package game

import (
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 魔法
//
// 施法的流程完全照原版 `sub_1994C`,順序很重要 —— 每一關都有自己的失敗訊息,
// 而且**已調配的份數在魔力檢查之前就先扣掉了**:
//
//	1. 場合對不對(`byte_40E7C[咒語]` vs 當前地點)  → 「Not here!」
//	2. 有沒有調配好的份數                            → 「None mixed!」
//	3. ★ 份數 −1                                     ← 後面失敗也不退
//	4. 魔力夠不夠(消耗 = 圈數)                      → 「M.P. too low!」
//	5. 魔力 −= 圈數
//	6. 等級夠不夠(等級 ≥ 圈數)                      → 失敗
//	7. 跑效果
//	8. 「Success!」/「Failed!」
//
// 戰鬥中還多一關(`sub_200BC`):**上一個打你的敵人如果就站在旁邊,
// 咒語會被打斷**（印「X interferes!」)。

// MagicResult 是一次施法的結果。
type MagicResult int

const (
	// MagicNotHere:這個場合不能施這個咒語。
	MagicNotHere MagicResult = iota
	// MagicNoneMixed:沒有調配好的份數。
	MagicNoneMixed
	// MagicNoMana:魔力不足。
	MagicNoMana
	// MagicInterfered:戰鬥中被相鄰的敵人打斷。
	MagicInterfered
	// MagicFailed:發動了但失敗(等級不足,或效果本身回報失敗)。
	MagicFailed
	// MagicSuccess:成功。
	MagicSuccess
	// MagicAbsorbed:這個地點把咒語吸走了(見 magicAbsorbedHere)。
	MagicAbsorbed
)

// magicAbsorbedHere 回報這個地點會不會把咒語吸掉(原版 `sub_1994C` 的地點分派)。
//
//	if (地點 == 0x12 && 還沒拿到王冠) → "Absorbed!"
//	if (地點 == 0x1D)                 → "Absorbed!"
//
// ★ 這兩個地點正好就是**王冠與權杖的所在**(`docs/re/57`):
// 第二座城堡在你拿到王冠之前壓制魔法,`STONEGATE` 永遠壓制。
// 兩個號碼分別由信物擺放位置獨立佐證,不是巧合。
//
// ⚠ 這一關在「場合對不對」與「有沒有調配」**之前** —— 所以在這兩處施法
// 不會扣份數也不會扣魔力。順序寫錯的話玩家會白白損失藥草。
func (s *State) magicAbsorbedHere() bool {
	if s.InCombat() {
		return false
	}
	switch s.Location {
	case u5data.CrownNPCLocation:
		return !s.Regalia.Crown
	case u5data.SceptreNPCLocation:
		return true
	}
	return false
}

// castLocation 回傳給咒語場合判斷用的地點值:戰鬥中是 −1。
func (s *State) castLocation() int {
	if s.InCombat() {
		return u5data.CombatLocation
	}
	return s.Location
}

// Cast 讓名冊第 caster 個角色施第 spell 個咒語。
func (s *State) Cast(caster, spell int) MagicResult {
	if s.Spells == nil || spell < 0 || spell >= u5data.SpellCount {
		return MagicFailed
	}
	ch := s.rosterAt(caster)
	if ch == nil {
		return MagicFailed
	}
	sp := s.Spells.Spells[spell]

	// 戰鬥中:上一個打你的人站在旁邊就會被打斷。
	if s.InCombat() {
		if who, ok := s.interferer(caster); ok {
			s.Log(who + "干擾了施法!")
			return MagicInterfered
		}
	}
	// ★ 有兩個地點會直接把咒語吸走,連場合都不看(原版的地點分派最前面)。
	if s.magicAbsorbedHere() {
		s.Log(MsgMagicAbsorbed)
		return MagicAbsorbed
	}
	if !sp.CanCastAt(s.castLocation()) {
		s.Log("此地不可!")
		return MagicNotHere
	}
	if s.Inventory.Spells[spell] <= 0 {
		s.Log("尚未調配!")
		return MagicNoneMixed
	}
	// ★ 先扣份數 —— 後面任何一步失敗都不退。原版就是這個順序。
	s.Inventory.Spells[spell]--

	if int(ch.MP) < sp.Circle {
		s.Log("魔力不足!")
		return MagicNoMana
	}
	ch.MP -= byte(sp.Circle)

	if int(ch.Level) < sp.Circle {
		s.Log("失敗!(等級不足以驅動第 " + itoa(sp.Circle) + " 圈)")
		return MagicFailed
	}
	if s.spellEffect(caster, spell) {
		s.Log("成功!")
		return MagicSuccess
	}
	s.Log(MsgFailed)
	return MagicFailed
}

// interferer 找出戰鬥中打斷施法的那個敵人(原版 `sub_200BC`)。
//
// 條件全部成立才算干擾:那個單位是**上一個攻擊施法者的人**、還活著、
// 站在敵方、沒被凍住也沒睡著、時間沒被停住,而且**距離正好 1**。
func (s *State) interferer(caster int) (string, bool) {
	c := s.Combat
	if c == nil {
		return "", false
	}
	self := s.combatSlotOfRoster(caster)
	if self < 0 {
		return "", false
	}
	by := int(c.LastAttacker[self])
	if by < 0 || by >= CombatUnitSlots {
		return "", false
	}
	u, t := &c.Units[self], &c.Units[by]
	if !t.Active() || !s.hostile(t) || s.hostile(u) {
		return "", false
	}
	if t.Flags&(UnitGrabbed|UnitAsleep) != 0 {
		return "", false
	}
	if combatDistance(u.X, u.Y, t.X, t.Y) != 1 {
		return "", false
	}
	return s.unitName(t), true
}

// combatSlotOfRoster 找名冊第 n 個角色在戰場上的槽號;不在場回 -1。
func (s *State) combatSlotOfRoster(n int) int {
	c := s.Combat
	if c == nil {
		return -1
	}
	for i := 0; i < u5data.CombatPartySlots; i++ {
		if c.Units[i].IsParty() && c.Units[i].Roster == n {
			return i
		}
	}
	return -1
}

// rosterAt 取名冊第 n 個角色。
func (s *State) rosterAt(n int) *u5data.Character {
	if n < 0 || n >= len(s.Roster) {
		return nil
	}
	return &s.Roster[n]
}

// 咒語索引。用得到的才取名 —— 其餘照編號走 spellEffect 的 default。
const (
	SpellInLor       = 0  // 光
	SpellGravPor     = 1  // 魔法飛彈
	SpellAnZu        = 2  // 喚醒
	SpellAnNox       = 3  // 解毒
	SpellMani        = 4  // 治療
	SpellAnYlem      = 5  // 消除
	SpellAnSanct     = 6  // 解陷阱 / 開鎖 / 開箱
	SpellAnXenCor    = 7  // 驅離生物
	SpellRelHur      = 8  // 改風向
	SpellInWis       = 9  // 定位
	SpellKalXen      = 10 // 召喚野獸
	SpellInXenMani   = 11 // 造食物
	SpellVasLor      = 12 // 大光明
	SpellVasFlam     = 13 // 火球
	SpellInFlamGrav  = 14 // 烈焰力場
	SpellInNoxGrav   = 15 // 毒力場
	SpellInZuGrav    = 16 // 睡眠力場
	SpellInPor       = 17 // 瞬移
	SpellAnGrav      = 18 // 破力場
	SpellInSanct     = 19 // 防護
	SpellInSanctGrav = 20 // 防護力場
	SpellUusPor      = 21 // 上樓
	SpellDesPor      = 22 // 下樓
	SpellWisQuas     = 23 // 顯形
	SpellInBetXen    = 24 // 召喚蟲群
	SpellAnExPor     = 25 // 魔法上鎖
	SpellInExPor     = 26 // 魔法解鎖
	SpellVasMani     = 27 // 大治療
	SpellInZu        = 28 // 睡眠風
	SpellRelTym      = 29 // 緩速
	SpellInVasPorY   = 30 // 能量爆
	SpellQuasAnWis   = 31 // 混亂
	SpellInAn        = 32 // 抗魔
	SpellWisAnYlem   = 33 // 顯示地圖
	SpellAnXenEx     = 34 // 魅惑
	SpellRelXenBet   = 35 // 變形
	SpellSanctLor    = 36 // 隱形
	SpellXenCorp     = 37 // 殺
	SpellInQuasXen   = 38 // 幻影
	SpellInQuasWis   = 39 // 全景(Peer)
	SpellInNoxHur    = 40 // 毒風
	SpellInQuasCorp  = 41 // 恐懼
	SpellInManiCorp  = 42 // 復活
	SpellKalXenCorp  = 43 // 召喚惡魔
	SpellInVasGravC  = 44 // 能量風
	SpellInFlamHur   = 45 // 火風
	SpellVasRelPor   = 46 // 大傳送門
	SpellAnTym       = 47 // 時間停止
)

// 三個「指定目標打一下」的咒語各自帶一個攻擊碼(原版 `sub_189E4` 寫進 `byte_3E0AD`)。
const (
	spellAttackGravPor = u5data.AttackGravPor
	spellAttackVasFlam = u5data.AttackVasFlam
	spellAttackXenCorp = u5data.AttackXenCorp
)

// TimeStopTurns 是 An Tym 停多久(原版 `sub_198E0` 的 `byte_3E09E = 0Ah`)。
const TimeStopTurns = 10

// LightSpellTurns 是 In Lor / Vas Lor 的持續回合(原版 `sub_1D310` 的參數)。
const (
	LightSpellTurns     = 100
	GreatLightSpellTurn = 255
)

// spellEffect 跑咒語的效果,回傳成功與否(對應原版跳表 `jpt_19B27` 的 48 個 case)。
//
// ⚠ 這裡只實作**效果已經逆完**的那些;其餘照實說明還沒做,而且**照原版一樣
// 已經把份數與魔力扣掉了** —— 假裝成功比誠實失敗更糟。
func (s *State) spellEffect(caster, spell int) bool {
	switch spell {
	case SpellInLor:
		s.LightTurns = LightSpellTurns
		return true
	case SpellVasLor:
		s.LightTurns = GreatLightSpellTurn
		return true

	case SpellAnZu: // 喚醒:狀態 'S' → 'G',戰場上也清掉睡著旗標
		return s.cureStatus(u5data.StatusAsleep)
	case SpellAnNox: // 解毒:狀態 'P' → 'G'
		return s.cureStatus(u5data.StatusPoisoned)

	case SpellMani: // 治療:回 1..30,上限 MaxHP;死人治不了
		return s.healTarget(s.spellTarget(caster, false), s.AttackRoll())
	case SpellVasMani: // 大治療:回滿
		t := s.spellTarget(caster, false)
		ch := s.rosterAt(t)
		if ch == nil || ch.Status == u5data.StatusDead {
			return false
		}
		ch.HP = ch.MaxHP
		s.Log(ch.Name + "的傷全好了。")
		return true

	case SpellInManiCorp: // 復活
		t := s.spellTarget(caster, true)
		ch := s.rosterAt(t)
		if ch == nil || ch.Status != u5data.StatusDead {
			s.Log("此人未死!")
			return false
		}
		ch.Status = u5data.StatusGood
		ch.HP = 1
		s.Log(ch.Name + "回到了人世。")
		return true

	case SpellGravPor:
		return s.spellAttack(caster, spellAttackGravPor)
	case SpellVasFlam:
		return s.spellAttack(caster, spellAttackVasFlam)
	case SpellXenCorp:
		return s.spellAttack(caster, spellAttackXenCorp)

	case SpellAnXenEx: // 魅惑:讓一隻怪物改站我方
		return s.charmNearest(caster)

	case SpellAnTym: // 時間停止
		s.Log("時間靜止了。")
		return s.setCombatMode(CombatModeTimeStop, TimeStopTurns)

	case SpellUusPor:
		return s.spellChangeFloor(true)
	case SpellDesPor:
		return s.spellChangeFloor(false)

	// 五個共用 `byte_3E08A` 的持續效果(`sub_1D31C(模式, 回合, 音效)`)。
	case SpellInSanct:
		return s.setCombatMode(CombatModeProtected, 20)
	case SpellRelTym:
		return s.setCombatMode(CombatModeSlow, 30)
	case SpellQuasAnWis:
		return s.setCombatMode(CombatModeConfuse, 20)
	case SpellInAn:
		return s.setCombatMode(CombatModeNegate, 10)

	case SpellInXenMani: // 造食物:糧食 + random(1,3),上限 9999
		s.Inventory.Food = addCap(s.Inventory.Food, s.Roll(1, 3), 9999)
		s.Log("糧食憑空出現。")
		return true

	case SpellInWis: // 定位:報出目前的世界座標
		s.Log("汝身在 (" + itoa(s.X) + ", " + itoa(s.Y) + ")。")
		return true

	case SpellSanctLor: // 隱形:掛上隱形位元,選目標時就選不到了
		return s.hideCaster(caster)

	case SpellInVasPorY: // 能量爆:對場上每個敵人各擲一次
		return s.energyBlast(caster)

	case SpellKalXen: // 召喚巨鼠
		return s.summonCreature(caster, summonRat, 1)
	case SpellInBetXen: // 召喚蟲群,一次 4 隻
		return s.summonCreature(caster, summonInsect, 4)
	case SpellKalXenCorp: // 召喚惡魔
		return s.summonCreature(caster, summonDaemon, 1)

	case SpellAnXenCor: // 驅離 —— 只對帶 0x20 位元的四種生物
		return s.frighten(caster, true)
	case SpellInQuasCorp: // 恐懼 —— 場上所有敵人
		return s.frighten(caster, false)
	case SpellRelXenBet: // 變形
		return s.polymorph(caster)
	case SpellWisQuas: // 顯形
		return s.revealHidden()

	case SpellVasRelPor: // 大傳送門:走到此刻月相的目的地
		return s.CastGreatGate(s.MoonPhaseNow())

	case SpellAnSanct: // 解陷阱 / 開箱 / 開鎖
		return s.disarmOrUnlock()
	case SpellRelHur: // 改風向 —— 原版問方向
		s.AskDirection(func(d Direction) { s.ChangeWind(d) })
		return true

	case SpellInExPor: // 魔法解鎖 —— 原版問方向
		s.AskDirection(func(d Direction) { s.MagicUnlockAhead(d) })
		return true
	case SpellAnExPor: // 魔法上鎖
		s.AskDirection(func(d Direction) { s.MagicLockAhead(d) })
		return true
	case SpellAnYlem: // 消除
		s.AskDirection(func(d Direction) { s.DispelAhead(d) })
		return true
	case SpellInPor: // 瞬移
		return s.blink(caster)
	case SpellWisAnYlem: // 牆壁擋不住視線
		s.SeeThroughWalls = true
		s.Log("石牆在汝眼前變得透明。")
		return true
	case SpellInQuasWis: // 全景
		return s.Peer()
	case SpellInQuasXen: // 幻影:複製一個目標
		return s.illusion(caster)
	case SpellAnGrav: // 破力場
		return s.destroyField()

	// 四個力場咒語 —— 同一支函式,差一個種類碼與一個攻擊碼。
	case SpellInFlamGrav:
		return s.castField(fieldFire, u5data.AttackInFlamGrav)
	case SpellInNoxGrav:
		return s.castField(fieldPoison, u5data.AttackInNoxGrav)
	case SpellInZuGrav:
		return s.castField(fieldSleep, u5data.AttackInZuGrav)
	case SpellInSanctGrav:
		return s.castField(fieldElectric, u5data.AttackInSanctGrav)

	// 四種風。原版每一次都問方向(`sub_1CC50`),引擎現在也問。
	case SpellInZu:
		return s.askWind(caster, windSleep)
	case SpellInNoxHur:
		return s.askWind(caster, windPoison)
	case SpellInFlamHur:
		return s.askWind(caster, windFire)
	case SpellInVasGravC:
		return s.askWind(caster, windEnergy)
	}
	s.Log(spellUndispatched)
	return false
}

// spellUndispatched 是「跳表漏了這一格」的訊息。
//
// 原版 `jpt_19B27` 是一張滿的 48 格跳表,沒有漏接這回事 ——
// 所以這一行印出來就代表引擎有缺口,不是咒語失敗。
// `TestEverySpellIsDispatched` 拿它當偵測訊號。
const spellUndispatched = "(此咒語的效果尚未實作 —— 藥草與魔力已照原版消耗)"

// spellTarget 挑「對某個隊員生效」的咒語要作用在誰身上。
//
// 原版會先問「Who?」(`sub_1C9C0` 印出隊伍清單等一個數字鍵)。引擎還沒有
// 那一層選單,先套用「戰鬥中是施法者自己、平時是隊伍裡傷得最重的那一個」
// —— **這是介面上的近似,不是規則**,補上選單之後就換掉。
//
// wantDead 為真時改找死掉的那一個(復活咒語用)。
func (s *State) spellTarget(caster int, wantDead bool) int {
	if wantDead {
		for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
			if s.Roster[i].Status == u5data.StatusDead {
				return i
			}
		}
		return caster
	}
	if s.InCombat() {
		return caster
	}
	best, worst := caster, 1<<30
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		c := &s.Roster[i]
		if c.Status == u5data.StatusDead {
			continue
		}
		if d := int(c.MaxHP) - int(c.HP); d > 0 && int(c.HP) < worst {
			worst, best = int(c.HP), i
		}
	}
	return best
}

