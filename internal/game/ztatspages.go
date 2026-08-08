package game

import (
	"fmt"
	"strings"

	"github.com/wicanr2/u5-cht/internal/i18n"
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// Ztats 的頁面模型(原版 `sub_1E9A0`,推導見 `docs/re/94`)
//
// 原版用**一個游標 `esi`(0..16)**表示 17 頁,不是「哪個人 + 哪一頁」兩個變數:
//
//	0..11  六名隊員 × 兩頁(偶數 = 數值頁 `sub_1DECC`、奇數 = 裝備頁 `sub_1E128`)
//	12     Equipment  `sub_1E210`   全隊共用的糧食 / 金幣 / 鑰匙 / 寶石 / 火把 / 抓鉤
//	13     Reagents   `sub_1E588(…, 8,    byte_3E060)`
//	14     Spells     `sub_1E588(…, 0x30, byte_3E000)`
//	15     Items      `sub_1E588(…, 0x26, byte_40BA0)`
//	16     Armaments  `sub_1E588(…, 0x30, byte_3DFD0)`
//
// ★ 「一個游標」是關鍵:翻頁時「最後一名的裝備頁 → Equipment」與
// 「Armaments → 繞回第一名」這兩個接縫,用兩個變數會寫不出來。

// Ztats 的頁號。
const (
	// ZtatsMemberPages 是每名隊員佔幾頁(數值 + 裝備)。
	ZtatsMemberPages = 2
	// ZtatsEquipmentPage 起是全隊共用的五頁。
	ZtatsEquipmentPage = 0x0C
	ZtatsReagentsPage  = 0x0D
	ZtatsSpellsPage    = 0x0E
	ZtatsItemsPage     = 0x0F
	ZtatsArmamentsPage = 0x10
	// ZtatsLastPage 是游標上限(原版 `cmp esi, 10h; jge`)。
	ZtatsLastPage = ZtatsArmamentsPage
)

// ZtatsNext 往後翻一頁(原版按鍵 `2` / `4`)。
//
//	在最後一名的裝備頁 → 跳到 Equipment(不是 +1)
//	在 Armaments       → 繞回第 0 頁
//	其餘               → +1
//
// ⚠ 第一條判斷用的是 `隊伍人數 × 2 − 1`,所以四人隊伍在第 7 頁就跳走 ——
// 沒有人的第 8..11 頁**看不到**。
func (z *Ztats) ZtatsNext(partySize int) {
	last := partySize*ZtatsMemberPages - 1
	switch {
	case z.Page == last:
		z.Page = ZtatsEquipmentPage
	case z.Page >= ZtatsLastPage:
		z.Page = 0
	default:
		z.Page++
	}
}

// ZtatsPrev 往前翻一頁(原版按鍵 `1` / `3`)。
//
//	在 Equipment → 隊伍人數 × 2 − 1(最後一名的裝備頁)
//	在第 0 頁     → 繞到 Armaments
//	其餘         → −1
func (z *Ztats) ZtatsPrev(partySize int) {
	switch {
	case z.Page == ZtatsEquipmentPage:
		z.Page = partySize*ZtatsMemberPages - 1
	case z.Page <= 0:
		z.Page = ZtatsLastPage
	default:
		z.Page--
	}
}

// ZtatsJumpMember 跳到第 n 名(0-based;原版按鍵 `1`..`6`)。
//
// 原版 `esi = 鍵碼*2 − 0x62`,並先擋 `鍵碼 − 0x31 >= 隊伍人數`
// ⇒ **超出隊伍人數的按鍵什麼都不做**(不是繞回、不是報錯)。
func (z *Ztats) ZtatsJumpMember(n, partySize int) {
	if n < 0 || n >= partySize {
		return
	}
	z.Page = n * ZtatsMemberPages
}

// ZtatsJumpEquipment 是原版按鍵 `0`。
func (z *Ztats) ZtatsJumpEquipment() { z.Page = ZtatsEquipmentPage }

// ztatsListPage 是四個「名稱 + 數量」清單頁的共同形狀(原版 `sub_1E588`)。
type ztatsListPage struct {
	Title  string
	Names  []string
	Counts []int
}

// ZtatsPageTitle 回傳這一頁的標題(隊員頁是角色名)。
func (s *State) ZtatsPageTitle() string {
	p := s.Zstats.Page
	if p < ZtatsEquipmentPage {
		if m := p / ZtatsMemberPages; m < len(s.Roster) {
			return s.Roster[m].Name
		}
		return ""
	}
	switch p {
	case ZtatsEquipmentPage:
		return i18n.Name("Equipment")
	case ZtatsReagentsPage:
		return i18n.Name("Reagents")
	case ZtatsSpellsPage:
		return i18n.Name("Spells")
	case ZtatsItemsPage:
		return i18n.Name("Items")
	case ZtatsArmamentsPage:
		return i18n.Name("Armaments")
	}
	return ""
}

// ZtatsBody 回傳這一頁的內容行。
func (s *State) ZtatsBody() []string {
	if s.Zstats == nil {
		return nil
	}
	p := s.Zstats.Page
	if p < ZtatsEquipmentPage {
		m := p / ZtatsMemberPages
		if m >= len(s.Roster) {
			return nil
		}
		if p%ZtatsMemberPages == 0 {
			return s.ztatsStatsPage(&s.Roster[m])
		}
		return s.ztatsArmsPage(&s.Roster[m])
	}
	if p == ZtatsEquipmentPage {
		return s.ztatsEquipmentPage()
	}
	if l := s.ztatsList(p); l != nil {
		return ztatsListLines(l)
	}
	return nil
}

// ZtatsStatusFieldWidth 是狀態名置中用的欄寬(原版 `mov eax, 0Fh`)。
//
// 原版:`if (0 < len(狀態名) < 15) 先印 (15 − len)/2 個空白`。
// ⚠ 長度 **等於** 15 時不置中也不截斷 —— `jge` 把它排除了。
const ZtatsStatusFieldWidth = 15

// ztatsCenter 重現那段置中(不是一般的「填滿到欄寬」——只補前導空白)。
func ztatsCenter(s string) string {
	n := len([]rune(s))
	if n <= 0 || n >= ZtatsStatusFieldWidth {
		return s
	}
	return strings.Repeat(" ", (ZtatsStatusFieldWidth-n)/2) + s
}

// ztatsStatsPage 是隊員的數值頁(原版 `sub_1DECC`)。
//
// ★★ 三個容易漏的排版細節:
//
//  1. **Str / Int / Dex 用 `'0'` 補位**(原版 `push 30h`),
//     而 HP / HM / Ex / Lv / Magic 用空白(`push 20h`)。
//  2. 狀態名在 15 欄裡**置中**。
//  3. 第一行開頭直接把**性別位元組當字元印**(0x0B / 0x0C)——
//     在 FM Towns 字型裡那兩格是♂/♀。引擎改印中文,見下。
//
// ⚠ 顯示順序是 Str / Int / Dex,而紀錄位移是 0x0C / 0x0E / 0x0D ——
// **Dex 與 Int 在紀錄裡是反的**,照著顯示順序抄會抄錯。
func (s *State) ztatsStatsPage(c *u5data.Character) []string {
	return []string{
		fmt.Sprintf("%s Lv-%d %s", genderMark(c.Gender), c.Level, c.ClassName()),
		ztatsCenter(u5data.StatusName(c.Status)),
		fmt.Sprintf("Str=%02d  HP:%4d", c.Strength, c.HP),
		fmt.Sprintf("Int=%02d  HM:%4d", c.Intel, c.MaxHP),
		fmt.Sprintf("Dex=%02d  Ex:%4d", c.Dex, c.Exp),
		fmt.Sprintf("    %s:%2d", i18n.Name("Magic"), c.MP),
	}
}

// genderMark 是原版第一行開頭那個字元。
//
// 原版直接 `sub_27230(rec[9])` —— 把性別位元組(0x0B 男 / 0x0C 女)
// 當字元丟給印字常式,靠 FM Towns 字型的那兩格是♂/♀。
// 引擎沒有那套字型,改用 Unicode 的同義符號:語意相同、來源照舊。
func genderMark(g byte) string {
	if g == u5data.GenderMale {
		return "♂"
	}
	return "♀"
}

// ztatsArmsPage 是隊員的裝備頁(原版 `sub_1E128`)。
//
// 掃紀錄 0x19..0x1E 六個部位,一件都沒有就印 "(None ready)"。
// ⚠ 判準是「六格**加起來**是 0」,不是逐格判斷 —— 所以只要有一件就不印那句。
func (s *State) ztatsArmsPage(c *u5data.Character) []string {
	out := []string{i18n.Name("Arms")}
	worn := 0
	for _, slot := range ztatsEquipSlots {
		code := c.Raw[slot]
		if code == u5data.ItemNone {
			continue
		}
		worn++
		out = append(out, "  "+i18n.Name(s.equipName(int(code))))
	}
	if worn == 0 {
		out = append(out, "("+i18n.Name("None ready")+")")
	}
	return out
}

// ztatsEquipmentPage 是全隊共用的那一頁(原版 `sub_1E210`)。
//
// ★ 欄寬照原版:糧食與金幣 **4**,鑰匙 / 寶石 / 火把 **2**。
// ★★ **抓鉤只在有的時候出現,而且不顯示數量** —— 原版
// `cmp byte_3DFBB, 0; jz` 之後只印 "\n Grapple",沒有 `sub_23A24`。
func (s *State) ztatsEquipmentPage() []string {
	inv := &s.Inventory
	out := []string{
		fmt.Sprintf(" %s: %4d", i18n.Name("Food"), inv.Food),
		fmt.Sprintf(" %s: %4d", i18n.Name("Gold"), inv.Gold),
		"",
		fmt.Sprintf(" %s.......%2d", i18n.Name("Keys"), inv.Keys),
		fmt.Sprintf(" %s.......%2d", i18n.Name("Gems"), inv.Gems),
		fmt.Sprintf(" %s....%2d", i18n.Name("Torches"), inv.Torches),
	}
	if inv.Grapple != 0 {
		out = append(out, " "+i18n.Name("Grapple"))
	}
	return out
}

// ztatsList 組出四個清單頁之一。
func (s *State) ztatsList(page int) *ztatsListPage {
	inv := &s.Inventory
	switch page {
	case ZtatsReagentsPage:
		return &ztatsListPage{i18n.Name("Reagents"),
			u5data.ReagentNames[:], inv.Reagents[:]}
	case ZtatsSpellsPage:
		if s.Spells == nil {
			return nil
		}
		names := make([]string, u5data.SpellCount)
		for i := range names {
			names[i] = s.Spells.Spells[i].Name
		}
		return &ztatsListPage{i18n.Name("Spells"), names, inv.Spells[:]}
	case ZtatsItemsPage:
		counts := s.ztatsItemCounts()
		return &ztatsListPage{i18n.Name("Items"),
			u5data.ZtatsItemNames[:], counts[:]}
	case ZtatsArmamentsPage:
		if s.Items == nil {
			return nil
		}
		names := make([]string, u5data.ItemCount)
		for i := range names {
			names[i] = s.Items.Name(byte(i))
		}
		return &ztatsListPage{i18n.Name("Armaments"), names, inv.Items[:]}
	}
	return nil
}

// ztatsListLines 把清單頁攤成行。
//
// ★ **只列數量非 0 的** —— 原版 `sub_1E588` 逐筆檢查計數陣列,是 0 就跳過。
// 全部是 0 時原版就是一頁空白(沒有「(無)」那種提示),照抄。
func ztatsListLines(l *ztatsListPage) []string {
	var out []string
	for i, n := range l.Counts {
		if n == 0 || i >= len(l.Names) {
			continue
		}
		out = append(out, fmt.Sprintf("%-14s%3d", i18n.Name(l.Names[i]), n))
	}
	return out
}

// ztatsEquipSlots 是角色紀錄裡的六個裝備部位(原版 `sub_1E128` 掃 0x19..0x1E)。
var ztatsEquipSlots = [...]int{
	u5data.CharHelm, u5data.CharArmour, u5data.CharWeapon,
	u5data.CharShield, u5data.CharRing, u5data.CharAmulet,
}

// ztatsItemCounts 把散落各處的持有量搬成 Items 頁那 38 筆(原版 `sub_1E8D4`)。
//
// 逐筆來源見 `u5data/ztatsitems.go` 的常數群。⚠ 旗標型的東西原版填 **0xFF**
// ⇒ 畫面上那一行的數字是 255。看起來像壞掉,是原版行為。
func (s *State) ztatsItemCounts() [u5data.ZtatsItemCount]int {
	var out [u5data.ZtatsItemCount]int
	inv := &s.Inventory
	for i := 0; i < u5data.ScrollCount && u5data.ZtatsScrollBase+i < u5data.ZtatsPotionBase; i++ {
		out[u5data.ZtatsScrollBase+i] = inv.Scrolls[i]
	}
	for i := 0; i < u5data.PotionCount && u5data.ZtatsPotionBase+i < u5data.ZtatsCarpetSlot; i++ {
		out[u5data.ZtatsPotionBase+i] = inv.Potions[i]
	}
	out[u5data.ZtatsCarpetSlot] = inv.Carpets
	out[u5data.ZtatsOddKeySlot] = inv.OddKeys
	out[u5data.ZtatsAmuletSlot] = ztatsFlag(s.Regalia.Amulet)
	out[u5data.ZtatsCrownSlot] = ztatsFlag(s.Regalia.Crown)
	out[u5data.ZtatsSceptreSlot] = ztatsFlag(s.Regalia.Sceptre)
	// ★ 月石:**還在身上**(地點 0xFF)才列出來,埋下去的就消失。
	for i := range s.Moongates {
		if u5data.ZtatsMoonstoneBase+i >= u5data.ZtatsShardBase {
			break
		}
		if inv.Moonstones[i].InHand() {
			out[u5data.ZtatsMoonstoneBase+i] = u5data.ZtatsFlagShown
		}
	}
	for i := range s.Shards {
		if u5data.ZtatsShardBase+i >= u5data.ZtatsSpyglassSlot {
			break
		}
		out[u5data.ZtatsShardBase+i] = ztatsFlag(s.Shards[i])
	}
	out[u5data.ZtatsSpyglassSlot] = ztatsFlag(s.HasSpyglass)
	out[u5data.ZtatsPlansSlot] = ztatsFlag(s.Regalia.Plans)
	out[u5data.ZtatsSextantSlot] = ztatsFlag(s.HasSextant)
	out[u5data.ZtatsWatchSlot] = ztatsFlag(s.HasWatch)
	out[u5data.ZtatsBadgeSlot] = ztatsFlag(s.HasBadge)
	out[u5data.ZtatsBoxSlot] = ztatsFlag(s.SandalwoodBox)
	return out
}

// ztatsFlag 把布林旗標換成原版顯示的數量(0 或 255)。
func ztatsFlag(b bool) int {
	if b {
		return u5data.ZtatsFlagShown
	}
	return 0
}
