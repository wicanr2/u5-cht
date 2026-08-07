// u5dump 把原版資料解碼成看得見的東西(PNG / JSON),當作解碼器的驗收工具。
//
//	u5dump tiles-fmtowns <U5_E 目錄> <out.png>   FM Towns EGA0-3.TIL → tile sheet
//	u5dump charset <IBM.CH> <out.png>            原版 8×8 字型 → 字元表
//	u5dump tlk <檔案> [--sjis] [--n 3]           對話檔 → 前 n 筆的欄位
//	u5dump tiles-raw <U5_E 目錄> <out.bin>       FM Towns tileset → DOS 4bpp 佈局
//	                                             (破 TILES.16 壓縮時的對答案基準)
//
// 每個子命令都應該產出「能用眼睛或 diff 判對錯」的東西 —— 解碼器沒有 oracle 就等於沒驗過。
package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wicanr2/u5-cht/internal/assets"
	"github.com/wicanr2/u5-cht/internal/game"
	"github.com/wicanr2/u5-cht/internal/render"
	"github.com/wicanr2/u5-cht/internal/u5data"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "tiles-fmtowns":
		err = cmdTilesFMTowns(os.Args[2:])
	case "tiles-raw":
		err = cmdTilesRaw(os.Args[2:])
	case "charset":
		err = cmdCharset(os.Args[2:])
	case "tlk":
		err = cmdTLK(os.Args[2:])
	case "map":
		err = cmdMap(os.Args[2:])
	case "world":
		err = cmdWorld(os.Args[2:])
	case "text":
		err = cmdText(os.Args[2:])
	case "scene":
		err = cmdScene(os.Args[2:])
	case "scenemaps":
		err = cmdSceneMaps(os.Args[2:])
	case "save":
		err = cmdSave(os.Args[2:])
	case "conv":
		err = cmdConv(os.Args[2:])
	case "npc":
		err = cmdNPC(os.Args[2:])
	case "town":
		err = cmdTown(os.Args[2:])
	default:
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "錯誤:%v\n", err)
		os.Exit(1)
	}
}

const usage = `u5dump — 原版資料解碼驗收工具

  u5dump tiles-fmtowns <U5_E 目錄> <out.png>   FM Towns EGA0-3.TIL → tile sheet(512 tile)
  u5dump tiles-raw     <U5_E 目錄> <out.bin>   同上但輸出 DOS 4bpp 佈局(65,536 B)
  u5dump charset       <IBM.CH>    <out.png>   原版 8×8 字型 → 字元表
  u5dump tlk           <檔案> [--sjis] [--n N] 對話檔 → 前 N 筆欄位(預設 3)
  u5dump map           <地圖檔> <U5_E 目錄> <out.png> [--side N] [--cols N] [--max N]
  u5dump world         <gamedata 目錄> <U5_E 目錄> <out.png> [--water N]
  u5dump text          <檔案> [--n N]          明文訊息檔 → 前 N 筆(預設 5)
  u5dump scene         <gamedata> <U5_E> <out.png> [--font 前綴] [--at X Y]
  u5dump scenemaps     <場景檔.DAT> <U5_E> <out.png>  16 張 32×32 場景地圖
  u5dump town          <gamedata> <U5_E> <地名> <out.png>  依原版地點表進城,畫出每一層
                                               遊戲畫面 headless 截圖(純 CPU,不需 GPU)
`

func fmTownsTilePaths(dir string) []string {
	return []string{
		filepath.Join(dir, "EGA0.TIL"),
		filepath.Join(dir, "EGA1.TIL"),
		filepath.Join(dir, "EGA2.TIL"),
		filepath.Join(dir, "EGA3.TIL"),
	}
}

