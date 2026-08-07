package game

import (
	"strings"

	"github.com/wicanr2/u5-cht/internal/i18n"
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 黑棘的審問(原版 `sub_C414` → `sub_C318`)
//
// 被抓進黑棘宮殿之後,他會就**第一座還沒被玷污的聖壇**逼問真言,問四次。
//
//	招了(答案裡出現真言) → 那座聖壇被玷污、業報 −5,
//	                        隊伍還有第二個人的話,同伴被「仁慈地」處決
//	一路不招               → 隊伍只剩一人:「小孩都拆穿得了汝的謊」,下地牢
//	                        還有同伴:第一次只被嗆,之後刑具一格一格推進,
//	                        第四次鍘刀落下,一名同伴被斬
//
// 不論結局,最後都被丟進宮殿地牢((10,7) 第 −1 層),鑰匙歸零。
//
// ⚠ **招與不招都會死人。** 招了是「賞汝一個痛快」,不招是「為汝的背叛付出代價」——
// 兩條路各自的訊息不同,但都走 `sub_C13C`。寫成「招了就沒事」會把這一幕
// 最重要的張力抹掉。
//
// ⚠ **八座聖壇全被玷污之後這一幕不會發生**(`sub_C414` 掃不到就直接跳過)。

// Blackthorn 是進行中的審問。
type Blackthorn struct {
	// Virtue 是這一次被逼問的聖壇(第一座還沒被玷污的)。
	Virtue int
	// Round 是問到第幾次(0..BlackthornRounds−1)。
	Round int
	// Taunted 記錄有沒有嗆過話 —— 原版的 `ebx`,第一次拒絕只嗆不動刑。
	Taunted bool
}

// BlackthornShrine 回傳第一座還沒被玷污的聖壇;全都玷污了就回 −1。
//
// 原版 `for (i = 0; i < 8 && byte_3E0E8[i] != 0; i++)` —— 注意判的是
// **整個位元組非 0**,不只是 bit 7。復原之後那一格留下的是 0x7F(仍然非 0),
// 所以**復原過的聖壇不會再被挑中**。這不是 bug,是原版的行為。
func (s *State) BlackthornShrine() int {
	for i := range s.ShrineFlag {
		if s.ShrineFlag[i] == 0 {
			return i
		}
	}
	return -1
}

// BeginInterrogation 開始審問。回傳有沒有真的開始。
func (s *State) BeginInterrogation() bool {
	v := s.BlackthornShrine()
	if v < 0 {
		// 八座全被玷污 —— 原版直接跳過整幕。
		s.releaseToCell()
		return false
	}
	s.Log(MsgSubdued)
	s.Log(MsgDraggedAway)
	s.Log("黑棘說道:「啊," + s.AvatarName() + "!能與汝相見,實為榮幸。」")
	s.logMisc(u5data.MsgBlackthornWait)
	s.Blackthorn = &Blackthorn{Virtue: v}
	s.Prompt = PromptBlackthorn
	s.askBlackthorn()
	return true
}

// askBlackthorn 問這一輪的問題。
func (s *State) askBlackthorn() {
	b := s.Blackthorn
	q := s.miscText(u5data.BlackthornQuestion(b.Round))
	if u5data.BlackthornQuestionNamesVirtue(b.Round) {
		// ⚠ 前三輪的問句都是**半句**,句尾要接美德名再補 `?」`
		//(原版 `sub_BFFC` 接 `off_411BC[i]` 之後再接 `asc_48C60` = `?"`)。
		// 這裡自己補一個空格:譯文原本就以「—— 」收尾,但 `miscText`
		// 會把尾巴的空白修掉,不補的話會變成「什麼 ——Honesty?」。
		q += " " + u5data.VirtueNames[b.Virtue] + "?」"
	}
	s.Log(q)
	s.Log(MsgYourResponse)
}

// AnswerBlackthorn 收玩家打的一行。回傳審問還在不在進行中。
func (s *State) AnswerBlackthorn(text string) bool {
	b := s.Blackthorn
	if b == nil {
		return false
	}
	s.Input = ""
	if u5data.MantraSpoken(u5data.Shrines[b.Virtue].Mantra, text) {
		s.confess()
		return false
	}
	// 沒招。隊伍只剩一個人時黑棘不再玩下去。
	if s.livingCompanions() < 2 {
		s.logMisc(u5data.MsgBlackthornLies)
		s.endInterrogation()
		return false
	}
	if !b.Taunted {
		// ⚠ 第一次拒絕只嗆一句,不動刑(原版 `and ebx, ebx; jz`)。
		b.Taunted = true
		s.logMisc(u5data.MsgBlackthornTaunt)
		// ⚠ 第 8 筆也是半句 —— 原版接的是**隊伍第 2 人的名字**再接 " die!\""
		//(`sub_C2D0` push 的 `dword_3DDD4` 就是名冊第 1 格)。
		// 不接的話這句會停在逗號上,而且威脅對象變得不明。
		s.Log(s.miscText(u5data.MsgBlackthornSand) + s.nextVictimName() + "就得死!」")
	} else if b.Round == u5data.BlackthornRounds-1 {
		// 最後一輪:鍘刀落下。
		s.executeCompanion(true)
	} else {
		// 中間幾輪:刑具往前推一格,時間走兩分鐘。
		s.AdvanceTime(BlackthornTortureMinutes)
	}
	b.Round++
	if b.Round >= u5data.BlackthornRounds {
		s.endInterrogation()
		return false
	}
	s.askBlackthorn()
	return true
}

// BlackthornTortureMinutes 是每推進一格刑具走掉的時間(原版 `sub_29304(2)`)。
const BlackthornTortureMinutes = 2

// nextVictimName 是下一個會被處決的同伴 —— 名冊第 1 格(原版 `dword_3DDD4`)。
func (s *State) nextVictimName() string {
	if len(s.Roster) > 1 {
		return i18n.Name(s.Roster[1].Name)
	}
	return ""
}

// confess 是招供之後的結局。
func (s *State) confess() {
	b := s.Blackthorn
	// ⚠ 玷污寫的是 **0xFF 整個位元組**,不是只設 bit 7 ——
	// 復原時只清 bit 7 留下 0x7F,而 `BlackthornShrine` 判的是「非 0」,
	// 所以復原過的聖壇之後不會再被挑中。兩邊要一起看才對得起來。
	s.ShrineFlag[b.Virtue] = 0xFF
	s.Karma = subFloor(s.Karma, u5data.BlackthornKarmaPenalty)
	if s.livingCompanions() > 1 {
		s.executeCompanion(false)
	} else {
		s.logMisc(u5data.MsgBlackthornTruth)
	}
	s.endInterrogation()
}

// livingCompanions 是隊伍裡還活著的人數(原版 `byte_3DDBF[i*32] != 'D'`)。
func (s *State) livingCompanions() int {
	n := 0
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		if s.Roster[i].Status != u5data.StatusDead {
			n++
		}
	}
	return n
}

