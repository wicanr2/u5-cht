package game

import (
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 白天全亮、夜裡只剩兩格 —— 日照半徑的主幹。
func TestLightRadiusFollowsTheClock(t *testing.T) {
	cases := []struct {
		name        string
		hour, minут int
		want        int
	}{
		{"正午", 12, 0, SightRadiusDay},
		{"清晨六點", 6, 0, SightRadiusDay},
		{"傍晚六點五十九", 18, 59, SightRadiusDay},
		{"午夜", 0, 0, SightRadiusNight},
		{"凌晨四點五十九", 4, 59, SightRadiusNight},
		{"晚間八點", 20, 0, SightRadiusNight},
	}
	for _, c := range cases {
		s := newState(t)
		s.Clock.Hour, s.Clock.Minute = c.hour, c.minут
		if got := s.LightRadius2(); got != c.want {
			t.Errorf("%s:半徑² %d,預期 %d", c.name, got, c.want)
		}
	}
}

// 日出從 2 開到 49,日落反著走 —— 六段各十分鐘。
func TestDawnAndDuskRampInOppositeDirections(t *testing.T) {
	// 五點整還是夜間的亮度,五點五十分幾乎全亮。
	for i, want := range sightDawnRamp {
		s := newState(t)
		s.Clock.Hour, s.Clock.Minute = SightDawnHour, i*10
		if got := s.LightRadius2(); got != want {
			t.Errorf("5:%02d 半徑² %d,預期 %d", i*10, got, want)
		}
	}
	// 日落是鏡像:19:00 幾乎全亮、19:50 剩夜間亮度。
	for i, want := range sightDawnRamp {
		min := 59 - i*10
		s := newState(t)
		s.Clock.Hour, s.Clock.Minute = SightDuskHour, min
		if got := s.LightRadius2(); got != want {
			t.Errorf("19:%02d 半徑² %d,預期 %d", min, got, want)
		}
	}
	// 銜接處要對得上:日出第一段 = 夜、最後一段 ≈ 白天。
	if sightDawnRamp[0] != SightRadiusNight {
		t.Errorf("日出第一段是 %d,應與夜間亮度 %d 相同",
			sightDawnRamp[0], SightRadiusNight)
	}
	if sightDawnRamp[len(sightDawnRamp)-1] >= SightRadiusDay {
		t.Errorf("日出最後一段 %d 不該達到白天的 %d",
			sightDawnRamp[len(sightDawnRamp)-1], SightRadiusDay)
	}
}

// 地下與亞拉臘號殘骸不分晝夜都是夜間亮度。
func TestUndergroundAndAraratAreAlwaysDark(t *testing.T) {
	s := newState(t)
	s.Clock.Hour = 12
	s.Floor = -1
	if got := s.LightRadius2(); got != SightRadiusNight {
		t.Errorf("正午的地下層半徑² %d,應為 %d", got, SightRadiusNight)
	}

	s2 := newState(t)
	s2.Clock.Hour = 12
	s2.Location = SightAlwaysDarkLocation
	if got := s2.LightRadius2(); got != SightRadiusNight {
		t.Errorf("正午的亞拉臘號半徑² %d,應為 %d", got, SightRadiusNight)
	}
	// 那個地點就是殘骸,不是隨手挑的編號。
	if u5data.Locations[SightAlwaysDarkLocation-1].Name != "ARARAT" {
		t.Errorf("地點 %d 是 %q,不是 ARARAT —— 常數對錯了",
			SightAlwaysDarkLocation, u5data.Locations[SightAlwaysDarkLocation-1].Name)
	}
}

// 咒語比火把亮 —— 反直覺但原版如此,而且兩個都只是「下限」。
func TestInLorIsBrighterThanATorch(t *testing.T) {
	if SightRadiusInLor <= SightRadiusTorch {
		t.Fatalf("In Lor 的 %d 應該大於火把的 %d", SightRadiusInLor, SightRadiusTorch)
	}
	// 夜裡點火把。
	s := newState(t)
	s.Clock.Hour = 0
	s.TorchTurns = 100
	if got := s.LightRadius2(); got != SightRadiusTorch {
		t.Errorf("夜裡的火把半徑² %d,應為 %d", got, SightRadiusTorch)
	}
	// 夜裡唸 In Lor。
	s.TorchTurns = 0
	s.LightTurns = 100
	if got := s.LightRadius2(); got != SightRadiusInLor {
		t.Errorf("夜裡的 In Lor 半徑² %d,應為 %d", got, SightRadiusInLor)
	}
	// 兩個一起 —— 取亮的那個,不是相加。
	s.TorchTurns = 100
	if got := s.LightRadius2(); got != SightRadiusInLor {
		t.Errorf("咒語加火把半徑² %d,應該還是 %d(取亮的,不相加)",
			got, SightRadiusInLor)
	}
}

// 白天有火把也不會變暗 —— 下限是「不足才補」。
func TestLightSourcesNeverDimTheDaylight(t *testing.T) {
	s := newState(t)
	s.Clock.Hour = 12
	s.TorchTurns = 100
	s.LightTurns = 100
	if got := s.LightRadius2(); got != SightRadiusDay {
		t.Errorf("正午拿火把半徑² %d,應該還是白天的 %d", got, SightRadiusDay)
	}
}

// 白天的半徑正好蓋滿整個 11×11 —— 所以晴天在空曠處視線判定等於沒啟用。
//
// ★ 這條把兩個獨立來的數字釘在一起:日照表的 0x32 與距離表角落的
// dx²+dy² = 50。兩邊任一個抄錯都會在這裡爆掉。
func TestDaylightCoversTheWholeWindow(t *testing.T) {
	corner := 0
	for y := 0; y < u5data.SightSide; y++ {
		for x := 0; x < u5data.SightSide; x++ {
			d := (x - 5) * (x - 5)
			d += (y - 5) * (y - 5)
			if d > corner {
				corner = d
			}
		}
	}
	if SightRadiusDay != corner {
		t.Fatalf("白天半徑² 是 %d,視窗角落的平方距離是 %d —— 兩者應相等",
			SightRadiusDay, corner)
	}
}

// 戰鬥、石室不做視線遮蔽(原版 `cmp byte_3E0A3, 80h` 那一岔)。
func TestNoSightMaskInCombatOrChambers(t *testing.T) {
	s := newState(t)
	s.Clock.Hour = 0 // 夜:正常情況會有遮蔽
	if s.SightMask() == nil {
		t.Fatal("夜裡的城鎮應該有視線遮蔽")
	}

	s.MiscMaps = &u5data.MiscMapSet{}
	if !s.enterChamber(u5data.MiscMapIndexShrine) {
		t.Fatal("進不了石室")
	}
	if s.SightMask() != nil {
		t.Error("石室裡不該有視線遮蔽")
	}
}

// Wis An Ylem 期間整張視窗都看得到。
func TestWisAnYlemDefeatsTheSightMask(t *testing.T) {
	s := newState(t)
	s.Clock.Hour = 0
	s.SeeThroughWalls = true
	mask := s.SightMask()
	if mask == nil {
		t.Fatal("應該還是有罩子,只是全開")
	}
	for dy := -5; dy <= 5; dy++ {
		for dx := -5; dx <= 5; dx++ {
			if !SightVisible(mask, dx, dy) {
				t.Fatalf("(%+d,%+d) 看不到 —— Wis An Ylem 該讓整張攤開", dx, dy)
			}
		}
	}
}

// visibleCount 數這一幀看得到幾格。
func visibleCount(s *State) int {
	mask := s.SightMask()
	n := 0
	for dy := -5; dy <= 5; dy++ {
		for dx := -5; dx <= 5; dx++ {
			if SightVisible(mask, dx, dy) {
				n++
			}
		}
	}
	return n
}

// openScene 造一張全是地板的場景,玩家站在中間。
func openScene(t *testing.T, hour int) *State {
	t.Helper()
	s := &State{Scenes: synthScenes(t, walkable(t)), MaxMessages: 8}
	if err := s.SetScene(britain, 0, 15, 15); err != nil {
		t.Fatalf("進不了不列顛城:%v", err)
	}
	s.Clock.Hour = hour
	return s
}

// 空曠處日夜都看得到整個視窗。
//
// ⚠⚠ **這條的結論與直覺相反,而它就是原版的行為。** 我原本寫的測試假設
// 「夜裡看得到的格數會變少」,結果對不上 —— 回去讀組語才發現半徑根本
// 不參與傳播:`sub_2DDB0` 是否往外傳只看地形擋不擋視線
//(`sub_2E1D0`),與 `arg_0` 無關。半徑只在**判定**時當捷徑,
// 而場景內的判定另有一條「來源不是黑的就顯示」的路,一樣繞過半徑。
//
// ⚠ 尚未對 DOSBox 原版實測(見 `docs/re/31` §5)。在那之前不宣稱
// 「夜間效果已完成」—— 這裡釘住的是**程式碼確實如此**,不是「畫面對了」。
func TestOpenGroundIsFullyVisibleDayAndNight(t *testing.T) {
	full := u5data.SightSide * u5data.SightSide
	for _, hour := range []int{12, 0} {
		if got := visibleCount(openScene(t, hour)); got != full {
			t.Errorf("%d 時在空曠處看得到 %d 格,應該是全部 %d 格", hour, got, full)
		}
	}
}

// 一間房間:看得到房間與四面牆,看不到牆外。
//
// ★ 這是視線遮蔽最實際的效果 —— 站在室內看不到隔壁房間。
// 而且**日夜相同**(理由同上一條)。
func TestYouSeeYourRoomAndItsWallsButNotBeyond(t *testing.T) {
	blocker := u5data.SightBlockers[0]
	floor := walkable(t)
	scenes := synthScenes(t, floor)
	m := &scenes.Files[0][u5data.Locations[britain-1].SceneIndex]
	// 內部 x 3..7、y 4..8(5×5),其餘全是牆。玩家站在正中央 (5,6)。
	for y := 0; y < u5data.SceneSide; y++ {
		for x := 0; x < u5data.SceneSide; x++ {
			if x < 3 || x > 7 || y < 4 || y > 8 {
				m.Tiles[y*u5data.SceneSide+x] = blocker
			}
		}
	}
	for _, hour := range []int{12, 0} {
		s := &State{Scenes: scenes, MaxMessages: 8}
		if err := s.SetScene(britain, 0, 5, 6); err != nil {
			t.Fatalf("進不了不列顛城:%v", err)
		}
		s.Clock.Hour = hour
		mask := s.SightMask()

		// 房間內部看得到。
		for dy := -2; dy <= 2; dy++ {
			for dx := -2; dx <= 2; dx++ {
				if !SightVisible(mask, dx, dy) {
					t.Fatalf("%d 時 (%+d,%+d) 在房間內卻看不到", hour, dx, dy)
				}
			}
		}
		// 貼著房間的那一圈牆看得到。
		for _, p := range [][2]int{{-3, 0}, {3, 0}, {0, -3}, {0, 3}} {
			if !SightVisible(mask, p[0], p[1]) {
				t.Errorf("%d 時 (%+d,%+d) 的牆應該看得到", hour, p[0], p[1])
			}
		}
		// 牆外看不到。
		for _, p := range [][2]int{{-5, 0}, {5, 0}, {0, -5}, {0, 5}, {-4, -4}} {
			if SightVisible(mask, p[0], p[1]) {
				t.Errorf("%d 時 (%+d,%+d) 在牆外,不該看得到", hour, p[0], p[1])
			}
		}
	}
}
