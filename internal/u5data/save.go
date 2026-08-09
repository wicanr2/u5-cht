package u5data

import (
	"encoding/binary"
	"fmt"
)

// 存檔格式(`SAVED.GAM` / `INIT.GAM`,4,192 B)
//
// 出自 FM Towns 版的讀取函式 `sub_27D24` —— 它把整份存檔**逐欄位循序**讀進一串
// 全域變數(44 次 fread + 82 次 fgetc,寫入函式 `sub_284CC` 完全對稱)。
// 因為是逐欄位而不是整塊記憶體 dump,檔案位移得把每個欄位的大小累加起來算。
//
// ⚠ FM Towns 的讀取序列加起來是 **4,226** B,比檔案多 34 B。兩個東西一拿掉就正好
// 對上 4,192,而且最後一個欄位剛好收在 EOF:
//
//   - `byte_3EDB0`(32 B)在 FM Towns 版被**連續讀了兩次**(組語確認,不是反編譯錯覺)
//   - 結尾多一個 2 B 欄位 `byte_3EE18`
//
// 兩者都是 FM Towns 移植時多出來的;隨遊戲附的 `SAVED.GAM`/`INIT.GAM` 是 DOS 版寫的
// (FM Towns 的資料檔與 DOS 版逐一同大小),所以**本套件用 4,192 的 DOS 版面**。
// 「最後一個欄位剛好收在 EOF」是這個判斷的主要證據,不是湊出來的。
//
// 前段已逐欄位對真實檔案驗證過:0x02CE 的年份讀出 139(不列顛尼亞的紀年)、
// 0x02D6 的載具是 0x1C(步行,與隊伍 tile 相符)、之後接著月 4 / 日 5 / 時 8 / 分 35。
const (
	// SaveFileSize 是存檔大小。
	SaveFileSize = 4192

	// LocationCodeMax 是 `byte_3E0A3` 的合理上限。
	//
	// ⚠⚠ **不是 `len(Locations)` = 32。** 地點表其實有 **40 筆**:前 32 筆是
	// 城鎮與城堡,後 8 筆是地牢(`sub_2D564` 掃索引 0x20..0x27,見 `dungeon.go`)。
	// 拿 32 當上限會讓**在地牢裡存不了檔** —— 而原版的 `byte_3E0A3` 在地牢裡
	// 就是 0x21..0x28。
	LocationCodeMax = DungeonLocationBase + DungeonCount - 1

	// 角色名冊:16 筆 32 B 的角色紀錄,從位移 2 開始。
	SaveRosterOffset = 0x0002
	RosterSize       = 16
	CharRecordSize   = 32

	// 以下位移由讀取序列累加得出,並以真實檔案交叉驗證。
	//
	// ⚠ 記憶體位址與檔案位移**不是固定差值**:讀取是逐欄位 fread/fgetc,
	// 而全域變數之間有對齊留下的空隙(`byte_3DFB8` 讀完 10 B 到 0x3DFC2,
	// 下一個欄位卻在 0x3DFC4)。整份存檔累積 4 B 的漂移,所以不能拿
	// 「線性位址 − 某個常數」去推位移,一定要跟著讀取序列走。
	SaveFoodOffset    = 0x0202 // word_3DFB4,u16(存糧;開局 63 份)
	SaveGoldOffset    = 0x0204 // word_3DFB6,u16
	SaveKeysOffset    = 0x0206 // byte_3DFB8
	SaveGemsOffset    = 0x0207 // byte_3DFB9
	SaveTorchesOffset = 0x0208 // byte_3DFBA
	// 0x0209..0x020F 是 `byte_3DFBB`..`byte_3DFC1` 七個單位元組。
	//
	// ★ 這一段**兩端都已經釘死**:前面 0x0208 是火把(已驗)、後面 0x0210 是
	// 寶石碎片(已驗),中間剛好七個位元組對七個變數,沒有排列的餘地。
	// 名字則來自 `sub_154BC` 的撿取分支(`mov byte_3DFC0, 0FFh` 後面接
	// "The crown of Lord British"…)。
	// SaveGrappleOffset 是**抓鉤**(byte_3DFBB)。
	//
	// ★ 這一格此前被寫成「疑為繩索」而且「位移沒對出來」—— 兩件事都不對:
	// 位移就在上面那段七個位元組的第一格(0x0209),而它的用途由
	// `sub_188C4`(大地圖上的攀爬)寫得很明白:
	// `cmp byte_3DFBB, 0; jnz; 印 "With what?"` —— 沒有它就爬不了山。
	// U5 的道具清單裡那件東西叫 **Grapple**(抓鉤),不是繩索。
	SaveGrappleOffset = 0x0209 // byte_3DFBB
	SaveCarpetsOffset = 0x020A // byte_3DFBC(魔毯;先前記為「位移未知」)
	// SaveOddKeysOffset 是「怪鑰匙」(byte_3DFBD)—— `sub_154BC` 的鑰匙分支
	// 有兩條:品質 ≥ 0x80 進這一格,其餘進一般鑰匙(0x0206)。
	SaveOddKeysOffset = 0x020B
	// SaveAmuletOffset 是**不列顛王的護符**(byte_3DFBF)。
	//
	// ⚠⚠ 這一格此前叫「寶珠(Orb)」——**錯的**。`sub_154BC` 的撿取分支
	// 寫得很明白:`mov byte_3DFBF, 0FFh` 後面接的字串是
	// "The Amulet of Lord British!"。而 `sub_1E8D4` 建可用道具清單時,
	// 它落在特殊道具表第 2 筆(`Amulet`)的位置,兩處一致。
	//
	// U5 的三件信物是**王冠 / 權杖 / 護符**,沒有寶珠 —— 那是 U6 的東西。
	// 這個錯名一路影響到結局判定的變數名(`Regalia.Orb`),已全部更正。
	SaveAmuletOffset = 0x020D // byte_3DFBF
	SaveCrownOffset   = 0x020E // byte_3DFC0 王冠
	SaveSceptreOffset = 0x020F // byte_3DFC1 權杖
	// 0x0214..0x0219 是 `byte_3DFC8`..`byte_3DFCD` 六個單位元組,
	// 同樣兩端釘死(前為 0x0210 起 4 B 的碎片,後為已驗的檀香木盒 0x0219)。
	//
	// ✅ **六格全數定名了**(`docs/re/79`)。兩條互不相干的證據:
	//
	//  1. `sub_1E8D4`(建 U 的清單)把它們抄成 +32..+37,而 `sub_1A5E8` 的跳表
	//     case 標註直接給出名字:`aSpyglass` = 32、`aPlans` = 33、
	//     `aSextant` = 34、`aWatchThePocket` = 35、`aBadge` = 36、木盒 = 37。
	//  2. `sub_1B964`(對話 opcode 0x86「給汝一樣東西」)裡三個信物是
	//     **直接寫 0xFF**:`'H'` → `byte_3DFCA`、`'I'` → `byte_3DFC8`、
	//     `'J'` → `byte_3DFCC`。與上面逐一相符。
	//
	// ⇒ 望遠鏡 / 六分儀 / 懷錶的持有旗標**不再是「位移未釘死」**
	// (`docs/re/44` §4 的那條 ⬜ 結案)。
	SaveSpyglassOffset = 0x0214 // byte_3DFC8 望遠鏡
	SavePlansOffset    = 0x0215 // byte_3DFC9 圖紙
	SaveSextantOffset  = 0x0216 // byte_3DFCA 六分儀
	SaveWatchOffset    = 0x0217 // byte_3DFCB 懷錶
	SaveBadgeOffset    = 0x0218 // byte_3DFCC 黑徽章
	// SaveItemsOffset 起 48 B,索引就是裝備編號(sub_11AF0 的 byte_3DFD0[裝備編號])。
	SaveItemsOffset = 0x021A
	// 卷軸 / 藥水 / 月石:與記憶體佈局同序,而且**兩端都有錨點**。
	//
	//	存檔 0x024A ↔ byte_3E000(已驗的 SaveSpellsOffset,48 B 咒語)
	//	存檔 0x02AA ↔ byte_3E060(已驗的 SaveReagentsOffset,8 B 藥草)
	//
	// 兩者相距 0x60,記憶體位址也正好相距 0x60 → 中間這一段沒有滑動的餘地,
	// `byte_3E030` / `byte_3E038` / `byte_3E050` 的位移是**算出來的,不是猜的**。

	// SaveScrollsOffset 起 8 B(byte_3E030):八種卷軸各持有幾捲。
	SaveScrollsOffset = 0x027A
	// SavePotionsOffset 起 8 B(byte_3E038):八色藥水各持有幾瓶。
	SavePotionsOffset = 0x0282
	// 月石是**八顆 × 四個欄位**,不是十六顆旗標。
	//
	// ⚠⚠ **更正(2026-08-08)**:此處原本寫「`0x029A` 起 16 B,十六顆月石,
	// 0xFF = 在手上」。長度夾對了(下一個已知欄位是 0x02AA 藥草),
	// **顆數與語意錯了** —— 那 16 B 是「八顆的地點」加「八顆的樓層」兩段。
	//
	// 一手證據是 `sub_1A2F8`(埋月石)那四行連寫:
	//
	//	byte_3E040[i] = byte_3E0A6   ; 玩家 X
	//	byte_3E048[i] = byte_3E0A7   ; 玩家 Y
	//	byte_3E050[i] = byte_3E0A3   ; ★ 地點
	//	byte_3E058[i] = byte_3E0A5   ; ★ 樓層
	//
	// 四個陣列各間隔 8 B ⇒ 每個都是 8 格。而 `sub_1E8D4`(建 U 的清單)
	// 只掃 `ecx < 8`、拿 `byte_3E050[i] == 0FFh` 當「還拿得出來」——
	// 兩處獨立咬合:**八顆**,而 0xFF 是「地點欄沒被寫過」= 還在身上。
	//
	// 位移換算的錨是同一組:`byte_3E000 ↔ 0x024A`、`byte_3E060 ↔ 0x02AA`,
	// 兩端夾住中間 0x60 B,沒有滑動空間。所以 0x028A / 0x0292 這 16 B
	// (原本整段沒解碼)就是 X 與 Y。
	SaveMoonstoneXOffset     = 0x028A // byte_3E040
	SaveMoonstoneYOffset     = 0x0292 // byte_3E048
	SaveMoonstoneLocOffset   = 0x029A // byte_3E050
	SaveMoonstoneFloorOffset = 0x02A2 // byte_3E058
	// MoonstoneCount 是月石的顆數。
	MoonstoneCount = 8

	// SaveReagentsOffset 起 8 B,順序同 ReagentNames(sub_11588 的 byte_3E060[藥草編號])。
	SaveReagentsOffset = 0x02AA

	SavePartySizeOffset = 0x02B5 // byte_3E06B
	SaveYearOffset      = 0x02CE // word_3E084
	SaveTimeStopOffset  = 0x02D4 // byte_3E08A('T' 停止時間、'Q' 速度加倍)
	// SaveActiveMemberOffset 是「指定行動者」(byte_3E08B,0xFF = 沒指定)。
	//
	// ★ 位移由 `sub_27D24` 的讀取序列推出來:那一段是**一格一格 `fgetc`**,
	// 所以檔案位移與全域位址同步遞增。三個錨點交叉驗證:
	// `byte_3E08C`=0x02D6、`byte_3E08F`=0x02D9、`byte_3E098`=0x02E2,
	// 差值與位址差完全一致 ⇒ `byte_3E08B` = 0x02D6 − 1。
	// 而 `INIT.GAM` / `SAVED.GAM` 這一格都是 **0xFF**,正是它的哨兵值。
	SaveActiveMemberOffset = 0x02D5 // byte_3E08B
	SaveTransportOffset    = 0x02D6 // byte_3E08C(隊伍當前載具 tile)
	SaveMonthOffset     = 0x02D7 // byte_3E08D
	SaveDayOffset       = 0x02D8 // byte_3E08E
	SaveHourOffset      = 0x02D9 // byte_3E08F
	SaveMinuteOffset    = 0x02DB // byte_3E091
	SaveKarmaOffset     = 0x02E2 // byte_3E098
	// SaveTurnCounterOffset 是每回合 +1 的計數器(byte_3E09B,上限 255)——
	// 施捨給乞丐的業報靠它節流(`docs/re/99` §5b)。`INIT.GAM` 是 0。
	SaveTurnCounterOffset = 0x02E5 // byte_3E09B
	SaveLocationOffset  = 0x02ED // byte_3E0A3
	SaveFloorOffset     = 0x02EF // byte_3E0A5
	SaveXOffset         = 0x02F0 // byte_3E0A6
	SaveYOffset         = 0x02F1 // byte_3E0A7

	// 以下六個欄位是任務進度,位移一樣跟著 `sub_27D24` 的讀取序列累加:
	// 0x02F1 之後是 byte_3E0A8..byte_3E0B7 十六個單位元組,接著 byte_3E0B8 的
	// 32 B 區塊,然後才輪到這一段。
	//
	// ★ 位移對不對有一個現成的證據:`INIT.GAM` 與 `SAVED.GAM` 在 0x0322..0x0339
	// 這 24 B 裡**只有 0x0325 是 0xFF,其餘全是 0**,而 0x0325 正是
	// 「現在被召喚出來的是哪一個暗影君主」(0xFF = 沒有)。位移若偏一格,
	// 那個 0xFF 就會落在別的欄位上。

	// SaveShadowlordAtOffset 起 3 B(byte_3E0D8):三個暗影君主各自盤據哪個地點。
	// 0 = 不在城裡,0xFF = 已被消滅。
	SaveShadowlordAtOffset = 0x0322
	// SaveShadowlordHereOffset 是現在被召喚出來的是哪一個(byte_3E0DB,0xFF = 沒有)。
	SaveShadowlordHereOffset = 0x0325
	// SaveShrineQuestOffset 起 2 B(byte_3E0DC):**進行中**的聖壇試煉,一德一位元。
	SaveShrineQuestOffset = 0x0326
	// SaveCodexLearnedOffset 起 2 B(byte_3E0DE):已在寶典上讀到的美德,一德一位元。
	//
	// ⚠ 這一個**不是**聖壇設的 —— 是寶典(`sub_1D850`)設的。
	SaveCodexLearnedOffset = 0x0328
	// SaveDungeonSealOffset 起 8 B(byte_3E0E0):八座地牢入口,bit 0x80 = 已封印。
	SaveDungeonSealOffset = 0x032A
	// SaveShrineFlagOffset 起 8 B(byte_3E0E8):八座聖壇,bit 0x80 = 已被玷污。
	SaveShrineFlagOffset = 0x0332
	// SaveRoomsClearedOffset 起 **14 B**(byte_3E0F0):清過的地牢房間位元陣列。
	//
	// ★★ **14 這個長度本身就是一條證據。** 索引是
	// `DungeonRoomBlock(地點碼)*16 + 房號`,而 `DungeonRoomBlock` 有一個
	// 「≥1 就 −1」的修正 ⇒ 八座地牢只佔 **7** 個區塊 ⇒ 7 × 16 = 112 位元
	// = 剛好 14 位元組。若索引是 0..7 就會需要 16 位元組。
	// ⇒ 原版 `sub_27D24` 的 `push 0Eh` 獨立佐證了那個共用區塊的怪處。
	SaveRoomsClearedOffset = 0x033A
	// RoomsClearedBytes 是那個位元陣列的長度(原版 `push 0Eh`)。
	RoomsClearedBytes = 14

	// SaveRemovedNPCOffset 起 128 B(`dword_3E36C`):**32 個地點各一個 u32**,
	// 位元 i = 那個地點的第 i 個 NPC 已經被永久清掉(`sub_218` 設)。
	//
	// 位移一樣跟著讀取序列累加:0x0332 之後是 byte_3E0F0(14 B)、
	// byte_3E100 / byte_3E120 / byte_3E140(各 32 B)、byte_3E160..3E16B
	// (12 個單位元組)、dword_3E16C(512 B),然後才是這一段。
	//
	// ⚠ 原版的陣列基底是 `dword_3E368`,索引是**地點編號 1..32** ——
	// 所以存檔裡的第 0 個 u32 對應**地點 1**,不是地點 0(大地圖沒有場景 NPC)。
	SaveRemovedNPCOffset = 0x05B4
	// RemovedNPCLocations 是上面那一段涵蓋幾個地點。
	RemovedNPCLocations = 32
)

