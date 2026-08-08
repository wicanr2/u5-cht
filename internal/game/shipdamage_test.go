package game

import (
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// shipScene 造一艘停在海上的船。
//
// ⚠ 要載一張真的世界地圖,否則 `TileAt` 全回 0、`SetTileAt` 無效 ——
// 而失敗訊息會長得像「撞擊沒判對」,把人指向錯的地方(同 `fire_test.go` 的註記)。
func shipScene(t *testing.T, hull int) *State {
	t.Helper()
	s := newCreateState(t)
	s.MaxMessages = 32
	s.Location = 0
	s.Transport = u5data.VehicleShip // 0x24:收帆的船,朝北
	s.ShipHull, s.ShipSkiffs = hull, 0
	s.Inventory.Carpets = 0
	if dir := os.Getenv("U5_GAMEDATA"); dir != "" {
		if w, err := u5data.LoadFlatMap(dir + "/UNDER.DAT"); err == nil {
			s.World, s.Under = w, w
		}
	}
	s.Messages = nil
	return s
}

// TestOnlyBigShipsTakeHullDamage:小艇、魔毯、馬、步行進來都是空轉。
//
// 這一條擋的是「把 Rough seas 的傷害套到小艇上」—— 原版那六個觸發點
// 有兩個會在小艇 / 魔毯上呼叫 `sub_22F0`,靠 `& 0xF8 == 0x20` 擋掉。
func TestOnlyBigShipsTakeHullDamage(t *testing.T) {
	for _, tr := range []byte{
		u5data.VehicleWalk, u5data.TileHorse | 2, u5data.VehicleCarpet,
		u5data.VehicleSkiff, u5data.VehicleSkiff | 3,
	} {
		s := shipScene(t, 50)
		s.Transport = tr
		if s.DamageShip() {
			t.Errorf("載具 0x%02X 竟然沉了", tr)
		}
		if s.ShipHull != 50 {
			t.Errorf("載具 0x%02X 的耐久被動了(%d)", tr, s.ShipHull)
		}
	}
	// 揚帆與收帆兩組都要吃傷害。
	for _, tr := range []byte{
		u5data.VehicleSailing, u5data.VehicleSailing | 3,
		u5data.VehicleShip, u5data.VehicleShip | 3,
	} {
		if !u5data.ShipTakesDamage(tr) {
			t.Errorf("載具 0x%02X 應該吃船身傷害", tr)
		}
	}
}

// TestHullLossStaysWithinTheDieRange:耐久夠高時只掉 1..30。
func TestHullLossStaysWithinTheDieRange(t *testing.T) {
	for i := 0; i < 200; i++ {
		s := shipScene(t, 99)
		if s.DamageShip() {
			t.Fatalf("耐久 99 不該一次沉(第 %d 次)", i)
		}
		lost := 99 - s.ShipHull
		if lost < 1 || lost > u5data.ShipDamageMax {
			t.Fatalf("掉了 %d 點,超出 1..%d", lost, u5data.ShipDamageMax)
		}
	}
}

// TestAbandonShipLadderIsSkiffThenCarpetThenDrown 釘住三層階梯的順序。
func TestAbandonShipLadderIsSkiffThenCarpetThenDrown(t *testing.T) {
	// (1) 有小艇 → 換小艇,朝向保留,而且**不扣小艇數**。
	s := shipScene(t, 1)
	s.Transport = u5data.VehicleShip | 3 // 朝西
	s.ShipSkiffs, s.Inventory.Carpets = 2, 1
	if !s.DamageShip() {
		t.Fatal("耐久 1 應該一定沉")
	}
	if s.Transport != u5data.VehicleSkiff|3 {
		t.Errorf("載具變成 0x%02X,預期小艇且朝向保留(0x%02X)",
			s.Transport, u5data.VehicleSkiff|3)
	}
	if s.Inventory.Carpets != 1 {
		t.Error("有小艇時不該動到魔毯")
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgAbandonShip) {
		t.Errorf("沒印「棄船」:%q", s.Messages)
	}

	// (2) 沒小艇但有魔毯 → 扣一張魔毯,換成魔毯的兩個朝向之一。
	s = shipScene(t, 1)
	s.Inventory.Carpets = 2
	if !s.DamageShip() {
		t.Fatal("耐久 1 應該一定沉")
	}
	if s.Inventory.Carpets != 1 {
		t.Errorf("魔毯剩 %d 張,預期扣掉一張", s.Inventory.Carpets)
	}
	if s.Transport != u5data.VehicleCarpet && s.Transport != u5data.VehicleCarpet+1 {
		t.Errorf("載具變成 0x%02X,預期魔毯 0x%02X / 0x%02X",
			s.Transport, u5data.VehicleCarpet, u5data.VehicleCarpet+1)
	}

	// (3) 兩者都沒有 → 溺水,而且全隊掉血。
	s = shipScene(t, 1)
	before := make([]uint16, s.PartySize)
	for i := range before {
		s.Roster[i].Status = u5data.StatusGood
		s.Roster[i].HP = 200
		before[i] = s.Roster[i].HP
	}
	if !s.DamageShip() {
		t.Fatal("耐久 1 應該一定沉")
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgDrowning) {
		t.Errorf("沒印「溺水」:%q", s.Messages)
	}
	if strings.Contains(strings.Join(s.Messages, "|"), MsgAbandonShip) {
		t.Error("既沒小艇也沒魔毯卻印了「棄船」")
	}
	hurt := 0
	for i := range before {
		if s.Roster[i].HP < before[i] {
			hurt++
			if lost := int(before[i] - s.Roster[i].HP); lost > u5data.DrownDamageMax {
				t.Errorf("第 %d 位掉了 %d 血,超出 1..%d", i, lost, u5data.DrownDamageMax)
			}
		}
	}
	if hurt != s.PartySize {
		t.Errorf("只有 %d 位掉血,預期全隊 %d 位", hurt, s.PartySize)
	}
}

// TestDrowningSkipsDamageWhenNobodyCanAct:全滅的隊伍不再吃傷害。
//
// 原版的閘門是 `sub_2B67C() != -1`,而 −1 正是「沒人能行動也沒人睡著」。
func TestDrowningSkipsDamageWhenNobodyCanAct(t *testing.T) {
	s := shipScene(t, 1)
	for i := 0; i < s.PartySize; i++ {
		s.Roster[i].Status = u5data.StatusDead
		s.Roster[i].HP = 0
	}
	if s.anyoneCanAct() {
		t.Fatal("全隊身亡卻回報還有人能行動")
	}
	s.DamageShip()
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgDrowning) {
		t.Errorf("還是要印「溺水」:%q", s.Messages)
	}
}

