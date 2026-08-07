package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// codexPages 讀完一次寶典,回傳它唸出來的每一頁。
func codexPages(t *testing.T, s *State) []string {
	t.Helper()
	if !s.ReadCodex() {
		t.Fatalf("讀不了寶典:\n%s", s.log())
	}
	pages := append([]string(nil), s.Codex.Pages...)
	for s.AdvanceCodex() {
	}
	return pages
}

// TestCodexAnswersAreTheEightVirtuePassages:八段箴言的記錄序號要對得上。
//
// ★ 依據見 `docs/re/26` §? 與 `u5data.CodexAnswerRecord` 的說明:
// `dword_5604C` 的八個位移加上 `MISCMSG.DAT` 的載入起點 0x3AB,
// 正好落在第 20..27 筆記錄的開頭。這條把那個對應釘住 ——
// 位移算錯的話,寶典會唸出別人的句子而且不會有人發現。
func TestCodexAnswersAreTheEightVirtuePassages(t *testing.T) {
	dir := gameDataDir(t)
	if dir == "" {
		return
	}
	tf, err := u5data.LoadText(dir + "/MISCMSG.DAT")
	if err != nil {
		t.Fatal(err)
	}
	// 原版 `dword_5604C`,相對於載入起點 0x3AB。
	want := []int{256, 334, 393, 468, 555, 639, 704, 784}
	const base = 0x3AB
	for v, off := range want {
		rec := u5data.CodexAnswerRecord(v)
		if rec >= len(tf.Records) {
			t.Fatalf("記錄 %d 超出 %d 筆", rec, len(tf.Records))
		}
		if got := tf.Records[rec].Offset; got != base+off {
			t.Errorf("第 %d 德:CodexAnswerRecord = %d(位移 0x%X),"+
				"原版指的是 0x%X", v, rec, got, base+off)
		}
	}
}

// TestCodexGivesTheLowestActiveQuest:同時有多項試煉時,寶典先給編號小的。
//
// 原版是 `for (i = 0; i < 8 && !(byte_3E0DC & (1<<i)); i++)` —— 由小到大掃,
// 停在第一個。寫成「最後領的那一項」的話,玩家同時開兩項試煉時會拿錯答案。
func TestCodexGivesTheLowestActiveQuest(t *testing.T) {
	s := shrineState(t)
	if s == nil {
		return
	}
	s.ShrineQuestActive = 1<<u5data.VirtueValor | 1<<u5data.VirtueHumility
	pages := codexPages(t, s)
	want := s.miscText(u5data.CodexAnswerRecord(u5data.VirtueValor))
	if !strings.Contains(strings.Join(pages, "\n"), want) {
		t.Errorf("寶典沒有唸出勇氣那一段:\n%s", strings.Join(pages, "\n"))
	}
	if s.ShrineQuestGiven != 1<<u5data.VirtueValor {
		t.Errorf("已讀位元是 %02X,只該設勇氣那一位", s.ShrineQuestGiven)
	}
}

// TestCodexWithNoQuestAsksHowYouGotHere:沒有進行中的試煉就問「汝是怎麼來的」。
func TestCodexWithNoQuestAsksHowYouGotHere(t *testing.T) {
	s := shrineState(t)
	if s == nil {
		return
	}
	s.ShrineQuestActive = 0
	pages := codexPages(t, s)
	want := s.miscText(u5data.MsgCodexNoQuest)
	if !strings.Contains(strings.Join(pages, "\n"), want) {
		t.Errorf("沒有印出「%s」:\n%s", want, strings.Join(pages, "\n"))
	}
	if s.ShrineQuestGiven != 0 {
		t.Errorf("沒有試煉卻設了已讀位元 %02X", s.ShrineQuestGiven)
	}
}

