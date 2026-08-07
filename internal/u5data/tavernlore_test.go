package u5data

import (
	"os"
	"testing"
)

func loadLore(t *testing.T) *TavernLoreTable {
	t.Helper()
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	lore, err := LoadTavernLore(dir)
	if err != nil {
		t.Fatal(err)
	}
	return lore
}

// 26 個主題正好蓋住整個尋寶清單:八德 + 七座地牢 + 三信物 + 三碎片 + 四項雜項。
//
// ⚠ 這條不只在驗「表讀得到」,而是在驗**表偏沒偏**:五張表首尾相接,
// 位移錯一格,關鍵字就不會是四個小寫字母、價格就不會落在 25..250。
func TestTavernLoreCoversTheWholeTreasureList(t *testing.T) {
	lore := loadLore(t)
	// 八德的四字母縮寫。
	virtues := []string{"hone", "comp", "valo", "just", "sacr", "hono", "spir", "humi"}
	for i, want := range virtues {
		if lore.Entries[i].Keyword != want {
			t.Errorf("第 %d 題是 %q,預期 %q(八德要在最前面)", i, lore.Entries[i].Keyword, want)
		}
	}
	// 三信物與三碎片各自連號 —— 這是「表沒錯位」最強的一條。
	for i, want := range []string{"crow", "scep", "amul", "fals", "hatr", "cowa"} {
		if got := lore.Entries[15+i].Keyword; got != want {
			t.Errorf("第 %d 題是 %q,預期 %q", 15+i, got, want)
		}
	}
	// 每一題都要有人名與地名。
	for i, e := range lore.Entries {
		if e.Who == "" || e.Where == "" {
			t.Errorf("第 %d 題(%s)缺人名或地名:%q / %q", i, e.Keyword, e.Who, e.Where)
		}
	}
}

// 打「honesty」要問得到誠實 —— 比對是子字串,不是完全相等。
func TestLoreMatchesBySubstring(t *testing.T) {
	lore := loadLore(t)
	cases := map[string]string{
		"honesty":  "hone",
		"HONESTY":  "hone", // 不分大小寫
		"Hone":     "hone",
		"crown":    "crow",
		"sceptre":  "scep",
		"amulet":   "amul",
		"deceit":   "dece",
		"hythloth": "hyth",
	}
	for in, want := range cases {
		i := lore.Match(in)
		if i < 0 {
			t.Errorf("打 %q 沒有比中任何主題", in)
			continue
		}
		if lore.Entries[i].Keyword != want {
			t.Errorf("打 %q 比到 %q,預期 %q", in, lore.Entries[i].Keyword, want)
		}
	}
	// 由前往後掃,第一個命中的就算 —— 「honour」含 hono,但 hone 在它前面且不相符,
	// 所以應該比到 hono。
	if i := lore.Match("honour"); i < 0 || lore.Entries[i].Keyword != "hono" {
		t.Errorf("打 honour 比到 %d", i)
	}
	// 完全沒關係的字要回 −1(原版印「That, I cannot help thee with.」)。
	for _, in := range []string{"", "xyzzy", "beer"} {
		if i := lore.Match(in); i >= 0 {
			t.Errorf("打 %q 竟然比到第 %d 題(%s)", in, i, lore.Entries[i].Keyword)
		}
	}
}

// 價格:三信物 200、三碎片 250 —— 這兩組同值,是價目表對得上的獨立證據。
func TestLorePricesGroupByWhatYouAsk(t *testing.T) {
	lore := loadLore(t)
	for _, kw := range []string{"crow", "scep", "amul"} {
		i := lore.Match(kw)
		if lore.Entries[i].Price != 200 {
			t.Errorf("%s 的價格是 %d,預期 200", kw, lore.Entries[i].Price)
		}
	}
	for _, kw := range []string{"fals", "hatr", "cowa"} {
		i := lore.Match(kw)
		if lore.Entries[i].Price != 250 {
			t.Errorf("%s 的價格是 %d,預期 250", kw, lore.Entries[i].Price)
		}
	}
	// 最便宜的是靈性(25)—— 表偏一格這條就不成立。
	if got := lore.Entries[lore.Match("spir")].Price; got != 25 {
		t.Errorf("spir 的價格是 %d,預期 25", got)
	}
}

// 13 個地名不重複,而且 26 個主題的索引都在範圍內(ParseTavernLore 已擋,這裡是回歸)。
func TestLorePlacesAreDistinct(t *testing.T) {
	lore := loadLore(t)
	seen := map[string]bool{}
	for i, p := range lore.Places {
		if p == "" {
			t.Errorf("第 %d 個地名是空的", i)
		}
		if seen[p] {
			t.Errorf("地名 %q 重複", p)
		}
		seen[p] = true
	}
}

// 檔案太短要回錯,不能 panic。
func TestTavernLoreRejectsShortFiles(t *testing.T) {
	for _, n := range []int{0, 0x4C84, tavernLorePrice} {
		if _, err := ParseTavernLore(make([]byte, n)); err == nil {
			t.Errorf("%d B 的檔案竟然解得出情報表", n)
		}
	}
}
