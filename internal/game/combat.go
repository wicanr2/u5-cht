package game

import (
	"strconv"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 戰鬥
//
// 進戰鬥的入口是 `sub_2E58C(物件槽)`:玩家撞上地圖上的怪物物件時觸發。
// 它做四件事 ——
//
//  1. 印敵人的名字(種類 < 0x40 一律叫 "PIRATES",否則查生物名表)
//  2. 依「隊伍在不在船上 / 敵人是不是船 / 敵人腳下的地形」挑戰鬥地圖
//     (`u5data.SelectCombatMap`)
//  3. `sub_2E51C` 載入那張圖,把隊員與敵人放到圖裡的入場位置
//  4. 進戰鬥迴圈
//
// 戰場是 11×11 —— 正好等於一般遊玩的視窗大小,所以戰鬥時整個視窗就是戰場。

// 戰場單位的旗標(原版的 `unit[+2]`)。
//
// 高兩位元說明「這一格站著誰」,低位元是狀態。`sub_AC40` 選目標時
// `flags == 0` 代表空槽,所以 unitParty / unitMonster 必須有一個成立。
const (
	// UnitSideFlip 是**陣營反轉**:隊員被魅惑就變敵方,怪物被馴服就變我方。
	// 同一個位元在兩種單位上意義相反,`sub_29A64` 就是靠這一點把兩邊算清楚。
	UnitSideFlip = 0x01
	// UnitFleeing 是逃跑中:朝目標的**反方向**走,而且**走得出戰場**
	//(`sub_16454` 對出界的格子只在這個旗標成立時放行)。
	UnitFleeing = 0x02
	// UnitFrozen 這一回合不能行動,也不會被選為目標。
	UnitFrozen = 0x04
	// UnitAsleep 睡著了:每輪有 1/17 機率醒來(`sub_A108` 的 `random(0,16) == 16`)。
	UnitAsleep = 0x08
	// UnitHidden 看不見,選不到當目標。
	UnitHidden = 0x10
	// UnitDead 死了。
	UnitDead = 0x20
	// UnitMonster / UnitParty 是「這一槽有東西」的兩種。
	UnitMonster = 0x40
	UnitParty   = 0x80
)

// CombatUnitSlots 是戰場單位表的長度(原版 `dword_3EF50` 是 32 槽 × 8 B)。
const CombatUnitSlots = 32

// initiativeBase 是行動倒數的基準:重設時算 `36 − 敏捷`
//(原版 `mov al, 24h; sub al, [ebx+1]`)。敏捷越高、倒數越短、出手越密。
const initiativeBase = 36

// Combatant 是戰場上的一個單位。
type Combatant struct {
	// Roster 是名冊索引(隊員);敵人是 -1。
	Roster int
	// Kind 是敵人的種類碼;隊員是 0。
	Kind byte
	// Creature 是生物索引(0..47),用來查屬性與旗標;隊員是 -1。
	Creature int
	// Tile 是畫出來的樣子。
	Tile int
	X, Y int
	// HP 是怪物的血。隊員的血記在名冊裡(原版也是這樣分開放的)。
	HP int
	// Dex 是敏捷,決定行動倒數。
	Dex int
	// Flags 是上面那組 Unit* 位元。
	Flags byte
	// Init 是行動倒數,每掃一輪減一,歸零就輪到它。
	Init int
}

// Dead 回報這個單位死了沒。
func (c *Combatant) Dead() bool { return c.Flags&UnitDead != 0 }

// Active 回報這一槽有沒有活著的單位。
func (c *Combatant) Active() bool {
	return c.Flags&(UnitParty|UnitMonster) != 0 && c.Flags&UnitDead == 0
}

// IsParty 回報這一槽放的是不是隊員(**不是**「站在哪一邊」——
// 被魅惑的隊員還是隊員,但陣營在敵方)。
func (c *Combatant) IsParty() bool { return c.Flags&UnitParty != 0 }

// Hostile 回報這個單位站在敵方(原版 `sub_29A64`)。
//
//	死了            → 不算任何一邊
//	隊員            → 陣營反轉位元;Saduj 例外(見 sadujName)
//	怪物            → 陣營反轉位元的**反面**
func (c *Combatant) Hostile(name string) bool {
	if c.Flags&UnitDead != 0 {
		return false
	}
	if c.Flags&UnitParty != 0 {
		if isSaduj(name) {
			return true
		}
		return c.Flags&UnitSideFlip != 0
	}
	return c.Flags&UnitSideFlip == 0
}

// isSaduj 認出那個叛徒。
//
// 原版寫得很硬:`cmp byte_3DDB8[角色*32], 'j'` —— 名字的**第 5 個字母**是 j
// 就永遠站在敵方。全檔只有這一處讀那個位元組。
//
// 會這樣寫是因為要認的只有一個人:**Saduj**(S-a-d-u-**j**),
// 名字倒過來是 Judas。加入隊伍之後在戰鬥裡反咬 —— 這是 U5 的既定劇情,
// 不是 bug。名冊 0 號(聖者本人)不套這條(原版 `and ecx,ecx; jz` 先擋掉)。
func isSaduj(name string) bool { return len(name) > 4 && name[4] == 'j' }

// Combat 是一場進行中的戰鬥。
type Combat struct {
	Map *u5data.CombatMap
	// MapIndex 是這張圖在 `.CBT` 裡的編號,方便對照 `u5dump cbt` 的輸出。
	MapIndex int
	// Units 是場上所有單位,隊員在前(0..5)、敵人在後(6..31)。
	//
	// 長度固定 32 —— 原版的槽號會被 AI 拿來當「離我最近的目標」的判準
	//(`sub_AC40` 由 31 往 0 掃),所以順序不能重排。
	Units [CombatUnitSlots]Combatant
	// Turn 是輪到誰行動(Units 的索引);-1 代表現在沒有玩家單位在等輸入。
	Turn int
	// EnemyName 是開場印的那個名字。
	EnemyName string
	// Over 為真時勝負已定,下一次輸入就離開戰場。
	Over bool
	// Won 記勝負,離場時決定要不要把怪物從地圖上清掉。
	Won bool
	// LastAttacker[槽] 是上一個攻擊這一槽的單位(原版 `byte_3E0B8`);
	// −1 代表沒有。施法被打斷的判定只看這個人。
	LastAttacker [CombatUnitSlots]int8
	// fromSlot 是觸發戰鬥的那個物件槽,打完要清掉。
	fromSlot int
	// scan 是排程掃到第幾槽,actions 是累計行動數(每 10 次 = 1 分鐘)。
	scan, actions int
}

// InCombat 回報是不是正在戰鬥。
func (s *State) InCombat() bool { return s.Combat != nil }

// enemyDisplayName 照 `sub_2E58C`:種類碼 < 0x40 的一律叫 PIRATES,
// 其餘查生物名表(索引 = (種類 − 64) / 4,與 docs/re/09 同一條公式)。
func (s *State) enemyDisplayName(kind byte) string {
	if kind < u5data.CreatureBase {
		return "PIRATES"
	}
	if n := s.Creatures.Name(kind); n != "" {
		return n
	}
	return "#" + strconv.Itoa(int(kind))
}

// BeginCombat 把玩家帶進戰鬥(原版 `sub_2E58C` → `sub_2E364` → `sub_2F0EC`)。
//
// slot 是撞上的那個怪物物件的槽號。回傳 false 代表打不起來
// (缺戰鬥地圖資料時誠實拒絕,而不是生一場空的仗)。
func (s *State) BeginCombat(slot int) bool {
	objs := s.currentObjects()
	if objs == nil || s.CombatMaps == nil {
		return false
	}
	o := &objs.Objects[slot]
	if !o.Present() {
		return false
	}
	kind := o.Kind &^ 0x03
	terrain := int(s.TileAt(o.X, o.Y))
	idx := u5data.SelectCombatMap(int(kind), terrain, s.Transport, !s.InScene())
	if idx < 0 || idx >= len(s.CombatMaps.Maps) {
		return false
	}
	m := &s.CombatMaps.Maps[idx]

	c := &Combat{Map: m, MapIndex: idx, fromSlot: slot, Turn: -1,
		EnemyName: s.enemyDisplayName(o.Kind)}
	for i := range c.LastAttacker {
		c.LastAttacker[i] = -1
	}
	// 隊員照圖裡的入場位置排;人數不足就只排前 n 個。
	for i, ch := range s.Party() {
		if i >= u5data.CombatPartySlots {
			break
		}
		u := &c.Units[i]
		*u = Combatant{
			Roster:   i,
			Creature: -1,
			Tile:     u5data.NPCTileBase + int(partyTileFor(ch)),
			X:        int(m.PartyX[i]),
			Y:        int(m.PartyY[i]),
			Dex:      int(ch.Dex),
			Flags:    UnitParty,
		}
		u.Init = u.resetInit()
		if ch.Status == u5data.StatusDead {
			u.Flags |= UnitDead
		}
		// 混沌之劍拿在手上的人不歸玩家指揮 —— 原版 `sub_A360` 一開始就檢查
		// 兩個武器欄位是不是 0x23,是的話立刻掛上陣營反轉位元交給 AI。
		if wieldsSwordOfChaos(ch) {
			u.Flags |= UnitSideFlip
		}
	}
	s.spawnEnemies(c, o)

	s.Combat = c
	s.Prompt = PromptCombat
	s.Log("「" + c.EnemyName + "」來襲!")
	s.Log("(戰場 #" + strconv.Itoa(idx) + ")")
	s.advanceCombat()
	return true
}

// wieldsSwordOfChaos 回報這名角色是不是握著混沌之劍。
//
// 原版查的是紀錄的 0x1B 與 0x1C 兩個裝備欄(左右手),值 0x23 就成立。
func wieldsSwordOfChaos(ch *u5data.Character) bool {
	e := ch.Equipment()
	return e.Weapon == u5data.ItemSwordOfChaos || e.Shield == u5data.ItemSwordOfChaos
}

// resetInit 算行動倒數的重設值(原版 `36 − 敏捷`)。
//
// 敏捷 30 的蝙蝠是 6、敏捷 6 的史萊姆是 30 —— 蝙蝠出手快五倍。
// 下限夾在 1,免得敏捷 ≥ 36 的單位一輪動無限次。
func (c *Combatant) resetInit() int {
	n := initiativeBase - c.Dex
	if n < 1 {
		n = 1
	}
	return n
}

// spawnEnemies 依原版 `sub_2F0EC` 決定這一場出現幾隻、哪幾種、站在哪。
func (s *State) spawnEnemies(c *Combat, o *u5data.MapObject) {
	base, ok := s.creatureIndexOf(o.Kind)
	if !ok {
		// 種類碼查不到生物索引(海盜船之類)——放一隻,屬性走預設。
		s.placeEnemy(c, 0, o.Kind, -1)
		return
	}
	// 物件槽 +5 的最高位元會讓原版把生物索引 +256,而那個「大版本」
	// **不吃**城鎮裡只出一隻的規則。原版寫成 `arg_4 -= 0x100` 之後
	// 用一個旗標記著,這裡直接用旗標表示。
	small := o.Raw[u5data.ObjShipHull] <= 0x7F

	count := 1
	if st := &s.Stats.Creature[base]; true {
		inTown := s.Location > 0 && s.Location <= 0x20
		if inTown && base != u5data.CreatureGuardIdx && small {
			// 在城鎮 / 城堡裡動手只打得到眼前那一個。衛兵例外 —— 叫來一整隊。
			count = 1
		} else {
			count = int(st.GroupMax)
			// 1 / 8 / 16 這三個值原版直接用,不擲骰。其餘擲 random(1, 上限)。
			if count != 1 && count != 8 && count != 16 {
				count = s.Roll(1, count)
			}
			// 隊伍佔 6 槽,加起來不能超過 32 槽。
			if count+u5data.CombatPartySlots > CombatUnitSlots-1 {
				count = 26
			}
		}
	}
	// 前四分之一的同伴各有 1/9 機率換成混編表指定的另一種怪物。
	mixUntil := count/4 + 1
	for i := 0; i < count; i++ {
		cre := base
		if i > 0 && i < mixUntil && s.Roll(0, 8) == 0 {
			cre = int(s.Stats.CreatureMix[base])
		}
		s.placeEnemy(c, i, u5data.CreatureBase+byte(cre*4), cre)
	}
}

// creatureIndexOf 把種類碼換成生物索引((編號 − 64) / 4)。
func (s *State) creatureIndexOf(kind byte) (int, bool) {
	if s.Stats == nil || kind < u5data.CreatureBase {
		return 0, false
	}
	i := int(kind-u5data.CreatureBase) / 4
	if i >= u5data.CreatureCount {
		return 0, false
	}
	return i, true
}

// placeEnemy 把第 n 隻敵人放到圖裡的第 n 個敵方入場點。
func (s *State) placeEnemy(c *Combat, n int, kind byte, creature int) {
	slot := u5data.CombatPartySlots + n
	if slot >= CombatUnitSlots || n >= len(c.Map.EnemyX) {
		return
	}
	u := &c.Units[slot]
	*u = Combatant{
		Roster:   -1,
		Kind:     kind,
		Creature: creature,
		Tile:     u5data.NPCTileBase + int(kind),
		X:        int(c.Map.EnemyX[n]),
		Y:        int(c.Map.EnemyY[n]),
		Flags:    UnitMonster,
		HP:       1,
		Dex:      1,
	}
	if creature >= 0 && s.Stats != nil {
		st := &s.Stats.Creature[creature]
		u.HP = int(st.MaxHP)
		u.Dex = int(st.Dex)
	}
	u.Init = u.resetInit()
}

// creatureOf 取一個單位的怪物屬性;隊員或查不到就回 nil。
func (s *State) creatureOf(u *Combatant) *u5data.CreatureStats {
	if s.Stats == nil || u.Creature < 0 || u.Creature >= u5data.CreatureCount {
		return nil
	}
	return &s.Stats.Creature[u.Creature]
}

// charOf 取一個隊員單位對應的名冊紀錄;敵人回 nil。
func (s *State) charOf(u *Combatant) *u5data.Character {
	if u.Roster < 0 || u.Roster >= len(s.Roster) {
		return nil
	}
	return &s.Roster[u.Roster]
}

// unitName 是印訊息用的名字。
func (s *State) unitName(u *Combatant) string {
	if ch := s.charOf(u); ch != nil {
		return ch.Name
	}
	return s.enemyDisplayName(u.Kind)
}

// hostile 回報某個單位站在敵方(原版 `sub_29A64`,含 Saduj 的例外)。
func (s *State) hostile(u *Combatant) bool {
	name := ""
	if ch := s.charOf(u); ch != nil {
		name = ch.Name
	}
	// 名冊 0 號是聖者本人,不套 Saduj 那條(原版先 `and ecx,ecx; jz` 擋掉)。
	if u.Roster == 0 {
		name = ""
	}
	return u.Hostile(name)
}

// sideCounts 數兩邊還剩幾個(原版 `sub_15DD4`)。
func (s *State) sideCounts(c *Combat) (enemies, party int) {
	for i := range c.Units {
		u := &c.Units[i]
		if !u.Active() {
			continue
		}
		if s.hostile(u) {
			enemies++
		} else {
			party++
		}
	}
	return
}

// playerControlled 回報這一槽現在該由玩家下指令。
//
// 站在我方 = 玩家控制。被魅惑的隊員與握著混沌之劍的人因此自動轉給 AI,
// 被馴服的怪物則反過來由玩家指揮 —— 這三種情況共用同一個判斷。
func (s *State) playerControlled(u *Combatant) bool { return !s.hostile(u) }

// advanceCombat 跑排程,直到輪到一個玩家控制的單位、或勝負已定。
//
// 原版 `sub_A9EC` 是一個 0..31 的無窮掃描:每掃到一個活著的單位就把它的
// 行動倒數減一,歸零的那個行動。**行動不是輪流,是按敏捷的頻率**——
// 敏捷 30 的蝙蝠每 6 輪動一次,敏捷 6 的史萊姆每 30 輪才動一次。
func (s *State) advanceCombat() {
	c := s.Combat
	if c == nil {
		return
	}
	c.Turn = -1
	// 掃描上限只是防呆:正常情況下最多 36 輪就會有人的倒數歸零。
	for guard := 0; guard < CombatUnitSlots*initiativeBase*2; guard++ {
		if c.Over {
			return
		}
		i := c.scan
		c.scan = (c.scan + 1) % CombatUnitSlots
		u := &c.Units[i]
		if !u.Active() {
			continue
		}
		// 隊員在名冊裡已經死了 → 同步到戰場上。
		if ch := s.charOf(u); ch != nil && ch.Status == u5data.StatusDead {
			u.Flags |= UnitDead
			continue
		}
		u.Init--
		if u.Init > 0 {
			continue
		}
		u.Init = u.resetInit()
		c.actions++
		if c.actions%10 == 0 {
			// 每 10 個單位行動走 1 分鐘(原版 `byte_3E092` 數到 10)。
			s.AdvanceTime(1)
		}
		if u.Flags&UnitFrozen != 0 {
			continue
		}
		if u.Flags&UnitAsleep != 0 {
			// 睡著的每次有 1/17 機率醒來。
			if s.Roll(0, 16) == 16 {
				u.Flags &^= UnitAsleep
				s.Log(s.unitName(u) + "醒了。")
			}
			continue
		}
		if s.playerControlled(u) {
			c.Turn = i
			return
		}
		// An Tym 期間敵人整個不動(原版 `sub_A108` 一開頭
		// `cmp byte_3E08A, 'T'` 就直接 return)。
		if s.TimeStop > 0 {
			continue
		}
		s.aiTurn(i)
		if s.checkCombatOver() {
			return
		}
	}
}

// checkCombatOver 判勝負(原版 `sub_A9EC` 每個單位行動完都查一次)。
func (s *State) checkCombatOver() bool {
	c := s.Combat
	enemies, party := s.sideCounts(c)
	switch {
	case enemies == 0:
		s.Log("勝利!")
		c.Over, c.Won = true, true
	case party == 0:
		s.Log("敗北!")
		c.Over, c.Won = true, false
	default:
		return false
	}
	c.Turn = -1
	return true
}

// partyTileFor 挑隊員在戰場上的圖。
//
// 原版用角色的職業決定 sprite;職業碼與生物編號的對應還沒追到,
// 先一律用聖者那一組 —— 這是**顯示**的近似,不影響規則。
func partyTileFor(_ *u5data.Character) byte { return 0x4C }

// CombatTileAt 回傳戰場上 (x, y) 該顯示什麼。
func (s *State) CombatTileAt(x, y int) byte {
	if s.Combat == nil {
		return u5data.TileBlank
	}
	return s.Combat.Map.At(x, y)
}

// CombatUnitAt 回報戰場上某一格站著誰。
func (c *Combat) CombatUnitAt(x, y int) (*Combatant, bool) {
	for i := range c.Units {
		u := &c.Units[i]
		if u.Active() && u.X == x && u.Y == y {
			return u, true
		}
	}
	return nil, false
}

// EndCombat 結束戰鬥回到地圖。won 為真時把觸發戰鬥的怪物從地圖上清掉。
func (s *State) EndCombat(won bool) {
	if s.Combat == nil {
		return
	}
	if won {
		if objs := s.currentObjects(); objs != nil {
			objs.Remove(s.Combat.fromSlot)
		}
	}
	s.Combat = nil
	s.Prompt = PromptNone
}

// CombatMove 讓目前輪到的單位往 d 走一步。
//
// 走出戰場邊緣就是**撤離**(原版沒有「逃跑鍵」——`sub_2F294` 判出界之後
// 把單位移出場,印「escapes!」)。
func (s *State) CombatMove(d Direction) {
	c := s.Combat
	if c == nil {
		return
	}
	if c.Over {
		s.EndCombat(c.Won)
		return
	}
	if c.Turn < 0 || c.Turn >= CombatUnitSlots {
		return
	}
	u := &c.Units[c.Turn]
	dx, dy := d.Delta()
	nx, ny := u.X+dx, u.Y+dy
	if nx < 0 || nx >= u5data.CombatSide || ny < 0 || ny >= u5data.CombatSide {
		// 走出邊緣 = 這名隊員離開戰場。
		s.Log(s.unitName(u) + "退出了戰場。")
		u.Flags |= UnitDead
		s.afterPlayerAction()
		return
	}
	if u5data.TileBlocksWalking(int(c.Map.At(nx, ny))) {
		s.Log(MsgBlocked)
		return
	}
	if _, taken := c.CombatUnitAt(nx, ny); taken {
		s.Log(MsgBlocked)
		return
	}
	u.X, u.Y = nx, ny
	s.afterPlayerAction()
}

// CombatAttack 讓目前輪到的隊員攻擊相鄰的敵人。
func (s *State) CombatAttack(d Direction) {
	c := s.Combat
	if c == nil || c.Over || c.Turn < 0 {
		return
	}
	u := &c.Units[c.Turn]
	dx, dy := d.Delta()
	target, ok := c.CombatUnitAt(u.X+dx, u.Y+dy)
	if !ok || s.hostile(target) == s.hostile(u) {
		s.Log("那個方向沒有敵人。")
		return
	}
	s.resolveAttack(c.Turn, s.unitIndex(target))
	s.afterPlayerAction()
}

// unitIndex 回傳某個單位在 Units 裡的槽號。
func (s *State) unitIndex(u *Combatant) int {
	c := s.Combat
	for i := range c.Units {
		if &c.Units[i] == u {
			return i
		}
	}
	return -1
}

// afterPlayerAction 玩家動完之後接回排程。
//
// 時間停止的倒數在這裡走 —— 原版 `sub_16370` 是在**玩家單位的回合結束時**
// 遞減 `byte_3E09E`,所以 An Tym 是「十個玩家回合」而不是「十分鐘」。
func (s *State) afterPlayerAction() {
	if s.TimeStop > 0 {
		s.TimeStop--
		if s.TimeStop == 0 {
			s.Log("時間又開始流動了。")
		}
	}
	if s.checkCombatOver() {
		return
	}
	s.advanceCombat()
}

// CombatPass 讓目前輪到的單位這一回合什麼都不做。
//
// 原版是空白鍵(玩家指令表 `jpt_A5C8` 的 case 32,印「Pass」)。
// 四面被自己人堵住時這是唯一的出路,少了它排程會停在那個人身上不動。
func (s *State) CombatPass() {
	c := s.Combat
	if c == nil {
		return
	}
	if c.Over {
		s.EndCombat(c.Won)
		return
	}
	if c.Turn < 0 {
		return
	}
	s.Log(s.unitName(&c.Units[c.Turn]) + "按兵不動。")
	s.afterPlayerAction()
}

// CombatFlee 讓整隊撤離。
//
// ⚠ 原版**沒有這個指令** —— 撤離是一格一格走出戰場邊緣。這裡保留一個
// 快捷鍵是為了不讓玩家卡在半完成的戰鬥系統裡,訊息也照實說明。
func (s *State) CombatFlee() {
	if s.Combat == nil {
		return
	}
	if s.Combat.Over {
		s.EndCombat(s.Combat.Won)
		return
	}
	s.Log("汝撤離了戰場。(原版沒有撤離鍵 —— 要一步步走出戰場邊緣)")
	s.EndCombat(false)
}

// 命中判定
//
// 原版 `sub_B484`:
//
//	threshold = (目標的防禦 + 30 − 攻擊者的攻擊) / 2
//	roll      = max(1, random(0, 60) / 2)        ← sub_2B724,值域 1..30
//	命中 ⟺ roll >= threshold
//
// 防禦高 → threshold 高 → 難命中;攻擊高 → threshold 低 → 易命中。
//
// 「攻擊」取哪一項由武器決定(`sub_B398`):**鈍器**(類別 8:釘盔、釘盾、
// 棍、釘頭錘、雙手錘)看**力量**,其餘看裝備的防禦加總。
// 幾種武器(0x23、0x27、0x28)則**必中**,不擲骰。

// AlwaysHitWeapons 是不用擲骰的武器編號(原版 `sub_B484` 的三個 cmp)。
var AlwaysHitWeapons = map[byte]bool{0x23: true, 0x27: true, 0x28: true}

// AttackRoll 是命中骰,值域 1..30(原版 `sub_2B724`)。
//
//	r = random(0, 60) / 2      ← 閉區間,所以是 0..30
//	if r == 0 { r = 1 }
func (s *State) AttackRoll() int {
	r := s.Roll(0, 60) / 2
	if r == 0 {
		r = 1
	}
	return r
}

// attackValueOf 取一名角色的「攻擊」那一項(原版 `sub_B398` 對角色的分支)。
func (s *State) attackValueOf(c *u5data.Character, weapon byte) int {
	if s.Stats == nil || c == nil {
		return 0
	}
	if s.Stats.IsBlunt(weapon) {
		return int(c.Strength)
	}
	return s.Stats.DefenceOf(c)
}

// HitChance 回報這一擊的門檻值(threshold)。roll >= 門檻就命中。
//
// 門檻 ≤ 1 代表必中(骰子最小是 1),≥ 31 代表必不中(骰子最大是 30)。
func HitThreshold(defence, attack int) int {
	return (defence + 30 - attack) / 2
}

// AttackerHits 判定一名隊員用某件武器打某個防禦值的目標會不會命中。
func (s *State) AttackerHits(c *u5data.Character, weapon byte, targetDefence int) bool {
	if AlwaysHitWeapons[weapon] {
		return true
	}
	return s.AttackRoll() >= HitThreshold(targetDefence, s.attackValueOf(c, weapon))
}

// 傷害計算
//
// 原版 `sub_B274(攻擊者, 目標)`:
//
//	攻擊者是怪物 → base = 怪物屬性的 +4(攻擊力)
//	攻擊者是角色 → 依武器:
//	    玻璃劍(0x27)  → 99,而且**劍會碎掉**(印「Thy sword hath shattered!」)
//	    寶石劍(0x28)  → 0
//	    空手(0xFF)    → 1
//	    其餘           → base = 武器傷害表[武器]
//	                     base > 1 且 != 99 時擲 random(1, base)
//
//	base == 99 → 傷害 99(必殺,不扣防禦)
//	否則:
//	    目標是怪物 → 減 random(1, 怪物屬性的 +3)
//	    目標是角色 → 減 random(1, 角色紀錄的 0x18)
//	    減值為 0 時不扣
//
// 注意兩處**閉區間**的 random(1, n) —— 與命中骰同一個 `sub_28E14`。

// DamageResult 是一次傷害計算的結果。
type DamageResult struct {
	// Amount 是最後扣的血;可能是負數(防禦擲得比傷害高),原版不夾。
	Amount int
	// Shattered 為真時武器碎了(玻璃劍)。
	Shattered bool
}

// WeaponDamageRoll 算攻擊者這一擊的基礎傷害(還沒扣目標的防禦)。
func (s *State) WeaponDamageRoll(weapon byte) (base int, shattered bool) {
	switch weapon {
	case u5data.ItemGlassSword:
		return u5data.InstantKillDamage, true
	case u5data.ItemJeweledSword:
		return 0, false
	case u5data.ItemNone:
		return u5data.BareHandDamage, false
	}
	if s.Stats == nil || int(weapon) >= u5data.ItemCount {
		return 0, false
	}
	base = s.Stats.ItemDamage[weapon]
	if base > 1 && base != u5data.InstantKillDamage {
		base = s.Roll(1, base)
	}
	return base, false
}

// DamageToCharacter 算一名隊員被打時實際掉多少血。
func (s *State) DamageToCharacter(base int, c *u5data.Character) int {
	return s.applyResist(base, int(c.Raw[u5data.CharDamageResist]))
}

// DamageToCreature 算一隻怪物被打時實際掉多少血。
func (s *State) DamageToCreature(base int, creature byte) int {
	resist := 0
	if s.Stats != nil {
		if st, ok := s.Stats.StatsFor(creature); ok {
			resist = int(st.Armour)
		}
	}
	return s.applyResist(base, resist)
}

// applyResist 套用減傷。必殺(99)不扣;減傷值 0 也不擲骰。
func (s *State) applyResist(base, resist int) int {
	if base == u5data.InstantKillDamage {
		return u5data.InstantKillDamage
	}
	if resist == 0 {
		return base
	}
	return base - s.Roll(1, resist)
}
