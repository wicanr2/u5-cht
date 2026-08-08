package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

func scrollScene(t *testing.T) *State {
	t.Helper()
	s := upkeepScene(t)
	for i := range s.Inventory.Scrolls {
		s.Inventory.Scrolls[i] = 1
	}
	for i := range s.Inventory.Potions {
		s.Inventory.Potions[i] = 1
	}
	s.Messages = nil
	return s
}

// TestEveryScrollIsReachableFromTheUseList 是這一輪的核心回歸:
// **八捲卷軸、八色藥水、八顆月石都要出現在 U 的清單裡。**
//
// 引擎原本一格都沒有 —— 撿得到、存得進檔、選單上找不到。
// 這條測試釘住的不是效果,是「有沒有入口」。
func TestEveryScrollPotionAndMoonstoneIsReachableFromTheUseList(t *testing.T) {
	s := scrollScene(t)
	for i := range s.Inventory.Moonstones {
		s.Inventory.Moonstones[i] = u5data.Moonstone{Location: u5data.MoonstoneInHand}
	}
	want := map[int]string{}
	for i := 0; i < u5data.ScrollCount; i++ {
		want[UseScrollFirst+i] = "卷軸 " + u5data.ScrollSpell(i)
	}
	for i := 0; i < u5data.PotionCount; i++ {
		want[UsePotionFirst+i] = "藥水 " + u5data.PotionColours[i]
	}
	for i := 0; i < u5data.MoonstoneCount; i++ {
		want[UseMoonstoneFirst+i] = "月石"
	}
	got := map[int]bool{}
	for _, e := range s.usableEntries() {
		got[e.Value] = true
	}
	for v, name := range want {
		if !got[v] {
			t.Errorf("U 的清單裡沒有編號 %d(%s)", v, name)
		}
	}
}

// TestScrollIsConsumedEvenWhenItFails 釘住原版的順序:
// `dec byte_3E030[edi]` 在所有判斷之前 —— 場合不對照樣少一捲。
func TestScrollIsConsumedEvenWhenItFails(t *testing.T) {
	s := scrollScene(t)
	// 召喚惡魔只在戰鬥中有效,而這裡不在戰鬥中。
	s.ReadScroll(ScrollSummonDaemon)
	if n := s.Inventory.Scrolls[ScrollSummonDaemon]; n != 0 {
		t.Errorf("失敗了卻沒扣掉卷軸,還剩 %d 捲", n)
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgNotHere) {
		t.Errorf("沒印「%s」:\n%s", MsgNotHere, strings.Join(s.Messages, "\n"))
	}
}

// TestScrollDurationsDifferFromTheSpells 是這一輪最容易寫錯的一條。
//
// 四個卷軸與同名咒語的持續時間**全部不同**。照 `spellEffect` 轉發會全弄錯,
// 而且不會有任何症狀 —— 效果照樣發生,只是長度不對。
func TestScrollDurationsDifferFromTheSpells(t *testing.T) {
	cases := []struct {
		name       string
		scroll     int
		read       func(s *State) int
		wantScroll int
		wantSpell  int
		spell      int
	}{
		{"光明", ScrollLight, func(s *State) int { return s.LightTurns },
			ScrollLightTurns, GreatLightSpellTurn, SpellVasLor},
		{"防護", ScrollProtection, func(s *State) int { return s.CombatModeTurns },
			ScrollProtectTurns, 20, SpellInSanct},
		{"抗魔", ScrollNegateMagic, func(s *State) int { return s.CombatModeTurns },
			ScrollNegateMagicTurn, 10, SpellInAn},
		{"停時", ScrollNegateTime, func(s *State) int { return s.CombatModeTurns },
			ScrollNegateTimeTurns, TimeStopTurns, SpellAnTym},
	}
	for _, c := range cases {
		if c.wantScroll == c.wantSpell {
			t.Errorf("%s:卷軸與咒語的時間被寫成一樣(%d)—— 原版是不同的",
				c.name, c.wantScroll)
			continue
		}
		s := scrollScene(t)
		s.ReadScroll(c.scroll)
		if got := c.read(s); got != c.wantScroll {
			t.Errorf("%s卷軸給了 %d 回合,原版是 %d", c.name, got, c.wantScroll)
		}
		s2 := scrollScene(t)
		s2.spellEffect(0, c.spell)
		if got := c.read(s2); got != c.wantSpell {
			t.Errorf("%s咒語給了 %d 回合,原版是 %d", c.name, got, c.wantSpell)
		}
	}
}

