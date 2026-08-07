// Package game 是遊戲狀態機:大地圖與場景之間的進出、移動、樓層、通行判定。
//
// 這一層**不畫任何像素、不依賴 ebiten**,所以整段劇本(走到城門 → 進城 → 上樓 →
// 走出邊界 → 回到大地圖)可以在單元測試裡跑完,不需要顯示環境。
// 這是 internal/render 那個「headless 驗證不該需要 GPU」決策的延伸:
// 邏輯與畫面各自可獨立驗證。
//
// 規則一律照原版執行檔,不自創(CLAUDE.md §3.0)。主要來源:
//
//	sub_86C  場景內移動 —— 邊界偵測、離開詢問、樓梯、阻擋(docs/re/03 §7)
//	sub_758  樓梯升降 —— tile & 0xFC == 0xC4,朝向決定上或下
//	sub_5C8  場景載入 —— 地點編號 + 樓層 → 哪一張 32×32 地圖
//	sub_10928 進入場景 —— 比對地點表座標,設 (15, 30) 為入口
package game

import (
	"math/rand"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// Direction 是原版內部使用的朝向碼(sub_86C 把按鍵轉成這組值後才往下傳)。
//
// 原版的按鍵碼是 1=西 2=東 3=北 4=南,進函式後立刻映射成下面這組 —— 之所以要換,
// 是因為這組的**反向 = XOR 2**,樓梯判定(sub_758)直接用到這個性質。
type Direction int

const (
	North Direction = 0
	East  Direction = 1
	South Direction = 2
	West  Direction = 3
)

// Opposite 回傳反向。原版就是 XOR 2 —— 不是查表。
func (d Direction) Opposite() Direction { return d ^ 2 }

// Delta 回傳這個方向的座標增量。
func (d Direction) Delta() (dx, dy int) {
	switch d {
	case North:
		return 0, -1
	case East:
		return 1, 0
	case South:
		return 0, 1
	case West:
		return -1, 0
	}
	return 0, 0
}

// TurnLeft / TurnRight 是第一人稱轉向。方向碼順時針排(北 0 東 1 南 2 西 3),
// 所以右轉是 +1、左轉是 −1,兩者都對 4 取模。
func (d Direction) TurnRight() Direction { return (d + 1) & 3 }

// TurnLeft 見 TurnRight。
func (d Direction) TurnLeft() Direction { return (d + 3) & 3 }

// Name 回傳中文方位(訊息欄用)。
func (d Direction) Name() string {
	switch d {
	case North:
		return "北"
	case East:
		return "東"
	case South:
		return "南"
	case West:
		return "西"
	}
	return "?"
}

// 場景入口座標。原版 sub_10928 進場景時固定設 (15, 30):32 格寬的中央、靠底部,
// 也就是城鎮的南門。
const (
	SceneEntryX = 15
	SceneEntryY = 30
)

// UnderworldLocation 是唯一一個「離開後回到地下世界而非地表」的地點編號。
// 原版 sub_86C:`if (byte_3E0A3 == 0x19)` → 印 "Underworld!"、樓層設 -1、BGM 換 10 號曲。
const UnderworldLocation = 25 // ARARAT

// Prompt 是等待玩家回答的提問。原版在場景邊界會問 "Dost thou wish to leave?",
// 只收 Y / N / ESC。
type Prompt int

const (
	// PromptNone 表示沒有待回答的提問。
	PromptNone Prompt = iota
	// PromptLeave 是「是否離開此地?」
	PromptLeave
	// PromptTalk 是對話中:鍵盤輸入的是關鍵字,不是指令鍵。
	PromptTalk
	// PromptAnswer 是 NPC 反問之後:輸入的是對那個問題的回答。
	PromptAnswer
	// PromptShop 是在店裡:按鍵是選單字母或 Y/N,不是指令鍵。
	PromptShop
	// PromptCombat 是戰鬥中:方向鍵移動、A 攻擊、C 施法、空白 Pass、Esc 撤離。
	PromptCombat
	// PromptSpell 是在打咒語名:輸入的是上古語,不是指令鍵。
	PromptSpell
)

// State 是一局遊戲的位置狀態。
type State struct {
	World  *u5data.WorldMap  // 地表(BRIT.DAT)
	Under  *u5data.WorldMap  // 地下世界(UNDER.DAT);可為 nil
	Scenes *u5data.SceneSet  // 城鎮 / 城堡 / 民居 / 要塞;可為 nil
	NPCs   *u5data.NPCSet    // 各地點的居民與排程;可為 nil
	Talks  *u5data.TalkSet   // 對話文字;可為 nil
	Shops  *u5data.ShopSet   // 商店目錄與對白;可為 nil
	Items  *u5data.ItemTable // 裝備名字;可為 nil
	// Objects 是大地圖上「會動的東西」:隊伍自己、坐騎、船、地上的物品、遊蕩的怪物
	// (原版 dword_3E46C,來自 BRIT.OOL / UNDER.OOL)。地表與地下各一份。
	Objects, UnderObjects *u5data.ObjectSet

	// CombatMaps 是地表的戰鬥地圖(BRIT.CBT);可為 nil。
	CombatMaps *u5data.CombatMapSet
	// Creatures 是生物名表,戰鬥時報敵人名字用;可為 nil。
	Creatures *u5data.CreatureTable
	// Stats 是戰鬥數值(怪物三圍、裝備防禦 / 射程 / 類別);可為 nil。
	Stats *u5data.CombatStats
	// Spells 是咒語表(名稱 / 圈數 / 藥草 / 可施法場合);可為 nil。
	Spells *u5data.SpellTable
	// Dungeons 是八座地牢的地圖(DUNGEON.DAT);可為 nil。
	Dungeons *u5data.DungeonSet
	// DungeonRooms 是地牢房間(DUNGEON.CBT,格式同地表戰鬥地圖);可為 nil。
	DungeonRooms *u5data.CombatMapSet
	// Dungeon 是「正在某座地牢裡」的狀態;nil 代表不在地牢。
	Dungeon *DungeonState
	// LightTurns 是光明咒語還亮幾分鐘(原版 byte_3E0B6)。
	LightTurns int
	// TorchTurns 是火把還亮幾分鐘(原版 byte_3E0B7)。點一把是 random(0,15) + 0x70。
	TorchTurns int
	// TimeStop 是 An Tym 還停幾回合(原版 byte_3E09E;byte_3E08A == 'T' 期間)。
	TimeStop int
	// CombatMode 是全域的戰鬥模式(原版 byte_3E08A),CombatModeTurns 是剩幾回合。
	//
	// 四個咒語共用這一個位元組(`sub_1D31C(模式, 回合, 音效)`):
	// In Sanct 'P'、Rel Tym 'Q'、Quas An Wis 'C'、In An 'N'、An Tym 'T'。
	// 後來的會蓋掉前面的 —— 原版就是一個位元組,不是一組旗標。
	CombatMode      byte
	CombatModeTurns int
	// rng 是戰鬥骰子。留空時用固定種子 —— headless 與測試要可重現;
	// 遊戲啟動時 cmd/u5cht 換成時間種子。
	rng *rand.Rand
	// Combat 是進行中的戰鬥(Prompt == PromptCombat 時有效)。
	Combat *Combat
	// castReturn 是打完咒語名要回到哪個 Prompt,castBy 是誰在施法。
	castReturn Prompt
	castBy     int

	// BaseSave 是讀進來的那份存檔,存檔時當底稿用 —— 引擎還沒解出來的欄位
	// (魔法、任務旗標、地牢狀態…)靠它原樣保留。見 savegame.go。
	BaseSave *u5data.Save

	// Clock 是遊戲內時間。NPC 站在哪裡完全由它決定(排程以小時為單位)。
	Clock Clock

	// Location 是原版的 byte_3E0A3:0 代表在大地圖,1..32 是地點編號。
	Location int
	// Floor 是原版的 byte_3E0A5。在大地圖:0 = 地表、-1 = 地下世界;在場景:樓層。
	Floor int
	// X, Y 是原版的 byte_3E0A6 / byte_3E0A7。大地圖時是世界座標,場景時是場景座標
	// —— 原版共用同一對變數,離開場景時從地點表把世界座標讀回來。
	X, Y int

	// HasShip 與 DockX/DockY 是買到的船停在哪(原版 byte_3EE17 的旗標
	// 與 byte_3E165 / byte_3E166 的座標)。船不佔物件槽。
	HasShip      bool
	DockX, DockY int

	// ShipHull / ShipSkiffs 是**目前搭乘中**那艘船的耐久與載著的小艇數。
	// 原版把它們暫存在物件槽 0(隊伍那一槽)的 +5 / +7,下船時再搬回船物件。
	ShipHull, ShipSkiffs int

	// Transport 是原版的 byte_3E08C:隊伍當前的載具 tile(0 = 步行)。
	// 通行判定 sub_2A694 的第一個參數就是它(docs/re/02)。
	Transport byte

	// Prompt 是目前的輸入模式;非 PromptNone 時,移動輸入無效。
	Prompt Prompt
	// Conv 是進行中的對話(Prompt == PromptTalk 時有效)。
	Conv *u5data.Conversation
	// Shop 是進行中的交易(Prompt == PromptShop 時有效)。
	Shop *ShopSession
	// Inventory 是隊伍共用的背包(金幣、鑰匙、寶石、火把、裝備、藥草)。
	Inventory u5data.Inventory
	// Input 是對話中已經打進去的關鍵字。
	Input string
	// pending 是正在等玩家回答的提問區塊。
	pending *u5data.Question
	// Karma 是業報(0..99)。對話裡的 opcode 0x89/0x8A 會動到它。
	Karma int
	// Roster 是全部 16 名可用角色;隊伍就是名冊的前 PartySize 名
	// (原版 sub_1BB5C 讓人入隊的方式是把名冊裡的那一筆與隊伍位置**對調**,
	//  所以「隊伍」不是另一個清單,而是名冊的前綴)。
	Roster    []u5data.Character
	PartySize int

	Messages    []string
	MaxMessages int

	// talkingTo 是正在交談的 NPC 槽號;-1 代表沒有。入隊之後要靠它把人移出場景。
	talkingTo int

	// removed 記錄已經離場的 NPC(入隊之後原版會把他從場景移除)。
	// key 是 地點編號<<8 | 槽號 —— 換地點回來時他不該又站在原地。
	removed map[int]bool

	// sceneObjects 是場景裡的物件槽;進場景時清空(原版 sub_1678)。
	sceneObjects *u5data.ObjectSet

	// rtNPCs 是 NPC 的執行期狀態(原版 word_3E770)。位置由它決定,
	// 不是每次現算排程 —— NPC 是一步一步走過去的。
	rtNPCs []RuntimeNPC

	scene *u5data.SceneMap                    // 目前這一層的場景地圖快取
	npcs  *[u5data.NPCsPerLocation]u5data.NPC // 目前地點的 NPC 槽
}

// currentObjects 回傳目前該用哪一份物件表。
//
// 大地圖用 `BRIT.OOL` / `UNDER.OOL` 讀進來的那兩份;場景裡另有一份,
// 進場景時**整份清空**(原版 `sub_1678` 把槽 1..31 的種類碼歸零再載入場景),
// 所以在城裡買的馬離開就不見了 —— 那是原版行為,不是漏做。
func (s *State) currentObjects() *u5data.ObjectSet {
	if s.InScene() {
		if s.sceneObjects == nil {
			s.sceneObjects = &u5data.ObjectSet{}
		}
		return s.sceneObjects
	}
	if s.Floor < 0 {
		return s.UnderObjects
	}
	return s.Objects
}

// Objects 回傳目前這一層的物件表,讓引擎其他部分能放東西進去。
func (s *State) CurrentObjects() *u5data.ObjectSet { return s.currentObjects() }

// VisibleObjects 回傳此刻該畫出來的地圖物件。
//
// 隊伍自己那一槽不畫 —— 玩家的位置由 State.X/Y 決定,畫兩次只會疊在一起。
func (s *State) VisibleObjects() []VisibleObject {
	objs := s.currentObjects()
	if objs == nil {
		return nil
	}
	var out []VisibleObject
	for i := range objs.Objects {
		if i == u5data.PartyObjectSlot {
			continue
		}
		o := &objs.Objects[i]
		if !o.Present() || o.Floor != s.Floor {
			continue
		}
		out = append(out, VisibleObject{Slot: i, X: o.X, Y: o.Y, Tile: int(o.Tile), Object: o})
	}
	return out
}

// VisibleObject 是「此刻該畫在這一格的地圖物件」。
type VisibleObject struct {
	Slot   int
	X, Y   int
	Tile   int
	Object *u5data.MapObject
}

// ObjectAt 回報某一格上有沒有物件。
func (s *State) ObjectAt(x, y int) (*u5data.MapObject, int, bool) {
	objs := s.currentObjects()
	if objs == nil {
		return nil, 0, false
	}
	for i := range objs.Objects {
		if i == u5data.PartyObjectSlot {
			continue
		}
		o := &objs.Objects[i]
		if o.Present() && o.X == WrapWorld(x) && o.Y == WrapWorld(y) && o.Floor == s.Floor {
			return o, i, true
		}
	}
	return nil, 0, false
}

// VisibleNPC 是「此刻該畫在這一格的 NPC」。
type VisibleNPC struct {
	Index int // 槽號(0 是隊伍自己,不會出現在這裡)
	X, Y  int
	Tile  int // 已經換算過的 tile 索引(u5data.NPCTileBase + 生物編號)
	NPC   *u5data.NPC
}

// VisibleNPCs 回傳此時此層看得到的 NPC。
//
// 位置取自**執行期狀態**(原版 word_3E770),不是每次拿時鐘現算排程 ——
// NPC 是一步一步走向排程位置的,中途會停在路上。
func (s *State) VisibleNPCs() []VisibleNPC {
	if !s.InScene() || s.npcs == nil {
		return nil
	}
	if len(s.rtNPCs) != u5data.NPCsPerLocation {
		s.initRuntimeNPCs()
	}
	var out []VisibleNPC
	for i := range s.npcs {
		if i == u5data.PartySlot {
			continue // 0 號是隊伍自己
		}
		n := &s.npcs[i]
		if !n.Present() || s.removed[s.Location<<8|i] {
			continue
		}
		rt := &s.rtNPCs[i]
		if rt.Mode == ModeAbsent || rt.Floor != s.Floor {
			continue
		}
		if rt.X < 0 || rt.X >= u5data.SceneSide || rt.Y < 0 || rt.Y >= u5data.SceneSide {
			continue
		}
		out = append(out, VisibleNPC{Index: i, X: rt.X, Y: rt.Y, Tile: n.TileIndex(), NPC: n})
	}
	return out
}

// NPCAt 回報某一格上有沒有 NPC。
func (s *State) NPCAt(x, y int) (*VisibleNPC, bool) {
	for _, v := range s.VisibleNPCs() {
		if v.X == x && v.Y == y {
			vv := v
			return &vv, true
		}
	}
	return nil, false
}

// loadNPCs 換地點或換樓層之後重取 NPC 槽。
func (s *State) loadNPCs() {
	s.npcs = nil
	if s.NPCs == nil || !s.InScene() {
		return
	}
	if n, err := s.NPCs.At(s.Location); err == nil {
		s.npcs = n
	}
}

// tick 推進遊戲時間,然後讓 NPC 走一回合。
//
// 原版一般行動每回合 1 分鐘(sub_1DC8 → sub_29304(1)),之後才輪到
// `sub_9690` 讓 NPC 動。順序不能反 —— NPC 的模式判定要看新的小時。
func (s *State) tick() {
	s.AdvanceTime(MinutesPerTurn)
	s.advanceNPCs()
}

// InScene 回報玩家是否在場景(城鎮 / 城堡 / 民居 / 要塞)裡。
func (s *State) InScene() bool { return s.Location != 0 }

// Location 名稱 —— 在大地圖時回傳空字串。
func (s *State) LocationName() string {
	loc, err := u5data.LocationByNumber(s.Location)
	if err != nil {
		return ""
	}
	return loc.DisplayName()
}

// Log 加一條訊息到訊息欄。
func (s *State) Log(msg string) {
	if s.MaxMessages <= 0 {
		s.MaxMessages = 2
	}
	s.Messages = append(s.Messages, msg)
	if len(s.Messages) > s.MaxMessages {
		s.Messages = s.Messages[len(s.Messages)-s.MaxMessages:]
	}
}

// currentWorld 回傳目前所在的大地圖(地表或地下世界)。
func (s *State) currentWorld() *u5data.WorldMap {
	if s.Floor < 0 && s.Under != nil {
		return s.Under
	}
	return s.World
}

// TileAt 回報 (x, y) 在**視窗上**該顯示什麼 tile。
//
// 大地圖會環繞(不列顛尼亞是 wrap-around);場景則在邊界外回傳 TileBlank ——
// 原版的 11×11 視窗緩衝就是先 memset 成 0xFF 再填入地圖內容,所以城鎮外緣
// 看到的是黑格,不是別的地形。
func (s *State) TileAt(x, y int) byte {
	if s.InScene() {
		if s.scene == nil || x < 0 || x >= u5data.SceneSide || y < 0 || y >= u5data.SceneSide {
			return u5data.TileBlank
		}
		return s.scene.At(x, y)
	}
	w := s.currentWorld()
	if w == nil {
		return 0
	}
	return w.At(WrapWorld(x), WrapWorld(y))
}

// WrapWorld 把座標折回大地圖範圍(不列顛尼亞是環繞的)。
func WrapWorld(v int) int {
	v %= u5data.WorldSide
	if v < 0 {
		v += u5data.WorldSide
	}
	return v
}

// Move 往 d 走一步。
//
// 場景內的完整流程照 sub_86C:先判斷是不是踩到邊界(x<1 / x>30 / y<1 / y>30),
// 是的話問「是否離開」;否則檢查通行,能走就走,並看看踩到的是不是樓梯。
func (s *State) Move(d Direction) {
	if s.Prompt != PromptNone {
		return // 有提問待答時,移動輸入無效(原版是 do-while 只收 Y/N/ESC)
	}
	if s.InScene() {
		s.moveInScene(d)
		return
	}
	s.moveInWorld(d)
}

func (s *State) moveInWorld(d Direction) {
	dx, dy := d.Delta()
	nx, ny := WrapWorld(s.X+dx), WrapWorld(s.Y+dy)
	// 撞上怪物就開打(原版 sub_2E58C 的入口)。
	if o, slot, ok := s.ObjectAt(nx, ny); ok && o.IsCreature() {
		if s.BeginCombat(slot) {
			return
		}
	}
	if u5data.TileBlocksWalking(int(s.TileAt(nx, ny))) {
		s.Log(MsgBlocked)
		return
	}
	s.X, s.Y = nx, ny
	s.tick()
	if loc, ok := u5data.LocationAt(nx, ny); ok {
		s.Log("往" + d.Name() + "方前行 —— 此處是" + loc.DisplayName() + "。")
	} else {
		s.Log("往" + d.Name() + "方前行。")
	}
}

func (s *State) moveInScene(d Direction) {
	dx, dy := d.Delta()
	nx, ny := s.X+dx, s.Y+dy

	// 邊界偵測。原版比的是**移動前**的座標:x<1 / x>30 / y<1 / y>30
	// (asm 是 `cmp byte_3E0A6, 1 / jnb` 與 `cmp byte_3E0A6, 1Eh / jbe`),
	// 也就是「已經站在最外圈,又要往外走」才算離開,不是「走到會出界」。
	atEdge := false
	switch d {
	case West:
		atEdge = s.X < 1
	case East:
		atEdge = s.X > 30
	case North:
		atEdge = s.Y < 1
	case South:
		atEdge = s.Y > 30
	}

	// 通行判定。原版還會先問「這格上有沒有 NPC / 物件」(sub_2B360),
	// 那要等 .NPC 格式解出來才做得了;目前只判地形。
	if u5data.TileBlocksWalking(int(s.TileAt(nx, ny))) && !atEdge {
		s.Log(MsgBlocked)
		return
	}

	if atEdge {
		s.Prompt = PromptLeave
		s.Log(MsgLeaveQuestion)
		return
	}

	s.X, s.Y = nx, ny
	s.tick()

	// 踩到樓梯就換層。原版 sub_758:同向走進去 inc 樓層,反向 dec。
	if facing, ok := u5data.StairsFacing(s.TileAt(s.X, s.Y)); ok {
		switch Direction(facing) {
		case d:
			s.changeFloor(+1)
		case d.Opposite():
			s.changeFloor(-1)
		}
	}
}

// changeFloor 換一層並重載場景地圖。原版是 inc/dec byte_3E0A5 後呼叫 sub_5C8(1)。
func (s *State) changeFloor(delta int) {
	next := s.Floor + delta
	m, err := s.Scenes.Map(s.Location, next)
	if err != nil {
		// 樓層超出範圍不該發生(梯子只出現在有下一層的地方)。真的發生時要看得見,
		// 不要靜默留在原地 —— 那會變成「梯子壞了」這種很難查的 bug。
		s.Log(MsgNothingToClimb)
		return
	}
	s.Floor = next
	s.scene = m
	s.loadNPCs()
	if delta > 0 {
		s.Log(MsgUp)
	} else {
		s.Log(MsgDown)
	}
}

// Klimb 是原版的「攀爬」指令(sub_EA0,按鍵 K)。
//
// 它看的是**腳下**那一格,不是要走進去的那一格 —— 這跟樓梯(走進去觸發)是兩套機制,
// 而梯子才是主力:四座燈塔、兩座城堡、修道院、巨蛇要塞都只有梯子。
func (s *State) Klimb() {
	if !s.InScene() {
		// 大地圖上的攀爬(上山、進地牢)是另一條路徑,還沒做。
		s.Log(MsgNothingToClimb)
		return
	}
	delta := u5data.ClimbDelta(s.TileAt(s.X, s.Y))
	if delta == 0 {
		s.Log(MsgNothingToClimb)
		return
	}
	s.changeFloor(delta)
}

// Enter 是原版的「進入」指令(sub_10928)。
//
// 原版做的是:掃地點表找出座標與玩家世界座標相符的那一筆;找不到就回「此處無可進入之地」。
func (s *State) Enter() {
	if s.InScene() {
		s.Log(MsgNothingToEnter)
		return
	}
	// 地牢入口與城鎮共用同一張地點表(索引 0x20..0x27),先查地牢。
	if n, ok := u5data.DungeonAt(s.X, s.Y); ok {
		if s.Transport != u5data.VehicleWalk && s.Transport != u5data.VehicleWalk+1 {
			s.Log("得下來走路才進得去!")
			return
		}
		s.Log("進入" + u5data.DungeonEntrances[n].DisplayName() + "。")
		s.EnterDungeon(n, false)
		return
	}
	loc, ok := u5data.LocationAt(s.X, s.Y)
	if !ok {
		s.Log(MsgNothingToEnter)
		return
	}
	num := loc.Number()
	m, err := s.Scenes.Map(num, 0)
	if err != nil {
		s.Log("進不去" + loc.DisplayName() + ":" + err.Error())
		return
	}
	s.Location = num
	s.Floor = 0
	s.X, s.Y = SceneEntryX, SceneEntryY
	s.scene = m
	s.sceneObjects = &u5data.ObjectSet{}
	s.loadNPCs()
	// 原版 sub_8924 只在**進場景**時建立執行期狀態;換樓層不重建
	// (NPC 還在原本走到的位置上,只是玩家換層了)。
	s.initRuntimeNPCs()
	s.Log("進入" + loc.DisplayName() + "。")
}

// Answer 回答待決的提問。原版只收 Y / N / ESC,ESC 等同 N。
func (s *State) Answer(yes bool) {
	if s.Prompt != PromptLeave {
		return
	}
	s.Prompt = PromptNone
	if !yes {
		s.Log(MsgNo)
		return
	}
	s.leaveScene()
}

// leaveScene 離開場景回到大地圖。
//
// 原版 sub_86C 的離開分支:世界座標**從地點表讀回來**(byte_410F3/byte_4111B 用
// 1-based 地點編號索引,就是同一張座標表),不是記著進來前的位置 ——
// 所以無論在城裡走到哪一格出去,都會回到城門那一格。
func (s *State) leaveScene() {
	loc, err := u5data.LocationByNumber(s.Location)
	if err != nil {
		return
	}
	underworld := s.Location == UnderworldLocation
	s.X, s.Y = loc.X, loc.Y
	s.Location = 0
	s.scene = nil
	s.npcs = nil
	s.rtNPCs = nil
	s.sceneObjects = nil
	if underworld {
		s.Floor = -1
		s.Log(MsgExitTo + MsgUnderworld)
	} else {
		s.Floor = 0
		s.Log(MsgExitTo + MsgBritannia)
	}
}

// LoadFrom 把一份原版存檔套進遊戲狀態。
//
// 這讓「匯入原版存檔」直接可用,也讓開局時間、隊伍與位置都來自原版而不是寫死的預設
// (CLAUDE.md §3.0:沒有證據就不要自己編數值)。
func (s *State) LoadFrom(sv *u5data.Save) {
	if sv == nil {
		return
	}
	s.BaseSave = sv
	s.Clock = Clock{Minute: sv.Minute, Hour: sv.Hour, Day: sv.Day, Month: sv.Month, Year: sv.Year}
	s.Karma = sv.Karma
	s.Transport = sv.Transport
	s.X, s.Y = sv.X, sv.Y
	s.Floor = sv.Floor
	s.Roster = append(s.Roster[:0], sv.Roster[:]...)
	s.PartySize = sv.PartySize
	s.Inventory = sv.Inventory
	s.removed = nil
	// 存檔可能是在城裡存的 —— 把場景與 NPC 一起載回來。
	s.Location = 0
	s.scene, s.npcs = nil, nil
	if sv.Location > 0 && s.Scenes != nil {
		if err := s.SetScene(sv.Location, sv.Floor, sv.X, sv.Y); err != nil {
			// 載不回場景就退回大地圖,並說明 —— 靜默把玩家丟到別處更糟。
			s.Log("讀檔:回不到原本的場景(" + err.Error() + "),改由大地圖開始。")
		}
	}
}

// Party 回傳目前隊伍 —— 名冊的前 PartySize 名。
func (s *State) Party() []*u5data.Character {
	n := s.PartySize
	if n > len(s.Roster) {
		n = len(s.Roster)
	}
	out := make([]*u5data.Character, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, &s.Roster[i])
	}
	return out
}

