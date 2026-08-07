package u5data

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// 商店
//
// NPC 的對話號碼落在 0x81–0xFC 就是商人(見 sub_1B52C 的分派)。原版 `sub_1B294`:
//
//	店種 = 對話號碼 - 0x81                       ← 只有 0x81..0x88 這 8 種
//	在 byte_4185C[店種][16] 裡找值等於當前地點編號的那一格 i
//	店名 = off_4145C[店種][i]   店主 = off_4165C[店種][i]
//
// 店名與店主的字串在 DOS 的 `DATA.OVL` 裡就有(0xBFE 起 46 個店名、0xEFA 起 47 個店主,
// 店主表結束處正好接上詞典)。所以本套件只把**結構**(哪一種店開在哪個地點、
// 哪一家沒有店名)寫死,字串一律讀玩家自己的檔 —— 和詞典的作法一致。
//
// ⚠ 全遊戲 47 家店,但只有 46 個店名:類型 2(馬廄)開在地點 30 的那家在
// `DATA.OVL` 裡沒有名字(執行檔的指標是 NULL)。這不是解析漏掉,是原版資料就這樣。
type ShopType int

const (
	ShopArmoury    ShopType = iota // 0x81 武具店
	ShopTavern                     // 0x82 酒館
	ShopStable                     // 0x83 馬廄
	ShopShipwright                 // 0x84 造船廠
	ShopReagents                   // 0x85 藥草與魔法用品
	ShopGuild                      // 0x86 公會(盜賊用品)
	ShopHealer                     // 0x87 治療所
	ShopInn                        // 0x88 旅店

	ShopTypeCount = 8
	// ShopDialogueBase 是第一種商店的對話號碼。
	ShopDialogueBase = 0x81
)

// TypeName 回傳店種的中文名。
func (t ShopType) TypeName() string {
	switch t {
	case ShopArmoury:
		return "武具店"
	case ShopTavern:
		return "酒館"
	case ShopStable:
		return "馬廄"
	case ShopShipwright:
		return "造船廠"
	case ShopReagents:
		return "藥草鋪"
	case ShopGuild:
		return "公會"
	case ShopHealer:
		return "治療所"
	case ShopInn:
		return "旅店"
	}
	return "?"
}

// ShopEntry 是「哪一種店開在哪個地點」。NameIdx 是它在 DATA.OVL 店名表裡的序號,
// -1 代表這家店沒有名字。
type ShopEntry struct {
	Type     ShopType
	Location int
	NameIdx  int
}

// TypeCounts 是每種店的家數,依 shopEntries 數出來。
//
// 這組數字同時是各張價目表的列數 —— `price.go` 的 TestPriceTableRows 拿它當閘門,
// 任何一邊算錯都會對不上。
func TypeCounts() [ShopTypeCount]int {
	var n [ShopTypeCount]int
	for _, e := range shopEntries {
		n[e.Type]++
	}
	return n
}

// shopEntries 出自執行檔的三張平行表(off_4145C / off_4165C / byte_4185C)。
// 順序就是店主表在 DATA.OVL 裡的順序 —— 兩邊已逐筆比對相符。
var shopEntries = [...]ShopEntry{
	{Type: 0, Location: 2, NameIdx: 0},
	{Type: 0, Location: 3, NameIdx: 1},
	{Type: 0, Location: 4, NameIdx: 2},
	{Type: 0, Location: 5, NameIdx: 3},
	{Type: 0, Location: 6, NameIdx: 4},
	{Type: 0, Location: 17, NameIdx: 5},
	{Type: 0, Location: 24, NameIdx: 6},
	{Type: 0, Location: 26, NameIdx: 7},
	{Type: 0, Location: 32, NameIdx: 8},
	{Type: 1, Location: 1, NameIdx: 9},
	{Type: 1, Location: 2, NameIdx: 10},
	{Type: 1, Location: 3, NameIdx: 11},
	{Type: 1, Location: 4, NameIdx: 12},
	{Type: 1, Location: 8, NameIdx: 13},
	{Type: 1, Location: 19, NameIdx: 14},
	{Type: 1, Location: 22, NameIdx: 15},
	{Type: 1, Location: 24, NameIdx: 16},
	{Type: 1, Location: 30, NameIdx: 17},
	{Type: 2, Location: 6, NameIdx: 18},
	{Type: 2, Location: 20, NameIdx: 19},
	{Type: 2, Location: 22, NameIdx: 20},
	{Type: 2, Location: 30, NameIdx: -1},
	{Type: 3, Location: 3, NameIdx: 21},
	{Type: 3, Location: 5, NameIdx: 22},
	{Type: 3, Location: 21, NameIdx: 23},
	{Type: 3, Location: 24, NameIdx: 24},
	{Type: 4, Location: 1, NameIdx: 25},
	{Type: 4, Location: 4, NameIdx: 26},
	{Type: 4, Location: 7, NameIdx: 27},
	{Type: 4, Location: 23, NameIdx: 28},
	{Type: 4, Location: 30, NameIdx: 29},
	{Type: 5, Location: 8, NameIdx: 30},
	{Type: 5, Location: 22, NameIdx: 31},
	{Type: 5, Location: 24, NameIdx: 32},
	{Type: 6, Location: 5, NameIdx: 33},
	{Type: 6, Location: 6, NameIdx: 34},
	{Type: 6, Location: 7, NameIdx: 35},
	{Type: 6, Location: 21, NameIdx: 36},
	{Type: 6, Location: 23, NameIdx: 37},
	{Type: 6, Location: 30, NameIdx: 38},
	{Type: 6, Location: 31, NameIdx: 39},
	{Type: 7, Location: 2, NameIdx: 40},
	{Type: 7, Location: 3, NameIdx: 41},
	{Type: 7, Location: 7, NameIdx: 42},
	{Type: 7, Location: 20, NameIdx: 43},
	{Type: 7, Location: 22, NameIdx: 44},
	{Type: 7, Location: 24, NameIdx: 45},
}

