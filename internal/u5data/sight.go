package u5data

// 視線遮蔽(原版 `sub_2E0E8` → `sub_2DDB0`,判定用 `sub_2E1D0` 與 `sub_2E8D0`)
//
// 11×11 的視窗裡,哪一格看得到、哪一格被牆擋住。原版每次重畫畫面都跑一次
//(`sub_29D64`),Wis An Ylem 則是**把整張罩子填成全可見**再重畫二十幀
//(`sub_1CE0C`)—— 所以那個咒語不是「揭開地圖」,是「這一瞬間牆擋不住視線」。
//
// ⚠ 結果**不是**一張 0/1 的可見遮罩,而是「這一格該畫什麼」:
// 看得到就放地形碼,看不到(牆後、燈光外)一律放 `SightHidden`(0xFF)。
// 原版的 `sub_2E0E8` 在最後把值為 0 的格子 `dec` 成 0xFF —— 兩種「看不見」
// 就這樣併成同一個值。照抄,呈現端只要判 0xFF。

// SightHidden 是「這一格看不到」。
const SightHidden = 0xFF

// SightSide 是視線視窗的邊長。
const SightSide = 11

// sightDistance 是 11×11 的**平方距離**表(原版 `byte_601F0`)。
//
// ★ 它就是 `dx² + dy²`(中心 (5,5)):角落 25+25 = 50、正上方 25、
// 斜對角一格 1+1 = 2。整張表 121 個值全部對得上 ——
// 所以引擎直接算,不必把表抄進來(`TestSightDistanceIsSquaredEuclidean`
// 拿原版的值逐格核對)。
func sightDistance(x, y int) int {
	dx, dy := x-SightSide/2, y-SightSide/2
	return dx*dx + dy*dy
}

// SightBlockers 是擋住視線的地形(原版 `dword_601D0`,19 筆)。
//
//	09 0A 0C 0D  山與丘
//	4D 4E 4F     樹與牆
//	5A 97
//	B8 B9 BC     門
//	D0 D1 D2 D3
//	F8 FE FF
var SightBlockers = [19]byte{
	0x09, 0x0A, 0x0C, 0x0D, 0x4D, 0x4E, 0x4F, 0x5A, 0x97,
	0xB8, 0xB9, 0xBC, 0xD0, 0xD1, 0xD2, 0xD3, 0xF8, 0xFE, 0xFF,
}

// SightDoors 是「**貼著才看得穿**」的地形(原版 `sub_2E1D0` 前半的五個比較)。
//
// 窗與門:站在旁邊(平方距離 1,也就是正上下左右一格)看得進去,
// 退一步就看不到了。
//
// ⚠ 用的是**平方**距離,所以「1」只涵蓋正交相鄰,斜角(2)不算。
var SightDoors = [5]byte{0x4A, 0x4B, 0xBA, 0xBB, 0x98}

// SightPasses 回報視線穿不穿得過這一格。dist2 是離視野中心的平方距離。
func SightPasses(tile byte, dist2 int) bool {
	for _, d := range SightDoors {
		if tile == d {
			return dist2 == 1
		}
	}
	for _, b := range SightBlockers {
		if tile == b {
			return false
		}
	}
	return true
}

// sightRing 是原版走訪八鄰的順序(`sub_2DDB0` 的 `dir` 從 7 遞減到 0)。
//
// ★ 原版**沒有八組偏移量** —— 它讓 `esi`/`edi` 一路累加,一圈走完剛好回到起點
// 附近。攤開來就是從西邊起、逆時針繞一圈:
//
//	dir 7  x−1        dir 3  y−1
//	dir 6  y+1        dir 2  y−1
//	dir 5  x+1        dir 1  x−1
//	dir 4  x+1        dir 0  x−1
//
// 這裡改寫成八個絕對偏移(等價,而且讀得懂)。順序不能換 ——
// 它決定了「同時有兩條路徑照到同一格」時誰先寫進去。
var sightRing = [8][2]int{
	{-1, 0}, {-1, 1}, {0, 1}, {1, 1}, {1, 0}, {1, -1}, {0, -1}, {-1, -1},
}

// SightWindow 是餵給 `ComputeSight` 的 11×11 地形。
type SightWindow struct {
	// Tile 取視窗座標 (x, y) 的地形;超出場景時回 0。
	Tile func(x, y int) byte
	// InScene 回報視窗座標 (x, y) 在不在場景範圍內。
	//
	// 原版判的是「加上視窗左上角之後落在 0..31」而且 `byte_3F8F4` 那一格非 0
	//(那是已經搬進視窗緩衝的場景地圖)。
	InScene func(x, y int) bool
}

