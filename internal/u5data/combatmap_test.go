package u5data

import (
	"os"
	"testing"
)

func loadCBT(t *testing.T, name string) *CombatMapSet {
	t.Helper()
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	s, err := LoadCombatMaps(dir + "/" + name)
	if err != nil {
		t.Fatalf("%s:%v", name, err)
	}
	return s
}

// TestCombatMapCounts:352 B 一張,兩個檔各自切得整齊。
func TestCombatMapCounts(t *testing.T) {
	if n := len(loadCBT(t, "BRIT.CBT").Maps); n != 16 {
		t.Errorf("BRIT.CBT 有 %d 張,預期 16", n)
	}
	if n := len(loadCBT(t, "DUNGEON.CBT").Maps); n != 112 {
		t.Errorf("DUNGEON.CBT 有 %d 張,預期 112", n)
	}
}

// TestCombatEntryPositionsInBounds:入場位置全部落在 11×11 內。
//
// 這是位移解對了的主要證據 —— 352 B 裡大部分位元組是地形碼(值域到 0xB3)
// 與 0,只有真正的座標欄位會**整批**落在 0..10。位移偏一格就會撈到地形碼,
// ParsePrices 式的驗證會立刻炸。這裡再獨立確認一次。
func TestCombatEntryPositionsInBounds(t *testing.T) {
	for _, name := range []string{"BRIT.CBT", "DUNGEON.CBT"} {
		s := loadCBT(t, name)
		for i := range s.Maps {
			m := &s.Maps[i]
			for k := 0; k < CombatPartySlots; k++ {
				if m.PartyX[k] >= CombatSide || m.PartyY[k] >= CombatSide {
					t.Fatalf("%s 第 %d 張:隊員 %d 在 (%d,%d)", name, i, k, m.PartyX[k], m.PartyY[k])
				}
			}
			for k := 0; k < CombatEnemySlots; k++ {
				if m.EnemyX[k] >= CombatSide || m.EnemyY[k] >= CombatSide {
					t.Fatalf("%s 第 %d 張:敵人 %d 在 (%d,%d)", name, i, k, m.EnemyX[k], m.EnemyY[k])
				}
			}
		}
	}
}

// TestCombatFirstMapLayout 鎖住 BRIT.CBT 第 0 張的實際值。
//
// 隊員圍在中央((6,6)(4,4)(6,4)(4,6)(5,3)(3,5)),敵人散在四角與邊上
// ((1,1)(9,1)(1,9)(9,9)…)—— 那正是 U5 戰鬥的擺法。這個「形狀對」比
// 「數字讀得出來」更能證明欄位沒錯位。
func TestCombatFirstMapLayout(t *testing.T) {
	m := &loadCBT(t, "BRIT.CBT").Maps[0]
	if m.PartyX != [CombatPartySlots]byte{6, 4, 6, 4, 5, 3} ||
		m.PartyY != [CombatPartySlots]byte{6, 4, 4, 6, 3, 5} {
		t.Errorf("隊員入場 X=%v Y=%v", m.PartyX, m.PartyY)
	}
	want := [4][2]byte{{1, 1}, {9, 1}, {1, 9}, {9, 9}}
	for k, w := range want {
		if m.EnemyX[k] != w[0] || m.EnemyY[k] != w[1] {
			t.Errorf("敵人 %d 在 (%d,%d),預期 (%d,%d)", k, m.EnemyX[k], m.EnemyY[k], w[0], w[1])
		}
	}
}

// TestCombatMapsHaveTerrain:每張圖都要有地形,不能整片是 0。
//
// 位移算錯最典型的症狀就是「解得出來但全是 0」——
// 那不會觸發任何範圍檢查,只會讓戰場變成一片空白。
func TestCombatMapsHaveTerrain(t *testing.T) {
	for _, name := range []string{"BRIT.CBT", "DUNGEON.CBT"} {
		s := loadCBT(t, name)
		for i := range s.Maps {
			m := &s.Maps[i]
			nonZero := 0
			for y := 0; y < CombatSide; y++ {
				for x := 0; x < CombatSide; x++ {
					if m.Tiles[y][x] != 0 {
						nonZero++
					}
				}
			}
			if nonZero < CombatSide*CombatSide/2 {
				t.Errorf("%s 第 %d 張只有 %d/%d 格有地形", name, i, nonZero, CombatSide*CombatSide)
			}
		}
	}
}

// TestCombatPartySlotsOverlapIsRare:入場位置多半互不重疊,但**不保證**。
//
// 我原本斷言「六個隊員一定不重疊」——那是我自己想的性質,不是原版保證:
// `BRIT.CBT` 第 9 張(白磚走廊那張)就有三個隊員的入場位置都是 (1,1)。
// 那張地形狹窄,原版資料就這樣。
//
// 所以這裡改成鎖住「16 張裡只有 1 張重疊」:位移一旦錯位,重疊的張數
// 會立刻暴增(讀到地形碼會撞出一堆相同值),這條仍然守得住門。
func TestCombatPartySlotsOverlapIsRare(t *testing.T) {
	s := loadCBT(t, "BRIT.CBT")
	overlapped := 0
	for i := range s.Maps {
		m := &s.Maps[i]
		seen := map[[2]byte]bool{}
		dup := false
		for k := 0; k < CombatPartySlots; k++ {
			p := [2]byte{m.PartyX[k], m.PartyY[k]}
			if seen[p] {
				dup = true
			}
			seen[p] = true
		}
		if dup {
			overlapped++
		}
	}
	if overlapped != 1 {
		t.Errorf("%d 張圖的隊員入場位置有重疊,預期只有 1 張(第 9 張)", overlapped)
	}
}

// TestDungeonEntryVariesByRoom:地牢的隊員入場與地表用同一組欄位,
// 而且**每個房間的入場方向不同**。
//
// 一開始我只看了第 0 張(整批 0)就寫成「地牢用另一套入場規則」——
// 錯了。第 1 張是 X=[5 6 4 5 6 4] Y=[8 9 9 10 10 10](從南邊進、排成三角),
// 第 4 張是 X=[3 2 2 1 0 0](從西邊進)。112 張裡有 33 張全 0,
// 那些是沒在用的房間,不是「欄位不適用」。
func TestDungeonEntryVariesByRoom(t *testing.T) {
	s := loadCBT(t, "DUNGEON.CBT")
	zero, shapes := 0, map[[2]byte]bool{}
	for i := range s.Maps {
		m := &s.Maps[i]
		if m.PartyX == [CombatPartySlots]byte{} && m.PartyY == [CombatPartySlots]byte{} {
			zero++
			continue
		}
		shapes[[2]byte{m.PartyX[0], m.PartyY[0]}] = true
	}
	if zero != 33 {
		t.Errorf("%d 張的隊員入場全是 0,預期 33", zero)
	}
	if len(shapes) < 4 {
		t.Errorf("有值的房間只有 %d 種入場位置,預期不只(每個房間的入口方向不同)", len(shapes))
	}
}
