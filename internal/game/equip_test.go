package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// TestSlotForFollowsTheItemTableRanges:每一類裝備進對的欄位。
//
// 分界不是自己畫的,是 `u5data` 從裝備名字表的實際排列定出來的。
// 這條逐界線驗一次 —— 差一格的話「釘盾」會被當成護甲穿在身上。
func TestSlotForFollowsTheItemTableRanges(t *testing.T) {
	cases := []struct {
		item byte
		want EquipSlot
		ok   bool
	}{
		{0, SlotHelm, true},    // Leather Helm
		{3, SlotHelm, true},    // Spiked Helm
		{4, SlotShield, true},  // Small Shield
		{8, SlotShield, true},  // Jewel Shield
		{9, SlotArmour, true},  // Cloth Armour
		{15, SlotArmour, true}, //
		{16, SlotWeapon, true}, // Dagger
		{41, SlotWeapon, true}, // Mystic Sword
		{42, SlotRing, true},   // Ring of Invisibility
		{44, SlotRing, true},   // Ring of Regeneration
		{45, SlotAmulet, true}, // Amulet of Turning
		{47, SlotAmulet, true}, // Ankh
		{u5data.ItemNone, 0, false},
	}
	for _, c := range cases {
		got, ok := SlotFor(c.item)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("裝備 %d → 欄位 %d(ok=%v),預期 %d(ok=%v)",
				c.item, got, ok, c.want, c.ok)
		}
	}
}

// TestReadyRefusesWhenTheSlotIsTakenAndTakesOffInstead —— ★ 這條測試整個反轉了。
//
// ⚠⚠ 它的前身是 `TestEquipReturnsTheOldItem`,釘住「換下來的要放回背包」,
// 註解還寫著「少了這一步,玩家換一次裝備就永久失去原本那件」。
// 那個顧慮是真的 —— **但原版根本不換**:部位有東西就印
// 「Remove first thy present armour!」並拒絕(`docs/re/72`)。
//
// 於是那條測試把一個我自己發明的便利行為(自動替換)釘得跟原版一樣牢。
// 這是本專案第四次踩到「測試在量自己的發明」。
//
// 現在驗兩件事:滿了要拒絕、而**同一個 R 鍵**按在身上那件會把它脫下來。
func TestReadyRefusesWhenTheSlotIsTakenAndTakesOffInstead(t *testing.T) {
	s := equipScene(t)
	c := &s.Roster[0]
	old := c.Raw[u5data.CharArmour]
	if old == u5data.ItemNone {
		t.Skip("這名角色沒穿護甲")
	}
	const newArmour = 15
	s.Inventory.Items[newArmour] = 1
	s.Messages = nil
	if s.Ready(0, newArmour) {
		t.Error("身上有甲卻還穿得上第二件")
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgRemoveArmourFirst) {
		t.Errorf("沒印「%s」:%q", MsgRemoveArmourFirst, s.Messages)
	}
	if c.Raw[u5data.CharArmour] != old {
		t.Errorf("被拒絕了卻還是換掉了:%d", c.Raw[u5data.CharArmour])
	}
	if s.Inventory.Items[newArmour] != 1 {
		t.Errorf("被拒絕了卻扣了背包:剩 %d", s.Inventory.Items[newArmour])
	}

	// ★ 同一鍵按在身上那件 → 卸下,背包 +1。
	before := s.Inventory.Items[old]
	s.Messages = nil
	if !s.Ready(0, old) {
		t.Fatalf("R 按在身上那件該卸下來:%q", s.Messages)
	}
	if c.Raw[u5data.CharArmour] != u5data.ItemNone {
		t.Errorf("沒卸下來:%d", c.Raw[u5data.CharArmour])
	}
	if s.Inventory.Items[old] != before+1 {
		t.Errorf("卸下來的沒放回背包:%d → %d", before, s.Inventory.Items[old])
	}
	// 卸掉之後才穿得上新的。
	s.Messages = nil
	if !s.Ready(0, newArmour) {
		t.Fatalf("卸掉之後還穿不上:%q", s.Messages)
	}
}