// healTarget 回血,上限 MaxHP;死人治不了(原版 `sub_1CD3C`)。
func (s *State) healTarget(target, amount int) bool {
	ch := s.rosterAt(target)
	if ch == nil || ch.Status == u5data.StatusDead {
		return false
	}
	hp := int(ch.HP) + amount
	if hp > int(ch.MaxHP) {
		hp = int(ch.MaxHP)
	}
	ch.HP = uint16(hp)
	s.Log(ch.Name + "回復了 " + itoa(amount) + " 點體力。")
	return true
}

// cureStatus 把隊伍裡所有處在某個狀態的人恢復成良好。
func (s *State) cureStatus(status byte) bool {
	any := false
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		ch := &s.Roster[i]
		if ch.Status != status {
			continue
		}
		ch.Status = u5data.StatusGood
		any = true
		s.Log(ch.Name + "恢復了。")
		if status == u5data.StatusAsleep && s.InCombat() {
			if slot := s.combatSlotOfRoster(i); slot >= 0 {
				s.Combat.Units[slot].Flags &^= UnitAsleep
			}
		}
	}
	return any
}

// spellAttack 是**七個**「指定目標打一下」的咒語(原版 `sub_189E4` / `sub_18A08`
// → `sub_20360` → `sub_20134`)。
//
// 三個攻擊咒語(Grav Por / Vas Flam / Xen Corp)加上四個 `*Grav` 力場在**戰鬥中**
// 走的是同一條路 —— 都是射程 15 的指定目標攻擊,傷害與射程都在那四張 56 筆的
// 裝備表裡(見 `u5data.AttackCodeCount`)。
//
// ⚠ **四個 `*Grav` 在戰鬥中不是「放一片持續力場」**,是一發遠程攻擊。
// 引擎原本在戰鬥中直接回失敗並印「尚未實作」—— 那個判斷是對的
// (當時真的沒證據),但答案其實就在同一組表的後面 8 筆裡。
//
//	In Nox Grav  傷害 18 → 但命中後走**中毒**分支,不扣血
//	In Zu Grav   傷害  0 → 命中後**睡著**
//	In Flam Grav 傷害 21 → 一般傷害
//	In Sanct Grav傷害  0 → 追到這裡是**沒有效果**;若它其實會留下力場,
//	                       那要在投射物飛行(`sub_1FE54`)裡,還沒讀
func (s *State) spellAttack(caster, code int) bool {
	c := s.Combat
	if c == nil {
		s.Log("此地無敵可擊。")
		return false
	}
	self := s.combatSlotOfRoster(caster)
	if self < 0 {
		return false
	}
	target, _, _ := s.aiTarget(self)
	if target < 0 {
		s.Log("沒有目標。")
		return false
	}
	// 命中判定:只有 Xen Corp 要擲(`sub_B484` 的施法分支)。
	//
	// ⚠ 用 `SpellResisted` 而不是 `resists` —— **這裡沒有免疫名單**。
	// `sub_B484` 只擲那一顆骰,`sub_189BC`(黑棘 / 不列顛王 / 暗影領主免疫)
	// 是操控類咒語才查的。混用會讓 Xen Corp 打不動劇情人物,而原版打得動。
	if !u5data.AttackAlwaysHits(code) && s.SpellResisted(self, target) {
		// ★ 原版咒語落空印的是一句**不具名**的 `Failed!`(`sub_1F570` 的
		// `byte_3E09F != 0` 那一支),不是「某人擋下了咒語」——
		// 後者把因果講成目標的功勞,而原版沒有那個意思(`docs/re/74`)。
		s.Log(MsgFailed)
		return true
	}
	// 命中後的特例先於傷害(`sub_B9A8` 的兩個 `cmp byte_3E0AD`)。
	switch code {
	case u5data.AttackInNoxGrav:
		s.poisonHit(self, target)
		return true
	case u5data.AttackInZuGrav:
		u := &c.Units[target]
		if u.Flags&UnitAsleep == 0 {
			u.Flags |= UnitAsleep
			s.Log(s.unitName(u) + "睡著了!")
		}
		return true
	}
	dmg := s.spellAttackDamage(code)
	if dmg == u5data.InstantKillDamage {
		s.Log(s.unitName(&c.Units[target]) + "被咒語擊中!")
	}
	if dmg == 0 {
		s.Log("咒語掠過,什麼也沒發生。")
		return true
	}
	s.applyDamage(self, target, dmg)
	return true
}

