package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// wellScene 造一個站在井旁邊的隊伍。
//
// 井本身不必真的存在於地圖上 —— 這幾條測的是 `sub_CD28` 的流程,
// 而 Look 走到那裡的路徑另有測試。
func wellScene(t *testing.T, location int) *State {
	t.Helper()
	s := newCreateState(t)
	s.MaxMessages = 32
	s.Location = location
	s.X, s.Y, s.Floor = 10, 10, 0
	s.Inventory.Gold = 100
	s.Objects = &u5data.ObjectSet{}
	s.sceneObjects = &u5data.ObjectSet{}
	s.Messages = nil
	return s
}

// horsesNear 數一下 (X+1, Y) 有幾匹馬。
func horsesNear(s *State) int {
	objs := s.currentObjects()
	if objs == nil {
		return 0
	}
	n := 0
	for i := range objs.Objects {
		o := &objs.Objects[i]
		if o.Present() && o.Kind == u5data.TileHorse && o.X == s.X+1 && o.Y == s.Y {
			n++
		}
	}
	return n
}

// TestWellIsInteractiveNotJustOneLine 是這一條的核心。
//
// ⚠ 引擎原本只印「一口深井。」就結束,因為當初讀的是 Hex-Rays 截斷後的
// `sub_CD28`。這條擋的是回歸到那個狀態。
func TestWellIsInteractiveNotJustOneLine(t *testing.T) {
	s := wellScene(t, u5data.WellLocationPaws)
	s.lookAtWell()
	if s.Prompt != PromptYesNo {
		t.Fatalf("看了井之後 Prompt=%v,預期在問 Y/N", s.Prompt)
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgDropACoin) {
		t.Errorf("沒問「要投一枚錢幣嗎」:%q", s.Messages)
	}
}

// TestWellFullFlowSpawnsAHorse 走完整條路:投錢 → 許願 → 馬。
func TestWellFullFlowSpawnsAHorse(t *testing.T) {
	for _, loc := range []int{u5data.WellLocationPaws, u5data.WellLocationEmpathAbbey} {
		s := wellScene(t, loc)
		s.lookAtWell()
		s.AnswerYesNo(true)
		if s.Prompt != PromptText {
			t.Fatalf("地點 %d:答 Y 之後 Prompt=%v,預期在收願望", loc, s.Prompt)
		}
		if s.Inventory.Gold != 99 {
			t.Errorf("地點 %d:金錢 %d,預期先扣一枚變 99", loc, s.Inventory.Gold)
		}
		for _, r := range "FERRARI" {
			s.TypeText(r)
		}
		s.SubmitText()
		if got := horsesNear(s); got != 1 {
			t.Errorf("地點 %d:旁邊有 %d 匹馬,預期 1 匹(%q)", loc, got, s.Messages)
		}
		if !strings.Contains(strings.Join(s.Messages, "|"), MsgPoof) {
			t.Errorf("地點 %d:沒印「砰!」:%q", loc, s.Messages)
		}
	}
}

// TestWellOnlyWorksAtPawsAndEmpathAbbey 釘住那兩個地點編號。
func TestWellOnlyWorksAtPawsAndEmpathAbbey(t *testing.T) {
	// 21 與 23 夾住 22(PAWS),30 與 32 夾住 31(EMPATH ABBEY)——
	// 相鄰地點也不行,這樣才擋得住「地點條件寫成範圍」的錯。
	for _, loc := range []int{0, 1, 21, 23, 30, 32} {
		s := wellScene(t, loc)
		s.lookAtWell()
		s.AnswerYesNo(true)
		for _, r := range "HORSE" {
			s.TypeText(r)
		}
		s.SubmitText()
		if got := horsesNear(s); got != 0 {
			t.Errorf("地點 %d 竟然生出了 %d 匹馬", loc, got)
		}
		if !strings.Contains(strings.Join(s.Messages, "|"), MsgNoEffect) {
			t.Errorf("地點 %d:沒印「毫無效果」:%q", loc, s.Messages)
		}
		// ⚠ 但錢**照扣** —— 原版扣錢在判定之前。
		if s.Inventory.Gold != 99 {
			t.Errorf("地點 %d:金錢 %d,預期照扣變 99", loc, s.Inventory.Gold)
		}
	}
}

// TestAllSixWishesWork 六個字面值一個都不能漏。
func TestAllSixWishesWork(t *testing.T) {
	if len(u5data.WellWishes) != 6 {
		t.Fatalf("願望表有 %d 筆,預期 6 筆", len(u5data.WellWishes))
	}
	for _, wish := range u5data.WellWishes {
		s := wellScene(t, u5data.WellLocationPaws)
		s.lookAtWell()
		s.AnswerYesNo(true)
		for _, r := range wish {
			s.TypeText(r)
		}
		s.SubmitText()
		if got := horsesNear(s); got != 1 {
			t.Errorf("願望 %q 生出 %d 匹馬,預期 1 匹", wish, got)
		}
	}
}

