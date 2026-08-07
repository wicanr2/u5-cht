package game

import (
	"os"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// TestCreationRunsSevenQuestions:賽制是 4 + 2 + 1,不是七題平鋪。
//
// 輪次結構決定了「哪些美德會對上」:每輪之間清掉「本輪抽過」但保留
// 「已淘汰」。併成一個旗標的話第二輪會重新抽到已淘汰的美德 ——
// 而那看起來只是「題目怪怪的」,不會壞掉。
func TestCreationRunsSevenQuestions(t *testing.T) {
	s := newCreateState(t)
	s.BeginCreation()
	s.AdvanceCreation() // 過開場白

	asked := 0
	for s.Create.Stage == CreationQuestion {
		a, b := s.Create.A, s.Create.B
		if a == b {
			t.Fatalf("第 %d 題把同一個美德對上自己(%d)", asked, a)
		}
		if !s.Create.Alive[a] || !s.Create.Alive[b] {
			t.Fatalf("第 %d 題抽到已淘汰的美德(%d vs %d)", asked, a, b)
		}
		s.AnswerCreation(asked%2 == 0)
		asked++
		if asked > 20 {
			t.Fatal("問不完 —— 賽制寫錯了")
		}
	}
	if asked != u5data.CreateQuestions {
		t.Errorf("問了 %d 題,原版是 %d 題", asked, u5data.CreateQuestions)
	}
	alive := 0
	for _, ok := range s.Create.Alive {
		if ok {
			alive++
		}
	}
	if alive != 1 {
		t.Errorf("最後剩 %d 個美德,應該只剩 1 個", alive)
	}
}

// TestCreationAccumulatesStats:七題的加分要**累加**,不是只算最後一題。
//
// ⚠ 這是 Hex-Rays 的陷阱:它把 `add [eax+2Ah], dl` 反編譯成賦值。
// 照著寫的話屬性會少掉六題份,而角色仍然「看起來正常」——
// 只是弱得莫名其妙,沒有人查得出來。
func TestCreationAccumulatesStats(t *testing.T) {
	s := newCreateState(t)
	s.BeginCreation()
	base := *s.Create
	s.AdvanceCreation()

	want := [3]int{base.Intel, base.Dex, base.Str}
	for s.Create.Stage == CreationQuestion {
		w := s.Create.A
		b := u5data.VirtueBonus[w]
		want[0] += b[u5data.BonusIntel]
		want[1] += b[u5data.BonusDex]
		want[2] += b[u5data.BonusStr]
		s.AnswerCreation(true) // 一律選 A
	}
	got := [3]int{s.Create.Intel, s.Create.Dex, s.Create.Str}
	if got != want {
		t.Errorf("三圍 %v,預期 %v(七題累加)", got, want)
	}
	// 七題最多加 2 分,起點又不是 0 —— 總和一定比起點大。
	if got[0]+got[1]+got[2] <= base.Intel+base.Dex+base.Str {
		t.Error("七題下來三圍完全沒漲 —— 累加沒生效")
	}
}

// TestCreationWritesTheCharacter:名字、性別、三圍要真的寫進名冊。
//
// 力量有下限 20(原版 `cmp byte_57190, 14h`),而魔力等於智力
//(原版 byte_3DDC2 與 byte_3DDC3 寫同一個值)—— 兩條都容易漏。
func TestCreationWritesTheCharacter(t *testing.T) {
	s := newCreateState(t)
	s.BeginCreation()
	s.AdvanceCreation()
	for s.Create.Stage == CreationQuestion {
		s.AnswerCreation(true)
	}
	s.AdvanceCreation() // 過結語
	for _, r := range "阿凡達" {
		s.TypeCreationName(string(r))
	}
	s.ConfirmCreationName()
	s.AnswerCreationGender(true)

	ch := &s.Roster[0]
	if ch.Name != "阿凡達" {
		t.Errorf("名字是 %q", ch.Name)
	}
	if ch.Gender != u5data.GenderMale {
		t.Errorf("性別是 0x%02X,預期 0x%02X", ch.Gender, u5data.GenderMale)
	}
	if int(ch.Intel) != s.Create.Intel {
		t.Errorf("智力 %d,預期 %d", ch.Intel, s.Create.Intel)
	}
	if ch.MP != ch.Intel {
		t.Errorf("魔力 %d 該等於智力 %d", ch.MP, ch.Intel)
	}
	if int(ch.Strength) < u5data.CreateMinStrength {
		t.Errorf("力量 %d 低於下限 %d", ch.Strength, u5data.CreateMinStrength)
	}
	if s.Prompt != PromptNone {
		t.Errorf("做完之後還卡在 Prompt %v", s.Prompt)
	}
}

// TestCreationNameHasALimit:名字欄位只有 9 B,留一格給結尾的 NUL。
func TestCreationNameHasALimit(t *testing.T) {
	s := newCreateState(t)
	s.BeginCreation()
	s.Create.Stage = CreationName
	for i := 0; i < 30; i++ {
		s.TypeCreationName("A")
	}
	if n := len([]rune(s.Create.Name)); n != u5data.CharNameLen-1 {
		t.Errorf("名字長 %d,上限該是 %d", n, u5data.CharNameLen-1)
	}
	s.TypeCreationName("") // 退格
	if n := len([]rune(s.Create.Name)); n != u5data.CharNameLen-2 {
		t.Errorf("退格之後長 %d", n)
	}
}

// TestVirtueQuestionsAreSymmetricAndComplete:28 個配對各有一題,且對稱。
//
// 表是從執行檔抽出來的 8×8 位移矩陣換算的。少一格、或哪裡不對稱,
// 都會讓某個配對抽到別人的題目 —— 而題目本身讀起來仍然通順。
func TestVirtueQuestionsAreSymmetricAndComplete(t *testing.T) {
	seen := map[int]bool{}
	for a := 0; a < u5data.VirtueCount; a++ {
		for b := 0; b < u5data.VirtueCount; b++ {
			q := u5data.VirtueQuestion(a, b)
			if a == b {
				if q != 0 {
					t.Errorf("%d 對自己不該有題目,卻是 %d", a, q)
				}
				continue
			}
			if q != u5data.VirtueQuestion(b, a) {
				t.Errorf("(%d,%d) 與 (%d,%d) 不對稱", a, b, b, a)
			}
			if q < 2 {
				t.Errorf("(%d,%d) 指到記錄 %d —— 那是開場或結語,不是題目", a, b, q)
			}
			seen[q] = true
		}
	}
	if len(seen) != 28 {
		t.Errorf("只用到 %d 個題目,8 取 2 應該是 28 個", len(seen))
	}
}

func newCreateState(t *testing.T) *State {
	t.Helper()
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	s := realState(t, dir)
	q, err := u5data.LoadText(dir + "/QUESTION.DAT")
	if err != nil {
		t.Fatalf("讀 QUESTION.DAT:%v", err)
	}
	s.Question = q
	s.SeedRandom(1)
	return s
}
