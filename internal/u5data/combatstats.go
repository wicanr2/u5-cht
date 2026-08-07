package u5data

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

// 戰鬥數值
//
// 八張表,全部在 DOS `DATA.OVL` 裡(位址是拿 FM Towns 的資料段逐位元組比對出來的)。
// 前四張緊緊相鄰,後四張間距 0x38(48 件 + 8 B 空白):
//
//	FM Towns   DOS       內容
//	0x3F050    0x13CC    怪物屬性   48 筆 × 8 B
//	0x3F1D0    0x154C    怪物旗標   48 × u16
//	0x3F230    0x15AC    怪物射程   48 B
//	0x3F260    0x15DC    投射物圖號 48 B
//	0x3F290    0x160C    武器傷害   **56 B**
//	0x3F2C8    0x1644    裝備防禦值 **56 B**
//	0x3F2F8    0x1674    武器射程   **56 B**
//	0x3F330    0x16AC    裝備類別   **56 B**
//	0x3F368    0x16E4    混編生物   48 B
//
// ⚠ **四張裝備表是 56 筆不是 48**(2026-08-07 更正)。相鄰兩張表的位移差
// 是 0x38 = 56,不是 0x30 = 48 —— 而多出來的 8 筆不是留白,是**攻擊咒語**:
//
//	48 (0x30) Grav Por     傷害 16  射程 15
//	49 (0x31) Vas Flam     傷害 30  射程 15
//	50 (0x32) Xen Corp     傷害 99  射程 15   ← 必殺
//	51 (0x33) In Nox Grav  傷害 18  射程 15   ← 實際效果是中毒
//	52 (0x34) In Zu Grav   傷害  0  射程 15   ← 實際效果是睡眠
//	53 (0x35) In Flam Grav 傷害 21  射程 15
//	54 (0x36) In Sanct Grav傷害  0  射程 15
//	55        (空)
//
// 怎麼發現的:追戰鬥中的力場咒語時,`sub_20360` 對隊員取
// `byte_3F2F8[效果碼]` 與 `byte_3F330[效果碼]` —— 而效果碼是 0x30..0x36,
// 早就超出 48。**「這個索引超出表的範圍」不是 bug,是表比我以為的長。**
// 相鄰位移一減就看得出來,只是先前照「48 種裝備」的印象填了 48。
//
// ⇒ 這同時把 `docs/re/17` 裡「Grav Por 與 Vas Flam 的傷害是估計值」那條解決掉:
// 16 與 30 都在表上,不用估。
const (
	statsCreature = 0x13CC // 48 × 8
	statsCreFlags = 0x154C // 48 × u16
	statsCreRange = 0x15AC // 48
	statsCreMisl  = 0x15DC // 48
	statsItemDmg  = 0x160C // 56
	statsItemDef  = 0x1644 // 56
	statsItemRnge = 0x1674 // 56
	statsItemKind = 0x16AC // 56
	statsCreMix   = 0x16E4 // 48
)

// AttackCodeCount 是四張裝備表的實際長度:48 件裝備 + 8 個攻擊咒語欄位。
const AttackCodeCount = 56

// 攻擊咒語在那四張表裡的索引(= 原版寫進 `byte_3E0AD` 的攻擊碼)。
const (
	AttackGravPor     = 0x30
	AttackVasFlam     = 0x31
	AttackXenCorp     = 0x32
	AttackInNoxGrav   = 0x33 // 命中後中毒(`sub_B9A8` 的 `cmp byte_3E0AD, '3'`)
	AttackInZuGrav    = 0x34 // 命中後睡著(同上,`'4'`)
	AttackInFlamGrav  = 0x35
	AttackInSanctGrav = 0x36
)

