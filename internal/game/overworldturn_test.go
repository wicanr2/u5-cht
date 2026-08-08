package game

import (
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// overworldScene 是一個站在**真的世界地圖**上的隊伍。
//
// ⚠ 需要真地圖:`SetTileAt` 在沒有世界地圖時**無聲失敗**,而那會讓整條測試
// 變成永遠通過的空殼(本專案踩過,見 `upkeep_test.go` 的註解)。所以載不到
// 就 `Skip`,而寫不進去就 `Fatal` —— 兩者要分開。
func overworldScene(t *testing.T) *State {
	t.Helper()
	s := upkeepScene(t)
	s.Location, s.Floor = 0, 0
	w, err := u5data.LoadFlatMap(os.Getenv("U5_GAMEDATA") + "/UNDER.DAT")
	if err != nil {
		t.Skipf("載不到平面地圖:%v", err)
	}
	s.World, s.Under = w, w
	s.X, s.Y = 64, 64
	s.Transport = u5data.VehicleWalk
	if !s.SetTileAt(s.X, s.Y, tileGrass) {
		t.Fatal("寫不進世界地圖 —— 這條測試沒有在驗任何東西")
	}
	s.Messages = nil
	return s
}

// TestOverworldTurnIsTwoMinutesAndSceneTurnIsOne —— `docs/re/81` §2。
//
//	大地圖  sub_2D9D0 → sub_29304(2)
//	場景    sub_1A54  → sub_29304(1)
//	地牢    sub_5378  → sub_29304(1)
//
// ⚠ 引擎原本三個模式共用 `MinutesPerTurn = 1` ⇒ **大地圖的時鐘走一半速度**,
// 連帶 NPC 排程、月相、日夜全部偏掉。這不是誤差是系統性的一半。
func TestOverworldTurnIsTwoMinutesAndSceneTurnIsOne(t *testing.T) {
	s := overworldScene(t)
	if got := s.minutesPerTurn(); got != OverworldMinutesPerTurn {
		t.Errorf("大地圖一回合 %d 分,原版是 %d", got, OverworldMinutesPerTurn)
	}
	// 進場景 → 回到 1 分鐘。
	s.Location = 1
	if got := s.minutesPerTurn(); got != MinutesPerTurn {
		t.Errorf("場景一回合 %d 分,原版是 %d", got, MinutesPerTurn)
	}
	s.Location = 0
	// 地牢同樣是 1(`sub_5378`)。
	s.Dungeon = &DungeonState{}
	if got := s.minutesPerTurn(); got != MinutesPerTurn {
		t.Errorf("地牢一回合 %d 分,原版是 %d", got, MinutesPerTurn)
	}
	s.Dungeon = nil

	// 真的走一步:時鐘該前進 2 分(草地是 0 級地形,沒有額外代價)。
	before := s.Clock
	s.tick()
	adv := (s.Clock.Hour-before.Hour)*MinutesPerHour + s.Clock.Minute - before.Minute
	if adv != OverworldMinutesPerTurn {
		t.Errorf("大地圖 tick 推了 %d 分,預期 %d", adv, OverworldMinutesPerTurn)
	}
}

// TestEveryOverworldTurnRunsAWorldTurn —— ★★ 這是十條裡玩家最感覺得到的一條。
//
// `sub_2D9D0` 的 `loc_2DD2F: call sub_2E24` 是**無條件**的:每個用掉回合的
// 大地圖動作都跑一個完整世界回合。
//
// ⚠⚠ 引擎原本只從 `payTerrainCost` 呼叫 `extraWorldTurn()`,而那只在地形
// **1 級或 2 級**時才有 ⇒ 草地(0 級)上走路連世界回合都不跑。
//
// ⚠ 這條測的是**世界回合有沒有被呼叫**(`WorldTurns` 計數),不是「有沒有冒出怪」
// —— 因為 `extraWorldTurn` 的**本體還沒實作**(遭遇門檻 / 生怪 / 怪物移動 /
// 漂流 / 清場是五支函式的另一個子系統,見那支函式的說明與 `WORKLIST §5.1` 第 11 條)。
// 把測試寫成「該冒出怪」會紅得很誠實但擋住這一批的落地;寫成「該被呼叫」
// 是**這一步真正做完的事**,而缺的那一塊有自己的 ⬜ 守著。
func TestEveryOverworldTurnRunsAWorldTurn(t *testing.T) {
	s := overworldScene(t)
	before := s.WorldTurns
	const steps = 10
	for i := 0; i < steps; i++ {
		s.tick()
	}
	if got := s.WorldTurns - before; got != steps {
		t.Errorf("草地上走 %d 回合只跑了 %d 個世界回合,原版是每回合一個", steps, got)
	}
}

// TestAnTymStopsTheWorldAndRelTymHalvesIt —— `sub_2E24` 的前兩道閘門。
func TestAnTymStopsTheWorldAndRelTymHalvesIt(t *testing.T) {
	s := overworldScene(t)

	// An Tym('T'):整個跳過,一個世界回合都不跑。
	s.CombatMode = CombatModeTimeStop
	before := s.WorldTurns
	for i := 0; i < 20; i++ {
		s.tick()
	}
	if s.WorldTurns != before {
		t.Errorf("An Tym 期間跑了 %d 個世界回合,原版是 0", s.WorldTurns-before)
	}

	// Rel Tym('Q'):隔次跳過 ⇒ 20 回合該剛好 10 個。
	s.CombatMode = CombatModeSlow
	s.gates = worldTurnGates{}
	before = s.WorldTurns
	for i := 0; i < 20; i++ {
		s.tick()
	}
	if got := s.WorldTurns - before; got != 10 {
		t.Errorf("Rel Tym 期間 20 回合跑了 %d 個世界回合,原版隔次跳過該是 10", got)
	}
}

// TestRidingHalvesTheWorldTurns —— 第三道閘門,而它回答了 `docs/re/38` §5 的問號。
//
// `byte_4FDD7` 是**持久切換位元**不是機率,而全檔只有這一個 ⇒ `sub_2E24` 的
// 四個呼叫點共用它。所以騎馬 / 坐魔毯時**所有**世界回合都隔次跳過,
// 等於「怪物只有一半的行動機會」—— 那是騎馬真正的好處,不只是走得快。
func TestRidingHalvesTheWorldTurns(t *testing.T) {
	s := overworldScene(t)
	for _, tc := range []struct {
		name      string
		transport byte
	}{
		{"騎馬", 0x12},
		{"魔毯", 0x14},
	} {
		s.Transport = tc.transport
		s.gates = worldTurnGates{}
		before := s.WorldTurns
		for i := 0; i < 20; i++ {
			s.tick()
		}
		if got := s.WorldTurns - before; got != 10 {
			t.Errorf("%s 走 20 回合跑了 %d 個世界回合,隔次跳過該是 10", tc.name, got)
		}
	}

	// ★ 步行不受影響 —— 原版比的是 0x12 / 0x14 兩對值。
	s.Transport = u5data.VehicleWalk
	s.gates = worldTurnGates{}
	before := s.WorldTurns
	for i := 0; i < 20; i++ {
		s.tick()
	}
	if got := s.WorldTurns - before; got != 20 {
		t.Errorf("步行走 20 回合只跑了 %d 個世界回合,該是 20", got)
	}
}

// TestGrassCostsNoExtraTurnButStillRunsOne —— 基本回合與額外回合是兩件事。
//
// `docs/re/38`:1 級地形 = 1 個**額外**世界回合、2 級 = 2 個。基本的那一個
// 在 `sub_2D9D0` 尾段,不歸地形代價管。所以山裡走一步怪走三步(1 + 2),
// 草地上走一步怪走一步(1 + 0)—— **不是 0**。
func TestGrassCostsNoExtraTurnButStillRunsOne(t *testing.T) {
	if TerrainCost(tileGrass) != terrainCostNone {
		t.Fatalf("草地被算成 %d 級 —— 原版把它從 4..15 裡挑出來當 0 級", TerrainCost(tileGrass))
	}
	if TerrainCost(TileSwamp) == terrainCostNone {
		t.Error("沼澤該是 1 級")
	}
}

// TestEarthquakeOnlyHappensInTheUnderworld —— `sub_2D998`。
//
//	if (byte_3E0A5 == 0) return                ; 地表不震
//	if (random(0, 0FFh) != 69h) return         ; 1/256,恰好等於
//
// ⚠ 判準是**等於** 0x69,不是「小於門檻」。寫成 `< 0x69` 會變成 105/256,
// 那是四十倍的差距 —— 而兩種寫法都「偶爾會地震」,玩起來分不出來。
func TestEarthquakeOnlyHappensInTheUnderworld(t *testing.T) {
	s := overworldScene(t)

	// 地表:擲多少次都不該震。
	s.Floor = 0
	for i := 0; i < 3000; i++ {
		s.Messages = nil
		s.underworldEarthquake()
		if strings.Contains(strings.Join(s.Messages, "|"), MsgEarthquake) {
			t.Fatal("地表也地震了 —— 原版第一行就 `cmp byte_3E0A5, 0; jz`")
		}
	}

	// 幽冥界:1/256,3000 次該中好幾次。
	s.Floor = UnderworldFloor
	quakes, damaged := 0, false
	for i := 0; i < 3000; i++ {
		s.Messages = nil
		for j := 0; j < s.PartySize && j < len(s.Roster); j++ {
			s.Roster[j].HP = 200
		}
		s.underworldEarthquake()
		if strings.Contains(strings.Join(s.Messages, "|"), MsgEarthquake) {
			quakes++
			if s.Roster[0].HP < 200 {
				damaged = true
			}
		}
	}
	if quakes == 0 {
		t.Error("幽冥界擲 3000 次一次都沒地震 —— 1/256 不該這樣")
	}
	if quakes > 300 {
		t.Errorf("3000 次震了 %d 次(約 1/%d)—— 原版是**恰好等於** 0x69 的 1/256,不是小於", quakes, 3000/max(quakes, 1))
	}
	if !damaged {
		t.Error("地震沒有造成傷害 —— 原版 `sub_2A4D0` 是全隊各 random(1, 8)")
	}
}

// TestSacredQuestGatePushesYouBackWithoutAQuest —— `sub_2D9D0` 的 `loc_2DC5F`。
//
// 通關條件是「**此刻手上有任何一項未完成的聖壇試煉**」(`byte_3E0DC != 0`),
// 不是「領過某個試煉」。所以八德全部領完獎之後這道關卡會**重新關上** ——
// 那是原版行為,不要「順手」改成單向開啟。
//
// ★ 被拒絕時原版是 `inc byte_3E0A7`(往南推一格),不是「擋住不讓走」:
// 玩家已經站上去了,才被推回來。
func TestSacredQuestGatePushesYouBackWithoutAQuest(t *testing.T) {
	s := overworldScene(t)
	s.X, s.Y = SacredGateX, SacredGateY

	// 沒有進行中的試煉 → 兩句拒絕 + 往南一格。
	s.ShrineQuestActive = 0
	s.Messages = nil
	s.sacredQuestGate()
	joined := strings.Join(s.Messages, "|")
	if !strings.Contains(joined, MsgNotOnSacredQuest) || !strings.Contains(joined, MsgPassageDenied) {
		t.Errorf("拒絕時少了訊息:%q", s.Messages)
	}
	if s.Y != SacredGateY+1 {
		t.Errorf("被拒絕後 Y = %d,原版 `inc byte_3E0A7` 該是 %d", s.Y, SacredGateY+1)
	}

	// 有任一項進行中的試煉 → 通行,而且**不動座標**。
	s.X, s.Y = SacredGateX, SacredGateY
	s.ShrineQuestActive = 1 << 3
	s.Messages = nil
	s.sacredQuestGate()
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgPassSeeker) {
		t.Errorf("有試煉卻沒放行:%q", s.Messages)
	}
	if s.Y != SacredGateY {
		t.Errorf("放行卻把人移動了,Y = %d", s.Y)
	}

	// 別的格子不受影響。
	s.X, s.Y = SacredGateX+1, SacredGateY
	s.ShrineQuestActive = 0
	s.Messages = nil
	s.sacredQuestGate()
	if len(s.Messages) != 0 {
		t.Errorf("旁邊那一格也觸發了關卡:%q", s.Messages)
	}

	// ★ 幽冥界不觸發(原版 `cmp byte_3E0A5, 0`)。
	s.X, s.Y = SacredGateX, SacredGateY
	s.Floor = UnderworldFloor
	s.Messages = nil
	s.sacredQuestGate()
	if len(s.Messages) != 0 {
		t.Errorf("幽冥界的同一座標也觸發了:%q", s.Messages)
	}
}

