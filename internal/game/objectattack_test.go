package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// putObject 把一個東西放進物件槽(測試用)。
func putObject(t *testing.T, s *State, slot int, kind byte, x, y int) {
	t.Helper()
	set := s.currentObjects()
	if set == nil {
		t.Fatal("沒有物件表")
	}
	o := &set.Objects[slot]
	*o = u5data.MapObject{Kind: kind, Tile: kind, X: x, Y: y, Floor: s.Floor}
	o.Raw[u5data.ObjKind] = kind
	o.Raw[u5data.ObjTile] = kind
	o.Raw[u5data.ObjX] = byte(x)
	o.Raw[u5data.ObjY] = byte(y)
}

// TestAdjacencyIsOrthogonalOnly —— ★ 斜對角不算相鄰。
//
// 原版判準是 `(dx==1 && dy==0) || (dx==0 && dy==1)`。寫成「切比雪夫距離 1」
// 會讓怪物比原版兇一圈 —— 而那在遊玩時只覺得難,查不出來。
func TestAdjacencyIsOrthogonalOnly(t *testing.T) {
	for _, tc := range []struct {
		dx, dy int
		want   bool
		why    string
	}{
		{1, 0, true, "正東"},
		{-1, 0, true, "正西"},
		{0, 1, true, "正南"},
		{0, -1, true, "正北"},
		{1, 1, false, "★ 斜對角不算"},
		{-1, -1, false, "★ 斜對角不算"},
		{2, 0, false, "兩格外"},
		{0, 0, false, "同一格"},
	} {
		s := overworldScene(t)
		// 用 Orc:沒有遠程攻擊,所以「動手了」只可能來自相鄰那一條。
		putObject(t, s, 3, 0xC0, s.X+tc.dx, s.Y+tc.dy)
		s.Messages = nil
		got := s.objectAttacks(3)
		if got != tc.want {
			t.Errorf("(%+d, %+d) 動手 = %v,預期 %v(%s)", tc.dx, tc.dy, got, tc.want, tc.why)
		}
	}
}

// TestOnlySeaSerpentAndDragonShootFromAfar —— 只有兩種有遠程,而且射程是方形。
func TestOnlySeaSerpentAndDragonShootFromAfar(t *testing.T) {
	// ⚠ **同一個 scene 重複擲**,不要每輪重建 —— `Roll` 用固定種子,
	// 重建等於重播同一顆骰子,1/8 的事件會永遠不發生。
	// (第一版就這樣寫,紅燈看起來像「射程算錯」。)
	s := overworldScene(t)
	fires := func(kind byte, dx, dy int) bool {
		for i := 0; i < 400; i++ {
			putObject(t, s, 3, kind, s.X+dx, s.Y+dy)
			s.Messages = nil
			if s.objectAttacks(3) {
				return true
			}
		}
		return false
	}
	// 海蛇與龍在射程內會開火(1/8,400 次一定中)。
	for _, kind := range []byte{RangedSeaSerpent, RangedDragon} {
		if !fires(kind, 3, 3) {
			t.Errorf("kind 0x%02X 在 (3,3) 不開火 —— 射程是兩軸各自 ≤ 3", kind)
		}
		// ★ 方形範圍:(3,3) 在射程內,(4,0) 不在。
		if fires(kind, 4, 0) {
			t.Errorf("kind 0x%02X 在 (4,0) 開火了 —— 射程只到 3", kind)
		}
	}
	// ★ 其餘的怪沒有遠程 —— Orc 站在 (3,3) 什麼都不做。
	if fires(0xC0, 3, 3) {
		t.Error("Orc 也有遠程攻擊了 —— 原版只有海蛇(0x88)與龍(0xDC)")
	}
}

// TestRangedFireIsOneInEight —— 1/8,不是「一定開火」。
func TestRangedFireIsOneInEight(t *testing.T) {
	s := overworldScene(t)
	putObject(t, s, 3, RangedDragon, s.X+3, s.Y+3)
	hits := 0
	const tries = 4000
	for i := 0; i < tries; i++ {
		if s.objectAttacks(3) {
			hits++
		}
	}
	rate := float64(hits) / tries
	if rate < 0.09 || rate > 0.16 {
		t.Errorf("開火率 %.1f%%,原版 1/8 = 12.5%%", rate*100)
	}
}

