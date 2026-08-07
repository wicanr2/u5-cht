package u5data

// 移動者 → 移動模式(原版 `sub_2A694` 的 `byte_5FF8C[mover >> 2]`)
//
// `docs/re/02` 早就記下了分派器的形狀,但**表的內容當時標「待確認」**。
// 這一份把 64 項全部抽出來,並把 11 種模式的判定各自寫清楚。
//
// # 為什麼是 `mover >> 2`
//
// 載具碼是「四個朝向一組」(大船 0x24..0x27、小艇 0x28..0x2B、馬 0x10..0x13…),
// 所以除以 4 就把一組收成一格。64 項覆蓋 0x00..0xFF 的全部移動者。
//
// # 兩個基礎判定
//
//	tileIsWater(tile)      tile < 4 或 (tile & 0xF0) == 0x60   (原版 `sub_2A674`)
//	TileBlocksWalking(tile) 那張 32 B 的阻擋位圖               (原版 `sub_2A610`)
//
// 11 種模式就是這兩者的不同組合。

// MoveMode 是一種移動方式。
type MoveMode int

// 11 種模式(原版跳表 `jpt_2A6B4` 的 case 0..10)。
//
// 名字是**依判定內容命名**的,不是照原版(原版沒有名字)。每一條的依據
// 寫在 `MoveModeAllows` 的註解裡。
const (
	// MoveWalk 是一般陸行:只看阻擋位圖。步行、多數 NPC 都走這條。
	MoveWalk MoveMode = 0
	// MoveWaterOnly 只能待在水上 —— 水生怪物。
	MoveWaterOnly MoveMode = 1
	// MoveAmphibious 水陸兩棲:水面可以、陸上照阻擋位圖。魔毯也走這條。
	MoveAmphibious MoveMode = 2
	// MoveHorse 是坐騎:陸行,但水一律不行。
	MoveHorse MoveMode = 3
	// MoveFlyer 是飛行:阻擋位圖與水都不管。
	MoveFlyer MoveMode = 4
	// MoveSkiff 是小艇:只走水,而且淺水也行。
	MoveSkiff MoveMode = 5
	// MoveShip 是大船:只走**深水**(淺灘會擱淺)。
	MoveShip MoveMode = 6
	// MoveMode7..10 是四種還沒逐條確認語意的模式,先照判定式實作。
	MoveMode7  MoveMode = 7
	MoveMode8  MoveMode = 8
	MoveMode9  MoveMode = 9
	MoveMode10 MoveMode = 10

	// MoveModeNone 是表上的 0xFF —— 落在 switch 範圍外,走 default。
	MoveModeNone MoveMode = 0xFF
)

// moverMode 是 `byte_5FF8C` 的 64 項,索引 = 移動者碼 >> 2。
//
// 抽自 `WORRIORS.EXP` 線性位址 0x5FF8C(檔案位移 = 線性 + 0x200)。
// 對得上的錨點(**這些常數都是別處獨立推出來的**):
//
//	索引  4(mover 0x10..0x13)= 3  ⇄ TileHorse 0x10 / 馬的載具碼 0x12,0x13
//	索引  5(mover 0x14..0x17)= 2  ⇄ VehicleCarpet 0x14  —— 魔毯兩棲,對得上「能飛過水」
//	索引  7(mover 0x1C..0x1F)= 0  ⇄ VehicleWalk 0x1C     —— 步行是一般陸行
//	索引  8(mover 0x20..0x23)= 6  ⇄ VehicleSailing 0x20  —— 揚帆中
//	索引  9(mover 0x24..0x27)= 6  ⇄ VehicleShip 0x24     —— 大船
//	索引 10(mover 0x28..0x2B)= 5  ⇄ VehicleSkiff 0x28    —— 小艇
//
// 六個錨點同時對上,表的對齊沒有滑動的餘地(`rulebook/62`)。
var moverMode = [64]byte{
	0, 0, 0, 0, 3, 2, 0, 0, // 0x00..0x1F
	6, 6, 5, 6, 0, 0, 0, 0, // 0x20..0x3F
	0, 0, 0, 0, 0, 0, 0, 0, // 0x40..0x5F
	0, 0, 0, 0, 0, 2, 0, 2, // 0x60..0x7F
	1, 1, 1, 1, 0, 2, 0, 4, // 0x80..0x9F
	0, 0, 0, 0, 2, 0, 0, 0, // 0xA0..0xBF
	0, 0, 0, 0, 0, 0, 2, 2, // 0xC0..0xDF
	10, 0, 0xFF, 9, 2, 8, 7, 4, // 0xE0..0xFF
}

// ModeOf 回傳這個移動者用哪一種模式。
func ModeOf(mover byte) MoveMode { return MoveMode(moverMode[mover>>2]) }

// 水的判定用既有的 `TileIsWater`(`tileflags.go`)—— 那一支就是 `sub_2A674`。
// ⚠ 我一度在這裡重寫了一份,`vet` 才擋下來。同 `docs/re/45` 的教訓:
// **加東西前先 grep 名字。**

// waterGroup 是第二段水的高位(`tile & 0xF0 == 0x60`)。
const waterGroup = 0x60

// DeepWaterMax 是大船走得的水:只有 tile < 3(深水與中水),淺灘 3 會擱淺。
const DeepWaterMax = 3

// MoveModeAllows 回報某種模式能不能進這一格。
//
// ⚠ 這一支**只管地形**。物件、NPC、力場那些擋路的東西由呼叫端另外判 ——
// 原版也是分開的(`sub_2A694` 只看 tile)。
func MoveModeAllows(mode MoveMode, tile int) bool {
	water := TileIsWater(tile)
	switch mode {
	case MoveWalk:
		return !TileBlocksWalking(tile)
	case MoveWaterOnly:
		return water
	case MoveAmphibious:
		// 水面直接可以;陸上照阻擋位圖。
		return water || !TileBlocksWalking(tile)
	case MoveHorse:
		return !water && !TileBlocksWalking(tile)
	case MoveFlyer:
		return true
	case MoveSkiff:
		return water
	case MoveShip:
		// 大船吃水深:淺灘(3)過不去。
		return tile >= 0 && (tile < DeepWaterMax || tile&0xF0 == waterGroup)
	case MoveMode7, MoveMode8, MoveMode9, MoveMode10:
		// ⚠ 這四種的判定式還沒逐條核完(`docs/re/47` §4)。
		// 先當成兩棲 —— 它們對應的都是 0xE0 之後的特殊移動者
		// (暗影君主、旋風之類),而那些東西在原版裡幾乎哪裡都去得。
		// **標成 TODO 而不是假裝確定**。
		return water || !TileBlocksWalking(tile)
	}
	// 表上的 0xFF 落在 switch 之外,原版走 default —— 與一般陸行同。
	return !TileBlocksWalking(tile)
}

// MoverCanEnter 是給呼叫端的入口:某個移動者進不進得了某一格。
func MoverCanEnter(mover byte, tile int) bool {
	return MoveModeAllows(ModeOf(mover), tile)
}
