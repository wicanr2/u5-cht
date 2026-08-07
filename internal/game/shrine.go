package game

import (
	"strings"

	"github.com/wicanr2/u5-cht/internal/i18n"
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 聖壇與冥想(原版 `sub_1D394`)
//
// 流程逐條照抄:
//
//	1. 站在哪一座 —— 掃座標表,沒中就當**靈性**(見 u5data.Shrines 的說明)
//	2. 問「汝欲冥想何種美德?」—— 比對美德名的**前四個字母**
//	3. 問**三次**「Mantra:」—— 每一次都要打對
//	4. 任何一次打錯 → 印「思緒散亂」,結束(而且**還是花掉了時間**)
//	5. 打對之後分三條路:
//	     還沒領過這座的試煉 → 領試煉(「去寶典學 X」)
//	     試煉進行中而且已回報 → 領獎:業報 +3、三圍 +1(謙遜額外 +3 業報)
//	     其餘                → 只能捐錢:每重 100 金,業報 +重數
//
// ⚠ **真言與美德名要打英文。** `[HARD]` 玩家在遊戲裡打不出中文,
// 而這兩個都是要玩家親手輸入的(`docs/manual/術語對照.md` §6)。
// 畫面上的提示會寫「誠實(honesty)」,讓人知道要按什麼。

// ShrineMantraTries 是原版問幾次真言(`cmp esi, 3`)。
//
// ⚠ **三次都要打對,不是「三次機會」。** 任何一次錯就失敗 ——
// 寫成「三次機會」會讓聖壇變得非常好過。
const ShrineMantraTries = 3

// Shrine 是進行中的冥想。
type Shrine struct {
	// Virtue 是這座聖壇的美德(0..7)。
	Virtue int
	// Stage 是進行到哪一步。
	Stage ShrineStage
	// Tries 是真言已經問過幾次。
	Tries int
	// OK 記錄到目前為止有沒有全對。
	//
	// ⚠ 原版**不會一錯就跳出** —— 它把三次真言問完才判。所以打錯之後
	// 還是得把流程走完。這個旗標就是原版的 `var_4`。
	OK bool
}

// ShrineStage 是冥想的步驟。
type ShrineStage int

const (
	// ShrineAskVirtue 在問「汝欲冥想何種美德?」
	ShrineAskVirtue ShrineStage = iota
	// ShrineAskMantra 在問真言。
	ShrineAskMantra
	// ShrineAskOffer 在問要獻多少金。
	ShrineAskOffer
)

// ShrineHere 回報玩家腳下是不是聖壇。
//
// ⚠ 原版**沒有**「這一格是不是聖壇」的檢查 —— 它是靠地圖上的
// 聖壇 tile 觸發,進來之後才反查是哪一座。這裡也一樣:
// 由 `Meditate` 的呼叫端決定何時觸發,`ShrineAt` 只負責認出是哪一座。
func (s *State) ShrineHere() int { return u5data.ShrineAt(s.X, s.Y) }

// Meditate 開始冥想。
func (s *State) Meditate() bool {
	if s.InScene() || s.InCombat() || s.InDungeon() {
		s.Log("此處無壇可拜。")
		return false
	}
	v := s.ShrineHere()
	s.Shrine = &Shrine{Virtue: v, OK: true, Stage: ShrineAskVirtue}
	s.Prompt = PromptShrine
	s.logMisc(u5data.MsgShrineApproach)
	s.logMisc(u5data.MsgShrineKneel)
	s.logMisc(u5data.MsgShrineWhich)
	return true
}

// ShrineAnswer 收玩家打的一行字,推進冥想。回傳還在不在冥想中。
func (s *State) ShrineAnswer(text string) bool {
	sh := s.Shrine
	if sh == nil {
		return false
	}
	text = strings.TrimSpace(text)
	switch sh.Stage {
	case ShrineAskVirtue:
		if text == "" {
			s.EndMeditate() // 空字串 = 作罷(原版 `cmp byte_55FDC, 0; jz` 直接離開)
			return false
		}
		// ⚠ 比的是**前四個字母**,而且大小寫不分。
		// `hone`(誠實)與 `hono`(榮譽)只差第四個字母 —— 少比一個字母
		// 兩座聖壇就分不出來。
		if !strings.Contains(strings.ToLower(text), u5data.Shrines[sh.Virtue].Prefix) {
			sh.OK = false
		}
		sh.Stage = ShrineAskMantra
		s.Log("真言(Mantra):")
		return true

	case ShrineAskMantra:
		if text == "" {
			s.EndMeditate()
			return false
		}
		if !strings.EqualFold(text, u5data.Shrines[sh.Virtue].Mantra) {
			sh.OK = false
		}
		sh.Tries++
		if sh.Tries < ShrineMantraTries {
			s.Log("真言(Mantra):")
			return true
		}
		return s.shrineResolve()

	case ShrineAskOffer:
		return s.shrineOffer(text)
	}
	return true
}

// shrineResolve 是三次真言問完之後的判定。
func (s *State) shrineResolve() bool {
	sh := s.Shrine
	if !sh.OK {
		s.logMisc(u5data.MsgShrineUnfocus)
		s.EndMeditate()
		return false
	}
	bit := byte(1) << uint(sh.Virtue)
	switch {
	case s.ShrineQuestGiven&bit == 0:
		// 第一次來:領試煉。
		//
		// ⚠ **這裡與原版有一處簡化。** 原版只設「試煉進行中」
		//(`byte_3E0DC`),「領過試煉」(`byte_3E0DE`)是**讀寶典時**才設的 ——
		// 也就是「去寶典」是必經的一步。寶典那一段還沒逆,所以引擎把兩個
		// 位元一起設,效果變成「拜兩次就拿得到獎」。
		// 寶典做出來之後要把這一行拿掉(`docs/re/25` §2.2)。
		s.ShrineQuestGiven |= bit
		s.ShrineQuestActive |= bit
		s.logMisc(u5data.MsgShrineQuestOn)
		s.Log(i18n.Text("MISCMSG.DAT", u5data.MsgShrineQuestIs, "") +
			i18n.Text("MISCMSG.DAT", u5data.ShrineQuestRecord(sh.Virtue), ""))
		s.logMisc(u5data.MsgShrineReturn)
		s.EndMeditate()
		return false

	case s.ShrineQuestActive&bit != 0:
		// 試煉完成回來:領獎。
		s.ShrineQuestActive &^= bit
		s.shrineReward(sh.Virtue)
		s.EndMeditate()
		return false

	default:
		// 已經領過獎,只剩捐錢。
		sh.Stage = ShrineAskOffer
		s.logMisc(u5data.MsgShrineOffer)
		return true
	}
}

// shrineReward 是完成試煉的獎賞。
func (s *State) shrineReward(virtue int) {
	s.AddKarma(u5data.ShrineQuestKarma)
	sh := u5data.Shrines[virtue]
	ch := &s.Roster[0]
	if sh.Str {
		ch.Strength = byte(addCap(int(ch.Strength), 1, u5data.StatMax))
		ch.Raw[u5data.CharStrength] = ch.Strength
		s.Log("力量 +1")
	}
	if sh.Dex {
		ch.Dex = byte(addCap(int(ch.Dex), 1, u5data.StatMax))
		ch.Raw[u5data.CharDex] = ch.Dex
		s.Log("敏捷 +1")
	}
	if sh.Int {
		ch.Intel = byte(addCap(int(ch.Intel), 1, u5data.StatMax))
		ch.Raw[u5data.CharIntel] = ch.Intel
		s.Log("智力 +1")
	}
	// ⚠ 謙遜是八德裡唯一三圍都不加的一座 —— 原版用雙倍業報補回來。
	if virtue == u5data.VirtueHumility {
		s.AddKarma(u5data.ShrineHumilityBonus)
	}
}

// shrineOffer 處理獻金。
//
// ⚠ 原版只讀**一個** '0'..'9' 的按鍵,所以一次最多獻 9 重 = 900 金。
// 讀成多位數會讓玩家一口氣把業報衝滿。
func (s *State) shrineOffer(text string) bool {
	if text == "" {
		s.EndMeditate()
		return false
	}
	c := text[0]
	if c < '0' || c > '9' {
		s.logMisc(u5data.MsgShrineOffer) // 不是數字就再問一次(原版的忙等迴圈)
		return true
	}
	n := int(c - '0')
	if n == 0 {
		s.Log("0 金。")
		s.EndMeditate()
		return false
	}
	gold := n * u5data.ShrineGoldPerUnit
	if s.Inventory.Gold < gold {
		s.logMisc(u5data.MsgShrineNoGold)
		s.logMisc(u5data.MsgShrineOffer)
		return true
	}
	s.Inventory.Gold -= gold
	s.AddKarma(n)
	s.Log("ALAKAZAM!")
	s.EndMeditate()
	return false
}

// EndMeditate 收掉冥想。
func (s *State) EndMeditate() {
	s.Shrine = nil
	if s.Prompt == PromptShrine {
		s.Prompt = PromptNone
	}
}

// AddKarma 加業報,上限 99(原版 `cmp byte_3E098, 63h`)。
//
// ⚠ 上限是 **99 不是 100**。差一點看起來沒差,但業報 99 是很多判定的門檻。
func (s *State) AddKarma(n int) {
	s.Karma = addCap(s.Karma, n, u5data.KarmaMax)
}

// logMisc 印一筆 `MISCMSG.DAT` 的訊息(已經是譯文)。
func (s *State) logMisc(record int) {
	en := ""
	if s.Misc != nil && record < len(s.Misc.Records) {
		en = s.Misc.Records[record].Text()
	}
	txt := strings.TrimSpace(i18n.Text("MISCMSG.DAT", record, en))
	if txt != "" {
		s.Log(txt)
	}
}

// SubmitShrine 把打好的一行送進冥想流程,並清空輸入框。
func (s *State) SubmitShrine() {
	text := s.Input
	s.Input = ""
	s.ShrineAnswer(text)
}

// OnShrineTile 回報玩家是不是站在**座標表上真的有的**那七座聖壇上。
//
// ⚠ 與 `ShrineHere` 不同:那一支永遠會回一個美德(掃不到就回靈性),
// 是原版 `sub_1D394` 進來之後才用的。這一支是**入口判斷**,
// 不能用那個 fallback —— 否則玩家在任何一格按 E 都會開始冥想。
//
// 靈性聖壇在幽冥界,座標不在表上;它要靠另一條路觸發,還沒接。
func (s *State) OnShrineTile() bool {
	if s.InScene() || s.InDungeon() || s.InCombat() {
		return false
	}
	for i := range u5data.Shrines {
		sh := &u5data.Shrines[i]
		if sh.X == 0 && sh.Y == 0 {
			continue // 靈性:不在表上
		}
		if sh.X == s.X && sh.Y == s.Y {
			return true
		}
	}
	return false
}
