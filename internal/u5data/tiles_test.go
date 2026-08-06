package u5data

import (
	"os"
	"path/filepath"
	"testing"
)

func fmTownsDir(t *testing.T) string {
	t.Helper()
	dir := os.Getenv("U5_FMTOWNS")
	if dir == "" {
		t.Skip("未設 U5_FMTOWNS(FM Towns ISO 抽出的目錄,需含 U5_E/),跳過")
	}
	p := filepath.Join(dir, "U5_E")
	if _, err := os.Stat(p); err != nil {
		t.Skipf("%s 讀不到:%v", p, err)
	}
	return p
}

func TestParseFMTownsTilesRejectsWrongSize(t *testing.T) {
	if _, err := ParseFMTownsTiles(make([]byte, 1024)); err == nil {
		t.Fatal("大小不對的 .TIL 應該被拒絕")
	}
}

// TestParseFMTownsTilesRejectsNonScaled 守住「機械 2× 放大」這個假設:
// 若餵進去的資料不符(高低 nibble 不同、或相鄰列不同),必須報錯而不是安靜解出垃圾。
func TestParseFMTownsTilesRejectsNonScaled(t *testing.T) {
	raw := make([]byte, fmTownsTilesPerFile*fmTownsTileBytes)
	raw[0] = 0x2A // 高低 nibble 不同
	if _, err := ParseFMTownsTiles(raw); err == nil {
		t.Error("高低 nibble 不同時應報錯(水平 2× 假設不成立)")
	}

	raw = make([]byte, fmTownsTilesPerFile*fmTownsTileBytes)
	raw[16] = 0x22 // 第 1 列與第 0 列不同(第 0 列全 0)
	if _, err := ParseFMTownsTiles(raw); err == nil {
		t.Error("相鄰列不同時應報錯(垂直 2× 假設不成立)")
	}
}

func TestPack4bppRoundTrip(t *testing.T) {
	var tile Tile
	for i := range tile.Pix {
		tile.Pix[i] = byte(i % 16)
	}
	packed := tile.Pack4bpp()
	if len(packed) != tileBytes4bpp {
		t.Fatalf("packed 長度 %d,預期 %d", len(packed), tileBytes4bpp)
	}
	for i := 0; i < TileSize*TileSize; i += 2 {
		b := packed[i/2]
		if b>>4 != tile.Pix[i] || b&0x0F != tile.Pix[i+1] {
			t.Fatalf("第 %d 個像素對 pack 錯:%#02x", i, b)
		}
	}
}

// TestLoadFMTownsTileSet 是 tileset 解碼的整合驗收:
// 512 個 tile 全部要通過 2× 放大的一致性檢查,且 Pack4bpp 串起來剛好 65,536 B ——
// 那個數字正是 DOS 版 TILES.16 檔頭宣稱的解壓後長度(兩個獨立來源互相印證)。
func TestLoadFMTownsTileSet(t *testing.T) {
	dir := fmTownsDir(t)
	paths := []string{
		filepath.Join(dir, "EGA0.TIL"),
		filepath.Join(dir, "EGA1.TIL"),
		filepath.Join(dir, "EGA2.TIL"),
		filepath.Join(dir, "EGA3.TIL"),
	}
	tiles, err := LoadFMTownsTileSet(paths)
	if err != nil {
		t.Fatalf("載入 FM Towns tileset: %v", err)
	}
	if len(tiles) != TileCount {
		t.Fatalf("tile 數 %d,預期 %d", len(tiles), TileCount)
	}

	total := 0
	for i := range tiles {
		total += len(tiles[i].Pack4bpp())
	}
	const wantTotal = TileCount * tileBytes4bpp // 65,536
	if total != wantTotal {
		t.Errorf("4bpp 總長 %d B,預期 %d B(= DOS TILES.16 宣稱的解壓後長度)", total, wantTotal)
	}

	// 色號必須全在 EGA 16 色範圍內。
	for i := range tiles {
		for _, p := range tiles[i].Pix {
			if p > 15 {
				t.Fatalf("tile %d 出現色號 %d,超出 EGA 16 色", i, p)
			}
		}
	}

	// tileset 不該整片空白(若全 0,通常表示切法錯了)。
	nonEmpty := 0
	for i := range tiles {
		for _, p := range tiles[i].Pix {
			if p != 0 {
				nonEmpty++
				break
			}
		}
	}
	if nonEmpty < TileCount/2 {
		t.Errorf("只有 %d/%d 個 tile 有內容,切法可能錯了", nonEmpty, TileCount)
	}
}

// TestDOSTilesHeaderDeclaresOracleSize 記錄兩個獨立來源的一致性:
// DOS 壓縮檔宣稱的解壓後長度,必須等於 FM Towns tileset 還原後的總長。
func TestDOSTilesHeaderDeclaresOracleSize(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過")
	}
	raw, err := os.ReadFile(filepath.Join(dir, "TILES.16"))
	if err != nil {
		t.Skipf("讀不到 TILES.16:%v", err)
	}
	if len(raw) < 4 {
		t.Fatal("TILES.16 太短")
	}
	declared := uint32(raw[0]) | uint32(raw[1])<<8 | uint32(raw[2])<<16 | uint32(raw[3])<<24
	if declared != TileCount*tileBytes4bpp {
		t.Errorf("TILES.16 宣稱解壓後 %d B,但 FM Towns 還原後是 %d B —— 其中一邊的假設錯了",
			declared, TileCount*tileBytes4bpp)
	}
}