// spellAttackDamage 讀攻擊咒語的傷害。
//
// **不再是估計值** —— `DATA.OVL` 0x160C 那張表其實有 56 筆,
// 第 48..54 筆就是攻擊碼 0x30..0x36 的傷害(見 `u5data.AttackCodeCount` 的說明)。
func (s *State) spellAttackDamage(code int) int {
	if s.Stats == nil || code < 0 || code >= u5data.AttackCodeCount {
		return 0
	}
	return s.Stats.ItemDamage[code]
}

// charmNearest 是 An Xen Ex:把最近的怪物變成我方(原版 `sub_194CC`,
// 印「Creature: X charmed!」)。
func (s *State) charmNearest(caster int) bool {
	c := s.Combat
	if c == nil {
		return false
	}
	self := s.combatSlotOfRoster(caster)
	if self < 0 {
		return false
	}
	target, _, _ := s.aiTarget(self)
	if target < 0 {
		return false
	}
	t := &c.Units[target]
	if t.IsParty() {
		return false
	}
	// `sub_194CC` 的三重閘門:不是劇情人物、站在敵方(`sub_29A64`)、抗性沒擋下。
	if s.resists(self, target) {
		s.Log(s.unitName(t) + "不為所動。")
		return false
	}
	t.Flags |= UnitSideFlip
	s.Log(s.unitName(t) + "被馴服了!")
	return true
}

