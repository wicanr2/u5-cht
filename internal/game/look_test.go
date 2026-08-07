package game

import (
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// TestClockFaceMatchesTheOriginalFormat:老爺鐘的時刻格式。
//
// 原版 `sub_D258`:`hour % 12`,餘 0 印 12;分鐘補兩位;`hour <= 11` 是上午。
// 兩個邊界最容易寫錯 —— 0 時該印「12:00 上午」而不是「0:00」,
// 12 時該印「12:00 下午」而不是「12:00 上午」。
func TestClockFaceMatchesTheOriginalFormat(t *testing.T) {
	s := &State{}
	cases := []struct {
		hour, minute int
		want         string
	}{
		{0, 0, "12:00" + MsgClockAM},
		{0, 5, "12:05" + MsgClockAM},
		{11, 59, "11:59" + MsgClockAM},
		{12, 0, "12:00" + MsgClockPM},
		{13, 30, "1:30" + MsgClockPM},
		{23, 45, "11:45" + MsgClockPM},
	}
	for _, c := range cases {
		s.Clock.Hour, s.Clock.Minute = c.hour, c.minute
		if got := s.clockFace(); got != c.want {
			t.Errorf("%d 時 %d 分 → %q,預期 %q", c.hour, c.minute, got, c.want)
		}
	}
}

// TestLookFollowsRedirectTiles:轉向格會一路跟下去,不是只跟一步。
//
// 原版 `loc_D268` 是個迴圈。寫成「查一次」的話,兩格相連的轉向就會停在
// 第二個轉向格上,印出它自己的敘述(「parched desert」)—— 而那看起來
// 完全合理,不會有人發現錯了。
func TestLookFollowsRedirectTiles(t *testing.T) {
	s := newLookState(t)
	x, y := s.X+1, s.Y
	// 目標格與它西邊那格都是「往西看」,再西邊才是真正的東西。
	s.SetTileAt(x, y, LookRedirectWest)
	s.SetTileAt(x-1, y, LookRedirectWest)
	s.SetTileAt(x-2, y, u5data.TileShrine)

	s.MaxMessages = 20
	s.Messages = nil
	s.lookTerrain(s.TileAt(x, y), x, y)
	joined := strings.Join(s.Messages, "\n")
	if !strings.Contains(joined, s.lookTerrainText(u5data.TileShrine)) {
		t.Errorf("轉向沒跟到底,印出來的是:\n%s", joined)
	}
}

// TestLookAtAnObjectHidesTheTerrain:有物件就只印物件。
//
// 原版是 `jmp`,不是接著往下印。放在招牌前面的一袋金幣**會蓋掉招牌** ——
// 這條把那個行為釘住,免得日後有人「順手」改成兩個都印。
func TestLookAtAnObjectHidesTheTerrain(t *testing.T) {
	s := newLookState(t)
	objs := s.CurrentObjects()
	if objs == nil {
		t.Skip("這一局沒有地圖物件可用")
	}
	// ⚠ 不能拿槽 0 —— 那是隊伍自己(原版 `sub_2B360` 從槽 1 起掃),
	// 而它的敘述在表上是佔位符 `x`。找第一個真的有話可說的物件。
	var o *u5data.MapObject
	for i := range objs.Objects {
		if i == u5data.PartyObjectSlot {
			continue
		}
		c := &objs.Objects[i]
		if c.Present() && c.Floor == s.Floor && !u5data.LookPlaceholder(s.Look2.Object(int(c.Kind))) {
			o = c
			break
		}
	}
	if o == nil {
		t.Skip("這一層沒有帶敘述的物件")
	}
	s.MaxMessages = 20
	s.Messages = nil
	s.lookAt(int(o.X), int(o.Y))
	joined := strings.Join(s.Messages, "\n")
	if !strings.Contains(joined, MsgThouDostSee) {
		t.Fatalf("沒印出「%s」:\n%s", MsgThouDostSee, joined)
	}
	if want := s.lookObject(o.Kind); want != "" && !strings.Contains(joined, want) {
		t.Errorf("沒印出物件敘述 %q:\n%s", want, joined)
	}
}

// TestFountainDoesNotHeal:噴泉在原版**沒有療效**。
//
// 這是「不要自己加東西」的守門測試。組語裡沒有任何寫入(`docs/re/37` §4),
// 所以喝完之後 HP 與狀態都要原封不動。
func TestFountainDoesNotHeal(t *testing.T) {
	s := newLookState(t)
	if len(s.Roster) == 0 {
		t.Skip("沒有角色")
	}
	s.Roster[0].HP = 5
	before := s.Roster[0].HP
	s.drinkFromFountain()
	for i := range s.Roster {
		if i == 0 && s.Roster[i].HP != before {
			t.Errorf("噴泉把 HP 從 %d 改成 %d —— 原版不會", before, s.Roster[i].HP)
		}
	}
}

// TestSunBurnsTheLooker:白天抬頭看太陽會扣 1 點 HP。
//
// 反直覺,所以特別釘住:原版 `sub_D064` 的白天分支就是
// 印「the sun!」再 `sub_2A464(member, 1)`。
func TestSunBurnsTheLooker(t *testing.T) {
	s := newLookState(t)
	if len(s.Roster) == 0 {
		t.Skip("沒有角色")
	}
	for i := range s.Roster {
		s.Roster[i].Status = u5data.StatusGood
		s.Roster[i].HP = 20
	}
	s.Clock.Hour = 12
	total := func() int {
		n := 0
		for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
			n += int(s.Roster[i].HP)
		}
		return n
	}
	before := total()
	s.lookAtTheSky()
	if got := total(); got != before-skySunDamage {
		t.Errorf("看太陽之後全隊 HP 共 %d,預期 %d", got, before-skySunDamage)
	}

	// 夜裡不扣。
	s.Clock.Hour = 2
	before = total()
	s.lookAtTheSky()
	if got := total(); got != before {
		t.Errorf("夜裡看天也扣了血:%d → %d", before, got)
	}
}

func newLookState(t *testing.T) *State {
	t.Helper()
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	s := realState(t, dir)
	lt, err := u5data.LoadLook(dir)
	if err != nil {
		t.Fatalf("讀 LOOK2.DAT:%v", err)
	}
	ss, err := u5data.LoadSigns(dir)
	if err != nil {
		t.Fatalf("讀 SIGNS.DAT:%v", err)
	}
	s.Look2, s.Signs = lt, ss
	return s
}
