package u5data

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
)

// TileSize 是原版 tile 的邏輯邊長(16×16)。
const TileSize = 16

// TileCount 是 Ultima V 的 tile 總數。
//
// 佐證:DOS 版 TILES.16 檔頭前 4 byte 宣稱解壓後 65,536 B,
// 而 65,536 / (16×16 個 4bpp 像素 = 128 B) = 512。
const TileCount = 512

// tileBytes4bpp 是一個 16×16 4bpp tile 的位元組數(每列 8 B × 16 列)。
const tileBytes4bpp = TileSize * TileSize / 2

// Tile 是一個 16×16 的索引色點陣,每個元素是 0–15 的 EGA 色號。
type Tile struct {
	Pix [TileSize * TileSize]byte
}

// At 回報 (x, y) 的色號。超出範圍回 0。
func (t *Tile) At(x, y int) byte {
	if x < 0 || x >= TileSize || y < 0 || y >= TileSize {
		return 0
	}
	return t.Pix[y*TileSize+x]
}

// EGAPalette 是標準 EGA 16 色(IRGB 位元順序:bit3=I, bit2=R, bit1=G, bit0=B)。
//
// **Tile.Pix 存的就是這一組索引。** 兩個來源在載入時都已經正規化成標準 EGA:
// DOS 的 `TILES.16` 本來就是,FM Towns 的 `EGA*.TIL` 由 `tileColorRemap` 換過。
var EGAPalette = [16]color.NRGBA{
	{0x00, 0x00, 0x00, 0xFF}, // 0  黑
	{0x00, 0x00, 0xAA, 0xFF}, // 1  藍
	{0x00, 0xAA, 0x00, 0xFF}, // 2  綠
	{0x00, 0xAA, 0xAA, 0xFF}, // 3  青
	{0xAA, 0x00, 0x00, 0xFF}, // 4  紅
	{0xAA, 0x00, 0xAA, 0xFF}, // 5  洋紅
	{0xAA, 0x55, 0x00, 0xFF}, // 6  棕
	{0xAA, 0xAA, 0xAA, 0xFF}, // 7  淺灰
	{0x55, 0x55, 0x55, 0xFF}, // 8  深灰
	{0x55, 0x55, 0xFF, 0xFF}, // 9  亮藍
	{0x55, 0xFF, 0x55, 0xFF}, // 10 亮綠
	{0x55, 0xFF, 0xFF, 0xFF}, // 11 亮青
	{0xFF, 0x55, 0x55, 0xFF}, // 12 亮紅
	{0xFF, 0x55, 0xFF, 0xFF}, // 13 亮洋紅
	{0xFF, 0xFF, 0x55, 0xFF}, // 14 黃
	{0xFF, 0xFF, 0xFF, 0xFF}, // 15 白
}

// tileColorRemap 把 **FM Towns** `EGA*.TIL` 裡的色號換成標準 EGA 色號。
//
// FM Towns 那份用的是 **IGRB** 順序(bit2 = G、bit1 = R),與標準 EGA 的
// **IRGB** 差在 R 與 G 對調 —— 所以 remap 就是「把 bit1 與 bit2 互換」,
// 而且它是自己的反函數。
//
// 這條結論走過兩個階段,值得留著當對照:
//
// **第一階段(2026-08-07 上午,語意推導)**——當時手上只有 FM Towns 一個來源。
// 先用標準 EGA palette 畫出 tileset 與世界地圖,再逐 tile 對照語意:
//
//	tile 1–3   黑底 + 藍波紋   水    色號 1/9   → 互換後不變 ✓(本來就對)
//	tile 11–13 灰色山脈        山    色號 7/8   → 互換後不變 ✓(本來就對)
//	tile 16+   黃屋頂、白城堡  建築  色號 14/15 → 互換後不變 ✓(本來就對)
//	tile 5–10  草地與森林      應綠  色號 4/12  → 互換後 4→綠、12→亮綠 ✓(修正)
//
// 凡「本來就對」的顏色在此變換下都不動,而唯一錯的那一類正好被修好 ——
// 單一變換同時修正多處而不弄壞任何一處,這種一致性就是當時的證據。
// 但文件也誠實標了:那是語意推導,不是對第二個來源核實過。
//
// **第二階段(同日下午,兩個來源互證)**——`TILES.16` 的 LZW 壓縮破了之後,
// DOS 版的 tile 資料直接攤在眼前。結果是:
//
//	DOS 解出來的 65,536 B  ==  FM Towns 四檔降回 16×16、套上這張 remap、壓回 4bpp
//
// **逐位元組完全相同**(`TestDOSTilesMatchFMTowns`)。這張表不再是推導,
// 是兩份獨立資料之間量出來的差。
//
// ⇒ 順帶確認 **DOS 的色號本來就是標準 EGA**,不需要任何轉換。
var tileColorRemap = [16]byte{
	0, 1, 4, 5, // 0000→0000, 0001→0001, 0010→0100, 0011→0101
	2, 3, 6, 7, // 0100→0010, 0101→0011, 0110→0110, 0111→0111
	8, 9, 12, 13, // 1000→1000, 1001→1001, 1010→1100, 1011→1101
	10, 11, 14, 15, // 1100→1010, 1101→1011, 1110→1110, 1111→1111
}

