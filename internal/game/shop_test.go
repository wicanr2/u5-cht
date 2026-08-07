package game

import (
	"os"
	"strings"
	"testing"

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
	s := &State{Shops: shops, Items: items, Clock: NewClock(), MaxMessages: 64}
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
	if s.Shop.Menu[0].Name != "Dagger" {
		t.Fatalf("第一項是 %q,預期 Dagger", s.Shop.Menu[0].Name)
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

// TestUnimplementedShopsSaySo:酒館 / 造船廠 / 旅店還沒逆完,要誠實說明,
// 不能靜默什麼都不做,也不能假裝談成了(CLAUDE.md §3.0)。
func TestUnimplementedShopsSaySo(t *testing.T) {
	for _, ty := range []u5data.ShopType{u5data.ShopTavern, u5data.ShopShipwright, u5data.ShopInn} {
		s := shopState(t, 3)
		shop, ok := s.Shops.At(3, ty)
		if !ok {
			continue
		}
		if s.openShop(shop) {
			t.Errorf("%s 的流程還沒逆完,卻開得起來", ty.TypeName())
		}
	}
}
