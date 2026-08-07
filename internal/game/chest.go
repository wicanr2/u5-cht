package game

import (
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 寶箱、力場、鎖著的門
//
// 三樣共用「看腳下,不是的話看面前那一格」這個模式(原版 `sub_18D18`
// 與 `sub_1D1B8` 都是同一段:先查自己站的格子,高四位元不對才用面向
// 的增量表 `word_41970` / `word_41978` 去看前面那一格,座標 `& 7` 環繞)。

// dungeonFacingTile 回傳「腳下或面前」那一格的座標。
//
// wantKind 是在找哪一種(0x40 寶箱 / 0x80 力場);腳下就是的話直接用腳下。
func (s *State) dungeonFacingTile(wantKind byte) (x, y int, tile byte, ok bool) {
	d := s.Dungeon
	if d == nil {
		return 0, 0, 0, false
	}
	if u5data.DungeonKind(s.DungeonTileHere()) == wantKind {
		return d.X, d.Y, s.DungeonTileHere(), true
	}
	dx, dy := d.Facing.Delta()
	// ⚠ 原版是 `& 7` 環繞,不是夾住 —— 站在邊上面向外會繞到另一邊。
	nx, ny := (d.X+dx)&7, (d.Y+dy)&7
	t := s.DungeonTileAt(nx, ny)
	return nx, ny, t, u5data.DungeonKind(t) == wantKind
}

// OpenChest 是 Open 指令(原版按鍵 O)。
//
// 地牢裡看腳下 / 面前那一格是不是 0x4x;地表看腳下有沒有種類 1 的物件。
func (s *State) OpenChest() {
	if s.InDungeon() {
		x, y, tile, ok := s.dungeonFacingTile(u5data.DungeonChest)
		if !ok {
			s.Log("沒有東西可以打開。")
			return
		}
		s.openDungeonChest(x, y, tile)
		return
	}
	s.Log("此處沒有寶箱。")
}

// openDungeonChest 打開地牢寶箱。
//
// 順序照 `sub_15108` + `sub_18D18`:
//
//	有陷阱(bit 0)→ 印「Trapped!」,開箱的人受傷
//	格子換成 `(tile & 8) | 0x70` —— 開過的寶箱
//	依三張表擲獎品,生成在那一格
// ★ **「地牢寶箱的等級從哪來」這個懸案結掉了 —— 答案是沒有那個東西。**
//
// 我原本假設地牢寶箱與地表寶箱共用 `sub_15020` 的獎品表,只差找不到「等級」
// 從哪來,於是拿「深度 × 4」硬湊。實際上**它們是兩套完全不同的程式碼**:
// 地牢寶箱走 `sub_15930`,只擲 `random(1, 樓層×4 + 4)` 再比七個門檻,
// 根本沒有等級這個概念。
//
// ⚠ 而且**開箱與取物是兩個步驟**:
//
//	Open(O)  0x4x → 0x7x,該中的陷阱在這裡中
//	Get(G)   0x7x → 清空,獎品在這裡擲(見 get.go 的 getDungeonChest)
//
// 原版 `sub_15930` 對 0x4x 印的正是「得先打開它!」。引擎先前把兩步併成一步,
// 所以玩家開完箱就自動拿到東西 —— 少了一個指令,而且獎品表是錯的那一張。
func (s *State) openDungeonChest(x, y int, tile byte) {
	d := s.Dungeon
	if tile&u5data.ChestTrappedDungeon != 0 {
		s.Log("有陷阱!")
		s.chestTrapVictim()
	}
	s.Dungeons.Set(d.Index, d.Level, x, y, u5data.DungeonOpenedChest(tile))
	s.Log("箱子打開了!")
}

// chestTrapVictim 讓開箱的人吃陷阱(原版 `sub_2AB38`)。
//
// ⚠ 傷害值還沒逆完 —— 這裡用 `random(1, 20)`,與地牢陷阱同量級,
// 程式碼與文件都標明是**估計值**。
func (s *State) chestTrapVictim() {
	if s.PartySize <= 0 {
		return
	}
	i := s.Roll(0, s.PartySize-1)
	ch := &s.Roster[i]
	if ch.Status == u5data.StatusDead {
		return
	}
	dmg := s.Roll(1, 20)
	hp := int(ch.HP) - dmg
	if hp <= 0 {
		hp = 0
		ch.Status = u5data.StatusDead
		s.Log(s.charName(ch) + "倒下了!")
	}
	ch.HP = uint16(hp)
}

// rollChestContents 依三張表擲獎品。回傳有沒有掉出東西。
//
// 原版 `sub_15020` 由 i = 7 往 0 掃,每一項各擲一次:
//
//	等級 >= 門檻 且 random(1,30) >= 門檻
//
// **門檻同時當「等級下限」與「擲骰難度」** —— 門檻 3 的那一項幾乎必中,
// 門檻 25 的那一項要高等寶箱加上好運。
func (s *State) rollChestContents(level int) bool {
	got := false
	for i := len(u5data.ChestKind) - 1; i >= 0; i-- {
		th := int(u5data.ChestThreshold[i])
		if level < th || s.Roll(1, 30) < th {
			continue
		}
		max := int(u5data.ChestMax[i])
		n := 1
		if max != 1 {
			n = s.Roll(1, max)
		}
		kind := u5data.ChestKind[i]
		if kind == u5data.ChestKindGold {
			// 金幣是唯一語意確定的那一種:`random(1, 等級×3)`(`sub_14F68`)。
			gold := s.Roll(1, level*3)
			s.Inventory.Gold = addCap(s.Inventory.Gold, gold, 9999)
			s.Log("找到 " + itoa(gold) + " 枚金幣。")
			got = true
			continue
		}
		// 其餘七種原版是**在那一格生一個物件**讓玩家自己撿(`sub_2B6C8`)。
		// 種類碼的語意還沒對出來,所以照抄「生成」這個行為而不去猜它是什麼。
		s.Log("找到了什麼東西(物件種類 " + itoa(int(kind)) + " ×" + itoa(n) + ")。")
		got = true
	}
	return got
}

// DestroyField 是 An Grav:破除面前的力場(原版 `sub_1D1B8`)。
func (s *State) destroyField() bool {
	if !s.InDungeon() {
		s.Log("此處沒有力場。")
		return false
	}
	x, y, tile, ok := s.dungeonFacingTile(u5data.DungeonMagic)
	if !ok {
		s.Log("此處沒有力場。")
		return false
	}
	// `and byte ptr [ebx], 8` —— 只留「頭上有洞」那一位元。
	s.Dungeons.Set(s.Dungeon.Index, s.Dungeon.Level, x, y, tile&u5data.DungeonHoleAbove)
	s.Log("力場被破除了!")
	return true
}

// disarmOrUnlock 是 An Sanct 的完整行為(原版 `sub_18D18`)。
//
// 地牢裡:面前是寶箱就解陷阱並打開。
// 地表 / 城鎮:面前是鎖著的門(0xB9 / 0xBB)就 `tile − 1` 解鎖;
// 否則找腳下那一格的寶箱物件,清掉 +5 的最高位元(解陷阱)。
func (s *State) disarmOrUnlock() bool {
	if s.InDungeon() {
		x, y, tile, ok := s.dungeonFacingTile(u5data.DungeonChest)
		if !ok {
			return false
		}
		if tile&u5data.ChestTrappedDungeon != 0 {
			s.Log("陷阱解除了!")
		}
		s.openDungeonChest(x, y, tile&^u5data.ChestTrappedDungeon)
		return true
	}
	return s.UnlockAhead(s.spellFacing())
}

// spellFacing 是地表施法要用的方向。
//
// 原版每次都問一次;引擎的方向選單還沒接,先取「四鄰裡有鎖著的門的那一邊」,
// 沒有就朝北。**這是介面的近似,不是規則。**
func (s *State) spellFacing() Direction {
	for _, d := range []Direction{North, East, South, West} {
		dx, dy := d.Delta()
		if u5data.TileIsLockedDoor(s.TileAt(s.X+dx, s.Y+dy)) {
			return d
		}
	}
	return North
}

// UnlockAhead 是 **An Sanct** 解一般的鎖:`tile − 1`(0xB9 → 0xB8、0xBB → 0xBA)。
func (s *State) UnlockAhead(dir Direction) bool {
	dx, dy := dir.Delta()
	x, y := s.X+dx, s.Y+dy
	t := s.TileAt(x, y)
	if !u5data.TileIsLockedDoor(t) {
		s.Log("那裡沒有鎖著的東西。")
		return false
	}
	if !s.SetTileAt(x, y, t-1) {
		s.Log("開不了。")
		return false
	}
	s.Log("鎖開了。")
	return true
}

// MagicUnlockAhead 是 **In Ex Por**:解魔法鎖(0x97 → 0xB8、0x98 → 0xBA)。
//
// ⚠ 與 An Sanct **不是同一條規則**。我第一版把兩個都寫成 `tile − 1`,
// 那會讓 In Ex Por 把 0x97 變成 0x96(一個完全不相干的圖)。
func (s *State) MagicUnlockAhead(dir Direction) bool {
	dx, dy := dir.Delta()
	x, y := s.X+dx, s.Y+dy
	next := u5data.MagicUnlock(s.TileAt(x, y))
	if next == 0 {
		s.Log("那裡沒有魔法的封印。")
		return false
	}
	if !s.SetTileAt(x, y, next) {
		return false
	}
	s.Log("封印解開了。")
	return true
}

// MagicLockAhead 是 **An Ex Por**:魔法上鎖(0xB8/0xB9 → 0x97、0xBA/0xBB → 0x98)。
func (s *State) MagicLockAhead(dir Direction) bool {
	dx, dy := dir.Delta()
	x, y := s.X+dx, s.Y+dy
	next := u5data.MagicLock(s.TileAt(x, y))
	if next == 0 {
		s.Log("那裡沒有門。")
		return false
	}
	if !s.SetTileAt(x, y, next) {
		return false
	}
	s.Log("門被魔法封住了。")
	return true
}

// DispelAhead 是 **An Ylem**:抹掉某個方向的力場 / 能量(原版 `sub_18C00`)。
func (s *State) DispelAhead(dir Direction) bool {
	dx, dy := dir.Delta()
	x, y := s.X+dx, s.Y+dy
	if !u5data.AnYlemTiles[s.TileAt(x, y)] {
		s.Log("此處沒有可以抹去的東西。")
		return false
	}
	if !s.SetTileAt(x, y, u5data.TileBrickFloor) {
		return false
	}
	s.Log("噗!")
	return true
}
