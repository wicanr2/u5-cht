// Package appui 放**應用層**的互動狀態,不是遊戲機制。
//
// 這裡只有一件事:離開遊戲的語意。原版的 U5 只能用 `Q)uit & Save` 離開,
// 而且那會直接結束程式;現代玩家的肌肉記憶不是這樣 —— 按 ESC 是「我按錯了,
// 回上一層」,不是「把一小時的進度丟掉」。
//
// ⚠ **為什麼不放進 `internal/game`**:`internal/game` 是原版行為的還原,
// 混進「確認框開著沒有」這種只有本重製版才有的狀態,會讓「這條規則是不是原版的」
// 這個問題變得難回答(`CLAUDE.md §3.0`:機制一律還原原版,沒證據就不要實作)。
// 存檔與離開是平台層的事,放這裡。
//
// 規則來源:`~/.claude/knowledge-base/retro-cht/esc-cancel-f10-quit-autosave/SKILL.md`
//(姊妹專案 u1-cht 已照這條做:`F5` 手動存檔、`F10` / `Ctrl+Q` 跳確認框、
// `ESC` 永遠是取消)。
package appui

// Keys 是這一帧按下了哪些「應用層」的鍵。
//
// 用布林結構而不是直接吃 ebiten 的鍵碼,是為了讓下面那台狀態機
// **在沒有視窗、沒有 GPU 的測試裡也跑得完**（`CLAUDE.md §3.1`:
// 驗證畫面與行為都不該需要 GPU）。鍵碼對應留在 `cmd/u5cht`。
type Keys struct {
	// Quit 是明確的「我要離開」訊號(F10 或 Ctrl+Q)。
	Quit bool
	// Yes / No 是確認框的兩顆單鍵(Y / Enter、N)。
	Yes bool
	No  bool
	// Escape 是 ESC。**它永遠只有取消的意思。**
	Escape bool
	// Save / Load 是即時存檔與讀回(F5 / F6)。
	Save bool
	Load bool
	// Help 是說明畫面的開關鍵(F1),見 `help.go`。
	Help bool
	// PageUp / PageDown 給說明畫面翻頁用。
	PageUp   bool
	PageDown bool
}

// Action 是狀態機這一帧要求呼叫端做什麼。
type Action int

// 這些是「要做什麼」而不是「按了什麼」—— 呼叫端不必再判一次鍵。
const (
	// ActNone 什麼都不做。
	ActNone Action = iota
	// ActOpenedQuit 剛剛把確認框打開(呼叫端只要重畫)。
	ActOpenedQuit
	// ActCancelled 確認框被取消了。
	ActCancelled
	// ActSaveAndQuit 玩家確認要離開 ⇒ **先存檔再結束**,順序不能反。
	ActSaveAndQuit
	// ActQuickSave 即時存檔(不離開遊戲)。
	ActQuickSave
	// ActQuickLoad 讀回上次的存檔。
	ActQuickLoad
)

// QuitDialog 是離開確認框的狀態。零值 = 沒開。
type QuitDialog struct {
	open bool
}

// IsOpen 回報確認框開著沒有(繪圖層用它決定要不要疊那個框)。
func (d *QuitDialog) IsOpen() bool { return d != nil && d.open }

// Step 吃這一帧的按鍵,回報要做什麼。
//
// 三條鐵則就寫在這裡,不散在呼叫端:
//
//  1. **ESC 永遠不會結束遊戲。** 框開著 → 取消;框沒開 → 什麼都不做。
//  2. **F10 只把框打開**,不會直接離開。
//  3. **框開著的時候它是 modal** —— 存檔 / 讀檔鍵一律不作用,
//     否則玩家在確認框上按 F5 會得到一個「存了但沒離開、框還開著」的怪狀態。
func (d *QuitDialog) Step(k Keys) Action {
	if d == nil {
		return ActNone
	}
	if d.open {
		switch {
		case k.Yes:
			// ⚠ 只回報「要存檔並離開」。**真的存檔由呼叫端做**,
			// 而且存失敗時要留在遊戲裡 —— 玩家選了 Yes 是相信你不會吃他的存檔,
			// 靜默結束比報錯糟得多。
			return ActSaveAndQuit
		case k.No, k.Escape:
			d.open = false
			return ActCancelled
		}
		return ActNone
	}
	switch {
	case k.Quit:
		d.open = true
		return ActOpenedQuit
	case k.Save:
		return ActQuickSave
	case k.Load:
		return ActQuickLoad
	}
	return ActNone
}

// Close 把框關掉(存檔失敗要留在遊戲裡時用)。
func (d *QuitDialog) Close() {
	if d != nil {
		d.open = false
	}
}
