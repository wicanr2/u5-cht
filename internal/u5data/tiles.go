package u5data

import (
	"strings"
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
//
// ★ **色號 6 是暗黃(0xAAAA00),不是一般說的棕(0xAA5500)。**
// 這不是選擇,是原版自己載進去的:`EGA.DRV` 與 `T1K.DRV` 都用
// `int 10h` AH=10h/AL=02h 一次載入 17 B 的調色盤表,而那張表在
// `DATA.OVL` 0x52EE(`EGAPaletteTableOffset`),第 7 個值是 **0x06**。
// EGA 的調色盤值裡棕色是 **0x14**(次級綠位元),0x06 是「紅 + 綠皆為主級」= 暗黃。
//
// 色號 6 佔整份 tileset 的 **7.70%**(黑與白之後第三多)—— 木地板、門、桌子、
// 箱子、泥土全是它,所以這一格錯了每一張畫面都會偏色。
//
// 其餘 15 格與那張表逐一相符(`TestEGAPaletteMatchesTheTableInDataOVL`),
// 所以這份常數現在是**從原版的表推導出來的**,不是抄慣例。
var EGAPalette = [16]color.NRGBA{
	{0x00, 0x00, 0x00, 0xFF}, // 0  黑
	{0x00, 0x00, 0xAA, 0xFF}, // 1  藍
	{0x00, 0xAA, 0x00, 0xFF}, // 2  綠
	{0x00, 0xAA, 0xAA, 0xFF}, // 3  青
	{0xAA, 0x00, 0x00, 0xFF}, // 4  紅
	{0xAA, 0x00, 0xAA, 0xFF}, // 5  洋紅
	{0xAA, 0xAA, 0x00, 0xFF}, // 6  暗黃 ★ 不是棕 —— 見 EGAPaletteTableOffset
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

// DOS 版 TILES.4 —— CGA 的 2bpp tileset
//
// 四種顯示模式的驅動程式一開頭就把答案講完了(`CGA.DRV` / `EGA.DRV` /
// `T1K.DRV` 各自的 `int 10h` 序列,`HER.DRV` 直接寫 CRTC):
//
//	CGA.DRV   mov ah,0 / mov al,4  / int 10h   → 模式 4(320×200 四色)
//	          mov bl,1 / mov bh,1  / mov ah,0Bh / int 10h  → **調色盤 1**
//	          xor bl,bl / xor bh,bh / mov ah,0Bh / int 10h → 背景 / 邊框 = 黑
//	          mov ah,5 / xor al,al / int 10h              → 顯示頁 0
//	EGA.DRV   mov al,0Dh → 模式 0Dh(320×200 十六色),再 AH=10h/AL=02h 載調色盤
//	T1K.DRV   mov al,09h → 模式 9(Tandy 320×200 十六色),同樣 AH=10h/AL=02h
//	HER.DRV   沒有 int 10h;直接寫 0x3B4 / 0x3B8 / 0x3BF(Hercules 單色 720×348)
//
// ⇒ **Tandy 與 EGA 讀同一批 `.16` 素材**(都是十六色、都用 BIOS 載調色盤),
// 差別在硬體調色盤;CGA 讀 `.4`;Hercules 是單色,沒有自己的 tileset
//(`IBM.HCS` / `RUNES.HCS` 各 3,072 B 是**字型**,不是圖磚)。
//
// `TILES.4` 的檔頭宣稱解壓後 32,768 B = 512 tile × 64 B = 16×16 的 **2bpp**,
// 剛好是 `TILES.16` 的一半 —— 兩個數字互相印證,壓縮格式與 `.16` 相同。

// tileBytesCGA 是一個 CGA tile 的位元組數(16×16 × 2bpp)。
const tileBytesCGA = TileSize * TileSize / 4

// CGAColorCount 是 CGA 模式 4 的色數。
const CGAColorCount = 4

// CGAPaletteIndex 是 CGA 調色盤 1(低亮度)在十六色盤裡的色號。
//
// `int 10h` AH=0Bh 的兩次呼叫:BH=1/BL=1 選調色盤 1、BH=0/BL=0 把背景設成黑
//(BL 的 bit 3 是亮度位元,這裡是 0 → **低亮度**)。
// 調色盤 1 低亮度的三個前景色就是十六色盤的青(3)、洋紅(5)、淺灰(7)。
//
// ⚠ 不要寫成高亮度的 11 / 13 / 15 —— 那要 BL 帶 bit 3,而原版沒有。
var CGAPaletteIndex = [CGAColorCount]byte{0, 3, 5, 7}

// CGAPalette 是 CGA 的四個顏色,直接從十六色盤取 —— 兩者共用同一份 RGB。
var CGAPalette = [CGAColorCount]color.NRGBA{
	EGAPalette[CGAPaletteIndex[0]],
	EGAPalette[CGAPaletteIndex[1]],
	EGAPalette[CGAPaletteIndex[2]],
	EGAPalette[CGAPaletteIndex[3]],
}

// ParseCGATiles 把解壓後的 TILES.4 內容切成 512 個 tile。
//
// 色號是 0..3(CGA 的四色),**不是** EGA 的 0..15 —— 要畫出來請用 CGAPalette,
// 或用 CGAToEGA 轉成十六色盤的色號。
func ParseCGATiles(raw []byte) ([]Tile, error) {
	want := TileCount * tileBytesCGA
	if len(raw) != want {
		return nil, fmt.Errorf("解壓後 %d B,預期 %d B(%d tile × %d B)",
			len(raw), want, TileCount, tileBytesCGA)
	}
	tiles := make([]Tile, TileCount)
	for i := range tiles {
		base := i * tileBytesCGA
		for y := 0; y < TileSize; y++ {
			for x := 0; x < TileSize; x += 4 {
				b := raw[base+y*TileSize/4+x/4]
				// 一個位元組四個像素,**高位在左**(與 4bpp 的 hi/lo 同一個方向)。
				tiles[i].Pix[y*TileSize+x] = (b >> 6) & 3
				tiles[i].Pix[y*TileSize+x+1] = (b >> 4) & 3
				tiles[i].Pix[y*TileSize+x+2] = (b >> 2) & 3
				tiles[i].Pix[y*TileSize+x+3] = b & 3
			}
		}
	}
	return tiles, nil
}

// LoadCGATileSet 從 gamedata 目錄讀 TILES.4 並解壓成 512 個 CGA tile。
func LoadCGATileSet(dir string) ([]Tile, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "TILES.4"))
	if err != nil {
		return nil, err
	}
	out, err := Decompress(raw)
	if err != nil {
		return nil, fmt.Errorf("TILES.4: %w", err)
	}
	return ParseCGATiles(out)
}