// TestSleepersDoNotWakeOnTheOverworld —— ★ 這條是「原版沒有 X」型的斷言,
// 而它有**全檔掃描**佐證(`CLAUDE.md §4.5` 要求):
//
//	把名冊狀態寫成 'G' 的位置:全檔 22 處
//	`random(0, 15) == 15` 之後寫 'G':**只有 1 處**,在 `sub_1318`(場景)
//
// ⇒ 睡著的隊員在大地圖上不會自己醒。引擎原本所有模式都跑 `terrainEffects()`,
// 所以醒得比原版容易 —— 而那讓睡眠系咒語與橙藥水在大地圖上失去重量。
func TestSleepersDoNotWakeOnTheOverworld(t *testing.T) {
	s := overworldScene(t)
	s.Roster[0].Status = u5data.StatusAsleep
	for i := 0; i < 500; i++ {
		s.tick()
		if s.InCombat() {
			// 遭遇打斷了測試,重來一輪就好(世界回合是無條件跑的)。
			s.Combat = nil
		}
		if s.Roster[0].Status != u5data.StatusAsleep {
			t.Fatalf("大地圖走 %d 回合就自己醒了 —— 那條 1/16 只在 `sub_1318`(場景)裡", i+1)
		}
	}
}

// TestSceneSleepersStillWake —— 上一條的正對照。
//
// 少了這一條,「大地圖不會醒」有可能只是因為我把整段醒來邏輯弄壞了。
func TestSceneSleepersStillWake(t *testing.T) {
	s := upkeepScene(t)
	s.Location = 1
	s.Roster[0].Status = u5data.StatusAsleep
	woke := false
	for i := 0; i < 500; i++ {
		s.terrainEffects()
		if s.Roster[0].Status == u5data.StatusGood {
			woke = true
			break
		}
	}
	if !woke {
		t.Error("場景裡走 500 回合都沒醒 —— 1/16 該早就中了")
	}
}

