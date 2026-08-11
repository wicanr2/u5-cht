package render

import (
	"image"
	"testing"
)

func TestClassicChromeLeavesFooterForHints(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, CanvasWidth, CanvasHeight))
	fill(img, ColorBackground)

	s := &Scene{UI: UIOriginal}
	s.drawClassicChrome(img)

	if got := img.At(PanelX, HintY); got != ColorBackground {
		t.Fatalf("古典框仍畫進命令提示 footer: y=%d, got=%v", HintY, got)
	}
	if got := img.At(PanelX, HintY-classicFooterGap-1); got == ColorBackground {
		t.Fatalf("古典框未在 footer 前收邊: y=%d", HintY-classicFooterGap-1)
	}
}
