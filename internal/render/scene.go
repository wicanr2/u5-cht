package render

import (
	"fmt"
	"image"
	"image/color"

	"github.com/wicanr2/u5-cht/internal/game"
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 畫面配置
//
// 原版 320×200 EGA 是「左側 11×11 tile 地圖 + 右側狀態 + 下方訊息」。
// 本專案的邏輯畫布是 640×400(原版的乾淨 2×),tile 用 2 倍 nearest 畫成 32 px,
// 地圖視窗仍是 11×11 tile —— 佔畫面比例與原版一致,而中文有空間用正常字級
// (rulebook/81「拉畫布別縮字」)。
const (
	CanvasWidth  = 640
	CanvasHeight = 400

	ViewTiles  = 11
	TileScale  = 2
	TilePixels = u5data.TileSize * TileScale
	ViewPixels = ViewTiles * TilePixels // 352

	MapOriginX = 8
	MapOriginY = 8

	// 右側直欄:狀態在上、對話與訊息在下。
	//
	// 原版就是這個配置 —— 320×200 下地圖佔左邊 (8,8)-(183,183),文字走右邊那一欄
	// (FM Towns 版的視窗矩形表 dword_3DD6C 也是同一個形狀)。
	// 一開始把訊息放在地圖下方,結果 400 − 360 = 40px 只塞得下兩行,
	// 對話一長就直接畫到畫布外。
	PanelX     = MapOriginX + ViewPixels + 8 // 368
	PanelWidth = CanvasWidth - PanelX - 8    // 264
	// PanelTextY 是右欄裡「狀態結束、對話開始」的分界。
	PanelTextY = MapOriginY + LineHeight*11

	// 地圖下方剩下的一條給操作提示。
	HintY = MapOriginY + ViewPixels + 6
)

var (
	ColorBackground = color.NRGBA{R: 0x10, G: 0x10, B: 0x28, A: 0xFF}
	ColorText       = color.NRGBA{R: 0xE8, G: 0xE8, B: 0xD8, A: 0xFF}
	ColorMarker     = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
)

// Scene 是「畫面上該有什麼」的完整描述。它不知道 ebiten 存在。
//
// 這是 headless 驗證與實機顯示的**共同來源**:兩邊都呼叫 Render(),
// 所以截圖就是實機畫面,不會漂移。
type Scene struct {
	State *game.State
	Tiles []u5data.Tile
	Text  *TextRenderer
}

// Render 畫出整個 640×400 畫面。
func (s *Scene) Render() *image.NRGBA {
	dst := image.NewNRGBA(image.Rect(0, 0, CanvasWidth, CanvasHeight))
	fill(dst, ColorBackground)

	s.drawMapView(dst)
	s.drawPanel(dst)
	s.drawMessages(dst)
	s.drawHints(dst)
	return dst
}

func (s *Scene) drawMapView(dst *image.NRGBA) {
	if s.State == nil || len(s.Tiles) == 0 {
		return
	}
	// 11×11 視窗永遠以玩家為中心 —— 原版就是這樣:
	// 場景移動函式 sub_86C 讀鄰格用的 byte_3F789[32*dy+dx] 是一個固定位址,
	// 也就是視窗緩衝裡玩家那一格,四鄰用 ±1 / ±32 直接定址(docs/re/03 §7)。
	half := ViewTiles / 2
	// 戰鬥中:11×11 的戰場正好等於視窗大小,直接鋪滿。
	if s.State.InCombat() {
		s.drawCombat(dst)
		return
	}
	for dy := -half; dy <= half; dy++ {
		for dx := -half; dx <= half; dx++ {
			s.drawTile(dst, int(s.State.TileAt(s.State.X+dx, s.State.Y+dy)),
				MapOriginX+(dx+half)*TilePixels,
				MapOriginY+(dy+half)*TilePixels)
		}
	}
	// NPC 疊在地形上。原版把 NPC 寫進同一個 11×11 視窗緩衝再一起畫,
	// 效果等同於「先地形後 NPC」。
	for _, n := range s.State.VisibleNPCs() {
		dx, dy := n.X-s.State.X, n.Y-s.State.Y
		if dx < -half || dx > half || dy < -half || dy > half {
			continue
		}
		s.drawTile(dst, n.Tile,
			MapOriginX+(dx+half)*TilePixels,
			MapOriginY+(dy+half)*TilePixels)
	}

	// 大地圖的物件(坐騎、船、地上的東西、遊蕩的怪物)也疊在地形上。
	for _, o := range s.State.VisibleObjects() {
		dx, dy := o.X-s.State.X, o.Y-s.State.Y
		if dx < -half || dx > half || dy < -half || dy > half {
			continue
		}
		s.drawTile(dst, o.Tile,
			MapOriginX+(dx+half)*TilePixels,
			MapOriginY+(dy+half)*TilePixels)
	}

	// 玩家位置標記。
	// TODO(P3):換成原版的 Avatar tile —— 索引要從反編譯碼確認,不猜。
	DrawFrame(dst,
		MapOriginX+half*TilePixels, MapOriginY+half*TilePixels,
		TilePixels, TilePixels, ColorMarker)
}

// drawTile 以 TileScale 倍(nearest)把一個 tile 畫到 (x, y)。
func (s *Scene) drawTile(dst *image.NRGBA, idx, x, y int) {
	if idx < 0 || idx >= len(s.Tiles) {
		idx = 0
	}
	t := &s.Tiles[idx]
	for ty := 0; ty < u5data.TileSize; ty++ {
		for tx := 0; tx < u5data.TileSize; tx++ {
			// 用 TilePalette(已套色號 IGRB→IRGB remap),不要用 EGAPalette。
			c := u5data.TilePalette[t.At(tx, ty)&0x0F]
			for sy := 0; sy < TileScale; sy++ {
				for sx := 0; sx < TileScale; sx++ {
					SetPixel(dst, x+tx*TileScale+sx, y+ty*TileScale+sy, c)
				}
			}
		}
	}
}

func (s *Scene) drawPanel(dst *image.NRGBA) {
	if s.Text == nil {
		return
	}
	st := s.State
	if st == nil {
		return
	}
	y := MapOriginY
	y = s.Text.DrawLines(dst, PanelX, y, []string{"創世紀 V", "命運勇士"})
	y += LineHeight / 2
	s.Text.Draw(dst, PanelX, y, fmt.Sprintf("%s  業報 %2d", st.Clock, st.Karma))
	y += LineHeight
	s.Text.Draw(dst, PanelX, y, fmt.Sprintf("座標 %3d,%3d  地形 %3d", st.X, st.Y, st.TileAt(st.X, st.Y)))
	y += LineHeight
	if st.InScene() {
		s.Text.Draw(dst, PanelX, y, fmt.Sprintf("★ %s 第 %d 層  居民 %d",
			st.LocationName(), st.Floor+1, len(st.VisibleNPCs())))
	} else {
		// 地點名取自原版執行檔的地點表(u5data.Locations),不是自己打的清單。
		line := "不列顛尼亞"
		if st.Floor < 0 {
			line = "地下世界"
		}
		if loc, ok := u5data.LocationAt(st.X, st.Y); ok {
			line = "★ " + loc.DisplayName()
		}
		s.Text.Draw(dst, PanelX, y, line)
	}

	// 隊伍:名字 + HP。原版右欄就是這個位置。
	y += LineHeight
	for _, c := range st.Party() {
		s.Text.Draw(dst, PanelX, y, fmt.Sprintf("%-9s %-4s %3d/%-3d",
			c.Name, c.ClassName(), c.HP, c.MaxHP))
		y += LineHeight
	}

	// 分隔線:狀態與對話之間
	for x := PanelX; x < PanelX+PanelWidth; x++ {
		SetPixel(dst, x, PanelTextY-LineHeight/2, ColorText)
	}
}

// drawHints 在地圖下方畫操作提示。
func (s *Scene) drawHints(dst *image.NRGBA) {
	if s.Text == nil || s.State == nil {
		return
	}
	hint := "方向鍵移動   E 進入   K 攀爬   T 交談   F10 離開"
	switch s.State.Prompt {
	case game.PromptLeave:
		hint = "Y 是 / N 否"
	case game.PromptTalk:
		hint = "輸入關鍵字後按 Enter,打 bye 或 ESC 結束"
	case game.PromptAnswer:
		hint = "回答對方的問題後按 Enter"
	}
	s.Text.Draw(dst, MapOriginX, HintY, hint)
}

// drawInputLine 在對話中畫出玩家正在打的字,後面帶一個游標。
// 沒有這一行,玩家打字時畫面完全沒反應 —— 會以為鍵盤壞了。
func (s *Scene) drawInputLine(dst *image.NRGBA, y int) int {
	if s.State == nil {
		return y
	}
	label := "汝問:"
	switch s.State.Prompt {
	case game.PromptTalk:
	case game.PromptAnswer:
		label = "汝答:"
	default:
		return y
	}
	s.Text.Draw(dst, PanelX, y, label+s.State.Input+"_")
	return y + LineHeight
}

func (s *Scene) drawMessages(dst *image.NRGBA) {
	if s.Text == nil {
		return
	}
	if s.State == nil {
		return
	}
	// 從下往上排:先算出全部行數,不夠高就只留最新的幾行 ——
	// 訊息比空間多的時候,該掉的是最舊的,不是最新的。
	var lines []string
	for _, m := range s.State.Messages {
		lines = append(lines, Wrap(m, PanelWidth)...)
	}
	avail := (CanvasHeight - 8 - PanelTextY) / LineHeight
	if s.State.Prompt == game.PromptTalk || s.State.Prompt == game.PromptAnswer {
		avail-- // 留一行給輸入列
	}
	if avail < 1 {
		avail = 1
	}
	if len(lines) > avail {
		lines = lines[len(lines)-avail:]
	}
	y := s.Text.DrawLines(dst, PanelX, PanelTextY, lines)
	s.drawInputLine(dst, y)
}

func fill(dst *image.NRGBA, c color.NRGBA) {
	b := dst.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			dst.SetNRGBA(x, y, c)
		}
	}
}

// WrapCoord 把座標折回世界地圖範圍(Britannia 是環繞的)。
func WrapCoord(v int) int {
	v %= u5data.WorldSide
	if v < 0 {
		v += u5data.WorldSide
	}
	return v
}

// drawCombat 畫戰場。11×11 與視窗同大,所以座標直接對應。
func (s *Scene) drawCombat(dst *image.NRGBA) {
	c := s.State.Combat
	for y := 0; y < ViewTiles; y++ {
		for x := 0; x < ViewTiles; x++ {
			s.drawTile(dst, int(c.Map.At(x, y)),
				MapOriginX+x*TilePixels, MapOriginY+y*TilePixels)
		}
	}
	for i := range c.Units {
		u := &c.Units[i]
		if !u.Active() {
			continue
		}
		s.drawTile(dst, u.Tile, MapOriginX+u.X*TilePixels, MapOriginY+u.Y*TilePixels)
	}
	// 目前輪到誰,用外框標出來。
	if c.Turn >= 0 && c.Turn < len(c.Units) {
		u := &c.Units[c.Turn]
		DrawFrame(dst, MapOriginX+u.X*TilePixels, MapOriginY+u.Y*TilePixels,
			TilePixels, TilePixels, ColorMarker)
	}
}
