package game

import (
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 上下載具(B-oard / X-it)
//
// 原版是主指令表 `jpt_2AD0D` 的 case 66('B' → `sub_16F08`)與
// case 88('X' → `sub_177AC`)。
//
// 兩支函式對稱:上載具是「腳下的物件 tile → 載具碼,物件從地圖消失」,
// 下載具是「載具碼 → 物件 tile,在腳下重新生成」。
// 換算寫在 `u5data` 的載具常數旁邊 —— 馬要 ±2,其餘同值。

// Board 是原版的 B-oard 指令。
func (s *State) Board() {
	if s.Prompt != PromptNone {
		return
	}
	// 地牢裡不能上載具(原版 `if 0x20 < byte_3E0A3 < 0x29`)。
	if s.Location > 0x20 && s.Location < 0x29 {
		s.Log(MsgNotHere)
		return
	}
	o, slot, ok := s.ObjectAt(s.X, s.Y)
	if !ok {
		s.Log(MsgWhat)
		return
	}
	tile := o.Tile
	switch {
	case tile&0xFE == u5data.TileHorse:
		if !s.mustBeOnFoot() {
			return
		}
		s.mount(slot, tile+u5data.HorseToVehicle, "汝跨上了馬。")
	case tile == u5data.TileCarpetObj:
		if !s.mustBeOnFoot() {
			return
		}
		s.mount(slot, u5data.VehicleCarpet, "汝踏上了魔毯。")
	case tile&0xFC == u5data.VehicleSkiff:
		if !s.mustBeOnFoot() {
			return
		}
		s.mount(slot, tile, "汝登上了小艇。")
	case tile&0xFC == u5data.VehicleShip:
		// 上大船的限制與其他不同:原版 `sub_16DC8` 的跳表允許
		// 魔毯(0x14/0x15)、步行(0x1C/0x1D)、小艇(0x28..0x2B)——
		// **騎著馬不能上船**。
		if !s.canBoardShip() {
			s.Log(MsgOnFoot)
			return
		}
		if h := o.Hull(); h < u5data.ShipHullWarning {
			s.Log(MsgShipDamaged)
		}
		skiffs := o.Skiffs()
		// 原本騎的東西會一起帶上船:魔毯記一筆、小艇加一艘。
		if s.Transport&0xFE == u5data.VehicleCarpet {
			s.Inventory.Carpets++
		}
		if s.Transport&0xFC == u5data.VehicleSkiff {
			skiffs++
		}
		if skiffs == 0 {
			s.Log(MsgNoSkiffs)
		}
		s.ShipHull, s.ShipSkiffs = o.Hull(), skiffs
		s.mount(slot, tile, "汝登上了船。")
	default:
		s.Log(MsgWhat)
	}
}

// mount 把物件收進來、換上載具碼。
func (s *State) mount(slot int, transport byte, msg string) {
	s.currentObjects().Remove(slot)
	s.Transport = transport
	s.Log(msg)
	s.tick()
}

// mustBeOnFoot 檢查「先下來走路」。原版 `sub_16DA4` 印的是 "On foot"。
func (s *State) mustBeOnFoot() bool {
	if u5data.IsOnFoot(s.Transport) {
		return true
	}
	s.Log(MsgOnFoot)
	return false
}

// canBoardShip 回報現在騎的東西能不能直接上船(原版 sub_16DC8 的跳表)。
func (s *State) canBoardShip() bool {
	switch u5data.VehicleKind(s.Transport) {
	case u5data.VehicleCarpet, u5data.VehicleWalk, u5data.VehicleSkiff:
		return true
	}
	return false
}

// Exit 是原版的 X-it 指令:從載具上下來。
func (s *State) Exit() {
	if s.Prompt != PromptNone {
		return
	}
	switch u5data.VehicleKind(s.Transport) {
	case u5data.TileHorse: // 0x10:騎馬(載具碼 0x12/0x13,&^3 之後是 0x10)
		s.dismount(s.Transport-u5data.HorseToVehicle, "汝下了馬。")
	case u5data.VehicleCarpet:
		// 魔毯要落在能站的地方:附近有陸地,而且腳下這一格步行走得過去。
		if !s.landNearby() || u5data.TileBlocksWalking(int(s.TileAt(s.X, s.Y))) {
			s.Log(MsgNoLandNearby)
			return
		}
		s.dismount(u5data.TileCarpetObj, "汝收起了魔毯。")
	case u5data.VehicleWalk:
		s.Log(MsgWhat) // 已經在走路了
	case u5data.VehicleSailing:
		s.Log(MsgUnderSail)
	case u5data.VehicleSkiff:
		if !s.landNearby() {
			s.Log(MsgNoLandNearby)
			return
		}
		s.dismount(s.Transport, "汝離開了小艇。")
	case u5data.VehicleShip:
		s.exitShip()
	default:
		s.Log(MsgWhat)
	}
}

// exitShip 下大船。旁邊沒有陸地時原版不是直接拒絕,而是依序找替代:
// 船上還有小艇就換乘小艇,沒有小艇但有魔毯就改騎魔毯,兩者都沒有才拒絕。
func (s *State) exitShip() {
	ship := s.Transport
	if s.landNearby() {
		s.dismountShip(ship, u5data.VehicleWalk, "汝上了岸。")
		return
	}
	switch {
	case s.ShipSkiffs > 0:
		s.ShipSkiffs--
		// 原版 `add byte_3E08C, 4` —— 大船碼 +4 正好是同朝向的小艇碼。
		s.Transport = ship + 4
		s.Log("汝放下小艇。")
		s.tick()
	case s.Inventory.Carpets > 0:
		s.Inventory.Carpets--
		s.Transport = u5data.VehicleCarpet
		s.Log("汝展開魔毯離船。")
		s.tick()
	default:
		s.Log(MsgNoSkiffs)
	}
}

// dismount 把載具放回腳下的地圖,自己變回步行。
func (s *State) dismount(tile byte, msg string) {
	objs := s.currentObjects()
	if objs == nil {
		s.Log(MsgNotHere)
		return
	}
	if _, ok := objs.Spawn(tile, s.X, s.Y, s.Floor); !ok {
		s.Log(MsgNoRoom)
		return
	}
	s.Transport = u5data.VehicleWalk
	s.Log(msg)
	s.tick()
}

// dismountShip 下船時要把耐久與小艇數搬回新生成的船物件
// (原版 `sub_2B6C8(..., hull, slot)` + `dword_3E470+3[slot*8] = 小艇數`)。
func (s *State) dismountShip(shipTile, transport byte, msg string) {
	objs := s.currentObjects()
	if objs == nil {
		s.Log(MsgNotHere)
		return
	}
	slot, ok := objs.Spawn(shipTile, s.X, s.Y, s.Floor)
	if !ok {
		s.Log(MsgNoRoom)
		return
	}
	objs.Objects[slot].SetHull(s.ShipHull)
	objs.Objects[slot].SetSkiffs(s.ShipSkiffs)
	s.ShipHull, s.ShipSkiffs = 0, 0
	s.Transport = transport
	s.Log(msg)
	s.tick()
}

// landNearby 回報四鄰有沒有走得上去的地(原版 `sub_16E58` 查視窗的
// (4,5)(6,5)(5,4)(5,6) 四格,也就是玩家的上下左右)。
func (s *State) landNearby() bool {
	for _, d := range []Direction{North, East, South, West} {
		dx, dy := d.Delta()
		if !u5data.TileBlocksWalking(int(s.TileAt(s.X+dx, s.Y+dy))) {
			return true
		}
	}
	return false
}
