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

// TestKeywordMatchingIsFourChars:原版只比前 4 個字母,兩邊都截斷。
// 這條錯了,玩家打完整單字(telescope)會被判定成聽不懂。
func TestKeywordMatchingIsFourChars(t *testing.T) {
	d := synthDict(t)
	r := buildRecord(
		lit("A"), lit("b"), lit("c"), lit("d"), lit("e"),
		lit("star"), lit("I watch the sky."),
	)
	c := ParseConversation(r, d)
	for _, in := range []string{"star", "stars", "STARS", "Starlight"} {
		got, _, ok := c.Respond(in)
		if !ok || got != "I watch the sky." {
			t.Errorf("問 %q 得到 %q(ok=%v)", in, got, ok)
		}
	}
	if _, _, ok := c.Respond("moon"); ok {
		t.Error("不認得的關鍵字竟然有回應")
	}
	if _, _, ok := c.Respond("st"); ok {
		t.Error("只打兩個字母不該命中 star —— 比對是雙方都截到 4 個字元後相等")
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
