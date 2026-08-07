package game

import (
	"strings"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 對話流程
//
// 原版按 T 選方向 → 找到人 → 印外貌與招呼 → 玩家逐字輸入關鍵字 → NPC 回應,
// 打 bye 或按 ESC 結束。關鍵字只比前 4 個字母(u5data.KeywordLen)。
//
// name / job / bye 這三個字**不在記錄的關鍵字表裡** —— 它們由引擎直接對應到
// 記錄的固定欄位。只實作關鍵字表的話,玩家問名字會得到「他聽不懂」。

// 內建關鍵字。這些字玩家一定打得出來(英文),所以 canonical 值維持英文
// (CLAUDE.md §5.2:會拿去跟玩家輸入比對的字串不翻譯)。
const (
	KeywordName = "name"
	KeywordJob  = "job"
	KeywordWork = "work"
	KeywordBye  = "bye"
)

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
	n := found.NPC
	switch {
	case n.Dialogue == u5data.DialogueNone:
		s.Log(MsgNoResponse)
	case n.IsShopkeeper():
		// 商店交易要 SHOPPE.DAT 的內容與營業時間表,還沒接上。
		s.Log(MsgMerchantClosed)
	case n.Dialogue == u5data.DialogueFrightened:
		s.Log(MsgFrightened)
	case n.Dialogue >= u5data.DialogueSpecialFE:
		s.Log(MsgNoResponse)
	default:
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
	if s.Prompt != PromptTalk {
		return
	}
	switch {
	case r >= 'a' && r <= 'z':
	case r >= 'A' && r <= 'Z':
		r = r - 'A' + 'a'
	default:
		return // 關鍵字只有英文字母
	}
	if len(s.Input) >= 16 {
		return
	}
	s.Input += string(r)
}

// Backspace 刪掉輸入列最後一個字元。
func (s *State) Backspace() {
	if s.Prompt == PromptTalk && s.Input != "" {
		s.Input = s.Input[:len(s.Input)-1]
	}
}

// Submit 送出目前的輸入。空字串等同結束對話。
func (s *State) Submit() {
	if s.Prompt != PromptTalk || s.Conv == nil {
		return
	}
	word := strings.TrimSpace(s.Input)
	s.Input = ""
	if word == "" {
		s.EndConversation()
		return
	}
	s.Log("汝問:" + word)
	s.answer(word)
}

func (s *State) answer(word string) {
	c := s.Conv
	lower := strings.ToLower(word)
	switch {
	case strings.HasPrefix(lower, KeywordBye):
		if c.Bye != "" {
			s.Log("「" + oneLine(c.Bye) + "」")
		}
		s.EndConversation()
		return
	case strings.HasPrefix(lower, KeywordName):
		s.Log("「" + oneLine(nonEmpty(c.Name, MsgNoResponse)) + "」")
		return
	case strings.HasPrefix(lower, KeywordJob), strings.HasPrefix(lower, KeywordWork):
		s.Log("「" + oneLine(nonEmpty(c.Job, MsgNoResponse)) + "」")
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
		s.Log(MsgJoinNotImplemented)
	}
	if fx.CallGuards {
		s.Log(MsgGuardsNotImplemented)
	}
	if fx.AsksPlayer {
		s.Log(MsgAskNotImplemented)
	}
	if fx.EndTalk {
		s.EndConversation()
	}
}

// EndConversation 結束對話。
func (s *State) EndConversation() {
	s.Conv = nil
	s.Input = ""
	if s.Prompt == PromptTalk {
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
