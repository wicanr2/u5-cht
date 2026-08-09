package appui

import (
	"strings"
	"testing"
)

// F1 是開關:按一次開、再按一次收。
func TestHelpKeyToggles(t *testing.T) {
	var h HelpPanel
	if h.IsOpen() {
		t.Fatal("零值應該是關著的")
	}
	if !h.Step(Keys{Help: true}, 2) || !h.IsOpen() {
		t.Fatal("F1 沒有把說明打開")
	}
	if !h.Step(Keys{Help: true}, 2) || h.IsOpen() {
		t.Fatal("再按 F1 沒有收起")
	}
}

// ★ ESC 收起說明 —— 與 `QuitDialog` 同一條鐵則:ESC 只取消,永遠不結束遊戲。
func TestHelpEscapeClosesAndNeverQuits(t *testing.T) {
	var h HelpPanel
	h.Step(Keys{Help: true}, 2)
	if !h.Step(Keys{Escape: true}, 2) || h.IsOpen() {
		t.Fatal("ESC 沒有收起說明")
	}
	// 關著時按 ESC 什麼都不該發生(不是「再開一次」也不是別的)。
	if h.Step(Keys{Escape: true}, 2) {
		t.Error("關著時 ESC 不該有作用")
	}
}

func TestHelpPagesClampAtBothEnds(t *testing.T) {
	var h HelpPanel
	h.Step(Keys{Help: true}, 2)
	if h.Page() != 0 {
		t.Fatalf("開起來應該在第 0 頁,得到 %d", h.Page())
	}
	if h.Step(Keys{PageUp: true}, 2) {
		t.Error("第 0 頁還能往前翻")
	}
	if !h.Step(Keys{PageDown: true}, 2) || h.Page() != 1 {
		t.Fatalf("翻不到第 1 頁(現在 %d)", h.Page())
	}
	if h.Step(Keys{PageDown: true}, 2) {
		t.Error("最後一頁還能往後翻")
	}
	// ⚠ 收起來要**回到第 0 頁** —— 否則下次按 F1 會停在上次翻到的地方,
	// 而玩家按 F1 的意圖是「我要看說明」,不是「回到我上次讀到哪」。
	h.Close()
	h.Step(Keys{Help: true}, 2)
	if h.Page() != 0 {
		t.Errorf("重新打開應該回到第 0 頁,得到 %d", h.Page())
	}
}

// ★★ 說明表本身要**能回答問題**,不是只有列出來。
//
// 使用者問的具體問題是「開門是什麼指令」—— 所以這條測試就驗那件事:
// 表裡必須有一列同時提到「開門」與正確的鍵 `O`。
// 這比「表有 26 列」有意義:少了那一列,說明畫面就沒有達到它存在的目的。
func TestHelpAnswersHowToOpenADoor(t *testing.T) {
	var found *HelpEntry
	for _, sec := range HelpSections() {
		for i, e := range sec.Entries {
			if strings.Contains(e.Title, "開門") {
				found = &sec.Entries[i]
			}
		}
	}
	if found == nil {
		t.Fatal("說明表裡找不到「開門」—— 使用者問的就是這一題")
	}
	if found.Key != "O" {
		t.Errorf("開門寫成按 %q,原版是 O(Open)", found.Key)
	}
	if !strings.Contains(found.Note, "方向") {
		t.Errorf("開門的補充說明沒提到要接方向:%q", found.Note)
	}
}

// 每一列都要有鍵與標題,而且鍵不重複 —— 重複的鍵表示表抄錯了。
func TestHelpEntriesAreWellFormed(t *testing.T) {
	seen := map[string]bool{}
	for _, sec := range HelpSections() {
		if sec.Heading == "" {
			t.Error("有一段沒有標題")
		}
		for _, e := range sec.Entries {
			if e.Key == "" || e.Title == "" {
				t.Errorf("不完整的一列:%+v", e)
			}
			if seen[e.Key] {
				t.Errorf("鍵 %q 出現兩次", e.Key)
			}
			seen[e.Key] = true
		}
	}
	// 原版把 A–Z 全用掉了 —— 抽查幾個最常用的,漏了就是表不完整。
	for _, k := range []string{"O", "T", "G", "K", "Z", "E"} {
		if !seen[k] {
			t.Errorf("說明表漏了原版指令 %q", k)
		}
	}
}
