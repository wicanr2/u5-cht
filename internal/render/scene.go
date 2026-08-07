package render

import (
	"fmt"
	"image"
	"image/color"

	"github.com/wicanr2/u5-cht/internal/game"
	"github.com/wicanr2/u5-cht/internal/i18n"
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

// EGABlack 是地牢視野外圍的底色。
var EGABlack = color.NRGBA{A: 0xFF}

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
	// DungeonViews 是 DNG1/2/3.16 的透視切片。空的話地牢退回俯視圖。
	DungeonViews []u5data.PictureSet
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
	// In Quas Wis 的全景:32×32 一次攤開,蓋掉平常的 11×11。
	if s.State.Prompt == game.PromptPeer {
		s.drawPeer(dst)
		return
	}
	// 戰鬥中:11×11 的戰場正好等於視窗大小,直接鋪滿。
	if s.State.InCombat() {
		s.drawCombat(dst)
		return
	}
	// 地牢:8×8 俯視。⚠ 原版是第一人稱透視 —— 那要先解 DNG1-3.16,還沒做。
	if s.State.InDungeon() {
		s.drawDungeon(dst)
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
			// 色號在載入時就正規化成標準 EGA 了(見 u5data.tileColorRemap),這裡直接查表。
			c := u5data.EGAPalette[t.At(tx, ty)&0x0F]
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
			i18n.Name(c.Name), c.ClassName(), c.HP, c.MaxHP))
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

