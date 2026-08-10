package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// potionScene 是一個「骰子不會走偏」的場景。
//
// ⚠ **這支存在的理由**:每一瓶藥水都有 1/8 機率不照顏色走
// (`random(0,15)` 命中 0 或 1)。要驗顏色本身的效果就得先把骰子挪開,
// 不然測試會**偶爾**紅 —— 而偶爾紅的測試比沒有測試更糟。
// 走偏那條另外有 `TestOneInEightPotionsMisfires` 專門驗。
func potionScene(t *testing.T) *State {
	t.Helper()
	s := upkeepScene(t)
	for i := range s.Inventory.Potions {
		s.Inventory.Potions[i] = 9
	}
	// 骰到 2..15 都是照顏色走。固定種子下先把骰子推到不會命中 0/1 的位置:
	// 直接改成「先抽掉會走偏的那幾顆」不可靠,所以改用重試 —— 見下面的 drink()。
	s.Messages = nil
	return s
}

// drinkUntil 喝到「這一次沒有走偏」為止。
//
// ⚠ **這支是被一條假紅逼出來的**:第一版直接喝一瓶就驗,結果紫藥水那條紅了,
// 訊息卻是「眼前一片通透……」—— 那是**白**藥水的句子。骰子重骰成了顏色 7。
// 換句話說**失敗訊息指向結果,原因在骰子**(這個形狀本專案已經踩到第三次)。
//
// 驗法是驗結果而不是驗骰子:走偏只有兩種結局(變成睡著、或重骰成別的顏色),
// 兩者都會讓「這一瓶該有的判準」不成立。所以最多試 40 次,取第一次照顏色走的。
func drinkUntil(t *testing.T, s *State, colour int, what string, ok func() bool) {
	t.Helper()
	for i := 0; i < 40; i++ {
		s.Messages = nil
		s.Inventory.Potions[colour] = 9
		s.DrinkPotion(colour)
		if ok() {
			return
		}
	}
	t.Fatalf("喝了 40 瓶%s藥水都沒出現「%s」:\n%s",
		u5data.PotionColoursZH[colour], what, strings.Join(s.Messages, "\n"))
}

// drinkNoMisfire 是 drinkUntil 的訊息版。
func drinkNoMisfire(t *testing.T, s *State, colour int, want string) {
	t.Helper()
	drinkUntil(t, s, colour, want, func() bool {
		return strings.Contains(strings.Join(s.Messages, "|"), want)
	})
}

// TestPotionIsConsumedEvenWhenItDoesNothing —— 與卷軸同一條順序:
// `dec byte_3E038[eax]` 在所有判斷之前。
func TestPotionIsConsumedEvenWhenItDoesNothing(t *testing.T) {
	s := potionScene(t)
	s.Inventory.Potions[PotionCurePoison] = 1
	// 沒中毒的人喝解毒藥水 —— 原版是失敗(狀態不是 'P' 就 `esi = 0`)。
	s.Roster[0].Status = u5data.StatusGood
	s.DrinkPotion(PotionCurePoison)
	if n := s.Inventory.Potions[PotionCurePoison]; n != 0 {
		t.Errorf("沒效果卻沒扣掉藥水,還剩 %d 瓶", n)
	}
}

// TestGreenPotionIsPoison 是最容易憑印象寫反的一瓶:綠色是**毒藥**。
func TestGreenPotionIsPoison(t *testing.T) {
	s := potionScene(t)
	s.Roster[0].Status = u5data.StatusGood
	drinkNoMisfire(t, s, PotionPoison, MsgPotionPoisoned)
	if s.Roster[0].Status != u5data.StatusPoisoned {
		t.Errorf("喝綠藥水沒中毒,狀態是 %q", string(s.Roster[0].Status))
	}
	// 反向:已經中毒的人喝綠藥水無效(原版要求狀態是 'G')。
	s2 := potionScene(t)
	s2.Roster[0].Status = u5data.StatusPoisoned
	s2.Inventory.Potions[PotionPoison] = 1
	s2.DrinkPotion(PotionPoison)
	if s2.Roster[0].Status != u5data.StatusPoisoned {
		t.Errorf("狀態被綠藥水改成 %q —— 原版只對 'G' 生效", string(s2.Roster[0].Status))
	}
}

// TestRedPotionCuresPoison 與上面成對:紅色才是解毒。
func TestRedPotionCuresPoison(t *testing.T) {
	s := potionScene(t)
	s.Roster[0].Status = u5data.StatusPoisoned
	drinkNoMisfire(t, s, PotionCurePoison, MsgPotionPoisonCured)
	if s.Roster[0].Status != u5data.StatusGood {
		t.Errorf("紅藥水沒解毒,狀態是 %q", string(s.Roster[0].Status))
	}
}

