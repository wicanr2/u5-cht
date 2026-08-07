package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/wicanr2/u5-cht/internal/i18n"
	"github.com/wicanr2/u5-cht/internal/u5data"
)

func main() {
	dir := os.Getenv("U5_GAMEDATA")
	set, err := u5data.LoadTalkSet(dir)
	if err != nil {
		panic(err)
	}
	dict, _ := u5data.LoadDictionary(dir)
	seen := map[string]int{}
	for _, tf := range set.Files {
		if tf == nil {
			continue
		}
		for i := range tf.Records {
			n := strings.TrimSpace(tf.Records[i].Name(dict))
			if n != "" {
				seen[n]++
			}
		}
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	done := 0
	for _, k := range keys {
		if i18n.Has(k) {
			done++
		}
	}
	fmt.Printf("# %d 個不重複的 NPC 名,已翻 %d\n", len(keys), done)
	for _, k := range keys {
		mark := " "
		if i18n.Has(k) {
			mark = "✓"
		}
		fmt.Printf("%s %s\n", mark, k)
	}
}
