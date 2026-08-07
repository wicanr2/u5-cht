package game

import (
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 魔法
//
// 施法的流程完全照原版 `sub_1994C`,順序很重要 —— 每一關都有自己的失敗訊息,
// 而且**已調配的份數在魔力檢查之前就先扣掉了**:
//
//	1. 場合對不對(`byte_40E7C[咒語]` vs 當前地點)  → 「Not here!」
//	2. 有沒有調配好的份數                            → 「None mixed!」
//	3. ★ 份數 −1                                     ← 後面失敗也不退
//	4. 魔力夠不夠(消耗 = 圈數)                      → 「M.P. too low!」
//	5. 魔力 −= 圈數
//	6. 等級夠不夠(等級 ≥ 圈數)                      → 失敗
//	7. 跑效果
//	8. 「Success!」/「Failed!」
//
// 戰鬥中還多一關(`sub_200BC`):**上一個打你的敵人如果就站在旁邊,
// 咒語會被打斷**（印「X interferes!」)。

// MagicResult 是一次施法的結果。
type MagicResult int

const (
	// MagicNotHere:這個場合不能施這個咒語。
	MagicNotHere MagicResult = iota
	// MagicNoneMixed:沒有調配好的份數。
	MagicNoneMixed
	// MagicNoMana:魔力不足。
	MagicNoMana
	// MagicInterfered:戰鬥中被相鄰的敵人打斷。
	MagicInterfered
	// MagicFailed:發動了但失敗(等級不足,或效果本身回報失敗)。
	MagicFailed
	// MagicSuccess:成功。
	MagicSuccess
)

// castLocation 回傳給咒語場合判斷用的地點值:戰鬥中是 −1。
func (s *State) castLocation() int {
	if s.InCombat() {
		return u5data.CombatLocation
	}
	return s.Location
}

// Cast 讓名冊第 caster 個角色施第 spell 個咒語。
func (s *State) Cast(caster, spell int) MagicResult {
	if s.Spells == nil || spell < 0 || spell >= u5data.SpellCount {
		return MagicFailed
	}
	ch := s.rosterAt(caster)
	if ch == nil {
		return MagicFailed
	}
	sp := s.Spells.Spells[spell]

	// 戰鬥中:上一個打你的人站在旁邊就會被打斷。
	if s.InCombat() {
		if who, ok := s.interferer(caster); ok {
			s.Log(who + "干擾了施法!")
			return MagicInterfered
		}
	}
	if !sp.CanCastAt(s.castLocation()) {
		s.Log("此地不可!")
		return MagicNotHere
	}
	if s.Inventory.Spells[spell] <= 0 {
		s.Log("尚未調配!")
		return MagicNoneMixed
	}
	// ★ 先扣份數 —— 後面任何一步失敗都不退。原版就是這個順序。
	s.Inventory.Spells[spell]--

	if int(ch.MP) < sp.Circle {
		s.Log("魔力不足!")
		return MagicNoMana
	}
	ch.MP -= byte(sp.Circle)

	if int(ch.Level) < sp.Circle {
		s.Log("失敗!(等級不足以驅動第 " + itoa(sp.Circle) + " 圈)")
		return MagicFailed
	}
	if s.spellEffect(caster, spell) {
		s.Log("成功!")
		return MagicSuccess
	}
	s.Log("失敗!")
	return MagicFailed
}

// CastByName 讓玩家用上古語名稱施法(名稱一律用英文 canonical 值比對)。
func (s *State) CastByName(caster int, name string) MagicResult {
	idx := s.Spells.Find(name)
	if idx < 0 {
		s.Log("無此咒語!")
		return MagicFailed
	}
	return s.Cast(caster, idx)
}

// interferer 找出戰鬥中打斷施法的那個敵人(原版 `sub_200BC`)。
//
// 條件全部成立才算干擾:那個單位是**上一個攻擊施法者的人**、還活著、
// 站在敵方、沒被凍住也沒睡著、時間沒被停住,而且**距離正好 1**。
func (s *State) interferer(caster int) (string, bool) {
	c := s.Combat
	if c == nil {
		return "", false
	}
	self := s.combatSlotOfRoster(caster)
	if self < 0 {
		return "", false
	}
	by := int(c.LastAttacker[self])
	if by < 0 || by >= CombatUnitSlots {
		return "", false
	}
	u, t := &c.Units[self], &c.Units[by]
	if !t.Active() || !s.hostile(t) || s.hostile(u) {
		return "", false
	}
	if t.Flags&(UnitFrozen|UnitAsleep) != 0 {
		return "", false
	}
	if combatDistance(u.X, u.Y, t.X, t.Y) != 1 {
		return "", false
	}
	return s.unitName(t), true
}

// combatSlotOfRoster 找名冊第 n 個角色在戰場上的槽號;不在場回 -1。
func (s *State) combatSlotOfRoster(n int) int {
	c := s.Combat
	if c == nil {
		return -1
	}
	for i := 0; i < u5data.CombatPartySlots; i++ {
		if c.Units[i].IsParty() && c.Units[i].Roster == n {
			return i
		}
	}
	return -1
}

// rosterAt 取名冊第 n 個角色。
func (s *State) rosterAt(n int) *u5data.Character {
	if n < 0 || n >= len(s.Roster) {
		return nil
	}
	return &s.Roster[n]
}

// 咒語索引。用得到的才取名 —— 其餘照編號走 spellEffect 的 default。
const (
	SpellInLor      = 0  // 光
	SpellGravPor    = 1  // 魔法飛彈
	SpellAnZu       = 2  // 喚醒
	SpellAnNox      = 3  // 解毒
	SpellMani       = 4  // 治療
	SpellRelHur     = 8  // 改風向
	SpellInWis      = 9  // 定位
	SpellKalXen     = 10 // 召喚野獸
	SpellInXenMani  = 11 // 造食物
	SpellVasLor     = 12 // 大光明
	SpellVasFlam    = 13 // 火球
	SpellInSanct    = 19 // 防護
	SpellUusPor     = 21 // 上樓
	SpellDesPor     = 22 // 下樓
	SpellVasMani    = 27 // 大治療
	SpellRelTym     = 29 // 緩速
	SpellInVasPorY  = 30 // 能量爆
	SpellQuasAnWis  = 31 // 混亂
	SpellInAn       = 32 // 抗魔
	SpellAnXenEx    = 34 // 魅惑
	SpellSanctLor   = 36 // 隱形
	SpellXenCorp    = 37 // 殺
	SpellInManiCorp = 42 // 復活
	SpellAnTym      = 47 // 時間停止
)

// 三個「指定目標打一下」的咒語各自帶一個攻擊碼(原版 `sub_189E4` 寫進 `byte_3E0AD`)。
const (
	spellAttackGravPor = 0x30
	spellAttackVasFlam = 0x31
	spellAttackXenCorp = 0x32
)

// TimeStopTurns 是 An Tym 停多久(原版 `sub_198E0` 的 `byte_3E09E = 0Ah`)。
const TimeStopTurns = 10

// LightSpellTurns 是 In Lor / Vas Lor 的持續回合(原版 `sub_1D310` 的參數)。
const (
	LightSpellTurns     = 100
	GreatLightSpellTurn = 255
)

// spellEffect 跑咒語的效果,回傳成功與否(對應原版跳表 `jpt_19B27` 的 48 個 case)。
//
// ⚠ 這裡只實作**效果已經逆完**的那些;其餘照實說明還沒做,而且**照原版一樣
// 已經把份數與魔力扣掉了** —— 假裝成功比誠實失敗更糟。
func (s *State) spellEffect(caster, spell int) bool {
	switch spell {
	case SpellInLor:
		s.LightTurns = LightSpellTurns
		return true
	case SpellVasLor:
		s.LightTurns = GreatLightSpellTurn
		return true

	case SpellAnZu: // 喚醒:狀態 'S' → 'G',戰場上也清掉睡著旗標
		return s.cureStatus(u5data.StatusAsleep)
	case SpellAnNox: // 解毒:狀態 'P' → 'G'
		return s.cureStatus(u5data.StatusPoisoned)

	case SpellMani: // 治療:回 1..30,上限 MaxHP;死人治不了
		return s.healTarget(s.spellTarget(caster, false), s.AttackRoll())
	case SpellVasMani: // 大治療:回滿
		t := s.spellTarget(caster, false)
		ch := s.rosterAt(t)
		if ch == nil || ch.Status == u5data.StatusDead {
			return false
		}
		ch.HP = ch.MaxHP
		s.Log(ch.Name + "的傷全好了。")
		return true

	case SpellInManiCorp: // 復活
		t := s.spellTarget(caster, true)
		ch := s.rosterAt(t)
		if ch == nil || ch.Status != u5data.StatusDead {
			s.Log("此人未死!")
			return false
		}
		ch.Status = u5data.StatusGood
		ch.HP = 1
		s.Log(ch.Name + "回到了人世。")
		return true

	case SpellGravPor:
		return s.spellAttack(caster, spellAttackGravPor)
	case SpellVasFlam:
		return s.spellAttack(caster, spellAttackVasFlam)
	case SpellXenCorp:
		return s.spellAttack(caster, spellAttackXenCorp)

	case SpellAnXenEx: // 魅惑:讓一隻怪物改站我方
		return s.charmNearest(caster)

	case SpellAnTym: // 時間停止
		s.Log("時間靜止了。")
		return s.setCombatMode(CombatModeTimeStop, TimeStopTurns)

	case SpellUusPor:
		return s.spellChangeFloor(true)
	case SpellDesPor:
		return s.spellChangeFloor(false)

	// 五個共用 `byte_3E08A` 的持續效果(`sub_1D31C(模式, 回合, 音效)`)。
	case SpellInSanct:
		return s.setCombatMode(CombatModeProtected, 20)
	case SpellRelTym:
		return s.setCombatMode(CombatModeSlow, 30)
	case SpellQuasAnWis:
		return s.setCombatMode(CombatModeConfuse, 20)
	case SpellInAn:
		return s.setCombatMode(CombatModeNegate, 10)

	case SpellInXenMani: // 造食物:糧食 + random(1,3),上限 9999
		s.Inventory.Food = addCap(s.Inventory.Food, s.Roll(1, 3), 9999)
		s.Log("糧食憑空出現。")
		return true

	case SpellInWis: // 定位:報出目前的世界座標
		s.Log("汝身在 (" + itoa(s.X) + ", " + itoa(s.Y) + ")。")
		return true

	case SpellSanctLor: // 隱形:掛上隱形位元,選目標時就選不到了
		return s.hideCaster(caster)

	case SpellInVasPorY: // 能量爆:對場上每個敵人各擲一次
		return s.energyBlast(caster)

	case SpellKalXen: // 召喚野獸
		return s.summonCreature(caster, 20) // 巨鼠
	}
	s.Log("(此咒語的效果尚未實作 —— 藥草與魔力已照原版消耗)")
	return false
}

// spellTarget 挑「對某個隊員生效」的咒語要作用在誰身上。
//
// 原版會先問「Who?」(`sub_1C9C0` 印出隊伍清單等一個數字鍵)。引擎還沒有
// 那一層選單,先套用「戰鬥中是施法者自己、平時是隊伍裡傷得最重的那一個」
// —— **這是介面上的近似,不是規則**,補上選單之後就換掉。
//
// wantDead 為真時改找死掉的那一個(復活咒語用)。
func (s *State) spellTarget(caster int, wantDead bool) int {
	if wantDead {
		for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
			if s.Roster[i].Status == u5data.StatusDead {
				return i
			}
		}
		return caster
	}
	if s.InCombat() {
		return caster
	}
	best, worst := caster, 1<<30
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		c := &s.Roster[i]
		if c.Status == u5data.StatusDead {
			continue
		}
		if d := int(c.MaxHP) - int(c.HP); d > 0 && int(c.HP) < worst {
			worst, best = int(c.HP), i
		}
	}
	return best
}

