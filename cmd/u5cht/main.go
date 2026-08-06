// u5cht 是 Ultima V 重製版(繁體中文)的執行檔。
//
// 目前是 P2 垂直切片:載入原版 tileset 與世界地圖,用 11×11 地圖視窗走 Britannia,
// HUD 與訊息欄走倚天中文點陣字。遊戲邏輯(碰撞、時間、NPC、戰鬥)在 P4 之後。
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/wicanr2/u5-cht/internal/cjk"
	"github.com/wicanr2/u5-cht/internal/render"
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 邏輯畫布 = 原版 320×200 的乾淨 2 倍(見 CLAUDE.md §3、rulebook/81)。
const (
	logicalWidth  = 640
	logicalHeight = 400

	mapOriginX = 8
	mapOriginY = 8

	panelX      = mapOriginX + render.ViewPixels + 8 // 右側狀態欄
	messageY    = mapOriginY + render.ViewPixels + 6 // 下方訊息欄
	maxMessages = 2
)

var (
	colorBackground = color.NRGBA{R: 0x10, G: 0x10, B: 0x28, A: 0xFF}
	colorText       = color.NRGBA{R: 0xE8, G: 0xE8, B: 0xD8, A: 0xFF}
	colorMarker     = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
)

// version 由建置時的 -ldflags 注入。
var version = "dev"

type game struct {
	canvas *ebiten.Image

	tiles *render.TileSet
	world *u5data.WorldMap
	text  *render.TextRenderer

	px, py   int      // 玩家在世界地圖的座標
	messages []string // 訊息欄(最新在後)

	// headless 截圖
	headless bool
	shotPath string
	frame    int
	done     bool
}

func (g *game) log(format string, args ...any) {
	g.messages = append(g.messages, fmt.Sprintf(format, args...))
	if len(g.messages) > maxMessages {
		g.messages = g.messages[len(g.messages)-maxMessages:]
	}
}

func (g *game) Update() error {
	if g.done {
		return ebiten.Termination
	}
	// 離開語意:F10 / Ctrl+Q 才離開,ESC 永遠是取消(P4 補確認框與自動存檔)。
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControlLeft) || ebiten.IsKeyPressed(ebiten.KeyControlRight)
	if inpututil.IsKeyJustPressed(ebiten.KeyF10) || (ctrl && inpututil.IsKeyJustPressed(ebiten.KeyQ)) {
		return ebiten.Termination
	}

	// 方向鍵移動。碰撞與地形效果屬於遊戲邏輯,P4 才做 —— 現在誠實地讓玩家走過所有地形。
	type move struct {
		key    ebiten.Key
		dx, dy int
		name   string
	}
	for _, m := range []move{
		{ebiten.KeyArrowUp, 0, -1, "北"},
		{ebiten.KeyArrowDown, 0, 1, "南"},
		{ebiten.KeyArrowLeft, -1, 0, "西"},
		{ebiten.KeyArrowRight, 1, 0, "東"},
	} {
		if inpututil.IsKeyJustPressed(m.key) {
			g.px = wrapCoord(g.px+m.dx)
			g.py = wrapCoord(g.py+m.dy)
			g.log("往%s方前行。", m.name)
		}
	}
	return nil
}

func wrapCoord(v int) int {
	v %= u5data.WorldSide
	if v < 0 {
		v += u5data.WorldSide
	}
	return v
}

