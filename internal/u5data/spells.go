package u5data

import (
	"fmt"
	"os"
	"path/filepath"
)

// 咒語表
//
// 三張表都在 DOS `DATA.OVL` 裡,位址是拿 FM Towns 的資料段逐位元組比對出來的:
//
//	FM Towns   DOS       內容
//	0x46884    0x0709    48 個咒語名 + "Frotz"(NUL 分隔的字串,不是指標表)
//	0x40E7C    0x1CA0    可施法的場合 48 B(位元遮罩)
//	0x40EAC    0x1CD0    藥草組合   48 B(位元遮罩,最高位元 = 第 0 種藥草)
//
// **圈數 = 咒語索引 / 6 + 1**(原版 `sub_1994C` 的 `idiv 6; inc eax`),
// 每圈正好 6 個、共 8 圈。圈數同時是**魔力消耗**與**最低等級**:
// 等級不足的人連咒語都發不動(`byte_3DDCA[角色*32] < 圈數` → 直接失敗)。
const (
	spellNames   = 0x0709
	spellContext = 0x1CA0
	spellReagent = 0x1CD0
)

// SpellCount 是咒語數(8 圈 × 6)。
const SpellCount = 48

// SpellsPerCircle 是每圈幾個。
const SpellsPerCircle = 6

// 可施法場合的位元(`byte_40E7C[咒語]`,原版 `sub_1994C` 依當前地點挑一個來測)。
//
//	地點 == 0        → CastInCombat 之外的 CastOutdoors ... 見 CanCastAt
//	地點 > 0x7F      → CastInCombat(戰鬥中原版把地點設成 −1)
//	地點 < 0x21      → CastInTown
//	其餘(0x21..0x7F)→ CastInDungeon
const (
	CastInCombat  = 0x01
	CastInDungeon = 0x02
	CastInTown    = 0x04
	CastOutdoors  = 0x08
)

// 藥草(順序 = 位元由高到低,`sub_18704` 的 `mov ebx, 80h; sar ebx, 1`)。
const (
	ReagentSulfurousAsh = iota
	ReagentGinseng
	ReagentGarlic
	ReagentSpiderSilk
	ReagentBloodMoss
	ReagentBlackPearl
	ReagentNightshade
	ReagentMandrakeRoot
)

// Spell 是一個咒語。
type Spell struct {
	// Name 是咒語的**上古語**名稱。⚠ 這是玩家要打進去的字串,
	// 所以 canonical 值永遠維持英文(見 CLAUDE.md §5.2 的硬規則)。
	Name string
	// Circle 是圈數 1..8,同時是魔力消耗與最低等級需求。
	Circle int
	// Reagents 是需要的藥草(位元遮罩,**最高位元 0x80 = 第 0 種藥草**)。
	//
	// ⚠ 這個順序是反的:原版 `sub_18704` 用 `mov ebx, 80h` 起頭、每輪
	// `sar ebx, 1`,所以藥草 0(硫磺灰)對應 0x80 而不是 0x01。
	// 用 ReagentBit 取位元,不要直接 `1 << r`。
	Reagents byte
	// Context 是可以在哪裡施(Cast* 位元)。
	Context byte
}

// SpellTable 是 48 個咒語。
type SpellTable struct {
	Spells [SpellCount]Spell
	// Extra 是索引 48 的 "Frotz" —— 排在咒語表後面但不在 8 圈之內。
	Extra string
}

// LoadSpells 從 DATA.OVL 讀出咒語表。
func LoadSpells(dir string) (*SpellTable, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "DATA.OVL"))
	if err != nil {
		return nil, err
	}
	return ParseSpells(raw)
}

// ParseSpells 從 DATA.OVL 的內容取出咒語表。
func ParseSpells(ovl []byte) (*SpellTable, error) {
	if len(ovl) < spellReagent+SpellCount {
		return nil, fmt.Errorf("DATA.OVL 只有 %d B,放不下咒語表", len(ovl))
	}
	t := &SpellTable{}
	off := spellNames
	for i := 0; i <= SpellCount; i++ {
		end := off
		for end < len(ovl) && ovl[end] != 0 {
			end++
		}
		if end >= len(ovl) {
			return nil, fmt.Errorf("咒語名在第 %d 個就跑出檔尾了", i)
		}
		name := string(ovl[off:end])
		off = end + 1
		if i == SpellCount {
			t.Extra = name
			break
		}
		t.Spells[i] = Spell{
			Name:     name,
			Circle:   i/SpellsPerCircle + 1,
			Reagents: ovl[spellReagent+i],
			Context:  ovl[spellContext+i],
		}
	}
	if err := t.validate(); err != nil {
		return nil, err
	}
	return t, nil
}

