package game

import "github.com/wicanr2/u5-cht/internal/u5data"

// 讀卷軸(原版 `sub_19ED8`,201 行)
//
// # 這裡原本是個洞
//
// `Inventory.Scrolls` 只有 `Get` 會寫,**沒有任何地方讀**;`usableEntries()`
// (U 的清單)也沒列卷軸。⇒ 撿得到、存得進檔、永遠讀不了,八個效果一個都沒接。
//
// 洞是這樣浮出來的:`sub_19ED8` 在 `docs/re/66` 的截斷清單上,
// 而追它的呼叫者 `sub_1A5E8`(U 指令)才發現**清單的前 29 格全是空的**——
// 卷軸(0..7)、藥水(8..15)、月石(21..28)三整族都沒接上。
// 見 `potion.go` 與 `moonstone.go`。
//
// # 八捲卷軸
//
//	byte_3E030[idx]--                       ; ★ 先扣一捲,不論成不成功
//	印 "Scroll"
//	switch (idx):
//	  0 VL  "Light!"          sub_1D310(0F0h) → byte_3E0B6 = 240
//	  1 RH  "Wind change!"    問方向;★ 只有 byte_3E0A3 < 0x21 才真的改
//	  2 IS  "Protection!"     sub_1D31C('P', 100, 音效 2)
//	  3 IA  "Negate magic!"   sub_1D31C('N',  20, 音效 3)
//	  4 IQW "View!"           ★ 戰鬥中印 "Not here!";其餘 sub_EDD4 / sub_F7C0
//	  5 KXC "Summon Daemon!"  ★ **只有戰鬥中**;其餘印 "Not here!"
//	  6 IMC "Resurrection!"   ★ 戰鬥中不行;其餘「On who:」+ sub_1CFC8
//	  7 AT  "Negate time!"    ★ 地點 0x1D / 0x28 印 "No effect!";其餘 sub_1D31C('T', 20, 7)
//
// # ★ 三個「照抄咒語就會寫錯」的地方
//
//  1. **卷軸不用魔力、不用藥草、不看等級。** 這是卷軸存在的意義:一級的角色也能
//     放出復活。所以不能走 `Cast()`,只能借 `spellEffect` 那一層。
//
//  2. **持續時間與同名咒語不一樣,三個都不一樣。**
//
//     | | 卷軸 | 咒語 |
//     |---|---|---|
//     | 光明 | **240**(`sub_1D310(0F0h)`) | In Lor 100 / Vas Lor 255 |
//     | 防護 | **100** | In Sanct 20 |
//     | 抗魔 | **20** | In An 10 |
//     | 停時 | **20** | An Tym 10 |
//
//     照著 `spellEffect` 轉發會把四個時間全弄錯,而且**不會有任何症狀** ——
//     效果照樣發生,只是短了或長了。這種差異只有讀原版數字才抓得到。
//
//  3. **召喚惡魔與復活的場合條件是相反的。** 兩條都比 `byte_3E0A3` 與 0x7F/0x80,
//     但一個 `jbe` 一個 `jnb`:惡魔**只能在戰場上**召,復活**只能在戰場外**做。
//     憑印象寫一定會把兩條寫成同一個方向。
//
// # 回傳值
//
// `ebx` 一開始是 1,只有「換風向但地點 ≥ 0x21」那條會歸零;
// **`Not here!` 三條都留著 1**。所以回傳值不是「效果成功了嗎」,
// 是原版 `sub_1A5E8` 的 `var_10` —— 這一回合算不算用掉了。
//
// ⚠ 還沒查的:地點 **0x28** 是什麼(0x1D 是 STONEGATE)。0x28 同時是地牢
// Doom 的地點編號(`DungeonDoomLocation`),但停時那條是在**非戰鬥**下比的,
// 而地牢裡 `byte_3E0A3` 也真的會是 0x21..0x28 —— 所以「Doom 裡不能停時」是
// 目前最合理的讀法,但沒有第二個證據,先照值實作、不寫進註解當定論。

