package u5data

import (
	"os"
	"testing"
)

func loadWorldObjects(t *testing.T) (*ObjectSet, *ObjectSet) {
	t.Helper()
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	sur, und, err := LoadWorldObjects(dir)
	if err != nil {
		t.Fatal(err)
	}
	return sur, und
}

// TestPartyObjectSlot:`BRIT.OOL` 的槽 0 是「步行的隊伍」。
//
// tile 值 28 與存檔的載具欄位是**兩個獨立檔案**讀出來的同一個數 ——
// 對得上,「+0 是種類」這個欄位位移就不是猜的。
func TestPartyObjectSlot(t *testing.T) {
	sur, _ := loadWorldObjects(t)
	party := &sur.Objects[PartyObjectSlot]
	if !party.Present() {
		t.Fatal("BRIT.OOL 的槽 0 是空的")
	}
	if party.Kind != TileWalking || party.Tile != TileWalking {
		t.Errorf("槽 0 是 kind=%d tile=%d,預期都是步行的隊伍 %d",
			party.Kind, party.Tile, TileWalking)
	}
	if party.X < 0 || party.X >= WorldSide || party.Y < 0 || party.Y >= WorldSide {
		t.Errorf("槽 0 在 (%d,%d),超出 %d×%d 的世界", party.X, party.Y, WorldSide, WorldSide)
	}
	dir := os.Getenv("U5_GAMEDATA")
	sv, err := LoadSave(dir + "/INIT.GAM")
	if err != nil {
		t.Fatal(err)
	}
	if sv.Transport != TileWalking {
		t.Errorf("存檔的載具是 %d,預期與物件表同樣是 %d", sv.Transport, TileWalking)
	}
	// 槽 0 站的那一格必須是地圖上真的有的地點(它是世界座標,不是場景座標)。
	if _, ok := LocationAt(party.X, party.Y); !ok {
		t.Errorf("槽 0 在 (%d,%d),那裡沒有任何地點", party.X, party.Y)
	}
}

// TestSaveObjectLayout:`SAVED.OOL` 是「地表 256 B + 地下 256 B」。
//
// 證據是後半與 `UNDER.OOL` 逐位元組相同,而前半全零。
// 256 B 的 `INIT.OOL` 則整份等於 `UNDER.OOL` —— 開新遊戲時地表沒有物件。
func TestSaveObjectLayout(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	_, und := loadWorldObjects(t)
	for _, name := range []string{"SAVED.OOL", "INIT.OOL"} {
		sur, u, err := LoadSaveObjects(dir + "/" + name)
		if err != nil {
			t.Fatalf("%s:%v", name, err)
		}
		for i := range sur.Objects {
			if sur.Objects[i].Present() {
				t.Errorf("%s 的地表半段第 %d 槽不是空的", name, i)
			}
		}
		if *u != *und {
			t.Errorf("%s 的地下半段與 UNDER.OOL 不同", name)
		}
	}
}

// TestUnderworldObjectsAreUnderground:地下世界那份的樓層欄是 −1。
//
// 樓層是**有號數**(0xFF = −1);當成無號讀會變 255,而 255 不會等於任何
// 真實樓層,結果是那些物件永遠畫不出來 —— 症狀看起來像「地下世界沒東西」。
func TestUnderworldObjectsAreUnderground(t *testing.T) {
	_, und := loadWorldObjects(t)
	n := 0
	for i := range und.Objects {
		o := &und.Objects[i]
		if !o.Present() {
			continue
		}
		n++
		if o.Floor != -1 {
			t.Errorf("UNDER.OOL 的槽 %d 樓層是 %d,預期 −1", i, o.Floor)
		}
	}
	if n == 0 {
		t.Error("UNDER.OOL 一個物件都沒有")
	}
}

// TestSpawnAndRemove:生成用的是空槽,而且不會佔到隊伍那一格。
func TestSpawnAndRemove(t *testing.T) {
	s := &ObjectSet{}
	s.Objects[PartyObjectSlot] = MapObject{Kind: TileWalking}
	slot, ok := s.Spawn(TileHorse, 10, 20, 0)
	if !ok {
		t.Fatal("空表裡也生不出東西")
	}
	if slot == PartyObjectSlot {
		t.Error("生成佔用了隊伍的槽")
	}
	o, found := s.At(10, 20, 0)
	if !found || o.Kind != TileHorse || o.Tile != TileHorse {
		t.Errorf("生成後在 (10,20) 找不到馬:%+v", o)
	}
	s.Remove(slot)
	if _, found := s.At(10, 20, 0); found {
		t.Error("移除之後還在")
	}
}

// TestSpawnRunsOut:32 槽用完就生不出來,不會覆蓋既有物件。
func TestSpawnRunsOut(t *testing.T) {
	s := &ObjectSet{}
	n := 0
	for {
		if _, ok := s.Spawn(TileHorse, n%32, n/32, 0); !ok {
			break
		}
		n++
		if n > ObjectSlots {
			t.Fatal("生成不會停 —— 空槽判斷壞了")
		}
	}
	if n != ObjectSlots-1 {
		t.Errorf("生了 %d 個就滿,預期 %d(槽 0 保留給隊伍)", n, ObjectSlots-1)
	}
}
