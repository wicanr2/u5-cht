package u5data

// 地牢戰場(原版 `sub_FE48` → `sub_FD54` → `sub_FC7C` → `sub_FAC4`/`sub_FB6C`/`sub_FBF0`)
//
// ★ **地牢的遭遇戰不用 `DUNGEON.CBT`。** 那 112 張全是房間(7 座 × 16 間),
// 遊蕩怪物的戰場是**當場畫出來的** —— 程式把 352 B 的戰鬥地圖緩衝
// (`byte_3F8F4`,先被 `rep stosd` 填成 0xFF)當畫布,照六張小表描出來。
//
// 一開始我找不到它讀哪一張 `.CBT`,差點下「地牢戰場資料還沒破」的結論;
// 真相是**根本沒有那份資料**。找不到的東西有兩種:還沒找到的,與不存在的。
//
// # 畫出來的樣子
//
//	FFwwwwwwwwwFF        w = 牆(byte_418DC)
//	wwwwwwwwwwww         列 1 與列 9 整條、行 1 與行 9 整條
//	 w         w         四角是 0xFF(空白)
//	 w    C    w         C = 腳下地形對應的擺設(梯子 / 寶箱 / 噴泉)
//	 w         w         內部 7×7 留白(0xFF)
//	wwwwwwwwwwww
//	FFwwwwwwwwwFF
//
// 內部與外環**沒有被寫過** —— 緩衝的初值 0xFF 就是它們的內容。0xFF 在本專案
// 已經是 `TileBlank`(「此處無物」,見 `sceneset.go`),兩邊獨立對上。
//
// # 四面牆會照你站的地方開口
//
// `sub_FC7C(方位)` 去看**地牢裡**那個方位的鄰格,再決定那面牆怎麼處理:
//
//	鄰格高四位元 < 0xA0        → 開七格(`sub_FBF0`):走得過去的通道
//	鄰格是 0xB0 / 0xC0 / 0xD0  → 整條邊封成 0xFF(`sub_FAC4`):實心牆
//	其餘(房間 0xA0/0xF0、門 0xE0)→ 開五格(`sub_FB6C`):窄口
//
// ⇒ **戰場的形狀就是你剛才站的那個路口。** 死巷裡打就只有一個出口。
// 這是「機制與素材忠於原版」在細節上的樣子:少了這一步,戰術完全不同。

// 地牢戰場的幾何常數(全部來自上述四支的內嵌值)。
const (
	// dungeonArenaBox 是內牆框的兩條線所在的行 / 列。
	dungeonArenaBoxLo = 1
	dungeonArenaBoxHi = 9
	// dungeonArenaWideGap 是通道口的寬度(`sub_FBF0` 的 `cmp edi, 7`)。
	dungeonArenaWideGap = 7
	// dungeonArenaNarrowGap 是窄口的寬度(`sub_FB6C` 的 `cmp edi, 5`)。
	dungeonArenaNarrowGap = 5
	// dungeonArenaCentre 是擺設放的位置(`byte_3F999` = 5×32 + 5)。
	dungeonArenaCentre = 5
)

// dungeonArenaWall / Floor 是牆與地板的 tile。
//
// **依地牢不同**(原版 `sub_5378` 的 `loc_53FD` / `loc_5414`):
// 地牢編號(地點 − 0x20)為 1 / 4 / 5 的三座用一組,其餘七座用另一組,
// 而且前者另外把圖組換成 `byte_3EE16 = 3`。
//
// 對上 `DungeonEntrances`(同樣抽自執行檔的地點表後 8 筆),編號 1 / 4 / 5
// 就是索引 0 / 3 / 4 —— **欺瞞、謬誤、貪婪**。這三座在原版裡另配一組圖組,
// 所以牆與地板也換一套。
//
// ⚠ 名字是查一手表得到的,不是憑「U5 地牢的習慣順序」推的。
const (
	dungeonArenaWallA  = 0x4F // 編號 1 / 4 / 5
	dungeonArenaFloorA = 0x45
	dungeonArenaWallB  = 0x4D // 其餘
	dungeonArenaFloorB = 0x05
)

