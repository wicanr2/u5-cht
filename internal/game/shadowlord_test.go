package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// TestFlamesSitInTheShadowlordKeeps:三團聖火的地點要與召喚地點是同一組。
//
// ★ 兩張表是**獨立來源**:`ShadowlordKeeps` 來自 `sub_17A14` 的
// `cmp al, 1Eh/1Fh/20h`,`Flames` 的地點欄來自 `sub_1A38C` 用的並列表
//(`aNoNoticeableEf + 0x22`)。對不上就代表有一邊抄錯。
func TestFlamesSitInTheShadowlordKeeps(t *testing.T) {
	for i, f := range u5data.Flames {
		if f.Location != u5data.ShadowlordKeeps[i] {
			t.Errorf("第 %d 團聖火在地點 %d,召喚表說是 %d",
				i, f.Location, u5data.ShadowlordKeeps[i])
		}
		if f.Location < 1 || f.Location > len(u5data.Locations) {
			t.Errorf("第 %d 團聖火的地點 %d 不在地點表範圍內", i, f.Location)
		}
		loc := u5data.Locations[f.Location-1]
		if f.Floor < loc.FloorMin || f.Floor > loc.FloorMax {
			t.Errorf("第 %d 團聖火在第 %d 層,而 %s 只有 %d..%d 層",
				i, f.Floor, loc.Name, loc.FloorMin, loc.FloorMax)
		}
	}
}

// TestRoamingNeverStacksAndAvoidsThePlayer:午夜分派的兩個排除條件。
//
// 少了「不重複」三位會疊在同一座城,少了「避開玩家」會憑空出現在腳下 ——
// 兩種都會讓主線的節奏走樣。這裡跑很多輪,因為它是抽籤:單跑一次證明不了事。
func TestRoamingNeverStacksAndAvoidsThePlayer(t *testing.T) {
	s := shrineState(t)
	if s == nil {
		return
	}
	s.ShadowlordAt = [u5data.ShadowlordCount]byte{1, 2, 3}
	for round := 0; round < 200; round++ {
		s.Location = round%u5data.ShadowlordCityMax + 1
		s.roamShadowlords()
		seen := map[byte]bool{}
		for i, v := range s.ShadowlordAt {
			if v < u5data.ShadowlordCityMin || v > u5data.ShadowlordCityMax {
				t.Fatalf("第 %d 輪:第 %d 位跑到地點 %d(該在 1..8)", round, i, v)
			}
			if int(v) == s.Location {
				t.Fatalf("第 %d 輪:第 %d 位跑到玩家所在的地點 %d", round, i, v)
			}
			if seen[v] {
				t.Fatalf("第 %d 輪:兩位暗影君主疊在地點 %d", round, v)
			}
			seen[v] = true
		}
	}
}

// TestRoamingSkipsTheDead:已被消滅的不會復活去別的城。
func TestRoamingSkipsTheDead(t *testing.T) {
	s := shrineState(t)
	if s == nil {
		return
	}
	s.ShadowlordAt = [u5data.ShadowlordCount]byte{u5data.ShadowlordGone, 2, 0}
	for i := 0; i < 50; i++ {
		s.roamShadowlords()
		if s.ShadowlordAt[0] != u5data.ShadowlordGone {
			t.Fatalf("已消滅的暗影君主被重新分派到 %d", s.ShadowlordAt[0])
		}
	}
	// 第 2 位的初值是 0(不在城裡)—— 0 < 0x80,所以它**會**被分派。
	if s.ShadowlordAt[2] == 0 {
		t.Error("初值 0 的那一位一直沒被分派出去")
	}
}

