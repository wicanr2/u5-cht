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
