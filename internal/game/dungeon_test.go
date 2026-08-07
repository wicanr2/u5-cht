package game

import (
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

func dungeonState(t *testing.T) *State {
	t.Helper()
	s := magicState(t)
	dg, err := u5data.LoadDungeons("../../gamedata")
	if err != nil {
		t.Fatal(err)
	}
	dr, err := u5data.LoadDungeonRooms("../../gamedata")
	if err != nil {
		t.Fatal(err)
	}
	s.Dungeons, s.DungeonRooms = dg, dr
	return s
}

// TestEnterDungeonFromWorldMap:站在入口按 E 就進得去。
func TestEnterDungeonFromWorldMap(t *testing.T) {
	s := dungeonState(t)
	s.Location, s.Floor = 0, 0
	e := u5data.DungeonEntrances[0] // Deceit
	s.X, s.Y = e.X, e.Y
	s.Enter()
	if !s.InDungeon() {
		t.Fatalf("站在 %s 入口卻進不去:\n%s", e.Name, s.log())
	}
	d := s.Dungeon
	if d.Index != 0 || d.Level != 0 || d.X != DungeonEntryX || d.Y != DungeonEntryY {
		t.Errorf("落點是第 %d 座第 %d 層 (%d,%d),預期第 0 座第 0 層 (1,1)",
			d.Index, d.Level, d.X, d.Y)
	}
	// 地點編號要落在咒語場合表認得的地牢範圍(0x21..0x7F)。
	if s.Location < 0x21 || s.Location > 0x7F {
		t.Errorf("地牢裡的地點編號是 %d,不在 0x21..0x7F", s.Location)
	}
}

// TestDungeonExitByClimbingUp:第 1 層入口的梯子爬上去就出來。
func TestDungeonExitByClimbingUp(t *testing.T) {
	s := dungeonState(t)
	s.Location = 0
	e := u5data.DungeonEntrances[0]
	s.X, s.Y = e.X, e.Y
	s.Enter()
	if !s.InDungeon() {
		t.Fatal("進不去")
	}
	s.DungeonKlimb(true)
	if s.InDungeon() {
		t.Errorf("在入口的梯子往上爬卻沒出來:\n%s", s.log())
	}
	if s.Location != 0 {
		t.Errorf("出來之後地點是 %d,預期 0(大地圖)", s.Location)
	}
}

// TestDungeonTurningIsFirstPerson:轉向四次回到原點,而且不移動。
func TestDungeonTurningIsFirstPerson(t *testing.T) {
	s := dungeonState(t)
	if !s.EnterDungeon(0, false) {
		t.Fatal("進不去")
	}
	d := s.Dungeon
	start, x, y := d.Facing, d.X, d.Y
	for i := 0; i < 4; i++ {
		s.DungeonTurn(false)
	}
	if d.Facing != start {
		t.Errorf("右轉四次之後朝向 %v,預期回到 %v", d.Facing, start)
	}
	for i := 0; i < 4; i++ {
		s.DungeonTurn(true)
	}
	if d.Facing != start {
		t.Errorf("左轉四次之後朝向 %v,預期回到 %v", d.Facing, start)
	}
	if d.X != x || d.Y != y {
		t.Errorf("轉向卻移動了:(%d,%d) → (%d,%d)", x, y, d.X, d.Y)
	}
}

// TestDungeonWalkStaysInsideAndOffWalls:走再多步也不會出界或站進牆裡。
//
// 這是地牢移動唯一的決定性驗收 —— 通行判定寫反(把通道當牆)會讓玩家
// 一步都走不動,寫漏會讓人穿牆,兩種都被這條抓到。
func TestDungeonWalkStaysInsideAndOffWalls(t *testing.T) {
	s := dungeonState(t)
	s.SeedRandom(5)
	for dg := 0; dg < u5data.DungeonCount; dg++ {
		if !s.EnterDungeon(dg, false) {
			t.Fatalf("第 %d 座進不去", dg)
		}
		d := s.Dungeon
		s.Combat = nil // 入口就是房間的那幾座(Shame / Doom)先把戰鬥收掉
		moved := 0
		for step := 0; step < 400; step++ {
			s.Combat = nil // 走進房間會開打;這條測的是走路,打完就繼續
			if !s.InDungeon() {
				break
			}
			before := [2]int{d.X, d.Y}
			s.DungeonMove(Direction(s.Roll(0, 3)))
			if !s.InDungeon() {
				break
			}
			if d.X < 0 || d.X >= u5data.DungeonSide || d.Y < 0 || d.Y >= u5data.DungeonSide {
				t.Fatalf("第 %d 座走到 (%d,%d),出了 8×8", dg, d.X, d.Y)
			}
			if u5data.DungeonPlayerBlocks(s.DungeonTileHere()) {
				t.Fatalf("第 %d 座站進了 %02X(走不過去的格子)", dg, s.DungeonTileHere())
			}
			if [2]int{d.X, d.Y} != before {
				moved++
			}
		}
		// ⚠ 「一步都沒動」不一定是 bug:Hythloth 的入口 (1,1) 是一座
		// **四面都是牆**的梯子井,只能上下不能走。所以要嘛動過,
		// 要嘛四鄰真的都堵死 —— 兩者皆非才是通行判定寫反了。
		if moved == 0 {
			open := 0
			for _, dxy := range [][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}} {
				if !u5data.DungeonPlayerBlocks(s.DungeonTileAt(d.X+dxy[0], d.Y+dxy[1])) {
					open++
				}
			}
			if open > 0 {
				t.Errorf("第 %d 座 400 步一格都沒動,但 (%d,%d) 有 %d 個方向是通的"+
					" —— 通行判定大概反了", dg, d.X, d.Y, open)
			}
		}
		s.Combat = nil
		s.Dungeon = nil
	}
}

