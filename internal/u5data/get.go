package u5data

// Get 指令:撿東西(原版 `sub_15A94` → `sub_154BC`,地牢走 `sub_15930`)
//
// 這是三塊寶石碎片、檀香木盒、王冠、權杖、寶珠、月石、魔毯的**唯一**入口 ——
// 少了它,真結局那條路根本走不到,而引擎在此之前完全沒有這個指令。
//
// 兩條路:那一格上**有物件**就撿物件,沒有就看**地形**能不能拿(牆上的火把、
// 田裡的作物、桌上的食物)。兩條都不成立才是「這裡沒有東西可拿。」

// 物件種類碼 → 撿到什麼(`sub_154BC` 的跳表與其後的比較鏈)。
const (
	ItemClosedChest  = 0x01 // 「得先打開它!」
	ItemGold         = 0x02
	ItemPotion       = 0x03
	ItemPlans        = 0x04 // 圖紙(byte_3DFC9)
	ItemKey          = 0x07
	ItemGem          = 0x08
	ItemTorch        = 0x0D
	ItemSandalwood   = 0x0E // ★ 檀香木盒 —— 真結局的條件
	ItemFood         = 0x0F
	ItemMoonstone    = 0x19
	ItemMagicCarpet  = 0x1B
	ItemShardBase    = 0xB4 // 0xB4..0xB7:碎片 / 王冠 / 權杖 / 寶珠
	ItemShard        = 0xB4
	ItemCrown        = 0xB5
	ItemSceptre      = 0xB6
	ItemOrb          = 0xB7
)

// GetPickable 回報那一格的物件撿不撿得起來(`sub_15A94` 的四個條件)。
//
// 原版是一串 `break` 條件,不是白名單:
//
//	種類 < 0x10          撿(一般物品:錢、藥水、鑰匙、寶石、火把、盒子、食物…)
//	種類 == 0x19         月石
//	種類 == 0x1B         魔毯
//	(種類 & 0xFC) == 0xB4 碎片 / 王冠 / 權杖 / 寶珠
//	其餘                 跳過,繼續找下一個槽
//
// ⚠ 「其餘」包含**怪物與坐騎**(種類 ≥ 0x40)—— 站在馬旁邊按 Get 不會把馬
// 撿起來,而是繼續掃後面的槽。這個「繼續掃」而不是「就此放棄」的差別,
// 決定了同一格疊了兩樣東西時撿到哪一個。
func GetPickable(kind byte) bool {
	switch {
	case kind < 0x10:
		return true
	case kind == ItemMoonstone, kind == ItemMagicCarpet:
		return true
	case kind&0xFC == ItemShardBase:
		return true
	}
	return false
}

// ShardIndex 是這個碎片物件對應第幾塊(原版 `mov eax, edi; and eax, 3`)。
//
// ⚠ 用的是物件的**品質欄**(槽位移 +5),不是種類碼。四個 0xB4..0xB7 共用
// 同一段程式碼,種類碼只決定「是碎片還是王冠」,是哪一塊碎片看品質欄。
func ShardIndex(quality int) int { return quality & 3 }

// 地形能不能拿(`sub_15A94` 後半的 switch)。
const (
	// TileWallTorchA / TileWallTorchB 是牆上的火把。拿下來換成磚地。
	TileWallTorchA = 0xB0
	TileWallTorchB = 0xB1
	// TileCrops / TileCropsPicked 是田裡的作物與收割後的空地。
	TileCrops       = 0x2D
	TileCropsPicked = 0x2C
	// 桌上的食物:一張桌子橫跨三格,中間那格拿走之後會變成左半或右半。
	TilePlateMiddle = 0x9C
	TilePlateNorth  = 0x9A
	TilePlateSouth  = 0x9B
	TilePlateEmpty  = 0x95
)

// BorrowedTorchTurns 是從牆上拿下來的火把能撐多久(`mov byte_3E0B7, 64h`)。
//
// = 100 分鐘,固定值 —— 與自己點一把(`random(0,15) + 0x70`,112..127)不同,
// 而且訊息是「借用!」不是「點亮了」。
const BorrowedTorchTurns = 100

// GetKarmaPenalty 是拿食物扣的業報(`dec byte_3E098`)。
//
// ★ **牆上的火把不扣**(那條路直接跳到結尾,原版還特地說「借用!」),
// **作物與桌上的食物扣 1**。這就是原版對「偷竊」的定義 ——
// 不是有沒有人看到,而是拿了誰的東西。
const GetKarmaPenalty = 1

