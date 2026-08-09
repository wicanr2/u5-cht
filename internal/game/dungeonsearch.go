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
// ✅ **擲骰那一層已補上**(2026-08-08):門檻是 `(樓層 × 2 + 30 − 敏捷) / 2`,
// 擲 `random(1,30)`,**大於**門檻才看得清楚;沒看清楚就用 `random(1,8)` 亂猜。
// 三種要擲的地形:寶箱(查陷阱)、炸彈坑、以及暗門那條的解除。
//
// ★ 陷阱的「複雜度」就是**樓層數** —— 不是每個箱子自己的等級。

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
// 的 13 路比較各自 push。空字串代表那一種**另有邏輯**(寶箱、坑、力場、暗門)。
var dungeonSearchText = [16]string{
	0x0: "此處無異狀。",
	0x1: "梯子上沒有藏東西。",
	0x2: "梯子上沒有藏東西。",
	0x3: "梯子上沒有藏東西。",
	0x4: "", // 寶箱 —— 要擲骰查陷阱
	0x5: "泉水裡沒有藏東西。",
	0x6: "", // 坑與炸彈 —— 依低位元分三種
	0x7: "門上沒有藏東西。",
	0x8: "", // 力場 —— 依低位元分四種
	0x9: "這一格不該存在。", // 原文 "This tile is impossible."
	0xA: "此處是一間房。",
	0xB: "牆上沒有藏東西。",
	0xC: "崩落的通道裡什麼也沒有。",
	0xD: "", // 暗門
	0xE: "門上沒有藏東西。",
	0xF: "門上沒有藏東西。",
}

// 搜尋的擲骰(原版 `sub_142EC`)。
//
// 門檻是 `(樓層 × 2 + 30 − 敏捷) / 2`:**越深越難、敏捷越高越容易**。
// 擲 `random(1,30)`,**大於**門檻才算看清楚;沒看清楚就用 `random(1,8)` 亂猜。
const (
	dungeonSearchRollMax  = 30
	dungeonSearchBase     = 30
	dungeonSearchGuessMax = 8
	// 陷阱複雜度的兩道界線(原版 `cmp var_10, 4` / `cmp var_10, 7`)。
	dungeonTrapSimpleBelow  = 4
	dungeonTrapComplexAtOr  = 7
	dungeonFieldSleep       = 0x80
	dungeonFieldPoison      = 0x81
	dungeonFieldFire        = 0x82
	dungeonPitEmpty         = 0x60
	dungeonPitOpen          = 0x61
	dungeonPitBomb          = 0x62
	dungeonChestNoTrap      = 0x40
)

// dungeonSearchThreshold 是這一次搜尋的難度門檻。
func (s *State) dungeonSearchThreshold(member int) int {
	dex := 0
	if member >= 0 && member < len(s.Roster) {
		dex = int(s.Roster[member].Dex)
	}
	level := 0
	if s.Dungeon != nil {
		level = s.Dungeon.Level
	}
	return (level*2 + dungeonSearchBase - dex) / 2
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
	kind := u5data.DungeonKind(tile)
	switch kind {
	case DungeonSecretDoor:
		s.Log("一道暗門!")
		s.Dungeons.Set(d.Index, d.Level, x, y, dungeonSecretDoorOpens(tile))
		return
	case u5data.DungeonChest:
		s.searchDungeonChest(tile)
		return
	case u5data.DungeonTrap:
		s.searchDungeonPit(tile, x, y)
		return
	case u5data.DungeonMagic:
		s.searchDungeonField(tile)
		return
	}
	if txt := dungeonSearchText[kind>>4]; txt != "" {
		s.Log(txt)
	}
}

// searchDungeonChest 是寶箱的陷阱偵測(原版 `loc_144A8`)。
//
// ★ **陷阱的「複雜度」就是樓層數。** 原版查的不是格子裡的什麼欄位,
// 而是 `byte_3E0A5` —— 越深的陷阱越複雜。憑「每個箱子有自己的陷阱等級」
// 去實作會多出一個原版沒有的維度。
//
// 沒看清楚時用 `random(1,8)` 亂猜,所以**會報錯三種答案中的任一種** ——
// 與地表的搜尋同一個形狀(`docs/re/43`)。
func (s *State) searchDungeonChest(tile byte) {
	// ⚠ **不檢查 member < 0** —— 原版拿 `sub_E19C` 的回傳值直接去查敏捷,
	// 沒人可選時算出來的門檻就是「敏捷 0」。照抄(`dungeonSearchThreshold`
	// 對 −1 已經回 dex 0)。
	s.pickMember("", func(member int) { s.searchDungeonChestBy(tile, member) })
}

// searchDungeonChestBy 是由 member 去查箱子的陷阱。
func (s *State) searchDungeonChestBy(tile byte, member int) {
	level := 0
	if s.Dungeon != nil {
		level = s.Dungeon.Level
	}
	v := level
	if s.Roll(1, dungeonSearchRollMax) <= s.dungeonSearchThreshold(member) {
		v = s.Roll(1, dungeonSearchGuessMax) // 沒看清楚 → 亂猜
	} else if tile == dungeonChestNoTrap {
		s.Log("沒有陷阱。")
		return
	}
	switch {
	case v < dungeonTrapSimpleBelow:
		s.Log("一個簡單的機關。")
	case v >= dungeonTrapComplexAtOr:
		s.Log("一個複雜的機關。")
	default:
		s.Log("有機關。")
	}
}

// searchDungeonPit 是坑與炸彈(原版 `loc_14517`)。
//
// 三種低位元各自不同,而且**搜到就解除** —— 原版把格子改成
// `(原值 & 8) + 0x60`,保留「頭上有洞」那一位元(同暗門那條)。
func (s *State) searchDungeonPit(tile byte, x, y int) {
	d := s.Dungeon
	disarm := func() {
		s.Dungeons.Set(d.Index, d.Level, x, y,
			u5data.DungeonTrap|(tile&u5data.DungeonHoleAbove))
	}
	switch tile {
	case dungeonPitEmpty:
		s.Log("坑裡沒有藏東西。")
	case dungeonPitOpen:
		s.Log("一個陷坑!")
		disarm()
	case dungeonPitBomb:
		// ⚠ 同 `searchDungeonChest`:原版不檢查有沒有人可選。
		s.pickMember("", func(member int) {
			if s.Roll(1, dungeonSearchRollMax) <= s.dungeonSearchThreshold(member) {
				s.Log("坑裡沒有藏東西。")
				return
			}
			s.Log("一個炸彈陷阱!")
			disarm()
		})
	}
}

// searchDungeonField 是三種力場(原版 `jpt_14627`)。
//
// ⚠ 0x83(電擊力場)**刻意沒有訊息** —— 跳表把它送進 default。
// 同 `dungeon.go` 裡踩踏分派的那條:電擊力場在原版是「走進去被彈回來」,
// 沒有人站上去過,也沒有人搜到它。
func (s *State) searchDungeonField(tile byte) {
	switch tile {
	case dungeonFieldSleep:
		s.Log("一道睡眠力場。")
	case dungeonFieldPoison:
		s.Log("一道毒氣力場。")
	case dungeonFieldFire:
		s.Log("一道火牆。")
	}
}
