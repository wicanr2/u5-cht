package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// harpScene 造一個站在豎琴前的狀態。
func harpScene(t *testing.T, location, floor int) *State {
	t.Helper()
	s := &State{Scenes: synthScenes(t, walkable(t)), NPCs: &u5data.NPCSet{}, MaxMessages: 32}
	if err := s.SetScene(location, floor, 15, 15); err != nil {
		t.Fatalf("進不了地點 %d 第 %d 層:%v", location, floor, err)
	}
	// 豎琴在玩家**正南**那一格。
	s.SetTileAt(15, 16, u5data.HarpTile)
	return s
}

// playTune 把曲子彈完。
func playTune(s *State) {
	for _, n := range u5data.HarpTune {
		s.PlayNote(rune('0' + n))
	}
}

// ⚠ 只有正南那一格是豎琴才彈得出聲音。
func TestTheHarpMustBeDueSouth(t *testing.T) {
	s := harpScene(t, u5data.HarpDoorLocation, u5data.HarpDoorFloor)
	if !s.AtHarp() {
		t.Fatal("南邊有豎琴卻說沒有")
	}
	if !s.PlayNote('5') {
		t.Error("站在豎琴前應該彈得出聲音")
	}
	// 把豎琴搬到北邊。
	s.SetTileAt(15, 16, walkable(t))
	s.SetTileAt(15, 14, u5data.HarpTile)
	if s.AtHarp() {
		t.Error("豎琴在北邊不算 —— 原版只看正南那一格")
	}
	if s.PlayNote('5') {
		t.Error("不在豎琴前不該發音")
	}
}

// 在不列顛王城堡二樓彈對曲子 → 那一格被切換。
func TestTheTuneOpensTheSecretDoor(t *testing.T) {
	s := harpScene(t, u5data.HarpDoorLocation, u5data.HarpDoorFloor)
	before := s.TileAt(u5data.HarpDoorX, u5data.HarpDoorY)
	playTune(s)
	after := s.TileAt(u5data.HarpDoorX, u5data.HarpDoorY)
	if after != before^u5data.HarpDoorXor {
		t.Errorf("(%d,%d) 從 %02X 變成 %02X,應該是 %02X",
			u5data.HarpDoorX, u5data.HarpDoorY, before, after, before^u5data.HarpDoorXor)
	}
	if !strings.Contains(allLogs(s), MsgSecretDoor) {
		t.Error("應該有一行提示")
	}
	// ⚠ 一個 XOR 同時做開與關 —— 再彈一次關回去。
	playTune(s)
	if s.TileAt(u5data.HarpDoorX, u5data.HarpDoorY) != before {
		t.Error("再彈一次應該關回去")
	}
}

// 換個地方彈對也沒用。
func TestTheTuneOnlyWorksInThatOneRoom(t *testing.T) {
	// 對的地點、錯的樓層。
	s := harpScene(t, u5data.HarpDoorLocation, 0)
	before := s.TileAt(u5data.HarpDoorX, u5data.HarpDoorY)
	playTune(s)
	if s.TileAt(u5data.HarpDoorX, u5data.HarpDoorY) != before {
		t.Error("樓層不對不該有反應")
	}
	// 錯的地點。
	s2 := harpScene(t, britain, 0)
	before2 := s2.TileAt(u5data.HarpDoorX, u5data.HarpDoorY)
	playTune(s2)
	if s2.TileAt(u5data.HarpDoorX, u5data.HarpDoorY) != before2 {
		t.Error("地點不對不該有反應")
	}
}

// ⚠ 彈錯**不是直接歸零** —— 原版手寫了一組回退規則(等同 KMP 的 failure function)。
func TestAWrongNoteBacksOffInsteadOfResetting(t *testing.T) {
	// 進度 10 彈 8 → 退回 3(不是 0)。
	if got := u5data.HarpNext(10, 8); got != 3 {
		t.Errorf("進度 10 彈 8 應該退回 3,實得 %d", got)
	}
	// 進度 11 彈 7 → 退回 2。
	if got := u5data.HarpNext(11, 7); got != 2 {
		t.Errorf("進度 11 彈 7 應該退回 2,實得 %d", got)
	}
	// 彈到曲子的第一個音 → 退回 1。
	if got := u5data.HarpNext(5, u5data.HarpTune[0]); got != 1 {
		t.Errorf("彈第一個音應該退回 1,實得 %d", got)
	}
	// 其餘歸零。
	if got := u5data.HarpNext(5, 1); got != 0 {
		t.Errorf("完全不相干的音應該歸零,實得 %d", got)
	}
	// 正常前進。
	for i := range u5data.HarpTune {
		if got := u5data.HarpNext(i, u5data.HarpTune[i]); got != i+1 {
			t.Errorf("進度 %d 彈對應該變成 %d,實得 %d", i, i+1, got)
		}
	}

	// 端對端:彈一半、彈錯一個、從回退點接回去,一樣開得了門。
	s := harpScene(t, u5data.HarpDoorLocation, u5data.HarpDoorFloor)
	before := s.TileAt(u5data.HarpDoorX, u5data.HarpDoorY)
	for i := 0; i < 5; i++ {
		s.PlayNote(rune('0' + u5data.HarpTune[i]))
	}
	s.PlayNote('1') // 錯音 → 歸零
	playTune(s)
	if s.TileAt(u5data.HarpDoorX, u5data.HarpDoorY) == before {
		t.Error("彈錯之後重來一次應該還是開得了門")
	}
}

// 音階是一個大調音階,而且鍵 '0' 是音階外的高音。
func TestTheHarpScaleIsAMajorScale(t *testing.T) {
	// 鍵 1..9 = 0 2 4 5 7 9 11 12 14 半音。
	want := []int{0, 2, 4, 5, 7, 9, 11, 12, 14}
	for i, w := range want {
		if u5data.HarpScale[i+1] != w {
			t.Errorf("鍵 %d 是 %d 半音,預期 %d", i+1, u5data.HarpScale[i+1], w)
		}
	}
	// ⚠ 鍵 '0' 是 16,不是 0 —— 別為了讓表整齊而「修正」它。
	if u5data.HarpScale[0] != 16 {
		t.Errorf("鍵 0 是 %d,原版是 16", u5data.HarpScale[0])
	}
}
