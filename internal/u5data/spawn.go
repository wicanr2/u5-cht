package u5data

// 大地圖的遊蕩怪生成(原版 `sub_1F98` / `sub_203C` / `sub_2008` / `sub_215C` / `sub_2B518`)
//
// 完整推導與表格見 `docs/re/82`。這一檔只放**純資料與純函式** ——
// 需要物件槽或隊伍座標的部分在 `internal/game`。

// EncounterRollLo / Hi 是遭遇擲骰的範圍(原版 `sub_28E14(1, 1Eh)`)。
//
// ⚠⚠ 判定是**嚴格小於門檻**(`cmp eax, ebx; jge → 跳過`)。所以**門檻 1 等於
// 不會發生** —— `random(1, 30)` 的最小值就是 1,而 `1 < 1` 不成立。
// 草地的門檻正好是 1 ⇒ **白天在草地上走路,原版永遠不會有隨機遭遇。**
//
// 這一條很容易寫成 `<=` 而看不出差別(平原上偶爾遇敵「感覺也很合理」),
// 但那會讓整個世界比原版兇。`TestGrasslandNeverSpawnsByDay` 釘住它。
const (
	EncounterRollLo = 1
	EncounterRollHi = 30
)

// EncounterThreshold 是這一格、這個時刻的遭遇門檻(原版 `sub_1F98`)。
//
//	if (幽冥界) return 3                       ; ★ 固定 3,不加夜間加成
//	tile 0x20..0x26            → 0            ; ★ 這一族完全不生怪
//	tile 4 或 9..0x0F          → 2            ; 沼澤、林 / 丘 / 山
//	其餘                       → 1
//	凌晨(小時 < 5)            → 再 +3
//
// ⚠ 原版還有一條 `if (小時 >= 0x20)` 也走 +3,而小時只到 23 ⇒ **死碼**。
// 照原樣留著(參數用 int,所以呼叫端傳得進 32 以上時行為與原版一致)。
func EncounterThreshold(tile byte, hour int, underworld bool) int {
	if underworld {
		return 3
	}
	var base int
	switch {
	case tile >= 0x20 && tile <= 0x26:
		base = 0
	case tile == 4 || (tile >= 9 && tile <= 0x0F):
		base = 2
	default:
		base = 1
	}
	// ★ `hour >= 0x20` 是原版的死碼(小時只到 23),不要「順手」刪掉。
	if hour >= 0x20 || hour < 5 {
		base += 3
	}
	return base
}

