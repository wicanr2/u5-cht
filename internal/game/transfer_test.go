package game

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// writeU4Save 造一份最小可用的 U4 PARTY.SAV。
func writeU4Save(t *testing.T, name string, gender byte) string {
	t.Helper()
	raw := make([]byte, 0x140+0xB6)
	rec := raw[0x008:]
	put := func(off, v int) { binary.LittleEndian.PutUint16(rec[off:], uint16(v)) }
	put(0x00, 180) // HP
	put(0x02, 350) // 最大 HP
	put(0x04, 4000)
	put(0x06, 40) // 力量
	put(0x08, 34) // 敏捷
	put(0x0A, 50) // 智力
	put(0x0C, 9)
	copy(rec[0x14:], name)
	rec[0x24] = gender
	rec[0x25] = 5 // Paladin
	path := filepath.Join(t.TempDir(), u5data.U4TransferFile)
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// 轉入的第一步是問「保留此名否」,不是直接把角色塞進名冊。
func TestTransferAsksBeforeInstalling(t *testing.T) {
	s := &State{Clock: NewClock(), MaxMessages: 32}
	if !s.TransferFromUltimaIV(writeU4Save(t, "AVATAR", u5data.GenderMale)) {
		t.Fatalf("轉入沒開始:%v", s.Messages)
	}
	if !s.AwaitingYesNo() {
		t.Fatalf("沒有在問問題,Prompt = %v", s.Prompt)
	}
	if len(s.Roster) != 0 {
		t.Error("還在問問題就把角色放進名冊了")
	}
	// 兩個問題都答 Y → 名字與性別都保留。
	s.AnswerYesNo(true) // 保留名字
	if !s.AwaitingYesNo() {
		t.Fatal("第二個問題沒問出來")
	}
	s.AnswerYesNo(true) // 保留性別
	if len(s.Roster) != 1 {
		t.Fatalf("答完之後名冊有 %d 人", len(s.Roster))
	}
	c := &s.Roster[0]
	if c.Name != "AVATAR" {
		t.Errorf("名字變成 %q", c.Name)
	}
	if c.Gender != u5data.GenderMale {
		t.Errorf("性別變成 0x%02X", c.Gender)
	}
	// 換算跑過了(經驗 4000 / 10 = 400 → 4 級 → HP 120)。
	if c.Exp != 400 || c.Level != 4 || c.MaxHP != 120 {
		t.Errorf("換算沒跑:經驗 %d 等級 %d 生命 %d", c.Exp, c.Level, c.MaxHP)
	}
}

// 答 N 就問新名字;打了就換,**什麼都不打就保留原名**。
func TestTransferRenameKeepsOldNameWhenBlank(t *testing.T) {
	for _, c := range []struct {
		typed string
		want  string
	}{
		{"CONNOR", "CONNOR"},
		{"", "AVATAR"}, // ★ 空字串保留原名(原版 `cmp byte_3DDB4, 0; jz`)
		{"VERYLONGNAMEHERE", "VERYLONG"}, // 上限 8 個字元
	} {
		s := &State{Clock: NewClock(), MaxMessages: 32}
		s.TransferFromUltimaIV(writeU4Save(t, "AVATAR", u5data.GenderMale))
		s.AnswerYesNo(false) // 不保留名字
		if !s.AwaitingText() {
			t.Fatalf("答 N 之後沒有問新名字,Prompt = %v", s.Prompt)
		}
		for _, r := range c.typed {
			s.TypeText(r)
		}
		s.SubmitText()
		s.AnswerYesNo(true) // 性別保留
		if got := s.Roster[0].Name; got != c.want {
			t.Errorf("打 %q → 名字 %q,預期 %q", c.typed, got, c.want)
		}
		// Raw 也要跟著改,否則存檔寫回舊名字。
		raw := strings.TrimRight(string(s.Roster[0].Raw[u5data.CharName:u5data.CharName+u5data.CharNameLen]), "\x00")
		if raw != c.want {
			t.Errorf("打 %q → Raw 裡是 %q,預期 %q", c.typed, raw, c.want)
		}
	}
}

// ★ 「保留原性別?」的 N 是**翻轉**,不是「設成女」。
//
// 寫成「Y = 男 / N = 女」的話,一個女角色答 Y 會變成男的 —— 那是原版沒有的行為。
func TestTransferSexAnswerFlipsRatherThanSets(t *testing.T) {
	cases := []struct {
		from byte
		keep bool
		want byte
	}{
		{u5data.GenderMale, true, u5data.GenderMale},
		{u5data.GenderMale, false, u5data.GenderFemale},
		{u5data.GenderFemale, true, u5data.GenderFemale}, // ← 這條擋住「Y = 男」的寫法
		{u5data.GenderFemale, false, u5data.GenderMale},
	}
	for _, c := range cases {
		s := &State{Clock: NewClock(), MaxMessages: 32}
		s.TransferFromUltimaIV(writeU4Save(t, "AVATAR", c.from))
		s.AnswerYesNo(true) // 名字保留
		s.AnswerYesNo(c.keep)
		got := s.Roster[0].Gender
		if got != c.want {
			t.Errorf("原本 0x%02X、答 %v → 0x%02X,預期 0x%02X", c.from, c.keep, got, c.want)
		}
		if s.Roster[0].Raw[u5data.CharGender] != got {
			t.Error("Raw 的性別沒跟著改")
		}
	}
}

// 讀不到檔要印原版那兩句,而且不留下半途的狀態。
func TestTransferFailurePrintsTheTwoLines(t *testing.T) {
	s := &State{Clock: NewClock(), MaxMessages: 32}
	if s.TransferFromUltimaIV(filepath.Join(t.TempDir(), "NOPE.SAV")) {
		t.Error("讀不到檔卻回報成功")
	}
	if len(s.Messages) < 2 ||
		s.Messages[len(s.Messages)-2] != MsgTransferError ||
		s.Messages[len(s.Messages)-1] != MsgTransferUnable {
		t.Errorf("印的不是原版那兩句:%v", s.Messages)
	}
	if s.AwaitingYesNo() || s.AwaitingText() {
		t.Error("失敗卻留在提問狀態")
	}
}
