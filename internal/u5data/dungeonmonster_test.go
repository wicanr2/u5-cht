package u5data

import (
	"os"
	"testing"
)

// 八種遊蕩怪物的生物編號要落在生物名表上,而且名字要真的取得出來 ——
// 表偏一格的話至少有一筆會撈到空指標或人類角色。
func TestDungeonMonsterKindsAreRealCreatures(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	tbl, err := LoadCreatureTable(dir)
	if err != nil {
		t.Fatalf("讀不到 DATA.OVL:%v", err)
	}
	want := []string{
		"Giant Rat", "Bat", "Giant Spider", "Ghost",
		"Slime", "Gremlin", "Gazer", "Reaper",
	}
	for k := 0; k < DungeonMonsterKinds; k++ {
		got := tbl.Name(DungeonMonsterCreature(k))
		if got != want[k] {
			t.Errorf("MON%d 是 %q,預期 %q", k, got, want[k])
		}
	}
}

// 不會移動的那一種必須是收割者。
//
// 這條看起來像重言,其實是**表對齊的檢查**:原版豁免的是索引 0x1B,
// 而 0x1B 要剛好落在第 7 種身上。表滑一格,豁免的就會變成別種能走的怪物。
func TestOnlyTheReaperStandsStill(t *testing.T) {
	still := 0
	for k := 0; k < DungeonMonsterKinds; k++ {
		if DungeonMonsterCreatureIndex(k) == DungeonMonsterStill {
			still++
			if k != 7 {
				t.Errorf("不動的那一種是第 %d 種,預期第 7 種(MON7)", k)
			}
		}
	}
	if still != 1 {
		t.Errorf("有 %d 種不移動,預期恰好 1 種", still)
	}
}

// 生得出來的格子與走得進去的格子**不是同一組**。
//
// 0x90 是那個差異點:走得進去,卻生不出來。寫成一條規則就會少掉這個行為。
func TestSpawnAndMoveRulesDiffer(t *testing.T) {
	if DungeonMonsterBlocked(0x90) {
		t.Errorf("0x90 應該走得進去")
	}
	if DungeonSpawnAllows(0x90) {
		t.Errorf("0x90 不應該生得出來")
	}
	// 陷阱兩者都不行;門兩者都可以。
	if DungeonSpawnAllows(0x61) || !DungeonMonsterBlocked(0x61) {
		t.Errorf("陷阱(0x61)應該兩者都不行")
	}
	if !DungeonSpawnAllows(0x70) || DungeonMonsterBlocked(0x70) {
		t.Errorf("門(0x70)應該兩者都可以")
	}
	// 房間與牆走不進去。
	for _, tile := range []byte{DungeonRoomA, DungeonRoomF, DungeonWall} {
		if !DungeonMonsterBlocked(tile) {
			t.Errorf("0x%02X 應該走不進去", tile)
		}
	}
}

// 攻擊方位要用 8 格的環面算,而且四個方向都對得上。
func TestAttackDirectionWrapsAroundTheLevel(t *testing.T) {
	cases := []struct {
		mx, my, px, py int
		want           int
		why            string
	}{
		{4, 4, 3, 4, 1, "怪在東"},
		{4, 4, 5, 4, 3, "怪在西"},
		{4, 4, 4, 3, 2, "怪在南"},
		{4, 4, 4, 5, 0, "怪在北"},
		// 環面:怪在 x=0,隊伍在 x=7 → 怪其實在隊伍的東邊(繞過去)。
		{0, 4, 7, 4, 1, "跨邊界時怪在東"},
		{7, 4, 0, 4, 3, "跨邊界時怪在西"},
		{4, 0, 4, 7, 2, "跨邊界時怪在南"},
	}
	for _, c := range cases {
		if got := DungeonAttackDirection(c.mx, c.my, c.px, c.py); got != c.want {
			t.Errorf("%s:怪 (%d,%d) 隊伍 (%d,%d) 算出 %d,預期 %d",
				c.why, c.mx, c.my, c.px, c.py, got, c.want)
		}
	}
}
