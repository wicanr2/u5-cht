package u5data

import (
	"fmt"
	"path/filepath"
)

// NPC 資料(`TOWNE/DWELLING/CASTLE/KEEP.NPC`)
//
// 格式出自原版的載入函式 `sub_8858`(docs/re/04):
//
//	檔案   = {TOWNE,DWELLING,CASTLE,KEEP}.NPC[(地點編號-1)/8]   ← 與場景檔同一套分檔
//	位移   = ((地點編號-1) % 8) × 576
//	  +0x000  512 B  32 × 16 B 排程記錄
//	  +0x200   32 B  每個 NPC 的**生物編號**(0 = 這個槽是空的)
//	  +0x220   32 B  每個 NPC 的對話號碼
//
// 8 個地點 × 576 B = 4,608 B = 檔案大小 ✓
const (
	// NPCsPerLocation 是每個地點的 NPC 槽數。
	NPCsPerLocation = 32
	// NPCRecordSize 是一筆排程記錄的大小。
	NPCRecordSize = 16
	// NPCBlockSize 是一個地點的 NPC 資料大小。
	NPCBlockSize = NPCsPerLocation*NPCRecordSize + NPCsPerLocation*2 // 576
	// NPCFileSize 是一個 .NPC 檔的大小(8 個地點)。
	NPCFileSize = NPCBlockSize * 8 // 4,608

	// PartySlot 是隊伍自己佔的槽。
	//
	// 原版的 NPC 更新迴圈 `sub_8924` 從 index 1 起跑,0 號留給玩家隊伍。
	// 檔案裡 0 號槽的內容**不是正常的 NPC 記錄**:多數地點是 0,
	// 月光城與不列顛城留著 0x1C(聖者的生物編號),還有一個地點是 0x29
	// (連 4 的倍數都不是)—— 都是殘值,讀到就略過,不要拿來推論。
	PartySlot = 0
)

// NPCFiles 是 NPC 檔名,順序與 SceneFiles 相同(原版 sub_8858 的 switch 就是這個順序)。
var NPCFiles = [4]string{"TOWNE.NPC", "DWELLING.NPC", "CASTLE.NPC", "KEEP.NPC"}

// 對話號碼的分類。原版 `sub_1B52C` 依這個值分派:
//
//	0            → "No response!"
//	1 .. 0x7F    → 查 .TLK(id 就是這個值)
//	0x80 .. 0xFC → 商人;低 7 位選店類,還要過營業時間才談得成
//	0xFD         → "Don't hurt me! Please go away!"(被嚇到的平民)
//	0xFE / 0xFF  → 兩個特殊角色,各自有專屬流程
const (
	DialogueNone       = 0x00
	DialogueShopFirst  = 0x80
	DialogueShopLast   = 0xFC
	DialogueFrightened = 0xFD
	DialogueSpecialFE  = 0xFE
	DialogueSpecialFF  = 0xFF
)

// NPCSchedule 是一筆 16 B 的排程記錄。
//
// 欄位位置由 `sub_9358` 證實:它用 `rec[slot+3]` / `rec[slot+6]` / `rec[slot+9]`
// 取 X / Y / 樓層,slot ∈ 0..2。
type NPCSchedule struct {
	AI    [3]byte // 每個 slot 的行為型別
	X     [3]byte // 每個 slot 的場景 X
	Y     [3]byte // 每個 slot 的場景 Y
	Floor [3]byte // 每個 slot 的樓層
	Times [4]byte // 四個時刻(見 Slot)
}

// Slot 回報 hour 這個時刻該用哪一個位置(0..2)。
//
// 原版 `sub_9C7C` 的算法很精簡:
//
//	for i in 0..3:  d[i] = (hour - Times[i]) & 0xFF      // 8-bit 減法,24 小時自然環繞
//	slot = argmin(d[0], d[1], d[2])
//	if d[slot] > d[3]: slot = 1
//
// 也就是「取最近剛過的那個時刻」。**四個時刻但只有三個位置** —— 最後那行就是答案:
// 第四個時刻也指向 slot 1,所以 NPC 一天會回到位置 1 兩次。
// (實例:不列顛城的 1 號 NPC 06:00 上工、11:00 外出、13:00 回崗位、17:00 上樓就寢。)
func (s *NPCSchedule) Slot(hour int) int {
	var d [4]byte
	for i := range d {
		d[i] = byte(hour) - s.Times[i]
	}
	best, slot := d[0], 0
	if d[1] < best {
		best, slot = d[1], 1
	}
	if d[2] < best {
		best, slot = d[2], 2
	}
	if best > d[3] {
		slot = 1
	}
	return slot
}

