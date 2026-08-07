package u5data

// 暗影君主的一生:遊走 → 玷污聖壇 → 被碎片消滅
//
// 三支函式串起整條線,全部在 FM Towns `WORRIORS.EXP` 裡:
//
//	sub_29304  每天午夜把活著的暗影君主重新分派到八德城市之一
//	sub_C414   黑棘的審問 —— 招了就玷污一座聖壇(`sub_C318`)
//	sub_1A38C  在對應的聖火前把寶石碎片舉起來 → 消滅
//
// 召喚(`sub_17A14`,Yell 一個名字)在 `wordofpower.go`。

// 遊走(`sub_29304` 的午夜段)

// ShadowlordCityMin / ShadowlordCityMax 是暗影君主會盤據的地點編號範圍。
//
// 原版 `sub_28E14(1, 8)` —— 就是八德城市(月光城…新馬金夏),不含城堡與村落。
const (
	ShadowlordCityMin = 1
	ShadowlordCityMax = 8
)

// 聖火與寶石碎片(`sub_1A38C`)

// Flame 是一團聖火的所在。
type Flame struct {
	X, Y     int
	Location int
	Floor    int
	// Shard 是要投進來的那塊碎片叫什麼,Name 是這團火叫什麼。
	//
	// **[HARD] 兩個都維持英文** —— 訊息裡照原版並排,譯文另走 i18n。
	Shard, Name string
	// ShardZH / NameZH 是譯名。
	ShardZH, NameZH string
}

// Flames 是三團聖火。索引與暗影君主、碎片共用。
//
// ★ 座標不是從遊戲印象打的。`sub_1A38C` 把四個並列表放在字串
// `aNoNoticeableEf`(0x55E28)後面,四段各 3 B:
//
//	+0x1C  X        0F 0F 0F
//	+0x1F  Y        09 03 10
//	+0x22  地點     1E 1F 20   = 30 / 31 / 32
//	+0x25  樓層     02 01 FF   = 2 / 1 / −1
//
// 而 30 / 31 / 32 正是 `u5data.Locations` 裡的學術之城 / 共感修道院 / 巨蛇要塞
// —— 與 `ShadowlordKeeps`(召喚地點)同一組,兩邊獨立對上。
var Flames = [ShadowlordCount]Flame{
	{X: 15, Y: 9, Location: 30, Floor: 2,
		Shard: "Falsehood", Name: "Truth", ShardZH: "虛偽", NameZH: "真理"},
	{X: 15, Y: 3, Location: 31, Floor: 1,
		Shard: "Hatred", Name: "Love", ShardZH: "憎恨", NameZH: "愛"},
	{X: 15, Y: 16, Location: 32, Floor: -1,
		Shard: "Cowardice", Name: "Courage", ShardZH: "怯懦", NameZH: "勇氣"},
}

// ShadowlordDoomBit 是消滅第 i 位之後 OR 進 `dword_3E3DC` 的位元(`byte_55E50`)。
//
// 值是 `02 04 08` —— 三位各一個,第 4 個位元組是 0(表尾)。
var ShadowlordDoomBit = [ShadowlordCount]byte{0x02, 0x04, 0x08}

// SaveShardsOffset 起 4 B(`byte_3DFC4`):三塊寶石碎片各持有與否,第 4 B 未用。
//
// 位移跟著 `sub_27D24` 的讀取序列從 0x0208(火把)累加:
// 0x0209..0x020F 是 `byte_3DFBB`..`byte_3DFC1` 七個單位元組,
// 接著 4 B 的 `byte_3DFC4`,再六個單位元組之後就是 `SaveItemsOffset`(0x021A)——
// **後面那個位移是既有的、已驗過的**,所以這一段沒有算錯的空間。
const SaveShardsOffset = 0x0210

// ShardCount 是存檔裡那一段的長度(3 塊碎片 + 1 B 未用)。
const ShardCount = 4

// 黑棘的審問(`sub_C414` → `sub_C318`)

// BlackthornRounds 是黑棘問幾次(`cmp esi, 4`)。
const BlackthornRounds = 4

// BlackthornKarmaPenalty 是招供扣的業報(`sub_2BBFC(&byte_3E098, 5)`)。
const BlackthornKarmaPenalty = 5

// BlackthornLocation 是黑棘宮殿的地點編號(`cmp byte_3E0A3, 12h`)。
//
// = 18,座標 (196,245) —— 與 `docs/re/03` 從進場地形表得到的
// 「61 = 黑棘宮殿」對得上。
const BlackthornLocation = 18

// 審問結束後被丟到哪(`sub_C414` 尾段)。
const (
	BlackthornCellX     = 10
	BlackthornCellY     = 7
	BlackthornCellFloor = -1
)

// 審問用的 `MISCMSG.DAT` 記錄序號。
//
// ★ 這一段的緩衝區與寶典**不是同一個載入起點**:`sub_C414` 呼叫的是
// `sub_2C740("MISCMSG.DAT", byte_54700, 0x3E8, 0)` —— **位移 0**,
// 所以 `byte_54700 + n` 直接就是檔頭起算的第 n 個位元組。
// (寶典那邊是從 0x3AB 載入,見 `CodexAnswerRecord`。同一個緩衝區、
// 兩個不同的起點 —— 只記一個會把另一邊的記錄全部算錯。)
const (
	MsgBlackthornAsk1  = 0 // 「神秘聖壇的真言是什麼 —— 」+ 美德名 + `?"`
	MsgBlackthornAsk2  = 1 // 「現在告訴吾,真言是什麼 —— 」
	MsgBlackthornAsk3  = 2 // 「抵抗是徒勞的!……」
	MsgBlackthornAsk4  = 3 // 「吾對汝的耐心已經耗盡!現在就把真言說出來!」
	MsgBlackthornBlade = 4 // 「黑棘手一揮,擺錘上的刀落下!」
	MsgBlackthornMercy = 5 // 「吾謝過汝……賜汝的同伴一個痛快。」
	MsgBlackthornPaid  = 6 // 「……汝的朋友為汝的背叛付出了代價!」
	MsgBlackthornTaunt = 7 // 「別犯下嘲笑吾的錯,蠢貨!」
	MsgBlackthornSand  = 8 // 「吾會一直問到沙漏流盡。然後……」
	MsgBlackthornTruth = 9 // 「吾在汝身上感到真實。這將以汝的性命為報!」
	MsgBlackthornLies  = 10
	MsgBlackthornWait  = 11
)

