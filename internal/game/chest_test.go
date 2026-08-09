package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// findDungeonTile 在八座地牢裡找一格符合條件的,把玩家放上去。
func findDungeonTile(t *testing.T, s *State, want byte) bool {
	t.Helper()
	for dg := 0; dg < u5data.DungeonCount; dg++ {
		for l := 0; l < u5data.DungeonLevels; l++ {
			for y := 0; y < u5data.DungeonSide; y++ {
				for x := 0; x < u5data.DungeonSide; x++ {
					if u5data.DungeonKind(s.Dungeons.At(dg, l, x, y)) != want {
						continue
					}
					s.Dungeon = &DungeonState{Index: dg,
						Location: u5data.DungeonLocationBase + dg,
						Level:    l, X: x, Y: y, Facing: South}
					s.Location = s.Dungeon.Location
					return true
				}
			}
		}
	}
	return false
}

// TestOpenChestTurnsItIntoAnOpenedOne:開過的寶箱格會變成 0x70,而且保留位元 3。
//
// `sub_18D18` 是 `tile = (tile & 8) | 0x70` —— 所以 **0x70 不是門,
// 是「開過的寶箱」**。這條同時把那個位元的語意釘住。
func TestOpenChestTurnsItIntoAnOpenedOne(t *testing.T) {
	s := dungeonState(t)
	if !findDungeonTile(t, s, u5data.DungeonChest) {
		t.Skip("八座地牢裡找不到寶箱")
	}
	d := s.Dungeon
	// 手動加上「頭上有洞」,驗它會被保留。
	orig := s.DungeonTileHere() | u5data.DungeonHoleAbove
	s.Dungeons.Set(d.Index, d.Level, d.X, d.Y, orig)
	actAs(t, s, 0) // 開箱要先決定由誰開;指定過就不會跳選單
	s.OpenChest()
	got := s.DungeonTileHere()
	if got&0xF0 != 0x70 {
		t.Errorf("開完之後是 %02X,預期高四位元 0x70", got)
	}
	if got&u5data.DungeonHoleAbove == 0 {
		t.Error("「頭上有洞」那一位元被清掉了")
	}
	// 開過的箱子不能再開。
	before := s.Inventory.Gold
	s.OpenChest()
	if s.Inventory.Gold != before {
		t.Error("同一個箱子開了兩次還有錢")
	}
}

// TestTrappedChestHurtsSomeone:bit 0 設著的寶箱會發動陷阱。
//
// ⚠⚠ 這條原本斷言「一定有人扣血」,而那是**舊的發明模型**(陷阱一律
// `random(1,20)` 扣血)留下的。逆完 `sub_2AB38` 之後陷阱有四種,其中
// **毒(2/8)與毒氣(1/8)不扣血只改狀態** ⇒ 舊斷言有 3/8 的機率誤判。
// 現在改成「扣血**或**中毒」,並把四種都跑一遍確保不會有哪一種靜默無效。
func TestTrappedChestHurtsSomeone(t *testing.T) {
	s := dungeonState(t)
	if !findDungeonTile(t, s, u5data.DungeonChest) {
		t.Skip("找不到寶箱")
	}
	d := s.Dungeon
	actAs(t, s, 0)
	hurt := map[string]bool{}
	for seed := 1; seed <= 40 && len(hurt) < 2; seed++ {
		s.SeedRandom(int64(seed))
		s.Dungeons.Set(d.Index, d.Level, d.X, d.Y,
			u5data.DungeonChest|u5data.ChestTrappedDungeon)
		total := 0
		for i := 0; i < s.PartySize; i++ {
			s.Roster[i].HP, s.Roster[i].MaxHP = 200, 200
			s.Roster[i].Status = u5data.StatusGood
			total += int(s.Roster[i].HP)
		}
		s.OpenChest()
		after, poisoned := 0, false
		for i := 0; i < s.PartySize; i++ {
			after += int(s.Roster[i].HP)
			if s.Roster[i].Status == u5data.StatusPoisoned {
				poisoned = true
			}
		}
		switch {
		case after < total:
			hurt["扣血"] = true
		case poisoned:
			hurt["中毒"] = true
		default:
			t.Fatalf("種子 %d:陷阱發動了卻既沒扣血也沒中毒:\n%s", seed, s.log())
		}
	}
	if len(hurt) < 2 {
		t.Errorf("四十顆種子只看到 %v —— 預期扣血與中毒兩類都出現過", hurt)
	}
}