// fillRect 把一塊矩形填成單色。
func fillRect(dst *image.NRGBA, x, y, w, h int, c color.NRGBA) {
	for py := y; py < y+h; py++ {
		for px := x; px < x+w; px++ {
			SetPixel(dst, px, py, c)
		}
	}
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

// 地牢的俯視畫面。
//
// ⚠ **這不是原版的樣子。** 原版的地牢是第一人稱透視圖(素材在 `DNG1-3.16`),
// 那三組圖還沒解碼。規則(移動、朝向、爬梯、陷阱、房間)是照原版的,
// 呈現方式是暫時的 —— 解完透視圖之後這一段會整個換掉。
//
// 8×8 的地牢畫在 11×11 視窗的中央,朝向用一個外框加箭頭表示。
func (s *Scene) drawDungeon(dst *image.NRGBA) {
	if s.drawDungeonFirstPerson(dst) {
		return
	}
	d := s.State.Dungeon
	const side = u5data.DungeonSide
	off := (ViewTiles - side) / 2
	for y := 0; y < side; y++ {
		for x := 0; x < side; x++ {
			s.drawTile(dst, dungeonViewTile(s.State.DungeonTileAt(x, y)),
				MapOriginX+(off+x)*TilePixels, MapOriginY+(off+y)*TilePixels)
		}
	}
	// 隊伍與朝向。
	px := MapOriginX + (off+d.X)*TilePixels
	py := MapOriginY + (off+d.Y)*TilePixels
	s.drawTile(dst, u5data.NPCTileBase+u5data.VehicleWalk, px, py)
	DrawFrame(dst, px, py, TilePixels, TilePixels, ColorMarker)
}

// DungeonViewScale 是透視畫面的放大倍率。
//
// 原版是 160×164;畫布是原版的乾淨 2×,所以這裡也是 2 → 320×328,
// 擺進 352×352 的地圖窗還有餘裕。
const DungeonViewScale = 2

// drawDungeonFirstPerson 畫原版的第一人稱透視走廊。切片沒載到就回 false。
//
// 畫的順序照 `sub_3D14`:
//
//	由近而遠 —— 每一階先看正面擋不擋得住(擋住就畫正面並停),再畫左右側牆
//
// ⚠ 深度 0 是**腳下這一格**,不是前面那一格。搞錯會整個往前偏一格,
// 而畫面看起來仍然「像個走廊」,不會當場穿幫。
func (s *Scene) drawDungeonFirstPerson(dst *image.NRGBA) bool {
	st := s.State
	d := st.Dungeon
	if d == nil || len(s.DungeonViews) == 0 {
		return false
	}
	theme := u5data.DungeonTheme(d.Location) - 1
	if theme < 0 || theme >= len(s.DungeonViews) {
		return false
	}
	set := s.DungeonViews[theme]

	const vw, vh = u5data.DungeonViewWidth, u5data.DungeonViewHeight
	ox := MapOriginX + (ViewPixels-vw*DungeonViewScale)/2
	oy := MapOriginY + (ViewPixels-vh*DungeonViewScale)/2
	fillRect(dst, ox, oy, vw*DungeonViewScale, vh*DungeonViewScale, EGABlack)

	// 沒有光就是全黑 —— `sub_3D14` 一開頭 `if (!byte_3E0B6 && !byte_3E0B7)`
	// 直接畫一個黑框收工。地牢裡沒火把是真的什麼都看不到。
	if !st.HasLight() {
		return true
	}

	fdx, fdy := d.Facing.Delta()
	// 側向 = 朝向轉 90 度(`byte_4FFA4` / `byte_4FFA8`)。
	sdx, sdy := -fdy, fdx

	x, y := d.X, d.Y
	for depth := 0; depth < u5data.DungeonViewDepths; depth++ {
		tile := st.DungeonTileAt(u5data.DungeonWrap(x), u5data.DungeonWrap(y))
		if !u5data.DungeonSeeThrough(tile, depth) {
			if n := u5data.DungeonFrontShape(tile, depth); n >= 0 {
				s.blitSlice(dst, set, n, ox, oy, depth)
			}
			break
		}
		// 站在門口時不畫腳下這一格的側牆 —— 門框就在身邊,擋住了。
		if depth > 0 || u5data.DungeonKind(tile) != u5data.DungeonDoorway {
			for _, sign := range [2]int{1, -1} {
				sx := u5data.DungeonWrap(x + sdx*sign)
				sy := u5data.DungeonWrap(y + sdy*sign)
				n := u5data.DungeonSideShape(st.DungeonTileAt(sx, sy), depth)
				s.blitSliceSide(dst, set, n, ox, oy, depth, sign < 0)
			}
		}
		x, y = x+fdx, y+fdy
	}
	return true
}

// blitSlice 把一張切片畫在左右兩半(右半水平鏡射)。
func (s *Scene) blitSlice(dst *image.NRGBA, set u5data.PictureSet, n, ox, oy, depth int) {
	s.blitSliceSide(dst, set, n, ox, oy, depth, false)
	s.blitSliceSide(dst, set, n, ox, oy, depth, true)
}

// blitSliceSide 畫半邊。mirror 為真時畫右半(水平翻轉)。
func (s *Scene) blitSliceSide(dst *image.NRGBA, set u5data.PictureSet, n, ox, oy, depth int, mirror bool) {
	if n < 0 || n >= len(set) || set[n] == nil {
		return
	}
	p := set[n]
	bx := u5data.DungeonBandX(depth)
	for py := 0; py < p.Height; py++ {
		for px := 0; px < p.Width; px++ {
			vx := bx + px
			if mirror {
				vx = u5data.DungeonViewWidth - 1 - vx
			}
			c := u5data.EGAPalette[p.Pix[py*p.Width+px]&0x0F]
			for ky := 0; ky < DungeonViewScale; ky++ {
				for kx := 0; kx < DungeonViewScale; kx++ {
					SetPixel(dst, ox+vx*DungeonViewScale+kx, oy+py*DungeonViewScale+ky, c)
				}
			}
		}
	}
}

// dungeonViewTile 把地牢格子換成一個看得懂的世界 tile。
//
// 只是為了讓俯視圖分得出通道 / 牆 / 梯子 / 房間,**不是原版的對應** ——
// 原版根本不畫俯視圖。
func dungeonViewTile(t byte) int {
	switch u5data.DungeonKind(t) {
	case u5data.DungeonWall, u5data.DungeonUnknownC,
		u5data.DungeonUnknownD, u5data.DungeonDoorway:
		return 0x0E // 山:走不過去
	case u5data.DungeonLadderUp:
		return 0xC8 // 上行梯
	case u5data.DungeonLadderDown:
		return 0xC9 // 下行梯
	case u5data.DungeonLadderBoth:
		return 0xC8
	case u5data.DungeonChest:
		return 0x101 // 寶箱
	case u5data.DungeonFountain:
		return 0x01 // 水
	case u5data.DungeonTrap, u5data.DungeonMagic:
		return 0x8F // 有東西不對勁
	case u5data.DungeonDoor:
		return 0x44 // 門
	case u5data.DungeonRoomA, u5data.DungeonRoomF:
		return 0x43 // 房間入口
	}
	return 0x04 // 通道:草地當地板
}

// PeerTilePixels 是全景模式下一格佔幾個像素。
//
// 352 / 32 = 11 —— 地圖視窗的大小不變,只是塞進 32×32 格。原版在
// 320×200 下是同一個做法(`sub_24824(16, 80, 367, 431)` 圈出的正是地圖窗,
// 而裡面畫的是 32×32),所以縮小倍率與原版一致。
const PeerTilePixels = ViewPixels / game.PeerSide

// drawPeer 畫 In Quas Wis 的 32×32 全景。
//
// 每一格用 nearest 縮到 11×11 —— tile 是 16×16,縮小倍率不是整數,
// 但這是原版就有的取捨(它也是把 16×16 塞進 11 px),照做。
func (s *Scene) drawPeer(dst *image.NRGBA) {
	half := game.PeerSide / 2
	for dy := -half; dy < half; dy++ {
		for dx := -half; dx < half; dx++ {
			s.drawTileScaled(dst, int(s.State.PeerTile(dx, dy)),
				MapOriginX+(dx+half)*PeerTilePixels,
				MapOriginY+(dy+half)*PeerTilePixels,
				PeerTilePixels)
		}
	}
	DrawFrame(dst,
		MapOriginX+half*PeerTilePixels, MapOriginY+half*PeerTilePixels,
		PeerTilePixels, PeerTilePixels, ColorMarker)
}

// drawTileScaled 把一個 tile 用 nearest 取樣畫成 size×size。
func (s *Scene) drawTileScaled(dst *image.NRGBA, idx, x, y, size int) {
	if idx < 0 || idx >= len(s.Tiles) {
		return
	}
	t := &s.Tiles[idx]
	for py := 0; py < size; py++ {
		sy := py * u5data.TileSize / size
		for px := 0; px < size; px++ {
			sx := px * u5data.TileSize / size
			SetPixel(dst, x+px, y+py, u5data.EGAPalette[t.At(sx, sy)&0x0F])
		}
	}
}
