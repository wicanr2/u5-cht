package i18n

import (
	"strings"
	"testing"
)

// textFiles 是五個明文訊息檔與它們的筆數(實測)。
var textFiles = map[string]int{
	"STORY.DAT": 20, "QUESTION.DAT": 30, "KARMA.DAT": 6,
	"MISCMSG.DAT": 47, "ENDMSG.DAT": 11,
}

// TestEveryPlainTextRecordIsTranslated:114 筆一個都不能漏。
//
// 漏一筆的後果是**畫面上突然冒出一句英文**,而那通常出現在劇情關鍵處
// (聖壇、黑棘的審問、結局)—— 正好是最不該出戲的地方。
func TestEveryPlainTextRecordIsTranslated(t *testing.T) {
	total, missing := 0, []string{}
	for f, n := range textFiles {
		for i := 0; i < n; i++ {
			total++
			if !TextTranslated(f, i) {
				missing = append(missing, TextKey(f, i))
			}
		}
	}
	if len(missing) > 0 {
		t.Errorf("%d / %d 筆沒翻:%v", len(missing), total, missing)
	}
	// 另外兩筆是開場第 6 頁那兩句 —— 它們寫死在原版執行檔裡,不在任何 .DAT。
	const hardcoded = 2
	if TextCount() != total+hardcoded {
		t.Errorf("譯文表有 %d 筆,五個檔 %d + 寫死 %d = %d —— 有多餘或重複的 key",
			TextCount(), total, hardcoded, total+hardcoded)
	}
	for i := 0; i < hardcoded; i++ {
		if !TextTranslated("INTRO", i) {
			t.Errorf("開場寫死的第 %d 句沒翻", i)
		}
	}
}

// TestNoStrayEnglishInTranslations:譯文裡不該殘留成串的英文單字。
//
// 專有名詞(VERAMOCOR 這種要玩家打出來的咒語)例外 —— 那是刻意留的。
func TestNoStrayEnglishInTranslations(t *testing.T) {
	allowed := map[string]bool{"VERAMOCOR": true}
	for k, zh := range texts {
		word := ""
		for _, r := range zh + " " {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				word += string(r)
				continue
			}
			if len(word) >= 3 && !allowed[word] {
				t.Errorf("%s 的譯文裡有英文字 %q", k, word)
			}
			word = ""
		}
	}
}

// TestSentenceFragmentsStayOpen:句子的前半譯完之後也要留成可接的形狀。
//
// `MISCMSG#0` 原文結尾是「…the Mystic Shrine of 」,後面由程式接美德名。
// 譯文若自己補上句號,接出來就會變成「…真言是什麼。誠實」。
func TestSentenceFragmentsStayOpen(t *testing.T) {
	// 這幾筆原文結尾沒有標點,是等著被接的。
	for _, k := range []string{
		"MISCMSG.DAT#0", "MISCMSG.DAT#1", "MISCMSG.DAT#2",
		"MISCMSG.DAT#8", "MISCMSG.DAT#32",
		"ENDMSG.DAT#1", "ENDMSG.DAT#2",
	} {
		zh := texts[k]
		if zh == "" {
			t.Errorf("%s 沒有譯文", k)
			continue
		}
		trimmed := strings.TrimRight(zh, " \n")
		last := []rune(trimmed)
		if len(last) == 0 {
			continue
		}
		switch last[len(last)-1] {
		case '。', '!', '」', '?':
			t.Errorf("%s 的譯文以 %q 收尾 —— 它是句子的前半,後面還要接東西",
				k, string(last[len(last)-1]))
		}
	}
}

// TestKarmaMessagesAreOrdered:六段業報訊息的語氣要從責備走到讚許。
//
// 這條擋的是「複製貼上時把順序弄反」—— 那會讓高業報的玩家被罵、
// 低業報的被誇,而測試若只檢查「有沒有翻」是看不出來的。
// 用最粗的訊號:前兩段該出現負面詞,後兩段該出現正面詞。
func TestKarmaMessagesAreOrdered(t *testing.T) {
	negative := []string{"遠離", "迷途"}
	positive := []string{"啟蒙", "啟蒙"}
	for i, want := range negative {
		if !strings.Contains(texts[TextKey("KARMA.DAT", i)], want) {
			t.Errorf("業報第 %d 段應該是責備的語氣(找不到 %q)", i, want)
		}
	}
	for i, want := range positive {
		k := TextKey("KARMA.DAT", 4+i)
		if !strings.Contains(texts[k], want) {
			t.Errorf("業報第 %d 段應該是讚許的語氣(找不到 %q)", 4+i, want)
		}
	}
}
