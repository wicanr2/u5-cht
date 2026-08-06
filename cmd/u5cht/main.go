// u5cht 是 Ultima V 重製版(繁體中文)的執行檔。
//
// P0 階段的範圍刻意很小:開一個 640×400 的邏輯畫布、用**原版字型**畫出字元表,
// 證明「Ebiten 視窗 + 原版資料解碼 + 整數倍 nearest 放大」這條管線是通的。
// 中文點陣字(倚天)在 P1.5 進來,tile 地圖在 P1,遊戲邏輯在 P2 之後。
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 邏輯畫布 = 原版 320×200 的乾淨 2 倍。底圖用 nearest 整數放大,
// 中文才有空間用正常點陣尺寸(見 CLAUDE.md §3、rulebook/81)。
const (
	logicalWidth  = 640
	logicalHeight = 400

	// glyphAdvance 是原版 8×8 字元的格推進量(格間留 1 px)。
	glyphAdvance = u5data.GlyphWidth + 1
)

var (
	colorBackground = color.NRGBA{R: 0x10, G: 0x10, B: 0x28, A: 0xFF} // 近 EGA 深藍
	colorText       = color.NRGBA{R: 0xE8, G: 0xE8, B: 0xD8, A: 0xFF}
)

// version 由建置時的 -ldflags 注入。
var version = "dev"

type game struct {
	canvas *ebiten.Image // 640×400 的邏輯畫布

	charset      *u5data.Charset // 原版 8×8 字型(可能為 nil:玩家沒放素材)
	charsetAtlas *ebiten.Image   // 由 charset 烘成的 atlas,首次 Draw 時建立
	status       string          // 英文狀態列(中文要等 P1.5 的 CJK 點陣路徑)
}

func (g *game) Update() error {
	// 離開語意:F10 / Ctrl+Q 才離開,ESC 永遠是取消(P4 補確認框與自動存檔)。
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControlLeft) || ebiten.IsKeyPressed(ebiten.KeyControlRight)
	if inpututil.IsKeyJustPressed(ebiten.KeyF10) || (ctrl && inpututil.IsKeyJustPressed(ebiten.KeyQ)) {
		return ebiten.Termination
	}
	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	if g.canvas == nil {
		g.canvas = ebiten.NewImage(logicalWidth, logicalHeight)
	}
	// atlas 延後到這裡才烘:Ebiten 的圖形資源在遊戲迴圈啟動後才保證可用。
	if g.charsetAtlas == nil && g.charset != nil {
		g.charsetAtlas = bakeCharsetAtlas(g.charset)
	}

	g.canvas.Fill(colorBackground)
	g.drawText(4, 4, "ULTIMA V: WARRIORS OF DESTINY")
	g.drawText(4, 14, "TRADITIONAL CHINESE REMAKE - PHASE 0 ("+version+")")
	g.drawText(4, 32, g.status)

	if g.charsetAtlas != nil {
		g.drawText(4, 52, "ORIGINAL 8X8 CHARSET (IBM.CH):")
		// 把 0x20–0x7F 排成三列,肉眼即可確認字型解碼正確。
		for i := 0; i < 96; i++ {
			c := byte(0x20 + i)
			g.drawGlyph(4+(i%32)*glyphAdvance, 64+(i/32)*(u5data.GlyphHeight+2), c)
		}
	}
	g.drawText(4, logicalHeight-12, "F10 / CTRL+Q TO QUIT")

	// 邏輯畫布 → 視窗:整數倍 nearest,pixel art 不糊。
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	scale := min(sw/logicalWidth, sh/logicalHeight)
	if scale < 1 {
		scale = 1
	}
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
	op.GeoM.Scale(float64(scale), float64(scale))
	op.GeoM.Translate(
		float64((sw-logicalWidth*scale)/2),
		float64((sh-logicalHeight*scale)/2),
	)
	screen.DrawImage(g.canvas, op)
}

// Layout 回傳視窗實際大小 —— 縮放由 Draw 自己管,才能保證 nearest 而非線性。
func (g *game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func (g *game) drawGlyph(x, y int, c byte) {
	if g.charsetAtlas == nil || c >= 128 {
		return
	}
	src := g.charsetAtlas.SubImage(image.Rect(
		int(c)*u5data.GlyphWidth, 0,
		(int(c)+1)*u5data.GlyphWidth, u5data.GlyphHeight,
	)).(*ebiten.Image)
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
	op.GeoM.Translate(float64(x), float64(y))
	g.canvas.DrawImage(src, op)
}

func (g *game) drawText(x, y int, s string) {
	for i := 0; i < len(s); i++ {
		g.drawGlyph(x+i*glyphAdvance, y, s[i])
	}
}

// bakeCharsetAtlas 把 128 個 8×8 字形橫排成一張 1024×8 的紋理。
func bakeCharsetAtlas(cs *u5data.Charset) *ebiten.Image {
	w := len(cs.Glyphs) * u5data.GlyphWidth
	img := image.NewNRGBA(image.Rect(0, 0, w, u5data.GlyphHeight))
	for i, gl := range cs.Glyphs {
		for y := 0; y < u5data.GlyphHeight; y++ {
			for x := 0; x < u5data.GlyphWidth; x++ {
				if gl.At(x, y) {
					img.SetNRGBA(i*u5data.GlyphWidth+x, y, colorText)
				}
			}
		}
	}
	return ebiten.NewImageFromImage(img)
}

func main() {
	gamedata := flag.String("gamedata", "gamedata",
		"原版 Ultima V(DOS 版)資料目錄;版權素材由玩家自備,不隨本專案散布")
	scale := flag.Int("scale", 2, "視窗放大倍率(整數;邏輯畫布固定 640×400)")
	showVersion := flag.Bool("version", false, "印出版本後結束")
	flag.Parse()

	if *showVersion {
		fmt.Printf("u5cht %s\n", version)
		return
	}

	g := &game{}

	// 字型是 P0 唯一會用到的原版資料。找不到時不要當錯誤 ——
	// 優雅降級並明說缺什麼,不拿自製素材充數(CLAUDE.md §3.0)。
	charsetPath := filepath.Join(*gamedata, "IBM.CH")
	if cs, err := u5data.LoadCharset(charsetPath); err != nil {
		g.status = "ORIGINAL CHARSET NOT FOUND - SEE CONSOLE"
		fmt.Fprintf(os.Stderr,
			"找不到原版字型 %s(%v)\n"+
				"請把 DOS 版 Ultima V 的資料檔放進 %s/,或用 -gamedata 指定目錄。\n"+
				"素材為版權所有,需自備合法副本;本專案不隨附。\n",
			charsetPath, err, *gamedata)
	} else {
		g.charset = cs
		g.status = fmt.Sprintf("CHARSET OK - %d GLYPHS 8X8", len(cs.Glyphs))
		fmt.Printf("已載入原版字型 %s(%d 個 8×8 字形)\n", charsetPath, len(cs.Glyphs))
	}

	if *scale < 1 {
		*scale = 1
	}
	ebiten.SetWindowSize(logicalWidth**scale, logicalHeight**scale)
	ebiten.SetWindowTitle("創世紀 V:命運勇士 — 繁體中文版")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(g); err != nil {
		fmt.Fprintf(os.Stderr, "執行失敗:%v\n", err)
		os.Exit(1)
	}
}
