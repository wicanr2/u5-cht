package u5data

import (
	"os"
	"testing"
)

// TestTileFlagsMatchOriginal 用原版執行檔核對寫進程式碼的通行表。
//
// 這張表是遊戲規則,不是猜的 —— 所以要能隨時對回原版。
// 沒有 FM Towns 執行檔時跳過(表本身仍在,只是這次不核對)。
func TestTileFlagsMatchOriginal(t *testing.T) {
	path := os.Getenv("U5_FMTOWNS_EXE")
	if path == "" {
		path = "../../re_work/fmtowns/WORRIORS.EXP"
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("找不到 %s,跳過與原版的核對", path)
	}
	const addr, imageOff = 0x5FF6C, 0x200
	off := addr + imageOff
	if len(raw) < off+len(tileBlockBits) {
		t.Fatalf("執行檔只有 %d B,取不到 0x%X 的表", len(raw), off)
	}
	for i := range tileBlockBits {
		if got := raw[off+i]; got != tileBlockBits[i] {
			t.Fatalf("第 %d 個 byte 與原版不符:程式碼 0x%02X,原版 0x%02X", i, tileBlockBits[i], got)
		}
	}
}

// TestWaterTilesAreBlocked 是這張表的 oracle:三種水必須擋住一般行走。
// 這條同時驗證了「反編譯出的 bit 判定式」有沒有抄錯。
func TestWaterTilesAreBlocked(t *testing.T) {
	for _, tile := range []int{1, 2, 3} {
		if !TileBlocksWalking(tile) {
			t.Errorf("tile %d 是水,應該阻擋一般行走", tile)
		}
		if !TileIsWater(tile) {
			t.Errorf("tile %d 應該被判定為水", tile)
		}
	}
	// 0x60–0x6F 是另一類水域(sub_2A674 的第二個條件)
	for _, tile := range []int{0x60, 0x6A, 0x6F} {
		if !TileIsWater(tile) {
			t.Errorf("tile 0x%02X 應該被判定為水", tile)
		}
	}
	if TileIsWater(0x70) {
		t.Error("tile 0x70 不在 0x60–0x6F,不該判定為水")
	}
}

// TestTileFlagsDistribution:表若讀錯位址,分布會明顯異常(全 0 或全 1)。
func TestTileFlagsDistribution(t *testing.T) {
	blocked := 0
	for i := 0; i < 512; i++ {
		if TileBlocksWalking(i) {
			blocked++
		}
	}
	// 實測:阻擋 195 / 512(這個數字是從原版執行檔算出來的,不是估的)。
	if blocked != 195 {
		t.Errorf("阻擋 %d/512 個 tile,實測應為 195 —— 表或判定式有一邊錯了", blocked)
	}
}

func TestTileNeedsSpecialMover(t *testing.T) {
	for _, tile := range []int{0x90, 0x91, 0x92, 0x93} {
		if !TileNeedsSpecialMover(tile) {
			t.Errorf("tile 0x%02X 應該需要特定移動者", tile)
		}
	}
	if TileNeedsSpecialMover(0x94) {
		t.Error("tile 0x94 不在 0x90–0x93")
	}
}
