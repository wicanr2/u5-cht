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
	// SaveItemsOffset 起 48 B,索引就是裝備編號(sub_11AF0 的 byte_3DFD0[裝備編號])。
	SaveItemsOffset = 0x021A
	// SaveReagentsOffset 起 8 B,順序同 ReagentNames(sub_11588 的 byte_3E060[藥草編號])。
	SaveReagentsOffset = 0x02AA

	SavePartySizeOffset = 0x02B5 // byte_3E06B
	SaveYearOffset      = 0x02CE // word_3E084
	SaveTimeStopOffset  = 0x02D4 // byte_3E08A('T' 停止時間、'Q' 速度加倍)
	SaveTransportOffset = 0x02D6 // byte_3E08C(隊伍當前載具 tile)
	SaveMonthOffset     = 0x02D7 // byte_3E08D
	SaveDayOffset       = 0x02D8 // byte_3E08E
	SaveHourOffset      = 0x02D9 // byte_3E08F
	SaveMinuteOffset    = 0x02DB // byte_3E091
	SaveKarmaOffset     = 0x02E2 // byte_3E098
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

	// Raw 是完整的 32 B 紀錄。裝備欄位(0x17 之後)還沒對到物品表,
	// 先原樣保留而不是硬取名字。
	Raw [CharRecordSize]byte
}

// Present 回報這一格名冊有沒有角色。
func (c *Character) Present() bool { return c.Name != "" || c.Class != 0 }

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
	// Carpets 是持有的魔毯數(原版 byte_3DFBC)。
	// ⚠ 存檔位移還沒對出來,目前只在遊戲中維護,讀檔不會帶進來。
	Carpets int
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
	// SandalwoodBox 是有沒有那只檀香木盒(byte_3DFCD)—— 真結局的條件。
	SandalwoodBox byte
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
	for i := 0; i < ItemCount; i++ {
		s.Inventory.Items[i] = int(raw[SaveItemsOffset+i])
	}
	for i := 0; i < SpellCount; i++ {
		s.Inventory.Spells[i] = int(raw[SaveSpellsOffset+i])
	}
	for i := 0; i < ReagentCount; i++ {
		s.Inventory.Reagents[i] = int(raw[SaveReagentsOffset+i])
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
	copy(s.ShadowlordAt[:], raw[SaveShadowlordAtOffset:])
	s.ShadowlordHere = raw[SaveShadowlordHereOffset]
	s.ShrineQuest = raw[SaveShrineQuestOffset]
	s.CodexLearned = raw[SaveCodexLearnedOffset]
	copy(s.DungeonSeal[:], raw[SaveDungeonSealOffset:])
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
	case s.Location > len(Locations):
		return fmt.Errorf("地點編號 %d 超出 0..%d", s.Location, len(Locations))
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
	for i := 0; i < ItemCount; i++ {
		out[SaveItemsOffset+i] = byte(s.Inventory.Items[i])
	}
	for i := 0; i < SpellCount; i++ {
		out[SaveSpellsOffset+i] = byte(s.Inventory.Spells[i])
	}
	for i := 0; i < ReagentCount; i++ {
		out[SaveReagentsOffset+i] = byte(s.Inventory.Reagents[i])
	}

	out[SavePartySizeOffset] = byte(s.PartySize)
	binary.LittleEndian.PutUint16(out[SaveYearOffset:], uint16(s.Year))
	out[SaveMonthOffset] = byte(s.Month)
	out[SaveDayOffset] = byte(s.Day)
	out[SaveHourOffset] = byte(s.Hour)
	out[SaveMinuteOffset] = byte(s.Minute)
	out[SaveKarmaOffset] = byte(s.Karma)
	out[SaveTransportOffset] = s.Transport
	out[SaveLocationOffset] = byte(s.Location)
	out[SaveFloorOffset] = byte(int8(s.Floor))
	out[SaveXOffset] = byte(s.X)
	out[SaveYOffset] = byte(s.Y)

	copy(out[SaveShardsOffset:], s.Shards[:])
	out[SaveSandalwoodBoxOffset] = s.SandalwoodBox
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