// validate 用「位移偏掉就一定違反」的性質把三張表釘住。
func (t *SpellTable) validate() error {
	// 名字:第一個一定是 In Lor、最後一個一定是 An Tym,後面接 Frotz。
	if t.Spells[0].Name != "In Lor" {
		return fmt.Errorf("第 0 個咒語是 %q,預期 In Lor", t.Spells[0].Name)
	}
	if t.Spells[SpellCount-1].Name != "An Tym" {
		return fmt.Errorf("第 47 個咒語是 %q,預期 An Tym", t.Spells[SpellCount-1].Name)
	}
	if t.Extra != "Frotz" {
		return fmt.Errorf("咒語表後面接的是 %q,預期 Frotz", t.Extra)
	}
	// 藥草:In Lor 只要硫磺灰,Mani 是人參 + 蛛絲 —— 兩條都是 U5 的常識,
	// 而且用到不同的位元,一起對上才算表沒偏。
	if got := t.Spells[0].Reagents; got != ReagentBit(ReagentSulfurousAsh) {
		return fmt.Errorf("In Lor 的藥草是 %08b,預期只有硫磺灰", got)
	}
	if want := ReagentBit(ReagentGinseng) | ReagentBit(ReagentSpiderSilk); t.Spells[4].Reagents != want {
		return fmt.Errorf("Mani 的藥草是 %08b,預期人參 + 蛛絲(%08b)",
			t.Spells[4].Reagents, want)
	}
	// 場合:Uus Por / Des Por(上下樓)只在地牢、Rel Hur(改風向)只在野外。
	for _, i := range []int{21, 22} {
		if t.Spells[i].Context != CastInDungeon {
			return fmt.Errorf("%s 的可施法場合是 %04b,預期只有地牢",
				t.Spells[i].Name, t.Spells[i].Context)
		}
	}
	if t.Spells[8].Context != CastOutdoors {
		return fmt.Errorf("Rel Hur 的可施法場合是 %04b,預期只有野外", t.Spells[8].Context)
	}
	return nil
}

// ReagentBit 是某種藥草在遮罩裡的位元(最高位元 = 第 0 種)。
func ReagentBit(r int) byte { return 0x80 >> uint(r) }

// NeedsReagent 回報這個咒語要不要某種藥草。
func (s Spell) NeedsReagent(r int) bool { return s.Reagents&ReagentBit(r) != 0 }

// ReagentList 依藥草編號順序回傳需要哪幾種。
func (s Spell) ReagentList() []int {
	var out []int
	for r := 0; r < ReagentCount; r++ {
		if s.NeedsReagent(r) {
			out = append(out, r)
		}
	}
	return out
}

// Find 依上古語名稱找咒語(不分大小寫)。回傳 -1 代表沒有這個咒語。
func (t *SpellTable) Find(name string) int {
	if t == nil {
		return -1
	}
	for i := range t.Spells {
		if equalFold(t.Spells[i].Name, name) {
			return i
		}
	}
	return -1
}

// equalFold 是 ASCII 的不分大小寫比較(咒語名一定是 ASCII)。
func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		x, y := a[i], b[i]
		if 'A' <= x && x <= 'Z' {
			x += 32
		}
		if 'A' <= y && y <= 'Z' {
			y += 32
		}
		if x != y {
			return false
		}
	}
	return true
}

// CanCastAt 回報在某個地點施不施得了這個咒語。
//
// location 就是原版的 `byte_3E0A3`:0 是大地圖、1..0x20 是城鎮與城堡、
// 0x21..0x7F 是地牢、戰鬥中原版把它設成 −1(這裡用 CombatLocation 表示)。
func (s Spell) CanCastAt(location int) bool {
	switch {
	case location == 0:
		return s.Context&CastOutdoors != 0
	case location > 0x7F || location < 0:
		return s.Context&CastInCombat != 0
	case location < 0x21:
		return s.Context&CastInTown != 0
	default:
		return s.Context&CastInDungeon != 0
	}
}

// CombatLocation 是「戰鬥中」的地點值(原版 `sub_2E364` 把 `byte_3E0A3` 設成 −1)。
const CombatLocation = -1

// SaveSpellsOffset 是存檔裡「已調配的咒語」那 48 B。
//
// 位移是從 FM Towns 的 `byte_3E000` 推出來的:同一段記憶體裡
// `byte_3DFD0`(裝備)對應存檔 0x021A、`byte_3E060`(藥草)對應 0x02AA,
// 兩者的差 0x3DDB6 一致 —— 所以 `byte_3E000` 對應 0x024A。
const SaveSpellsOffset = 0x024A

// SpellStackLimit 是每種咒語能存幾份(原版 `cmp byte_3E000[i], 63h` → 99)。
const SpellStackLimit = 99

// MixAmountMax 是「要幾份?」收得下的最大值。
//
// 原版問份數是 `sub_2B7F0(2)` —— **兩位數**,所以上限就是 99,
// 與 SpellStackLimit 同值但來源不同(一個是輸入欄寬,一個是儲存上限)。
const MixAmountMax = 99
