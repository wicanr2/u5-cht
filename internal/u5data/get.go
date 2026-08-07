package u5data

// Get 指令:撿東西(原版 `sub_15A94` → `sub_154BC`,地牢走 `sub_15930`)
//
// 這是三塊寶石碎片、檀香木盒、王冠、權杖、護符、月石、魔毯的**唯一**入口 ——
// 少了它,真結局那條路根本走不到,而引擎在此之前完全沒有這個指令。
//
// 兩條路:那一格上**有物件**就撿物件,沒有就看**地形**能不能拿(牆上的火把、
// 田裡的作物、桌上的食物)。兩條都不成立才是「這裡沒有東西可拿。」

// 物件種類碼 → 撿到什麼(`sub_154BC` 的跳表與其後的比較鏈)。
const (
	ItemClosedChest  = 0x01 // 「得先打開它!」
	ItemGold         = 0x02
	ItemPotion       = 0x03 // 藥水;品質選顏色(byte_3E038[品質])
	// ItemScroll 是卷軸(byte_3E030[品質])。
	//
	// ⚠ **圖紙不是自己的種類碼** —— 它是這一類裡品質 0xFF 的那一筆
	// (`sub_154BC` 的 `cmp edi, 0FFh`,見 ItemPlansQuality)。
	// 把種類 4 一律當成圖紙的話,隨便一捲卷軸都會變成攻城圖。
	ItemScroll       = 0x04
	// ItemKindEquipMax 是「這個種類碼走裝備那一條」的上限。
	//
	// 跳表涵蓋 1..0x0B,再加上比較鏈的 0x0C —— 落在其中而又沒有專屬分支的
	// (5、6、9、10、0x0B、0x0C)全部進 `byte_3DFD0[品質]`,品質就是裝備編號。
	ItemKindEquipMax = 0x0C
	ItemKey          = 0x07
	ItemGem          = 0x08
	ItemTorch        = 0x0D
	ItemSandalwood   = 0x0E // ★ 檀香木盒 —— 真結局的條件
	ItemFood         = 0x0F
	ItemMoonstone    = 0x19
	ItemMagicCarpet  = 0x1B
	ItemShardBase    = 0xB4 // 0xB4..0xB7:碎片 / 王冠 / 權杖 / 護符
	ItemShard        = 0xB4
	ItemCrown        = 0xB5
	ItemSceptre      = 0xB6
	ItemAmulet          = 0xB7
)

// GetPickable 回報那一格的物件撿不撿得起來(`sub_15A94` 的四個條件)。
//
// 原版是一串 `break` 條件,不是白名單:
//
//	種類 < 0x10          撿(一般物品:錢、藥水、鑰匙、寶石、火把、盒子、食物…)
//	種類 == 0x19         月石
//	種類 == 0x1B         魔毯
//	(種類 & 0xFC) == 0xB4 碎片 / 王冠 / 權杖 / 護符
//	其餘                 跳過,繼續找下一個槽
//
// ⚠ 「其餘」包含**怪物與坐騎**(種類 ≥ 0x40)—— 站在馬旁邊按 Get 不會把馬
// 撿起來,而是繼續掃後面的槽。這個「繼續掃」而不是「就此放棄」的差別,
// 決定了同一格疊了兩樣東西時撿到哪一個。
func GetPickable(kind byte) bool {
	switch {
	case kind < 0x10:
		return true
	case kind == ItemMoonstone, kind == ItemMagicCarpet:
		return true
	case kind&0xFC == ItemShardBase:
		return true
	}
	return false
}

// ShardIndex 是這個碎片物件對應第幾塊(原版 `mov eax, edi; and eax, 3`)。
//
// ⚠ 用的是物件的**品質欄**(槽位移 +5),不是種類碼。四個 0xB4..0xB7 共用
// 同一段程式碼,種類碼只決定「是碎片還是王冠」,是哪一塊碎片看品質欄。
func ShardIndex(quality int) int { return quality & 3 }

// 地形能不能拿(`sub_15A94` 後半的 switch)。
const (
	// TileWallTorchA / TileWallTorchB 是牆上的火把。拿下來換成磚地。
	TileWallTorchA = 0xB0
	TileWallTorchB = 0xB1
	// TileCrops / TileCropsPicked 是田裡的作物與收割後的空地。
	TileCrops       = 0x2D
	TileCropsPicked = 0x2C
	// 桌上的食物:一張桌子橫跨三格,中間那格拿走之後會變成左半或右半。
	TilePlateMiddle = 0x9C
	TilePlateNorth  = 0x9A
	TilePlateSouth  = 0x9B
	TilePlateEmpty  = 0x95
)

