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
			// ★ 原版 `sub_21108` 開頭:`cmp dword_56E1C, 3; jnz` ——
			// **剛好第三杯**時店主先問一句,第四杯之後就不問了(`docs/re/93`)。
			if sess.drinks == TavernDrinkNagCount {
				sess.Mode = ShopModeTavernNag
				s.Log("「" + s.addressWord() + ",恕吾直言,」" +
					sess.Shop.Owner + "說道。「汝喝得還不夠麼?」")
				s.Log("(Y/N)")
				return
			}
			sess.Mode = ShopModeTavernWine
			s.showTavernMenu()
		case TavernRations:
			sess.Mode = ShopModeTavernQty
			s.showTavernMenu()
		case TavernLore:
			s.askTavernLore()
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
	case ShopModeTavernNag:
		// ⚠ 原版收的是 Y / N:答 Y 才進酒單,答 N 回到選單。
		if r == 'y' || r == 'Y' {
			sess.Mode = ShopModeTavernWine
			s.showTavernMenu()
			return
		}
		s.Log("否。")
		sess.Mode = ShopModeTavernMenu
		s.showTavernMenu()
	case ShopModeTavernLoreConfirm:
		s.answerTavernLore(r)
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
	sess.Choice = ShopItem{Goods: GoodsFood, ID: 0, Name: "餐點", Qty: alive, ServeOnTable: true}
	// ⚠⚠ **餐點不議價**(`docs/re/93`)。此前這裡套了 `u5data.Haggle`,
	// 而全檔掃描議價公式(`mov ecx, 64h` + `sub ecx, edx`)只在**九支**函式裡出現,
	// 餐點路徑三支(`sub_210D8` / `sub_20E6C` / `sub_20F60`)一支都不在 ——
	// 乾糧(`sub_21310`)在、餐點不在。`docs/re/10` 那一列也一併更正了。
	sess.Price = p.TavernFood[sess.Shop.TypeIndex] * alive
	sess.Action = ActionTavern
	// 原版:`"That will be <價> gold for the <two..six> of ye, <sir|milady>.`
	// ★ `sub_20ED0` 只認 2..6 的英文數字 ⇒ **一個人時那個位置是空的**
	// (原版就印成 "for the  of ye"),照抄。
	s.Log("「這樣是 " + strconv.Itoa(sess.Price) + " 金," + partyCountWord(alive) +
		"位" + s.addressWord() + "。」")
	s.Log("「汝要嗎?(Y/N)」")
	sess.Mode = ShopModeConfirm
}

// partyCountWord 是原版 `sub_20ED0` 的英文數字("two".."six")。
//
// ⚠ 它只認 **2..6**;其他值(含 1)走 `default` 什麼都不印 ——
// 一人隊伍在原版會看到 "gold for the  of ye"。這裡照樣留空。
func partyCountWord(n int) string {
	if n < 2 || n > u5data.CombatPartySlots {
		return ""
	}
	return [...]string{"二", "三", "四", "五", "六"}[n-2]
}

// addressWord 是原版 `sub_20F24` 的 "sir" / "milady"。
//
// ★★ 那支函式有**死碼**:開頭 `al = 0FFh; byte_3E08B = al; and al, al; jz`——
// `al` 永遠不是 0,所以走的是 `edi = 0` 那條 ⇒ **永遠看名冊第 0 筆的性別**,
// 不是說話對象。原版的 bug,照抄(`CLAUDE.md §3.0`)。
//
// ⚠ 附帶副作用:它把 `byte_3E08B` 寫成 0xFF。⬜ 那個位元組的語意還沒查,
// 所以引擎沒有重現這個副作用 —— 有誰讀它的話會是一條 finding。
func (s *State) addressWord() string {
	if len(s.Roster) > 0 && s.Roster[0].Gender == u5data.GenderMale {
		return "先生"
	}
	return "夫人"
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
		if it.ServeOnTable {
			s.serveOnTheTable()
		}
	case GoodsDrink:
		s.Shop.drinks++
		s.Log("「請慢用。」")
		// ⚠ **更正**:此前這裡寫「原版只把『這趟喝了幾杯』加一,沒有其他效果」——
		// 那是錯的。同一支 `sub_21108` 還有一行 `mov byte_3E169, 19h`,
		// 也就是**喝一杯醉 25 次**:接下來每次按鍵有一半機率變成隨機走一步
		// (`sub_1158` 的 `Hic!`,見 `upkeep.go`)。
		//
		// 沒讀到的原因與許願井同一個 —— `sub_21108` 也在截斷清單上
		// (153 行 → 6 行 C、17 個字串全掉,`docs/re/66`)。
		s.GetDrunk()
	}
}

// ---------------------------------------------------------------- 打聽消息

// 原版 `sub_21500`。26 個主題、四句隨機模板,表在 `DATA.OVL` 0x4C84 起
//(見 `u5data/tavernlore.go` 與 `docs/re/61`)。
//
// ⚠ 這是**打字**的一步,不是按選單字母:玩家打進去的字要跟 26 個四字母關鍵字
// 做子字串比對。所以 canonical 值一律維持英文(CLAUDE.md §5.2)——
// 打「honesty」「hone」都問得到誠實,而中文提示只出現在旁邊的說明裡。

// askTavernLore 問「汝想聽哪一樁事」。
func (s *State) askTavernLore() {
	sess := s.Shop
	if s.Lore == nil {
		// 表沒載進來就誠實說,不要假裝談成了(CLAUDE.md §3.0)。
		s.Log("(情報表未載入 —— 缺 DATA.OVL 的酒館表)")
		s.backToMenu()
		return
	}
	sess.LoreInput = ""
	sess.LoreTopic = -1
	sess.Mode = ShopModeTavernLoreAsk
	s.Log("「" + s.AvatarName() + "啊,汝想聽我說哪一樁事?」")
	s.Log("(打英文關鍵字,如 honesty / crown / deceit)")
}

