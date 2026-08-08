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
// ⚠ **更正一條長期的誤記**:此前這裡寫「原版用字母選(「Item: 」後面按一個
// 字母)」—— 那是錯的。把原版的清單瀏覽器 `sub_1EFC8` 整支掃過,它比對過的
// 鍵碼只有 **1..4(方向)、0xD3..0xD6(翻頁)、13(Enter)、0x20(空白)、
// 0x1B(ESC)**,**沒有任何一個字母鍵**(0x41..0x5A 一次都沒出現)。
// `Item: ` 是「目前停在哪一項」的標籤,不是「請按字母」的提示。
// ⇒ 引擎的方向鍵 + Enter **本來就是照原版**。見 `docs/re/60`。
//
// 唯一漏掉的是**翻頁鍵一次移 7 項**(原版 0xD5 / 0xD6 把移動次數設成 7),
// 已補上 `PickPage`。

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
	// Cursor 是游標停在哪一項(整份清單的絕對索引)。
	Cursor int
	// Top 是視窗最上面那一項 —— 原版 `sub_1EFC8` 的 `arg_0`。
	//
	// 原版只畫 **7 列**(文字列 2..8,畫到第 9 列就 `break`),
	// 游標列 `j`(1..7)與視窗頂端是兩個獨立的量。少了 Top 就無法重現
	// 「游標黏在中間、視窗自己捲」那個行為。
	Top int
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
	// Top / Cursor 都從 0 開始 —— 原版每次進瀏覽器也是從第一個候選起算。
	s.Prompt = PromptPick
	return true
}

// 清單瀏覽器的捲動規則 —— 逐條照 `sub_1EFC8`(`docs/re/60` 追記)
//
// 原版不是「一次列完、游標繞回」,而是一個**七列的視窗**加一個游標列:
//
//	視窗頂端  a1     ← Picker.Top
//	游標列    j      ← 1..7,= Cursor − Top + 1
//
// 而移動的規則有一個很有個性的地方:**游標黏在第 4 列(正中間)**,
// 到那裡之後改成捲視窗;只有靠近頭尾捲不動了,游標才自己在視窗裡走。
// 原版的上移迴圈長這樣(下移對稱,多一個 `j >= 已畫列數` 的條件):
//
//	for 步數 {
//	    if j == 4 || j == 1 {
//	        if 上面還有東西 { 視窗往上一項; continue }
//	        if j != 4      { continue }          // j==1 且到頂 → 這一步空轉
//	    }
//	    j--                                     // 游標在視窗裡往上
//	}
//
// ⚠ **原版不繞回**:到頂 / 到底時那一步是空轉(`sub_1E3D8` / `sub_1E418`
// 回 −1),游標留在原地。此前引擎的 `PickMove` 會繞回 —— 那是我自己加的,
// 而 `TestPickMoveWraps` 每次都綠,因為它量的是我自己的發明。

// PickRows 是視窗畫幾列(原版畫到文字列 9 就 `break`,起點是列 2 → 7 列)。
const PickRows = 7

// PickCenterRow 是游標會黏住的那一列(原版寫死 `j == 4`)。
const PickCenterRow = 4

// PickPageRows 是翻頁鍵一次移幾項(原版 `sub_1EFC8` 對 0xD5 / 0xD6 把
// 移動次數設成 **7**,不是「一整頁」)。
const PickPageRows = PickRows

// rows 是這一幀實際畫出來幾列(清單比視窗短的時候會少於 PickRows)。
func (p *Picker) rows() int {
	n := len(p.Entries) - p.Top
	if n > PickRows {
		return PickRows
	}
	return n
}

// row 是游標所在的視窗列號(原版的 `j`,1..7)。
func (p *Picker) row() int { return p.Cursor - p.Top + 1 }

// stepUp 是往上一步(原版 0xD5 分支的迴圈本體,鍵 1 / 3 走同一段)。
func (p *Picker) stepUp() {
	if j := p.row(); j == PickCenterRow || j == 1 {
		if p.Top > 0 { // 上面還有東西 → 捲視窗,游標列不變
			p.Top--
			p.Cursor--
			return
		}
		if j != PickCenterRow { // j == 1 且到頂 → 這一步空轉
			return
		}
	}
	p.Cursor-- // 游標在視窗裡往上
}

// stepDown 是往下一步(原版 0xD6 分支)。
func (p *Picker) stepDown() {
	rows := p.rows()
	if j := p.row(); j == PickCenterRow || j == PickRows || j >= rows {
		if p.Top+rows < len(p.Entries) { // 下面還有東西 → 捲視窗
			p.Top++
			p.Cursor++
			return
		}
		// 捲不動了。只有「游標剛好在正中間、而且清單長過中間」才讓游標自己走。
		if j != PickCenterRow || rows <= PickCenterRow {
			return
		}
	}
	p.Cursor++
}

// PickMove 移動游標。**不繞回** —— 到頭到尾就停住(原版行為)。
func (s *State) PickMove(delta int) {
	if s.Pick == nil || len(s.Pick.Entries) == 0 {
		return
	}
	p := s.Pick
	for n := delta; n > 0; n-- {
		p.stepDown()
	}
	for n := delta; n < 0; n++ {
		p.stepUp()
	}
}

