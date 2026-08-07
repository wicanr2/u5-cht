package game

import (
	"os"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 王冠與權杖真的撿得到 —— 用原版資料走到密室裡按 G。
//
// 這條測的是「知道位置」與「機制跑得通」是兩件事:引擎這邊一行都沒為信物加過
// 特例(它們就是兩個行為型別 0 的 NPC,由 `syncNPCObjects` 鏡射進物件表),
// 所以這條紅掉代表鏡射或 Get 的泛用路徑壞了,不是信物本身的問題。
func TestRegaliaCanBePickedUpWhereTheyStand(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	scenes, err := u5data.LoadSceneSet(dir)
	if err != nil {
		t.Fatal(err)
	}
	npcs, err := u5data.LoadNPCSet(dir)
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range u5data.RegaliaNPCPlacement {
		s := &State{Scenes: scenes, NPCs: npcs, Clock: NewClock(), MaxMessages: 32}
		// 站在信物北邊一格,面向南 —— 密室四周是牆,但 Get 只看目標格。
		if err := s.SetScene(want.Location, want.Floor, want.X, want.Y-1); err != nil {
			t.Fatalf("%s:進不了地點 %d 第 %d 層:%v", want.Name, want.Location, want.Floor, err)
		}
		slot, o := s.pickableAt(want.X, want.Y)
		if o == nil {
			t.Errorf("%s:(%d,%d) 上沒有撿得起來的東西 —— 鏡射沒跑?",
				want.Name, want.X, want.Y)
			continue
		}
		if o.Kind != want.Kind {
			t.Errorf("%s:(%d,%d) 上的是種類 0x%02X,預期 0x%02X",
				want.Name, want.X, want.Y, o.Kind, want.Kind)
			continue
		}
		s.getAt(0, 1)
		switch want.Kind {
		case u5data.ItemCrown:
			if !s.Regalia.Crown {
				t.Errorf("按了 G 卻沒拿到王冠:%v", s.Messages)
			}
		case u5data.ItemSceptre:
			if !s.Regalia.Sceptre {
				t.Errorf("按了 G 卻沒拿到權杖:%v", s.Messages)
			}
		}
		// 撿完那一格就空了。
		if _, still := s.pickableAt(want.X, want.Y); still != nil {
			t.Errorf("%s:撿完之後物件槽 %d 還在", want.Name, slot)
		}
	}
}

// 王冠會永久消失,權杖不會 —— 原版只有王冠配了 `sub_218`。
//
// ⚠ 這不是 bug,是照抄。權杖可以刷,與不列顛王城堡二樓那張魔毯同一種行為。
func TestOnlyTheCrownIsRemovedForGood(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	scenes, err := u5data.LoadSceneSet(dir)
	if err != nil {
		t.Fatal(err)
	}
	npcs, err := u5data.LoadNPCSet(dir)
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range u5data.RegaliaNPCPlacement {
		s := &State{Scenes: scenes, NPCs: npcs, Clock: NewClock(), MaxMessages: 32}
		if err := s.SetScene(want.Location, want.Floor, want.X, want.Y-1); err != nil {
			t.Fatal(err)
		}
		s.getAt(0, 1)
		// 離開再回來。
		if err := s.SetScene(want.Location, want.Floor, want.X, want.Y-1); err != nil {
			t.Fatal(err)
		}
		_, o := s.pickableAt(want.X, want.Y)
		gone := o == nil
		wantGone := want.Kind == u5data.ItemCrown
		if gone != wantGone {
			verb := map[bool]string{true: "不見了", false: "還在原地"}
			t.Errorf("%s:回來之後%s,原版是%s",
				want.Name, verb[gone], verb[wantGone])
		}
	}
}
