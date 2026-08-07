package u5data

import (
	"fmt"
	"os"
)

// 戰鬥地圖(`BRIT.CBT` / `DUNGEON.CBT`)
//
// 載入寫在 `sub_2E51C`:
//
//	mov  eax, arg_0
//	lea  edx, [eax+eax*4]      ; 5×
//	lea  edx, [eax+edx*2]      ; eax + 10× = 11×
//	shl  edx, 5                ; ×32  →  offset = 編號 × 352
//	push edx
//	push 160h                  ; 352 B
//	push offset byte_3F8F4
//	push offset aBritCbt
//	call sub_2C740
//
// 每張 **352 B = 11 列 × 32 B**。每列的前 11 B 是地形(11×11 的戰場),
// 後 21 B 放各種入場位置 —— 地形與位置**交錯**在同一塊裡,不是分開兩段。
//
//	BRIT.CBT     5,632 B → 16 張(地表的各種地形:草原、森林、沼澤、海灘…)
//	DUNGEON.CBT 39,424 B → 112 張(地牢房間)
const (
	// CombatMapSize 是一張戰鬥地圖的位元組數。
	CombatMapSize = 352
	// CombatSide 是戰場的邊長(11×11)。
	CombatSide = 11
	// CombatRowStride 是每一列在檔案裡佔的位元組數。
	CombatRowStride = 32
	// CombatPartySlots 是玩家方的位置數(隊伍上限 6 人)。
	CombatPartySlots = 6
	// CombatEnemySlots 是敵方的位置數。
	CombatEnemySlots = 16
)

// 入場位置在 352 B 裡的位移,由 `sub_2E51C` 的搬移逐行反推:
//
//	byte_3F95F − byte_3F8F4 = 107 = 3×32 + 11   玩家 X ×6
//	byte_3F965 − byte_3F8F4 = 113 = 3×32 + 17   玩家 Y ×6
//	byte_3F9BF − byte_3F8F4 = 203 = 6×32 + 11   敵人 X ×16
//	byte_3F9DF − byte_3F8F4 = 235 = 7×32 + 11   敵人 Y ×16
const (
	combatPartyX = 107
	combatPartyY = 113
	combatEnemyX = 203
	combatEnemyY = 235
	// combatEnemyKind 是 16 個敵人各自的種類碼(5×32 + 11 = 171)。
	//
	// 依據:`sub_FE48` 在隨機遭遇那條路徑上寫 `byte_3F99F[槽] = 生物*4 + 0x40`,
	// 而 `0x3F99F − 0x3F8F4 = 171`。地牢房間則是從檔案裡讀 —— 兩條路徑
	// 共用同一個陣列。`DUNGEON.CBT` 的這 16 B 幾乎全是 4 的倍數 + 0x40,
	// 對得上生物編號的公式。
	combatEnemyKind = 171
)

// CombatMap 是一張戰鬥地圖。
type CombatMap struct {
	// Tiles[y][x] 是 11×11 的地形。
	Tiles [CombatSide][CombatSide]byte
	// PartyX / PartyY 是六名隊員的入場位置。
	PartyX, PartyY [CombatPartySlots]byte
	// EnemyX / EnemyY 是十六個敵人的入場位置。
	EnemyX, EnemyY [CombatEnemySlots]byte
	// EnemyKind 是十六個敵人各自的種類碼;0 代表這一槽空著。
	//
	// 地表的 `BRIT.CBT` 這一段全是 0(地表遭遇的怪物由撞到的那個物件決定),
	// 地牢房間的 `DUNGEON.CBT` 才用得到。
	EnemyKind [CombatEnemySlots]byte
	// Raw 是完整的 352 B。
	//
	// ✅ 列 1、2、4、5 的 +11..+22 **已解**(`docs/re/48` §6):那是**四個入場方向**
	// 各自的隊伍位置,側 = 由朝向決定(`DungeonArenaSide`)。
	// 上面的 `PartyX`/`PartyY` 取的是列 3(= 朝北入場那一組)——
	// 這是原版 `combatPartyX = 3×32+11` 的那一組,不是任意挑的,
	// 但**地牢遭遇戰會依朝向換組**,見 `DungeonArenaParty`。
	Raw [CombatMapSize]byte
}

// CombatMapSet 是一個 `.CBT` 檔裡的全部戰鬥地圖。
type CombatMapSet struct {
	Maps []CombatMap
}

// LoadCombatMaps 讀入 `BRIT.CBT` 或 `DUNGEON.CBT`。
func LoadCombatMaps(path string) (*CombatMapSet, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseCombatMaps(raw)
}

// ParseCombatMaps 解析 `.CBT` 的內容。
func ParseCombatMaps(raw []byte) (*CombatMapSet, error) {
	if len(raw) == 0 || len(raw)%CombatMapSize != 0 {
		return nil, fmt.Errorf("戰鬥地圖檔 %d B,不是 %d 的整數倍", len(raw), CombatMapSize)
	}
	n := len(raw) / CombatMapSize
	s := &CombatMapSet{Maps: make([]CombatMap, n)}
	for i := 0; i < n; i++ {
		blk := raw[i*CombatMapSize : (i+1)*CombatMapSize]
		m := &s.Maps[i]
		copy(m.Raw[:], blk)
		for y := 0; y < CombatSide; y++ {
			copy(m.Tiles[y][:], blk[y*CombatRowStride:y*CombatRowStride+CombatSide])
		}
		copy(m.PartyX[:], blk[combatPartyX:])
		copy(m.PartyY[:], blk[combatPartyY:])
		copy(m.EnemyX[:], blk[combatEnemyX:])
		copy(m.EnemyY[:], blk[combatEnemyY:])
		copy(m.EnemyKind[:], blk[combatEnemyKind:])
	}
	if err := s.validate(); err != nil {
		return nil, err
	}
	return s, nil
}