// AttackAlwaysHits 回報這個攻擊碼會不會自動命中(`sub_B484` 的施法分支)。
//
//	0x30 / 0x31          自動命中
//	0x32(Xen Corp)      **要擲** —— 用智力對智力,與抗性判定同一個形狀
//	>= 0x33              自動命中
//
// ⚠ 唯一會失手的攻擊咒語是必殺的那一個。寫成「攻擊咒語都自動命中」
// 會讓 Xen Corp 變成無條件秒殺。
func AttackAlwaysHits(code int) bool {
	return (code > 0x2F && code < 0x32) || code >= 0x33
}

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
// +5 與 +6 由死亡處理(`sub_B51C`)與遭遇生成(`sub_2F0EC`)認出來。
// +7 還沒對出語意,原樣留在 Raw 裡。
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
	// CreatureMaxHP 是生命上限(`sub_B51C` 死亡時用 `byte_3F055[生物*8] / 4 + 1`
	// 算經驗值)。蝙蝠 5、史萊姆 10、巨龍 99、精靈守衛 99。
	CreatureMaxHP = 5
	// CreatureGroupMax 是一次遭遇最多幾隻(`sub_2F0EC` 的 `byte_3F056[生物*8]`)。
	// 蝙蝠與史萊姆 16、蟲群與鯊魚與獸人 10、衛兵 8、巨龍 2、
	// 巨蟒陷阱與石像鬼與擬態怪 1。
	CreatureGroupMax = 6
)

// 怪物旗標(`word_3F1D0[生物]`,48 × u16)。
//
// 每一條都是從實際讀那一位元的程式碼認出來的,不是從名字猜的;
// 認不出來的就照實留白(見 `docs/re/16`)。
const (
	// CreatureNoRemains:死了不留屍體也不留寶箱(`sub_B51C` 的 `test …, 0x1001`)。
	// 海洋生物、蝙蝠、蟲群、幽靈這些都有。
	CreatureNoRemains = 0x0001
	// CreatureStealsFood:命中時有 3/4 機率偷食物而不造成傷害
	//(`sub_9F08` 的 `test al, 2`)。全表只有小魔怪(Gremlin)。
	CreatureStealsFood = 0x0002
	// CreaturePoison1 / CreaturePoison9:任一位元成立就走下毒攻擊
	//(`sub_B9A8` 的 `test …, 0x204` → `sub_B8DC`)。
	// 巨蟒與大烏賊是 0x0004,巨鼠與擬態怪是 0x0200,巨蜘蛛兩個都有。
	CreaturePoison1 = 0x0004
	CreaturePoison9 = 0x0200
	// CreatureInvulnerable:傷害直接歸零(`sub_B51C` 的 `test al, 8`)。
	// 旅人、黑刺、不列顛王 —— 這三個本來就殺不死。
	CreatureInvulnerable = 0x0008
	// CreatureDivides:被打到但沒死會分裂(`sub_B51C` 的 `test al, 0x10`,
	// 最多試 8 個空位,印「X divides!」)。全表只有史萊姆。
	CreatureDivides = 0x0010
	// CreatureHalfDamage:傷害減半,除非 `byte_3E0A0`(疑為魔法武器旗標)成立
	//(`sub_B51C` 的 `test al, 0x20`)。
	//
	// ★ 這個位元同時是 **An Xen Corp(驅離)的唯一閘門** ——
	// `sub_18EB0` 掃全場時只處理帶這一位元的目標。實測全 48 種帶它的是
	// **幽靈、骷髏、惡魔、暗影領主**四種。
	//
	// ⚠ 別把它讀成「不死生物」:惡魔不是不死,而屍鬼(Rot Worm)也沒帶。
	// 原版沒有另一張不死表 —— 減傷與可驅離就是同一個位元,兩種語意。
	CreatureHalfDamage = 0x0020

	// CreatureRepellable 是 CreatureHalfDamage 的另一個名字,
	// 用在講 An Xen Corp 驅不驅得動的地方。同一位元,換個說法比較讀得懂。
	CreatureRepellable = CreatureHalfDamage

	// CreatureCharms:遠程回合可能對隨機隊員施展魅惑(`sub_1F5A4` 的 `test al, 0x40`)。
	// 注視者(Gazer)是代表。
	CreatureCharms = 0x0040
	// CreatureVanishes:死時不留屍體而是印「X vanishes!」並換成 tile 0x16
	//(`sub_B51C` 的 `test 高位元組, 0x10`)。旅人、黑刺、不列顛王、暗影領主。
	CreatureVanishes = 0x1000
	// CreatureTeleports:移動時可能瞬間移動到別處(`sub_AE20` 的 `test 高位元組, 0x20`)。
	// 鬼火(Wisp)是代表。
	CreatureTeleports = 0x2000
	// CreatureCasts:走施法那條遠程路徑(`sub_9E10` / `sub_9F08` 的 `test ah, 0x80`)。
	// 法師、注視者、收割者、惡魔、海馬、黑刺、不列顛王、暗影領主。
	CreatureCasts = 0x8000
)

