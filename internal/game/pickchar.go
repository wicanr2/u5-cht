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
// 引擎目前沒有單人狀態,所以第 1 條先不做;等單人模式接上再補
//(留白比猜著寫好,見 CLAUDE.md §3.0)。

// pickCharacter 決定由誰動手。回傳名冊索引,沒人可選時回 -1。
//
// prompt 是要問的話;傳空字串代表這個呼叫端不印提示(原版有些路徑就是不印)。
//
// ⚠ 目前多人時取**最後一位**能動的隊員,與原版「印 Player: 讓玩家選」不同。
// 原版 `sub_E19C` 在只有一人能動時回的正是掃描過程留下的那一位,這裡沿用
// 同一個掃描;人數 >1 的分支要等 PromptCharacter 選單接上才會一致。
// 這是**已知的落差**,不要當成完成。
func (s *State) pickCharacter(prompt string) int {
	if prompt != "" {
		s.Log(prompt)
	}
	// 戰鬥中沒有「挑人」這回事 —— **輪到誰就是誰**。
	//
	// 原版把「目前是哪個角色」記在 `byte_3E08B`,而 `sub_A360` 進來就
	// 把它設成行動中的那個單位;所有指令讀的都是它。這裡回行動者的名冊索引。
	if m := s.actingMember(); m >= 0 {
		return m
	}
	last, n := -1, 0
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		switch s.Roster[i].Status {
		case u5data.StatusGood, u5data.StatusPoisoned:
			last, n = i, n+1
		}
	}
	if n == 0 {
		return -1
	}
	return last
}

// damageMember 在戰鬥之外扣一名隊員的血(原版 `sub_2A464`)。
//
// 與 `applyDamage` 的差別:那一支是戰鬥用的,要算抗性、發經驗值、通知 AI。
// 這一支是「地圖上的東西讓你掉血」,原版就只有減血與判死。
func (s *State) damageMember(i, dmg int) {
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
