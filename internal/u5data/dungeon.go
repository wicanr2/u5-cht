package u5data

import (
	"fmt"
	"os"
	"path/filepath"
)

// 地牢
//
// `DUNGEON.DAT` 是 4096 B = **8 座地牢 × 8 層 × 8×8 格,一格 1 B**。
// 載入的算式直接寫在 `sub_2D564` 裡:
//
//	sub_2C740("DUNGEON.DAT", dword_3E16C, 0x200, 地點 * 0x200 − 0x4000)
//
// 一次只讀 512 B(一座地牢的 8 層),而 `地點 * 512 − 0x4000` 代表
// **地點 0x20 對應檔案位移 0**;實際進去之後 `byte_3E0A3 = 地點 + 1`,
// 所以八座地牢的地點編號是 **0x21..0x28** —— 這正好對得上咒語場合表
// 「0x21..0x7F 算地牢」那一條(見 `docs/re/17`)。
//
// 層內定址是 `dword_3E16C[樓層 << 6 + y*8 + x]`(`sub_4594`、`sub_5150` 等處
// 一致),所以一層 64 B、row stride 8。

// 地牢的形狀。
const (
	DungeonCount  = 8
	DungeonLevels = 8
	DungeonSide   = 8
	DungeonLevelB = DungeonSide * DungeonSide // 64
	DungeonBlockB = DungeonLevels * DungeonLevelB
	DungeonFileB  = DungeonCount * DungeonBlockB // 4096
)

// DungeonLocationBase 是第一座地牢的地點編號(`sub_2D564` 的 `byte_3E0A3 = edi + 1`)。
const DungeonLocationBase = 0x21

// 地牢格子的**高四位元**就是種類(全檔都寫成 `tile & 0xF0`)。
const (
	// DungeonPassage 通道(可走)。**這是可走的那一個,不是牆** ——
	// 全檔出現 1199 次,而牆(0xB0)出現 1926 次。
	DungeonPassage = 0x00
	// DungeonLadderUp / Down / Both:`sub_417C` 的 Klimb 判定直接比這三個值。
	DungeonLadderUp   = 0x10
	DungeonLadderDown = 0x20
	DungeonLadderBoth = 0x30
	// DungeonChest 寶箱。
	DungeonChest = 0x40
	// DungeonFountain 噴泉(`sub_CE78` 印「a gurgling fountain!」)。
	DungeonFountain = 0x50
	// DungeonTrap 陷阱:0x61/0x69 是坑(掉下一層)、0x62/0x6A 是炸彈。
	DungeonTrap = 0x60
	// DungeonDoor 門。
	DungeonDoor = 0x70
	// DungeonMagic 魔法陷阱:0x80/0x89 睡眠、0x82/0x8A 火焰。
	DungeonMagic = 0x80
	// DungeonRoomA / DungeonRoomF 是地牢房間 —— **兩個高位元都算**
	//(`sub_5150` 的 `cmp al, 0F0h / jz / cmp al, 0A0h`)。低四位元是房號。
	DungeonRoomA = 0xA0
	DungeonRoomF = 0xF0
	// DungeonWall 牆。
	DungeonWall = 0xB0
	// 0xC0 / 0xD0 走不過去,語意還沒完全定。
	DungeonUnknownC = 0xC0
	DungeonUnknownD = 0xD0
	// DungeonDoorway 是**門口**(2026-08-07 修正,原本記成「語意未定,當牆」)。
	//
	// 依據 `sub_48F4`:站在 0xE0 上時左轉右轉都會被擋下,印「Not in doorway!」
	//(轉身 180 度不受限)。而移動的阻擋判定 `0xA0 < kind < 0xE0` **不含 0xE0**
	// —— 門口是走得過去的。
	DungeonDoorway = 0xE0
)

