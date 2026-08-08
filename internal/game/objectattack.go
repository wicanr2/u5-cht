package game

import "github.com/wicanr2/u5-cht/internal/u5data"

// 大地圖上的東西怎麼碰到隊伍(原版 `sub_25F0` → `sub_24DC`)
//
// 推導見 `docs/re/83`。⚠ `sub_25F0` 的名字看起來像「讓這一槽動一步」,
// 實際上它**完全不改座標** —— 三條路全部是攻擊。
// 「怪物移動」在 `sub_2E24` 的兩支下游裡都找不到(`docs/re/83` §1 的 ⬜)。

// 遠程攻擊者與射程(原版 `sub_25F0` 的 `cmp eax, 88h` / `0DCh`)。
const (
	// RangedSeaSerpent / RangedDragon 是**唯二**有遠程攻擊的種類。
	RangedSeaSerpent = 0x88
	RangedDragon     = 0xDC
	// RangedReach 是射程:兩軸各自 ≤ 3 ⇒ **方形**範圍,不是圓形。
	RangedReach = 3
	// RangedFireOneIn 是開火機率的分母:`random(0, 7) != 0` 就不開火 ⇒ 1/8。
	RangedFireOneIn = 8
)

// 貼身接觸的三個特例(原版 `sub_24DC`)。
const (
	// WhirlpoolKind 是漩渦(`kind & 0xFC == 0xEC`)。
	//
	// ★★ 它**不是生物**,是地理特徵 —— 這解釋了為什麼 `DATA.OVL` 的生物名表
	// 在索引 43 是空的。三重佐證:
	//  1. `sub_24DC` 印的字串是 `"\nWHIRLPOOL!\n"`
	//  2. 物件種類碼 +256 = tile 0x1EC = 492,而 `look#492` 就是「漩渦」
	//  3. 已翻好的 NPC 對白:「有人說被漩渦吸下去的船、後來擱淺在地底什麼地方」
	//     (`DWELLING.TLK#2#e11`)—— 遊戲內的傳聞描述的正是這個機制
	WhirlpoolKind = 0xEC
	// SandTrapKind 是流沙陷阱那一族(`kind & 0xFC == 0xE0`)——
	// 貼上去**只損船、不開戰**。
	SandTrapKind = 0xE0
	// WhirlpoolExitX / Y 是被漩渦吸下去之後的落點(原版寫死 `0x22` / `0x12`)。
	WhirlpoolExitX = 0x22
	WhirlpoolExitY = 0x12
)

// objectAttacks 是一個物件槽對隊伍做的事(原版 `sub_25F0`)。
//
// 回傳值同原版:非 0 代表「這一槽出事了」。`sub_2E24` 用它決定後面的槽
// 還要不要跑漂流閘門(`docs/re/83` §1)。
func (s *State) objectAttacks(slot int) bool {
	set := s.currentObjects()
	if set == nil || slot < 0 || slot >= len(set.Objects) {
		return false
	}
	o := &set.Objects[slot]
	kind := o.Raw[u5data.ObjKind]
	dx := wrapDistance(int(o.Raw[u5data.ObjX]) - s.X)
	dy := wrapDistance(int(o.Raw[u5data.ObjY]) - s.Y)

	// ★ 相鄰**只算正交**:斜對角(1,1)不算,所以斜著貼著隊伍的怪不會動手。
	// 寫成「切比雪夫距離 1」會讓怪物比原版兇一圈。
	if (dx == 1 && dy == 0) || (dx == 0 && dy == 1) {
		s.contactParty(slot)
		return true
	}

	// ★ 只有海蛇與龍有遠程攻擊。
	if kind == RangedSeaSerpent || kind == RangedDragon {
		if dx > RangedReach || dy > RangedReach {
			return false
		}
		if s.Roll(0, RangedFireOneIn-1) != 0 {
			return false
		}
		// ⚠ 原版接的是船損那一對(`sub_2B1C8` + `sub_22F0`)——
		// **遠程命中打的是船不是人**。在陸地上被龍噴到會發生什麼,取決於
		// `sub_2B1C8` 對非船載具怎麼處理,⬜ 未追(`docs/re/83`)。
		s.Log(MsgRangedAttack)
		s.DamageShip()
		return true
	}

	// ⬜ 敵船開砲(`kind & 0xFC == 0x2C`,正交對齊且距離 < 4 → `"* BOOOM! *"`)
	// 還沒做 —— `sub_23FC` 未讀,而照印象寫「船會開砲」等於自創傷害規則。
	return false
}

