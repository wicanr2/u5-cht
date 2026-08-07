package game

import (
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// NPC 的執行期狀態與每回合的移動
//
// `.NPC` 檔只回答「幾點該在哪」。「這一回合 NPC 做什麼」由另外三層決定:
//
//	sub_8924(hour)   進場景時建立執行期狀態
//	sub_9690(hour)   每回合跑一次,對 32 個 NPC 各做一次
//	  ├ sub_91A4     判行為模式
//	  ├ sub_8A1C     建格子圖
//	  ├ sub_8BA0     環狀佇列的有界搜尋
//	  ├ sub_8D28     從目標回溯成 (步數, 方向) 序列,再反轉
//	  └ sub_9428     「這一格站得上去嗎」
//
// 完整說明見 docs/re/12。

// NPCMode 是 NPC 這一回合的行為模式(原版 `word_3E770[npc*16 + 0]`)。
type NPCMode int

const (
	// ModeAbsent 是這個槽沒有人。
	ModeAbsent NPCMode = 0
	// ModeIdle 是已經在該在的位置上,不用動。
	ModeIdle NPCMode = 1
	// ModeWalk 是與玩家同層、目標也在同層 —— 一步一步走過去。
	ModeWalk NPCMode = 2
	// ModeStairs 是正在走向樓梯口(原版模式 3;`sub_9358` 對它特別放行樓梯 tile)。
	ModeStairs NPCMode = 3
	// ModeUpToPlayer / ModeDownToPlayer:NPC 在別層,目標在玩家這層,要過來。
	ModeUpToPlayer   NPCMode = 4
	ModeDownToPlayer NPCMode = 5
	// ModeUpAway / ModeDownAway:NPC 在玩家這層,目標在別層,要離開。
	ModeUpAway   NPCMode = 6
	ModeDownAway NPCMode = 7
	// ModeOffscreen 是 NPC 與目標都不在玩家這一層 —— 玩家看不到,直接搬過去。
	ModeOffscreen NPCMode = 8
)

// 尋路的界限,全部照原版。
const (
	// pathQueueSize 是搜尋佇列的長度。原版是 32 格的**環狀**佇列,
	// 滿了就不再加入 —— 所以這不是完整 BFS,是有界搜尋,搜不到就是搜不到。
	pathQueueSize = 32
	// pathMaxSteps 是路徑最多幾個 byte(16 段 (步數, 方向))。
	pathMaxSteps = 32
	// retryGiveUp 是連續尋路失敗到這個數就放棄(原版 `cmp ax, 0C8h`)。
	retryGiveUp = 200
	// retryCeiling 是放棄之後計數還會再爬到這裡才歸零重來(原版 0xCC)。
	retryCeiling = 204
)

// 格子圖的編碼(原版 `dword_3EF4C` 指向的 32×32 buffer)。
const (
	cellOpen    = 0x00 // 可通行、未訪問
	cellBlocked = 0x90
	cellTarget  = 0x05
	cellStart   = 0x46
)

// 方向編碼 1..4。
//
// ⚠⚠ **原版有兩張方向表,而且互為相反。** 這是這一段最容易踩的地雷:
//
//	dir   BFS(sub_8BA0)   走路 / 回溯(sub_8EA4、sub_8D28)
//	 1    x − 1(西)        x + 1(東)
//	 2    y + 1(南)        y − 1(北)
//	 3    x + 1(東)        x − 1(西)
//	 4    y − 1(北)        y + 1(南)
//
// 兩張表混用的話,BFS 與回溯都各自正確,但走出來的路徑會往反方向跑穿牆而過
// —— 我第一版就是這樣,症狀是 NPC 沿著一條合法的路徑走進建築物裡再走出地圖。
//
// 也因為兩表相反,`sub_8D28` 尾端反轉序列時才要把方向也反向:
// 回溯是用走路那張表在「倒著走」,反轉之後方向自然要翻過來。
const (
	dirA = 1
	dirB = 2
	dirC = 3
	dirD = 4
)

// bfsStep 是搜尋用的方向表(原版 `sub_8BA0` 的跳表)。
func bfsStep(x, y, dir int) (int, int) {
	switch dir {
	case dirA:
		return x - 1, y
	case dirB:
		return x, y + 1
	case dirC:
		return x + 1, y
	case dirD:
		return x, y - 1
	}
	return x, y
}

// walkStep 是走路與回溯用的方向表(原版 `sub_8EA4` 與 `sub_8D28` 的跳表)。
func walkStep(x, y, dir int) (int, int) {
	switch dir {
	case dirA:
		return x + 1, y
	case dirB:
		return x, y - 1
	case dirC:
		return x - 1, y
	case dirD:
		return x, y + 1
	}
	return x, y
}

// pathReverse 把方向反過來:原版的公式就是 `((d + 1) & 3) + 1`。
func pathReverse(dir int) int { return ((dir + 1) & 3) + 1 }

// RuntimeNPC 是一個 NPC 的執行期狀態(原版 word_3E770 的 16 B)。
type RuntimeNPC struct {
	Mode     NPCMode
	X, Y     int
	Floor    int
	Creature byte
	Slot     int // 目前採用的排程 slot(0..2)

	// path 是 (步數, 方向) 成對的序列,pathIdx 是走到第幾對。
	path    []byte
	pathIdx int
	// retries 是尋路連續失敗次數(原版 word_3EDD4)。
	retries int
}

// initRuntimeNPCs 依當前時刻建立執行期狀態(原版 `sub_8924`)。
func (s *State) initRuntimeNPCs() {
	s.rtNPCs = nil
	if s.npcs == nil {
		return
	}
	rt := make([]RuntimeNPC, u5data.NPCsPerLocation)
	for i := range s.npcs {
		n := &s.npcs[i]
		if !n.Present() || s.removed[s.Location<<8|i] {
			continue
		}
		slot := n.Schedule.Slot(s.Clock.Hour)
		rt[i] = RuntimeNPC{
			Mode:     ModeIdle,
			X:        int(n.Schedule.X[slot]),
			Y:        int(n.Schedule.Y[slot]),
			Floor:    int(n.Schedule.Floor[slot]),
			Creature: n.Creature,
			Slot:     slot,
		}
	}
	s.rtNPCs = rt
}

// npcMode 判出一個 NPC 這一回合的模式(原版 `sub_91A4`)。
//
// 只有在排程的四個時刻之一**正好是現在**時才重新評估;其餘時候維持原模式。
// 這就是 NPC 在整點換班、而不是每回合重新決定要去哪的原因。
func (s *State) npcMode(i int) NPCMode {
	rt := &s.rtNPCs[i]
	if rt.Mode == ModeAbsent {
		return ModeAbsent
	}
	sched := &s.npcs[i].Schedule
	for t := 0; t < len(sched.Times); t++ {
		if int(sched.Times[t]) != s.Clock.Hour {
			continue
		}
		slot := sched.Slot(s.Clock.Hour)
		if slot == rt.Slot {
			return ModeIdle
		}
		rt.Slot = slot
		rt.path, rt.pathIdx, rt.retries = nil, 0, 0
		here, from, to := s.Floor, rt.Floor, int(sched.Floor[slot])
		switch {
		case here != from && here != to:
			return ModeOffscreen
		case here == from && here == to:
			return ModeWalk
		case here == from && to > here:
			return ModeUpAway
		case here == from:
			return ModeDownAway
		case from > here:
			return ModeUpToPlayer
		default:
			return ModeDownToPlayer
		}
	}
	// 已經站在目標位置上就是閒置(原版在 sub_91A4 尾端補的那一段)。
	if rt.X == int(sched.X[rt.Slot]) && rt.Y == int(sched.Y[rt.Slot]) &&
		rt.Floor == int(sched.Floor[rt.Slot]) {
		return ModeIdle
	}
	return rt.Mode
}

// advanceNPCs 讓所有 NPC 走一回合。每次時鐘前進之後呼叫。
func (s *State) advanceNPCs() {
	if !s.InScene() || s.npcs == nil {
		return
	}
	if len(s.rtNPCs) != u5data.NPCsPerLocation {
		s.initRuntimeNPCs()
	}
	for i := range s.rtNPCs {
		if i == u5data.PartySlot || s.rtNPCs[i].Mode == ModeAbsent {
			continue
		}
		// ⚠ **只有模式 ≤ 1(不存在或閒置)才重新判定。**
		// 原版 `sub_9690` 的 `cmp word ptr [ebx], 1 / jg` 就是這個條件 ——
		// 正在移動中(模式 ≥ 2)的 NPC 直接繼續走,不重判。
		// 漏了它的話,因為換班時刻整整持續一小時,NPC 會每分鐘被打回閒置,
		// 路徑算得出來卻一步都走不了。
		if s.rtNPCs[i].Mode <= ModeIdle {
			s.rtNPCs[i].Mode = s.npcMode(i)
		}
		s.stepNPC(i)
	}
}

// stepNPC 讓一個 NPC 走一步。
func (s *State) stepNPC(i int) {
	rt := &s.rtNPCs[i]
	sched := &s.npcs[i].Schedule
	tx, ty, tf := int(sched.X[rt.Slot]), int(sched.Y[rt.Slot]), int(sched.Floor[rt.Slot])

	switch rt.Mode {
	case ModeIdle:
		return
	case ModeOffscreen:
		// 玩家看不到那一層,原版直接把 NPC 放到目標位置。
		rt.X, rt.Y, rt.Floor = tx, ty, tf
		rt.Mode = ModeIdle
		return
	case ModeUpAway, ModeDownAway, ModeUpToPlayer, ModeDownToPlayer:
		// 跨樓層:原版先走到樓梯口再上下樓。樓梯口的選擇(`sub_89EC`)
		// 還沒還原,先直接換層 —— 這是本檔唯一還沒逐格對上原版的地方,
		// 而它只在 NPC 換樓層的那一刻發生。
		rt.X, rt.Y, rt.Floor = tx, ty, tf
		rt.Mode = ModeIdle
		return
	}

	// ModeWalk:同層走過去。
	if rt.X == tx && rt.Y == ty {
		rt.Mode = ModeIdle
		return
	}
	if !s.followPath(i, tx, ty) {
		s.retryPath(i, tx, ty)
	}
}

// followPath 沿著已算好的路徑走一步。回傳 false 代表沒有路徑可走。
func (s *State) followPath(i, tx, ty int) bool {
	rt := &s.rtNPCs[i]
	for rt.pathIdx+1 < len(rt.path) {
		count := int(rt.path[rt.pathIdx])
		dir := int(rt.path[rt.pathIdx+1])
		if count == 0 {
			rt.pathIdx += 2
			continue
		}
		nx, ny := walkStep(rt.X, rt.Y, dir)
		if !s.npcCanStand(i, nx, ny, tx, ty) {
			// 被擋住了:整條路徑作廢,下回合再算(原版是把失敗計數 +1)。
			rt.path, rt.pathIdx = nil, 0
			return false
		}
		rt.X, rt.Y = nx, ny
		rt.path[rt.pathIdx] = byte(count - 1)
		rt.retries = 0
		if rt.X == tx && rt.Y == ty {
			rt.Mode = ModeIdle
			rt.path, rt.pathIdx = nil, 0
		}
		return true
	}
	return false
}

// retryPath 重新尋路。原版不是每回合都算:
//
//	失敗計數 >= 200            → 放棄(計數繼續爬到 204 才歸零重來)
//	失敗計數 != 0 且 亂數 != 1 → 這回合不算
func (s *State) retryPath(i, tx, ty int) {
	rt := &s.rtNPCs[i]
	if rt.retries >= retryGiveUp {
		rt.retries++
		if rt.retries > retryCeiling {
			rt.retries = 0
		}
		return
	}
	// 原版是 `random(0, 2) == 1`;這裡用時鐘當種子,讓 headless 可重現
	// (與問候語挑句同樣的理由:隨機在測試裡只會製造雜訊)。
	if rt.retries != 0 && (s.Clock.Minute+i)%3 != 1 {
		rt.retries++
		return
	}
	if path, ok := s.findPath(i, tx, ty); ok {
		rt.path, rt.pathIdx, rt.retries = path, 0, 0
		return
	}
	rt.retries = retryGiveUp
}

// npcTerrainOK 只看座標與地形(原版 `sub_9358`)。
//
// **NPC 有自己的通行表**,與玩家那張差 89 個 tile。目標格永遠算可站 ——
// 原版 `sub_9358` 一開頭就對目標回 2,不管地形。
func (s *State) npcTerrainOK(self, x, y, tx, ty int) bool {
	if x == tx && y == ty {
		return true
	}
	if x < 0 || x >= u5data.SceneSide || y < 0 || y >= u5data.SceneSide {
		return false
	}
	tile := int(s.TileAt(x, y))
	if tile == u5data.TileStairsDown || tile == u5data.TileStairsUp {
		// 樓梯只有正在找樓梯的 NPC(模式 3)能踩。
		return self >= 0 && self < len(s.rtNPCs) && s.rtNPCs[self].Mode == ModeStairs
	}
	return !u5data.TileBlocksNPC(tile)
}

// occupiedBy 回報 (x, y) 這一格站著誰(隊伍算 PartySlot),沒有回 -1。
func (s *State) occupiedBy(self, x, y int) int {
	if x == s.X && y == s.Y {
		return u5data.PartySlot
	}
	for j := range s.rtNPCs {
		if j == self || j == u5data.PartySlot {
			continue
		}
		o := &s.rtNPCs[j]
		if o.Mode != ModeAbsent && o.Floor == s.Floor && o.X == x && o.Y == y {
			return j
		}
	}
	return -1
}

// npcCanStand 是**實際走一步**時的判定(原版 `sub_9428`:`sub_9358` 之後
// 再問一次 `sub_2B3DC`「那一格有東西嗎」)。地形要過,而且不能有人站著。
func (s *State) npcCanStand(self, x, y, tx, ty int) bool {
	return s.npcTerrainOK(self, x, y, tx, ty) && s.occupiedBy(self, x, y) < 0
}

// manhattan 是原版 `sub_8F3C`。
func manhattan(ax, ay, bx, by int) int {
	dx, dy := ax-bx, ay-by
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

// nearTargetSlack 是「離目標多近的話,別人就不算障礙」(原版 `cmp eax, 4`)。
//
// 建格子圖時對其他 NPC 特別放寬:離目標 4 格以內的不擋路。
// 這讓 NPC 能規劃到目的地附近而不會被同事卡死;真的走過去時再嚴格檢查
// (`sub_9428` 那一層),走不過去就整條路徑作廢、下回合重算。
const nearTargetSlack = 4

// findPath 算出從 NPC 現在的位置到 (tx, ty) 的路徑,格式與原版相同:
// (步數, 方向) 成對,最多 16 段。
//
// 演算法照 `sub_8BA0` + `sub_8D28`:
//
//  1. 32 格**環狀佇列**的有界搜尋,從起點擴散到目標;
//     每格記下「是從哪個方向走過來的」。
//  2. **方向的嘗試順序取決於這一格是從哪來的** —— 從那個方向開始
//     依 1→2→3→4 循環,不是固定的上下左右。起點記 0x46,`>> 4` 是 4(北)。
//  3. 從目標沿著方向欄位回溯到起點,收集 (步數, 方向)。
//  4. 反轉序列,並把每個方向反向(`((d+1)&3)+1`)。
func (s *State) findPath(i, tx, ty int) ([]byte, bool) {
	rt := &s.rtNPCs[i]
	const side = u5data.SceneSide
	grid := make([]byte, side*side)
	for y := 0; y < side; y++ {
		for x := 0; x < side; x++ {
			if !s.npcTerrainOK(i, x, y, tx, ty) {
				grid[y*side+x] = cellBlocked
				continue
			}
			// 別人站的格子:離目標 4 格以上才算障礙(原版 sub_8A1C)。
			if who := s.occupiedBy(i, x, y); who >= 0 &&
				manhattan(x, y, tx, ty) >= nearTargetSlack {
				grid[y*side+x] = cellBlocked
				continue
			}
			grid[y*side+x] = cellOpen
		}
	}
	grid[ty*side+tx] = cellTarget
	grid[rt.Y*side+rt.X] = cellStart

	var qx, qy [pathQueueSize]int
	qx[0], qy[0] = rt.X, rt.Y
	head, tail := 0, 1
	found := false
	fx, fy := 0, 0

	for head != tail && !found {
		x, y := qx[head], qy[head]
		dir := int(grid[y*side+x]) >> 4
		for k := 0; k < 4 && !found; k++ {
			nx, ny := bfsStep(x, y, dir)
			if nx >= 0 && nx < side && ny >= 0 && ny < side {
				v := grid[ny*side+nx]
				if v < 0x10 { // 還沒訪問過
					grid[ny*side+nx] = byte(dir << 4)
					if v == cellTarget {
						found, fx, fy = true, nx, ny
						break
					}
					if tail != head { // 佇列沒滿才加入 —— 滿了就放棄這一格
						qx[tail], qy[tail] = nx, ny
						tail++
						if tail >= pathQueueSize {
							tail = 0
						}
					}
				}
			}
			dir = (dir & 3) + 1
		}
		head++
		if head >= pathQueueSize {
			head = 0
		}
	}
	if !found {
		return nil, false
	}
	return backtrack(grid, side, fx, fy), true
}

// backtrack 從目標沿方向欄位回到起點,產出 (步數, 方向) 序列(原版 `sub_8D28`)。
func backtrack(grid []byte, side, x, y int) []byte {
	path := make([]byte, 0, pathMaxSteps)
	cell := grid[y*side+x]
	dir := int(cell) >> 4
	low := int(cell) & 0x0F
	prev := dir
	count := 0

	for len(path) < pathMaxSteps {
		// 往回走一格。grid 記的是「從前驅到我」的**搜尋方向**,
		// 而 walkStep 那張表正好是它的反向,所以直接套 walkStep 就是往回走。
		x, y = walkStep(x, y, dir)
		atStart := low == (cellStart & 0x0F)
		if dir == prev && !atStart {
			count++
		}
		if dir != prev || atStart {
			path = append(path, byte(count), byte(prev))
			if atStart {
				break
			}
			prev = dir
			count = 1
		}
		if x < 0 || x >= side || y < 0 || y >= side {
			break
		}
		cell = grid[y*side+x]
		dir = int(cell) >> 4
		low = int(cell) & 0x0F
		if dir == 0 {
			break
		}
	}

	// 反轉:序列頭尾對調,方向同時反向。
	for i, j := 0, len(path)-2; i <= j; i, j = i+2, j-2 {
		path[i], path[j] = path[j], path[i]
		di, dj := int(path[i+1]), int(path[j+1])
		path[i+1] = byte(pathReverse(dj))
		path[j+1] = byte(pathReverse(di))
	}
	return path
}
