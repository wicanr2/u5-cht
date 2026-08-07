package game

import (
	"strings"
	"testing"
)

// 每個指令都要有回顯,而且**只有原版不印的那幾個**才是空的。
//
// ⚠ 這條擋的是「加了新指令卻忘了回顯」:原版沒有任何一個字母指令是靜靜跑掉的,
// 所以清單上出現新的空字串就是漏了。
func TestEveryCommandEchoesItsName(t *testing.T) {
	s := &State{MaxMessages: 8}
	// 原版分派器不 push 字串的:空白(由 Pass 自己印)、A(三個位置各自的處理
	// 函式才印)、以及**地表的 E**(交給處理函式;離開地表就會印了)。
	silent := map[rune]bool{' ': true, 'A': true, 'E': true}
	for key := 'A'; key <= 'Z'; key++ {
		echo := s.CommandEcho(key)
		switch {
		case silent[key] && echo != "":
			t.Errorf("%c 應該不回顯,卻印 %q", key, echo)
		case !silent[key] && echo == "":
			t.Errorf("%c 沒有回顯 —— 原版每個字母指令都先印自己的名字", key)
		}
	}
	if got := s.CommandEcho(' '); got != "" {
		t.Errorf("空白鍵的回顯是 %q,預期空(由 Pass 自己印)", got)
	}
}

// 要問方向的指令,名字必須以「——」結尾;不問方向的不能有。
//
// 那個破折號就是原版的「等方向」提示(`Get-`),所以它同時是**功能標記**,
// 不只是標點。標錯的話方向提示會多印一行,或者按下去完全沒反應。
func TestOnlyDirectionCommandsEndWithTheDash(t *testing.T) {
	s := &State{MaxMessages: 8}
	// 原版以 `-` 結尾的:Fire Get Jimmy Klimb Open Search Talk Hole-up,
	// 加上 Look 在**非地牢**時。
	wantDash := map[rune]bool{
		'F': true, 'G': true, 'J': true, 'K': true,
		'O': true, 'S': true, 'T': true, 'H': true, 'L': true,
	}
	for key := 'A'; key <= 'Z'; key++ {
		echo := s.CommandEcho(key)
		has := strings.HasSuffix(echo, dirSuffix)
		if has != wantDash[key] {
			t.Errorf("%c 回顯 %q:破折號 = %v,預期 %v", key, echo, has, wantDash[key])
		}
	}
}

// L 與 S 的回顯**依所在位置而不同** —— 地牢裡不問方向。
func TestLookAndSearchEchoDependOnWhereYouAre(t *testing.T) {
	s := &State{MaxMessages: 8}
	for _, key := range []rune{'L', 'S'} {
		out := s.CommandEcho(key)
		if !strings.HasSuffix(out, dirSuffix) {
			t.Errorf("地表的 %c 回顯 %q,預期以破折號結尾(要問方向)", key, out)
		}
	}
	// 進地牢:兩者都變成「……」,不問方向。
	s.Dungeon = &DungeonState{}
	for _, key := range []rune{'L', 'S'} {
		out := s.CommandEcho(key)
		if !strings.HasSuffix(out, ellipsis) {
			t.Errorf("地牢的 %c 回顯 %q,預期以刪節號結尾(不問方向)", key, out)
		}
	}
}

// E 在地表交給處理函式,其餘地方印「進入什麼?」。
func TestEnterEchoOnlyOutsideTheOverworld(t *testing.T) {
	s := &State{MaxMessages: 8}
	if got := s.CommandEcho('E'); got != "" {
		t.Errorf("地表的 E 回顯 %q,預期空(交給處理函式)", got)
	}
	s.Location = 2
	if got := s.CommandEcho('E'); got == "" {
		t.Error("城鎮裡的 E 沒有回顯,預期「進入什麼?」")
	}
}

// 指令名以「——」結尾時,方向會被接到**同一則**訊息後面(原版 `Get-North`)。
func TestDirectionIsAppendedToTheCommandName(t *testing.T) {
	s := &State{MaxMessages: 8}
	s.EchoCommand('G')
	n := len(s.Messages)
	s.AskDirection(func(Direction) {})
	if len(s.Messages) != n {
		t.Errorf("指令名已經以破折號結尾了,卻又多印一行:%v", s.Messages)
	}
	s.AnswerDirection(North)
	last := s.Messages[len(s.Messages)-1]
	if !strings.HasSuffix(last, North.Name()) {
		t.Errorf("方向沒有接在指令名後面:%q", last)
	}
	if !strings.HasPrefix(last, "拿取") {
		t.Errorf("接錯了行:%q", last)
	}
}

// 沒有指令名在前面時,方向提示照樣要自己印一行 —— 例如咒語的方向。
func TestDirectionPromptStillPrintsOnItsOwn(t *testing.T) {
	s := &State{MaxMessages: 8}
	s.Log("施法……")
	n := len(s.Messages)
	s.AskDirection(func(Direction) {})
	if len(s.Messages) == n {
		t.Error("前面不是等方向的指令名,方向提示卻沒印")
	}
}

// 按到非指令鍵印「What?」,不是靜靜吃掉。
func TestUnknownKeySaysWhat(t *testing.T) {
	s := &State{MaxMessages: 8}
	s.UnknownCommand()
	if last := s.Messages[len(s.Messages)-1]; last != MsgWhat {
		t.Errorf("印的是 %q,預期 %q", last, MsgWhat)
	}
}

// 小寫也要吃 —— 玩家不會為了打指令去按 Shift。
func TestEchoAcceptsLowercase(t *testing.T) {
	s := &State{MaxMessages: 8}
	if !s.EchoCommand('g') {
		t.Fatal("小寫 g 沒有回顯")
	}
	if !strings.HasPrefix(s.Messages[len(s.Messages)-1], "拿取") {
		t.Errorf("小寫 g 印成 %q", s.Messages[len(s.Messages)-1])
	}
}
