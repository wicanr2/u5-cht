package game

import (
	"regexp"
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// englishRun 是「連續三個以上的英文單字」—— 殘留原文的指紋。
//
// ⚠ 不能用「有沒有 ASCII 字母」當判準:譯文裡本來就會有英文
// (真言 AHM、符文 FALLAX、地名、玩家打得出來的關鍵字),
// 而那些是**刻意保留的**(`CLAUDE.md §5.2`)。所以判的是「像句子的一整串」。
var englishRun = regexp.MustCompile(`[A-Za-z]{2,}[ ,]+[A-Za-z]{2,}[ ,]+[A-Za-z]{3,}`)

// allCapsRun 是連續的全大寫詞 —— 咒語符文(`VAS REL XEN BET`)、
// 力量之言(`FALLAX`)、真言(`AHM`)那一族。
//
// ⚠⚠ **這個豁免是實測逼出來的,不是預先想到的。** 第一版偵測器在
// 1307 個關鍵字回應裡命中 3 條,而三條全是誤判:譯文正確地把
// 「吾將其名為『REL XEN BET』」的符文留成英文。
// ⇒ 判英文之前先把全大寫的串挖掉,而不是放寬「三個詞」那個門檻 ——
// 放寬門檻會讓真的英文句子也漏掉。
var allCapsRun = regexp.MustCompile(`[A-Z]{2,}(?:[ ,]+[A-Z]{2,})*`)

// looksEnglish 回報這一行像不像殘留的英文原文。
func looksEnglish(line string) bool {
	return englishRun.MatchString(allCapsRun.ReplaceAllString(line, ""))
}

// TestConversationsRenderInChinese —— ★★★ 端到端:真的開一段對話,看輸出。
//
// i18n 那一層早就有測試(1712 個 key、查得到、無重複),但那驗的是**表**。
// 玩家看到的是**遊戲路徑**:`beginConversation` → `trTalk` → `Log`。
// 中間任何一環(檔名對不上、`Conv.ID` 取錯、忘了呼叫 `trTalk`)都會讓
// 表全綠而畫面全英文 —— 這條就是補那個縫。
func TestConversationsRenderInChinese(t *testing.T) {
	s := openCmdScene(t)
	if s.Talks == nil {
		t.Skip("沒有對話資料")
	}
	checked, bad := 0, 0
	// ★★ 掃**全部**有對話檔的地點,不只一座城 —— 這也是 P5 的
	// 「逐畫面巡查英文殘留」在對白這一半上的自動化。
	// ⚠ **每一層都要進。** 只掃樓層 0 的話,城堡樓上與地下室的 NPC
	// 一個都碰不到 —— 而那些正是對白最長的一批(不列顛王、黑棘、聖者議會)。
	seen := map[int]bool{}
	for loc := 1; loc <= u5data.LastSceneLocation; loc++ {
		for floor := -1; floor <= 4; floor++ {
			if err := s.SetScene(loc, floor, 15, 15); err != nil {
				continue
			}
			c, b2 := s.sweepConversations(t, seen)
			checked += c
			bad += b2
		}
	}
	// ★★ **135 是全部有對白的 NPC 數**(`.TLK` 四檔的相異 NPC:
	// CASTLE 40 + KEEP 32 + TOWNE 48 + DWELLING 15)。寫成斷言而不是
	// 只記在 log 裡 —— 覆蓋率掉下來的時候要紅燈,而不是靜靜少掃幾個。
	if checked != talkNPCTotal {
		t.Errorf("掃到 %d 段對話,預期全部 %d 段 —— 少掃的那些沒有被驗到",
			checked, talkNPCTotal)
	}
	t.Logf("全地點全樓層:查了 %d 段對話的開場,%d 行疑似英文", checked, bad)
}

// talkNPCTotal 是四個 `.TLK` 裡有對白的 NPC 總數。
//
//	CASTLE 40 + KEEP 32 + TOWNE 48 + DWELLING 15 = 135
//
// 數法:`grep -oE '\`[A-Z]+\.TLK#[0-9]+#' talkwork.md | … | sort -u | wc -l`。
const talkNPCTotal = 135

// sweepConversations 把當前地點每個 NPC 的開場四句都跑一次。
//
// seen 記住「(地點, 對話號碼)」—— 同一個 NPC 在多個樓層的排程裡都出現,
// 不去重的話同一段對白會被查很多次,而數字會看起來比實際覆蓋率高。
func (s *State) sweepConversations(t *testing.T, seen map[int]bool) (checked, bad int) {
	t.Helper()
	for i := 1; i < len(s.npcs); i++ {
		n := &s.npcs[i]
		if !n.Present() || n.Dialogue == 0 {
			continue
		}
		key := s.Location<<8 | int(n.Dialogue)
		if seen[key] {
			continue
		}
		seen[key] = true
		s.Messages = nil
		s.talkingTo = i
		s.beginConversation(n.Dialogue)
		if s.Conv == nil {
			continue
		}
		checked++
		for _, line := range s.Messages {
			if looksEnglish(line) {
				bad++
				if bad <= 5 {
					t.Errorf("對話 %d 的這一行像英文原文:%q", n.Dialogue, line)
				}
			}
		}
		s.EndConversation()
	}
	return checked, bad
}

// TestKeywordAnswersRenderInChinese —— 關鍵字的回應也要中譯。
//
// 開場四句(外貌 / 招呼 / 職業 / 道別)與**關鍵字回應**走的是不同的
// 譯文欄位(`e1`、`e2`…),所以要分開驗 —— 只驗開場會漏掉一整族。
func TestKeywordAnswersRenderInChinese(t *testing.T) {
	s := openCmdScene(t)
	if s.Talks == nil {
		t.Skip("沒有對話資料")
	}
	checked, bad := 0, 0
	// 同開場那條:掃全部地點,而且去重。
	seen := map[int]bool{}
	for loc := 1; loc <= u5data.LastSceneLocation; loc++ {
		if err := s.SetScene(loc, 0, 15, 15); err != nil {
			continue
		}
		c, b2 := s.sweepKeywords(t, seen)
		checked += c
		bad += b2
	}
	if checked == 0 {
		t.Skip("沒有關鍵字可問")
	}
	t.Logf("全地點:查了 %d 個關鍵字回應,%d 行疑似英文", checked, bad)
}

// sweepKeywords 把當前地點每個 NPC 的每個關鍵字都問一次。
func (s *State) sweepKeywords(t *testing.T, seen map[int]bool) (checked, bad int) {
	t.Helper()
	for i := 1; i < len(s.npcs); i++ {
		n := &s.npcs[i]
		if !n.Present() || n.Dialogue == 0 {
			continue
		}
		key := s.Location<<8 | int(n.Dialogue)
		if seen[key] {
			continue
		}
		seen[key] = true
		s.talkingTo = i
		s.beginConversation(n.Dialogue)
		if s.Conv == nil {
			continue
		}
		for _, kw := range s.Conv.Keywords() {
			s.Messages = nil
			// 走玩家真的走的那條路:一個一個字元打進去,再 Submit。
			s.Input = ""
			for _, r := range kw {
				s.TypeRune(r)
			}
			s.Submit()
			checked++
			for _, line := range s.Messages {
				if looksEnglish(line) {
					bad++
					if bad <= 5 {
						t.Errorf("對話 %d 的關鍵字 %q 回了英文:%q", n.Dialogue, kw, line)
					}
				}
			}
		}
		s.EndConversation()
	}
	return checked, bad
}

// TestTheDetectorActuallyDetects —— ★ 反對照:偵測器要抓得到真的英文。
//
// 少了這一條,上面兩條可能只是「regexp 永遠不中」。
func TestTheDetectorActuallyDetects(t *testing.T) {
	for _, en := range []string{
		"Welcome, thou art most kind!",
		"I can see many things beyond thy keen!",
		"an old, strangely familiar gypsy.",
	} {
		if !englishRun.MatchString(en) {
			t.Errorf("偵測器漏掉了英文原文:%q", en)
		}
	}
	// 而**刻意保留的**英文不該被誤判。
	for _, zh := range []string{
		"「汝欲何為?」",
		"我見到一名誠實之人吟誦 AHM!",
		"入口刻著符文 FALLAX。",
		"他住在 Verity Isle 上。",
		// ★ 這三條是實測抓出來的誤判(見 `allCapsRun` 的說明)。
		"「吾將其名為『REL XEN BET』。此乃第六圈之咒!」",
		"「吾當時在研究一個叫『VAS REL XEN BET』的咒語。」",
	} {
		if looksEnglish(zh) {
			t.Errorf("誤判了譯文:%q", zh)
		}
	}
	_ = strings.TrimSpace("")
	_ = u5data.KarmaMax
}