// ItemAmuletOfTurning 是「轉化護符」。
//
// 戴在護符欄(紀錄 0x1E)時,施法者的遠程攻擊有 **1/2 機率被抵銷**
// (`sub_9F08` 先看目標的 `byte_3DDD2[角色*32] == 0x2D`,
// 再看攻擊者有沒有 CreatureCasts,兩個都成立才擲那一枚硬幣)。
const ItemAmuletOfTurning = 0x2D

// CreatureStats 是一種怪物的戰鬥數值。
type CreatureStats struct {
	Strength, Dex, Intel byte
	// Armour 是被打時的減傷,Attack 是打人時的傷害上限。
	Armour, Attack byte
	// MaxHP 是生命上限,GroupMax 是一次遭遇的隻數上限。
	MaxHP, GroupMax byte
	// Flags 是旗標位元(見上面那組 Creature* 常數)。
	Flags uint16
	// Range 是這種怪物打得到幾格(近戰是 1、衛兵 15、巨龍 9、法師 7)。
	Range byte
	Raw   [CreatureStatSize]byte
}

// Has 回報這種怪物有沒有某個旗標。
func (c *CreatureStats) Has(flag uint16) bool { return c != nil && c.Flags&flag != 0 }

// IsPoisonous 回報這種怪物的攻擊會不會下毒(`sub_B9A8` 的 0x204 是**兩個位元任一**)。
func (c *CreatureStats) IsPoisonous() bool {
	return c.Has(CreaturePoison1) || c.Has(CreaturePoison9)
}