// TestMidnightTriggersRoaming:跨過午夜才重新分派,而且跨一次算一次。
//
// ⚠ 判斷要看**日期變了沒**,不是「小時剛好是 0」——
// 休息或進出聖壇石室一次推進超過一小時,看小時會整個跳過去。
func TestMidnightTriggersRoaming(t *testing.T) {
	s := shrineState(t)
	if s == nil {
		return
	}
	s.ShadowlordAt = [u5data.ShadowlordCount]byte{1, 2, 3}
	s.Clock.Hour, s.Clock.Minute = 12, 0
	before := s.ShadowlordAt
	s.AdvanceTime(60)
	if s.ShadowlordAt != before {
		t.Error("白天走一小時就重新分派了")
	}
	// 一口氣跨過午夜(而且不是整點落在 0 點)。
	s.Clock.Hour, s.Clock.Minute = 23, 30
	day := s.Clock.Day
	s.AdvanceTime(90) // → 01:00,隔天
	if s.Clock.Day == day {
		t.Fatal("時鐘沒有跨日")
	}
	if s.ShadowlordAt == before {
		t.Error("跨過午夜卻沒有重新分派 —— 大概是拿小時 == 0 在判")
	}
}

// TestGemShardNeedsFlamePlaceAndShadowlord:碎片的三個條件缺一不可。
func TestGemShardNeedsFlamePlaceAndShadowlord(t *testing.T) {
	s := shrineState(t)
	if s == nil {
		return
	}
	const which = 1 // ASTAROTH / 愛之火 / 共感修道院
	f := u5data.Flames[which]
	s.Shards[which] = true

	// 沒到聖火前:毫無效果。
	s.Messages = nil
	if s.UseGemShard(which) {
		t.Error("不在聖火前也消滅得掉")
	}
	if !strings.Contains(s.log(), MsgNoEffect) {
		t.Errorf("不在聖火前沒有印「%s」:\n%s", MsgNoEffect, s.log())
	}

	// 站到聖火那一格,但沒有暗影君主 —— 原版**什麼都不說**。
	if err := s.SetScene(f.Location, f.Floor, f.X, f.Y); err != nil {
		t.Skipf("進不了 %d 號地點第 %d 層:%v", f.Location, f.Floor, err)
	}
	s.Messages = nil
	if s.UseGemShard(which) {
		t.Error("沒有暗影君主在場也消滅得掉")
	}
	if strings.Contains(s.log(), MsgNoEffect) {
		t.Errorf("位置對但沒目標時原版是沉默的,卻印了「%s」:\n%s", MsgNoEffect, s.log())
	}

	// 召出來 —— 牠在玩家上方**兩**格,碎片要的是**一**格。
	s.Yell()
	s.Input = u5data.Shadowlords[which]
	s.SubmitYell()
	if !s.shadowlordPresent() {
		t.Fatalf("召不出來:\n%s", s.log())
	}
	if s.UseGemShard(which) {
		t.Error("暗影君主在兩格外就被消滅了 —— 原版比的是正北一格")
	}

	// 把牠挪到正北一格。
	n := -1
	for _, v := range s.VisibleNPCs() {
		if v.NPC.Creature == u5data.TileShadowlord {
			n = v.Index
		}
	}
	if n < 0 {
		t.Fatal("找不到召出來的暗影君主")
	}
	s.rtNPCs[n].Y = s.Y - 1
	s.Messages = nil
	if !s.UseGemShard(which) {
		t.Fatalf("三個條件都滿足了卻沒消滅:\n%s", s.log())
	}
	if s.ShadowlordAt[which] != u5data.ShadowlordGone {
		t.Errorf("消滅之後 ShadowlordAt[%d] = %d", which, s.ShadowlordAt[which])
	}
	if s.Shards[which] {
		t.Error("碎片用掉了卻還在背包裡")
	}
	if s.DoomFlags()&u5data.ShadowlordDoomBit[which] == 0 {
		t.Error("末日位元沒有設起來")
	}
	if s.shadowlordPresent() {
		t.Error("消滅之後暗影君主還在場上")
	}
}

