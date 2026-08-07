package u5data

import "testing"

// origSightDistance 是從 `WORRIORS.EXP` 的 `byte_601F0`(檔案位移 0x601F0+0x200)
// 直接 dump 出來的 121 個位元組。
//
// 放在測試裡而不是放進引擎:引擎用 `dx²+dy²` 算,這張表是**核對用的 oracle**。
// 兩邊只要有一格對不上就是解錯了。
var origSightDistance = [SightSide * SightSide]byte{
	50, 41, 34, 29, 26, 25, 26, 29, 34, 41, 50,
	41, 32, 25, 20, 17, 16, 17, 20, 25, 32, 41,
	34, 25, 18, 13, 10, 9, 10, 13, 18, 25, 34,
	29, 20, 13, 8, 5, 4, 5, 8, 13, 20, 29,
	26, 17, 10, 5, 2, 1, 2, 5, 10, 17, 26,
	25, 16, 9, 4, 1, 0, 1, 4, 9, 16, 25,
	26, 17, 10, 5, 2, 1, 2, 5, 10, 17, 26,
	29, 20, 13, 8, 5, 4, 5, 8, 13, 20, 29,
	34, 25, 18, 13, 10, 9, 10, 13, 18, 25, 34,
	41, 32, 25, 20, 17, 16, 17, 20, 25, 32, 41,
	50, 41, 34, 29, 26, 25, 26, 29, 34, 41, 50,
}

// 原版那張表就是 dx²+dy² —— 逐格核對,不是抽樣。
func TestSightDistanceIsSquaredEuclidean(t *testing.T) {
	for y := 0; y < SightSide; y++ {
		for x := 0; x < SightSide; x++ {
			got := sightDistance(x, y)
			want := int(origSightDistance[y*SightSide+x])
			if got != want {
				t.Fatalf("(%d,%d): 算出 %d,原版表是 %d", x, y, got, want)
			}
		}
	}
	if sightDistance(5, 5) != 0 {
		t.Fatal("中心的距離應該是 0")
	}
}

// 貼著才看得穿的五個地形:平方距離 1 通,其餘一律不通。
func TestDoorsAreOnlyTransparentWhenAdjacent(t *testing.T) {
	for _, tile := range SightDoors {
		if !SightPasses(tile, 1) {
			t.Fatalf("地形 %02X 貼著應該看得穿", tile)
		}
		// 2 是斜角相鄰 —— 原版判的是**平方**距離,所以斜角就不通了。
		for _, d := range []int{0, 2, 4, 5, 9, 50} {
			if SightPasses(tile, d) {
				t.Fatalf("地形 %02X 在平方距離 %d 不該看得穿", tile, d)
			}
		}
	}
}

// 十九個擋視線的地形,不管多遠都擋。
func TestBlockersAlwaysBlock(t *testing.T) {
	for _, tile := range SightBlockers {
		for _, d := range []int{0, 1, 2, 9, 25, 50} {
			if SightPasses(tile, d) {
				t.Fatalf("地形 %02X 在平方距離 %d 不該看得穿", tile, d)
			}
		}
	}
	// 沒列進去的一律通(草地 0x03、地板 0x01…)。
	for _, tile := range []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x40, 0x7F} {
		if !SightPasses(tile, 9) {
			t.Fatalf("地形 %02X 不在擋視線清單裡,應該看得穿", tile)
		}
	}
}

// 兩張清單不能有交集 —— 有的話 `SightPasses` 的判斷順序就會決定結果,
// 那是很容易在改動時翻掉的隱性依賴。
func TestDoorsAndBlockersDoNotOverlap(t *testing.T) {
	for _, d := range SightDoors {
		for _, b := range SightBlockers {
			if d == b {
				t.Fatalf("地形 %02X 同時在門與牆兩張清單裡", d)
			}
		}
	}
}

// openWindow 是一片空地(全部看得穿、全部亮著)。
func openWindow(tile byte) SightWindow {
	return SightWindow{
		Tile:    func(x, y int) byte { return tile },
		Lit: func(x, y int) bool { return true },
	}
}

