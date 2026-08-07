package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// TestYellOnShipTogglesSailsWithoutAsking:在船上按 Yell 直接收放帆,不問任何字。
//
// ⚠ 兩個容易寫反的地方各驗一次:
//  1. **不會跳出輸入框** —— 多一個「喊什麼?」就是多一個原版沒有的按鍵
//  2. **0x20 是揚著帆**,按下去是「收帆」(+4);寫反的話玩家會發現
//     停在港口的船一按就開始跑
func TestYellOnShipTogglesSailsWithoutAsking(t *testing.T) {
	s := shrineState(t)
	if s == nil {
		return
	}
	s.Transport = u5data.VehicleSailing
	if s.Yell() {
		t.Error("在船上按 Yell 卻跳出了輸入框")
	}
	if s.Prompt != PromptNone {
		t.Errorf("輸入模式變成 %v,應該維持 PromptNone", s.Prompt)
	}
	if s.Transport != u5data.VehicleShip {
		t.Errorf("收帆之後載具是 %02X,預期 %02X", s.Transport, u5data.VehicleShip)
	}
	if !strings.Contains(s.log(), MsgFurl) {
		t.Errorf("沒有印出「%s」:\n%s", MsgFurl, s.log())
	}
	s.Yell()
	if s.Transport != u5data.VehicleSailing {
		t.Errorf("揚帆之後載具是 %02X,預期 %02X", s.Transport, u5data.VehicleSailing)
	}
	// 四個朝向都要能收放,不是只有 0x20 那一個。
	for _, tr := range []byte{
		u5data.VehicleSailing + 1, u5data.VehicleSailing + 2, u5data.VehicleSailing + 3,
	} {
		s.Transport = tr
		s.Yell()
		if s.Transport != tr+4 {
			t.Errorf("載具 %02X 收帆變成 %02X,預期 %02X", tr, s.Transport, tr+4)
		}
		s.Yell()
		if s.Transport != tr {
			t.Errorf("載具 %02X 收放各一次沒有回到原點(現在 %02X)", tr, s.Transport)
		}
	}
}

// TestYellOnFootAsks:不在船上時 Yell 會問「喊什麼?」。
func TestYellOnFootAsks(t *testing.T) {
	s := shrineState(t)
	if s == nil {
		return
	}
	s.Transport = u5data.VehicleWalk
	if !s.Yell() {
		t.Fatal("走路時按 Yell 沒有問話")
	}
	if s.Prompt != PromptYell {
		t.Errorf("輸入模式是 %v,預期 PromptYell", s.Prompt)
	}
	// 什麼都沒打就送出 = 什麼也沒喊,而且要退出輸入模式。
	s.SubmitYell()
	if s.Prompt != PromptNone {
		t.Error("送出空字串之後還卡在輸入模式")
	}
	if !strings.Contains(s.log(), MsgYellNothing) {
		t.Errorf("沒有印出「%s」:\n%s", MsgYellNothing, s.log())
	}
}

// TestYellRoutesByLocation:同一個鍵在不同地點做不同的事。
//
// 大地圖(0)→ 力量之言;三團聖火所在(30/31/32)→ 暗影君主;
// 其餘城鎮 → 毫無效果。這條驗的是**分派**,不是各分支的內容。
func TestYellRoutesByLocation(t *testing.T) {
	s := shrineState(t)
	if s == nil {
		return
	}
	s.Transport = u5data.VehicleWalk
	s.X, s.Y = 100, 100 // 空地:力量之言會走到「毫無效果」

	// 大地圖:認得 FALLAX 是力量之言(雖然旁邊沒有目標)。
	s.Location = 0
	s.Messages = nil
	s.Yell()
	s.Input = "FALLAX"
	s.SubmitYell()
	if !strings.Contains(s.log(), MsgWordUttered) {
		t.Errorf("大地圖上 FALLAX 沒被當成力量之言:\n%s", s.log())
	}

	// 一般城鎮:同一個字什麼也不會發生。
	s.Location = 2 // 不列顛城
	s.Messages = nil
	s.Yell()
	s.Input = "FALLAX"
	s.SubmitYell()
	if strings.Contains(s.log(), MsgWordUttered) {
		t.Errorf("在城裡 FALLAX 也被當成力量之言了:\n%s", s.log())
	}
	if !strings.Contains(s.log(), MsgNoEffect) {
		t.Errorf("在城裡喊力量之言沒有印「%s」:\n%s", MsgNoEffect, s.log())
	}
}

