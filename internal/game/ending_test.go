package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

func endState(t *testing.T) *State {
	t.Helper()
	s := shrineState(t)
	if s == nil {
		return nil
	}
	tf, err := u5data.LoadText(gameDataDir(t) + "/ENDMSG.DAT")
	if err != nil {
		t.Fatal(err)
	}
	s.EndMsg = tf
	return s
}

// TestEndingMessageOffsetsMatchTheOriginal:十一筆結局訊息的序號要對得上。
//
// ★ `sub_135FC` 是拿**寫死的位移**取字(`byte_54700 + n`,`ENDMSG.DAT`
// 從檔頭載入)。這條把那十個位移逐一對回記錄開頭 —— 算錯的話結局會唸錯句子,
// 而且不會有任何東西報錯。
func TestEndingMessageOffsetsMatchTheOriginal(t *testing.T) {
	s := endState(t)
	if s == nil {
		return
	}
	want := []struct {
		offset, record int
	}{
		{0x000, u5data.MsgEndWellMet},
		{0x021, u5data.MsgEndAskBox},
		{0x049, u5data.MsgEndAskBox2},
		{0x0AB, u5data.MsgEndOpensBox},
		{0x0D3, u5data.MsgEndArtifact},
		{0x128, u5data.MsgEndOurWorld},
		{0x167, u5data.MsgEndFreeUs},
		{0x1C9, u5data.MsgEndOlder},
		{0x211, u5data.MsgEndOrb},
		{0x24B, u5data.MsgEndFollow},
		{0x2D5, u5data.MsgEndPullUpAChair},
	}
	for _, w := range want {
		if w.record >= len(s.EndMsg.Records) {
			t.Fatalf("記錄 %d 超出 %d 筆", w.record, len(s.EndMsg.Records))
		}
		if got := s.EndMsg.Records[w.record].Offset; got != w.offset {
			t.Errorf("記錄 %d 的位移是 0x%X,原版指的是 0x%X", w.record, got, w.offset)
		}
	}
}

// TestEndingRevivesTheDeadUnconditionally:結局一開始就把死掉的隊員救回來。
//
// ⚠ 兩件事都要做:狀態改回 'G',**而且**目前 HP 補成最大 HP。
// 只改狀態的話人活過來但血是 0。
// ⚠ 而且是**無條件**的 —— 在問「盒子帶了沒」之前就做完了。
func TestEndingRevivesTheDeadUnconditionally(t *testing.T) {
	s := endState(t)
	if s == nil {
		return
	}
	if s.PartySize < 2 {
		t.Skip("隊伍太少")
	}
	s.Roster[1].Status = u5data.StatusDead
	s.Roster[1].HP = 0
	s.SandalwoodBox = false // 就算沒帶盒子也要復活
	if !s.BeginEnding() {
		t.Fatalf("結局沒有開始:\n%s", s.log())
	}
	if s.Roster[1].Status != u5data.StatusGood {
		t.Errorf("狀態還是 %c", s.Roster[1].Status)
	}
	if s.Roster[1].HP != s.Roster[1].MaxHP {
		t.Errorf("HP 是 %d,預期補滿到 %d", s.Roster[1].HP, s.Roster[1].MaxHP)
	}
	if !strings.Contains(s.log(), "活過來了") {
		t.Errorf("沒有印復活訊息:\n%s", s.log())
	}
}

// TestSayingYesWithoutTheBoxIsTheBadEnding:嘴上說有不算,得真的帶著盒子。
//
// ⚠ 原版是 `and`:回答要是 Y,**而且** `byte_3DFCD != 0`。
// 少了後者,那只盒子(整條主線最後一件收集品)就變成可有可無。
func TestSayingYesWithoutTheBoxIsTheBadEnding(t *testing.T) {
	s := endState(t)
	if s == nil {
		return
	}
	s.SandalwoodBox = false
	s.BeginEnding()
	s.AnswerEnding(true)
	if s.Ending == nil {
		t.Fatal("回答之後結局就結束了")
	}
	if s.Ending.Good {
		t.Error("沒帶盒子卻走了真結局")
	}
	for s.AdvanceEnding() {
	}
	if !strings.Contains(s.log(), s.endText(u5data.MsgEndPullUpAChair)) {
		t.Errorf("沒有印「搬張椅子坐下吧」:\n%s", s.log())
	}
	if strings.Contains(s.log(), s.endText(u5data.MsgEndOrb)) {
		t.Error("壞結局卻唸出了月之球那一段")
	}
}

