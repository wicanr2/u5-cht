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
	// ⚠ **tile 0 不在這份清單裡。** `TileIsWater(0)` 為真(原版 `sub_2A674`
	// 把 0 併進去了),但阻擋位圖的第 0 位是 **0**,而坐騎判的是位圖 ——
	// 所以原版的馬**踩得上 tile 0**。第一版的測試把 0 也列進來,
	// 釘住的是「馬應該怕水」的直覺,不是原版的判定式。
	for _, tile := range []int{1, 2, 3, 0x60, 0x6F} {
		if MoveModeAllows(MoveHorse, tile) {
			t.Errorf("馬走進了水 0x%02X", tile)
		}
		if !MoveModeAllows(MoveAmphibious, tile) {
			t.Errorf("兩棲的過不了水 0x%02X", tile)
		}
	}
	// 而沼澤(4)是坐騎**額外**擋的那一格 —— 位圖沒擋它,case 3 自己擋。
	if TileBlocksWalking(horseBlockSwamp) {
		t.Errorf("沼澤竟然在阻擋位圖裡,那 case 3 多寫的那條就沒意義了")
	}
	if MoveModeAllows(MoveHorse, horseBlockSwamp) {
		t.Error("馬過了沼澤")
	}
	if MoveModeAllows(MoveHorse, horseBlockOdd) {
		t.Errorf("馬過了 0x%02X", horseBlockOdd)
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

// TestFlyerIgnoresWallsButNotWater:飛行的不看位圖,但**過不了水**。
//
// ⚠ 這支的前身是 `TestFlyerGoesAnywhere`,釘住「飛行的什麼都不管」——
// 而原版 case 4 是 `sub_2A674(tile); setz` = **只放行非水**。
// 少了後半,飛行怪物會直接飛過海洋追殺玩家,而測試會一路綠。
func TestFlyerIgnoresWallsButNotWater(t *testing.T) {
	// 牆與山:位圖擋得住,但飛行的過。
	for _, tile := range []int{0x3E, 0x4F, 0xFF} {
		if !MoveModeAllows(MoveFlyer, tile) {
			t.Errorf("飛行的過不了 0x%02X", tile)
		}
	}
	// 水:飛不過。
	for _, tile := range []int{0, 1, 2, 3, 0x60, 0x6F} {
		if MoveModeAllows(MoveFlyer, tile) {
			t.Errorf("飛行的飛過了水 0x%02X", tile)
		}
	}
}

// TestOneTerrainModesAcceptExactlyOneTile:模式 7..10 各只認一種地形。
func TestOneTerrainModesAcceptExactlyOneTile(t *testing.T) {
	want := map[MoveMode]int{
		MoveSwampOnly:   swampTile,
		MoveGrassOnly:   grassTile,
		MoveShallowOnly: shallowTile,
		MoveDesertOnly:  desertTile,
	}
	for mode, only := range want {
		n := 0
		for tile := 0; tile < TileCount; tile++ {
			if MoveModeAllows(mode, tile) {
				n++
				if tile != only {
					t.Errorf("模式 %d 放行了 0x%02X,只該放行 0x%02X", mode, tile, only)
				}
			}
		}
		if n != 1 {
			t.Errorf("模式 %d 放行了 %d 種地形,應該只有 1 種", mode, n)
		}
	}
}

// TestShipStaysInDeepWater:大船只走 tile 0..2,連 0x6x 那族水都進不去。
//
// ★ 這就是「大船靠不了岸、得放小艇」的真正機制。第一版放行了 0x6x,
// 等於讓大船開得進河道。
func TestShipStaysInDeepWater(t *testing.T) {
	for tile := 0; tile <= ShipDeepMax; tile++ {
		if !MoveModeAllows(MoveShip, tile) {
			t.Errorf("大船走不了 0x%02X", tile)
		}
	}
	for _, tile := range []int{3, 4, 0x60, 0x6F, 0x70} {
		if MoveModeAllows(MoveShip, tile) {
			t.Errorf("大船開進了 0x%02X", tile)
		}
	}
}

// TestUnknownModeBlocks:表上的 0xFF 走 default,而 default 是**擋住**。
//
// ⚠ 第一版寫成「與一般陸行同」,那讓 0xE8..0xEB 那組移動者憑空獲得通行能力。
func TestUnknownModeBlocks(t *testing.T) {
	for _, tile := range []int{0, 5, 0x4F} {
		if MoveModeAllows(MoveModeNone, tile) {
			t.Errorf("未知模式放行了 0x%02X", tile)
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