// TestShadowlordSummonNeedsAFlameKeep:暗影君主只在三團聖火所在召得出來。
//
// ⚠ 而且**沒有「名字要配這座城」的檢查** —— 原版在三座裡的任何一座喊任何
// 一個名字都成立。這裡連同「別的城喊不出來」一起驗,免得日後有人「順手」
// 加上配對檢查。
func TestShadowlordSummonNeedsAFlameKeep(t *testing.T) {
	s := shrineState(t)
	if s == nil {
		return
	}
	// 慈悲修道院(31)—— 三團聖火之一。
	if err := s.SetScene(31, 0, 15, 15); err != nil {
		t.Skipf("進不了慈悲修道院:%v", err)
	}
	s.Messages = nil
	s.Yell()
	s.Input = "NOSFENTOR" // 名字與這座城「不配」,原版照樣召得出來
	s.SubmitYell()
	if !strings.Contains(s.log(), MsgShadowlordAppears) {
		t.Fatalf("在慈悲修道院喊 NOSFENTOR 沒有召出暗影君主:\n%s", s.log())
	}
	if s.ShadowlordHere != 2 {
		t.Errorf("召出來的是第 %d 個,預期第 2 個(NOSFENTOR)", s.ShadowlordHere)
	}
	if !s.shadowlordPresent() {
		t.Error("召喚成功但場上找不到暗影君主")
	}
	// 場上已經有一個了,再喊一次不會多出第二個。
	s.Messages = nil
	s.Yell()
	s.Input = "FAULINEI"
	s.SubmitYell()
	if strings.Contains(s.log(), MsgShadowlordAppears) {
		t.Errorf("場上已有暗影君主卻又召出一個:\n%s", s.log())
	}

	// 換一座不是聖火所在的城:同樣的名字沒有反應。
	s2 := shrineState(t)
	if err := s2.SetScene(2, 0, 15, 15); err != nil { // 不列顛城
		t.Skipf("進不了不列顛城:%v", err)
	}
	s2.Messages = nil
	s2.Yell()
	s2.Input = "FAULINEI"
	s2.SubmitYell()
	if strings.Contains(s2.log(), MsgShadowlordAppears) {
		t.Errorf("在不列顛城也召出了暗影君主:\n%s", s2.log())
	}
}

// TestSummonedShadowlordDoesNotLeakIntoTheNPCTable:召出來的東西不能污染 NPC 表。
//
// `NPCSet.At` 回傳的是**指向共用陣列的指標**。直接往那裡寫的話,離場再進來
// 暗影君主還在,而且同一個 `.NPC` 檔裡別的地點也會跟著中招 ——
// 原版那張表是每次進場景重新載入的暫存副本,行為完全不同。
func TestSummonedShadowlordDoesNotLeakIntoTheNPCTable(t *testing.T) {
	s := shrineState(t)
	if s == nil {
		return
	}
	if err := s.SetScene(31, 0, 15, 15); err != nil {
		t.Skipf("進不了慈悲修道院:%v", err)
	}
	s.Yell()
	s.Input = "ASTAROTH"
	s.SubmitYell()
	if !s.shadowlordPresent() {
		t.Fatalf("沒召出來:\n%s", s.log())
	}
	// 重新進同一個場景 —— NPC 表該是乾淨的。
	if err := s.SetScene(31, 0, 15, 15); err != nil {
		t.Fatal(err)
	}
	if s.shadowlordPresent() {
		t.Error("重新進場景之後暗影君主還在 —— NPC 表被污染了")
	}
}

// TestMatchPrefixIsPrefixNotSubstring:原版的字串比對是前綴,不是相等也不是子字串。
//
// `sub_27C98` 把參考字截到 9 個字元、轉大寫,再交給 `sub_39554` 做
// `strncmp(參考字, 玩家打的, 參考字長度)`。三種語意的差別實際會影響玩法:
//
//	相等  → `ahmxyz` 不算(原版算)
//	子字串 → `the ahm` 算(原版不算)
func TestMatchPrefixIsPrefixNotSubstring(t *testing.T) {
	cases := []struct {
		needle, typed string
		want          bool
	}{
		{"Ahm", "Ahm", true},
		{"Ahm", "ahm", true},      // 大小寫不分
		{"Ahm", "ahmxyz", true},   // 後面多打字照樣過
		{"Ahm", "the ahm", false}, // 不是開頭就不算
		{"Ahm", "Ah", false},      // 打不完不算
		{"hone", "honesty", true},
		{"hono", "honesty", false}, // 誠實與榮譽只差第四個字母
		{"FALLAX", "fallax", true},
		{"VERAMOCOR", "veramocor", true}, // 剛好九個字元
		// Compassion 有十個字元,原版只比得到前九個。
		{"Compassion", "compassio", true},
	}
	for _, c := range cases {
		if got := u5data.MatchPrefix(c.needle, c.typed); got != c.want {
			t.Errorf("MatchPrefix(%q, %q) = %v,預期 %v", c.needle, c.typed, got, c.want)
		}
	}
}

