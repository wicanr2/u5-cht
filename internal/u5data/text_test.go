package u5data

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseTextBasics(t *testing.T) {
	// NUL 分隔(不是 '|' —— 那個說法已更正)
	tf, err := ParseText([]byte("{From no_where, smoky wisps\x00Hello there\x00"))
	if err != nil {
		t.Fatalf("解析失敗: %v", err)
	}
	if len(tf.Records) != 2 {
		t.Fatalf("記錄數 %d,預期 2", len(tf.Records))
	}
	if !tf.Records[0].Page {
		t.Error("第一筆以 '{' 開頭,Page 應為 true")
	}
	if got := tf.Records[0].Text(); got != "From nowhere, smoky wisps" {
		t.Errorf("Text() = %q —— 換頁標記與斷字提示應被移除", got)
	}
	if tf.HyphenHintCount() != 1 {
		t.Errorf("斷字提示數 %d,預期 1", tf.HyphenHintCount())
	}
}

// TestTextTokensArePreservedNotGuessed:token → 詞的映射還沒定,
// 所以必須原樣標示成 <XX>,不可以拿「差 10」的錯映射硬展開。
func TestTextTokensArePreservedNotGuessed(t *testing.T) {
	tf, err := ParseText([]byte("Thanks\x86nothing!\"\x00"))
	if err != nil {
		t.Fatalf("解析失敗: %v", err)
	}
	r := tf.Records[0]
	if !r.HasTokens() {
		t.Error("這筆含 0x86,HasTokens 應為 true")
	}
	if got := r.Text(); !strings.Contains(got, "<86>") {
		t.Errorf("Text() = %q,token 應保留成 <86>", got)
	}
	if tf.TokenCount() != 1 {
		t.Errorf("token 數 %d,預期 1", tf.TokenCount())
	}
}

func TestParseTextRejectsNoSeparator(t *testing.T) {
	if _, err := ParseText([]byte("no separator here")); err == nil {
		t.Error("沒有 NUL 分隔時應該報錯")
	}
	if _, err := ParseText(nil); err == nil {
		t.Error("空檔案應該報錯")
	}
}

// TestLoadTextFromGameData 對真素材驗收,並記錄每個檔的 token 用量 ——
// 有 token 的檔案在 token 映射定案前不能進翻譯流程。
func TestLoadTextFromGameData(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過")
	}
	files := []struct {
		name    string
		minRecs int
		wantSub string
	}{
		{"STORY.DAT", 5, "smoky wisps"},
		{"QUESTION.DAT", 5, "gyp"},
		{"KARMA.DAT", 2, "Avatar"},
		{"MISCMSG.DAT", 5, "Mantra"},
		{"ENDMSG.DAT", 2, "Lord British"},
		{"SHOPPE.DAT", 5, "Thanks"},
	}
	for _, f := range files {
		tf, err := LoadText(filepath.Join(dir, f.name))
		if err != nil {
			t.Errorf("%s: %v", f.name, err)
			continue
		}
		if len(tf.Records) < f.minRecs {
			t.Errorf("%s 只切出 %d 筆,預期至少 %d", f.name, len(tf.Records), f.minRecs)
		}
		var joined strings.Builder
		for _, r := range tf.Records {
			joined.WriteString(r.Text())
			joined.WriteByte('\n')
		}
		if !strings.Contains(joined.String(), f.wantSub) {
			t.Errorf("%s 內容找不到 %q —— 可能解錯", f.name, f.wantSub)
		}
		t.Logf("%s:%d 筆,斷字提示 %d,詞典 token %d", f.name, len(tf.Records), tf.HyphenHintCount(), tf.TokenCount())
	}
}

// TestReadTextDictionary 固定住字典位置與開頭內容。
func TestReadTextDictionary(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過")
	}
	ovl, err := os.ReadFile(filepath.Join(dir, "DATA.OVL"))
	if err != nil {
		t.Fatalf("讀 DATA.OVL: %v", err)
	}
	words, err := ReadTextDictionary(ovl, 118)
	if err != nil {
		t.Fatalf("讀字典: %v", err)
	}
	want := []string{"the", "thou", "of", "to", "and", "that", "for"}
	for i, w := range want {
		if words[i] != w {
			t.Errorf("字典第 %d 個是 %q,預期 %q", i, words[i], w)
		}
	}
	// 字典裡應該有這些 U5 專有詞(證明找對了位置)
	joined := strings.Join(words, " ")
	for _, s := range []string{"Blackthorn", "Shadowlords", "Mantra", "British"} {
		if !strings.Contains(joined, s) {
			t.Errorf("字典裡找不到 %q —— 位置可能不對", s)
		}
	}
	// 字典後面緊接檔名表,所以不該把檔名讀進來
	if strings.Contains(joined, ".16") || strings.Contains(joined, ".PCS") {
		t.Error("字典讀太多了,把後面的檔名表也吃進來")
	}
}
