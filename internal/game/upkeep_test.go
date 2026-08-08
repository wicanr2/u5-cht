package game

import (
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

func upkeepScene(t *testing.T) *State {
	t.Helper()
	s := newCreateState(t)
	s.MaxMessages = 32
	s.mealHour = newUpkeepHour
	if s.PartySize < 1 {
		s.PartySize = 1
	}
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		s.Roster[i].Status = u5data.StatusGood
		s.Roster[i].HP = 200
	}
	s.Inventory.Food = 100
	s.Messages = nil
	return s
}

// TestPoisonCostsOneHitPointPerTurn:中毒每回合掉 1 血。
//
// ⚠ 引擎在此之前**中毒完全不扣血** —— 狀態掛著卻沒有代價。
func TestPoisonCostsOneHitPointPerTurn(t *testing.T) {
	s := upkeepScene(t)
	s.Roster[0].Status = u5data.StatusPoisoned
	s.Clock.Hour = 3 // 不是用餐時刻
	for i := 0; i < 10; i++ {
		s.upkeep()
	}
	if got := 200 - int(s.Roster[0].HP); got != 10*PoisonDamagePerTurn {
		t.Errorf("中毒 10 回合掉了 %d 血,預期 %d", got, 10*PoisonDamagePerTurn)
	}
	// 沒中毒的人不該掉血。
	if s.PartySize > 1 && s.Roster[1].HP != 200 {
		t.Errorf("沒中毒的隊員掉了 %d 血", 200-s.Roster[1].HP)
	}
}

// TestFoodIsEatenAtSixTwelveAndEighteen 釘住那三個時刻。
func TestFoodIsEatenAtSixTwelveAndEighteen(t *testing.T) {
	for hour := 0; hour < 24; hour++ {
		s := upkeepScene(t)
		s.Clock.Hour = hour
		s.upkeep()
		eaten := 100 - s.Inventory.Food
		meal := hour == 6 || hour == 12 || hour == 18
		if meal && eaten != s.PartySize {
			t.Errorf("%d 點吃了 %d 份,預期 %d 份(隊伍人數)", hour, eaten, s.PartySize)
		}
		if !meal && eaten != 0 {
			t.Errorf("%d 點竟然吃了 %d 份 —— 只有 6 / 12 / 18 點才扣糧", hour, eaten)
		}
	}
}

// TestSameHourOnlyEatsOnce:同一個小時走幾百步也只扣一次糧。
func TestSameHourOnlyEatsOnce(t *testing.T) {
	s := upkeepScene(t)
	s.Clock.Hour = 12
	for i := 0; i < 200; i++ {
		s.upkeep()
	}
	if eaten := 100 - s.Inventory.Food; eaten != s.PartySize {
		t.Errorf("同一小時走 200 步吃了 %d 份,預期 %d 份", eaten, s.PartySize)
	}
}

// TestSleepingAndDeadMembersDoNotEat:睡著的人與死人都不吃。
func TestSleepingAndDeadMembersDoNotEat(t *testing.T) {
	s := upkeepScene(t)
	if s.PartySize < 2 {
		t.Skip("隊伍太小")
	}
	s.Roster[0].Status = u5data.StatusAsleep
	s.Roster[1].Status = u5data.StatusDead
	s.Clock.Hour = 6
	s.upkeep()
	want := s.PartySize - 2
	if eaten := 100 - s.Inventory.Food; eaten != want {
		t.Errorf("吃了 %d 份,預期 %d 份(睡著的與死人都不吃)", eaten, want)
	}
}

// TestStarvingHurtsEveryTurn:斷糧是**每回合**掉血,不是一天三次。
func TestStarvingHurtsEveryTurn(t *testing.T) {
	s := upkeepScene(t)
	s.Inventory.Food = 0
	s.Clock.Hour = 3 // 不是用餐時刻,照樣要餓
	s.upkeep()
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgStarving) {
		t.Fatalf("存糧 0 卻沒挨餓:%q", s.Messages)
	}
	hurt := 0
	for i := 0; i < s.PartySize; i++ {
		if s.Roster[i].HP < 200 {
			hurt++
		}
	}
	if hurt != s.PartySize {
		t.Errorf("只有 %d 位掉血,預期全隊 %d 位", hurt, s.PartySize)
	}
	// 再走一步還要再餓一次(同一個小時內也算)。
	s.Messages = nil
	s.upkeep()
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgStarving) {
		t.Errorf("第二回合沒再挨餓 —— 原版不更新 byte_3E090,所以每回合都判:%q", s.Messages)
	}
}