// TestRestoreShrineWantsTheFullVirtueName:復原聖壇問的是完整美德名,不是四字母前綴。
//
// 冥想查 `off_55FEC`(`valo`),復原查 `off_411BC`(`Valour`)—— 兩張不同的表。
// 而且 `off_411BC` 用的是**英式拼法 Valour**,與 `u5data.Shrines[2].Name`
// 的 `Valor` 不同;原版兩處就不一致,不要「順手統一」。
func TestRestoreShrineWantsTheFullVirtueName(t *testing.T) {
	if u5data.VirtueNames[u5data.VirtueValor] == u5data.Shrines[u5data.VirtueValor].Name {
		t.Fatal("兩張表的勇氣拼法變成一樣了 —— 原版是 Valour / Valor,有一邊被改過")
	}
	v := u5data.VirtueValor
	// 冥想:四字母前綴,打 Valor 或 Valour 都過。
	for _, typed := range []string{"valor", "valour"} {
		if !u5data.MatchPrefix(u5data.Shrines[v].Prefix, typed) {
			t.Errorf("冥想時打 %q 卻不認得", typed)
		}
	}
	// 復原:要完整的 Valour,少一個 u 不算。
	if u5data.MatchPrefix(u5data.VirtueNames[v], "Valor") {
		t.Error("復原勇氣聖壇打 Valor 也過了 —— 原版要的是 Valour")
	}
	if !u5data.MatchPrefix(u5data.VirtueNames[v], "Valour") {
		t.Error("復原勇氣聖壇打 Valour 卻不過")
	}
}

// TestQuestFlagsSurviveSaveRoundTrip:任務旗標要存得回去也讀得回來。
//
// ⚠ 這幾個位移是「跟著讀取序列累加」算出來的,算錯不會報錯 ——
// 只會讓玩家的封印、玷污、試煉進度在讀檔之後靜靜歸零。
// 所以這裡改了值再走一次 Encode → ParseSave,順便確認沒有踩到隔壁欄位。
func TestQuestFlagsSurviveSaveRoundTrip(t *testing.T) {
	s := shrineState(t)
	if s == nil {
		return
	}
	if s.BaseSave == nil {
		t.Skip("沒有底稿存檔")
	}
	s.ShrineQuestActive = 0x25
	s.ShrineQuestGiven = 0x81
	s.DungeonSeal[3] = u5data.DungeonSealedBit
	s.ShrineFlag[6] = 0xFF
	s.ShadowlordAt = [u5data.ShadowlordCount]byte{4, 0xFF, 7}
	s.ShadowlordHere = 1
	karma := s.Karma

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
	switch {
	case back.ShrineQuest != 0x25:
		t.Errorf("試煉進度讀回來是 %02X,預期 25", back.ShrineQuest)
	case back.CodexLearned != 0x81:
		t.Errorf("寶典進度讀回來是 %02X,預期 81", back.CodexLearned)
	case back.DungeonSeal[3] != u5data.DungeonSealedBit:
		t.Errorf("地牢封印讀回來是 %02X", back.DungeonSeal[3])
	case back.ShrineFlag[6] != 0xFF:
		t.Errorf("聖壇旗標讀回來是 %02X", back.ShrineFlag[6])
	case back.ShadowlordAt != [u5data.ShadowlordCount]byte{4, 0xFF, 7}:
		t.Errorf("暗影君主位置讀回來是 %v", back.ShadowlordAt)
	case back.ShadowlordHere != 1:
		t.Errorf("召喚中的暗影君主讀回來是 %d", back.ShadowlordHere)
	case back.Karma != karma:
		// 隔壁欄位沒被踩到的哨兵。
		t.Errorf("業報從 %d 變成 %d —— 新欄位大概寫到別人身上了", karma, back.Karma)
	}
}

// TestShadowlordNamesAreDistinct:三個名字不能互為前綴。
//
// 比對是前綴,所以只要有一個名字是另一個的開頭,短的那個永遠搶先。
func TestShadowlordNamesAreDistinct(t *testing.T) {
	for i, a := range u5data.Shadowlords {
		if u5data.ShadowlordIndex(a) != i {
			t.Errorf("%s 查回第 %d 個,預期第 %d 個", a, u5data.ShadowlordIndex(a), i)
		}
	}
	if u5data.ShadowlordIndex("FALLAX") != -1 {
		t.Error("力量之言被當成暗影君主的名字")
	}
}