func cmdTilesFMTowns(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("用法:u5dump tiles-fmtowns <U5_E 目錄> <out.png>")
	}
	tiles, err := u5data.LoadFMTownsTileSet(fmTownsTilePaths(args[0]))
	if err != nil {
		return err
	}
	if err := writePNG(args[1], u5data.TileSheet(tiles, 32)); err != nil {
		return err
	}
	fmt.Printf("✓ %d 個 tile → %s(32 個一列,16×16 每格)\n", len(tiles), args[1])
	fmt.Println("  驗收方式:與 DOSBox 跑原版的畫面逐格比對(水、草、樹、城牆、人物…)")
	return nil
}

func cmdTilesRaw(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("用法:u5dump tiles-raw <U5_E 目錄> <out.bin>")
	}
	tiles, err := u5data.LoadFMTownsTileSet(fmTownsTilePaths(args[0]))
	if err != nil {
		return err
	}
	var out []byte
	for i := range tiles {
		out = append(out, tiles[i].Pack4bpp()...)
	}
	if err := os.WriteFile(args[1], out, 0o644); err != nil {
		return err
	}
	fmt.Printf("✓ %d B → %s\n", len(out), args[1])
	fmt.Println("  這是破 DOS TILES.16 壓縮時的對答案基準(解壓器輸出應與此逐位元組相同)")
	return nil
}

func cmdCharset(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("用法:u5dump charset <IBM.CH> <out.png>")
	}
	cs, err := u5data.LoadCharset(args[0])
	if err != nil {
		return err
	}
	const cols = 16
	rows := (len(cs.Glyphs) + cols - 1) / cols
	img := image.NewNRGBA(image.Rect(0, 0, cols*u5data.GlyphWidth, rows*u5data.GlyphHeight))
	for i, g := range cs.Glyphs {
		ox, oy := (i%cols)*u5data.GlyphWidth, (i/cols)*u5data.GlyphHeight
		for y := 0; y < u5data.GlyphHeight; y++ {
			for x := 0; x < u5data.GlyphWidth; x++ {
				if g.At(x, y) {
					img.SetNRGBA(ox+x, oy+y, u5data.EGAPalette[15])
				}
			}
		}
	}
	if err := writePNG(args[1], img); err != nil {
		return err
	}
	fmt.Printf("✓ %d 個字形 → %s\n", len(cs.Glyphs), args[1])
	return nil
}

func cmdTLK(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("用法:u5dump tlk <檔案> [--sjis] [--n N] [--dict gamedata]")
	}
	enc := u5data.TalkEncodingHighBit
	n := 3
	dictDir := filepath.Dir(args[0]) // 詞典在同一個資料目錄的 DATA.OVL
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--dict":
			if i+1 < len(args) {
				dictDir = args[i+1]
				i++
			}
		case "--sjis":
			enc = u5data.TalkEncodingShiftJIS
		case "--n":
			if i+1 < len(args) {
				n, _ = strconv.Atoi(args[i+1])
				i++
			}
		}
	}
	tf, err := u5data.LoadTalk(args[0], enc)
	if err != nil {
		return err
	}
	var dict *u5data.Dictionary
	if enc == u5data.TalkEncodingHighBit {
		if dict, err = u5data.LoadDictionary(dictDir); err != nil {
			fmt.Fprintf(os.Stderr, "⚠ 讀不到詞典(%v)—— token 會留成 <XX>\n", err)
		}
	}
	fmt.Printf("%s:%d 筆,編碼 %s\n", args[0], len(tf.Records), tf.Encoding)
	for i, r := range tf.Records {
		if i >= n {
			break
		}
		fmt.Printf("\n-- 第 %d 筆(NPC %d,offset 0x%X,%d B)--\n", i, r.NPCIndex, r.Offset, len(r.Data))
		for j, s := range r.Strings(dict) {
			if j >= 8 {
				fmt.Println("   …")
				break
			}
			fmt.Printf("   [%d] %q\n", j, s)
		}
	}
	return nil
}

