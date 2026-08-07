package game

import (
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// castState 造一個能施法的最小狀態(讀真的咒語表與符文表)。
func castState(t *testing.T) *State {
	t.Helper()
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	sp, err := u5data.LoadSpells(dir)
	if err != nil {
		t.Fatal(err)
	}
	rt, err := u5data.LoadRuneTable(dir)
	if err != nil {
		t.Fatal(err)
	}
	s := &State{Spells: sp, Runes: rt, Clock: NewClock(), MaxMessages: 32}
	s.Roster = []u5data.Character{{
		Name: "AVATAR", Class: 'A', Status: u5data.StatusGood,
		Level: 8, MP: 40, HP: 100, MaxHP: 100,
	}}
	s.PartySize = 1
	for i := range s.Inventory.Spells {
		s.Inventory.Spells[i] = 9
	}
	return s
}

// 打 I、L 兩個字母就是 In Lor —— 畫面上長出的是整個符文詞。
func TestCastingTakesRuneInitialsNotTheSpellName(t *testing.T) {
	s := castState(t)
	s.BeginCastPrompt()
	if s.Prompt != PromptSpell {
		t.Fatalf("按 C 沒有進入施法輸入:%v", s.Prompt)
	}
	s.TypeSpellLetter('i') // 小寫照樣收
	if s.Input != "IN" {
		t.Errorf("打 i 之後回顯是 %q,預期 \"IN\"", s.Input)
	}
	s.TypeSpellLetter('L')
	if s.Input != "IN LOR" {
		t.Errorf("打 iL 之後回顯是 %q,預期 \"IN LOR\"", s.Input)
	}
	if got := s.SpellLetters(); got != "IL" {
		t.Errorf("收到的字母是 %q,預期 \"IL\"", got)
	}
	s.SubmitSpell()
	if s.LightTurns != LightSpellTurns {
		t.Errorf("In Lor 沒生效:LightTurns = %d", s.LightTurns)
	}
	if s.Prompt == PromptSpell {
		t.Error("送出之後還留在輸入模式")
	}
}

// 收不下的鍵要**默默丟掉**:J / O、數字、以及第五個字母。
func TestSpellInputSilentlyDropsWhatItCannotTake(t *testing.T) {
	s := castState(t)
	s.BeginCastPrompt()
	before := len(s.Messages)
	for _, r := range []rune{'J', 'O', '4', '-'} {
		s.TypeSpellLetter(r)
	}
	if got := s.SpellLetters(); got != "" {
		t.Errorf("J / O / 數字竟然收進來了:%q", got)
	}
	if len(s.Messages) != before {
		t.Errorf("被丟掉的鍵印了訊息:%v", s.Messages[before:])
	}
	// 四個是上限,第五個丟掉。
	for _, r := range []rune{'A', 'B', 'C', 'D', 'E'} {
		s.TypeSpellLetter(r)
	}
	if got := s.SpellLetters(); got != "ABCD" {
		t.Errorf("上限沒守住:%q", got)
	}
	// Backspace 退一個,連回顯一起退。
	s.BackspaceSpell()
	if got := s.SpellLetters(); got != "ABC" {
		t.Errorf("Backspace 之後是 %q,預期 \"ABC\"", got)
	}
	if strings.Contains(s.Input, "DES") {
		t.Errorf("回顯還留著退掉的那個詞:%q", s.Input)
	}
}

// 空白鍵就是送出 —— 與 Enter 同義。
func TestSpaceSubmitsLikeEnter(t *testing.T) {
	s := castState(t)
	s.BeginCastPrompt()
	s.TypeSpellLetter('I')
	s.TypeSpellLetter('L')
	if submitted := s.TypeSpellLetter(' '); !submitted {
		t.Fatal("空白鍵沒有送出")
	}
	if s.LightTurns != LightSpellTurns {
		t.Errorf("空白鍵送出之後咒語沒生效:%v", s.Messages)
	}
}

// 兩種失敗訊息不一樣,而且都不消耗份數。
func TestEmptyAndUnknownSpellSayDifferentThings(t *testing.T) {
	s := castState(t)
	// 什麼都沒打就送出 → 「無!」
	s.BeginCastPrompt()
	s.SubmitSpell()
	if last := s.Messages[len(s.Messages)-1]; last != MsgSpellNone {
		t.Errorf("空輸入印的是 %q,預期 %q", last, MsgSpellNone)
	}
	// ESC 走同一句。
	s.BeginCastPrompt()
	s.TypeSpellLetter('I')
	s.CancelSpell()
	if last := s.Messages[len(s.Messages)-1]; last != MsgSpellNone {
		t.Errorf("ESC 印的是 %q,預期 %q", last, MsgSpellNone)
	}
	// 湊不出咒語 → 「毫無效果!」
	s.BeginCastPrompt()
	s.TypeSpellLetter('B')
	s.SubmitSpell()
	if last := s.Messages[len(s.Messages)-1]; last != MsgSpellNoEffect {
		t.Errorf("湊不出來印的是 %q,預期 %q", last, MsgSpellNoEffect)
	}
	// 三次都沒有動到任何份數。
	for i, n := range s.Inventory.Spells {
		if n != 9 {
			t.Errorf("第 %d 個咒語的份數變成 %d —— 失敗的輸入不該扣", i, n)
		}
	}
	if s.Roster[0].MP != 40 {
		t.Errorf("魔力變成 %d —— 失敗的輸入不該扣", s.Roster[0].MP)
	}
}

// 打字母的順序不影響結果,而且真的會施到同一個咒語。
func TestTypingRunesInAnyOrderCastsTheSameSpell(t *testing.T) {
	for _, in := range []string{"IL", "LI"} {
		s := castState(t)
		s.BeginCastPrompt()
		for _, r := range in {
			s.TypeSpellLetter(r)
		}
		s.SubmitSpell()
		if s.LightTurns != LightSpellTurns {
			t.Errorf("打 %q 沒有施出 In Lor:%v", in, s.Messages)
		}
	}
}

// 兩個地點會把咒語吸走,而且**不扣份數也不扣魔力**。
func TestMagicIsAbsorbedAtTheTwoRegaliaSites(t *testing.T) {
	cases := []struct {
		name     string
		loc      int
		crown    bool
		absorbed bool
	}{
		{"第二座城堡・還沒拿王冠", u5data.CrownNPCLocation, false, true},
		{"第二座城堡・拿到王冠", u5data.CrownNPCLocation, true, false},
		{"STONEGATE・還沒拿王冠", u5data.SceptreNPCLocation, false, true},
		{"STONEGATE・拿到王冠", u5data.SceptreNPCLocation, true, true}, // 永遠壓制
		{"不列顛城", 2, false, false},
	}
	for _, c := range cases {
		s := castState(t)
		s.Location = c.loc
		s.Regalia.Crown = c.crown
		got := s.Cast(0, SpellInLor)
		if (got == MagicAbsorbed) != c.absorbed {
			t.Errorf("%s:結果是 %v,預期吸收 = %v", c.name, got, c.absorbed)
		}
		if !c.absorbed {
			continue
		}
		if s.Inventory.Spells[SpellInLor] != 9 {
			t.Errorf("%s:份數被扣掉了", c.name)
		}
		if s.Roster[0].MP != 40 {
			t.Errorf("%s:魔力被扣掉了", c.name)
		}
	}
}
