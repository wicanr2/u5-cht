package game

import (
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// clearObjects 把物件表清空,免得別的槽干擾。
func clearObjects(t *testing.T, s *State) {
	t.Helper()
	set := s.currentObjects()
	if set == nil {
		t.Fatal("沒有物件表")
	}
	for i := range set.Objects {
		set.Objects[i] = u5data.MapObject{}
	}
}

// flatGrass 把隊伍周圍鋪成草地,讓通行判定不受地形干擾。
func flatGrass(t *testing.T, s *State, r int) {
	t.Helper()
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			if !s.SetTileAt(s.X+dx, s.Y+dy, tileGrass) {
				t.Fatal("寫不進世界地圖")
			}
		}
	}
}

// TestChaseMovesTowardTheParty —— `sub_2B24`:往隊伍走一格。
func TestChaseMovesTowardTheParty(t *testing.T) {
	s := overworldScene(t)
	clearObjects(t, s)
	flatGrass(t, s, 12)
	// Orc 放在東南方 8 格處。
	putObject(t, s, 5, 0xC0, s.X+8, s.Y+8)
	before := [2]int{s.X + 8, s.Y + 8}
	s.chaseParty(5)
	o := s.currentObjects().Objects[5]
	nx, ny := int(o.Raw[u5data.ObjX]), int(o.Raw[u5data.ObjY])
	if nx == before[0] && ny == before[1] {
		t.Fatal("怪一步都沒動")
	}
	// 只能動一軸,而且是往隊伍靠的那一邊。
	dx, dy := nx-before[0], ny-before[1]
	if absInt(dx)+absInt(dy) != 1 {
		t.Errorf("走了 (%+d,%+d) —— 原版一次只動一格、一個軸", dx, dy)
	}
	if dx > 0 || dy > 0 {
		t.Errorf("走了 (%+d,%+d) —— 隊伍在西北邊,該往負向走", dx, dy)
	}
}

// TestChaseFallsBackToWanderingWhenBoxedIn —— ★ 兩軸都進不去就改隨機遊走。
//
// 這是原版不會讓怪物卡在牆邊發呆的原因(`sub_2B24` 最後一行 `call sub_2A54`)。
func TestChaseFallsBackToWanderingWhenBoxedIn(t *testing.T) {
	s := overworldScene(t)
	clearObjects(t, s)
	flatGrass(t, s, 12)
	// 怪在正東 8 格;把牠往西的那一格改成深水(Orc 過不去),
	// Y 軸的步是 0(同一列)⇒ 追人的兩條路都走不通。
	ox, oy := s.X+8, s.Y
	putObject(t, s, 5, 0xC0, ox, oy)
	if !s.SetTileAt(ox-1, oy, u5data.RoughSeasTile) {
		t.Fatal("寫不進世界地圖")
	}
	moved := false
	for i := 0; i < 60 && !moved; i++ {
		s.chaseParty(5)
		o := s.currentObjects().Objects[5]
		if int(o.Raw[u5data.ObjX]) != ox || int(o.Raw[u5data.ObjY]) != oy {
			moved = true
		}
	}
	if !moved {
		t.Error("追不到又不會遊走 —— 原版卡住時會退回 `sub_2A54`")
	}
}

// TestFlyersIgnoreTerrainSlowdown —— ★ 會飛的四種免疫地形延遲。
//
// 沼澤/灌木/荒漠 1/2、林/丘/山 1/3 才走得成;龍、蝙蝠、惡魔、Mongbat 免疫。
func TestFlyersIgnoreTerrainSlowdown(t *testing.T) {
	for _, k := range []byte{u5data.FlyerDragon, u5data.FlyerBat, u5data.FlyerDaemon, u5data.FlyerMongbat} {
		if !u5data.CreatureIgnoresTerrain(k) {
			t.Errorf("0x%02X 該免疫地形延遲", k)
		}
	}
	if u5data.CreatureIgnoresTerrain(0xC0) {
		t.Error("Orc 不該免疫 —— 原版只列了四種")
	}

	// 走進山地(2 級):Orc 需要多試幾次,龍一次就過。
	step := func(kind byte) int {
		s := overworldScene(t)
		clearObjects(t, s)
		flatGrass(t, s, 12)
		ox, oy := s.X+4, s.Y
		putObject(t, s, 5, kind, ox, oy)
		if !s.SetTileAt(ox-1, oy, 11) { // 丘陵 = 2 級
			t.Fatal("寫不進世界地圖")
		}
		for i := 1; i <= 60; i++ {
			s.stepObject(5, -1, 0)
			o := s.currentObjects().Objects[5]
			if int(o.Raw[u5data.ObjX]) != ox {
				return i
			}
		}
		return 0
	}
	if got := step(u5data.FlyerDragon); got != 1 {
		t.Errorf("龍走進丘陵花了 %d 次,免疫的話該是 1 次", got)
	}
	if got := step(0xC0); got == 0 {
		t.Error("Orc 60 次都走不進丘陵 —— 1/3 不該這樣")
	}
}

