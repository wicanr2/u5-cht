package game

import (
	"strconv"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 戰鬥
//
// 進戰鬥的入口是 `sub_2E58C(物件槽)`:玩家撞上地圖上的怪物物件時觸發。
// 它做四件事 ——
//
//  1. 印敵人的名字(種類 < 0x40 一律叫 "PIRATES",否則查生物名表)
//  2. 依「隊伍在不在船上 / 敵人是不是船 / 敵人腳下的地形」挑戰鬥地圖
//     (`u5data.SelectCombatMap`)
//  3. `sub_2E51C` 載入那張圖,把隊員與敵人放到圖裡的入場位置
//  4. 進戰鬥迴圈
//
// 戰場是 11×11 —— 正好等於一般遊玩的視窗大小,所以戰鬥時整個視窗就是戰場。

// Combatant 是戰場上的一個單位。
type Combatant struct {
	// Roster 是名冊索引(隊員);敵人是 -1。
	Roster int
	// Kind 是敵人的種類碼;隊員是 0。
	Kind byte
	// Tile 是畫出來的樣子。
	Tile int
	X, Y int
	// Dead 為真時不再畫、也不再行動。
	Dead bool
}

// IsParty 回報這是不是玩家這一方。
func (c *Combatant) IsParty() bool { return c.Roster >= 0 }

// Combat 是一場進行中的戰鬥。
type Combat struct {
	Map *u5data.CombatMap
	// MapIndex 是這張圖在 `.CBT` 裡的編號,方便對照 `u5dump cbt` 的輸出。
	MapIndex int
	// Units 是場上所有單位,隊員在前、敵人在後。
	Units []Combatant
	// Turn 是輪到誰行動(Units 的索引)。
	Turn int
	// EnemyName 是開場印的那個名字。
	EnemyName string
	// fromSlot 是觸發戰鬥的那個物件槽,打完要清掉。
	fromSlot int
}

// InCombat 回報是不是正在戰鬥。
func (s *State) InCombat() bool { return s.Combat != nil }

// enemyDisplayName 照 `sub_2E58C`:種類碼 < 0x40 的一律叫 PIRATES,
// 其餘查生物名表(索引 = (種類 − 64) / 4,與 docs/re/09 同一條公式)。
func (s *State) enemyDisplayName(kind byte) string {
	if kind < u5data.CreatureBase {
		return "PIRATES"
	}
	if n := s.Creatures.Name(kind); n != "" {
		return n
	}
	return "#" + strconv.Itoa(int(kind))
}

// BeginCombat 把玩家帶進戰鬥(原版 `sub_2E58C`)。
//
// slot 是撞上的那個怪物物件的槽號。回傳 false 代表打不起來
// (缺戰鬥地圖資料時誠實拒絕,而不是生一場空的仗)。
func (s *State) BeginCombat(slot int) bool {
	objs := s.currentObjects()
	if objs == nil || s.CombatMaps == nil {
		return false
	}
	o := &objs.Objects[slot]
	if !o.Present() {
		return false
	}
	kind := o.Kind &^ 0x03
	terrain := int(s.TileAt(o.X, o.Y))
	idx := u5data.SelectCombatMap(int(kind), terrain, s.Transport, !s.InScene())
	if idx < 0 || idx >= len(s.CombatMaps.Maps) {
		return false
	}
	m := &s.CombatMaps.Maps[idx]

	c := &Combat{Map: m, MapIndex: idx, fromSlot: slot, EnemyName: s.enemyDisplayName(o.Kind)}
	// 隊員照圖裡的入場位置排;人數不足就只排前 n 個。
	for i, ch := range s.Party() {
		if i >= u5data.CombatPartySlots {
			break
		}
		c.Units = append(c.Units, Combatant{
			Roster: i,
			Tile:   u5data.NPCTileBase + int(partyTileFor(ch)),
			X:      int(m.PartyX[i]),
			Y:      int(m.PartyY[i]),
			Dead:   ch.Status == u5data.StatusDead,
		})
	}
	// 敵人:原版一次擺幾隻由遭遇決定,那部分還沒逆完 —— 這裡先放一隻,
	// 位置用圖裡的第 0 個敵人入場點。多隻遭遇見 §還沒做的。
	c.Units = append(c.Units, Combatant{
		Roster: -1,
		Kind:   o.Kind,
		Tile:   u5data.NPCTileBase + int(o.Kind),
		X:      int(m.EnemyX[0]),
		Y:      int(m.EnemyY[0]),
	})

	s.Combat = c
	s.Prompt = PromptCombat
	s.Log("「" + c.EnemyName + "」來襲!")
	s.Log("(戰場 #" + strconv.Itoa(idx) + ")")
	return true
}

// partyTileFor 挑隊員在戰場上的圖。
//
// 原版用角色的職業決定 sprite;職業碼與生物編號的對應還沒追到,
// 先一律用聖者那一組 —— 這是**顯示**的近似,不影響規則。
func partyTileFor(_ *u5data.Character) byte { return 0x4C }

// CombatTileAt 回傳戰場上 (x, y) 該顯示什麼。
func (s *State) CombatTileAt(x, y int) byte {
	if s.Combat == nil {
		return u5data.TileBlank
	}
	return s.Combat.Map.At(x, y)
}

// CombatUnitAt 回報戰場上某一格站著誰。
func (c *Combat) CombatUnitAt(x, y int) (*Combatant, bool) {
	for i := range c.Units {
		u := &c.Units[i]
		if !u.Dead && u.X == x && u.Y == y {
			return u, true
		}
	}
	return nil, false
}

// EndCombat 結束戰鬥回到地圖。won 為真時把觸發戰鬥的怪物從地圖上清掉。
func (s *State) EndCombat(won bool) {
	if s.Combat == nil {
		return
	}
	if won {
		if objs := s.currentObjects(); objs != nil {
			objs.Remove(s.Combat.fromSlot)
		}
	}
	s.Combat = nil
	s.Prompt = PromptNone
}

// CombatMove 讓目前輪到的單位往 d 走一步。
//
// 通行用**玩家那張表**(戰場上的隊員是玩家),而且不能疊在別人身上。
func (s *State) CombatMove(d Direction) {
	c := s.Combat
	if c == nil || c.Turn < 0 || c.Turn >= len(c.Units) {
		return
	}
	u := &c.Units[c.Turn]
	dx, dy := d.Delta()
	nx, ny := u.X+dx, u.Y+dy
	if nx < 0 || nx >= u5data.CombatSide || ny < 0 || ny >= u5data.CombatSide {
		s.Log(MsgBlocked)
		return
	}
	if u5data.TileBlocksWalking(int(c.Map.At(nx, ny))) {
		s.Log(MsgBlocked)
		return
	}
	if _, taken := c.CombatUnitAt(nx, ny); taken {
		s.Log(MsgBlocked)
		return
	}
	u.X, u.Y = nx, ny
	s.nextCombatTurn()
}

// nextCombatTurn 換下一個還活著的單位行動。
func (s *State) nextCombatTurn() {
	c := s.Combat
	for i := 1; i <= len(c.Units); i++ {
		n := (c.Turn + i) % len(c.Units)
		if !c.Units[n].Dead {
			c.Turn = n
			return
		}
	}
}

// CombatFlee 逃離戰鬥。原版的逃跑要走到戰場邊緣才行,判定還沒逆完 ——
// 這裡先讓玩家隨時能離開,並誠實說明。
func (s *State) CombatFlee() {
	if s.Combat == nil {
		return
	}
	s.Log("汝撤離了戰場。(逃跑判定尚未實作 —— 原版要走到戰場邊緣)")
	s.EndCombat(false)
}

// 命中判定
//
// 原版 `sub_B484`:
//
//	threshold = (目標的防禦 + 30 − 攻擊者的攻擊) / 2
//	roll      = max(1, random(0, 60) / 2)        ← sub_2B724,值域 1..30
//	命中 ⟺ roll >= threshold
//
// 防禦高 → threshold 高 → 難命中;攻擊高 → threshold 低 → 易命中。
//
// 「攻擊」取哪一項由武器決定(`sub_B398`):**鈍器**(類別 8:釘盔、釘盾、
// 棍、釘頭錘、雙手錘)看**力量**,其餘看裝備的防禦加總。
// 幾種武器(0x23、0x27、0x28)則**必中**,不擲骰。

// AlwaysHitWeapons 是不用擲骰的武器編號(原版 `sub_B484` 的三個 cmp)。
var AlwaysHitWeapons = map[byte]bool{0x23: true, 0x27: true, 0x28: true}

// AttackRoll 是命中骰,值域 1..30(原版 `sub_2B724`)。
//
//	r = random(0, 60) / 2      ← 閉區間,所以是 0..30
//	if r == 0 { r = 1 }
func (s *State) AttackRoll() int {
	r := s.Roll(0, 60) / 2
	if r == 0 {
		r = 1
	}
	return r
}

// attackValueOf 取一名角色的「攻擊」那一項(原版 `sub_B398` 對角色的分支)。
func (s *State) attackValueOf(c *u5data.Character, weapon byte) int {
	if s.Stats == nil || c == nil {
		return 0
	}
	if s.Stats.IsBlunt(weapon) {
		return int(c.Strength)
	}
	return s.Stats.DefenceOf(c)
}

// HitChance 回報這一擊的門檻值(threshold)。roll >= 門檻就命中。
//
// 門檻 ≤ 1 代表必中(骰子最小是 1),≥ 31 代表必不中(骰子最大是 30)。
func HitThreshold(defence, attack int) int {
	return (defence + 30 - attack) / 2
}

// AttackerHits 判定一名隊員用某件武器打某個防禦值的目標會不會命中。
func (s *State) AttackerHits(c *u5data.Character, weapon byte, targetDefence int) bool {
	if AlwaysHitWeapons[weapon] {
		return true
	}
	return s.AttackRoll() >= HitThreshold(targetDefence, s.attackValueOf(c, weapon))
}

// 傷害計算
//
// 原版 `sub_B274(攻擊者, 目標)`:
//
//	攻擊者是怪物 → base = 怪物屬性的 +4(攻擊力)
//	攻擊者是角色 → 依武器:
//	    玻璃劍(0x27)  → 99,而且**劍會碎掉**(印「Thy sword hath shattered!」)
//	    寶石劍(0x28)  → 0
//	    空手(0xFF)    → 1
//	    其餘           → base = 武器傷害表[武器]
//	                     base > 1 且 != 99 時擲 random(1, base)
//
//	base == 99 → 傷害 99(必殺,不扣防禦)
//	否則:
//	    目標是怪物 → 減 random(1, 怪物屬性的 +3)
//	    目標是角色 → 減 random(1, 角色紀錄的 0x18)
//	    減值為 0 時不扣
//
// 注意兩處**閉區間**的 random(1, n) —— 與命中骰同一個 `sub_28E14`。

// DamageResult 是一次傷害計算的結果。
type DamageResult struct {
	// Amount 是最後扣的血;可能是負數(防禦擲得比傷害高),原版不夾。
	Amount int
	// Shattered 為真時武器碎了(玻璃劍)。
	Shattered bool
}

// WeaponDamageRoll 算攻擊者這一擊的基礎傷害(還沒扣目標的防禦)。
func (s *State) WeaponDamageRoll(weapon byte) (base int, shattered bool) {
	switch weapon {
	case u5data.ItemGlassSword:
		return u5data.InstantKillDamage, true
	case u5data.ItemJeweledSword:
		return 0, false
	case u5data.ItemNone:
		return u5data.BareHandDamage, false
	}
	if s.Stats == nil || int(weapon) >= u5data.ItemCount {
		return 0, false
	}
	base = s.Stats.ItemDamage[weapon]
	if base > 1 && base != u5data.InstantKillDamage {
		base = s.Roll(1, base)
	}
	return base, false
}

// DamageToCharacter 算一名隊員被打時實際掉多少血。
func (s *State) DamageToCharacter(base int, c *u5data.Character) int {
	return s.applyResist(base, int(c.Raw[u5data.CharDamageResist]))
}

// DamageToCreature 算一隻怪物被打時實際掉多少血。
func (s *State) DamageToCreature(base int, creature byte) int {
	resist := 0
	if s.Stats != nil {
		if st, ok := s.Stats.StatsFor(creature); ok {
			resist = int(st.Armour)
		}
	}
	return s.applyResist(base, resist)
}

// applyResist 套用減傷。必殺(99)不扣;減傷值 0 也不擲骰。
func (s *State) applyResist(base, resist int) int {
	if base == u5data.InstantKillDamage {
		return u5data.InstantKillDamage
	}
	if resist == 0 {
		return base
	}
	return base - s.Roll(1, resist)
}