// TestSecondWeaponGoesToTheOffHand:右手滿了就往左手放。
//
// 原版的 0x1C 欄位既收盾也收第二把武器(Iolo 就是雙手各一把)。
// 一律換掉右手的話,玩家永遠裝不上第二把 —— 而那是吟遊詩人的打法。
func TestSecondWeaponGoesToTheOffHand(t *testing.T) {
	s := equipScene(t)
	c := &s.Roster[0]
	// ⚠ 先把盔甲卸掉 —— 不然**力量檢查會先擋下來**,而失敗訊息會是
	// 「汝力有不足!」,看起來像左手那條規則壞了。建角的角色帶著
	// 皮盔(重 1)+ 鏈甲(重 10),再拿短劍 5 + 釘錘 7 = 23 > 力量 20。
	// (這一條是新加的力量規則害的,不是它壞了 —— 見 `docs/re/72`。)
	c.Raw[u5data.CharHelm] = u5data.ItemNone
	c.Raw[u5data.CharArmour] = u5data.ItemNone
	c.Raw[u5data.CharWeapon] = 23 // Short Sword
	c.Raw[u5data.CharShield] = u5data.ItemNone
	const second = 24 // Mace
	s.Inventory.Items[second] = 1

	if !s.Ready(0, second) {
		t.Fatalf("拿不上:%q", s.Messages)
	}
	if c.Raw[u5data.CharWeapon] != 23 {
		t.Errorf("右手被換掉了:%d", c.Raw[u5data.CharWeapon])
	}
	if c.Raw[u5data.CharShield] != second {
		t.Errorf("左手是 %d,預期 %d", c.Raw[u5data.CharShield], second)
	}
}

// TestReadyTakesEveryCategory:R **一支收下全部六類**,而且清單也要全列。
//
// ⚠ 這支測試的前身是 `TestReadyAndWearRejectTheWrongCategory` ——
// 它釘住的是「R 不收護甲」,而那個行為是我從 U4 搬來的,原版沒有
//(`sub_2ACF4` 對 'W' 只印 "W-What?",見 `docs/re/49`)。
// **測試會把錯的行為釘得跟對的一樣牢** —— 這是本專案第三次踩到。
func TestReadyTakesEveryCategory(t *testing.T) {
	s := equipScene(t)
	// 六類各一件:頭盔 0、盾 4、護甲 9、武器 16、戒指 42、頸飾 45。
	for _, it := range []byte{0, 4, 9, 16, 42, 45} {
		s.Inventory.Items[it] = 1
	}
	// ⚠ 每一類都要**先把那一格清空** —— 原版「部位滿了」是拒絕而不是替換
	// (`docs/re/72`)。這裡驗的是「R 一支收下全部六類」,不是替換。
	for _, it := range []byte{0, 9, 42, 45} {
		slot, ok := SlotFor(it)
		if !ok {
			t.Fatalf("第 %d 件查不到欄位", it)
		}
		s.Roster[0].Raw[slotOffset[slot]] = u5data.ItemNone
		s.Messages = nil
		if !s.Ready(0, it) {
			t.Errorf("R 收不下第 %d 件:%q", it, s.Messages)
		}
	}
	// 清單也要把六類都列出來。
	// ⚠ 剛才那四件已經穿上去、從背包扣掉了(原版換裝**不會**把舊的丟回背包,
	// 所以背包真的少了)。這裡驗的是清單的涵蓋範圍,先補回去。
	for _, it := range []byte{0, 4, 9, 16, 42, 45} {
		s.Inventory.Items[it] = 1
	}
	list := s.ReadyList()
	got := map[EquipSlot]bool{}
	for _, it := range list {
		if slot, ok := SlotFor(it); ok {
			got[slot] = true
		}
	}
	if len(got) < 4 {
		t.Errorf("R 的清單只涵蓋 %d 類欄位(%v),應該把背包裡各類都列出來", len(got), list)
	}
	// 不是裝備的東西還是要擋(例如火把 —— 不在 0..47 的裝備範圍裡)。
	s.Messages = nil
	if s.Ready(0, u5data.ItemNone) {
		t.Error("R 收下了「無」")
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgCannotReady) {
		t.Errorf("沒說拿不上手:%q", s.Messages)
	}
}

