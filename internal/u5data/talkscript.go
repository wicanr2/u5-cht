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
//	段 2i+5   第 i 個關鍵字      段 2i+6  它的回應
//
// 這個 2i+5 / 2i+6 的算式不是猜的:`sub_1BD50(i)` 把指標重設到開頭後跳 `2i+5` 段,
// 而命中之後 `sub_1BF08` 用 `sub_1BAFC(i*2 + 6)` 印回應。
//
// **關鍵字表在遇到位元組 0x90 時結束** —— 跳段用的 `sub_1BA80(0, 0x90)` 是
// 「前進到 NUL(成功)或 0x90(失敗)」。0x90 之後的段落是提問分支用的資料,
// 不是關鍵字對。少了這個終止條件會多解出幾組假的對,把回應接到錯的字上。
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

	// KeywordEndMarker 是關鍵字表的結束標記(sub_1BA80 的第二個停止位元組)。
	KeywordEndMarker = 0x90
	// JoinNameLen 是入隊時比對名字的長度。原版 sub_1BB5C 只讀 3 個位元組。
	JoinNameLen = 3
	// KarmaMax 是業報上限(原版 sub_2BBB8 的第三個參數 0x63)。
	KarmaMax = 99
)

// Effects 是一則回應除了文字之外造成的影響。
type Effects struct {
	JoinParty   bool // 對方要求加入隊伍
	CallGuards  bool // 叫衛兵
	KarmaDelta  int  // 業報增減
	EndTalk     bool // 結束整段對話
	AsksPlayer  bool // 中途向玩家提問(0x91–0x9F)
	AskCode     byte // 提問碼,用來找對應的提問區塊
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
	// Questions 是關鍵字表之後的提問區塊(見 Question)。
	Questions []Question

	// AvatarName 是玩家角色的名字,opcode 0x81 會插入它。
	// 留空時用 AvatarNamePlaceholder —— 印空字串會讓句子少一個主詞,
	// 看起來像翻譯漏了。
	AvatarName string

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

	// 段 2i+5 / 2i+6 成對,遇到含 0x90 的段就停;之後是提問區塊。
	//
	// end 一定要在迴圈外先算好:對數用完(i+1 越界)時迴圈是**正常結束**的,
	// 那時提問區塊就從 i 開始。只在 break 時設 end 的話,這種情況會漏掉整批提問。
	i := TalkFixedFields
	for ; i+1 < len(segs); i += 2 {
		if hasByte(segs[i], KeywordEndMarker) || hasByte(segs[i+1], KeywordEndMarker) {
			break
		}
		kw := normalizeKeyword(d.Expand(segs[i]))
		if kw == "" {
			continue
		}
		c.Entries = append(c.Entries, Entry{Keyword: kw, Raw: segs[i+1]})
	}
	c.parseQuestions(segs, i)
	return c
}

// normalizeKeyword 把記錄裡的關鍵字正規化:去控制碼、小寫、去頭尾空白。
//
// **不截斷**。原版不是「比前 4 個字母」,而是把關鍵字當**子字串**去玩家輸入裡找
// (見 MatchKeyword);記錄裡的關鍵字本來就長短不一(`bow`、`star`、`art thou`)。
func normalizeKeyword(s string) string {
	return strings.TrimSpace(strings.ToLower(cleanText(s)))
}

// MatchKeyword 判斷玩家輸入有沒有命中某個關鍵字。
//
// 原版的規則(`sub_1BD8C` → `sub_27C98`,以及內建關鍵字那一段同樣的判斷):
// **關鍵字要出現在玩家輸入裡,而且必須落在詞首** ——
// 命中位置是 0,或前一個字元是空白。
//
//	關鍵字 "bow"  ← 輸入 "bow" ✓  "bows" ✓  "my bow" ✓  "elbow" ✗
//
// 這與「雙方都截到 4 個字母再比對」不一樣:那樣 `bows` 會截成 `bows` 而配不上 `bow`。
// 短關鍵字在記錄裡很常見,所以這個差別會讓一堆問話得到「聽不懂」。
func MatchKeyword(keyword, input string) bool {
	if keyword == "" {
		return false
	}
	in := strings.ToLower(input)
	for i := 0; i+len(keyword) <= len(in); i++ {
		if in[i:i+len(keyword)] != keyword {
			continue
		}
		if i == 0 || in[i-1] == ' ' {
			return true
		}
	}
	return false
}

