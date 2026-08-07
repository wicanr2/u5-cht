package u5data

import (
	"os"
	"strings"
	"testing"
)

// TestSignsDecodeToReadableText:招牌不是美術,是真的文字。
//
// `SIGNS.DAT` 此前標著「用字元畫的框線,同 u4-cht 的美術內嵌字母問題」。
// 其實只有**框**是字母美術,框裡是大寫 ASCII 加五個符文合字。
// 這條把地表那塊路標解出來當證明。
func TestSignsDecodeToReadableText(t *testing.T) {
	ss := loadSignsOrSkip(t)
	sg, ok := ss.At(0, 0, 0x5F, 0x94)
	if !ok {
		t.Fatal("地表 (0x5F,0x94) 找不到路標")
	}
	lines := sg.Lines()
	joined := strings.Join(lines, "\n")
	for _, want := range []string{"NORTH BRITAIN", "EAST PAWS", "SOUTH TRINSIC"} {
		if !strings.Contains(joined, want) {
			t.Errorf("路標裡沒有 %q:\n%s", want, joined)
		}
	}
}

// TestSignAliasSharesOneBoard:相鄰兩格共用一塊招牌。
//
// 資料裡的手法是:第一筆的內容只有 `0x0A`,渲染器看到它就往後跳 6 個位元組,
// 而那 6 個位元組是 `0x0A` + NUL + 第二筆的表頭 —— 跳完正好落在共用的內容上。
// 這條同時驗證了「跳 6」這個常數:少跳或多跳都會解到垃圾。
func TestSignAliasSharesOneBoard(t *testing.T) {
	ss := loadSignsOrSkip(t)
	a, ok1 := ss.At(28, 0, 0x10, 0x17)
	b, ok2 := ss.At(28, 0, 0x0E, 0x17)
	if !ok1 || !ok2 {
		t.Fatalf("兩格都該有招牌,實際 %v / %v", ok1, ok2)
	}
	ta, tb := strings.Join(a.Lines(), "\n"), strings.Join(b.Lines(), "\n")
	if ta != tb {
		t.Errorf("共用的招牌解出兩種內容:\n--- %d,%d\n%s\n--- %d,%d\n%s",
			a.X, a.Y, ta, b.X, b.Y, tb)
	}
	if !strings.Contains(ta, "BEWARE") || !strings.Contains(ta, "UNUSUAL SIZE") {
		t.Errorf("內容看起來不對:\n%s", ta)
	}
}

// TestSignMacrosExpandToFullRows:巨集一個位元組展開成整列。
//
// 招牌寬 16 欄。巨集(0x29..0x31)每一條都是 16 個字,所以用巨集畫框的招牌
// 每一列剛好對齊。這條擋住「巨集表抄錯一條長度」——那種錯不會讓程式爆掉,
// 只會讓框歪掉,而歪掉的框在測試裡看不出來,除非量長度。
func TestSignMacrosExpandToFullRows(t *testing.T) {
	for i, m := range signMacros {
		if len(m) != SignWidth {
			t.Errorf("巨集 0x%02X 長 %d,預期 %d:%q", signMacroLo+i, len(m), SignWidth, m)
		}
	}
}

// TestEverySignRendersWithoutJunk:全部 78 塊招牌都要解得出可讀內容。
//
// 「有印出東西」不等於「解對了」——把任意位元組當文字印一定印得出東西。
// 所以這條驗的是**沒有非字模的位元組漏出來**:渲染完只該剩 `RUNES.CH`
// 真的畫得出來的槽位,也就是可見 ASCII 加上唯一的裝飾字模 0x0E。
func TestEverySignRendersWithoutJunk(t *testing.T) {
	ss := loadSignsOrSkip(t)
	all := ss.All()
	if len(all) < 50 {
		t.Fatalf("只解出 %d 塊招牌,少得可疑", len(all))
	}
	for _, sg := range all {
		for _, c := range sg.Render() {
			if c.PageBreak || c.NewLine {
				continue
			}
			if c.Ch == signOrnament {
				continue
			}
			if c.Ch < 0x20 || c.Ch > 0x7E {
				t.Errorf("地點 %d (%d,%d) 渲染出控制碼 0x%02X", sg.Location, sg.X, sg.Y, c.Ch)
				break
			}
		}
	}
}

func loadSignsOrSkip(t *testing.T) *SignSet {
	t.Helper()
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	ss, err := LoadSigns(dir)
	if err != nil {
		t.Fatalf("讀 SIGNS.DAT:%v", err)
	}
	return ss
}
