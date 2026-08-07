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
	// SaveFile 是要載入的存檔。空字串時依序試 SAVED.GAM、INIT.GAM。
	SaveFile string
	// WaterTile 是世界地圖全水 chunk 要填的 tile 索引。
	WaterTile byte
}

// Bundle 是載好的素材。任一欄位都可能是 nil —— 呼叫端要能在缺件時繼續跑。
type Bundle struct {
	Tiles     []u5data.Tile
	World     *u5data.WorldMap      // 地表(BRIT.DAT,chunk 組裝)
	Under     *u5data.WorldMap      // 地下世界(UNDER.DAT,256×256 直接存,不分 chunk)
	Scenes    *u5data.SceneSet      // 城鎮 / 民居 / 城堡 / 要塞
	NPCs      *u5data.NPCSet        // 各地點的居民與排程
	Talks     *u5data.TalkSet       // 對話文字 + 展開詞典
	Save      *u5data.Save          // 存檔:名冊、隊伍、時間、位置
	Objects   *u5data.ObjectSet     // 地表的地圖物件(BRIT.OOL)
	UnderObjs *u5data.ObjectSet     // 地下世界的地圖物件(UNDER.OOL)
	Shops     *u5data.ShopSet       // 商店目錄與商店對白
	Items     *u5data.ItemTable     // 48 件裝備的名字
	Creatures *u5data.CreatureTable // 生物名
	Charset   *u5data.Charset
	CJK       *cjk.Font
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

	if u, err := u5data.LoadFlatMap(filepath.Join(opts.GameData, "UNDER.DAT")); err != nil {
		warn = append(warn, fmt.Sprintf("地下世界:%v", err))
	} else {
		b.Under = u
	}

	if sc, err := u5data.LoadSceneSet(opts.GameData); err != nil {
		warn = append(warn, fmt.Sprintf("場景地圖:%v", err))
	} else {
		b.Scenes = sc
	}

	if n, err := u5data.LoadNPCSet(opts.GameData); err != nil {
		warn = append(warn, fmt.Sprintf("NPC:%v", err))
	} else {
		b.NPCs = n
	}

	if t, err := u5data.LoadTalkSet(opts.GameData); err != nil {
		warn = append(warn, fmt.Sprintf("對話:%v", err))
	} else {
		b.Talks = t
	}

	if it, err := u5data.LoadItemTable(opts.GameData); err != nil {
		warn = append(warn, fmt.Sprintf("裝備表:%v", err))
	} else {
		b.Items = it
	}

	if ct, err := u5data.LoadCreatureTable(opts.GameData); err != nil {
		warn = append(warn, fmt.Sprintf("生物名表:%v", err))
	} else {
		b.Creatures = ct
	}

	if sur, und, err := u5data.LoadWorldObjects(opts.GameData); err != nil {
		warn = append(warn, fmt.Sprintf("地圖物件表(.OOL):%v", err))
	} else {
		b.Objects, b.UnderObjs = sur, und
	}

	if b.Talks != nil {
		if sh, err := u5data.LoadShops(opts.GameData, b.Talks.Dict); err != nil {
			warn = append(warn, fmt.Sprintf("商店:%v", err))
		} else {
			b.Shops = sh
		}
	}

	if sv, name, err := loadSave(opts); err != nil {
		warn = append(warn, fmt.Sprintf("存檔:%v", err))
	} else {
		b.Save = sv
		_ = name
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

// loadSave 載入存檔。沒指定就依序試 SAVED.GAM、INIT.GAM ——
// 前者是玩家進行中的遊戲,後者是原版附的新遊戲範本。
// 兩個都沒有也不致命:上層會退回「找一塊陸地站上去」的起始狀態。
func loadSave(opts Options) (*u5data.Save, string, error) {
	candidates := []string{opts.SaveFile}
	if opts.SaveFile == "" {
		candidates = []string{
			filepath.Join(opts.GameData, "SAVED.GAM"),
			filepath.Join(opts.GameData, "INIT.GAM"),
		}
	}
	var firstErr error
	for _, p := range candidates {
		sv, err := u5data.LoadSave(p)
		if err == nil {
			return sv, p, nil
		}
		if firstErr == nil {
			firstErr = err
		}
	}
	return nil, "", firstErr
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