// 一片空地、燈光半徑夠大 —— 整張都該看得到。
func TestOpenGroundIsFullyVisible(t *testing.T) {
	mask := ComputeSight(openWindow(0x03), 50)
	for i, v := range mask {
		if v != 0x03 {
			t.Fatalf("第 %d 格是 %02X,一片空地應該全部看得到", i, v)
		}
	}
}

// 燈光半徑不是「看得到多遠」的上限 —— 空地沒有東西擋,再小的半徑也傳得到角落。
//
// ⚠ 這條是刻意寫的:原版的半徑判斷是**捷徑**(半徑內不必檢查是否擋視線),
// 不是**距離上限**。抄成「超過半徑就黑掉」的話白天在野外只看得到身邊一圈。
func TestLightRadiusIsAShortcutNotADistanceLimit(t *testing.T) {
	mask := ComputeSight(openWindow(0x03), 1)
	for i, v := range mask {
		if v != 0x03 {
			t.Fatalf("第 %d 格是 %02X,燈光半徑 1 的空地仍該全部看得到", i, v)
		}
	}
}

// 半徑 0 = 沒有光源 = **整張全黑**,連腳下都看不到。
//
// ⚠ 原版是 `jle`(跳過視線計算)配 `jge`(跳過全攤開)兩個條件夾出來的,
// 0 落在中間,兩邊都不做,罩子維持初值。這一格很容易在重寫時「順手補上
// 至少看得到自己」—— 那就與原版不同了(無光的地牢原版是真的什麼都沒有)。
func TestNoLightMeansTotalDarkness(t *testing.T) {
	mask := ComputeSight(openWindow(0x03), 0)
	for i, v := range mask {
		if v != SightHidden {
			t.Fatalf("第 %d 格是 %02X,沒有光源時整張都該是 %02X", i, v, byte(SightHidden))
		}
	}
}

// Wis An Ylem(半徑 −1):整張直接可見,連牆後面都看得到。
func TestWisAnYlemRevealsEverything(t *testing.T) {
	w := SightWindow{
		// 除了中心那一格,整張都是牆。
		Tile: func(x, y int) byte {
			if x == 5 && y == 5 {
				return 0x03
			}
			return 0x4E
		},
		Lit: func(x, y int) bool { return true },
	}
	mask := ComputeSight(w, -1)
	for y := 0; y < SightSide; y++ {
		for x := 0; x < SightSide; x++ {
			v := mask[y*SightSide+x]
			want := byte(0x4E)
			if x == 5 && y == 5 {
				want = 0x03
			}
			if v != want {
				t.Fatalf("(%d,%d) 是 %02X,Wis An Ylem 下應該是 %02X", x, y, v, want)
			}
		}
	}
}

// 一道牆:牆本身看得到,牆後面看不到。
func TestAWallHidesWhatIsBehindIt(t *testing.T) {
	// 第 2 列(y=2)整排是牆,玩家在 (5,5)。
	w := SightWindow{
		Tile: func(x, y int) byte {
			if y == 2 {
				return 0x4E // 擋視線
			}
			return 0x03
		},
		Lit: func(x, y int) bool { return true },
	}
	mask := ComputeSight(w, 1)

	// 牆自己看得到。
	for x := 0; x < SightSide; x++ {
		if mask[2*SightSide+x] != 0x4E {
			t.Fatalf("(%d,2) 是 %02X,那面牆自己應該看得到", x, mask[2*SightSide+x])
		}
	}
	// 牆正後方(正上方那幾格)看不到。
	//
	// ⚠ 只驗正上方一小段:斜著繞過整排牆的邊緣是原版 BFS 的行為,
	// 拿「整排 y<2 都看不到」當條件會誤判成 bug。
	for x := 3; x <= 7; x++ {
		for y := 0; y <= 1; y++ {
			if mask[y*SightSide+x] != SightHidden {
				t.Fatalf("(%d,%d) 是 %02X,牆後面不該看得到", x, y, mask[y*SightSide+x])
			}
		}
	}
}