// 戰場模式(原版 `byte_3E0B1`,由 `sub_2E364` 的第一個參數設定)。
//
// 帶 bit 2(4 或 6)的走「營地」那種佈局;bit 7(0x80)會啟用
// 「全隊必須從同一出口」那條規則,而這四個值都不帶它。
const (
	DungeonArenaModeEncounter = 0 // 一般遭遇(`sub_2E364(0, …)`)
	DungeonArenaModeWander    = 2 // 地牢遊蕩怪物(`sub_5008`)
	DungeonArenaModeCampField = 4 // 地表紮營(`sub_2E8B0`)
	DungeonArenaModeCamp      = 6 // 地牢紮營(`sub_2B8CC`)
)

// DungeonArenaAltSet 是用另一組牆地板的三個地牢編號(地點 − 0x20)。
var DungeonArenaAltSet = [3]int{1, 4, 5}

// DungeonArenaTiles 回傳這座地牢的牆與地板 tile。number 是地點 − 0x20(1..8)。
func DungeonArenaTiles(number int) (wall, floor byte) {
	for _, n := range DungeonArenaAltSet {
		if number == n {
			return dungeonArenaWallA, dungeonArenaFloorA
		}
	}
	return dungeonArenaWallB, dungeonArenaFloorB
}

// dungeonArenaCentreTile 是 `byte_418E0[8]`:腳下地形的高四位元 >> 4 → 擺在
// 戰場正中央的 tile。0 代表什麼都不擺。
//
//	1 上行梯 → 0xC8    2 下行梯 → 0xC9    3 兩向梯 → 0xC8(畫成上行)
//	4 寶箱   → 0xDC    5 噴泉   → 0xD8    6 陷阱 / 7 門 → 無
//
// ★ 0xC8 / 0xC9 正是本專案早已知道的上下梯 tile(`DungeonCanClimbUp` 那一組),
// 而 0xDC / 0xD8 對上寶箱與噴泉 —— 四個值都能獨立對照,位移沒有偏。
var dungeonArenaCentreTile = [8]byte{0x00, 0xC8, 0xC9, 0xC8, 0xDC, 0xD8, 0x00, 0x00}

// 隊伍入場位置(原版 `byte_418F8` / `byte_418FE` / `byte_41904` / `byte_4190A`,各 6 B)。
//
// ★ 四張表被**同時當 X 也當 Y**,只是配對方式不同 —— 因為四個方位的隊形
// 就是同一個楔形轉 90 度。原版沒有存四份座標,存了兩份形狀。
var (
	dungeonArenaSpread = [CombatPartySlots]byte{5, 4, 6, 3, 5, 7} // byte_418F8 == byte_4190A
	dungeonArenaFar    = [CombatPartySlots]byte{6, 7, 7, 8, 8, 8} // byte_418FE
	dungeonArenaNear   = [CombatPartySlots]byte{4, 3, 3, 2, 2, 2} // byte_41904
)

// 敵人入場位置(原版 `byte_41910` / `41920` / `41930` / `41940`,各 16 B)。
//
// 41920 = 10 − 41930 逐項成立、41910 與 41940 只差第 10、11 項 ——
// 同樣是旋轉對稱,不是四份獨立資料。
var (
	dungeonArenaEnemyA = [CombatEnemySlots]byte{5, 4, 6, 3, 7, 2, 8, 5, 2, 8, 3, 7, 2, 4, 6, 8}  // 41910
	dungeonArenaEnemyB = [CombatEnemySlots]byte{8, 8, 8, 7, 7, 6, 6, 9, 8, 8, 9, 9, 10, 10, 10, 10} // 41920
	dungeonArenaEnemyC = [CombatEnemySlots]byte{2, 2, 2, 3, 3, 4, 4, 1, 2, 2, 1, 1, 0, 0, 0, 0}  // 41930
	dungeonArenaEnemyD = [CombatEnemySlots]byte{5, 4, 6, 3, 7, 2, 8, 5, 2, 8, 7, 3, 2, 4, 6, 8}  // 41940
)

