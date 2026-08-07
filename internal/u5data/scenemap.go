package u5data

import (
	"fmt"
	"image"
	"os"
)

// 場景地圖(城鎮 / 城堡 / 民居 / 要塞)
//
// 已確認的事實:
//   - 檔案:TOWNE.DAT / CASTLE.DAT / KEEP.DAT / DWELLING.DAT,**各 16,384 B**
//     (DATA.OVL 0x2636 就列著這四個檔名)
//   - **列寬 32**:反編譯的移動函式 sub_86C 用 `byte_3F789[32 * dy + dx]` 定址場景地圖緩衝
//     (docs/re/02)
//   ⇒ 16,384 / (32×32) = **每檔 16 張地圖**
//   - 哪一張對應哪個地點與樓層:見 SceneSet(規則出自 sub_5C8)與 Location.FloorMin/Max
//     (樓層範圍用梯子拓撲反推,docs/re/03 §8)。四個檔各 16 張,恰好被 8 棟建築整除。
const (
	// SceneSide 是場景地圖的邊長(tile 數)。
	SceneSide = 32
	// SceneTiles 是一張場景地圖的 tile 數。
	SceneTiles = SceneSide * SceneSide
	// ScenesPerFile 是每個場景檔內含的地圖數。
	ScenesPerFile = 16
	// SceneFileSize 是場景檔的大小。
	SceneFileSize = SceneTiles * ScenesPerFile // 16,384
)

// SceneMap 是一張 32×32 的場景地圖。
type SceneMap struct {
	Tiles [SceneTiles]byte
}

// At 回報 (x, y) 的 tile 索引。
func (s *SceneMap) At(x, y int) byte {
	if x < 0 || x >= SceneSide || y < 0 || y >= SceneSide {
		return 0
	}
	return s.Tiles[y*SceneSide+x]
}

// ParseSceneMaps 把一個場景檔切成 16 張 32×32 地圖。
func ParseSceneMaps(raw []byte) ([]SceneMap, error) {
	if len(raw) != SceneFileSize {
		return nil, fmt.Errorf("場景檔 %d B,預期 %d B(%d 張 %d×%d)",
			len(raw), SceneFileSize, ScenesPerFile, SceneSide, SceneSide)
	}
	out := make([]SceneMap, ScenesPerFile)
	for i := range out {
		copy(out[i].Tiles[:], raw[i*SceneTiles:(i+1)*SceneTiles])
	}
	return out, nil
}

// LoadSceneMaps 讀取並解析一個場景檔。
func LoadSceneMaps(path string) ([]SceneMap, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("讀取場景檔 %s: %w", path, err)
	}
	maps, err := ParseSceneMaps(raw)
	if err != nil {
		return nil, fmt.Errorf("解析 %s: %w", path, err)
	}
	return maps, nil
}

// RenderSceneMaps 把場景地圖排成一張圖,供目視驗收(切對了會看到建築與道路)。
func RenderSceneMaps(scenes []SceneMap, tiles []Tile, cols int) (*image.NRGBA, error) {
	if len(tiles) == 0 {
		return nil, fmt.Errorf("沒有 tileset,無法算繪")
	}
	if cols <= 0 {
		cols = 4
	}
	rows := (len(scenes) + cols - 1) / cols
	side := SceneSide * TileSize
	img := image.NewNRGBA(image.Rect(0, 0, cols*side, rows*side))
	for i := range scenes {
		ox, oy := (i%cols)*side, (i/cols)*side
		for ty := 0; ty < SceneSide; ty++ {
			for tx := 0; tx < SceneSide; tx++ {
				idx := int(scenes[i].At(tx, ty))
				if idx >= len(tiles) {
					idx = 0
				}
				t := &tiles[idx]
				for y := 0; y < TileSize; y++ {
					for x := 0; x < TileSize; x++ {
						img.SetNRGBA(ox+tx*TileSize+x, oy+ty*TileSize+y, TilePalette[t.At(x, y)&0x0F])
					}
				}
			}
		}
	}
	return img, nil
}
