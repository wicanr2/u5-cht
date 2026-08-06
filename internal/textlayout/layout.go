// Package textlayout 是文字度量與斷行 —— 純邏輯,不依賴 ebiten。
//
// 抽出來的理由有兩個:
//  1. 度量與斷行可以在沒有顯示環境的地方測試(ebiten 的 init 需要 X11)。
//  2. [HARD] 量測與繪製必須走同一個 gate。把 Advance / LineHeight 放在單一處,
//     render 層只能引用、不能自己另算 —— 否則 Wrap 算出來的斷行會跟畫出來的不同,
//     文字溢出訊息框,而症狀看起來完全像「譯文太長」(這個坑 kb 記過)。
package textlayout

// ASCII 走原版 8×8 字型、CJK 走倚天 16×15。CJK 寬度剛好是 ASCII 的兩倍,
// 混排時欄位才對得齊,「幾格寬」的算術也保持整數。
const (
	ASCIIAdvance = 8
	CJKAdvance   = ASCIIAdvance * 2
	LineHeight   = 16
)

// Advance 回傳一個字佔的水平寬度 —— 量測與繪製共用的唯一來源。
func Advance(r rune) int {
	if r < 0x80 {
		return ASCIIAdvance
	}
	return CJKAdvance
}

// Width 回傳整串字的寬度。
func Width(s string) int {
	w := 0
	for _, r := range s {
		w += Advance(r)
	}
	return w
}

// Wrap 依「可用像素寬」斷行。用的是同一個 Advance,所以斷行結果與實際繪製一致。
//
// CJK 不做音節斷字(原版 .DAT 裡的 `_` 斷字提示在解碼時已移除);
// 英文則盡量在空白處斷,斷不了才硬斷。
func Wrap(s string, maxWidth int) []string {
	if maxWidth < CJKAdvance {
		return []string{s}
	}
	var lines []string
	var cur []rune
	curW := 0
	lastSpace, widthAtSpace := -1, 0

	flush := func() {
		if len(cur) > 0 {
			lines = append(lines, string(cur))
		}
		cur = nil
		curW = 0
		lastSpace, widthAtSpace = -1, 0
	}

	for _, r := range s {
		if r == '\n' {
			flush()
			continue
		}
		adv := Advance(r)
		if curW+adv > maxWidth {
			// 英文優先在最後一個空白斷,避免把單字切開。
			if lastSpace > 0 && lastSpace < len(cur) {
				head := string(cur[:lastSpace])
				tail := append([]rune{}, cur[lastSpace+1:]...)
				lines = append(lines, head)
				cur = tail
				curW = curW - widthAtSpace - ASCIIAdvance
			} else {
				flush()
			}
			lastSpace, widthAtSpace = -1, 0
		}
		if r == ' ' {
			lastSpace, widthAtSpace = len(cur), curW
		}
		cur = append(cur, r)
		curW += adv
	}
	flush()
	return lines
}