func (g *game) Draw(screen *ebiten.Image) {
	if g.canvas == nil {
		g.canvas = ebiten.NewImage(logicalWidth, logicalHeight)
	}
	g.canvas.Fill(colorBackground)

	// 地圖視窗
	if g.tiles != nil && g.world != nil {
		g.tiles.DrawWorldView(g.canvas, g.world, g.px, g.py, mapOriginX, mapOriginY)
		// 玩家位置標記。
		// TODO(P3):換成原版的 Avatar tile —— 索引要從反編譯碼確認,不猜。
		cx := mapOriginX + (render.ViewTiles/2)*render.TilePixels
		cy := mapOriginY + (render.ViewTiles/2)*render.TilePixels
		drawFrame(g.canvas, cx, cy, render.TilePixels, render.TilePixels, colorMarker)
	}

	// 右側狀態欄(全中文)
	if g.text != nil {
		y := mapOriginY
		g.text.Draw(g.canvas, panelX, y, "創世紀 V")
		y += render.LineHeight
		g.text.Draw(g.canvas, panelX, y, "命運勇士")
		y += render.LineHeight * 2
		g.text.Draw(g.canvas, panelX, y, fmt.Sprintf("座標 %3d,%3d", g.px, g.py))
		y += render.LineHeight
		if g.world != nil {
			g.text.Draw(g.canvas, panelX, y, fmt.Sprintf("地形 %3d", g.world.At(g.px, g.py)))
		}
		y += render.LineHeight * 2
		g.text.Draw(g.canvas, panelX, y, "方向鍵移動")
		y += render.LineHeight
		g.text.Draw(g.canvas, panelX, y, "F10 離開")

		// 下方訊息欄:斷行用的是同一個 Advance,所以不會溢框
		msgY := messageY
		for _, m := range g.messages {
			for _, line := range render.Wrap(m, logicalWidth-2*mapOriginX) {
				g.text.Draw(g.canvas, mapOriginX, msgY, line)
				msgY += render.LineHeight
			}
		}
	}

	// 邏輯畫布 → 視窗:整數倍 nearest,pixel art 不糊。
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	scale := min(sw/logicalWidth, sh/logicalHeight)
	if scale < 1 {
		scale = 1
	}
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
	op.GeoM.Scale(float64(scale), float64(scale))
	op.GeoM.Translate(float64((sw-logicalWidth*scale)/2), float64((sh-logicalHeight*scale)/2))
	screen.DrawImage(g.canvas, op)

	// headless:等資源就緒後截一張圖就結束(給 CI 與逐畫面比對用)
	g.frame++
	if g.headless && g.frame >= 2 && !g.done {
		if err := savePNG(g.shotPath, g.canvas); err != nil {
			fmt.Fprintf(os.Stderr, "截圖失敗:%v\n", err)
		} else {
			fmt.Printf("✓ 截圖 → %s\n", g.shotPath)
		}
		g.done = true
	}
}

// Layout 回傳視窗實際大小 —— 縮放由 Draw 自己管,才能保證 nearest 而非線性。
func (g *game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func drawFrame(dst *ebiten.Image, x, y, w, h int, c color.NRGBA) {
	for i := 0; i < w; i++ {
		dst.Set(x+i, y, c)
		dst.Set(x+i, y+h-1, c)
	}
	for j := 0; j < h; j++ {
		dst.Set(x, y+j, c)
		dst.Set(x+w-1, y+j, c)
	}
}

func savePNG(path string, img *ebiten.Image) error {
	b := img.Bounds()
	buf := make([]byte, 4*b.Dx()*b.Dy())
	img.ReadPixels(buf)
	out := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	copy(out.Pix, buf)
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, out)
}

