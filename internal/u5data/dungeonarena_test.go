package u5data

import "testing"

// 死巷裡打,戰場就只有一個出口。
//
// 這是整個地牢戰場最重要的一條:形狀來自你剛才站的路口。
func TestArenaOpensOnlyWhereTheCorridorIs(t *testing.T) {
	a := DungeonArena{Number: 2, Here: DungeonPassage, Facing: 0}
	// 只有北邊是通道,其餘三面是牆。
	a.Around = [4]byte{DungeonPassage, DungeonWall, DungeonWall, DungeonWall}
	m := BuildDungeonArena(a)
	wall, floor := DungeonArenaTiles(2)

	// 北面在內牆框那一列(y = 1)開了七格。
	for x := 2; x <= 8; x++ {
		if m.At(x, 1) != floor {
			t.Errorf("北面 (%d,1) 是 0x%02X,預期開成地板 0x%02X", x, m.At(x, 1), floor)
		}
	}
	// 開口以外還是牆。
	//
	// ⚠ 只檢查 x = 1 與 9。x = 0 與 10 在**外環**上,而東西兩面被封的時候
	// `sub_FAC4` 是整條外環一起抹成空白 —— 它跑在框畫好之後,所以會蓋掉那兩格。
	// 我第一版把這兩格也當成牆,測試就紅了;紅的是預期,不是實作。
	for _, x := range []int{1, 9} {
		if m.At(x, 1) != wall {
			t.Errorf("北面 (%d,1) 是 0x%02X,預期還是牆 0x%02X", x, m.At(x, 1), wall)
		}
	}
	// 東 / 南 / 西三面被整條封成空白。
	for i := 0; i < CombatSide; i++ {
		if m.At(CombatSide-1, i) != TileBlank {
			t.Errorf("東面 (%d,%d) 沒有封起來(0x%02X)", CombatSide-1, i, m.At(CombatSide-1, i))
		}
		if m.At(i, CombatSide-1) != TileBlank {
			t.Errorf("南面 (%d,%d) 沒有封起來", i, CombatSide-1)
		}
		if m.At(0, i) != TileBlank {
			t.Errorf("西面 (0,%d) 沒有封起來", i)
		}
	}
}

// 房間(0xA0 / 0xF0)與門(0xE0)開的是**五格窄口**,不是七格。
func TestArenaNarrowGapForRoomsAndDoors(t *testing.T) {
	_, floor := DungeonArenaTiles(2)
	for _, tile := range []byte{DungeonRoomA, DungeonRoomF, DungeonDoorway} {
		m := BuildDungeonArena(DungeonArena{
			Number: 2, Here: DungeonPassage, Facing: 0,
			Around: [4]byte{tile, DungeonWall, DungeonWall, DungeonWall},
		})
		open := 0
		for x := 0; x < CombatSide; x++ {
			if m.At(x, 1) == floor {
				open++
			}
		}
		if open != 5 {
			t.Errorf("鄰格 0x%02X 開了 %d 格,預期 5 格", tile, open)
		}
	}
}

// 腳下的地形會擺在戰場正中央:梯子、寶箱、噴泉各有對應的 tile。
func TestArenaCentreShowsWhatYouStandOn(t *testing.T) {
	// ⚠ 「什麼都不擺」的結果是**地板**,不是空白 —— 底色由 `sub_FE48` 先填成
	// `byte_418DD`(地板),`sub_FD54` 只在有擺設時覆蓋中央那一格。
	// 我第一版預期空白,那是因為把底色也寫成了 0xFF(見 `docs/re/53`)。
	_, floor := DungeonArenaTiles(2)
	cases := map[byte]byte{
		DungeonPassage:    floor, // 什麼都不擺 → 留著地板
		DungeonLadderUp:   0xC8,
		DungeonLadderDown: 0xC9,
		DungeonLadderBoth: 0xC8, // 兩向梯畫成上行
		DungeonChest:      0xDC,
		DungeonFountain:   0xD8,
		DungeonTrap:       floor, // 陷阱不擺
	}
	for here, want := range cases {
		m := BuildDungeonArena(DungeonArena{Number: 2, Here: here, Facing: 0})
		if got := m.At(dungeonArenaCentre, dungeonArenaCentre); got != want {
			t.Errorf("腳下 0x%02X:中央是 0x%02X,預期 0x%02X", here, got, want)
		}
	}
	// 而且戰場中間**站得住人** —— 底色是地板不是空白。
	m := BuildDungeonArena(DungeonArena{Number: 2, Here: DungeonPassage, Facing: 0})
	for y := 2; y <= 8; y++ {
		for x := 2; x <= 8; x++ {
			if m.At(x, y) == TileBlank {
				t.Fatalf("框內 (%d,%d) 是空白,底色應該是地板 0x%02X", x, y, floor)
			}
		}
	}
}

