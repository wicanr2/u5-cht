package game

import (
	"fmt"

	"github.com/wicanr2/u5-cht/internal/i18n"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 選單 —— 讓 R / W / N / U / M 這幾個指令按得到
//
// 這五支的規則早就實作了,但它們的簽章是 `Ready(member, item)` 這種 API,
// 鍵盤上按不到 —— **做完了卻用不到,等於沒做**。
//
// 原版每一支各有自己的選單畫面(`sub_1F3A4` 印「Item: 」、`sub_1E418` 印
// 「Items:」、`sub_18468` 印「Reagents:」),但形狀都一樣:列一排候選、
// 挑一個、可以放棄。所以這裡收成**一個**選單,由呼叫端給標題與候選。
//
// ⚠ 原版用**字母**選(「Item: 」後面按一個字母)。引擎用方向鍵 + Enter ——
// 與主選單、Ztats 一致。字母對應哪一項在跳表裡看得出來有,但沒逐一核過,
// 所以先不做(`docs/re/45` §5 同一條原則:沒證據不猜)。

// PickEntry 是選單上的一項。
type PickEntry struct {
	// Label 是顯示的文字。
	Label string
	// Value 是選中之後回傳給呼叫端的值(裝備編號、名冊索引、咒語編號…)。
	Value int
}

// Picker 是進行中的選單。
type Picker struct {
	Title   string
	Entries []PickEntry
	Cursor  int
	// Marks 是複選模式下每一項的勾選狀態;nil 代表單選。
	//
	// 只有調藥的藥草清單是複選(原版 `sub_18468`):空白鍵或 Enter 勾 / 取消勾,
	// 按 **M** 才確定。單選的選單 Enter 就是確定 —— 兩種模式的 Enter 語意不同,
	// 所以要靠這個欄位分。
	Marks []bool
	// confirm 是複選模式按下 M 之後要做的事,參數是勾選遮罩。
	confirm func(mask byte) bool
	// then 是選定之後要做的事。回傳 true 代表流程結束(選單關掉);
	// false 代表**還有下一步**,由 then 自己接著開下一個選單。
	then func(value int) bool
	// back 是放棄時要回到哪個 Prompt。
	back Prompt
}

// beginPick 打開一個選單。候選是空的就直接回報 empty 那句話。
func (s *State) beginPick(title string, entries []PickEntry, empty string, then func(int) bool) bool {
	if len(entries) == 0 {
		s.Log(empty)
		return false
	}
	back := s.Prompt
	if back == PromptPick {
		back = PromptNone // 選單接選單時不要回到自己
	}
	s.Pick = &Picker{Title: title, Entries: entries, then: then, back: back}
	s.Prompt = PromptPick
	return true
}

// PickMove 移動游標,會繞回。
func (s *State) PickMove(delta int) {
	if s.Pick == nil {
		return
	}
	n := len(s.Pick.Entries)
	s.Pick.Cursor = ((s.Pick.Cursor+delta)%n + n) % n
}

// PickChoose 按下 Enter。
func (s *State) PickChoose() {
	p := s.Pick
	if p == nil || p.Cursor >= len(p.Entries) {
		return
	}
	value, then, back := p.Entries[p.Cursor].Value, p.then, p.back
	s.Pick, s.Prompt = nil, back
	if then != nil && !then(value) {
		return // then 自己開了下一個選單,或者只是印了一句話
	}
}

// PickCancel 放棄(ESC)。
func (s *State) PickCancel() {
	if s.Pick == nil {
		return
	}
	s.Prompt = s.Pick.back
	s.Pick = nil
	s.Log(MsgNevermind)
}

// PickLines 是選單現在該顯示的每一行。游標那一項前面加箭頭。
func (s *State) PickLines() []string {
	p := s.Pick
	if p == nil {
		return nil
	}
	out := make([]string, 0, len(p.Entries)+1)
	out = append(out, p.Title)
	for i, e := range p.Entries {
		mark := "  "
		if i == p.Cursor {
			mark = "→"
		}
		// 複選模式多一個勾記 —— 原版在數量與名字之間印一個特殊字元(0x0F)。
		sel := ""
		if p.Marks != nil {
			sel = " "
			if p.Marks[i] {
				sel = "*"
			}
		}
		out = append(out, mark+sel+e.Label)
	}
	return out
}

// PickToggle 是複選模式下的空白鍵 / Enter:勾或取消勾游標那一項。
func (s *State) PickToggle() {
	p := s.Pick
	if p == nil || p.Marks == nil || p.Cursor >= len(p.Marks) {
		return
	}
	p.Marks[p.Cursor] = !p.Marks[p.Cursor]
}

// PickConfirm 是複選模式下的 M:把勾選遮罩交出去。
func (s *State) PickConfirm() {
	p := s.Pick
	if p == nil || p.Marks == nil {
		return
	}
	var mask byte
	for i, on := range p.Marks {
		if on {
			mask |= u5data.ReagentBit(p.Entries[i].Value)
		}
	}
	confirm, back := p.confirm, p.back
	s.Pick, s.Prompt = nil, back
	if confirm != nil {
		confirm(mask)
	}
}

// PickMulti 回報現在的選單是不是複選模式。
func (s *State) PickMulti() bool { return s.Pick != nil && s.Pick.Marks != nil }

// ---------------------------------------------------------------- 各指令的入口

// partyEntries 是隊伍的候選清單。skip 那一位不列(換位時不能跟自己換)。
func (s *State) partyEntries(skip int) []PickEntry {
	var out []PickEntry
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		if i == skip {
			continue
		}
		out = append(out, PickEntry{Label: s.Roster[i].Name, Value: i})
	}
	return out
}

