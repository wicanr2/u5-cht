package game

import (
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
// 「從創世紀 IV 轉入」做一個什麼都不做卻關掉選單的分支,玩家會以為
// 角色轉進來了 —— 那比缺這一項更糟(CLAUDE.md §3.0)。
func TestUnimplementedMenuItemsSaySo(t *testing.T) {
	for _, c := range []struct {
		item MenuItem
		want string
	}{
		{MenuTransferU4, MsgTransferNotImplemented},
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
