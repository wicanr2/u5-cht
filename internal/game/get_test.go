package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// getScene 造一個場景,在玩家北邊放一個物件。
func getScene(t *testing.T, kind byte, quality int) *State {
	t.Helper()
	s := &State{Scenes: synthScenes(t, walkable(t)), NPCs: &u5data.NPCSet{}, MaxMessages: 16}
	s.Karma = 50
	if err := s.SetScene(britain, 0, 15, 15); err != nil {
		t.Fatalf("進不了不列顛城:%v", err)
	}
	o := &s.sceneObjects.Objects[3]
	o.Kind, o.Tile, o.X, o.Y, o.Floor = kind, kind, 15, 14, 0
	o.Raw[5] = byte(quality)
	return s
}

// getNorth 往北撿。
func getNorth(s *State) {
	s.Get()
	s.AnswerDirection(North)
}

// ★ 三塊寶石碎片撿得起來了 —— 在此之前引擎沒有 Get,這條路根本走不到。
func TestPickingUpTheThreeShards(t *testing.T) {
	for i := 0; i < u5data.ShadowlordCount; i++ {
		s := getScene(t, u5data.ItemShard, i)
		getNorth(s)
		if !s.Shards[i] {
			t.Fatalf("撿了品質 %d 的碎片,Shards[%d] 還是 false", i, i)
		}
		for j := 0; j < u5data.ShadowlordCount; j++ {
			if j != i && s.Shards[j] {
				t.Errorf("撿第 %d 塊卻連第 %d 塊也拿到了", i, j)
			}
		}
		// 訊息要說出是哪一塊。
		if !strings.Contains(allLogs(s), u5data.Flames[i].ShardZH) {
			t.Errorf("訊息沒提到「%s」:%q", u5data.Flames[i].ShardZH, allLogs(s))
		}
		// 地上不該還有一個。
		if s.sceneObjects.Objects[3].Present() {
			t.Error("撿完之後物件槽沒清掉")
		}
	}
}

// ⚠ 是哪一塊碎片看**品質欄**,不是種類碼 —— 四塊共用同一個種類碼 0xB4。
func TestWhichShardComesFromTheQualityByte(t *testing.T) {
	if u5data.ShardIndex(0) != 0 || u5data.ShardIndex(2) != 2 {
		t.Fatal("品質欄的低兩位就是碎片編號")
	}
	// 4 會繞回 0(`and eax, 3`)。
	if u5data.ShardIndex(4) != 0 {
		t.Error("品質 4 應該繞回第 0 塊")
	}
}

// 檀香木盒 —— 真結局的唯一額外條件。
func TestPickingUpTheSandalwoodBox(t *testing.T) {
	s := getScene(t, u5data.ItemSandalwood, 0)
	if s.SandalwoodBox {
		t.Fatal("一開始不該有盒子")
	}
	getNorth(s)
	if !s.SandalwoodBox {
		t.Fatal("撿了盒子卻沒拿到")
	}
}

// 王冠 / 權杖 / 寶珠 / 圖紙各有各的旗標。
func TestPickingUpTheRegalia(t *testing.T) {
	cases := []struct {
		kind byte
		get  func(*State) bool
		name string
	}{
		{u5data.ItemCrown, func(s *State) bool { return s.Regalia.Crown }, "王冠"},
		{u5data.ItemSceptre, func(s *State) bool { return s.Regalia.Sceptre }, "權杖"},
		{u5data.ItemOrb, func(s *State) bool { return s.Regalia.Orb }, "寶珠"},
		{u5data.ItemPlans, func(s *State) bool { return s.Regalia.Plans }, "圖紙"},
	}
	for _, c := range cases {
		s := getScene(t, c.kind, 0)
		getNorth(s)
		if !c.get(s) {
			t.Errorf("撿了%s卻沒拿到", c.name)
		}
	}
}

// 數量類的東西照品質欄加進背包。
func TestPickingUpStackables(t *testing.T) {
	cases := []struct {
		kind byte
		n    int
		get  func(*State) int
		name string
	}{
		{u5data.ItemGold, 250, func(s *State) int { return s.Inventory.Gold }, "金幣"},
		{u5data.ItemGem, 3, func(s *State) int { return s.Inventory.Gems }, "寶石"},
		{u5data.ItemKey, 2, func(s *State) int { return s.Inventory.Keys }, "鑰匙"},
		{u5data.ItemTorch, 5, func(s *State) int { return s.Inventory.Torches }, "火把"},
		{u5data.ItemFood, 30, func(s *State) int { return s.Inventory.Food }, "糧食"},
	}
	for _, c := range cases {
		s := getScene(t, c.kind, c.n)
		getNorth(s)
		if got := c.get(s); got != c.n {
			t.Errorf("撿了 %d 個%s,背包裡是 %d", c.n, c.name, got)
		}
	}
}

