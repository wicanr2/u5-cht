package game

import (
	"fmt"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// Ready(R)—— 換裝備
//
// ⚠⚠ **更正(2026-08-08,`docs/re/49`)**:原版**只有一個 R**,沒有 W。
// 我第一版把 U4 的「Ready 武器 / Wear 護甲」兩鍵模型搬過來,而 U5 的
// 主指令分派器 `sub_2ACF4` 對 'W' 只印一句 **"W-What?"** —— 它不是指令
//('D' 也一樣)。兩個分派器(`sub_2ACF4` 世界/場景/地牢、`sub_A360` 戰鬥)
// 都各自印這句,兩處獨立佐證。
//
// 原版 R(`sub_1F3A4`)的清單範圍是 `sub_1E418(-1, 0x30, …)` ——
// **0x30 = 48 = 全部裝備**,頭盔護甲戒指頸飾武器盾一起列。
//
//	R Ready  頭盔(0..3)→ 頭;盾(4..8)→ 左手;護甲(9..15)→ 身;
//	         武器(16..41)→ 右手 / 左手;戒指(42..44)→ 指;頸飾(45..47)→ 頸
//
// 「哪個欄位收得下」由 `SlotFor` 判斷,那部分沒變 —— 錯的只是鍵位與清單切分。
// 這是 `rulebook/65` 的典型:自家測試全綠,而行為與原版不同。
//
// 裝備欄位的位移早在 `u5data/items.go` 就解出來了(`CharHelm` … `CharAmulet`),
// 由六名初始角色橫向對照定出 —— 法師沒頭盔穿布甲拿匕首、戰士釘盔加釘盾、
// 吟遊詩人雙手各一把武器。缺的只是換裝的動作。
//
// ⚠ **換下來的要放回背包。** 少了這一步,玩家換一次裝備就永久失去原本那件 ——
// 而症狀要等到很久以後翻背包才會發現。

// EquipSlot 是身上的一個裝備欄位。
type EquipSlot int

// 六個欄位,順序與 `u5data.Equipment.Slots` 相同。
const (
	SlotHelm EquipSlot = iota
	SlotArmour
	SlotWeapon
	SlotShield
	SlotRing
	SlotAmulet
	SlotCount
)

// slotOffset 是每個欄位在角色紀錄裡的位移。
var slotOffset = [SlotCount]int{
	SlotHelm:   u5data.CharHelm,
	SlotArmour: u5data.CharArmour,
	SlotWeapon: u5data.CharWeapon,
	SlotShield: u5data.CharShield,
	SlotRing:   u5data.CharRing,
	SlotAmulet: u5data.CharAmulet,
}

// SlotFor 回報這件裝備該放哪個欄位,以及放不放得下。
//
// 分界照 `u5data` 的裝備分類(`ItemHelmFirst` 那組),那是從名字表的實際排列
// 定出來的 —— 不是自己畫的界線。
//
// ⚠ 武器與盾**都收在左手**:原版的 0x1C 欄位既放盾也放第二把武器
//(Iolo 就是雙手各一把)。所以「左手」不叫 shield 而叫副手。
func SlotFor(item byte) (EquipSlot, bool) {
	i := int(item)
	switch {
	case item == u5data.ItemNone:
		return 0, false
	case i >= u5data.ItemHelmFirst && i <= u5data.ItemHelmLast:
		return SlotHelm, true
	case i >= u5data.ItemArmourFirst && i <= u5data.ItemArmourLast:
		return SlotArmour, true
	case i >= u5data.ItemShieldFirst && i <= u5data.ItemShieldLast:
		return SlotShield, true
	case i >= u5data.ItemWeaponFirst && i <= u5data.ItemWeaponLast:
		return SlotWeapon, true
	case i >= u5data.ItemRingFirst && i <= u5data.ItemRingLast:
		// 42..44 是戒指、45..47 是頸飾 —— 名字表就是這樣排的
		//(Ring of Invisibility / Protection / Regeneration,
		//  Amulet of Turning / Spiked Collar / Ankh)。
		if i <= u5data.ItemRingFirst+2 {
			return SlotRing, true
		}
		return SlotAmulet, true
	}
	return 0, false
}

// Ready 是 R 指令:給某人換上一件裝備。**六個欄位都走這一支。**
func (s *State) Ready(member int, item byte) bool {
	slot, ok := SlotFor(item)
	if !ok {
		s.Log(MsgCannotReady)
		return false
	}
	return s.equip(member, slot, item)
}

// equip 換裝:背包 −1、身上原本那件放回背包。
//
// 武器有個例外:右手滿了就往左手放(原版的 0x1C 既收盾也收第二把武器)。
// 兩手都滿才換掉右手 —— 不然玩家永遠裝不上第二把。
func (s *State) equip(member int, slot EquipSlot, item byte) bool {
	if member < 0 || member >= len(s.Roster) {
		return false
	}
	if s.Inventory.Items[item] <= 0 {
		s.Log(MsgDontHaveThat)
		return false
	}
	c := &s.Roster[member]
	if slot == SlotWeapon && c.Raw[u5data.CharWeapon] != u5data.ItemNone &&
		c.Raw[u5data.CharShield] == u5data.ItemNone {
		slot = SlotShield
	}
	off := slotOffset[slot]
	if old := c.Raw[off]; old != u5data.ItemNone {
		s.Inventory.Items[old]++
	}
	c.Raw[off] = item
	s.Inventory.Items[item]--
	s.Log(fmt.Sprintf(MsgEquipped, c.Name, s.equipName(int(item))))
	return true
}

// Unequip 把某個欄位上的東西卸回背包。
func (s *State) Unequip(member int, slot EquipSlot) bool {
	if member < 0 || member >= len(s.Roster) || slot < 0 || slot >= SlotCount {
		return false
	}
	c := &s.Roster[member]
	off := slotOffset[slot]
	old := c.Raw[off]
	if old == u5data.ItemNone {
		return false
	}
	c.Raw[off] = u5data.ItemNone
	s.Inventory.Items[old]++
	s.Log(fmt.Sprintf(MsgUnequipped, c.Name, s.equipName(int(old))))
	return true
}

// ReadyList 是背包裡**所有**穿戴得上的東西(給 R 指令的選單)。
//
// 原版的範圍是 `sub_1E418(-1, 0x30, …)` —— 48 件裝備全列,不分武器與護甲。
func (s *State) ReadyList() []byte {
	var out []byte
	for i := 0; i < u5data.ItemCount; i++ {
		if s.Inventory.Items[i] <= 0 {
			continue
		}
		if _, ok := SlotFor(byte(i)); !ok {
			continue
		}
		out = append(out, byte(i))
	}
	return out
}
