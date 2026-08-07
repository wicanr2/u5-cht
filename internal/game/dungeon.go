package game

import (
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 地牢
//
// 八座地牢的地點編號是 0x21..0x28,每座 8 層、每層 8×8 格。
// 進去之後的操作與地面完全不同 —— 原版是**第一人稱**視角:
// 方向鍵不是走四方,而是「前進 / 左轉 / 右轉 / 後退」。
//
// ⚠ **畫面還是俯視的。** 原版的第一人稱透視圖要先把 `DNG1-3.16`
// 那三組圖解出來,那還沒做;規則(移動、朝向、爬梯、房間、陷阱)是照原版的,
// 呈現方式不是。文件與 README 都要照實寫,不能讓「地牢做完了」變成一句謊。

// DungeonEntryFromAbove / FromBelow 是進地牢的落點與樓層。
//
// `sub_2D564`:從地面進去是第 1 層(索引 0);從地底世界鑽上來是第 8 層
//(索引 7)而且落在 (7,7) —— 唯一的例外是 **Doom**(地點 0x28),
// 它從地底世界進去也走地面那條。
const (
	DungeonEntryX = 1
	DungeonEntryY = 1
	// DungeonUnderX / Y 是從地底世界進來的落點。
	DungeonUnderX = 7
	DungeonUnderY = 7
	// DungeonDoomLocation 是 Doom,從地底世界進去不套用上面那條。
	DungeonDoomLocation = 0x28
)

// DungeonState 是「正在地牢裡」的狀態。
type DungeonState struct {
	// Index 是第幾座(0..7);Location 是地點編號(0x21 + Index)。
	Index    int
	Location int
	// Level 是第幾層(0 = 最上面一層)。
	Level int
	// X, Y 是層內座標。
	X, Y int
	// Facing 是面向哪邊 —— 第一人稱的「前」。
	Facing Direction
}

// InDungeon 回報是不是在地牢裡。
func (s *State) InDungeon() bool { return s.Dungeon != nil }

// EnterDungeon 進第 n 座地牢(0..7)。fromBelow 為真代表從地底世界鑽上來。
func (s *State) EnterDungeon(n int, fromBelow bool) bool {
	if s.Dungeons == nil || n < 0 || n >= u5data.DungeonCount {
		s.Log("此處沒有地牢入口。")
		return false
	}
	loc := u5data.DungeonLocationBase + n
	d := &DungeonState{Index: n, Location: loc, Facing: South}
	if fromBelow && loc != DungeonDoomLocation {
		d.Level = u5data.DungeonLevels - 1
		d.X, d.Y = DungeonUnderX, DungeonUnderY
	} else {
		d.Level = 0
		d.X, d.Y = DungeonEntryX, DungeonEntryY
	}
	s.Dungeon = d
	s.Location = loc
	s.Log("汝踏入了幽深的地底。")
	s.onDungeonTile()
	return true
}

// LeaveDungeon 離開地牢回到大地圖。
func (s *State) LeaveDungeon() {
	if s.Dungeon == nil {
		return
	}
	s.Dungeon = nil
	s.Location = 0
	s.Log("汝回到了地面。")
}

// DungeonTileAt 取地牢裡某一格。
func (s *State) DungeonTileAt(x, y int) byte {
	d := s.Dungeon
	if d == nil {
		return u5data.DungeonWall
	}
	return s.Dungeons.At(d.Index, d.Level, x, y)
}

// DungeonTileHere 是腳下這一格。
func (s *State) DungeonTileHere() byte {
	d := s.Dungeon
	if d == nil {
		return u5data.DungeonWall
	}
	return s.DungeonTileAt(d.X, d.Y)
}

// DungeonTurn 轉向(左 / 右各 90 度)。轉向**不花時間**。
func (s *State) DungeonTurn(left bool) {
	d := s.Dungeon
	if d == nil {
		return
	}
	if left {
		d.Facing = d.Facing.TurnLeft()
	} else {
		d.Facing = d.Facing.TurnRight()
	}
	s.Log("汝轉向" + d.Facing.Name() + "。")
}

// DungeonForward 往面向的方向走一步。back 為真時後退(朝向不變)。
func (s *State) DungeonForward(back bool) {
	d := s.Dungeon
	if d == nil {
		return
	}
	dx, dy := d.Facing.Delta()
	if back {
		dx, dy = -dx, -dy
	}
	s.dungeonStep(d.X+dx, d.Y+dy)
}

// DungeonMove 讓玩家直接往某個絕對方向走(給俯視畫面用的方便鍵)。
//
// 原版沒有這個 —— 第一人稱只有前後與轉向。留著是因為目前畫面是俯視的,
// 兩者並存不影響規則。
func (s *State) DungeonMove(dir Direction) {
	d := s.Dungeon
	if d == nil {
		return
	}
	d.Facing = dir
	dx, dy := dir.Delta()
	s.dungeonStep(d.X+dx, d.Y+dy)
}

// dungeonStep 走到 (nx, ny)。
func (s *State) dungeonStep(nx, ny int) {
	d := s.Dungeon
	if nx < 0 || nx >= u5data.DungeonSide || ny < 0 || ny >= u5data.DungeonSide {
		s.Log(MsgBlocked)
		return
	}
	if u5data.DungeonPlayerBlocks(s.DungeonTileAt(nx, ny)) {
		s.Log(MsgBlocked)
		return
	}
	d.X, d.Y = nx, ny
	s.AdvanceTime(MinutesPerTurn)
	s.onDungeonTile()
}

// onDungeonTile 處理踩到的那一格(原版 `sub_5150`)。
//
// 優先序照原版:先看是不是房間(高四位元 0xA0 或 0xF0),再看完整位元組
// 對應到哪一種陷阱。噴泉與寶箱要玩家自己動手(Look / Open),不是踩到就觸發。
func (s *State) onDungeonTile() {
	d := s.Dungeon
	if d == nil {
		return
	}
	tile := s.DungeonTileHere()
	if u5data.DungeonIsRoom(tile) {
		s.enterDungeonRoom(tile)
		return
	}
	switch tile {
	case u5data.DungeonPitTrapA, u5data.DungeonPitTrapB:
		s.dungeonPitTrap()
	case u5data.DungeonBombTrapA, u5data.DungeonBombTrapB:
		s.Log("炸彈陷阱!")
		s.Log("轟隆!!")
		// 原版把這一格清成 `& 8` —— 只留「頭上有洞」那一位元,陷阱用掉了。
		s.Dungeons.Set(d.Index, d.Level, d.X, d.Y, tile&u5data.DungeonHoleAbove)
		s.damageWholeParty()
	case u5data.DungeonSleepA, u5data.DungeonSleepB:
		s.Log("睡眠法術!")
		s.sleepWholeParty()
	case u5data.DungeonFireA, u5data.DungeonFireB:
		s.Log("烈焰!")
		s.damageWholeParty()
	}
}

// dungeonPitTrap 掉進陷阱坑(原版 `sub_4EB8`)。
//
//	把腳下這一格的低三位元清掉(陷阱用過了)
//	樓層 +1
//	落點那一格加上「頭上有洞」的位元 —— 之後可以用繩索爬回去
//	全隊受傷
func (s *State) dungeonPitTrap() {
	d := s.Dungeon
	if d.Level >= u5data.DungeonLevels-1 {
		return
	}
	s.Log("陷阱坑!")
	s.Log("墜落中……")
	here := s.DungeonTileHere()
	s.Dungeons.Set(d.Index, d.Level, d.X, d.Y, here&0xF8)
	d.Level++
	below := s.DungeonTileAt(d.X, d.Y)
	if below < u5data.DungeonMagic {
		s.Dungeons.Set(d.Index, d.Level, d.X, d.Y, below|u5data.DungeonHoleAbove)
	}
	s.damageWholeParty()
	s.onDungeonTile()
}

// DungeonKlimb 是地牢裡的 Klimb(原版 `sub_417C`)。up 為真往上。
func (s *State) DungeonKlimb(up bool) {
	d := s.Dungeon
	if d == nil {
		return
	}
	tile := s.DungeonTileHere()
	if up {
		if !u5data.DungeonCanClimbUp(tile, s.hasRope()) {
			s.Log(MsgNothingToClimb)
			return
		}
		if d.Level == 0 {
			s.LeaveDungeon()
			return
		}
		d.Level--
	} else {
		if !u5data.DungeonCanClimbDown(tile) {
			s.Log(MsgNothingToClimb)
			return
		}
		if d.Level >= u5data.DungeonLevels-1 {
			s.Log("再下去就沒有路了。")
			return
		}
		d.Level++
	}
	s.AdvanceTime(MinutesPerTurn)
	if up {
		s.Log(MsgUp)
	} else {
		s.Log(MsgDown)
	}
	s.onDungeonTile()
}

// hasRope 回報身上有沒有繩索(原版 `byte_3DFBB`)。
//
// ⚠ 那個位元組在存檔裡的位移還沒對出來,所以先一律當成**沒有** ——
// 少一條爬回去的路,總比讓玩家憑空穿牆好。
func (s *State) hasRope() bool { return false }

// DungeonChangeLevel 是 Uus Por / Des Por 在地牢裡的效果(原版 `sub_3F34` → `sub_3ED0`)。
//
// ⚠ 目的地那一格必須是**純通道**(高四位元 0)。原版寫成
// `if (t == 0 || t == 0xE0 || …) return 0`,而第一個條件一成立後面都到不了
// —— 組語裡看得到那幾個 `cmp` 是死碼。照抄實際行為,不照抄意圖。
func (s *State) DungeonChangeLevel(delta int) bool {
	d := s.Dungeon
	if d == nil {
		return false
	}
	next := d.Level + delta
	if next < 0 || next >= u5data.DungeonLevels {
		s.Log("此處無路可去。")
		return false
	}
	if u5data.DungeonKind(s.Dungeons.At(d.Index, next, d.X, d.Y)) != u5data.DungeonPassage {
		s.Log("此處無路可去。")
		return false
	}
	d.Level = next
	if delta > 0 {
		s.Log(MsgDown)
	} else {
		s.Log(MsgUp)
	}
	s.onDungeonTile()
	return true
}

// enterDungeonRoom 進地牢房間(原版 `sub_42CC`)。
//
// 房間用的是**與地表戰鬥完全相同的 352 B 地圖格式**,所以直接開一場戰鬥。
func (s *State) enterDungeonRoom(tile byte) {
	d := s.Dungeon
	if s.DungeonRooms == nil {
		s.Log("(地牢房間資料未載入)")
		return
	}
	idx := u5data.DungeonRoomIndex(d.Location, tile)
	if idx < 0 || idx >= len(s.DungeonRooms.Maps) {
		s.Log("(房號 " + itoa(u5data.DungeonRoomNumber(tile)) + " 超出範圍)")
		return
	}
	s.Log("進入房間……")
	s.beginRoomCombat(&s.DungeonRooms.Maps[idx], idx)
}

// damageWholeParty 讓全隊受傷(原版 `sub_2A4D0`)。
//
// ⚠ 傷害值還沒逆出來 —— `sub_2A4D0` 內部另有一套。這裡用陷阱最常見的
// `random(1, 20)`(與下毒攻擊 `sub_B8DC` 的 `random(0,20)` 同量級),
// 並在文件裡標明是**估計值**。
func (s *State) damageWholeParty() {
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		ch := &s.Roster[i]
		if ch.Status == u5data.StatusDead {
			continue
		}
		dmg := s.Roll(1, 20)
		hp := int(ch.HP) - dmg
		if hp <= 0 {
			hp = 0
			ch.Status = u5data.StatusDead
			s.Log(ch.Name + "倒下了!")
		}
		ch.HP = uint16(hp)
	}
}

// sleepWholeParty 讓全隊睡著(原版 `sub_4DC8` 印「Sleep spell!」)。
func (s *State) sleepWholeParty() {
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		ch := &s.Roster[i]
		if ch.Status == u5data.StatusGood {
			ch.Status = u5data.StatusAsleep
		}
	}
}
