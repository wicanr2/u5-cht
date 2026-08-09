package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// pickedMember 同步跑一次 `pickMember`,回傳 (挑到誰, 有沒有開選單)。
//
// ★ 第二個回傳值是**判別力**:這幾條路徑該「不問直接答」,而
// 「開了選單」與「同步回 −1」用單一個 int 分不出來 —— 兩者都會讓
// 呼叫端什麼也沒做,但原因完全不同。
func pickedMember(s *State) (who int, opened bool) {
	who, called := -1, false
	s.pickMember("", func(m int) { who, called = m, true })
	return who, !called
}

// TestOneAbleMemberIsNotAsked —— ★ 只有一個能動就**不問**。
//
// 原版 `sub_E19C` 是 `if (count <= 1) return esi` —— 直接回掃到的那一位。
// 多問一次的話,一人隊伍每個指令都要按兩次鍵。
func TestOneAbleMemberIsNotAsked(t *testing.T) {
	s := activeState(t)
	s.ClearActiveMember()
	for i := 1; i < s.PartySize && i < len(s.Roster); i++ {
		s.Roster[i].Status = u5data.StatusDead
	}
	who, opened := pickedMember(s)
	if opened {
		t.Fatal("只有一個人能動卻開了選單")
	}
	if who != 0 {
		t.Errorf("挑到第 %d 位,預期唯一能動的第 0 位", who)
	}
}

// TestNobodyAbleAnswersMinusOne —— 全隊倒下:同步回 −1,不開選單。
func TestNobodyAbleAnswersMinusOne(t *testing.T) {
	s := activeState(t)
	s.ClearActiveMember()
	for i := 0; i < len(s.Roster); i++ {
		s.Roster[i].Status = u5data.StatusDead
	}
	who, opened := pickedMember(s)
	if opened {
		t.Fatal("沒人能動卻開了選單")
	}
	if who != -1 {
		t.Errorf("回 %d,預期 −1", who)
	}
}

// TestTwoAbleMembersOpenTheMenu —— ★ 2 人以上要**問**。
//
// 這是 `sub_2A7F4` 存在的理由,也是此前引擎最大的落差:舊版直接取
// 最後一位能動的隊員,玩家完全沒有選擇權。
func TestTwoAbleMembersOpenTheMenu(t *testing.T) {
	s := activeState(t)
	s.ClearActiveMember()
	_, opened := pickedMember(s)
	if !opened {
		t.Fatal("兩人以上能動卻沒開選單")
	}
	if s.Prompt != PromptPick || s.Pick == nil {
		t.Fatalf("Prompt=%v Pick=%v,預期選單開著", s.Prompt, s.Pick)
	}
	if s.Pick.Title != MsgSelect {
		t.Errorf("標題是 %q,預期 %q", s.Pick.Title, MsgSelect)
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgPlayerPrompt) {
		t.Errorf("沒印「%s」:%q", MsgPlayerPrompt, s.Messages)
	}
}

// TestTheMenuListsEveryoneIncludingTheDead —— ★★ 游標可以停在死人身上。
//
// 原版**不把死人從清單裡藏起來**:你選得到他,按下去才說「無法行事!」。
// 藏起來的話玩家不會知道那是狀態問題,只會覺得少了一個人。
func TestTheMenuListsEveryoneIncludingTheDead(t *testing.T) {
	s := activeState(t)
	s.ClearActiveMember()
	if s.PartySize < 3 {
		t.Skip("隊伍太小")
	}
	s.Roster[1].Status = u5data.StatusDead
	if _, opened := pickedMember(s); !opened {
		t.Fatal("沒開選單")
	}
	if n := len(s.Pick.Entries); n != s.PartySize {
		t.Fatalf("清單有 %d 項,預期整隊 %d 人", n, s.PartySize)
	}
	if got := s.Pick.Entries[1].Value; got != 1 {
		t.Errorf("第 2 項的值是 %d,預期名冊索引 1", got)
	}
}

