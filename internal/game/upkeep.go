package game

import "github.com/wicanr2/u5-cht/internal/u5data"

// 每回合的維生開銷:中毒扣血、三餐扣糧、沒糧就餓(原版 `sub_2A50C`)
//
// 這一支也在 `docs/re/66` 的截斷清單上(113 行 → 4 行 C、`'Starving!'` 掉了),
// 而引擎在此之前:**中毒不會扣血、隊伍永遠不會餓**。
// 存糧只有「怪物偷糧」那一處會減 —— 買糧買了也用不掉。
//
// # `sub_2A50C`
//
//	for (每個隊員 i) {
//	    狀態 = byte_3DDBF[i*32]
//	    if (狀態 == 'D' && byte_3E08B == i) byte_3E08B = 0FFh   ; 死了就取消選中
//	    if (狀態 == 'D' || 狀態 == 'S') continue
//	    if (狀態 == 'P') sub_2A464(i, 1)                        ; ★ 中毒每回合 1 血
//	    活人數++                                                ; ★ 睡著的不算
//	}
//	if (byte_3E08F == byte_3E090) return                        ; 這個小時已經結算過
//	if (word_3DFB4 == 0) {                                      ; ★ 存糧 0
//	    印 "Starving!"
//	    sub_2A4D0()                                             ; 全隊 rand(1, 8) 傷
//	} else if (byte_3E08F ∈ {6, 12, 18}) {                      ; ★ 6 / 12 / 18 點
//	    存糧 -= 活人數
//	    byte_3E090 = byte_3E08F                                 ; 記下「這個小時結算過了」
//	    byte_3E09B += 1(上限 255)
//	}
//
// ★ 三個容易寫錯的地方:
//
//  1. **一天扣三次,不是每回合扣。** `byte_3E090` 記著上一次結算的小時,
//     同一個小時內走幾百步也只扣一次。
//  2. **睡著的人不吃**(`'S'` 那條 `continue` 在計數之前),死人也不吃。
//  3. **餓的判定在扣糧之前,而且是每小時一次** —— `loc_2A5C0`(「記下這個
//     小時處理過了」)是三條路的匯流點,挨餓那條也 `jmp` 到它。
//     所以存糧 0 的隊伍每小時掉一次血,不是每走一步掉一次。
//
// ✅ `byte_3E09B` 的語意已定(`docs/re/99` §5b):**每回合 +1**(上限 255),
// 與吃飯無關 —— 它是施捨業報的節流計數器(見 `ringtick.go`)。

// 維生開銷的常數。
const (
	// PoisonDamagePerTurn 是中毒每回合扣的血(原版 `sub_2A464(i, 1)`)。
	PoisonDamagePerTurn = 1
)

// MealHours 是一天扣糧的三個時刻(原版 `cmp al, 6 / 0Ch / 12h`)。
var MealHours = [3]int{6, 12, 18}

// newUpkeepHour 是 mealHour 的初值:−1 代表「還沒結算過任何一個小時」。
const newUpkeepHour = -1

// upkeep 是每回合的維生開銷。
func (s *State) upkeep() {
	if s.mealHour == 0 && s.Clock.Hour != 0 {
		// 零值修正:第一次進來就把「還沒結算過」記成 −1,
		// 免得午夜那一小時被零值誤判成「剛結算過」。
		s.mealHour = newUpkeepHour
	}
	alive := 0
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		switch s.Roster[i].Status {
		case u5data.StatusDead, u5data.StatusAsleep:
			// ★ 死人與睡著的人都不算「要吃飯的人」。
			continue
		case u5data.StatusPoisoned:
			s.damageMember(i, PoisonDamagePerTurn)
		}
		alive++
	}
	s.settleHour(alive)
	// ★★ 尾段:**三條路都會走到這裡**(原版 `loc_2A5CA`,含「這小時已結算過」
	// 那條)。此前整段都沒做,而漏掉的理由是 `call sub_2BCC8` 在 `retn`
	// 前面一行 —— 讀函式讀到「看起來收尾了」就停(見 `ringtick.go` 檔頭)。
	s.countTurn()
	s.tickModeOutOfCombat()
	s.regenerateParty()
}

