package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

func combatState(t *testing.T) *State {
	t.Helper()
	s := worldState(t)
	cm, err := u5data.LoadCombatMaps("../../gamedata/BRIT.CBT")
	if err != nil {
		t.Skipf("讀不到 BRIT.CBT:%v", err)
	}
	ct, err := u5data.LoadCreatureTable("../../gamedata")
	if err != nil {
		t.Fatal(err)
	}
	s.CombatMaps, s.Creatures = cm, ct
	return s
}

// TestSelectCombatMap 鎖住地形 → 戰鬥地圖的對照(原版 sub_2E58C 的 73-case 跳表)。
func TestSelectCombatMap(t *testing.T) {
	const walk = u5data.VehicleWalk
	const ship = u5data.VehicleShip
	for _, c := range []struct {
		name      string
		kind      int
		terrain   int
		transport byte
		inWorld   bool
		want      int
	}{
		{"權杖", u5data.EnemySceptre, 5, walk, true, u5data.CombatMapSceptre},
		{"在船上遇敵船", u5data.EnemyShip, 1, ship, true, u5data.CombatMapShipShip},
		{"在船上打水戰", 0x40, 1, ship, true, u5data.CombatMapShipSea},
		{"在船上打陸戰", 0x40, 5, ship, true, u5data.CombatMapShipLnd},
		{"陸上遇敵船", u5data.EnemyShip, 5, walk, true, u5data.CombatMapShipVs},
		{"水面", 0x40, 2, walk, true, u5data.CombatMapOpenSea},
		{"水生怪物在陸上也打水戰", 0x80, 5, walk, true, u5data.CombatMapOpenSea},
		{"地形 4", 0x40, 4, walk, true, 1},
		{"地形 5", 0x40, 5, walk, true, 2},
		{"地形 6", 0x40, 6, walk, true, 3},
		{"地形 8", 0x40, 8, walk, true, 3},
		{"地形 7", 0x40, 7, walk, true, 4},
		{"地形 31", 0x40, 31, walk, true, 4},
		{"地形 10", 0x40, 10, walk, true, 5},
		{"地形 13", 0x40, 13, walk, true, 6},
		{"地形 29", 0x40, 29, walk, true, 7},
		{"地形 0x6B", 0x40, 0x6B, walk, true, 7},
		{"地形 68", 0x40, 68, walk, true, 8},
		{"沒對照 + 在大地圖", 0x40, 40, walk, true, 2},
		{"沒對照 + 在場景", 0x40, 40, walk, false, 8},
	} {
		got := u5data.SelectCombatMap(c.kind, c.terrain, c.transport, c.inWorld)
		if got != c.want {
			t.Errorf("%s:選了第 %d 張,預期第 %d 張", c.name, got, c.want)
		}
	}
}

// TestWaterBattleException:0x6A / 0x6B 是 0x60 那一族裡的兩個例外,不算水戰。
func TestWaterBattleException(t *testing.T) {
	for _, tile := range []int{0, 1, 2, 3, 0x60, 0x65, 0x6F} {
		if !u5data.IsWaterBattle(tile) {
			t.Errorf("tile 0x%02X 應該算水戰", tile)
		}
	}
	for _, tile := range []int{4, 5, 0x6A, 0x6B, 0x70} {
		if u5data.IsWaterBattle(tile) {
			t.Errorf("tile 0x%02X 不該算水戰", tile)
		}
	}
}

