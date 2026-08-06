package render

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 地圖視窗
//
// 原版 320×200 EGA 的畫面配置是「左側 11×11 tile 地圖 + 右側狀態 + 下方訊息」。
// 本專案的邏輯畫布是 640×400(原版的乾淨 2×),所以 tile 用 **2 倍 nearest** 畫成 32 px,
// 地圖視窗仍是 11×11 tile = 352×352 px —— 佔畫面比例與原版一致,而中文有空間用正常字級。
// (這就是 rulebook/81「拉畫布別縮字」的做法。)
const (
	// ViewTiles 是地圖視窗的邊長(tile 數)。
	ViewTiles = 11
	// TileScale 是 tile 的整數放大倍率。
	TileScale = 2
	// TilePixels 是放大後一個 tile 的邊長。
	TilePixels = u5data.TileSize * TileScale
	// ViewPixels 是地圖視窗的邊長(像素)。
	ViewPixels = ViewTiles * TilePixels
)

// TileSet 是已上傳到 GPU 的 tileset。
type TileSet struct {
	tex   *ebiten.Image
	count int
}

// NewTileSet 把解碼好的 tile 烘成一張橫排紋理。
func NewTileSet(tiles []u5data.Tile) *TileSet {
	if len(tiles) == 0 {
		return &TileSet{}
	}
	img := image.NewNRGBA(image.Rect(0, 0, len(tiles)*u5data.TileSize, u5data.TileSize))
	for i := range tiles {
		ox := i * u5data.TileSize
		for y := 0; y < u5data.TileSize; y++ {
			for x := 0; x < u5data.TileSize; x++ {
				// 用 TilePalette(已套 IGRB→IRGB 的色號 remap),不要用 EGAPalette。
				img.SetNRGBA(ox+x, y, u5data.TilePalette[tiles[i].At(x, y)&0x0F])
			}
		}
	}
	return &TileSet{tex: ebiten.NewImageFromImage(img), count: len(tiles)}
}

// Count 回傳 tile 數。
func (ts *TileSet) Count() int { return ts.count }

// DrawTile 把第 idx 個 tile 以 TileScale 倍畫到 (x, y)。
func (ts *TileSet) DrawTile(dst *ebiten.Image, idx, x, y int) {
	if ts.tex == nil {
		return
	}
	if idx < 0 || idx >= ts.count {
		idx = 0
	}
	sub := ts.tex.SubImage(image.Rect(
		idx*u5data.TileSize, 0,
		(idx+1)*u5data.TileSize, u5data.TileSize,
	)).(*ebiten.Image)
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
	op.GeoM.Scale(TileScale, TileScale)
	op.GeoM.Translate(float64(x), float64(y))
	dst.DrawImage(sub, op)
}

// DrawWorldView 以 (centerX, centerY) 為中心畫 11×11 的地圖視窗。
// 世界地圖是環繞的(Britannia 在原版就是 wrap-around),所以座標取模。
func (ts *TileSet) DrawWorldView(dst *ebiten.Image, world *u5data.WorldMap, centerX, centerY, originX, originY int) {
	half := ViewTiles / 2
	for dy := -half; dy <= half; dy++ {
		for dx := -half; dx <= half; dx++ {
			wx := wrap(centerX+dx, u5data.WorldSide)
			wy := wrap(centerY+dy, u5data.WorldSide)
			ts.DrawTile(dst, int(world.At(wx, wy)),
				originX+(dx+half)*TilePixels,
				originY+(dy+half)*TilePixels)
		}
	}
}

func wrap(v, n int) int {
	v %= n
	if v < 0 {
		v += n
	}
	return v
}
