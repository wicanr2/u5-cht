package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 紮營能回血 —— 但**要睡滿六小時以上**,而且回的是 random(1,63) 不是回滿。
func TestCampRestoresTheParty(t *testing.T) {
	s := dungeonState(t)
	s.Location, s.Floor = 0, 0
	s.Transport = u5data.VehicleWalk
	s.X, s.Y = 82, 106 // 不列顛城旁的陸地
	c := &s.Roster[0]
	c.HP = 1
	c.MP = 0
	s.Clock.Hour, s.Clock.Minute = 10, 0

	s.HoleUp()
	if !s.AwaitingNumber() {
		t.Fatalf("紮營沒有問時數:\n%s", s.log())
	}
	s.AnswerNumber(9)
	if s.AwaitingYesNo() {
		s.AnswerYesNo(false)
	}
	if s.InCombat() {
		t.Skip("這一次被突襲了,換一顆種子再測恢復")
	}
	if c.HP <= 1 {
		t.Errorf("睡了九小時 HP 還是 %d:\n%s", c.HP, s.log())
	}
	for _, m := range s.Party() {
		if m.Status == u5data.StatusAsleep {
			t.Errorf("%s 醒不過來", m.Name)
		}
	}
}

// ★ 睡不到六小時**完全沒效果**(原版 `cmp arg_8, 5; jle → "No effect..."`)。
//
// 我第一版把旅店那支恢復拿來用,等於「睡三小時就回滿血」——
// 那讓紮營變成無限回血機,而原版刻意設了門檻與冷卻。
func TestShortCampDoesNothing(t *testing.T) {
	s := dungeonState(t)
	s.Location, s.Floor = 0, 0
	s.Transport = u5data.VehicleWalk
	s.X, s.Y = 82, 106
	c := &s.Roster[0]
	c.HP = 1
	s.Clock.Hour, s.Clock.Minute = 10, 0

	s.HoleUp()
	s.AnswerNumber(3) // 只睡三小時
	if s.AwaitingYesNo() {
		s.AnswerYesNo(false)
	}
	if s.InCombat() {
		t.Skip("被突襲了")
	}
	if c.HP != 1 {
		t.Errorf("睡三小時竟然回了血:%d", c.HP)
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), "毫無效果") {
		t.Errorf("沒說毫無效果:\n%s", s.log())
	}
}

// 冷卻沒退完再紮營也是「毫無效果」—— 擋住連續紮營刷血。
func TestCampCooldownBlocksRepeats(t *testing.T) {
	s := dungeonState(t)
	s.Location, s.Floor = 0, 0
	s.Transport = u5data.VehicleWalk
	s.X, s.Y = 82, 106
	s.Clock.Hour = 10
	s.RestCooldown = campRestCooldown
	c := &s.Roster[0]
	c.HP = 1

	s.HoleUp()
	s.AnswerNumber(9)
	if s.AwaitingYesNo() {
		s.AnswerYesNo(false)
	}
	if s.InCombat() {
		t.Skip("被突襲了")
	}
	if c.HP != 1 {
		t.Errorf("冷卻期間竟然回了血:%d", c.HP)
	}
}

// 守夜的人什麼都不恢復 —— 那是派人守夜要付的代價。
func TestTheWatchGetsNoRest(t *testing.T) {
	s := dungeonState(t)
	s.Location, s.Floor = 0, 0
	s.Transport = u5data.VehicleWalk
	s.X, s.Y = 82, 106
	s.Clock.Hour = 10
	for i := 0; i < s.PartySize; i++ {
		s.Roster[i].HP = 1
		s.Roster[i].Status = u5data.StatusGood
	}
	// 直接呼叫 camp,跳過選單(選單的路徑另有測試)。
	s.camp(9, 0)
	if s.InCombat() {
		t.Skip("被突襲了")
	}
	if s.Roster[0].HP != 1 {
		t.Errorf("守夜的人竟然回了血:%d", s.Roster[0].HP)
	}
	if s.PartySize > 1 && s.Roster[1].HP <= 1 {
		t.Errorf("沒守夜的人反而沒回血:%d", s.Roster[1].HP)
	}
}

