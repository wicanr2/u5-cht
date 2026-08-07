package u5data

import (
	"fmt"
	"os"
	"path/filepath"
)

// 地圖物件表(`BRIT.OOL` / `UNDER.OOL`,各 256 B)
//
// 啟動時 `sub_0` 直接把整份讀進 `dword_3E46C`:
//
//	sub_2C740(sub_10910(), dword_3E46C, 0x100, 0)
//	          └─ byte_3E0A5 == 0 ? "A:BRIT.OOL" : "A:UNDER.OOL"
//
// 256 B = **32 槽 × 8 B**。這張表放的是「會動的東西」:隊伍本身、玩家買的坐騎與船、
// 地上的物品、遊蕩的怪物。城鎮裡的 NPC 走的是另一套(`.NPC` + 排程),不在這裡。
//
// 實際檔案佐證:`BRIT.OOL` 只有槽 0 有東西 —— tile 28(0x1C,步行的隊伍),
// 與存檔的載具欄位同值;`UNDER.OOL` 有五個,樓層欄都是 0xFF(地下世界 = −1)。
//
// ⚠ 槽 0 的座標是 (86,107),那是地點表第 17 筆(一個無名地點)的位置,
// **不是** `INIT.GAM` 的開場位置(IOLO'S HUT,(45,62))。這兩份檔案不同步 ——
// `BRIT.OOL` 是世界的預設物件,開新遊戲要讀的是 `INIT.OOL`。
const (
	// ObjectSlots 是物件槽的數量。
	ObjectSlots = 32
	// ObjectRecordSize 是一個槽的大小。
	ObjectRecordSize = 8
	// ObjectFileSize 是一份 .OOL 的大小。
	ObjectFileSize = ObjectSlots * ObjectRecordSize
	// PartyObjectSlot 是隊伍自己佔的槽(與 .NPC 的 PartySlot 同樣是 0)。
	PartyObjectSlot = 0
)

// 槽內的欄位位移。
//
// 由買馬那段(`sub_118CC` 尾段)逐行讀出來的:它把 +0 與 +1 都設成馬的 tile、
// +2/+3 設成座標、+4 設成當前樓層、+5..+7 清零。
//
// +0 與 +1 的差別在船的轉向那段(`sub_23FC`)看得出來:轉向只改 +1
// (tile 0x2C/0x2E ⇄ 0x2D/0x2F),+0 不動 —— 所以 **+0 是這是什麼東西、
// +1 是此刻該畫哪一格**。
const (
	ObjKind  = 0 // 種類(基礎 tile);0 = 這個槽是空的
	ObjTile  = 1 // 目前要畫的 tile(船會隨朝向變)
	ObjX     = 2
	ObjY     = 3
	ObjFloor = 4 // 0 地表 / 0xFF 地下世界;場景裡是樓層
	// +5..+7 買馬時清零,但全檔沒有其他讀寫處 —— 用途不明,原樣保留不猜。
)

// 幾個有名字的種類碼。
const (
	// ObjEmpty 是空槽。
	ObjEmpty = 0
	// ObjCreatureBase 是怪物種類碼的起點,與生物名表的 CreatureBase 同源。
	ObjCreatureBase = 0x40
	// ObjSpecialFC 是某種特殊物件;`sub_48C` 開場會掃全表看在不在場,
	// 但它是什麼還沒追到。標記出來免得被當成一般物件處理。
	ObjSpecialFC = 0xFC
)

// MapObject 是一個物件槽。
type MapObject struct {
	Kind  byte
	Tile  byte
	X, Y  int
	Floor int
	// Raw 是完整的 8 B;+5..+7 還沒解,保留原樣。
	Raw [ObjectRecordSize]byte
}

// Present 回報這個槽有沒有東西。
func (o *MapObject) Present() bool { return o.Kind != ObjEmpty }

// IsCreature 回報這是不是怪物(而非坐騎、船或地上的物品)。
func (o *MapObject) IsCreature() bool {
	return o.Kind >= ObjCreatureBase && o.Kind != ObjSpecialFC
}

// ObjectSet 是一份物件表。
type ObjectSet struct {
	Objects [ObjectSlots]MapObject
}

// LoadObjects 讀入 `BRIT.OOL` 或 `UNDER.OOL`。
func LoadObjects(path string) (*ObjectSet, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseObjects(raw)
}

// LoadWorldObjects 讀入地表與地下世界兩份物件表。
func LoadWorldObjects(dir string) (surface, under *ObjectSet, err error) {
	surface, err = LoadObjects(filepath.Join(dir, "BRIT.OOL"))
	if err != nil {
		return nil, nil, err
	}
	under, err = LoadObjects(filepath.Join(dir, "UNDER.OOL"))
	if err != nil {
		return nil, nil, err
	}
	return surface, under, nil
}