// spellChangeFloor 是 Uus Por / Des Por。up 為真是往上。
//
// 兩個咒語的可施法場合表只放行**地牢**,所以正常情況一定走地牢那條;
// 場景那條留著是防呆(場景的樓層號與地牢相反 —— 場景 +1 是往上,
// 地牢 +1 是往下)。
func (s *State) spellChangeFloor(up bool) bool {
	if s.InDungeon() {
		d := -1
		if !up {
			d = 1
		}
		return s.DungeonChangeLevel(d)
	}
	if !s.InScene() || s.Scenes == nil {
		s.Log("此處無路可去。")
		return false
	}
	d := 1
	if !up {
		d = -1
	}
	if _, err := s.Scenes.Map(s.Location, s.Floor+d); err != nil {
		s.Log("此處無路可去。")
		return false
	}
	s.changeFloor(d)
	return true
}

// 調配咒語(原版 `sub_18704`)
//
// 玩家挑一個咒語、挑一組藥草、挑份數。**挑到的藥草一定會被扣掉**,
// 挑錯了也一樣 —— 錯的組合只會得到一團廢渣。組合要**完全相符**才算成功。

// Mix 調配 count 份 spell,用 reagents 指定的那幾種藥草。
//
// 回傳成功與否;不論成敗,選到的藥草都會被扣掉(這是原版行為)。
// ⚠ spell 可以是 `u5data.SpellInputNoSpell` —— 玩家打的符文湊不出咒語時原版
// **照樣讓他調**,藥草照扣、結果一定是廢渣(`sub_18704` 只擋 −1)。
// 所以這裡不能用 `spell < 0` 提早 return:那會把藥草還給玩家。
func (s *State) Mix(spell, count int, reagents []int) bool {
	if s.Spells == nil || spell >= u5data.SpellCount || count <= 0 {
		return false
	}
	var picked byte
	for _, r := range reagents {
		if r < 0 || r >= u5data.ReagentCount {
			continue
		}
		picked |= u5data.ReagentBit(r)
	}
	// 藥草不夠就**拒絕**,不是把份數調低。
	//
	// ⚠ 原版 `sub_18698` 印「Insufficient reagents!」然後**重問一次份數** ——
	// 它不會替玩家改成調得出來的份數。靜靜調低看起來很體貼,但玩家會以為
	// 自己調到了要的份數,而實際上少了 —— 那種落差要等到施法時才發現。
	for r := 0; r < u5data.ReagentCount; r++ {
		if picked&u5data.ReagentBit(r) == 0 {
			continue
		}
		if s.Inventory.Reagents[r] < count {
			s.Log(MsgInsufficientReagents)
			return false
		}
	}
	if picked == 0 {
		s.Log(MsgInsufficientReagents)
		return false
	}
	// ★ 先扣藥草 —— 配錯了也不退。
	for r := 0; r < u5data.ReagentCount; r++ {
		if picked&u5data.ReagentBit(r) != 0 {
			s.Inventory.Reagents[r] -= count
		}
	}
	if spell < 0 || picked != s.Spells.Spells[spell].Reagents {
		s.Log(MsgMixFailed)
		return false
	}
	n := s.Inventory.Spells[spell] + count
	if n > u5data.SpellStackLimit {
		n = u5data.SpellStackLimit
	}
	s.Inventory.Spells[spell] = n
	s.Log("完成!" + s.Spells.Spells[spell].Name + " ×" + itoa(count))
	return true
}

// MixByRecipe 用咒語自己的配方調配 count 份(玩家不用一個個挑藥草)。
func (s *State) MixByRecipe(spell, count int) bool {
	if s.Spells == nil || spell < 0 || spell >= u5data.SpellCount {
		return false
	}
	return s.Mix(spell, count, s.Spells.Spells[spell].ReagentList())
}

// AdvanceTime 推進遊戲時鐘,順便讓光源燒掉同樣的分鐘數。
//
// 原版 `sub_29304(分鐘)` 的順序:**先看時間有沒有被停住**
// (`byte_3E08A == 'T'` 就整段跳過),否則對 `byte_3E0B7`(火把)與
// `byte_3E0B6`(光明咒語)各做一次飽和減法。
//
// ⇒ **An Tym 期間火把不會燒。** 這不是小事:U5 玩家在地牢裡靠這一點省火把。
func (s *State) AdvanceTime(minutes int) {
	before := s.Clock
	if s.TimeStop > 0 {
		s.Clock.Advance(minutes)
		s.afterMidnight(before)
		return
	}
	s.Clock.Advance(minutes)
	s.LightTurns = subFloor(s.LightTurns, minutes)
	s.TorchTurns = subFloor(s.TorchTurns, minutes)
	s.tickHourly(before)
	s.afterMidnight(before)
}

// tickHourly 是**每過一個遊戲小時**才減的那些計數。
//
// 原版 `sub_29304` 把它們放在「分鐘計數器溢出 60」的那一支裡面 ——
// 與火把 / 光明咒語(每分鐘減)是不同的節奏。目前只有休息冷卻
// (`byte_3E09C`,`sub_2BBFC(&byte_3E09C, 1)`)。
//
// ⚠ 算的是**跨過幾個小時邊界**,不是「小時剛好變了」——
// 一次推進超過一小時(紮營、聖壇石室)時後者會整段跳過去。
func (s *State) tickHourly(before Clock) {
	if hours := s.Clock.HoursSince(before); hours > 0 {
		s.RestCooldown = subFloor(s.RestCooldown, hours)
	}
}

// afterMidnight 處理「跨過午夜」才發生的事(原版 `sub_29304` 在
// `byte_3E08F` 從 23 進位成 0 的那一支)。
//
// 目前只有一件:把活著的暗影君主重新分派到八德城市之一(見 shadowlord.go)。
//
// ⚠ 判斷用的是**日期變了沒**,不是「小時剛好等於 0」——
// 一次推進超過一小時(休息、進出聖壇石室)時後者會整個跳過去。
func (s *State) afterMidnight(before Clock) {
	if s.Clock.Day == before.Day && s.Clock.Month == before.Month && s.Clock.Year == before.Year {
		return
	}
	s.roamShadowlords()
}

// subFloor 是飽和減法(原版 `sub_2BBFC`)。
func subFloor(v, n int) int {
	if v -= n; v < 0 {
		return 0
	}
	return v
}

// HasLight 回報現在看不看得見(原版 `byte_3E0B6` 或 `byte_3E0B7` 任一非零)。
func (s *State) HasLight() bool { return s.LightTurns > 0 || s.TorchTurns > 0 }

// TorchMinutes 是點一把火把能撐多久(原版 `sub_17630`:`random(0,15) + 0x70`)。
func (s *State) TorchMinutes() int { return s.Roll(0, 15) + 0x70 }

// LightTorch 點一把火把。
func (s *State) LightTorch() bool {
	if s.Inventory.Torches <= 0 {
		s.Log("沒有火把了。")
		return false
	}
	s.Inventory.Torches--
	n := s.TorchMinutes()
	if n > s.TorchTurns {
		s.TorchTurns = n
	}
	s.Log("火把點亮了。")
	return true
}

