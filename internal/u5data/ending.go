package u5data

// 結局(原版 `sub_135FC`)
//
// 走到王座廳、把不列顛王救出來之後的那一幕。
//
// ⚠ 這一幕**有兩個結局**,分岔點是「汝帶了那只檀香木盒沒有」——
// 帶了才看得到真結局(月之球、回到舊世界);沒帶就只有一句
// 「那麼,搬張椅子坐下吧。我們得在這兒待上一陣子了。」然後遊戲就停在那裡。
// 只做真結局的話,玩家會以為盒子是可有可無的道具。

// EndingMessage 是 `ENDMSG.DAT` 的記錄序號。
//
// ★ 與黑棘的審問一樣,`sub_135FC` 是用 `sub_2C740(…, byte_54700, 0x3E8, 0)`
// **從檔頭**把整份 `ENDMSG.DAT` 載進同一個緩衝區,所以 `byte_54700 + n`
// 就是檔頭起算第 n 個位元組。十個「寫死的字串指標」逐一落在記錄開頭:
// 0x000 / 0x021 / 0x049 / 0x0AB / 0x0D3 / 0x128 / 0x167 / 0x1C9 / 0x211 / 0x24B
// 對到第 0..9 筆,而「沒帶盒子」那句 `byte_549D5`(0x2D5)是第 10 筆。
const (
	MsgEndWellMet   = 0  // 「不列顛王說道:『幸會了,」+ 聖者名 + `!"`
	MsgEndAskBox    = 1  // 「汝可帶來了吾的盒子?」
	MsgEndAskBox2   = 2  // 「那只檀香木盒……汝可帶來了?」(答 N 之後再問一次)
	MsgEndOpensBox  = 3  // 「不列顛王小心翼翼地打開盒子……」
	MsgEndArtifact  = 4  // 以下六句是真結局
	MsgEndOurWorld  = 5
	MsgEndFreeUs    = 6
	MsgEndOlder     = 7
	MsgEndAmulet       = 8
	MsgEndFollow    = 9
	MsgEndPullUpAChair = 10 // 沒帶盒子的結局
)

// EndingFinale 是打開盒子之後不列顛王說的那幾句(記錄 3..9)。
var EndingFinale = []int{
	MsgEndOpensBox, MsgEndArtifact, MsgEndOurWorld,
	MsgEndFreeUs, MsgEndOlder, MsgEndAmulet, MsgEndFollow,
}

// SaveSandalwoodBoxOffset 是「有沒有那只檀香木盒」(`byte_3DFCD`)。
//
// 位移跟著讀取序列:0x0214 起是 `byte_3DFC8`..`byte_3DFCD` 六個單位元組,
// 之後 0x021A 就是既有且已驗過的 `SaveItemsOffset` —— 兩端都釘住了,
// 中間這六格沒有算錯的空間。
const SaveSandalwoodBoxOffset = 0x0219

// 製作名單 / 頒獎狀(原版 `sub_13258`)
//
// 真結局播完之後的那一頁 —— 用羊皮紙腔宣告「某年某月某日,某某聖者救了
// 不列顛王」,然後結算汝花了多久。
//
// ★ 起算日 **139 年 4 月 5 日** 不是猜的,兩個獨立來源對得上:
//
//	程式碼  `sub_13258` 直接減:年 −0x8B(139)、月 −4、日 −5
//	資料    `INIT.GAM` 的年 / 月 / 日欄位就是 139 / 4 / 5
//
// 借位用的是不列顛尼亞曆:**每月 28 天、每年 13 個月**(與 `game.Clock` 一致)。
const (
	EpochYear  = 139
	EpochMonth = 4
	EpochDay   = 5
)

// Elapsed 是從開局到現在過了多久(原版 `sub_13258` 尾段)。
//
// 借位順序照原版:先補日、再補月。
func Elapsed(year, month, day int) (years, months, days int) {
	years = year - EpochYear
	months = month - EpochMonth
	days = day - EpochDay
	if days < 0 {
		days += CalendarDaysPerMonth
		months--
	}
	if months < 0 {
		months += CalendarMonthsPerYear
		years--
	}
	return
}

// 不列顛尼亞曆(與 `game.Clock` 同一組值,放這裡是為了 `Elapsed` 不必反向依賴)。
const (
	CalendarDaysPerMonth  = 28
	CalendarMonthsPerYear = 13
)

// CreditsRunes 是名單中間那兩行符文(原版 `aEQueOfEAvatar` / `aIsForever`)。
//
// 原始位元組是 `[E@QUE_@OF@[E@AVATAR` 與 `IS@FOREVER` —— 那是**符文字型**的
// 編碼,不是亂碼:`@` 是空白、`[` 是 TH 合字、`_` 是 ST 合字。
// 拆開來就是 `THE QUEST OF THE AVATAR / IS FOREVER`。
//
// **[HARD] 這兩行維持原樣不譯** —— 它們在原版是用符文字型畫出來的圖形,
// 譯成中文就沒有那個效果了。中譯放在下面一行當註解式的補述。
var CreditsRunes = [2]string{"THE QUEST OF THE AVATAR", "IS FOREVER"}

// 結局的觸發條件(原版 `sub_161E4`)
//
// 全遊戲只有兩處呼叫結局那一幕(`sub_135FC`),兩處都守著 `byte_3E0B0 == 'M'`,
// 而寫入 'M' 的只有 `sub_161E4` 這一支:
//
//	單位 = dword_3EF50[byte_3E0AE]        // 這一回合行動的那個
//	if (旗標 == 0)        return          // 空槽
//	if (旗標 & 0x20)      return          // 已死
//	if (戰場 Y != 2)      return
//	if ((byte_3F854[戰場 X] & 0xFC) != 0x3C) return
//	byte_3E0B0 = 'M';  印 "<某某> is absorbed!"
//
// ★ **`byte_3F854` 是戰場暫存的第 1 列,不是腳下那一格。**
// `byte_3F844` 是 11×11、**列距 16** 的戰場 / 石室暫存(見 miscmap.go),
// 而 `0x3F854 = 0x3F844 + 16` —— 也就是 `[y=1][x]`。單位在第 2 列,
// 檢查的是**它正北那一格**。
//
// ★ **0x3C..0x3F 是城堡的外牆與城門。** 切出來看是白色城垛與黃色拱門;
// 在世界地圖上剛好只出現七格,全部圍著兩座城堡的入口:
//
//	(87,106) (85,107) (86,107) (87,107)   ← 不列顛王的城堡(地點 17)
//	(197,244) (195,245) (197,245)          ← 黑棘的宮殿(地點 18)
const (
	// AbsorbRow 是單位必須站的戰場列。
	AbsorbRow = 2
	// AbsorbTileGroup 是「會吸收人」的地形群(`& 0xFC == 0x3C`)。
	AbsorbTileGroup = 0x3C
)

// OverlayEmpty 是疊圖層的空值(原版 `sub_29D64` / `sub_C778` 用 `rep stosd` 填的 0xFF)。
//
// ★ `0xFF & 0xFC = 0xFC`,剛好不等於 0x3C —— 空格不會誤觸吸收判定。
const OverlayEmpty = 0xFF

// AbsorbTile 回報疊圖層的這個位元組會不會把站在它南邊的人吸進去。
func AbsorbTile(t byte) bool { return t&0xFC == AbsorbTileGroup }
