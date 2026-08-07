package u5data

import (
	"strings"
	"path/filepath"
	"encoding/binary"
	"fmt"
	"os"
)

// TalkFile 是一份 NPC 對話檔(DOS 的 .TLK 或 FM Towns 日文版的 .JPN)。
//
// 檔案結構(靜態推導 + 自我一致性驗證,2026-08-06/07):
//
//	u16 recordCount
//	recordCount × { u16 npcIndex; u16 dataOffset }
//	…各筆對話資料…
//
// 支持這個解讀的證據:TOWNE.TLK 開頭是 30 00 | 01 00 c2 00 | 02 00 2b 03 …
// → recordCount = 0x30 = 48;而 2 + 48×4 = 0xC2,**正好等於第一筆的 dataOffset**。
// 表尾與第一筆資料無縫接合,這個自我一致性是主要佐證。
// FM Towns 日文版 TOWNE.JPN 檔頭同結構(30 00 01 00 c2 00 02 00 74 04 …),
// 只有 offset 因日文長度不同而變大 → 兩版可靠 npcIndex 逐筆對齊。
//
// TODO(P3):以 IDA 反編譯的讀取常式(錨點 sub_2C740 / 緩衝 byte_54700,
// 見 docs/re/00-hexrays-p3-verified.md)確認 (index, offset) 的欄位順序與
// 記錄內部的欄位切分(U4 的 .TLK 是固定 12 欄,U5 尚未確認)。在確認前
// 本套件只提供「整筆記錄的位元組」,不假裝知道欄位邊界。
type TalkFile struct {
	// Encoding 是這份檔案的文字編碼方式。
	Encoding TalkEncoding
	// Records 依檔案順序排列。
	Records []TalkRecord
}

// TalkRecord 是單一 NPC 的對話資料。
type TalkRecord struct {
	// NPCIndex 是檔頭索引表給的編號(從 1 起)。英日兩版靠它對齊。
	NPCIndex int
	// Offset 是這筆資料在檔案中的起始位移(除錯與交叉比對用)。
	Offset int
	// Data 是**原始**位元組,尚未展開。內部是多個 NUL 結尾的段落。
	// high-bit 版本要用 Dictionary.Expand 展開(bit7 有設 = 字面,沒設 = 詞典 token)。
	Data []byte
}

// TalkEncoding 標示對話文字的位元組編碼。
type TalkEncoding int

const (
	// TalkEncodingHighBit 是 DOS/英文版:每個位元組的 bit7 被設為 1,清掉即 ASCII 明文。
	TalkEncodingHighBit TalkEncoding = iota
	// TalkEncodingShiftJIS 是 FM Towns 日文版 .JPN:Shift-JIS 原樣(高位元本身有意義,不可清)。
	TalkEncodingShiftJIS
)

func (e TalkEncoding) String() string {
	switch e {
	case TalkEncodingHighBit:
		return "high-bit(DOS/英文)"
	case TalkEncodingShiftJIS:
		return "Shift-JIS(FM Towns 日文)"
	default:
		return fmt.Sprintf("未知(%d)", int(e))
	}
}

// LoadTalk 讀取並解析一份對話檔。
func LoadTalk(path string, enc TalkEncoding) (*TalkFile, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("讀取對話檔 %s: %w", path, err)
	}
	tf, err := ParseTalk(raw, enc)
	if err != nil {
		return nil, fmt.Errorf("解析 %s: %w", path, err)
	}
	return tf, nil
}

