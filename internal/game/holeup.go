package game

import (
	"encoding/binary"
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

// 紮營的規則(原版 `sub_165C8`)。
const (
	// campAmbushOneIn 是**每過一個遊戲小時**擲一次的突襲機率分母
	//(原版 `random(0, 0x3F) == 0`)。睡九小時大約 13% 會被襲擊。
	campAmbushOneIn = 0x3F
	// campStepMinutes 是紮營時每次推的分鐘數(原版 `sub_29304(5)`)。
	campStepMinutes = 5
	// campMinHours 是「有效果」的最低時數 —— **睡不到 6 小時完全沒用**
	//(原版 `cmp arg_8, 5; jle → "No effect..."`)。
	campMinHours = 5
	// campHealMax 是紮營回的血量上限(原版 `random(1, 0x3F)`)。
	//
	// ⚠ **不是回滿。** 回滿是旅店(`sub_21D48`)的規則,兩條路不一樣;
	// 我第一版把旅店那支拿來用,等於讓野外紮營跟住旅店一樣好。
	campHealMax = 0x3F
	// campRestCooldown 是休息後的冷卻(原版 `byte_3E09C = 0x0E`)。
	// 冷卻沒退完再紮營只會得到「毫無效果」。
	campRestCooldown = 0x0E
	// campApparitionPercent 是老人出現的機率(原版 `random(0,99) < 25`)。
	campApparitionPercent = 25
	// campWrapAt 是紮營的時刻環繞值。
	//
	// ★ **這裡減的是 24(0x18),而睡床減的是 23(0x17)。**
	// 同一個遊戲的兩條休息路徑,一條對一條錯 —— 所以那個 off-by-one
	// 是真的寫錯了,不是某種刻意的設計。兩邊都照抄。
	campWrapAt = 0x18
)

// campAmbushCreature 是紮營被突襲時抽到的八種怪物(原版 `byte_3F3C8`)。
//
// 值是生物索引:41 巨魔 / 20 巨鼠 / 21 蝙蝠 / 24 泥怪 / 22 巨蛛 / 25 小鬼 /
// 36 無頭者 / 20 巨鼠。⚠ 巨鼠出現**兩次**(索引 1 與 7)—— 機率是雙倍,
// 不是我抄重複了。而且這張表與地牢的 `byte_3F3D0` **只差在頭尾兩筆**
// (地牢有收割者與凝視魔,野外換成巨魔與無頭者),兩張表緊鄰在執行檔裡。
var campAmbushCreature = [8]byte{0x29, 0x14, 0x15, 0x18, 0x16, 0x19, 0x24, 0x14}

// camp 紮營歇息 hours 小時,watch 是守夜的人(−1 = 無人)。
//
// 原版 `sub_165C8`。三件事按這個順序:睡 → 可能被襲擊 → 恢復。
func (s *State) camp(hours int, watch int) {
	s.Watch = watch
	if watch >= 0 && watch < len(s.Roster) {
		s.Log(s.Roster[watch].Name + "守夜。")
	}
	// 中毒的人記下來 —— 他們不會恢復(而且**也不會死**;毒死是旅店的規則)。
	poisoned := map[int]bool{}
	for i, c := range s.Party() {
		if c.Status == u5data.StatusPoisoned {
			poisoned[i] = true
		}
		if c.Status == u5data.StatusGood && i != watch {
			c.Status = u5data.StatusAsleep
			c.Raw[u5data.CharStatus] = u5data.StatusAsleep
		}
	}
	s.Log("Zzzzzz……")

	target := s.Clock.Hour + hours
	if target > holeUpWrapAt {
		target -= campWrapAt
	}
	last := s.Clock.Hour
	for guard := 0; s.Clock.Hour != target && guard < 24*12+1; guard++ {
		s.AdvanceTime(campStepMinutes)
		if s.Clock.Hour == last {
			continue
		}
		last = s.Clock.Hour
		// ★ 突襲的骰子**一小時擲一次**,不是一步一次。
		if s.Roll(0, campAmbushOneIn) != 0 {
			continue
		}
		s.campAmbush(watch)
		return
	}
	s.finishCamp(hours, watch, poisoned)
}

// campAmbush 是「遭到突襲!」(原版 `loc_167CE` 那一段的骰子中了之後)。
//
// ★ **守夜的意義全在這裡**:有人守夜,全隊在開打前就站起來了;
// 沒人守夜,大家還躺著('S')就被打 —— 睡著的人在戰鬥裡動不了。
func (s *State) campAmbush(watch int) {
	idx := campAmbushCreature[s.Roll(0, len(campAmbushCreature)-1)]
	s.Log("遭到突襲!")
	if watch >= 0 {
		for _, c := range s.Party() {
			if c.Status == u5data.StatusAsleep {
				c.Status = u5data.StatusGood
				c.Raw[u5data.CharStatus] = u5data.StatusGood
			}
		}
	}
	kind := u5data.CreatureBase + idx*4
	if s.InDungeon() {
		s.campDungeonAmbush(kind)
		return
	}
	o := &u5data.MapObject{X: s.X, Y: s.Y, Kind: kind}
	o.Raw[0] = kind
	o.Raw[u5data.ObjX], o.Raw[u5data.ObjY] = byte(s.X), byte(s.Y)
	s.beginCombatWith(o)
}

// campDungeonAmbush 是地牢裡紮營被突襲(原版走 `byte_3E0B1 = 6` 那條)。
func (s *State) campDungeonAmbush(kind byte) {
	d := s.Dungeon
	arena := u5data.BuildDungeonArena(u5data.DungeonArena{
		Number: d.Location - u5data.DungeonLocationBase + 1,
		Here:   s.DungeonTileHere(),
		Around: s.dungeonNeighbours(),
		Facing: int(d.Facing),
	})
	arena.EnemyKind[0] = kind
	if s.beginRoomCombat(arena, -1) {
		s.markArena(u5data.DungeonArenaModeCamp)
	}
}

// finishCamp 是睡完一整段之後的恢復(原版 `loc_16A33`)。
//
// ⚠⚠ **三條與旅店完全不同的規則**,我第一版全部弄錯(拿旅店那支來用):
//
//	1. **睡不到 6 小時完全沒效果。** 原版 `cmp arg_8, 5; jle → "No effect..."`。
//	2. **回的血是 `random(1, 63)`,不是回滿。** 回滿是旅店。
//	3. **守夜的人什麼都不恢復。** 那是派人守夜要付的代價。
//
// 另外中毒的人不恢復,但**不會死** —— 「睡覺毒發身亡」也是旅店的規則。
func (s *State) finishCamp(hours int, watch int, poisoned map[int]bool) {
	defer s.wakeCamp()
	if s.RestCooldown >= 1 || hours <= campMinHours {
		s.Log("毫無效果……")
		return
	}
	s.Log("隊伍歇息過了!")
	for i, c := range s.Party() {
		if poisoned[i] || c.Status == u5data.StatusDead || i == watch {
			continue
		}
		hp := int(c.HP) + s.Roll(1, campHealMax)
		if hp > int(c.MaxHP) {
			hp = int(c.MaxHP)
		}
		c.HP = uint16(hp)
		binary.LittleEndian.PutUint16(c.Raw[u5data.CharHP:], c.HP)
		// 法力:法師回滿智力,其餘只有一半。
		// ⚠ 與旅店的分類不同 —— 這裡只認 'A',旅店把 'M' 也算進去。
		mp := int(c.Intel)
		if c.Class != 'A' {
			mp /= 2
		}
		c.MP = byte(mp)
		c.Raw[u5data.CharMP] = c.MP
	}
	// 老人(升級)只在紮營成功之後擲,而且四分之一。
	s.MaybeApparition()
	s.RestCooldown = campRestCooldown
}

// wakeCamp 把睡著的人叫起來(原版兩條路共用的收尾)。
func (s *State) wakeCamp() {
	for _, c := range s.Party() {
		if c.Status == u5data.StatusAsleep {
			c.Status = u5data.StatusGood
			c.Raw[u5data.CharStatus] = u5data.StatusGood
		}
	}
	s.Log("汝醒來了。")
}

// sleepInBed 睡床(原版 `sub_16BA0`)。
//
// ★★ **睡床什麼都不恢復。** 原版的尾巴只做三件事:全隊 'S' → 'G'、
// 把隊伍座標 **x + 1**(從床上下來往東挪一格)、還原音樂。
// 沒有一行動到 HP 或 MP。
//
// 所以睡床的用途是**讓時間過去**(等 NPC 上班、等商店開門),不是治傷 ——
// 治傷要紮營(≥ 6 小時)或住旅店。憑「睡在床上當然比野地舒服」的直覺
// 去補恢復,就是自創遊戲。
//
// # 目標時刻的算法有個 off-by-one
//
//	edi = 現在的小時 + 要睡的小時
//	if (edi > 0x17) edi -= 0x17       ← **減 23,不是 24**
//
// 22 時睡 4 小時 → 26 − 23 = 3 時,而正確的環繞是 2 時。
// **紮營那條路減的是 24**(`campWrapAt`)—— 同一個遊戲兩條路,一條對一條錯,
// 所以這是真的寫錯,不是刻意設計。兩邊都照抄(CLAUDE.md §3.0)。
func (s *State) sleepInBed(hours int) {
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
	for guard := 0; s.Clock.Hour != target && guard < 24*6+1; guard++ {
		s.AdvanceTime(holeUpStepMinutes)
	}
	s.wakeCamp()
	// 從床上下來 —— 原版 `inc byte_3E0A6`,往東挪一格。
	s.X++
}

const (
	// holeUpWrapAt 是那個 off-by-one 的環繞值(原版 `cmp edi, 17h` / `sub edi, 17h`)。
	holeUpWrapAt = 0x17
	// holeUpStepMinutes 是睡覺時每次推的分鐘數(原版 `sub_29304(0Ah)`)。
	holeUpStepMinutes = 10
)

// ---------------------------------------------------------------- 通用提問

// AskNumber 問一個一位數(原版的「(1-9)」,`sub_2B7F0(1)`)。
//
// 原版收 '0'..'9' 與空白,其餘按鍵一律忽略繼續等;空白與 '0' 都是放棄。
// 這裡把「放棄」統一成回呼收到 0。
func (s *State) AskNumber(max int, then func(int)) {
	s.askNumber(1, max, then)
}

// AskNumberDigits 問一個最多 digits 位的數字(原版 `sub_2B7F0(digits)`)。
//
// 兩位數的地方目前只有調藥的「要幾份?」。一位數按下去就送出,
// 兩位以上要按 Enter —— 這是原版輸入欄的行為,不是機制差異。
func (s *State) AskNumberDigits(digits, max int, then func(int)) {
	s.askNumber(digits, max, then)
}

func (s *State) askNumber(digits, max int, then func(int)) {
	s.numReturn = s.Prompt
	s.numThen, s.numMax, s.numDigits = then, max, digits
	s.numInput = ""
	s.Prompt = PromptNumber
}

// AnswerNumber 是玩家按下數字鍵(0 = 放棄)。超出上限的照原版忽略。
//
// 一位數模式按下去就結束;多位數模式是把數字累積起來,由 SubmitNumber 送出。
func (s *State) AnswerNumber(n int) {
	if s.Prompt != PromptNumber {
		return
	}
	if s.numDigits > 1 {
		if n < 0 || n > 9 || len(s.numInput) >= s.numDigits {
			return
		}
		// 開頭的 0 沒有意義,照原版的輸入欄一樣不收。
		if n == 0 && s.numInput == "" {
			s.finishNumber(0)
			return
		}
		s.numInput += string(rune('0' + n))
		return
	}
	if n < 0 || n > s.numMax {
		return // 原版繼續等,不是取消
	}
	s.finishNumber(n)
}

// SubmitNumber 是多位數模式的 Enter。
func (s *State) SubmitNumber() {
	if s.Prompt != PromptNumber {
		return
	}
	n := 0
	for _, c := range s.numInput {
		n = n*10 + int(c-'0')
	}
	// 超過上限的照原版繼續等,不是取消。
	if n > s.numMax {
		s.numInput = ""
		return
	}
	s.finishNumber(n)
}

// BackspaceNumber 是多位數模式的 Backspace。
func (s *State) BackspaceNumber() {
	if s.Prompt == PromptNumber && s.numInput != "" {
		s.numInput = s.numInput[:len(s.numInput)-1]
	}
}

// NumberInput 是多位數模式目前打進去的字串(給算繪看)。
func (s *State) NumberInput() string { return s.numInput }

func (s *State) finishNumber(n int) {
	then := s.numThen
	s.numThen, s.numInput = nil, ""
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

// AskText 問一行字(原版 `sub_239B4(緩衝區, 上限)`)。
//
// max 是收得下幾個字元;送出時把去掉頭尾空白的結果交給 then。
// **空字串是合法的回答** —— 原版對「什麼都沒打」有自己的處理(改名時就是保留原名),
// 所以不能在這一層擋掉。
func (s *State) AskText(question string, max int, then func(string)) {
	s.textReturn = s.Prompt
	s.textThen, s.textMax = then, max
	s.Input = ""
	s.Prompt = PromptText
	if question != "" {
		s.Log(question)
	}
}

// TypeText 收一個字元(只收可列印的 ASCII —— 原版的輸入欄也只收這些)。
func (s *State) TypeText(r rune) {
	if s.Prompt != PromptText || len(s.Input) >= s.textMax {
		return
	}
	if r < ' ' || r > '~' {
		return
	}
	s.Input += string(r)
}

// BackspaceText 退一個字元。
func (s *State) BackspaceText() {
	if s.Prompt == PromptText && s.Input != "" {
		s.Input = s.Input[:len(s.Input)-1]
	}
}

// SubmitText 送出。
func (s *State) SubmitText() {
	if s.Prompt != PromptText {
		return
	}
	text := trimSpace(s.Input)
	then := s.textThen
	s.Input, s.textThen = "", nil
	s.Prompt = s.textReturn
	if then != nil {
		then(text)
	}
}

// AwaitingText 回報是不是正在等一行字。
func (s *State) AwaitingText() bool { return s.Prompt == PromptText }
