package u5data

import (
	"os"
	"strings"
	"testing"
)

// synthDict 造一份最小詞典,避免這些測試依賴原版資料。
func synthDict(t *testing.T) *Dictionary {
	t.Helper()
	d := &Dictionary{}
	d.words[1] = "the"
	d.words[3] = "of"
	return d
}

// lit 把字面文字編成 .TLK 的 high-bit 形式。
func lit(s string) []byte {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		b[i] = s[i] | 0x80
	}
	return b
}

func join(parts ...[]byte) []byte {
	var out []byte
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

// buildRecord 依 .TLK 的段落結構組一筆記錄:5 個固定欄位 + 成對的關鍵字/回應。
func buildRecord(segs ...[]byte) *TalkRecord {
	var data []byte
	for i, s := range segs {
		if i > 0 {
			data = append(data, 0)
		}
		data = append(data, s...)
	}
	return &TalkRecord{NPCIndex: 1, Data: data}
}

func TestParseConversationFields(t *testing.T) {
	d := synthDict(t)
	r := buildRecord(
		lit("Zachariah"), lit("a stately man."), lit("Welcome!"),
		join(lit("I study"), []byte{1}, lit("stars.")), lit("Good journeys."),
	)
	c := ParseConversation(r, d)
	if c.Name != "Zachariah" {
		t.Errorf("名字 %q", c.Name)
	}
	if c.Description != "a stately man." {
		t.Errorf("外貌 %q", c.Description)
	}
	if c.Job != "I study the stars." {
		t.Errorf("職業 %q —— 詞典 token 沒展開?", c.Job)
	}
	if c.Bye != "Good journeys." {
		t.Errorf("道別 %q", c.Bye)
	}
}

// TestKeywordMatchingIsSubstringAtWordStart:原版不是「比前 4 個字母」,
// 而是把記錄裡的關鍵字當**子字串**去玩家輸入裡找,且必須落在詞首
// (`sub_1BD8C` → `sub_27C98`,再檢查前一個字元是不是空白)。
//
// 這個差別對短關鍵字很致命:記錄裡的 `bow` 要能被輸入 `bows` 命中,
// 但截斷比對會把 `bows` 截成 `bows` 而配不上。
func TestKeywordMatchingIsSubstringAtWordStart(t *testing.T) {
	d := synthDict(t)
	r := buildRecord(
		lit("A"), lit("b"), lit("c"), lit("d"), lit("e"),
		lit("star"), lit("I watch the sky."),
	)
	c := ParseConversation(r, d)
	for _, in := range []string{"star", "stars", "STARS", "Starlight", "the stars"} {
		got, _, ok := c.Respond(in)
		if !ok || got != "I watch the sky." {
			t.Errorf("問 %q 得到 %q(ok=%v)", in, got, ok)
		}
	}
	if _, _, ok := c.Respond("moon"); ok {
		t.Error("不認得的關鍵字竟然有回應")
	}
	if _, _, ok := c.Respond("st"); ok {
		t.Error("只打兩個字母不該命中 star —— 關鍵字要整個出現在輸入裡")
	}
	// 必須落在詞首:moonstar 裡的 star 前面不是空白,不算。
	if _, _, ok := c.Respond("moonstar"); ok {
		t.Error("moonstar 不該命中 star —— 命中位置必須是 0 或前一字元為空白")
	}
}

// TestSameAsNextChain:回應只有 0x87 代表「同下一則」,而且可能連續好幾層。
// 占星師的 tele/star 就是這樣共用一段回答。
func TestSameAsNextChain(t *testing.T) {
	d := synthDict(t)
	r := buildRecord(
		lit("A"), lit("b"), lit("c"), lit("d"), lit("e"),
		lit("tele"), []byte{OpSameAsNext},
		lit("scop"), []byte{OpSameAsNext},
		lit("star"), lit("I watch the sky."),
	)
	c := ParseConversation(r, d)
	for _, in := range []string{"tele", "scop", "star"} {
		got, fx, ok := c.Respond(in)
		if !ok || got != "I watch the sky." {
			t.Errorf("問 %q 得到 %q(ok=%v)", in, got, ok)
		}
		if fx.SameAsNext {
			t.Errorf("問 %q 的結果仍標著「同下一則」—— 別名鏈沒走完", in)
		}
	}
}

// TestEffects:opcode 會改變遊戲狀態,不能只當成要濾掉的雜訊。
func TestEffects(t *testing.T) {
	d := synthDict(t)
	r := buildRecord(
		lit("A"), lit("b"), lit("c"), lit("d"), lit("e"),
		lit("join"), join(lit("I shall come!"), []byte{OpJoinParty}),
		lit("stea"), join(lit("Thief!"), []byte{OpCallGuards, OpKarmaDown}),
		lit("alms"), join(lit("Bless thee."), []byte{OpKarmaUp}),
		lit("who"), join(lit("I am"), []byte{OpAvatarName}),
	)
	c := ParseConversation(r, d)

	if _, fx, _ := c.Respond("join"); !fx.JoinParty {
		t.Error("0x84 沒有被解讀成「加入隊伍」")
	}
	_, fx, _ := c.Respond("stea")
	if !fx.CallGuards || fx.KarmaDelta != -1 {
		t.Errorf("叫衛兵 %v、業報 %+d,預期 true / -1", fx.CallGuards, fx.KarmaDelta)
	}
	if _, fx, _ := c.Respond("alms"); fx.KarmaDelta != +1 {
		t.Errorf("業報 %+d,預期 +1", fx.KarmaDelta)
	}
	// 0x81 是「插入聖者的名字」,不是要丟掉的控制碼。
	if got, _, _ := c.Respond("who"); !strings.Contains(got, AvatarNamePlaceholder) {
		t.Errorf("回應 %q 裡沒有聖者名字的佔位符", got)
	}
}

// TestAvatarNameSurvivesTheFixedFields:招呼語裡的玩家名字不准被吃掉。
//
// 四個固定欄位走的是 Expand + cleanText,不是 render 的 opcode 路徑。
// 0x81 在 Expand 眼中是字面文字,清掉 bit7 之後變成 0x01,再被 cleanText
// 當控制碼丟掉 —— 招呼語就印成 `G'day, !`(CASTLE.TLK#20 的真實症狀)。
//
// ⚠ 這條壞在**英文原文那一支**,所以「譯文漏字」的直覺是錯的:
// 工作單上根本看不到那個名字,譯者無從得知該補。
func TestAvatarNameSurvivesTheFixedFields(t *testing.T) {
	d := synthDict(t)
	r := buildRecord(
		lit("A"),
		lit("a farmer."),
		join(lit("G'day, "), []byte{OpAvatarName}, lit("!")),
		lit("I farm."),
		lit("Bye."),
	)
	c := ParseConversation(r, d)
	if !strings.Contains(c.Greeting, AvatarToken) {
		t.Errorf("招呼語 %q 裡沒有 %s —— 玩家的名字被 cleanText 吃掉了",
			c.Greeting, AvatarToken)
	}
}

// TestAllConversationsParse 用原版資料把四個 .TLK 全部走一遍。
//
// 檢查的是「不會爆、不會無窮迴圈、每個關鍵字都給得出答案」——
// 別名鏈若成環,Respond 會卡住,這個測試會直接吊死而不是靜默通過。
func TestAllConversationsParse(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	set, err := LoadTalkSet(dir)
	if err != nil {
		t.Fatal(err)
	}
	convs, entries, noName := 0, 0, 0
	for fi := range TalkFiles {
		tf := set.Files[fi]
		for i := range tf.Records {
			c := ParseConversation(&tf.Records[i], set.Dict)
			convs++
			if c.Name == "" {
				noName++
			}
			for _, kw := range c.Keywords() {
				text, _, ok := c.Respond(kw)
				if !ok {
					t.Errorf("%s id %d:自己列出的關鍵字 %q 卻答不出來",
						TalkFiles[fi], c.ID, kw)
				}
				entries++
				_ = text
			}
		}
	}
	if convs < 100 {
		t.Errorf("只解出 %d 段對話,太少", convs)
	}
	// 名字是最穩定的欄位;大量缺名代表段落切法錯了。
	if noName*4 > convs {
		t.Errorf("%d/%d 段對話沒有名字,超過四分之一 —— 段落切法有問題", noName, convs)
	}
	t.Logf("四個 .TLK 共 %d 段對話、%d 個關鍵字,%d 段無名", convs, entries, noName)
}

// TestKeywordListEndsAt0x90:關鍵字表在遇到 0x90 時結束。
//
// 少了這個終止條件,0x90 之後的提問區塊會被當成關鍵字對,
// 把回應接到錯的字上 —— 而且錯得很難察覺(每個字都「有」答案,只是答錯)。
// 實測影響很大:四個 .TLK 的關鍵字數從 1767 掉到 1307,26% 是假的。
func TestKeywordListEndsAt0x90(t *testing.T) {
	d := synthDict(t)
	r := buildRecord(
		lit("A"), lit("b"), lit("c"), lit("d"), lit("e"),
		lit("star"), lit("I watch the sky."),
		append([]byte{KeywordEndMarker, 0x91}, lit("Art thou from here?")...),
		lit("I thought not."),
		lit("y"),
		lit("Then perhaps not too smart!"),
	)
	c := ParseConversation(r, d)
	if kws := c.Keywords(); len(kws) != 1 || kws[0] != "star" {
		t.Errorf("關鍵字表是 %v,預期只有 [star] —— 0x90 之後不該被當成關鍵字", kws)
	}
	q, ok := c.Question(0x91)
	if !ok {
		t.Fatal("沒有解出 0x91 的提問區塊")
	}
	if got, _ := c.Render(q.Text); got != "Art thou from here?" {
		t.Errorf("問題文字是 %q", got)
	}
	if q.Terminal {
		t.Error("這個區塊有三段分支,不該是終端")
	}
	if q.YesWord != "y" {
		t.Errorf("「是」的觸發字是 %q,預期 \"y\"", q.YesWord)
	}
	if got, _ := c.Render(q.AnswerQuestion("y")); got != "Then perhaps not too smart!" {
		t.Errorf("答 y 得到 %q", got)
	}
	if got, _ := c.Render(q.AnswerQuestion("n")); got != "I thought not." {
		t.Errorf("答 n 得到 %q", got)
	}
}

// TestTerminalQuestionBlock:只有問題文字、而且文字裡有「會讓輸出停下」的指令
// (0x82 結束回應 / 0x84 入隊)的區塊,印完就結束,不向玩家要輸入。
func TestTerminalQuestionBlock(t *testing.T) {
	d := synthDict(t)
	r := buildRecord(
		lit("A"), lit("b"), lit("c"), lit("d"), lit("e"),
		lit("join"), []byte{0x94},
		append(append([]byte{KeywordEndMarker, 0x94}, lit("We both thank thee!")...), OpJoinParty),
	)
	c := ParseConversation(r, d)
	// 問 join → 回應是純提問碼,不該被誤判成「同下一則」
	_, fx, ok := c.Respond("join")
	if !ok || !fx.AsksPlayer || fx.AskCode != 0x94 {
		t.Fatalf("問 join 得到 ok=%v asks=%v code=0x%02X", ok, fx.AsksPlayer, fx.AskCode)
	}
	q, found := c.Question(0x94)
	if !found {
		t.Fatal("沒有解出 0x94 的區塊")
	}
	if !q.Terminal {
		t.Error("含入隊指令的單段區塊應該是終端")
	}
	text, tfx := c.Render(q.Text)
	if text != "We both thank thee!" || !tfx.JoinParty {
		t.Errorf("區塊文字 %q,入隊 %v", text, tfx.JoinParty)
	}
}

// TestBuiltinKeywords:內建表的順序決定分派,不能動。
func TestBuiltinKeywords(t *testing.T) {
	for kw, want := range map[string]int{
		"name": BuiltinName, "job": BuiltinJob, "work": BuiltinWork,
		"bye": BuiltinBye, "thank": BuiltinThank,
	} {
		if got := MatchBuiltin(kw); got != want {
			t.Errorf("%q 命中索引 %d,預期 %d", kw, got, want)
		}
	}
	// 髒話一律落在 BuiltinProfanity 之後(原版對它們回同一句)。
	if got := MatchBuiltin("damn"); got < BuiltinProfanity {
		t.Errorf("damn 命中索引 %d,應該落在髒話區(>= %d)", got, BuiltinProfanity)
	}
	if got := MatchBuiltin("telescope"); got != -1 {
		t.Errorf("telescope 不該命中任何內建關鍵字,實得 %d", got)
	}
	// 詞首規則同樣適用:surname 裡的 name 前面不是空白。
	if got := MatchBuiltin("surname"); got != -1 {
		t.Errorf("surname 不該命中 name(命中位置要在詞首),實得 %d", got)
	}
}