// ★★ 睡床**什麼都不恢復**,只是讓時間過去,而且起床會往東挪一格。
//
// 憑「睡在床上當然比野地舒服」的直覺去補恢復,就是自創遊戲。
func TestSleepingInABedHealsNothing(t *testing.T) {
	s := dungeonState(t)
	s.Location, s.Floor = 0, 0
	s.Clock.Hour = 10
	c := &s.Roster[0]
	c.HP, c.MP = 1, 0
	x := s.X

	s.sleepInBed(8)
	if c.HP != 1 {
		t.Errorf("睡床竟然回了血:%d", c.HP)
	}
	if c.MP != 0 {
		t.Errorf("睡床竟然回了法力:%d", c.MP)
	}
	if s.Clock.Hour != 18 {
		t.Errorf("10 時睡 8 小時醒在 %d 時,預期 18 時", s.Clock.Hour)
	}
	if s.X != x+1 {
		t.Errorf("起床沒有從床上挪開:x %d → %d", x, s.X)
	}
}

// ★ 跨午夜時原版會**多醒一小時** —— `sub edi, 17h` 減的是 23 不是 24。
//
// 這是原版的 bug,而 CLAUDE.md §3.0 要求機制與原版一模一樣、包括它的 bug。
// 「順手修好」會讓時間對不上原版,而那種差異只有並排跑才看得出來。
func TestSleepPastMidnightKeepsTheOriginalOffByOne(t *testing.T) {
	s := dungeonState(t)
	s.Location, s.Floor = 0, 0
	s.Transport = u5data.VehicleWalk
	s.X, s.Y = 82, 106
	s.Clock.Hour, s.Clock.Minute = 22, 0

	s.sleepInBed(4) // 22 + 4 = 26 → 26 − 23 = 3(正確的環繞會是 2)
	if s.Clock.Hour != 3 {
		t.Errorf("22 時睡 4 小時醒在 %d 時;原版的算法會給 3 時(減 23 而非 24)", s.Clock.Hour)
	}

	// ★ 對照:**紮營那條路減的是 24**,環繞是對的。
	// 同一個遊戲兩條休息路徑,一條對一條錯 —— 所以那是真的寫錯,不是刻意設計。
	s2 := dungeonState(t)
	s2.Location, s2.Floor = 0, 0
	s2.Transport = u5data.VehicleWalk
	s2.X, s2.Y = 82, 106
	s2.Clock.Hour, s2.Clock.Minute = 22, 0
	s2.camp(4, -1)
	if s2.InCombat() {
		t.Skip("被突襲了")
	}
	if s2.Clock.Hour != 2 {
		t.Errorf("紮營:22 時睡 4 小時醒在 %d 時,預期 2 時(減 24)", s2.Clock.Hour)
	}
}

// 城裡沒躺在床上就睡不著。
func TestHoleUpInTownNeedsABed(t *testing.T) {
	s := dungeonState(t)
	loc := &u5data.Locations[1] // 不列顛城
	s.Location, s.Floor = 0, 0
	s.X, s.Y = loc.X, loc.Y
	s.Transport = u5data.VehicleWalk
	s.Enter()
	if !s.InScene() {
		t.Skip("進不了城")
	}
	s.Messages = nil
	s.HoleUp()
	if s.AwaitingNumber() {
		t.Fatalf("沒躺在床上卻問了時數:\n%s", s.log())
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), "床") {
		t.Errorf("沒說要在床上:\n%s", s.log())
	}
	// 腳下換成床就問得出時數了。
	s.SetTileAt(s.X, s.Y, HoleUpBedTile)
	s.HoleUp()
	if !s.AwaitingNumber() {
		t.Errorf("站在床上還是睡不著:\n%s", s.log())
	}
}

// 在船上按 H 是修船,不是紮營;而且揚著帆修不了。
func TestHoleUpOnAShipRepairsIt(t *testing.T) {
	s := dungeonState(t)
	s.Location, s.Floor = 0, 0
	s.Transport = u5data.VehicleSailing
	s.ShipHull = 3
	s.Messages = nil
	s.HoleUp()
	if !strings.Contains(strings.Join(s.Messages, "|"), "收帆") {
		t.Errorf("揚著帆竟然修得起來:\n%s", s.log())
	}
	if s.ShipHull != 3 {
		t.Errorf("耐久被改了:%d", s.ShipHull)
	}

	// 收帆之後修得動,而且**一定會修到 10 以上**(原版是 do-while)。
	s.Transport = u5data.VehicleShip
	s.HoleUp()
	if s.ShipHull < shipRepairUntil {
		t.Errorf("修完耐久只有 %d,原版的 do-while 會一路加到 %d 以上",
			s.ShipHull, shipRepairUntil)
	}
	if s.ShipHull > ShipHullMax {
		t.Errorf("耐久 %d 超過上限 %d", s.ShipHull, ShipHullMax)
	}
	// 修船不該問時數 —— 它與睡覺是兩條路。
	if s.AwaitingNumber() {
		t.Error("修船竟然問了時數")
	}
}

