package game

import (
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// TestTalkFlowWithRealData 用原版資料跑一整段對話:
// 走到 NPC 旁邊 → T → 問 job → 問一個他認得的關鍵字 → bye 結束。
func TestTalkFlowWithRealData(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	scenes, err := u5data.LoadSceneSet(dir)
	if err != nil {
		t.Fatal(err)
	}
	npcs, err := u5data.LoadNPCSet(dir)
	if err != nil {
		t.Fatal(err)
	}
	talks, err := u5data.LoadTalkSet(dir)
	if err != nil {
		t.Fatal(err)
	}

	s := &State{Scenes: scenes, NPCs: npcs, Talks: talks, Clock: NewClock(), MaxMessages: 32}
	if err := s.SetScene(britain, 0, 0, 0); err != nil {
		t.Fatal(err)
	}
	// 找一個此刻在這一層、且有正常對話的 NPC,站到他旁邊。
	var target *VisibleNPC
	for _, v := range s.VisibleNPCs() {
		if v.NPC.Dialogue > 0 && v.NPC.Dialogue < u5data.DialogueShopFirst {
			vv := v
			target = &vv
			break
		}
	}
	if target == nil {
		t.Skip("不列顛城此刻沒有可交談的 NPC")
	}
	s.X, s.Y = target.X, target.Y-1 // 站在他北邊,他就在南邊

	s.Talk()
	if s.Prompt != PromptTalk || s.Conv == nil {
		t.Fatalf("按 T 沒有進入對話:%v", s.Messages)
	}
	name := s.Conv.Name

	// 問 job —— 這是引擎內建的關鍵字,不在記錄的關鍵字表裡。
	typeWord(s, KeywordJob)
	if !anyContains(s.Messages, firstWords(s.Conv.Job)) {
		t.Errorf("問 job 沒有得到職業回答(%q):%v", s.Conv.Job, s.Messages)
	}

	// 問一個他自己列出的關鍵字。
	if kws := s.Conv.Keywords(); len(kws) > 0 {
		want, _, _ := s.Conv.Respond(kws[0])
		typeWord(s, kws[0])
		if want != "" && !anyContains(s.Messages, firstWords(want)) {
			t.Errorf("問 %q 沒有得到對應回答(%q):%v", kws[0], want, s.Messages)
		}
	}

	// 打不存在的字 → 聽不懂
	typeWord(s, "zzzz")
	if last := s.Messages[len(s.Messages)-1]; last != MsgDoesNotUnderstand {
		t.Errorf("問不存在的關鍵字得到 %q,預期 %q", last, MsgDoesNotUnderstand)
	}

	// bye → 結束
	typeWord(s, KeywordBye)
	if s.Prompt != PromptNone || s.Conv != nil {
		t.Errorf("打 bye 之後對話沒結束(對象 %s)", name)
	}
}

func TestTypeRuneOnlyAcceptsLetters(t *testing.T) {
	s := &State{Prompt: PromptTalk}
	for _, r := range []rune{'a', 'B', '1', '-', ' ', '中', 'z'} {
		s.TypeRune(r)
	}
	if s.Input != "abz" {
		t.Errorf("輸入列是 %q,預期 %q —— 關鍵字只收英文字母且自動轉小寫", s.Input, "abz")
	}
	s.Backspace()
	if s.Input != "ab" {
		t.Errorf("退格後是 %q", s.Input)
	}
}

// TestTypeRuneIgnoredOutsideTalk:不在對話中時打字不該累積,
// 否則玩家按 E/K/T 這些指令鍵會被偷偷記進輸入列。
func TestTypeRuneIgnoredOutsideTalk(t *testing.T) {
	s := &State{}
	s.TypeRune('a')
	if s.Input != "" {
		t.Errorf("非對話狀態卻收了字:%q", s.Input)
	}
}

func typeWord(s *State, w string) {
	for _, r := range w {
		s.TypeRune(r)
	}
	s.Submit()
}

func anyContains(msgs []string, sub string) bool {
	if sub == "" {
		return true
	}
	for _, m := range msgs {
		if strings.Contains(m, sub) {
			return true
		}
	}
	return false
}

// firstWords 取回應的前幾個字當比對錨點 —— 訊息欄會把換行壓成空白,
// 整句比對會因為空白差異而假失敗。
func firstWords(s string) string {
	f := strings.Fields(strings.ReplaceAll(s, "\n", " "))
	if len(f) > 3 {
		f = f[:3]
	}
	return strings.Join(f, " ")
}
