package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// TestHiddenItemTableShape —— 113 筆表的形狀。
func TestHiddenItemTableShape(t *testing.T) {
	if n := len(u5data.HiddenItems); n != u5data.HiddenItemCount {
		t.Fatalf("表有 %d 筆,預期 %d", n, u5data.HiddenItemCount)
	}
	// 每一筆都有種類 —— 全 0 的空列表示抽表位址錯了。
	empty := 0
	for _, it := range u5data.HiddenItems {
		if it.Kind == 0 {
			empty++
		}
	}
	if empty > 0 {
		t.Errorf("有 %d 筆種類是 0 —— 抽表位址可能錯了", empty)
	}
	// 可重複拿的只有 13..15。
	for i := 0; i < u5data.HiddenItemCount; i++ {
		want := i >= 0x0D && i <= 0x0F
		if u5data.HiddenItemRepeatable(i) != want {
			t.Errorf("第 %d 筆可重複 = %v,預期 %v", i, !want, want)
		}
	}
}

// TestDungeonRoomNamesDifferFromTheLookTable —— ★ 兩張表是不同的。
//
// 六處差異都是刻意的,其中最有意思的是月石:搜出來時只說「一塊奇怪的石頭」。
func TestDungeonRoomNamesDifferFromTheLookTable(t *testing.T) {
	lt := loadLookOrSkipGame(t)
	diff := map[byte][2]string{}
	for kind := byte(1); kind <= 31; kind++ {
		a := u5data.DungeonRoomItemName(kind)
		b := lt.Object(int(kind))
		if b != "" && a != b {
			diff[kind] = [2]string{b, a}
		}
	}
	// 至少那六處要不一樣 —— 若兩張表變得相同,表示有人把它們合併了。
	for _, kind := range []byte{2, 7, 13, 25, 30, 31} {
		if _, ok := diff[kind]; !ok {
			t.Errorf("種類 %d 兩張表相同 —— 原版是不同的", kind)
		}
	}
	if got := u5data.DungeonRoomItemName(25); !strings.Contains(got, "strange rock") {
		t.Errorf("月石在這張表叫 %q,原版是 a strange rock", got)
	}
	// ★ 檀香木盒(14)落到 default。
	if got := u5data.DungeonRoomItemName(14); !strings.Contains(got, "nothing") {
		t.Errorf("檀香木盒印 %q,原版落到 default 的 nothing of note", got)
	}
	t.Logf("兩張表共 %d 處不同", len(diff))
}

// TestHerbSpotsNeedMidnight —— ★ 秘密採藥點只在午夜、每天每點一次。
func TestHerbSpotsNeedMidnight(t *testing.T) {
	s := worldState(t)
	spot := u5data.HerbSpots[0]
	x, y := int(spot.X), int(spot.Y)
	before := s.Inventory.Reagents[spot.Herb]

	// 白天:採不到。
	s.Clock.Hour, s.Clock.Day = 12, 5
	if s.gatherHerbs(x, y) {
		t.Error("中午竟然採到了藥草")
	}
	if s.Inventory.Reagents[spot.Herb] != before {
		t.Error("白天沒採到卻加了藥草")
	}

	// 午夜:採到。
	s.Clock.Hour = HerbGatherHour
	if !s.gatherHerbs(x, y) {
		t.Fatal("午夜採不到藥草")
	}
	got := s.Inventory.Reagents[spot.Herb] - before
	if got < HerbGatherMin || got > HerbGatherMax {
		t.Errorf("採到 %d 株,原版是 random(%d,%d)", got, HerbGatherMin, HerbGatherMax)
	}

	// ★ 同一天同一點採不到第二次。
	if s.gatherHerbs(x, y) {
		t.Error("同一天同一點採了兩次")
	}

	// ★ 隔天可以再採。
	s.Clock.Day++
	if !s.gatherHerbs(x, y) {
		t.Error("隔天採不到")
	}
}

// TestHerbSpotsAreTrackedPerSpot —— ★ 記帳是每點各一份。
//
// 同一個午夜跑三個點,三個點都該採得到 —— 全域一份的話第二點就會被擋掉。
func TestHerbSpotsAreTrackedPerSpot(t *testing.T) {
	s := worldState(t)
	s.Clock.Hour, s.Clock.Day = HerbGatherHour, 9
	for i := range u5data.HerbSpots {
		spot := u5data.HerbSpots[i]
		if !s.gatherHerbs(int(spot.X), int(spot.Y)) {
			t.Errorf("第 %d 個採藥點在同一個午夜採不到 —— 記帳可能是全域一份", i)
		}
	}
}

// TestHerbCapAtNinetyNine —— 上限 0x63。
func TestHerbCapAtNinetyNine(t *testing.T) {
	s := worldState(t)
	spot := u5data.HerbSpots[0]
	s.Inventory.Reagents[spot.Herb] = HerbCarryLimit
	s.Clock.Hour, s.Clock.Day = HerbGatherHour, 1
	s.gatherHerbs(int(spot.X), int(spot.Y))
	if got := s.Inventory.Reagents[spot.Herb]; got != HerbCarryLimit {
		t.Errorf("採完是 %d,上限應該夾在 %d", got, HerbCarryLimit)
	}
}

