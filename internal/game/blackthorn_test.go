package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/i18n"
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// arrested 準備一個「被抓進黑棘宮殿」的狀態。
func arrested(t *testing.T) *State {
	t.Helper()
	s := shrineState(t)
	if s == nil {
		return nil
	}
	if s.PartySize < 2 {
		t.Skipf("這份存檔只有 %d 個人,審問要至少兩個人才走得完", s.PartySize)
	}
	return s
}

// TestBlackthornAsksTheFirstCleanShrine:逼問的是第一座還沒被玷污的聖壇。
//
// ⚠ 原版判的是 `byte_3E0E8[i] != 0`,**整個位元組**,不只是 bit 7。
// 復原之後那一格留下 0x7F(仍然非 0),所以復原過的聖壇不會再被挑中 ——
// 寫成 `& 0x80` 的話,玩家復原一座就會被同一座反覆逼問。
func TestBlackthornAsksTheFirstCleanShrine(t *testing.T) {
	s := arrested(t)
	if s == nil {
		return
	}
	if got := s.BlackthornShrine(); got != 0 {
		t.Errorf("全乾淨時該挑第 0 座,挑了第 %d 座", got)
	}
	s.ShrineFlag[0] = 0xFF // 已玷污
	s.ShrineFlag[1] = 0x7F // 玷污過又復原
	if got := s.BlackthornShrine(); got != 2 {
		t.Errorf("復原過的第 1 座不該再被挑中,結果挑了第 %d 座", got)
	}
	for i := range s.ShrineFlag {
		s.ShrineFlag[i] = 0xFF
	}
	if got := s.BlackthornShrine(); got != -1 {
		t.Errorf("八座全玷污了還挑得出第 %d 座", got)
	}
	// 全玷污時整幕不發生。
	s.Messages = nil
	if s.BeginInterrogation() {
		t.Error("八座全玷污了審問還是開始了")
	}
}

// TestConfessDesecratesAndCostsKarma:招了 → 聖壇被玷污、業報 −5。
//
// ⚠ 比對是**子字串**不是前綴(原版 `sub_C098` 拿 edi 滑過整行)——
// 「the mantra is Ahm」也算招供。寫成前綴的話,玩家在真言前加一個字
// 就能白嫖過關,而那正是這一幕唯一的抉擇。
func TestConfessDesecratesAndCostsKarma(t *testing.T) {
	s := arrested(t)
	if s == nil {
		return
	}
	s.Karma = 50
	if !s.BeginInterrogation() {
		t.Fatalf("審問沒有開始:\n%s", s.log())
	}
	v := s.Blackthorn.Virtue
	s.AnswerBlackthorn("the mantra is " + u5data.Shrines[v].Mantra)
	if s.ShrineFlag[v] != 0xFF {
		t.Errorf("招了之後第 %d 座的旗標是 %02X,預期 FF", v, s.ShrineFlag[v])
	}
	if s.Karma != 45 {
		t.Errorf("業報 %d,招供該扣 5 變成 45", s.Karma)
	}
	if s.Prompt == PromptBlackthorn {
		t.Error("招了之後還在審問中")
	}
}

// TestRefusingKeepsTheShrineButCostsACompanion:一路不招 → 聖壇保住,同伴被斬。
//
// ⚠ **招與不招都會死人**,只是訊息不同。寫成「招了就沒事」會把這一幕
// 最重要的張力抹掉,而且玩家會發現招供沒有代價。
func TestRefusingKeepsTheShrineButCostsACompanion(t *testing.T) {
	s := arrested(t)
	if s == nil {
		return
	}
	size := s.PartySize
	// ⚠ 訊息裡是**譯名**(i18n.Name),不是名冊裡的英文原名。
	victim := i18n.Name(s.Roster[1].Name)
	s.Karma = 50
	s.BeginInterrogation()
	v := s.Blackthorn.Virtue
	for i := 0; i < u5data.BlackthornRounds && s.Blackthorn != nil; i++ {
		s.AnswerBlackthorn("no")
	}
	if s.ShrineFlag[v] != 0 {
		t.Errorf("一路不招,第 %d 座卻被玷污了(%02X)", v, s.ShrineFlag[v])
	}
	if s.Karma != 50 {
		t.Errorf("業報 %d,不招不該扣", s.Karma)
	}
	if s.PartySize != size-1 {
		t.Errorf("隊伍剩 %d 人,預期 %d —— 拒絕到底該死一個同伴", s.PartySize, size-1)
	}
	if !strings.Contains(s.log(), victim) {
		t.Errorf("沒有點名被處決的是誰:\n%s", s.log())
	}
}

// TestFirstRefusalOnlyTaunts:第一次拒絕只被嗆,不動刑。
//
// 原版的 `ebx`:第一輪 `and ebx, ebx; jz` 走「只嗆一句」那條,之後才推刑具。
// 少了這個旗標,第一次拒絕就會直接進入處決倒數,四輪變三輪。
func TestFirstRefusalOnlyTaunts(t *testing.T) {
	s := arrested(t)
	if s == nil {
		return
	}
	size := s.PartySize
	s.BeginInterrogation()
	s.AnswerBlackthorn("no")
	if s.PartySize != size {
		t.Error("第一次拒絕就死人了")
	}
	if !strings.Contains(s.log(), s.miscText(u5data.MsgBlackthornTaunt)) {
		t.Errorf("第一次拒絕沒有嗆話:\n%s", s.log())
	}
	if s.Blackthorn == nil || s.Blackthorn.Round != 1 {
		t.Error("第一次拒絕之後沒有進到第二輪")
	}
}

