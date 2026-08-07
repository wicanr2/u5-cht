package game

import (
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/i18n"
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// TestIntroPagesCoverEveryStoryRecord:21 頁要正好用完 `STORY.DAT` 的 20 筆。
//
// 頁表是從原版 dump 出來的,但「有沒有抄漏一頁」光看表看不出來。
// 這條反過來驗:每一筆記錄都有頁用到、而且沒有一筆被用兩次。
// 少一頁的症狀是**開場少講一段故事**,玩家不會知道自己漏了。
func TestIntroPagesCoverEveryStoryRecord(t *testing.T) {
	used := map[int]int{}
	hardcoded := 0
	for i, p := range u5data.IntroPages {
		if p.Record < 0 {
			hardcoded++
			if i != u5data.IntroHardcodedPage {
				t.Errorf("第 %d 頁的文字寫死,但寫死的應該只有第 %d 頁",
					i, u5data.IntroHardcodedPage)
			}
			continue
		}
		used[p.Record]++
	}
	if hardcoded != 1 {
		t.Errorf("有 %d 頁的文字寫死,原版只有一頁", hardcoded)
	}
	for r := 0; r < 20; r++ {
		switch used[r] {
		case 1:
		case 0:
			t.Errorf("STORY.DAT 第 %d 筆沒有任何一頁用到 —— 開場會少講一段", r)
		default:
			t.Errorf("STORY.DAT 第 %d 筆被 %d 頁用到", r, used[r])
		}
	}
	// 記錄序號必須遞增 —— 開場是照順序講的。
	prev := -1
	for _, p := range u5data.IntroPages {
		if p.Record < 0 {
			continue
		}
		if p.Record <= prev {
			t.Errorf("記錄序號從 %d 退回 %d —— 頁表抄亂了", prev, p.Record)
		}
		prev = p.Record
	}
}

// TestIntroArtReferencesAreInRange:每一頁指到的插圖都要真的存在。
func TestIntroArtReferencesAreInRange(t *testing.T) {
	dir := gameDataDir(t)
	if dir == "" {
		return
	}
	sets := make([]u5data.PictureSet, len(u5data.IntroStoryFiles))
	for i, name := range u5data.IntroStoryFiles {
		set, err := u5data.LoadPictures(dir + "/" + name)
		if err != nil {
			t.Fatal(err)
		}
		sets[i] = set
	}
	for i, p := range u5data.IntroPages {
		if p.Story < 0 || p.Story >= len(sets) {
			t.Fatalf("第 %d 頁指到第 %d 個插圖檔,只有 %d 個", i, p.Story, len(sets))
		}
		for _, shape := range []int{p.Shape, p.Shape2} {
			if shape < 0 {
				continue
			}
			set := sets[p.Story]
			if shape >= len(set) || set[shape] == nil {
				t.Errorf("第 %d 頁要 %s 的形狀 %d,但那個檔只有 %d 個形狀",
					i, u5data.IntroStoryFiles[p.Story], shape, len(set))
			}
		}
	}
}

// TestIntroTextIsChinese:每一頁都要拿得到中文。
//
// 包含那一頁**寫死在執行檔裡**的文字 —— 它不在 `STORY.DAT`,
// 只看資料檔會漏掉,而漏掉的表現是「開場中間突然冒出兩句英文」。
func TestIntroTextIsChinese(t *testing.T) {
	s := introState(t)
	if s == nil {
		return
	}
	for i := 0; i < u5data.IntroPageCount; i++ {
		txt := s.IntroText(i)
		if txt == "" {
			t.Errorf("第 %d 頁沒有文字", i)
			continue
		}
		if !hasHan(txt) {
			t.Errorf("第 %d 頁的文字沒有漢字:%.60s", i, txt)
		}
	}
	// 寫死那一頁的兩句都要在。
	hard := s.IntroText(u5data.IntroHardcodedPage)
	for _, en := range u5data.IntroHardcoded {
		if strings.Contains(hard, en) {
			t.Errorf("寫死的那一頁還留著英文原文:%q", en)
		}
	}
	if !strings.Contains(hard, i18n.Text("INTRO", 0, "")) {
		t.Error("寫死那一頁少了第一句")
	}
}

// TestIntroAdvancesThroughEveryPage:一路按下去要走完 21 頁才結束。
//
// ⚠ 文字比框高的那幾頁會**先捲完才翻頁**。這條同時驗兩件事:
// 捲動不會卡住(無限迴圈),而且捲動也不會讓某些頁被跳過。
func TestIntroAdvancesThroughEveryPage(t *testing.T) {
	s := introState(t)
	if s == nil {
		return
	}
	if !s.BeginIntro() {
		t.Fatal("開場播不起來")
	}
	seen := map[int]bool{0: true}
	steps := 0
	for s.Intro != nil {
		steps++
		if steps > 500 {
			t.Fatal("按了 500 次還沒播完 —— 大概卡在某一頁的捲動裡")
		}
		s.AdvanceIntro()
		if s.Intro != nil {
			seen[s.Intro.Page] = true
		}
	}
	if len(seen) != u5data.IntroPageCount {
		t.Errorf("只走過 %d 頁,共 %d 頁", len(seen), u5data.IntroPageCount)
	}
	if s.Prompt != PromptNone {
		t.Errorf("播完之後 Prompt 是 %v,應該回到 PromptNone", s.Prompt)
	}
}

// TestIntroWrapKeepsPunctuationOffLineStart:標點不能被推到行首。
//
// 中文沒有詞邊界,是照字數斷行的 —— 天真的實作會讓句號、逗號、
// 右引號跑到下一行開頭,那在中文排版裡很刺眼。
func TestIntroWrapKeepsPunctuationOffLineStart(t *testing.T) {
	lines := IntroWrap("這是一段夠長的中文,用來測試斷行會不會把標點推到行首。"+
		"再來一句,同樣夠長,結尾也有句號。還有第三句,加上引號「像這樣」。", 12)
	if len(lines) < 3 {
		t.Fatalf("只斷成 %d 行,測不出東西", len(lines))
	}
	for i, l := range lines {
		if l == "" {
			continue
		}
		if isTrailingPunct([]rune(l)[0]) {
			t.Errorf("第 %d 行以標點 %q 開頭:%q", i, string([]rune(l)[0]), l)
		}
	}
}

// TestSkipIntroLeavesNoState:ESC 跳過之後不能殘留狀態。
//
// ⚠ 跳過的只有敘事,**不是任何遊戲狀態** —— 玩家跳過開場之後仍要走完
// 該走的流程。這條盯的是 Prompt 有沒有還給遊戲。
func TestSkipIntroLeavesNoState(t *testing.T) {
	s := introState(t)
	if s == nil {
		return
	}
	s.BeginIntro()
	s.SkipIntro()
	if s.Intro != nil || s.Prompt != PromptNone {
		t.Errorf("跳過之後 Intro=%v Prompt=%v", s.Intro, s.Prompt)
	}
}

func hasHan(s string) bool {
	for _, r := range s {
		if r >= 0x3400 && r <= 0x9FFF {
			return true
		}
	}
	return false
}

// gameDataDir 回傳原版資料目錄;沒設就 skip 並回空字串。
func gameDataDir(t *testing.T) string {
	t.Helper()
	if os.Getenv("U5_GAMEDATA") == "" {
		t.Skip("未設 U5_GAMEDATA")
		return ""
	}
	return "../../gamedata"
}

// introState 準備一個載好 STORY.DAT 的 State。
func introState(t *testing.T) *State {
	t.Helper()
	dir := gameDataDir(t)
	if dir == "" {
		return nil
	}
	s := combatState(t)
	tf, err := u5data.LoadText(dir + "/STORY.DAT")
	if err != nil {
		t.Fatal(err)
	}
	s.Story = tf
	return s
}
