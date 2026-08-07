package game

import (
	"strconv"
	"strings"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 商店交易
//
// 八種店在原版是八支獨立的處理函式(`sub_1B294` 的跳表 `jpt_1B363`),
// 但買東西那一段是同一個模子:
//
//  1. 從價目表查底價 —— 索引一律是「這型店的第幾家」(`word_3EF34`)。
//  2. 套議價公式 `u5data.Haggle`(**用交談者的智力**,不是隊長)。
//  3. 問 Y/N;答 Yes 才比對金幣。
//  4. 錢不夠 → 一句冷嘲,交易失敗;錢夠 → 扣錢、加貨(上限 99)。
//
// 八種店的流程都逐條對過反編譯:旅店與造船廠在 inn.go、酒館在 tavern.go。
// 還沒做的只剩兩處,兩處都不是「不知道規則」而是缺前置系統:
// 買馬買船要地圖物件層才生得出坐騎,酒館的打聽消息要八德與地牢的知識表。

// ShopMode 是商店互動目前停在哪一步。
type ShopMode int

const (
	// ShopModeNone 表示不在店裡。
	ShopModeNone ShopMode = iota
	// ShopModeBuySell 是武具店的第一層:買(B)還是賣(S)。
	ShopModeBuySell
	// ShopModeMenu 是在看貨架,等玩家按 a/b/c… 選一項。
	ShopModeMenu
	// ShopModeSell 是在看背包,等玩家挑一件要賣的。
	ShopModeSell
	// ShopModeConfirm 是已經報過價,等玩家答 Y/N。
	ShopModeConfirm
	// ShopModeHealMenu 是治療所問「Cure / Heal / Resurrect」。
	ShopModeHealMenu
	// ShopModeHealTarget 是治療所問「治誰」。
	ShopModeHealTarget
	// ShopModeInnMenu 是旅店問「領人(P)、寄放(L)、還是住宿(R)」。
	ShopModeInnMenu
	// ShopModeInnLeave 是旅店問「誰要留下」。
	ShopModeInnLeave
	// ShopModeInnPick 是旅店問「誰要退房」。
	ShopModeInnPick
	// ShopModeTavernMenu 是酒館的四欄菜單(熱鍵字母每家不同)。
	ShopModeTavernMenu
	// ShopModeTavernWine 是酒館的酒單。
	ShopModeTavernWine
	// ShopModeTavernQty 是酒館問「乾糧要幾份」。
	ShopModeTavernQty
)

// ShopSession 是一次進店的完整狀態。
type ShopSession struct {
	Shop *u5data.Shop
	Mode ShopMode
	// Menu 是目前列出的品項,順序就是玩家按鍵 a/b/c… 的順序。
	Menu []ShopItem
	// Choice 是玩家選中的那一項(Mode 為 ShopModeConfirm 時有效)。
	Choice ShopItem
	// Price 是已經套過議價公式、玩家實際要付的錢。
	Price int
	// Service 是治療所選的服務。
	Service HealService
	// Target 是治療所要治的名冊索引。
	Target int
	// Action 是這次要結的帳做什麼用(旅店住宿 / 寄放 / 退房)。
	Action ShopAction
	// guests 是旅店 REGISTER 上列出的名冊索引,順序就是按鍵順序。
	guests []int
	// drinks 是這趟在酒館喝了幾杯(原版 dword_56E1C)。
	drinks int
	// clerk 是應對玩家的角色索引(名冊位置);議價用他的智力。
	clerk int
}

// ShopAction 是「付了錢之後要發生什麼」。
//
// 買東西以外的交易(住宿、寄放同伴、退房、買船)沒有「貨」可以進背包,
// 各自要做的事不同,用這個分。
type ShopAction int

const (
	// ActionBuy 是一般的買貨。
	ActionBuy ShopAction = iota
	// ActionSell 是賣貨給店家。
	ActionSell
	// ActionHeal 是治療所的服務。
	ActionHeal
	// ActionInnRest 是在旅店過夜。
	ActionInnRest
	// ActionInnLeave 是把同伴寄放在旅店。
	ActionInnLeave
	// ActionInnPick 是把寄放的同伴領回來(結清住宿費)。
	ActionInnPick
	// ActionShip 是在造船廠買船。
	ActionShip
	// ActionTavern 是在酒館點餐 / 點酒 / 買乾糧。
	ActionTavern
)

// ShopGoods 分辨這一項是什麼東西 —— 決定買下去要加到哪個欄位。
type ShopGoods int

const (
	// GoodsItem 是裝備(加到 Inventory.Items)。
	GoodsItem ShopGoods = iota
	// GoodsReagent 是藥草(加到 Inventory.Reagents)。
	GoodsReagent
	// GoodsGuild 是公會的鑰匙 / 寶石 / 火把。
	GoodsGuild
	// GoodsHorse 是馬。
	GoodsHorse
	// GoodsShip 是船(ID 0 帆船、1 小艇)。
	GoodsShip
	// GoodsFood 是存糧(酒館的一餐或乾糧)。
	GoodsFood
	// GoodsDrink 是酒館的酒。
	GoodsDrink
)

// ShopItem 是貨架上的一項。
type ShopItem struct {
	Goods ShopGoods
	// ID 依 Goods 而定:裝備編號 / 藥草編號 / 公會品項(0..2)。
	ID int
	// Name 是顯示用的名字。
	Name string
	// Base 是底價(還沒議價)。
	Base int
	// Qty 是成交一次拿到幾個。
	Qty int
	// Pitch 是店員的說詞在 SHOPPE.DAT 的位移;0 代表沒有。
	Pitch int
}

// HealService 是治療所的三種服務。
type HealService int

const (
	// HealCure 解毒。
	HealCure HealService = iota
	// HealHeal 補血。
	HealHeal
	// HealResurrect 復活。
	HealResurrect
)

// Price 回傳這項服務的固定價(原版寫在程式碼裡的立即數)。
func (h HealService) Price() int {
	switch h {
	case HealCure:
		return u5data.CurePrice
	case HealHeal:
		return u5data.HealPrice
	case HealResurrect:
		return u5data.ResurrectPrice
	}
	return 0
}

// Name 回傳服務的中文名。
func (h HealService) Name() string {
	switch h {
	case HealCure:
		return "解毒"
	case HealHeal:
		return "療傷"
	case HealResurrect:
		return "復生"
	}
	return "?"
}

// openShop 依店種擺出貨架。回傳 false 代表這種店的流程還沒實作。
func (s *State) openShop(shop *u5data.Shop) bool {
	p := s.Shops.Prices
	if p == nil {
		return false
	}
	sess := &ShopSession{Shop: shop, Mode: ShopModeMenu, clerk: 0}
	switch shop.Type {
	case u5data.ShopArmoury:
		// 武具店先問買還是賣(原版 sub_1258C 只收 B / S / 空白)。
		sess.Mode = ShopModeBuySell
	case u5data.ShopReagents:
		for _, r := range p.ReagentStockList(shop.TypeIndex) {
			sess.Menu = append(sess.Menu, ShopItem{
				Goods: GoodsReagent, ID: r, Name: u5data.ReagentNames[r],
				Base:  p.ReagentPrice[shop.TypeIndex][r],
				Qty:   p.ReagentQty[shop.TypeIndex][r],
				Pitch: p.ReagentPitch[r],
			})
		}
	case u5data.ShopGuild:
		for g := 0; g < u5data.GuildGoods; g++ {
			sess.Menu = append(sess.Menu, ShopItem{
				Goods: GoodsGuild, ID: g, Name: u5data.GuildGoodsNames[g],
				Base: p.Guild[shop.TypeIndex][g], Qty: u5data.GuildGoodsQty[g],
				Pitch: p.GuildPitch[g],
			})
		}
	case u5data.ShopStable:
		sess.Menu = append(sess.Menu, ShopItem{
			Goods: GoodsHorse, ID: 0, Name: "Horse",
			Base: p.Stable[shop.TypeIndex], Qty: 1,
		})
	case u5data.ShopHealer:
		sess.Mode = ShopModeHealMenu
	case u5data.ShopInn:
		return s.openInn(shop)
	case u5data.ShopShipwright:
		return s.openShipwright(shop)
	case u5data.ShopTavern:
		return s.openTavern(shop)
	default:
		return false
	}
	if sess.Mode == ShopModeMenu && len(sess.Menu) == 0 {
		return false
	}
	s.Shop = sess
	s.Prompt = PromptShop
	s.showShopMenu()
	return true
}

// fillBuyMenu 擺出武具店的貨架。
func (s *State) fillBuyMenu() {
	sess := s.Shop
	p := s.Shops.Prices
	sess.Menu = sess.Menu[:0]
	for _, id := range p.ArmouryStockList(sess.Shop.TypeIndex) {
		sess.Menu = append(sess.Menu, ShopItem{
			Goods: GoodsItem, ID: int(id), Name: s.itemName(id),
			Base: p.Item[id], Qty: 1, Pitch: p.ItemPitch[id],
		})
	}
	sess.Mode = ShopModeMenu
}

// fillSellMenu 列出背包裡「有貨而且店家肯收」的裝備。
//
// 原版是一個可捲動的視窗(sub_12198),逐格掃 byte_3DFD0 的 48 個欄位;
// 不收的兩類在 sub_12060 才擋 —— 箭矢與弩矢(「不收二手彈藥」)、
// 底價 0 的東西(聖物與任務物品,「這個我不收」)。這裡提前濾掉,
// 讓玩家不必按了才被拒絕;被擋的理由本身仍保留在 sellQuote。
func (s *State) fillSellMenu() {
	sess := s.Shop
	p := s.Shops.Prices
	sess.Menu = sess.Menu[:0]
	for id := 0; id < u5data.ItemCount; id++ {
		if s.Inventory.Items[id] == 0 || p.Item[id] == 0 {
			continue
		}
		if id == u5data.ItemArrows || id == u5data.ItemQuarrels {
			continue
		}
		sess.Menu = append(sess.Menu, ShopItem{
			Goods: GoodsItem, ID: id, Name: s.itemName(byte(id)),
			Base: p.Item[id], Qty: 1,
		})
	}
	sess.Mode = ShopModeSell
}

func (s *State) itemName(id byte) string {
	if n := s.Items.Name(id); n != "" {
		return n
	}
	return "#" + strconv.Itoa(int(id))
}

// showShopMenu 印出貨架或服務選單。
func (s *State) showShopMenu() {
	sess := s.Shop
	if sess == nil {
		return
	}
	switch sess.Mode {
	case ShopModeHealMenu:
		s.Log("「吾等能" + HealCure.Name() + "、" + HealHeal.Name() + "、" + HealResurrect.Name() + "。汝所需為何?」")
		s.Log("  c) " + HealCure.Name() + "   h) " + HealHeal.Name() + "   r) " + HealResurrect.Name())
	case ShopModeHealTarget:
		var b strings.Builder
		for i, c := range s.Party() {
			if i > 0 {
				b.WriteString("  ")
			}
			b.WriteString(string(rune('a'+i)) + ") " + c.Name)
		}
		s.Log("「為誰施為?」")
		s.Log("  " + b.String())
	case ShopModeBuySell:
		s.Log("「汝欲買(B)抑或賣(S)?」")
	case ShopModeMenu:
		for i, it := range sess.Menu {
			s.Log("  " + string(rune('a'+i)) + " ... " + it.Name)
		}
		s.Log("「汝意欲何物?」")
	case ShopModeSell:
		if len(sess.Menu) == 0 {
			s.Log("「汝身上沒有我收得起的東西。」")
			sess.Mode = ShopModeBuySell
			s.showShopMenu()
			return
		}
		for i, it := range sess.Menu {
			s.Log("  " + string(rune('a'+i)) + " ... " + it.Name +
				" ×" + strconv.Itoa(s.Inventory.Items[it.ID]))
		}
		s.Log("「汝欲售何物?」")
	}
}

// ShopChoose 處理玩家在店裡按的一個字母鍵。
//
// 原版每一步都是「等一個按鍵」而不是打字,所以這裡也走按鍵而非 Input。
func (s *State) ShopChoose(r rune) {
	sess := s.Shop
	if sess == nil || s.Prompt != PromptShop {
		return
	}
	if r >= 'A' && r <= 'Z' {
		r = r - 'A' + 'a'
	}
	switch sess.Mode {
	case ShopModeInnMenu, ShopModeInnLeave, ShopModeInnPick:
		s.innChoose(r)
	case ShopModeTavernMenu, ShopModeTavernWine, ShopModeTavernQty:
		s.tavernChoose(r)
	case ShopModeBuySell:
		switch r {
		case 'b':
			s.Log("買。")
			s.fillBuyMenu()
			s.showShopMenu()
		case 's':
			s.Log("賣。")
			s.fillSellMenu()
			s.showShopMenu()
		default:
			s.LeaveShop()
		}
	case ShopModeMenu:
		i := int(r - 'a')
		if i < 0 || i >= len(sess.Menu) {
			s.LeaveShop()
			return
		}
		s.quote(sess.Menu[i])
	case ShopModeSell:
		i := int(r - 'a')
		if i < 0 || i >= len(sess.Menu) {
			s.LeaveShop()
			return
		}
		s.sellQuote(sess.Menu[i])
	case ShopModeConfirm:
		switch r {
		case 'y':
			s.settle()
		default:
			s.Log("否。")
			s.backToMenu()
		}
	case ShopModeHealMenu:
		switch r {
		case 'c':
			sess.Service = HealCure
		case 'h':
			sess.Service = HealHeal
		case 'r':
			sess.Service = HealResurrect
		default:
			s.LeaveShop()
			return
		}
		sess.Mode = ShopModeHealTarget
		s.showShopMenu()
	case ShopModeHealTarget:
		i := int(r - 'a')
		party := s.Party()
		if i < 0 || i >= len(party) {
			s.LeaveShop()
			return
		}
		s.quoteHeal(i)
	}
}

// backToMenu 回到貨架。原版每筆交易後都會再問一次「還要別的嗎」。
func (s *State) backToMenu() {
	sess := s.Shop
	if sess == nil {
		return
	}
	s.Log("「還需要別的嗎?」")
	switch {
	case sess.Shop.Type == u5data.ShopInn:
		sess.Mode = ShopModeInnMenu
		s.showInnMenu()
		return
	case sess.Shop.Type == u5data.ShopTavern:
		sess.Mode = ShopModeTavernMenu
		s.showTavernMenu()
		return
	case sess.Shop.Type == u5data.ShopHealer:
		sess.Mode = ShopModeHealMenu
	case sess.Action == ActionSell:
		s.fillSellMenu()
	case sess.Shop.Type == u5data.ShopArmoury:
		s.fillBuyMenu()
	default:
		sess.Mode = ShopModeMenu
	}
	s.showShopMenu()
}

// LeaveShop 離開商店。
func (s *State) LeaveShop() {
	if s.Shop == nil {
		return
	}
	s.Shop = nil
	s.Prompt = PromptNone
	s.Log("「後會有期。」")
}

// clerkIntel 是議價時用的智力。
//
// 原版 `sub_11AF0` 拿的是 `byte_3DDC2[arg_0*32]` —— arg_0 一路從 `sub_1B294`
// 傳下來,是**跟商人交談的那名角色**(隊伍第 0 位)。智力每點折 3%。
func (s *State) clerkIntel() int {
	party := s.Party()
	if len(party) == 0 {
		return 0
	}
	i := s.Shop.clerk
	if i < 0 || i >= len(party) {
		i = 0
	}
	return int(party[i].Intel)
}

// sellQuote 店家對一件裝備開價(原版 sub_12060)。
func (s *State) sellQuote(it ShopItem) {
	sess := s.Shop
	sess.Action = ActionSell
	sess.Choice = it
	if it.ID == u5data.ItemArrows || it.ID == u5data.ItemQuarrels {
		s.Log("「二手彈藥恕不收購。」")
		s.backToMenu()
		return
	}
	if it.Base == 0 {
		s.Log("「此物我不收。」")
		s.backToMenu()
		return
	}
	sess.Price = u5data.SellValue(it.Base, s.clerkIntel())
	// 八句開價說詞隨機挑一句;`&` 換成裝備名(有另一種說法就用那個)。
	pitches := s.Shops.Prices.SellPitch
	off := pitches[s.greetVariant()%len(pitches)]
	name := it.Name
	if alt, ok := u5data.ItemAltNames[it.ID]; ok {
		name = alt
	}
	if line := s.pitchWithItem(off, sess.Price, name); line != "" {
		s.Log("「" + line + "」")
	} else {
		s.Log("「這件" + name + ",我出 " + strconv.Itoa(sess.Price) + " 金。」")
	}
	sess.Mode = ShopModeConfirm
	s.Log("「成交嗎?(Y/N)」")
}

// sell 收下玩家的東西。金幣有 9999 的上限(原版 sub_2BBDC 的第三個參數)。
func (s *State) sell(it ShopItem, price int) {
	s.Inventory.Gold += price
	if s.Inventory.Gold > u5data.GoldLimit {
		s.Inventory.Gold = u5data.GoldLimit
	}
	if s.Inventory.Items[it.ID] > 0 {
		s.Inventory.Items[it.ID]--
	}
	s.Log("成交!")
}

// quote 報價。
func (s *State) quote(it ShopItem) {
	sess := s.Shop
	sess.Action = ActionBuy
	sess.Choice = it
	sess.Price = u5data.Haggle(it.Base, s.clerkIntel())
	if p := s.pitch(it.Pitch, sess.Price, it.Qty); p != "" {
		s.Log("「" + p + "」")
	} else {
		s.Log("「" + it.Name + ",要價 " + strconv.Itoa(sess.Price) + " 金。」")
	}
	if s.carried(it) >= u5data.CarryLimit {
		s.Log("「汝已拿不下更多了!」")
		s.backToMenu()
		return
	}
	sess.Mode = ShopModeConfirm
	s.Log("「汝要買嗎?(Y/N)」")
}

// quoteHeal 報治療的價。
func (s *State) quoteHeal(target int) {
	sess := s.Shop
	party := s.Party()
	c := party[target]
	sess.Target = target
	sess.Price = sess.Service.Price()
	if !healNeeded(sess.Service, c) {
		s.Log("「汝無需此術!」")
		s.backToMenu()
		return
	}
	s.Log("「吾可為" + c.Name + "行" + sess.Service.Name() + "之事,需 " +
		strconv.Itoa(sess.Price) + " 金。汝願付否?(Y/N)」")
	sess.Mode = ShopModeConfirm
}

// healNeeded 對應原版三個分支各自的前置判斷。
func healNeeded(svc HealService, c *u5data.Character) bool {
	switch svc {
	case HealCure:
		return c.Status == u5data.StatusPoisoned
	case HealHeal:
		return c.Status != u5data.StatusDead && c.HP < c.MaxHP
	case HealResurrect:
		return c.Status == u5data.StatusDead
	}
	return false
}

// settle 收錢交貨。
func (s *State) settle() {
	sess := s.Shop
	s.Log("是。")
	if sess.Action == ActionSell {
		s.sell(sess.Choice, sess.Price)
		s.backToMenu()
		return
	}
	if s.Inventory.Gold < sess.Price {
		// 旅店付不起是被轟出去,不是回到選單(原版三段各有一句台詞)。
		// 治療所在特定地點(原版 `cmp byte_3E0A3, 7`)對 100 金以下的服務
		// 不趕人 —— 付不起也照做。復活 200 金不在此列。
		charity := sess.Shop.Type == u5data.ShopHealer &&
			s.Location == u5data.CharityLocation && sess.Price <= u5data.CharityMax
		if !charity {
			s.Log("「付不出錢?滾出去!」")
			s.backToMenu()
			return
		}
	} else {
		s.Inventory.Gold -= sess.Price
	}
	switch {
	case sess.Shop.Type == u5data.ShopHealer:
		s.applyHeal()
	case sess.Shop.Type == u5data.ShopInn:
		s.settleInn()
	case sess.Shop.Type == u5data.ShopTavern:
		s.settleTavern()
	default:
		s.deliver(sess.Choice)
		s.Log("成交!")
	}
	s.backToMenu()
}

// applyHeal 套用治療效果。
func (s *State) applyHeal() {
	sess := s.Shop
	party := s.Party()
	if sess.Target < 0 || sess.Target >= len(party) {
		return
	}
	c := party[sess.Target]
	switch sess.Service {
	case HealCure:
		c.Status = u5data.StatusGood
		c.Raw[u5data.CharStatus] = u5data.StatusGood
	case HealHeal:
		c.HP = c.MaxHP
	case HealResurrect:
		c.Status = u5data.StatusGood
		c.Raw[u5data.CharStatus] = u5data.StatusGood
		c.HP = c.MaxHP
	}
	s.Log(c.Name + "已" + sess.Service.Name() + "。")
}

// carried 回報目前持有幾個 —— 上限判斷用。
func (s *State) carried(it ShopItem) int {
	switch it.Goods {
	case GoodsItem:
		return s.Inventory.Items[it.ID]
	case GoodsReagent:
		return s.Inventory.Reagents[it.ID]
	case GoodsGuild:
		switch it.ID {
		case 0:
			return s.Inventory.Keys
		case 1:
			return s.Inventory.Gems
		case 2:
			return s.Inventory.Torches
		}
	}
	return 0
}

// deliver 把買到的東西放進背包。
//
// ⚠ 箭矢(27)與弩矢(29)是特例:原版直接把數量設成 99,而不是加 1
// (`sub_11AF0` 的 `cmp edi, 1Bh / cmp edi, 1Dh`)。買一次就補滿。
func (s *State) deliver(it ShopItem) {
	add := func(cur, n int) int {
		cur += n
		if cur > u5data.CarryLimit {
			cur = u5data.CarryLimit
		}
		return cur
	}
	switch it.Goods {
	case GoodsItem:
		if it.ID == u5data.ItemArrows || it.ID == u5data.ItemQuarrels {
			s.Inventory.Items[it.ID] = u5data.CarryLimit
		} else {
			s.Inventory.Items[it.ID] = add(s.Inventory.Items[it.ID], it.Qty)
		}
	case GoodsReagent:
		s.Inventory.Reagents[it.ID] = add(s.Inventory.Reagents[it.ID], it.Qty)
	case GoodsGuild:
		switch it.ID {
		case 0:
			s.Inventory.Keys = add(s.Inventory.Keys, it.Qty)
		case 1:
			s.Inventory.Gems = add(s.Inventory.Gems, it.Qty)
		case 2:
			s.Inventory.Torches = add(s.Inventory.Torches, it.Qty)
		}
	case GoodsHorse:
		s.spawnMount(u5data.TileHorse, "馬")
	case GoodsShip:
		// ⚠ 買船**不生成物件槽** —— 原版 `sub_218DC` 只把停泊座標寫進
		// byte_3E165 / byte_3E166,船在碼頭等你。
		p := s.Shops.Prices
		i := s.Shop.Shop.TypeIndex
		s.DockX, s.DockY = p.DockX[i], p.DockY[i]
		s.HasShip = true
		s.Log(it.Name + "已備妥,停在碼頭 (" +
			strconv.Itoa(s.DockX) + "," + strconv.Itoa(s.DockY) + ")。")
	}
}

// spawnMount 把買到的坐騎或船放到店旁邊的空地上。
//
// 原版 `sub_118CC` 開頭就先找位置:依 **南、北、東、西** 的順序看四個鄰格
// (`dword_555E8` = {0,0,1,-1}、`dword_555F8` = {1,-1,0,0}),
// 挑第一個「沒有東西擋著、而且地形是 5 / 68 / 69」的格子。
// 四格都不行就「馬廄關門了」—— 買賣根本不會開始。
//
// 這裡在成交後才放,結果一樣:找不到位置就誠實說明,而不是讓坐騎憑空消失。
func (s *State) spawnMount(tile byte, what string) {
	objs := s.currentObjects()
	if objs == nil {
		s.Log("(" + what + "無處可放)")
		return
	}
	for _, d := range []Direction{South, North, East, West} {
		dx, dy := d.Delta()
		x, y := s.X+dx, s.Y+dy
		if !u5data.TileAllowsMount(int(s.TileAt(x, y))) {
			continue
		}
		if _, _, occupied := s.ObjectAt(x, y); occupied {
			continue
		}
		if _, taken := s.NPCAt(x, y); taken {
			continue
		}
		if _, ok := objs.Spawn(tile, x, y, s.Floor); ok {
			s.Log(what + "已備妥,就在" + d.Name() + "邊。")
			return
		}
	}
	s.Log("(此處放不下" + what + ")")
}

// pitch 取店員的推銷詞:% 是價格,^ 是一次賣幾份(藥草才有)。
func (s *State) pitch(off, price, count int) string {
	if off == 0 || s.Shops == nil || s.Shop == nil {
		return ""
	}
	ph := u5data.Placeholders{Hour: s.Clock.Hour, Number: price, Count: count}
	return oneLine(s.Shops.SayWith(off, s.Shop.Shop, ph))
}

// pitchWithItem 取店員的收購說詞:% 是價格,& 是裝備名。
func (s *State) pitchWithItem(off, price int, item string) string {
	if off == 0 || s.Shops == nil || s.Shop == nil {
		return ""
	}
	ph := u5data.Placeholders{Hour: s.Clock.Hour, Number: price, Item: item}
	return oneLine(s.Shops.SayWith(off, s.Shop.Shop, ph))
}
