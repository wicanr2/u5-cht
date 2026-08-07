package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

func combatState(t *testing.T) *State {
	t.Helper()
	s := worldState(t)
	cm, err := u5data.LoadCombatMaps("../../gamedata/BRIT.CBT")
	if err != nil {
		t.Skipf("讀不到 BRIT.CBT:%v", err)
	}
	ct, err := u5data.LoadCreatureTable("../../gamedata")
	if err != nil {
		t.Fatal(err)
	}
	st, err := u5data.LoadCombatStats("../../gamedata")
	if err != nil {
		t.Fatal(err)
	}
	s.CombatMaps, s.Creatures, s.Stats = cm, ct, st
	return s
}

// TestSelectCombatMap 鎖住地形 → 戰鬥地圖的對照(原版 sub_2E58C 的 73-case 跳表)。
func TestSelectCombatMap(t *testing.T) {
	const walk = u5data.VehicleWalk
	const ship = u5data.VehicleShip
	for _, c := range []struct {
		name      string
		kind      int
		terrain   int
		transport byte
		inWorld   bool
		want      int
	}{
		{"權杖", u5data.EnemySceptre, 5, walk, true, u5data.CombatMapSceptre},
		{"在船上遇敵船", u5data.EnemyShip, 1, ship, true, u5data.CombatMapShipShip},
		{"在船上打水戰", 0x40, 1, ship, true, u5data.CombatMapShipSea},
		{"在船上打陸戰", 0x40, 5, ship, true, u5data.CombatMapShipLnd},
		{"陸上遇敵船", u5data.EnemyShip, 5, walk, true, u5data.CombatMapShipVs},
		{"水面", 0x40, 2, walk, true, u5data.CombatMapOpenSea},
		{"水生怪物在陸上也打水戰", 0x80, 5, walk, true, u5data.CombatMapOpenSea},
		{"地形 4", 0x40, 4, walk, true, 1},
		{"地形 5", 0x40, 5, walk, true, 2},
		{"地形 6", 0x40, 6, walk, true, 3},
		{"地形 8", 0x40, 8, walk, true, 3},
		{"地形 7", 0x40, 7, walk, true, 4},
		{"地形 31", 0x40, 31, walk, true, 4},
		{"地形 10", 0x40, 10, walk, true, 5},
		{"地形 13", 0x40, 13, walk, true, 6},
		{"地形 29", 0x40, 29, walk, true, 7},
		{"地形 0x6B", 0x40, 0x6B, walk, true, 7},
		{"地形 68", 0x40, 68, walk, true, 8},
		{"沒對照 + 在大地圖", 0x40, 40, walk, true, 2},
		{"沒對照 + 在場景", 0x40, 40, walk, false, 8},
	} {
		got := u5data.SelectCombatMap(c.kind, c.terrain, c.transport, c.inWorld)
		if got != c.want {
			t.Errorf("%s:選了第 %d 張,預期第 %d 張", c.name, got, c.want)
		}
	}
}

// TestWaterBattleException:0x6A / 0x6B 是 0x60 那一族裡的兩個例外,不算水戰。
func TestWaterBattleException(t *testing.T) {
	for _, tile := range []int{0, 1, 2, 3, 0x60, 0x65, 0x6F} {
		if !u5data.IsWaterBattle(tile) {
			t.Errorf("tile 0x%02X 應該算水戰", tile)
		}
	}
	for _, tile := range []int{4, 5, 0x6A, 0x6B, 0x70} {
		if u5data.IsWaterBattle(tile) {
			t.Errorf("tile 0x%02X 不該算水戰", tile)
		}
	}
}

// activeUnits 收集場上還在的單位(Units 現在是固定 32 槽,大半是空的)。
func activeUnits(c *Combat) []*Combatant {
	var out []*Combatant
	for i := range c.Units {
		if c.Units[i].Active() {
			out = append(out, &c.Units[i])
		}
	}
	return out
}

