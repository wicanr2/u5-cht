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
	// DungeonItems 是 ITEMS.16 —— 走廊裡的梯子、寶箱、噴泉、陷阱。
	DungeonItems u5data.PictureSet
	// IntroArt 是 STORY1-6.16 —— 開場的插圖。
	IntroArt []u5data.PictureSet
}

// Render 畫出整個 640×400 畫面。
func (s *Scene) Render() *image.NRGBA {
	dst := image.NewNRGBA(image.Rect(0, 0, CanvasWidth, CanvasHeight))
	fill(dst, ColorBackground)

	// 開場動畫佔滿整個畫面 —— 沒有狀態欄也沒有地圖窗。
	if s.State != nil && s.State.Intro != nil {
		s.drawIntro(dst)
		return dst
	}
	// 主選單也佔滿整個畫面 —— 原版是在那張「窗外景色」上疊選單,
	// 這一層還沒有那張圖,先用單獨一頁,不要疊在地圖上讓玩家分心。
	if s.State != nil && s.State.InMainMenu() {
		s.drawMainMenu(dst)
		s.drawHints(dst)
		return dst
	}
	// 角色數值也是整頁 —— 原版 Ztats 就是蓋掉地圖窗的一整面。
	if s.State != nil && s.State.Prompt == game.PromptZtats {
		s.drawZtats(dst)
		s.drawHints(dst)
		return dst
	}
	// 選單同樣整頁。原版是在地圖旁邊的側欄列候選,但側欄只有 ~16 欄寬,
	// 中文一列放得下 8 個字 —— 裝備名放不進去。整頁才擺得開。
	if s.State != nil && s.State.Prompt == game.PromptPick {
		s.drawLines(dst, s.State.PickLines())
		s.drawHints(dst)
		return dst
	}
	s.drawMapView(dst)
	s.drawPanel(dst)
	s.drawMessages(dst)
	s.drawHints(dst)
	return dst
}

