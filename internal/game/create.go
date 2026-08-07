package game

import (
	"strings"

	"github.com/wicanr2/u5-cht/internal/i18n"
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 建立新角色(原版主選單的 `Create New Character`,流程在 `.text:000235B6`)
//
// 這是引擎此前**最大的缺口**:沒有它,玩家只能從既有的 `SAVED.GAM` /
// `INIT.GAM` 開局 —— 也就是只能扮演別人做好的聖者。
//
// ⚠ 這一條與 CLAUDE.md §6.1 的驗收鐵則直接相關:
// 「不要打包測試角色存檔,玩家該走建角流程」。流程不存在的時候,
// 那條規則是空的。
//
// # 流程
//
//	1. 吉普賽的開場白(`QUESTION.DAT` 記錄 0)
//	2. 七題八德淘汰賽(4 → 2 → 1),每題答 A 或 B
//	3. 結語(記錄 1)
//	4. 問名字
//	5. 問性別
//	6. 把結果寫進 `INIT.GAM` 的聖者那一格
//
// 屬性怎麼算見 `u5data/create.go` —— 那裡有「原版是 add 不是覆蓋」的說明。

// Creation 是進行中的建角。
type Creation struct {
	// Alive[v] 是這個美德還沒被淘汰。
	Alive [u5data.VirtueCount]bool
	// picked[v] 是**本輪**已經抽到過(原版 byte_5717C,每輪清空)。
	picked [u5data.VirtueCount]bool
	// A, B 是這一題對上的兩個美德;Swapped 記錄畫面上 A/B 的呈現順序。
	A, B int
	// Round 是第幾輪(0..2),Asked 是本輪問到第幾題。
	Round, Asked int
	// Total 是總共問完幾題。
	Total int
	// Intel, Dex, Str 是累加中的三圍。起點是既有角色的值,不是 0。
	Intel, Dex, Str int
	// Stage 是流程走到哪。
	Stage CreationStage
	// Name 是玩家打到一半的名字。
	Name string
	// Winner 是最後存活的美德;流程沒走完時是 -1。
	Winner int
}

// CreationStage 是建角流程的階段。
type CreationStage int

const (
	// CreationIntro 是吉普賽的開場白,按任意鍵繼續。
	CreationIntro CreationStage = iota
	// CreationQuestion 是七題之一,只收 A / B。
	CreationQuestion
	// CreationClosing 是問完之後的結語,按任意鍵繼續。
	CreationClosing
	// CreationName 是打名字。
	CreationName
	// CreationGender 是選性別,只收 M / F。
	CreationGender
	// CreationDone 是做完了。
	CreationDone
)

// BeginCreation 開始建角。
//
// 三圍的起點取自名冊第 0 名(`INIT.GAM` 的聖者)—— 原版在問答前
// 把 `byte_3DDC0..C2` 抄進累加器,那三個位元組就是既有角色的值。
func (s *State) BeginCreation() bool {
	c := &Creation{Winner: -1}
	for i := range c.Alive {
		c.Alive[i] = true
	}
	if len(s.Roster) > 0 {
		c.Intel = int(s.Roster[0].Intel)
		c.Dex = int(s.Roster[0].Dex)
		c.Str = int(s.Roster[0].Strength)
	}
	s.Create = c
	s.Prompt = PromptCreate
	s.Log(s.creationText(u5data.QuestionIntro))
	return true
}

// creationText 取 `QUESTION.DAT` 某一筆的譯文。
func (s *State) creationText(rec int) string {
	en := ""
	if s.Question != nil && rec >= 0 && rec < len(s.Question.Records) {
		en = s.Question.Records[rec].Text()
	}
	return i18n.Text(u5data.QuestionFile, rec, en)
}

// CreationPrompt 是現在畫面上該顯示的字。
func (s *State) CreationPrompt() string {
	c := s.Create
	if c == nil {
		return ""
	}
	switch c.Stage {
	case CreationIntro:
		return s.creationText(u5data.QuestionIntro)
	case CreationQuestion:
		return s.creationText(u5data.VirtueQuestion(c.A, c.B))
	case CreationClosing:
		return s.creationText(u5data.QuestionClosing)
	case CreationName:
		return MsgCreateName + c.Name
	case CreationGender:
		return MsgCreateGender
	}
	return ""
}

// AdvanceCreation 是「按任意鍵」:開場白與結語用。
func (s *State) AdvanceCreation() {
	c := s.Create
	if c == nil {
		return
	}
	switch c.Stage {
	case CreationIntro:
		c.Stage = CreationQuestion
		s.nextQuestion()
	case CreationClosing:
		c.Stage = CreationName
		s.Log(MsgCreateName)
	}
}

// nextQuestion 抽下一題的兩個美德(原版 `sub_23248` 抽兩次)。
//
// 抽法是**拒絕取樣**:隨機 0..7,抽到已抽過或已淘汰的就重抽。
// 兩個旗標分開 —— 「本輪抽過」每輪清空,「已淘汰」一路留著。
func (s *State) nextQuestion() {
	c := s.Create
	c.A = s.drawVirtue()
	c.B = s.drawVirtue()
	if c.A < 0 || c.B < 0 {
		c.Stage = CreationClosing
		return
	}
	s.Log(s.creationText(u5data.VirtueQuestion(c.A, c.B)))
}

// drawVirtue 抽一個還沒抽過、也還沒被淘汰的美德。回 -1 代表沒得抽了。
func (s *State) drawVirtue() int {
	c := s.Create
	// 拒絕取樣可能永遠抽不中,所以要有上限 —— 原版沒有(它保證抽得到),
	// 但引擎不該因為狀態被改壞就吊死。上限之外的行為與原版無異。
	for try := 0; try < 1000; try++ {
		v := s.Roll(0, u5data.VirtueCount-1)
		if !c.picked[v] && c.Alive[v] {
			c.picked[v] = true
			return v
		}
	}
	for v := range c.picked {
		if !c.picked[v] && c.Alive[v] {
			c.picked[v] = true
			return v
		}
	}
	return -1
}

// AnswerCreation 是玩家在七題裡選了 A 還是 B。
//
// pickA 為真代表選了畫面上的第一個(也就是 c.A)。
func (s *State) AnswerCreation(pickA bool) {
	c := s.Create
	if c == nil || c.Stage != CreationQuestion {
		return
	}
	winner, loser := c.A, c.B
	if !pickA {
		winner, loser = c.B, c.A
	}
	b := u5data.VirtueBonus[winner]
	c.Intel += b[u5data.BonusIntel]
	c.Dex += b[u5data.BonusDex]
	c.Str += b[u5data.BonusStr]
	c.Alive[loser] = false
	c.Winner = winner

	c.Asked++
	c.Total++
	if c.Asked < u5data.CreateRounds[c.Round] {
		s.nextQuestion()
		return
	}
	// 一輪結束:清掉「本輪抽過」,但**不清**「已淘汰」。
	for i := range c.picked {
		c.picked[i] = false
	}
	c.Round++
	c.Asked = 0
	if c.Round >= len(u5data.CreateRounds) {
		c.Stage = CreationClosing
		s.Log(s.creationText(u5data.QuestionClosing))
		return
	}
	s.nextQuestion()
}

// TypeCreationName 收名字的一個字元。空字串代表退格。
func (s *State) TypeCreationName(ch string) {
	c := s.Create
	if c == nil || c.Stage != CreationName {
		return
	}
	if ch == "" {
		if n := []rune(c.Name); len(n) > 0 {
			c.Name = string(n[:len(n)-1])
		}
		return
	}
	if len([]rune(c.Name)) < u5data.CharNameLen-1 {
		c.Name += ch
	}
}

// ConfirmCreationName 名字打完了(按 Enter)。
func (s *State) ConfirmCreationName() {
	c := s.Create
	if c == nil || c.Stage != CreationName || strings.TrimSpace(c.Name) == "" {
		return
	}
	c.Stage = CreationGender
	s.Log(MsgCreateGender)
}

// AnswerCreationGender 選性別,然後把角色寫進名冊。
func (s *State) AnswerCreationGender(male bool) {
	c := s.Create
	if c == nil || c.Stage != CreationGender {
		return
	}
	gender := byte(u5data.GenderFemale)
	if male {
		gender = u5data.GenderMale
	}
	s.applyCreation(gender)
	c.Stage = CreationDone
	s.Prompt = PromptNone
	s.Log(MsgCreateDone)
}

// applyCreation 把建角的結果寫進名冊第 0 名。
//
// ⚠ **只改該改的欄位。** 其餘(裝備、HP、經驗、位置)沿用 `INIT.GAM` ——
// 原版也是把新角色寫進那份初始存檔,不是從零造一個。自己補值等於自創數值。
func (s *State) applyCreation(gender byte) {
	c := s.Create
	if len(s.Roster) == 0 {
		return
	}
	ch := &s.Roster[0]
	ch.Name = strings.TrimSpace(c.Name)
	ch.Gender = gender
	ch.Intel = byte(c.Intel)
	// 初始魔力等於智力 —— 原版 byte_3DDC2 與 byte_3DDC3 寫的是同一個值。
	ch.MP = byte(c.Intel)
	ch.Dex = byte(c.Dex)
	str := c.Str
	if str < u5data.CreateMinStrength {
		str = u5data.CreateMinStrength
	}
	ch.Strength = byte(str)
}
