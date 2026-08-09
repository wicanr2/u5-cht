package render

import (
	"image"
	"image/color"
)

// 離開確認框(疊在整個畫面上的 modal)
//
// ⚠ **這個框是本重製版加的,不是原版的畫面。** 原版按 `Q` 就是
// `Quit & Save` 直接結束,沒有確認這一步。加它的理由與規則見
// `internal/appui/quit.go` 的檔頭 —— ESC 是取消、F10 才離開、離開前自動存檔。
//
// 版面刻意與遊戲的訊息欄不同(置中、有邊框、背景壓暗):玩家要一眼看出
// 「這不是遊戲裡的問句,這是要不要離開」。用訊息欄問的話,它會混在
// 「是否離開此地?」(`PromptLeave`,那是原版的機制)旁邊,兩個都是
// 「Y / N」而意思差很多。

const (
	quitBoxWidth  = 380
	quitBoxHeight = 5 * LineHeight
)

// 確認框的用字。**Y / Enter 與 N / ESC 都要寫出來** ——
// 「確定 / 取消」兩顆模糊按鈕是反模式(玩家不知道該按哪個鍵)。
const (
	quitTitle  = "確定離開遊戲?"
	quitNote   = "離開前會自動存檔。"
	quitChoice = "Y / Enter 存檔離開    N / ESC 取消"
)

var (
	colorQuitFrame = color.NRGBA{R: 0xFF, G: 0xD0, B: 0x40, A: 0xFF}
	colorQuitFill  = color.NRGBA{R: 0x18, G: 0x18, B: 0x30, A: 0xFF}
)

// drawQuitDialog 把背景壓暗,然後在正中央畫確認框。
func (s *Scene) drawQuitDialog(dst *image.NRGBA) {
	if s.Text == nil {
		return
	}
	dimAll(dst)
	x := (CanvasWidth - quitBoxWidth) / 2
	y := (CanvasHeight - quitBoxHeight) / 2
	fillRect(dst, x, y, quitBoxWidth, quitBoxHeight, colorQuitFill)
	DrawFrame(dst, x, y, quitBoxWidth, quitBoxHeight, colorQuitFrame)
	ty := y + LineHeight/2
	for _, line := range []string{quitTitle, quitNote, "", quitChoice} {
		if line == "" {
			ty += LineHeight / 2
			continue
		}
		// 每一行各自置中 —— 中文與 ASCII 的寬度不同,不能用固定縮排。
		lx := x + (quitBoxWidth-Width(line))/2
		s.Text.Draw(dst, lx, ty, line)
		ty += LineHeight
	}
}

// dimAll 把整張畫面壓暗,當 modal 的 scrim。
//
// ⚠ 這裡是**就地改像素**而不是疊一層半透明色:`fillRect` 只會畫實色,
// 疊上去就把背後的畫面蓋掉了 —— 而 scrim 的用意正是「還看得見背後,
// 但明顯退到後面」。除以 3 是實際看過的結果,比 /2 更能讓框跳出來。
func dimAll(dst *image.NRGBA) {
	b := dst.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := dst.NRGBAAt(x, y)
			dst.SetNRGBA(x, y, color.NRGBA{R: c.R / 3, G: c.G / 3, B: c.B / 3, A: c.A})
		}
	}
}
