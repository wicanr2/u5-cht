package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 進地牢就該有一隻遊蕩怪物,而且**不在隊伍的同一行或同一列**。
//
// 這條擋的是「生成只避開同一格」的寫法 —— 那樣一進門就可能被正前方撲上來,
// 而原版刻意讓第一步是安全的。
func TestDungeonSpawnsAMonsterOffTheParty(t *testing.T) {
	s := dungeonState(t)
	s.Location = 0
	e := u5data.DungeonEntrances[0]
	s.X, s.Y = e.X, e.Y
	s.Enter()
	if !s.InDungeon() {
		t.Fatalf("進不去:\n%s", s.log())
	}
	d := s.Dungeon
	if d.Monster == nil {
		t.Fatal("進地牢卻沒有遊蕩怪物")
	}
	if d.Monster.X == d.X || d.Monster.Y == d.Y {
		t.Errorf("怪物在 (%d,%d),與隊伍 (%d,%d) 同行或同列",
			d.Monster.X, d.Monster.Y, d.X, d.Y)
	}
	if d.Monster.Creature < u5data.CreatureBase {
		t.Errorf("怪物的生物編號是 %d,不像生物", d.Monster.Creature)
	}
}

// 遊蕩怪物走不進牆、房間與陷阱。
//
// 做法:讓它在一層裡走很多步,每一步都檢查落點合法 —— 這比檢查判定式本身
// 更接近玩家會遇到的情況(判定式對了但呼叫端忘了用,測試才抓得到)。
func TestDungeonMonsterNeverStepsIntoWalls(t *testing.T) {
	s := dungeonState(t)
	s.Location = 0
	e := u5data.DungeonEntrances[0]
	s.X, s.Y = e.X, e.Y
	s.Enter()
	d := s.Dungeon
	if d.Monster == nil {
		t.Fatal("沒有怪物可測")
	}
	moved := 0
	for i := 0; i < 400; i++ {
		m := d.Monster
		if m == nil {
			s.spawnDungeonMonster()
			continue
		}
		bx, by := m.X, m.Y
		if s.moveDungeonMonster() {
			// 撲到隊伍了 —— 這一步不算移動,重新放一隻繼續測。
			s.spawnDungeonMonster()
			continue
		}
		tile := s.Dungeons.At(d.Index, d.Level, m.X, m.Y)
		if u5data.DungeonMonsterBlocked(tile) {
			t.Fatalf("怪物走進了 0x%02X (%d,%d)", tile, m.X, m.Y)
		}
		if m.X != bx || m.Y != by {
			moved++
		}
	}
	if moved == 0 {
		t.Error("四百步下來怪物一步都沒動,移動邏輯沒有真的跑到")
	}
}

// 收割者紮了根,一步都不會走。
func TestReaperNeverMoves(t *testing.T) {
	s := dungeonState(t)
	s.Location = 0
	e := u5data.DungeonEntrances[0]
	s.X, s.Y = e.X, e.Y
	s.Enter()
	d := s.Dungeon
	// 手動換成第 7 種(收割者),放在一個離隊伍遠的空格。
	d.Monster = &DungeonMonster{
		Kind: 7, Creature: u5data.DungeonMonsterCreature(7),
		X: 5, Y: 5, PrevX: 5, PrevY: 5,
	}
	if d.X == 5 && d.Y == 5 {
		t.Skip("隊伍剛好站在測試點上")
	}
	for i := 0; i < 50; i++ {
		s.moveDungeonMonster()
		if d.Monster.X != 5 || d.Monster.Y != 5 {
			t.Fatalf("收割者移動到 (%d,%d) 了", d.Monster.X, d.Monster.Y)
		}
	}
}

