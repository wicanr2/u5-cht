package u5data

// 固定物品表的兩條特例(原版 `sub_14160` 的 `if (i < 0x0D || i > 0x0F)`)

// 可重複拿的索引範圍(原版 0x0D..0x0F 不寫「已撿過」位元圖)。
const (
	HiddenItemRepeatFirst = 0x0D
	HiddenItemRepeatLast  = 0x0F
	// HiddenItemSpareKeys 是那一串「鑰匙用完才會再長出來」的鑰匙
	//(原版 `if (i == 0x0D && byte_3DFB8 == 0)`)。
	HiddenItemSpareKeys = 0x0D
)

// HiddenItemRepeatable 回報第 i 筆撿走之後會不會再出現。
func HiddenItemRepeatable(i int) bool {
	return i >= HiddenItemRepeatFirst && i <= HiddenItemRepeatLast
}

// HiddenTakenBytes 是「已撿過」位元圖的長度(原版 `byte_3E06C`,113 位元)。
const HiddenTakenBytes = (HiddenItemCount + 7) / 8

// DungeonRoomItemName 是地牢房間 / 搜尋結果的物件描述(原版 `sub_13CB0`)。
//
// ⚠⚠ 這**不是** `LOOK2.DAT` 的物件表 —— 兩張表有六處不同,而差異是刻意的:
//
//	種類  LOOK2.DAT(地面 Look)   sub_13CB0(這一張)
//	  2   gold                    a sack of gold
//	  7   a key                   a ring of keys
//	 13   a torch                 some torches
//	 25   a moon stone            ★ a strange rock   ← 搜出來時**認不出**是月石
//	 30   a corpse                a rotting body
//	 31   a corpse                a moldy corpse
//	 14   a sandalwood box        ★ 落到 default「什麼也沒有」
//
// ★ 第 25 筆最有意思:自己埋的月石被搜出來時,遊戲只說「一塊奇怪的石頭」。
// ⚠ 第 14 筆(檀香木盒)落到 default —— 它是真結局的關鍵道具,而這張表沒有它。
// 照原版保留。
func DungeonRoomItemName(kind byte) string {
	if n, ok := dungeonRoomItemNames[kind]; ok {
		return n
	}
	return "nothing of note."
}

var dungeonRoomItemNames = map[byte]string{
	1:  "a chest",
	2:  "a sack of gold",
	3:  "a potion",
	4:  "a scroll",
	5:  "a weapon",
	6:  "a shield",
	7:  "a ring of keys",
	8:  "a gem",
	9:  "a helm",
	10: "a ring",
	11: "some armour",
	12: "an amulet",
	13: "some torches",
	15: "some food",
	25: "a strange rock",
	30: "a rotting body",
	31: "a moldy corpse",
}