// 施法的輸入流程 —— 打符文首字母,不是打咒語全名
//
// ⚠ **這裡原本做錯了**:原本讓玩家把 `In Lor` 整串打進去,而原版 `sub_1CA0C`
// 收的是**符文詞的首字母**(最多四個),每收一個就把整個符文詞回顯出來。
// 打 `I` `L` 畫面上長出 `IN LOR`,送出才查表。詳見 `u5data/runeinput.go`
// 與 `docs/re/58`。
//
// 咒語名與符文詞的 canonical 值一律維持英文(CLAUDE.md §5.2 的硬規則)——
// 玩家按的是英文字母鍵,中文只出現在說明裡。

// BeginCastPrompt 開始收符文首字母。
func (s *State) BeginCastPrompt() {
	if s.Spells == nil || s.Runes == nil {
		s.Log("咒語表未載入。")
		return
	}
	s.castBy = s.currentCaster()
	s.beginRunePrompt(MsgSpellName, func(idx int) {
		switch idx {
		case u5data.SpellInputCancelled:
			s.Log(MsgSpellNone)
		case u5data.SpellInputNoSpell:
			s.Log(MsgSpellNoEffect)
		default:
			res := s.Cast(s.castBy, idx)
			// 戰鬥中不論成敗都算用掉一個回合。
			if s.InCombat() && res != MagicNotHere {
				s.afterPlayerAction()
			}
		}
	})
}

// beginRunePrompt 開一個符文輸入(施法與調藥共用 —— 原版就是同一支 `sub_1CA0C`)。
//
// then 收到的是咒語索引,或 `u5data.SpellInputCancelled` / `SpellInputNoSpell`;
// **兩種失敗要怎麼處理由呼叫端決定** —— 施法把 −2 當「毫無效果」擋掉,
// 調藥卻讓 −2 繼續走下去(原版只擋 −1)。
func (s *State) beginRunePrompt(question string, then func(idx int)) {
	if s.Runes == nil {
		return
	}
	s.castReturn = s.Prompt
	s.runeThen = then
	s.Prompt = PromptSpell
	s.spellLetters = s.spellLetters[:0]
	s.Input = ""
	s.Log(question)
}

// TypeSpellLetter 收一個符文首字母(原版 `sub_1CA0C` 的讀鍵迴圈)。
//
// 回報這一鍵有沒有把輸入送出去 —— **空白鍵在原版就是送出**,與 Enter 同義。
func (s *State) TypeSpellLetter(r rune) bool {
	if s.Prompt != PromptSpell || s.Runes == nil {
		return false
	}
	if r == ' ' {
		s.SubmitSpell()
		return true
	}
	if r >= 'a' && r <= 'z' {
		r = r - 'a' + 'A'
	}
	// ⚠ 收不下的鍵(J / O、非字母、已經四個了)原版是**默默丟掉**,
	// 不印任何訊息 —— 照抄,不要「好心」提示。
	if len(s.spellLetters) >= u5data.RuneInputMax || !s.Runes.AcceptsLetter(byte(r)) {
		return false
	}
	s.spellLetters = append(s.spellLetters, byte(r))
	s.Input = s.spellEcho()
	return false
}

// BackspaceSpell 退掉最後一個字母(連同它回顯的那個符文詞)。
func (s *State) BackspaceSpell() {
	if s.Prompt != PromptSpell || len(s.spellLetters) == 0 {
		return
	}
	s.spellLetters = s.spellLetters[:len(s.spellLetters)-1]
	s.Input = s.spellEcho()
}

// spellEcho 是目前輸入回顯成的符文詞串。
//
// 原版每個詞後面補一個空白,而且累計超過 13 欄會先換行(`RuneInputWrapAt`)——
// 換行是主機端的欄寬行為,本引擎的訊息版面自己會折,所以這裡只組出詞串。
// 這是**版面差異**,不是機制差異。
func (s *State) spellEcho() string {
	out := ""
	for _, c := range s.spellLetters {
		w, ok := s.Runes.RuneWord(c)
		if !ok {
			continue
		}
		if out != "" {
			out += " "
		}
		out += w
	}
	return out
}

// SpellLetters 是目前打進去的符文首字母(給算繪與測試看)。
func (s *State) SpellLetters() string { return string(s.spellLetters) }

// currentCaster 是現在該由誰施法:戰鬥中是輪到的那個隊員,平時是隊長。
func (s *State) currentCaster() int {
	if c := s.Combat; c != nil && c.Turn >= 0 && c.Turn < CombatUnitSlots {
		if r := c.Units[c.Turn].Roster; r >= 0 {
			return r
		}
	}
	return 0
}

// SubmitSpell 把打好的符文送出去查表,結果交給 beginRunePrompt 給的處理函式。
func (s *State) SubmitSpell() {
	letters := s.spellLetters
	then := s.runeThen
	s.spellLetters = s.spellLetters[:0]
	s.runeThen = nil
	s.Input = ""
	s.Prompt = s.castReturn
	if s.Runes == nil || then == nil {
		return
	}
	then(s.Runes.Match(letters))
}

// CancelSpell 取消輸入(ESC)。
//
// ⚠ 原版的 ESC 走的是「當成一個字母都沒打就送出」那條路 —— 所以它跟空輸入
// **完全同一條路徑**,而不是另一個「作罷」分支。照抄:清掉字母再送出。
func (s *State) CancelSpell() {
	s.spellLetters = s.spellLetters[:0]
	s.SubmitSpell()
}

// trimSpace 去掉頭尾空白(咒語名前後可能被多打了空格)。
func trimSpace(v string) string {
	i, j := 0, len(v)
	for i < j && v[i] == ' ' {
		i++
	}
	for j > i && v[j-1] == ' ' {
		j--
	}
	return v[i:j]
}

// 全域戰鬥模式(原版 `byte_3E08A`,由 `sub_1D31C(模式, 回合, 音效)` 設定)
//
// 五個咒語共用**同一個位元組** —— 後施的會蓋掉先施的,不是一組可疊加的旗標。
// 每個模式的讀取處都在別的地方,所以語意是從「誰在讀它」反推的:
//
//	'T' An Tym       `sub_A108` 開頭直接 return  → 敵人整個不動
//	'Q' Rel Tym      `sub_A108` 擲 1/2 才行動    → 敵人只有一半的回合能動
//	'C' Quas An Wis  `sub_AC40` 可能把陣營看反   → 敵人亂打
//	'N' In An        `sub_9E10` / `sub_AE20` 擋掉 → 施法者不能放遠程、不能瞬移
//	'P' In Sanct     讀它的地方還沒找到           → 只記著,沒有效果
const (
	CombatModeNone      = 0
	CombatModeTimeStop  = 'T'
	CombatModeSlow      = 'Q'
	CombatModeConfuse   = 'C'
	CombatModeNegate    = 'N'
	CombatModeProtected = 'P'
)

// setCombatMode 設定全域戰鬥模式(原版 `sub_1D31C`)。
func (s *State) setCombatMode(mode byte, turns int) bool {
	s.CombatMode, s.CombatModeTurns = mode, turns
	if mode == CombatModeTimeStop {
		s.TimeStop = turns
	}
	return true
}

// tickCombatMode 讓模式倒數一回合(原版 `sub_16370`,在玩家單位回合結束時)。
func (s *State) tickCombatMode() {
	if s.CombatModeTurns <= 0 {
		return
	}
	s.CombatModeTurns--
	if s.CombatModeTurns == 0 {
		s.CombatMode = CombatModeNone
		s.Log("咒語的效力消退了。")
	}
}

