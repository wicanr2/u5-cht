package game

import "github.com/wicanr2/u5-cht/internal/u5data"

// 戰鬥中站在有害的格子上(原版 `sub_BBA0`,由 `sub_A9EC` 每個單位回合呼叫)
//
// 推導見 `docs/re/99`。這是**戰鬥地圖版**的 `terrainEffects` ——
// 引擎此前只有場景 / 大地圖那一份(`sub_1318`),而戰場上熔岩、壁爐、沼澤
// 與三種力場**一件都不作用**:把敵人推到熔岩上完全沒事。
//
// 原版分成兩層找:**先看地形**,地形沒中才**掃同一格的物件**。
//
//	地形 0x8F 熔岩 / 0xBC 壁爐  → 100(傷害)
//	地形 0x04 沼澤              →  50(毒)
//	物件 0xEA 火力場            → 100
//	物件 0xE8 毒力場            →  50
//	物件 0xE9 睡眠力場          → 150(睡著)
//
// ★ 那三個數字(50 / 100 / 150)在原版就是 `switch` 的 case 值 ——
// 不是機率也不是傷害量,只是「哪一種效果」的代碼。照原樣命名。

// 戰場上有害格子的效果代碼(原版 `sub_BBA0` 的 `esi`)。
const (
	// HarmNone 是沒有效果。
	HarmNone = 0
	// HarmPoison 是毒(沼澤與毒力場)。
	HarmPoison = 50
	// HarmBurn 是燒傷(熔岩、壁爐、火力場)。
	HarmBurn = 100
	// HarmSleep 是睡著(睡眠力場)。
	HarmSleep = 150
)

// 有害的地形與物件。
const (
	// TileMoltenLava 是熔岩(`look#143`)。
	TileMoltenLava = 0x8F
	// TileFireplaceCombat 是壁爐(`look#188`)。
	//
	// ⚠ 與 `upkeep.go` 的 `TileFireplace` 是**同一個 tile**,而兩支各自
	// 從自己的組語抄出來。不合併是因為兩邊的效果不同(這裡是單一單位、
	// 那裡是全隊),合併之後很容易把兩套規則混成一套。
	TileFireplaceCombat = 0xBC
	// TileSwampCombat 是沼澤(`look#4`)。
	TileSwampCombat = 0x04

	// FieldFire / FieldPoisonObj / FieldSleepObj 是三種會作用的力場物件。
	//
	// ⚠ 第四種(0xEB,純力場)**沒有** case —— 站在上面什麼都不會發生。
	// 那是原版的樣子,不要「補齊」。
	FieldPoisonObj = 0xE8
	FieldSleepObj  = 0xE9
	FieldFire      = 0xEA

	// FieldBurnDamageMax 是燒傷的擲骰上限(原版 `sub_2B710(10)` = `random(0,10)`)。
	FieldBurnDamageMax = 10
	// FieldPoisonDamageMax 是「毒不上身時改扣的血」(原版 `sub_B8DC` 的 `random(0,20)`)。
	FieldPoisonDamageMax = 20
	// PoisonImmuneKindFrom 是「這種怪物不吃毒」的種類碼下界(原版 `kind < 80h`)。
	PoisonImmuneKindFrom = 0x80
)

// harmUnderUnit 回報某個單位腳下該吃哪一種效果(原版 `sub_BBA0` 的前半)。
//
// ⚠ **順序不能反**:地形先、物件後。地形中了就**不掃物件** ——
// 所以站在熔岩上的毒力場只會燒不會毒。
func (s *State) harmUnderUnit(slot int) int {
	c := s.Combat
	if c == nil || slot < 0 || slot >= len(c.Units) {
		return HarmNone
	}
	u := &c.Units[slot]
	switch s.TileAt(u.X, u.Y) {
	case TileMoltenLava, TileFireplaceCombat:
		return HarmBurn
	case TileSwampCombat:
		return HarmPoison
	}
	objs := s.currentObjects()
	if objs == nil {
		return HarmNone
	}
	for i := range objs.Objects {
		o := &objs.Objects[i]
		if !o.Present() || o.X != u.X || o.Y != u.Y {
			continue
		}
		switch o.Kind {
		case FieldFire:
			return HarmBurn
		case FieldPoisonObj:
			return HarmPoison
		case FieldSleepObj:
			return HarmSleep
		}
	}
	return HarmNone
}

