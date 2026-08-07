package u5data

import (
	"os"
	"testing"
)

// `TILES.4` 解壓後必須正好是 512 個 2bpp tile,而且是 `TILES.16` 的**一半**。
//
// 「一半」這條是兩個獨立數字互相印證:`.16` 的檔頭宣稱 65,536、`.4` 宣稱 32,768,
// 而 4bpp → 2bpp 剛好折半。任一邊的解壓或切法錯了,這條就不成立。
func TestCGATilesAreHalfTheSizeOfEGA(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	cga, err := LoadCGATileSet(dir)
	if err != nil {
		t.Fatal(err)
	}
	ega, err := LoadDOSTileSet(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(cga) != TileCount || len(ega) != TileCount {
		t.Fatalf("CGA %d 個 / EGA %d 個 tile,預期各 %d 個", len(cga), len(ega), TileCount)
	}
	if tileBytesCGA*2 != tileBytesDOS {
		t.Errorf("每個 tile:CGA %d B、EGA %d B —— 不是一半", tileBytesCGA, tileBytesDOS)
	}
	// 色號只能是 0..3。超出範圍代表切法(高位在左)錯了。
	for i := range cga {
		for j, p := range cga[i].Pix {
			if p > 3 {
				t.Fatalf("第 %d 個 tile 的第 %d 個像素色號是 %d,CGA 只有四色", i, j, p)
			}
		}
	}
	// 四個色號都要真的出現過 —— 只用到兩色的話幾乎一定是位元取錯了。
	seen := map[byte]bool{}
	for i := range cga {
		for _, p := range cga[i].Pix {
			seen[p] = true
		}
	}
	if len(seen) != CGAColorCount {
		t.Errorf("整份 tileset 只用了 %d 種色號,預期 %d 種", len(seen), CGAColorCount)
	}
}

// CGA 的調色盤是**十六色盤的 0 / 3 / 5 / 7**(模式 4、調色盤 1、低亮度)。
//
// 依據:`CGA.DRV` 的 `mov bl,1; mov bh,1; mov ah,0Bh; int 10h` 選調色盤 1,
// 接著 `xor bl,bl; xor bh,bh` 把背景設成黑 —— BL 的 bit 3(亮度)是 0。
// ⚠ 寫成高亮度的 11 / 13 / 15 就錯了。
func TestCGAPaletteIsPaletteOneLowIntensity(t *testing.T) {
	if CGAPaletteIndex != [CGAColorCount]byte{0, 3, 5, 7} {
		t.Errorf("調色盤色號是 %v,預期 {0,3,5,7}", CGAPaletteIndex)
	}
	// RGB 要與十六色盤同一份,不是另外抄一組。
	for i, idx := range CGAPaletteIndex {
		if CGAPalette[i] != EGAPalette[idx] {
			t.Errorf("第 %d 色與十六色盤第 %d 色不同", i, idx)
		}
	}
	// 四個顏色互不相同 —— 抄錯索引最常見的症狀是兩色相同。
	for i := 0; i < CGAColorCount; i++ {
		for j := i + 1; j < CGAColorCount; j++ {
			if CGAPalette[i] == CGAPalette[j] {
				t.Errorf("第 %d 與第 %d 色相同", i, j)
			}
		}
	}
	// CGAToEGA 要與表一致。
	for c := byte(0); c < CGAColorCount; c++ {
		if CGAToEGA(c) != CGAPaletteIndex[c] {
			t.Errorf("CGAToEGA(%d) = %d,預期 %d", c, CGAToEGA(c), CGAPaletteIndex[c])
		}
	}
}

// 長度不對要回錯,不能 panic。
func TestParseCGATilesRejectsWrongLength(t *testing.T) {
	for _, n := range []int{0, TileCount*tileBytesCGA - 1, TileCount * tileBytesDOS} {
		if _, err := ParseCGATiles(make([]byte, n)); err == nil {
			t.Errorf("%d B 竟然切得出 CGA tileset", n)
		}
	}
}

// 四種模式名要認得,而且不分大小寫;不認得的要回錯。
func TestParseDisplayMode(t *testing.T) {
	cases := map[string]DisplayMode{
		"EGA": DisplayEGA, "ega": DisplayEGA,
		"CGA": DisplayCGA, "cga": DisplayCGA,
		"Tandy": DisplayTandy, "TANDY": DisplayTandy,
		"Hercules": DisplayHercules,
	}
	for in, want := range cases {
		got, err := ParseDisplayMode(in)
		if err != nil {
			t.Errorf("%q 認不出來:%v", in, err)
			continue
		}
		if got != want {
			t.Errorf("%q → %v,預期 %v", in, got, want)
		}
	}
	if _, err := ParseDisplayMode("VGA"); err == nil {
		t.Error("VGA 竟然認得 —— 原版沒有這個模式")
	}
	// Hercules 是唯一一個素材還沒解的 —— 誠實回報,不要拿 EGA 冒充。
	for m := DisplayMode(0); m < DisplayModeCount; m++ {
		want := m != DisplayHercules
		if m.Implemented() != want {
			t.Errorf("%s 的 Implemented() = %v,預期 %v", m.Name(), m.Implemented(), want)
		}
	}
}

// 依模式載 tileset:CGA 那條的色號要落在 CGA 的四色上(換算成十六色色號之後)。
func TestLoadTileSetForMapsCGAOntoTheSixteenColourIndex(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	cga, err := LoadTileSetFor(dir, DisplayCGA)
	if err != nil {
		t.Fatal(err)
	}
	allowed := map[byte]bool{}
	for _, c := range CGAPaletteIndex {
		allowed[c] = true
	}
	for i := range cga {
		for j, p := range cga[i].Pix {
			if !allowed[p] {
				t.Fatalf("第 %d 個 tile 的第 %d 個像素色號是 %d,不在 %v 之內",
					i, j, p, CGAPaletteIndex)
			}
		}
	}
	// EGA / Tandy 讀同一批 `.16` —— 兩者的 tileset 必須逐位元組相同。
	ega, err := LoadTileSetFor(dir, DisplayEGA)
	if err != nil {
		t.Fatal(err)
	}
	tandy, err := LoadTileSetFor(dir, DisplayTandy)
	if err != nil {
		t.Fatal(err)
	}
	for i := range ega {
		if ega[i] != tandy[i] {
			t.Fatalf("第 %d 個 tile:EGA 與 Tandy 不同 —— 兩者應讀同一批 .16", i)
		}
	}
	// 而 CGA 一定要與 EGA 不同(否則就是根本沒讀 .4)。
	same := true
	for i := range ega {
		if ega[i] != cga[i] {
			same = false
			break
		}
	}
	if same {
		t.Error("CGA 與 EGA 的 tileset 完全相同 —— .4 沒有被讀進來")
	}
}
