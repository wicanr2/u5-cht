package game

import (
	"strconv"

	"github.com/wicanr2/u5-cht/internal/i18n"
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
	// UnitGrabbed 被**拖屍怪(Corpser)拖到水下**了。
	//
	// ⚠ **更正**:這個旗標此前叫 `UnitGrabbed`,註解寫「這一回合不能行動」——
	// 那是從「它會讓單位跳過回合」反推出來的名字,不是它的語意。
	// 真正的來源是 `sub_1F840`:攻擊者的生物編號是 **0x2D(Corpser)**
	// 而目標是隊員時,印 `" dragged under!"`、設這個旗標、把顯示 tile 設成 0
	// (所以那個人**看起來消失了**)。
	//
	// 之後每一次輪到他:`sub_A360` 印 `ARGH!` 並呼叫 `sub_BCC4` 擲掙脫
	//(敏捷 > `max(1, rand(0,60)/2)` → 印 `" regurgitated!"` 並清旗標)。
	// 不論掙不掙脫,**這一回合都用掉了**。見 `docs/re/67`。
	UnitGrabbed = 0x04
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
	// Mode 是「怎麼進到這場戰鬥的」(原版 `byte_3E0B1`,由 `sub_2E364` 的第一個
	// 參數設定)。目前只用到 bit 0x80(不能用 ESC 離開),其餘位元是佈陣的變體:
	//
	//	2    地牢遊蕩怪物        4    地表紮營
	//	6    地牢紮營(2|4)     0x82 ★ 地牢房間 —— 唯一設 0x80 的地方
	//
	// 全檔只有 `sub_42CC`(「Entering room...」)一處寫 0x82,所以
	// **「離不開」是地牢房間獨有的**。見 `docs/re/73`。
	Mode byte
	// LastAttacker[槽] 是上一個攻擊這一槽的單位(原版 `byte_3E0B8`);
	// −1 代表沒有。施法被打斷的判定只看這個人。
	LastAttacker [CombatUnitSlots]int8
	// fromSlot 是觸發戰鬥的那個物件槽,打完要清掉。
	fromSlot int
	// scan 是排程掃到第幾槽,actions 是累計行動數(每 10 次 = 1 分鐘)。
	scan, actions int
	// LadderBoth 記著腳下那格地牢地形是不是**兩向梯**(原版 `byte_418DE`)——
	// 戰鬥中按 K 要不要問「上還是下」看它。
	LadderBoth bool
	// ArenaMode 是原版的 `byte_3E0B1`:0 一般遭遇 / 2 地牢遊蕩 / 4 地表紮營 / 6 地牢紮營。
	// 目前只有「全隊同一出口」那條規則讀它(`sameExitEnforced`)。
	ArenaMode int
	// ExitDir 是第一個離場的人選的出口(0 = 還沒有人離場)。
	ExitDir int
	// Left 是已經離場的單位數。
	Left int
	// savedX / savedY 是進戰鬥前的世界座標。
	//
	// 戰鬥時 `State.X/Y` 借給行動中的單位當戰場座標(見 `focusCombatUnit`),
	// 離場時還原 —— 原版 `sub_2E364` 的 var_4 / var_8 就是這個用途。
	savedX, savedY int
}

// InCombat 回報是不是正在戰鬥。
func (s *State) InCombat() bool { return s.Combat != nil }

