package u5data

import (
	"strings"
	"testing"
)

// ★ 風名**自己就帶「風」字**(無風 / 北風 / …)⇒ 顯示層不可以再接一個。
//
// 這條測試是並排比對抓到「無風風」之後補的(`docs/playtest-checkpoints.md` A1)。
// 值得寫下來的原因:字串拼接的錯誤**不會讓任何測試變紅**,也不會讓程式壞掉 ——
// 它只會在畫面上多一個字,而那要有人真的看畫面才發現。
// 把「名字已經完整」寫成不變式,下一個要接後綴的人就會先撞到這裡。
func TestWindNamesAlreadyIncludeTheWordWind(t *testing.T) {
	for i, n := range WindNameZH {
		if !strings.HasSuffix(n, "風") {
			t.Errorf("風向 %d 的名字 %q 沒有帶「風」字 —— 顯示層可能會自己補一個", i, n)
		}
	}
}
