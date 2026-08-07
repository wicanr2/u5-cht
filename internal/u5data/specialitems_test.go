package u5data

import (
	"os"
	"testing"
)

// 短名字表要解得出來,而且**前 22 筆與 `docs/re/44` 手抄的那份一致**。
func TestSpecialItemTableMatchesTheHandDecodedList(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	tbl, err := LoadSpecialItems(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"Magic Crpt", "Skull Keys", "Amulet", "Crown", "Sceptre",
		"(0", "(1", "(2", "(3", "(4", "(5", "(6", "(7",
		"Shard/Falsehd", "Shard/Hatred", "Shard/Cowrdce",
		"Spyglass", "HMS Cape Plan", "Sextant", "Pocket Watch",
		"Black Badge", "Wooden Box",
	}
	if len(want) != SpecialItemCount {
		t.Fatalf("對照表 %d 筆,SpecialItemCount 是 %d", len(want), SpecialItemCount)
	}
	for i, w := range want {
		if tbl.Names[i] != w {
			t.Errorf("第 %d 筆是 %q,預期 %q", i, tbl.Names[i], w)
		}
	}
	// 索引 → U 指令 case 編號:第 0 筆是 case 16。
	if got := tbl.NameForUseCode(SpecialItemUseBase); got != "Magic Crpt" {
		t.Errorf("case %d 查到 %q", SpecialItemUseBase, got)
	}
	if got := tbl.NameForUseCode(SpecialItemUseBase + SpecialItemCount - 1); got != "Wooden Box" {
		t.Errorf("最後一個 case 查到 %q", got)
	}
	if got := tbl.NameForUseCode(0); got != "" {
		t.Errorf("case 0 不在表裡,卻查到 %q", got)
	}
}

// ★ 同一批裝備有**兩套名字**:0x1806 是長名、0x1946 後段是短名。
//
// 這條擋的是「兩張表其實是同一張」的誤判 —— 若真是同一張,
// 那 `WORKLIST` 上「特殊物品 158 項」就會被當成另一批道具去找,而它不是。
func TestEquipmentHasBothLongAndShortNames(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA")
	}
	long, err := LoadItemTable(dir)
	if err != nil {
		t.Fatal(err)
	}
	short, err := LoadSpecialItems(dir)
	if err != nil {
		t.Fatal(err)
	}
	diff := 0
	for i := 0; i < ItemCount; i++ {
		if long.Names[i] == "" || short.EquipNames[i] == "" {
			continue
		}
		if long.Names[i] != short.EquipNames[i] {
			diff++
			// 短名不該比長名長 —— 那就不叫短名了。
			if len(short.EquipNames[i]) > len(long.Names[i]) {
				t.Errorf("第 %d 件:短名 %q 比長名 %q 還長",
					i, short.EquipNames[i], long.Names[i])
			}
		}
	}
	if diff == 0 {
		t.Error("兩張表完全相同 —— 那就不是「長名 / 短名」兩套了")
	}
	t.Logf("48 件裝備裡有 %d 件的短名與長名不同", diff)
}

// `(0`..`(7` 是原版留的佔位名,要認得出來(清單得跳過它們)。
func TestPlaceholderNamesAreRecognised(t *testing.T) {
	for _, s := range []string{"(0", "(3", "(7"} {
		if !SpecialItemPlaceholder(s) {
			t.Errorf("%q 應該算佔位名", s)
		}
	}
	for _, s := range []string{"", "(", "(8", "(12", "Amulet", "Crown"} {
		if SpecialItemPlaceholder(s) {
			t.Errorf("%q 不該算佔位名", s)
		}
	}
}