// 四個朝向的入場位置必須互不相同,而且全部落在框內(2..8)。
//
// ⚠ 這條擋的是「四張表配對接錯」——接錯了座標還是 0..10,只是隊伍會站到
// 牆上或錯的一側,而那種錯在測試不看內容時完全看不出來。
func TestArenaPartyEntersFromTheSideItFaces(t *testing.T) {
	seen := map[string]int{}
	for facing := 0; facing < 4; facing++ {
		side := DungeonArenaSide(facing)
		x, y := DungeonArenaParty(side)
		key := ""
		for i := 0; i < CombatPartySlots; i++ {
			if x[i] < 2 || x[i] > 8 || y[i] < 2 || y[i] > 8 {
				t.Errorf("朝向 %d(側 %d)第 %d 人在 (%d,%d),超出框內 2..8",
					facing, side, i, x[i], y[i])
			}
			key += string(rune('0'+x[i])) + string(rune('0'+y[i]))
		}
		seen[key]++
	}
	if len(seen) != 4 {
		t.Errorf("四個朝向只產生 %d 種隊形,預期 4 種", len(seen))
	}
	// 朝北 → 從南邊入場,所以六人的 y 都要在下半場。
	_, y := DungeonArenaParty(DungeonArenaSide(0))
	for i, v := range y {
		if v <= dungeonArenaCentre {
			t.Errorf("朝北時第 %d 人的 y = %d,應該在南半場(> %d)", i, v, dungeonArenaCentre)
		}
	}
}

// 敵人入場點在四個朝向下是同一個形狀轉 90 度 —— 用原版表自己的對稱性檢查。
func TestArenaEnemyTablesAreRotations(t *testing.T) {
	// byte_41920 = 10 − byte_41930 逐項成立。
	for i := 0; i < CombatEnemySlots; i++ {
		if dungeonArenaEnemyB[i]+dungeonArenaEnemyC[i] != 10 {
			t.Errorf("第 %d 項:%d + %d != 10", i, dungeonArenaEnemyB[i], dungeonArenaEnemyC[i])
		}
	}
	// 朝東與朝南只是把 X / Y 互換。
	ex1, ey1 := DungeonArenaEnemies(1)
	ex2, ey2 := DungeonArenaEnemies(2)
	if ex1 != ey2 || ey1 != ex2 {
		t.Errorf("朝東與朝南不是互換 X / Y")
	}
	for i := 0; i < CombatEnemySlots; i++ {
		if ex1[i] >= CombatSide || ey1[i] >= CombatSide {
			t.Errorf("第 %d 個敵人入場點 (%d,%d) 超出 11×11", i, ex1[i], ey1[i])
		}
	}
}

// 兩組牆地板要真的不同,而且只有三個地牢用第一組。
func TestArenaTilesDifferForThreeDungeons(t *testing.T) {
	alt := 0
	for n := 1; n <= DungeonCount; n++ {
		w, f := DungeonArenaTiles(n)
		if w == dungeonArenaWallA {
			alt++
			if f != dungeonArenaFloorA {
				t.Errorf("地牢 %d 的牆地板配錯了", n)
			}
		}
	}
	if alt != len(DungeonArenaAltSet) {
		t.Errorf("%d 座地牢用第一組牆地板,預期 %d 座", alt, len(DungeonArenaAltSet))
	}
	if dungeonArenaWallA == dungeonArenaWallB || dungeonArenaFloorA == dungeonArenaFloorB {
		t.Errorf("兩組牆地板不該相同")
	}
}
