package u5data

import (
	"os"
	"strings"
	"testing"
)

// TestLookTableCrossChecksKnownTiles:敘述表的索引不是「看起來對」就算數。
//
// 拿**別處獨立推出來的 tile 常數**當對照 —— 那些常數的來源是原版的
// 移動判定、買馬、上下載具,與 LOOK2.DAT 完全無關。兩邊對得上,
// 索引空間才算確立(單點碰運氣的機率很低,十一點同時中就不是運氣)。
func TestLookTableCrossChecksKnownTiles(t *testing.T) {
	lt := loadLookOrSkip(t)

	terrain := map[int]string{
		5:                     "grass",   // mountableTiles 說 5 是草地
		68:                    "cobble",  // mountableTiles 說 68/69 是城鎮地面
		69:                    "cobble",  //
		TileCodex:             "the Shrine of the Codex",
		TileShrine:            "a mystic shrine",
		TileShrineDesecrated:  "a ruined shrine",
		TileDoorA:             "a wooden door",
		TileLockedDoor:        "a locked door",
		int(TileStairsDown):   "a ladder",
		int(TileStairsUp):     "a ladder",
	}
	for tile, want := range terrain {
		if got := lt.Terrain(tile); got != want {
			t.Errorf("地形 %d(0x%02X)= %q,預期 %q", tile, tile, got, want)
		}
	}

	// 物件走的是另一個索引空間(+256)。這幾條是最容易寫錯的地方:
	// 直接查 Terrain(0x10) 得到的是「a small hut」,那是地形的第 16 格。
	object := map[int]string{
		TileHorse:     "a horse",
		TileCarpetObj: "an odd rug",
		VehicleSkiff:  "a skiff",
	}
	for tile, want := range object {
		if got := lt.Object(tile); got != want {
			t.Errorf("物件 0x%02X = %q,預期 %q", tile, got, want)
		}
		// 同一個號碼查地形一定不一樣 —— 一樣就代表兩個空間被寫成同一個。
		if lt.Terrain(tile) == want {
			t.Errorf("物件 0x%02X 在地形空間也查到 %q —— 索引沒有分開", tile, want)
		}
	}
}

// TestLookStemsKeepTheirTrailingSpace:會接後綴的那幾筆,結尾的空格是接縫。
//
// 原版印完敘述之後直接接上火名 / 地牢名 / 時刻,沒有另外補空白。
// 誰要是「順手」把尾巴的空白 trim 掉,畫面上就會變成
// 「the Flame ofTruth」—— 而測試若只比對前綴就抓不到。
func TestLookStemsKeepTheirTrailingSpace(t *testing.T) {
	lt := loadLookOrSkip(t)
	for _, tile := range []int{0xDE, 0xDF, 0xFA, 0xFB} {
		got := lt.Terrain(tile)
		if !strings.HasSuffix(got, " ") {
			t.Errorf("0x%02X = %q,結尾少了接後綴用的空白", tile, got)
		}
	}
}

// TestLookPlaceholdersMarkSpecialCasedTiles:被程式特判掉的格子在表上是佔位符。
//
// 這條反過來驗證了「哪些 tile 有特別處理」的清單:水晶球、天空、噴泉、
// 招牌 —— 原版都不從表上取字,所以表上留的是 `*`。若哪天有一格從佔位符
// 變成真的敘述,代表清單抄漏了一種。
func TestLookPlaceholdersMarkSpecialCasedTiles(t *testing.T) {
	lt := loadLookOrSkip(t)
	for _, tile := range []int{0x59, 0xD8, 0xD9, 0xDA, 0xDB, 0x89, 0x8A, 0xA0, 0xA4, 0xF8} {
		if got := lt.Terrain(tile); !LookPlaceholder(got) {
			t.Errorf("0x%02X = %q,預期是佔位符 —— 這一格本該由程式特判", tile, got)
		}
	}
}

func loadLookOrSkip(t *testing.T) *LookTable {
	t.Helper()
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	lt, err := LoadLook(dir)
	if err != nil {
		t.Fatalf("讀 LOOK2.DAT:%v", err)
	}
	return lt
}
