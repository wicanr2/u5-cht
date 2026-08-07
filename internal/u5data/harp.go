package u5data

// 豎琴與那首曲子(原版 `sub_11E0`)
//
// 站在豎琴前面按數字鍵就會發出音。在**不列顛王城堡二樓**把那首十三個音的
// 曲子彈對,牆上會開一道暗門。
//
// 追出來的過程:
//
//	byte_3F7A9 == 0x8D            玩家**正南**那一格是豎琴才進得來
//	                              (byte_3F789 是玩家那一格,+0x20 = 往南一列)
//	edi = 按鍵 − 0x30             '0'..'9' → 0..9
//	音高 = 音階[edi] + 0x3C       0x3C = 60 = 中央 C
//	序列比對成功十三次 → 地點 17 二樓時 scene[13][17] ^= 0x0B

// HarpTile 是豎琴那一格。切出來看確實是一把豎琴(0x8B 椅子、0x8E 鐵砧、0x8F 熔岩)。
const HarpTile = 0x8D

// HarpScale 是十個按鍵對應的半音(原版 `dword_478C8` 前 11 B 的前 10 個)。
//
//	鍵  1  2  3  4  5  6  7  8  9   → 0 2 4 5 7 9 11 12 14  = 一個大調音階
//	鍵  0                            → 16
//
// ⚠ 索引 0(按鍵 '0')的值是 16,不是 0 —— 它是音階外的高音,
// 不要「順手」把它排成 0 讓表看起來整齊。
var HarpScale = [10]int{16, 0, 2, 4, 5, 7, 9, 11, 12, 14}

// HarpMiddleC 是加在半音上的基準(`add eax, 3Ch`)。
const HarpMiddleC = 0x3C

// HarpTune 是那首曲子(原版 `byte_4FC80` 起 13 B)。
var HarpTune = [13]int{6, 7, 8, 9, 8, 7, 8, 7, 6, 7, 6, 5, 3}

// 彈對之後的效果(`sub_11E0` 的 `loc_1269` 尾段)。
const (
	// HarpDoorLocation / HarpDoorFloor 是唯一有反應的地方:不列顛王的城堡二樓。
	HarpDoorLocation = 17
	HarpDoorFloor    = 2
	// HarpDoorX / HarpDoorY 是被切換的那一格。
	//
	// ★ 座標是算出來的,不是猜的:`sub_11E0` 對 `byte_402A5` 做 XOR,
	// 而 `sub_DB10` 告訴我們**場景緩衝的基底是 `byte_400F4`、列距 32**
	//(地點 0 走 `byte_404F4` 的四象限、地點 > 0x7F 走 `byte_3F8F4`)。
	// 0x402A5 − 0x400F4 = 0x1B1 = 433 = 13×32 + 17。
	HarpDoorX = 17
	HarpDoorY = 13
	// HarpDoorXor 是切換用的 XOR 值。一個 XOR 同時做開與關。
	HarpDoorXor = 0x0B
)

// HarpNext 回傳彈完這個音之後的進度(原版 `loc_1269` / `loc_12BB`)。
//
// 對了就往前一格;錯了**不是直接歸零** —— 原版手寫了一組回退規則,
// 效果等同 KMP 的 failure function:
//
//	進度 10 而且彈了 8  → 退回 3
//	進度 11 而且彈了 7  → 退回 2
//	彈的是曲子的第一個音 → 退回 1
//	其餘                → 0
//
// ⚠ 少了前兩條,玩到一半彈錯一個音就得從頭來 —— 而原版讓汝從中間接回去。
// 那兩個數字(10→3、11→2)不是隨便選的:曲子的第 3..4 個音是 `8 9`、
// 第 10..11 個音是 `7 6`,回退點正是「已經彈對的尾巴還能當開頭」的位置。
func HarpNext(progress, note int) int {
	if progress >= 0 && progress < len(HarpTune) && HarpTune[progress] == note {
		return progress + 1
	}
	switch {
	case progress == 10 && note == 8:
		return 3
	case progress == 11 && note == 7:
		return 2
	case note == HarpTune[0]:
		return 1
	}
	return 0
}
