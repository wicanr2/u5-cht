package game

import (
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 投射物飛行
//
// 箭、弩矢、投擲武器、攻擊法術與四種風走的是同一條路(原版 `sub_20CB4`):
//
//	1. 用 Bresenham 從射手拉一條直線到目標(`sub_2055C`,在 ×16 的像素空間算,
//	   每格再 +16 對到格子中心)
//	2. 沿線一格一格前進,每一步查 `sub_2BC34` —— **擋住就停在那裡**
//	3. 停下來的那一格上如果站著人,那個人就是實際被打到的
//
// ⇒ **會誤傷自己人。** 站在射線上的隊友會替敵人擋箭,這是原版行為,
// 不是我加的難度。

// iabs 是絕對值(測試檔已經有一個 abs,不重名)。
func iabs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// projectilePath 回傳從 (x0,y0) 到 (x1,y1) 的格子序列,**不含起點**。
//
// 用 Bresenham。原版在 ×16 的像素空間算再除回來,格子層級的結果一致 ——
// 差別只在動畫的平滑度,那是畫面的事。
func projectilePath(x0, y0, x1, y1 int) [][2]int {
	dx, dy := iabs(x1-x0), iabs(y1-y0)
	sx, sy := sign(x1-x0), sign(y1-y0)
	err := dx - dy
	var out [][2]int
	x, y := x0, y0
	for i := 0; i < 64; i++ {
		if x == x1 && y == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
		out = append(out, [2]int{x, y})
	}
	return out
}

func sign(v int) int {
	switch {
	case v > 0:
		return 1
	case v < 0:
		return -1
	}
	return 0
}

// FlyProjectile 讓一個投射物從 shooter 飛向 (tx, ty)。
//
// 回傳實際打到的單位槽號(−1 = 什麼都沒打到)與飛到哪一格停下。
// 停下的原因有三種:打到人、撞到擋箭的地形、飛出戰場。
func (s *State) FlyProjectile(shooter, tx, ty int) (hit, endX, endY int) {
	c := s.Combat
	if c == nil {
		return -1, tx, ty
	}
	u := &c.Units[shooter]
	x, y := u.X, u.Y
	for _, p := range projectilePath(u.X, u.Y, tx, ty) {
		if p[0] < 0 || p[0] >= u5data.CombatSide || p[1] < 0 || p[1] >= u5data.CombatSide {
			return -1, x, y
		}
		// 地形擋在人之前判 —— 原版 `sub_2BC34` 是在移動到那一格**之後**、
		// 畫下一步**之前**查的,所以擋住的那一格自己不算命中。
		if u5data.TileBlocksProjectile(int(c.Map.At(p[0], p[1]))) {
			return -1, x, y
		}
		x, y = p[0], p[1]
		if v, ok := c.CombatUnitAt(x, y); ok {
			return s.unitIndex(v), x, y
		}
	}
	return -1, x, y
}

// HasLineOfFire 回報從 from 射得到 (tx, ty) 而中間沒有東西擋。
func (s *State) HasLineOfFire(from, tx, ty int) bool {
	hit, ex, ey := s.FlyProjectile(from, tx, ty)
	_ = hit
	return ex == tx && ey == ty
}
