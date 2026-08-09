package game

import (
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// roomScene 進第 n 座地牢,並把腳下那一格設成第 room 號房間。
func roomScene(t *testing.T, n, room int) *State {
	t.Helper()
	s := dungeonState(t)
	s.Location, s.Floor = 0, 0
	e := u5data.DungeonEntrances[n]
	s.X, s.Y = e.X, e.Y
	if !s.EnterDungeon(n, false) {
		t.Skip("進不了地牢")
	}
	d := s.Dungeon
	s.Dungeons.Set(d.Index, d.Level, d.X, d.Y, u5data.DungeonRoomF|byte(room))
	return s
}

// TestClearedRoomIsWipedOffTheMap —— ★★ 清過的房間再進地牢就不見了。
//
// 原版 `sub_FA7C` 在進地牢時把 `0xFn` 改成 `0xAn`(可走的空房間)——
// 引擎此前每次踏上房間格都會再打一場。
func TestClearedRoomIsWipedOffTheMap(t *testing.T) {
	const dungeon, room = 0, 3
	s := roomScene(t, dungeon, room)
	d := s.Dungeon
	x, y, level := d.X, d.Y, d.Level
	tile := u5data.DungeonRoomF | byte(room)

	// 還沒清過 → 進地牢不會動它。
	s.applyClearedRooms()
	if got := s.Dungeons.At(dungeon, level, x, y); got != tile {
		t.Fatalf("沒清過卻被改成 0x%02X", got)
	}

	s.markRoomCleared(s.locationCode(), tile)
	if !s.roomIsCleared(s.locationCode(), tile) {
		t.Fatal("記了卻查不到")
	}
	s.applyClearedRooms()
	got := s.Dungeons.At(dungeon, level, x, y)
	if want := tile & u5data.DungeonRoomClearedMask; got != want {
		t.Errorf("清過之後是 0x%02X,預期 0x%02X(0xFn → 0xAn)", got, want)
	}
	// ★ 變成的是「空房間」那一族,不是通道也不是牆。
	if u5data.DungeonKind(got) != u5data.DungeonRoomA {
		t.Errorf("清過之後的種類是 0x%02X,預期 0x%02X", u5data.DungeonKind(got), u5data.DungeonRoomA)
	}
	// 房號要保留 —— 遮罩只碰高四位元。
	if u5data.DungeonRoomNumber(got) != room {
		t.Errorf("房號被改掉了:%d → %d", room, u5data.DungeonRoomNumber(got))
	}
}

// TestSixRoomsAreNeverMarkedCleared —— ★★ 六間例外房永遠有怪。
//
// 資料是原版 `byte_55110` 的六筆,鍵 = `房號 | ((地點碼 & 0x0F) << 4)`。
// 地點碼低四位元 4 / 5 = 謬誤 / 貪婪。
func TestSixRoomsAreNeverMarkedCleared(t *testing.T) {
	for _, tc := range []struct {
		location, room int
		armed          bool
		name           string
	}{
		{0x24, 1, true, "謬誤 房 1"},
		{0x24, 6, true, "謬誤 房 6"},
		{0x24, 11, true, "謬誤 房 11"},
		{0x24, 12, true, "謬誤 房 12"},
		{0x25, 0, true, "貪婪 房 0"},
		{0x25, 11, true, "貪婪 房 11"},
		// 反對照:同一座地牢的別的房間會被記。
		{0x24, 2, false, "謬誤 房 2 —— 不在清單上"},
		{0x25, 1, false, "貪婪 房 1 —— 不在清單上"},
		// 反對照:別的地牢的同一個房號不受影響。
		{0x21, 1, false, "欺瞞 房 1"},
		{0x28, 11, false, "末日 房 11"},
	} {
		if got := u5data.DungeonRoomAlwaysArmed(tc.location, tc.room); got != tc.armed {
			t.Errorf("%s:永遠有怪 = %v,預期 %v", tc.name, got, tc.armed)
		}
	}
}

// TestAlwaysArmedRoomIsNotRemembered —— 例外房打完也不會被記。
func TestAlwaysArmedRoomIsNotRemembered(t *testing.T) {
	// 地點碼 0x24 = 索引 3 = 謬誤;房 1 在清單上。
	const dungeon, room = 3, 1
	s := roomScene(t, dungeon, room)
	tile := u5data.DungeonRoomF | byte(room)
	if s.locationCode() != 0x24 {
		t.Skipf("地點碼是 0x%02X,預期 0x24", s.locationCode())
	}
	s.markRoomCleared(s.locationCode(), tile)
	if s.roomIsCleared(s.locationCode(), tile) {
		t.Error("例外清單上的房間被記成清過了")
	}
	// 反對照:同一座地牢不在清單上的房間**會**被記。
	other := byte(u5data.DungeonRoomF | 2)
	s.markRoomCleared(s.locationCode(), other)
	if !s.roomIsCleared(s.locationCode(), other) {
		t.Error("不在清單上的房間卻沒被記 —— 那是整支關掉,不是例外生效")
	}
}

// TestRoomMemoryIsPerDungeon —— 不同地牢的同一個房號互不干擾。
//
// ⚠ **除了 0x21 與 0x22** —— 它們在位元陣列裡共用同一批位元
// (`DungeonRoomBlock` 的「≥1 就 −1」修正)。那是原版的索引方式,
// 而 `DUNGEON.CBT` 的房間查表用的是同一個函式 ⇒ 不是我算錯。
func TestRoomMemoryIsPerDungeon(t *testing.T) {
	s := roomScene(t, 0, 3)
	tile := byte(u5data.DungeonRoomF | 3)
	s.markRoomCleared(0x23, tile) // 第三座
	if s.roomIsCleared(0x25, tile) {
		t.Error("記了 0x23 卻讓 0x25 也算清過")
	}
	if !s.roomIsCleared(0x23, tile) {
		t.Error("記了 0x23 自己卻查不到")
	}
	// ★ 0x21 與 0x22 共用 —— 釘住這個原版的怪處,免得日後被「修正」掉。
	s.markRoomCleared(0x21, tile)
	if !s.roomIsCleared(0x22, tile) {
		t.Error("0x21 與 0x22 該共用同一批位元(DungeonRoomBlock 的修正)")
	}
}

// deepWaterTile 是深水(`look#1` = deep water)。
//
// ⚠ 這裡就地定義而不是加進 `u5data` —— 只有這一條測試需要它,
// 而 `u5data` 裡再多一個 tile 常數會讓「哪些 tile 有名字」更難數。
const deepWaterTile = 0x01

// TestDungeonsStartCollapsed —— ★★★ 開局八座地牢入口都是崩塌的。
//
// 這是 U5 的主線閘門:要各自喊對力量之言才進得去。引擎此前直接用
// `BRIT.DAT` 的原始地形 ⇒ 一開局所有地牢都能走進去,包括末日。
//
// 極性的證據見 `u5data.DungeonIsSealed` 的檔頭(四條獨立來源)。
func TestDungeonsStartCollapsed(t *testing.T) {
	s := worldState(t)
	if s.World == nil {
		t.Skip("沒有世界地圖")
	}
	// 開局:八個旗標全 0(`INIT.GAM` 實測)。
	s.DungeonSeal = [u5data.VirtueCount]byte{}
	s.applyWorldFlags()
	collapsed := 0
	for i := range u5data.DungeonEntrances {
		e := u5data.DungeonEntrances[i]
		got := s.TileAt(e.X, e.Y)
		if got == u5data.TileDungeonSealed {
			collapsed++
			continue
		}
		// ★ **末日(DOOM)的座標 (128,128) 在地表是深水** —— 它是從幽冥界
		// 進去的,地表上沒有入口。`applyWorldFlags` 的「目前是原始入口地形
		// 才改」那道守衛正是為了這種格子,所以它不會被誤改成崩塌。
		if e.Name == "DOOM" && got == deepWaterTile {
			continue
		}
		t.Errorf("%s 的入口是 0x%02X,開局應該是崩塌的 0x%02X",
			e.DisplayName(), got, u5data.TileDungeonSealed)
	}
	// 反對照:不能一個都沒改(那樣上面的迴圈會全部走 continue 而不報錯)。
	if collapsed != len(u5data.DungeonEntrances)-1 {
		t.Errorf("只有 %d 座崩塌,預期 %d 座(末日在幽冥界)",
			collapsed, len(u5data.DungeonEntrances)-1)
	}
}

// TestOpenedDungeonStaysOpen —— 反對照:旗標設著(= 已開封)就不改寫。
//
// 少了這一條,「開局全崩塌」與「永遠全崩塌」用同一個觀察分不開 ——
// 而後者會讓遊戲整個過不去。
func TestOpenedDungeonStaysOpen(t *testing.T) {
	s := worldState(t)
	if s.World == nil {
		t.Skip("沒有世界地圖")
	}
	for i := range s.DungeonSeal {
		s.DungeonSeal[i] = u5data.DungeonSealedBit // ★ 設著 = 通的
	}
	// 先把地形擺成原始入口,再看 applyWorldFlags 會不會動它。
	for i := range u5data.DungeonEntrances {
		e := u5data.DungeonEntrances[i]
		s.SetTileAt(e.X, e.Y, u5data.DungeonEntranceTile[i])
	}
	s.applyWorldFlags()
	for i := range u5data.DungeonEntrances {
		e := u5data.DungeonEntrances[i]
		if got := s.TileAt(e.X, e.Y); got != u5data.DungeonEntranceTile[i] {
			t.Errorf("%s 已開封卻被改成 0x%02X", e.DisplayName(), got)
		}
	}
}

// TestDungeonIsSealedHasTheOppositePolarity —— ★ 釘住那個反直覺的極性。
//
// `flag & 0x80 != 0` 這個直覺寫法方向剛好相反。這條測試存在的理由就是
// 讓「順手改回直覺寫法」的人立刻紅燈。
func TestDungeonIsSealedHasTheOppositePolarity(t *testing.T) {
	if !u5data.DungeonIsSealed(0) {
		t.Error("旗標 0 應該是崩塌(`sub_1056C` 的 `setz`)")
	}
	if u5data.DungeonIsSealed(u5data.DungeonSealedBit) {
		t.Error("旗標 0x80 應該是通的 —— 喊力量之言把 0 變成 0x80 是開封")
	}
}

// TestDesecratedShrineShowsAsRuined —— 被玷污的聖壇在地圖上是毀壞的樣子。
//
// ⚠ 聖壇那一邊的極性是**正常的**(`& 0x80` 設著 = 毀壞),與地牢相反。
// 兩支原版函式(`sub_1056C` / `sub_105AC`)一個 `setz` 一個 `setnbe`。
func TestDesecratedShrineShowsAsRuined(t *testing.T) {
	s := worldState(t)
	if s.World == nil {
		t.Skip("沒有世界地圖")
	}
	// 挑一座不在 (0,0) 的(靈性聖壇在幽冥界)。
	v := -1
	for i := range u5data.Shrines {
		if u5data.Shrines[i].X != 0 || u5data.Shrines[i].Y != 0 {
			v = i
			break
		}
	}
	if v < 0 {
		t.Skip("沒有地表聖壇")
	}
	sh := u5data.Shrines[v]
	s.SetTileAt(sh.X, sh.Y, u5data.TileShrine)
	s.ShrineFlag[v] = 0 // 沒被玷污 → 不動
	s.applyWorldFlags()
	if got := s.TileAt(sh.X, sh.Y); got != u5data.TileShrine {
		t.Fatalf("%s 沒被玷污卻改成 0x%02X", sh.NameZH, got)
	}
	s.ShrineFlag[v] = u5data.ShrineDesecratedBit
	s.applyWorldFlags()
	if got := s.TileAt(sh.X, sh.Y); got != u5data.TileShrineDesecrated {
		t.Errorf("%s 被玷污卻是 0x%02X,預期 0x%02X",
			sh.NameZH, got, u5data.TileShrineDesecrated)
	}
}
