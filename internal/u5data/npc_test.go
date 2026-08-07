package u5data

import (
	"os"
	"testing"
)

// TestNPCScheduleSlot 鎖住排程的 slot 選擇。
//
// 用不列顛城 1 號 NPC 的真實排程當範例(位置 (5,18)F1 / (4,19)F0 / (2,5)F0,
// 時刻 17/6/11/13)。它的一日作息應該是:
//
//	00:00 slot0(二樓,睡覺)→ 06:00 slot1(上工)→ 11:00 slot2(外出)
//	→ 13:00 slot1(回崗位)→ 17:00 slot0(上樓就寢)
//
// 13:00 那一段是關鍵:第四個時刻指向 **slot 1**,不是 slot 3 ——
// 原版 sub_9C7C 最後那行 `if d[slot] > d[3] { slot = 1 }` 就是在做這件事。
// 少了它,NPC 下午會憑空消失(或停在錯的位置)。
func TestNPCScheduleSlot(t *testing.T) {
	s := NPCSchedule{Times: [4]byte{17, 6, 11, 13}}
	want := map[int]int{
		0: 0, 5: 0, 6: 1, 10: 1, 11: 2, 12: 2, 13: 1, 16: 1, 17: 0, 23: 0,
	}
	for hour, w := range want {
		if got := s.Slot(hour); got != w {
			t.Errorf("%02d:00 → slot %d,預期 %d", hour, got, w)
		}
	}
}

// TestNPCScheduleWrapsMidnight:8-bit 減法讓 24 小時自然環繞,
// 跨午夜不該挑錯 slot。
func TestNPCScheduleWrapsMidnight(t *testing.T) {
	s := NPCSchedule{Times: [4]byte{22, 6, 11, 13}} // 22:00 就寢
	for _, h := range []int{22, 23, 0, 1, 5} {
		if got := s.Slot(h); got != 0 {
			t.Errorf("%02d:00 應該還在 slot 0(夜間),實得 %d", h, got)
		}
	}
	if got := s.Slot(6); got != 1 {
		t.Errorf("06:00 應該換到 slot 1,實得 %d", got)
	}
}

func TestParseNPCBlockRejectsWrongSize(t *testing.T) {
	if _, err := ParseNPCBlock(make([]byte, 100)); err == nil {
		t.Error("大小不對卻沒報錯")
	}
}

// TestRealNPCData 用原版資料檢查整批 NPC 的合理性。
func TestRealNPCData(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	set, err := LoadNPCSet(dir)
	if err != nil {
		t.Fatal(err)
	}
	total, shops := 0, 0
	for n := 1; n <= len(Locations); n++ {
		loc := &Locations[n-1]
		npcs, err := set.At(n)
		if err != nil {
			t.Fatal(err)
		}
		// 0 號槽是隊伍自己,檔案裡的值不是正常的 NPC 記錄(多數地點是 0,
		// 少數留著 0x1C 或 0x29 這類殘值)。原版的 NPC 更新迴圈 sub_8924
		// 就是從 esi = 1 起跑的 —— 所以這一格一律略過,不要對它做任何斷言。
		for i := 1; i < NPCsPerLocation; i++ {
			n2 := &npcs[i]
			if !n2.Present() {
				continue
			}
			total++
			if n2.IsShopkeeper() {
				shops++
			}
			// 人物的生物編號都是 4 的倍數 —— 人物圖以四張(四個朝向)為一組。
			// 但 NPC 槽也放非人物(馬匹、箱子之類):它們的編號 < 0x40,
			// tile 落在 256–287 那一列的物件圖,單張不成組。
			// sub_B98 判「這是不是可以被嚇跑的平民」用的正是 0x40 <= tile < 0x74。
			// 上界同樣取自 sub_B98:0x74 以上是怪物(STONEGATE 就有一批),
			// 那些也不照四張一組排。
			if n2.Creature >= 0x40 && n2.Creature < 0x74 && n2.Creature%4 != 0 {
				t.Errorf("%s #%d 的人物生物編號 %d 不是 4 的倍數", loc.Name, i, n2.Creature)
			}
			// 排程位置必須落在 32×32 之內,樓層必須在這個地點的範圍內。
			for s := 0; s < 3; s++ {
				x, y := int(n2.Schedule.X[s]), int(n2.Schedule.Y[s])
				f := int(int8(n2.Schedule.Floor[s]))
				if x >= SceneSide || y >= SceneSide {
					t.Errorf("%s #%d slot%d 的位置 (%d,%d) 超出 32×32", loc.Name, i, s, x, y)
				}
				if f < loc.FloorMin || f > loc.FloorMax {
					t.Errorf("%s #%d slot%d 的樓層 %+d 超出 %+d..%+d",
						loc.Name, i, s, f, loc.FloorMin, loc.FloorMax)
				}
			}
			// 時刻必須是合法的小時。
			for _, h := range n2.Schedule.Times {
				if h >= 24 {
					t.Errorf("%s #%d 的時刻 %d 不是合法小時", loc.Name, i, h)
				}
			}
		}
	}
	if total < 100 {
		t.Errorf("全遊戲只找到 %d 個 NPC,太少了 —— 解析大概錯了", total)
	}
	t.Logf("32 個地點共 %d 個 NPC,其中 %d 個是商人", total, shops)
}
