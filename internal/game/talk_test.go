package game

import (
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/i18n"
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// TestTalkFlowWithRealData 用原版資料跑一整段對話:
// 走到 NPC 旁邊 → T → 問 job → 問一個他認得的關鍵字 → bye 結束。
func TestTalkFlowWithRealData(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	scenes, err := u5data.LoadSceneSet(dir)
	if err != nil {
		t.Fatal(err)
	}
	npcs, err := u5data.LoadNPCSet(dir)
	if err != nil {
		t.Fatal(err)
	}
	talks, err := u5data.LoadTalkSet(dir)
	if err != nil {
		t.Fatal(err)
	}

	s := &State{Scenes: scenes, NPCs: npcs, Talks: talks, Clock: NewClock(), MaxMessages: 32}
	if err := s.SetScene(britain, 0, 0, 0); err != nil {
		t.Fatal(err)
	}
	// 找一個此刻在這一層、且有正常對話的 NPC,站到他旁邊。
	var target *VisibleNPC
	for _, v := range s.VisibleNPCs() {
		if v.NPC.Dialogue > 0 && v.NPC.Dialogue < u5data.DialogueShopFirst {
			vv := v
			target = &vv
			break
		}
	}
	if target == nil {
		t.Skip("不列顛城此刻沒有可交談的 NPC")
	}
	s.X, s.Y = target.X, target.Y-1 // 站在他北邊,他就在南邊

	talkAround(s)
	if s.Prompt != PromptTalk || s.Conv == nil {
		t.Fatalf("按 T 沒有進入對話:%v", s.Messages)
	}
	name := s.Conv.Name

	// 問 job —— 這是引擎內建的關鍵字,不在記錄的關鍵字表裡。
	//
	// ⚠ **不要拿英文原文當預期值。** 這一條原本比對 `s.Conv.Job` 的英文,
	// 對白中譯之後就整批紅了 —— 而紅的原因是「翻對了」。
	// 測的是**流程**(問了有回應、回應不是「聽不懂」),不是文字內容。
	before := len(s.Messages)
	typeWord(s, "job")
	if got := s.Messages[before:]; len(got) < 2 || anyContains(got, MsgDoesNotUnderstand) {
		t.Errorf("問 job 沒有得到職業回答:%v", got)
	}

	// 問一個他自己列出的關鍵字。
	if kws := s.Conv.Keywords(); len(kws) > 0 {
		want, _, _ := s.Conv.Respond(kws[0])
		before = len(s.Messages)
		typeWord(s, kws[0])
		if got := s.Messages[before:]; want != "" && anyContains(got, MsgDoesNotUnderstand) {
			t.Errorf("問 %q(原文回應 %q)卻聽不懂:%v", kws[0], want, got)
		}
	}

	// 打不存在的字 → 聽不懂
	typeWord(s, "zzzz")
	if last := s.Messages[len(s.Messages)-1]; last != MsgDoesNotUnderstand {
		t.Errorf("問不存在的關鍵字得到 %q,預期 %q", last, MsgDoesNotUnderstand)
	}

	// bye → 結束
	typeWord(s, "bye")
	if s.Prompt != PromptNone || s.Conv != nil {
		t.Errorf("打 bye 之後對話沒結束(對象 %s)", name)
	}
}

func TestTypeRuneOnlyAcceptsLetters(t *testing.T) {
	s := &State{Prompt: PromptTalk}
	for _, r := range []rune{'a', 'B', '1', '-', ' ', '中', 'z'} {
		s.TypeRune(r)
	}
	if s.Input != "abz" {
		t.Errorf("輸入列是 %q,預期 %q —— 關鍵字只收英文字母且自動轉小寫", s.Input, "abz")
	}
	s.Backspace()
	if s.Input != "ab" {
		t.Errorf("退格後是 %q", s.Input)
	}
}

// TestTypeRuneIgnoredOutsideTalk:不在對話中時打字不該累積,
// 否則玩家按 E/K/T 這些指令鍵會被偷偷記進輸入列。
func TestTypeRuneIgnoredOutsideTalk(t *testing.T) {
	s := &State{}
	s.TypeRune('a')
	if s.Input != "" {
		t.Errorf("非對話狀態卻收了字:%q", s.Input)
	}
}

func typeWord(s *State, w string) {
	for _, r := range w {
		s.TypeRune(r)
	}
	s.Submit()
}

func anyContains(msgs []string, sub string) bool {
	if sub == "" {
		return true
	}
	for _, m := range msgs {
		if strings.Contains(m, sub) {
			return true
		}
	}
	return false
}

// firstWords 取回應的前幾個字當比對錨點 —— 訊息欄會把換行壓成空白,
// 整句比對會因為空白差異而假失敗。
func firstWords(s string) string {
	f := strings.Fields(strings.ReplaceAll(s, "\n", " "))
	if len(f) > 3 {
		f = f[:3]
	}
	return strings.Join(f, " ")
}

// TestGwennoJoinsParty 用原版資料跑一次完整的入隊:
// 走到 Gwenno 旁邊 → T → 問 join → 她道謝 → 進隊伍 → 從場上消失。
//
// 這條路徑同時驗了四件事:關鍵字表在 0x90 結束(否則 join 會接到別的回應)、
// 提問區塊解析、終端區塊不向玩家要輸入、名冊對調式的入隊。
func TestGwennoJoinsParty(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	s := realState(t, dir)
	// Gwenno 是不列顛城 8 號槽,中午在 (3,29) 地面層。
	if err := s.SetScene(britain, 0, 3, 28); err != nil {
		t.Fatal(err)
	}
	s.Clock.Hour = 12
	before := s.PartySize
	seen := len(s.VisibleNPCs())

	talkAround(s)
	if s.Conv == nil || s.Conv.Name != "Gwenno" {
		t.Fatalf("沒跟 Gwenno 說到話:%v", s.Messages)
	}
	typeWord(s, "join")

	if s.PartySize != before+1 {
		t.Fatalf("隊伍人數 %d,預期 %d:%v", s.PartySize, before+1, s.Messages)
	}
	if got := s.Roster[before].Name; got != "Gwenno" {
		t.Errorf("隊伍新成員是 %q,預期 Gwenno —— 入隊是把名冊該筆與隊伍位置對調", got)
	}
	if now := len(s.VisibleNPCs()); now != seen-1 {
		t.Errorf("場上還剩 %d 人,預期 %d —— 入隊後她應該從場景消失", now, seen-1)
	}
}

// TestGwennoYesNoBranch:NPC 反問之後,答 y 與答 n 走不同分支。
func TestGwennoYesNoBranch(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	// ⚠ 原本比對英文原文("Then perhaps" / "I thought not"),對白中譯之後
	// 整批紅 —— 而紅的原因是**翻對了**。改成測「兩條分支的回應不一樣」,
	// 那才是分支有沒有接對的本質,而且與譯到哪裡無關。
	reply := map[string]string{}
	for _, answer := range []string{"y", "n"} {
		s := realState(t, dir)
		if err := s.SetScene(britain, 0, 3, 28); err != nil {
			t.Fatal(err)
		}
		s.Clock.Hour = 12
		talkAround(s)
		typeWord(s, "yew") // 這句的回應會拋出提問碼 0x91
		if s.Prompt != PromptAnswer {
			t.Fatalf("問 yew 之後沒有進入回答模式:%v", s.Messages)
		}
		before := len(s.Messages)
		typeWord(s, answer)
		got := s.Messages[before:]
		if len(got) < 2 {
			t.Errorf("答 %q 沒有得到回應:%v", answer, got)
			continue
		}
		reply[answer] = strings.Join(got, " ")
	}
	if reply["y"] == reply["n"] {
		t.Errorf("答 y 與答 n 的回應一樣,分支沒接對:%q", reply["y"])
	}
}

// TestProfanityGetsRebuke:對 NPC 罵髒話有固定回應(原版 29 個字都導到同一句)。
func TestProfanityGetsRebuke(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	s := realState(t, dir)
	if err := s.SetScene(britain, 0, 3, 28); err != nil {
		t.Fatal(err)
	}
	s.Clock.Hour = 12
	talkAround(s)
	typeWord(s, "damn")
	if last := s.Messages[len(s.Messages)-1]; last != "「"+MsgFoulLanguage+"」" {
		t.Errorf("罵髒話得到 %q,預期 %q", last, MsgFoulLanguage)
	}
}

func realState(t *testing.T, dir string) *State {
	t.Helper()
	scenes, err := u5data.LoadSceneSet(dir)
	if err != nil {
		t.Fatal(err)
	}
	npcs, err := u5data.LoadNPCSet(dir)
	if err != nil {
		t.Fatal(err)
	}
	talks, err := u5data.LoadTalkSet(dir)
	if err != nil {
		t.Fatal(err)
	}
	sv, err := u5data.LoadSave(dir + "/SAVED.GAM")
	if err != nil {
		t.Fatal(err)
	}
	// 咒語表與符文表也載進來 —— 少了它們,凡是碰到施法 / 調藥的測試都會
	// **靜靜跳過**,而一直跳過的測試等於沒有測試。
	spells, err := u5data.LoadSpells(dir)
	if err != nil {
		t.Fatal(err)
	}
	runes, err := u5data.LoadRuneTable(dir)
	if err != nil {
		t.Fatal(err)
	}
	s := &State{Scenes: scenes, NPCs: npcs, Talks: talks,
		Spells: spells, Runes: runes, MaxMessages: 64}
	s.LoadFrom(sv)
	return s
}

// talkAround 是測試用的便利函式。
//
// 原版的 Talk 會**先問方向**(見 `TalkToward`),而多數測試只關心
// 「跟旁邊那個人說到話」,不關心方向選單。這裡四個方向試一輪。
func talkAround(s *State) {
	for _, d := range []Direction{North, East, South, West} {
		s.Talk()
		if !s.AwaitingDirection() {
			return
		}
		s.AnswerDirection(d)
		// 只要不是「無人在此」就當作談到了(可能進對話、進盤查,
		// 也可能只是被回一句「滾開,害蟲!」)。
		if len(s.Messages) > 0 && s.Messages[len(s.Messages)-1] != MsgNobodyHere {
			return
		}
	}
}

// 對話本文的譯文覆蓋層:有譯文就換掉,而且**只換顯示的字**。
//
// 用真的原版資料跑 —— 這一條同時在驗 key 對得上:
// 譯文掛在 `CASTLE.TLK#1#job`,而地點 17 的對話正好出自 `CASTLE.TLK`。
// 檔名算錯的話譯文查不到,測試會退回英文而失敗。
func TestTalkUsesTheChineseOverlay(t *testing.T) {
	dir := gameDataDir(t)
	if dir == "" {
		return
	}
	s := realState(t, dir)
	if err := s.SetScene(17, 0, 15, 15); err != nil {
		t.Fatal(err)
	}
	s.beginConversation(1) // Alistair the Bard
	if s.Conv == nil {
		t.Fatal("開不了對話")
	}
	if s.convFile != "CASTLE.TLK" {
		t.Fatalf("對話檔算成 %q,應該是 CASTLE.TLK", s.convFile)
	}
	s.Input = "job"
	s.Submit()
	if !strings.Contains(allLogs(s), "樂音") {
		t.Errorf("job 沒走中文譯文:%q", allLogs(s))
	}
}

// 沒翻的段落照樣出英文 —— 半套中文比整段消失好。
func TestUntranslatedTalkStaysEnglish(t *testing.T) {
	dir := gameDataDir(t)
	if dir == "" {
		return
	}
	s := realState(t, dir)
	if err := s.SetScene(17, 0, 15, 15); err != nil {
		t.Fatal(err)
	}
	// 找一筆**確定還沒翻**的外觀敘述 —— 隨著翻譯推進,寫死號碼會失效。
	for id := 1; id < 60; id++ {
		s.EndConversation()
		s.Messages = nil
		s.beginConversation(byte(id))
		if s.Conv == nil || strings.TrimSpace(s.Conv.Description) == "" {
			continue
		}
		if i18n.TalkTranslated("CASTLE.TLK", s.Conv.ID, i18n.TalkFieldDesc) {
			continue
		}
		if !strings.Contains(allLogs(s), s.Conv.Description) {
			t.Errorf("沒翻的段落應該原樣出英文:%q", allLogs(s))
		}
		return
	}
	t.Skip("這個地點的外觀敘述都翻完了 —— 這是好事")
}

// ─── 對話 opcode 的引擎端(`docs/re/79`)──────────────────────────────

// TestDemandGoldTakesTheMoneyOrRefuses —— opcode 0x85 的兩條路。
func TestDemandGoldTakesTheMoneyOrRefuses(t *testing.T) {
	s := upkeepScene(t)
	s.Inventory.Gold = 250
	s.Messages = nil
	s.demandGold(100)
	if s.Inventory.Gold != 150 {
		t.Errorf("付了 100 之後剩 %d,預期 150", s.Inventory.Gold)
	}
	if strings.Contains(strings.Join(s.Messages, "|"), MsgNotEnoughGold) {
		t.Error("付得出來卻說錢不夠")
	}
	// 付不出來:印那句話,而且**一枚都不扣**。
	s.Messages = nil
	s.demandGold(999)
	if s.Inventory.Gold != 150 {
		t.Errorf("付不出來卻扣了錢:剩 %d", s.Inventory.Gold)
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgNotEnoughGold) {
		t.Errorf("沒印「%s」:%q", MsgNotEnoughGold, s.Messages)
	}
	// ★ 剛好夠要付得出來(原版 `cmp edi, eax; jg` 是「大於」才拒絕)。
	s.Messages = nil
	s.demandGold(150)
	if s.Inventory.Gold != 0 {
		t.Errorf("剛好夠卻沒付:剩 %d", s.Inventory.Gold)
	}
}

// TestGiveThingCoversAllElevenResourcesAndEquipment —— opcode 0x86。
func TestGiveThingCoversAllElevenResourcesAndEquipment(t *testing.T) {
	s := upkeepScene(t)
	// 裝備:參數就是背包索引。
	before := s.Inventory.Items[0x17]
	s.giveThing(0x17)
	if s.Inventory.Items[0x17] != before+1 {
		t.Errorf("裝備 0x17 沒 +1:%d → %d", before, s.Inventory.Items[0x17])
	}

	type check struct {
		code byte
		name string
		get  func() int
	}
	inv := &s.Inventory
	checks := []check{
		{u5data.GiveFood, "糧食", func() int { return inv.Food }},
		{u5data.GiveGold, "金幣", func() int { return inv.Gold }},
		{u5data.GiveKeys, "鑰匙", func() int { return inv.Keys }},
		{u5data.GiveGems, "寶石", func() int { return inv.Gems }},
		{u5data.GiveTorches, "火把", func() int { return inv.Torches }},
		{u5data.GiveGrapple, "抓鉤", func() int { return inv.Grapple }},
		{u5data.GiveCarpet, "魔毯", func() int { return inv.Carpets }},
		{u5data.GiveSkullKey, "骷髏鑰匙", func() int { return inv.OddKeys }},
	}
	for _, c := range checks {
		was := c.get()
		s.giveThing(c.code)
		if got := c.get(); got != was+1 {
			t.Errorf("%s(%q)沒 +1:%d → %d", c.name, string(c.code), was, got)
		}
	}
	// ★ 三個信物是**直接設成有**,不是 +1。
	s.HasSextant, s.HasSpyglass, s.HasBadge = false, false, false
	s.giveThing(u5data.GiveSextant)
	s.giveThing(u5data.GiveSpyglass)
	s.giveThing(u5data.GiveBadge)
	if !s.HasSextant || !s.HasSpyglass || !s.HasBadge {
		t.Errorf("三個信物沒設起來:六分儀 %v 望遠鏡 %v 徽章 %v",
			s.HasSextant, s.HasSpyglass, s.HasBadge)
	}
}

// TestTheThreeNauticalFlagsSurviveASaveRoundTrip —— 位移釘死了才做得到。
//
// `docs/re/44` §4 原本記「望遠鏡 / 六分儀 / 懷錶的持有旗標存檔位移還沒釘死,
// 所以 U 的清單無條件列出來」。現在釘死了(`docs/re/79`),所以:
// 存得進去、讀得回來,而清單也照旗標列。
func TestTheThreeNauticalFlagsSurviveASaveRoundTrip(t *testing.T) {
	s := upkeepScene(t)
	if s.BaseSave == nil {
		t.Skip("沒有底稿存檔")
	}
	s.HasSpyglass, s.HasSextant, s.HasWatch = true, false, true
	sv, err := s.ExportSave(s.BaseSave)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := sv.Encode()
	if err != nil {
		t.Fatal(err)
	}
	back, err := u5data.ParseSave(raw)
	if err != nil {
		t.Fatal(err)
	}
	if back.Spyglass == 0 || back.Sextant != 0 || back.Watch == 0 {
		t.Errorf("旗標沒 round-trip:望遠鏡 %#x 六分儀 %#x 懷錶 %#x",
			back.Spyglass, back.Sextant, back.Watch)
	}
	// 而且清單要照旗標列。
	s.HasSpyglass, s.HasSextant, s.HasWatch = true, false, false
	seen := map[int]bool{}
	for _, e := range s.usableEntries() {
		seen[e.Value] = true
	}
	if !seen[UseSpyglass] {
		t.Error("有望遠鏡卻沒列出來")
	}
	if seen[UseSextant] || seen[UseWatch] {
		t.Error("沒有的東西還列在清單上 —— 此前是無條件列出,現在該照旗標")
	}
}