// 角色紀錄的欄位位移。
//
// 由六名初始角色橫向對照得出:性別欄只有 0x0B/0x0C 兩種而 Mariah 與 Jaana 是 0x0C;
// 職業欄是可讀字母 A/F/B/M(Avatar/Fighter/Bard/Mage);
// 目前 HP 與最大 HP 在五名角色身上相等(全新角色),經驗值與等級也彼此相符
// (150/167 → 等級 2,249/260/278 → 等級 3)。
const (
	CharName     = 0  // 9 B,NUL 補齊
	CharGender   = 9  // 0x0B 男 / 0x0C 女
	CharClass    = 10 // 'A' 聖者 / 'F' 戰士 / 'B' 吟遊詩人 / 'M' 法師
	CharStatus   = 11 // 'G' 良好
	CharStrength = 12
	CharDex      = 13
	CharIntel    = 14
	CharMP       = 15
	CharHP       = 16 // u16 LE
	CharMaxHP    = 18 // u16 LE
	CharExp      = 20 // u16 LE
	CharLevel    = 22
	// CharDamageResist 是傷害計算時的減傷值(`sub_B274` 讀
	// `byte_3DDCC[角色*32]`,而 0x3DDCC − 0x3DDB4 = 0x18)。
	//
	// ⚠ 名字刻意不叫 armour:它**不是**裝備防禦的加總(那是 CharArmour
	// 那個裝備欄位加 ItemDefence 算的)。六名初始角色的這一欄全都是 7,
	// 但聖者穿鎖甲、Mariah 只穿布甲 —— 是什麼算出來的還沒追到,先照原樣讀。
	CharDamageResist = 24

	CharNameLen = 9
)

