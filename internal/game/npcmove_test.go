package game

import (
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// TestPathReverseFormula:方向反轉就是 `((d+1)&3)+1`(原版 sub_8D28 的尾段)。
//
// 順帶鎖住「BFS 與走路是兩張相反的方向表」—— 混用會讓路徑穿牆。
func TestPathReverseFormula(t *testing.T) {
	for _, c := range [][2]int{{dirA, dirC}, {dirB, dirD}, {dirC, dirA}, {dirD, dirB}} {
		if got := pathReverse(c[0]); got != c[1] {
			t.Errorf("pathReverse(%d) = %d,預期 %d", c[0], got, c[1])
		}
		if pathReverse(pathReverse(c[0])) != c[0] {
			t.Errorf("方向 %d 反兩次沒回到原點", c[0])
		}
	}
	// 兩張表必須互為相反:bfsStep 走一步再用同一個 dir 的 walkStep 走回來,
	// 應該回到原地。
	for d := 1; d <= 4; d++ {
		x, y := bfsStep(10, 10, d)
		bx, by := walkStep(x, y, d)
		if bx != 10 || by != 10 {
			t.Errorf("方向 %d:bfsStep 後 walkStep 沒回到原地,到了 (%d,%d)", d, bx, by)
		}
	}
}

// TestNPCUsesOwnBlockTable:NPC 的通行表與玩家的不是同一張。
//
// 最明顯的是 tile 16..25(馬與各種載具):玩家站得上去,NPC 不行。
// 拿玩家那張表給 NPC 用,NPC 就會走到船上與馬背上。
func TestNPCUsesOwnBlockTable(t *testing.T) {
	diff := 0
	for tile := 0; tile < u5data.TileCount; tile++ {
		if u5data.TileBlocksNPC(tile) != u5data.TileBlocksWalking(tile) {
			diff++
		}
	}
	if diff != 89 {
		t.Errorf("兩張通行表差 %d 個 tile,預期 89", diff)
	}
	for tile := 16; tile <= 25; tile++ {
		if !u5data.TileBlocksNPC(tile) {
			t.Errorf("tile %d(載具)不該讓 NPC 站上去", tile)
		}
	}
	if u5data.TileBlocksWalking(u5data.TileHorse) {
		t.Error("玩家該站得上馬那一格")
	}
}

// TestNPCsWalkNotTeleport:NPC 換班之後是一步一步走,不是瞬移。
//
// 驗收方式:讓時鐘跳到某個換班時刻,然後一回合一回合推進,
// 檢查每一步的位移都是 1 格(曼哈頓距離),而不是一次跳到定位。
func TestNPCsWalkNotTeleport(t *testing.T) {
	s := shopState(t, 0)
	if err := s.SetScene(britain, 0, 15, 30); err != nil {
		t.Fatal(err)
	}
	// 找一個接下來會換班、而且新舊位置都在這一層的 NPC。
	target, hour := -1, 0
	for i := range s.npcs {
		if i == u5data.PartySlot || !s.npcs[i].Present() {
			continue
		}
		sc := &s.npcs[i].Schedule
		for _, h := range sc.Times {
			a := sc.Slot(int(h))
			b := sc.Slot((int(h) + 23) % 24)
			if a == b || sc.Floor[a] != 0 || sc.Floor[b] != 0 {
				continue
			}
			if sc.X[a] == sc.X[b] && sc.Y[a] == sc.Y[b] {
				continue
			}
			target, hour = i, int(h)
			break
		}
		if target >= 0 {
			break
		}
	}
	if target < 0 {
		t.Skip("不列顛城沒有在同一層換班的 NPC")
	}

	s.Clock.Hour, s.Clock.Minute = (hour+23)%24, 59
	s.initRuntimeNPCs()
	prevX, prevY := s.rtNPCs[target].X, s.rtNPCs[target].Y
	moved, jumped := 0, 0
	for turn := 0; turn < 60; turn++ {
		s.tick()
		rt := &s.rtNPCs[target]
		d := abs(rt.X-prevX) + abs(rt.Y-prevY)
		switch {
		case d == 0:
		case d == 1:
			moved++
		default:
			jumped++
		}
		prevX, prevY = rt.X, rt.Y
	}
	if jumped > 0 {
		t.Errorf("NPC %d 有 %d 次一口氣跳超過一格", target, jumped)
	}
	if moved == 0 {
		t.Errorf("NPC %d 換班後完全沒動", target)
	}
}

// TestNPCsStayOnWalkableTiles:NPC 走過的每一格都必須是它能站的。
//
// 唯一的例外是**排程位置本身** —— 原版 `sub_9358` 一開頭就對目標格回 2,
// 不看地形。床、椅子這類 tile 擋一般通行,但 NPC 的崗位就在上面。
func TestNPCsStayOnWalkableTiles(t *testing.T) {
	s := shopState(t, 0)
	if err := s.SetScene(britain, 0, 15, 30); err != nil {
		t.Fatal(err)
	}
	exempt := 0
	for turn := 0; turn < 24*60; turn += 7 {
		s.Clock.Advance(7)
		s.advanceNPCs()
		for _, v := range s.VisibleNPCs() {
			tile := int(s.TileAt(v.X, v.Y))
			if !u5data.TileBlocksNPC(tile) {
				continue
			}
			// 「崗位」要比對**三個 slot 的任一個** —— 換班的那一刻 rt.Slot
			// 已經更新成新崗位,但 NPC 還站在舊崗位上(剛離開床鋪、還沒走第一步)。
			sc := &s.npcs[v.Index].Schedule
			onPost := false
			for k := 0; k < 3; k++ {
				if v.X == int(sc.X[k]) && v.Y == int(sc.Y[k]) {
					onPost = true
					break
				}
			}
			if onPost {
				exempt++
				continue
			}
			t.Fatalf("NPC %d 站在擋路的 tile %d 上 (%d,%d),而且那不是它的崗位,時刻 %02d:%02d",
				v.Index, tile, v.X, v.Y, s.Clock.Hour, s.Clock.Minute)
		}
	}
	if exempt == 0 {
		t.Log("(這一天沒有 NPC 站上擋路 tile 的崗位)")
	}
}

// TestNPCsDoNotOverlap:兩個 NPC 不會站在同一格。
func TestNPCsDoNotOverlap(t *testing.T) {
	s := shopState(t, 0)
	if err := s.SetScene(britain, 0, 15, 30); err != nil {
		t.Fatal(err)
	}
	for turn := 0; turn < 24*60; turn += 11 {
		s.Clock.Advance(11)
		s.advanceNPCs()
		seen := map[[2]int]int{}
		for _, v := range s.VisibleNPCs() {
			k := [2]int{v.X, v.Y}
			if prev, dup := seen[k]; dup {
				t.Fatalf("NPC %d 與 %d 都站在 (%d,%d),時刻 %02d:%02d",
					prev, v.Index, v.X, v.Y, s.Clock.Hour, s.Clock.Minute)
			}
			seen[k] = v.Index
		}
	}
}