// askWind 問方向,然後放那一道風。
func (s *State) askWind(caster, kind int) bool {
	s.AskDirection(func(d Direction) { s.castWind(caster, kind, d) })
	return true
}

// castDirection 是「沒問到方向時」的退路:離最近的敵人是哪一邊。
//
// 正常路徑已經改成真的問(AskDirection);這支留給沒有互動的呼叫端
// (headless 驗收與測試)。
func (s *State) castDirection(caster int) Direction {
	self := s.combatSlotOfRoster(caster)
	if self < 0 {
		return North
	}
	_, dx, dy := s.aiTarget(self)
	if iabs(dx) >= iabs(dy) {
		if dx >= 0 {
			return East
		}
		return West
	}
	if dy >= 0 {
		return South
	}
	return North
}

// hideCaster 是 Sanct Lor(原版 `sub_19674`)。
//
// 只做兩件事:給自己掛上隱形位元(0x10)、把地圖物件的 tile 換成 0x1D。
// 隱形位元就是 `sub_AC40` 選目標時會跳過的那一個 —— 隱形不是減傷,
// 是**根本不會被鎖定**。
func (s *State) hideCaster(caster int) bool {
	slot := s.combatSlotOfRoster(caster)
	if slot < 0 {
		return false
	}
	s.Combat.Units[slot].Flags |= UnitHidden
	s.Log(s.unitName(&s.Combat.Units[slot]) + "的身影淡去了。")
	return true
}

// energyBlast 是 In Vas Por Ylem(原版 `sub_19440`)。
//
// 對場上**每一個**敵人各判一次:擲 1..30 對上該單位的防禦,擲贏才吃
// `random(1, 20)` 的傷害,而經驗值照樣進施法者的帳。
func (s *State) energyBlast(caster int) bool {
	c := s.Combat
	if c == nil {
		s.Log("此地無敵可擊。")
		return false
	}
	self := s.combatSlotOfRoster(caster)
	if self < 0 {
		return false
	}
	hit := false
	for i := range c.Units {
		t := &c.Units[i]
		if !t.Active() || !s.hostile(t) {
			continue
		}
		def := 0
		if st := s.creatureOf(t); st != nil {
			def = int(st.Armour)
		} else if ch := s.charOf(t); ch != nil {
			def = s.Stats.DefenceOf(ch)
		}
		if s.AttackRoll() < def {
			continue
		}
		s.applyDamage(self, i, s.Roll(1, 20))
		hit = true
	}
	return hit
}

// summonCreature 是召喚類咒語(Kal Xen / In Bet Xen / Kal Xen Corp,
// 原版 `sub_18F2C` / `sub_192BC` / `sub_1CE70`)。
//
// 三支都是「挑一個空位 → `sub_2EAE4(生物索引, 0, x, y, 樓層)` 生一隻」,
// 而且都給新單位掛上 `unit[+2] |= 1`(陣營反轉 = 站我方)。
//
// **召哪一種是從通行判定那一行讀出來的**:`sub_9CE8(tile, x, y)` 的第一個
// 參數是生物的 **tile 碼**,換算回索引就是答案 ——
//
//	Kal Xen       `sub_9CE8(0x90, …)` → (144−64)/4 = 20 巨鼠
//	In Bet Xen    `sub_9CE8(0xBC, …)` → (188−64)/4 = 31 蟲群,一次 **4 隻**
//	Kal Xen Corp  `sub_9CE8(0xD8, …)` → (216−64)/4 = 38 惡魔
//
// ⚠ 這條差點抄錯:`sub_2EAE4` 的第一個參數是**生物索引**(0..47),
// 不是 tile 碼 —— 同一支函式裡兩個參數用兩套編號。
func (s *State) summonCreature(caster, creature, count int) bool {
	c := s.Combat
	if c == nil {
		s.Log("此地不可召喚。")
		return false
	}
	self := s.combatSlotOfRoster(caster)
	if self < 0 {
		return false
	}
	u := &c.Units[self]
	got := 0
	for try := 0; try < 32 && got < count; try++ {
		x, y := u.X+s.Roll(-1, 1), u.Y+s.Roll(-1, 1)
		if s.combatBlocked(self, x, y) {
			continue
		}
		for i := u5data.CombatPartySlots; i < CombatUnitSlots; i++ {
			if c.Units[i].Flags != 0 {
				continue
			}
			s.placeEnemy(c, i-u5data.CombatPartySlots,
				u5data.CreatureBase+byte(creature*4), creature)
			c.Units[i].X, c.Units[i].Y = x, y
			// 召出來的站我方 —— 怪物的陣營反轉位元代表「被馴服」。
			c.Units[i].Flags |= UnitSideFlip
			s.Log(s.unitName(&c.Units[i]) + "應召而來!")
			got++
			break
		}
	}
	return got > 0
}

// 三個召喚咒語各自召的生物索引與隻數。
const (
	summonRat    = 20 // Kal Xen
	summonInsect = 31 // In Bet Xen,一次 4 隻
	summonDaemon = 38 // Kal Xen Corp
)

// frighten 是 An Xen Corp(驅離不死)與 In Quas Corp(恐懼)。
//
// **兩個都是群體法術,不是選單一目標** —— 原版 `sub_18EB0` 與 `sub_19810`
// 都是 `for i in 0..31` 掃全場。我第一版寫成「嚇跑最近的一隻」,那是憑印象
// 補的,組語裡沒有那回事。
//
// 兩者只差一個閘門:
//
//	An Xen Corp  額外要求 `word_3F1D0[生物] & 0x20` —— 只有四種生物吃
//	             (幽靈、骷髏、惡魔、暗影領主)
//	In Quas Corp 沒有這個要求 —— 場上每個敵人都吃
//
// 其餘完全相同:跳過劇情人物(`sub_189BC`)、擲抗性(`sub_1F48C`),
// 然後 `Init = 1`(下一個就輪到它,馬上跑)+ 掛上逃跑旗標。
func (s *State) frighten(caster int, repellableOnly bool) bool {
	c := s.Combat
	if c == nil {
		s.Log("此地無人可嚇。")
		return false
	}
	self := s.combatSlotOfRoster(caster)
	if self < 0 {
		return false
	}
	any := false
	for i := 0; i < CombatUnitSlots; i++ {
		u := &c.Units[i]
		// `(BYTE2 & 0xC0) == 0x40`:是怪物、不是隊員。
		if u.Flags&(UnitMonster|UnitParty) != UnitMonster {
			continue
		}
		if repellableOnly && !s.creatureOf(u).Has(u5data.CreatureRepellable) {
			continue
		}
		if s.resists(self, i) {
			continue
		}
		u.Init = 1
		u.Flags |= UnitFleeing
		any = true
	}
	if !any {
		s.Log("沒有東西被嚇到。")
		return false
	}
	s.Log("牠們轉身逃走了!")
	return true
}

// polymorph 是 Rel Xen Bet(原版 `sub_195C0`)。
//
// 把目標從戰場上移除(`sub_B210`),在同一格生一個**種類 0x14** 的東西。
// ⚠ 0x14 那個編號的語意還沒對出來 —— 它是地圖物件的種類碼,不是生物索引
// (生物索引 0x14 = 20 是巨鼠,但這裡走的是 `sub_2EAE4(0x14, …)`
// 而該函式吃的正是生物索引,所以**很可能就是巨鼠**)。
// 兩種讀法都說得通,沒有第三個證據,所以照數值實作並在此標明。
func (s *State) polymorph(caster int) bool {
	c := s.Combat
	if c == nil {
		return false
	}
	self := s.combatSlotOfRoster(caster)
	if self < 0 {
		return false
	}
	target, _, _ := s.aiTarget(self)
	if target < 0 || c.Units[target].IsParty() {
		return false
	}
	if s.resists(self, target) {
		s.Log(s.unitName(&c.Units[target]) + "不為所動。")
		return false
	}
	t := &c.Units[target]
	x, y := t.X, t.Y
	name := s.unitName(t)
	*t = Combatant{}
	s.placeEnemy(c, target-u5data.CombatPartySlots,
		u5data.CreatureBase+byte(0x14*4), 0x14)
	c.Units[target].X, c.Units[target].Y = x, y
	s.Log(name + "變成了別的東西!")
	return true
}