// 被襲擊時:方位只在**沒面對它**時才報,而且報完會把隊伍轉過去。
func TestAttackReportsDirectionOnlyWhenNotFacingIt(t *testing.T) {
	s := dungeonState(t)
	s.Location = 0
	e := u5data.DungeonEntrances[0]
	s.X, s.Y = e.X, e.Y
	s.Enter()
	d := s.Dungeon
	d.X, d.Y = 4, 4
	d.Facing = South

	// 怪物從北邊(上一格 y = 3)撲上來。
	d.Monster = &DungeonMonster{
		Kind: 0, Creature: u5data.DungeonMonsterCreature(0),
		X: 4, Y: 3, PrevX: 4, PrevY: 3,
	}
	s.Messages = nil
	s.dungeonMonsterAttacks()
	if !strings.Contains(strings.Join(s.Messages, "\n"), "自北方") {
		t.Errorf("面朝南被北邊撲上,應報出方位:\n%s", s.log())
	}
	if d.Facing != North {
		t.Errorf("被襲擊後朝向是 %s,應轉為北", d.Facing.Name())
	}

	// 這次面朝北,同樣從北邊撲上來 —— 不該再報方位。
	d.X, d.Y = 4, 4
	d.Facing = North
	d.Monster = &DungeonMonster{
		Kind: 0, Creature: u5data.DungeonMonsterCreature(0),
		X: 4, Y: 3, PrevX: 4, PrevY: 3,
	}
	s.Messages = nil
	s.dungeonMonsterAttacks()
	joined := strings.Join(s.Messages, "\n")
	if strings.Contains(joined, "自北方") {
		t.Errorf("已經面朝北了還報方位:\n%s", s.log())
	}
	if !strings.Contains(joined, "遭到襲擊") {
		t.Errorf("沒有報出遭襲:\n%s", s.log())
	}
}

// 撲上來會真的開一場戰鬥,而戰場是當場畫的(不是 DUNGEON.CBT 的房間)。
func TestAttackStartsCombatOnAGeneratedArena(t *testing.T) {
	s := dungeonState(t)
	s.Location = 0
	e := u5data.DungeonEntrances[0]
	s.X, s.Y = e.X, e.Y
	s.Enter()
	d := s.Dungeon
	d.X, d.Y = 4, 4
	d.Facing = North
	d.Monster = &DungeonMonster{
		Kind: 0, Creature: u5data.DungeonMonsterCreature(0),
		X: 4, Y: 3, PrevX: 4, PrevY: 3,
	}
	s.dungeonMonsterAttacks()
	if !s.InCombat() {
		t.Fatalf("被撲上卻沒有開打:\n%s", s.log())
	}
	c := s.Combat
	if c.MapIndex != -1 {
		t.Errorf("戰場編號是 %d,現場畫的戰場不該有 .CBT 編號", c.MapIndex)
	}
	// 場上至少要有一隻敵人,而且是巨鼠。
	enemies, _ := s.sideCounts(c)
	if enemies < 1 {
		t.Errorf("戰場上沒有敵人")
	}
	if c.EnemyName == "" || strings.Contains(c.EnemyName, "房間") {
		t.Errorf("敵人名字是 %q,應是那隻遊蕩怪物", c.EnemyName)
	}
	// 打完之後這一層要換上一隻新的怪物。
	if d.Monster == nil {
		t.Error("打完之後沒有補上新的遊蕩怪物")
	}
}

// 睡著的隊員在地牢裡走著會自己醒過來,而中毒的不會被順手治好。
func TestDungeonStepsWakeSleepersOnly(t *testing.T) {
	s := dungeonState(t)
	s.Location = 0
	e := u5data.DungeonEntrances[0]
	s.X, s.Y = e.X, e.Y
	s.Enter()
	for i := 0; i < s.PartySize; i++ {
		s.Roster[i].Status = u5data.StatusAsleep
	}
	if s.PartySize < 2 {
		t.Skip("隊伍太小")
	}
	s.Roster[1].Status = u5data.StatusPoisoned

	for i := 0; i < 300; i++ {
		s.wakeDungeonSleepers()
	}
	if s.Roster[0].Status != u5data.StatusGood {
		t.Errorf("走了三百步第一位還在睡(狀態 %c)", s.Roster[0].Status)
	}
	if s.Roster[1].Status != u5data.StatusPoisoned {
		t.Errorf("中毒的被順手治好了(狀態 %c)", s.Roster[1].Status)
	}
	if s.Roster[0].Raw[u5data.CharStatus] != u5data.StatusGood {
		t.Error("醒過來只改了欄位、沒改存檔位元組")
	}
}