// healTarget 回血,上限 MaxHP;死人治不了(原版 `sub_1CD3C`)。
func (s *State) healTarget(target, amount int) bool {
	ch := s.rosterAt(target)
	if ch == nil || ch.Status == u5data.StatusDead {
		return false
	}
	hp := int(ch.HP) + amount
	if hp > int(ch.MaxHP) {
		hp = int(ch.MaxHP)
	}
	ch.HP = uint16(hp)
	s.Log(ch.Name + "回復了 " + itoa(amount) + " 點體力。")
	return true
}

// cureStatus 把隊伍裡所有處在某個狀態的人恢復成良好。
func (s *State) cureStatus(status byte) bool {
	any := false
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		ch := &s.Roster[i]
		if ch.Status != status {
			continue
		}
		ch.Status = u5data.StatusGood
		any = true
		s.Log(ch.Name + "恢復了。")
		if status == u5data.StatusAsleep && s.InCombat() {
			if slot := s.combatSlotOfRoster(i); slot >= 0 {
				s.Combat.Units[slot].Flags &^= UnitAsleep
			}
		}
	}
	return any
}

// spellAttack 是三個「指定目標打一下」的咒語(原版 `sub_189E4`)。
//
// 攻擊碼寫進原版的 `byte_3E0AD`,由投射物與傷害那一段解讀;
// 差別在傷害來源與動畫,命中判定與一般攻擊共用。
func (s *State) spellAttack(caster, kind int) bool {
	c := s.Combat
	if c == nil {
		s.Log("此地無敵可擊。")
		return false
	}
	self := s.combatSlotOfRoster(caster)
	if self < 0 {
		return false
	}
	target, _, _ := s.aiTarget(self)
	if target < 0 {
		s.Log("沒有目標。")
		return false
	}
	dmg := spellAttackDamage(kind)
	if dmg == u5data.InstantKillDamage {
		s.Log(s.unitName(&c.Units[target]) + "被咒語擊中!")
	}
	s.applyDamage(self, target, dmg)
	return true
}

