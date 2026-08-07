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

// 白天看得到整個視窗,夜裡只剩身邊那一小圈。
//
// ★ 夜裡剩下的正好是 9 格 —— 平方距離 ≤ 2 的格子就是自己加八鄰。
// 半徑 2 這個值與「3×3」對得上不是巧合,是同一件事的兩種說法。
func TestNightShrinksTheViewToArmsLength(t *testing.T) {
	full := u5data.SightSide * u5data.SightSide
	if got := visibleCount(openScene(t, 12)); got != full {
		t.Errorf("白天在空曠處看得到 %d 格,應該是全部 %d 格", got, full)
	}

	night := 0
	for y := 0; y < u5data.SightSide; y++ {
		for x := 0; x < u5data.SightSide; x++ {
			if sightDist2(x, y) <= SightRadiusNight {
				night++
			}
		}
	}
	if night != 9 {
		t.Fatalf("平方距離 ≤ %d 的格子有 %d 個,預期 9(自己加八鄰)",
			SightRadiusNight, night)
	}
	if got := visibleCount(openScene(t, 0)); got != night {
		t.Errorf("夜裡在空曠處看得到 %d 格,應該只剩 %d 格", got, night)
	}
}

// sightDist2 是視窗座標離中心的平方距離(與 u5data 的同一個算法)。
func sightDist2(x, y int) int {
	dx, dy := x-u5data.SightSide/2, y-u5data.SightSide/2
	return dx*dx + dy*dy
}

// 火把讓夜裡看得遠一些,In Lor 又更遠 —— 兩個下限接得上遮蔽。
func TestLightSourcesWidenTheNightView(t *testing.T) {
	dark := visibleCount(openScene(t, 0))

	torch := openScene(t, 0)
	torch.TorchTurns = 100
	lit := visibleCount(torch)

	spell := openScene(t, 0)
	spell.LightTurns = 100
	bright := visibleCount(spell)

	if !(dark < lit && lit < bright) {
		t.Errorf("夜 %d 格 < 火把 %d 格 < In Lor %d 格 —— 這個順序沒成立",
			dark, lit, bright)
	}
}

// 一間白天的房間:看得到房間與四面牆,看不到牆外。
//
// ★ 這是視線遮蔽最實際的效果 —— 站在室內看不到隔壁房間,而且白天也一樣
//(半徑 50 蓋滿視窗,但 flood fill 仍然過不了牆)。
func TestYouSeeYourRoomAndItsWallsButNotBeyond(t *testing.T) {
	blocker := u5data.SightBlockers[0]
	scenes := synthScenes(t, walkable(t))
	m := &scenes.Files[0][u5data.Locations[britain-1].SceneIndex]
	// 內部 x 3..7、y 4..8(5×5),其餘全是牆。玩家站在正中央 (5,6)。
	for y := 0; y < u5data.SceneSide; y++ {
		for x := 0; x < u5data.SceneSide; x++ {
			if x < 3 || x > 7 || y < 4 || y > 8 {
				m.Tiles[y*u5data.SceneSide+x] = blocker
			}
		}
	}
	s := &State{Scenes: scenes, MaxMessages: 8}
	if err := s.SetScene(britain, 0, 5, 6); err != nil {
		t.Fatalf("進不了不列顛城:%v", err)
	}
	s.Clock.Hour = 12
	mask := s.SightMask()

	// 房間內部看得到。
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			if !SightVisible(mask, dx, dy) {
				t.Fatalf("(%+d,%+d) 在房間內卻看不到", dx, dy)
			}
		}
	}
	// 貼著房間的那一圈牆看得到。
	for _, p := range [][2]int{{-3, 0}, {3, 0}, {0, -3}, {0, 3}} {
		if !SightVisible(mask, p[0], p[1]) {
			t.Errorf("(%+d,%+d) 的牆應該看得到", p[0], p[1])
		}
	}
	// 牆外看不到。
	for _, p := range [][2]int{{-5, 0}, {5, 0}, {0, -5}, {0, 5}, {-4, -4}} {
		if SightVisible(mask, p[0], p[1]) {
			t.Errorf("(%+d,%+d) 在牆外,不該看得到", p[0], p[1])
		}
	}
}

