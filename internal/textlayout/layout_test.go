package textlayout

import (
	"strings"
	"testing"
)

func TestAdvanceGate(t *testing.T) {
	if Advance('A') != ASCIIAdvance {
		t.Errorf("ASCII 寬度 %d,預期 %d", Advance('A'), ASCIIAdvance)
	}
	if Advance('中') != CJKAdvance {
		t.Errorf("CJK 寬度 %d,預期 %d", Advance('中'), CJKAdvance)
	}
	// CJK 必須剛好是 ASCII 的兩倍,混排欄位才對得齊
	if CJKAdvance != ASCIIAdvance*2 {
		t.Errorf("CJK(%d)不是 ASCII(%d)的兩倍", CJKAdvance, ASCIIAdvance)
	}
	if got := Width("汝好 abc"); got != 2*CJKAdvance+4*ASCIIAdvance {
		t.Errorf("Width(\"汝好 abc\") = %d,預期 %d", got, 2*CJKAdvance+4*ASCIIAdvance)
	}
}

// TestWrapNeverExceedsWidth 是防「文字溢框」的核心測試:
// Wrap 用的 Advance 必須與 Draw 用的完全相同,所以每一行的 Width 都不能超過上限。
// 這個坑 kb 記過 —— 量測與繪製不同步時,症狀看起來完全像「譯文太長」。
func TestWrapNeverExceedsWidth(t *testing.T) {
	samples := []string{
		"汝已抵達不列顛尼亞。此地由不列顛王治理,然黑刺爵僭越其位。",
		"Thou hast strayed far from the path of the Avatar.",
		"混合 mixed 中英 text 一起 wrap 的情況,看斷行會不會爆。",
		strings.Repeat("聖", 100),
		strings.Repeat("a", 200),
		"短",
		"",
	}
	for _, maxW := range []int{CJKAdvance, 64, 128, 320, 624} {
		for _, s := range samples {
			for _, line := range Wrap(s, maxW) {
				if w := Width(line); w > maxW {
					t.Errorf("maxWidth=%d 時,行 %q 的寬度 %d 超出上限", maxW, line, w)
				}
			}
		}
	}
}

func TestWrapKeepsAllRunes(t *testing.T) {
	s := "汝已抵達不列顛尼亞 the Avatar returns"
	joined := strings.Join(Wrap(s, 128), "")
	// 英文在空白處斷行會吃掉那個空白,所以比對時把空白去掉
	want := strings.ReplaceAll(s, " ", "")
	got := strings.ReplaceAll(joined, " ", "")
	if got != want {
		t.Errorf("Wrap 掉字了:\n得到 %q\n預期 %q", got, want)
	}
}

func TestWrapHandlesNewline(t *testing.T) {
	lines := Wrap("第一行\n第二行", 320)
	if len(lines) != 2 || lines[0] != "第一行" || lines[1] != "第二行" {
		t.Errorf("換行沒有分成兩行:%q", lines)
	}
}