// findLandStart 從地圖中央往外找一個非水的落點。
//
// 這是「正常玩家路徑」的最小版本:玩家不該開場就泡在海裡。
// 完整的可達性檢查(flood-fill 連通分量、城鎮與落點同分量)在 P4 —— 那個坑 u2-cht 踩過:
// 回歸測試全綠,但新角色被放在只連城堡的 12 格小島上 soft-lock。
func findLandStart(w *u5data.WorldMap, waterTile byte) (int, int) {
	const c = u5data.WorldSide / 2
	for r := 0; r < u5data.WorldSide/2; r++ {
		for dy := -r; dy <= r; dy++ {
			for dx := -r; dx <= r; dx++ {
				if max(abs(dx), abs(dy)) != r {
					continue // 只掃這一圈
				}
				x, y := wrapCoord(c+dx), wrapCoord(c+dy)
				if t := w.At(x, y); t != waterTile && t != 0 {
					return x, y
				}
			}
		}
	}
	return c, c
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func main() {
	gamedata := flag.String("gamedata", "gamedata",
		"原版 Ultima V(DOS 版)資料目錄;版權素材由玩家自備,不隨本專案散布")
	fmtowns := flag.String("fmtowns", "re_work/fmtowns/iso/U5_E",
		"FM Towns 版 U5_E 目錄(未壓縮 tileset 來源)")
	fontPrefix := flag.String("font", "assets/fonts/eten-15",
		"倚天中文點陣字 atlas 前綴(用 tools/dev.sh font 15 產生)")
	scale := flag.Int("scale", 2, "視窗放大倍率(整數;邏輯畫布固定 640×400)")
	headless := flag.Bool("headless", false, "不等使用者操作,畫一幀截圖後結束")
	shot := flag.String("shot", "build/shot.png", "headless 模式的截圖輸出路徑")
	showVersion := flag.Bool("version", false, "印出版本後結束")
	flag.Parse()

	if *showVersion {
		fmt.Printf("u5cht %s\n", version)
		return
	}

	g := &game{headless: *headless, shotPath: *shot}

	// --- 原版素材(缺了就優雅降級並明說,不拿自製素材充數;CLAUDE.md §3.0)---
	charset, err := u5data.LoadCharset(filepath.Join(*gamedata, "IBM.CH"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "找不到原版字型:%v\n", err)
	}

	tiles, err := u5data.LoadFMTownsTileSet([]string{
		filepath.Join(*fmtowns, "EGA0.TIL"),
		filepath.Join(*fmtowns, "EGA1.TIL"),
		filepath.Join(*fmtowns, "EGA2.TIL"),
		filepath.Join(*fmtowns, "EGA3.TIL"),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "載入 tileset 失敗:%v\n", err)
	} else {
		g.tiles = render.NewTileSet(tiles)
		fmt.Printf("已載入 %d 個 tile\n", len(tiles))
	}

	if chunks, err := u5data.LoadChunks(filepath.Join(*gamedata, "BRIT.DAT"), u5data.ChunkSide); err != nil {
		fmt.Fprintf(os.Stderr, "載入 BRIT.DAT 失敗:%v\n", err)
	} else if ovl, err := os.ReadFile(filepath.Join(*gamedata, "DATA.OVL")); err != nil {
		fmt.Fprintf(os.Stderr, "讀 DATA.OVL 失敗:%v\n", err)
	} else if index, err := u5data.ReadWorldChunkIndex(ovl); err != nil {
		fmt.Fprintf(os.Stderr, "取 chunk 索引表失敗:%v\n", err)
	} else if world, err := u5data.BuildWorldMap(chunks, index, 1); err != nil {
		fmt.Fprintf(os.Stderr, "組世界地圖失敗:%v\n", err)
	} else {
		g.world = world
		g.px, g.py = findLandStart(world, 1)
		fmt.Printf("世界地圖就緒(%d chunk),落點 (%d, %d)\n", len(chunks), g.px, g.py)
	}

	cjkFont, err := cjk.Load(*fontPrefix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "載入中文字型失敗:%v\n", err)
	} else {
		fmt.Printf("已載入倚天中文點陣字(%d 字)\n", cjkFont.Count())
	}

	// Ebiten 的紋理要在遊戲迴圈啟動後才建立才安全,但 NewImageFromImage 在 RunGame
	// 之前呼叫也可以(延後配置);為求穩妥,文字繪製器在這裡建立即可。
	g.text = render.NewTextRenderer(charset, cjkFont, colorText)
	g.log("汝已抵達不列顛尼亞。")

	if *scale < 1 {
		*scale = 1
	}
	ebiten.SetWindowSize(logicalWidth**scale, logicalHeight**scale)
	ebiten.SetWindowTitle("創世紀 V:命運勇士 — 繁體中文版")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(g); err != nil {
		fmt.Fprintf(os.Stderr, "執行失敗:%v\n", err)
		os.Exit(1)
	}
}
