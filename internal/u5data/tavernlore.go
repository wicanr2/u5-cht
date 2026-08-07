package u5data

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

// 酒館的打聽消息(原版 `sub_21500`)—— U5 的線索系統
//
// 你付錢問一個主題,酒保回你「去某地找某人」。26 個主題正好把整個遊戲的
// 尋寶清單蓋完:
//
//	八德    hone comp valo just sacr hono spir humi
//	地牢    dece desp dest wron cove sham hyth
//	三信物  crow scep amul                        ← 王冠 / 權杖 / 護符(docs/re/57)
//	三碎片  fals hatr cowa
//	其他    astr oppr brit resi unde
//
// # 五張表在 `DATA.OVL` 裡是連著的
//
//	0x4C84  26 × u16  關鍵字指標(`hone` `comp` …,各 4 個字母)
//	0x4CB8  26 × u16  人名指標  (`Malik` `Greyson` …)
//	0x4CEC  26 × u8   地名索引  (指向下面那 13 個)
//	0x4D06  13 × u16  地名指標  (`Moonglow` … `Serpent's Hold`)
//	0x4D20  26 × u16  價格      (50..250 金)
//	0x4D52  結束
//
// ★ **連續無縫本身就是錨**:26 / 26 / 26 / 13 / 26 五段首尾相接,
// 中間一個位元組的空隙都沒有,而五段的長度全部剛好對上各自的筆數。
// 位移只要偏一格,後面四張表會一起爛掉。
//
// # 比對規則(`sub_27C98`)
//
// 關鍵字先被大寫化(`al & 0x7F` 之後 `sub al, 0x20`),然後在玩家的答案裡
// 找**子字串**;由前往後掃 26 個主題,**第一個命中的就是答案**。
//
// ⚠ 原版在命中之後還有一段 `cmp byte_55F37[ebx], 20h`(看命中位置前一個字元
// 是不是空白),但**兩個條件分支都落到同一個 `mov edi, esi`** ——
// 也就是說那個檢查對控制流沒有任何影響,是編譯器留下的殘骸。
// 所以「純子字串比對」就是原版的行為,不是簡化。
//
// # 流程
//
//	1. 印「Of what wouldst thou hear my lore, <隊長>?」→ 讀最多 15 個字
//	2. 空的 → 直接結束
//	3. 比對 26 個關鍵字;都不中 → 「That, I cannot help thee with.」
//	4. 報價(`SHOPPE.DAT` 0x134E,`%` = 價格)+「Fair 'nuff?」→ **只收 Y / N**
//	5. N → 「No」結束;Y → 金幣不夠 → 「Sorry, <隊長>」+ 0x146A
//	6. 夠 → 扣錢,`&` = 人名、`*` = 地名,四句模板**隨機挑一句**印出來,
//	   後面接「says <酒保>.」
const (
	// TavernLoreTopics 是主題數。
	TavernLoreTopics = 26
	// TavernLorePlaces 是地名數。
	TavernLorePlaces = 13
	// TavernLoreAnswerMax 是玩家能打幾個字(原版 `sub_2B770(…, 0x0F)`)。
	TavernLoreAnswerMax = 15
)

// 五張表的位移。
const (
	tavernLoreKeyword  = 0x4C84
	tavernLoreWho      = 0x4CB8
	tavernLoreWhereIdx = 0x4CEC
	tavernLoreWhere    = 0x4D06
	tavernLorePrice    = 0x4D20
)

// `SHOPPE.DAT` 裡用到的位移(原版 `sub_11168(…)` 的參數)。
const (
	// TavernLorePriceLine 是報價那句(含 `%` 價格佔位符)。
	TavernLorePriceLine = 0x134E
	// TavernLoreCannotAfford 是金幣不夠那句(前面接「Sorry, <隊長>」)。
	TavernLoreCannotAfford = 0x146A
)

// TavernLoreSay 是四句「去某地找某人」的模板,原版隨機挑一句
//(`sub_28E14(0, 3)`,四句都可能中)。
//
//	0x13A2  "Seek ye & in *!"
//	0x13AE  "Rumour has it that &, who lives in *, doth possess such knowledge."
//	0x13D9  "It may be that &, of *, may be able to help thee!"
//	0x13F3  "Mayhap & in * wilt see fit to aid thee!"
var TavernLoreSay = [4]int{0x13A2, 0x13AE, 0x13D9, 0x13F3}

// TavernLoreEntry 是一個主題。
type TavernLoreEntry struct {
	// Keyword 是要比對的四個字母(原版是小寫,比對時不分大小寫)。
	//
	// ⚠ canonical 值維持英文 —— 玩家打進去的字要跟它比(CLAUDE.md §5.2)。
	Keyword string
	// Who 是「去找誰」。
	Who string
	// Where 是「在哪裡」(已經套過地名索引)。
	Where string
	// Price 是問這一題要多少金幣。
	Price int
}

// TavernLoreTable 是 26 個主題與 13 個地名。
type TavernLoreTable struct {
	Entries [TavernLoreTopics]TavernLoreEntry
	Places  [TavernLorePlaces]string
}

