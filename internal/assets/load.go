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
	"github.com/wicanr2/u5-cht/internal/i18n"
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
	// Display 是顯示模式(原版四個 `.DRV` 的其中一個)。預設 EGA。
	Display u5data.DisplayMode
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
	Combat    *u5data.CombatMapSet  // 地表的戰鬥地圖(BRIT.CBT)
	Stats     *u5data.CombatStats   // 戰鬥數值(怪物三圍、裝備防禦/射程/類別)
	Spells    *u5data.SpellTable    // 咒語表(名稱 / 圈數 / 藥草 / 可施法場合)
	Runes     *u5data.RuneTable     // 符文詞與咒語代碼(施法 / 調藥的輸入法)
	Lore      *u5data.TavernLoreTable // 酒館的打聽消息(26 個主題)
	Dungeons  *u5data.DungeonSet    // 八座地牢的地圖(DUNGEON.DAT)
	Moons     *u5data.MoonPhases    // 月相表(DATA.OVL)
	Look2     *u5data.LookTable     // Look 指令的敘述表(LOOK2.DAT)
	Signs     *u5data.SignSet       // 城鎮招牌與墓碑(SIGNS.DAT)
	WindDelay *u5data.WindDelay     // 航行延遲表(DATA.OVL)
	DngRooms  *u5data.CombatMapSet  // 地牢房間(DUNGEON.CBT)
	Objects   *u5data.ObjectSet     // 地表的地圖物件(BRIT.OOL)
	UnderObjs *u5data.ObjectSet     // 地下世界的地圖物件(UNDER.OOL)
	Shops     *u5data.ShopSet       // 商店目錄與商店對白
	Items     *u5data.ItemTable     // 48 件裝備的名字(長名)
	// SpecialItems 是短名字表:U 指令的 22 個特殊道具 + 48 件裝備的縮寫名
	// (同一批裝備在原版有長短兩套名字,見 `docs/re/56`)。
	SpecialItems *u5data.SpecialItemTable
	Creatures *u5data.CreatureTable // 生物名
	Charset   *u5data.Charset
	// DungeonViews 是 DNG1/2/3.16 的透視切片,索引 0..2 對應三種外觀。
	DungeonViews []u5data.PictureSet
	// DungeonItems 是 ITEMS.16(走廊裡的梯子、寶箱、噴泉、陷阱)。
	DungeonItems u5data.PictureSet
	// Story 是 STORY.DAT,IntroArt 是六個 STORY*.16。
	Story    *u5data.TextFile
	// Question 是 QUESTION.DAT —— 建角時吉普賽的三十筆問答。
	Question *u5data.TextFile
	Misc     *u5data.TextFile
	// EndMsg 是 ENDMSG.DAT —— 結局那一幕的十一段文字。
	EndMsg   *u5data.TextFile
	// MiscMaps 是 MISCMAPS.DAT —— 四張 11×11 的石室。
	MiscMaps *u5data.MiscMapSet
	IntroArt []u5data.PictureSet
	CJK      *cjk.Font
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

	// tileset:**DOS 的 `TILES.16` 優先**。
	//
	// 從 2026-08-07 起 DOS 的 LZW 壓縮已破(`internal/u5data/lzw.go`),
	// 而兩份資料的解碼結果**逐位元組相同**(`TestDOSTilesMatchFMTowns`)——
	// 所以沒有理由再要求玩家準備 FM Towns 光碟。
	// FM Towns 那條保留當後援與 oracle。
	if tiles, err := u5data.LoadTileSetFor(opts.GameData, opts.Display); err == nil {
		b.Tiles = tiles
	} else if opts.FMTowns != "" {
		paths := []string{"EGA0.TIL", "EGA1.TIL", "EGA2.TIL", "EGA3.TIL"}
		for i := range paths {
			paths[i] = filepath.Join(opts.FMTowns, paths[i])
		}
		if t2, err2 := u5data.LoadFMTownsTileSet(paths); err2 != nil {
			warn = append(warn, fmt.Sprintf("tileset:DOS %v;FM Towns %v", err, err2))
		} else {
			b.Tiles = t2
		}
	} else {
		warn = append(warn, fmt.Sprintf("tileset:%v", err))
	}

	// 地牢透視圖組(DNG1/2/3.16)—— 三種外觀。
	for i := 1; i <= u5data.DungeonThemes; i++ {
		name := fmt.Sprintf("DNG%d.16", i)
		set, err := u5data.LoadPictures(filepath.Join(opts.GameData, name))
		if err != nil {
			warn = append(warn, fmt.Sprintf("%s:%v", name, err))
			continue
		}
		b.DungeonViews = append(b.DungeonViews, set)
	}
	// 開場:STORY.DAT 的二十段文字 + 六個插圖檔。
	if tf, err := u5data.LoadText(filepath.Join(opts.GameData, "STORY.DAT")); err != nil {
		warn = append(warn, fmt.Sprintf("STORY.DAT:%v", err))
	} else {
		b.Story = tf
	}
	if tf, err := u5data.LoadText(filepath.Join(opts.GameData, "QUESTION.DAT")); err != nil {
		warn = append(warn, fmt.Sprintf("QUESTION.DAT:%v", err))
	} else {
		b.Question = tf
	}
	if tf, err := u5data.LoadText(filepath.Join(opts.GameData, "MISCMSG.DAT")); err != nil {
		warn = append(warn, fmt.Sprintf("MISCMSG.DAT:%v", err))
	} else {
		b.Misc = tf
	}
	if tf, err := u5data.LoadText(filepath.Join(opts.GameData, "ENDMSG.DAT")); err != nil {
		warn = append(warn, fmt.Sprintf("ENDMSG.DAT:%v", err))
	} else {
		b.EndMsg = tf
	}
	if mm, err := u5data.LoadMiscMaps(filepath.Join(opts.GameData, "MISCMAPS.DAT")); err != nil {
		warn = append(warn, fmt.Sprintf("MISCMAPS.DAT:%v", err))
	} else {
		b.MiscMaps = mm
	}
	for _, name := range u5data.IntroStoryFiles {
		set, err := u5data.LoadPictures(filepath.Join(opts.GameData, name))
		if err != nil {
			warn = append(warn, fmt.Sprintf("%s:%v", name, err))
			set = nil
		}
		b.IntroArt = append(b.IntroArt, set)
	}

	if set, err := u5data.LoadPictures(filepath.Join(opts.GameData, "ITEMS.16")); err != nil {
		warn = append(warn, fmt.Sprintf("ITEMS.16:%v", err))
	} else {
		b.DungeonItems = set
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

	// 短名字表(U 指令的特殊道具 + 裝備的縮寫名,見 `docs/re/56`)。
	if si, err := u5data.LoadSpecialItems(opts.GameData); err != nil {
		warn = append(warn, fmt.Sprintf("特殊道具短名表:%v", err))
	} else {
		b.SpecialItems = si
	}

	if ct, err := u5data.LoadCreatureTable(opts.GameData); err != nil {
		warn = append(warn, fmt.Sprintf("生物名表:%v", err))
	} else {
		b.Creatures = ct
	}

	if cs, err := u5data.LoadCombatStats(opts.GameData); err != nil {
		warn = append(warn, fmt.Sprintf("戰鬥數值:%v", err))
	} else {
		b.Stats = cs
	}

	if sp, err := u5data.LoadSpells(opts.GameData); err != nil {
		warn = append(warn, fmt.Sprintf("咒語表:%v", err))
	} else {
		b.Spells = sp
	}

	if rt, err := u5data.LoadRuneTable(opts.GameData); err != nil {
		warn = append(warn, fmt.Sprintf("符文詞表:%v", err))
	} else {
		b.Runes = rt
	}

	if tl, err := u5data.LoadTavernLore(opts.GameData); err != nil {
		warn = append(warn, fmt.Sprintf("酒館情報表:%v", err))
	} else {
		b.Lore = tl
	}

	if mp, err := u5data.LoadMoonPhases(opts.GameData); err != nil {
		warn = append(warn, fmt.Sprintf("月相表:%v", err))
	} else {
		b.Moons = mp
	}

	if lk, err := u5data.LoadLook(opts.GameData); err != nil {
		warn = append(warn, fmt.Sprintf("Look 敘述表(LOOK2.DAT):%v", err))
	} else {
		b.Look2 = lk
	}

	if sn, err := u5data.LoadSigns(opts.GameData); err != nil {
		warn = append(warn, fmt.Sprintf("招牌(SIGNS.DAT):%v", err))
	} else {
		b.Signs = sn
	}

	if wd, err := u5data.LoadWindDelay(opts.GameData); err != nil {
		warn = append(warn, fmt.Sprintf("航行延遲表:%v", err))
	} else {
		b.WindDelay = wd
	}

	if dg, err := u5data.LoadDungeons(opts.GameData); err != nil {
		warn = append(warn, fmt.Sprintf("地牢地圖(DUNGEON.DAT):%v", err))
	} else {
		b.Dungeons = dg
	}

	if dr, err := u5data.LoadDungeonRooms(opts.GameData); err != nil {
		warn = append(warn, fmt.Sprintf("地牢房間(DUNGEON.CBT):%v", err))
	} else {
		b.DngRooms = dr
	}

	if cm, err := u5data.LoadCombatMaps(filepath.Join(opts.GameData, "BRIT.CBT")); err != nil {
		warn = append(warn, fmt.Sprintf("戰鬥地圖(BRIT.CBT):%v", err))
	} else {
		b.Combat = cm
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
			// 商店對白走中譯覆蓋層。掛在載入處而不是 u5data 裡面 ——
			// 資料層只讀原版檔,譯文是上層的事。
			sh.Translate = i18n.Shop
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
