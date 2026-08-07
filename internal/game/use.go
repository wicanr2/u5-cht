package game

import (
	"fmt"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// Use 指令(原版 `sub_1A5E8`)
//
// 「用道具」是剩下指令裡最雜的一支:候選清單來自 `DATA.OVL 0x04C3` 的
// **特殊道具表**,而每一項的效果各自不同 —— 有的上載具、有的開力場、
// 有的只是報時。
//
// # 道具編號 = 特殊道具表的索引 + 16
//
// 原版的跳表 case 就是這麼來的(`jumptable 0001A6DD case 16` 對上表裡的
// 第 0 筆 `Magic Crpt`)。表的前 22 筆:
//
//	 0 Magic Crpt     → case 16   魔毯
//	 1 Skull Keys     → case 17   骷髏鑰匙
//	 2 Amulet         → case 18   不列顛王的護符
//	 3 Crown          → case 19   王冠
//	 4 Sceptre        → case 20   權杖
//	 5..12 "(0".."(7" → case 21..28  **不可用**(原版跳到 default)
//	13..15 Shard/…    → case 29..31 三塊碎片
//	16 Spyglass       → case 32   望遠鏡
//	17 HMS Cape Plan  → case 33   那份圖紙
//	18 Sextant        → case 34   六分儀
//	19 Pocket Watch   → case 35   懷錶
//	20 Black Badge    → case 36   黑徽章
//	21 Wooden Box     → case 37   木盒
//
// ⚠ 5..12 那八筆的名字是 `(0`..`(7` —— 看起來像資料損毀,其實是原版留的
// 佔位名(它們在跳表裡走 default,根本不會被用),**不要試著「修好」它們**。

// 可用道具的 case 編號。
const (
	UseCarpet     = 16
	UseSkullKey   = 17
	UseAmulet     = 18
	UseCrown      = 19
	UseSceptre    = 20
	UseShardFirst = 29
	UseShardLast  = 31
	UseSpyglass   = 32
	UsePlans      = 33
	UseSextant    = 34
	UseWatch      = 35
	UseBadge      = 36
	UseWoodenBox  = 37
)

// 信物穿戴時寫進全域模式位元組的碼(原版 `sub_1D31C(模式, 回合, 音效)`)。
//
// ★ **與四個咒語共用同一個位元組**(`byte_3E08A`,見 `State.CombatMode`)。
// 所以戴上王冠會蓋掉 In Sanct、而 An Tym 也會蓋掉王冠 —— 原版就是一個位元組,
// 不是一組旗標。這很容易寫成獨立的布林值,然後行為就跟原版不一樣了。
const (
	ModeAmulet = 0x0E
	ModeCrown  = 0x1C
	// 信物的持續回合是 0xFF —— 實務上等於「戴著就一直有效」。
	regaliaTurns = 0xFF
)

// Use 是 U 指令:用一件特殊道具。
//
// item 是上面那組 case 編號。
func (s *State) Use(item int) bool {
	switch {
	case item == UseCarpet:
		return s.useCarpet()
	case item == UseSkullKey:
		return s.useSkullKey()
	case item == UseAmulet:
		return s.wearRegalia(s.Regalia.Amulet, ModeAmulet, MsgUseAmulet)
	case item == UseCrown:
		return s.wearRegalia(s.Regalia.Crown, ModeCrown, MsgUseCrown)
	case item == UseSceptre:
		return s.useSceptre()
	case item == UseSpyglass:
		return s.useSpyglass()
	case item == UsePlans:
		return s.usePlans()
	case item == UseSextant:
		return s.useSextant()
	case item == UseWatch:
		return s.useWatch()
	case item == UseBadge:
		return s.useBadge()
	case item >= UseShardFirst && item <= UseShardLast:
		// 碎片的用法是「丟進聖火」,而那條路走的是聖火那一支
		//(`docs/re/26`),不是這裡。這裡只擋住「身上沒有」。
		if !s.Shards[item-UseShardFirst] {
			s.Log(MsgDontHaveThat)
			return false
		}
		s.Log(MsgShardOnlyAtFlame)
		return false
	case item == UseWoodenBox:
		// 原版印「Box- How?」再依答案分支,而那條路還沒逆完。
		s.Log(MsgBoxHow)
		return false
	}
	s.Log(MsgNoUsableItems)
	return false
}

// useCarpet 攤開魔毯(原版 case 16)。
//
// 三道前置照原版:在船上要先下船、必須步行、而且不是每個地方都攤得開。
func (s *State) useCarpet() bool {
	if s.Inventory.Carpets <= 0 {
		s.Log(MsgDontHaveThat)
		return false
	}
	if u5data.VehicleKind(s.Transport) == u5data.VehicleShip ||
		u5data.VehicleKind(s.Transport) == u5data.VehicleSailing {
		s.Log(MsgXitShipFirst)
		return false
	}
	if !u5data.IsOnFoot(s.Transport) {
		s.Log(MsgOnlyOnFoot)
		return false
	}
	if s.InScene() || s.InDungeon() {
		s.Log(MsgNotHere)
		return false
	}
	s.Transport = u5data.VehicleCarpet
	s.Log(MsgBoarded)
	return true
}

// useSkullKey 用骷髏鑰匙(原版 case 17):化掉眼前的力場。
//
// ⚠ 骷髏鑰匙是**一次性**的:用掉就沒了。原版把它記在「怪鑰匙」那個計數
//(`byte_3DFBD`)—— 名字奇怪,但它就是這個東西。
func (s *State) useSkullKey() bool {
	if s.Inventory.OddKeys <= 0 {
		s.Log(MsgDontHaveThat)
		return false
	}
	if !s.destroyField() {
		s.Log(MsgNotHere)
		return false
	}
	s.Inventory.OddKeys--
	s.Log(MsgSkullKey)
	return true
}

// wearRegalia 戴上或脫下護符 / 王冠(原版 `sub_1A5B0` + `sub_1D31C`)。
//
// 再用一次就脫下來(原版印「Removed!」)—— 它是切換而不是單向。
func (s *State) wearRegalia(have bool, mode byte, msg string) bool {
	if !have {
		s.Log(MsgDontHaveThat)
		return false
	}
	if s.CombatMode == mode {
		s.CombatMode, s.CombatModeTurns = 0, 0
		s.Log(MsgRemoved)
		return true
	}
	s.CombatMode, s.CombatModeTurns = mode, regaliaTurns
	s.Log(msg)
	return true
}

// useSceptre 舉起權杖(原版 case 20):化掉力場,而且**不設模式**。
//
// ⚠ 權杖與護符 / 王冠不同 —— 它不寫模式位元組,只放三個音效然後化力場。
// 寫成「跟王冠一樣設一個模式」會多出一個原版沒有的持續效果。
func (s *State) useSceptre() bool {
	if !s.Regalia.Sceptre {
		s.Log(MsgDontHaveThat)
		return false
	}
	s.Log(MsgUseSceptre)
	if s.destroyField() {
		s.Log(MsgFieldDissolved)
		return true
	}
	s.Log(MsgNoEffect)
	return false
}

// useSpyglass 用望遠鏡(原版 case 32):看星象。
//
// 兩道前置:要在戶外、而且要看得到星星(也就是夜裡)。
func (s *State) useSpyglass() bool {
	if s.InScene() || s.InDungeon() {
		s.Log(MsgNotHere)
		return false
	}
	if s.isDaylight() {
		s.Log(MsgNoStars)
		return false
	}
	s.Log(MsgLooking)
	return true
}

// usePlans 用那份圖紙(原版 case 33):船速加倍。
//
// ⚠ 只在船上有用。這是 U5 的航海捷徑,而它**不是一次性的** ——
// 原版沒有扣除任何東西,圖紙留在身上。
func (s *State) usePlans() bool {
	if !s.Regalia.Plans {
		s.Log(MsgDontHaveThat)
		return false
	}
	if u5data.VehicleKind(s.Transport) != u5data.VehicleShip &&
		u5data.VehicleKind(s.Transport) != u5data.VehicleSailing {
		s.Log(MsgOnlyOnShipboard)
		return false
	}
	s.ShipRigged = true
	s.Log(MsgShipRigged)
	return true
}

// useSextant 用六分儀(原版 case 34):報出座標。
//
// 兩道前置:戶外、夜裡(要看得到星星才測得出位置)。
func (s *State) useSextant() bool {
	if s.InScene() || s.InDungeon() {
		s.Log(MsgOnlyOutdoors)
		return false
	}
	if s.isDaylight() {
		s.Log(MsgOnlyAtNight)
		return false
	}
	// 原版印的是 chunk 座標(把 256×256 的世界切成 16×16 塊之後的塊號),
	// 不是格座標 —— 那才是手冊上那張地圖的格線。
	s.Log(fmt.Sprintf("%s%d, %d", MsgPosition, s.X/u5data.ChunkSide, s.Y/u5data.ChunkSide))
	return true
}

// useWatch 看懷錶(原版 case 35):報時,格式與老爺鐘相同。
func (s *State) useWatch() bool {
	s.Log(MsgPocketWatch + s.clockFace())
	return true
}

// useBadge 戴上黑徽章(原版 case 36)。
//
// 徽章與「戰鬥模式的咒語」共用同一個位元組 —— 衛兵盤查那一支
//(`docs/re/32`)讀的就是它。所以戴徽章會蓋掉戰鬥模式的咒語,反之亦然。
func (s *State) useBadge() bool {
	if !s.HasBadge {
		s.Log(MsgDontHaveThat)
		return false
	}
	s.Log(MsgBadgeWorn)
	return true
}

// isDaylight 回報現在是不是白天。與抬頭看天用同一組邊界(6 時 ≤ hour < 18 時)。
func (s *State) isDaylight() bool {
	return s.Clock.Hour >= skyDayFrom && s.Clock.Hour < skyDayUntil
}