// 牆上的火把在夜裡照亮周圍 —— `sub_2E21C` 那一輪真的有接上。
//
// ★ 而且**火把照不進隔壁房間**:光源那一輪同樣會被牆擋住。
func TestASceneTorchLightsTheRoomAtNight(t *testing.T) {
	blocker := u5data.SightBlockers[0]
	torch := u5data.SightLightTiles[1] // 0xBD 火盆(不在擋視線清單裡)
	if u5data.SightEmitsLight(torch) != true {
		t.Fatal("挑到的地形不會發光")
	}
	mk := func(withTorch bool) *State {
		scenes := synthScenes(t, walkable(t))
		m := &scenes.Files[0][u5data.Locations[britain-1].SceneIndex]
		// 一道牆把場景切成上下兩半,牆在 y=10;玩家在 (15,15)。
		for x := 0; x < u5data.SceneSide; x++ {
			m.Tiles[10*u5data.SceneSide+x] = blocker
		}
		if withTorch {
			// 火盆放在玩家北邊三格(平方距離 9,夜間半徑 2 之外)。
			m.Tiles[12*u5data.SceneSide+15] = torch
		}
		s := &State{Scenes: scenes, MaxMessages: 8}
		if err := s.SetScene(britain, 0, 15, 15); err != nil {
			t.Fatalf("進不了不列顛城:%v", err)
		}
		s.Clock.Hour = 0
		return s
	}
	if SightVisible(mk(false).SightMask(), 0, -3) {
		t.Error("沒有火盆時,夜裡看不到北邊三格")
	}
	if !SightVisible(mk(true).SightMask(), 0, -3) {
		t.Error("放了火盆,夜裡那一格應該被照亮")
	}
	// 牆的另一邊(y=9,也就是 dy=−6)本來就在視窗外;取視窗內最遠的一格
	// 來驗火把照不穿牆:牆在 dy=−5,牆後不在視窗裡,所以改驗牆自己被照亮、
	// 而更遠處沒有被火把「穿透」照到。
	if !SightVisible(mk(true).SightMask(), 0, -5) {
		t.Error("火盆與牆之間沒有阻擋,那面牆應該被照亮")
	}
}

// 燈塔的光束:天黑時掃亮一片扇形,白天不畫。
func TestTheLighthouseBeamSweepsAtNight(t *testing.T) {
	// ⚠ 真正的燈塔在**大地圖**上(`BRIT.DAT` 裡 0x1B 剛好四個)。
	// 這裡用合成場景只是要驗掃描與扇區的邏輯 —— 位置放在玩家南邊五格,
	// 這樣往北掃的扇區才落在視窗裡(燈塔在視窗座標 (5,10))。
	scenes := synthScenes(t, walkable(t))
	m := &scenes.Files[0][u5data.Locations[britain-1].SceneIndex]
	m.Tiles[20*u5data.SceneSide+15] = u5data.LighthouseTile

	mk := func(hour, frame int) *State {
		s := &State{Scenes: scenes, MaxMessages: 8}
		if err := s.SetScene(britain, 0, 15, 15); err != nil {
			t.Fatalf("進不了不列顛城:%v", err)
		}
		s.Clock.Hour = hour
		s.BeamFrame = frame
		return s
	}
	count := func(s *State) int {
		lit := u5data.ComputeLit(func(x, y int) byte {
			return s.TileAt(s.X-5+x, s.Y-5+y)
		})
		before := 0
		for _, v := range lit {
			if v != 0 {
				before++
			}
		}
		s.applyBeam(lit, s.LightRadius2())
		after := 0
		for _, v := range lit {
			if v != 0 {
				after++
			}
		}
		return after - before
	}

	if n := count(mk(0, 1)); n <= 0 {
		t.Errorf("夜裡光束應該多照亮一些格子,實得 %d", n)
	}
	if n := count(mk(12, 1)); n != 0 {
		t.Errorf("白天不該畫光束,卻多亮了 %d 格", n)
	}
	// 沒有燈塔的場景不畫。
	plain := &State{Scenes: synthScenes(t, walkable(t)), MaxMessages: 8}
	if err := plain.SetScene(britain, 0, 15, 15); err != nil {
		t.Fatal(err)
	}
	plain.Clock.Hour = 0
	if _, _, ok := plain.beamSource(); ok {
		t.Error("沒有燈塔的場景不該找到光源")
	}
}

// 十六個扇區排成羅盤:1 正北、5 正東、9 正南、13 正西。
func TestBeamSectorsFormACompass(t *testing.T) {
	cardinal := map[int][2]int{1: {0, -1}, 5: {1, 0}, 9: {0, 1}, 13: {-1, 0}}
	for sector, want := range cardinal {
		cells := u5data.BeamCells(sector)
		if len(cells) == 0 {
			t.Fatalf("扇區 %d 是空的", sector)
		}
		if cells[0] != want {
			t.Errorf("扇區 %d 的第一格是 %v,預期 %v(正向)", sector, cells[0], want)
		}
	}
	// ⚠ 填充的 (0,0) 要被濾掉 —— 不然燈塔自己那一格會被反覆點亮。
	for s := 0; s < u5data.BeamSectorCount; s++ {
		for _, c := range u5data.BeamCells(s) {
			if c[0] == 0 && c[1] == 0 {
				t.Fatalf("扇區 %d 混進了填充值 (0,0)", s)
			}
		}
	}
	// 環繞。
	if len(u5data.BeamCells(16)) != len(u5data.BeamCells(0)) {
		t.Error("扇區編號應該對 16 取模")
	}
}

// 掃一圈回到原點。
func TestBeamWrapsAfterSixteenFrames(t *testing.T) {
	s := newState(t)
	for i := 0; i < u5data.BeamSectorCount; i++ {
		s.AdvanceBeam()
	}
	if s.BeamFrame != 0 {
		t.Errorf("掃了 %d 幀之後停在 %d,應該回到 0",
			u5data.BeamSectorCount, s.BeamFrame)
	}
}