// TestLavaBurnsOnTheOverworldToo —— `sub_10BC4`(大地圖)與 `sub_1318`(場景)
// 是兩支不同的函式做同一件事。⚠ 場景那條還包含壁爐(0xBC),大地圖只有熔岩。
func TestLavaBurnsOnTheOverworldToo(t *testing.T) {
	s := overworldScene(t)
	if !s.SetTileAt(s.X, s.Y, TileLava) {
		t.Fatal("寫不進世界地圖")
	}
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		s.Roster[i].HP = 200
	}
	s.Messages = nil
	s.overworldTurnEnd()
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgBurning) {
		t.Errorf("大地圖的熔岩沒有燒人:%q", s.Messages)
	}
	if s.Roster[0].HP >= 200 {
		t.Error("熔岩沒扣血 —— `sub_10BC4` 接的是 `sub_2A4D0`")
	}
}

// TestPassRunsTheTurnEnd —— 空白鍵一樣要結算。
//
// 原版 case 32 印完字就落回 `sub_2D9D0` / `sub_1A54` 的尾段。引擎原本自己
// `AdvanceTime` + `extraWorldTurn()` ⇒ **站著不動不會餓、中毒也不會痛**。
// 而「按空白鍵可以無代價等待」會讓斷糧與中毒兩個機制整個失效。
func TestPassRunsTheTurnEnd(t *testing.T) {
	s := overworldScene(t)
	s.Roster[0].Status = u5data.StatusPoisoned
	s.Roster[0].HP = 200
	s.Clock.Hour = 3 // 不是用餐時刻,排除扣糧的干擾
	s.Pass()
	if s.Roster[0].HP >= 200 {
		t.Error("按空白鍵中毒沒扣血 —— Pass 沒有走每回合收尾")
	}
}

