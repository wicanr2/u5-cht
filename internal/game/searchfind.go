package game

import (
	"strconv"

	"github.com/wicanr2/u5-cht/internal/i18n"
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 搜尋的三層「真的找到東西」(原版 `sub_147A8` 的尾段,推導見 `docs/re/96`)
//
// 引擎此前只有隨機那一層(`sub_13DD8` 的金幣 / 食物 / 垃圾 / 瘟疫),
// 而原版在它之後還有**三層固定內容**,順序是:
//
//	sub_13FC4(X, Y, 樓層)  → 挖出**自己埋的月石**
//	sub_14090(X, Y)        → **午夜**的秘密採藥點
//	sub_14160(X, Y)        → 113 筆**固定物品**;都沒中才印「什麼也沒有」
//
// ★ 三層是**短路**的:前一層回傳非 0 就不跑後面。
// ⚠ 而「什麼也沒有」那一句只在**第三層**沒中時才印 —— 前兩層落空是靜默的
//(它們會落到下一層)。

// searchFixedContents 依序跑那三層(原版 `sub_147A8` 尾段)。
//
// 回傳 false 表示三層都沒中 —— 呼叫端該印「什麼也沒有」。
func (s *State) searchFixedContents(x, y int) bool {
	if s.digUpMoonstone(x, y) {
		return true
	}
	if s.gatherHerbs(x, y) {
		return true
	}
	return s.findHiddenItem(x, y)
}

// digUpMoonstone 把埋在這一格的月石挖出來(原版 `sub_13FC4`)。
//
// ★ **倒著掃八顆**(原版 `esi = 8; dec esi; ... jl`)—— 同一格埋了兩顆時
// 先挖出**編號大**的那一顆。
//
// ⚠ 四個欄位**全部**要對上:X / Y / 樓層 / **地點碼**。少比地點碼的話
// 「埋在城裡 (15,15)」與「埋在大地圖 (15,15)」會混在一起。
func (s *State) digUpMoonstone(x, y int) bool {
	for i := u5data.MoonstoneCount - 1; i >= 0; i-- {
		m := s.Inventory.Moonstones[i]
		if m.X != x || m.Y != y || m.Floor != s.Floor || m.Location != s.locationCode() {
			continue
		}
		// ★ 那一格已經有這一顆的物件了就跳過 —— 否則每搜一次就多生一顆。
		if s.moonstoneObjectHere(x, y, byte(i)) {
			continue
		}
		slot := s.freeObjectSlot()
		if slot < 0 {
			s.Log(MsgNoRoom)
			return true
		}
		// ★ 品質欄放的是**月石編號** —— 撿起來才知道是第幾顆。
		s.placeFoundObjectAt(slot, u5data.ItemMoonstone, byte(i), x, y)
		s.Log(MsgStrangeRock)
		return true
	}
	return false
}

// moonstoneObjectHere 回報 (x, y) 上是不是已經躺著第 n 顆月石。
func (s *State) moonstoneObjectHere(x, y int, n byte) bool {
	objs := s.CurrentObjects()
	if objs == nil {
		return false
	}
	o, ok := objs.At(x, y, s.Floor)
	if !ok {
		return false
	}
	return o.Raw[u5data.ObjKind] == u5data.ItemMoonstone &&
		o.Raw[u5data.ObjQuality] == n
}

// 採藥的規則(原版 `sub_14090`)。
const (
	// HerbGatherHour 是唯一採得到的小時(原版 `cmp byte_3E08F, 0`)。
	//
	// ★ **午夜**。NPC 對話裡就講了「曼德拉草可於彼或亡者沼地採得」,
	// 而少了時間這一條的話玩家白天去會什麼都沒有卻不知道為什麼。
	HerbGatherHour = 0
	// 一次採到的株數(原版 `random(2, 15)`)。
	HerbGatherMin = 2
	HerbGatherMax = 15
	// HerbCarryLimit 是每種藥草的上限(原版 `cmp …, 63h` → 夾到 0x63)。
	HerbCarryLimit = 0x63
)

// gatherHerbs 是秘密採藥點(原版 `sub_14090`)。
//
// 四個條件都要成立:座標對上其中一點、**小時是 0**、**今天還沒採過這一點**、
// 而「今天」是靠 `byte_3E068[i] == byte_3E08E`(日期)判的。
//
// ⚠ 記帳是**每點各一份**,不是全域一份 —— 同一個午夜可以跑三個點各採一次。
func (s *State) gatherHerbs(x, y int) bool {
	for i := range u5data.HerbSpots {
		spot := u5data.HerbSpots[i]
		if int(spot.X) != x || int(spot.Y) != y {
			continue
		}
		if s.Clock.Hour != HerbGatherHour {
			continue
		}
		if s.herbPickedDay[i] == byte(s.Clock.Day) {
			continue
		}
		s.herbPickedDay[i] = byte(s.Clock.Day)
		n := s.Roll(HerbGatherMin, HerbGatherMax)
		h := int(spot.Herb)
		if h >= 0 && h < len(s.Inventory.Reagents) {
			s.Inventory.Reagents[h] += n
			if s.Inventory.Reagents[h] > HerbCarryLimit {
				s.Inventory.Reagents[h] = HerbCarryLimit
			}
		}
		s.Log(strconv.Itoa(n) + " 株" + i18n.Name(spot.Name))
		return true
	}
	return false
}

// findHiddenItem 是那 113 筆固定物品(原版 `sub_14160`)。
//
// 五個條件:地點碼、樓層、X、Y 都對上,而且**還沒被撿過**。
// 撿過的記在一張位元圖裡(原版 `byte_3E06C[i>>3]` 的 `1 << (i&7)`)。
//
// ★★ **索引 13..15 不記帳**(原版 `if (i < 0x0D || i > 0x0F)` 才寫位元圖)
// ⇒ 那三筆**可以重複拿**。其中第 13 筆是一串鑰匙,而它另外要求
// **手上一把鑰匙都沒有** —— 那是防卡死:鑰匙用完了才會再長出來。
func (s *State) findHiddenItem(x, y int) bool {
	for i := 0; i < u5data.HiddenItemCount; i++ {
		it := u5data.HiddenItems[i]
		if int(it.Loc) != s.locationCode() || int(int8(it.Floor)) != s.Floor {
			continue
		}
		if int(it.X) != x || int(it.Y) != y {
			continue
		}
		if !s.hiddenItemAvailable(i) {
			continue
		}
		slot := s.freeObjectSlot()
		if slot < 0 {
			s.Log(MsgNoRoom)
			return true
		}
		if !u5data.HiddenItemRepeatable(i) {
			s.markHiddenItemTaken(i)
		}
		s.placeFoundObjectAt(slot, it.Kind, it.Quality, x, y)
		s.Log(i18n.Look(0, u5data.DungeonRoomItemName(it.Kind)))
		return true
	}
	return false
}

// hiddenItemAvailable 回報第 i 筆還拿不拿得到。
//
// ⚠ 原版對索引 13 / 14 / 15 各有一條額外條件;13 那條(鑰匙用完才長)
// 已還原,14 與 15 的兩個位元組語意未解 ⇒ **先當成永遠可拿**
// (`CLAUDE.md §3.0`:留白比猜著寫好,而這個留白的方向是「玩家拿得到」)。
func (s *State) hiddenItemAvailable(i int) bool {
	if s.hiddenTaken[i>>3]&(1<<(i&7)) != 0 {
		return false
	}
	if i == u5data.HiddenItemSpareKeys && s.Inventory.Keys > 0 {
		return false
	}
	return true
}

// markHiddenItemTaken 記下第 i 筆已經被撿走(原版 `byte_3E06C[i>>3] |= bit`)。
func (s *State) markHiddenItemTaken(i int) {
	s.hiddenTaken[i>>3] |= 1 << (i & 7)
}
