package game

import "github.com/wicanr2/u5-cht/internal/u5data"

// 船的損傷、沉船、棄船,以及「轉向要花一回合」
//
// 規則與組語全部寫在 `u5data/shipdamage.go`。這一整組是**因為 Hex-Rays
// 安靜截斷而漏掉的**(`docs/re/66`)—— 在此之前引擎裡的船永遠不會沉。

// DamageShip 跑一次船身損傷判定(原版 `sub_22F0`)。
//
// 回傳 true 代表船沉了。**只有大船會受損** —— 小艇與魔毯進來是空轉,
// 那是刻意的:原版六個觸發點裡有兩個(Rough seas / 觸礁)在小艇與魔毯上
// 也會呼叫它,靠這一行擋掉。
func (s *State) DamageShip() bool {
	if !u5data.ShipTakesDamage(s.Transport) {
		return false
	}
	dmg := s.Roll(1, u5data.ShipDamageMax)
	if dmg < s.ShipHull {
		s.ShipHull -= dmg
		return false // 撐住了
	}
	s.ShipHull = 0
	s.Log(MsgShipSunk)
	s.abandonShip()
	return true
}

// abandonShip 是沉船之後的三層階梯:小艇 → 魔毯 → 溺水。
func (s *State) abandonShip() {
	facing := s.Transport & 0x03
	switch {
	case s.ShipSkiffs > 0:
		s.Log(MsgAbandonShip)
		// ⚠ 原版**不扣小艇數**(組語裡沒有 `dec`)—— 照原樣。
		s.Transport = u5data.VehicleSkiff | facing
	case s.Inventory.Carpets > 0:
		s.Log(MsgAbandonShip)
		s.Inventory.Carpets--
		// 魔毯只有東西兩個朝向,原版是 `sub_28E14(0, 1)` 隨機挑一個。
		s.Transport = u5data.VehicleCarpet + byte(s.Roll(0, 1))
	default:
		s.drown()
	}
	s.ShipHull, s.ShipSkiffs = 0, 0
}

// drown 是連小艇與魔毯都沒有的下場。
//
// ⚠ 原版把載具碼設成 **0** —— 那不是「步行」(0x1C),是「不畫載具圖」
// (同樣的手法在 `sub_10A1C` 的墜落動畫裡出現過:先存起舊值、設 0、重畫、再還原,
// 而這裡**沒有還原**)。所以隊伍的圖示會消失在水裡。
func (s *State) drown() {
	s.Transport = 0
	s.Log(MsgDrowning)
	// 閘門是 `sub_2B67C() != -1`:回 −1 代表沒人能行動、也沒人睡著,
	// 也就是隊伍已經全滅 —— 那時不必再灑傷害。
	if s.anyoneCanAct() {
		s.damageWholeParty()
	}
}

// anyoneCanAct 重現 `sub_2B67C` 的三態回傳,只取「不是 −1」這一面。
//
// 原版掃隊員:'G'(良好)或 'P'(中毒)→ 能行動;'S'(睡著)→ 記一筆;
// 其他('D' 死 / 'C' 魅惑…)→ 跳出迴圈。掃完沒人能行動時,有人睡著回 1、
// 否則回 −1。
func (s *State) anyoneCanAct() bool {
	asleep := false
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		switch s.Roster[i].Status {
		case u5data.StatusGood, u5data.StatusPoisoned:
			return true
		case u5data.StatusAsleep:
			asleep = true
		default:
			return asleep
		}
	}
	return asleep
}

// RoughSeas 是小艇或魔毯走到那個水 tile 上的遭遇(原版 `sub_2D9D0`)。
//
// ⚠ 原版接著呼叫 `sub_22F0`,但那一支只動大船 —— 所以對小艇與魔毯來說
// **只有訊息與動畫**,沒有實際損傷。這裡照做(`DamageShip` 自己會擋),
// 不要「順手」補一個傷害進去。
func (s *State) RoughSeas() {
	if !u5data.RoughSeasAffects(s.Transport) {
		return
	}
	s.Log(MsgRoughSeas)
	s.DamageShip()
}

// RoughSeasHere 是「移動之後」該不該遇到風浪(原版 `sub_2D9D0` 的
// `cmp esi, 1` 那一段,`esi` 是腳下的 tile)。
//
// ⚠⚠ **這一支原本沒有任何呼叫者。** `RoughSeas` 上一輪就寫好了、也有測試,
// 但**沒有人叫它** ⇒ 遊戲裡的風浪永遠不會發生。這與卷軸 / 藥水 / 月石 / 碎片
// 是同一類:**做完了卻接不到,等於沒做**(`docs/re/80`)。
//
// 觸發條件兩個都要成立:腳下是**深水**(tile 1),而且載具是小艇或魔毯。
func (s *State) RoughSeasHere() {
	if s.InScene() || s.InDungeon() || s.InCombat() {
		return
	}
	if s.TileAt(s.X, s.Y) != u5data.RoughSeasTile {
		return
	}
	s.RoughSeas()
}

