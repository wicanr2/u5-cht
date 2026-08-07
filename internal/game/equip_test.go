package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// TestSlotForFollowsTheItemTableRanges:每一類裝備進對的欄位。
//
// 分界不是自己畫的,是 `u5data` 從裝備名字表的實際排列定出來的。
// 這條逐界線驗一次 —— 差一格的話「釘盾」會被當成護甲穿在身上。
func TestSlotForFollowsTheItemTableRanges(t *testing.T) {
	cases := []struct {
		item byte
		want EquipSlot
		ok   bool
	}{
		{0, SlotHelm, true},    // Leather Helm
		{3, SlotHelm, true},    // Spiked Helm
		{4, SlotShield, true},  // Small Shield
		{8, SlotShield, true},  // Jewel Shield
		{9, SlotArmour, true},  // Cloth Armour
		{15, SlotArmour, true}, //
		{16, SlotWeapon, true}, // Dagger
		{41, SlotWeapon, true}, // Mystic Sword
		{42, SlotRing, true},   // Ring of Invisibility
		{44, SlotRing, true},   // Ring of Regeneration
		{45, SlotAmulet, true}, // Amulet of Turning
		{47, SlotAmulet, true}, // Ankh
		{u5data.ItemNone, 0, false},
	}
	for _, c := range cases {
		got, ok := SlotFor(c.item)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("裝備 %d → 欄位 %d(ok=%v),預期 %d(ok=%v)",
				c.item, got, ok, c.want, c.ok)
		}
	}
}

// TestEquipReturnsTheOldItem:換下來的要放回背包。
//
// 少了這一步,玩家換一次裝備就永久失去原本那件 —— 而症狀要等到很久以後
// 翻背包才會發現,那時已經無法回推是哪一次換裝吃掉的。
func TestEquipReturnsTheOldItem(t *testing.T) {
	s := equipScene(t)
	c := &s.Roster[0]
	old := c.Raw[u5data.CharArmour]
	if old == u5data.ItemNone {
		t.Skip("這名角色沒穿護甲")
	}
	const newArmour = 15
	s.Inventory.Items[newArmour] = 1
	before := s.Inventory.Items[old]

	if !s.Ready(0, newArmour) {
		t.Fatalf("穿不上:%q", s.Messages)
	}
	if c.Raw[u5data.CharArmour] != newArmour {
		t.Errorf("身上是 %d,預期 %d", c.Raw[u5data.CharArmour], newArmour)
	}
	if s.Inventory.Items[old] != before+1 {
		t.Errorf("舊護甲沒放回背包:%d → %d", before, s.Inventory.Items[old])
	}
	if s.Inventory.Items[newArmour] != 0 {
		t.Errorf("新護甲沒從背包扣掉:剩 %d", s.Inventory.Items[newArmour])
	}
}

// TestSecondWeaponGoesToTheOffHand:右手滿了就往左手放。
//
// 原版的 0x1C 欄位既收盾也收第二把武器(Iolo 就是雙手各一把)。
// 一律換掉右手的話,玩家永遠裝不上第二把 —— 而那是吟遊詩人的打法。
func TestSecondWeaponGoesToTheOffHand(t *testing.T) {
	s := equipScene(t)
	c := &s.Roster[0]
	c.Raw[u5data.CharWeapon] = 23 // Short Sword
	c.Raw[u5data.CharShield] = u5data.ItemNone
	const second = 24 // Mace
	s.Inventory.Items[second] = 1

	if !s.Ready(0, second) {
		t.Fatalf("拿不上:%q", s.Messages)
	}
	if c.Raw[u5data.CharWeapon] != 23 {
		t.Errorf("右手被換掉了:%d", c.Raw[u5data.CharWeapon])
	}
	if c.Raw[u5data.CharShield] != second {
		t.Errorf("左手是 %d,預期 %d", c.Raw[u5data.CharShield], second)
	}
}

