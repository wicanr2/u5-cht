package game

import (
	"strings"
	"testing"
)

// TestPushMovesFurnitureAndTheParty:推完之後隊伍會跟著走一步。
//
// 原版 `loc_1830F` 在推與拉兩個分支**匯流之後**才加座標 —— 兩種都會走。
// 寫成「只有推才走」的話,拉完隊伍留在原地,而家具就疊在隊伍腳下。
func TestPushMovesFurnitureAndTheParty(t *testing.T) {
	s := pushScene(t)
	x, y := s.X, s.Y
	s.SetTileAt(x+1, y, 0xA6)             // 橡木桶
	s.SetTileAt(x+2, y, pushFloorDefault) // 前面是地板 → 推得動

	s.pushToward(1, 0)
	if got := s.TileAt(x+2, y); got != 0xA6 {
		t.Errorf("桶子沒有前進,前面那格是 0x%02X", got)
	}
	if got := s.TileAt(x+1, y); got != pushFloorDefault {
		t.Errorf("原地沒有變地板,是 0x%02X", got)
	}
	if s.X != x+1 || s.Y != y {
		t.Errorf("隊伍在 (%d,%d),應該跟著走到 (%d,%d)", s.X, s.Y, x+1, y)
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgPushed) {
		t.Errorf("沒印出「%s」:%q", MsgPushed, s.Messages)
	}
}

// TestPushFallsBackToPull:推不動就改拉,而且兩者交換位置。
//
// 這**不是另一個指令** —— 原版是同一支函式的 else 分支。
// 寫成「推不動就算了」會少掉半個機制,而且玩家永遠搬不動靠牆的家具。
func TestPushFallsBackToPull(t *testing.T) {
	s := pushScene(t)
	x, y := s.X, s.Y
	s.SetTileAt(x+1, y, 0xA6)             // 桶子
	s.SetTileAt(x+2, y, 0x3E)             // 前面是牆 → 推不動
	s.SetTileAt(x, y, pushFloorDefault)   // 但腳下是地板 → 拉得動

	s.pushToward(1, 0)
	if got := s.TileAt(x, y); got != 0xA6 {
		t.Errorf("桶子沒有被拉到隊伍原處,那格是 0x%02X", got)
	}
	if got := s.TileAt(x+1, y); got != pushFloorDefault {
		t.Errorf("桶子原處沒有變地板,是 0x%02X", got)
	}
	if s.X != x+1 {
		t.Errorf("拉完隊伍該走到 %d,實際在 %d", x+1, s.X)
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgPulled) {
		t.Errorf("沒印出「%s」:%q", MsgPulled, s.Messages)
	}
}

// TestChairsAndCannonsTurn:椅子與大砲會轉向,其餘家具不會。
//
// 朝向 北 +0 / 東 +1 / 南 +2 / 西 +3(原版 `sub_18028`),
// **拉的時候 ^2**。少了轉向的話椅子推一推會朝著奇怪的方向,
// 而那看起來只像美術瑕疵,不像 bug。
func TestChairsAndCannonsTurn(t *testing.T) {
	s := pushScene(t)
	cases := []struct {
		tile   byte
		dx, dy int
		pull   bool
		want   byte
	}{
		{0x90, 1, 0, false, 0x91},  // 椅子往東 → +1
		{0x90, 0, 1, false, 0x92},  // 往南 → +2
		{0x90, -1, 0, false, 0x93}, // 往西 → +3
		{0x90, 0, -1, false, 0x90}, // 往北 → +0
		{0x91, 1, 0, true, 0x93},   // 拉:東 +1 再 ^2 = 3
		{0xB4, 0, 1, false, 0xB6},  // 大砲往南
		{0xA6, 1, 0, false, 0xA6},  // 桶子不轉向
		{0xAF, 0, 1, true, 0xAF},   // 置物箱也不轉
	}
	for _, c := range cases {
		if got := s.pushFacing(c.tile, c.dx, c.dy, c.pull); got != c.want {
			t.Errorf("0x%02X 往 (%d,%d) pull=%v → 0x%02X,預期 0x%02X",
				c.tile, c.dx, c.dy, c.pull, got, c.want)
		}
	}
}

// TestCannonsNeedTheirOwnFloor:大砲要 0x45,家具要 0x44,兩種不能互換。
//
// 原版 `(v8 & 0xFC) == 0xB4 ? 69 : 68`,而目的地必須「正好是」那一種。
// 兩種都是 cobble,肉眼分不出來 —— 只有比對數值才看得見這條規則。
func TestCannonsNeedTheirOwnFloor(t *testing.T) {
	s := pushScene(t)
	x, y := s.X, s.Y
	s.SetTileAt(x+1, y, 0xB4)             // 大砲
	s.SetTileAt(x+2, y, pushFloorDefault) // 給錯地板(0x44)
	s.SetTileAt(x, y, 0x3E)               // 腳下是牆 → 也拉不動

	s.pushToward(1, 0)
	if got := s.TileAt(x+1, y); got != 0xB4 {
		t.Errorf("大砲不該動,那格變成 0x%02X", got)
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgWontBudge) {
		t.Errorf("沒印出「%s」:%q", MsgWontBudge, s.Messages)
	}
}

// TestUnpushableStaysPut:不在清單上的東西一概推不動。
func TestUnpushableStaysPut(t *testing.T) {
	s := pushScene(t)
	x, y := s.X, s.Y
	s.SetTileAt(x+1, y, 0x3E) // 牆
	s.SetTileAt(x+2, y, pushFloorDefault)
	s.pushToward(1, 0)
	if s.X != x {
		t.Errorf("推不動卻走了一步,隊伍到了 %d", s.X)
	}
	if got := s.TileAt(x+1, y); got != 0x3E {
		t.Errorf("牆被推走了,那格是 0x%02X", got)
	}
}

func pushScene(t *testing.T) *State {
	t.Helper()
	s := newState(t)
	s.MaxMessages = 16
	s.Location = 1
	s.Floor = 0
	s.X, s.Y = 10, 10
	if err := s.SetScene(1, 0, 10, 10); err != nil {
		t.Skipf("合成場景不可用:%v", err)
	}
	s.X, s.Y = 10, 10
	return s
}
