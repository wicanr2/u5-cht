package game

import (
	"strings"

	"github.com/wicanr2/u5-cht/internal/i18n"
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 對話流程
//
// 原版按 T 選方向 → 找到人 → 印外貌與招呼 → 玩家逐字輸入關鍵字 → NPC 回應,
// 打 bye 或按 ESC 結束。關鍵字只比前 4 個字母(u5data.KeywordLen)。
//
// 引擎另有一張 34 個字的內建關鍵字表(NAME/JOB/WORK/BYE/THANK + 29 個髒話),
// 它們**不在**記錄的關鍵字表裡,而是直接對應到記錄的固定欄位或固定回應。

// 內建關鍵字的清單在 u5data.BuiltinKeywords —— 一律維持英文,
// 因為那是要拿去跟玩家輸入比對的(CLAUDE.md §5.2)。

// Talk 是原版的「交談」指令(sub_1B658 → sub_1B52C,按鍵 T)。
//
// 原版先問方向,這裡簡化成「看四鄰有沒有人」—— 找到第一個就跟他說話。
// 對話號碼的分派完全照 sub_1B52C(見 u5data 的 Dialogue 常數)。
func (s *State) Talk() {
	if !s.InScene() {
		s.Log(MsgNobodyHere)
		return
	}
	var found *VisibleNPC
	for _, d := range []Direction{North, East, South, West} {
		dx, dy := d.Delta()
		if v, ok := s.NPCAt(s.X+dx, s.Y+dy); ok {
			found = v
			break
		}
	}
	if found == nil {
		s.Log(MsgNobodyHere)
		return
	}
	s.talkToNPC(found.Index)
}

// talkToNPC 對指定槽位的 NPC 起話頭。
//
// 拆出來是因為**對話不只玩家能發起** —— 行為型別 4 / 5 的 NPC 走到玩家
// 身邊時會自己搭話(原版 `sub_8F60` 設 `byte_3EDD0 = 't'`,見 npcai.go)。
func (s *State) talkToNPC(idx int) {
	if s.npcs == nil || idx < 0 || idx >= len(s.npcs) {
		return
	}
	n := &s.npcs[idx]
	found := &VisibleNPC{Index: idx, NPC: n}
	switch {
	case n.Dialogue == u5data.DialogueNone:
		s.Log(MsgNoResponse)
	case n.IsShopkeeper():
		s.talkingTo = found.Index
		s.enterShop(n)
	case n.Dialogue == u5data.DialogueFrightened:
		s.Log(MsgFrightened)
	case n.Dialogue == u5data.DialogueSpecialFE:
		// 「滾開,害蟲!」說完那個人就跑掉(原版 `sub_154C`)。
		s.shooNPC(idx)
	case n.Dialogue == u5data.DialogueGuardChallenge:
		// 攔人盤查(原版 `sub_1B3D0`)。答不出來就當場逮捕。
		if !s.guardChallenge() {
			s.Arrest()
		}
	default:
		s.talkingTo = found.Index
		s.beginConversation(n.Dialogue)
	}
}

func (s *State) beginConversation(dialogue byte) {
	if s.Talks == nil {
		s.Log(MsgNoResponse)
		return
	}
	rec, ok := s.Talks.Record(s.Location, int(dialogue))
	if !ok {
		s.Log(MsgNoResponse)
		return
	}
	c := u5data.ParseConversation(rec, s.Talks.Dict)
	// 對話裡的 opcode 0x81 要插入聖者的名字 —— 從名冊拿真名,不要留佔位符。
	c.AvatarName = s.AvatarName()
	s.Conv = c
	s.Prompt = PromptTalk
	s.Input = ""
	// ⚠ 這裡說的還是英文原文。對話中譯要等譯名對過《軟體世界》手冊之後才做,
	// 現在硬翻會變成之後要整批重來的二手轉譯。
	if c.Description != "" {
		s.Log("汝見到" + c.Description)
	}
	if c.Greeting != "" {
		s.Log("「" + oneLine(c.Greeting) + "」")
	}
}

// TypeRune 把一個字元加進輸入列(對話中用)。
func (s *State) TypeRune(r rune) {
	if s.Prompt != PromptTalk && s.Prompt != PromptAnswer &&
		s.Prompt != PromptSpell && s.Prompt != PromptShrine &&
		s.Prompt != PromptYell && s.Prompt != PromptBlackthorn &&
		s.Prompt != PromptGuard {
		return
	}
	switch {
	case r >= 'a' && r <= 'z':
	case r >= 'A' && r <= 'Z':
		// 咒語名、真言與力量之言保留大小寫原樣(比對本來就不分大小寫),
		// 關鍵字一律小寫。
		if s.Prompt != PromptSpell && s.Prompt != PromptShrine &&
			s.Prompt != PromptYell && s.Prompt != PromptBlackthorn &&
			s.Prompt != PromptGuard {
			r = r - 'A' + 'a'
		}
	case r == ' ' && (s.Prompt == PromptSpell || s.Prompt == PromptBlackthorn):
		// 上古語咒語名有空格(「An Tym」);回答黑棘也可以打一整句。
	case r >= '0' && r <= '9' && s.Prompt == PromptShrine:
		// 聖壇獻金是打數字。
	default:
		return
	}
	// Yell 原版只讀 12 個字元(`sub_17E74` 的 `push 0Ch`);其餘輸入列到 16。
	limit := 16
	switch s.Prompt {
	case PromptYell:
		limit = u5data.YellInputMax
	case PromptGuard:
		limit = u5data.BlackthornPasswordMax
	}
	if len(s.Input) >= limit {
		return
	}
	s.Input += string(r)
}

// Backspace 刪掉輸入列最後一個字元。
func (s *State) Backspace() {
	if (s.Prompt == PromptTalk || s.Prompt == PromptAnswer || s.Prompt == PromptSpell) && s.Input != "" {
		s.Input = s.Input[:len(s.Input)-1]
	}
}

// Submit 送出目前的輸入。
func (s *State) Submit() {
	if s.Conv == nil {
		return
	}
	word := strings.TrimSpace(s.Input)
	s.Input = ""

	// 正在報上名字(opcode 0x88)—— 與一般的提問分支不同,沒有問題區塊。
	if s.askingName {
		s.answerName(word)
		return
	}

	// 正在回答 NPC 的提問。
	if s.Prompt == PromptAnswer {
		q := s.pending
		s.pending = nil
		s.Prompt = PromptTalk
		s.Log(MsgYouRespond + word)
		if q == nil {
			return
		}
		text, fx := s.Conv.Render(q.AnswerQuestion(word))
		if text != "" {
			s.Log("「" + oneLine(text) + "」")
		}
		s.applyEffects(fx)
		return
	}

	if s.Prompt != PromptTalk {
		return
	}
	if word == "" {
		// 原版:空輸入等同 BYE,印道別後結束。
		if s.Conv.Bye != "" {
			s.Log("「" + oneLine(s.Conv.Bye) + "」")
		}
		s.EndConversation()
		return
	}
	s.Log("汝問:" + word)
	s.answer(word)
}

// answer 處理玩家的一句輸入,流程照原版 `sub_1BF08`:
//
//  1. 空輸入 → 印道別、結束對話
//  2. 先掃 34 個**內建**關鍵字(NAME/JOB/WORK/BYE/THANK + 29 個髒話)
//  3. 再掃這筆記錄自己的關鍵字
//  4. 都沒中 → 「I cannot help thee with that.」
//
// 順序不能顛倒:內建的先。有些 NPC 的記錄裡也有 `name` 之類的字,
// 原版是內建的先命中。
func (s *State) answer(word string) {
	c := s.Conv
	switch b := u5data.MatchBuiltin(word); {
	case b == u5data.BuiltinName:
		// 原版:"My name is " + 段 0
		s.Log("「" + MsgMyNameIs + nonEmpty(i18n.Name(c.Name), "?") + "」")
		return
	case b == u5data.BuiltinJob || b == u5data.BuiltinWork:
		s.Log("「" + oneLine(nonEmpty(c.Job, MsgNoResponse)) + "」")
		return
	case b == u5data.BuiltinBye || b == u5data.BuiltinThank:
		if c.Bye != "" {
			s.Log("「" + oneLine(c.Bye) + "」")
		}
		s.EndConversation()
		return
	case b >= u5data.BuiltinProfanity:
		// 原版對所有髒話回同一句。少了這段,對 NPC 罵髒話會變成沒反應。
		s.Log("「" + MsgFoulLanguage + "」")
		return
	}

	text, fx, ok := c.Respond(word)
	if !ok {
		s.Log(MsgDoesNotUnderstand)
		return
	}
	if text != "" {
		s.Log("「" + oneLine(text) + "」")
	}
	s.applyEffects(fx)
}

// applyEffects 套用回應帶來的遊戲狀態改變。
//
// 業報是真的會動的(原版 opcode 0x89/0x8A 加減 byte_3E098,上限 99);
// 加入隊伍與叫衛兵還沒有對應的子系統,先誠實說明而不是假裝發生了。
func (s *State) applyEffects(fx u5data.Effects) {
	if fx.KarmaDelta != 0 {
		s.Karma += fx.KarmaDelta
		if s.Karma > u5data.KarmaMax {
			s.Karma = u5data.KarmaMax
		}
		if s.Karma < 0 {
			s.Karma = 0
		}
	}
	if fx.JoinParty {
		s.joinParty()
	}
	if fx.CallGuards {
		// ⚠ 原版 `sub_C10` 做兩件事:衛兵變敵對**而且**其餘的人一半機率逃跑。
		// 見 npcai.go 的 CallGuards。
		s.Log(MsgGuardsCalled)
		s.EndConversation()
		s.CallGuards()
	}
	if fx.AsksName {
		s.askName()
	}
	if fx.AsksPlayer {
		s.ask(fx.AskCode)
	}
	if fx.EndTalk {
		s.EndConversation()
	}
}

// enterShop 與商人搭話(原版 sub_1B52C 的 0x80–0xFC 分支 → sub_1B294)。
//
// 原版先看營業時間:`sub_9C7C(npc, hour)` 挑出這名商人此刻的排程 slot,
// 只有在店裡(而不是回家睡覺)才談得成,否則回一句「營業時再來」。
// 這裡用同一份排程判斷:商人不在他的工作位置就是沒開門。
//
// 騎在馬上時原版只有馬廄肯招呼你,其餘一律「GET THAT HORSE OUT...」。
func (s *State) enterShop(n *u5data.NPC) {
	if s.Shops == nil {
		s.Log(MsgMerchantClosed)
		return
	}
	shop, ok := s.Shops.ForDialogue(n.Dialogue, s.Location)
	if !ok {
		// 對話號碼 0x89–0xFC 沒有對應的店種(原版只有 0x81–0x88 八種)。
		s.Log(MsgMerchantClosed)
		return
	}
	if s.Transport&0xFE == u5data.TransportHorse && shop.Type != u5data.ShopStable {
		s.Log("「" + MsgGetThatHorseOut + "」")
		return
	}
	name := shop.Name
	if name == "" {
		name = shop.Owner + "的店"
	}
	s.Log(shop.Type.TypeName() + "「" + name + "」")
	// 四句問候語隨機挑一句(原版 sub_111CC 用 random(0,3))。
	if g := s.Shops.Greeting(shop, s.greetVariant(), s.Clock.Hour); g != "" {
		s.Log("「" + oneLine(g) + "」")
	}
	if !s.openShop(shop) {
		// 酒館 / 造船廠 / 旅店的流程還沒逆完 —— 誠實說明,不要假裝談成了。
		s.Log(MsgShopNotImplemented)
	}
}

// greetVariant 挑一句問候語。原版是亂數;這裡用時鐘當種子,
// 讓 headless 截圖可以重現 —— 隨機在測試裡只會製造雜訊。
func (s *State) greetVariant() int {
	return (s.Clock.Hour + s.Clock.Minute) % 4
}

// ask 讓 NPC 反問玩家(原版 opcode 0x91–0x9F → sub_1C0AC)。
//
// 找到碼相同的提問區塊,印出問題。**終端區塊**印完就結束、不要輸入 ——
// 原版是靠「印的過程中碰到會回傳 1 的指令(0x82/0x84)就停」判斷的,
// Gwenno 的 0x94 區塊(道謝 + 入隊)正是這種。
func (s *State) ask(code byte) {
	if s.Conv == nil {
		return
	}
	q, ok := s.Conv.Question(code)
	if !ok {
		// 記錄裡沒有這個碼的區塊 —— 資料或解析的問題,讓它看得見。
		s.Log(MsgNoQuestionBlock)
		return
	}
	text, fx := s.Conv.Render(q.Text)
	if text != "" {
		s.Log("「" + oneLine(text) + "」")
	}
	if q.Terminal {
		// 終端區塊的副作用(入隊、結束)在這裡生效。
		s.applyEffects(fx)
		return
	}
	s.pending = q
	s.Prompt = PromptAnswer
}

// joinParty 讓正在交談的 NPC 加入隊伍(原版 opcode 0x84 → sub_1BB5C)。
//
// 原版的作法很特別,值得照抄而不是自己設計:
//
//  1. 隊伍滿 6 人 → 回「Thou hast no room for me in thy party」,不入隊。
//  2. 把腳本指標**重設到記錄開頭**(sub_1BAA4(0)),讀 3 個位元組 ——
//     那就是這位 NPC 名字的前三個字母。
//  3. 從名冊**尾端往回**掃(index 15 → 1),找名字前三個字母相符的那一筆
//     (比對遮掉 bit7 且大小寫無關,sub_1B140)。
//  4. 把找到的那筆與名冊第 PartySize 格**對調**,隊伍人數 +1。
//     所以「隊伍」不是另一個清單,而是名冊的前綴。
//  5. 把該 NPC 從場景移除。
//
// 為什麼是前三個字母而不是整個名字:名冊裡的名字與對話記錄裡的名字不一定完全一樣
// (對話記錄的名字後面常跟著控制位元組)。三個字母對這 16 人足以唯一識別。
func (s *State) joinParty() {
	if s.Conv == nil {
		return
	}
	if s.PartySize >= u5data.MaxPartySize {
		s.Log(MsgPartyFull)
		return
	}
	prefix := namePrefix(s.Conv.Name)
	if prefix == "" {
		s.Log(MsgNoMatch)
		return
	}
	for i := len(s.Roster) - 1; i >= 1; i-- {
		if namePrefix(s.Roster[i].Name) != prefix {
			continue
		}
		if i < s.PartySize {
			return // 已經在隊伍裡了
		}
		n := s.PartySize
		s.Roster[i], s.Roster[n] = s.Roster[n], s.Roster[i]
		s.PartySize++
		s.removeCurrentNPC()
		s.Log(s.Roster[n].Name + MsgJoined)
		return
	}
	// 原版在這裡印 "System Error - No Match!" —— 照樣要讓它看得見,
	// 因為那代表對話記錄與名冊對不上,是資料或解析的問題,不該靜默吞掉。
	s.Log(MsgNoMatch)
}

// namePrefix 取名字的前三個字母,小寫。原版比對就是這樣做的(遮 bit7、大小寫無關)。
func namePrefix(name string) string {
	n := strings.ToLower(strings.TrimSpace(name))
	if len(n) > u5data.JoinNameLen {
		n = n[:u5data.JoinNameLen]
	}
	return n
}

// removeCurrentNPC 把正在交談的那個 NPC 從場景移除(入隊之後他不該還站在原地)。
func (s *State) removeCurrentNPC() {
	if s.talkingTo <= 0 {
		return // 0 號槽是隊伍自己,不會是交談對象
	}
	// ⚠ 入隊要記進**存檔**,不只是這一次進場景 ——
	// 只設 session 的話,離場再回來城裡會多出一個分身。
	s.markNPCRemoved(s.talkingTo)
	s.removeNPC(s.talkingTo)
	s.talkingTo = 0
}

// EndConversation 結束對話。
func (s *State) EndConversation() {
	s.Conv = nil
	s.Input = ""
	s.pending = nil
	s.talkingTo = 0
	if s.Prompt == PromptTalk || s.Prompt == PromptAnswer {
		s.Prompt = PromptNone
	}
}

// oneLine 把多行回應壓成一行 —— 訊息欄自己會依寬度斷行(internal/textlayout),
// 原文裡的換行是給原版 320×200 版面用的,照搬會在中文版面上斷得很怪。
func oneLine(s string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(s, "\n", " ")), " ")
}

