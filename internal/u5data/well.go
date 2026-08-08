package u5data

// 許願井(原版 `sub_CD28`)—— 一個藏了三十多年的彩蛋
//
// `docs/re/xx` 之前把這一格結論寫成「原版接著問『Drop a coin?』但**不管答什麼
// 都沒有後續**,組語裡沒有任何寫入」。**那是錯的**,而錯因正是 `CLAUDE.md` §4.4
// 那條 `[HARD]`:當時讀的是 Hex-Rays 的輸出,而它把 `sub_CD28` **截斷在 Y/N
// 迴圈那裡**就 return 了 ——
//
//	char sub_CD28()          // ← Hex-Rays 版:連參數都沒了
//	{
//	  sub_23C18("a well.\n\nDrop a coin?");
//	  do v2 = sub_29EEC(...); while ( v2 != 'Y' && v2 != 'N' );
//	  if ( v2 == 'N' ) return sub_23C18("No\n");
//	  else             return sub_23C18("Yes\n");   // ← 真正的內容從這裡才開始
//	}
//
// 組語裡它有 **三個參數**、長度是反編譯版的四倍,而且**寫得很滿**。
//
// # 完整流程(`sub_CD28(玩家X, 玩家Y, 樓層)`,由 Look 的分派 `sub_D258` 在
// tile 0xA1 時呼叫,三個參數是 `byte_3E0A6` / `byte_3E0A7` / `byte_3E0A5`)
//
//	印 "a well.\n\nDrop a coin?"
//	等 'Y' 或 'N'(其他鍵繼續等)
//	'N' → 印 "No\n"、結束
//	'Y' → 印 "Yes\n"
//	      if (金錢 == 0) → 結束     ← ★ 一句話都不印,靜靜地什麼都沒發生
//	      印 "\nThy wish?\n"
//	      金錢--                    ← ★ 先扣一枚,不論許什麼願
//	      讀入最多 12 個字元
//	      空的 → 印 "Nothing\n"、結束
//	      六個字串比對(見 WellWishes)
//	      都不符   → 印 "\nNo effect...\n"
//	      符合但地點不是 22 / 31 → 印 "\nNo effect...\n"
//	      符合且地點對 → 印 "\nPoof!\n"、放音效,
//	                     在 (玩家X + 1, 玩家Y, 樓層) 生一匹**馬**
//
// # 兩個地點
//
//	0x16 = 22 → PAWS
//	0x1F = 31 → EMPATH ABBEY
//
// 這兩個數字是 `byte_3E0A3`(1-based 地點編號,`docs/re/03`),對回 `locations.go`
// 就是這兩座有井的城鎮。
//
// # 比對規則:比的是「大寫後的字面值」,而且是**前綴**
//
// `sub_27C98(字面值, 玩家輸入)` 先把字面值逐字轉大寫(最多 9 個字元,
// 用 `byte_738D8[c] & 2` 判小寫、`sub al,20h` 轉大寫),再 `strncmp`
// **只比字面值的長度**。所以:
//
//	輸入 "HORSEY..." → 只比前 5 個 → **相符**
//	輸入 "HORS"      → 第 5 個位元組是 NUL vs 'E' → 不符
//
// 回 0 代表相符、−1 代表不符(呼叫端 `inc eax; jg` = 相符)。
//
// ⚠ 比對用的是**大寫**字面值,而輸入常式 `sub_2B770` 收 32..122
// (含小寫)。所以小寫的願望在原版**不會**觸發 —— 這一條照原樣保留,
// 不「順手」改成不分大小寫(那就變成自創了)。

// WellLookTile 是井的 tile(Look 分派 `sub_D258` 比的那個值)。
const WellLookTile = 0xA1

// WellWishMax 是願望收得下幾個字元(原版 `sub_2B770(buf, 0Ch)`)。
const WellWishMax = 12

// WellCoin 是投一次井要花多少錢(原版 `dec word_3DFB6`)。
const WellCoin = 1

// WellWishes 是六個會生效的願望,**順序照原版的比對順序**。
//
// 五輛跑車加一匹馬 —— 1988 年的 Origin 團隊在許願井裡藏了自己的夢想車單。
var WellWishes = []string{
	"CORVETTE",
	"FERRARI",
	"LAMBORGHINI",
	"LOTUS",
	"PORSCHE",
	"HORSE",
}

// 彩蛋只在這兩個地點生效(原版 `cmp al,16h` / `cmp al,1Fh`)。
const (
	// WellLocationPaws 是 PAWS(地點 22)。
	WellLocationPaws = 0x16
	// WellLocationEmpathAbbey 是 EMPATH ABBEY(地點 31)。
	WellLocationEmpathAbbey = 0x1F
)

// WellWishMatches 照原版的規則比對玩家輸入。
//
// 回傳命中的是 `WellWishes` 的哪一個(−1 = 都不符)。
// 規則是「大寫字面值的前綴比對」,所以輸入比字面值長也算。
func WellWishMatches(wish string) int {
	for i, w := range WellWishes {
		if len(wish) < len(w) {
			continue
		}
		if wish[:len(w)] == w {
			return i
		}
	}
	return -1
}

// WellGrantsHere 回報這個地點會不會兌現願望。
func WellGrantsHere(location int) bool {
	return location == WellLocationPaws || location == WellLocationEmpathAbbey
}
