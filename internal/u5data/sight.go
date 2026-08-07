package u5data

// 視線遮蔽與場景照明(原版 `sub_2E0E8` / `sub_2E21C` → `sub_2DDB0`)
//
// 兩件事共用同一支 flood fill,差別在一個旗標:
//
//	sub_2E21C  掃出場景裡會發光的地形,**每一個各跑一次** → 得到「哪裡是亮的」
//	sub_2E0E8  以玩家為中心跑一次,查上面那張亮度圖 → 得到「哪一格該畫什麼」
//
// ★ 這個順序很要緊。玩家那一輪判「看不看得到」時,牆外那一格的依據是
// **它亮不亮**,而不是「它在不在場景裡」—— 這正是夜晚會變暗的原因:
// 天黑時玩家自己的半徑只剩 2,再遠的東西要有場景光源照著才看得見。
//
// ⚠ 我第一版把那個查表誤讀成「在不在場景範圍內」,於是夜晚與白天畫面相同。
// 是把 `sub_2E21C` 讀完才發現 `byte_3F8F4` 是**亮度圖**不是場景圖
//(它一開始填 0xFF,只有光源照到的格子留著值,最後把沒照到的 0xFF 變成 0)。
// 記在這裡:同一個緩衝區被兩支函式用作不同語意,只讀其中一支必定推錯。

// SightHidden 是「這一格看不到」。
const SightHidden = 0xFF

const (
	// SightSide 是視線視窗的邊長。
	SightSide = 11
	// SightStride 是罩子與亮度圖的列距(原版一律 `shl eax, 5`)。
	SightStride = 32
	// SightSceneSide 是亮度圖的邊長(`byte_3F8F4` 是 0x400 B)。
	//
	// 它涵蓋的是**以視窗左上角為原點**的 32×32,不是整張場景地圖。
	SightSceneSide = 32
)

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

// SightBlockers 是擋住視線的地形(原版 `dword_601D0`,19 筆,後面接一個 0 收尾)。
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

// SightLightTiles 是會發光的地形(原版 `dword_601E4`,10 筆,後面接一個 0 收尾)。
//
//	DC     月門(`sub_DEE4` 把它畫進場景時就是這個碼)
//	BD BE  火盆
//	B2 BF
//	B0 B1 B3
//	DE     營火
//	BC     ★ **同時也在 `SightBlockers` 裡** —— 牆上帶火的窗:
//	       擋住視線,但會照亮四周。兩張表的交集只有這一個。
var SightLightTiles = [10]byte{0xDC, 0xBD, 0xBE, 0xB2, 0xDE, 0xBF, 0xB0, 0xB1, 0xB3, 0xBC}

// SightLightRadius2 是一個場景光源照多遠(`sub_2E21C` 的 `push 0Ah`)。
//
// = 10,約三格半。與玩家自己的半徑無關 —— 這是固定值。
const SightLightRadius2 = 0x0A

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

