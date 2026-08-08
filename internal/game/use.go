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
// # 清單其實有 38 格,不是 22 格
//
// 清單本身是 `sub_1E8D4` 從六個地方**抄**進 `byte_40BA0` 的一整段:
//
//	+0 ..+7   byte_3E030[8]   八種卷軸的持有數           → sub_19ED8(i)
//	+8 ..+15  byte_3E038[8]   八色藥水的持有數           → sub_1A0B0(i−8)
//	+16       byte_3DFBC      魔毯      ┐
//	+17       byte_3DFBD      骷髏鑰匙  │
//	+18       byte_3DFBF      護符      ├ case 16..20
//	+19       byte_3DFC0      王冠      │
//	+20       byte_3DFC1      權杖      ┘
//	+21..+28  byte_3E050[8]   八顆月石(== 0xFF 才列)   → sub_1A2F8(i−21)
//	+29..+31  byte_3DFC4[3]   三塊碎片    ┐
//	+32       byte_3DFC8      望遠鏡      │
//	+33       byte_3DFC9      圖紙        │
//	+34       byte_3DFCA      六分儀      ├ case 29..37
//	+35       byte_3DFCB      懷錶        │
//	+36       byte_3DFCC      徽章        │
//	+37       byte_3DFCD      檀香木盒    ┘
//
// ⚠ 後六筆的順序**用跳表自己的 case 標註核對過**(`aSpyglass` = case 32、
// `aPlans` = 33、`aSextant` = 34、`aWatchThePocket` = 35、`aBadge` = 36、
// 木盒 = 37),不是照 `sub_1E8D4` 的抄寫順序猜的 —— 我第一版就猜錯了兩筆。
// 另一個獨立佐證是存檔:`SavePlansOffset = 0x0215 ↔ byte_3DFC9`、
// 已驗過的檀香木盒 `0x0219 ↔ byte_3DFCD`,兩端夾住中間四筆。
//
// 分派在 `sub_1A5E8`,而**前三段是在進跳表之前就被接走的**:
//
//	if (n < 8)   sub_19ED8(n)          ; 卷軸
//	if (n < 16)  sub_1A0B0(n − 8)      ; 藥水
//	if (n > 20 && n < 29) sub_1A2F8(n − 21)  ; 月石
//	否則 switch (n − 16) 的 22 格跳表
//
// ⚠⚠ **更正(2026-08-08)**:這裡原本寫
//
//	5..12 "(0".."(7" → case 21..28  **不可用**(原版跳到 default)
//	⚠ … 它們在跳表裡走 default,根本不會被用,**不要試著「修好」它們**
//
// 跳表那八格**確實**指向 `def_1A6DD` —— 但那是因為 21..28 已經被上面那三行
// 接去 `sub_1A2F8`(埋月石)了。只讀跳表會讀成「沒用」;真相在跳表**前面**六行。
// 名字 `(0`..`(7` 也不是損毀,是八顆月石的短名(`(` 在原版字型裡是月相符號)。
//
// ⇒ 引擎原本只接了 22 格中的 14 格,**卷軸、藥水、月石三整族撿得到卻用不了**。
// 見 `scroll.go` / `potion.go` / `moonstone.go` 與 `docs/re/71`。

// 可用道具的 case 編號。
const (
	// 前三段:編號**不加 16**,它們在跳表之前就被接走。
	UseScrollFirst = 0
	UseScrollLast  = 7
	UsePotionFirst = 8
	UsePotionLast  = 15

	UseCarpet     = 16
	UseSkullKey   = 17
	UseAmulet     = 18
	UseCrown      = 19
	UseSceptre = 20
	// 月石也是在跳表之前接走的那一段(21..28)。
	UseMoonstoneFirst = 21
	UseMoonstoneLast  = 28

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
	// ★ 三段在跳表之前:順序與原版一致(卷軸 → 藥水 → 月石 → 跳表)。
	//
	// ⚠ **只有這兩段的回傳值會被讀**:原版 `var_10` 一開始是 1,
	// 只有卷軸與藥水那兩個分支會把它換成函式的回傳值,而結尾
	// `if (var_10 == 0) 印 "Failed!"`。月石與其餘道具都跳過那個賦值,
	// 所以永遠不會印。⇒ **「Failed!」是卷軸與藥水專屬的收尾**。
	case item >= UseScrollFirst && item <= UseScrollLast:
		return s.useOrFail(s.ReadScroll(item - UseScrollFirst))
	case item >= UsePotionFirst && item <= UsePotionLast:
		return s.useOrFail(s.DrinkPotion(item - UsePotionFirst))
	case item >= UseMoonstoneFirst && item <= UseMoonstoneLast:
		return s.BuryMoonstone(item - UseMoonstoneFirst)
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
		// ⚠⚠ **更正**:這裡原本寫「原版印『Box- How?』再依答案分支,
		// 而那條路還沒逆完」,並印一句「(木盒的用法尚未實作)」。
		//
		// **沒有那條路。** `sub_1A5E8` 的 case 37(`loc_1AB32`)整段只有兩行:
		//
		//	push offset aBoxHow ; call sub_23C18 ; jmp loc_1AB3C
		//
		// 而那個字串是 `aBoxHow db 'Box',0Ah` —— **就只有 "Box"**。
		// IDA 的自動命名 `aBoxHow` 把鄰居的字尾黏進來了,而我把那個
		// **工具產生的名字**當成了原版的字串內容。
		//
		// 檀香木盒的用途在別處(不列顛王城堡的 NPC 那條線,`docs/re/36`),
		// U 對它就是印一個名字。⇒ 這不是缺口,是一個假缺口(`docs/re/77`)。
		s.Log(MsgItemWoodenBox)
		return true
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

// useOrFail 是原版 U 的收尾(`sub_1A5E8` 的 `if (var_10 == 0) 印 "Failed!"`)。
//
// ⚠ 只有卷軸與藥水會走到這裡 —— 見 `Use` 裡的說明。
func (s *State) useOrFail(ok bool) bool {
	if !ok {
		s.Log(MsgFailed)
	}
	return ok
}