// TestChestContentsScaleWithDepth:越深的寶箱掉得越多。
//
// 門檻同時是「等級下限」與「擲骰難度」,所以第 1 層幾乎只掉金幣,
// 第 8 層才有機會掉到門檻 25 那一項。
func TestChestContentsScaleWithDepth(t *testing.T) {
	s := magicState(t)
	s.SeedRandom(2)
	count := func(level, tries int) int {
		n := 0
		for i := 0; i < tries; i++ {
			if s.rollChestContents(level) {
				n++
			}
		}
		return n
	}
	// ⚠ 獎品表最低的門檻是 3,所以**等級 1 或 2 的寶箱永遠是空的**。
	// 這是表本身的性質,不是判斷寫反 —— 我第一版拿深度 1..8 當等級,
	// 第 1 層的箱子就全空了,那正是「等級的來源還沒逆對」的訊號。
	if count(2, 200) != 0 {
		t.Error("等級 2 低於所有門檻,不該掉出東西")
	}
	shallow := count(4, 200)
	deep := count(30, 200)
	if shallow == 0 {
		t.Error("等級 4 過得了門檻 3,200 次全空 —— 門檻判斷大概反了")
	}
	if deep <= shallow {
		t.Errorf("深處的寶箱 %d 次有東西,淺處 %d 次 —— 應該更多", deep, shallow)
	}
	t.Logf("等級 4:%d/200 有東西;等級 30:%d/200", shallow, deep)
}

// TestAnGravDestroysField:破力場把 0x8x 清成只剩位元 3。
func TestAnGravDestroysField(t *testing.T) {
	s := dungeonState(t)
	if !findDungeonTile(t, s, u5data.DungeonMagic) {
		t.Skip("八座地牢裡找不到力場")
	}
	if !s.destroyField() {
		t.Fatalf("破不了力場:\n%s", s.log())
	}
	if got := s.DungeonTileHere(); got&^u5data.DungeonHoleAbove != 0 {
		t.Errorf("破完之後是 %02X,預期只剩位元 3", got)
	}
	// 破過就沒了。
	if s.destroyField() {
		t.Error("同一個力場破了兩次")
	}
}

// TestUnlockLoweringTheTileByOne:解鎖就是 tile − 1。
func TestUnlockLoweringTheTileByOne(t *testing.T) {
	s := magicState(t)
	s.Location = 0
	// 在旁邊放一扇鎖著的門。
	if !s.SetTileAt(s.X+1, s.Y, u5data.TileLockedDoor) {
		t.Skip("這一層的地圖不支援寫入")
	}
	if !s.UnlockAhead(East) {
		t.Fatalf("解不開:\n%s", s.log())
	}
	if got := s.TileAt(s.X+1, s.Y); got != u5data.TileLockedDoor-1 {
		t.Errorf("解鎖之後 tile 是 %02X,預期 %02X", got, u5data.TileLockedDoor-1)
	}
	// 沒鎖的東西解不了。
	if s.UnlockAhead(East) {
		t.Error("已經解開的門又解了一次")
	}
	// 0xBB 那一扇也是同一條規則。
	s.SetTileAt(s.X+1, s.Y, u5data.TileLockedMagicDoor)
	if !s.UnlockAhead(East) {
		t.Error("0xBB 解不開")
	}
	if got := s.TileAt(s.X+1, s.Y); got != u5data.TileLockedMagicDoor-1 {
		t.Errorf("0xBB 解鎖之後是 %02X", got)
	}
}

