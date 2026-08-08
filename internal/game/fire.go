package game

import "github.com/wicanr2/u5-cht/internal/u5data"

// Fire(F)—— 開砲
//
// 原版位址:`sub_172C4`、舷側判定在 `sub_17120`。
//
// ⚠ Mix(M)**不在這裡** —— 它早就實作在 `magic.go` 了(`State.Mix`)。
// 我一度在這裡重寫了一份,`vet` 才擋下來。教訓:**加新指令前先 grep 方法名**,
// 指令表有 26 支,記不住哪些做過了。

// ---------------------------------------------------------------- Fire

// Fire 是 F 指令:開砲。
//
// # 兩種砲
//
//	在船上  → 只能打**舷側**(垂直於船首方向)。原版 `sub_17120` 的
//	          「Fire broadsides only!」就是這一條。
//	在陸上  → 隊伍**緊鄰**的那一格要有大砲(tile 0xB4)。
//
// 原版查的是視野緩衝裡玩家四鄰的四格(`byte_3F769` 北 / `byte_3F78A` 東 /
// `byte_3F7A9` 南 / `byte_3F788` 西 —— 玩家自己在 `byte_3F789`,
// 緩衝一列 0x20 格,所以 ±1 是東西、±0x20 是南北)。
//
// ⚠ **只在地表**(原版 `cmp byte_3E0A3, 0; jnz` → 不是 0 就印「What?」)。
// 城裡的大砲要用 Push 推,不能開。
func (s *State) Fire() {
	if s.InScene() || s.InDungeon() {
		s.Log(MsgWhat)
		return
	}
	s.AskDirection(func(d Direction) { s.fireToward(d) })
}

// fireToward 往某個方向開砲。
func (s *State) fireToward(d Direction) {
	if s.isAboardShip() {
		if !s.isBroadside(d) {
			s.Log(MsgBroadsidesOnly)
			return
		}
	} else if !s.cannonBeside(d) {
		s.Log(MsgWhat)
		return
	}
	dx, dy := d.Delta()
	s.Log(MsgBooom)
	if s.InCombat() {
		// 戰場上與投射物同一條路。
		s.FlyProjectile(0, s.X+dx*cannonRange, s.Y+dy*cannonRange)
		return
	}
	// ⚠ 大地圖上 `FlyProjectile` 一開頭就 `if s.Combat == nil { return }` ——
	// 也就是說在此之前**開砲什麼都不會發生**,只印一句 BOOOM。
	// 原版 `sub_172C4` 的後半是另一條路,見 `fireCannonball`。
	s.fireCannonball(dx, dy)
}

// 大地圖上的砲彈(原版 `sub_172C4` 的後半)
//
// 砲彈沿方向飛最多 5 格,每一格先問「有沒有物件」再問「是不是門」:
//
//	物件是有效目標 → 打中:**整個槽直接清掉**(沒有血量、不進戰鬥)、
//	                 **業報 −5**(下限 0)、若那是個人就**被逮捕**
//	物件是隊伍自己(槽 0)→ ★ `sub_2A4D0()`:**全隊受傷**(自己打自己)
//	七種門         → 印 "Door destroyed!",那一格變成 **0x44 磚地**
//
// # 有效目標的判準(原版 `loc_1751C` 那一段)
//
//	kind == 0x10 或 0x11              → 是目標(★ 馬)
//	kind <  0x1C                      → 不是
//	kind & 0xF8 == 0x78               → 不是 ★ 那是**黑刺(14)與不列顛王(15)**
//	                                     (0x78 = 14×4+0x40、0x7C = 15×4+0x40)
//	其餘                              → 是目標
//
// ⚠ 中間還有一道 `and eax, 0FCh; cmp eax, 2Fh` —— **那是死碼**:
// `& 0xFC` 會清掉低兩位,結果永遠不可能等於 0x2F。照實際行為實作,不照抄意圖
// (同 `docs/re/61` 的酒館關鍵字那個死碼)。
//
// # 七種門
//
//	0x97 0x98 奇怪的門   0x99 柵門
//	0xB8 木門   0xB9 上鎖的門   0xBA 有窗戶的木門   0xBB 有窗戶的上鎖的門
//
// 打掉之後那一格寫成 **0x44**(`TileBrickFloor`)—— 與 An Ylem(消除)寫的是同一個值。