// TestWhirlpoolSucksYouIntoTheUnderworld —— ★★ 0xEC 是漩渦。
//
// 三重佐證見 `docs/re/83`:組語字串 `"WHIRLPOOL!"`、`look#492` =「漩渦」
// (物件種類碼 +256 = tile 0x1EC)、以及 NPC 對白 `DWELLING.TLK#2#e11`。
func TestWhirlpoolSucksYouIntoTheUnderworld(t *testing.T) {
	s := overworldScene(t)
	s.Transport = u5data.VehicleShip // 在船上
	putObject(t, s, 3, WhirlpoolKind, s.X+1, s.Y)
	s.Messages = nil
	if !s.objectAttacks(3) {
		t.Fatal("貼著漩渦什麼都沒發生")
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgWhirlpool) {
		t.Errorf("沒印漩渦:%q", s.Messages)
	}
	if s.Floor != UnderworldFloor {
		t.Errorf("沒被吸到幽冥界,樓層是 %d", s.Floor)
	}
	if s.X != WhirlpoolExitX || s.Y != WhirlpoolExitY {
		t.Errorf("落點是 (%d, %d),原版寫死 (%d, %d)", s.X, s.Y, WhirlpoolExitX, WhirlpoolExitY)
	}
	// ★ 漩渦是一次性的 —— 吸完就消失。
	if s.currentObjects() != nil {
		for i := range s.currentObjects().Objects {
			if s.currentObjects().Objects[i].Raw[u5data.ObjKind]&0xFC == WhirlpoolKind {
				t.Error("漩渦吸完還留在場上 —— 原版把種類碼與 tile 都歸零")
			}
		}
	}
}

// TestWhirlpoolDoesNotTouchYouOnFoot —— 步行時漩渦碰不到你。
func TestWhirlpoolDoesNotTouchYouOnFoot(t *testing.T) {
	s := overworldScene(t)
	s.Transport = u5data.VehicleWalk
	before := s.Floor
	putObject(t, s, 3, WhirlpoolKind, s.X+1, s.Y)
	s.objectAttacks(3)
	if s.Floor != before {
		t.Error("步行也被漩渦吸走了 —— 原版 `cmp byte_3E08C, 1Ch; jz` 先擋掉")
	}
}

// TestSandTrapOnlyDamagesTheShip —— 流沙陷阱族貼上來不開戰。
func TestSandTrapOnlyDamagesTheShip(t *testing.T) {
	s := overworldScene(t)
	putObject(t, s, 3, SandTrapKind, s.X+1, s.Y)
	if !s.objectAttacks(3) {
		t.Fatal("流沙陷阱貼上來什麼都沒發生")
	}
	if s.InCombat() {
		t.Error("流沙陷阱貼上來開戰了 —— 原版只走 `sub_22F0`(船損)")
	}
}

// TestSkiffTakesHullDamageInsteadOfCombat —— ★ 在小艇上被貼身不會開戰。
//
// 原版在水上(腳下 tile < 4)分三條:魔毯與小艇只損船,其餘開戰。
// 寫成「一律開戰」會讓小艇變成活棺材。
func TestSkiffTakesHullDamageInsteadOfCombat(t *testing.T) {
	// 腳下要是水(tile < 4),否則走的是「陸地一律開戰」那條。
	for _, tc := range []struct {
		name      string
		transport byte
		wantFight bool
	}{
		{"小艇", u5data.VehicleSkiff, false},
		{"魔毯", 0x14, false},
		{"大船", u5data.VehicleShip, true},
	} {
		s := overworldScene(t)
		if !s.SetTileAt(s.X, s.Y, u5data.RoughSeasTile) {
			t.Fatal("寫不進世界地圖")
		}
		s.Transport = tc.transport
		putObject(t, s, 3, 0xC0, s.X+1, s.Y)
		s.objectAttacks(3)
		if got := s.InCombat(); got != tc.wantFight {
			t.Errorf("%s:開戰 = %v,預期 %v", tc.name, got, tc.wantFight)
		}
	}
}

// TestOnLandItAlwaysFights —— 腳下是陸地就開戰,不管載具。
func TestOnLandItAlwaysFights(t *testing.T) {
	s := overworldScene(t)
	if !s.SetTileAt(s.X, s.Y, tileGrass) {
		t.Fatal("寫不進世界地圖")
	}
	s.Transport = u5data.VehicleSkiff // 就算坐在小艇上
	putObject(t, s, 3, 0xC0, s.X+1, s.Y)
	s.objectAttacks(3)
	if !s.InCombat() {
		t.Error("陸地上被貼身沒開戰 —— 原版 `cmp byte ptr [eax], 4; jnb → 開戰`")
	}
}

