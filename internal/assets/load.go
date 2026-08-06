// Package assets 把「載入原版素材」這件事收在一處,讓遊戲執行檔與驗收工具共用同一套邏輯。
//
// 素材一律用原版(CLAUDE.md §3.0);缺件時**優雅降級並明說**,不拿自製素材充數。
// 所以 Load 不會因為少一樣東西就整個失敗,而是回傳警告清單讓呼叫端印出來。
package assets

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/wicanr2/u5-cht/internal/cjk"
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// Options 是素材的來源路徑。
type Options struct {
	// GameData 是 DOS 版原版資料目錄(BRIT.DAT / DATA.OVL / IBM.CH …)。
	GameData string
	// FMTowns 是 FM Towns ISO 抽出的 U5_E 目錄(未壓縮 tileset 來源)。
	FMTowns string
	// FontPrefix 是倚天 atlas 前綴(tools/dev.sh font 15 產生)。
	FontPrefix string
	// WaterTile 是世界地圖全水 chunk 要填的 tile 索引。
	WaterTile byte
}

// Bundle 是載好的素材。任一欄位都可能是 nil —— 呼叫端要能在缺件時繼續跑。
type Bundle struct {
	Tiles   []u5data.Tile
	World   *u5data.WorldMap
	Charset *u5data.Charset
	CJK     *cjk.Font
}

// Load 依 opts 載入素材,回傳 bundle 與「哪些沒載到」的警告。
func Load(opts Options) (*Bundle, []string) {
	if opts.WaterTile == 0 {
		opts.WaterTile = 1 // tile 1 是水(實測:占世界地圖 37.6%,且填充全水 chunk 後輪廓正確)
	}
	b := &Bundle{}
	var warn []string

	if cs, err := u5data.LoadCharset(filepath.Join(opts.GameData, "IBM.CH")); err != nil {
		warn = append(warn, fmt.Sprintf("原版 8×8 字型:%v", err))
	} else {
		b.Charset = cs
	}

	if opts.FMTowns != "" {
		paths := []string{"EGA0.TIL", "EGA1.TIL", "EGA2.TIL", "EGA3.TIL"}
		for i := range paths {
			paths[i] = filepath.Join(opts.FMTowns, paths[i])
		}
		if tiles, err := u5data.LoadFMTownsTileSet(paths); err != nil {
			warn = append(warn, fmt.Sprintf("tileset:%v", err))
		} else {
			b.Tiles = tiles
		}
	}

	if w, err := loadWorld(opts.GameData, opts.WaterTile); err != nil {
		warn = append(warn, fmt.Sprintf("世界地圖:%v", err))
	} else {
		b.World = w
	}

	if opts.FontPrefix != "" {
		if f, err := cjk.Load(opts.FontPrefix); err != nil {
			warn = append(warn, fmt.Sprintf("中文字型:%v", err))
		} else {
			b.CJK = f
		}
	}
	return b, warn
}

func loadWorld(gameData string, water byte) (*u5data.WorldMap, error) {
	chunks, err := u5data.LoadChunks(filepath.Join(gameData, "BRIT.DAT"), u5data.ChunkSide)
	if err != nil {
		return nil, err
	}
	ovl, err := os.ReadFile(filepath.Join(gameData, "DATA.OVL"))
	if err != nil {
		return nil, err
	}
	index, err := u5data.ReadWorldChunkIndex(ovl)
	if err != nil {
		return nil, err
	}
	return u5data.BuildWorldMap(chunks, index, water)
}

// FindLandStart 從地圖中央往外找一個非水的落點。
//
// 這是「正常玩家路徑」的最小版本:玩家不該開場就泡在海裡。
// 完整的可達性檢查(flood-fill 連通分量、城鎮與落點同分量)在 P4 —— 那個坑 u2-cht 踩過:
// 回歸測試全綠,但新角色被放在只連城堡的 12 格小島上 soft-lock。
func FindLandStart(w *u5data.WorldMap, waterTile byte) (int, int) {
	const c = u5data.WorldSide / 2
	if w == nil {
		return c, c
	}
	for r := 0; r < u5data.WorldSide/2; r++ {
		for dy := -r; dy <= r; dy++ {
			for dx := -r; dx <= r; dx++ {
				if absInt(dx) != r && absInt(dy) != r {
					continue // 只掃這一圈的邊
				}
				x, y := wrapCoord(c+dx), wrapCoord(c+dy)
				// 用原版執行檔取出的通行表判斷,不是自己列的清單(docs/re/02)。
				if t := int(w.At(x, y)); !u5data.TileBlocksWalking(t) && !u5data.TileIsWater(t) {
					return x, y
				}
			}
		}
	}
	return c, c
}

func wrapCoord(v int) int {
	v %= u5data.WorldSide
	if v < 0 {
		v += u5data.WorldSide
	}
	return v
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