// CombatStats 是全部四張表。
type CombatStats struct {
	// Creature[索引] —— 索引與生物名表同一套((編號 − 64) / 4)。
	Creature [CreatureCount]CreatureStats
	// ItemDefence[裝備編號] 是這件裝備加多少防禦。
	//
	// 值本身就說明了語意:頭盔 1/2/3/3、盾 2/3/3/5、護甲 1→10 遞增、
	// 武器幾乎都是 0(只有 Main Gauche 是 1)、防護戒指 2、釘項圈 2。
	ItemDefence [AttackCodeCount]int
	// ItemRange[裝備編號] 是遠程武器打得到幾格;0 = 近戰。
	//
	// 匕首 3、投石索 4、火油 4、矛 5、投擲斧 4、弓 7、十字弓 8、
	// 魔法弓 15、魔法斧 15 —— 正好是 U5 能扔能射的那幾樣。
	ItemRange [AttackCodeCount]int
	// ItemDamage[裝備編號] 是武器的傷害上限。
	//
	// 匕首 6 → 長劍 15 → 雙手武器 20 → 戟 30;
	// **混沌之劍與玻璃劍都是 99**(必殺),而箭矢與弩矢只有 1
	// —— 彈藥本身不造成傷害,傷害來自弓。
	ItemDamage [AttackCodeCount]int
	// ItemKind[裝備編號] 是武器類別。
	//
	// 已知的只有一個值:**8 = 鈍器**(釘盔、釘盾、棍、釘頭錘、雙手錘)。
	// `sub_B398` 用它決定命中要看力量還是看另一項:類別 8 走力量。
	// 其餘的值(2、3、7)還沒對出語意。
	ItemKind [AttackCodeCount]int
	// CreatureMix[索引] 是遭遇時可能混進來的另一種怪物(`sub_2F0EC` 的
	// `byte_3F368[生物]`)。前四分之一的同伴各有 1/9 機率換成這一種。
	CreatureMix [CreatureCount]byte
	// CreatureMissile[索引] 是遠程攻擊飛出去的那個東西的圖號。
	//
	// 依據:`sub_9E10` 一開頭就 `movzx eax, byte_3F260[生物]`,原封不動
	// 當成 `sub_1FE54`(投射物飛行)的最後一個參數。近戰生物那一批全是 0,
	// 有射程的那一批全非 0 —— 與射程表互為佐證。
	CreatureMissile [CreatureCount]byte
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

// ParseCombatStats 從 DATA.OVL 的內容取出全部八張表。
func ParseCombatStats(ovl []byte) (*CombatStats, error) {
	need := statsCreMix + CreatureCount
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
		c.MaxHP = c.Raw[CreatureMaxHP]
		c.GroupMax = c.Raw[CreatureGroupMax]
		c.Flags = binary.LittleEndian.Uint16(ovl[statsCreFlags+i*2:])
		c.Range = ovl[statsCreRange+i]
		s.CreatureMix[i] = ovl[statsCreMix+i]
		s.CreatureMissile[i] = ovl[statsCreMisl+i]
	}
	for i := 0; i < AttackCodeCount; i++ {
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
	// 旗標表:殺不死的那三位(旅人 / 黑刺 / 不列顛王)必須同時是無敵與消失。
	// 六個位元一起對上,位移偏一格不可能還成立。
	for _, i := range []int{CreatureWanderer, CreatureBlackthorn, CreatureLordBritish} {
		c := &s.Creature[i]
		if !c.Has(CreatureInvulnerable) || !c.Has(CreatureVanishes) {
			return fmt.Errorf("生物 %d 的旗標是 %04X,預期同時有無敵(%04X)與消失(%04X)",
				i, c.Flags, CreatureInvulnerable, CreatureVanishes)
		}
	}
	// 射程表:近戰的一定是 1,衛兵(拿十字弓)是全表最大。
	if r := s.Creature[CreatureFighterIdx].Range; r != 1 {
		return fmt.Errorf("戰士的射程是 %d,近戰該是 1", r)
	}
	if g, f := s.Creature[CreatureGuardIdx].Range, s.Creature[CreatureFighterIdx].Range; g <= f {
		return fmt.Errorf("衛兵射程 %d 不大於戰士 %d —— 射程表的位移大概錯了", g, f)
	}
	return nil
}

// 幾個在驗證與規則裡點名到的生物索引(索引 = (編號 − 64) / 4)。
const (
	CreatureFighterIdx  = 2
	CreatureGuardIdx    = 12
	CreatureWanderer    = 13
	CreatureShadowLord  = 47
	CreatureBlackthorn  = 14
	CreatureLordBritish = 15
	// CreatureMimicIdx / CreatureReaperIdx 是**不會移動**的兩種
	//(`sub_AE20` 開頭 `cmp dl, 1Ah` / `1Bh` 直接 return)。
	CreatureMimicIdx  = 0x1A
	CreatureReaperIdx = 0x1B
	// CreatureGazerIdx 的攻擊會讓目標睡著(`sub_B9A8` 的 `cmp …, 1Ch`)。
	CreatureGazerIdx = 0x1C
)

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

// CreatureSpellImmune 回報這種生物是不是**完全不吃法術抗性判定的對象**。
//
// `sub_189BC` 只認三個編號:**黑刺(14)、不列顛王(15)、暗影領主(47)**。
// 魅惑、變形、驅離、恐懼、風系睡眠與能量,全部先問這一句 —— 三個都答「不」。
// 劇情人物不能被一發二圈咒語處理掉,原版是用寫死的編號擋的,不是靠數值。
func CreatureSpellImmune(creature int) bool {
	return creature == CreatureBlackthorn ||
		creature == CreatureLordBritish ||
		creature == CreatureShadowLord
}