// ★ 地表寶箱的種類碼就是 `sub_154BC` 的物品碼 —— 兩張獨立的表用同一組編碼。
func TestSurfaceChestKindsAreItemCodes(t *testing.T) {
	want := map[byte]string{
		u5data.ItemClosedChest: "上鎖的箱子",
		u5data.ItemGold:        "金幣",
		u5data.ItemPotion:      "藥水",
		u5data.ItemScroll:      "卷軸",
		u5data.ItemKey:         "鑰匙",
		u5data.ItemGem:         "寶石",
		u5data.ItemTorch:       "火把",
		u5data.ItemFood:        "食物",
	}
	for _, k := range u5data.ChestKind {
		if _, ok := want[k]; !ok {
			t.Errorf("寶箱種類碼 %d 不在 sub_154BC 的物品碼裡", k)
		}
	}
	if len(u5data.ChestKind) != len(want) {
		t.Errorf("寶箱有 %d 種,物品碼列了 %d 種", len(u5data.ChestKind), len(want))
	}
	// ⚠ 檀香木盒(0x0E)**刻意不在**任何獎品表裡 —— 它是唯一的劇情物品。
	for _, k := range u5data.ChestKind {
		if k == u5data.ItemSandalwood {
			t.Error("檀香木盒不該從寶箱掉出來")
		}
	}
	for _, k := range u5data.DungeonLootKind {
		if k == u5data.ItemSandalwood {
			t.Error("檀香木盒不該從地牢寶箱掉出來")
		}
	}
}

// 寶箱掉出來的東西真的進背包(原本只印一行「找到了什麼東西」)。
func TestChestLootReachesTheBackpack(t *testing.T) {
	s := getScene(t, 0, 0)
	s.sceneObjects.Objects[3] = u5data.MapObject{}
	before := s.Inventory.Gold + s.Inventory.Gems + s.Inventory.Keys +
		s.Inventory.Torches + s.Inventory.Food
	// 等級開到最大,八種門檻全過的機率很高;跑幾次總會拿到東西。
	got := false
	for i := 0; i < 40 && !got; i++ {
		s.rollChestContents(99)
		after := s.Inventory.Gold + s.Inventory.Gems + s.Inventory.Keys +
			s.Inventory.Torches + s.Inventory.Food
		got = after > before
	}
	if !got {
		t.Error("開了四十次高等寶箱,背包一樣都沒多 —— 獎品沒接進背包")
	}
}

// ─── 地牢裡的 J(原版 `sub_14B2C`,`docs/re/76`)────────────────────────

// jimmyDungeonScene 把玩家放在一個寶箱格上,並給足鑰匙。
func jimmyDungeonScene(t *testing.T, trap byte) *State {
	t.Helper()
	s := dungeonState(t)
	if !findDungeonTile(t, s, u5data.DungeonChest) {
		t.Skip("八座地牢裡找不到寶箱")
	}
	d := s.Dungeon
	s.Dungeons.Set(d.Index, d.Level, d.X, d.Y, u5data.DungeonChest|trap)
	s.Inventory.Keys = 50
	s.Messages = nil
	return s
}

// TestJimmyInADungeonDisarmsTheTrapNotALock —— ★★ 解的是陷阱,不是鎖。
//
// 成功之後那一格還是 0x4x(箱子),只是低三位元被清掉。
// 所以 J 在地牢裡的用途是「先解陷阱,再用 O 開」。
func TestJimmyInADungeonDisarmsTheTrapNotALock(t *testing.T) {
	s := jimmyDungeonScene(t, 0x03) // 有陷阱
	d := s.Dungeon
	// 敏捷拉高讓它一定成功(門檻 = (樓層×2 + 30 − 敏捷)/2,骰 1..30 要大於門檻)。
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		s.Roster[i].Dex = 99
	}
	actAs(t, s, 0)
	before := s.Inventory.Keys
	s.Jimmy()
	got := s.DungeonTileHere()
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgChestUnlocked) {
		t.Fatalf("沒印「%s」:%q", MsgChestUnlocked, s.Messages)
	}
	if u5data.DungeonKind(got) != u5data.DungeonChest {
		t.Errorf("解完之後是 0x%02X —— 該還是箱子(0x4x)", got)
	}
	if got&u5data.DungeonChestTrapMask != 0 {
		t.Errorf("陷阱位元沒清掉:0x%02X", got)
	}
	if s.Inventory.Keys != before {
		t.Errorf("成功了卻扣了鑰匙:%d → %d", before, s.Inventory.Keys)
	}
	_ = d
}

