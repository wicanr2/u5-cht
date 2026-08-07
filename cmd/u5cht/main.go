// u5cht 是 Ultima V 重製版(繁體中文)的執行檔。
//
// 目前可以走大地圖、按 E 進城鎮 / 城堡 / 燈塔、在場景裡走動與上下樓、走到邊界離開。
// 規則全部照原版執行檔(見 internal/game 的套件說明)。時間、NPC、戰鬥尚未實作。
//
// 架構:遊戲規則在 internal/game(純邏輯)、畫面在 internal/render(純 CPU),
// 這一層只負責把成品上傳成紋理、整數倍放大顯示、把按鍵轉成 game 的動作。
// 兩者都不依賴 ebiten,所以規則與畫面都能 headless 驗證。
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/wicanr2/u5-cht/internal/assets"
	gamestate "github.com/wicanr2/u5-cht/internal/game"
	"github.com/wicanr2/u5-cht/internal/render"
)

const maxMessages = 8

// version 由建置時的 -ldflags 注入。
var version = "dev"

type game struct {
	scene *render.Scene
	state *gamestate.State

	tex   *ebiten.Image // 上傳到 GPU 的畫面
	dirty bool          // 畫面是否需要重畫(回合制遊戲多數幀不需要)
}

func (g *game) Update() error {
	// 離開語意:F10 / Ctrl+Q 才離開,ESC 永遠是取消(P5 補確認框與自動存檔)。
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControlLeft) || ebiten.IsKeyPressed(ebiten.KeyControlRight)
	if inpututil.IsKeyJustPressed(ebiten.KeyF10) || (ctrl && inpututil.IsKeyJustPressed(ebiten.KeyQ)) {
		return ebiten.Termination
	}
	st := g.state
	before := len(st.Messages)
	snapshot := g.key()

	// 對話中鍵盤打的是關鍵字,不是指令鍵。ESC 退出對話(不是離開遊戲)。
	if st.Prompt == gamestate.PromptTalk || st.Prompt == gamestate.PromptAnswer {
		for _, r := range ebiten.AppendInputChars(nil) {
			st.TypeRune(r)
			g.dirty = true
		}
		switch {
		case inpututil.IsKeyJustPressed(ebiten.KeyEnter),
			inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter):
			st.Submit()
		case inpututil.IsKeyJustPressed(ebiten.KeyBackspace):
			st.Backspace()
		case inpututil.IsKeyJustPressed(ebiten.KeyEscape):
			st.EndConversation()
		}
		g.dirty = true
		return nil
	}

	// 在店裡:每一步都是等一個字母鍵(選單的 a/b/c… 或 Y/N),ESC 走人。
	if st.Prompt == gamestate.PromptShop {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			st.LeaveShop()
		} else {
			for _, r := range ebiten.AppendInputChars(nil) {
				st.ShopChoose(r)
			}
		}
		g.dirty = true
		return nil
	}

	// 有提問待答時只收 Y / N / ESC —— 原版 sub_86C 的 do-while 就是這個行為,
	// ESC 等同於 N。
	if st.Prompt != gamestate.PromptNone {
		switch {
		case inpututil.IsKeyJustPressed(ebiten.KeyY):
			st.Answer(true)
		case inpututil.IsKeyJustPressed(ebiten.KeyN), inpututil.IsKeyJustPressed(ebiten.KeyEscape):
			st.Answer(false)
		}
	} else {
		if inpututil.IsKeyJustPressed(ebiten.KeyE) {
			st.Enter()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyK) {
			st.Klimb()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyT) {
			st.Talk()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyB) {
			st.Board()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyX) {
			st.Exit()
		}
		for key, dir := range map[ebiten.Key]gamestate.Direction{
			ebiten.KeyArrowUp:    gamestate.North,
			ebiten.KeyArrowDown:  gamestate.South,
			ebiten.KeyArrowLeft:  gamestate.West,
			ebiten.KeyArrowRight: gamestate.East,
		} {
			if inpututil.IsKeyJustPressed(key) {
				st.Move(dir)
			}
		}
	}

	if len(st.Messages) != before || g.key() != snapshot {
		g.dirty = true
	}
	return nil
}

// key 是「畫面該不該重畫」的判斷依據 —— 回合制遊戲多數幀什麼都沒變。
func (g *game) key() [6]int {
	st := g.state
	return [6]int{st.X, st.Y, st.Location, st.Floor, st.Clock.Hour*60 + st.Clock.Minute, int(st.Prompt)}
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
	saveFile := flag.String("save", "",
		"要載入的存檔;留空則依序試 gamedata/SAVED.GAM、INIT.GAM")
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
		SaveFile:   *saveFile,
	})
	for _, w := range warns {
		fmt.Fprintf(os.Stderr, "⚠ %s\n", w)
	}

	st := &gamestate.State{
		World:        bundle.World,
		Under:        bundle.Under,
		Scenes:       bundle.Scenes,
		NPCs:         bundle.NPCs,
		Talks:        bundle.Talks,
		Shops:        bundle.Shops,
		Items:        bundle.Items,
		Objects:      bundle.Objects,
		UnderObjects: bundle.UnderObjs,
		MaxMessages:  maxMessages,
	}
	if bundle.Save != nil {
		// 開局狀態一律取自原版存檔:時間、隊伍、位置都不是自己編的。
		st.LoadFrom(bundle.Save)
		st.Log(fmt.Sprintf("汝已抵達不列顛尼亞 —— %d 年 %d 月 %d 日。",
			st.Clock.Year, st.Clock.Month, st.Clock.Day))
	} else {
		// 沒有存檔時退回「找一塊陸地站上去」,並說明這不是原版的開局。
		st.Clock = gamestate.NewClock()
		st.X, st.Y = assets.FindLandStart(bundle.World, 1)
		st.Log("未載入存檔,由任意陸地起始。")
	}

	g := &game{
		state: st,
		scene: &render.Scene{
			State: st,
			Tiles: bundle.Tiles,
			Text:  render.NewTextRenderer(bundle.Charset, bundle.CJK, render.ColorText),
		},
	}

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
