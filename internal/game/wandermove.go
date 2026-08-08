package game

import "github.com/wicanr2/u5-cht/internal/u5data"

// 大地圖上的東西怎麼移動(原版 `sub_2D38` → `sub_2B24` / `sub_2A54` → `sub_2870`)
//
// 推導見 `docs/re/85`。⚠ 這一整條當初找不到,是因為沒讀完 `sub_2D38` 的尾巴 ——
// 它最後一行是 `call sub_2B24`,而那才是移動(`docs/re/84` §1)。

// stepObject 真的把物件挪一格(原版 `sub_2870`)。
//
//	kind 是敵船(& 0xFC == 0x2C)→ 依方向換朝向圖,**不吃地形延遲**
//	kind 是龍 / 蝙蝠 / 惡魔 / Mongbat → **免疫地形延遲**
//	其餘 → 依目標格的地形擲一次「這一步走不走得成」
//	走成了:記下舊座標(給 `notJustVacated` 用)、寫新座標
//	★ 走到 tile 0xDC 上就**整個消失**(種類碼與 tile 都歸零)
func (s *State) stepObject(slot, dx, dy int) {
	set := s.currentObjects()
	if set == nil {
		return
	}
	o := &set.Objects[slot]
	kind := o.Raw[u5data.ObjKind]
	nx := int(o.Raw[u5data.ObjX]) + dx
	ny := int(o.Raw[u5data.ObjY]) + dy
	tile := s.TileAt(WrapWorld(nx), WrapWorld(ny))

	switch {
	case kind&0xFC == u5data.SpawnEnemyShip:
		// 敵船換朝向圖:低兩位就是朝向(北東南西,見 `u5data.ShipFacing*`)。
		facing := u5data.ShipFacingForDelta(dx, dy)
		o.Raw[u5data.ObjKind] = byte(u5data.ShipTileBase + facing)
		o.Raw[u5data.ObjTile] = o.Raw[u5data.ObjKind]
		o.Kind, o.Tile = o.Raw[u5data.ObjKind], o.Raw[u5data.ObjTile]
	case u5data.CreatureIgnoresTerrain(kind):
		// ★ 會飛的四種免疫地形延遲。
	default:
		if !s.terrainLetsCreatureThrough(tile) {
			return // 這一步被地形吃掉了,原地不動
		}
	}

	// ★ 記下舊座標:`notJustVacated` 會擋住「下一個處理到的東西補進這一格」。
	s.lastVacatedX = int(o.Raw[u5data.ObjX])
	s.lastVacatedY = int(o.Raw[u5data.ObjY])
	o.Raw[u5data.ObjX] = byte(nx)
	o.Raw[u5data.ObjY] = byte(ny)
	o.X, o.Y = int(o.Raw[u5data.ObjX]), int(o.Raw[u5data.ObjY])

	// ★ 走到 tile 0xDC 就消失(原版 `cmp byte ptr [eax], 0DCh` → 兩個位元組歸零)。
	if tile == u5data.CreatureVanishTile {
		*o = u5data.MapObject{}
	}
}

// terrainLetsCreatureThrough 是地形對怪物的拖慢(原版 `sub_2870` 的跳表)。
//
//	沼澤 / 灌木 / 焦灼荒漠(4, 6, 7, 8)與荒漠(30, 31) → `random(0,1) == 0` 才走
//	林 / 熱帶林 / 丘 / 山 / 高峰(9..15)               → `random(0,2) == 2` 才走
//	其餘                                              → 一定走
//
// ★★ 這正好是玩家那張地形代價表的同一組地形(`docs/re/38`:4/6/7/8 是 1 級、
// 9–15 是 2 級)—— **同一組地形,玩家付額外回合,怪物付機率**。
// 兩邊是不同的機制,不要合併。
func (s *State) terrainLetsCreatureThrough(tile byte) bool {
	switch {
	case tile == 4 || tile == 6 || tile == 7 || tile == 8 || tile == 30 || tile == 31:
		return s.Roll(0, 1) == 0
	case tile >= 9 && tile <= 15:
		return s.Roll(0, 2) == 2
	}
	return true
}

// objectCanEnter 是「這個東西進得去那一格嗎」(原版 `sub_2778`)。
//
//	地形對這個 mover 通不通(`sub_2A694`)**而且**那一格沒有別的物件。
func (s *State) objectCanEnter(slot, x, y int) bool {
	set := s.currentObjects()
	if set == nil {
		return false
	}
	x, y = WrapWorld(x), WrapWorld(y)
	if !u5data.MoverCanEnter(set.Objects[slot].Raw[u5data.ObjKind], int(s.TileAt(x, y))) {
		return false
	}
	if _, other, ok := s.ObjectAt(x, y); ok && other != slot {
		return false
	}
	return true
}

// notJustVacated 擋住「補進上一個東西剛離開的那一格」(原版 `sub_27CC`)。
//
// ⚠ 那對座標是**全域**的(`byte_4FD94` / `byte_4FD95`),不是每一槽一份 ——
// 只由 `sub_2870` 在移動前寫、只由這裡讀。世界回合倒著掃槽,所以它擋的是
// 「這一輪剛移動過的那一個留下的空格」。看起來像防呆,照原樣做。
//
// ★ `sub_2A54`(隨機遊走)**不查這一條**,只有 `sub_2B24`(追人)會查。
func (s *State) notJustVacated(x, y int) bool {
	return WrapWorld(x) != WrapWorld(s.lastVacatedX) || WrapWorld(y) != WrapWorld(s.lastVacatedY)
}