// TestOrangePotionSleepsTheDrinker —— 橙色照原版要求狀態是 'G'。
func TestOrangePotionSleepsTheDrinker(t *testing.T) {
	s := potionScene(t)
	s.Roster[0].Status = u5data.StatusGood
	drinkNoMisfire(t, s, PotionSleep, MsgPotionSlept)
	if s.Roster[0].Status != u5data.StatusAsleep {
		t.Errorf("橙藥水沒讓人睡著,狀態是 %q", string(s.Roster[0].Status))
	}
}

// TestOneInEightPotionsMisfires 釘住那兩行「在跳表之前」的骰子。
//
// 1/16 變成睡著、1/16 重骰顏色 ⇒ 合起來 1/8。這裡驗的是**它真的會發生**:
// 拿一瓶「對已經中毒的人毫無效果」的紅藥水連喝很多次,
// 若骰子那兩行不存在,結果會是**一次也不會睡著**。
func TestOneInEightPotionsMisfires(t *testing.T) {
	s := potionScene(t)
	s.Roster[0].Status = u5data.StatusPoisoned
	slept := 0
	const tries = 400
	for i := 0; i < tries; i++ {
		s.Roster[0].Status = u5data.StatusGood
		s.Inventory.Potions[PotionCurePoison] = 9
		s.Messages = nil
		s.DrinkPotion(PotionCurePoison)
		if s.Roster[0].Status == u5data.StatusAsleep {
			slept++
		}
	}
	if slept == 0 {
		t.Fatal("喝了 400 瓶解毒藥水一次也沒睡著 —— 走偏那兩行沒實作")
	}
	// 期望值:1/16 直接變睡,加上 1/16 重骰命中橙色的 1/8 → 約 1/16 + 1/128 ≈ 7%。
	// 只驗量級,不驗精確比例(固定種子,樣本 400)。
	if lo, hi := tries/40, tries/4; slept < lo || slept > hi {
		t.Errorf("400 瓶裡睡著 %d 次,超出合理量級 [%d, %d]", slept, lo, hi)
	}
}

// TestPurplePotionOnlyChangesTheSprite:紫藥水把人畫成老鼠,**屬性一格都不動**。
//
// 這是「看起來很像變形其實不是」的那一條 —— 原版只寫物件的兩個 tile 位元組。
func TestPurplePotionOnlyChangesTheSprite(t *testing.T) {
	s := corpserArena(t)
	slot := -1
	for i := 0; i < u5data.CombatPartySlots; i++ {
		if s.Combat.Units[i].IsParty() {
			slot = i
			break
		}
	}
	if slot < 0 {
		t.Fatal("戰場上找不到隊員")
	}
	s.Combat.Turn = slot
	who := s.Combat.Units[slot].Roster
	hp, dex := s.Roster[who].HP, s.Roster[who].Dex
	drinkUntil(t, s, PotionPoof, MsgPotionPoof, func() bool {
		return s.Combat.Units[slot].Tile == PotionRatCombatTile
	})
	if s.Roster[who].HP != hp || s.Roster[who].Dex != dex {
		t.Error("紫藥水動到了屬性 —— 原版只寫 tile")
	}
	// 戰場外只印「沒什麼感覺」。
	s2 := potionScene(t)
	drinkNoMisfire(t, s2, PotionPoof, MsgNoNoticeableEffect)
}

// TestBlackPotionSetsTheHiddenBit —— 黑藥水真正的效果是 `UnitHidden`。
func TestBlackPotionSetsTheHiddenBit(t *testing.T) {
	s := corpserArena(t)
	slot := -1
	for i := 0; i < u5data.CombatPartySlots; i++ {
		if s.Combat.Units[i].IsParty() {
			slot = i
			break
		}
	}
	if slot < 0 {
		t.Fatal("戰場上找不到隊員")
	}
	s.Combat.Turn = slot
	drinkUntil(t, s, PotionInvisible, MsgPotionInvisible, func() bool {
		return s.Combat.Units[slot].Flags&UnitHidden != 0
	})
	if got := s.Combat.Units[slot].Tile; got != combatTileIndex(PartyTileStanding) {
		t.Errorf("黑藥水該把圖寫成站著的最終索引 0x%03X,結果是 0x%03X",
			combatTileIndex(PartyTileStanding), got)
	}
}

// TestWhitePotionOnlyWorksOutsideDungeons —— 白藥水掀開整張遮蔽罩,
// 但只在大地圖與場景;地牢與戰鬥都是「沒什麼感覺」。
func TestWhitePotionOnlyWorksOutsideDungeons(t *testing.T) {
	s := potionScene(t)
	drinkUntil(t, s, PotionReveal, MsgPotionReveal, func() bool { return s.RevealFlash > 0 })

	// 地牢裡:一瓶都不該生效,所以這裡**不重試** —— 連喝 40 瓶都不准掀開。
	s2 := potionScene(t)
	s2.Dungeon = &DungeonState{Index: 0, Location: u5data.DungeonLocationBase}
	for i := 0; i < 40; i++ {
		s2.Inventory.Potions[PotionReveal] = 9
		s2.DrinkPotion(PotionReveal)
		if s2.RevealFlash > 0 {
			t.Fatal("地牢裡的白藥水不該生效")
		}
	}
}