// TestWishMatchingIsAPrefixOfTheUppercaseLiteral 釘住比對規則。
//
// 原版 `sub_27C98` 把**字面值**轉大寫後 `strncmp` 字面值的長度 ——
// 所以輸入比字面值長算相符,短的不算,小寫的不算。
func TestWishMatchingIsAPrefixOfTheUppercaseLiteral(t *testing.T) {
	cases := []struct {
		wish  string
		match bool
		why   string
	}{
		{"HORSE", true, "剛好相等"},
		{"HORSEY", true, "只比字面值的 5 個位元組 → 相符"},
		{"HORS", false, "第 5 個位元組是 NUL vs 'E'"},
		{"horse", false, "★ 原版比的是大寫字面值,小寫不生效"},
		{"LAMBORGHINI", true, "最長的那個"},
		{"", false, "空的走 Nothing 那條"},
		{"MUSTANG", false, "不在表裡"},
		{" HORSE", false, "★ 比對層不做 trim —— 前面多一個空白就對不上"},
	}
	for _, c := range cases {
		if got := u5data.WellWishMatches(c.wish) >= 0; got != c.match {
			t.Errorf("%q 相符 = %v,預期 %v(%s)", c.wish, got, c.match, c.why)
		}
	}
}

// TestNoCoinNoWishAndNoMessage 是最容易寫錯的一段:沒錢時原版**一句話都不印**。
func TestNoCoinNoWishAndNoMessage(t *testing.T) {
	s := wellScene(t, u5data.WellLocationPaws)
	s.Inventory.Gold = 0
	s.lookAtWell()
	s.AnswerYesNo(true)
	if s.Prompt == PromptText {
		t.Error("沒錢卻還問願望")
	}
	joined := strings.Join(s.Messages, "|")
	if strings.Contains(joined, MsgThyWish) {
		t.Errorf("沒錢卻印了「汝所願為何」:%q", s.Messages)
	}
	// 最後一句應該就是「是。」—— 之後什麼都沒有(貼合原版的靜默)。
	if len(s.Messages) == 0 || !strings.Contains(s.Messages[len(s.Messages)-1], MsgYes) {
		t.Errorf("最後一句是 %q,預期停在「是。」", s.Messages)
	}
}

// TestSayingNoCostsNothing:答 N 不扣錢、不問願望。
func TestSayingNoCostsNothing(t *testing.T) {
	s := wellScene(t, u5data.WellLocationPaws)
	s.lookAtWell()
	s.AnswerYesNo(false)
	if s.Inventory.Gold != 100 {
		t.Errorf("答 N 卻扣了錢,剩 %d", s.Inventory.Gold)
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgNo) {
		t.Errorf("沒印「否」:%q", s.Messages)
	}
	if horsesNear(s) != 0 {
		t.Error("答 N 卻生出了馬")
	}
}

// TestEmptyWishSaysNothing:什麼都不打 → 「無。」,而錢已經扣掉了。
func TestEmptyWishSaysNothing(t *testing.T) {
	s := wellScene(t, u5data.WellLocationPaws)
	s.lookAtWell()
	s.AnswerYesNo(true)
	s.SubmitText() // 直接送出空字串
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgNothing) {
		t.Errorf("沒印「無。」:%q", s.Messages)
	}
	if s.Inventory.Gold != 99 {
		t.Errorf("金錢 %d,預期已經扣掉一枚", s.Inventory.Gold)
	}
	if horsesNear(s) != 0 {
		t.Error("空願望卻生出了馬")
	}
}

// TestWishLengthCapIsTwelve:原版 `sub_2B770(buf, 0Ch)` 收 12 個字元。
func TestWishLengthCapIsTwelve(t *testing.T) {
	if u5data.WellWishMax != 12 {
		t.Fatalf("上限是 %d,預期 12", u5data.WellWishMax)
	}
	s := wellScene(t, u5data.WellLocationPaws)
	s.lookAtWell()
	s.AnswerYesNo(true)
	for i := 0; i < 20; i++ {
		s.TypeText('X')
	}
	if got := len(s.Input); got != u5data.WellWishMax {
		t.Errorf("打了 20 個字收下 %d 個,預期 %d 個", got, u5data.WellWishMax)
	}
	// ⚠ 最長的願望 LAMBORGHINI 是 11 個字元 —— 剛好放得進 12。
	longest := 0
	for _, w := range u5data.WellWishes {
		if len(w) > longest {
			longest = len(w)
		}
	}
	if longest > u5data.WellWishMax {
		t.Errorf("最長的願望 %d 個字元,收不進 %d 的上限", longest, u5data.WellWishMax)
	}
}
