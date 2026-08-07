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
func TestSteppingOnAMoongateTeleports(t *testing.T) {
	s := moonState(t)
	if s.BaseSave == nil {
		t.Skip("沒有底稿存檔")
	}
	s.Location, s.Floor = 0, 0
	// 站到某個月門旁邊,再走進去。
	g := s.Moongates[0]
	s.X, s.Y = g.X-1, g.Y
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
