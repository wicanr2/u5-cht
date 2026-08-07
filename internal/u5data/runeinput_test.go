package u5data

import (
	"os"
	"strings"
	"testing"
)

// loadRunes 讀真檔;沒有原版資料就跳過。
func loadRunes(t *testing.T) (*RuneTable, *SpellTable) {
	t.Helper()
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	rt, err := LoadRuneTable(dir)
	if err != nil {
		t.Fatal(err)
	}
	st, err := LoadSpells(dir)
	if err != nil {
		t.Fatal(err)
	}
	return rt, st
}

// 48 個代碼必須逐一等於咒語名的**首字母排序**。
//
// 這是整張表的意義所在:0 筆例外才算解對。逐筆比對比抽樣強得多 ——
// 抽樣抽到 `IL`(In Lor)看不出有排序,要抽到 `HR`(Rel Hur)才看得出來。
func TestSpellCodesAreSortedInitials(t *testing.T) {
	rt, st := loadRunes(t)
	for i := 0; i < SpellCount; i++ {
		want := CanonicalSpellCode(initialsOf(st.Spells[i].Name))
		if rt.Codes[i] != want {
			t.Errorf("第 %d 個咒語 %q:代碼是 %q,首字母排序是 %q",
				i, st.Spells[i].Name, rt.Codes[i], want)
		}
	}
	// 代碼互不重複 —— 重複的話 Match 永遠只會回傳靠前的那一個。
	seen := map[string]int{}
	for i, c := range rt.Codes {
		if j, dup := seen[c]; dup {
			t.Errorf("代碼 %q 同時屬於第 %d 與第 %d 個咒語", c, j, i)
		}
		seen[c] = i
	}
}

// initialsOf 取每個詞的首字母。
func initialsOf(name string) []byte {
	var out []byte
	for _, w := range strings.Fields(name) {
		out = append(out, w[0])
	}
	return out
}

// 24 個符文詞:A–Z 去掉 J 與 O,而且每個詞的首字母就是它自己的索引。
func TestRuneWordsCoverEveryLetterButJAndO(t *testing.T) {
	rt, _ := loadRunes(t)
	n := 0
	for c := byte('A'); c <= 'Z'; c++ {
		w, ok := rt.RuneWord(c)
		if c == 'J' || c == 'O' {
			if ok {
				t.Errorf("%c 竟然有符文詞 %q", c, w)
			}
			continue
		}
		if !ok {
			t.Errorf("%c 沒有符文詞", c)
			continue
		}
		if w[0] != c {
			t.Errorf("%c 的符文詞是 %q,首字母對不上", c, w)
		}
		n++
	}
	if n != RuneWordCount {
		t.Errorf("符文詞有 %d 個,預期 %d 個", n, RuneWordCount)
	}
	// 非字母一律不收。
	for _, c := range []byte{'@', '[', '0', ' ', 0} {
		if rt.AcceptsLetter(c) {
			t.Errorf("0x%02X 竟然收得進來", c)
		}
	}
}

// 打字母的順序不影響結果 —— 排序就是為了這個。
func TestMatchIgnoresTheOrderYouTypeIn(t *testing.T) {
	rt, st := loadRunes(t)
	cases := []struct {
		in   string
		want string // 預期咒語名
	}{
		{"IL", "In Lor"},
		{"LI", "In Lor"}, // 反過來打
		{"RH", "Rel Hur"},
		{"HR", "Rel Hur"},
		{"AXC", "An Xen Cor"}, // 名字的順序
		{"CXA", "An Xen Cor"}, // 完全倒過來
		{"IVPY", "In Vas P Y"},
		{"YPVI", "In Vas P Y"},
		{"M", "Mani"},
		{"MM", "Mani"}, // ★ 連續重複只算一個
	}
	for _, c := range cases {
		got := rt.Match([]byte(c.in))
		if got < 0 {
			t.Errorf("打 %q 得到 %d,預期咒語 %q", c.in, got, c.want)
			continue
		}
		if st.Spells[got].Name != c.want {
			t.Errorf("打 %q 得到 %q,預期 %q", c.in, st.Spells[got].Name, c.want)
		}
	}
}

// 兩種非咒語結果各自要對:什麼都沒打 → −1,湊不出來 → −2。
func TestMatchReportsTheTwoFailureKinds(t *testing.T) {
	rt, _ := loadRunes(t)
	if got := rt.Match(nil); got != SpellInputCancelled {
		t.Errorf("什麼都沒打得到 %d,預期 %d", got, SpellInputCancelled)
	}
	for _, in := range []string{"B", "ZZ", "BCDE", "QQ"} {
		if got := rt.Match([]byte(in)); got != SpellInputNoSpell {
			t.Errorf("打 %q 得到 %d,預期 %d", in, got, SpellInputNoSpell)
		}
	}
}

// 去重只吃**連續**重複 —— 排序過的輸入裡重複一定相鄰,所以等價;
// 但這條把「先排序再去重」的順序釘住,免得誰改成先去重再排序。
func TestCanonicalCodeSortsThenDedupes(t *testing.T) {
	cases := map[string]string{
		"IL":   "IL",
		"LI":   "IL",
		"MM":   "M",
		"AAB":  "AB",
		"ABA":  "AB", // 不相鄰的重複,排序之後才變相鄰
		"YPVI": "IPVY",
		"":     "",
	}
	for in, want := range cases {
		if got := CanonicalSpellCode([]byte(in)); got != want {
			t.Errorf("%q → %q,預期 %q", in, got, want)
		}
	}
}
