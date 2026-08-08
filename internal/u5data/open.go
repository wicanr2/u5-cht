package u5data

// Open(O)指令的地形與物件規則(原版 `sub_15374` / `sub_15108` / `sub_152B8`)
//
// 一個鍵三條路:
//
//	地點 0x21..0x28(地牢)→ sub_152B8   開**腳下**那一格的寶箱,不問方向
//	其餘                   → 問方向,對目標格查下面這張表
//	表裡沒有的 tile         → sub_15108   看那一格的**物件層**有沒有箱子
//
// ⚠ 地牢那條**不問方向**,也不看面前那一格 —— 它讀的是
// `byte_3E0A6/A7` 也就是玩家自己站的位置。引擎原本用的是
// `dungeonFacingTile`(腳下或面前),而那條規則的出處是 `sub_18D18` ——
// **施法**選目標用的,不是 Open。見 `docs/re/75`。

// Open 指令對某個 tile 的處置。
type OpenAction int

// 五種處置(原版 `sub_15374` 的跳表 + 三個前置 `cmp`)。
const (
	// OpenObjectLayer 是「這個 tile 不特別,去看物件層」(原版走 `sub_15108`)。
	OpenObjectLayer OpenAction = iota
	// OpenAlreadyOpen 印 "It's open!"。
	OpenAlreadyOpen
	// OpenTooHeavy 印 "Too heavy!" —— 只有柵門(0x99)。
	OpenTooHeavy
	// OpenLocked 印 "Locked!"。
	OpenLocked
	// OpenDoor 打開:那一格變成 0x44 磚地,並排定 4 回合後自動關上。
	OpenDoor
)

// OpenedDoorTile 是門打開之後那一格變成什麼(原版 `mov byte ptr [eax], 44h`)。
//
// ★ 與 An Ylem(消除)寫的是**同一個值** —— 0x44 是「什麼都沒有的地板」。
const OpenedDoorTile = 0x44

// DoorAutoCloseTurns 是打開的門幾回合後自己關上(原版 `byte_3E164 = 4`)。
//
// ★★ 這條機制引擎完全沒有。主迴圈 `sub_1A54` 每回合做:
//
//	if (byte_3E161 != 0 && --byte_3E164 == 0)
//	    sub_2B64C(byte_3E161, byte_3E162, byte_3E163)   // 把 tile 寫回去
//
// 而 `sub_2B64C` 只在**場景裡**(1 <= 地點 < 0x21)才真的寫回 ——
// 大地圖與地牢沒有門要關。
//
// ⚠ 而且那四個變數**只有一組**:`sub_15374` 開門前先呼叫一次
// `sub_2B64C(byte_3E161, …)` 把上一扇關掉 ⇒ **同時只能有一扇門是開的**。
// 這不是最佳化,是可觀察的行為:走過一長串門會看到後面的自己關上。
const DoorAutoCloseTurns = 4

// OpenActionFor 回報 Open 指令碰到這個 tile 該做什麼。
func OpenActionFor(tile byte) OpenAction {
	switch tile {
	case TileMagicLockedA, TileMagicLockedB: // 0x97 / 0x98 魔法鎖
		return OpenLocked
	case TilePortcullis: // 0x99 柵門
		return OpenTooHeavy
	case TileItsOpen: // 0xAF
		return OpenAlreadyOpen
	case TileDoorA, TileDoorB: // 0xB8 / 0xBA
		return OpenDoor
	case TileLockedDoor, TileLockedMagicDoor: // 0xB9 / 0xBB
		return OpenLocked
	}
	return OpenObjectLayer
}

// TilePortcullis 是柵門(`look#153`)—— Open 對它印 "Too heavy!"。
const TilePortcullis = 0x99

// TileItsOpen 是 Open 印 "It's open!" 的那一格(原版跳表 case 175)。
//
// ⚠ `look#175` 給它的名字是「沉重的行李箱」,而「It's open!」對行李箱說不太通。
// 名字來自 `LOOK2.DAT`,行為來自 `sub_15374` 的跳表 —— **照值實作,不編故事**。
// 名字與行為為何對不上,記在 `docs/re/75` 的 ⬜。
const TileItsOpen = 0xAF

// 物件層的兩個特例(原版 `sub_15108` 的兩個 `cmp al`)。
const (
	// ObjLockedChest 是可以開的箱子(物件種類 1)。
	ObjLockedChest = 1
	// ObjSandalwoodBox 是檀香木盒 —— Open 對它印 "Can't!"(打不開)。
	ObjSandalwoodBox = 0x0E
)

// ChestTrapQualityBit 是 `ChestTrappedWorld` 的別名 —— 品質這**一個位元組**
// 同時裝兩件事:最高位 = 有陷阱、低七位 = 獎品等級。
// 原版開箱時 `and var_5, 7Fh` 把陷阱位清掉之後才拿去擲獎品。
const ChestTrapQualityBit = ChestTrappedWorld

// LastSceneLocation 是最後一個「場景」地點編號(原版一族 `cmp al, 20h` 的上界)。
//
// 1..0x20 是城鎮 / 城堡 / 民居 / 要塞;0 是大地圖、0x21..0x28 是地牢、
// > 0x7F 是戰鬥。開箱扣業報與門會自己關上兩條都只在這個範圍成立。
const LastSceneLocation = 0x20

// DungeonChestTrapMask 是地牢寶箱的陷阱判準 —— ★ **低三位元全部**。
//
// ⚠ `sub_152B8`(Open)用 `test di, 7`;而 `sub_18D18`(An Sanct 解陷阱)
// 用的是單一位元 `ChestTrappedDungeon`。**同一件事兩支各做一半**,
// 判準不同(`docs/re/74` §1 的形狀)。照各自原樣實作,不統一。
const DungeonChestTrapMask = 0x07

// DungeonOpenedChestKind 是「開過的寶箱」的高四位元(`DungeonOpenedChest` 寫的 0x70)。
const DungeonOpenedChestKind = 0x70

// ChestOpenKarmaPenalty 是在場景裡開箱子的業報代價(原版 `sub byte_3E098, 2`,下限 0)。
//
// ★ 只在 **1 <= 地點 <= 0x20**(城鎮 / 城堡 / 民居 / 要塞)扣 ——
// 大地圖上的箱子是無主的,地牢裡的也不扣。那是「翻別人家的箱子」的代價。
const ChestOpenKarmaPenalty = 2
