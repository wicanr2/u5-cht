package game

import (
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/i18n"
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// shopState 造一個「站在店裡」的狀態:真的資料、真的名冊,但位置直接指定。
func shopState(t *testing.T, location int) *State {
	t.Helper()
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	dict, err := u5data.LoadDictionary(dir)
	if err != nil {
		t.Fatal(err)
	}
	shops, err := u5data.LoadShops(dir, dict)
	if err != nil {
		t.Fatal(err)
	}
	items, err := u5data.LoadItemTable(dir)
	if err != nil {
		t.Fatal(err)
	}
	sv, err := u5data.LoadSave(dir + "/INIT.GAM")
	if err != nil {
		t.Fatal(err)
	}
	scenes, err := u5data.LoadSceneSet(dir)
	if err != nil {
		t.Fatal(err)
	}
	npcs, err := u5data.LoadNPCSet(dir)
	if err != nil {
		t.Fatal(err)
	}
	sur, und, err := u5data.LoadWorldObjects(dir)
	if err != nil {
		t.Fatal(err)
	}
	// 世界地圖一定要載 —— 少了它 TileAt 一律回 0,而 tile 0 不阻擋通行,
	// 於是「附近有沒有陸地」這類判斷全部恆真,測試會靜靜地失去意義。
	chunks, err := u5data.LoadChunks(dir+"/BRIT.DAT", u5data.ChunkSide)
	if err != nil {
		t.Fatal(err)
	}
	ovl, err := os.ReadFile(dir + "/DATA.OVL")
	if err != nil {
		t.Fatal(err)
	}
	index, err := u5data.ReadWorldChunkIndex(ovl)
	if err != nil {
		t.Fatal(err)
	}
	world, err := u5data.BuildWorldMap(chunks, index, 1)
	if err != nil {
		t.Fatal(err)
	}
	// 酒館的打聽消息要這張表 —— 少了它那一欄會走「表未載入」那條路,
	// 於是整條流程的測試會靜靜地測不到東西。
	lore, err := u5data.LoadTavernLore(dir)
	if err != nil {
		t.Fatal(err)
	}
	s := &State{
		Shops: shops, Items: items, Scenes: scenes, World: world, NPCs: npcs,
		Objects: sur, UnderObjects: und, Lore: lore,
		Clock: NewClock(), MaxMessages: 64,
	}
	s.LoadFrom(sv)
	s.Location = location
	return s
}

func (s *State) log() string { return strings.Join(s.Messages, "\n") }

// TestArmouryPurchase 走一次完整的買賣:進店 → 選第一項 → 答 Yes → 檢查扣錢與進貨。
//
// 地點 2 的武具店是第 0 家(Iolo's Bows),第一項是匕首(底價 1)。
func TestArmouryPurchase(t *testing.T) {
	s := shopState(t, 2)
	shop, ok := s.Shops.At(2, u5data.ShopArmoury)
	if !ok {
		t.Fatal("地點 2 沒有武具店")
	}
	if !s.openShop(shop) {
		t.Fatal("武具店開不起來")
	}
	if s.Prompt != PromptShop {
		t.Fatalf("進店後 Prompt 是 %v", s.Prompt)
	}
	// 武具店先問買還是賣(原版 sub_1258C 只收 B / S / 空白)。
	if s.Shop.Mode != ShopModeBuySell {
		t.Fatalf("進武具店停在 %v,預期先問買賣", s.Shop.Mode)
	}
	s.ShopChoose('b')
	if len(s.Shop.Menu) != 7 {
		t.Fatalf("貨架有 %d 項,預期 7", len(s.Shop.Menu))
	}
	if s.Shop.Menu[0].Name != "匕首" {
		t.Fatalf("第一項是 %q,預期「匕首」", s.Shop.Menu[0].Name)
	}

	goldBefore := s.Inventory.Gold
	daggersBefore := s.Inventory.Items[16]
	intel := int(s.Party()[0].Intel)
	want := u5data.Haggle(1, intel)

	s.ShopChoose('a')
	if s.Shop.Mode != ShopModeConfirm {
		t.Fatalf("報價後停在 %v\n%s", s.Shop.Mode, s.log())
	}
	if s.Shop.Price != want {
		t.Errorf("報價 %d,依智力 %d 應為 %d", s.Shop.Price, intel, want)
	}
	s.ShopChoose('y')
	if got := goldBefore - s.Inventory.Gold; got != want {
		t.Errorf("扣了 %d 金,預期 %d", got, want)
	}
	if got := s.Inventory.Items[16] - daggersBefore; got != 1 {
		t.Errorf("匕首多了 %d 把,預期 1", got)
	}
	// 成交後回到貨架,可以繼續買。
	if s.Shop.Mode != ShopModeMenu {
		t.Errorf("成交後停在 %v", s.Shop.Mode)
	}
}

// TestArrowsFillToLimit:買箭矢不是加一,是**直接補滿 99**。
//
// 這是原版 sub_11AF0 對編號 27 / 29 的特例(`cmp edi, 1Bh` / `cmp edi, 1Dh`)。
// 少了這條,玩家得按 99 次才裝滿一壺箭。
func TestArrowsFillToLimit(t *testing.T) {
	s := shopState(t, 2)
	shop, _ := s.Shops.At(2, u5data.ShopArmoury)
	s.openShop(shop)
	s.ShopChoose('b')
	s.Inventory.Gold = 9999
	// 貨架第 4 項是 Arrows(第 0 家:Dagger Sling Bow Arrows …)。
	if s.Shop.Menu[3].ID != u5data.ItemArrows {
		t.Fatalf("第 4 項是 %q,預期 Arrows", s.Shop.Menu[3].Name)
	}
	s.ShopChoose('d')
	s.ShopChoose('y')
	if s.Inventory.Items[u5data.ItemArrows] != u5data.CarryLimit {
		t.Errorf("買一次箭矢後有 %d 支,預期補滿 %d",
			s.Inventory.Items[u5data.ItemArrows], u5data.CarryLimit)
	}
}

// TestCannotAfford:錢不夠時不扣錢、不進貨,而且交易失敗。
func TestCannotAfford(t *testing.T) {
	s := shopState(t, 2)
	shop, _ := s.Shops.At(2, u5data.ShopArmoury)
	s.openShop(shop)
	s.ShopChoose('b')
	s.Inventory.Gold = 0
	before := s.Inventory.Items[u5data.ItemArrows]
	s.ShopChoose('g') // 第 7 項:Magic Bow,底價 800
	s.ShopChoose('y')
	if s.Inventory.Gold != 0 {
		t.Errorf("沒錢卻扣了錢,剩 %d", s.Inventory.Gold)
	}
	if s.Inventory.Items[36] != 0 {
		t.Errorf("沒錢卻拿到了魔法弓")
	}
	if s.Inventory.Items[u5data.ItemArrows] != before {
		t.Errorf("動到了不相干的欄位")
	}
	if !strings.Contains(s.log(), "付不出錢") {
		t.Errorf("沒有拒絕訊息:\n%s", s.log())
	}
}

// TestArmourySell:賣東西用的是另一條公式 —— 買價從兩倍底價往下折,
// 賣價則是 `3 × INT × base / 100 + 1` 從零往上加。
//
// 開局背包裡有 1 面小盾(底價 40),賣掉之後金幣要增加、盾要消失。
func TestArmourySell(t *testing.T) {
	s := shopState(t, 2)
	shop, _ := s.Shops.At(2, u5data.ShopArmoury)
	s.openShop(shop)
	s.ShopChoose('s')
	if s.Shop.Mode != ShopModeSell {
		t.Fatalf("按 S 之後停在 %v\n%s", s.Shop.Mode, s.log())
	}
	// 背包裡有貨、店家又肯收的第一件是小盾(編號 4)。
	first := s.Shop.Menu[0]
	if first.ID != 4 {
		t.Fatalf("待售清單第一件是 %q(編號 %d),預期 Small Shield", first.Name, first.ID)
	}
	intel := int(s.Party()[0].Intel)
	want := u5data.SellValue(first.Base, intel)
	if want >= u5data.Haggle(first.Base, intel) {
		t.Errorf("賣價 %d 不該高於買價 %d", want, u5data.Haggle(first.Base, intel))
	}
	gold := s.Inventory.Gold
	had := s.Inventory.Items[first.ID]
	s.ShopChoose('a')
	if s.Shop.Price != want {
		t.Errorf("開價 %d,依智力 %d 應為 %d", s.Shop.Price, intel, want)
	}
	s.ShopChoose('y')
	if got := s.Inventory.Gold - gold; got != want {
		t.Errorf("收到 %d 金,預期 %d", got, want)
	}
	if s.Inventory.Items[first.ID] != had-1 {
		t.Errorf("賣掉後還有 %d 個", s.Inventory.Items[first.ID])
	}
}

// TestSellMenuExcludesUnsellable:彈藥與底價 0 的東西不會出現在待售清單。
//
// 原版是按了才拒絕(sub_12060 的兩個分支),這裡提前濾掉;
// 兩種東西都不該賣得掉,這一條就是在鎖那個結果。
func TestSellMenuExcludesUnsellable(t *testing.T) {
	s := shopState(t, 2)
	shop, _ := s.Shops.At(2, u5data.ShopArmoury)
	s.openShop(shop)
	s.Inventory.Items[u5data.ItemArrows] = 50
	s.Inventory.Items[u5data.ItemQuarrels] = 50
	s.Inventory.Items[47] = 1 // Ankh:底價 0,店家不收
	s.ShopChoose('s')
	for _, it := range s.Shop.Menu {
		switch it.ID {
		case u5data.ItemArrows, u5data.ItemQuarrels:
			t.Errorf("彈藥 %q 出現在待售清單", it.Name)
		case 47:
			t.Errorf("底價 0 的 %q 出現在待售清單", it.Name)
		}
	}
}

// TestGoldCap:金幣上限 9999(原版 sub_2BBDC 的第三個參數)。
func TestGoldCap(t *testing.T) {
	s := shopState(t, 2)
	shop, _ := s.Shops.At(2, u5data.ShopArmoury)
	s.openShop(shop)
	s.ShopChoose('s')
	s.Inventory.Gold = u5data.GoldLimit - 1
	s.ShopChoose('a')
	s.ShopChoose('y')
	if s.Inventory.Gold != u5data.GoldLimit {
		t.Errorf("賣出後金幣 %d,預期夾在 %d", s.Inventory.Gold, u5data.GoldLimit)
	}
}

// TestReagentPurchaseQuantity:藥草是「一價買一包 N 份」,N 來自另一張表。
func TestReagentPurchaseQuantity(t *testing.T) {
	s := shopState(t, 1)
	shop, ok := s.Shops.At(1, u5data.ShopReagents)
	if !ok {
		t.Fatal("地點 1 沒有藥草鋪")
	}
	if !s.openShop(shop) {
		t.Fatal("藥草鋪開不起來")
	}
	s.Inventory.Gold = 9999
	it := s.Shop.Menu[0]
	if it.Qty < 1 {
		t.Fatalf("第一味 %s 的份數是 %d", it.Name, it.Qty)
	}
	before := s.Inventory.Reagents[it.ID]
	s.ShopChoose('a')
	s.ShopChoose('y')
	if got := s.Inventory.Reagents[it.ID] - before; got != it.Qty {
		t.Errorf("買 %s 拿到 %d 份,預期 %d", it.Name, got, it.Qty)
	}
}

// TestGuildQuantities:公會的鑰匙 3 / 寶石 4 / 火把 5,是原版三個分支的立即數。
func TestGuildQuantities(t *testing.T) {
	s := shopState(t, 8)
	shop, ok := s.Shops.At(8, u5data.ShopGuild)
	if !ok {
		t.Fatal("地點 8 沒有公會")
	}
	if !s.openShop(shop) {
		t.Fatal("公會開不起來")
	}
	s.Inventory.Gold = 9999
	before := [3]int{s.Inventory.Keys, s.Inventory.Gems, s.Inventory.Torches}
	for i, key := range []rune{'a', 'b', 'c'} {
		s.ShopChoose(key)
		s.ShopChoose('y')
		_ = i
	}
	after := [3]int{s.Inventory.Keys, s.Inventory.Gems, s.Inventory.Torches}
	for i, want := range u5data.GuildGoodsQty {
		if got := after[i] - before[i]; got != want {
			t.Errorf("%s 多了 %d,預期 %d", u5data.GuildGoodsNames[i], got, want)
		}
	}
}

// TestHealerNeedsAilment:治療所對沒病的人會回「汝無需此術」,而且不收錢。
func TestHealerNeedsAilment(t *testing.T) {
	s := shopState(t, 5)
	shop, ok := s.Shops.At(5, u5data.ShopHealer)
	if !ok {
		t.Fatal("地點 5 沒有治療所")
	}
	if !s.openShop(shop) {
		t.Fatal("治療所開不起來")
	}
	gold := s.Inventory.Gold
	s.ShopChoose('c') // 解毒
	s.ShopChoose('a') // 第一位隊員(健康)
	if !strings.Contains(s.log(), "無需此術") {
		t.Errorf("健康的人卻能解毒:\n%s", s.log())
	}
	if s.Inventory.Gold != gold {
		t.Errorf("沒治療卻收了錢")
	}
}

// TestHealerCuresPoison:中毒的人解毒後狀態變回 G,並照價付錢。
func TestHealerCuresPoison(t *testing.T) {
	s := shopState(t, 5)
	shop, _ := s.Shops.At(5, u5data.ShopHealer)
	s.openShop(shop)
	s.Roster[0].Status = u5data.StatusPoisoned
	gold := s.Inventory.Gold
	s.ShopChoose('c')
	s.ShopChoose('a')
	if s.Shop.Price != u5data.CurePrice {
		t.Errorf("解毒報價 %d,預期 %d(固定價,不議價)", s.Shop.Price, u5data.CurePrice)
	}
	s.ShopChoose('y')
	if s.Roster[0].Status != u5data.StatusGood {
		t.Errorf("解毒後狀態是 %c", s.Roster[0].Status)
	}
	if gold-s.Inventory.Gold != u5data.CurePrice {
		t.Errorf("扣了 %d 金,預期 %d", gold-s.Inventory.Gold, u5data.CurePrice)
	}
}

// TestHealerCharityLocation:地點 7 的治療所對 100 金以下的服務,付不起也照做。
//
// 原版 sub_12794 的 `cmp byte_3E0A3, 7` —— 這是唯一一條與地點綁定的例外,
// 而復活(200 金)不在其中。
func TestHealerCharityLocation(t *testing.T) {
	s := shopState(t, u5data.CharityLocation)
	shop, ok := s.Shops.At(u5data.CharityLocation, u5data.ShopHealer)
	if !ok {
		t.Fatalf("地點 %d 沒有治療所", u5data.CharityLocation)
	}
	s.openShop(shop)
	s.Roster[0].Status = u5data.StatusPoisoned
	s.Inventory.Gold = 0
	s.ShopChoose('c')
	s.ShopChoose('a')
	s.ShopChoose('y')
	if s.Roster[0].Status != u5data.StatusGood {
		t.Errorf("身無分文卻沒被免費解毒:\n%s", s.log())
	}

	// 同一家店,復活 200 金就不在慈善範圍。
	s2 := shopState(t, u5data.CharityLocation)
	s2.openShop(shop)
	s2.Roster[0].Status = u5data.StatusDead
	s2.Inventory.Gold = 0
	s2.ShopChoose('r')
	s2.ShopChoose('a')
	s2.ShopChoose('y')
	if s2.Roster[0].Status != u5data.StatusDead {
		t.Errorf("身無分文卻被免費復活了")
	}
}

// TestEveryShopTypeOpens:八種店現在都進得去。
//
// 這條取代了先前那個「還沒逆完的店要誠實說明」的守門測試 ——
// 守門的對象都做完了,留著只會變成永遠為真的空測試。
func TestEveryShopTypeOpens(t *testing.T) {
	for ty := u5data.ShopType(0); ty < u5data.ShopTypeCount; ty++ {
		var opened bool
		for loc := 1; loc <= 32 && !opened; loc++ {
			s := shopState(t, loc)
			shop, ok := s.Shops.At(loc, ty)
			if !ok {
				continue
			}
			opened = s.openShop(shop)
		}
		if !opened {
			t.Errorf("%s 一家都開不起來", ty.TypeName())
		}
	}
}

// TestTavernHotkeysVaryByShop:酒館的熱鍵字母每家不同,由菜單樣式決定。
//
// 這是很容易漏掉的一條:寫死成 a/b/c 的話,原版按 M 點餐的酒館就點不到東西。
func TestTavernHotkeysVaryByShop(t *testing.T) {
	seen := map[[4]byte]int{}
	for loc := 1; loc <= 32; loc++ {
		s := shopState(t, loc)
		shop, ok := s.Shops.At(loc, u5data.ShopTavern)
		if !ok {
			continue
		}
		s.openShop(shop)
		seen[s.tavernHotkeys()]++
	}
	if len(seen) < 2 {
		t.Errorf("九家酒館只有 %d 種熱鍵配置,原版有 4 套菜單樣式", len(seen))
	}
}

// TestTavernMealCountsLivingOnly:一餐的價是「每人單價 × 活著的人數」,
// 死人不算(原版 sub_20E6C 跳過狀態 'D')。買完存糧增加。
func TestTavernMealCountsLivingOnly(t *testing.T) {
	s := shopState(t, 2)
	shop, ok := s.Shops.At(2, u5data.ShopTavern)
	if !ok {
		t.Fatal("地點 2 沒有酒館")
	}
	if !s.openShop(shop) {
		t.Fatal("酒館開不起來")
	}
	s.Roster[1].Status = u5data.StatusDead
	alive := s.PartySize - 1
	unit := s.Shops.Prices.TavernFood[shop.TypeIndex]
	want := u5data.Haggle(unit*alive, s.clerkIntel())

	food, gold := s.Inventory.Food, s.Inventory.Gold
	s.ShopChoose(rune(s.tavernHotkeys()[TavernMeal]))
	if s.Shop.Price != want {
		t.Errorf("一餐報價 %d,預期 %d(%d 個活人 × %d)", s.Shop.Price, want, alive, unit)
	}
	s.ShopChoose('y')
	if gold-s.Inventory.Gold != want {
		t.Errorf("扣了 %d 金,預期 %d", gold-s.Inventory.Gold, want)
	}
	if s.Inventory.Food-food != alive {
		t.Errorf("存糧多了 %d 份,預期 %d", s.Inventory.Food-food, alive)
	}
}

// TestTavernWineIsNotHaggled:酒**不議價**。
//
// 原版 sub_21108 直接拿 dword_56E44[i] 跟金幣比,沒有套智力折扣 ——
// 全遊戲的交易只有這一項是這樣。
func TestTavernWineIsNotHaggled(t *testing.T) {
	s := shopState(t, 2)
	shop, _ := s.Shops.At(2, u5data.ShopTavern)
	s.openShop(shop)
	keys := s.tavernHotkeys()
	if keys[TavernDrink] == ' ' {
		t.Skip("這家酒館沒有酒單")
	}
	s.Inventory.Gold = 9999
	s.ShopChoose(rune(keys[TavernDrink]))
	s.ShopChoose('a') // Rose,18 金
	if s.Shop.Price != u5data.WineFirstPrice {
		t.Errorf("Rose 報價 %d,預期 %d(不議價)", s.Shop.Price, u5data.WineFirstPrice)
	}
	gold := s.Inventory.Gold
	s.ShopChoose('y')
	if gold-s.Inventory.Gold != u5data.WineFirstPrice {
		t.Errorf("扣了 %d 金", gold-s.Inventory.Gold)
	}
}

// TestTavernRationsQuantity:乾糧是「單價 × 數量」,數量由玩家按數字鍵給。
func TestTavernRationsQuantity(t *testing.T) {
	s := shopState(t, 2)
	shop, _ := s.Shops.At(2, u5data.ShopTavern)
	s.openShop(shop)
	keys := s.tavernHotkeys()
	if keys[TavernRations] == ' ' {
		t.Skip("這家酒館不賣乾糧")
	}
	s.Inventory.Gold = 9999
	food := s.Inventory.Food
	unit := s.Shops.Prices.TavernRation[shop.TypeIndex]
	s.ShopChoose(rune(keys[TavernRations]))
	s.ShopChoose('5')
	if want := u5data.Haggle(unit*5, s.clerkIntel()); s.Shop.Price != want {
		t.Errorf("5 份乾糧報價 %d,預期 %d", s.Shop.Price, want)
	}
	s.ShopChoose('y')
	if s.Inventory.Food-food != 5 {
		t.Errorf("存糧多了 %d 份,預期 5", s.Inventory.Food-food)
	}
}

// TestInnRestRecovers:住一晚 —— 扣錢、時間跳到早上六點、HP 回滿、法力補到智力。
//
// 法力上限就是智力(原版 sub_21D48 的 `[ebx+0Fh] = [ebx+0Eh]`),
// 吟遊詩人只有一半。戰士沒有法力,所以不動。
func TestInnRestRecovers(t *testing.T) {
	s := shopState(t, 2)
	shop, ok := s.Shops.At(2, u5data.ShopInn)
	if !ok {
		t.Fatal("地點 2 沒有旅店")
	}
	if !s.openShop(shop) {
		t.Fatal("旅店開不起來")
	}
	// 先打傷全隊、掏空法力。
	for _, c := range s.Party() {
		c.HP = 1
		c.MP = 0
	}
	s.Clock.Hour, s.Clock.Minute = 20, 0
	gold := s.Inventory.Gold
	want := u5data.Haggle(s.Shops.Prices.Inn[shop.TypeIndex]*s.PartySize, s.clerkIntel())

	s.ShopChoose('r')
	if s.Shop.Price != want {
		t.Errorf("住宿報價 %d,預期 %d(每人每天 × %d 人再議價)",
			s.Shop.Price, want, s.PartySize)
	}
	s.ShopChoose('y')
	if gold-s.Inventory.Gold != want {
		t.Errorf("扣了 %d 金,預期 %d", gold-s.Inventory.Gold, want)
	}
	if s.Clock.Hour != WakeHour {
		t.Errorf("醒來是 %d 點,預期 %d 點", s.Clock.Hour, WakeHour)
	}
	for _, c := range s.Party() {
		if c.HP != c.MaxHP {
			t.Errorf("%s 醒來 HP %d/%d,沒回滿", c.Name, c.HP, c.MaxHP)
		}
		switch c.Class {
		case 'A', 'M':
			if c.MP != c.Intel {
				t.Errorf("%s(%c)法力 %d,預期等於智力 %d", c.Name, c.Class, c.MP, c.Intel)
			}
		case 'B':
			if c.MP != c.Intel/2 {
				t.Errorf("%s(吟遊詩人)法力 %d,預期智力的一半 %d", c.Name, c.MP, c.Intel/2)
			}
		}
		if c.Status != u5data.StatusGood {
			t.Errorf("%s 醒來狀態是 %c", c.Name, c.Status)
		}
	}
}

// TestSleepKillsThePoisoned:中毒的人睡覺會死。
//
// 這是原版的真規則(sub_21D48 的 `cmp [ebx+0Bh], 'P'` → 改成 'D'、HP 歸 0),
// 不是 bug。少了這條,毒在遊戲裡就只是個無關痛癢的狀態。
func TestSleepKillsThePoisoned(t *testing.T) {
	s := shopState(t, 2)
	shop, _ := s.Shops.At(2, u5data.ShopInn)
	s.openShop(shop)
	s.Roster[1].Status = u5data.StatusPoisoned
	s.ShopChoose('r')
	s.ShopChoose('y')
	if s.Roster[1].Status != u5data.StatusDead {
		t.Errorf("中毒的人睡了一晚還活著(狀態 %c)", s.Roster[1].Status)
	}
	if s.Roster[1].HP != 0 {
		t.Errorf("死了卻還有 %d HP", s.Roster[1].HP)
	}
	if !strings.Contains(s.log(), "毒發身亡") {
		t.Errorf("沒有死訊:\n%s", s.log())
	}
	// 沒中毒的人照常恢復。
	if s.Roster[0].Status != u5data.StatusGood {
		t.Errorf("旁邊的人也出事了(狀態 %c)", s.Roster[0].Status)
	}
}

// TestInnLeaveAndPickUp:寄放同伴 → 隊伍少一人、名冊記下地點;領回 → 補回隊伍。
//
// `CharInnFlag` 存的**不是布林而是地點編號**(sub_22018 寫的是 byte_3E0A3),
// 所以在別的城市的旅店領不到人 —— 這一條也一起驗。
func TestInnLeaveAndPickUp(t *testing.T) {
	s := shopState(t, 2)
	shop, _ := s.Shops.At(2, u5data.ShopInn)
	s.openShop(shop)
	before := s.PartySize
	name := s.Party()[1].Name

	s.ShopChoose('l')
	s.ShopChoose('b') // 第二位隊員
	s.ShopChoose('y')
	if s.PartySize != before-1 {
		t.Fatalf("寄放後隊伍 %d 人,預期 %d\n%s", s.PartySize, before-1, s.log())
	}
	if got := s.guestNames(); len(got) != 1 || got[0] != name {
		t.Fatalf("住宿名冊是 %v,預期只有 %s", got, name)
	}
	// 換一家旅店(地點 3)就找不到這個人。
	s.Location = 3
	if got := s.guestNames(); len(got) != 0 {
		t.Errorf("在別的城市也列出了 %v", got)
	}
	s.Location = 2

	s.ShopChoose('p')
	s.ShopChoose('a')
	s.ShopChoose('y')
	if s.PartySize != before {
		t.Errorf("領回後隊伍 %d 人,預期 %d\n%s", s.PartySize, before, s.log())
	}
	if len(s.guestNames()) != 0 {
		t.Errorf("領回後名冊上還有人")
	}
}

// TestInnRoomsLimit:寄放的人不能超過房間數(原版 sub_21CE4)。
func TestInnRoomsLimit(t *testing.T) {
	s := shopState(t, 2)
	shop, _ := s.Shops.At(2, u5data.ShopInn)
	rooms := s.Shops.Prices.InnRooms[shop.TypeIndex]
	// 先把名冊上非隊伍的人全部塞進這家旅店,塞到客滿。
	n := 0
	for i := s.PartySize; i < len(s.Roster) && n < rooms; i++ {
		if s.Roster[i].Present() {
			s.Roster[i].Raw[u5data.CharInnFlag] = byte(2)
			n++
		}
	}
	if n < rooms {
		t.Skipf("名冊上只有 %d 個閒人,填不滿 %d 間房", n, rooms)
	}
	s.openShop(shop)
	s.ShopChoose('l')
	s.ShopChoose('a')
	if !strings.Contains(s.log(), "沒有空房") {
		t.Errorf("客滿了還收人:\n%s", s.log())
	}
}

// TestShipwrightPrices:造船廠報帆船與小艇的價,兩者都套議價公式。
func TestShipwrightPrices(t *testing.T) {
	s := shopState(t, 3)
	shop, ok := s.Shops.At(3, u5data.ShopShipwright)
	if !ok {
		t.Fatal("地點 3 沒有造船廠")
	}
	if !s.openShop(shop) {
		t.Fatal("造船廠開不起來")
	}
	p := s.Shops.Prices
	i := shop.TypeIndex
	intel := s.clerkIntel()
	for k, want := range []int{
		u5data.Haggle(p.Frigate[i], intel),
		u5data.Haggle(p.Skiff[i], intel),
	} {
		s.ShopChoose(rune('a' + k))
		if s.Shop.Price != want {
			t.Errorf("第 %d 項報價 %d,預期 %d", k, s.Shop.Price, want)
		}
		s.ShopChoose('n')
	}
}

// TestStableSpawnsHorse:買了馬,馬要真的出現在旁邊的格子上。
//
// 原版 `sub_118CC` 依 **南、北、東、西** 的順序找第一個地形是 5/68/69
// 又沒被佔住的鄰格。生成之後那一格就查得到物件 —— 這是物件層接上了的證據。
func TestStableSpawnsHorse(t *testing.T) {
	s := shopState(t, 22) // PAWS 有馬廄
	shop, ok := s.Shops.At(22, u5data.ShopStable)
	if !ok {
		t.Fatal("PAWS 沒有馬廄")
	}
	if err := s.SetScene(22, 0, 8, 19); err != nil {
		t.Fatal(err)
	}
	if !s.openShop(shop) {
		t.Fatal("馬廄開不起來")
	}
	s.Inventory.Gold = 9999
	s.ShopChoose('a')
	s.ShopChoose('y')

	found := false
	for _, o := range s.VisibleObjects() {
		if o.Object.Kind == u5data.TileHorse {
			found = true
			if o.X == s.X && o.Y == s.Y {
				t.Error("馬生在玩家自己那一格")
			}
			dist := abs(o.X-s.X) + abs(o.Y-s.Y)
			if dist != 1 {
				t.Errorf("馬在 (%d,%d),離玩家 %d 格,預期就在隔壁", o.X, o.Y, dist)
			}
		}
	}
	if !found {
		t.Errorf("買了馬卻不在地圖上:\n%s", s.log())
	}
}

// TestSceneEntryClearsObjects:進場景要清空物件槽(原版 sub_1678)。
//
// 在城裡買的馬,離開再回來就不在了 —— 那是原版行為,不是漏做。
func TestSceneEntryClearsObjects(t *testing.T) {
	s := shopState(t, 22)
	if err := s.SetScene(22, 0, 8, 19); err != nil {
		t.Fatal(err)
	}
	if _, ok := s.CurrentObjects().Spawn(u5data.TileHorse, 9, 19, 0); !ok {
		t.Fatal("放不下馬")
	}
	if len(s.VisibleObjects()) != 1 {
		t.Fatalf("放了一匹馬卻看到 %d 個物件", len(s.VisibleObjects()))
	}
	if err := s.SetScene(22, 0, 8, 19); err != nil {
		t.Fatal(err)
	}
	if n := len(s.VisibleObjects()); n != 0 {
		t.Errorf("重新進場景後還有 %d 個物件", n)
	}
}

// TestShipwrightSetsDock:買船不生成物件,只記下停泊座標。
//
// 原版 `sub_218DC` 寫的是 byte_3E165 / byte_3E166 —— 船在碼頭等你。
// 之前我一度把船當成物件槽處理,那是沒有依據的:買船那段程式碼裡
// 根本沒有出現任何 tile 值。
func TestShipwrightSetsDock(t *testing.T) {
	s := shopState(t, 3)
	shop, ok := s.Shops.At(3, u5data.ShopShipwright)
	if !ok {
		t.Fatal("地點 3 沒有造船廠")
	}
	s.openShop(shop)
	s.Inventory.Gold = 9999
	s.ShopChoose('a') // 帆船
	s.ShopChoose('y')
	if !s.HasShip {
		t.Fatal("買了船卻沒記下來")
	}
	p := s.Shops.Prices
	if s.DockX != p.DockX[shop.TypeIndex] || s.DockY != p.DockY[shop.TypeIndex] {
		t.Errorf("停泊座標 (%d,%d),預期 (%d,%d)",
			s.DockX, s.DockY, p.DockX[shop.TypeIndex], p.DockY[shop.TypeIndex])
	}
	if len(s.VisibleObjects()) != 0 {
		t.Error("買船不該生成地圖物件")
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// ⚠⚠ **佔位符的個數必須與原文完全相同。**
//
// 商店對白裡 `#$%&*@^` 七個字元會被代換成店名 / 店主 / 價格 / 物品 / 地名 /
// 時段 / 數量。譯文少打一個就少一個資訊,多打一個會把中文裡本來的字
// 吃掉換成價格 —— 而兩種錯誤在畫面上都只是「這句話怪怪的」,不會報錯。
//
// 這一條逐段比對 194 筆,是譯文品質的主要防線。
func TestShopPlaceholdersMatchTheOriginal(t *testing.T) {
	dir := gameDataDir(t)
	if dir == "" {
		return
	}
	dict, err := u5data.LoadDictionary(dir)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(dir + "/SHOPPE.DAT")
	if err != nil {
		t.Fatal(err)
	}
	count := func(s string) map[rune]int {
		m := map[rune]int{}
		for _, r := range s {
			switch r {
			case '#', '$', '%', '&', '*', '@', '^':
				m[r]++
			}
		}
		return m
	}
	total, translated := 0, 0
	for off := 0; off < len(raw); {
		end := off
		for end < len(raw) && raw[end] != 0 {
			end++
		}
		if end > off {
			en := dict.ExpandDAT(raw[off:end])
			if strings.TrimSpace(en) != "" {
				total++
				if i18n.ShopTranslated(off) {
					translated++
					zh := i18n.Shop(off, en)
					we, wz := count(en), count(zh)
					for r, n := range we {
						if wz[r] != n {
							t.Errorf("位移 %d:原文有 %d 個 %q,譯文有 %d 個\n  EN: %s\n  ZH: %s",
								off, n, string(r), wz[r], en, zh)
						}
					}
					for r, n := range wz {
						if we[r] != n {
							t.Errorf("位移 %d:譯文多了 %d 個 %q(原文 %d 個)\n  ZH: %s",
								off, n, string(r), we[r], zh)
						}
					}
				}
			}
		}
		off = end + 1
	}
	t.Logf("商店對白 %d 段,已翻 %d 段", total, translated)
	if translated != total {
		t.Errorf("還有 %d 段沒翻", total-translated)
	}
}

// 譯文接得上引擎:買賣時顯示的是中文。
func TestShopSpeaksChinese(t *testing.T) {
	dir := gameDataDir(t)
	if dir == "" {
		return
	}
	set, err := u5data.LoadTalkSet(dir)
	if err != nil {
		t.Fatal(err)
	}
	shops, err := u5data.LoadShops(dir, set.Dict)
	if err != nil {
		t.Fatal(err)
	}
	shops.Translate = i18n.Shop
	sh, ok := shops.At(britain, u5data.ShopTavern)
	if !ok {
		t.Skip("不列顛城沒有酒館")
	}
	got := shops.Greeting(sh, 0, 10)
	if !strings.ContainsAny(got, "歡迎你我") {
		t.Errorf("酒館招呼還是英文:%q", got)
	}
	// 店名與店主要被代換掉,不能留佔位符。
	if strings.ContainsAny(got, "#$@") {
		t.Errorf("佔位符沒代換乾淨:%q", got)
	}
}

// 酒館的打聽消息:打關鍵字 → 報價 → 付錢 → 得到「去某地找某人」。
//
// 走完整條原版流程(`sub_21500`),而且用真的資料表。
func TestTavernLoreTellsYouWhereToLook(t *testing.T) {
	s := shopState(t, 2)
	shop, ok := s.Shops.At(2, u5data.ShopTavern)
	if !ok {
		t.Fatal("地點 2 沒有酒館")
	}
	if !s.openShop(shop) {
		t.Fatal("酒館開不起來")
	}
	keys := s.tavernHotkeys()
	if keys[TavernLore] == ' ' {
		t.Skip("這家酒館沒有打聽消息那一欄")
	}
	s.SeedRandom(1)
	s.Inventory.Gold = 999
	s.ShopChoose(rune(keys[TavernLore]))
	if s.Shop.Mode != ShopModeTavernLoreAsk {
		t.Fatalf("按下打聽消息之後停在 %v,預期等打字", s.Shop.Mode)
	}
	// 打「crown」→ 對到 crow 那一題,價格 200。
	for _, r := range "crown" {
		s.ShopChoose(r)
	}
	s.ShopChoose('\r')
	if s.Shop.Mode != ShopModeTavernLoreConfirm {
		t.Fatalf("送出關鍵字之後停在 %v,預期等 Y/N:%s", s.Shop.Mode, s.log())
	}
	if s.Shop.Price != 200 {
		t.Errorf("報價 %d,預期 200", s.Shop.Price)
	}
	gold := s.Inventory.Gold
	s.ShopChoose('y')
	if gold-s.Inventory.Gold != 200 {
		t.Errorf("扣了 %d 金,預期 200", gold-s.Inventory.Gold)
	}
	// 線索裡要出現那一題的人名與地名(有譯名就是譯名)。
	e := s.Lore.Entries[s.Lore.Match("crown")]
	out := s.log()
	if !strings.Contains(out, s.loreWho(e.Who)) {
		t.Errorf("線索裡沒有人名 %q:%s", s.loreWho(e.Who), out)
	}
	if !strings.Contains(out, s.lorePlace(e.Where)) {
		t.Errorf("線索裡沒有地名 %q:%s", s.lorePlace(e.Where), out)
	}
}

// 打不出關鍵字要**回到問題本身**,不是回菜單 —— 原版可以一直猜。
func TestTavernLoreAsksAgainWhenItCannotHelp(t *testing.T) {
	s := shopState(t, 2)
	shop, _ := s.Shops.At(2, u5data.ShopTavern)
	s.openShop(shop)
	keys := s.tavernHotkeys()
	if keys[TavernLore] == ' ' {
		t.Skip("這家酒館沒有打聽消息那一欄")
	}
	s.ShopChoose(rune(keys[TavernLore]))
	for _, r := range "xyzzy" {
		s.ShopChoose(r)
	}
	s.ShopChoose('\r')
	if s.Shop.Mode != ShopModeTavernLoreAsk {
		t.Errorf("猜錯之後停在 %v,預期又問一次", s.Shop.Mode)
	}
}

// 金幣不夠:不扣錢、不給線索。
//
// ⚠ 原版比的是 `jle`,所以**剛好等於價格是付得出來的** —— 這條同時釘住界線。
func TestTavernLoreRefusesWhenYouCannotPay(t *testing.T) {
	s := shopState(t, 2)
	shop, _ := s.Shops.At(2, u5data.ShopTavern)
	s.openShop(shop)
	keys := s.tavernHotkeys()
	if keys[TavernLore] == ' ' {
		t.Skip("這家酒館沒有打聽消息那一欄")
	}
	for _, c := range []struct {
		gold int
		paid bool
	}{{199, false}, {200, true}} {
		s.openShop(shop)
		s.SeedRandom(1)
		s.Inventory.Gold = c.gold
		s.ShopChoose(rune(keys[TavernLore]))
		for _, r := range "crown" {
			s.ShopChoose(r)
		}
		s.ShopChoose('\r')
		s.ShopChoose('y')
		got := c.gold - s.Inventory.Gold
		if c.paid && got != 200 {
			t.Errorf("身上 %d 金卻扣了 %d", c.gold, got)
		}
		if !c.paid && got != 0 {
			t.Errorf("身上 %d 金付不出來卻扣了 %d", c.gold, got)
		}
	}
}
