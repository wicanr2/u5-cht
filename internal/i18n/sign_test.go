package i18n

import "testing"

// 招牌的**框**由這一層負責,不由譯者負責。
//
// 78 塊招牌實測出三種只會發生在框上的錯,每一種都改不到譯文本身:
//
//	440 列  欄寬對不上(譯者要一邊翻一邊數欄、左右補空白)
//	 14 列  譯者重打 ASCII 美術時少一根線
//	 45 塊  拆行之後忘了把底框接回去
//
// 全部都是機械規則能修的。這裡把規則釘住 —— 這幾條一旦鬆掉,
// 症狀是招牌歪掉或沒有底,而那在單元測試裡看不出來,只有肉眼看得到。

func TestSignFitReCentresContent(t *testing.T) {
	got := signFit("g 三個字 g", "g   FIVE WORD   g")
	if signCols(got) != signCols("g   FIVE WORD   g") {
		t.Errorf("置中後 %q 是 %d 欄,範本是 %d 欄",
			got, signCols(got), signCols("g   FIVE WORD   g"))
	}
	if got[0] != 'g' || got[len(got)-1] != 'g' {
		t.Errorf("兩側的框線被吃掉了:%q", got)
	}
}

func TestSignFitKeepsTheOriginalFrame(t *testing.T) {
	// 譯者少打一根線 → 一律用原文那一列。
	if got := signFit("8lllllll9", "8llllllllllllll9"); got != "8llllllllllllll9" {
		t.Errorf("框線列沒有換回原文:%q", got)
	}
}

func TestSignFitKeepsDeliberateBlankRows(t *testing.T) {
	// ⚠ 留白列也沒有中文,但它是**內容**不是框。
	// 早期把它當框換成原文,結果原文那一句英文整個活了回來(62 列)。
	got := signFit("g   g", "g  PROSECUTED g")
	if got == "g  PROSECUTED g" {
		t.Error("留白列被當成框線,英文原句被叫回來了")
	}
	if signCols(got) != signCols("g  PROSECUTED g") {
		t.Errorf("留白列 %q 是 %d 欄,範本是 %d 欄",
			got, signCols(got), signCols("g  PROSECUTED g"))
	}
}

func TestSignLinesAppendsTheMissingBottom(t *testing.T) {
	en := []string{"8llll9", "g ONE g", "jllllk"}
	addSign(map[string]string{
		"sign#99#1#1#0": "8llll9",
		"sign#99#1#1#1": "g 一 g",
	})
	got := SignLines(99, 1, 1, en)
	if len(got) != 3 || got[2] != "jllllk" {
		t.Errorf("底框沒補回來:%q", got)
	}
}

func TestSignLinesFallsBackWholesale(t *testing.T) {
	en := []string{"8llll9", "g ONE g"}
	got := SignLines(98, 7, 7, en)
	if len(got) != len(en) || got[1] != en[1] {
		t.Errorf("整塊沒譯時該原樣回去,得到 %q", got)
	}
}
