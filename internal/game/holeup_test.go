package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 紮營能回血 —— 但**要睡滿六小時以上**,而且回的是 random(1,63) 不是回滿。
func TestCampRestoresTheParty(t *testing.T) {
	s := dungeonState(t)
	s.Location, s.Floor = 0, 0
	s.Transport = u5data.VehicleWalk
	s.X, s.Y = 82, 106 // 不列顛城旁的陸地
	c := &s.Roster[0]
	c.HP = 1
	c.MP = 0
	s.Clock.Hour, s.Clock.Minute = 10, 0

	s.HoleUp()
	if !s.AwaitingNumber() {
		t.Fatalf("紮營沒有問時數:\n%s", s.log())
	}
	s.AnswerNumber(9)
	if s.AwaitingYesNo() {
		s.AnswerYesNo(false)
	}
	if s.InCombat() {
		t.Skip("這一次被突襲了,換一顆種子再測恢復")
	}
	if c.HP <= 1 {
		t.Errorf("睡了九小時 HP 還是 %d:\n%s", c.HP, s.log())
	}
	for _, m := range s.Party() {
		if m.Status == u5data.StatusAsleep {
			t.Errorf("%s 醒不過來", m.Name)
		}
	}
}

// ★ 睡不到六小時**完全沒效果**(原版 `cmp arg_8, 5; jle → "No effect..."`)。
//
// 我第一版把旅店那支恢復拿來用,等於「睡三小時就回滿血」——
// 那讓紮營變成無限回血機,而原版刻意設了門檻與冷卻。
func TestShortCampDoesNothing(t *testing.T) {
	s := dungeonState(t)
	s.Location, s.Floor = 0, 0
	s.Transport = u5data.VehicleWalk
	s.X, s.Y = 82, 106
	c := &s.Roster[0]
	c.HP = 1
	s.Clock.Hour, s.Clock.Minute = 10, 0

	s.HoleUp()
	s.AnswerNumber(3) // 只睡三小時
	if s.AwaitingYesNo() {
		s.AnswerYesNo(false)
	}
	if s.InCombat() {
		t.Skip("被突襲了")
	}
	if c.HP != 1 {
		t.Errorf("睡三小時竟然回了血:%d", c.HP)
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), "毫無效果") {
		t.Errorf("沒說毫無效果:\n%s", s.log())
	}
}

// 冷卻沒退完再紮營也是「毫無效果」—— 擋住連續紮營刷血。
func TestCampCooldownBlocksRepeats(t *testing.T) {
	s := dungeonState(t)
	s.Location, s.Floor = 0, 0
	s.Transport = u5data.VehicleWalk
	s.X, s.Y = 82, 106
	s.Clock.Hour = 10
	s.RestCooldown = campRestCooldown
	c := &s.Roster[0]
	c.HP = 1

	s.HoleUp()
	s.AnswerNumber(9)
	if s.AwaitingYesNo() {
		s.AnswerYesNo(false)
	}
	if s.InCombat() {
		t.Skip("被突襲了")
	}
	if c.HP != 1 {
		t.Errorf("冷卻期間竟然回了血:%d", c.HP)
	}
}

// 守夜的人什麼都不恢復 —— 那是派人守夜要付的代價。
func TestTheWatchGetsNoRest(t *testing.T) {
	s := dungeonState(t)
	s.Location, s.Floor = 0, 0
	s.Transport = u5data.VehicleWalk
	s.X, s.Y = 82, 106
	s.Clock.Hour = 10
	for i := 0; i < s.PartySize; i++ {
		s.Roster[i].HP = 1
		s.Roster[i].Status = u5data.StatusGood
	}
	// 直接呼叫 camp,跳過選單(選單的路徑另有測試)。
	s.camp(9, 0)
	if s.InCombat() {
		t.Skip("被突襲了")
	}
	if s.Roster[0].HP != 1 {
		t.Errorf("守夜的人竟然回了血:%d", s.Roster[0].HP)
	}
	if s.PartySize > 1 && s.Roster[1].HP <= 1 {
		t.Errorf("沒守夜的人反而沒回血:%d", s.Roster[1].HP)
	}
}

// ★★ 睡床**什麼都不恢復**,只是讓時間過去,而且起床會往東挪一格。
//
// 憑「睡在床上當然比野地舒服」的直覺去補恢復,就是自創遊戲。
func TestSleepingInABedHealsNothing(t *testing.T) {
	s := dungeonState(t)
	s.Location, s.Floor = 0, 0
	s.Clock.Hour = 10
	c := &s.Roster[0]
	c.HP, c.MP = 1, 0
	x := s.X

	s.sleepInBed(8)
	if c.HP != 1 {
		t.Errorf("睡床竟然回了血:%d", c.HP)
	}
	if c.MP != 0 {
		t.Errorf("睡床竟然回了法力:%d", c.MP)
	}
	if s.Clock.Hour != 18 {
		t.Errorf("10 時睡 8 小時醒在 %d 時,預期 18 時", s.Clock.Hour)
	}
	if s.X != x+1 {
		t.Errorf("起床沒有從床上挪開:x %d → %d", x, s.X)
	}
}

