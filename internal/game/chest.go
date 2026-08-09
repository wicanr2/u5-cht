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
	// ★ 地牢那條**不問方向、也不看面前那一格** —— `sub_152B8` 讀的是
	// `byte_3E0A6/A7`,玩家自己站的位置。引擎原本走 `dungeonFacingTile`
	// (腳下或面前),而那條規則的出處是 `sub_18D18`(**施法**選目標用的)。
	// 兩支的行為不同,而把施法的規則套到 Open 上會讓玩家隔一格就開得到箱子。
	if s.InDungeon() {
		s.openDungeonChestUnderfoot()
		return
	}
	// 地表 / 場景:問方向,再看那一格(原版 `sub_15374` 的 `sub_2B2AC`)。
	s.AskDirection(func(d Direction) { s.openToward(d) })
}

// openDungeonChestUnderfoot 是地牢的 Open(原版 `sub_152B8`)。
//
// ⚠ 陷阱的判準是 **`tile & 7`(低三位元全部)**,不是只有 bit 0。
// `sub_18D18`(An Sanct)用的才是單一位元 —— 又是「同一件事兩支各做一半」
// (`docs/re/74` §1 的形狀)。低三位元有值就中陷阱,照原樣實作。
func (s *State) openDungeonChestUnderfoot() {
	d := s.Dungeon
	tile := s.DungeonTileHere()
	switch {
	case u5data.DungeonKind(tile) == u5data.DungeonChest:
		s.pickMember("", func(who int) {
			if who < 0 {
				return
			}
			if tile&u5data.DungeonChestTrapMask != 0 {
				s.Log(MsgTrapped)
				s.chestTrapVictim()
			}
			s.Dungeons.Set(d.Index, d.Level, d.X, d.Y, u5data.DungeonOpenedChest(tile))
			s.Log(MsgChestOpened)
		})
	case u5data.DungeonKind(tile) == u5data.DungeonOpenedChestKind:
		s.Log(MsgAlreadyOpen)
	default:
		s.Log(MsgWhat)
	}
}

// openToward 是地表 / 場景的 Open(原版 `sub_15374` 問完方向之後那一段)。
func (s *State) openToward(d Direction) {
	// ★ 開新的門之前先把上一扇關掉 —— 原版在問方向**之前**就呼叫一次
	// `sub_2B64C(byte_3E161, …)`,而那四個變數只有一組。
	s.closePendingDoor()
	dx, dy := d.Delta()
	x, y := s.X+dx, s.Y+dy
	tile := s.TileAt(x, y)
	switch u5data.OpenActionFor(tile) {
	case u5data.OpenAlreadyOpen:
		s.Log(MsgItsOpen)
	case u5data.OpenTooHeavy:
		s.Log(MsgTooHeavy)
	case u5data.OpenLocked:
		s.Log(MsgLocked)
	case u5data.OpenDoor:
		// 那一格變成磚地,並排定 4 回合後把原本的 tile 寫回去。
		s.SetTileAt(x, y, u5data.OpenedDoorTile)
		s.door = pendingDoor{Tile: tile, X: x, Y: y, Turns: u5data.DoorAutoCloseTurns}
		s.Log(MsgOpened)
	default:
		s.openObjectChest(x, y)
	}
}