// spellAttackDamage 是三個攻擊咒語的傷害。
//
// ⚠ `sub_189E4` 只把攻擊碼記到 `byte_3E0AD`,實際數值在投射物那一段
// (`sub_20134` / `sub_1FE54`),還沒逆完。這裡用「Xen Corp 是必殺」這條
// 確定的事實,另外兩個先按圈數換算並在文件裡標明是**估計值**。
func spellAttackDamage(kind int) int {
	switch kind {
	case spellAttackXenCorp:
		return u5data.InstantKillDamage
	case spellAttackVasFlam:
		return 20
	default:
		return 10
	}
}

// charmNearest 是 An Xen Ex:把最近的怪物變成我方(原版 `sub_194CC`,
// 印「Creature: X charmed!」)。
func (s *State) charmNearest(caster int) bool {
	c := s.Combat
	if c == nil {
		return false
	}
	self := s.combatSlotOfRoster(caster)
	if self < 0 {
		return false
	}
	target, _, _ := s.aiTarget(self)
	if target < 0 {
		return false
	}
	t := &c.Units[target]
	if t.IsParty() {
		return false
	}
	t.Flags |= UnitSideFlip
	s.Log(s.unitName(t) + "被馴服了!")
	return true
}

// spellChangeFloor 是 Uus Por / Des Por。up 為真是往上。
//
// 兩個咒語的可施法場合表只放行**地牢**,所以正常情況一定走地牢那條;
// 場景那條留著是防呆(場景的樓層號與地牢相反 —— 場景 +1 是往上,
// 地牢 +1 是往下)。
func (s *State) spellChangeFloor(up bool) bool {
	if s.InDungeon() {
		d := -1
		if !up {
			d = 1
		}
		return s.DungeonChangeLevel(d)
	}
	if !s.InScene() || s.Scenes == nil {
		s.Log("此處無路可去。")
		return false
	}
	d := 1
	if !up {
		d = -1
	}
	if _, err := s.Scenes.Map(s.Location, s.Floor+d); err != nil {
		s.Log("此處無路可去。")
		return false
	}
	s.changeFloor(d)
	return true
}