// cannonDoors 是砲彈打得掉的七種門(原版 `sub_172C4` 的七個 `cmp`)。
var cannonDoors = map[byte]bool{
	0x97: true, 0x98: true, 0x99: true,
	0xB8: true, 0xB9: true, 0xBA: true, 0xBB: true,
}

// cannonKarmaPenalty 是打中東西的業報代價(原版 `sub byte_3E098, 5`,下限 0)。
const cannonKarmaPenalty = 5

// cannonTargets 回報砲彈打不打得到這種物件。
func cannonTargets(kind byte) bool {
	if kind == u5data.TileHorse || kind == u5data.TileHorse+1 {
		return true // ★ 馬是特例:編號比 0x1C 小,但原版明文放行
	}
	if kind < 0x1C {
		return false
	}
	// ★ 黑刺與不列顛王打不掉(0x78..0x7F)。
	return kind&0xF8 != 0x78
}

// fireCannonball 讓砲彈飛出去。
func (s *State) fireCannonball(dx, dy int) {
	x, y := s.X, s.Y
	for step := 0; step < cannonRange; step++ {
		x, y = WrapWorld(x+dx), WrapWorld(y+dy)
		// 原版先問物件(`sub_2B3DC` 從槽 31 往下掃),再問地形。
		if o, slot, ok := s.ObjectAt(x, y); ok && cannonTargets(o.Kind) {
			s.cannonHit(slot, x, y)
			return
		}
		if cannonDoors[s.TileAt(x, y)] {
			s.Log(MsgDoorDestroyed)
			s.SetTileAt(x, y, u5data.TileBrickFloor)
			return
		}
	}
}

// cannonHit 是砲彈打中東西的後果。
func (s *State) cannonHit(slot, x, y int) {
	// ★ 打到槽 0(隊伍自己)→ 全隊受傷。原版的 `var_20 == 0` 那條路。
	if slot == u5data.PartyObjectSlot {
		s.damageWholeParty()
		return
	}
	s.currentObjects().Remove(slot)
	// 業報 −5,下限 0(原版 `cmp al,5; jbe → mov byte_3E098, 0`)。
	if s.Karma > cannonKarmaPenalty {
		s.Karma -= cannonKarmaPenalty
	} else {
		s.Karma = 0
	}
	// 打的是人就被逮捕(原版 `sub_2E0` 查 NPC → `sub_218` + `sub_268`)。
	if _, ok := s.NPCAt(x, y); ok {
		s.Arrest()
	}
}

// cannonRange 是砲彈飛多遠。原版沿著視野緩衝一路掃到邊界,
// 而視野是 11×11 —— 半徑 5 就到底了。
const cannonRange = 5

// isAboardShip 回報隊伍是不是在大船上(揚帆與否都算)。
func (s *State) isAboardShip() bool {
	k := u5data.VehicleKind(s.Transport)
	return k == u5data.VehicleShip || k == u5data.VehicleSailing
}

// isBroadside 回報這個方向是不是舷側。
//
// 船首方向就是載具碼的低兩位(0x24..0x27 / 0x20..0x23 各對四個朝向),
// 而舷側 = 與船首**垂直**。判斷法:方向碼與船首碼的奇偶不同就是垂直
//(方向碼北 0 / 東 1 / 南 2 / 西 3 —— 偶數是南北、奇數是東西)。
func (s *State) isBroadside(d Direction) bool {
	heading := int(s.Transport & 0x03)
	return (int(d)^heading)&1 == 1
}

// cannonBeside 回報那個方向的鄰格有沒有大砲。
func (s *State) cannonBeside(d Direction) bool {
	dx, dy := d.Delta()
	return s.TileAt(s.X+dx, s.Y+dy)&0xFC == pushGroupCannon
}