// 騎著馬紮不了營(原版 `cmp byte…, 1Ch` 只認步行)。
func TestCampNeedsToBeOnFoot(t *testing.T) {
	s := dungeonState(t)
	s.Location, s.Floor = 0, 0
	s.X, s.Y = 82, 106
	s.Transport = u5data.TileHorse
	s.Messages = nil
	s.HoleUp()
	if s.AwaitingNumber() {
		t.Fatalf("騎著馬竟然紮起營來:\n%s", s.log())
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), "步行") {
		t.Errorf("沒說要步行:\n%s", s.log())
	}
}

// 戰鬥中不可用的鍵**各有自己的回應**,不是統一一句「不行」。
//
// 而且 D 與 W 在戰鬥分派器裡也是 "-What?" —— 與主分派器一致,
// 兩處獨立佐證它們不是指令。
func TestCombatRefusalsAreIndividual(t *testing.T) {
	cases := map[rune]string{
		'E': "此處不可",
		'T': "無人回應",
		'B': "對什麼",
		'D': "何事",
		'W': "何事",
	}
	for key, want := range cases {
		got, ok := CombatRefuse(key)
		if !ok {
			t.Errorf("%c 應該在拒絕清單裡", key)
			continue
		}
		if !strings.Contains(got, want) {
			t.Errorf("%c 回的是 %q,應含 %q", key, got, want)
		}
	}
	// 可用的鍵不該落在拒絕清單裡 —— 兩張表重疊就是抄錯了。
	for _, key := range CombatAllowedKeys {
		if _, ok := CombatRefuse(key); ok {
			t.Errorf("%c 同時出現在可用與拒絕兩張表", key)
		}
	}
	// 兩張表加起來應該蓋掉 A..Z 全部 26 個字母(原版 case 65..90 全有著落)。
	seen := map[rune]bool{}
	for _, k := range CombatAllowedKeys {
		seen[k] = true
	}
	for k := range combatRefusals {
		seen[k] = true
	}
	for k := 'A'; k <= 'Z'; k++ {
		if !seen[k] {
			t.Errorf("字母 %c 兩張表都沒有 —— jpt_A5C8 的 case %d 漏抄了", k, k)
		}
	}
}

// 戰鬥中的 Get / Search / Push 那幾支**與地圖上是同一份程式**。
//
// 驗的是三件事:(1) 戰鬥中 `TileAt` 讀的是戰場、(2) 行動回合開始時
// `State.X/Y` 指到行動者身上、(3) 離場時還原成進戰鬥前的世界座標。
// 這三件湊齊,那九個指令就不需要「戰場版」—— 原版也沒有。
func TestCombatSharesTheMapAccessors(t *testing.T) {
	s := dungeonState(t)
	s.Location = 0
	e := u5data.DungeonEntrances[0]
	s.X, s.Y = e.X, e.Y
	s.Enter()
	d := s.Dungeon
	d.X, d.Y = 4, 4
	d.Facing = North
	worldX, worldY := s.X, s.Y

	d.Monster = &DungeonMonster{
		Kind: 0, Creature: u5data.DungeonMonsterCreature(0),
		X: 4, Y: 3, PrevX: 4, PrevY: 3,
	}
	s.dungeonMonsterAttacks()
	if !s.InCombat() {
		t.Fatalf("沒開打:\n%s", s.log())
	}
	c := s.Combat
	if c.Turn < 0 {
		t.Skip("這一輪先由敵人行動")
	}
	u := &c.Units[c.Turn]
	// (2) 座標借給了行動者。
	if s.X != u.X || s.Y != u.Y {
		t.Errorf("隊伍座標是 (%d,%d),行動者在 (%d,%d)", s.X, s.Y, u.X, u.Y)
	}
	// (1) TileAt 讀的是戰場;寫進去也只動戰場那份副本。
	if got, want := s.TileAt(u.X, u.Y), s.CombatTileAt(u.X, u.Y); got != want {
		t.Errorf("TileAt 在戰鬥中回 0x%02X,戰場上是 0x%02X", got, want)
	}
	const probe = 0x44
	if !s.SetTileAt(u.X, u.Y, probe) {
		t.Fatal("戰鬥中寫不進戰場")
	}
	if s.CombatTileAt(u.X, u.Y) != probe {
		t.Error("寫進去的沒有落在戰場上")
	}
	// 而且沒有污染地牢那一層。
	if s.Dungeons.At(d.Index, d.Level, u.X, u.Y) == probe && (u.X != 4 || u.Y != 4) {
		t.Error("戰鬥中的寫入漏到地牢地圖上了")
	}
	// (3) 離場還原世界座標。
	s.EndCombat(true)
	if s.X != worldX || s.Y != worldY {
		t.Errorf("離場後座標是 (%d,%d),應還原成 (%d,%d)", s.X, s.Y, worldX, worldY)
	}
}

