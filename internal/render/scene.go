package render

import (
	"fmt"
	"image"
	"image/color"

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
	World    *u5data.WorldMap
	Tiles    []u5data.Tile
	Text     *TextRenderer
	PX, PY   int      // 玩家在世界地圖的座標
	Messages []string // 訊息欄(最新在後)
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
	if s.World == nil || len(s.Tiles) == 0 {
		return
	}
	half := ViewTiles / 2
	for dy := -half; dy <= half; dy++ {
		for dx := -half; dx <= half; dx++ {
			// 世界地圖是環繞的(原版 Britannia 就是 wrap-around)。
			wx, wy := WrapCoord(s.PX+dx), WrapCoord(s.PY+dy)
			s.drawTile(dst, int(s.World.At(wx, wy)),
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
	y := MapOriginY
	y = s.Text.DrawLines(dst, PanelX, y, []string{"創世紀 V", "命運勇士"})
	y += LineHeight
	s.Text.Draw(dst, PanelX, y, fmt.Sprintf("座標 %3d,%3d", s.PX, s.PY))
	y += LineHeight
	if s.World != nil {
		s.Text.Draw(dst, PanelX, y, fmt.Sprintf("地形 %3d", s.World.At(s.PX, s.PY)))
	}
	y += LineHeight * 2
	s.Text.DrawLines(dst, PanelX, y, []string{"方向鍵移動", "F10 離開"})
}

func (s *Scene) drawMessages(dst *image.NRGBA) {
	if s.Text == nil {
		return
	}
	y := MessageY
	maxW := CanvasWidth - 2*MapOriginX
	for _, m := range s.Messages {
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
