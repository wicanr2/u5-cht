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