// 調配咒語(原版 `sub_18704`)
//
// 玩家挑一個咒語、挑一組藥草、挑份數。**挑到的藥草一定會被扣掉**,
// 挑錯了也一樣 —— 錯的組合只會得到一團廢渣。組合要**完全相符**才算成功。

// Mix 調配 count 份 spell,用 reagents 指定的那幾種藥草。
//
// 回傳成功與否;不論成敗,選到的藥草都會被扣掉(這是原版行為)。
func (s *State) Mix(spell, count int, reagents []int) bool {
	if s.Spells == nil || spell < 0 || spell >= u5data.SpellCount || count <= 0 {
		return false
	}
	var picked byte
	for _, r := range reagents {
		if r < 0 || r >= u5data.ReagentCount {
			continue
		}
		picked |= u5data.ReagentBit(r)
	}
	// 份數受限於最少的那一種藥草。
	for r := 0; r < u5data.ReagentCount; r++ {
		if picked&u5data.ReagentBit(r) == 0 {
			continue
		}
		if have := s.Inventory.Reagents[r]; have < count {
			count = have
		}
	}
	if count <= 0 {
		s.Log("藥草不足。")
		return false
	}
	// ★ 先扣藥草 —— 配錯了也不退。
	for r := 0; r < u5data.ReagentCount; r++ {
		if picked&u5data.ReagentBit(r) != 0 {
			s.Inventory.Reagents[r] -= count
		}
	}
	if picked != s.Spells.Spells[spell].Reagents {
		s.Log("配方不對,藥草白費了。")
		return false
	}
	n := s.Inventory.Spells[spell] + count
	if n > u5data.SpellStackLimit {
		n = u5data.SpellStackLimit
	}
	s.Inventory.Spells[spell] = n
	s.Log("完成!" + s.Spells.Spells[spell].Name + " ×" + itoa(count))
	return true
}

