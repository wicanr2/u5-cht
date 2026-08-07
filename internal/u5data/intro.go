package u5data

// 開場動畫的頁表(從 `sub_584B` 的並列表 dump 出來,2026-08-07)
//
// 開場是 **21 頁**(`cmp edi, 15h`)。每一頁由六張並列表決定:
//
//	byte_54280[頁]  用哪一個 STORY*.16(0..5)
//	byte_54268[頁]  形狀編號
//	byte_542C8[頁]  頁的種類 —— >= 4 時**多畫一張**,形狀是 `種類 × 2 − 5`
//	dword_541B4[頁] 這一頁的文字在 `STORY.DAT` 裡的**檔案位移**
//	byte_54298/542B0  FM Towns 的擺放座標
//
// ★ **`off_41BE8` 存的是 DOS 的檔名**(`STORY1.16`…`STORY6.16`)——
// FM Towns 把遊戲中的美術換成 `.PNL` / `.TIF`,**唯獨開場插圖沿用 DOS 那六個檔**。
// 所以這張頁表的「哪一頁配哪一張圖」可以直接用,不是只適用於 FM Towns。
//
// ★ **文字位移全部對得上 `STORY.DAT` 的記錄邊界。** 表裡的 20 個非零位移
// (0, 273, 971, 1424, …, 10903)與用 NUL 切出來的記錄起點**一個不差**。
// 兩份獨立資料對上,所以「頁 → 記錄」這條對應是量出來的不是猜的。
//
// ⚠ 第 6 頁的位移是 **0**,那不是「指回第一筆」——
// 它的種類是 3,走另一條路:文字是**寫死在執行檔裡的兩句**
// (`Instantly, a shimmering blue door springs up!` /
// `With heart beating rapidly, you step into it.`),不從 `STORY.DAT` 讀。
// 只看位移表會把第 6 頁的文字弄成第 0 頁的重播。
//
// ⚠ **擺放座標沒有沿用。** `byte_54298` / `byte_542B0` 是 FM Towns 640×480
// 的版面,套到本專案的 640×400 會歪。DOS 的座標在 `INTRO.OVL` 裡,還沒讀 ——
// 所以引擎自己排版(圖置中、文字在下),並在 README 標明這一點。

// IntroPageCount 是開場的頁數。
const IntroPageCount = 21

// IntroPage 是開場的一頁。
type IntroPage struct {
	// Story 是用第幾個 `STORY*.16`(0 = STORY1.16)。
	Story int
	// Shape 是主圖的形狀編號;Shape2 是第二張,−1 表示沒有。
	Shape, Shape2 int
	// Record 是文字在 `STORY.DAT` 裡的記錄序號;−1 表示文字寫死在執行檔裡。
	Record int
}

// IntroPages 是那 21 頁。
//
// 第 6 頁的 `Record` 是 −1 —— 見上面的說明。
var IntroPages = [IntroPageCount]IntroPage{
	{Story: 0, Shape: 0, Shape2: -1, Record: 0},
	{Story: 0, Shape: 1, Shape2: -1, Record: 1},
	{Story: 1, Shape: 0, Shape2: -1, Record: 2},
	{Story: 1, Shape: 1, Shape2: -1, Record: 3},
	{Story: 1, Shape: 2, Shape2: -1, Record: 4},
	{Story: 1, Shape: 2, Shape2: -1, Record: 5},
	{Story: 1, Shape: 2, Shape2: -1, Record: -1}, // 種類 3:文字寫死
	{Story: 2, Shape: 0, Shape2: -1, Record: 6},
	{Story: 2, Shape: 1, Shape2: -1, Record: 7},
	{Story: 3, Shape: 0, Shape2: -1, Record: 8},
	{Story: 3, Shape: 1, Shape2: -1, Record: 9},
	{Story: 4, Shape: 0, Shape2: -1, Record: 10},
	{Story: 4, Shape: 1, Shape2: -1, Record: 11},
	{Story: 5, Shape: 0, Shape2: -1, Record: 12},
	{Story: 5, Shape: 1, Shape2: -1, Record: 13},
	{Story: 5, Shape: 2, Shape2: 3, Record: 14}, // 種類 4 → 第二張 = 4×2−5
	{Story: 5, Shape: 6, Shape2: 5, Record: 15}, // 種類 5 → 5×2−5
	{Story: 5, Shape: 4, Shape2: 7, Record: 16}, // 種類 6 → 6×2−5
	{Story: 5, Shape: 2, Shape2: 5, Record: 17},
	{Story: 5, Shape: 6, Shape2: 7, Record: 18},
	{Story: 5, Shape: 4, Shape2: 3, Record: 19},
}

// IntroStoryFiles 是六個開場插圖檔(`off_41BE8`)。
var IntroStoryFiles = [6]string{
	"STORY1.16", "STORY2.16", "STORY3.16", "STORY4.16", "STORY5.16", "STORY6.16",
}

// IntroHardcoded 是第 6 頁那兩句寫死在執行檔裡的文字。
//
// 它們**不在 `STORY.DAT` 裡**,所以譯文的 key 另外開一組
//(`i18n` 的 `INTRO#0` / `INTRO#1`)。
var IntroHardcoded = [2]string{
	"Instantly, a shimmering blue door springs up!",
	"With heart beating rapidly, you step into it.",
}

// IntroHardcodedPage 是走寫死文字那一頁的頁號。
const IntroHardcodedPage = 6
