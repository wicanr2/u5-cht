package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// TestSearchFindsTheSecretDoor:有缺口的牆(0x4E)搜得出密門,普通牆(0x4F)搜不出。
//
// 兩格差一號、行為完全不同。`LOOK2` 把 0x4E 叫「a wall with a nick」——
// 那個缺口就是給玩家的提示。搞混的話密門變成隨處都有或永遠找不到。
func TestSearchFindsTheSecretDoor(t *testing.T) {
	s := searchScene(t)
	x, y := s.X+1, s.Y
	s.SetTileAt(x, y, SearchSecretDoor)
	s.searchAt(x, y)
	if got := s.TileAt(x, y); got != u5data.TileDoorA {
		t.Errorf("密門沒開,那格是 0x%02X", got)
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgHiddenDoor) {
		t.Errorf("沒印出密門:%q", s.Messages)
	}

	s2 := searchScene(t)
	s2.SetTileAt(x, y, 0x4F) // 普通的牆
	s2.searchAt(x, y)
	if got := s2.TileAt(x, y); got != 0x4F {
		t.Errorf("普通的牆也開了門:0x%02X", got)
	}
}

// TestSearchPhraseComesFromTheFurniture:地點語是句子的前半,不是額外一行。
//
// 原版的字串以 `t` 結尾(`"\nIn the stump\nt"` + `"hou dost find"`)。
// 拆成兩行印的話畫面上會多一行空隙,而且對照原版截圖會對不上。
func TestSearchPhraseComesFromTheFurniture(t *testing.T) {
	s := searchScene(t)
	x, y := s.X+1, s.Y
	s.SetTileAt(x, y, 0x2B) // a hollow stump
	s.searchAt(x, y)
	first := s.Messages[0]
	if !strings.HasPrefix(first, "在樹洞裡") {
		t.Errorf("第一行是 %q,該以地點語開頭", first)
	}
	if !strings.Contains(first, MsgThouDostFind) {
		t.Errorf("地點語與「汝翻到了」不在同一行:%q", first)
	}
}

// TestTrapDetectionMakesBothKindsOfMistake:兩種錯都會發生。
//
// 這是特色不是 bug。只做「漏看」的話玩家會學到「說沒陷阱就一定安全」,
// 而原版不是 —— 兩種錯都要在。
//
// ⚠ 這條**必須是統計性的**。第一版寫成「智力 0 就一定漏看」,但難度只有
// 一半要壓過 random(1,30):等級 25 + 智力 0 的難度是 55、一半 27,
// 擲到 27..30 就還是看清了(4/30)。單次斷言於是隨種子飄 ——
// 而它「通過」的時候完全看不出前提是錯的。
func TestTrapDetectionMakesBothKindsOfMistake(t *testing.T) {
	s := searchScene(t)
	for i := range s.Roster {
		s.Roster[i].Intel = 0 // 智力越低,兩種錯越常出現
		s.Roster[i].Status = u5data.StatusGood
	}
	member := s.PartySize - 1

	missed, phantom := 0, 0
	for i := 0; i < 400; i++ {
		// 有陷阱卻報「沒有陷阱」= 漏看
		s.Messages = nil
		trapped := &u5data.MapObject{}
		trapped.Raw[u5data.ObjQuality] = 0x80 | 25
		s.reportTrap(trapped, member)
		if strings.Contains(strings.Join(s.Messages, "|"), MsgNoTrap) {
			missed++
		}
		// 沒陷阱卻報「有陷阱」= 幻覺
		s.Messages = nil
		clean := &u5data.MapObject{}
		s.reportTrap(clean, member)
		if strings.Contains(strings.Join(s.Messages, "|"), MsgATrap) {
			phantom++
		}
	}
	if missed == 0 {
		t.Error("四百次都沒漏看真陷阱 —— 偵測判定可能寫成必定成功")
	}
	if phantom == 0 {
		t.Error("四百次都沒出現幻覺陷阱 —— 少了原版的另一半錯誤")
	}
}

// TestTrapDetectionDescribesTheLevel:看清了才分「簡單 / 複雜」。
//
// 智力 30 讓難度歸零,`detected` 必為 true。
func TestTrapDetectionDescribesTheLevel(t *testing.T) {
	for _, c := range []struct {
		level byte
		want  string
	}{
		{5, MsgSimpleTrap},   // < 10
		{25, MsgComplexTrap}, // > 20
		{15, MsgATrap},       // 中間
	} {
		s := searchScene(t)
		for i := range s.Roster {
			s.Roster[i].Intel = 30
			s.Roster[i].Status = u5data.StatusGood
		}
		o := &u5data.MapObject{}
		o.Raw[u5data.ObjQuality] = 0x80 | c.level
		s.reportTrap(o, s.PartySize-1)
		if !strings.Contains(strings.Join(s.Messages, "|"), c.want) {
			t.Errorf("等級 %d 該報「%s」,實際 %q", c.level, c.want, s.Messages)
		}
	}
}

// TestSearchIsMostlyJunk:八分之七是垃圾。
//
// 把機率調高會讓搜家具變成刷錢手段 —— 原版刻意讓它不划算。
// 這條用固定種子跑一百次,只驗「命中率遠低於一半」,不驗確切次數
//(那會綁死亂數實作)。
func TestSearchIsMostlyJunk(t *testing.T) {
	s := searchScene(t)
	for i := range s.Roster {
		s.Roster[i].Status = u5data.StatusGood
	}
	hits := 0
	for i := 0; i < 200; i++ {
		s.Messages = nil
		s.rollSearchFind(s.PartySize - 1)
		j := strings.Join(s.Messages, "|")
		if strings.Contains(j, MsgFoundGold) || strings.Contains(j, MsgFoundFood) {
			hits++
		}
	}
	if hits == 0 {
		t.Error("兩百次一次都沒翻到 —— 機率算錯了")
	}
	if hits > 60 {
		t.Errorf("兩百次翻到 %d 次,原版是八分之一 —— 太容易了", hits)
	}
}

// TestSearchPhraseTilesAllHaveLookText:地點語的每個 tile 在敘述表裡都是家具。
//
// 這是交叉檢查:地點語表若抄錯一個號碼,那一格在 `LOOK2` 會是水或草地 ——
// 而「在草地裡翻到金幣」看起來只是怪,不像 bug。
func TestSearchPhraseTilesAllHaveLookText(t *testing.T) {
	lt := loadLookOrSkipGame(t)
	for tile := range searchPhrase {
		desc := lt.Terrain(int(tile))
		if u5data.LookPlaceholder(desc) || desc == "" {
			t.Errorf("0x%02X 在敘述表裡沒有東西(%q)—— 地點語表可能抄錯", tile, desc)
		}
	}
}

func loadLookOrSkipGame(t *testing.T) *u5data.LookTable {
	t.Helper()
	s := searchScene(t)
	if s.Look2 == nil {
		t.Skip("沒有敘述表")
	}
	return s.Look2
}

func searchScene(t *testing.T) *State {
	t.Helper()
	s := newLookState(t)
	s.MaxMessages = 16
	s.Messages = nil
	s.SeedRandom(7)
	return s
}