// TestOnlyTwoRingsVanish 是這一條最容易寫錯的地方。
//
// 原版只比 0x2A(隱形)與 0x2C(再生)兩個值 —— **防護戒指(0x2B)不會消失**。
func TestOnlyTwoRingsVanish(t *testing.T) {
	cases := []struct {
		ring    byte
		vanishe bool
		name    string
	}{
		{u5data.ItemRingInvisibility, true, "隱形戒指"},
		{u5data.ItemRingProtection, false, "★ 防護戒指 —— 原版不比這個值"},
		{u5data.ItemRingRegeneration, true, "再生戒指"},
		{u5data.ItemNone, false, "沒戴戒指"},
	}
	// ⚠ **同一個 State 重複擲**,不要每次都 `upkeepScene(t)`:
	// 新 State 的骰子種子是固定的 1(見 `SeedRandom` 的說明),
	// 每次重建就永遠拿到同一個骰值 —— 測試會假紅。
	s := upkeepScene(t)
	for _, c := range cases {
		gone := false
		for try := 0; try < 400 && !gone; try++ {
			s.Roster[0].Raw[u5data.CharRing] = c.ring
			s.vanishRings()
			if s.Roster[0].Raw[u5data.CharRing] != c.ring {
				gone = true
			}
		}
		if gone != c.vanishe {
			t.Errorf("%s(0x%02X):會消失 = %v,預期 %v", c.name, c.ring, gone, c.vanishe)
		}
	}
}

// TestRingVanishRollIsOneInSixteen:機率是 1/16,而判準是 `== 11` 不是 `< 1`。
func TestRingVanishRollIsOneInSixteen(t *testing.T) {
	if RingVanishRollMax != 15 || RingVanishHit != 11 {
		t.Fatalf("判準是 rand(0, %d) == %d,預期 rand(0, 15) == 11",
			RingVanishRollMax, RingVanishHit)
	}
	// 骰的範圍是 0..15 共 16 個值,命中一個 → 1/16。
	if RingVanishHit < 0 || RingVanishHit > RingVanishRollMax {
		t.Error("命中值落在骰的範圍外 —— 那樣戒指永遠不會消失")
	}
	s := upkeepScene(t)
	gone, tries := 0, 4000
	for i := 0; i < tries; i++ {
		s.Roster[0].Raw[u5data.CharRing] = u5data.ItemRingInvisibility
		s.vanishRings()
		if s.Roster[0].Raw[u5data.CharRing] == u5data.ItemNone {
			gone++
		}
	}
	// 1/16 ≈ 250/4000。放寬到 150..400 只是擋「機率差一個數量級」。
	if gone < 150 || gone > 400 {
		t.Errorf("4000 次消失了 %d 次,1/16 該落在 150..400", gone)
	}
}

// terrainScene 造一個站在指定地形上的隊伍(需要真的世界地圖才寫得進 tile)。
func terrainScene(t *testing.T, tile byte) *State {
	t.Helper()
	s := upkeepScene(t)
	s.Location, s.Floor = 0, 0
	s.Transport = u5data.VehicleWalk
	if dir := envGamedata(); dir != "" {
		if w, err := u5data.LoadFlatMap(dir + "/UNDER.DAT"); err == nil {
			s.World, s.Under = w, w
		}
	}
	s.X, s.Y = 20, 20
	if !s.SetTileAt(s.X, s.Y, tile) {
		t.Skip("寫不進世界地圖")
	}
	s.Messages = nil
	return s
}

