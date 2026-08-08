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

// TestPickMoveDoesNotWrap:原版**不繞回** —— 到頭到尾那一步是空轉。
//
// ⚠ 這條原本叫 `TestPickMoveWraps`,測的是「往上繞到最後一項」——
// 而繞回是我自己加的,原版 `sub_1EFC8` 的移動靠 `sub_1E3D8` / `sub_1E418`
// 回 −1 停住。測試每次都綠,因為它量的是我自己的發明(`docs/re/60` 追記)。
func TestPickMoveDoesNotWrap(t *testing.T) {
	s := pickScene(t)
	s.beginPick("t", []PickEntry{{Label: "a"}, {Label: "b"}, {Label: "c"}}, "", nil)
	s.PickMove(-1)
	if s.Pick.Cursor != 0 {
		t.Errorf("在第一項往上該停住,實際跑到 %d", s.Pick.Cursor)
	}
	for i := 0; i < 10; i++ {
		s.PickMove(1)
	}
	if s.Pick.Cursor != 2 {
		t.Errorf("按到底該停在最後一項,實際 %d", s.Pick.Cursor)
	}
	s.PickMove(1)
	if s.Pick.Cursor != 2 {
		t.Errorf("在最後一項往下該停住,實際跑到 %d", s.Pick.Cursor)
	}
}

// TestCursorSticksToTheMiddleRow 釘住原版最有個性的那一段:
// 游標走到第 4 列就黏住,改成捲視窗。
func TestCursorSticksToTheMiddleRow(t *testing.T) {
	s := &State{MaxMessages: 8}
	var out []PickEntry
	for i := 0; i < 20; i++ {
		out = append(out, PickEntry{Label: "x", Value: i})
	}
	s.beginPick("t", out, "empty", func(int) bool { return true })
	// 前三步游標自己往下走,視窗不動。
	for i := 1; i <= 3; i++ {
		s.PickMove(1)
		if s.Pick.Top != 0 {
			t.Fatalf("第 %d 步視窗就動了(Top=%d)", i, s.Pick.Top)
		}
		if got := s.Pick.row(); got != i+1 {
			t.Fatalf("第 %d 步游標在第 %d 列,預期第 %d 列", i, got, i+1)
		}
	}
	// 到第 4 列之後改成捲視窗,游標列不再變。
	for i := 4; i <= 10; i++ {
		s.PickMove(1)
		if got := s.Pick.row(); got != PickCenterRow {
			t.Fatalf("第 %d 步游標跑到第 %d 列,預期黏在第 %d 列", i, got, PickCenterRow)
		}
		if s.Pick.Top != i-3 {
			t.Fatalf("第 %d 步視窗頂端 %d,預期 %d", i, s.Pick.Top, i-3)
		}
	}
	// 捲到底之後(Top = 20−7 = 13)游標才自己走到第 7 列。
	for i := 0; i < 20; i++ {
		s.PickMove(1)
	}
	if s.Pick.Top != len(out)-PickRows {
		t.Errorf("按到底視窗頂端 %d,預期 %d", s.Pick.Top, len(out)-PickRows)
	}
	if got := s.Pick.row(); got != PickRows {
		t.Errorf("按到底游標在第 %d 列,預期第 %d 列", got, PickRows)
	}
	if s.Pick.Cursor != len(out)-1 {
		t.Errorf("按到底停在第 %d 項,預期第 %d 項", s.Pick.Cursor, len(out)-1)
	}
}

// TestPickOnlyDrawsSevenRows:原版視窗只有七列。
func TestPickOnlyDrawsSevenRows(t *testing.T) {
	s := &State{MaxMessages: 8}
	var out []PickEntry
	for i := 0; i < 20; i++ {
		out = append(out, PickEntry{Label: "x", Value: i})
	}
	s.beginPick("t", out, "empty", func(int) bool { return true })
	lines := s.PickLines()
	if len(lines) != PickRows+1 { // 標題 + 七列
		t.Errorf("畫了 %d 行(含標題),預期 %d 行", len(lines), PickRows+1)
	}
}

// TestScrollHintMatchesTheThreeArrows 對上原版 `sub_29008` 的三個字元碼。
func TestScrollHintMatchesTheThreeArrows(t *testing.T) {
	s := &State{MaxMessages: 8}
	var out []PickEntry
	for i := 0; i < 20; i++ {
		out = append(out, PickEntry{Label: "x", Value: i})
	}
	s.beginPick("t", out, "empty", func(int) bool { return true })
	if got := s.PickScrollHint(); got != "\u2193" {
		t.Errorf("在最上面該只有 ↓,實際 %q", got)
	}
	s.PickMove(5) // 捲一點,兩邊都有東西
	if got := s.PickScrollHint(); got != "\u2195" {
		t.Errorf("中間該是 ↕,實際 %q", got)
	}
	s.PickEnd()
	if got := s.PickScrollHint(); got != "\u2191" {
		t.Errorf("在最下面該只有 ↑,實際 %q", got)
	}
	// 清單短到裝得進視窗 → 不畫箭頭。
	s2 := pickScene(t)
	s2.beginPick("t", []PickEntry{{Label: "a"}, {Label: "b"}}, "", nil)
	if got := s2.PickScrollHint(); got != "" {
		t.Errorf("清單裝得下卻畫了 %q", got)
	}
}

