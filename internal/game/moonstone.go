package game

import "github.com/wicanr2/u5-cht/internal/u5data"

// 埋月石(原版 `sub_1A2F8`)
//
// # 這裡原本被一句註解封起來了
//
// `use.go` 的說明寫著:
//
//	5..12 "(0".."(7" → case 21..28  **不可用**(原版跳到 default)
//	⚠ … 它們在跳表裡走 default,根本不會被用,**不要試著「修好」它們**
//
// 那句話是錯的,而且錯得很有說服力 —— 跳表裡那八格**確實**指向 `def_1A6DD`。
// 但它們指向 default 的原因是**在進 switch 之前就已經被接走了**:
//
//	loc_1A6B3:
//	    cmp eax, 14h ; jle → 走 switch      (≤ 20)
//	    cmp eax, 1Dh ; jge → 走 switch      (≥ 29)
//	    sub eax, 15h
//	    push eax
//	    call sub_1A2F8                      ; ★ 21..28 = 八顆月石
//
// ⇒ 只讀跳表會得到「這八格沒用」;要讀跳表**前面**那六行才看得到真相。
// (`rulebook/63`:程式碼是唯一真相 —— 但「讀了程式碼」不等於「讀對了範圍」。)
//
// # 埋在哪裡才算
//
//	tile = sub_DB10(玩家x, 玩家y)        ; 腳下那一格
//	印 "Moonstone "
//	if (byte_3E0A3 >= 0x21) → "cannot be buried here!"   ; 地牢與戰鬥都不行
//	可埋的地形:tile == 0x2C 或 0x2D,或 0x04 <= tile <= 0x0A
//	不然 → "cannot be buried here!"
//	印 "buried!"
//	byte_3E040[i] = 玩家x
//	byte_3E048[i] = 玩家y
//	byte_3E050[i] = byte_3E0A3    ; ★ 地點,不是「有沒有」
//	byte_3E058[i] = byte_3E0A5    ; ★ 樓層
//
// # ★ 順手更正一條存檔格式
//
// `save.go` 原本寫「`SaveMoonstonesOffset` 起 **16 B**,十六顆月石,0xFF = 在手上」,
// 並註明長度是被下一個已知欄位夾出來的。夾得沒錯,**顆數與語意錯了**:
//
//	byte_3E040[8]  存檔 0x028A   月石埋的 X      ← 原本整段沒解碼
//	byte_3E048[8]  存檔 0x0292   月石埋的 Y      ← 原本整段沒解碼
//	byte_3E050[8]  存檔 0x029A   月石埋在哪個地點(0xFF = 還在手上)
//	byte_3E058[8]  存檔 0x02A2   月石埋在哪一層
//
// (位移換算用同一個錨:`byte_3E000 ↔ 0x024A`、`byte_3E060 ↔ 0x02AA`,兩端夾住,
// 中間 0x60 個位元組沒有滑動空間。)
//
// 所以是**八顆**月石各四個欄位,不是十六顆旗標;而 `0xFF` 不是「在手上」的
// 特殊旗標,是「地點欄還沒被寫過」—— `sub_1E8D4` 拿 `== 0xFF` 當「可用」正是
// 因為沒埋才拿得出來。兩個佐證互相咬合:清單只掃 8 格(`cmp ecx, 8`)。
//
// ⚠ 埋下去之後那顆月石**怎麼變成月門**還沒追(`sub_E084` 讀的是另一組
// `Moongates` 表)。`⬜` 留在 `WORKLIST.md`,不在這裡猜。

// 可以埋月石的地形(原版的四個 `cmp`,注意兩個是 `jle`/`jge` 的開區間):
//
//	cmp esi, 2Ch ; jz  → 可
//	cmp esi, 2Dh ; jz  → 可
//	cmp esi, 3   ; jle → 不可
//	cmp esi, 0Bh ; jge → 不可
//	                   → 可(也就是 4..10)
//
// ★ 判準讀對了,是因為**九個 `look#` 名字一次全對**:
//
//	  3 淺灘        ← 下界外,水裡埋不了
//	  4 沼澤   5 草地   6 灌木   7 焦灼荒漠
//	  8 灌木   9 樹林  10 熱帶森林                ← 界內,全是挖得動的地面
//	 11 山麓        ← 上界外,石頭地
//	0x2C 犁過的地  0x2D 豐收莊稼                  ← 兩個單獨列的是**農地**
//
// 上下界都恰好卡在「水」與「岩」上,而兩個例外恰好是農田 —— 這不是巧合對得上,
// 是同一個判準的九個獨立錨點。(`rulebook/62`:同時命中的錨點證明表沒有滑動。)
const (
	MoonstoneBuryLoTile = 0x04
	MoonstoneBuryHiTile = 0x0A
	// MoonstoneBuryFieldTile 是犁過的地;+1 是豐收莊稼。
	MoonstoneBuryFieldTile = 0x2C
)

// MoonstoneBuryable 回報這一格地形埋不埋得下去。
func MoonstoneBuryable(tile byte) bool {
	if tile == MoonstoneBuryFieldTile || tile == MoonstoneBuryFieldTile+1 {
		return true
	}
	return tile >= MoonstoneBuryLoTile && tile <= MoonstoneBuryHiTile
}

// BuryMoonstone 埋下第 i 顆月石(原版 `sub_1A2F8`)。
func (s *State) BuryMoonstone(i int) bool {
	if i < 0 || i >= u5data.MoonstoneCount {
		return false
	}
	if !s.Inventory.Moonstones[i].InHand() {
		s.Log(MsgDontHaveThat)
		return false
	}
	s.Log(MsgMoonstone)
	if !s.SceneOrOverworld() || !MoonstoneBuryable(s.TileAt(s.X, s.Y)) {
		s.Log(MsgMoonstoneCannotBury)
		return false
	}
	s.Log(MsgMoonstoneBuried)
	s.Inventory.Moonstones[i] = u5data.Moonstone{
		X: s.X, Y: s.Y, Location: s.locationCode(), Floor: s.Floor,
	}
	return true
}
