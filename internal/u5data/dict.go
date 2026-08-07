package u5data

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// 文字壓縮:詞典與展開規則
//
// 遊戲裡幾乎所有英文文字(`.TLK` 對話、`SHOPPE.DAT`、各種 `.DAT`)都用同一份
// 118 字的常用詞詞典壓縮 —— 但**兩類檔案的極性相反**,踩錯會解出完全不同的東西:
//
//	.TLK :  bit7 有設 = 字面文字   bit7 沒設 = token(slot = b)
//	.DAT :  bit7 沒設 = 字面文字   bit7 有設 = token(slot = b - 0x7F)
//
// 這一條特別容易誤判,因為錯的解法「看起來只差一點」:把 `.TLK` 整份清掉 bit7,
// 結果是 `I study\x01stars.` —— 讀起來像混進一個雜訊字元,實際上 0x01 是 "the",
// 正確答案是 `I study the stars.`。
//
// `.TLK` 的展開規則出自 FM Towns 版的 `sub_1C3F8`(docs/re/05);
// `.DAT` 的極性與偏移是實測出來的(見 ExpandDAT)。
//
//	b < 0x81:  輸出一個空格,再輸出 dict[b];並設下 pendingSpace
//	b >= 0x81: 字面字元 (b & 0x7F);0x8D(CR)改成 0x8A(LF);
//	           若 pendingSpace 則先補一個空格,然後清掉它
//
// 注意空格只由 token 自己前置、由下一個**字面**字元補後置 ——
// 所以連續的 token 之間不會出現雙空格。
const (
	// DictTokenMax 是最大的詞典索引(1..128)。
	DictTokenMax = 128
	// DictWordCount 是實際有字的槽數(128 減掉 10 個空槽)。
	DictWordCount = 118
	// DictLiteralMin 是「這個位元組是字面文字」的下限。
	DictLiteralMin = 0x81
)

// dictHoles 是詞典指標表裡的空槽。
//
// 原版 FM Towns 執行檔的 `dword_41990` 是一張**用 token 值直接索引**的指標表,
// 中間有 10 個 NULL。DOS 版的 `DATA.OVL` 則存成緊密的 118 字清單,
// 所以要把清單塞回槽位時得跳過這些洞。
//
// 這 10 個洞可以獨立驗證:統計四個 `.TLK` 的 token 值,
// 8 / 28 / 50 / 65 / 67 / 70 / 72–75 **一次都沒出現過**。
// (先前分析 SHOPPE.DAT 時觀察到的「token 與清單索引固定差 10」,
// 差的就是這 10 個洞 —— 最後一個洞在 75,之後差值就固定了。)
var dictHoles = [...]int{8, 28, 50, 65, 67, 70, 72, 73, 74, 75}

// DictOffset 是 DOS 版 `DATA.OVL` 裡詞典的位移。
const DictOffset = 0x104C

// Dictionary 是展開壓縮文字用的詞典,依 token 值索引。
type Dictionary struct {
	words [DictTokenMax + 1]string
}

// LoadDictionary 從 DATA.OVL 讀出詞典。
//
// 只有「哪些槽是空的」來自反組譯,單字本身一律讀玩家自己的原版檔 ——
// 這樣 repo 裡不需要放任何遊戲文字。
func LoadDictionary(gameDataDir string) (*Dictionary, error) {
	raw, err := os.ReadFile(filepath.Join(gameDataDir, "DATA.OVL"))
	if err != nil {
		return nil, err
	}
	return ParseDictionary(raw)
}