// settleHour 是維生開銷的「每小時結算一次」那一段(原版 `loc_2A563`..`loc_2A5C0`)。
//
//	if (這個小時已經結算過) return
//	if (存糧 == 0)  { 印 "Starving!"; 全隊 rand(1,8) }
//	else if (是用餐時刻) 存糧 −= 活人數
//	記下「這個小時結算過了」          ; ★ 三條路都設,不只用餐那條
//
// ⚠⚠ **最後那一行此前只寫在用餐那條路上**,而原版的 `loc_2A5C0` 是
// 三條路的匯流點(`jmp short loc_2A5C0` 從挨餓與「不是用餐時刻」兩處跳來)。
// 差別很具體:**挨餓的傷害是每小時一次,不是每回合一次**。
// 引擎此前每走一步就掉一次血,斷糧會在幾十步內滅團 —— 那不是原版的難度。
func (s *State) settleHour(alive int) {
	// ⚠ mealHour 的零值 0 會與「午夜 0 點」相等,所以初值用 −1(見 `newUpkeepHour`)。
	if s.mealHour == s.Clock.Hour {
		return
	}
	switch {
	case s.Inventory.Food == 0:
		s.Log(MsgStarving)
		s.damageWholeParty()
	case isMealHour(s.Clock.Hour):
		s.Inventory.Food -= alive
		if s.Inventory.Food < 0 {
			s.Inventory.Food = 0
		}
	}
	s.mealHour = s.Clock.Hour
}

// isMealHour 回報現在是不是三個用餐時刻之一。
func isMealHour(hour int) bool {
	for _, h := range MealHours {
		if hour == h {
			return true
		}
	}
	return false
}

// tickModeOutOfCombat 是維生開銷尾段的模式倒數(原版 `sub_2A50C` 的 `byte_3E09E`)。
//
//	if (byte_3E09E != 0 && != 0FFh) {
//	    if (--byte_3E09E == 0) { byte_3E08A = 0; 重畫面板 }
//	}
//
// ★ 與戰鬥裡那一份(`tickCombatMode`,原版 `sub_16370`)是**同一個計數器
// 在兩個迴圈裡各減一次** —— 而戰鬥與非戰鬥是互斥的(`docs/re/81`),
// 所以實際上每回合只減一次。⚠ `0xFF` 是「不倒數」的哨兵。
func (s *State) tickModeOutOfCombat() {
	if s.CombatModeTurns <= 0 || s.CombatModeTurns == ModeTurnsForever {
		return
	}
	s.CombatModeTurns--
	if s.CombatModeTurns == 0 {
		s.CombatMode = CombatModeNone
	}
}

// ModeTurnsForever 是「這個模式不倒數」的哨兵(原版 `cmp al, 0FFh; jz`)。
const ModeTurnsForever = 0xFF

// 開戰時戒指會消失(原版 `sub_2EE84` 佈陣那一段)
//
// 佈陣的迴圈裡夾了一段與佈陣無關的判定:
//
//	if (byte_3DDD1[i*32] == 2Ah) 戒指 = 2Ah        ; 隱形戒指
//	if (byte_3DDD1[i*32] == 2Ch) 戒指 = 2Ch        ; 再生戒指
//	if (戒指 != 0 && rand(0, 15) == 11) {          ; ★ 1/16
//	    印 "A ring has vanished!"
//	    sub_2F35C(i, 戒指)                         ; 把它從身上拿掉
//	}
//
// `byte_3DDD1` − 名冊基底 `byte_3DDB4` = **0x1D** = `CharRing` ✓
//
// ★ **只有兩種戒指會消失**:
//
//	0x2A = 42 Ring of Invisibility 隱形戒指   → 會
//	0x2B = 43 Ring of Protection   防護戒指   → **不會**
//	0x2C = 44 Ring of Regeneration 再生戒指   → 會
//
// 名字是從 `DATA.OVL` 0x1806 的指標表 + 0x10 偏移逐筆解出來的(`items.go`)。
// 兩個會消失的正好是「效果最強」的兩個 —— 防護戒指沒事,這不是隨便挑的。
//
// ⚠ 判定是 `rand(0, 15) == 11`,不是 `< 1` 也不是 `== 0` ——
// 機率一樣是 1/16,但**寫成別的等號就不是同一組隨機序列**了。照原樣。

// RingVanishRollMax / RingVanishHit 是戒指消失的判定(原版 `rand(0, 15) == 11`)。
const (
	RingVanishRollMax = 15
	RingVanishHit     = 11
)

// RingsThatVanish 是會在開戰時消失的兩種戒指。
//
// ⚠ 防護戒指(43)**不在**這裡 —— 原版只比 0x2A 與 0x2C 兩個值。
var RingsThatVanish = [2]byte{u5data.ItemRingInvisibility, u5data.ItemRingRegeneration}