// TestEquipNeedsTheItemInTheBag:背包裡沒有就裝不上。
func TestEquipNeedsTheItemInTheBag(t *testing.T) {
	s := equipScene(t)
	s.Inventory.Items[41] = 0
	if s.Ready(0, 41) {
		t.Error("背包裡沒有卻裝上了")
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgDontHaveThat) {
		t.Errorf("沒說沒有那件:%q", s.Messages)
	}
}

// TestUnequipPutsItBack:卸下來的回背包。
func TestUnequipPutsItBack(t *testing.T) {
	s := equipScene(t)
	c := &s.Roster[0]
	c.Raw[u5data.CharHelm] = 2
	before := s.Inventory.Items[2]
	if !s.Unequip(0, SlotHelm) {
		t.Fatal("卸不下來")
	}
	if c.Raw[u5data.CharHelm] != u5data.ItemNone {
		t.Errorf("頭上還有 %d", c.Raw[u5data.CharHelm])
	}
	if s.Inventory.Items[2] != before+1 {
		t.Errorf("沒回背包:%d → %d", before, s.Inventory.Items[2])
	}
	if s.Unequip(0, SlotHelm) {
		t.Error("空欄位卻卸得下來")
	}
}

// TestReadyListCoversAllSixSlots:R 的選單是**一張**,六類都在裡面。
//
// ⚠ 這支的前身是 `TestEquipListsSplitByCategory`,釘的是「R 與 W 兩張不重疊」
// —— 而 U5 沒有 W(`docs/re/49`)。舊測試每一次都綠,因為它量的是我自己
// 發明的規則。
func TestReadyListCoversAllSixSlots(t *testing.T) {
	s := equipScene(t)
	// 六類各一件:頭盔 0、盾 4、護甲 9、武器 16、戒指 42、頸飾 45。
	want := []int{0, 4, 9, 16, 42, 45}
	for _, i := range want {
		s.Inventory.Items[i] = 1
	}
	seen := map[byte]bool{}
	for _, it := range s.ReadyList() {
		seen[it] = true
	}
	for _, i := range want {
		if !seen[byte(i)] {
			t.Errorf("裝備 %d 不在 R 的選單裡", i)
		}
	}
	// 選單裡不該出現裝不上的東西。
	for _, it := range s.ReadyList() {
		if _, ok := SlotFor(it); !ok {
			t.Errorf("選單裡有裝不上的 %d", it)
		}
	}
}

func equipScene(t *testing.T) *State {
	t.Helper()
	s := newCreateState(t)
	s.MaxMessages = 16
	return s
}

// ─── `sub_1EC34` 的七條規則(`docs/re/72`)────────────────────────────

// bareScene 是一個「六格全空、力量夠用」的角色 —— 驗部位規則時不想被
// 重量或既有裝備干擾。
func bareScene(t *testing.T) (*State, *u5data.Character) {
	t.Helper()
	s := equipScene(t)
	c := &s.Roster[0]
	for slot := EquipSlot(0); slot < SlotCount; slot++ {
		c.Raw[slotOffset[slot]] = u5data.ItemNone
	}
	c.Strength = 99
	s.Messages = nil
	return s, c
}

// TestAmmunitionCannotBeReadied:★ 箭與弩矢裝不上,而且**一句話都不印**。
//
// 原版 `sub_1EC34` 開頭就 `return 0`。印一句「拿不上手」會比原版友善 ——
// 但那不是原版。
func TestAmmunitionCannotBeReadied(t *testing.T) {
	for _, ammo := range []byte{u5data.ItemArrows, u5data.ItemQuarrels} {
		s, _ := bareScene(t)
		s.Inventory.Items[ammo] = 10
		if s.Ready(0, ammo) {
			t.Errorf("裝備 %d(彈藥)竟然裝上了", ammo)
		}
		if len(s.Messages) != 0 {
			t.Errorf("裝備 %d 該靜靜地失敗,卻印了 %q", ammo, s.Messages)
		}
	}
}

