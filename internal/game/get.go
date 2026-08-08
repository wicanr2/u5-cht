package game

import (
	"fmt"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// Get 指令(原版 `sub_15A94` → `sub_154BC`)
//
// 這是三塊寶石碎片、檀香木盒、王冠、權杖、護符、月石、魔毯的**唯一**入口。
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
	inv := &s.Inventory
	switch kind {
	case u5data.ItemClosedChest:
		// 關著的箱子不能撿 —— 而且**不清槽**,箱子留在原地等 Open。
		s.Log(MsgOpenItFirst)
		return

	case u5data.ItemGold:
		inv.Gold = addCapped(inv.Gold, quality, u5data.GoldLimit)
		s.Log(fmt.Sprintf(MsgGotGold, quality))
	case u5data.ItemFood:
		inv.Food = addCapped(inv.Food, quality, u5data.GoldLimit)
		s.Log(fmt.Sprintf(MsgGotFood, quality))
	case u5data.ItemGem:
		inv.Gems = addCapped(inv.Gems, quality, u5data.CarryLimit)
		s.Log(fmt.Sprintf(MsgGotGems, quality))
	case u5data.ItemTorch:
		inv.Torches = addCapped(inv.Torches, quality, u5data.CarryLimit)
		s.Log(fmt.Sprintf(MsgGotTorches, quality))

	case u5data.ItemKey:
		// ⚠ 鑰匙有**兩個**計數欄。品質最高位被設起來的是「怪鑰匙」,
		// 而且數量取的是清掉最高位之後的值 —— 混在一起算的話,
		// 一把怪鑰匙會變成一百多把普通鑰匙。
		if quality&u5data.KeyOddBit != 0 {
			n := quality &^ u5data.KeyOddBit
			inv.OddKeys = addCapped(inv.OddKeys, n, u5data.CarryLimit)
			s.Log(fmt.Sprintf(MsgGotOddKeys, n))
			break
		}
		inv.Keys = addCapped(inv.Keys, quality, u5data.CarryLimit)
		s.Log(fmt.Sprintf(MsgGotKeys, quality))

	case u5data.ItemPotion:
		// 品質選顏色。
		inv.Potions[quality%u5data.PotionCount] = addCapped(
			inv.Potions[quality%u5data.PotionCount], 1, u5data.CarryLimit)
		s.Log(fmt.Sprintf(MsgGotPotion, u5data.PotionColoursZH[quality%u5data.PotionCount]))

	case u5data.ItemScroll:
		// ★ 圖紙**不是自己的種類碼** —— 它是卷軸裡品質 0xFF 的那一筆。
		if quality == u5data.ItemPlansQuality {
			s.Regalia.Plans = true
			s.Log(MsgGotPlans)
			break
		}
		// ⚠ 計數用**完整**品質當索引,顯示的咒語名卻只取低三位
		//(原版 `and eax, 7` 只套在名字上)。照抄。
		i := quality % u5data.ScrollCount
		inv.Scrolls[i] = addCapped(inv.Scrolls[i], 1, u5data.CarryLimit)
		s.Log(fmt.Sprintf(MsgGotScroll, u5data.ScrollSpells[quality&7]))

	case u5data.ItemMagicCarpet:
		inv.Carpets = addCapped(inv.Carpets, 1, u5data.CarryLimit)
		s.Log(MsgGotCarpet)
		// ⚠⚠ **原版只做暫時移除,沒有寫進存檔的遮罩**(`sub_268(0x16)`,
		// 沒有配套的 `sub_218`)。所以離開城堡再回來,毯子又躺在原地 ——
		// 這是可以刷魔毯的。照抄:「機制與原版一模一樣」包含原版的 bug。
		if s.Location == u5data.CarpetNPCLocation {
			s.removeNPC(u5data.CarpetNPCSlot)
		}

	case u5data.ItemMoonstone:
		// 品質是第幾顆月石。撿起來 = 把地點欄寫回 0xFF(見 `moonstone.go`)。
		if quality >= 0 && quality < u5data.MoonstoneCount {
			inv.Moonstones[quality] = u5data.Moonstone{Location: u5data.MoonstoneInHand}
		}
		s.Log(MsgGotMoonstone)

	case u5data.ItemSandalwood:
		s.SandalwoodBox = true
		s.Log(MsgGotSandalwood)
		// ★ 原版在這裡直接寫 `byte_3E3AF |= 0x80` —— 而
		// `0x3E3AF = 0x3E36C + 16×4 + 3`,也就是**地點 17 的第 31 號 NPC**
		// 那一個永久移除位元。硬編碼的位址獨立指回 `CASTLE.NPC` 槽 31,
		// 與 docs/re/36 由資料檔得到的結論完全對上(兩個來源,同一個答案)。
		s.markNPCRemovedAt(u5data.SandalwoodNPCLocation, u5data.SandalwoodNPCSlot)

	case u5data.ItemCrown:
		s.Regalia.Crown = true
		s.Log(MsgGotCrown)
		// 王冠是唯一走「反查 NPC → 永久移除」那條泛用路徑的(sub_2E0 + sub_218 + sub_268)。
		if i, mirrored := s.npcOfObject(slot); mirrored {
			s.takeNPCObject(i)
		}
	case u5data.ItemSceptre:
		s.Regalia.Sceptre = true
		s.Log(MsgGotSceptre)
	case u5data.ItemAmulet:
		s.Regalia.Amulet = true
		s.Log(MsgGotAmulet)
	case u5data.ItemShard:
		i := u5data.ShardIndex(quality)
		s.Shards[i] = true
		s.Log(fmt.Sprintf(MsgGotShard, u5data.Flames[i].ShardZH))

	default:
		// 種類 5 / 6 / 9 / 10 / 0x0B / 0x0C 全走同一條:品質就是**裝備編號**。
		if kind > u5data.ItemKindEquipMax {
			s.Log(MsgNothingToGet)
			return
		}
		if quality < 0 || quality >= u5data.ItemCount {
			s.Log(MsgNothingToGet)
			return
		}
		// 箭矢與弩矢一次五支,其餘一件。
		n := 1
		if quality == u5data.ItemArrows || quality == u5data.ItemQuarrels {
			n = u5data.AmmoPerPickup
		}
		inv.Items[quality] = addCapped(inv.Items[quality], n, u5data.CarryLimit)
		s.Log(fmt.Sprintf(MsgGotItem, s.equipName(quality)))
	}
	s.clearObject(slot)
}