// 群體隻數:上限剛好 8 或 16 的怪物一次出滿,其餘擲骰。
func TestDungeonGroupSizeFollowsTheStatsTable(t *testing.T) {
	s := dungeonState(t)
	if s.Stats == nil {
		t.Skip("沒有戰鬥數值表")
	}
	for k := 0; k < u5data.DungeonMonsterKinds; k++ {
		idx := u5data.DungeonMonsterCreatureIndex(k)
		max := int(s.Stats.Creature[idx].GroupMax)
		seen := map[int]bool{}
		for i := 0; i < 200; i++ {
			seen[len(s.dungeonEnemySlots(idx))] = true
		}
		for n := range seen {
			if n < 1 || n > max && max >= 1 {
				t.Errorf("第 %d 種(上限 %d)出了 %d 隻", k, max, n)
			}
		}
		if max == 8 || max == 16 {
			if len(seen) != 1 || !seen[max] {
				t.Errorf("第 %d 種上限 %d,應該每次都出滿,實得 %v", k, max, seen)
			}
		}
	}
}

// 地牢的搜尋:方向是相對的,而且搜得出暗門。
func TestDungeonSearchFindsSecretDoors(t *testing.T) {
	s := dungeonState(t)
	s.Location = 0
	e := u5data.DungeonEntrances[0]
	s.X, s.Y = e.X, e.Y
	s.Enter()
	d := s.Dungeon
	d.X, d.Y = 4, 4
	d.Facing = North
	s.TorchTurns = 10 // 有光才搜得到

	// 正前方(北)放一道暗門,而且帶著「頭上有洞」那一位元。
	const holed = DungeonSecretDoor | u5data.DungeonHoleAbove
	s.Dungeons.Set(d.Index, d.Level, 4, 3, holed)

	s.Messages = nil
	s.Search()
	if !s.AwaitingDirection() {
		t.Fatalf("地牢的搜尋沒有問方向:\n%s", s.log())
	}
	s.AnswerDirection(North) // ↑ = 前方
	joined := strings.Join(s.Messages, "\n")
	if !strings.Contains(joined, "暗門") {
		t.Errorf("沒找到暗門:\n%s", s.log())
	}
	after := s.Dungeons.At(d.Index, d.Level, 4, 3)
	if u5data.DungeonKind(after) != u5data.DungeonDoorway {
		t.Errorf("暗門沒有變成門:0x%02X", after)
	}
	// ★ 「頭上有洞」那一位元不能被抹掉,否則之後爬不回上一層。
	if after&u5data.DungeonHoleAbove == 0 {
		t.Errorf("開門把「頭上有洞」抹掉了:0x%02X", after)
	}
}

// 沒有火把也沒有光明咒語時,搜尋只會回「一片漆黑」。
//
// ⚠ 原版判的是**兩個光源計時器**,不是視野半徑 —— 地牢的基礎半徑永遠 > 0,
// 拿半徑判會變成「永遠有光」。
func TestDungeonSearchNeedsLight(t *testing.T) {
	s := dungeonState(t)
	s.Location = 0
	e := u5data.DungeonEntrances[0]
	s.X, s.Y = e.X, e.Y
	s.Enter()
	d := s.Dungeon
	d.X, d.Y = 4, 4
	d.Facing = North
	s.TorchTurns, s.LightTurns = 0, 0
	s.Dungeons.Set(d.Index, d.Level, 4, 3, DungeonSecretDoor)

	s.Messages = nil
	s.Search()
	s.AnswerDirection(North)
	if !strings.Contains(strings.Join(s.Messages, "\n"), "漆黑") {
		t.Errorf("沒有光卻搜得到東西:\n%s", s.log())
	}
	if u5data.DungeonKind(s.Dungeons.At(d.Index, d.Level, 4, 3)) != DungeonSecretDoor {
		t.Error("摸黑竟然把暗門開了")
	}
}
