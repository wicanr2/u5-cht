package game

import (
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 暗影君主的遊走與消滅
//
// 召喚在 `wordofpower.go`(Yell 一個名字),玷污聖壇在 `blackthorn.go`。
// 這一份是另外兩塊:每天午夜換城市,以及在聖火前用寶石碎片了結。

// roamShadowlords 是午夜把活著的暗影君主重新分派(原版 `sub_29304` 的午夜段)。
//
// 原版的迴圈:
//
//	do {
//	    b = random(1, 8);
//	    if (byte_3E0A3 == b) b = 0;              // 玩家就在那座城 → 重抽
//	    for (j = 0; j < 3; j++)
//	        if (byte_3E0D8[j] == b) b = 0;       // 已經有別位在那 → 重抽
//	} while (b == 0);
//	byte_3E0D8[i] = b;
//
// ⚠ 兩個排除條件都不能省:少了「玩家所在」那條,暗影君主會憑空出現在
// 玩家腳下的城裡;少了「不重複」那條,三位會疊在同一座城,
// 而剩下的城永遠遇不到 —— 兩種都會讓整條主線的節奏走樣。
//
// ⚠ 這個迴圈**理論上會空轉**:八座城全被佔滿又剛好都排除時抽不到。
// 實際上三位 + 玩家最多佔 4 座,八座裡永遠有得選,所以原版沒有防護。
// 這裡加了上限只是不讓引擎在被改壞的存檔上卡死,不是行為差異。
func (s *State) roamShadowlords() {
	const maxTries = 1000
	for i := range s.ShadowlordAt {
		// 0x80 以上是「不在城裡」(0xFF = 已被消滅),不參與分派。
		if s.ShadowlordAt[i] >= 0x80 {
			continue
		}
		for try := 0; try < maxTries; try++ {
			b := byte(s.Roll(u5data.ShadowlordCityMin, u5data.ShadowlordCityMax))
			if int(b) == s.Location {
				continue
			}
			if s.shadowlordInCity(b) {
				continue
			}
			s.ShadowlordAt[i] = b
			break
		}
	}
}

// shadowlordInCity 回報這個地點編號是不是已經有暗影君主了。
func (s *State) shadowlordInCity(loc byte) bool {
	for _, v := range s.ShadowlordAt {
		if v == loc {
			return true
		}
	}
	return false
}

// ShadowlordHauntsHere 回報玩家現在所在的地點有沒有暗影君主盤據。
//
// 原版 `sub_48C`:掃三筆,`byte_3E0D8[i] == byte_3E0A3` 就把 `byte_3E16A` 設成 i。
// 回傳 −1 代表沒有。
func (s *State) ShadowlordHauntsHere() int {
	if s.Location == 0 {
		return -1
	}
	for i, v := range s.ShadowlordAt {
		if int(v) == s.Location {
			return i
		}
	}
	return -1
}

// 寶石碎片(原版 `sub_1A38C`)
//
// 舉起碎片 → 站的位置要正好是對應的那團聖火 → 正北一格要站著**那一位**
// 暗影君主 → 才消滅得掉。三個條件缺一不可,而且原版對後兩個條件的失敗
// 反應不同:位置不對會說「毫無效果」,位置對但沒有那位在,則**什麼都不說**。

// UseGemShard 把第 i 塊寶石碎片舉起來。回傳有沒有消滅掉暗影君主。
func (s *State) UseGemShard(i int) bool {
	if i < 0 || i >= u5data.ShadowlordCount {
		return false
	}
	if !s.Shards[i] {
		s.Log("汝沒有那塊碎片。")
		return false
	}
	f := u5data.Flames[i]
	s.Log("寶石碎片 —— 汝在頭頂高舉起" + f.ShardZH + "之碎片……")

	if s.X != f.X || s.Y != f.Y || s.Location != f.Location || s.Floor != f.Floor {
		s.Log(MsgNoEffect)
		return false
	}
	s.Log("……並將它擲入" + f.NameZH + "之火!")

	// 正北一格必須是那一位暗影君主。
	//
	// ⚠ 原版比的是**兩件事**:那一格的 tile 是 0xFC,而且 `byte_3E0DB == i`
	//(現在被召喚出來的就是這一位)。少了後者,任何一位站在火前都能被
	// 任何一塊碎片打掉。
	n := s.northOfPlayer()
	if n < 0 || int(s.ShadowlordHere) != i {
		// ⚠ 原版在這裡**不印任何東西**就 return —— 火燒起來了,但什麼也沒發生。
		return false
	}
	s.removeNPC(n)
	s.ShadowlordAt[i] = u5data.ShadowlordGone
	s.ShadowlordHere = u5data.ShadowlordNone
	s.Shards[i] = false
	s.DoomFlags |= u5data.ShadowlordDoomBit[i]
	s.Log("暗影君主" + u5data.Shadowlords[i] + "的末日已然降臨!")
	return true
}

// northOfPlayer 回傳站在玩家正北一格的暗影君主是第幾號 NPC 槽;沒有就回 −1。
func (s *State) northOfPlayer() int {
	for _, v := range s.VisibleNPCs() {
		if v.X == s.X && v.Y == s.Y-1 && v.NPC.Creature == u5data.TileShadowlord {
			return v.Index
		}
	}
	return -1
}

// removeNPC 把一個 NPC 槽從場上移除。
func (s *State) removeNPC(i int) {
	if s.removed == nil {
		s.removed = map[int]bool{}
	}
	s.removed[s.Location<<8|i] = true
	if i < len(s.rtNPCs) {
		s.rtNPCs[i].Mode = ModeAbsent
	}
}