// TestSayingYesWithTheBoxIsTheTrueEnding:帶著盒子答 Y 才看得到真結局。
func TestSayingYesWithTheBoxIsTheTrueEnding(t *testing.T) {
	s := endState(t)
	if s == nil {
		return
	}
	s.SandalwoodBox = true
	s.BeginEnding()
	s.AnswerEnding(true)
	if s.Ending == nil || !s.Ending.Good {
		t.Fatalf("帶著盒子答 Y 卻不是真結局:\n%s", s.log())
	}
	for s.AdvanceEnding() {
	}
	for _, r := range u5data.EndingFinale {
		if !strings.Contains(s.log(), s.endText(r)) {
			t.Errorf("真結局漏了記錄 %d:\n%s", r, s.log())
		}
	}
}

// TestAnsweringNoAsksOneMoreTime:答 N 會換個說法再問一次,而且只再問一次。
//
// 原版的第二問是另一段文字(把盒子講清楚:「吾寢室密道裡的那只檀香木盒」)。
// 少了它,玩家隨手按到 N 就直接吃到壞結局。
func TestAnsweringNoAsksOneMoreTime(t *testing.T) {
	s := endState(t)
	if s == nil {
		return
	}
	s.SandalwoodBox = true
	s.BeginEnding()
	s.AnswerEnding(false)
	if s.Ending == nil || !s.Ending.Asking {
		t.Fatalf("答 N 之後沒有再問一次:\n%s", s.log())
	}
	if !strings.Contains(s.log(), s.endText(u5data.MsgEndAskBox2)) {
		t.Errorf("第二問用的不是另一段文字:\n%s", s.log())
	}
	// 第二次還是 N → 壞結局,不再追問。
	s.AnswerEnding(false)
	if s.Ending.Asking {
		t.Error("問了第三次")
	}
	if s.Ending.Good {
		t.Error("兩次都答 N 卻走了真結局")
	}
}

// TestSandalwoodBoxSurvivesSaveRoundTrip:盒子存得回去也讀得回來。
func TestSandalwoodBoxSurvivesSaveRoundTrip(t *testing.T) {
	s := endState(t)
	if s == nil || s.BaseSave == nil {
		return
	}
	s.SandalwoodBox = true
	items := s.Inventory.Items
	sv, err := s.ExportSave(s.BaseSave)
	if err != nil {
		t.Fatal(err)
	}
	blob, err := sv.Encode()
	if err != nil {
		t.Fatal(err)
	}
	back, err := u5data.ParseSave(blob)
	if err != nil {
		t.Fatal(err)
	}
	if back.SandalwoodBox == 0 {
		t.Error("盒子讀回來不見了")
	}
	if back.Inventory.Items != items {
		// 0x0219 就緊貼在 0x021A 的裝備表前面 —— 偏一格最先被踩到的就是它。
		t.Error("裝備表被踩到了 —— 0x0219 這個位移大概算錯了")
	}
}

// 製作名單接在真結局後面,而且起算日與 INIT.GAM 對得上。
func TestCreditsFollowTheTrueEnding(t *testing.T) {
	if u5data.EpochYear != 139 || u5data.EpochMonth != 4 || u5data.EpochDay != 5 {
		t.Fatalf("起算日是 %d/%d/%d,原版是 139/4/5",
			u5data.EpochYear, u5data.EpochMonth, u5data.EpochDay)
	}
	// 剛好一年又一個月又一天。
	y, m, d := u5data.Elapsed(140, 5, 6)
	if y != 1 || m != 1 || d != 1 {
		t.Errorf("140/5/6 距起算日 %d 年 %d 月 %d 日,預期 1/1/1", y, m, d)
	}
	// 借位:日不夠先跟月借 28、月不夠再跟年借 13。
	y, m, d = u5data.Elapsed(140, 4, 4)
	if y != 0 || m != 12 || d != 27 {
		t.Errorf("140/4/4 算出 %d/%d/%d,預期 0 年 12 個月 27 天", y, m, d)
	}
	// 同一天通關 → 三個都是 0。
	if y, m, d = u5data.Elapsed(139, 4, 5); y != 0 || m != 0 || d != 0 {
		t.Errorf("起算日當天算出 %d/%d/%d,預期全 0", y, m, d)
	}
}

// ⚠ 那兩行符文維持英文 —— 原版是用符文字型畫出來的圖形。
func TestTheRuneLinesStayInEnglish(t *testing.T) {
	for _, line := range u5data.CreditsRunes {
		for _, r := range line {
			if r > 127 {
				t.Fatalf("符文行 %q 混進了非 ASCII 字元", line)
			}
		}
	}
	if u5data.CreditsRunes[0] != "THE QUEST OF THE AVATAR" {
		t.Errorf("第一行是 %q", u5data.CreditsRunes[0])
	}
}