// TestHomeAndEndAreTheTwoUntracedKeyCodes 釘住 0xD3 / 0xD4 的語意。
func TestHomeAndEndAreTheTwoUntracedKeyCodes(t *testing.T) {
	s := &State{MaxMessages: 8}
	var out []PickEntry
	for i := 0; i < 20; i++ {
		out = append(out, PickEntry{Label: "x", Value: i})
	}
	s.beginPick("t", out, "empty", func(int) bool { return true })
	s.PickEnd() // 0xD4
	if s.Pick.Cursor != len(out)-1 {
		t.Errorf("End 停在第 %d 項,預期最後一項 %d", s.Pick.Cursor, len(out)-1)
	}
	if s.Pick.Top != len(out)-PickRows {
		t.Errorf("End 之後視窗頂端 %d,預期 %d(最後一頁)", s.Pick.Top, len(out)-PickRows)
	}
	s.PickHome() // 0xD3
	if s.Pick.Cursor != 0 || s.Pick.Top != 0 {
		t.Errorf("Home 之後 Cursor=%d Top=%d,預期都是 0", s.Pick.Cursor, s.Pick.Top)
	}
	// 清單比視窗短時 End 不該把 Top 推成負的。
	s2 := pickScene(t)
	s2.beginPick("t", []PickEntry{{Label: "a"}, {Label: "b"}}, "", nil)
	s2.PickEnd()
	if s2.Pick.Top != 0 || s2.Pick.Cursor != 1 {
		t.Errorf("短清單 End 之後 Cursor=%d Top=%d,預期 1 / 0", s2.Pick.Cursor, s2.Pick.Top)
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

// 翻頁鍵一次走 7 步(不繞回)。
//
// ⚠ 7 是原版的數字(`sub_1EFC8` 對 0xD5 / 0xD6 把移動次數設成 7),
// 不是「一整頁」——清單再長也是 7。
func TestPickPageMovesSevenRows(t *testing.T) {
	s := &State{MaxMessages: 8}
	var out []PickEntry
	for i := 0; i < 20; i++ {
		out = append(out, PickEntry{Label: "x", Value: i})
	}
	s.beginPick("t", out, "empty", func(int) bool { return true })
	s.PickPage(1)
	if s.Pick.Cursor != PickPageRows {
		t.Errorf("往下翻頁到第 %d 項,預期第 %d 項", s.Pick.Cursor, PickPageRows)
	}
	s.PickPage(-1)
	if s.Pick.Cursor != 0 {
		t.Errorf("往上翻回第 %d 項,預期第 0 項", s.Pick.Cursor)
	}
	// ⚠ 更正:原版**不繞回**。在第 0 項再往上翻,七步全部空轉。
	s.PickPage(-1)
	if s.Pick.Cursor != 0 || s.Pick.Top != 0 {
		t.Errorf("在第一項往上翻該停住,實際 Cursor=%d Top=%d", s.Pick.Cursor, s.Pick.Top)
	}
}

// 原版的清單瀏覽器**沒有字母捷徑** —— 這條把那個更正釘住。
//
// 依據:`sub_1EFC8` 整支比對過的鍵碼只有 1..4 / 0xD3..0xD6 / 13 / 0x20 / 0x1B,
// 0x41..0x5A 一次都沒出現(`docs/re/60`)。所以選單不該去解讀字母鍵 ——
// 若哪天有人「補上字母捷徑」,那是加了原版沒有的東西。
func TestPickerHasNoLetterShortcut(t *testing.T) {
	s := &State{MaxMessages: 8}
	s.beginPick("t", []PickEntry{{Label: "a", Value: 7}, {Label: "b", Value: 9}}, "empty",
		func(int) bool { return true })
	before := s.Pick.Cursor
	// 字母鍵在選單裡不該有任何效果 —— TypeRune 是打字用的路徑。
	s.TypeRune('b')
	if s.Pick == nil {
		t.Fatal("按了字母鍵選單就關了 —— 那是原版沒有的行為")
	}
	if s.Pick.Cursor != before {
		t.Errorf("字母鍵移動了游標(%d → %d)", before, s.Pick.Cursor)
	}
}