// TestBumpingMonsterStartsCombat:走進怪物就開打,而且隊員與敵人各就各位。
func TestBumpingMonsterStartsCombat(t *testing.T) {
	s := combatState(t)
	// 在旁邊放一隻怪物(種類碼 ≥ 0x40 才算怪物)。
	if _, ok := s.CurrentObjects().Spawn(0x40, s.X+1, s.Y, s.Floor); !ok {
		t.Fatal("放不下怪物")
	}
	s.Move(East)
	if !s.InCombat() {
		t.Fatalf("撞上怪物卻沒開打:\n%s", s.log())
	}
	if s.Prompt != PromptCombat {
		t.Errorf("戰鬥中 Prompt 是 %v", s.Prompt)
	}
	c := s.Combat
	// 隊員一律排在 0..5,位置要與圖裡的入場點一致。
	m := c.Map
	for i := 0; i < s.PartySize; i++ {
		u := c.Units[i]
		if !u.IsParty() {
			t.Fatalf("第 %d 個單位不是隊員", i)
		}
		if u.X != int(m.PartyX[i]) || u.Y != int(m.PartyY[i]) {
			t.Errorf("隊員 %d 在 (%d,%d),圖裡的入場點是 (%d,%d)",
				i, u.X, u.Y, m.PartyX[i], m.PartyY[i])
		}
	}
	// 敵人從第 6 槽起,第一隻站在圖裡的第 0 個敵方入場點。
	e := c.Units[u5data.CombatPartySlots]
	if e.IsParty() || !e.Active() {
		t.Fatal("第 6 槽應該是一隻活著的敵人")
	}
	if e.X != int(m.EnemyX[0]) || e.Y != int(m.EnemyY[0]) {
		t.Errorf("敵人在 (%d,%d),圖裡的入場點是 (%d,%d)", e.X, e.Y, m.EnemyX[0], m.EnemyY[0])
	}
	// 敵人名字要查得出來(種類 0x40 = Mage → 法師)。
	if c.EnemyName != "法師" {
		t.Errorf("敵人名字是 %q,預期「法師」", c.EnemyName)
	}
}

// TestEnemyGroupSizeComesFromTheTable:出現幾隻要落在生物屬性表的上限裡。
//
// 法師(索引 0)的上限是 3,所以擲出來一定是 1..3;史萊姆(24)的 16
// 是原版直接用的固定值,不擲骰 —— 兩種路徑各驗一次。
func TestEnemyGroupSizeComesFromTheTable(t *testing.T) {
	s := combatState(t)
	count := func(kind byte) int {
		s.Combat = nil
		objs := s.CurrentObjects()
		slot, ok := objs.Spawn(kind, s.X+1, s.Y, s.Floor)
		if !ok {
			t.Fatal("放不下怪物")
		}
		if !s.BeginCombat(slot) {
			t.Fatal("打不起來")
		}
		n := 0
		for i := u5data.CombatPartySlots; i < CombatUnitSlots; i++ {
			if s.Combat.Units[i].Flags != 0 {
				n++
			}
		}
		objs.Remove(slot)
		return n
	}
	for try := 0; try < 20; try++ {
		if n := count(0x40); n < 1 || n > 3 {
			t.Fatalf("法師出現 %d 隻,表裡的上限是 3", n)
		}
	}
	// 史萊姆 = 生物索引 24 → 種類碼 64 + 24*4 = 160。
	if n := count(160); n != 16 {
		t.Errorf("史萊姆出現 %d 隻,預期固定 16", n)
	}
}

