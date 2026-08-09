package game

import "github.com/wicanr2/u5-cht/internal/u5data"

// 無主的馬會在城裡自己走動,除非牠被繫住(原版 `sub_FEC` + `sub_FB4`)
//
// 推導見 `docs/re/99`。這一支掛在**場景主迴圈** `sub_1A54` 上,
// 與「打開的門四回合後自己關上」是同一個位置的鄰居 —— 而引擎只做了門。
//
// ★★ 認出這是「馬」靠的是 look 表的**物件段**:索引 256 + 種類碼,
// 而 `look#272` / `look#273` 都寫著「a horse」⇒ 種類碼 0x10 / 0x11 是馬的
// 兩個朝向(原版比的正是 `kind & 0xFE == 0x10`)。
// 同樣的方法認出牠避開的那兩格:`look#162` = 繫馬柱、`look#67` = 欄杆。
//
// ⚠ 別拿地形段的 look 去查物件種類碼:索引 < 256 才是地形,
// 而 tile 0x10 在地形段是「a small hut」—— 用錯段會得到「小屋會自己走」。

// 馬的種類碼與繫馬的兩種地形。
const (
	// HorseKindEast 是朝東的馬(原版往 +x 走時寫 0x10)。
	HorseKindEast = 0x10
	// HorseKindWest 是朝西的馬(往 −x 走時寫 0x11)。
	HorseKindWest = 0x11
	// HorseKindMask 是原版的比對遮罩(`kind & 0xFE == 0x10`)。
	HorseKindMask = 0xFE

	// TileHitchingPost 是繫馬柱(`look#162`)。
	TileHitchingPost = 0xA2
	// TileRail 是欄杆(`look#67`)。
	TileRail = 0x43

	// SceneCoordMax 是場景座標的上限(原版 `cmp edi, 1Fh; jg`)。
	//
	// ⚠ 原版**不環繞**,超出就不走 —— 與大地圖的 `WrapWorld` 不同。
	SceneCoordMax = 31
)

// horseTied 回報 (x, y) 那一格是不是繫馬柱或欄杆(原版 `sub_FB4`)。
func (s *State) horseTied(x, y int) bool {
	t := s.TileAt(x, y)
	return t == TileHitchingPost || t == TileRail
}

// wanderHorses 讓場景裡沒被繫住的馬走一步(原版 `sub_FEC`)。
//
//	for (槽 = 0; 槽 < 32; 槽++) {
//	    if ((kind & 0FEh) != 10h)        continue   ; 不是馬
//	    if (槽.樓層 != 當前樓層)          continue
//	    if (random(0,1) != 0)            continue   ; ★ 1/2 才考慮動
//	    if (四個鄰格任一是柱或欄)         continue   ; ★ 被繫住了
//	    if (random(0,1) != 0) {                     ; 水平
//	        d = 2*random(0,1) − 1                   ; ±1
//	        x += d;  kind = (d > 0) ? 10h : 11h     ; ★ 朝向跟著改
//	    } else {
//	        y += 2*random(0,1) − 1                  ; ★ 垂直**不改朝向**
//	    }
//	    if (x 或 y 出了 0..31)            continue
//	    if (馬走不上那個地形)             continue
//	    if (那一格已經有東西)             continue
//	    寫回 kind / tile / x / y
//	}
//
// ⚠ 四個容易寫錯的地方:
//
//  1. **繫住的判準是「四個鄰格」不是「腳下」。** 馬站在柱子**旁邊**就算被繫住,
//     所以柱子那一格自己空著。寫成看腳下的話城裡的馬會全部跑掉。
//  2. **垂直移動不改朝向。** 只有左右走才換 tile —— 上下走維持原本的朝向。
//  3. **座標不環繞。** 場景是 32×32 而原版直接比 0..31,出界就這一回合不動。
//  4. **1/2 的閘門在最前面**,在讀座標之前 —— 所以「有沒有被繫住」這件事
//     一半的回合根本不查。行為上看不出差別,但隨機序列會不同。
func (s *State) wanderHorses() {
	if !s.InScene() {
		return
	}
	objs := s.currentObjects()
	if objs == nil {
		return
	}
	for i := range objs.Objects {
		o := &objs.Objects[i]
		if o.Kind&HorseKindMask != HorseKindEast {
			continue
		}
		if o.Floor != s.Floor {
			continue
		}
		if s.Roll(0, 1) != 0 {
			continue
		}
		x, y := o.X, o.Y
		if s.horseTied(x, y+1) || s.horseTied(x+1, y) ||
			s.horseTied(x, y-1) || s.horseTied(x-1, y) {
			continue
		}
		kind := o.Kind
		if s.Roll(0, 1) != 0 {
			d := 2*s.Roll(0, 1) - 1
			x += d
			if d > 0 {
				kind = HorseKindEast
			} else {
				kind = HorseKindWest
			}
		} else {
			y += 2*s.Roll(0, 1) - 1
		}
		if x > SceneCoordMax || y > SceneCoordMax || x < 0 || y < 0 {
			continue
		}
		if !u5data.MoverCanEnter(HorseKindEast, int(s.TileAt(x, y))) {
			continue
		}
		if _, _, taken := s.ObjectAt(x, y); taken {
			continue
		}
		o.Kind, o.Tile = kind, kind
		o.Raw[u5data.ObjKind], o.Raw[u5data.ObjTile] = kind, kind
		o.X, o.Y = x, y
		o.Raw[u5data.ObjX], o.Raw[u5data.ObjY] = byte(x), byte(y)
	}
}
