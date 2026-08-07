package game

import (
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// mixState 造一個身上有藥草的狀態。
func mixState(t *testing.T) *State {
	t.Helper()
	s := castState(t)
	for r := range s.Inventory.Reagents {
		s.Inventory.Reagents[r] = 20
	}
	for i := range s.Inventory.Spells {
		s.Inventory.Spells[i] = 0
	}
	return s
}

// toggleReagents 把清單上對應那幾種藥草勾起來。
func toggleReagents(t *testing.T, s *State, want ...int) {
	t.Helper()
	if s.Pick == nil || s.Pick.Marks == nil {
		t.Fatal("現在不是複選模式")
	}
	for _, r := range want {
		found := false
		for i, e := range s.Pick.Entries {
			if e.Value == r {
				s.Pick.Cursor = i
				s.PickToggle()
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("清單上沒有藥草 %d", r)
		}
	}
}

// 一整趟調藥:M → 打符文 → 勾藥草 → 按 M → 答份數。
//
// In Lor 的配方只要硫磺灰一味,所以這一趟從頭到尾都用得上原版資料。
func TestMixingWalksTheOriginalFourSteps(t *testing.T) {
	s := mixState(t)
	if !s.BeginMix() {
		t.Fatal("按 M 沒有開始調藥")
	}
	// 第一步是符文輸入,不是咒語選單。
	if s.Prompt != PromptSpell {
		t.Fatalf("第一步是 %v,預期符文輸入", s.Prompt)
	}
	s.TypeSpellLetter('I')
	s.TypeSpellLetter('L')
	s.SubmitSpell()

	// 第二步是藥草複選。
	if s.Prompt != PromptPick || !s.PickMulti() {
		t.Fatalf("第二步是 %v(複選 = %v),預期藥草複選", s.Prompt, s.PickMulti())
	}
	toggleReagents(t, s, u5data.ReagentSulfurousAsh)
	s.PickConfirm()

	// 第三步是「要幾份?」,兩位數。
	if !s.AwaitingNumber() {
		t.Fatalf("第三步是 %v,預期問份數", s.Prompt)
	}
	s.AnswerNumber(1)
	s.AnswerNumber(2)
	if got := s.NumberInput(); got != "12" {
		t.Fatalf("兩位數沒收進來:%q", got)
	}
	s.SubmitNumber()

	if got := s.Inventory.Spells[SpellInLor]; got != 12 {
		t.Errorf("調出 %d 份 In Lor,預期 12:%v", got, s.Messages)
	}
	if got := s.Inventory.Reagents[u5data.ReagentSulfurousAsh]; got != 20-12 {
		t.Errorf("硫磺灰剩 %d,預期 %d", got, 20-12)
	}
}

// 藥草不足要**重問份數**,不是取消整個流程。
func TestNotEnoughReagentsReAsksTheAmount(t *testing.T) {
	s := mixState(t)
	s.Inventory.Reagents[u5data.ReagentSulfurousAsh] = 3
	s.BeginMix()
	s.TypeSpellLetter('I')
	s.TypeSpellLetter('L')
	s.SubmitSpell()
	toggleReagents(t, s, u5data.ReagentSulfurousAsh)
	s.PickConfirm()

	// 先要 9 份 —— 只有 3 份,要被擋下並重問。
	s.AnswerNumber(9)
	s.SubmitNumber()
	if !s.AwaitingNumber() {
		t.Fatalf("藥草不足之後沒有重問份數:%v", s.Messages)
	}
	// 「藥草不足!」之後緊接著又問一次份數,所以它不是最後一句。
	if !logged(s, MsgInsufficientReagents) {
		t.Errorf("沒有印「%s」:%v", MsgInsufficientReagents, s.Messages)
	}
	if last := s.Messages[len(s.Messages)-1]; last != MsgHowMuch {
		t.Errorf("重問的那句是 %q,預期 %q", last, MsgHowMuch)
	}
	if s.Inventory.Reagents[u5data.ReagentSulfurousAsh] != 3 {
		t.Error("被擋下卻扣了藥草")
	}
	// 改成 3 份就過。
	s.AnswerNumber(3)
	s.SubmitNumber()
	if got := s.Inventory.Spells[SpellInLor]; got != 3 {
		t.Errorf("調出 %d 份,預期 3:%v", got, s.Messages)
	}
	if got := s.Inventory.Reagents[u5data.ReagentSulfurousAsh]; got != 0 {
		t.Errorf("硫磺灰剩 %d,預期 0", got)
	}
}

// 配錯藥草:藥草照扣,咒語一份都沒有。
func TestWrongRecipeWastesTheReagents(t *testing.T) {
	s := mixState(t)
	s.BeginMix()
	s.TypeSpellLetter('I')
	s.TypeSpellLetter('L')
	s.SubmitSpell()
	// In Lor 只要硫磺灰,這裡故意多勾一味大蒜。
	toggleReagents(t, s, u5data.ReagentSulfurousAsh, u5data.ReagentGarlic)
	s.PickConfirm()
	s.AnswerNumber(2)
	s.SubmitNumber()

	if s.Inventory.Spells[SpellInLor] != 0 {
		t.Error("配方不對卻調出了咒語")
	}
	for _, r := range []int{u5data.ReagentSulfurousAsh, u5data.ReagentGarlic} {
		if got := s.Inventory.Reagents[r]; got != 18 {
			t.Errorf("藥草 %d 剩 %d,預期 18(配錯也要扣)", r, got)
		}
	}
}

// 一味都沒勾就按 M:原版**還是會先問份數**,答完才說「沒東西可調!」。
func TestNothingPickedStillAsksTheAmountFirst(t *testing.T) {
	s := mixState(t)
	s.BeginMix()
	s.TypeSpellLetter('I')
	s.TypeSpellLetter('L')
	s.SubmitSpell()
	s.PickConfirm() // 什麼都沒勾
	if !s.AwaitingNumber() {
		t.Fatal("沒勾任何藥草卻沒問份數 —— 順序與原版不同")
	}
	s.AnswerNumber(1)
	s.SubmitNumber()
	if last := s.Messages[len(s.Messages)-1]; last != MsgNothingToMix {
		t.Errorf("印的是 %q,預期 %q", last, MsgNothingToMix)
	}
}

// 符文湊不出咒語(−2)**不會**中止調藥 —— 原版只擋「什麼都沒打」。
//
// ⚠ 這條看起來像 bug,實際上是原版的 `inc eax; jnz`:只有 −1 被擋。
// 結果是玩家亂打符文照樣能把藥草調成廢渣。照抄。
func TestUnknownRunesStillReachTheReagentPicker(t *testing.T) {
	s := mixState(t)
	s.BeginMix()
	s.TypeSpellLetter('B') // BET 一個字母湊不出任何咒語
	s.SubmitSpell()
	if s.Prompt != PromptPick || !s.PickMulti() {
		t.Fatalf("湊不出咒語就被中止了:%v / %v", s.Prompt, s.Messages)
	}
	toggleReagents(t, s, u5data.ReagentSulfurousAsh)
	s.PickConfirm()
	s.AnswerNumber(1)
	s.SubmitNumber()
	if got := s.Inventory.Reagents[u5data.ReagentSulfurousAsh]; got != 19 {
		t.Errorf("硫磺灰剩 %d,預期 19(藥草照扣)", got)
	}
	for i, n := range s.Inventory.Spells {
		if n != 0 {
			t.Errorf("第 %d 個咒語竟然調出了 %d 份", i, n)
		}
	}
}

// 一種藥草都沒有就按 M:直接一句話,不進符文輸入。
func TestNoReagentsAtAllStopsImmediately(t *testing.T) {
	s := mixState(t)
	for r := range s.Inventory.Reagents {
		s.Inventory.Reagents[r] = 0
	}
	if s.BeginMix() {
		t.Error("沒有藥草卻開始調藥")
	}
	if last := s.Messages[len(s.Messages)-1]; last != MsgNoReagents {
		t.Errorf("印的是 %q,預期 %q", last, MsgNoReagents)
	}
	if s.Prompt == PromptSpell {
		t.Error("竟然進了符文輸入")
	}
}

// 藥草清單只列身上有的。
func TestReagentListShowsOnlyWhatYouOwn(t *testing.T) {
	s := mixState(t)
	s.Inventory.Reagents[u5data.ReagentGinseng] = 0
	s.Inventory.Reagents[u5data.ReagentNightshade] = 0
	s.BeginMix()
	s.TypeSpellLetter('M')
	s.SubmitSpell()
	if s.Pick == nil {
		t.Fatal("沒有開出藥草清單")
	}
	for _, e := range s.Pick.Entries {
		if e.Value == u5data.ReagentGinseng || e.Value == u5data.ReagentNightshade {
			t.Errorf("清單列出了沒有的藥草 %d", e.Value)
		}
	}
	if len(s.Pick.Entries) != u5data.ReagentCount-2 {
		t.Errorf("清單有 %d 項,預期 %d 項", len(s.Pick.Entries), u5data.ReagentCount-2)
	}
}

// logged 回報訊息列裡有沒有某一句。
func logged(s *State, want string) bool {
	for _, m := range s.Messages {
		if m == want {
			return true
		}
	}
	return false
}
