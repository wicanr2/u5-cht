package render

import (
	"image/color"
	"testing"

	"github.com/wicanr2/u5-cht/internal/game"
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 這些測試能存在,本身就是架構決策的成果:
// render 不依賴 ebiten,所以畫面組裝可以在沒有顯示環境的地方測。
// (綁 ebiten 的版本連 go test 都會因為 ui.init 找不到 X11 而 panic。)

func testScene() *Scene {
	tiles := make([]u5data.Tile, 8)
	for i := range tiles {
		for p := range tiles[i].Pix {
			tiles[i].Pix[p] = byte(i % 16)
		}
	}
	world := &u5data.WorldMap{}
	for i := range world.Tiles {
		world.Tiles[i] = byte(i % 8)
	}
	return &Scene{
		State: &game.State{
			World: world, X: 100, Y: 100,
			Messages: []string{"汝已抵達不列顛尼亞。"},
		},
		Tiles: tiles,
		Text:  NewTextRenderer(nil, nil, ColorText),
	}
}

func TestSceneRenderSize(t *testing.T) {
	img := testScene().Render()
	if b := img.Bounds(); b.Dx() != CanvasWidth || b.Dy() != CanvasHeight {
		t.Errorf("畫面 %d×%d,預期 %d×%d", b.Dx(), b.Dy(), CanvasWidth, CanvasHeight)
	}
}

// TestSceneSurvivesMissingAssets:素材缺件時要優雅降級,不能 panic。
// 玩家可能沒放原版資料、沒烘字型 —— 那時該畫出空畫面並由上層報告缺什麼。
func TestSceneSurvivesMissingAssets(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("缺素材時 panic 了:%v", r)
		}
	}()
	empty := &Scene{
		State: &game.State{Messages: []string{"測試"}},
		Text:  NewTextRenderer(nil, nil, ColorText),
	}
	if img := empty.Render(); img == nil {
		t.Fatal("Render 回傳 nil")
	}
	(&Scene{}).Render() // 連 Text 都沒有也不該爆
}

func TestScenePlayerMarkerAtCenter(t *testing.T) {
	img := testScene().Render()
	half := ViewTiles / 2
	mx := MapOriginX + half*TilePixels
	my := MapOriginY + half*TilePixels
	// 標記是空心白框:四個角應該是 marker 色
	for _, p := range [][2]int{
		{mx, my},
		{mx + TilePixels - 1, my},
		{mx, my + TilePixels - 1},
		{mx + TilePixels - 1, my + TilePixels - 1},
	} {
		if got := img.NRGBAAt(p[0], p[1]); got != ColorMarker {
			t.Errorf("(%d,%d) 應為玩家標記色 %v,實得 %v", p[0], p[1], ColorMarker, got)
		}
	}
	// 框內部不該被填滿
	if img.NRGBAAt(mx+TilePixels/2, my+TilePixels/2) == ColorMarker {
		t.Error("玩家標記應該是空心框,不是實心方塊")
	}
}

// TestSceneWrapAroundWorld:世界地圖是環繞的,在邊界附近取景不該越界或 panic。
func TestSceneWrapAroundWorld(t *testing.T) {
	s := testScene()
	for _, pos := range [][2]int{{0, 0}, {255, 255}, {0, 255}, {255, 0}} {
		s.State.X, s.State.Y = pos[0], pos[1]
		if img := s.Render(); img == nil {
			t.Fatalf("在 (%d,%d) 取景失敗", pos[0], pos[1])
		}
	}
	if WrapCoord(-1) != u5data.WorldSide-1 || WrapCoord(u5data.WorldSide) != 0 {
		t.Error("WrapCoord 沒有正確環繞")
	}
}

func TestDrawFrameAndSetPixelClamp(t *testing.T) {
	img := testScene().Render()
	// 畫在畫布外不該 panic
	DrawFrame(img, -50, -50, 10, 10, color.NRGBA{R: 1, A: 1})
	DrawFrame(img, CanvasWidth+10, CanvasHeight+10, 10, 10, color.NRGBA{R: 1, A: 1})
	SetPixel(img, -1, -1, color.NRGBA{})
	SetPixel(img, CanvasWidth, CanvasHeight, color.NRGBA{})
}