// FM Towns 版 EGA*.TIL 的格式(靜態推導 + 降採樣一致性驗證,2026-08-07)
//
// 每個檔案 65,536 B,內含 128 個 tile,每個 tile 是**原版 16×16 機械放大 2 倍**後的
// 32×32 4bpp packed 點陣(每列 16 B × 32 列 = 512 B)。
//
// 支持這個解讀的證據:
//   - 檔內位元組只有 15–16 種唯一值,而且全部是「重複 nibble」(0x00/0x22/0xAA/0xEE/0xFF…)。
//     4bpp packed 若沒有水平 2× 放大,圖形邊緣必然出現混合 nibble(如 0x2A)——實測一個都沒有。
//   - 相鄰列成對相同 → 垂直也是 2×。
//   - 4 檔 × 128 tile = 512 tile,而降回 16×16 後 512 × 128 B = 65,536 B,
//     **正好等於 DOS 版 TILES.16 檔頭宣稱的解壓後長度** —— 兩個獨立來源互相印證。
//
// 因為是機械放大,資訊量等於原版 16×16,所以本解碼器直接降回 16×16 存放:
// 一來與 DOS 版共用同一個 Tile 型別,二來降採樣時可以檢查「每個 2×2 區塊是否真的同色」,
// 當成格式假設的驗證條件(FM Towns 若其實是重繪過的高解圖,這裡就會報錯而不是安靜吃下去)。
const (
	fmTownsTilesPerFile = 128
	fmTownsTileBytes    = (TileSize * 2) * (TileSize * 2) / 2 // 512
)

// ParseFMTownsTiles 解析一個 FM Towns EGA*.TIL(128 個 tile)。
func ParseFMTownsTiles(raw []byte) ([]Tile, error) {
	want := fmTownsTilesPerFile * fmTownsTileBytes
	if len(raw) != want {
		return nil, fmt.Errorf("EGA*.TIL 大小 %d B,預期 %d B(128 tile × 512 B)", len(raw), want)
	}
	tiles := make([]Tile, fmTownsTilesPerFile)
	const srcRowBytes = TileSize * 2 / 2 // 放大後一列 32 px = 16 B

	for i := range tiles {
		base := i * fmTownsTileBytes
		for y := 0; y < TileSize; y++ {
			// 垂直 2×:只取每兩列的第一列。
			row := raw[base+(y*2)*srcRowBytes : base+(y*2+1)*srcRowBytes]
			// 同時檢查被丟掉的那一列是否真的相同。
			dup := raw[base+(y*2+1)*srcRowBytes : base+(y*2+2)*srcRowBytes]
			for k := range row {
				if row[k] != dup[k] {
					return nil, fmt.Errorf(
						"tile %d 第 %d 列與下一列不同(垂直 2× 放大的假設不成立):"+
							"FM Towns 這份圖可能是重繪過的,不能當 DOS tileset 的 oracle",
						i, y*2)
				}
			}
			for x := 0; x < TileSize; x++ {
				// 水平 2×:放大後的兩個像素被 pack 進同一個 byte,高低 nibble 應相同。
				b := row[x]
				hi, lo := b>>4, b&0x0F
				if hi != lo {
					return nil, fmt.Errorf(
						"tile %d (%d,%d) 的位元組 %#02x 高低 nibble 不同"+
							"(水平 2× 放大的假設不成立)", i, x, y, b)
				}
				// 存進去之前先換成標準 EGA 色號 —— Tile.Pix 全 engine 一種意思。
				tiles[i].Pix[y*TileSize+x] = tileColorRemap[hi]
			}
		}
	}
	return tiles, nil
}