// TestBuriedMoonstoneCanBeDugUp —— 埋下去的月石搜得回來。
//
// ⚠ 四個欄位都要對上(X / Y / 樓層 / **地點碼**)—— 少比地點碼的話
// 「埋在城裡 (15,15)」與「埋在大地圖 (15,15)」會混在一起。
func TestBuriedMoonstoneCanBeDugUp(t *testing.T) {
	s := worldState(t)
	clearObjects(t, s)
	const mx, my = 40, 50
	s.Inventory.Moonstones[3] = u5data.Moonstone{X: mx, Y: my, Location: 0, Floor: 0}

	if !s.digUpMoonstone(mx, my) {
		t.Fatal("埋在這一格的月石挖不出來")
	}
	if got := strings.Join(s.Messages, "|"); !strings.Contains(got, "奇怪的石頭") {
		t.Errorf("挖出來卻印 %q", got)
	}
	// ★ 已經躺在那一格了 → 不再生第二顆。
	if s.digUpMoonstone(mx, my) {
		t.Error("同一格挖了兩次,會生出兩顆月石")
	}
	// 品質欄要帶著月石編號。
	objs := s.CurrentObjects()
	o, ok := objs.At(mx, my, s.Floor)
	if !ok {
		t.Fatal("月石物件不在那一格")
	}
	if o.Raw[u5data.ObjQuality] != 3 {
		t.Errorf("品質欄是 %d,預期月石編號 3", o.Raw[u5data.ObjQuality])
	}
}

// TestMoonstoneDigChecksTheLocationCode —— 地點碼不對就挖不到。
func TestMoonstoneDigChecksTheLocationCode(t *testing.T) {
	s := worldState(t)
	clearObjects(t, s)
	// 埋在地點 7,而玩家在大地圖(地點 0)。
	s.Inventory.Moonstones[0] = u5data.Moonstone{X: 20, Y: 20, Location: 7, Floor: 0}
	if s.digUpMoonstone(20, 20) {
		t.Error("在大地圖挖出了埋在城裡的月石")
	}
}

// TestMoonstoneDigScansBackwards —— ★ 倒著掃,同一格先挖出編號大的。
func TestMoonstoneDigScansBackwards(t *testing.T) {
	s := worldState(t)
	clearObjects(t, s)
	const mx, my = 31, 32
	s.Inventory.Moonstones[2] = u5data.Moonstone{X: mx, Y: my, Location: 0, Floor: 0}
	s.Inventory.Moonstones[6] = u5data.Moonstone{X: mx, Y: my, Location: 0, Floor: 0}
	if !s.digUpMoonstone(mx, my) {
		t.Fatal("挖不出來")
	}
	o, _ := s.CurrentObjects().At(mx, my, s.Floor)
	if o.Raw[u5data.ObjQuality] != 6 {
		t.Errorf("先挖出第 %d 顆,原版倒著掃該是第 6 顆", o.Raw[u5data.ObjQuality])
	}
}

// TestSpareKeysOnlyAppearWhenYouHaveNone —— ★ 防卡死:鑰匙用完才會再長。
func TestSpareKeysOnlyAppearWhenYouHaveNone(t *testing.T) {
	s := worldState(t)
	s.Inventory.Keys = 3
	if s.hiddenItemAvailable(u5data.HiddenItemSpareKeys) {
		t.Error("身上有鑰匙時那一串備用鑰匙還是拿得到")
	}
	s.Inventory.Keys = 0
	if !s.hiddenItemAvailable(u5data.HiddenItemSpareKeys) {
		t.Error("鑰匙用完了卻拿不到備用的 —— 那會卡死")
	}
}

// TestHiddenItemTakenOnce —— 撿過就不再出現(可重複的三筆除外)。
func TestHiddenItemTakenOnce(t *testing.T) {
	s := worldState(t)
	const once = 20 // 不在 13..15 的一筆
	if !s.hiddenItemAvailable(once) {
		t.Fatal("一開始就拿不到")
	}
	s.markHiddenItemTaken(once)
	if s.hiddenItemAvailable(once) {
		t.Error("撿過了還拿得到")
	}
	// ★ 可重複的那三筆:記了也照樣拿得到(因為呼叫端不會記)。
	for i := 0x0D; i <= 0x0F; i++ {
		if !u5data.HiddenItemRepeatable(i) {
			t.Errorf("第 %d 筆該是可重複的", i)
		}
	}
}

// TestFixedContentsRunBeforeTheRandomRoll —— 三層在隨機那一層之前。
//
// 判準:站在一個有固定物品的格子上搜,拿到的是**那一筆**而不是隨機的金幣。
func TestFixedContentsRunBeforeTheRandomRoll(t *testing.T) {
	s := worldState(t)
	clearObjects(t, s)
	// 找一筆在大地圖(地點 0、樓層 0)的固定物品。
	idx := -1
	for i := 0; i < u5data.HiddenItemCount; i++ {
		it := u5data.HiddenItems[i]
		if it.Loc == 0 && it.Floor == 0 && !u5data.HiddenItemRepeatable(i) {
			idx = i
			break
		}
	}
	if idx < 0 {
		t.Skip("表裡沒有大地圖上的固定物品")
	}
	it := u5data.HiddenItems[idx]
	if !s.findHiddenItem(int(it.X), int(it.Y)) {
		t.Fatalf("第 %d 筆(%d,%d)搜不到", idx, it.X, it.Y)
	}
	o, ok := s.CurrentObjects().At(int(it.X), int(it.Y), 0)
	if !ok {
		t.Fatal("物件沒被放到被搜的那一格")
	}
	if o.Raw[u5data.ObjKind] != it.Kind {
		t.Errorf("放的是種類 %d,預期 %d", o.Raw[u5data.ObjKind], it.Kind)
	}
	// 再搜一次拿不到(已記帳)。
	if s.findHiddenItem(int(it.X), int(it.Y)) {
		t.Error("同一筆固定物品拿了兩次")
	}
}