// TestJimmyWastesAKeyOnAnUntrappedChest —— ★ 沒有陷阱的箱子撬不開,鑰匙照斷。
//
// 判準是 `tile & 0xF7 == 0x40`(tile ∈ {0x40, 0x48})—— 低三位元為 0,
// 沒有東西可解,所以原版直接跳到「鑰匙斷了」。寫成「沒陷阱就成功」
// 會讓鑰匙變成萬能鑰匙。
func TestJimmyWastesAKeyOnAnUntrappedChest(t *testing.T) {
	for _, extra := range []byte{0, u5data.DungeonHoleAbove} {
		s := jimmyDungeonScene(t, extra)
		for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
			s.Roster[i].Dex = 99 // 就算敏捷拉滿也一樣
		}
		actAs(t, s, 0)
		before := s.Inventory.Keys
		s.Jimmy()
		joined := strings.Join(s.Messages, "|")
		if !strings.Contains(joined, MsgKeyBroke) {
			t.Errorf("tile 0x%02X 該印「%s」:%q",
				u5data.DungeonChest|extra, MsgKeyBroke, s.Messages)
		}
		if strings.Contains(joined, MsgChestUnlocked) {
			t.Errorf("沒有陷阱的箱子竟然「解開」了:%q", s.Messages)
		}
		if s.Inventory.Keys != before-1 {
			t.Errorf("鑰匙沒斷:%d → %d", before, s.Inventory.Keys)
		}
	}
}

// TestJimmyInADungeonAsksWhoBeforeCheckingKeys —— ★ 順序與門那條相反。
//
// 地牢這條是「選人 → 讀格子 → 查鑰匙」,所以沒鑰匙時**照樣先問**;
// 門那條是先查鑰匙(`sub_14CAC` 的 `cmp byte_3DFB8, 0` 在問方向之前)。
func TestJimmyInADungeonAsksWhoBeforeCheckingKeys(t *testing.T) {
	s := jimmyDungeonScene(t, 0x03)
	actAs(t, s, 0)
	s.Inventory.Keys = 0
	s.Messages = nil
	s.Jimmy()
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgNoKeys) {
		t.Errorf("沒鑰匙該印「%s」:%q", MsgNoKeys, s.Messages)
	}
	// 而且不該把陷阱清掉。
	if s.DungeonTileHere()&u5data.DungeonChestTrapMask == 0 {
		t.Error("沒鑰匙卻把陷阱解掉了")
	}
}

// TestJimmyOnAnOpenedChestSaysAlreadyOpen。
func TestJimmyOnAnOpenedChestSaysAlreadyOpen(t *testing.T) {
	s := jimmyDungeonScene(t, 0x03)
	d := s.Dungeon
	s.Dungeons.Set(d.Index, d.Level, d.X, d.Y, u5data.DungeonOpenedChestKind)
	actAs(t, s, 0)
	s.Messages = nil
	s.Jimmy()
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgAlreadyOpen) {
		t.Errorf("開過的箱子該印「%s」:%q", MsgAlreadyOpen, s.Messages)
	}
}

// TestChestTrapKindsFollowTheWeightTable —— ★ 四種陷阱的權重表。
//
// `byte_5FFEC = {0,0,0,1,1,2,2,3}` ⇒ 酸 3/8、毒 2/8、炸彈 2/8、毒氣 1/8
//(`docs/re/91`)。⚠ 此前引擎用 `random(1,20)` 的估計傷害,連陷阱**種類**都沒有。
func TestChestTrapKindsFollowTheWeightTable(t *testing.T) {
	want := map[byte]int{0: 3, 1: 2, 2: 2, 3: 1}
	got := map[byte]int{}
	for _, v := range u5data.TrapKindRoll {
		got[v]++
	}
	for k, n := range want {
		if got[k] != n {
			t.Errorf("種類 %d 在表裡出現 %d 次,預期 %d", k, got[k], n)
		}
	}
	if len(u5data.TrapKindRoll) != u5data.TrapKindRollMax+1 {
		t.Errorf("表長 %d,而擲骰上限是 %d", len(u5data.TrapKindRoll), u5data.TrapKindRollMax)
	}
}

