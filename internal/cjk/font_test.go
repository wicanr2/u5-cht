package cjk

import (
	"os"
	"testing"
)

func loadTestFont(t *testing.T) *Font {
	t.Helper()
	prefix := os.Getenv("U5_CJK_FONT")
	if prefix == "" {
		prefix = "../../assets/fonts/eten-15"
	}
	if _, err := os.Stat(prefix + ".png"); err != nil {
		t.Skipf("找不到 %s.png(先跑 tools/dev.sh font 15),跳過", prefix)
	}
	f, err := Load(prefix)
	if err != nil {
		t.Fatalf("載入字型: %v", err)
	}
	return f
}

func TestLoadFontAndLookup(t *testing.T) {
	f := loadTestFont(t)
	w, h := f.GlyphSize()
	if w != 16 || h != 15 {
		t.Errorf("字模 %d×%d,預期 16×15", w, h)
	}
	if f.Count() < 10000 {
		t.Errorf("只收 %d 字,倚天漢字區應有 13,000 以上", f.Count())
	}

	// 專案會用到的字與全形標點都要在(全形標點來自 SPCFONT,漏帶就會缺)
	for _, r := range "創世紀命運勇士不列顛尼亞聖者美德真言月門汝卿，。！？「」『』（）" {
		if !f.Has(r) {
			t.Errorf("缺字 %q —— 檢查烘字時是否漏帶 SPCFONT", r)
		}
	}

	rect, ok := f.Glyph('中')
	if !ok {
		t.Fatal("取不到「中」")
	}
	if rect.Dx() != w || rect.Dy() != h {
		t.Errorf("glyph 矩形 %v 尺寸不對", rect)
	}
	if !rect.In(f.Mask.Bounds()) {
		t.Errorf("glyph 矩形 %v 超出 atlas 範圍 %v", rect, f.Mask.Bounds())
	}
}

func TestMissingRunesIgnoresASCII(t *testing.T) {
	f := loadTestFont(t)
	// ASCII 走原版 8×8 字型,不該被算成缺字
	if got := f.MissingRunes("abc,!? 123"); len(got) != 0 {
		t.Errorf("ASCII 被誤判為缺字:%q", got)
	}
}

func TestLoadRejectsMissingFiles(t *testing.T) {
	if _, err := Load("/nonexistent/font"); err == nil {
		t.Error("檔案不存在時應該報錯")
	}
}
