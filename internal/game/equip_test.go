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

	if !s.Wear(0, newArmour) {
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

// TestReadyAndWearRejectTheWrongCategory:R 不收護甲、W 不收武器。
func TestReadyAndWearRejectTheWrongCategory(t *testing.T) {
	s := equipScene(t)
	s.Inventory.Items[9] = 1  // Cloth Armour
	s.Inventory.Items[16] = 1 // Dagger

	if s.Ready(0, 9) {
		t.Error("R 竟然收下了護甲")
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgCannotReady) {
		t.Errorf("沒說拿不上手:%q", s.Messages)
	}
	s.Messages = nil
	if s.Wear(0, 16) {
		t.Error("W 竟然收下了武器")
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgCannotWear) {
		t.Errorf("沒說穿不上身:%q", s.Messages)
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

// TestEquipListsSplitByCategory:R 與 W 的選單不重疊,且合起來涵蓋全部可裝備的。
func TestEquipListsSplitByCategory(t *testing.T) {
	s := equipScene(t)
	for _, i := range []int{0, 4, 9, 16, 42, 45} {
		s.Inventory.Items[i] = 1
	}
	ready, wear := s.ReadyList(), s.WearList()
	seen := map[byte]bool{}
	for _, it := range append(append([]byte{}, ready...), wear...) {
		if seen[it] {
			t.Errorf("裝備 %d 同時出現在兩張選單", it)
		}
		seen[it] = true
	}
	for _, i := range []int{0, 4, 9, 16, 42, 45} {
		if !seen[byte(i)] {
			t.Errorf("裝備 %d 兩張選單都沒有", i)
		}
	}
	// 盾與武器在 R,其餘在 W。
	for _, it := range ready {
		if slot, _ := SlotFor(it); slot != SlotWeapon && slot != SlotShield {
			t.Errorf("R 的選單裡有 %d(欄位 %d)", it, slot)
		}
	}
}

func equipScene(t *testing.T) *State {
	t.Helper()
	s := newCreateState(t)
	s.MaxMessages = 16
	return s
}
