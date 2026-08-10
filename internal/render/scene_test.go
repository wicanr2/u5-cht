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

// ★ 隊伍畫在正中央,而且畫的是**載具那一頁的 tile**,不是白色方框。
//
// 這一條是對 DOSBox 原版並排比對之後改的(`docs/playtest-findings.md`):
// 原版中央是聖者的小人,站在城鎮 tile 上時小人蓋住那格地形;
// 引擎原本畫一個 `ColorMarker` 空心框當替代品,而那在並排時最顯眼 ——
// 看起來像選取狀態,不像角色。
//
// ⚠ 這裡驗的是「中央那一格畫的是 `NPCTileBase + 載具碼` 那個 tile」,
// 不是「中央有東西」—— 後者連白框都能通過。
func TestScenePartySpriteAtCenter(t *testing.T) {
	s := testScene()
	// testScene 的 tile 是人工造的漸層色塊,第 n 個 tile 整格都是色號 n%16。
	// 隊伍那一格因此應該整格同色,而且色號對得上載具碼。
	want := u5data.NPCTileBase + int(s.State.Transport)
	img := s.Render()
	half := ViewTiles / 2
	mx := MapOriginX + half*TilePixels
	my := MapOriginY + half*TilePixels
	if want >= len(s.Tiles) {
		// 假 tileset 放不到那個索引 —— 那就至少要確認**不是**白框了。
		for _, p := range [][2]int{{mx, my}, {mx + TilePixels - 1, my}} {
			if got := img.NRGBAAt(p[0], p[1]); got == ColorMarker {
				t.Errorf("(%d,%d) 還是白色方框 —— 原版沒有這個框", p[0], p[1])
			}
		}
		return
	}
	exp := u5data.EGAPalette[s.Tiles[want].At(0, 0)&0x0F]
	if got := img.NRGBAAt(mx+TilePixels/2, my+TilePixels/2); got != exp {
		t.Errorf("中央那一格是 %v,預期畫 tile %d(%v)", got, want, exp)
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

// ★★ 隊伍那一欄要畫**狀態**。
//
// 原版那一欄是 `Elwood     60G` —— 名字、HP、然後一個狀態字母;
// 本專案把字母展開成中文說明(使用者指示 2026-08-09)。
//
// ⚠ **這裡不能用「健康 vs 中毒 兩張圖不同」當判準**:測試用的 `TextRenderer`
// 沒有字庫(render 的測試不依賴原版資料),兩個中文詞都會畫成同樣的缺字方框
// ⇒ 兩張圖一模一樣,而那不代表狀態沒畫。改成兩件事各驗一次:
//
//  1. 狀態欄那一段 x 範圍**真的被畫過**(有隊員 vs 沒隊員的差異要涵蓋它)
//  2. 五個狀態碼的中文說明**互不相同**、也不是 `?`(純邏輯,不必字庫)
func TestPartyLineShowsStatus(t *testing.T) {
	withParty := func(n int) *image.NRGBA {
		s := testScene()
		s.State.Roster = []u5data.Character{{Name: "Elwood", HP: 60, MaxHP: 60, Status: u5data.StatusGood}}
		s.State.PartySize = n
		return s.Render()
	}
	empty, one := withParty(0), withParty(1)
	x0, x1 := PanelX+partyStatusColumn, PanelX+partyStatusColumn+CJKAdvance*2
	painted := false
	for y := 0; y < CanvasHeight && !painted; y++ {
		for x := x0; x < x1; x++ {
			if empty.NRGBAAt(x, y) != one.NRGBAAt(x, y) {
				painted = true
				break
			}
		}
	}
	if !painted {
		t.Errorf("狀態欄(x %d..%d)完全沒有被畫過", x0, x1)
	}
}

// 五個狀態碼各有各的說明,而且沒有一個掉到 `?`。
//
// ⚠ 狀態碼是**可讀字母**('G'/'P'/'D'/'C'/'S'),原版直接
// `cmp byte_3DDBF[32*i], 'P'` 這樣比 —— 譯名換了不該影響那些判定,
// 所以這條只驗顯示用的說明,不碰常數本身。
func TestEveryStatusCodeHasItsOwnDescription(t *testing.T) {
	codes := []byte{u5data.StatusGood, u5data.StatusPoisoned, u5data.StatusDead,
		u5data.StatusCharmed, u5data.StatusAsleep}
	seen := map[string]byte{}
	for _, c := range codes {
		name := u5data.StatusName(c)
		if name == "" || name == "?" {
			t.Errorf("狀態 %q 沒有說明(得到 %q)", string(rune(c)), name)
			continue
		}
		if prev, dup := seen[name]; dup {
			t.Errorf("狀態 %q 與 %q 共用說明 %q", string(rune(c)), string(rune(prev)), name)
		}
		seen[name] = c
	}
}
