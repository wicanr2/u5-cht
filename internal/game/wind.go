package game

import (
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 風
//
// 風向影響兩件事:**船走得多快**,以及**海上的船會不會自己漂**。
// 兩者查的是同一張 4 朝向 × 5 風的延遲表(`docs/re/23`)。
//
// Rel Hur(改風向)是玩家唯一能主動改它的手段 —— 也是為什麼那個咒語的
// 可施法場合表**只放行野外**:風只在海上有意義(`docs/re/17`)。

// SetWind 換風向(原版 `sub_2A984`:設 `byte_3E0A2` 並把變化計時器歸零)。
func (s *State) SetWind(wind int) {
	if wind < 0 || wind >= u5data.WindCount {
		return
	}
	s.Wind = wind
	s.windTimer = 0
	s.Log("風轉成了" + u5data.WindNameZH[wind] + "。")
}

// ChangeWind 是 Rel Hur:把風轉到指定方向(原版 `sub_1CDA4` 的跳表)。
//
// 方向 → 風值的對應寫死在那張跳表裡:
//
//	北 → 風 1   南 → 風 2   西 → 風 3   東 → 風 4
func (s *State) ChangeWind(d Direction) bool {
	switch d {
	case North:
		s.SetWind(u5data.WindNorth)
	case South:
		s.SetWind(u5data.WindSouth)
	case West:
		s.SetWind(u5data.WindWest)
	default:
		s.SetWind(u5data.WindEast)
	}
	return true
}

// sailDelay 是隊伍這艘船此刻的延遲(ShipNeverMoves = 逆風動不了)。
func (s *State) sailDelay(dx, dy int) int {
	if s.WindDelay == nil {
		return 2
	}
	return s.WindDelay.Delay(u5data.ShipFacingForDirection(dx, dy), s.Wind)
}

// CanSail 回報在目前的風下,往這個方向的船這一步走不走得動。
//
// 原版 `sub_2D38` 的邏輯:延遲 4 一律不動;否則累加一個計數器,
// 計數器超過延遲才動一格然後歸零。
//
// ⇒ **逆風完全動不了**,側風要兩三拍才走一格。這就是 U5 航海的節奏,
// 也是 Rel Hur 有用的原因。
func (s *State) CanSail(dx, dy int) bool {
	if s.Wind == u5data.WindCalm {
		// 無風:原版 `sub_2D38` 在查延遲表**之前**就把它 `jz` 掉,所以不查表。
		//
		// ⚠ 但這**不等於「照走」**:揚著帆(0x20..0x23)而且無風時,
		// `sub_2CCFC` 會回傳「這一步用掉了」—— 船動不了。那個判斷在
		// `turnShipInstead`,在呼叫本函式之前就攔下來了(`docs/re/66`)。
		// 走到這裡的是收帆的船(0x24..0x27),它不受風影響。
		return true
	}
	delay := s.sailDelay(dx, dy)
	if delay >= u5data.ShipNeverMoves {
		s.Log("逆風!船動不了。")
		return false
	}
	s.windTimer++
	if s.windTimer <= delay {
		s.Log("船在頂風前進……")
		return false
	}
	s.windTimer = 0
	return true
}

// IsSailing 回報隊伍是不是正在駕船。
func (s *State) IsSailing() bool {
	k := s.Transport &^ 0x03
	return k == u5data.VehicleShip || k == u5data.VehicleSailing
}

// WindName 是給狀態列用的風向名。
func (s *State) WindName() string {
	if s.Wind < 0 || s.Wind >= u5data.WindCount {
		return u5data.WindNameZH[u5data.WindCalm]
	}
	return u5data.WindNameZH[s.Wind]
}
