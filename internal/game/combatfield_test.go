package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// harmScene 造一個戰鬥,並把第 0 槽單位腳下的地形設成 tile。
func harmScene(t *testing.T, tile byte) (*State, int) {
	t.Helper()
	s := combatState(t)
	// ⚠ `combatState` 只是**備好戰鬥資料**,不會進戰鬥 —— 要走進一隻怪物
	// 才會開打(同 `TestCombatMoveStaysOnField`)。第一版少了這兩行,
	// 八條測試全部 SKIP,而 SKIP 在摘要裡看起來跟 PASS 一樣。
	s.CurrentObjects().Spawn(0x40, s.X+1, s.Y, s.Floor)
	s.Move(East)
	c := s.Combat
	if c == nil || !s.InCombat() {
		t.Skip("沒開打")
	}
	slot := -1
	for i := range c.Units {
		if c.Units[i].Active() && c.Units[i].IsParty() {
			slot = i
			break
		}
	}
	if slot < 0 {
		t.Skip("戰場上沒有隊員")
	}
	u := &c.Units[slot]
	if !s.SetTileAt(u.X, u.Y, tile) {
		t.Skip("寫不進戰場")
	}
	s.Messages = nil
	return s, slot
}

// TestLavaAndFireplaceBurnInCombat —— ★ 戰場上的熔岩與壁爐會燒人。
//
// 引擎此前只有場景 / 大地圖那一份地形效果(`sub_1318`),戰場上
// 熔岩、壁爐、沼澤**一件都不作用** —— 把敵人推到熔岩上完全沒事。
func TestLavaAndFireplaceBurnInCombat(t *testing.T) {
	for _, tile := range []byte{TileMoltenLava, TileFireplaceCombat} {
		s, slot := harmScene(t, tile)
		if got := s.harmUnderUnit(slot); got != HarmBurn {
			t.Errorf("tile 0x%02X 的效果是 %d,預期 %d(燒傷)", tile, got, HarmBurn)
		}
		ch := s.charOf(&s.Combat.Units[slot])
		if ch == nil {
			t.Skip("那一槽不是隊員")
		}
		ch.HP, ch.MaxHP = 200, 200
		hurt := false
		for i := 0; i < 60 && !hurt; i++ {
			ch.HP = 200
			s.harmStandingUnit(slot)
			if ch.HP < 200 {
				hurt = true
			}
		}
		if !hurt {
			t.Errorf("tile 0x%02X 站了六十回合都沒受傷", tile)
		}
	}
}

// TestSwampPoisonsInCombat —— 戰場上的沼澤會上毒。
func TestSwampPoisonsInCombat(t *testing.T) {
	s, slot := harmScene(t, TileSwampCombat)
	if got := s.harmUnderUnit(slot); got != HarmPoison {
		t.Fatalf("沼澤的效果是 %d,預期 %d(毒)", got, HarmPoison)
	}
	ch := s.charOf(&s.Combat.Units[slot])
	if ch == nil {
		t.Skip("那一槽不是隊員")
	}
	ch.Status = u5data.StatusGood
	ch.Raw[u5data.CharStatus] = u5data.StatusGood
	s.harmStandingUnit(slot)
	if ch.Status != u5data.StatusPoisoned {
		t.Errorf("踩沼澤之後狀態是 0x%02X,預期中毒", ch.Status)
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgIsPoisoned) {
		t.Errorf("沒印中毒:%q", s.Messages)
	}
}

// TestPoisonHurtsWhenItCannotStick —— ★★ 毒上不了身就改成扣血。
//
// 原版 `sub_B8DC` 的第二條路。少了它,毒力場對**已經中毒的人**與
// **所有怪物**完全無效 —— 而對敵人放毒力場正是玩家最常用它的方式。
func TestPoisonHurtsWhenItCannotStick(t *testing.T) {
	s, slot := harmScene(t, TileSwampCombat)
	ch := s.charOf(&s.Combat.Units[slot])
	if ch == nil {
		t.Skip("那一槽不是隊員")
	}
	ch.Status = u5data.StatusPoisoned // 已經中毒 ⇒ 狀態不是 'G'
	ch.Raw[u5data.CharStatus] = u5data.StatusPoisoned
	ch.HP, ch.MaxHP = 200, 200
	hurt := false
	for i := 0; i < 60 && !hurt; i++ {
		ch.HP = 200
		s.harmStandingUnit(slot)
		if ch.HP < 200 {
			hurt = true
		}
	}
	if !hurt {
		t.Error("已經中毒的人再踩沼澤六十回合都沒受傷")
	}
}