// LoadSaveObjects 讀入存檔的物件表(`SAVED.OOL` / `INIT.OOL`)。
//
// **512 B 的版本是「地表 256 B + 地下 256 B」**。實測佐證:`SAVED.OOL`
// 前半全零(存檔當下地表沒有物件)、後半與 `UNDER.OOL` 逐位元組相同。
//
// 256 B 的版本只有一半 —— `INIT.OOL` 就是這樣,而且與 `UNDER.OOL` 完全相同,
// 所以那 256 B 是**地下世界**那一份,地表視為空的。
func LoadSaveObjects(path string) (surface, under *ObjectSet, err error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	switch len(raw) {
	case ObjectFileSize * 2:
		if surface, err = ParseObjects(raw[:ObjectFileSize]); err != nil {
			return nil, nil, err
		}
		under, err = ParseObjects(raw[ObjectFileSize:])
		return surface, under, err
	case ObjectFileSize:
		under, err = ParseObjects(raw)
		return &ObjectSet{}, under, err
	}
	return nil, nil, fmt.Errorf("存檔物件表 %d B,預期 %d 或 %d",
		len(raw), ObjectFileSize, ObjectFileSize*2)
}

// ParseObjects 解析一份 256 B 的物件表。
func ParseObjects(raw []byte) (*ObjectSet, error) {
	if len(raw) < ObjectFileSize {
		return nil, fmt.Errorf("物件表 %d B,至少要 %d B", len(raw), ObjectFileSize)
	}
	s := &ObjectSet{}
	for i := 0; i < ObjectSlots; i++ {
		rec := raw[i*ObjectRecordSize : (i+1)*ObjectRecordSize]
		o := &s.Objects[i]
		copy(o.Raw[:], rec)
		o.Kind = rec[ObjKind]
		o.Tile = rec[ObjTile]
		o.X = int(rec[ObjX])
		o.Y = int(rec[ObjY])
		o.Floor = int(int8(rec[ObjFloor]))
	}
	return s, nil
}

// At 回報某一格上有沒有物件(只看同一層)。
func (s *ObjectSet) At(x, y, floor int) (*MapObject, bool) {
	if s == nil {
		return nil, false
	}
	for i := range s.Objects {
		o := &s.Objects[i]
		if o.Present() && o.X == x && o.Y == y && o.Floor == floor {
			return o, true
		}
	}
	return nil, false
}

// FreeSlot 找一個空槽。槽 0 是隊伍的,不給別人用。
func (s *ObjectSet) FreeSlot() (int, bool) {
	if s == nil {
		return 0, false
	}
	for i := PartyObjectSlot + 1; i < ObjectSlots; i++ {
		if !s.Objects[i].Present() {
			return i, true
		}
	}
	return 0, false
}

// Spawn 在指定位置放一個物件,回傳用掉的槽號。
//
// 照原版買馬那段的寫法:種類與顯示 tile 都設成同一個值,座標與樓層照給,
// 後三個位元組清零。
func (s *ObjectSet) Spawn(kind byte, x, y, floor int) (int, bool) {
	i, ok := s.FreeSlot()
	if !ok {
		return 0, false
	}
	o := &s.Objects[i]
	*o = MapObject{Kind: kind, Tile: kind, X: x, Y: y, Floor: floor}
	o.Raw[ObjKind] = kind
	o.Raw[ObjTile] = kind
	o.Raw[ObjX] = byte(x)
	o.Raw[ObjY] = byte(y)
	o.Raw[ObjFloor] = byte(int8(floor))
	return i, true
}

// Remove 清空一個槽(玩家騎上坐騎、撿起物品時)。
func (s *ObjectSet) Remove(slot int) {
	if s == nil || slot < 0 || slot >= ObjectSlots {
		return
	}
	s.Objects[slot] = MapObject{}
}

// 船專用的兩個欄位
//
// `sub_16F08`(上船)與 `sub_177AC`(下船)把槽裡的 +5 與 +7 當成船的屬性用:
//
//	movzx eax, byte ptr dword_3E470+1[edi*8]   ; 3E46C + 5 + slot*8 → 船的耐久
//	movzx esi, byte ptr dword_3E470+3[edi*8]   ; 3E46C + 7 + slot*8 → 船上的小艇數
//
// 耐久 < 10 會警告「DANGER: SHIP BADLY DAMAGED!」,小艇數 0 會警告
// 「WARNING: NO SKIFFS ON BOARD!」。上船時兩個值搬到**槽 0**(隊伍那一槽),
// 下船時再搬回新生成的船物件 —— 所以槽 0 在載具期間兼作「當前載具的狀態」。
//
// 這填掉了先前標「用途不明」的三個位元組其中兩個;+6 仍然沒有讀寫處。
const (
	ObjShipHull   = 5 // 船的耐久
	ObjShipSkiffs = 7 // 船上載的小艇數
)

// Hull 回傳船的耐久。
func (o *MapObject) Hull() int { return int(o.Raw[ObjShipHull]) }

// Skiffs 回傳船上載的小艇數。
func (o *MapObject) Skiffs() int { return int(o.Raw[ObjShipSkiffs]) }

// SetHull / SetSkiffs 設定船的兩個屬性。
func (o *MapObject) SetHull(v int)   { o.Raw[ObjShipHull] = byte(v) }
func (o *MapObject) SetSkiffs(v int) { o.Raw[ObjShipSkiffs] = byte(v) }
