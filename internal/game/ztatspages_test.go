package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

func ztatsState(t *testing.T) *State {
	t.Helper()
	s := lockScene(t)
	if !s.BeginZtats() {
		t.Fatal("打不開數值畫面")
	}
	return s
}

// TestZtatsHasSeventeenPages —— ★ 17 頁,而且順序照原版的單一游標。
//
// 一直按「下一頁」要能走遍全部 17 頁再回到起點:六名 × 2 頁只走到
// **實際人數** × 2,之後直接跳 Equipment(原版 `esi == 人數×2 − 1` 那條)。
func TestZtatsHasSeventeenPages(t *testing.T) {
	s := ztatsState(t)
	n := s.PartySize
	if n > len(s.Roster) {
		n = len(s.Roster)
	}
	seen := []int{s.Zstats.Page}
	for i := 0; i < 40; i++ {
		s.ZtatsPage(1)
		if s.Zstats.Page == 0 {
			break
		}
		seen = append(seen, s.Zstats.Page)
	}
	want := n*ZtatsMemberPages + 5 // 隊員頁 + Equipment + 四個清單頁
	if len(seen) != want {
		t.Errorf("走了 %d 頁才繞回,預期 %d(%d 名 × 2 + 5):%v",
			len(seen), want, n, seen)
	}
	// 最後一名的裝備頁之後必須是 Equipment,不是 +1。
	for i, p := range seen {
		if p == n*ZtatsMemberPages-1 {
			if i+1 >= len(seen) || seen[i+1] != ZtatsEquipmentPage {
				t.Errorf("第 %d 頁之後是 %v,原版該跳 Equipment(%d)",
					p, seen[i+1:], ZtatsEquipmentPage)
			}
			break
		}
	}
}

// TestZtatsPrevFromEquipmentLandsOnTheLastMember —— 反向的那個接縫。
func TestZtatsPrevFromEquipmentLandsOnTheLastMember(t *testing.T) {
	s := ztatsState(t)
	n := s.PartySize
	if n > len(s.Roster) {
		n = len(s.Roster)
	}
	s.Zstats.ZtatsJumpEquipment()
	s.ZtatsPage(-1)
	if want := n*ZtatsMemberPages - 1; s.Zstats.Page != want {
		t.Errorf("從 Equipment 往前到第 %d 頁,預期 %d(最後一名的裝備頁)",
			s.Zstats.Page, want)
	}
}

// TestZtatsNumberKeys —— `0` 跳 Equipment、`1`..`6` 跳隊員。
//
// ⚠ 超出隊伍人數的鍵**什麼都不做**(原版先擋 `鍵碼 − 0x31 >= 人數`)。
func TestZtatsNumberKeys(t *testing.T) {
	s := ztatsState(t)
	n := s.PartySize
	if n > len(s.Roster) {
		n = len(s.Roster)
	}
	s.ZtatsKey('0')
	if s.Zstats.Page != ZtatsEquipmentPage {
		t.Errorf("按 0 到第 %d 頁,預期 Equipment(%d)", s.Zstats.Page, ZtatsEquipmentPage)
	}
	for i := 0; i < 6; i++ {
		s.Zstats.Page = ZtatsEquipmentPage
		s.ZtatsKey(rune('1' + i))
		if i < n {
			if want := i * ZtatsMemberPages; s.Zstats.Page != want {
				t.Errorf("按 %d 到第 %d 頁,預期 %d", i+1, s.Zstats.Page, want)
			}
			continue
		}
		if s.Zstats.Page != ZtatsEquipmentPage {
			t.Errorf("按 %d(隊伍只有 %d 人)竟然翻到第 %d 頁 —— 原版什麼都不做",
				i+1, n, s.Zstats.Page)
		}
	}
}

