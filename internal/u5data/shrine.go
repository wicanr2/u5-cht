package u5data

// 八座聖壇(`sub_1D394`,2026-08-07)
//
// 資料全部來自原版的並列表:
//
//	byte_411FC[8]  X            byte_41204[8]  Y
//	off_411DC[8]   真言          off_55FEC[8]   美德名的前四個字母
//	byte_5606C[8]  力量 +1       byte_56074[8]  敏捷 +1
//	byte_5607C[8]  智力 +1       dword_5602C[8] 寶典答案在 MISCMSG 裡的位移
//
// ★ `dword_5602C` 的值 `[0, 27, 57, 95, 123, 159, 185, 214]` 與
// `MISCMSG.DAT` 第 12..19 筆**相對第 12 筆的位移完全相同** ——
// 所以「這座聖壇要汝去寶典學什麼」就是第 `12 + 美德` 筆。
// 兩份獨立資料對上,不是猜的。

// VirtueCount 是八德。
const VirtueCount = 8

// 八德的順序 —— 全遊戲的旗標位元、真言表、聖壇座標都照這個順序。
const (
	VirtueHonesty = iota
	VirtueCompassion
	VirtueValor
	VirtueJustice
	VirtueSacrifice
	VirtueHonor
	VirtueSpirituality
	VirtueHumility
)

// Shrine 是一座聖壇。
type Shrine struct {
	// X, Y 是世界座標。
	X, Y int
	// Mantra 是真言。**[HARD] 這是玩家要打出來的字,永遠維持英文。**
	Mantra string
	// Prefix 是美德名的前四個字母(小寫)——「汝欲冥想何種美德?」比對用。
	Prefix string
	// Str / Dex / Int 是完成試煉後加哪一項。
	Str, Dex, Int bool
	// Name 是英文美德名,NameZH 是中文。
	Name, NameZH string
}

// Shrines 是八座聖壇。
//
// ⚠ **靈性聖壇的座標是 (0,0)** —— 那不是錯,是原版就沒有把它放進座標表。
// `sub_1D394` 掃完八筆沒中的話 `if (i == 8) i = 6`,也就是
// **站在任何一個不在表上的聖壇,一律當成靈性**。U5 的靈性聖壇在幽冥界,
// 靠這個 fallback 生效。照抄。
//
// ⚠ **`hone`(誠實)與 `hono`(榮譽)只差第四個字母** —— 這就是原版比四個字母
// 而不是三個的原因。縮成三個字母會讓兩座聖壇分不出來。
var Shrines = [VirtueCount]Shrine{
	{X: 233, Y: 66, Mantra: "Ahm", Prefix: "hone", Int: true, Name: "Honesty", NameZH: "誠實"},
	{X: 128, Y: 92, Mantra: "Mu", Prefix: "comp", Dex: true, Name: "Compassion", NameZH: "慈悲"},
	{X: 36, Y: 229, Mantra: "Ra", Prefix: "valo", Str: true, Name: "Valor", NameZH: "勇氣"},
	{X: 73, Y: 11, Mantra: "Beh", Prefix: "just", Dex: true, Int: true, Name: "Justice", NameZH: "正義"},
	{X: 205, Y: 45, Mantra: "Cah", Prefix: "sacr", Str: true, Dex: true, Name: "Sacrifice", NameZH: "犧牲"},
	{X: 81, Y: 207, Mantra: "Summ", Prefix: "hono", Str: true, Int: true, Name: "Honor", NameZH: "榮譽"},
	{X: 0, Y: 0, Mantra: "Om", Prefix: "spir", Str: true, Dex: true, Int: true, Name: "Spirituality", NameZH: "靈性"},
	{X: 231, Y: 216, Mantra: "Lum", Prefix: "humi", Name: "Humility", NameZH: "謙遜"},
}

// ShrineAt 回傳站在 (x, y) 是哪一座聖壇。
//
// 找不到就回**靈性**(6)—— 原版的 `if (i == 8) i = 6`,見 `Shrines` 的說明。
func ShrineAt(x, y int) int {
	for i := range Shrines {
		if Shrines[i].X == x && Shrines[i].Y == y {
			return i
		}
	}
	return VirtueSpirituality
}

// 聖壇的獎懲數值(`sub_1D394` 尾段)
const (
	// ShrineQuestKarma 是完成試煉加的業報。
	ShrineQuestKarma = 3
	// ShrineHumilityBonus 是**謙遜**額外再加的業報(`cmp edi, 7`)。
	//
	// ⚠ 只有第 7 座有這個加成。它是八德裡唯一三圍都不加的一座 ——
	// 換成業報補回來,而且是雙倍。
	ShrineHumilityBonus = 3
	// StatMax 是三圍上限(`cmp byte_3DDC0, 1Eh`)。
	//
	// (業報上限 `KarmaMax` 在 `talkscript.go` —— 對話與聖壇共用同一個 0x63。)
	StatMax = 30
	// ShrineGoldPerUnit 是「幾百重黃金」的一重是多少金幣。
	//
	// 算式是 `n × 5 × 5 × 4 = n × 100`(原版拆成三個 `lea` 做乘法)。
	ShrineGoldPerUnit = 100
	// ShrineMaxOffer 是一次最多獻幾重 —— 原版只讀**一個** '0'..'9' 的按鍵。
	ShrineMaxOffer = 9
)

// ShrineQuestRecord 是「去寶典學什麼」那句話在 `MISCMSG.DAT` 裡的記錄序號。
func ShrineQuestRecord(virtue int) int { return 12 + virtue }

// MISCMSG 裡與聖壇有關的幾筆(序號,不是位移)。
const (
	MsgShrineApproach = 45 // 「汝走近那座靜謐的聖壇……」
	MsgShrineKneel    = 28 // 「……汝跪倒在聖壇之前。」
	MsgShrineWhich    = 29 // 「汝欲冥想何種美德?」
	MsgShrineUnfocus  = 30 // 「汝之思緒散亂無所凝聚。」
	MsgShrineQuestOn  = 31 // 「聖壇開口,一項試煉就此降下!」
	MsgShrineQuestIs  = 32 // 「今汝之神聖試煉,是前往寶典之處,習得」
	MsgShrineReturn   = 33 // 「試煉既成,再來此地!」
	MsgShrineOffer    = 34 // 「欲獻上多少百重的黃金?」
	MsgShrineNoGold   = 35 // 「汝沒有那麼多金子!」
	MsgShrineWellDone = 36 // 「做得好!」
)