// vanishRings 在開戰時擲一次戒指消失(原版 `sub_2EE84` 的那一段)。
func (s *State) vanishRings() {
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		ch := &s.Roster[i]
		if ch.Status == u5data.StatusDead {
			continue
		}
		ring := ch.Raw[u5data.CharRing]
		if ring != RingsThatVanish[0] && ring != RingsThatVanish[1] {
			continue
		}
		if s.Roll(0, RingVanishRollMax) != RingVanishHit {
			continue
		}
		ch.Raw[u5data.CharRing] = u5data.ItemNone
		s.Log(MsgRingVanished)
	}
}

// 腳下那一格每回合的作用:活門、沼澤、火(原版 `sub_1318`)
//
// `sub_1318` 也在截斷清單上(197 行 → 29 行 C,三個字串全掉),
// 而它同時是**維生開銷的呼叫端** —— 最後一行是 `if (ebx == 0) sub_2A50C()`。
// 也就是說:**掉下活門的那一回合不算吃飯**。
//
//	for (每個隊員)
//	    if (狀態 == 'S' && rand(0, 15) == 15) 狀態 = 'G'    ; ★ 睡著的 1/16 自己醒
//	tile = 腳下那一格
//	if (tile == 8Ch && (載具 & 0FEh) != 14h) {              ; ★ 鬆動的磚,而且不在魔毯上
//	    印 "A TRAPDOOR!"
//	    sub_2A4D0()                                         ; 全隊 rand(1, 8) 傷
//	    if (地點 == 1Dh) {                                  ; ★★ STONEGATE
//	        …把地圖緩衝填滿 0x8F(熔岩)、清空物件表…
//	        每個隊員 血 = 0、狀態 = 'D'                      ; ★★ 全隊當場死亡
//	    }
//	    樓層--;  ebx = 1                                    ; 兩條路都掉下一層
//	}
//	else if (tile == 4 && 載具 == 1Ch) {                     ; ★ 沼澤,而且**步行**
//	    for (每個隊員)
//	        if (狀態 == 'D' || 狀態 == 'P') continue
//	        if (rand(0, 29) > 敏捷) { 印 "Poisoned!"; 狀態 = 'P' }
//	}
//	else if (tile == 0BCh || tile == 8Fh) {                  ; ★ 壁爐 / 熔岩
//	    印 "Burning!"
//	    sub_2A4D0()
//	}
//	if (ebx == 0) sub_2A50C()
//
// 四個 tile 都用 `look#<tile>` 查得出名字,四個都與效果相符:
//
//	tile 4    沼澤       → 中毒
//	tile 0x8C 鬆動的磚   → 活門
//	tile 0x8F 熔岩       → 燃燒
//	tile 0xBC 壁爐       → 燃燒
//
// ★★ **地點 29 是 STONEGATE**(`locations.go` 的第 29 筆)。踩到那裡的鬆動磚,
// 原版把整張地圖緩衝填滿熔岩 tile 再讓全隊死亡 —— 那不是普通陷阱,是死路。
// 這一條**不能「順手加個存活判定」**:原版就是直接把血設 0、狀態設 'D'。
//
// ⚠ 沼澤中毒判的是 `rand(0, 29) > 敏捷`(**大於**才中毒)——
// 所以敏捷 30 以上完全不會中毒。而**只有步行會中**:騎馬、坐船、坐魔毯都免疫,
// 而且載具判的又是**單一值 0x1C**(與食人妖、大地圖攀爬同一個寫法)。

// 地形效果的常數。
const (
	// TileTrapdoor 是鬆動的磚(`look#140`)—— 踩到就掉下一層。
	TileTrapdoor = 0x8C
	// TileSwamp 是沼澤(`look#4`)—— 步行走過去會中毒。
	TileSwamp = 0x04
	// TileLava / TileFireplace 是熔岩與壁爐(`look#143` / `look#188`)。
	TileLava      = 0x8F
	TileFireplace = 0xBC
	// SwampPoisonRollMax 是沼澤中毒的骰上限(`rand(0, 29)`,**大於**敏捷才中毒)。
	SwampPoisonRollMax = 29
	// WakeUpRollMax / WakeUpHit 是睡著的人自己醒(`rand(0, 15) == 15`)。
	WakeUpRollMax = 15
	WakeUpHit     = 15
	// StonegateLocation 是 STONEGATE —— 那裡的活門會讓全隊死亡。
	StonegateLocation = 0x1D
)

