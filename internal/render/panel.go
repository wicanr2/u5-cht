package render

import (
	"fmt"
	"image"

	"github.com/wicanr2/u5-cht/internal/i18n"
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 兩種介面模式:原版版面 / 現代版面
//
// 使用者 2026-08-09:「希望遊戲界面能夠有原版跟現代兩種模式,目前的 GUI 排版
// 雖然跟原版一樣但感覺不夠友善簡潔。」
//
// 兩種模式**畫的是同一批資料**,差別只在版面 —— 所以資料組裝集中在
// `panelData()`,兩個 layout 各自排。這樣「現代模式少了一個欄位」
// 這種問題不可能發生:少了就是 layout 沒排,不是資料沒算。
//
//	原版模式  欄位順序與行距照原版的右欄(緊密、一行接一行)
//	現代模式  分組 + 留白 + 分隔線,訊息欄從更高的位置開始(能顯示更多行)
//
// ⚠ **兩種模式都不是「原版逐像素重現」。** 原版右欄是 320×200 下的 8×8 ASCII,
// 這裡是 640×400 下的中文點陣字 —— 字寬就不一樣,不可能逐像素相同。
// 「原版模式」的意思是**欄位組成與順序照原版**,不是像素級複製。

// UIMode 是介面版面。
type UIMode int

const (
	// UIOriginal 是照原版右欄的緊密版面。
	UIOriginal UIMode = iota
	// UIModern 是分組留白的版面(預設)。
	UIModern
)

// UIModeNames 給 `-ui` 旗標與畫面上的提示用。
var UIModeNames = map[UIMode]string{UIOriginal: "原版版面", UIModern: "現代版面"}

// ParseUIMode 讀 `-ui` 的值。
func ParseUIMode(v string) (UIMode, bool) {
	switch v {
	case "", "modern", "現代":
		return UIModern, true
	case "original", "classic", "原版":
		return UIOriginal, true
	}
	return UIModern, false
}

// panelRow 是右欄的一列文字。
type panelRow struct {
	// 三欄式的隊伍列(name / hp / status)三個都非空;其餘只用 text。
	text            string
	name, hp, state string
	party           bool
}

// panelGroup 是現代版面裡的一個分組(組間留白 + 分隔線)。
type panelGroup struct {
	rows []panelRow
}

// panelData 組出右欄要顯示的所有內容,分成四組:
//
//	① 標題      ② 時間 / 業報 / 地點    ③ 糧食 / 金幣 / 日期 / 風向    ④ 隊伍
//
// `debug` 為真時在第 ② 組多一列座標與地形碼 —— 那**不是原版的欄位**,
// 是開發時看的,預設不顯示(使用者:版面要簡潔)。
func (s *Scene) panelData(debug bool) []panelGroup {
	st := s.State
	var where string
	if st.InScene() {
		where = fmt.Sprintf("★ %s 第 %d 層", st.LocationName(), st.Floor+1)
	} else {
		where = "不列顛尼亞"
		if st.Floor < 0 {
			where = "地下世界"
		}
		if loc, ok := u5data.LocationAt(st.X, st.Y); ok {
			where = "★ " + loc.DisplayName()
		}
	}

	// ★ 原版狀態列的四樣東西(`sub_2A1E8` / `sub_2A0C4` / `sub_2A984`):
	//
	//	F:<糧食>      永遠畫
	//	Ship:<耐久>   ★ 只在**揚著帆的大船上而且不在戰鬥中**才畫;否則畫 G:<金幣>
	//	<月>-<日>-<年>
	//	<模式字母>    byte_3E08A 非 0 時畫('P'/'N'/'T'/'Q'/'C')
	//
	// ⚠ **Ship 與 G 共用同一格**:原版是 if/else,所以在船上看不到金幣 ——
	// 那不是遺漏,是版面只有一格。
	second := fmt.Sprintf("G:%d", st.Inventory.Gold)
	if hull, ok := st.ShipHullShown(); ok {
		second = fmt.Sprintf("船:%d", hull)
	}
	date := st.Clock.DateString()
	if st.CombatMode != 0 {
		date += "  [" + string(rune(st.CombatMode)) + "]"
	}
	if st.WindShown() {
		// ⚠ `WindName()` 已經帶「風」字(無風 / 北風 / …),再接一個會變「無風風」。
		date += "  " + st.WindName()
	}

	status := []panelRow{
		{text: fmt.Sprintf("%s  業報 %2d", st.Clock, st.Karma)},
		{text: where},
	}
	if debug {
		status = append(status, panelRow{
			text: fmt.Sprintf("座標 %3d,%3d  地形 %3d", st.X, st.Y, st.TileAt(st.X, st.Y)),
		})
	}

	var party []panelRow
	for _, c := range st.Party() {
		party = append(party, panelRow{
			party: true,
			name:  i18n.Name(c.Name),
			hp:    fmt.Sprintf("%3d/%-3d", c.HP, c.MaxHP),
			state: u5data.StatusName(c.Status),
		})
	}

	groups := []panelGroup{
		{rows: status},
		{rows: []panelRow{
			{text: fmt.Sprintf("糧:%-5d %s", st.Inventory.Food, second)},
			{text: date},
		}},
		{rows: party},
	}
	// ⚠ **現代版面不畫標題。** 「創世紀 V / 命運勇士」在原版右欄佔兩行,
	// 而視窗標題列已經寫著同一句話 —— 那兩行加上組間留白會把訊息欄擠掉三行,
	// 而訊息欄被擠掉的症狀很具體:交談時對方那段話被切一半。
	// 原版模式照原版留著。
	if s.UI == UIOriginal {
		groups = append([]panelGroup{
			{rows: []panelRow{{text: "創世紀 V"}, {text: "命運勇士"}}},
		}, groups...)
	}
	return groups
}

// drawPanelRow 畫一列(三欄式的隊伍列走固定像素欄位)。
//
// ⚠ 隊伍那三欄**不能用 `%-9s` 補空白**:中文一個字 16 px、ASCII 8 px,
// 而 `%-9s` 補的是**位元組** —— `Elwood`(6 B)補三格剛好 72 px,
// `夏米諾`(9 B)一格都不補只有 48 px ⇒ HP 欄會左右跳。
func (s *Scene) drawPanelRow(dst *image.NRGBA, x, y int, r panelRow) {
	if !r.party {
		s.Text.Draw(dst, x, y, r.text)
		return
	}
	s.Text.Draw(dst, x, y, r.name)
	s.Text.Draw(dst, x+partyHPColumn, y, r.hp)
	s.Text.Draw(dst, x+partyStatusColumn, y, r.state)
}

// drawPanelOriginal 是照原版右欄的緊密版面:一行接一行,只有標題後面留半行。
func (s *Scene) drawPanelOriginal(dst *image.NRGBA, debug bool) int {
	y := MapOriginY
	for gi, g := range s.panelData(debug) {
		for _, r := range g.rows {
			s.drawPanelRow(dst, PanelX, y, r)
			y += LineHeight
		}
		if gi == 0 {
			y += LineHeight / 2 // 標題後面留半行,其餘一行接一行(原版就是這樣)
		}
	}
	return y
}

// drawPanelModern 是分組留白的版面:組間空半行 + 一條細分隔線。
//
// 目標是使用者說的「友善簡潔」:同樣的資訊,但眼睛找得到分界 ——
// 時間與地點是一組、家當是一組、隊伍是一組。
func (s *Scene) drawPanelModern(dst *image.NRGBA, debug bool) int {
	y := MapOriginY
	groups := s.panelData(debug)
	for gi, g := range groups {
		if gi > 0 {
			y += LineHeight / 2
			// 分隔線畫在組與組之間(不畫在標題下面 —— 那裡已經有留白)。
			if gi > 1 {
				drawPanelRule(dst, y-LineHeight/4)
			}
		}
		for _, r := range g.rows {
			s.drawPanelRow(dst, PanelX, y, r)
			y += LineHeight
		}
	}
	return y
}

// drawPanelRule 在右欄畫一條橫向細線。
func drawPanelRule(dst *image.NRGBA, y int) {
	for x := PanelX; x < PanelX+PanelWidth; x++ {
		SetPixel(dst, x, y, ColorRule)
	}
}

// PanelMinMessageTop 是訊息欄最高只能從這裡開始(原版右欄的分界)。
//
// ⚠ 實際起點取 `max(這個值, 右欄真正畫到哪 + 半行)` —— 不能寫死。
// 第一版寫死成一個常數,而現代版面的組間留白讓右欄長了三行 ⇒
// 訊息直接**蓋在隊伍最後一行上**。症狀看起來像「字疊在一起」,
// 根因是「版面高度是算出來的,不是常數」。
const PanelMinMessageTop = PanelTextY
