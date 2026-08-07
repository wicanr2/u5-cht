package u5data

import (
	"fmt"
	"os"
	"path/filepath"
)

// 施法與調藥的輸入法:打符文首字母,不是打咒語全名(原版 `sub_1CA0C`)
//
// ★ 這一條把 `WORKLIST` 上「法術代碼表 98 項縮寫,對應關係待解」結掉了,
// 而且結論改了引擎的輸入模型 —— 原本 `CastByName` 讓玩家把 `In Lor` 整串打進去,
// **那不是原版的玩法**。
//
// # 兩張相鄰的表
//
// `DATA.OVL` 裡這兩張表是連著的,FM Towns 的指標表也連著
//(`0x40D54` 起 26 個 → `0x40DBC` 起 48 個),這個相鄰本身就是結構佐證:
//
//	DOS 0x0941  24 個符文詞  `AN BET CORP DES EX FLAM GRAV HUR IN KAL LOR MANI
//	                          NOX POR QUAS REL SANCT TYM UUS VAS WIS XEN YLEM ZU`
//	DOS 0x09A5  48 個代碼    `IL GP AZ AN M AY AS ACX HR IW KX IMX LV FV …`
//
// 24 個詞的首字母正好是 **A–Z 去掉 J 與 O** —— 所以原版讀到 `J` 或 `O`
// 是**默默丟掉重讀**(`cmp al,'J'; jz 重讀` / `cmp al,'O'; jz 重讀`),
// 不是印錯誤訊息。少了這一條,按到 J 會查到表外。
//
// 而 48 個代碼是**咒語名的首字母排序後**的樣子:
//
//	In Lor        → I,L       → IL
//	Rel Hur       → R,H  排序 → HR      ← 順序反了才看得出是排序過的
//	Vas Lor       → V,L  排序 → LV
//	An Xen Corp   → A,X,C 排序 → ACX
//	In Vas Por Ylem → I,V,P,Y 排序 → IPVY
//
// 48 筆逐一驗過(`TestSpellCodesAreSortedInitials`),0 筆例外。
//
// # 流程(照抄 `sub_1CA0C`)
//
//	1. 收字母,**最多 4 個**(`cmp ebx,4; jge 丟掉`);每收一個就把整個符文詞
//	   回顯在畫面上,詞間一個空格,累計超過 13 欄先換行
//	2. Backspace 退一個(把剛才那個詞從畫面上擦掉)
//	3. **Enter 或空白鍵**都是送出;ESC 是取消(等同一個字母都沒打)
//	4. 送出時把字母**氣泡排序**成遞增
//	5. 逐一比對 48 個代碼;比對時**連續重複的輸入字母只算一個**
//	   (所以打 `AA` 等於打 `A`)
//	6. 一個字母都沒打 → −1;打了但查不到 → −2;查到 → 咒語索引
//
// 呼叫端有兩個:`sub_1994C`(施法)與 `sub_18704`(調藥)——
// **同一個輸入法**,所以調藥也是打符文首字母。
const (
	// RuneInputMax 是最多能打幾個字母。
	RuneInputMax = 4
	// RuneInputWrapAt 是回顯累計到幾欄要換行(原版 `cmp eax, 0Dh; jle`)。
	RuneInputWrapAt = 13
)

// 輸入的兩種非咒語結果(原版的回傳值)。
const (
	// SpellInputCancelled 是一個字母都沒打就送出(含 ESC)—— 原版印 "None!"。
	SpellInputCancelled = -1
	// SpellInputNoSpell 是打了字母但湊不出咒語 —— 原版印 "No effect!"。
	SpellInputNoSpell = -2
)

// runeWordsOffset / spellCodesOffset 是 `DATA.OVL` 裡兩張表的位移。
const (
	runeWordsOffset  = 0x0941
	spellCodesOffset = 0x09A5
)

// RuneWordCount 是符文詞的數量(A–Z 去掉 J 與 O)。
const RuneWordCount = 24

// RuneTable 是符文詞與咒語代碼。
type RuneTable struct {
	// Words[i] 是首字母 'A'+i 的符文詞;J 與 O 是空字串。
	Words [26]string
	// Codes[i] 是第 i 個咒語的代碼(首字母排序後)。
	Codes [SpellCount]string
}