// MixByRecipe 用咒語自己的配方調配 count 份(玩家不用一個個挑藥草)。
func (s *State) MixByRecipe(spell, count int) bool {
	if s.Spells == nil || spell < 0 || spell >= u5data.SpellCount {
		return false
	}
	return s.Mix(spell, count, s.Spells.Spells[spell].ReagentList())
}

// AdvanceTime 推進遊戲時鐘,順便讓光源燒掉同樣的分鐘數。
//
// 原版 `sub_29304(分鐘)` 的順序:**先看時間有沒有被停住**
//(`byte_3E08A == 'T'` 就整段跳過),否則對 `byte_3E0B7`(火把)與
// `byte_3E0B6`(光明咒語)各做一次飽和減法。
//
// ⇒ **An Tym 期間火把不會燒。** 這不是小事:U5 玩家在地牢裡靠這一點省火把。
func (s *State) AdvanceTime(minutes int) {
	if s.TimeStop > 0 {
		s.Clock.Advance(minutes)
		return
	}
	s.Clock.Advance(minutes)
	s.LightTurns = subFloor(s.LightTurns, minutes)
	s.TorchTurns = subFloor(s.TorchTurns, minutes)
}

// subFloor 是飽和減法(原版 `sub_2BBFC`)。
func subFloor(v, n int) int {
	if v -= n; v < 0 {
		return 0
	}
	return v
}

// HasLight 回報現在看不看得見(原版 `byte_3E0B6` 或 `byte_3E0B7` 任一非零)。
func (s *State) HasLight() bool { return s.LightTurns > 0 || s.TorchTurns > 0 }