// TestTownFightIsOneOnOne:在城鎮裡動手只打得到眼前那一個。
//
// 原版 `sub_2F0EC` 的條件是「地點編號 1..32、而且不是衛兵」。
// 衛兵是唯一的例外 —— 叫來一整隊。
func TestTownFightIsOneOnOne(t *testing.T) {
	s := combatState(t)
	s.Location = 2 // 不列顛城
	spawn := func(kind byte) int {
		s.Combat = nil
		objs := s.CurrentObjects()
		slot, ok := objs.Spawn(kind, s.X+1, s.Y, s.Floor)
		if !ok {
			t.Fatal("放不下怪物")
		}
		if !s.BeginCombat(slot) {
			t.Fatal("打不起來")
		}
		n := 0
		for i := u5data.CombatPartySlots; i < CombatUnitSlots; i++ {
			if s.Combat.Units[i].Flags != 0 {
				n++
			}
		}
		objs.Remove(slot)
		return n
	}
	// 史萊姆在野外是 16 隻,在城裡只有 1 隻。
	if n := spawn(160); n != 1 {
		t.Errorf("城裡的史萊姆出現 %d 隻,預期 1", n)
	}
	// 衛兵(索引 12 → 種類碼 112)是例外,整隊 8 個。
	if n := spawn(112); n != 8 {
		t.Errorf("城裡的衛兵出現 %d 個,預期 8", n)
	}
}

// TestInitiativeFavoursDexterity:敏捷高的出手比較密。
//
// 原版的行動倒數是 `36 − 敏捷`,而不是輪流。蝙蝠(敏捷 30)每 6 輪動一次、
// 史萊姆(敏捷 6)每 30 輪 —— 差五倍。這條把「排程不是 round-robin」釘住。
func TestInitiativeFavoursDexterity(t *testing.T) {
	fast := Combatant{Dex: 30}
	slow := Combatant{Dex: 6}
	if fast.resetInit() >= slow.resetInit() {
		t.Errorf("敏捷 30 的倒數 %d 不短於敏捷 6 的 %d",
			fast.resetInit(), slow.resetInit())
	}
	if got := fast.resetInit(); got != 6 {
		t.Errorf("敏捷 30 的倒數是 %d,預期 36 − 30 = 6", got)
	}
	// 敏捷 ≥ 36 不能讓倒數變成 0 或負數,否則排程會在同一個單位上打轉。
	if got := (&Combatant{Dex: 99}).resetInit(); got < 1 {
		t.Errorf("敏捷 99 的倒數是 %d,不該 < 1", got)
	}
}

// TestCombatDistanceIsFlooredEuclidean:距離是 floor(√(dx²+dy²))。
//
// 斜角相鄰要算 1 格(才打得到),隔一格的斜角是 2。
func TestCombatDistanceIsFlooredEuclidean(t *testing.T) {
	cases := []struct{ x1, y1, x2, y2, want int }{
		{0, 0, 0, 0, 0},
		{0, 0, 1, 0, 1},
		{0, 0, 1, 1, 1}, // √2 = 1.41 → 1
		{0, 0, 2, 2, 2}, // √8 = 2.83 → 2
		{0, 0, 3, 4, 5}, // 3-4-5
		{5, 5, 0, 0, 7}, // √50 = 7.07 → 7
	}
	for _, c := range cases {
		if got := combatDistance(c.x1, c.y1, c.x2, c.y2); got != c.want {
			t.Errorf("(%d,%d)→(%d,%d) 距離 %d,預期 %d",
				c.x1, c.y1, c.x2, c.y2, got, c.want)
		}
	}
}

// TestSadujIsAlwaysHostile:名字第 5 個字母是 j 的隊員永遠站在敵方。
//
// 這是原版寫死的一條(`cmp byte_3DDB8[角色*32], 'j'`),要認的只有 Saduj。
func TestSadujIsAlwaysHostile(t *testing.T) {
	u := Combatant{Flags: UnitParty}
	if u.Hostile("Iolo") {
		t.Error("Iolo 不該是敵方")
	}
	if !u.Hostile("Saduj") {
		t.Error("Saduj 該永遠站在敵方")
	}
	// 名字太短的不能越界。
	if u.Hostile("Gorn") {
		t.Error("Gorn 不該是敵方")
	}
	// 死了就不算任何一邊(原版先查 0x20 再查別的)。
	dead := Combatant{Flags: UnitParty | UnitDead}
	if dead.Hostile("Saduj") {
		t.Error("死掉的 Saduj 不該再算敵方")
	}
	// 怪物反過來:陣營反轉位元成立才是我方。
	mon := Combatant{Flags: UnitMonster}
	if !mon.Hostile("") {
		t.Error("怪物預設該是敵方")
	}
	mon.Flags |= UnitSideFlip
	if mon.Hostile("") {
		t.Error("被馴服的怪物該站在我方")
	}
}

