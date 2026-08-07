package u5data

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDungeonBandsTileTheHalfWidth:切片寬度必須把半個畫面剛好鋪滿。
//
// 這是整套透視幾何的**唯一證據**,所以拿真的檔案來驗,不寫死數字:
//
//	側牆四階的寬度相加 == 80
//	正面第 d 階的寬度  == 80 − (前 d 階側牆寬度之和)
//
// 五條算式同時成立,而且兩組數字來自檔案裡兩批不同的形狀。
// 只要 `dungeonBandX` 或任一張圖的尺寸讀錯,這裡就會爆。
func TestDungeonBandsTileTheHalfWidth(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA")
	}
	set, err := LoadPictures(filepath.Join(dir, "DNG1.16"))
	if err != nil {
		t.Fatal(err)
	}
	sum := 0
	for d := 0; d < DungeonViewDepths; d++ {
		side := set[dungeonSideSolid+d]
		if side == nil {
			t.Fatalf("側牆第 %d 階是空的", d)
		}
		if got := DungeonBandX(d); got != sum {
			t.Errorf("第 %d 階的起點是 %d,但前面幾階加起來是 %d", d, got, sum)
		}
		front := set[dungeonFrontDoor+d]
		if front == nil {
			t.Fatalf("正面第 %d 階是空的", d)
		}
		if want := DungeonViewHalfWidth - sum; front.Width != want {
			t.Errorf("正面第 %d 階寬 %d,預期 %d(= 80 − 前面側牆之和)",
				d, front.Width, want)
		}
		sum += side.Width
	}
	if sum != DungeonViewHalfWidth {
		t.Errorf("側牆四階加起來 %d,預期正好 %d", sum, DungeonViewHalfWidth)
	}
}

// TestDungeonThemeGrouping:八座地牢的外觀分組。
//
// `sub_5378` 是寫死的分組,不是「第 n 座用第 n 套」——
// 引擎第一版拿索引對 3 取餘數,結果欺瞞從磚牆變成洞穴。
func TestDungeonThemeGrouping(t *testing.T) {
	want := map[string]int{
		"DECEIT": 3, "DESPISE": 1, "DESTARD": 1, "WRONG": 3,
		"COVETOUS": 3, "SHAME": 2, "HYTHLOTH": 2, "DOOM": 1,
	}
	for i, e := range DungeonEntrances {
		loc := DungeonLocationBase + i
		if got := DungeonTheme(loc); got != want[e.Name] {
			t.Errorf("%s(地點 %#x)的外觀是 %d,預期 %d", e.Name, loc, got, want[e.Name])
		}
	}
	// 三套都要被用到 —— 只用一套代表分組寫錯了。
	used := map[int]bool{}
	for i := range DungeonEntrances {
		used[DungeonTheme(DungeonLocationBase+i)] = true
	}
	if len(used) != DungeonThemes {
		t.Errorf("八座地牢只用到 %d 套外觀,應該三套都有", len(used))
	}
}

// TestDungeonShapeSelection:選圖規則逐條對照 `sub_3878` / `sub_36C0`。
func TestDungeonShapeSelection(t *testing.T) {
	side := []struct {
		tile byte
		want int
		what string
	}{
		{DungeonPassage, dungeonSideOpen, "通道 → 看得穿的開口"},
		{DungeonLadderUp, dungeonSideOpen, "梯子也是通道類"},
		{DungeonRoomA, dungeonSideDoor, "房間 → 側牆有門"},
		{DungeonDoorway, dungeonSideDoor, "門口 → 側牆有門"},
		{DungeonRoomF, dungeonSideDoor, "房間 → 側牆有門"},
		{DungeonUnknownC, dungeonSideC0, "0xC0 自成一組"},
		{DungeonWall, dungeonSideSolid, "牆 → 實心側牆"},
		{DungeonUnknownD, dungeonSideSolid, "0xD0 也是實心"},
	}
	for _, c := range side {
		for d := 0; d < DungeonViewDepths; d++ {
			if got := DungeonSideShape(c.tile, d); got != c.want+d {
				t.Errorf("側牆 %02X 深度 %d → %d,預期 %d(%s)",
					c.tile, d, got, c.want+d, c.what)
			}
		}
	}
	// 正面:深度 0 一律 12,其餘照種類分組。
	for _, tile := range []byte{DungeonWall, DungeonRoomA, DungeonUnknownC} {
		if got := DungeonFrontShape(tile, 0); got != dungeonFrontDoor {
			t.Errorf("正面 %02X 深度 0 → %d,原版寫死 %d", tile, got, dungeonFrontDoor)
		}
	}
	front := map[byte]int{
		DungeonWall: dungeonFrontWall, DungeonUnknownD: dungeonFrontWall,
		DungeonRoomA: dungeonFrontDoor, DungeonDoorway: dungeonFrontDoor,
		DungeonRoomF: dungeonFrontDoor, DungeonUnknownC: dungeonFrontC0,
	}
	for tile, base := range front {
		for d := 1; d < DungeonViewDepths; d++ {
			if got := DungeonFrontShape(tile, d); got != base+d {
				t.Errorf("正面 %02X 深度 %d → %d,預期 %d", tile, d, got, base+d)
			}
		}
	}
	// 通道不擋視線。
	if DungeonFrontShape(DungeonPassage, 2) != -1 {
		t.Error("通道不該有正面")
	}
}

// TestEmptyShapesAreExactlyTheDepthZeroOnes:兩個空格不是漏做。
//
// 第 8 格與第 24 格是空的,因為深度 0 的正面**一律**用第 12 格 ——
// 那兩個基底的第 0 階永遠不會被要求。這條把「空格」與「選圖規則」綁在一起:
// 哪天有人把深度 0 的特例拿掉,這裡會立刻抓到少了圖。
func TestEmptyShapesAreExactlyTheDepthZeroOnes(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA")
	}
	set, err := LoadPictures(filepath.Join(dir, "DNG1.16"))
	if err != nil {
		t.Fatal(err)
	}
	empty := map[int]bool{}
	for i, p := range set {
		if p == nil {
			empty[i] = true
		}
	}
	if len(empty) != 2 || !empty[dungeonFrontWall] || !empty[dungeonFrontC0] {
		t.Fatalf("空的格子是 %v,預期正好 {%d, %d}",
			empty, dungeonFrontWall, dungeonFrontC0)
	}
	// 反過來:選圖規則不會去要那兩格。
	for _, tile := range []byte{
		DungeonPassage, DungeonWall, DungeonRoomA, DungeonRoomF,
		DungeonDoorway, DungeonUnknownC, DungeonUnknownD,
	} {
		for d := 0; d < DungeonViewDepths; d++ {
			for _, n := range []int{DungeonSideShape(tile, d), DungeonFrontShape(tile, d)} {
				if n >= 0 && set[n] == nil {
					t.Errorf("tile %02X 深度 %d 要了空的第 %d 格", tile, d, n)
				}
			}
		}
	}
}
