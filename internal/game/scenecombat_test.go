package game

import (
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// TestNPCKillIsPermanentReadsTheBranchesRight:哪些 NPC 死了會被永久記下。
//
// ⚠ `sub_218` 的分支很容易讀反:`cmp esi,70h; jz loc_243` 是**跳去檢查 0xB4**,
// 而 0x70 ≠ 0xB4 所以就地返回 —— 也就是「**衛兵不記**」。
// 讀成「衛兵要記」的話,玩家把守衛殺光之後那座城就永遠沒有衛兵了。
func TestNPCKillIsPermanentReadsTheBranchesRight(t *testing.T) {
	cases := []struct {
		creature byte
		want     bool
		why      string
	}{
		{u5data.CreatureGuard, false, "衛兵會補上"},
		{u5data.CreatureGuard + 3, false, "衛兵的四個朝向都算衛兵(& 0xFC)"},
		{0x40, true, "一般居民"},
		{0x7C, true, "0x80 以下都記"},
		{0xB4, true, "0x80 以上唯一的例外"},
		{0xB7, true, "0xB4 的四個朝向"},
		{0x80, false, "0x80 以上、不是 0xB4"},
		{u5data.TileShadowlord, false, "暗影君主要用碎片才殺得死"},
	}
	for _, c := range cases {
		if got := u5data.NPCKillIsPermanent(c.creature); got != c.want {
			t.Errorf("生物 %02X:NPCKillIsPermanent = %v,預期 %v(%s)",
				c.creature, got, c.want, c.why)
		}
	}
}

// TestKillingACivilianIsPermanent:打死居民之後離場再回來,他不會復活。
//
// ⚠ 這條同時驗兩件事:`removed` 這次有效、`RemovedNPC` 存檔位元也設了。
// 只做前者的話,離場再進來人就回來了。
func TestKillingACivilianIsPermanent(t *testing.T) {
	s := inTown(t)
	if s == nil {
		return
	}
	victim := -1
	for _, v := range s.VisibleNPCs() {
		if u5data.NPCKillIsPermanent(v.NPC.Creature) {
			victim = v.Index
			break
		}
	}
	if victim < 0 {
		t.Skip("此刻城裡沒有會被永久記下的 NPC")
	}
	if !s.beginNPCCombat(victim) {
		t.Skipf("打不起來(缺戰鬥地圖?):\n%s", s.log())
	}
	if s.RemovedNPC[s.Location-1]&(1<<uint(victim)) == 0 {
		t.Errorf("存檔位元沒設 —— 離場再回來這個人會復活")
	}
	// 離開戰鬥、重新進場景。
	loc := s.Location
	s.Combat, s.Prompt = nil, PromptNone
	if err := s.SetScene(loc, 0, 15, 15); err != nil {
		t.Fatal(err)
	}
	for _, v := range s.VisibleNPCs() {
		if v.Index == victim {
			t.Error("重新進場景之後被打死的人又出現了")
		}
	}
}

// TestKillingAGuardIsNotPermanent:衛兵死了會補上。
func TestKillingAGuardIsNotPermanent(t *testing.T) {
	s := inTown(t)
	if s == nil {
		return
	}
	guard := -1
	for _, v := range s.VisibleNPCs() {
		if v.NPC.Creature&0xFC == u5data.CreatureGuard {
			guard = v.Index
			break
		}
	}
	if guard < 0 {
		t.Skip("此刻城裡沒有衛兵")
	}
	s.beginNPCCombat(guard)
	if s.RemovedNPC[s.Location-1]&(1<<uint(guard)) != 0 {
		t.Error("衛兵被永久記下了 —— 那座城之後就沒有守衛了")
	}
	loc := s.Location
	s.Combat, s.Prompt = nil, PromptNone
	if err := s.SetScene(loc, 0, 15, 15); err != nil {
		t.Fatal(err)
	}
	back := false
	for _, v := range s.VisibleNPCs() {
		if v.Index == guard {
			back = true
		}
	}
	if !back {
		t.Error("重新進場景之後衛兵沒有補上")
	}
}

// TestRemovedNPCsSurviveSaveRoundTrip:永久移除的位元存得回去也讀得回來。
//
// ⚠ 這一段 128 B 的位移是從 0x0332 往後累加七個欄位算出來的,
// 算錯不會報錯 —— 只會讓玩家殺掉的人在讀檔之後復活。
// 所以拿隔壁的月門目的地當哨兵,確認沒有踩到別人。
func TestRemovedNPCsSurviveSaveRoundTrip(t *testing.T) {
	s := shrineState(t)
	if s == nil || s.BaseSave == nil {
		return
	}
	s.RemovedNPC[1] = 0x8000_0005
	s.RemovedNPC[31] = 0x0000_00FF
	gates := s.Moongates

	sv, err := s.ExportSave(s.BaseSave)
	if err != nil {
		t.Fatal(err)
	}
	blob, err := sv.Encode()
	if err != nil {
		t.Fatal(err)
	}
	back, err := u5data.ParseSave(blob)
	if err != nil {
		t.Fatal(err)
	}
	if back.RemovedNPC[1] != 0x8000_0005 || back.RemovedNPC[31] != 0x0000_00FF {
		t.Errorf("讀回來是 %08X / %08X", back.RemovedNPC[1], back.RemovedNPC[31])
	}
	if back.Moongates != gates {
		t.Error("月門目的地被踩到了 —— 128 B 那一段的位移大概算錯了")
	}
}

// TestSceneCombatUsesTheTownRule:在城裡動手只打得到眼前那一個。
//
// 原版 `spawnEnemies` 的 `inTown && base != 衛兵 && small` → 一隻。
// 這條確認合成物件走的是同一條規則(`Raw[ObjShipHull]` 留 0 才算 small)。
func TestSceneCombatUsesTheTownRule(t *testing.T) {
	s := inTown(t)
	if s == nil {
		return
	}
	victim := -1
	for _, v := range s.VisibleNPCs() {
		if v.NPC.Creature&0xFC != u5data.CreatureGuard {
			victim = v.Index
			break
		}
	}
	if victim < 0 {
		t.Skip("此刻城裡只有衛兵")
	}
	if !s.beginNPCCombat(victim) {
		t.Skip("打不起來")
	}
	if s.Combat == nil {
		t.Fatal("沒有進到戰鬥")
	}
	n := 0
	for i := range s.Combat.Units {
		u := &s.Combat.Units[i]
		if u.Active() && u.Flags&UnitMonster != 0 {
			n++
		}
	}
	if n != 1 {
		t.Errorf("城裡動手卻出現 %d 個敵人,預期 1 個", n)
	}
}