// TestDungeonPitTrapFallsOneLevel:掉進坑會下一層,而且落點多出「頭上的洞」。
func TestDungeonPitTrapFallsOneLevel(t *testing.T) {
	s := dungeonState(t)
	if !s.EnterDungeon(0, false) {
		t.Fatal("進不去")
	}
	d := s.Dungeon
	// 找一個陷阱坑,把玩家放上去。
	found := false
	for l := 0; l < u5data.DungeonLevels-1 && !found; l++ {
		for y := 0; y < u5data.DungeonSide && !found; y++ {
			for x := 0; x < u5data.DungeonSide && !found; x++ {
				tile := s.Dungeons.At(d.Index, l, x, y)
				if tile != u5data.DungeonPitTrapA && tile != u5data.DungeonPitTrapB {
					continue
				}
				d.Level, d.X, d.Y = l, x, y
				found = true
			}
		}
	}
	if !found {
		t.Skip("這座地牢沒有陷阱坑")
	}
	level, x, y := d.Level, d.X, d.Y
	s.onDungeonTile()
	if d.Level != level+1 {
		t.Errorf("踩到陷阱坑卻還在第 %d 層,預期掉到第 %d 層", d.Level, level+1)
	}
	// 原本那一格的陷阱要被用掉。
	if before := s.Dungeons.At(d.Index, level, x, y); before&0x07 != 0 {
		t.Errorf("陷阱格還是 %02X,低三位元該被清掉", before)
	}
	// 落點要能爬回去(頭上有洞)。
	if s.Dungeons.At(d.Index, d.Level, x, y)&u5data.DungeonHoleAbove == 0 {
		t.Error("落點沒有標上「頭上有洞」,之後爬不回去")
	}
}

// TestDungeonRoomStartsCombat:走進房間格就開一場戰鬥,怪物來自房間資料。
//
// ⚠ 入口房不一定有怪 —— Shame 的第 0 號房整排 EnemyKind 都是 0,那是原版
// 資料就這樣(一間空房)。所以分兩件事驗:**開得起來**用入口房,
// **怪物讀得到**掃全部 112 間房。
func TestDungeonRoomStartsCombat(t *testing.T) {
	s := dungeonState(t)
	// Shame(索引 5)的入口 (1,1) 本身就是房間。
	if !s.EnterDungeon(5, false) {
		t.Fatal("進不去")
	}
	if !s.InCombat() {
		t.Fatalf("Shame 的入口是房間卻沒開打:\n%s", s.log())
	}
	if _, party := s.sideCounts(s.Combat); party == 0 {
		t.Error("隊伍沒進場")
	}

	// 全部房間裡至少要有一大半排得出怪物。
	withMonsters := 0
	for i := range s.DungeonRooms.Maps {
		s.Combat = nil
		s.beginRoomCombat(&s.DungeonRooms.Maps[i], i)
		if e, _ := s.sideCounts(s.Combat); e > 0 {
			withMonsters++
		}
	}
	if withMonsters < len(s.DungeonRooms.Maps)/2 {
		t.Errorf("%d 間房裡只有 %d 間排得出怪物", len(s.DungeonRooms.Maps), withMonsters)
	}
	t.Logf("%d 間房、%d 間有怪物", len(s.DungeonRooms.Maps), withMonsters)
}

// TestDungeonSpellsWorkOnlyInDungeons:Uus Por / Des Por 在地牢裡才施得動。
func TestDungeonSpellsWorkOnlyInDungeons(t *testing.T) {
	s := dungeonState(t)
	uus := s.Spells.Find("Uus Por")
	des := s.Spells.Find("Des Por")
	s.Roster[0].Level, s.Roster[0].MP = 8, 40

	// 大地圖上施不動。
	s.Location = 0
	s.Inventory.Spells[uus] = 1
	if got := s.Cast(0, uus); got != MagicNotHere {
		t.Errorf("大地圖上施 Uus Por 回傳 %v,預期 MagicNotHere", got)
	}

	// 地牢裡:找一個上下都是純通道的位置,Des Por 要下得去。
	if !s.EnterDungeon(0, false) {
		t.Fatal("進不去")
	}
	s.Combat = nil
	d := s.Dungeon
	for l := 0; l < u5data.DungeonLevels-1; l++ {
		for y := 0; y < u5data.DungeonSide; y++ {
			for x := 0; x < u5data.DungeonSide; x++ {
				if u5data.DungeonKind(s.Dungeons.At(d.Index, l, x, y)) != u5data.DungeonPassage ||
					u5data.DungeonKind(s.Dungeons.At(d.Index, l+1, x, y)) != u5data.DungeonPassage {
					continue
				}
				d.Level, d.X, d.Y = l, x, y
				s.Roster[0].MP = 40
				s.Inventory.Spells[des] = 1
				if got := s.Cast(0, des); got != MagicSuccess {
					t.Fatalf("在 (%d,%d) 第 %d 層施 Des Por 回傳 %v:\n%s",
						x, y, l, got, s.log())
				}
				if d.Level != l+1 {
					t.Errorf("Des Por 之後在第 %d 層,預期第 %d 層", d.Level, l+1)
				}
				return
			}
		}
	}
	t.Skip("找不到上下都是純通道的位置")
}