// revealHidden 是 Wis Quas:讓隱形的東西現形(原版 `sub_19264` 只重畫,
// 實際效果是把隱形位元清掉)。
func (s *State) revealHidden() bool {
	c := s.Combat
	if c == nil {
		s.Log("此處沒有藏著的東西。")
		return false
	}
	any := false
	for i := range c.Units {
		if c.Units[i].Flags&UnitHidden != 0 {
			c.Units[i].Flags &^= UnitHidden
			any = true
		}
	}
	if any {
		s.Log("藏著的身影現形了。")
	}
	return any
}

// 四種「風」(原版 `sub_1AEB4(施法者, 種類, 範圍參數)`)
//
// In Zu(睡眠)、In Nox Hur(毒)、In Flam Hur(火)、In Vas Grav Corp(能量)
// 走同一支函式,流程是:
//
//  1. 問方向(`sub_1CC50`)
//  2. `sub_1AC20` 算出一串格子
//  3. 對那串格子上的每個單位各作用一次(用 `unit[+5] |= 0x80` 標記,
//     同一發不會打到同一個人兩次)
//
// ⚠ **範圍的形狀還沒逆完**。`sub_1AC20` 吃一個每個咒語各自不同的參數
// (`word_3EF44` / `word_3EF42` / `word_3EF3C` / `off_3EF3E+2`),看起來是
// 寬度或長度。這裡先用「從施法者往那個方向的一條直線,直到戰場邊緣或
// 被地形擋住」——**射線本身是照原版的投射物規則走的**(`docs/re/20`),
// 只有「寬度」是近似。文件與這裡都標明了。
const (
	windSleep  = 1 // In Zu
	windPoison = 2 // In Nox Hur
	windFire   = 3 // In Flam Hur
	windEnergy = 4 // In Vas Grav Corp

	// windEnergyDamage 是能量風的傷害:`sub_B51C(目標, 99)`,寫死不擲。
	windEnergyDamage = 99
)

// castWind 是四種風的共同實作。dir 是玩家挑的方向。
func (s *State) castWind(caster, kind int, dir Direction) bool {
	c := s.Combat
	if c == nil {
		s.Log("此地無風可興。")
		return false
	}
	self := s.combatSlotOfRoster(caster)
	if self < 0 {
		return false
	}
	u := &c.Units[self]
	dx, dy := dir.Delta()
	hitAny := false
	x, y := u.X, u.Y
	for step := 0; step < u5data.CombatSide; step++ {
		x, y = x+dx, y+dy
		if x < 0 || x >= u5data.CombatSide || y < 0 || y >= u5data.CombatSide {
			break
		}
		if u5data.TileBlocksProjectile(int(c.Map.At(x, y))) {
			break
		}
		v, ok := c.CombatUnitAt(x, y)
		if !ok {
			continue
		}
		i := s.unitIndex(v)
		if i == self {
			continue
		}
		hitAny = true
		switch kind {
		// ⚠ 四種風的抗性判定**不一樣**(`sub_1AC20` 的 switch):
		// 睡眠與能量擲 `sub_1F48C`、毒走自己那一套(`sub_B398(…,-2)`)、
		// **火完全不擲** —— 火風打到誰就是誰。
		case windSleep:
			if s.resists(self, i) {
				continue
			}
			if v.Flags&UnitAsleep == 0 {
				v.Flags |= UnitAsleep
				s.Log(s.unitName(v) + "睡著了!")
			}
		case windPoison:
			s.poisonHit(self, i)
		case windFire:
			// 原版 `sub_2B710(0x1E)` → random(0, 30)。
			s.applyDamage(self, i, s.Roll(0, 30))
		case windEnergy:
			if s.resists(self, i) {
				continue
			}
			// ★ 不是隨機 —— `sub_B51C(j, 99)` 是**寫死的 99**。
			// 99 在傷害函式裡是特別值:對隊員直接歸零(即死),
			// 對怪物就是 99 點(不死生物減半成 49)。
			s.applyDamage(self, i, windEnergyDamage)
		}
	}
	if !hitAny {
		s.Log("風掃過,什麼也沒碰到。")
	}
	return hitAny
}

// 方向選單(原版 `sub_1CC50`,印「Direction-」)
//
// 好幾個咒語與指令都要先問方向:An Ylem、An Sanct、An Grav、In Por、
// In Ex Por、An Ex Por、Rel Hur、四種風。原版每一支都是先呼叫 `sub_1CC50`
// 讀一個方向鍵,ESC 作罷。
//
// 引擎先前是用「猜一個合理的方向」代替 —— 那是介面的近似。這裡把真的選單
// 接上:`AskDirection` 進 PromptDirection,方向鍵按下去才跑後續。

// AskDirection 問一個方向,拿到之後呼叫 then。
func (s *State) AskDirection(then func(Direction)) {
	s.dirReturn = s.Prompt
	s.dirThen = then
	s.Prompt = PromptDirection
	// 指令名已經以「——」結尾了(原版的 `Get-`)就不另起一行 ——
	// 那個破折號本身就是「等方向」的提示,方向名會接在它後面。
	if !s.awaitingAfterDash() {
		s.Log("方向 ——")
	}
}

// AnswerDirection 是玩家按下方向鍵。
func (s *State) AnswerDirection(d Direction) {
	then := s.dirThen
	s.dirThen = nil
	s.Prompt = s.dirReturn
	// 原版把方向名**接在指令名後面**(`Get-North`),同一行。
	// 沒有指令名可接時才另起一行。
	if s.awaitingAfterDash() {
		s.Append(d.Name())
	} else {
		s.Log(d.Name())
	}
	if then == nil {
		return
	}
	then(d)
}

// CancelDirection 是玩家按 ESC。
func (s *State) CancelDirection() {
	s.dirThen = nil
	s.Prompt = s.dirReturn
	s.Log("作罷。")
}

// AwaitingDirection 回報是不是正在等方向。
func (s *State) AwaitingDirection() bool { return s.Prompt == PromptDirection }

// blink 是 In Por(原版 `sub_19098`)。
//
// 兩條路:
//
//	戰鬥中 → 最多試 **7 次**隨機落點(`sub_B1D8`),第一個站得住的就過去
//	地圖上 → 問方向,往那個方向直線瞬移
//
// 戰鬥中那條是**隨機**的,不是玩家指定 —— 這點很容易寫成「傳到想去的地方」。
func (s *State) blink(caster int) bool {
	if c := s.Combat; c != nil {
		self := s.combatSlotOfRoster(caster)
		if self < 0 {
			return false
		}
		u := &c.Units[self]
		for try := 0; try < 7; try++ {
			x, y := s.Roll(0, u5data.CombatSide-1), s.Roll(0, u5data.CombatSide-1)
			if s.combatBlocked(self, x, y) {
				continue
			}
			u.X, u.Y = x, y
			s.Log(s.unitName(u) + "消失又出現在別處!")
			return true
		}
		return false
	}
	// 地圖上:問方向再瞬移。
	s.AskDirection(func(d Direction) { s.blinkTowards(d) })
	return true
}

