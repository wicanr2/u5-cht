package game

import (
	"strconv"
	"strings"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 旅店與造船廠
//
// 旅店(原版 `sub_2274C`)先問一句 Y/N,再分三條路:
//
//	P  Pick up  領回寄放的同伴(`sub_22280`)—— 結清住宿費
//	L  Leave    把同伴寄放在這裡(`sub_22018`)
//	R  Rest     過夜(`sub_21D48`)
//
// 三條路的價都建立在同一個底價 `byte_57090[旅店]`(每人每天 2 或 3 金)上,
// 再各自乘上人數或月數,最後套與買東西同一條議價公式。

// 角色紀錄裡與旅店有關的兩個欄位(位址由 `sub_22018` 的寫入處反推:
// `byte_3DDD3` − 0x3DDB4 = 0x1F、`byte_3DDCB` − 0x3DDB4 = 0x17)。
//
// ⚠ `CharInnFlag` 存的**不是布林**,是「寄放在哪一個地點」——
// `sub_22018` 寫進去的正是 `byte_3E0A3`(當前地點編號),
// 而 `sub_22280` 列 REGISTER 時就拿它跟當前地點比。0 代表沒寄放。

// openInn 進旅店。
func (s *State) openInn(shop *u5data.Shop) bool {
	s.Shop = &ShopSession{Shop: shop, Mode: ShopModeInnMenu}
	s.Prompt = PromptShop
	s.showInnMenu()
	return true
}

func (s *State) showInnMenu() {
	sess := s.Shop
	switch sess.Mode {
	case ShopModeInnMenu:
		s.Log("「汝是來領人(P)、寄放同伴(L),還是要住一晚(R)?」")
	case ShopModeInnLeave:
		s.Log("「誰要留下?」")
		s.logRoster(s.partyNames())
	case ShopModeInnPick:
		s.Log("  住 宿 名 冊:")
		s.logRoster(s.guestNames())
		s.Log("「誰要退房?」")
	}
}

func (s *State) logRoster(names []string) {
	var b strings.Builder
	for i, n := range names {
		if i > 0 {
			b.WriteString("  ")
		}
		b.WriteString(string(rune('a'+i)) + ") " + n)
	}
	s.Log("  " + b.String())
}

func (s *State) partyNames() []string {
	var out []string
	for _, c := range s.Party() {
		out = append(out, c.Name)
	}
	return out
}

// innGuests 找出寄放在這家旅店的名冊索引。
//
// 原版 `sub_22280` 掃全部 16 筆,比對 `byte_3DDD3[i*32] == byte_3E0A3`。
func (s *State) innGuests() []int {
	var out []int
	for i := range s.Roster {
		c := &s.Roster[i]
		if !c.Present() || i < s.PartySize {
			continue
		}
		if int(c.Raw[u5data.CharInnFlag]) == s.Location {
			out = append(out, i)
		}
	}
	return out
}

func (s *State) guestNames() []string {
	var out []string
	for _, i := range s.innGuests() {
		out = append(out, s.Roster[i].Name)
	}
	return out
}

// innChoose 處理旅店裡按下的鍵。
func (s *State) innChoose(r rune) {
	sess := s.Shop
	p := s.Shops.Prices
	idx := sess.Shop.TypeIndex
	switch sess.Mode {
	case ShopModeInnMenu:
		switch r {
		case 'r':
			s.Log("住宿。")
			s.quoteRest()
		case 'l':
			s.Log("寄放。")
			if s.PartySize <= 1 {
				s.Log("「汝的同伴不會離開汝!」")
				s.showInnMenu()
				return
			}
			sess.Mode = ShopModeInnLeave
			s.showInnMenu()
		case 'p':
			s.Log("領人。")
			sess.guests = s.innGuests()
			if len(sess.guests) == 0 {
				s.Log("「此處沒有汝的同伴。」")
				s.showInnMenu()
				return
			}
			if s.PartySize >= u5data.MaxPartySize {
				s.Log("「汝得先留下一人!」")
				s.showInnMenu()
				return
			}
			sess.Mode = ShopModeInnPick
			s.showInnMenu()
		default:
			s.LeaveShop()
		}
	case ShopModeInnLeave:
		i := int(r - 'a')
		if i < 0 || i >= s.PartySize {
			s.LeaveShop()
			return
		}
		// 客滿判斷(原版 sub_21CE4):寄放的人不能超過房間數。
		if len(s.innGuests()) >= p.InnRooms[idx] {
			s.Log("「抱歉,已經沒有空房了。」")
			sess.Mode = ShopModeInnMenu
			s.showInnMenu()
			return
		}
		sess.Target = i
		sess.Action = ActionInnLeave
		sess.Price = u5data.Haggle(p.Inn[idx], s.clerkIntel())
		s.Log("「最好的房間一個月 " + strconv.Itoa(sess.Price) + " 金,退房時結算。汝要嗎?(Y/N)」")
		sess.Mode = ShopModeConfirm
	case ShopModeInnPick:
		i := int(r - 'a')
		if i < 0 || i >= len(sess.guests) {
			s.LeaveShop()
			return
		}
		who := sess.guests[i]
		months := int(s.Roster[who].Raw[u5data.CharInnDays])
		if months == 0 {
			months = 1 // 原版:不足一期也照一期算
		}
		sess.Target = who
		sess.Action = ActionInnPick
		sess.Price = u5data.Haggle(p.Inn[idx], s.clerkIntel()) * months
		s.Log("「這樣是 " + strconv.Itoa(sess.Price) + " 金,謝謝。」")
		sess.Mode = ShopModeConfirm
	}
}

// quoteRest 報住一晚的價。原版:每人每天的價 × 隊伍人數,再議價。
func (s *State) quoteRest() {
	sess := s.Shop
	p := s.Shops.Prices
	idx := sess.Shop.TypeIndex
	sess.Action = ActionInnRest
	sess.Price = u5data.Haggle(p.Inn[idx]*s.PartySize, s.clerkIntel())
	if line := s.pitch(p.InnRestPitch[idx], sess.Price, 0); line != "" {
		s.Log("「" + line + "」")
	} else {
		s.Log("「住一晚 " + strconv.Itoa(sess.Price) + " 金。」")
	}
	s.Log("「汝要嗎?(Y/N)」")
	sess.Mode = ShopModeConfirm
}

// settleInn 旅店三條路各自的結果。
func (s *State) settleInn() {
	sess := s.Shop
	p := s.Shops.Prices
	idx := sess.Shop.TypeIndex
	switch sess.Action {
	case ActionInnRest:
		s.Log("「祝汝一夜好眠!」")
		// 原版把玩家移到床鋪那一格再睡(byte_570B8 / byte_570C0)。
		s.X, s.Y = p.InnBedX[idx], p.InnBedY[idx]
		s.SleepUntilMorning()
		// 醒來往東踏一步下床(原版 `inc byte_3E0A6`)。
		s.X++
	case ActionInnLeave:
		c := &s.Roster[sess.Target]
		c.Raw[u5data.CharInnFlag] = byte(s.Location)
		c.Raw[u5data.CharInnDays] = 0
		name := c.Name
		// 隊伍是名冊的前綴 —— 把人跟隊尾對調再縮短,順序才不會亂
		// (與入隊 sub_1BB5C 的作法對稱)。
		last := s.PartySize - 1
		s.Roster[sess.Target], s.Roster[last] = s.Roster[last], s.Roster[sess.Target]
		s.PartySize--
		s.Log(name + "留了下來。")
	case ActionInnPick:
		who := sess.Target
		c := &s.Roster[who]
		c.Raw[u5data.CharInnFlag] = 0
		c.Raw[u5data.CharInnDays] = 0
		n := s.PartySize
		s.Roster[who], s.Roster[n] = s.Roster[n], s.Roster[who]
		s.PartySize++
		s.Log(s.Roster[n].Name + "回到了隊伍。")
	}
}

// SleepUntilMorning 睡到早上六點,並套用休息的恢復規則。
//
// 出自原版 `sub_21D48` 的後半 —— 旅店與野外紮營共用同一段:
//
//	睡前:狀態 'G' 的人改成 'S'(睡著)
//	推進:12 × 5 分鐘,之後每次 9 分鐘,直到 byte_3E08F(小時)== 6
//	醒來:HP 回滿;'A'/'M' 的 MP = 智力,'B' 的 MP = 智力 / 2,戰士不變
//	      中毒('P')的人**會死**,狀態改 'D'、HP 歸 0,印「XXX has passed away.」
//	      其餘 'S' 的人改回 'G'
//
// ⚠ 「中毒睡覺會死」是原版的真規則,不是 bug —— 出門前記得解毒。
// 這條沒實作的話,毒在遊戲裡就變成無關痛癢的狀態。
func (s *State) SleepUntilMorning() {
	for _, c := range s.Party() {
		if c.Status == u5data.StatusGood {
			c.Status = u5data.StatusAsleep
			c.Raw[u5data.CharStatus] = u5data.StatusAsleep
		}
	}
	s.Log("Zzzzzz....")
	// 原版先推 12 次 5 分鐘,再每次 9 分鐘直到早上六點。
	for i := 0; i < 12; i++ {
		s.AdvanceTime(5)
	}
	for guard := 0; s.Clock.Hour != WakeHour && guard < 24*60; guard++ {
		// ★ 再生戒指在**每一次 9 分鐘的推進**都擲一次(原版 `loc_21F0C`
		// 的 `call sub_2BCC8` 就在 `sub_29304(9)` 前面)—— 不是每小時一次。
		// 一夜下來擲很多次,但旅店本來就會把 HP 補滿,所以看不出差別;
		// 位置照抄是為了「中毒的人不會被補滿」那條路上仍然有回血。
		s.regenerateParty()
		s.AdvanceTime(9)
	}
	for _, c := range s.Party() {
		s.wakeUp(c)
	}
	s.Log("天亮了。")
	// 醒來時有 1/4 機率遇上那位老人(原版 `sub_165C8` 的 `random(0,99) < 25`)。
	s.MaybeApparition()
}

// WakeHour 是休息結束的時刻(原版 `cmp byte_3E08F, 6`)。
const WakeHour = 6

// wakeUp 套用一名角色醒來時的恢復。
func (s *State) wakeUp(c *u5data.Character) {
	if c.Status == u5data.StatusDead {
		return
	}
	c.HP = c.MaxHP
	// 法力上限就是智力 —— 吟遊詩人只有一半,戰士沒有法力所以不動。
	switch c.Class {
	case 'A', 'M':
		c.MP = c.Intel
	case 'B':
		c.MP = c.Intel / 2
	}
	c.Raw[u5data.CharMP] = c.MP
	if c.Status == u5data.StatusPoisoned {
		c.Status = u5data.StatusDead
		c.Raw[u5data.CharStatus] = u5data.StatusDead
		c.HP = 0
		s.Log(c.Name + "毒發身亡。")
		return
	}
	if c.Status == u5data.StatusAsleep {
		c.Status = u5data.StatusGood
		c.Raw[u5data.CharStatus] = u5data.StatusGood
	}
}

// openShipwright 進造船廠。
//
// 原版 `sub_219B0` 依序問三句:修船(固定 10,000 金)、帆船、小艇,
// 每句都是 Y/N。這裡列成選單,價與判斷照原版。
func (s *State) openShipwright(shop *u5data.Shop) bool {
	p := s.Shops.Prices
	i := shop.TypeIndex
	sess := &ShopSession{Shop: shop, Mode: ShopModeMenu}
	sess.Menu = []ShopItem{
		{Goods: GoodsShip, ID: 0, Name: "Frigate", Base: p.Frigate[i], Qty: 1},
		{Goods: GoodsShip, ID: 1, Name: "Skiff", Base: p.Skiff[i], Qty: 1},
	}
	s.Shop = sess
	s.Prompt = PromptShop
	s.showShopMenu()
	return true
}
