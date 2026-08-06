package u5data

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

// 明文訊息檔(2026-08-07 一手驗證;**此前的 `|` 分隔說法是錯的**)
//
//	STORY.DAT     11,679 B  開場故事      NUL 21 · '{' 36 · LF 28 · '_' 654
//	QUESTION.DAT   7,746 B  吉普賽問答    NUL 31 · '{'  5 · LF  3 · '_' 116
//	KARMA.DAT        761 B  業報訊息      NUL  6
//	MISCMSG.DAT    2,745 B  系統訊息      NUL 48 · LF 77
//	ENDMSG.DAT       786 B  結局訊息      NUL 11 · LF 25
//	SHOPPE.DAT    10,135 B  商店對白      NUL 195 + **大量高位元組(詞典 token)**
//	LOOK2.DAT      3,622 B  觀察敘述      NUL 218 + 0x01–0x1F 控制碼 → 格式不同,另案
//	SIGNS.DAT      8,364 B  城鎮招牌      u16 offset 表 + 用字元畫的框線 → 另案
//
// ⚠ **更正**:一開始記載「`|` 分隔記錄」,那是錯的 —— `STORY.DAT` 裡 0x7C 出現 **0 次**。
// 錯誤來源是當初用 `strings … | tr '\n' '|'` 檢視輸出,`|` 是那個 `tr` 自己加的,
// 卻被當成檔案內容。教訓:一手資料(原始位元組)贏二手推論(工具輸出)。
//
// 三個已確證的標記:
//   - **NUL (0x00)**:記錄分隔
//   - `{`:段落 / 換頁起始(只出現在 STORY 與 QUESTION)
//   - `_`:**英文斷字提示**(`be_gin`、`mys_te_ri_ous`)。中文化時一律移除
const (
	textRecordSep  = 0x00
	textPageMark   = '{'
	textHyphenHint = '_'
	// TextTokenBase 是詞典 token 的起始位元組:≥ 0x80 代表「這是一個詞典代碼」。
	TextTokenBase = 0x80
)

// TextFile 是一份明文訊息檔。
type TextFile struct {
	Records []TextRecord
}

// TextRecord 是一筆訊息。
type TextRecord struct {
	// Index 是這筆訊息在檔案中的序號(翻譯表的 key 就用它)。
	Index int
	// Raw 是原始位元組(保留 token 與所有標記)。
	Raw []byte
	// Page 表示這筆訊息以 '{' 開頭。
	Page bool
}

