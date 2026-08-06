// u5cht 是 Ultima V 重製版(繁體中文)的執行檔。
//
// 目前是 P2 垂直切片:載入原版 tileset 與世界地圖,用 11×11 地圖視窗走 Britannia,
// HUD 與訊息欄走倚天中文點陣字。遊戲邏輯(碰撞、時間、NPC、戰鬥)在 P4 之後。
//
// 架構:畫面完全由 internal/render 在 CPU 上組好(不依賴 ebiten),
// 這一層只負責把成品上傳成紋理、整數倍放大顯示、收鍵盤。
// 理由見 internal/render 的套件說明(headless 驗證不該需要 GPU)。
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/wicanr2/u5-cht/internal/assets"
	"github.com/wicanr2/u5-cht/internal/render"
	"github.com/wicanr2/u5-cht/internal/u5data"
)

const maxMessages = 2

// version 由建置時的 -ldflags 注入。
var version = "dev"

type game struct {
	scene *render.Scene

	tex   *ebiten.Image // 上傳到 GPU 的畫面
	dirty bool          // 畫面是否需要重畫(回合制遊戲多數幀不需要)
}

func (g *game) log(format string, args ...any) {
	g.scene.Messages = append(g.scene.Messages, fmt.Sprintf(format, args...))
	if len(g.scene.Messages) > maxMessages {
		g.scene.Messages = g.scene.Messages[len(g.scene.Messages)-maxMessages:]
	}
	g.dirty = true
}

func (g *game) Update() error {
	// 離開語意:F10 / Ctrl+Q 才離開,ESC 永遠是取消(P4 補確認框與自動存檔)。
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControlLeft) || ebiten.IsKeyPressed(ebiten.KeyControlRight)
	if inpututil.IsKeyJustPressed(ebiten.KeyF10) || (ctrl && inpututil.IsKeyJustPressed(ebiten.KeyQ)) {
		return ebiten.Termination
	}

	// 方向鍵移動。碰撞與地形效果屬遊戲邏輯,P4 才做 —— 現在誠實地讓玩家走過所有地形。
	for _, m := range []struct {
		key    ebiten.Key
		dx, dy int
		name   string
	}{
		{ebiten.KeyArrowUp, 0, -1, "北"},
		{ebiten.KeyArrowDown, 0, 1, "南"},
		{ebiten.KeyArrowLeft, -1, 0, "西"},
		{ebiten.KeyArrowRight, 1, 0, "東"},
	} {
		if !inpututil.IsKeyJustPressed(m.key) {
			continue
		}
		nx := render.WrapCoord(g.scene.PX + m.dx)
		ny := render.WrapCoord(g.scene.PY + m.dy)
		// 通行規則取自原版執行檔的 tile bitmap(u5data.TileBlocksWalking),
		// 不是自己定的 —— 見 docs/re/02。
		if g.scene.World != nil && u5data.TileBlocksWalking(int(g.scene.World.At(nx, ny))) {
			g.log("受阻!")
			continue
		}
		g.scene.PX, g.scene.PY = nx, ny
		g.log("往%s方前行。", m.name)
	}
	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	if g.tex == nil || g.dirty {
		g.tex = ebiten.NewImageFromImage(g.scene.Render())
		g.dirty = false
	}
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	scale := min(sw/render.CanvasWidth, sh/render.CanvasHeight)
	if scale < 1 {
		scale = 1
	}
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
	op.GeoM.Scale(float64(scale), float64(scale))
	op.GeoM.Translate(
		float64((sw-render.CanvasWidth*scale)/2),
		float64((sh-render.CanvasHeight*scale)/2),
	)
	screen.DrawImage(g.tex, op)
}

// Layout 回傳視窗實際大小 —— 縮放由 Draw 自己管,才能保證 nearest 而非線性。
func (g *game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func main() {
	gamedata := flag.String("gamedata", "gamedata",
		"原版 Ultima V(DOS 版)資料目錄;版權素材由玩家自備,不隨本專案散布")
	fmtowns := flag.String("fmtowns", "re_work/fmtowns/iso/U5_E",
		"FM Towns 版 U5_E 目錄(未壓縮 tileset 來源)")
	fontPrefix := flag.String("font", "assets/fonts/eten-15",
		"倚天中文點陣字 atlas 前綴(用 tools/dev.sh font 15 產生)")
	scale := flag.Int("scale", 2, "視窗放大倍率(整數;邏輯畫布固定 640×400)")
	showVersion := flag.Bool("version", false, "印出版本後結束")
	flag.Parse()

	if *showVersion {
		fmt.Printf("u5cht %s\n", version)
		return
	}

	// 素材缺件不致命:優雅降級並明說缺什麼(CLAUDE.md §3.0)。
	// headless 截圖請用 `u5dump scene`(純 CPU,不需要 GPU)。
	bundle, warns := assets.Load(assets.Options{
		GameData:   *gamedata,
		FMTowns:    *fmtowns,
		FontPrefix: *fontPrefix,
	})
	for _, w := range warns {
		fmt.Fprintf(os.Stderr, "⚠ %s\n", w)
	}

	g := &game{scene: &render.Scene{
		World: bundle.World,
		Tiles: bundle.Tiles,
		Text:  render.NewTextRenderer(bundle.Charset, bundle.CJK, render.ColorText),
	}}
	g.scene.PX, g.scene.PY = assets.FindLandStart(bundle.World, 1)
	g.log("汝已抵達不列顛尼亞。")

	if *scale < 1 {
		*scale = 1
	}
	ebiten.SetWindowSize(render.CanvasWidth**scale, render.CanvasHeight**scale)
	ebiten.SetWindowTitle("創世紀 V:命運勇士 — 繁體中文版")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(g); err != nil {
		fmt.Fprintf(os.Stderr, "執行失敗:%v\n", err)
		os.Exit(1)
	}
}
