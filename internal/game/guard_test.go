package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// guardScene 造一個城鎮場景,在玩家北邊放一個對話號碼 0xFF 的攔路衛兵。
func guardScene(t *testing.T, location int) *State {
	t.Helper()
	s := &State{Scenes: synthScenes(t, walkable(t)), NPCs: &u5data.NPCSet{}, MaxMessages: 16}
	s.Roster = []u5data.Character{
		{Name: "Avatar", Status: u5data.StatusGood},
		{Name: "Iolo", Status: u5data.StatusGood},
		{Name: "Shamino", Status: u5data.StatusGood},
	}
	s.PartySize = 3
	s.Inventory.Gold = 1000
	if err := s.SetScene(location, 0, 15, 15); err != nil {
		t.Fatalf("進不了地點 %d:%v", location, err)
	}
	// 1 號槽放衛兵,站在玩家北邊。
	n := &s.npcs[1]
	n.Creature = u5data.CreatureGuard
	n.Dialogue = u5data.DialogueGuardChallenge
	for k := range n.Schedule.X {
		n.Schedule.X[k], n.Schedule.Y[k], n.Schedule.Floor[k] = 15, 14, 0
	}
	n.Schedule.Times[0] = 1
	s.initRuntimeNPCs()
	return s
}

func lastLog(s *State) string {
	if len(s.Messages) == 0 {
		return ""
	}
	return s.Messages[len(s.Messages)-1]
}

func allLogs(s *State) string { return strings.Join(s.Messages, "\n") }

// 一般城鎮:每個活著的隊員 10 gp 貢金。
func TestGuardDemandsTenGoldPerLivingMember(t *testing.T) {
	s := guardScene(t, britain)
	s.Talk()
	if s.Prompt != PromptGuard {
		t.Fatalf("跟攔路衛兵說話後 Prompt 是 %v,應該進盤查", s.Prompt)
	}
	if s.Guard == nil || s.Guard.Tribute != 30 {
		t.Fatalf("三個人應該要 30 gp,實得 %+v", s.Guard)
	}
	if !strings.Contains(allLogs(s), "30") {
		t.Errorf("訊息裡沒有金額:%q", allLogs(s))
	}
	s.AnswerGuard(true)
	if s.Inventory.Gold != 970 {
		t.Errorf("付完剩 %d gp,應該是 970", s.Inventory.Gold)
	}
	if s.Prompt == PromptGuard {
		t.Error("付完之後盤查應該結束")
	}
}

// 死掉的隊員不用繳。
func TestDeadCompanionsPayNoTribute(t *testing.T) {
	s := guardScene(t, britain)
	s.Roster[2].Status = u5data.StatusDead
	s.Talk()
	if s.Guard == nil || s.Guard.Tribute != 20 {
		t.Fatalf("兩個活人應該要 20 gp,實得 %+v", s.Guard)
	}
}

// 不付 → 當場逮捕。
func TestRefusingTheTributeGetsThouArrested(t *testing.T) {
	s := guardScene(t, britain)
	s.Talk()
	s.AnswerGuard(false)
	if s.Prompt != PromptArrest {
		t.Fatalf("拒繳之後 Prompt 是 %v,應該是 %v", s.Prompt, PromptArrest)
	}
	if s.Inventory.Gold != 1000 {
		t.Errorf("拒繳不該扣錢,剩 %d", s.Inventory.Gold)
	}
}

// ⚠ 付不出來也是逮捕 —— 「我願意但沒錢」在原版不是理由。
func TestWillingButBrokeIsStillArrested(t *testing.T) {
	s := guardScene(t, britain)
	s.Inventory.Gold = 5
	s.Talk()
	s.AnswerGuard(true)
	if s.Prompt != PromptArrest {
		t.Fatalf("錢不夠時 Prompt 是 %v,應該還是被抓", s.Prompt)
	}
	if s.Inventory.Gold != 5 {
		t.Errorf("錢不夠不該扣,剩 %d", s.Inventory.Gold)
	}
}

// 米諾克要的是一半家財,不是人頭稅。
func TestMinocTakesHalfThyGold(t *testing.T) {
	s := guardScene(t, u5data.GuardHalfGoldLocation)
	s.Talk()
	if s.Guard == nil || s.Guard.Tribute != 500 {
		t.Fatalf("1000 gp 的一半應該是 500,實得 %+v", s.Guard)
	}
	s.AnswerGuard(true)
	if s.Inventory.Gold != 500 {
		t.Errorf("付完剩 %d gp,應該是 500", s.Inventory.Gold)
	}
	// 而且不是人頭稅 —— 換個隊伍人數,金額不該變。
	s2 := guardScene(t, u5data.GuardHalfGoldLocation)
	s2.PartySize = 1
	s2.Talk()
	if s2.Guard.Tribute != 500 {
		t.Errorf("米諾克的金額不該隨人數變,實得 %d", s2.Guard.Tribute)
	}
}