// TestWrongShardDoesNotKill:碎片與現身的那一位要對得上。
//
// 原版比的是 `byte_3E0DB == i`。少了這一條,任何一塊碎片都能打掉任何一位,
// 而「哪塊碎片配哪團火」正是這條支線的全部內容。
func TestWrongShardDoesNotKill(t *testing.T) {
	s := shrineState(t)
	if s == nil {
		return
	}
	f := u5data.Flames[1]
	s.Shards[0], s.Shards[1] = true, true
	if err := s.SetScene(f.Location, f.Floor, f.X, f.Y); err != nil {
		t.Skipf("進不了:%v", err)
	}
	s.Yell()
	s.Input = u5data.Shadowlords[1] // 召 ASTAROTH
	s.SubmitYell()
	for _, v := range s.VisibleNPCs() {
		if v.NPC.Creature == u5data.TileShadowlord {
			s.rtNPCs[v.Index].Y = s.Y - 1
		}
	}
	// 拿第 0 塊碎片(虛偽)—— 位置雖然不是它的火,所以先驗位置那一關。
	if s.UseGemShard(0) {
		t.Error("在別團火前用第 0 塊碎片也成功了")
	}
	// 把玩家搬到第 0 塊碎片的火(位置條件過),但現身的仍是第 1 位。
	f0 := u5data.Flames[0]
	s.Location, s.Floor, s.X, s.Y = f0.Location, f0.Floor, f0.X, f0.Y
	if s.UseGemShard(0) {
		t.Error("現身的是第 1 位,第 0 塊碎片卻消滅得掉")
	}
	if s.ShadowlordAt[1] == u5data.ShadowlordGone {
		t.Error("拿錯碎片卻把第 1 位打掉了")
	}
}

// TestShardsSurviveSaveRoundTrip:碎片存得回去也讀得回來。
func TestShardsSurviveSaveRoundTrip(t *testing.T) {
	s := shrineState(t)
	if s == nil || s.BaseSave == nil {
		return
	}
	s.Shards = [u5data.ShadowlordCount]bool{true, false, true}
	gold := s.Inventory.Gold
	sv, err := s.ExportSave(s.BaseSave)
	if err != nil {
		t.Fatal(err)
	}
	blob, err := sv.Encode()
	if err != nil {
		t.Fatal(err)
	}
	back, err := u5data.ParseSave(blob)
	if err != nil {
		t.Fatal(err)
	}
	want := [u5data.ShardCount]byte{1, 0, 1, sv.Shards[3]}
	if back.Shards != want {
		t.Errorf("碎片讀回來是 %v,預期 %v", back.Shards, want)
	}
	if back.Inventory.Gold != gold {
		// 0x0210 就在背包那一段裡 —— 位移錯的話最先被踩到的是金幣與裝備。
		t.Errorf("金幣從 %d 變成 %d —— 碎片的位移大概寫到別人身上了", gold, back.Inventory.Gold)
	}
}

// TestHauntedTownSpawnsTheShadowlord:進到被盤據的城,牠會現身。
//
// 三件事一起發生:印「空氣中瀰漫著……」、暗影君主出現在 (15, 表[地點])、
// 居民依是哪一位而變。少任何一件,玩家就不知道這座城出事了。
func TestHauntedTownSpawnsTheShadowlord(t *testing.T) {
	s := shrineState(t)
	if s == nil {
		return
	}
	const town = 2 // 不列顛城
	s.ShadowlordAt = [u5data.ShadowlordCount]byte{0, town, 0}
	if err := s.SetScene(town, 0, 15, 15); err != nil {
		t.Skipf("進不了不列顛城:%v", err)
	}
	if !s.shadowlordPresent() {
		t.Fatalf("被盤據的城裡沒有暗影君主:\n%s", s.log())
	}
	if !strings.Contains(s.log(), u5data.ShadowlordAir[1]) {
		t.Errorf("沒有印「空氣中瀰漫著憎恨」:\n%s", s.log())
	}
	// 落點要照表。
	wantY := int(u5data.ShadowlordTownY[town])
	found := false
	for _, v := range s.VisibleNPCs() {
		if v.NPC.Creature == u5data.TileShadowlord {
			found = true
			if v.X != u5data.ShadowlordTownX || v.Y != wantY {
				t.Errorf("暗影君主在 (%d,%d),表上是 (%d,%d)",
					v.X, v.Y, u5data.ShadowlordTownX, wantY)
			}
		}
	}
	if !found {
		t.Error("看不到牠")
	}
}

