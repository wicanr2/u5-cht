package game

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMenuOrderMatchesTheOriginal:六項的順序照原版,不重排。
//
// 老遊戲玩家的肌肉記憶是「第幾項」。把常用的往上搬會踩到那個記憶 ——
// 那不是體貼是干擾,而且會讓對照原版錄影的驗收對不上。
func TestMenuOrderMatchesTheOriginal(t *testing.T) {
	want := []MenuItem{
		MenuJourneyOnward,    // Journey Onward
		MenuCreateCharacter,  // Create New Character
		MenuTransferU4,       // Transfer from Ultima IV
		MenuIntroduction,     // Ultima V Introduction
		MenuAcknowledgements, // Acknowledgements
		MenuReturnToView,     // Return to the View
	}
	if int(MenuItemCount) != len(want) {
		t.Fatalf("選單有 %d 項,原版是 %d 項", MenuItemCount, len(want))
	}
	for i, m := range want {
		if int(m) != i {
			t.Errorf("第 %d 項的常數值是 %d —— 順序被改過", i, m)
		}
		if MenuLabels[m] == "" {
			t.Errorf("第 %d 項沒有中文標籤", i)
		}
	}
}

// TestMenuCursorWraps:游標上下繞回。
func TestMenuCursorWraps(t *testing.T) {
	s := &State{MaxMessages: 8}
	s.BeginMainMenu()
	s.MenuMove(-1)
	if s.Menu.Cursor != MenuReturnToView {
		t.Errorf("從第一項往上該繞到最後一項,實際是 %d", s.Menu.Cursor)
	}
	s.MenuMove(1)
	if s.Menu.Cursor != MenuJourneyOnward {
		t.Errorf("再往下該回到第一項,實際是 %d", s.Menu.Cursor)
	}
}

// TestUnimplementedMenuItemsSaySo:沒做的項目要照實說,不能假裝有用。
//
// ⚠ 「從創世紀 IV 轉入」**已經實作了**(`docs/re/55`),所以從這張表裡移除;
// 但它在**沒給存檔路徑**時仍然要照實說,那一條移到下面單獨測。
func TestUnimplementedMenuItemsSaySo(t *testing.T) {
	for _, c := range []struct {
		item MenuItem
		want string
	}{
		{MenuAcknowledgements, MsgAcknowledgementsNotImplemented},
	} {
		s := &State{MaxMessages: 8}
		s.BeginMainMenu()
		s.Menu.Cursor = c.item
		if s.MenuChoose() {
			t.Errorf("第 %d 項沒實作,不該關掉選單", c.item)
		}
		if !s.InMainMenu() {
			t.Errorf("第 %d 項選完之後選單不見了", c.item)
		}
		if !strings.Contains(strings.Join(s.Messages, "|"), c.want) {
			t.Errorf("第 %d 項沒有說明未實作:%q", c.item, s.Messages)
		}
	}
}

// TestJourneyOnwardClosesTheMenu:「繼續前行」就是關掉選單開始玩。
func TestJourneyOnwardClosesTheMenu(t *testing.T) {
	s := &State{MaxMessages: 8}
	s.BeginMainMenu()
	if !s.MenuChoose() {
		t.Fatal("「繼續前行」該關掉選單")
	}
	if s.InMainMenu() || s.Prompt != PromptNone {
		t.Errorf("選單沒關乾淨:Prompt=%v Menu=%v", s.Prompt, s.Menu)
	}
}

// 「從創世紀 IV 轉入」在沒給存檔路徑時要照實說,不能假裝轉入。
//
// 原版寫死讀 `a:party.sav`;引擎讓 `-u4save` 指定。沒指定就沒東西可讀 ——
// 這時做一個「什麼都不做卻關掉選單」的分支,玩家會以為角色轉進來了。
func TestTransferWithoutAPathSaysSo(t *testing.T) {
	s := &State{MaxMessages: 8}
	s.BeginMainMenu()
	s.Menu.Cursor = MenuTransferU4
	if s.MenuChoose() {
		t.Error("沒給路徑卻關掉了選單")
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgTransferNeedsPath) {
		t.Errorf("沒有提示要給路徑:%q", s.Messages)
	}
}

// 給了一份壞掉的存檔,要照原版印同一句「無法完成轉入」,而且不關選單。
func TestTransferOnABadSaveKeepsTheMenu(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "PARTY.SAV")
	if err := os.WriteFile(bad, []byte{1, 2, 3}, 0o644); err != nil {
		t.Fatal(err)
	}
	s := &State{MaxMessages: 8, U4SavePath: bad}
	s.BeginMainMenu()
	s.Menu.Cursor = MenuTransferU4
	if s.MenuChoose() {
		t.Error("壞存檔卻關掉了選單")
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), "無法完成轉入") {
		t.Errorf("沒有印「無法完成轉入」:%q", s.Messages)
	}
}

// ★★ 主選單第二項要真的走到建角。
//
// 2026-08-09 實機回報:「五代沒有人物建立嗎?一開始就進入遊戲?」——
// 建角流程早就實作好、還有五條測試,但 **`cmd/u5cht` 開機直接讀檔進場**,
// 於是玩家永遠走不到那張選單。這條測試釘的是**入口**,不是流程:
// 「功能在、入口不在」不會讓任何既有測試變紅。
func TestMenuSecondItemStartsCreation(t *testing.T) {
	s := newCreateState(t)
	s.BeginMainMenu()
	if !s.InMainMenu() {
		t.Fatal("主選單沒開起來")
	}
	s.MenuMove(1)
	if s.Menu.Cursor != MenuCreateCharacter {
		t.Fatalf("往下一格是 %v,預期建立新角色", s.Menu.Cursor)
	}
	if !s.MenuChoose() {
		t.Fatal("選了建立新角色,選單沒關")
	}
	if s.Prompt != PromptCreate {
		t.Fatalf("選完之後 Prompt 是 %v,預期 PromptCreate", s.Prompt)
	}
	// 反對照:第一項是「繼續前行」,不該開始建角。
	s2 := newCreateState(t)
	s2.BeginMainMenu()
	if !s2.MenuChoose() || s2.Prompt == PromptCreate {
		t.Error("第一項應該直接進遊戲,不是建角")
	}
}
