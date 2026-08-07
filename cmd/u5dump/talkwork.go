package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/text/encoding/japanese"

	"github.com/wicanr2/u5-cht/internal/i18n"
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// `talkwork` —— 產出 `.TLK` 對話的翻譯工作清單(英日對照)
//
// 對話本文是 P5 的大宗:四個檔數百筆記錄,每筆又分名字、外觀、招呼、職業、
// 道別、若干關鍵字回應與提問區塊。要翻得動,得先有一份**逐段列出、
// 附上譯文 key、而且英日並排**的清單。
//
// 英日對齊靠 `.TLK` 索引表的 **id 欄**,不是記錄順序(CLAUDE.md §2.4)——
// 兩版的檔頭結構相同(`(u16 位移, u16 id)` 交錯),同一個 id 就是同一個 NPC。
//
// 日文只當**語意佐證**,不當譯名來源:日譯多用片假名音譯,直接搬過來會變成
// 二手轉譯。名詞一律回到《軟體世界》手冊 → 聖者之書體系 → 現代直觀的優先序。
//
// ⚠ 產出的檔含原版對話全文 → **不入庫**(與 IDA 產物、烘出來的字庫同一條規則)。

func cmdTalkWork(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("用法:u5dump talkwork <gamedata> <out.md> [U5_J 目錄]")
	}
	dir, out := args[0], args[1]
	jpnDir := ""
	if len(args) > 2 {
		jpnDir = args[2]
	}

	set, err := u5data.LoadTalkSet(dir)
	if err != nil {
		return err
	}
	jpn := loadJapanese(jpnDir)

	var b strings.Builder
	b.WriteString("# `.TLK` 對話翻譯工作清單\n\n")
	b.WriteString("由 `u5dump talkwork` 產生。**含原版對話全文,不要入庫。**\n\n")
	b.WriteString("- `key` 欄直接貼進 `internal/i18n/talk_*.go` 的譯文表。\n")
	b.WriteString("- 譯文要放玩家名字的地方寫 `%A`(對應原版 opcode 0x81)。\n")
	b.WriteString("- 日文只當語意佐證,**不是譯名來源** —— 名詞回到手冊與聖者之書體系。\n\n")

	total, done := 0, 0
	for fi, name := range u5data.TalkFiles {
		tf := set.Files[fi]
		if tf == nil {
			continue
		}
		fmt.Fprintf(&b, "\n## %s(%d 筆)\n", name, len(tf.Records))
		jf := jpn[strings.TrimSuffix(name, ".TLK")+".JPN"]
		for i := range tf.Records {
			r := &tf.Records[i]
			c := u5data.ParseConversation(r, set.Dict)
			// ★ 讓玩家名字的位置在工作單上**看得見**。
			//
			// 不設的話 opcode 0x81 展開成空的,工作單上是
			// `A fine day to thee, .` —— 譯者看不出那裡有個名字要留,
			// 譯文就把它吃掉了(第 05 批真的發生了)。用 `%A` 當標記,
			// 與譯文表的寫法一致,譯者照抄就對。
			c.AvatarName = i18n.AvatarToken
			var ja [][]byte
			if jf != nil {
				if jr, ok := recordByID(jf, r.NPCIndex); ok {
					// ⚠ **要生的段落,不要走 `Strings`。** 那條路會呼叫
					// `Dictionary.Expand`,而 nil 詞典把每個 ≥ 0x80 的位元組
					// 換成 `<XX>` —— Shift-JIS 的高位元組正好全中,
					// 出來是一片 `<4A>0H` 這種東西,而且看起來「有填」。
					ja = jr.Segments()
				}
			}
			n, d := writeRecord(&b, name, c, ja)
			total += n
			done += d
		}
	}
	fmt.Fprintf(&b, "\n---\n\n共 %d 段,已翻 %d 段(%.1f%%)。\n",
		total, done, 100*float64(done)/float64(max1(total)))

	if err := os.WriteFile(out, []byte(b.String()), 0o644); err != nil {
		return err
	}
	fmt.Printf("✓ %s —— %d 段,已翻 %d 段(%.1f%%)\n",
		out, total, done, 100*float64(done)/float64(max1(total)))
	if jpnDir == "" {
		fmt.Println("  ⚠ 沒給 U5_J 目錄 → 沒有日文對照。語意吃不準時會很吃力,建議帶上。")
	}
	return nil
}