// CGAToEGA 把 CGA 色號(0..3)換成十六色盤的色號,方便與 EGA 走同一條算繪路徑。
func CGAToEGA(c byte) byte { return CGAPaletteIndex[c&3] }

// DisplayMode 是原版的四種顯示模式(對應四個 `.DRV`)。
//
// 只有前兩種在素材上分得開:EGA / Tandy 讀 `.16`(十六色),CGA 讀 `.4`(四色)。
// Hercules 是單色而且沒有自己的 tileset,所以還沒實作 —— 見 `docs/re/62`。
type DisplayMode int

const (
	// DisplayEGA 是 `EGA.DRV`:BIOS 模式 0Dh,320×200 十六色。
	DisplayEGA DisplayMode = iota
	// DisplayCGA 是 `CGA.DRV`:BIOS 模式 4,320×200 四色(調色盤 1、低亮度)。
	DisplayCGA
	// DisplayTandy 是 `T1K.DRV`:BIOS 模式 9,320×200 十六色。
	//
	// ⚠ 素材與 EGA **同一批 `.16`**,兩者都用 `int 10h` AH=10h/AL=02h 載調色盤;
	// 差別在硬體調色盤的內容,還沒抽出來,所以目前與 EGA 同畫面。
	DisplayTandy
	// DisplayHercules 是 `HER.DRV`:不走 BIOS,直接寫 0x3B4 / 0x3B8 / 0x3BF。
	//
	// 單色。素材是 **`.4`(2bpp)**,四個色號各對一個位元圖樣
	//(`HerculesPattern`)—— 見該變數的說明。
	// ⚠ 完整的 blit 迴圈(縱向放大、位元加倍表怎麼配合)還沒追完。
	DisplayHercules
	// DisplayModeCount 是模式數。
	DisplayModeCount
)

