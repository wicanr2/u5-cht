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

// MoongateAt 回報世界座標上有沒有**開著的**月門。
//
// ✅ 「哪幾個小時、tile 是什麼」已經逆完了(`docs/re/86`,原版 `sub_DEE4`):
//
//	夜裡(20:00–04:59)→ 把 tile 0xDC 寫進**八顆月石埋藏的座標**
//	白天               → 計數器遞減,歸零才寫回草地(tile 5)
//
// ★★ 座標就是月石的埋藏點 —— **月門長在月石被埋的地方**,那是「埋月石
// 有什麼用」的答案。而 `sub_DE74` **完全不看月相**:它只查那顆月石在不在
// 當前的地點 / 樓層 / 視窗裡。
//
// ⇒ 三件事各由不同的東西決定:**開不開**看時間、**在哪裡**看月石埋在哪、
// **去哪裡**看月相(`TravelByMoongate`)。
//
// ⬜ 仍未做:把 tile 真的寫進地圖緩衝(原版每次重畫都寫),以及天亮後
// 那 0x10 次重畫的淡出。引擎用「座標 + 時段」判定,效果相同但畫面上看不到門。
func (s *State) MoongateAt(x, y int) (int, bool) {
	if s.BaseSave == nil {
		return 0, false
	}
	// ★ 白天沒有月門(原版 `sub_DEE4` 天亮之後把那一格寫回草地)。
	if !u5data.MoongateOpenAtHour(s.Clock.Hour) {
		return 0, false
	}
	for i := range s.Moongates {
		d := s.Moongates[i]
		if d.Known() && d.Location == 0 && d.X == x && d.Y == y {
			return i, true
		}
	}
	return 0, false
}

// EnterMoongateHere 踏上月門時的處理。回傳有沒有真的傳送。
func (s *State) EnterMoongateHere() bool {
	if s.InScene() || s.InDungeon() {
		return false
	}
	if _, ok := s.MoongateAt(s.X, s.Y); !ok {
		return false
	}
	return s.TravelByMoongate(s.MoonPhaseNow())
}
