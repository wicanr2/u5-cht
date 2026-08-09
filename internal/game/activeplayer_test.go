package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

func activeState(t *testing.T) *State {
	t.Helper()
	s := lockScene(t)
	if s.PartySize < 2 {
		t.Skip("隊伍太小")
	}
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		s.Roster[i].Status = u5data.StatusGood
	}
	return s
}

// TestFreshStateHasNoActiveMember —— ★ 結構常值的零值不能被當成「指定了第一位」。
//
// ⚠ 這條就是 `activeSet` 旗標存在的理由。少了它,任何 `&State{…}` 一開始
// 就等於「指定了隊員 0」,而那會讓喝噴泉、看水晶球全部落在同一個人頭上。
func TestFreshStateHasNoActiveMember(t *testing.T) {
	var s State
	if got := s.ActiveMember(); got != -1 {
		t.Errorf("全新的 State 指定了第 %d 位,預期 -1", got)
	}
}

// TestSetActivePlayerAcceptsTenDigits —— ★ 收 '0'..'9' 十個鍵。
//
// ⚠ 不是 '0'..'6'。'7'..'9' 會算出超出人數的索引 → 印「無效!」,
// 但**仍然算它管的鍵**(回 true)。少收三個鍵的話玩家按 8 會落到別的指令。
func TestSetActivePlayerAcceptsTenDigits(t *testing.T) {
	s := activeState(t)
	for r := '0'; r <= '9'; r++ {
		if !s.SetActivePlayer(r) {
			t.Errorf("按 %q 沒被當成指定行動者的鍵", r)
		}
	}
	// 非數字鍵不歸它管。
	for _, r := range []rune{'a', 'Z', ' ', '/'} {
		if s.SetActivePlayer(r) {
			t.Errorf("按 %q 竟然被當成指定行動者的鍵", r)
		}
	}
}

// TestSetActivePlayerPicksAndClears —— 指定與取消。
func TestSetActivePlayerPicksAndClears(t *testing.T) {
	s := activeState(t)
	s.Messages = nil
	s.SetActivePlayer('2')
	if got := s.ActiveMember(); got != 1 {
		t.Errorf("按 2 指定了第 %d 位,預期第 1 位(0-based)", got)
	}
	if got := strings.Join(s.Messages, "|"); !strings.Contains(got, s.Roster[1].Name) {
		t.Errorf("沒印出被指定的人名:%q", got)
	}
	s.Messages = nil
	s.SetActivePlayer('0')
	if got := s.ActiveMember(); got != -1 {
		t.Errorf("按 0 之後還指定著第 %d 位", got)
	}
	if got := strings.Join(s.Messages, "|"); !strings.Contains(got, MsgActiveNone) {
		t.Errorf("按 0 沒印「%s」:%q", MsgActiveNone, got)
	}
}

// TestDeadOrAsleepCannotBeActive —— ★ 死了或睡著了不能被指定,中毒可以。
func TestDeadOrAsleepCannotBeActive(t *testing.T) {
	cases := []struct {
		status byte
		ok     bool
		name   string
	}{
		{u5data.StatusGood, true, "良好"},
		{u5data.StatusPoisoned, true, "中毒"},
		{u5data.StatusDead, false, "死亡"},
		{u5data.StatusAsleep, false, "睡著"},
	}
	for _, c := range cases {
		s := activeState(t)
		s.ClearActiveMember()
		s.Roster[1].Status = c.status
		s.Messages = nil
		s.SetActivePlayer('2')
		got := s.ActiveMember() == 1
		if got != c.ok {
			t.Errorf("狀態「%s」被指定 = %v,預期 %v(訊息 %q)",
				c.name, got, c.ok, s.Messages)
		}
		if !c.ok && !strings.Contains(strings.Join(s.Messages, "|"), MsgActiveInvalid) {
			t.Errorf("狀態「%s」該印「%s」:%q", c.name, MsgActiveInvalid, s.Messages)
		}
	}
}

// TestOutOfPartyIsInvalid —— 超出隊伍人數印「無效!」。
func TestOutOfPartyIsInvalid(t *testing.T) {
	s := activeState(t)
	s.ClearActiveMember()
	s.Messages = nil
	s.SetActivePlayer('9') // 第 8 位,任何隊伍都沒這麼多人
	if got := s.ActiveMember(); got != -1 {
		t.Errorf("按 9 指定了第 %d 位", got)
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgActiveInvalid) {
		t.Errorf("沒印「%s」:%q", MsgActiveInvalid, s.Messages)
	}
}

// TestPickCharacterHonoursTheActiveMember —— ★ 指定過就不再自動掃。
//
// 這是接上 `sub_E19C` 第一條分支的驗收:自動掃回**最後一位**能動的,
// 而指定之後要回**被指定的那一位**。兩者不同才驗得出來。
func TestPickCharacterHonoursTheActiveMember(t *testing.T) {
	s := activeState(t)
	s.ClearActiveMember()
	auto, opened := pickedMember(s)
	if opened {
		// 沒指定 + 多人能動 → 原版會開選單。取最後一位當對照組。
		s.PickCancel()
		auto = -1
		for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
			switch s.Roster[i].Status {
			case u5data.StatusGood, u5data.StatusPoisoned:
				auto = i
			}
		}
	}
	if auto < 0 {
		t.Skip("沒人可選")
	}
	// 挑一個與自動結果不同的隊員。
	want := 0
	if auto == 0 {
		want = 1
	}
	s.SetActivePlayer(rune('1' + want))
	got, opened2 := pickedMember(s)
	if opened2 {
		t.Fatal("指定過了卻還開選單 —— `sub_E19C` 第一條分支沒生效")
	}
	if got != want {
		t.Errorf("指定第 %d 位之後挑到 %d(自動會挑 %d)", want, got, auto)
	}
}

// TestActiveMemberIsNotRevalidated —— ★ 原版不重驗狀態,照抄。
//
// 指定之後那個人死了,指令**照樣**落在他頭上。這看起來像 bug,
// 但原版 `sub_E19C` 只檢查 `byte_3E08B != 0xFF`(`docs/re/97`)。
func TestActiveMemberIsNotRevalidated(t *testing.T) {
	s := activeState(t)
	s.ClearActiveMember()
	s.SetActivePlayer('2')
	if s.ActiveMember() != 1 {
		t.Skip("指定不成功")
	}
	s.Roster[1].Status = u5data.StatusDead
	got, opened := pickedMember(s)
	if opened {
		t.Fatal("被指定的人死了就改成問玩家 —— 原版不重驗狀態")
	}
	if got != 1 {
		t.Errorf("被指定的人死了之後挑到 %d,原版仍回 1(不重驗)", got)
	}
}
