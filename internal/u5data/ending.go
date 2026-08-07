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
	MsgEndOrb       = 8
	MsgEndFollow    = 9
	MsgEndPullUpAChair = 10 // 沒帶盒子的結局
)

// EndingFinale 是打開盒子之後不列顛王說的那幾句(記錄 3..9)。
var EndingFinale = []int{
	MsgEndOpensBox, MsgEndArtifact, MsgEndOurWorld,
	MsgEndFreeUs, MsgEndOlder, MsgEndOrb, MsgEndFollow,
}

// SaveSandalwoodBoxOffset 是「有沒有那只檀香木盒」(`byte_3DFCD`)。
//
// 位移跟著讀取序列:0x0214 起是 `byte_3DFC8`..`byte_3DFCD` 六個單位元組,
// 之後 0x021A 就是既有且已驗過的 `SaveItemsOffset` —— 兩端都釘住了,
// 中間這六格沒有算錯的空間。
const SaveSandalwoodBoxOffset = 0x0219