// 陷阱與魔法力場的完整位元組(`sub_5150` 比的是**整個位元組**,不是高四位元)。
//
// 0x80..0x83 這一組不只是「地圖上本來就有的魔法陷阱」,它同時是**四個
// `*Grav` 咒語放出來的力場**(`sub_18A08` 的 `byte_55E24[種類]`
// = `82 81 80 83`)。玩家放的與地圖上的走同一段程式碼,所以是同一組編號。
//
//	0x80 睡眠(In Zu Grav)     0x81 毒(In Nox Grav)
//	0x82 烈焰(In Flam Grav)   0x83 電擊(In Sanct Grav)—— 見下
//
// 每一個都有 **+8 的變體**(0x88..0x8B),那是「頭上有洞」的位元疊上去的:
// `sub_18A08` 寫回的是 `(舊值 & 8) | 力場編號`。
//
// ⚠ 我第一版把 0x89 寫成睡眠 —— 錯了。`jpt_52C7` 這張 10 格跳表
//(索引 = 格子 − 0x81)排得清清楚楚:0x80 與 0x88 是睡眠、
// 0x81 與 0x89 是毒、0x82 與 0x8A 是烈焰,0x83..0x87 與 0x8B 什麼都不做。
// 錯因是憑「0x80 系列 = 睡眠」的印象往下填,沒把跳表逐格數完。
//
// ⚠ **0x83 在踩踏那條路徑上什麼都不做,不代表它無害。** 它是
// **電擊力場**,處理在**移動**那一段(`sub_48F4` → `sub_4834`):
// 走進去會印「Ouch! Electric field!」、全隊受傷、然後**被彈回原格**。
// 所以玩家從來沒有真的站上去過 —— 踩踏分派表當然看不到它。
// 一個效果沒出現在你正在讀的那張表裡,先想想它是不是在另一條路徑上。
const (
	DungeonPitTrapA  = 0x61
	DungeonPitTrapB  = 0x69
	DungeonBombTrapA = 0x62
	DungeonBombTrapB = 0x6A

	DungeonSleepA   = 0x80
	DungeonSleepB   = 0x88
	DungeonPoisonA  = 0x81
	DungeonPoisonB  = 0x89
	DungeonFireA     = 0x82
	DungeonFireB     = 0x8A
	DungeonElectricA = 0x83
	DungeonElectricB = 0x8B
)

// DungeonFieldTile 是四個 `*Grav` 咒語在地牢裡放出來的力場格子。
//
// 索引就是 `sub_1994C` 傳給 `sub_18A08` 的種類碼:
// 0 = In Flam Grav、1 = In Nox Grav、2 = In Zu Grav、3 = In Sanct Grav。
// 值取自 `byte_55E24`(FM Towns 0x55E24:`82 81 80 83`)。
var DungeonFieldTile = [4]byte{
	DungeonFireA, DungeonPoisonA, DungeonSleepA, DungeonElectricA,
}

// DungeonHoleAbove 是「頭上有個洞」的位元。
//
// 掉進陷阱坑之後 `sub_4EB8` 對落點 `or bl, 8`,而 `sub_417C` 的 Klimb
// 就是看這一位元(外加 `byte_3DFBB`,疑為繩索)決定爬不爬得上去。
const DungeonHoleAbove = 0x08

// Dungeon 是一座地牢的 8 層。
type Dungeon struct {
	// Levels[層][y*8+x];層 0 是最上面一層。
	Levels [DungeonLevels][DungeonLevelB]byte
}

// DungeonSet 是全部 8 座。
type DungeonSet struct {
	Dungeons [DungeonCount]Dungeon
}

// LoadDungeons 讀 `DUNGEON.DAT`。
func LoadDungeons(dir string) (*DungeonSet, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "DUNGEON.DAT"))
	if err != nil {
		return nil, err
	}
	return ParseDungeons(raw)
}

// ParseDungeons 解析 `DUNGEON.DAT` 的內容。
func ParseDungeons(raw []byte) (*DungeonSet, error) {
	if len(raw) != DungeonFileB {
		return nil, fmt.Errorf("DUNGEON.DAT %d B,預期 %d B", len(raw), DungeonFileB)
	}
	s := &DungeonSet{}
	for d := 0; d < DungeonCount; d++ {
		for l := 0; l < DungeonLevels; l++ {
			off := d*DungeonBlockB + l*DungeonLevelB
			copy(s.Dungeons[d].Levels[l][:], raw[off:off+DungeonLevelB])
		}
	}
	return s, nil
}

// At 取某座地牢某層某格。超出範圍回牆。
func (s *DungeonSet) At(dungeon, level, x, y int) byte {
	if s == nil || dungeon < 0 || dungeon >= DungeonCount ||
		level < 0 || level >= DungeonLevels ||
		x < 0 || x >= DungeonSide || y < 0 || y >= DungeonSide {
		return DungeonWall
	}
	return s.Dungeons[dungeon].Levels[level][y*DungeonSide+x]
}

