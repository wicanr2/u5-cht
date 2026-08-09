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
	// 四種模式的素材現在都解得出來(Hercules 走 `HER.DRV` 的圖樣表)。
	for m := DisplayMode(0); m < DisplayModeCount; m++ {
		if !m.Implemented() {
			t.Errorf("%s 的素材竟然還沒解", m.Name())
		}
	}
	// 範圍外的不算實作 —— 這條擋的是「將來加了第五種卻忘了處理」。
	if DisplayModeCount.Implemented() {
		t.Error("範圍外的模式竟然回報已實作")
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

// ★ 這條測試同時釘住**兩個不同來源**,而它們在色號 6 上不一致 —— 刻意的。
//
//	資料事實:`DATA.OVL` 0x52EE 那張表的第 7 個值是 0x06(暗黃)
//	行為事實:對 DOSBox 跑的原版取色,畫面上是 #AA5500(棕)
//
// 十五格相符、第七格不同,而那一格顯示的正好是 EGA 預設值 ⇒ 表沒有活到遊戲畫面
// (推測是之後又設了一次顯示模式;`int 10h` 設模式會重設調色盤暫存器)。
// 完整量測與推理在 `docs/re/62`。
//
// 釘住的三件事:表的位移 0x52EE、6-bit 值 → RGB 的換算、以及
// **`EGAPalette[6]` 用的是量到的棕,不是表裡的暗黃**。
func TestEGAPaletteMatchesTheTableInDataOVL(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	pal, overscan, err := LoadEGAPalette(dir)
	if err != nil {
		t.Fatal(err)
	}
	// ★ 十五格相符;**第 7 格(色號 6)刻意不同**,理由見下面。
	for i := range pal {
		if i == 6 {
			continue
		}
		if pal[i] != EGAPalette[i] {
			t.Errorf("色號 %d:表裡是 %v,常數是 %v", i, pal[i], EGAPalette[i])
		}
	}
	// 邊框是黑的(表的第 17 個值是 0)。
	if overscan != (EGAPalette[0]) {
		t.Errorf("overscan 是 %v,預期與色號 0 相同(黑)", overscan)
	}
	// 表裡確實是 0x06 = 暗黃 —— 這是**資料事實**,要釘住。
	if got := pal[6]; got.R != 0xAA || got.G != 0xAA || got.B != 0x00 {
		t.Errorf("表的第 7 個值換算出 %v,預期 0x06 = 暗黃 (AA,AA,00) —— 位移可能不對", got)
	}
	// ★★ 但**畫面上是棕的** —— 這是**行為事實**,量自 DOSBox 跑的原版
	// (小屋內部 4,184 px、大地圖森林 420 px,`machine=ega` 與 `vgaonly` 都是 #AA5500)。
	// 兩件事同時成立:表被載入,但沒活到遊戲畫面(見 `tiles.go` 的說明)。
	// ⚠ **不要把這一格「修回」表裡的值** —— 那會讓每一張畫面的木頭都偏色。
	if got := EGAPalette[6]; got.R != 0xAA || got.G != 0x55 || got.B != 0x00 {
		t.Errorf("色號 6 是 %v,原版畫面上量到的是棕 (AA,55,00)", got)
	}
}

// 6-bit 調色盤值 → RGB 的四級換算,連兩個容易搞混的值一起釘住。
func TestEGAColorFromValueHasFourLevelsPerChannel(t *testing.T) {
	cases := map[byte][3]byte{
		0x00: {0x00, 0x00, 0x00}, // 黑
		0x01: {0x00, 0x00, 0xAA}, // 藍(主級)
		0x06: {0xAA, 0xAA, 0x00}, // ★ 暗黃 —— 原版色號 6
		0x14: {0xAA, 0x55, 0x00}, // ★ 棕 —— 慣例的色號 6,原版**沒用**
		0x07: {0xAA, 0xAA, 0xAA}, // 淺灰
		0x38: {0x55, 0x55, 0x55}, // 深灰(只有次級)
		0x3F: {0xFF, 0xFF, 0xFF}, // 白(主級 + 次級)
		0x3E: {0xFF, 0xFF, 0x55}, // 黃
	}
	for v, want := range cases {
		got := EGAColorFromValue(v)
		if got.R != want[0] || got.G != want[1] || got.B != want[2] {
			t.Errorf("0x%02X → (%02X,%02X,%02X),預期 (%02X,%02X,%02X)",
				v, got.R, got.G, got.B, want[0], want[1], want[2])
		}
	}
}

// Hercules 的四個圖樣就是 `HER.DRV` 0x0318 的四個位元組。
//
// 語意:0 全暗、3 全亮、1 與 2 都是 50% 灰但**相位相反**(所以相鄰時分得出來)。
// ⚠ 四筆就是全部 —— 若原版還要依列號交替相位,表就得有八筆。
func TestHerculesPatternsAreTheFourFromTheDriver(t *testing.T) {
	if HerculesPattern != [HerculesPatternCount]byte{0x00, 0x55, 0xAA, 0xFF} {
		t.Fatalf("圖樣表是 %v,預期 {00,55,AA,FF}", HerculesPattern)
	}
	// 色號 0 每一欄都暗、色號 3 每一欄都亮。
	for x := 0; x < 8; x++ {
		if HerculesBit(0, x) {
			t.Errorf("色號 0 在第 %d 欄竟然亮著", x)
		}
		if !HerculesBit(3, x) {
			t.Errorf("色號 3 在第 %d 欄竟然是暗的", x)
		}
	}
	// 色號 1 與 2 每一欄都相反 —— 這是「相位相反」的定義。
	for x := 0; x < 8; x++ {
		if HerculesBit(1, x) == HerculesBit(2, x) {
			t.Errorf("色號 1 與 2 在第 %d 欄相同 —— 相位沒有相反", x)
		}
	}
	// 兩者各自都是一半亮 —— 50% 灰。
	for _, c := range []byte{1, 2} {
		on := 0
		for x := 0; x < 8; x++ {
			if HerculesBit(c, x) {
				on++
			}
		}
		if on != 4 {
			t.Errorf("色號 %d 有 %d/8 欄亮著,預期 4(50%%)", c, on)
		}
	}
}

// Hercules 的 tileset 讀的是 `.4`(2bpp),而且結果只有黑與白兩種色號。
func TestHerculesTileSetIsMonochromeFromTheTwoBitAssets(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	her, err := LoadTileSetFor(dir, DisplayHercules)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[byte]bool{}
	for i := range her {
		for _, p := range her[i].Pix {
			seen[p] = true
		}
	}
	if len(seen) != 2 || !seen[0] || !seen[15] {
		t.Errorf("色號用到 %v,預期只有 0 與 15", seen)
	}
	// 與 CGA 的來源比:亮的像素數必須介於「色號 3 的數量」與「3+1+2 的數量」之間。
	cga, err := LoadCGATileSet(dir)
	if err != nil {
		t.Fatal(err)
	}
	var solid, half, lit int
	for i := range cga {
		for _, p := range cga[i].Pix {
			switch p {
			case 3:
				solid++
			case 1, 2:
				half++
			}
		}
	}
	for i := range her {
		for _, p := range her[i].Pix {
			if p == 15 {
				lit++
			}
		}
	}
	if lit <= solid || lit >= solid+half {
		t.Errorf("亮的像素 %d 個,預期落在 (%d, %d) 之間 —— 半色調沒有生效",
			lit, solid, solid+half)
	}
}