// TestSwampPoisonsOnlyOnFootAndOnlyAboveDexterity 釘住兩個閘門。
func TestSwampPoisonsOnlyOnFootAndOnlyAboveDexterity(t *testing.T) {
	// 敏捷 0 → 一定中毒。
	s := terrainScene(t, TileSwamp)
	for i := 0; i < s.PartySize; i++ {
		s.Roster[i].Dex = 0
	}
	s.terrainEffects()
	for i := 0; i < s.PartySize; i++ {
		if s.Roster[i].Status != u5data.StatusPoisoned {
			t.Errorf("第 %d 位敏捷 0 走過沼澤卻沒中毒", i)
		}
	}

	// ★ 敏捷 30 以上完全免疫(判準是 `rand(0,29) > 敏捷`,大於才中毒)。
	s = terrainScene(t, TileSwamp)
	for i := 0; i < s.PartySize; i++ {
		s.Roster[i].Dex = 30
	}
	for k := 0; k < 200; k++ {
		s.terrainEffects()
	}
	for i := 0; i < s.PartySize; i++ {
		if s.Roster[i].Status == u5data.StatusPoisoned {
			t.Errorf("第 %d 位敏捷 30 卻中毒了 —— 骰上限是 %d,大於才中毒",
				i, SwampPoisonRollMax)
		}
	}

	// ★ 只有步行會中毒 —— 而且原版比的是**單一值** 0x1C。
	for _, tr := range []byte{
		u5data.VehicleWalk + 1, u5data.TileHorse | 2,
		u5data.VehicleCarpet, u5data.VehicleSkiff,
	} {
		s = terrainScene(t, TileSwamp)
		s.Transport = tr
		for i := 0; i < s.PartySize; i++ {
			s.Roster[i].Dex = 0
		}
		s.terrainEffects()
		for i := 0; i < s.PartySize; i++ {
			if s.Roster[i].Status == u5data.StatusPoisoned {
				t.Errorf("載具 0x%02X 竟然也會中毒", tr)
				break
			}
		}
	}
}

// TestLavaAndFireplaceBurn:兩個 tile 都會燒。
func TestLavaAndFireplaceBurn(t *testing.T) {
	for _, tile := range []byte{TileLava, TileFireplace} {
		s := terrainScene(t, tile)
		s.terrainEffects()
		if !strings.Contains(strings.Join(s.Messages, "|"), MsgBurning) {
			t.Errorf("tile 0x%02X 沒燒起來:%q", tile, s.Messages)
		}
		hurt := false
		for i := 0; i < s.PartySize; i++ {
			if s.Roster[i].HP < 200 {
				hurt = true
			}
		}
		if !hurt {
			t.Errorf("tile 0x%02X 燒了卻沒人受傷", tile)
		}
	}
}

// TestTrapdoorSkipsTheMeal 是這一條最容易漏的:掉下去的那一回合不算吃飯。
//
// 原版 `sub_1318` 的最後一行是 `if (ebx == 0) sub_2A50C()`,
// 而 `ebx` 只在活門那一條被設成 1。
func TestTrapdoorSkipsTheMeal(t *testing.T) {
	s := terrainScene(t, TileTrapdoor)
	s.Clock.Hour = 12 // 用餐時刻
	before := s.Inventory.Food
	s.terrainEffects()
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgTrapdoor) {
		t.Fatalf("沒印活門:%q", s.Messages)
	}
	if s.Inventory.Food != before {
		t.Errorf("掉下活門那一回合吃了飯:%d → %d", before, s.Inventory.Food)
	}
	// ★ 坐在魔毯上不會掉下去。
	s = terrainScene(t, TileTrapdoor)
	s.Transport = u5data.VehicleCarpet
	s.Messages = nil
	s.terrainEffects()
	if strings.Contains(strings.Join(s.Messages, "|"), MsgTrapdoor) {
		t.Errorf("坐魔毯竟然也掉下活門:%q", s.Messages)
	}
}