// TestPirateNaming:種類碼 < 0x40 的敵人一律叫 PIRATES(原版 sub_2E58C)。
func TestPirateNaming(t *testing.T) {
	s := combatState(t)
	if got := s.enemyDisplayName(u5data.EnemyShip); got != "海盜" {
		t.Errorf("敵船的名字是 %q,預期「海盜」", got)
	}
	if got := s.enemyDisplayName(0x40); got != "法師" {
		t.Errorf("種類 0x40 的名字是 %q,預期「法師」", got)
	}
	// ⚠ 顯示名是中文,但**資料層仍是英文** —— 比對與存檔都靠英文原文。
	if n := s.Creatures.Name(0x40); n != "Mage" {
		t.Errorf("生物名表回 %q,原文應該還是 Mage", n)
	}
}

// TestCombatMoveStaysOnField:戰場上的移動不能走出 11×11,也不能疊在別人身上。
func TestCombatMoveStaysOnField(t *testing.T) {
	s := combatState(t)
	s.CurrentObjects().Spawn(0x40, s.X+1, s.Y, s.Floor)
	s.Move(East)
	if !s.InCombat() {
		t.Skip("沒開打")
	}
	c := s.Combat
	for turn := 0; turn < 200; turn++ {
		for _, d := range []Direction{North, East, South, West} {
			s.CombatMove(d)
		}
		for i := range c.Units {
			u := &c.Units[i]
			if !u.Active() {
				continue
			}
			if u.X < 0 || u.X >= u5data.CombatSide || u.Y < 0 || u.Y >= u5data.CombatSide {
				t.Fatalf("單位 %d 走到 (%d,%d),出了戰場", i, u.X, u.Y)
			}
		}
		seen := map[[2]int]int{}
		for i := range c.Units {
			u := &c.Units[i]
			if !u.Active() {
				continue
			}
			k := [2]int{u.X, u.Y}
			if prev, dup := seen[k]; dup {
				t.Fatalf("單位 %d 與 %d 疊在 (%d,%d)", prev, i, u.X, u.Y)
			}
			seen[k] = i
		}
		if c.Over {
			break
		}
	}
}

// TestFleeLeavesCombat:撤離回到地圖,而且怪物還在(沒打贏就不該消失)。
func TestFleeLeavesCombat(t *testing.T) {
	s := combatState(t)
	s.CurrentObjects().Spawn(0x40, s.X+1, s.Y, s.Floor)
	s.Move(East)
	if !s.InCombat() {
		t.Skip("沒開打")
	}
	s.CombatFlee()
	if s.InCombat() {
		t.Error("撤離之後還在戰鬥")
	}
	if s.Prompt != PromptNone {
		t.Errorf("撤離之後 Prompt 是 %v", s.Prompt)
	}
	if _, _, ok := s.ObjectAt(s.X+1, s.Y); !ok {
		t.Error("沒打贏,怪物卻從地圖上消失了")
	}
	if !strings.Contains(s.log(), "撤離") {
		t.Errorf("沒有撤離訊息:\n%s", s.log())
	}
}

// TestHitThresholdFormula:命中門檻就是 `(防禦 + 30 − 攻擊) / 2`(原版 sub_B484)。
//
// 骰子值域 1..30,所以門檻 ≤ 1 必中、≥ 31 必不中。
func TestHitThresholdFormula(t *testing.T) {
	for _, c := range []struct{ def, atk, want int }{
		{0, 0, 15},
		{10, 10, 15},
		{0, 30, 0},  // 攻擊拉滿 → 門檻 0 → 必中
		{40, 0, 35}, // 防禦很高 → 門檻 35 > 30 → 必不中
		{7, 5, 16},
	} {
		if got := HitThreshold(c.def, c.atk); got != c.want {
			t.Errorf("HitThreshold(%d, %d) = %d,預期 %d", c.def, c.atk, got, c.want)
		}
	}
}

