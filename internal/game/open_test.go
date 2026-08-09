package game

import (
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// openCmdScene 是一個「站在城鎮裡、腳邊那一格可以隨意改」的場景。
//
// Open 的地表 / 場景那一條**完全沒被實作過**(引擎原本只印「此處沒有寶箱。」),
// 所以這一整組測試都是新的。
func openCmdScene(t *testing.T) *State {
	t.Helper()
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("需要 U5_GAMEDATA")
	}
	s := realState(t, dir)
	const britain = 2
	if err := s.SetScene(britain, 0, 15, 15); err != nil {
		t.Skipf("進不了場景:%v", err)
	}
	s.Messages = nil
	return s
}

// openNorthAt 把北邊那一格設成 tile,然後對北按 O。
func openNorthAt(t *testing.T, s *State, tile byte) {
	t.Helper()
	if !s.SetTileAt(s.X, s.Y-1, tile) {
		t.Fatalf("寫不進地圖 —— 這條測試沒有在驗任何東西")
	}
	s.Messages = nil
	s.OpenChest()
	if s.Prompt != PromptDirection {
		t.Fatalf("Open 沒問方向(Prompt = %v)—— 原版 `sub_15374` 問完才動作", s.Prompt)
	}
	s.AnswerDirection(North)
}

// TestOpenRefusalsMatchTheJumpTable 逐一釘住 `sub_15374` 的四種拒絕。
func TestOpenRefusalsMatchTheJumpTable(t *testing.T) {
	cases := []struct {
		tile byte
		name string
		want string
	}{
		{u5data.TileMagicLockedA, "魔法鎖 A", MsgLocked},
		{u5data.TileMagicLockedB, "魔法鎖 B", MsgLocked},
		{u5data.TilePortcullis, "柵門", MsgTooHeavy},
		{u5data.TileItsOpen, "0xAF", MsgItsOpen},
		{u5data.TileLockedDoor, "上鎖的門", MsgLocked},
		{u5data.TileLockedMagicDoor, "有窗戶的上鎖的門", MsgLocked},
	}
	for _, c := range cases {
		s := openCmdScene(t)
		openNorthAt(t, s, c.tile)
		if got := strings.Join(s.Messages, "|"); !strings.Contains(got, c.want) {
			t.Errorf("%s(0x%02X)該印「%s」,實得 %q", c.name, c.tile, c.want, s.Messages)
		}
		// 拒絕的那幾種都不該改動地圖。
		if s.TileAt(s.X, s.Y-1) != c.tile {
			t.Errorf("%s 被拒絕了卻改了地圖:0x%02X", c.name, s.TileAt(s.X, s.Y-1))
		}
	}
}

// TestOpenedDoorBecomesFloorAndClosesAfterFourTurns —— ★★ 這條機制引擎完全沒有。
//
// 原版主迴圈 `sub_1A54` 每回合 `dec byte_3E164`,歸零就 `sub_2B64C` 把
// 原本的 tile 寫回去。所以**門會自己關上**,而且只有 4 回合。
func TestOpenedDoorBecomesFloorAndClosesAfterFourTurns(t *testing.T) {
	for _, door := range []byte{u5data.TileDoorA, u5data.TileDoorB} {
		s := openCmdScene(t)
		x, y := s.X, s.Y-1
		openNorthAt(t, s, door)
		if !strings.Contains(strings.Join(s.Messages, "|"), MsgOpened) {
			t.Fatalf("門 0x%02X 沒印「%s」:%q", door, MsgOpened, s.Messages)
		}
		if got := s.TileAt(x, y); got != u5data.OpenedDoorTile {
			t.Errorf("門開了之後是 0x%02X,原版是 0x%02X(磚地)",
				got, u5data.OpenedDoorTile)
		}
		// 撐滿 DoorAutoCloseTurns − 1 回合都還是開的。
		for i := 0; i < u5data.DoorAutoCloseTurns-1; i++ {
			s.tick()
			if s.TileAt(x, y) != u5data.OpenedDoorTile {
				t.Fatalf("第 %d 回合門就關了 —— 原版是 %d 回合",
					i+1, u5data.DoorAutoCloseTurns)
			}
		}
		s.tick()
		if got := s.TileAt(x, y); got != door {
			t.Errorf("第 %d 回合門該關回 0x%02X,實得 0x%02X",
				u5data.DoorAutoCloseTurns, door, got)
		}
	}
}