// 戰鬥中不必挑人 —— 輪到誰就是誰。
func TestCombatPicksTheActingMember(t *testing.T) {
	s := dungeonState(t)
	s.Location = 0
	e := u5data.DungeonEntrances[0]
	s.X, s.Y = e.X, e.Y
	s.Enter()
	d := s.Dungeon
	d.X, d.Y = 4, 4
	d.Monster = &DungeonMonster{
		Kind: 0, Creature: u5data.DungeonMonsterCreature(0),
		X: 4, Y: 3, PrevX: 4, PrevY: 3,
	}
	s.dungeonMonsterAttacks()
	if !s.InCombat() || s.Combat.Turn < 0 {
		t.Skip("這一輪先由敵人行動")
	}
	want := s.Combat.Units[s.Combat.Turn].Roster
	if want < 0 {
		t.Skip("行動者不是隊員")
	}
	if got := s.pickCharacter(""); got != want {
		t.Errorf("戰鬥中挑到第 %d 位,應是行動者第 %d 位", got, want)
	}
}

// 戰鬥中踩在梯子上按 K 會**離場**,而且地牢會換樓層。
func TestCombatKlimbLeavesViaTheLadder(t *testing.T) {
	s := dungeonState(t)
	s.Location = 0
	e := u5data.DungeonEntrances[0]
	s.X, s.Y = e.X, e.Y
	s.Enter()
	d := s.Dungeon
	d.X, d.Y = 4, 4
	d.Level = 3
	d.Monster = &DungeonMonster{
		Kind: 0, Creature: u5data.DungeonMonsterCreature(0),
		X: 4, Y: 3, PrevX: 4, PrevY: 3,
	}
	s.dungeonMonsterAttacks()
	if !s.InCombat() || s.Combat.Turn < 0 {
		t.Skip("這一輪先由敵人行動")
	}
	c := s.Combat
	u := &c.Units[c.Turn]
	// 把行動者腳下換成上行梯。
	if !s.SetCombatTileAt(u.X, u.Y, CombatLadderUp) {
		t.Fatal("寫不進戰場")
	}
	c.LadderBoth = false
	before := d.Level

	s.Messages = nil
	s.CombatKlimb()
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgUp) {
		t.Errorf("踩在梯子上按 K 沒有往上:\n%s", s.log())
	}
	if c.ExitDir != CombatExitUp {
		t.Errorf("出口記成 %d,預期 %d", c.ExitDir, CombatExitUp)
	}
	// 隊伍只剩三人時,一個人離場不會馬上結束;真的結束了就要換層。
	if !s.InCombat() && s.Dungeon != nil && s.Dungeon.Level != before-1 {
		t.Errorf("離場後樓層是 %d,預期 %d", s.Dungeon.Level, before-1)
	}
}

// 不在梯子上按 K 會問方向,而**只有 tile 0x4C 爬得過去**。
func TestCombatKlimbNeedsAClimbableTile(t *testing.T) {
	s := dungeonState(t)
	s.Location = 0
	e := u5data.DungeonEntrances[0]
	s.X, s.Y = e.X, e.Y
	s.Enter()
	d := s.Dungeon
	d.X, d.Y = 4, 4
	d.Monster = &DungeonMonster{
		Kind: 0, Creature: u5data.DungeonMonsterCreature(0),
		X: 4, Y: 3, PrevX: 4, PrevY: 3,
	}
	s.dungeonMonsterAttacks()
	if !s.InCombat() || s.Combat.Turn < 0 {
		t.Skip("這一輪先由敵人行動")
	}
	c := s.Combat
	u := &c.Units[c.Turn]
	s.SetCombatTileAt(u.X, u.Y, 0x44) // 普通地板 → 不是梯子
	if u.Y < 1 {
		t.Skip("行動者貼著上緣,測不到北邊")
	}

	// 北邊放一塊不能爬的東西。
	s.SetCombatTileAt(u.X, u.Y-1, 0x44)
	s.Messages = nil
	s.CombatKlimb()
	if !s.AwaitingDirection() {
		t.Fatalf("不在梯子上按 K 沒有問方向:\n%s", s.log())
	}
	bx, by := u.X, u.Y
	s.AnswerDirection(North)
	if u.X != bx || u.Y != by {
		t.Errorf("爬上了不能爬的東西:(%d,%d) → (%d,%d)", bx, by, u.X, u.Y)
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgWhat) {
		t.Errorf("沒回「%s」:\n%s", MsgWhat, s.log())
	}

	// 換成 0x4C 就爬得上去。
	s.SetCombatTileAt(u.X, u.Y-1, CombatClimbable)
	s.CombatKlimb()
	s.AnswerDirection(North)
	if u.Y != by-1 {
		t.Errorf("0x4C 爬不上去:(%d,%d) → (%d,%d)", bx, by, u.X, u.Y)
	}
}

