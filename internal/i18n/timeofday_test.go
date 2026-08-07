package i18n

import "testing"

// TestTimeOfDayBoundariesMatchTheOriginal:時段的邊界照原版,不照中文作息直覺。
//
// 原版 `sub_10FEC`:`hour < 0x0C` 是 morning、`< 0x12` 是 afternoon、其餘 evening。
//
// ⚠ 中文的「下午」直覺上從 13 時起,但原版是 **12 時**;「晚上」直覺上更晚,
// 原版是 **18 時**。照直覺重畫的話,店家在 12 時會說「早好!」——
// 而那種錯只有拿原版並排才看得出來。
func TestTimeOfDayBoundariesMatchTheOriginal(t *testing.T) {
	cases := []struct {
		hour int
		want string
	}{
		{0, "早"}, {11, "早"},
		{12, "午"}, {17, "午"},
		{18, "晚"}, {23, "晚"},
	}
	for _, c := range cases {
		if got := TimeOfDay(c.hour); got != c.want {
			t.Errorf("%d 時 → %q,預期 %q", c.hour, got, c.want)
		}
	}
}

// TestTimeOfDayIsNotEnglish:代入的必須是中文。
//
// 這條的存在理由:譯好的對白是「@好!我是 $……」,`@` 代入英文的話
// 玩家看到「morning好!」。**譯文沒錯,錯的是代入的字** ——
// 這種缺陷在譯文檢查裡完全看不出來。
func TestTimeOfDayIsNotEnglish(t *testing.T) {
	for _, hour := range []int{0, 12, 18} {
		w := TimeOfDay(hour)
		for _, r := range w {
			if r < 0x80 {
				t.Errorf("%d 時的時段字 %q 含 ASCII —— 沒譯", hour, w)
				break
			}
		}
	}
}

// U 指令的 22 個特殊道具:除了原版留的八個佔位名,其餘都要有中譯。
//
// ⚠ key 是原版的**縮寫**(`Magic Crpt` 不是 `Magic Carpet`)—— 查表用的
// 就是檔案裡那個字串,寫成完整拼法會查不到而默默顯示英文。
func TestSpecialItemNamesAreTranslated(t *testing.T) {
	want := []string{
		"Magic Crpt", "Skull Keys", "Amulet", "Crown", "Sceptre",
		"Shard/Falsehd", "Shard/Hatred", "Shard/Cowrdce",
		"Spyglass", "HMS Cape Plan", "Sextant", "Pocket Watch",
		"Black Badge", "Wooden Box",
	}
	for _, en := range want {
		zh := Name(en)
		if zh == "" || zh == en {
			t.Errorf("%q 沒有中譯", en)
			continue
		}
		for _, r := range zh {
			if r < 128 {
				t.Errorf("%q 的譯名 %q 裡有 ASCII 字元 %q", en, zh, r)
				break
			}
		}
	}
}
