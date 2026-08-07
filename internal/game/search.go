package game

import "github.com/wicanr2/u5-cht/internal/u5data"

// Search 指令(原版 `sub_147A8`,兩支結算在 `sub_13F04` / `sub_13DD8`)
//
// 搜一格東西,可能有三種結果:
//
//	那一格有物件(寶箱之類) → 查陷阱(`sub_13F04`)
//	那一格是有缺口的牆      → 找到密門
//	其餘家具                → 擲一次「翻到了什麼」(`sub_13DD8`)
//
// # 敘述句是拼出來的
//
// 原版的字串長這樣:`"\nIn the stump\nt"` + `"hou dost find\n"` ——
// 地點語**以 `t` 結尾**,接上 `hou dost find` 才成句。所以地點語不是
// 「額外的一行」,它是同一句話的前半。中文照這個結構走:
// 「在樹洞裡」+「汝翻到了」。
//
// # 陷阱偵測會出兩種錯(這是特色不是 bug)
//
//	偵測成功 + 真的有陷阱 → 照等級描述
//	偵測成功 + 沒有陷阱   → 「沒有陷阱!」
//	偵測失敗 + 真的有陷阱 → **「沒有陷阱!」** ← 漏看了
//	偵測失敗 + 沒有陷阱   → **「有陷阱!」**   ← 幻覺
//
// 低智力的角色既會漏看真陷阱、也會看到不存在的陷阱。兩種錯都要照做 ——
// 只做「漏看」的話,玩家會學到「說沒陷阱就一定安全」,而原版不是。

// searchPhrase 是每種家具的地點語(原版那一串以 `t` 結尾的字串)。
//
// 每一個 tile 都拿 `LOOK2.DAT` 的敘述交叉對過(`docs/re/43`)——
// 0xAD 是「chest of drawers」對上「in the dresser」、0xAF 是
// 「heavy footlocker」對上「in the trunk」,不是憑順序猜的。
var searchPhrase = map[byte]string{
	0x2B: "在樹洞裡",   // a hollow stump
	0x4F: "在牆裡",     // a wall
	0x5A: "在窗台上",   // a window shelf
	0x5C: "在書架裡",   // a crowded bookshelf
	0x5D: "在書架裡",   //
	0xA1: "在井邊",     // a deep well
	0xA5: "在書桌裡",   // a desk
	0xA6: "在木桶裡",   // an oaken barrel
	0xA8: "在梳妝台裡", // a vanity
	0xAB: "在床底下",   // a bed
	0xAC: "在床底下",   //
	0xAD: "在五斗櫃裡", // a chest of drawers
	0xAF: "在置物箱裡", // a heavy footlocker
	0xB2: "在火盆裡",   // a hot brazier
	0xBC: "在壁爐裡",   // a fireplace
}

// SearchSecretDoor 是「有缺口的牆」——搜它會找到密門。
//
// `LOOK2` 把 0x4E 叫「a wall with a nick」,那個缺口就是給玩家的提示;
// 而 0x4F 是普通的牆,搜了只會翻到垃圾。**兩格差一號,行為完全不同。**
const SearchSecretDoor = 0x4E

// 陷阱偵測的常數(原版 `sub_13F04`)。
const (
	trapBaseDifficulty = 30 // 0x1E
	trapRollMax        = 30
	trapSimpleBelow    = 10 // 等級低於它是「簡單的陷阱」
	trapComplexAbove   = 20 // 高於它是「複雜的陷阱」
	trapFlag           = 0x80
	trapLevelMask      = 0x7F
)

// 搜到東西的機率(原版 `sub_13DD8`)。
const (
	searchFindOneIn = 8 // random(0,7) == 0 才是真的翻到東西
	searchFoodOneIn = 4 // 真的翻到之後,random(0,3) == 0 是糧食,否則金幣
	searchPlagueMax = 31
	searchPlagueHit = 19 // random(0,31) 剛好是它就中瘟疫
)

// Search 是 S 指令。
func (s *State) Search() {
	// 地牢有自己的一支(原版 `sub_147A8` 依地點分派到 `sub_142EC`)——
	// 方向是相對的、報的是地形、而且搜得出暗門。見 dungeonsearch.go。
	if s.InDungeon() {
		s.searchDungeon()
		return
	}
	s.AskDirection(func(d Direction) {
		dx, dy := d.Delta()
		s.searchAt(s.X+dx, s.Y+dy)
	})
}

