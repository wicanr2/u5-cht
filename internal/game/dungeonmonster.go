package game

import "github.com/wicanr2/u5-cht/internal/u5data"

// 地牢的遊蕩怪物(原版 `sub_5150` 的前半 + `sub_4C6C` + `sub_5008` + `sub_4460`)
//
// 地牢裡每一層養著**一隻**遊蕩怪物,你走一步它走一步;走到你腳下就開打。
// 這是 U5 地牢的壓力來源 —— 少了它,地牢只是一張要繞的迷宮。
//
// 每一步的順序照原版 `sub_5150`,**不能重排**:
//
//	1. 睡著的隊員擲醒不醒(`random(0,63) < 4`)
//	2. 遊蕩怪物走一步;走到隊伍身上 → 「遭到襲擊!」→ 開打 → 換一隻新的
//	3. 才處理腳下那一格(陷阱 / 房間)
//
// 把 2 和 3 對調的話,你會先中陷阱再被襲擊,而原版是先被襲擊 ——
// 而且原版在打完之後**重讀腳下的格子**再跑 3,因為戰鬥可能把人挪走了。

// DungeonMonster 是某一層的那隻遊蕩怪物(原版物件槽 1,`dword_3E474`)。
type DungeonMonster struct {
	// Kind 是八種裡的第幾種(0..7),同時就是 `MONn.16` 的檔號。
	Kind int
	// Creature 是生物編號(= 生物索引 × 4 + 64)。
	Creature byte
	// X, Y 是它在這一層的座標。
	X, Y int
	// PrevX, PrevY 是它上一格(原版物件槽 2,`dword_3E47C`)。
	//
	// 兩個用途:不讓它原地來回踱步,以及**算出「從哪邊被攻擊」** ——
	// 撞上隊伍之後它會退回這一格,所以方位是用退回後的位置算的。
	PrevX, PrevY int
}

// 睡著的隊員每走一步的甦醒判定(原版 `sub_5150` 開頭:`random(0,63) < 4`)。
const (
	dungeonWakeRollMax = 63
	dungeonWakeUnder   = 4
)

// dungeonGroupAlwaysFull 是「不擲骰,直接出滿」的兩個群體上限
//(原版 `cmp al, 8` / `cmp al, 10h`)。
const (
	dungeonGroupFull8  = 8
	dungeonGroupFull16 = 16
)

// spawnDungeonMonster 生一隻新的遊蕩怪物(原版 `sub_10208(1)` + `sub_4460(1)`)。
//
// 換層、掉進陷阱坑、以及**打完一場之後**都會換一隻。
func (s *State) spawnDungeonMonster() {
	d := s.Dungeon
	if d == nil || s.Dungeons == nil {
		return
	}
	d.Monster = nil
	// ⚠ 種類只抽一次,**在找落點之前** —— 落點找不到就這一層沒有怪,
	// 不會換一種再試(原版 `sub_4460` 先抽 r 才呼叫 `sub_4594`)。
	k := s.Roll(0, u5data.DungeonMonsterKinds-1)
	for try := 0; try < u5data.DungeonSpawnTries; try++ {
		cell := s.Roll(0, u5data.DungeonLevelB-1)
		x, y := cell%u5data.DungeonSide, cell/u5data.DungeonSide
		if !u5data.DungeonSpawnAllows(s.Dungeons.At(d.Index, d.Level, x, y)) {
			continue
		}
		// ⚠ 排除的是**同一行或同一列**,不只是同一格。所以剛換層時
		// 怪物一定不在你的正前方或正側方 —— 第一步不會馬上被撲。
		if x == d.X || y == d.Y {
			continue
		}
		d.Monster = &DungeonMonster{
			Kind: k, Creature: u5data.DungeonMonsterCreature(k),
			X: x, Y: y, PrevX: x, PrevY: y,
		}
		return
	}
}

// dungeonTurnEnd 是地牢裡「一個回合結束」要跑的三件事。
func (s *State) dungeonTurnEnd() {
	s.wakeDungeonSleepers()
	if s.moveDungeonMonster() {
		s.dungeonMonsterAttacks()
	}
	// 原版打完之後重讀腳下的格子才跑分派 —— `onDungeonTile` 本來就是現讀,
	// 所以這裡直接叫它就等於照抄。
	s.onDungeonTile()
}

// wakeDungeonSleepers 讓睡著的隊員有機會自己醒過來。
//
// 原版每走一步、每個睡著的隊員各擲一次 `random(0,63)`,小於 4 就醒 ——
// 大約十六步醒一個。⚠ 只有**睡著**('S')會醒;中毒與死亡不在此列。
func (s *State) wakeDungeonSleepers() {
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		ch := &s.Roster[i]
		if ch.Status != u5data.StatusAsleep {
			continue
		}
		if s.Roll(0, dungeonWakeRollMax) >= dungeonWakeUnder {
			continue
		}
		ch.Status = u5data.StatusGood
		ch.Raw[u5data.CharStatus] = u5data.StatusGood
		s.Log(s.charName(ch) + "醒了過來。")
	}
}