// ★ 跨午夜時原版會**多醒一小時** —— `sub edi, 17h` 減的是 23 不是 24。
//
// 這是原版的 bug,而 CLAUDE.md §3.0 要求機制與原版一模一樣、包括它的 bug。
// 「順手修好」會讓時間對不上原版,而那種差異只有並排跑才看得出來。
func TestSleepPastMidnightKeepsTheOriginalOffByOne(t *testing.T) {
	s := dungeonState(t)
	s.Location, s.Floor = 0, 0
	s.Transport = u5data.VehicleWalk
	s.X, s.Y = 82, 106
	s.Clock.Hour, s.Clock.Minute = 22, 0

	s.sleepInBed(4) // 22 + 4 = 26 → 26 − 23 = 3(正確的環繞會是 2)
	if s.Clock.Hour != 3 {
		t.Errorf("22 時睡 4 小時醒在 %d 時;原版的算法會給 3 時(減 23 而非 24)", s.Clock.Hour)
	}

	// ★ 對照:**紮營那條路減的是 24**,環繞是對的。
	// 同一個遊戲兩條休息路徑,一條對一條錯 —— 所以那是真的寫錯,不是刻意設計。
	s2 := dungeonState(t)
	s2.Location, s2.Floor = 0, 0
	s2.Transport = u5data.VehicleWalk
	s2.X, s2.Y = 82, 106
	s2.Clock.Hour, s2.Clock.Minute = 22, 0
	s2.camp(4, -1)
	if s2.InCombat() {
		t.Skip("被突襲了")
	}
	if s2.Clock.Hour != 2 {
		t.Errorf("紮營:22 時睡 4 小時醒在 %d 時,預期 2 時(減 24)", s2.Clock.Hour)
	}
}

// 城裡沒躺在床上就睡不著。
func TestHoleUpInTownNeedsABed(t *testing.T) {
	s := dungeonState(t)
	loc := &u5data.Locations[1] // 不列顛城
	s.Location, s.Floor = 0, 0
	s.X, s.Y = loc.X, loc.Y
	s.Transport = u5data.VehicleWalk
	s.Enter()
	if !s.InScene() {
		t.Skip("進不了城")
	}
	s.Messages = nil
	s.HoleUp()
	if s.AwaitingNumber() {
		t.Fatalf("沒躺在床上卻問了時數:\n%s", s.log())
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), "床") {
		t.Errorf("沒說要在床上:\n%s", s.log())
	}
	// 腳下換成床就問得出時數了。
	s.SetTileAt(s.X, s.Y, HoleUpBedTile)
	s.HoleUp()
	if !s.AwaitingNumber() {
		t.Errorf("站在床上還是睡不著:\n%s", s.log())
	}
}

// 在船上按 H 是修船,不是紮營;而且揚著帆修不了。
func TestHoleUpOnAShipRepairsIt(t *testing.T) {
	s := dungeonState(t)
	s.Location, s.Floor = 0, 0
	s.Transport = u5data.VehicleSailing
	s.ShipHull = 3
	s.Messages = nil
	s.HoleUp()
	if !strings.Contains(strings.Join(s.Messages, "|"), "收帆") {
		t.Errorf("揚著帆竟然修得起來:\n%s", s.log())
	}
	if s.ShipHull != 3 {
		t.Errorf("耐久被改了:%d", s.ShipHull)
	}

	// 收帆之後修得動,而且**一定會修到 10 以上**(原版是 do-while)。
	s.Transport = u5data.VehicleShip
	s.HoleUp()
	if s.ShipHull < shipRepairUntil {
		t.Errorf("修完耐久只有 %d,原版的 do-while 會一路加到 %d 以上",
			s.ShipHull, shipRepairUntil)
	}
	if s.ShipHull > ShipHullMax {
		t.Errorf("耐久 %d 超過上限 %d", s.ShipHull, ShipHullMax)
	}
	// 修船不該問時數 —— 它與睡覺是兩條路。
	if s.AwaitingNumber() {
		t.Error("修船竟然問了時數")
	}
}

// 騎著馬紮不了營(原版 `cmp byte…, 1Ch` 只認步行)。
func TestCampNeedsToBeOnFoot(t *testing.T) {
	s := dungeonState(t)
	s.Location, s.Floor = 0, 0
	s.X, s.Y = 82, 106
	s.Transport = u5data.TileHorse
	s.Messages = nil
	s.HoleUp()
	if s.AwaitingNumber() {
		t.Fatalf("騎著馬竟然紮起營來:\n%s", s.log())
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), "步行") {
		t.Errorf("沒說要步行:\n%s", s.log())
	}
}

// 戰鬥中不可用的鍵**各有自己的回應**,不是統一一句「不行」。
//
// 而且 D 與 W 在戰鬥分派器裡也是 "-What?" —— 與主分派器一致,
// 兩處獨立佐證它們不是指令。
func TestCombatRefusalsAreIndividual(t *testing.T) {
	cases := map[rune]string{
		'E': "此處不可",
		'T': "無人回應",
		'B': "對什麼",
		'D': "何事",
		'W': "何事",
	}
	for key, want := range cases {
		got, ok := CombatRefuse(key)
		if !ok {
			t.Errorf("%c 應該在拒絕清單裡", key)
			continue
		}
		if !strings.Contains(got, want) {
			t.Errorf("%c 回的是 %q,應含 %q", key, got, want)
		}
	}
	// 可用的鍵不該落在拒絕清單裡 —— 兩張表重疊就是抄錯了。
	for _, key := range CombatAllowedKeys {
		if _, ok := CombatRefuse(key); ok {
			t.Errorf("%c 同時出現在可用與拒絕兩張表", key)
		}
	}
	// 兩張表加起來應該蓋掉 A..Z 全部 26 個字母(原版 case 65..90 全有著落)。
	seen := map[rune]bool{}
	for _, k := range CombatAllowedKeys {
		seen[k] = true
	}
	for k := range combatRefusals {
		seen[k] = true
	}
	for k := 'A'; k <= 'Z'; k++ {
		if !seen[k] {
			t.Errorf("字母 %c 兩張表都沒有 —— jpt_A5C8 的 case %d 漏抄了", k, k)
		}
	}
}
