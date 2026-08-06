package u5data

import "testing"

// TestLocationsMatchWorldMap 是地點表的 oracle:
// 座標必須落在已知的入口 tile 上。這條同時驗證了三張平行表沒有錯位。
func TestLocationsCount(t *testing.T) {
	named := 0
	for i := range Locations {
		if Locations[i].Name != "" {
			named++
		}
	}
	// 實測:32 筆裡 27 筆有名字,13–17 之中有 5 筆無名(與 sub_10928 的判斷吻合)
	if named != 27 {
		t.Errorf("有名字的地點 %d 筆,實測應為 27", named)
	}
	for i := 13; i <= 17; i++ {
		if Locations[i].Name != "" {
			t.Errorf("索引 %d 應該無名(sub_10928 裡 i<13||i>17 才印名字),卻是 %q", i, Locations[i].Name)
		}
	}
}

func TestLocationAt(t *testing.T) {
	// 不列顛城在 (81,106) —— 從原版執行檔 dump 出來的
	loc, ok := LocationAt(81, 106)
	if !ok {
		t.Fatal("(81,106) 應該是不列顛城")
	}
	if loc.Name != "BRITAIN" {
		t.Errorf("(81,106) 是 %q,預期 BRITAIN", loc.Name)
	}
	if loc.DisplayName() != "不列顛城" {
		t.Errorf("顯示名 %q,預期中文譯名", loc.DisplayName())
	}
	if _, ok := LocationAt(0, 0); ok {
		t.Error("(0,0) 不該有地點")
	}
}

// TestLocationDisplayNameFallsBack:未定案的譯名要退回英文,不能顯示空白。
func TestLocationDisplayNameFallsBack(t *testing.T) {
	l := Location{Name: "COVE"}
	if l.DisplayName() != "COVE" {
		t.Errorf("沒有中文時應退回英文,實得 %q", l.DisplayName())
	}
	if (&Location{}).DisplayName() != "?" {
		t.Error("兩者皆空時應回 ?")
	}
}

// TestEightVirtueCitiesHaveChineseNames:八德城市的譯名已定案(對齊聖者之書體系)。
func TestEightVirtueCitiesHaveChineseNames(t *testing.T) {
	want := map[string]string{
		"MOONGLOW": "月光城", "BRITAIN": "不列顛城", "JHELOM": "哲倫", "YEW": "紫衫城",
		"MINOC": "米諾克", "TRINSIC": "特林希克", "SKARA BRAE": "史卡拉布雷",
		"NEW MAGINCIA": "新馬精西亞",
	}
	got := map[string]string{}
	for i := range Locations {
		if zh, ok := want[Locations[i].Name]; ok {
			got[Locations[i].Name] = Locations[i].NameZH
			if Locations[i].NameZH != zh {
				t.Errorf("%s 的譯名是 %q,應為 %q", Locations[i].Name, Locations[i].NameZH, zh)
			}
		}
	}
	if len(got) != len(want) {
		t.Errorf("八德城市只找到 %d 個,預期 %d", len(got), len(want))
	}
}