// 隊員在戰場上的圖 —— ★ **依職業**,而 0x4C 就是其中一格。
//
// ⚠⚠ 這條測試整個反轉過(`docs/re/72`)。它原本叫
// `TestPartyBattlefieldTileIsNotTheClimbable`,主張「0x1D / 0x1E 兩個值,
// **不是** 0x4C,因為 0x4C 是 `sub_16058` 判爬得過去的那一格」。
//
// 那個推論的前提是「一個 tile 不會同時是隊伍自己與爬得過去的東西」——
// **前提本身錯了**。開戰佈陣寫的是 `byte_40C34[職業]`,而聖者那一格正是 0x4C;
// `sub_16058` 拿 0x4C 判「爬得過去」,恰恰是因為那一格站著隊員。
//
// 現在這條改成釘住職業表,並保留原來那兩條**還是對的**斷言
// (0x1D / 0x1E 相鄰、0x1D 緊接在步行的 0x1C 之後)。
func TestPartyBattlefieldTileComesFromTheClassTable(t *testing.T) {
	// 站著與躺著要是相鄰的兩格 —— 原版一支寫 0x1E、另一支寫 0x1D。
	if PartyTileLying != PartyTileStanding+1 {
		t.Errorf("0x%02X 與 0x%02X 不相鄰", PartyTileStanding, PartyTileLying)
	}
	// 而且要與世界地圖上步行的隊伍同一族(`sub_16DA4` 收 0x1C 與 0x1D)。
	if PartyTileStanding != int(u5data.VehicleWalk)+1 {
		t.Errorf("站著的 0x%02X 不緊接在步行的 0x%02X 之後",
			PartyTileStanding, u5data.VehicleWalk)
	}

	// ★ 九個職業字母逐一對上 `byte_40C34`,而且 `CombatClimbable` 必須是
	// 其中之一 —— 那是「前提錯了」的直接證據,不是巧合。
	want := map[byte]byte{
		'A': 0x4C, 'M': 0x40, 'B': 0x44, 'F': 0x48,
		'D': 0x4C, 'T': 0x4C, 'P': 0x4C, 'R': 0x4C, 'S': 0x4C,
	}
	for class, tile := range want {
		ch := &u5data.Character{Status: u5data.StatusGood, Class: class}
		if got := partyTileFor(ch); got != tile {
			t.Errorf("職業 %q 的圖是 0x%02X,原版是 0x%02X", string(class), got, tile)
		}
	}
	if u5data.PartyCombatTile(&u5data.Character{Class: 'A'}) != byte(CombatClimbable) {
		t.Errorf("聖者的圖該正是 `sub_16058` 判爬得過去的那格 0x%02X", CombatClimbable)
	}
	// 四個不同的值,不多不少。
	seen := map[byte]bool{}
	for _, tile := range u5data.PartyCombatTiles {
		seen[tile] = true
	}
	if len(seen) != 4 {
		t.Errorf("職業表有 %d 個不同的值,原版是 4 個(0x40/0x44/0x48/0x4C)", len(seen))
	}

	// 睡著與倒下走恢復路徑,不查職業表。
	ch := &u5data.Character{Status: u5data.StatusAsleep, Class: 'M'}
	if got := partyTileFor(ch); got != PartyTileLying {
		t.Errorf("睡著的隊員畫成 0x%02X", got)
	}
	ch.Status = u5data.StatusDead
	if got := partyTileFor(ch); got != PartyTileLying {
		t.Errorf("倒下的隊員畫成 0x%02X", got)
	}
}