// 性別碼。原版只有兩種(由六名初始角色橫向對照得出,見上)。
const (
	GenderMale   = 0x0B
	GenderFemale = 0x0C
)

// 角色狀態碼。都是可讀字母,原版直接拿 `cmp byte_3DDBF[32*i], 'P'` 這樣比
// (治療所 sub_12838 的三個分支)。
const (
	StatusGood     = 'G'
	StatusPoisoned = 'P'
	StatusDead     = 'D'
	StatusCharmed  = 'C'
	StatusAsleep   = 'S'
)

// Character 是一名角色。
type Character struct {
	Name     string
	Gender   byte
	Class    byte
	Status   byte
	Strength byte
	Dex      byte
	Intel    byte
	MP       byte
	HP       uint16
	MaxHP    uint16
	Exp      uint16
	Level    byte

	// Raw 是完整的 32 B 紀錄。0x17 之後的欄位**已經解出來了** ——
	// 見 `items.go` 的 `CharInnDays`(0x17)/ `CharHelm`..`CharAmulet`(0x19..0x1E)/
	// `CharInnFlag`(0x1F)。留著 Raw 是為了保住還沒命名的那幾格,不是因為沒解。
	Raw [CharRecordSize]byte
}

// Present 回報這一格名冊有沒有角色。
func (c *Character) Present() bool { return c.Name != "" || c.Class != 0 }