// TestAnyoneCanActMatchesTheThreeWayReturn 逐狀態對回 `sub_2B67C`。
func TestAnyoneCanActMatchesTheThreeWayReturn(t *testing.T) {
	cases := []struct {
		status []byte
		want   bool
		why    string
	}{
		{[]byte{u5data.StatusGood}, true, "'G' 能行動 → 回 0"},
		{[]byte{u5data.StatusPoisoned}, true, "★ 'P' 也能行動 —— 兩者走同一條分支"},
		{[]byte{u5data.StatusAsleep}, true, "睡著 → 回 1,不是 −1"},
		{[]byte{u5data.StatusDead}, false, "只有死人 → 回 −1"},
		{[]byte{u5data.StatusCharmed}, false, "被惑也跳出迴圈 → −1"},
		{[]byte{u5data.StatusDead, u5data.StatusGood}, false,
			"★ 遇到 'D' 就跳出 —— 後面那個好人掃不到"},
		{[]byte{u5data.StatusAsleep, u5data.StatusGood}, true, "睡著會繼續掃,掃到好人"},
	}
	for _, c := range cases {
		s := shipScene(t, 50)
		s.PartySize = len(c.status)
		for i, st := range c.status {
			s.Roster[i].Status = st
		}
		if got := s.anyoneCanAct(); got != c.want {
			t.Errorf("%v → %v,預期 %v(%s)", c.status, got, c.want, c.why)
		}
	}
}

