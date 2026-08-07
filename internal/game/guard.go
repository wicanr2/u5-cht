package game

import (
	"fmt"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 衛兵的盤查(原版 `sub_1B3D0`,對話號碼 0xFF)
//
// 引擎原本只有「被逮捕之後會怎樣」,沒有「為什麼會被逮捕」。
// 這一層補上中間那一段:攔下來、開口要東西、給了就過、不給就抓。
//
// ⚠ 三種盤查是**互斥**的,依地點分派(見 `u5data/guard.go`)。

// GuardChallenge 是進行中的盤查。
type GuardChallenge struct {
	// Tribute 是這次要繳多少(密語那一種為 0)。
	Tribute int
	// Password 為真時要打字答密語,否則是 Y / N。
	Password bool
}

// guardChallenge 跑一次盤查(原版 `sub_1B3D0`)。
//
// 回傳 true 代表**當場結束**(給了錢、或已經進到等玩家回答的狀態);
// false 代表盤查失敗,呼叫端要接著逮捕。
func (s *State) guardChallenge() bool {
	switch {
	case s.Location == u5data.BlackthornLocation:
		// 黑棘的宮殿:沒戴徽章連問都不問,直接抓。
		if s.CombatMode != u5data.BadgeMode {
			return false
		}
		s.Log(MsgGuardPassword)
		s.Guard = &GuardChallenge{Password: true}
		s.Prompt = PromptGuard
		return true
	case s.Location == u5data.GuardHalfGoldLocation:
		s.Log(MsgGuardHalfGold)
		s.Guard = &GuardChallenge{Tribute: s.Inventory.Gold / 2}
		s.Prompt = PromptGuard
		return true
	default:
		n := u5data.GuardTribute(s.livingCompanions())
		s.Log(fmt.Sprintf(MsgGuardTribute, n))
		s.Guard = &GuardChallenge{Tribute: n}
		s.Prompt = PromptGuard
		return true
	}
}

// AnswerGuard 回答「汝可願付?」。
//
// ⚠ 付不出來也是逮捕(原版 `cmp eax, esi; jb loc_1B522`)——
// 「我願意但沒錢」不是理由。
func (s *State) AnswerGuard(pay bool) {
	g := s.Guard
	if g == nil {
		return
	}
	s.endGuard()
	if !pay {
		s.Log(MsgGuardNo)
		s.Arrest()
		return
	}
	s.Log(MsgGuardYes)
	if s.Inventory.Gold < g.Tribute {
		// 原版在這裡不另外說話,直接進逮捕。
		s.Arrest()
		return
	}
	s.Inventory.Gold -= g.Tribute
}

// SubmitGuard 交出打好的密語。
func (s *State) SubmitGuard() {
	if s.Guard == nil || !s.Guard.Password {
		return
	}
	typed := s.Input
	s.endGuard()
	if !u5data.PasswordMatches(u5data.BlackthornPassword, typed) {
		s.Arrest()
		return
	}
	s.Log(MsgGuardPass)
}

// CancelGuard 不理會盤查 —— 與答「不」同義,原版沒有第三條路。
func (s *State) CancelGuard() {
	if s.Guard == nil {
		return
	}
	if s.Guard.Password {
		s.endGuard()
		s.Arrest()
		return
	}
	s.AnswerGuard(false)
}

func (s *State) endGuard() {
	s.Guard = nil
	s.Input = ""
	if s.Prompt == PromptGuard {
		s.Prompt = PromptNone
	}
}

// shooNPC 是對話號碼 0xFE 的反應(原版 `sub_154C` → `sub_B98`)。
//
// 「滾開,害蟲!」說完那個人就跑掉 —— 對話號碼改成 0xFD(怕你),
// 行為型別改成 3(逃跑)。
//
// ⚠ 有兩個前提,少了就會讓不該跑的東西也跑:
//
//	生物編號要在 0x40..0x73(是人,不是動物或怪物)
//	而且**原本就是 0xFE、或本來有排程**(是這座城的居民)
func (s *State) shooNPC(i int) {
	s.Log(MsgBegoneVermin)
	if s.npcs == nil || i < 0 || i >= len(s.npcs) {
		return
	}
	n := &s.npcs[i]
	if n.Creature < 0x40 || n.Creature >= 0x74 {
		return
	}
	if n.Dialogue != u5data.DialogueSpecialFE && !n.Schedule.Scheduled() {
		return
	}
	n.Dialogue = u5data.DialogueFrightened
	s.makeFlee(i)
}