// Text 回傳可讀的英文:去掉換頁標記與斷字提示。
// **詞典 token 不展開**,而是保留成 `<XX>` —— 見 TextDictionary 的說明。
func (r TextRecord) Text() string {
	var b strings.Builder
	for i, c := range r.Raw {
		switch {
		case c >= TextTokenBase:
			fmt.Fprintf(&b, "<%02X>", c)
		case c == textHyphenHint:
			// 斷字提示:丟掉
		case c == textPageMark && i == 0:
			// 換頁標記:丟掉
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

// HasTokens 回報這筆記錄是否含詞典 token(含的話就還不能直接拿去翻譯)。
func (r TextRecord) HasTokens() bool {
	for _, c := range r.Raw {
		if c >= TextTokenBase {
			return true
		}
	}
	return false
}

// ParseText 依 NUL 切開明文訊息檔。
func ParseText(raw []byte) (*TextFile, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("檔案是空的")
	}
	if !bytes.Contains(raw, []byte{textRecordSep}) {
		return nil, fmt.Errorf("找不到 NUL 分隔 —— 這可能不是明文訊息檔")
	}
	tf := &TextFile{}
	for i, part := range bytes.Split(raw, []byte{textRecordSep}) {
		if len(part) == 0 {
			continue
		}
		rec := make([]byte, len(part))
		copy(rec, part)
		tf.Records = append(tf.Records, TextRecord{
			Index: i,
			Raw:   rec,
			Page:  rec[0] == textPageMark,
		})
	}
	if len(tf.Records) == 0 {
		return nil, fmt.Errorf("切不出任何記錄")
	}
	return tf, nil
}

// LoadText 讀取並解析一份明文訊息檔。
func LoadText(path string) (*TextFile, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("讀取 %s: %w", path, err)
	}
	tf, err := ParseText(raw)
	if err != nil {
		return nil, fmt.Errorf("解析 %s: %w", path, err)
	}
	return tf, nil
}

// HyphenHintCount 數斷字提示總數(譯文側應為 0,可當回歸檢查)。
func (t *TextFile) HyphenHintCount() int {
	n := 0
	for _, r := range t.Records {
		n += bytes.Count(r.Raw, []byte{textHyphenHint})
	}
	return n
}

// TokenCount 數詞典 token 總數。
func (t *TextFile) TokenCount() int {
	n := 0
	for _, r := range t.Records {
		for _, c := range r.Raw {
			if c >= TextTokenBase {
				n++
			}
		}
	}
	return n
}

// 詞典壓縮(部分破解,2026-08-07)
//
// SHOPPE.DAT(以及推測其他文字來源)用**常用詞字典**壓縮:位元組 ≥ 0x80 不是文字,
// 而是一個詞的代碼。實例:
//
//	"Thanks\x86nothing!\""              → \x86 展開後是 "for" → Thanks for nothing!
//	"…ready\x83buy something!\""        → \x83 = "to"        → ready to buy something!
//
// 詞典本身在 **DOS 版 DATA.OVL offset 0x104C**,是連續的 NUL 結尾字串:
//
//	the thou of to and that for in is have with thee this not my it me but dost know
//	be was Blackthorn from thy one are here many Lord am we they he would art on …
//
// 共 **118 個詞**(index 0–117);緊接其後就是檔名表(PROPORT.PCS、TILES.16 …),
// 所以詞典到 117 為止。
//
// ⚠ **token → index 的精確映射還沒定**,所以本套件不展開 token:
//   - `\x86` 需要 index 6,而 0x104C 起算的第 6 個詞正是 "for" ✓
//   - 但 `\xD7` / `\xDE` 需要 index 87 / 94,實際在 index 77 / 84 —— **固定差 10**
//   - token 空間 0x80–0xFF 有 128 個,詞典只有 118 個 ⇒ 有 10 個 token 不是詞
//     (推測是控制碼:換行、玩家名代入之類),這很可能就是那個 10 的來源
//
// 硬猜下去只會得到似通順但錯的譯文,所以留給 P3:FM Towns 的 `WORRIORS.EXP` 可反編譯,
// 裡面必然有展開 token 的字串輸出函式,讀它就有確定答案(見 docs/re/01)。
const TextDictionaryOffset = 0x104C

// ReadTextDictionary 從 DOS 版 DATA.OVL 取出常用詞字典。
// 回傳的順序就是檔案順序;**索引與 token 的對應關係尚未確定**(見上方說明)。
func ReadTextDictionary(dataOVL []byte, n int) ([]string, error) {
	if n <= 0 {
		n = 118
	}
	if len(dataOVL) <= TextDictionaryOffset {
		return nil, fmt.Errorf("DATA.OVL 只有 %d B,取不到 0x%X 的字典", len(dataOVL), TextDictionaryOffset)
	}
	words := make([]string, 0, n)
	i := TextDictionaryOffset
	for len(words) < n {
		j := bytes.IndexByte(dataOVL[i:], 0)
		if j < 0 || j > 24 {
			break
		}
		words = append(words, string(dataOVL[i:i+j]))
		i += j + 1
	}
	if len(words) < n {
		return words, fmt.Errorf("只讀到 %d 個詞(要 %d)——字典邊界可能不對", len(words), n)
	}
	// 正確性檢查:第 0 個詞應該是 "the"、第 6 個應該是 "for"(已實測)。
	if words[0] != "the" || words[6] != "for" {
		return words, fmt.Errorf(
			"字典開頭不對(得到 %q … %q,預期 \"the\" … \"for\")——offset 0x%X 可能只對 DOS 版成立",
			words[0], words[6], TextDictionaryOffset)
	}
	return words, nil
}
