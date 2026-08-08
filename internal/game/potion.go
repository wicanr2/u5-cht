package game

import "github.com/wicanr2/u5-cht/internal/u5data"

// 喝藥水(原版 `sub_1A0B0`)
//
// 與卷軸同一個洞:`Inventory.Potions` 只有 `Get` 會寫,從來沒有人讀。
// 八色藥水撿得到、存得進檔、喝不了。
//
// # ★ 最容易漏掉的一段:每一瓶都有 1/8 機率不照顏色走
//
//	var_4 = random(0, 15)
//	if (var_4 == 0)      顏色 = 4            ; ★ 1/16 → 不論喝哪一瓶都變成「睡著」
//	else if (var_4 == 1) 顏色 = random(0, 7) ; ★ 1/16 → 效果整個重骰
//
// 這兩行在跳表**之前**,所以是對全部八色生效的。合起來 **1/8 的瓶子會出錯**,
// 其中一半直接把喝的人打昏。這是原版「不明藥水」風險感的來源,
// 而它完全不在任何攻略的顏色對照表裡 —— 只有讀碼才看得到。
//
// # 八色(順序即 `u5data.PotionColours`)
//
//	0 藍 Awaken        狀態 'S' → 'G';★ 戰場上若正是行動中的那個單位,還要把他站起來
//	1 黃 Healed!       sub_1CD3C = Mani 的同一支(回 1..30,上限 MaxHP)
//	2 紅 Poison cured! 'P' → 'G'
//	3 綠 POISONED!     'G' → 'P'   ★ 綠色是**毒藥**,喝了會中毒
//	4 橙 Slept!        必須是 'G';戰場外設 'S',戰場上讓那個單位躺下
//	5 紫 Poof!         ★ 只有戰鬥中;把喝的人的 tile 換成 0x90 = 老鼠
//	6 黑 Invisible!    ★ 只有戰鬥中;旗標 |= 0x10(UnitHidden)、tile 換成 0x1D
//	7 白 (無訊息)      ★ 只有 byte_3E0A3 < 0x21;sub_1CE0C = 把視線遮蔽罩整片填掉
//
// ★ 紫色的 0x90 不是猜的:`0x90 = CreatureBase + 0x14*4`,而 `0x14` 正是
// Kal Xen 召喚的那一種(`summonRat`)、也是 Rel Xen Bet 變形寫進去的同一個值。
// 三處獨立命中同一個編號 ⇒ 紫藥水把人變成老鼠。**而且只換 tile** ——
// 原版只寫物件的兩個 tile 位元組(`[ebx]`、`[ebx+1]`),屬性一格都沒動。
// 所以它是純外觀,不是變形。這種「看起來很像變形其實不是」的事,
// 照印象寫必錯。
//
// ★ 白色的 `sub_1CE0C` 是 `sub_2E0E8(-1, …)` —— 半徑 −1 = 把整張視線罩填成
// 可見(`docs/re/17` §496、`docs/re/31`)。所以白藥水**沒有回合數**:
// 它把遮蔽掀開,重畫 20 幀(`sub_DCAC`+`sub_297F4`+`sub_29BEC` 跑 20 次),
// 最後 `sub_29D64` 收尾恢復正常 —— 是**一瞬間看穿**,不是持續效果。
// ⚠ 那 20 幀在原版是一個阻塞迴圈(期間遊戲不前進)。引擎沒有阻塞動畫層,
// 近似成「到下一個動作為止」(`RevealFlash`),差異記在 `docs/re/71`。
//
// ★ 黑色寫的 0x1D 不是不明編號:`combat.go` 的 `PartyTileStanding` 就是它
// (`docs/re/53` 已釘死站著 0x1D / 躺著 0x1E)。所以黑藥水順手把人「扶起來」,
// 真正的效果是那個 `UnitHidden` 位元。

// 藥水的八個顏色索引(順序即 `u5data.PotionColours`)。
const (
	PotionAwaken     = 0 // 藍
	PotionHeal       = 1 // 黃
	PotionCurePoison = 2 // 紅
	PotionPoison     = 3 // 綠 —— ★ 是毒藥
	PotionSleep      = 4 // 橙
	PotionPoof       = 5 // 紫
	PotionInvisible  = 6 // 黑
	PotionReveal     = 7 // 白
)

// 藥水走偏的兩個骰值(原版 `sub_28E14(0, 0Fh)` 的結果)。
const (
	PotionMisfireRollMax = 15 // random(0, 15)
	PotionMisfireSleep   = 0  // → 效果變成「睡著」
	PotionMisfireRandom  = 1  // → 效果重骰成 0..7
)

// PotionRatTile 是紫藥水寫進物件 tile 的值(= 老鼠)。
const PotionRatTile = u5data.CreatureBase + summonRat*4

// PotionInvisibleTile 是黑藥水寫進物件 tile 的值 —— 就是站著的隊員圖。
const PotionInvisibleTile = PartyTileStanding