// TestOnlyOneDoorCanBeOpenAtATime —— ★ 那四個變數只有一組。
//
// 原版 `sub_15374` 在問方向**之前**就先呼叫一次
// `sub_2B64C(byte_3E161, byte_3E162, byte_3E163)` 把上一扇關掉。
// 所以走過一長串門會看到後面的自己關上 —— 那是可觀察的行為,不是最佳化。
func TestOnlyOneDoorCanBeOpenAtATime(t *testing.T) {
	s := openCmdScene(t)
	firstX, firstY := s.X, s.Y-1
	openNorthAt(t, s, u5data.TileDoorA)
	if s.TileAt(firstX, firstY) != u5data.OpenedDoorTile {
		t.Fatal("第一扇門沒開")
	}
	// 再開東邊那一扇 —— 第一扇該立刻關上,不必等倒數。
	if !s.SetTileAt(s.X+1, s.Y, u5data.TileDoorB) {
		t.Fatal("寫不進地圖")
	}
	s.Messages = nil
	s.OpenChest()
	s.AnswerDirection(East)
	if got := s.TileAt(firstX, firstY); got != u5data.TileDoorA {
		t.Errorf("開第二扇門時第一扇該立刻關上(回到 0x%02X),實得 0x%02X",
			u5data.TileDoorA, got)
	}
	if s.TileAt(s.X+1, s.Y) != u5data.OpenedDoorTile {
		t.Error("第二扇門沒開")
	}
}

// TestOpeningAChestInTownCostsKarma —— ★★ 在場景裡開箱子扣 2 點業報。
//
// 「翻別人家的箱子」的代價。原版 `sub_15108` 只在 1 <= 地點 <= 0x20 扣,
// 大地圖上的箱子無主、不扣。
func TestOpeningAChestInTownCostsKarma(t *testing.T) {
	s := openCmdScene(t)
	if s.Objects == nil {
		s.Objects = &u5data.ObjectSet{}
	}
	objs := s.currentObjects()
	if objs == nil {
		t.Skip("這個場景沒有物件層")
	}
	x, y := s.X, s.Y-1
	if !s.SetTileAt(x, y, u5data.OpenedDoorTile) { // 一個「表裡沒有」的 tile
		t.Fatal("寫不進地圖")
	}
	// ⚠ **要記住 Spawn 回傳的槽號**,不能事後用 `ObjectAt(x, y)` 找 ——
	// 城裡那一格上可能還站著 NPC 的鏡射物件,`ObjectAt` 會先撞到它。
	// (第一版就是這樣紅的,而失敗訊息「箱子還在物件層上」指向錯的地方。)
	slot, ok := objs.Spawn(u5data.ObjLockedChest, x, y, s.Floor)
	if !ok {
		t.Skip("物件槽滿了")
	}
	actAs(t, s, 0)
	s.Karma = 50
	s.Messages = nil
	s.OpenChest()
	s.AnswerDirection(North)
	if s.Karma != 50-u5data.ChestOpenKarmaPenalty {
		t.Errorf("在城裡開箱子後業報是 %d,預期 %d",
			s.Karma, 50-u5data.ChestOpenKarmaPenalty)
	}
	// 箱子該從物件層消失(原版 `sub_2B6C8` 清掉那一槽)。
	if objs.Objects[slot].Present() {
		t.Error("開完之後箱子還在物件層上")
	}

	// ★ 業報有下限 0(原版 `cmp al, 2; jbe → mov byte_3E098, 0`)。
	s2 := openCmdScene(t)
	if s2.Objects == nil {
		s2.Objects = &u5data.ObjectSet{}
	}
	o2 := s2.currentObjects()
	if o2 == nil {
		t.Skip("沒有物件層")
	}
	s2.SetTileAt(s2.X, s2.Y-1, u5data.OpenedDoorTile)
	if _, ok := o2.Spawn(u5data.ObjLockedChest, s2.X, s2.Y-1, s2.Floor); !ok {
		t.Skip("物件槽滿了")
	}
	actAs(t, s2, 0)
	s2.Karma = 1
	s2.OpenChest()
	s2.AnswerDirection(North)
	if s2.Karma != 0 {
		t.Errorf("業報 1 開完箱子變成 %d,原版下限是 0", s2.Karma)
	}
}

