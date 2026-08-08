package game

import (
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 敵人 AI 與傷害結算
//
// 一個 AI 單位的一回合(原版 `sub_A108`):
//
//	睡著        → 1/17 機率醒來,這回合不動(在排程裡處理)
//	被凍住      → 跳過
//	逃跑中      → 有 1/4 機率回一點血,然後只走不打
//	否則        → 先試攻擊(`sub_9F08`),打不到才移動(`sub_AE20`)
//	移動走出邊緣 → 印「escapes!」離場
//
// 選目標(`sub_AC40`)是**由 31 槽往 0 掃**、取歐氏距離最近的敵對單位。
// 由高往低掃這件事有觀察得到的後果:距離相同時**槽號小的贏**,
// 也就是隊伍前排(名冊順序在前的隊員)比較容易被鎖定。

// aiTurn 跑一個 AI 單位的回合。
func (s *State) aiTurn(idx int) {
	c := s.Combat
	u := &c.Units[idx]

	// 全域戰鬥模式先擋一關(原版 `sub_A108` 一開頭的兩個 cmp)。
	switch s.CombatMode {
	case CombatModeTimeStop:
		return
	case CombatModeSlow:
		// Rel Tym:這一回合有 1/2 機率整個跳過。
		if s.Roll(0, 1) == 0 {
			return
		}
	}

	if u.Flags&UnitFleeing != 0 {
		// 逃跑中的怪物每回合有 1/4 機率回一點血(原版 `sub_A108` 的
		// `random(0,3) == 3 → inc unit[+0]`;隊員不適用)。
		if !u.IsParty() && s.Roll(0, 3) == 3 {
			u.HP++
		}
		s.aiMove(idx)
		return
	}
	if s.aiAttack(idx) {
		return
	}
	s.aiMove(idx)
}

// combatDistance 是戰場上的距離:**floor(√(dx² + dy²))**。
//
// 原版 `sub_1F9F8` 拿 `sub_1F9CC` 算出 dx²+dy² 之後,用「連續減奇數」
// 那個古典整數開根號求 floor —— 所以斜角相鄰算 1 格,與正交相鄰同等。
func combatDistance(x1, y1, x2, y2 int) int {
	d := (x1-x2)*(x1-x2) + (y1-y2)*(y1-y2)
	n, i := 0, 1
	for i <= d {
		d -= i
		i += 2
		n++
	}
	return n
}

// aiTarget 選目標,並回傳往它走的一步方向(原版 `sub_AC40`)。
//
// 回傳的 dx/dy 已經套過逃跑反轉:逃跑中的單位拿到的是**背對目標**那一步。
func (s *State) aiTarget(idx int) (target, dx, dy int) {
	c := s.Combat
	u := &c.Units[idx]
	mine := s.hostile(u)
	// Quas An Wis:混亂的時候 AI 可能把陣營看反,結果打自己人。
	//
	// 原版 `sub_AC40`:`if (byte_3E08A == 'C' && random(0,60)/2 > sub_B398(自己, -1))
	// mySide = 0` —— 擲得過該單位的某個防禦值就翻面。
	if s.CombatMode == CombatModeConfuse {
		def := 0
		if st := s.creatureOf(u); st != nil {
			def = int(st.Intel)
		}
		if s.AttackRoll() > def {
			mine = false
		}
	}
	target, best := -1, 99
	tx, ty := u.X, u.Y
	// 由 31 往 0 掃 —— 距離平手時槽號小的勝出。
	for i := CombatUnitSlots - 1; i >= 0; i-- {
		if i == idx {
			continue
		}
		t := &c.Units[i]
		if !t.Active() || s.hostile(t) == mine {
			continue
		}
		// 看不見與被凍住的都選不到。
		if t.Flags&(UnitHidden|UnitGrabbed) != 0 {
			continue
		}
		if d := combatDistance(u.X, u.Y, t.X, t.Y); d < best {
			best, target = d, i
			tx, ty = t.X, t.Y
		}
	}
	if target < 0 {
		// 沒有目標了:原版讓所有還在場的怪物血量歸 1 並掛上逃跑旗標,
		// 然後往戰場正中央(5,5)算方向 —— 實際效果是四散離場。
		for i := u5data.CombatPartySlots; i < CombatUnitSlots; i++ {
			t := &c.Units[i]
			if t.Flags&UnitMonster != 0 {
				t.HP = 1
				t.Flags |= UnitFleeing
			}
		}
		tx, ty = u5data.CombatSide/2, u5data.CombatSide/2
	}
	if u.X > tx {
		dx = -1
	} else if u.X < tx {
		dx = 1
	}
	if u.Y > ty {
		dy = -1
	} else if u.Y < ty {
		dy = 1
	}
	if u.Flags&UnitFleeing != 0 {
		dx, dy = -dx, -dy
	}
	return target, dx, dy
}

// aiAttack 試著攻擊(原版 `sub_9F08`)。回傳是否用掉了這個回合。
func (s *State) aiAttack(idx int) bool {
	c := s.Combat
	u := &c.Units[idx]
	target, _, _ := s.aiTarget(idx)
	if target < 0 {
		return false
	}
	t := &c.Units[target]
	dist := combatDistance(u.X, u.Y, t.X, t.Y)

	reach := 1
	if st := s.creatureOf(u); st != nil {
		reach = int(st.Range)
	}
	if dist > reach {
		return false
	}
	if dist != 1 {
		return s.aiRangedAttack(idx, target)
	}
	s.resolveAttack(idx, target)
	return true
}

// aiRangedAttack 是遠程那條路(原版 `sub_9E10`)。回傳是否用掉了這個回合。
//
// 兩件在原版裡很關鍵、少了就會把遠程怪物寫得過強的事:
//
//  1. **每回合只有 1/2 機率真的射出來**(`random(0,255) >= 128 → return 0`)。
//     沒射成就回 0,`sub_A108` 接著去跑移動 —— 所以法師是「射一次、走一步」
//     而不是每回合都轟。擬態怪(0x1A)是唯一的例外,牠不動所以每回合都出手。
//  2. **轉化護符**(裝備 0x2D)戴在護符欄時,施法者的這一擊有 1/2 機率
//     直接算失手。
func (s *State) aiRangedAttack(idx, target int) bool {
	c := s.Combat
	u, t := &c.Units[idx], &c.Units[target]
	st := s.creatureOf(u)
	// In An:抗魔期間施法者放不出遠程(原版 `sub_9E10` 的
	// `cmp al, 'N' / jz loc_9F00`)。
	if st.Has(u5data.CreatureCasts) && s.CombatMode == CombatModeNegate {
		return false
	}
	if u.Creature != u5data.CreatureMimicIdx && s.Roll(0, 255) >= 128 {
		return false
	}
	forceMiss := false
	if st.Has(u5data.CreatureCasts) {
		if ch := s.charOf(t); ch != nil &&
			ch.Equipment().Amulet == u5data.ItemAmuletOfTurning && s.Roll(0, 255) < 128 {
			forceMiss = true
		}
	}
	s.Log(s.unitName(u) + "朝" + s.unitName(t) + "發動遠程攻擊。")
	tx, ty := t.X, t.Y
	if forceMiss || !s.attackHits(u, t, u5data.ItemNone) {
		// 失手的投射物**不會消失**,它會偏到目標附近一格繼續飛
		//(原版 `sub_1FE54` 在沒命中時用 `sub_1FDE8` 挑一個鄰格當終點)。
		tx, ty = t.X+s.Roll(-1, 1), t.Y+s.Roll(-1, 1)
	}
	// 真正打到誰由飛行路徑決定 —— 站在射線上的人會替目標擋下來。
	victim, _, _ := s.FlyProjectile(idx, tx, ty)
	if victim < 0 {
		s.Log("投射物落空了。")
		return true
	}
	s.resolveAttack(idx, victim)
	return true
}

// aiMove 移動一步(原版 `sub_AE20`)。回傳有沒有真的動。
//
// 原版的挑法:先擲一枚硬幣決定要不要優先走 X 軸,走不了就試 Y 軸,
// 兩軸都不行才隨機試四個方向、最多四次。**不是最短路徑演算法** ——
// 這也是原版怪物會卡在牆角繞的原因。
func (s *State) aiMove(idx int) bool {
	c := s.Combat
	u := &c.Units[idx]
	if st := s.creatureOf(u); st != nil {
		// 擬態怪與收割者從不移動 —— 原版 `sub_AE20` 開頭就 return。
		if u.Creature == u5data.CreatureMimicIdx || u.Creature == u5data.CreatureReaperIdx {
			return false
		}
		// 會瞬移的(鬼火)先試瞬移。原版挑落點的 `sub_B1D8` 還沒逆完,
		// 這裡在戰場上隨機找一格空地,行為等價。
		// In An 也擋瞬移(原版 `sub_AE20` 的同一個 `cmp dl, 'N'`)。
		if st.Has(u5data.CreatureTeleports) && s.CombatMode != CombatModeNegate &&
			s.Roll(0, 3) != 3 {
			if s.aiTeleport(idx) {
				return true
			}
		}
	}
	_, dx, dy := s.aiTarget(idx)

	moved := false
	if s.Roll(0, 255) > 127 && dx != 0 && !s.combatBlocked(idx, u.X+dx, u.Y) {
		dy, moved = 0, true
	} else if dy != 0 && !s.combatBlocked(idx, u.X, u.Y+dy) {
		dx, moved = 0, true
	}
	if !moved {
		// 四次隨機嘗試,方向照原版跳表的順序:南 東 北 西。
		for try := 0; try < 4 && !moved; try++ {
			switch s.Roll(0, 3) {
			case 0:
				dx, dy = 0, 1
			case 1:
				dx, dy = 1, 0
			case 2:
				dx, dy = 0, -1
			default:
				dx, dy = -1, 0
			}
			if !s.combatBlocked(idx, u.X+dx, u.Y+dy) {
				moved = true
			}
		}
	}
	if !moved {
		return false
	}
	u.X, u.Y = u.X+dx, u.Y+dy
	if u.X < 0 || u.X >= u5data.CombatSide || u.Y < 0 || u.Y >= u5data.CombatSide {
		s.Log(s.unitName(u) + "逃走了!")
		u.Flags |= UnitDead
	}
	return true
}

// aiTeleport 把單位挪到戰場上隨機一格空地。
func (s *State) aiTeleport(idx int) bool {
	c := s.Combat
	u := &c.Units[idx]
	for try := 0; try < 8; try++ {
		x, y := s.Roll(0, u5data.CombatSide-1), s.Roll(0, u5data.CombatSide-1)
		if s.combatBlocked(idx, x, y) {
			continue
		}
		u.X, u.Y = x, y
		s.Log(s.unitName(u) + "瞬間移動了!")
		return true
	}
	return false
}

// combatBlocked 回報 (x, y) 這一格走不走得過去(原版 `sub_16454`)。
//
// 出界一律算擋住,**除非**這個單位正在逃跑 —— 逃跑就是靠走出邊緣完成的。
func (s *State) combatBlocked(idx, x, y int) bool {
	c := s.Combat
	u := &c.Units[idx]
	if x < 0 || x >= u5data.CombatSide || y < 0 || y >= u5data.CombatSide {
		return u.Flags&UnitFleeing == 0
	}
	if u5data.TileBlocksWalking(int(c.Map.At(x, y))) {
		return true
	}
	_, taken := c.CombatUnitAt(x, y)
	return taken
}

// resolveAttack 結算一次攻擊:命中 → 特殊效果 → 傷害 → 死亡 → 經驗值。
//
// 對應原版的 `sub_B484`(命中)→ `sub_B9A8`(命中之後)→ `sub_B274`(傷害)
// → `sub_B51C`(扣血與死亡)這一條鏈。
func (s *State) resolveAttack(attacker, target int) {
	c := s.Combat
	a, t := &c.Units[attacker], &c.Units[target]
	an, tn := s.unitName(a), s.unitName(t)

	// 記下「誰打了我」—— 原版 `byte_3E0B8[目標] = 攻擊者`,
	// 施法被打斷的判定只認這個人。
	c.LastAttacker[target] = int8(attacker)

	weapon := byte(u5data.ItemNone)
	if ch := s.charOf(a); ch != nil {
		weapon = ch.Equipment().Weapon
	}
	if !s.attackHits(a, t, weapon) {
		s.Log(an + "揮空了。")
		return
	}
	// 小魔怪偷食物:命中之後有 3/4 機率不造成傷害,改偷糧食。
	if st := s.creatureOf(a); st != nil && st.Has(u5data.CreatureStealsFood) &&
		s.Roll(0, 3) != 0 && s.Inventory.Food > 0 {
		stolen := s.Inventory.Food
		if stolen > 5 {
			stolen = 5
		}
		s.Inventory.Food -= stolen
		s.Log(an + "偷走了糧食!")
		return
	}
	// 下毒的怪物:命中之後有 3/4 機率改成下毒。
	if st := s.creatureOf(a); st != nil && st.IsPoisonous() && s.Roll(0, 3) != 0 {
		s.poisonHit(attacker, target)
		return
	}
	// 注視者的凝視讓人睡著。
	if a.Creature == u5data.CreatureGazerIdx && t.Flags&UnitAsleep == 0 {
		t.Flags |= UnitAsleep
		s.Log(tn + "睡著了!")
		return
	}
	// 拖屍怪把隊員拖到水下(原版 `sub_1F840`)。與注視者同一層:
	// 命中之後改成施加狀態,**不再算傷害**。
	if s.corpserGrab(attacker, target) {
		return
	}

	base, shattered := s.attackDamage(a, weapon)
	if shattered {
		if ch := s.charOf(a); ch != nil {
			ch.Raw[u5data.CharWeapon] = u5data.ItemNone
		}
		s.Log("汝的劍碎了!")
	}
	dmg := s.damageAfterResist(base, t)
	if dmg < 0 && t.IsParty() {
		// 防禦擲贏了傷害:對隊員就是完全擋下。
		s.Log(tn + "擋下了這一擊。")
		return
	}
	s.applyDamage(attacker, target, dmg)
}

// attackHits 判命中(原版 `sub_B484`)。
func (s *State) attackHits(a, t *Combatant, weapon byte) bool {
	if a.IsParty() && AlwaysHitWeapons[weapon] {
		return true
	}
	attack := 0
	if ch := s.charOf(a); ch != nil {
		attack = s.attackValueOf(ch, weapon)
	} else if st := s.creatureOf(a); st != nil {
		attack = int(st.Strength)
	}
	defence := 0
	if ch := s.charOf(t); ch != nil {
		defence = s.Stats.DefenceOf(ch)
	} else if st := s.creatureOf(t); st != nil {
		defence = int(st.Intel)
	}
	return s.AttackRoll() >= HitThreshold(defence, attack)
}

// attackDamage 算基礎傷害:怪物走屬性表的攻擊力,隊員走武器表。
func (s *State) attackDamage(a *Combatant, weapon byte) (base int, shattered bool) {
	if !a.IsParty() {
		if st := s.creatureOf(a); st != nil {
			return int(st.Attack), false
		}
		return u5data.BareHandDamage, false
	}
	return s.WeaponDamageRoll(weapon)
}

// damageAfterResist 扣掉目標的減傷。
func (s *State) damageAfterResist(base int, t *Combatant) int {
	if ch := s.charOf(t); ch != nil {
		return s.DamageToCharacter(base, ch)
	}
	if st := s.creatureOf(t); st != nil {
		return s.applyResist(base, int(st.Armour))
	}
	return base
}

// poisonHit 是下毒攻擊(原版 `sub_B8DC`)。
//
// 對狀態「良好」的隊員 → 中毒;其餘一律當成 0..20 的傷害。
func (s *State) poisonHit(attacker, target int) {
	c := s.Combat
	t := &c.Units[target]
	if ch := s.charOf(t); ch != nil && ch.Status == u5data.StatusGood {
		ch.Status = u5data.StatusPoisoned
		s.Log(ch.Name + "中毒了!")
		return
	}
	s.applyDamage(attacker, target, s.Roll(0, 20))
}

// applyDamage 扣血、判死亡、發經驗值(原版 `sub_B51C` + `sub_B9A8` 的尾段)。
func (s *State) applyDamage(attacker, target, dmg int) {
	c := s.Combat
	a, t := &c.Units[attacker], &c.Units[target]
	tn := s.unitName(t)

	if st := s.creatureOf(t); st != nil {
		// 殺不死的三位:傷害直接歸零。
		if st.Has(u5data.CreatureInvulnerable) {
			s.Log(tn + "毫髮無傷。")
			return
		}
		// 幽靈那一類實體攻擊只吃一半。原版的例外條件 `byte_3E0A0`
		// (疑為魔法武器)還沒逆完,先一律減半。
		if st.Has(u5data.CreatureHalfDamage) && dmg != u5data.InstantKillDamage {
			dmg /= 2
		}
	}

	xp := 0
	if ch := s.charOf(t); ch != nil {
		hp := int(ch.HP) - dmg
		if hp < 1 || dmg == u5data.InstantKillDamage {
			hp = 0
			ch.Status = u5data.StatusDead
			t.Flags |= UnitDead
			s.Log(tn + "倒下了!")
		}
		if hp < 0 {
			hp = 0
		}
		ch.HP = uint16(hp)
		if !t.Dead() {
			// ★ 隊員被打**只說一句「被擊中了」,不報數字**(原版 `sub_1F840`
			// 的 `" hit!"`)—— 血量在 Ztats 看得到,不必在訊息欄重複。
			s.Log(tn + MsgWasHit)
		}
	} else {
		// ⚠ 原版在這裡是位元組減法:傷害是負數時怪物**反而回血**。
		// 照抄 —— 這是防禦擲贏傷害時的既有行為,不是我加的平衡。
		if t.HP < dmg {
			t.HP = 0
		} else {
			t.HP -= dmg
		}
		if t.HP == 0 || dmg == u5data.InstantKillDamage {
			t.Flags |= UnitDead
			xp = s.killCreature(target)
		} else {
			// ★ 怪物被打**報的是傷勢等級,不是數字**(原版 `sub_1F840` 的四句)。
			// 這是刻意的設計:玩家看不到敵人的血量,只能靠形容詞判斷 ——
			// 報數字等於洩漏原版藏起來的資訊。
			s.Log(tn + s.woundReport(target))
			s.maybeDivide(target)
		}
	}
	// 經驗值只有隊員拿得到,而且上限 9999(原版 `sub_2BBDC` 的第三個參數)。
	if xp > 0 {
		if ch := s.charOf(a); ch != nil {
			e := int(ch.Exp) + xp
			if e > 9999 {
				e = 9999
			}
			ch.Exp = uint16(e)
		}
	}
}

// killCreature 處理怪物死亡,回傳該給的經驗值。
//
// 經驗值 = **生命上限 / 4 + 1**(原版 `sub_B51C` 的 `byte_3F055[生物*8] >> 2 + 1`)。
// 所以經驗值只跟這種怪物有多耐打有關,與實際打了幾下無關。
func (s *State) killCreature(idx int) int {
	c := s.Combat
	u := &c.Units[idx]
	name := s.unitName(u)
	st := s.creatureOf(u)
	if st == nil {
		s.Log(name + "被打倒了!")
		return 0
	}
	if st.Has(u5data.CreatureVanishes) {
		s.Log(name + "消失了!")
	} else {
		s.Log(name + "被打倒了!")
	}
	return int(st.MaxHP)/4 + 1
}

// maybeDivide 是史萊姆的分裂(原版 `sub_B51C` 的 `test al, 0x10`)。
//
// 沒死才會分裂,最多試 8 次找空位,分身的血與本體同步。
func (s *State) maybeDivide(idx int) {
	c := s.Combat
	u := &c.Units[idx]
	st := s.creatureOf(u)
	if st == nil || !st.Has(u5data.CreatureDivides) {
		return
	}
	for try := 0; try < 8; try++ {
		x, y := u.X+s.Roll(-1, 1), u.Y+s.Roll(-1, 1)
		if s.combatBlocked(idx, x, y) {
			continue
		}
		for i := u5data.CombatPartySlots; i < CombatUnitSlots; i++ {
			if c.Units[i].Flags != 0 {
				continue
			}
			n := *u
			n.X, n.Y = x, y
			n.Init = n.resetInit()
			c.Units[i] = n
			s.Log(s.unitName(u) + "分裂了!")
			return
		}
		return
	}
}

// itoa 是給訊息用的小工具(避免整支檔案為了一個數字去 import strconv)。
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [8]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

// strugglingUnderwater 是被拖屍怪拖到水下的那個人的一回合(原版 `sub_A360`
// 的旗標 4 分支 + `sub_BCC4`)。
//
//	印 "ARGH!"
//	門檻 = max(1, rand(0, 60) / 2)
//	if (敏捷 > 門檻) 印 <名字> " regurgitated!"、清旗標
//	不論掙不掙脫,這一回合都用掉了
//
// ⚠ 掙脫判的是**敏捷**(`unit[+1]`,與行動倒數 `36 − 敏捷` 讀的是同一格),
// 不是力量、也不是生命。
func (s *State) strugglingUnderwater(u *Combatant) {
	s.Log(MsgArgh)
	dex := 0
	if ch := s.charOf(u); ch != nil {
		dex = int(ch.Dex)
	} else if st := s.creatureOf(u); st != nil {
		dex = int(st.Dex)
	}
	if dex > u5data.CorpserEscapeThreshold(s.Roll(0, u5data.CorpserEscapeRollMax)) {
		u.Flags &^= UnitGrabbed
		s.Log(s.unitName(u) + MsgRegurgitated)
	}
}

// corpserGrab 是拖屍怪打中**隊員**時的特殊效果(原版 `sub_1F840`)。
//
// ⚠ 只對隊員生效 —— 原版那個分支的前提是 `test si, 80h`(目標是隊員)。
// 拖屍怪打怪物就只是普通命中。
func (s *State) corpserGrab(attacker, target int) bool {
	a, t := &s.Combat.Units[attacker], &s.Combat.Units[target]
	if a.Creature != u5data.CreatureCorpserIdx || !t.IsParty() {
		return false
	}
	if t.Flags&UnitGrabbed != 0 {
		return false
	}
	t.Flags |= UnitGrabbed
	s.Log(s.unitName(t) + MsgDraggedUnder)
	return true
}


// 傷勢等級與「打怕了就跑」(原版 `sub_BAFC`)
//
// 這一支同時做兩件事,而引擎原本兩件都沒有:
//
//	① 把怪物的血量換成四句形容詞 —— 玩家判斷敵人狀況的唯一依據
//	② **設逃跑旗標** —— 打到剩 1/4 血以下的怪物會跑
//
// `docs/re/16` 原本寫「逃跑就是靠走出邊緣完成的,沒有另外的『逃跑判定』」——
// 那句話錯了,判定就在這裡(`docs/re/68`)。
//
//	上限 = byte_3F055[生物 × 8]
//	if (上限/4     > 血) { 等級 = 1; 逃 }                      ; critical
//	if (上限/2     > 血) { 等級 = 2; rand(0,256) > 251 也逃 }   ; heavily
//	if (上限×3/4   > 血) { 等級 = 3 }                          ; lightly
//	否則                  { 等級 = 4 }                          ; barely
//
// ⚠ 第三個門檻原版算的是 `(上限/4/2)*3`,不是 `上限*3/4` ——
// 整數除法先做,兩者在上限不是 8 的倍數時**會差**。照原版的順序算。

// WoundLevel 是四個傷勢等級,1 最重。
const (
	WoundCritical = 1
	WoundHeavily  = 2
	WoundLightly  = 3
	WoundBarely   = 4
)

// woundLevel 算傷勢等級,並依原版的規則決定要不要掛上逃跑旗標。
func (s *State) woundLevel(idx int) int {
	u := &s.Combat.Units[idx]
	st := s.creatureOf(u)
	if st == nil {
		return WoundBarely
	}
	hp := int(u.HP)
	quarter := int(st.MaxHP) / 4
	flee := false
	level := WoundBarely
	switch {
	case quarter > hp:
		level, flee = WoundCritical, true
	case quarter*2 > hp:
		level = WoundHeavily
		// 半血以下還有一小段機率會跑(原版 `rand(0, 256) > 0FBh`)。
		flee = s.Roll(0, 256) > 0xFB
	case (quarter*2/2)*3 > hp:
		// ⚠ 照原版的算式:先 `sar eax,1` 再 `lea esi,[esi+esi*2]`。
		level = WoundLightly
	}
	if flee {
		u.Flags |= UnitFleeing
	} else {
		u.Flags &^= UnitFleeing
	}
	return level
}

// woundReport 回傳要接在名字後面的那句話。
func (s *State) woundReport(idx int) string {
	switch s.woundLevel(idx) {
	case WoundCritical:
		return MsgWoundCritical
	case WoundHeavily:
		return MsgWoundHeavily
	case WoundLightly:
		return MsgWoundLightly
	}
	return MsgWoundBarely
}