// openObjectChest 開那一格**物件層**上的箱子(原版 `sub_15108`)。
//
// ★ 四條容易漏的規則:
//
//  1. **從槽 1 開始掃**(槽 0 是隊伍自己)。
//  2. **檀香木盒(種類 0x0E)打不開** —— 印 "Can't!",不是「沒東西可開」。
//  3. **品質這一個位元組裝兩件事**:最高位 = 有陷阱、低七位 = 獎品等級。
//     原版 `and var_5, 7Fh` 把陷阱位清掉之後才拿去擲獎品。
//  4. ★★ **在場景裡開箱子扣 2 點業報**(下限 0)—— 只有地點 1..0x20 扣。
//     那是「翻別人家的箱子」的代價;大地圖上的箱子無主,不扣。
func (s *State) openObjectChest(x, y int) {
	objs := s.currentObjects()
	if objs == nil {
		s.Log(MsgNothingToOpen)
		return
	}
	slot := -1
	for i := 1; i < len(objs.Objects); i++ {
		o := &objs.Objects[i]
		if !o.Present() || o.X != WrapWorld(x) || o.Y != WrapWorld(y) {
			continue
		}
		if !s.InCombat() && o.Floor != s.Floor {
			continue
		}
		if o.Kind == u5data.ObjSandalwoodBox {
			s.Log(MsgCantOpenThat)
			return
		}
		if o.Kind == u5data.ObjLockedChest {
			slot = i
			break
		}
	}
	if slot < 0 {
		s.Log(MsgNothingToOpen)
		return
	}
	// ⚠ `slot` 由回呼捕捉。選單開著的時候不會過回合 ⇒ 物件表不會變動,
	// 所以捕捉索引是安全的(換成捕捉物件本身反而會漏掉 `Remove` 的副作用)。
	s.pickMember("", func(who int) {
		if who >= 0 {
			s.openChestSlot(objs, slot)
		}
	})
}

// openChestSlot 真的把第 slot 個箱子打開。
func (s *State) openChestSlot(objs *u5data.ObjectSet, slot int) {
	quality := int(objs.Objects[slot].Raw[u5data.ObjQuality])
	objs.Remove(slot)
	// ★ 業報 −2,只在場景裡。
	if s.Location >= 1 && s.Location <= u5data.LastSceneLocation {
		s.Karma = subFloor(s.Karma, u5data.ChestOpenKarmaPenalty)
	}
	if quality&u5data.ChestTrapQualityBit != 0 {
		quality &^= u5data.ChestTrapQualityBit
		s.Log(MsgTrapped)
		s.chestTrapVictim()
	}
	if !s.rollChestContents(quality) {
		s.Log(MsgChestEmpty)
	}
}

// pendingDoor 是「打開著、等著自己關上」的那一扇門(原版 `byte_3E161..164`)。
//
// ⚠ **只有一組**,所以同時只能有一扇門是開的。
type pendingDoor struct {
	Tile  byte
	X, Y  int
	Turns int
}

// tickDoor 是主迴圈每回合對那一扇門做的事(原版 `sub_1A54` 的
// `dec byte_3E164; jnz` 那一段)。
func (s *State) tickDoor() {
	if s.door.Tile == 0 || s.door.Turns <= 0 {
		return
	}
	s.door.Turns--
	if s.door.Turns == 0 {
		s.closePendingDoor()
	}
}

