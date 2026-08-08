package game

import "github.com/wicanr2/u5-cht/internal/u5data"

// 移動時的載具動詞與朝向(原版 `sub_7C0`)
//
// 每走一步,原版會先依載具印一個動詞:
//
//	載具 & 0xFC == 0x10  馬     → "Ride "
//	              0x14  魔毯   → "Fly "
//	              0x28  小艇   → "Row "
//	              0x20/0x24 船 → 不印動詞
//	其餘(步行)              → 不印
//
// ★ 而且同一支還**改載具碼的低位元** —— 那是精靈圖的朝向:
//
//	馬  :往東(dir 1)→ 0x12   往西(dir 3)→ 0x13
//	魔毯:往東        → 0x14   往西        → 0x15
//	船 / 小艇:`載具 = (載具 & 0xFC) + dir`(四個朝向都用)
//
// ⚠ 馬與魔毯**只在東西向換**(原版只比 dir 1 與 3),往南北走保持原樣;
// 船是四向全換。兩種規則不一樣,不能寫成一條。
//
// 這件事看起來只是「印一個字」,但朝向位元同時是 `isBroadside`(開砲判舷側)
// 與 `ModeOf`(通行判定)讀的東西 —— 不更新的話船的舷側永遠算錯。

// vehicleVerb 回傳這個載具走一步要印的動詞;空字串代表不印。
func vehicleVerb(transport byte) string {
	switch transport & 0xFC {
	case u5data.TileHorse:
		return "騎行 "
	case u5data.VehicleCarpet:
		return "飛行 "
	case u5data.VehicleSkiff:
		return "划行 "
	}
	return ""
}

// faceVehicle 依走的方向更新載具碼的朝向位元,並回傳要印的動詞。
func (s *State) faceVehicle(d Direction) string {
	verb := vehicleVerb(s.Transport)
	switch s.Transport & 0xFC {
	case u5data.TileHorse:
		// 只有東西向換圖(原版 `cmp edi, 1` / `cmp edi, 3`)。
		if d == East {
			s.Transport = u5data.TileHorse | 0x02
		} else if d == West {
			s.Transport = u5data.TileHorse | 0x03
		}
	case u5data.VehicleCarpet:
		if d == East {
			s.Transport = u5data.VehicleCarpet
		} else if d == West {
			s.Transport = u5data.VehicleCarpet | 0x01
		}
	case u5data.VehicleSailing, u5data.VehicleShip, u5data.VehicleSkiff:
		// 船與小艇是四向:低兩位元直接放方向碼。
		s.Transport = (s.Transport & 0xFC) | byte(d)
	}
	return verb
}

// Pass 是空白鍵(原版 `sub_2ACF4` case 32)。
//
// ★ 它**不只是跳過一回合**:在地表而且揚著帆時,空白鍵是**收帆**
//(原版印 "Sheets in irons!" 並把 `byte_3E167` 清掉)。
// 只有不在那個狀態下才印 "Pass"。
//
// 寫成「空白鍵就是跳過」的話,玩家在海上會找不到收帆的辦法 ——
// 而 Y(Yell)那條路是**放**帆,不是收。
func (s *State) Pass() {
	// 揚著帆的狀態就是載具碼 0x20..0x23(`VehicleSailing`)。
	if s.Location == 0 && !s.InCombat() &&
		u5data.VehicleKind(s.Transport) == u5data.VehicleSailing {
		// 收帆:載具碼從揚帆那一組回到大船那一組,朝向保留。
		s.Transport = u5data.VehicleShip | (s.Transport & 0x03)
		s.Log("收帆!")
		// 收帆與 Pass 在原版是同一個 case 的兩條印字分支,兩條都落到
		// 同一個收尾(`loc_2AE30`)⇒ 一樣用掉一回合、一樣結算。
		s.tick()
		return
	}
	s.Log("按兵不動。")
	// ★ 空白鍵一樣要走每回合收尾:原版 case 32 回到 `sub_2D9D0` / `sub_1A54`
	// 的尾段,所以按空白鍵照樣結算地形、維生開銷、世界回合。
	// ⚠ 引擎原本自己 `AdvanceTime` + `extraWorldTurn()`,**跳過了地形與維生開銷**
	// —— 站著不動不會餓、中毒不會痛(`docs/re/81` §5)。
	s.tick()
}