// terrainEffects 是腳下那一格每回合的作用(原版 `sub_1318`)。
//
// ⚠⚠ **只在場景裡跑。** `sub_1318` 的唯一呼叫者是 `sub_1A54`(場景主迴圈),
// 而大地圖走的是 `sub_2D9D0`(見 `overworldturn.go`)、地牢走 `sub_5150`。
// 三個迴圈互斥(`docs/re/81` §1),所以這裡的「每回合」是「**場景裡的**每回合」。
// 由 `tick()` 的模式分流保證,不要從別的地方叫它。
func (s *State) terrainEffects() {
	// ★ 睡著的人每回合有 1/16 自己醒(戰鬥外的那一條,與戰鬥中的兩條都不同)。
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		if s.Roster[i].Status == u5data.StatusAsleep && s.Roll(0, WakeUpRollMax) == WakeUpHit {
			s.Roster[i].Status = u5data.StatusGood
		}
	}
	switch tile := s.TileAt(s.X, s.Y); {
	case tile == TileTrapdoor && s.Transport&0xFE != u5data.VehicleCarpet:
		s.Log(MsgTrapdoor)
		s.damageWholeParty()
		if s.Location == StonegateLocation {
			// ★★ STONEGATE 的活門下面是熔岩 —— 原版當場把全隊寫成死亡。
			for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
				s.Roster[i].HP = 0
				s.Roster[i].Status = u5data.StatusDead
			}
		}
		if s.InScene() {
			s.changeFloor(-1)
		}
		// ★ 掉下去的這一回合不算吃飯(原版 `if (ebx == 0) sub_2A50C()`)。
		return
	case tile == TileSwamp && s.Transport == u5data.VehicleWalk:
		// ⚠ **場景裡**的沼澤擲 0..29;大地圖那顆是 1..30
		// (`SwampOverworldPoison*`,在 `overworldturn.go`)。
		// 兩顆都是原版的,區分的是**地點**,別合併也別互相套。
		s.poisonPartyBySwamp(0, SwampPoisonRollMax)
	case tile == TileLava || tile == TileFireplace:
		s.Log(MsgBurning)
		s.damageWholeParty()
	}
	s.upkeep()
}

// 喝醉(原版 `sub_1158` —— 讀鍵的那一層)
//
// `sub_1158` 是**讀一個鍵**的包裝函式,而它在讀完之後夾了一段:
//
//	if (byte_3E169 != 0 && rand(0, 1) != 0) {      ; ★ 醉酒計數 > 0,而且擲中 1/2
//	    sub_C10()
//	    byte_3E169--                               ; ★ 每次踉蹺就少一次
//	    印 "Hic!"
//	    return byte_4FC54[rand(0, 3)]              ; ★★ 回傳一個**亂的方向鍵**
//	}
//	return 真正讀到的鍵
//
// `byte_4FC54` 的前四個位元組是 `3, 4, 2, 1` —— 正是四個方向鍵的鍵碼
// (與 `sub_1EFC8` 的清單瀏覽器用的 1..4 同一組)。
//
// # 誰讓你醉的
//
//	sub_21108  mov byte_3E169, 19h    ← ★ **酒館的酒單**,一杯醉 25 次
//	sub_1678   mov byte_3E169, 0      ← 進場景時清掉
//
// ⇒ **在酒館點一杯酒,接下來有 25 次「按鍵變成隨機走一步」**,
// 而且是「有一半機率發生,發生了才扣一次」—— 所以實際會持續相當久。
//
// ★ 這是**輸入層**的效果,不是狀態欄上的一個圖示:玩家按 Z 想看數值,
// 結果往東走了一格。少了它,酒館的酒就只是「花錢買一句話」。

// 醉酒的常數。
const (
	// TavernDrunkTurns 是點一杯酒醉多久(原版 `sub_21108` 的 `mov byte_3E169, 19h`)。
	TavernDrunkTurns = 0x19
	// DrunkStaggerOdds 是每次按鍵踉蹺的機率分母(原版 `rand(0, 1) != 0` → 1/2)。
	DrunkStaggerOdds = 2
)

// DrunkKeys 是踉蹺時會走的四個方向(原版 `byte_4FC54` 的前四個位元組 3,4,2,1)。
//
// ⚠ 順序照原樣 —— 四個都在裡面,機率相同,但**表的順序就是原版的順序**。
var DrunkKeys = [4]Direction{West, East, South, North}