// TorchMinutes 是點一把火把能撐多久(原版 `sub_17630`:`random(0,15) + 0x70`)。
func (s *State) TorchMinutes() int { return s.Roll(0, 15) + 0x70 }

// LightTorch 點一把火把。
func (s *State) LightTorch() bool {
	if s.Inventory.Torches <= 0 {
		s.Log("沒有火把了。")
		return false
	}
	s.Inventory.Torches--
	n := s.TorchMinutes()
	if n > s.TorchTurns {
		s.TorchTurns = n
	}
	s.Log("火把點亮了。")
	return true
}

// 施法的輸入流程
//
// 原版問「Spell name:」然後等玩家把**上古語**打進去(`sub_1CA0C`)。
// 咒語名因此是玩家要輸入的字串 —— canonical 值一律維持英文,
// 中文只出現在說明裡(CLAUDE.md §5.2 的硬規則)。

// BeginCastPrompt 開始問咒語名。
func (s *State) BeginCastPrompt() {
	if s.Spells == nil {
		s.Log("咒語表未載入。")
		return
	}
	s.castReturn = s.Prompt
	s.castBy = s.currentCaster()
	s.Prompt = PromptSpell
	s.Input = ""
	s.Log("咒語名:")
}

// currentCaster 是現在該由誰施法:戰鬥中是輪到的那個隊員,平時是隊長。
func (s *State) currentCaster() int {
	if c := s.Combat; c != nil && c.Turn >= 0 && c.Turn < CombatUnitSlots {
		if r := c.Units[c.Turn].Roster; r >= 0 {
			return r
		}
	}
	return 0
}

// SubmitSpell 把打好的咒語名送出去。
func (s *State) SubmitSpell() {
	name := trimSpace(s.Input)
	s.Input = ""
	s.Prompt = s.castReturn
	if name == "" {
		s.Log("作罷。")
		return
	}
	res := s.CastByName(s.castBy, name)
	// 戰鬥中不論成敗都算用掉一個回合。
	if s.InCombat() && res != MagicNotHere {
		s.afterPlayerAction()
	}
}

// CancelSpell 取消輸入。
func (s *State) CancelSpell() {
	s.Input = ""
	s.Prompt = s.castReturn
	s.Log("作罷。")
}

// trimSpace 去掉頭尾空白(咒語名前後可能被多打了空格)。
func trimSpace(v string) string {
	i, j := 0, len(v)
	for i < j && v[i] == ' ' {
		i++
	}
	for j > i && v[j-1] == ' ' {
		j--
	}
	return v[i:j]
}

// 全域戰鬥模式(原版 `byte_3E08A`,由 `sub_1D31C(模式, 回合, 音效)` 設定)
//
// 五個咒語共用**同一個位元組** —— 後施的會蓋掉先施的,不是一組可疊加的旗標。
// 每個模式的讀取處都在別的地方,所以語意是從「誰在讀它」反推的:
//
//	'T' An Tym       `sub_A108` 開頭直接 return  → 敵人整個不動
//	'Q' Rel Tym      `sub_A108` 擲 1/2 才行動    → 敵人只有一半的回合能動
//	'C' Quas An Wis  `sub_AC40` 可能把陣營看反   → 敵人亂打
//	'N' In An        `sub_9E10` / `sub_AE20` 擋掉 → 施法者不能放遠程、不能瞬移
//	'P' In Sanct     讀它的地方還沒找到           → 只記著,沒有效果
const (
	CombatModeNone      = 0
	CombatModeTimeStop  = 'T'
	CombatModeSlow      = 'Q'
	CombatModeConfuse   = 'C'
	CombatModeNegate    = 'N'
	CombatModeProtected = 'P'
)

// setCombatMode 設定全域戰鬥模式(原版 `sub_1D31C`)。
func (s *State) setCombatMode(mode byte, turns int) bool {
	s.CombatMode, s.CombatModeTurns = mode, turns
	if mode == CombatModeTimeStop {
		s.TimeStop = turns
	}
	return true
}

