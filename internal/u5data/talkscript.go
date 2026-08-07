package u5data

import "strings"

// 對話腳本
//
// `.TLK` 的記錄不只是文字,是一段**腳本**:原始位元組 0x81–0x9F 與 0xFF 是指令,
// 由原版的 `sub_1C3F8`(一張 31 路的跳表)逐位元組解釋。把它們當文字印出來,
// 畫面上就會出現一堆控制字元;把它們丟掉,則會失去加入隊伍、業報增減、
// 呼叫衛兵這些**實際會改變遊戲狀態**的效果。
//
// 記錄的排法(展開 48 筆 TOWNE.TLK 對照得出,並與 sub_1C840 的讀取方式一致):
//
//	段 0..4   名字 / 外貌 / 招呼 / 職業 / 道別
//	段 5 之後 成對的(關鍵字, 回應),關鍵字最多 4 個字元
//
// 「回應只有一個 0x87」代表**同下一則** —— 例如占星師的 `tele`(telescope)與
// `star` 共用同一段回答。原版 0x87 的實作是把文字指標存起來、往下讀一則再還原。
const (
	// OpAvatarName 插入聖者的名字(sub_1BB3C 逐字印出 byte_3DDB4)。
	OpAvatarName = 0x81
	// OpEndResponse 結束這一則回應。
	OpEndResponse = 0x82
	// OpPause 停頓(原版跑 0x1C 次的等待迴圈,可按鍵略過)。
	OpPause = 0x83
	// OpJoinParty 邀請加入隊伍(sub_1BB5C,滿員時回「Thou hast no room for me…」)。
	OpJoinParty = 0x84
	// OpSameAsNext 表示「這個關鍵字的回答與下一則相同」。
	OpSameAsNext = 0x87
	// OpKarmaUp / OpKarmaDown 增減業報(byte_3E098,上限 99)。
	OpKarmaUp   = 0x89
	OpKarmaDown = 0x8A
	// OpCallGuards 叫衛兵(sub_C10 掃 32 個 NPC 槽找 tile 0x70 那批)。
	OpCallGuards = 0x8B
	// OpNewline 換行(原版把字面的 0x8D 轉成 0x8A)。
	OpNewline = 0x8D
	// OpToggleEmphasis 切換強調(byte_55F1A ^= 0x80,影響後續字元的輸出屬性)。
	OpToggleEmphasis = 0x8E
	// OpWaitKey 等玩家按鍵。
	OpWaitKey = 0x8F
	// OpAskFirst..OpAskLast 是 15 個「向玩家提問並讀取回答」的指令(sub_1C0AC)。
	OpAskFirst = 0x91
	OpAskLast  = 0x9F
	// OpEndConversation 結束整段對話。
	OpEndConversation = 0xFF

	// KeywordLen 是關鍵字比對的長度。Ultima 系列一貫只看前 4 個字母。
	KeywordLen = 4
	// KarmaMax 是業報上限(原版 sub_2BBB8 的第三個參數 0x63)。
	KarmaMax = 99
)

// Effects 是一則回應除了文字之外造成的影響。
type Effects struct {
	JoinParty   bool // 對方要求加入隊伍
	CallGuards  bool // 叫衛兵
	KarmaDelta  int  // 業報增減
	EndTalk     bool // 結束整段對話
	AsksPlayer  bool // 中途向玩家提問(0x91–0x9F);問答流程尚未實作
	SameAsNext  bool // 這一則只是「同下一則」的別名
	WaitForKey  bool // 有停頓 / 等待按鍵
	HasEmphasis bool // 用到強調切換
}

// Entry 是一組「關鍵字 → 回應」。
type Entry struct {
	Keyword string // 已正規化成小寫、最多 KeywordLen 個字元
	Raw     []byte // 回應的原始位元組(尚未展開)
}

// Conversation 是一筆 .TLK 記錄解析後的樣子。
type Conversation struct {
	ID          int // .TLK 索引表裡的 id(= NPC 的對話號碼)
	Name        string
	Description string
	Greeting    string
	Job         string
	Bye         string
	Entries     []Entry

	dict *Dictionary
}

// ParseConversation 把一筆記錄解析成對話。
func ParseConversation(r *TalkRecord, d *Dictionary) *Conversation {
	segs := r.Segments()
	c := &Conversation{ID: r.NPCIndex, dict: d}
	get := func(i int) string {
		if i < len(segs) {
			return cleanText(d.Expand(segs[i]))
		}
		return ""
	}
	c.Name = r.Name(d)
	c.Description = get(TalkFieldDescription)
	c.Greeting = get(TalkFieldGreeting)
	c.Job = get(TalkFieldJob)
	c.Bye = get(TalkFieldBye)

	// 段 5 之後成對:關鍵字、回應。
	for i := TalkFixedFields; i+1 < len(segs); i += 2 {
		kw := normalizeKeyword(d.Expand(segs[i]))
		if kw == "" {
			continue
		}
		c.Entries = append(c.Entries, Entry{Keyword: kw, Raw: segs[i+1]})
	}
	return c
}