// turnShipInstead 是大船的轉向(原版 `sub_2CCFC` 的大船分支)。
//
// 回傳 true 代表**這一步被用掉了**,呼叫端不要移動。三種情況:
//
//	想去的方向 != 目前朝向 → 轉向 + 印 "Head <方向>",耐久 < 50 再加一句 "Hull weak!"
//	揚著帆(0x20..0x23)而且無風 → 動不了
//	其餘 → false,照走
func (s *State) turnShipInstead(d Direction) bool {
	if u5data.VehicleKind(s.Transport) != u5data.VehicleSailing &&
		u5data.VehicleKind(s.Transport) != u5data.VehicleShip {
		return false
	}
	want := (s.Transport & 0xFC) | byte(d)
	if want != s.Transport {
		s.Transport = want
		s.Log(MsgHeading + d.Name())
		if s.ShipHull < u5data.ShipHullWeak {
			s.Log(MsgHullWeak)
		}
		return true
	}
	// 已經朝著那個方向。收帆的船(0x24..0x27)不受風影響。
	if s.Transport >= u5data.VehicleShip {
		return false
	}
	// ★ 揚著帆 + 無風 = 動不了。原版 `cmp byte_3E0A2, 0; jnz →照走`。
	return s.Wind == u5data.WindCalm
}

// 撞上東西:`Blocked!` / `COLLISION!` / `Docked!` / `OUCH!`(原版 `sub_2CE70`)
//
// 這一支是**通行判定 + 撞擊處理**合在一起(回傳「過不過得去」),
// 所以它同時是 `sub_22F0` 的第二個觸發點 —— `docs/re/66` 當時只知道
// 「有這條路,觸發條件沒讀」。讀完之後四個地形值全部對得上語意:
//
//	tile 3    淺灘   → "BREAKING UP!"   ★ 揚著帆衝進淺灘,船身裂開
//	tile 0x47 碼頭   → "Docked!"        ★ 靠岸:載具碼 += 4(揚帆組 → 收帆組)
//	tile 0x2F 仙人掌 → "OUCH!" + 全隊受傷
//	其餘             → "COLLISION!"(揚帆)或 "Blocked!"(沒揚帆)
//
// 四個都用 `internal/i18n/look_text.go` 的 `look#<tile>` 查得出中文名,
// 而且每一個都與訊息**語意相符** —— 這比逐個位址對照更有說服力。
// 順帶把 `RoughSeasTile = 1` 也認出來了:`look#1` 是**深水**,
// 所以小艇與魔毯是在深水上遇到風浪。
//
// 原版的順序有一個容易寫錯的地方:
//
//	if (地形 == 3)       印 "BREAKING UP!"    → 接著仍走「受損」那條
//	else if (地形 != 47h) 印 "COLLISION!"     → 接著也走「受損」
//	if (地形 == 47h) { 印 "Docked!"; 載具 += 4; 收帆 }
//	else             { 音效; sub_22F0(); 收帆 }
//
// 也就是**淺灘會印兩件事**(裂開 + 受損),而碼頭**一句都不印撞擊**。

// blockedMove 是走不過去時的處理(原版 `sub_2CE70` 的後半)。
func (s *State) blockedMove(nx, ny int) {
	tile := s.TileAt(nx, ny)
	if u5data.VehicleKind(s.Transport) != u5data.VehicleSailing {
		// 沒揚帆:一句 Blocked,而仙人掌會扎人。
		s.Log(MsgBlocked)
		if tile == u5data.TileCactus {
			s.Log(MsgOuch)
			s.damageWholeParty()
		}
		return
	}
	// 揚著帆撞上東西。
	switch {
	case tile == u5data.TileShallowWater:
		s.Log(MsgBreakingUp)
	case tile != u5data.TileDock:
		s.Log(MsgCollision)
	}
	if tile == u5data.TileDock {
		s.Log(MsgDocked)
		// ★ 載具碼 += 4:揚帆那一組(0x20..0x23)回到收帆那一組(0x24..0x27),
		// 朝向自然保留 —— 原版就是一個 `add byte_3E08C, 4`。
		s.Transport += 4
		return
	}
	s.DamageShip()
	// 撞完自動收帆(原版 `mov byte_3E167, 0`)。
	s.Transport = u5data.VehicleShip | (s.Transport & 0x03)
}

// ShipHullShown 回報狀態列該不該畫船身耐久,以及要畫多少
// (原版 `sub_2A1E8` 的 `Ship:` 分支)。
//
// 兩個條件都要成立:**不在戰鬥中**(`byte_3E0A3 < 0x80`)而且
// **載具是揚著帆的大船**(`byte_3E08C & 0xF8 == 0x20`)。
//
// ⚠ 遮罩是 **0xF8**,涵蓋 0x20..0x27 —— 揚帆組(0x20..0x23)**與**收帆組
// (0x24..0x27)都算。所以收了帆照樣看得到耐久,只有下了船才看不到。
func (s *State) ShipHullShown() (int, bool) {
	if s.InCombat() {
		return 0, false
	}
	if !u5data.ShipTakesDamage(s.Transport) {
		return 0, false
	}
	// ⚠ 原版讀的是 `dword_3E470+1` = **物件槽 0(隊伍自己)的 +5** ——
	// 上了船之後那一槽就是這艘船。引擎把「正在騎的船」的耐久放在
	// `State.ShipHull`(`DamageShip` 也是改那一個),兩者是同一件事的
	// 兩種存法;這裡跟著引擎的模型,不去讀物件槽,免得兩個來源分歧。
	return s.ShipHull, true
}