// StatusName 回傳狀態的中文名。
//
// ⚠ 只影響顯示。比對一律用位元組常數(`StatusGood` 等)——
// 治療所與復活判定讀的是那個位元組,譯名換了也不該影響它們。
func StatusName(status byte) string {
	switch status {
	case StatusGood:
		return "康健"
	case StatusPoisoned:
		return "中毒"
	case StatusDead:
		return "身亡"
	case StatusAsleep:
		return "沉睡"
	case StatusCharmed:
		return "被惑"
	}
	return "?"
}

// ClassName 回傳職業的中文名。
//
// 職業代碼是可讀字母,譯名對齊 u4-cht / u6-cht 的《創世紀聖者之書》體系。
func (c *Character) ClassName() string {
	switch c.Class {
	case 'A':
		return "聖者"
	case 'F':
		return "戰士"
	case 'B':
		return "吟遊詩人"
	case 'M':
		return "法師"
	case 'T':
		return "盜賊"
	}
	return string(rune(c.Class))
}

// Inventory 是隊伍共用的背包(不隸屬個別角色)。
type Inventory struct {
	// Food 是隊伍的存糧(原版 word_3DFB4)。在酒館買一餐或買幾份乾糧都會增加。
	Food    int
	Gold    int
	Keys    int
	Gems    int
	Torches int
	// Items[裝備編號] 是持有數量,上限 CarryLimit。
	Items [ItemCount]int
	// Reagents[藥草編號] 是持有份數,順序同 ReagentNames。
	Reagents [ReagentCount]int
	// Spells[咒語編號] 是已調配好的份數,上限 SpellStackLimit。
	Spells [SpellCount]int
	// Grapple 是抓鉤(原版 byte_3DFBB,存檔 0x0209)—— 大地圖上爬山的前提。
	Grapple int
	// Carpets 是持有的魔毯數(原版 byte_3DFBC,存檔 0x020A)。
	Carpets int
	// OddKeys 是「怪鑰匙」(原版 byte_3DFBD)—— 品質最高位被設起來的那種鑰匙。
	OddKeys int
	// Scrolls[卷軸編號] 是持有幾捲(原版 byte_3E030),名字見 ScrollSpells。
	Scrolls [ScrollCount]int
	// Potions[顏色編號] 是持有幾瓶(原版 byte_3E038),顏色見 PotionColours。
	Potions [PotionCount]int
	// Moonstones[i] 是第 i 顆月石埋在哪 —— 見 Moonstone 與上面的位移更正。
	Moonstones [MoonstoneCount]Moonstone
}

