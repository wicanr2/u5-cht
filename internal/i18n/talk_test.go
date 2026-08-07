package i18n

import "testing"

// key 的形狀:檔名#對話id#欄位。
func TestTalkKeyShape(t *testing.T) {
	if got := TalkKey("CASTLE.TLK", 8, TalkFieldGreet); got != "CASTLE.TLK#8#greet" {
		t.Errorf("key 是 %q", got)
	}
	if got := TalkEntryField(3); got != "e3" {
		t.Errorf("關鍵字回應欄位是 %q", got)
	}
	if TalkQuestionField(1) != "q1" ||
		TalkQuestionNoField(1) != "q1n" || TalkQuestionYesField(1) != "q1y" {
		t.Error("提問區塊的三個欄位名不對")
	}
}

// 查不到就回原文 —— 半套中文比整段消失好。
func TestTalkFallsBackToEnglish(t *testing.T) {
	const en = "I know nothing of that."
	if got := Talk("TOWNE.TLK", 9999, TalkFieldGreet, en, "Avatar"); got != en {
		t.Errorf("查不到應該回原文,實得 %q", got)
	}
}

// 有譯文就換掉。
func TestTalkUsesTheTranslation(t *testing.T) {
	got := Talk("CASTLE.TLK", 1, TalkFieldJob, "I try to lift people's spirits!", "")
	if got != "吾以樂音提振眾人的心緒！" {
		t.Errorf("實得 %q", got)
	}
	if !TalkTranslated("CASTLE.TLK", 1, TalkFieldJob) {
		t.Error("這一段應該算已翻")
	}
	if TalkTranslated("CASTLE.TLK", 1, "e99") {
		t.Error("不存在的欄位不該算已翻")
	}
}

// %A 代入玩家名字(對應原版 opcode 0x81)。
func TestAvatarTokenIsSubstituted(t *testing.T) {
	addTalk(map[string]string{"TOWNE.TLK#77#greet": "願" + AvatarToken + "安好。"})
	if got := Talk("TOWNE.TLK", 77, TalkFieldGreet, "", "曉風"); got != "願曉風安好。" {
		t.Errorf("實得 %q", got)
	}
	// 沒給名字就原樣留著記號 —— dump 工具會這樣用。
	if got := Talk("TOWNE.TLK", 77, TalkFieldGreet, "", ""); got != "願%A安好。" {
		t.Errorf("沒給名字時實得 %q", got)
	}
	delete(talks, "TOWNE.TLK#77#greet")
}

// ⚠ 譯文裡不該出現半形標點 —— 倚天字庫走 Big5,半形逗號句號在中文句子裡
// 會與全形混排,而且是**看得出來的醜**。這一條在守譯文的排版一致性。
func TestTranslationsUseFullWidthPunctuation(t *testing.T) {
	for k, v := range talks {
		for _, r := range []rune{',', '!', '?', ';', ':'} {
			if containsRune(v, r) {
				t.Errorf("%s 用了半形標點 %q:%s", k, string(r), v)
			}
		}
	}
}

func containsRune(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}
	return false
}
