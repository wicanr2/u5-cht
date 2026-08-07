package game

import (
	"fmt"
	"strings"

	"github.com/wicanr2/u5-cht/internal/i18n"
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 結局(原版 `sub_135FC`)
//
//	1. **隊伍裡死掉的人全部復活**(狀態 'D' → 'G',HP 補滿),各印一句「⋯⋯活過來了!」
//	2. 大家走進王座廳,不列顛王開口:「幸會了,<聖者名>!」
//	3. 「汝可帶來了吾的盒子?」—— 答 N 會**再問一次**(換個說法)
//	4. 答 Y **而且真的有那只檀香木盒** → 真結局(打開盒子、月之球、回到舊世界)
//	   否則 → 「原來如此……那麼,搬張椅子坐下吧。」遊戲就停在那裡
//
// ⚠ 兩個條件是 `and`:`cmp ebx,'Y'; jnz 壞結局` **接著** `cmp byte_3DFCD,0; jz 壞結局`。
// 少了後者,玩家嘴上說有就真的有了 —— 而那只盒子是整條主線最後一件收集品。
//
// ⚠ **復活是無條件的**,不看有沒有盒子、也不看回答 —— 它在問話之前就做完了。

// Ending 是進行中的結局。
type Ending struct {
	// Pages 是還沒唸完的段落。
	Pages []string
	// Page 是現在停在第幾段。
	Page int
	// Asking 為真時在等 Y / N(「汝可帶來了吾的盒子?」)。
	Asking bool
	// SecondAsk 記錄已經追問過一次了 —— 原版只追問一次。
	SecondAsk bool
	// Good 記錄走的是不是真結局。
	Good bool
}

// BeginEnding 播結局。回傳有沒有開始。
func (s *State) BeginEnding() bool {
	if s.EndMsg == nil {
		// 沒有 ENDMSG.DAT 就不假裝有結局(誠實跳過比放一段空白好)。
		return false
	}
	s.enterChamber(u5data.MiscMapIndexThrone)
	s.FinishChamberWalk()
	s.revivePartyForEnding()
	s.Ending = &Ending{}
	s.Prompt = PromptEnding
	s.Log(s.endText(u5data.MsgEndWellMet) + s.AvatarName() + "!」")
	s.askForBox(u5data.MsgEndAskBox)
	return true
}

// revivePartyForEnding 把隊伍裡死掉的人全部救回來。
//
// ⚠ 原版順手做的兩件事都要照做:狀態改回 'G',**而且**目前 HP 補成最大 HP
//(`mov dx, word_3DDC6[eax]; mov word_3DDC4[eax], dx` —— 那是 MaxHP → HP)。
// 只改狀態的話,人活過來但血是 0。
func (s *State) revivePartyForEnding() {
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		c := &s.Roster[i]
		if c.Status != u5data.StatusDead {
			continue
		}
		c.Status = u5data.StatusGood
		c.HP = c.MaxHP
		c.Raw[u5data.CharStatus] = u5data.StatusGood
		s.Log(c.Name + "活過來了!")
	}
}

// askForBox 問「汝可帶來了吾的盒子?」。
func (s *State) askForBox(record int) {
	s.Ending.Asking = true
	s.Log(s.endText(record))
}

// AnswerEnding 回答不列顛王的問題。
func (s *State) AnswerEnding(yes bool) {
	e := s.Ending
	if e == nil || !e.Asking {
		return
	}
	if yes {
		s.Log("是。")
	} else {
		s.Log("否。")
	}
	if !yes && !e.SecondAsk {
		// ⚠ 原版答 N 之後會**換個說法再問一次**,不是直接判定。
		e.SecondAsk = true
		s.askForBox(u5data.MsgEndAskBox2)
		return
	}
	e.Asking = false
	// ⚠ 兩個條件都要:嘴上說有,而且背包裡真的有。
	e.Good = yes && s.SandalwoodBox
	if e.Good {
		for _, r := range u5data.EndingFinale {
			e.Pages = append(e.Pages, s.endText(r))
		}
		// 真結局播完接製作名單(原版 `sub_13258`)。
		e.Pages = append(e.Pages, s.creditsPages()...)
	} else {
		e.Pages = append(e.Pages,
			"「原來如此……」",
			s.endText(u5data.MsgEndPullUpAChair))
	}
	s.Log(e.Pages[0])
}

// AdvanceEnding 翻到下一段。唸完就結束並回 false。
func (s *State) AdvanceEnding() bool {
	e := s.Ending
	if e == nil || e.Asking {
		return false
	}
	if e.Page+1 >= len(e.Pages) {
		s.EndEnding()
		return false
	}
	e.Page++
	s.Log(e.Pages[e.Page])
	return true
}