// executeCompanion 處決一名同伴(原版 `sub_C13C`)。
//
// ⚠ 死的是**第二個活著的人**,不是隨機挑、也不是最弱的那個 ——
// 原版數到 `ebx == 2` 就停手(所以聖者本人永遠活著)。
//
// ⚠ 而且不是「標記成死亡」:那個人被**搬到名冊第 15 格**、名冊其餘往前遞補、
// 隊伍人數 −1,並在第 15 格的位移 31 寫下 0x7F。那個 0x7F 就是寶典石室前
// 「骨灰罈」的來源(`sub_1DA10` 掃的就是它)。只標記死亡的話,
// 之後去寶典看不到罈子,而且名冊會留一個空洞。
func (s *State) executeCompanion(blade bool) {
	idx, seen := -1, 0
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		if s.Roster[i].Status == u5data.StatusDead {
			continue
		}
		if seen++; seen == 2 {
			idx = i
			break
		}
	}
	if idx < 0 {
		return
	}
	victim := s.Roster[idx]
	// 名冊往前遞補,被處決的那個移到最後一格。
	last := len(s.Roster) - 1
	copy(s.Roster[idx:last], s.Roster[idx+1:])
	victim.Raw[u5data.CharUrn] = u5data.CharUrnMark
	s.Roster[last] = victim
	s.PartySize--

	if blade {
		s.logMisc(u5data.MsgBlackthornBlade)
		s.Log(i18n.Name(victim.Name) + "被劈成了兩半!")
		s.logMisc(u5data.MsgBlackthornPaid)
	} else {
		s.logMisc(u5data.MsgBlackthornMercy)
	}
}

// endInterrogation 收掉審問,把隊伍丟進宮殿地牢。
func (s *State) endInterrogation() {
	s.Blackthorn = nil
	if s.Prompt == PromptBlackthorn {
		s.Prompt = PromptNone
	}
	s.releaseToCell()
}

// releaseToCell 把隊伍搬到黑棘宮殿的地牢(原版 `sub_C414` 尾段)。
//
// ⚠ **鑰匙歸零**(`byte_3DFB8 = 0`)—— 不然玩家一被關進去就能開門走人。
func (s *State) releaseToCell() {
	s.Inventory.Keys = 0
	s.Transport = u5data.VehicleWalk
	if s.Scenes == nil {
		return
	}
	if err := s.SetScene(u5data.BlackthornLocation,
		u5data.BlackthornCellFloor, u5data.BlackthornCellX, u5data.BlackthornCellY); err != nil {
		s.Log("讀不到黑棘宮殿的地牢(" + err.Error() + ")。")
	}
}

// SubmitBlackthorn 把打好的一行送進審問。
func (s *State) SubmitBlackthorn() {
	text := s.Input
	s.Input = ""
	s.AnswerBlackthorn(strings.TrimSpace(text))
}