// AvatarName 回傳聖者的名字(名冊第 0 名)。對話裡的 opcode 0x81 用它。
func (s *State) AvatarName() string {
	if len(s.Roster) > 0 && s.Roster[0].Name != "" {
		return s.Roster[0].Name
	}
	return ""
}

// SetScene 直接把狀態放進某個地點的某一層(給工具與測試用,不是遊戲流程)。
func (s *State) SetScene(num, floor, x, y int) error {
	m, err := s.Scenes.Map(num, floor)
	if err != nil {
		return err
	}
	s.Location, s.Floor, s.X, s.Y, s.scene = num, floor, x, y, m
	// 進場景要清空物件槽(原版 sub_1678 把槽 1..31 歸零再載入場景)。
	s.sceneObjects = &u5data.ObjectSet{}
	s.loadNPCs()
	s.initRuntimeNPCs()
	return nil
}

// SeedRandom 換掉戰鬥骰子的種子。遊戲啟動時用時間,測試不呼叫就是固定種子。
func (s *State) SeedRandom(seed int64) { s.rng = rand.New(rand.NewSource(seed)) }

// Roll 回傳 [lo, hi] 的亂數(原版 `sub_28E14(lo, hi)`)。
//
// ⚠ **閉區間**,兩端都取得到 —— 原版算範圍時是 `hi − lo` 之後 `inc ecx`。
// 少了那個 +1,命中骰就永遠擲不到上界 30,最強的攻擊也會偶爾落空。
func (s *State) Roll(lo, hi int) int {
	if hi < lo {
		return lo
	}
	if s.rng == nil {
		s.rng = rand.New(rand.NewSource(1))
	}
	return lo + s.rng.Intn(hi-lo+1)
}

// SetTileAt 改寫地圖上的一格(解鎖的門、燒掉的東西之類)。
//
// 回傳 false 代表這一層的地圖不支援寫入。⚠ 只改記憶體裡的副本 ——
// **絕不寫回原版檔案**(CLAUDE.md §3.2 的硬規則:`internal/u5data` 只讀)。
func (s *State) SetTileAt(x, y int, tile byte) bool {
	if s.InScene() {
		if s.scene == nil || x < 0 || x >= u5data.SceneSide ||
			y < 0 || y >= u5data.SceneSide {
			return false
		}
		s.scene.Tiles[y*u5data.SceneSide+x] = tile
		return true
	}
	w := s.currentWorld()
	if w == nil {
		return false
	}
	w.Tiles[WrapWorld(y)*u5data.WorldSide+WrapWorld(x)] = tile
	return true
}
