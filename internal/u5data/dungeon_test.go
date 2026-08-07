package u5data

import (
	"os"
	"testing"
)

func loadDungeonSet(t *testing.T) *DungeonSet {
	t.Helper()
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	d, err := LoadDungeons(dir)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// TestDungeonShape:4096 B 要正好切成 8 座 × 8 層 × 8×8。
func TestDungeonShape(t *testing.T) {
	if DungeonFileB != 4096 {
		t.Fatalf("形狀算出來是 %d B,DUNGEON.DAT 是 4096 B", DungeonFileB)
	}
	if _, err := ParseDungeons(make([]byte, 4095)); err == nil {
		t.Error("大小不對卻讀得起來")
	}
}

// TestEveryLevelHasFloorAndWall:64 層每一層都要同時有通道與牆。
//
// 這條擋的是「8×8 切錯」——切成 16×16 或 4×4 都會讓某些層變成整片同一個值。
func TestEveryLevelHasFloorAndWall(t *testing.T) {
	s := loadDungeonSet(t)
	for d := 0; d < DungeonCount; d++ {
		for l := 0; l < DungeonLevels; l++ {
			passage, wall := 0, 0
			for i := 0; i < DungeonLevelB; i++ {
				switch DungeonKind(s.Dungeons[d].Levels[l][i]) {
				case DungeonPassage:
					passage++
				case DungeonWall:
					wall++
				}
			}
			if passage == 0 || wall == 0 {
				t.Errorf("%s 第 %d 層:通道 %d 格、牆 %d 格 —— 兩者都該有",
					DungeonEntrances[d].Name, l+1, passage, wall)
			}
		}
	}
}

// TestEveryDungeonIsTraversable:每一層都要有往下的路,而入口一定不是死路。
//
// ⚠ 「往下的路」不只是梯子。實測發現四座地牢有整層一個下行梯都沒有
//(Destard 第 4 層、Wrong 第 1 層、Covetous 第 6 層、Doom 第 3 層)——
// 那幾層是靠**房間**接下去的(房間的 11×11 地圖裡自己有梯子)。
// 我第一版把條件寫成「每層都要有梯子」,四座地牢一起紅 —— 斷言寫太死。
//
// 入口 (1,1) 也一樣:五座是上行梯(0x10 / 0x30)、Destard 是一個
// **頭上的洞**(位元 3)、Shame 與 Doom 直接是**房間**(0xF0)。
func TestEveryDungeonIsTraversable(t *testing.T) {
	s := loadDungeonSet(t)
	for d := 0; d < DungeonCount; d++ {
		name := DungeonEntrances[d].Name
		// 入口那一格一定是「有作用」的東西,不會是牆或空通道。
		entry := s.Dungeons[d].Levels[0][1*DungeonSide+1]
		ok := DungeonCanClimbUp(entry, true) || DungeonIsRoom(entry)
		if !ok {
			t.Errorf("%s 的入口 (1,1) 是 %02X —— 既不是梯子 / 洞也不是房間", name, entry)
		}
		// 最底層以外,每一層都要有往下的路(梯子、坑,或一間房)。
		for l := 0; l < DungeonLevels-1; l++ {
			down, rooms := 0, 0
			for i := 0; i < DungeonLevelB; i++ {
				tile := s.Dungeons[d].Levels[l][i]
				if DungeonCanClimbDown(tile) {
					down++
				}
				if DungeonIsRoom(tile) {
					rooms++
				}
			}
			if down == 0 && rooms == 0 {
				t.Errorf("%s 第 %d 層既沒有下行的路也沒有房間 —— 走到這裡就斷了",
					name, l+1)
			}
		}
	}
}

// TestDungeonRoomNumbersAreInRange:地圖上出現的房號都要查得到房間。
func TestDungeonRoomNumbersAreInRange(t *testing.T) {
	s := loadDungeonSet(t)
	rooms, err := LoadDungeonRooms(os.Getenv("U5_GAMEDATA"))
	if err != nil {
		t.Fatal(err)
	}
	seen := 0
	for d := 0; d < DungeonCount; d++ {
		loc := DungeonLocationBase + d
		for l := 0; l < DungeonLevels; l++ {
			for i := 0; i < DungeonLevelB; i++ {
				tile := s.Dungeons[d].Levels[l][i]
				if !DungeonIsRoom(tile) {
					continue
				}
				seen++
				idx := DungeonRoomIndex(loc, tile)
				if idx < 0 || idx >= len(rooms.Maps) {
					t.Fatalf("%s 第 %d 層有房號 %d → 索引 %d,而 DUNGEON.CBT 只有 %d 間",
						DungeonEntrances[d].Name, l+1, DungeonRoomNumber(tile),
						idx, len(rooms.Maps))
				}
			}
		}
	}
	if seen == 0 {
		t.Error("整個 DUNGEON.DAT 一間房都沒有?房間的高位元大概認錯了")
	}
	t.Logf("八座地牢共 %d 個房間格子,DUNGEON.CBT 有 %d 間房", seen, len(rooms.Maps))
}

// TestDungeonRoomsHaveMonsters:地牢房間的怪物欄位要放得下真的生物編號。
//
// `DUNGEON.CBT` 的偏移 171 那 16 B 若不是怪物,值不會幾乎全是「4 的倍數 + 64」。
func TestDungeonRoomsHaveMonsters(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA")
	}
	rooms, err := LoadDungeonRooms(dir)
	if err != nil {
		t.Fatal(err)
	}
	creature, other, roomsWith := 0, 0, 0
	for i := range rooms.Maps {
		any := false
		for _, k := range rooms.Maps[i].EnemyKind {
			switch {
			case k == 0:
			case k >= CreatureBase && (k-CreatureBase)%4 == 0:
				creature++
				any = true
			default:
				other++
			}
		}
		if any {
			roomsWith++
		}
	}
	if roomsWith < len(rooms.Maps)/2 {
		t.Errorf("%d 間房裡只有 %d 間有怪物 —— 偏移 171 大概認錯了",
			len(rooms.Maps), roomsWith)
	}
	// ⚠ 這一欄不是只放怪物:小值(1..15、30、31)與 235/237/238/239 這幾個
	// 不是 4 的倍數的值也出現不少(約四分之一),看起來是**房間裡的物件**
	// (寶箱之類)。所以只要求「合法生物編號佔多數」,不要求全部。
	if creature <= other {
		t.Errorf("合法生物編號 %d 個,其他 %d 個 —— 生物該佔多數", creature, other)
	}
	t.Logf("%d 間房、%d 間有怪物;合法生物編號 %d 個、其他(疑為物件)%d 個",
		len(rooms.Maps), roomsWith, creature, other)
}

