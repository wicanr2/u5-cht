package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// dungeonLookState 進地牢、點著火把。
func dungeonLookState(t *testing.T) *State {
	t.Helper()
	s := dungeonState(t)
	if s.Dungeon == nil {
		if !findDungeonTile(t, s, u5data.DungeonPassage) {
			t.Skip("進不了地牢")
		}
	}
	s.TorchTurns = 100
	return s
}

// TestDungeonLookIsADifferentProgram —— L 在地牢裡走另一支。
//
// 判準:地牢版**不問絕對方向的四個方位**,而是印相對方向的結果;
// 而且地面版的「汝所見為」後面接的是家具語,地牢版是十六種地形之一。
func TestDungeonLookIsADifferentProgram(t *testing.T) {
	s := dungeonLookState(t)
	d := s.Dungeon
	s.Dungeons.Set(d.Index, d.Level, d.X, d.Y, u5data.DungeonPassage)
	s.Messages = nil
	s.Look()
	if s.Prompt != PromptDirection {
		t.Fatalf("L 之後沒有等方向(Prompt %v)", s.Prompt)
	}
	s.AnswerDirection(South) // ↓ = 腳下
	got := strings.Join(s.Messages, "|")
	if !strings.Contains(got, "一條通道") {
		t.Errorf("看腳下的通道卻印 %q", got)
	}
}

// TestDungeonLookNeedsLight —— 沒光源就是一片漆黑。
//
// ⚠ 比的是**兩個光源計時器**而不是視野半徑 —— 地牢的基礎半徑是 2,
// 拿半徑判會永遠有光。這條就是釘住那個差別。
func TestDungeonLookNeedsLight(t *testing.T) {
	s := dungeonLookState(t)
	s.TorchTurns, s.LightTurns = 0, 0
	s.Messages = nil
	s.Look()
	s.AnswerDirection(South)
	got := strings.Join(s.Messages, "|")
	if !strings.Contains(got, "漆黑") {
		t.Errorf("沒光源卻看得到:%q", got)
	}
	// 正對照:點上火把就看得到了。
	s.TorchTurns = 50
	s.Messages = nil
	s.Look()
	s.AnswerDirection(South)
	if got := strings.Join(s.Messages, "|"); strings.Contains(got, "漆黑") {
		t.Errorf("點了火把還是漆黑:%q", got)
	}
}

// TestDungeonLookFieldsComeBeforeTheJumpTable —— ★ 分派順序。
//
// 0x80 與 0xC0 在原版被**前面的特判**攔掉,所以十六路跳表裡那兩格是死碼。
// 這條驗力場印的是四種之一而**不是**跳表裡的「一片能量場」。
func TestDungeonLookFieldsComeBeforeTheJumpTable(t *testing.T) {
	s := dungeonLookState(t)
	want := map[byte]string{
		0x80: "睡眠場", 0x81: "毒氣場", 0x82: "火牆", 0x83: "電場",
	}
	for tile, w := range want {
		if got := dungeonLookDesc(tile, s); !strings.Contains(got, w) {
			t.Errorf("tile %02X 印 %q,預期含 %q", tile, got, w)
		}
	}
	// 0x84..0x8F 落到 default「能量場」。
	if got := dungeonLookDesc(0x8F, s); !strings.Contains(got, "能量場") {
		t.Errorf("tile 8F 印 %q,預期 default 的能量場", got)
	}
	// ★ 跳表裡索引 8 與 12 那兩格取不到。
	//
	// ⚠⚠ 這條原本比 `== dungeonLookText[0x8]` —— **毫無鑑別力**:
	// 原版索引 8 的字串與力場 default 的字**完全一樣**,兩條路印出來的
	// 東西無法區分。改成在那兩格放哨兵值,任何走到就是分派順序壞了。
	for tile := 0x00; tile <= 0xFF; tile++ {
		if got := dungeonLookDesc(byte(tile), s); got == dungeonLookUnreachable {
			t.Errorf("tile %02X 走到了取不到的跳表格 —— 分派順序被改壞了", tile)
		}
	}
	// 反對照:哨兵確實在表裡(否則上面那個迴圈永遠綠)。
	if dungeonLookText[0x8] != dungeonLookUnreachable ||
		dungeonLookText[0xC] != dungeonLookUnreachable {
		t.Error("跳表索引 8 / 12 沒有放哨兵 —— 上面那個迴圈沒有鑑別力")
	}
	// 原文有留下來供對碼。
	if len(dungeonLookDeadArms) != 2 {
		t.Errorf("死碼原文只記了 %d 筆,預期 2 筆", len(dungeonLookDeadArms))
	}
}

