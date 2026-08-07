package i18n

import "testing"

// TestNoDuplicateChineseNames:兩個不同的英文名不該撞同一個中文譯名。
//
// 撞名的後果不是「醜」而是**玩家分不出兩個 NPC**,而畫面上完全看不出問題。
// 少數幾組是原版就同名(例如生物表裡兩筆 Bard),那些列在例外裡。
func TestNoDuplicateChineseNames(t *testing.T) {
	seen := map[string][]string{}
	for en, zh := range names {
		seen[zh] = append(seen[zh], en)
	}
	allowed := map[string]bool{
		"吟遊詩人": true, // 生物表裡 Bard 出現兩次(索引 1 與 7)
	}
	for zh, ens := range seen {
		if len(ens) > 1 && !allowed[zh] {
			t.Errorf("譯名 %q 被 %v 共用 —— 玩家會分不出來", zh, ens)
		}
	}
}

// TestSeriesNamesMatchU6:系列共通名要與 u6-cht 一致。
//
// 這一條是使用者明確要求的政策(2026-08-07):**名稱跟 u6-cht 對齊**。
// 值抄自 `~/u3-cht/u6-cht/dumps/glossary.json` 與該專案的譯文。
// 本專案第一版有五個地名與兩個人名跟 u6 不同,這條把它們釘回去。
func TestSeriesNamesMatchU6(t *testing.T) {
	u6 := map[string]string{
		// glossary.json 的 people / places
		"Lord British":   "不列顛王",
		"Blackthorn":     "黑棘",
		"Avatar":         "聖者",
		"Iolo":           "尤洛",
		"Shamino":        "夏米諾",
		"Dupre":          "杜普雷",
		"Geoffrey":       "傑佛瑞",
		"Gwenno":         "葛雯",
		"Sentri":         "山特利",
		"Mariah":         "瑪萊雅",
		"Nystul":         "尼斯托",
		"Sin'Vraal":      "辛弗拉",
		"Britannia":      "不列顛尼亞",
		"BRITAIN":        "不列顛城",
		"TRINSIC":        "特林希克",
		"NEW MAGINCIA":   "新馬精西亞",
		"MINOC":          "米諾克",
		"COVE":           "海灣鎮",
		"STONEGATE":      "石門",
		"THE LYCAEUM":    "學苑",
		"EMPATH ABBEY":   "共感修道院",
		"SERPENT'S HOLD": "巨蛇堡",
		"Underworld":     "幽冥界",
		// 本專案第一版與 u6 不同、已改過來的五個
		"YEW":        "尤伊",
		"MOONGLOW":   "月華城",
		"JHELOM":     "傑隆",
		"SKARA BRAE": "斯卡拉布雷",
		// u6 譯文裡出現的 NPC
		"Charlotte": "夏綠蒂",
		"Jaana":     "雅娜",
		"Katrina":   "卡翠娜",
		"Sutek":     "蘇特克",
		"Smith":     "史密斯",
		// u6 手冊的怪物表
		"Skeleton": "骷髏怪",
		"Gremlin":  "小妖怪",
		"Mimic":    "化形怪",
		"Reaper":   "樹妖",
		"Gazer":    "多眼妖",
		"Corpser":  "拖屍怪",
		"Mongbat":  "蝙猴",
		"Headless": "無頭怪",
		"Daemon":   "惡魔",
		// u6 手冊的武器表
		"Dagger":    "匕首",
		"Spear":     "長矛",
		"Bow":       "弓",
		"Crossbow":  "十字弓",
		"Magic Bow": "魔法弓",
		"Magic Axe": "魔斧",
		"Ankh":      "安卡",
	}
	for en, want := range u6 {
		if got := Name(en); got != want {
			t.Errorf("%s → %q,u6-cht 是 %q", en, got, want)
		}
	}
}

// TestUnknownFallsBackToEnglish:查不到就原樣回傳,不要回空字串。
func TestUnknownFallsBackToEnglish(t *testing.T) {
	if got := Name("Nonexistent Widget"); got != "Nonexistent Widget" {
		t.Errorf("查不到卻回 %q", got)
	}
	if Has("Nonexistent Widget") {
		t.Error("Has 對沒翻的東西回了 true")
	}
}

// TestBilingualKeepsEnglishVisible:雙語顯示要保留英文。
//
// 咒語與符文玩家得**打得出來**,所以顯示一定要帶原文。
func TestBilingualKeepsEnglishVisible(t *testing.T) {
	if got := Bilingual("Dagger"); got != "匕首(Dagger)" {
		t.Errorf("Bilingual 回 %q", got)
	}
	// 沒翻的只回一份,不要變成「Foo(Foo)」。
	if got := Bilingual("Foo"); got != "Foo" {
		t.Errorf("沒翻的 Bilingual 回 %q", got)
	}
}

// TestEveryNameIsChinese:表裡的值不能還是英文(那代表忘了翻)。
func TestEveryNameIsChinese(t *testing.T) {
	for en, zh := range names {
		if zh == en {
			t.Errorf("%q 的譯名跟原文一樣 —— 忘了翻?", en)
		}
		hasHan := false
		for _, r := range zh {
			if r >= 0x3400 && r <= 0x9FFF {
				hasHan = true
				break
			}
		}
		if !hasHan {
			t.Errorf("%q → %q 裡沒有漢字", en, zh)
		}
	}
}