// TestHatredMakesTownsfolkHostileAndFearMakesThemFlee:兩位的效果不一樣。
//
// ⚠ 三位一視同仁的話,玩家就分辨不出這座城被誰佔了 ——
// 而「憎恨的城裡人人撲上來、怯懦的城裡人人跑光」正是這條支線的表現方式。
// 虛偽(0)在這裡什麼都不做。
func TestHatredMakesTownsfolkHostileAndFearMakesThemFlee(t *testing.T) {
	count := func(which int, ai byte) int {
		s := shrineState(t)
		if s == nil {
			return -1
		}
		s.ShadowlordAt = [u5data.ShadowlordCount]byte{0xFF, 0xFF, 0xFF}
		s.ShadowlordAt[which] = 2
		if err := s.SetScene(2, 0, 15, 15); err != nil {
			t.Skipf("進不了不列顛城:%v", err)
		}
		n := 0
		for i := range s.npcs {
			if s.npcs[i].Present() && s.npcs[i].Schedule.AI[0] == ai {
				n++
			}
		}
		return n
	}
	hostile := count(1, u5data.NPCAIHostile)
	fleeing := count(2, u5data.NPCAIFleeing)
	if hostile < 0 || fleeing < 0 {
		return
	}
	if hostile == 0 && fleeing == 0 {
		t.Skip("這座城此刻沒有符合條件的居民(原版的 4 號槽 bug 會整城跳過)")
	}
	if hostile == 0 {
		t.Error("憎恨佔城卻沒有人變敵對")
	}
	if fleeing == 0 {
		t.Error("怯懦佔城卻沒有人逃跑")
	}
}

// TestStonegateAnnouncesEverySurvivor:石門會為每一位還活著的各印一次。
func TestStonegateAnnouncesEverySurvivor(t *testing.T) {
	s := shrineState(t)
	if s == nil {
		return
	}
	s.ShadowlordAt = [u5data.ShadowlordCount]byte{1, u5data.ShadowlordGone, 3}
	if err := s.SetScene(u5data.StonegateLocation, 0, 15, 15); err != nil {
		t.Skipf("進不了石門:%v", err)
	}
	log := s.log()
	if !strings.Contains(log, u5data.ShadowlordAir[0]) {
		t.Error("沒有印虛偽的氣息")
	}
	if strings.Contains(log, u5data.ShadowlordAir[1]) {
		t.Error("已經被消滅的那一位還是印了")
	}
	if !strings.Contains(log, u5data.ShadowlordAir[2]) {
		t.Error("沒有印怯懦的氣息")
	}
}

// TestShadowlordTownYCoversExactlyTheEightCities:落點表只有八座城非零。
//
// ★ 這是 `byte_3EF1B` 的自我驗證:遊走的範圍是 `random(1,8)`,
// 而落點表恰好也只有 1..8 有值。兩份獨立資料對上 —— 表讀錯的話對不上。
func TestShadowlordTownYCoversExactlyTheEightCities(t *testing.T) {
	for loc, y := range u5data.ShadowlordTownY {
		inRange := loc >= u5data.ShadowlordCityMin && loc <= u5data.ShadowlordCityMax
		if inRange && y == 0 {
			t.Errorf("地點 %d 是八德城市,落點卻是 0", loc)
		}
		if !inRange && y != 0 {
			t.Errorf("地點 %d 不在遊走範圍內,落點卻是 %d", loc, y)
		}
	}
}