// TestAttackRollRange:命中骰的值域是 1..30,而且 0 會被抬成 1(原版 sub_2B724)。
func TestAttackRollRange(t *testing.T) {
	s := &State{}
	s.SeedRandom(42)
	seen := map[int]bool{}
	for i := 0; i < 5000; i++ {
		r := s.AttackRoll()
		if r < 1 || r > 30 {
			t.Fatalf("擲出 %d,超出 1..30", r)
		}
		seen[r] = true
	}
	if len(seen) < 25 {
		t.Errorf("5000 次只擲出 %d 種值,分布不像 1..30", len(seen))
	}
	if !seen[1] || !seen[30] {
		t.Error("沒擲出過 1 或 30 —— 值域的兩端要涵蓋得到")
	}
}

// TestBluntWeaponUsesStrength:鈍器的命中看力量,其他武器看裝備防禦加總。
//
// 原版 sub_B398:武器類別 == 8 走一條分支(角色紀錄的 0x0C = 力量),
// 否則走另一條(sub_2F2BC = 裝備防禦加總)。
func TestBluntWeaponUsesStrength(t *testing.T) {
	s := combatState(t)
	st, err := u5data.LoadCombatStats("../../gamedata")
	if err != nil {
		t.Fatal(err)
	}
	s.Stats = st
	c := &s.Roster[0]
	c.Strength = 99                 // 力量拉到很高,裝備不變
	blunt := s.attackValueOf(c, 24) // Mace,類別 8
	edged := s.attackValueOf(c, 30) // Long Sword,類別 0
	if blunt != 99 {
		t.Errorf("鈍器的攻擊值是 %d,預期等於力量 99", blunt)
	}
	if edged == 99 {
		t.Errorf("非鈍器不該用力量")
	}
	if edged != st.DefenceOf(c) {
		t.Errorf("非鈍器的攻擊值是 %d,預期等於裝備防禦加總 %d", edged, st.DefenceOf(c))
	}
}

// TestAlwaysHitWeapons:三種武器不用擲骰(原版 sub_B484 的三個 cmp)。
func TestAlwaysHitWeapons(t *testing.T) {
	s := combatState(t)
	st, _ := u5data.LoadCombatStats("../../gamedata")
	s.Stats = st
	s.SeedRandom(7)
	c := &s.Roster[0]
	for w := range AlwaysHitWeapons {
		for i := 0; i < 100; i++ {
			if !s.AttackerHits(c, w, 999) {
				t.Fatalf("武器 0x%02X 應該必中,卻沒中", w)
			}
		}
	}
	// 一般武器打防禦 999 的目標永遠不中。
	for i := 0; i < 100; i++ {
		if s.AttackerHits(c, 30, 999) {
			t.Fatal("長劍打防禦 999 的目標不該命中")
		}
	}
}