// 站在門邊看得進去,退一步就看不到 —— 五個「貼著才透」的地形的實際效果。
func TestStandingNextToADoorSeesThrough(t *testing.T) {
	// 門在 (5,4)(玩家正上方,平方距離 1),門後 (5,3) 是空地。
	w := SightWindow{
		Tile: func(x, y int) byte {
			switch {
			case x == 5 && y == 4:
				return 0x4B // 門
			case y <= 4:
				return 0x03
			}
			return 0x03
		},
		Lit: func(x, y int) bool { return true },
	}
	mask := ComputeSight(w, 1)
	if mask[4*SightSide+5] != 0x4B {
		t.Fatalf("門那一格是 %02X,應該看得到", mask[4*SightSide+5])
	}
	if mask[3*SightSide+5] != 0x03 {
		t.Fatalf("門後 (5,3) 是 %02X,貼著門應該看得進去", mask[3*SightSide+5])
	}

	// 同一扇門移到平方距離 4(正上方兩格)—— 這次擋住。
	w2 := SightWindow{
		Tile: func(x, y int) byte {
			if x == 5 && y == 3 {
				return 0x4B
			}
			// 兩側全是牆,逼視線只能從門過。
			if x != 5 {
				return 0x4E
			}
			return 0x03
		},
		Lit: func(x, y int) bool { return true },
	}
	mask2 := ComputeSight(w2, 1)
	if mask2[2*SightSide+5] != SightHidden {
		t.Fatalf("(5,2) 是 %02X,隔兩格的門不該看得穿", mask2[2*SightSide+5])
	}
}

// 沒有光照到的格子看不見 —— 這就是夜晚會變暗的機制。
//
// ⚠ **但玩家自己的半徑內是例外**:原版的半徑判斷(`jl loc_2E065`)排在
// 亮度檢查**之前**,所以半徑內的格子連亮不亮都不問就直接畫。
// 也就是說再暗都看得到身邊那一小圈。重寫時很容易「順手」把亮度檢查提前,
// 那會讓玩家在全黑處連腳邊都看不到。這裡把兩件事都釘住。
func TestUnlitCellsAreDarkExceptInsideTheLightRadius(t *testing.T) {
	w := SightWindow{
		Tile: func(x, y int) byte { return 0x03 },
		// 只有右半邊亮著。
		Lit: func(x, y int) bool { return x >= 5 },
	}
	mask := ComputeSight(w, 1)
	for y := 0; y < SightSide; y++ {
		for x := 0; x < 5; x++ {
			// (4,5) 的平方距離是 1,落在燈光半徑內 —— 原版照畫。
			if x == 4 && y == 5 {
				if mask[y*SightSide+x] != 0x03 {
					t.Fatalf("(4,5) 是 %02X,半徑內原版不問亮度直接畫",
						mask[y*SightSide+x])
				}
				continue
			}
			if mask[y*SightSide+x] != SightHidden {
				t.Fatalf("(%d,%d) 是 %02X,沒照到光應該是黑的", x, y, mask[y*SightSide+x])
			}
		}
	}
	if mask[5*SightSide+5] != 0x03 {
		t.Fatal("中心應該看得到")
	}
}

// 中心那一格永遠看得到 —— 就算腳下站的是擋視線的地形。
func TestTheCentreIsAlwaysVisible(t *testing.T) {
	for _, tile := range SightBlockers {
		w := openWindow(tile)
		mask := ComputeSight(w, 1)
		if mask[5*SightSide+5] != tile {
			t.Fatalf("腳下是 %02X 時中心變成 %02X", tile, mask[5*SightSide+5])
		}
	}
}