// writeRecord 寫一筆記錄的所有可翻欄位,回傳(段數, 已翻段數)。
func writeRecord(b *strings.Builder, file string, c *u5data.Conversation, ja [][]byte) (int, int) {
	fmt.Fprintf(b, "\n### id %d ─ %s\n\n", c.ID, strings.TrimSpace(c.Name))
	fmt.Fprintln(b, "| key | 英文 | 日文 |")
	fmt.Fprintln(b, "|---|---|---|")

	total, done := 0, 0
	row := func(field, en, ja string) {
		// ⚠ 沒有半個字母的段落**不是文字**,是渲染的殘留(多半是 `@`,
		// 對話腳本的終止標記展開後留下的)。放進工作單只會讓每個 agent
		// 「把 @ 翻成 @」,或更糟 —— 填一個空字串,那會讓覆蓋層回傳空的,
		// 畫面上那句話直接消失。第 03 批就這樣交了 12 段空的。
		if !hasLetter(en) {
			return
		}
		total++
		mark := ""
		if i18n.TalkTranslated(file, c.ID, field) {
			done++
			mark = " ✅"
		}
		fmt.Fprintf(b, "| `%s`%s | %s | %s |\n",
			i18n.TalkKey(file, c.ID, field), mark, cell(en), cell(ja))
	}

	// ⚠ 日文那一欄走 `Strings(nil)` 的**原始段落**,不跑英文那套 opcode / 詞典
	// 展開 —— Shift-JIS 的高位元組會被當成詞典 token 嚼爛。段落順序兩版相同
	// (名字 / 外觀 / 招呼 / 職業 / 道別),所以位置對得上。
	row(i18n.TalkFieldDesc, c.Description, at(ja, u5data.TalkFieldDescription))
	row(i18n.TalkFieldGreet, c.Greeting, at(ja, u5data.TalkFieldGreeting))
	row(i18n.TalkFieldJob, c.Job, at(ja, u5data.TalkFieldJob))
	row(i18n.TalkFieldBye, c.Bye, at(ja, u5data.TalkFieldBye))

	for i := range c.Entries {
		en, _ := c.Render(c.Entries[i].Raw)
		// 關鍵字本身留英文 —— 那是玩家要打進去的 canonical 值。
		fmt.Fprintf(b, "| *關鍵字* `%s` | | |\n", c.Entries[i].Keyword)
		// 關鍵字與回應成對(段 2i+5 / 2i+6),所以回應在奇數那一段。
		row(i18n.TalkEntryField(i), en, at(ja, u5data.TalkFixedFields+2*i+1))
	}
	for i := range c.Questions {
		q := &c.Questions[i]
		en, _ := c.Render(q.Text)
		row(i18n.TalkQuestionField(i), en, "")
		en, _ = c.Render(q.No)
		row(i18n.TalkQuestionNoField(i), en, "")
		en, _ = c.Render(q.Yes)
		row(i18n.TalkQuestionYesField(i), en, "")
	}
	return total, done
}

// at 取日文段落並從 Shift-JIS 解成 UTF-8。
//
// 超出範圍回空字串 —— 兩版的段落數不一定一樣,對不上就留白,不要硬湊。
//
// ⚠ **一定要解碼。** 不解碼的話 markdown 裡是一片 mojibake,而它看起來
// 「有東西」—— 表格填滿了、行數也對,只是每一格都是垃圾。第一版就是這樣,
// 直到把 CASTLE.TLK 第一筆貼出來看才發現。
func at(ja [][]byte, i int) string {
	if i < 0 || i >= len(ja) {
		return ""
	}
	return fromShiftJIS(ja[i])
}

// fromShiftJIS 把一段 Shift-JIS 位元組轉成 UTF-8。
//
// 先把 0x00–0x1F 的控制碼換成空白(那些是對話腳本的指令,不是文字),
// 再整段解碼。解不動就退回可讀字元 —— 工作清單少一格比整份跑不出來好。
func fromShiftJIS(raw []byte) string {
	buf := make([]byte, 0, len(raw))
	for _, c := range raw {
		if c < 0x20 {
			buf = append(buf, ' ')
			continue
		}
		buf = append(buf, c)
	}
	out, err := japanese.ShiftJIS.NewDecoder().Bytes(buf)
	if err != nil {
		return strings.Map(func(r rune) rune {
			if r < 0x20 || r > 0x7E {
				return -1
			}
			return r
		}, string(buf))
	}
	return string(out)
}