// equipEntries 把裝備編號變成候選,標上持有數量。
func (s *State) equipEntries(items []byte) []PickEntry {
	out := make([]PickEntry, 0, len(items))
	for _, it := range items {
		out = append(out, PickEntry{
			Label: fmt.Sprintf("%s ×%d", s.equipName(int(it)), s.Inventory.Items[it]),
			Value: int(it),
		})
	}
	return out
}

// BeginReady 是 R 指令:先挑人,再挑裝備。
//
// ⚠ 原版**沒有 W(Wear)** —— 六個欄位全部走這一支,清單是 48 件裝備全列。
// 舊版拆成 R / W 兩支是把 U4 的模型搬過來,已更正(`docs/re/49`)。
func (s *State) BeginReady() bool {
	return s.beginPick(MsgPickWho, s.partyEntries(-1), MsgNobodyHere, func(member int) bool {
		return !s.beginPick(MsgPickItem, s.equipEntries(s.ReadyList()), MsgNoUsableItems,
			func(item int) bool { s.Ready(member, byte(item)); return true })
	})
}

// BeginNewOrder 是 N 指令:挑兩個人交換。
//
// ⚠ 第一份清單**不能排除聖者** —— 原版是讓玩家選得到、然後才說
// 「<名字> must lead!」。從清單裡藏起來的話玩家不會知道那是規則,
// 只會以為選單壞了。
func (s *State) BeginNewOrder() bool {
	return s.beginPick(MsgPickSwap, s.partyEntries(-1), MsgNobodyHere, func(a int) bool {
		if a == 0 {
			s.Log(s.Roster[0].Name + MsgMustLead)
			return true
		}
		return !s.beginPick(MsgPickWith, s.partyEntries(a), MsgNobodyHere,
			func(b int) bool { s.NewOrder(a, b); return true })
	})
}

// BeginUse 是 U 指令:挑一件特殊道具。
func (s *State) BeginUse() bool {
	return s.beginPick(MsgPickItem, s.usableEntries(), MsgNoUsableItems,
		func(item int) bool { s.Use(item); return true })
}

// usableEntries 是身上有的特殊道具(原版 `sub_1E8D4` 那份清單)。
//
// 名字走**原版的短名字表**(`docs/re/56`):索引 0 = `Magic Crpt` = case 16,
// 由 `i18n.Name` 翻成中文,翻不到就照原樣顯示英文短名。
// 這樣清單的內容與順序來自玩家自己的檔案,而不是我在程式裡打的一串字。
func (s *State) usableEntries() []PickEntry {
	var out []PickEntry
	// label 先查短名字表 + i18n,查不到才退回寫死的中文。
	label := func(code int, fallback string) string {
		en := s.SpecialItems.NameForUseCode(code)
		if en == "" || u5data.SpecialItemPlaceholder(en) {
			return fallback
		}
		if zh := i18n.Name(en); zh != "" && zh != en {
			return zh
		}
		return en
	}
	add := func(have bool, name string, item int) {
		if have {
			out = append(out, PickEntry{Label: label(item, name), Value: item})
		}
	}
	add(s.Inventory.Carpets > 0, MsgItemCarpet, UseCarpet)
	add(s.Inventory.OddKeys > 0, MsgItemSkullKey, UseSkullKey)
	add(s.Regalia.Amulet, MsgItemAmulet, UseAmulet)
	add(s.Regalia.Crown, MsgItemCrown, UseCrown)
	add(s.Regalia.Sceptre, MsgItemSceptre, UseSceptre)
	for i := 0; i < u5data.ShadowlordCount; i++ {
		add(s.Shards[i], MsgItemShard, UseShardFirst+i)
	}
	add(s.Regalia.Plans, MsgItemPlans, UsePlans)
	add(s.HasBadge, MsgItemBadge, UseBadge)
	add(s.SandalwoodBox, MsgItemWoodenBox, UseWoodenBox)
	// ⚠ 望遠鏡 / 六分儀 / 懷錶的持有旗標存檔位移還沒釘死(`docs/re/44` §4),
	// 所以**先無條件列出來**。列了但沒有比不列好:效果本身是照原版做的,
	// 缺的只是「有沒有」那一格 —— 而藏起來會讓玩家以為這幾樣沒實作。
	out = append(out,
		PickEntry{Label: label(UseSpyglass, MsgItemSpyglass), Value: UseSpyglass},
		PickEntry{Label: label(UseSextant, MsgItemSextant), Value: UseSextant},
		PickEntry{Label: label(UseWatch, MsgItemWatch), Value: UseWatch})
	return out
}

