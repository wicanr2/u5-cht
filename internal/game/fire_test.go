package game

import (
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// TestFireOnlyBroadsides:船上只打得到舷側。
//
// 原版 `sub_17120` 的「Fire broadsides only!」。少了這道限制,玩家可以往
// 船首正前方轟 —— 而那是帆船完全辦不到的事,也讓海戰的走位變得沒意義。
func TestFireOnlyBroadsides(t *testing.T) {
	s := fireScene(t)
	// 船首朝北(載具碼低兩位 = 0)。
	s.Transport = u5data.VehicleShip
	for _, c := range []struct {
		d    Direction
		want bool
	}{
		{North, false}, // 船首方向
		{South, false}, // 船尾方向
		{East, true},   // 舷側
		{West, true},   // 舷側
	} {
		if got := s.isBroadside(c.d); got != c.want {
			t.Errorf("船首朝北、往%s開砲:isBroadside=%v,預期 %v", c.d.Name(), got, c.want)
		}
	}
	// 船首朝東 → 南北才是舷側。
	s.Transport = u5data.VehicleShip + 1
	if !s.isBroadside(North) || s.isBroadside(East) {
		t.Error("船首朝東時,舷側該是南北")
	}
}

// TestFireFromShipRejectsTheBow:往船首開砲會被擋下並說明原因。
func TestFireFromShipRejectsTheBow(t *testing.T) {
	s := fireScene(t)
	s.Transport = u5data.VehicleShip // 朝北
	s.Messages = nil
	s.fireToward(North)
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgBroadsidesOnly) {
		t.Errorf("沒說只能打舷側:%q", s.Messages)
	}
	s.Messages = nil
	s.fireToward(East)
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgBooom) {
		t.Errorf("舷側該打得出去:%q", s.Messages)
	}
}

// TestFireOnLandNeedsAnAdjacentCannon:陸上要緊鄰大砲。
//
// 原版查的是視野緩衝裡玩家四鄰那四格是不是 0xB4。走路時對著空地按 F
// 只會得到「What?」。
func TestFireOnLandNeedsAnAdjacentCannon(t *testing.T) {
	s := fireScene(t)
	s.Transport = u5data.VehicleWalk
	x, y := s.X, s.Y

	s.Messages = nil
	s.fireToward(East)
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgWhat) {
		t.Errorf("旁邊沒砲卻開得出來:%q", s.Messages)
	}

	s.SetTileAt(x+1, y, 0xB4)
	s.Messages = nil
	s.fireToward(East)
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgBooom) {
		t.Errorf("旁邊有砲卻開不出來:%q", s.Messages)
	}
	// 四個朝向的大砲都算(0xB4..0xB7 是同一群)。
	for _, tile := range []byte{0xB4, 0xB5, 0xB6, 0xB7} {
		s.SetTileAt(x+1, y, tile)
		if !s.cannonBeside(East) {
			t.Errorf("0x%02X 沒被認成大砲", tile)
		}
	}
}

// TestFireNotInsideTowns:城裡不能開砲。
//
// 原版 `cmp byte_3E0A3, 0; jnz` —— 不在地表就印「What?」。
// 城裡的大砲要用 Push 推,不能開。
func TestFireNotInsideTowns(t *testing.T) {
	s := fireScene(t)
	s.Location = 1
	s.Messages = nil
	s.Fire()
	if s.AwaitingDirection() {
		t.Error("城裡竟然問了方向")
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgWhat) {
		t.Errorf("沒印出「%s」:%q", MsgWhat, s.Messages)
	}
}

func fireScene(t *testing.T) *State {
	t.Helper()
	s := newCreateState(t)
	s.MaxMessages = 16
	s.Messages = nil
	s.Location = 0
	s.Floor = 0
	// Fire 只在地表能用,所以測試需要真的世界地圖 —— `realState` 只載場景。
	// ⚠ 少了這一步 `TileAt` 全回 0、`SetTileAt` 無效,而測試的失敗訊息
	// 會長得像「大砲沒被認出來」,把人指向錯的地方。
	dir := os.Getenv("U5_GAMEDATA")
	w, err := u5data.LoadFlatMap(dir + "/UNDER.DAT")
	if err != nil {
		t.Skipf("載不到平面地圖:%v", err)
	}
	s.Under = w
	s.World = w // 地表也用同一張 —— 這條測試只在意 tile 讀寫,不在意地形長相
	return s
}
