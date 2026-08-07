package game

import (
	"fmt"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// Hole up(H)—— 紮營、睡床、修船
//
// 原版位址:`sub_2ACF4` case 72 分派 → `sub_2B8CC`(紮營 / 修船)/ `sub_16BA0`(睡床)。
//
// ★ **同一個鍵做三件事**,而且是**先看載具、再看地點**:
//
//	在船上(載具 0x20..0x27)  → 修船,與地點無關
//	在城鎮 / 城堡(1..0x20)   → 必須站在**床**上(tile 0xAB),否則「唯有於床上。」
//	地表(0)或地牢(≥0x21)   → 紮營
//
// 順序不能倒過來:在城裡的碼頭上船照樣是修船,不是「城裡不能紮營」。
//
// 這一支之前完全沒接 —— 玩家出了城就**無法休息**,而休息是 U5 唯一的
// 回血與升級管道(`levelup.go` 的老人只在休息後出現)。
// 指令表全解之後才發現漏了它(`docs/re/49`)。

// HoleUpBedTile 是床(原版 `cmp edi, 0ABh`)。
const HoleUpBedTile = 0xAB

// 修船的規則(原版 `sub_2B8CC` 前半)。
const (
	// shipRepairTurns 是修船要跑幾個世界回合。
	shipRepairTurns = 5
	// shipRepairMinutes 是每個回合推的分鐘數。
	shipRepairMinutes = 5
	// shipRepairUntil 是「至少要修回這麼多」——
	// 迴圈是 do-while,所以**一定會加至少一次**,而且會一路加到 10 以上。
	shipRepairUntil = 10
	// ShipHullMax 是耐久上限(原版 `cmp …, 63h`)。
	ShipHullMax = 99
)

// holeUpMaxHours 是能睡幾小時(原版問「(1-9)」,收 '0'..'9' 與空白)。
const holeUpMaxHours = 9

// HoleUp 是 H 指令。
func (s *State) HoleUp() {
	s.Log("紮營 ——")
	// 先看載具:在大船上就是修船,不管站在哪裡。
	if k := u5data.VehicleKind(s.Transport); k == u5data.VehicleShip || k == u5data.VehicleSailing {
		s.repairShip()
		return
	}
	if s.InScene() {
		if s.TileAt(s.X, s.Y) != HoleUpBedTile {
			s.Log("唯有於床上。")
			return
		}
		s.askHoleUpHours(true)
		return
	}
	// 地表與地牢都是紮營。
	if !s.canCampHere() {
		return
	}
	s.askHoleUpHours(false)
}

// canCampHere 是紮營的兩道地表限制(原版 `loc_2B9E9` / `loc_2BA04`)。
//
// ⚠ 兩條都**只在地表以外的地點不成立時才檢查** —— 地牢裡(≥0x21)兩條都跳過,
// 所以在地牢裡騎馬也紮得起來。照抄。
func (s *State) canCampHere() bool {
	if s.InDungeon() {
		return true
	}
	// 水上不行。⚠ 原版寫的是 `tile != 0 && tile < 4`,**0 被排除在外** ——
	// 也就是深水(0)竟然過得了這一關,而淺水(1..3)不行。
	// 看起來像 off-by-one,但那是原版的行為;真的走到深水上早就在船上了,
	// 而在船上這一段根本到不了(前面就分去修船)。照抄,不「修好」它。
	if t := int(s.TileAt(s.X, s.Y)); t != 0 && t < 4 {
		s.Log("須在陸上或船上。")
		return false
	}
	if !u5data.IsOnFoot(s.Transport) {
		s.Log("汝須步行。")
		return false
	}
	return true
}

// repairShip 修船(原版 `sub_2B8CC` 的前半)。
func (s *State) repairShip() {
	s.Log("修補船身……")
	// 揚著帆修不了 —— 要先收帆(Yell 的「收帆」)。
	if u5data.VehicleKind(s.Transport) == u5data.VehicleSailing {
		s.Log("須先收帆!")
		return
	}
	// 五個世界回合,每回合推 5 分鐘。原版每一輪都重新確認還在船上
	// (`byte_3E08C & 0xFC != 0x24 → 中止`)—— 修船途中被打下船就停。
	for i := 0; i < shipRepairTurns; i++ {
		if u5data.VehicleKind(s.Transport) != u5data.VehicleShip {
			return
		}
		s.extraWorldTurn()
		s.AdvanceTime(shipRepairMinutes)
	}
	// ⚠ 這是 do-while:**至少加一次**,而且一路加到 10 以上才停。
	// 所以修一次船最少也會回到 10 —— 破成 1 的船修一輪就能航行。
	for {
		s.ShipHull += s.Roll(1, 3)
		if s.ShipHull > ShipHullMax {
			s.ShipHull = ShipHullMax
		}
		if s.ShipHull >= shipRepairUntil {
			break
		}
	}
	s.Log(fmt.Sprintf("船身耐久 %d!", s.ShipHull))
}

// askHoleUpHours 問要睡幾小時。bed 為真代表睡床(不問守夜)。
func (s *State) askHoleUpHours(bed bool) {
	s.Log("欲歇息幾時辰?(1-9)")
	s.AskNumber(holeUpMaxHours, func(hours int) {
		if hours <= 0 {
			s.Log("作罷。")
			return
		}
		if bed {
			s.sleepInBed(hours)
			return
		}
		s.askWatch(hours)
	})
}

// askWatch 問要不要派人守夜(原版 `loc_2BAA4`)。
//
// ⚠ 只有**能動的人超過一個**才問 —— 一個人的隊伍守了夜就沒人睡。
// 「能動」算的是狀態 'G' 或 'P'(中毒也算),但**真的派得上的只有 'G'**:
// 選了中毒的人,原版會回「無人守夜!」。兩個判定用不同的集合,不是筆誤。
func (s *State) askWatch(hours int) {
	able := 0
	for _, c := range s.Party() {
		if c.Status == u5data.StatusGood || c.Status == u5data.StatusPoisoned {
			able++
		}
	}
	if able <= 1 {
		s.camp(hours, -1)
		return
	}
	s.Ask("欲派人守夜否?", func(yes bool) {
		if !yes {
			s.camp(hours, -1)
			return
		}
		if !s.beginPick("何人守夜?", s.partyEntries(-1), MsgNobodyHere, func(member int) bool {
			if member < 0 || member >= len(s.Roster) ||
				s.Roster[member].Status != u5data.StatusGood {
				s.Log("無人守夜!")
				s.camp(hours, -1)
				return true
			}
			s.camp(hours, member)
			return true
		}) {
			s.camp(hours, -1)
		}
	})
}

// camp 紮營歇息 hours 小時,watch 是守夜的人(−1 = 無人)。
//
// ⚠ **遭遇那一層還沒做。** 原版走 `sub_2E364(4 或 6, watch, hours)` 進一個
// 專用戰場(地表 type 4、地牢 type 6),在那裡判要不要被襲擊、守夜的人有沒有
// 發現。那條路的解析見 `docs/re/48` §8 的未完項 —— 沒有證據就不補一個機率,
// 所以目前紮營一定睡得安穩。**這是已知的落差,不是「紮營做完了」。**
func (s *State) camp(hours int, watch int) {
	s.Watch = watch
	if watch >= 0 && watch < len(s.Roster) {
		s.Log(s.Roster[watch].Name + "守夜。")
	}
	s.restHours(hours)
}

// sleepInBed 睡床(原版 `sub_16BA0`)。
func (s *State) sleepInBed(hours int) {
	s.restHours(hours)
}

// restHours 是「睡 N 小時」的本體(原版 `sub_16BA0` 的後半)。
//
// # 目標時刻的算法有個 off-by-one
//
//	edi = 現在的小時 + 要睡的小時
//	if (edi > 0x17) edi -= 0x17       ← **減 23,不是 24**
//
// 所以跨過午夜時會多醒一小時:22 時睡 4 小時 → 26 → 26 − 23 = 3 時,
// 而正確的環繞是 2 時。**這是原版的 bug,照抄**(CLAUDE.md §3.0:
// 機制與原版一模一樣,包括它的 bug)。
//
// 時間是每次推 10 分鐘直到小時數對上 —— 不是一次加 hours×60,
// 因為途中要讓 NPC 走、讓事件觸發。
func (s *State) restHours(hours int) {
	target := s.Clock.Hour + hours
	if target > holeUpWrapAt {
		target -= holeUpWrapAt
	}
	for _, c := range s.Party() {
		if c.Status == u5data.StatusGood {
			c.Status = u5data.StatusAsleep
			c.Raw[u5data.CharStatus] = u5data.StatusAsleep
		}
	}
	s.Log("Zzzzzz……")
	// 上限擋住「target 永遠對不上」的情況(推 10 分鐘一次,一天 144 次)。
	for guard := 0; s.Clock.Hour != target && guard < 24*6+1; guard++ {
		s.AdvanceTime(holeUpStepMinutes)
	}
	for _, c := range s.Party() {
		s.wakeUp(c)
	}
	s.Log("汝醒來了。")
	s.MaybeApparition()
}

const (
	// holeUpWrapAt 是那個 off-by-one 的環繞值(原版 `cmp edi, 17h` / `sub edi, 17h`)。
	holeUpWrapAt = 0x17
	// holeUpStepMinutes 是睡覺時每次推的分鐘數(原版 `sub_29304(0Ah)`)。
	holeUpStepMinutes = 10
)

// ---------------------------------------------------------------- 通用提問

// AskNumber 問一個數字(原版的「(1-9)」)。
//
// 原版收 '0'..'9' 與空白,其餘按鍵一律忽略繼續等;空白與 '0' 都是放棄。
// 這裡把「放棄」統一成回呼收到 0。
func (s *State) AskNumber(max int, then func(int)) {
	s.numReturn = s.Prompt
	s.numThen, s.numMax = then, max
	s.Prompt = PromptNumber
}

// AnswerNumber 是玩家按下數字鍵(0 = 放棄)。超出上限的照原版忽略。
func (s *State) AnswerNumber(n int) {
	if s.Prompt != PromptNumber {
		return
	}
	if n < 0 || n > s.numMax {
		return // 原版繼續等,不是取消
	}
	then := s.numThen
	s.numThen = nil
	s.Prompt = s.numReturn
	if n > 0 {
		s.Log(fmt.Sprintf("%d", n))
	}
	if then != nil {
		then(n)
	}
}

// AwaitingNumber 回報是不是正在等數字。
func (s *State) AwaitingNumber() bool { return s.Prompt == PromptNumber }

// Ask 問一個 Y / N。
func (s *State) Ask(question string, then func(bool)) {
	s.ynReturn = s.Prompt
	s.yesNoThen = then
	s.Prompt = PromptYesNo
	s.Log(question)
}

// AnswerYesNo 是玩家按下 Y / N(ESC 等同 N,與原版一致)。
func (s *State) AnswerYesNo(yes bool) {
	if s.Prompt != PromptYesNo {
		return
	}
	then := s.yesNoThen
	s.yesNoThen = nil
	s.Prompt = s.ynReturn
	if yes {
		s.Log("是。")
	} else {
		s.Log(MsgNo)
	}
	if then != nil {
		then(yes)
	}
}

// AwaitingYesNo 回報是不是正在等 Y / N。
func (s *State) AwaitingYesNo() bool { return s.Prompt == PromptYesNo }