// Moonstone 是一顆月石的四個欄位(原版 byte_3E040/48/50/58 的同一格)。
//
// 「在手上」不是另一個欄位,是 `Location == MoonstoneInHand`(0xFF)——
// 原版 `sub_1E8D4` 就是這樣判「這顆拿得出來」的。
type Moonstone struct {
	X, Y     int
	Location int
	Floor    int
}

// MoonstoneInHand 是地點欄的「還沒埋」值。
const MoonstoneInHand = 0xFF

// InHand 回報這顆月石還在身上(還沒被埋)。
func (m Moonstone) InHand() bool { return m.Location == MoonstoneInHand }

// Regalia 是不列顛王的信物與圖紙 —— 各佔一個 0/0xFF 的位元組。
//
// 四個都由 Get 指令撿到(`sub_154BC` 的 0xB5 / 0xB6 / 0xB7 / 0x04 分支),
// 撿到時各印一句 "The crown of Lord British" 之類。
type Regalia struct {
	Crown, Sceptre, Amulet bool
	// Plans 是圖紙(`byte_3DFC9`)。
	Plans bool
}

// putFlag 寫一個 0 / 0xFF 的旗標(原版一律用 0xFF 當真)。
func putFlag(out []byte, off int, v bool) {
	if v {
		out[off] = 0xFF
		return
	}
	out[off] = 0
}

// Save 是一份存檔。
//
// 只解出已經驗證過的欄位;其餘保留在 Raw 裡。與其對沒把握的位移硬取名字,
// 不如讓呼叫端看得到「這一段還沒解」。
type Save struct {
	Roster    [RosterSize]Character
	Inventory Inventory
	PartySize int
	Year      int
	Month     int
	Day       int
	Hour      int
	Minute    int
	Karma     int
	Transport byte
	// Moongates 是八個月相各自的目的地(存檔 0x028A 起)。
	Moongates [MoonPhaseCount]MoongateDest
	Location  int // 0 = 大地圖
	Floor     int // signed:負數是地下
	X, Y      int

	// Shards[i] 是有沒有第 i 塊寶石碎片(存檔 0x0210 起 4 B,第 4 B 未用)。
	Shards [ShardCount]byte
	// Regalia 是王冠 / 權杖 / 護符 / 圖紙。
	Regalia Regalia
	// SandalwoodBox 是有沒有那只檀香木盒(byte_3DFCD)—— 真結局的條件。
	SandalwoodBox byte
	// Spyglass / Sextant / Watch 是三件航海道具的持有旗標
	// (byte_3DFC8 / byte_3DFCA / byte_3DFCB,位移見上面的說明)。
	//
	// ⚠ 這三格此前是「位移未釘死」所以 U 的清單**無條件列出**它們
	// (`docs/re/44` §4)。現在釘死了(`docs/re/79`),改成照旗標列。
	Spyglass byte
	Sextant  byte
	Watch    byte
	// Badge 是黑棘的黑徽章(byte_3DFCC,存檔 0x0218)。
	//
	// ★ 位移的證據來自 Ztats 的 Items 頁:`sub_1E8D4` 把 `byte_3DFC8`..`byte_3DFCD`
	// 六格連續搬進清單第 32..37 筆,而那六個名字(`Spyglass` `HMS Cape Plan`
	// `Sextant` `Pocket Watch` `Black Badge` `Wooden Box`)與 U 指令的 case
	// 標註逐一相符 ⇒ 中間沒有空格,徽章就是 0x0218(`docs/re/94`)。
	Badge byte
	// ShadowlordAt[i] 是第 i 個暗影君主盤據的地點編號(0 = 不在城裡,0xFF = 已消滅)。
	ShadowlordAt [ShadowlordCount]byte
	// ShadowlordHere 是現在被召喚出來的那一個(0xFF = 沒有)。
	ShadowlordHere byte
	// ShrineQuest 是進行中的聖壇試煉,一德一位元。
	ShrineQuest byte
	// CodexLearned 是已在寶典上讀到的美德,一德一位元。
	CodexLearned byte
	// DungeonSeal[i] 的 bit 0x80 = 第 i 座地牢入口已被力量之言封印。
	DungeonSeal [VirtueCount]byte
	// ShrineFlag[i] 的 bit 0x80 = 第 i 座聖壇已被玷污。
	ShrineFlag [VirtueCount]byte
	// ActiveMember 是「指定行動者」的名冊索引;0xFF = 沒指定。
	ActiveMember byte
	// TurnCounter 是每回合 +1 的計數器(施捨業報的節流)。
	TurnCounter byte
	// RoomsCleared 是清過的地牢房間位元陣列(14 B = 7 區塊 × 16 間)。
	RoomsCleared [RoomsClearedBytes]byte
	// RemovedNPC[地點-1] 的位元 i = 那個地點的第 i 個 NPC 已被永久清掉。
	RemovedNPC [RemovedNPCLocations]uint32

	Raw []byte
}

