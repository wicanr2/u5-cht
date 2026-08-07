package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 檀香木盒的整條路:坐上豎琴前的椅子 → 彈十三個音 → 牆開了 → 走進去 → 撿走。
//
// 這一條刻意用**真的原版資料**跑,不用合成場景。理由是它要驗的正是資料本身:
// 盒子躺在 `CASTLE.NPC` 的第 31 槽(生物編號 0x0E),而那間密室在
// `CASTLE.DAT` 裡是被牆封死的 —— 兩份檔案要對得上,這條路才走得通。
func TestTheSandalwoodBoxBehindTheHarpDoor(t *testing.T) {
	dir := gameDataDir(t)
	if dir == "" {
		return
	}
	s := realState(t, dir)

	// 二樓,坐在豎琴北邊那張椅子上(豎琴在 (17,18),椅子在 (17,17))。
	if err := s.SetScene(u5data.HarpDoorLocation, u5data.HarpDoorFloor, 17, 17); err != nil {
		t.Fatalf("進不了不列顛王城堡二樓:%v", err)
	}
	if !s.AtHarp() {
		t.Fatalf("(17,17) 的正南應該是豎琴,實得地形 %02X", s.TileAt(17, 18))
	}

	// 開之前那一格是牆。
	const wall, floor = 0x4F, 0x44
	if got := s.TileAt(u5data.HarpDoorX, u5data.HarpDoorY); got != wall {
		t.Fatalf("暗門那一格開之前是 %02X,原版資料裡應該是石牆 %02X", got, wall)
	}

	// 十三個音。
	for _, n := range u5data.HarpTune {
		if !s.PlayNote(rune('0' + n)) {
			t.Fatalf("彈第 %d 個音沒出聲", n)
		}
	}
	if got := s.TileAt(u5data.HarpDoorX, u5data.HarpDoorY); got != floor {
		t.Fatalf("彈完曲子那一格是 %02X,應該變成地板 %02X", got, floor)
	}

	// 密室在 (17,12) 與 (18,12),盒子在 (18,12)。走進去站在 (17,12)。
	s.X, s.Y = u5data.HarpDoorX, u5data.HarpDoorY-1
	o, slot, ok := s.ObjectAt(18, 12)
	if !ok {
		t.Fatal("(18,12) 上沒有物件 —— 盒子應該由 NPC 槽鏡射過來")
	}
	if o.Kind != u5data.ItemSandalwood {
		t.Fatalf("(18,12) 的物件種類是 %02X,應該是檀香木盒 %02X",
			o.Kind, u5data.ItemSandalwood)
	}
	if i, mirrored := s.npcOfObject(slot); !mirrored || i != 31 {
		t.Errorf("那一槽應該是 NPC 31 的鏡射,實得 npc=%d mirrored=%v", i, mirrored)
	}

	// 往東撿。
	s.getAt(1, 0)
	if !s.SandalwoodBox {
		t.Fatalf("撿完之後沒拿到盒子,訊息:%q", allLogs(s))
	}
	if !strings.Contains(allLogs(s), "檀香") {
		t.Errorf("訊息裡沒提到檀香木盒:%q", allLogs(s))
	}
	// 地上不該還有一個,而且要記進存檔用的遮罩。
	if _, _, still := s.ObjectAt(18, 12); still {
		t.Error("撿走之後 (18,12) 還躺著一個")
	}
	if s.RemovedNPC[u5data.HarpDoorLocation-1]&(1<<31) == 0 {
		t.Error("撿走之後沒把 NPC 31 記進永久移除遮罩 —— 離場再回來會再長出來")
	}
	// 過一回合也不能被 syncNPCObjects 配回來。
	s.tick()
	if _, _, back := s.ObjectAt(18, 12); back {
		t.Error("下一回合盒子又出現了 —— 鏡射沒有尊重永久移除遮罩")
	}
}

// 地下室的東西要看得見:樓層位元組 0xFF 是 −1,不是 255。
//
// 不列顛王城堡地下室有三個關著的寶箱(槽 23/24/25,樓層位元組 0xFF)。
// 少了有號轉換,它們會拿 255 去跟樓層 −1 比,整層的東西一個都不出現 ——
// 而畫面上只會看到一間空房,完全不像 bug。
func TestBasementChestsAreVisibleAndOpenable(t *testing.T) {
	dir := gameDataDir(t)
	if dir == "" {
		return
	}
	s := realState(t, dir)
	if err := s.SetScene(u5data.HarpDoorLocation, -1, 16, 20); err != nil {
		t.Fatalf("進不了地下室:%v", err)
	}
	want := [][2]int{{16, 21}, {17, 22}, {13, 23}}
	for _, p := range want {
		o, _, ok := s.ObjectAt(p[0], p[1])
		if !ok {
			t.Errorf("(%d,%d) 上沒有寶箱", p[0], p[1])
			continue
		}
		if o.Kind != u5data.ItemClosedChest {
			t.Errorf("(%d,%d) 的種類是 %02X,應該是關著的寶箱 %02X",
				p[0], p[1], o.Kind, u5data.ItemClosedChest)
		}
		// 品質是 sub_1E74 幫寶箱填的 0x1E —— 地表寶箱擲獎 random(1,30) 的上限。
		if o.Raw[u5data.ObjQuality] != u5data.NPCObjectQualityChest {
			t.Errorf("(%d,%d) 的品質是 %02X,應該是 %02X",
				p[0], p[1], o.Raw[u5data.ObjQuality], u5data.NPCObjectQualityChest)
		}
	}
}

// 有號樓層本身。
func TestFloorByteIsSigned(t *testing.T) {
	cases := map[byte]int{0: 0, 1: 1, 3: 3, 0x7F: 127, 0xFF: -1, 0xFE: -2}
	for in, want := range cases {
		if got := u5data.SignedFloor(in); got != want {
			t.Errorf("SignedFloor(%02X) = %d,應該是 %d", in, got, want)
		}
	}
}

// 被點名的槽拿 0xFF,其餘拿 0 —— 遮罩照抄,不賦予語意。
func TestFlaggedSlotsGetQualityFF(t *testing.T) {
	// 石門(地點 29)的槽 5..8。
	for _, slot := range []int{5, 6, 7, 8} {
		if got := u5data.NPCObjectQuality(0x94, 29, slot); got != u5data.NPCObjectQualityFlagged {
			t.Errorf("石門槽 %d 的品質是 %02X,應該是 %02X",
				slot, got, u5data.NPCObjectQualityFlagged)
		}
	}
	if got := u5data.NPCObjectQuality(0x94, 29, 4); got != 0 {
		t.Errorf("石門槽 4 沒被點名,品質應該是 0,實得 %02X", got)
	}
	// 寶箱優先於遮罩(原版的判斷順序就是先看生物編號 1)。
	if got := u5data.NPCObjectQuality(1, 29, 5); got != u5data.NPCObjectQualityChest {
		t.Errorf("寶箱的品質是 %02X,應該是 %02X", got, u5data.NPCObjectQualityChest)
	}
}