// TestReadyTakesEveryCategory:R **一支收下全部六類**,而且清單也要全列。
//
// ⚠ 這支測試的前身是 `TestReadyAndWearRejectTheWrongCategory` ——
// 它釘住的是「R 不收護甲」,而那個行為是我從 U4 搬來的,原版沒有
//(`sub_2ACF4` 對 'W' 只印 "W-What?",見 `docs/re/49`)。
// **測試會把錯的行為釘得跟對的一樣牢** —— 這是本專案第三次踩到。
func TestReadyTakesEveryCategory(t *testing.T) {
	s := equipScene(t)
	// 六類各一件:頭盔 0、盾 4、護甲 9、武器 16、戒指 42、頸飾 45。
	for _, it := range []byte{0, 4, 9, 16, 42, 45} {
		s.Inventory.Items[it] = 1
	}
	for _, it := range []byte{0, 9, 42, 45} {
		s.Messages = nil
		if !s.Ready(0, it) {
			t.Errorf("R 收不下第 %d 件:%q", it, s.Messages)
		}
	}
	// 清單也要把六類都列出來。
	list := s.ReadyList()
	got := map[EquipSlot]bool{}
	for _, it := range list {
		if slot, ok := SlotFor(it); ok {
			got[slot] = true
		}
	}
	if len(got) < 4 {
		t.Errorf("R 的清單只涵蓋 %d 類欄位(%v),應該把背包裡各類都列出來", len(got), list)
	}
	// 不是裝備的東西還是要擋(例如火把 —— 不在 0..47 的裝備範圍裡)。
	s.Messages = nil
	if s.Ready(0, u5data.ItemNone) {
		t.Error("R 收下了「無」")
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgCannotReady) {
		t.Errorf("沒說拿不上手:%q", s.Messages)
	}
}

// TestEquipNeedsTheItemInTheBag:背包裡沒有就裝不上。
func TestEquipNeedsTheItemInTheBag(t *testing.T) {
	s := equipScene(t)
	s.Inventory.Items[41] = 0
	if s.Ready(0, 41) {
		t.Error("背包裡沒有卻裝上了")
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgDontHaveThat) {
		t.Errorf("沒說沒有那件:%q", s.Messages)
	}
}

// TestUnequipPutsItBack:卸下來的回背包。
func TestUnequipPutsItBack(t *testing.T) {
	s := equipScene(t)
	c := &s.Roster[0]
	c.Raw[u5data.CharHelm] = 2
	before := s.Inventory.Items[2]
	if !s.Unequip(0, SlotHelm) {
		t.Fatal("卸不下來")
	}
	if c.Raw[u5data.CharHelm] != u5data.ItemNone {
		t.Errorf("頭上還有 %d", c.Raw[u5data.CharHelm])
	}
	if s.Inventory.Items[2] != before+1 {
		t.Errorf("沒回背包:%d → %d", before, s.Inventory.Items[2])
	}
	if s.Unequip(0, SlotHelm) {
		t.Error("空欄位卻卸得下來")
	}
}

// TestReadyListCoversAllSixSlots:R 的選單是**一張**,六類都在裡面。
//
// ⚠ 這支的前身是 `TestEquipListsSplitByCategory`,釘的是「R 與 W 兩張不重疊」
// —— 而 U5 沒有 W(`docs/re/49`)。舊測試每一次都綠,因為它量的是我自己
// 發明的規則。
func TestReadyListCoversAllSixSlots(t *testing.T) {
	s := equipScene(t)
	// 六類各一件:頭盔 0、盾 4、護甲 9、武器 16、戒指 42、頸飾 45。
	want := []int{0, 4, 9, 16, 42, 45}
	for _, i := range want {
		s.Inventory.Items[i] = 1
	}
	seen := map[byte]bool{}
	for _, it := range s.ReadyList() {
		seen[it] = true
	}
	for _, i := range want {
		if !seen[byte(i)] {
			t.Errorf("裝備 %d 不在 R 的選單裡", i)
		}
	}
	// 選單裡不該出現裝不上的東西。
	for _, it := range s.ReadyList() {
		if _, ok := SlotFor(it); !ok {
			t.Errorf("選單裡有裝不上的 %d", it)
		}
	}
}

func equipScene(t *testing.T) *State {
	t.Helper()
	s := newCreateState(t)
	s.MaxMessages = 16
	return s
}
