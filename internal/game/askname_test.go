package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// nameScene 造一個正在跟 NPC 對話的狀態。
func nameScene(t *testing.T) *State {
	t.Helper()
	s := &State{Scenes: synthScenes(t, walkable(t)), NPCs: &u5data.NPCSet{}, MaxMessages: 16}
	s.Roster = []u5data.Character{
		{Name: "Elwood", Status: u5data.StatusGood},
		{Name: "Shamino", Status: u5data.StatusGood},
	}
	s.PartySize = 2
	if err := s.SetScene(britain, 0, 15, 15); err != nil {
		t.Fatalf("進不了不列顛城:%v", err)
	}
	s.talkingTo = 4
	s.Conv = &u5data.Conversation{Name: "Zachariah"}
	s.Prompt = PromptTalk
	return s
}

// 報上隊裡的名字 → 對方記住汝。
func TestGivingAPartyNameIsRemembered(t *testing.T) {
	s := nameScene(t)
	s.askName()
	if s.Prompt != PromptAnswer || !s.askingName {
		t.Fatalf("問完名字之後 Prompt=%v askingName=%v", s.Prompt, s.askingName)
	}
	s.Input = "Elwood"
	s.Submit()
	if !strings.Contains(allLogs(s), MsgAPleasure) {
		t.Errorf("報對名字應該回「%s」,實得 %q", MsgAPleasure, allLogs(s))
	}
	if !s.KnowsThyName(4) {
		t.Error("報過名字之後這個 NPC 應該認得汝")
	}
	if s.askingName || s.Prompt != PromptTalk {
		t.Error("回答完應該回到一般對話")
	}
}

// ⚠ 遮罩是**每個地點一份** —— 換一座城,同一個槽不算認得。
func TestBeingKnownIsPerLocation(t *testing.T) {
	s := nameScene(t)
	s.askName()
	s.Input = "Elwood"
	s.Submit()
	if !s.KnowsThyName(4) {
		t.Fatal("這座城的 4 號應該認得汝")
	}
	if err := s.SetScene(3, 0, 15, 15); err != nil { // 換到傑隆
		t.Fatalf("換城失敗:%v", err)
	}
	if s.KnowsThyName(4) {
		t.Error("換了一座城,同樣是 4 號槽不該認得汝")
	}
}

// 空輸入直接被打發掉。
//
// ⚠ 第一次讀 `sub_1C2FC` 時我把 `cmp al, byte_55F38` 讀成「跟對方的名字比」,
// 於是寫了一條「不能拿人家的名字當自己的」規則 —— 那是**我加的,原版沒有**。
// `al` 在那一刻是 0(迴圈變數還沒開始跑),比的是輸入的第一個位元組是不是 0。
func TestEmptyAnswerIsShruggedOff(t *testing.T) {
	s := nameScene(t)
	s.askName()
	s.Input = ""
	s.Submit()
	if !strings.Contains(allLogs(s), MsgIfYouSaySo) {
		t.Errorf("空輸入應該回「%s」,實得 %q", MsgIfYouSaySo, allLogs(s))
	}
	if s.KnowsThyName(4) {
		t.Error("空輸入不該讓他認得汝")
	}
}

// 亂報一個不在隊裡的名字。
func TestAnUnknownNameIsShruggedOff(t *testing.T) {
	s := nameScene(t)
	s.askName()
	s.Input = "Mondain"
	s.Submit()
	if !strings.Contains(allLogs(s), MsgIfYouSaySo) {
		t.Errorf("亂報名字應該回「%s」,實得 %q", MsgIfYouSaySo, allLogs(s))
	}
	if s.KnowsThyName(4) {
		t.Error("亂報名字不該被記住")
	}
}

// ⚠ 比對長度是 **4**,不是對話關鍵字那個 9。
//
// 隊裡有 Elwood → needle 是「Elwo」,所以打「Elwo」算、打「Elw」不算。
// 抄成 9 的話玩家得打完整個名字才行,與原版不同。
func TestNameMatchingTruncatesToFour(t *testing.T) {
	if u5data.NameMatchLen != 4 {
		t.Fatalf("比對長度是 %d,原版是 4", u5data.NameMatchLen)
	}
	if !u5data.NameSpoken("Elwood", "Elwo") {
		t.Error("「Elwo」應該對得上 Elwood(needle 只有前四個字元)")
	}
	if u5data.NameSpoken("Elwood", "Elw") {
		t.Error("「Elw」比 needle 短,對不上")
	}
	if !u5data.NameSpoken("Elwood", "elwood the bard") {
		t.Error("不分大小寫,而且後面多打不影響")
	}
	if u5data.NameSpoken("Elwood", "Shamino") {
		t.Error("不同的名字不該對上")
	}

	// 端對端。
	s := nameScene(t)
	s.askName()
	s.Input = "Elwo"
	s.Submit()
	if !strings.Contains(allLogs(s), MsgAPleasure) {
		t.Errorf("打「Elwo」應該算,實得 %q", allLogs(s))
	}
}

// 0x88 要解析成 AsksName,而且不能被誤判成「同下一則」。
func TestAskNameOpcodeIsNotMistakenForAnAlias(t *testing.T) {
	c := &u5data.Conversation{}
	_, fx := c.Render([]byte{u5data.OpAskName})
	if !fx.AsksName {
		t.Fatal("0x88 應該解析成 AsksName")
	}
	if fx.SameAsNext {
		t.Error("整則只有 0x88 的回應不是「同下一則」—— 它是反問名字")
	}
}