// BeginMix 是 M 指令(原版 `sub_18704` → `sub_1CA0C` → `sub_18468` → `sub_18698`)。
//
// ⚠ **這裡原本是捷徑**:列一張咒語選單、照配方自動調。那把原版最大的張力
// 拿掉了 —— 原版是你**自己挑藥草**,挑錯就白費。現在照原版的四步走:
//
//	1. 身上一種藥草都沒有 → 「無藥草!」
//	2. 「為哪個咒語?」→ 打符文首字母(與施法**同一個輸入法**)
//	   ⚠ 只有「什麼都沒打」會中止;**湊不出咒語(−2)照樣往下走** ——
//	   原版是 `inc eax; jnz`,只擋 −1。所以亂打符文會讓你把藥草調成廢渣。
//	3. 藥草清單(只列身上有的)勾選:方向鍵移動、空白 / Enter 勾、**M 確定**、ESC 放棄
//	4. 「要幾份?」→ 兩位數;不夠就印「藥草不足!」**重問一次**
//
// 挑到的藥草一定會被扣掉,遮罩要與配方**完全相符**才調得出咒語。
func (s *State) BeginMix() bool {
	if !s.hasAnyReagent() {
		s.Log(MsgNoReagents)
		return false
	}
	if s.Spells == nil || s.Runes == nil {
		return false
	}
	s.beginRunePrompt(MsgForWhatSpell, func(spell int) {
		if spell == u5data.SpellInputCancelled {
			s.Log(MsgSpellNone)
			return
		}
		s.beginReagentPick(spell)
	})
	return true
}

// beginReagentPick 打開藥草勾選畫面(原版 `sub_18468`)。
//
// 只列**身上有的**藥草,每一列印出擁有數量;spell 可能是
// `u5data.SpellInputNoSpell`,那代表玩家打的符文湊不出咒語 ——
// 照原版仍然讓他挑、仍然扣藥草,只是永遠調不出東西。
func (s *State) beginReagentPick(spell int) bool {
	var out []PickEntry
	for r := 0; r < u5data.ReagentCount; r++ {
		n := s.Inventory.Reagents[r]
		if n < 1 {
			continue
		}
		out = append(out, PickEntry{
			Label: fmt.Sprintf("%02d %s", n, s.reagentName(r)),
			Value: r,
		})
	}
	if !s.beginPick(MsgReagents, out, MsgNoReagents, nil) {
		return false
	}
	s.Pick.Marks = make([]bool, len(out))
	s.Pick.confirm = func(mask byte) bool {
		s.askMixAmount(spell, mask)
		return true
	}
	return true
}

// askMixAmount 問份數並收尾(原版 `sub_18698` + `sub_18704` 的後半)。
//
// ⚠ 順序照原版:**先問份數,才檢查有沒有勾到東西**。所以一個都沒勾就按 M,
// 原版還是會問「要幾份?」,答完才說「沒東西可調!」。
func (s *State) askMixAmount(spell int, mask byte) {
	s.Log(MsgHowMuch)
	s.AskNumberDigits(2, u5data.MixAmountMax, func(n int) {
		if n <= 0 {
			return
		}
		// 藥草不夠 → 印一句然後**重問份數**,不是取消整個流程。
		for r := 0; r < u5data.ReagentCount; r++ {
			if mask&u5data.ReagentBit(r) == 0 {
				continue
			}
			if s.Inventory.Reagents[r] < n {
				s.Log(MsgInsufficientReagents)
				s.askMixAmount(spell, mask)
				return
			}
		}
		if mask == 0 {
			s.Log(MsgNothingToMix)
			return
		}
		var picked []int
		for r := 0; r < u5data.ReagentCount; r++ {
			if mask&u5data.ReagentBit(r) != 0 {
				picked = append(picked, r)
			}
		}
		s.Log(MsgMixing)
		s.Mix(spell, n, picked)
	})
}

// reagentName 是藥草的顯示名(有譯名就用譯名,沒有就照原樣顯示英文)。
func (s *State) reagentName(r int) string {
	if r < 0 || r >= len(u5data.ReagentNames) {
		return ""
	}
	en := u5data.ReagentNames[r]
	if zh := i18n.Name(en); zh != "" {
		return zh
	}
	return en
}

// hasAnyReagent 回報身上有沒有任何藥草。
//
// (原本放在 fire.go,那份 Mix 重複實作被移除時一起沒了 —— 搬過來。)
func (s *State) hasAnyReagent() bool {
	for _, n := range s.Inventory.Reagents {
		if n > 0 {
			return true
		}
	}
	return false
}
