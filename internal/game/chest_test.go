package game

import (
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// findDungeonTile 在八座地牢裡找一格符合條件的,把玩家放上去。
func findDungeonTile(t *testing.T, s *State, want byte) bool {
	t.Helper()
	for dg := 0; dg < u5data.DungeonCount; dg++ {
		for l := 0; l < u5data.DungeonLevels; l++ {
			for y := 0; y < u5data.DungeonSide; y++ {
				for x := 0; x < u5data.DungeonSide; x++ {
					if u5data.DungeonKind(s.Dungeons.At(dg, l, x, y)) != want {
						continue
					}
					s.Dungeon = &DungeonState{Index: dg,
						Location: u5data.DungeonLocationBase + dg,
						Level:    l, X: x, Y: y, Facing: South}
					s.Location = s.Dungeon.Location
					return true
				}
			}
		}
	}
	return false
}

// TestOpenChestTurnsItIntoAnOpenedOne:開過的寶箱格會變成 0x70,而且保留位元 3。
//
// `sub_18D18` 是 `tile = (tile & 8) | 0x70` —— 所以 **0x70 不是門,
// 是「開過的寶箱」**。這條同時把那個位元的語意釘住。
func TestOpenChestTurnsItIntoAnOpenedOne(t *testing.T) {
	s := dungeonState(t)
	if !findDungeonTile(t, s, u5data.DungeonChest) {
		t.Skip("八座地牢裡找不到寶箱")
	}
	d := s.Dungeon
	// 手動加上「頭上有洞」,驗它會被保留。
	orig := s.DungeonTileHere() | u5data.DungeonHoleAbove
	s.Dungeons.Set(d.Index, d.Level, d.X, d.Y, orig)
	s.OpenChest()
	got := s.DungeonTileHere()
	if got&0xF0 != 0x70 {
		t.Errorf("開完之後是 %02X,預期高四位元 0x70", got)
	}
	if got&u5data.DungeonHoleAbove == 0 {
		t.Error("「頭上有洞」那一位元被清掉了")
	}
	// 開過的箱子不能再開。
	before := s.Inventory.Gold
	s.OpenChest()
	if s.Inventory.Gold != before {
		t.Error("同一個箱子開了兩次還有錢")
	}
}

// TestTrappedChestHurtsSomeone:bit 0 設著的寶箱會傷人。
func TestTrappedChestHurtsSomeone(t *testing.T) {
	s := dungeonState(t)
	s.SeedRandom(4)
	if !findDungeonTile(t, s, u5data.DungeonChest) {
		t.Skip("找不到寶箱")
	}
	d := s.Dungeon
	s.Dungeons.Set(d.Index, d.Level, d.X, d.Y,
		u5data.DungeonChest|u5data.ChestTrappedDungeon)
	total := 0
	for i := 0; i < s.PartySize; i++ {
		s.Roster[i].HP, s.Roster[i].MaxHP = 200, 200
		total += int(s.Roster[i].HP)
	}
	s.OpenChest()
	after := 0
	for i := 0; i < s.PartySize; i++ {
		after += int(s.Roster[i].HP)
	}
	if after >= total {
		t.Errorf("有陷阱的箱子開下去沒人受傷(%d → %d):\n%s", total, after, s.log())
	}
}

// TestChestContentsScaleWithDepth:越深的寶箱掉得越多。
//
// 門檻同時是「等級下限」與「擲骰難度」,所以第 1 層幾乎只掉金幣,
// 第 8 層才有機會掉到門檻 25 那一項。
func TestChestContentsScaleWithDepth(t *testing.T) {
	s := magicState(t)
	s.SeedRandom(2)
	count := func(level, tries int) int {
		n := 0
		for i := 0; i < tries; i++ {
			if s.rollChestContents(level) {
				n++
			}
		}
		return n
	}
	// ⚠ 獎品表最低的門檻是 3,所以**等級 1 或 2 的寶箱永遠是空的**。
	// 這是表本身的性質,不是判斷寫反 —— 我第一版拿深度 1..8 當等級,
	// 第 1 層的箱子就全空了,那正是「等級的來源還沒逆對」的訊號。
	if count(2, 200) != 0 {
		t.Error("等級 2 低於所有門檻,不該掉出東西")
	}
	shallow := count(4, 200)
	deep := count(30, 200)
	if shallow == 0 {
		t.Error("等級 4 過得了門檻 3,200 次全空 —— 門檻判斷大概反了")
	}
	if deep <= shallow {
		t.Errorf("深處的寶箱 %d 次有東西,淺處 %d 次 —— 應該更多", deep, shallow)
	}
	t.Logf("等級 4:%d/200 有東西;等級 30:%d/200", shallow, deep)
}

// TestAnGravDestroysField:破力場把 0x8x 清成只剩位元 3。
func TestAnGravDestroysField(t *testing.T) {
	s := dungeonState(t)
	if !findDungeonTile(t, s, u5data.DungeonMagic) {
		t.Skip("八座地牢裡找不到力場")
	}
	if !s.destroyField() {
		t.Fatalf("破不了力場:\n%s", s.log())
	}
	if got := s.DungeonTileHere(); got&^u5data.DungeonHoleAbove != 0 {
		t.Errorf("破完之後是 %02X,預期只剩位元 3", got)
	}
	// 破過就沒了。
	if s.destroyField() {
		t.Error("同一個力場破了兩次")
	}
}

// TestUnlockLoweringTheTileByOne:解鎖就是 tile − 1。
func TestUnlockLoweringTheTileByOne(t *testing.T) {
	s := magicState(t)
	s.Location = 0
	// 在旁邊放一扇鎖著的門。
	if !s.SetTileAt(s.X+1, s.Y, u5data.TileLockedDoor) {
		t.Skip("這一層的地圖不支援寫入")
	}
	if !s.UnlockAhead(East) {
		t.Fatalf("解不開:\n%s", s.log())
	}
	if got := s.TileAt(s.X+1, s.Y); got != u5data.TileLockedDoor-1 {
		t.Errorf("解鎖之後 tile 是 %02X,預期 %02X", got, u5data.TileLockedDoor-1)
	}
	// 沒鎖的東西解不了。
	if s.UnlockAhead(East) {
		t.Error("已經解開的門又解了一次")
	}
	// 0xBB 那一扇也是同一條規則。
	s.SetTileAt(s.X+1, s.Y, u5data.TileLockedMagicDoor)
	if !s.UnlockAhead(East) {
		t.Error("0xBB 解不開")
	}
	if got := s.TileAt(s.X+1, s.Y); got != u5data.TileLockedMagicDoor-1 {
		t.Errorf("0xBB 解鎖之後是 %02X", got)
	}
}

// ★ 地表寶箱的種類碼就是 `sub_154BC` 的物品碼 —— 兩張獨立的表用同一組編碼。
func TestSurfaceChestKindsAreItemCodes(t *testing.T) {
	want := map[byte]string{
		u5data.ItemClosedChest: "上鎖的箱子",
		u5data.ItemGold:        "金幣",
		u5data.ItemPotion:      "藥水",
		u5data.ItemPlans:       "卷軸",
		u5data.ItemKey:         "鑰匙",
		u5data.ItemGem:         "寶石",
		u5data.ItemTorch:       "火把",
		u5data.ItemFood:        "食物",
	}
	for _, k := range u5data.ChestKind {
		if _, ok := want[k]; !ok {
			t.Errorf("寶箱種類碼 %d 不在 sub_154BC 的物品碼裡", k)
		}
	}
	if len(u5data.ChestKind) != len(want) {
		t.Errorf("寶箱有 %d 種,物品碼列了 %d 種", len(u5data.ChestKind), len(want))
	}
	// ⚠ 檀香木盒(0x0E)**刻意不在**任何獎品表裡 —— 它是唯一的劇情物品。
	for _, k := range u5data.ChestKind {
		if k == u5data.ItemSandalwood {
			t.Error("檀香木盒不該從寶箱掉出來")
		}
	}
	for _, k := range u5data.DungeonLootKind {
		if k == u5data.ItemSandalwood {
			t.Error("檀香木盒不該從地牢寶箱掉出來")
		}
	}
}

// 寶箱掉出來的東西真的進背包(原本只印一行「找到了什麼東西」)。
func TestChestLootReachesTheBackpack(t *testing.T) {
	s := getScene(t, 0, 0)
	s.sceneObjects.Objects[3] = u5data.MapObject{}
	before := s.Inventory.Gold + s.Inventory.Gems + s.Inventory.Keys +
		s.Inventory.Torches + s.Inventory.Food
	// 等級開到最大,八種門檻全過的機率很高;跑幾次總會拿到東西。
	got := false
	for i := 0; i < 40 && !got; i++ {
		s.rollChestContents(99)
		after := s.Inventory.Gold + s.Inventory.Gems + s.Inventory.Keys +
			s.Inventory.Torches + s.Inventory.Food
		got = after > before
	}
	if !got {
		t.Error("開了四十次高等寶箱,背包一樣都沒多 —— 獎品沒接進背包")
	}
}
