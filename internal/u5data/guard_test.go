package u5data

import (
	"os"
	"path/filepath"
	"testing"
)

// 完整相等,不是詞首比對。
func TestPasswordNeedsBothSidesToEnd(t *testing.T) {
	if !PasswordMatches("IMPE", "IMPE") || !PasswordMatches("IMPE", "impe") {
		t.Error("完整相等(不分大小寫)應該通過")
	}
	for _, wrong := range []string{"IMP", "IMPERA", "", " IMPE"} {
		if PasswordMatches("IMPE", wrong) {
			t.Errorf("%q 不該通過", wrong)
		}
	}
	// 最高位元先清掉(原版 `and al, 7Fh`)—— 這是 `.TLK` 那套 bit7 編碼的殘留。
	if !PasswordMatches("IMPE", string([]byte{'I' | 0x80, 'M', 'P', 'E'})) {
		t.Error("比對前應該先清最高位元")
	}
}

// 貢金是每個活人 10 gp。
func TestTributeIsTenPerHead(t *testing.T) {
	for members, want := range map[int]int{0: 0, 1: 10, 3: 30, 6: 60} {
		if got := GuardTribute(members); got != want {
			t.Errorf("%d 人要 %d gp,預期 %d", members, got, want)
		}
	}
}

// ★ 原版資料的交叉核對:哪些地點真的有「對話號碼 0xFF」的攔路衛兵。
//
// 這條把**程式碼裡的常數**與**資料檔裡的實際內容**對起來:
//
//	`sub_1B3D0` 單獨挑出地點 5 走「一半家財」那條  →  地點 5 真的有一個 0xFF 衛兵
//	`sub_1B3D0` 單獨挑出地點 18 走「徽章 + 密語」  →  地點 18 有整整八個
//
// 兩邊獨立來源對得上,才算真的讀懂了。而且順帶證明**全部 13 個都是
// 生物編號 0x70(衛兵)** —— 「0xFF = 攔路盤查」不是碰巧對上一兩筆。
func TestOnlyGuardsCarryTheChallengeDialogue(t *testing.T) {
	dir := gamedataDir(t)
	set, err := LoadNPCSet(dir)
	if err != nil {
		t.Fatalf("讀不到 NPC 檔:%v", err)
	}
	byLocation := map[int]int{}
	total := 0
	for loc := 1; loc <= 32; loc++ {
		npcs, err := set.At(loc)
		if err != nil {
			continue
		}
		for i := range npcs {
			if npcs[i].Dialogue != DialogueGuardChallenge {
				continue
			}
			total++
			byLocation[loc]++
			if npcs[i].Creature != CreatureGuard {
				t.Errorf("地點 %d 槽 %d 的對話號碼是 0xFF,生物編號卻是 %02X —— 不是衛兵",
					loc, i, npcs[i].Creature)
			}
		}
	}
	if total != 13 {
		t.Errorf("全遊戲有 %d 個攔路衛兵,預期 13", total)
	}
	// 黑棘的宮殿:八個,對應「徽章 + 密語」那一條。
	if byLocation[BlackthornLocation] != 8 {
		t.Errorf("黑棘宮殿(地點 %d)有 %d 個攔路衛兵,預期 8",
			BlackthornLocation, byLocation[BlackthornLocation])
	}
	// 米諾克:一個,對應「一半家財」那一條。
	if byLocation[GuardHalfGoldLocation] != 1 {
		t.Errorf("米諾克(地點 %d)有 %d 個攔路衛兵,預期 1",
			GuardHalfGoldLocation, byLocation[GuardHalfGoldLocation])
	}
	// 其餘四座城各一個,走人頭貢金那一條。
	for _, loc := range []int{1, 3, 4, 6} {
		if byLocation[loc] != 1 {
			t.Errorf("地點 %d 有 %d 個攔路衛兵,預期 1", loc, byLocation[loc])
		}
	}
}

// 密語字串在 DOS 版與 FM Towns 版一致 —— 沒有被哪一版截短。
//
// FM Towns 的 `aImpe` 是 `'IMPE',0`;這裡驗 DOS 的 `DATA.OVL` 也一樣。
func TestPasswordIsInTheDOSDataToo(t *testing.T) {
	dir := gamedataDir(t)
	raw, err := os.ReadFile(filepath.Join(dir, "DATA.OVL"))
	if err != nil {
		t.Fatalf("讀不到 DATA.OVL:%v", err)
	}
	const off = 0x4AAA
	if off+len(BlackthornPassword)+1 > len(raw) {
		t.Fatalf("DATA.OVL 只有 %d B,位移 0x%X 讀不到", len(raw), off)
	}
	got := string(raw[off : off+len(BlackthornPassword)])
	if got != BlackthornPassword {
		t.Errorf("DATA.OVL 0x%X 是 %q,預期 %q", off, got, BlackthornPassword)
	}
	// 前後都要是 NUL —— 證明它是獨立一筆,不是某個長字串的中間。
	if raw[off-1] != 0 || raw[off+len(BlackthornPassword)] != 0 {
		t.Errorf("0x%X 前後不是 NUL(前 %02X 後 %02X)—— 它可能是別的字串的一段",
			off, raw[off-1], raw[off+len(BlackthornPassword)])
	}
}