func writePNG(path string, img image.Image) error {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// --- map 子命令(附加於此以保持單檔;若再長就拆檔)---

func cmdMap(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("用法:u5dump map <地圖檔> <U5_E 目錄> <out.png> [--side 16] [--cols 16] [--max N]")
	}
	side, cols, max := 16, 16, 0
	for i := 3; i < len(args); i++ {
		if i+1 >= len(args) {
			break
		}
		switch args[i] {
		case "--side":
			side, _ = strconv.Atoi(args[i+1])
		case "--cols":
			cols, _ = strconv.Atoi(args[i+1])
		case "--max":
			max, _ = strconv.Atoi(args[i+1])
		}
	}
	tiles, err := u5data.LoadFMTownsTileSet(fmTownsTilePaths(args[1]))
	if err != nil {
		return err
	}
	chunks, err := u5data.LoadChunks(args[0], side)
	if err != nil {
		return err
	}
	fmt.Printf("%s → %d 個 %d×%d chunk\n", args[0], len(chunks), side, side)
	if max > 0 && max < len(chunks) {
		chunks = chunks[:max]
	}
	img, err := u5data.RenderChunks(chunks, tiles, cols, side)
	if err != nil {
		return err
	}
	if err := writePNG(args[2], img); err != nil {
		return err
	}
	fmt.Printf("✓ %s(%d×%d px)\n", args[2], img.Bounds().Dx(), img.Bounds().Dy())
	fmt.Println("  驗收方式:切對了會看到海岸線與地形;切錯了是雜訊")
	return nil
}

// cmdWorld 組出完整 256×256 世界地圖(chunk 資料 + DATA.OVL 的索引表)。
func cmdWorld(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("用法:u5dump world <gamedata 目錄> <U5_E 目錄> <out.png> [--water N]")
	}
	water := 1
	for i := 3; i+1 < len(args); i++ {
		if args[i] == "--water" {
			water, _ = strconv.Atoi(args[i+1])
		}
	}
	tiles, err := u5data.LoadFMTownsTileSet(fmTownsTilePaths(args[1]))
	if err != nil {
		return err
	}
	chunks, err := u5data.LoadChunks(filepath.Join(args[0], "BRIT.DAT"), u5data.ChunkSide)
	if err != nil {
		return err
	}
	ovl, err := os.ReadFile(filepath.Join(args[0], "DATA.OVL"))
	if err != nil {
		return err
	}
	index, err := u5data.ReadWorldChunkIndex(ovl)
	if err != nil {
		return err
	}
	world, err := u5data.BuildWorldMap(chunks, index, byte(water))
	if err != nil {
		return err
	}
	img, err := world.Render(tiles)
	if err != nil {
		return err
	}
	if err := writePNG(args[2], img); err != nil {
		return err
	}
	fmt.Printf("✓ 完整世界地圖 %d×%d tile → %s(%d×%d px)\n",
		u5data.WorldSide, u5data.WorldSide, args[2], img.Bounds().Dx(), img.Bounds().Dy())
	fmt.Println("  驗收方式:與已知的 Britannia 世界地圖比對輪廓")
	return nil
}

// cmdText 印明文訊息檔的前幾筆(驗收解碼是否正確)。
func cmdText(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("用法:u5dump text <檔案> [--n N] [--dict gamedata]")
	}
	n := 5
	dictDir := filepath.Dir(args[0])
	for i := 1; i+1 < len(args); i++ {
		switch args[i] {
		case "--n":
			n, _ = strconv.Atoi(args[i+1])
		case "--dict":
			dictDir = args[i+1]
		}
	}
	tf, err := u5data.LoadText(args[0])
	if err != nil {
		return err
	}
	dict, derr := u5data.LoadDictionary(dictDir)
	if derr != nil {
		fmt.Fprintf(os.Stderr, "⚠ 讀不到詞典(%v)—— token 會留成 <XX>\n", derr)
	}
	fmt.Printf("%s:%d 筆記錄,斷字提示 %d 個\n", args[0], len(tf.Records), tf.HyphenHintCount())
	for i, r := range tf.Records {
		if i >= n {
			break
		}
		mark := " "
		if r.Page {
			mark = "{"
		}
		text := r.Expand(dict)
		if len(text) > 200 {
			text = text[:200] + "…"
		}
		fmt.Printf("\n[%3d]%s %s\n", r.Index, mark, text)
	}
	return nil
}

