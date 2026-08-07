package u5data

import (
	"os"
	"testing"
)

func loadStats(t *testing.T) *CombatStats {
	t.Helper()
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	s, err := LoadCombatStats(dir)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// TestCreatureStatsMatchArchetype:怪物三圍要符合各自的定位。
//
// 四個原型同時對上才算數 —— 位移偏一格就會兩三個一起垮:
// 法師智力最高、戰士力量最高、吟遊詩人敏捷最高、聖者三圍相等。
func TestCreatureStatsMatchArchetype(t *testing.T) {
	s := loadStats(t)
	mage, bard, fighter, avatar := &s.Creature[0], &s.Creature[1], &s.Creature[2], &s.Creature[3]
	if mage.Intel != 20 || mage.Strength != 10 {
		t.Errorf("法師 力%d 敏%d 智%d,預期力 10 智 20", mage.Strength, mage.Dex, mage.Intel)
	}
	if fighter.Strength != 20 || fighter.Intel != 10 {
		t.Errorf("戰士 力%d 智%d,預期力 20 智 10", fighter.Strength, fighter.Intel)
	}
	if bard.Dex != 20 {
		t.Errorf("吟遊詩人 敏%d,預期 20", bard.Dex)
	}
	if avatar.Strength != avatar.Dex || avatar.Dex != avatar.Intel {
		t.Errorf("聖者 力%d 敏%d 智%d,預期三圍相等", avatar.Strength, avatar.Dex, avatar.Intel)
	}
}

// TestItemDefenceValues:裝備防禦值的語意。
//
// 頭盔 1/2/3/3、盾 2/3/3/5/0、護甲 1→10 遞增、武器幾乎都是 0、
// 防護戒指 2。這一整組同時對上,位移就不可能是湊巧。
func TestItemDefenceValues(t *testing.T) {
	s := loadStats(t)
	for _, c := range []struct {
		id   int
		name string
		want int
	}{
		{0, "Leather Helm", 1}, {1, "Chain Coif", 2}, {2, "Iron Helm", 3},
		{4, "Small Shield", 2}, {5, "Large Shield", 3}, {7, "Magic Shield", 5},
		{9, "Cloth Armour", 1}, {14, "Plate Mail", 7}, {15, "Mystic Armour", 10},
		{30, "Long Sword", 0}, {43, "Ring of Protection", 2}, {46, "Spiked Collar", 2},
	} {
		if s.ItemDefence[c.id] != c.want {
			t.Errorf("%s(%d)防禦 %d,預期 %d", c.name, c.id, s.ItemDefence[c.id], c.want)
		}
	}
}

// TestItemRangeValues:遠程武器的射程,近戰一律 0。
func TestItemRangeValues(t *testing.T) {
	s := loadStats(t)
	for _, c := range []struct {
		id   int
		name string
		want int
	}{
		{16, "Dagger", 3}, {17, "Sling", 4}, {19, "Flaming Oil", 4},
		{21, "Spear", 5}, {22, "Throwing Axe", 4},
		{26, "Bow", 7}, {28, "Crossbow", 8},
		{36, "Magic Bow", 15}, {38, "Magic Axe", 15},
		{30, "Long Sword", 0}, {33, "2H Sword", 0}, {0, "Leather Helm", 0},
	} {
		if s.ItemRange[c.id] != c.want {
			t.Errorf("%s(%d)射程 %d,預期 %d", c.name, c.id, s.ItemRange[c.id], c.want)
		}
	}
}

// TestBluntWeapons:類別 8 是鈍器 —— 釘盔、釘盾、棍、釘頭錘、雙手錘。
//
// 這一組共同點很明顯(都是砸的),而且它決定命中要看力量還是別的,
// 所以值得單獨釘住。
func TestBluntWeapons(t *testing.T) {
	s := loadStats(t)
	blunt := []int{}
	for i := 0; i < ItemCount; i++ {
		if s.ItemKind[i] == ItemKindBlunt {
			blunt = append(blunt, i)
		}
	}
	want := []int{3, 6, 18, 24, 31} // 釘盔 釘盾 棍 釘頭錘 雙手錘
	if len(blunt) != len(want) {
		t.Fatalf("鈍器有 %v,預期 %v", blunt, want)
	}
	for i := range want {
		if blunt[i] != want[i] {
			t.Fatalf("鈍器是 %v,預期 %v", blunt, want)
		}
	}
}

// TestDefenceOfAvatar:聖者的裝備防禦加總。
//
// 鎖甲頭巾 2 + 鎖甲 5 + 長劍 0 + 安卡 0 = 7。
func TestDefenceOfAvatar(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	s := loadStats(t)
	sv, err := LoadSave(dir + "/INIT.GAM")
	if err != nil {
		t.Fatal(err)
	}
	if got := s.DefenceOf(&sv.Roster[0]); got != 7 {
		t.Errorf("聖者的裝備防禦是 %d,預期 7(鎖甲頭巾 2 + 鎖甲 5)", got)
	}
}

// TestWeaponDamageValues:武器傷害表的語意。
//
// 匕首 6 → 長劍 15 → 雙手武器 20 → 戟 30 遞增,兩把神劍 99(必殺),
// 箭矢與弩矢只有 1(彈藥本身不造成傷害,傷害來自弓),防具一律 0。
func TestWeaponDamageValues(t *testing.T) {
	s := loadStats(t)
	for _, c := range []struct {
		id   int
		name string
		want int
	}{
		{16, "Dagger", 6}, {21, "Spear", 10}, {23, "Short Sword", 12},
		{24, "Mace", 15}, {30, "Long Sword", 15},
		{31, "2H Hammer", 20}, {33, "2H Sword", 20}, {34, "Halberd", 30},
		{35, "Sword of Chaos", 99}, {39, "Glass Sword", 99},
		{27, "Arrows", 1}, {29, "Quarrels", 1},
		{0, "Leather Helm", 0}, {14, "Plate Mail", 0},
		// ⚠ 釘盔與釘盾**有**傷害 —— 它們是鈍器,拿來撞人打得痛。
		{3, "Spiked Helm", 4}, {6, "Spiked Shield", 6},
	} {
		if s.ItemDamage[c.id] != c.want {
			t.Errorf("%s(%d)傷害 %d,預期 %d", c.name, c.id, s.ItemDamage[c.id], c.want)
		}
	}
}

// TestCreatureArmourAndAttack:怪物屬性的 +3 護甲與 +4 攻擊力。
//
// 戰士護甲 8 > 法師 0(法師本來就脆),聖者攻擊 30 最高。
func TestCreatureArmourAndAttack(t *testing.T) {
	s := loadStats(t)
	mage, fighter, avatar := &s.Creature[0], &s.Creature[2], &s.Creature[3]
	if mage.Armour != 0 {
		t.Errorf("法師護甲 %d,預期 0", mage.Armour)
	}
	if fighter.Armour != 8 {
		t.Errorf("戰士護甲 %d,預期 8", fighter.Armour)
	}
	if avatar.Attack != 30 {
		t.Errorf("聖者攻擊 %d,預期 30", avatar.Attack)
	}
	if fighter.Armour <= mage.Armour {
		t.Error("戰士的護甲該高於法師")
	}
}

// TestCreatureGroupSizeMatchesArchetype:一次遭遇幾隻該符合各自的定位。
//
// 成群的小怪 16、獨行的大怪 1 —— 兩端同時對上,位移偏一格就會一起垮。
func TestCreatureGroupSizeMatchesArchetype(t *testing.T) {
	s := loadStats(t)
	swarm := map[int]string{21: "蝙蝠", 24: "史萊姆"}
	for i, name := range swarm {
		if g := s.Creature[i].GroupMax; g != 16 {
			t.Errorf("%s 一次最多 %d 隻,預期 16", name, g)
		}
	}
	lone := map[int]string{14: "黑刺", 15: "不列顛王", 26: "擬態怪", 40: "巨蟒陷阱"}
	for i, name := range lone {
		if g := s.Creature[i].GroupMax; g != 1 {
			t.Errorf("%s 一次最多 %d 隻,預期獨行(1)", name, g)
		}
	}
	if d, b := s.Creature[39].GroupMax, s.Creature[21].GroupMax; d >= b {
		t.Errorf("巨龍 %d 隻不該多於蝙蝠 %d 隻", d, b)
	}
}

// TestCreatureMaxHPOrder:生命上限要跟強弱一致。
func TestCreatureMaxHPOrder(t *testing.T) {
	s := loadStats(t)
	bat, slime, dragon := s.Creature[21].MaxHP, s.Creature[24].MaxHP, s.Creature[39].MaxHP
	if !(bat < slime && slime < dragon) {
		t.Errorf("生命上限 蝙蝠%d < 史萊姆%d < 巨龍%d 不成立", bat, slime, dragon)
	}
	if dragon != 99 {
		t.Errorf("巨龍的生命上限是 %d,預期 99", dragon)
	}
}

// TestCreatureRangeSplitsMeleeFromRanged:射程表要把近戰與遠程分乾淨。
//
// 近戰的一律 1、遠程的一律 > 1 —— 這是「射程 1 的那一批對應的 0x15DC 欄位
// 也全是 0」的另一面,兩張表互相佐證。
func TestCreatureRangeSplitsMeleeFromRanged(t *testing.T) {
	s := loadStats(t)
	melee := []int{2, 3, 4, 19, 20, 21, 22, 23, 24, 25, 29, 31, 32, 33, 36, 40}
	for _, i := range melee {
		if r := s.Creature[i].Range; r != 1 {
			t.Errorf("生物 %d 的射程是 %d,預期近戰 1", i, r)
		}
		if u := s.CreatureMissile[i]; u != 0 {
			t.Errorf("生物 %d 是近戰,投射物圖號卻是 %d(預期 0)", i, u)
		}
	}
	ranged := map[int]byte{0: 7, 12: 15, 28: 5, 39: 9}
	for i, want := range ranged {
		if r := s.Creature[i].Range; r != want {
			t.Errorf("生物 %d 的射程是 %d,預期 %d", i, r, want)
		}
		if s.CreatureMissile[i] == 0 {
			t.Errorf("生物 %d 有射程,投射物圖號卻是 0", i)
		}
	}
}

// TestCreatureFlagsPickOutTheRightCreatures:每個旗標指到的那批生物要說得通。
//
// 這是旗標表最強的驗收 —— 位移只要偏一格,「只有小魔怪會偷食物」
// 「只有那三位殺不死」這些條件會同時垮掉。
func TestCreatureFlagsPickOutTheRightCreatures(t *testing.T) {
	s := loadStats(t)
	with := func(flag uint16) []int {
		var got []int
		for i := range s.Creature {
			if s.Creature[i].Has(flag) {
				got = append(got, i)
			}
		}
		return got
	}
	eq := func(name string, flag uint16, want ...int) {
		t.Helper()
		got := with(flag)
		if len(got) != len(want) {
			t.Errorf("%s:%v,預期 %v", name, got, want)
			return
		}
		for i := range got {
			if got[i] != want[i] {
				t.Errorf("%s:%v,預期 %v", name, got, want)
				return
			}
		}
	}
	// 小魔怪偷食物,而且全表只有牠。
	eq("偷食物", CreatureStealsFood, 25)
	// 殺不死的三位 = 旅人 / 黑刺 / 不列顛王。
	eq("無敵", CreatureInvulnerable, 13, 14, 15)
	// 死了會消失的多一位暗影領主(47)。
	eq("消失", CreatureVanishes, 13, 14, 15, 47)
	// 下毒:巨蜘蛛 / 大烏賊 / 巨蟒(0x0004)+ 巨鼠 / 擬態怪(0x0200)。
	var poison []int
	for i := range s.Creature {
		if s.Creature[i].IsPoisonous() {
			poison = append(poison, i)
		}
	}
	for _, i := range []int{17, 20, 22, 26, 34} {
		found := false
		for _, p := range poison {
			if p == i {
				found = true
			}
		}
		if !found {
			t.Errorf("生物 %d 該會下毒,卻不在 %v 裡", i, poison)
		}
	}
	// 史萊姆會分裂。⚠ 不能寫「只有史萊姆」—— 石像鬼(30)的旗標也有這一位,
	// 原因不明,但那是原版資料寫的。
	if !s.Creature[24].Has(CreatureDivides) {
		t.Error("史萊姆不會分裂?")
	}
	// 鬼火會瞬移。
	if !s.Creature[37].Has(CreatureTeleports) {
		t.Error("鬼火不會瞬移?")
	}
	// 會施法的一定包含法師與注視者,而戰士一定不會。
	for _, i := range []int{0, 28} {
		if !s.Creature[i].Has(CreatureCasts) {
			t.Errorf("生物 %d 該會施法", i)
		}
	}
	if s.Creature[2].Has(CreatureCasts) {
		t.Error("戰士不該會施法")
	}
}

// TestCreatureMixIsSelfConsistent:混編表指到的一定是合法索引。
func TestCreatureMixIsSelfConsistent(t *testing.T) {
	s := loadStats(t)
	for i, m := range s.CreatureMix {
		if int(m) >= CreatureCount {
			t.Errorf("生物 %d 的混編對象是 %d,超出 0..%d", i, m, CreatureCount-1)
		}
	}
	// 幾個明顯是「同類的頭目」的對應:小魔怪 → 擬態怪、擬態怪 → 巨人。
	if s.CreatureMix[25] != 26 || s.CreatureMix[26] != 35 {
		t.Errorf("混編表:25→%d、26→%d,預期 26 與 35", s.CreatureMix[25], s.CreatureMix[26])
	}
	// 殺不死的三位只會混到自己(不會帶小弟)。
	for _, i := range []int{13, 14, 15} {
		if int(s.CreatureMix[i]) != i {
			t.Errorf("生物 %d 的混編對象是 %d,預期自己", i, s.CreatureMix[i])
		}
	}
}
