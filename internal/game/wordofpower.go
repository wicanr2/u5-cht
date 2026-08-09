package game

import (
	"strings"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// Yell 指令(原版 `sub_17E74`)
//
// 一個鍵,三種完全不同的用途 —— 原版就是這樣分派的:
//
//	在有帆的船上(載具 & 0xF8 == 0x20)  → 收帆 / 揚帆,不問任何字
//	在地點 30/31/32(三團聖火所在)      → 喊名字召來暗影君主(sub_17A14)
//	在大地圖上(地點 0)                 → 力量之言(sub_17CFC)
//	其餘地點                            → 「毫無效果!」
//
// ⚠ **收帆那條是「不問就做」。** 原版在船上按 Yell 根本不會跳出輸入框;
// 寫成「先問再判斷是不是在船上」會多出一個原版沒有的按鍵。

// Yell 按下 Yell 鍵。回傳是不是進到了輸入模式。
func (s *State) Yell() bool {
	// ⚠ 戰鬥中(地點 ≥ 0x80,原版把 byte_3E0A3 設成 −1)不收放帆。
	if s.Transport&0xF8 == u5data.VehicleSailing && s.Location < 0x80 {
		if s.Transport&0xFC == u5data.VehicleSailing {
			s.Log(MsgFurl)
			s.Transport += 4 // 0x20..0x23(揚帆)→ 0x24..0x27(收帆)
		} else {
			s.Log(MsgHoist)
			s.Transport -= 4
		}
		return false
	}
	s.Prompt = PromptYell
	s.Input = ""
	s.Log(MsgYellWhat)
	return true
}

// CancelYell 作罷,不喊了。
func (s *State) CancelYell() {
	s.Input = ""
	if s.Prompt == PromptYell {
		s.Prompt = PromptNone
	}
}

// SubmitYell 把打好的一行送出去。
func (s *State) SubmitYell() {
	text := strings.TrimSpace(s.Input)
	s.Input = ""
	s.Prompt = PromptNone
	if text == "" {
		s.Log(MsgYellNothing)
		return
	}
	switch {
	case s.Location >= 1 && s.Location <= len(u5data.Locations):
		s.yellInTown(text)
	case s.Location == 0:
		s.SpeakWord(text)
	default:
		s.Log(MsgNoEffect)
	}
}

// yellInTown 是在城裡喊出聲(原版 `sub_17A14`)—— 目前只有召喚暗影君主一種效果。
//
// ⚠ **地點不對就沒事,而且沒有「這座城配這個名字」的檢查**:三團聖火所在的
// 任何一座裡喊任何一個名字,召來的都是那個名字的主人。照抄。
func (s *State) yellInTown(text string) {
	if !isFlameKeep(s.Location) {
		s.Log(MsgNoEffect)
		return
	}
	i := u5data.ShadowlordIndex(text)
	switch {
	case i < 0:
		s.Log(MsgNoEffect)
		return
	case s.Y < u5data.ShadowlordSpawnAhead:
		// 上方兩格擠不出位置(原版 `cmp byte_3E0A7, 2; jb`)。
		s.Log(MsgNoEffect)
		return
	case s.ShadowlordAt[i] == u5data.ShadowlordGone:
		s.Log(MsgNoEffect)
		return
	case s.shadowlordPresent():
		// 場上已經有一個了 —— 原版掃過三十二個 NPC 找 tile 0xFC。
		s.Log(MsgNoEffect)
		return
	}
	if !s.spawnShadowlord(i, s.X, s.Y-u5data.ShadowlordSpawnAhead) {
		s.Log(MsgNoEffect)
		return
	}
	s.ShadowlordHere = byte(i)
	s.Log(MsgShadowlordAppears)
	// ⚠ 原版尾端那段 `byte_3E08A` 存 'T' → 清 0 → 重畫 → 存回 'T' **不是**解除停時,
	// 是「重畫時不要走停時那條路徑」的臨時開關。停時的回合數不受影響,別照抄成取消。
}

// shadowlordPresent 回報場上是不是已經有一個暗影君主。
//
// 原版掃的是三十二個地圖物件(`dword_3E46C[esi*8] == 0xFC`);這裡掃 NPC 槽,
// 因為引擎把暗影君主放在 NPC 那一側(牠會走動、會攻擊)。
func (s *State) shadowlordPresent() bool {
	if s.npcs == nil {
		return false
	}
	for i := range s.npcs {
		if s.npcs[i].Creature == u5data.TileShadowlord && !s.removed[s.Location<<8|i] {
			return true
		}
	}
	return false
}

// spawnShadowlord 把暗影君主塞進一個空的 NPC 槽。
//
// ⚠ **找槽的方式照抄:從 31 往下掃,停在第一個空的**(不是最小的空槽)。
// 全滿時原版直接蓋掉 31 號 —— 也照抄,不要改成「放不下就失敗」。
func (s *State) spawnShadowlord(which, x, y int) bool {
	if s.npcs == nil {
		return false
	}
	slot := u5data.NPCsPerLocation - 1
	for j := u5data.NPCsPerLocation - 1; j >= 0; j-- {
		if s.npcs[j].Creature == 0 {
			slot = j
			break
		}
	}
	n := &s.npcs[slot]
	*n = u5data.NPC{Creature: u5data.TileShadowlord}
	// 三個排程 slot 全指向牠出現的那一格,四個時刻全是 0 ——
	// 原版就是這樣填的(`byte_3E570[..] = 6` 是 AI 型別,X/Y/樓層三份都一樣)。
	for k := 0; k < 3; k++ {
		n.Schedule.AI[k] = shadowlordAI
		n.Schedule.X[k] = byte(x)
		n.Schedule.Y[k] = byte(y)
		n.Schedule.Floor[k] = byte(s.Floor)
	}
	delete(s.removed, s.Location<<8|slot)
	if len(s.rtNPCs) != u5data.NPCsPerLocation {
		s.initRuntimeNPCs()
	}
	s.rtNPCs[slot] = RuntimeNPC{
		Mode:     ModeIdle,
		X:        x,
		Y:        y,
		Floor:    s.Floor,
		Creature: u5data.TileShadowlord,
	}
	return true
}

// shadowlordAI 是原版填進排程 AI 欄的值(`byte_3E570[..] = 6`)。
const shadowlordAI = 6

// isFlameKeep 回報這個地點是不是三團聖火之一所在。
func isFlameKeep(loc int) bool {
	for _, n := range u5data.ShadowlordKeeps {
		if n == loc {
			return true
		}
	}
	return false
}

// 力量之言(原版 `sub_17CFC`)
//
// 說出八個字之一,會對**鄰格**做事:
//
//	鄰格是地牢入口(原地形或 0xDF) → 切換封印
//	鄰格是被玷污的聖壇(0x1A)      → 進入復原流程(要答美德名 + 三次真言)
//	都不是                          → 「毫無效果!」
//
// ⚠ **檢查鄰格的順序是固定的:西、南、東、北**,取第一個命中的
//(`byte_3F788` → `byte_3F7A9` → `byte_3F78A` → `byte_3F769`,
// 也就是視窗緩衝裡玩家那一格的 −1 / +32 / +1 / −32)。
// 兩個目標同時在旁邊時,順序決定了動到哪一個。
//
// ⚠ **地牢那條路還要座標對得上。** 光是「鄰格看起來像地牢入口」不夠,
// 那一格必須真的是第 i 座地牢的入口座標 —— 不然 FALLAX 可以拿去開任何一座山洞。
// 而且座標不合時原版**就此結束**:不繼續看別的方向,也**不印「毫無效果」**
//(`var_8` 已經被設成 1)。這個沉默是原版的行為,不是漏寫。

// wordNeighbours 是原版檢查鄰格的順序與方向。
var wordNeighbours = [4][2]int{
	{-1, 0}, // 西 byte_3F788
	{0, 1},  // 南 byte_3F7A9
	{1, 0},  // 東 byte_3F78A
	{0, -1}, // 北 byte_3F769
}

// SpeakWord 說出一個力量之言。回傳有沒有產生效果。
func (s *State) SpeakWord(spoken string) bool {
	i := u5data.WordOfPowerIndex(spoken)
	if i < 0 {
		s.Log(MsgNoEffect)
		return false
	}
	s.Log(MsgWordUttered)
	for _, d := range wordNeighbours {
		x, y := WrapWorld(s.X+d[0]), WrapWorld(s.Y+d[1])
		tile := s.TileAt(x, y)
		if !u5data.WordTargetTile(tile, i) {
			continue
		}
		if tile == u5data.TileShrineDesecrated {
			return s.beginShrineRestore(i, x, y)
		}
		// 地牢:座標必須真的是第 i 座的入口,不對就靜靜結束(見上方說明)。
		e := u5data.DungeonEntrances[i]
		if x != e.X || y != e.Y {
			return false
		}
		s.DungeonSeal[i] ^= u5data.DungeonSealedBit
		s.SetTileAt(x, y, u5data.ToggleDungeonSeal(tile, i))
		// ⚠⚠ **兩句話此前是反的。** 旗標的 bit 7 設著代表**通的**,
		// 清掉才是崩塌(`u5data.DungeonIsSealed` 的檔頭列了四條證據)。
		// 而 U5 的主線正是「八座地牢一開始進不去,要各自喊對那個字」——
		// 反過來的話玩家第一次喊對反而把路封死了。
		if u5data.DungeonIsSealed(s.DungeonSeal[i]) {
			s.Log(e.DisplayName() + "的入口崩塌了。")
		} else {
			s.Log(e.DisplayName() + "的入口開了。")
		}
		return true
	}
	s.Log(MsgNoEffect)
	return false
}

// beginShrineRestore 開始復原一座被玷污的聖壇(原版 `sub_17C2C`)。
//
// 流程與冥想**很像**(問美德名、問三次真言),但有三處不同,都照抄了:
//
//	1. 美德名比的是**完整的**英文名(`off_411BC` = "Honesty"),
//	   不是冥想用的四字母前綴(`off_55FEC` = "hone")
//	2. 失敗**不印任何訊息**,只停一下(`sub_27230(10)`)
//	3. 全對之後**還要再比一次座標** —— 在別的地方唸對真言不算數
func (s *State) beginShrineRestore(virtue, x, y int) bool {
	s.Shrine = &Shrine{
		Virtue:    virtue,
		OK:        true,
		Stage:     ShrineAskVirtue,
		Restoring: true,
		TargetX:   x,
		TargetY:   y,
	}
	s.Prompt = PromptShrine
	s.logMisc(u5data.MsgShrineWhich)
	return true
}

// shrineRestoreResolve 是復原流程的結尾。
func (s *State) shrineRestoreResolve() bool {
	sh := s.Shrine
	e := u5data.Shrines[sh.Virtue]
	// 真言錯、或者這一格不是那一座聖壇 —— 原版都是靜靜結束,沒有訊息。
	if !sh.OK || e.X != sh.TargetX || e.Y != sh.TargetY {
		s.EndMeditate()
		return false
	}
	s.ShrineFlag[sh.Virtue] &^= u5data.ShrineDesecratedBit
	s.SetTileAt(sh.TargetX, sh.TargetY, u5data.TileShrine)
	s.Log(MsgShrineRestored)
	s.EndMeditate()
	return true
}
