package render

import (
	"image"
	"testing"

	"github.com/wicanr2/u5-cht/internal/game"
)

// 確認框關著時,畫面要與完全沒有這個功能時**一模一樣**。
//
// ★ 這條看起來多餘,但它擋的是「dimAll 在旗標關著時也跑了」這種
// 只在視覺上看得出來的迴歸 —— 而視覺迴歸沒有測試就等於沒人看。
func TestQuitDialogOffChangesNothing(t *testing.T) {
	a := testScene().Render()
	s := testScene()
	s.QuitAsk = false
	b := s.Render()
	if !sameImage(a, b) {
		t.Fatal("QuitAsk=false 時畫面不該有任何差別")
	}
}

func TestQuitDialogDimsAndDrawsBox(t *testing.T) {
	base := testScene().Render()
	s := testScene()
	s.QuitAsk = true
	got := s.Render()

	// ① 角落被壓暗 —— 那裡一定在框外面。
	corner := image.Pt(2, 2)
	before := base.NRGBAAt(corner.X, corner.Y)
	after := got.NRGBAAt(corner.X, corner.Y)
	if after.R > before.R || after.G > before.G || after.B > before.B {
		t.Fatalf("角落沒有被壓暗:%v → %v", before, after)
	}

	// ② 框的邊線畫出來了 —— 取上緣正中央那一點。
	bx := (CanvasWidth - quitBoxWidth) / 2
	by := (CanvasHeight - quitBoxHeight) / 2
	edge := got.NRGBAAt(bx+quitBoxWidth/2, by)
	if edge != colorQuitFrame {
		t.Fatalf("框的上緣不是邊框色:%v", edge)
	}

	// ③ 框裡面**沒有**被壓暗成一片黑 —— 底色要是 `colorQuitFill`。
	inner := got.NRGBAAt(bx+4, by+quitBoxHeight-4)
	if inner != colorQuitFill {
		t.Fatalf("框內底色不對:%v(預期 %v)", inner, colorQuitFill)
	}
}

// ★★ 框要蓋在**每一種**畫面之上。原本的 `Render` 有五條 early return
//(開場動畫、主選單、Ztats、選單、正常畫面),漏掉任一條的症狀就是
// 「在某個畫面按 F10 沒反應」—— 而那個畫面通常不是開發時最常看的那個。
func TestQuitDialogCoversEveryScreen(t *testing.T) {
	cases := map[string]func(*Scene){
		"正常畫面": func(*Scene) {},
		"開場動畫": func(s *Scene) { s.State.Intro = &game.Intro{} },
		"Ztats": func(s *Scene) { s.State.Prompt = game.PromptZtats },
		"選單":   func(s *Scene) { s.State.Prompt = game.PromptPick },
	}
	bx := (CanvasWidth - quitBoxWidth) / 2
	by := (CanvasHeight - quitBoxHeight) / 2
	for name, setup := range cases {
		s := testScene()
		setup(s)
		s.QuitAsk = true
		got := s.Render()
		if edge := got.NRGBAAt(bx+quitBoxWidth/2, by); edge != colorQuitFrame {
			t.Fatalf("%s:確認框沒有疊上去(上緣是 %v)", name, edge)
		}
	}
}

func sameImage(a, b *image.NRGBA) bool {
	if a.Bounds() != b.Bounds() {
		return false
	}
	for i := range a.Pix {
		if a.Pix[i] != b.Pix[i] {
			return false
		}
	}
	return true
}
