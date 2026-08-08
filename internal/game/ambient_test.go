package game

import (
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// ambientState 造一片可以自由改 tile 的空地。
func ambientState(t *testing.T) *State {
	t.Helper()
	s := worldState(t)
	clearObjects(t, s)
	flatGrass(t, s, 8)
	s.resetLoadWindow()
	return s
}

// TestAmbientPicksTheNearestEmitter —— 只有**最近**的那一個出聲。
//
// 原版是 `if (d² >= 目前最小) continue`,所以掃描結束時留下的是距離最小的。
func TestAmbientPicksTheNearestEmitter(t *testing.T) {
	s := ambientState(t)
	// 瀑布放遠(4 格),噴泉放近(1 格)→ 該聽到噴泉。
	s.SetTileAt(s.X+4, s.Y, 0xD4)
	s.SetTileAt(s.X+1, s.Y, 0xD8)
	if got := s.scanAmbient(); got != u5data.AmbientFountain {
		t.Errorf("掃到 %v,預期噴泉(較近)", got)
	}
	// 把噴泉搬遠一點,瀑布就該贏。
	s.SetTileAt(s.X+1, s.Y, 5)
	s.SetTileAt(s.X+5, s.Y, 0xD8)
	if got := s.scanAmbient(); got != u5data.AmbientWaterfall {
		t.Errorf("掃到 %v,預期瀑布(較近)", got)
	}
}

// TestAmbientRangeMatchesTheDistanceSquaredLimit —— 範圍是距離平方 < 0x33。
//
// ⚠ 這**不是**方形範圍:11×11 的角落 (5,5) 是 50 < 51 剛好進得去,
// 而 (5,5) 之外的都出局。所以正對邊 5 格聽得到、斜角 5 格也聽得到,
// 但斜角 6 格(72)聽不到 —— 就算它還在 11×11 掃描框裡也不算(掃不到)。
func TestAmbientRangeMatchesTheDistanceSquaredLimit(t *testing.T) {
	cases := []struct {
		dx, dy int
		want   bool
	}{
		{5, 0, true},   // 25
		{5, 5, true},   // 50 —— 角落剛好進得去
		{0, 0, true},   // 0 —— 腳下那一格
		{6, 0, false},  // 掃描框外
		{-5, -5, true}, // 對稱
	}
	for _, c := range cases {
		s := ambientState(t)
		s.SetTileAt(s.X+c.dx, s.Y+c.dy, 0xD4)
		got := s.scanAmbient() == u5data.AmbientWaterfall
		if got != c.want {
			t.Errorf("(%+d,%+d) d²=%d:聽到 %v,預期 %v",
				c.dx, c.dy, c.dx*c.dx+c.dy*c.dy, got, c.want)
		}
	}
}

// TestClockTicksAndTocksOnDifferentPitches —— 滴與答音高不同,其餘六相靜音。
func TestClockTicksAndTocksOnDifferentPitches(t *testing.T) {
	s := ambientState(t)
	s.SetTileAt(s.X+1, s.Y, 0xFA)
	pitches := map[int]int{}
	for i := 0; i < u5data.ClockPhases; i++ {
		phase := s.clockPhase
		a := s.TickAmbient()
		if a.Kind != u5data.AmbientClock {
			t.Fatalf("相位 %d 沒掃到落地鐘", phase)
		}
		switch phase {
		case u5data.ClockTickPhase, u5data.ClockTockPhase:
			if a.SFX != u5data.SFXClock {
				t.Errorf("相位 %d 該滴答,卻是音效 %d", phase, a.SFX)
			}
			pitches[phase] = a.ClockPitch
		default:
			if a.SFX != u5data.SFXNone {
				t.Errorf("相位 %d 不該出聲,卻放了音效 %d", phase, a.SFX)
			}
		}
	}
	if pitches[u5data.ClockTickPhase] == pitches[u5data.ClockTockPhase] {
		t.Errorf("滴與答音高相同(都是 %d)—— 原版是 0xBB8 vs 0x7D0",
			pitches[u5data.ClockTickPhase])
	}
	// 八相之後回到原點。
	if s.clockPhase != 0 {
		t.Errorf("跑完 %d 相之後計時器是 %d,預期歸零", u5data.ClockPhases, s.clockPhase)
	}
}

// TestClockChimesTwelveHourStyle —— 報時是十二小時制。
func TestClockChimesTwelveHourStyle(t *testing.T) {
	cases := []struct{ hour, want int }{
		{0, 12}, {1, 1}, {11, 11}, {12, 12}, {13, 1}, {23, 11},
	}
	s := ambientState(t)
	for _, c := range cases {
		s.StartClockChime(c.hour)
		if got := s.ClockStrikesLeft(); got != c.want {
			t.Errorf("%d 點敲 %d 下,預期 %d", c.hour, got, c.want)
		}
	}
}

// TestChimeReplacesTickAndCountsDownTwoPerCycle —— 鐘響取代滴答。
//
// ⚠ 兩件事一起驗:① 報時中第 0/4 相放的是鐘不是滴答;
// ② 每個八相週期只遞減 **2**(第 0 相與第 4 相各一次)。
func TestChimeReplacesTickAndCountsDownTwoPerCycle(t *testing.T) {
	s := ambientState(t)
	s.SetTileAt(s.X+1, s.Y, 0xFA)
	s.StartClockChime(3) // 敲三下
	rang := 0
	for i := 0; i < u5data.ClockPhases; i++ {
		if a := s.TickAmbient(); a.SFX == u5data.SFXAlarm {
			rang++
		} else if a.SFX == u5data.SFXClock {
			t.Errorf("報時中第 %d 相放了滴答,原版是鐘響取代滴答", i)
		}
	}
	if rang != 2 {
		t.Errorf("一個八相週期響了 %d 下,原版是 2 下(第 0 與第 4 相)", rang)
	}
	if left := s.ClockStrikesLeft(); left != 1 {
		t.Errorf("敲三下跑完一輪還剩 %d,預期 1", left)
	}
}

// TestChimeFinishesEvenWhenYouWalkAway —— 走遠了鐘照樣敲完。
//
// ★ 原版的遞減在 switch 的**匯流處**(default),不在鐘那一支裡面
// ⇒ 附近沒有鐘也照樣倒數。這條驗的是「證據放在哪」而不是「聽起來像什麼」。
func TestChimeFinishesEvenWhenYouWalkAway(t *testing.T) {
	s := ambientState(t) // 附近沒有任何發聲物
	s.StartClockChime(12)
	for i := 0; i < u5data.ClockPhases*6; i++ {
		s.TickAmbient()
	}
	if left := s.ClockStrikesLeft(); left != 0 {
		t.Errorf("跑了六個週期(12 次遞減)還剩 %d 下,預期敲完", left)
	}
}

// TestMusicStopsNearTheInstrumentAndResumesOnlyWhenAllQuiet —— 配樂壓制。
//
// ⚠⚠ 接回去的條件是「附近**完全沒有**發聲物」,不是「不是樂器了」。
// 從樂器走到瀑布旁邊配樂還是停著 —— 這條就是驗那個差別。
func TestMusicStopsNearTheInstrumentAndResumesOnlyWhenAllQuiet(t *testing.T) {
	s := ambientState(t)
	s.playSong(SongBritannia)
	// 樂器在疊圖層 —— 用物件表放一個 tile 0x5E 的物件。
	putObject(t, s, 3, 0x5E, s.X+1, s.Y)

	a := s.TickAmbient()
	if a.Kind != u5data.AmbientMusic {
		t.Fatalf("站在樂器旁掃到 %v", a.Kind)
	}
	if !a.MusicStopped || s.CurrentSong() != SongNone {
		t.Errorf("配樂沒被停掉(曲號 %d)", s.CurrentSong())
	}
	if a.BeeperNote == 0 {
		t.Error("旋律第一步是休止?原版第一步是 1 → D4(62)")
	}
	// 走到瀑布旁邊(樂器移出範圍)—— 配樂**還是**停著。
	putObject(t, s, 3, 0x5E, s.X+9, s.Y)
	s.SetTileAt(s.X+1, s.Y, 0xD4)
	if a := s.TickAmbient(); !a.MusicStopped {
		t.Error("換成瀑布之後配樂就接回去了 —— 原版要等到完全安靜")
	}
	// 瀑布也離開 → 接回原本那一首。
	s.SetTileAt(s.X+1, s.Y, 5)
	a = s.TickAmbient()
	if a.MusicStopped {
		t.Error("附近安靜了配樂還沒接回去")
	}
	if s.CurrentSong() != SongBritannia {
		t.Errorf("接回來的是曲號 %d,預期 %d", s.CurrentSong(), SongBritannia)
	}
}

// TestBeeperMelodyMatchesTheOriginalTable —— 旋律的形狀。
//
// 驗導出性質而不是抄一遍表:長度 53、值域 0..9、第一步是 D4、最後一步也是 D4、
// 而且**結尾是下行音階**(…G4 F#4 E4 D4)。
func TestBeeperMelodyMatchesTheOriginalTable(t *testing.T) {
	if n := len(u5data.BeeperMelody); n != 53 {
		t.Errorf("旋律 %d 步,原版是 53(游標上限 0x35)", n)
	}
	for i, v := range u5data.BeeperMelody {
		if int(v) > len(u5data.BeeperScale) {
			t.Errorf("第 %d 步的值 %d 超出九個音", i, v)
		}
	}
	first := u5data.BeeperNote(u5data.BeeperMelody[0])
	last := u5data.BeeperNote(u5data.BeeperMelody[len(u5data.BeeperMelody)-1])
	if first != 62 || last != 62 {
		t.Errorf("首尾是 %d / %d,原版兩端都是 D4(62)", first, last)
	}
	// 最後四步遞減。
	tail := u5data.BeeperMelody[len(u5data.BeeperMelody)-4:]
	for i := 1; i < len(tail); i++ {
		if u5data.BeeperNote(tail[i]) >= u5data.BeeperNote(tail[i-1]) {
			t.Errorf("結尾第 %d 步沒有下行(%v)", i, tail)
		}
	}
	// 休止符回 0,越界也回 0(不 panic)。
	if u5data.BeeperNote(0) != 0 || u5data.BeeperNote(99) != 0 {
		t.Error("休止符與越界值應該都回 0")
	}
}

// TestBeeperCursorWrapsAtFiftyThree —— 游標繞回。
func TestBeeperCursorWrapsAtFiftyThree(t *testing.T) {
	s := ambientState(t)
	putObject(t, s, 3, 0x5E, s.X+1, s.Y)
	var got []byte
	for i := 0; i < len(u5data.BeeperMelody)+3; i++ {
		got = append(got, s.TickAmbient().BeeperNote)
	}
	for i := 0; i < 3; i++ {
		if got[len(u5data.BeeperMelody)+i] != got[i] {
			t.Errorf("繞回之後第 %d 步是 %d,預期與第一輪相同的 %d",
				i, got[len(u5data.BeeperMelody)+i], got[i])
		}
	}
}

// TestAmbientCenterIsFixedInCombat —— 戰鬥中掃描中心固定在 (5,5)。
//
// 原版 `cmp byte_3E0A3, 80h; jnb` → 地點碼 ≥ 0x80 就 `eax = 5` 給兩個座標。
func TestAmbientCenterIsFixedInCombat(t *testing.T) {
	s := combatState(t)
	slot, ok := s.CurrentObjects().Spawn(0xC0, s.X+1, s.Y, s.Floor)
	if !ok {
		t.Skip("放不下怪物")
	}
	if !s.BeginCombat(slot) {
		t.Skipf("開不了戰:\n%s", s.log())
	}
	cx, cy := s.ambientCenter()
	if cx != u5data.AmbientScanRadius || cy != u5data.AmbientScanRadius {
		t.Errorf("戰鬥中中心是 (%d,%d),原版固定 (%d,%d)",
			cx, cy, u5data.AmbientScanRadius, u5data.AmbientScanRadius)
	}
}

// TestWaterfallOnlyTriggersOnceUntilYouLeave —— 進入範圍才觸發一次。
func TestWaterfallOnlyTriggersOnceUntilYouLeave(t *testing.T) {
	s := ambientState(t)
	s.SetTileAt(s.X+1, s.Y, 0xD4)
	if a := s.TickAmbient(); a.SFX != u5data.SFXWaterfall {
		t.Fatalf("第一次進入範圍沒放瀑布(音效 %d)", a.SFX)
	}
	for i := 0; i < 5; i++ {
		if a := s.TickAmbient(); a.SFX != u5data.SFXNone {
			t.Errorf("第 %d 次又放了音效 %d —— 旗標沒擋住", i+2, a.SFX)
		}
	}
	// 走遠 → 旗標清掉 → 回來會再放一次。
	s.SetTileAt(s.X+1, s.Y, 5)
	s.TickAmbient()
	s.SetTileAt(s.X+1, s.Y, 0xD4)
	if a := s.TickAmbient(); a.SFX != u5data.SFXWaterfall {
		t.Errorf("走遠再回來沒有重放(音效 %d)", a.SFX)
	}
}
