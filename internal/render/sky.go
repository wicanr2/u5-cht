package render

import (
	"image"
	"image/color"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 原版 sub_E2A4 的 12 格日月時間帶。座標是 640×400 畫布上的實測 2× 值:
// DOS 320×200 的起點為 (48,0)，每格 8×8；放大後就是 (96,0)，每格 16×16。
const (
	SkyBandX      = 96
	SkyBandY      = 0
	SkyBandSlots  = 12
	SkyGlyphScale = 2

	RuneSunGlyph = 0x0E
	RuneMoonBase = 0x30
)

// skyMarker 是一個要寫進 12 格緩衝的原版 glyph。
type skyMarker struct {
	slot  int
	glyph byte
	color color.NRGBA
}

// skyBandMarkers 將原版的三個位置公式保留成純函式，方便邊界測試。
//
//	太陽    0x11 - hour
//	Trammel 0x08 - hour，若小於 -12 則加 24
//	Felucca 0x02 - hour，若小於 -12 則加 24
//
// 可見範圍都是 0..11；後寫的 Felucca 會覆蓋同一格，與原版逐字元寫入
// 12 格緩衝的順序一致。
func skyBandMarkers(hour, trammel, felucca int) []skyMarker {
	markers := make([]skyMarker, 0, 3)
	if slot := skyBandSlot(0x11-hour, false); slot >= 0 {
		markers = append(markers, skyMarker{slot: slot, glyph: RuneSunGlyph, color: u5data.EGAPalette[14]})
	}
	if slot := skyBandSlot(0x08-hour, true); slot >= 0 {
		markers = append(markers, skyMarker{slot: slot, glyph: byte(RuneMoonBase + trammel), color: u5data.EGAPalette[7]})
	}
	if slot := skyBandSlot(0x02-hour, true); slot >= 0 {
		markers = append(markers, skyMarker{slot: slot, glyph: byte(RuneMoonBase + felucca), color: u5data.EGAPalette[7]})
	}
	return markers
}

func skyBandSlot(pos int, wrapNegative bool) int {
	if wrapNegative && pos < -12 {
		pos += 24
	}
	if pos < 0 || pos >= SkyBandSlots {
		return -1
	}
	return pos
}

// drawTimeBand 畫大地圖／場景的日月指示；地牢與戰鬥依原版位置碼不畫。
// 現代版與 UIOriginal 都保留，差別只在 UIOriginal 另外有藍白框線。
func (s *Scene) drawTimeBand(dst *image.NRGBA) {
	if s.State == nil || s.RuneCharset == nil || !s.State.SceneOrOverworld() {
		return
	}
	day := s.State.Clock.Day
	if day < 1 || day > u5data.DaysPerMonth {
		return
	}
	if s.State.Moons == nil {
		// RUNES.CH 缺月相表時仍可顯示日間太陽，但不猜月相。
		for _, marker := range skyBandMarkers(s.State.Clock.Hour, -99, -99) {
			if marker.glyph == RuneSunGlyph {
				s.drawRuneGlyph(dst, marker.glyph, SkyBandX+marker.slot*u5data.GlyphWidth*SkyGlyphScale, SkyBandY, marker.color)
			}
		}
		return
	}
	phases := s.State.Moons[day]
	for _, marker := range skyBandMarkers(s.State.Clock.Hour, int(phases[0]), int(phases[1])) {
		s.drawRuneGlyph(dst, marker.glyph,
			SkyBandX+marker.slot*u5data.GlyphWidth*SkyGlyphScale,
			SkyBandY, marker.color)
	}
}

// drawRuneGlyph 以原版 8×8 glyph 的乾淨 2× 放大畫到畫布。
func (s *Scene) drawRuneGlyph(dst *image.NRGBA, glyph byte, x, y int, c color.NRGBA) {
	if s.RuneCharset == nil {
		return
	}
	g, ok := s.RuneCharset.Glyph(glyph)
	if !ok {
		return
	}
	for gy := 0; gy < u5data.GlyphHeight; gy++ {
		for gx := 0; gx < u5data.GlyphWidth; gx++ {
			if !g.At(gx, gy) {
				continue
			}
			for sy := 0; sy < SkyGlyphScale; sy++ {
				for sx := 0; sx < SkyGlyphScale; sx++ {
					SetPixel(dst, x+gx*SkyGlyphScale+sx, y+gy*SkyGlyphScale+sy, c)
				}
			}
		}
	}
}