// TestZtatsStatsPageFormatting —— ★ 原版的補位與置中。
//
// Str/Int/Dex 用 `'0'` 補到兩位,HP/HM/Ex 用空白補到四位。
func TestZtatsStatsPageFormatting(t *testing.T) {
	s := ztatsState(t)
	c := &s.Roster[0]
	c.Strength, c.Intel, c.Dex = 7, 8, 9
	c.HP, c.MaxHP, c.Exp = 12, 34, 56
	lines := s.ztatsStatsPage(c)
	joined := strings.Join(lines, "\n")
	for _, want := range []string{"Str=07", "Int=08", "Dex=09"} {
		if !strings.Contains(joined, want) {
			t.Errorf("少了 %q(三圍要用 '0' 補到兩位):\n%s", want, joined)
		}
	}
	for _, want := range []string{"HP:  12", "HM:  34", "Ex:  56"} {
		if !strings.Contains(joined, want) {
			t.Errorf("少了 %q(HP/HM/Ex 用空白補到四位):\n%s", want, joined)
		}
	}
}

// TestZtatsStatusIsCentredInFifteenColumns —— 狀態名置中。
//
// ⚠ 長度**等於** 15 時不置中(原版 `jge` 排除了)—— 這條邊界要驗。
func TestZtatsStatusIsCentredInFifteenColumns(t *testing.T) {
	cases := []struct{ in string; wantPad int }{
		{"Good", (ZtatsStatusFieldWidth - 4) / 2},
		{strings.Repeat("x", ZtatsStatusFieldWidth), 0},
		{strings.Repeat("x", ZtatsStatusFieldWidth + 3), 0},
		{"", 0},
	}
	for _, c := range cases {
		got := ztatsCenter(c.in)
		pad := len(got) - len(strings.TrimLeft(got, " "))
		if c.in == "" {
			pad = 0
		}
		if pad != c.wantPad {
			t.Errorf("%q 前面補了 %d 個空白,預期 %d", c.in, pad, c.wantPad)
		}
	}
}

// TestZtatsArmsPageSaysNoneWhenNothingWorn —— 一件都沒穿才印那句。
//
// ⚠ 原版判準是「六格**加起來**是 0」,所以只要有一件就不印。
func TestZtatsArmsPageSaysNoneWhenNothingWorn(t *testing.T) {
	s := ztatsState(t)
	c := &s.Roster[0]
	for _, slot := range ztatsEquipSlots {
		c.Raw[slot] = u5data.ItemNone
	}
	if got := strings.Join(s.ztatsArmsPage(c), "|"); !strings.Contains(got, "None") &&
		!strings.Contains(got, "無") {
		t.Errorf("六格全空卻沒印「一件都沒有」:%q", got)
	}
	// 只戴一件 → 那句就不該出現。
	c.Raw[ztatsEquipSlots[len(ztatsEquipSlots)-1]] = 1
	got := s.ztatsArmsPage(c)
	if len(got) < 2 {
		t.Fatalf("戴了一件卻只有 %d 行:%v", len(got), got)
	}
	if strings.Contains(strings.Join(got, "|"), "None ready") {
		t.Errorf("戴了一件還印 (None ready):%v", got)
	}
}

// TestZtatsEquipmentPageShowsGrappleOnlyWhenHeld —— 抓鉤有才顯示、且不帶數量。
func TestZtatsEquipmentPageShowsGrappleOnlyWhenHeld(t *testing.T) {
	s := ztatsState(t)
	s.Inventory.Grapple = 0
	if got := strings.Join(s.ztatsEquipmentPage(), "|"); strings.Contains(got, "Grapple") {
		t.Errorf("沒抓鉤卻列出來了:%q", got)
	}
	s.Inventory.Grapple = 1
	lines := s.ztatsEquipmentPage()
	last := lines[len(lines)-1]
	if !strings.Contains(last, "Grapple") && !strings.Contains(last, "抓鉤") {
		t.Errorf("有抓鉤卻沒列出:%q", last)
	}
	// ★ 原版那一行**沒有數字** —— 只印名字。
	if strings.ContainsAny(last, "0123456789") {
		t.Errorf("抓鉤那一行帶了數字 %q,原版只印名字", last)
	}
}