// LoadSave 讀入一份存檔(`SAVED.GAM` 或 `INIT.GAM`)。
func LoadSave(path string) (*Save, error) {
	raw, err := readExact(path, SaveFileSize)
	if err != nil {
		return nil, err
	}
	return ParseSave(raw)
}

// ParseSave 解析存檔內容。
func ParseSave(raw []byte) (*Save, error) {
	if len(raw) != SaveFileSize {
		return nil, fmt.Errorf("存檔 %d B,預期 %d B", len(raw), SaveFileSize)
	}
	s := &Save{Raw: raw}
	for i := range s.Roster {
		off := SaveRosterOffset + i*CharRecordSize
		rec := raw[off : off+CharRecordSize]
		c := &s.Roster[i]
		copy(c.Raw[:], rec)
		c.Name = cstring(rec[CharName : CharName+CharNameLen])
		c.Gender = rec[CharGender]
		c.Class = rec[CharClass]
		c.Status = rec[CharStatus]
		c.Strength = rec[CharStrength]
		c.Dex = rec[CharDex]
		c.Intel = rec[CharIntel]
		c.MP = rec[CharMP]
		c.HP = binary.LittleEndian.Uint16(rec[CharHP:])
		c.MaxHP = binary.LittleEndian.Uint16(rec[CharMaxHP:])
		c.Exp = binary.LittleEndian.Uint16(rec[CharExp:])
		c.Level = rec[CharLevel]
	}
	s.Moongates = parseMoongates(raw)
	s.Inventory.Food = int(binary.LittleEndian.Uint16(raw[SaveFoodOffset:]))
	s.Inventory.Gold = int(binary.LittleEndian.Uint16(raw[SaveGoldOffset:]))
	s.Inventory.Keys = int(raw[SaveKeysOffset])
	s.Inventory.Gems = int(raw[SaveGemsOffset])
	s.Inventory.Torches = int(raw[SaveTorchesOffset])
	s.Inventory.Grapple = int(raw[SaveGrappleOffset])
	s.Inventory.Carpets = int(raw[SaveCarpetsOffset])
	s.Inventory.OddKeys = int(raw[SaveOddKeysOffset])
	s.Regalia.Amulet = raw[SaveAmuletOffset] != 0
	s.Regalia.Crown = raw[SaveCrownOffset] != 0
	s.Regalia.Sceptre = raw[SaveSceptreOffset] != 0
	s.Regalia.Plans = raw[SavePlansOffset] != 0
	for i := 0; i < ItemCount; i++ {
		s.Inventory.Items[i] = int(raw[SaveItemsOffset+i])
	}
	for i := 0; i < SpellCount; i++ {
		s.Inventory.Spells[i] = int(raw[SaveSpellsOffset+i])
	}
	for i := 0; i < ReagentCount; i++ {
		s.Inventory.Reagents[i] = int(raw[SaveReagentsOffset+i])
	}
	for i := 0; i < ScrollCount; i++ {
		s.Inventory.Scrolls[i] = int(raw[SaveScrollsOffset+i])
	}
	for i := 0; i < PotionCount; i++ {
		s.Inventory.Potions[i] = int(raw[SavePotionsOffset+i])
	}
	for i := 0; i < MoonstoneCount; i++ {
		s.Inventory.Moonstones[i] = Moonstone{
			X:        int(raw[SaveMoonstoneXOffset+i]),
			Y:        int(raw[SaveMoonstoneYOffset+i]),
			Location: int(raw[SaveMoonstoneLocOffset+i]),
			// 樓層與 SaveFloorOffset 同一種表示:補數,地底世界是負的。
			Floor: int(int8(raw[SaveMoonstoneFloorOffset+i])),
		}
	}

	s.PartySize = int(raw[SavePartySizeOffset])
	s.Year = int(binary.LittleEndian.Uint16(raw[SaveYearOffset:]))
	s.Month = int(raw[SaveMonthOffset])
	s.Day = int(raw[SaveDayOffset])
	s.Hour = int(raw[SaveHourOffset])
	s.Minute = int(raw[SaveMinuteOffset])
	s.Karma = int(raw[SaveKarmaOffset])
	s.Transport = raw[SaveTransportOffset]
	s.Location = int(raw[SaveLocationOffset])
	s.Floor = int(int8(raw[SaveFloorOffset]))
	s.X = int(raw[SaveXOffset])
	s.Y = int(raw[SaveYOffset])

	copy(s.Shards[:], raw[SaveShardsOffset:])
	s.SandalwoodBox = raw[SaveSandalwoodBoxOffset]
	s.Badge = raw[SaveBadgeOffset]
	s.Spyglass = raw[SaveSpyglassOffset]
	s.Sextant = raw[SaveSextantOffset]
	s.Watch = raw[SaveWatchOffset]
	copy(s.ShadowlordAt[:], raw[SaveShadowlordAtOffset:])
	s.ShadowlordHere = raw[SaveShadowlordHereOffset]
	s.ShrineQuest = raw[SaveShrineQuestOffset]
	s.CodexLearned = raw[SaveCodexLearnedOffset]
	copy(s.DungeonSeal[:], raw[SaveDungeonSealOffset:])
	s.ActiveMember = raw[SaveActiveMemberOffset]
	s.TurnCounter = raw[SaveTurnCounterOffset]
	copy(s.RoomsCleared[:], raw[SaveRoomsClearedOffset:])
	copy(s.ShrineFlag[:], raw[SaveShrineFlagOffset:])
	for i := range s.RemovedNPC {
		s.RemovedNPC[i] = binary.LittleEndian.Uint32(raw[SaveRemovedNPCOffset+i*4:])
	}

	if err := s.validate(); err != nil {
		return nil, err
	}
	return s, nil
}

