package u5data

import (
	"os"
	"testing"
)

func loadDict(t *testing.T) *Dictionary {
	t.Helper()
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	d, err := LoadDictionary(dir)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// TestDictionarySpotChecks 固定幾個由反組譯確認過的槽位。
func TestDictionarySpotChecks(t *testing.T) {
	d := loadDict(t)
	// 這些值來自 FM Towns 執行檔的 dword_41990(用 token 值直接索引的指標表)。
	for tok, want := range map[byte]string{
		1: "the", 3: "of", 9: "in", 31: "many", 91: "these", 105: "man", 128: "work",
	} {
		if got := d.Word(tok); got != want {
			t.Errorf("token %d = %q,預期 %q", tok, got, want)
		}
	}
	// 空槽必須是空的 —— 這 10 個洞正是 DOS 密集清單與 token 值之間的差。
	for _, h := range dictHoles {
		if got := d.Word(byte(h)); got != "" {
			t.Errorf("token %d 應該是空槽,卻是 %q", h, got)
		}
	}
}

// TestExpandTLK 用三句已經人工對過的句子鎖住 .TLK 的展開規則。
//
// 這三句同時涵蓋了三種情況:單一 token 夾在字面文字中間、連續三個 token、
// 以及 token 後面接標點。空白處理錯的話這三句至少會壞一句。
func TestExpandTLK(t *testing.T) {
	d := loadDict(t)
	lit := func(s string) []byte { // 把字面文字編回 .TLK 的 high-bit 形式
		b := make([]byte, len(s))
		for i := 0; i < len(s); i++ {
			b[i] = s[i] | 0x80
		}
		return b
	}
	cat := func(parts ...[]byte) []byte {
		var out []byte
		for _, p := range parts {
			out = append(out, p...)
		}
		return out
	}
	cases := []struct {
		raw  []byte
		want string
	}{
		{cat(lit("I study"), []byte{1}, lit("stars.")), "I study the stars."},
		{cat(lit("a stately, white-haired"), []byte{105, 3, 31}, lit("years.")),
			"a stately, white-haired man of many years."},
		{cat(lit("Welcome,"), []byte{9, 91}, lit("dark times.")),
			"Welcome, in these dark times."},
	}
	for _, c := range cases {
		if got := d.Expand(c.raw); got != c.want {
			t.Errorf("展開得到 %q,預期 %q", got, c.want)
		}
	}
}

// TestExpandDAT:.DAT 的極性與槽位偏移和 .TLK 不同,用真實的 SHOPPE.DAT 鎖住。
func TestExpandDAT(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	d := loadDict(t)
	tf, err := LoadText(dir + "/SHOPPE.DAT")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		`Thanks for nothing!"`,
		`Come back when you're ready to buy something!"`,
		`Harrumph!"`,
		`Be off with ye, then..."`,
	}
	for i, w := range want {
		if i >= len(tf.Records) {
			t.Fatalf("SHOPPE.DAT 只有 %d 筆", len(tf.Records))
		}
		if got := tf.Records[i].Expand(d); got != w {
			t.Errorf("第 %d 筆展開成 %q,預期 %q", i, got, w)
		}
	}
}

// TestNoUnexpandedTokens:整份 SHOPPE.DAT 展開後不該剩下任何 <XX>。
// 剩下就代表槽表或偏移有洞 —— 那會在翻譯時變成一句句缺字的英文。
//
// 只驗 SHOPPE.DAT。LOOK2.DAT 也用同一套壓縮,但它前面有幾筆不是文字
// (看起來是索引表,被目前的切法當成記錄了),那是 TextFile 的切法問題,
// 不是詞典的問題 —— 別把兩件事綁在同一個測試裡。
func TestNoUnexpandedTokens(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	d := loadDict(t)
	tf, err := LoadText(dir + "/SHOPPE.DAT")
	if err != nil {
		t.Fatal(err)
	}
	var bad []int
	for i, r := range tf.Records {
		if containsToken(r.Expand(d)) {
			bad = append(bad, i)
		}
	}
	if len(bad) > 0 {
		t.Errorf("SHOPPE.DAT 有 %d/%d 筆展開後仍有未解的 token(第 %v 筆)",
			len(bad), len(tf.Records), bad)
	}
}

func containsToken(s string) bool {
	for i := 0; i+3 < len(s); i++ {
		if s[i] == '<' && s[i+3] == '>' {
			return true
		}
	}
	return false
}
