package u5data

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

// 裝備表
//
// `DATA.OVL` 裡的裝備名字**分成兩段**放:0x0052 起是舊的一批、0x175C 起是另一批
// (Chain Coif、Iron Helm、Ring Mail、Dagger、Sling…)。單獨讀任何一段都會缺一半,
// 而且缺的是最常見的那些(長劍、十字弓、鎖甲頭巾)。
//
// 真正的順序在 **0x1806 的指標表**:48 個 u16,依裝備編號排好,交錯指向兩段字串。
//
// ⚠ 指標要 **加 0x10** 才是檔案位移(`0x0042 + 0x10 = 0x0052` = "Leather Helm"、
// `0x174C + 0x10 = 0x175C` = "Chain Coif")。這 0x10 是 DOS overlay 的載入偏移;
// 忘了加會讀到字串中間,解出 " Helm"、"elm" 這種殘片 —— 看起來像編碼問題,
// 其實是位移問題。
const (
	// ItemPointerTable 是裝備名指標表在 DATA.OVL 裡的位移。
	ItemPointerTable = 0x1806
	// ItemPointerBias 是指標要加上的偏移。
	ItemPointerBias = 0x10
	// ItemCount 是裝備的數量(編號 0..47)。48 之後接的是生物名,不是裝備。
	ItemCount = 48
	// ItemNone 是「這個欄位沒有裝備」。
	ItemNone = 0xFF
)

// 裝備分類的邊界(依名字表的實際排列)。
const (
	ItemHelmFirst   = 0
	ItemHelmLast    = 3
	ItemShieldFirst = 4
	ItemShieldLast  = 8
	ItemArmourFirst = 9
	ItemArmourLast  = 15
	ItemWeaponFirst = 16
	ItemWeaponLast  = 41
	ItemRingFirst   = 42
	ItemRingLast    = 47
)

// ItemTable 是 48 件裝備的名字,索引就是裝備編號。
type ItemTable struct {
	Names [ItemCount]string
}

// LoadItemTable 從 DATA.OVL 讀出裝備表。
func LoadItemTable(dir string) (*ItemTable, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "DATA.OVL"))
	if err != nil {
		return nil, err
	}
	return ParseItemTable(raw)
}

// ParseItemTable 從 DATA.OVL 的內容取出裝備表。
func ParseItemTable(ovl []byte) (*ItemTable, error) {
	need := ItemPointerTable + ItemCount*2
	if len(ovl) < need {
		return nil, fmt.Errorf("DATA.OVL 只有 %d B,放不下 0x%X 的指標表", len(ovl), ItemPointerTable)
	}
	t := &ItemTable{}
	for i := 0; i < ItemCount; i++ {
		off := int(binary.LittleEndian.Uint16(ovl[ItemPointerTable+i*2:])) + ItemPointerBias
		e := indexByte(ovl, off)
		if off <= 0 || off >= len(ovl) || e < 0 {
			return nil, fmt.Errorf("第 %d 件裝備的指標指到 0x%X,超出檔案", i, off)
		}
		name := string(ovl[off:e])
		if name == "" || !printableASCII(name) {
			return nil, fmt.Errorf("第 %d 件裝備的名字不像名字:%q", i, name)
		}
		t.Names[i] = name
	}
	// 自我一致性:頭尾對不上就是位移或偏移算錯了 —— 早點失敗比默默解出半截名字好。
	if t.Names[0] != "Leather Helm" || t.Names[ItemCount-1] != "Ankh" {
		return nil, fmt.Errorf("裝備表頭尾是 %q / %q,預期 \"Leather Helm\" / \"Ankh\"",
			t.Names[0], t.Names[ItemCount-1])
	}
	return t, nil
}

// Name 回傳裝備編號的名字;ItemNone 或超出範圍回傳空字串。
func (t *ItemTable) Name(id byte) string {
	if t == nil || id == ItemNone || int(id) >= ItemCount {
		return ""
	}
	return t.Names[id]
}

// 角色紀錄裡的裝備欄位。
//
// 由六名初始角色橫向對照定出來,而且每個人的配備都符合他的定位:
// 法師 Mariah 與 Jaana 沒有頭盔、穿布甲拿匕首;戰士 Geoffrey 是釘盔 + 鱗甲 +
// 釘頭錘 + 釘盾;吟遊詩人 Iolo 雙手各一把武器;聖者是鎖甲頭巾 + 鎖甲 + 長劍 + 安卡。
const (
	CharHelm   = 0x19 // 25
	CharArmour = 0x1A // 26
	CharWeapon = 0x1B // 27 右手
	CharShield = 0x1C // 28 左手(盾或第二把武器)
	CharRing   = 0x1D // 29
	CharAmulet = 0x1E // 30
	// CharInnFlag 是「留在旅店」的標記。入隊時原版把它設成 0
	// (sub_1BB5C 寫 byte_3DDD3[i*32],而 0x3DDD3 - 0x3DDB4 = 0x1F)。
	CharInnFlag = 0x1F // 31
	// CharInnDays 是留在旅店的天數:每過一天 +1,上限 25;0xFF 代表沒在計數。
	// (日期進位那段 sub_29304 對 16 名角色跑 `byte_3DDCB[i*32]`,
	//  而 0x3DDCB - 0x3DDB4 = 0x17。)
	CharInnDays = 0x17 // 23
)

