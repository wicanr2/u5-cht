package game

import "github.com/wicanr2/u5-cht/internal/u5data"

// 清過的地牢房間不會再有怪(原版 `sub_F9A0` / `sub_FA20` / `sub_FA7C`)
//
// 推導見 `docs/re/99`。三支函式組成一個機制:
//
//	sub_F9A0(房號)   打完一間房 → 在位元陣列 `byte_3E0F0` 上記一筆
//	sub_FA20(房號)   查那一筆
//	sub_FA7C()       ★ 進地牢時掃整座地牢的 512 格,把清過的房間
//	                 tile 從 `0xFn` 改成 `0xAn` —— 於是踏上去不再開打
//
// ⇒ 引擎此前**每次踏上房間格都會再打一場**,而原版打過就沒了。
// 這在毀滅或海斯洛斯那種要來回穿越的地牢裡差別非常大。

// markRoomCleared 把一間房記成「清過了」(原版 `sub_F9A0`)。
//
// ⚠ **兩件事用兩套不同的索引**,而那是原版的樣子:
//
//	例外清單的鍵 = 房號 | ((地點碼 & 0x0F) << 4)      ← 原始的低四位元
//	位元陣列的索引 = DungeonRoomBlock(地點碼)*16 + 房號  ← 有「≥1 就 −1」的修正
//
// 所以地點碼 0x21 與 0x22 在**位元陣列裡共用同一批位元**,但在例外清單裡
// 是不同的鍵。看起來像 bug,但兩處各自讀完都是這樣寫的 ——
// 而且 `DungeonRoomBlock` 那個修正早就存在(`DUNGEON.CBT` 的房間查表也用它),
// 所以它不是我算錯,是原版的索引方式。
func (s *State) markRoomCleared(location int, tile byte) {
	room := u5data.DungeonRoomNumber(tile)
	// ★ 例外清單上的房間**永遠不記** ⇒ 那幾間每次進去都有怪。
	if u5data.DungeonRoomAlwaysArmed(location, room) {
		return
	}
	idx := u5data.DungeonRoomIndex(location, tile)
	if idx < 0 || idx/8 >= len(s.roomsCleared) {
		return
	}
	s.roomsCleared[idx/8] |= 1 << (idx & 7)
}

// roomIsCleared 查一間房清過了沒(原版 `sub_FA20`)。
//
// ⚠ 它**不查例外清單** —— 例外只在「記」的那一邊生效。兩邊都查會得到
// 同樣的結果(沒記過就查不到),但少一次查表比較貼近原版。
func (s *State) roomIsCleared(location int, tile byte) bool {
	idx := u5data.DungeonRoomIndex(location, tile)
	if idx < 0 || idx/8 >= len(s.roomsCleared) {
		return false
	}
	return s.roomsCleared[idx/8]&(1<<(idx&7)) != 0
}

// applyClearedRooms 進地牢時把清過的房間從地圖上抹掉(原版 `sub_FA7C`)。
//
//	for (i = 0; i < 0x200; i++)                 ; 8 層 × 64 格 = 整座地牢
//	    if ((tile & 0F0h) == 0F0h && 清過了)
//	        tile &= 0AFh                         ; ★ 0xFn → 0xAn
//
// ★ `& 0xAF` 清掉的是 **0x50 兩個位元**,所以房間格會變成 `0xAn` ——
// 那是「空房間」那一族(可走、不觸發戰鬥),不是通道也不是牆。
// 用「設成 0」之類的簡化寫法會讓地圖上多出一片假通道。
func (s *State) applyClearedRooms() {
	d := s.Dungeon
	if d == nil || s.Dungeons == nil {
		return
	}
	for level := 0; level < u5data.DungeonLevels; level++ {
		for y := 0; y < u5data.DungeonSide; y++ {
			for x := 0; x < u5data.DungeonSide; x++ {
				tile := s.Dungeons.At(d.Index, level, x, y)
				if u5data.DungeonKind(tile) != u5data.DungeonRoomF {
					continue
				}
				if !s.roomIsCleared(s.locationCode(), tile) {
					continue
				}
				s.Dungeons.Set(d.Index, level, x, y, tile&u5data.DungeonRoomClearedMask)
			}
		}
	}
}

// 世界地圖依旗標重建:崩塌的地牢入口與毀壞的聖壇(原版 `sub_105E4`)
//
// 原版**每次載入世界地圖區塊**都重跑一次:
//
//	區塊裡的 tile 22 / 23 / 24(洞穴 / 礦坑 / 地牢)
//	    → 若 sub_1056C(區塊) 為真,改寫成 0xDF(崩塌的入口)
//	區塊裡的 tile 25(神秘聖壇)
//	    → 若 sub_105AC(區塊) 為真,改寫成 0x1A(毀壞的聖壇)
//
//	sub_1056C:八座地牢的區塊表裡找到 → `byte_3E0E0[i] == 0`;找不到 → **真**
//	sub_105AC:八座聖壇的區塊表裡找到 → `byte_3E0E8[i] > 0x7F`;找不到 → 假
//
// ⇒ 開局八個旗標全是 0(`INIT.GAM` 實測)⇒ **八座地牢入口都是崩塌的**,
// 而引擎此前直接用 `BRIT.DAT` 的原始地形 ⇒ 一開局所有地牢都能走進去,
// 包括末日。喊力量之言反而會把路封死(訊息也是反的)。
//
// ⬜ **兩處已知差異**,都寫在這裡而不是假裝一樣:
//
//  1. 原版是**整個區塊**掃 22/23/24 → 同一區塊裡**其他**的洞穴與礦坑也會
//     一起崩塌。這裡只改八個入口本身 —— 要做整區塊得先把 `byte_55140`
//     那張區塊表 dump 出來。
//  2. 原版的判準是「區塊不在表裡 → 真」,所以**不在任何地牢區塊裡的**
//     洞穴/礦坑一律崩塌。同 1,需要那張表才能重現。

// applyWorldFlags 依旗標把八座地牢入口與八座聖壇的地形改寫回去。
//
// 載入存檔之後要跑一次 —— 旗標存在存檔裡,而地形來自唯讀的 `BRIT.DAT`。
func (s *State) applyWorldFlags() {
	if s.World == nil {
		return
	}
	for i := range u5data.DungeonEntrances {
		e := u5data.DungeonEntrances[i]
		if !u5data.DungeonIsSealed(s.DungeonSeal[i]) {
			continue
		}
		// ⚠ 只在那一格**目前是原始入口地形**時才改 —— 已經是 0xDF 就不必動,
		// 而動了也無害(冪等),但少一次寫入比較容易在測試裡看清因果。
		if s.TileAt(e.X, e.Y) == u5data.DungeonEntranceTile[i] {
			s.SetTileAt(e.X, e.Y, u5data.TileDungeonSealed)
		}
	}
	for i := range s.ShrineFlag {
		if s.ShrineFlag[i]&u5data.ShrineDesecratedBit == 0 {
			continue
		}
		sh := u5data.Shrines[i]
		// ⚠ 靈性聖壇的座標是 (0,0) —— 它在**幽冥界**,不在地表這張圖上。
		// 不擋掉的話會把地表 (0,0) 那一格改成聖壇。
		if sh.X == 0 && sh.Y == 0 {
			continue
		}
		if s.TileAt(sh.X, sh.Y) == u5data.TileShrine {
			s.SetTileAt(sh.X, sh.Y, u5data.TileShrineDesecrated)
		}
	}
}