// TestDungeonEntrancesAreDistinct:八個入口座標不能重複。
func TestDungeonEntrancesAreDistinct(t *testing.T) {
	seen := map[[2]int]string{}
	for i, e := range DungeonEntrances {
		k := [2]int{e.X, e.Y}
		if prev, dup := seen[k]; dup {
			t.Errorf("%s 與 %s 的入口都在 (%d,%d)", prev, e.Name, e.X, e.Y)
		}
		seen[k] = e.Name
		if _, ok := DungeonAt(e.X, e.Y); !ok {
			t.Errorf("查不到第 %d 座地牢的入口", i)
		}
	}
	// Doom 在世界地圖正中央 —— 那是地底世界的入口。
	if DungeonEntrances[7].X != 128 || DungeonEntrances[7].Y != 128 {
		t.Errorf("Doom 在 (%d,%d),預期 (128,128)",
			DungeonEntrances[7].X, DungeonEntrances[7].Y)
	}
}

// TestDungeonRoomBlockSharesFirstTwo:前兩座地牢共用同一批房間。
//
// 這是原版 `sub_42CC` 的 `if (idx >= 1) idx--`,不是我算錯 ——
// 39424 / 5632 = 7 個區塊配 8 座地牢,少的那一個就是這樣補的。
func TestDungeonRoomBlockSharesFirstTwo(t *testing.T) {
	if a, b := DungeonRoomBlock(0x21), DungeonRoomBlock(0x22); a != b {
		t.Errorf("前兩座地牢的房間區塊是 %d 與 %d,原版讓它們共用", a, b)
	}
	if got := DungeonRoomBlock(0x28); got != 6 {
		t.Errorf("Doom 的房間區塊是 %d,預期 6", got)
	}
}