// DATA.OVL 裡兩張名單的位移。
const (
	shopNamesOffset  = 0x0BFE
	shopOwnersOffset = 0x0EFA
	shopNameCount    = 46
	shopOwnerCount   = 47
)

// Shop 是一家店。
type Shop struct {
	Type     ShopType
	Location int
	Name     string // 可能是空的(全遊戲有一家沒名字)
	Owner    string
	// TypeIndex 是「同種店裡的第幾家」,也就是原版的 word_3EF34
	// (`sub_1B294` 在 byte_4185C[店種] 這一列裡找當前地點的位置)。
	// 所有價目表都用它當列索引。
	TypeIndex int
}

// ShopSet 是全遊戲的商店目錄與商店對白。
type ShopSet struct {
	Shops []Shop
	// Prices 是價目表(讀自 DATA.OVL)。
	Prices *PriceTable
	// text 是 SHOPPE.DAT 的原始內容,依位元組位移取用。
	text []byte
	dict *Dictionary
}

// LoadShops 讀入商店目錄(DATA.OVL)與對白(SHOPPE.DAT)。
func LoadShops(dir string, dict *Dictionary) (*ShopSet, error) {
	ovl, err := os.ReadFile(filepath.Join(dir, "DATA.OVL"))
	if err != nil {
		return nil, err
	}
	text, err := os.ReadFile(filepath.Join(dir, "SHOPPE.DAT"))
	if err != nil {
		return nil, err
	}
	names, err := readStringList(ovl, shopNamesOffset, shopNameCount)
	if err != nil {
		return nil, fmt.Errorf("店名表:%w", err)
	}
	owners, err := readStringList(ovl, shopOwnersOffset, shopOwnerCount)
	if err != nil {
		return nil, fmt.Errorf("店主表:%w", err)
	}
	if len(shopEntries) != len(owners) {
		return nil, fmt.Errorf("目錄有 %d 家店,店主表有 %d 筆", len(shopEntries), len(owners))
	}
	prices, err := ParsePrices(ovl)
	if err != nil {
		return nil, fmt.Errorf("價目表:%w", err)
	}
	s := &ShopSet{text: text, dict: dict, Prices: prices}
	var seen [ShopTypeCount]int
	for i, e := range shopEntries {
		sh := Shop{Type: e.Type, Location: e.Location, Owner: owners[i], TypeIndex: seen[e.Type]}
		seen[e.Type]++
		if e.NameIdx >= 0 {
			if e.NameIdx >= len(names) {
				return nil, fmt.Errorf("第 %d 家店的店名序號 %d 超出 %d", i, e.NameIdx, len(names))
			}
			sh.Name = names[e.NameIdx]
		}
		s.Shops = append(s.Shops, sh)
	}
	return s, nil
}

// At 找出某個地點、某一種店。
func (s *ShopSet) At(location int, t ShopType) (*Shop, bool) {
	for i := range s.Shops {
		if s.Shops[i].Location == location && s.Shops[i].Type == t {
			return &s.Shops[i], true
		}
	}
	return nil, false
}

