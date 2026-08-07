package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// worldState 造一個站在大地圖上的狀態。
func worldState(t *testing.T) *State {
	t.Helper()
	s := shopState(t, 0)
	s.Location, s.Floor = 0, 0
	// 找一塊走得上去的陸地站著,四鄰也是陸地 —— 下載具需要「附近有陸地」。
	for y := 0; y < 256; y++ {
		for x := 0; x < 256; x++ {
			if u5data.TileBlocksWalking(int(s.TileAt(x, y))) {
				continue
			}
			s.X, s.Y = x, y
			if s.landNearby() {
				return s
			}
		}
	}
	t.Skip("世界地圖上找不到合適的落腳點")
	return nil
}

// TestBoardHorseAddsTwo:騎上馬時載具碼是**馬的 tile + 2**。
//
// 原版 `sub_16F08` 的 `mov eax, ebx; add al, 2; mov byte_3E08C, al`。
// 這條同時解釋了通行判定裡的 `byte_3E08C & 0xFE == 0x12` ——
// 馬物件是 0x10/0x11,騎上去才變 0x12/0x13。
func TestBoardHorseAddsTwo(t *testing.T) {
	s := worldState(t)
	if _, ok := s.CurrentObjects().Spawn(u5data.TileHorse, s.X, s.Y, s.Floor); !ok {
		t.Fatal("放不下馬")
	}
	s.Board()
	if s.Transport != u5data.TileHorse+u5data.HorseToVehicle {
		t.Errorf("騎上馬後載具是 0x%02X,預期 0x%02X",
			s.Transport, u5data.TileHorse+u5data.HorseToVehicle)
	}
	if s.Transport&0xFE != u5data.TransportHorse {
		t.Errorf("載具 0x%02X 不符合通行判定的騎馬條件 0x%02X", s.Transport, u5data.TransportHorse)
	}
	// 馬從地圖上消失了。
	if _, _, ok := s.ObjectAt(s.X, s.Y); ok {
		t.Error("騎上去了,馬還留在地上")
	}
}

// TestExitHorseRestoresObject:下馬會把馬放回腳下,自己變回步行。
func TestExitHorseRestoresObject(t *testing.T) {
	s := worldState(t)
	s.CurrentObjects().Spawn(u5data.TileHorse, s.X, s.Y, s.Floor)
	s.Board()
	s.Exit()
	if !u5data.IsOnFoot(s.Transport) {
		t.Errorf("下馬後載具是 0x%02X,預期步行", s.Transport)
	}
	o, _, ok := s.ObjectAt(s.X, s.Y)
	if !ok {
		t.Fatal("下馬之後馬不見了")
	}
	if o.Tile != u5data.TileHorse {
		t.Errorf("放回去的是 0x%02X,預期馬 0x%02X", o.Tile, u5data.TileHorse)
	}
}

// TestBoardNeedsOnFoot:騎著東西不能直接換另一個(原版 sub_16DA4)。
func TestBoardNeedsOnFoot(t *testing.T) {
	s := worldState(t)
	s.CurrentObjects().Spawn(u5data.TileHorse, s.X, s.Y, s.Floor)
	s.Transport = u5data.VehicleCarpet // 已經在魔毯上
	s.Board()
	if s.Transport != u5data.VehicleCarpet {
		t.Errorf("在魔毯上還騎得上馬,載具變成 0x%02X", s.Transport)
	}
	if !strings.Contains(s.log(), "步行") {
		t.Errorf("沒有提示要先下來:\n%s", s.log())
	}
}

