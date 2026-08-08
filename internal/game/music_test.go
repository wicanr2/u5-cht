package game

import (
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// TestFreshStateHasNoSong —— ★ 零值不能看起來像「正在播勝利曲」。
//
// 曲號 0 是勝利那一首,而 `State` 到處被寫成結構常值(沒有建構子)。
// 這條測試就是 `song` / `songSet` 不匯出的理由 —— 匯出的話這個陷阱一定會踩。
func TestFreshStateHasNoSong(t *testing.T) {
	var s State
	if got := s.CurrentSong(); got != SongNone {
		t.Errorf("空的 State 回曲號 %d,預期 %d(未定)", got, SongNone)
	}
	if got := s.PreviousSong(); got != SongNone {
		t.Errorf("空的 State 的前一首是 %d,預期 %d", got, SongNone)
	}
}

// TestPlaySongRemembersThePrevious —— `dword_65334` / `dword_65338`。
func TestPlaySongRemembersThePrevious(t *testing.T) {
	var s State
	s.playSong(SongTown)
	s.playSong(SongCombat)
	if got := s.CurrentSong(); got != SongCombat {
		t.Errorf("現在是 %d,預期 %d", got, SongCombat)
	}
	if got := s.PreviousSong(); got != SongTown {
		t.Errorf("前一首是 %d,預期 %d", got, SongTown)
	}
	// 重複指定同一首不算換 —— 否則每回合都會把前一首覆蓋掉。
	s.playSong(SongCombat)
	if got := s.PreviousSong(); got != SongTown {
		t.Errorf("重複指定同一首之後前一首變成 %d,預期還是 %d", got, SongTown)
	}
}

// TestEnteringAndLeavingASceneSwitchesSongs —— 進場景 7、出來 1。
func TestEnteringAndLeavingASceneSwitchesSongs(t *testing.T) {
	s := overworldScene(t)
	if s.BaseSave == nil {
		t.Skip("沒有底稿存檔")
	}
	// 站到一座城的座標上再進去。
	loc := u5data.Locations[1] // BRITAIN
	s.Location, s.Floor = 0, 0
	s.X, s.Y = loc.X, loc.Y
	s.Enter()
	if !s.InScene() {
		t.Fatalf("進不去%s:\n%s", loc.Name, s.log())
	}
	if got := s.CurrentSong(); got != SongTown {
		t.Errorf("進城之後曲號 %d,預期 %d(原版 `sub_2D72C`)", got, SongTown)
	}
	s.leaveScene()
	if got := s.CurrentSong(); got != SongBritannia {
		t.Errorf("出城之後曲號 %d,預期 %d", got, SongBritannia)
	}
}

// TestLeavingTheUnderworldSceneUsesItsOwnSong —— ★★ 幽冥界另有一首。
//
// 原版 `sub_86C` 的離場分支:`byte_3E0A3 == 19h` 印 "Underworld!" 配曲 0x0A,
// 否則印 "Britannia!" 配曲 1。⚠ 判準是**地點碼**,不是樓層。
func TestLeavingTheUnderworldSceneUsesItsOwnSong(t *testing.T) {
	s := overworldScene(t)
	if s.BaseSave == nil {
		t.Skip("沒有底稿存檔")
	}
	s.Location = UnderworldLocation
	s.leaveScene()
	if got := s.CurrentSong(); got != SongUnderworld {
		t.Errorf("從幽冥界出來曲號 %d,預期 %d", got, SongUnderworld)
	}
	if SongUnderworld == SongBritannia {
		t.Error("幽冥界與地表用了同一個常數 —— 原版是兩首")
	}
}

// TestCombatAndVictorySwitchSongs —— 戰鬥 3、打贏 0。
func TestCombatAndVictorySwitchSongs(t *testing.T) {
	s := overworldScene(t)
	if s.BaseSave == nil {
		t.Skip("沒有底稿存檔")
	}
	clearObjects(t, s)
	flatGrass(t, s, 8)
	putObject(t, s, 5, 0xC0, s.X+1, s.Y) // Orc 貼著隊伍
	if !s.BeginCombat(5) {
		t.Fatalf("開不了戰:\n%s", s.log())
	}
	if got := s.CurrentSong(); got != SongCombat {
		t.Errorf("戰鬥中曲號 %d,預期 %d", got, SongCombat)
	}
	// 把敵人全部標死再判勝負(用 `UnitDead`,與原版同一個位元)。
	c := s.Combat
	for i := range c.Units {
		if u := &c.Units[i]; u.Active() && s.hostile(u) {
			u.Flags |= UnitDead
		}
	}
	s.checkCombatOver()
	if got := s.CurrentSong(); got != SongVictory {
		t.Errorf("打贏之後曲號 %d,預期 %d", got, SongVictory)
	}
}

// TestBoardingAVesselSwitchesToTheShipSong —— 上船與上小艇同一首。
//
// 原版 `sub_16F08` 印 "skiff\n" 與 "Ship\n" 各自之後都 push 2;
// **馬與魔毯不換曲**(在別的分支)。
func TestBoardingAVesselSwitchesToTheShipSong(t *testing.T) {
	s := overworldScene(t)
	if s.BaseSave == nil {
		t.Skip("沒有底稿存檔")
	}
	board := func(tile byte) int {
		clearObjects(t, s)
		flatGrass(t, s, 4)
		s.Transport = u5data.VehicleWalk
		s.playSong(SongBritannia)
		putObject(t, s, 5, tile, s.X, s.Y)
		s.Board()
		return s.CurrentSong()
	}
	for _, tile := range []byte{u5data.VehicleShip, u5data.VehicleSkiff} {
		if got := board(tile); got != SongShip {
			t.Errorf("登上 0x%02X 之後曲號 %d,預期 %d", tile, got, SongShip)
		}
	}
	// 馬不換曲 —— 原版那兩個 push 2 只在船與小艇的分支裡。
	if got := board(u5data.TileHorse); got == SongShip {
		t.Error("騎馬也換成船的曲子了 —— 原版只有船與小艇會換")
	}
}