// chaseParty 是往隊伍走一格(原版 `sub_2B24`)。
//
//	dx = 帶號環繞(物件X − 隊伍X);dy 同理
//	stepX = −sign(dx);stepY = −sign(dy)          ; 往隊伍靠
//	random(0,1) 決定**先試哪一軸**
//	兩軸都進不去 → ★ 退回隨機遊走(`sub_2A54`)
//
// ★ 最後那一條是原版不會讓怪物卡在牆邊發呆的原因。
func (s *State) chaseParty(slot int) {
	set := s.currentObjects()
	if set == nil {
		return
	}
	o := &set.Objects[slot]
	ox, oy := int(o.Raw[u5data.ObjX]), int(o.Raw[u5data.ObjY])
	stepX := -signOf(signedWrap(ox - s.X))
	stepY := -signOf(signedWrap(oy - s.Y))

	tryX := func() bool {
		if stepX == 0 {
			return false
		}
		nx := ox + stepX
		if !s.objectCanEnter(slot, nx, oy) || !s.notJustVacated(nx, oy) {
			return false
		}
		s.stepObject(slot, stepX, 0)
		return true
	}
	tryY := func() bool {
		if stepY == 0 {
			return false
		}
		ny := oy + stepY
		if !s.objectCanEnter(slot, ox, ny) || !s.notJustVacated(ox, ny) {
			return false
		}
		s.stepObject(slot, 0, stepY)
		return true
	}

	// ★ 先試哪一軸是擲出來的 —— 少了這一擲,怪物會沿著固定的階梯路線接近,
	// 看起來像在走格線而不是在追人。
	if s.Roll(0, 1) == 1 {
		if tryX() || tryY() {
			return
		}
	} else if tryY() || tryX() {
		return
	}
	s.wanderRandomly(slot)
}

// WanderTries 是隨機遊走的嘗試次數(原版 `cmp [ebp+var_4], 3; jge`)。
const WanderTries = 3

// wanderRandomly 是隨機挑一個方向走(原版 `sub_2A54`)。
//
// 最多試 3 次,每次擲 `random(0,3)` 選四方向之一;進得去就走,否則再試。
// ⚠ 它**只查地形與佔位**(`sub_2778`),不查 `notJustVacated`。
func (s *State) wanderRandomly(slot int) {
	set := s.currentObjects()
	if set == nil {
		return
	}
	o := &set.Objects[slot]
	ox, oy := int(o.Raw[u5data.ObjX]), int(o.Raw[u5data.ObjY])
	deltas := [4][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}
	for i := 0; i < WanderTries; i++ {
		d := deltas[s.Roll(0, 3)]
		if s.objectCanEnter(slot, ox+d[0], oy+d[1]) {
			s.stepObject(slot, d[0], d[1])
			return
		}
	}
}

// objectMoveGate 是「這一槽這回合動不動」(原版 `sub_2D38`)。
//
// ⚠⚠ 回傳值的語意當初標反了(`docs/re/84` §1):走到 `sub_2B24` 的那條路
// 才是**移動**。三個特例:
//
//	漩渦(& 0xFC == 0xEC):+5 位元是持久切換 ⇒ 隔次才有機會動;
//	                       輪到它時再擲 1/2,擲中走一般移動、否則隨機遊走
//	0xFC:`sub_27F0` 為 0 → 移動;否則 +5 累加到 0x14 之前都不動(⬜ 兩支下游未讀)
//	敵船(& 0xFC == 0x2C):無風不動;查速度表,**值越大越快**,
//	                       `4` = 每回合都動;計數器 ≤ 值就動,超過才跳一回合並歸零
func (s *State) objectMoveGate(slot int) {
	set := s.currentObjects()
	if set == nil {
		return
	}
	o := &set.Objects[slot]
	kind := o.Raw[u5data.ObjKind]

	switch {
	case kind&0xFC == WhirlpoolKind:
		o.Raw[u5data.ObjQuality] ^= 1
		if o.Raw[u5data.ObjQuality] == 0 {
			return
		}
		if s.Roll(0, 1) != 0 {
			s.chaseParty(slot)
			return
		}
		s.wanderRandomly(slot)
		return

	case kind&0xFC == u5data.SpawnEnemyShip:
		if s.Wind == u5data.WindCalm {
			return // ★ 無風,敵船不動
		}
		speed := s.enemyShipSpeed(kind)
		if speed >= u5data.ShipNeverMoves {
			s.chaseParty(slot)
			return
		}
		o.Raw[u5data.ObjShipTick]++
		if int(o.Raw[u5data.ObjShipTick]) <= speed {
			s.chaseParty(slot)
			return
		}
		o.Raw[u5data.ObjShipTick] = 0
		return
	}
	s.chaseParty(slot)
}

// enemyShipSpeed 查敵船這個朝向在目前的風下的速度值(原版 `dword_4FD50`)。
//
// ⚠ 名字裡沒有「延遲」是刻意的:**值越大越快**(`docs/re/84` §1)。
func (s *State) enemyShipSpeed(kind byte) int {
	if s.WindDelay == nil {
		return u5data.ShipNeverMoves
	}
	return s.WindDelay.Delay(int(kind-u5data.ShipTileBase)&0x03, s.Wind)
}

// signedWrap 把 8 位元的差值變成帶號距離(原版 `and 0FFh` + `> 7Fh` 就減 0x100)。
func signedWrap(d int) int {
	d &= 0xFF
	if d > 0x7F {
		d -= 0x100
	}
	return d
}

func signOf(v int) int {
	switch {
	case v > 0:
		return 1
	case v < 0:
		return -1
	}
	return 0
}