// Name 回傳模式的顯示名。
func (m DisplayMode) Name() string {
	switch m {
	case DisplayEGA:
		return "EGA"
	case DisplayCGA:
		return "CGA"
	case DisplayTandy:
		return "Tandy"
	case DisplayHercules:
		return "Hercules"
	}
	return "?"
}

// Implemented 回報這個模式的素材解得出來沒有。
//
// 四種都解得出來了。⚠ Hercules 用的是原版自己的圖樣表,但 blit 迴圈還沒追完,
// 所以它**不是逐像素重現** —— 這一點在 `HerculesPattern` 的說明裡寫清楚了。
func (m DisplayMode) Implemented() bool { return m >= 0 && m < DisplayModeCount }

// ParseDisplayMode 把命令列上的字串轉成模式(不分大小寫)。
func ParseDisplayMode(s string) (DisplayMode, error) {
	for m := DisplayMode(0); m < DisplayModeCount; m++ {
		if strings.EqualFold(s, m.Name()) {
			return m, nil
		}
	}
	return DisplayEGA, fmt.Errorf("不認得的顯示模式 %q(可用:EGA / CGA / Tandy / Hercules)", s)
}

// LoadTileSetFor 依顯示模式讀 tileset,色號一律回十六色盤的 0..15。
//
// CGA 的 tile 本身只有 0..3,這裡透過 CGAToEGA 換成十六色色號 ——
// 這樣算繪層只需要一條路徑,不必為四色模式另寫一份。
func LoadTileSetFor(dir string, m DisplayMode) ([]Tile, error) {
	if m == DisplayHercules {
		return LoadHerculesTileSet(dir)
	}
	if m != DisplayCGA {
		return LoadDOSTileSet(dir)
	}
	tiles, err := LoadCGATileSet(dir)
	if err != nil {
		return nil, err
	}
	for i := range tiles {
		for j, p := range tiles[i].Pix {
			tiles[i].Pix[j] = CGAToEGA(p)
		}
	}
	return tiles, nil
}

// 原版載入的 EGA / Tandy 調色盤表
//
// `EGA.DRV` 與 `T1K.DRV` 的初始化序列在設完模式之後都做同一件事:
//
//	pop     dx              ; DX = 呼叫端傳進來的 BX
//	add     dx, 24h         ; DX = BX + 0x24
//	push    ds / pop es     ; ES = DS
//	mov     al, 2
//	mov     ah, 10h
//	int     10h             ; 一次載入 ES:DX 起的 17 B 調色盤表
//
// ★ **表的位址是呼叫端給的,不在驅動程式裡** —— 而兩個驅動程式算的都是
// `BX + 0x24`、呼叫端又是同一段遊戲程式碼。所以 **EGA 與 Tandy 載的是同一張表**:
// 這不是「比對兩張表發現一樣」,是**根本只有一張**。
// 兩者的差別只在 BIOS 模式(0Dh vs 9),畫面顏色相同。
//
// 表本體在 `DATA.OVL` 0x52EE,內容是:
//
//	00 01 02 03 04 05 06 07 38 39 3A 3B 3C 3D 3E 3F   16 個調色盤暫存器
//	00                                                overscan(邊框)
//
// # 三條獨立佐證,不是「找到一段看起來像的資料」
//
// 1. 全部 17 個值都 ≤ 0x3F(調色盤值只有 6 bit),而且這段前後都是 0 填充;
//    整個 `gamedata/` 只有這一處出現這個 pattern。
// 2. 它正是 EGA 十六色的標準表 —— **只有第 7 個值不同**(0x06 而非慣例的 0x14)。
// 3. ★ 反推 `BX = 0x52EE − 0x24 = 0x52CA`,那裡是一個顯示參數結構:
//
//	+0x00  0
//	+0x02  8, 8            ← 字元格 8×8
//	+0x06  16, 16          ← 圖磚 16×16
//	+0x18  0x013F = 319    ← 螢幕最大 X(320 − 1)
//	+0x1C  0x00C7 = 199    ← 螢幕最大 Y(200 − 1)
//	+0x24  調色盤表        ← 驅動程式算的正是這個位移
//
//    319×199 與 8×8 / 16×16 一次全部對上 —— 這塊就是驅動程式收到的那個結構,
//    而調色盤表確實掛在它的 +0x24。
const (
	// EGAPaletteTableOffset 是那張 17 B 表在 `DATA.OVL` 裡的位移。
	EGAPaletteTableOffset = 0x52EE
	// EGAPaletteTableSize 是表的長度(16 個暫存器 + 1 個 overscan)。
	EGAPaletteTableSize = 17
)