// cmdScene 用純 CPU 畫出遊戲畫面。
//
// 這是本專案的 headless 驗收路徑:與實機顯示共用同一個 render.Scene,
// 所以截圖就是實機畫面。不需要 X11、不需要 GL —— 之前綁 ebiten 的版本在
// xvfb + 軟體 GL 下死鎖了五小時,那正是改成 CPU 繪製的原因。
func cmdScene(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("用法:u5dump scene <gamedata> <U5_E 目錄> <out.png> [--font 前綴] [--at X Y] [--script nsewEKTyN…] [--scene 地點名 樓層 X Y] [--hour H]")
	}
	fontPrefix := "assets/fonts/eten-15"
	atX, atY := -1, -1
	script := ""
	sceneName, sceneFloor := "", 0
	hour := -1
	for i := 3; i < len(args); i++ {
		switch args[i] {
		case "--hour":
			if i+1 < len(args) {
				hour, _ = strconv.Atoi(args[i+1])
			}
		case "--scene":
			if i+4 < len(args) {
				sceneName = args[i+1]
				sceneFloor, _ = strconv.Atoi(args[i+2])
				atX, _ = strconv.Atoi(args[i+3])
				atY, _ = strconv.Atoi(args[i+4])
			}
		case "--script":
			if i+1 < len(args) {
				script = args[i+1]
			}
		case "--font":
			if i+1 < len(args) {
				fontPrefix = args[i+1]
			}
		case "--at":
			if i+2 < len(args) {
				atX, _ = strconv.Atoi(args[i+1])
				atY, _ = strconv.Atoi(args[i+2])
			}
		}
	}

	bundle, warns := assets.Load(assets.Options{
		GameData: args[0], FMTowns: args[1], FontPrefix: fontPrefix,
	})
	for _, w := range warns {
		fmt.Fprintf(os.Stderr, "⚠ %s\n", w)
	}

	st := &game.State{
		World: bundle.World, Under: bundle.Under, Scenes: bundle.Scenes,
		NPCs: bundle.NPCs, Talks: bundle.Talks, Clock: game.NewClock(), MaxMessages: 8,
	}
	if bundle.Save != nil {
		st.LoadFrom(bundle.Save)
	}
	if hour >= 0 {
		st.Clock.Hour = hour
	}
	// 存檔先套(時間、隊伍、位置),命令列參數再覆蓋。順序反了的話
	// --at / --scene 會被存檔的位置蓋掉,截圖就永遠停在同一個地方。
	if atX >= 0 && atY >= 0 {
		st.X, st.Y = atX, atY
	} else if bundle.Save == nil {
		st.X, st.Y = assets.FindLandStart(bundle.World, 1)
	}
	if sceneName != "" {
		loc, ok := findLocation(sceneName)
		if !ok {
			return fmt.Errorf("地點表裡沒有 %q", sceneName)
		}
		if err := st.SetScene(loc.Number(), sceneFloor, atX, atY); err != nil {
			return err
		}
	}
	st.Log("汝已抵達不列顛尼亞。此地由不列顛王治理,然其人已然失蹤。")
	if err := playScript(st, script); err != nil {
		return err
	}

	sc := &render.Scene{
		State: st,
		Tiles: bundle.Tiles,
		Text:  render.NewTextRenderer(bundle.Charset, bundle.CJK, render.ColorText),
	}

	if err := writePNG(args[2], sc.Render()); err != nil {
		return err
	}
	where := fmt.Sprintf("大地圖 %d,%d", st.X, st.Y)
	if st.InScene() {
		where = fmt.Sprintf("%s 第 %d 層 %d,%d", st.LocationName(), st.Floor+1, st.X, st.Y)
	}
	fmt.Printf("✓ 畫面 %d×%d → %s(%s)\n", render.CanvasWidth, render.CanvasHeight, args[2], where)
	for _, m := range st.Messages {
		fmt.Printf("  訊息:%s\n", m)
	}
	if bundle.CJK != nil {
		var miss []rune
		for _, m := range st.Messages {
			miss = append(miss, bundle.CJK.MissingRunes(m)...)
		}
		if len(miss) > 0 {
			fmt.Printf("⚠ 訊息裡有 %d 個缺字:%s\n", len(miss), string(miss))
		}
	}
	_ = u5data.TileSize
	return nil
}

