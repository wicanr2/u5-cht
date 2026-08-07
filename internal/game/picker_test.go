package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// TestReadyMenuChainsTwoPicks:R 是「先挑人、再挑武器」兩層選單。
//
// 第一層選完之後選單**不能關掉** —— 它要接著開第二層。`beginPick` 的 then
// 回傳 false 就是「還有下一步」,寫成一律 true 的話玩家挑完人就沒反應了。
func TestReadyMenuChainsTwoPicks(t *testing.T) {
	s := pickScene(t)
	s.Inventory.Items[24] = 1 // Mace,讓 ReadyList 有東西

	if !s.BeginReady() {
		t.Fatalf("開不了選單:%q", s.Messages)
	}
	if s.Prompt != PromptPick {
		t.Fatalf("Prompt 是 %v", s.Prompt)
	}
	if !strings.Contains(s.PickLines()[0], MsgPickWho) {
		t.Errorf("第一層該問「由誰」:%q", s.PickLines()[0])
	}
	s.PickChoose() // 選第一個人
	if s.Prompt != PromptPick {
		t.Fatalf("挑完人之後選單關掉了 —— 第二層沒開,Prompt=%v", s.Prompt)
	}
	if !strings.Contains(s.PickLines()[0], MsgPickItem) {
		t.Errorf("第二層該問「哪一件」:%q", s.PickLines()[0])
	}
	// ⚠ 不能直接 PickChoose ——存檔的背包本來就有武器,游標停在第一項
	// 不見得是我們放進去的那把。要驗的是「選到的那一件真的裝上了」,
	// 所以得先把游標移到它上面。
	found := false
	for i, e := range s.Pick.Entries {
		if e.Value == 24 {
			s.Pick.Cursor, found = i, true
		}
	}
	if !found {
		t.Fatalf("選單裡沒有剛放進背包的武器:%q", s.PickLines())
	}
	s.PickChoose() // 選那件武器
	if s.Prompt != PromptNone {
		t.Errorf("選完之後該收起來,Prompt=%v", s.Prompt)
	}
	if s.Roster[0].Raw[u5data.CharWeapon] != 24 &&
		s.Roster[0].Raw[u5data.CharShield] != 24 {
		t.Errorf("武器沒裝上:右手 %d 左手 %d",
			s.Roster[0].Raw[u5data.CharWeapon], s.Roster[0].Raw[u5data.CharShield])
	}
}

// TestNewOrderMenuStillOffersTheAvatar:換位的第一層要列得出聖者。
//
// ⚠ 原版是讓玩家**選得到**、然後才說「must lead!」。從清單裡藏起來的話
// 玩家不會知道那是規則,只會以為選單壞了 —— 而「看不見」與「選了被拒」
// 對玩家是兩種完全不同的訊息。
func TestNewOrderMenuStillOffersTheAvatar(t *testing.T) {
	s := pickScene(t)
	if s.PartySize < 2 {
		t.Skip("隊伍不到兩人")
	}
	if !s.BeginNewOrder() {
		t.Fatalf("開不了選單:%q", s.Messages)
	}
	lines := s.PickLines()
	if !strings.Contains(strings.Join(lines, "|"), s.Roster[0].Name) {
		t.Errorf("清單裡沒有聖者:%q", lines)
	}
	// 選了聖者要被拒,而且說明原因。
	s.Pick.Cursor = 0
	s.Messages = nil
	s.PickChoose()
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgMustLead) {
		t.Errorf("選了聖者沒說「必須走在最前面」:%q", s.Messages)
	}
}

// TestPickCancelGoesBack:ESC 放棄不留下半完成的狀態。
func TestPickCancelGoesBack(t *testing.T) {
	s := pickScene(t)
	s.Inventory.Items[24] = 1
	s.BeginReady()
	s.PickChoose() // 進第二層
	s.PickCancel()
	if s.Prompt != PromptNone || s.Pick != nil {
		t.Errorf("放棄之後沒收乾淨:Prompt=%v Pick=%v", s.Prompt, s.Pick)
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgNevermind) {
		t.Errorf("沒印出「作罷」:%q", s.Messages)
	}
}

// TestPickMoveWraps:游標上下繞回。
func TestPickMoveWraps(t *testing.T) {
	s := pickScene(t)
	s.beginPick("t", []PickEntry{{Label: "a"}, {Label: "b"}, {Label: "c"}}, "", nil)
	s.PickMove(-1)
	if s.Pick.Cursor != 2 {
		t.Errorf("往上該繞到最後,實際 %d", s.Pick.Cursor)
	}
	s.PickMove(1)
	if s.Pick.Cursor != 0 {
		t.Errorf("再往下該回到 0,實際 %d", s.Pick.Cursor)
	}
}

// TestEmptyListSaysSoInsteadOfOpening:候選是空的就不要開一個空選單。
func TestEmptyListSaysSoInsteadOfOpening(t *testing.T) {
	s := pickScene(t)
	for i := range s.Inventory.Items {
		s.Inventory.Items[i] = 0
	}
	s.Messages = nil
	if s.BeginReady() {
		// 第一層是「挑人」,一定有人,所以會開;挑完人之後第二層才是空的。
		s.PickChoose()
	}
	if s.Prompt == PromptPick {
		t.Errorf("背包空的卻開了裝備選單:%q", s.PickLines())
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgNoUsableItems) {
		t.Errorf("沒說沒有可用的道具:%q", s.Messages)
	}
}

// TestMixDoesNotOpenASpellMenu:調藥的第一步是**符文輸入**,不是咒語選單。
//
// ⚠ 這條原本測的是「選單只列調得出來的咒語」—— 而那張選單是我自己加的,
// 原版沒有。測試每次都綠,因為它量的是我自己的發明(`docs/re/58`)。
// 現在改成釘住原版的第一步:`sub_18704` 印「For what spell?」然後叫 `sub_1CA0C`。
func TestMixDoesNotOpenASpellMenu(t *testing.T) {
	s := pickScene(t)
	if s.Spells == nil || s.Runes == nil {
		t.Skip("沒有咒語表 / 符文表")
	}
	for r := range s.Inventory.Reagents {
		s.Inventory.Reagents[r] = 5
	}
	if !s.BeginMix() {
		t.Fatalf("按 M 沒有開始調藥:%q", s.Messages)
	}
	if s.Prompt != PromptSpell {
		t.Errorf("第一步是 %v,預期符文輸入", s.Prompt)
	}
	if s.Pick != nil {
		t.Error("開出了咒語選單 —— 原版沒有這一步")
	}
}

func pickScene(t *testing.T) *State {
	t.Helper()
	s := newCreateState(t)
	s.MaxMessages = 16
	s.Messages = nil
	return s
}
