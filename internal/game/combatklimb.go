package game

import "github.com/wicanr2/u5-cht/internal/u5data"

// 戰鬥中的 K(原版 `sub_16058`)—— 與地圖上那三支都不同
//
// 地圖上的 Klimb 有三個入口(地表 `sub_188C4`、場景 `sub_EA0`、地牢 `sub_417C`),
// 戰鬥又是第四支。它做兩件事:
//
//	腳下是梯子 → **離開戰場**(向上或向下),走 `sub_15E20`
//	否則       → 問方向,爬過那一格的 **tile 0x4C**
//
// ★ 離場那一支順帶解出一條規則:**全隊必須從同一個出口離開**
// (`sub_15E20` 的 "All must use the same exit!")。第一個離場的人選定出口,
// 之後的人只能跟。⚠ 那條有個 `byte_3E0B1 & 0x80` 的閘門,而目前四種戰場模式
//(0 一般遭遇 / 2 地牢遊蕩 / 4 地表紮營 / 6 地牢紮營)**都不帶 0x80**,
// 所以規則存在但還踩不到。照抄閘門,不假裝它一定生效。

// 戰場上的三種梯子 tile(原版 `cmp esi, 0C8h` / `0C9h` / `86h`)。
const (
	CombatLadderUp   = 0xC8
	CombatLadderDown = 0xC9
	// CombatHatch 是另一種向下的出口(0x86),而且**還多一道閘門**:
	// `byte_3E0B1 & 0x80` 設起來時它不算(原版 `test al, 80h; jnz`)。
	CombatHatch = 0x86
	// CombatClimbable 是「爬得過去」的那一種 tile(原版 `cmp byte ptr [eax], 4Ch`)。
	CombatClimbable = 0x4C
)

// 離場方向碼,與 `sub_100F8` 的上下樓同一組(5 上 / 6 下)。
const (
	CombatExitUp   = 5
	CombatExitDown = 6
)

// CombatKlimb 是戰鬥中的 K。
func (s *State) CombatKlimb() {
	c := s.Combat
	if c == nil || c.Turn < 0 || c.Turn >= len(c.Units) {
		return
	}
	u := &c.Units[c.Turn]
	s.Log("攀爬 ——")
	switch tile := s.CombatTileAt(u.X, u.Y); tile {
	case CombatLadderUp:
		// 兩向梯要問上下 —— 而「這是兩向梯」是**畫戰場時**記下來的
		// (原版 `sub_FD54` 對地牢地形 kind 3 設 `byte_418DE = 1`)。
		if c.LadderBoth {
			s.Ask("上(Y)還是下(N)?", func(up bool) {
				if up {
					s.combatLeave(u, CombatExitUp)
					return
				}
				s.combatLeave(u, CombatExitDown)
			})
			return
		}
		s.combatLeave(u, CombatExitUp)
	case CombatLadderDown:
		s.combatLeave(u, CombatExitDown)
	case CombatHatch:
		s.combatLeave(u, CombatExitDown)
	default:
		s.AskDirection(func(d Direction) { s.combatClimbOver(u, d) })
	}
}

// combatClimbOver 爬過鄰格的那一種地形(原版 `loc_1612A`)。
//
// ⚠ 只有 **tile 0x4C** 爬得過去,而且那一格不能有人站著;
// 其餘一律「此為何意?」——不是「無處可攀」之類的自創訊息。
func (s *State) combatClimbOver(u *Combatant, d Direction) {
	c := s.Combat
	dx, dy := d.Delta()
	nx, ny := u.X+dx, u.Y+dy
	if s.CombatTileAt(nx, ny) != CombatClimbable {
		s.Log(MsgWhat)
		return
	}
	if _, taken := c.CombatUnitAt(nx, ny); taken {
		s.Log(MsgWhat)
		return
	}
	u.X, u.Y = nx, ny
	s.focusCombatUnit(u)
	s.afterPlayerAction()
}

// combatLeave 讓一個單位從戰場離開(原版 `sub_15E20`)。
func (s *State) combatLeave(u *Combatant, exit int) {
	c := s.Combat
	// 在船上的隊伍不能自己走掉。
	if k := u5data.VehicleKind(s.Transport); k == u5data.VehicleShip || k == u5data.VehicleSailing {
		s.Log("須與船同行!")
		return
	}
	// 全隊必須從同一個出口 —— 見檔頭的閘門說明。
	if u.IsParty() {
		if c.ExitDir == 0 {
			c.ExitDir = exit
		} else if c.ExitDir != exit && c.sameExitEnforced() {
			s.Log("全隊須從同一出口離去!")
			return
		}
	}
	if exit == CombatExitUp {
		s.Log(MsgUp)
	} else {
		s.Log(MsgDown)
	}
	u.Flags |= UnitDead // 離場 —— 用「不在場上」表示,名冊裡的人沒事
	c.Left++
	// 隊伍全走了就結束,並把梯子的效果套到地牢樓層上。
	if enemies, party := s.sideCounts(c); party == 0 || enemies == 0 {
		s.finishCombatExit(exit)
		return
	}
	s.afterPlayerAction()
}

// finishCombatExit 收尾:離開戰場,若是地牢的梯子就換樓層。
func (s *State) finishCombatExit(exit int) {
	d := s.Dungeon
	s.EndCombat(false)
	if d == nil {
		return
	}
	switch {
	case exit == CombatExitUp && d.Level > 0:
		d.Level--
	case exit == CombatExitDown && d.Level < u5data.DungeonLevels-1:
		d.Level++
	case exit == CombatExitUp:
		s.LeaveDungeon() // 第一層往上就是出地面
	}
}

// sameExitEnforced 回報這一場要不要強制「同一個出口」。
//
// 原版的閘門是 `byte_3E0B1 & 0x80`。目前四種模式都不帶 0x80,所以一律 false ——
// **照抄閘門,不假裝規則一定生效**。等哪天逆出帶 0x80 的那條路再開。
func (c *Combat) sameExitEnforced() bool { return c.ArenaMode&0x80 != 0 }
