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

// Ready 是 R 指令:給某人換上(或卸下)一件裝備 —— 原版 `sub_1EC34`。
//
// ⚠⚠ **這一支原本漏了七條規則**(`docs/re/72`)。引擎的第一版是
// 「查編號範圍決定部位 → 塞進去 → 舊的放回背包」,而原版有一整套前置檢查,
// 每一條都有自己的拒絕訊息。漏掉的後果不是崩潰,是**遊戲變簡單了**:
//
//  1. **同一鍵也是卸下。** 原版先用 `sub_1E38C` 掃六格,已經穿在身上就脫下來。
//     引擎沒有這一步,所以脫裝備得走另一個指令 —— 而原版沒有那個指令。
//  2. **部位滿了是拒絕,不是換掉。** 原版印「Remove first thy present helm!」
//     之類五句;引擎會**默默把舊的換下來**。差別在玩家的操作步數與心智模型。
//  3. **雙手武器要兩手都空**(部位碼 0x30)。而單手 / 雙手在編號上是**交錯**的,
//     用編號區間永遠推不出來 —— 必須查表(`u5data.EquipSlotCode`)。
//  4. **力量不夠裝不上。** 六格重量加總 + 新件重量 > 力量 → 「Thou art not
//     strong enough!」。引擎完全沒有重量概念。
//  5. **弓沒有箭裝不上。** 弓 / 魔法弓要 `ItemArrows`、十字弓要 `ItemQuarrels`。
//  6. **箭與弩矢本身裝不上**(部位碼 0x00,而且函式開頭就擋掉)。
//  7. **戰鬥中不能換護甲**,除非勝負已定(`byte_3E0B3 != 0` ≈ `Combat.Over`)。
//
// 另外還有一條與 `docs/re/70` 成對的:**戴上隱形戒指 / 再生戒指時 1/16 會消失**
// (`docs/re/70` 找到的是每場戰鬥開始時的那一次,這裡是穿戴的那一次 ——
// 兩處是不同的骰子,不要合併)。
func (s *State) Ready(member int, item byte) bool {
	if member < 0 || member >= len(s.Roster) {
		return false
	}
	// ★ 箭與弩矢:原版開頭就擋,而且**一句話都不印**(`return 0`)。
	if item == u5data.ItemArrows || item == u5data.ItemQuarrels {
		return false
	}
	code := u5data.SlotCodeNone
	if int(item) < u5data.ItemCount {
		code = int(u5data.EquipSlotCode[item])
	}
	if code == u5data.SlotCodeNone {
		s.Log(MsgCannotReady)
		return false
	}
	// ★ 戰鬥中不能換護甲 —— 但勝負已定之後可以。
	if code == u5data.SlotCodeArmour && s.InCombat() && !s.Combat.Over {
		s.Log(MsgNoArmourInBattle)
		return false
	}
	c := &s.Roster[member]
	// ★ 已經穿在身上 → 這一鍵是「卸下」。
	if slot, worn := s.wornSlot(c, item); worn {
		return s.takeOff(member, slot, item)
	}
	// ★ 弓要有箭。
	if ammo, needs := u5data.EquipAmmoFor(item); needs && s.Inventory.Items[ammo] <= 0 {
		s.Log(MsgNoAmmunition)
		return false
	}
	if s.Inventory.Items[item] <= 0 {
		s.Log(MsgDontHaveThat)
		return false
	}
	// ★ 部位規則:滿了就拒絕,不換掉。
	off, ok := s.slotOffsetFor(c, item, code)
	if !ok {
		return false
	}
	// ★ 力量檢查在部位規則**之後**(原版 `var_14` 算得早、用得晚)——
	// 所以「部位滿了」與「力量不夠」同時成立時,印的是部位那一句。
	if u5data.EquipTotalWeight(c)+int(u5data.EquipWeight[item]) > int(c.Strength) {
		s.Log(MsgNotStrongEnough)
		return false
	}
	c.Raw[off] = item
	s.Inventory.Items[item]--
	s.Log(fmt.Sprintf(MsgEquipped, c.Name, s.equipName(int(item))))
	// ★ 穿戴的那一刻,兩種魔法戒指有 1/16 會消失。
	if s.ringVanishesOnWearing(member, item) {
		return true
	}
	// 戰鬥中戴上隱形戒指 → 圖換成通用的站姿(原版 `mov al, 1Dh`)。
	if item == u5data.ItemRingInvisibility && s.InCombat() {
		if slot := s.combatSlotOfRoster(member); slot >= 0 {
			s.Combat.Units[slot].Tile = PartyTileStanding
		}
	}
	return true
}

// wornSlot 掃六格看這件裝備是不是已經穿在身上(原版 `sub_1E38C`)。
func (s *State) wornSlot(c *u5data.Character, item byte) (EquipSlot, bool) {
	for slot := EquipSlot(0); slot < SlotCount; slot++ {
		if c.Raw[slotOffset[slot]] == item {
			return slot, true
		}
	}
	return 0, false
}

// takeOff 卸下(原版 `sub_2F35C` + 背包 +1 + 隱形戒指的收尾)。
func (s *State) takeOff(member int, slot EquipSlot, item byte) bool {
	c := &s.Roster[member]
	c.Raw[slotOffset[slot]] = u5data.ItemNone
	// ★ 背包 +1,而且上限是 99(原版 `cmp dl, 63h; jnb` 跳過遞增)。
	if s.Inventory.Items[item] < u5data.CarryLimit {
		s.Inventory.Items[item]++
	}
	s.Log(fmt.Sprintf(MsgUnequipped, c.Name, s.equipName(int(item))))
	// ★ 戰鬥中脫下隱形戒指要做四件事,而不只是清一個旗標。
	if item == u5data.ItemRingInvisibility && s.InCombat() {
		s.unhideAfterRing(member)
	}
	return true
}