// wrapDistance 是原版那兩行環繞距離(`abs(d)`,`> 0x7F` 就取 `0x100 − d`)。
func wrapDistance(d int) int {
	d = absInt(d) & 0xFF
	if d > 0x7F {
		d = 0x100 - d
	}
	return d
}

// contactParty 是「貼著隊伍」的三條路(原版 `sub_24DC`)。
//
//	if (漩渦) { 步行 → 只損船;否則 → 吸下去 }
//	if (流沙陷阱族) → 只損船
//	印 "Attacked!"
//	腳下 tile >= 4(陸地)     → 開戰
//	魔毯                      → 只損船
//	不是小艇                  → 開戰
//	是小艇                    → 只損船
//
// ★ 所以**在小艇上被貼身只會掉耐久,不會開戰** —— 在船上是安全的,
// 而那正是 U5 海戰的節奏。寫成「一律開戰」會讓小艇變成活棺材。
func (s *State) contactParty(slot int) {
	set := s.currentObjects()
	if set == nil {
		return
	}
	o := &set.Objects[slot]
	kind := o.Raw[u5data.ObjKind]

	if kind&0xFC == WhirlpoolKind {
		// ★ 步行時漩渦碰不到你(原版 `cmp byte_3E08C, 1Ch; jz → 只損船`)。
		if s.Transport == u5data.VehicleWalk {
			s.DamageShip()
			return
		}
		s.suckIntoWhirlpool(slot)
		return
	}
	if kind&0xFC == SandTrapKind {
		s.DamageShip()
		return
	}

	s.Log(MsgAttacked)
	// ★ 腳下是陸地(tile >= 4)就開戰,不管載具。
	if s.TileAt(s.X, s.Y) >= 4 {
		s.beginObjectCombat(slot)
		return
	}
	// 水上:魔毯與小艇只損船,其餘開戰。
	if s.Transport&0xFE == ridingCarpet || u5data.VehicleKind(s.Transport) == u5data.VehicleSkiff {
		s.DamageShip()
		return
	}
	s.beginObjectCombat(slot)
}

// suckIntoWhirlpool 是被漩渦吸下去(原版 `sub_24DC` 的漩渦分支)。
//
// 順序照原版:**漩渦自己先消失**(種類碼與 tile 都歸零)→ 印字 → 船損 →
// 傳送到幽冥界的固定座標。⚠ 漩渦是**一次性**的,吸完就沒了。
func (s *State) suckIntoWhirlpool(slot int) {
	set := s.currentObjects()
	if set != nil && slot >= 0 && slot < len(set.Objects) {
		set.Objects[slot] = u5data.MapObject{}
	}
	s.Log(MsgWhirlpool)
	s.DamageShip()
	if s.Under == nil {
		// 沒載幽冥界地圖就誠實停在這裡,不假裝吸下去(同 `fallDownTheWaterfall`
		// 的處理;`CLAUDE.md §3.0`:缺素材要優雅降級並明說)。
		return
	}
	s.Floor = UnderworldFloor
	s.X, s.Y = WhirlpoolExitX, WhirlpoolExitY
	s.placeUnderworldItems()
}

// beginObjectCombat 把那一槽當敵人開戰(原版 `sub_2E58C`)。
func (s *State) beginObjectCombat(slot int) {
	s.BeginCombat(slot)
}