// SightEmitsLight 回報這個地形會不會發光。
func SightEmitsLight(tile byte) bool {
	for _, t := range SightLightTiles {
		if tile == t {
			return true
		}
	}
	return false
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

// sightFill 是 `sub_2DDB0` 的參數。全部照原版的 arg 對應,不重新命名語意 ——
// 兩個呼叫端傳的東西差很多,硬套一個「好聽」的名字反而對不回組語。
type sightFill struct {
	// radius2 是 arg_0(**還沒 +1**;原版進來第一件事就是 `inc`)。
	radius2 int
	// dstY / dstX 是 arg_C / arg_10:視窗座標寫進 dst 時要加的位移。
	dstY, dstX int
	// sentinel 對應 arg_C == −0x6F 那個特例。
	//
	//	true(玩家)  走訪判斷看 dst 那一格是不是 0xFF;半徑外仍可能顯示
	//	false(光源)走訪判斷看另一張 visited 表;**半徑外一律不照亮**
	//
	// ★ 「半徑是不是上限」就差在這裡。玩家那一輪不是上限(牆外靠亮度圖決定),
	// 光源那一輪是上限(照不到就是照不到)。
	sentinel bool
	// dst 是要寫的緩衝(玩家 → 罩子;光源 → 亮度圖),列距 SightStride。
	dst []byte
	// visited 是 sentinel == false 時的走訪表(原版寫死用 `byte_3F6E4`)。
	visited []byte
	// tile 取視窗座標 (0..10, 0..10) 的地形。位移由呼叫端包進來。
	tile func(x, y int) byte
	// lit 回報視窗座標那一格亮不亮(原版 `byte_3F8F4[…] != 0`)。
	// 只有 sentinel == true 會用到。
	lit func(x, y int) bool
}

// run 跑一次 flood fill(`sub_2DDB0`)。
//
//	佇列從中心 (5,5) 出發
//	取出一格 → 依 sightRing 繞它的八鄰
//	  出界、或已經走過 → 跳過
//	  平方距離 ≤ 半徑²  → 寫地形(捷徑)
//	  sentinel == false → **不寫**,而且當成 0xFF(照不到,傳播就此打住)
//	  穿不過去          → 來源亮著、而且兩格都亮 → 寫地形;否則寫 0xFF
//	  穿得過            → 亮著就寫地形,否則寫 0(黑)
//	  這一格的地形穿得過視線 → 入列
func (f *sightFill) run() {
	if f.radius2 <= 0 {
		return
	}
	// ⚠ 原版一開始 `inc arg_0`,判斷寫成 `dist2 < radius²+1`,
	// 等價於 `dist2 <= radius²` —— 抄成 `<` 會少掉整整一圈。
	limit := f.radius2

	const cx, cy = SightSide / 2, SightSide / 2
	f.put(cx, cy, f.tile(cx, cy))
	queue := [][2]int{{cx, cy}}

	for len(queue) > 0 {
		from := queue[0]
		queue = queue[1:]
		for _, d := range sightRing {
			x, y := from[0]+d[0], from[1]+d[1]
			if x < 0 || x >= SightSide || y < 0 || y >= SightSide {
				continue
			}
			if !f.claim(x, y) {
				continue
			}

			tile := f.tile(x, y)
			dist2 := sightDistance(x, y)
			switch {
			case dist2 <= limit:
				f.put(x, y, tile)
			case !f.sentinel:
				// 光源那一輪:超過半徑就照不到。原版連寫都不寫
				//(`loc_2E074` 只設 var_41D = 0xFF),於是下面的
				// `SightPasses(0xFF)` 為假 → 不入列 → 傳播停在半徑上。
				tile = SightHidden
			case !SightPasses(tile, dist2):
				// 擋視線的格子**自己還是看得到**(汝看得到那面牆),
				// 但來源不能是黑的,而且兩格都要亮著。
				if f.get(from[0], from[1]) != sightDark &&
					f.lit(from[0], from[1]) && f.lit(x, y) {
					f.put(x, y, tile)
				} else {
					f.put(x, y, SightHidden)
					tile = SightHidden
				}
			case !f.lit(x, y):
				// 沒有光照到:黑的。
				f.put(x, y, sightDark)
			default:
				f.put(x, y, tile)
			}
			if SightPasses(tile, dist2) {
				queue = append(queue, [2]int{x, y})
			}
		}
	}
}

// sightDark 是**演算過程中**「這一格是黑的」。
//
// ⚠ 它與「還沒走過」是**兩個不同的值**,不能合併:
// 「還沒走過」是 0xFF、「黑的」是 0。牆要不要露出來的判斷讀的正是
// 來源那一格「是不是 0」(`cmp byte[edx+eax], 0`)—— 用 0xFF 去比會反過來。
// 跑完之後呼叫端才把 0 一律併成 0xFF。
const sightDark = 0

// index 把視窗座標換成 dst 裡的位置;超出緩衝回 −1。
//
// ⚠ **原版沒有下界檢查。** 它只擋 `arg_C + y > 0x1F`(而且是有號比較),
// 所以場景上緣附近的光源會用負的索引寫到 `byte_3F8F4` **前面**去 ——
// 一個真實的緩衝區低位越界。這裡改成擋掉:重現記憶體毀損不叫忠於原版,
// 那是未定義行為,同一份程式在不同機器上的結果都不一樣。
func (f *sightFill) index(x, y int) int {
	dy, dx := f.dstY+y, f.dstX+x
	if dy < 0 || dx < 0 || dy >= SightSceneSide || dx >= SightSceneSide {
		return -1
	}
	i := dy*SightStride + dx
	if i >= len(f.dst) {
		return -1
	}
	return i
}

func (f *sightFill) put(x, y int, v byte) {
	if i := f.index(x, y); i >= 0 {
		f.dst[i] = v
	}
}

func (f *sightFill) get(x, y int) byte {
	if i := f.index(x, y); i >= 0 {
		return f.dst[i]
	}
	return SightHidden
}

// claim 是「這一格還沒走過嗎」,走過就記下來。
//
// 兩種走訪表對應原版的兩條分支:
//
//	sentinel  看 dst 那一格是不是 0xFF —— 沒有另外記,所以**被判成看不見的牆
//	          會被別的方向再判一次**。那正是「同一面牆從亮的那側看得到、
//	          從暗的那側看不到」的來源;換成走訪表會讓結果取決於 BFS 的順序。
//	光源      另有一張 visited(原版寫死 `byte_3F6E4`),走過就填 0。
func (f *sightFill) claim(x, y int) bool {
	if f.sentinel {
		return f.get(x, y) == SightHidden
	}
	i := y*SightStride + x
	if i >= len(f.visited) || f.visited[i] == sightDark {
		return false
	}
	if f.index(x, y) < 0 {
		return false
	}
	f.visited[i] = sightDark
	return true
}

// SightWindow 是餵給 `ComputeSight` 的 11×11 地形與亮度。
type SightWindow struct {
	// Tile 取視窗座標 (x, y) 的地形。
	Tile func(x, y int) byte
	// Lit 回報視窗座標 (x, y) 亮不亮(來自 `ComputeLit`)。
	Lit func(x, y int) bool
}

// ComputeLit 掃出場景裡的光源,算出 32×32 的亮度圖(原版 `sub_2E21C`)。
//
// tile 取的是**以視窗左上角為原點**的座標 (0..31, 0..31)。
// 回傳的切片列距 `SightStride`,0 代表暗、非 0 代表亮(值是那一格的地形)。
//
// 每個光源各跑一次半徑² 10 的 flood fill,而且是**會被牆擋住**的那種 ——
// 火把照不進隔壁房間。
func ComputeLit(tile func(x, y int) byte) []byte {
	lit := make([]byte, SightSceneSide*SightStride)
	for i := range lit {
		lit[i] = SightHidden
	}
	var sources [][2]int
	for y := 0; y < SightSceneSide; y++ {
		for x := 0; x < SightSceneSide; x++ {
			t := tile(x, y)
			if !SightEmitsLight(t) {
				continue
			}
			sources = append(sources, [2]int{x, y})
			lit[y*SightStride+x] = t
		}
	}

	visited := make([]byte, SightSide*SightStride)
	half := SightSide / 2
	for _, src := range sources {
		for i := range visited {
			visited[i] = SightHidden
		}
		lx, ly := src[0], src[1]
		f := &sightFill{
			radius2:  SightLightRadius2,
			dstY:     ly - half,
			dstX:     lx - half,
			sentinel: false,
			dst:      lit,
			visited:  visited,
			tile: func(x, y int) byte {
				return tile(lx-half+x, ly-half+y)
			},
		}
		f.run()
	}

	// `sub_2E21C` 的收尾:沒照到的 0xFF `inc` 成 0。
	for i, v := range lit {
		if v == SightHidden {
			lit[i] = sightDark
		}
	}
	return lit
}

// ComputeSight 算出 11×11 的「這一格該畫什麼」(原版 `sub_2E0E8`)。
//
// radius2 是玩家自己的平方半徑(原版 `byte_3E0B5`),三種情形**完全不同**:
//
//	> 0  正常:半徑內無條件看得到,半徑外交給視線與亮度判定
//	= 0  **全黑** —— 沒有光源,連腳下都看不到
//	< 0  Wis An Ylem:整張直接攤開,牆擋不住
//
// ⚠ 中間那一條容易抄錯。原版是 `jle`(跳過 flood fill)配 `jge`(跳過全攤開)
// 兩個條件夾出來的:0 兩邊都跳過,罩子維持初值全 0xFF。抄成「半徑 0 = 只看腳下」
// 的話,無光的地方會變成能看到自己站的那一格,而原版是真的什麼都沒有。
func ComputeSight(w SightWindow, radius2 int) [SightSide * SightSide]byte {
	var out [SightSide * SightSide]byte
	for i := range out {
		out[i] = SightHidden
	}
	if radius2 < 0 {
		// Wis An Ylem:`sub_2E0E8(-1, …)` 跳過 flood fill,整張直接填地形。
		for y := 0; y < SightSide; y++ {
			for x := 0; x < SightSide; x++ {
				out[y*SightSide+x] = w.Tile(x, y)
			}
		}
		return out
	}
	if radius2 == 0 {
		return out
	}

	// 原版的罩子 `byte_3F6E4` 是 11 列 × 32 列距。
	mask := make([]byte, SightSide*SightStride)
	for i := range mask {
		mask[i] = SightHidden
	}
	f := &sightFill{
		radius2:  radius2,
		sentinel: true,
		dst:      mask,
		tile:     w.Tile,
		lit:      w.Lit,
	}
	f.run()

	// `sub_2E0E8` 的收尾:黑的(0)併成 0xFF,再攤成連續的 11×11。
	//
	// ⚠ 地形碼 0x00 的格子會被一起併掉 —— 原版就是這樣,照抄。
	for y := 0; y < SightSide; y++ {
		for x := 0; x < SightSide; x++ {
			v := mask[y*SightStride+x]
			if v == sightDark {
				v = SightHidden
			}
			out[y*SightSide+x] = v
		}
	}
	return out
}
