package game

import "github.com/wicanr2/u5-cht/internal/u5data"

// 配樂:什麼時候換第幾首(原版 32 個 `sub_3181C(n)` 呼叫點)
//
// 推導見 `docs/re/87`。`internal/game` **只決定曲號**,不碰音訊 ——
// 播放在 `internal/audio`,headless 完全不需要音效裝置(同 `render` 不綁 GPU)。
//
// ⚠ **只接有明確呼叫點的場合。** 15 首要服務整個遊戲,原作者本來就重用同一首;
// 憑「這裡該有音樂吧」補的都是自創(`CLAUDE.md §3.0`)。沒證據的留 ⬜。

const (
	// SongNone 是「還沒決定要播什麼」。原版沒有這個狀態(它開機就選好),
	// 引擎需要它,因為 `State` 可能在沒有存檔的情況下被建起來(單元測試)。
	SongNone = -1

	// SongVictory 是打贏之後(原版 `sub_A9EC` 印完 `"\nVICTORY!\n"`)。
	SongVictory = 0
	// SongBritannia 是地表大地圖(原版 `sub_86C` 的離場分支印 `"Britannia!\n"`)。
	SongBritannia = u5data.SongOverworld
	// SongShip 是船上(原版 `sub_16F08` 上船、`sub_31DC0` 開局在船上)。
	SongShip = u5data.SongShip
	// SongCombat 是戰鬥(原版 `sub_A9EC` 進迴圈時)。
	SongCombat = 3
	// SongTown 是場景裡 —— 城鎮 / 城堡 / 民居 / 要塞(原版 `sub_2D72C`)。
	SongTown = 7
	// SongDungeon 是地牢 / 洞穴 / 礦坑(原版 `sub_2D564`、`byte_3E0A3 >= 0x21`)。
	SongDungeon = u5data.SongDungeon
	// SongUnderworld 是幽冥界大地圖。
	//
	// ★★ 這一條是從 `sub_86C` 的離場分支挖出來的,而且它與地表**分開**:
	//
	//	if (byte_3E0A3 == 19h) { 印 "Underworld!\n"; 樓層 = −1; 曲 = 0Ah }
	//	else                   { 印 "Britannia!\n";  樓層 =  0; 曲 = 1   }
	//
	// ⚠ 同一首也出現在 `sub_32244`(開場 `U5INTR_E.TXT` / `SUMMONIN.TIF`)
	// ⇒ **不要把它單獨叫「幽冥界主題」**;它至少服務兩個場合。
	SongUnderworld = 0x0A
)

// CurrentSong 回報現在該播第幾首,還沒選曲時回 `SongNone`。
//
// ⚠ `find_unwired.py` 會把這一支與 `PreviousSong` 列成「只有測試引用」,
// **而那是有理由的**:它們是給 `internal/audio` 用的介面,而那個套件還不存在
// (⬜ C 階段)。已在原地寫明,免得下一輪被當成死碼刪掉
// —— 同 `u5data.WindDelay.Delay` 的處置。
func (s *State) CurrentSong() int {
	if !s.songSet {
		return SongNone
	}
	return s.song
}

// PreviousSong 回報上一首(原版 `dword_65338`)。⬜ 原版有幾個呼叫點是
// 「放完這段回到剛才那首」(`sub_165C8` / `sub_21D48` 的 `[ebp+var_8]`),
// 那幾條還沒接 —— 接的時候會用到這個值。
func (s *State) PreviousSong() int {
	if !s.songSet {
		return SongNone
	}
	return s.prevSong
}

// playSong 換曲。重複指定同一首**不算換**(原版 `sub_3181C` 會重播,但那會
// 讓引擎每回合都從頭播 —— 原版的呼叫點本來就不在每回合的路徑上)。
func (s *State) playSong(n int) {
	if s.songSet && n == s.song {
		return
	}
	s.prevSong = s.song
	s.song = n
	s.songSet = true
}

// overworldSong 是「回到大地圖該放哪一首」(原版 `sub_86C` 的離場分支)。
//
// ⚠ 判準是**地點碼**,不是樓層 —— 原版比的是 `byte_3E0A3 == 19h`。
// 用樓層判會在「地表的地下一層」出錯。
func (s *State) overworldSong() int {
	if s.Location == UnderworldLocation {
		return SongUnderworld
	}
	return SongBritannia
}

// StartupSong 在載入存檔之後決定第一首(原版 `sub_31DC0`,只跑一次)。
//
// ⚠ 存檔裡 `byte_3DDB0` 對應的欄位還沒定位 ⇒ 一律當「沒有有效曲號」,
// 走推導那條路。`docs/re/87` §3 記了這個差異與它可觀察的後果
// (存檔在城裡開局會先響大地圖的曲子)。
func (s *State) StartupSong() {
	s.playSong(u5data.StartupSong(u5data.NoStartupSong,
		int(s.Transport), s.locationCode(), s.X, s.Y))
}