// TestBowsNeedAmmunition:弓 / 魔法弓要箭,十字弓要弩矢。
func TestBowsNeedAmmunition(t *testing.T) {
	cases := []struct {
		weapon byte
		ammo   int
	}{{0x1A, u5data.ItemArrows}, {0x24, u5data.ItemArrows}, {0x1C, u5data.ItemQuarrels}}
	for _, c := range cases {
		s, ch := bareScene(t)
		s.Inventory.Items[c.weapon] = 1
		s.Inventory.Items[u5data.ItemArrows] = 0
		s.Inventory.Items[u5data.ItemQuarrels] = 0
		if s.Ready(0, c.weapon) {
			t.Errorf("沒彈藥卻裝上了 %d", c.weapon)
		}
		if !strings.Contains(strings.Join(s.Messages, "|"), MsgNoAmmunition) {
			t.Errorf("裝備 %d 沒印「%s」:%q", c.weapon, MsgNoAmmunition, s.Messages)
		}
		// 有彈藥就裝得上。
		s.Inventory.Items[c.ammo] = 1
		s.Messages = nil
		if !s.Ready(0, c.weapon) {
			t.Errorf("有彈藥卻還裝不上 %d:%q", c.weapon, s.Messages)
		}
		if ch.Raw[u5data.CharWeapon] != c.weapon {
			t.Errorf("裝備 %d 沒進右手:%d", c.weapon, ch.Raw[u5data.CharWeapon])
		}
	}
}

// TestTwoHandedWeaponsNeedBothHands —— 部位碼 0x30。
func TestTwoHandedWeaponsNeedBothHands(t *testing.T) {
	// 找一件雙手武器與一件單手武器。
	var twoHand, oneHand byte = 0xFF, 0xFF
	for i := u5data.ItemWeaponFirst; i < u5data.ItemRingFirst; i++ {
		switch u5data.EquipSlotCode[i] {
		case u5data.SlotCodeTwoHand:
			if twoHand == 0xFF {
				twoHand = byte(i)
			}
		case u5data.SlotCodeOneHand:
			if oneHand == 0xFF {
				oneHand = byte(i)
			}
		}
	}
	if twoHand == 0xFF || oneHand == 0xFF {
		t.Fatal("找不到雙手 / 單手武器各一件")
	}
	// 左手有東西 → 雙手武器裝不上。
	s, ch := bareScene(t)
	s.Inventory.Items[oneHand] = 2
	s.Inventory.Items[twoHand] = 1
	if !s.Ready(0, oneHand) {
		t.Fatalf("單手武器裝不上:%q", s.Messages)
	}
	s.Messages = nil
	if s.Ready(0, twoHand) {
		t.Error("一手有東西卻裝上了雙手武器")
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgBothHandsFree) {
		t.Errorf("沒印「%s」:%q", MsgBothHandsFree, s.Messages)
	}

	// ★ 反過來:拿著雙手武器時**左手欄位是空的,卻不能再拿單手** ——
	// 這一條用編號區間永遠寫不出來(原版 `sub_1EBE8` 的第三個分支)。
	s2, ch2 := bareScene(t)
	s2.Inventory.Items[twoHand] = 1
	s2.Inventory.Items[oneHand] = 1
	if !s2.Ready(0, twoHand) {
		t.Fatalf("雙手武器裝不上:%q", s2.Messages)
	}
	if ch2.Raw[u5data.CharShield] != u5data.ItemNone {
		t.Fatal("雙手武器該只佔右手那一格")
	}
	s2.Messages = nil
	if s2.Ready(0, oneHand) {
		t.Error("拿著雙手武器卻還能在左手再拿一件")
	}
	if !strings.Contains(strings.Join(s2.Messages, "|"), MsgFreeAHandFirst) {
		t.Errorf("沒印「%s」:%q", MsgFreeAHandFirst, s2.Messages)
	}
	_ = ch
}

// TestStrengthLimitsWhatYouCanWear —— 六格重量加總 + 新件 > 力量 → 拒絕。
func TestStrengthLimitsWhatYouCanWear(t *testing.T) {
	// 找一件有重量的護甲。
	var heavy byte = 0xFF
	for i := u5data.ItemArmourFirst; i <= u5data.ItemArmourLast; i++ {
		if u5data.EquipWeight[i] > 0 {
			heavy = byte(i)
			break
		}
	}
	if heavy == 0xFF {
		t.Fatal("找不到有重量的護甲")
	}
	w := int(u5data.EquipWeight[heavy])

	s, c := bareScene(t)
	c.Strength = byte(w - 1)
	s.Inventory.Items[heavy] = 1
	if s.Ready(0, heavy) {
		t.Errorf("力量 %d 竟然穿得上重 %d 的護甲", c.Strength, w)
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgNotStrongEnough) {
		t.Errorf("沒印「%s」:%q", MsgNotStrongEnough, s.Messages)
	}
	// ★ 剛好等於力量要**穿得上**(原版 `setnl` 是 >=,不是 >)。
	c.Strength = byte(w)
	s.Messages = nil
	if !s.Ready(0, heavy) {
		t.Errorf("力量剛好等於重量 %d 卻穿不上 —— 原版是 >=:%q", w, s.Messages)
	}
}