// enemyDisplayName 照 `sub_2E58C`:種類碼 < 0x40 的一律叫 PIRATES,
// 其餘查生物名表(索引 = (種類 − 64) / 4,與 docs/re/09 同一條公式)。
func (s *State) enemyDisplayName(kind byte) string {
	if kind < u5data.CreatureBase {
		return i18n.Name("Pirates")
	}
	if n := s.Creatures.Name(kind); n != "" {
		return i18n.Name(n)
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
	return s.beginCombatFrom(o, slot)
}

// beginCombatWith 用一個**合成的**物件開打(場景裡的 NPC 走這條,見 scenecombat.go)。
//
// 場景 NPC 不佔物件槽,所以沒有可以回填的來源槽 —— 傳 −1。
func (s *State) beginCombatWith(o *u5data.MapObject) bool {
	if s.CombatMaps == nil {
		return false
	}
	return s.beginCombatFrom(o, -1)
}

// beginCombatFrom 是兩條路共用的本體。
func (s *State) beginCombatFrom(o *u5data.MapObject, slot int) bool {
	kind := o.Kind &^ 0x03
	terrain := int(s.TileAt(o.X, o.Y))
	idx := u5data.SelectCombatMap(int(kind), terrain, s.Transport, !s.InScene())
	if idx < 0 || idx >= len(s.CombatMaps.Maps) {
		return false
	}
	m := &s.CombatMaps.Maps[idx]

	c := &Combat{Map: m, MapIndex: idx, fromSlot: slot, Turn: -1,
		savedX: s.X, savedY: s.Y,
		EnemyName: s.enemyDisplayName(o.Kind)}
	for i := range c.LastAttacker {
		c.LastAttacker[i] = -1
	}
	// 隊員照圖裡的入場位置排;人數不足就只排前 n 個。
	s.placeParty(c, m)
	s.spawnEnemies(c, o)

	s.Combat = c
	s.Prompt = PromptCombat
	// 原版 `sub_A9EC`(戰鬥迴圈)一進來就換曲 3(`docs/re/87`)。
	s.playSong(SongCombat)
	s.Log("「" + c.EnemyName + "」來襲!")
	s.Log("(戰場 #" + strconv.Itoa(idx) + ")")
	s.advanceCombat()
	return true
}

// beginRoomCombat 用一張現成的地圖開一場戰鬥(地牢房間走這條)。
//
// 與撞上怪物那條的差別只在**敵人從哪來**:房間的怪物寫在地圖自己的
// `EnemyKind` 裡(檔案位移 171),不是由撞到的物件決定。
func (s *State) beginRoomCombat(m *u5data.CombatMap, idx int, mode byte) bool {
	// ⚠ **mode 一定要由呼叫端給**:這一支被三種場合共用(地牢房間 0x82、
	// 地牢遊蕩怪物 2、地牢紮營 6),而**只有房間那個帶 0x80**(離不開)。
	// 寫死成房間模式會讓遊蕩怪物與紮營也變成離不開的死戰。
	c := &Combat{Map: m, MapIndex: idx, fromSlot: -1, Turn: -1,
		savedX: s.X, savedY: s.Y, EnemyName: "房間裡的東西", Mode: mode}
	// 地牢房間 / 遊蕩怪 / 紮營三種場合都走 `sub_A9EC`,所以一樣換曲 3。
	s.playSong(SongCombat)
	for i := range c.LastAttacker {
		c.LastAttacker[i] = -1
	}
	s.placeParty(c, m)
	for n, kind := range m.EnemyKind {
		if kind < u5data.CreatureBase {
			continue
		}
		cre, ok := s.creatureIndexOf(kind)
		if !ok {
			continue
		}
		s.placeEnemy(c, n, kind, cre)
		if c.EnemyName == "房間裡的東西" {
			c.EnemyName = s.enemyDisplayName(kind)
		}
	}
	s.Combat = c
	s.Prompt = PromptCombat
	s.Log("「" + c.EnemyName + "」!")
	s.advanceCombat()
	return true
}

// placeParty 把隊員排到圖裡的入場位置。
//
// ⚠ 原版的佈陣函式 `sub_2EE84` 裡**夾了一段與佈陣無關的判定**:
// 開戰時隱形戒指與再生戒指有 1/16 會消失。那一段在 `upkeep.go` 的
// `vanishRings`,由這裡呼叫 —— 位置照原版(在排位置的同一趟)。
func (s *State) placeParty(c *Combat, m *u5data.CombatMap) {
	s.vanishRings()
	for i, ch := range s.Party() {
		if i >= u5data.CombatPartySlots {
			break
		}
		u := &c.Units[i]
		*u = Combatant{
			Roster:   i,
			Creature: -1,
			// ⚠ **不加 `NPCTileBase`**:0x1D / 0x1E 是前 256 格(地形與物件那一頁)
			// 裡的直接 tile 號,不是生物編號。加了 256 會畫成某隻怪物。
			Tile: int(partyTileFor(ch)),
			X:        int(m.PartyX[i]),
			Y:        int(m.PartyY[i]),
			Dex:      int(ch.Dex),
			Flags:    UnitParty,
		}
		u.Init = u.resetInit()
		if ch.Status == u5data.StatusDead {
			u.Flags |= UnitDead
		}
		if wieldsSwordOfChaos(ch) {
			u.Flags |= UnitSideFlip
		}
	}
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
		return i18n.Name(ch.Name)
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
		if u.Flags&UnitGrabbed != 0 {
			// 被拖屍怪拖到水下:印 ARGH! 並擲一次掙脫,這一回合照樣用掉。
			s.strugglingUnderwater(u)
			continue
		}
		if u.Flags&UnitAsleep != 0 {
			// ★ 醒來的機率**兩條路不一樣**,而且原版就是兩支不同的函式:
			//
			//	怪物 / AI(`sub_A108`)  rand(0, 16) == 16   → 1/17
			//	隊員  (`sub_A360`)      rand(0, 255) < 16   → 1/16
			//
			// 而隊員那條**不論醒不醒都印 "Zzzzz..." 並用掉這一回合** ——
			// 少了那句話,玩家只會看到自己的角色莫名其妙不動(`docs/re/67`)。
			if s.playerControlled(u) {
				if s.Roll(0, 255) < 16 {
					u.Flags &^= UnitAsleep
				}
				s.Log(s.unitName(u) + MsgZzzzz)
			} else if s.Roll(0, 16) == 16 {
				u.Flags &^= UnitAsleep
				s.Log(s.unitName(u) + "醒了。")
			}
			continue
		}
		if s.playerControlled(u) {
			c.Turn = i
			s.focusCombatUnit(u)
			// 原版每個隊員的回合都以「<名字>, armed with …:」開場。
			s.announceCombatTurn(u)
			return
		}
		s.aiTurn(i)
		if s.checkCombatOver() {
			return
		}
	}
}