// TestHorseCannotBoardShip:騎著馬不能上船。
//
// 原版 `sub_16DC8` 的跳表只放行魔毯(0x14/0x15)、步行(0x1C/0x1D)、
// 小艇(0x28..0x2B)—— 騎馬那幾格是 default(拒絕)。
func TestHorseCannotBoardShip(t *testing.T) {
	s := worldState(t)
	s.CurrentObjects().Spawn(u5data.VehicleShip, s.X, s.Y, s.Floor)
	s.Transport = u5data.TileHorse + u5data.HorseToVehicle
	s.Board()
	if s.Transport != u5data.TileHorse+u5data.HorseToVehicle {
		t.Errorf("騎著馬上了船,載具變成 0x%02X", s.Transport)
	}
}

// TestShipHullAndSkiffsCarryOver:船的耐久與小艇數在上下船之間要保住。
//
// 原版把兩個值暫存在物件槽 0 的 +5 / +7,下船時再搬回新生成的船物件。
// 耐久低於 10 會警告。
func TestShipHullAndSkiffsCarryOver(t *testing.T) {
	s := worldState(t)
	slot, _ := s.CurrentObjects().Spawn(u5data.VehicleShip, s.X, s.Y, s.Floor)
	s.CurrentObjects().Objects[slot].SetHull(7)
	s.CurrentObjects().Objects[slot].SetSkiffs(2)
	s.Board()
	if s.ShipHull != 7 || s.ShipSkiffs != 2 {
		t.Errorf("上船後耐久 %d 小艇 %d,預期 7 / 2", s.ShipHull, s.ShipSkiffs)
	}
	if !strings.Contains(s.log(), "受損嚴重") {
		t.Errorf("耐久 7 沒有警告:\n%s", s.log())
	}
	s.Exit()
	o, _, ok := s.ObjectAt(s.X, s.Y)
	if !ok {
		t.Fatal("下船後船不見了")
	}
	if o.Hull() != 7 || o.Skiffs() != 2 {
		t.Errorf("放回去的船耐久 %d 小艇 %d,預期 7 / 2", o.Hull(), o.Skiffs())
	}
}

// TestExitShipFallsBackToSkiff:海上沒有陸地時,下船會改成放小艇。
//
// 原版依序找:附近有陸地 → 上岸;沒有就用船上的小艇;連小艇都沒有
// 但有魔毯 → 騎魔毯;三者皆無才拒絕。
func TestExitShipFallsBackToSkiff(t *testing.T) {
	s := worldState(t)
	// 找一塊四鄰全是水的海面。
	found := false
	for y := 0; y < 256 && !found; y++ {
		for x := 0; x < 256; x++ {
			s.X, s.Y = x, y
			if u5data.TileBlocksWalking(int(s.TileAt(x, y))) && !s.landNearby() {
				found = true
				break
			}
		}
	}
	if !found {
		t.Skip("找不到遠離陸地的海面")
	}
	s.Transport = u5data.VehicleShip
	s.ShipSkiffs = 1
	s.Exit()
	if s.Transport&0xFC != u5data.VehicleSkiff {
		t.Errorf("海上下船後載具是 0x%02X,預期小艇", s.Transport)
	}
	if s.ShipSkiffs != 0 {
		t.Errorf("放了小艇但數量還是 %d", s.ShipSkiffs)
	}

	// 沒有小艇也沒有魔毯就只能待在船上。
	s.Transport = u5data.VehicleShip
	s.ShipSkiffs, s.Inventory.Carpets = 0, 0
	s.Exit()
	if s.Transport != u5data.VehicleShip {
		t.Errorf("沒有小艇卻下了船,載具 0x%02X", s.Transport)
	}
	if !strings.Contains(s.log(), "沒有小艇") {
		t.Errorf("沒有拒絕訊息:\n%s", s.log())
	}
}

// TestExitOnFootSaysWhat:已經在走路時按 X 會被回「此為何意」。
func TestExitOnFootSaysWhat(t *testing.T) {
	s := worldState(t)
	s.Transport = u5data.VehicleWalk
	s.Exit()
	if !strings.Contains(s.log(), "此為何意") {
		t.Errorf("走路時按 X 沒有回應:\n%s", s.log())
	}
}