// equipName 取裝備名。裝備表還沒載入時退回編號 —— 不要靜默印空字串。
func (s *State) equipName(id int) string {
	if s.Items != nil {
		if n := s.Items.Name(byte(id)); n != "" {
			return n
		}
	}
	return fmt.Sprintf("#%d", id)
}

// clearObject 把物件槽清空(原版 `sub_154BC` 收尾的 `sub_2B6C8(0,…,槽)`)。
//
// ⚠⚠ **只清槽,不動 NPC。** 城裡的物品是 NPC 鏡射出來的槽,只清這裡的話
// 下一回合 `sub_1E74` 會照原樣再配一格回來 —— 而原版**就是這樣**。
// 要它不再長回來得另外除籍,而原版是**逐案硬編碼**的,沒有通則:
//
//	檀香木盒  直接寫 `byte_3E3AF |= 0x80`(= 地點 17 槽 31 的永久位元)
//	王冠      `sub_2E0` 反查 NPC → `sub_218` + `sub_268`(唯一走泛用路徑的)
//	魔毯      只有 `sub_268(0x16)`,**沒有** `sub_218` → 離場再回來又長出來
//
// 我第一版在這裡加了「是鏡射就一律永久除籍」的通則 —— 那會把魔毯那個
// 可以刷的行為「修好」,而修好就是與原版不同。已改回逐案處理。
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

// placeUnderworldItems 進地下世界時把護符與碎片放進物件槽(原版 `sub_10B3C`)。
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
	if !s.Regalia.Amulet {
		set(u5data.UnderworldAmuletSlot, u5data.ItemAmulet,
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
