package render

import (
	"image"
	"testing"
)

func TestBothUIModesLeaveFooterForHints(t *testing.T) {
	for _, mode := range []UIMode{UIModern, UIOriginal} {
		img := image.NewNRGBA(image.Rect(0, 0, CanvasWidth, CanvasHeight))
		fill(img, ColorBackground)

		s := &Scene{UI: mode}
		s.drawClassicChrome(img)

		if got := img.At(PanelX, HintY); got != ColorBackground {
			t.Errorf("%s 框線仍畫進命令提示 footer: y=%d, got=%v", UIModeNames[mode], HintY, got)
		}
		if got := img.At(PanelX, HintY-classicFooterGap-1); got == ColorBackground {
			t.Errorf("%s 框線未在 footer 前收邊: y=%d", UIModeNames[mode], HintY-classicFooterGap-1)
		}
	}
}