// ParseDictionary 從 DATA.OVL 的內容取出詞典。
func ParseDictionary(dataOVL []byte) (*Dictionary, error) {
	if len(dataOVL) <= DictOffset {
		return nil, fmt.Errorf("DATA.OVL 只有 %d B,取不到 0x%X 的詞典", len(dataOVL), DictOffset)
	}
	hole := map[int]bool{}
	for _, h := range dictHoles {
		hole[h] = true
	}

	d := &Dictionary{}
	p := DictOffset
	n := 0
	for tok := 1; tok <= DictTokenMax; tok++ {
		if hole[tok] {
			continue
		}
		e := indexByte(dataOVL, p)
		if e < 0 {
			return nil, fmt.Errorf("詞典在第 %d 個字之後沒有結尾", n)
		}
		w := string(dataOVL[p:e])
		if w == "" || !printableASCII(w) {
			return nil, fmt.Errorf("詞典第 %d 個字(token %d)不像單字:%q", n, tok, w)
		}
		d.words[tok] = w
		p = e + 1
		n++
	}
	// 自我一致性:第一個字必須是 the、最後一個必須是 work、總數 118。
	// 對不上通常代表拿到別版的 DATA.OVL,而不是程式錯 —— 早點失敗比默默解錯好。
	if n != DictWordCount {
		return nil, fmt.Errorf("詞典讀到 %d 個字,預期 %d", n, DictWordCount)
	}
	if d.words[1] != "the" || d.words[DictTokenMax] != "work" {
		return nil, fmt.Errorf("詞典頭尾是 %q / %q,預期 \"the\" / \"work\"",
			d.words[1], d.words[DictTokenMax])
	}
	return d, nil
}

// Word 回傳 token 對應的字;空槽或超出範圍回傳空字串。
func (d *Dictionary) Word(tok byte) string {
	if int(tok) > DictTokenMax {
		return ""
	}
	return d.words[tok]
}

// Expand 把一段壓縮文字展開成純文字。
//
// d 可為 nil —— 那時 token 會保留成 `<XX>`,方便在還沒有 DATA.OVL 時看出
// 「這裡有一個沒展開的字」,而不是靜默生出破碎的句子。
func (d *Dictionary) Expand(raw []byte) string {
	var b strings.Builder
	pendingSpace := false
	for _, c := range raw {
		if c >= DictLiteralMin {
			ch := c & 0x7F
			if ch == '\r' {
				ch = '\n' // 原版 sub_1C3F8:0x8D → 0x8A
			}
			if pendingSpace {
				b.WriteByte(' ')
				pendingSpace = false
			}
			b.WriteByte(ch)
			continue
		}
		// token
		var w string
		if d != nil {
			w = d.Word(c)
		}
		if w == "" {
			// 空槽在原版是「把原位元組照樣輸出」。沒有詞典時則標記出來。
			if d != nil {
				b.WriteByte(c)
			} else {
				fmt.Fprintf(&b, "<%02X>", c)
			}
			continue
		}
		b.WriteByte(' ')
		b.WriteString(w)
		pendingSpace = true
	}
	return b.String()
}

// ExpandDAT 展開 `.DAT` 系列(SHOPPE.DAT 等)的壓縮文字。
//
// ⚠ 極性與 `.TLK` **相反**:`.DAT` 的本文是普通 ASCII(bit7 清除),
// bit7 有設的才是 token;而且 token 的槽位差 0x7F ——
//
//	.TLK :  slot = b        (b <= 0x80 是 token)
//	.DAT :  slot = b - 0x7F (b >= 0x80 是 token)
//
// 兩邊都對到同一張 1..128 的槽表。這個差別是實測出來的(還沒在反編譯裡找到
// `.DAT` 的輸出常式),證據是同一批句子在 +0x7F 之下才讀得通:
//
//	"Thanks {86} nothing!"                   → for      (86-7F = 7)
//	"Come back {94} you're ready {83} buy"   → when / to
//	"Be off {8B} ye, then..."                → with
//	"Our Iron Helms {9C} padded"             → are      (槽 29;29-1=28 正好是一個空槽)
//
// 最後一例特別有力:0x9C 若照 .TLK 的算法會落在空槽,照 .DAT 的算法才有字。
func (d *Dictionary) ExpandDAT(raw []byte) string {
	var b strings.Builder
	pendingSpace := false
	for _, c := range raw {
		if c < 0x80 {
			ch := c
			if ch == '\r' {
				ch = '\n'
			}
			if pendingSpace {
				b.WriteByte(' ')
				pendingSpace = false
			}
			b.WriteByte(ch)
			continue
		}
		var w string
		if d != nil {
			w = d.Word(c - 0x7F)
		}
		if w == "" {
			fmt.Fprintf(&b, "<%02X>", c)
			continue
		}
		b.WriteByte(' ')
		b.WriteString(w)
		pendingSpace = true
	}
	return b.String()
}

func indexByte(b []byte, from int) int {
	for i := from; i < len(b); i++ {
		if b[i] == 0 {
			return i
		}
	}
	return -1
}

func printableASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < 0x20 || s[i] >= 0x7F {
			return false
		}
	}
	return true
}