// ParseRuneTable 從 `DATA.OVL` 的內容取出兩張表。
func ParseRuneTable(ovl []byte) (*RuneTable, error) {
	t := &RuneTable{}
	read := func(off, n int) ([]string, int, error) {
		out := make([]string, 0, n)
		for len(out) < n {
			end := off
			for end < len(ovl) && ovl[end] != 0 {
				end++
			}
			if end >= len(ovl) {
				return nil, 0, fmt.Errorf("第 %d 筆就跑出檔尾了", len(out))
			}
			out = append(out, string(ovl[off:end]))
			off = end + 1
		}
		return out, off, nil
	}
	words, _, err := read(runeWordsOffset, RuneWordCount)
	if err != nil {
		return nil, fmt.Errorf("符文詞表:%w", err)
	}
	for _, w := range words {
		if w == "" {
			return nil, fmt.Errorf("符文詞表裡有空字串,位移 0x%04X 對不上", runeWordsOffset)
		}
		i := int(w[0] - 'A')
		if i < 0 || i >= len(t.Words) {
			return nil, fmt.Errorf("符文詞 %q 的首字母不是大寫字母", w)
		}
		t.Words[i] = w
	}
	codes, _, err := read(spellCodesOffset, SpellCount)
	if err != nil {
		return nil, fmt.Errorf("咒語代碼表:%w", err)
	}
	copy(t.Codes[:], codes)
	if err := t.validate(); err != nil {
		return nil, err
	}
	return t, nil
}

// LoadRuneTable 從 `DATA.OVL` 讀出兩張表。
func LoadRuneTable(dir string) (*RuneTable, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "DATA.OVL"))
	if err != nil {
		return nil, err
	}
	return ParseRuneTable(raw)
}

// validate 用「位移偏掉就一定違反」的性質把兩張表釘住。
func (t *RuneTable) validate() error {
	// 符文詞:兩頭各一個,加上 J / O 必須是空的 —— 後者同時證明「24 而不是 26」。
	if t.Words['A'-'A'] != "AN" {
		return fmt.Errorf("A 的符文詞是 %q,預期 AN", t.Words[0])
	}
	if t.Words['Z'-'A'] != "ZU" {
		return fmt.Errorf("Z 的符文詞是 %q,預期 ZU", t.Words['Z'-'A'])
	}
	for _, c := range []byte{'J', 'O'} {
		if w := t.Words[c-'A']; w != "" {
			return fmt.Errorf("%c 竟然有符文詞 %q —— 表偏了", c, w)
		}
	}
	// 代碼:兩頭各一個,中間挑一個**首字母順序被排序打亂**的
	//(`Rel Hur` → HR)。三個一起中,表就沒有滑動的餘地。
	for _, c := range []struct {
		i    int
		want string
	}{{0, "IL"}, {8, "HR"}, {SpellCount - 1, "AT"}} {
		if t.Codes[c.i] != c.want {
			return fmt.Errorf("第 %d 個咒語代碼是 %q,預期 %q", c.i, t.Codes[c.i], c.want)
		}
	}
	// 代碼裡不能有重複字母 —— 有的話那一筆永遠比不中(見 Match 的說明)。
	for i, code := range t.Codes {
		for j := 1; j < len(code); j++ {
			if code[j] == code[j-1] {
				return fmt.Errorf("第 %d 個咒語代碼 %q 有連續重複字母", i, code)
			}
		}
	}
	return nil
}

// RuneWord 回傳首字母 letter 的符文詞;J / O 與非字母回空字串與 false。
func (t *RuneTable) RuneWord(letter byte) (string, bool) {
	if letter < 'A' || letter > 'Z' {
		return "", false
	}
	w := t.Words[letter-'A']
	return w, w != ""
}

// AcceptsLetter 回報這個按鍵算不算一個符文首字母。
//
// ⚠ 原版對 `J` / `O` 是**默默丟掉**(跳回讀鍵),不是印錯誤 —— 照抄。
func (t *RuneTable) AcceptsLetter(letter byte) bool {
	_, ok := t.RuneWord(letter)
	return ok
}

// Match 把打好的字母對成咒語索引,或 SpellInputCancelled / SpellInputNoSpell。
//
// 原版是氣泡排序 + 逐筆比對,比對時連續重複的輸入字母只吃掉表上一個字元。
// 因為表上沒有任何一筆有重複字母(`validate` 釘住了),
// 「排序後去重再整串比對」與原版逐字元那一段**等價**,而且好讀得多。
func (t *RuneTable) Match(letters []byte) int {
	if len(letters) == 0 {
		return SpellInputCancelled
	}
	want := CanonicalSpellCode(letters)
	for i, code := range t.Codes {
		if code == want {
			return i
		}
	}
	return SpellInputNoSpell
}

// CanonicalSpellCode 把輸入的字母排序並去掉重複,得到查表用的代碼。
func CanonicalSpellCode(letters []byte) string {
	sorted := make([]byte, len(letters))
	copy(sorted, letters)
	// 氣泡排序 —— 與原版同一種寫法,長度最多 4 沒有效能問題。
	for swapped := true; swapped; {
		swapped = false
		for i := 1; i < len(sorted); i++ {
			if sorted[i-1] > sorted[i] {
				sorted[i-1], sorted[i] = sorted[i], sorted[i-1]
				swapped = true
			}
		}
	}
	out := make([]byte, 0, len(sorted))
	for i, c := range sorted {
		if i > 0 && c == sorted[i-1] {
			continue
		}
		out = append(out, c)
	}
	return string(out)
}