// GetDrunk 讓隊伍醉 25 次(酒館點酒時呼叫)。
func (s *State) GetDrunk() { s.Drunk = TavernDrunkTurns }

// DrunkStagger 回報這一次按鍵是不是變成了隨機走一步。
//
// 回傳 (方向, true) 代表踉蹺了 —— 呼叫端應該**丟掉原本的按鍵**、改成往那個方向走。
func (s *State) DrunkStagger() (Direction, bool) {
	if s.Drunk <= 0 {
		return North, false
	}
	if s.Roll(0, DrunkStaggerOdds-1) == 0 {
		return North, false
	}
	s.Drunk--
	s.Log(MsgHic)
	return DrunkKeys[s.Roll(0, len(DrunkKeys)-1)], true
}

// 踏進沼澤的那一次中毒(原版 `sub_10BDC`,由移動後的分派 `sub_2D9D0` 呼叫)
//
// ★★ **沼澤有兩個中毒判定,不是一個。**
//
//	sub_2D9D0(移動之後)→ sub_10BDC     擲 random(1, 30),roll > 敏捷 → 中毒
//	sub_1318 (每回合)  → 自己那一段    擲 random(0, 29),roll > 敏捷 → 中毒
//
// 兩者的條件完全相同(腳下 tile == 4、`byte_3E08C == 0x1C` 步行、
// 跳過 'D' 與 'P'、逐個隊員各擲一次、印 `Poisoned!`),**只有骰子的範圍差一格**。
//
// ⇒ **踏進沼澤的那一步會被擲兩次**(先 `sub_10BDC`,同一回合再 `sub_1318`),
// 之後每站一回合擲一次。所以「走進去」比「站著」危險一倍。
//
// ⚠ 兩顆骰子的範圍不同幾乎可以確定是原作者的手誤(1..30 vs 0..29)——
// 但那個差異**改變機率**:敏捷 29 的人只有 `sub_10BDC` 那次毒得到
// (`sub_1318` 最大只擲到 29,而 29 > 29 不成立);敏捷 30 兩邊全免疫。
// 照原樣實作,不「統一」成同一顆(`CLAUDE.md` §3.0:不自行平衡)。
//
// ⚠⚠ 這一支是**先有 `sub_1318` 那一半、後來才發現另一半**的:
// `docs/re/70` 寫沼澤中毒時只讀了 `sub_1318`,而 `sub_10BDC` 在
// `docs/re/66` 的截斷清單上獨立列著,兩者當時沒被聯想到一起。
// **同一個機制被兩支函式各做一半,是這個執行檔的常見形狀**(見 `docs/re/72`
// 的兩顆戒指骰子)—— 找到一處之後要問「還有誰做同一件事」。

// SwampOverworldPoisonLo / Hi 是**大地圖**沼澤的骰範圍(原版 `sub_10BDC` 的
// `sub_28E14(1, 1Eh)`)。場景那顆是 `random(0, SwampPoisonRollMax)`。
//
// ⚠⚠ 這一格之差**改變機率**,不能統一:
//
//	敏捷 29 → 場景裡永遠免疫(29 > 29 不成立)、大地圖有 1/30 會中
//	敏捷 30 → 兩邊都免疫
//	敏捷 0  → 場景 29/30、大地圖 30/30 必中
//
// 幾乎確定是原作者手誤,但 `CLAUDE.md §3.0` 說得很清楚:**不自行平衡**。
// 兩顆照原樣留著,`TestSwampDiceDifferByPlace` 用敏捷 29 把差異釘住
// (那個值是唯一能區分兩顆骰子的敏捷)。
const (
	SwampOverworldPoisonLo = 1
	SwampOverworldPoisonHi = 30
)

// poisonPartyBySwamp 是兩支共用的本體:逐個隊員擲一次,擲贏敏捷就中毒。
//
// 跳過死人與已經中毒的人(原版兩支都是 `cmp dl, 'D'` / `cmp dl, 'P'`)。
func (s *State) poisonPartyBySwamp(lo, hi int) {
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		st := s.Roster[i].Status
		if st == u5data.StatusDead || st == u5data.StatusPoisoned {
			continue
		}
		// ⚠ **大於**敏捷才中毒 —— 兩支原版函式都是 `jle → 跳過`。
		if s.Roll(lo, hi) > int(s.Roster[i].Dex) {
			s.Log(s.Roster[i].Name + MsgPoisonedBySwamp)
			s.Roster[i].Status = u5data.StatusPoisoned
		}
	}
}