// TestStonegateTrapdoorKillsEveryone 是那條死路。
//
// ⚠ 原版在地點 29(STONEGATE)踩到活門,直接把全隊的血設 0、狀態設 'D'。
// 這條擋「順手加個存活判定」。
func TestStonegateTrapdoorKillsEveryone(t *testing.T) {
	// ⚠ 活門在**場景**裡,所以要真的進 STONEGATE ——
	// 只把 `Location` 設成 29 的話 `TileAt` 會去讀場景圖,而我們寫的是世界圖,
	// 失敗訊息會長得像「全隊沒死」,其實是地形根本沒被讀到。
	s := upkeepScene(t)
	if s.Scenes == nil {
		t.Skip("沒載場景圖")
	}
	if err := s.SetScene(StonegateLocation, 0, 15, 15); err != nil {
		t.Skipf("進不了 STONEGATE:%v", err)
	}
	s.Transport = u5data.VehicleWalk
	if !s.SetTileAt(s.X, s.Y, TileTrapdoor) {
		t.Skip("寫不進場景圖")
	}
	s.Messages = nil
	s.terrainEffects()
	for i := 0; i < s.PartySize; i++ {
		if s.Roster[i].Status != u5data.StatusDead {
			t.Errorf("第 %d 位在 STONEGATE 的活門下沒死 —— 原版是全隊當場死亡", i)
		}
		if s.Roster[i].HP != 0 {
			t.Errorf("第 %d 位的血是 %d,原版直接設 0", i, s.Roster[i].HP)
		}
	}
	// 別的地點不會死(大地圖上的活門)。
	s = terrainScene(t, TileTrapdoor)
	s.terrainEffects()
	allDead := true
	for i := 0; i < s.PartySize; i++ {
		if s.Roster[i].Status != u5data.StatusDead {
			allDead = false
		}
	}
	if allDead && s.PartySize > 0 {
		t.Error("在 STONEGATE 以外的活門下全隊都死了 —— 那一條只在地點 29 成立")
	}
}

// TestSleepersWakeOnTheirOwn:戰鬥外每回合 1/16 自己醒。
func TestSleepersWakeOnTheirOwn(t *testing.T) {
	s := terrainScene(t, 0x05) // 草地,沒有地形效果
	s.Roster[0].Status = u5data.StatusAsleep
	woke := false
	for i := 0; i < 400 && !woke; i++ {
		s.terrainEffects()
		if s.Roster[0].Status == u5data.StatusGood {
			woke = true
		}
	}
	if !woke {
		t.Errorf("擲 400 回都沒醒 —— 判準是 rand(0, %d) == %d", WakeUpRollMax, WakeUpHit)
	}
}

// envGamedata 是 `U5_GAMEDATA`(itest 才有;test 會是空字串 → 測試跳過)。
func envGamedata() string { return os.Getenv("U5_GAMEDATA") }

// TestBuyingADrinkMakesYouDrunk 是這一條的入口。
//
// ⚠ `tavern.go` 原本寫「原版只把『這趟喝了幾杯』加一,沒有其他效果」——
// 那是錯的:同一支 `sub_21108` 還有 `mov byte_3E169, 19h`。
func TestBuyingADrinkMakesYouDrunk(t *testing.T) {
	s := upkeepScene(t)
	if s.Drunk != 0 {
		t.Fatalf("一開始就醉了(%d)", s.Drunk)
	}
	s.GetDrunk()
	if s.Drunk != TavernDrunkTurns {
		t.Errorf("喝一杯醉 %d 次,預期 %d", s.Drunk, TavernDrunkTurns)
	}
	if TavernDrunkTurns != 25 {
		t.Errorf("醉的次數是 %d,原版是 0x19 = 25", TavernDrunkTurns)
	}
}

