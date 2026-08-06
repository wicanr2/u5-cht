package u5data

import (
	"os"
	"path/filepath"
	"testing"
)

// TestTileColorRemapIsInvolution:R↔G bit 互換做兩次應該回到原點。
// 這個性質保證 remap 不是隨手湊出來的表,而真的是一個 bit 對調。
func TestTileColorRemapIsInvolution(t *testing.T) {
	for i := 0; i < 16; i++ {
		if got := tileColorRemap[tileColorRemap[i]]; got != byte(i) {
			t.Errorf("remap(remap(%d)) = %d,應該回到 %d —— 這個表不是單純的 bit 對調", i, got, i)
		}
	}
	// 直接驗算 bit1 與 bit2 互換。
	for i := 0; i < 16; i++ {
		want := byte(i&0b1001 | (i&0b0010)<<1 | (i&0b0100)>>1)
		if tileColorRemap[i] != want {
			t.Errorf("remap[%d] = %d,但 bit1↔bit2 互換應得 %d", i, tileColorRemap[i], want)
		}
	}
}

// TestTilePaletteKnownColors 固定住色彩校正的結論:
// 水藍、山灰、建築黃白在互換下不變;草地與森林被修正成綠色系。
func TestTilePaletteKnownColors(t *testing.T) {
	cases := []struct {
		code byte
		want int // 標準 EGA 色號
		what string
	}{
		{1, 1, "水(藍)——互換下不變"},
		{9, 9, "水波紋(亮藍)——互換下不變"},
		{7, 7, "山(淺灰)——互換下不變"},
		{8, 8, "山陰影(深灰)——互換下不變"},
		{14, 14, "建築(黃)——互換下不變"},
		{15, 15, "建築(白)——互換下不變"},
		{4, 2, "草地:紅 → 綠(修正)"},
		{12, 10, "森林高光:亮紅 → 亮綠(修正)"},
	}
	for _, c := range cases {
		if TilePalette[c.code] != EGAPalette[c.want] {
			t.Errorf("色號 %d(%s):得到 %v,預期 EGA %d = %v",
				c.code, c.what, TilePalette[c.code], c.want, EGAPalette[c.want])
		}
	}
}

func TestParseChunks(t *testing.T) {
	raw := make([]byte, 3*ChunkTiles)
	for i := range raw {
		raw[i] = byte(i % 251)
	}
	chunks, err := ParseChunks(raw, ChunkSide)
	if err != nil {
		t.Fatalf("解析失敗: %v", err)
	}
	if len(chunks) != 3 {
		t.Fatalf("chunk 數 %d,預期 3", len(chunks))
	}
	if chunks[1].At(0, 0) != raw[ChunkTiles] {
		t.Error("第二個 chunk 的起點對不上")
	}
	if _, err := ParseChunks(make([]byte, 100), ChunkSide); err == nil {
		t.Error("非整數倍的資料應該被拒絕")
	}
}

// TestBuildWorldMapConsistency 守住組圖時的一致性檢查:
// 索引表用到的 chunk 數必須等於檔案裡的 chunk 數,否則其中一邊的假設錯了。
func TestBuildWorldMapConsistency(t *testing.T) {
	chunks := make([]Chunk, 2)
	chunks[0].Tiles[0] = 7
	chunks[1].Tiles[0] = 9

	index := make([]byte, WorldChunkIndexSize)
	for i := range index {
		index[i] = ChunkAllWater
	}
	index[0] = 0
	index[1] = 1

	w, err := BuildWorldMap(chunks, index, 1)
	if err != nil {
		t.Fatalf("組圖失敗: %v", err)
	}
	if w.At(0, 0) != 7 {
		t.Errorf("(0,0) = %d,預期 7", w.At(0, 0))
	}
	if w.At(ChunkSide, 0) != 9 {
		t.Errorf("(%d,0) = %d,預期 9", ChunkSide, w.At(ChunkSide, 0))
	}
	if w.At(WorldSide-1, WorldSide-1) != 1 {
		t.Error("全水位置應該被填成 waterTile")
	}

	// chunk 數與索引表用量不符 → 必須報錯
	index[2] = 0
	if _, err := BuildWorldMap(chunks, index, 1); err == nil {
		t.Error("索引表用了 3 個 chunk 但只有 2 個時應該報錯")
	}
	// 索引指向不存在的 chunk → 必須報錯
	index[2] = ChunkAllWater
	index[1] = 99
	if _, err := BuildWorldMap(chunks, index, 1); err == nil {
		t.Error("索引指向不存在的 chunk 時應該報錯")
	}
}

// TestWorldMapFromGameData 是世界地圖的整合驗收。
// 除了「組得起來」,還檢查地形比例合理 —— 若 chunk 切法或索引位移錯了,
// 這些比例會明顯偏掉(例如整片水或整片雜訊)。
func TestWorldMapFromGameData(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過")
	}
	chunks, err := LoadChunks(filepath.Join(dir, "BRIT.DAT"), ChunkSide)
	if err != nil {
		t.Fatalf("載入 BRIT.DAT: %v", err)
	}
	if len(chunks) != 205 {
		t.Errorf("BRIT.DAT chunk 數 %d,實測應為 205", len(chunks))
	}

	ovl, err := os.ReadFile(filepath.Join(dir, "DATA.OVL"))
	if err != nil {
		t.Fatalf("讀 DATA.OVL: %v", err)
	}
	index, err := ReadWorldChunkIndex(ovl)
	if err != nil {
		t.Fatalf("取索引表: %v", err)
	}

	water := 0
	for _, v := range index {
		if v == ChunkAllWater {
			water++
		}
	}
	if water != 51 {
		t.Errorf("索引表有 %d 個全水位置,預期 51(= 256 - 205)", water)
	}

	world, err := BuildWorldMap(chunks, index, 1)
	if err != nil {
		t.Fatalf("組世界地圖: %v", err)
	}

	// 地形比例:水應該是最多的單一 tile,但不該吃掉整張圖。
	counts := map[byte]int{}
	for _, v := range world.Tiles {
		counts[v]++
	}
	total := len(world.Tiles)
	if got := counts[1] * 100 / total; got < 30 || got > 75 {
		t.Errorf("水(tile 1)占 %d%%,超出合理範圍 30–75%% —— chunk 切法或索引位移可能有問題", got)
	}
	if len(counts) < 50 {
		t.Errorf("整張地圖只用了 %d 種 tile,太少 —— 解碼可能錯了", len(counts))
	}
}