// focusCombatUnit 把「隊伍座標」指到目前行動的那個單位身上。
//
// ★ 原版 `sub_A360` 的**第一件事**就是這個:
//
//	movzx eax, byte_3E0AE                    ; 目前行動的單位
//	mov   dl, byte ptr dword_3EF54+2[eax*8]
//	mov   byte_3E0A6, dl                     ; 「隊伍 X」← 該單位在戰場上的 X
//	mov   dl, byte ptr dword_3EF54+3[eax*8]
//	mov   byte_3E0A7, dl
//
// 也就是說原版**只有一對座標**,戰鬥時把它借給行動中的單位用。
// 這正是 Get / Jimmy / Open / Push / Search 那幾支不用改就能在戰場上
// 運作的原因 —— 它們讀的一直是同一對全域座標。
//
// 進戰鬥前的世界座標由 `Combat.savedX/savedY` 收著,`EndCombat` 還回去
//(原版 `sub_2E364` 的 var_4 / var_8 也是這樣存還)。
func (s *State) focusCombatUnit(u *Combatant) {
	if s.Combat == nil {
		return
	}
	s.X, s.Y = u.X, u.Y
}

// checkCombatOver 判勝負(原版 `sub_A9EC` 每個單位行動完都查一次)。
func (s *State) checkCombatOver() bool {
	c := s.Combat
	enemies, party := s.sideCounts(c)
	switch {
	case enemies == 0:
		s.Log("勝利!")
		// 原版 `sub_A9EC` 印完 `"\nVICTORY!\n"` 才換曲 0(`docs/re/87`)。
		s.playSong(SongVictory)
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

// 隊員在戰場上的兩個 tile(原版直接寫進物件記錄的 byte +1)。
//
// ★ **2026-08-08 定案,有直接證據**(`docs/re/53`):
//
//	sub_2EDF8(躺下):mov byte ptr dword_3E46C+1[eax*8], 1Eh
//	sub_2ED50(起身):mov byte ptr dword_3E46C+1[eax*8], 1Dh
//
// 兩支都是「把這個戰場單位對應的物件記錄的圖換掉」,而且成對出現 ——
// 站著 0x1D、躺著 0x1E。0x1C 是世界地圖上步行的隊伍,而 `sub_16DA4` 判
// 「在步行嗎」收的是 **0x1C 或 0x1D 兩個值**,獨立佐證 0x1D 屬於同一族。
//
// ⚠⚠ 先前寫的 0x4C 是**猜的**,而且撞到了別的語意:`sub_16058`(戰鬥中的
// Klimb)判「爬得過去」用的就是 tile 0x4C。一格不會同時是「隊伍自己」與
// 「戰場上爬得過去的東西」—— 那個矛盾是我自己造出來的。
const (
	// PartyTileStanding 是站著的隊員。
	PartyTileStanding = 0x1D
	// PartyTileLying 是睡著 / 倒下的隊員。
	PartyTileLying = 0x1E
)

// partyTileFor 挑隊員在戰場上的圖。
//
// ⚠⚠ **更正(2026-08-08,`docs/re/72`)**:此前這裡寫「**與職業無關**,原版就
// 只有站著與躺著兩個值」—— **反了**。開戰佈陣的迴圈(`sub_C414+1D0`)寫的是
// `byte_40C34[職業字母在 "AMBFDTPRS" 裡的位置]` = 法師 0x40 / 吟遊詩人 0x44 /
// 戰士 0x48 / 其餘五種 0x4C(見 `u5data.PartyCombatTile`)。
//
// 0x1D / 0x1E 只出現在**恢復路徑**上:`sub_2ED50` 醒來寫 0x1D、`sub_2EDF8`
// 躺下寫 0x1E、戴上隱形戒指寫 0x1D。當初只讀了那兩支寫 `+1` 的函式,
// 沒有全檔掃「還有誰寫這個欄位」,於是把「我看到的兩個值」寫成「原版只有兩個值」。
//
// ★ 而 0x4C 正是 `docs/re/52` §5 記下的那個矛盾裡被「修掉」的值 ——
// 那個猜是對的:`sub_16058`(戰鬥中的 Klimb)判「爬得過去」用 0x4C,
// 正因為那一格站著隊員。
func partyTileFor(ch *u5data.Character) byte {
	if ch != nil && (ch.Status == u5data.StatusAsleep || ch.Status == u5data.StatusDead) {
		return PartyTileLying
	}
	return u5data.PartyCombatTile(ch)
}

// CombatTileAt 回傳戰場上 (x, y) 該顯示什麼。
func (s *State) CombatTileAt(x, y int) byte {
	if s.Combat == nil {
		return u5data.TileBlank
	}
	return s.Combat.Map.At(x, y)
}

// SetCombatTileAt 改寫戰場上的一格(撬開的門、燒掉的東西、推走的家具)。
//
// ⚠ 改的是 `Combat.Map` 這份**副本**。它是 `BuildDungeonArena` 現畫的,
// 或是從 `.CBT` 讀進來的值傳副本 —— 不會寫回原版檔案(CLAUDE.md §3.2)。
func (s *State) SetCombatTileAt(x, y int, tile byte) bool {
	if s.Combat == nil || s.Combat.Map == nil {
		return false
	}
	if x < 0 || x >= u5data.CombatSide || y < 0 || y >= u5data.CombatSide {
		return false
	}
	s.Combat.Map.Tiles[y][x] = tile
	s.Combat.Map.Raw[y*u5data.CombatRowStride+x] = tile
	return true
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
	// 把進戰鬥前的世界座標還回去(原版 `sub_2E364` 尾端把 var_4 / var_8
	// 寫回 `byte_3E0A6/A7`)。少了這一步,打完架人會出現在戰場的格座標上。
	s.X, s.Y = s.Combat.savedX, s.Combat.savedY
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

// CombatAttack 讓目前輪到的隊員往某個方向攻擊。
//
// **遠程武器往那個方向射出去**,由投射物的飛行決定打到誰
//(原版 `sub_1FA6C` 瞄準 → `sub_20134` → `sub_1FE54`)。近戰只打相鄰那一格。
func (s *State) CombatAttack(d Direction) {
	c := s.Combat
	if c == nil || c.Over || c.Turn < 0 {
		return
	}
	u := &c.Units[c.Turn]
	dx, dy := d.Delta()

	reach := 1
	weapon := byte(u5data.ItemNone)
	if ch := s.charOf(u); ch != nil {
		weapon = ch.Equipment().Weapon
		if r := s.Stats.ItemRange[weapon]; r > 0 {
			reach = r
		}
	}
	if reach > 1 {
		// 遠程:朝那個方向射到射程盡頭,路上第一個擋下來的就是目標。
		victim, _, _ := s.FlyProjectile(c.Turn, u.X+dx*reach, u.Y+dy*reach)
		if victim < 0 {
			s.Log("射空了。")
		} else {
			s.resolveAttack(c.Turn, victim)
		}
		s.afterPlayerAction()
		return
	}
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
	// ★ 原版 `sub_16370` 的三件事,順序照抄:戒指效果 → 力場消散 → 模式倒數。
	s.ringUpkeep()
	s.expireFields()
	s.tickCombatMode()
	if s.TimeStop > 0 {
		s.TimeStop--
		if s.TimeStop == 0 {
			s.Log("時間又開始流動了。")
		}
	}
	if s.checkCombatOver() {
		return
	}
	// 原版 `sub_A360` 在每個單位行動完之後就查一次(`call sub_161E4`)。
	if s.checkAbsorbed() {
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
func (s *State) CombatFlee() bool {
	c := s.Combat
	if c == nil {
		return false
	}
	s.Log(MsgEscape)
	// ★ 兩道閘門**只在還有活著的隊員時才檢查**。全隊倒下時 ESC 一定放行 ——
	// 那是「打輸了離場」那條路,不能被「勝負未定」卡住。
	if s.anyLivingPartyUnit() {
		if c.Mode&CombatNoEscape != 0 {
			s.Log(MsgEscapeNotHere)
			return false
		}
		if !c.Over {
			s.Log(MsgEscapeNotYet)
			return false
		}
	}
	// 原版接著把場上**所有單位與所有物件**逐一移除(`sub_B210(−槽−1)` 與
	// `sub_B210(槽+1)`)再離場。引擎的 `EndCombat` 直接丟掉整個 `Combat`,
	// 等價 —— 但**戰勝時要清掉地圖上那隻怪**這件事得照 `Won` 走。
	s.EndCombat(c.Over && c.Won)
	return true
}

// anyLivingPartyUnit 是原版那個掃 32 槽的迴圈:`(flags & 0A0h) == 80h`
// —— 隊員(0x80)而且沒死(0x20)。
//
// ⚠ 遮罩是 0xA0 不是 0x80:少了 `UnitDead` 那一位,全隊倒下之後
// ESC 會被「勝負未定」擋住,玩家就卡在戰場上了。
func (s *State) anyLivingPartyUnit() bool {
	c := s.Combat
	if c == nil {
		return false
	}
	for i := range c.Units {
		if c.Units[i].Flags&(UnitParty|UnitDead) == UnitParty {
			return true
		}
	}
	return false
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

// 戰鬥回合的開場提示(原版 `sub_A360` 開頭 + `sub_A310`)
//
// 原版每一個**隊員**的回合都以這一行開場:
//
//	<名字>, armed with <武器…>:        或        <名字>, armed with bare hands:
//
// 引擎原本什麼都不印,`focusCombatUnit` 只把鏡頭移過去 —— 玩家得自己記得
// 現在輪到誰、手上拿的是什麼。
//
// # `sub_A310(裝備編號, 要不要先加分隔)`
//
//	if (編號 == 0FFh)          return 0        ; 空手
//	if (byte_3F290[編號] == 0) return 0        ; ★ 傷害是 0 → 不算武器
//	if (要不要先加分隔)  strcat(緩衝, ", ")
//	strcat(緩衝, 物品名表[編號])
//	return 1
//
// ⚠ 分隔字串在 IDA 裡被讀成 `off_48A88 dd offset loc_202C` —— 那是**誤判**:
// 呼叫端推的是 `offset off_48A88`(那四個位元組本身),而 0x0000202C 的
// 小端位元組就是 `2C 20 00 00` = **`", "`**。IDA 把內嵌字面值當成指標了。
//
// # 三個欄位,不是六個
//
// `sub_A360` 只問三格:`byte_3DDCD` / `byte_3DDCF` / `byte_3DDD0`,
// 以名冊基底 `byte_3DDB4` 回推是**偏移 0x19 / 0x1B / 0x1C** ——
// 頭盔、右手、左手。護甲(0x1A)、戒指(0x1D)、護符(0x1E)不問。
//
// (基底可以獨立驗:`byte_3DDBF` − `byte_3DDB4` = 0x0B = `CharStatus` ✓)
//
// 三次都回 0 才印 `bare hands`。所以「傷害 0」的東西一律不列 ——
// ⚠ 這包含**寶石劍**(`ItemJeweledSword`,傷害被特例設成 0),
// 拿著它的角色會被報成空手。那是原版行為,不是這裡算錯。

// combatArmSlots 是開場提示會問的三個裝備欄位(原版只問這三格)。
var combatArmSlots = [3]int{u5data.CharHelm, u5data.CharWeapon, u5data.CharShield}

// armedWith 組出「, armed with …」後面那一段(不含前綴與結尾的冒號)。
func (s *State) armedWith(ch *u5data.Character) string {
	if ch == nil || s.Stats == nil {
		return MsgBareHands
	}
	var out string
	for _, slot := range combatArmSlots {
		id := int(ch.Raw[slot])
		if id == int(u5data.ItemNone) || id >= len(s.Stats.ItemDamage) {
			continue
		}
		// ★ 判準是「這件東西有傷害」,不是「這一格是武器欄」。
		if s.Stats.ItemDamage[id] == 0 {
			continue
		}
		if out != "" {
			out += MsgArmedSeparator
		}
		out += s.equipName(id)
	}
	if out == "" {
		return MsgBareHands
	}
	return out
}

// announceCombatTurn 印出隊員回合的開場提示。
func (s *State) announceCombatTurn(u *Combatant) {
	ch := s.charOf(u)
	if ch == nil {
		return
	}
	s.Log(s.unitName(u) + MsgArmedWith + s.armedWith(ch) + MsgArmedColon)
}
