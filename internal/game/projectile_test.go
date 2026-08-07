package game

import (
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// TestProjectilePathIsAStraightLine:路徑要從起點旁邊出發、正好停在終點。
func TestProjectilePathIsAStraightLine(t *testing.T) {
	cases := []struct{ x0, y0, x1, y1, steps int }{
		{5, 5, 5, 5, 0},  // 原地
		{0, 0, 3, 0, 3},  // 水平
		{0, 0, 0, 4, 4},  // 垂直
		{0, 0, 3, 3, 3},  // 正斜角
		{0, 0, 6, 3, 6},  // 淺斜角:步數 = 長邊
		{10, 10, 0, 0, 10},
	}
	for _, c := range cases {
		p := projectilePath(c.x0, c.y0, c.x1, c.y1)
		if len(p) != c.steps {
			t.Errorf("(%d,%d)→(%d,%d) 走了 %d 步,預期 %d",
				c.x0, c.y0, c.x1, c.y1, len(p), c.steps)
			continue
		}
		if c.steps == 0 {
			continue
		}
		// 不含起點。
		if p[0] == [2]int{c.x0, c.y0} {
			t.Errorf("(%d,%d)→(%d,%d) 的第一步就是起點", c.x0, c.y0, c.x1, c.y1)
		}
		// 正好停在終點。
		if last := p[len(p)-1]; last != [2]int{c.x1, c.y1} {
			t.Errorf("(%d,%d)→(%d,%d) 停在 (%d,%d)", c.x0, c.y0, c.x1, c.y1, last[0], last[1])
		}
		// 每一步只能挪一格(含斜角)。
		prev := [2]int{c.x0, c.y0}
		for _, q := range p {
			if iabs(q[0]-prev[0]) > 1 || iabs(q[1]-prev[1]) > 1 {
				t.Errorf("從 (%d,%d) 跳到 (%d,%d),一步不該超過一格",
					prev[0], prev[1], q[0], q[1])
			}
			prev = q
		}
	}
}

// TestProjectileTablePolarity:bit = 1 是「穿得過」,與行走那兩張相反。
//
// 抄錯極性的話箭會變成只能穿牆、不能穿空地 —— 而畫面上只看得到「射不到」。
// 用水與山兩端一起驗:水擋行走但不擋箭,山兩者都擋。
func TestProjectileTablePolarity(t *testing.T) {
	// ⚠ 三種深度的水是 tile **1..3**;tile 0 在行走表裡是通的
	//(那一格不是水,是空白 / 未定義),所以不能一起算進來。
	for tile := 1; tile < 4; tile++ {
		if u5data.TileBlocksProjectile(tile) {
			t.Errorf("水 tile %d 不該擋箭", tile)
		}
		if !u5data.TileBlocksWalking(tile) {
			t.Errorf("水 tile %d 該擋行走", tile)
		}
	}
	// 擋箭的一定不多於擋行走的,而且兩者的差集就是那七個載具 / 坐騎。
	both, projOnly := 0, 0
	for tile := 0; tile < 256; tile++ {
		if !u5data.TileBlocksProjectile(tile) {
			continue
		}
		if u5data.TileBlocksWalking(tile) {
			both++
		} else {
			projOnly++
		}
	}
	if both == 0 {
		t.Fatal("沒有任何 tile 同時擋箭與擋行走 —— 極性大概反了")
	}
	if projOnly > 10 {
		t.Errorf("有 %d 個 tile 擋箭卻不擋行走,預期只有少數載具 / 坐騎", projOnly)
	}
	t.Logf("擋箭的 %d 個,其中 %d 個同時擋行走、%d 個只擋箭", both+projOnly, both, projOnly)
}

// TestProjectileStopsAtFirstUnit:射線上的人會替目標擋下來 —— 包括自己人。
//
// 這是原版行為(`sub_1FE54` 回傳實際打到的那一個),不是我加的難度。
func TestProjectileStopsAtFirstUnit(t *testing.T) {
	s := magicState(t)
	slot, _ := s.CurrentObjects().Spawn(0x40, s.X+1, s.Y, s.Floor)
	if !s.BeginCombat(slot) {
		t.Fatal("打不起來")
	}
	c := s.Combat
	// 排一條乾淨的橫線:射手在 (1,5)、擋路的在 (3,5)、目標在 (6,5)。
	shooter, blocker, target := 0, 1, u5data.CombatPartySlots
	place := func(i, x, y int) {
		c.Units[i].X, c.Units[i].Y = x, y
	}
	// 先把其他人挪開。
	for i := range c.Units {
		if c.Units[i].Active() && i != shooter && i != blocker && i != target {
			c.Units[i].Flags |= UnitDead
		}
	}
	place(shooter, 1, 5)
	place(blocker, 3, 5)
	place(target, 6, 5)
	// 那一整列要是箭飛得過的地形,否則測的就不是「被人擋住」。
	for x := 2; x <= 6; x++ {
		if u5data.TileBlocksProjectile(int(c.Map.At(x, 5))) {
			t.Skipf("第 5 列的 (%d,5) 是擋箭的地形,換一張圖再測", x)
		}
	}
	got, _, _ := s.FlyProjectile(shooter, 6, 5)
	if got != blocker {
		t.Errorf("箭打到第 %d 槽,預期被擋路的第 %d 槽吃掉", got, blocker)
	}
	// 把擋路的移開就打得到目標。
	c.Units[blocker].Flags |= UnitDead
	if got, _, _ := s.FlyProjectile(shooter, 6, 5); got != target {
		t.Errorf("擋路的走開之後打到第 %d 槽,預期第 %d 槽", got, target)
	}
}

// TestProjectileStopsAtBlockingTerrain:撞到擋箭的地形就停,不會穿過去。
func TestProjectileStopsAtBlockingTerrain(t *testing.T) {
	s := magicState(t)
	slot, _ := s.CurrentObjects().Spawn(0x40, s.X+1, s.Y, s.Floor)
	if !s.BeginCombat(slot) {
		t.Fatal("打不起來")
	}
	c := s.Combat
	// 找一張列上有擋箭地形的圖。
	for idx := range s.CombatMaps.Maps {
		m := &s.CombatMaps.Maps[idx]
		for y := 0; y < u5data.CombatSide; y++ {
			wall := -1
			for x := 1; x < u5data.CombatSide-1; x++ {
				if u5data.TileBlocksProjectile(int(m.At(x, y))) {
					wall = x
					break
				}
			}
			if wall < 2 {
				continue
			}
			c.Map = m
			for i := range c.Units {
				c.Units[i].Flags = 0
			}
			c.Units[0] = Combatant{Roster: 0, Creature: -1, Flags: UnitParty, X: 0, Y: y}
			c.Units[u5data.CombatPartySlots] = Combatant{
				Roster: -1, Creature: 0, Flags: UnitMonster, HP: 10,
				X: u5data.CombatSide - 1, Y: y,
			}
			hit, ex, _ := s.FlyProjectile(0, u5data.CombatSide-1, y)
			if hit >= 0 {
				t.Errorf("第 %d 列 (%d,%d) 有牆,箭卻打到第 %d 槽", y, wall, y, hit)
			}
			if ex >= wall {
				t.Errorf("箭停在 x=%d,牆在 x=%d —— 穿過去了", ex, wall)
			}
			return
		}
	}
	t.Skip("找不到列上有擋箭地形的戰鬥地圖")
}

// TestRangedWeaponFiresAcrossTheField:拿弓的隊員打得到隔幾格的敵人。
func TestRangedWeaponFiresAcrossTheField(t *testing.T) {
	s := magicState(t)
	slot, _ := s.CurrentObjects().Spawn(0x40, s.X+1, s.Y, s.Floor)
	if !s.BeginCombat(slot) {
		t.Fatal("打不起來")
	}
	c := s.Combat
	for i := range c.Units {
		if c.Units[i].Active() && i != 0 && i != u5data.CombatPartySlots {
			c.Units[i].Flags |= UnitDead
		}
	}
	// 找一列箭飛得過的。
	row := -1
	for y := 0; y < u5data.CombatSide && row < 0; y++ {
		clear := true
		for x := 0; x < 5; x++ {
			if u5data.TileBlocksProjectile(int(c.Map.At(x, y))) {
				clear = false
			}
		}
		if clear {
			row = y
		}
	}
	if row < 0 {
		t.Skip("找不到乾淨的一列")
	}
	c.Units[0].X, c.Units[0].Y = 0, row
	foe := &c.Units[u5data.CombatPartySlots]
	foe.X, foe.Y, foe.HP = 4, row, 40
	// 給隊長一把弓(裝備 26,射程 7)。
	s.Roster[0].Raw[u5data.CharWeapon] = 26
	if s.Stats.ItemRange[26] == 0 {
		t.Skip("裝備 26 沒有射程")
	}
	c.Turn = 0
	before := foe.HP
	for i := 0; i < 30 && foe.HP == before && foe.Active(); i++ {
		c.Turn = 0
		s.CombatAttack(East)
	}
	if foe.HP == before && foe.Active() {
		t.Errorf("拿弓射了 30 次,隔 4 格的敵人一點傷都沒受:\n%s", s.log())
	}
}
