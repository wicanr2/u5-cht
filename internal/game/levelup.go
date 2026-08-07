package game

import (
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 升級:「奇異而熟悉的老人」
//
// U5 沒有「打完怪就跳等」這回事 —— 升級是一個**事件**:紮營休息時有
// **1/4 機率**(`sub_165C8` 的 `random(0,99) < 25`)出現一位老人,
// 印「An apparition!」,把全隊照經驗值重算等級,然後消失。
//
// 等級的算式(`sub_10C34`):
//
//	等級 = 1
//	n = 經驗值 / 100
//	while n > 0 { 等級++; n >>= 1 }
//
// 也就是門檻 **100 / 200 / 400 / 800 / 1600 / 3200 / 6400**,而經驗值上限
// 9999 → 最高 8 級。8 級正好是最高的咒語圈數 —— 兩者是同一個數字。
//
// 升級時:
//
//	生命上限 = 目前生命 = 等級 × 30
//	隨機一項三圍 +1(上限 30),印「stronger! / quicker! / wiser!」
//	魔力照職業重算

// MaxCharacterLevel 是等級上限(經驗值 9999 算出來就是 8)。
const MaxCharacterLevel = 8

// HPPerLevel 是每一級的生命(`lea edx,[ebx+ebx*4]` → ×5、`[edx+edx*2]` → ×15、
// `add edx,edx` → ×30)。
const HPPerLevel = 30

// StatCap 是三圍的上限(`sub_2BBB8` 的第三個參數 0x1E)。
const StatCap = 30

// ApparitionChance 是紮營時老人出現的機率分子(百分之)。
const ApparitionChance = 25

// LevelForExp 依經驗值算等級。
func LevelForExp(exp int) int {
	level := 1
	for n := exp / 100; n > 0; n >>= 1 {
		level++
	}
	if level > MaxCharacterLevel {
		level = MaxCharacterLevel
	}
	return level
}

// ManaForClass 依職業與智力算魔力上限(`sub_10C34` 的尾段)。
//
//	'A' 聖者 / 'M' 法師 → 智力
//	'B' 吟遊詩人        → 智力 / 2
//	其餘('F' 戰士)     → 0
func ManaForClass(class byte, intel int) int {
	switch class {
	case 'A', 'M':
		return intel
	case 'B':
		return intel / 2
	}
	return 0
}

// Apparition 是老人現身的整個事件。回傳有沒有人升級。
func (s *State) Apparition() bool {
	s.Log("一個幻影!")
	any := false
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		if s.levelUp(i) {
			any = true
		}
	}
	s.Log("那位奇異而熟悉的老人消失了……")
	return any
}

// levelUp 依經驗值把一名角色調到該有的等級。回傳有沒有真的升。
func (s *State) levelUp(i int) bool {
	ch := &s.Roster[i]
	want := LevelForExp(int(ch.Exp))
	if int(ch.Level) == want {
		// 等級沒變就什麼都不做 —— 原版 `cmp edx, ebx; jz loc_10ECE` 直接跳過,
		// 連魔力都不重算。
		return false
	}
	ch.Level = byte(want)
	hp := want * HPPerLevel
	ch.MaxHP, ch.HP = uint16(hp), uint16(hp)

	s.Log("「向汝致敬," + s.charName(ch) + "!")
	s.Log("為表彰汝之壯舉,吾將賜汝獎賞。")
	s.Log("汝已達第 " + itoa(want) + " 級,並且")
	switch s.Roll(1, 3) {
	case 1:
		s.Log("更強壯了!」")
		ch.Strength = byte(addCap(int(ch.Strength), 1, StatCap))
		ch.Raw[u5data.CharStrength] = ch.Strength
	case 2:
		s.Log("更敏捷了!」")
		ch.Dex = byte(addCap(int(ch.Dex), 1, StatCap))
		ch.Raw[u5data.CharDex] = ch.Dex
	default:
		s.Log("更睿智了!」")
		ch.Intel = byte(addCap(int(ch.Intel), 1, StatCap))
		ch.Raw[u5data.CharIntel] = ch.Intel
	}
	// 魔力照職業重算 —— 戰士永遠是 0。
	if mp := ManaForClass(ch.Class, int(ch.Intel)); ch.Class == 'A' || ch.Class == 'M' || ch.Class == 'B' {
		ch.MP = byte(mp)
	}
	return true
}

// addCap 是有上限的加法(原版 `sub_2BBB8`)。
func addCap(v, add, cap int) int {
	if v += add; v > cap {
		return cap
	}
	return v
}

// charName 是角色的顯示名;沒名字時退回職業。
func (s *State) charName(ch *u5data.Character) string {
	if ch.Name != "" {
		return ch.Name
	}
	return ch.ClassName()
}

// MaybeApparition 是紮營 / 休息之後那一擲。
//
// 原版 `sub_165C8`:`random(0, 99) < 25` 才出現(而且要不是被埋伏那種情況)。
func (s *State) MaybeApparition() bool {
	if s.Roll(0, 99) >= ApparitionChance {
		return false
	}
	s.Apparition()
	return true
}