func nonEmpty(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

// NPC 反問「汝名為何?」(原版 opcode 0x88 → `sub_1C2FC`)
//
// 這一格在 `docs/re/06` 的指令表裡原本標「未定」。追出來的流程:
//
//	1. 印「汝名為何?」,收一行輸入
//	2. 拿去跟**隊伍每一名成員**的名字比 —— 前 4 個字元的詞首比對
//	   (`sub_27C98`,與對話關鍵字同一支)
//	3. 對上 → 記下「這座城的這個 NPC 認得汝了」,回「幸會!」
//	   對不上 → 「汝說是就是吧。」
//
// ⚠ 兩個容易漏的細節:
//
//   - **空輸入直接「汝說是就是吧。」**(原版 `mov eax, edi`(此時 edi 還是 0)
//     配 `cmp al, byte_55F38` —— 比的是輸入緩衝的第一個位元組是不是 0)。
//     第一次讀這段時我把它讀成「報對方自己的名字」,那是錯的。
//   - **比對長度是 4,不是 9。** 原版只把成員名字的前 4 個位元組複製進
//     needle 緩衝。與對話關鍵字共用同一支比對函式,但截斷長度不同 ——
//     用 9 的話「Elwo」報不進去。

// askName 是 NPC 開口問名字。
func (s *State) askName() {
	s.Log(MsgWhatIsThyName)
	s.askingName = true
	s.Prompt = PromptAnswer
}

// answerName 處理玩家報上的名字。
func (s *State) answerName(word string) {
	s.askingName = false
	s.Prompt = PromptTalk
	s.Log(MsgYouRespond + word)

	if word == "" {
		s.Log("「" + MsgIfYouSaySo + "」")
		return
	}
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		if !u5data.NameSpoken(s.Roster[i].Name, word) {
			continue
		}
		s.markIntroduced()
		s.Log("「" + MsgAPleasure + "」")
		return
	}
	s.Log("「" + MsgIfYouSaySo + "」")
}

// markIntroduced 記下「這座城的這個 NPC 認得汝了」(原版 `sub_1C1AC`)。
//
// ⚠ 遮罩是**每個地點一個 32 位元**(`dword_3E3E8[地點] |= 1 << 槽`)——
// 同一個人在別的地點不算認得。做成全域的話,汝一報名字全世界都認識汝。
func (s *State) markIntroduced() {
	if s.talkingTo < 0 || s.talkingTo >= 32 {
		return
	}
	if s.Location < 0 || s.Location >= len(s.Introduced) {
		return
	}
	s.Introduced[s.Location] |= 1 << uint(s.talkingTo)
}

// KnowsThyName 回報這座城的第 slot 個 NPC 認不認得汝。
func (s *State) KnowsThyName(slot int) bool {
	if slot < 0 || slot >= 32 || s.Location < 0 || s.Location >= len(s.Introduced) {
		return false
	}
	return s.Introduced[s.Location]&(1<<uint(slot)) != 0
}