// EndEnding 收掉結局。
//
// ⚠ 原版在這裡**不會回到遊戲**:真結局跑製作名單,壞結局是一個無窮迴圈
//(玩家真的只能坐在那裡)。引擎不做無窮迴圈 —— 那會讓人以為當機 ——
// 改成把「這一局結束了」寫進訊息欄,由呼叫端決定接下來做什麼。
func (s *State) EndEnding() {
	if s.Ending != nil && s.Ending.Good {
		s.Log("(真結局 —— 原版在此播製作名單)")
	} else {
		s.Log("(原版在此停住不動 —— 沒有那只盒子,故事就到這裡)")
	}
	s.leaveChamber()
	s.Ending = nil
	if s.Prompt == PromptEnding {
		s.Prompt = PromptNone
	}
}

// endText 取一筆 `ENDMSG.DAT` 的譯文。
func (s *State) endText(record int) string {
	en := ""
	if s.EndMsg != nil && record >= 0 && record < len(s.EndMsg.Records) {
		en = s.EndMsg.Records[record].Text()
	}
	return strings.TrimSpace(i18n.Text("ENDMSG.DAT", record, en))
}

// 製作名單 / 頒獎狀(原版 `sub_13258`)
//
// 真結局播完之後的那一頁。原版是一頁一頁翻的羊皮紙,這裡照同樣的節奏
// 併進 `Ending` 的翻頁流程。

// creditsPages 組出名單的每一頁。
func (s *State) creditsPages() []string {
	c := s.Clock
	name := i18n.Name(s.AvatarName())
	if name == "" {
		name = "聖者"
	}
	years, months, days := u5data.Elapsed(c.Year, c.Month, c.Day)

	pages := []string{
		fmt.Sprintf("茲證明,於%d年第%d月第%d日,",
			c.Year, c.Month, c.Day),
		name + "  聖者",
		"拯救了吾王不列顛之性命,",
		"從而拯救了吾等子民與這片土地。",
		// ⚠ 這兩行原版是用符文字型畫出來的,維持英文原樣。
		u5data.CreditsRunes[0],
		u5data.CreditsRunes[1],
		"(聖者的追尋,永無止盡)",
	}
	// 原版最後兩行是「Report now, thy Quest compleat in <歷時>」+
	// 「to Lord British at Origin Systems!」—— 當年的註冊回函梗,照留。
	//
	// ⚠ 歷時三個單位全為 0 時原版**印不出東西**(同一天通關)。
	// 照抄,但那一行就不放了,免得畫面上出現「歷時  」。
	if e := elapsedText(years, months, days); e != "" {
		pages = append(pages, "汝之追尋,歷時"+e+"達成 ——")
	}
	return append(pages, "謹報予 Origin Systems 的不列顛王!")
}

// elapsedText 把年 / 月 / 日排成一句。
//
// ⚠ 原版**只印非零的單位**,而且相鄰兩個之間才放逗號
//(`if (年) 印年; if (月 || 日) 印逗號; …`)。三個都是 0 時整句是空的 ——
// 那代表同一天通關,原版就是印不出東西,照抄。
func elapsedText(years, months, days int) string {
	var parts []string
	if years != 0 {
		parts = append(parts, fmt.Sprintf("%d年", years))
	}
	if months != 0 {
		parts = append(parts, fmt.Sprintf("%d個月", months))
	}
	if days != 0 {
		parts = append(parts, fmt.Sprintf("%d天", days))
	}
	return strings.Join(parts, "又")
}

// checkAbsorbed 是 `sub_161E4`:每個單位行動完檢查一次。
//
// 條件見 `u5data.AbsorbRow` / `AbsorbTile` —— 活著的單位站在戰場第 2 列、
// **正北**那一格是城堡外牆或城門,就會被「吸收」,而戰鬥結束時
// (`sub_42CC` 與 `sub_163B0` 兩處都查 `byte_3E0B0 == 'M'`)播結局。
//
// ⚠ **目前引擎的戰場地形全部來自 `BRIT.CBT` / `DUNGEON.CBT` 的固定地圖,
// 而那兩個檔一格 0x3C..0x3F 都沒有**(全檔掃過:BRIT.CBT 0 次,
// DUNGEON.CBT 唯一一次落在第 111 張的中繼資料區,不在 11×11 的格子裡)。
// 城堡地形只存在於**世界地圖**,所以這條規則要真的能觸發,
// 得先做「在野外開打時用周遭地形當場組戰場」那條路(原版 `sub_295AC`)。
// 那一段還沒逆完 —— 規則先照原版放進來,並在 docs/re/34 記下缺口。
func (s *State) checkAbsorbed() bool {
	c := s.Combat
	if c == nil || c.Over || c.Turn < 0 || c.Turn >= len(c.Units) {
		return false
	}
	u := &c.Units[c.Turn]
	if !u.Active() || u.Y != u5data.AbsorbRow {
		return false
	}
	if !u5data.AbsorbTile(c.Map.At(u.X, u5data.AbsorbRow-1)) {
		return false
	}
	s.Log(s.unitName(u) + MsgIsAbsorbed)
	c.Over, c.Won = true, true
	s.EndCombat(true)
	s.BeginEnding()
	return true
}
