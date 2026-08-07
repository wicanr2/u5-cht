package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 紮營能回血、能回法力 —— 這是 U5 唯一的恢復管道,少了它出城就沒得治。
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
	s.AnswerNumber(3)
	// 隊伍超過一人時會問守夜。
	if s.AwaitingYesNo() {
		s.AnswerYesNo(false)
	}
	if c.HP != c.MaxHP {
		t.Errorf("睡完 HP 是 %d,預期回滿 %d", c.HP, c.MaxHP)
	}
	if s.Clock.Hour != 13 {
		t.Errorf("10 時睡 3 小時醒在 %d 時,預期 13 時", s.Clock.Hour)
	}
	for _, m := range s.Party() {
		if m.Status == u5data.StatusAsleep {
			t.Errorf("%s 醒不過來", m.Name)
		}
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

	s.restHours(4) // 22 + 4 = 26 → 26 − 23 = 3(正確的環繞會是 2)
	if s.Clock.Hour != 3 {
		t.Errorf("22 時睡 4 小時醒在 %d 時;原版的算法會給 3 時(減 23 而非 24)", s.Clock.Hour)
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
