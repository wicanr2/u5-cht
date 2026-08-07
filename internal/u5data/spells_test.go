package u5data

import (
	"os"
	"testing"
)

func loadSpellTable(t *testing.T) *SpellTable {
	t.Helper()
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	s, err := LoadSpells(dir)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// TestSpellCirclesAreSixEach:8 圈 × 6 個,而且圈界要落在對的咒語上。
//
// 圈數是 `索引 / 6 + 1` 算出來的,所以這條同時驗「名字表沒偏」——
// 只要少讀或多讀一個字串,下面每個圈界都會錯開。
func TestSpellCirclesAreSixEach(t *testing.T) {
	s := loadSpellTable(t)
	bounds := map[string]int{
		"In Lor":     1, // 第 1 圈第一個
		"An Ylem":    1, // 第 1 圈最後一個
		"An Sanct":   2, // 第 2 圈第一個
		"Vas Lor":    3,
		"In Por":     3, // 第 3 圈最後一個
		"An Grav":    4,
		"In Bet Xen": 5,
		"In Vas P Y": 6,
		"Sanct Lo":   7,
		"In Mani Co": 8,
		"An Tym":     8, // 第 8 圈最後一個
	}
	for name, circle := range bounds {
		i := s.Find(name)
		if i < 0 {
			t.Errorf("找不到咒語 %q", name)
			continue
		}
		if s.Spells[i].Circle != circle {
			t.Errorf("%s(索引 %d)是第 %d 圈,預期第 %d 圈",
				name, i, s.Spells[i].Circle, circle)
		}
	}
	// 每一圈都要正好 6 個。
	count := map[int]int{}
	for i := range s.Spells {
		count[s.Spells[i].Circle]++
	}
	for c := 1; c <= 8; c++ {
		if count[c] != SpellsPerCircle {
			t.Errorf("第 %d 圈有 %d 個咒語,預期 %d", c, count[c], SpellsPerCircle)
		}
	}
}

// TestSpellReagentsMatchCanon:幾個大家都背得出來的配方要對。
//
// 挑的這幾個橫跨全部 8 個位元,而且分佈在不同圈 —— 表偏一格就會一起錯。
func TestSpellReagentsMatchCanon(t *testing.T) {
	s := loadSpellTable(t)
	canon := map[string][]int{
		"In Lor":     {ReagentSulfurousAsh},
		"Grav Por":   {ReagentSulfurousAsh, ReagentBlackPearl},
		"An Nox":     {ReagentGinseng, ReagentGarlic},
		"Mani":       {ReagentGinseng, ReagentSpiderSilk},
		"In Wis":     {ReagentNightshade},
		"Kal Xen":    {ReagentSpiderSilk, ReagentMandrakeRoot},
		"Rel Hur":    {ReagentSulfurousAsh, ReagentBloodMoss},
		"In Mani Co": {ReagentSulfurousAsh, ReagentGinseng, ReagentGarlic, ReagentSpiderSilk, ReagentBloodMoss, ReagentMandrakeRoot},
	}
	for name, want := range canon {
		i := s.Find(name)
		if i < 0 {
			t.Errorf("找不到咒語 %q", name)
			continue
		}
		got := s.Spells[i].ReagentList()
		if len(got) != len(want) {
			t.Errorf("%s 的藥草是 %v,預期 %v", name, got, want)
			continue
		}
		for k := range got {
			if got[k] != want[k] {
				t.Errorf("%s 的藥草是 %v,預期 %v", name, got, want)
				break
			}
		}
	}
	// 每個咒語都要至少一種藥草 —— 一個都不用的話配方表大概讀到空白區了。
	for i := range s.Spells {
		if s.Spells[i].Reagents == 0 {
			t.Errorf("%s 不需要任何藥草?", s.Spells[i].Name)
		}
	}
}

// TestSpellContextsMatchWhereTheyMakeSense:能在哪裡施要說得通。
func TestSpellContextsMatchWhereTheyMakeSense(t *testing.T) {
	s := loadSpellTable(t)
	at := func(name string) Spell {
		i := s.Find(name)
		if i < 0 {
			t.Fatalf("找不到咒語 %q", name)
		}
		return s.Spells[i]
	}
	// 改風向只在大海上有意義 → 只有野外。
	if got := at("Rel Hur"); got.Context != CastOutdoors {
		t.Errorf("Rel Hur 的場合是 %04b,預期只有野外", got.Context)
	}
	// 上下樓只在地牢。
	for _, n := range []string{"Uus Por", "Des Por"} {
		if got := at(n); got.Context != CastInDungeon {
			t.Errorf("%s 的場合是 %04b,預期只有地牢", n, got.Context)
		}
	}
	// 攻擊咒語只在戰鬥。
	for _, n := range []string{"Grav Por", "Vas Flam", "Xen Corp", "Kal Xen"} {
		if got := at(n); got.Context != CastInCombat {
			t.Errorf("%s 的場合是 %04b,預期只有戰鬥", n, got.Context)
		}
	}
	// 治療與解毒到處都能用。
	for _, n := range []string{"Mani", "An Nox", "An Zu", "Vas Mani", "An Tym"} {
		g := at(n)
		if g.Context != CastInCombat|CastInDungeon|CastInTown|CastOutdoors {
			t.Errorf("%s 的場合是 %04b,預期四處皆可", n, g.Context)
		}
	}
	// 復活與大傳送門**不能在戰鬥中用**。
	for _, n := range []string{"In Mani Co", "Vas Rel Po"} {
		if g := at(n); g.Context&CastInCombat != 0 {
			t.Errorf("%s 不該能在戰鬥中施(%04b)", n, g.Context)
		}
	}
}

// TestCanCastAtMapsLocationRanges:地點值 → 場合位元的對照。
func TestCanCastAtMapsLocationRanges(t *testing.T) {
	only := func(bit byte) Spell { return Spell{Context: bit} }
	cases := []struct {
		loc  int
		want byte
	}{
		{0, CastOutdoors},           // 大地圖
		{1, CastInTown},             // 月光城
		{0x20, CastInTown},          // 城鎮範圍的上界
		{0x21, CastInDungeon},       // 地牢
		{0x7F, CastInDungeon},       // 地牢範圍的上界
		{0x80, CastInCombat},        // 原版用 > 0x7F 表示戰鬥
		{CombatLocation, CastInCombat}, // −1 也是戰鬥
	}
	for _, c := range cases {
		for _, bit := range []byte{CastInCombat, CastInDungeon, CastInTown, CastOutdoors} {
			got := only(bit).CanCastAt(c.loc)
			if want := bit == c.want; got != want {
				t.Errorf("地點 %d、位元 %04b:CanCastAt = %v,預期 %v",
					c.loc, bit, got, want)
			}
		}
	}
}