// DungeonArenaSide 把朝向換成隊伍從哪一邊入場(原版 `sub_10334` 的六路跳表)。
//
//	朝北(0)→ 3(從南邊進場,敵人在北)
//	朝東(1)→ 2
//	朝南(2)→ 4
//	朝西(3)→ 1
//
// 跳表另有 4 → default(4)、5 → 3 兩路,但**到不了** ——
// 寫朝向的 `sub_100F8` 只把 4 / 5(上下樓)寫進 `byte_3EE14`,
// `byte_3EE15` 永遠是 0..3。照抄跳表,並記著那兩路是死碼。
func DungeonArenaSide(facing int) int {
	switch facing {
	case 0, 5:
		return 3
	case 1:
		return 2
	case 3:
		return 1
	}
	return 4
}

// DungeonArenaParty 回傳某一邊入場時六名隊員的座標。
func DungeonArenaParty(side int) (x, y [CombatPartySlots]byte) {
	switch side {
	case 1:
		return dungeonArenaFar, dungeonArenaSpread
	case 2:
		return dungeonArenaNear, dungeonArenaSpread
	case 3:
		return dungeonArenaSpread, dungeonArenaFar
	}
	return dungeonArenaSpread, dungeonArenaNear // 4
}

// DungeonArenaEnemies 回傳這個朝向下十六個敵人入場點的座標。
func DungeonArenaEnemies(facing int) (x, y [CombatEnemySlots]byte) {
	switch facing {
	case 0:
		return dungeonArenaEnemyA, dungeonArenaEnemyC
	case 1:
		return dungeonArenaEnemyB, dungeonArenaEnemyD
	case 2:
		return dungeonArenaEnemyD, dungeonArenaEnemyB
	}
	return dungeonArenaEnemyC, dungeonArenaEnemyA // 3
}

// DungeonArena 是畫一張地牢戰場需要的輸入。
type DungeonArena struct {
	// Number 是地牢編號(地點 − 0x20,1..8)—— 決定牆與地板的 tile。
	Number int
	// Here 是隊伍腳下那一格的地牢 tile。
	Here byte
	// Around 是四個方位的鄰格(索引 = 0 北 / 1 東 / 2 南 / 3 西)。
	Around [4]byte
	// Facing 是隊伍的朝向(0..3)。
	Facing int
}

