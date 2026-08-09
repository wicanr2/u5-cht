package game

import "github.com/wicanr2/u5-cht/internal/u5data"

// 「誰來做?」(原版 `sub_E19C`,選單本身是 `sub_2A7F4`)
//
// 好幾個指令要先決定由哪位隊員動手:喝噴泉的水、凝視水晶球、之後的用物品。
// 原版把這件事收在一支函式裡,而且**只有在必要時才問**:
//
//	1. 目前是單人狀態(`byte_3E08B != 0xFF`)→ 就是那個人,不問。
//	2. 否則數狀態 'G'(良好)或 'P'(中毒)的隊員:
//	   0 人 → -1;1 人 → 就是他,不問;2 人以上 → 印「Player: 」問。
//
// ⚠ 第 2 點的「能動」判定**含中毒**。寫成「只有 G 才算」會讓中毒的隊員
// 突然不能做事 —— 那不是原版。
//
// ✅ 第 1 條已接上(`docs/re/97`):數字鍵指令 `SetActivePlayer` 就是原版的
// 「Set Active Plr」(`sub_2BD40`),而 `byte_3E08B` 對應 `State.activeMember`。
//
// ✅ 第 2 點的「2 人以上 → 問」也接上了 —— 選單本體 `sub_2A7F4` 在
// `pickmember.go`(`docs/re/98`)。
//
// ⚠⚠ **這裡曾經有一支同步的 `pickCharacter(prompt) int`**,它在多人時
// 直接取最後一位能動的隊員、不問玩家。已刪掉,原因不只是「行為不對」:
// 只要那一支還在,新的呼叫端就會挑它寫(同步的比較好寫),而**每一個這樣的
// 呼叫端都是一個靜靜跳過選單的地方**。⇒ 不留同步版本,讓「要問」成為唯一的路。

// damageMember 在戰鬥之外扣一名隊員的血(原版 `sub_2A464`)。
//
// 與 `applyDamage` 的差別:那一支是戰鬥用的,要算抗性、發經驗值、通知 AI。
// 這一支是「地圖上的東西讓你掉血」,原版就只有減血與判死。
func (s *State) damageMember(i, dmg int) {
	// ★ A 級證據:原版 `sub_2AC08` 用索引 7 = DAME1(「ダメージ」)。
	s.PlaySFX(u5data.SFXDamage1)
	if i < 0 || i >= len(s.Roster) {
		return
	}
	ch := &s.Roster[i]
	hp := int(ch.HP) - dmg
	if hp <= 0 {
		hp = 0
		ch.Status = u5data.StatusDead
		s.Log(ch.Name + "倒下了!")
	}
	ch.HP = uint16(hp)
}

// peerAtTheLand 是水晶球看見的全景 —— 與 In Quas Wis 同一支(原版 `sub_EDD4`)。
func (s *State) peerAtTheLand() { s.Peer() }

// actingMember 回報戰鬥中行動者的名冊索引;不在戰鬥中或行動者不是隊員回 −1。
func (s *State) actingMember() int {
	c := s.Combat
	if c == nil || c.Turn < 0 || c.Turn >= len(c.Units) {
		return -1
	}
	u := &c.Units[c.Turn]
	if !u.IsParty() || u.Roster < 0 || u.Roster >= len(s.Roster) {
		return -1
	}
	return u.Roster
}
