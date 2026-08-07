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
	"time"

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
	// combatAiming 是「按了 A、正在等方向」的中間狀態。
	combatAiming bool
	// dungeonKlimb 是「按了 K、正在等 U/D」的中間狀態。
	dungeonKlimb bool
}

func (g *game) Update() error {
	// 離開語意:F10 / Ctrl+Q 才離開,ESC 永遠是取消(P5 補確認框與自動存檔)。
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControlLeft) || ebiten.IsKeyPressed(ebiten.KeyControlRight)
	if inpututil.IsKeyJustPressed(ebiten.KeyF10) || (ctrl && inpututil.IsKeyJustPressed(ebiten.KeyQ)) {
		// 離開前自動存檔 —— 按錯鍵不該丟進度。
		if dir, err := g.state.WriteSave(g.state.BaseSave); err != nil {
			fmt.Fprintf(os.Stderr, "⚠ 離開前存檔失敗:%v\n", err)
		} else {
			fmt.Printf("已存檔於 %s\n", dir)
		}
		return ebiten.Termination
	}
	st := g.state
	before := len(st.Messages)
	snapshot := g.key()

	// 對話中鍵盤打的是關鍵字,不是指令鍵。ESC 退出對話(不是離開遊戲)。
	// 等方向:方向鍵決定,ESC 作罷(原版的「Direction-」)。
	if st.Prompt == gamestate.PromptDirection {
		for key, dir := range map[ebiten.Key]gamestate.Direction{
			ebiten.KeyArrowUp:    gamestate.North,
			ebiten.KeyArrowDown:  gamestate.South,
			ebiten.KeyArrowLeft:  gamestate.West,
			ebiten.KeyArrowRight: gamestate.East,
		} {
			if inpututil.IsKeyJustPressed(key) {
				st.AnswerDirection(dir)
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			st.CancelDirection()
		}
		g.dirty = true
		return nil
	}

	// 聖壇冥想:打字回答(美德名、真言、獻金),Enter 送出、ESC 作罷。
	if st.Prompt == gamestate.PromptShrine {
		for _, r := range ebiten.AppendInputChars(nil) {
			st.TypeRune(r)
		}
		switch {
		case inpututil.IsKeyJustPressed(ebiten.KeyEnter),
			inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter):
			st.SubmitShrine()
		case inpututil.IsKeyJustPressed(ebiten.KeyBackspace):
			st.Backspace()
		case inpututil.IsKeyJustPressed(ebiten.KeyEscape):
			st.EndMeditate()
		}
		g.dirty = true
		return nil
	}

	// 開場動畫:任意鍵翻頁,ESC 跳過整段。
	if st.Prompt == gamestate.PromptIntro {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			st.SkipIntro()
			g.dirty = true
		} else if len(inpututil.AppendJustPressedKeys(nil)) > 0 {
			st.AdvanceIntro()
			g.dirty = true
		}
		return nil
	}

	// In Quas Wis 的全景:原版畫完就卡著等一個按鍵,按什麼都收起來。
	if st.Prompt == gamestate.PromptPeer {
		if len(inpututil.AppendPressedKeys(nil)) > 0 {
			st.ClosePeer()
			g.dirty = true
		}
		return nil
	}

	// 打咒語名:上古語(含空格),Enter 送出、ESC 作罷。
	if st.Prompt == gamestate.PromptSpell {
		for _, r := range ebiten.AppendInputChars(nil) {
			st.TypeRune(r)
		}
		switch {
		case inpututil.IsKeyJustPressed(ebiten.KeyEnter),
			inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter):
			st.SubmitSpell()
		case inpututil.IsKeyJustPressed(ebiten.KeyBackspace):
			st.Backspace()
		case inpututil.IsKeyJustPressed(ebiten.KeyEscape):
			st.CancelSpell()
		}
		g.dirty = true
		return nil
	}

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

	// 戰鬥中:方向鍵移動、A + 方向攻擊、空白鍵按兵不動、ESC 撤離。
	//
	// 鍵位照原版的玩家指令表(`jpt_A5C8`):A 攻擊、空白 Pass、ESC 離開。
	if st.Prompt == gamestate.PromptCombat {
		dirs := map[ebiten.Key]gamestate.Direction{
			ebiten.KeyArrowUp:    gamestate.North,
			ebiten.KeyArrowDown:  gamestate.South,
			ebiten.KeyArrowLeft:  gamestate.West,
			ebiten.KeyArrowRight: gamestate.East,
		}
		if g.combatAiming {
			// A 按下之後在等方向。
			for key, dir := range dirs {
				if inpututil.IsKeyJustPressed(key) {
					st.CombatAttack(dir)
					g.combatAiming = false
				}
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
				g.combatAiming = false
			}
			g.dirty = true
			return nil
		}
		switch {
		case inpututil.IsKeyJustPressed(ebiten.KeyA):
			g.combatAiming = true
			st.Log("攻擊 —— 哪個方向?")
		case inpututil.IsKeyJustPressed(ebiten.KeyC):
			st.BeginCastPrompt()
		case inpututil.IsKeyJustPressed(ebiten.KeySpace):
			st.CombatPass()
		case inpututil.IsKeyJustPressed(ebiten.KeyEscape):
			st.CombatFlee()
		default:
			for key, dir := range dirs {
				if inpututil.IsKeyJustPressed(key) {
					st.CombatMove(dir)
				}
			}
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
	} else if st.InDungeon() {
		// 地牢:↑ 前進、↓ 後退、← → 轉向,K 爬梯(再按 U/D),C 施法,I 火把。
		//
		// ⚠ 這是照原版的**第一人稱**操作;畫面卻還是俯視的(透視圖要先解
		// DNG1-3.16)。兩者並存看起來有點怪,但規則是對的。
		switch {
		case inpututil.IsKeyJustPressed(ebiten.KeyArrowUp):
			st.DungeonForward(false)
		case inpututil.IsKeyJustPressed(ebiten.KeyArrowDown):
			st.DungeonForward(true)
		case inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft):
			st.DungeonTurn(true)
		case inpututil.IsKeyJustPressed(ebiten.KeyArrowRight):
			st.DungeonTurn(false)
		case inpututil.IsKeyJustPressed(ebiten.KeyK):
			g.dungeonKlimb = true
			st.Log("攀爬 —— 上(U)還是下(D)?")
		case inpututil.IsKeyJustPressed(ebiten.KeyC):
			st.BeginCastPrompt()
		case inpututil.IsKeyJustPressed(ebiten.KeyI):
			st.LightTorch()
		case inpututil.IsKeyJustPressed(ebiten.KeyO):
			st.OpenChest()
		case inpututil.IsKeyJustPressed(ebiten.KeyQ) && !ctrl:
			st.Save()
		}
		if g.dungeonKlimb {
			if inpututil.IsKeyJustPressed(ebiten.KeyU) {
				st.DungeonKlimb(true)
				g.dungeonKlimb = false
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyD) {
				st.DungeonKlimb(false)
				g.dungeonKlimb = false
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
				g.dungeonKlimb = false
			}
		}
		g.dirty = true
		return nil
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
		// C 施法(問上古語咒語名),I 點火把 —— 兩個都照原版鍵位。
		if inpututil.IsKeyJustPressed(ebiten.KeyC) {
			st.BeginCastPrompt()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyI) {
			st.LightTorch()
		}
		// Q 存檔(不離開)。原版 Q 是「存檔並離開」,這裡拆開:
		// 離開走 F10 / Ctrl+Q 並自動存檔,見上面與 esc-cancel-f10-quit-autosave。
		if inpututil.IsKeyJustPressed(ebiten.KeyQ) && !ctrl {
			st.Save()
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
	playIntro := flag.Bool("intro", false, "強制播開場動畫(沒有存檔時本來就會播)")
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
		CombatMaps:   bundle.Combat,
		Stats:        bundle.Stats,
		Spells:       bundle.Spells,
		Story:        bundle.Story,
		Misc:         bundle.Misc,
		Dungeons:     bundle.Dungeons,
		Moons:        bundle.Moons,
		WindDelay:    bundle.WindDelay,
		DungeonRooms: bundle.DngRooms,
		Creatures:    bundle.Creatures,
		UnderObjects: bundle.UnderObjs,
		MaxMessages:  maxMessages,
	}
	st.SeedRandom(time.Now().UnixNano())
	loaded := false
	if sv, from, err := gamestate.FindSave(*gamedata, *saveFile); err == nil {
		loaded = true
		// 開局狀態一律取自存檔:先找設定目錄的進度,再退回原版存檔。
		st.LoadFrom(sv)
		// 同一份進度的物件表(玩家買的馬、放下的船)也要一起讀回來。
		if so, uo := gamestate.FindSaveObjects(); so != nil {
			st.Objects, st.UnderObjects = so, uo
		}
		fmt.Printf("讀取進度:%s\n", from)
		st.Log(fmt.Sprintf("汝已抵達不列顛尼亞 —— %d 年 %d 月 %d 日。",
			st.Clock.Year, st.Clock.Month, st.Clock.Day))
	} else {
		// 沒有存檔時退回「找一塊陸地站上去」,並說明這不是原版的開局。
		st.Clock = gamestate.NewClock()
		st.X, st.Y = assets.FindLandStart(bundle.World, 1)
		st.Log("未載入存檔,由任意陸地起始。")
	}

	// 開場動畫:原版是新遊戲才播。這裡沿用 —— 有進度就直接進遊戲,
	// 不要每次開機都逼玩家看一遍。`-intro` 可以強制播。
	if *playIntro || !loaded {
		st.BeginIntro()
	}

	g := &game{
		state: st,
		scene: &render.Scene{
			State:        st,
			Tiles:        bundle.Tiles,
			Text:         render.NewTextRenderer(bundle.Charset, bundle.CJK, render.ColorText),
			DungeonViews: bundle.DungeonViews,
			DungeonItems: bundle.DungeonItems,
			IntroArt:     bundle.IntroArt,
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
