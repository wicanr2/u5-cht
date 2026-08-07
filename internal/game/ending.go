package game

import (
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