// LoadFMTownsTileSet 依序讀 EGA0–EGA3.TIL,合成完整的 512 個 tile。
func LoadFMTownsTileSet(paths []string) ([]Tile, error) {
	if len(paths) != 4 {
		return nil, fmt.Errorf("需要 4 個檔(EGA0–EGA3.TIL),給了 %d 個", len(paths))
	}
	all := make([]Tile, 0, TileCount)
	for _, p := range paths {
		raw, err := os.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("讀取 %s: %w", p, err)
		}
		tiles, err := ParseFMTownsTiles(raw)
		if err != nil {
			return nil, fmt.Errorf("解析 %s: %w", p, err)
		}
		all = append(all, tiles...)
	}
	if len(all) != TileCount {
		return nil, fmt.Errorf("合成後 %d 個 tile,預期 %d", len(all), TileCount)
	}
	return all, nil
}

// Pack4bpp 把 tile 壓回 DOS 版的 4bpp packed 佈局(每列 8 B、共 128 B)。
//
// 這是破解 DOS TILES.16 壓縮時的對答案基準:解壓器的輸出應該與
// 「FM Towns tileset 全部 Pack4bpp 串起來」逐位元組相同(共 65,536 B)。
func (t *Tile) Pack4bpp() []byte {
	out := make([]byte, tileBytes4bpp)
	for y := 0; y < TileSize; y++ {
		for x := 0; x < TileSize; x += 2 {
			out[y*TileSize/2+x/2] = t.Pix[y*TileSize+x]<<4 | t.Pix[y*TileSize+x+1]&0x0F
		}
	}
	return out
}

// TileSheet 把一批 tile 排成一張圖,供目視驗收(cols 個一列)。
func TileSheet(tiles []Tile, cols int) *image.NRGBA {
	if cols <= 0 {
		cols = 16
	}
	rows := (len(tiles) + cols - 1) / cols
	img := image.NewNRGBA(image.Rect(0, 0, cols*TileSize, rows*TileSize))
	for i := range tiles {
		ox, oy := (i%cols)*TileSize, (i/cols)*TileSize
		for y := 0; y < TileSize; y++ {
			for x := 0; x < TileSize; x++ {
				img.SetNRGBA(ox+x, oy+y, EGAPalette[tiles[i].At(x, y)&0x0F])
			}
		}
	}
	return img
}

// DOS 版 TILES.16
//
// 破解過程見 `internal/u5data/lzw.go`:就是標準 LZW(LSB-first、9→12 bit、
// 不做 early change),檔頭 4 B 是解壓後長度 65,536 = 512 tile × 128 B。
//
// 每個 tile 是 16×16 的 4bpp packed(每列 8 B),色號**已經是標準 EGA**。

// tileBytesDOS 是一個 DOS tile 的位元組數。
const tileBytesDOS = TileSize * TileSize / 2

// ParseDOSTiles 把解壓後的 TILES.16 內容切成 512 個 tile。
func ParseDOSTiles(raw []byte) ([]Tile, error) {
	want := TileCount * tileBytesDOS
	if len(raw) != want {
		return nil, fmt.Errorf("解壓後 %d B,預期 %d B(%d tile × %d B)",
			len(raw), want, TileCount, tileBytesDOS)
	}
	tiles := make([]Tile, TileCount)
	for i := range tiles {
		base := i * tileBytesDOS
		for y := 0; y < TileSize; y++ {
			for x := 0; x < TileSize; x += 2 {
				b := raw[base+y*TileSize/2+x/2]
				tiles[i].Pix[y*TileSize+x] = b >> 4
				tiles[i].Pix[y*TileSize+x+1] = b & 0x0F
			}
		}
	}
	return tiles, nil
}

// LoadDOSTileSet 從 gamedata 目錄讀 TILES.16 並解壓成 512 個 tile。
func LoadDOSTileSet(dir string) ([]Tile, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "TILES.16"))
	if err != nil {
		return nil, err
	}
	out, err := Decompress(raw)
	if err != nil {
		return nil, fmt.Errorf("TILES.16: %w", err)
	}
	return ParseDOSTiles(out)
}
