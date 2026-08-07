package u5data

// 創角 —— 吉普賽的八德淘汰賽(原版 `.text:000235B6` 起,問答在 `sub_23274`)
//
// # 賽制:8 → 4 → 2 → 1,一共七題
//
//	第一輪  4 題    每題淘汰一個,8 剩 4
//	第二輪  2 題    4 剩 2
//	第三輪  1 題    決出唯一存活的美德
//
// 每一輪之間會把「本輪抽過了」的旗標清掉(原版 `byte_5717C`),
// 但「已被淘汰」的旗標(`byte_57184`)一路留著 —— 兩個旗標分開,
// 併成一個就會在第二輪重新抽到已淘汰的美德。
//
// # 每一題的效果:**勝方加分,敗方出局**
//
//	byte_57164[勝方 +  0] 加到 智力
//	byte_57164[勝方 +  8] 加到 敏捷
//	byte_57164[勝方 + 16] 加到 力量
//	byte_57164[敗方 + 32] = 1        ← 淘汰
//
// ⚠ 原版是 `add`,不是覆蓋。Hex-Rays 把這三行反編譯成賦值
//(`byte_5718E = *(...)`),照著寫的話七題只有最後一題算數 ——
// 屬性會少掉六題份。真值在組語:`add [eax+2Ah], dl`(CLAUDE.md §4.4)。
//
// # 收尾
//
//	智力 → CharIntel,**而且同時寫進 CharMP**(原版 byte_3DDC2 與 byte_3DDC3
//	       是同一個值 —— 初始魔力等於智力)
//	敏捷 → CharDex
//	力量 → max(力量, 20)   ← 有下限,原版 `cmp byte_57190, 14h; ja …`
//
// 三個屬性的**起點是既有角色的值**(原版在問答前把 `byte_3DDC0..C2` 抄進
// `byte_5718E..90`),不是從 0 開始。引擎照做:從 `INIT.GAM` 的聖者讀起。

// 八德的順序與常數在 `shrine.go`(`VirtueHonesty`…`VirtueHumility`)——
// 全遊戲共用同一組,這裡不另立。

// VirtueBonus 是每個美德勝出一次加的 (智力, 敏捷, 力量)。
//
// 原版 `dword_57164` 的 24 個位元組,拆成三列八欄。
// 與系列的職業對應完全吻合,可當第二重佐證:
//
//	誠實 +2 智  → 法師      慈悲 +2 敏  → 吟遊詩人   勇氣 +2 力  → 戰士
//	正義 智+敏  → 德魯伊    犧牲 敏+力  → 工匠       榮譽 智+力  → 聖騎士
//	靈性 三者各 +1 → 遊俠    謙遜 全 0    → 牧羊人
var VirtueBonus = [VirtueCount][3]int{
	VirtueHonesty:      {2, 0, 0},
	VirtueCompassion:   {0, 2, 0},
	VirtueValor:       {0, 0, 2},
	VirtueJustice:      {1, 1, 0},
	VirtueSacrifice:    {0, 1, 1},
	VirtueHonor:       {1, 0, 1},
	VirtueSpirituality: {1, 1, 1},
	VirtueHumility:     {0, 0, 0},
}

// VirtueBonus 的三個欄位。
const (
	BonusIntel = 0
	BonusDex   = 1
	BonusStr   = 2
)

// CreateMinStrength 是力量的下限(原版 `cmp byte_57190, 14h`)。
const CreateMinStrength = 20

// questionRecord[a][b] 是「a 對上 b」那一題在 `QUESTION.DAT` 的記錄索引。
//
// 原版存的是**檔案位移**(`dword_57194[8*a+b]`,8×8 對稱、對角線為 0),
// 這裡換算成記錄索引 —— 引擎的 `TextFile` 本來就按記錄索引取字,
// 留著位移只會多一層換算,而且位移在不同語言版的 `QUESTION.DAT` 會不一樣。
//
// 28 個不重複的題目佔記錄 2..29;記錄 0 是開場、記錄 1 是問完之後的結語。
var questionRecord = [VirtueCount][VirtueCount]int{
	{0, 2, 3, 4, 5, 6, 7, 8},
	{2, 0, 9, 10, 11, 12, 13, 14},
	{3, 9, 0, 15, 16, 17, 18, 19},
	{4, 10, 15, 0, 20, 21, 22, 23},
	{5, 11, 16, 20, 0, 24, 25, 26},
	{6, 12, 17, 21, 24, 0, 27, 28},
	{7, 13, 18, 22, 25, 27, 0, 29},
	{8, 14, 19, 23, 26, 28, 29, 0},
}

// 開場與結語的記錄索引。
const (
	QuestionIntro   = 0
	QuestionClosing = 1
)

// VirtueQuestion 回傳「a 對上 b」那一題的記錄索引。同一個美德對自己回 0。
func VirtueQuestion(a, b int) int {
	if a < 0 || a >= VirtueCount || b < 0 || b >= VirtueCount {
		return 0
	}
	return questionRecord[a][b]
}

// CreateRounds 是每一輪的題數(原版寫死的 4 / 2 / 1)。
var CreateRounds = [...]int{4, 2, 1}

// CreateQuestions 是全部的題數。
const CreateQuestions = 7
