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
	"github.com/wicanr2/u5-cht/internal/audio"
	gamestate "github.com/wicanr2/u5-cht/internal/game"
	"github.com/wicanr2/u5-cht/internal/u5data"
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
	// music 把 `state.CurrentSong()` 同步到實際播放。可為 nil(沒有音樂資料)。
	music *audio.Player
}

func (g *game) Update() error {
	// 配樂:`internal/game` 只決定曲號,這裡把它同步到播放層(`docs/re/87`)。
	// ⚠ 每一帧呼叫是刻意的 —— `Player.Update` 對「同一首」是 no-op,
	// 而換曲點散在引擎各處(進城、開戰、上船…),用輪詢比到處插回呼乾淨。
	if g.music != nil {
		g.music.Update(g.state.CurrentSong())
	}
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
	// 燈塔的光束每一幀往前掃一格(原版 `sub_2E944` 每次重畫都做一次)。
	// ⚠ 只有夜裡、而且場景裡真的有燈塔時才看得出來 —— `applyBeam` 自己會判。
	st.AdvanceBeam()
	g.dirty = true
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

	// Yell 的「喊什麼?」:打一個力量之言或暗影君主的名字,Enter 送出、ESC 作罷。
	if st.Prompt == gamestate.PromptYell {
		for _, r := range ebiten.AppendInputChars(nil) {
			st.TypeRune(r)
		}
		switch {
		case inpututil.IsKeyJustPressed(ebiten.KeyEnter),
			inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter):
			st.SubmitYell()
		case inpututil.IsKeyJustPressed(ebiten.KeyBackspace):
			st.Backspace()
		case inpututil.IsKeyJustPressed(ebiten.KeyEscape):
			st.CancelYell()
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

	// 黑棘的審問:打字回答,Enter 送出。**沒有 ESC** —— 汝逃不掉。
	if st.Prompt == gamestate.PromptBlackthorn {
		for _, r := range ebiten.AppendInputChars(nil) {
			st.TypeRune(r)
		}
		switch {
		case inpututil.IsKeyJustPressed(ebiten.KeyEnter),
			inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter):
			st.SubmitBlackthorn()
		case inpututil.IsKeyJustPressed(ebiten.KeyBackspace):
			st.Backspace()
		}
		g.dirty = true
		return nil
	}

	// 結局:先答一個 Y / N,之後按任意鍵翻頁。**沒有 ESC** —— 這是最後一幕。
	if st.Prompt == gamestate.PromptEnding {
		if st.Ending != nil && st.Ending.Asking {
			switch {
			case inpututil.IsKeyJustPressed(ebiten.KeyY):
				st.AnswerEnding(true)
			case inpututil.IsKeyJustPressed(ebiten.KeyN):
				st.AnswerEnding(false)
			}
		} else if len(inpututil.AppendJustPressedKeys(nil)) > 0 {
			st.AdvanceEnding()
		}
		g.dirty = true
		return nil
	}

	// 讀寶典:任意鍵翻頁,ESC 直接闔上。
	if st.Prompt == gamestate.PromptCodex {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			st.EndCodex()
			g.dirty = true
		} else if len(inpututil.AppendJustPressedKeys(nil)) > 0 {
			st.AdvanceCodex()
			g.dirty = true
		}
		return nil
	}

	// 角色數值畫面:左右翻頁,ESC 收起。
	if st.Prompt == gamestate.PromptZtats {
		switch {
		case inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft):
			st.ZtatsPage(-1)
		case inpututil.IsKeyJustPressed(ebiten.KeyArrowRight):
			st.ZtatsPage(1)
		case inpututil.IsKeyJustPressed(ebiten.KeyEscape),
			inpututil.IsKeyJustPressed(ebiten.KeyZ):
			st.EndZtats()
		}
		g.dirty = true
		return nil
	}

	// 通用選單(R / N / U / M):上下移動、Enter 選定、ESC 放棄。
	//
	// 調藥的藥草清單是**複選**(原版 `sub_18468`):四個方向鍵都能移動
	//(其中兩個往上、兩個往下)、空白或 Enter 勾選、**M 才確定**。
	if st.Prompt == gamestate.PromptPick {
		multi := st.PickMulti()
		switch {
		case inpututil.IsKeyJustPressed(ebiten.KeyArrowUp):
			st.PickMove(-1)
		case inpututil.IsKeyJustPressed(ebiten.KeyArrowDown):
			st.PickMove(1)
		case multi && inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft):
			st.PickMove(-1)
		case multi && inpututil.IsKeyJustPressed(ebiten.KeyArrowRight):
			st.PickMove(1)
		// 翻頁鍵一次移 7 項 —— 原版 `sub_1EFC8` 對 0xD5 / 0xD6 就是這個數字。
		case inpututil.IsKeyJustPressed(ebiten.KeyPageUp):
			st.PickPage(-1)
		case inpututil.IsKeyJustPressed(ebiten.KeyPageDown):
			st.PickPage(1)
		// Home / End —— 原版鍵碼 0xD3 / 0xD4(`docs/re/60` 追記)。
		case inpututil.IsKeyJustPressed(ebiten.KeyHome):
			st.PickHome()
		case inpututil.IsKeyJustPressed(ebiten.KeyEnd):
			st.PickEnd()
		case multi && (inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
			inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
			inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter)):
			st.PickToggle()
		case multi && inpututil.IsKeyJustPressed(ebiten.KeyM):
			st.PickConfirm()
		case !multi && (inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
			inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter)):
			st.PickChoose()
		case inpututil.IsKeyJustPressed(ebiten.KeyEscape):
			st.PickCancel()
		}
		g.dirty = true
		return nil
	}

	// 主選單:上下移動、Enter 選定、ESC 收起(等同「回到景色」)。
	if st.Prompt == gamestate.PromptMenu {
		switch {
		case inpututil.IsKeyJustPressed(ebiten.KeyArrowUp):
			st.MenuMove(-1)
		case inpututil.IsKeyJustPressed(ebiten.KeyArrowDown):
			st.MenuMove(1)
		case inpututil.IsKeyJustPressed(ebiten.KeyEnter),
			inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter):
			st.MenuChoose()
		case inpututil.IsKeyJustPressed(ebiten.KeyEscape):
			st.MenuMove(int(gamestate.MenuReturnToView) - int(st.Menu.Cursor))
			st.MenuChoose()
		}
		g.dirty = true
		return nil
	}

	// 建立新角色:開場白 / 結語按任意鍵,七題只收 A / B,名字打字,性別 M / F。
	if st.Prompt == gamestate.PromptCreate {
		switch st.Create.Stage {
		case gamestate.CreationIntro, gamestate.CreationClosing:
			if len(inpututil.AppendJustPressedKeys(nil)) > 0 {
				st.AdvanceCreation()
			}
		case gamestate.CreationQuestion:
			switch {
			case inpututil.IsKeyJustPressed(ebiten.KeyA):
				st.AnswerCreation(true)
			case inpututil.IsKeyJustPressed(ebiten.KeyB):
				st.AnswerCreation(false)
			}
		case gamestate.CreationName:
			for _, r := range ebiten.AppendInputChars(nil) {
				st.TypeCreationName(string(r))
			}
			switch {
			case inpututil.IsKeyJustPressed(ebiten.KeyEnter),
				inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter):
				st.ConfirmCreationName()
			case inpututil.IsKeyJustPressed(ebiten.KeyBackspace):
				st.TypeCreationName("")
			}
		case gamestate.CreationGender:
			switch {
			case inpututil.IsKeyJustPressed(ebiten.KeyM):
				st.AnswerCreationGender(true)
			case inpututil.IsKeyJustPressed(ebiten.KeyF):
				st.AnswerCreationGender(false)
			}
		}
		g.dirty = true
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

	// 施法:收符文首字母(最多四個),**Enter 或空白鍵**送出、ESC 作罷。
	//
	// ⚠ 空白鍵在原版就是送出鍵,所以要讓它走 `TypeSpellLetter` ——
	// 它送出之後就不能再處理後面的按鍵了(不然會把 Enter 當成新一輪的輸入)。
	if st.Prompt == gamestate.PromptSpell {
		for _, r := range ebiten.AppendInputChars(nil) {
			if st.TypeSpellLetter(r) {
				g.dirty = true
				return nil
			}
		}
		switch {
		case inpututil.IsKeyJustPressed(ebiten.KeyEnter),
			inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter):
			st.SubmitSpell()
		case inpututil.IsKeyJustPressed(ebiten.KeyBackspace):
			st.BackspaceSpell()
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
		// 戰鬥中也能用的字母鍵(`jpt_A5C8` 有實作的那些)。
		//
		// ★ 這幾支**與地圖上是同一份程式** —— 原版只有一對座標
		// (`byte_3E0A6/A7`),戰鬥時借給行動中的單位;`sub_DB10` 再依
		// 地點編號決定要讀世界、場景還是戰場緩衝。引擎照這個形狀做了
		// (`TileAt` / `SetTileAt` 加戰鬥分支 + `focusCombatUnit`),
		// 所以這裡直接呼叫同樣的方法就對了。
		case inpututil.IsKeyJustPressed(ebiten.KeyG):
			st.Get()
		case inpututil.IsKeyJustPressed(ebiten.KeyJ):
			st.Jimmy()
		case inpututil.IsKeyJustPressed(ebiten.KeyO):
			st.OpenChest()
		case inpututil.IsKeyJustPressed(ebiten.KeyP):
			st.Push()
		case inpututil.IsKeyJustPressed(ebiten.KeyR):
			st.BeginReady()
		case inpututil.IsKeyJustPressed(ebiten.KeyS):
			st.Search()
		case inpututil.IsKeyJustPressed(ebiten.KeyU):
			st.BeginUse()
		case inpututil.IsKeyJustPressed(ebiten.KeyY):
			st.Yell()
		case inpututil.IsKeyJustPressed(ebiten.KeyZ):
			st.BeginZtats()
		// K 在戰鬥中是**另一支**(`sub_16058`):踩在梯子上就離場,
		// 否則問方向爬過鄰格的 tile 0x4C。
		case inpututil.IsKeyJustPressed(ebiten.KeyK):
			st.CombatKlimb()
		default:
			for key, dir := range dirs {
				if inpututil.IsKeyJustPressed(key) {
					st.CombatMove(dir)
				}
			}
			// 原版戰鬥中不可用的字母鍵**各有自己的回應**,不是統一一句話
			// (`sub_A360` 的 `jpt_A5C8`,見 gamestate/combatcmd.go)。
			for _, r := range ebiten.AppendInputChars(nil) {
				up := r
				if up >= 'a' && up <= 'z' {
					up -= 32
				}
				if msg, ok := gamestate.CombatRefuse(up); ok {
					st.Log(msg)
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
			// 酒館的打聽消息是**打字**,所以 Enter 與 Backspace 也要送進去。
			switch {
			case inpututil.IsKeyJustPressed(ebiten.KeyEnter),
				inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter):
				st.ShopChoose('\r')
			case inpututil.IsKeyJustPressed(ebiten.KeyBackspace):
				st.ShopChoose(8)
			}
		}
		g.dirty = true
		return nil
	}

	// 衛兵的盤查:黑棘宮殿要打密語,其餘地方只收 Y / N。
	if st.Prompt == gamestate.PromptGuard {
		if st.Guard != nil && st.Guard.Password {
			for _, r := range ebiten.AppendInputChars(nil) {
				st.TypeRune(r)
			}
			switch {
			case inpututil.IsKeyJustPressed(ebiten.KeyEnter),
				inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter):
				st.SubmitGuard()
			case inpututil.IsKeyJustPressed(ebiten.KeyBackspace):
				st.Backspace()
			case inpututil.IsKeyJustPressed(ebiten.KeyEscape):
				st.CancelGuard()
			}
		} else {
			switch {
			case inpututil.IsKeyJustPressed(ebiten.KeyY):
				st.AnswerGuard(true)
			case inpututil.IsKeyJustPressed(ebiten.KeyN),
				inpututil.IsKeyJustPressed(ebiten.KeyEscape):
				st.AnswerGuard(false)
			}
		}
		g.dirty = true
		return nil
	}

	// 數字輸入 —— 只收數字鍵;空白與 0 是放棄,其餘按鍵原版會繼續等。
	//
	// 兩位數的地方(調藥的「要幾份?」)要按 Enter 送出,Backspace 退一位。
	if st.AwaitingNumber() {
		for n := 0; n <= 9; n++ {
			if inpututil.IsKeyJustPressed(ebiten.Key0 + ebiten.Key(n)) {
				st.AnswerNumber(n)
			}
		}
		switch {
		case inpututil.IsKeyJustPressed(ebiten.KeyEnter),
			inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter):
			st.SubmitNumber()
		case inpututil.IsKeyJustPressed(ebiten.KeyBackspace):
			st.BackspaceNumber()
		case inpututil.IsKeyJustPressed(ebiten.KeySpace),
			inpututil.IsKeyJustPressed(ebiten.KeyEscape):
			st.AnswerNumber(0)
		}
		g.dirty = true
		return nil
	}

	// 通用「打一行字」提問(轉入 U4 的改名)。Enter 送出、ESC 等同不改。
	if st.AwaitingText() {
		for _, r := range ebiten.AppendInputChars(nil) {
			st.TypeText(r)
		}
		switch {
		case inpututil.IsKeyJustPressed(ebiten.KeyEnter),
			inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter):
			st.SubmitText()
		case inpututil.IsKeyJustPressed(ebiten.KeyBackspace):
			st.BackspaceText()
		case inpututil.IsKeyJustPressed(ebiten.KeyEscape):
			// 原版的「什麼都不打」就是保留原值 —— ESC 走同一條路。
			st.Input = ""
			st.SubmitText()
		}
		g.dirty = true
		return nil
	}

	// 通用 Y / N 提問(紮營的守夜之類)。ESC 等同 N,與原版一致。
	if st.AwaitingYesNo() {
		switch {
		case inpututil.IsKeyJustPressed(ebiten.KeyY):
			st.AnswerYesNo(true)
		case inpututil.IsKeyJustPressed(ebiten.KeyN), inpututil.IsKeyJustPressed(ebiten.KeyEscape):
			st.AnswerYesNo(false)
		}
		g.dirty = true
		return nil
	}

	// 「汝束手就擒否?」—— 只收 Y / N。**ESC 不是取消**:原版沒有第三條路。
	if st.Prompt == gamestate.PromptArrest {
		switch {
		case inpututil.IsKeyJustPressed(ebiten.KeyY):
			st.AnswerArrest(true)
		case inpututil.IsKeyJustPressed(ebiten.KeyN):
			st.AnswerArrest(false)
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
		// 地牢:↑ 前進、↓ 後退、← → 轉向;字母鍵走**與地面同一張指令表**。
		//
		// ★ 原版的字母指令只有一個分派器(`sub_2ACF4`),地牢的迴圈
		// (`sub_4B14`)在方向鍵與數字鍵之後就把鍵交給它 —— 位置差異是在
		// 各指令**內部**判的(`byte_3E0A3` 是 0 / <0x21 / >=0x21)。
		// 舊版在這裡自己列了一份短清單,結果 S 搜尋、Z 數值、U 用道具、
		// R 換裝、A 攻擊在地牢裡全部按不到 —— **做完了卻用不到,等於沒做**。
		//
		// ⚠ 畫面還是俯視的(透視圖要先解 DNG1-3.16),但操作與規則是照原版的。
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
		// 站在豎琴前(正南那一格)時數字鍵是彈音,不是指令。
		case st.AtHarp() && harpKey(st) != 0:
			st.PlayNote(harpKey(st))
		default:
			g.commandKeys(st, ctrl)
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
		// ★ 喝醉了先攔一手:原版 `sub_1158` 是**讀鍵那一層**在做這件事 ——
		// 有一半機率把按到的鍵換成一個隨機方向,而且印 "Hic!"。
		// 攔在這裡而不是各指令裡面,才會像原版一樣**所有鍵都會踉蹺**。
		if anyGameKeyPressed() {
			if d, staggered := st.DrunkStagger(); staggered {
				st.Move(d)
				if len(st.Messages) != before || g.key() != snapshot {
					g.dirty = true
				}
				return nil
			}
		}
		g.commandKeys(st, ctrl)
		if inpututil.IsKeyJustPressed(ebiten.KeyK) {
			st.Klimb()
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

// commandKeys 是**字母指令表**,地面、場景與地牢共用一份。
//
// 這對應原版的 `sub_2ACF4` —— A..Z 加空白鍵共 27 個 case 的單一分派器。
// 各指令自己判位置(`byte_3E0A3`:0 = 地表 / <0x21 = 場景 / >=0x21 = 地牢),
// 所以**不該在輸入層按位置分兩份清單**;分了就會漏,而漏掉的指令
// 在測試裡完全看不出來(規則都實作了,只是按不到)。
//
// ⚠ K(Klimb)不在這裡:地牢的 K 要先問上 / 下,兩邊的互動流程不同。
// 方向鍵也不在這裡(地牢是前進 / 轉向,地面是四方走)。
//
// ★ D 與 W **不是指令** —— 原版只印 `D-What?` / `W-What?`(`docs/re/49`)。
// 這裡照樣印,而不是靜靜吃掉:玩家按錯鍵會看到跟原版一樣的回應。
func (g *game) commandKeys(st *gamestate.State, ctrl bool) {
	for _, c := range commandTable {
		if !inpututil.IsKeyJustPressed(c.key) {
			continue
		}
		if c.key == ebiten.KeyQ && ctrl {
			continue // Ctrl+Q 是離開,不是存檔
		}
		// ★ 原版每個指令**先印自己的名字**,再跑處理函式(`sub_2ACF4`)。
		// 名字末尾的「——」就是「等方向」的提示,少了它按下 G 之後畫面毫無反應。
		st.EchoCommand(c.echo)
		if c.run != nil {
			c.run(st)
		}
		return
	}
}

// commandEntry 是一個字母指令:按鍵、回顯用的字母、處理函式。
//
// `run` 為 nil 代表原版只印一句話就結束(D 與 W 那兩個空鍵)。
type commandEntry struct {
	key  ebiten.Key
	echo rune
	run  func(*gamestate.State)
}

// commandTable 是**字母指令表**,地面、場景與地牢共用一份。
//
// 這對應原版的 `sub_2ACF4` —— A..Z 加空白鍵共 27 個 case 的單一分派器。
// 各指令自己判位置(`byte_3E0A3`:0 = 地表 / <0x21 = 場景 / >=0x21 = 地牢),
// 所以**不該在輸入層按位置分兩份清單**;分了就會漏,而漏掉的指令
// 在測試裡完全看不出來(規則都實作了,只是按不到)。
//
// ⚠ K(Klimb)不在這裡:地牢的 K 要先問上 / 下,兩邊的互動流程不同。
// 方向鍵也不在這裡(地牢是前進 / 轉向,地面是四方走)。
//
// ★ D 與 W **不是指令** —— 原版只印 `D-What?` / `W-What?`(`docs/re/49`),
// 所以它們的 `run` 是 nil,只靠回顯。
var commandTable = []commandEntry{
	{ebiten.KeyE, 'E', func(st *gamestate.State) { st.Enter() }},
	{ebiten.KeyT, 'T', func(st *gamestate.State) { st.Talk() }},
	// L 是原版的 Look:先問方向,再看那一格(LOOK2.DAT / SIGNS.DAT)。
	{ebiten.KeyL, 'L', func(st *gamestate.State) { st.Look() }},
	// P 是原版的 Push:推家具,推不動就改拉。
	{ebiten.KeyP, 'P', func(st *gamestate.State) { st.Push() }},
	// J 撬鎖、V 看寶石(攤開全景)、Z 角色數值 —— 都照原版鍵位。
	{ebiten.KeyJ, 'J', func(st *gamestate.State) { st.Jimmy() }},
	{ebiten.KeyV, 'V', func(st *gamestate.State) { st.ViewGem() }},
	{ebiten.KeyZ, 'Z', func(st *gamestate.State) { st.BeginZtats() }},
	// S 搜尋:查陷阱、找密門、翻家具。地牢裡不問方向,搜腳下。
	{ebiten.KeyS, 'S', func(st *gamestate.State) { st.Search() }},
	// F 開砲(船上打舷側 / 陸上要緊鄰大砲)。
	{ebiten.KeyF, 'F', func(st *gamestate.State) { st.Fire() }},
	// R 換裝備(六個欄位一支)、N 換位、U 用道具、M 調藥 —— 都走通用選單。
	{ebiten.KeyR, 'R', func(st *gamestate.State) { st.BeginReady() }},
	{ebiten.KeyN, 'N', func(st *gamestate.State) { st.BeginNewOrder() }},
	{ebiten.KeyU, 'U', func(st *gamestate.State) { st.BeginUse() }},
	{ebiten.KeyM, 'M', func(st *gamestate.State) { st.BeginMix() }},
	{ebiten.KeyO, 'O', func(st *gamestate.State) { st.OpenChest() }},
	{ebiten.KeyG, 'G', func(st *gamestate.State) { st.Get() }},
	{ebiten.KeyB, 'B', func(st *gamestate.State) { st.Board() }},
	{ebiten.KeyX, 'X', func(st *gamestate.State) { st.Exit() }},
	// C 施法(收符文首字母),I 點火把 —— 兩個都照原版鍵位。
	{ebiten.KeyC, 'C', func(st *gamestate.State) { st.BeginCastPrompt() }},
	{ebiten.KeyI, 'I', func(st *gamestate.State) { st.LightTorch() }},
	// 空白鍵是原版的 Pass —— 而在海上揚著帆時它是**收帆**。
	// 那兩句由 `Pass` 自己印,所以回顯給 0(不印)。
	{ebiten.KeySpace, 0, func(st *gamestate.State) { st.Pass() }},
	// H 是原版的 Hole up:在船上修船、在城裡要站床上、其餘紮營。
	{ebiten.KeyH, 'H', func(st *gamestate.State) { st.HoleUp() }},
	// Y 是原版的 Yell:船上收放帆,城裡喊暗影君主的名字,野外說力量之言。
	{ebiten.KeyY, 'Y', func(st *gamestate.State) { st.Yell() }},
	// D 與 W 是原版留著的空鍵 —— 只有回顯,沒有處理函式。
	{ebiten.KeyD, 'D', nil},
	{ebiten.KeyW, 'W', nil},
	// Q 存檔(不離開)。原版 Q 是「存檔並離開」,這裡拆開:
	// 離開走 F10 / Ctrl+Q 並自動存檔,見 esc-cancel-f10-quit-autosave。
	{ebiten.KeyQ, 'Q', func(st *gamestate.State) { st.Save() }},
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
	audioDir := flag.String("audio", "assets/audio",
		"渲染好的配樂目錄(ogg;由 .EUP 離線轉出,不入 git)")
	fontPrefix := flag.String("font", "assets/fonts/eten-15",
		"倚天中文點陣字 atlas 前綴(用 tools/dev.sh font 15 產生)")
	saveFile := flag.String("save", "",
		"要載入的存檔;留空則依序試 gamedata/SAVED.GAM、INIT.GAM")
	scale := flag.Int("scale", 2, "視窗放大倍率(整數;邏輯畫布固定 640×400)")
	display := flag.String("display", "EGA",
		"顯示模式:EGA(.16 十六色)/ CGA(.4 四色)/ Tandy(同 EGA 素材)/ Hercules(單色)")
	showVersion := flag.Bool("version", false, "印出版本後結束")
	playIntro := flag.Bool("intro", false, "強制播開場動畫(沒有存檔時本來就會播)")
	newChar := flag.Bool("create", false,
		"直接走建角流程(吉普賽的七題八德),覆寫載入的那名聖者")
	showMenu := flag.Bool("menu", false, "開機先進主選單(原版的六個項目)")
	// 原版的「Transfer from Ultima IV」寫死讀 `a:party.sav`。現代環境沒有
	// A 磁碟,所以路徑用旗標給;不給就在選單裡照實說,不假裝轉入。
	u4save := flag.String("u4save", "",
		"《創世紀 IV》的 PARTY.SAV 路徑(主選單「從創世紀 IV 轉入」要用)")
	flag.Parse()

	if *showVersion {
		fmt.Printf("u5cht %s\n", version)
		return
	}

	// 素材缺件不致命:優雅降級並明說缺什麼(CLAUDE.md §3.0)。
	// headless 截圖請用 `u5dump scene`(純 CPU,不需要 GPU)。
	mode, err := u5data.ParseDisplayMode(*display)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(2)
	}
	if !mode.Implemented() {
		fmt.Fprintf(os.Stderr, "⚠ %s 模式尚未實作,改用 EGA\n", mode.Name())
		mode = u5data.DisplayEGA
	}
	// Hercules 用的是原版自己的圖樣表,但 blit 迴圈還沒追完 —— 照實說一句,
	// 不要讓玩家以為那是逐像素重現(CLAUDE.md §3.0)。
	if mode == u5data.DisplayHercules {
		fmt.Fprintln(os.Stderr, "⚠ Hercules:圖樣表照原版,但 blit 迴圈未追完(見 docs/re/64)")
	}
	bundle, warns := assets.Load(assets.Options{
		GameData:   *gamedata,
		FMTowns:    *fmtowns,
		FontPrefix: *fontPrefix,
		SaveFile:   *saveFile,
		Display:    mode,
	})
	for _, w := range warns {
		fmt.Fprintf(os.Stderr, "⚠ %s\n", w)
	}

	// ⚠ 缺**世界地圖**不是「降級」,是根本沒得玩 —— 那代表玩家還沒把原版資料
	// 放進來。這時候要一句話講清楚就結束,不要開一個空白視窗、
	// 也不要讓上面那三十行警告淹掉重點。
	//
	// 交付包的 `PLAY.bat` 在前面也擋一次;這一條是給直接跑執行檔的人。
	if bundle.World == nil {
		fmt.Fprintf(os.Stderr, `
找不到原版遊戲資料(在 %q 底下讀不到 BRIT.DAT)。

本程式只是引擎,不含原版的地圖、美術與文字 —— 請自備一份合法的
DOS 版《Ultima V》,把資料檔複製到那個目錄裡,或用 -gamedata 指定位置:

    u5cht -gamedata /path/to/ultima5

詳見同目錄的 README-CHT.txt。
`, *gamedata)
		os.Exit(1)
	}

	st := &gamestate.State{
		World:        bundle.World,
		Under:        bundle.Under,
		Scenes:       bundle.Scenes,
		NPCs:         bundle.NPCs,
		Talks:        bundle.Talks,
		Shops:        bundle.Shops,
		Items:        bundle.Items,
		SpecialItems: bundle.SpecialItems,
		Objects:      bundle.Objects,
		CombatMaps:   bundle.Combat,
		Stats:        bundle.Stats,
		Spells:       bundle.Spells,
		Runes:        bundle.Runes,
		Lore:         bundle.Lore,
		Story:        bundle.Story,
		Question:     bundle.Question,
		Misc:         bundle.Misc,
		EndMsg:       bundle.EndMsg,
		MiscMaps:     bundle.MiscMaps,
		Dungeons:     bundle.Dungeons,
		Moons:        bundle.Moons,
		Look2:        bundle.Look2,
		Signs:        bundle.Signs,
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
	// 「從創世紀 IV 轉入」要用的存檔路徑(原版寫死 `a:party.sav`)。
	st.U4SavePath = *u4save
	// 建角。原版是主選單的「Create New Character」;引擎還沒有主選單,
	// 先用旗標接上 —— 沒有它玩家只能扮演存檔裡做好的聖者。
	// ⚠ 開場動畫還在播的時候不能同時建角,兩者都吃「任意鍵」。
	if *newChar && st.Prompt == gamestate.PromptNone {
		st.BeginCreation()
	} else if *showMenu && st.Prompt == gamestate.PromptNone {
		st.BeginMainMenu()
	}

	// 配樂。⚠ 目前**沒有後端**(傳 nil)—— `.EUP` → ogg 的離線渲染還沒做,
	// 所以接上去也還沒有聲音。先把管線與「缺什麼」的回報接好:曲號在引擎裡
	// 已經會正確切換(`docs/re/87`),缺的只有音訊本身。
	// ⚠ 讀不到 `U5_BGM.TBL` 只是「沒有 FM Towns 資料」,不擋遊戲。
	var music *audio.Player
	if mp, err := audio.New(*fmtowns, *audioDir, nil); err != nil {
		fmt.Fprintf(os.Stderr, "⚠ 配樂:%v\n", err)
	} else {
		music = mp
		if n := len(mp.Missing()); n > 0 {
			fmt.Fprintf(os.Stderr,
				"⚠ 配樂:%d/%d 首還沒渲染成 %s(離線轉檔還沒做,遊戲照樣可玩)\n",
				n, u5data.BGMSongCount, audio.Ext)
		}
	}

	g := &game{
		state: st,
		music: music,
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

// harpKey 回傳這一幀按下的數字鍵('0'..'9'),沒有就回 0。
func harpKey(st *gamestate.State) rune {
	_ = st
	for k := ebiten.Key0; k <= ebiten.Key9; k++ {
		if inpututil.IsKeyJustPressed(k) {
			return rune('0' + (k - ebiten.Key0))
		}
	}
	return 0
}

// anyGameKeyPressed 回報這一幀有沒有按下「會被醉酒攔掉」的鍵。
//
// 原版的 `sub_1158` 攔的是**任何一次讀鍵**,所以這裡把字母指令表、K 與四個
// 方向鍵都算進去 —— 但**不含** Ctrl 組合與 ESC 那類離開 / 系統鍵。
func anyGameKeyPressed() bool {
	for _, c := range commandTable {
		if inpututil.IsKeyJustPressed(c.key) {
			return true
		}
	}
	for _, k := range []ebiten.Key{
		ebiten.KeyK,
		ebiten.KeyArrowUp, ebiten.KeyArrowDown,
		ebiten.KeyArrowLeft, ebiten.KeyArrowRight,
	} {
		if inpututil.IsKeyJustPressed(k) {
			return true
		}
	}
	return false
}