// TestWorldTurnLetsAdjacentCreaturesAct —— 接線測試:世界回合真的會叫 objectAttacks。
//
// ⚠ 這一條測的是**入口**。本專案已經五次遇到「規則寫好了卻沒人叫」
// (`docs/re/80`),所以每一條新規則都要有一條測入口的。
func TestWorldTurnLetsAdjacentCreaturesAct(t *testing.T) {
	s := overworldScene(t)
	if !s.SetTileAt(s.X, s.Y, tileGrass) {
		t.Fatal("寫不進世界地圖")
	}
	putObject(t, s, 3, 0xC0, s.X+1, s.Y)
	s.Messages = nil
	s.extraWorldTurn()
	if !s.InCombat() {
		t.Error("世界回合沒有讓貼著隊伍的怪動手 —— objectAttacks 沒接上")
	}
}

// TestEnemyShipsFireBroadsideEveryTurnWhenAligned —— ★★ 敵船開砲沒有機率閘門。
//
// 條件是「正交同線且相隔 1..3 格」;海蛇與龍是 1/8,敵船是**每回合都開**。
func TestEnemyShipsFireBroadsideEveryTurnWhenAligned(t *testing.T) {
	s := overworldScene(t)
	fire := func(dx, dy int) bool {
		clearObjects(t, s)
		flatGrass(t, s, 8)
		putObject(t, s, 5, u5data.SpawnEnemyShip, s.X+dx, s.Y+dy)
		s.ShipHull = 100
		return s.objectAttacks(5)
	}
	for _, tc := range []struct {
		dx, dy int
		want   bool
		why    string
	}{
		{3, 0, true, "正東 3 格"},
		{-3, 0, true, "正西 3 格"},
		{0, 3, true, "正南 3 格"},
		{0, -2, true, "正北 2 格"},
		{4, 0, false, "正東 4 格 —— 條件是嚴格 < 4"},
		{0, 4, false, "正南 4 格"},
		{2, 2, false, "斜對角 —— 要正交同線"},
		{1, 3, false, "不同線"},
	} {
		if got := fire(tc.dx, tc.dy); got != tc.want {
			t.Errorf("%s:開砲 = %v,預期 %v", tc.why, got, tc.want)
		}
	}
	// 沒有 1/8 那種閘門 —— 同一個位置連開十次都要成立。
	for i := 0; i < 10; i++ {
		if !fire(2, 0) {
			t.Fatalf("第 %d 次沒開砲 —— 敵船不該有機率閘門", i+1)
		}
	}
}

// TestBroadsideTurnChangesTheTileNotTheKind —— ⚠ 只改船圖,不動種類碼。
//
// 原版 `sub_23FC` 寫的是位移 1(`ObjTile`),而 `sub_2870`(移動時轉向)
// **兩個都寫**。差別是可觀察的:風速表查 `ObjKind` ⇒ 側身開砲不影響船速。
func TestBroadsideTurnChangesTheTileNotTheKind(t *testing.T) {
	s := overworldScene(t)
	clearObjects(t, s)
	flatGrass(t, s, 8)
	// 船圖朝北(南北向),玩家在正東 ⇒ dy == 0 ⇒ 不該轉(已經能打東西向嗎?)
	// 先驗 dx == 0 那條:玩家在正南,船身南北向 ⇒ 轉成東西向。
	putObject(t, s, 5, u5data.SpawnEnemyShip, s.X, s.Y+2)
	set := s.currentObjects()
	set.Objects[5].Raw[u5data.ObjTile] = u5data.ShipTileBase + u5data.ShipFacingN
	kindBefore := set.Objects[5].Raw[u5data.ObjKind]

	s.turnBroadside(5, 0, 2)
	got := set.Objects[5].Raw[u5data.ObjTile]
	east := byte(u5data.ShipTileBase + u5data.ShipFacingE)
	west := byte(u5data.ShipTileBase + u5data.ShipFacingW)
	if got != east && got != west {
		t.Errorf("船圖是 0x%02X,預期東(0x%02X)或西(0x%02X)", got, east, west)
	}
	if set.Objects[5].Raw[u5data.ObjKind] != kindBefore {
		t.Errorf("種類碼被改成 0x%02X 了 —— 原版 `sub_23FC` 只寫位移 1",
			set.Objects[5].Raw[u5data.ObjKind])
	}

	// 已經是東西向、玩家又在正南 ⇒ 不動(原版兩個 if 都不成立)。
	set.Objects[5].Raw[u5data.ObjTile] = east
	s.turnBroadside(5, 0, 2)
	if set.Objects[5].Raw[u5data.ObjTile] != east {
		t.Error("船身已經對得到目標了還轉 —— 原版只在需要時轉")
	}
}
