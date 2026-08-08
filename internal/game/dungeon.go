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
	// Monster 是這一層的遊蕩怪物(見 dungeonmonster.go);nil 代表這一層沒有。
	Monster *DungeonMonster
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
	s.spawnDungeonMonster()
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
	// ⚠ 站在門口轉不了身(`sub_48F4` 的「Not in doorway!」)。
	// 轉身 180 度不受限,但那走 DungeonTurnAround。
	if !u5data.DungeonCanTurn(s.DungeonTileHere()) {
		s.Log("在門口轉不了身!")
		return
	}
	if left {
		d.Facing = d.Facing.TurnLeft()
	} else {
		d.Facing = d.Facing.TurnRight()
	}
	s.Log("汝轉向" + d.Facing.Name() + "。")
}

// DungeonTurnAround 轉身 180 度。**在門口也做得到** —— 原版的
// 「Not in doorway!」只擋左右轉,轉身走的是 default case。
func (s *State) DungeonTurnAround() {
	d := s.Dungeon
	if d == nil {
		return
	}
	d.Facing = d.Facing.TurnLeft().TurnLeft()
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
	s.dungeonStep(d.X+dx, d.Y+dy, back)
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
	s.dungeonStep(d.X+dx, d.Y+dy, false)
}

// dungeonStep 走到 (nx, ny)。
func (s *State) dungeonStep(nx, ny int, back bool) {
	d := s.Dungeon
	// ⚠ 原版的地牢座標是**繞的**,不是撞牆(`sub_48F4` 的 `if (v<0) v=7`)。
	// 引擎原本在邊界印「Blocked!」—— 那讓 8×8 的地圖變成有牆的盒子,
	// 而原版是環面。
	nx, ny = u5data.DungeonWrap(nx), u5data.DungeonWrap(ny)
	tile := s.DungeonTileAt(nx, ny)
	// 電擊力場:走進去 → 受傷 → 被彈回原格(`sub_4834`)。
	// ⚠ 這一段在**移動**裡,不在踩踏分派表裡 —— 玩家從來沒真的站上去過。
	if tile == u5data.DungeonElectricA || tile == u5data.DungeonElectricB {
		s.Log("好痛!")
		s.Log("電擊力場!")
		s.damageWholeParty()
		s.AdvanceTime(MinutesPerTurn)
		return
	}
	if u5data.DungeonPlayerBlocks(tile, back) {
		s.Log(MsgBlocked)
		return
	}
	d.X, d.Y = nx, ny
	s.AdvanceTime(MinutesPerTurn)
	s.dungeonTurnEnd()
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
		s.fieldAffectsParty(u5data.StatusAsleep)
	case u5data.DungeonPoisonA, u5data.DungeonPoisonB:
		s.Log("中毒!")
		s.fieldAffectsParty(u5data.StatusPoisoned)
	case u5data.DungeonFireA, u5data.DungeonFireB:
		s.Log("烈焰!")
		s.damageWholeParty()
	}
	// ⚠ 0x83 / 0x8B(電擊力場)**刻意沒有 case** —— `jpt_52C7` 把它們送進
	// default。它的效果在 `dungeonStep` 裡:走進去就被彈回來,
	// 所以玩家從來沒真的站在上面過。
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
	s.spawnDungeonMonster()
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
	s.spawnDungeonMonster()
	s.dungeonTurnEnd()
}

// hasRope 回報身上有沒有**抓鉤**(原版 `byte_3DFBB`)。
//
// ⚠⚠ **更正兩件事**(`docs/re/68` 追記):
//
//	① 位移**早就對出來了** —— 它是 0x0209..0x020F 那段七個位元組的第一格
//	   (`SaveGrappleOffset = 0x0209`),而那一段兩端都已釘死。
//	   這裡的 `return false` 是**陳舊標記**:註解說的「位移還沒對出來」
//	   在 save.go 補完那一段之後就不成立了,只是沒回來改(`rulebook/63`)。
//	② 它不是「繩索」,是**抓鉤(Grapple)** —— `sub_188C4`(大地圖攀爬)
//	   的第一道閘門就是它:`cmp byte_3DFBB, 0; jnz; 印 "With what?"`。
//
// 函式名保留 `hasRope` 是為了不動地牢那邊的呼叫端;語意見 `Inventory.Grapple`。
func (s *State) hasRope() bool { return s.Inventory.Grapple != 0 }

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
	s.spawnDungeonMonster()
	s.dungeonTurnEnd()
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
	// ★ 地牢房間是全遊戲唯一「ESC 離不開」的戰場(`docs/re/73`)。
	s.beginRoomCombat(&s.DungeonRooms.Maps[idx], idx, CombatModeRoom)
}

// damageWholeParty 讓全隊受傷(原版 `sub_2A4D0`)。
//
// ⚠ **更正**:此前這裡寫「傷害值還沒逆出來,用 `random(1, 20)` 是估計值」——
// 其實 `sub_2A4D0` 整支只有九行,讀出來就是:
//
//	for (i = 0; i < 6 && i < 隊伍人數; i++)
//	    if (狀態[i] != 'D') sub_2A464(i, sub_28E14(1, 8))
//
// 沒讀到的原因與許願井同一個:反編譯版的參數列是空的(`docs/re/66`)。
// 所以傷害是 **rand(1, 8)**,不是 1..20;而且**最多只掃 6 個人**
// (原版 `cmp edi, 6`;`CombatPartySlots` 就是那個 6)。
//
// `sub_2A464` 就是 `damageMember`(見 `pickchar.go`),所以這裡直接用它,
// 不再自己重寫扣血與判死。
func (s *State) damageWholeParty() {
	for i := 0; i < s.PartySize && i < len(s.Roster) && i < u5data.CombatPartySlots; i++ {
		if s.Roster[i].Status == u5data.StatusDead {
			continue
		}
		s.damageMember(i, s.Roll(1, u5data.DrownDamageMax))
	}
}

// fieldAffectsParty 是睡眠力場與毒力場對全隊的判定
//(原版 `sub_4DC8` / `sub_4E58` —— 兩支只差印出來的字與掛上的狀態)。
//
//	每個隊員各擲一次 random(1, 30)
//	敏捷 > 擲出來的點數 → 躲掉
//	已經死掉的跳過(`cmp byte_3DDBF[edx], 'D'`)
//
// ⚠ **不是「全隊一律中」**。我第一版寫成無條件掛狀態,那讓高敏捷完全沒有
// 意義 —— 而敏捷正是原版在這裡唯一看的東西。
func (s *State) fieldAffectsParty(status byte) {
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		ch := &s.Roster[i]
		if ch.Status == u5data.StatusDead {
			continue
		}
		if int(ch.Dex) > s.Roll(1, 30) {
			continue
		}
		ch.Status = status
		ch.Raw[u5data.CharStatus] = status
		s.Log(s.charName(ch) + "受到了影響。")
	}
}
