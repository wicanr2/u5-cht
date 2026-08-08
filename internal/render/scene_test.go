package render

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
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

// TestDungeonFirstPersonIsSymmetric:透視畫面必須左右對稱。
//
// 每張切片都是**半邊**,右半是水平鏡射。只要 x 算錯一格(off-by-one)
// 或鏡射公式寫錯,對稱就破了 —— 而畫面看起來仍然「像個走廊」,
// 光用眼睛很難發現。這條用像素比對把它釘死。
//
// ⚠ 只在「左右兩側的地牢格子相同」時才成立,所以先找一個左右對稱的位置。
func TestDungeonFirstPersonIsSymmetric(t *testing.T) {
	sc, st := dungeonScene(t)
	if sc == nil {
		return
	}
	d := st.Dungeon
	// 找一個左右鄰格種類相同的朝向,不然畫面本來就不該對稱。
	fdx, fdy := d.Facing.Delta()
	sdx, sdy := -fdy, fdx
	l := st.DungeonTileAt(u5data.DungeonWrap(d.X+sdx), u5data.DungeonWrap(d.Y+sdy))
	r := st.DungeonTileAt(u5data.DungeonWrap(d.X-sdx), u5data.DungeonWrap(d.Y-sdy))
	if u5data.DungeonSideShape(l, 0) != u5data.DungeonSideShape(r, 0) {
		t.Skip("入口左右兩側不同,這一格測不了對稱")
	}
	img := sc.Render()
	ox := MapOriginX + (ViewPixels-u5data.DungeonViewWidth*DungeonViewScale)/2
	oy := MapOriginY + (ViewPixels-u5data.DungeonViewHeight*DungeonViewScale)/2
	w := u5data.DungeonViewWidth * DungeonViewScale
	bad := 0
	for y := 0; y < u5data.DungeonViewHeight*DungeonViewScale; y++ {
		for x := 0; x < w/2; x++ {
			a := img.NRGBAAt(ox+x, oy+y)
			b := img.NRGBAAt(ox+w-1-x, oy+y)
			if a != b {
				if bad < 3 {
					t.Errorf("(%d,%d) 是 %v,鏡射位置是 %v", x, y, a, b)
				}
				bad++
			}
		}
	}
	if bad != 0 {
		t.Fatalf("%d 個像素左右不對稱", bad)
	}
}

// TestDungeonWithoutLightIsBlack:沒光就是全黑。
//
// `sub_3D14` 一開頭 `if (!byte_3E0B6 && !byte_3E0B7)` 直接畫一個黑框收工。
// 這條擋的是「忘了實作黑暗,玩家沒火把也看得見整條走廊」。
func TestDungeonWithoutLightIsBlack(t *testing.T) {
	sc, st := dungeonScene(t)
	if sc == nil {
		return
	}
	st.LightTurns, st.TorchTurns = 0, 0
	img := sc.Render()
	ox := MapOriginX + (ViewPixels-u5data.DungeonViewWidth*DungeonViewScale)/2
	oy := MapOriginY + (ViewPixels-u5data.DungeonViewHeight*DungeonViewScale)/2
	for y := 0; y < u5data.DungeonViewHeight*DungeonViewScale; y += 7 {
		for x := 0; x < u5data.DungeonViewWidth*DungeonViewScale; x += 7 {
			if c := img.NRGBAAt(ox+x, oy+y); c != EGABlack {
				t.Fatalf("沒有光,(%d,%d) 卻是 %v", x, y, c)
			}
		}
	}
	// 有光就不該是全黑 —— 不然上面那條測的只是「畫面本來就黑」。
	st.TorchTurns = 100
	img = sc.Render()
	lit := false
	for y := 0; y < u5data.DungeonViewHeight*DungeonViewScale && !lit; y += 3 {
		for x := 0; x < u5data.DungeonViewWidth*DungeonViewScale; x += 3 {
			if img.NRGBAAt(ox+x, oy+y) != EGABlack {
				lit = true
				break
			}
		}
	}
	if !lit {
		t.Error("點了火把畫面還是全黑")
	}
}

