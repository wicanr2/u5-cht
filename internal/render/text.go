// Package render 是畫面組裝層:tile 地圖、HUD、混合 ASCII/CJK 的文字。
//
// [架構決策 2026-08-07] 本套件**不依賴 ebiten**,全部畫在 image.NRGBA 上。
//
// 原因是踩到的坑:一開始把繪製綁在 ebiten 上,結果 headless 截圖必須起 xvfb + 軟體 GL,
// 而那個組合會死鎖 —— 實測卡了五小時、沒有任何輸出,最後只能砍掉容器
// (u4-cht 也踩過同一類問題:「F2 圖形熱切換在軟體 GL 死鎖」)。
// 加逾時或換 GL 後端都只是治標;根因是「驗證畫面」本來就不該需要 GPU。
//
// 改成單一 CPU 繪製路徑之後:
//   - headless 驗證是秒級的純函式呼叫,CI 不需要顯示環境
//   - 只有一份繪製實作 ⇒ 截圖與實機畫面保證一致,不會漂移
//   - ebiten 那層只剩「把 640×400 的成品上傳成紋理 + 整數倍放大 + 收鍵盤」
//
// 每幀重畫 640×400(約 256K 像素)對回合制 retro 遊戲綽綽有餘,
// 而且畫面只在狀態改變時才需要重畫。
package render

import (
	"image"
	"image/color"

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
	charset *u5data.Charset
	cjkFont *cjk.Font
	fg      color.NRGBA

	// 字模比行高矮時垂直居中用的位移。
	cjkOffsetY   int
	asciiOffsetY int
}

// NewTextRenderer 建立文字繪製器。任一字型可為 nil(缺 CJK 時中文會畫成紅框)。
func NewTextRenderer(charset *u5data.Charset, cjkFont *cjk.Font, fg color.NRGBA) *TextRenderer {
	t := &TextRenderer{
		charset:      charset,
		cjkFont:      cjkFont,
		fg:           fg,
		asciiOffsetY: (LineHeight - u5data.GlyphHeight) / 2,
	}
	if cjkFont != nil {
		_, gh := cjkFont.GlyphSize()
		t.cjkOffsetY = (LineHeight - gh) / 2
	}
	return t
}

// Draw 從 (x, y) 畫一行字,回傳畫完後的 x。
func (t *TextRenderer) Draw(dst *image.NRGBA, x, y int, s string) int {
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

// DrawLines 從 (x, y) 逐行畫,回傳畫完後的 y。
func (t *TextRenderer) DrawLines(dst *image.NRGBA, x, y int, lines []string) int {
	for _, ln := range lines {
		t.Draw(dst, x, y, ln)
		y += LineHeight
	}
	return y
}

func (t *TextRenderer) drawASCII(dst *image.NRGBA, x, y int, c byte) {
	if t.charset == nil {
		return
	}
	g, ok := t.charset.Glyph(c)
	if !ok {
		return
	}
	oy := y + t.asciiOffsetY
	for gy := 0; gy < u5data.GlyphHeight; gy++ {
		for gx := 0; gx < u5data.GlyphWidth; gx++ {
			if g.At(gx, gy) {
				SetPixel(dst, x+gx, oy+gy, t.fg)
			}
		}
	}
}

func (t *TextRenderer) drawCJK(dst *image.NRGBA, x, y int, r rune) {
	if t.cjkFont == nil {
		t.drawMissingBox(dst, x, y)
		return
	}
	rect, ok := t.cjkFont.Glyph(r)
	if !ok {
		// 缺字畫成紅框而不是靜默跳過 —— 畫面上看得見才會有人去修。
		t.drawMissingBox(dst, x, y)
		return
	}
	oy := y + t.cjkOffsetY
	mask := t.cjkFont.Mask
	for my := rect.Min.Y; my < rect.Max.Y; my++ {
		for mx := rect.Min.X; mx < rect.Max.X; mx++ {
			if v, _, _, _ := mask.At(mx, my).RGBA(); v > 0x7FFF {
				SetPixel(dst, x+(mx-rect.Min.X), oy+(my-rect.Min.Y), t.fg)
			}
		}
	}
}

func (t *TextRenderer) drawMissingBox(dst *image.NRGBA, x, y int) {
	c := color.NRGBA{R: 0xAA, G: 0x00, B: 0x00, A: 0xFF}
	w, h := CJKAdvance, LineHeight-2
	for i := 0; i < w; i++ {
		SetPixel(dst, x+i, y+1, c)
		SetPixel(dst, x+i, y+h, c)
	}
	for j := 1; j <= h; j++ {
		SetPixel(dst, x, y+j, c)
		SetPixel(dst, x+w-1, y+j, c)
	}
}

// SetPixel 是帶邊界檢查的畫點。
func SetPixel(dst *image.NRGBA, x, y int, c color.NRGBA) {
	if !(image.Point{X: x, Y: y}).In(dst.Bounds()) {
		return
	}
	dst.SetNRGBA(x, y, c)
}

// DrawFrame 畫一個空心矩形(玩家位置標記等)。
func DrawFrame(dst *image.NRGBA, x, y, w, h int, c color.NRGBA) {
	for i := 0; i < w; i++ {
		SetPixel(dst, x+i, y, c)
		SetPixel(dst, x+i, y+h-1, c)
	}
	for j := 0; j < h; j++ {
		SetPixel(dst, x, y+j, c)
		SetPixel(dst, x+w-1, y+j, c)
	}
}
