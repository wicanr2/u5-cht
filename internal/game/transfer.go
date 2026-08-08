package game

import (
	"fmt"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 從 Ultima IV 轉入角色 —— 主選單的第三項(原版 `sub_7594`)
//
// 這是系列特色:U4 打完的那個角色可以帶進 U5。轉進來的東西比想像中少 ——
// 三圍、生命、經驗、名字、性別、職業,加上一個「是不是聖者」的旗標。
// 裝備、金錢、道具**都不帶**(原版那兩塊 read 裡沒有它們)。
//
// ⚠ 原版寫死讀 `a:party.sav`(U4 的存檔磁碟)。現代環境沒有 A 磁碟,
// 所以引擎讓呼叫端給路徑;檔名 `PARTY.SAV` 保留原樣供辨識。
//
// # 三個階段,順序照原版
//
//	1. `sub_71D0` 讀檔 + 驗界線(`u5data.ParseU4Transfer`)
//	2. 印 "Found:" 那一頁,然後問**兩個問題**:改不改名、改不改性別
//	3. 換算(`u5data.Convert`)並逐項報告
//
// ★ 第 2 步的兩個問題是**玩家真的做得到的選擇**,不是裝飾:
//
//	"Keep this name?"  → N 就問 "Enter new name: "(最多 8 個字元;
//	                      **什麼都不打就保留原名**)
//	"Keep same sex?"   → Y 保留、**N 翻轉**
//
// 兩個問題都**只收 Y / N**,其他鍵原版是繼續等(`cmp bl,'Y'; jz; cmp bl,'N'; jnz 重讀`)。
//
// 性別的那一段值得看一眼,因為它不是「N 就設成女」:
//
//	if (Y 且目前是男 0x0B) → 印 Male
//	if (N 且目前是女 0x0C) → 印 Male      ← 翻轉
//	否則                    → 印 Female
//
// 也就是 **Y 保留、N 翻轉**,而不是「Y=男 / N=女」。寫成後者的話,
// 一個女角色答 Y 會變成男的。

// U4NewNameMax 是改名時收得下幾個字元(原版 `sub_239B4(byte_3DDB4, 8)`)。
const U4NewNameMax = 8

// pendingTransfer 是「讀好了、還在問問題」的轉入。
type pendingTransfer struct {
	t *u5data.U4Transfer
}

// TransferFromUltimaIV 讀一份 U4 `PARTY.SAV` 並開始轉入流程。
//
// 回傳 true 代表**讀檔成功、流程已經開始**(接著會問兩個問題);
// 失敗時**照原版印同一句話**,不分原因 —— 原版對所有失敗都只印
// `Error: Your Ultima IV game contains …` + `Unable to continue transfer.`。
func (s *State) TransferFromUltimaIV(path string) bool {
	t, err := u5data.LoadU4Transfer(path)
	if err != nil {
		s.Log(MsgTransferError)
		s.Log(MsgTransferUnable)
		return false
	}
	s.xfer = &pendingTransfer{t: t}
	// 原版的 "Found:" 那一頁:等級、性別、職業、三圍、名字、是不是聖者。
	c := &t.Char
	s.Log(fmt.Sprintf("%s:%s,%s%s", MsgTransferFound, c.Name,
		s.genderWord(c.Gender), string(c.Class)))
	s.Log(fmt.Sprintf("力 %d 敏 %d 智 %d", c.Strength, c.Dex, c.Intel))
	if t.Avatar {
		s.Log(c.Name + MsgIsAnAvatar)
	} else {
		s.Log(c.Name + MsgIsNotAnAvatar)
	}
	s.askTransferName()
	return true
}

// askTransferName 問「保留這個名字嗎」。
func (s *State) askTransferName() {
	s.Ask(MsgKeepThisName, func(keep bool) {
		if keep {
			s.askTransferSex()
			return
		}
		s.AskText(MsgEnterNewName, U4NewNameMax, func(name string) {
			// ⚠ 什麼都不打就**保留原名** —— 原版 `cmp byte_3DDB4, 0; jz` 那條路。
			if name != "" {
				s.xfer.t.Char.Name = name
				copy(s.xfer.t.Char.Raw[u5data.CharName:u5data.CharName+u5data.CharNameLen],
					make([]byte, u5data.CharNameLen))
				copy(s.xfer.t.Char.Raw[u5data.CharName:], name)
			}
			s.askTransferSex()
		})
	})
}

// askTransferSex 問「保留同樣的性別嗎」——**Y 保留、N 翻轉**。
func (s *State) askTransferSex() {
	s.Ask(MsgKeepSameSex, func(keep bool) {
		c := &s.xfer.t.Char
		if !keep {
			if c.Gender == u5data.GenderMale {
				c.Gender = u5data.GenderFemale
			} else {
				c.Gender = u5data.GenderMale
			}
			c.Raw[u5data.CharGender] = c.Gender
		}
		s.Log(s.genderWord(c.Gender))
		s.finishTransfer()
	})
}

// finishTransfer 跑換算、逐項報告,並把角色放進名冊。
func (s *State) finishTransfer() {
	t := s.xfer.t
	s.xfer = nil
	// ★ 讀進來還沒完 —— 原版 `sub_7594` 接著逐項換算並把變化印出來。
	// 少了這一步,轉入的角色會帶著 U4 的三圍與經驗值直接進 U5,強得離譜。
	before := t.Char
	t.Convert()
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
	c := &t.Char
	s.Log(c.Name + "自《創世紀 IV》而來。")
	// 原版逐項報告換算結果(`Experience has been converted` 等)。
	s.Log(fmt.Sprintf("經驗 %d → %d,等級 %d", before.Exp, c.Exp, c.Level))
	s.Log(fmt.Sprintf("生命 %d(等級 × %d)", c.MaxHP, u5data.U4TransferHPPerLevel))
	s.Log(fmt.Sprintf("力量 %d(50)→ %d(30)", before.Strength, c.Strength))
	s.Log(fmt.Sprintf("敏捷 %d(50)→ %d(30)", before.Dex, c.Dex))
	s.Log(fmt.Sprintf("智力 %d(50)→ %d(30),法力同智力", before.Intel, c.Intel))
	if t.Avatar {
		s.Log(c.Name + "乃聖者。")
	}
}

// genderWord 是性別的顯示字(原版 `Male` / `Female`)。
func (s *State) genderWord(g byte) string {
	if g == u5data.GenderMale {
		return MsgMale
	}
	return MsgFemale
}
