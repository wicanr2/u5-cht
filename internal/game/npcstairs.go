package game

import "github.com/wicanr2/u5-cht/internal/u5data"

// NPC 換樓層要先走到樓梯口(原版 `sub_92C0` / `sub_89EC` / `sub_8BA0`)
//
// 推導見 `docs/re/100`。原版模式 6 / 7(往上 / 往下)那一段:
//
//	if (sub_92C0(npc, 排程槽))     → 已經站在對的樓梯上 → 換層
//	if (這回合已經找過)            → return
//	目標 = sub_89EC(x, y, 要往上, npc)      ; 找最近的樓梯口
//	if (目標) sub_8BA0(…)                   ; 往那裡走一步
//	if (走不動) return                      ; ★ 這一回合就不動,下回合再試
//
// 而 `sub_89EC(x, y, up, npc)` 只是 `sub_8BA0(x, y, 0, 0, up ? −1 : −2, npc)` ——
// 同一支 BFS 尋路,用 **−1 / −2 當「目標是最近的上 / 下樓梯」的特殊碼**。
//
// ⇒ 引擎此前在這裡**直接換層**(`npcmove.go` 自己標著「樓梯口的選擇還沒還原」),
// 所以玩家會看到 NPC 憑空消失、在另一層憑空出現。

// npcNeedsStairUp 回報這個 NPC 要往上還是往下(原版 `sub_92C0` 的那個比較)。
//
//	當前樓層 >  排程樓層 → 往下(找 0xC9)
//	當前樓層 <= 排程樓層 → 往上(找 0xC8)
//
// ⚠ 相等時算「往上」—— 原版是 `jle`(小於或等於就走上樓那條)。
// 實務上相等不會進這段(同層走 `ModeWalk`),但照抄比較安全。
func (s *State) npcNeedsStairUp(i int) bool {
	if i < 0 || i >= len(s.rtNPCs) {
		return true
	}
	rt := &s.rtNPCs[i]
	target := u5data.SignedFloor(s.npcs[i].Schedule.Floor[rt.Slot])
	return s.Floor <= target
}

// npcOnUsableStair 回報這個 NPC 腳下能不能通往它的排程樓層(原版 `sub_92C0`)。
//
//	要往下 → tile == 0xC9(`LadderDown`)
//	要往上 → tile == 0xC8(`LadderUp`)
//	或者   → (tile & 0xF4) == 0xC4    ★ 兩個方向都算
//
// ⚠⚠ 最後那個遮罩是 **0xF4 不是 0xFC** ⇒ 除了 0xC4..0xC7(四個朝向的樓梯)
// 還收 **0xCC..0xCF**(bit 3 沒被遮住)。`u5data.StairsFacing` 用的是 0xFC,
// 範圍比較窄 —— 兩支的定義域不同,不要互相套用。
func (s *State) npcOnUsableStair(i int) bool {
	if i < 0 || i >= len(s.rtNPCs) {
		return false
	}
	rt := &s.rtNPCs[i]
	tile := s.TileAt(rt.X, rt.Y)
	want := byte(u5data.LadderDown)
	if s.npcNeedsStairUp(i) {
		want = u5data.LadderUp
	}
	return tile == want || tile&u5data.NPCStairMask == u5data.NPCStairBase
}

// nearestNPCStair 找離這個 NPC 最近的可用樓梯口(原版 `sub_89EC`)。
//
// 原版是同一支 BFS 用特殊目標碼走出來的「最近」;引擎的 `findPath` 只吃座標,
// 所以這裡先用**曼哈頓距離**挑一格再交給 `findPath`。
//
// ⚠ 兩者不一定挑到同一格:BFS 的「最近」是**走得到的步數**,曼哈頓距離
// 不考慮牆。差別會在「隔著牆更近的那座樓梯」上顯出來 —— 已知差異,
// 不假裝一樣。要完全一致得把 `findPath` 改成支援「目標是一個謂詞」。
func (s *State) nearestNPCStair(i int) (int, int, bool) {
	if i < 0 || i >= len(s.rtNPCs) || s.scene == nil {
		return 0, 0, false
	}
	rt := &s.rtNPCs[i]
	want := byte(u5data.LadderDown)
	if s.npcNeedsStairUp(i) {
		want = u5data.LadderUp
	}
	bx, by, best := 0, 0, -1
	for y := 0; y < u5data.SceneSide; y++ {
		for x := 0; x < u5data.SceneSide; x++ {
			tile := s.TileAt(x, y)
			if tile != want && tile&u5data.NPCStairMask != u5data.NPCStairBase {
				continue
			}
			d := iabs(x-rt.X) + iabs(y-rt.Y)
			if best < 0 || d < best {
				bx, by, best = x, y, d
			}
		}
	}
	return bx, by, best >= 0
}

// stepNPCToStairs 是模式 6 / 7 的一回合:站對了就換層,否則往樓梯走一步。
//
// 回傳 false 表示「這一回合什麼都沒做」——與原版 `sub_89EC` 回 0 之後
// `return` 的行為相同(不動,下回合再試)。
//
// ⚠ **不要在找不到樓梯時直接換層。** 那是引擎此前的做法,而它把
// 「NPC 走上樓」變成「NPC 瞬移」。原版找不到就是不動。
func (s *State) stepNPCToStairs(i int) bool {
	rt := &s.rtNPCs[i]
	sched := &s.npcs[i].Schedule
	if s.npcOnUsableStair(i) {
		// 站在對的樓梯上 → 換層,並且落在排程的座標上。
		rt.X = int(sched.X[rt.Slot])
		rt.Y = int(sched.Y[rt.Slot])
		rt.Floor = u5data.SignedFloor(sched.Floor[rt.Slot])
		rt.Mode = ModeIdle
		rt.path, rt.pathIdx = nil, 0
		return true
	}
	sx, sy, ok := s.nearestNPCStair(i)
	if !ok {
		return false
	}
	// ★ 走向樓梯時要掛 `ModeStairs` —— `npcTerrainOK` 只對這個模式放行
	// 樓梯 tile,否則 NPC 永遠踩不上去(原版 `sub_9358` 的模式 3)。
	rt.Mode = ModeStairs
	if len(rt.path) == 0 {
		s.retryPath(i, sx, sy)
	}
	return s.followPath(i, sx, sy)
}