// ParseTavernLore 從 `DATA.OVL` 的內容取出五張表。
func ParseTavernLore(ovl []byte) (*TavernLoreTable, error) {
	need := tavernLorePrice + TavernLoreTopics*2
	if len(ovl) < need {
		return nil, fmt.Errorf("DATA.OVL 只有 %d B,放不下 0x%X 的酒館情報表",
			len(ovl), tavernLoreKeyword)
	}
	str := func(table, i int) (string, error) {
		off := int(binary.LittleEndian.Uint16(ovl[table+i*2:])) + ItemPointerBias
		if off <= ItemPointerBias || off >= len(ovl) {
			return "", fmt.Errorf("第 %d 筆的指標指到 0x%X,超出檔案", i, off)
		}
		e := indexByte(ovl, off)
		if e < 0 {
			return "", fmt.Errorf("第 %d 筆沒有結尾的 NUL", i)
		}
		s := string(ovl[off:e])
		if !printableASCII(s) {
			return "", fmt.Errorf("第 %d 筆不是可讀文字:%q", i, s)
		}
		return s, nil
	}
	t := &TavernLoreTable{}
	for i := 0; i < TavernLorePlaces; i++ {
		p, err := str(tavernLoreWhere, i)
		if err != nil {
			return nil, fmt.Errorf("地名表:%w", err)
		}
		t.Places[i] = p
	}
	for i := 0; i < TavernLoreTopics; i++ {
		kw, err := str(tavernLoreKeyword, i)
		if err != nil {
			return nil, fmt.Errorf("關鍵字表:%w", err)
		}
		who, err := str(tavernLoreWho, i)
		if err != nil {
			return nil, fmt.Errorf("人名表:%w", err)
		}
		wi := int(ovl[tavernLoreWhereIdx+i])
		if wi >= TavernLorePlaces {
			return nil, fmt.Errorf("第 %d 筆的地名索引是 %d,超出 %d 個地名",
				i, wi, TavernLorePlaces)
		}
		t.Entries[i] = TavernLoreEntry{
			Keyword: kw,
			Who:     who,
			Where:   t.Places[wi],
			Price:   int(binary.LittleEndian.Uint16(ovl[tavernLorePrice+i*2:])),
		}
	}
	if err := t.validate(); err != nil {
		return nil, err
	}
	return t, nil
}

// LoadTavernLore 從 `DATA.OVL` 讀出酒館情報表。
func LoadTavernLore(dir string) (*TavernLoreTable, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "DATA.OVL"))
	if err != nil {
		return nil, err
	}
	return ParseTavernLore(raw)
}

// validate 用「位移偏掉就一定違反」的性質把五張表一起釘住。
func (t *TavernLoreTable) validate() error {
	// 每個關鍵字都是四個小寫字母 —— 26 筆全部成立,表就沒有滑動的餘地。
	for i, e := range t.Entries {
		if len(e.Keyword) != 4 {
			return fmt.Errorf("第 %d 個關鍵字是 %q,預期四個字母", i, e.Keyword)
		}
		for j := 0; j < 4; j++ {
			if e.Keyword[j] < 'a' || e.Keyword[j] > 'z' {
				return fmt.Errorf("第 %d 個關鍵字 %q 含非小寫字母", i, e.Keyword)
			}
		}
	}
	// 兩頭各釘一個,加上八德的頭尾 —— 四個一起中。
	for _, c := range []struct {
		i  int
		kw string
	}{{0, "hone"}, {7, "humi"}, {15, "crow"}, {TavernLoreTopics - 1, "unde"}} {
		if t.Entries[c.i].Keyword != c.kw {
			return fmt.Errorf("第 %d 個關鍵字是 %q,預期 %q", c.i, t.Entries[c.i].Keyword, c.kw)
		}
	}
	if t.Places[0] != "Moonglow" {
		return fmt.Errorf("第 0 個地名是 %q,預期 Moonglow", t.Places[0])
	}
	// 價格全部落在合理範圍,而且不是零 —— 表偏一格會抓到別的資料。
	for i, e := range t.Entries {
		if e.Price < 1 || e.Price > 999 {
			return fmt.Errorf("第 %d 題的價格是 %d", i, e.Price)
		}
	}
	return nil
}

// Match 把玩家的答案對成主題編號;−1 代表沒有一個主題吃得下。
//
// 由前往後掃,**第一個命中的就算**(原版的迴圈是 break,不是取最佳)。
func (t *TavernLoreTable) Match(answer string) int {
	if t == nil || answer == "" {
		return -1
	}
	up := upperASCII(answer)
	for i, e := range t.Entries {
		if contains(up, upperASCII(e.Keyword)) {
			return i
		}
	}
	return -1
}

// upperASCII 把 ASCII 小寫轉大寫(原版 `al & 0x7F` 之後 `sub al, 0x20`)。
func upperASCII(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'a' && c <= 'z' {
			b[i] = c - 0x20
		}
	}
	return string(b)
}

// contains 是最樸素的子字串搜尋 —— 與 `sub_27C98` 的語意一致。
func contains(hay, needle string) bool {
	if needle == "" || len(needle) > len(hay) {
		return false
	}
	for i := 0; i+len(needle) <= len(hay); i++ {
		if hay[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