// ParseTalk 解析已讀入記憶體的對話檔內容。
func ParseTalk(raw []byte, enc TalkEncoding) (*TalkFile, error) {
	const headerEntrySize = 4
	if len(raw) < 2 {
		return nil, fmt.Errorf("檔案只有 %d B,連記錄數都放不下", len(raw))
	}
	count := int(binary.LittleEndian.Uint16(raw[:2]))
	tableEnd := 2 + count*headerEntrySize
	if count == 0 || tableEnd > len(raw) {
		return nil, fmt.Errorf("記錄數 %d 不合理(索引表會到 %d B,檔案只有 %d B)", count, tableEnd, len(raw))
	}

	type entry struct {
		index  int
		offset int
	}
	entries := make([]entry, 0, count)
	for i := 0; i < count; i++ {
		base := 2 + i*headerEntrySize
		entries = append(entries, entry{
			index:  int(binary.LittleEndian.Uint16(raw[base : base+2])),
			offset: int(binary.LittleEndian.Uint16(raw[base+2 : base+4])),
		})
	}

	// 自我一致性檢查:索引表尾應該正好接上第一筆資料。
	// 這是支持「(index, offset) 而非 (offset, index)」這個解讀的主要證據,
	// 對不上就是格式假設錯了 —— 寧可報錯,也不要安靜地解出垃圾。
	if entries[0].offset != tableEnd {
		return nil, fmt.Errorf(
			"第一筆 offset %d(0x%X)不等於索引表尾 %d(0x%X):"+
				"欄位順序或記錄數的解讀有誤,先回去對 docs/re/ 的讀取常式再改",
			entries[0].offset, entries[0].offset, tableEnd, tableEnd)
	}

	tf := &TalkFile{Encoding: enc, Records: make([]TalkRecord, 0, count)}
	for i, e := range entries {
		end := len(raw)
		if i+1 < len(entries) {
			end = entries[i+1].offset
		}
		if e.offset > len(raw) || end > len(raw) || end < e.offset {
			return nil, fmt.Errorf("第 %d 筆(NPC %d)的範圍 [%d, %d) 超出檔案(%d B)", i, e.index, e.offset, end, len(raw))
		}
		// Data 保留**原始**位元組。high-bit 版本絕對不能在這裡就把 bit7 清掉 ——
		// 那會把「字面文字」和「詞典 token」壓成同一種東西,再也分不開
		// (見 Dictionary 的說明:0x01 是 "the",不是控制字元)。
		data := make([]byte, end-e.offset)
		copy(data, raw[e.offset:end])
		tf.Records = append(tf.Records, TalkRecord{NPCIndex: e.index, Offset: e.offset, Data: data})
	}
	return tf, nil
}

// 對話記錄的前五段。
//
// 展開 48 筆 TOWNE.TLK 之後,**多數**記錄的前五段是這個順序,之後才是
// 「關鍵字 → 回應」的成對資料。
//
// ⚠ 不是每一筆都乾淨:有些記錄的第 0 段名字後面還跟著控制位元組,
// 有些第 2 段是空的而問候語併進了第 1 段。這代表記錄開頭可能還有一個
// 尚未解出的小檔頭(`sub_1C840` 讀進 0x400 B 後交給 sub_1C660,
// 而另一處把資料當成 6 B 一組在掃)。在解出來之前,取欄位要能容忍雜訊。
const (
	TalkFieldName        = 0 // NPC 名字
	TalkFieldDescription = 1 // 「看」到的樣子
	TalkFieldGreeting    = 2 // 打招呼
	TalkFieldJob         = 3 // 問 job / work
	TalkFieldBye         = 4 // 道別
	TalkFixedFields      = 5
)

// Segments 把一筆記錄切成 NUL 分隔的段落(原始位元組,未展開)。
func (r TalkRecord) Segments() [][]byte {
	var out [][]byte
	start := 0
	for i, b := range r.Data {
		if b == 0 {
			if i > start {
				out = append(out, r.Data[start:i])
			} else {
				out = append(out, nil)
			}
			start = i + 1
		}
	}
	if start < len(r.Data) {
		out = append(out, r.Data[start:])
	}
	return out
}

// Strings 把一筆記錄展開成可讀文字。d 可為 nil(那時 token 會留成 <XX>)。
func (r TalkRecord) Strings(d *Dictionary) []string {
	segs := r.Segments()
	out := make([]string, 0, len(segs))
	for _, s := range segs {
		out = append(out, d.Expand(s))
	}
	return out
}