// TestWeaponDamageRoll:三把神器的特例,以及一般武器擲 random(1, 上限)。
func TestWeaponDamageRoll(t *testing.T) {
	s := combatState(t)
	st, err := u5data.LoadCombatStats("../../gamedata")
	if err != nil {
		t.Fatal(err)
	}
	s.Stats = st
	s.SeedRandom(3)

	if d, sh := s.WeaponDamageRoll(u5data.ItemGlassSword); d != u5data.InstantKillDamage || !sh {
		t.Errorf("玻璃劍 傷害 %d 碎裂 %v,預期 99 / true", d, sh)
	}
	if d, sh := s.WeaponDamageRoll(u5data.ItemJeweledSword); d != 0 || sh {
		t.Errorf("寶石劍 傷害 %d,預期 0", d)
	}
	if d, _ := s.WeaponDamageRoll(u5data.ItemNone); d != u5data.BareHandDamage {
		t.Errorf("空手傷害 %d,預期 %d", d, u5data.BareHandDamage)
	}
	if d, _ := s.WeaponDamageRoll(u5data.ItemSwordOfChaos); d != u5data.InstantKillDamage {
		t.Errorf("混沌之劍傷害 %d,預期 99(不擲骰)", d)
	}
	// 長劍上限 15:擲很多次應該落在 1..15,而且兩端都碰得到。
	lo, hi := 99, 0
	for i := 0; i < 3000; i++ {
		d, _ := s.WeaponDamageRoll(30)
		if d < 1 || d > 15 {
			t.Fatalf("長劍擲出 %d,超出 1..15", d)
		}
		if d < lo {
			lo = d
		}
		if d > hi {
			hi = d
		}
	}
	if lo != 1 || hi != 15 {
		t.Errorf("長劍的傷害範圍是 %d..%d,預期 1..15(閉區間)", lo, hi)
	}
	// 箭矢上限 1 → 不擲骰,固定 1。
	for i := 0; i < 50; i++ {
		if d, _ := s.WeaponDamageRoll(u5data.ItemArrows); d != 1 {
			t.Fatalf("箭矢傷害 %d,預期固定 1", d)
		}
	}
}

// TestInstantKillIgnoresArmour:必殺(99)不扣防禦。
func TestInstantKillIgnoresArmour(t *testing.T) {
	s := combatState(t)
	st, _ := u5data.LoadCombatStats("../../gamedata")
	s.Stats = st
	s.SeedRandom(11)
	if got := s.DamageToCreature(u5data.InstantKillDamage, 0x48); got != u5data.InstantKillDamage {
		t.Errorf("必殺被扣成 %d", got)
	}
	if got := s.DamageToCharacter(u5data.InstantKillDamage, &s.Roster[0]); got != u5data.InstantKillDamage {
		t.Errorf("必殺被扣成 %d", got)
	}
}

// TestDamageSubtractsResist:減傷擲 random(1, 減傷值),而且值為 0 時不擲。
func TestDamageSubtractsResist(t *testing.T) {
	s := combatState(t)
	st, _ := u5data.LoadCombatStats("../../gamedata")
	s.Stats = st
	s.SeedRandom(5)
	// 法師護甲 0(怪物編號 0x40)→ 傷害原封不動。
	for i := 0; i < 50; i++ {
		if got := s.DamageToCreature(10, 0x40); got != 10 {
			t.Fatalf("打護甲 0 的法師,傷害變成 %d", got)
		}
	}
	// 戰士護甲 8(怪物編號 0x48)→ 傷害落在 10−8 .. 10−1。
	lo, hi := 99, -99
	for i := 0; i < 3000; i++ {
		got := s.DamageToCreature(10, 0x48)
		if got < 2 || got > 9 {
			t.Fatalf("打護甲 8 的戰士,傷害 %d 超出 2..9", got)
		}
		if got < lo {
			lo = got
		}
		if got > hi {
			hi = got
		}
	}
	if lo != 2 || hi != 9 {
		t.Errorf("傷害範圍 %d..%d,預期 2..9", lo, hi)
	}
}

// TestBattleReachesAConclusion:一場仗打得完,而且結局只有勝或敗。
//
// 這是整套回合機制唯一的決定性驗收 —— 排程、AI、傷害、死亡判定任何一環
// 卡住,這條就會逾時而不是安靜地跑出一個看起來還好的畫面。
func TestBattleReachesAConclusion(t *testing.T) {
	for seed := int64(1); seed <= 12; seed++ {
		battleToConclusion(t, seed)
	}
}

