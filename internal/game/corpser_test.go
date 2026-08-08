package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// TestCorpserIndexIsFortyFive 釘住那個編號。
//
// 0x2D = 45 對回怪物名表 `off_3F564` 的第 45 筆(0x3F618)= Corpser。
// 這條擋的是「把 0x2D 當成別的東西」—— 同一個位元組在物品命名空間裡
// 是 `ItemAmuletOfTurning`,兩者無關(本專案踩過好幾次同值不同命名空間)。
func TestCorpserIndexIsFortyFive(t *testing.T) {
	if u5data.CreatureCorpserIdx != 45 {
		t.Fatalf("拖屍怪是 %d,預期 45", u5data.CreatureCorpserIdx)
	}
	if u5data.CreatureCorpserIdx == u5data.CreatureGazerIdx {
		t.Error("拖屍怪與注視者不該是同一個編號")
	}
}

// TestCorpserEscapeThresholdIsAtLeastOne:`sar eax,1` 之後原版硬補一個 1。
func TestCorpserEscapeThresholdIsAtLeastOne(t *testing.T) {
	for roll := 0; roll <= u5data.CorpserEscapeRollMax; roll++ {
		got := u5data.CorpserEscapeThreshold(roll)
		want := roll / 2
		if want < 1 {
			want = 1
		}
		if got != want {
			t.Fatalf("骰 %d → 門檻 %d,預期 %d", roll, got, want)
		}
		if got < 1 {
			t.Fatalf("骰 %d 的門檻是 %d —— 原版保證至少 1", roll, got)
		}
	}
	// 門檻最大是 30,所以敏捷 31 以上一定第一回合就掙脫。
	if u5data.CorpserEscapeThreshold(u5data.CorpserEscapeRollMax) != 30 {
		t.Errorf("最大門檻是 %d,預期 30",
			u5data.CorpserEscapeThreshold(u5data.CorpserEscapeRollMax))
	}
}

// TestOnlyPartyMembersGetDraggedUnder:原版那個分支的前提是「目標是隊員」。
func TestOnlyPartyMembersGetDraggedUnder(t *testing.T) {
	s := corpserArena(t)
	// 找一個隊員與一隻怪物。
	party, monster := -1, -1
	for i := range s.Combat.Units {
		u := &s.Combat.Units[i]
		if u.Flags == 0 {
			continue
		}
		if u.IsParty() && party < 0 {
			party = i
		}
		if !u.IsParty() && monster < 0 {
			monster = i
		}
	}
	if party < 0 || monster < 0 {
		t.Skip("戰場上湊不出一個隊員與一隻怪物")
	}
	// 攻擊者不是拖屍怪 → 不抓。
	s.Combat.Units[monster].Creature = u5data.CreatureGazerIdx
	if s.corpserGrab(monster, party) {
		t.Error("注視者不該把人拖下水")
	}
	// 是拖屍怪、目標是隊員 → 抓。
	s.Combat.Units[monster].Creature = u5data.CreatureCorpserIdx
	s.Messages = nil
	if !s.corpserGrab(monster, party) {
		t.Fatal("拖屍怪打隊員該把人拖下水")
	}
	if s.Combat.Units[party].Flags&UnitGrabbed == 0 {
		t.Error("旗標沒設上")
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgDraggedUnder) {
		t.Errorf("沒印「被拖入水中」:%q", s.Messages)
	}
	// 已經被抓住了就不重複抓。
	if s.corpserGrab(monster, party) {
		t.Error("已經在水下卻又被抓一次")
	}
	// ★ 目標是怪物 → 只是普通命中,不抓。
	other := -1
	for i := range s.Combat.Units {
		if i != monster && s.Combat.Units[i].Flags != 0 && !s.Combat.Units[i].IsParty() {
			other = i
			break
		}
	}
	if other >= 0 && s.corpserGrab(monster, other) {
		t.Error("拖屍怪打怪物不該觸發拖入水中")
	}
}