// Set 寫回某一格(踩過的陷阱要清掉、掉下去要打洞)。
func (s *DungeonSet) Set(dungeon, level, x, y int, v byte) {
	if s == nil || dungeon < 0 || dungeon >= DungeonCount ||
		level < 0 || level >= DungeonLevels ||
		x < 0 || x >= DungeonSide || y < 0 || y >= DungeonSide {
		return
	}
	s.Dungeons[dungeon].Levels[level][y*DungeonSide+x] = v
}

// DungeonKind 取高四位元。
func DungeonKind(tile byte) byte { return tile & 0xF0 }

// DungeonBlocks 回報這一格走不走得過去。
//
// 依據是怪物移動那一段(`sub_AC40` 的地牢版,反編譯在 0x4DC8 附近):
// `kind != 0x60 && kind != 0x80 && kind < 0xA0` 才放行。
// 也就是 **0x00..0x50 與 0x70、0x90 可走**;陷阱(0x60)與魔法陷阱(0x80)
// 怪物會避開,牆與門與房間(≥ 0xA0)走不過去。
//
// ⚠ 玩家**踩得到**陷阱 —— 那條路徑另外走(見 `DungeonPlayerBlocks`)。
func DungeonBlocks(tile byte) bool {
	k := DungeonKind(tile)
	return k == DungeonTrap || k == DungeonMagic || k >= DungeonRoomA
}

// DungeonPlayerBlocks 是玩家能不能走進去。`backing` 為真時是後退。
//
// `sub_48F4` 的判定,逐條照抄:
//
//	前進:0xA0 < kind < 0xE0 擋下 —— 也就是 **0xB0、0xC0、0xD0** 三種
//	後退:上面三種之外,**再加 0xA0 與 0xF0**(房間)
//
// 也就是說:
//
//	陷阱與力場不擋 —— 那是踩上去才觸發的
//	房間(0xA0 / 0xF0)**可以走進去,但不能倒退進去**
//	門口(0xE0)兩個方向都通
//
// ⚠ 「後退不能進房間」很容易漏掉,而漏掉的話玩家可以倒著走進戰鬥,
// 原版不允許。
func DungeonPlayerBlocks(tile byte, backing bool) bool {
	k := DungeonKind(tile)
	if k > DungeonRoomA && k < DungeonDoorway {
		return true
	}
	if backing && (k == DungeonRoomA || k == DungeonRoomF) {
		return true
	}
	return false
}

// DungeonCanTurn 回報站在這一格能不能左右轉(`sub_48F4` 的「Not in doorway!」)。
//
// 站在門口(0xE0)時左轉右轉都被擋下 —— 門框太窄,轉不了身。
// **轉身 180 度不受限**(那一支走 default case,沒有這道檢查)。
func DungeonCanTurn(tile byte) bool { return DungeonKind(tile) != DungeonDoorway }

// DungeonWrap 把地牢座標繞回 0..7(`sub_48F4` 的 `if (v<0) v=7; if (v>7) v=0`)。
//
// ⚠ 原版是**繞的**,不是撞牆。走出東邊會從西邊出來。
func DungeonWrap(v int) int {
	switch {
	case v < 0:
		return DungeonSide - 1
	case v >= DungeonSide:
		return 0
	}
	return v
}

// DungeonIsRoom 回報這一格是不是地牢房間。
func DungeonIsRoom(tile byte) bool {
	k := DungeonKind(tile)
	return k == DungeonRoomA || k == DungeonRoomF
}

// DungeonRoomNumber 取房號(低四位元,`sub_42CC` 的 `and al, 0Fh`)。
func DungeonRoomNumber(tile byte) int { return int(tile & 0x0F) }

// DungeonRoomClearedMask 是「這間房清過了」之後套在 tile 上的遮罩
// (原版 `sub_FA7C` 的 `and byte ptr [esi], 0AFh`)。
//
// ★ `0xAF` 清掉 **0x50** 兩個位元 ⇒ `0xFn` 變成 `0xAn`,也就是
// `DungeonRoomA` 那一族(可走的空房間),不是通道也不是牆。
const DungeonRoomClearedMask = 0xAF

