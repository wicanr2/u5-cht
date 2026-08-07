package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// inTown 把隊伍放進不列顛城,回傳可以直接跑回合的狀態。
func inTown(t *testing.T) *State {
	t.Helper()
	s := shrineState(t)
	if s == nil {
		return nil
	}
	if err := s.SetScene(2, 0, 15, 15); err != nil { // 不列顛城
		t.Skipf("進不了不列顛城:%v", err)
	}
	return s
}

// TestNPCAITypesAreActuallyUsed:排程裡的行為型別不能是死欄位。
//
// 引擎原本只用了排程的 X / Y / 樓層,行為型別整欄沒人讀 —— 症狀是
// 城裡每個人走到崗位就變蠟像。這條先確認資料裡真的有多種型別,
// 免得日後有人以為那一欄永遠是 0 而把它刪掉。
func TestNPCAITypesAreActuallyUsed(t *testing.T) {
	s := inTown(t)
	if s == nil {
		return
	}
	seen := map[byte]int{}
	for i := range s.npcs {
		n := &s.npcs[i]
		if !n.Present() {
			continue
		}
		for _, a := range n.Schedule.AI {
			seen[a]++
		}
	}
	if len(seen) < 2 {
		t.Errorf("不列顛城的 NPC 只有 %d 種行為型別:%v —— 那一欄大概解錯了", len(seen), seen)
	}
	for a := range seen {
		if a > u5data.NPCAIDrunk {
			t.Errorf("出現型別 %d,而原版的跳表只有 0..7", a)
		}
	}
}

// TestIdleNPCsStillMove:走到崗位之後還是有人會動。
//
// ⚠ 這條是「閒置 ≠ 靜止」的回歸測試。原版在 NPC 抵達崗位之後才呼叫
// `sub_95BC`(行為型別),少了它整座城靜止不動,但**所有既有測試都會照樣全綠**
// —— 因為它們驗的是「不要走到不該走的地方」,不是「要走」。
func TestIdleNPCsStillMove(t *testing.T) {
	s := inTown(t)
	if s == nil {
		return
	}
	// 先讓所有人走到崗位。
	for i := 0; i < 60; i++ {
		s.advanceNPCs()
	}
	type pos struct{ x, y int }
	start := map[int]pos{}
	for _, v := range s.VisibleNPCs() {
		start[v.Index] = pos{v.X, v.Y}
	}
	if len(start) == 0 {
		t.Skip("這個時刻城裡沒人")
	}
	moved := 0
	for i := 0; i < 60; i++ {
		s.advanceNPCs()
	}
	for _, v := range s.VisibleNPCs() {
		if p, ok := start[v.Index]; ok && (p.x != v.X || p.y != v.Y) {
			moved++
		}
	}
	if moved == 0 {
		t.Errorf("跑了 60 回合,%d 個 NPC 一個都沒動過 —— 行為型別大概沒接上", len(start))
	}
}

// TestWanderStaysNearItsPost:遊走型不會離開崗位三格以外。
func TestWanderStaysNearItsPost(t *testing.T) {
	s := inTown(t)
	if s == nil {
		return
	}
	for i := 0; i < 400; i++ {
		s.advanceNPCs()
		for _, v := range s.VisibleNPCs() {
			rt := &s.rtNPCs[v.Index]
			sched := &s.npcs[v.Index].Schedule
			ai := sched.AI[rt.Slot]
			if ai != u5data.NPCAIWander && ai != u5data.NPCAIStay {
				continue
			}
			limit := u5data.NPCAIWanderRange
			if ai == u5data.NPCAIStay {
				limit = 0
			}
			d := manhattan(int(sched.X[rt.Slot]), int(sched.Y[rt.Slot]), rt.X, rt.Y)
			// ⚠ 排程剛換班時 NPC 還在往新崗位走,那時距離當然會超過 ——
			// 只在「已經閒置」的狀態下檢查。
			if rt.Mode == ModeIdle && d > limit {
				t.Fatalf("NPC %d(型別 %d)離崗位 %d 格,上限 %d", v.Index, ai, d, limit)
			}
		}
	}
}

// TestHostileGuardWalksOverAndArrests:叫衛兵之後衛兵會走過來逮捕。
//
// 這是「叫衛兵」→ 敵對 → 追擊 → 接觸 → 逮捕整條鏈的端對端測試。
// 中間任何一環沒接上,玩家喊完衛兵就什麼事都不會發生。
func TestHostileGuardWalksOverAndArrests(t *testing.T) {
	s := inTown(t)
	if s == nil {
		return
	}
	// 找一個衛兵,把他放在玩家附近。
	guard := -1
	for i := range s.npcs {
		if s.npcs[i].Creature == u5data.CreatureGuard {
			guard = i
			break
		}
	}
	if guard < 0 {
		t.Skip("不列顛城此刻沒有衛兵")
	}
	s.CallGuards()
	if got := s.npcs[guard].Schedule.AI[0]; got != u5data.NPCAIHostile &&
		got != u5data.NPCAIHostileBig {
		t.Fatalf("叫完衛兵之後衛兵的型別是 %d,預期 6 或 7", got)
	}
	// ⚠ 排程時刻要一起清掉,不然下一個整點就把他打回崗位。
	for _, tm := range s.npcs[guard].Schedule.Times {
		if tm != 0 {
			t.Errorf("排程時刻沒有清掉(%v)—— 敵意會在下一個整點消失",
				s.npcs[guard].Schedule.Times)
			break
		}
	}
	// 把他挪到玩家旁邊,再跑一回合就該逮捕。
	s.rtNPCs[guard].X, s.rtNPCs[guard].Y = s.X, s.Y-1
	s.rtNPCs[guard].Floor = s.Floor
	s.rtNPCs[guard].Mode = ModeIdle
	s.Messages = nil
	s.advanceNPCs()
	if s.Prompt != PromptArrest {
		t.Fatalf("衛兵貼到身邊卻沒有逮捕(Prompt=%v):\n%s", s.Prompt, s.log())
	}
	if !strings.Contains(s.log(), MsgUnderArrest) {
		t.Errorf("沒有印「%s」:\n%s", MsgUnderArrest, s.log())
	}
}

