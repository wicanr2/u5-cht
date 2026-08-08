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
	if s.Objects == nil {
		s.Objects = &u5data.ObjectSet{}
	}
	if s.X == 0 && s.Y == 0 {
		s.X, s.Y = 64, 64
	}
	return s
}

// TestCannonballDestroysDoors 是這次補上的:大地圖開砲原本什麼都不會發生。
//
// ⚠ `fireToward` 原本呼叫 `FlyProjectile`,而它一開頭就 `if s.Combat == nil { return }`
// —— 所以在此之前大地圖上的砲彈只印一句 BOOOM 就結束。
func TestCannonballDestroysDoors(t *testing.T) {
	s := fireScene(t)
	for _, door := range []byte{0x97, 0x98, 0x99, 0xB8, 0xB9, 0xBA, 0xBB} {
		s.Messages = nil
		if !s.SetTileAt(s.X+2, s.Y, door) {
			t.Skip("寫不進世界地圖")
		}
		s.fireCannonball(1, 0)
		if !strings.Contains(strings.Join(s.Messages, "|"), MsgDoorDestroyed) {
			t.Errorf("門 0x%02X 沒被轟掉:%q", door, s.Messages)
		}
		if got := s.TileAt(s.X+2, s.Y); got != u5data.TileBrickFloor {
			t.Errorf("門 0x%02X 轟完變成 0x%02X,預期磚地 0x%02X",
				door, got, u5data.TileBrickFloor)
		}
	}
	// 不是門的東西不會變。
	s.Messages = nil
	if !s.SetTileAt(s.X+2, s.Y, 0x05) { // 草地
		t.Skip("寫不進世界地圖")
	}
	s.fireCannonball(1, 0)
	if strings.Contains(strings.Join(s.Messages, "|"), MsgDoorDestroyed) {
		t.Errorf("草地被當成門轟掉了:%q", s.Messages)
	}
}

// TestCannonCannotShootBlackthornOrLordBritish 釘住那個 `& 0xF8 == 0x78`。
//
// 0x78 = 14×4+0x40(黑刺)、0x7C = 15×4+0x40(不列顛王)——
// 兩個都在 0x78..0x7F 這一組裡。
func TestCannonCannotShootBlackthornOrLordBritish(t *testing.T) {
	for kind := 0x78; kind <= 0x7F; kind++ {
		if cannonTargets(byte(kind)) {
			t.Errorf("種類 0x%02X 竟然打得掉 —— 那一組是黑刺與不列顛王", kind)
		}
	}
	// 其他生物打得掉。
	for _, kind := range []byte{0x40, 0x44, 0x80, 0xE4} {
		if !cannonTargets(kind) {
			t.Errorf("種類 0x%02X 該打得掉", kind)
		}
	}
	// ★ 馬是特例:編號比 0x1C 小卻放行。
	for _, kind := range []byte{u5data.TileHorse, u5data.TileHorse + 1} {
		if !cannonTargets(kind) {
			t.Errorf("馬(0x%02X)該打得掉 —— 原版明文放行", kind)
		}
	}
	// 小物件打不掉。
	for _, kind := range []byte{0x01, 0x0F, 0x1B} {
		if cannonTargets(kind) {
			t.Errorf("種類 0x%02X 不該是目標(< 0x1C)", kind)
		}
	}
	// ⚠ 那道 `& 0xFC == 0x2F` 是死碼 —— 船(0x2C..0x2F)照樣打得掉。
	for kind := 0x2C; kind <= 0x2F; kind++ {
		if !cannonTargets(byte(kind)) {
			t.Errorf("種類 0x%02X 打不掉 —— 那道 `& 0xFC == 0x2F` 永遠不成立,是死碼", kind)
		}
	}
}

// TestCannonHitCostsKarma:打中東西業報 −5,下限 0。
func TestCannonHitCostsKarma(t *testing.T) {
	for _, c := range []struct{ before, after int }{
		{50, 45}, {6, 1}, {5, 0}, {3, 0}, {0, 0},
	} {
		s := fireScene(t)
		objs := s.currentObjects()
		if objs == nil {
			t.Skip("沒有物件層")
		}
		slot, ok := objs.Spawn(0xE4, s.X+1, s.Y, s.Floor)
		if !ok {
			t.Skip("生不出目標")
		}
		s.Karma = c.before
		s.cannonHit(slot, s.X+1, s.Y)
		if s.Karma != c.after {
			t.Errorf("業報 %d → %d,預期 %d", c.before, s.Karma, c.after)
		}
		if objs.Objects[slot].Present() {
			t.Error("打中的東西沒有消失 —— 原版是整個槽清掉")
		}
	}
}

// TestShootingYourOwnSlotHurtsTheParty:打到槽 0 就是打自己。
func TestShootingYourOwnSlotHurtsTheParty(t *testing.T) {
	s := fireScene(t)
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		s.Roster[i].Status = u5data.StatusGood
		s.Roster[i].HP = 200
	}
	before := s.Karma
	s.cannonHit(u5data.PartyObjectSlot, s.X, s.Y)
	if s.Karma != before {
		t.Errorf("打到自己卻扣了業報:%d → %d", before, s.Karma)
	}
	hurt := false
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		if s.Roster[i].HP < 200 {
			hurt = true
		}
	}
	if !hurt {
		t.Error("打到自己卻沒人受傷")
	}
}
