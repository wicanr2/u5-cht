package game

import (
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// horseScene 在場景裡放一匹馬,回傳 (狀態, 槽號)。
func horseScene(t *testing.T, x, y int) (*State, int) {
	t.Helper()
	s := openCmdScene(t)
	objs := s.currentObjects()
	if objs == nil {
		t.Skip("這個場景沒有物件層")
	}
	// 先把馬要走的那一圈清成可通行的地板,否則測到的是地形而不是規則。
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			if !s.SetTileAt(x+dx, y+dy, u5data.OpenedDoorTile) {
				t.Skip("寫不進場景地圖")
			}
		}
	}
	slot, ok := objs.Spawn(HorseKindEast, x, y, s.Floor)
	if !ok {
		t.Skip("物件槽滿了")
	}
	return s, slot
}

// TestLooseHorseWanders —— 沒被繫住的馬會走動。
func TestLooseHorseWanders(t *testing.T) {
	x, y := 10, 10
	s, slot := horseScene(t, x, y)
	objs := s.currentObjects()
	moved := false
	for i := 0; i < 200 && !moved; i++ {
		s.wanderHorses()
		o := &objs.Objects[slot]
		if o.X != x || o.Y != y {
			moved = true
		}
	}
	if !moved {
		t.Error("兩百回合都沒動 —— 沒被繫住的馬該會走")
	}
}

// TestTiedHorseStaysPut —— ★★ 旁邊有繫馬柱或欄杆就不走。
//
// 判準是**四個鄰格**不是腳下:馬站在柱子旁邊才算被繫住。
// 寫成看腳下的話城裡的馬會全部跑掉。
func TestTiedHorseStaysPut(t *testing.T) {
	for _, tie := range []byte{TileHitchingPost, TileRail} {
		for _, d := range []struct{ dx, dy int }{{0, 1}, {1, 0}, {0, -1}, {-1, 0}} {
			x, y := 10, 10
			s, slot := horseScene(t, x, y)
			objs := s.currentObjects()
			if !s.SetTileAt(x+d.dx, y+d.dy, tie) {
				t.Skip("寫不進場景地圖")
			}
			for i := 0; i < 200; i++ {
				s.wanderHorses()
			}
			if o := &objs.Objects[slot]; o.X != x || o.Y != y {
				t.Errorf("tile 0x%02X 在 (%+d,%+d) 卻沒繫住馬:走到 (%d,%d)",
					tie, d.dx, d.dy, o.X, o.Y)
			}
		}
	}
}

// TestHorseFacingFollowsHorizontalMovesOnly —— ★ 只有左右走才換朝向。
//
//	往 +x(東)→ 0x10   往 −x(西)→ 0x11   上下走 → **朝向不變**
func TestHorseFacingFollowsHorizontalMovesOnly(t *testing.T) {
	x, y := 10, 10
	s, slot := horseScene(t, x, y)
	objs := s.currentObjects()
	sawVerticalKeepingFacing := false
	for i := 0; i < 400; i++ {
		o := &objs.Objects[slot]
		bx, by, bk := o.X, o.Y, o.Kind
		s.wanderHorses()
		switch {
		case o.X > bx:
			if o.Kind != HorseKindEast {
				t.Fatalf("往東走卻是 0x%02X,預期 0x%02X", o.Kind, HorseKindEast)
			}
		case o.X < bx:
			if o.Kind != HorseKindWest {
				t.Fatalf("往西走卻是 0x%02X,預期 0x%02X", o.Kind, HorseKindWest)
			}
		case o.Y != by:
			if o.Kind != bk {
				t.Fatalf("上下走卻換了朝向:0x%02X → 0x%02X", bk, o.Kind)
			}
			sawVerticalKeepingFacing = true
		}
		// 走遠了拉回來,免得撞到測試範圍外的地形。
		if o.X != x || o.Y != y {
			o.X, o.Y = x, y
			o.Raw[u5data.ObjX], o.Raw[u5data.ObjY] = byte(x), byte(y)
		}
	}
	if !sawVerticalKeepingFacing {
		t.Error("四百回合一次都沒有垂直移動 —— 這條規則沒被驗到")
	}
}

// TestHorseStaysInsideTheScene —— ★ 座標不環繞:出界就不走。
//
// 大地圖用 `WrapWorld`,而場景是 32×32 且原版直接比 0..31。
// 用環繞的話站在邊上的馬會瞬移到對面。
func TestHorseStaysInsideTheScene(t *testing.T) {
	// ⚠ 從 (2,2) 開場再把馬搬到 (0,0):`horseScene` 會清 ±2 圈的地形,
	// 而從 (1,1) 開場那一圈會落到負座標上、寫不進去 → 測試會**跳過**。
	s, slot := horseScene(t, 2, 2)
	objs := s.currentObjects()
	o := &objs.Objects[slot]
	o.X, o.Y = 0, 0
	o.Raw[u5data.ObjX], o.Raw[u5data.ObjY] = 0, 0
	for i := 0; i < 400; i++ {
		s.wanderHorses()
		if o.X < 0 || o.Y < 0 || o.X > SceneCoordMax || o.Y > SceneCoordMax {
			t.Fatalf("馬走到場景外 (%d,%d)", o.X, o.Y)
		}
		o.X, o.Y = 0, 0
		o.Raw[u5data.ObjX], o.Raw[u5data.ObjY] = 0, 0
	}
}

// TestOnlyHorsesWander —— 反對照:別的物件不會自己走。
//
// 少了這一條,「馬會走」與「所有物件都會走」用同一個觀察分不開,
// 而後者會讓地上的寶箱自己跑掉。
func TestOnlyHorsesWander(t *testing.T) {
	for _, kind := range []byte{u5data.ObjLockedChest, 0x0E, 0x12, 0x0F} {
		x, y := 10, 10
		s, _ := horseScene(t, x, y)
		objs := s.currentObjects()
		slot, ok := objs.Spawn(kind, x+2, y, s.Floor)
		if !ok {
			t.Skip("物件槽滿了")
		}
		for i := 0; i < 200; i++ {
			s.wanderHorses()
		}
		if o := &objs.Objects[slot]; o.X != x+2 || o.Y != y {
			t.Errorf("種類 0x%02X 也走動了:(%d,%d) —— 原版只比 `kind & 0xFE == 0x10`",
				kind, o.X, o.Y)
		}
	}
}

// TestHorseDoesNotWalkOntoAnotherObject —— 目標格有東西就不走。
func TestHorseDoesNotWalkOntoAnotherObject(t *testing.T) {
	x, y := 10, 10
	s, slot := horseScene(t, x, y)
	objs := s.currentObjects()
	// 把四周全部佔滿。
	for _, d := range []struct{ dx, dy int }{{0, 1}, {1, 0}, {0, -1}, {-1, 0}} {
		if _, ok := objs.Spawn(u5data.ObjLockedChest, x+d.dx, y+d.dy, s.Floor); !ok {
			t.Skip("物件槽不夠")
		}
	}
	for i := 0; i < 200; i++ {
		s.wanderHorses()
	}
	if o := &objs.Objects[slot]; o.X != x || o.Y != y {
		t.Errorf("四周都被佔住還是走了:(%d,%d)", o.X, o.Y)
	}
}
