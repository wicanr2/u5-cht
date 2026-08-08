package u5data

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func loadHCS(t *testing.T, name string) *HCSCharset {
	t.Helper()
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	c, err := LoadHCSCharset(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// 16×12 這個切法要用**字形**證明,不是用檔案大小 —— 8×24 也整除 3072。
//
// 判準:'A' 要有「上尖、中間一橫、兩腳分開」的形狀。用 ASCII 圖比對三個特徵:
// 最上面一列窄、中段有一列幾乎全滿(橫槓)、最下面一列左右各有筆畫而中間空。
func TestHCSFontIsSixteenByTwelve(t *testing.T) {
	c := loadHCS(t, "IBM.HCS")
	g, ok := c.Glyph('A')
	if !ok {
		t.Fatal("取不到 'A'")
	}
	art := g.ASCIIArt('#', '.')
	t.Logf("'A' 長這樣:\n%s", art)
	lines := strings.Split(strings.TrimRight(art, "\n"), "\n")
	if len(lines) != HCSGlyphHeight {
		t.Fatalf("有 %d 列,預期 %d 列", len(lines), HCSGlyphHeight)
	}
	count := func(s string) int { return strings.Count(s, "#") }
	// 找出有筆畫的列範圍。
	first, last := -1, -1
	for i, l := range lines {
		if count(l) > 0 {
			if first < 0 {
				first = i
			}
			last = i
		}
	}
	if first < 0 {
		t.Fatal("'A' 整個是空的 —— 切法錯了")
	}
	// 'A' 的橫槓:中段某一列的筆畫數要明顯多於頂端那一列。
	top := count(lines[first])
	best := 0
	for i := first + 1; i < last; i++ {
		if n := count(lines[i]); n > best {
			best = n
		}
	}
	if best <= top {
		t.Errorf("找不到比頂端(%d 點)更寬的橫槓(最寬 %d 點)—— 不像 'A'", top, best)
	}
	// 最下面一列要左右分開(中間有空隙)—— 'A' 的兩腳。
	bottom := lines[last]
	if !strings.Contains(strings.Trim(bottom, "."), ".") {
		t.Errorf("最下面一列沒有中間空隙,不像 'A' 的兩腳:%q", bottom)
	}
}

// 索引就是 ASCII 碼:空白是空的,而可見字元不是空的。
func TestHCSFontIsIndexedByASCII(t *testing.T) {
	for _, name := range []string{"IBM.HCS", "RUNES.HCS"} {
		c := loadHCS(t, name)
		g, _ := c.Glyph(' ')
		for _, b := range g {
			if b != 0 {
				t.Errorf("%s:空白不是空的 —— 索引可能不是 ASCII", name)
				break
			}
		}
		// 字母一定要有筆畫。
		//
		// ⚠ 標點與 `x y z` 在 **`RUNES.HCS` 是空的** —— 不列顛符文表沒有那幾個,
		// 那是資料本身,不是解析錯誤。所以只對字母設要求,而 `IBM.HCS`
		// 額外要求 0x21..0x7A **一格都不空**。
		blankInRunes := "!#$%&()<=>?xyz"
		for ch := byte(0x21); ch <= 0x7A; ch++ {
			if name == "RUNES.HCS" && strings.ContainsRune(blankInRunes, rune(ch)) {
				continue
			}
			g, _ := c.Glyph(ch)
			empty := true
			for _, b := range g {
				if b != 0 {
					empty = false
					break
				}
			}
			if empty {
				t.Errorf("%s:0x%02X(%q)是空的", name, ch, string(rune(ch)))
			}
		}
	}
}

// 每列兩個位元組、MSB 在左 —— 位元順序反了字會左右鏡射。
//
// 判準:大寫 'L' 的左半邊要有筆畫、右上角要是空的。
func TestHCSBitOrderIsMSBFirst(t *testing.T) {
	c := loadHCS(t, "IBM.HCS")
	g, _ := c.Glyph('L')
	leftTop, rightTop := 0, 0
	for y := 0; y < HCSGlyphHeight/2; y++ {
		for x := 0; x < HCSGlyphWidth; x++ {
			if !g.At(x, y) {
				continue
			}
			if x < HCSGlyphWidth/2 {
				leftTop++
			} else {
				rightTop++
			}
		}
	}
	if leftTop <= rightTop {
		t.Errorf("'L' 的上半左邊 %d 點、右邊 %d 點 —— 位元順序可能反了", leftTop, rightTop)
	}
}

// 大小不對要回錯,不能 panic。
func TestParseHCSRejectsWrongSize(t *testing.T) {
	for _, n := range []int{0, hcsFileSize - 1, hcsFileSize + 1, charsetFileSize} {
		if _, err := ParseHCSCharset(make([]byte, n)); err == nil {
			t.Errorf("%d B 竟然解得出 .HCS", n)
		}
	}
}
