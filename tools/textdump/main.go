// textdump 列出五個明文訊息檔的原文與譯文,並統計覆蓋率。
//
//	tools/dev.sh go run ./tools/textdump                 全部五個檔
//	tools/dev.sh go run ./tools/textdump MISCMSG.DAT     指定檔案
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wicanr2/u5-cht/internal/i18n"
	"github.com/wicanr2/u5-cht/internal/u5data"
)

var allFiles = []string{"STORY.DAT", "QUESTION.DAT", "KARMA.DAT", "MISCMSG.DAT", "ENDMSG.DAT"}

func main() {
	dir := os.Getenv("U5_GAMEDATA")
	files := os.Args[1:]
	if len(files) == 0 {
		files = allFiles
	}
	total, done := 0, 0
	for _, f := range files {
		tf, err := u5data.LoadText(filepath.Join(dir, f))
		if err != nil {
			panic(err)
		}
		fmt.Printf("=== %s  %d 筆\n", f, len(tf.Records))
		for i, r := range tf.Records {
			total++
			mark := " "
			if i18n.TextTranslated(f, i) {
				mark = "✓"
				done++
			}
			en := strings.ReplaceAll(r.Text(), "\n", "\\n")
			zh := strings.ReplaceAll(i18n.Text(f, i, r.Text()), "\n", "\\n")
			fmt.Printf("%s [%d] %s\n     → %s\n", mark, i, clip(en), clip(zh))
		}
	}
	fmt.Printf("\n覆蓋率:%d / %d\n", done, total)
}

func clip(s string) string {
	rs := []rune(s)
	if len(rs) > 110 {
		return string(rs[:110]) + "…"
	}
	return s
}