// tickCombatMode 讓模式倒數一回合(原版 `sub_16370`,在玩家單位回合結束時)。
func (s *State) tickCombatMode() {
	if s.CombatModeTurns <= 0 {
		return
	}
	s.CombatModeTurns--
	if s.CombatModeTurns == 0 {
		s.CombatMode = CombatModeNone
		s.Log("咒語的效力消退了。")
	}
}


// hideCaster 是 Sanct Lor(原版 `sub_19674`)。
//
// 只做兩件事:給自己掛上隱形位元(0x10)、把地圖物件的 tile 換成 0x1D。
// 隱形位元就是 `sub_AC40` 選目標時會跳過的那一個 —— 隱形不是減傷,
// 是**根本不會被鎖定**。
func (s *State) hideCaster(caster int) bool {
	slot := s.combatSlotOfRoster(caster)
	if slot < 0 {
		return false
	}
	s.Combat.Units[slot].Flags |= UnitHidden
	s.Log(s.unitName(&s.Combat.Units[slot]) + "的身影淡去了。")
	return true
}

// energyBlast 是 In Vas Por Ylem(原版 `sub_19440`)。
//
// 對場上**每一個**敵人各判一次:擲 1..30 對上該單位的防禦,擲贏才吃
// `random(1, 20)` 的傷害,而經驗值照樣進施法者的帳。
func (s *State) energyBlast(caster int) bool {
	c := s.Combat
	if c == nil {
		s.Log("此地無敵可擊。")
		return false
	}
	self := s.combatSlotOfRoster(caster)
	if self < 0 {
		return false
	}
	hit := false
	for i := range c.Units {
		t := &c.Units[i]
		if !t.Active() || !s.hostile(t) {
			continue
		}
		def := 0
		if st := s.creatureOf(t); st != nil {
			def = int(st.Armour)
		} else if ch := s.charOf(t); ch != nil {
			def = s.Stats.DefenceOf(ch)
		}
		if s.AttackRoll() < def {
			continue
		}
		s.applyDamage(self, i, s.Roll(1, 20))
		hit = true
	}
	return hit
}

// summonCreature 是召喚類咒語(Kal Xen / In Bet Xen / Kal Xen Corp,
// 原版 `sub_18F2C` / `sub_192BC` / `sub_1CE70`)。
//
// 三支都是「挑一個空位 → `sub_2EAE4` 生一隻」,差別只在生的是什麼。
// ⚠ **各自召哪一種還沒逆完** —— `sub_B1D8` 挑位置那段讀懂了,
// 生物編號那一段沒有。這裡用「Kal Xen 召野獸」這個 U5 常識填 Kal Xen,
// 另外兩支先不接,文件裡標明是**推測**。
func (s *State) summonCreature(caster, creature int) bool {
	c := s.Combat
	if c == nil {
		s.Log("此地不可召喚。")
		return false
	}
	self := s.combatSlotOfRoster(caster)
	if self < 0 {
		return false
	}
	u := &c.Units[self]
	for try := 0; try < 16; try++ {
		x, y := u.X+s.Roll(-1, 1), u.Y+s.Roll(-1, 1)
		if s.combatBlocked(self, x, y) {
			continue
		}
		for i := u5data.CombatPartySlots; i < CombatUnitSlots; i++ {
			if c.Units[i].Flags != 0 {
				continue
			}
			s.placeEnemy(c, i-u5data.CombatPartySlots,
				u5data.CreatureBase+byte(creature*4), creature)
			c.Units[i].X, c.Units[i].Y = x, y
			// 召出來的站我方 —— 怪物的陣營反轉位元代表「被馴服」。
			c.Units[i].Flags |= UnitSideFlip
			s.Log(s.unitName(&c.Units[i]) + "應召而來!")
			return true
		}
		return false
	}
	return false
}