func hasByte(b []byte, v byte) bool {
	for _, c := range b {
		if c == v {
			return true
		}
	}
	return false
}

// Respond 回答玩家的一句輸入。
//
// 依記錄裡的順序找第一個命中的關鍵字(原版 `sub_1BD8C` 就是從 index 0 往上掃)。
// 找不到回傳 ok=false —— 原版那時說「I cannot help thee with that.」。
func (c *Conversation) Respond(input string) (text string, fx Effects, ok bool) {
	if strings.TrimSpace(input) == "" {
		return "", Effects{}, false
	}
	for i := range c.Entries {
		if !MatchKeyword(c.Entries[i].Keyword, input) {
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
			if c.AvatarName != "" {
				b.WriteString(c.AvatarName)
			} else {
				b.WriteString(AvatarNamePlaceholder)
			}
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
			fx.AskCode = ch
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
	// 只有在「一個字都沒有、而且什麼副作用也沒有」時才當成別名。
	//
	// ⚠ 不能只看「沒有文字」:有些回應整則只有一個提問指令(0x91–0x9F),
	// 那是「NPC 反問你」而不是「同下一則」。把它當別名的話,
	// 玩家問這個字會拿到下一個關鍵字的答案 —— 錯得很難察覺。
	if onlyOps && !fx.SameAsNext && strings.TrimSpace(b.String()) == "" &&
		!fx.AsksPlayer && !fx.JoinParty && !fx.CallGuards && !fx.EndTalk && fx.KarmaDelta == 0 {
		fx.SameAsNext = true
	}
	return strings.TrimSpace(b.String()), fx
}

// AvatarNamePlaceholder 是還不知道聖者名字時的佔位符
// (例如單獨解析對話、沒載存檔的工具用途)。
const AvatarNamePlaceholder = "汝"

// BuiltinKeywords 是引擎內建的關鍵字表(原版 `off_55E88`,34 個)。
//
// 順序不能動 —— `sub_1BE28` 是用**索引**分派的:
// 0 = NAME、1/2 = JOB/WORK、3/4 = BYE/THANK,5 之後全是髒話,
// 一律回同一句「"With language like that, how did you become an Avatar?"」。
//
// 這些字**不在**記錄的關鍵字表裡,只實作記錄那份的話,玩家問名字會得到「聽不懂」。
// 髒話那一大段是原版就有的內容,照實收錄 —— 少了它,對 NPC 罵髒話會變成沒反應,
// 與原版行為不同。
var BuiltinKeywords = [34]string{
	"name", "job", "work", "bye", "thank",
	"fuck", "shit", "damn", "dick", "prick", "pussy", "cunt", "ass", "butt",
	"booger", "piss", "jack off", "masturbate", "suck", "fart", "tits", "boob",
	"melons", "blow", "penis", "breast", "clit", "balls", "scrotum", "nuts",
	"bullshit", "cum", "crotch", "motherfucker",
}

// 內建關鍵字的索引分界(對應 sub_1BE28 的跳表)。
const (
	BuiltinName      = 0
	BuiltinJob       = 1
	BuiltinWork      = 2
	BuiltinBye       = 3
	BuiltinThank     = 4
	BuiltinProfanity = 5 // 5 之後全部同一個處理
)

// MatchBuiltin 回傳玩家輸入命中的內建關鍵字索引,-1 代表沒中。
// 掃描順序與原版一致(索引由小到大),所以 "name" 會比記錄裡的關鍵字先命中。
func MatchBuiltin(input string) int {
	for i, kw := range BuiltinKeywords {
		if MatchKeyword(kw, input) {
			return i
		}
	}
	return -1
}

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

// 提問區塊
//
// 關鍵字表(0x90 之前)後面接的是一連串**提問區塊**,結構:
//
//	0x90 <碼> <問題文字> ¶ <「否」的回答> ¶ <「是」的觸發字> ¶ <「是」的回答> ¶
//
// 碼是 0x91–0x9F 其中之一。某個關鍵字的回應裡出現同一個碼時(opcode 0x91–0x9F),
// 引擎就跳到對應的區塊發問(`sub_1C0AC` → `sub_1BCB8` 找碼相同的區塊並印出問題)。
//
// 玩家回答之後:輸入含「是」的觸發字 → 印第 4 段(`sub_1BCF4`),
// 否則印第 2 段(`sub_1BD0C`)。
//
// **終端區塊**只有問題文字、沒有分支:印的過程中碰到會讓 `sub_1C3F8` 回傳 1 的
// 指令(0x82 結束回應、0x84 入隊)就停,`sub_1C0AC` 直接返回、不向玩家要輸入。
// Gwenno 的 0x94 區塊就是這種:「Iolo and I both thank thee!」+ 入隊。
//
// 實際串起來是兩段式的:某句話拋出 0x93「Aren't ye going to ask me to join with thee?」
// → 玩家答 y → 「是」的回答是 `[94]` → 再拋出 0x94 → 終端區塊 → 道謝並入隊。
type Question struct {
	Code byte
	// Text / No / Yes 是原始位元組(尚未展開),交給 Conversation.Render 處理。
	Text    []byte
	No      []byte
	Yes     []byte
	YesWord string // 觸發「是」的字,通常是 "y"
	// Terminal 表示這個區塊印完就結束,不向玩家要輸入。
	Terminal bool
}

// stopsOutput 回報這段位元組裡有沒有「會讓輸出提前結束」的指令。
// 原版判斷終端區塊靠的就是這個(sub_1C3F8 對這些碼回傳 1)。
func stopsOutput(b []byte) bool {
	return hasByte(b, OpEndResponse) || hasByte(b, OpJoinParty)
}

// parseQuestions 解析關鍵字表之後的提問區塊。
func (c *Conversation) parseQuestions(segs [][]byte, from int) {
	for i := from; i < len(segs); i++ {
		s := segs[i]
		if len(s) < 2 || s[0] != KeywordEndMarker {
			continue
		}
		code := s[1]
		if code < OpAskFirst || code > OpAskLast {
			continue // 0x9F 是結束標記,不是提問
		}
		q := Question{Code: code, Text: s[2:]}
		if stopsOutput(q.Text) {
			q.Terminal = true
		} else {
			// 後面三段是「否的回答 / 是的觸發字 / 是的回答」,遇到下一個 0x90 就停。
			var branch [][]byte
			for j := i + 1; j < len(segs) && len(branch) < 3; j++ {
				if len(segs[j]) > 0 && segs[j][0] == KeywordEndMarker {
					break
				}
				branch = append(branch, segs[j])
			}
			if len(branch) < 3 {
				q.Terminal = true
			} else {
				q.No = branch[0]
				q.YesWord = normalizeKeyword(c.dict.Expand(branch[1]))
				q.Yes = branch[2]
			}
		}
		c.Questions = append(c.Questions, q)
	}
}

// Question 依碼取提問區塊。
func (c *Conversation) Question(code byte) (*Question, bool) {
	for i := range c.Questions {
		if c.Questions[i].Code == code {
			return &c.Questions[i], true
		}
	}
	return nil, false
}

// Render 展開一段原始位元組(問題、分支回答都走這裡)。
func (c *Conversation) Render(raw []byte) (string, Effects) {
	return c.render(raw)
}

// AnswerQuestion 依玩家的回答挑分支。
//
// 原版的判斷:輸入裡含「是」的觸發字就走 Yes,否則走 No —— 沒有第三種結果。
func (q *Question) AnswerQuestion(input string) []byte {
	if q.YesWord != "" && MatchKeyword(q.YesWord, input) {
		return q.Yes
	}
	return q.No
}
