package game

import (
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// TestLevelThresholdsDouble:等級門檻是 100 起跳、之後每次翻倍。
//
// 算式是「經驗值 / 100 之後一直右移到 0,移幾次就加幾級」——
// 用邊界值驗最有效:99 與 100、199 與 200、6399 與 6400 各差一級。
func TestLevelThresholdsDouble(t *testing.T) {
	cases := []struct{ exp, level int }{
		{0, 1}, {99, 1},
		{100, 2}, {199, 2},
		{200, 3}, {399, 3},
		{400, 4}, {799, 4},
		{800, 5}, {1599, 5},
		{1600, 6}, {3199, 6},
		{3200, 7}, {6399, 7},
		{6400, 8}, {9999, 8},
	}
	for _, c := range cases {
		if got := LevelForExp(c.exp); got != c.level {
			t.Errorf("經驗值 %d 算出第 %d 級,預期第 %d 級", c.exp, got, c.level)
		}
	}
	// 經驗值上限 9999 → 最高 8 級,而 8 也正好是最高的咒語圈數。
	if LevelForExp(9999) != MaxCharacterLevel {
		t.Errorf("經驗值上限算出第 %d 級,預期 %d", LevelForExp(9999), MaxCharacterLevel)
	}
}

// TestManaByClass:魔力上限照職業算。
func TestManaByClass(t *testing.T) {
	cases := []struct {
		class byte
		intel int
		want  int
	}{
		{'A', 20, 20}, // 聖者:全額
		{'M', 25, 25}, // 法師:全額
		{'B', 21, 10}, // 吟遊詩人:一半(整數除法)
		{'F', 30, 0},  // 戰士:沒有
	}
	for _, c := range cases {
		if got := ManaForClass(c.class, c.intel); got != c.want {
			t.Errorf("職業 %c 智力 %d 的魔力是 %d,預期 %d",
				c.class, c.intel, got, c.want)
		}
	}
}

// TestLevelUpRaisesHPAndOneStat:升級把生命設成 等級×30,並且**只**加一項三圍。
func TestLevelUpRaisesHPAndOneStat(t *testing.T) {
	s := magicState(t)
	s.SeedRandom(3)
	ch := &s.Roster[0]
	ch.Level, ch.Exp = 1, 800 // → 第 5 級
	ch.Class = 'A'
	before := [3]byte{ch.Strength, ch.Dex, ch.Intel}

	if !s.levelUp(0) {
		t.Fatalf("該升級卻沒升:\n%s", s.log())
	}
	if ch.Level != 5 {
		t.Errorf("升到第 %d 級,預期第 5 級", ch.Level)
	}
	if int(ch.MaxHP) != 5*HPPerLevel || ch.HP != ch.MaxHP {
		t.Errorf("生命 %d/%d,預期 %d 補滿", ch.HP, ch.MaxHP, 5*HPPerLevel)
	}
	after := [3]byte{ch.Strength, ch.Dex, ch.Intel}
	raised := 0
	for i := range before {
		switch {
		case after[i] == before[i]+1:
			raised++
		case after[i] != before[i]:
			t.Errorf("第 %d 項三圍 %d → %d,一次只該 +1", i, before[i], after[i])
		}
	}
	if raised != 1 {
		t.Errorf("有 %d 項三圍上升,一次只該有一項", raised)
	}
	// 聖者的魔力 = 智力。
	if int(ch.MP) != int(ch.Intel) {
		t.Errorf("聖者的魔力是 %d,預期等於智力 %d", ch.MP, ch.Intel)
	}
}

// TestLevelUpIsIdempotent:等級已經對了就什麼都不做。
//
// 原版 `cmp edx, ebx; jz` 直接跳過,連魔力都不重算 —— 所以反覆遇到老人
// 不會一直長三圍。這條擋的正是「每次紮營都 +1 力量」那種走鐘。
func TestLevelUpIsIdempotent(t *testing.T) {
	s := magicState(t)
	ch := &s.Roster[0]
	ch.Exp, ch.Class = 800, 'A'
	s.levelUp(0)
	snap := *ch
	for i := 0; i < 20; i++ {
		if s.levelUp(0) {
			t.Fatal("等級沒變卻回報升級了")
		}
	}
	if ch.Strength != snap.Strength || ch.Dex != snap.Dex || ch.Intel != snap.Intel {
		t.Errorf("反覆升級把三圍推高了:%d/%d/%d → %d/%d/%d",
			snap.Strength, snap.Dex, snap.Intel, ch.Strength, ch.Dex, ch.Intel)
	}
}

// TestStatCap:三圍不會超過 30。
func TestStatCap(t *testing.T) {
	s := magicState(t)
	ch := &s.Roster[0]
	ch.Strength, ch.Dex, ch.Intel = StatCap, StatCap, StatCap
	ch.Class = 'A'
	for lv := 2; lv <= MaxCharacterLevel; lv++ {
		ch.Exp = 100 << uint(lv-2)
		s.levelUp(0)
	}
	if ch.Strength > StatCap || ch.Dex > StatCap || ch.Intel > StatCap {
		t.Errorf("三圍超過上限:%d/%d/%d", ch.Strength, ch.Dex, ch.Intel)
	}
}

// TestSleepCanBringTheApparition:睡一晚有機會遇到老人並升級。
func TestSleepCanBringTheApparition(t *testing.T) {
	seen := false
	for seed := int64(1); seed <= 30 && !seen; seed++ {
		s := magicState(t)
		s.SeedRandom(seed)
		for i := 0; i < s.PartySize; i++ {
			s.Roster[i].Exp = 6400 // 該是第 8 級了
			s.Roster[i].Level = 1
		}
		s.SleepUntilMorning()
		if s.Roster[0].Level == MaxCharacterLevel {
			seen = true
		}
		// 中毒的人睡覺會死,這裡不該有人是那種狀態。
		if s.Roster[0].Status == u5data.StatusDead {
			t.Fatal("睡一覺就死了?")
		}
	}
	if !seen {
		t.Error("30 個亂數種子睡下去都沒遇到老人 —— 1/4 的機率不該這樣")
	}
}