// 怪物與坐騎撿不起來,而且原版是**繼續往下掃**不是放棄。
func TestUnpickableObjectsAreSkippedNotFatal(t *testing.T) {
	if u5data.GetPickable(0x40) || u5data.GetPickable(0x10) || u5data.GetPickable(0xB8) {
		t.Error("怪物 / 0x10 / 0xB8 不該撿得起來")
	}
	if !u5data.GetPickable(u5data.ItemMoonstone) || !u5data.GetPickable(u5data.ItemMagicCarpet) {
		t.Error("月石與魔毯撿得起來")
	}

	// 同一格上先放一隻怪物,後面放碎片 —— 應該撿到碎片。
	s := getScene(t, u5data.ItemShard, 1)
	monster := &s.sceneObjects.Objects[2] // 槽 2 排在碎片(槽 3)前面
	monster.Kind, monster.Tile, monster.X, monster.Y, monster.Floor = 0x50, 0x50, 15, 14, 0
	getNorth(s)
	if !s.Shards[1] {
		t.Error("怪物擋在前面時應該繼續往下掃到碎片")
	}
	if !monster.Present() {
		t.Error("怪物不該被撿走")
	}
}

// 牆上的火把:拿得到、不扣業報,而且原版說「借用!」。
func TestBorrowingAWallTorchCostsNoKarma(t *testing.T) {
	s := getScene(t, 0, 0)
	s.sceneObjects.Objects[3] = u5data.MapObject{} // 清掉物件,走地形那條路
	s.SetTileAt(15, 14, u5data.TileWallTorchA)
	getNorth(s)
	if s.TorchTurns != u5data.BorrowedTorchTurns {
		t.Errorf("火把時間是 %d,應該是 %d", s.TorchTurns, u5data.BorrowedTorchTurns)
	}
	if s.Karma != 50 {
		t.Errorf("業報變成 %d —— 牆上的火把不該扣", s.Karma)
	}
	if !strings.Contains(allLogs(s), "借用") {
		t.Errorf("訊息是 %q,原版說的是「借用!」", allLogs(s))
	}
	if s.TileAt(15, 14) != u5data.TileBrickFloor {
		t.Errorf("拿走之後那一格是 %02X,應該變成磚地", s.TileAt(15, 14))
	}
}

// 作物與桌上的食物**要扣業報** —— 那是原版對「偷竊」的定義。
func TestTakingFoodCostsKarma(t *testing.T) {
	s := getScene(t, 0, 0)
	s.sceneObjects.Objects[3] = u5data.MapObject{}
	s.SetTileAt(15, 14, u5data.TileCrops)
	getNorth(s)
	if s.Karma != 49 {
		t.Errorf("業報是 %d,採了作物應該扣 1", s.Karma)
	}
	if s.Inventory.Food != 1 {
		t.Errorf("糧食是 %d,應該 +1", s.Inventory.Food)
	}
	if s.TileAt(15, 14) != u5data.TileCropsPicked {
		t.Error("採完之後那一格應該變成收割後的空地")
	}
	// 業報 0 時不會變成負的。
	s.Karma = 0
	s.SetTileAt(15, 14, u5data.TileCrops)
	getNorth(s)
	if s.Karma != 0 {
		t.Errorf("業報 0 之後變成 %d,不該是負的", s.Karma)
	}
}

