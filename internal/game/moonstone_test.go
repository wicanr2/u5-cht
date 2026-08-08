package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// TestMoonstoneBuryTerrainBoundaries 是這一輪最硬的一條:
// **九個 `look#` 名字必須同時對得上判準的上下界。**
//
// 判準是 `tile == 0x2C || tile == 0x2D || 4 <= tile <= 10`。若我把界讀錯一格,
// 下面就會有一格的「該不該」翻掉 —— 而那一格恰好是水或岩石,一眼看得出不對。
func TestMoonstoneBuryTerrainBoundaries(t *testing.T) {
	cases := []struct {
		tile byte
		name string
		want bool
	}{
		{0x02, "水", false},
		{0x03, "淺灘", false}, // ★ 下界外一格
		{0x04, "沼澤", true},  // ★ 下界
		{0x05, "草地", true},
		{0x06, "灌木", true},
		{0x07, "焦灼荒漠", true},
		{0x08, "灌木", true},
		{0x09, "樹林", true},
		{0x0A, "熱帶森林", true},  // ★ 上界
		{0x0B, "山麓", false},   // ★ 上界外一格
		{0x2B, "?", false},     // ★ 農地前一格
		{0x2C, "犁過的地", true}, // ★ 兩個單獨列的例外
		{0x2D, "豐收莊稼", true},
		{0x2E, "?", false}, // ★ 農地後一格
	}
	for _, c := range cases {
		if got := MoonstoneBuryable(c.tile); got != c.want {
			t.Errorf("tile 0x%02X(%s)埋得下去 = %v,預期 %v", c.tile, c.name, got, c.want)
		}
	}
}

// TestBuryingAMoonstoneRecordsFourFields 釘住原版那四行連寫。
//
// ⚠ 這條同時是存檔格式更正的守門員:若哪天有人把 `Moonstone` 改回一個布林值,
// 「埋在哪裡」就沒地方放,這條會編不過。
func TestBuryingAMoonstoneRecordsFourFields(t *testing.T) {
	s := upkeepScene(t)
	s.Inventory.Moonstones[2] = u5data.Moonstone{Location: u5data.MoonstoneInHand}
	// 站在一格草地上。
	if !s.SetTileAt(s.X, s.Y, 0x05) {
		t.Skip("這一層地圖不支援寫入")
	}
	s.Messages = nil
	if !s.BuryMoonstone(2) {
		t.Fatalf("草地上該埋得下去:\n%s", strings.Join(s.Messages, "\n"))
	}
	got := s.Inventory.Moonstones[2]
	if got.X != s.X || got.Y != s.Y {
		t.Errorf("記下的座標是 (%d, %d),預期 (%d, %d)", got.X, got.Y, s.X, s.Y)
	}
	if got.Location != s.locationCode() {
		t.Errorf("記下的地點是 %d,預期 %d", got.Location, s.locationCode())
	}
	if got.Floor != s.Floor {
		t.Errorf("記下的樓層是 %d,預期 %d", got.Floor, s.Floor)
	}
	if got.InHand() {
		t.Error("埋掉了卻還算在手上")
	}
	// 埋掉之後 U 的清單裡不該再出現它。
	for _, e := range s.usableEntries() {
		if e.Value == UseMoonstoneFirst+2 {
			t.Error("埋掉的月石還留在 U 的清單裡")
		}
	}
}

// TestMoonstoneCannotBeBuriedOnHardGround 反面:界外的地形要拒絕,而且**不能**
// 把月石記成已埋(原版那條路整段跳過四個寫入)。
func TestMoonstoneCannotBeBuriedOnHardGround(t *testing.T) {
	s := upkeepScene(t)
	s.Inventory.Moonstones[0] = u5data.Moonstone{Location: u5data.MoonstoneInHand}
	if !s.SetTileAt(s.X, s.Y, 0x0B) { // 山麓
		t.Skip("這一層地圖不支援寫入")
	}
	s.Messages = nil
	if s.BuryMoonstone(0) {
		t.Error("山麓上竟然埋下去了")
	}
	if !strings.Contains(strings.Join(s.Messages, "|"), MsgMoonstoneCannotBury) {
		t.Errorf("沒印「%s」:\n%s", MsgMoonstoneCannotBury, strings.Join(s.Messages, "\n"))
	}
	if !s.Inventory.Moonstones[0].InHand() {
		t.Error("埋失敗卻把月石弄丟了 —— 原版那條路一個欄位都沒寫")
	}
}

// TestMoonstoneCannotBeBuriedInDungeons —— `byte_3E0A3 >= 0x21` 直接拒絕。
func TestMoonstoneCannotBeBuriedInDungeons(t *testing.T) {
	s := upkeepScene(t)
	s.Inventory.Moonstones[0] = u5data.Moonstone{Location: u5data.MoonstoneInHand}
	s.SetTileAt(s.X, s.Y, 0x05) // 就算腳下是草地
	s.Dungeon = &DungeonState{Index: 0, Location: u5data.DungeonLocationBase}
	s.Messages = nil
	if s.BuryMoonstone(0) {
		t.Error("地牢裡竟然埋下去了")
	}
	if !s.Inventory.Moonstones[0].InHand() {
		t.Error("地牢裡埋失敗卻把月石弄丟了")
	}
}