// sightDark 是**演算過程中**「這一格是黑的」。
//
// ⚠ 它與「還沒走過」是**兩個不同的值**,不能合併:
// 「還沒走過」是 0xFF、「黑的」是 0。牆要不要露出來的判斷讀的正是
// 來源那一格「是不是 0」(`cmp byte[edx+eax], 0`)—— 用 0xFF 去比會反過來。
// 跑完之後 `sub_2E0E8` 才把 0 一律 `dec` 成 0xFF,對外只剩一種「看不見」。
const sightDark = 0

// ComputeSight 算出 11×11 的「這一格該畫什麼」。
//
// radius2 是燈光的平方半徑(原版 `byte_3E0B5`),三種情形**完全不同**:
//
//	> 0  正常:半徑內無條件看得到,半徑外才輪到視線判定
//	= 0  **全黑** —— 沒有光源,連腳下都看不到
//	< 0  Wis An Ylem:整張直接攤開,牆擋不住
//
// ⚠ 中間那一條容易抄錯。原版是 `jle`(跳過視線計算)配 `jge`(跳過全攤開)
// 兩個條件夾出來的:0 兩邊都跳過,罩子維持初值全 0xFF。抄成「半徑 0 = 只看腳下」
// 的話,無光的地牢會變成能看到自己站的那一格,而原版是真的什麼都沒有。
//
// 演算法照 `sub_2DDB0`,是一個佇列式的 flood fill:
//
//	罩子初值全 0xFF(= 還沒走過);中心 (5,5) 先填上腳下的地形並入列
//	取出一格 → 依 sightRing 繞它的八鄰(順序見上)
//	  出界、或那一格不是 0xFF(走過了) → 跳過
//	  平方距離 ≤ radius²  → 看得到,寫地形
//	  穿不過去            → 三個條件都成立才讓那面牆露出來:
//	                          來源不是黑的、來源在場景內、自己也在場景內
//	                        否則寫 0xFF(看不到)
//	  穿得過              → 在場景內寫地形,否則寫 0(黑)
//	  這一格的地形穿得過視線 → 入列繼續往外傳
//
// ★ 「走過了」這個判斷用的是**罩子本身**,不是另一份走訪表 ——
// 所以被判成 0xFF 的牆**還會被別的方向再判一次**。這不是 bug:
// 同一面牆從亮的那側看得到、從暗的那側看不到,靠的就是這個重判。
// 換成走訪表會讓牆的可見與否取決於 BFS 先碰到哪一邊。
func ComputeSight(w SightWindow, radius2 int) [SightSide * SightSide]byte {
	var mask [SightSide * SightSide]byte
	for i := range mask {
		mask[i] = SightHidden
	}
	if radius2 < 0 {
		// Wis An Ylem:`sub_2E0E8(-1, …)` 跳過 flood fill,整張直接填地形。
		for y := 0; y < SightSide; y++ {
			for x := 0; x < SightSide; x++ {
				mask[y*SightSide+x] = w.Tile(x, y)
			}
		}
		return mask
	}
	if radius2 == 0 {
		// 沒有光源 —— 整張維持 0xFF。
		return mask
	}
	// ⚠ 原版一開始 `inc arg_0`,判斷寫成 `dist2 < radius²+1`。
	// 等價於 `dist2 <= radius²` —— 抄成 `<` 會少掉整整一圈。
	limit := radius2

	const cx, cy = SightSide / 2, SightSide / 2
	mask[cy*SightSide+cx] = w.Tile(cx, cy)
	queue := [][2]int{{cx, cy}}

	for len(queue) > 0 {
		from := queue[0]
		queue = queue[1:]
		for _, d := range sightRing {
			x, y := from[0]+d[0], from[1]+d[1]
			if x < 0 || x >= SightSide || y < 0 || y >= SightSide {
				continue
			}
			idx := y*SightSide + x
			if mask[idx] != SightHidden {
				continue
			}

			tile := w.Tile(x, y)
			dist2 := sightDistance(x, y)
			switch {
			case dist2 <= limit:
				// 燈光範圍內 —— 無條件看得到。
				mask[idx] = tile
			case !SightPasses(tile, dist2):
				// 擋視線的格子**自己還是看得到**(汝看得到那面牆),
				// 但來源不能是黑的,而且兩格都要在場景內。
				if mask[from[1]*SightSide+from[0]] != sightDark &&
					w.InScene(from[0], from[1]) && w.InScene(x, y) {
					mask[idx] = tile
				} else {
					mask[idx] = SightHidden
					tile = SightHidden
				}
			case !w.InScene(x, y):
				// 場景外:黑的。
				mask[idx] = sightDark
			default:
				mask[idx] = tile
			}
			// 穿得過的才繼續往外傳。
			if SightPasses(tile, dist2) {
				queue = append(queue, [2]int{x, y})
			}
		}
	}

	// `sub_2E0E8` 的收尾:黑的(0)併成 0xFF。
	//
	// ⚠ 地形碼 0x00 的格子會被一起併掉 —— 原版就是這樣,照抄。
	for i, v := range mask {
		if v == sightDark {
			mask[i] = SightHidden
		}
	}
	return mask
}
