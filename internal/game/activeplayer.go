package game

import "github.com/wicanr2/u5-cht/internal/u5data"

// 「Set Active Plr」—— 數字鍵指定接下來由誰動手(原版 `sub_2BD40`)
//
// 推導見 `docs/re/97`。這是 `pickCharacter`(`sub_E19C`)第一條分支要的那個狀態:
//
//	byte_3E08B != 0xFF  →  接下來的指令都由那個人做,不再問
//
// ★ 它是**跨指令**的,而且**進存檔**(`sub_27D24` 讀 / `sub_284CC` 寫)。
// 許多指令用完會把它清回 0xFF —— 見 `docs/re/95` §3 的全檔掃描。

// ActiveNone 是「沒有指定」(原版 `0xFF`)。
const ActiveNone = 0xFF

// 數字鍵的兩端(原版 `cmp edi, 30h` / `cmp edi, 39h`)。
//
// ⚠ 收的是 **'0'..'9' 十個鍵**,不是 '0'..'6'。'7'..'9' 會算出
// 超出隊伍人數的索引 → 印「Invalid!」。照抄:少收三個鍵的話,
// 玩家按 8 會落到別的指令去。
const (
	ActiveKeyNone  = '0'
	ActiveKeyFirst = '1'
	ActiveKeyLast  = '9'
)

// ActiveMember 回報現在指定了誰;沒指定時回 -1。
func (s *State) ActiveMember() int {
	// ⚠ 兩個條件都要 —— `activeSet` 擋住「結構常值的零值」,
	// `!= ActiveNone` 擋住「明確清掉」。少一個就會有一種情況判錯。
	if !s.activeSet || s.activeMember == ActiveNone {
		return -1
	}
	return int(s.activeMember)
}

// ClearActiveMember 把指定清掉(原版那一大批 `mov byte_3E08B, 0FFh`)。
func (s *State) ClearActiveMember() { s.activeMember, s.activeSet = ActiveNone, true }

// SetActivePlayer 是數字鍵指令(原版 `sub_2BD40`)。
//
//	'0'      → 「無!」,清掉指定
//	'1'..'9' → 指定第 (鍵 − '1') 位;超出人數、死了、睡著了都印「無效!」
//
// 回傳 false 表示這個鍵不是它管的(呼叫端該往下傳)。
//
// ⚠⚠ **死了或睡著了不能被指定** —— 而「中毒」可以。
// `pickCharacter` 的自動掃描也是把 'G' 與 'P' 都算能動,兩處一致。
func (s *State) SetActivePlayer(key rune) bool {
	if key < ActiveKeyNone || key > ActiveKeyLast {
		return false
	}
	s.Log(MsgSetActivePlayer)
	if key == ActiveKeyNone {
		s.ClearActiveMember()
		s.Log(MsgActiveNone)
		return true
	}
	i := int(key - ActiveKeyFirst)
	if i >= s.PartySize || i >= len(s.Roster) {
		s.Log(MsgActiveInvalid)
		return true
	}
	switch s.Roster[i].Status {
	case u5data.StatusDead, u5data.StatusAsleep:
		s.Log(MsgActiveInvalid)
		return true
	}
	s.activeMember, s.activeSet = byte(i), true
	s.Log(s.Roster[i].Name)
	return true
}

// activeIfUsable 回報指定的那位還能不能動手。
//
// ⚠ 指定之後狀態**可能變**(被打死、被催眠),而原版 `sub_E19C` 只檢查
// `byte_3E08B != 0xFF` 就直接用 —— **不重驗狀態**。照抄:
// 死人被指定著的時候,那些指令真的會落在死人頭上(原版的行為)。
func (s *State) activeIfUsable() int {
	i := s.ActiveMember()
	if i < 0 || i >= s.PartySize || i >= len(s.Roster) {
		return -1
	}
	return i
}