// TestLightScrollOverwritesVasLor:`sub_1D310` 是指派不是取大值,
// 所以 240 會把 Vas Lor 的 255 蓋掉。
func TestLightScrollOverwritesVasLor(t *testing.T) {
	s := scrollScene(t)
	s.LightTurns = GreatLightSpellTurn // 255
	s.ReadScroll(ScrollLight)
	if s.LightTurns != ScrollLightTurns {
		t.Errorf("光明卷軸該把亮度蓋成 %d,結果是 %d", ScrollLightTurns, s.LightTurns)
	}
}

// TestSummonDaemonAndResurrectionHaveOppositeGates 是第三個陷阱:
// 召喚惡魔**只在戰場上**、復活**只在戰場外**。
func TestSummonDaemonAndResurrectionHaveOppositeGates(t *testing.T) {
	// 戰場外:召喚惡魔失敗、復活可以進到效果那一層。
	s := scrollScene(t)
	s.ReadScroll(ScrollSummonDaemon)
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgNotHere) {
		t.Error("戰場外召喚惡魔該印「此處不可」")
	}
	s = scrollScene(t)
	s.ReadScroll(ScrollResurrection)
	if strings.Contains(strings.Join(s.Messages, "|"), MsgNotHere) {
		t.Error("戰場外的復活卷軸不該印「此處不可」")
	}
}

// TestNegateTimeHasTwoDeadLocations 釘住 0x1D / 0x28 那兩個地點。
func TestNegateTimeHasTwoDeadLocations(t *testing.T) {
	for _, loc := range NegateTimeDeadLocations {
		s := scrollScene(t)
		s.Location = loc
		s.ReadScroll(ScrollNegateTime)
		if !strings.Contains(strings.Join(s.Messages, "|"), MsgNoEffect) {
			t.Errorf("地點 0x%02X 的停時卷軸該印「%s」:\n%s",
				loc, MsgNoEffect, strings.Join(s.Messages, "\n"))
		}
		if s.CombatModeTurns != 0 {
			t.Errorf("地點 0x%02X 竟然還是停了 %d 回合", loc, s.CombatModeTurns)
		}
	}
	// 正對照:別的地點要真的停。
	s := scrollScene(t)
	s.Location = 1
	s.ReadScroll(ScrollNegateTime)
	if s.CombatModeTurns != ScrollNegateTimeTurns {
		t.Errorf("一般地點的停時卷軸沒生效(CombatModeTurns = %d)", s.CombatModeTurns)
	}
}

// TestWindChangeScrollAsksAnywayButOnlyWorksOutdoors —— 方向照問,
// 地牢裡白問(原版 `sub_1CC50` 排在比地點之前)。
func TestWindChangeScrollOnlyWorksOutsideDungeons(t *testing.T) {
	s := scrollScene(t)
	s.Wind = u5data.WindNorth
	if !s.ReadScroll(ScrollWindChange) {
		t.Fatal("大地圖上的換風卷軸該成立")
	}
	if s.Prompt != PromptDirection {
		t.Fatal("換風卷軸沒問方向")
	}
	s.AnswerDirection(East)
	if s.Wind != u5data.WindEast {
		t.Errorf("風向沒改成東,現在是 %d", s.Wind)
	}

	// 地牢裡:照樣問,但改不動。
	s = scrollScene(t)
	s.Wind = u5data.WindNorth
	s.Dungeon = &DungeonState{Index: 0, Location: u5data.DungeonLocationBase}
	if s.ReadScroll(ScrollWindChange) {
		t.Error("地牢裡的換風卷軸該回失敗")
	}
	if s.Prompt != PromptDirection {
		t.Fatal("地牢裡也該照樣問方向 —— 原版問完才比地點")
	}
	s.AnswerDirection(East)
	if s.Wind != u5data.WindNorth {
		t.Errorf("地牢裡的風向竟然被改成 %d", s.Wind)
	}
}