// BorrowedTorchTurns 是從牆上拿下來的火把能撐多久(`mov byte_3E0B7, 64h`)。
//
// = 100 分鐘,固定值 —— 與自己點一把(`random(0,15) + 0x70`,112..127)不同,
// 而且訊息是「借用!」不是「點亮了」。
const BorrowedTorchTurns = 100

// GetKarmaPenalty 是拿食物扣的業報(`dec byte_3E098`)。
//
// ★ **牆上的火把不扣**(那條路直接跳到結尾,原版還特地說「借用!」),
// **作物與桌上的食物扣 1**。這就是原版對「偷竊」的定義 ——
// 不是有沒有人看到,而是拿了誰的東西。
const GetKarmaPenalty = 1

// PlateReach 回報從 (dx, dy) 這個方向伸手拿不拿得到這一格的食物。
//
// 原版對三種盤子各有各的規矩,而且**只看 Y**:
//
//	0x9A(北半)  要從南邊拿(dy == 1)
//	0x9B(南半)  要從北邊拿(dy == -1)
//	0x9C(整張)  不能從東西向拿(dx 必須是 0)
//
// 拿不到時印「碰不到那個盤子!」。也就是說站在桌子側面伸手是搆不著的 ——
// 得繞到桌子的長邊。
func PlateReach(tile byte, dx, dy int) bool {
	switch tile {
	case TilePlateNorth:
		return dy == 1
	case TilePlateSouth:
		return dy == -1
	case TilePlateMiddle:
		return dx == 0
	}
	return false
}

// PlateAfter 回傳拿走食物之後那一格變成什麼。
//
// 整張桌子(0x9C)被拿走一半之後**變成剩下的那一半**,不是直接清空:
// 從南邊拿 → 剩南半(0x9B);從北邊拿 → 剩北半(0x9A)。
// 所以同一張桌子可以拿兩次。
func PlateAfter(tile byte, dy int) byte {
	if tile == TilePlateMiddle {
		if dy == 1 {
			return TilePlateSouth
		}
		if dy == -1 {
			return TilePlateNorth
		}
		return tile
	}
	return TilePlateEmpty
}

// DungeonGetLocationMin / Max 是「按 Get 要去開地牢寶箱」的地點編號範圍。
//
// 原版 `cmp al, 20h; jbe` / `cmp al, 29h; jnb` —— 也就是 **33..40**,
// 八座地牢。落在這個範圍時 `sub_15A94` 一開頭就轉給 `sub_15930`。
const (
	DungeonGetLocationMin = 33
	DungeonGetLocationMax = 40
)

// 地牢寶箱的獎品(原版 `sub_15930`,三張並列表在 0x55DD4 / 0x55DDC / 0x55DE4)
//
// ★ 這一組**與地表寶箱(`sub_15020`)完全是兩套**。地表寶箱有「等級」
// (物件的品質位元組),地牢寶箱**沒有** —— 它只擲 `random(1, 樓層×4 + 4)`,
// 再拿七個門檻各比一次。`docs/re/` 先前記著「地牢寶箱的等級從哪來還沒找到」,
// 答案是:根本沒有那個東西,是另一支函式。
//
//	i  門檻  數量上限  種類碼  掉什麼
//	0    2      31      0x0F   食物
//	1    4   樓層×8     0x02   金幣      ← 數量上限是算出來的,不是查表
//	2    5       3      0x07   鑰匙
//	3   10       3      0x08   寶石
//	4   20       3      0x0D   火把
//	5   25       7      特例    藥水(種類碼 3,數量 random(0,7))
//	6   25       7      特例    卷軸(種類碼 4,數量 random(0,7))
//	7    0       0        0    表尾
var (
	// DungeonLootThreshold 是每一類的門檻:擲出來 ≥ 它才拿得到。
	DungeonLootThreshold = [7]byte{2, 4, 5, 10, 20, 25, 25}
	// DungeonLootMax 是數量的上限(索引 1 的金幣不用它,見上表)。
	DungeonLootMax = [7]byte{31, 0, 3, 3, 3, 7, 7}
	// DungeonLootKind 是種類碼;0xFF 代表由程式碼直接指定(藥水 / 卷軸)。
	DungeonLootKind = [7]byte{0x0F, 0x02, 0x07, 0x08, 0x0D, 0xFF, 0xFF}
)

