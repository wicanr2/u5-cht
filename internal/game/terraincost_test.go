package game

import "testing"

// TestTerrainCostGrid:分級表逐格對原版的 switch。
//
// ⚠ 草地(5)是這張表最容易寫錯的一格:它**落在 4..15 這個「粗糙」範圍裡,
// 但代價是 0**。原版把它獨立成 switch 的第一條;寫成「4..8 是一級」
// 就會讓平原走起來跟沼澤一樣慢,而那在遊玩時只覺得「怪」,不會有人查得出來。
func TestTerrainCostGrid(t *testing.T) {
	want := map[int]int{
		0: terrainCostNone, // 未使用
		3: terrainCostNone, // 淺灘
		4: terrainCostSlow, // 沼澤
		5: terrainCostNone, // ★ 草地:在範圍內,但不慢
		6: terrainCostSlow, // 灌木
		7: terrainCostSlow, // 荒漠
		8: terrainCostSlow, // 灌木
		9: terrainCostVery, // 樹林
		10: terrainCostVery, // 熱帶林
		11: terrainCostVery, // 丘陵
		12: terrainCostVery, // 山
		13: terrainCostVery, // 高峰
		15: terrainCostVery, // 丘陵
		16: terrainCostNone, // 範圍外:小屋
		0x1E: terrainCostSlow, // 另一組荒漠
		0x1F: terrainCostSlow,
		0x20: terrainCostNone,
		68:   terrainCostNone, // 城鎮地面
	}
	for tile, w := range want {
		if got := TerrainCost(tile); got != w {
			t.Errorf("tile %d(0x%02X)代價 %d,預期 %d", tile, tile, got, w)
		}
	}
}

// TestTerrainCostSpendsTimeAndTurns:代價是「回合 + 時間」兩件事。
//
// 額外的回合是**完整的世界回合**(怪物照走),時間則是固定的 2 / 4 分鐘 ——
// 不是「每個額外回合 1 分鐘」。兩者分開算,混在一起會少算時間。
func TestTerrainCostSpendsTimeAndTurns(t *testing.T) {
	cases := []struct {
		tile    int
		minutes int
	}{
		{5, 0},  // 草地:什麼都不付
		{4, 2},  // 沼澤
		{12, 4}, // 山
	}
	for _, c := range cases {
		s := &State{MaxMessages: 8}
		before := s.Clock
		s.payTerrainCost(c.tile)
		got := (s.Clock.Hour-before.Hour)*MinutesPerHour + (s.Clock.Minute - before.Minute)
		if got != c.minutes {
			t.Errorf("tile %d 走了 %d 分鐘,預期 %d", c.tile, got, c.minutes)
		}
	}
}

// TestTerrainCostAnnounces:一級印「步履維艱」,二級印「寸步難行」。
func TestTerrainCostAnnounces(t *testing.T) {
	cases := []struct {
		tile int
		want string
	}{
		{4, MsgSlowProgress},
		{12, MsgVerySlow},
	}
	for _, c := range cases {
		s := &State{MaxMessages: 8}
		s.payTerrainCost(c.tile)
		found := false
		for _, m := range s.Messages {
			if m == c.want {
				found = true
			}
		}
		if !found {
			t.Errorf("tile %d 沒印出 %q,訊息是 %q", c.tile, c.want, s.Messages)
		}
	}
	// 草地什麼都不印。
	s := &State{MaxMessages: 8}
	s.payTerrainCost(5)
	if len(s.Messages) != 0 {
		t.Errorf("草地不該有訊息,得到 %q", s.Messages)
	}
}