// BuildDungeonArena 照原版把一張地牢戰場畫出來。
//
// 回傳的 `CombatMap` 只缺 `EnemyKind` —— 那要由呼叫端依遭遇到的怪物與隻數填。
func BuildDungeonArena(a DungeonArena) *CombatMap {
	m := &CombatMap{}
	for i := range m.Raw {
		m.Raw[i] = TileBlank
	}
	wall, floor := DungeonArenaTiles(a.Number)
	// ⚠⚠ **更正(`docs/re/53`)**:底色是**地板**,不是空白。
	// `sub_FE48` 的第一個迴圈把 11 列 × 11 格全部填成 `byte_418DD`(地板):
	//
	//	for (edi = 0; edi < 0x0B; edi++) {
	//	    esi = &byte_3F8F4[edi*32]; eax = byte_418DD 重複四份
	//	    stosd; stosd; stosw; stosb        ; 11 B
	//	}
	//
	// 我第一版把整塊填成 0xFF,理由是別處有個 `rep stosd` 填 0xFFFFFFFF ——
	// 但那是**畫面緩衝**的初始化,不在這條路徑上。跑錯了來源。
	// 症狀會是「戰場除了牆之外全是空白」,而那在沒有畫面的測試裡看不出來。
	for y := 0; y < CombatSide; y++ {
		for x := 0; x < CombatSide; x++ {
			m.Tiles[y][x] = floor
			m.Raw[y*CombatRowStride+x] = floor
		}
	}
	set := func(x, y int, t byte) {
		if x < 0 || x >= CombatSide || y < 0 || y >= CombatSide {
			return
		}
		m.Tiles[y][x] = t
		m.Raw[y*CombatRowStride+x] = t
	}

	// --- sub_FD54:框 + 四角 + 正中央的擺設 ---
	for i := 0; i < CombatSide; i++ {
		set(i, dungeonArenaBoxLo, wall)
		set(i, dungeonArenaBoxHi, wall)
		set(dungeonArenaBoxLo, i, wall)
		set(dungeonArenaBoxHi, i, wall)
	}
	set(0, 0, TileBlank)
	set(CombatSide-1, 0, TileBlank)
	set(0, CombatSide-1, TileBlank)
	set(CombatSide-1, CombatSide-1, TileBlank)
	if k := DungeonKind(a.Here); k != 0 && k < DungeonMagic {
		if t := dungeonArenaCentreTile[k>>4]; t != 0 {
			set(dungeonArenaCentre, dungeonArenaCentre, t)
		}
	}

	// --- sub_FC7C:四面牆各自照鄰格處理 ---
	for side := 0; side < 4; side++ {
		k := DungeonKind(a.Around[side])
		switch {
		case k < DungeonRoomA:
			// 走得過去 → 開七格。原版之後還會再跑一次 sub_FB6C 的五格,
			// 那五格落在同樣的七格裡面,所以是無效動作 —— 不必模擬。
			dungeonArenaGap(set, side, dungeonArenaWideGap, floor)
		case k == DungeonWall || k == DungeonUnknownC || k == DungeonUnknownD:
			dungeonArenaSeal(set, side, a.Here, wall)
		default:
			dungeonArenaGap(set, side, dungeonArenaNarrowGap, floor)
		}
	}

	// --- sub_FE48:入場位置 ---
	px, py := DungeonArenaParty(DungeonArenaSide(a.Facing))
	m.PartyX, m.PartyY = px, py
	ex, ey := DungeonArenaEnemies(a.Facing)
	m.EnemyX, m.EnemyY = ex, ey
	copy(m.Raw[combatPartyX:], px[:])
	copy(m.Raw[combatPartyY:], py[:])
	copy(m.Raw[combatEnemyX:], ex[:])
	copy(m.Raw[combatEnemyY:], ey[:])
	for i := range m.EnemyKind {
		m.EnemyKind[i] = 0
		m.Raw[combatEnemyKind+i] = 0
	}
	return m
}

// dungeonArenaGap 在某一面牆上開一個 n 格寬的口(原版 `sub_FBF0` / `sub_FB6C`)。
//
// 兩支的差別只有起點與長度:七格從 2 起、五格從 3 起,都貼著內牆框那一條線。
func dungeonArenaGap(set func(x, y int, t byte), side, n int, floor byte) {
	start := (CombatSide - n) / 2 // 7 → 2、5 → 3
	for i := 0; i < n; i++ {
		switch side {
		case 0:
			set(start+i, dungeonArenaBoxLo, floor)
		case 1:
			set(dungeonArenaBoxHi, start+i, floor)
		case 2:
			set(start+i, dungeonArenaBoxHi, floor)
		case 3:
			set(dungeonArenaBoxLo, start+i, floor)
		}
	}
}

// dungeonArenaSeal 把某一面**整條邊**封成空白(原版 `sub_FAC4`)。
//
// ⚠ 封的是 11×11 的最外一圈那一整條,不是內牆框 —— 封起來之後那一側連
// 外環都過不去。另外,站在門口(0xE0)時北面與西面各多兩塊牆:
// 那兩對座標(5,2)/(5,8) 與 (2,5)/(8,5) 是原版寫死的,語意不明,照抄。
func dungeonArenaSeal(set func(x, y int, t byte), side int, here, wall byte) {
	for i := 0; i < CombatSide; i++ {
		switch side {
		case 0:
			set(i, 0, TileBlank)
		case 1:
			set(CombatSide-1, i, TileBlank)
		case 2:
			set(i, CombatSide-1, TileBlank)
		case 3:
			set(0, i, TileBlank)
		}
	}
	if DungeonKind(here) != DungeonDoorway {
		return
	}
	switch side {
	case 0:
		set(dungeonArenaCentre, 2, wall)
		set(dungeonArenaCentre, 8, wall)
	case 3:
		set(2, dungeonArenaCentre, wall)
		set(8, dungeonArenaCentre, wall)
	}
}
