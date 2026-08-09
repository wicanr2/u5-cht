package u5data

// tile 通行規則(從原版執行檔取得,2026-08-07)
//
// 來源:FM Towns `WORRIORS.EXP`(SHA-256 見 docs/re/00)記憶體位址 `0x5FF6C`,
// 檔案位移 `0x5FF6C + 0x200`(P3 image offset)。共 64 byte = 512 tile × 1 bit。
//
// 反編譯出的判定邏輯(`sub_2A610`):
//
//	v2 = ((128 >> (tile & 7)) & byte_5FF6C[tile >> 3]) == 0;
//	if ((mover & 0xFE) != 0x1C && (mover & 0xF0) != 0x40 && (tile & 0xFC) == 0x90)
//		return false;
//	return v2;
//
// 也就是 **bit = 1 代表阻擋**;另外 tile 0x90–0x93 只有特定移動者能通過。
//
// 為什麼把表寫進程式碼:這是**遊戲規則**(等同「這把劍傷害多少」),不是美術或音樂素材。
// 重寫引擎本來就要還原規則,而規則若不入庫,引擎就得依賴玩家手上剛好有 FM Towns 版。
// 原版資料檔本身仍然一律不散布(CLAUDE.md §3.0)。
// 可用 `tools/dump_tile_flags.py` 從自備的執行檔重新產生本表來核對。
var tileBlockBits = [64]byte{
	0x70, 0x0C, 0x00, 0x28, 0x01, 0xF3, 0x00, 0xBD,
	0x72, 0x3F, 0xFF, 0xFF, 0xFF, 0xCF, 0xFF, 0xFF,
	0xFC, 0xF6, 0x0F, 0xFF, 0xFF, 0xC7, 0xFF, 0xF7,
	0xF0, 0x3F, 0xFF, 0xF3, 0xFF, 0xFF, 0xFF, 0xBE,
	0x00, 0x00, 0x00, 0x00, 0x03, 0x02, 0x00, 0x00,
	0x06, 0x06, 0x05, 0x06, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0x02,
}

// TileBlocksWalking 回報這個 tile 是否阻擋一般行走。
//
// ⚠ 參數是 int 不是 byte:tileset 有 512 個 tile,byte 放不下。
// (世界地圖每格確實是 1 byte —— 實測 BRIT.DAT 值域 0..255 —— 但 sprite 用到 256 以上。)
func TileBlocksWalking(tile int) bool {
	if tile < 0 || tile >= TileCount {
		return true // 超出範圍當作阻擋,比讓玩家走進未定義區域安全
	}
	return tileBlockBits[tile>>3]&(128>>(tile&7)) != 0
}

// TileIsWater 回報這個 tile 是否為水。
//
// 反編譯自 `sub_2A674`:`tile < 4 || (tile & 0xF0) == 0x60`。
// 前者是三種深度的水(tile 1/2/3)加上 tile 0;後者 0x60–0x6F 是另一類水域物件。
//
// ⚠ tile 0 落在 `< 4` 裡,但它在 tileset 上是綠色星芒(特效),而且**不被阻擋** ——
// 原版這條判斷把 0 併進來的用意待確認,先照抄行為不自行「修正」。
func TileIsWater(tile int) bool {
	return tile >= 0 && (tile < 4 || tile&0xF0 == 0x60)
}

// TileNeedsSpecialMover 回報這個 tile 是否只有特定移動者能通過(`tile & 0xFC == 0x90`)。
func TileNeedsSpecialMover(tile int) bool {
	return tile&0xFC == 0x90
}

// 幾個有名字的 tile。兩個都有直接證據,不是看圖猜的:
//
//   - `TileHorse` 是買馬時寫進物件槽的值(`sub_118CC` 的 `mov al, 10h`)。
//   - `TileWalking` 是隊伍步行的 tile —— `BRIT.OOL` 槽 0 的實際內容就是它,
//     而存檔的載具欄位(`SaveTransportOffset`)讀出來也是 0x1C。
//
// ⚠ 船的 tile **沒有列在這裡**:買船在原版不生成物件,所以那段程式碼裡
// 根本沒有船的 tile 值可抄。船的四個朝向出現在轉向函式 `sub_23FC`
// (0x2C..0x2F),但那是「船在海上時的顯示」,與「買到的是哪種船」是兩件事,
// 還沒對上。沒有證據就不填。
const (
	TileHorse   = 0x10 // 16
	TileWalking = 0x1C // 28
)