// PlateReach 回報從 (dx, dy) 這個方向伸手拿不拿得到這一格的食物。
//
// 原版對三種盤子各有各的規矩,而且**只看 Y**:
//
//	0x9A(北半)  要從南邊拿(dy == 1)
//	0x9B(南半)  要從北邊拿(dy == -1)
//	0x9C(整張)  不能從東西向拿(dx 必須是 0)
//
// 拿不到時印「碰不到那個盤子!」。也就是說站在桌子側面伸手是搆不著的 ——
// 得繞到桌子的長邊。
func PlateReach(tile byte, dx, dy int) bool {
	switch tile {
	case TilePlateNorth:
		return dy == 1
	case TilePlateSouth:
		return dy == -1
	case TilePlateMiddle:
		return dx == 0
	}
	return false
}

// PlateAfter 回傳拿走食物之後那一格變成什麼。
//
// 整張桌子(0x9C)被拿走一半之後**變成剩下的那一半**,不是直接清空:
// 從南邊拿 → 剩南半(0x9B);從北邊拿 → 剩北半(0x9A)。
// 所以同一張桌子可以拿兩次。
func PlateAfter(tile byte, dy int) byte {
	if tile == TilePlateMiddle {
		if dy == 1 {
			return TilePlateSouth
		}
		if dy == -1 {
			return TilePlateNorth
		}
		return tile
	}
	return TilePlateEmpty
}

// DungeonGetLocationMin / Max 是「按 Get 要去開地牢寶箱」的地點編號範圍。
//
// 原版 `cmp al, 20h; jbe` / `cmp al, 29h; jnb` —— 也就是 **33..40**,
// 八座地牢。落在這個範圍時 `sub_15A94` 一開頭就轉給 `sub_15930`。
const (
	DungeonGetLocationMin = 33
	DungeonGetLocationMax = 40
)

// 地牢寶箱的獎品(原版 `sub_15930`,三張並列表在 0x55DD4 / 0x55DDC / 0x55DE4)
//
// ★ 這一組**與地表寶箱(`sub_15020`)完全是兩套**。地表寶箱有「等級」
// (物件的品質位元組),地牢寶箱**沒有** —— 它只擲 `random(1, 樓層×4 + 4)`,
// 再拿七個門檻各比一次。`docs/re/` 先前記著「地牢寶箱的等級從哪來還沒找到」,
// 答案是:根本沒有那個東西,是另一支函式。
//
//	i  門檻  數量上限  種類碼  掉什麼
//	0    2      31      0x0F   食物
//	1    4   樓層×8     0x02   金幣      ← 數量上限是算出來的,不是查表
//	2    5       3      0x07   鑰匙
//	3   10       3      0x08   寶石
//	4   20       3      0x0D   火把
//	5   25       7      特例    藥水(種類碼 3,數量 random(0,7))
//	6   25       7      特例    卷軸(種類碼 4,數量 random(0,7))
//	7    0       0        0    表尾
var (
	// DungeonLootThreshold 是每一類的門檻:擲出來 ≥ 它才拿得到。
	DungeonLootThreshold = [7]byte{2, 4, 5, 10, 20, 25, 25}
	// DungeonLootMax 是數量的上限(索引 1 的金幣不用它,見上表)。
	DungeonLootMax = [7]byte{31, 0, 3, 3, 3, 7, 7}
	// DungeonLootKind 是種類碼;0xFF 代表由程式碼直接指定(藥水 / 卷軸)。
	DungeonLootKind = [7]byte{0x0F, 0x02, 0x07, 0x08, 0x0D, 0xFF, 0xFF}
)

// DungeonLootSpecial 是索引 5 / 6 那兩類寫死的種類碼。
var DungeonLootSpecial = map[int]byte{5: ItemPotion, 6: ItemPlans}

// DungeonLootRollMax 是那顆骰子的上限:`random(1, 樓層×4 + 4)`。
//
// 所以第 1 層(樓層 0)最多擲到 4 —— 只拿得到食物(門檻 2)與金幣(門檻 4);
// 火把要第 5 層以後、藥水與卷軸要第 6 層以後才可能出現。
func DungeonLootRollMax(floor int) int { return floor*4 + 4 }

// DungeonEmptiedChest 是把開過的寶箱清空之後那一格變成什麼。
//
// 原版 `and byte ptr [eax+ecx], 8` —— 只留「頭上有洞」那一位元。
func DungeonEmptiedChest(tile byte) byte { return tile & DungeonHoleAbove }
