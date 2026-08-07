package game

import (
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

func windState(t *testing.T) *State {
	t.Helper()
	s := magicState(t)
	wd, err := u5data.LoadWindDelay("../../gamedata")
	if err != nil {
		t.Fatal(err)
	}
	s.WindDelay = wd
	return s
}

// TestWindDelayTableIsSymmetric:延遲表的對稱性。
//
// 每個朝向都要正好「順風 1 種、側風 1 種、逆風 2 種」,而且南北向的船
// 只吃南北風、東西向的只吃東西風。位移偏一格這個對稱就垮了 ——
// 這比逐值比對硬。
func TestWindDelayTableIsSymmetric(t *testing.T) {
	s := windState(t)
	for f := 0; f < u5data.ShipFacings; f++ {
		var n [5]int
		for w := 1; w < u5data.WindCount; w++ {
			n[s.WindDelay.Delay(f, w)]++
		}
		if n[2] != 1 || n[3] != 1 || n[4] != 2 {
			t.Errorf("朝向 %d:順風 %d 種、側風 %d 種、逆風 %d 種,預期 1/1/2",
				f, n[2], n[3], n[4])
		}
	}
	// 每個朝向的順風就是它自己的方向。
	for _, c := range []struct {
		facing, wind int
		name         string
	}{
		{u5data.ShipFacingN, u5data.WindNorth, "朝北的船遇北風"},
		{u5data.ShipFacingS, u5data.WindSouth, "朝南的船遇南風"},
		{u5data.ShipFacingW, u5data.WindWest, "朝西的船遇西風"},
		{u5data.ShipFacingE, u5data.WindEast, "朝東的船遇東風"},
	} {
		if got := s.WindDelay.Delay(c.facing, c.wind); got != 2 {
			t.Errorf("%s 的延遲是 %d,預期 2(最快)", c.name, got)
		}
	}
	// ⚠ 「動不了」的是**橫向**的風,不是反向的。
	// 朝北的船遇南風是延遲 3(還能搶風調向),遇東西風才是 4。
	// 我第一版直覺寫成「南風把朝北的船擋死」,測試打臉 —— 表說的是
	// 「同一軸線的風都能用,一快一慢;垂直軸線的風完全沒用」。
	if got := s.WindDelay.Delay(u5data.ShipFacingN, u5data.WindSouth); got != 3 {
		t.Errorf("朝北的船遇南風延遲 %d,預期 3(同軸線,慢但走得動)", got)
	}
	for _, w := range []int{u5data.WindWest, u5data.WindEast} {
		if got := s.WindDelay.Delay(u5data.ShipFacingN, w); got != u5data.ShipNeverMoves {
			t.Errorf("朝北的船遇橫風 %d 延遲 %d,預期動不了", w, got)
		}
	}
}

// TestSailingAgainstTheWindStalls:逆風走不動、順風走得動。
//
// 這是 U5 航海的節奏,也是 Rel Hur 有用的原因。
func TestSailingAgainstTheWindStalls(t *testing.T) {
	s := windState(t)
	s.Location, s.Floor = 0, 0
	s.Transport = u5data.VehicleShip
	// 朝北走(dy = −1):北風順、南風逆。
	s.Wind = u5data.WindNorth
	moved := 0
	for i := 0; i < 30; i++ {
		if s.CanSail(0, -1) {
			moved++
		}
	}
	if moved == 0 {
		t.Error("順風 30 拍一步都走不了")
	}
	// 橫風才是真的動不了。
	s.Wind = u5data.WindEast
	s.windTimer = 0
	for i := 0; i < 30; i++ {
		if s.CanSail(0, -1) {
			t.Fatal("橫風竟然走得動")
		}
	}
	// 側風走得動但比順風慢。
	s.Wind = u5data.WindNorth
	s.windTimer = 0
	fast := 0
	for i := 0; i < 60; i++ {
		if s.CanSail(0, -1) {
			fast++
		}
	}
	// 同軸線的反向風走得動,但比順風慢(延遲 3 vs 2)。
	s.Wind = u5data.WindSouth
	s.windTimer = 0
	slow := 0
	for i := 0; i < 60; i++ {
		if s.CanSail(0, -1) { // 朝北 + 南風 = 同軸線但較慢
			slow++
		}
	}
	if fast == 0 || slow == 0 {
		t.Errorf("順風或側風走不動:順 %d、側 %d", fast, slow)
	}
	if slow >= fast {
		t.Errorf("側風走了 %d 步、順風 %d 步 —— 順風該比較快", slow, fast)
	}
}

// TestRelHurSetsTheWind:改風向把風轉到指定方向。
func TestRelHurSetsTheWind(t *testing.T) {
	s := windState(t)
	for _, c := range []struct {
		d    Direction
		want int
	}{
		{North, u5data.WindNorth},
		{South, u5data.WindSouth},
		{West, u5data.WindWest},
		{East, u5data.WindEast},
	} {
		s.ChangeWind(c.d)
		if s.Wind != c.want {
			t.Errorf("往 %s 改風向之後是 %d,預期 %d", c.d.Name(), s.Wind, c.want)
		}
	}
}

// TestRelHurAsksForDirection:Rel Hur 會問方向,而且只在野外施得動。
func TestRelHurAsksForDirection(t *testing.T) {
	s := windState(t)
	s.Location = 0
	relHur := s.Spells.Find("Rel Hur")
	s.Inventory.Spells[relHur] = 1
	s.Roster[0].Level, s.Roster[0].MP = 8, 40
	if got := s.Cast(0, relHur); got != MagicSuccess {
		t.Fatalf("野外施 Rel Hur 回傳 %v:\n%s", got, s.log())
	}
	if !s.AwaitingDirection() {
		t.Fatal("Rel Hur 沒有問方向")
	}
	s.AnswerDirection(West)
	if s.Wind != u5data.WindWest {
		t.Errorf("答了西之後風是 %d,預期 %d", s.Wind, u5data.WindWest)
	}
	// 城鎮裡施不動 —— 風只在海上有意義。
	s.Location = 2
	s.Inventory.Spells[relHur] = 1
	if got := s.Cast(0, relHur); got != MagicNotHere {
		t.Errorf("城鎮裡施 Rel Hur 回傳 %v,預期 MagicNotHere", got)
	}
}
