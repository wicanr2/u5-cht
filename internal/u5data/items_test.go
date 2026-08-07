package u5data

import (
	"os"
	"testing"
)

func gamedataDir(t *testing.T) string {
	t.Helper()
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	return dir
}

// TestItemTable:48 件裝備,順序由 0x1806 的指標表決定。
//
// 名字分成兩段放在 DATA.OVL(0x52 與 0x175C),只讀其中一段會缺一半 ——
// 而且缺的是長劍、十字弓、鎖甲頭巾這些最常見的。指標表把兩段交錯串起來。
func TestItemTable(t *testing.T) {
	tbl, err := LoadItemTable(gamedataDir(t))
	if err != nil {
		t.Fatal(err)
	}
	// 幾個定位點:分類邊界與兩段字串各出一個。
	for id, want := range map[byte]string{
		0:  "Leather Helm", // 第一段
		1:  "Chain Coif",   // 第二段 —— 交錯的證據
		4:  "Small Shield",
		9:  "Cloth Armour",
		11: "Ring Mail", // 第二段
		16: "Dagger",    // 第二段
		30: "Long Sword",
		41: "Mystic Sword",
		47: "Ankh",
	} {
		if got := tbl.Name(id); got != want {
			t.Errorf("裝備 %d 是 %q,預期 %q", id, got, want)
		}
	}
	for i := 0; i < ItemCount; i++ {
		if tbl.Names[i] == "" {
			t.Errorf("裝備 %d 沒有名字", i)
		}
	}
	if tbl.Name(ItemNone) != "" {
		t.Error("ItemNone 應該回傳空字串")
	}
}

// TestCharacterEquipment:角色紀錄的裝備欄位對得上每個人的定位。
//
// 這是欄位位移正確的橫向佐證:法師沒有頭盔、穿布甲拿匕首;
// 戰士 Geoffrey 是釘盔 + 鱗甲 + 釘頭錘 + 釘盾;聖者是鎖甲頭巾 + 鎖甲 + 長劍 + 安卡。
func TestCharacterEquipment(t *testing.T) {
	dir := gamedataDir(t)
	tbl, err := LoadItemTable(dir)
	if err != nil {
		t.Fatal(err)
	}
	sv, err := LoadSave(dir + "/INIT.GAM")
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		idx                            int
		helm, armour, weapon, shield   string
	}{
		{0, "Chain Coif", "Chain Mail", "Long Sword", ""},           // 聖者
		{1, "Leather Helm", "Ring Mail", "Short Sword", "Small Shield"}, // Shamino
		{2, "Leather Helm", "Leather Armour", "Main Gauche", "Short Sword"}, // Iolo
		{3, "", "Cloth Armour", "Dagger", ""},                       // Mariah(法師)
		{4, "Spiked Helm", "Scale Mail", "Mace", "Spiked Shield"},   // Geoffrey
	}
	for _, c := range cases {
		eq := sv.Roster[c.idx].Equipment()
		got := [4]string{tbl.Name(eq.Helm), tbl.Name(eq.Armour), tbl.Name(eq.Weapon), tbl.Name(eq.Shield)}
		want := [4]string{c.helm, c.armour, c.weapon, c.shield}
		if got != want {
			t.Errorf("%s 的裝備是 %v,預期 %v", sv.Roster[c.idx].Name, got, want)
		}
	}
}

// TestCreatureTable:索引 = (生物編號 − 64) / 4。
func TestCreatureTable(t *testing.T) {
	tbl, err := LoadCreatureTable(gamedataDir(t))
	if err != nil {
		t.Fatal(err)
	}
	for cr, want := range map[byte]string{
		64: "Mage", 68: "Bard", 80: "Villager", 84: "Merchant",
		104: "Child", 112: "Guard", 124: "Lord British",
	} {
		if got := tbl.Name(cr); got != want {
			t.Errorf("生物 %d 是 %q,預期 %q", cr, got, want)
		}
	}
	// 「可被嚇跑的平民」範圍(sub_B98 的 0x40..0x73)剛好排除衛兵與不列顛王。
	if !tbl.IsHuman(0x40) || !tbl.IsHuman(0x68) {
		t.Error("0x40 / 0x68 應該算人")
	}
	if tbl.IsHuman(0x74) || tbl.IsHuman(0x30) {
		t.Error("0x74(怪物)與 0x30(物件)不該算人")
	}
	// 物件與動物(編號 < 64)沒有名字,這是原版資料就這樣。
	if tbl.Name(17) != "" {
		t.Errorf("編號 17 是物件,不該有名字,實得 %q", tbl.Name(17))
	}
}

// TestMostNPCsHaveCreatureNames:全遊戲多數 NPC 都對得上生物名。
// 對不上的應該只有編號 < 64 的物件動物與非 4 倍數的怪物。
func TestMostNPCsHaveCreatureNames(t *testing.T) {
	dir := gamedataDir(t)
	tbl, err := LoadCreatureTable(dir)
	if err != nil {
		t.Fatal(err)
	}
	set, err := LoadNPCSet(dir)
	if err != nil {
		t.Fatal(err)
	}
	named, unnamed := 0, 0
	for n := 1; n <= len(Locations); n++ {
		npcs, err := set.At(n)
		if err != nil {
			t.Fatal(err)
		}
		for i := 1; i < NPCsPerLocation; i++ {
			c := npcs[i].Creature
			if c == 0 {
				continue
			}
			if tbl.Name(c) != "" {
				named++
			} else {
				unnamed++
				if c >= CreatureBase && (c-CreatureBase)%4 == 0 {
					t.Errorf("生物 %d 是 4 的倍數且 >= 64,卻查不到名字", c)
				}
			}
		}
	}
	if named < 250 {
		t.Errorf("只有 %d 個 NPC 查得到名字,太少", named)
	}
	t.Logf("全遊戲 NPC:%d 個有名字、%d 個沒有(物件與怪物)", named, unnamed)
}
