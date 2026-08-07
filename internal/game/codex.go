package game

import (
	"strings"

	"github.com/wicanr2/u5-cht/internal/i18n"
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 終極智慧之寶典(原版 `sub_1D850`)
//
// 聖壇給汝一項試煉(「去寶典學 X」),寶典就是去領答案的地方。
// **兩者靠兩組不同的位元銜接**:
//
//	byte_3E0DC 進行中的試煉 —— 聖壇設(冥想過關時)
//	byte_3E0DE 已在寶典讀到 —— **寶典設**,聖壇只讀不寫
//
// ⚠ 少了寶典這一段,聖壇會退化成「拜兩次就領獎」。這正是這一支存在的理由:
// 在它做出來之前,引擎是把兩個位元一起設的(見 shrine.go 的說明)。
//
// 流程照抄:
//
//	1. 「書已翻至汝所尋的那一頁!」
//	2. 「汝在那神聖的一頁上讀到:」
//	3. 找**索引最小**的那一項進行中的試煉
//	     沒有 → 「汝是怎麼來到這裡的?」結束
//	     有   → 設下已讀位元,唸出那一句箴言
//	4. 八德**全部**讀完的那一次,額外翻出四段符文(通往末日之路)

// Codex 是進行中的閱讀。一頁一個按鍵,與開場動畫同樣的節奏。
type Codex struct {
	// Pages 是這一次要唸的每一頁。
	Pages []string
	// Page 是現在停在第幾頁。
	Page int
}

// Text 是這一頁的文字。
func (c *Codex) Text() string {
	if c.Page < 0 || c.Page >= len(c.Pages) {
		return ""
	}
	return c.Pages[c.Page]
}

// ReadCodex 開始讀寶典。回傳有沒有讀成。
func (s *State) ReadCodex() bool {
	if s.InScene() || s.InCombat() || s.InDungeon() {
		s.Log("此處沒有寶典。")
		return false
	}
	pages := []string{
		s.miscText(u5data.MsgCodexApproach),
		s.miscText(u5data.MsgCodexOpen),
		s.miscText(u5data.MsgCodexPage),
	}
	// ⚠ 找的是**索引最小**的那一項,不是最後領的那一項
	//(原版 `for (i = 0; i < 8 && !(byte_3E0DC & (1<<i)); i++)`)。
	// 同時進行多項試煉時,寶典先給編號小的那一項。
	v := -1
	for i := 0; i < u5data.VirtueCount; i++ {
		if s.ShrineQuestActive&(1<<uint(i)) != 0 {
			v = i
			break
		}
	}
	if v < 0 {
		// 沒有進行中的試煉 —— 原版在這裡問「汝是怎麼來到這裡的?」
		pages = append(pages, s.miscText(u5data.MsgCodexNoQuest))
		return s.beginCodex(pages)
	}
	s.ShrineQuestGiven |= 1 << uint(v)
	pages = append(pages, "「"+s.miscText(u5data.CodexAnswerRecord(v))+"」")

	// ⚠ 原版這裡**沒有防重播的旗標** —— 它就只是 `cmp byte_3E0DE, 0FFh`,
	// 設完位元之後測一次。「符文只出現一次」是**流程逼出來的**,不是擋出來的:
	// 八德都領過之後,聖壇不會再發新的試煉(`given` 那一位已經有了,
	// 只剩捐錢那條路),所以再來寶典時找不到進行中的試煉,連這一行都到不了。
	// 照抄成有旗標的版本會偏離原版,而且在「八德全給、又有試煉」這種
	// 只有改存檔才做得出來的狀態下行為不同。
	if s.ShrineQuestGiven == 0xFF {
		pages = append(pages, s.miscText(u5data.MsgCodexWind))
		for i := 0; i < u5data.CodexRunePages; i++ {
			pages = append(pages, s.miscText(u5data.MsgCodexRuneOne+i))
		}
	}
	return s.beginCodex(pages)
}

// beginCodex 進入閱讀模式並唸出第一頁。
func (s *State) beginCodex(pages []string) bool {
	s.Codex = &Codex{Pages: pages}
	s.Prompt = PromptCodex
	s.Log(s.Codex.Text())
	return true
}

// AdvanceCodex 翻到下一頁。唸完就結束並回 false。
func (s *State) AdvanceCodex() bool {
	if s.Codex == nil {
		return false
	}
	if s.Codex.Page+1 >= len(s.Codex.Pages) {
		s.EndCodex()
		return false
	}
	s.Codex.Page++
	s.Log(s.Codex.Text())
	return true
}

// EndCodex 收掉閱讀。
//
// ⚠ **離開時要走掉十六分鐘**(原版 `sub_1DA10` 尾端的 `sub_29304(0x10)`)——
// 進出聖壇與寶典都一樣。少了它,玩家可以在寶典前面反覆進出而時間不動。
func (s *State) EndCodex() {
	s.Codex = nil
	if s.Prompt == PromptCodex {
		s.Prompt = PromptNone
	}
	s.AdvanceTime(u5data.ShrineChamberMinutes)
}

// miscText 取一筆 `MISCMSG.DAT` 的譯文(整理掉前後空白與換行)。
func (s *State) miscText(record int) string {
	en := ""
	if s.Misc != nil && record >= 0 && record < len(s.Misc.Records) {
		en = s.Misc.Records[record].Text()
	}
	return strings.TrimSpace(i18n.Text("MISCMSG.DAT", record, en))
}