// TestDisabledMemberReAsks —— ★★ 選到不能動的人 → 印一句,然後**回到選單**。
//
// 原版是 `aDisabled` + `jmp loc_E20E`(繞回選單),不是取消。
// 做成取消的話玩家按錯一次就得把整個指令重打。
func TestDisabledMemberReAsks(t *testing.T) {
	s := activeState(t)
	s.ClearActiveMember()
	if s.PartySize < 3 {
		t.Skip("隊伍太小")
	}
	s.Roster[1].Status = u5data.StatusAsleep
	got, called := -1, 0
	s.pickMember("", func(m int) { got, called = m, called+1 })
	if s.Pick == nil {
		t.Fatal("沒開選單")
	}
	// 把游標移到睡著的那一位再按 Enter。
	if !s.PickMemberDigit('2') {
		t.Fatal("數字鍵 2 沒被選單吃掉")
	}
	if s.Pick.Cursor != 1 {
		t.Fatalf("游標在第 %d 項,預期第 1 項", s.Pick.Cursor)
	}
	s.PickChoose()
	if called != 0 {
		t.Errorf("選到睡著的人卻已經回呼(值 %d)", got)
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgDisabled) {
		t.Errorf("沒印「%s」:%q", MsgDisabled, s.Messages)
	}
	// ★ 關鍵:選單要**還開著**。
	if s.Pick == nil || s.Prompt != PromptPick {
		t.Fatalf("選到不能動的人之後選單關了(Pick=%v Prompt=%v)", s.Pick, s.Prompt)
	}
	// 再選一個能動的 —— 這次要成功。
	s.PickMemberDigit('1')
	s.PickChoose()
	if called != 1 || got != 0 {
		t.Errorf("重選之後 called=%d got=%d,預期 1 / 0", called, got)
	}
}

// TestCancelStillCallsThen —— ★ 取消也要回呼一次(值 −1)。
//
// 少了這一條,玩家按 ESC 之後呼叫端的流程會**靜靜停住** —— 比報錯難查,
// 因為畫面上什麼異常都沒有(`docs/re/98`)。
func TestCancelStillCallsThen(t *testing.T) {
	s := activeState(t)
	s.ClearActiveMember()
	got, called := -2, 0
	s.pickMember("", func(m int) { got, called = m, called+1 })
	if s.Pick == nil {
		t.Fatal("沒開選單")
	}
	s.PickCancel()
	if called != 1 {
		t.Fatalf("取消之後回呼 %d 次,預期 1 次", called)
	}
	if got != -1 {
		t.Errorf("取消時回呼收到 %d,預期 −1", got)
	}
}

// TestDigitOutsideThePartyIsSwallowed —— 超出人數的數字鍵:不動游標,但算吃掉。
//
// 原版 `sub_2A7F4` 對 `'1'..'6'` 一律進那個分支,只有 `< 隊伍人數` 才採用;
// 沒採用時**不會**往下傳給別的指令。
func TestDigitOutsideThePartyIsSwallowed(t *testing.T) {
	s := activeState(t)
	s.ClearActiveMember()
	if _, opened := pickedMember(s); !opened {
		t.Fatal("沒開選單")
	}
	before := s.Pick.Cursor
	if !s.PickMemberDigit('6') && s.PartySize < 6 {
		t.Error("超出人數的數字鍵沒被吃掉")
	}
	if s.PartySize < 6 && s.Pick.Cursor != before {
		t.Errorf("超出人數卻移動了游標到第 %d 項", s.Pick.Cursor)
	}
	// 非數字、以及 '7' 以上不歸它管。
	//
	// ⚠ **'7' 這一條是刻意與 `SetActivePlayer` 不同的**:那一支收 '0'..'9'
	// 十個鍵(超出人數印「無效!」),而選單只收 '0'..'6'。兩支的鍵範圍
	// 在原版就不一樣,別「順手統一」。
	if s.PickMemberDigit('7') {
		t.Error("'7' 被當成隊員選擇")
	}
	if s.PickMemberDigit('a') {
		t.Error("字母被當成隊員選擇")
	}
}