// TestDrunkStaggerHalfTheTimeAndCountsDown 釘住兩件事:1/2 機率,而且**踉蹺才扣**。
func TestDrunkStaggerHalfTheTimeAndCountsDown(t *testing.T) {
	// 沒醉就不會踉蹺。
	s := upkeepScene(t)
	for i := 0; i < 100; i++ {
		if _, ok := s.DrunkStagger(); ok {
			t.Fatal("沒醉卻踉蹺了")
		}
	}

	// 醉了:擲很多次,踉蹺的次數應該把計數扣到 0,而總按鍵數大約是兩倍。
	s = upkeepScene(t)
	s.GetDrunk()
	presses, staggers := 0, 0
	for s.Drunk > 0 && presses < 1000 {
		presses++
		if _, ok := s.DrunkStagger(); ok {
			staggers++
		}
	}
	if staggers != TavernDrunkTurns {
		t.Errorf("踉蹺了 %d 次才醒,預期 %d 次", staggers, TavernDrunkTurns)
	}
	// ★ 一半機率 → 按鍵數該在踉蹺數的 1..4 倍之間(放寬只為擋數量級錯誤)。
	if presses < staggers || presses > staggers*4 {
		t.Errorf("按了 %d 次才用完 %d 次踉蹺 —— 1/2 的機率不該這樣", presses, staggers)
	}
	// 醒了就不再踉蹺。
	for i := 0; i < 100; i++ {
		if _, ok := s.DrunkStagger(); ok {
			t.Fatal("醉意退了卻還在踉蹺")
		}
	}
}

// TestDrunkStaggerUsesAllFourDirections:四個方向都要出得來。
//
// 原版 `byte_4FC54` 的前四個位元組是 3, 4, 2, 1 —— 四個方向鍵碼。
func TestDrunkStaggerUsesAllFourDirections(t *testing.T) {
	s := upkeepScene(t)
	seen := map[Direction]bool{}
	for i := 0; i < 2000; i++ {
		s.Drunk = TavernDrunkTurns // 一直保持醉著
		if d, ok := s.DrunkStagger(); ok {
			seen[d] = true
		}
	}
	for _, d := range []Direction{North, South, East, West} {
		if !seen[d] {
			t.Errorf("方向 %v 一次都沒出現 —— 四個方向該等機率", d)
		}
	}
	if len(DrunkKeys) != 4 {
		t.Errorf("方向表有 %d 筆,原版是 4 筆", len(DrunkKeys))
	}
}

// TestSwampDiceDifferByPlace —— ★ 沼澤有兩顆骰子,而它們區分的是**地點**。
//
//	sub_10BDC(大地圖,由 sub_2D9D0 呼叫)   random(1, 30)
//	sub_1318 (場景,  由 sub_1A54  呼叫)   random(0, 29)
//
// 兩者條件完全相同,**只有範圍差一格**,而那個差異改變機率:
// 敏捷 29 的人只有在大地圖的沼澤裡才會中毒。
//
// ⚠⚠ 這個測試的前身叫 `TestSwampHasTwoSeparatePoisonRolls`,釘的是
// 「踏進沼澤那一步會被擲兩次」——**那個結論是錯的**(`docs/re/81` §3)。
// 兩支函式的呼叫者是**互斥的兩個模式主迴圈**,同一回合不會都跑。
// 骰子的性質沒變,錯的是「兩顆骰子代表什麼」。
func TestSwampDiceDifferByPlace(t *testing.T) {
	if SwampOverworldPoisonHi == SwampPoisonRollMax {
		t.Fatal("兩顆骰子的上限被寫成一樣了 —— 原版是 30(大地圖)與 29(場景)")
	}
	if SwampOverworldPoisonLo != 1 {
		t.Errorf("大地圖那顆的下限是 %d,原版是 1", SwampOverworldPoisonLo)
	}

	// 敏捷 29:場景那顆(0..29)永遠毒不到,大地圖那顆(1..30)有 1/30。
	s := upkeepScene(t)
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		s.Roster[i].Dex = 29
	}
	for i := 0; i < 600; i++ {
		s.Roster[0].Status = u5data.StatusGood
		s.poisonPartyBySwamp(0, SwampPoisonRollMax)
		if s.Roster[0].Status == u5data.StatusPoisoned {
			t.Fatal("敏捷 29 被場景那顆(0..29)毒到了 —— 上限 29 不可能大於 29")
		}
	}
	poisoned := 0
	for i := 0; i < 600; i++ {
		s.Roster[0].Status = u5data.StatusGood
		s.poisonPartyBySwamp(SwampOverworldPoisonLo, SwampOverworldPoisonHi)
		if s.Roster[0].Status == u5data.StatusPoisoned {
			poisoned++
		}
	}
	if poisoned == 0 {
		t.Error("敏捷 29 連大地圖那顆(1..30)都毒不到 —— 上限該是 30")
	}

	// 敏捷 30 兩邊都免疫。
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		s.Roster[i].Dex = 30
	}
	for i := 0; i < 300; i++ {
		s.Roster[0].Status = u5data.StatusGood
		s.poisonPartyBySwamp(0, SwampPoisonRollMax)
		s.poisonPartyBySwamp(SwampOverworldPoisonLo, SwampOverworldPoisonHi)
		if s.Roster[0].Status == u5data.StatusPoisoned {
			t.Fatal("敏捷 30 該對兩顆骰子都免疫")
		}
	}
}