// battleToConclusion 用固定種子打一場到結束。
func battleToConclusion(t *testing.T, seed int64) {
	t.Helper()
	s := combatState(t)
	s.SeedRandom(seed)
	slot, ok := s.CurrentObjects().Spawn(0x40, s.X+1, s.Y, s.Floor)
	if !ok {
		t.Fatal("放不下怪物")
	}
	if !s.BeginCombat(slot) {
		t.Fatal("打不起來")
	}
	c := s.Combat
	for act := 0; act < 4000 && !c.Over; act++ {
		if c.Turn < 0 {
			t.Fatalf("種子 %d 第 %d 步輪不到任何人,而勝負也還沒定", seed, act)
		}
		playerActs(s)
	}
	if !c.Over {
		t.Fatalf("種子 %d:4000 步之後還沒打完 —— 回合排程大概卡住了。log:\n%s",
			seed, s.log())
	}
	// 勝負要與場上實況一致。
	enemies, party := s.sideCounts(c)
	if c.Won && enemies != 0 {
		t.Errorf("種子 %d 宣告勝利,場上卻還有 %d 個敵人", seed, enemies)
	}
	if !c.Won && party != 0 {
		t.Errorf("種子 %d 宣告敗北,我方卻還有 %d 個人", seed, party)
	}
}

// playerActs 幫輪到的隊員下一個會**改變狀態**的指令:先找相鄰的敵人打,
// 沒有就往最近的敵人走;被擋住就換方向。回傳 false 代表四個方向都試過了。
func playerActs(s *State) bool {
	c := s.Combat
	turn := c.Turn
	u := &c.Units[turn]
	dirs := []Direction{North, East, South, West}
	// 相鄰有敵人就打。
	for _, d := range dirs {
		dx, dy := d.Delta()
		if t, ok := c.CombatUnitAt(u.X+dx, u.Y+dy); ok && s.hostile(t) != s.hostile(u) {
			s.CombatAttack(d)
			return true
		}
	}
	// 否則朝目標走,擋住就換一個方向。
	_, tdx, tdy := s.aiTarget(turn)
	try := []Direction{}
	if tdx != 0 {
		try = append(try, deltaDirection(tdx, 0))
	}
	if tdy != 0 {
		try = append(try, deltaDirection(0, tdy))
	}
	try = append(try, dirs...)
	for _, d := range try {
		x, y := u.X, u.Y
		s.CombatMove(d)
		if c.Over || c.Turn != turn || u.X != x || u.Y != y {
			return true
		}
	}
	// 四面被堵住 —— 原版的出路是空白鍵(Pass),那也是一個完整的回合。
	s.CombatPass()
	return true
}

// deltaDirection 把一步的位移換成方向(斜角優先取水平)。
func deltaDirection(dx, dy int) Direction {
	switch {
	case dx > 0:
		return East
	case dx < 0:
		return West
	case dy > 0:
		return South
	default:
		return North
	}
}

// TestBattlesGoBothWays:同一種遭遇要有輸有贏。
//
// 這條擋的是兩種都會「全綠但玩壞」的情況:公式寫錯讓怪物變沙包(場場贏)、
// 或遠程規則漏了那個 1/2 讓法師每回合都轟(場場輸)。
// 兩者都不會讓別的測試變紅。
func TestBattlesGoBothWays(t *testing.T) {
	wins := 0
	const runs = 20
	for seed := int64(1); seed <= runs; seed++ {
		s := combatState(t)
		s.SeedRandom(seed)
		slot, _ := s.CurrentObjects().Spawn(0x40, s.X+1, s.Y, s.Floor)
		if !s.BeginCombat(slot) {
			t.Fatal("打不起來")
		}
		c := s.Combat
		for act := 0; act < 4000 && !c.Over; act++ {
			if c.Turn < 0 {
				break
			}
			playerActs(s)
		}
		if c.Won {
			wins++
		}
	}
	if wins == 0 || wins == runs {
		t.Errorf("%d 場對三名法師全部同一個結果(勝 %d)—— 不像在擲骰", runs, wins)
	}
	t.Logf("對法師群 %d 場勝 %d 場", runs, wins)
}

