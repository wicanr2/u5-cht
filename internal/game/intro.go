package game

import (
	"strings"

	"github.com/wicanr2/u5-cht/internal/i18n"
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 開場動畫
//
// 21 頁,每頁一張插圖加一段文字。頁表與「哪一頁配哪張圖」來自原版
//(`u5data.IntroPages`,依據見那邊的說明);文字來自 `STORY.DAT`,
// 譯文走 `internal/i18n`。
//
// 原版每頁停下來等一個按鍵(`sub_27814` 的忙等)。這裡也是:
// `Prompt = PromptIntro`,按任意鍵翻頁,ESC 直接跳過整段。
//
// ⚠ **這不是 debug 捷徑。** 玩家跳過開場之後仍然要走完建角流程 ——
// 跳過的只有敘事,不是任何遊戲狀態(`rulebook` 的教訓:
// 用 debug hook 串起來的「能跑完」會遮住真 bug)。

// Intro 是進行中的開場動畫。
type Intro struct {
	// Page 是現在停在第幾頁(0..IntroPageCount−1)。
	Page int
	// Lines 是這一頁的文字,已經照畫面寬度斷好行。
	Lines []string
	// Scroll 是文字框從第幾行開始顯示。
	//
	// ⚠ 有幾頁的文字比框高 —— 尤其中文比英文多佔行。**先把同一頁的文字捲完
	// 才翻頁**,不然玩家會看不到後半段而完全不知道自己漏了。
	Scroll int
}

// VisibleLines 是文字框現在該顯示的那幾行。
func (in *Intro) VisibleLines() []string {
	if in.Scroll >= len(in.Lines) {
		return nil
	}
	end := in.Scroll + IntroLinesPerScreen
	if end > len(in.Lines) {
		end = len(in.Lines)
	}
	return in.Lines[in.Scroll:end]
}

// MoreOnThisPage 回報同一頁還有沒有沒顯示完的文字。
func (in *Intro) MoreOnThisPage() bool {
	return in.Scroll+IntroLinesPerScreen < len(in.Lines)
}

// IntroText 取第 n 頁的文字(已經是譯文)。
//
// 第 6 頁的文字寫死在原版執行檔裡,不在 `STORY.DAT` —— 走另一組 key。
func (s *State) IntroText(page int) string {
	if page < 0 || page >= u5data.IntroPageCount {
		return ""
	}
	p := u5data.IntroPages[page]
	if p.Record < 0 {
		return i18n.Text("INTRO", 0, u5data.IntroHardcoded[0]) + "\n\n" +
			i18n.Text("INTRO", 1, u5data.IntroHardcoded[1])
	}
	en := ""
	if s.Story != nil && p.Record < len(s.Story.Records) {
		en = s.Story.Records[p.Record].Text()
	}
	return i18n.Text("STORY.DAT", p.Record, en)
}

// BeginIntro 從第 0 頁開始播開場。
func (s *State) BeginIntro() bool {
	if s.Story == nil {
		// 沒有 STORY.DAT 就不假裝有開場 —— 誠實跳過比放一段空白好。
		return false
	}
	s.Intro = &Intro{}
	s.Prompt = PromptIntro
	s.setIntroPage(0)
	return true
}

// AdvanceIntro 翻到下一頁。翻完就結束並回 false。
func (s *State) AdvanceIntro() bool {
	if s.Intro == nil {
		return false
	}
	// 同一頁還有沒讀完的文字就先捲,不要翻頁。
	if s.Intro.MoreOnThisPage() {
		s.Intro.Scroll += IntroLinesPerScreen
		return true
	}
	if s.Intro.Page+1 >= u5data.IntroPageCount {
		s.EndIntro()
		return false
	}
	s.setIntroPage(s.Intro.Page + 1)
	return true
}

// SkipIntro 直接跳過整段開場(原版沒有,但按 ESC 是這個引擎的一貫語意)。
func (s *State) SkipIntro() { s.EndIntro() }

// EndIntro 收掉開場。
func (s *State) EndIntro() {
	s.Intro = nil
	if s.Prompt == PromptIntro {
		s.Prompt = PromptNone
	}
}

// setIntroPage 換頁並重新斷行。
func (s *State) setIntroPage(n int) {
	s.Intro.Page = n
	s.Intro.Scroll = 0
	s.Intro.Lines = IntroWrap(s.IntroText(n), IntroColumns)
}

// IntroColumns 是開場文字一行放幾個中文字。
//
// 畫布 640 寬、左右各留 24 px,中文字 16 px → (640−48)/16 = 37,取 36 留一點餘裕。
const IntroColumns = 36

// IntroLinesPerScreen 是文字框一次放得下幾行(框高 138 px / 行高 16 px,留一行給提示)。
const IntroLinesPerScreen = 7

// IntroWrap 把一段文字斷成行。
//
// ⚠ **中文不能照英文的規則斷**:原版靠空白斷字(而且資料裡還有 `_` 斷字提示),
// 中文沒有詞邊界,是照**字數**斷的。所以這裡數的是 rune 不是 byte,
// 而且不在標點前換行(避免行首出現逗號句號)。
//
// 原文裡的 `\n` 是原版排版留下的硬換行,一律當成段落分隔照舊處理 ——
// 譯文已經按中文的節奏重下了 `\n`,不再沿用英文的位置。
func IntroWrap(text string, cols int) []string {
	var out []string
	for _, para := range strings.Split(text, "\n") {
		rs := []rune(strings.TrimRight(para, " "))
		if len(rs) == 0 {
			out = append(out, "")
			continue
		}
		for len(rs) > 0 {
			n := cols
			if n > len(rs) {
				n = len(rs)
			}
			// 避免把標點推到下一行的行首。
			for n < len(rs) && isTrailingPunct(rs[n]) {
				n++
			}
			out = append(out, string(rs[:n]))
			rs = rs[n:]
		}
	}
	return out
}

// isTrailingPunct 回報這個字元不該出現在行首。
func isTrailingPunct(r rune) bool {
	switch r {
	case '。', ',', '、', ';', ':', '!', '?', '」', '』', ')', '】', '…', '·':
		return true
	}
	return false
}
