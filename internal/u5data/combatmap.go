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
)

// CombatMap 是一張戰鬥地圖。
type CombatMap struct {
	// Tiles[y][x] 是 11×11 的地形。
	Tiles [CombatSide][CombatSide]byte
	// PartyX / PartyY 是六名隊員的入場位置。
	PartyX, PartyY [CombatPartySlots]byte
	// EnemyX / EnemyY 是十六個敵人的入場位置。
	EnemyX, EnemyY [CombatEnemySlots]byte
	// Raw 是完整的 352 B。每列 +11..+31 還有幾組沒解出來的資料
	// (列 1、2、4、5 各有一批,格式看起來像「另一種入場方向」的位置),
	// 沒有證據就不取名字,原樣留著。
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