// TestSwampPoisonSkipsTheDeadAndTheAlreadyPoisoned —— 兩支原版函式都有這兩個 cmp。
func TestSwampPoisonSkipsTheDeadAndTheAlreadyPoisoned(t *testing.T) {
	s := upkeepScene(t)
	if s.PartySize < 2 {
		t.Skip("需要兩個人以上")
	}
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		s.Roster[i].Dex = 0 // 敏捷 0 → 一定中毒(除非被跳過)
	}
	s.Roster[0].Status = u5data.StatusDead
	s.Roster[1].Status = u5data.StatusGood
	s.poisonPartyBySwamp(SwampOverworldPoisonLo, SwampOverworldPoisonHi)
	if s.Roster[0].Status != u5data.StatusDead {
		t.Errorf("死人被沼澤毒成了 %q", string(s.Roster[0].Status))
	}
	if s.Roster[1].Status != u5data.StatusPoisoned {
		t.Errorf("敏捷 0 的活人沒中毒,狀態是 %q", string(s.Roster[1].Status))
	}
}

// TestSwampOnlyPoisonsOnFoot —— 步行才中毒(`byte_3E08C == 0x1C`,**單一值**)。
//
// ★ 走的是 `overworldTurnEnd()` 而不是骰子本體,因為**閘門在呼叫端**:
// `sub_10BDC` 自己只有「逐個隊員擲一次」的迴圈,`tile == 4 && 載具 == 0x1C`
// 兩個條件都在 `sub_2D9D0` 裡。⚠ `docs/re/74` 曾寫「兩支的條件完全相同」——
// 那是把呼叫端的條件算進被呼叫的函式了;效果相同,但**檢查的位置不同**,
// 而位置決定了測試該打哪裡。
func TestSwampOnlyPoisonsOnFoot(t *testing.T) {
	s := upkeepScene(t)
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		s.Roster[i].Dex = 0
		s.Roster[i].Status = u5data.StatusGood
	}
	// ⚠ 大地圖那顆骰子只在大地圖擲(`sub_1318` 那顆是場景的)。
	// 所以這條測試需要**真的世界地圖** —— `upkeepScene` 只載場景,
	// 少了這一步 `SetTileAt` 會無聲失敗、測試變成永遠 skip 的空殼
	// (同 `fireScene` 的註解;那個坑本專案踩過)。
	s.Location, s.Floor = 0, 0
	w, err := u5data.LoadFlatMap(os.Getenv("U5_GAMEDATA") + "/UNDER.DAT")
	if err != nil {
		t.Skipf("載不到平面地圖:%v", err)
	}
	s.World, s.Under = w, w
	if s.X == 0 && s.Y == 0 {
		s.X, s.Y = 64, 64
	}
	if !s.SetTileAt(s.X, s.Y, TileSwamp) {
		t.Fatal("寫不進世界地圖 —— 這條測試沒有在驗任何東西")
	}
	s.Transport = 0x12 // 騎馬(載具 = 物件 + 2)
	s.overworldTurnEnd()
	if s.Roster[0].Status == u5data.StatusPoisoned {
		t.Error("騎著馬也被沼澤毒到了 —— 原版只在步行時判")
	}
	s.Transport = u5data.VehicleWalk
	s.overworldTurnEnd()
	if s.Roster[0].Status != u5data.StatusPoisoned {
		t.Error("步行走在沼澤上卻沒中毒")
	}
}