// 只印非零的單位。
func TestElapsedTextSkipsZeroUnits(t *testing.T) {
	if got := elapsedText(0, 0, 3); got != "3天" {
		t.Errorf("只有天數時是 %q", got)
	}
	if got := elapsedText(2, 0, 1); got != "2年又1天" {
		t.Errorf("跳過 0 個月時是 %q", got)
	}
	if got := elapsedText(0, 0, 0); got != "" {
		t.Errorf("同一天通關應該是空字串,實得 %q", got)
	}
}

// 結局的觸發條件(原版 `sub_161E4`)。
//
// 站在戰場第 2 列、正北是城堡外牆 → 被吸進去 → 結局那一幕。
func TestStandingBelowACastleGateEndsTheGame(t *testing.T) {
	dir := gameDataDir(t)
	if dir == "" {
		return
	}
	s := endState(t)
	if s == nil {
		return
	}
	if !s.beginRoomCombat(&s.CombatMaps.Maps[0], 0) {
		t.Fatal("開不了戰鬥")
	}
	c := s.Combat
	// 把一個疊圖位元組落在 0x3C..0x3F 的單位放在 (5,1),
	// 讓第一個行動的單位站到它正南。
	// ⚠ 查的是**疊圖層**(物件 / 單位),不是地形 —— 見 docs/re/34 §4。
	putOverlay(c, 31, 0x3E, 5, 1)
	u := &c.Units[c.Turn]
	u.X, u.Y = 5, u5data.AbsorbRow
	if !s.checkAbsorbed() {
		t.Fatal("站在城門正南應該被吸進去")
	}
	if s.Ending == nil {
		t.Fatal("被吸進去之後應該進結局那一幕")
	}
	if !strings.Contains(allLogs(s), "吸") {
		t.Errorf("訊息裡沒說被吸進去:%q", allLogs(s))
	}
}

// ⚠ 差一列就不算 —— 查的是**正北**那一格,不是腳下。
//
// 這一條在釘 `byte_3F854` 的身分:它是戰場暫存(列距 16)的第 1 列,
// 不是「單位站的那一格」。當初讀成腳下的地形,整條鏈就接不起來。
func TestTheAbsorbCheckLooksNorthNotUnderfoot(t *testing.T) {
	dir := gameDataDir(t)
	if dir == "" {
		return
	}
	s := endState(t)
	if s == nil {
		return
	}
	s.beginRoomCombat(&s.CombatMaps.Maps[0], 0)
	c := s.Combat
	// 地形是城門**不算** —— 疊圖層與地形是兩層,原版查的是疊圖層。
	c.Map.Tiles[1][5] = 0x3E
	u := &c.Units[c.Turn]
	u.X, u.Y = 5, u5data.AbsorbRow
	if s.checkAbsorbed() {
		t.Error("地形是城門不該觸發 —— byte_3F844 是疊圖層,地形進不去")
	}
	// 疊圖在腳下那一格也不算。
	putOverlay(c, 31, 0x3E, 5, u5data.AbsorbRow+1)
	if s.checkAbsorbed() {
		t.Error("正南有疊圖不該觸發 —— 查的是正北")
	}
	// 列數不是 2 也不算。
	c.Units[31].X, c.Units[31].Y = 5, 3
	u.Y = 4
	if s.checkAbsorbed() {
		t.Error("第 4 列不該觸發 —— 原版寫死了第 2 列")
	}
}

// 四個城堡地形碼都算,鄰居不算(`& 0xFC == 0x3C`)。
func TestAbsorbTileGroup(t *testing.T) {
	for _, t2 := range []byte{0x3C, 0x3D, 0x3E, 0x3F} {
		if !u5data.AbsorbTile(t2) {
			t.Errorf("%02X 應該算城堡", t2)
		}
	}
	for _, t2 := range []byte{0x3B, 0x40, 0x00, 0x44} {
		if u5data.AbsorbTile(t2) {
			t.Errorf("%02X 不該算城堡", t2)
		}
	}
}

// putOverlay 在戰場某一槽塞一個單位,用來造出疊圖層的內容。
//
// 原版的疊圖是 `byte_3F844`(物件與 NPC 的 tile 位元組畫上去);
// 本引擎在戰鬥中的等價物就是站在那一格的單位。
func putOverlay(c *Combat, slot int, kind byte, x, y int) {
	c.Units[slot] = Combatant{
		Roster: -1, Creature: -1, Kind: kind, X: x, Y: y, HP: 5,
		Flags: UnitMonster,
	}
}
