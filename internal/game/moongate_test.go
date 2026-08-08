package game

import (
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

func moonState(t *testing.T) *State {
	t.Helper()
	s := magicState(t)
	mp, err := u5data.LoadMoonPhases("../../gamedata")
	if err != nil {
		t.Fatal(err)
	}
	s.Moons = mp
	return s
}

// TestMoongatesSitNextToTheirCity:八座月門每一座都貼著對應的城市。
//
// 這是月門表位移最硬的驗收 —— 八組座標同時落在八座城市 10 格內,
// 位移偏一個位元組就不可能還成立。
func TestMoongatesSitNextToTheirCity(t *testing.T) {
	s := moonState(t)
	if s.BaseSave == nil {
		t.Skip("沒有底稿存檔")
	}
	// 一城一門 —— 八座最近的城市必須互不相同,這比單純的距離門檻硬。
	nearest := map[string]int{}
	for i := 0; i < u5data.MoonPhaseCount; i++ {
		g := s.Moongates[i]
		if !g.Known() {
			t.Errorf("第 %d 個相位的目的地是未知的", i)
			continue
		}
		if g.Location != 0 {
			t.Errorf("第 %d 個相位的地點是 %d,月門該通到大地圖", i, g.Location)
		}
		// 找最近的城市。
		best, bestD := "", 1<<30
		for _, loc := range u5data.Locations {
			d := iabs(loc.X-g.X) + iabs(loc.Y-g.Y)
			if d < bestD {
				bestD, best = d, loc.Name
			}
		}
		// ⚠ 距離門檻不能訂太緊:紫衫城(Yew)在森林裡,它那座門離城 14 格。
		// 我第一版寫 12,只有那一座紅 —— 斷言寫太死。
		if bestD > 16 {
			t.Errorf("第 %d 個相位在 (%d,%d),最近的城市 %s 也有 %d 格遠",
				i, g.X, g.Y, best, bestD)
		}
		if prev, dup := nearest[best]; dup {
			t.Errorf("相位 %d 與 %d 最近的城市都是 %s —— 一城一門才對", prev, i, best)
		}
		nearest[best] = i
		t.Logf("相位 %d → (%3d,%3d) 貼著 %s(%d 格)", i, g.X, g.Y, best, bestD)
	}
	if len(nearest) != u5data.MoonPhaseCount {
		t.Errorf("八座月門只對到 %d 座城市", len(nearest))
	}
}

// TestMoonPhasesAreValidAndMove:月相表每一天都在 0..7,而且會變。
func TestMoonPhasesAreValidAndMove(t *testing.T) {
	s := moonState(t)
	seenT, seenF := map[int]bool{}, map[int]bool{}
	for d := 1; d <= u5data.DaysPerMonth; d++ {
		tr := s.Moons.PhaseAt(d, 0)  // 上午 → Trammel
		fe := s.Moons.PhaseAt(d, 13) // 下午 → Felucca
		if tr < 0 || tr >= 8 || fe < 0 || fe >= 8 {
			t.Fatalf("第 %d 日的月相是 %d/%d,不在 0..7", d, tr, fe)
		}
		seenT[tr], seenF[fe] = true, true
	}
	// 一個月裡兩顆月亮都該走完全部八個相位。
	if len(seenT) != 8 {
		t.Errorf("Trammel 一個月只出現 %d 個相位,預期 8", len(seenT))
	}
	if len(seenF) != 8 {
		t.Errorf("Felucca 一個月只出現 %d 個相位,預期 8", len(seenF))
	}
}

// TestMoongateDependsOnTimeOfDay:同一座門中午前後通到不同地方。
//
// 原版 `sub_E084`:`小時 < 12` 看 Trammel、否則看 Felucca。少了這一條,
// 月門會變成固定傳送點,整個系統就沒意義了。
func TestMoongateDependsOnTimeOfDay(t *testing.T) {
	s := moonState(t)
	differ := 0
	for d := 1; d <= u5data.DaysPerMonth; d++ {
		if s.Moons.PhaseAt(d, 6) != s.Moons.PhaseAt(d, 18) {
			differ++
		}
	}
	if differ == 0 {
		t.Error("28 天裡上午與下午的相位都一樣 —— 時段判斷大概沒生效")
	}
	t.Logf("28 天裡有 %d 天上午與下午通到不同的門", differ)
}

// TestSteppingOnAMoongateTeleports:走上月門就被捲走。
//
// ⚠ **要在夜裡。** 原版的月門是 `sub_DEE4` 每次重畫時寫進地圖的 tile 0xDC,
// 而它只在 20:00–04:59 寫(白天寫回草地)⇒ **白天沒有月門可踏**
// (`docs/re/86`)。第一版沒設時間,紅燈看起來像傳送壞了。
func TestSteppingOnAMoongateTeleports(t *testing.T) {
	s := moonState(t)
	if s.BaseSave == nil {
		t.Skip("沒有底稿存檔")
	}
	s.Location, s.Floor = 0, 0
	s.Clock.Hour = u5data.MoongateNightFrom // 夜裡
	// 站到某個月門旁邊,再走進去。
	g := s.Moongates[0]
	s.X, s.Y = g.X-1, g.Y
	// ⚠ 視窗原點現在是真的狀態(原版 `byte_3E0AB`/`byte_3E0AC`)——
	// 直接改座標之後要跟著定位,否則 `moongateWritesHere` 的視窗檢查
	// 會拿 (0,0) 當原點。真實遊戲裡這由 `resetLoadWindow` 負責。
	s.resetLoadWindow()
	s.Move(East)
	want := s.Moongates[s.MoonPhaseNow()]
	if s.X != want.X || s.Y != want.Y {
		t.Errorf("踏進月門之後在 (%d,%d),預期相位 %d 的 (%d,%d)",
			s.X, s.Y, s.MoonPhaseNow(), want.X, want.Y)
	}
}

// TestGreatGateRefusesUnderSail:揚帆中不能用大傳送門。
func TestGreatGateRefusesUnderSail(t *testing.T) {
	s := moonState(t)
	s.Location = 0
	s.Transport = u5data.VehicleSailing
	if s.CastGreatGate(0) {
		t.Error("揚帆中還是傳送了")
	}
	s.Transport = u5data.VehicleWalk
	if !s.CastGreatGate(0) {
		t.Errorf("走路時大傳送門卻失敗:\n%s", s.log())
	}
}

// TestMoongateTileIsWrittenAtNightAndErasedByDay —— ★★ 月門是寫進地圖的一格。
//
// 原版 `sub_DEE4` 每次重畫都跑:夜裡把 tile 0xDC 寫進**月石埋藏的座標**,
// 白天把計數器降到 0 之後寫回草地(tile 5)。⇒ 「月門存在」的唯一來源是
// 地圖上那一格,不是「座標 + 時段」的判斷(`docs/re/86`)。
func TestMoongateTileIsWrittenAtNightAndErasedByDay(t *testing.T) {
	s := moonState(t)
	if s.BaseSave == nil {
		t.Skip("沒有底稿存檔")
	}
	s.Location, s.Floor = 0, 0
	g := s.Moongates[0]
	// 站到那顆月石旁邊,月門才在載入視窗內(原版 `sub_DE74` 的範圍檢查)。
	s.X, s.Y = g.X-2, g.Y
	s.resetLoadWindow()

	// 夜裡:那一格該變成月門。
	s.Clock.Hour = u5data.MoongateNightFrom
	s.RefreshMoongateTiles()
	if got := s.TileAt(g.X, g.Y); got != u5data.MoongateOpenTile {
		t.Fatalf("夜裡那一格是 0x%02X,預期月門 0x%02X", got, u5data.MoongateOpenTile)
	}

	// ⚠ 要先把計數器養滿才驗得到殘留 —— 只開一回合的門,天一亮就該關,
	// **而那正是原版行為**(計數器 1 → 遞減一次就歸零)。
	// 第一版沒養就直接驗殘留,紅燈看起來像實作錯了。
	for i := 0; i < MoongateFrameMax; i++ {
		s.RefreshMoongateTiles()
	}
	if s.MoongateFrame != MoongateFrameMax {
		t.Fatalf("計數器養到 %d,預期上限 %d", s.MoongateFrame, MoongateFrameMax)
	}

	// ★ 白天不是立刻關 —— 計數器要先降到 0。
	s.Clock.Hour = 12
	s.RefreshMoongateTiles()
	if got := s.TileAt(g.X, g.Y); got != u5data.MoongateOpenTile {
		t.Error("天一亮月門就消失了 —— 原版是計數器歸零才寫回草地")
	}
	for i := 0; i < MoongateFrameMax+2; i++ {
		s.RefreshMoongateTiles()
	}
	if got := s.TileAt(g.X, g.Y); got != u5data.MoongateClosedTile {
		t.Errorf("計數器歸零之後那一格是 0x%02X,預期草地 0x%02X",
			got, u5data.MoongateClosedTile)
	}

	// 時段判準的邊界(>= 20 或 < 5)。
	for _, tc := range []struct {
		hour int
		open bool
	}{
		{4, true}, {5, false}, {12, false}, {19, false}, {20, true}, {23, true},
	} {
		if got := u5data.MoongateOpenAtHour(tc.hour); got != tc.open {
			t.Errorf("%02d 點 → %v,預期 %v", tc.hour, got, tc.open)
		}
	}
}

// TestEnteringReadsTheTileNotTheCoordinates —— ★★ 只有一個真相來源。
//
// 原版 `sub_E084` 只做一件事:讀腳下那一格的 tile。所以
//
//	月石的埋藏座標**上沒有月門 tile** → 踏上去什麼都不會發生
//	不是埋藏點的格子**有月門 tile**   → 踏上去照樣被傳送
//
// ⚠ 此前 `EnterMoongateHere` 查的是「座標 + 時段」,而 tile 是另一條路 ——
// **兩個真相來源遲早會漂**。這條測試就是防止它漂回去。
func TestEnteringReadsTheTileNotTheCoordinates(t *testing.T) {
	s := moonState(t)
	if s.BaseSave == nil {
		t.Skip("沒有底稿存檔")
	}
	s.Location, s.Floor = 0, 0
	g := s.Moongates[0]

	// (a) 站在埋藏點上,但那一格是草地 → 不該傳送。
	s.X, s.Y = g.X, g.Y
	if !s.SetTileAt(g.X, g.Y, u5data.MoongateClosedTile) {
		t.Fatal("寫不進世界地圖")
	}
	if s.EnterMoongateHere() {
		t.Error("埋藏點上沒有月門 tile 卻傳送了 —— 判準該是 tile 不是座標")
	}

	// (b) 隨便一格寫上月門 tile → 該傳送。
	//
	// ⚠ 要避開午夜前十分鐘的空窗(見 `TestMidnightMoongateSwallowsWithoutSending`)
	// —— 那時 `EnterMoongateHere` 也回 true,只驗回傳值會變成假綠燈。
	// 所以這裡連**人有沒有真的移動**一起驗。
	s.Clock.Hour = u5data.MoongateNightFrom
	fx, fy := g.X+7, g.Y+7
	s.X, s.Y = fx, fy
	if !s.SetTileAt(fx, fy, u5data.MoongateOpenTile) {
		t.Fatal("寫不進世界地圖")
	}
	if !s.EnterMoongateHere() {
		t.Error("有月門 tile 卻沒傳送 —— 原版只讀那一格")
	}
	if s.X == fx && s.Y == fy {
		t.Error("回了 true 但人沒動")
	}
}

// TestSteppingThroughClosesTheGateBehindYou —— ★ 踏過的月門立刻變回草地。
//
// 原版 `sub_E084` 在判「要不要傳送」**之前**就 `mov byte ptr [eax], 5` ——
// 把腳下那一格寫回草地(`docs/re/86` §6)。夜裡下一次 `RefreshMoongateTiles`
// 會再寫回 0xDC ⇒ 畫面上是「吸走你之後閉合、再張開」。
func TestSteppingThroughClosesTheGateBehindYou(t *testing.T) {
	s := moonState(t)
	if s.BaseSave == nil {
		t.Skip("沒有底稿存檔")
	}
	s.Location, s.Floor = 0, 0
	s.Clock.Hour = u5data.MoongateNightFrom
	// 挑一格離月石很遠的地方,免得 `RefreshMoongateTiles` 立刻又寫回去。
	fx, fy := s.Moongates[0].X+40, s.Moongates[0].Y+40
	s.X, s.Y = fx, fy
	if !s.SetTileAt(fx, fy, u5data.MoongateOpenTile) {
		t.Fatal("寫不進世界地圖")
	}
	if !s.EnterMoongateHere() {
		t.Fatal("沒傳送")
	}
	if got := s.TileAt(fx, fy); got != u5data.MoongateClosedTile {
		t.Errorf("踏過的那一格是 0x%02X,預期草地 0x%02X", got, u5data.MoongateClosedTile)
	}
}

// TestMidnightMoongateSwallowsWithoutSending —— ★★ 午夜前十分鐘踏上月門不會傳送。
//
// 原版 `sub_E084`:`if (byte_3E08F == 0 && byte_3E091 < 0Ah) esi = 1`,
// 直接跳過 `sub_DF84`(傳送)。門照樣關掉,人留在原地 ——
// **沒有訊息、沒有動畫**(`docs/re/86` §6)。
//
// ⚠ 這看起來像 bug,而它可能真的是。`CLAUDE.md §3.0`:照原樣做,
// 不補提示、不「順手」修掉。已列進 A 階段對 DOSBox 的核對清單。
func TestMidnightMoongateSwallowsWithoutSending(t *testing.T) {
	s := moonState(t)
	if s.BaseSave == nil {
		t.Skip("沒有底稿存檔")
	}
	s.Location, s.Floor = 0, 0
	fx, fy := s.Moongates[0].X+40, s.Moongates[0].Y+40

	step := func(hour, minute int) (moved bool, tile byte) {
		s.Clock.Hour, s.Clock.Minute = hour, minute
		s.X, s.Y = fx, fy
		if !s.SetTileAt(fx, fy, u5data.MoongateOpenTile) {
			t.Fatal("寫不進世界地圖")
		}
		s.EnterMoongateHere()
		return s.X != fx || s.Y != fy, s.TileAt(fx, fy)
	}

	// 00:00–00:09:門關掉,人不動。
	for _, m := range []int{0, 5, 9} {
		moved, tile := step(0, m)
		if moved {
			t.Errorf("00:%02d 傳送了 —— 原版那十分鐘不送人", m)
		}
		if tile != u5data.MoongateClosedTile {
			t.Errorf("00:%02d 之後那一格是 0x%02X,門該關掉", m, tile)
		}
	}
	// 00:10 起恢復正常。
	if moved, _ := step(0, u5data.MoongateDeadMinutes); !moved {
		t.Error("00:10 還是不送人 —— 判準是 `分鐘 < 10`")
	}
	// 別的小時的第 0 分鐘不受影響(判準含 `小時 == 0`)。
	if moved, _ := step(u5data.MoongateNightFrom, 0); !moved {
		t.Error("20:00 不送人 —— 空窗只在午夜那個小時")
	}
}

// TestMoongateOnlyWritesInTheLoadedWindow —— `sub_DE74` 的範圍檢查。
//
// ⚠ 這一條有可觀察的後果:**離開視窗時原版不會把那一格寫回草地**,
// 所以遠處的月門 tile 會留在地圖上直到玩家再走近。照原樣做。
func TestMoongateOnlyWritesInTheLoadedWindow(t *testing.T) {
	s := moonState(t)
	if s.BaseSave == nil {
		t.Skip("沒有底稿存檔")
	}
	s.Location, s.Floor = 0, 0
	g := s.Moongates[0]
	// 站得很遠 → 不該寫。
	s.X, s.Y = g.X+80, g.Y+80
	before := s.TileAt(g.X, g.Y)
	s.Clock.Hour = u5data.MoongateNightFrom
	s.RefreshMoongateTiles()
	if s.TileAt(g.X, g.Y) != before {
		t.Error("站在 80 格外也把月門寫進地圖了 —— `sub_DE74` 有 32×32 的範圍檢查")
	}
}

// TestCreatureWalkingIntoAMoongateVanishes —— ★★ 怪跟玩家一樣被月門捲走。
//
// `sub_2870` 尾段:新格 tile == 0xDC 就把種類碼與 tile 都歸零。
//
// ⚠ 這裡的 0xDC 是**地圖 tile**;`sub_2870` 上面那個 0xDC 是**龍的物件種類碼**。
// 同一支函式、同一個值、兩個命名空間 —— 這條測試順便釘住兩者沒被搞混。
func TestCreatureWalkingIntoAMoongateVanishes(t *testing.T) {
	s := overworldScene(t)
	clearObjects(t, s)
	flatGrass(t, s, 8)
	ox, oy := s.X+4, s.Y
	putObject(t, s, 5, 0xC0, ox, oy) // Orc
	if !s.SetTileAt(ox-1, oy, u5data.MoongateOpenTile) {
		t.Fatal("寫不進世界地圖")
	}
	s.stepObject(5, -1, 0)
	if got := s.currentObjects().Objects[5].Raw[u5data.ObjKind]; got != 0 {
		t.Errorf("怪走進月門之後種類碼還是 0x%02X,原版把它歸零", got)
	}

	// ★ 反面:龍(物件種類碼 0xDC)走在**草地**上不該消失。
	s2 := overworldScene(t)
	clearObjects(t, s2)
	flatGrass(t, s2, 8)
	putObject(t, s2, 6, u5data.FlyerDragon, s2.X+4, s2.Y)
	s2.stepObject(6, -1, 0)
	if s2.currentObjects().Objects[6].Raw[u5data.ObjKind] != u5data.FlyerDragon {
		t.Error("龍在草地上消失了 —— 把物件種類碼 0xDC 當成地圖 tile 0xDC 了")
	}
}
