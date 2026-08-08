package game

import "github.com/wicanr2/u5-cht/internal/u5data"

// 大地圖每回合的收尾(原版 `sub_2D9D0` 的 `loc_2DBF4` .. `loc_2DD34`)
//
// # 為什麼要有這一支
//
// `sub_0` 不是主迴圈,是**三個模式主迴圈之間的切換器**(`docs/re/81`):
//
//	大地圖   sub_2DD44 → sub_2D9D0
//	場景     sub_1A54  → sub_1318
//	地牢     sub_5378  → sub_5150
//
// 三支都「一路跑到玩家離開這個模式才回來」⇒ **同一回合絕不會走兩支**。
// 引擎原本用單一 `tick()` 服務所有模式,於是大地圖同時吃到場景那一套
// (`terrainEffects`)與自己那一套,而它獨有的那幾件事一件都沒做。
//
// # 原版的順序(不要重排)
//
//	if (edi == 0) 什麼都不做          ; 這個指令沒有用掉回合
//	sub_29304(2)                      ; ★ 大地圖是 2 分鐘,場景與地牢是 1
//	esi = tile(X, Y)                  ; ★ 只讀一次,後面全部用這個值
//	if ((esi & 0xFE) == 0x6A) sub_3010()          ; 橋 → 橋下的食人妖 ┐
//	else if (esi == 4 && 步行)  sub_10BDC()       ; 沼澤 → 中毒       │ 四條
//	                            sub_2B740(1)     ;   + 重畫一次      │ 互斥
//	else if (esi == 0x8F)      sub_10BC4()        ; 熔岩 → 燒        │
//	else                       聖域關卡           ; (0xE9, 0xEB)     ┘
//	sub_2D998()                       ; 幽冥界地震(1/256)
//	sub_2A50C()                       ; 維生開銷
//	if (esi == 1 && 小艇或魔毯) 風浪
//	if ((esi & 0xFC) == 0xD4)   瀑布
//	sub_2E24()                        ; ★★ 無條件的一個世界回合
//
// ★★ 最後那一行是**無條件**的。引擎的 `extraWorldTurn()` 原本只從
// `payTerrainCost` 呼叫(地形 1/2 級),所以**在草地上走路野怪不會動、
// 也不擲遭遇** —— 而草地是最常見的地形。地形代價的 1–2 個回合是**額外加**的
// (`docs/re/38` 的「山裡走一步怪走三步」= 1 個基本 + 2 個額外)。
//
// ⚠ tile 只讀一次這件事有觀察後果:橋下的食人妖把隊伍拉進戰鬥之後,
// 風浪與瀑布的判定用的還是**進戰鬥前**那一格。照原樣做。

// 大地圖收尾用到的常數。
const (
	// OverworldMinutesPerTurn 是大地圖一回合的分鐘數(原版 `sub_29304(2)`)。
	// ⚠ 與場景 / 地牢的 `MinutesPerTurn`(1)不同 —— 三個模式各自寫死。
	OverworldMinutesPerTurn = 2

	// EarthquakeRollMax / EarthquakeHit 是幽冥界地震(`sub_2D998`):
	// `random(0, 0xFF) == 0x69` ⇒ 1/256,而且必須**恰好等於** 0x69,
	// 不是「小於某個門檻」。
	EarthquakeRollMax = 0xFF
	EarthquakeHit     = 0x69

	// SacredGateX / SacredGateY 是地表大地圖上那一格聖域關卡
	// (`sub_2D9D0` 的 `loc_2DC5F`:`byte_3E0A6 == 0E9h && byte_3E0A7 == 0EBh`)。
	SacredGateX = 0xE9
	SacredGateY = 0xEB
)