// TestStrugglingUnderwaterAlwaysCostsTheTurnAndArghs 釘住「不論掙不掙脫都印 ARGH!」。
func TestStrugglingUnderwaterAlwaysCostsTheTurnAndArghs(t *testing.T) {
	s := corpserArena(t)
	party := -1
	for i := range s.Combat.Units {
		if s.Combat.Units[i].Flags != 0 && s.Combat.Units[i].IsParty() {
			party = i
			break
		}
	}
	if party < 0 {
		t.Skip("戰場上沒有隊員")
	}
	u := &s.Combat.Units[party]

	// 敏捷 31 以上一定掙脫得了(門檻最大 30)。
	if ch := s.charOf(u); ch != nil {
		ch.Dex = 99
	}
	u.Flags |= UnitGrabbed
	s.Messages = nil
	s.strugglingUnderwater(u)
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgArgh) {
		t.Errorf("沒印「啊啊啊」:%q", s.Messages)
	}
	if u.Flags&UnitGrabbed != 0 {
		t.Error("敏捷 99 該一定掙脫")
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgRegurgitated) {
		t.Errorf("沒印「被吐了出來」:%q", s.Messages)
	}

	// 敏捷 1 幾乎不可能掙脫,但 ARGH! 一定要印。
	if ch := s.charOf(u); ch != nil {
		ch.Dex = 1
	}
	stuck := 0
	for i := 0; i < 60; i++ {
		u.Flags |= UnitGrabbed
		s.Messages = nil
		s.strugglingUnderwater(u)
		if !strings.Contains(strings.Join(s.Messages, "|"), MsgArgh) {
			t.Fatalf("第 %d 次沒印「啊啊啊」:%q", i, s.Messages)
		}
		if u.Flags&UnitGrabbed != 0 {
			stuck++
		}
	}
	if stuck == 0 {
		t.Error("敏捷 1 掙脫了 60 次 —— 門檻至少是 1,不該次次成功")
	}
}

// TestSleepingPartyMemberAlwaysSaysZzzzz 是這次補的另一半。
//
// 原版 `sub_A360`(隊員)與 `sub_A108`(怪物)是**兩支不同的函式**,
// 醒來機率不同,而隊員那條不論醒不醒都印 "Zzzzz..." 並用掉回合。
func TestSleepingPartyMemberAlwaysSaysZzzzz(t *testing.T) {
	s := corpserArena(t)
	party := -1
	for i := range s.Combat.Units {
		if s.Combat.Units[i].Flags != 0 && s.Combat.Units[i].IsParty() {
			party = i
			break
		}
	}
	if party < 0 {
		t.Skip("戰場上沒有隊員")
	}
	// 讓所有人都睡著,這樣一定會走到那條路。
	for i := range s.Combat.Units {
		if s.Combat.Units[i].Flags != 0 {
			s.Combat.Units[i].Flags |= UnitAsleep
		}
	}
	s.Messages = nil
	// 推進排程幾次,睡著的隊員應該唸出 Zzzzz。
	for i := 0; i < 40 && s.Combat != nil; i++ {
		s.advanceCombat()
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgZzzzz) {
		t.Errorf("睡著的隊員沒唸「呼……呼……」:%q", s.Messages)
	}
}

// corpserArena 借地牢遊蕩怪物那條路開一場戰鬥 —— 只要場上有隊員與怪物就夠了。
func corpserArena(t *testing.T) *State {
	t.Helper()
	s := dungeonState(t)
	s.MaxMessages = 64
	s.Location = 0
	e := u5data.DungeonEntrances[0]
	s.X, s.Y = e.X, e.Y
	s.Enter()
	d := s.Dungeon
	d.X, d.Y = 4, 4
	d.Facing = North
	d.Monster = &DungeonMonster{
		Kind: 0, Creature: u5data.DungeonMonsterCreature(0),
		X: 4, Y: 3, PrevX: 4, PrevY: 3,
	}
	s.dungeonMonsterAttacks()
	if !s.InCombat() {
		t.Skip("開不了戰場")
	}
	s.Messages = nil
	return s
}
