package u5data

// 門與鎖:格子狀態、四支原版函式,以及 An Ylem 抹得掉的東西。
//
// 四條規則散在四支函式裡,湊起來才看得懂一整套:
//
//	`sub_18D18` An Sanct(開鎖)     0xB9 → 0xB8、0xBB → 0xBA(`dec`)
//	`sub_19354` An Ex Por(魔法上鎖) 0xB8/0xB9 → 0x97、0xBA/0xBB → 0x98
//	`sub_1D15C` In Ex Por(魔法解鎖) 0x97 → 0xB8、0x98 → 0xBA
//	`sub_18C00` An Ylem(消除)       清單裡的格子一律 → 0x44
//
// ⚠ **An Sanct 與 In Ex Por 不是同一條規則。** 前者做 `tile − 1` 開一般的鎖,
// 後者只認 0x97/0x98 這兩個魔法鎖。我第一版把兩個都寫成 `tile − 1`,
// 那會讓 In Ex Por 把魔法鎖變成 0x96 —— 一個不存在的格子。
//
// ⚠ **上鎖與解鎖也不對稱。** An Ex Por 吃「一般門」與「上鎖門」兩種
//(0xB8 與 0xB9 都變成 0x97),而 In Ex Por 只還原成**一般門**。
// 所以拿 An Ex Por 鎖住一扇本來就上鎖的門、再用 In Ex Por 解開,
// 那扇門會變成沒鎖的。這是原版的行為,不是四捨五入。

// 上鎖的門(地表 / 城鎮):**0xB9 與 0xBB 是鎖著的**。
const (
	TileLockedDoor      = 0xB9
	TileLockedMagicDoor = 0xBB
)

// TileIsLockedDoor 回報這個世界 tile 是不是鎖著的門。
func TileIsLockedDoor(tile byte) bool {
	return tile == TileLockedDoor || tile == TileLockedMagicDoor
}

// 魔法鎖與一般門的四個格子編號。
const (
	TileMagicLockedA = 0x97
	TileMagicLockedB = 0x98
	TileDoorA        = 0xB8
	TileDoorB        = 0xBA
)

// MagicLock 回傳這扇門魔法上鎖之後變成什麼;不是門就回 0。
func MagicLock(tile byte) byte {
	switch tile {
	case TileDoorA, TileLockedDoor:
		return TileMagicLockedA
	case TileDoorB, TileLockedMagicDoor:
		return TileMagicLockedB
	}
	return 0
}

// MagicUnlock 回傳這扇魔法鎖解開之後變成什麼;不是魔法鎖就回 0。
func MagicUnlock(tile byte) byte {
	switch tile {
	case TileMagicLockedA:
		return TileDoorA
	case TileMagicLockedB:
		return TileDoorB
	}
	return 0
}

// AnYlemTiles 是 An Ylem(消除)吃得掉的 tile。
//
// 從 `sub_18C00` 的 32-case 跳表讀出來(索引 = tile − 0x90),外加特例的 0x5B。
// 那一批是各種力場與能量的圖 —— An Ylem 就是把它們抹掉。
// TileBrickFloor 是室內地板(An Ylem 抹掉東西之後填回去的那一格)。
//
// 依據是 `sub_18C00` 的 `mov byte ptr [esi], 44h` —— 寫死的 0x44。
// 佐證:0x44 在 `TOWNE.DAT` 出現 3,142 次、`CASTLE.DAT` 3,927 次
// (兩者都是第二多),而在世界地圖 `BRIT.DAT` **一次都沒有** ——
// 它是純室內地形。An Ylem 只能在戰鬥與城鎮裡施放(場合表 0x05),
// 兩邊對得上。
const TileBrickFloor = 0x44

// AnYlemTiles 是 An Ylem 抹得掉的東西。
var AnYlemTiles = map[byte]bool{
	0x5B: true,
	0x90: true, 0x91: true, 0x92: true, 0x93: true,
	0x9D: true,
	0xA5: true, 0xA6: true, 0xA8: true, 0xA9: true,
	0xAD: true, 0xAE: true, 0xAF: true,
}
