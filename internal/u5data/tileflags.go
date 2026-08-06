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
