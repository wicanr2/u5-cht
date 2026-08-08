package game

import (
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// TestTavernMealIsNotHaggled —— ★ 餐點不議價(`docs/re/93`)。
//
// 這條與 `TestTavernWineIsNotHaggled` 成對:兩項都不議價,而**乾糧會**。
// 三者放一起才看得出「不是全部不議價,也不是只有酒不議價」。
func TestTavernMealIsNotHaggled(t *testing.T) {
	s := shopState(t, 2)
	shop, ok := s.Shops.At(2, u5data.ShopTavern)
	if !ok {
		t.Fatal("地點 2 沒有酒館")
	}
	if !s.openShop(shop) {
		t.Fatal("酒館開不起來")
	}
	unit := s.Shops.Prices.TavernFood[shop.TypeIndex]
	alive := 0
	for _, c := range s.Party() {
		if c.Status != u5data.StatusDead {
			alive++
		}
	}
	s.ShopChoose(rune(s.tavernHotkeys()[TavernMeal]))
	if got := s.Shop.Price; got != unit*alive {
		t.Errorf("一餐報 %d,不議價的話該是 %d(%d × %d)", got, unit*alive, unit, alive)
	}
	// 反對照:議價過的價**不等於**原價,否則這條測試沒有鑑別力。
	if h := u5data.Haggle(unit*alive, s.clerkIntel()); h == unit*alive {
		t.Skipf("這家店的智力剛好讓議價 = 原價(%d)⇒ 這條分不出來", h)
	}
}

// TestMealIsServedOnTheTable —— 點完餐桌上真的出現菜。
//
// 原版 `sub_20F60` 尾段:北邊那一格是空桌(0x95)就改成 0x9B,
// 否則試南邊改成 0x9A,兩邊都不是桌子就什麼都不放。
func TestMealIsServedOnTheTable(t *testing.T) {
	cases := []struct {
		name       string
		north, south byte
		wantN, wantS byte
	}{
		{"只有北邊有桌", TableEmpty, 5, TableServedNorth, 5},
		{"只有南邊有桌", 5, TableEmpty, 5, TableServedSouth},
		{"兩邊都有桌 → 只放北邊", TableEmpty, TableEmpty, TableServedNorth, TableEmpty},
		{"兩邊都沒桌 → 什麼都不放", 5, 5, 5, 5},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := shopState(t, 2)
			s.SetTileAt(s.X, s.Y-1, c.north)
			s.SetTileAt(s.X, s.Y+1, c.south)
			s.serveOnTheTable()
			if got := s.TileAt(s.X, s.Y-1); got != c.wantN {
				t.Errorf("北邊變成 %02X,預期 %02X", got, c.wantN)
			}
			if got := s.TileAt(s.X, s.Y+1); got != c.wantS {
				t.Errorf("南邊變成 %02X,預期 %02X", got, c.wantS)
			}
		})
	}
}

// TestBartenderNagsAtExactlyThreeDrinks —— ★ 剛好第三杯才勸。
//
// ⚠⚠ 原版是 `cmp dword_56E1C, 3; jnz` ⇒ **等於 3**,不是 `>= 3`。
// 所以第四杯之後那一問就再也不出現。這條就是驗那個「看起來像 bug 的原版行為」。
func TestBartenderNagsAtExactlyThreeDrinks(t *testing.T) {
	for drinks := 0; drinks <= 5; drinks++ {
		s := shopState(t, 2)
		shop, ok := s.Shops.At(2, u5data.ShopTavern)
		if !ok {
			t.Fatal("地點 2 沒有酒館")
		}
		if !s.openShop(shop) {
			t.Fatal("酒館開不起來")
		}
		if s.tavernHotkeys()[TavernDrink] == ' ' {
			t.Skip("這家酒館沒有酒單")
		}
		s.Shop.drinks = drinks
		s.ShopChoose(rune(s.tavernHotkeys()[TavernDrink]))
		nagged := s.Shop.Mode == ShopModeTavernNag
		if want := drinks == TavernDrinkNagCount; nagged != want {
			t.Errorf("喝過 %d 杯:勸了 %v,預期 %v", drinks, nagged, want)
		}
	}
}

// TestNagLetsYouThroughOnYes —— 那一問答 Y 進酒單、答 N 回選單。
func TestNagLetsYouThroughOnYes(t *testing.T) {
	for _, c := range []struct {
		key  rune
		want ShopMode
	}{{'y', ShopModeTavernWine}, {'n', ShopModeTavernMenu}} {
		s := shopState(t, 2)
		shop, _ := s.Shops.At(2, u5data.ShopTavern)
		if !s.openShop(shop) {
			t.Fatal("酒館開不起來")
		}
		if s.tavernHotkeys()[TavernDrink] == ' ' {
			t.Skip("這家酒館沒有酒單")
		}
		s.Shop.drinks = TavernDrinkNagCount
		s.ShopChoose(rune(s.tavernHotkeys()[TavernDrink]))
		s.ShopChoose(c.key)
		if s.Shop.Mode != c.want {
			t.Errorf("答 %q 之後模式是 %v,預期 %v", c.key, s.Shop.Mode, c.want)
		}
	}
}

// TestAddressWordAlwaysLooksAtRosterZero —— ★ 原版的死碼照抄。
//
// `sub_20F24` 開頭 `al = 0FFh; and al, al; jz` ⇒ 那個 jz 永遠不跳,
// 所以索引永遠是 0 ⇒ **不管跟誰說話都看名冊第 0 筆的性別**。
func TestAddressWordAlwaysLooksAtRosterZero(t *testing.T) {
	s := shopState(t, 2)
	if len(s.Roster) < 2 {
		t.Skip("名冊太短")
	}
	s.Roster[0].Gender = u5data.GenderMale
	s.Roster[1].Gender = u5data.GenderFemale
	male := s.addressWord()
	// 只改第 1 筆 —— 稱呼不該變。
	s.Roster[1].Gender = u5data.GenderMale
	if s.addressWord() != male {
		t.Error("改了名冊第 1 筆就換了稱呼 —— 原版只看第 0 筆")
	}
	// 改第 0 筆才會變。
	s.Roster[0].Gender = u5data.GenderFemale
	if s.addressWord() == male {
		t.Error("改了名冊第 0 筆稱呼卻沒變")
	}
}

// TestPartyCountWordOnlyCoversTwoToSix —— 原版只認 2..6。
//
// ⚠ 一人隊伍在原版會印成 "gold for the  of ye"(中間是空的)——
// 那是原版的行為,不要「順手」補上「一」。
func TestPartyCountWordOnlyCoversTwoToSix(t *testing.T) {
	for n := 0; n <= 8; n++ {
		got := partyCountWord(n)
		if want := n >= 2 && n <= u5data.CombatPartySlots; (got != "") != want {
			t.Errorf("%d 人回 %q,預期 %s", n, got, map[bool]string{true: "有字", false: "空字串"}[want])
		}
	}
}
