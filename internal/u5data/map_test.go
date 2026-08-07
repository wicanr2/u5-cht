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

// TestGrassIsGreenAfterNormalisation:正規化之後草地要是綠的。
//
// 這條取代了原本的「TilePalette 查表值等於某個 EGA 色號」——
// 色號正規化搬到載入時之後,那種查表比對只是在驗一個恆等式。
// 現在驗的是**真的資料**:從 `TILES.16` 讀出來的草地 tile,
// 主色必須落在綠色系(2 綠 / 10 亮綠),而水必須落在藍色系(1 / 9)。
//
// 認錯色號順序時,草會變紅、水不變 —— 所以草是那個會說話的樣本。
func TestGrassIsGreenAfterNormalisation(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA")
	}
	tiles, err := LoadDOSTileSet(dir)
	if err != nil {
		t.Fatal(err)
	}
	dominant := func(i int) byte {
		var n [16]int
		for _, v := range tiles[i].Pix {
			n[v&0x0F]++
		}
		// 黑色不算 —— 多數 tile 的底色都是黑,那不帶資訊。
		best := byte(1)
		for c := 2; c < 16; c++ {
			if n[c] > n[best] {
				best = byte(c)
			}
		}
		return best
	}
	green := map[byte]bool{2: true, 10: true}
	blue := map[byte]bool{1: true, 9: true}
	for _, i := range []int{5, 6} { // 草地與更多草
		if c := dominant(i); !green[c] {
			t.Errorf("tile %d 的主色是 %d,草地應該是綠色系(2 / 10)", i, c)
		}
	}
	for _, i := range []int{1, 2} { // 深水與淺水
		if c := dominant(i); !blue[c] {
			t.Errorf("tile %d 的主色是 %d,水應該是藍色系(1 / 9)", i, c)
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