// ForDialogue 依商人的對話號碼與所在地點找店。
func (s *ShopSet) ForDialogue(dialogue byte, location int) (*Shop, bool) {
	t := ShopType(int(dialogue) - ShopDialogueBase)
	if t < 0 || t >= ShopTypeCount {
		return nil, false
	}
	return s.At(location, t)
}

// Greeting 取某家店的第 variant 句問候語(0..3),已完成佔位符代換。
func (s *ShopSet) Greeting(sh *Shop, variant, hour int) string {
	if sh == nil {
		return ""
	}
	if variant < 0 || variant > 3 {
		variant = 0
	}
	return s.Say(s.Prices.Greet[sh.Type][variant], sh, hour, 0)
}

// Say 取 SHOPPE.DAT 位移 off 起的一段文字,展開詞典並代換佔位符。
//
// 原版 `sub_11168` 就做這件事:讀 SHOPPE.DAT 的第 off 個位元組起到 NUL,
// 丟給 `sub_10FEC` 展開。位移 0 或超出檔案代表「這條沒有文字」。
func (s *ShopSet) Say(off int, sh *Shop, hour, number int) string {
	return s.SayWith(off, sh, Placeholders{Hour: hour, Number: number})
}

// SayWith 同 Say,但可以指定全部六個佔位符。
func (s *ShopSet) SayWith(off int, sh *Shop, ph Placeholders) string {
	if s == nil || off <= 0 || off >= len(s.text) {
		return ""
	}
	return s.FormatWith(s.readRecord(off), sh, ph)
}

// readRecord 取 off 起到下一個 NUL 為止的原始位元組。
func (s *ShopSet) readRecord(off int) []byte {
	for i := off; i < len(s.text); i++ {
		if s.text[i] == 0 {
			return s.text[off:i]
		}
	}
	return s.text[off:]
}

// Placeholders 是商店文字裡七個佔位符要填的東西。
//
// 全部出自原版 `sub_10FEC` 的跳表(`jpt_11093`),一個不多一個不少:
//
//	#  0x23  店名          dword_3EF24
//	$  0x24  店主          dword_3EF28
//	%  0x25  價格          word_3EF38
//	&  0x26  物品名        dword_3EF2C(賣出對白用;有別稱就用別稱)
//	*  0x2A  地名          dword_3EF30(酒館八卦用)
//	@  0x40  時段          由當前小時決定,不是參數
//	^  0x5E  數量          word_3EF3A(藥草一次幾份)
//
// 之前把 `^` 與 `*` 原樣留著,是因為還不知道它們填什麼;現在兩者都追到了
// 設定處,所以照原版填。
type Placeholders struct {
	Hour   int
	Number int    // %  價格
	Count  int    // ^  數量;0 時不代換,避免把「沒給值」印成 0
	Item   string // &  物品名
	Place  string // *  地名
}

// Format 展開一段商店文字,只代換店名 / 店主 / 時段 / 價格。
func (s *ShopSet) Format(raw []byte, sh *Shop, hour, number int) string {
	return s.FormatWith(raw, sh, Placeholders{Hour: hour, Number: number})
}

// FormatWith 展開一段商店文字並代換全部佔位符。
func (s *ShopSet) FormatWith(raw []byte, sh *Shop, ph Placeholders) string {
	expanded := s.dict.ExpandDAT(raw)
	pairs := []string{
		"#", sh.Name,
		"$", sh.Owner,
		"@", TimeOfDay(ph.Hour),
		"%", strconv.Itoa(ph.Number),
	}
	if ph.Count > 0 {
		pairs = append(pairs, "^", strconv.Itoa(ph.Count))
	}
	if ph.Item != "" {
		pairs = append(pairs, "&", ph.Item)
	}
	if ph.Place != "" {
		pairs = append(pairs, "*", ph.Place)
	}
	return strings.NewReplacer(pairs...).Replace(expanded)
}

// TimeOfDay 回傳原版用的時段字(sub_10FEC 的 `@`)。
func TimeOfDay(hour int) string {
	switch {
	case hour < 12:
		return "morning"
	case hour < 18:
		return "afternoon"
	}
	return "evening"
}

func readStringList(b []byte, off, n int) ([]string, error) {
	out := make([]string, 0, n)
	p := off
	for i := 0; i < n; i++ {
		e := indexByte(b, p)
		if e < 0 {
			return nil, fmt.Errorf("讀到第 %d 筆就沒有結尾了", i)
		}
		w := string(b[p:e])
		if w == "" || !printableASCII(w) {
			return nil, fmt.Errorf("第 %d 筆不像名字:%q", i, w)
		}
		out = append(out, w)
		p = e + 1
	}
	return out, nil
}
