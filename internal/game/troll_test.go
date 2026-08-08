package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// trollScene 借 `combatState` —— 過路費那條路要能真的開打,
// 而 `newCreateState` 沒有載戰場資料(`CombatMaps` 是 nil)。
func trollScene(t *testing.T) *State {
	t.Helper()
	s := combatState(t)
	s.MaxMessages = 64
	s.Location = 0
	s.Transport = u5data.VehicleWalk
	s.X, s.Y = 10, 10
	s.Inventory.Gold = 500
	if s.PartySize < 1 {
		s.PartySize = 1
	}
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		s.Roster[i].Status = u5data.StatusGood
		if s.Roster[i].Name == "" {
			s.Roster[i].Name = "測試者"
		}
	}
	s.Messages = nil
	return s
}

// TestOnlyFootTravellersMeetTheTrolls:原版比的是**單一值** 0x1C。
//
// 這一條擋兩件事:(1) 忘了判載具 → 騎馬渡橋也會被抓;
// (2) 用 `IsOnFoot`(0x1C 或 0x1D)代替 —— 原版 `cmp byte_3E08C, 1Ch` 只認一個值。
func TestOnlyFootTravellersMeetTheTrolls(t *testing.T) {
	for _, tr := range []byte{
		u5data.TileHorse | 2, u5data.VehicleCarpet, u5data.VehicleSkiff,
		u5data.VehicleShip, u5data.VehicleWalk + 1,
	} {
		s := trollScene(t)
		s.Transport = tr
		// 擲一百次,只要有一次觸發就是錯的。
		for i := 0; i < 100; i++ {
			s.Messages = nil
			s.crossBridge()
			if strings.Contains(strings.Join(s.Messages, "|"), MsgTrollsUnderBridge) {
				t.Fatalf("載具 0x%02X 竟然遇到了食人妖", tr)
			}
		}
	}
}

// TestSneakUsesDexterityAndTollUsesStrength 是最容易看錯的一組。
//
// `byte_3DDC1`(偏移 0x0D)是**敏捷**、`byte_3DDC0`(0x0C)是**力量** ——
// 兩個相鄰欄位,一個管偷渡、一個管過路費。
func TestSneakUsesDexterityAndTollUsesStrength(t *testing.T) {
	// 敏捷 99 → 一定溜過去(骰上限 30)。
	s := trollScene(t)
	for i := 0; i < s.PartySize; i++ {
		s.Roster[i].Dex = 99
	}
	got := false
	for i := 0; i < 200 && !got; i++ {
		s.Messages = nil
		s.crossBridge()
		joined := strings.Join(s.Messages, "|")
		if strings.Contains(joined, MsgCaught) {
			t.Fatalf("敏捷 99 卻被抓到了:%q", s.Messages)
		}
		if strings.Contains(joined, MsgTrollsEvaded) {
			got = true
		}
	}
	if !got {
		t.Error("擲 200 次都沒觸發過遭遇 —— 1/8 的機率不該這樣")
	}

	// 敏捷 0 → 一定被抓,而過路費看的是**力量**。
	s = trollScene(t)
	for i := 0; i < s.PartySize; i++ {
		s.Roster[i].Dex = 0
		s.Roster[i].Strength = 10
	}
	for i := 0; i < 200; i++ {
		s.Messages = nil
		s.Prompt = PromptNone
		s.crossBridge()
		if strings.Contains(strings.Join(s.Messages, "|"), MsgCaught) {
			want := TrollTollBase - 10*TrollTollPerStrength // 99 − 30 = 69
			if !strings.Contains(strings.Join(s.Messages, "|"), "69") {
				t.Errorf("過路費不是 %d:%q", want, s.Messages)
			}
			return
		}
	}
	t.Error("敏捷 0 擲 200 次都沒被抓到")
}

// TestTollFormulaHasNoFloor:力量 34 以上通行費是負數 —— 照原版保留。
func TestTollFormulaHasNoFloor(t *testing.T) {
	for _, str := range []int{0, 10, 33, 34, 50} {
		toll := TrollTollBase - str*TrollTollPerStrength
		if str <= 33 && toll < 0 {
			t.Errorf("力量 %d 的通行費 %d 不該是負的", str, toll)
		}
		if str >= 34 && toll >= 0 {
			t.Errorf("力量 %d 的通行費是 %d —— 原版的算式沒有下限,該變負數", str, toll)
		}
	}
}

// TestRefusingToPayStartsAFight:不付就開打,而且怪物是食人妖。
func TestRefusingToPayStartsAFight(t *testing.T) {
	s := trollScene(t)
	before := s.Inventory.Gold
	s.trollToll()
	if s.Prompt != PromptYesNo {
		t.Fatalf("沒在問「汝要付嗎」,Prompt=%v", s.Prompt)
	}
	s.AnswerYesNo(false)
	if s.Inventory.Gold != before {
		t.Errorf("不付卻扣了錢:%d → %d", before, s.Inventory.Gold)
	}
	if !s.InCombat() {
		t.Error("不付卻沒開打")
	}
}

// TestPayingTheTollDeductsGold / 錢不夠就開打並退回。
func TestPayingTheTollDeductsGold(t *testing.T) {
	s := trollScene(t)
	for i := 0; i < s.PartySize; i++ {
		s.Roster[i].Strength = 10
	}
	toll := TrollTollBase - 10*TrollTollPerStrength
	s.Inventory.Gold = toll + 5
	s.trollToll()
	s.AnswerYesNo(true)
	if s.Inventory.Gold != 5 {
		t.Errorf("付完剩 %d,預期 5", s.Inventory.Gold)
	}
	if s.InCombat() {
		t.Error("付了錢卻還開打")
	}

	// ★ 錢不夠:退回原數並開打。
	s = trollScene(t)
	for i := 0; i < s.PartySize; i++ {
		s.Roster[i].Strength = 10
	}
	s.Inventory.Gold = toll - 1
	s.trollToll()
	s.AnswerYesNo(true)
	if s.Inventory.Gold != toll-1 {
		t.Errorf("錢不夠卻扣了錢:剩 %d,預期 %d", s.Inventory.Gold, toll-1)
	}
	if !s.InCombat() {
		t.Error("錢不夠卻沒開打")
	}
}

// TestSleepingAndDeadMembersDoNotSneak:死的與睡著的不參加擲骰。
func TestSleepingAndDeadMembersDoNotSneak(t *testing.T) {
	s := trollScene(t)
	if s.PartySize < 2 {
		t.Skip("隊伍太小")
	}
	// 第一位睡著、其餘敏捷 99 → 只有醒著的人會被點名。
	s.Roster[0].Status = u5data.StatusAsleep
	for i := 0; i < s.PartySize; i++ {
		s.Roster[i].Dex = 99
	}
	for i := 0; i < 300; i++ {
		s.Messages = nil
		s.crossBridge()
		joined := strings.Join(s.Messages, "|")
		if !strings.Contains(joined, MsgTrollsUnderBridge) {
			continue
		}
		if strings.Contains(joined, s.Roster[0].Name+MsgSneaksAcross) {
			t.Errorf("睡著的 %s 也被點名潛行:%q", s.Roster[0].Name, s.Messages)
		}
		return
	}
	t.Skip("300 次都沒觸發遭遇")
}
