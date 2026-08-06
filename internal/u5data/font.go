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
