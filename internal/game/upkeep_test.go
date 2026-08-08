package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

func upkeepScene(t *testing.T) *State {
	t.Helper()
	s := newCreateState(t)
	s.MaxMessages = 32
	s.mealHour = newUpkeepHour
	if s.PartySize < 1 {
		s.PartySize = 1
	}
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		s.Roster[i].Status = u5data.StatusGood
		s.Roster[i].HP = 200
	}
	s.Inventory.Food = 100
	s.Messages = nil
	return s
}

// TestPoisonCostsOneHitPointPerTurn:中毒每回合掉 1 血。
//
// ⚠ 引擎在此之前**中毒完全不扣血** —— 狀態掛著卻沒有代價。
func TestPoisonCostsOneHitPointPerTurn(t *testing.T) {
	s := upkeepScene(t)
	s.Roster[0].Status = u5data.StatusPoisoned
	s.Clock.Hour = 3 // 不是用餐時刻
	for i := 0; i < 10; i++ {
		s.upkeep()
	}
	if got := 200 - int(s.Roster[0].HP); got != 10*PoisonDamagePerTurn {
		t.Errorf("中毒 10 回合掉了 %d 血,預期 %d", got, 10*PoisonDamagePerTurn)
	}
	// 沒中毒的人不該掉血。
	if s.PartySize > 1 && s.Roster[1].HP != 200 {
		t.Errorf("沒中毒的隊員掉了 %d 血", 200-s.Roster[1].HP)
	}
}

// TestFoodIsEatenAtSixTwelveAndEighteen 釘住那三個時刻。
func TestFoodIsEatenAtSixTwelveAndEighteen(t *testing.T) {
	for hour := 0; hour < 24; hour++ {
		s := upkeepScene(t)
		s.Clock.Hour = hour
		s.upkeep()
		eaten := 100 - s.Inventory.Food
		meal := hour == 6 || hour == 12 || hour == 18
		if meal && eaten != s.PartySize {
			t.Errorf("%d 點吃了 %d 份,預期 %d 份(隊伍人數)", hour, eaten, s.PartySize)
		}
		if !meal && eaten != 0 {
			t.Errorf("%d 點竟然吃了 %d 份 —— 只有 6 / 12 / 18 點才扣糧", hour, eaten)
		}
	}
}

// TestSameHourOnlyEatsOnce:同一個小時走幾百步也只扣一次糧。
func TestSameHourOnlyEatsOnce(t *testing.T) {
	s := upkeepScene(t)
	s.Clock.Hour = 12
	for i := 0; i < 200; i++ {
		s.upkeep()
	}
	if eaten := 100 - s.Inventory.Food; eaten != s.PartySize {
		t.Errorf("同一小時走 200 步吃了 %d 份,預期 %d 份", eaten, s.PartySize)
	}
}

// TestSleepingAndDeadMembersDoNotEat:睡著的人與死人都不吃。
func TestSleepingAndDeadMembersDoNotEat(t *testing.T) {
	s := upkeepScene(t)
	if s.PartySize < 2 {
		t.Skip("隊伍太小")
	}
	s.Roster[0].Status = u5data.StatusAsleep
	s.Roster[1].Status = u5data.StatusDead
	s.Clock.Hour = 6
	s.upkeep()
	want := s.PartySize - 2
	if eaten := 100 - s.Inventory.Food; eaten != want {
		t.Errorf("吃了 %d 份,預期 %d 份(睡著的與死人都不吃)", eaten, want)
	}
}

// TestStarvingHurtsEveryTurn:斷糧是**每回合**掉血,不是一天三次。
func TestStarvingHurtsEveryTurn(t *testing.T) {
	s := upkeepScene(t)
	s.Inventory.Food = 0
	s.Clock.Hour = 3 // 不是用餐時刻,照樣要餓
	s.upkeep()
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgStarving) {
		t.Fatalf("存糧 0 卻沒挨餓:%q", s.Messages)
	}
	hurt := 0
	for i := 0; i < s.PartySize; i++ {
		if s.Roster[i].HP < 200 {
			hurt++
		}
	}
	if hurt != s.PartySize {
		t.Errorf("只有 %d 位掉血,預期全隊 %d 位", hurt, s.PartySize)
	}
	// 再走一步還要再餓一次(同一個小時內也算)。
	s.Messages = nil
	s.upkeep()
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgStarving) {
		t.Errorf("第二回合沒再挨餓 —— 原版不更新 byte_3E090,所以每回合都判:%q", s.Messages)
	}
}

// TestOnlyTwoRingsVanish 是這一條最容易寫錯的地方。
//
// 原版只比 0x2A(隱形)與 0x2C(再生)兩個值 —— **防護戒指(0x2B)不會消失**。
func TestOnlyTwoRingsVanish(t *testing.T) {
	cases := []struct {
		ring    byte
		vanishe bool
		name    string
	}{
		{u5data.ItemRingInvisibility, true, "隱形戒指"},
		{u5data.ItemRingProtection, false, "★ 防護戒指 —— 原版不比這個值"},
		{u5data.ItemRingRegeneration, true, "再生戒指"},
		{u5data.ItemNone, false, "沒戴戒指"},
	}
	// ⚠ **同一個 State 重複擲**,不要每次都 `upkeepScene(t)`:
	// 新 State 的骰子種子是固定的 1(見 `SeedRandom` 的說明),
	// 每次重建就永遠拿到同一個骰值 —— 測試會假紅。
	s := upkeepScene(t)
	for _, c := range cases {
		gone := false
		for try := 0; try < 400 && !gone; try++ {
			s.Roster[0].Raw[u5data.CharRing] = c.ring
			s.vanishRings()
			if s.Roster[0].Raw[u5data.CharRing] != c.ring {
				gone = true
			}
		}
		if gone != c.vanishe {
			t.Errorf("%s(0x%02X):會消失 = %v,預期 %v", c.name, c.ring, gone, c.vanishe)
		}
	}
}

// TestRingVanishRollIsOneInSixteen:機率是 1/16,而判準是 `== 11` 不是 `< 1`。
func TestRingVanishRollIsOneInSixteen(t *testing.T) {
	if RingVanishRollMax != 15 || RingVanishHit != 11 {
		t.Fatalf("判準是 rand(0, %d) == %d,預期 rand(0, 15) == 11",
			RingVanishRollMax, RingVanishHit)
	}
	// 骰的範圍是 0..15 共 16 個值,命中一個 → 1/16。
	if RingVanishHit < 0 || RingVanishHit > RingVanishRollMax {
		t.Error("命中值落在骰的範圍外 —— 那樣戒指永遠不會消失")
	}
	s := upkeepScene(t)
	gone, tries := 0, 4000
	for i := 0; i < tries; i++ {
		s.Roster[0].Raw[u5data.CharRing] = u5data.ItemRingInvisibility
		s.vanishRings()
		if s.Roster[0].Raw[u5data.CharRing] == u5data.ItemNone {
			gone++
		}
	}
	// 1/16 ≈ 250/4000。放寬到 150..400 只是擋「機率差一個數量級」。
	if gone < 150 || gone > 400 {
		t.Errorf("4000 次消失了 %d 次,1/16 該落在 150..400", gone)
	}
}