// TestTerrainWinsOverObjects —— ★ 地形先、物件後,而且地形中了就不掃物件。
//
// 站在熔岩上的毒力場只會燒不會毒。
func TestTerrainWinsOverObjects(t *testing.T) {
	s, slot := harmScene(t, TileMoltenLava)
	objs := s.currentObjects()
	if objs == nil {
		t.Skip("沒有物件層")
	}
	u := &s.Combat.Units[slot]
	if _, ok := objs.Spawn(FieldPoisonObj, u.X, u.Y, s.Floor); !ok {
		t.Skip("物件槽滿了")
	}
	if got := s.harmUnderUnit(slot); got != HarmBurn {
		t.Errorf("熔岩 + 毒力場的效果是 %d,預期 %d(地形優先)", got, HarmBurn)
	}
}

// TestThreeFieldObjectsAndTheFourthDoesNothing —— ★ 第四種力場沒有 case。
//
// 0xEB(純力場)站上去什麼都不會發生 —— 原版的 switch 沒有那一格。
// 「補齊」它就是自創規則。
func TestThreeFieldObjectsAndTheFourthDoesNothing(t *testing.T) {
	for _, tc := range []struct {
		kind byte
		want int
		name string
	}{
		{FieldPoisonObj, HarmPoison, "毒力場"},
		{FieldSleepObj, HarmSleep, "睡眠力場"},
		{FieldFire, HarmBurn, "火力場"},
		{0xEB, HarmNone, "★ 第四種 —— 原版沒有 case"},
	} {
		s, slot := harmScene(t, u5data.OpenedDoorTile) // 無害的地板
		objs := s.currentObjects()
		if objs == nil {
			t.Skip("沒有物件層")
		}
		u := &s.Combat.Units[slot]
		if _, ok := objs.Spawn(tc.kind, u.X, u.Y, s.Floor); !ok {
			t.Skip("物件槽滿了")
		}
		if got := s.harmUnderUnit(slot); got != tc.want {
			t.Errorf("%s(0x%02X):效果 %d,預期 %d", tc.name, tc.kind, got, tc.want)
		}
	}
}

// TestSleepFieldClearsPoison —— ★★ 睡眠力場會把中毒擦掉。
//
// 狀態是**單一位元組**,設成 'S' 就把 'P' 蓋掉了。與紮營同一個副作用。
func TestSleepFieldClearsPoison(t *testing.T) {
	s, slot := harmScene(t, u5data.OpenedDoorTile)
	objs := s.currentObjects()
	if objs == nil {
		t.Skip("沒有物件層")
	}
	u := &s.Combat.Units[slot]
	if _, ok := objs.Spawn(FieldSleepObj, u.X, u.Y, s.Floor); !ok {
		t.Skip("物件槽滿了")
	}
	ch := s.charOf(u)
	if ch == nil {
		t.Skip("那一槽不是隊員")
	}
	ch.Status = u5data.StatusPoisoned
	ch.Raw[u5data.CharStatus] = u5data.StatusPoisoned
	s.harmStandingUnit(slot)
	if ch.Status != u5data.StatusAsleep {
		t.Errorf("踩睡眠力場之後狀態是 0x%02X,預期沉睡", ch.Status)
	}
	if u.Flags&UnitAsleep == 0 {
		t.Error("狀態改了但單位旗標沒掛上 UnitAsleep")
	}
}

// TestSleepingDeadStaysDead —— 死人不會被叫去睡(原版 `sub_2EDF8` 擋 'D')。
func TestSleepingDeadStaysDead(t *testing.T) {
	s, slot := harmScene(t, u5data.OpenedDoorTile)
	ch := s.charOf(&s.Combat.Units[slot])
	if ch == nil {
		t.Skip("那一槽不是隊員")
	}
	ch.Status = u5data.StatusDead
	ch.Raw[u5data.CharStatus] = u5data.StatusDead
	s.putUnitToSleep(slot)
	if ch.Status != u5data.StatusDead {
		t.Errorf("死人被改成 0x%02X", ch.Status)
	}
}

// TestHarmlessFloorDoesNothing —— 反對照:普通地板什麼都不會發生。
//
// 少了這一條,「有害格子會作用」與「每回合都掉血」用同一個觀察分不開。
func TestHarmlessFloorDoesNothing(t *testing.T) {
	s, slot := harmScene(t, u5data.OpenedDoorTile)
	if got := s.harmUnderUnit(slot); got != HarmNone {
		t.Errorf("普通地板的效果是 %d,預期 %d", got, HarmNone)
	}
	ch := s.charOf(&s.Combat.Units[slot])
	if ch == nil {
		t.Skip("那一槽不是隊員")
	}
	ch.HP, ch.MaxHP = 200, 200
	ch.Status = u5data.StatusGood
	for i := 0; i < 100; i++ {
		s.harmStandingUnit(slot)
	}
	if ch.HP != 200 || ch.Status != u5data.StatusGood {
		t.Errorf("站在普通地板上一百回合:HP %d、狀態 0x%02X", ch.HP, ch.Status)
	}
}
