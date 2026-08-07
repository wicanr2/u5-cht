package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// TestRegaliaShareOneModeByte:護符與王冠共用全域模式位元組。
//
// ★ 與四個咒語是**同一個位元組**(`byte_3E08A`)。所以戴上王冠會蓋掉
// In Sanct、而 An Tym 也會蓋掉王冠。寫成兩個獨立的布林值,行為就與原版不同 ——
// 而那種差異在遊玩時只覺得「怪」,追不到源頭。
func TestRegaliaShareOneModeByte(t *testing.T) {
	s := useScene(t)
	s.Regalia.Amulet, s.Regalia.Crown = true, true

	s.Use(UseAmulet)
	if s.CombatMode != ModeAmulet {
		t.Errorf("戴護符之後模式是 0x%02X,預期 0x%02X", s.CombatMode, ModeAmulet)
	}
	// 戴王冠會蓋掉護符 —— 一個位元組放不下兩件。
	s.Use(UseCrown)
	if s.CombatMode != ModeCrown {
		t.Errorf("戴王冠之後模式是 0x%02X,預期 0x%02X", s.CombatMode, ModeCrown)
	}
	// 再用一次就取下。
	s.Messages = nil
	s.Use(UseCrown)
	if s.CombatMode != 0 {
		t.Errorf("取下之後模式還是 0x%02X", s.CombatMode)
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgRemoved) {
		t.Errorf("沒印出「取下了」:%q", s.Messages)
	}
}

// TestSceptreSetsNoMode:權杖只化力場,**不設模式**。
//
// 護符與王冠都寫模式位元組,權杖不寫 —— 原版那一支只放三個音效然後化力場。
// 寫成「跟王冠一樣」會多出一個原版沒有的持續效果。
func TestSceptreSetsNoMode(t *testing.T) {
	s := useScene(t)
	s.Regalia.Sceptre = true
	s.CombatMode = 0
	s.Use(UseSceptre)
	if s.CombatMode != 0 {
		t.Errorf("權杖設了模式 0x%02X —— 原版不會", s.CombatMode)
	}
}

// TestUseNeedsTheItem:身上沒有就用不了。
func TestUseNeedsTheItem(t *testing.T) {
	s := useScene(t)
	s.Regalia = u5data.Regalia{}
	s.Inventory.Carpets = 0
	s.Inventory.OddKeys = 0
	for _, item := range []int{UseAmulet, UseCrown, UseSceptre, UseCarpet, UseSkullKey, UsePlans} {
		s.Messages = nil
		if s.Use(item) {
			t.Errorf("道具 %d 身上沒有卻用得了", item)
		}
	}
}

// TestSextantAndSpyglassNeedNightOutdoors:六分儀與望遠鏡都要戶外 + 夜裡。
//
// 兩道前置各有專屬訊息,而且**順序有意義**:原版先擋室內、再擋白天。
// 反過來的話在城裡的白天會說「只有夜裡才行」,而正確的回答是「只有戶外才行」。
func TestSextantAndSpyglassNeedNightOutdoors(t *testing.T) {
	s := useScene(t)
	s.Location = 1 // 在場景裡
	s.Clock.Hour = 2
	s.Messages = nil
	if s.useSextant() {
		t.Error("在室內卻測得出位置")
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgOnlyOutdoors) {
		t.Errorf("室內該說「只有在戶外才行」:%q", s.Messages)
	}

	s.Location = 0
	s.Clock.Hour = 12
	s.Messages = nil
	if s.useSextant() {
		t.Error("白天卻測得出位置")
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgOnlyAtNight) {
		t.Errorf("白天該說「只有夜裡才行」:%q", s.Messages)
	}

	s.Clock.Hour = 2
	s.Messages = nil
	if !s.useSextant() {
		t.Errorf("戶外的夜裡該測得出位置:%q", s.Messages)
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgPosition) {
		t.Errorf("沒報出位置:%q", s.Messages)
	}
}

// TestPlansOnlyOnShipboard:圖紙只在船上有用,而且不會被用掉。
func TestPlansOnlyOnShipboard(t *testing.T) {
	s := useScene(t)
	s.Regalia.Plans = true
	s.Transport = u5data.VehicleWalk
	if s.usePlans() {
		t.Error("走路時也改裝得了船")
	}
	s.Transport = u5data.VehicleShip
	if !s.usePlans() {
		t.Errorf("在船上卻改裝不了:%q", s.Messages)
	}
	if !s.ShipRigged {
		t.Error("船速沒加倍")
	}
	if !s.Regalia.Plans {
		t.Error("圖紙被用掉了 —— 原版不會扣")
	}
}

// TestWatchReadsLikeTheGrandfatherClock:懷錶與老爺鐘同一個格式。
//
// 兩處若各寫一套,遲早會有一處的 12 時 / 0 時邊界不一樣。
func TestWatchReadsLikeTheGrandfatherClock(t *testing.T) {
	s := useScene(t)
	s.Clock.Hour, s.Clock.Minute = 0, 5
	s.Messages = nil
	s.useWatch()
	if !strings.Contains(strings.Join(s.Messages, "|"), "12:05") {
		t.Errorf("0 時 5 分該印 12:05:%q", s.Messages)
	}
}

func useScene(t *testing.T) *State {
	t.Helper()
	s := newCreateState(t)
	s.MaxMessages = 16
	s.Messages = nil
	return s
}