// DungeonLootSpecial 是索引 5 / 6 那兩類寫死的種類碼。
var DungeonLootSpecial = map[int]byte{5: ItemPotion, 6: ItemScroll}

// DungeonLootRollMax 是那顆骰子的上限:`random(1, 樓層×4 + 4)`。
//
// 所以第 1 層(樓層 0)最多擲到 4 —— 只拿得到食物(門檻 2)與金幣(門檻 4);
// 火把要第 5 層以後、藥水與卷軸要第 6 層以後才可能出現。
func DungeonLootRollMax(floor int) int { return floor*4 + 4 }

// DungeonEmptiedChest 是把開過的寶箱清空之後那一格變成什麼。
//
// 原版 `and byte ptr [eax+ecx], 8` —— 只留「頭上有洞」那一位元。
func DungeonEmptiedChest(tile byte) byte { return tile & DungeonHoleAbove }

// 地下世界的固定寶物(原版 `sub_10B3C`,由地圖載入 `sub_2CBEC` 呼叫)
//
// ★ 「碎片與盒子擺在哪裡」這一題,碎片與護符的部分在這裡結掉了 ——
// 它們**不在**任何資料檔裡(`.OOL` 沒有、`DUNGEON.CBT` 也沒有),
// 是進地下世界時由程式**當場塞進物件槽**的:
//
//	if (樓層 == 0) return;                        // 只在地下世界
//	if (還沒拿到護符) 槽 28 = 護符 @ (105, 225)
//	for (i = 0; i < 3; i++) {
//	    if (已經拿到第 i 塊碎片) continue;
//	    if (第 i 位暗影君主已被消滅) continue;      // 用掉的碎片不會重生
//	    槽 29+i = 碎片 @ (表)
//	}
//
// 第二個條件很要緊:`sub_1A38C` 用碎片消滅暗影君主時會把碎片**清掉**
//(`byte_3DFC4[i] = 0`),如果只看「有沒有拿到」,那塊碎片會在原地重生。

// UnderworldItemSlot 是這幾樣東西各自佔的物件槽。
//
// 由位址算出來:`dword_3E54C` 與 `3E554h[i*8]` 相對於物件表起點 `dword_3E46C`
// 分別是 +0xE0 與 +0xE8,除以 8 就是槽 28 與 29..31。
const (
	UnderworldAmuletSlot   = 28
	UnderworldShardSlot = 29
)

// UnderworldOrb 是護符固定擺放的位置與品質。
var UnderworldOrb = struct {
	X, Y, Quality int
}{X: 105, Y: 225, Quality: 0xF3}

// UnderworldShards 是三塊碎片固定擺放的位置與品質(`byte_55140 + 0x110`)。
//
// ⚠ 品質是 **0xF0 / 0xF1 / 0xF2**,不是 0 / 1 / 2 —— `ShardIndex` 取的是
// 低兩位(`and eax, 3`),所以剛好對到 0 / 1 / 2。高位那幾個位元在這裡沒有用途,
// 但**不能把品質改寫成 0/1/2** :那會讓存檔與原版不同。
var UnderworldShards = [ShadowlordCount]struct {
	X, Y, Quality int
}{
	{X: 192, Y: 80, Quality: 0xF0},
	{X: 130, Y: 65, Quality: 0xF1},
	{X: 176, Y: 184, Quality: 0xF2},
}