// 黑棘的宮殿:沒戴徽章連問都不問。
func TestNoBadgeMeansNoQuestionsJustArrest(t *testing.T) {
	s := guardScene(t, u5data.BlackthornLocation)
	s.Talk()
	if s.Prompt == PromptGuard {
		t.Fatal("沒戴徽章不該進到問密語那一步")
	}
	// 黑棘宮殿的逮捕直接進審問。
	if s.Prompt != PromptBlackthorn {
		t.Errorf("Prompt 是 %v,黑棘宮殿的逮捕應該進審問", s.Prompt)
	}
}

// 戴了徽章就問密語;答對放行、答錯抓人。
func TestTheBadgeBuysAQuestionNotAPass(t *testing.T) {
	ask := func() *State {
		s := guardScene(t, u5data.BlackthornLocation)
		s.CombatMode = u5data.BadgeMode
		s.Talk()
		if s.Prompt != PromptGuard || s.Guard == nil || !s.Guard.Password {
			t.Fatalf("戴著徽章應該被問密語,Prompt=%v Guard=%+v", s.Prompt, s.Guard)
		}
		return s
	}
	// 答對。
	s := ask()
	s.Input = u5data.BlackthornPassword
	s.SubmitGuard()
	if s.Prompt == PromptBlackthorn {
		t.Error("答對密語不該被抓")
	}
	if !strings.Contains(lastLog(s), "過去吧") {
		t.Errorf("答對之後的訊息是 %q", lastLog(s))
	}
	// 小寫也算(比對不分大小寫)。
	s = ask()
	s.Input = strings.ToLower(u5data.BlackthornPassword)
	s.SubmitGuard()
	if s.Prompt == PromptBlackthorn {
		t.Error("密語比對應該不分大小寫")
	}
	// 答錯。
	s = ask()
	s.Input = "MONDAIN"
	s.SubmitGuard()
	if s.Prompt != PromptBlackthorn {
		t.Errorf("答錯密語後 Prompt 是 %v,應該進審問", s.Prompt)
	}
}

// ⚠ 密語是**完整相等**,不是詞首比對 —— 這與對話關鍵字是兩套規則。
//
// 混用的話「IMPERA」會被誤判成通過,而玩家永遠不知道自己其實打錯了。
func TestPasswordIsExactNotAPrefix(t *testing.T) {
	if !u5data.PasswordMatches("IMPE", "impe") {
		t.Error("不分大小寫的完整相等應該通過")
	}
	for _, wrong := range []string{"IMP", "IMPERA", "IMPE ", "", "XIMPE"} {
		if u5data.PasswordMatches("IMPE", wrong) {
			t.Errorf("%q 不該通過完整相等比對", wrong)
		}
	}
	// 對照:對話關鍵字那一套是詞首比對,兩者行為刻意不同。
	if !u5data.MatchPrefix("IMPE", "IMPERA") {
		t.Error("關鍵字比對本來就是詞首 —— 這條在對照兩套規則的差異")
	}
}

// 徽章與戰鬥模式咒語共用同一個位元組 —— 原版的共用儲存,照抄。
func TestTheBadgeSharesItsByteWithCombatModeSpells(t *testing.T) {
	s := guardScene(t, u5data.BlackthornLocation)
	s.CombatMode = u5data.BadgeMode
	// 施一個戰鬥模式咒語會蓋掉徽章。
	s.CombatMode = 'T' // An Tym
	s.Talk()
	if s.Prompt == PromptGuard {
		t.Error("咒語蓋掉徽章之後,衛兵不該再問密語 —— 應該直接抓")
	}
}

// 對話號碼 0xFE:「滾開,害蟲!」說完那個人就跑掉。
func TestBegoneVerminMakesThemFlee(t *testing.T) {
	s := guardScene(t, britain)
	n := &s.npcs[1]
	n.Creature = 0x50 // 一般居民(在 0x40..0x73 內)
	n.Dialogue = u5data.DialogueSpecialFE
	s.Talk()
	if !strings.Contains(lastLog(s), "滾開") {
		t.Errorf("訊息是 %q,應該是「滾開,害蟲!」", lastLog(s))
	}
	if n.Dialogue != u5data.DialogueFrightened {
		t.Errorf("對話號碼是 %02X,應該改成 %02X(怕你)",
			n.Dialogue, u5data.DialogueFrightened)
	}
	if n.Schedule.AI[0] != u5data.NPCAIFleeing {
		t.Errorf("行為型別是 %d,應該是 %d(逃跑)",
			n.Schedule.AI[0], u5data.NPCAIFleeing)
	}
}

// ⚠ 不是每個 0xFE 都會跑 —— 生物編號要在 0x40..0x73 才算「人」。
func TestOnlyPeopleFlee(t *testing.T) {
	s := guardScene(t, britain)
	n := &s.npcs[1]
	n.Creature = 0x20 // 動物 / 怪物
	n.Dialogue = u5data.DialogueSpecialFE
	s.Talk()
	if n.Dialogue != u5data.DialogueSpecialFE {
		t.Errorf("生物編號 0x20 不該被改成怕你,實得 %02X", n.Dialogue)
	}
}