// TestZtatsItemsPageGathersFromEverywhere —— 38 筆的搬運。
//
// ⚠⚠ 旗標型的東西原版填 **0xFF** ⇒ 畫面上是 255。看起來像壞掉,是原版行為。
func TestZtatsItemsPageGathersFromEverywhere(t *testing.T) {
	s := ztatsState(t)
	if n := len(u5data.ZtatsItemNames); n != u5data.ZtatsItemCount {
		t.Fatalf("名稱表 %d 筆,預期 %d", n, u5data.ZtatsItemCount)
	}
	s.HasBadge = true
	s.Regalia.Crown = true
	s.Inventory.Carpets = 2
	got := s.ztatsItemCounts()
	if got[u5data.ZtatsBadgeSlot] != u5data.ZtatsFlagShown {
		t.Errorf("黑徽章顯示 %d,原版旗標填 %d",
			got[u5data.ZtatsBadgeSlot], u5data.ZtatsFlagShown)
	}
	if got[u5data.ZtatsCrownSlot] != u5data.ZtatsFlagShown {
		t.Errorf("王冠顯示 %d,預期 %d", got[u5data.ZtatsCrownSlot], u5data.ZtatsFlagShown)
	}
	// ★ 魔毯是**真的數量**,不是旗標。
	if got[u5data.ZtatsCarpetSlot] != 2 {
		t.Errorf("魔毯顯示 %d,預期 2(它是數量不是旗標)", got[u5data.ZtatsCarpetSlot])
	}
	// 反對照:沒有的東西是 0。
	s.HasBadge = false
	if s.ztatsItemCounts()[u5data.ZtatsBadgeSlot] != 0 {
		t.Error("拿掉黑徽章之後還是非 0")
	}
}

// TestZtatsListSkipsZeroCounts —— 數量 0 的不列。
func TestZtatsListSkipsZeroCounts(t *testing.T) {
	l := &ztatsListPage{"T",
		[]string{"a", "b", "c"}, []int{0, 5, 0}}
	got := ztatsListLines(l)
	if len(got) != 1 {
		t.Fatalf("列了 %d 行,預期只列非 0 的那一筆:%v", len(got), got)
	}
	if !strings.Contains(got[0], "b") || !strings.Contains(got[0], "5") {
		t.Errorf("那一行是 %q,預期含 b 與 5", got[0])
	}
	// 全 0 → 一頁空白(原版沒有「(無)」提示)。
	if n := len(ztatsListLines(&ztatsListPage{"T", []string{"a"}, []int{0}})); n != 0 {
		t.Errorf("全 0 卻列了 %d 行 —— 原版就是一頁空白", n)
	}
}

// TestEveryZtatsPageRendersSomething —— 17 頁逐一不 panic、不整片空白。
//
// ⚠ 清單頁在「持有量全 0」時**本來就空**,所以只驗標題;隊員頁與 Equipment
// 一定要有內容。
func TestEveryZtatsPageRendersSomething(t *testing.T) {
	s := ztatsState(t)
	for p := 0; p <= ZtatsLastPage; p++ {
		s.Zstats.Page = p
		title := s.ZtatsPageTitle()
		body := s.ZtatsBody()
		if p >= ZtatsReagentsPage {
			if title == "" {
				t.Errorf("第 %d 頁沒有標題", p)
			}
			continue
		}
		n := s.PartySize
		if n > len(s.Roster) {
			n = len(s.Roster)
		}
		if p >= n*ZtatsMemberPages && p < ZtatsEquipmentPage {
			continue // 沒有人的頁,原版翻不到
		}
		if title == "" || len(body) == 0 {
			t.Errorf("第 %d 頁標題 %q 內容 %d 行", p, title, len(body))
		}
	}
}
