package render

import (
	"testing"

	"github.com/wicanr2/u5-cht/internal/game"
	"github.com/wicanr2/u5-cht/internal/u5data"
)

func testRuneCharset() *u5data.Charset {
	c := &u5data.Charset{Glyphs: make([]u5data.Glyph, 128)}
	// 只點亮第一個像素，測試可直接辨認 slot、色號與 2× 放大。
	c.Glyphs[RuneSunGlyph][0] = 0x80
	c.Glyphs[RuneMoonBase+2][0] = 0x40
	c.Glyphs[RuneMoonBase+3][0] = 0x20
	return c
}

func testMoonState(hour int) *game.State {
	m := &u5data.MoonPhases{}
	m[5] = [2]byte{2, 3}
	return &game.State{
		Clock:    game.Clock{Hour: hour, Day: 5, Month: 1, Year: 139},
		Moons:    m,
		Messages: []string{"測試日月時間帶"},
	}
}

func TestSkyBandUsesOriginalTimePositions(t *testing.T) {
	day := skyBandMarkers(8, 2, 3)
	if len(day) != 2 {
		t.Fatalf("08 點應有太陽與 Trammel 兩個 glyph,實際 %d 個", len(day))
	}
	if day[0].slot != 9 || day[0].glyph != RuneSunGlyph {
		t.Errorf("太陽位置=(%d,0x%02X),預期 (9,0x%02X)", day[0].slot, day[0].glyph, RuneSunGlyph)
	}
	if day[1].slot != 0 || day[1].glyph != RuneMoonBase+2 {
		t.Errorf("Trammel 位置=(%d,0x%02X),預期 (0,0x%02X)", day[1].slot, day[1].glyph, RuneMoonBase+2)
	}

	night := skyBandMarkers(21, 2, 3)
	if len(night) != 2 {
		t.Fatalf("21 點應有兩個月亮 glyph,實際 %d 個", len(night))
	}
	if night[0].slot != 11 || night[0].glyph != RuneMoonBase+2 {
		t.Errorf("Trammel 夜間位置=(%d,0x%02X),預期 (11,0x32)", night[0].slot, night[0].glyph)
	}
	if night[1].slot != 5 || night[1].glyph != RuneMoonBase+3 {
		t.Errorf("Felucca 夜間位置=(%d,0x%02X),預期 (5,0x33)", night[1].slot, night[1].glyph)
	}
}

func TestBothUIModesKeepTheSunMoonIndicator(t *testing.T) {
	for _, mode := range []UIMode{UIModern, UIOriginal} {
		s := testScene()
		s.State = testMoonState(8)
		s.UI = mode
		s.RuneCharset = testRuneCharset()
		img := s.Render()

		sunX := SkyBandX + 9*u5data.GlyphWidth*SkyGlyphScale
		if got := img.NRGBAAt(sunX, SkyBandY); got != u5data.EGAPalette[14] {
			t.Errorf("%s 太陽 glyph 是 %v,預期 EGA 14=%v", UIModeNames[mode], got, u5data.EGAPalette[14])
		}
		moonX := SkyBandX
		if got := img.NRGBAAt(moonX+2, SkyBandY); got != u5data.EGAPalette[7] {
			t.Errorf("%s Trammel glyph 是 %v,預期 EGA 7=%v", UIModeNames[mode], got, u5data.EGAPalette[7])
		}
	}
}

func TestBothUIModesDrawClassicChrome(t *testing.T) {
	for _, mode := range []UIMode{UIModern, UIOriginal} {
		s := testScene()
		s.UI = mode
		pixel := s.Render().NRGBAAt(0, 8)
		if pixel != u5data.EGAPalette[1] {
			t.Errorf("%s 框線外層是 %v,預期 EGA 深藍=%v", UIModeNames[mode], pixel, u5data.EGAPalette[1])
		}
	}
}
