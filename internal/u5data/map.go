package u5data

import (
	"fmt"
	"image"
	"os"
)

// 地圖資料的已驗證事實與待確認項(2026-08-07)
//
// 已知(檔案大小,一手觀察):
//   DOS / FM Towns  BRIT.DAT   52,480 B      地表世界
//   DOS / FM Towns  UNDER.DAT  65,536 B      地底世界
//   PC-98           BRIT.DAT   65,536 B      ← 與 DOS 不同,PC-98 版不省略
//   TOWNE/CASTLE/KEEP/DWELLING.DAT  各 16,384 B
//
// 推導:65,536 = 256×256 tile(每 tile 1 byte)= 完整世界地圖。
// 而 52,480 = 205 × 256,恰好是「256 個 16×16 chunk 裡只存了 205 個」——
// 缺的 51 個推測是全水 chunk(這種省略法在 Ultima 系列常見)。
//
// ⚠ 已被推翻:一開始猜 BRIT.OOL(256 B)是 chunk 索引表(0xFF = 全水),
// 實測 OOL 的值大多是 0、沒有 0xFF,**不是索引表**。
//
// 索引表實際在 DATA.OVL offset 0x3886(見 WorldChunkIndexOffset 的說明)。
const (
	// ChunkSide 是一個地圖 chunk 的邊長(tile 數)。
	ChunkSide = 16
	// ChunkTiles 是一個 chunk 的 tile 數。
	ChunkTiles = ChunkSide * ChunkSide
	// WorldSide 是完整世界地圖的邊長(tile 數)。
	WorldSide = 256
)

// Chunk 是一塊 16×16 的地圖,每個元素是 tile 索引。
type Chunk struct {
	Tiles [ChunkTiles]byte
}

// At 回報 chunk 內 (x, y) 的 tile 索引。
func (c *Chunk) At(x, y int) byte {
	if x < 0 || x >= ChunkSide || y < 0 || y >= ChunkSide {
		return 0
	}
	return c.Tiles[y*ChunkSide+x]
}

// ParseChunks 把地圖資料切成連續的 chunk。side 是 chunk 邊長(探索格式時可改)。
func ParseChunks(raw []byte, side int) ([]Chunk, error) {
	if side <= 0 || side > ChunkSide {
		return nil, fmt.Errorf("chunk 邊長 %d 不支援(目前只處理 ≤ %d)", side, ChunkSide)
	}
	per := side * side
	if len(raw)%per != 0 {
		return nil, fmt.Errorf("資料 %d B 不是 chunk 大小 %d B 的整數倍", len(raw), per)
	}
	n := len(raw) / per
	out := make([]Chunk, n)
	for i := 0; i < n; i++ {
		for y := 0; y < side; y++ {
			for x := 0; x < side; x++ {
				out[i].Tiles[y*ChunkSide+x] = raw[i*per+y*side+x]
			}
		}
	}
	return out, nil
}

// LoadChunks 讀一個地圖檔並切成 chunk。
func LoadChunks(path string, side int) ([]Chunk, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("讀取地圖 %s: %w", path, err)
	}
	return ParseChunks(raw, side)
}

// 世界地圖的 chunk 索引表
//
// WorldChunkIndexOffset 是索引表在 **DOS 版 DATA.OVL** 內的位移。
//
// 怎麼找到的(2026-08-07):BRIT.DAT 是 52,480 B = 205 個 16×16 chunk,而完整世界地圖
// 需要 16×16 = 256 個 chunk 位置 ⇒ 有 51 個位置沒存資料(推測全水)。於是掃過整個
// 48,464 B 的 DATA.OVL,找「恰好 51 個 0xFF、其餘 205 個值互不重複且落在 0..204」
// 的 256 B 區塊 —— **全檔只有一個命中**,就是 0x3886。
//
// 這種唯一性本身就是證據:表要同時滿足「0xFF 的個數 = 缺少的 chunk 數」與
// 「非 0xFF 的值是 0..204 的一個排列」,隨機資料撞出來的機率極低。
//
// ⚠ 這個位移只對 DOS 版 DATA.OVL 成立。FM Towns 版的 U5_E/ 沒有 DATA.OVL,
// 資料放在別處;所以 API 設計成「索引表由呼叫者提供」,不把位移寫死進解碼流程。
const (
	WorldChunkIndexOffset = 0x3886
	WorldChunkIndexSize   = 256
	// ChunkAllWater 在索引表裡代表「這個位置沒有存資料,整塊都是水」。
	ChunkAllWater = 0xFF
	// WorldChunksPerSide 是世界地圖每邊的 chunk 數(16 × 16 chunk = 256×256 tile)。
	WorldChunksPerSide = WorldSide / ChunkSide
)