// dungeonRoomAlwaysArmed 是「永遠不會被記成清過」的六間房。
//
// 原版 `byte_55110`(6 筆,筆數在 `byte_55116`):
//
//	50h 5Bh 41h 46h 4Bh 4Ch
//
// 鍵 = `房號 | ((地點碼 & 0x0F) << 4)` —— ⚠ 用的是**原始的低四位元**,
// 不是 `DungeonRoomBlock` 那個有修正的索引(見 `game/dungeonroom.go`)。
//
// 地點碼 0x21 是第一座地牢,所以低四位元 4 / 5 = 地點碼 0x24 / 0x25
// = 索引 3 / 4 = **謬誤(WRONG)** 與 **貪婪(COVETOUS)**:
//
//	謬誤   房 1、6、11、12
//	貪婪   房 0、11
//
// ⚠ 這六筆是**資料不是規則** —— 不要試著從「哪幾間該有怪」推它,
// 反過來也不要「順手補齊」成整座地牢。
var dungeonRoomAlwaysArmed = [6]byte{0x50, 0x5B, 0x41, 0x46, 0x4B, 0x4C}

// DungeonRoomAlwaysArmed 回報這一間房是不是永遠有怪。
func DungeonRoomAlwaysArmed(location, room int) bool {
	key := byte(room) | byte((location&0x0F)<<4)
	for _, k := range dungeonRoomAlwaysArmed {
		if k == key {
			return true
		}
	}
	return false
}

// DungeonCanClimbUp / Down 是 Klimb 的判定(`sub_417C`)。
//
// 往上:梯子(0x10 / 0x30),或這一格有「頭上的洞」而且身上有繩索。
// 往下:梯子(0x20 / 0x30),或陷阱坑(0x60 —— 掉下去的洞可以自己爬)。
func DungeonCanClimbUp(tile byte, hasRope bool) bool {
	switch DungeonKind(tile) {
	case DungeonLadderUp, DungeonLadderBoth:
		return true
	}
	return hasRope && tile&DungeonHoleAbove != 0
}

// DungeonCanClimbDown 見 DungeonCanClimbUp。
func DungeonCanClimbDown(tile byte) bool {
	switch DungeonKind(tile) {
	case DungeonLadderDown, DungeonLadderBoth, DungeonTrap:
		return true
	}
	return false
}

// 地牢房間(`DUNGEON.CBT`)
//
// `sub_42CC` 的位移算式:
//
//	房號     = tile & 0x0F
//	地牢索引 = 地點 − 0x21;**索引 ≥ 1 的再減一**
//	位移     = 地牢索引 × 11 × 512 + 房號 × 11 × 32
//	         = 地牢索引 × 5632 + 房號 × 352
//
// 一間房 352 B —— 與地表的戰鬥地圖同一個格式,所以直接沿用 CombatMap。
// 一座地牢 16 間房 = 5632 B,檔案 39424 B = **7 個區塊**。
//
// ⚠ 八座地牢卻只有七個區塊:`if (idx >= 1) idx--` 讓**前兩座共用同一批房間**。
// 這是原版的算式,照抄。
const (
	DungeonRoomsPerDungeon = 16
	DungeonRoomBlockB      = DungeonRoomsPerDungeon * CombatMapSize // 5632
)

// DungeonRoomBlock 把地點編號換成 `DUNGEON.CBT` 裡的區塊索引。
func DungeonRoomBlock(location int) int {
	idx := location - DungeonLocationBase
	if idx >= 1 {
		idx--
	}
	return idx
}

// LoadDungeonRooms 讀 `DUNGEON.CBT`,回傳全部房間(格式與地表戰鬥地圖相同)。
func LoadDungeonRooms(dir string) (*CombatMapSet, error) {
	return LoadCombatMaps(filepath.Join(dir, "DUNGEON.CBT"))
}

// DungeonRoomIndex 把(地點, 格子)換成 `DUNGEON.CBT` 裡的房間索引。
func DungeonRoomIndex(location int, tile byte) int {
	return DungeonRoomBlock(location)*DungeonRoomsPerDungeon + DungeonRoomNumber(tile)
}

// 八座地牢的地表入口
//
// 地點表其實有 **40 筆**,不是 32 —— `sub_2D564` 從索引 **0x20 掃到 0x27**
// 找座標與玩家相符的那一筆,用的是同一組 `byte_410F4`(X)/ `byte_4111C`(Y)/
// `off_41054`(名稱)。前 32 筆是城鎮與城堡(見 `locations.go`),
// 後 8 筆就是地牢。
//
// ⚠ Doom(索引 0x27)還有一道額外的門檻:`byte_3E0D8 & 3E0D9 & 3E0DA`
// (三個都是什麼還沒定,像是三把鑰匙或三個護符)全部 ≥ 0x80 才進得去,
// 否則印「Attacked at entrance!」並直接開打。那三個位元組還沒追到,
// 所以這條**還沒實作**。
type DungeonEntrance struct {
	X, Y int
	Name string
	// NameZH 是中譯。譯名對齊 u4-cht / u6-cht 的系列共通名。
	NameZH string
}