// 四張「地形族 → 怪物」的加權表(原版 `byte_4FD1C` / `byte_4FD28` /
// `byte_4FD30` / `byte_4FD35`,權重在 `dword_4FD38` / `a88` /
// `dword_4FD4C` / `dword_4FD50+1`)。
//
// ★★ **四張的權重都恰好加到 256** —— 那是「怪物清單與權重清單配對正確」的
// 硬佐證,不是巧合。任何改動之後 `TestSpawnWeightsSumTo256` 會抓到。
//
// ★ 陸地・地表那張的順序本身就是難度曲線:Orc 23.4% → … → Dragon 0.8%
// → Daemon 0.4%。**弱的常見、強的稀有**,而最後三筆(3 / 2 / 1)幾乎見不到。
//
// ⚠ `a88` 在反組譯檔裡被 IDA 命名成字串 `'@88  ',8,8,0` —— 那是**權重表**
// (`40 38 38 20 20 08 08 00`),不是字串。同 `docs/re/67` 的 `off_48A88`
// 與 `docs/re/77` 的 `aBoxHow`:**`aXxx` 這種自動名字不是資料**。
var (
	// SpawnLandSurface 是地表陸地的 12 種(`byte_4FD1C`)。
	SpawnLandSurface = []byte{0xC0, 0xC8, 0x90, 0x98, 0xBC, 0xC4, 0xD0, 0xE4, 0xCC, 0xD4, 0xDC, 0xD8}
	// SpawnLandSurfaceWeights 是對應權重(`dword_4FD38`)。
	SpawnLandSurfaceWeights = []byte{0x3C, 0x32, 0x28, 0x1E, 0x14, 0x0F, 0x0F, 0x0A, 0x0A, 0x03, 0x02, 0x01}

	// SpawnLandUnderworld 是幽冥界陸地的 8 種(`byte_4FD28`)。
	// ⚠ 最後一筆是 0x00 且權重是 0 ⇒ **選不到**。照原樣保留:
	// 刪掉它會讓表長度與權重表不一致,而那個一致性正是配對的佐證。
	SpawnLandUnderworld = []byte{0x94, 0x90, 0x98, 0xF0, 0xF4, 0xD8, 0xDC, 0x00}
	// SpawnLandUnderworldWeights 是對應權重(`a88`)。
	SpawnLandUnderworldWeights = []byte{0x40, 0x38, 0x38, 0x20, 0x20, 0x08, 0x08, 0x00}

	// SpawnWaterSurface 是地表水域的 5 種(`byte_4FD30`)。最後一筆 0x2C 是**敵船**。
	SpawnWaterSurface = []byte{0x8C, 0x84, 0x88, 0x80, 0x2C}
	// SpawnWaterSurfaceWeights 是對應權重。
	//
	// ⚠⚠ 原版的權重只有**四個位元組**(`dword_4FD4C` = `48 48 28 26`,和 222),
	// 第五個是**溢出到下一個變數**取到的:`sub_2008` 沒有邊界檢查,擲到 222 以上時
	// 會讀 `dword_4FD50` 的第 0 個位元組 `0x22`(34),而 222 + 34 = 256。
	// 這就是為什麼 IDA 把水域幽冥界的表標成 `dword_4FD50+1` —— **+0 屬於上一張表**。
	// 這裡把 0x22 明確寫成第五個權重,行為相同而意圖看得見。
	SpawnWaterSurfaceWeights = []byte{0x48, 0x48, 0x28, 0x26, 0x22}

	// SpawnWaterUnderworld 是幽冥界水域(`byte_4FD35`)。
	// ⚠ 索引 0(0x84 Squid)的權重是 **0** ⇒ 選不到;實際只有 50% 海蛇 / 50% 空手。
	SpawnWaterUnderworld = []byte{0x84, 0x88, 0x00}
	// SpawnWaterUnderworldWeights 是對應權重(`dword_4FD50+1` = `00 80 80`)。
	SpawnWaterUnderworldWeights = []byte{0x00, 0x80, 0x80}
)

// 三個地形特例(原版 `sub_203C` 的三條提早 return)。
const (
	// SpawnDeepWaterSpecial 是深水(tile 1)`random(0, 7) == 7` 時的那一種。
	// ⚠ 名字在 `DATA.OVL` 的生物名表裡是**空的**(索引 43,與 42 一起空著)——
	// 語意未定,照編號實作、不猜(`CLAUDE.md §3.0`)。
	SpawnDeepWaterSpecial = 0xEC
	// SpawnScorchedSpecial 是焦灼荒漠(tile 7)的 Sand Trap;`random(0, 3) != 0` 就不生。
	SpawnScorchedSpecial = 0xE0
	// SpawnUnderworldSwamp 是幽冥界沼澤(tile 4)的 Rot Worm。
	SpawnUnderworldSwamp = 0xF8
	// SpawnEnemyShip 是水域表最後那一筆 —— 敵船,不是生物。
	//
	// ★ 判準不是猜的:`sub_2218` 在 `kind == 0x2C` 時額外把物件槽的
	// **+5(`ObjQuality`)設成 100**,而 `docs/re/11` 已經釘死 +5 是船的耐久。
	// 而且它只生在 `tile & 0xF0 == 0x60` 那一族水上。
	SpawnEnemyShip = 0x2C
	// SpawnEnemyShipHull 是敵船的初始耐久(`0x64`)。
	SpawnEnemyShipHull = 100
)

// SpawnPickWeighted 依權重表挑一個索引(原版 `sub_2008`)。
//
//	r = random(0, 0xFF)
//	i = 0
//	while (weights[i] <= r) { r -= weights[i]; i++ }
//	return i
//
// ⚠ 原版**沒有邊界檢查** —— 它靠「權重加起來剛好 256」保證走不出去。
// 這裡多一道 `i < len` 的保險,並回傳最後一個索引;真的走到那裡就是資料壞了,
// 而測試已經釘住四張表都加到 256。
func SpawnPickWeighted(weights []byte, roll int) int {
	r := roll
	for i := 0; i < len(weights); i++ {
		if int(weights[i]) > r {
			return i
		}
		r -= int(weights[i])
	}
	return len(weights) - 1
}

