package u5data

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

// 特殊道具名表 —— 以及一個「同一批東西有兩套名字」的發現
//
// `DATA.OVL` 裡有**兩張**名字指標表指著同一批裝備:
//
//	0x1806  48 筆  長名字:`Leather Helm` `Spiked Helm` `Small Shield`
//	0x1946  ≥70 筆 短名字:`Leath Helm`   `Spkd. Helm`  `Sm. Shield`
//
// ★ 而且 **0x1946 的前 22 筆是 U 指令的特殊道具**,第 22 筆之後才接上
// 那 48 件裝備(用短名字)。所以它不是「另一份裝備表」,是**背包清單用的那一份**
// —— U5 的道具面板很窄,長名字塞不進去。
//
// 我原本以為 `WORKLIST` 上那條「特殊物品 158 項」是另一批道具,
// 實際上它是**這張表的總長度**:22 個特殊道具 + 48 件裝備(短名)+ 狀態名等等。
//
// # 索引就是 U 指令的 case 編號 − 16
//
// `docs/re/44` 用手抄出前 22 筆(`jumptable 0001A6DD case 16` 對上第 0 筆),
// 現在整張表都從玩家自己的檔案讀出來了 —— 索引 0 = `Magic Crpt` = case 16。
//
// ⚠ 第 5..12 筆叫 `(0`..`(7`。看起來像資料損毀,其實是原版留的佔位名
// (它們在跳表裡走 default,根本用不到)。**不要試著「修好」它們**。

const (
	// SpecialItemTableOffset 是短名字表的指標表位移。
	SpecialItemTableOffset = 0x1946
	// SpecialItemCount 是 U 指令用得到的前段長度(22 筆)。
	SpecialItemCount = 22
	// SpecialItemUseBase 是「索引 → U 指令 case 編號」的偏移。
	SpecialItemUseBase = 16
	// specialItemEquipFirst 是短名字表裡開始接裝備的索引。
	specialItemEquipFirst = 22
)

// SpecialItemTable 是短名字表。
type SpecialItemTable struct {
	// Names[i] 是第 i 筆的英文短名。
	Names [SpecialItemCount]string
	// EquipNames[i] 是同一張表後段那 48 件裝備的**短名**
	//(與 `ItemNames` 的長名一一對應,但字串不同)。
	EquipNames [ItemCount]string
}

// ParseSpecialItems 從 `DATA.OVL` 的內容取出短名字表。
func ParseSpecialItems(ovl []byte) (*SpecialItemTable, error) {
	need := SpecialItemTableOffset + (specialItemEquipFirst+ItemCount)*2
	if len(ovl) < need {
		return nil, fmt.Errorf("DATA.OVL 只有 %d B,放不下 0x%X 的短名字表",
			len(ovl), SpecialItemTableOffset)
	}
	read := func(i int) string {
		off := int(binary.LittleEndian.Uint16(ovl[SpecialItemTableOffset+i*2:])) + ItemPointerBias
		if off <= ItemPointerBias || off >= len(ovl) {
			return ""
		}
		e := indexByte(ovl, off)
		if e < 0 {
			return ""
		}
		if s := string(ovl[off:e]); printableASCII(s) {
			return s
		}
		return ""
	}
	t := &SpecialItemTable{}
	for i := 0; i < SpecialItemCount; i++ {
		t.Names[i] = read(i)
	}
	for i := 0; i < ItemCount; i++ {
		t.EquipNames[i] = read(specialItemEquipFirst + i)
	}
	// 兩頭各釘一個錨:第 0 筆與第 21 筆是 U 指令那份清單的頭尾,
	// 而第 22 筆必須接上裝備 —— 三個一起中,表就沒有滑動的餘地。
	if t.Names[0] != "Magic Crpt" {
		return nil, fmt.Errorf("短名字表第 0 筆是 %q,預期 \"Magic Crpt\"", t.Names[0])
	}
	if t.Names[SpecialItemCount-1] != "Wooden Box" {
		return nil, fmt.Errorf("短名字表第 %d 筆是 %q,預期 \"Wooden Box\"",
			SpecialItemCount-1, t.Names[SpecialItemCount-1])
	}
	if t.EquipNames[0] != "Leath Helm" {
		return nil, fmt.Errorf("短名字表第 %d 筆是 %q,預期 \"Leath Helm\"",
			specialItemEquipFirst, t.EquipNames[0])
	}
	return t, nil
}

// LoadSpecialItems 從 `DATA.OVL` 讀出短名字表。
func LoadSpecialItems(dir string) (*SpecialItemTable, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "DATA.OVL"))
	if err != nil {
		return nil, err
	}
	return ParseSpecialItems(raw)
}

// NameForUseCode 依 U 指令的 case 編號回傳英文短名;不在範圍內回空字串。
func (t *SpecialItemTable) NameForUseCode(code int) string {
	if t == nil {
		return ""
	}
	i := code - SpecialItemUseBase
	if i < 0 || i >= SpecialItemCount {
		return ""
	}
	return t.Names[i]
}

// SpecialItemPlaceholder 回報這個名字是不是原版留的佔位名(`(0`..`(7`)。
//
// 那八筆在跳表裡走 default,用不到 —— 拿它們去查譯名只會得到一堆問號,
// 所以清單要跳過。
func SpecialItemPlaceholder(name string) bool {
	return len(name) == 2 && name[0] == '(' && name[1] >= '0' && name[1] <= '7'
}