// EGAColorFromValue 把一個 EGA 調色盤值(6 bit)換成 RGB。
//
// 位元佈局是 `r g b R G B`:bit 5/4/3 是次級(secondary)的紅綠藍,
// bit 2/1/0 是主級(primary)。每個通道兩個位元 → 四級:
//
//	00 → 0x00    01(只有次級)→ 0x55    10(只有主級)→ 0xAA    11 → 0xFF
//
// 所以 0x06(000110)是紅與綠皆主級 = 暗黃,而棕色是 0x14(010100:紅主級 + 綠次級)。
func EGAColorFromValue(v byte) color.NRGBA {
	level := func(primary, secondary bool) byte {
		switch {
		case primary && secondary:
			return 0xFF
		case primary:
			return 0xAA
		case secondary:
			return 0x55
		}
		return 0x00
	}
	return color.NRGBA{
		R: level(v&0x04 != 0, v&0x20 != 0),
		G: level(v&0x02 != 0, v&0x10 != 0),
		B: level(v&0x01 != 0, v&0x08 != 0),
		A: 0xFF,
	}
}

// ParseEGAPaletteTable 從 `DATA.OVL` 的內容讀出原版載入的 16 色調色盤。
//
// 回傳的第 17 個值(overscan / 邊框色)另外給,因為它不是圖磚用的顏色。
func ParseEGAPaletteTable(ovl []byte) (pal [16]color.NRGBA, overscan color.NRGBA, err error) {
	end := EGAPaletteTableOffset + EGAPaletteTableSize
	if len(ovl) < end {
		return pal, overscan, fmt.Errorf("DATA.OVL 只有 %d B,放不下 0x%X 的調色盤表",
			len(ovl), EGAPaletteTableOffset)
	}
	raw := ovl[EGAPaletteTableOffset:end]
	// 調色盤值只有 6 bit —— 有超出的就是位移不對。
	for i, v := range raw {
		if v > 0x3F {
			return pal, overscan, fmt.Errorf("第 %d 個調色盤值是 0x%02X,超過 6 bit", i, v)
		}
	}
	for i := 0; i < 16; i++ {
		pal[i] = EGAColorFromValue(raw[i])
	}
	return pal, EGAColorFromValue(raw[16]), nil
}

// LoadEGAPalette 從 `DATA.OVL` 讀出原版的調色盤。
func LoadEGAPalette(dir string) ([16]color.NRGBA, color.NRGBA, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "DATA.OVL"))
	if err != nil {
		var pal [16]color.NRGBA
		return pal, color.NRGBA{}, err
	}
	return ParseEGAPaletteTable(raw)
}