// ReadWorldChunkIndex 從 DOS 版 DATA.OVL 取出 chunk 索引表。
func ReadWorldChunkIndex(dataOVL []byte) ([]byte, error) {
	end := WorldChunkIndexOffset + WorldChunkIndexSize
	if len(dataOVL) < end {
		return nil, fmt.Errorf("DATA.OVL 只有 %d B,取不到 0x%X 的索引表", len(dataOVL), WorldChunkIndexOffset)
	}
	table := make([]byte, WorldChunkIndexSize)
	copy(table, dataOVL[WorldChunkIndexOffset:end])
	return table, nil
}

// WorldMap 是組好的完整世界地圖(256×256 tile)。
type WorldMap struct {
	Tiles [WorldSide * WorldSide]byte
}

// At 回報 (x, y) 的 tile 索引。
func (w *WorldMap) At(x, y int) byte {
	if x < 0 || x >= WorldSide || y < 0 || y >= WorldSide {
		return 0
	}
	return w.Tiles[y*WorldSide+x]
}

// BuildWorldMap 用 chunk 資料 + 索引表組出完整世界地圖。
// waterTile 是全水 chunk 要填的 tile 索引(索引表標 0xFF 的位置)。
func BuildWorldMap(chunks []Chunk, index []byte, waterTile byte) (*WorldMap, error) {
	if len(index) != WorldChunkIndexSize {
		return nil, fmt.Errorf("索引表 %d B,預期 %d B", len(index), WorldChunkIndexSize)
	}
	// 一致性檢查:索引表引用的 chunk 編號必須都存在,而且非 0xFF 的項數要等於 chunk 數。
	used := 0
	for i, v := range index {
		if v == ChunkAllWater {
			continue
		}
		used++
		if int(v) >= len(chunks) {
			return nil, fmt.Errorf("索引表第 %d 項指向 chunk %d,但只有 %d 個 chunk", i, v, len(chunks))
		}
	}
	if used != len(chunks) {
		return nil, fmt.Errorf(
			"索引表用到 %d 個 chunk,但檔案有 %d 個 —— 索引表位移或 chunk 切法有一邊錯了",
			used, len(chunks))
	}

	w := &WorldMap{}
	for cy := 0; cy < WorldChunksPerSide; cy++ {
		for cx := 0; cx < WorldChunksPerSide; cx++ {
			id := index[cy*WorldChunksPerSide+cx]
			for ty := 0; ty < ChunkSide; ty++ {
				for tx := 0; tx < ChunkSide; tx++ {
					x, y := cx*ChunkSide+tx, cy*ChunkSide+ty
					if id == ChunkAllWater {
						w.Tiles[y*WorldSide+x] = waterTile
					} else {
						w.Tiles[y*WorldSide+x] = chunks[id].At(tx, ty)
					}
				}
			}
		}
	}
	return w, nil
}

// Render 把世界地圖畫成一張圖(256×256 tile × 16 px = 4096×4096)。
func (w *WorldMap) Render(tiles []Tile) (*image.NRGBA, error) {
	if len(tiles) == 0 {
		return nil, fmt.Errorf("沒有 tileset,無法算繪")
	}
	img := image.NewNRGBA(image.Rect(0, 0, WorldSide*TileSize, WorldSide*TileSize))
	for ty := 0; ty < WorldSide; ty++ {
		for tx := 0; tx < WorldSide; tx++ {
			idx := int(w.At(tx, ty))
			if idx >= len(tiles) {
				idx = 0
			}
			t := &tiles[idx]
			ox, oy := tx*TileSize, ty*TileSize
			for y := 0; y < TileSize; y++ {
				for x := 0; x < TileSize; x++ {
					img.SetNRGBA(ox+x, oy+y, TilePalette[t.At(x, y)&0x0F])
				}
			}
		}
	}
	return img, nil
}

// RenderChunks 把 chunk 用 tileset 畫成一張圖,供目視判斷格式對不對 ——
// 地圖若切錯,畫出來會是雜訊;切對了會看到海岸線與地形。
func RenderChunks(chunks []Chunk, tiles []Tile, cols, side int) (*image.NRGBA, error) {
	if len(tiles) == 0 {
		return nil, fmt.Errorf("沒有 tileset,無法算繪")
	}
	if cols <= 0 {
		cols = 16
	}
	rows := (len(chunks) + cols - 1) / cols
	img := image.NewNRGBA(image.Rect(0, 0, cols*side*TileSize, rows*side*TileSize))
	for i := range chunks {
		cx, cy := (i%cols)*side*TileSize, (i/cols)*side*TileSize
		for ty := 0; ty < side; ty++ {
			for tx := 0; tx < side; tx++ {
				idx := int(chunks[i].At(tx, ty))
				if idx >= len(tiles) {
					idx = 0
				}
				t := &tiles[idx]
				ox, oy := cx+tx*TileSize, cy+ty*TileSize
				for y := 0; y < TileSize; y++ {
					for x := 0; x < TileSize; x++ {
						img.SetNRGBA(ox+x, oy+y, TilePalette[t.At(x, y)&0x0F])
					}
				}
			}
		}
	}
	return img, nil
}
