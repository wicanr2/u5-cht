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
//  3. **餓的判定在扣糧之前**,而且**每回合都判** —— 存糧 0 的隊伍走一步掉一次血,
//     不是一天掉三次。所以斷糧是會死人的。
//
// ⚠ `byte_3E09B` 是另一個計數器(每次進餐 +1,上限 255),用途還沒追 ——
// 先不接,標在這裡。

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
	// ★ 同一個小時只結算一次(原版 `byte_3E090` 記著上次結算的小時)。
	//
	// ⚠ mealHour 的零值 0 會與「午夜 0 點」相等 —— 而午夜不是用餐時刻,
	// 唯一的差別是「存糧 0 又剛好在 0 點」那一小時不會挨餓。
	// 為了不留這個縫,初值用 −1(見 `newUpkeepHour`)。
	if s.mealHour == s.Clock.Hour {
		return
	}
	if s.Inventory.Food == 0 {
		// ⚠ 餓的判定**每回合都跑**,不是一天三次 —— 斷糧的隊伍走一步掉一次血。
		s.Log(MsgStarving)
		s.damageWholeParty()
		return
	}
	for _, h := range MealHours {
		if s.Clock.Hour != h {
			continue
		}
		s.Inventory.Food -= alive
		if s.Inventory.Food < 0 {
			s.Inventory.Food = 0
		}
		s.mealHour = s.Clock.Hour
		return
	}
}

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
		for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
			st := s.Roster[i].Status
			if st == u5data.StatusDead || st == u5data.StatusPoisoned {
				continue
			}
			// ⚠ **大於**敏捷才中毒 —— 敏捷 30 以上完全免疫。
			if s.Roll(0, SwampPoisonRollMax) > int(s.Roster[i].Dex) {
				s.Log(s.Roster[i].Name + MsgPoisonedBySwamp)
				s.Roster[i].Status = u5data.StatusPoisoned
			}
		}
	case tile == TileLava || tile == TileFireplace:
		s.Log(MsgBurning)
		s.damageWholeParty()
	}
	s.upkeep()
}
