package game

// 主選單(原版 `sub_60BC` 畫、`sub_6730` 驅動)
//
// 六個項目,原版逐字:
//
//	Journey Onward          載入進度繼續玩
//	Create New Character    走建角流程
//	Transfer from Ultima IV 從 U4 的存檔轉入角色
//	Ultima V Introduction   重播開場動畫
//	Acknowledgements        製作名單
//	Return to the View      回到那張「窗外景色」的待機畫面
//
// ⚠ 順序照原版,不重排。玩家對老遊戲的肌肉記憶就是「第幾項」——
// 把常用的往上搬會踩到那個記憶,而那不是體貼是干擾。

// MenuItem 是主選單的一個項目。
type MenuItem int

const (
	// MenuJourneyOnward 是「繼續前行」:用載入的存檔開始玩。
	MenuJourneyOnward MenuItem = iota
	// MenuCreateCharacter 是「建立新角色」。
	MenuCreateCharacter
	// MenuTransferU4 是「從創世紀 IV 轉入」。**尚未實作** —— 轉換規則未逆。
	MenuTransferU4
	// MenuIntroduction 是「重播開場」。
	MenuIntroduction
	// MenuAcknowledgements 是「製作名單」。
	MenuAcknowledgements
	// MenuReturnToView 是「回到景色」:收起選單。
	MenuReturnToView
	// MenuItemCount 是項目數。
	MenuItemCount
)

// MenuLabels 是六個項目的中文。英文原文寫在型別的說明裡。
var MenuLabels = [MenuItemCount]string{
	MenuJourneyOnward:    "繼續前行",
	MenuCreateCharacter:  "建立新角色",
	MenuTransferU4:       "從創世紀 IV 轉入",
	MenuIntroduction:     "創世紀 V 開場",
	MenuAcknowledgements: "製作群",
	MenuReturnToView:     "回到景色",
}

// MainMenu 是進行中的主選單。
type MainMenu struct {
	// Cursor 是游標停在第幾項。
	Cursor MenuItem
}

// BeginMainMenu 打開主選單。
func (s *State) BeginMainMenu() {
	s.Menu = &MainMenu{}
	s.Prompt = PromptMenu
}

// MenuMove 上下移動游標。delta 是 -1 或 +1,會繞回。
func (s *State) MenuMove(delta int) {
	if s.Menu == nil {
		return
	}
	n := int(MenuItemCount)
	s.Menu.Cursor = MenuItem(((int(s.Menu.Cursor)+delta)%n + n) % n)
}

// MenuChoose 按下 Enter。回報選單有沒有關掉。
func (s *State) MenuChoose() bool {
	if s.Menu == nil {
		return false
	}
	switch s.Menu.Cursor {
	case MenuJourneyOnward:
		s.closeMenu()
		return true
	case MenuCreateCharacter:
		s.closeMenu()
		s.BeginCreation()
		return true
	case MenuTransferU4:
		// ⚠ 沒實作就說沒實作。做一個「假裝轉入」的分支比缺這一項更糟 ——
		// 玩家會以為角色轉進來了(CLAUDE.md §3.0)。
		s.Log(MsgTransferNotImplemented)
		return false
	case MenuIntroduction:
		s.closeMenu()
		s.BeginIntro()
		return true
	case MenuAcknowledgements:
		// 原版這一項畫的是一張製作群的畫面,不是文字表 —— 素材還沒對到,
		// 所以照實說。**不要**拿結局那份頒獎狀名單頂替:那是通關的獎勵,
		// 內容與這裡完全不同,混用等於偽造。
		s.Log(MsgAcknowledgementsNotImplemented)
		return false
	case MenuReturnToView:
		s.closeMenu()
		return true
	}
	return false
}

func (s *State) closeMenu() {
	s.Menu = nil
	s.Prompt = PromptNone
}

// InMainMenu 回報選單開著沒有。
func (s *State) InMainMenu() bool { return s.Prompt == PromptMenu && s.Menu != nil }
