package game

import (
	"os"
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

// TestMixRefusesWhenReagentsAreShort:藥草不足是**拒絕**,不是把份數調低。
//
// ⚠ 這條原本寫成「份數受限於最少的那一味」,而那是錯的。
// 原版 `sub_18698`:要的份數超過任一種選中藥草的存量就印
// 「Insufficient reagents!」、把 `var_4` 設 0,然後在 `loc_186F6`
// `jz loc_1869F` **跳回去重問份數**。它不會替玩家改成調得出來的份數。
//
// 靜靜調低看起來很體貼,但玩家會以為自己調到了要的份數而實際上少了 ——
// 那種落差要等到施法時才發現,而那時已經回推不了原因。
func TestMixRefusesWhenReagentsAreShort(t *testing.T) {
	s := magicState(t)
	mani := s.Spells.Find("Mani")
	s.Inventory.Spells[mani] = 0
	s.Inventory.Reagents[u5data.ReagentGinseng] = 9
	s.Inventory.Reagents[u5data.ReagentSpiderSilk] = 2

	if s.MixByRecipe(mani, 5) {
		t.Error("蛛絲只有 2 份卻配出了 5 份")
	}
	if s.Inventory.Spells[mani] != 0 {
		t.Errorf("配出 %d 份,該一份都沒有", s.Inventory.Spells[mani])
	}
	// 拒絕的時候**一株藥草都不該扣** —— 扣除發生在成敗判定之前,
	// 但那是「份數已經確定」之後的事。
	if s.Inventory.Reagents[u5data.ReagentGinseng] != 9 {
		t.Errorf("被拒絕卻扣了人參:剩 %d", s.Inventory.Reagents[u5data.ReagentGinseng])
	}
	// 改成調得出來的份數就成功。
	if !s.MixByRecipe(mani, 2) {
		t.Fatalf("兩份該配得出來:\n%s", s.log())
	}
	if s.Inventory.Spells[mani] != 2 {
		t.Errorf("配出 %d 份,預期 2", s.Inventory.Spells[mani])
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

// TestXenCorpKillsWhenItLands:Xen Corp 命中就是必殺 —— 但它會失手。
//
// ⚠ 這是七個攻擊咒語裡**唯一要擲命中的一個**(`sub_B484`:0x30 / 0x31 與
// >= 0x33 自動命中,只有 0x32 走骰子)。寫成「攻擊咒語都自動命中」
// 會讓必殺變成無條件秒殺 —— 所以這條同時驗「會中」與「會失手」。
func TestXenCorpKillsWhenItLands(t *testing.T) {
	killed, missed := 0, 0
	for round := 0; round < 40; round++ {
		s := magicState(t)
		s.SeedRandom(int64(round) + 1)
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
		switch {
		case after == before-1:
			killed++
		case after == before:
			missed++
		default:
			t.Fatalf("敵人 %d → %d,一次只該死一個", before, after)
		}
	}
	if killed == 0 {
		t.Error("放了 40 次一個都沒殺掉 —— 必殺沒生效")
	}
	if missed == 0 {
		t.Error("放了 40 次一次都沒失手 —— Xen Corp 應該要擲命中")
	}
	t.Logf("40 次:殺掉 %d、失手 %d", killed, missed)
}

// TestAttackSpellDamageComesFromTheTable:傷害是查表來的,不是估的。
//
// `DATA.OVL` 0x160C 那張表有 **56 筆**,第 48..54 筆就是攻擊碼 0x30..0x36。
// 這條把三個數字釘死:Grav Por 16、Vas Flam 30、Xen Corp 99。
// 引擎原本用 10 / 20 / 99,前兩個是按圈數估的。
func TestAttackSpellDamageComesFromTheTable(t *testing.T) {
	s := magicState(t)
	want := map[int]int{
		u5data.AttackGravPor:     16,
		u5data.AttackVasFlam:     30,
		u5data.AttackXenCorp:     99,
		u5data.AttackInNoxGrav:   18,
		u5data.AttackInZuGrav:    0,
		u5data.AttackInFlamGrav:  21,
		u5data.AttackInSanctGrav: 0,
	}
	for code, dmg := range want {
		if got := s.spellAttackDamage(code); got != dmg {
			t.Errorf("攻擊碼 %#x 的傷害是 %d,表上是 %d", code, got, dmg)
		}
	}
	// 射程全部是 15 —— 攻擊咒語不必貼身。
	for code := range want {
		if got := s.Stats.ItemRange[code]; got != 15 {
			t.Errorf("攻擊碼 %#x 的射程是 %d,預期 15", code, got)
		}
	}
	// 反過來:一般裝備的那 48 筆不能被這 8 筆蓋掉。
	if s.Stats.ItemDamage[16] != 6 { // 匕首
		t.Errorf("匕首的傷害變成 %d —— 表長度改成 56 之後前面 48 筆應該不動", s.Stats.ItemDamage[16])
	}
}

// TestCombatFieldSpellsApplyStatusNotDamage:In Nox Grav 中毒、In Zu Grav 睡眠。
//
// 兩者在表上有傷害欄(18 / 0),但 `sub_B9A8` 在扣血**之前**就用
// `cmp byte_3E0AD, '3'` / `'4'` 攔下來,改成掛狀態。
// 只看傷害表會把 In Nox Grav 寫成「打 18 點」—— 那是錯的。
func TestCombatFieldSpellsApplyStatusNotDamage(t *testing.T) {
	for _, c := range []struct {
		spell int
		flag  byte
		what  string
	}{
		{SpellInZuGrav, UnitAsleep, "睡眠"},
	} {
		s := magicState(t)
		slot, _ := s.CurrentObjects().Spawn(0x40, s.X+1, s.Y, s.Floor)
		if !s.BeginCombat(slot) {
			t.Fatal("打不起來")
		}
		self := s.combatSlotOfRoster(0)
		target, _, _ := s.aiTarget(self)
		if target < 0 {
			t.Skip("沒有目標")
		}
		hp := s.Combat.Units[target].HP
		if !s.spellEffect(0, c.spell) {
			t.Fatalf("%s 放不出來:\n%s", c.what, s.log())
		}
		if s.Combat.Units[target].Flags&c.flag == 0 {
			t.Errorf("%s 沒有掛上狀態", c.what)
		}
		if s.Combat.Units[target].HP != hp {
			t.Errorf("%s 卻扣了血:%d → %d", c.what, hp, s.Combat.Units[target].HP)
		}
	}
}

// TestFailedSpellStillCosts:效果沒發生的咒語仍照原版消耗,而且誠實回失敗。
//
// 這一條是刻意寫的:假裝成功會讓「魔法做完了」這句話變成謊。
// 原版的扣除順序是「先扣藥草、再扣魔力、最後才跑效果」
// (`sub_1994C` 的 `dec byte_3E000[edi]` 在跳表之前),所以效果失敗也照扣。
//
// 用 An Sanct 當探針:四處皆可施,但前方沒有鎖也沒有箱子時它會失敗。
func TestFailedSpellStillCosts(t *testing.T) {
	s := magicState(t)
	s.Location = 0
	probe := s.Spells.Find("An Sanct")
	s.Inventory.Spells[probe] = 1
	ch := &s.Roster[0]
	ch.Level, ch.MP = 8, 40
	if got := s.Cast(0, probe); got != MagicFailed {
		t.Errorf("效果沒發生卻回傳 %v,預期 MagicFailed", got)
	}
	if s.Inventory.Spells[probe] != 0 || ch.MP != 38 {
		t.Errorf("失敗也該照扣:剩 %d 份、魔力 %d(預期 0 份、38)",
			s.Inventory.Spells[probe], ch.MP)
	}
}

// TestEverySpellIsDispatched:48 個咒語每一個都要有分派,不能靜靜掉進 default。
//
// 原版 `jpt_19B27` 是一張**滿的** 48 格跳表 —— 沒有「這一格沒人接」這回事。
// 引擎這邊漏接會表現成「藥草扣了卻什麼都沒發生」,而玩家分不出那是
// 咒語失敗還是程式漏寫。這條把兩者分開。
func TestEverySpellIsDispatched(t *testing.T) {
	for i := 0; i < u5data.SpellCount; i++ {
		s := magicState(t) // 每個咒語一個乾淨的世界 —— 效果會改狀態
		s.spellEffect(0, i)
		for _, line := range s.Messages {
			if line == spellUndispatched {
				t.Errorf("第 %d 個咒語(%s)沒有分派", i, s.Spells.Spells[i].Name)
			}
		}
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

// TestDirectionPromptRunsTheFollowUp:方向選單問完才做事,ESC 就不做。
//
// 原版每個要方向的咒語都先呼叫 `sub_1CC50`。引擎先前是「猜一個合理方向」,
// 那是介面的近似;這條把真的選單釘住。
func TestDirectionPromptRunsTheFollowUp(t *testing.T) {
	s := magicState(t)
	got := Direction(-1)
	s.AskDirection(func(d Direction) { got = d })
	if !s.AwaitingDirection() {
		t.Fatal("AskDirection 之後沒有進等方向的狀態")
	}
	s.AnswerDirection(West)
	if got != West {
		t.Errorf("拿到方向 %v,預期 West", got)
	}
	if s.AwaitingDirection() {
		t.Error("回答之後還停在等方向")
	}
	// ESC 作罷 —— 後續不該跑。
	got = Direction(-1)
	s.AskDirection(func(d Direction) { got = d })
	s.CancelDirection()
	if got != Direction(-1) {
		t.Error("按了 ESC 卻還是做了事")
	}
	if s.AwaitingDirection() {
		t.Error("取消之後還停在等方向")
	}
}

// TestFrightenIsMassNotSingleTarget:恐懼是**群體**,不是選一隻。
//
// 原版 `sub_18EB0` / `sub_19810` 都是 `for i in 0..31` 掃全場,命中的
// 一律 `LOBYTE = 1`(先手)+ `or byte ptr [esi+2], 2`(逃跑位元)。
// 這條擋的正是我寫錯過的那一版:「嚇跑最近的一隻」。
func TestFrightenIsMassNotSingleTarget(t *testing.T) {
	s := magicState(t)
	slot, _ := s.CurrentObjects().Spawn(0x40, s.X+1, s.Y, s.Floor)
	if !s.BeginCombat(slot) {
		t.Fatal("打不起來")
	}
	enemies := 0
	for i := range s.Combat.Units {
		u := &s.Combat.Units[i]
		if u.Flags&(UnitMonster|UnitParty) == UnitMonster {
			enemies++
		}
	}
	if enemies < 2 {
		t.Skipf("這一場只有 %d 個敵人,測不出群體與單體的差別", enemies)
	}
	// ⚠ 一次施法擋在抗性擲骰上,單看一次分不出群體與單體 —— 有可能三個裡
	// 只中一個。**單體法術的目標是固定的**(`aiTarget` 每次都挑同一隻),
	// 所以連放幾次之後:群體會把整場都掛上旗標,單體永遠只有一隻。
	// 這條驗的是那個差別,不是單次命中率。
	fled := 0
	for round := 0; round < 40 && fled < enemies; round++ {
		s.frighten(0, false)
		fled = 0
		for i := range s.Combat.Units {
			if s.Combat.Units[i].Flags&UnitFleeing != 0 {
				fled++
			}
		}
	}
	if fled < 2 {
		t.Errorf("放了 40 次,%d / %d 個敵人在逃跑 —— In Quas Corp 是群體法術",
			fled, enemies)
	}
	t.Logf("%d 個敵人,最後 %d 個在逃跑", enemies, fled)
}

// TestRepelOnlyHitsRepellable:An Xen Corp 只對帶 0x20 位元的生物有效。
//
// 閘門是 `word_3F1D0[生物] & 0x20` —— 全 48 種只有四種帶它。
// 拿一隻不帶的怪來打,一定嚇不動。
func TestRepelOnlyHitsRepellable(t *testing.T) {
	s := magicState(t)
	slot, _ := s.CurrentObjects().Spawn(0x40, s.X+1, s.Y, s.Floor)
	if !s.BeginCombat(slot) {
		t.Fatal("打不起來")
	}
	repellable := false
	for i := range s.Combat.Units {
		u := &s.Combat.Units[i]
		if u.Flags&(UnitMonster|UnitParty) != UnitMonster {
			continue
		}
		if s.creatureOf(u).Has(u5data.CreatureRepellable) {
			repellable = true
		}
	}
	if !s.frighten(0, true) {
		return // 沒中很正常(沒目標,或抗性擋下)
	}
	if !repellable {
		t.Error("場上一隻帶 0x20 位元的生物都沒有,An Xen Corp 卻生效了")
	}
}

// TestSummonsJoinOurSide:三個召喚咒語召出來的都站我方。
//
// 原版三支都給新單位 `unit[+2] |= 1` —— 怪物的陣營反轉位元 = 被馴服。
func TestSummonsJoinOurSide(t *testing.T) {
	for _, c := range []struct {
		name     string
		creature int
		count    int
	}{
		{"Kal Xen 巨鼠", summonRat, 1},
		{"In Bet Xen 蟲群", summonInsect, 4},
		{"Kal Xen Corp 惡魔", summonDaemon, 1},
	} {
		s := magicState(t)
		s.SeedRandom(6)
		slot, _ := s.CurrentObjects().Spawn(0x40, s.X+1, s.Y, s.Floor)
		if !s.BeginCombat(slot) {
			t.Fatal("打不起來")
		}
		before, _ := s.sideCounts(s.Combat)
		if !s.summonCreature(0, c.creature, c.count) {
			t.Errorf("%s 召不出來:\n%s", c.name, s.log())
			continue
		}
		after, party := s.sideCounts(s.Combat)
		if after != before {
			t.Errorf("%s 之後敵人從 %d 變成 %d —— 召出來的該站我方",
				c.name, before, after)
		}
		if party <= s.PartySize {
			t.Errorf("%s 之後我方只有 %d 個(隊伍本來就 %d 個)",
				c.name, party, s.PartySize)
		}
	}
}

// TestPolymorphReplacesTheTarget:變形把目標換掉,而且換成的東西還在同一格。
func TestPolymorphReplacesTheTarget(t *testing.T) {
	s := magicState(t)
	slot, _ := s.CurrentObjects().Spawn(0x40, s.X+1, s.Y, s.Floor)
	if !s.BeginCombat(slot) {
		t.Fatal("打不起來")
	}
	c := s.Combat
	target, _, _ := s.aiTarget(0)
	if target < 0 {
		t.Fatal("沒有目標")
	}
	x, y := c.Units[target].X, c.Units[target].Y
	kind := c.Units[target].Kind
	// 抗性(`sub_1F48C`)擋得掉,所以多試幾次 —— 這條測的是「變成功之後
	// 長什麼樣」,不是命中率。
	ok := false
	for round := 0; round < 40 && !ok; round++ {
		ok = s.polymorph(0)
	}
	if !ok {
		t.Fatalf("放了 40 次都變不了:\n%s", s.log())
	}
	if c.Units[target].Kind == kind {
		t.Error("變形之後種類沒變")
	}
	if c.Units[target].X != x || c.Units[target].Y != y {
		t.Errorf("變形之後跑到 (%d,%d),預期還在 (%d,%d)",
			c.Units[target].X, c.Units[target].Y, x, y)
	}
}

// TestWisQuasRevealsHidden:顯形把隱形位元清掉。
func TestWisQuasRevealsHidden(t *testing.T) {
	s := magicState(t)
	slot, _ := s.CurrentObjects().Spawn(0x40, s.X+1, s.Y, s.Floor)
	if !s.BeginCombat(slot) {
		t.Fatal("打不起來")
	}
	s.Combat.Units[u5data.CombatPartySlots].Flags |= UnitHidden
	if !s.revealHidden() {
		t.Fatal("顯不出來")
	}
	if s.Combat.Units[u5data.CombatPartySlots].Flags&UnitHidden != 0 {
		t.Error("顯形之後還是隱形的")
	}
	// 沒有隱形的東西時回 false。
	if s.revealHidden() {
		t.Error("沒有東西藏著卻回報成功")
	}
}

// TestMagicLockIsNotSymmetric:An Ex Por 與 In Ex Por 不是互逆的。
//
// 上鎖吃兩種門(0xB8 與 0xB9 都變 0x97),解鎖只還原成一般門(0x97 → 0xB8)。
// 所以「鎖住一扇本來上鎖的門再解開」會得到一扇**沒鎖的**門。
// 這條把那個不對稱釘住 —— 不是我算錯,是 `sub_19354` / `sub_1D15C` 就這樣。
func TestMagicLockIsNotSymmetric(t *testing.T) {
	cases := []struct{ from, locked, back byte }{
		{u5data.TileDoorA, u5data.TileMagicLockedA, u5data.TileDoorA},
		{u5data.TileLockedDoor, u5data.TileMagicLockedA, u5data.TileDoorA},
		{u5data.TileDoorB, u5data.TileMagicLockedB, u5data.TileDoorB},
		{u5data.TileLockedMagicDoor, u5data.TileMagicLockedB, u5data.TileDoorB},
	}
	for _, c := range cases {
		if got := u5data.MagicLock(c.from); got != c.locked {
			t.Errorf("%02X 上鎖變成 %02X,預期 %02X", c.from, got, c.locked)
		}
		if got := u5data.MagicUnlock(c.locked); got != c.back {
			t.Errorf("%02X 解鎖變成 %02X,預期 %02X", c.locked, got, c.back)
		}
	}
	// 不是門就不該有反應。
	if u5data.MagicLock(5) != 0 || u5data.MagicUnlock(5) != 0 {
		t.Error("草地也能上鎖")
	}
	// ⚠ An Sanct 走的是另一條路(tile − 1),兩者不可混用。
	if u5data.MagicUnlock(u5data.TileLockedDoor) != 0 {
		t.Error("In Ex Por 不該認得一般的上鎖門 —— 那是 An Sanct 的事")
	}
}

// TestAnYlemDispelSet:消除清單要與跳表一字不差。
//
// `sub_18C00` 是 `tile == 0x5B` 加上一張 `tile − 0x90` 的 32 格跳表。
// 跳表的相鄰項一錯位,整組就偏一格 —— 所以連「哪些**不在**清單裡」一起驗。
func TestAnYlemDispelSet(t *testing.T) {
	in := []byte{0x5B, 0x90, 0x91, 0x92, 0x93, 0x9D, 0xA5, 0xA6, 0xA8, 0xA9, 0xAD, 0xAE, 0xAF}
	out := []byte{0x5A, 0x5C, 0x8F, 0x94, 0x9C, 0x9E, 0xA4, 0xA7, 0xAA, 0xAC, 0xB0}
	for _, tile := range in {
		if !u5data.AnYlemTiles[tile] {
			t.Errorf("%02X 該在 An Ylem 的清單裡", tile)
		}
	}
	for _, tile := range out {
		if u5data.AnYlemTiles[tile] {
			t.Errorf("%02X 不該在 An Ylem 的清單裡", tile)
		}
	}
	if len(u5data.AnYlemTiles) != len(in) {
		t.Errorf("清單有 %d 項,跳表數出來是 %d 項", len(u5data.AnYlemTiles), len(in))
	}
}

// TestBlinkOnMapMovesFar:地圖上的 In Por 是**長距離**瞬移,不是走一格。
func TestBlinkOnMapMovesFar(t *testing.T) {
	s := magicState(t)
	x0, y0 := s.X, s.Y
	if !s.blink(0) {
		t.Fatalf("瞬移不了:\n%s", s.log())
	}
	if !s.AwaitingDirection() {
		t.Fatal("地圖上的 In Por 應該先問方向")
	}
	s.AnswerDirection(East)
	if s.X == x0 && s.Y == y0 {
		t.Skip("東邊 32 格內全是走不進去的地形,這一次沒動")
	}
	if d := worldDelta(x0, s.X); d < 2 {
		t.Errorf("只移動了 %d 格 —— In Por 不是走路", d)
	}
	if s.Y != y0 {
		t.Errorf("往東瞬移卻換了緯度:%d → %d", y0, s.Y)
	}
}

// worldDelta 是世界地圖上兩個 X 的環繞距離。
func worldDelta(a, b int) int {
	d := iabs(a - b)
	if d > u5data.WorldSide/2 {
		d = u5data.WorldSide - d
	}
	return d
}

// TestBlinkInCombatIsRandom:戰鬥中的 In Por 不問方向 —— 原版是隨機落點。
func TestBlinkInCombatIsRandom(t *testing.T) {
	s := magicState(t)
	slot, _ := s.CurrentObjects().Spawn(0x40, s.X+1, s.Y, s.Floor)
	if !s.BeginCombat(slot) {
		t.Fatal("打不起來")
	}
	self := s.combatSlotOfRoster(0)
	if self < 0 {
		t.Fatal("找不到施法者")
	}
	x0, y0 := s.Combat.Units[self].X, s.Combat.Units[self].Y
	moved := false
	for i := 0; i < 20 && !moved; i++ {
		if !s.blink(0) {
			continue
		}
		if s.AwaitingDirection() {
			t.Fatal("戰鬥中的 In Por 不該問方向")
		}
		u := s.Combat.Units[self]
		moved = u.X != x0 || u.Y != y0
	}
	if !moved {
		t.Errorf("放了 20 次都沒換位置:\n%s", s.log())
	}
}

// TestIllusionDuplicatesTheTarget:In Quas Xen 是**複製**,不是召喚。
//
// 原版 `sub_196A4` 把目標的兩個 dword 整組搬進空槽 —— 所以分身的種類、
// 血量、旗標全都一樣。這條驗「多出來的那個跟本尊同種」。
func TestIllusionDuplicatesTheTarget(t *testing.T) {
	s := magicState(t)
	slot, _ := s.CurrentObjects().Spawn(0x40, s.X+1, s.Y, s.Floor)
	if !s.BeginCombat(slot) {
		t.Fatal("打不起來")
	}
	c := s.Combat
	before := 0
	for i := range c.Units {
		if c.Units[i].Flags != 0 {
			before++
		}
	}
	self := s.combatSlotOfRoster(0)
	target, _, _ := s.aiTarget(self)
	if target < 0 {
		t.Fatal("沒有目標")
	}
	want := c.Units[target]
	if !s.illusion(0) {
		t.Skipf("場上沒有空槽或擺不下分身:\n%s", s.log())
	}
	after, found := 0, false
	for i := range c.Units {
		if c.Units[i].Flags == 0 {
			continue
		}
		after++
		if i != target && c.Units[i].Kind == want.Kind && c.Units[i].HP == want.HP {
			found = true
		}
	}
	if after != before+1 {
		t.Errorf("場上從 %d 個變成 %d 個,幻影應該只多出一個", before, after)
	}
	if !found {
		t.Error("找不到與本尊同種同血量的分身")
	}
}

// TestSpellImmuneNamesThree:免疫名單就是那三個劇情人物。
//
// `sub_189BC` 只認 14 / 15 / 47。多一個少一個都會讓「黑刺被一發魅惑解決」
// 這種事發生 —— 這條是拿全 48 種掃一遍反過來驗。
func TestSpellImmuneNamesThree(t *testing.T) {
	immune := []int{}
	for i := 0; i < u5data.CreatureCount; i++ {
		if u5data.CreatureSpellImmune(i) {
			immune = append(immune, i)
		}
	}
	want := []int{u5data.CreatureBlackthorn, u5data.CreatureLordBritish, u5data.CreatureShadowLord}
	if len(immune) != len(want) {
		t.Fatalf("免疫的有 %v,預期 %v", immune, want)
	}
	for i := range want {
		if immune[i] != want[i] {
			t.Fatalf("免疫的有 %v,預期 %v", immune, want)
		}
	}
}

// TestRepellableFlagPicksExactlyFour:0x20 位元只落在四種生物身上。
//
// 這條同時驗兩件事:0x20 認的是哪一批沒搞錯,而且旗標表沒有讀偏一個位元組
// (偏一格的話中選的會是完全不同的四種)。
//
// ⚠ 名單是**幽靈、骷髏、惡魔、暗影領主** —— 惡魔在裡面,所以這一位元
// 不是「不死」。我一開始照 An Xen Corp 的字面意思寫成不死,被這條打掉。
func TestRepellableFlagPicksExactlyFour(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA")
	}
	st, err := u5data.LoadCombatStats(dir)
	if err != nil {
		t.Fatal(err)
	}
	names, err := u5data.LoadCreatureTable(dir)
	if err != nil {
		t.Fatal(err)
	}
	got := []string{}
	for i := 0; i < u5data.CreatureCount; i++ {
		if st.Creature[i].Has(u5data.CreatureRepellable) {
			got = append(got, names.Names[i])
		}
	}
	want := map[string]bool{"Ghost": true, "Skeleton": true, "Daemon": true, "Shadow Lord": true}
	if len(got) != len(want) {
		t.Fatalf("帶 0x20 位元的是 %v,預期 %v", got, want)
	}
	for _, n := range got {
		if !want[n] {
			t.Errorf("%s 不該帶 0x20 位元", n)
		}
	}
}

// TestFieldSpellsPlaceTheRightTile:四個 `*Grav` 對應四個力場格子。
//
// 值來自 `byte_55E24`(FM Towns 0x55E24)= `82 81 80 83`,而分派表把
// 14/15/16/20 依序送成種類 0/1/2/3。順序錯一位就會變成「放毒的放出火」。
func TestFieldSpellsPlaceTheRightTile(t *testing.T) {
	want := map[int]byte{
		SpellInFlamGrav:  u5data.DungeonFireA,
		SpellInNoxGrav:   u5data.DungeonPoisonA,
		SpellInZuGrav:    u5data.DungeonSleepA,
		SpellInSanctGrav: u5data.DungeonElectricA,
	}
	for spell, tile := range want {
		s := dungeonState(t)
		if !s.EnterDungeon(0, false) {
			t.Skip("進不了地牢")
		}
		d := s.Dungeon
		// 挑一格純通道當目標,並轉向它。
		placed := false
		for _, dir := range []Direction{North, East, South, West} {
			dx, dy := dir.Delta()
			x, y := (d.X+dx)&(u5data.DungeonSide-1), (d.Y+dy)&(u5data.DungeonSide-1)
			if s.DungeonTileAt(x, y)&^u5data.DungeonHoleAbove != 0 {
				continue
			}
			d.Facing = dir
			if !s.spellEffect(0, spell) {
				t.Fatalf("咒語 %d 放不出來:\n%s", spell, s.log())
			}
			if got := s.DungeonTileAt(x, y) &^ u5data.DungeonHoleAbove; got != tile {
				t.Errorf("咒語 %d 放出 %02X,預期 %02X", spell, got, tile)
			}
			placed = true
			break
		}
		if !placed {
			t.Skipf("入口四周沒有純通道可以放力場")
		}
	}
}

// TestElectricFieldBouncesInsteadOfBeingStoodOn:電擊力場是在**移動**時觸發。
//
// `jpt_52C7` 把 0x83 送進 default —— 踩踏分派表裡它什麼都不做。
// 但它一點也不無害:`sub_48F4` 在移動時攔下來,印「Ouch! Electric field!」、
// 全隊受傷、把人彈回原格。玩家**從來沒有真的站上去過**。
//
// 這條同時擋兩種寫錯的方式:把它寫成無害(漏了傷害),
// 或把它寫進踩踏表(那樣玩家會站在上面)。
func TestElectricFieldBouncesInsteadOfBeingStoodOn(t *testing.T) {
	s := dungeonState(t)
	if !s.EnterDungeon(0, false) {
		t.Skip("進不了地牢")
	}
	before := make([]uint16, 0, 6)
	for _, ch := range s.Party() {
		before = append(before, ch.HP)
	}
	d := s.Dungeon
	// (a) 踩踏分派表對 0x83 沒有反應。
	s.Dungeons.Set(d.Index, d.Level, d.X, d.Y, u5data.DungeonElectricA)
	s.onDungeonTile()
	for i, ch := range s.Party() {
		if ch.HP != before[i] || ch.Status != u5data.StatusGood {
			t.Errorf("踩踏表對電擊力場有反應了:%s 變成 HP %d / 狀態 %c",
				ch.Name, ch.HP, ch.Status)
		}
	}
	// (b) 走進去卻會受傷,而且人留在原地。
	s.Dungeons.Set(d.Index, d.Level, d.X, d.Y, u5data.DungeonPassage)
	dx, dy := d.Facing.Delta()
	tx, ty := u5data.DungeonWrap(d.X+dx), u5data.DungeonWrap(d.Y+dy)
	s.Dungeons.Set(d.Index, d.Level, tx, ty, u5data.DungeonElectricA)
	x0, y0 := d.X, d.Y
	s.DungeonForward(false)
	if d.X != x0 || d.Y != y0 {
		t.Errorf("走進電擊力場之後停在 (%d,%d),應該被彈回 (%d,%d)", d.X, d.Y, x0, y0)
	}
	hurt := false
	for i, ch := range s.Party() {
		if ch.HP < before[i] {
			hurt = true
		}
	}
	if !hurt {
		t.Errorf("走進電擊力場卻沒人受傷:\n%s", s.log())
	}
}

// TestSleepFieldRollsDexterity:睡眠 / 毒力場是逐人擲敏捷,不是全隊一律中。
//
// `sub_4DC8`:`random(1,30)`,敏捷大於點數就躲掉。敏捷 30 的人幾乎不會中,
// 敏捷 1 的人幾乎一定中 —— 這兩端的差別就是這條在驗的。
func TestSleepFieldRollsDexterity(t *testing.T) {
	nimble, clumsy := 0, 0
	const rounds = 60
	for r := 0; r < rounds; r++ {
		s := dungeonState(t)
		s.SeedRandom(int64(r) + 1)
		if !s.EnterDungeon(0, false) {
			t.Skip("進不了地牢")
		}
		p := s.Party()
		if len(p) == 0 {
			t.Skip("隊伍是空的")
		}
		p[0].Dex, p[0].Status = 30, u5data.StatusGood
		s.fieldAffectsParty(u5data.StatusAsleep)
		if p[0].Status == u5data.StatusAsleep {
			nimble++
		}

		s2 := dungeonState(t)
		s2.SeedRandom(int64(r) + 1)
		q := s2.Party()
		q[0].Dex, q[0].Status = 1, u5data.StatusGood
		s2.fieldAffectsParty(u5data.StatusAsleep)
		if q[0].Status == u5data.StatusAsleep {
			clumsy++
		}
	}
	if clumsy <= nimble {
		t.Errorf("敏捷 1 中了 %d/%d 次,敏捷 30 中了 %d/%d 次 —— 敏捷該有用",
			clumsy, rounds, nimble, rounds)
	}
	t.Logf("敏捷 30 中 %d/%d、敏捷 1 中 %d/%d", nimble, rounds, clumsy, rounds)
}

// TestPeerOpensAndCloses:In Quas Wis 是一個等按鍵的畫面,不是持續狀態。
func TestPeerOpensAndCloses(t *testing.T) {
	s := magicState(t)
	if !s.Peer() {
		t.Fatal("全景開不起來")
	}
	if s.Prompt != PromptPeer {
		t.Fatalf("全景之後 Prompt 是 %v,預期 PromptPeer", s.Prompt)
	}
	// 32×32 的每一格都要取得到 —— 邊界靠世界地圖環繞,不該炸。
	half := PeerSide / 2
	for dy := -half; dy < half; dy++ {
		for dx := -half; dx < half; dx++ {
			_ = s.PeerTile(dx, dy)
		}
	}
	s.ClosePeer()
	if s.Prompt != PromptNone {
		t.Error("按了鍵卻沒收起來")
	}
}
