package game

import (
	"os"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// britain 是不列顛城的地點編號(TOWNE.DAT 索引 2,兩層)。測試裡拿它當代表。
const britain = 2

// synthScenes 造一組合成場景:每張地圖全是可走的 tile,方便單獨驗狀態機的規則。
// 用真實的地點編號(所以 SceneIndex / Floors 走的是原版表),但地圖內容是自己填的。
func synthScenes(t *testing.T, fill byte) *u5data.SceneSet {
	t.Helper()
	s := &u5data.SceneSet{}
	for i := range s.Files {
		s.Files[i] = make([]u5data.SceneMap, u5data.ScenesPerFile)
		for j := range s.Files[i] {
			for k := range s.Files[i][j].Tiles {
				s.Files[i][j].Tiles[k] = fill
			}
		}
	}
	return s
}

// walkable 找一個原版通行表允許行走的 tile,當合成地圖的地板。
// 不寫死數字 —— 通行表是從執行檔取出來的,寫死會在表更新時悄悄失效。
//
// ⚠ **跳過 0**。通行表說 0 可以走,但 0 在引擎裡是**「這一格沒有資料」的哨兵**:
// 視線遮蔽判「在不在場景內」看的就是場景緩衝那一格是不是 0
//(`cmp byte_3F8F4[edx], 0`),而遮蔽罩最後也把 0 併成「看不見」。
// 拿 0 當地板的話,合成地圖等於一整片「沒有資料」——
// 移動規則照樣驗得過,但任何跟可見度有關的測試都會全黑而看不出原因。
func walkable(t *testing.T) byte {
	t.Helper()
	for i := 1; i < 256; i++ {
		if !u5data.TileBlocksWalking(i) && i != u5data.TileBlank {
			if _, isStairs := u5data.StairsFacing(byte(i)); !isStairs {
				return byte(i)
			}
		}
	}
	t.Fatal("通行表裡找不到任何可走的 tile")
	return 0
}

func newState(t *testing.T) *State {
	t.Helper()
	return &State{Scenes: synthScenes(t, walkable(t)), MaxMessages: 8}
}

func TestEnterUsesLocationTable(t *testing.T) {
	s := newState(t)
	loc := &u5data.Locations[britain-1]
	s.X, s.Y = loc.X, loc.Y

	s.Enter()
	if !s.InScene() {
		t.Fatal("站在不列顛城的座標上按進入,卻沒進場景")
	}
	if s.Location != britain {
		t.Errorf("進到地點 %d,預期 %d", s.Location, britain)
	}
	if s.X != SceneEntryX || s.Y != SceneEntryY {
		t.Errorf("入口在 (%d,%d),原版固定是 (%d,%d)", s.X, s.Y, SceneEntryX, SceneEntryY)
	}
	if s.Floor != 0 {
		t.Errorf("進場景時樓層是 %d,應為 0(地面層)", s.Floor)
	}
}

func TestEnterNowhereDoesNothing(t *testing.T) {
	s := newState(t)
	s.X, s.Y = 1, 1 // 地點表裡沒有這一格
	s.Enter()
	if s.InScene() {
		t.Error("在空地按進入竟然進了場景")
	}
	if len(s.Messages) == 0 || s.Messages[len(s.Messages)-1] != MsgNothingToEnter {
		t.Errorf("應回報「%s」,實得 %v", MsgNothingToEnter, s.Messages)
	}
}

// TestEdgeDetectionUsesPreMoveCoord:原版比的是**移動前**的座標
// (`cmp byte_3E0A6, 1 / jnb`),所以站在 x=1 往西走只是走到 x=0,不會問要不要離開;
// 要站在 x=0 再往西才問。這個 off-by-one 弄反的話,城鎮最外圈就永遠走不到。
func TestEdgeDetectionUsesPreMoveCoord(t *testing.T) {
	cases := []struct {
		name      string
		x, y      int
		dir       Direction
		wantAsk   bool
		wantX     int
		wantY     int
		rationale string
	}{
		{"站在 x=1 往西", 1, 15, West, false, 0, 15, "只是走到最外圈"},
		{"站在 x=0 往西", 0, 15, West, true, 0, 15, "已在最外圈,再往外才是離開"},
		{"站在 x=30 往東", 30, 15, East, false, 31, 15, "走到最外圈"},
		{"站在 x=31 往東", 31, 15, East, true, 31, 15, "離開"},
		{"站在 y=1 往北", 15, 1, North, false, 15, 0, "走到最外圈"},
		{"站在 y=0 往北", 15, 0, North, true, 15, 0, "離開"},
		{"站在 y=30 往南", 15, 30, South, false, 15, 31, "走到最外圈"},
		{"站在 y=31 往南", 15, 31, South, true, 15, 31, "離開"},
	}
	for _, c := range cases {
		s := newState(t)
		if err := s.SetScene(britain, 0, c.x, c.y); err != nil {
			t.Fatal(err)
		}
		s.Move(c.dir)
		if got := s.Prompt == PromptLeave; got != c.wantAsk {
			t.Errorf("%s:問要不要離開 = %v,預期 %v(%s)", c.name, got, c.wantAsk, c.rationale)
		}
		if s.X != c.wantX || s.Y != c.wantY {
			t.Errorf("%s:走到 (%d,%d),預期 (%d,%d)", c.name, s.X, s.Y, c.wantX, c.wantY)
		}
	}
}

func TestLeaveReturnsToLocationCoord(t *testing.T) {
	s := newState(t)
	loc := &u5data.Locations[britain-1]
	if err := s.SetScene(britain, 0, 5, 5); err != nil { // 從城中央離開
		t.Fatal(err)
	}
	s.Prompt = PromptLeave
	s.Answer(true)

	if s.InScene() {
		t.Fatal("回答「是」之後還在場景裡")
	}
	// 原版是從地點表把座標讀回來,不是記住進來前的位置 —— 所以在城裡走到哪裡出去
	// 都會回到城門那一格。
	if s.X != loc.X || s.Y != loc.Y {
		t.Errorf("回到 (%d,%d),預期地點表上的 (%d,%d)", s.X, s.Y, loc.X, loc.Y)
	}
	if s.Floor != 0 {
		t.Errorf("回到地表時樓層是 %d,應為 0", s.Floor)
	}
}

// TestLeaveAraratGoesToUnderworld:原版 sub_86C 對地點 25 特判 —— 印 "Underworld!"
// 而不是 "Britannia!",樓層設 -1。
func TestLeaveAraratGoesToUnderworld(t *testing.T) {
	s := newState(t)
	if err := s.SetScene(UnderworldLocation, 0, 5, 5); err != nil {
		t.Fatal(err)
	}
	s.Prompt = PromptLeave
	s.Answer(true)
	if s.Floor != -1 {
		t.Errorf("離開 ARARAT 後樓層是 %d,預期 -1(地下世界)", s.Floor)
	}
	if last := s.Messages[len(s.Messages)-1]; last != MsgExitTo+MsgUnderworld {
		t.Errorf("訊息是 %q,預期 %q", last, MsgExitTo+MsgUnderworld)
	}
}

func TestAnswerNoStaysInScene(t *testing.T) {
	s := newState(t)
	if err := s.SetScene(britain, 0, 0, 15); err != nil {
		t.Fatal(err)
	}
	s.Move(West)
	if s.Prompt != PromptLeave {
		t.Fatal("走出邊界卻沒有問是否離開")
	}
	// 有提問待答時,移動輸入必須無效
	s.Move(East)
	if s.X != 0 {
		t.Errorf("提問未答時竟然移動了,現在 x=%d", s.X)
	}
	s.Answer(false)
	if !s.InScene() || s.Prompt != PromptNone {
		t.Errorf("回答「否」後應留在場景且提問清空,實得 InScene=%v Prompt=%v", s.InScene(), s.Prompt)
	}
}

// TestStairsDirection:原版 sub_758 —— 同向走進樓梯往上,反向往下,側向不動。
func TestStairsDirection(t *testing.T) {
	const facing = int(North)
	stairs := byte(u5data.StairsBase + facing)

	cases := []struct {
		dir       Direction
		wantFloor int
		why       string
	}{
		{North, 1, "與樓梯同向走進去 → 上樓"},
		{South, -1, "反向走進去 → 下樓(這裡沒有地下層,所以會被擋下)"},
		{East, 0, "側向走進去 → 不換層"},
	}
	for _, c := range cases {
		s := newState(t)
		if err := s.SetScene(britain, 0, 15, 15); err != nil {
			t.Fatal(err)
		}
		dx, dy := c.dir.Delta()
		// 把樓梯放在玩家要走進去的那一格
		s.Scenes.Files[0][u5data.Locations[britain-1].SceneIndex].
			Tiles[(15+dy)*u5data.SceneSide+(15+dx)] = stairs
		s.Move(c.dir)

		want := c.wantFloor
		if want < 0 {
			want = 0 // 不列顛城沒有地下層,changeFloor 會拒絕並留在原地
		}
		if s.Floor != want {
			t.Errorf("往%s走進朝北的樓梯 → 樓層 %d,預期 %d(%s)", c.dir.Name(), s.Floor, want, c.why)
		}
	}
}

func TestOppositeIsXor2(t *testing.T) {
	// 原版就是靠 XOR 2 取反向(sub_758 的 `xor dl, 2`),不是查表。
	for _, d := range []Direction{North, East, South, West} {
		if got := d.Opposite().Opposite(); got != d {
			t.Errorf("%s 的反向的反向是 %s", d.Name(), got.Name())
		}
		dx, dy := d.Delta()
		ox, oy := d.Opposite().Delta()
		if dx != -ox || dy != -oy {
			t.Errorf("%s(%d,%d)與其反向(%d,%d)不是相反", d.Name(), dx, dy, ox, oy)
		}
	}
}

// TestSceneViewportBlankOutsideBounds:場景邊界外要是 TileBlank(0xFF,純黑),
// 不是 tile 0 —— tile 0 在 tileset 裡是一團紅黃爆裂圖案,填錯會在城外鋪出一片火花。
func TestSceneViewportBlankOutsideBounds(t *testing.T) {
	s := newState(t)
	if err := s.SetScene(britain, 0, 15, 30); err != nil {
		t.Fatal(err)
	}
	for _, p := range [][2]int{{-1, 15}, {32, 15}, {15, -1}, {15, 32}, {40, 40}} {
		if got := s.TileAt(p[0], p[1]); got != u5data.TileBlank {
			t.Errorf("場景外 (%d,%d) 的 tile 是 %d,預期 %d", p[0], p[1], got, u5data.TileBlank)
		}
	}
}

// TestRealSceneRoundTrip 用原版資料跑一遍「進城 → 走動 → 出城」。
// 沒設 U5_GAMEDATA 就跳過 —— 版權素材由玩家自備,不入庫。
func TestRealSceneRoundTrip(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	scenes, err := u5data.LoadSceneSet(dir)
	if err != nil {
		t.Fatal(err)
	}
	s := &State{Scenes: scenes, MaxMessages: 8}
	loc := &u5data.Locations[britain-1]
	s.X, s.Y = loc.X, loc.Y

	s.Enter()
	if !s.InScene() {
		t.Fatalf("進不了不列顛城:%v", s.Messages)
	}
	// 入口那一格必須可走 —— 否則玩家一進城就卡在牆裡
	if u5data.TileBlocksWalking(int(s.TileAt(s.X, s.Y))) {
		t.Errorf("入口 (%d,%d) 的 tile %d 不可通行", s.X, s.Y, s.TileAt(s.X, s.Y))
	}
	s.Move(South) // 30 → 31
	s.Move(South) // 站在最外圈再往外 → 問是否離開
	if s.Prompt != PromptLeave {
		t.Fatalf("從南門走出去卻沒問是否離開:%v", s.Messages)
	}
	s.Answer(true)
	if s.InScene() || s.X != loc.X || s.Y != loc.Y {
		t.Errorf("離開後在 (%d,%d) 地點 %d,預期回到 (%d,%d) 大地圖", s.X, s.Y, s.Location, loc.X, loc.Y)
	}
}

// TestEveryLocationEnterable:32 個地點每一個都要進得去,且入口那一格可通行。
// 這是「正常玩家路徑」的最小保證 —— 進城卡在牆裡是 u2-cht 踩過的 soft-lock。
func TestEveryLocationEnterable(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	scenes, err := u5data.LoadSceneSet(dir)
	if err != nil {
		t.Fatal(err)
	}
	var stuck []string
	for i := range u5data.Locations {
		loc := &u5data.Locations[i]
		s := &State{Scenes: scenes, X: loc.X, Y: loc.Y, MaxMessages: 4}
		s.Enter()
		if !s.InScene() {
			t.Errorf("%s 進不去:%v", loc.Name, s.Messages)
			continue
		}
		if u5data.TileBlocksWalking(int(s.TileAt(s.X, s.Y))) {
			stuck = append(stuck, loc.Name)
		}
		// 每一層都要載得起來(含地下層)
		for f := loc.FloorMin; f <= loc.FloorMax; f++ {
			if _, err := scenes.Map(i+1, f); err != nil {
				t.Errorf("%s 第 %+d 層:%v", loc.Name, f, err)
			}
		}
	}
	if len(stuck) > 0 {
		t.Errorf("這些地點的入口格不可通行(玩家一進去就卡住):%v", stuck)
	}
}

// TestLadderChain 用原版資料驗證每一棟多層建築的梯子鏈。
//
// 原版的梯子是成對的:某層 (x,y) 有上行梯(0xC8),上一層的同一格就必須是下行梯(0xC9)。
// 這條性質**獨立於執行檔**,所以它同時驗了三件事:樓層範圍對不對、
// SceneSet.Map 的索引算式對不對、Klimb 的方向對不對。
func TestLadderChain(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	scenes, err := u5data.LoadSceneSet(dir)
	if err != nil {
		t.Fatal(err)
	}
	climbed := 0
	for i := range u5data.Locations {
		loc := &u5data.Locations[i]
		for f := loc.FloorMin; f < loc.FloorMax; f++ {
			m, err := scenes.Map(i+1, f)
			if err != nil {
				t.Fatalf("%s 第 %+d 層:%v", loc.Name, f, err)
			}
			for p := 0; p < u5data.SceneTiles; p++ {
				if m.Tiles[p] != u5data.LadderUp {
					continue
				}
				x, y := p%u5data.SceneSide, p/u5data.SceneSide
				s := &State{Scenes: scenes, MaxMessages: 4}
				if err := s.SetScene(i+1, f, x, y); err != nil {
					t.Fatal(err)
				}
				s.Klimb()
				if s.Floor != f+1 {
					t.Errorf("%s:站在第 %+d 層 (%d,%d) 的上行梯按 K,樓層變成 %+d,預期 %+d(%v)",
						loc.Name, f, x, y, s.Floor, f+1, s.Messages)
					continue
				}
				// 爬上去之後,腳下必須是「可以往下」的格子 —— 否則就下不來了。
				// 通常是下行梯 0xC9,但 ARARAT 用的是活板門 0x86
				// (原版 sub_EA0 把 134 和 201 當同一件事處理)。
				if got := s.TileAt(x, y); u5data.ClimbDelta(got) != -1 {
					t.Errorf("%s:從第 %+d 層 (%d,%d) 爬上去,落點 tile 是 %d,下不來了",
						loc.Name, f, x, y, got)
				}
				climbed++
			}
		}
	}
	if climbed == 0 {
		t.Fatal("一個梯子都沒爬到 —— 測試本身沒生效")
	}
	t.Logf("驗過 %d 段梯子", climbed)
}
