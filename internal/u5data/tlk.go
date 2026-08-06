package u5data

import (
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
	// Data 是已解碼的位元組(bit7 編碼的版本已清除高位元)。
	// 內部是多個 NUL 結尾字串,欄位語意待 P3 確認。
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
		data := make([]byte, end-e.offset)
		copy(data, raw[e.offset:end])
		if enc == TalkEncodingHighBit {
			for j := range data {
				data[j] &= 0x7F
			}
		}
		tf.Records = append(tf.Records, TalkRecord{NPCIndex: e.index, Offset: e.offset, Data: data})
	}
	return tf, nil
}

// Strings 把一筆記錄切成 NUL 分隔的字串。
// 欄位語意未定(見 TalkFile 的 TODO),所以只回切好的片段,不命名欄位。
func (r TalkRecord) Strings() []string {
	var out []string
	start := 0
	for i, b := range r.Data {
		if b == 0 {
			if i > start {
				out = append(out, string(r.Data[start:i]))
			}
			start = i + 1
		}
	}
	if start < len(r.Data) {
		out = append(out, string(r.Data[start:]))
	}
	return out
}
