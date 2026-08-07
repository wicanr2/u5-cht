package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 空白鍵在海上揚著帆時是**收帆**,不是跳過一回合。
//
// 寫成「空白鍵就是跳過」的話玩家在海上找不到收帆的辦法 ——
// Y(Yell)那條路是**放**帆,不是收。
func TestSpaceLowersTheSails(t *testing.T) {
	s := dungeonState(t)
	s.Location, s.Floor = 0, 0
	s.Transport = u5data.VehicleSailing | 0x02 // 揚著帆、朝南
	s.Messages = nil
	s.Pass()
	if !strings.Contains(strings.Join(s.Messages, "|"), "收帆") {
		t.Errorf("揚著帆按空白鍵沒有收帆:\n%s", s.log())
	}
	if u5data.VehicleKind(s.Transport) != u5data.VehicleShip {
		t.Errorf("收帆後載具是 0x%02X,應回到大船那一組", s.Transport)
	}
	// ★ 朝向要保留 —— 低兩位元同時是開砲判舷側讀的東西。
	if s.Transport&0x03 != 0x02 {
		t.Errorf("收帆把朝向弄掉了:0x%02X", s.Transport)
	}
	// 收好帆之後再按就只是跳過。
	s.Messages = nil
	s.Pass()
	if !strings.Contains(strings.Join(s.Messages, "|"), "按兵不動") {
		t.Errorf("收好帆之後按空白鍵沒有跳過:\n%s", s.log())
	}
}

// 載具的動詞與朝向:馬與魔毯只在東西向換圖,船是四向全換。
//
// ⚠ 兩種規則不一樣(原版馬與魔毯只比 dir 1 與 3),不能寫成一條。
func TestVehicleFacingFollowsTheOriginal(t *testing.T) {
	s := dungeonState(t)
	s.Location, s.Floor = 0, 0

	// 馬:往東 0x12、往西 0x13;南北不動。
	s.Transport = u5data.TileHorse
	if v := s.faceVehicle(East); !strings.Contains(v, "騎行") {
		t.Errorf("騎馬走一步沒有動詞:%q", v)
	}
	if s.Transport != u5data.TileHorse|0x02 {
		t.Errorf("馬往東是 0x%02X,預期 0x%02X", s.Transport, u5data.TileHorse|0x02)
	}
	s.faceVehicle(West)
	if s.Transport != u5data.TileHorse|0x03 {
		t.Errorf("馬往西是 0x%02X", s.Transport)
	}
	before := s.Transport
	s.faceVehicle(North)
	if s.Transport != before {
		t.Errorf("馬往北竟然換了圖:0x%02X → 0x%02X", before, s.Transport)
	}

	// 船:四個朝向都換,而且低兩位元就是方向碼。
	s.Transport = u5data.VehicleShip
	for _, d := range []Direction{North, East, South, West} {
		if v := s.faceVehicle(d); v != "" {
			t.Errorf("船不該印動詞,卻印了 %q", v)
		}
		if s.Transport != u5data.VehicleShip|byte(d) {
			t.Errorf("船朝 %s 是 0x%02X,預期 0x%02X",
				d.Name(), s.Transport, u5data.VehicleShip|byte(d))
		}
	}

	// 小艇有動詞、魔毯有動詞、步行沒有。
	s.Transport = u5data.VehicleSkiff
	if !strings.Contains(s.faceVehicle(East), "划行") {
		t.Error("小艇沒有動詞")
	}
	s.Transport = u5data.VehicleCarpet
	if !strings.Contains(s.faceVehicle(East), "飛行") {
		t.Error("魔毯沒有動詞")
	}
	s.Transport = u5data.VehicleWalk
	if v := s.faceVehicle(East); v != "" {
		t.Errorf("步行不該有動詞,卻印了 %q", v)
	}
}

// Ztats 最後那一句預設是「並非聖者」。
//
// ★ 那個旗標唯一的來源是從 Ultima IV 轉入角色。新建的角色一律不是聖者 ——
// 這不是待補的預設值,是原版行為(U5 開場就是「你回來了,但不是以聖者的身分」)。
func TestZtatsSaysNotAnAvatarByDefault(t *testing.T) {
	s := dungeonState(t)
	if !s.BeginZtats() {
		t.Skip("開不了數值畫面")
	}
	joined := strings.Join(s.ZtatsLines(), "\n")
	if !strings.Contains(joined, "並非聖者") {
		t.Errorf("預設沒有印「並非聖者」:\n%s", joined)
	}
	s.TransferredAvatar = true
	joined = strings.Join(s.ZtatsLines(), "\n")
	if !strings.Contains(joined, "乃聖者") {
		t.Errorf("轉入過的角色沒有印「乃聖者」:\n%s", joined)
	}
}