// 兩個寫死在程式裡的 NPC 槽(`sub_154BC` 的魔毯與檀香木盒分支)
//
// 城裡的物品是 `.NPC` 檔裡生物編號 < 0x40 的槽,由 `sub_1E74` 鏡射進物件表
//(見 `docs/re/36`)。撿走之後要不要讓它復活,原版是**逐案硬編碼**的,
// 沒有通則 —— 而全遊戲撿得起來的物品型 NPC 只有**四個**:這兩個,
// 加下面的王冠與權杖(原本寫「加王冠」漏了權杖,見 `docs/re/57`)。
const (
	// SandalwoodNPCLocation / SandalwoodNPCSlot 是檀香木盒那一槽。
	//
	// ★ 這兩個數字不是從 `.NPC` 檔數出來的,是從程式碼算出來的:
	// 原版寫的是 `byte_3E3AF |= 0x80`,而永久移除遮罩的基底是 `dword_3E36C`,
	// 於是 `0x3E3AF = 0x3E36C + 16×4 + 3` → 陣列第 16 格(地點 17)的位元 31。
	// 資料檔那邊獨立給出同一個答案(`CASTLE.NPC` 地點 17 槽 31 生物編號 0x0E)。
	SandalwoodNPCLocation = 17
	SandalwoodNPCSlot     = 31

	// CarpetNPCLocation / CarpetNPCSlot 是不列顛王城堡二樓那張魔毯(`sub_268(0x16)`)。
	//
	// ⚠⚠ **原版只做暫時移除,沒有配套的 `sub_218`** —— 離開再回來毯子又在原地。
	// 這是可以刷的。照抄,不「修好」。
	CarpetNPCLocation = 17
	CarpetNPCSlot     = 22
)

// 王冠與權杖擺在哪裡(`.NPC` 檔裡各一槽,全遊戲各只有一個)
//
// ★ 這一題卡了很久,而**卡住的原因是我在錯的命名空間裡找**。
//
// 0xB4..0xB7 這四個號碼同時活在兩套索引裡,語意完全不同:
//
//	地形 tile 0xB4..0xB7 → `LOOK2.DAT` 的 look#180..183 = **四個朝向的加農砲**
//	物件種類 0xB4..0xB7 → look#436..439 = 碎片 / 王冠 / 權杖 / 護符
//
// 於是「在地圖裡 grep 0xB5」會撈到一堆**砲**:不列顛王城堡的上兩層、
// 亞拉拉特、邊境哨、巨蛇要塞,而且清一色左右對稱 —— 對稱正是它們不可能是
// 「全世界只有一個」的信物的證據。我當時看到對稱卻沒把它當線索。
//
// 找對命名空間之後答案是唯一的:掃四份 `.NPC` 的生物編號欄,
// 全遊戲只有兩槽是 0xB5 / 0xB6。
//
//	王冠 0xB5  `CASTLE.NPC` 地點 18(第二座城堡,大地圖 (196,245))槽 1
//	           排程三個 slot 全部 (15,13) 第 +3 層,行為型別 0(不動),四個時刻全 0
//	權杖 0xB6  `KEEP.NPC`   地點 29 STONEGATE 槽 9
//	           排程三個 slot 全部 (15,15) 地面層,同樣型別 0、時刻全 0
//
// 兩處都是密室:王冠那間是雉堞(0x4F)圍出來的小室,門口一格 0x97「怪門」、
// 兩側各一座火盆;權杖那格被八格 0x8C「鬆動的磚」團團圍住,外面再一圈石柱。
//
// # 為什麼不用寫「放置」程式碼
//
// 因為它們是 NPC。`sub_1E74` 每回合把「此刻在本層」的 NPC 鏡射進物件表,
// Get 掃的就是物件表 —— 這條路已經在跑了(見上面檀香木盒那段)。
// 引擎這邊**一行都不用加**:行為型別 0 = 原地不動,時刻全 0 = 不換崗位。
//
// ⚠ 而且兩者的善後**不一樣**:王冠走 `sub_2E0` + `sub_218` + `sub_268`
// 全套(永久移除),權杖只靠共同尾段清掉物件槽,**沒有 `sub_218`** ——
// 所以離開 STONEGATE 再回來,權杖躺在原地第二次。與魔毯同一種原版行為,照抄。
const (
	// CrownNPCLocation / CrownNPCSlot 是王冠那一槽。
	CrownNPCLocation = 18
	CrownNPCSlot     = 1

	// SceptreNPCLocation / SceptreNPCSlot 是權杖那一槽。
	SceptreNPCLocation = 29
	SceptreNPCSlot     = 9
)

// RegaliaNPCPlacement 是王冠與權杖在 `.NPC` 檔裡的位置,供測試對真檔核對。
var RegaliaNPCPlacement = []struct {
	Name     string
	Kind     byte
	Location int
	Slot     int
	X, Y     int
	Floor    int
}{
	{Name: "王冠", Kind: ItemCrown, Location: CrownNPCLocation, Slot: CrownNPCSlot,
		X: 15, Y: 13, Floor: 3},
	{Name: "權杖", Kind: ItemSceptre, Location: SceptreNPCLocation, Slot: SceptreNPCSlot,
		X: 15, Y: 15, Floor: 0},
}