// hasLetter 回報這段字裡有沒有半個英文字母。
func hasLetter(s string) bool {
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			return true
		}
	}
	return false
}

// cell 把一段文字塞進 markdown 表格的一格。
func cell(s string) string {
	s = strings.ReplaceAll(strings.TrimSpace(s), "\n", " ")
	s = strings.ReplaceAll(s, "|", "\\|")
	return s
}

func max1(n int) int {
	if n < 1 {
		return 1
	}
	return n
}

// loadJapanese 讀 FM Towns 的四份 `.JPN`。目錄沒給或讀不到就回空 map ——
// 沒有日文對照仍然要能產清單,只是會少一欄。
func loadJapanese(dir string) map[string]*u5data.TalkFile {
	out := map[string]*u5data.TalkFile{}
	if dir == "" {
		return out
	}
	names := make([]string, 0, len(u5data.TalkFiles))
	for _, n := range u5data.TalkFiles {
		names = append(names, strings.TrimSuffix(n, ".TLK")+".JPN")
	}
	sort.Strings(names)
	for _, n := range names {
		tf, err := u5data.LoadTalk(filepath.Join(dir, n), u5data.TalkEncodingShiftJIS)
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠ 讀不到 %s:%v\n", n, err)
			continue
		}
		out[n] = tf
	}
	return out
}

// recordByID 依 id 欄找日文那一筆。**不是用記錄順序** —— 兩版的筆數不一定相同。
func recordByID(tf *u5data.TalkFile, id int) (*u5data.TalkRecord, bool) {
	for i := range tf.Records {
		if tf.Records[i].NPCIndex == id {
			return &tf.Records[i], true
		}
	}
	return nil, false
}

// `shopwork` —— 產出 `SHOPPE.DAT` 的翻譯工作清單
//
// 商店對白比 `.TLK` 小得多(194 筆),而且玩家每次買賣都會看到 ——
// 投報率比對話本文高。key 是**檔案位移**,因為原版本來就是拿位移取文字的。
func cmdShopWork(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("用法:u5dump shopwork <gamedata> <out.md>")
	}
	dir, out := args[0], args[1]
	dict, err := u5data.LoadDictionary(dir)
	if err != nil {
		return err
	}
	raw, err := os.ReadFile(filepath.Join(dir, "SHOPPE.DAT"))
	if err != nil {
		return err
	}

	var b strings.Builder
	b.WriteString("# `SHOPPE.DAT` 商店對白翻譯工作清單\n\n")
	b.WriteString("由 `u5dump shopwork` 產生。**含原版全文,不要入庫。**\n\n")
	b.WriteString("佔位符要照抄進譯文:`#`店名 `$`店主 `%`價格 `&`物品 `*`地名 `@`時段 `^`數量。\n\n")
	b.WriteString("| 位移 | 英文 |\n|---|---|\n")

	total, done := 0, 0
	off := 0
	for off < len(raw) {
		end := off
		for end < len(raw) && raw[end] != 0 {
			end++
		}
		if end > off {
			en := dict.ExpandDAT(raw[off:end])
			if strings.TrimSpace(en) != "" {
				total++
				mark := ""
				if i18n.ShopTranslated(off) {
					done++
					mark = " ✅"
				}
				fmt.Fprintf(&b, "| `%d`%s | %s |\n", off, mark, cell(en))
			}
		}
		off = end + 1
	}
	fmt.Fprintf(&b, "\n共 %d 段,已翻 %d 段(%.1f%%)。\n",
		total, done, 100*float64(done)/float64(max1(total)))
	if err := os.WriteFile(out, []byte(b.String()), 0o644); err != nil {
		return err
	}
	fmt.Printf("✓ %s —— %d 段,已翻 %d 段(%.1f%%)\n",
		out, total, done, 100*float64(done)/float64(max1(total)))
	return nil
}