// Field 取一個固定欄位;沒有就回空字串。
func (r TalkRecord) Field(d *Dictionary, i int) string {
	segs := r.Segments()
	if i < 0 || i >= len(segs) {
		return ""
	}
	return d.Expand(segs[i])
}

// TalkFiles 是對話檔名,順序與 SceneFiles / NPCFiles 相同
// (原版 off_55E78,由 `(地點編號-1)/8` 選出 —— 見 sub_1C840)。
var TalkFiles = [4]string{"TOWNE.TLK", "DWELLING.TLK", "CASTLE.TLK", "KEEP.TLK"}

// TalkSet 是四個對話檔加上展開用的詞典。
type TalkSet struct {
	Files [len(TalkFiles)]*TalkFile
	Dict  *Dictionary
}

// LoadTalkSet 讀入四個 .TLK 與 DATA.OVL 的詞典。
func LoadTalkSet(dir string) (*TalkSet, error) {
	s := &TalkSet{}
	var err error
	if s.Dict, err = LoadDictionary(dir); err != nil {
		return nil, err
	}
	for i, name := range TalkFiles {
		if s.Files[i], err = LoadTalk(filepath.Join(dir, name), TalkEncodingHighBit); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// Record 依地點編號與對話號碼取出對話記錄。
//
// 原版 sub_1C840 就是這樣做的:先用 `(地點編號-1)/8` 選檔,再**線性搜尋**
// 索引表找 id 相符的那一筆 —— id 不是陣列下標,所以不能直接用它當索引。
func (s *TalkSet) Record(location, dialogue int) (*TalkRecord, bool) {
	loc, err := LocationByNumber(location)
	if err != nil {
		return nil, false
	}
	tf := s.Files[loc.SceneFile]
	if tf == nil {
		return nil, false
	}
	for i := range tf.Records {
		if tf.Records[i].NPCIndex == dialogue {
			return &tf.Records[i], true
		}
	}
	return nil, false
}

// cleanText 去掉對話腳本的控制位元組,只留可讀文字。
//
// 記錄裡除了文字還混著對話引擎的指令(原始位元組 0x81–0x9F,展開後就是
// 0x01–0x1F 的控制字元)—— `sub_1C3F8` 有一張跳表在處理它們(cases 133/134/
// 140/141/143/144/145-159)。整套腳本語意還沒解,但把控制碼濾掉之後
// 文字本身是完好的,足以拿來翻譯與顯示。
func cleanText(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == '\n':
			b.WriteRune('\n')
		case r < 0x20 || r == 0x7F:
			// 控制碼:丟掉
		default:
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}

// Name 取 NPC 的名字。
//
// 名字是第 0 段開頭那一串文字,後面常跟著控制位元組(有些記錄還多一個引號),
// 所以要在第一個控制碼處切斷。
func (r TalkRecord) Name(d *Dictionary) string {
	raw := d.Expand(r.Field2(TalkFieldName))
	for i, c := range raw {
		if c < 0x20 || c == '"' || c == 0x7F {
			return strings.TrimSpace(raw[:i])
		}
	}
	return strings.TrimSpace(raw)
}

// Field2 回傳某一段的原始位元組。
func (r TalkRecord) Field2(i int) []byte {
	segs := r.Segments()
	if i < 0 || i >= len(segs) {
		return nil
	}
	return segs[i]
}

// Line 取某一段的可讀文字(已去控制碼)。
func (r TalkRecord) Line(d *Dictionary, i int) string {
	return cleanText(d.Expand(r.Field2(i)))
}

// Greeting 取打招呼的話;有些記錄的第 2 段只有控制碼,那就退回第 3 段(自我介紹)。
func (r TalkRecord) Greeting(d *Dictionary) string {
	if s := r.Line(d, TalkFieldGreeting); s != "" {
		return s
	}
	return r.Line(d, TalkFieldJob)
}