// TestTrapIsTheTopBitOfTheChestQuality —— ★ 品質一個位元組裝兩件事。
//
// 最高位 = 有陷阱、低七位 = 獎品等級。原版 `and var_5, 7Fh` 清掉陷阱位
// 之後才拿去擲獎品 —— 少了那一步,有陷阱的箱子會被當成 128 級的寶箱。
func TestTrapIsTheTopBitOfTheChestQuality(t *testing.T) {
	s := openCmdScene(t)
	if s.Objects == nil {
		s.Objects = &u5data.ObjectSet{}
	}
	objs := s.currentObjects()
	if objs == nil {
		t.Skip("沒有物件層")
	}
	x, y := s.X, s.Y-1
	s.SetTileAt(x, y, u5data.OpenedDoorTile)
	slot, ok := objs.Spawn(u5data.ObjLockedChest, x, y, s.Floor)
	if !ok {
		t.Skip("物件槽滿了")
	}
	// 同上:用 Spawn 回傳的槽號,不要用 `ObjectAt` 找。
	objs.Objects[slot].Raw[u5data.ObjQuality] = u5data.ChestTrapQualityBit | 5
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		s.Roster[i].HP = 200
		s.Roster[i].Status = u5data.StatusGood
	}
	actAs(t, s, 0)
	s.Messages = nil
	s.OpenChest()
	s.AnswerDirection(North)
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgTrapped) {
		t.Errorf("最高位設著卻沒印「%s」:%q", MsgTrapped, s.Messages)
	}
	// ⚠ 判準是「扣血**或**中毒」。四種陷阱裡毒(2/8)與毒氣(1/8)只改狀態
	// 不扣血(`docs/re/91`)⇒ 只看 HP 的話這條有 3/8 的機率誤判,
	// 而 itest 沒有固定種子 → 會變成間歇紅燈。
	hurt := false
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		if s.Roster[i].HP < 200 || s.Roster[i].Status == u5data.StatusPoisoned {
			hurt = true
		}
	}
	if !hurt {
		t.Errorf("有陷阱卻既沒人扣血也沒人中毒:%q", s.Messages)
	}
}

// TestTheSandalwoodBoxCannotBeOpened —— 物件種類 0x0E 印 "Can't!"。
//
// ⚠ 與「沒東西可開」是**兩句不同的話**:原版分得很清楚,
// 因為木盒就在那裡,只是這個指令打不開它。
func TestTheSandalwoodBoxCannotBeOpened(t *testing.T) {
	s := openCmdScene(t)
	if s.Objects == nil {
		s.Objects = &u5data.ObjectSet{}
	}
	objs := s.currentObjects()
	if objs == nil {
		t.Skip("沒有物件層")
	}
	x, y := s.X, s.Y-1
	s.SetTileAt(x, y, u5data.OpenedDoorTile)
	objs.Spawn(u5data.ObjSandalwoodBox, x, y, s.Floor)
	s.Messages = nil
	s.OpenChest()
	s.AnswerDirection(North)
	got := strings.Join(s.Messages, "|")
	if !strings.Contains(got, MsgCantOpenThat) {
		t.Errorf("檀香木盒該印「%s」,實得 %q", MsgCantOpenThat, s.Messages)
	}
	if strings.Contains(got, MsgNothingToOpen) {
		t.Error("印成了「沒有東西可以開」—— 原版那是另一句")
	}
}

// TestNothingToOpenWhenTheSquareIsBare —— 空地印 "Nothing to open!"。
func TestNothingToOpenWhenTheSquareIsBare(t *testing.T) {
	s := openCmdScene(t)
	s.SetTileAt(s.X, s.Y-1, u5data.OpenedDoorTile)
	s.Messages = nil
	s.OpenChest()
	s.AnswerDirection(North)
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgNothingToOpen) {
		t.Errorf("空地該印「%s」,實得 %q", MsgNothingToOpen, s.Messages)
	}
}
