package game

import (
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 月門
//
// 八座月門散在不列顛尼亞,每一座就在對應城市旁邊。踏進去被送到**哪一座**,
// 由當下的月相決定 —— 而**上午看特拉梅爾、下午看費盧卡**
//(`sub_E084` 的 `cmp byte_3E08F, 0Ch`)。同一座門在中午前後通到不同地方。
//
// Vas Rel Por(大傳送門)是同一條路的手動版:問「To phase: 」讀一個 '1'..'8',
// 減 '1' 之後直接呼叫同一支 `sub_DF84`。

// MoonPhaseNow 是此刻該用的月相(0..7)。
func (s *State) MoonPhaseNow() int {
	if s.Moons == nil {
		return 0
	}
	return s.Moons.PhaseAt(s.Clock.Day, s.Clock.Hour)
}

// MoongateDest 取某個相位的目的地。
func (s *State) MoongateDest(phase int) (u5data.MoongateDest, bool) {
	if s.BaseSave == nil || phase < 0 || phase >= u5data.MoonPhaseCount {
		return u5data.MoongateDest{}, false
	}
	d := s.Moongates[phase]
	return d, d.Known()
}

// TravelByMoongate 把隊伍送到某個相位的目的地(原版 `sub_DF84`)。
func (s *State) TravelByMoongate(phase int) bool {
	d, ok := s.MoongateDest(phase)
	if !ok {
		s.Log("那道門不通。")
		return false
	}
	// 從場景裡出來的話要先離開場景(原版 `sub_1678`)。
	if s.InScene() && d.Location < u5data.DungeonLocationBase {
		s.leaveScene()
	}
	s.Dungeon = nil
	s.Location = d.Location
	s.Floor = d.Floor
	s.X, s.Y = d.X, d.Y
	if d.Location == 0 {
		s.scene = nil
		s.sceneObjects = nil
	}
	s.Log("月光把汝捲了進去……")
	return true
}

// CastGreatGate 是 Vas Rel Por(原版 `sub_1986C`)。
//
// phase 是玩家打的 '1'..'8' 減 1。⚠ **揚帆中不能用**
//(`byte_3E08C & 0xF0 == 0x20` 直接失敗)。
func (s *State) CastGreatGate(phase int) bool {
	if s.Transport&0xF0 == u5data.VehicleSailing {
		s.Log("揚帆中不行!")
		return false
	}
	if phase < 0 || phase >= u5data.MoonPhaseCount {
		s.Log("沒有那個相位。")
		return false
	}
	return s.TravelByMoongate(phase)
}

// ⚠ 這裡原本有一支 `MoongateAt(x, y)`。`EnterMoongateHere` 改成讀 tile 之後
// 它就沒有任何非測試呼叫者了 —— 而「座標是不是埋藏點」的資料本來就在
// `s.Moongates` 裡,不需要包一層。已刪:留著會在下一輪盤點被算成「已完成」,
// 而且它曾經是**第二個真相來源**(同時查座標與時段),那正是要拆掉的東西。

// EnterMoongateHere 踏上月門時的處理。回傳有沒有真的傳送。
//
// ★ 判準與原版一致:**只讀腳下那一格的 tile**(`sub_E084` 的
// `sub_DB10(隊伍X, 隊伍Y)` + `cmp byte ptr [eax], 0DCh`)。
// 月門存不存在那一格,是 `RefreshMoongateTiles` 寫進去的結果。
func (s *State) EnterMoongateHere() bool {
	if s.InScene() || s.InDungeon() {
		return false
	}
	if s.TileAt(s.X, s.Y) != u5data.MoongateOpenTile {
		return false
	}
	return s.TravelByMoongate(s.MoonPhaseNow())
}
