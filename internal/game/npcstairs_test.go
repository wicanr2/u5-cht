package game

import (
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// TestStairConstantsAgreeAcrossFiles —— ★★ 兩個檔案曾經給同一個值相反的名字。
//
// `sceneset.go` 說 `LadderUp = 0xC8`(而 `ClimbDelta` 對它回 +1),
// `tileflags.go` 曾經說 `TileStairsDown = 0xC8`。`sub_92C0` 站在前者那邊:
// 要往下找 0xC9、要往上找 0xC8。
func TestStairConstantsAgreeAcrossFiles(t *testing.T) {
	if u5data.TileStairsUp != u5data.LadderUp {
		t.Errorf("TileStairsUp = 0x%02X,LadderUp = 0x%02X —— 名字又分岔了",
			u5data.TileStairsUp, u5data.LadderUp)
	}
	if u5data.TileStairsDown != u5data.LadderDown {
		t.Errorf("TileStairsDown = 0x%02X,LadderDown = 0x%02X",
			u5data.TileStairsDown, u5data.LadderDown)
	}
	// 而「往上」那一個的 ClimbDelta 必須是 +1 —— 這是名字對不對的判準。
	if got := u5data.ClimbDelta(u5data.LadderUp); got != +1 {
		t.Errorf("LadderUp 的 ClimbDelta 是 %d,預期 +1", got)
	}
	if got := u5data.ClimbDelta(u5data.LadderDown); got != -1 {
		t.Errorf("LadderDown 的 ClimbDelta 是 %d,預期 −1", got)
	}
}

// TestNPCStairMaskIsWiderThanStairsFacing —— ★ 兩個遮罩不同,不要互相套用。
//
//	sub_92C0        `& 0xF4 == 0xC4` → 收 0xC4..0xC7 **與** 0xCC..0xCF
//	StairsFacing    `& 0xFC == 0xC4` → 只收 0xC4..0xC7
func TestNPCStairMaskIsWiderThanStairsFacing(t *testing.T) {
	if u5data.NPCStairMask != 0xF4 {
		t.Fatalf("NPCStairMask = 0x%02X,原版是 0xF4", u5data.NPCStairMask)
	}
	for _, tile := range []byte{0xC4, 0xC5, 0xC6, 0xC7} {
		if tile&u5data.NPCStairMask != u5data.NPCStairBase {
			t.Errorf("0x%02X 該算樓梯", tile)
		}
		if _, ok := u5data.StairsFacing(tile); !ok {
			t.Errorf("0x%02X 該是四朝向樓梯之一", tile)
		}
	}
	// ★ 0xCC..0xCF 只有寬的那個遮罩收。
	for _, tile := range []byte{0xCC, 0xCD, 0xCE, 0xCF} {
		if tile&u5data.NPCStairMask != u5data.NPCStairBase {
			t.Errorf("0x%02X 該被 0xF4 那個遮罩收進來", tile)
		}
		if _, ok := u5data.StairsFacing(tile); ok {
			t.Errorf("0x%02X 不該被 StairsFacing(0xFC)收 —— 兩個定義域不同", tile)
		}
	}
	// 反對照:樓梯以外的不能中。
	for _, tile := range []byte{0x00, 0x43, 0xA2, 0xD4, 0xE4} {
		if tile&u5data.NPCStairMask == u5data.NPCStairBase {
			t.Errorf("0x%02X 被誤判成樓梯", tile)
		}
	}
}

// TestNPCWalksToStairsInsteadOfTeleporting —— ★★ NPC 換樓層要先走到樓梯。
//
// 引擎此前直接改 `rt.Floor` ⇒ 玩家會看到 NPC 憑空消失、在另一層憑空出現。
func TestNPCWalksToStairsInsteadOfTeleporting(t *testing.T) {
	s := openCmdScene(t)
	if len(s.rtNPCs) == 0 {
		t.Skip("這個場景沒有 NPC")
	}
	// 找一個在場的 NPC,把它擺在一格**不是**樓梯的地板上,
	// 並且要求它換到另一層。
	i := -1
	for k := 1; k < len(s.rtNPCs); k++ {
		if s.npcs[k].Present() && s.rtNPCs[k].Mode != ModeAbsent {
			i = k
			break
		}
	}
	if i < 0 {
		t.Skip("沒有在場的 NPC")
	}
	rt := &s.rtNPCs[i]
	rt.Mode = ModeUpAway
	beforeFloor := rt.Floor
	// 確認它腳下不是樓梯(否則這條測不到「要先走過去」)。
	if s.npcOnUsableStair(i) {
		t.Skip("這個 NPC 剛好站在樓梯上")
	}
	s.stepNPCToStairs(i)
	if rt.Floor != beforeFloor {
		t.Errorf("沒站在樓梯上就換了樓層(%d → %d)—— 那是瞬移", beforeFloor, rt.Floor)
	}
}

// TestNPCOnStairChangesFloor —— 反對照:站對了就會換層。
//
// 少了這一條,「不瞬移」與「永遠不換層」用同一個觀察分不開 ——
// 而後者會讓 NPC 永遠回不到排程的樓層。
func TestNPCOnStairChangesFloor(t *testing.T) {
	s := openCmdScene(t)
	if len(s.rtNPCs) == 0 {
		t.Skip("這個場景沒有 NPC")
	}
	i := -1
	for k := 1; k < len(s.rtNPCs); k++ {
		if s.npcs[k].Present() && s.rtNPCs[k].Mode != ModeAbsent {
			i = k
			break
		}
	}
	if i < 0 {
		t.Skip("沒有在場的 NPC")
	}
	rt := &s.rtNPCs[i]
	rt.Mode = ModeUpAway
	// 把它腳下寫成「往上」的梯子 —— 而 `npcNeedsStairUp` 要真的是往上。
	want := byte(u5data.LadderDown)
	if s.npcNeedsStairUp(i) {
		want = u5data.LadderUp
	}
	if !s.SetTileAt(rt.X, rt.Y, want) {
		t.Skip("寫不進場景地圖")
	}
	if !s.npcOnUsableStair(i) {
		t.Fatalf("腳下寫成 0x%02X 了卻不算可用的樓梯", want)
	}
	target := u5data.SignedFloor(s.npcs[i].Schedule.Floor[rt.Slot])
	s.stepNPCToStairs(i)
	if rt.Floor != target {
		t.Errorf("站在樓梯上卻沒換到排程樓層:%d(預期 %d)", rt.Floor, target)
	}
	if rt.Mode != ModeIdle {
		t.Errorf("換完層之後模式是 %v,預期回到 ModeIdle", rt.Mode)
	}
}

// grassTile 是草地(`look#5` = grass)。
const grassTile = 0x05

// TestDarknessTileBlindsOnTheOverworld —— ★★ 站在「黑暗」格上視野歸零。
//
// tile 0xFF 是 look 表第 255 筆「darkness!」,在 `UNDER.DAT` 出現 106 次、
// `BRIT.DAT` 一次都沒有 ⇒ 幽冥界專屬。
func TestDarknessTileBlindsOnTheOverworld(t *testing.T) {
	s := worldState(t)
	if s.World == nil {
		t.Skip("沒有世界地圖")
	}
	s.Location, s.Floor = 0, 0
	s.Clock.Hour = 12 // 正午,本來最亮
	s.X, s.Y = 60, 60
	if !s.SetTileAt(s.X, s.Y, grassTile) {
		t.Skip("寫不進世界地圖")
	}
	bright := s.LightRadius2()
	if bright == 0 {
		t.Skip("這個狀態本來就是全黑")
	}
	if !s.SetTileAt(s.X, s.Y, u5data.TileDarkness) {
		t.Skip("寫不進世界地圖")
	}
	if got := s.LightRadius2(); got != 0 {
		t.Errorf("站在黑暗格上半徑² 是 %d,預期 0(正常時是 %d)", got, bright)
	}
	// ★ 那個豁免模式要真的豁免。
	s.CombatMode = DarknessExemptMode
	if got := s.LightRadius2(); got != bright {
		t.Errorf("模式 0x%02X 下半徑² 是 %d,預期照原本算出 %d",
			DarknessExemptMode, got, bright)
	}
}

// TestDarknessOnlyAppliesOnTheOverworld —— ★★ 只在大地圖上判。
//
// `TileAt` 在場景資料缺失或座標出界時回的 `TileBlank` **也是 0xFF** ——
// 同一個值兩個意思。少了這道閘門,任何沒載入場景的狀態都會變成全黑。
func TestDarknessOnlyAppliesOnTheOverworld(t *testing.T) {
	if u5data.TileDarkness != u5data.TileBlank {
		t.Skip("兩個常數已經不同值了,這條的前提消失")
	}
	s := newState(t)
	s.Clock.Hour = 12
	s.Location = 2 // 場景 ⇒ TileAt 會回 TileBlank
	if got := s.LightRadius2(); got == 0 {
		t.Error("場景裡沒有地圖資料就被判成全黑 —— 那是把 TileBlank 當 darkness")
	}
}