// NPC 是一個 NPC 槽。Creature 為 0 代表這個槽是空的。
type NPC struct {
	Schedule NPCSchedule
	// Creature 是生物編號 —— **不是** tile 索引。要畫圖請用 TileIndex()。
	Creature byte
	Dialogue byte // 對話號碼,語意見上面的常數
}

// NPCTileBase 是生物編號要加上的偏移。
//
// 原版的算繪路徑就是 `movzx eax, al; add eax, 100h` 之後才呼叫畫 tile 的函式 ——
// 生物編號存成 byte,而人物圖都在 tileset 的後半(256 之後)。
//
// 兩個獨立的佐證:所有生物編號都是 4 的倍數,而人物圖正好以 4 張(四個朝向)
// 為一組;衛兵的編號 0x70 → tile 368,算繪出來是持戟的鎧甲士兵,
// 恰好對應 sub_1B52C 那句 "The guard offers no response!"。
const NPCTileBase = 256

// TileIndex 回傳算繪用的 tile 索引。
func (n *NPC) TileIndex() int { return NPCTileBase + int(n.Creature) }

// Present 回報這個槽有沒有人。
func (n *NPC) Present() bool { return n.Creature != 0 }

// IsShopkeeper 回報這是不是商人(對話走 SHOPPE.DAT 而非 .TLK)。
func (n *NPC) IsShopkeeper() bool {
	return n.Dialogue >= DialogueShopFirst && n.Dialogue <= DialogueShopLast
}

// At 回報 hour 這個時刻 NPC 在哪裡。
func (n *NPC) At(hour int) (x, y, floor int) {
	s := n.Schedule.Slot(hour)
	return int(n.Schedule.X[s]), int(n.Schedule.Y[s]), int(n.Schedule.Floor[s])
}

// NPCSet 是一個 .NPC 檔載好之後的內容(8 個地點 × 32 個槽)。
type NPCSet struct {
	Files [len(NPCFiles)][8][NPCsPerLocation]NPC
}

// ParseNPCBlock 解一個地點的 576 B。
func ParseNPCBlock(raw []byte) ([NPCsPerLocation]NPC, error) {
	var out [NPCsPerLocation]NPC
	if len(raw) != NPCBlockSize {
		return out, fmt.Errorf("NPC 區塊 %d B,預期 %d B", len(raw), NPCBlockSize)
	}
	creatures := raw[NPCsPerLocation*NPCRecordSize:]
	dialogues := creatures[NPCsPerLocation:]
	for i := range out {
		r := raw[i*NPCRecordSize : (i+1)*NPCRecordSize]
		copy(out[i].Schedule.AI[:], r[0:3])
		copy(out[i].Schedule.X[:], r[3:6])
		copy(out[i].Schedule.Y[:], r[6:9])
		copy(out[i].Schedule.Floor[:], r[9:12])
		copy(out[i].Schedule.Times[:], r[12:16])
		out[i].Creature = creatures[i]
		out[i].Dialogue = dialogues[i]
	}
	return out, nil
}

// LoadNPCSet 讀入四個 .NPC 檔。
func LoadNPCSet(dir string) (*NPCSet, error) {
	s := &NPCSet{}
	for fi, name := range NPCFiles {
		raw, err := readExact(filepath.Join(dir, name), NPCFileSize)
		if err != nil {
			return nil, err
		}
		for li := 0; li < 8; li++ {
			block, err := ParseNPCBlock(raw[li*NPCBlockSize : (li+1)*NPCBlockSize])
			if err != nil {
				return nil, fmt.Errorf("%s 第 %d 個地點:%w", name, li, err)
			}
			s.Files[fi][li] = block
		}
	}
	return s, nil
}

// At 回傳地點 num(1-based)的 32 個 NPC 槽。
func (s *NPCSet) At(num int) (*[NPCsPerLocation]NPC, error) {
	loc, err := LocationByNumber(num)
	if err != nil {
		return nil, err
	}
	return &s.Files[loc.SceneFile][(num-1)%8], nil
}
