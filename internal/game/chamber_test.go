package game

import (
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 走進石室是一格一格走的(原版 `sub_134CC`)。
func TestWalkingIntoTheChamberOneStepAtATime(t *testing.T) {
	s := newState(t)
	s.MiscMaps = &u5data.MiscMapSet{}
	if !s.enterChamber(u5data.MiscMapIndexCodex) {
		t.Fatal("進不了寶典石室")
	}
	// 一進去站在入口。
	if s.X != u5data.MiscMapEnterX || s.Y != u5data.MiscMapEnterY {
		t.Fatalf("進場站在 (%d,%d),應該是入口 (%d,%d)",
			s.X, s.Y, u5data.MiscMapEnterX, u5data.MiscMapEnterY)
	}
	// 寶典要走 7 步。
	want := u5data.MiscMapWalkCodex
	steps := 0
	for {
		y := s.Y
		more := s.StepChamberWalk()
		if s.Y != y {
			steps++
		}
		// 每一步只能動一格。
		if d := y - s.Y; d != 1 && d != 0 {
			t.Fatalf("一步走了 %d 格", d)
		}
		if !more {
			break
		}
	}
	if steps != want {
		t.Errorf("走了 %d 步,原版是 %d 步", steps, want)
	}
	if s.Y != u5data.MiscMapStandY(u5data.MiscMapIndexCodex) {
		t.Errorf("走完停在 Y=%d,應該是 %d",
			s.Y, u5data.MiscMapStandY(u5data.MiscMapIndexCodex))
	}
}

// ⚠ 平手時走橫向 —— 原版 `cmp dy, dx; jle 走橫向`。
func TestDiagonalTiesGoHorizontal(t *testing.T) {
	x, y, moving := u5data.WalkStep(0, 0, 3, 3)
	if !moving || x != 1 || y != 0 {
		t.Errorf("正斜角應該先走橫的,實得 (%d,%d) moving=%v", x, y, moving)
	}
	// 縱向差距較大時才走縱向。
	x, y, _ = u5data.WalkStep(0, 0, 1, 3)
	if x != 0 || y != 1 {
		t.Errorf("dy > dx 應該走縱的,實得 (%d,%d)", x, y)
	}
	// 到了就不動。
	if _, _, moving := u5data.WalkStep(4, 4, 4, 4); moving {
		t.Error("已經在目標上不該再動")
	}
}

// 四張石室各自的步數對得上原版的常數。
func TestChamberWalkLengths(t *testing.T) {
	cases := map[int]int{
		u5data.MiscMapIndexShrine: u5data.MiscMapWalkShrine,
		u5data.MiscMapIndexCodex:  u5data.MiscMapWalkCodex,
	}
	for which, want := range cases {
		p := u5data.WalkPath(u5data.MiscMapEnterX, u5data.MiscMapEnterY,
			u5data.MiscMapEnterX, u5data.MiscMapStandY(which))
		if len(p) != want {
			t.Errorf("第 %d 張走了 %d 步,原版是 %d", which, len(p), want)
		}
	}
}
