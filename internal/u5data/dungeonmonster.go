package u5data

// 地牢的遊蕩怪物(原版 `sub_4460` 生成 / `sub_4594` 找落點 / `sub_4C6C` 每步移動)
//
// 每一層地牢有**一隻**遊蕩怪物,住在物件槽 1(`dword_3E474`),它的上一格
// 記在物件槽 2(`dword_3E47C`)。走到隊伍腳下就開打。
//
// # 八種怪物,一種一張圖
//
// 生成時抽 `random(0, 7)`,查 `byte_3F3D0[8]` 得生物索引 —— 而**索引 0..7
// 同時就是 `MON0.16`..`MON7.16` 的檔號**(`off_41BC0[記錄[0]]`)。
// 這解掉了 WORKLIST 上「怪物圖 MON0–7.16(8 組)」為什麼剛好八組:
// 那不是任意八種怪物的圖組,而是**地牢遊蕩怪物的名冊**。
//
//	MON0  索引 20  Giant Rat     巨鼠
//	MON1  索引 21  Bat           蝙蝠
//	MON2  索引 22  Giant Spider  巨蛛
//	MON3  索引 23  Ghost         幽魂
//	MON4  索引 24  Slime         泥怪
//	MON5  索引 25  Gremlin       小鬼
//	MON6  索引 28  Gazer         凝視魔
//	MON7  索引 27  Reaper        收割者
//
// ★ 最後兩筆是 **28、27** —— 順序反的。這種「資料本身不整齊」正是它是真表
// 而不是我照印象填出來的證據(`rulebook/62`)。

// DungeonMonsterKinds 是遊蕩怪物的種類數,也是 `MONn.16` 的組數。
const DungeonMonsterKinds = 8

// dungeonMonsterIndex 是 `byte_3F3D0` 的八項 —— **生物索引**,不是生物編號。
//
// 抽自 `WORRIORS.EXP` 線性 0x3F3D0(檔案位移 = 線性 + 0x200)。
var dungeonMonsterIndex = [DungeonMonsterKinds]byte{
	0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1C, 0x1B,
}

// dungeonMonsterView 是 `byte_3F3D8` 的八項。
//
// ⚠ 語意**未解**。取用它的 `sub_31F0` 只抽 0x90 / 0x60 / 0x0F 三個遮罩,
// 而 0x0F 在這八項裡全是 0 —— 看得出是第一人稱視圖的擺法(幾隻、多遠),
// 不是戰鬥數值(拿去對 `byte_3F1D0` 的怪物旗標,八項全不合)。
// 第一人稱透視還沒做,所以先原樣留著,不取名字。
var dungeonMonsterView = [DungeonMonsterKinds]byte{
	0x60, 0xA0, 0x00, 0x90, 0x80, 0x60, 0x00, 0x00,
}

// DungeonMonsterCreatureIndex 回傳第 k 種的生物索引(0..47)。
func DungeonMonsterCreatureIndex(k int) int {
	if k < 0 || k >= DungeonMonsterKinds {
		return -1
	}
	return int(dungeonMonsterIndex[k])
}

// DungeonMonsterCreature 回傳第 k 種的生物編號(餵得進 CreatureTable.Name)。
func DungeonMonsterCreature(k int) byte {
	i := DungeonMonsterCreatureIndex(k)
	if i < 0 {
		return 0
	}
	return CreatureBase + byte(i)*4
}

// DungeonMonsterView 回傳第 k 種的視圖位元組(語意未解,見上)。
func DungeonMonsterView(k int) byte {
	if k < 0 || k >= DungeonMonsterKinds {
		return 0
	}
	return dungeonMonsterView[k]
}

// DungeonMonsterStill 是**不會移動**的那一種(`sub_4C6C` 的 `cmp dl, 1Bh`)。
//
// ★ 0x1B = 索引 27 = **Reaper(收割者)** —— U5 裡它是紮了根的樹妖。
// 「唯一跳過移動的那一種剛好是不會走的怪物」是這張表對得上的最強證據:
// 表要是偏一格,豁免的就會是別種能走的怪物。
const DungeonMonsterStill = 0x1B

// 生成與移動的嘗試次數(兩處都是 8,原版 `cmp [ebp+var_8], 8`)。
const (
	DungeonSpawnTries = 8
	DungeonMoveTries  = 8
)

// DungeonSpawnBackoff 是「想走回上一格」時要擲的骰:只有 `random(0,7) == 1`
// 才准回頭(原版 `sub_28E14(0,7); dec eax; jnz retry`)。
const DungeonSpawnBackoff = 7

// DungeonSpawnAllows 回報遊蕩怪物**生**得生在這一格(原版 `sub_4594`)。
//
//	高四位元 < 0x60          可以(通道、梯子、寶箱、噴泉)
//	高四位元 == 0x70(門)    可以
//	其餘                     不行
func DungeonSpawnAllows(tile byte) bool {
	k := DungeonKind(tile)
	return k < DungeonTrap || k == DungeonDoor
}

// DungeonMonsterBlocked 回報遊蕩怪物**走**不走得進這一格(原版 `sub_4C6C`)。
//
//	高四位元 == 0x60(陷阱)      不行
//	高四位元 == 0x80(魔法陷阱)  不行
//	高四位元 >= 0xA0(房間 / 牆)  不行
//	其餘                          可以
//
// ⚠ 這兩個判定**不是互補的**:0x90 走得進去卻生不出來。0x90 在地牢資料裡
// 沒有已命名的語意(`DungeonMagic` 是 0x80),所以照抄兩條規則,不硬湊成一條。
func DungeonMonsterBlocked(tile byte) bool {
	k := DungeonKind(tile)
	return k == DungeonTrap || k == DungeonMagic || k >= DungeonRoomA
}

// DungeonAttackDirection 算出「從哪邊被攻擊」(原版 `sub_5008` 的四段比較)。
//
// 傳進來的是**怪物的上一格**與隊伍的位置。判定順序照原版,不可重排 ——
// 它是 if / else if 的鏈,先中的先算:
//
//	隊伍 x == (怪 x − 1) & 7  → 東(怪在東邊)
//	隊伍 x == (怪 x + 1) & 7  → 西
//	隊伍 y == (怪 y − 1) & 7  → 南
//	其餘                       → 北
//
// `& 7` 不是隨手加的:地牢一層是 8×8 而且**繞回**,所以 x = 0 的西邊是 x = 7。
func DungeonAttackDirection(mx, my, px, py int) int {
	switch {
	case px == (mx-1)&7:
		return 1 // 東
	case px == (mx+1)&7:
		return 3 // 西
	case py == (my-1)&7:
		return 2 // 南
	}
	return 0 // 北
}
