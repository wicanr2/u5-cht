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
	// 砲彈沿著那個方向飛,打到第一個擋住的東西為止 —— 與投射物同一條路。
	s.FlyProjectile(0, s.X+dx*cannonRange, s.Y+dy*cannonRange)
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