// Equipment 是一名角色身上的裝備編號。
type Equipment struct {
	Helm, Armour, Weapon, Shield, Ring, Amulet byte
}

// Equipment 從角色紀錄取出裝備欄位。
func (c *Character) Equipment() Equipment {
	return Equipment{
		Helm:   c.Raw[CharHelm],
		Armour: c.Raw[CharArmour],
		Weapon: c.Raw[CharWeapon],
		Shield: c.Raw[CharShield],
		Ring:   c.Raw[CharRing],
		Amulet: c.Raw[CharAmulet],
	}
}

// Slots 依「頭盔 / 護甲 / 武器 / 副手 / 戒指 / 護符」的順序回傳裝備編號。
func (e Equipment) Slots() [6]byte {
	return [6]byte{e.Helm, e.Armour, e.Weapon, e.Shield, e.Ring, e.Amulet}
}

// EquipmentSlotNames 是裝備欄位的中文名,順序與 Equipment.Slots 相同。
var EquipmentSlotNames = [6]string{"頭盔", "護甲", "武器", "副手", "戒指", "護符"}

// 生物名表
//
// 緊接在裝備指標表後面(0x1866)的是另一張 48 個 u16 的指標表,指向生物名:
// Mage / Bard / Fighter / Avatar / Villager / Merchant / … / Guard / Blackthorn /
// Lord British / Sea Horse / … / Shadow Lord。
//
// 索引 = **(生物編號 − 64) / 4**。人物圖以四張(四個朝向)為一組,所以編號是 4 的倍數;
// 減 64 是因為表從編號 64 開始。
//
//	64 Mage   68 Bard   80 Villager   84 Merchant   104 Child   112 Guard
//
// 這個公式與 `sub_B98` 的「可被嚇跑的平民」範圍 0x40..0x73 互相印證:
// 那個範圍換算成索引正好是 0..11,**剛好排除**衛兵(12)、遊民(13)、
// 黑刺(14)與不列顛王(15)—— 這幾種本來就不該被嚇跑。
//
// ⚠ 不是每個 NPC 槽都對得上:全遊戲 325 個 NPC 裡有 29 個編號 < 64
// (物件與動物,tile 落在 256–319 那一列)或不是 4 的倍數(兩隻怪物),
// 它們畫得出來但這張表裡沒有名字。這是原版資料就這樣,不是解析漏掉。
const (
	// CreaturePointerTable 是生物名指標表在 DATA.OVL 裡的位移。
	CreaturePointerTable = ItemPointerTable + ItemCount*2 // 0x1866
	// CreatureCount 是表裡的筆數。
	CreatureCount = 48
	// CreatureBase 是生物編號的起點(索引 0 對應的編號)。
	CreatureBase = 64
	// CreatureHumanLast 是最後一個「人」的索引;之後是怪物。
	CreatureHumanLast = 15
)

// CreatureTable 是生物名字,索引 = (生物編號 − 64) / 4。
type CreatureTable struct {
	Names [CreatureCount]string
}

// ParseCreatureTable 從 DATA.OVL 的內容取出生物名表。
func ParseCreatureTable(ovl []byte) (*CreatureTable, error) {
	need := CreaturePointerTable + CreatureCount*2
	if len(ovl) < need {
		return nil, fmt.Errorf("DATA.OVL 只有 %d B,放不下 0x%X 的指標表", len(ovl), CreaturePointerTable)
	}
	t := &CreatureTable{}
	for i := 0; i < CreatureCount; i++ {
		off := int(binary.LittleEndian.Uint16(ovl[CreaturePointerTable+i*2:])) + ItemPointerBias
		if off <= ItemPointerBias || off >= len(ovl) {
			continue // 表裡有兩個空指標(索引 8、9),照原樣留空
		}
		e := indexByte(ovl, off)
		if e < 0 {
			continue
		}
		if name := string(ovl[off:e]); printableASCII(name) {
			t.Names[i] = name
		}
	}
	if t.Names[0] != "Mage" || t.Names[CreatureHumanLast] != "Lord British" {
		return nil, fmt.Errorf("生物名表的第 0 / %d 筆是 %q / %q,預期 \"Mage\" / \"Lord British\"",
			CreatureHumanLast, t.Names[0], t.Names[CreatureHumanLast])
	}
	return t, nil
}

// LoadCreatureTable 從 DATA.OVL 讀出生物名表。
func LoadCreatureTable(dir string) (*CreatureTable, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "DATA.OVL"))
	if err != nil {
		return nil, err
	}
	return ParseCreatureTable(raw)
}

// Name 依生物編號回傳名字。編號 < 64 或不是 4 的倍數時回傳空字串 ——
// 那些是物件與動物,這張表裡沒有它們。
func (t *CreatureTable) Name(creature byte) string {
	if t == nil || creature < CreatureBase || (creature-CreatureBase)%4 != 0 {
		return ""
	}
	i := int(creature-CreatureBase) / 4
	if i >= CreatureCount {
		return ""
	}
	return t.Names[i]
}

// IsHuman 回報這個生物編號算不算「人」。
//
// 依據是 `sub_B98`(平民被攻擊會嚇到):它判的範圍是 0x40 <= 編號 < 0x74,
// 換算成索引 0..11 —— 衛兵以後的不算。
func (t *CreatureTable) IsHuman(creature byte) bool {
	return creature >= 0x40 && creature < 0x74
}
