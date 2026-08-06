// Package render 是 Ebiten 繪圖層:tile 地圖、HUD、混合 ASCII/CJK 的文字。
package render

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/u5-cht/internal/cjk"
	"github.com/wicanr2/u5-cht/internal/textlayout"
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 文字度量一律走 internal/textlayout —— render 只負責畫,不自己算寬度。
// (量測與繪製同一個 gate;見 textlayout 套件說明。)
const (
	ASCIIAdvance = textlayout.ASCIIAdvance
	CJKAdvance   = textlayout.CJKAdvance
	LineHeight   = textlayout.LineHeight
)

// Advance / Width / Wrap 轉呼叫 textlayout,方便呼叫端只 import render。
func Advance(r rune) int            { return textlayout.Advance(r) }
func Width(s string) int            { return textlayout.Width(s) }
func Wrap(s string, w int) []string { return textlayout.Wrap(s, w) }

// TextRenderer 畫混合 ASCII 與 CJK 的文字。
type TextRenderer struct {
	asciiTex *ebiten.Image // 原版 8×8 字型烘成的 atlas(128 glyph 橫排)
	cjkTex   *ebiten.Image // 倚天 atlas
	cjkFont  *cjk.Font

	// cjkOffsetY 讓 15 px 高的字模在 16 px 行高裡垂直居中。
	cjkOffsetY int
}

// NewTextRenderer 建立文字繪製器。cjkFont 可為 nil(那時 CJK 會畫成缺字框)。
func NewTextRenderer(charset *u5data.Charset, cjkFont *cjk.Font, fg color.NRGBA) *TextRenderer {
	t := &TextRenderer{cjkFont: cjkFont}

	if charset != nil {
		w := len(charset.Glyphs) * u5data.GlyphWidth
		img := image.NewNRGBA(image.Rect(0, 0, w, u5data.GlyphHeight))
		for i, g := range charset.Glyphs {
			for y := 0; y < u5data.GlyphHeight; y++ {
				for x := 0; x < u5data.GlyphWidth; x++ {
					if g.At(x, y) {
						img.SetNRGBA(i*u5data.GlyphWidth+x, y, fg)
					}
				}
			}
		}
		t.asciiTex = ebiten.NewImageFromImage(img)
	}

	if cjkFont != nil {
		_, gh := cjkFont.GlyphSize()
		t.cjkOffsetY = (LineHeight - gh) / 2
		// atlas 是單色遮罩:把亮點染成前景色。
		b := cjkFont.Mask.Bounds()
		img := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				if r, _, _, _ := cjkFont.Mask.At(x, y).RGBA(); r > 0x7FFF {
					img.SetNRGBA(x-b.Min.X, y-b.Min.Y, fg)
				}
			}
		}
		t.cjkTex = ebiten.NewImageFromImage(img)
	}
	return t
}

// Draw 從 (x, y) 畫一行字,回傳畫完後的 x。
func (t *TextRenderer) Draw(dst *ebiten.Image, x, y int, s string) int {
	for _, r := range s {
		if r < 0x80 {
			t.drawASCII(dst, x, y, byte(r))
		} else {
			t.drawCJK(dst, x, y, r)
		}
		x += Advance(r)
	}
	return x
}

func (t *TextRenderer) drawASCII(dst *ebiten.Image, x, y int, c byte) {
	if t.asciiTex == nil || c >= 128 {
		return
	}
	// ASCII 字模只有 8 px 高,在 16 px 行高裡垂直居中。
	sub := t.asciiTex.SubImage(image.Rect(
		int(c)*u5data.GlyphWidth, 0,
		(int(c)+1)*u5data.GlyphWidth, u5data.GlyphHeight,
	)).(*ebiten.Image)
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
	op.GeoM.Translate(float64(x), float64(y+(LineHeight-u5data.GlyphHeight)/2))
	dst.DrawImage(sub, op)
}

func (t *TextRenderer) drawCJK(dst *ebiten.Image, x, y int, r rune) {
	if t.cjkTex == nil || t.cjkFont == nil {
		return
	}
	rect, ok := t.cjkFont.Glyph(r)
	if !ok {
		// 缺字畫成空框而不是靜默跳過 —— 畫面上看得見才會有人去修。
		t.drawMissingBox(dst, x, y)
		return
	}
	sub := t.cjkTex.SubImage(rect).(*ebiten.Image)
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
	op.GeoM.Translate(float64(x), float64(y+t.cjkOffsetY))
	dst.DrawImage(sub, op)
}

func (t *TextRenderer) drawMissingBox(dst *ebiten.Image, x, y int) {
	gw, gh := CJKAdvance, LineHeight-2
	c := color.NRGBA{R: 0xAA, G: 0x00, B: 0x00, A: 0xFF}
	for i := 0; i < gw; i++ {
		dst.Set(x+i, y+1, c)
		dst.Set(x+i, y+gh, c)
	}
	for j := 1; j <= gh; j++ {
		dst.Set(x, y+j, c)
		dst.Set(x+gw-1, y+j, c)
	}
}
