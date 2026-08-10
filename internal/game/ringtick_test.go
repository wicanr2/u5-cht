package game

import (
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// TestRegenerationRingHealsOneInEight —— 再生戒指每回合 1/8 回 1 點。
//
// ⚠ 這條機制此前**完全不存在**:`sub_2BCC8` 被 Hex-Rays 反編譯成
// `return 0;`,而組語是 55 行(`CLAUDE.md §4.4` 第一種失效形態)。
func TestRegenerationRingHealsOneInEight(t *testing.T) {
	s := upkeepScene(t)
	s.Roster[0].MaxHP, s.Roster[0].HP = 200, 100
	s.Roster[0].Raw[u5data.CharRing] = u5data.ItemRingRegeneration
	healed := 0
	for i := 0; i < 800; i++ {
		before := s.Roster[0].HP
		s.regenerateParty()
		if s.Roster[0].HP > before {
			healed++
		}
		s.Roster[0].HP = 100 // 每次回到原點,才數得出命中次數
	}
	// 1/8 ≈ 100/800。放寬到 55..160 只擋「機率差一個數量級」。
	if healed < 55 || healed > 160 {
		t.Errorf("800 回合回血 %d 次,1/8 該落在 55..160", healed)
	}
}

// TestOnlyTheRegenerationRingHeals —— 反對照:另外兩只戒指不回血。
func TestOnlyTheRegenerationRingHeals(t *testing.T) {
	s := upkeepScene(t)
	for _, ring := range []byte{
		u5data.ItemNone, u5data.ItemRingInvisibility, u5data.ItemRingProtection,
	} {
		s.Roster[0].MaxHP, s.Roster[0].HP = 200, 100
		s.Roster[0].Raw[u5data.CharRing] = ring
		for i := 0; i < 400; i++ {
			s.regenerateParty()
		}
		if s.Roster[0].HP != 100 {
			t.Errorf("戒指 0x%02X 回了 %d 血 —— 只有再生戒指(0x%02X)會",
				ring, s.Roster[0].HP-100, u5data.ItemRingRegeneration)
		}
	}
}

// TestRegenerationStopsAtMaxHP —— 上限是**那個人的** MaxHP。
func TestRegenerationStopsAtMaxHP(t *testing.T) {
	s := upkeepScene(t)
	s.Roster[0].MaxHP, s.Roster[0].HP = 120, 120
	s.Roster[0].Raw[u5data.CharRing] = u5data.ItemRingRegeneration
	for i := 0; i < 400; i++ {
		s.regenerateParty()
	}
	if s.Roster[0].HP != 120 {
		t.Errorf("滿血還回到 %d,上限該是 MaxHP 120", s.Roster[0].HP)
	}
}

// TestRegenerationSkipsOnlyTheDead —— ★ 只跳過 'D';睡著與中毒的照樣回血。
//
// 寫成「只有 'G' 才回」會讓中毒的人永遠回不了血,而中毒在 U5 很常見 ——
// 那個 bug 會被當成「戒指好像沒用」。
func TestRegenerationSkipsOnlyTheDead(t *testing.T) {
	for _, tc := range []struct {
		status byte
		heals  bool
		name   string
	}{
		{u5data.StatusGood, true, "健康"},
		{u5data.StatusPoisoned, true, "★ 中毒照樣回血"},
		{u5data.StatusAsleep, true, "★ 睡著照樣回血"},
		{u5data.StatusCharmed, true, "★ 被惑照樣回血"},
		{u5data.StatusDead, false, "身亡 —— 唯一跳過的"},
	} {
		s := upkeepScene(t)
		s.Roster[0].Status = tc.status
		s.Roster[0].MaxHP, s.Roster[0].HP = 200, 100
		s.Roster[0].Raw[u5data.CharRing] = u5data.ItemRingRegeneration
		for i := 0; i < 400; i++ {
			s.regenerateParty()
		}
		if got := s.Roster[0].HP > 100; got != tc.heals {
			t.Errorf("%s:回血 = %v,預期 %v(HP %d)", tc.name, got, tc.heals, s.Roster[0].HP)
		}
	}
}

// TestTurnCounterCountsEveryTurnNotEveryMeal —— ★★ 語意更正的驗收。
//
// `byte_3E09B` 此前被記成「進餐計數器」。它的 +1 在 `sub_2A50C` 的尾段,
// 而「這個小時已經結算過」那條路**也會**走到那裡 ⇒ 每回合都 +1。
func TestTurnCounterCountsEveryTurnNotEveryMeal(t *testing.T) {
	s := upkeepScene(t)
	s.Clock.Hour = 3 // 不是用餐時刻
	for i := 0; i < 50; i++ {
		s.upkeep()
	}
	if got := s.TurnsSinceAlms(); got != 50 {
		t.Errorf("走 50 回合(都不是用餐時刻)計數器是 %d,預期 50 —— 它與吃飯無關", got)
	}
	// 上限 255,不繞回 0。
	for i := 0; i < 400; i++ {
		s.upkeep()
	}
	if got := s.TurnsSinceAlms(); got != TurnCounterMax {
		t.Errorf("計數器爆到 %d,預期夾在 %d", got, TurnCounterMax)
	}
}

// TestFieldsExpireOneInSixteen —— 力場每回合 1/16 消散,而且**不印訊息**。
func TestFieldsExpireOneInSixteen(t *testing.T) {
	s := upkeepScene(t)
	objs := s.currentObjects()
	if objs == nil {
		t.Skip("這個場景沒有物件層")
	}
	gone, tries := 0, 800
	for i := 0; i < tries; i++ {
		slot, ok := objs.Spawn(FieldObjectKind, s.X+2, s.Y+2, s.Floor)
		if !ok {
			t.Skip("物件槽滿了")
		}
		s.Messages = nil
		s.expireFields()
		if !objs.Objects[slot].Present() {
			gone++
		} else {
			objs.Remove(slot)
		}
		if len(s.Messages) != 0 {
			t.Fatalf("力場消散印了訊息 %q —— 原版那一段沒有任何輸出", s.Messages)
		}
	}
	// 1/16 ≈ 50/800。放寬到 20..100。
	if gone < 20 || gone > 100 {
		t.Errorf("800 次消散了 %d 次,1/16 該落在 20..100", gone)
	}
}

// TestOnlyFieldsExpire —— 反對照:別的物件不會自己消失。
//
// 少了這一條,「力場會消散」與「物件會隨機消失」用同一個觀察分不開,
// 而後者會讓玩家放在地上的東西不見。
func TestOnlyFieldsExpire(t *testing.T) {
	s := upkeepScene(t)
	objs := s.currentObjects()
	if objs == nil {
		t.Skip("這個場景沒有物件層")
	}
	for _, kind := range []byte{u5data.ObjLockedChest, 0x10, 0xE7, 0xEC} {
		slot, ok := objs.Spawn(kind, s.X+2, s.Y+2, s.Floor)
		if !ok {
			t.Skip("物件槽滿了")
		}
		for i := 0; i < 400; i++ {
			s.expireFields()
		}
		if !objs.Objects[slot].Present() {
			t.Errorf("種類 0x%02X 也消散了 —— 原版只比 `kind & 0xFC == 0xE8`", kind)
		}
		objs.Remove(slot)
	}
}

// almsScene 造一個「正在跟第 1 號 NPC 說話」的場景,並把那個 NPC 的
// 生物編號改成想測的值。
//
// ⚠ 直接改 `Creature` 而不是去找一個真的有乞丐的城鎮:那樣測試會依賴
// 「哪一座城有乞丐」,而那是資料而不是規則 —— 資料一變測試就跳過,
// 而跳過的測試看起來跟通過一樣。
func almsScene(t *testing.T, creature byte) *State {
	t.Helper()
	s := openCmdScene(t)
	if len(s.npcs) < 2 {
		t.Skip("這個場景沒有 NPC")
	}
	s.npcs[1].Creature = creature
	s.talkingTo = 1
	s.Karma, s.Inventory.Gold = 50, 100
	return s
}

// TestAlmsKarmaNeedsTheThrottle —— ★ 施捨的業報要等 100 回合。
//
// 這條機制此前**刻意留白**:閘門 `byte_3E09B >= 100` 的語意未定
// (`docs/re/79` 寫著「缺一段比多一段錯的好」)。語意定出來之後才接。
func TestAlmsKarmaNeedsTheThrottle(t *testing.T) {
	s := almsScene(t, BeggarKind)
	// 計數器不足 → 不給業報。
	s.turnsSinceAlms = AlmsThrottleTurns - 1
	s.demandGold(10)
	if s.Karma != 50 {
		t.Errorf("計數器只有 %d 就給了業報(業報 %d)", AlmsThrottleTurns-1, s.Karma)
	}
	if s.Inventory.Gold != 90 {
		t.Errorf("錢沒扣:%d", s.Inventory.Gold)
	}
	// 足夠 → +1,而且計數器歸零。
	s.turnsSinceAlms = AlmsThrottleTurns
	s.demandGold(10)
	if s.Karma != 50+AlmsKarma {
		t.Errorf("業報是 %d,預期 %d", s.Karma, 50+AlmsKarma)
	}
	if s.turnsSinceAlms != 0 {
		t.Errorf("給完業報計數器是 %d,預期歸零", s.turnsSinceAlms)
	}
}

// TestAlmsGivingEverythingAddsTwoMore —— ★ 把錢給光,再加 2(總共 +3)。
func TestAlmsGivingEverythingAddsTwoMore(t *testing.T) {
	s := almsScene(t, BeggarKind)
	s.turnsSinceAlms = AlmsThrottleTurns
	s.Inventory.Gold = 10
	s.demandGold(10) // 給光
	if s.Karma != 50+AlmsKarma+AlmsKarmaBroke {
		t.Errorf("給光錢之後業報是 %d,預期 %d", s.Karma, 50+AlmsKarma+AlmsKarmaBroke)
	}
	// 反對照:沒給光只加 1。
	s2 := almsScene(t, BeggarKind)
	s2.turnsSinceAlms = AlmsThrottleTurns
	s2.Inventory.Gold = 100
	s2.demandGold(10)
	if s2.Karma != 50+AlmsKarma {
		t.Errorf("還有剩錢卻加了 %d 業報,預期只加 %d", s2.Karma-50, AlmsKarma)
	}
}

// TestAlmsKarmaIsCappedAt99 —— 兩次都各自夾在 99,不是「+3 一次算完」。
func TestAlmsKarmaIsCappedAt99(t *testing.T) {
	s := almsScene(t, BeggarKind)
	s.turnsSinceAlms = AlmsThrottleTurns
	s.Karma = u5data.KarmaMax
	s.Inventory.Gold = 10
	s.demandGold(10)
	if s.Karma != u5data.KarmaMax {
		t.Errorf("業報 %d 超過上限 %d", s.Karma, u5data.KarmaMax)
	}
}

// TestAlmsOnlyForBeggars —— 反對照:給別人錢不加業報,也不清節流計數器。
func TestAlmsOnlyForBeggars(t *testing.T) {
	// 0x6C..0x6F 是乞丐的四個朝向(原版 `& 0xFC`),0x70 就不是了。
	// ⚠ 0x6B 是**範圍下緣外**的那一個(`0x6B & 0xFC == 0x68`),0x70 是上緣外。
	for _, creature := range []byte{0x70, 0x68, 0x00, 0x6B} {
		s := almsScene(t, creature)
		s.turnsSinceAlms = AlmsThrottleTurns
		s.demandGold(10)
		if s.Karma != 50 {
			t.Errorf("生物 0x%02X 也拿到了施捨業報:%d", creature, s.Karma)
		}
		if s.turnsSinceAlms != AlmsThrottleTurns {
			t.Errorf("生物 0x%02X 把節流計數器清掉了", creature)
		}
	}
	// 正對照:四個朝向都算乞丐。
	for _, creature := range []byte{0x6C, 0x6D, 0x6E, 0x6F} {
		s := almsScene(t, creature)
		s.turnsSinceAlms = AlmsThrottleTurns
		s.demandGold(10)
		if s.Karma == 50 {
			t.Errorf("生物 0x%02X 該算乞丐(`& 0xFC == 0x6C`)卻沒給業報", creature)
		}
	}
}