// TestCavedPassageHasThreeOutcomes —— 崩落通道的低四位元分三種。
//
// ⚠ 第三種是 1/255 的反盜版彩蛋 —— 用固定種子掃到它,而不是「跑很多次看看」。
func TestCavedPassageHasThreeOutcomes(t *testing.T) {
	s := dungeonLookState(t)
	if got := dungeonLookDesc(0xC0|CavedStalactite, s); !strings.Contains(got, "石筍") {
		t.Errorf("低位元 1 印 %q,預期石筍", got)
	}
	if got := dungeonLookDesc(0xC0|CavedPassage, s); !strings.Contains(got, "崩落") {
		t.Errorf("低位元 2 印 %q,預期崩落的通道", got)
	}
	// 低位元 0 或 ≥3 → 擲 1..255,只有擲到 255 才是盜版者。
	pirate, other := 0, 0
	for seed := 1; seed <= 400; seed++ {
		s.SeedRandom(int64(seed))
		if strings.Contains(dungeonLookDesc(0xC0, s), "盜版") {
			pirate++
		} else {
			other++
		}
	}
	if other == 0 {
		t.Error("四百次全是盜版者 —— 1/255 的機率不該這樣")
	}
	t.Logf("四百顆種子:盜版者 %d 次、冒險者 %d 次(原版是 1/255)", pirate, other)
}

// TestFountainDrinkEffectsUseTheFullTile —— ★ 喝的效果看**完整 tile** 不是高四位。
func TestFountainDrinkEffectsUseTheFullTile(t *testing.T) {
	cases := []struct {
		tile byte
		name string
	}{
		{FountainCure, "解毒"}, {FountainHeal, "痊癒"},
		{FountainPoison, "中毒"}, {0x5F, "味道很糟"},
	}
	for _, c := range cases {
		s := dungeonLookState(t)
		if s.PartySize < 1 {
			t.Skip("隊伍是空的")
		}
		for i := 0; i < s.PartySize; i++ {
			s.Roster[i].Status = u5data.StatusGood
			s.Roster[i].HP, s.Roster[i].MaxHP = 10, 40
		}
		s.Messages = nil
		s.drinkFromDungeonFountain(c.tile, true)
		got := strings.Join(s.Messages, "|")
		if !strings.Contains(got, c.name) {
			t.Errorf("tile %02X 印 %q,預期含 %q", c.tile, got, c.name)
		}
	}
}

// TestFountainHealFillsToMaxRegardlessOfStatus —— 補血不看死活。
//
// ⚠ 原版那兩行(`word_3DDC4 = word_3DDC6`)沒有任何前置判斷 ——
// 死人的 HP 也會被填滿,但狀態仍是 'D'。照抄,不「順手」加判斷。
func TestFountainHealFillsToMaxRegardlessOfStatus(t *testing.T) {
	s := dungeonLookState(t)
	if s.PartySize < 1 {
		t.Skip("隊伍是空的")
	}
	// 讓 pickCharacter 選到隊員 0:只留它是活的。
	for i := 1; i < s.PartySize; i++ {
		s.Roster[i].Status = u5data.StatusDead
	}
	s.Roster[0].Status = u5data.StatusGood
	s.Roster[0].HP, s.Roster[0].MaxHP = 3, 55
	s.drinkFromDungeonFountain(FountainHeal, true)
	if s.Roster[0].HP != 55 {
		t.Errorf("喝完 HP 是 %d,預期補到上限 55", s.Roster[0].HP)
	}
}

// TestFountainRefusalDoesNothing —— 答 N 什麼都不發生。
func TestFountainRefusalDoesNothing(t *testing.T) {
	s := dungeonLookState(t)
	if s.PartySize < 1 {
		t.Skip("隊伍是空的")
	}
	s.Roster[0].Status = u5data.StatusPoisoned
	s.drinkFromDungeonFountain(FountainCure, false)
	if s.Roster[0].Status != u5data.StatusPoisoned {
		t.Error("答 N 卻還是解了毒")
	}
}

// TestDungeonLookOffersTheDrinkOnlyAtFountains —— 只有噴泉會問。
func TestDungeonLookOffersTheDrinkOnlyAtFountains(t *testing.T) {
	s := dungeonLookState(t)
	d := s.Dungeon
	for _, c := range []struct {
		tile byte
		ask  bool
	}{
		{u5data.DungeonFountain, true},
		{u5data.DungeonFountain | 2, true}, // 高四位一樣 → 照問
		{u5data.DungeonPassage, false},
		{u5data.DungeonChest, false},
	} {
		s.Dungeons.Set(d.Index, d.Level, d.X, d.Y, c.tile)
		s.Prompt = PromptNone
		s.Messages = nil
		s.lookDungeonRelative(SearchHere)
		asked := s.Prompt == PromptYesNo
		if asked != c.ask {
			t.Errorf("tile %02X:問了要不要喝 = %v,預期 %v(訊息 %q)",
				c.tile, asked, c.ask, s.Messages)
		}
		if asked {
			s.AnswerYesNo(false)
		}
	}
}