// TestTurningAShipCostsTheStep:船轉向要花掉這一步。
func TestTurningAShipCostsTheStep(t *testing.T) {
	s := shipScene(t, 99)
	s.Transport = u5data.VehicleShip | byte(North)
	if !s.turnShipInstead(East) {
		t.Error("轉向該吃掉這一步")
	}
	if s.Transport != u5data.VehicleShip|byte(East) {
		t.Errorf("轉完是 0x%02X,預期朝東 0x%02X", s.Transport, u5data.VehicleShip|byte(East))
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgHeading) {
		t.Errorf("沒印「轉向」:%q", s.Messages)
	}
	// 已經朝東,收帆的船就直接走。
	if s.turnShipInstead(East) {
		t.Error("已經朝東的收帆船不該再被吃掉一步")
	}
}

// TestHullWeakThresholdIsFiftyNotTen 是這一條最容易寫錯的地方。
//
// 上船時的警告門檻是 **10**(`sub_16F08`,"Danger! Ship badly damaged!"),
// 航行轉向時是 **50**(`sub_2CCFC`,"Hull weak!")—— 兩個不同的警告。
func TestHullWeakThresholdIsFiftyNotTen(t *testing.T) {
	if u5data.ShipHullWeak != 50 {
		t.Fatalf("門檻是 %d,預期 50", u5data.ShipHullWeak)
	}
	if u5data.ShipHullWarning == u5data.ShipHullWeak {
		t.Error("兩個警告的門檻不該相同")
	}
	for _, c := range []struct {
		hull int
		warn bool
	}{{49, true}, {50, false}, {99, false}, {1, true}} {
		s := shipScene(t, c.hull)
		s.Transport = u5data.VehicleShip | byte(North)
		s.turnShipInstead(South)
		got := strings.Contains(strings.Join(s.Messages, "|"), MsgHullWeak)
		if got != c.warn {
			t.Errorf("耐久 %d 警告 = %v,預期 %v", c.hull, got, c.warn)
		}
	}
}

// TestSailsUpWithNoWindCannotMove 是修掉的那個錯。
//
// `CanSail` 原本寫「無風 → 照走」,依據是 `sub_2D38` 在查表前就 `jz` 掉無風。
// 那句話只對一半:不查表是對的,但揚著帆就是動不了 —— 判斷在 `sub_2CCFC`。
func TestSailsUpWithNoWindCannotMove(t *testing.T) {
	// 揚帆 + 無風 + 已經朝那個方向 → 動不了。
	s := shipScene(t, 99)
	s.Transport = u5data.VehicleSailing | byte(North)
	s.Wind = u5data.WindCalm
	if !s.turnShipInstead(North) {
		t.Error("揚著帆又無風該動不了")
	}
	// 有風就走。
	s.Wind = u5data.WindNorth
	if s.turnShipInstead(North) {
		t.Error("揚著帆又有風該照走")
	}
	// ★ 收帆的船不受風影響。
	s.Transport = u5data.VehicleShip | byte(North)
	s.Wind = u5data.WindCalm
	if s.turnShipInstead(North) {
		t.Error("收帆的船無風也該照走")
	}
}

// TestRoughSeasHitsSmallCraftOnlyAndDealsNoDamage 釘住那個「刻意的空轉」。
func TestRoughSeasHitsSmallCraftOnlyAndDealsNoDamage(t *testing.T) {
	for _, tr := range []byte{u5data.VehicleSkiff, u5data.VehicleSkiff | 2,
		u5data.VehicleCarpet, u5data.VehicleCarpet | 1} {
		s := shipScene(t, 99)
		s.Transport = tr
		s.RoughSeas()
		if !strings.Contains(strings.Join(s.Messages, "|"), MsgRoughSeas) {
			t.Errorf("載具 0x%02X:沒印「風浪險惡」:%q", tr, s.Messages)
		}
		if s.ShipHull != 99 {
			t.Errorf("載具 0x%02X:耐久被動了 —— 原版對小艇 / 魔毯只有訊息", tr)
		}
		if s.Transport != tr {
			t.Errorf("載具 0x%02X 被換成 0x%02X", tr, s.Transport)
		}
	}
	// 大船與步行不會遇到 Rough seas。
	for _, tr := range []byte{u5data.VehicleShip, u5data.VehicleSailing, u5data.VehicleWalk} {
		s := shipScene(t, 99)
		s.Transport = tr
		s.RoughSeas()
		if strings.Contains(strings.Join(s.Messages, "|"), MsgRoughSeas) {
			t.Errorf("載具 0x%02X 不該遇到風浪:%q", tr, s.Messages)
		}
	}
}

