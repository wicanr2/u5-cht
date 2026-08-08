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

// TestOverworldKlimbNeedsTheGrapple:第一道閘門是抓鉤。
func TestOverworldKlimbNeedsTheGrapple(t *testing.T) {
	s := trollScene(t)
	s.Inventory.Grapple = 0
	s.Klimb()
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgWithWhat) {
		t.Errorf("沒抓鉤該問「用什麼爬」:%q", s.Messages)
	}
	if s.Prompt == PromptDirection {
		t.Error("沒抓鉤卻已經在問方向")
	}
	// 有抓鉤但騎著馬 → 只能徒步。
	s = trollScene(t)
	s.Inventory.Grapple = 0xFF
	s.Transport = u5data.TileHorse | 2
	s.Messages = nil
	s.Klimb()
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgOnFootOnly) {
		t.Errorf("騎著馬該說須徒步:%q", s.Messages)
	}
}

// TestCliffAndUnclimbableAreDifferentMessages:峭壁與「不能爬」是兩句話。
func TestCliffAndUnclimbableAreDifferentMessages(t *testing.T) {
	cases := []struct {
		tile byte
		want string
		move bool
	}{
		{ClimbCliff, MsgImpassable, false},
		{0x05, MsgNotClimbable, false}, // 草地
		{ClimbMountain, "", true},
	}
	for _, c := range cases {
		s := trollScene(t)
		s.Inventory.Grapple = 0xFF
		s.X, s.Y = 10, 10
		if !s.SetTileAt(11, 10, c.tile) {
			t.Skip("寫不進世界地圖")
		}
		s.Messages = nil
		s.Klimb()
		if s.Prompt != PromptDirection {
			t.Fatalf("地形 0x%02X:沒在問方向", c.tile)
		}
		s.AnswerDirection(East)
		joined := strings.Join(s.Messages, "|")
		if c.want != "" && !strings.Contains(joined, c.want) {
			t.Errorf("地形 0x%02X:少了 %q,實際 %q", c.tile, c.want, s.Messages)
		}
		if c.want == MsgImpassable && strings.Contains(joined, MsgNotClimbable) {
			t.Errorf("峭壁印成了「不能爬」:%q", s.Messages)
		}
		moved := s.X == 11
		if moved != c.move {
			t.Errorf("地形 0x%02X:移動 = %v,預期 %v", c.tile, moved, c.move)
		}
	}
}

// TestFallingDoesNotBlockTheClimb 是最容易寫錯的一條。
//
// 原版的 `sub_2D014`(移動)在摔倒迴圈**之後、無條件**執行 ——
// 「有人摔倒就不過去」是很自然的想像,但那不是原版。
func TestFallingDoesNotBlockTheClimb(t *testing.T) {
	s := trollScene(t)
	s.Inventory.Grapple = 0xFF
	s.X, s.Y = 10, 10
	if !s.SetTileAt(11, 10, ClimbMountain) {
		t.Skip("寫不進世界地圖")
	}
	// 敏捷 0 → 每個人都摔。
	for i := 0; i < s.PartySize; i++ {
		s.Roster[i].Dex = 0
		s.Roster[i].Status = u5data.StatusGood
		s.Roster[i].HP = 200
	}
	s.Messages = nil
	s.Klimb()
	s.AnswerDirection(East)
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgFell) {
		t.Errorf("敏捷 0 卻沒人摔倒:%q", s.Messages)
	}
	if s.X != 11 {
		t.Errorf("摔倒之後停在 X=%d —— 原版無條件過去", s.X)
	}
	for i := 0; i < s.PartySize; i++ {
		lost := 200 - int(s.Roster[i].HP)
		if lost < 1 || lost > ClimbFallDamageMax {
			t.Errorf("第 %d 位掉了 %d 血,超出 1..%d", i, lost, ClimbFallDamageMax)
		}
	}
}

// TestGrappleOffsetIsKnown:更正掉 `hasRope` 的陳舊 `return false`。
func TestGrappleOffsetIsKnown(t *testing.T) {
	if u5data.SaveGrappleOffset != 0x0209 {
		t.Fatalf("抓鉤的位移是 0x%04X,預期 0x0209", u5data.SaveGrappleOffset)
	}
	// 七個位元組要正好排滿 0x0209..0x020F。
	if u5data.SaveCarpetsOffset != u5data.SaveGrappleOffset+1 {
		t.Errorf("魔毯在 0x%04X,該緊接在抓鉤後面", u5data.SaveCarpetsOffset)
	}
	if u5data.SaveTorchesOffset != u5data.SaveGrappleOffset-1 {
		t.Errorf("火把在 0x%04X,該緊接在抓鉤前面", u5data.SaveTorchesOffset)
	}
	// hasRope 要真的看那個欄位,不是寫死 false。
	s := trollScene(t)
	s.Inventory.Grapple = 0
	if s.hasRope() {
		t.Error("沒抓鉤卻回報有")
	}
	s.Inventory.Grapple = 0xFF
	if !s.hasRope() {
		t.Error("有抓鉤卻回報沒有 —— `return false` 的陳舊標記又回來了")
	}
}