// moveDungeonMonster 讓遊蕩怪物走一步,回報有沒有撲到隊伍(原版 `sub_4C6C`)。
func (s *State) moveDungeonMonster() bool {
	d := s.Dungeon
	if d == nil || d.Monster == nil || s.Dungeons == nil {
		return false
	}
	m := d.Monster
	// 收割者紮了根,不移動 —— 只可能是**你走到它身上**。
	if u5data.DungeonMonsterCreatureIndex(m.Kind) != u5data.DungeonMonsterStill {
		s.stepDungeonMonster(m)
	}
	if m.X != d.X || m.Y != d.Y {
		return false
	}
	// 撲到了就退回上一格 —— 原版不讓怪物站在隊伍那一格上。
	m.X, m.Y = m.PrevX, m.PrevY
	return true
}

// stepDungeonMonster 試著走一步(最多八次)。
func (s *State) stepDungeonMonster(m *DungeonMonster) {
	d := s.Dungeon
	for try := 0; try < u5data.DungeonMoveTries; try++ {
		dx, dy := Direction(s.Roll(0, 3)).Delta()
		nx := u5data.DungeonWrap(m.X + dx)
		ny := u5data.DungeonWrap(m.Y + dy)
		if u5data.DungeonMonsterBlocked(s.Dungeons.At(d.Index, d.Level, nx, ny)) {
			continue
		}
		// 想走回上一格?八分之一才准。原版寫成 `random(0,7) − 1 != 0 → 重試`,
		// 所以准許的是**恰好擲出 1**,不是「擲出 0」——差一格就變成 1/8 vs 1/8
		// 的同機率,但寫錯就不是照抄了。
		if nx == m.PrevX && ny == m.PrevY && s.Roll(0, u5data.DungeonSpawnBackoff) != 1 {
			continue
		}
		m.PrevX, m.PrevY = m.X, m.Y
		m.X, m.Y = nx, ny
		return
	}
}

// dungeonMonsterAttacks 是「遭到襲擊!」那一段(原版 `sub_5008`)。
//
// ★ 兩個細節憑印象一定寫不出來:
//
//  1. **方位只在你沒面對它時才印。** 從你正對著的方向撲上來,原版只說
//     「Attacked!」—— 因為你看得見它,不需要人家告訴你在哪邊。
//  2. **被襲擊會把隊伍轉過去。** 印完方位之後 `byte_3EE15 = byte_3EE14`,
//     也就是朝向被改成攻擊來向。下一步的「前進」方向因此變了。
func (s *State) dungeonMonsterAttacks() {
	d := s.Dungeon
	m := d.Monster
	if m == nil {
		return
	}
	dir := Direction(u5data.DungeonAttackDirection(m.X, m.Y, d.X, d.Y))
	if dir == d.Facing {
		s.Log("遭到襲擊!")
	} else {
		s.Log("自" + dir.Name() + "方遭到襲擊!")
		d.Facing = dir
	}
	s.beginDungeonCombat(m)
	// 這一隻用掉了,換一隻新的守著這一層。
	s.spawnDungeonMonster()
}

// beginDungeonCombat 開一場地牢遭遇戰(原版 `sub_2E364(2, …)` → `sub_FE48`)。
//
// 戰場是**當場畫的**,不是從 `DUNGEON.CBT` 讀的(見 `u5data.BuildDungeonArena`)。
func (s *State) beginDungeonCombat(m *DungeonMonster) bool {
	d := s.Dungeon
	if d == nil || s.Stats == nil {
		return false
	}
	arena := u5data.BuildDungeonArena(u5data.DungeonArena{
		Number: d.Location - u5data.DungeonLocationBase + 1,
		Here:   s.DungeonTileHere(),
		Around: s.dungeonNeighbours(),
		Facing: int(d.Facing),
	})
	idx := u5data.DungeonMonsterCreatureIndex(m.Kind)
	if idx < 0 || idx >= u5data.CreatureCount {
		return false
	}
	for _, slot := range s.dungeonEnemySlots(idx) {
		arena.EnemyKind[slot] = u5data.CreatureBase + byte(idx)*4
	}
	return s.beginRoomCombat(arena, -1)
}

// dungeonEnemySlots 決定這一群怪物佔哪幾個入場點(原版 `sub_FE48` 尾段)。
//
// 原版先把 0..15 洗牌,再把前 n 個填上怪物 —— 所以**同一種怪物每次的站位
// 都不一樣**。隻數 n = `random(1, 群體上限)`,但上限剛好是 8 或 16 時直接出滿。
func (s *State) dungeonEnemySlots(creature int) []int {
	slots := make([]int, u5data.CombatEnemySlots)
	for i := range slots {
		slots[i] = i
	}
	for i := range slots {
		j := s.Roll(0, u5data.CombatEnemySlots-1)
		slots[i], slots[j] = slots[j], slots[i]
	}
	n := int(s.Stats.Creature[creature].GroupMax)
	switch {
	case n < 1:
		n = 1
	case n != dungeonGroupFull8 && n != dungeonGroupFull16:
		n = s.Roll(1, n)
	}
	return slots[:n]
}

// dungeonNeighbours 取四個方位的鄰格(順序 = 北 / 東 / 南 / 西,與方向碼同)。
func (s *State) dungeonNeighbours() [4]byte {
	d := s.Dungeon
	var out [4]byte
	if d == nil {
		return out
	}
	for i := 0; i < 4; i++ {
		dx, dy := Direction(i).Delta()
		out[i] = s.DungeonTileAt(u5data.DungeonWrap(d.X+dx), u5data.DungeonWrap(d.Y+dy))
	}
	return out
}