// Hercules 的單色轉換(`HER.DRV`)
//
// `HER.DRV` 沒有任何 `int 10h` —— Hercules 不在 BIOS 支援範圍內,它直接寫暫存器。
// 需要的東西全部在檔案裡找得到,而且**都不是猜的**:
//
//	0x070C  12 B   `35 2D 2E 07 5B 02 57 57 02 03 00 00`
//	               ← **標準 Hercules 圖形模式的 6845 參數表**,一個位元組都不差
//	0x071A         `mov al,1 / mov dx,3BFh / out dx,al`  ← 設定暫存器:允許圖形模式
//	0x0734         `mov dx,3B8h / out dx,al`             ← 模式暫存器
//	0x0318   4 B   `00 55 AA FF`   ← ★ 色號 → 位元圖樣
//	0x031C  16 B   `00 03 0C 0F 30 33 3C 3F C0 C3 CC CF F0 F3 FC FF`
//	               ← 「每個 bit 複製成兩個」的表 = **橫向 2 倍**
//	0x032C   8 B   `80 40 20 10 08 04 02 01`  ← 單一位元遮罩
//
// 用到 0x0318 的有兩處,而**兩處都是在講顏色**:
//
//	#15 @0x09E4  設定畫筆顏色
//	    and ax, 3 / mov bx, 0318h / xlat / mov cs:[0334], al
//	#31 @0x1282  畫字(前景 / 背景各一個色號)
//	    and dx, 0303h            ; ★ DL 與 DH 各取**低 2 bit**
//	    mov bx, 0318h
//	    mov al, dl / xlat / mov cs:[130E], al   ; 前景圖樣
//	    mov al, dh / xlat / mov cs:[130F], al   ; 背景圖樣
//
// 兩處都把色號夾成 **2 bit** → Hercules 的顏色空間就是 2bpp 的那四色,
// 與 `TILES.4` 的色數一致。所以「Hercules 沒有自己的 tileset,那它畫什麼」
// 這個問題的答案是 **CGA 那一份**。
//
// ⚠ **更正**:此前這裡寫「`and dx, 0303h` 是關鍵,所以 Hercules 吃 `.4`」——
// 那條推論引錯了常式。`and dx,0303h` 是**字的前景 / 背景色**,與圖磚資料無關;
// 而驅動程式唯一的圖形 blit(#26 @0x118C)收的是 **1bpp**。
// 「2bpp → 單色」的轉換**不在驅動裡**,在遊戲本體某個 `.OVL`(見 `hercules.go` §五)。
// 結論(色數是 4)沒變,但支持它的是那兩道 `and`,不是圖磚路徑。
//
// 四個圖樣的語意很直白:
//
//	色號 0 → 0x00  全暗
//	色號 1 → 0x55  隔一點亮(50%,相位 A)
//	色號 2 → 0xAA  隔一點亮(50%,**相位 B**)
//	色號 3 → 0xFF  全亮
//
// 中間兩色都是 50% 灰但**相位相反** —— 所以色號 1 與 2 的區域相鄰時分得出來。
// 而且**四個圖樣就是全部**:若還要依列號交替相位,表就得有八筆。
//
// blit 迴圈與整個畫面幾何(640×300 置中在 720×348、橫 2 倍縱 1.5 倍、
// 加倍表與圖樣表為什麼從不一起用)已經追完,寫在 **`hercules.go`**。
//
// 這裡只需要知道一件事:驅動程式對遊戲呈現的是 **320×200 / 40×25**,
// 與其他三種模式一模一樣,放大全部藏在驅動裡。所以 `HerculesTile` 產出
// **1× 的 16×16 單色圖**是對的抽象層 —— 每個像素取「圖樣在該 x 位置的那個 bit」。
//
// ⚠ 仍未證實的是**「2bpp → 單色」這步發生在哪**(驅動只收 1bpp,所以在遊戲本體),
// 以及是否逐列換相位。`hercules.go` §五 有記帳。

// HerculesPatternCount 是圖樣表的筆數(= 2bpp 的色數)。
const HerculesPatternCount = 4

// HerculesPattern 是 `HER.DRV` 0x0318 的四筆圖樣。
var HerculesPattern = [HerculesPatternCount]byte{0x00, 0x55, 0xAA, 0xFF}

// HerculesBit 回報 2bpp 色號 c 在 x 這一欄是亮還是暗。
//
// 圖樣是 8 個像素一輪(MSB 在左),與 `HER.DRV` 的 `xlat` 取出來那個位元組同一個意思。
func HerculesBit(c byte, x int) bool {
	p := HerculesPattern[c&3]
	return p&(1<<(7-uint(x&7))) != 0
}

// HerculesTile 把一個 2bpp 的 tile 轉成單色:亮的像素給色號 15、暗的給 0。
//
// 回傳的 Tile 走的還是十六色色號(黑與白),所以算繪層不必為單色另開一條路。
func HerculesTile(t Tile) Tile {
	var out Tile
	for y := 0; y < TileSize; y++ {
		for x := 0; x < TileSize; x++ {
			if HerculesBit(t.Pix[y*TileSize+x], x) {
				out.Pix[y*TileSize+x] = 15
			}
		}
	}
	return out
}

// LoadHerculesTileSet 讀 `TILES.4` 並轉成單色。
//
// ⚠ 這是**依 `HER.DRV` 的圖樣表**做出來的,但 blit 迴圈還沒追完(見上面的說明)——
// 所以它不等於原版 Hercules 畫面的逐像素重現。
func LoadHerculesTileSet(dir string) ([]Tile, error) {
	tiles, err := LoadCGATileSet(dir)
	if err != nil {
		return nil, err
	}
	for i := range tiles {
		tiles[i] = HerculesTile(tiles[i])
	}
	return tiles, nil
}
