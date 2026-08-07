package game

import (
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 彈豎琴(原版 `sub_11E0`)
//
// 站在豎琴前面按數字鍵發音;在不列顛王城堡二樓把那首十三個音的曲子彈對,
// 牆上會開一道暗門。

// AtHarp 回報玩家**正南**那一格是不是豎琴。
//
// ⚠ 只看南邊一格 —— 原版讀的是 `byte_3F7A9`,也就是視窗緩衝裡玩家那一格
//(`byte_3F789`)往南一列。站在豎琴的別的方向按數字鍵不會發音。
func (s *State) AtHarp() bool {
	return s.TileAt(s.X, s.Y+1) == u5data.HarpTile
}

// PlayNote 彈一個音(key 是 '0'..'9')。回傳有沒有真的彈出聲音。
//
// 曲子彈對之後會在不列顛王城堡二樓切換那一格暗門。
func (s *State) PlayNote(key rune) bool {
	if key < '0' || key > '9' {
		return false
	}
	if !s.AtHarp() {
		return false
	}
	note := int(key - '0')
	s.Log(harpNoteName(note))

	s.HarpProgress = u5data.HarpNext(s.HarpProgress, note)
	if s.HarpProgress < len(u5data.HarpTune) {
		return true
	}
	// 十三個音全對了。進度先歸零 —— 原版無論在哪裡都會歸零,
	// 不是「只有在對的地方才算」。
	s.HarpProgress = 0
	if s.Location != u5data.HarpDoorLocation || s.Floor != u5data.HarpDoorFloor {
		return true
	}
	t := s.TileAt(u5data.HarpDoorX, u5data.HarpDoorY)
	if s.SetTileAt(u5data.HarpDoorX, u5data.HarpDoorY, t^u5data.HarpDoorXor) {
		s.Log(MsgSecretDoor)
	}
	return true
}

// harpNoteName 是訊息欄上顯示的音。
//
// 原版只發聲不印字(`sub_33D78(音高, 0x78, 0xC350)`);這裡多印一行,
// 因為引擎還沒接音訊,不印的話玩家完全看不出自己彈了什麼。
// 音高照原版算,所以之後接上音訊時數字是對的。
func harpNoteName(note int) string {
	semitone := u5data.HarpScale[note] + u5data.HarpMiddleC
	names := [12]string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}
	return "♪ " + names[semitone%12]
}
