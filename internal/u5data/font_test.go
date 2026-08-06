package u5data

import (
	"os"
	"path/filepath"
	"testing"
)

// gameDataDir 回傳 DOS 版原版資料目錄。沒設就跳過整合測試 ——
// 原版資料是版權素材,不入庫,由開發者自備(見 CLAUDE.md §3.0)。
func gameDataDir(t *testing.T) string {
	t.Helper()
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA(DOS 版 ultima5/ 目錄),跳過需要原版素材的測試")
	}
	if _, err := os.Stat(dir); err != nil {
		t.Skipf("U5_GAMEDATA=%s 讀不到:%v", dir, err)
	}
	return dir
}

func TestParseCharsetRejectsWrongSize(t *testing.T) {
	if _, err := ParseCharset(make([]byte, 512)); err == nil {
		t.Fatal("512 B 的字型檔應該被拒絕(預期 1024 B),卻通過了")
	}
}

func TestGlyphAtAndASCIIArt(t *testing.T) {
	// 一條對角線,用來確認位元順序是 MSB 在左。
	var g Glyph
	for i := 0; i < GlyphHeight; i++ {
		g[i] = 1 << (7 - uint(i))
	}
	for i := 0; i < GlyphHeight; i++ {
		if !g.At(i, i) {
			t.Errorf("(%d,%d) 應該點亮", i, i)
		}
		if g.At(GlyphWidth-1-i, i) && i != GlyphWidth-1-i {
			t.Errorf("(%d,%d) 不該點亮 —— 位元順序反了", GlyphWidth-1-i, i)
		}
	}
	if g.At(-1, 0) || g.At(0, GlyphHeight) {
		t.Error("超出範圍的座標應回 false")
	}
	if art := g.ASCIIArt('#', '.'); len(art) != GlyphHeight*(GlyphWidth+1) {
		t.Errorf("ASCIIArt 長度 %d,預期 %d", len(art), GlyphHeight*(GlyphWidth+1))
	}
}

// TestLoadCharsetIBM 是字型解碼的 oracle:'A' 必須逐位元組等於實測 dump 的結果。
// 這比「能載入不報錯」強得多 —— 索引公式或位元序錯了都會在這裡現形。
func TestLoadCharsetIBM(t *testing.T) {
	dir := gameDataDir(t)
	cs, err := LoadCharset(filepath.Join(dir, "IBM.CH"))
	if err != nil {
		t.Fatalf("載入 IBM.CH: %v", err)
	}
	if len(cs.Glyphs) != charsetGlyphCount {
		t.Fatalf("glyph 數 %d,預期 %d", len(cs.Glyphs), charsetGlyphCount)
	}

	// 2026-08-06 實測 dump 的 'A'(索引即 ASCII 碼 65)。
	wantA := Glyph{0x1E, 0x36, 0x66, 0x7E, 0x66, 0x66, 0xC6, 0x00}
	gotA, ok := cs.Glyph('A')
	if !ok {
		t.Fatal("取不到 'A' 的字形")
	}
	if gotA != wantA {
		t.Errorf("'A' 字形不符 —— 索引不是 ASCII 直對應,或位元序錯了\n實得:\n%s預期:\n%s",
			gotA.ASCIIArt('#', '.'), wantA.ASCIIArt('#', '.'))
	}

	// 空白應該是全暗;若不是,通常表示整批字偏移了。
	if sp, _ := cs.Glyph(' '); sp != (Glyph{}) {
		t.Errorf("空白字元不是全暗,整套字型可能偏移:\n%s", sp.ASCIIArt('#', '.'))
	}
}
