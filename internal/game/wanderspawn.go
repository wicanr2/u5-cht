package game

import "github.com/wicanr2/u5-cht/internal/u5data"

// 大地圖的遊蕩怪生成(原版 `sub_2218` → `sub_215C` / `sub_203C` / `sub_2B57C`)
//
// 表格與門檻在 `internal/u5data/spawn.go`,推導在 `docs/re/82`。
// 這一檔是需要物件槽與隊伍座標的那一半。

// SpawnAttempts 是找落點的重試上限(原版 `for (i = 0; i < 0x80; i++)`)。
//
// ★ 128 次全失敗就**這一回合不生**,不是「一定要生出來」。海上一片深水時
// 水域那條路有 3/4 的機率放棄,所以失敗是常態。
const SpawnAttempts = 0x80

// spawnWanderingCreature 是一次生怪嘗試(原版 `sub_2218`)。
//
//	for (i = 0; i < 0x80; i++) {
//	    sub_215C()                              ; 挑候選座標
//	    kind = sub_203C(tile(候選))              ; 依地形選一種;0 = 這裡不行
//	    if (kind == 0) continue
//	    if (kind != 0x2C) break
//	    if ((tile & 0xF0) == 0x60) break        ; ★ 敵船只能生在那一族水上
//	}
//	if (i == 0x80) return
//	slot = sub_2B57C()
//	sub_2B6C8(kind, kind, X, Y, 樓層, 0, slot)
//	if (kind == 0x2C) 物件[slot].+5 = 100       ; 船的耐久
func (s *State) spawnWanderingCreature() {
	for i := 0; i < SpawnAttempts; i++ {
		x, y, ok := s.spawnCandidate()
		if !ok {
			continue
		}
		tile := s.TileAt(x, y)
		kind := s.spawnKindFor(tile)
		if kind == 0 {
			continue
		}
		if kind == u5data.SpawnEnemyShip && !u5data.SpawnWaterFamily(tile) {
			continue
		}
		s.placeWanderer(kind, x, y)
		return
	}
}

// spawnCandidate 挑一個落點(原版 `sub_215C`)。
//
// 原版是 do-while 一直重擲到合格;這裡回傳 `ok=false` 讓外層那個 128 次的
// 迴圈負責重試 —— 兩者的期望行為相同,而**不會有無限迴圈**。
// (原版那個 do-while 在隊伍貼著視窗邊緣時理論上會轉很久,實務上不會卡死。)
func (s *State) spawnCandidate() (int, int, bool) {
	ox, oy := s.spawnWindowOrigin()
	x := (ox + s.Roll(0, u5data.SpawnWindowSpan)) & 0xFF
	y := (oy + s.Roll(0, u5data.SpawnWindowSpan)) & 0xFF
	if !u5data.SpawnFarEnough(x-s.X, y-s.Y) {
		return 0, 0, false
	}
	return x, y, true
}

// spawnWindowOrigin 是載入視窗的左上角(原版 `byte_3E0AB` / `byte_3E0AC`)。
//
// ✅ **不再是近似**(`docs/re/88`):原點對齊 16 的倍數,只在隊伍走進邊緣 5 格
// 以內才整塊捲 16 格。維護在 `internal/game/loadwindow.go`。
//
// ⚠ 此前用「隊伍為中心的 32×32」當近似,而差別是可觀察的:真視窗的原點
// 在同一塊裡**不動**,所以生怪的那一圈落在固定位置;近似版永遠貼著隊伍。
func (s *State) spawnWindowOrigin() (int, int) {
	return s.WindowX, s.WindowY
}

