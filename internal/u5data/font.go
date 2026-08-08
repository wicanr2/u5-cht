package u5data

import (
	"fmt"
	"os"
)

// GlyphWidth / GlyphHeight 是原版點陣字元的尺寸。
//
// 已驗證(2026-08-06,DOS 版 IBM.CH):檔案 1024 B = 128 glyph × 8 B,
// 每列 1 byte、MSB 在左、由上而下;索引就是 ASCII 碼(dump idx 65 得到 'A'、
// idx 64 得到 '@'、idx 33 得到 '!')。RUNES.CH 同格式(符文字型)。
const (
	GlyphWidth  = 8
	GlyphHeight = 8

	charsetGlyphCount = 128
	charsetFileSize   = charsetGlyphCount * GlyphHeight // 1024
)

// Glyph 是一個 8×8 單色點陣字元,每個元素是一列的 8 個 bit(MSB 在左)。
type Glyph [GlyphHeight]byte

// At 回報 (x, y) 這個像素是否點亮。座標超出範圍回 false。
func (g Glyph) At(x, y int) bool {
	if x < 0 || x >= GlyphWidth || y < 0 || y >= GlyphHeight {
		return false
	}
	return g[y]&(1<<(7-uint(x))) != 0
}

// Charset 是一整套原版點陣字型(IBM.CH 或 RUNES.CH)。
type Charset struct {
	// Glyphs 以 ASCII 碼為索引,長度固定 128。
	Glyphs []Glyph
}

// Glyph 取某個位元組對應的字形。超出 0x00–0x7F 回空字形與 false ——
// 這正是原版沒有 CJK 空間的原因:中文一律走另一條點陣路徑(見 internal/cjk)。
func (c *Charset) Glyph(b byte) (Glyph, bool) {
	if int(b) >= len(c.Glyphs) {
		return Glyph{}, false
	}
	return c.Glyphs[b], true
}

// LoadCharset 讀取 IBM.CH / RUNES.CH 這類 8×8 字型檔。
func LoadCharset(path string) (*Charset, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("讀取字型檔 %s: %w", path, err)
	}
	return ParseCharset(raw)
}

// ParseCharset 解析已讀入記憶體的 8×8 字型內容。
func ParseCharset(raw []byte) (*Charset, error) {
	if len(raw) != charsetFileSize {
		return nil, fmt.Errorf(
			"字型檔大小 %d B,預期 %d B(128 glyph × 8 B)——確認這不是 FM Towns U5.FNT 或 PC-98 FONT98.CH,那兩個佈局不同",
			len(raw), charsetFileSize)
	}
	cs := &Charset{Glyphs: make([]Glyph, charsetGlyphCount)}
	for i := range cs.Glyphs {
		copy(cs.Glyphs[i][:], raw[i*GlyphHeight:(i+1)*GlyphHeight])
	}
	return cs, nil
}

// ASCIIArt 把字形印成可目視檢查的字串(驗證 oracle 用:'A' 必須看起來像 A)。
func (g Glyph) ASCIIArt(on, off byte) string {
	buf := make([]byte, 0, GlyphHeight*(GlyphWidth+1))
	for y := 0; y < GlyphHeight; y++ {
		for x := 0; x < GlyphWidth; x++ {
			if g.At(x, y) {
				buf = append(buf, on)
			} else {
				buf = append(buf, off)
			}
		}
		buf = append(buf, '\n')
	}
	return string(buf)
}

// Hercules 的高解析字型 `IBM.HCS` / `RUNES.HCS`
//
// 兩個檔案各 3,072 B = 128 glyph × 24 B。24 B 的切法有兩種可能,
// **dump 出來用眼睛看就分得出來**:
//
//	8 × 24   glyph 65 是一團斜線,認不出字
//	16 × 12  glyph 65 是乾淨的 'A'   ← 就是這個
//
// 整張表(0x20..0x7F)畫出來是完整的 ASCII:空白、`!`、數字、大小寫字母、
// 以及最後幾格畫框用的圖形字元。索引與 `IBM.CH` 一樣**就是 ASCII 碼**。
//
// 幾何上也說得通:Hercules 是 720×348,720 / 16 = 45 欄、348 / 12 = 29 列;
// 相對於 DOS 的 8×8 @ 320×200 是**橫向 2 倍、縱向 1.5 倍** ——
// 而那正好是 320×200 撐到 640×300 塞進 720×348 的比例。
const (
	// HCSGlyphWidth / HCSGlyphHeight 是 Hercules 字型的尺寸。
	HCSGlyphWidth  = 16
	HCSGlyphHeight = 12

	// hcsBytesPerRow 是一列的位元組數(16 px = 2 B)。
	hcsBytesPerRow = HCSGlyphWidth / 8
	// hcsGlyphBytes 是一個 glyph 的位元組數。
	hcsGlyphBytes = HCSGlyphHeight * hcsBytesPerRow // 24
	// hcsFileSize 是一份 `.HCS` 的大小。
	hcsFileSize = charsetGlyphCount * hcsGlyphBytes // 3072
)

// HCSGlyph 是一個 16×12 單色點陣字元,每列兩個位元組(MSB 在左)。
type HCSGlyph [hcsGlyphBytes]byte

// At 回報 (x, y) 這個像素是否點亮。座標超出範圍回 false。
func (g HCSGlyph) At(x, y int) bool {
	if x < 0 || x >= HCSGlyphWidth || y < 0 || y >= HCSGlyphHeight {
		return false
	}
	b := g[y*hcsBytesPerRow+x/8]
	return b&(1<<(7-uint(x%8))) != 0
}

// HCSCharset 是一份 Hercules 字型(128 個 glyph,索引 = ASCII)。
type HCSCharset struct {
	Glyphs [charsetGlyphCount]HCSGlyph
}

// Glyph 依 ASCII 碼取字元;超出 128 回 false。
func (c *HCSCharset) Glyph(b byte) (HCSGlyph, bool) {
	if int(b) >= charsetGlyphCount {
		return HCSGlyph{}, false
	}
	return c.Glyphs[b], true
}

// ParseHCSCharset 解一份 `.HCS`。
func ParseHCSCharset(raw []byte) (*HCSCharset, error) {
	if len(raw) != hcsFileSize {
		return nil, fmt.Errorf("`.HCS` 是 %d B,預期 %d B(%d glyph × %d B)",
			len(raw), hcsFileSize, charsetGlyphCount, hcsGlyphBytes)
	}
	c := &HCSCharset{}
	for i := range c.Glyphs {
		copy(c.Glyphs[i][:], raw[i*hcsGlyphBytes:(i+1)*hcsGlyphBytes])
	}
	return c, nil
}

// LoadHCSCharset 讀一份 `.HCS`。
func LoadHCSCharset(path string) (*HCSCharset, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseHCSCharset(raw)
}

// ASCIIArt 把 glyph 畫成文字圖,方便在測試裡用眼睛看。
func (g HCSGlyph) ASCIIArt(on, off byte) string {
	out := make([]byte, 0, (HCSGlyphWidth+1)*HCSGlyphHeight)
	for y := 0; y < HCSGlyphHeight; y++ {
		for x := 0; x < HCSGlyphWidth; x++ {
			if g.At(x, y) {
				out = append(out, on)
			} else {
				out = append(out, off)
			}
		}
		out = append(out, '\n')
	}
	return string(out)
}