// TestMimicAndReaperNeverMove:擬態怪與收割者釘在原地。
//
// 原版 `sub_AE20` 開頭 `cmp dl, 1Bh / 1Ah` 就直接 return —— 這是牠們
// 「守在寶箱旁 / 長在地上」的設定,不是漏做。
func TestMimicAndReaperNeverMove(t *testing.T) {
	for _, cre := range []int{u5data.CreatureMimicIdx, u5data.CreatureReaperIdx} {
		s := combatState(t)
		s.SeedRandom(3)
		kind := u5data.CreatureBase + byte(cre*4)
		slot, _ := s.CurrentObjects().Spawn(kind, s.X+1, s.Y, s.Floor)
		if !s.BeginCombat(slot) {
			t.Fatal("打不起來")
		}
		c := s.Combat
		u := &c.Units[u5data.CombatPartySlots]
		x, y := u.X, u.Y
		for i := 0; i < 40; i++ {
			s.aiMove(u5data.CombatPartySlots)
		}
		if u.X != x || u.Y != y {
			t.Errorf("生物 %d 從 (%d,%d) 移到 (%d,%d),牠不該動", cre, x, y, u.X, u.Y)
		}
	}
}

// TestGremlinStealsFood:小魔怪命中之後多半是偷糧食而不是造成傷害。
func TestGremlinStealsFood(t *testing.T) {
	s := combatState(t)
	s.SeedRandom(11)
	// 小魔怪 = 生物索引 25 → 種類碼 64 + 100 = 164。
	slot, _ := s.CurrentObjects().Spawn(164, s.X+1, s.Y, s.Floor)
	if !s.BeginCombat(slot) {
		t.Fatal("打不起來")
	}
	s.Inventory.Food = 500
	gremlin := u5data.CombatPartySlots
	if s.Combat.Units[gremlin].Creature != 25 {
		t.Fatalf("第一隻不是小魔怪,是生物 %d", s.Combat.Units[gremlin].Creature)
	}
	before := s.Inventory.Food
	for i := 0; i < 60; i++ {
		s.resolveAttack(gremlin, 0)
	}
	if s.Inventory.Food >= before {
		t.Errorf("被小魔怪打了 60 次,糧食還是 %d(原本 %d)", s.Inventory.Food, before)
	}
}

// TestKillGivesExperience:打死怪物拿的經驗值是「生命上限 / 4 + 1」。
//
// 原版 `sub_B51C` 只看這種怪物有多耐打,與實際打了幾下無關。
func TestKillGivesExperience(t *testing.T) {
	s := combatState(t)
	slot, _ := s.CurrentObjects().Spawn(0x40, s.X+1, s.Y, s.Floor)
	if !s.BeginCombat(slot) {
		t.Fatal("打不起來")
	}
	c := s.Combat
	enemy := u5data.CombatPartySlots
	before := s.Roster[0].Exp
	// 直接送一記必殺,略過命中骰。
	s.applyDamage(0, enemy, u5data.InstantKillDamage)
	if !c.Units[enemy].Dead() {
		t.Fatal("吃了必殺卻沒死")
	}
	want := int(s.Stats.Creature[0].MaxHP)/4 + 1
	if got := int(s.Roster[0].Exp) - int(before); got != want {
		t.Errorf("拿到 %d 點經驗,預期 %d(法師生命上限 %d)",
			got, want, s.Stats.Creature[0].MaxHP)
	}
}

// TestInvulnerableCreaturesTakeNoDamage:旅人 / 黑刺 / 不列顛王打不死。
func TestInvulnerableCreaturesTakeNoDamage(t *testing.T) {
	s := combatState(t)
	// 不列顛王 = 生物索引 15 → 種類碼 64 + 15*4 = 124。
	slot, _ := s.CurrentObjects().Spawn(124, s.X+1, s.Y, s.Floor)
	if !s.BeginCombat(slot) {
		t.Fatal("打不起來")
	}
	c := s.Combat
	enemy := u5data.CombatPartySlots
	hp := c.Units[enemy].HP
	for i := 0; i < 20; i++ {
		s.applyDamage(0, enemy, u5data.InstantKillDamage)
	}
	if c.Units[enemy].Dead() || c.Units[enemy].HP != hp {
		t.Errorf("不列顛王被打掉了(HP %d → %d、死亡 %v)",
			hp, c.Units[enemy].HP, c.Units[enemy].Dead())
	}
}
