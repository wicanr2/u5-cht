package u5data

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

// `LOOK2.DAT` —— Look 指令的敘述表
//
// # 格式(2026-08-08 一手驗證,並以 FM Towns `sub_CC44` / `sub_D45C` 對照)
//
//	0x0000  u16 × 512   位移表(**絕對檔案位移**,不是相對資料區)
//	0x0400  …           NUL 結尾的 ASCII 字串,重複的敘述共用同一段
//
// 表長剛好 1024 B,而第 0 筆位移就是 1024 —— 表尾接資料頭,沒有空隙。
// 512 項對上 tileset 的 512 格,這是「這一格畫的是什麼」的文字版。
//
// ⚠ **此前 `docs/formats/02-text-files.md` 記載的「218 個 NUL + 大量 0x01–0x1F
// 控制碼,結構化格式未解」是錯的。** 那些「控制碼」其實是位移表的高位位元組
// (0x0400–0x0E17 這個範圍的 u16,高位元組自然落在 0x04–0x0E)。
// 錯因是先數位元組分布再猜格式,而沒有先試最單純的假設:檔頭是不是一張表。
//
// # 兩個索引空間
//
// 原版有兩支讀取函式,差別只在位移:
//
//	sub_CC44(t)  →  表[t]         地形:世界地圖 / 場景地圖那一層的 tile
//	sub_D45C(t)  →  表[t + 256]   物件:BRIT.OOL 那一層的 tile(+ 生物)
//
// 組語裡是 `2 * a1` 與 `2 * a1 + 512`(位元組數),換成項數就是 +0 與 +256。
// 這解釋了為什麼 `TileHorse = 0x10` 在表上查到的是「a small hut」——
// 馬是**物件**,要查 0x110,那裡才是 `a horse`。
//
// 交叉驗證(不是單點碰運氣):
//
//	表[5]   = grass   ⇄  mountableTiles 說 5 是草地
//	表[68]  = cobble  ⇄  mountableTiles 說 68/69 是城鎮地面
//	表[69]  = cobble
//	表[17]  = the Shrine of the Codex  ⇄  TileCodex = 0x11
//	表[25]  = a mystic shrine          ⇄  TileShrine = 0x19
//	表[26]  = a ruined shrine          ⇄  TileShrineDesecrated = 0x1A
//	表[184] = a wooden door            ⇄  TileDoorA = 0xB8
//	表[185] = a locked door            ⇄  TileLockedDoor = 0xB9
//	表[0x110] = a horse                ⇄  TileHorse = 0x10(物件空間)
//	表[0x11B] = an odd rug             ⇄  TileCarpetObj = 0x1B(物件空間)
//	表[0x128] = a skiff                ⇄  VehicleSkiff = 0x28
//
// # 尾巴的空白是有意義的
//
//	表[0xDF] = "the collapsed entrance to the dungeon "   ← 結尾一個空格
//
// 原版接著依**入口的 x 座標**印出地牢名(見 `LookDungeonName`),
// 所以那個空格是接縫,不是髒資料。中文化時這一筆的譯文也要留成可接的形狀。

// LookTiles 是敘述表的項數 —— 與 tileset 同為 512 格。
const LookTiles = 512

// LookObjectBase 是物件索引空間的起點。原版 `sub_D45C` 的 `+512` 位元組。
const LookObjectBase = 256

// lookHeaderSize 是位移表本身的長度,也是第一筆敘述的位移。
const lookHeaderSize = LookTiles * 2

// LookTable 是 `LOOK2.DAT` 解出來的敘述表。
type LookTable struct {
	// text[i] 是第 i 格的敘述。重複的敘述在檔案裡共用一段,
	// 這裡展開成各自一份 —— 512 個短字串,不值得為了省記憶體換來索引間接。
	text [LookTiles]string
}

// LoadLook 讀 `LOOK2.DAT`。
func LoadLook(dir string) (*LookTable, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "LOOK2.DAT"))
	if err != nil {
		return nil, err
	}
	return ParseLook(raw)
}

// ParseLook 解一份 `LOOK2.DAT` 的內容。
func ParseLook(raw []byte) (*LookTable, error) {
	if len(raw) < lookHeaderSize {
		return nil, fmt.Errorf("LOOK2.DAT 只有 %d B,連 %d B 的位移表都放不下", len(raw), lookHeaderSize)
	}
	t := &LookTable{}
	for i := 0; i < LookTiles; i++ {
		off := int(binary.LittleEndian.Uint16(raw[i*2:]))
		// 位移必須落在資料區內。越界就是解錯了,不要靜靜吞掉 ——
		// 這張表是「畫面上要印什麼」,錯了會直接印給玩家看。
		if off < lookHeaderSize || off >= len(raw) {
			return nil, fmt.Errorf("第 %d 格的位移 %d 落在資料區外(%d..%d)",
				i, off, lookHeaderSize, len(raw)-1)
		}
		end := bytes.IndexByte(raw[off:], 0)
		if end < 0 {
			return nil, fmt.Errorf("第 %d 格的敘述從 %d 起沒有結尾的 NUL", i, off)
		}
		t.text[i] = string(raw[off : off+end])
	}
	return t, nil
}

// Terrain 回傳地形 tile 的敘述(原版 `sub_CC44`)。
func (t *LookTable) Terrain(tile int) string { return t.at(tile) }

// Object 回傳物件 tile 的敘述(原版 `sub_D45C`,索引加 256)。
func (t *LookTable) Object(tile int) string { return t.at(tile + LookObjectBase) }

func (t *LookTable) at(i int) string {
	if i < 0 || i >= LookTiles {
		return ""
	}
	return t.text[i]
}

// LookPlaceholder 是原版用來佔位的敘述:沒有安排說明的格子填 `x`(少數填 `*`)。
//
// 判斷「這一格有沒有敘述」要用它,而不是判斷空字串 —— 表上每一格都有字串。
func LookPlaceholder(s string) bool { return s == "x" || s == "*" }