// TestCodexRunesOnTheEighth:四段符文在讀到第八德的那一次出現。
//
// ⚠ 原版**沒有防重播的旗標**,只有一句 `cmp byte_3E0DE, 0FFh`。
// 「符文只看得到一次」是**流程逼出來的**:八德都領過之後聖壇不會再發試煉,
// 所以再來寶典時找不到進行中的試煉,根本走不到播符文那一行。
// 這條測試兩件事都驗 —— 湊滿那次會播、而且之後聖壇確實不再發試煉。
func TestCodexRunesOnTheEighth(t *testing.T) {
	s := shrineState(t)
	if s == nil {
		return
	}
	rune1 := s.miscText(u5data.MsgCodexRuneOne)
	if rune1 == "" {
		t.Fatal("第 41 筆是空的 —— MISCMSG 位移大概錯了")
	}

	// 只差最後一德:這一次要播符文。
	s.ShrineQuestGiven = 0x7F
	s.ShrineQuestActive = 1 << 7
	joined := strings.Join(codexPages(t, s), "\n")
	if !strings.Contains(joined, rune1) {
		t.Errorf("湊滿八德的那一次沒有播符文:\n%s", joined)
	}
	if s.ShrineQuestGiven != 0xFF {
		t.Errorf("已讀位元是 %02X,應該滿了", s.ShrineQuestGiven)
	}

	// 八德都領過之後,聖壇只剩捐錢那條路 —— 不會再有進行中的試煉。
	// 這才是「符文看不到第二次」真正的原因。
	s.ShrineQuestActive = 0
	s2 := atShrine(t, u5data.VirtueHonesty)
	s2.ShrineQuestGiven, s2.ShrineQuestActive = 0xFF, 0
	answer(s2, u5data.VirtueHonesty)
	if s2.ShrineQuestActive != 0 {
		t.Errorf("八德全領過之後聖壇又發了試煉(active=%02X)", s2.ShrineQuestActive)
	}
	// 所以再去寶典只會被問「汝是怎麼來到這裡的?」。
	joined = strings.Join(codexPages(t, s), "\n")
	if strings.Contains(joined, rune1) {
		t.Errorf("沒有進行中的試煉卻播了符文:\n%s", joined)
	}
	if !strings.Contains(joined, s.miscText(u5data.MsgCodexNoQuest)) {
		t.Errorf("沒有進行中的試煉時寶典沒問「汝是怎麼來的」:\n%s", joined)
	}
}

// TestCodexRunePassageNamesVeramocor:符文那四段要提到 VERAMOCOR。
//
// 這是遊戲的主線提示:末日之門要用第八個力量之言打開。
// **[HARD] 它是玩家要打出來的字,譯文裡必須維持英文原樣。**
func TestCodexRunePassageNamesVeramocor(t *testing.T) {
	s := shrineState(t)
	if s == nil {
		return
	}
	var all string
	for i := 0; i < u5data.CodexRunePages; i++ {
		all += s.miscText(u5data.MsgCodexRuneOne+i) + "\n"
	}
	if !strings.Contains(all, u5data.WordsOfPower[u5data.VirtueHumility]) {
		t.Errorf("四段符文裡找不到 VERAMOCOR —— 譯文大概把它翻成中文了:\n%s", all)
	}
}

// TestShrineChamberCostsSixteenMinutes:進出聖壇 / 寶典要走掉十六分鐘。
//
// 原版 `sub_1DA10` 尾端的 `sub_29304(0x10)`。少了它,拜壇與讀寶典都是免費的,
// 而 U5 的 NPC 排程、月相、光源全都吃這個時鐘。
func TestShrineChamberCostsSixteenMinutes(t *testing.T) {
	s := shrineState(t)
	if s == nil {
		return
	}
	before := s.Clock.Hour*60 + s.Clock.Minute
	codexPages(t, s)
	after := s.Clock.Hour*60 + s.Clock.Minute
	if d := (after - before + 24*60) % (24 * 60); d != u5data.ShrineChamberMinutes {
		t.Errorf("讀完寶典過了 %d 分鐘,預期 %d", d, u5data.ShrineChamberMinutes)
	}

	s2 := atShrine(t, u5data.VirtueHonesty)
	before = s2.Clock.Hour*60 + s2.Clock.Minute
	answer(s2, u5data.VirtueHonesty)
	after = s2.Clock.Hour*60 + s2.Clock.Minute
	if d := (after - before + 24*60) % (24 * 60); d != u5data.ShrineChamberMinutes {
		t.Errorf("拜完壇過了 %d 分鐘,預期 %d", d, u5data.ShrineChamberMinutes)
	}
}

// TestEnterDispatchesByTile:進入是看腳下的地形,不是查座標表。
//
// 原版 `sub_2D72C` 用地形分派(17 = 法典聖壇、25 = 八德聖壇)。
// ⚠ 用地形而不是座標,幽冥界的靈性聖壇才會動 —— 它的座標不在
// `u5data.Shrines` 表上,那一格的 (0,0) 是 `ShrineAt` 的 fallback。
func TestEnterDispatchesByTile(t *testing.T) {
	s := shrineState(t)
	if s == nil {
		return
	}
	// 隨便一格,踩上聖壇地形就該開始冥想。
	s.X, s.Y = 100, 100
	s.SetTileAt(s.X, s.Y, u5data.TileShrine)
	s.Enter()
	if s.Prompt != PromptShrine {
		t.Errorf("站在聖壇地形上按 E,輸入模式是 %v:\n%s", s.Prompt, s.log())
	}
	s.EndMeditate()

	s.SetTileAt(s.X, s.Y, u5data.TileCodex)
	s.Enter()
	if s.Prompt != PromptCodex {
		t.Errorf("站在寶典地形上按 E,輸入模式是 %v:\n%s", s.Prompt, s.log())
	}
	for s.AdvanceCodex() {
	}
	if s.Prompt != PromptNone {
		t.Error("讀完寶典還卡在輸入模式")
	}
}
