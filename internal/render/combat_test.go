package render

import (
	"testing"

	"github.com/wicanr2/u5-cht/internal/game"
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// TestCombatPartyTileUsesCreaturePage 是 issue1 的算繪端回歸：
// raw 0x4C 與最終 0x14C 故意做成不同顏色，漏加 NPCTileBase 就會立刻失敗。
func TestCombatPartyTileUsesCreaturePage(t *testing.T) {
	const raw = 0x4C
	final := u5data.NPCTileBase + raw
	tiles := make([]u5data.Tile, final+1)
	for i := range tiles[raw].Pix {
		tiles[raw].Pix[i] = 1    // raw namespace：藍色錯誤圖
		tiles[final].Pix[i] = 14 // creature namespace：黃色隊伍圖
	}

	st := &game.State{Prompt: game.PromptCombat}
	st.Combat = &game.Combat{Map: &u5data.CombatMap{}, Turn: -1}
	st.Combat.Units[0] = game.Combatant{
		Tile:  final,
		X:     5,
		Y:     5,
		Flags: game.UnitParty,
	}

	img := (&Scene{State: st, Tiles: tiles, UI: UIModern}).Render()
	x := MapOriginX + 5*TilePixels + TilePixels/2
	y := MapOriginY + 5*TilePixels + TilePixels/2
	if got := img.NRGBAAt(x, y); got != u5data.EGAPalette[14] {
		t.Fatalf("戰鬥隊伍中心像素是 %v,預期生物頁 tile %d 的 %v", got, final, u5data.EGAPalette[14])
	}
}
