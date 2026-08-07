package u5data

import (
	"fmt"
	"os"
	"path/filepath"
)

// 戰鬥數值
//
// 四張表,全部在 DOS `DATA.OVL` 裡(位址是拿 FM Towns 的資料段逐位元組比對出來的):
//
//	FM Towns   DOS       內容
//	0x3F050    0x13CC    怪物屬性  48 筆 × 8 B
//	0x3F290    0x160C    武器傷害   48 B
//	0x3F2C8    0x1644    裝備防禦值 48 B
//	0x3F2F8    0x1674    武器射程   48 B
//	0x3F330    0x16AC    裝備類別   48 B
const (
	statsCreature = 0x13CC // 48 × 8
	statsItemDmg  = 0x160C // 48
	statsItemDef  = 0x1644 // 48
	statsItemRnge = 0x1674 // 48
	statsItemKind = 0x16AC // 48
)

// CreatureStatSize 是怪物屬性一筆的大小。
const CreatureStatSize = 8

// 怪物屬性的欄位。
//
// 前三個一眼可辨,而且與角色的三圍一一對應:
//
//	Mage    [10, 15, 20, …]   力量低、智力高
//	Fighter [20, 15, 10, …]   反過來
//	Bard    [15, 20, 10, …]   敏捷最高
//	Avatar  [25, 25, 25, …]   全能
//
// `sub_B398` 也印證了這件事:它取 `byte_3F050[生物*8]` 當「力量那一項」、
// `byte_3F052[生物*8]` 當「智力那一項」,而角色走的是紀錄裡的 0x0C 與 0x0E
// —— 正好是力量與智力。
//
// +3 與 +4 由傷害公式(`sub_B274`)認出來:護甲與攻擊力。
// +5..+7 還沒對出語意,原樣留在 Raw 裡。
const (
	CreatureStrength = 0
	CreatureDex      = 1
	CreatureIntel    = 2
	// CreatureArmour 是減傷用的護甲(`sub_B274` 讀 `byte_3F053[生物*8]`)。
	// 戰士 8、法師 0 —— 對得上定位。
	CreatureArmour = 3
	// CreatureAttack 是傷害用的攻擊力(`sub_B274` 讀 `byte_3F054[生物*8]`)。
	// 聖者 30 最高。
	CreatureAttack = 4
)

// CreatureStats 是一種怪物的戰鬥數值。
type CreatureStats struct {
	Strength, Dex, Intel byte
	// Armour 是被打時的減傷,Attack 是打人時的傷害上限。
	Armour, Attack byte
	Raw            [CreatureStatSize]byte
}

// CombatStats 是全部四張表。
type CombatStats struct {
	// Creature[索引] —— 索引與生物名表同一套((編號 − 64) / 4)。
	Creature [CreatureCount]CreatureStats
	// ItemDefence[裝備編號] 是這件裝備加多少防禦。
	//
	// 值本身就說明了語意:頭盔 1/2/3/3、盾 2/3/3/5、護甲 1→10 遞增、
	// 武器幾乎都是 0(只有 Main Gauche 是 1)、防護戒指 2、釘項圈 2。
	ItemDefence [ItemCount]int
	// ItemRange[裝備編號] 是遠程武器打得到幾格;0 = 近戰。
	//
	// 匕首 3、投石索 4、火油 4、矛 5、投擲斧 4、弓 7、十字弓 8、
	// 魔法弓 15、魔法斧 15 —— 正好是 U5 能扔能射的那幾樣。
	ItemRange [ItemCount]int
	// ItemDamage[裝備編號] 是武器的傷害上限。
	//
	// 匕首 6 → 長劍 15 → 雙手武器 20 → 戟 30;
	// **混沌之劍與玻璃劍都是 99**(必殺),而箭矢與弩矢只有 1
	// —— 彈藥本身不造成傷害,傷害來自弓。
	ItemDamage [ItemCount]int
	// ItemKind[裝備編號] 是武器類別。
	//
	// 已知的只有一個值:**8 = 鈍器**(釘盔、釘盾、棍、釘頭錘、雙手錘)。
	// `sub_B398` 用它決定命中要看力量還是看另一項:類別 8 走力量。
	// 其餘的值(2、3、7)還沒對出語意。
	ItemKind [ItemCount]int
}

// ItemKindBlunt 是「鈍器」的類別碼(`sub_B398` 的 `cmp byte_3F330[edi], 8`)。
const ItemKindBlunt = 8

// LoadCombatStats 從 DATA.OVL 讀出四張表。
func LoadCombatStats(dir string) (*CombatStats, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "DATA.OVL"))
	if err != nil {
		return nil, err
	}
	return ParseCombatStats(raw)
}

// ParseCombatStats 從 DATA.OVL 的內容取出四張表。
func ParseCombatStats(ovl []byte) (*CombatStats, error) {
	need := statsItemKind + ItemCount
	if len(ovl) < need {
		return nil, fmt.Errorf("DATA.OVL 只有 %d B,放不下 0x%X 的戰鬥數值區", len(ovl), need)
	}
	s := &CombatStats{}
	for i := 0; i < CreatureCount; i++ {
		off := statsCreature + i*CreatureStatSize
		c := &s.Creature[i]
		copy(c.Raw[:], ovl[off:off+CreatureStatSize])
		c.Strength = c.Raw[CreatureStrength]
		c.Dex = c.Raw[CreatureDex]
		c.Intel = c.Raw[CreatureIntel]
		c.Armour = c.Raw[CreatureArmour]
		c.Attack = c.Raw[CreatureAttack]
	}
	for i := 0; i < ItemCount; i++ {
		s.ItemDamage[i] = int(ovl[statsItemDmg+i])
		s.ItemDefence[i] = int(ovl[statsItemDef+i])
		s.ItemRange[i] = int(ovl[statsItemRnge+i])
		s.ItemKind[i] = int(ovl[statsItemKind+i])
	}
	if err := s.validate(); err != nil {
		return nil, err
	}
	return s, nil
}