// harmStandingUnit 套用腳下那一格的效果(原版 `sub_BBA0` 的 switch)。
func (s *State) harmStandingUnit(slot int) {
	c := s.Combat
	if c == nil || slot < 0 || slot >= len(c.Units) {
		return
	}
	switch s.harmUnderUnit(slot) {
	case HarmPoison:
		// ⚠ 種類碼 ≥ 0x80 的怪物**不吃這一條**(原版
		// `if (物件[自己].kind < 80h)`)。隊員的物件是隊伍那一格,種類碼很小。
		if u := &c.Units[slot]; !u.IsParty() && u.Kind >= PoisonImmuneKindFrom {
			return
		}
		s.poisonOrHurt(slot)
	case HarmBurn:
		s.Log(MsgBurning)
		s.applyDamage(slot, slot, s.Roll(0, FieldBurnDamageMax))
	case HarmSleep:
		s.putUnitToSleep(slot)
	}
}

// poisonOrHurt 是毒的兩條路(原版 `sub_B8DC`)。
//
//	是隊員 && 狀態 == 'G'  → 狀態改 'P',印「<名字> is poisoned!」
//	否則                   → 改成扣 random(0, 20) 血
//
// ★★ 第二條容易漏,而它的後果很具體:**已經中毒的人再踩毒力場會受傷**
// (狀態已經不是 'G'),而**怪物踩毒力場也是受傷**(怪物沒有狀態欄)。
// 只寫第一條的話,毒力場對敵人完全無效 —— 而那是玩家最常用它的方式。
func (s *State) poisonOrHurt(slot int) {
	c := s.Combat
	u := &c.Units[slot]
	if u.IsParty() && u.Roster >= 0 && u.Roster < len(s.Roster) {
		if ch := &s.Roster[u.Roster]; ch.Status == u5data.StatusGood {
			ch.Status = u5data.StatusPoisoned
			ch.Raw[u5data.CharStatus] = u5data.StatusPoisoned
			s.Log(s.charName(ch) + MsgIsPoisoned)
			return
		}
	}
	s.applyDamage(slot, slot, s.Roll(0, FieldPoisonDamageMax))
}

// putUnitToSleep 讓一個單位睡著(原版 `sub_2EDF8`)。
//
//	怪物         → 掛上「退場」位元,從地圖上移掉
//	隊員(非 'D')→ 狀態改 'S';如果他是被指定的行動者,取消指定
//
// ★★ 而「狀態改 'S'」會把原本的狀態**擦掉**(狀態是單一位元組)——
// 所以睡眠力場**會解掉中毒與魅惑**。與紮營同一個副作用(`holeup.go`)。
func (s *State) putUnitToSleep(slot int) {
	c := s.Combat
	u := &c.Units[slot]
	if !u.IsParty() {
		u.Flags |= UnitAsleep
		s.Log(s.unitName(u) + MsgFallsAsleep)
		return
	}
	if u.Roster < 0 || u.Roster >= len(s.Roster) {
		return
	}
	ch := &s.Roster[u.Roster]
	if ch.Status == u5data.StatusDead {
		return
	}
	ch.Status = u5data.StatusAsleep
	ch.Raw[u5data.CharStatus] = u5data.StatusAsleep
	u.Flags |= UnitAsleep
	// 睡著的人不能當「指定的行動者」—— 原版 `if (單位 == 0) byte_3E08B = 0FFh`。
	if s.ActiveMember() == u.Roster {
		s.ClearActiveMember()
	}
	s.Log(s.charName(ch) + MsgFallsAsleep)
}
