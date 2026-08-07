package u5data

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMiscMapsDecodeWithStride16:四張石室的列距是 16 不是 11。
//
// ⚠ 照 11 讀的話第二列開始整張歪掉,而畫面看起來只是「有點亂」——
// 不會有任何東西報錯。所以這裡用**內容**當 oracle:
//
//	牢房   有上鎖的門 0xB9 與上鎖的魔法門 0xBB
//	寶典   正中央那一格是 0x41(那本書)
//	王座廳 有窗 0xAB / 0xAC / 0xAF
//
// 列距讀錯的話這三條全部落空。
func TestMiscMapsDecodeWithStride16(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("沒有 U5_GAMEDATA")
	}
	set, err := LoadMiscMaps(filepath.Join(dir, "MISCMAPS.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	has := func(m *MiscMap, tile byte) bool {
		for _, v := range m.Tiles {
			if v == tile {
				return true
			}
		}
		return false
	}
	cell := &set.Maps[MiscMapIndexCell]
	if !has(cell, TileLockedDoor) || !has(cell, TileLockedMagicDoor) {
		t.Error("牢房裡找不到上鎖的門 —— 列距大概讀成 11 了")
	}
	codex := &set.Maps[MiscMapIndexCodex]
	// ★ 那本書落在玩家站定的位置正前方一格 —— 位移(0x160)與步數(往上 7)
	// 是兩份獨立的來源,對得上才算兩邊都對。
	standY := MiscMapStandY(MiscMapIndexCodex)
	if got := codex.At(MiscMapEnterX, standY-1); got != tileCodexBook {
		t.Errorf("寶典應該在 (%d,%d),那一格卻是 %02X",
			MiscMapEnterX, standY-1, got)
	}
	throne := &set.Maps[MiscMapIndexThrone]
	if !has(throne, 0xAB) || !has(throne, 0xAF) {
		t.Error("王座廳裡找不到窗")
	}
}

// tileCodexBook 是寶典那本書的 tile。
const tileCodexBook = 0x41

// TestMiscMapsRejectShortFile:檔案不夠長要報錯,不要讀出一張空的。
func TestMiscMapsRejectShortFile(t *testing.T) {
	if _, err := ParseMiscMaps(make([]byte, MiscMapThrone)); err == nil {
		t.Error("少了最後一張卻沒有報錯")
	}
}