// BlackthornQuestion 是第 n 輪的問句記錄序號。
//
// ⚠ **前三輪句尾要接美德名**(原版 `off_411BC[虛擬索引]` 之後再補 `?"`),
// 第四輪不接 —— 它本身就是完整的一句。
func BlackthornQuestion(round int) int {
	if round < 0 || round >= BlackthornRounds {
		return MsgBlackthornAsk4
	}
	return MsgBlackthornAsk1 + round
}

// BlackthornQuestionNamesVirtue 回報第 n 輪的問句要不要接美德名。
func BlackthornQuestionNamesVirtue(round int) bool { return round < BlackthornRounds-1 }

// CharUrn 是角色紀錄裡「這個人被黑棘處決了」的欄位(位移 31)。
//
// ★ 這一格的語意原本是空白的(`docs/re/27` §5 記著「還沒追」)。
// 追出來了:`sub_C13C` 把被處決的同伴搬到名冊第 15 格,然後
// `mov byte_3DFB3, 7Fh` —— 0x3DFB3 = 0x3DDB4 + 15×32 + 31。
// 而 `sub_1DA10` 進寶典石室時掃的正是 `byte_3DDD3[i*32] == 0x7F`,
// 掃到就印「Thou dost see an urn marked: <名字>」。
//
// 也就是說:**位移 31 的 0x7F = 這個人的骨灰罈擺在寶典之前。**
const (
	CharUrn     = 31
	CharUrnMark = 0x7F
)

// 暗影君主盤據一座城時會發生什麼(`sub_48C` + `sub_15C4`,由 `sub_1678` 在進場景時呼叫)

// ShadowlordTownX 是暗影君主在城裡現身的 X(`sub_48C` 的 `mov dl, 0Fh`)。
const ShadowlordTownX = 15

// ShadowlordTownY 是牠在各地點現身的 Y(`byte_3EF1B[地點]`)。
//
// ★ 只有索引 1..8 非零 —— 那正是八座八德城市,與 `sub_29304` 分派的範圍
//(`random(1,8)`)完全吻合。兩份獨立資料對上,不必另找 oracle。
var ShadowlordTownY = [33]byte{
	0, 4, 9, 15, 8, 17, 10, 11, 10,
}

// ShadowlordAir 是進到有暗影君主的地方時那句「空氣中瀰漫著……」的第二段
//(`off_4FC90`)。**這是敘述,不是玩家要輸入的字**,所以譯成中文。
var ShadowlordAir = [ShadowlordCount]string{"虛偽", "憎恨", "怯懦"}

// 暗影君主對城裡居民的影響(`sub_15C4` 依 `byte_3E16A` 分派)
//
//	0 FAULINEI  虛偽 → 這裡**什麼都不做**(它的手段在別處)
//	1 ASTAROTH  憎恨 → 居民變**敵對**,對話號碼改成 0xFE
//	2 NOSFENTOR 怯懦 → 居民**逃跑**,對話號碼改成 0xFD
//
// ⚠ 三個各有各的,不能一視同仁 —— 憎恨讓人撲上來、怯懦讓人跑光,
// 那正是玩家分辨「這座城被誰佔了」的方式。
const (
	ShadowlordHateDialogue  = 0xFE // ASTAROTH 改成的對話號碼
	ShadowlordFearDialogue  = 0xFD // NOSFENTOR 改成的對話號碼
)

// StonegateLocation 是石門(地點 29)——三位暗影君主的出身地。
//
// 進去時會為**每一位還沒被消滅的**印一次「空氣中瀰漫著……」(`sub_1678`)。
const StonegateLocation = 29

// ShadowlordAffectsNPC 回報這名 NPC 會不會被城裡的暗影君主影響(`sub_1568`)。
//
// 三個條件:排程時刻**不全為 0**(是有作息的居民)、生物編號在 0x40..0x73、
// 再擲一次硬幣(一半機率)。
//
// ⚠⚠ **原版這裡有一個 bug,而且照抄了。** 那句
// `movzx eax, byte_3EDB0[esi]` 用的是 `esi` —— 但 `esi` 在上面那個
// `for (esi = 0; esi < 4; esi++)` 迴圈結束時是 **4**,不是傳進來的 `edi`。
// 也就是說**生物編號那一關檢查的永遠是第 4 號槽的居民**,與正在判的那個無關。
//
// 照抄。與末日位元那個共用儲存同理:自己「修好」它會讓行為與原版不同,
// 而這裡的差別看得出來 —— 4 號槽剛好是隻動物的城,整城就沒人會被影響。
func ShadowlordAffectsNPC(scheduled bool, slot4Creature byte, coin int) bool {
	if !scheduled {
		return false
	}
	if slot4Creature < 0x40 || slot4Creature >= 0x74 {
		return false
	}
	return coin == 0
}