// tavernLoreType 收打聽消息的輸入(最多 15 個字,Enter 送出)。
func (s *State) tavernLoreType(r rune) {
	sess := s.Shop
	switch {
	case r == '\r' || r == '\n':
		s.submitTavernLore()
	case r == 8 || r == 127:
		if sess.LoreInput != "" {
			sess.LoreInput = sess.LoreInput[:len(sess.LoreInput)-1]
		}
	case r >= ' ' && r < 0x7F:
		if len(sess.LoreInput) < u5data.TavernLoreAnswerMax {
			sess.LoreInput += string(r)
		}
	}
}

// submitTavernLore 比對關鍵字並報價。
func (s *State) submitTavernLore() {
	sess := s.Shop
	answer := sess.LoreInput
	sess.LoreInput = ""
	// 什麼都沒打就結束 —— 原版 `cmp byte_55F38, 0; jz` 直接回 0。
	if answer == "" {
		s.backToMenu()
		return
	}
	topic := s.Lore.Match(answer)
	if topic < 0 {
		s.Log("「這一樁,我幫不上汝。」")
		// 原版跳回**問題本身**,不是回菜單 —— 可以一直問到猜中。
		s.askTavernLore()
		return
	}
	sess.LoreTopic = topic
	sess.Price = s.Lore.Entries[topic].Price
	sess.Mode = ShopModeTavernLoreConfirm
	if line := s.pitch(u5data.TavernLorePriceLine, sess.Price, 0); line != "" {
		s.Log(line)
	}
	s.Log("「這樣可以吧?(Y/N)」")
}

// answerTavernLore 是報價之後的 Y / N(原版只收這兩鍵,其餘繼續等)。
func (s *State) answerTavernLore(r rune) {
	switch r {
	case 'n':
		s.Log("否。")
		s.backToMenu()
	case 'y':
		s.payTavernLore()
	}
	// 其他鍵:原版 `call sub_29EEC` 迴圈繼續等,所以這裡什麼都不做。
}

// payTavernLore 收錢並說出線索。
func (s *State) payTavernLore() {
	sess := s.Shop
	e := s.Lore.Entries[sess.LoreTopic]
	// 金幣不夠:「Sorry, <隊長>」+ 0x146A。⚠ 原版比的是 `jle`,
	// 所以**剛好等於價格是付得出來的**。
	if s.Inventory.Gold < e.Price {
		s.Log("「抱歉," + s.AvatarName())
		if line := s.pitch(u5data.TavernLoreCannotAfford, 0, 0); line != "" {
			s.Log(line)
		}
		s.backToMenu()
		return
	}
	s.Inventory.Gold -= e.Price
	// 四句模板隨機挑一句(原版 `random(0, 3)`),`&` 填人名、`*` 填地名。
	say := u5data.TavernLoreSay[s.Roll(0, len(u5data.TavernLoreSay)-1)]
	if line := s.loreSay(say, s.loreWho(e.Who), s.lorePlace(e.Where)); line != "" {
		s.Log(line)
	}
	s.Log("—— " + s.Shop.Shop.Owner + "如是說。")
	s.backToMenu()
}

// loreWho / lorePlace 有譯名就用譯名,沒有就照原樣顯示英文。
func (s *State) loreWho(en string) string {
	if zh := i18n.Name(en); zh != "" {
		return zh
	}
	return en
}

func (s *State) lorePlace(en string) string { return s.loreWho(en) }

// loreSay 取一句線索模板:& 填人名、* 填地名。
func (s *State) loreSay(off int, who, place string) string {
	if off == 0 || s.Shops == nil || s.Shop == nil {
		return ""
	}
	return oneLine(s.Shops.SayWith(off, s.Shop.Shop, u5data.Placeholders{
		Hour: s.Clock.Hour, Item: who, Place: place,
		TimeWord: i18n.TimeOfDay(s.Clock.Hour),
	}))
}

// TavernDrinkNagCount 是店主開口勸的杯數(原版 `cmp dword_56E1C, 3`)。
//
// ⚠ 判準是 `jnz` ⇒ **剛好等於 3**。喝到第四杯之後這一問就再也不出現 ——
// 原版的寫法,不要「順手」改成 `>= 3`(`CLAUDE.md §3.0`)。
const TavernDrinkNagCount = 3

// 空桌與上了菜的桌子(原版 `sub_20F60` 尾段直接改地圖那一格)。
const (
	TableEmpty      = 0x95 // 空桌
	TableServedNorth = 0x9B // 玩家北邊那一格上了菜
	TableServedSouth = 0x9A // 玩家南邊那一格上了菜
)

// serveOnTheTable 把菜端上桌(原版 `sub_20F60` 的最後三個分支)。
//
//	北邊那一格是空桌 → 改成 0x9B
//	否則南邊那一格是空桌 → 改成 0x9A(原版這一支還多叫一次重畫)
//	兩邊都不是桌子     → 什麼都不放
//
// ★ 這是**寫進地圖緩衝**的一格,與月門同一個手法(`docs/re/86`)——
// 不是動畫、不是物件。⬜ 原版沒有把它收回去的程式碼,菜就一直留在桌上。
func (s *State) serveOnTheTable() {
	if s.TileAt(s.X, s.Y-1) == TableEmpty {
		s.SetTileAt(s.X, s.Y-1, TableServedNorth)
		return
	}
	if s.TileAt(s.X, s.Y+1) == TableEmpty {
		s.SetTileAt(s.X, s.Y+1, TableServedSouth)
	}
}