// SpawnTerrainKind 是 `sub_203C` 開頭那個地形分類。
type SpawnTerrainKind int

// 兩族地形 + 一族不生怪。
const (
	// SpawnTerrainWater 是水路那一族:tile < 4、0x60–0x6F、0xD4–0xD7、0xE4–0xE7。
	// ⚠ 「0xE4–0xE7」在**地圖 tile** 的意義下才成立;那不是怪物編號。
	SpawnTerrainWater SpawnTerrainKind = iota
	// SpawnTerrainLand 是其餘可生怪的地形。
	SpawnTerrainLand
	// SpawnTerrainNone 是明確不生怪的:山(0x0C)、高峰(0x0D)、
	// 以及 tile >= 0x10 而不在 0x30–0x33 的全部。
	SpawnTerrainNone
)

// SpawnTerrainOf 把地圖 tile 分到三族(原版 `sub_203C` 的前兩段判斷)。
func SpawnTerrainOf(tile byte) SpawnTerrainKind {
	if tile < 4 ||
		(tile >= 0x60 && tile <= 0x6F) ||
		(tile >= 0xD4 && tile <= 0xD7) ||
		(tile >= 0xE4 && tile <= 0xE7) {
		return SpawnTerrainWater
	}
	// 陸路:三條明確不生的先擋掉。
	if tile == 0x0C || tile == 0x0D {
		return SpawnTerrainNone
	}
	// ★ tile >= 0x10 只有 0x30..0x33 能生(原版 `and eax, 0FCh; cmp eax, 30h`)。
	if tile >= 0x10 && tile&0xFC != 0x30 {
		return SpawnTerrainNone
	}
	return SpawnTerrainLand
}

// SpawnWaterFamily 回報這一格屬不屬於敵船能生的那一族水(`tile & 0xF0 == 0x60`)。
func SpawnWaterFamily(tile byte) bool { return tile&0xF0 == 0x60 }

// WaterSpawnGiveUpRollHi / Keep 是水域那一族額外的「大多數時候放棄」判定
// (原版 `random(0, 0x40) >= 0x10 → return 0`)。
//
// ⚠ 是 `random(0, 64)`(**65 個值**)而不是 0..63,所以留下的比例是
// 16/65 ≈ 24.6%,不是剛好 1/4。差別很小但寫成 `random(0,63)` 就不是原版。
const (
	WaterSpawnGiveUpRollHi = 0x40
	WaterSpawnKeepBelow    = 0x10
)

// IsCreatureTile 回報物件槽的 `ObjKind` 位元組算不算生物(原版 `sub_22B0`)。
//
//	0x2C..0x2F  → 是      ★ 船也算(世界回合要讓敵船移動)
//	0x30..0x7F  → 不是
//	0x80..0xB3  → 是
//	0xB4..0xB7  → 不是    ★ 四個信物
//	0xB8..0xE7  → 是
//	0xE8..0xEB  → 不是
//	0xEC..0xFF  → 是
//	< 0x2C      → 不是
//
// ⚠⚠ **與 `MapObject.IsCreature()` 不是同一個判準。** 那一支判的是
// `Kind >= 0x40 && Kind != 0xFC`,範圍差很多(0x40–0x7F 在這裡不算生物)。
// 兩支各有出處,不要合併 —— 世界回合要用的是這一支。
func IsCreatureTile(kind byte) bool {
	switch {
	case kind >= 0x2C && kind <= 0x2F:
		return true
	case kind < 0x80:
		return false
	case kind <= 0xB3:
		return true
	case kind <= 0xB7:
		return false
	case kind <= 0xE7:
		return true
	case kind <= 0xEB:
		return false
	}
	return true
}

// 生怪落點的兩個限制(原版 `sub_215C`)。
//
//	X = (視窗左上角X + random(0, 0x1F)) & 0xFF
//	Y = (視窗左上角Y + random(0, 0x1F)) & 0xFF
//	重擲直到:abs(X − 隊伍X) > 6 且 abs(Y − 隊伍Y) > 6
//	          而且 abs(…) < 0xFA(這一條是**環繞**:−253 其實等於 +3)
//
// ★ 兩條加起來的意思是「**在載入視窗內,但兩軸都離隊伍 7 格以上**」——
// 怪不會在眼前冒出來。
const (
	SpawnWindowSpan   = 0x1F
	SpawnMinDistance  = 6    // abs <= 6 → 太近,重擲
	SpawnWrapDistance = 0xFA // abs >= 0xFA → 環繞之後也太近
)