// TestEnemyShipTurnsToFaceItsHeading —— 敵船移動時換朝向圖。
//
// 朝向的低兩位就是**北東南西**(`sub_2CCFC` 的跳表定案,`docs/re/84` §3)。
func TestEnemyShipTurnsToFaceItsHeading(t *testing.T) {
	for _, tc := range []struct {
		dx, dy int
		want   int
		name   string
	}{
		{0, -1, u5data.ShipFacingN, "北"},
		{1, 0, u5data.ShipFacingE, "東"},
		{0, 1, u5data.ShipFacingS, "南"},
		{-1, 0, u5data.ShipFacingW, "西"},
	} {
		s := overworldScene(t)
		clearObjects(t, s)
		flatGrass(t, s, 12)
		// 目標格要是水,船才進得去 —— 但 stepObject 不查通行,直接測換圖。
		putObject(t, s, 5, u5data.SpawnEnemyShip, s.X+5, s.Y+5)
		s.stepObject(5, tc.dx, tc.dy)
		got := int(s.currentObjects().Objects[5].Raw[u5data.ObjKind]) - u5data.ShipTileBase
		if got != tc.want {
			t.Errorf("往%s走之後朝向是 %d,預期 %d", tc.name, got, tc.want)
		}
	}
}

// TestCalmStopsEnemyShips —— 無風時敵船不動(原版第一行就 `jz`)。
func TestCalmStopsEnemyShips(t *testing.T) {
	s := overworldScene(t)
	clearObjects(t, s)
	flatGrass(t, s, 12)
	ox, oy := s.X+6, s.Y+6
	putObject(t, s, 5, u5data.SpawnEnemyShip, ox, oy)
	s.Wind = u5data.WindCalm
	for i := 0; i < 30; i++ {
		s.objectMoveGate(5)
	}
	o := s.currentObjects().Objects[5]
	if int(o.Raw[u5data.ObjX]) != ox || int(o.Raw[u5data.ObjY]) != oy {
		t.Error("無風時敵船動了 —— 原版 `cmp byte_3E0A2, 0; jz` 先擋掉")
	}
}

// TestWhirlpoolMovesEveryOtherTurn —— 漩渦的 +5 是持久切換位元。
func TestWhirlpoolMovesEveryOtherTurn(t *testing.T) {
	s := overworldScene(t)
	clearObjects(t, s)
	flatGrass(t, s, 12)
	putObject(t, s, 5, WhirlpoolKind, s.X+6, s.Y+6)
	set := s.currentObjects()
	// 連續兩次:切換位元該在 1 / 0 之間翻。
	s.objectMoveGate(5)
	first := set.Objects[5].Raw[u5data.ObjQuality]
	s.objectMoveGate(5)
	second := set.Objects[5].Raw[u5data.ObjQuality]
	if first == second {
		t.Errorf("切換位元沒翻:%d → %d", first, second)
	}
}

// TestSailRhythmBurnsWorldTurnsByWind —— ★ 揚帆時風向多燒幾個世界回合。
//
// n = 1 + 不同的分量數;跑 n%3 次 ⇒ **同向 1 次、反向 2 次、垂直 0 次**。
//
// ⚠ 「垂直 0 次」反直覺(側風反而不花額外時間),但組語就是 `n % 3`。
// `CLAUDE.md §3.0` 不自行平衡 —— 照原樣,並列進 A 階段對 DOSBox 的核對清單。
func TestSailRhythmBurnsWorldTurnsByWind(t *testing.T) {
	// 帆朝北(Transport 低兩位 = North = 0)。
	for _, tc := range []struct {
		wind int
		want int
		why  string
	}{
		{u5data.WindSouth, 1, "南風把你往北推 = 與帆同向"},
		{u5data.WindNorth, 2, "北風把你往南推 = 與帆反向"},
		{u5data.WindEast, 0, "東風往西推 = 垂直"},
		{u5data.WindWest, 0, "西風往東推 = 垂直"},
	} {
		s := overworldScene(t)
		clearObjects(t, s)
		flatGrass(t, s, 12)
		s.Transport = u5data.VehicleSailing | byte(North)
		s.Wind = tc.wind
		before := s.WorldTurns
		s.sailRhythm()
		if got := s.WorldTurns - before; got != tc.want {
			t.Errorf("%s:多跑了 %d 個世界回合,預期 %d", tc.why, got, tc.want)
		}
	}

	// 沒揚帆或無風 → 完全不跑。
	s := overworldScene(t)
	s.Transport = u5data.VehicleShip
	s.Wind = u5data.WindNorth
	before := s.WorldTurns
	s.sailRhythm()
	if s.WorldTurns != before {
		t.Error("收帆的船也吃了揚帆節奏")
	}
}
