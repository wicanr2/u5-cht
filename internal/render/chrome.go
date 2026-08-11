package render

import (
	"image"
	"image/color"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

const (
	classicChromeTop    = 8
	classicFooterGap    = 2
	classicChromeHeight = HintY - classicChromeTop - classicFooterGap
)

// drawClassicChrome 畫 UIOriginal 的原版藍白框線。
//
// 這裡只畫外框，不重排右欄內容：中文狀態與 remake 擴充訊息仍由既有
// adaptive layout 決定，避免為了像素框線退回原版狹窄的英文欄位。
func (s *Scene) drawClassicChrome(dst *image.NRGBA) {
	if s.UI != UIOriginal {
		return
	}

	// MapOriginY=16 正好把頂端 16 px 留給日月帶，框線不會蓋住第一列地形。
	// 古典框在命令提示前收邊，下面的 footer 留給繁中提示列，避免底框穿過字形。
	drawClassicFrame(dst, 0, classicChromeTop, PanelX, classicChromeHeight)
	drawClassicFrame(dst, PanelX-8, classicChromeTop,
		CanvasWidth-PanelX+8, classicChromeHeight)
}

// drawClassicFrame 是原版 EGA 藍／亮藍／白三層框線的 2× 畫布版本。
func drawClassicFrame(dst *image.NRGBA, x, y, w, h int) {
	if w < 16 || h < 16 {
		return
	}
	drawFrameBand(dst, x, y, w, h, 2, u5data.EGAPalette[1])
	drawFrameBand(dst, x+2, y+2, w-4, h-4, 2, u5data.EGAPalette[9])
	drawFrameBand(dst, x+6, y+6, w-12, h-12, 2, u5data.EGAPalette[15])
}

func drawFrameBand(dst *image.NRGBA, x, y, w, h, thickness int, c color.NRGBA) {
	for t := 0; t < thickness; t++ {
		for px := x + t; px < x+w-t; px++ {
			SetPixel(dst, px, y+t, c)
			SetPixel(dst, px, y+h-1-t, c)
		}
		for py := y + t; py < y+h-t; py++ {
			SetPixel(dst, x+t, py, c)
			SetPixel(dst, x+w-1-t, py, c)
		}
	}
}
