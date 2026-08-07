package u5data

// 地牢第一人稱透視:幾何與選圖規則
//
// 這一套是把兩件事對起來得到的:
//
//  1. **選哪一張圖** —— 直接讀 FM Towns 的繪圖流程
//     (`sub_3D14` 主迴圈、`sub_3878` 側牆、`sub_36C0` 正面)。
//     那是遊戲邏輯,兩版共用;FM Towns 換掉的只是美術(`.PNL` vs `DNG*.16`)。
//  2. **畫在哪裡** —— FM Towns 的座標表是它自己的 640×480 版面,套不到 DOS。
//     但 DOS 的切片寬度自己就把答案講完了:
//
//     側牆四階  24 + 32 + 16 + 8 = **80**
//     正面四階  80 / 56 / 24 / 8
//
//     80 就是半個畫面。而正面第 d 階的寬度**正好等於 80 減掉前面每一階側牆的寬度**:
//
//     80 − 0 = 80 ✓   80 − 24 = 56 ✓   80 − 24 − 32 = 24 ✓   80 − 24 − 32 − 16 = 8 ✓
//
//     四條同時成立,而且兩組數字來自檔案裡兩批不同的形狀 ——
//     這不是湊出來的,是幾何本身。組出來的畫面也確實是一條對稱的透視走廊。
//
// ⚠ 每張切片都是**半邊**:同一張圖畫左邊,再水平鏡射畫右邊。
// 所以整個視野是 160 × 164。原版把它擺在 176×176 的地圖窗裡,
// 精確的內縮量要讀 DOS `DUNGEON.OVL` 才知道 —— 引擎目前置中。

const (
	// DungeonViewWidth 是透視畫面的總寬(兩個半邊)。
	DungeonViewWidth = 160
	// DungeonViewHalfWidth 是半邊的寬度。
	DungeonViewHalfWidth = DungeonViewWidth / 2
	// DungeonViewDepths 是往前看幾格(含腳下這一格)。
	//
	// `sub_3D14` 的 `for (depth = 0; depth < 4; depth++)` —— 深度 0 是**腳下這一格**,
	// 不是前面那一格。所以側牆第 0 階畫的是自己左右兩側的牆。
	DungeonViewDepths = 4
)

// dungeonBandX 是第 d 階在半邊裡的起點 x。
//
// 側牆與正面共用同一組起點 —— 那正是上面那四條算式的意思。
var dungeonBandX = [DungeonViewDepths]int{0, 24, 56, 72}

// DungeonBandX 回傳第 d 階在左半邊的起點 x。
func DungeonBandX(depth int) int {
	if depth < 0 || depth >= DungeonViewDepths {
		return 0
	}
	return dungeonBandX[depth]
}

// 形狀群的基底編號(`DNG*.16` 的 28 格)
//
// 側牆(`sub_3878`,依**側邊那一格**的種類):
//
//	tile < 0xA0                → 16 + d   看得穿的開口
//	0xA0 / 0xE0 / 0xF0         →  4 + d   側牆上有門
//	0xC0                       → 20 + d
//	其餘(0xB0 / 0xD0)         →  0 + d   實心側牆
//
// 正面(`sub_36C0`,依**擋住視線的那一格**的種類,查 `byte_4FEEE`):
//
//	0xB0 / 0xD0                →  8 + d
//	0xA0 / 0xE0 / 0xF0         → 12 + d
//	0xC0                       → 24 + d
//	**但深度 0 一律用 12**(`if (depth == 0) shape = 0x0C`)
//
// ⚠ 第 8 格與第 24 格是空的 —— 因為深度 0 永遠走 12,那兩個基底的第 0 階
// 從來不會被要求。原版留白不是漏做。
const (
	dungeonSideSolid = 0
	dungeonSideDoor  = 4
	dungeonSideOpen  = 16
	dungeonSideC0    = 20

	dungeonFrontWall = 8
	dungeonFrontDoor = 12
	dungeonFrontC0   = 24
)

// DungeonSideShape 回傳側牆該用第幾個形狀。
func DungeonSideShape(tile byte, depth int) int {
	if tile < DungeonRoomA {
		return dungeonSideOpen + depth
	}
	switch DungeonKind(tile) {
	case DungeonRoomA, DungeonDoorway, DungeonRoomF:
		return dungeonSideDoor + depth
	case DungeonUnknownC:
		return dungeonSideC0 + depth
	}
	return dungeonSideSolid + depth
}

// DungeonFrontShape 回傳擋住視線的正面該用第幾個形狀;不擋則回 −1。
func DungeonFrontShape(tile byte, depth int) int {
	if tile < DungeonRoomA {
		return -1
	}
	// 深度 0 一律 12 —— 那是唯一有第 0 階的正面群。
	if depth == 0 {
		return dungeonFrontDoor
	}
	switch DungeonKind(tile) {
	case DungeonRoomA, DungeonDoorway, DungeonRoomF:
		return dungeonFrontDoor + depth
	case DungeonUnknownC:
		return dungeonFrontC0 + depth
	}
	return dungeonFrontWall + depth
}

// DungeonSeeThrough 回報視線能不能穿過這一格繼續往前(`sub_36C0` 的回傳值)。
//
// 通道類(< 0xA0)一律看得穿;此外只有一個例外 ——
// **站在門口(0xE0)時,深度 0 看得穿**,不然人在門框裡會什麼都看不到。
func DungeonSeeThrough(tile byte, depth int) bool {
	if tile < DungeonRoomA {
		return true
	}
	return depth == 0 && DungeonKind(tile) == DungeonDoorway
}

// DungeonTheme 回傳這座地牢用哪一套外觀圖組(1 = DNG1、2 = DNG2、3 = DNG3)。
//
// `sub_5378` 依**地點編號**分三組(`n = 地點 − 0x20`):
//
//	n == 1, 4, 5  → 3  磚牆   欺瞞 / 謬誤 / 貪婪
//	n == 6, 7     → 2  熔岩   羞恥 / 海斯洛斯
//	其餘          → 1  洞穴   輕蔑 / 毀滅 / 末日
//
// ⚠ 不是「第 n 座用第 n 套」。引擎第一版拿地牢索引對 3 取餘數 ——
// 那讓欺瞞變成洞穴,而原版是磚牆。分組是寫死的,照抄。
//
// 同一段程式碼還設了 `byte_418DC/418DD`(磚牆組是 'O','E',其餘是 'M',5),
// 用途未追 —— 疑為配樂或音效編號。
func DungeonTheme(location int) int {
	switch location - 0x20 {
	case 1, 4, 5:
		return 3
	case 6, 7:
		return 2
	}
	return 1
}
