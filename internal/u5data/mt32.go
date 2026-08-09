package u5data

// 第二套音樂:DOS `upgrade/` 的 16 首 `.XMI`(Roland MT-32)
//
// 推導見 `docs/formats/13`(XMI 格式)與本檔 §「怎麼對上曲號」。
// FM Towns 那套在 `bgm.go`(`U5_BGM.TBL` → `M*.EUP`)。
//
// ## 為什麼需要這張表
//
// `U5_BGM.TBL` 給的是**曲號 → `.EUP` 檔名**,而 upgrade 那套的檔名是**曲名**
// (`U5THEME.XMI`、`BRITLAND.XMI`…)。兩邊沒有共同的鍵:
//
//	曲號 0  →  M1.EUP     ⇄  ???.XMI
//
// 而「哪一首 XMI 配哪個場景」**不在任何資料檔裡** —— upgrade 是把它改寫進
// `ULTIMA.EXE` 的程式碼。⇒ 不能讀表拿到,只能另外找證據。
//
// ## 怎麼對上曲號:比旋律
//
// 兩套是**同一批曲子的不同編曲** ⇒ 拿音高**差分序列**(不受移調影響)的
// 最長共同子串配對(`tools/match_songs.py`,可重跑)。結果:
//
//	15 首 `.EUP` × 16 首 XMI = 240 對,**雙向**取最佳
//	14 組雙向一致,分數 47..423(次佳一律 < 1/2)
//	剩下的 REUNION ⇄ M14 由音數 128/126、聲道 6/6 收尾(見下)
//	AMIGA 對任何 `.EUP` 與任何其他 XMI 都 ≤ 10 ⇒ **它沒有曲號**
//
// ★★ 而這張表可以**三方交叉驗證**:`docs/re/87` 是從 32 個 `sub_3181C`
// 呼叫點逆出曲號的用途,與 upgrade 作者自己寫的曲名獨立吻合 ——
//
//	曲 1  Britannic Lands         ⇄ `sub_86C` 回地表
//	曲 2  Cap'n Johne's Hornpipe  ⇄ `sub_16F08` 上船
//	曲 3  Engagement and Melee    ⇄ `sub_A9EC` 進戰鬥
//	曲 9  Halls of Doom           ⇄ `sub_2D564` 進地牢
//	曲 0Ah Worlds Below           ⇄ `sub_86C` 進幽冥界
//
// 五條獨立吻合 ⇒ 旋律配對、「曲號 = `U5_BGM.TBL` 的列號」、逆出來的曲號語意
// 三者互為佐證。這比其中任何一條單獨的證據都強。
//
// ## ⚠ 兩個差點判錯的地方
//
//  1. **分數要對照曲子長度讀。** `M14.EUP` 每聲道只有 ~22 個音,所以它拿 21
//     已經是整條旋律都一樣;用「≥12 且是次佳兩倍」的絕對門檻會判成沒對上。
//  2. **同一句樂句會出現在兩首曲子裡。** `REUNION`(短的重奏)與 `RULEBRIT`
//     共用開頭那句,所以 `REUNION×M14`、`REUNION×M152`、`RULEBRIT×M14`
//     **三個組合都拿 21** —— 旋律這一個訊號分不開它們。
//     分開它們的是規模:REUNION 128 音/6 聲道 ⇄ M14 126 音/6 聲道;
//     RULEBRIT 460 音/8 聲道 ⇄ M152 397 音/6 聲道。

// MT32Track 是 upgrade 那一版的一首曲子。
type MT32Track struct {
	// Base 是 `.XMI` 的基名,也是渲染出來的 ogg 檔名。
	//
	// ⚠ **大小寫不可靠**:光碟上 15 個是大寫、`trntlla.xmi` 是小寫。
	// 查檔案一律走 `audio` 那邊的大小寫無關查表 —— 這個坑已經咬過一次
	// (`*.XMI` 這個 glob 漏掉了 `trntlla.xmi`,害「15 首」寫錯成事實)。
	Base string
	// Title 是 `upgrade/Files.txt` 裡作者標的曲名。
	Title string
}

// MT32Tracks 是曲號 → upgrade 的曲子。下標就是曲號(對齊 `U5_BGM.TBL` 的列號)。
//
// 第三欄的 `.EUP` 是旋律配對的對象,留著讓人能回頭核對 `tools/match_songs.py`。
var MT32Tracks = [BGMSongCount]MT32Track{
	0:  {"U5THEME", "Ultima V Theme"},           // ⇄ M1  旋律 119
	1:  {"BRITLAND", "Britannic Lands"},         // ⇄ M2  旋律 185
	2:  {"HORNPIPE", "Cap'n Johne's Hornpipe"},  // ⇄ M3  旋律  91
	3:  {"ENGGMNT", "Engagement and Melee"},     // ⇄ M4  旋律  98
	4:  {"STONES", "Stones"},                    // ⇄ M5  旋律 423
	5:  {"GREYSON", "Greyson's Tale"},           // ⇄ M6  旋律  64
	6:  {"FANFARE", "Fanfare for the Virtuous"}, // ⇄ M7  旋律 185
	7:  {"MONARCH", "The Missing Monarch"},      // ⇄ M8  旋律  47
	8:  {"trntlla", "Villager Tarantella"},      // ⇄ M92 旋律 108(★ 檔名小寫)
	9:  {"HALLS", "Halls of Doom"},              // ⇄ M10 旋律  52
	10: {"WRLDBLW", "Worlds Below"},             // ⇄ M11 旋律 199
	11: {"BLCKTHRN", "Lord Blackthorn"},         // ⇄ M12 旋律 151
	12: {"LADYNAN", "Dream of Lady Nan"},        // ⇄ M13 旋律 164
	13: {"REUNION", "Joyous Reunion"},           // ⇄ M14 音數 128/126(旋律平手)
	14: {"RULEBRIT", "Rule Britannia"},          // ⇄ M152 旋律  53
}

// MT32ExtraTracks 是 upgrade 有、FM Towns 那套**沒有**的曲子。
//
// ★ `AMIGA`(Amiga Theme,383 秒)對 15 首 `.EUP` 的最高分是 10、
// 對其他 15 首 XMI 的最高分也是 10 ⇒ 它既不是某首的編曲版,也不是改編 ——
// 是 upgrade 額外附的一首。**沒有曲號,所以遊戲流程不會播它。**
//
// ⬜ 原版 upgrade 有沒有在某處播它未知(要逆 patch 過的 `ULTIMA.EXE`)。
// 引擎目前不播 —— 沒證據就不接(`CLAUDE.md §3.0`)。
var MT32ExtraTracks = []MT32Track{
	{"AMIGA", "Amiga Theme"},
}