// normalizeKeyword 把關鍵字正規化:去控制碼、小寫、**先截到 KeywordLen 再去尾空白**。
//
// 順序不能顛倒。記錄裡有不少多字關鍵字(`art thou`、`who art`、`how many`),
// 截到 4 個字元剛好切在空白上變成 `"art "`;若先 trim 再截,存進去的是 `"art "`、
// 查詢時算出來的是 `"art"`,同一個字自己對不上自己 ——
// 症狀是 NPC 明明列了這個關鍵字卻回「聽不懂」。
//
// 截斷後去尾空白也順帶讓玩家打 `art` 就能命中 `art thou`。原版是硬比 4 個位元組,
// 這樣做比它寬鬆一點;寬鬆的方向對玩家有利,而且不會誤命中別的關鍵字
// (前 4 字元相同的本來就是同一個)。
func normalizeKeyword(s string) string {
	s = strings.ToLower(cleanText(s))
	if len(s) > KeywordLen {
		s = s[:KeywordLen]
	}
	return strings.TrimRight(s, " ")
}

// Respond 回答玩家輸入的一個字。
//
// 比對方式與原版一致:雙方都截到前 4 個字母。找不到就回傳 ok=false,
// 由呼叫端決定要說什麼(原版是「I cannot help thee with that.」之類)。
func (c *Conversation) Respond(input string) (text string, fx Effects, ok bool) {
	kw := normalizeKeyword(input)
	if kw == "" {
		return "", Effects{}, false
	}
	for i := range c.Entries {
		if c.Entries[i].Keyword != kw {
			continue
		}
		// 「同下一則」可能連續好幾層,所以要一直往下找到真正有文字的那一則。
		for j := i; j < len(c.Entries); j++ {
			t, f := c.render(c.Entries[j].Raw)
			if !f.SameAsNext {
				return t, f, true
			}
		}
		return "", Effects{}, true
	}
	return "", Effects{}, false
}

// render 展開一則回應,並收集它造成的影響。
func (c *Conversation) render(raw []byte) (string, Effects) {
	var b strings.Builder
	var fx Effects
	pendingSpace := false
	onlyOps := true

	for _, ch := range raw {
		switch {
		case ch == OpAvatarName:
			if pendingSpace {
				b.WriteByte(' ')
				pendingSpace = false
			}
			b.WriteString(AvatarNamePlaceholder)
			onlyOps = false
			continue
		case ch == OpSameAsNext:
			fx.SameAsNext = true
			continue
		case ch == OpEndResponse, ch == OpEndConversation:
			if ch == OpEndConversation {
				fx.EndTalk = true
			}
			continue
		case ch == OpJoinParty:
			fx.JoinParty = true
			continue
		case ch == OpCallGuards:
			fx.CallGuards = true
			continue
		case ch == OpKarmaUp:
			fx.KarmaDelta++
			continue
		case ch == OpKarmaDown:
			fx.KarmaDelta--
			continue
		case ch == OpNewline:
			b.WriteByte('\n')
			pendingSpace = false
			onlyOps = false
			continue
		case ch == OpPause, ch == OpWaitKey:
			fx.WaitForKey = true
			continue
		case ch == OpToggleEmphasis:
			fx.HasEmphasis = true
			continue
		case ch >= OpAskFirst && ch <= OpAskLast:
			fx.AsksPlayer = true
			continue
		case ch >= DictLiteralMin:
			// 字面文字。0x8D 已在上面處理掉,這裡剩下一般字元。
			if pendingSpace {
				b.WriteByte(' ')
				pendingSpace = false
			}
			b.WriteByte(ch & 0x7F)
			onlyOps = false
			continue
		}
		// 詞典 token
		w := c.dict.Word(ch)
		if w == "" {
			continue
		}
		b.WriteByte(' ')
		b.WriteString(w)
		pendingSpace = true
		onlyOps = false
	}
	// 整則只有指令、沒有半個字 → 它是別名(0x87 那種)。
	if onlyOps && !fx.SameAsNext && strings.TrimSpace(b.String()) == "" {
		fx.SameAsNext = true
	}
	return strings.TrimSpace(b.String()), fx
}

// AvatarNamePlaceholder 是聖者名字的佔位符。
//
// 原版 0x81 會把玩家角色名逐字印出來;引擎目前還沒有建角流程(存檔格式未解),
// 所以先留一個看得出來的佔位符,而不是印空字串 —— 空字串會讓句子少一個主詞,
// 看起來像翻譯漏了。
const AvatarNamePlaceholder = "汝"

// Keywords 回傳這個 NPC 認得的所有關鍵字(去重、依出現順序)。
func (c *Conversation) Keywords() []string {
	seen := map[string]bool{}
	var out []string
	for _, e := range c.Entries {
		if seen[e.Keyword] {
			continue
		}
		seen[e.Keyword] = true
		out = append(out, e.Keyword)
	}
	return out
}