// validate 用「算錯就一定違反」的性質擋住位移偏掉。
//
// 入場位置全部落在 0..10 是最強的一條:352 B 裡大部分位元組是地形碼
// (值域到 0xB3)與 0,只有真正的座標欄位會整批落在 0..10。位移偏一格
// 就會撈到地形碼,立刻超界。
func (s *CombatMapSet) validate() error {
	for i := range s.Maps {
		m := &s.Maps[i]
		for k := 0; k < CombatPartySlots; k++ {
			if m.PartyX[k] >= CombatSide || m.PartyY[k] >= CombatSide {
				return fmt.Errorf("第 %d 張圖的第 %d 個隊員入場位置是 (%d,%d),超出 %d×%d",
					i, k, m.PartyX[k], m.PartyY[k], CombatSide, CombatSide)
			}
		}
		for k := 0; k < CombatEnemySlots; k++ {
			if m.EnemyX[k] >= CombatSide || m.EnemyY[k] >= CombatSide {
				return fmt.Errorf("第 %d 張圖的第 %d 個敵人入場位置是 (%d,%d),超出 %d×%d",
					i, k, m.EnemyX[k], m.EnemyY[k], CombatSide, CombatSide)
			}
		}
	}
	return nil
}

// At 回傳 (x, y) 的地形。
func (m *CombatMap) At(x, y int) byte {
	if x < 0 || x >= CombatSide || y < 0 || y >= CombatSide {
		return TileBlank
	}
	return m.Tiles[y][x]
}

// 挑戰鬥地圖
//
// 原版 `sub_2E58C` 在載入前先決定用哪一張。判斷依序看三件事:
// 隊伍在不在船上、敵人是不是船、以及**敵人腳下那一格的地形**。

// 幾個特別的種類碼。
const (
	// EnemyShip 是敵船(海盜船)。
	EnemyShip = 0x2C
	// EnemySceptre 是權杖(`sub_2E58C` 對它另有一段「The Sceptre is reclaimed!」)。
	EnemySceptre = 0xFC
	// enemyWaterFamily 是「水生怪物」那一族(種類 & 0xF0 == 0x80)—— 一律打水戰。
	enemyWaterFamily = 0x80
)

// 戰鬥地圖編號。名字取自 `u5dump cbt` 畫出來的樣子。
const (
	CombatMapSceptre  = 10 // 權杖
	CombatMapShipSea  = 11 // 在船上、水戰
	CombatMapShipVs   = 12 // 對上敵船(自己不在船上)
	CombatMapShipLnd  = 13 // 在船上、陸戰
	CombatMapShipShip = 14 // 在船上、對上敵船
	CombatMapOpenSea  = 15 // 純水面
)

// IsWaterBattle 回報敵人腳下那一格算不算水戰(原版 `sub_2E58C` 的 var_8)。
//
//	tile < 4                                → 水
//	tile & 0xFE == 0x6A                     → 不是(那兩格是例外)
//	tile & 0xF0 == 0x60                     → 水
//	其餘                                     → 不是
func IsWaterBattle(terrain int) bool {
	if terrain < 4 {
		return true
	}
	if terrain&0xFE == 0x6A {
		return false
	}
	return terrain&0xF0 == 0x60
}

// terrainCombatMap 是「敵人腳下的地形 → 戰鬥地圖」的對照,出自 `sub_2E58C`
// 的 73-case 跳表(`jpt_2E711`)。表裡沒有的地形走 fallback,見 SelectCombatMap。
var terrainCombatMap = map[int]int{
	1: 15, 2: 15, 3: 15, // 水
	4: 1,
	5: 2,
	6: 3, 8: 3,
	7: 4, 30: 4, 31: 4,
	9: 5, 10: 5,
	11: 6, 12: 6, 13: 6, 14: 6, 15: 6,
	29: 7, 72: 7, 73: 7,
	0x6A: 7, 0x6B: 7,
	68: 8,
}

// SelectCombatMap 依原版 `sub_2E58C` 挑出要用哪一張戰鬥地圖。
//
//	enemyKind  敵人的種類碼(物件槽的 +0,已經 & 0xFC)
//	terrain    敵人腳下那一格的地形 tile
//	transport  隊伍當前的載具(`byte_3E08C`)
//	inWorld    玩家是不是在大地圖上(地點編號 0)
func SelectCombatMap(enemyKind, terrain int, transport byte, inWorld bool) int {
	if enemyKind == EnemySceptre {
		return CombatMapSceptre
	}
	water := IsWaterBattle(terrain)
	if enemyKind&0xF0 == enemyWaterFamily {
		water = true
	}
	// 隊伍在船上(載具 & 0xF8 == 0x20 涵蓋揚帆 0x20..0x23 與大船 0x24..0x27)。
	if transport&0xF8 == 0x20 {
		switch {
		case enemyKind == EnemyShip:
			return CombatMapShipShip
		case water:
			return CombatMapShipSea
		default:
			return CombatMapShipLnd
		}
	}
	if enemyKind == EnemyShip {
		return CombatMapShipVs
	}
	if water {
		return CombatMapOpenSea
	}
	if m, ok := terrainCombatMap[terrain]; ok {
		return m
	}
	// 表裡沒有的地形:在大地圖上用 2(草原),在場景裡用 8。
	if inWorld {
		return 2
	}
	return 8
}