// validate 檢查解出來的值合不合理。
//
// 位移是用「把讀取序列的欄位大小累加起來」算的,只要中間漏算一個欄位,
// 後面全部會偏移。偏移之後讀出來的仍然是「某個數字」,不會自己報錯 ——
// 所以這裡主動撞一次牆:日期與隊伍人數只要有一個離譜,就是位移錯了。
func (s *Save) validate() error {
	switch {
	case s.Month < 1 || s.Month > MonthsPerYearSave:
		return fmt.Errorf("月份是 %d(合理範圍 1..%d)—— 存檔位移大概算錯了",
			s.Month, MonthsPerYearSave)
	case s.Day < 1 || s.Day > DaysPerMonthSave:
		return fmt.Errorf("日期是 %d(合理範圍 1..%d)", s.Day, DaysPerMonthSave)
	case s.Hour > 23:
		return fmt.Errorf("小時是 %d", s.Hour)
	case s.Minute > 59:
		return fmt.Errorf("分鐘是 %d", s.Minute)
	case s.PartySize > MaxPartySize:
		return fmt.Errorf("隊伍 %d 人(上限 %d)", s.PartySize, MaxPartySize)
	case s.Location > LocationCodeMax:
		return fmt.Errorf("地點編號 %d 超出 0..%d", s.Location, LocationCodeMax)
	}
	// 背包欄位與角色紀錄之間隔著會漂移的那 4 B(見上方位移註解),
	// 所以另外撞一次牆:持有數量的上限是 99,超過就是位移偏了。
	for i, n := range s.Inventory.Items {
		if n > CarryLimit {
			return fmt.Errorf("裝備 %d 持有 %d 個(上限 %d)—— 背包位移大概算錯了", i, n, CarryLimit)
		}
	}
	for i, n := range s.Inventory.Reagents {
		if n > CarryLimit {
			return fmt.Errorf("藥草 %d 持有 %d 份(上限 %d)", i, n, CarryLimit)
		}
	}
	return nil
}

// 曆法上限。與 internal/game 的 Clock 同源(原版 sub_29304 的進位條件),
// 這裡重複一份是為了讓 u5data 不必反向依賴 game。
const (
	DaysPerMonthSave  = 28
	MonthsPerYearSave = 13
	// MaxPartySize 是隊伍人數上限。原版 sub_1BB5C 用 `cmp byte_3E06B, 6` 判滿員。
	MaxPartySize = 6
)

// Party 回傳目前在隊伍裡的角色(名冊前 PartySize 名)。
func (s *Save) Party() []*Character {
	n := s.PartySize
	if n > RosterSize {
		n = RosterSize
	}
	out := make([]*Character, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, &s.Roster[i])
	}
	return out
}

