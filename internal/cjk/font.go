// Package cjk 載入由倚天(ETEN)點陣字烘出的 atlas,提供「碼點 → atlas 位置」查詢。
//
// 為什麼是倚天而不是 TTF:1990s DOS 中文遊戲的中文就長這樣;倚天的 16×15 / 24×24 是為
// 該尺寸手工調的點陣,TTF 縮到這個大小會糊、筆劃比例也不對。
// atlas 由 tools/build_eten_font.py 產生(**產物不入 git**,各自從自備字庫重跑)。
//
// 邊界:本套件不依賴 ebiten —— 回傳標準 image 與矩形,由 render 層轉成 GPU 紋理。
package cjk

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
)

// Font 是一套 CJK 點陣字。
type Font struct {
	// Mask 是 atlas 影像(單色;亮點 = 筆劃)。
	Mask image.Image

	glyphW, glyphH int
	cols           int
	index          map[rune]int // 碼點 → atlas 第幾格
}

type atlasMeta struct {
	Source      string `json:"source"`
	GlyphWidth  int    `json:"glyph_width"`
	GlyphHeight int    `json:"glyph_height"`
	Cols        int    `json:"cols"`
	Count       int    `json:"count"`
	Codepoints  []int  `json:"codepoints"`
}

// Load 讀取 atlas 的 .png 與 .json(共用同一個前綴)。
func Load(prefix string) (*Font, error) {
	metaRaw, err := os.ReadFile(prefix + ".json")
	if err != nil {
		return nil, fmt.Errorf("讀取字型索引 %s.json: %w(先跑 tools/dev.sh font 15 烘字型)", prefix, err)
	}
	var meta atlasMeta
	if err := json.Unmarshal(metaRaw, &meta); err != nil {
		return nil, fmt.Errorf("解析 %s.json: %w", prefix, err)
	}
	if meta.GlyphWidth <= 0 || meta.GlyphHeight <= 0 || meta.Cols <= 0 {
		return nil, fmt.Errorf("%s.json 的 glyph 尺寸或 cols 不合理", prefix)
	}
	if len(meta.Codepoints) != meta.Count {
		return nil, fmt.Errorf("%s.json 宣稱 %d 字但 codepoints 有 %d 個", prefix, meta.Count, len(meta.Codepoints))
	}

	f, err := os.Open(prefix + ".png")
	if err != nil {
		return nil, fmt.Errorf("讀取字型 atlas %s.png: %w", prefix, err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("解碼 %s.png: %w", prefix, err)
	}

	// atlas 尺寸要對得上索引數量,否則取字會取到隔壁格。
	rows := (meta.Count + meta.Cols - 1) / meta.Cols
	wantW, wantH := meta.Cols*meta.GlyphWidth, rows*meta.GlyphHeight
	if b := img.Bounds(); b.Dx() != wantW || b.Dy() != wantH {
		return nil, fmt.Errorf("atlas 是 %d×%d,但索引推算應為 %d×%d —— png 與 json 不同步",
			b.Dx(), b.Dy(), wantW, wantH)
	}

	fnt := &Font{
		Mask:   img,
		glyphW: meta.GlyphWidth,
		glyphH: meta.GlyphHeight,
		cols:   meta.Cols,
		index:  make(map[rune]int, meta.Count),
	}
	for i, cp := range meta.Codepoints {
		fnt.index[rune(cp)] = i
	}
	return fnt, nil
}

// GlyphSize 回傳字模尺寸。
func (f *Font) GlyphSize() (w, h int) { return f.glyphW, f.glyphH }

// Count 回傳收錄字數。
func (f *Font) Count() int { return len(f.index) }

// Glyph 回傳某個碼點在 atlas 中的矩形。找不到回 false ——
// 呼叫者要自行決定 fallback,不要靜默畫成空白(那會變成畫面上的空洞)。
func (f *Font) Glyph(r rune) (image.Rectangle, bool) {
	i, ok := f.index[r]
	if !ok {
		return image.Rectangle{}, false
	}
	x, y := (i%f.cols)*f.glyphW, (i/f.cols)*f.glyphH
	return image.Rect(x, y, x+f.glyphW, y+f.glyphH), true
}

// Has 回報字型是否收錄這個碼點。
func (f *Font) Has(r rune) bool {
	_, ok := f.index[r]
	return ok
}

// MissingRunes 回報字串裡沒收錄的字。
//
// 缺字數量是品質指標:Big5 沒有的字(簡體、罕用字)才該落到這裡。
// 若一大批字掉進來,先懷疑索引公式或漏帶 SPCFONT,不要無腦補字型。
func (f *Font) MissingRunes(s string) []rune {
	var out []rune
	seen := map[rune]bool{}
	for _, r := range s {
		if r < 0x80 || seen[r] {
			continue // ASCII 走原版 8×8 字型,不算缺字
		}
		if !f.Has(r) {
			seen[r] = true
			out = append(out, r)
		}
	}
	return out
}
