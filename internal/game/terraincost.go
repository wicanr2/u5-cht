package game

// 地形的移動代價(原版 `sub_2D0BC`,由移動後的 `sub_2D174` 呼叫)
//
// 在世界地圖上走進沼澤、樹林、山丘,要付的不是「走不動」而是**時間**:
// 原版讓世界多跑一到兩個回合、時鐘多走 2 或 4 分鐘,然後印
// 「Slow progress!」或「Very slow!」。
//
// ⚠ 那多出來的回合是**完整的世界回合** —— 怪物照樣移動、NPC 照樣排程。
// 所以在山裡走一步,附近的怪會走三步。這才是粗糙地形真正的代價,
// 「多幾分鐘」只是它的副作用。寫成「只加時間」會讓穿越山區變得毫無風險。
//
// # 分級(原版的 switch,照抄)
//
//	tile 5           草地            0 級  ← 特別挑出來,雖然它落在 4..15 裡
//	tile 4, 6, 7, 8  沼澤/灌木/荒漠   1 級
//	tile 9..15       林/丘/山/高峰    2 級
//	tile 0x1E, 0x1F  荒漠(另一組)   1 級
//	其餘             —               0 級
//
// 4..15 這個範圍是**先判上下界再切 9**:`if 4 <= t < 0x10 { if t < 9 → 1 else → 2 }`。
// 草地(5)在範圍內卻是 0 級,所以它在 switch 裡單獨一條 —— 不能寫成
// 「4..8 是 1 級」然後忘了把 5 挖掉。
//
// # 訊息什麼時候印
//
// 只有在**額外的回合什麼事都沒發生**時才印。原版:
//
//	v3 = sub_2E24(tile)          ← 跑一個世界回合,回傳「有沒有出事」
//	if v3 == 0 → "Slow progress!"
//
// 二級要**兩個回合的結果相加**都是 0 才印「Very slow!」。也就是說
// 遇上怪物的那一步不會印慢 —— 玩家已經有別的事要煩了。

// 地形代價的級距。值就是「要多跑幾個世界回合」。
const (
	terrainCostNone = 0
	terrainCostSlow = 1
	terrainCostVery = 2
)

// terrainCostMinutes 是各級要多走的分鐘數(原版 `sub_29304(2)` / `sub_29304(4)`)。
//
// 注意**不是**每個額外回合 1 分鐘 —— 一級走 2 分鐘、二級走 4 分鐘,
// 而額外回合各只有 1 個與 2 個。時間與回合數是分開算的。
var terrainCostMinutes = map[int]int{
	terrainCostSlow: 2,
	terrainCostVery: 4,
}

const (
	tileGrass       = 5
	tileRoughLo     = 4  // 含
	tileRoughHi     = 16 // 不含
	tileRoughSplit  = 9  // 這個值(含)以上是二級
	tileDesertAltLo = 0x1E
	tileDesertAltHi = 0x1F
)

// TerrainCost 回傳走進這一格要付的額外世界回合數。
func TerrainCost(tile int) int {
	switch {
	case tile == tileGrass:
		return terrainCostNone
	case tile == tileDesertAltLo, tile == tileDesertAltHi:
		return terrainCostSlow
	case tile >= tileRoughLo && tile < tileRoughHi:
		if tile < tileRoughSplit {
			return terrainCostSlow
		}
		return terrainCostVery
	default:
		return terrainCostNone
	}
}

// payTerrainCost 付掉粗糙地形的代價:多跑幾個世界回合,再多推時鐘。
//
// 回傳有沒有真的付(給測試與呼叫端判斷)。
func (s *State) payTerrainCost(tile int) bool {
	cost := TerrainCost(tile)
	if cost == terrainCostNone {
		return false
	}
	// 額外的回合是真的回合:怪物與 NPC 都動。
	quiet := true
	for i := 0; i < cost; i++ {
		if s.extraWorldTurn() {
			quiet = false
		}
	}
	s.AdvanceTime(terrainCostMinutes[cost])
	if !quiet {
		return true // 出事了就不印慢 —— 玩家已經有別的事要煩
	}
	switch cost {
	case terrainCostSlow:
		s.Log(MsgSlowProgress)
	case terrainCostVery:
		s.Log(MsgVerySlow)
	}
	return true
}

// extraWorldTurn 跑一個額外的世界回合,回報有沒有出事(遭遇 / 進戰鬥)。
//
// ⚠ 這裡**不推時鐘** —— 時間由 `payTerrainCost` 依級距一次加完
// (原版 `sub_2E24` 只動世界,`sub_29304` 才動時鐘)。混在一起會多算分鐘。
func (s *State) extraWorldTurn() bool {
	before := s.InCombat()
	s.advanceNPCs()
	s.syncNPCObjects()
	return s.InCombat() != before
}