// DungeonEntrances 依地牢索引排列(0 = Deceit … 7 = Doom)。
var DungeonEntrances = [DungeonCount]DungeonEntrance{
	{240, 73, "DECEIT", "欺瞞"},
	{91, 67, "DESPISE", "輕蔑"},
	{72, 168, "DESTARD", "毀滅"},
	{126, 20, "WRONG", "謬誤"},
	{156, 27, "COVETOUS", "貪婪"},
	{58, 102, "SHAME", "羞恥"},
	{239, 240, "HYTHLOTH", "海斯洛斯"},
	{128, 128, "DOOM", "末日"},
}

// DungeonAt 回報世界座標上有沒有地牢入口。
func DungeonAt(x, y int) (int, bool) {
	for i := range DungeonEntrances {
		if DungeonEntrances[i].X == x && DungeonEntrances[i].Y == y {
			return i, true
		}
	}
	return 0, false
}

// DisplayName 是訊息欄要印的名字。
func (e DungeonEntrance) DisplayName() string {
	if e.NameZH == "" {
		return e.Name
	}
	return e.NameZH + "(" + e.Name + ")"
}

// 寶箱的獎品表(DOS `0x4134` / `0x413C` / `0x4144`,各 8 B)
//
// `sub_15020(等級, …)` 由 i = 7 往 0 掃:
//
//	等級 >= ChestThreshold[i] 且 random(1,30) >= ChestThreshold[i]
//	    → 數量 = (ChestMax[i] == 1) ? 1 : random(1, ChestMax[i])
//	      在寶箱那一格生一個種類 ChestKind[i] 的物件
//
// 也就是**門檻同時當「等級下限」與「擲骰難度」**:門檻 3 的那一項幾乎必中,
// 門檻 25 的那一項要高等寶箱加上好運。
//
// ⚠ 八個種類碼(1/2/3/4/7/8/13/15)用的是**地圖物件的種類編號**,
// 與 `docs/re/11` 那一套同源;`sub_14F68` 只是把它們生在地上讓玩家撿。
// 只有種類 **2** 的語意確定:數量是 `random(1, 等級×3)`、上限 90 ——
// U5 的寶箱裡只有金幣是這樣長的。其餘七種還沒對出來,所以引擎照原版
// **生成物件**而不去猜它是什麼。
var (
	// ChestKind 的八個值就是 `sub_154BC` 的**物品碼**(見 `get.go`):
	//
	//	 1  上鎖的箱子   2  金幣   3  藥水    4  卷軸
	//	 7  鑰匙        8  寶石  13  火把   15  食物
	//
	// ★ 這一組原本標「語意未定」。逆完 Get 指令之後對上了 ——
	// 而且反過來也是佐證:兩張獨立的表(地表寶箱 vs `sub_154BC` 的跳表)
	// 用同一組編碼。
	ChestKind      = [8]byte{1, 2, 3, 4, 7, 8, 13, 15}
	ChestThreshold = [8]byte{25, 3, 17, 17, 9, 15, 7, 7}
	ChestMax       = [8]byte{10, 90, 8, 8, 2, 2, 2, 2}
)

// ChestKindGold 是唯一語意確定的那一種(數量 `random(1, 等級×3)`)。
const ChestKindGold = 2

// ChestTrapped 是寶箱「有陷阱」的位元。
//
// 地牢寶箱看格子的 **bit 0**(`sub_18D18` 的 `test byte [eax], 1`);
// 地表寶箱看物件 +5 的 **bit 7**(`sub_15108` 的 `cmp al, 7Fh; ja`,
// 而 An Sanct 解除的方式是 `and byte [ebx+5], 7Fh`)。
const (
	ChestTrappedDungeon = 0x01
	ChestTrappedWorld   = 0x80
)

// DungeonOpenedChest 是寶箱打開之後那一格變成什麼。
//
// `sub_18D18`:`tile = (tile & 8) | 0x70` —— 保留「頭上有洞」那一位元,
// 其餘換成 0x70。所以 **0x70 不是門,是「開過的寶箱」**。
func DungeonOpenedChest(tile byte) byte {
	return (tile & DungeonHoleAbove) | 0x70
}