// searchAt 搜某一格。
func (s *State) searchAt(x, y int) {
	member := s.pickCharacter("")
	if member < 0 {
		return
	}
	tile := s.TileAt(x, y)
	s.Log(searchPhrase[tile] + MsgThouDostFind)

	// 有物件就是查它的陷阱,不會再翻到別的東西。
	if o, _, ok := s.ObjectAt(x, y); ok {
		s.reportTrap(o, member)
		return
	}
	if tile == SearchSecretDoor {
		s.SetTileAt(x, y, u5data.TileDoorA)
		s.Log(MsgHiddenDoor)
		return
	}
	s.rollSearchFind(member)
}

// reportTrap 是陷阱偵測(原版 `sub_13F04`)。
//
// 難度 = 陷阱等級 + 30 − 智力,擲 random(1,30);難度的一半 ≤ 擲值才算看清。
// ⚠ 「看清」與「真的有陷阱」不一致時一律報反 —— 見檔頭那張表。
func (s *State) reportTrap(o *u5data.MapObject, member int) {
	q := o.Raw[u5data.ObjQuality]
	trapped := q&trapFlag != 0
	level := int(q & trapLevelMask)

	difficulty := trapBaseDifficulty - int(s.Roster[member].Intel)
	if trapped {
		difficulty += level
	}
	detected := difficulty/2 <= s.Roll(1, trapRollMax)

	if detected != trapped {
		s.Log(MsgNoTrap)
		return
	}
	switch {
	case detected && level < trapSimpleBelow:
		s.Log(MsgSimpleTrap)
	case detected && level > trapComplexAbove:
		s.Log(MsgComplexTrap)
	default:
		s.Log(MsgATrap)
	}
}

// rollSearchFind 擲「翻到了什麼」(原版 `sub_13DD8`)。
//
// ⚠ **八分之七是垃圾。** 只有 random(0,7) == 0 才真的翻到東西,
// 而那之中四分之三是金幣、四分之一是糧食。垃圾那一支裡還有
// 三十二分之一會中瘟疫(狀態變 'P')。
//
// 把機率調高會讓搜家具變成刷錢手段 —— 原版刻意讓它不划算。
func (s *State) rollSearchFind(member int) {
	if s.Roll(0, searchFindOneIn-1) != 0 {
		s.rollSearchJunk(member)
		return
	}
	slot := s.freeObjectSlot()
	if slot < 0 {
		s.Log(MsgNoRoom)
		return
	}
	kind := byte(u5data.ItemGold)
	msg := MsgFoundGold
	if s.Roll(0, searchFoodOneIn-1) == 0 {
		kind, msg = u5data.ItemFood, MsgFoundFood
	}
	s.placeFoundObject(slot, kind, byte(s.Roll(1, 3)))
	s.Log(msg)
}

// rollSearchJunk 是翻到垃圾的那一支,含瘟疫。
func (s *State) rollSearchJunk(member int) {
	if s.Roll(0, searchPlagueMax) == searchPlagueHit {
		s.Roster[member].Status = u5data.StatusPoisoned
		s.Log(MsgPlague)
		return
	}
	// ⚠ 原版是**嵌套的**亂數:`random(0, random(0,3))` ——
	// 所以「什麼也沒有」的機率遠高於其他三種。攤平成 random(0,3)
	// 會讓玩家一直翻到「血肉模糊的一團」,那不是原版的節奏。
	switch s.Roll(0, s.Roll(0, 3)) {
	case 1:
		s.Log(MsgFoundWorms)
	case 2:
		s.Log(MsgFoundGuts)
	case 3:
		s.Log(MsgFoundPulp)
	default:
		s.Log(MsgFoundNothing)
	}
}

// placeFoundObject 把翻到的東西放進物件槽。
func (s *State) placeFoundObject(slot int, kind, quality byte) {
	objs := s.currentObjects()
	if objs == nil || slot < 0 || slot >= len(objs.Objects) {
		return
	}
	o := &objs.Objects[slot]
	o.Raw[u5data.ObjKind] = kind
	o.Raw[u5data.ObjTile] = kind
	o.Raw[u5data.ObjX] = byte(s.X)
	o.Raw[u5data.ObjY] = byte(s.Y)
	o.Raw[u5data.ObjQuality] = quality
	o.Kind, o.Tile = kind, kind
	o.X, o.Y, o.Floor = s.X, s.Y, s.Floor
}
