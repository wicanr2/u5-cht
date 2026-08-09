package game

import "github.com/wicanr2/u5-cht/internal/u5data"

// 「Player: 」—— 多人可選時真的問(原版 `sub_E19C` 第三條 + `sub_2A7F4` 選單)
//
// 推導見 `docs/re/97` §3b 與 `docs/re/98`。原版的三條路:
//
//	戰鬥中             → 當前行動的單位,不問
//	已指定(`byte_3E08B`)→ 那個人,不問
//	否則               → 掃「'G' 或 'P'」;**只有一個就不問**,2 人以上才開選單
//
// ★★ 選單的兩個關鍵行為(少一個手感就不對):
//
//  1. **游標可以停在死人身上** —— 不是跳過他們。按下去才印「無法行動!」
//     然後**回到選單**(原版 `aDisabled` + `jmp loc_E20E` 繞回去)。
//  2. **數字鍵直接跳**('1'..'6'),而方向鍵是另一組(鍵碼 1..4,在 '0' 之下)——
//     兩者不衝突,所以原版兩種操作並存。
//
// ⚠ `then` **一定會被呼叫一次**:不必問時同步呼叫,要問時在玩家選完(或取消)
// 之後呼叫。取消 → `then(-1)`。呼叫端的後續步驟一律寫在 `then` 裡面,
// **不要**寫在 `pickMember` 之後 —— 那一段在「要問」的路徑上會提前執行。

// pickMember 決定由誰動手,必要時開選單問。
func (s *State) pickMember(prompt string, then func(int)) {
	// ① 戰鬥中:輪到誰就是誰。
	if m := s.actingMember(); m >= 0 {
		then(m)
		return
	}
	// ② 玩家用數字鍵指定過(原版 `byte_3E08B != 0xFF`,見 `docs/re/97`)。
	if m := s.activeIfUsable(); m >= 0 {
		then(m)
		return
	}
	able := s.ableMembers()
	switch len(able) {
	case 0:
		then(-1)
		return
	case 1:
		// ★ 只有一個能動就**不問** —— 原版 `if (count <= 1) return esi`。
		then(able[0])
		return
	}
	// ③ 2 人以上 → 開選單。
	if prompt != "" {
		s.Log(prompt)
	}
	s.Log(MsgPlayerPrompt)
	s.beginPickMember(then)
}

// ableMembers 回傳「能動」的名冊索引(狀態 'G' 或 'P')。
//
// ⚠ **中毒算能動**。寫成「只有 'G' 才算」會讓中毒的隊員突然不能做事 ——
// 那不是原版,而且中毒在 U5 裡很常見。
func (s *State) ableMembers() []int {
	var out []int
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		switch s.Roster[i].Status {
		case u5data.StatusGood, u5data.StatusPoisoned:
			out = append(out, i)
		}
	}
	return out
}

// beginPickMember 開「Select:」選單(原版 `sub_2A7F4`)。
//
// ★ 候選是**全部隊員**不是只有能動的 —— 見檔頭第 1 點。
func (s *State) beginPickMember(then func(int)) {
	var entries []PickEntry
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		c := &s.Roster[i]
		label := c.Name
		// 不能動的在名字後面標狀態,讓玩家看得出來為什麼按下去會被拒。
		//
		// ⚠ 這是**引擎加的**:原版靠玩家自己記狀態(狀態列上有)。
		// 加了不改變任何規則,只是把資訊放在眼前。
		switch c.Status {
		case u5data.StatusGood, u5data.StatusPoisoned:
		default:
			label += "(" + u5data.StatusName(c.Status) + ")"
		}
		entries = append(entries, PickEntry{Label: label, Value: i})
	}
	if !s.beginPick(MsgSelect, entries, MsgNobodyCanAct, func(v int) bool {
		if v < 0 || v >= len(s.Roster) {
			then(-1)
			return true
		}
		switch s.Roster[v].Status {
		case u5data.StatusGood, u5data.StatusPoisoned:
			then(v)
			return true
		}
		// ★★ 原版:印「Disabled!」然後**重問**(不是取消)。
		s.Log(MsgDisabled)
		s.beginPickMember(then)
		return false
	}) {
		then(-1)
		return
	}
	// 取消 → then(-1),否則呼叫端的流程會靜靜停住。
	s.Pick.onCancel = func() { then(-1) }
	// 原版取消時印的是「None!」(同 `SetActivePlayer` 按 '0' 那句)。
	s.Pick.cancelMsg = MsgActiveNone
}

// 選人選單自己收的三種鍵(原版 `sub_2A7F4` 的按鍵分派)。
const (
	// PickMemberKeyNone 是「不選任何人」(原版 `'0'` → `sel = −1`)。
	PickMemberKeyNone = '0'
	// PickMemberKeyFirst / Last 是直接跳到第 n 位(原版 `'1'..'6'`)。
	//
	// ⚠ 上限是 **'6'**(隊伍最多六人),不是 `SetActivePlayer` 的 `'9'`。
	// 兩支收的鍵範圍不同 —— 那是原版的樣子,不要「統一」。
	PickMemberKeyFirst = '1'
	PickMemberKeyLast  = '6'
	// PickMemberKeyQuit 是空白鍵結束(原版 `0x20`,與 ESC 同一個出口)。
	PickMemberKeyQuit = ' '
)

// PickMemberDigit 是選人選單自己收的按鍵(原版 `sub_2A7F4`)。
//
//	'1'..'6' → 游標跳到那一位。⚠ **不檢查狀態** —— 只是移游標,
//	           按 Enter 才觸發「無法行事!」的判斷。
//	'0'、空白 → 結束,回呼收到 −1
//
// 只在**選人選單**裡有效。回傳有沒有吃掉這個鍵 —— 沒吃掉才該往下傳。
func (s *State) PickMemberDigit(r rune) bool {
	if !s.PickIsMember() {
		return false
	}
	p := s.Pick
	if r == PickMemberKeyNone || r == PickMemberKeyQuit {
		s.PickCancel()
		return true
	}
	if r < PickMemberKeyFirst || r > PickMemberKeyLast {
		return false
	}
	i := int(r - PickMemberKeyFirst)
	if i >= len(p.Entries) {
		return true // 原版:超出人數就什麼都不做,但**算吃掉了**
	}
	p.Cursor = i
	return true
}

// PickIsMember 回報現在開著的是不是**選人**選單。
//
// ⚠ 拿標題比對而不是加一個布林欄位:`Picker` 是共用的,而選人選單是
// 由 `beginPickMember` 唯一產生的 —— 標題就是它的身分。加欄位的話
// 每個 `beginPick` 呼叫端都得記得填,漏一個就變成「有時候能按數字鍵」。
func (s *State) PickIsMember() bool { return s.Pick != nil && s.Pick.Title == MsgSelect }
