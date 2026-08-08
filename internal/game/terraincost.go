package game

import "github.com/wicanr2/u5-cht/internal/u5data"

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

// 世界回合的三道跳過閘門(原版 `sub_2E24` 開頭)
//
//	if (byte_3E08A == 'T') return 0           ; An Tym 停時 → 整個跳過
//	if (byte_3E08A == 'Q') {                  ; Rel Tym 加速
//	    byte_4FDD5 ^= 1
//	    if (byte_4FDD5 != 0) return 0         ; ★ 隔次跳過
//	}
//	if ((byte_3E08C & 0FEh) == 12h || == 14h) {   ; 馬 或 魔毯
//	    byte_4FDD7 ^= 1
//	    if (byte_4FDD7 != 0) return 0         ; ★ 隔次跳過
//	}
//
// ★★ **兩個都是持久的切換位元,不是機率。** 效果等於「怪物只有一半的行動機會」。
//
// ⚠ `docs/re/38` §5 留了一個問號:「`byte_4FDD7` 與 `sub_2E24` 的其他呼叫點
// 共不共用同一個位元?共用的話騎馬時連一般移動的回合也會隔次跳過。」
// **答案是共用** —— 全檔只有這一個 `byte_4FDD7`,而 `sub_2E24` 的四個呼叫點
// (`sub_2D9D0` 每回合、`sub_2D0BC` 地形代價、`sub_2D2D0`、`sub_2B8CC` 紮營修船)
// 全部經過同一道閘門。所以騎馬 / 坐魔毯時**所有**世界回合都隔次跳過。
type worldTurnGates struct {
	relTymBit bool // byte_4FDD5
	mountBit  bool // byte_4FDD7
}

// 兩個載具的比對值(原版 `cmp eax, 12h` / `cmp eax, 14h`,前面套過 `and 0FEh`)。
const (
	mountedHorse = 0x12 // 騎著的馬(0x12/0x13)—— 不是地上那匹(0x10/0x11)
	ridingCarpet = 0x14 // 魔毯(0x14/0x15)
)

// extraWorldTurn 跑一個世界回合,回報有沒有出事(遭遇 / 進戰鬥)。
//
// ⚠ 這裡**不推時鐘** —— 時間由呼叫端加(原版 `sub_2E24` 只動世界,
// `sub_29304` 才動時鐘)。混在一起會多算分鐘。
//
// ⚠⚠ **本體還沒實作。** `sub_2E24` 過了閘門之後做四件事:
//
//	threshold = sub_1F98()                        ; 遭遇門檻
//	if (random(1, 30) < threshold) sub_2218()     ; 生怪
//	倒著掃槽 0x1F..1:sub_25F0(槽) 讓怪動 / 攻擊;沒出事再 sub_2D38(槽) 漂流
//	再掃一遍:離視窗超過 0x1F 格的怪 → sub_2B6C8 清掉
//
// 目前這裡只有 `advanceNPCs`(而它**只在場景裡有作用**)⇒ **大地圖上怪物不會動、
// 不會擲遭遇、太遠的怪也不會被清掉**。`docs/re/38` §4 的對應表把
// `sub_2E24` 標成已實作,那是過期斷言。這一段列為 `WORKLIST §5.1` 第 11 條,
// 五支函式(`sub_1F98`/`sub_2218`/`sub_25F0`/`sub_2D38`/`sub_2B6C8`)另案處理。
func (s *State) extraWorldTurn() bool {
	// ⚠ 三道閘門讀的都是原版的原始欄位:`byte_3E08A`(= `CombatMode`)與
	// `byte_3E08C`(= `Transport`)。載具用 `& 0xFE` 而不是 `VehicleKind`,
	// 因為原版比的是 `0x12`(騎著的馬)與 `0x14`(魔毯)這兩對值 ——
	// `VehicleKind` 會把「地上那匹馬」(0x10/0x11)也算進來。
	if s.CombatMode == CombatModeTimeStop {
		return false
	}
	if s.CombatMode == CombatModeSlow {
		s.gates.relTymBit = !s.gates.relTymBit
		if s.gates.relTymBit {
			return false
		}
	}
	if m := s.Transport & 0xFE; m == mountedHorse || m == ridingCarpet {
		s.gates.mountBit = !s.gates.mountBit
		if s.gates.mountBit {
			return false
		}
	}
	s.WorldTurns++
	before := s.InCombat()
	// ★ 遭遇擲骰:`門檻 = sub_1F98()`,`random(1, 30) < 門檻` 才生。
	// ⚠ 只在大地圖 —— 場景與地牢有自己的遊蕩怪路徑。
	if !s.InScene() && !s.InDungeon() && !s.InCombat() {
		threshold := u5data.EncounterThreshold(s.TileAt(s.X, s.Y), s.Clock.Hour, s.Floor != 0)
		if s.Roll(u5data.EncounterRollLo, u5data.EncounterRollHi) < threshold {
			s.spawnWanderingCreature()
		}
	}
	// ★ 倒著掃槽 0x1F..1(槽 0 是隊伍),讓每一個「算生物」的槽對隊伍動手。
	// ⚠ 判準用 `u5data.IsCreatureTile`(依 `ObjKind` 位元組)**不是**
	// `MapObject.IsCreature()` —— 兩者的範圍不同,分歧點在 0x40(`docs/re/82` §3)。
	//
	// ⬜ 原版在這裡還有 `if (esi == 0) sub_2D38(槽)`(能不能動的閘門)與
	// 第二趟的清場,兩者都還沒做(`docs/re/83` §2、`WORKLIST §5.1b`)。
	if set := s.currentObjects(); set != nil && !s.InScene() && !s.InDungeon() {
		acted := false
		for slot := len(set.Objects) - 1; slot >= 1; slot-- {
			if !u5data.IsCreatureTile(set.Objects[slot].Raw[u5data.ObjKind]) {
				continue
			}
			if s.objectAttacks(slot) {
				if s.InCombat() {
					// 進了戰鬥就停下 —— 後面的槽這一回合不再動手。
					break
				}
				// ★ 出事的那一槽**不再移動**(原版 `if (esi == 0) sub_2D38(槽)`)。
				// ⚠ 而且 `esi` 是**累加**的:一旦有任何一槽出事,
				// 後面的槽也全部不再移動。照原樣。
				acted = true
			}
			if acted {
				continue
			}
			s.objectMoveGate(slot)
		}
	}
	s.advanceNPCs()
	s.syncNPCObjects()
	return s.InCombat() != before
}
