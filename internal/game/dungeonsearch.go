package game

import "github.com/wicanr2/u5-cht/internal/u5data"

// 地牢裡的 Search(原版 `sub_142EC`,由 `sub_147A8` 依所在地分派)
//
// ★ 地牢的搜尋是**另一支程式**,不是地面那支加個條件:
//
//	sub_147A8:  if (0x20 < 地點 < 0x29) → sub_142EC()   ← 地牢
//	                                 else → 問絕對方向,搜鄰格
//
// 兩者差在三處:
//
//  1. **方向是相對的。** 原版印「Dir-」然後收方向鍵,但意義是
//     前(Ahead)/ 左(Left)/ 右(Right)/ **腳下(Here)**,空白鍵 = 放棄。
//     第一人稱下「北」沒有意義,所以是相對的。
//  2. **看不見就搜不到。** 沒有光源時直接回「汝所見:一片漆黑。」
//  3. **報的是地形本身**,而不是地面那套「家具的地點語 + 翻到什麼」。
//     每一種地形有自己的一句話,而 **0xD0 是暗門** —— 搜到就變成門。
//
// ⚠ 原版還有一道**擲骰**決定要不要多找到東西(寶箱的陷阱、坑裡的東西、
// 骸骨崩解…),門檻是 `(樓層 × 2 + 30 − 敏捷) / 2`。骰子的上下界還沒讀出來,
// 所以**那一層先不做**並標在這裡 —— 少一個獎勵好過憑感覺補一個機率。

// SearchRelative 是地牢搜尋的相對方向。
type SearchRelative int

// 四個方向 + 放棄(原版 `sub_13BA8` 收的鍵碼 1..4 與空白)。
const (
	SearchLeft  SearchRelative = 1
	SearchRight SearchRelative = 2
	SearchAhead SearchRelative = 3
	SearchHere  SearchRelative = 4
)

// dungeonSearchText 是每一種地牢地形被搜過之後的那一句話。
//
// 索引 = 高四位元 >> 4。原文在 `WORRIORS.EXP` 的字串池裡,由 `sub_142EC`
// 的 13 路比較各自 push。
var dungeonSearchText = [16]string{
	0x0: "此處無異狀。",
	0x1: "梯子上沒有藏東西。",
	0x2: "梯子上沒有藏東西。",
	0x3: "梯子上沒有藏東西。",
	0x4: "箱子裡是財寶!",
	0x5: "泉水裡沒有藏東西。",
	0x6: "坑裡沒有藏東西。",
	0x7: "門上沒有藏東西。",
	0x8: "此處有一道力場。",
	0x9: "這一格不該存在。", // 原文 "This tile is impossible."
	0xA: "此處是一間房。",
	0xB: "牆上沒有藏東西。",
	0xC: "崩落的通道裡什麼也沒有。",
	0xD: "", // 暗門 —— 另外處理
	0xE: "門上沒有藏東西。",
	0xF: "門上沒有藏東西。",
}

// DungeonSecretDoor 是「搜得出來的暗門」那一種地形(原版 `cmp edx, 0D0h`)。
const DungeonSecretDoor = 0xD0

// dungeonSecretDoorOpens 是暗門被搜到之後變成什麼(原版 `and al, 8; add al, 0E0h`)。
//
// ⚠ **保留第 3 位元**(「頭上有洞」那一位)—— 直接寫 0xE0 會把洞抹掉,
// 之後就爬不回上一層了。原版是 `0xE0 | (原值 & 8)`,不是 `0xE0`。
func dungeonSecretDoorOpens(tile byte) byte {
	return u5data.DungeonDoorway | (tile & u5data.DungeonHoleAbove)
}

// searchDungeon 是地牢的 S 指令:問相對方向,再看那一格。
func (s *State) searchDungeon() {
	s.Log("搜尋……")
	s.AskDirection(func(d Direction) {
		// 絕對方向鍵在第一人稱下要換算成相對:↑ 前、← 左、→ 右、↓ 腳下。
		switch d {
		case North:
			s.searchDungeonRelative(SearchAhead)
		case West:
			s.searchDungeonRelative(SearchLeft)
		case East:
			s.searchDungeonRelative(SearchRight)
		default:
			s.searchDungeonRelative(SearchHere)
		}
	})
}

// searchDungeonRelative 搜相對方向的那一格。
func (s *State) searchDungeonRelative(rel SearchRelative) {
	d := s.Dungeon
	if d == nil {
		return
	}
	// 沒有光就什麼都搜不到。
	//
	// ⚠ 原版比的是**兩個光源計時器**(`byte_3E0B6`/`byte_3E0B7`,即火把與
	// In Lor 的剩餘回合)都為 0,**不是**視野半徑 —— 地牢的基礎半徑是夜間的 2,
	// 永遠大於 0,拿半徑判會永遠有光。
	if s.TorchTurns <= 0 && s.LightTurns <= 0 {
		s.Log("汝所見:一片漆黑。")
		return
	}
	x, y := d.X, d.Y
	switch rel {
	case SearchAhead:
		dx, dy := d.Facing.Delta()
		x, y = x+dx, y+dy
	case SearchLeft:
		dx, dy := d.Facing.TurnLeft().Delta()
		x, y = x+dx, y+dy
	case SearchRight:
		dx, dy := d.Facing.TurnRight().Delta()
		x, y = x+dx, y+dy
	}
	x, y = u5data.DungeonWrap(x), u5data.DungeonWrap(y)
	tile := s.DungeonTileAt(x, y)
	s.Log("汝所見:")
	if u5data.DungeonKind(tile) == DungeonSecretDoor {
		s.Log("一道暗門!")
		s.Dungeons.Set(d.Index, d.Level, x, y, dungeonSecretDoorOpens(tile))
		return
	}
	s.Log(dungeonSearchText[u5data.DungeonKind(tile)>>4])
}
