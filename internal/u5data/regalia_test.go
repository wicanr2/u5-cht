package u5data

import (
	"os"
	"testing"
)

// 王冠與權杖必須真的躺在 `.NPC` 檔裡宣稱的那兩槽上。
//
// 這條測試的價值不在「常數有沒有打錯」,而在**證明我沒有再一次找錯命名空間**:
// 它同時要求「這兩槽是 0xB5 / 0xB6」與「全遊戲只有這兩槽是」。
// 只要哪天有人把地形的加農砲(同樣是 0xB4..0xB7)混進來,第二個條件就會紅。
func TestRegaliaSitWhereTheNPCFilesSayTheyDo(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	set, err := LoadNPCSet(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range RegaliaNPCPlacement {
		slots, err := set.At(want.Location)
		if err != nil {
			t.Fatalf("%s:地點 %d 讀不到:%v", want.Name, want.Location, err)
		}
		n := &slots[want.Slot]
		if n.Creature != want.Kind {
			t.Errorf("%s:地點 %d 槽 %d 的生物編號是 0x%02X,預期 0x%02X",
				want.Name, want.Location, want.Slot, n.Creature, want.Kind)
			continue
		}
		// 三個 slot 都指同一格 —— 它不會換崗位。
		for k := 0; k < 3; k++ {
			if int(n.Schedule.X[k]) != want.X || int(n.Schedule.Y[k]) != want.Y ||
				int(n.Schedule.Floor[k]) != want.Floor {
				t.Errorf("%s:排程 slot %d 是 (%d,%d) 第 %d 層,預期 (%d,%d) 第 %d 層",
					want.Name, k, n.Schedule.X[k], n.Schedule.Y[k], n.Schedule.Floor[k],
					want.X, want.Y, want.Floor)
			}
			// 行為型別 0 = 原地不動。信物會走路的話就不是照抄了。
			if n.Schedule.AI[k] != NPCAIFixed {
				t.Errorf("%s:排程 slot %d 的行為型別是 %d,預期 %d(不動)",
					want.Name, k, n.Schedule.AI[k], NPCAIFixed)
			}
		}
		// 四個時刻全 0 → `Scheduled()` 為假,暗影君主與「叫人滾開」都不會挑到它。
		if n.Schedule.Scheduled() {
			t.Errorf("%s:竟然有作息時刻 %v", want.Name, n.Schedule.Times)
		}
		// 對話號碼 0 = 沒有對話(`sub_1B52C` 的 "No response!")。
		if n.Dialogue != 0 {
			t.Errorf("%s:對話號碼是 %d,預期 0", want.Name, n.Dialogue)
		}
	}

	// 全遊戲只有那兩槽 —— 這才是「找對命名空間」的證明。
	found := map[byte][]string{}
	for num := 1; num <= len(Locations); num++ {
		slots, err := set.At(num)
		if err != nil {
			t.Fatal(err)
		}
		for i := range slots {
			switch slots[i].Creature {
			case ItemCrown, ItemSceptre, ItemAmulet, ItemShard:
				found[slots[i].Creature] = append(found[slots[i].Creature],
					locSlotLabel(num, i))
			}
		}
	}
	for kind, want := range map[byte]int{
		ItemCrown: 1, ItemSceptre: 1,
		// 護符與碎片不在 `.NPC` 裡 —— 它們由 `sub_10B3C` 當場塞進地下世界。
		ItemAmulet: 0, ItemShard: 0,
	} {
		if got := len(found[kind]); got != want {
			t.Errorf("生物編號 0x%02X 在 `.NPC` 裡出現 %d 次(%v),預期 %d 次",
				kind, got, found[kind], want)
		}
	}
}

// locSlotLabel 只是給錯誤訊息用的短標籤。
func locSlotLabel(num, slot int) string {
	return string(rune('0'+num/10)) + string(rune('0'+num%10)) + "#" +
		string(rune('0'+slot/10)) + string(rune('0'+slot%10))
}

// 地形的 0xB4..0xB7 是加農砲,與物件種類的信物是**兩套索引** ——
// 這一條把那個陷阱釘住,免得誰又拿地圖 tile 去找信物。
func TestRegaliaNumbersCollideWithCannonTerrain(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	look, err := LoadLook(dir)
	if err != nil {
		t.Fatal(err)
	}
	for kind := ItemShard; kind <= ItemAmulet; kind++ {
		terrain := look.Terrain(int(kind))
		object := look.Object(int(kind))
		if terrain == object {
			t.Errorf("0x%02X:地形與物件的敘述相同(%q)—— 兩套索引沒分開", kind, terrain)
		}
		if terrain != "a cannon" {
			t.Errorf("地形 0x%02X 的敘述是 %q,預期 \"a cannon\"", kind, terrain)
		}
	}
	if got := look.Object(ItemCrown); got != "the Crown!" {
		t.Errorf("物件 0x%02X 的敘述是 %q", ItemCrown, got)
	}
	if got := look.Object(ItemSceptre); got != "the Sceptre!" {
		t.Errorf("物件 0x%02X 的敘述是 %q", ItemSceptre, got)
	}
}