// blinkTowards 是地圖上的 In Por:往一個方向直線瞬移。
//
// ⚠ 原版的距離是 `byte_3E0AB + 0x20`(上限 0x100)—— `byte_3E0AB` 是什麼
// 還沒追到,所以這裡用固定的 0x20(32 格)並標明。
func (s *State) blinkTowards(d Direction) bool {
	dx, dy := d.Delta()
	const dist = 0x20
	for n := dist; n >= 1; n-- {
		x, y := WrapWorld(s.X+dx*n), WrapWorld(s.Y+dy*n)
		if u5data.TileBlocksWalking(int(s.TileAt(x, y))) {
			continue
		}
		s.X, s.Y = x, y
		s.Log("汝在一瞬間移動了。")
		return true
	}
	s.Log("那個方向沒有落腳處。")
	return false
}

// illusion 是 In Quas Xen(原版 `sub_196A4`)。
//
// 把目標的整筆戰場紀錄**複製**到一個空槽 —— 場上多出一個一模一樣的它。
// 原版是 `dword_3EF50[空槽] = dword_3EF50[目標]` 連同 `dword_3EF54`,
// 兩個 dword 整組搬,所以連血量與旗標都一樣。
func (s *State) illusion(caster int) bool {
	c := s.Combat
	if c == nil {
		return false
	}
	self := s.combatSlotOfRoster(caster)
	if self < 0 {
		return false
	}
	target, _, _ := s.aiTarget(self)
	if target < 0 {
		return false
	}
	for i := u5data.CombatPartySlots; i < CombatUnitSlots; i++ {
		if c.Units[i].Flags != 0 {
			continue
		}
		n := c.Units[target]
		// 找一格空地擺分身。
		placed := false
		for try := 0; try < 8 && !placed; try++ {
			x, y := n.X+s.Roll(-1, 1), n.Y+s.Roll(-1, 1)
			if s.combatBlocked(target, x, y) {
				continue
			}
			n.X, n.Y = x, y
			placed = true
		}
		if !placed {
			return false
		}
		n.Init = n.resetInit()
		c.Units[i] = n
		s.Log(s.unitName(&c.Units[i]) + "的幻影出現了!")
		return true
	}
	return false
}

// SpellResisted 是抗性判定(原版 `sub_1F48C`)。
//
// **與命中判定同一套算式**:`門檻 = (目標 + 30 − 施法者) / 2`,擲 1..30,
// 只是比較方向相反 —— 命中是 `擲 >= 門檻`,抗性是 `擲 < 門檻` 就被擋下。
// 兩邊取的都是 `sub_B398(單位, −1)`(智力那一項)。
//
// ⚠ 攻擊碼 '0'、'1' 與 >= '3' 那幾種**不擲抗性**,一律成立。
func (s *State) SpellResisted(caster, target int) bool {
	c := s.Combat
	if c == nil {
		return false
	}
	stat := func(i int) int {
		u := &c.Units[i]
		if ch := s.charOf(u); ch != nil {
			return int(ch.Intel)
		}
		if st := s.creatureOf(u); st != nil {
			return int(st.Intel)
		}
		return 0
	}
	threshold := (stat(target) + 30 - stat(caster)) / 2
	return s.AttackRoll() < threshold
}

// SpellImmune 是 `sub_189BC`:三個劇情人物擋掉全部的操控類法術。
func (s *State) SpellImmune(slot int) bool {
	if s.Combat == nil || slot < 0 || slot >= CombatUnitSlots {
		return false
	}
	return u5data.CreatureSpellImmune(s.Combat.Units[slot].Creature)
}

// resists 把「免疫」與「抗性」併成一句 —— 原版每個操控類法術都是
// `!sub_1F48C(…) && !sub_189BC(…)` 這一對,兩者任一成立就沒效果。
func (s *State) resists(caster, target int) bool {
	return s.SpellImmune(target) || s.SpellResisted(caster, target)
}

// 四個 `*Grav` 力場咒語(原版 `sub_18A08(種類)`)
//
// 分派表 `jpt_19B27` 把四個咒語送進同一支函式,只差一個種類碼:
//
//	14 In Flam Grav  → 0 → 0x82 烈焰
//	15 In Nox Grav   → 1 → 0x81 毒
//	16 In Zu Grav    → 2 → 0x80 睡眠
//	20 In Sanct Grav → 3 → 0x83 防護(踩上去什麼都不會發生)
//
// 地牢裡的作法很單純:在**面向的下一格**寫進力場編號,而且
//
//	目標格必須是 `(tile & 0xF7) == 0` —— 純通道,只有「頭上有洞」那一位元可以留
//	寫回去的是 `(舊值 & 8) | 力場編號` —— 洞保留
//	座標 `& 7` 環繞 —— 站在邊上往外放會繞到另一側
//
// 戰鬥中走的是另一條路(`sub_20360` 的效果碼 0x33..0x36),那套目標選取
// 與效果分派還沒逆完 —— 見下面的說明。
const (
	fieldFire     = 0 // In Flam Grav
	fieldPoison   = 1 // In Nox Grav
	fieldSleep    = 2 // In Zu Grav
	fieldElectric = 3 // In Sanct Grav —— 地牢裡是電擊力場(0x83)
)

// castField 放一個力場。`code` 是它在戰鬥中的攻擊碼。回傳有沒有成功。
func (s *State) castField(kind, code int) bool {
	d := s.Dungeon
	if d == nil {
		// 戰鬥中走遠程攻擊那條路(`sub_20360` 的 `byte_55E20[種類]`)。
		return s.spellAttack(s.currentCaster(), code)
	}
	dx, dy := d.Facing.Delta()
	// `and ecx, 7` —— 地牢的 8×8 是環繞的。
	x, y := (d.X+dx)&(u5data.DungeonSide-1), (d.Y+dy)&(u5data.DungeonSide-1)
	tile := s.DungeonTileAt(x, y)
	if tile&^u5data.DungeonHoleAbove != 0 {
		s.Log("那裡放不下。")
		return false
	}
	s.Dungeons.Set(d.Index, d.Level, x, y,
		(tile&u5data.DungeonHoleAbove)|u5data.DungeonFieldTile[kind])
	s.Log("一道力場在汝面前成形。")
	return true
}

// PeerRadius 是全景視野的半徑 —— 原版 `sub_EDD4` 的兩層 `for (…; < 32; …)`,
// 32×32 格,以隊伍為中心。
const PeerSide = 32

// Peer 是 In Quas Wis(原版 `sub_EDD4`):把周圍 32×32 格一次攤開,
// 畫完卡住等一個按鍵。
//
// ⚠ 地牢裡走的是另一支(`sub_F7C0`)—— 那是從所在位置向八方泛洪、
// 把整層畫成一張平面圖。引擎的地牢畫面本來就是 8×8 俯視全圖,
// 資訊量已經等同,所以這裡不另外開一個模式,直接回報「已經看得到了」。
func (s *State) Peer() bool {
	if s.InDungeon() {
		s.Log("汝已將這一層盡收眼底。")
		return true
	}
	s.Prompt = PromptPeer
	s.Log("汝的視野向四方展開。")
	return true
}

// PeerTile 取全景視野裡的一格(dx, dy 以隊伍為原點,值域 −16..15)。
func (s *State) PeerTile(dx, dy int) byte {
	return s.TileAt(s.X+dx, s.Y+dy)
}

// ClosePeer 收起全景。原版是「按任意鍵」。
func (s *State) ClosePeer() {
	if s.Prompt == PromptPeer {
		s.Prompt = PromptNone
	}
}
