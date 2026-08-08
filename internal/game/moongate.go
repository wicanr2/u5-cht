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
	// ★ A 級證據:原版 `sub_135FC`(月門過場)用索引 0x0A = MOON2 三次(`docs/re/90`)。
	s.PlaySFX(u5data.SFXMoongate)
	// 原版 `sub_DF84` 傳送完呼叫 `sub_2CBEC`(視窗重新定位)再 push 字面 1。
	s.resetLoadWindow()
	s.playSong(SongBritannia)
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
	// ⬜ 原版在這裡跑一段**阻塞動畫**:把載具碼暫時換成 0x16、在視窗中心畫月門
	// (`sub_25DE4`),然後 `byte_3E097 = 0x0F` 倒數著呼叫 `sub_265F0` 15 次。
	// 引擎沒有阻塞動畫層,所以只有結果沒有過程(`docs/re/86` §6)。

	// ★★ 踏過的月門**立刻變回草地**(原版 `mov byte ptr [eax], 5`)。
	// 夜裡下一次 `RefreshMoongateTiles` 會再把它寫回 0xDC ⇒ 看起來是
	// 「吸走你之後閉合、再張開」。順序照原版:**先關門,才判要不要傳送**。
	s.SetTileAt(s.X, s.Y, u5data.MoongateClosedTile)
	// ★★ 午夜的前十分鐘踏上月門**不會傳送**,只是把門關掉
	// (原版 `if (byte_3E08F == 0 && byte_3E091 < 0Ah) esi = 1`)。
	// ⚠ 這一條沒有訊息也沒有動畫 —— 玩家只會看到門消失而人沒動。
	// 照原樣做(`CLAUDE.md §3.0`),不要「順手」補一句提示。
	if s.Clock.Hour == 0 && s.Clock.Minute < u5data.MoongateDeadMinutes {
		return true
	}
	return s.TravelByMoongate(s.MoonPhaseNow())
}