func cstring(b []byte) string {
	for i, c := range b {
		if c == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}

// 存檔寫回
//
// 寫入函式 `sub_284CC` 與讀取 `sub_27D24` **逐欄位對稱**(同樣的順序、
// 同樣的大小),所以寫回不需要另一套位移表 —— 讀得回來就寫得出去。
//
// 作法是**以讀進來的 4,192 B 為底**,只把已經解出來的欄位覆蓋回去。
// 還沒解的欄位(魔法、任務旗標、地牢狀態…)原樣保留,不會因為引擎還沒
// 實作那些系統就把它們清成 0 —— 那會讓存檔在原版裡壞掉。

// Encode 把 Save 序列化回 4,192 B。
func (s *Save) Encode() ([]byte, error) {
	if len(s.Raw) != SaveFileSize {
		return nil, fmt.Errorf("底稿 %d B,預期 %d B —— Encode 需要一份讀進來的存檔當底",
			len(s.Raw), SaveFileSize)
	}
	out := make([]byte, SaveFileSize)
	copy(out, s.Raw)

	for i := range s.Roster {
		off := SaveRosterOffset + i*CharRecordSize
		c := &s.Roster[i]
		rec := out[off : off+CharRecordSize]
		copy(rec, c.Raw[:])
		// 結構化欄位蓋過 Raw —— 遊戲中改到的是它們(療傷、入隊、狀態變化)。
		copy(rec[CharName:CharName+CharNameLen], padName(c.Name))
		rec[CharGender] = c.Gender
		rec[CharClass] = c.Class
		rec[CharStatus] = c.Status
		rec[CharStrength] = c.Strength
		rec[CharDex] = c.Dex
		rec[CharIntel] = c.Intel
		rec[CharMP] = c.MP
		binary.LittleEndian.PutUint16(rec[CharHP:], c.HP)
		binary.LittleEndian.PutUint16(rec[CharMaxHP:], c.MaxHP)
		binary.LittleEndian.PutUint16(rec[CharExp:], c.Exp)
		rec[CharLevel] = c.Level
	}

	encodeMoongates(out, s.Moongates)
	binary.LittleEndian.PutUint16(out[SaveFoodOffset:], clampU16(s.Inventory.Food))
	binary.LittleEndian.PutUint16(out[SaveGoldOffset:], clampU16(s.Inventory.Gold))
	out[SaveKeysOffset] = byte(s.Inventory.Keys)
	out[SaveGemsOffset] = byte(s.Inventory.Gems)
	out[SaveTorchesOffset] = byte(s.Inventory.Torches)
	out[SaveGrappleOffset] = byte(s.Inventory.Grapple)
	out[SaveCarpetsOffset] = byte(s.Inventory.Carpets)
	out[SaveOddKeysOffset] = byte(s.Inventory.OddKeys)
	putFlag(out, SaveAmuletOffset, s.Regalia.Amulet)
	putFlag(out, SaveCrownOffset, s.Regalia.Crown)
	putFlag(out, SaveSceptreOffset, s.Regalia.Sceptre)
	putFlag(out, SavePlansOffset, s.Regalia.Plans)
	for i := 0; i < ItemCount; i++ {
		out[SaveItemsOffset+i] = byte(s.Inventory.Items[i])
	}
	for i := 0; i < SpellCount; i++ {
		out[SaveSpellsOffset+i] = byte(s.Inventory.Spells[i])
	}
	for i := 0; i < ReagentCount; i++ {
		out[SaveReagentsOffset+i] = byte(s.Inventory.Reagents[i])
	}
	for i := 0; i < ScrollCount; i++ {
		out[SaveScrollsOffset+i] = byte(s.Inventory.Scrolls[i])
	}
	for i := 0; i < PotionCount; i++ {
		out[SavePotionsOffset+i] = byte(s.Inventory.Potions[i])
	}
	for i := 0; i < MoonstoneCount; i++ {
		m := s.Inventory.Moonstones[i]
		out[SaveMoonstoneXOffset+i] = byte(m.X)
		out[SaveMoonstoneYOffset+i] = byte(m.Y)
		out[SaveMoonstoneLocOffset+i] = byte(m.Location)
		out[SaveMoonstoneFloorOffset+i] = byte(int8(m.Floor))
	}

	out[SavePartySizeOffset] = byte(s.PartySize)
	binary.LittleEndian.PutUint16(out[SaveYearOffset:], uint16(s.Year))
	out[SaveMonthOffset] = byte(s.Month)
	out[SaveDayOffset] = byte(s.Day)
	out[SaveHourOffset] = byte(s.Hour)
	out[SaveMinuteOffset] = byte(s.Minute)
	out[SaveKarmaOffset] = byte(s.Karma)
	out[SaveActiveMemberOffset] = s.ActiveMember
	out[SaveTurnCounterOffset] = s.TurnCounter
	copy(out[SaveRoomsClearedOffset:], s.RoomsCleared[:])
	out[SaveTransportOffset] = s.Transport
	out[SaveLocationOffset] = byte(s.Location)
	out[SaveFloorOffset] = byte(int8(s.Floor))
	out[SaveXOffset] = byte(s.X)
	out[SaveYOffset] = byte(s.Y)

	copy(out[SaveShardsOffset:], s.Shards[:])
	out[SaveSandalwoodBoxOffset] = s.SandalwoodBox
	out[SaveBadgeOffset] = s.Badge
	out[SaveSpyglassOffset] = s.Spyglass
	out[SaveSextantOffset] = s.Sextant
	out[SaveWatchOffset] = s.Watch
	copy(out[SaveShadowlordAtOffset:], s.ShadowlordAt[:])
	out[SaveShadowlordHereOffset] = s.ShadowlordHere
	out[SaveShrineQuestOffset] = s.ShrineQuest
	out[SaveCodexLearnedOffset] = s.CodexLearned
	copy(out[SaveDungeonSealOffset:], s.DungeonSeal[:])
	copy(out[SaveShrineFlagOffset:], s.ShrineFlag[:])
	for i, v := range s.RemovedNPC {
		binary.LittleEndian.PutUint32(out[SaveRemovedNPCOffset+i*4:], v)
	}
	return out, nil
}

func padName(name string) []byte {
	b := make([]byte, CharNameLen)
	copy(b, name) // 超過 9 B 自動截斷,不足補 NUL
	return b
}

func clampU16(v int) uint16 {
	switch {
	case v < 0:
		return 0
	case v > 0xFFFF:
		return 0xFFFF
	}
	return uint16(v)
}
