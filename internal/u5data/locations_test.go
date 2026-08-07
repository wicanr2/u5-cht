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

// TestSceneMapping 驗證「地點 → 場景檔 + 索引」的對應(原版 sub_5C8 的規則)。
func TestSceneMapping(t *testing.T) {
	for i := range Locations {
		num := i + 1
		// 檔案 = (地點編號-1)/8
		if want := (num - 1) / 8; Locations[i].SceneFile != want {
			t.Errorf("地點 %d(%s)的檔案索引是 %d,依 (編號-1)/8 應為 %d",
				num, Locations[i].Name, Locations[i].SceneFile, want)
		}
		// 整個樓層範圍都必須落在每檔 16 張之內
		if lo := Locations[i].SceneIndex + Locations[i].FloorMin; lo < 0 {
			t.Errorf("地點 %d(%s)最低層算出索引 %d,小於 0", num, Locations[i].Name, lo)
		}
		if hi := Locations[i].SceneIndex + Locations[i].FloorMax; hi >= ScenesPerFile {
			t.Errorf("地點 %d(%s)最高層算出索引 %d,超出每檔 %d 張",
				num, Locations[i].Name, hi, ScenesPerFile)
		}
		if Locations[i].FloorMin > 0 {
			t.Errorf("地點 %d(%s)的最低層是 %+d —— 地面層(0)一定要在範圍內",
				num, Locations[i].Name, Locations[i].FloorMin)
		}
		if Locations[i].FloorMax < 0 {
			t.Errorf("地點 %d(%s)的最高層是 %+d —— 地面層(0)一定要在範圍內",
				num, Locations[i].Name, Locations[i].FloorMax)
		}
	}
}

// TestSceneMapsFullyPartitioned:四個場景檔各 16 張地圖,必須**恰好**被 8 個地點的
// 樓層範圍蓋滿 —— 不重疊也不留空。這是整張表最強的一致性檢查:任何一筆樓層數寫錯,
// 都會在這裡露出縫或疊到別人身上。
func TestSceneMapsFullyPartitioned(t *testing.T) {
	for f := range SceneFiles {
		used := make([]int, ScenesPerFile)
		for i := range Locations {
			if Locations[i].SceneFile != f {
				continue
			}
			for fl := Locations[i].FloorMin; fl <= Locations[i].FloorMax; fl++ {
				idx := Locations[i].SceneIndex + fl
				if idx < 0 || idx >= ScenesPerFile {
					continue // 上一個測試已經報過
				}
				used[idx]++
			}
		}
		for idx, n := range used {
			switch {
			case n == 0:
				t.Errorf("%s 第 %d 張地圖沒有任何地點宣告使用", SceneFiles[f], idx)
			case n > 1:
				t.Errorf("%s 第 %d 張地圖被 %d 個地點同時宣告", SceneFiles[f], idx, n)
			}
		}
	}
}

// TestKnownSceneTargets 固定幾個已用畫面驗收過的對應。
func TestKnownSceneTargets(t *testing.T) {
	cases := []struct {
		name     string
		file     int
		index    int
		lo, hi   int
		why      string
	}{
		{"BRITAIN", 0, 2, 0, 1, "TOWNE.DAT 索引 2,兩層"},
		{"MOONGLOW", 0, 0, 0, 1, "TOWNE.DAT 第一個"},
		{"YEW", 0, 7, -1, 0, "紫衫城的地下是監獄 —— 地圖排在地面層之前"},
		{"FOGSBANE", 1, 0, 0, 2, "燈塔三層 —— 畫面驗收過:底層有家具、二層剩塔身、頂層是圓形燈室"},
		{"IOLO'S HUT", 1, 12, 0, 0, "小屋只有一層"},
		{"SERPENT'S HOLD", 3, 14, -1, 1, "地面層索引 14,但地圖 13 是它的地下層"},
	}
	for _, c := range cases {
		var loc *Location
		for i := range Locations {
			if Locations[i].Name == c.name {
				loc = &Locations[i]
				break
			}
		}
		if loc == nil {
			t.Errorf("地點表裡沒有 %s", c.name)
			continue
		}
		if loc.SceneFile != c.file || loc.SceneIndex != c.index ||
			loc.FloorMin != c.lo || loc.FloorMax != c.hi {
			t.Errorf("%s → 檔案 %d 索引 %d 樓層 %+d..%+d,預期 %d/%d/%+d..%+d(%s)",
				c.name, loc.SceneFile, loc.SceneIndex, loc.FloorMin, loc.FloorMax,
				c.file, c.index, c.lo, c.hi, c.why)
		}
	}
}

// TestSceneOffset:原版是 (索引 + 樓層) << 10,且樓層 > 0x7F 視為負(地下層)。
func TestSceneOffset(t *testing.T) {
	l := Location{SceneIndex: 2}
	if got := l.SceneOffset(0); got != 2*SceneTiles {
		t.Errorf("地面層位移 %d,預期 %d", got, 2*SceneTiles)
	}
	if got := l.SceneOffset(1); got != 3*SceneTiles {
		t.Errorf("第二層位移 %d,預期 %d", got, 3*SceneTiles)
	}
	// 0xFF 代表 -1(地下一層)
	if got := l.SceneOffset(0xFF); got != 1*SceneTiles {
		t.Errorf("地下一層位移 %d,預期 %d —— 樓層 >0x7F 要當負數", got, 1*SceneTiles)
	}
}