// drawMainMenu 畫主選單的六個項目。游標那一項前面加箭頭 ——
// 反白需要另一套繪製路徑,而箭頭在點陣字下反而更清楚。
func (s *Scene) drawMainMenu(dst *image.NRGBA) {
	if s.Text == nil || s.State.Menu == nil {
		return
	}
	const top = 96
	s.Text.Draw(dst, MapOriginX, top-LineHeight*2, "創世紀 V:命運勇士")
	for i, label := range game.MenuLabels {
		mark := "  "
		if game.MenuItem(i) == s.State.Menu.Cursor {
			mark = "→"
		}
		s.Text.Draw(dst, MapOriginX, top+i*LineHeight, mark+label)
	}
	// 選了未實作的項目時訊息會留在 Messages 裡 —— 選單頁沒有訊息欄,
	// 所以在這裡把最後一句畫出來,不然玩家按了 Enter 完全沒有反應。
	if n := len(s.State.Messages); n > 0 {
		s.Text.Draw(dst, MapOriginX, top+int(game.MenuItemCount)*LineHeight+LineHeight,
			s.State.Messages[n-1])
	}
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
	// 視線遮蔽:被牆擋住、或落在光照之外的格子不畫(原版 `sub_2DDB0`)。
	// mask 為 nil 代表這個場合不做遮蔽(戰鬥、石室)。
	mask := s.State.SightMask()
	for dy := -half; dy <= half; dy++ {
		for dx := -half; dx <= half; dx++ {
			if !game.SightVisible(mask, dx, dy) {
				continue
			}
			s.drawTerrain(dst, int(s.State.TileAt(s.State.X+dx, s.State.Y+dy)),
				MapOriginX+(dx+half)*TilePixels,
				MapOriginY+(dy+half)*TilePixels)
		}
	}
	// NPC 疊在地形上。原版把 NPC 寫進同一個 11×11 視窗緩衝再一起畫,
	// 效果等同於「先地形後 NPC」。
	//
	// ⚠ 看不到的格子上的 NPC 也不畫 —— 原版是把 NPC 寫進**同一個罩子**,
	// 而罩子上那一格已經是 0xFF。少了這一段,黑暗裡會浮出一排無主的小人。
	for _, n := range s.State.VisibleNPCs() {
		dx, dy := n.X-s.State.X, n.Y-s.State.Y
		if dx < -half || dx > half || dy < -half || dy > half {
			continue
		}
		if !game.SightVisible(mask, dx, dy) {
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
		if !game.SightVisible(mask, dx, dy) {
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
// drawTerrain 畫地形格,並處理月門的長出動畫(原版 `sub_29BEC` → `sub_265F0`)。
//
// 原版的地圖重畫在畫每一格之前多問一句:
//
//	if (視窗緩衝[格] == 0DCh && byte_3E097 != 0 && byte_3E097 < 10h)
//	    sub_265F0(byte_3E097, 欄, 列)      ; ★ 中間格
//	else 一般畫法                            ; 0 或 >= 16 → 直接畫 tile
//
// ⇒ 計數器**滿格(0x10)才畫完整的月門**,1..15 是長出的過程。
// 而計數器為 0 時走一般畫法 —— 那正好是「離開載入視窗、tile 沒被寫回草地」
// 的殘留(`docs/re/86`):殘留的月門會被畫成**完整的門**。
func (s *Scene) drawTerrain(dst *image.NRGBA, idx, x, y int) {
	if idx != int(u5data.MoongateOpenTile) || s.State == nil {
		s.drawTile(dst, idx, x, y)
		return
	}
	f := s.State.MoongateFrame
	if f <= 0 || f >= game.MoongateFrameMax {
		s.drawTile(dst, idx, x, y)
		return
	}
	s.drawMoongateFrame(dst, f, x, y)
}

// drawMoongateFrame 疊出第 f 格的月門(原版 `sub_265F0`)。
//
// 組語裡是兩次 `rep movs` 疊在一個 512 byte 的暫存格上:
//
//	ebx = 基址 + 0A00h                  ; 0x0A00 / 0x200 = tile **5(草地)**
//	var_224 = 基址 + 1B800h             ; 0x1B800 / 0x200 = tile **0DCh(月門)**
//	整格填草地(rep movsd,ecx = 80h = 512 byte)
//	dest = 暫存格 + (20h − 2*f) * 10h   ; = 第 (16 − f) 列
//	copy f*20h byte                      ; = 月門圖的**前 f 列**
//
// ⇒ **底下先鋪草地,月門圖的前 f 列疊在第 (16−f) 列起** —— 門由下往上長出來,
// 而露出來的一直是門圖的上半部。f = 16 時就等於整格月門(走一般畫法)。
//
// ★ 每格 0x200 byte、每列 0x20 byte ⇒ 16×16 × 2 byte/px,正好是 FM Towns 的
// 16-bit 直色。這順帶佐證了「0x1B800 / 0x200 = 0xDC」不是巧合。
//
// ⚠ `offset sub_1B800` 是 **IDA 的自動命名,不是函式** —— 它被當成數值用。
// 這是本專案第四次踩到「`aXxx` / `sub_Xxxx` 自動名不是資料」(`CLAUDE.md §4.5`)。
func (s *Scene) drawMoongateFrame(dst *image.NRGBA, f, x, y int) {
	gate := int(u5data.MoongateOpenTile)
	grass := int(u5data.MoongateClosedTile)
	if gate >= len(s.Tiles) || grass >= len(s.Tiles) {
		return
	}
	g, m := &s.Tiles[grass], &s.Tiles[gate]
	top := u5data.TileSize - f
	for ty := 0; ty < u5data.TileSize; ty++ {
		for tx := 0; tx < u5data.TileSize; tx++ {
			// 第 (16−f) 列起換成月門圖的第 (ty − (16−f)) 列。
			px := g.At(tx, ty)
			if ty >= top {
				px = m.At(tx, ty-top)
			}
			c := u5data.EGAPalette[px&0x0F]
			for sy := 0; sy < TileScale; sy++ {
				for sx := 0; sx < TileScale; sx++ {
					SetPixel(dst, x+tx*TileScale+sx, y+ty*TileScale+sy, c)
				}
			}
		}
	}
}

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

	// ★ 原版狀態列的四樣東西(`sub_2A1E8` / `sub_2A0C4` / `sub_2A984`),
	// 引擎的右欄**一樣都沒畫**(`docs/re/80`):
	//
	//	F:<糧食>      永遠畫
	//	Ship:<耐久>   ★ 只在**揚著帆的大船上而且不在戰鬥中**才畫
	//	              (`byte_3E08C & 0F8h == 20h` 且 `byte_3E0A3 < 80h`);
	//	              否則那個位置畫的是 G:<金幣>
	//	<月>-<日>-<年>
	//	<模式字母>    byte_3E08A 非 0 時畫('P'/'N'/'T'/'Q'/'C')
	//
	// ⚠ **Ship 與 G 共用同一格**:原版 `jnz short loc_2A29F` → `sub_2A0C4`,
	// 兩者是 if/else。所以在船上看不到金幣 —— 那不是遺漏,是版面只有一格。
	y += LineHeight
	second := fmt.Sprintf("G:%d", st.Inventory.Gold)
	if hull, ok := st.ShipHullShown(); ok {
		second = fmt.Sprintf("船:%d", hull)
	}
	s.Text.Draw(dst, PanelX, y, fmt.Sprintf("糧:%-5d %s", st.Inventory.Food, second))
	y += LineHeight
	line := st.Clock.DateString()
	if st.CombatMode != 0 {
		// 模式字母原版直接畫那個位元組('P' 防護 / 'N' 抗魔 / 'T' 停時…)。
		line += "  [" + string(rune(st.CombatMode)) + "]"
	}
	if st.WindShown() {
		line += "  " + st.WindName() + "風"
	}
	s.Text.Draw(dst, PanelX, y, line)

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
	hint := "方向鍵移動   E 進入   K 攀爬   T 交談   F5 音源   F10 離開"
	switch s.State.Prompt {
	case game.PromptLeave:
		hint = "Y 是 / N 否"
	case game.PromptTalk:
		hint = "輸入關鍵字後按 Enter,打 bye 或 ESC 結束"
	case game.PromptAnswer:
		hint = "回答對方的問題後按 Enter"
	case game.PromptSpell:
		hint = "打符文首字母(最多四個),Enter 或空白鍵送出,ESC 作罷"
	case game.PromptShrine:
		hint = "打英文的美德名 / 真言後按 Enter,ESC 作罷"
	case game.PromptYell:
		hint = "喊一個字後按 Enter,ESC 作罷"
	case game.PromptBlackthorn:
		hint = "回答黑棘後按 Enter —— 說出真言就是招了"
	case game.PromptArrest:
		hint = "Y 束手就擒 / N 反抗"
	case game.PromptGuard:
		hint = "Y 付 / N 不付"
	case game.PromptCreate:
		hint = creationHint(s.State)
	case game.PromptMenu:
		hint = "↑↓ 移動,Enter 選定"
	case game.PromptZtats:
		hint = "←→ 翻頁,0 裝備頁,1-6 跳隊員,ESC 收起"
	case game.PromptPick:
		hint = "↑↓ 移動,PgUp/PgDn 跳 7 項,Enter 選定,ESC 作罷"
		if s.State.PickMulti() {
			// 調藥的藥草清單是複選 —— Enter 是勾選,**M 才確定**。
			hint = "↑↓ 移動,空白 / Enter 勾選,M 開始調配,ESC 作罷"
		}
		if s.State.PickIsMember() {
			// 選人選單多兩組操作(原版 `sub_2A7F4`):四顆方向鍵都能動游標、
			// 數字鍵直跳,而 0 與空白鍵是「不選任何人」。
			hint = "方向鍵移動,1-6 直接跳,Enter 選定,0 / 空白 / ESC 不選"
		}
		if s.State.Guard != nil && s.State.Guard.Password {
			hint = "打密語後按 Enter,ESC 作罷"
		}
	case game.PromptText:
		hint = "打字後按 Enter,ESC 保留原值"
	case game.PromptNumber:
		hint = "按數字鍵,0 或 ESC 放棄"
		if in := s.State.NumberInput(); in != "" {
			hint = in + "_ —— Enter 送出,Backspace 退一位"
		}
	case game.PromptEnding:
		hint = "Y 是 / N 否,之後按任意鍵繼續"
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
	case game.PromptSpell:
		label = "咒語:"
	case game.PromptText:
		// 通用文字提問(轉入 U4 的改名);問句本身已經在訊息欄裡。
		label = "汝打:"
	case game.PromptShrine:
		// 美德名、真言、獻金三步共用一行 —— 問句本身已經在訊息欄裡。
		label = "汝言:"
	case game.PromptYell:
		label = "汝喊:"
	case game.PromptBlackthorn:
		label = "汝答:"
	case game.PromptCreate:
		if s.State.Create == nil || s.State.Create.Stage != game.CreationName {
			return y
		}
		s.Text.Draw(dst, PanelX, y, "汝名:"+s.State.Create.Name+"_")
		return y + LineHeight
	case game.PromptGuard:
		if s.State.Guard == nil || !s.State.Guard.Password {
			return y
		}
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
	switch s.State.Prompt {
	case game.PromptTalk, game.PromptAnswer, game.PromptSpell, game.PromptShrine, game.PromptYell, game.PromptBlackthorn, game.PromptText:
		avail-- // 留一行給輸入列
	case game.PromptGuard:
		if s.State.Guard != nil && s.State.Guard.Password {
			avail--
		}
	case game.PromptCreate:
		if s.State.Create != nil && s.State.Create.Stage == game.CreationName {
			avail-- // 打名字那一步要留一行給輸入列
		}
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
	items := s.DungeonItems

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
	seen := u5data.DungeonViewDepths
	for depth := 0; depth < u5data.DungeonViewDepths; depth++ {
		tile := st.DungeonTileAt(u5data.DungeonWrap(x), u5data.DungeonWrap(y))
		if !u5data.DungeonSeeThrough(tile, depth) {
			if n := u5data.DungeonFrontShape(tile, depth); n >= 0 {
				s.blitSlice(dst, set, n, ox, oy, depth)
			}
			seen = depth + 1
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
	// 走廊裡的東西(梯子、寶箱、噴泉、陷阱)—— 由遠而近疊上去,
	// 照 `sub_3D14` 的第二個迴圈。近的蓋遠的。
	if items != nil {
		for depth := seen - 1; depth >= 0; depth-- {
			tile := st.DungeonTileAt(
				u5data.DungeonWrap(d.X+fdx*depth), u5data.DungeonWrap(d.Y+fdy*depth))
			if n := u5data.DungeonObjectUpper(tile, depth); n >= 0 {
				s.blitObject(dst, items, n, ox, oy, depth, true)
			}
			if n := u5data.DungeonObjectLower(tile, depth); n >= 0 {
				s.blitObject(dst, items, n, ox, oy, depth, false)
			}
			// 頭上有洞:另外一張,畫在上半(`byte_4FF9E`)。
			if tile&u5data.DungeonHoleAbove != 0 {
				s.blitObject(dst, items, u5data.DungeonHoleShape(depth), ox, oy, depth, true)
			}
		}
	}
	return true
}

// blitObject 畫走廊裡的一個物件。upper 為真時貼齊天花板,否則貼齊地板。
//
// 每張圖都是**半邊**,水平鏡射補成整個 —— 與牆一樣。
func (s *Scene) blitObject(dst *image.NRGBA, set u5data.PictureSet, n, ox, oy, depth int, upper bool) {
	if n < 0 || n >= len(set) || set[n] == nil {
		return
	}
	p := set[n]
	top := u5data.DungeonFloorY(depth) - p.Height
	if upper {
		top = u5data.DungeonCeilingY(depth)
	}
	for _, mirror := range [2]bool{false, true} {
		for py := 0; py < p.Height; py++ {
			for px := 0; px < p.Width; px++ {
				if p.Mask != nil && p.Mask[py*p.Width+px] != 0 {
					continue
				}
				// 物件靠中線:左半從 80−w 起,右半鏡射。
				vx := u5data.DungeonViewHalfWidth - p.Width + px
				if mirror {
					vx = u5data.DungeonViewWidth - 1 - vx
				}
				c := u5data.EGAPalette[p.Pix[py*p.Width+px]&0x0F]
				for ky := 0; ky < DungeonViewScale; ky++ {
					for kx := 0; kx < DungeonViewScale; kx++ {
						SetPixel(dst,
							ox+vx*DungeonViewScale+kx,
							oy+(top+py)*DungeonViewScale+ky, c)
					}
				}
			}
		}
	}
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

// 開場動畫的版面
//
// ⚠ **這個版面是本專案自己排的,不是原版的。** 原版的座標在 DOS `INTRO.OVL` 裡
// (FM Towns 那份 `byte_54298` / `byte_542B0` 是 640×480 的,套過來會歪)。
// 「哪一頁配哪張圖」是照原版的頁表,擺在哪裡不是 —— 這個分界寫在
// `u5data.IntroPages` 的說明裡。
// 原版在 320×200 下把插圖幾乎鋪滿整個畫面(最大一張 176×192),
// 文字直接**疊在圖上**的一個方框裡 —— 這裡照做:圖 2× 靠上,
// 下方一塊實心黑底放文字。中文 24 px 比原版的 8 px 字大得多,
// 所以文字框拉高一些,蓋掉插圖的下緣。
const (
	introArtTop   = 0
	introTextTop  = 262
	introMargin   = 24
	introTextRule = introTextTop - 10
)

// drawIntro 畫開場的一頁:上面插圖、下面文字。
func (s *Scene) drawIntro(dst *image.NRGBA) {
	fill(dst, EGABlack)
	in := s.State.Intro
	if in.Page >= 0 && in.Page < u5data.IntroPageCount {
		p := u5data.IntroPages[in.Page]
		// 第二張先畫,主圖疊上去 —— 原版也是這個順序(`kind >= 4` 那一段
		// 在主圖之後才畫,所以主圖在下)。
		s.drawIntroArt(dst, p.Story, p.Shape2, true)
		s.drawIntroArt(dst, p.Story, p.Shape, false)
	}
	// 文字底:蓋一塊實心黑,免得字疊在插圖上看不清。
	fillRect(dst, 0, introTextRule, CanvasWidth, CanvasHeight-introTextRule, EGABlack)
	for x := 0; x < CanvasWidth; x++ {
		SetPixel(dst, x, introTextRule, ColorText)
	}
	if s.Text == nil {
		return
	}
	y := introTextTop
	for _, line := range in.VisibleLines() {
		s.Text.Draw(dst, introMargin, y, line)
		y += LineHeight
	}
	hint := "任意鍵繼續    ESC 跳過"
	if in.MoreOnThisPage() {
		hint = "任意鍵看下文  ESC 跳過"
	}
	s.Text.Draw(dst, introMargin, CanvasHeight-LineHeight-4, hint)
}

// drawIntroArt 把一張插圖畫在上半部。second 為真時往右下偏,免得完全蓋住主圖。
func (s *Scene) drawIntroArt(dst *image.NRGBA, story, shape int, second bool) {
	if shape < 0 || story < 0 || story >= len(s.IntroArt) {
		return
	}
	set := s.IntroArt[story]
	if shape >= len(set) || set[shape] == nil {
		return
	}
	p := set[shape]
	x := (CanvasWidth - p.Width*2) / 2
	y := introArtTop
	if second {
		// 原版是 (x, y + 0x37) —— 相對位移照抄,絕對座標沒有。
		y += 0x37
	}
	for py := 0; py < p.Height; py++ {
		for px := 0; px < p.Width; px++ {
			if p.Mask != nil && p.Mask[py*p.Width+px] != 0 {
				continue
			}
			c := u5data.EGAPalette[p.Pix[py*p.Width+px]&0x0F]
			for ky := 0; ky < 2; ky++ {
				for kx := 0; kx < 2; kx++ {
					SetPixel(dst, x+px*2+kx, y+py*2+ky, c)
				}
			}
		}
	}
}

// creationHint 是建角每一步的操作提示。
//
// 建角是玩家**第一個**碰到的畫面,提示錯了就是第一印象壞掉 ——
// 而它有四種輸入方式(任意鍵 / A・B / 打字 / M・F),不逐步標會按不下去。
func creationHint(st *game.State) string {
	if st.Create == nil {
		return ""
	}
	switch st.Create.Stage {
	case game.CreationIntro, game.CreationClosing:
		return "按任意鍵繼續"
	case game.CreationQuestion:
		return "A 或 B —— 選一個"
	case game.CreationName:
		return "打上名字後按 Enter"
	case game.CreationGender:
		return "M 男 / F 女"
	}
	return ""
}

// drawZtats 畫角色數值畫面(原版的 Ztats)。
func (s *Scene) drawZtats(dst *image.NRGBA) { s.drawLines(dst, s.State.ZtatsLines()) }

// drawLines 從固定的起點往下畫幾行 —— 整頁式畫面共用。
func (s *Scene) drawLines(dst *image.NRGBA, lines []string) {
	if s.Text == nil {
		return
	}
	const top = 88
	for i, line := range lines {
		s.Text.Draw(dst, MapOriginX, top+i*LineHeight, line)
	}
}
