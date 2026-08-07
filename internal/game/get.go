package game

import (
	"fmt"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// Get 指令(原版 `sub_15A94` → `sub_154BC`)
//
// 這是三塊寶石碎片、檀香木盒、王冠、權杖、寶珠、月石、魔毯的**唯一**入口。
// 引擎在此之前完全沒有這個指令 —— 所以真結局那條路是走不到的,
// 只能用 `u5dump` 的腳本直接播。
//
// 流程:問一個方向 → 那一格上有可撿的物件就撿 → 沒有就看地形。

// Get 是玩家按下 G。
func (s *State) Get() {
	if s.InCombat() {
		s.Log(MsgNothingToGet)
		return
	}
	// 地牢裡的 Get 是開寶箱,另一支(原版一進 `sub_15A94` 就轉走)。
	if s.InDungeon() {
		s.getDungeonChest()
		return
	}
	s.AskDirection(func(d Direction) {
		dx, dy := d.Delta()
		s.getAt(dx, dy)
	})
}

// getAt 從 (dx, dy) 那一格拿東西。
func (s *State) getAt(dx, dy int) {
	x, y := s.X+dx, s.Y+dy
	if slot, o := s.pickableAt(x, y); o != nil {
		s.pickUp(o.Kind, int(o.Raw[5]), slot)
		return
	}
	s.getTerrain(x, y, dx, dy)
}

// pickableAt 找那一格上第一個撿得起來的物件。
//
// ⚠ 掃到不可撿的東西(怪物、坐騎)時原版是**繼續往下掃**,不是就此放棄 ——
// 同一格疊了兩樣東西時,這決定了撿到哪一個。
func (s *State) pickableAt(x, y int) (int, *u5data.MapObject) {
	objs := s.currentObjects()
	if objs == nil {
		return 0, nil
	}
	for i := 1; i < u5data.ObjectSlots; i++ {
		o := &objs.Objects[i]
		if !o.Present() || o.X != x || o.Y != y {
			continue
		}
		// 場景與大地圖要比樓層;戰鬥地圖(地點 ≥ 0x80)不比。
		if s.Location < 0x80 && o.Floor != s.Floor {
			continue
		}
		if !u5data.GetPickable(o.Kind) {
			continue
		}
		return i, o
	}
	return 0, nil
}

// pickUp 是 `sub_154BC`:依種類碼把東西收進背包,並清掉那個物件槽。
//
// quality 是物件槽 +5 那個位元組。
//
// ★ 那個欄位原本記著「+5..+7 買馬時清零,但全檔沒有其他讀寫處」——
// **這條被推翻了**:`sub_15A94` 就是把 +5 當數量 / 品質傳進來的,
// 而且哪一塊碎片完全由它決定(`and eax, 3`)。
func (s *State) pickUp(kind byte, quality, slot int) {
	taken := true
	switch {
	case kind == u5data.ItemClosedChest:
		s.Log(MsgOpenItFirst)
		return
	case kind == u5data.ItemGold:
		s.Inventory.Gold = addCapped(s.Inventory.Gold, quality, 9999)
		s.Log(fmt.Sprintf(MsgGotGold, quality))
	case kind == u5data.ItemGem:
		s.Inventory.Gems = addCapped(s.Inventory.Gems, quality, 99)
		s.Log(fmt.Sprintf(MsgGotGems, quality))
	case kind == u5data.ItemKey:
		s.Inventory.Keys = addCapped(s.Inventory.Keys, quality, 99)
		s.Log(fmt.Sprintf(MsgGotKeys, quality))
	case kind == u5data.ItemTorch:
		s.Inventory.Torches = addCapped(s.Inventory.Torches, quality, 99)
		s.Log(fmt.Sprintf(MsgGotTorches, quality))
	case kind == u5data.ItemFood:
		s.Inventory.Food = addCapped(s.Inventory.Food, quality, 9999)
		s.Log(fmt.Sprintf(MsgGotFood, quality))
	case kind == u5data.ItemMagicCarpet:
		s.Inventory.Carpets = addCapped(s.Inventory.Carpets, 1, 99)
		s.Log(MsgGotCarpet)
	case kind == u5data.ItemSandalwood:
		s.SandalwoodBox = true
		s.Log(MsgGotSandalwood)
	case kind == u5data.ItemPlans:
		s.Regalia.Plans = true
		s.Log(MsgGotPlans)
	case kind == u5data.ItemCrown:
		s.Regalia.Crown = true
		s.Log(MsgGotCrown)
	case kind == u5data.ItemSceptre:
		s.Regalia.Sceptre = true
		s.Log(MsgGotSceptre)
	case kind == u5data.ItemOrb:
		s.Regalia.Orb = true
		s.Log(MsgGotOrb)
	case kind == u5data.ItemShard:
		i := u5data.ShardIndex(quality)
		s.Shards[i] = true
		s.Log(fmt.Sprintf(MsgGotShard, u5data.Flames[i].ShardZH))
	case kind == u5data.ItemMoonstone:
		s.Log(MsgGotMoonstone)
	default:
		// 其餘種類碼原版也收(藥水、卷軸、裝備…),但那幾條要先把
		// `byte_3DFD0` / `byte_3E030` / `byte_3E038` 三張表接進背包。
		// 還沒做 —— 誠實回報,不假裝撿到了。
		s.Log(MsgCannotCarry)
		taken = false
	}
	if taken {
		s.clearObject(slot)
	}
}

// clearObject 把物件槽清空 —— 東西已經在背包裡了,地上不該還有一個。
func (s *State) clearObject(slot int) {
	objs := s.currentObjects()
	if objs == nil || slot <= 0 || slot >= u5data.ObjectSlots {
		return
	}
	objs.Objects[slot] = u5data.MapObject{}
}

// getTerrain 是那一格沒有物件時的第二條路。
func (s *State) getTerrain(x, y, dx, dy int) {
	tile := s.TileAt(x, y)
	switch tile {
	case u5data.TileWallTorchA, u5data.TileWallTorchB:
		// ⚠ 牆上的火把**不扣業報** —— 原版還特地說「借用!」。
		if !s.SetTileAt(x, y, u5data.TileBrickFloor) {
			s.Log(MsgNothingToGet)
			return
		}
		if s.TorchTurns < u5data.BorrowedTorchTurns {
			s.TorchTurns = u5data.BorrowedTorchTurns
		}
		s.Log(MsgBorrowed)
		return
	case u5data.TileCrops:
		if !s.SetTileAt(x, y, u5data.TileCropsPicked) {
			s.Log(MsgNothingToGet)
			return
		}
		s.Inventory.Food = addCapped(s.Inventory.Food, 1, 9999)
		s.Log(MsgCropsPicked)
		s.stealKarma()
		return
	case u5data.TilePlateNorth, u5data.TilePlateSouth, u5data.TilePlateMiddle:
		if !u5data.PlateReach(tile, dx, dy) {
			s.Log(MsgCantReachPlate)
			return
		}
		if !s.SetTileAt(x, y, u5data.PlateAfter(tile, dy)) {
			s.Log(MsgNothingToGet)
			return
		}
		s.Inventory.Food = addCapped(s.Inventory.Food, 1, 9999)
		s.Log(MsgMmmmm)
		s.stealKarma()
		return
	}
	s.Log(MsgNothingToGet)
}

// stealKarma 是拿食物的代價(原版 `dec byte_3E098`,而且 0 就不再往下扣)。
func (s *State) stealKarma() {
	if s.Karma <= 0 {
		return
	}
	s.Karma -= u5data.GetKarmaPenalty
}

// addCapped 加上去但不超過上限(原版 `sub_2BBB8(&欄位, 量, 上限)`)。
func addCapped(cur, add, limit int) int {
	cur += add
	if cur > limit {
		cur = limit
	}
	return cur
}

// getDungeonChest 是在地牢裡按 Get(原版 `sub_15930`)。
//
// ⚠ 對象是**已經打開**的寶箱(0x7x)。還沒開的(0x4x)印「得先打開它!」——
// 開箱與取物在原版是兩個指令,不是一個。
func (s *State) getDungeonChest() {
	x, y, tile, ok := s.dungeonFacingTile(u5data.DungeonDoor)
	if !ok {
		// 還沒開的寶箱要給不同的訊息。
		if _, _, t, isChest := s.dungeonFacingTile(u5data.DungeonChest); isChest {
			_ = t
			s.Log(MsgOpenItFirst)
			return
		}
		s.Log(MsgNothingToGet)
		return
	}
	d := s.Dungeon
	s.Dungeons.Set(d.Index, d.Level, x, y, u5data.DungeonEmptiedChest(tile))
	s.Log(MsgChestContents)
	s.rollDungeonLoot(d.Level)
}

// rollDungeonLoot 擲地牢寶箱的獎品。
//
// 七類各擲一次 `random(1, 樓層×4 + 4)`,擲出來 ≥ 門檻才拿得到。
// 所以第一層的箱子只可能有食物與金幣,火把要第五層以後 ——
// 深度直接決定獎品的種類,不是只影響數量。
func (s *State) rollDungeonLoot(floor int) {
	rollMax := u5data.DungeonLootRollMax(floor)
	for i := 0; i < len(u5data.DungeonLootThreshold); i++ {
		if s.Roll(1, rollMax) < int(u5data.DungeonLootThreshold[i]) {
			continue
		}
		if kind, special := u5data.DungeonLootSpecial[i]; special {
			s.pickUp(kind, s.Roll(0, 7), 0)
			continue
		}
		// ⚠ 金幣(索引 1)的數量上限是**算**出來的(樓層×8),不是查表 ——
		// 表裡那一格是 0,照抄查表會讓地牢裡永遠撿不到錢。
		max := int(u5data.DungeonLootMax[i])
		if i == 1 {
			max = floor * 8
		}
		if max < 1 {
			continue
		}
		s.pickUp(u5data.DungeonLootKind[i], s.Roll(1, max), 0)
	}
}

// placeUnderworldItems 進地下世界時把寶珠與碎片放進物件槽(原版 `sub_10B3C`)。
//
// 每次載入地下世界都跑一次,而且是**冪等**的:已經拿到的不再放,
// 已經用掉的(暗影君主被消滅)也不再放。
func (s *State) placeUnderworldItems() {
	if s.Floor >= 0 || s.UnderObjects == nil {
		return
	}
	set := func(slot int, kind byte, x, y, quality int) {
		o := &s.UnderObjects.Objects[slot]
		o.Kind, o.Tile = kind, kind
		o.X, o.Y, o.Floor = x, y, -1
		o.Raw[0], o.Raw[1] = kind, kind
		o.Raw[2], o.Raw[3] = byte(x), byte(y)
		o.Raw[4] = 0xFF
		o.Raw[5] = byte(quality)
	}
	if !s.Regalia.Orb {
		set(u5data.UnderworldOrbSlot, u5data.ItemOrb,
			u5data.UnderworldOrb.X, u5data.UnderworldOrb.Y, u5data.UnderworldOrb.Quality)
	}
	for i := 0; i < u5data.ShadowlordCount; i++ {
		if s.Shards[i] {
			continue
		}
		// ⚠ 暗影君主已被消滅 → 那塊碎片用掉了,不重生。
		if s.ShadowlordAt[i] >= 0x80 {
			continue
		}
		p := u5data.UnderworldShards[i]
		set(u5data.UnderworldShardSlot+i, u5data.ItemShard, p.X, p.Y, p.Quality)
	}
}
