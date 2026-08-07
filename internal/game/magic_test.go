package game

import (
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

func magicState(t *testing.T) *State {
	t.Helper()
	s := combatState(t)
	sp, err := u5data.LoadSpells("../../gamedata")
	if err != nil {
		t.Fatal(err)
	}
	s.Spells = sp
	return s
}

// TestMixNeedsTheExactRecipe:配方要完全相符,而且**配錯了藥草照樣沒了**。
//
// 原版 `sub_18704` 先扣藥草再比對 —— 這一條把「先扣後判」的順序釘住,
// 寫成「判對了才扣」會讓玩家可以無限試錯。
func TestMixNeedsTheExactRecipe(t *testing.T) {
	s := magicState(t)
	for i := range s.Inventory.Reagents {
		s.Inventory.Reagents[i] = 10
	}
	mani := s.Spells.Find("Mani")
	// ⚠ INIT.GAM 本來就帶著幾份調配好的咒語,不能假設從 0 開始。
	s.Inventory.Spells[mani] = 0
	// 錯的配方:少一味。
	before := s.Inventory.Reagents[u5data.ReagentGinseng]
	if s.Mix(mani, 2, []int{u5data.ReagentGinseng}) {
		t.Error("只放人參也配得出 Mani?")
	}
	if s.Inventory.Reagents[u5data.ReagentGinseng] != before-2 {
		t.Errorf("配錯了人參卻沒少:%d → %d",
			before, s.Inventory.Reagents[u5data.ReagentGinseng])
	}
	if s.Inventory.Spells[mani] != 0 {
		t.Errorf("配錯了卻拿到 %d 份 Mani", s.Inventory.Spells[mani])
	}
	// 對的配方。
	if !s.MixByRecipe(mani, 3) {
		t.Fatalf("照配方也配不出來:\n%s", s.log())
	}
	if s.Inventory.Spells[mani] != 3 {
		t.Errorf("配出 %d 份,預期 3", s.Inventory.Spells[mani])
	}
}

// TestMixLimitedByScarcestReagent:份數受限於最少的那一味。
func TestMixLimitedByScarcestReagent(t *testing.T) {
	s := magicState(t)
	mani := s.Spells.Find("Mani")
	s.Inventory.Spells[mani] = 0
	s.Inventory.Reagents[u5data.ReagentGinseng] = 9
	s.Inventory.Reagents[u5data.ReagentSpiderSilk] = 2
	if !s.MixByRecipe(mani, 5) {
		t.Fatalf("配不出來:\n%s", s.log())
	}
	if s.Inventory.Spells[mani] != 2 {
		t.Errorf("配出 %d 份,蛛絲只有 2 份時上限就是 2", s.Inventory.Spells[mani])
	}
	if s.Inventory.Reagents[u5data.ReagentGinseng] != 7 {
		t.Errorf("人參剩 %d,預期 9 − 2 = 7", s.Inventory.Reagents[u5data.ReagentGinseng])
	}
}

// TestCastConsumesEvenWhenItFails:施法失敗**照樣**扣掉調配好的那一份。
//
// 原版 `sub_1994C` 的順序是「份數 −1 → 檢查魔力 → 檢查等級」,
// 寫反了會讓玩家用零成本試探自己的等級夠不夠。
func TestCastConsumesEvenWhenItFails(t *testing.T) {
	s := magicState(t)
	s.Location = 0 // 大地圖
	mani := s.Spells.Find("Mani")
	s.Inventory.Spells[mani] = 2
	ch := &s.Roster[0]
	ch.MP = 0 // 魔力歸零 → 一定失敗
	if got := s.Cast(0, mani); got != MagicNoMana {
		t.Errorf("魔力 0 施法回傳 %v,預期 MagicNoMana", got)
	}
	if s.Inventory.Spells[mani] != 1 {
		t.Errorf("失敗後剩 %d 份,預期已經扣掉一份剩 1", s.Inventory.Spells[mani])
	}
}

// TestCastChecksPlaceBeforeConsuming:場合不對是在扣份數**之前**擋下的。
func TestCastChecksPlaceBeforeConsuming(t *testing.T) {
	s := magicState(t)
	s.Location = 0 // 大地圖
	uus := s.Spells.Find("Uus Por")
	s.Inventory.Spells[uus] = 1
	if got := s.Cast(0, uus); got != MagicNotHere {
		t.Errorf("在大地圖施 Uus Por 回傳 %v,預期 MagicNotHere", got)
	}
	if s.Inventory.Spells[uus] != 1 {
		t.Errorf("場合不對就該原封不動,卻剩 %d 份", s.Inventory.Spells[uus])
	}
}

// TestManaCostIsTheCircle:魔力消耗等於圈數。
func TestManaCostIsTheCircle(t *testing.T) {
	s := magicState(t)
	s.Location = 0
	ch := &s.Roster[0]
	ch.Level = 8
	for _, name := range []string{"Mani", "Vas Mani", "An Tym"} {
		i := s.Spells.Find(name)
		circle := s.Spells.Spells[i].Circle
		s.Inventory.Spells[i] = 1
		ch.MP = 40
		before := ch.MP
		s.Cast(0, i)
		if int(before)-int(ch.MP) != circle {
			t.Errorf("%s(第 %d 圈)扣了 %d 點魔力,預期 %d",
				name, circle, int(before)-int(ch.MP), circle)
		}
	}
}

// TestLevelGatesHighCircles:等級不足發不動高圈咒語 —— 但魔力照扣。
func TestLevelGatesHighCircles(t *testing.T) {
	s := magicState(t)
	s.Location = 0
	tym := s.Spells.Find("An Tym") // 第 8 圈
	s.Inventory.Spells[tym] = 1
	ch := &s.Roster[0]
	ch.Level, ch.MP = 3, 40
	if got := s.Cast(0, tym); got != MagicFailed {
		t.Errorf("等級 3 施第 8 圈回傳 %v,預期 MagicFailed", got)
	}
	if ch.MP != 32 {
		t.Errorf("魔力剩 %d,預期 40 − 8 = 32(等級檢查在扣魔力之後)", ch.MP)
	}
	if s.TimeStop != 0 {
		t.Error("等級不足卻還是把時間停了")
	}
}

// TestManiHealsWithinRange:Mani 回 1..30,而且不會超過上限。
func TestManiHealsWithinRange(t *testing.T) {
	s := magicState(t)
	s.Location = 0
	mani := s.Spells.Find("Mani")
	ch := &s.Roster[0]
	ch.Level, ch.MaxHP = 8, 200
	// 平時的治療對象是「傷得最重的那一個」,先讓其他隊員滿血。
	for i := 1; i < s.PartySize; i++ {
		s.Roster[i].HP = s.Roster[i].MaxHP
	}
	for i := 0; i < 40; i++ {
		ch.HP = 100
		ch.MP = 40
		s.Inventory.Spells[mani] = 1
		if got := s.Cast(0, mani); got != MagicSuccess {
			t.Fatalf("Mani 失敗:%v\n%s", got, s.log())
		}
		d := int(ch.HP) - 100
		if d < 1 || d > 30 {
			t.Fatalf("回了 %d 點,預期 1..30", d)
		}
	}
	// 不會超過上限。
	ch.HP, ch.MP = ch.MaxHP-2, 40
	s.Inventory.Spells[mani] = 1
	s.Cast(0, mani)
	if ch.HP != ch.MaxHP {
		t.Errorf("補到 %d,上限是 %d", ch.HP, ch.MaxHP)
	}
}

// TestVasManiFullHealAndResurrect:大治療補滿、復活救得回死人。
func TestVasManiFullHealAndResurrect(t *testing.T) {
	s := magicState(t)
	s.Location = 2 // 城鎮(復活不能在戰鬥中施)
	ch := &s.Roster[0]
	ch.Level, ch.MaxHP, ch.MP = 8, 200, 40
	for i := 1; i < s.PartySize; i++ {
		s.Roster[i].HP = s.Roster[i].MaxHP
	}

	vas := s.Spells.Find("Vas Mani")
	ch.HP = 5
	s.Inventory.Spells[vas] = 1
	s.Cast(0, vas)
	if ch.HP != ch.MaxHP {
		t.Errorf("大治療後 %d/%d,預期補滿", ch.HP, ch.MaxHP)
	}

	res := s.Spells.Find("In Mani Co")
	ch.Status, ch.HP, ch.MP = u5data.StatusDead, 0, 40
	s.Inventory.Spells[res] = 1
	if got := s.Cast(0, res); got != MagicSuccess {
		t.Fatalf("復活失敗:%v\n%s", got, s.log())
	}
	if ch.Status != u5data.StatusGood || ch.HP != 1 {
		t.Errorf("復活後狀態 %c、HP %d,預期 G / 1", ch.Status, ch.HP)
	}
}

// TestAnNoxCuresPoison:解毒把狀態 'P' 變回 'G'。
func TestAnNoxCuresPoison(t *testing.T) {
	s := magicState(t)
	s.Location = 0
	nox := s.Spells.Find("An Nox")
	s.Roster[0].Level, s.Roster[0].MP = 8, 40
	s.Roster[1].Status = u5data.StatusPoisoned
	s.Inventory.Spells[nox] = 1
	if got := s.Cast(0, nox); got != MagicSuccess {
		t.Fatalf("解毒失敗:%v\n%s", got, s.log())
	}
	if s.Roster[1].Status != u5data.StatusGood {
		t.Errorf("解毒後狀態是 %c", s.Roster[1].Status)
	}
}

// TestInterferenceBlocksCastingInCombat:相鄰的敵人會打斷施法。
//
// 原版 `sub_200BC` 只認**上一個打你的那個人**,而且要正好貼著。
func TestInterferenceBlocksCastingInCombat(t *testing.T) {
	s := magicState(t)
	slot, _ := s.CurrentObjects().Spawn(0x40, s.X+1, s.Y, s.Floor)
	if !s.BeginCombat(slot) {
		t.Fatal("打不起來")
	}
	c := s.Combat
	me, foe := 0, u5data.CombatPartySlots
	// 把敵人擺到貼身,並且記成「上一個打我的人」。
	c.Units[foe].X, c.Units[foe].Y = c.Units[me].X+1, c.Units[me].Y
	c.LastAttacker[me] = int8(foe)

	mani := s.Spells.Find("Mani")
	s.Inventory.Spells[mani] = 1
	s.Roster[0].Level, s.Roster[0].MP = 8, 40
	if got := s.Cast(0, mani); got != MagicInterfered {
		t.Errorf("貼身敵人在旁邊施法回傳 %v,預期 MagicInterfered", got)
	}
	if s.Inventory.Spells[mani] != 1 {
		t.Error("被打斷不該消耗調配好的咒語")
	}
	// 把敵人挪遠就施得成。
	c.Units[foe].X = c.Units[me].X + 5
	if got := s.Cast(0, mani); got != MagicSuccess && got != MagicFailed {
		t.Errorf("敵人挪遠之後回傳 %v,不該再是被打斷", got)
	}
}

// TestXenCorpKills:Xen Corp 是必殺。
func TestXenCorpKills(t *testing.T) {
	s := magicState(t)
	slot, _ := s.CurrentObjects().Spawn(0x40, s.X+1, s.Y, s.Floor)
	if !s.BeginCombat(slot) {
		t.Fatal("打不起來")
	}
	xen := s.Spells.Find("Xen Corp")
	s.Inventory.Spells[xen] = 1
	s.Roster[0].Level, s.Roster[0].MP = 8, 40
	before, _ := s.sideCounts(s.Combat)
	if got := s.Cast(0, xen); got != MagicSuccess {
		t.Fatalf("Xen Corp 失敗:%v\n%s", got, s.log())
	}
	after, _ := s.sideCounts(s.Combat)
	if after != before-1 {
		t.Errorf("敵人 %d → %d,必殺該少一個", before, after)
	}
}

// TestUnimplementedSpellsStillCost:還沒實作的咒語仍照原版消耗,而且誠實回失敗。
//
// 這一條是刻意寫的:假裝成功會讓「魔法做完了」這句話變成謊。
func TestUnimplementedSpellsStillCost(t *testing.T) {
	s := magicState(t)
	s.Location = 0
	// An Sanct(解陷阱 / 開箱)還沒實作,四處皆可施。
	wis := s.Spells.Find("An Sanct")
	s.Inventory.Spells[wis] = 1
	ch := &s.Roster[0]
	ch.Level, ch.MP = 8, 40
	if got := s.Cast(0, wis); got != MagicFailed {
		t.Errorf("未實作的咒語回傳 %v,預期 MagicFailed", got)
	}
	if s.Inventory.Spells[wis] != 0 || ch.MP != 38 {
		t.Errorf("未實作也該照扣:剩 %d 份、魔力 %d(預期 0 份、38)",
			s.Inventory.Spells[wis], ch.MP)
	}
}

// TestCombatModesSharaeOneByte:五個持續咒語共用同一個位元組,後施的蓋掉先施的。
//
// 原版 `sub_1D31C` 就是 `byte_3E08A = 模式; byte_3E09E = 回合` ——
// 不是旗標集合。寫成可疊加會讓玩家同時吃到緩速 + 混亂 + 抗魔。
func TestCombatModesShareOneByte(t *testing.T) {
	s := magicState(t)
	s.Location = 0
	ch := &s.Roster[0]
	ch.Level = 8
	cast := func(name string) {
		t.Helper()
		i := s.Spells.Find(name)
		s.Inventory.Spells[i] = 1
		ch.MP = 40
		if got := s.Cast(0, i); got != MagicSuccess {
			t.Fatalf("%s 失敗:%v\n%s", name, got, s.log())
		}
	}
	cast("Rel Tym")
	if s.CombatMode != CombatModeSlow || s.CombatModeTurns != 30 {
		t.Errorf("Rel Tym 之後模式 %q 剩 %d 回合,預期 'Q' / 30",
			string(s.CombatMode), s.CombatModeTurns)
	}
	// ⚠ Quas An Wis 與 In An 的可施法場合只有「戰鬥」,大地圖上施不動 ——
	// 所以這裡改用同樣四處皆可的 An Tym 驗「後施的蓋掉先施的」。
	cast("An Tym")
	if s.CombatMode != CombatModeTimeStop || s.CombatModeTurns != TimeStopTurns {
		t.Errorf("An Tym 沒有蓋掉 Rel Tym:模式 %q 剩 %d 回合",
			string(s.CombatMode), s.CombatModeTurns)
	}
	if s.TimeStop != TimeStopTurns {
		t.Errorf("TimeStop 是 %d,預期 %d", s.TimeStop, TimeStopTurns)
	}
	cast("In Sanct")
	if s.CombatMode != CombatModeProtected || s.CombatModeTurns != 20 {
		t.Errorf("In Sanct 之後模式 %q 剩 %d 回合,預期 'P' / 20",
			string(s.CombatMode), s.CombatModeTurns)
	}
}

// TestInXenManiMakesFood:造食物一次 1..3 份。
func TestInXenManiMakesFood(t *testing.T) {
	s := magicState(t)
	s.Location = 0
	ch := &s.Roster[0]
	ch.Level = 8
	i := s.Spells.Find("In Xen Man")
	for n := 0; n < 30; n++ {
		s.Inventory.Food = 100
		ch.MP = 40
		s.Inventory.Spells[i] = 1
		if got := s.Cast(0, i); got != MagicSuccess {
			t.Fatalf("造食物失敗:%v", got)
		}
		if d := s.Inventory.Food - 100; d < 1 || d > 3 {
			t.Fatalf("造出 %d 份糧食,預期 1..3", d)
		}
	}
}

// TestSanctLorHidesFromTargeting:隱形之後敵人選不到你。
func TestSanctLorHidesFromTargeting(t *testing.T) {
	s := magicState(t)
	slot, _ := s.CurrentObjects().Spawn(0x40, s.X+1, s.Y, s.Floor)
	if !s.BeginCombat(slot) {
		t.Fatal("打不起來")
	}
	c := s.Combat
	foe := u5data.CombatPartySlots
	// 先確認敵人本來選得到某個隊員。
	before, _, _ := s.aiTarget(foe)
	if before < 0 {
		t.Fatal("敵人本來就沒有目標")
	}
	// 全隊隱形。
	for i := 0; i < u5data.CombatPartySlots; i++ {
		if c.Units[i].Active() {
			c.Units[i].Flags |= UnitHidden
		}
	}
	if after, _, _ := s.aiTarget(foe); after >= 0 {
		t.Errorf("全隊隱形之後敵人還鎖定得到第 %d 槽", after)
	}
}

// TestWindSpellsSweepALine:四種風掃過一條線上的每個敵人,而且各只作用一次。
func TestWindSpellsSweepALine(t *testing.T) {
	s := magicState(t)
	s.SeedRandom(9)
	slot, _ := s.CurrentObjects().Spawn(0x40, s.X+1, s.Y, s.Floor)
	if !s.BeginCombat(slot) {
		t.Fatal("打不起來")
	}
	c := s.Combat
	// 找一列箭飛得過的,把施法者放在最左邊、兩隻敵人排在右邊。
	row := -1
	for y := 0; y < u5data.CombatSide && row < 0; y++ {
		clear := true
		for x := 0; x < 6; x++ {
			if u5data.TileBlocksProjectile(int(c.Map.At(x, y))) {
				clear = false
			}
		}
		if clear {
			row = y
		}
	}
	if row < 0 {
		t.Skip("找不到乾淨的一列")
	}
	for i := range c.Units {
		c.Units[i].Flags = 0
	}
	c.Units[0] = Combatant{Roster: 0, Creature: -1, Flags: UnitParty, X: 0, Y: row}
	for n, x := range []int{2, 4} {
		i := u5data.CombatPartySlots + n
		c.Units[i] = Combatant{Roster: -1, Creature: 0, Kind: 0x40,
			Flags: UnitMonster, HP: 99, X: x, Y: row}
	}
	before := [2]int{c.Units[6].HP, c.Units[7].HP}
	if !s.castWind(0, windFire, East) {
		t.Fatalf("火風沒打中任何人:\n%s", s.log())
	}
	for n := 0; n < 2; n++ {
		u := &c.Units[u5data.CombatPartySlots+n]
		if u.Active() && u.HP >= before[n] {
			t.Errorf("線上第 %d 隻沒受傷:%d → %d", n, before[n], u.HP)
		}
	}
}

// TestWindStopsAtBlockingTerrain:風也吃投射物的擋路規則。
func TestWindStopsAtBlockingTerrain(t *testing.T) {
	s := magicState(t)
	slot, _ := s.CurrentObjects().Spawn(0x40, s.X+1, s.Y, s.Floor)
	if !s.BeginCombat(slot) {
		t.Fatal("打不起來")
	}
	c := s.Combat
	for idx := range s.CombatMaps.Maps {
		m := &s.CombatMaps.Maps[idx]
		for y := 0; y < u5data.CombatSide; y++ {
			wall := -1
			for x := 1; x < u5data.CombatSide-2; x++ {
				if u5data.TileBlocksProjectile(int(m.At(x, y))) {
					wall = x
					break
				}
			}
			if wall < 2 {
				continue
			}
			c.Map = m
			for i := range c.Units {
				c.Units[i].Flags = 0
			}
			c.Units[0] = Combatant{Roster: 0, Creature: -1, Flags: UnitParty, X: 0, Y: y}
			foe := &c.Units[u5data.CombatPartySlots]
			*foe = Combatant{Roster: -1, Creature: 0, Kind: 0x40,
				Flags: UnitMonster, HP: 99, X: wall + 1, Y: y}
			hp := foe.HP
			s.castWind(0, windFire, East)
			if foe.HP != hp {
				t.Errorf("牆在 x=%d,牆後 x=%d 的敵人卻受了傷(%d → %d)",
					wall, foe.X, hp, foe.HP)
			}
			return
		}
	}
	t.Skip("找不到列上有擋箭地形的戰鬥地圖")
}