// TestWholePartyDamageIsOneToEight:更正掉「random(1,20) 是估計值」那條。
func TestWholePartyDamageIsOneToEight(t *testing.T) {
	if u5data.DrownDamageMax != 8 {
		t.Fatalf("上限是 %d,預期 8(原版 `sub_28E14(1, 8)`)", u5data.DrownDamageMax)
	}
	for i := 0; i < 200; i++ {
		s := shipScene(t, 99)
		for j := 0; j < s.PartySize; j++ {
			s.Roster[j].Status = u5data.StatusGood
			s.Roster[j].HP = 200
		}
		s.damageWholeParty()
		for j := 0; j < s.PartySize; j++ {
			lost := 200 - int(s.Roster[j].HP)
			if lost < 1 || lost > u5data.DrownDamageMax {
				t.Fatalf("第 %d 位掉了 %d 血,超出 1..%d", j, lost, u5data.DrownDamageMax)
			}
		}
	}
}

// TestCollisionOutcomesByTerrain 逐地形釘住撞擊的四種結果。
func TestCollisionOutcomesByTerrain(t *testing.T) {
	cases := []struct {
		tile    byte
		sailing bool
		want    []string
		absent  []string
		why     string
	}{
		{u5data.TileDock, true, []string{MsgDocked},
			[]string{MsgCollision, MsgBreakingUp, MsgShipSunk}, "碼頭是靠岸,不印撞擊"},
		{u5data.TileShallowWater, true, []string{MsgBreakingUp},
			[]string{MsgCollision}, "★ 淺灘印裂開,不印 COLLISION"},
		{0x0A, true, []string{MsgCollision},
			[]string{MsgBreakingUp, MsgDocked}, "其他地形印 COLLISION"},
		{u5data.TileCactus, false, []string{MsgBlocked, MsgOuch},
			[]string{MsgCollision}, "沒揚帆撞仙人掌"},
		{0x0A, false, []string{MsgBlocked},
			[]string{MsgOuch, MsgCollision}, "沒揚帆撞普通障礙"},
	}
	for _, c := range cases {
		s := shipScene(t, 99)
		if c.sailing {
			s.Transport = u5data.VehicleSailing | byte(North)
		} else {
			s.Transport = u5data.VehicleWalk
		}
		// 把要撞的那一格換成指定地形。
		s.X, s.Y = 10, 10
		if !s.SetTileAt(11, 10, c.tile) {
			t.Skip("寫不進世界地圖")
		}
		s.Messages = nil
		s.blockedMove(11, 10)
		joined := strings.Join(s.Messages, "|")
		for _, w := range c.want {
			if !strings.Contains(joined, w) {
				t.Errorf("地形 0x%02X(%s):少了 %q,實際 %q", c.tile, c.why, w, s.Messages)
			}
		}
		for _, w := range c.absent {
			if strings.Contains(joined, w) {
				t.Errorf("地形 0x%02X(%s):不該有 %q,實際 %q", c.tile, c.why, w, s.Messages)
			}
		}
	}
}

// TestDockingLowersTheSailsAndKeepsFacing:靠岸是 `add byte_3E08C, 4`。
func TestDockingLowersTheSailsAndKeepsFacing(t *testing.T) {
	for _, d := range []Direction{North, East, South, West} {
		s := shipScene(t, 99)
		s.Transport = u5data.VehicleSailing | byte(d)
		s.X, s.Y = 10, 10
		if !s.SetTileAt(11, 10, u5data.TileDock) {
			t.Skip("寫不進世界地圖")
		}
		s.blockedMove(11, 10)
		if s.Transport != u5data.VehicleShip|byte(d) {
			t.Errorf("朝向 %v:靠岸後載具是 0x%02X,預期 0x%02X",
				d, s.Transport, u5data.VehicleShip|byte(d))
		}
	}
}

