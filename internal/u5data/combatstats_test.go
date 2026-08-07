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
