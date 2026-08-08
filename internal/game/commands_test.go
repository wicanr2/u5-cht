package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// TestJimmyMagicLockAlwaysBreaksTheKey:魔法鎖必定失敗,**而且照樣扣鑰匙**。
//
// 原版 `loc_14DC0` 直接跳到扣鑰匙那段。寫成「魔法鎖不能撬,什麼都不發生」
// 會讓玩家可以無限試 —— 而原版是會把鑰匙耗光的,那是真正的代價。
func TestJimmyMagicLockAlwaysBreaksTheKey(t *testing.T) {
	for _, tile := range []byte{u5data.TileMagicLockedA, u5data.TileMagicLockedB} {
		s := lockScene(t)
		s.Inventory.Keys = 3
		s.SetTileAt(s.X+1, s.Y, tile)
		for i := 0; i < 3; i++ {
			s.jimmyAt(s.X+1, s.Y)
		}
		if s.Inventory.Keys != 0 {
			t.Errorf("0x%02X 撬三次之後還剩 %d 把鑰匙", tile, s.Inventory.Keys)
		}
		if got := s.TileAt(s.X+1, s.Y); got != tile {
			t.Errorf("魔法鎖被撬開了:0x%02X", got)
		}
	}
}

// TestJimmyOrdinaryLockRollsAgainstDex:普通鎖擲 random(0,29) 對上敏捷。
//
// 敏捷 30 一定成功(擲值上限 29)、敏捷 0 一定失敗 —— 兩端各驗一次,
// 才知道比較方向沒有寫反。寫反的話「手笨的人反而撬得開」,
// 而在隨機的表象下沒有人看得出來。
func TestJimmyOrdinaryLockRollsAgainstDex(t *testing.T) {
	for _, c := range []struct {
		dex      byte
		wantOpen bool
	}{{30, true}, {0, false}} {
		s := lockScene(t)
		s.Inventory.Keys = 1
		// ⚠ 全隊都設 —— `pickCharacter` 目前取的是最後一位能動的隊員
		// (多人選單還沒接,見 pickchar.go 的說明)。只設第 0 名的話
		// 這條測的是別人的敏捷,而它會「通過」得莫名其妙。
		for i := range s.Roster {
			s.Roster[i].Dex = c.dex
			s.Roster[i].Status = u5data.StatusGood
		}
		s.SetTileAt(s.X+1, s.Y, u5data.TileLockedDoor)
		s.jimmyAt(s.X+1, s.Y)
		opened := s.TileAt(s.X+1, s.Y) == u5data.TileDoorA
		if opened != c.wantOpen {
			t.Errorf("敏捷 %d:開了=%v,預期 %v(訊息 %q)", c.dex, opened, c.wantOpen, s.Messages)
		}
		if s.Inventory.Keys != 0 {
			t.Errorf("敏捷 %d:鑰匙沒扣", c.dex)
		}
	}
}

// TestJimmyWithoutKeysDoesNotEvenAsk:沒鑰匙就不問方向。
func TestJimmyWithoutKeysDoesNotEvenAsk(t *testing.T) {
	s := lockScene(t)
	s.Inventory.Keys = 0
	s.Jimmy()
	if s.AwaitingDirection() {
		t.Error("沒鑰匙卻還是問了方向")
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgNoKeys) {
		t.Errorf("沒印出「%s」:%q", MsgNoKeys, s.Messages)
	}
}

// TestAvatarMustLead:聖者不能離開第一位。
//
// 隊伍第 0 格是聖者,存檔格式、對話系統、結局判定都假設它在那裡。
// 少了這道檢查,玩家可以把聖者換到後面,然後上面那些全部找錯人 ——
// 而症狀會出現在很遠的地方(例如結局判定失敗),幾乎追不回來。
func TestAvatarMustLead(t *testing.T) {
	s := lockScene(t)
	if s.PartySize < 2 {
		t.Skip("隊伍不到兩人")
	}
	first := s.Roster[0].Name
	if s.NewOrder(0, 1) {
		t.Error("把聖者換到第二位竟然成功了")
	}
	if s.Roster[0].Name != first {
		t.Errorf("聖者被換走了:%q → %q", first, s.Roster[0].Name)
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgMustLead) {
		t.Errorf("沒印出「必須走在最前面」:%q", s.Messages)
	}
}

// TestNewOrderSwapsWholeRecords:交換的是整筆記錄,不是只換名字。
func TestNewOrderSwapsWholeRecords(t *testing.T) {
	s := lockScene(t)
	if s.PartySize < 3 {
		t.Skip("隊伍不到三人")
	}
	a, b := s.Roster[1], s.Roster[2]
	if !s.NewOrder(1, 2) {
		t.Fatal("交換第 2、3 位失敗")
	}
	if s.Roster[1].Name != b.Name || s.Roster[2].Name != a.Name {
		t.Errorf("名字沒換:%q %q", s.Roster[1].Name, s.Roster[2].Name)
	}
	if s.Roster[1].Strength != b.Strength || s.Roster[2].Strength != a.Strength {
		t.Error("只換了名字,力量沒跟著換 —— 交換的該是整筆記錄")
	}
}

// TestViewGemDoesNotSpendTheGem:看寶石**不扣寶石**。
//
// 原版只檢查「有沒有」(`You have none!`),看完寶石還在。
// 這很反直覺,所以特別釘住 —— 不要「順手」加一行扣除。
func TestViewGemDoesNotSpendTheGem(t *testing.T) {
	s := lockScene(t)
	s.Inventory.Gems = 2
	if !s.ViewGem() {
		t.Fatalf("有寶石卻看不了:%q", s.Messages)
	}
	if s.Inventory.Gems != 2 {
		t.Errorf("寶石被扣掉了:剩 %d", s.Inventory.Gems)
	}

	s2 := lockScene(t)
	s2.Inventory.Gems = 0
	if s2.ViewGem() {
		t.Error("沒寶石卻看得了")
	}
	if !strings.Contains(strings.Join(s2.Messages, "|"), MsgYouHaveNone) {
		t.Errorf("沒印出「%s」:%q", MsgYouHaveNone, s2.Messages)
	}
}

// TestZtatsWrapsThroughArmamentsNotThroughTheParty —— ★ 繞回的接縫在哪。
//
// ⚠⚠ 這條原本斷言「翻頁在隊伍範圍內繞回」,而那是**發明的模型**。
// 原版有 17 頁(六名 × 2 + Equipment + 四個清單頁),第 0 頁往前是
// **Armaments**(`docs/re/94`),不是最後一名。
func TestZtatsWrapsThroughArmamentsNotThroughTheParty(t *testing.T) {
	s := lockScene(t)
	if !s.BeginZtats() {
		t.Fatal("打不開數值畫面")
	}
	if len(s.ZtatsLines()) == 0 {
		t.Error("數值畫面是空的")
	}
	s.ZtatsPage(-1)
	if s.Zstats.Page != ZtatsArmamentsPage {
		t.Errorf("從第 0 頁往前到第 %d 頁,原版是 Armaments(%d)",
			s.Zstats.Page, ZtatsArmamentsPage)
	}
	s.ZtatsPage(1)
	if s.Zstats.Page != 0 {
		t.Errorf("再往後該繞回第 0 頁,實際 %d", s.Zstats.Page)
	}
	s.EndZtats()
	if s.Prompt != PromptNone {
		t.Errorf("收起之後還卡在 Prompt %v", s.Prompt)
	}
}

func lockScene(t *testing.T) *State {
	t.Helper()
	s := newCreateState(t) // 借它的「載入原版存檔 + 固定亂數種子」
	s.MaxMessages = 16
	return s
}