// overworldTurnEnd 是大地圖用掉一回合之後的全部結算。
//
// 呼叫點只有 `tick()`,而 `tick()` 只在「這一步算用掉回合」時被呼叫 ——
// 對應原版的 `if (edi == 0) 跳過`。
func (s *State) overworldTurnEnd() {
	// ★ tile 只讀一次(原版 `movzx esi, byte ptr [eax]`),後面四條分支與
	// 風浪、瀑布都用這個值。中途進了戰鬥也不重讀。
	tile := s.TileAt(s.X, s.Y)

	switch {
	case tile&0xFE == TrollBridgeTile:
		// 走上橋有 1/8 遇到橋下的食人妖。
		s.crossBridge()
	case tile == TileSwamp && s.Transport == u5data.VehicleWalk:
		// ⚠ 大地圖的沼澤骰子是 **random(1, 30)**,場景那顆是 random(0, 29)。
		// 兩顆區分的是**地點**不是「走進去 vs 站著」—— `docs/re/74` §1 曾寫成
		// 後者,由 `docs/re/81` §3 更正。同一回合只擲一次。
		s.poisonPartyBySwamp(SwampOverworldPoisonLo, SwampOverworldPoisonHi)
	case tile == TileLava:
		// 大地圖的熔岩走 `sub_10BC4` —— 與場景 `sub_1318` 那條同效果、
		// 不同函式。⚠ 場景那條還包含壁爐(0xBC),大地圖這條**只有熔岩**。
		s.Log(MsgBurning)
		s.damageWholeParty()
	default:
		s.sacredQuestGate()
	}

	s.underworldEarthquake()
	s.upkeep()
	// 小艇 / 魔毯走到深水上會遇到風浪。
	if tile == u5data.RoughSeasTile {
		s.RoughSeas()
	}
	// 踩到瀑布就掉下去(世界上只有一個瀑布通往幽冥界)。
	if tile&0xFC == FallTileGroup {
		s.fallDownTheWaterfall()
	}
	// ★★ 無條件的一個世界回合:怪物移動 + 遭遇擲骰。
	s.extraWorldTurn()
}

// underworldEarthquake 是幽冥界的地震(原版 `sub_2D998`)。
//
//	if (byte_3E0A5 == 0) return                  ; ★ 地表不震
//	if (random(0, 0FFh) != 69h) return           ; 1/256
//	印 "EARTHQUAKE!"
//	sub_2AC08()                                  ; 震動畫面 + 隨機音效
//	sub_2A4D0()                                  ; 全隊各 random(1, 8) 傷
//
// `byte_3E0A5` 在大地圖上是 0(地表)或 −1(幽冥界),就是引擎的 `Floor`。
//
// ⬜ `sub_2AC08` 是純視覺:把畫面區塊上下捲動、每三列配一次隨機音效
// (`sub_28E14(13h, 96h)`)。引擎沒有阻塞動畫層,所以只有訊息與傷害。
func (s *State) underworldEarthquake() {
	if s.Floor == 0 {
		return
	}
	if s.Roll(0, EarthquakeRollMax) != EarthquakeHit {
		return
	}
	s.Log(MsgEarthquake)
	s.damageWholeParty()
}

// sacredQuestGate 是地表大地圖 (0xE9, 0xEB) 那一格的關卡(原版 `loc_2DC5F`)。
//
//	if (X != 0E9h || Y != 0EBh || byte_3E0A5 != 0 || byte_3E0A3 != 0) return
//	印 "\n\""
//	if (byte_3E0DC != 0) 印 "Pass, Seeker!\"\n"
//	else {
//	    印 "Thou art not upon a Sacred Quest!\n"
//	    印 "Passage denied!\"\n"
//	    inc byte_3E0A7                           ; ★ 把玩家往南推回一格
//	}
//
// `byte_3E0DC` 是**進行中的聖壇試煉位元圖**(一德一位元,聖壇冥想過關時設,
// 領獎時清)—— 引擎的 `ShrineQuestActive`(`docs/re/25`、`27`)。
//
// ⇒ 通關條件不是「領過某個試煉」而是「**此刻手上有任何一項未完成的試煉**」。
// 八德全部領完獎之後這道關卡會重新關上,那是原版行為。
//
// ★ 被拒絕時原版是 `inc` 座標而不是「擋住不讓走」—— 玩家已經站上去了,
// 才被推回南邊一格。所以會看到自己動了又被推回來。
func (s *State) sacredQuestGate() {
	if s.X != SacredGateX || s.Y != SacredGateY || s.Floor != 0 {
		return
	}
	if s.ShrineQuestActive != 0 {
		s.Log(MsgPassSeeker)
		return
	}
	s.Log(MsgNotOnSacredQuest)
	s.Log(MsgPassageDenied)
	s.Y = WrapWorld(s.Y + 1)
}

// minutesPerTurn 是這一回合推進的分鐘數。
//
// 三個模式各自寫死(`docs/re/81` §2):大地圖 `sub_29304(2)`、
// 場景 `sub_1A54` 的 `sub_29304(1)`、地牢 `sub_5378` 的 `sub_29304(1)`。
func (s *State) minutesPerTurn() int {
	if !s.InScene() && !s.InDungeon() && !s.InCombat() {
		return OverworldMinutesPerTurn
	}
	return MinutesPerTurn
}