// TestCombatTrapsAreOnlyAcidOrPoison —— ★ 戰鬥中不查權重表。
//
// 原版 `cmp byte_3E0A3, 7Fh; ja` → 地點碼 > 0x7F 只擲 `random(0,1)`
// ⇒ **戰鬥中不會有炸彈與毒氣**。引擎的戰鬥地點碼是 0xFF。
func TestCombatTrapsAreOnlyAcidOrPoison(t *testing.T) {
	s := combatState(t)
	// 先在非戰鬥狀態下確認四種都出得來(正對照 —— 否則下面的「只有兩種」
	// 可能是因為擲骰壞了而不是因為分支)。
	seen := map[int]bool{}
	for i := 0; i < 400; i++ {
		seen[s.trapKind()] = true
	}
	if len(seen) < 4 {
		t.Errorf("非戰鬥時只看到 %v,預期四種都可能", seen)
	}
	slot, ok := s.CurrentObjects().Spawn(0xC0, s.X+1, s.Y, s.Floor) // 0xC0 = 法師
	if !ok {
		t.Skip("放不下怪物")
	}
	if !s.BeginCombat(slot) {
		t.Skipf("開不了戰:\n%s", s.log())
	}
	for i := 0; i < 400; i++ {
		if k := s.trapKind(); k > u5data.TrapCombatKindMax {
			t.Fatalf("戰鬥中擲出種類 %d,預期只有 0 或 1", k)
		}
	}
}

// TestGasPoisonsTheLivingOnly —— 毒氣掃全隊,但不動死人。
//
// ⚠ `sub_2AB08` 只在「槽在隊伍裡」且「狀態不是 'D'」時寫 'P'
// ⇒ 毒氣**不會治好也不會復活**任何人。
func TestGasPoisonsTheLivingOnly(t *testing.T) {
	s := dungeonState(t)
	if s.PartySize < 2 {
		t.Skip("隊伍太小")
	}
	for i := 0; i < s.PartySize; i++ {
		s.Roster[i].Status = u5data.StatusGood
	}
	s.Roster[1].Status = u5data.StatusDead
	for i := 0; i < u5data.CombatPartySlots; i++ {
		s.poisonMember(i)
	}
	for i := 0; i < s.PartySize; i++ {
		if i == 1 {
			if s.Roster[i].Status != u5data.StatusDead {
				t.Error("毒氣把死人的狀態改掉了")
			}
			continue
		}
		if s.Roster[i].Status != u5data.StatusPoisoned {
			t.Errorf("隊員 %d 沒中毒(狀態 %02X)", i, s.Roster[i].Status)
		}
	}
	// 越界的槽不能 panic。
	s.poisonMember(-1)
	s.poisonMember(u5data.CombatPartySlots + 5)
}

// TestBombHurtsEveryPartySlot —— 炸彈掃**槽 0..5**,不是掃 PartySize。
//
// 原版 `sub_2A4D0` 的迴圈上限是固定的 6,裡面才逐一檢查
//「槽 < byte_3E06B」與「狀態 != 'D'」⇒ 隊伍不滿時多出來的槽是空轉,
// 不是把迴圈縮短。行為等價,但寫成 PartySize 會在未來加欄位時錯開。
func TestBombHurtsEveryPartySlot(t *testing.T) {
	s := dungeonState(t)
	if s.PartySize < 2 {
		t.Skip("隊伍太小")
	}
	for i := 0; i < s.PartySize; i++ {
		s.Roster[i].Status = u5data.StatusGood
		s.Roster[i].HP, s.Roster[i].MaxHP = 200, 200
	}
	s.Roster[1].Status = u5data.StatusDead
	before := make([]uint16, s.PartySize)
	for i := range before {
		before[i] = s.Roster[i].HP
	}
	s.bombEveryone()
	for i := 0; i < s.PartySize; i++ {
		got := s.Roster[i].HP
		switch {
		case i == 1:
			if got != before[i] {
				t.Errorf("死人被炸彈扣了血(%d → %d)", before[i], got)
			}
		case got == before[i]:
			t.Errorf("隊員 %d 沒被炸到", i)
		case before[i]-got > u5data.TrapBombDamageMax:
			t.Errorf("隊員 %d 掉了 %d 點,上限是 %d",
				i, before[i]-got, u5data.TrapBombDamageMax)
		}
	}
}
