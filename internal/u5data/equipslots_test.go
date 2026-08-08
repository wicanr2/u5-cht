package u5data

import (
	"os"
	"testing"
)

// TestEquipSlotCodeTableIsExactlyFortyEight:表長必須恰好等於裝備數。
//
// 這是「從 `.asm` 展開對了」的第一個佐證 —— IDA 把那張表拆成
// `db 4 dup(80h)` + 一個字串 + `dw` + 四個 `dd`(因為 0x20/0x30/0x40 都是
// 可見字元),展開後要剛好 48 筆,多一筆或少一筆都代表哪一段讀錯了。
func TestEquipSlotCodeTableIsExactlyFortyEight(t *testing.T) {
	if len(EquipSlotCode) != ItemCount || len(EquipWeight) != ItemCount {
		t.Fatalf("部位碼 %d 筆、重量 %d 筆,裝備數是 %d",
			len(EquipSlotCode), len(EquipWeight), ItemCount)
	}
}

// TestOnlyAmmunitionHasNoSlot:★ 只有箭與弩矢裝不上,而它們正是兩種彈藥。
//
// 三條互不相干的證據指向同一對編號:
//
//  1. 部位碼表裡**只有兩筆** 0x00 —— 27 與 29。
//  2. `sub_1EC34` 開頭 `if (item == 1Bh || item == 1Dh) return`。
//  3. 弓查 `byte_3DFEB`、十字弓查 `byte_3DFED`,換算回存檔正是
//     `Items[27]` / `Items[29]`。
func TestOnlyAmmunitionHasNoSlot(t *testing.T) {
	var none []int
	for i, code := range EquipSlotCode {
		if code == SlotCodeNone {
			none = append(none, i)
		}
	}
	if len(none) != 2 || none[0] != ItemArrows || none[1] != ItemQuarrels {
		t.Errorf("沒有部位的是 %v,預期正好是箭 %d 與弩矢 %d",
			none, ItemArrows, ItemQuarrels)
	}
}

// TestOneAndTwoHandedWeaponsInterleave 是「必須查表、不能用編號區間」的證據。
//
// 若單手與雙手在編號上是分段的,`ItemWeaponFirst..Last` 那種區間就夠用了。
// 這條要求它們**真的交錯** —— 交錯一旦成立,區間法就永遠推不出來。
func TestOneAndTwoHandedWeaponsInterleave(t *testing.T) {
	// 16 單手、17 雙手、18 單手、19 雙手 —— 開頭四筆就交錯。
	want := []byte{SlotCodeOneHand, SlotCodeTwoHand, SlotCodeOneHand, SlotCodeTwoHand}
	for i, w := range want {
		if got := EquipSlotCode[16+i]; got != w {
			t.Errorf("裝備 %d 的部位碼是 0x%02X,預期 0x%02X", 16+i, got, w)
		}
	}
	flips := 0
	for i := ItemWeaponFirst; i < ItemRingFirst-1; i++ {
		a, b := EquipSlotCode[i], EquipSlotCode[i+1]
		if (a == SlotCodeOneHand || a == SlotCodeTwoHand) &&
			(b == SlotCodeOneHand || b == SlotCodeTwoHand) && a != b {
			flips++
		}
	}
	if flips < 4 {
		t.Errorf("單手 / 雙手只換手 %d 次 —— 交錯不成立的話這張表就沒有存在的必要", flips)
	}
}

// TestEquipAmmoOnlyForThreeWeapons:★ 只有弓 / 魔法弓 / 十字弓要查彈藥。
func TestEquipAmmoOnlyForThreeWeapons(t *testing.T) {
	want := map[byte]int{0x1A: ItemArrows, 0x24: ItemArrows, 0x1C: ItemQuarrels}
	for i := 0; i < ItemCount; i++ {
		ammo, needs := EquipAmmoFor(byte(i))
		w, ok := want[byte(i)]
		if needs != ok {
			t.Errorf("裝備 %d 要查彈藥 = %v,預期 %v", i, needs, ok)
			continue
		}
		if needs && ammo != w {
			t.Errorf("裝備 %d 查的是 %d,預期 %d", i, ammo, w)
		}
	}
}

// TestStartingRosterCanCarryItsOwnEquipment 是重量表最強的一個 oracle。
//
// ★ `INIT.GAM` 裡 15 名預設角色的裝備重量**全部**不超過各自的力量 ——
// 而其中好幾個貼著上限(Geoffrey 20/24、Dupre 21/22、Saduj 16/21)。
// 若重量表讀錯一格、或「六格加總」這條規則讀錯,幾乎一定有人會超標。
//
// ⇒ 這條不是在驗程式碼,是在驗**我抄的那 48 個數字**。
func TestStartingRosterCanCarryItsOwnEquipment(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("需要 U5_GAMEDATA")
	}
	raw, err := os.ReadFile(dir + "/INIT.GAM")
	if err != nil {
		t.Skip(err)
	}
	sv, err := ParseSave(raw)
	if err != nil {
		t.Fatal(err)
	}
	checked, tight := 0, 0
	for i := range sv.Roster {
		c := &sv.Roster[i]
		if c.Name == "" || c.Strength == 0 {
			continue
		}
		checked++
		w := EquipTotalWeight(c)
		if w > int(c.Strength) {
			t.Errorf("%s(力量 %d)的裝備重 %d —— 原版不可能讓預設角色超重",
				c.Name, c.Strength, w)
		}
		if int(c.Strength)-w <= 3 {
			tight++
		}
	}
	if checked < 10 {
		t.Fatalf("只驗到 %d 名角色,`INIT.GAM` 該有 15 名", checked)
	}
	// ★ 有人貼著上限才說明這張表**有在起作用** —— 全部遠低於上限的話,
	// 就算重量表整排寫成 0 這條測試也會過。
	if tight == 0 {
		t.Error("沒有任何角色貼近力量上限 —— 這樣的話重量表全填 0 也會過關")
	}
}
