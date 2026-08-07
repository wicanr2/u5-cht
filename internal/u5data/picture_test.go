package u5data

import (
	"os"
	"path/filepath"
	"testing"
)

// TestEveryPictureSetParses:全部 25 個 `.16` 與 25 個 `.4` 的形狀表都要完全對齊。
//
// 「完全對齊」的定義很硬:把所有 blob 的起點排序之後,**每一個 blob 依
// 寬 × 高 × 深度算出來的大小,必須正好等於它與下一個起點的距離**。
// 少一個位元組或多一個位元組都會被抓到,不需要看圖。
//
// 這條擋的是列寬補齊規則 —— EGA 補到 8 像素、CGA 不補。套錯任一邊,
// 268 個 EGA blob 或 266 個 CGA blob 會整批位移崩掉。
func TestEveryPictureSetParses(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA")
	}
	names, _ := filepath.Glob(filepath.Join(dir, "*.16"))
	more, _ := filepath.Glob(filepath.Join(dir, "*.4"))
	names = append(names, more...)
	shapes, files := 0, 0
	for _, n := range names {
		base := filepath.Base(n)
		if base == "TILES.16" || base == "TILES.4" {
			continue // tileset 不是形狀檔,沒有位移表
		}
		set, err := LoadPictures(n)
		if err != nil {
			t.Errorf("%s: %v", base, err)
			continue
		}
		files++
		for _, p := range set {
			if p != nil {
				shapes++
			}
		}
		if err := checkTightlyPacked(n, set); err != nil {
			t.Errorf("%s: %v", base, err)
		}
	}
	if files < 40 {
		t.Fatalf("只讀了 %d 個形狀檔,預期 48 個", files)
	}
	t.Logf("%d 個檔、%d 個形狀全部解得開且位移緊密相接", files, shapes)
}

// checkTightlyPacked 驗證形狀之間沒有縫也沒有重疊。
//
// 這是「格式對不對」最便宜也最嚴的檢查:檔案是機器產生的,blob 一定緊挨著。
// 只要列寬算式錯一點點,累積誤差就會讓某個 blob 撞進下一個。
func checkTightlyPacked(path string, set PictureSet) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	out, err := Decompress(raw)
	if err != nil {
		return err
	}
	ega := filepath.Ext(path) == ".16"
	depth := depthCGA
	if ega {
		depth = depthEGA
	}
	// 重建位移表並算每個 blob 的區間。
	type span struct{ start, end int }
	var spans []span
	n := int(out[0]) | int(out[1])<<8
	for i := 0; i < n; i++ {
		img := int(out[2+4*i]) | int(out[3+4*i])<<8
		msk := int(out[4+4*i]) | int(out[5+4*i])<<8
		for _, e := range []struct {
			off int
			d   pictureDepth
		}{{img, depth}, {msk, 1}} {
			if e.off == 0 {
				continue
			}
			w := int(out[e.off]) | int(out[e.off+1])<<8
			h := int(out[e.off+2]) | int(out[e.off+3])<<8
			spans = append(spans, span{e.off, e.off + 4 + h*rowBytes(w, e.d)})
		}
	}
	// 起點排序後逐一比對。
	for i := range spans {
		for j := i + 1; j < len(spans); j++ {
			if spans[j].start < spans[i].start {
				spans[i], spans[j] = spans[j], spans[i]
			}
		}
	}
	for i, s := range spans {
		next := len(out)
		if i+1 < len(spans) {
			next = spans[i+1].start
		}
		if s.end != next {
			return errf("第 %d 個 blob 結束於 %d,下一個從 %d 開始(差 %d B)",
				i, s.end, next, next-s.end)
		}
	}
	return nil
}

// TestDungeonPicturesHaveTheExpectedShape:三個地牢圖組的形狀表要一致。
//
// `DNG1/2/3.16` 是**同一套透視切片的三種外觀**(洞穴 / 熔岩 / 磚牆),
// 所以三個檔的形狀表應該完全相同 —— 尺寸與空格的位置都一樣。
// 這條同時證明「28 格裡有 2 格是空的」不是解析錯誤,而是原版就留白。
func TestDungeonPicturesHaveTheExpectedShape(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA")
	}
	var first PictureSet
	for i := 1; i <= 3; i++ {
		set, err := LoadPictures(filepath.Join(dir, dungeonPictureName(i)))
		if err != nil {
			t.Fatal(err)
		}
		if len(set) != DungeonViewShapes {
			t.Fatalf("DNG%d.16 有 %d 個形狀,預期 %d", i, len(set), DungeonViewShapes)
		}
		if first == nil {
			first = set
			continue
		}
		for k := range set {
			a, b := first[k], set[k]
			if (a == nil) != (b == nil) {
				t.Errorf("形狀 %d:DNG1 %v、DNG%d %v", k, a != nil, i, b != nil)
				continue
			}
			if a == nil {
				continue
			}
			if a.Width != b.Width || a.Height != b.Height {
				t.Errorf("形狀 %d:DNG1 是 %d×%d、DNG%d 是 %d×%d",
					k, a.Width, a.Height, i, b.Width, b.Height)
			}
		}
	}
	// 全部切片同高 —— 那是透視走廊的畫布高度。
	for k, p := range first {
		if p == nil {
			continue
		}
		if p.Height != DungeonViewHeight {
			t.Errorf("形狀 %d 高 %d,預期 %d", k, p.Height, DungeonViewHeight)
		}
	}
}