// TestBumpingMonsterStartsCombat:走進怪物就開打,而且隊員與敵人各就各位。
func TestBumpingMonsterStartsCombat(t *testing.T) {
	s := combatState(t)
	// 在旁邊放一隻怪物(種類碼 ≥ 0x40 才算怪物)。
	if _, ok := s.CurrentObjects().Spawn(0x40, s.X+1, s.Y, s.Floor); !ok {
		t.Fatal("放不下怪物")
	}
	s.Move(East)
	if !s.InCombat() {
		t.Fatalf("撞上怪物卻沒開打:\n%s", s.log())
	}
	if s.Prompt != PromptCombat {
		t.Errorf("戰鬥中 Prompt 是 %v", s.Prompt)
	}
	c := s.Combat
	if len(c.Units) != s.PartySize+1 {
		t.Errorf("場上 %d 個單位,預期 %d 名隊員 + 1 隻敵人", len(c.Units), s.PartySize)
	}
	// 位置要與圖裡的入場點一致。
	m := c.Map
	for i := 0; i < s.PartySize; i++ {
		u := c.Units[i]
		if !u.IsParty() {
			t.Fatalf("第 %d 個單位不是隊員", i)
		}
		if u.X != int(m.PartyX[i]) || u.Y != int(m.PartyY[i]) {
			t.Errorf("隊員 %d 在 (%d,%d),圖裡的入場點是 (%d,%d)",
				i, u.X, u.Y, m.PartyX[i], m.PartyY[i])
		}
	}
	e := c.Units[len(c.Units)-1]
	if e.IsParty() {
		t.Error("最後一個單位應該是敵人")
	}
	if e.X != int(m.EnemyX[0]) || e.Y != int(m.EnemyY[0]) {
		t.Errorf("敵人在 (%d,%d),圖裡的入場點是 (%d,%d)", e.X, e.Y, m.EnemyX[0], m.EnemyY[0])
	}
	// 敵人名字要查得出來(種類 0x40 = Mage)。
	if c.EnemyName != "Mage" {
		t.Errorf("敵人名字是 %q,預期 Mage", c.EnemyName)
	}
}

// TestPirateNaming:種類碼 < 0x40 的敵人一律叫 PIRATES(原版 sub_2E58C)。
func TestPirateNaming(t *testing.T) {
	s := combatState(t)
	if got := s.enemyDisplayName(u5data.EnemyShip); got != "PIRATES" {
		t.Errorf("敵船的名字是 %q,預期 PIRATES", got)
	}
	if got := s.enemyDisplayName(0x40); got != "Mage" {
		t.Errorf("種類 0x40 的名字是 %q,預期 Mage", got)
	}
}

// TestCombatMoveStaysOnField:戰場上的移動不能走出 11×11,也不能疊在別人身上。
func TestCombatMoveStaysOnField(t *testing.T) {
	s := combatState(t)
	s.CurrentObjects().Spawn(0x40, s.X+1, s.Y, s.Floor)
	s.Move(East)
	if !s.InCombat() {
		t.Skip("沒開打")
	}
	c := s.Combat
	for turn := 0; turn < 200; turn++ {
		for _, d := range []Direction{North, East, South, West} {
			s.CombatMove(d)
		}
		for i := range c.Units {
			u := &c.Units[i]
			if u.X < 0 || u.X >= u5data.CombatSide || u.Y < 0 || u.Y >= u5data.CombatSide {
				t.Fatalf("單位 %d 走到 (%d,%d),出了戰場", i, u.X, u.Y)
			}
		}
		seen := map[[2]int]int{}
		for i := range c.Units {
			u := &c.Units[i]
			k := [2]int{u.X, u.Y}
			if prev, dup := seen[k]; dup {
				t.Fatalf("單位 %d 與 %d 疊在 (%d,%d)", prev, i, u.X, u.Y)
			}
			seen[k] = i
		}
	}
}

// TestFleeLeavesCombat:撤離回到地圖,而且怪物還在(沒打贏就不該消失)。
func TestFleeLeavesCombat(t *testing.T) {
	s := combatState(t)
	s.CurrentObjects().Spawn(0x40, s.X+1, s.Y, s.Floor)
	s.Move(East)
	if !s.InCombat() {
		t.Skip("沒開打")
	}
	s.CombatFlee()
	if s.InCombat() {
		t.Error("撤離之後還在戰鬥")
	}
	if s.Prompt != PromptNone {
		t.Errorf("撤離之後 Prompt 是 %v", s.Prompt)
	}
	if _, _, ok := s.ObjectAt(s.X+1, s.Y); !ok {
		t.Error("沒打贏,怪物卻從地圖上消失了")
	}
	if !strings.Contains(s.log(), "撤離") {
		t.Errorf("沒有撤離訊息:\n%s", s.log())
	}
}