// PickPage 翻一頁:往 dir 的方向走 PickPageRows 步。
//
// 注意是「走 7 步」而不是「索引加 7」—— 因為每一步可能捲視窗、可能移游標,
// 兩者對索引的影響一樣,但對 `Top` 不一樣。
func (s *State) PickPage(dir int) {
	if dir < 0 {
		s.PickMove(-PickPageRows)
		return
	}
	s.PickMove(PickPageRows)
}

// PickHome 跳到第一項(原版鍵碼 **0xD3**:`sub_1E418(-1, …)` 從頭找第一個候選)。
func (s *State) PickHome() {
	if s.Pick == nil {
		return
	}
	s.Pick.Top, s.Pick.Cursor = 0, 0
}

// PickEnd 跳到最後一項(原版鍵碼 **0xD4**:先 `sub_1E3D8` 找最後一項,
// 再往前走最多 6 步把視窗頂端定出來 —— 也就是「最後一頁,游標在最後一項」)。
func (s *State) PickEnd() {
	if s.Pick == nil || len(s.Pick.Entries) == 0 {
		return
	}
	p := s.Pick
	p.Cursor = len(p.Entries) - 1
	p.Top = p.Cursor - (PickRows - 1)
	if p.Top < 0 {
		p.Top = 0
	}
}

// PickScrollHint 回報原版畫在清單旁邊的捲動指示字元。
//
// 原版用 CP437 的三個箭頭(`sub_29008` 的參數就是字元碼):
//
//	24 → ↑   上面還有(視窗頂端不是第一項)
//	25 → ↓   下面還有
//	18 → ↕   兩邊都有
//	（都沒有時走 `sub_28F80` 清掉,不畫）
//
// 回傳空字串代表不畫。
func (s *State) PickScrollHint() string {
	p := s.Pick
	if p == nil {
		return ""
	}
	up := p.Top > 0
	down := p.Top+p.rows() < len(p.Entries)
	switch {
	case up && down:
		return "\u2195" // ↕ CP437 18
	case up:
		return "\u2191" // ↑ CP437 24
	case down:
		return "\u2193" // ↓ CP437 25
	}
	return ""
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
	out := make([]string, 0, PickRows+2)
	title := p.Title
	// 捲動指示跟在標題後面 —— 原版畫在清單旁邊,這裡受 CJK 版面限制擺在標題列。
	if hint := s.PickScrollHint(); hint != "" {
		title += " " + hint
	}
	out = append(out, title)
	// ★ 只畫視窗裡的七列,不是整份清單(原版 `sub_1EFC8` 畫到文字列 9 就 break)。
	for i := p.Top; i < p.Top+p.rows(); i++ {
		e := p.Entries[i]
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
	// ★ 順序照 `sub_1E8D4` 抄清單的順序:卷軸 → 藥水 → 特殊道具 → 月石 → 其餘。
	// 卷軸與藥水列的是**數量**(原版那兩段抄的是持有數,不是旗標)。
	for i := 0; i < u5data.ScrollCount; i++ {
		if n := s.Inventory.Scrolls[i]; n > 0 {
			out = append(out, PickEntry{
				Label: MsgScroll + u5data.ScrollSpell(i) + " ×" + itoa(n),
				Value: UseScrollFirst + i,
			})
		}
	}
	for i := 0; i < u5data.PotionCount; i++ {
		if n := s.Inventory.Potions[i]; n > 0 {
			out = append(out, PickEntry{
				Label: u5data.PotionColoursZH[i] + "藥水 ×" + itoa(n),
				Value: UsePotionFirst + i,
			})
		}
	}
	add(s.Inventory.Carpets > 0, MsgItemCarpet, UseCarpet)
	add(s.Inventory.OddKeys > 0, MsgItemSkullKey, UseSkullKey)
	add(s.Regalia.Amulet, MsgItemAmulet, UseAmulet)
	add(s.Regalia.Crown, MsgItemCrown, UseCrown)
	add(s.Regalia.Sceptre, MsgItemSceptre, UseSceptre)
	// 月石 —— 只列還在手上的(原版 `byte_3E050[i] == 0xFF`)。
	for i := 0; i < u5data.MoonstoneCount; i++ {
		add(s.Inventory.Moonstones[i].InHand(), MsgItemMoonstone, UseMoonstoneFirst+i)
	}
	for i := 0; i < u5data.ShadowlordCount; i++ {
		add(s.Shards[i], MsgItemShard, UseShardFirst+i)
	}
	add(s.Regalia.Plans, MsgItemPlans, UsePlans)
	add(s.HasBadge, MsgItemBadge, UseBadge)
	add(s.SandalwoodBox, MsgItemWoodenBox, UseWoodenBox)
	// ✅ 望遠鏡 / 六分儀 / 懷錶的存檔位移**已釘死**(`docs/re/79`),
	// 所以改成照旗標列 —— 此前是「位移沒對出來所以無條件列出」。
	add(s.HasSpyglass, MsgItemSpyglass, UseSpyglass)
	add(s.HasSextant, MsgItemSextant, UseSextant)
	add(s.HasWatch, MsgItemWatch, UseWatch)
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
