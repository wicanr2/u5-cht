package u5data

import (
	"os"
	"testing"
)

func loadPrices(t *testing.T) *PriceTable {
	t.Helper()
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	p, err := LoadPrices(dir)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// TestPriceTableRows:每張價目表的列數必須等於該店種的家數。
//
// 兩邊是**各自獨立**得出的 —— 家數來自 shopEntries(執行檔的 byte_4185C 三張平行表),
// 列數來自 DATA.OVL 的位移常數。位移偏一列以上就對不上,而偏一列正是最容易犯的錯。
func TestPriceTableRows(t *testing.T) {
	loadPrices(t) // 位移錯到讀不出合理值時,ParsePrices 自己會擋
	n := TypeCounts()
	want := map[ShopType]int{
		ShopArmoury:    ArmouryCount,
		ShopTavern:     TavernCount,
		ShopStable:     StableCount,
		ShopShipwright: ShipwrightCount,
		ShopReagents:   ReagentShops,
		ShopGuild:      GuildCount,
		ShopHealer:     HealerCount,
		ShopInn:        InnCount,
	}
	for ty, w := range want {
		if n[ty] != w {
			t.Errorf("%s 有 %d 家,但價目表配 %d 列", ty.TypeName(), n[ty], w)
		}
	}
}

// TestItemPrices 鎖住幾個一眼就知道對不對的價格。
//
// 挑的是表頭、表尾與中段各一,位移只要偏一格就全數落空。
func TestItemPrices(t *testing.T) {
	p := loadPrices(t)
	for _, c := range []struct {
		id    int
		name  string
		price int
	}{
		{0, "Leather Helm", 15},
		{7, "Magic Shield", 2000},
		{16, "Dagger", 1},
		{27, "Arrows", 10},
		{30, "Long Sword", 70},
		{45, "Amulet/Turning", 900},
		{47, "Ankh", 0}, // 不賣的東西價格是 0
	} {
		if p.Item[c.id] != c.price {
			t.Errorf("%s(編號 %d)價格 %d,預期 %d", c.name, c.id, p.Item[c.id], c.price)
		}
	}
}

// TestArmouryStock:9 家武具店各賣 7 件,而且賣的全是真的裝備。
//
// 另外核對「Iolo's Bows」那一家(第 0 家)—— 它的貨架應該以弓箭為主,
// 這是店名與資料互相印證的一條:店名不是這張表讀出來的,對得上就不是巧合。
func TestArmouryStock(t *testing.T) {
	p := loadPrices(t)
	for s := 0; s < ArmouryCount; s++ {
		list := p.ArmouryStockList(s)
		if len(list) != 7 {
			t.Errorf("第 %d 家武具店有 %d 件貨,預期 7", s, len(list))
		}
	}
	first := p.ArmouryStockList(0)
	want := []byte{16, 17, 26, 27, 28, 29, 36} // Dagger Sling Bow Arrows Crossbow Quarrels MagicBow
	if len(first) != len(want) {
		t.Fatalf("第 0 家武具店 %v", first)
	}
	for i := range want {
		if first[i] != want[i] {
			t.Fatalf("第 0 家武具店的貨架是 %v,預期 %v", first, want)
		}
	}
}

// TestReagentShops:五家藥草鋪,有價就有份數,而且沒有一家超過選單能列的 5 味。
//
// 再核對「毒藤與曼陀羅只有三家賣」—— 原版把這兩味邪門藥草限制在特定地點,
// 這是資料語意上的旁證,不是解析出來的性質。
func TestReagentShops(t *testing.T) {
	p := loadPrices(t)
	evil := 0
	for s := 0; s < ReagentShops; s++ {
		list := p.ReagentStockList(s)
		if len(list) == 0 || len(list) > 5 {
			t.Errorf("第 %d 家藥草鋪列了 %d 味", s, len(list))
		}
		if p.ReagentPrice[s][6] > 0 || p.ReagentPrice[s][7] > 0 {
			evil++
		}
	}
	if evil != 3 {
		t.Errorf("賣毒藤 / 曼陀羅的有 %d 家,預期 3", evil)
	}
}

// TestHaggle 鎖住議價公式。
//
// 原版是 `base + base*(100 − 3*INT)/100`,x86 的 idiv 向零捨入。
// 智力 0 付兩倍、智力 33 幾乎等於底價 —— 這條曲線是公式解對的特徵。
func TestHaggle(t *testing.T) {
	for _, c := range []struct{ base, intel, want int }{
		{100, 0, 200},
		{100, 10, 170},
		{100, 33, 101},
		{15, 20, 21}, // 15 + 15*40/100 = 15 + 6
		{2000, 25, 2500},
		{1, 30, 1}, // 1 + 1*10/100 = 1 + 0(整數除法歸零)
	} {
		if got := Haggle(c.base, c.intel); got != c.want {
			t.Errorf("Haggle(%d, %d) = %d,預期 %d", c.base, c.intel, got, c.want)
		}
	}
}

// TestPitchOffsetsResolve:41 件有店在賣的裝備,說詞都讀得出可讀的英文;
// 7 件沒人賣的一句都沒有。
//
// 這條是整組表最強的一致性檢查 —— 貨架、價格、說詞位移是**三個不同位移**讀出來的,
// 能三方對上就不會是位移湊巧。
func TestPitchOffsetsResolve(t *testing.T) {
	s := loadShops(t)
	p := s.Prices
	sold, unsold := 0, 0
	for id := 0; id < ItemCount; id++ {
		text := s.Say(p.ItemPitch[id], &s.Shops[0], 9, p.Item[id])
		if p.ItemPitch[id] != 0 && text != "" {
			sold++
			if !printableASCII(text) {
				t.Errorf("裝備 %d 的說詞不像英文:%q", id, text)
			}
			continue
		}
		unsold++
	}
	if sold != 41 || unsold != 7 {
		t.Errorf("有說詞的 %d 件、沒說詞的 %d 件,預期 41 / 7", sold, unsold)
	}
}

// TestGreetOffsetsFromOVL:問候語位移改成讀 DATA.OVL 之後,值要與先前寫死的一致。
//
// 那組數字是逆向 dword_553CC 得來的,現在換成從玩家的檔案讀 —— 這是回歸保護:
// 位移常數 0x3B3A 若指錯地方,這裡會立刻炸。
func TestGreetOffsetsFromOVL(t *testing.T) {
	p := loadPrices(t)
	want := [ShopTypeCount][4]int{
		{0, 0, 0, 0}, // 武具店走另一條流程,四個位移都是 0
		{3436, 3494, 3545, 3604},
		{5265, 5345, 5395, 5463},
		{5783, 5848, 5904, 5965},
		{6759, 6826, 6885, 6949},
		{8034, 8082, 8145, 8230},
		{8764, 8819, 8875, 8927},
		{9188, 9242, 9308, 9365},
	}
	if p.Greet != want {
		t.Errorf("問候語位移\n讀到 %v\n預期 %v", p.Greet, want)
	}
}

// TestServicePrices 鎖住不在 DATA.OVL、只寫在程式碼裡的那幾個價。
func TestServicePrices(t *testing.T) {
	p := loadPrices(t)
	if p.Inn != [InnCount]int{2, 3, 2, 3, 2, 3} {
		t.Errorf("旅店價 %v,預期 [2 3 2 3 2 3]", p.Inn)
	}
	if p.Stable != [StableCount]int{100, 130, 160, 190} {
		t.Errorf("馬價 %v", p.Stable)
	}
	if p.Frigate != [ShipwrightCount]int{600, 753, 650, 700} {
		t.Errorf("船價 %v", p.Frigate)
	}
	if p.Skiff != [ShipwrightCount]int{200, 175, 125, 100} {
		t.Errorf("小艇價 %v", p.Skiff)
	}
	// 酒單價格在原版有兩份:一份在表裡,一份寫死在選單字串裡("a) Rose.......18")。
	// 兩份對得上,表示表沒讀錯。
	if p.Wine != [WineCount]int{18, 192, 79, 30, 275, 98} {
		t.Errorf("酒單價 %v", p.Wine)
	}
}

// TestInventoryFromSave:背包欄位在存檔裡的位移。
//
// 開局 150 金、2 把鑰匙、4 支火把,背包裡是六名初始角色用得上的雜物 ——
// 位移偏一點就會讀成別的欄位,而「數字仍然像數字」不會自己報錯,
// 所以這裡挑的是**有語意**的值。
func TestInventoryFromSave(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	sv, err := LoadSave(dir + "/INIT.GAM")
	if err != nil {
		t.Fatal(err)
	}
	inv := sv.Inventory
	if inv.Gold != 150 {
		t.Errorf("開局金幣 %d,預期 150", inv.Gold)
	}
	if inv.Keys != 2 || inv.Torches != 4 || inv.Gems != 0 {
		t.Errorf("鑰匙 %d 寶石 %d 火把 %d,預期 2 / 0 / 4", inv.Keys, inv.Gems, inv.Torches)
	}
	want := map[int]int{4: 1, 9: 2, 16: 6, 18: 1, 19: 3} // 小盾/布甲/匕首/棍/火油
	for id, n := range want {
		if inv.Items[id] != n {
			t.Errorf("裝備 %d 有 %d 個,預期 %d", id, inv.Items[id], n)
		}
	}
	if inv.Reagents != [ReagentCount]int{4, 6, 7, 6, 0, 3, 0, 0} {
		t.Errorf("藥草 %v", inv.Reagents)
	}
}