// SpawnFarEnough 回報候選座標離隊伍夠不夠遠(原版 `sub_215C` 的四道 while)。
func SpawnFarEnough(dx, dy int) bool {
	ax, ay := abs8(dx), abs8(dy)
	if ax <= SpawnMinDistance || ay <= SpawnMinDistance {
		return false
	}
	return ax < SpawnWrapDistance && ay < SpawnWrapDistance
}

func abs8(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// 找可用物件槽的優先序(原版 `sub_2B57C` 九段串接 → `sub_2B518`)。
//
// ⚠⚠ **只掃槽 1..0x17(23)**,而世界回合掃到 0x1F(31)⇒ **高 8 槽是保留區**:
// 生怪永遠不會用到它們,但世界回合照樣讓裡面的東西移動。這個不對稱是原版的,
// 別「順手」把兩邊對齊。
const (
	// SpawnSlotFirst / SpawnSlotLast 是生怪可以用的槽範圍(槽 0 是隊伍)。
	SpawnSlotFirst = 1
	SpawnSlotLast  = 0x17
	// SpawnSlotProtected 是**永不被回收**的種類碼(原版 `cmp al, 0B5h; jz → 跳過`)。
	// ⚠ 0xB4–0xB7 是四個信物,而原版**只保護 0xB5 一個**。看起來像遺漏,
	// 但 `CLAUDE.md §3.0` 說不自行補齊 —— 照原樣只擋 0xB5。
	SpawnSlotProtected = 0xB5
	// SpawnSlotNearRadius 是「太近所以不回收」的半徑(原版 `+5` 之後比 `> 0x0A`)。
	SpawnSlotNearRadius = 5
)

// SpawnSlotPass 是找槽的一趟掃描條件。
type SpawnSlotPass struct {
	Lo, Hi      byte // 可以被覆寫的 ObjKind 範圍(含兩端)
	RequireFar  bool // 只回收離隊伍 5 格以外的
	Description string
}

// SpawnSlotPasses 是原版 `sub_2B57C` 的九趟,順序不能動。
//
// ★ 順序本身是設計:**先找空槽,再犧牲物品(1–0x0F),然後才是生物(0x80+)**,
// 而且前四趟都要求「離隊伍夠遠」—— 玩家眼前的東西不會突然消失。
// 最後一趟 `{0, 0xFF, false}` 保證一定找得到(會覆寫槽 1)。
var SpawnSlotPasses = []SpawnSlotPass{
	{0x00, 0x00, false, "空槽"},
	{0x01, 0x0F, true, "遠處的物品"},
	{0x80, 0xFF, true, "遠處的生物"},
	{0x10, 0x11, true, "遠處的坐騎"},
	{0x30, 0x7F, true, "遠處的其他"},
	{0x01, 0x0F, false, "任何物品"},
	{0x80, 0xFF, false, "任何生物"},
	{0x10, 0x11, false, "任何坐騎"},
	{0x30, 0x7F, false, "任何其他"},
	{0x00, 0xFF, false, "隨便一個"},
}

// CreatureVanishTile 是「走上去就消失」的那一格(原版 `sub_2870` 尾段
// `cmp byte ptr [eax], 0DCh` → 種類碼與 tile 都歸零)。
//
// ⬜ 那是什麼地形還沒查(要抽 FM Towns 的 `EGA*.TIL` 才看得到圖)。
// 照編號實作、語意留白。
const CreatureVanishTile = 0xDC

// 免疫地形延遲的四種(原版 `sub_2870` 的四個 `cmp eax, …; jz def_295C`)。
//
// ★ 四種都會飛 —— 龍、蝙蝠、惡魔、Mongbat。這不是巧合,是設計:
// 沼澤與山地拖不慢飛行生物。
const (
	FlyerDragon  = 0xDC
	FlyerBat     = 0x94
	FlyerDaemon  = 0xD8
	FlyerMongbat = 0xF0
)

// CreatureIgnoresTerrain 回報這一種吃不吃地形延遲。
func CreatureIgnoresTerrain(kind byte) bool {
	return kind == FlyerDragon || kind == FlyerBat || kind == FlyerDaemon || kind == FlyerMongbat
}