// 四面都是牆的密室:只看得到那一圈牆,外面全黑。
func TestASealedRoomShowsOnlyItsWalls(t *testing.T) {
	w := SightWindow{
		Tile: func(x, y int) byte {
			if x >= 4 && x <= 6 && y >= 4 && y <= 6 {
				return 0x03
			}
			return 0x4E
		},
		Lit: func(x, y int) bool { return true },
	}
	mask := ComputeSight(w, 1)
	// 3×3 的房間裡面看得到。
	for y := 4; y <= 6; y++ {
		for x := 4; x <= 6; x++ {
			if mask[y*SightSide+x] != 0x03 {
				t.Fatalf("(%d,%d) 是 %02X,房間裡應該看得到", x, y, mask[y*SightSide+x])
			}
		}
	}
	// 貼著房間的那一圈牆看得到。
	for _, p := range [][2]int{{3, 5}, {7, 5}, {5, 3}, {5, 7}} {
		if mask[p[1]*SightSide+p[0]] != 0x4E {
			t.Fatalf("(%d,%d) 是 %02X,貼著的牆應該看得到", p[0], p[1], mask[p[1]*SightSide+p[0]])
		}
	}
	// 牆再外面一圈全黑。
	for _, p := range [][2]int{{1, 5}, {9, 5}, {5, 1}, {5, 9}, {0, 0}, {10, 10}} {
		if mask[p[1]*SightSide+p[0]] != SightHidden {
			t.Fatalf("(%d,%d) 是 %02X,密室外不該看得到", p[0], p[1], mask[p[1]*SightSide+p[0]])
		}
	}
}

// flood fill 一定收斂,而且 121 格全部走到。
//
// ⚠ 這裡**不能**要求「每格只判一次」:被判成看不見的牆會維持 0xFF,
// 於是從別的方向繞過來時會再判一次 —— 那正是「同一面牆從亮的那側看得到、
// 從暗的那側看不到」的來源。收斂靠的是「牆不入列」,不是走訪表。
//
// 上限抓得寬鬆但有限(121 格 × 8 鄰):真的漏掉終止條件時會直接吊死,
// 這條測試就會逾時而不是安靜通過。
func TestTheFloodFillTerminatesAndCoversEveryCell(t *testing.T) {
	visits := map[int]int{}
	total := 0
	w := SightWindow{
		// 一片牆:最容易讓「牆會重判」變成無窮迴圈的地形。
		Tile: func(x, y int) byte {
			visits[y*SightSide+x]++
			total++
			if x == 5 && y == 5 {
				return 0x03
			}
			return 0x4E
		},
		Lit: func(x, y int) bool { return true },
	}
	ComputeSight(w, 1)
	if total > SightSide*SightSide*8 {
		t.Fatalf("查了 %d 次地形,遠超過 121×8 —— flood fill 沒收斂", total)
	}

	// 空地要 121 格全走到。
	visits = map[int]int{}
	w2 := SightWindow{
		Tile: func(x, y int) byte {
			visits[y*SightSide+x]++
			return 0x03
		},
		Lit: func(x, y int) bool { return true },
	}
	ComputeSight(w2, 1)
	if len(visits) != SightSide*SightSide {
		t.Fatalf("只走訪了 %d 格,應該是 %d 格", len(visits), SightSide*SightSide)
	}
}

// 同一面牆:亮的那側看得到,暗的那側看不到。
//
// ★ 這條驗的是「走訪判斷用罩子本身、不用另一份走訪表」那個設計。
// 換成走訪表的話,牆的可見與否會變成「BFS 先從哪一側碰到它」——
// 畫面上看起來只是「有時候牆會閃一下」,很難追。
func TestAWallIsVisibleFromTheLitSideOnly(t *testing.T) {
	// (5,3) 是牆;它的北邊(5,2)沒照到光,南邊是玩家(亮)。
	w := SightWindow{
		Tile: func(x, y int) byte {
			if y == 3 {
				return 0x4E
			}
			return 0x03
		},
		Lit: func(x, y int) bool { return y >= 3 },
	}
	mask := ComputeSight(w, 1)
	if mask[3*SightSide+5] != 0x4E {
		t.Fatalf("(5,3) 是 %02X,從亮的南側應該看得到那面牆", mask[3*SightSide+5])
	}
	for x := 0; x < SightSide; x++ {
		for y := 0; y <= 2; y++ {
			if mask[y*SightSide+x] != SightHidden {
				t.Fatalf("(%d,%d) 是 %02X,沒照到光不該看得到", x, y, mask[y*SightSide+x])
			}
		}
	}
}