// closePendingDoor 把那一格寫回原本的 tile(原版 `sub_2B64C`)。
//
// ⚠ 原版**只在場景裡**才真的寫回(`地點 != 0 && 地點 < 0x21`)——
// 大地圖與地牢沒有門要關。
func (s *State) closePendingDoor() {
	d := s.door
	s.door = pendingDoor{}
	if d.Tile == 0 {
		return
	}
	if s.Location == 0 || s.Location > u5data.LastSceneLocation {
		return
	}
	s.SetTileAt(d.X, d.Y, d.Tile)
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
// ✅ **整支讀完了**(`docs/re/91`),此前的 `random(1,20)` 是估計值,已換掉。
//
//	if (地點碼 > 0x7F)            種類 = random(0, 1)      ; ★ 戰鬥中只有酸與毒
//	else                          種類 = byte_5FFEC[random(0,7)]
//	switch (種類) {
//	  0 "Acid"   → sub_2A464(隊員, sub_2B724())            ; 單人 1..30
//	  1 "Poison" → sub_2AB08(隊員)                         ; 單人中毒
//	  2 "Bomb"   → sub_2A4D0()                             ; 全隊各 random(1,8)
//	  3 "Gas"    → for (i=0..5) sub_2AB08(i)               ; 全隊中毒
//	  default    → 什麼都不做
//	}
//
// ⚠ `default` 走不到(`byte_5FFEC` 只有 0..3),但原版留著 —— 照原樣。
func (s *State) chestTrapVictim() {
	if s.PartySize <= 0 {
		return
	}
	// ★ 原版第一件事是 `sub_2C598(0x28, 0xBB8, 0x1F4)` —— 那三個數字是
	// **白噪參數(Rate/Dura/Limit)不是音效索引**,而 FM Towns 把這一組
	// 轉成 `sub_2C46C(7, 0x3C)` = `DAME1.SND`(`docs/re/92`)。
	s.PlaySFX(u5data.SFXDamage1)
	victim := s.Roll(0, s.PartySize-1)
	switch s.trapKind() {
	case u5data.TrapAcid:
		s.Log(MsgTrapAcid)
		// ★ 與命中骰、Mani 同一顆:`max(1, random(0,60)/2)` → 1..30
		// (`docs/re/15` 已記,`sub_2B724`)。
		s.damageMember(victim, s.AttackRoll())
	case u5data.TrapPoison:
		s.Log(MsgTrapPoison)
		s.poisonMember(victim)
	case u5data.TrapBomb:
		s.Log(MsgTrapBomb)
		s.bombEveryone()
	case u5data.TrapGas:
		s.Log(MsgTrapGas)
		for i := 0; i < u5data.CombatPartySlots; i++ {
			s.poisonMember(i)
		}
	}
}

// bombEveryone 是炸彈陷阱的全隊傷害(原版 `sub_2A4D0`)。
//
// ⚠ 原版的迴圈上限是**寫死的 6**(不是隊伍人數),裡面才逐一檢查
// 「槽 < byte_3E06B」與「狀態不是 'D'」。兩層檢查照抄:行為等價,
// 但寫成 `for i < PartySize` 會把原版的結構抹掉。
func (s *State) bombEveryone() {
	for i := 0; i < u5data.CombatPartySlots; i++ {
		if i >= s.PartySize || i >= len(s.Roster) {
			continue
		}
		if s.Roster[i].Status == u5data.StatusDead {
			continue
		}
		s.damageMember(i, s.Roll(1, u5data.TrapBombDamageMax))
	}
}

// trapKind 擲出陷阱種類(原版 `sub_2AB38` 的前半)。
//
// ★ **戰鬥中只有酸與毒** —— 地點碼 > 0x7F 時不查那張八筆權重表,
// 只擲 `random(0,1)`。引擎的戰鬥地點碼是 0xFF(`locationCode`)。
func (s *State) trapKind() int {
	if s.locationCode() > 0x7F {
		return s.Roll(0, u5data.TrapCombatKindMax)
	}
	return int(u5data.TrapKindRoll[s.Roll(0, u5data.TrapKindRollMax)])
}

// poisonMember 把一個隊員設成中毒(原版 `sub_2AB08`)。
//
//	槽 >= 隊伍人數 → 不動;已經死掉 → 不動;否則狀態設 'P'
//
// ⚠ **死人不會中毒**,而且**不會把中毒的人治好** —— 它只寫入 'P'。
func (s *State) poisonMember(i int) {
	if i < 0 || i >= s.PartySize || i >= len(s.Roster) {
		return
	}
	if s.Roster[i].Status == u5data.StatusDead {
		return
	}
	s.Roster[i].Status = u5data.StatusPoisoned
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
			// 金幣的數量另有算法:`random(1, 等級×3)`(`sub_14F68`),
			// 不走 `ChestMax` 那一格。
			gold := s.Roll(1, level*3)
			s.Inventory.Gold = addCap(s.Inventory.Gold, gold, 9999)
			s.Log("找到 " + itoa(gold) + " 枚金幣。")
			got = true
			continue
		}
		// ★ **種類碼的語意對出來了。** 這張表的八個值
		// `{1, 2, 3, 4, 7, 8, 13, 15}` 正是 `sub_154BC` 的物品碼:
		// 1 上鎖的箱、2 金幣、3 藥水、4 卷軸、7 鑰匙、8 寶石、13 火把、15 食物。
		// 原本這裡只印「找到了什麼東西(物件種類 N)」,因為當時還沒逆 `sub_154BC`。
		//
		// ⚠ 原版是**在那一格生一個物件**讓玩家自己 Get(`sub_2B6C8`);
		// 這裡直接收進背包,少了「掉在地上」那一步。差別在:原版可以把
		// 掉出來的東西留在原地不撿。標成已知差異,不假裝一樣。
		s.pickUp(kind, n, 0)
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