// unhideAfterRing 是戰鬥中脫下隱形戒指之後的收尾(原版 `sub_1EC34` 的
// `loc_1ECD5` 那一段)。
//
// ★ 第四件事最反直覺:**把場上所有怪物的「逃跑」旗標清掉**
// (`for esi = 31 downto 6: unit[esi].flags &= ~2`)。原版的邏輯是
// 「你現形了,怪物就沒有理由繼續亂跑」—— 隱形會讓怪物逃跑,而這是它的反向。
func (s *State) unhideAfterRing(member int) {
	if slot := s.combatSlotOfRoster(member); slot >= 0 {
		s.Combat.Units[slot].Flags &^= UnitHidden
		s.Combat.Units[slot].Tile = int(u5data.PartyCombatTile(&s.Roster[member]))
	}
	// ⚠ 原版這裡還有一行 `byte_3E0B2 &= ~0x10`。引擎**還沒有**這個位元組的欄位
	// (`docs/re/67` §五把它定成「這一擊的結果,已經報過了」的一次性旗標,
	// 而 0x10 是五個位元裡唯一沒解的那一個)。⇒ 這裡先不補一個假欄位;
	// 缺的是那個位元的語意,不是這一行。`docs/re/72` 記著。
	for i := u5data.CombatPartySlots; i < CombatUnitSlots; i++ {
		s.Combat.Units[i].Flags &^= UnitFleeing
	}
}

// slotOffsetFor 依部位碼決定裝進哪一格,並在放不下時印出原版那五句拒絕。
//
// 回傳 false 代表已經印過訊息、不要再往下走。
func (s *State) slotOffsetFor(c *u5data.Character, item byte, code int) (int, bool) {
	switch code {
	case u5data.SlotCodeHelm:
		if c.Raw[u5data.CharHelm] != u5data.ItemNone {
			s.Log(MsgRemoveHelmFirst)
			return 0, false
		}
		return u5data.CharHelm, true
	case u5data.SlotCodeArmour:
		if c.Raw[u5data.CharArmour] != u5data.ItemNone {
			s.Log(MsgRemoveArmourFirst)
			return 0, false
		}
		return u5data.CharArmour, true
	case u5data.SlotCodeOneHand:
		switch s.freeHands(c) {
		case handsNone:
			s.Log(MsgFreeAHandFirst)
			return 0, false
		case handsBoth, handsRightOnly:
			// ★ 兩手都空時原版把 2 改成 0 → 用右手。
			return u5data.CharWeapon, true
		default:
			return u5data.CharShield, true
		}
	case u5data.SlotCodeTwoHand:
		if s.freeHands(c) != handsBoth {
			s.Log(MsgBothHandsFree)
			return 0, false
		}
		// ★ 雙手武器記在**右手**那一格。
		return u5data.CharWeapon, true
	case u5data.SlotCodeAmulet:
		if c.Raw[u5data.CharAmulet] != u5data.ItemNone {
			s.Log(MsgRemoveAmuletFirst)
			return 0, false
		}
		return u5data.CharAmulet, true
	case u5data.SlotCodeRing:
		if c.Raw[u5data.CharRing] != u5data.ItemNone {
			s.Log(MsgOneRingOnly)
			return 0, false
		}
		return u5data.CharRing, true
	}
	s.Log(MsgCannotReady)
	return 0, false
}

// 空手狀況(原版 `sub_1EBE8` 的四種回傳)。
const (
	// handsBoth 兩手都空(原版回 2)。
	handsBoth = 2
	// handsRightOnly 右手空、左手有東西(原版回 0 → 裝右手)。
	handsRightOnly = 0
	// handsLeftOnly 右手有東西而左手空,**而且右手那件不是雙手武器**(原版回 1)。
	handsLeftOnly = 1
	// handsNone 沒手可用(原版回 0xFF)—— 含「右手拿著雙手武器」那一種。
	handsNone = -1
)

// freeHands 是原版 `sub_1EBE8`。
//
// ★ 第三個分支是重點:右手有東西、左手空,但**右手那件是雙手武器**
// (部位碼 0x30)時回「沒手可用」—— 拿著雙手劍就不能再拿盾,
// 即使左手欄位是空的。這條用編號範圍寫不出來。
func (s *State) freeHands(c *u5data.Character) int {
	right, left := c.Raw[u5data.CharWeapon], c.Raw[u5data.CharShield]
	if right == u5data.ItemNone && left == u5data.ItemNone {
		return handsBoth
	}
	if right == u5data.ItemNone {
		return handsRightOnly
	}
	if left == u5data.ItemNone && u5data.EquipSlotCode[right] != u5data.SlotCodeTwoHand {
		return handsLeftOnly
	}
	return handsNone
}

// ringVanishesOnWearing 是穿戴時那 1/16(原版 `random(0,15) == 0`)。
//
// ⚠ 與 `docs/re/70` 的每場戰鬥開始那一次是**不同的骰子**,不要合併:
// 那一次擲 `random(0,15) == 11`、由開戰佈陣觸發;這一次擲 `== 0`、由穿戴觸發。
// 兩者對象相同(隱形戒指 0x2A、再生戒指 0x2C),機率相同,時機與骰值不同。
func (s *State) ringVanishesOnWearing(member int, item byte) bool {
	if item != u5data.ItemRingInvisibility && item != u5data.ItemRingRegeneration {
		return false
	}
	if s.Roll(0, RingVanishRollMax) != 0 {
		return false
	}
	s.Log(MsgRingVanishes)
	s.Roster[member].Raw[u5data.CharRing] = u5data.ItemNone
	return true
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