// spawnKindFor 依地形選一種怪(原版 `sub_203C`)。回 0 = 這一格不生。
func (s *State) spawnKindFor(tile byte) byte {
	under := s.Floor != 0
	switch u5data.SpawnTerrainOf(tile) {
	case u5data.SpawnTerrainNone:
		return 0

	case u5data.SpawnTerrainWater:
		// ★ 水域那一族先擲一次「大多數時候放棄」——
		// `random(0, 64) >= 0x10` 就不生(留下 16/65 ≈ 24.6%)。
		if s.Roll(0, u5data.WaterSpawnGiveUpRollHi) >= u5data.WaterSpawnKeepBelow {
			return 0
		}
		if !under {
			// ★ 深水(tile 1)有 1/8 生那個沒有名字的 0xEC。
			if tile == u5data.RoughSeasTile && s.Roll(0, 7) == 7 {
				return u5data.SpawnDeepWaterSpecial
			}
			return s.pick(u5data.SpawnWaterSurface, u5data.SpawnWaterSurfaceWeights)
		}
		return s.pick(u5data.SpawnWaterUnderworld, u5data.SpawnWaterUnderworldWeights)

	default: // SpawnTerrainLand
		// ★ 焦灼荒漠(tile 7)只生 Sand Trap,而且 3/4 的機率什麼都不生。
		if tile == 7 {
			if s.Roll(0, 3) != 0 {
				return 0
			}
			return u5data.SpawnScorchedSpecial
		}
		// ★ 幽冥界的沼澤只生 Rot Worm(原版比的是 `byte_3E0A5 == 0FFh`,
		// 也就是**恰好** −1 那一層,不是「任何負的樓層」)。
		if tile == TileSwamp && s.Floor == UnderworldFloor {
			return u5data.SpawnUnderworldSwamp
		}
		if !under {
			return s.pick(u5data.SpawnLandSurface, u5data.SpawnLandSurfaceWeights)
		}
		return s.pick(u5data.SpawnLandUnderworld, u5data.SpawnLandUnderworldWeights)
	}
}

// pick 依權重挑一種(原版 `sub_2008`,擲 `random(0, 0xFF)`)。
func (s *State) pick(kinds, weights []byte) byte {
	i := u5data.SpawnPickWeighted(weights, s.Roll(0, 0xFF))
	if i < 0 || i >= len(kinds) {
		return 0
	}
	return kinds[i]
}

// placeWanderer 把怪寫進物件槽(原版 `sub_2B57C` 找槽 + `sub_2B6C8` 寫入)。
func (s *State) placeWanderer(kind byte, x, y int) {
	slot, ok := s.freeSpawnSlot()
	if !ok {
		return
	}
	set := s.currentObjects()
	if set == nil {
		return
	}
	o := &set.Objects[slot]
	*o = u5data.MapObject{Kind: kind, Tile: kind, X: x, Y: y, Floor: s.Floor}
	o.Raw[u5data.ObjKind] = kind
	o.Raw[u5data.ObjTile] = kind
	o.Raw[u5data.ObjX] = byte(x)
	o.Raw[u5data.ObjY] = byte(y)
	o.Raw[u5data.ObjFloor] = byte(s.Floor)
	// ★ 敵船帶滿耐久 —— 這一筆同時是「0x2C 是船不是怪」的證據(`docs/re/82` §4)。
	if kind == u5data.SpawnEnemyShip {
		o.Raw[u5data.ObjQuality] = u5data.SpawnEnemyShipHull
	}
}

// freeSpawnSlot 依原版的十趟優先序找一個可用槽(`sub_2B57C` → `sub_2B518`)。
//
// ★ 順序是設計:先空槽,再犧牲**遠處的**物品 → 生物 → 坐騎 → 其他,
// 然後才是近處的同四類,最後一趟保證一定找得到。
// ⇒ 玩家眼前的東西不會突然消失,而地上的物品比怪物先被回收。
func (s *State) freeSpawnSlot() (int, bool) {
	set := s.currentObjects()
	if set == nil {
		return 0, false
	}
	for _, pass := range u5data.SpawnSlotPasses {
		for slot := u5data.SpawnSlotFirst; slot <= u5data.SpawnSlotLast && slot < len(set.Objects); slot++ {
			kind := set.Objects[slot].Raw[u5data.ObjKind]
			if kind < pass.Lo || kind > pass.Hi {
				continue
			}
			// ★ 原版只保護 0xB5 一個(四個信物裡的一個)。照原樣。
			if kind == u5data.SpawnSlotProtected {
				continue
			}
			if !pass.RequireFar {
				return slot, true
			}
			dx := int(int8(set.Objects[slot].Raw[u5data.ObjX] - byte(s.X)))
			dy := int(int8(set.Objects[slot].Raw[u5data.ObjY] - byte(s.Y)))
			if absInt(dx) > u5data.SpawnSlotNearRadius || absInt(dy) > u5data.SpawnSlotNearRadius {
				return slot, true
			}
		}
	}
	return 0, false
}

// absInt 是絕對值(原版 `sub_39654`)。
//
// ⚠ 不叫 `abs` —— 那個名字已經被 `shop_test.go` 用掉了,而測試檔裡的宣告
// 一樣會撞到套件命名空間。
func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