// cmdSave 印出一份存檔的內容(名冊、隊伍、時間、位置)。
func cmdSave(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("用法:u5dump save <SAVED.GAM 或 INIT.GAM>")
	}
	sv, err := u5data.LoadSave(args[0])
	if err != nil {
		return err
	}
	where := fmt.Sprintf("大地圖 (%d,%d)", sv.X, sv.Y)
	if sv.Location > 0 {
		if loc, e := u5data.LocationByNumber(sv.Location); e == nil {
			where = fmt.Sprintf("%s 第 %+d 層 (%d,%d)", loc.DisplayName(), sv.Floor, sv.X, sv.Y)
		}
	} else if sv.Floor < 0 {
		where = fmt.Sprintf("地下世界 (%d,%d)", sv.X, sv.Y)
	}
	fmt.Printf("%s\n", args[0])
	fmt.Printf("  時間:%d 年 %d 月 %d 日 %02d:%02d\n", sv.Year, sv.Month, sv.Day, sv.Hour, sv.Minute)
	fmt.Printf("  位置:%s   載具 tile %d   業報 %d   隊伍 %d 人\n",
		where, sv.Transport, sv.Karma, sv.PartySize)
	fmt.Printf("  名冊:\n")
	for i := range sv.Roster {
		c := &sv.Roster[i]
		if !c.Present() {
			continue
		}
		mark := "  "
		if i < sv.PartySize {
			mark = "★ "
		}
		fmt.Printf("   %s%2d %-10s %-8s Lv%-2d  HP %3d/%-3d  MP %2d  力%2d 敏%2d 智%2d  經驗 %d\n",
			mark, i, c.Name, c.ClassName(), c.Level, c.HP, c.MaxHP, c.MP,
			c.Strength, c.Dex, c.Intel, c.Exp)
	}
	return nil
}

// cmdConv 把一筆對話完整解析出來:固定欄位 + 關鍵字表 + 每個關鍵字的回應與副作用。
func cmdConv(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("用法:u5dump conv <gamedata> <地點名> [對話號碼]")
	}
	set, err := u5data.LoadTalkSet(args[0])
	if err != nil {
		return err
	}
	loc, ok := findLocation(args[1])
	if !ok {
		return fmt.Errorf("地點表裡沒有 %q", args[1])
	}
	npcs, err := u5data.LoadNPCSet(args[0])
	if err != nil {
		return err
	}
	slots, err := npcs.At(loc.Number())
	if err != nil {
		return err
	}
	want := -1
	if len(args) > 2 {
		want, _ = strconv.Atoi(args[2])
	}
	for i := range slots {
		n := &slots[i]
		if !n.Present() || n.Dialogue == 0 || n.Dialogue >= u5data.DialogueShopFirst {
			continue
		}
		if want >= 0 && int(n.Dialogue) != want {
			continue
		}
		rec, ok := set.Record(loc.Number(), int(n.Dialogue))
		if !ok {
			fmt.Printf("#%d 對話 %d:.TLK 裡找不到\n", i, n.Dialogue)
			continue
		}
		c := u5data.ParseConversation(rec, set.Dict)
		fmt.Printf("\n=== #%d 對話 %d:%s ===\n", i, n.Dialogue, c.Name)
		fmt.Printf("  外貌:%s\n  招呼:%s\n  職業:%s\n  道別:%s\n",
			c.Description, c.Greeting, c.Job, c.Bye)
		fmt.Printf("  關鍵字(%d):%v\n", len(c.Keywords()), c.Keywords())
		for _, kw := range c.Keywords() {
			t, fx, _ := c.Respond(kw)
			var tags []string
			if fx.JoinParty {
				tags = append(tags, "加入隊伍")
			}
			if fx.CallGuards {
				tags = append(tags, "叫衛兵")
			}
			if fx.KarmaDelta != 0 {
				tags = append(tags, fmt.Sprintf("業報%+d", fx.KarmaDelta))
			}
			if fx.AsksPlayer {
				tags = append(tags, "反問玩家")
			}
			if fx.EndTalk {
				tags = append(tags, "結束對話")
			}
			suffix := ""
			if len(tags) > 0 {
				suffix = "  [" + strings.Join(tags, " ") + "]"
			}
			fmt.Printf("    %-5s → %s%s\n", kw, strings.ReplaceAll(t, "\n", " "), suffix)
		}
	}
	return nil
}