// TestZeroAndSpaceLeaveTheMenu —— ★ '0' 與空白鍵都是「不選任何人」。
//
// 原版 `sub_2A7F4`:`'0'` → `sel = −1`、`0x20`(空白)與 ESC 同一個出口,
// 而 `sel < 0` 印的是「None!」。⇒ 三個鍵都要回呼 −1,而且印的是
// 原版那句而不是通用的「作罷。」(一個動作印兩句話會看起來像 bug)。
func TestZeroAndSpaceLeaveTheMenu(t *testing.T) {
	for _, key := range []rune{PickMemberKeyNone, PickMemberKeyQuit} {
		s := activeState(t)
		s.ClearActiveMember()
		got, called := -2, 0
		s.pickMember("", func(m int) { got, called = m, called+1 })
		if s.Pick == nil {
			t.Fatalf("'%c':沒開選單", key)
		}
		s.Messages = nil
		if !s.PickMemberDigit(key) {
			t.Fatalf("'%c' 沒被選單吃掉", key)
		}
		if called != 1 || got != -1 {
			t.Errorf("'%c':called=%d got=%d,預期 1 / −1", key, called, got)
		}
		if s.Pick != nil {
			t.Errorf("'%c' 之後選單還開著", key)
		}
		joined := strings.Join(s.Messages, "|")
		if !strings.Contains(joined, MsgActiveNone) {
			t.Errorf("'%c':沒印「%s」:%q", key, MsgActiveNone, s.Messages)
		}
		if strings.Contains(joined, MsgNevermind) {
			t.Errorf("'%c':印了通用的「%s」,原版只印「%s」:%q",
				key, MsgNevermind, MsgActiveNone, s.Messages)
		}
	}
}

// TestDigitOnlyWorksInTheMemberMenu —— ★ 數字鍵直跳只在**選人**選單裡有效。
//
// 別的選單(裝備、藥草)也是 `Picker`,而它們的數字鍵不該跳游標 ——
// 那會讓「輸入數量」那類流程收不到數字。
func TestDigitOnlyWorksInTheMemberMenu(t *testing.T) {
	s := activeState(t)
	if !s.beginPick("哪一件?", []PickEntry{{Label: "甲", Value: 0}, {Label: "乙", Value: 1}},
		MsgNobodyHere, func(int) bool { return true }) {
		t.Fatal("開不了測試用選單")
	}
	if s.PickMemberDigit('2') {
		t.Error("非選人選單也吃掉了數字鍵")
	}
	if s.Pick.Cursor != 0 {
		t.Errorf("非選人選單的游標被移到第 %d 項", s.Pick.Cursor)
	}
}

// TestPoisonedCanStillAct —— ★ 中毒**算能動**。
//
// 原版判的是 'G' 或 'P' 兩種。寫成「只有 G」會讓中毒的隊員突然不能做事,
// 而中毒在 U5 裡很常見 —— 這個 bug 會一直被當成「狀態顯示怪怪的」。
func TestPoisonedCanStillAct(t *testing.T) {
	s := activeState(t)
	s.ClearActiveMember()
	for i := 0; i < len(s.Roster); i++ {
		s.Roster[i].Status = u5data.StatusDead
	}
	s.Roster[1].Status = u5data.StatusPoisoned
	who, opened := pickedMember(s)
	if opened {
		t.Fatal("只有一個(中毒的)能動卻開了選單")
	}
	if who != 1 {
		t.Errorf("挑到 %d,預期中毒但能動的第 1 位", who)
	}
}

// actAs 讓接下來的指令固定由第 who 位執行。
//
// ★★ 這**不是**測試專用的後門,而是原版自己的機制:數字鍵「指定行動者」
// (`sub_2BD40` → `byte_3E08B`)。差別很重要 —— 後門會遮住選單那條路的 bug
// (`CLAUDE.md §6.1`:debug hook 會讓回歸測試全綠而玩家一開就壞),
// 而這一支走的是玩家真的按得到的鍵。
//
// ⚠ 用它的測試在驗「效果對不對」;「該不該問」那件事由 pickmember_test.go
// 本身的那幾條驗。兩者不要混在一個測試裡。
func actAs(t *testing.T, s *State, who int) {
	t.Helper()
	if who < 0 || who >= len(s.Roster) {
		t.Fatalf("指定第 %d 位,但名冊只有 %d 人", who, len(s.Roster))
	}
	s.Roster[who].Status = u5data.StatusGood
	if !s.SetActivePlayer(rune(ActiveKeyFirst + who)) {
		t.Fatalf("'%c' 不被當成指定行動者的鍵", rune(ActiveKeyFirst+who))
	}
	if got := s.ActiveMember(); got != who {
		t.Fatalf("指定第 %d 位之後 ActiveMember 回 %d", who, got)
	}
}
