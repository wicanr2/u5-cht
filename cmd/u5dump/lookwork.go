package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/wicanr2/u5-cht/internal/i18n"
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// `lookwork` —— 產出 Look 敘述與招牌的翻譯工作單
//
// 兩份都是 P5 的文字源,但形狀不一樣,所以工作單也分兩段:
//
//	敘述  512 格但只有 216 段不重複的字。**按字合併、列出共用它的所有格號**,
//	      譯者一次決定一句話,而不是把同一句翻十幾遍。
//	招牌  78 塊,每塊是一個 16 欄的框。整塊一起譯 —— 框裡塞得下幾個中文字
//	      要由譯者當場數,逐列對譯沒有意義。
//
// ⚠ 產出的檔含原版文字 → **不入庫**。

func cmdLookWork(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("用法:u5dump lookwork <gamedata> <out.md>")
	}
	dir, out := args[0], args[1]

	lt, err := u5data.LoadLook(dir)
	if err != nil {
		return err
	}
	ss, err := u5data.LoadSigns(dir)
	if err != nil {
		return err
	}

	var b strings.Builder
	b.WriteString("# Look 敘述與招牌 翻譯工作清單\n\n")
	b.WriteString("由 `u5dump lookwork` 產生。**含原版文字,不要入庫。**\n\n")

	total, done := writeLookRows(&b, lt)
	st, sd := writeSignRows(&b, ss)

	fmt.Fprintf(&b, "\n---\n\n敘述 %d 段(已翻 %d)、招牌 %d 塊(已翻 %d)\n",
		total, done, st, sd)
	if err := os.WriteFile(out, []byte(b.String()), 0o644); err != nil {
		return err
	}
	fmt.Printf("✓ %s —— 敘述 %d 段(已翻 %d)、招牌 %d 塊(已翻 %d)\n",
		out, total, done, st, sd)
	return nil
}

// writeLookRows 列出敘述。同一句話掛在好幾格上時合併成一列。
func writeLookRows(b *strings.Builder, lt *u5data.LookTable) (total, done int) {
	// 用「英文原句」分組,但 key 仍是格號 —— 分組只是為了少譯幾次,
	// 譯文表上每一格各自有 key(見 i18n/look.go 說明為什麼不用原文當 key)。
	group := map[string][]int{}
	for i := 0; i < u5data.LookTiles; i++ {
		var en string
		if i < u5data.LookObjectBase {
			en = lt.Terrain(i)
		} else {
			en = lt.Object(i - u5data.LookObjectBase)
		}
		if u5data.LookPlaceholder(en) {
			continue // 佔位符不翻:那些格子由程式特判,表上的字不會被印出來
		}
		group[en] = append(group[en], i)
	}
	keys := make([]string, 0, len(group))
	for k := range group {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return group[keys[i]][0] < group[keys[j]][0] })

	b.WriteString("## Look 敘述(`LOOK2.DAT`)\n\n")
	b.WriteString("`key` 欄可能有多個 —— 同一句話掛在好幾格上,**每個 key 都要放進譯文表**。\n")
	b.WriteString("索引 < 256 是地形,≥ 256 是物件與生物。\n\n")
	b.WriteString("| key | 英文 |\n|---|---|\n")
	for _, en := range keys {
		idx := group[en]
		total++
		var ks []string
		allDone := true
		for _, i := range idx {
			ks = append(ks, fmt.Sprintf("`%s`", i18n.LookKey(i)))
			if !i18n.LookTranslated(i) {
				allDone = false
			}
		}
		mark := ""
		if allDone {
			mark = " ✅"
			done++
		}
		fmt.Fprintf(b, "| %s%s | %s |\n", strings.Join(ks, " "), mark, mdEscape(en))
	}
	return total, done
}

// writeSignRows 列出招牌。整塊一起譯,所以英文欄是多列拼起來的。
func writeSignRows(b *strings.Builder, ss *u5data.SignSet) (total, done int) {
	b.WriteString("\n## 招牌與墓碑(`SIGNS.DAT`)\n\n")
	b.WriteString("每塊一段。`key` 是**第一列**的 key,其餘列把尾巴的數字往上加。\n")
	b.WriteString("框線(`abbc` / `8lll9` / `g`)是 `RUNES.CH` 的美術字模,譯文照留;\n")
	b.WriteString("框裡一列 14 欄,中文一個字佔兩欄 → **一列最多 7 個中文字**。\n\n")
	for _, sg := range ss.All() {
		lines := sg.Lines()
		total++
		mark := ""
		if i18n.SignTranslated(sg.Location, sg.X, sg.Y) {
			mark = " ✅"
			done++
		}
		fmt.Fprintf(b, "### `%s`%s(地點 %d 樓 %d,%d,%d,共 %d 列)\n\n```\n%s\n```\n\n",
			i18n.SignKey(sg.Location, sg.X, sg.Y, 0), mark,
			sg.Location, sg.Floor, sg.X, sg.Y, len(lines),
			strings.Join(lines, "\n"))
	}
	return total, done
}

// mdEscape 讓字串在 markdown 表格裡不會把欄位切開。
func mdEscape(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "|", "\\|"), "\n", " ")
}