// cmdNPC 列出一個地點在某個時刻的居民 —— 位置、tile、對話號碼、一日作息。
func cmdNPC(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("用法:u5dump npc <gamedata> <地點名> [--hour H]")
	}
	hour := -1
	for i := 2; i < len(args); i++ {
		if args[i] == "--hour" && i+1 < len(args) {
			hour, _ = strconv.Atoi(args[i+1])
		}
	}
	set, err := u5data.LoadNPCSet(args[0])
	if err != nil {
		return err
	}
	loc, ok := findLocation(args[1])
	if !ok {
		return fmt.Errorf("地點表裡沒有 %q", args[1])
	}
	npcs, err := set.At(loc.Number())
	if err != nil {
		return err
	}
	fmt.Printf("%s(%s)—— 32 個槽\n", loc.Name, loc.DisplayName())
	for i := range npcs {
		n := &npcs[i]
		if !n.Present() {
			continue
		}
		kind := fmt.Sprintf("對話 %d", n.Dialogue)
		switch {
		case i == u5data.PartySlot:
			kind = "隊伍"
		case n.Dialogue == u5data.DialogueNone:
			kind = "不搭話"
		case n.IsShopkeeper():
			kind = fmt.Sprintf("商人 %02X", n.Dialogue)
		case n.Dialogue >= u5data.DialogueFrightened:
			kind = fmt.Sprintf("特殊 %02X", n.Dialogue)
		}
		fmt.Printf("  #%2d tile %3d  %-10s", i, n.TileIndex(), kind)
		if hour >= 0 {
			x, y, f := n.At(hour)
			fmt.Printf("  %02d:00 在 (%2d,%2d) 第 %d 層", hour, x, y, f+1)
		} else {
			// 一日作息:只印換位置的時刻
			prev := -1
			for h := 0; h < 24; h++ {
				sl := n.Schedule.Slot(h)
				if sl == prev {
					continue
				}
				prev = sl
				fmt.Printf("  %02d:00→(%2d,%2d)F%d", h,
					n.Schedule.X[sl], n.Schedule.Y[sl], n.Schedule.Floor[sl]+1)
			}
		}
		fmt.Println()
	}
	return nil
}

// findLocation 依英文名找地點(不分大小寫)。
func findLocation(name string) (*u5data.Location, bool) {
	for i := range u5data.Locations {
		if strings.EqualFold(u5data.Locations[i].Name, name) {
			return &u5data.Locations[i], true
		}
	}
	return nil, false
}

