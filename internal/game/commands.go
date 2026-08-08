package game

import (
	"fmt"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 四個指令:Jimmy(J)、New order(N)、View a gem(V)、Ztats(Z)
//
// 原版的指令派發在 `sub_2ACF4`,A..Z 各一格跳表。這四支的處理函式分別是
// `sub_14CAC` / `sub_17688` / `sub_EDD4` / `sub_1E9A0`。

// ---------------------------------------------------------------- Jimmy

// Jimmy 是 J 指令:用鑰匙撬鎖(原版 `sub_14CAC`)。
//
// # 規則
//
//	沒有鑰匙                        → 「沒有鑰匙!」,不問方向
//	目標是 0xB9 / 0xBB(普通鎖門)  → 擲 random(0, 29) 對上該員的**敏捷**
//	                                  擲值 < 敏捷 → 開了;否則「鑰匙斷了!」
//	目標是 0x97 / 0x98(魔法鎖)    → **一律「鑰匙斷了!」**
//	其餘                            → 撬不了
//
// ⚠ 魔法鎖那條是「必定失敗**而且照樣扣鑰匙**」——原版 `loc_14DC0` 直接跳到
// 扣鑰匙那段。寫成「魔法鎖不能撬,什麼都不發生」會讓玩家可以無限試,
// 而原版是會把鑰匙耗光的。
const jimmyRollMax = 29

// Jimmy 是 J 指令。
//
// ⚠ **地牢裡是另一支程式**(原版 `sub_14CAC` 開頭就分岔到 `sub_14B2C`)——
// 引擎原本只有門那條,而 `TileAt` 在地牢裡讀不到地牢格 ⇒ **地牢裡按 J 撬不了
// 任何東西**。見 `docs/re/76`。
func (s *State) Jimmy() {
	if s.InDungeon() {
		s.jimmyDungeonChest()
		return
	}
	if s.Inventory.Keys <= 0 {
		s.Log(MsgNoKeys)
		return
	}
	s.AskDirection(func(d Direction) {
		dx, dy := d.Delta()
		s.jimmyAt(s.X+dx, s.Y+dy)
	})
}

// jimmyDungeonChest 是地牢裡的 J(原版 `sub_14B2C`)。
//
// ★★ **它撬的不是鎖,是陷阱。** 成功之後那一格是
// `(tile & 8) | 0x40` —— 還是箱子(0x4x),只是**低三位元的陷阱被清掉了**。
// 所以 J 在地牢裡的用途是「先解陷阱,再用 O 開」。
//
// ★ 三個容易寫錯的地方:
//
//  1. **沒有陷阱的箱子撬不開,而且鑰匙照斷。**
//     判準是 `tile & 0xF7 == 0x40`(也就是 tile ∈ {0x40, 0x48})——
//     那是低三位元為 0 的箱子,沒有東西可解,所以原版直接跳到「鑰匙斷了」。
//     寫成「沒陷阱就成功」會讓玩家把鑰匙當萬能鑰匙。
//  2. **問人在查鑰匙之前**。地牢這條的順序是「選人 → 讀格子 → 查鑰匙」,
//     所以身上沒鑰匙時原版**照樣先問「Player:」**(門那條相反,先查鑰匙)。
//  3. 門檻與地牢搜尋**同一條式子**(`(樓層×2 + 30 − 敏捷) / 2`),
//     擲的是 `random(1, 30)` 且要**大於**門檻才成功。
func (s *State) jimmyDungeonChest() {
	d := s.Dungeon
	// ★ 先問人 —— 原版在這裡,而不是在查鑰匙之後。
	who := s.pickCharacter("")
	if who < 0 {
		return
	}
	tile := s.DungeonTileHere()
	switch {
	case tile&^u5data.DungeonHoleAbove == u5data.DungeonChest:
		// 沒有陷阱的箱子:撬不開,鑰匙照斷。
		if s.Inventory.Keys <= 0 {
			s.Log(MsgNoKeys)
			return
		}
		s.Log(MsgKeyBroke)
		s.Inventory.Keys--
	case u5data.DungeonKind(tile) == u5data.DungeonChest:
		// 有陷阱的箱子:擲骰解陷阱。
		if s.Inventory.Keys <= 0 {
			s.Log(MsgNoKeys)
			return
		}
		if s.Roll(1, dungeonSearchRollMax) > s.dungeonSearchThreshold(who) {
			s.Log(MsgChestUnlocked)
			s.Dungeons.Set(d.Index, d.Level, d.X, d.Y,
				(tile&u5data.DungeonHoleAbove)|u5data.DungeonChest)
			return
		}
		s.Log(MsgKeyBroke)
		s.Inventory.Keys--
	case u5data.DungeonKind(tile) == u5data.DungeonOpenedChestKind:
		s.Log(MsgAlreadyOpen)
	default:
		s.Log(MsgWhat)
	}
}

// jimmyAt 撬 (x, y) 那一格的鎖。
func (s *State) jimmyAt(x, y int) {
	tile := s.TileAt(x, y)
	switch tile {
	case u5data.TileLockedDoor, u5data.TileLockedMagicDoor:
		// 普通鎖:看撬鎖的人手多穩。
		i := s.pickCharacter("")
		if i < 0 {
			return
		}
		s.Inventory.Keys--
		if s.Roll(0, jimmyRollMax) < int(s.Roster[i].Dex) {
			s.SetTileAt(x, y, u5data.TileDoorA)
			s.Log(MsgUnlocked)
			return
		}
		s.Log(MsgKeyBroke)
	case u5data.TileMagicLockedA, u5data.TileMagicLockedB:
		// 魔法鎖:必定失敗,鑰匙照斷。
		s.Inventory.Keys--
		s.Log(MsgKeyBroke)
	default:
		s.Log(MsgNoLockHere)
	}
}

// ---------------------------------------------------------------- New order

// NewOrder 是 N 指令:交換兩名隊員的位置(原版 `sub_17688`)。
//
// ⚠ **聖者不能離開第一位。** 原版對 index 0 印「<名字> must lead!」——
// 隊伍第 0 格是聖者,整個存檔格式與對話系統都假設它在那裡。
// 少了這道檢查,玩家可以把聖者換到後面,然後對話與結局判定全部找錯人。
//
// 交換的是**整筆 32 B 記錄**,不是只換名字。
func (s *State) NewOrder(a, b int) bool {
	if a < 0 || b < 0 || a >= s.PartySize || b >= s.PartySize ||
		a >= len(s.Roster) || b >= len(s.Roster) {
		s.Log(MsgSwapNobody)
		return false
	}
	if a == 0 || b == 0 {
		s.Log(s.Roster[0].Name + MsgMustLead)
		return false
	}
	if a == b {
		return false
	}
	s.Roster[a], s.Roster[b] = s.Roster[b], s.Roster[a]
	s.Log(fmt.Sprintf(MsgSwapped, s.Roster[b].Name, s.Roster[a].Name))
	return true
}

// ---------------------------------------------------------------- View a gem

// ViewGem 是 V 指令:看一顆寶石,攤開周圍的全景(原版 `sub_EDD4`)。
//
// ★ **與 In Quas Wis 走同一支函式。** 咒語版與寶石版在原版是同一個畫面,
// 差別只在「寶石版要有寶石可用」。所以這裡直接呼叫 `Peer()`,
// 不另外寫一套 —— 兩套會漂掉,而畫面漂掉是看得見的。
//
// ⚠ 原版**不扣寶石**:`sub_2B115` 只檢查「有沒有」(`aYouHaveNone`
// = "You have none!"),看完寶石還在。這很反直覺,所以特別註明 ——
// 不要「順手」加一行扣除。
func (s *State) ViewGem() bool {
	if s.Inventory.Gems <= 0 {
		s.Log(MsgYouHaveNone)
		return false
	}
	return s.Peer()
}

// ---------------------------------------------------------------- Ztats

// Ztats 是 Z 指令:角色數值畫面(原版 `sub_1E9A0`,`ZSTATS.OVL`)。
//
// 資料全部來自存檔的角色記錄,沒有新的解碼工作 —— 缺的一直是畫面。
// 這裡先做**文字版**:一名隊員一頁,左右翻頁。
//
// ⚠ 職業與狀態用**中文顯示、英文 canonical**:狀態碼是 `'G'`/`'P'`/`'D'`/
// `'S'`/`'C'`,而治療所與復活判定比的是那個位元組。顯示層譯了不影響比對。
type Ztats struct {
	// Page 是原版那**一個**游標(`sub_1E9A0` 的 `esi`,0..16)。
	//
	// ★ 不拆成「哪個人 + 哪一頁」兩個變數是刻意的:翻頁時
	// 「最後一名的裝備頁 → Equipment」與「Armaments → 繞回第一名」
	// 這兩個接縫用兩個變數寫不出來(`internal/game/ztatspages.go`)。
	Page int
}

// BeginZtats 打開數值畫面。
func (s *State) BeginZtats() bool {
	if s.PartySize <= 0 || len(s.Roster) == 0 {
		return false
	}
	s.Zstats = &Ztats{}
	s.Prompt = PromptZtats
	return true
}

// ZtatsPage 翻頁。delta 是 -1 或 +1。
//
// ⚠ **不是**「在隊伍範圍內繞回」—— 原版有 17 頁(六名 × 2 + 五頁全隊共用),
// 而繞回的接縫在 Armaments 與第一名之間(`ZtatsNext` / `ZtatsPrev`)。
func (s *State) ZtatsPage(delta int) {
	if s.Zstats == nil {
		return
	}
	n := s.PartySize
	if n > len(s.Roster) {
		n = len(s.Roster)
	}
	if n <= 0 {
		return
	}
	if delta < 0 {
		s.Zstats.ZtatsPrev(n)
		return
	}
	s.Zstats.ZtatsNext(n)
}

// ZtatsKey 收數字鍵(原版按鍵 `0` 與 `1`..`6`)。
//
//	'0'      → Equipment 頁
//	'1'..'6' → 跳到第 N 名的數值頁
//
// ⚠ 超出隊伍人數的按鍵**什麼都不做** —— 原版先擋
// `鍵碼 − 0x31 >= 隊伍人數`,不是繞回也不是報錯。回傳有沒有吃掉這個鍵。
func (s *State) ZtatsKey(r rune) bool {
	if s.Zstats == nil {
		return false
	}
	if r == '0' {
		s.Zstats.ZtatsJumpEquipment()
		return true
	}
	if r < '1' || r > '6' {
		return false
	}
	n := s.PartySize
	if n > len(s.Roster) {
		n = len(s.Roster)
	}
	before := s.Zstats.Page
	s.Zstats.ZtatsJumpMember(int(r-'1'), n)
	return s.Zstats.Page != before || int(r-'1') < n
}

// EndZtats 關掉數值畫面。
func (s *State) EndZtats() {
	s.Zstats = nil
	s.Prompt = PromptNone
}

// ZtatsLines 是數值畫面現在該顯示的每一行(標題 + 內容)。
//
// 頁面模型與逐頁排版在 `internal/game/ztatspages.go`(原版 `sub_1E9A0` 一族)。
func (s *State) ZtatsLines() []string {
	if s.Zstats == nil {
		return nil
	}
	title := s.ZtatsPageTitle()
	body := s.ZtatsBody()
	if title == "" && body == nil {
		return nil
	}
	out := []string{title}
	out = append(out, body...)
	// 聖者那一行只在隊員頁出現(原版 `sub_1E0A4` 畫在框裡,不分頁 ——
	// 但它讀的是**那一頁的角色**,所以共用頁沒有它)。
	if s.Zstats.Page < ZtatsEquipmentPage {
		if m := s.Zstats.Page / ZtatsMemberPages; m < len(s.Roster) {
			out = append(out, s.avatarLine(&s.Roster[m]))
		}
	}
	return out
}

// avatarLine 是 Ztats 最後那一句「某某**是**聖者 / **不是**聖者」。
//
// 原版 `sub_7594`(Ztats 畫面)在名字後面接 " is " 再依 `dword_54498` 選
// `an Avatar.` 或 `not an Avatar`。
//
// ★ 那個旗標**只有一個來源**:從 Ultima IV 轉入角色時(`sub_71D0`),
// 若 U4 存檔裡那八個 word(`[+6]`..`[+0x14]`)全為 0 就設 1。
// 沒有轉入過就一直是 0 —— 所以**新建的角色一律「不是聖者」**,
// 這不是待補的預設值,是原版的行為。
//
// ⚠ 因此這一句不能寫成「主角一定是聖者」。U5 的開場正是「你回來了,
// 但這一次不是以聖者的身分」——那句話有它的意義。
func (s *State) avatarLine(c *u5data.Character) string {
	if s.TransferredAvatar {
		return c.Name + "乃聖者。"
	}
	return c.Name + "並非聖者。"
}