// 卷軸的八個索引(順序即 `u5data.ScrollSpells`)。
const (
	ScrollLight        = 0 // VL  Vas Lor
	ScrollWindChange   = 1 // RH  Rel Hur
	ScrollProtection   = 2 // IS  In Sanct
	ScrollNegateMagic  = 3 // IA  In An
	ScrollView         = 4 // IQW In Quas Wis
	ScrollSummonDaemon = 5 // KXC Kal Xen Corp
	ScrollResurrection = 6 // IMC In Mani Corp
	ScrollNegateTime   = 7 // AT  An Tym
)

// 卷軸自己的效果強度 —— ★ 與同名咒語不同,見上表。
const (
	ScrollLightTurns      = 0xF0 // sub_1D310(0F0h)
	ScrollProtectTurns    = 0x64 // sub_1D31C('P', 64h, 2)
	ScrollNegateMagicTurn = 0x14 // sub_1D31C('N', 14h, 3)
	ScrollNegateTimeTurns = 0x14 // sub_1D31C('T', 14h, 7)
)

// NegateTimeDeadLocations 是停時卷軸**沒有效果**的兩個地點
// (原版 `cmp al, 1Dh` / `cmp al, 28h`)。
var NegateTimeDeadLocations = [2]int{0x1D, 0x28}

// ReadScroll 讀第 i 捲卷軸(原版 `sub_19ED8`)。
//
// 回傳值同原版的 `ebx`:這一回合算不算用掉了(不是「效果成功了嗎」)。
func (s *State) ReadScroll(i int) bool {
	if i < 0 || i >= u5data.ScrollCount {
		return false
	}
	if s.Inventory.Scrolls[i] <= 0 {
		s.Log(MsgDontHaveThat)
		return false
	}
	// ★ 先扣,不論後面成不成功 —— 原版 `dec byte_3E030[edi]` 在所有判斷之前。
	s.Inventory.Scrolls[i]--
	s.Log(MsgScroll)

	switch i {
	case ScrollLight:
		// ★ 直接指派,不是取大值:原版 `sub_1D310` 就是 `byte_3E0B6 = al`,
		// 所以一捲光明卷軸會把 Vas Lor 的 255 **蓋成 240**。
		s.LightTurns = ScrollLightTurns
		s.Log(MsgScrollLight)
		return true

	case ScrollWindChange:
		s.Log(MsgScrollWindChange)
		// ★ 方向**照樣問**(`sub_1CC50` 在比地點之前),只是地牢與戰鬥中白問。
		ok := s.SceneOrOverworld()
		s.AskDirection(func(d Direction) {
			if ok {
				s.ChangeWind(d)
			}
		})
		return ok

	case ScrollProtection:
		s.Log(MsgScrollProtection)
		return s.setCombatMode(CombatModeProtected, ScrollProtectTurns)

	case ScrollNegateMagic:
		s.Log(MsgScrollNegateMagic)
		return s.setCombatMode(CombatModeNegate, ScrollNegateMagicTurn)

	case ScrollView:
		s.Log(MsgScrollView)
		if s.InCombat() {
			s.Log(MsgNotHere)
			return true
		}
		s.Peer()
		return true

	case ScrollSummonDaemon:
		s.Log(MsgScrollSummonDaemon)
		// ★ 與其他都相反:**只有戰鬥中**能召喚。
		if !s.InCombat() {
			s.Log(MsgNotHere)
			return true
		}
		return s.summonCreature(s.scrollReader(), summonDaemon, 1)

	case ScrollResurrection:
		s.Log(MsgScrollResurrection)
		if s.InCombat() {
			s.Log(MsgNotHere)
			return true
		}
		return s.spellEffect(s.scrollReader(), SpellInManiCorp)

	case ScrollNegateTime:
		for _, loc := range NegateTimeDeadLocations {
			if s.locationCode() == loc {
				s.Log(MsgNoEffect)
				return true
			}
		}
		s.Log(MsgScrollNegateTime)
		return s.setCombatMode(CombatModeTimeStop, ScrollNegateTimeTurns)
	}
	return true
}

// scrollReader 是「誰在讀」。
//
// 原版沒有這一步 —— 卷軸不屬於某個角色,復活那條走的是「On who:」選單
// (`sub_1C9C0`),而召喚惡魔根本不需要施法者。引擎的選單還沒做
// (同 `spellTarget` 的說明),先用**第一個能行動的人**;選單一落地兩邊一起換。
func (s *State) scrollReader() int {
	if i := s.firstAbleMember(); i >= 0 {
		return i
	}
	return 0
}
