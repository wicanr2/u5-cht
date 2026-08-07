package game

import (
	"strconv"

	"github.com/wicanr2/u5-cht/internal/i18n"
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 酒館
//
// 原版 `sub_216C8`。菜單有四個選項,但**每家酒館的熱鍵字母不一樣** ——
// `byte_57034[酒館]` 選出一套「菜單樣式」(0..3),樣式再決定四個字母:
//
//	樣式 0   M A R C      樣式 1   C W · T
//	樣式 2   B R · H      樣式 3   F S P A
//
// 四欄的功能是固定的,只有字母會變:
//
//	第 1 欄  sub_210D8  一餐    每人單價 × 活著的隊員數 → 存糧 += 人數
//	第 2 欄  sub_21108  酒單    六款固定價,**不議價**
//	第 3 欄  sub_21310  乾糧    單價 × 數量 → 存糧 += 數量
//	第 4 欄  sub_21500  打聽    付錢問八德與地牢的所在(還沒做)
//
// 字母是空白代表這家沒有那個選項。

// TavernChoice 是酒館菜單的四欄。
type TavernChoice int

const (
	// TavernMeal 是當場吃的一餐。
	TavernMeal TavernChoice = iota
	// TavernDrink 是酒單。
	TavernDrink
	// TavernRations 是帶著走的乾糧。
	TavernRations
	// TavernLore 是付錢打聽消息。
	TavernLore
)

// Name 回傳這一欄的中文名。
func (c TavernChoice) Name() string {
	switch c {
	case TavernMeal:
		return "餐點"
	case TavernDrink:
		return "酒"
	case TavernRations:
		return "乾糧"
	case TavernLore:
		return "打聽消息"
	}
	return "?"
}

// ShopModeTavern 系列補在 shop.go 的 ShopMode 之後,見該檔案的常數宣告。

// openTavern 進酒館。
func (s *State) openTavern(shop *u5data.Shop) bool {
	s.Shop = &ShopSession{Shop: shop, Mode: ShopModeTavernMenu}
	s.Prompt = PromptShop
	s.showTavernMenu()
	return true
}

// tavernHotkeys 回傳這家酒館四欄的熱鍵字母(小寫),空白表示沒有這一欄。
func (s *State) tavernHotkeys() [4]byte {
	p := s.Shops.Prices
	style := p.TavernStyle[s.Shop.Shop.TypeIndex]
	keys := u5data.TavernHotkeys[style]
	for i := range keys {
		if keys[i] >= 'A' && keys[i] <= 'Z' {
			keys[i] = keys[i] - 'A' + 'a'
		}
	}
	return keys
}

func (s *State) showTavernMenu() {
	sess := s.Shop
	switch sess.Mode {
	case ShopModeTavernMenu:
		line := ""
		for i, k := range s.tavernHotkeys() {
			if k == ' ' {
				continue
			}
			if line != "" {
				line += "   "
			}
			line += string(rune(k)) + ") " + TavernChoice(i).Name()
		}
		s.Log("  " + line)
		s.Log("「汝要點些什麼?」")
	case ShopModeTavernWine:
		p := s.Shops.Prices
		var line string
		for i, en := range u5data.WineNames {
			n := i18n.Name(en)
			if i > 0 {
				line += "  "
			}
			line += string(rune('a'+i)) + ") " + n + " " + strconv.Itoa(p.Wine[i])
		}
		s.Log("「本店酒單:」")
		s.Log("  " + line)
		s.Log("「汝的選擇?」")
	case ShopModeTavernQty:
		s.Log("「汝要幾份?(1-9)」")
	}
}

// tavernChoose 處理酒館裡按下的鍵。
func (s *State) tavernChoose(r rune) {
	sess := s.Shop
	p := s.Shops.Prices
	idx := sess.Shop.TypeIndex
	switch sess.Mode {
	case ShopModeTavernMenu:
		keys := s.tavernHotkeys()
		which := -1
		for i, k := range keys {
			if k != ' ' && rune(k) == r {
				which = i
				break
			}
		}
		switch TavernChoice(which) {
		case TavernMeal:
			s.quoteMeal()
		case TavernDrink:
			sess.Mode = ShopModeTavernWine
			s.showTavernMenu()
		case TavernRations:
			sess.Mode = ShopModeTavernQty
			s.showTavernMenu()
		case TavernLore:
			// sub_21500:打聽八德與地牢的所在。關鍵字表與人名/地名對照還沒解。
			s.Log("(打聽消息 —— 情報系統尚未實作)")
			s.backToMenu()
		default:
			s.LeaveShop()
		}
	case ShopModeTavernWine:
		i := int(r - 'a')
		if i < 0 || i >= u5data.WineCount {
			s.LeaveShop()
			return
		}
		// 酒**不議價**:原版直接拿 dword_56E44[i] 跟金幣比。
		sess.Choice = ShopItem{Goods: GoodsDrink, ID: i, Name: i18n.Name(u5data.WineNames[i]), Qty: 1}
		sess.Price = p.Wine[i]
		sess.Action = ActionTavern
		s.Log("「好眼光。" + i18n.Name(u5data.WineNames[i]) + ",要 " + strconv.Itoa(sess.Price) + " 金。」")
		s.Log("「汝要嗎?(Y/N)」")
		sess.Mode = ShopModeConfirm
	case ShopModeTavernQty:
		n := int(r - '0')
		if n < 1 || n > 9 {
			s.Log("「哼。」")
			sess.Mode = ShopModeTavernMenu
			s.showTavernMenu()
			return
		}
		sess.Choice = ShopItem{Goods: GoodsFood, ID: 0, Name: "乾糧", Qty: n}
		sess.Price = u5data.Haggle(p.TavernRation[idx]*n, s.clerkIntel())
		sess.Action = ActionTavern
		s.Log("「" + strconv.Itoa(n) + " 份乾糧,共 " + strconv.Itoa(sess.Price) + " 金。」")
		s.Log("「汝要嗎?(Y/N)」")
		sess.Mode = ShopModeConfirm
	}
}

// quoteMeal 報一餐的價。
//
// 原版 `sub_20E6C`:掃過隊伍,**死掉的人不算**,每個活人加一份單價。
// 這一餐同時補的是存糧,不是即時回血。
func (s *State) quoteMeal() {
	sess := s.Shop
	p := s.Shops.Prices
	alive := 0
	for _, c := range s.Party() {
		if c.Status != u5data.StatusDead {
			alive++
		}
	}
	if alive == 0 {
		s.Log("「這裡沒有活人要吃飯。」")
		s.backToMenu()
		return
	}
	sess.Choice = ShopItem{Goods: GoodsFood, ID: 0, Name: "餐點", Qty: alive}
	sess.Price = u5data.Haggle(p.TavernFood[sess.Shop.TypeIndex]*alive, s.clerkIntel())
	sess.Action = ActionTavern
	s.Log("「" + strconv.Itoa(alive) + " 位用餐,共 " + strconv.Itoa(sess.Price) + " 金。」")
	s.Log("「汝要嗎?(Y/N)」")
	sess.Mode = ShopModeConfirm
}

// settleTavern 收下錢之後的結果。
func (s *State) settleTavern() {
	it := s.Shop.Choice
	switch it.Goods {
	case GoodsFood:
		s.Inventory.Food += it.Qty
		if s.Inventory.Food > u5data.GoldLimit {
			s.Inventory.Food = u5data.GoldLimit
		}
		s.Log("存糧增為 " + strconv.Itoa(s.Inventory.Food) + " 份。")
	case GoodsDrink:
		// 原版只把「這趟喝了幾杯」加一(dword_56E1C),沒有其他效果。
		s.Shop.drinks++
		s.Log("「請慢用。」")
	}
}
