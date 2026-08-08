package game

import (
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// TestWindowOriginAlignsToSixteen —— `sub_2CBEC` 的兩行。
//
//	原點 = 座標 & 0xF0;若 (座標 & 0x0F) < 8 再退一塊
//
// ⇒ 隊伍在視窗裡的位置永遠落在 8..23(視窗 32 格寬)。
func TestWindowOriginAlignsToSixteen(t *testing.T) {
	for _, tc := range []struct{ v, want int }{
		{0x00, 0xF0}, // ★ 低 4 bit = 0 < 8 ⇒ 退一塊,而且會環繞到 0xF0
		{0x07, 0xF0},
		{0x08, 0x00}, // 後半 ⇒ 不退
		{0x0F, 0x00},
		{0x50, 0x40},
		{0x58, 0x50},
		{0x87, 0x70},
		{0xFF, 0xF0},
	} {
		if got := alignLoadWindow(tc.v); got != tc.want {
			t.Errorf("座標 0x%02X → 原點 0x%02X,預期 0x%02X", tc.v, got, tc.want)
		}
	}
	// 導出的性質:隊伍在視窗裡的位置永遠在 8..23。
	for v := 0; v < 256; v++ {
		in := (v - alignLoadWindow(v)) & u5data.SpawnWindowSpan
		if in < 8 || in > 23 {
			t.Fatalf("座標 0x%02X 在視窗裡的位置是 %d,超出 8..23", v, in)
		}
	}
}

// TestWindowOnlyScrollsAtTheEdge —— ★★ 一次捲半個視窗,不是跟著走一格。
//
// 原版 `sub_2D014`:隊伍在視窗內的座標落在 5..0x1A 就什麼都不做,
// 否則原點一次跳 16 格。這就是大地圖走到邊緣「畫面整塊跳一下」的來源。
func TestWindowOnlyScrollsAtTheEdge(t *testing.T) {
	var s State
	s.X, s.Y = 0x58, 0x58
	s.resetLoadWindow()
	ox, oy := s.WindowX, s.WindowY
	if ox != 0x50 || oy != 0x50 {
		t.Fatalf("原點 (0x%02X,0x%02X),預期 (0x50,0x50)", ox, oy)
	}

	// 往東走到還在緩衝區內 —— 原點不動。
	for s.X < ox+u5data.SpawnWindowSpan-WindowScrollMargin {
		s.X++
		s.scrollLoadWindow(1, 0)
		if s.WindowX != ox {
			t.Fatalf("隊伍 X=0x%02X 時原點就捲到 0x%02X 了(在視窗內第 %d 格)",
				s.X, s.WindowX, (s.X-ox)&u5data.SpawnWindowSpan)
		}
	}
	// 再走一步就踩出緩衝區 ⇒ 原點跳 16。
	s.X++
	s.scrollLoadWindow(1, 0)
	if s.WindowX != ox+WindowAlign {
		t.Errorf("踩出緩衝區之後原點是 0x%02X,預期 0x%02X(跳 %d 格)",
			s.WindowX, ox+WindowAlign, WindowAlign)
	}
	if s.WindowY != oy {
		t.Errorf("只往東走,Y 的原點卻動了:0x%02X → 0x%02X", oy, s.WindowY)
	}
}

// TestWindowScrollWrapsAtTheMapEdge —— 原點用 `& 0xF0` 收尾 ⇒ 會環繞。
//
// 世界地圖是 256×256 環形的,而 `and al, 0F0h` 對 8 位元運算 ⇒ 原點自然環繞。
func TestWindowScrollWrapsAtTheMapEdge(t *testing.T) {
	var s State
	s.WindowX, s.WindowY = 0xF0, 0xF0
	s.X, s.Y = 0xFF, 0xFF // 在視窗裡的位置是 15 —— 還在緩衝區內
	s.scrollLoadWindow(1, 1)
	if s.WindowX != 0xF0 {
		t.Fatalf("還在緩衝區內就捲了")
	}
	// 走到視窗內第 27 格(> 0x1A)就該捲,而 0xF0 + 0x10 = 0x100 → 0x00。
	s.X, s.Y = 0x0B, 0x0B
	s.scrollLoadWindow(1, 1)
	if s.WindowX != 0x00 || s.WindowY != 0x00 {
		t.Errorf("原點 (0x%02X,0x%02X),預期環繞到 (0x00,0x00)", s.WindowX, s.WindowY)
	}
}

// TestCullingUsesTheWindowNotTheParty —— ★★ 清場的判準是視窗原點。
//
// 這是「真視窗」與「隊伍為中心」兩種寫法**可觀察的差別**:視窗不捲的那些回合,
// 怪走遠了也不會被清掉;近似版會立刻清。原版 `sub_2E24` 尾段比的是原點。
func TestCullingUsesTheWindowNotTheParty(t *testing.T) {
	s := overworldScene(t)
	if s.BaseSave == nil {
		t.Skip("沒有底稿存檔")
	}
	clearObjects(t, s)
	s.X, s.Y = 0x58, 0x58
	s.resetLoadWindow() // 原點 (0x50,0x50)

	// (a) 落在視窗內 ⇒ 留著。原點 + 0x1F 是最遠的一格。
	putObject(t, s, 5, 0xC0, s.WindowX+u5data.SpawnWindowSpan, s.WindowY)
	// (b) 再遠一格 ⇒ 清掉。
	putObject(t, s, 6, 0xC0, s.WindowX+u5data.SpawnWindowSpan+1, s.WindowY)
	// (c) 不是生物的東西**不清**,不管多遠 —— 玩家停在岸邊的船不能憑空消失。
	putObject(t, s, 7, u5data.VehicleShip, s.WindowX+100, s.WindowY+100)
	// (d) 槽 0 不掃(原版 `while (edi > 0)`)。
	putObject(t, s, 0, 0xC0, s.WindowX+100, s.WindowY+100)

	s.cullDistantCreatures()
	set := s.currentObjects()
	if set.Objects[5].Raw[u5data.ObjKind] == 0 {
		t.Error("視窗邊界上那一格被清掉了 —— 條件是 `> 0x1F` 不是 `>=`")
	}
	if set.Objects[6].Raw[u5data.ObjKind] != 0 {
		t.Error("超出視窗的怪沒被清掉")
	}
	if set.Objects[7].Raw[u5data.ObjKind] == 0 {
		t.Error("船被清掉了 —— 原版只清生物(`sub_22B0`)")
	}
	if set.Objects[0].Raw[u5data.ObjKind] == 0 {
		t.Error("槽 0 被清掉了 —— 原版的迴圈是 `while (edi > 0)`,掃不到槽 0")
	}
}

// TestSpawnUsesTheRealWindowOrigin —— 生怪落點的基準換成真原點了。
func TestSpawnUsesTheRealWindowOrigin(t *testing.T) {
	var s State
	s.X, s.Y = 0x58, 0x58
	s.resetLoadWindow()
	ox, oy := s.spawnWindowOrigin()
	if ox != s.WindowX || oy != s.WindowY {
		t.Errorf("生怪原點 (0x%02X,0x%02X) 與視窗原點 (0x%02X,0x%02X) 不一致",
			ox, oy, s.WindowX, s.WindowY)
	}
	// 而且**不是**隊伍為中心 —— 那是此前的近似。
	if ox == s.X-u5data.SpawnWindowSpan/2 && oy == s.Y-u5data.SpawnWindowSpan/2 {
		t.Error("還是「隊伍為中心」的近似")
	}
}

// TestCullingJudgesByKindNotTile —— ⚠⚠ 位移 0 而不是位移 1。
//
// 原版是 `movzx eax, byte ptr dword_3E46C[edi*8]`(**沒有 `+1`**)⇒ 吃 `ObjKind`。
// 第一版寫成 `ObjTile`,而 `Spawn` 與 `putObject` 都把兩欄設成同一個值,
// 所以當時的測試完全抓不到。這一條把兩欄**故意設成不同**來釘住位移。
func TestCullingJudgesByKindNotTile(t *testing.T) {
	s := overworldScene(t)
	if s.BaseSave == nil {
		t.Skip("沒有底稿存檔")
	}
	clearObjects(t, s)
	s.X, s.Y = 0x58, 0x58
	s.resetLoadWindow()
	set := s.currentObjects()
	far := func(slot int, kind, tile byte) {
		o := &set.Objects[slot]
		*o = u5data.MapObject{Kind: kind, Tile: tile,
			X: s.WindowX + 100, Y: s.WindowY + 100, Floor: s.Floor}
		o.Raw[u5data.ObjKind] = kind
		o.Raw[u5data.ObjTile] = tile
		o.Raw[u5data.ObjX] = byte(o.X)
		o.Raw[u5data.ObjY] = byte(o.Y)
	}
	// (a) 種類碼是船(0x24,不是生物)、圖是 Orc(0xC0,是生物)→ **不該清**。
	far(5, u5data.VehicleShip, 0xC0)
	// (b) 種類碼是 Orc、圖是船 → **該清**。
	far(6, 0xC0, u5data.VehicleShip)

	s.cullDistantCreatures()
	if set.Objects[5].Raw[u5data.ObjKind] == 0 {
		t.Error("種類碼是船卻被清掉了 —— 判準讀成 ObjTile 了")
	}
	if set.Objects[6].Raw[u5data.ObjKind] != 0 {
		t.Error("種類碼是 Orc 卻沒被清 —— 判準讀成 ObjTile 了")
	}
}

// TestVehiclesOnTheMapAreNeverCulled —— ★ 玩家停在外面的載具不能消失。
//
// 靠的是 `sub_22B0` 對 `< 0x2C` 回 0。原版 `sub_2DD44` 把載具放回大地圖時寫的
// 種類碼是 **0x25(船)/ 0x29(小艇)**,引擎的 `dismountShip` 用 `Transport`
// (0x24..0x2B)—— 兩者都落在同一區。馬(0x10/0x11)與魔毯(0x14)也一樣。
//
// ⚠ 反面:**敵船(0x2C..0x2F)是生物,會被清** —— 那是對的,牠們是敵人。
func TestVehiclesOnTheMapAreNeverCulled(t *testing.T) {
	s := overworldScene(t)
	if s.BaseSave == nil {
		t.Skip("沒有底稿存檔")
	}
	for _, tc := range []struct {
		kind byte
		name string
		cull bool
	}{
		{u5data.VehicleShip, "大船 0x24", false},
		{u5data.VehicleShip + 1, "大船 0x25(原版放回地圖用的值)", false},
		{u5data.VehicleSkiff, "小艇 0x28", false},
		{u5data.VehicleSkiff + 1, "小艇 0x29(原版放回地圖用的值)", false},
		{u5data.TileHorse, "馬 0x10", false},
		{u5data.TileCarpetObj, "魔毯", false},
		{u5data.SpawnEnemyShip, "敵船 0x2C", true},
		{0xC0, "Orc", true},
	} {
		clearObjects(t, s)
		s.X, s.Y = 0x58, 0x58
		s.resetLoadWindow()
		putObject(t, s, 5, tc.kind, s.WindowX+100, s.WindowY+100)
		s.cullDistantCreatures()
		gone := s.currentObjects().Objects[5].Raw[u5data.ObjKind] == 0
		if gone != tc.cull {
			verb := "被清掉了"
			if !gone {
				verb = "沒被清"
			}
			t.Errorf("%s(0x%02X)%s,預期 cull=%v", tc.name, tc.kind, verb, tc.cull)
		}
	}
}