// TestAloneMeansStraightToTheDungeon:隊伍只剩一人時黑棘不玩了。
func TestAloneMeansStraightToTheDungeon(t *testing.T) {
	s := arrested(t)
	if s == nil {
		return
	}
	s.PartySize = 1
	s.BeginInterrogation()
	v := s.Blackthorn.Virtue
	s.AnswerBlackthorn("no")
	if s.Blackthorn != nil {
		t.Error("只剩一人時審問還在繼續")
	}
	if s.ShrineFlag[v] != 0 {
		t.Error("只剩一人拒絕卻還是被玷污了")
	}
	if !strings.Contains(s.log(), s.miscText(u5data.MsgBlackthornLies)) {
		t.Errorf("沒有印「小孩都拆穿得了汝的謊」:\n%s", s.log())
	}
}

// TestExecutedCompanionGetsAnUrn:被處決的人搬到名冊最後一格並打上 0x7F。
//
// ★ 那個 0x7F 就是寶典石室前「骨灰罈」的來源(`sub_1DA10` 掃的是
// `byte_3DDD3[i*32] == 0x7F`)—— `docs/re/27` §5 當時記著「還沒追」,
// 追出來就是這裡。只把狀態標成 'D' 的話,之後去寶典看不到罈子,
// 而且名冊會留一個空洞。
func TestExecutedCompanionGetsAnUrn(t *testing.T) {
	s := arrested(t)
	if s == nil {
		return
	}
	victim := s.Roster[1].Name
	third := ""
	if s.PartySize > 2 {
		third = s.Roster[2].Name
	}
	s.BeginInterrogation()
	for i := 0; i < u5data.BlackthornRounds && s.Blackthorn != nil; i++ {
		s.AnswerBlackthorn("no")
	}
	last := &s.Roster[len(s.Roster)-1]
	if last.Name != victim {
		t.Errorf("名冊最後一格是 %q,預期被處決的 %q", last.Name, victim)
	}
	if last.Raw[u5data.CharUrn] != u5data.CharUrnMark {
		t.Errorf("骨灰罈標記是 %02X,預期 %02X",
			last.Raw[u5data.CharUrn], u5data.CharUrnMark)
	}
	if third != "" && s.Roster[1].Name != third {
		t.Errorf("名冊沒有往前遞補:第 1 格是 %q,預期 %q", s.Roster[1].Name, third)
	}
}

// TestReleasedToTheCellWithoutKeys:不論結局都被丟進地牢,而且鑰匙歸零。
//
// ⚠ 鑰匙不歸零的話玩家一被關進去就能開門走人 —— 原版 `byte_3DFB8 = 0`。
func TestReleasedToTheCellWithoutKeys(t *testing.T) {
	s := arrested(t)
	if s == nil {
		return
	}
	s.Inventory.Keys = 5
	s.BeginInterrogation()
	v := s.Blackthorn.Virtue
	s.AnswerBlackthorn(u5data.Shrines[v].Mantra)
	if s.Inventory.Keys != 0 {
		t.Errorf("還剩 %d 把鑰匙", s.Inventory.Keys)
	}
	if s.Location != u5data.BlackthornLocation {
		t.Errorf("被放在地點 %d,預期黑棘宮殿 %d", s.Location, u5data.BlackthornLocation)
	}
	if s.Floor != u5data.BlackthornCellFloor ||
		s.X != u5data.BlackthornCellX || s.Y != u5data.BlackthornCellY {
		t.Errorf("被放在 (%d,%d) 第 %d 層,預期 (%d,%d) 第 %d 層",
			s.X, s.Y, s.Floor,
			u5data.BlackthornCellX, u5data.BlackthornCellY, u5data.BlackthornCellFloor)
	}
}

// TestMantraSpokenIsSubstringNotPrefix:招供的判定是子字串。
//
// 與聖壇的 `MatchPrefix` 是**兩種不同的比對** —— 原版一個用
// `sub_27C98`(前綴)、一個用 `sub_C098`(滑動子字串)。抄成同一支的話,
// 黑棘會變得很好騙。
func TestMantraSpokenIsSubstringNotPrefix(t *testing.T) {
	cases := []struct {
		typed string
		want  bool
	}{
		{"Ahm", true},
		{"ahm", true},
		{"the mantra is ahm", true},  // ★ 前綴比對會漏掉這一個
		{"I will never say Ahm", true},
		{"no", false},
		{"Mu", false},
	}
	for _, c := range cases {
		if got := u5data.MantraSpoken("Ahm", c.typed); got != c.want {
			t.Errorf("MantraSpoken(%q, %q) = %v,預期 %v", "Ahm", c.typed, got, c.want)
		}
	}
	// 反面對照:同一組字用聖壇那支比對,「the mantra is ahm」就不算。
	if u5data.MatchPrefix("Ahm", "the mantra is ahm") {
		t.Error("MatchPrefix 竟然也吃子字串 —— 兩支比對混在一起了")
	}
}

// TestEveryBlackthornRecordIsTranslated:審問那十二筆都要有譯文。
func TestEveryBlackthornRecordIsTranslated(t *testing.T) {
	s := shrineState(t)
	if s == nil {
		return
	}
	for i := u5data.MsgBlackthornAsk1; i <= u5data.MsgBlackthornWait; i++ {
		txt := s.miscText(i)
		if txt == "" {
			t.Errorf("MISCMSG#%d 是空的", i)
			continue
		}
		// 譯文裡不該還留著英文原句的痕跡(這幾筆全都有中譯)。
		if strings.Contains(txt, "Mantra") || strings.Contains(txt, "Blackthorn") {
			t.Errorf("MISCMSG#%d 看起來還是英文:%q", i, txt)
		}
	}
}