// TestCollisionDamagesTheHullAndLowersSails:撞完要掉耐久並自動收帆。
func TestCollisionDamagesTheHullAndLowersSails(t *testing.T) {
	s := shipScene(t, 99)
	s.Transport = u5data.VehicleSailing | byte(East)
	s.X, s.Y = 10, 10
	if !s.SetTileAt(11, 10, 0x0A) {
		t.Skip("寫不進世界地圖")
	}
	s.blockedMove(11, 10)
	if s.ShipHull >= 99 {
		t.Errorf("撞完耐久還是 %d —— 該掉血", s.ShipHull)
	}
	if u5data.VehicleKind(s.Transport) != u5data.VehicleShip {
		t.Errorf("撞完載具是 0x%02X —— 該自動收帆", s.Transport)
	}
	if s.Transport&0x03 != byte(East) {
		t.Errorf("撞完朝向變了:0x%02X", s.Transport)
	}
}

// ─── 做完了卻沒人叫的三支(`docs/re/80`)──────────────────────────────

// TestRoughSeasActuallyFiresOnDeepWater —— ★★ `RoughSeas` 原本沒有呼叫者。
//
// 上一輪把 `RoughSeas` 寫好了、也有測試,但**沒有任何地方叫它** ⇒
// 遊戲裡的風浪永遠不會發生。這條測的是**入口**,不是效果。
func TestRoughSeasActuallyFiresOnDeepWater(t *testing.T) {
	s := shipScene(t, 100)
	if !s.SetTileAt(s.X, s.Y, u5data.RoughSeasTile) {
		t.Skip("寫不進地圖")
	}
	// 小艇踩在深水上 → 該有風浪。
	s.Transport = u5data.VehicleSkiff
	s.Messages = nil
	s.RoughSeasHere()
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgRoughSeas) {
		t.Errorf("小艇在深水上沒有風浪:%q", s.Messages)
	}
	// 步行不算。
	s.Transport = u5data.VehicleWalk
	s.Messages = nil
	s.RoughSeasHere()
	if strings.Contains(strings.Join(s.Messages, "|"), MsgRoughSeas) {
		t.Error("步行竟然也有風浪")
	}
	// 淺灘不算(原版比的是 `esi == 1`,單一個 tile 值)。
	s.Transport = u5data.VehicleSkiff
	s.SetTileAt(s.X, s.Y, u5data.TileShallowWater)
	s.Messages = nil
	s.RoughSeasHere()
	if strings.Contains(strings.Join(s.Messages, "|"), MsgRoughSeas) {
		t.Error("淺灘竟然也有風浪 —— 原版只認 tile 1")
	}
}

// TestWindIsHiddenInDungeonsAndAtArarat —— `sub_2A984` 的三道閘門。
func TestWindIsHiddenInDungeonsAndAtArarat(t *testing.T) {
	s := shipScene(t, 100)
	s.Location, s.Floor = 0, 0
	if !s.WindShown() {
		t.Error("大地圖上該畫風向")
	}
	s.Location = SightAlwaysDarkLocation // ★ 亞拉臘號殘骸
	if s.WindShown() {
		t.Errorf("地點 %d(亞拉臘號)不該畫風向", SightAlwaysDarkLocation)
	}
	s.Location = 0
	s.Floor = -1
	if s.WindShown() {
		t.Error("地下世界不該走這一條(原版是另一個分支)")
	}
	s.Floor = 0
	s.Dungeon = &DungeonState{Index: 0, Location: u5data.DungeonLocationBase}
	if s.WindShown() {
		t.Error("地牢裡不該畫風向")
	}
}

// TestShipHullOnlyShowsAboardAShip —— `Ship:` 與 `G:` 共用同一格。
func TestShipHullOnlyShowsAboardAShip(t *testing.T) {
	s := shipScene(t, 100)
	s.Transport = u5data.VehicleWalk
	if _, ok := s.ShipHullShown(); ok {
		t.Error("步行時不該畫船身耐久")
	}
	// ★ 遮罩是 0xF8 → 揚帆組與收帆組都算。
	for _, v := range []byte{u5data.VehicleSailing, u5data.VehicleShip} {
		s.Transport = v
		if _, ok := s.ShipHullShown(); !ok {
			t.Errorf("載具 %#x 該畫船身耐久(遮罩是 0xF8,兩組都算)", v)
		}
	}
	// 小艇不算(0x28 & 0xF8 = 0x28)。
	s.Transport = u5data.VehicleSkiff
	if _, ok := s.ShipHullShown(); ok {
		t.Error("小艇不該畫船身耐久")
	}
}