// playScript 把一串按鍵餵給狀態機,讓「走進城裡再走出來」這種劇本能 headless 重現。
//
//	n s e w 移動   E 進入   K 攀爬   T 交談   y / N 回答提問   "keyword" 對話輸入
func playScript(st *game.State, script string) error {
	rs := []rune(script)
	for i := 0; i < len(rs); i++ {
		r := rs[i]
		// 對話中要打字:用 `"keyword"` 包起來,送出後自動按 Enter。
		if r == '"' {
			word := ""
			for i++; i < len(rs) && rs[i] != '"'; i++ {
				word += string(rs[i])
			}
			for _, c := range word {
				st.TypeRune(c)
			}
			st.Submit()
			continue
		}
		switch r {
		case 'n':
			st.Move(game.North)
		case 's':
			st.Move(game.South)
		case 'e':
			st.Move(game.East)
		case 'w':
			st.Move(game.West)
		case 'E':
			st.Enter()
		case 'K':
			st.Klimb()
		case 'T':
			st.Talk()
		case 'y':
			st.Answer(true)
		case 'N':
			st.Answer(false)
		case ' ':
		default:
			return fmt.Errorf("腳本裡看不懂的動作 %q(可用:n s e w 移動、E 進入、K 攀爬、T 交談、y/N 回答)", r)
		}
	}
	return nil
}

// cmdSceneMaps 把一個場景檔(TOWNE/CASTLE/KEEP/DWELLING.DAT)的 16 張地圖畫出來。
func cmdSceneMaps(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("用法:u5dump scenemaps <場景檔.DAT> <U5_E 目錄> <out.png>")
	}
	tiles, err := u5data.LoadFMTownsTileSet(fmTownsTilePaths(args[1]))
	if err != nil {
		return err
	}
	scenes, err := u5data.LoadSceneMaps(args[0])
	if err != nil {
		return err
	}
	img, err := u5data.RenderSceneMaps(scenes, tiles, 4)
	if err != nil {
		return err
	}
	if err := writePNG(args[2], img); err != nil {
		return err
	}
	fmt.Printf("✓ %d 張 %d×%d 場景地圖 → %s(%d×%d px)\n",
		len(scenes), u5data.SceneSide, u5data.SceneSide, args[2],
		img.Bounds().Dx(), img.Bounds().Dy())
	fmt.Println("  驗收方式:切對了會看到建築、道路、城牆;切錯了是雜訊")
	return nil
}

// cmdTown 用原版地點表把某個地點的所有樓層畫出來。
//
// 對應規則完全照 sub_5C8:檔案 = SceneFiles[(編號-1)/8]、索引 = 起始索引 + 樓層。
func cmdTown(args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("用法:u5dump town <gamedata> <U5_E 目錄> <地名> <out.png>")
	}
	want := strings.ToUpper(args[2])
	var loc *u5data.Location
	for i := range u5data.Locations {
		if strings.ToUpper(u5data.Locations[i].Name) == want {
			loc = &u5data.Locations[i]
			break
		}
	}
	if loc == nil {
		return fmt.Errorf("地點表裡沒有 %q(試試 BRITAIN / MOONGLOW / FOGSBANE / IOLO'S HUT)", args[2])
	}
	tiles, err := u5data.LoadFMTownsTileSet(fmTownsTilePaths(args[1]))
	if err != nil {
		return err
	}
	scenes, err := u5data.LoadSceneMaps(filepath.Join(args[0], u5data.SceneFiles[loc.SceneFile]))
	if err != nil {
		return err
	}
	var floors []u5data.SceneMap
	for f := loc.FloorMin; f <= loc.FloorMax; f++ {
		idx := loc.SceneIndex + f
		if idx >= len(scenes) {
			break
		}
		floors = append(floors, scenes[idx])
	}
	img, err := u5data.RenderSceneMaps(floors, tiles, len(floors))
	if err != nil {
		return err
	}
	if err := writePNG(args[3], img); err != nil {
		return err
	}
	fmt.Printf("✓ %s(%s)世界座標 (%d,%d) → %s 索引 %d,共 %d 層 → %s\n",
		loc.Name, loc.DisplayName(), loc.X, loc.Y,
		u5data.SceneFiles[loc.SceneFile], loc.SceneIndex, loc.FloorMax-loc.FloorMin+1, args[3])
	return nil
}