// TestCallGuardsScattersEveryoneElse:叫衛兵時其餘的人有一半會逃。
//
// ⚠ 少了這一段,叫完衛兵整條街的人還若無其事地站著。這是機率事件,
// 所以看的是「總共有人逃」而不是某一個人逃。
func TestCallGuardsScattersEveryoneElse(t *testing.T) {
	s := inTown(t)
	if s == nil {
		return
	}
	civilians := 0
	for i := range s.npcs {
		n := &s.npcs[i]
		if i != u5data.PartySlot && n.Present() &&
			n.Creature != u5data.CreatureGuard &&
			n.Creature != u5data.CreatureGuardCaptain {
			civilians++
		}
	}
	if civilians < 4 {
		t.Skipf("平民只有 %d 個,樣本太小", civilians)
	}
	s.CallGuards()
	fled := 0
	for i := range s.npcs {
		if s.npcs[i].Schedule.AI[0] == u5data.NPCAIFleeing {
			fled++
		}
	}
	if fled == 0 {
		t.Errorf("%d 個平民一個都沒逃 —— sub_C10 的後半段大概漏了", civilians)
	}
}

// TestComeQuietlyWakesInYewJail:束手就擒 → 在紫杉城的牢房醒來。
//
// ⚠ 三件事都要:地點是紫杉城 (25,4)、時刻是早上八點、**鑰匙歸零**。
// 鑰匙不歸零的話,玩家一被關進去就能開門走人。
func TestComeQuietlyWakesInYewJail(t *testing.T) {
	s := inTown(t)
	if s == nil {
		return
	}
	s.Inventory.Keys = 3
	s.Arrest()
	if s.Prompt != PromptArrest {
		t.Fatalf("沒有問「汝束手就擒否?」:\n%s", s.log())
	}
	s.AnswerArrest(true)
	if s.Location != u5data.ArrestJailLocation {
		t.Errorf("醒來在地點 %d,預期紫杉城 %d", s.Location, u5data.ArrestJailLocation)
	}
	if s.X != u5data.ArrestJailX || s.Y != u5data.ArrestJailY {
		t.Errorf("醒來在 (%d,%d),預期 (%d,%d)",
			s.X, s.Y, u5data.ArrestJailX, u5data.ArrestJailY)
	}
	if s.Clock.Hour != u5data.ArrestWakeHour {
		t.Errorf("醒來時是 %d 點,預期 %d 點", s.Clock.Hour, u5data.ArrestWakeHour)
	}
	if s.Inventory.Keys != 0 {
		t.Errorf("還剩 %d 把鑰匙", s.Inventory.Keys)
	}
}

// TestResistingCallsTheGuards:反抗 → 全城衛兵撲上來。
func TestResistingCallsTheGuards(t *testing.T) {
	s := inTown(t)
	if s == nil {
		return
	}
	before := s.Location
	s.Arrest()
	s.AnswerArrest(false)
	if s.Location != before {
		t.Error("反抗卻被搬到別的地方了")
	}
	if !strings.Contains(s.log(), MsgDefendThyself) {
		t.Errorf("沒有印「%s」:\n%s", MsgDefendThyself, s.log())
	}
	hostile := 0
	for i := range s.npcs {
		if a := s.npcs[i].Schedule.AI[0]; a == u5data.NPCAIHostile || a == u5data.NPCAIHostileBig {
			hostile++
		}
	}
	if hostile == 0 {
		t.Error("反抗之後沒有任何人變敵對")
	}
}

// TestArrestInThePalaceGoesStraightToBlackthorn:在黑棘宮殿被抓不問話。
//
// 原版 `sub_1884` 的第一個分支就是 `cmp byte_3E0A3, 12h` —— 在那裡
// 沒有「束手就擒否」這一問,直接進審問。
func TestArrestInThePalaceGoesStraightToBlackthorn(t *testing.T) {
	s := shrineState(t)
	if s == nil {
		return
	}
	if s.PartySize < 2 {
		t.Skip("隊伍太少")
	}
	s.Location = u5data.BlackthornLocation
	s.Arrest()
	if s.Prompt == PromptArrest {
		t.Error("在黑棘宮殿還問了「汝束手就擒否?」")
	}
	if s.Blackthorn == nil {
		t.Errorf("沒有進到審問:\n%s", s.log())
	}
}
