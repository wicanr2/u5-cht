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
// ⚠⚠ **2026-08-08 更正**:第一版有四個模式寫錯了(詳見 `docs/re/51`)。
// 錯因是照「名字聽起來該是什麼」寫,而沒有逐 case 讀跳表 ——
// 名字是我取的,判定式必須來自組語。
const (
	// MoveWalk 是一般陸行:只看阻擋位圖(`sub_2A610`)。
	MoveWalk MoveMode = 0
	// MoveWaterOnly 只能待在水上 —— 水生怪物(`sub_2A674`)。
	MoveWaterOnly MoveMode = 1
	// MoveAmphibious 水陸兩棲:水面可以、陸上照阻擋位圖。魔毯走這條。
	MoveAmphibious MoveMode = 2
	// MoveHorse 是坐騎:照阻擋位圖,**外加擋住 tile 4(沼澤)與 0x8F**。
	//
	// ⚠ 更正:原版不是「不能下水 + 位圖」,水本來就在位圖裡;
	// 它多擋的是**沼澤** —— 馬過不了沼澤,而這一條會直接影響玩家怎麼繞路。
	MoveHorse MoveMode = 3
	// MoveFlyer 是飛行:**不看位圖,但過不了水**(`!sub_2A674`)。
	//
	// ⚠ 更正:第一版寫成「什麼都不管」(`return true`)。原版飛得過山與牆,
	// 但飛不過水面 —— 少了後半,飛行怪物會直接飛過海洋追殺玩家。
	MoveFlyer MoveMode = 4
	// MoveSkiff 是小艇:水面 + **河道有流向**(見 `skiffDirMask`)。
	MoveSkiff MoveMode = 5
	// MoveShip 是大船:**只走 tile 0..2**。
	//
	// ⚠ 更正:第一版還放行了 0x6x 那一族水(河流)。原版是純 `tile <= 2` ——
	// 大船進不了河道,只走深海與中海。這正是「大船靠不了岸」的真正機制。
	MoveShip MoveMode = 6
	// MoveSwampOnly / GrassOnly / ShallowOnly / DesertOnly 是**只走一種地形**
	// 的四個模式(原版 case 7..10 各是一行 `cmp esi, N; setz al`)。
	//
	// ⚠ 更正:第一版把這四種當「兩棲」並標 TODO,實際上它們是**最嚴格**的 ——
	// 只認一種 tile。對應的移動者是 0xF8 / 0xF4 / 0xEC / 0xE0 那幾組,
	// 也就是長在特定地形上不動窩的東西(沼澤的觸手怪、荒漠的沙陷)。
	MoveSwampOnly   MoveMode = 7  // tile == 4
	MoveGrassOnly   MoveMode = 8  // tile == 5
	MoveShallowOnly MoveMode = 9  // tile == 1
	MoveDesertOnly  MoveMode = 10 // tile == 7

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

// 各模式用到的地形常數,全部來自 `jpt_2A6B4` 的各 case。
const (
	// ShipDeepMax 是大船走得的最深 tile(原版 case 6:`cmp esi, 2; setle`)。
	ShipDeepMax = 2
	// horseBlockSwamp / horseBlockOdd 是坐騎額外擋住的兩格(case 3)。
	horseBlockSwamp = 4
	horseBlockOdd   = 0x8F
	// 四個「只走一種地形」的模式各自認的 tile。
	swampTile   = 4
	grassTile   = 5
	shallowTile = 1
	desertTile  = 7
	// skiffPierLo / Hi 是小艇的另一組來源 tile(case 5:`tile & 0xFC == 0x34`)。
	skiffPierLo = 0x34
	skiffPierHi = 0x37
	// skiffFreeBelow 是「不查流向就通」的界線(case 5:`cmp esi, 60h; jl`)。
	skiffFreeBelow = 0x60
)

// skiffDirMask 是小艇的**流向遮罩**:每一格四個位元,對應能從哪個朝向進去。
//
// ★ 這解掉了 U5 河道的機制:河流是**有方向的**,小艇只能順著走。
// 原版查的是 `記憶體[0x5FF6C + tile]`(tile ≥ 0x60)與
// `記憶體[0x5FFA8 + tile]`(tile 0x34..0x37)—— 兩個算式**指到同一塊記憶體**,
// 0x5FFA8 + 0x34 == 0x5FF6C + 0x70。所以那是一張表用兩種偏移讀,
// 不是兩張表;0x34..0x37 借用了 0x70..0x73 的形狀。
//
// 這裡存的是 0x5FFCC..0x5FFEB 那 32 個位元組(= tile 0x60..0x7F)。
var skiffDirMask = [32]byte{
	0x0A, 0x05, 0x03, 0x09, 0x0C, 0x06, 0x0B, 0x0D,
	0x0E, 0x07, 0x0A, 0x05, 0x08, 0x04, 0x02, 0x01,
	0x06, 0x03, 0x09, 0x0C, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x01, 0x01, 0x02, 0x02, 0x03,
}

// skiffAllows 回報小艇能不能以這個朝向進這一格(原版 case 5)。
//
// 位元的取法是 `8 >> (移動者 & 3)` —— 朝向就是載具碼的低兩位。
func skiffAllows(mover byte, tile int) bool {
	bit := byte(8) >> (mover & 3)
	if tile >= skiffPierLo && tile <= skiffPierHi {
		// 0x34..0x37 借 0x70..0x73 的那四格(0x5FFA8 + tile 的實際落點)。
		return skiffDirMask[tile-skiffPierLo+0x10]&bit != 0
	}
	if !TileIsWater(tile) {
		return false
	}
	if tile < skiffFreeBelow {
		return true // 深水與中水不查流向
	}
	if tile >= skiffFreeBelow+len(skiffDirMask) {
		return false
	}
	return skiffDirMask[tile-skiffFreeBelow]&bit != 0
}

// MoverBlocked 是原版 `sub_2A610(移動者, tile)` 的完整判定 —— 它**不只是位圖**:
//
//	位圖那一位設起來            → 擋
//	移動者 & 0xFE == 0x1C(步行) → 過
//	移動者 & 0xF0 == 0x40        → 過
//	tile & 0xFC == 0x90          → 擋(只有上面兩種移動者過得去)
//
// ⚠ 第三條是本專案早就知道的 `TileNeedsSpecialMover`,但先前沒有接進通行判定 ——
// 只用了位圖。少了它,一般怪物會走進那四格「只有特定移動者能過」的地形。
func MoverBlocked(mover byte, tile int) bool {
	if TileBlocksWalking(tile) {
		return true
	}
	if mover&0xFE == VehicleWalk || mover&0xF0 == moverFreePass {
		return false
	}
	return TileNeedsSpecialMover(tile)
}

// moverFreePass 是「不受 0x90 那組限制」的移動者高位元(`cmp eax, 40h`)。
const moverFreePass = 0x40

// MoveModeAllows 回報某種模式能不能進這一格。
//
// ⚠ 這一支**只管地形,而且不知道移動者是誰** —— 模式 0 與 3 在原版是
// `sub_2A610(移動者, tile)`,會多看一條「0x90 那組地形」的規則。
// 真正的入口是 `MoverCanEnter`;這一支留給只有模式在手的呼叫端,
// 對模式 0 / 3 會**偏寬鬆**。
//
// ⚠ 小艇(模式 5)還要看**朝向**,所以它也在 `MoverCanEnter` 才完整。
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
		// 位圖 + 沼澤與 0x8F 兩格另外擋掉。
		return !TileBlocksWalking(tile) && tile != horseBlockSwamp && tile != horseBlockOdd
	case MoveFlyer:
		// 飛得過山與牆,飛不過水。
		return !water
	case MoveSkiff:
		// 朝向不明時放寬成「是水就行」;真正的判定在 MoverCanEnter。
		return water || (tile >= skiffPierLo && tile <= skiffPierHi)
	case MoveShip:
		return tile >= 0 && tile <= ShipDeepMax
	case MoveSwampOnly:
		return tile == swampTile
	case MoveGrassOnly:
		return tile == grassTile
	case MoveShallowOnly:
		return tile == shallowTile
	case MoveDesertOnly:
		return tile == desertTile
	}
	// 表上的 0xFF 落在 switch 之外,原版走 default —— **回 0(不能進)**,
	// 不是「與一般陸行同」。⚠ 第一版寫成陸行,那讓 0xE8..0xEB 那組移動者
	// 憑空獲得了通行能力。
	return false
}

// MoverCanEnter 是給呼叫端的入口:某個移動者進不進得了某一格。
//
// 小艇要看朝向(河道有流向),所以單獨走 `skiffAllows`。
func MoverCanEnter(mover byte, tile int) bool {
	mode := ModeOf(mover)
	switch mode {
	case MoveSkiff:
		return skiffAllows(mover, tile)
	case MoveWalk:
		return !MoverBlocked(mover, tile)
	case MoveAmphibious:
		return TileIsWater(tile) || !MoverBlocked(mover, tile)
	case MoveHorse:
		return !MoverBlocked(mover, tile) && tile != horseBlockSwamp && tile != horseBlockOdd
	}
	return MoveModeAllows(mode, tile)
}
