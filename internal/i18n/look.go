package i18n

import (
	"fmt"
	"strings"
)

// `LOOK2.DAT` 與 `SIGNS.DAT` 的譯文覆蓋層
//
// # 為什麼 key 是索引而不是原文
//
// `LOOK2.DAT` 有 512 格但只有 216 段不重複的敘述 —— 同一句話掛在好幾格上
// (草地就佔了十幾格)。拿原文當 key 會**把它們綁死**:之後想讓「城鎮的草地」
// 與「野外的草地」用不同說法就辦不到。索引當 key 保留了這個自由度,
// 代價只是譯文表大一點。
//
// key 的形狀:
//
//	look#<索引>      索引即 LookTable 的格號:地形 0..255,物件 256..511
//	looksuf#<英文>   接在敘述後面的專名(三聖火、八座地牢)
//	sign#<地點>#<x>#<y>#<第幾列>
//
// # 招牌為什麼連座標一起當 key
//
// 招牌是**畫出來的框 + 文字**,一列 16 欄。中文一個字佔兩欄,所以同一塊招牌
// 譯成中文之後列數與內容都會變 —— 沒辦法逐列對譯。因此譯文層直接接管
// **整塊招牌**(用 `SignLines` 一次換掉所有列),key 用「哪一塊」而不是「哪一句」。
//
// ⚠ 沒有譯文時回原文的那幾列 —— 框線是 `abbc` / `8lll9` 這種字母美術,
// 得靠招牌字型才看得懂。這是已知的降級,寫在 `docs/localization-notes.md`。

// LookKey 組出敘述表的 key。
func LookKey(index int) string { return fmt.Sprintf("look#%d", index) }

// Look 查一格的敘述譯文。查不到就回原文。
func Look(index int, en string) string {
	if zh, ok := looks[LookKey(index)]; ok {
		return zh
	}
	return en
}

// LookSuffixKey 組出後綴專名的 key。
func LookSuffixKey(en string) string { return "looksuf#" + en }

// LookSuffix 查接在敘述尾巴的專名(`the Flame of ` 後面的火名、
// `the collapsed entrance to the dungeon ` 後面的地牢名)。
func LookSuffix(en string) string {
	if zh, ok := looks[LookSuffixKey(en)]; ok {
		return zh
	}
	return en
}

// LookTranslated 回報這一格翻過了沒。統計覆蓋率用。
func LookTranslated(index int) bool {
	_, ok := looks[LookKey(index)]
	return ok
}

// LookCount 是已翻的筆數(含後綴專名)。
func LookCount() int { return len(looks) }

// SignKey 組出招牌某一列的 key。
func SignKey(location, x, y, line int) string {
	return fmt.Sprintf("sign#%d#%d#%d#%d", location, x, y, line)
}

// SignLines 用譯文換掉整塊招牌。
//
// 譯文的列數**可以與原文不同** —— 中文一個字佔兩欄,一列 14 欄只放得下
// 七個字,排不下就得多分一列。所以這裡從第 0 列連號往下取,取到缺號為止,
// 而不是照原文的列數配對。一列都沒譯就整塊原樣回去。
//
// # 欄寬由這裡負責,不是譯者
//
// 招牌是畫出來的框。譯者要一邊翻一邊數欄、還要左右補空白補到兩側框線對齊,
// 錯一格就歪 —— 而歪掉的框在任何機械檢查裡都看得出來,卻要人工一列一列改。
// 那是**該由程式做的事**:`signFit` 拿原文那一列的框當範本,把譯文的內容
// 重新置中。譯者只要把字寫對,對齊交給引擎。
func SignLines(location, x, y int, en []string) []string {
	var out []string
	for i := 0; ; i++ {
		zh, ok := signs[SignKey(location, x, y, i)]
		if !ok {
			break
		}
		out = append(out, signFit(zh, signTemplate(en, i)))
	}
	if len(out) == 0 {
		return en
	}
	return out
}

// signTemplate 挑原文的哪一列當對齊範本。
//
// 譯文多出來的列(中文拆行)沒有對應的原文,拿**最後一列有文字的**當範本 ——
// 那一定是「框 + 內容」的形狀,而不是純框線。
func signTemplate(en []string, i int) string {
	if i < len(en) {
		return en[i]
	}
	for j := len(en) - 1; j >= 0; j-- {
		if signHasContent(en[j]) {
			return en[j]
		}
	}
	if len(en) > 0 {
		return en[len(en)-1]
	}
	return ""
}

// signHasContent 回報這一列是不是「框 + 內容」而不是純框線。
//
// 判斷依據是**有沒有空白**:框線列是一整排連續的線條字模,不會有空格;
// 有內容的列一定有(內容兩側的留白,或字與字之間)。
func signHasContent(line string) bool {
	for _, r := range line {
		if r == ' ' {
			return true
		}
	}
	return false
}

// signFit 把一列譯文對齊到範本的欄寬。
//
// 只在**兩端的字元與範本一致**時才動手 —— 那代表譯者留著同樣的框,
// 中間才是內容。兩端不一致就原樣放行:可能是譯者刻意改了框,
// 也可能是純框線列,無論哪種都不該被程式亂改。
func signFit(zh, template string) string {
	w := signCols(template)
	if w == 0 {
		return zh
	}
	// 純框線列一律用原文。譯者重打一次 ASCII 美術只會少一根線 ——
	// 實測 78 塊裡有 14 列就是這樣歪掉的。框不是譯文,不該經過譯者的手。
	if !signHasCJK(zh) {
		return template
	}
	if signCols(zh) == w {
		return zh
	}
	tr, zr := []rune(template), []rune(zh)
	if len(tr) < 3 || len(zr) < 3 || tr[0] != zr[0] || tr[len(tr)-1] != zr[len(zr)-1] {
		return zh
	}
	inner := strings.TrimSpace(string(zr[1 : len(zr)-1]))
	room := w - signCols(string(tr[0])) - signCols(string(tr[len(tr)-1]))
	pad := room - signCols(inner)
	if pad < 0 {
		return zh // 塞不下就別動,讓它明顯地凸出來
	}
	left := pad / 2
	return string(tr[0]) + strings.Repeat(" ", left) +
		inner + strings.Repeat(" ", pad-left) + string(tr[len(tr)-1])
}

// signHasCJK 回報這一列有沒有中文。沒有就代表它是框線,不是譯文。
func signHasCJK(s string) bool {
	for _, r := range s {
		if r > 0x7F {
			return true
		}
	}
	return false
}

// signCols 是這一串字佔幾欄。中文全形算兩欄。
func signCols(s string) int {
	n := 0
	for _, r := range s {
		if r > 0x7F {
			n += 2
			continue
		}
		n++
	}
	return n
}

// SignTranslated 回報這一塊招牌翻過了沒。
func SignTranslated(location, x, y int) bool {
	_, ok := signs[SignKey(location, x, y, 0)]
	return ok
}

// SignCount 是已翻的招牌列數。
func SignCount() int { return len(signs) }

var (
	looks = map[string]string{}
	signs = map[string]string{}
)

func addLook(m map[string]string) {
	for k, v := range m {
		looks[k] = v
	}
}

func addSign(m map[string]string) {
	for k, v := range m {
		signs[k] = v
	}
}
