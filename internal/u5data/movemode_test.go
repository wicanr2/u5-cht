package u5data

import "testing"

// TestMoverModeTableAnchors:六個錨點同時對上,表的對齊就沒有滑動的餘地。
//
// 錨點全部是**別處獨立推出來的**載具常數(買馬、上下載具、風向那幾支),
// 與 `byte_5FF8C` 完全無關。六個一起中,不是巧合(`rulebook/62`)。
func TestMoverModeTableAnchors(t *testing.T) {
	cases := []struct {
		mover byte
		want  MoveMode
		why   string
	}{
		{TileHorse, MoveHorse, "馬"},
		{TileHorse + HorseToVehicle, MoveHorse, "騎在馬上"},
		{VehicleCarpet, MoveAmphibious, "魔毯兩棲"},
		{VehicleWalk, MoveWalk, "步行"},
		{VehicleSailing, MoveShip, "揚帆中"},
		{VehicleShip, MoveShip, "大船"},
		{VehicleSkiff, MoveSkiff, "小艇"},
	}
	for _, c := range cases {
		if got := ModeOf(byte(c.mover)); got != c.want {
			t.Errorf("%s(0x%02X)→ 模式 %d,預期 %d", c.why, c.mover, got, c.want)
		}
	}
}

// TestShipCannotEnterShoals:大船吃水深 —— 淺灘過不去。
//
// 小艇可以。這是 U5 航海的核心限制:大船靠不了岸,得放小艇。
// 兩種都寫成「只要是水就行」的話,玩家可以開大船直接撞上沙灘。
func TestShipCannotEnterShoals(t *testing.T) {
	const shoals = 3
	if MoveModeAllows(MoveShip, shoals) {
		t.Error("大船開進了淺灘")
	}
	if !MoveModeAllows(MoveSkiff, shoals) {
		t.Error("小艇進不了淺灘")
	}
	// 深水兩者都行。
	for _, mode := range []MoveMode{MoveShip, MoveSkiff} {
		if !MoveModeAllows(mode, 1) {
			t.Errorf("模式 %d 進不了深水", mode)
		}
	}
}

// TestHorseRefusesWater:馬不下水,而兩棲的會。
func TestHorseRefusesWater(t *testing.T) {
	for _, tile := range []int{0, 1, 2, 3, 0x60, 0x6F} {
		if MoveModeAllows(MoveHorse, tile) {
			t.Errorf("馬走進了水 0x%02X", tile)
		}
		if !MoveModeAllows(MoveAmphibious, tile) {
			t.Errorf("兩棲的過不了水 0x%02X", tile)
		}
	}
}

// TestWaterTileGroups:水有兩段,不是只有 tile < 4。
//
// 原版 `sub_2A674` 是 `tile < 4 || (tile & 0xF0) == 0x60`。
// 漏掉 0x60 那一段的話,那十六格水在船的判定裡會變成陸地。
func TestWaterTileGroups(t *testing.T) {
	for tile := 0; tile < 4; tile++ {
		if !TileIsWater(tile) {
			t.Errorf("0x%02X 該是水", tile)
		}
	}
	for tile := 0x60; tile <= 0x6F; tile++ {
		if !TileIsWater(tile) {
			t.Errorf("0x%02X 該是水(第二段)", tile)
		}
	}
	for _, tile := range []int{4, 5, 0x5F, 0x70} {
		if TileIsWater(tile) {
			t.Errorf("0x%02X 不該是水", tile)
		}
	}
}

// TestFlyerGoesAnywhere:飛行的什麼都不管。
func TestFlyerGoesAnywhere(t *testing.T) {
	for _, tile := range []int{0, 3, 0x3E, 0x4F, 0x60, 0xFF} {
		if !MoveModeAllows(MoveFlyer, tile) {
			t.Errorf("飛行的過不了 0x%02X", tile)
		}
	}
}

// TestWalkModeMatchesTheBitmap:一般陸行就是那張阻擋位圖,不多不少。
//
// 這條把兩者綁在一起 —— 哪天有人在 MoveWalk 裡加了額外條件,
// 玩家的通行判定就會與 NPC 的分岔,而那種分岔極難察覺。
func TestWalkModeMatchesTheBitmap(t *testing.T) {
	for tile := 0; tile < TileCount; tile++ {
		if MoveModeAllows(MoveWalk, tile) == TileBlocksWalking(tile) {
			t.Fatalf("tile 0x%02X:一般陸行與阻擋位圖不一致", tile)
		}
	}
}