// validate 用「算錯就一定違反」的性質擋住位移偏掉。
func (s *CombatStats) validate() error {
	// 法師的智力該高於力量,戰士反過來 —— 這兩條同時成立的機率不高,
	// 位移偏一格就會兩邊都不對。
	mage, fighter := &s.Creature[0], &s.Creature[2]
	if mage.Intel <= mage.Strength {
		return fmt.Errorf("法師的智力 %d 不高於力量 %d —— 怪物屬性表的位移大概錯了",
			mage.Intel, mage.Strength)
	}
	if fighter.Strength <= fighter.Intel {
		return fmt.Errorf("戰士的力量 %d 不高於智力 %d", fighter.Strength, fighter.Intel)
	}
	// 護甲的防禦值必須嚴格遞增(布甲 → 板甲)。
	for i := ItemArmourFirst; i < ItemArmourLast-1; i++ {
		if s.ItemDefence[i] >= s.ItemDefence[i+1] {
			return fmt.Errorf("護甲 %d 的防禦 %d 不低於 %d 的 %d",
				i, s.ItemDefence[i], i+1, s.ItemDefence[i+1])
		}
	}
	// 近戰武器不該有射程,弓與十字弓一定有。
	for _, id := range []int{ItemArrows, ItemQuarrels} {
		if s.ItemRange[id] != 0 {
			return fmt.Errorf("彈藥 %d 不該有射程,卻是 %d", id, s.ItemRange[id])
		}
	}
	if s.ItemRange[26] == 0 || s.ItemRange[28] == 0 {
		return fmt.Errorf("弓 / 十字弓沒有射程(%d / %d)", s.ItemRange[26], s.ItemRange[28])
	}
	// 傷害:防具**除了鈍器**之外不該有傷害值。
	//
	// ⚠ 釘盔(3)傷害 4、釘盾(6)也有 —— 它們的類別是 8(鈍器),
	// 拿來撞人本來就打得痛。我第一版把「所有防具傷害 0」寫進 validate,
	// 結果 LoadCombatStats 直接回錯,連帶讓測試裡的 Stats 變成 nil、
	// 減傷全部失效。斷言寫太死比不寫更危險。
	for i := ItemHelmFirst; i <= ItemArmourLast; i++ {
		if s.ItemDamage[i] != 0 && s.ItemKind[i] != ItemKindBlunt {
			return fmt.Errorf("非鈍器的防具 %d 傷害是 %d,不該有", i, s.ItemDamage[i])
		}
	}
	for _, id := range []int{ItemSwordOfChaos, ItemGlassSword} {
		if s.ItemDamage[id] != InstantKillDamage {
			return fmt.Errorf("裝備 %d 的傷害是 %d,預期必殺 %d",
				id, s.ItemDamage[id], InstantKillDamage)
		}
	}
	return nil
}

// StatsFor 依生物編號取屬性(編號 → 索引的公式與生物名表同一套)。
func (s *CombatStats) StatsFor(creature byte) (*CreatureStats, bool) {
	if s == nil || creature < CreatureBase || (creature-CreatureBase)%4 != 0 {
		return nil, false
	}
	i := int(creature-CreatureBase) / 4
	if i >= CreatureCount {
		return nil, false
	}
	return &s.Creature[i], true
}

// IsRanged 回報這件武器打不打得到遠處。
func (s *CombatStats) IsRanged(item byte) bool {
	if s == nil || int(item) >= ItemCount {
		return false
	}
	return s.ItemRange[item] > 0
}

// IsBlunt 回報這件武器算不算鈍器(命中靠力量那一類)。
func (s *CombatStats) IsBlunt(item byte) bool {
	if s == nil || int(item) >= ItemCount {
		return false
	}
	return s.ItemKind[item] == ItemKindBlunt
}

// DefenceOf 把一名角色身上的裝備防禦值加總(原版 `sub_2F2BC`)。
//
// 逐格看六個裝備欄位,`!= ItemNone` 的就把 `ItemDefence` 加進去。
func (s *CombatStats) DefenceOf(c *Character) int {
	if s == nil || c == nil {
		return 0
	}
	total := 0
	for _, id := range c.Equipment().Slots() {
		if id == ItemNone || int(id) >= ItemCount {
			continue
		}
		total += s.ItemDefence[id]
	}
	return total
}

// 傷害公式裡有特例的三把武器(`sub_B274` 的三個 cmp)。
//
// 它們同時也是 `sub_B484` 的**必中**清單 —— 三把神器不用擲命中骰。
const (
	// ItemSwordOfChaos 混沌之劍:傷害 99。
	ItemSwordOfChaos = 0x23 // 35
	// ItemGlassSword 玻璃劍:傷害 99,而且**用一次就碎**
	// (原版印「Thy sword hath shattered!」並消耗掉)。
	ItemGlassSword = 0x27 // 39
	// ItemJeweledSword 寶石劍:傷害被特例設成 **0**,不是表裡的 1。
	ItemJeweledSword = 0x28 // 40
	// InstantKillDamage 是「必殺」的傷害值;走到這個值就不再扣防禦。
	InstantKillDamage = 99
	// BareHandDamage 是空手的傷害(武器欄位是 ItemNone 時)。
	BareHandDamage = 1
)
