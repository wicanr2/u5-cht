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
		return fmt.Errorf("用法:u5dump tlk <檔案> [--sjis] [--n N]")
	}
	enc := u5data.TalkEncodingHighBit
	n := 3
	for i := 1; i < len(args); i++ {
		switch args[i] {
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
	fmt.Printf("%s:%d 筆,編碼 %s\n", args[0], len(tf.Records), tf.Encoding)
	for i, r := range tf.Records {
		if i >= n {
			break
		}
		fmt.Printf("\n-- 第 %d 筆(NPC %d,offset 0x%X,%d B)--\n", i, r.NPCIndex, r.Offset, len(r.Data))
		for j, s := range r.Strings() {
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