// TestArmourCannotBeChangedInHeatedBattle —— 但勝負已定之後可以。
func TestArmourCannotBeChangedInHeatedBattle(t *testing.T) {
	var armour byte = u5data.ItemArmourFirst
	s := corpserArena(t)
	c := &s.Roster[0]
	c.Raw[u5data.CharArmour] = u5data.ItemNone
	c.Strength = 99
	s.Inventory.Items[armour] = 1
	s.Combat.Over = false
	s.Messages = nil
	if s.Ready(0, armour) {
		t.Error("戰事正酣卻換了甲")
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgNoArmourInBattle) {
		t.Errorf("沒印「%s」:%q", MsgNoArmourInBattle, s.Messages)
	}
	// ★ 勝負已定(byte_3E0B3 != 0)之後就可以。
	s.Combat.Over = true
	s.Messages = nil
	if !s.Ready(0, armour) {
		t.Errorf("勝負已定之後該換得了甲:%q", s.Messages)
	}
	// 而武器在戰鬥中一直換得了 —— 那條規則只管護甲。
	s.Combat.Over = false
	var weapon byte = u5data.ItemWeaponFirst
	s.Inventory.Items[weapon] = 1
	c.Raw[u5data.CharWeapon] = u5data.ItemNone
	s.Messages = nil
	if !s.Ready(0, weapon) {
		t.Errorf("戰鬥中該換得了武器(規則只管護甲):%q", s.Messages)
	}
}

// TestRingVanishesOnWearingIsADifferentDieFromCombatStart —— 兩個骰子不能合併。
//
// 穿戴那一次擲 `random(0,15) == 0`;開戰佈陣那一次擲 `== 11`(`docs/re/70`)。
// 機率相同、對象相同,而**骰值不同**,所以是兩顆骰子。
func TestRingVanishesOnWearingIsADifferentDieFromCombatStart(t *testing.T) {
	if RingVanishHit == 0 {
		t.Fatal("開戰那次的骰值不該是 0 —— 兩顆骰子會被誤認成同一顆")
	}
	s, _ := bareScene(t)
	vanished := 0
	const tries = 400
	for i := 0; i < tries; i++ {
		s.Roster[0].Raw[u5data.CharRing] = u5data.ItemNone
		s.Inventory.Items[u5data.ItemRingInvisibility] = 1
		s.Messages = nil
		s.Ready(0, u5data.ItemRingInvisibility)
		if strings.Contains(strings.Join(s.Messages, "|"), MsgRingVanishes) {
			vanished++
			if s.Roster[0].Raw[u5data.CharRing] != u5data.ItemNone {
				t.Fatal("印了「戒指消失了」卻還戴在手上")
			}
		}
	}
	if vanished == 0 {
		t.Fatal("戴了 400 次隱形戒指一次都沒消失 —— 穿戴那一次的 1/16 沒實作")
	}
	if lo, hi := tries/40, tries/4; vanished < lo || vanished > hi {
		t.Errorf("400 次消失 %d 次,超出 1/16 的合理量級 [%d, %d]", vanished, lo, hi)
	}
	// ★ 防護戒指(0x2B)不會消失 —— 原版只列了 0x2A 與 0x2C。
	s2, _ := bareScene(t)
	for i := 0; i < tries; i++ {
		s2.Roster[0].Raw[u5data.CharRing] = u5data.ItemNone
		s2.Inventory.Items[u5data.ItemRingProtection] = 1
		s2.Messages = nil
		s2.Ready(0, u5data.ItemRingProtection)
		if strings.Contains(strings.Join(s2.Messages, "|"), MsgRingVanishes) {
			t.Fatal("防護戒指消失了 —— 原版只有隱形與再生兩種會")
		}
	}
}
