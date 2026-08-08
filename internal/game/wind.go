package game

import (
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 風
//
// 風向影響兩件事:**船走得多快**,以及**海上的船會不會自己漂**。
// 兩者查的是同一張 4 朝向 × 5 風的延遲表(`docs/re/23`)。
//
// Rel Hur(改風向)是玩家唯一能主動改它的手段 —— 也是為什麼那個咒語的
// 可施法場合表**只放行野外**:風只在海上有意義(`docs/re/17`)。

// SetWind 換風向(原版 `sub_2A984`:設 `byte_3E0A2` 並把變化計時器歸零)。
func (s *State) SetWind(wind int) {
	if wind < 0 || wind >= u5data.WindCount {
		return
	}
	s.Wind = wind
	s.windTimer = 0
	s.Log("風轉成了" + u5data.WindNameZH[wind] + "。")
}

// ChangeWind 是 Rel Hur:把風轉到指定方向(原版 `sub_1CDA4` 的跳表)。
//
// 方向 → 風值的對應寫死在 `sub_1CDA4` 的 `jpt_1CDD0` 裡(方向是**鍵碼** 1..4,
// 而鍵碼 1=西 2=東 3=北 4=南 —— 由 `sub_2D174`/`sub_2CCFC` 定案):
//
//	北 → 風 1   南 → 風 2   東 → 風 3   西 → 風 4
//
// ⚠ 原本這裡寫「西 → 風 3、東 → 風 4」,那是照引擎自己抄反的常數寫的。
// 更正的三個來源見 `u5data/wind.go` 的說明與 `docs/re/84` §3。
func (s *State) ChangeWind(d Direction) bool {
	switch d {
	case North:
		s.SetWind(u5data.WindNorth)
	case South:
		s.SetWind(u5data.WindSouth)
	case East:
		s.SetWind(u5data.WindEast)
	default:
		s.SetWind(u5data.WindWest)
	}
	return true
}

// CanSail 回報這一步走不走得動(原版 `sub_2CCFC` 的最後五行)。
//
//	loc_2CE50:                       ; 想去的方向就是現在的船頭(不必轉向)
//	    cmp  byte_3E08C, 24h
//	    jnb  short 回0               ; ★ >= 0x24(收帆的大船)→ 照走
//	    cmp  byte_3E0A2, 0
//	    jnz  short 回0               ; ★ 有風 → 照走
//	    mov  ebx, 1                  ; ★ 揚帆(0x20..0x23)且**完全無風** → 動不了
//
// ⇒ **玩家的航行只有這一條閘門:揚著帆而且風是 Calm。有風就照走,不管風向。**
//
// ⚠⚠ 這一支此前是**照敵船的規則寫的**(`docs/re/83`、`84`)。舊版查那張
// 4×5 的延遲表,逆風與橫風都走不動 —— 但那張表只被 `sub_2D38` 讀,
// 而 `sub_2D38` 只處理**物件槽**;玩家的船是 `byte_3E08C` 載具碼,不是物件槽。
// 而且連那張表的極性都反了(`4` 是每回合都動,不是動不了)。
//
// ★ 風向真正的作用在 `sub_2D2D0`:揚帆時它依「帆向 vs 風向」多燒
// **0 / 1 / 2 個世界回合**(垂直 0、同向 1、反向 2)——
// 加的是**時間與怪物的行動機會**,不是位移。⬜ 那一段還沒實作。
func (s *State) CanSail(dx, dy int) bool {
	// 收帆的大船(0x24..0x27)與其他載具不受風影響。
	if u5data.VehicleKind(s.Transport) != u5data.VehicleSailing {
		return true
	}
	if s.Wind != u5data.WindCalm {
		return true
	}
	s.Log(MsgBecalmed)
	return false
}

// IsSailing 回報隊伍是不是正在駕船。
func (s *State) IsSailing() bool {
	k := s.Transport &^ 0x03
	return k == u5data.VehicleShip || k == u5data.VehicleSailing
}

// WindShown 回報狀態列此刻該不該畫風向(原版 `sub_2A984` 的三道閘門)。
//
// ⚠⚠ **這一支與 `WindName` 原本都沒有呼叫者** —— 狀態列從來沒畫過風向,
// 而風向是航海的核心資訊(頂風走不動)。同一類問題見 `docs/re/80`。
//
//	byte_3E0A3 >= 0x21  → 不畫(地牢與戰鬥)
//	byte_3E0A3 == 0x19  → 不畫(★ 亞拉臘號殘骸)
//	byte_3E0A5 >= 0x80  → 走另一條分支(地下世界)
func (s *State) WindShown() bool {
	if s.InDungeon() || s.InCombat() {
		return false
	}
	if s.Location >= u5data.DungeonLocationBase {
		return false
	}
	// ★ 亞拉臘號殘骸不畫 —— 與 `SightAlwaysDarkLocation` 是同一個地點編號。
	if s.Location == SightAlwaysDarkLocation {
		return false
	}
	return s.Floor >= 0
}

// WindName 是給狀態列用的風向名。
func (s *State) WindName() string {
	if s.Wind < 0 || s.Wind >= u5data.WindCount {
		return u5data.WindNameZH[u5data.WindCalm]
	}
	return u5data.WindNameZH[s.Wind]
}
