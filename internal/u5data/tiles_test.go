package u5data

import (
	"bytes"
	"fmt"
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

// TestDOSTilesMatchFMTowns:DOS `TILES.16` 解壓的結果與 FM Towns 的
// `EGA*.TIL` **逐位元組相同**。
//
// 這是本專案最硬的一條驗收 —— 兩份來自不同平台、不同壓縮狀態的資料,
// 經過各自的解碼路徑之後應該收斂到同一組 65,536 個位元組。
// 只要 LZW 的任一個參數(打包方向、起始碼寬、加寬時機、清表處理)錯一項,
// 或色號 remap 錯一個位元,這裡就會爆。
//
// ⚠ 它同時是 `docs/formats/01`「TILES.16 不是標準 LZW」那條錯誤記載的墓碑:
// 當初的驗收條件是「解出來要等於 EGA*.TIL 的位元組」,而兩者中間還隔著一層
// 色號順序 —— **驗收條件本身有洞,讓對的答案看起來是錯的**。
func TestDOSTilesMatchFMTowns(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	fm := os.Getenv("U5_FMTOWNS")
	if dir == "" || fm == "" {
		t.Skip("未設 U5_GAMEDATA / U5_FMTOWNS")
	}
	dos, err := LoadDOSTileSet(dir)
	if err != nil {
		t.Fatal(err)
	}
	paths := make([]string, 4)
	for i := range paths {
		paths[i] = filepath.Join(fm, "U5_E", fmt.Sprintf("EGA%d.TIL", i))
	}
	towns, err := LoadFMTownsTileSet(paths)
	if err != nil {
		t.Fatal(err)
	}
	if len(dos) != len(towns) {
		t.Fatalf("DOS %d 個 tile、FM Towns %d 個", len(dos), len(towns))
	}
	bad := 0
	for i := range dos {
		a, b := dos[i].Pack4bpp(), towns[i].Pack4bpp()
		if !bytes.Equal(a, b) {
			if bad < 3 {
				t.Errorf("tile %d 不同:\nDOS      %x\nFM Towns %x", i, a, b)
			}
			bad++
		}
	}
	if bad != 0 {
		t.Fatalf("%d / %d 個 tile 不同", bad, len(dos))
	}
	t.Logf("512 個 tile、%d B 逐位元組相同", len(dos)*tileBytesDOS)
}

// TestDecompressRejectsWrongLength:解出來的長度對不上檔頭就要報錯。
//
// 這條擋的是「解壓器安靜回半截資料」—— 那會表現成畫面下半部是雜訊,
// 而症狀看起來像繪圖的問題,不像解壓的問題。
func TestDecompressRejectsWrongLength(t *testing.T) {
	// 檔頭宣稱 1000 B,但位元流只有一個結束碼。
	raw := []byte{0xE8, 0x03, 0x00, 0x00, 0x01, 0x02}
	if _, err := Decompress(raw); err == nil {
		t.Error("長度對不上卻沒報錯")
	}
	if _, err := Decompress([]byte{1, 2}); err == nil {
		t.Error("連檔頭都不完整卻沒報錯")
	}
}

// TestEveryPictureFileDecompresses:25 個 `.16` 與 25 個 `.4` 全部解得開。
//
// 一個格式如果只在一個檔案上成立,那多半是巧合。整批過才算數。
func TestEveryPictureFileDecompresses(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA")
	}
	names, err := filepath.Glob(filepath.Join(dir, "*.1[64]"))
	if err != nil {
		t.Fatal(err)
	}
	more, _ := filepath.Glob(filepath.Join(dir, "*.4"))
	names = append(names, more...)
	if len(names) < 20 {
		t.Fatalf("只找到 %d 個圖檔,預期至少 20 個", len(names))
	}
	for _, n := range names {
		raw, err := os.ReadFile(n)
		if err != nil {
			t.Fatal(err)
		}
		out, err := Decompress(raw)
		if err != nil {
			t.Errorf("%s: %v", filepath.Base(n), err)
			continue
		}
		if len(out) <= len(raw) {
			t.Errorf("%s 解出 %d B 卻比原本的 %d B 還小 —— 那不像壓縮檔",
				filepath.Base(n), len(out), len(raw))
		}
	}
	t.Logf("%d 個圖檔全部解得開", len(names))
}
