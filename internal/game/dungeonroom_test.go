package game

import (
	"strings"
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

// TestCollapsedEntranceCannotBeEntered —— ★★★ 崩塌的入口按 E 只印「What?」。
//
// 那道門在**分派表的定義域**上,不在進入函式裡:原版 `sub_2D72C` 是
// `switch (tile − 0x10)`,而 `0xDF − 0x10 = 0xCF > 0x2E` → default → 「What?」。
func TestCollapsedEntranceCannotBeEntered(t *testing.T) {
	s := worldState(t)
	if s.World == nil {
		t.Skip("沒有世界地圖")
	}
	const n = 1 // 輕蔑 —— 隨便挑一座地表上的
	e := u5data.DungeonEntrances[n]
	s.X, s.Y = e.X, e.Y
	s.Transport = u5data.VehicleWalk
	s.DungeonSeal[n] = 0 // 崩塌
	s.applyWorldFlags()
	s.Messages = nil
	s.Enter()
	if s.Dungeon != nil {
		t.Fatal("崩塌的入口竟然進去了")
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgWhat) {
		t.Errorf("沒印「%s」:%q", MsgWhat, s.Messages)
	}
}

// TestWordOfPowerOpensTheWayThrough —— ★★★ **無 debug 的正常玩家路徑**。
//
// 這一條是防 soft-lock 的:開局全部崩塌 + 進入被 tile 擋住 ⇒ 只要力量之言
// 那條路有一處不對,整個遊戲就過不去。所以要一路走完:
//
//	崩塌 → 按 E 進不去 → 站在入口喊那個字 → 地形變回原樣 → 按 E 進得去
func TestWordOfPowerOpensTheWayThrough(t *testing.T) {
	// ⚠ 用 `dungeonState` 而不是 `worldState` —— 這條要真的走進地牢,
	// 所以 `s.Dungeons` 不能是 nil(`worldState` 沒載它,測試會**跳過**)。
	s := dungeonState(t)
	s.Location, s.Floor = 0, 0
	if s.World == nil || s.Dungeons == nil {
		t.Skip("沒有世界地圖或地牢資料")
	}
	const n = 1
	e := u5data.DungeonEntrances[n]
	s.Transport = u5data.VehicleWalk
	s.DungeonSeal[n] = 0
	s.applyWorldFlags()
	if got := s.TileAt(e.X, e.Y); got != u5data.TileDungeonSealed {
		t.Fatalf("開局地形是 0x%02X,預期崩塌", got)
	}

	// ★ 喊力量之言要站在入口**旁邊**,不是站在上面 —— 原版掃的是視窗緩衝裡
	// 玩家那一格的 −1 / +32 / +1 / −32(`docs/re/26` §3.1),四個鄰格。
	// 而**進入**要站在上面。兩件事的站位不同,照原版。
	s.X, s.Y = e.X+1, e.Y
	word := u5data.WordsOfPower[n]
	s.Messages = nil
	if !s.SpeakWord(word) {
		t.Fatalf("喊 %q 沒有效果:%q", word, s.Messages)
	}
	if got := s.TileAt(e.X, e.Y); got != u5data.DungeonEntranceTile[n] {
		t.Fatalf("喊完之後地形是 0x%02X,預期回到 0x%02X",
			got, u5data.DungeonEntranceTile[n])
	}
	if u5data.DungeonIsSealed(s.DungeonSeal[n]) {
		t.Fatal("喊完之後旗標還是「崩塌」—— 極性反了")
	}
	// ★ 訊息要說「開了」,不是「崩塌了」。
	if !strings.Contains(strings.Join(s.Messages, "|"), "入口開了") {
		t.Errorf("沒說入口開了:%q", s.Messages)
	}

	// 現在站上去按 E 該進得去。
	s.X, s.Y = e.X, e.Y
	s.Messages = nil
	s.Enter()
	if s.Dungeon == nil {
		t.Fatalf("喊開之後還是進不去:%q", s.Messages)
	}
	// 反對照:再喊一次會封回去,而封回去之後又進不去。
	s.LeaveDungeon()
	s.X, s.Y = e.X+1, e.Y
	s.SpeakWord(word)
	if got := s.TileAt(e.X, e.Y); got != u5data.TileDungeonSealed {
		t.Errorf("再喊一次地形是 0x%02X,預期又崩塌(XOR 是對稱的)", got)
	}
}

// TestTheThreeEntranceTilesCoverTheEightDungeons —— 兩張表不能漂移。
//
// `DungeonEntranceTiles`(分派表的三個 case)與 `DungeonEntranceTile`
// (八座各自的原始地形)出處不同,所以刻意不合併 —— 但後者的每一個值
// 都必須在前者裡,否則某座地牢會變成「地形對但按 E 不認」。
func TestTheThreeEntranceTilesCoverTheEightDungeons(t *testing.T) {
	for i, tile := range u5data.DungeonEntranceTile {
		if !u5data.IsDungeonEntranceTile(tile) {
			t.Errorf("%s 的入口地形 0x%02X 不在三個 case 裡",
				u5data.DungeonEntrances[i].Name, tile)
		}
	}
	// 反對照:崩塌的地形不算入口(那就是整個機制的重點)。
	if u5data.IsDungeonEntranceTile(u5data.TileDungeonSealed) {
		t.Error("崩塌的地形被當成入口了")
	}
}

// TestThreeFieldsSurviveASaveRoundTrip —— ★ 三個新接上的存檔欄位。
//
// 位移都是從 `sub_27D24` 的讀取序列推出來,再拿 `INIT.GAM` 驗過
// (0x02D5 = 0xFF 正是「沒指定」的哨兵)。這條測試釘住 round-trip。
func TestThreeFieldsSurviveASaveRoundTrip(t *testing.T) {
	s := roomScene(t, 0, 5)
	// 指定第 1 位、走 137 回合、清掉兩間房。
	if s.PartySize < 2 {
		t.Skip("隊伍太小")
	}
	s.Roster[1].Status = u5data.StatusGood
	if !s.SetActivePlayer('2') {
		t.Fatal("'2' 不被當成指定行動者的鍵")
	}
	s.turnsSinceAlms = 137
	s.markRoomCleared(0x23, byte(u5data.DungeonRoomF|5))
	s.markRoomCleared(0x26, byte(u5data.DungeonRoomF|9))
	before := s.roomsCleared

	// ⚠ `ExportSave` 需要**底稿** —— 引擎只解出部分欄位,把未解欄位清成 0
	// 會讓存檔在原版裡壞掉(見 `savegame.go` 的說明)。
	base := s.BaseSave
	if base == nil {
		t.Skip("沒有底稿存檔")
	}
	sv, err := s.ExportSave(base)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := sv.Encode()
	if err != nil {
		t.Fatal(err)
	}
	// ★ 直接驗**位元組**,不只驗 round-trip —— round-trip 用錯的位移
	// 也會通過(寫進去再從同一個地方讀出來)。
	if got := raw[u5data.SaveActiveMemberOffset]; got != 1 {
		t.Errorf("0x%04X 是 %d,預期 1", u5data.SaveActiveMemberOffset, got)
	}
	if got := raw[u5data.SaveTurnCounterOffset]; got != 137 {
		t.Errorf("0x%04X 是 %d,預期 137", u5data.SaveTurnCounterOffset, got)
	}
	span := raw[u5data.SaveRoomsClearedOffset : u5data.SaveRoomsClearedOffset+u5data.RoomsClearedBytes]
	for i := range before {
		if span[i] != before[i] {
			t.Errorf("房間位元陣列第 %d 個位元組是 0x%02X,預期 0x%02X", i, span[i], before[i])
		}
	}

	back, err := u5data.ParseSave(raw)
	if err != nil {
		t.Fatal(err)
	}
	s2 := roomScene(t, 0, 5)
	s2.LoadFrom(back)
	if got := s2.ActiveMember(); got != 1 {
		t.Errorf("讀回來的指定行動者是 %d,預期 1", got)
	}
	if got := s2.TurnsSinceAlms(); got != 137 {
		t.Errorf("讀回來的回合計數是 %d,預期 137", got)
	}
	if !s2.roomIsCleared(0x23, byte(u5data.DungeonRoomF|5)) ||
		!s2.roomIsCleared(0x26, byte(u5data.DungeonRoomF|9)) {
		t.Error("讀回來之後房間紀錄不見了")
	}
	// 反對照:沒清過的房間不能變成清過。
	if s2.roomIsCleared(0x23, byte(u5data.DungeonRoomF|6)) {
		t.Error("沒清過的房間被讀成清過了")
	}
}

// TestRoomsClearedSpanIsFourteenBytes —— ★★ 14 這個長度本身是一條證據。
//
// 索引是 `DungeonRoomBlock*16 + 房號`,而 `DungeonRoomBlock` 有「≥1 就 −1」
// 的修正 ⇒ 八座地牢只佔 7 個區塊 ⇒ 7 × 16 = 112 位元 = 14 位元組。
// 原版 `sub_27D24` 的 `push 0Eh` 獨立佐證了那個共用區塊的怪處。
func TestRoomsClearedSpanIsFourteenBytes(t *testing.T) {
	if u5data.RoomsClearedBytes != 14 {
		t.Fatalf("長度是 %d,原版讀 0x0E = 14", u5data.RoomsClearedBytes)
	}
	// 最大索引要塞得進去。
	maxIdx := u5data.DungeonRoomBlock(u5data.DungeonLocationBase+u5data.DungeonCount-1)*
		u5data.DungeonRoomsPerDungeon + u5data.DungeonRoomsPerDungeon - 1
	if maxIdx/8 >= u5data.RoomsClearedBytes {
		t.Errorf("最大索引 %d 需要 %d 位元組,只有 %d",
			maxIdx, maxIdx/8+1, u5data.RoomsClearedBytes)
	}
	// ★ 而且要**剛好**塞滿 —— 差一個位元組就表示索引方式不是那個修正版。
	if maxIdx/8 != u5data.RoomsClearedBytes-1 {
		t.Errorf("最大索引 %d 落在第 %d 個位元組,預期剛好是最後一個(%d)",
			maxIdx, maxIdx/8, u5data.RoomsClearedBytes-1)
	}
}