// 可以放坐騎的地形。
//
// 原版 `sub_118CC` 找位置時比的三個值:`cmp [var_4], 'D'`(68)、`'E'`(69)、`5`。
// 前兩個是城鎮的地面磚,5 是草地。放不下就「馬廄關門了」。
var mountableTiles = map[int]bool{5: true, 68: true, 69: true}

// TileAllowsMount 回報這個地形能不能放坐騎。
func TileAllowsMount(tile int) bool { return mountableTiles[tile] }

// 載具碼(原版 `byte_3E08C`)
//
// 這是「隊伍現在騎/坐在什麼上面」,與地上那個物件的 tile **不是同一個值**。
// 上下載具的兩支函式(`sub_16F08` / `sub_177AC`)把換算寫得很明白:
//
//	馬    物件 0x10/0x11  ⇄  載具 = 物件 + 2 = 0x12/0x13
//	魔毯  物件 0x1B       ⇄  載具 0x14
//	小艇  物件 0x28..0x2B ⇄  載具 = 物件(同值)
//	大船  物件 0x24..0x27 ⇄  載具 = 物件(同值)
//
// 「馬 + 2」這條同時解釋了通行判定裡的 `byte_3E08C & 0xFE == 0x12`。
const (
	VehicleCarpet   = 0x14 // 0x14/0x15 魔毯
	VehicleWalk     = 0x1C // 0x1C/0x1D 步行
	VehicleSailing  = 0x20 // 0x20..0x23 揚帆中(船正在動)
	VehicleShip     = 0x24 // 0x24..0x27 大船
	VehicleSkiff    = 0x28 // 0x28..0x2B 小艇
	TileCarpetObj   = 0x1B // 地上的魔毯
	HorseToVehicle  = 2    // 騎上馬時載具碼要加的量
	ShipHullWarning = 10   // 耐久低於這個值會警告「船嚴重受損」
)

// VehicleKind 把載具碼歸類。回傳的是該類的起始碼(VehicleWalk / VehicleShip …)。
func VehicleKind(transport byte) int { return int(transport) &^ 0x03 }

// IsOnFoot 回報隊伍是不是在走路。
//
// 原版 `sub_16DA4` 判的就是 `byte_3E08C` 等於 0x1C 或 0x1D ——
// 上馬、上毯、上小艇都要求先下來走路。
func IsOnFoot(transport byte) bool {
	return transport == VehicleWalk || transport == VehicleWalk+1
}

// NPC 的通行規則(原版 `byte_54524`,位址 0x54524)
//
// ⚠ **與玩家的不是同一張表。** 玩家走 `byte_5FF6C`(上面那張),NPC 走這張,
// 兩者有 89 個 tile 判定不同 —— 最明顯的是 tile 16..25(馬與各種載具):
// 玩家可以站上去,NPC 不行。
//
// 判定寫在 `sub_9358`:`(byte_54524[tile>>3] & (0x80 >> (tile&7))) == 0` 才通行,
// 也就是 **bit = 1 代表阻擋**,與玩家那張同格式。
//
// 一開始我只有玩家那張表,差點就拿它給 NPC 用 —— 那會讓 NPC 走到船上與馬背上。
var npcBlockBits = [64]byte{
	0x70, 0x0C, 0xFF, 0xF8, 0x01, 0xF3, 0xFF, 0xFF,
	0x72, 0x3F, 0xFF, 0xFF, 0xFF, 0xCF, 0xFF, 0xFF,
	0xFE, 0xFF, 0x0F, 0xFF, 0xFF, 0xDF, 0xFF, 0x5F,
	0xF0, 0x3F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0x00, 0x00, 0x00, 0x00, 0x13, 0x06, 0x44, 0x00,
	0x04, 0x07, 0x11, 0x11, 0x02, 0x02, 0x01, 0x13,
	0x1F, 0x12, 0x11, 0x11, 0x10, 0x01, 0x13, 0x1F,
	0x01, 0x13, 0x1C, 0x01, 0x10, 0x1C, 0x1C, 0x12,
}

