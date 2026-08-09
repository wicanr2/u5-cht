package render

import (
	"image"

	"github.com/wicanr2/u5-cht/internal/appui"
)

// F1 指令說明的畫面(整頁,蓋掉地圖窗)
//
// ⚠ 這不是原版的畫面 —— 原版靠紙本手冊。內容與理由見 `internal/appui/help.go`。
//
// 版面刻意做成**兩欄**:鍵在左、說明在右,而補充另起一行縮排。
// 一頁裝不完就翻頁 —— 原版的指令有二十六個,硬塞成一頁會小到看不清。

const (
	helpTop     = MapOriginY + LineHeight
	helpKeyCol  = MapOriginX
	helpTextCol = MapOriginX + 92
	// helpNoteIndent 是補充說明的縮排(接在標題那一行下面)。
	helpNoteIndent = helpTextCol + 16
)

// helpLineBudget 是一頁放得下幾**行**(不是幾列)。
//
// ⚠ 一「列」不是固定高度:有補充說明的佔兩行、沒有的佔一行、段落標題還多半行。
// 用固定的「一頁幾列」會讓有些頁排到畫布外、有些頁空一大半 ——
// 所以改成按行數裝箱,`helpPages()` 與 `HelpPageCount()` 共用同一個計算。
const helpLineBudget = 20

// helpPages 把兩段內容切成頁。每一頁是「(段落標題, 幾列)」的序列。
type helpRow struct {
	heading string // 非空 = 這一列其實是段落標題
	entry   appui.HelpEntry
}

func helpRows() []helpRow {
	var rows []helpRow
	for _, sec := range appui.HelpSections() {
		rows = append(rows, helpRow{heading: sec.Heading})
		for _, e := range sec.Entries {
			rows = append(rows, helpRow{entry: e})
		}
	}
	return rows
}

// helpRowLines 是這一列要佔幾行(段落標題的半行間距算 1)。
func helpRowLines(r helpRow) int {
	if r.heading != "" {
		return 2
	}
	if r.entry.Note != "" {
		return 2
	}
	return 1
}

// helpPages 把所有列依行數裝箱成頁。回傳每一頁的列。
func helpPages() [][]helpRow {
	var pages [][]helpRow
	var cur []helpRow
	used := 0
	for _, r := range helpRows() {
		n := helpRowLines(r)
		if used+n > helpLineBudget && len(cur) > 0 {
			pages = append(pages, cur)
			cur, used = nil, 0
		}
		cur = append(cur, r)
		used += n
	}
	if len(cur) > 0 {
		pages = append(pages, cur)
	}
	if len(pages) == 0 {
		pages = [][]helpRow{nil}
	}
	return pages
}

// HelpPageCount 是說明畫面總共幾頁。
func HelpPageCount() int { return len(helpPages()) }

// drawHelp 畫出說明畫面的第 page 頁。
func (s *Scene) drawHelp(dst *image.NRGBA, page int) {
	if s.Text == nil {
		return
	}
	fill(dst, ColorBackground)
	pages := helpPages()
	total := len(pages)
	if page < 0 {
		page = 0
	}
	if page >= total {
		page = total - 1
	}
	title := appui.HelpTitle
	if total > 1 {
		title += "　（" + itoa(page+1) + " / " + itoa(total) + "）"
	}
	s.Text.Draw(dst, MapOriginX, MapOriginY, title)

	y := helpTop + LineHeight/2
	for _, r := range pages[page] {
		if r.heading != "" {
			y += LineHeight / 2
			s.Text.Draw(dst, helpKeyCol, y, "── "+r.heading+" ──")
			y += LineHeight
			continue
		}
		s.Text.Draw(dst, helpKeyCol, y, r.entry.Key)
		s.Text.Draw(dst, helpTextCol, y, r.entry.Title)
		y += LineHeight
		if r.entry.Note != "" {
			s.Text.Draw(dst, helpNoteIndent, y, r.entry.Note)
			y += LineHeight
		}
	}
}

// itoa 是給頁碼用的極小整數轉字串 —— 不想為了兩個數字把 `fmt` 拉進這一檔。
func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	var b [8]byte
	i := len(b)
	for v > 0 {
		i--
		b[i] = byte('0' + v%10)
		v /= 10
	}
	return string(b[i:])
}