// TestTerrainCostIsPaidBeforeTheTurnEnd —— `docs/re/81` §4 的順序。
//
// 原版 `sub_2D174`(移動 + 地形代價)在 `sub_2D9D0` 的 line 79899,
// 每回合收尾在 line 80013 ⇒ **代價在前**。`docs/re/38` §4 曾寫反。
//
// 怎麼驗:2 級地形走一步,時鐘該前進 2(基本)+ 4(代價)= 6 分。
// 順序錯不會改總和,但會改「NPC 排程看到的是哪一個小時」——
// 所以這條測的是**總和**,順序本身靠 `moveInWorld` 的程式碼順序與註解守。
func TestTerrainCostIsPaidBeforeTheTurnEnd(t *testing.T) {
	s := overworldScene(t)
	// ⚠ 不能用山(12):它**擋步行** ⇒ `Move` 直接回「去路受阻」,時鐘一分都不動,
	// 測試會紅得像順序錯了。第一版就踩了這個 —— **紅燈的成因在前置條件而不是結論**,
	// 這個形狀本專案已經踩過五次。丘陵(11)同樣是 2 級但走得上去。
	const hills = 11
	if TerrainCost(hills) != terrainCostVery {
		t.Fatalf("tile %d 不是 2 級", hills)
	}
	if u5data.TileBlocksWalking(hills) {
		t.Fatalf("tile %d 也擋步行 —— 換一個 2 級且可通行的", hills)
	}
	if !s.SetTileAt(s.X, s.Y+1, hills) {
		t.Fatal("寫不進世界地圖")
	}
	before := s.Clock
	s.Move(South)
	adv := (s.Clock.Hour-before.Hour)*MinutesPerHour + s.Clock.Minute - before.Minute
	if adv < OverworldMinutesPerTurn+terrainCostMinutes[terrainCostVery] {
		t.Errorf("走上山推了 %d 分,預期至少 %d(2 基本 + 4 代價)",
			adv, OverworldMinutesPerTurn+terrainCostMinutes[terrainCostVery])
	}
}
