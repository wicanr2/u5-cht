package u5data

// 裝備的部位碼與重量 —— 兩張 48 筆的表,直接抄自執行檔
//
// 原版 `sub_1EC34`(Ready)不用「編號範圍」判部位,而是查表:
//
//	byte_40BD4[裝備編號]  部位碼
//	byte_40C04[裝備編號]  重量
//
// ★ **為什麼一定要用表而不是範圍**:範圍表達不出「雙手武器」。
// 部位碼把單手(0x20)與雙手(0x30)分開,而它們在編號上是**交錯**的 ——
// 16 單手、17 雙手、18 單手、19 雙手、20..25 單手、26 雙手…
// 用 `ItemWeaponFirst..Last` 這種區間永遠推不出來。
//
// # 兩張表怎麼從 `.asm` 讀出來
//
// IDA 把它們拆成 `db` / 字串 / `dw` / `dd` 混排(因為 0x20 是空白、0x30 是 `'0'`、
// 0x40 是 `'@'`,整段被當成字串了):
//
//	byte_40BD4  db 4 dup(80h)
//	a000        db '     @@@@@@@ 0 0      0',0     ← 空白 = 0x20、'0' = 0x30、'@' = 0x40
//	a0_0        db '0',0
//	            dw 3020h
//	            dd 30303030h, 20202030h, 2023020h, 4040402h
//
// 逐段展開恰好 **4 + 24 + 2 + 2 + 16 = 48 筆**,與 `ItemCount` 相同 ——
// 這是「讀對了」的第一個佐證(表長剛好對上,沒有多也沒有少)。
//
// # ★★ 第二與第三個佐證:兩個 `0x00` 就是箭與弩矢
//
// 展開後只有**兩筆**是 0x00(不能裝):**編號 27 與 29**。
// 而 `sub_1EC34` 的第一件事就是:
//
//	if (item == 1Bh || item == 1Dh) return    ; 27 / 29,無訊息
//
// 兩處獨立命中同一對編號。第三處:弓(0x1A / 0x24)裝備前要查 `byte_3DFEB`、
// 十字弓(0x1C)要查 `byte_3DFED`,而那兩個位址換算回存檔正是
// `Items[27]` 與 `Items[29]` —— 也就是 `ItemArrows` / `ItemQuarrels`。
//
// ⇒ **箭與弩矢不是裝備,是彈藥**;它們沒有部位、裝不上,而是被弓查詢。
// 三條互不相干的證據指向同一件事(`rulebook/62`)。

// 部位碼(原版 `byte_40BD4` 的值)。
const (
	// SlotCodeNone 裝不上 —— 只有箭(27)與弩矢(29)。
	SlotCodeNone = 0x00
	// SlotCodeRing 戒指。
	SlotCodeRing = 0x02
	// SlotCodeAmulet 頸飾。
	SlotCodeAmulet = 0x04
	// SlotCodeOneHand 單手(含盾)。
	SlotCodeOneHand = 0x20
	// SlotCodeTwoHand 雙手 —— ★ 要兩手都空。
	SlotCodeTwoHand = 0x30
	// SlotCodeArmour 護甲。
	SlotCodeArmour = 0x40
	// SlotCodeHelm 頭盔。
	SlotCodeHelm = 0x80
)

// EquipSlotCode 是每件裝備的部位碼(原版 `byte_40BD4`,48 筆)。
var EquipSlotCode = [ItemCount]byte{
	// 0..3 頭盔
	SlotCodeHelm, SlotCodeHelm, SlotCodeHelm, SlotCodeHelm,
	// 4..8 盾(單手)
	SlotCodeOneHand, SlotCodeOneHand, SlotCodeOneHand, SlotCodeOneHand, SlotCodeOneHand,
	// 9..15 護甲
	SlotCodeArmour, SlotCodeArmour, SlotCodeArmour, SlotCodeArmour,
	SlotCodeArmour, SlotCodeArmour, SlotCodeArmour,
	// 16..26 武器 —— ★ 單手與雙手交錯
	SlotCodeOneHand, // 16
	SlotCodeTwoHand, // 17
	SlotCodeOneHand, // 18
	SlotCodeTwoHand, // 19
	SlotCodeOneHand, // 20
	SlotCodeOneHand, // 21
	SlotCodeOneHand, // 22
	SlotCodeOneHand, // 23
	SlotCodeOneHand, // 24
	SlotCodeOneHand, // 25
	SlotCodeTwoHand, // 26 弓
	SlotCodeNone,    // 27 ★ 箭 —— 彈藥,裝不上
	SlotCodeTwoHand, // 28 十字弓
	SlotCodeNone,    // 29 ★ 弩矢 —— 彈藥,裝不上
	SlotCodeOneHand, // 30
	SlotCodeTwoHand, // 31
	SlotCodeTwoHand, // 32
	SlotCodeTwoHand, // 33
	SlotCodeTwoHand, // 34
	SlotCodeTwoHand, // 35
	SlotCodeTwoHand, // 36 魔法弓
	SlotCodeOneHand, // 37
	SlotCodeOneHand, // 38
	SlotCodeOneHand, // 39
	SlotCodeOneHand, // 40
	SlotCodeTwoHand, // 41
	// 42..44 戒指
	SlotCodeRing, SlotCodeRing, SlotCodeRing,
	// 45..47 頸飾
	SlotCodeAmulet, SlotCodeAmulet, SlotCodeAmulet,
}

// EquipWeight 是每件裝備的重量(原版 `byte_40C04`,48 筆)。
//
// 換裝時原版加總**已裝六格**的重量再加上新件,與力量比:
//
//	力量 >= 已裝重量和 + 新件重量        → 裝得上
//	否則                                 → "Thou art not strong enough!"
//
// ⚠ 重量 0 的不少(空盾 7/8、布甲 9、法杖 27…),所以「重量 0」不代表資料缺 ——
// 原版就是讓一批輕裝完全不吃力量。
var EquipWeight = [ItemCount]byte{
	0, 1, 2, 3, // 0..3   頭盔
	2, 3, 4, 0, 0, // 4..8   盾
	0, 2, 4, 6, 10, 12, 0, // 9..15  護甲
	1, 2, 3, 2, 3, 4, 6, 5, 7, 8, 8, 0, 6, 0, 9, 16, // 16..31
	15, 13, 18, 0, 0, 8, 0, 5, // 32..39
	0, 0, // 40..41
	0, 0, 0, // 42..44 戒指
	0, 0, 0, // 45..47 頸飾
}

// EquipAmmoFor 回報這件武器要哪一種彈藥,以及要不要查。
//
// 原版三個 `cmp`:弓 `0x1A` 與魔法弓 `0x24` 查 `Items[ItemArrows]`,
// 十字弓 `0x1C` 查 `Items[ItemQuarrels]`。其餘不查。
//
// ⚠ 只有**這三件**要查 —— 別擅自推廣到「所有遠程武器」。
func EquipAmmoFor(item byte) (int, bool) {
	switch item {
	case 0x1A, 0x24:
		return ItemArrows, true
	case 0x1C:
		return ItemQuarrels, true
	}
	return 0, false
}

// EquipTotalWeight 是某個角色身上六格裝備的重量和(原版那個六次的迴圈)。
func EquipTotalWeight(c *Character) int {
	sum := 0
	for _, off := range []int{CharHelm, CharArmour, CharWeapon, CharShield, CharRing, CharAmulet} {
		if it := c.Raw[off]; it != ItemNone {
			sum += int(EquipWeight[it])
		}
	}
	return sum
}
