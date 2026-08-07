package game

import (
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 從 Ultima IV 轉入角色 —— 主選單的第三項(原版 `sub_71D0`)
//
// 這是系列特色:U4 打完的那個角色可以帶進 U5。轉進來的東西比想像中少 ——
// 三圍、生命、經驗、名字、性別、職業,加上一個「是不是聖者」的旗標。
// 裝備、金錢、道具**都不帶**(原版那兩塊 read 裡沒有它們)。
//
// ⚠ 原版寫死讀 `a:party.sav`(U4 的存檔磁碟)。現代環境沒有 A 磁碟,
// 所以引擎讓呼叫端給路徑;檔名 `PARTY.SAV` 保留原樣供辨識。

// TransferFromUltimaIV 把一份 U4 `PARTY.SAV` 轉成隊伍的第一名角色。
//
// 成功時回傳 true。失敗時**照原版印同一句話**,不分原因 ——
// 原版對所有失敗都只印 `Error: Your Ultima IV game contains …` +
// `Unable to continue transfer.`,不告訴玩家是哪裡不對。
func (s *State) TransferFromUltimaIV(path string) bool {
	t, err := u5data.LoadU4Transfer(path)
	if err != nil {
		s.Log("錯誤:汝之《創世紀 IV》存檔內容有誤。")
		s.Log("無法完成轉入。")
		return false
	}
	if len(s.Roster) == 0 {
		s.Roster = make([]u5data.Character, 1)
	}
	s.Roster[0] = t.Char
	if s.PartySize < 1 {
		s.PartySize = 1
	}
	// ★ 這個旗標只有這裡設得起來 —— Ztats 末行的「乃聖者 / 並非聖者」
	// 完全由它決定(`docs/re/54` §3)。
	s.TransferredAvatar = t.Avatar
	s.Log(t.Char.Name + "自《創世紀 IV》而來。")
	if t.Avatar {
		s.Log(t.Char.Name + "乃聖者。")
	}
	return true
}
