package u5data

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// buildTalk 造一份最小的合法對話檔:count + (index, offset) 表 + 資料。
func buildTalk(t *testing.T, records [][]byte, highBit bool) []byte {
	t.Helper()
	count := len(records)
	head := make([]byte, 2+count*4)
	binary.LittleEndian.PutUint16(head[:2], uint16(count))
	offset := len(head)
	var body []byte
	for i, rec := range records {
		base := 2 + i*4
		binary.LittleEndian.PutUint16(head[base:base+2], uint16(i+1)) // NPC index 從 1 起
		binary.LittleEndian.PutUint16(head[base+2:base+4], uint16(offset))
		enc := append([]byte(nil), rec...)
		if highBit {
			for j := range enc {
				enc[j] |= 0x80
			}
		}
		body = append(body, enc...)
		offset += len(enc)
	}
	return append(head, body...)
}

func TestParseTalkHighBit(t *testing.T) {
	raw := buildTalk(t, [][]byte{
		[]byte("Zachariah\x00a stately man\x00"),
		[]byte("Gwenno\x00"),
	}, true)

	tf, err := ParseTalk(raw, TalkEncodingHighBit)
	if err != nil {
		t.Fatalf("解析失敗: %v", err)
	}
	if len(tf.Records) != 2 {
		t.Fatalf("記錄數 %d,預期 2", len(tf.Records))
	}
	if tf.Records[0].NPCIndex != 1 || tf.Records[1].NPCIndex != 2 {
		t.Errorf("NPC 編號讀錯:%d, %d", tf.Records[0].NPCIndex, tf.Records[1].NPCIndex)
	}
	got := tf.Records[0].Strings()
	if len(got) != 2 || got[0] != "Zachariah" || got[1] != "a stately man" {
		t.Errorf("bit7 沒有正確清除或欄位切錯:%q", got)
	}
}

func TestParseTalkShiftJISKeepsHighBit(t *testing.T) {
	// Shift-JIS 的高位元本身有意義,清掉就毀了。「あ」= 0x82 0xA0。
	raw := buildTalk(t, [][]byte{{0x82, 0xA0, 0x00}}, false)
	tf, err := ParseTalk(raw, TalkEncodingShiftJIS)
	if err != nil {
		t.Fatalf("解析失敗: %v", err)
	}
	if got := tf.Records[0].Data[:2]; got[0] != 0x82 || got[1] != 0xA0 {
		t.Errorf("Shift-JIS 的高位元被清掉了:% X", got)
	}
}

// TestParseTalkRejectsInconsistentTable 守住格式假設:索引表尾必須接上第一筆資料。
// 這個檢查存在的理由是「寧可報錯,也不要安靜解出垃圾」。
func TestParseTalkRejectsInconsistentTable(t *testing.T) {
	raw := buildTalk(t, [][]byte{[]byte("x\x00")}, true)
	binary.LittleEndian.PutUint16(raw[4:6], 0x1234) // 把第一筆 offset 弄壞
	if _, err := ParseTalk(raw, TalkEncodingHighBit); err == nil {
		t.Fatal("第一筆 offset 與索引表尾不符時應該報錯")
	}
	if _, err := ParseTalk([]byte{0x01}, TalkEncodingHighBit); err == nil {
		t.Fatal("過短的檔案應該報錯")
	}
}

// TestLoadTalkDOS 用原版 TOWNE.TLK 驗證:48 筆、第一筆是 Zachariah。
// 這兩個數字都來自 2026-08-06 的實測 dump。
func TestLoadTalkDOS(t *testing.T) {
	dir := gameDataDir(t)
	tf, err := LoadTalk(filepath.Join(dir, "TOWNE.TLK"), TalkEncodingHighBit)
	if err != nil {
		t.Fatalf("載入 TOWNE.TLK: %v", err)
	}
	if len(tf.Records) != 48 {
		t.Errorf("TOWNE.TLK 記錄數 %d,實測應為 48", len(tf.Records))
	}
	first := tf.Records[0].Strings()
	if len(first) == 0 || !strings.HasPrefix(first[0], "Zachariah") {
		t.Errorf("第一筆應以 Zachariah 起頭,實得 %q", first)
	}
}

// TestLoadTalkFMTownsJapanese 驗證日文版同結構,且能靠 NPCIndex 與英文版對齊 ——
// 這是翻譯 oracle 成立的前提。設 U5_FMTOWNS 指向 ISO 抽出的 U5_J/ 目錄。
func TestLoadTalkFMTownsJapanese(t *testing.T) {
	dir := os.Getenv("U5_FMTOWNS")
	if dir == "" {
		t.Skip("未設 U5_FMTOWNS(FM Towns ISO 抽出的目錄,需含 U5_J/),跳過")
	}
	jp, err := LoadTalk(filepath.Join(dir, "U5_J", "TOWNE.JPN"), TalkEncodingShiftJIS)
	if err != nil {
		t.Fatalf("載入 TOWNE.JPN: %v", err)
	}
	if len(jp.Records) != 48 {
		t.Errorf("TOWNE.JPN 記錄數 %d,實測應為 48", len(jp.Records))
	}

	en := os.Getenv("U5_GAMEDATA")
	if en == "" {
		return
	}
	eng, err := LoadTalk(filepath.Join(en, "TOWNE.TLK"), TalkEncodingHighBit)
	if err != nil {
		t.Fatalf("載入英文 TOWNE.TLK: %v", err)
	}
	if len(eng.Records) != len(jp.Records) {
		t.Fatalf("英日記錄數不同(%d vs %d),無法逐筆對齊", len(eng.Records), len(jp.Records))
	}
	for i := range eng.Records {
		if eng.Records[i].NPCIndex != jp.Records[i].NPCIndex {
			t.Fatalf("第 %d 筆的 NPC 編號英日不一致(%d vs %d)", i, eng.Records[i].NPCIndex, jp.Records[i].NPCIndex)
		}
	}
}