// DrinkPotion 喝第 i 色藥水(原版 `sub_1A0B0`)。
//
// 回傳值同原版的 `esi`:這一回合算不算用掉了。
func (s *State) DrinkPotion(i int) bool {
	if i < 0 || i >= u5data.PotionCount {
		return false
	}
	if s.Inventory.Potions[i] <= 0 {
		s.Log(MsgDontHaveThat)
		return false
	}
	// ★ 先扣一瓶,與卷軸同 —— 原版 `dec byte_3E038[eax]` 在所有判斷之前。
	s.Inventory.Potions[i]--
	s.Log(MsgPotion)

	// 喝的人:戰場上是**此刻行動中的那個單位**;戰場外原版問「On who:」。
	who := s.potionDrinker()
	if who < 0 {
		return false
	}

	// ★ 1/8 的瓶子不照顏色走。
	switch s.Roll(0, PotionMisfireRollMax) {
	case PotionMisfireSleep:
		i = PotionSleep
	case PotionMisfireRandom:
		i = s.Roll(0, u5data.PotionCount-1)
	}

	ch := s.rosterAt(who)
	if ch == nil {
		return false
	}
	switch i {
	case PotionAwaken:
		if ch.Status != u5data.StatusAsleep {
			return false
		}
		ch.Status = u5data.StatusGood
		// ★ 戰場上還要把那個單位真的站起來(`sub_2ED50`)—— 但只在「他就是
		// 行動中的那個」而且旗標正好是「隊員 + 睡著、沒死也不是怪物」
		// (`flags & 0E8h == 88h`)時。那個遮罩比對是原版的原話,不要簡化成
		// 「有沒有睡著」:0xE8 同時排除了 UnitDead 與 UnitMonster。
		if s.InCombat() {
			slot := s.combatSlotOfRoster(who)
			if slot >= 0 && slot == s.Combat.Turn &&
				s.Combat.Units[slot].Flags&(UnitParty|UnitMonster|UnitAsleep|UnitDead) == UnitParty|UnitAsleep {
				s.Combat.Units[slot].Flags &^= UnitAsleep
				s.Combat.Units[slot].Tile = PartyTileStanding
			}
		}
		return true

	case PotionHeal:
		// 原版走 `sub_1CD3C` —— 與 Mani 同一支(回 1..30)。
		if !s.healTarget(who, s.AttackRoll()) {
			return false
		}
		s.Log(MsgPotionHealed)
		return true

	case PotionCurePoison:
		if ch.Status != u5data.StatusPoisoned {
			return false
		}
		ch.Status = u5data.StatusGood
		s.Log(MsgPotionPoisonCured)
		return true

	case PotionPoison:
		if ch.Status != u5data.StatusGood {
			return false
		}
		ch.Status = u5data.StatusPoisoned
		s.Log(MsgPotionPoisoned)
		return true

	case PotionSleep:
		if ch.Status != u5data.StatusGood {
			return false
		}
		// ⚠ **兩條路都會把名冊狀態設成 'S'**:戰鬥外是這裡直接寫,戰鬥中是
		// `sub_2EDF8`(躺下)自己第一行就寫(`docs/re/53`)。看 `sub_1A0B0`
		// 會以為戰鬥中不動狀態 —— 那是因為那一行藏在被呼叫的函式裡。
		ch.Status = u5data.StatusAsleep
		if s.InCombat() {
			if slot := s.combatSlotOfRoster(who); slot >= 0 {
				s.Combat.Units[slot].Flags |= UnitAsleep
				s.Combat.Units[slot].Tile = PartyTileLying
			}
		}
		s.Log(MsgPotionSlept)
		return true

	case PotionPoof:
		if !s.InCombat() {
			s.Log(MsgNoNoticeableEffect)
			return true
		}
		s.Log(MsgPotionPoof)
		return s.retileDrinker(who, PotionRatTile)

	case PotionInvisible:
		if !s.InCombat() {
			s.Log(MsgNoNoticeableEffect)
			return true
		}
		s.Log(MsgPotionInvisible)
		if slot := s.combatSlotOfRoster(who); slot >= 0 {
			s.Combat.Units[slot].Flags |= UnitHidden
		}
		return s.retileDrinker(who, PotionInvisibleTile)

	case PotionReveal:
		// ★ 只有大地圖與場景;地牢與戰鬥都印「沒什麼感覺」。
		if !s.SceneOrOverworld() {
			s.Log(MsgNoNoticeableEffect)
			return true
		}
		s.RevealFlash = 1
		s.Log(MsgPotionReveal)
		return true
	}
	return true
}

// potionDrinker 是「誰喝下去」。
//
// 原版:戰鬥中是 `dword_3EF50[byte_3E0AE]` 的角色(= 此刻行動的單位),
// 戰鬥外印「On who: 」讓玩家選(`sub_1C9C0`)。選單還沒做,戰鬥外沿用
// `spellTarget` 的同一套近似(傷得最重的那個),兩邊之後一起換。
func (s *State) potionDrinker() int {
	if c := s.Combat; c != nil {
		if c.Turn >= 0 && c.Turn < CombatUnitSlots && c.Units[c.Turn].IsParty() {
			return c.Units[c.Turn].Roster
		}
		return -1
	}
	return s.spellTarget(s.firstAbleMember(), false)
}

// retileDrinker 換掉戰場上那個人畫出來的樣子(原版寫物件的 `[+0]` 與 `[+1]`)。
func (s *State) retileDrinker(who, tile int) bool {
	slot := s.combatSlotOfRoster(who)
	if slot < 0 {
		return false
	}
	s.Combat.Units[slot].Tile = tile
	return true
}
