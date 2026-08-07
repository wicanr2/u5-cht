package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

func endState(t *testing.T) *State {
	t.Helper()
	s := shrineState(t)
	if s == nil {
		return nil
	}
	tf, err := u5data.LoadText(gameDataDir(t) + "/ENDMSG.DAT")
	if err != nil {
		t.Fatal(err)
	}
	s.EndMsg = tf
	return s
}

// TestEndingMessageOffsetsMatchTheOriginal:十一筆結局訊息的序號要對得上。
//
// ★ `sub_135FC` 是拿**寫死的位移**取字(`byte_54700 + n`,`ENDMSG.DAT`
// 從檔頭載入)。這條把那十個位移逐一對回記錄開頭 —— 算錯的話結局會唸錯句子,
// 而且不會有任何東西報錯。
func TestEndingMessageOffsetsMatchTheOriginal(t *testing.T) {
	s := endState(t)
	if s == nil {
		return
	}
	want := []struct {
		offset, record int
	}{
		{0x000, u5data.MsgEndWellMet},
		{0x021, u5data.MsgEndAskBox},
		{0x049, u5data.MsgEndAskBox2},
		{0x0AB, u5data.MsgEndOpensBox},
		{0x0D3, u5data.MsgEndArtifact},
		{0x128, u5data.MsgEndOurWorld},
		{0x167, u5data.MsgEndFreeUs},
		{0x1C9, u5data.MsgEndOlder},
		{0x211, u5data.MsgEndOrb},
		{0x24B, u5data.MsgEndFollow},
		{0x2D5, u5data.MsgEndPullUpAChair},
	}
	for _, w := range want {
		if w.record >= len(s.EndMsg.Records) {
			t.Fatalf("記錄 %d 超出 %d 筆", w.record, len(s.EndMsg.Records))
		}
		if got := s.EndMsg.Records[w.record].Offset; got != w.offset {
			t.Errorf("記錄 %d 的位移是 0x%X,原版指的是 0x%X", w.record, got, w.offset)
		}
	}
}

// TestEndingRevivesTheDeadUnconditionally:結局一開始就把死掉的隊員救回來。
//
// ⚠ 兩件事都要做:狀態改回 'G',**而且**目前 HP 補成最大 HP。
// 只改狀態的話人活過來但血是 0。
// ⚠ 而且是**無條件**的 —— 在問「盒子帶了沒」之前就做完了。
func TestEndingRevivesTheDeadUnconditionally(t *testing.T) {
	s := endState(t)
	if s == nil {
		return
	}
	if s.PartySize < 2 {
		t.Skip("隊伍太少")
	}
	s.Roster[1].Status = u5data.StatusDead
	s.Roster[1].HP = 0
	s.SandalwoodBox = false // 就算沒帶盒子也要復活
	if !s.BeginEnding() {
		t.Fatalf("結局沒有開始:\n%s", s.log())
	}
	if s.Roster[1].Status != u5data.StatusGood {
		t.Errorf("狀態還是 %c", s.Roster[1].Status)
	}
	if s.Roster[1].HP != s.Roster[1].MaxHP {
		t.Errorf("HP 是 %d,預期補滿到 %d", s.Roster[1].HP, s.Roster[1].MaxHP)
	}
	if !strings.Contains(s.log(), "活過來了") {
		t.Errorf("沒有印復活訊息:\n%s", s.log())
	}
}

// TestSayingYesWithoutTheBoxIsTheBadEnding:嘴上說有不算,得真的帶著盒子。
//
// ⚠ 原版是 `and`:回答要是 Y,**而且** `byte_3DFCD != 0`。
// 少了後者,那只盒子(整條主線最後一件收集品)就變成可有可無。
func TestSayingYesWithoutTheBoxIsTheBadEnding(t *testing.T) {
	s := endState(t)
	if s == nil {
		return
	}
	s.SandalwoodBox = false
	s.BeginEnding()
	s.AnswerEnding(true)
	if s.Ending == nil {
		t.Fatal("回答之後結局就結束了")
	}
	if s.Ending.Good {
		t.Error("沒帶盒子卻走了真結局")
	}
	for s.AdvanceEnding() {
	}
	if !strings.Contains(s.log(), s.endText(u5data.MsgEndPullUpAChair)) {
		t.Errorf("沒有印「搬張椅子坐下吧」:\n%s", s.log())
	}
	if strings.Contains(s.log(), s.endText(u5data.MsgEndOrb)) {
		t.Error("壞結局卻唸出了月之球那一段")
	}
}

// TestSayingYesWithTheBoxIsTheTrueEnding:帶著盒子答 Y 才看得到真結局。
func TestSayingYesWithTheBoxIsTheTrueEnding(t *testing.T) {
	s := endState(t)
	if s == nil {
		return
	}
	s.SandalwoodBox = true
	s.BeginEnding()
	s.AnswerEnding(true)
	if s.Ending == nil || !s.Ending.Good {
		t.Fatalf("帶著盒子答 Y 卻不是真結局:\n%s", s.log())
	}
	for s.AdvanceEnding() {
	}
	for _, r := range u5data.EndingFinale {
		if !strings.Contains(s.log(), s.endText(r)) {
			t.Errorf("真結局漏了記錄 %d:\n%s", r, s.log())
		}
	}
}

// TestAnsweringNoAsksOneMoreTime:答 N 會換個說法再問一次,而且只再問一次。
//
// 原版的第二問是另一段文字(把盒子講清楚:「吾寢室密道裡的那只檀香木盒」)。
// 少了它,玩家隨手按到 N 就直接吃到壞結局。
func TestAnsweringNoAsksOneMoreTime(t *testing.T) {
	s := endState(t)
	if s == nil {
		return
	}
	s.SandalwoodBox = true
	s.BeginEnding()
	s.AnswerEnding(false)
	if s.Ending == nil || !s.Ending.Asking {
		t.Fatalf("答 N 之後沒有再問一次:\n%s", s.log())
	}
	if !strings.Contains(s.log(), s.endText(u5data.MsgEndAskBox2)) {
		t.Errorf("第二問用的不是另一段文字:\n%s", s.log())
	}
	// 第二次還是 N → 壞結局,不再追問。
	s.AnswerEnding(false)
	if s.Ending.Asking {
		t.Error("問了第三次")
	}
	if s.Ending.Good {
		t.Error("兩次都答 N 卻走了真結局")
	}
}

// TestSandalwoodBoxSurvivesSaveRoundTrip:盒子存得回去也讀得回來。
func TestSandalwoodBoxSurvivesSaveRoundTrip(t *testing.T) {
	s := endState(t)
	if s == nil || s.BaseSave == nil {
		return
	}
	s.SandalwoodBox = true
	items := s.Inventory.Items
	sv, err := s.ExportSave(s.BaseSave)
	if err != nil {
		t.Fatal(err)
	}
	blob, err := sv.Encode()
	if err != nil {
		t.Fatal(err)
	}
	back, err := u5data.ParseSave(blob)
	if err != nil {
		t.Fatal(err)
	}
	if back.SandalwoodBox == 0 {
		t.Error("盒子讀回來不見了")
	}
	if back.Inventory.Items != items {
		// 0x0219 就緊貼在 0x021A 的裝備表前面 —— 偏一格最先被踩到的就是它。
		t.Error("裝備表被踩到了 —— 0x0219 這個位移大概算錯了")
	}
}
