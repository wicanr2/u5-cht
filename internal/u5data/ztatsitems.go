package u5data

// Ztats 的 Items 頁:38 筆散落的道具被搬進一個連續陣列(原版 `sub_1E8D4`)
//
// 推導見 `docs/re/94`。★ 這一頁的名字表(`off_408F4`,38 個指標)是從
// `WORRIORS.EXP` 的**一手位元組**跟指標讀出來的,不是從 `.asm` 的註解抄的
//(重複字串在 IDA 的列表裡沒有註解,靠註解會漏)。

// ZtatsItemCount 是這一頁的筆數(原版 `push 26h`)。
const ZtatsItemCount = 38

// ZtatsItemNames 是那 38 個**縮寫**名(原版 `off_408F4`)。
//
// ⚠⚠ 三處看起來像 bug 而其實是原版:
//
//  1. 八瓶藥水全都叫 `"!"` —— **顏色看不出來**。
//  2. 八顆月石叫 `"(0"`..`"(7"` —— 明顯是沒寫完的佔位字串。
//  3. 名字都被縮到能塞進面板:`Magic Crpt`、`Shard/Falsehd`、`Leath Helm`。
//
// 照抄(`CLAUDE.md §3.0`)。譯文由 `internal/i18n` 覆蓋層處理,
// 要在那裡把 `"!"` 譯成看得懂的東西是**在地化決定**,不是改機制。
var ZtatsItemNames = [ZtatsItemCount]string{
	// 0..7 卷軸(`byte_3E030`)—— 名字是咒語縮寫
	"*VL", "*RH", "*IS", "*IA", "*IQW", "*KXC", "*IMC", "*AT",
	// 8..15 藥水(`byte_3E038`)
	"!", "!", "!", "!", "!", "!", "!", "!",
	// 16..20 五件單品
	"Magic Crpt", "Skull Keys", "Amulet", "Crown", "Sceptre",
	// 21..28 月石(`byte_3E050`)
	"(0", "(1", "(2", "(3", "(4", "(5", "(6", "(7",
	// 29..31 三塊寶石碎片
	"Shard/Falsehd", "Shard/Hatred", "Shard/Cowrdce",
	// 32..37 航海與任務道具
	"Spyglass", "HMS Cape Plan", "Sextant", "Pocket Watch",
	"Black Badge", "Wooden Box",
}

// ZtatsFlagShown 是旗標型道具在這一頁顯示的數量(原版填 `0FFh`)。
//
// ★ 月石與 HMS 海圖不是計數而是有/無,原版把「有」搬成 **0xFF**
// ⇒ 畫面上那一行的數字是 **255**。看起來像壞掉,是原版行為。
const ZtatsFlagShown = 0xFF

// ZtatsItemSlot 是那 38 筆各自的來源說明(原版 `sub_1E8D4` 逐行搬運)。
//
//	 0..7   byte_3E030[i]                        卷軸
//	 8..15  byte_3E038[i]                        藥水
//	16      byte_3DFBC                           魔毯
//	17      byte_3DFBD                           骷髏鑰匙
//	18      byte_3DFBF                           護符
//	19      byte_3DFC0                           王冠
//	20      byte_3DFC1                           權杖
//	21..28  byte_3E050[i] == 0xFF ? 0xFF : 0      ★ 月石:**還在身上**才顯示
//	29..31  byte_3DFC4[i]                        三塊碎片
//	32      byte_3DFC8                           望遠鏡(直接搬位元組)
//	33      byte_3DFC9 != 0 ? 0xFF : 0            HMS 海圖(★ 唯一做布林轉換的一筆)
//	34..37  byte_3DFCA..CD                       六分儀 / 懷錶 / 黑徽章 / 木盒(直接搬)
//
// ⚠ 月石那一條的極性容易看反:`cmp …, 0FFh; jz` **跳到**設 0xFF 的那一支,
// 所以「等於 0xFF(還在身上)」才會出現在清單上 —— 埋下去的就從清單消失。
//
// 實際的搬運在 `internal/game/ztatspages.go` —— 那些旗標的欄位在 `game.State` 上,
// 不在 `u5data.Save` 上(引擎把它們攤成布林了)。
const (
	ZtatsScrollBase    = 0
	ZtatsPotionBase    = 8
	ZtatsCarpetSlot    = 16
	ZtatsOddKeySlot    = 17
	ZtatsAmuletSlot    = 18
	ZtatsCrownSlot     = 19
	ZtatsSceptreSlot   = 20
	ZtatsMoonstoneBase = 21
	ZtatsShardBase     = 29
	ZtatsSpyglassSlot  = 32
	ZtatsPlansSlot     = 33
	ZtatsSextantSlot   = 34
	ZtatsWatchSlot     = 35
	ZtatsBadgeSlot     = 36
	ZtatsBoxSlot       = 37
)