// TileBlocksNPC 回報這個 tile 擋不擋 NPC。
func TileBlocksNPC(tile int) bool {
	if tile < 0 || tile >= TileCount {
		return true
	}
	return npcBlockBits[tile>>3]&(128>>(tile&7)) != 0
}

// 樓梯的 tile。`sub_9358` 對「模式 3」的 NPC 特別放行這兩個值。
//
// ⚠⚠ **這兩個名字此前是反的。** 同一組值在 `sceneset.go` 裡叫
// `LadderUp = 0xC8` / `LadderDown = 0xC9`,而 `ClimbDelta` 對 0xC8 回 **+1**
// (往上)—— 兩個檔案給同一個值取了相反的名字。
//
// `sub_92C0`(NPC 判斷腳下能不能通往目標樓層)獨立佐證了 `sceneset.go` 那邊:
//
//	當前樓層 >  目標樓層(要往下)→ 找 tile **0xC9**
//	當前樓層 <= 目標樓層(要往上)→ 找 tile **0xC8**
//
// ⇒ 以 `sceneset.go` 的 `LadderUp` / `LadderDown` 為準,這裡不再另立名字。
// TileDarkness 是 look 表第 255 筆「darkness!」。
//
// ★ 它在 `UNDER.DAT` 出現 106 次、在 `BRIT.DAT` **一次都沒有**
// ⇒ 幽冥界專屬。站上去視野歸零(`game.LightRadius2`,原版 `sub_2D944`)。
const TileDarkness = 0xFF

const (
	TileStairsDown = LadderDown
	TileStairsUp   = LadderUp
)

// 投射物的穿透規則(原版 `byte_60018`,位址 0x60018)
//
// ⚠ **極性與上面兩張相反。** `sub_2BC34` 是
// `(byte_60018[tile>>3] & (0x80 >> (tile&7))) != 0` → `setnz al`,
// 也就是 **bit = 1 代表「箭飛得過去」**;行走與 NPC 那兩張是 bit = 1 代表擋住。
// 抄錯極性的話箭會只能穿牆、不能穿空地 —— 而畫面上看起來只是「射不到」。
//
// 這張只有 **32 B(256 個 tile)**,不像行走那兩張是 64 B。
// 世界地圖與戰場的格子值域本來就是 0..255,夠用。
//
// 交叉檢查:46 個 tile 擋投射物,其中 39 個同時也擋行走(牆、樹、山、屋)。
// 剩下 7 個 —— 0x12/0x13(騎著的馬)、0x14/0x15(魔毯)、0x19、0x1B、0x3E ——
// **擋箭卻不擋走路**,那些是載具與坐騎:擋在中間的馬會把箭吃掉,
// 但你可以走過去上馬。反過來,**水完全不擋箭**(0..3 全部通),
// 而水是擋行走的。兩張表的差集正好說得通,這比逐格核對可靠。
var projectilePassBits = [32]byte{
	0xFF, 0xF3, 0xC3, 0x8F, 0xFF, 0xFF, 0xFF, 0xC0,
	0xDD, 0xF8, 0x03, 0xDF, 0xFF, 0xFF, 0x00, 0x00,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x3F,
	0xFF, 0xFF, 0xFF, 0xFE, 0xFF, 0xFF, 0xFF, 0xFF,
}

// TileBlocksProjectile 回報這個 tile 擋不擋飛過去的東西(箭、法術、風)。
func TileBlocksProjectile(tile int) bool {
	if tile < 0 || tile >= 256 {
		return true
	}
	return projectilePassBits[tile>>3]&(128>>(tile&7)) == 0
}