// 盤子有方向規矩:得站在桌子的長邊,側面搆不著。
func TestYouMustReachTheePlateFromTheRightSide(t *testing.T) {
	// 0x9A(北半)要從南邊拿 —— 也就是玩家在它南邊、往北伸手(dy = −1)?
	// 不對:原版判的是 dy == 1。dy 是「從玩家往目標」的增量,
	// 所以 dy == 1 代表目標在玩家**南邊**。
	if !u5data.PlateReach(u5data.TilePlateNorth, 0, 1) {
		t.Error("0x9A 要從 dy=1 那個方向拿")
	}
	if u5data.PlateReach(u5data.TilePlateNorth, 0, -1) {
		t.Error("0x9A 從 dy=-1 拿不到")
	}
	if !u5data.PlateReach(u5data.TilePlateSouth, 0, -1) {
		t.Error("0x9B 要從 dy=-1 那個方向拿")
	}
	// 整張桌子:東西向拿不到。
	if u5data.PlateReach(u5data.TilePlateMiddle, 1, 0) ||
		u5data.PlateReach(u5data.TilePlateMiddle, -1, 0) {
		t.Error("0x9C 從東西向搆不著")
	}
	if !u5data.PlateReach(u5data.TilePlateMiddle, 0, 1) {
		t.Error("0x9C 從南北向拿得到")
	}

	// ★ 整張桌子被拿走一半之後**變成剩下的那一半**,可以再拿一次。
	if u5data.PlateAfter(u5data.TilePlateMiddle, 1) != u5data.TilePlateSouth {
		t.Error("從 dy=1 拿走之後應該剩南半")
	}
	if u5data.PlateAfter(u5data.TilePlateMiddle, -1) != u5data.TilePlateNorth {
		t.Error("從 dy=-1 拿走之後應該剩北半")
	}
	if u5data.PlateAfter(u5data.TilePlateNorth, 1) != u5data.TilePlateEmpty {
		t.Error("半張桌子拿完就空了")
	}
}

// 搆不著時什麼都不會發生 —— 不扣業報、不加糧食、地形不變。
func TestUnreachablePlateChangesNothing(t *testing.T) {
	s := getScene(t, 0, 0)
	s.sceneObjects.Objects[3] = u5data.MapObject{}
	// 盤子放在玩家**東邊**,而 0x9C 只能從南北向拿。
	s.SetTileAt(16, 15, u5data.TilePlateMiddle)
	s.Get()
	s.AnswerDirection(East)
	if s.Karma != 50 || s.Inventory.Food != 0 {
		t.Errorf("搆不著卻有副作用:業報 %d 糧食 %d", s.Karma, s.Inventory.Food)
	}
	if s.TileAt(16, 15) != u5data.TilePlateMiddle {
		t.Error("搆不著不該改地形")
	}
	if !strings.Contains(allLogs(s), "碰不到") {
		t.Errorf("訊息是 %q", allLogs(s))
	}
}

// 什麼都沒有時就說沒有。
func TestNothingToGet(t *testing.T) {
	s := getScene(t, 0, 0)
	s.sceneObjects.Objects[3] = u5data.MapObject{}
	getNorth(s)
	if !strings.Contains(allLogs(s), "沒有東西可拿") {
		t.Errorf("訊息是 %q", allLogs(s))
	}
}

// 地牢寶箱的獎品:深度決定**種類**,不只是數量。
//
// 第一層擲得再高也只有 4,拿不到門檻 5 的鑰匙;第八層才可能出藥水。
func TestDungeonLootDepthGatesTheCategories(t *testing.T) {
	if u5data.DungeonLootRollMax(0) != 4 {
		t.Fatalf("第一層的骰子上限是 %d,應該是 4", u5data.DungeonLootRollMax(0))
	}
	// 第一層拿不到的:鑰匙(5)、寶石(10)、火把(20)、藥水與卷軸(25)。
	for i, th := range u5data.DungeonLootThreshold {
		reachable := u5data.DungeonLootRollMax(0) >= int(th)
		want := i < 2 // 只有食物(2)與金幣(4)構得到
		if reachable != want {
			t.Errorf("第一層第 %d 類門檻 %d:可及 %v,預期 %v", i, th, reachable, want)
		}
	}
	// 第八層(樓層 7)擲得到 32,七類全部構得到。
	for i, th := range u5data.DungeonLootThreshold {
		if u5data.DungeonLootRollMax(7) < int(th) {
			t.Errorf("第八層應該構得到第 %d 類(門檻 %d)", i, th)
		}
	}
}

// ⚠ 金幣那一格的表值是 0 —— 照抄查表會讓地牢裡永遠撿不到錢。
func TestDungeonGoldQuantityIsComputedNotTabulated(t *testing.T) {
	if u5data.DungeonLootMax[1] != 0 {
		t.Fatal("表裡金幣那一格本來就是 0(數量另外算)")
	}
	if u5data.DungeonLootKind[1] != u5data.ItemGold {
		t.Fatalf("第 1 類應該是金幣,實得 %02X", u5data.DungeonLootKind[1])
	}
}
