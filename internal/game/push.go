package game

// Push 指令(原版 `sub_18154`,搬運在 `sub_1806C` / `sub_180E0`)
//
// 「推」在 U5 不只是把家具挪開:椅子與大砲有**四個朝向**,推完會轉向;
// 而推不動的時候會自動改成「拉」—— 兩者都以玩家往前走一步收尾。
//
// # 流程
//
//	1. 問方向,目標格 = 隊伍位置 + 方向
//	2. 目標格上有物件、或那個 tile 不在可推清單裡 → 「推不動!」
//	3. 決定該留下哪一種地板:大砲(0xB4 群)留 0x45,其餘留 0x44
//	4. **推**:目標格再往前一格沒有物件、而且正好是那種地板 → 家具前進
//	   **拉**:否則,若**隊伍腳下**正好是那種地板 → 家具退到隊伍原處
//	   兩者都不成立 → 「推不動」
//	5. 椅子(0x90 群)與大砲(0xB4 群)按方向轉向;拉的時候朝向再反過來
//	6. 隊伍往那個方向走一步
//
// ⚠ 第 4 步的「拉」不是另一個指令 —— 原版就是同一支函式的 else 分支,
// 而且拉完之後隊伍與家具**交換位置**。寫成「推不動就算了」會少掉半個機制。

// pushFloorDefault / pushFloorCannon 是家具挪走之後留下的地板。
//
// 兩個都是 cobble,但**不能互換**:目的地必須「正好是」對應的那一種,
// 所以大砲只推得動在 0x45 上、家具只推得動在 0x44 上。
const (
	pushFloorDefault = 0x44
	pushFloorCannon  = 0x45
)

// 會轉向的兩群:椅子與大砲。其餘家具推了只是移位。
const (
	pushGroupChair  = 0x90
	pushGroupCannon = 0xB4
)

// pushable 是推得動的東西(原版 `sub_17F58` 的 switch)。
var pushable = map[byte]bool{
	0x90: true, 0x91: true, 0x92: true, 0x93: true, // 椅子 ×4 朝向
	0xA5: true, // 書桌
	0xA6: true, // 橡木桶
	0xA8: true, // 梳妝台
	0xA9: true, // 水壺
	0xAD: true, // 五斗櫃
	0xAE: true, // 邊桌
	0xAF: true, // 沉重的置物箱
	0xB4: true, 0xB5: true, 0xB6: true, 0xB7: true, // 大砲 ×4 朝向
}

// Pushable 回報這個 tile 推不推得動。
func Pushable(tile byte) bool { return pushable[tile] }

// Push 是 P 指令。
func (s *State) Push() {
	s.AskDirection(func(d Direction) {
		dx, dy := d.Delta()
		s.pushToward(dx, dy)
	})
}

// pushToward 往 (dx, dy) 推。
func (s *State) pushToward(dx, dy int) {
	tx, ty := s.X+dx, s.Y+dy
	tile := s.TileAt(tx, ty)
	if _, _, hasObj := s.ObjectAt(tx, ty); hasObj || !Pushable(tile) {
		s.Log(MsgWontBudge)
		return
	}

	floor := byte(pushFloorDefault)
	if tile&0xFC == pushGroupCannon {
		floor = pushFloorCannon
	}

	bx, by := tx+dx, ty+dy
	_, _, beyondObj := s.ObjectAt(bx, by)

	switch {
	case !beyondObj && s.TileAt(bx, by) == floor:
		// 推:家具前進一格,原地變地板。
		s.SetTileAt(bx, by, s.pushFacing(tile, dx, dy, false))
		s.SetTileAt(tx, ty, floor)
	case s.TileAt(s.X, s.Y) == floor:
		// 拉:家具退到隊伍腳下,目標格變地板 —— 然後兩者交換位置。
		s.SetTileAt(s.X, s.Y, s.pushFacing(tile, dx, dy, true))
		s.SetTileAt(tx, ty, floor)
		s.Log(MsgPulled)
		s.stepAfterPush(dx, dy)
		return
	default:
		s.Log(MsgWontBudge)
		return
	}
	s.Log(MsgPushed)
	s.stepAfterPush(dx, dy)
}

// pushFacing 算出家具搬完之後該畫哪一格。
//
// 只有椅子與大砲會轉向(原版判 `tile & 0xFC` 是不是 0x90 或 0xB4);
// 其餘家具原樣搬過去。朝向:北 +0、東 +1、南 +2、西 +3(原版 `sub_18028`),
// **拉的時候整個反過來**(`^ 2` —— 與 `Direction.Opposite` 同一個手法)。
func (s *State) pushFacing(tile byte, dx, dy int, pull bool) byte {
	group := tile & 0xFC
	if group != pushGroupChair && group != pushGroupCannon {
		return tile
	}
	face := 0
	switch {
	case dx == 1 && dy == 0:
		face = 1
	case dx == 0 && dy == 1:
		face = 2
	case dx == -1 && dy == 0:
		face = 3
	}
	if pull {
		face ^= 2
	}
	return group | byte(face)
}

// stepAfterPush 是收尾:隊伍往推的方向走一步。
//
// ⚠ 推與拉**都會走**。原版 `loc_1830F` 在兩個分支匯流之後才加座標,
// 不是只有推才走 —— 拉的結果因此是「隊伍與家具交換位置」。
func (s *State) stepAfterPush(dx, dy int) {
	if s.InScene() {
		s.X, s.Y = s.X+dx, s.Y+dy
	} else {
		s.X, s.Y = WrapWorld(s.X+dx), WrapWorld(s.Y+dy)
	}
	s.tick()
}
