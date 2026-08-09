package game

import (
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// ★★★ 每走一步都要印方向名 —— 這條是**對 DOSBox 原版並排比對**抓到的
//(`docs/playtest-checkpoints.md` A2:原版三步印三行 `>South`,引擎一行都沒有)。
//
// 值得記下來的地方:這個缺口存在了整個 P2–P4,而**沒有任何測試會紅** ——
// 少一句回饋不會讓任何斷言失敗。只有把畫面並排才看得出來。
func TestSceneMoveEchoesTheDirection(t *testing.T) {
	s := shopState(t, 22)
	if err := s.SetScene(22, 0, 15, 15); err != nil {
		t.Fatal(err)
	}
	before := len(s.Messages)
	s.Move(North)
	if len(s.Messages) <= before {
		t.Fatal("場景移動一句話都沒印")
	}
	if got := s.Messages[before]; got != North.Name() {
		t.Errorf("第一句是 %q,預期方向名 %q", got, North.Name())
	}
}

// ⚠ 方向名要印在**移動判定之前** —— 撞牆時讀起來是「方向」→「去路受阻!」。
// 倒過來是另一個意思(先說走不動,再說往哪走)。
func TestBlockedMoveStillEchoesTheDirectionFirst(t *testing.T) {
	s := shopState(t, 22)
	if err := s.SetScene(22, 0, 15, 15); err != nil {
		t.Fatal(err)
	}
	// 在場景裡找一個「站得住、但某個方向有牆」的位置 ——
	// 起點四周不一定有牆,固定用起點會讓這條測試時而跳過。
	for y := 1; y < 30; y++ {
		for x := 1; x < 30; x++ {
			if u5data.TileBlocksWalking(int(s.TileAt(x, y))) {
				continue
			}
			for _, d := range []Direction{North, East, South, West} {
				dx, dy := d.Delta()
				if !u5data.TileBlocksWalking(int(s.TileAt(x+dx, y+dy))) {
					continue
				}
				s.X, s.Y = x, y
				s.Messages = nil
				s.Move(d)
				if len(s.Messages) < 2 {
					t.Fatalf("在 (%d,%d) 往%s撞牆只印了 %d 句:%v",
						x, y, d.Name(), len(s.Messages), s.Messages)
				}
				if s.Messages[0] != d.Name() {
					t.Errorf("撞牆時第一句是 %q,預期方向名 %q", s.Messages[0], d.Name())
				}
				return
			}
		}
	}
	t.Fatal("整個場景找不到一面牆 —— 場景資料可能沒載入")
}

// ★ 揚著帆的船**不印**方向名(原版 `sub_2D174` 的 `byte_3E167` 閘門)——
// 船改印轉向那一句,兩句都印會變成重複回報。
//
// 反對照在同一支測試裡:同一個位置**下船**之後就要印。少了反對照,
// 「整支回顯壞掉」與「船不印」會得到同一個綠燈。
func TestSailingDoesNotEchoButWalkingDoes(t *testing.T) {
	s := worldState(t)
	s.Transport = u5data.VehicleSailing
	s.Messages = nil
	s.Move(North)
	for _, m := range s.Messages {
		if m == North.Name() {
			t.Fatalf("揚帆時印了方向名:%v", s.Messages)
		}
	}
	// 反對照:步行時同一步要印。
	s.Transport = 0
	s.Messages = nil
	s.Move(North)
	if len(s.Messages) == 0 || s.Messages[0] != North.Name() {
		t.Fatalf("步行時沒印方向名:%v", s.Messages)
	}
}