// dungeonScene 準備一個「已經進到地牢裡、火把點著」的場景。
// 沒有原版素材就 skip,回傳 nil。
func dungeonScene(t *testing.T) (*Scene, *game.State) {
	t.Helper()
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA")
		return nil, nil
	}
	views := make([]u5data.PictureSet, 0, u5data.DungeonThemes)
	for i := 1; i <= u5data.DungeonThemes; i++ {
		set, err := u5data.LoadPictures(filepath.Join(dir, fmt.Sprintf("DNG%d.16", i)))
		if err != nil {
			t.Fatal(err)
		}
		views = append(views, set)
	}
	dg, err := u5data.LoadDungeons(dir)
	if err != nil {
		t.Fatal(err)
	}
	sc := testScene()
	sc.DungeonViews = views
	st := sc.State
	st.Dungeons = dg
	if !st.EnterDungeon(0, false) {
		t.Fatal("進不了地牢")
	}
	st.TorchTurns = 100
	return sc, st
}

// TestMoongateGrowsFromTheBottom —— ★ 月門的長出動畫(原版 `sub_265F0`)。
//
// 第 f 格 = 「草地 + 月門圖的前 f 列疊在第 (16−f) 列起」。所以:
//
//	f = 0 或 f >= 16 → 整格月門(一般畫法)
//	f = 1..15        → 上面 (16−f) 列是草地、下面 f 列是門
//
// ⚠ 這條測試比對的是**列的來源**,不是「有沒有變化」—— 只驗「畫面不一樣」
// 會讓「疊反了(門在上、草在下)」照樣過關。
func TestMoongateGrowsFromTheBottom(t *testing.T) {
	tiles := make([]u5data.Tile, 256)
	// 草地全填色號 1、月門每一列填「列號 + 2」,這樣就分得出來源是哪一列。
	for p := range tiles[u5data.MoongateClosedTile].Pix {
		tiles[u5data.MoongateClosedTile].Pix[p] = 1
	}
	for ty := 0; ty < u5data.TileSize; ty++ {
		for tx := 0; tx < u5data.TileSize; tx++ {
			tiles[u5data.MoongateOpenTile].Pix[ty*u5data.TileSize+tx] = byte(ty + 2)
		}
	}
	world := &u5data.WorldMap{}
	for i := range world.Tiles {
		world.Tiles[i] = u5data.MoongateOpenTile
	}
	s := &Scene{
		State: &game.State{World: world, X: 100, Y: 100},
		Tiles: tiles,
		Text:  NewTextRenderer(nil, nil, ColorText),
	}

	// 挑**非中央**的一格 —— 中央那格上面疊著玩家標記(ColorMarker),
	// 量到的會是白色而不是地形。第一版量中央,紅燈看起來像疊反了。
	half := ViewTiles / 2
	cx := MapOriginX + (half+2)*TilePixels
	cy := MapOriginY + half*TilePixels
	rowColor := func(img *image.NRGBA, ty int) color.Color {
		return img.At(cx+1, cy+ty*TileScale+1)
	}

	for _, f := range []int{1, 5, 15} {
		s.State.MoongateFrame = f
		img := s.Render()
		top := u5data.TileSize - f
		for ty := 0; ty < u5data.TileSize; ty++ {
			want := u5data.EGAPalette[1] // 草地
			why := "草地"
			if ty >= top {
				want = u5data.EGAPalette[(ty-top+2)&0x0F]
				why = fmt.Sprintf("月門圖第 %d 列", ty-top)
			}
			if got := rowColor(img, ty); got != color.Color(want) {
				t.Fatalf("f=%d 第 %d 列該是%s:得到 %v,預期 %v", f, ty, why, got, want)
			}
		}
	}

	// f = 0 與滿格都走一般畫法 = 整格月門(第 0 列就是門的第 0 列)。
	for _, f := range []int{0, game.MoongateFrameMax} {
		s.State.MoongateFrame = f
		img := s.Render()
		if got, want := rowColor(img, 0), color.Color(u5data.EGAPalette[2]); got != want {
			t.Errorf("f=%d 該畫完整的月門:第 0 列得到 %v,預期 %v", f, got, want)
		}
	}
}
