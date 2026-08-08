package game

import (
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// TestSpawnWeightsSumTo256 —— ★★ 四張表的權重都恰好加到 256。
//
// 這是「怪物清單與權重清單配對正確」的**硬佐證**:四張表長度各不相同
// (12 / 8 / 5 / 3),而四張都剛好 256 不可能是巧合。任何一張抄錯一個位元組
// 都會在這裡被抓到。
func TestSpawnWeightsSumTo256(t *testing.T) {
	for _, tc := range []struct {
		name    string
		kinds   []byte
		weights []byte
	}{
		{"陸地表", u5data.SpawnLandSurface, u5data.SpawnLandSurfaceWeights},
		{"陸幽冥", u5data.SpawnLandUnderworld, u5data.SpawnLandUnderworldWeights},
		{"水地表", u5data.SpawnWaterSurface, u5data.SpawnWaterSurfaceWeights},
		{"水幽冥", u5data.SpawnWaterUnderworld, u5data.SpawnWaterUnderworldWeights},
	} {
		if len(tc.kinds) != len(tc.weights) {
			t.Errorf("%s:%d 種怪對 %d 個權重", tc.name, len(tc.kinds), len(tc.weights))
		}
		sum := 0
		for _, w := range tc.weights {
			sum += int(w)
		}
		if sum != 256 {
			t.Errorf("%s 的權重加起來是 %d,原版是 256", tc.name, sum)
		}
	}
}

// TestSpawnPickWeightedMatchesTheOriginalLoop —— `sub_2008` 的逐格對照。
//
// 陸地・地表:權重 60 / 50 / 40 / … ⇒ r=0..59 → 0、r=60..109 → 1、r=110..149 → 2。
func TestSpawnPickWeightedMatchesTheOriginalLoop(t *testing.T) {
	w := u5data.SpawnLandSurfaceWeights
	for _, tc := range []struct{ roll, want int }{
		{0, 0}, {59, 0}, {60, 1}, {109, 1}, {110, 2}, {149, 2}, {255, 11},
	} {
		if got := u5data.SpawnPickWeighted(w, tc.roll); got != tc.want {
			t.Errorf("擲 %d 選到索引 %d,預期 %d", tc.roll, got, tc.want)
		}
	}
	// ★ 權重 0 的那一項永遠選不到(幽冥界水域的索引 0)。
	for r := 0; r < 256; r++ {
		if u5data.SpawnPickWeighted(u5data.SpawnWaterUnderworldWeights, r) == 0 {
			t.Fatalf("擲 %d 選到了權重 0 的索引 0", r)
		}
	}
}

// TestEncounterThresholdByTerrainAndHour —— `sub_1F98` 的逐格對照。
func TestEncounterThresholdByTerrainAndHour(t *testing.T) {
	const noon, dawn = 12, 3
	for _, tc := range []struct {
		name  string
		tile  byte
		hour  int
		under bool
		want  int
	}{
		{"草地・白天", 5, noon, false, 1},
		{"草地・凌晨", 5, dawn, false, 4},
		{"沼澤・白天", 4, noon, false, 2},
		{"森林・白天", 9, noon, false, 2},
		{"山・白天", 0x0C, noon, false, 2},
		{"沼澤・凌晨", 4, dawn, false, 5},
		{"0x20 那族・白天", 0x20, noon, false, 0},
		{"0x26 那族・白天", 0x26, noon, false, 0},
		{"0x20 那族・凌晨", 0x20, dawn, false, 3},
		{"幽冥界・白天", 5, noon, true, 3},
		{"幽冥界・凌晨", 5, dawn, true, 3}, // ★ 不加夜間加成
	} {
		if got := u5data.EncounterThreshold(tc.tile, tc.hour, tc.under); got != tc.want {
			t.Errorf("%s:門檻 %d,原版 %d", tc.name, got, tc.want)
		}
	}
}

// TestGrasslandNeverSpawnsByDay —— ★★ 本輪最重要的一條。
//
// 判定是 `random(1, 30) < 門檻`,而草地的門檻是 **1** ⇒ `1 < 1` 不成立
// ⇒ **白天在草地上走路,原版永遠不會有隨機遭遇。**
//
// ⚠ 寫成 `<=` 會讓平原上偶爾遇敵,而那「感覺也很合理」—— 沒有這條測試,
// 一個比原版兇的世界不會被發現。
func TestGrasslandNeverSpawnsByDay(t *testing.T) {
	s := overworldScene(t)
	s.Clock.Hour = 12
	for i := 0; i < 3000; i++ {
		if s.Roll(u5data.EncounterRollLo, u5data.EncounterRollHi) <
			u5data.EncounterThreshold(tileGrass, s.Clock.Hour, false) {
			t.Fatalf("第 %d 次擲骰讓草地生怪了 —— 門檻 1 配 random(1,30) 該永遠不成立", i)
		}
	}
}

// TestForestSpawnsByDayAndGrassSpawnsAtNight —— 門檻 2 與夜間 +3 都真的會生。
//
// 這是上一條的正對照:少了它,「草地不生」有可能只是因為我把生怪整段弄壞了。
func TestForestSpawnsByDayAndGrassSpawnsAtNight(t *testing.T) {
	s := overworldScene(t)
	hit := func(tile byte, hour int) bool {
		th := u5data.EncounterThreshold(tile, hour, false)
		for i := 0; i < 3000; i++ {
			if s.Roll(u5data.EncounterRollLo, u5data.EncounterRollHi) < th {
				return true
			}
		}
		return false
	}
	if !hit(9, 12) {
		t.Error("森林在白天擲 3000 次都沒生怪 —— 門檻 2 該有 1/30")
	}
	if !hit(tileGrass, 3) {
		t.Error("草地在凌晨擲 3000 次都沒生怪 —— 門檻 4 該有 3/30")
	}
}

// TestSpawnTerrainSplitsWaterLandAndNone —— `sub_203C` 的三族分類。
func TestSpawnTerrainSplitsWaterLandAndNone(t *testing.T) {
	for _, tc := range []struct {
		tile byte
		want u5data.SpawnTerrainKind
		why  string
	}{
		{0, u5data.SpawnTerrainWater, "tile < 4"},
		{1, u5data.SpawnTerrainWater, "深水"},
		{3, u5data.SpawnTerrainWater, "淺灘"},
		{4, u5data.SpawnTerrainLand, "沼澤算陸地"},
		{5, u5data.SpawnTerrainLand, "草地"},
		{0x0B, u5data.SpawnTerrainLand, "丘陵"},
		{0x0C, u5data.SpawnTerrainNone, "★ 山不生"},
		{0x0D, u5data.SpawnTerrainNone, "★ 高峰不生"},
		{0x0F, u5data.SpawnTerrainLand, "0x0F 仍是陸地"},
		{0x10, u5data.SpawnTerrainNone, "★ >= 0x10 而不在 0x30..0x33"},
		{0x30, u5data.SpawnTerrainLand, "★ 0x30..0x33 是例外"},
		{0x33, u5data.SpawnTerrainLand, "同上"},
		{0x34, u5data.SpawnTerrainNone, "0x34 不在例外裡"},
		{0x60, u5data.SpawnTerrainWater, "0x60..0x6F 水族"},
		{0x6F, u5data.SpawnTerrainWater, "同上"},
		{0x70, u5data.SpawnTerrainNone, "0x70 既非水也不能生"},
		{0xD4, u5data.SpawnTerrainWater, "0xD4..0xD7"},
		{0xE4, u5data.SpawnTerrainWater, "0xE4..0xE7"},
		{0xE8, u5data.SpawnTerrainNone, "0xE8 出了水族範圍"},
	} {
		if got := u5data.SpawnTerrainOf(tc.tile); got != tc.want {
			t.Errorf("tile 0x%02X → %d,預期 %d(%s)", tc.tile, got, tc.want, tc.why)
		}
	}
}

// TestIsCreatureTileIsNotTheSameAsIsCreature —— ★ 兩個判準各有出處,不要合併。
func TestIsCreatureTileIsNotTheSameAsIsCreature(t *testing.T) {
	for _, tc := range []struct {
		kind byte
		want bool
		why  string
	}{
		{0x2C, true, "★ 船算(世界回合要讓敵船動)"},
		{0x2F, true, "船的四個朝向"},
		{0x30, false, "0x30..0x7F 不算"},
		{0x40, false, "★ 這裡不算 —— 而 MapObject.IsCreature() 算"},
		{0x7F, false, ""},
		{0x80, true, "生物起點"},
		{0xB3, true, ""},
		{0xB4, false, "★ 四個信物"},
		{0xB7, false, "同上"},
		{0xB8, true, ""},
		{0xE7, true, ""},
		{0xE8, false, "0xE8..0xEB 不算"},
		{0xEB, false, ""},
		{0xEC, true, "深水那個沒名字的"},
		{0xFF, true, ""},
		{0x00, false, "空槽"},
	} {
		if got := u5data.IsCreatureTile(tc.kind); got != tc.want {
			t.Errorf("kind 0x%02X → %v,預期 %v(%s)", tc.kind, got, tc.want, tc.why)
		}
	}
	// ★ 兩個判準真的不同 —— 0x40 是分歧點。
	o := u5data.MapObject{Kind: 0x40}
	if !o.IsCreature() {
		t.Error("MapObject.IsCreature() 該把 0x40 當生物")
	}
	if u5data.IsCreatureTile(0x40) {
		t.Error("IsCreatureTile 不該把 0x40 當生物 —— 兩個判準的分歧點就在這裡")
	}
}

// TestSpawnStaysSevenTilesAway —— `sub_215C` 的距離限制。
//
// ★ 兩軸都要離隊伍 7 格以上,所以怪不會在眼前冒出來。
func TestSpawnStaysSevenTilesAway(t *testing.T) {
	for _, tc := range []struct {
		dx, dy int
		want   bool
	}{
		{0, 0, false},
		{6, 20, false},   // X 太近
		{20, 6, false},   // Y 太近
		{7, 7, true},     // 剛好夠
		{-7, -7, true},   // 負向也算
		{250, 20, false}, // ★ 環繞之後其實只差 6
		{20, 250, false},
		{-250, 20, false},
	} {
		if got := u5data.SpawnFarEnough(tc.dx, tc.dy); got != tc.want {
			t.Errorf("(%d, %d) → %v,預期 %v", tc.dx, tc.dy, got, tc.want)
		}
	}
}

// TestSandTrapOnlyInScorchedDesertAndRotWormOnlyInUnderworldSwamp —— 兩個地形特例。
func TestSandTrapOnlyInScorchedDesertAndRotWormOnlyInUnderworldSwamp(t *testing.T) {
	s := overworldScene(t)

	// 焦灼荒漠(tile 7):只可能生 Sand Trap 或不生。
	s.Floor = 0
	for i := 0; i < 400; i++ {
		if k := s.spawnKindFor(7); k != 0 && k != u5data.SpawnScorchedSpecial {
			t.Fatalf("焦灼荒漠生出了 0x%02X,原版只有 Sand Trap(0x%02X)", k, u5data.SpawnScorchedSpecial)
		}
	}

	// 幽冥界的沼澤:一定是 Rot Worm。
	s.Floor = UnderworldFloor
	for i := 0; i < 100; i++ {
		if k := s.spawnKindFor(TileSwamp); k != u5data.SpawnUnderworldSwamp {
			t.Fatalf("幽冥界沼澤生出了 0x%02X,原版是 Rot Worm(0x%02X)", k, u5data.SpawnUnderworldSwamp)
		}
	}

	// ★ 地表的沼澤走一般陸地表,**不是** Rot Worm。
	s.Floor = 0
	for i := 0; i < 400; i++ {
		if k := s.spawnKindFor(TileSwamp); k == u5data.SpawnUnderworldSwamp {
			t.Fatal("地表的沼澤生出了 Rot Worm —— 原版比的是樓層恰好 −1")
		}
	}
}

// TestEnemyShipCarriesFullHull —— ★ 「0x2C 是船不是怪」的證據就是這一筆。
func TestEnemyShipCarriesFullHull(t *testing.T) {
	s := overworldScene(t)
	s.placeWanderer(u5data.SpawnEnemyShip, s.X+10, s.Y+10)
	set := s.currentObjects()
	if set == nil {
		t.Fatal("沒有物件表")
	}
	found := false
	for i := range set.Objects {
		if set.Objects[i].Raw[u5data.ObjKind] != u5data.SpawnEnemyShip {
			continue
		}
		found = true
		if got := int(set.Objects[i].Raw[u5data.ObjQuality]); got != u5data.SpawnEnemyShipHull {
			t.Errorf("敵船的耐久是 %d,原版寫 %d", got, u5data.SpawnEnemyShipHull)
		}
	}
	if !found {
		t.Fatal("敵船沒有被放進物件槽")
	}

	// ★ 一般怪物不帶耐久。
	s2 := overworldScene(t)
	s2.placeWanderer(0xC0, s2.X+10, s2.Y+10) // Orc
	for i := range s2.currentObjects().Objects {
		o := &s2.currentObjects().Objects[i]
		if o.Raw[u5data.ObjKind] == 0xC0 && o.Raw[u5data.ObjQuality] != 0 {
			t.Errorf("Orc 也被寫了耐久 %d", o.Raw[u5data.ObjQuality])
		}
	}
}

// TestSpawnSlotPrefersEmptyThenFarThings —— `sub_2B57C` 的十趟優先序。
func TestSpawnSlotPrefersEmptyThenFarThings(t *testing.T) {
	s := overworldScene(t)
	set := s.currentObjects()
	if set == nil {
		t.Fatal("沒有物件表")
	}

	// 全部填滿「近處的生物」,只留一個空槽 → 該挑那個空槽。
	for i := u5data.SpawnSlotFirst; i <= u5data.SpawnSlotLast; i++ {
		set.Objects[i].Raw[u5data.ObjKind] = 0xC0
		set.Objects[i].Raw[u5data.ObjX] = byte(s.X)
		set.Objects[i].Raw[u5data.ObjY] = byte(s.Y)
	}
	const empty = 9
	set.Objects[empty].Raw[u5data.ObjKind] = 0
	if slot, ok := s.freeSpawnSlot(); !ok || slot != empty {
		t.Errorf("沒挑空槽,挑了 %d(ok=%v)", slot, ok)
	}

	// 沒有空槽時:遠處的**物品**該比近處的生物先被回收。
	set.Objects[empty].Raw[u5data.ObjKind] = 0xC0
	const farItem = 4
	set.Objects[farItem].Raw[u5data.ObjKind] = 0x05 // 物品範圍 1..0x0F
	set.Objects[farItem].Raw[u5data.ObjX] = byte(s.X + 20)
	set.Objects[farItem].Raw[u5data.ObjY] = byte(s.Y + 20)
	if slot, ok := s.freeSpawnSlot(); !ok || slot != farItem {
		t.Errorf("沒挑遠處的物品,挑了 %d(ok=%v)", slot, ok)
	}

	// ★ 0x B5 永不被回收 —— 全部填 0xB5 時最後那一趟也要跳過它。
	for i := u5data.SpawnSlotFirst; i <= u5data.SpawnSlotLast; i++ {
		set.Objects[i].Raw[u5data.ObjKind] = u5data.SpawnSlotProtected
	}
	if _, ok := s.freeSpawnSlot(); ok {
		t.Error("0xB5 被回收了 —— 原版 `cmp al, 0B5h; jz → 跳過`")
	}
}

// TestSpawnNeverUsesTheReservedHighSlots —— ★ 只掃 1..0x17,高 8 槽是保留區。
//
// 世界回合掃到 0x1F,生怪只掃到 0x17 —— 這個不對稱是原版的。
func TestSpawnNeverUsesTheReservedHighSlots(t *testing.T) {
	s := overworldScene(t)
	set := s.currentObjects()
	if set == nil {
		t.Fatal("沒有物件表")
	}
	// 1..0x17 全部塞滿不可回收的 0xB5,0x18..0x1F 留空 → 該找不到槽。
	for i := u5data.SpawnSlotFirst; i <= u5data.SpawnSlotLast; i++ {
		set.Objects[i].Raw[u5data.ObjKind] = u5data.SpawnSlotProtected
	}
	for i := u5data.SpawnSlotLast + 1; i < len(set.Objects); i++ {
		set.Objects[i].Raw[u5data.ObjKind] = 0
	}
	if slot, ok := s.freeSpawnSlot(); ok {
		t.Errorf("用到了保留槽 %d —— 原版只掃 1..0x%02X", slot, u5data.SpawnSlotLast)
	}
}
