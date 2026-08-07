package render

import (
	"fmt"
	"image"
	"image/color"

	"github.com/wicanr2/u5-cht/internal/game"
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 畫面配置
//
// 原版 320×200 EGA 是「左側 11×11 tile 地圖 + 右側狀態 + 下方訊息」。
// 本專案的邏輯畫布是 640×400(原版的乾淨 2×),tile 用 2 倍 nearest 畫成 32 px,
// 地圖視窗仍是 11×11 tile —— 佔畫面比例與原版一致,而中文有空間用正常字級
// (rulebook/81「拉畫布別縮字」)。
const (
	CanvasWidth  = 640
	CanvasHeight = 400

	ViewTiles  = 11
	TileScale  = 2
	TilePixels = u5data.TileSize * TileScale
	ViewPixels = ViewTiles * TilePixels // 352

	MapOriginX = 8
	MapOriginY = 8
	PanelX     = MapOriginX + ViewPixels + 8 // 右側狀態欄起點
	MessageY   = MapOriginY + ViewPixels + 6 // 下方訊息欄起點
)

var (
	ColorBackground = color.NRGBA{R: 0x10, G: 0x10, B: 0x28, A: 0xFF}
	ColorText       = color.NRGBA{R: 0xE8, G: 0xE8, B: 0xD8, A: 0xFF}
	ColorMarker     = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
)

// Scene 是「畫面上該有什麼」的完整描述。它不知道 ebiten 存在。
//
// 這是 headless 驗證與實機顯示的**共同來源**:兩邊都呼叫 Render(),
// 所以截圖就是實機畫面,不會漂移。
type Scene struct {
	State *game.State
	Tiles []u5data.Tile
	Text  *TextRenderer
}

// Render 畫出整個 640×400 畫面。
func (s *Scene) Render() *image.NRGBA {
	dst := image.NewNRGBA(image.Rect(0, 0, CanvasWidth, CanvasHeight))
	fill(dst, ColorBackground)

	s.drawMapView(dst)
	s.drawPanel(dst)
	s.drawMessages(dst)
	return dst
}

func (s *Scene) drawMapView(dst *image.NRGBA) {
	if s.State == nil || len(s.Tiles) == 0 {
		return
	}
	// 11×11 視窗永遠以玩家為中心 —— 原版就是這樣:
	// 場景移動函式 sub_86C 讀鄰格用的 byte_3F789[32*dy+dx] 是一個固定位址,
	// 也就是視窗緩衝裡玩家那一格,四鄰用 ±1 / ±32 直接定址(docs/re/03 §7)。
	half := ViewTiles / 2
	for dy := -half; dy <= half; dy++ {
		for dx := -half; dx <= half; dx++ {
			s.drawTile(dst, int(s.State.TileAt(s.State.X+dx, s.State.Y+dy)),
				MapOriginX+(dx+half)*TilePixels,
				MapOriginY+(dy+half)*TilePixels)
		}
	}
	// 玩家位置標記。
	// TODO(P3):換成原版的 Avatar tile —— 索引要從反編譯碼確認,不猜。
	DrawFrame(dst,
		MapOriginX+half*TilePixels, MapOriginY+half*TilePixels,
		TilePixels, TilePixels, ColorMarker)
}

// drawTile 以 TileScale 倍(nearest)把一個 tile 畫到 (x, y)。
func (s *Scene) drawTile(dst *image.NRGBA, idx, x, y int) {
	if idx < 0 || idx >= len(s.Tiles) {
		idx = 0
	}
	t := &s.Tiles[idx]
	for ty := 0; ty < u5data.TileSize; ty++ {
		for tx := 0; tx < u5data.TileSize; tx++ {
			// 用 TilePalette(已套色號 IGRB→IRGB remap),不要用 EGAPalette。
			c := u5data.TilePalette[t.At(tx, ty)&0x0F]
			for sy := 0; sy < TileScale; sy++ {
				for sx := 0; sx < TileScale; sx++ {
					SetPixel(dst, x+tx*TileScale+sx, y+ty*TileScale+sy, c)
				}
			}
		}
	}
}

func (s *Scene) drawPanel(dst *image.NRGBA) {
	if s.Text == nil {
		return
	}
	st := s.State
	if st == nil {
		return
	}
	y := MapOriginY
	y = s.Text.DrawLines(dst, PanelX, y, []string{"創世紀 V", "命運勇士"})
	y += LineHeight
	s.Text.Draw(dst, PanelX, y, fmt.Sprintf("座標 %3d,%3d", st.X, st.Y))
	y += LineHeight
	s.Text.Draw(dst, PanelX, y, fmt.Sprintf("地形 %3d", st.TileAt(st.X, st.Y)))
	y += LineHeight
	if st.InScene() {
		s.Text.Draw(dst, PanelX, y, "★ "+st.LocationName())
		y += LineHeight
		s.Text.Draw(dst, PanelX, y, fmt.Sprintf("第 %d 層", st.Floor+1))
		y += LineHeight
	} else {
		// 地點名取自原版執行檔的地點表(u5data.Locations),不是自己打的清單。
		if loc, ok := u5data.LocationAt(st.X, st.Y); ok {
			s.Text.Draw(dst, PanelX, y, "★ "+loc.DisplayName())
		}
		y += LineHeight
		if st.Floor < 0 {
			s.Text.Draw(dst, PanelX, y, "地下世界")
			y += LineHeight
		}
	}
	y += LineHeight
	keys := []string{"方向鍵移動", "E 進入", "K 攀爬", "F10 離開"}
	if st.Prompt != game.PromptNone {
		keys = []string{"Y 是 / N 否"}
	}
	s.Text.DrawLines(dst, PanelX, y, keys)
}

func (s *Scene) drawMessages(dst *image.NRGBA) {
	if s.Text == nil {
		return
	}
	if s.State == nil {
		return
	}
	y := MessageY
	maxW := CanvasWidth - 2*MapOriginX
	for _, m := range s.State.Messages {
		// 斷行用的是同一個 Advance,所以畫出來保證不溢框。
		y = s.Text.DrawLines(dst, MapOriginX, y, Wrap(m, maxW))
	}
}

func fill(dst *image.NRGBA, c color.NRGBA) {
	b := dst.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			dst.SetNRGBA(x, y, c)
		}
	}
}

// WrapCoord 把座標折回世界地圖範圍(Britannia 是環繞的)。
func WrapCoord(v int) int {
	v %= u5data.WorldSide
	if v < 0 {
		v += u5data.WorldSide
	}
	return v
}
