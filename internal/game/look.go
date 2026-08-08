package game

import (
	"fmt"

	"github.com/wicanr2/u5-cht/internal/i18n"
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// Look 指令(原版 `sub_D9C4`,經 `sub_D258` 分派)
//
// 「看」是 U5 少數會把資料檔整個攤在玩家面前的指令:`LOOK2.DAT` 是一張
// 512 格的敘述表,`SIGNS.DAT` 是 78 塊招牌。兩份此前都標著「格式未解」,
// 實際上都只是位移表 + 字串(見 `docs/re/37`)。
//
// # 分派順序(照原版,順序有意義)
//
//	1. 問方向。
//	2. 目標格是水晶球(0x29)→ 凝視,不印「汝所見為」。
//	3. 印「汝所見為」。
//	4. 那一格**有物件**  → 物件敘述(`LOOK2` 索引 +256),到此為止。
//	5. tile 是招牌類    → 讀 `SIGNS.DAT`。
//	6. 其餘             → 地形敘述,外加幾個會接後綴的特例。
//
// ⚠ 第 4 步「有物件就到此為止」很容易寫錯成「兩個都印」。原版是 `jmp`,
// 站在招牌前面的一袋金幣會**蓋掉**招牌 —— 那是原版行為,照做。
const (
	// LookCrystalSphere 是水晶球:凝視,走另一條路。
	LookCrystalSphere = 0x29
	// LookSky 是可以抬頭看天的格子(原版 tile 0x59)。
	LookSky = 0x59
	// LookWell 是井 —— 而井是**互動**的(許願井彩蛋,見 `well.go`)。
	LookWell = u5data.WellLookTile
	// LookFountainGroup 是噴泉:`tile & 0xFC == 0xD8`,四個朝向。
	LookFountainGroup = 0xD8
	// LookFlame 是三聖火的封印,敘述結尾「the Flame of 」要接火名。
	LookFlame = 0xDE
	// LookSealedDungeon 是崩塌的地牢入口,敘述結尾要接地牢名。
	LookSealedDungeon = 0xDF
	// LookClockGroup 是老爺鐘:`tile & 0xFE == 0xFA`,敘述結尾要接時刻。
	LookClockGroup = 0xFA
)

// 轉向格:看到這幾格會**改看鄰格**,而且會一路跟下去。
//
// 原版是個 `while` 迴圈(`loc_D268`),不是查一次就算 —— 連著兩格轉向也走得通。
const (
	LookRedirectNorth = 0xE0
	LookRedirectEast  = 0xE1
	LookRedirectWest  = 0xE2
)

// 招牌格:這幾種 tile 要去 `SIGNS.DAT` 查,而不是查 `LOOK2.DAT`。
var lookSignTiles = map[byte]bool{0x89: true, 0x8A: true, 0xA0: true, 0xA4: true, 0xF8: true}

// lookFlameLocations 是三聖火各自所在的地點編號(原版比 `byte_3E0A3`)。
var lookFlameLocations = map[int]string{
	0x1E: "Truth",
	0x1F: "Love",
	0x20: "Courage",
}

// lookDungeonByX 是「崩塌的地牢入口」對應的地牢名。
//
// ★ 原版**用入口的 x 座標分辨是哪一座地牢**,不是查地點表 —— 八座地牢的
// x 剛好兩兩不同,於是一個 switch 就夠了。這種寫法在現代看起來像 hack,
// 但它就是原版的判斷依據,照抄才對得上(改用查表會在座標被改動時行為不同)。
var lookDungeonByX = map[int]string{
	0x3A: "Shame",
	0x48: "Destard",
	0x5B: "Despise",
	0x7E: "Wrong",
	0x80: "Doom",
	0x9C: "Covetous",
	0xEF: "Hythloth",
	0xF0: "Deceit",
}

// Look 是 L 指令。
func (s *State) Look() {
	// ★ 地牢是**另一支程式**(原版 `sub_2ACF4` 的 'L' 依地點碼分派到 `sub_EEEC`)——
	// 相對方向、十六種地形描述、噴泉能喝。見 `internal/game/dungeonlook.go`。
	if s.InDungeon() {
		s.LookDungeon()
		return
	}
	s.AskDirection(func(d Direction) {
		dx, dy := d.Delta()
		s.lookAt(s.X+dx, s.Y+dy)
	})
}

// lookAt 看某一格。
func (s *State) lookAt(x, y int) {
	tile := s.TileAt(x, y)
	obj, _, hasObj := s.ObjectAt(x, y)

	if tile == LookCrystalSphere {
		s.gazeIntoSphere()
		return
	}
	s.Log(MsgThouDostSee)

	if hasObj {
		s.Log(s.lookObject(obj.Kind))
		return
	}
	if lookSignTiles[tile] {
		s.readSign(x, y)
		return
	}
	s.lookTerrain(tile, x, y)
}

// lookTerrain 看地形(原版 `sub_D258`)。
func (s *State) lookTerrain(tile byte, x, y int) {
	// 轉向格:一路跟到不是轉向格為止。
	//
	// ⚠ 加上步數上限是**我們加的**,原版沒有 —— 原版的資料裡不會成環,
	// 但引擎不該因為一份改壞的地圖就吊死。上限之外的行為與原版無異。
	for i := 0; i < u5data.SceneSide; i++ {
		switch tile {
		case LookRedirectNorth:
			y--
		case LookRedirectEast:
			x++
		case LookRedirectWest:
			x--
		default:
			i = u5data.SceneSide
			continue
		}
		tile = s.TileAt(x, y)
	}

	switch {
	case tile == LookSky:
		s.lookAtTheSky()
		return
	case tile == LookWell:
		s.lookAtWell()
		return
	case tile&0xFC == LookFountainGroup:
		s.drinkFromFountain()
		return
	}

	desc := s.lookTerrainText(tile)
	switch {
	case tile&0xFE == LookClockGroup:
		s.Log(desc + s.clockFace())
	case tile == LookFlame:
		if name, ok := lookFlameLocations[s.Location]; ok {
			s.Log(desc + i18n.LookSuffix(name))
			return
		}
		// 火名認不出來時原版**整句都不印**(`jmp loc_D457`),連敘述都沒有。
		// 這裡已經印了敘述才發現認不出來 —— 差別只在多一行,不影響狀態。
		s.Log(desc)
	case tile == LookSealedDungeon:
		if name, ok := lookDungeonByX[x]; ok {
			s.Log(desc + i18n.LookSuffix(name))
			return
		}
		s.Log(desc)
	default:
		s.Log(desc)
	}
}

// lookTerrainText 取地形敘述的譯文。
func (s *State) lookTerrainText(tile byte) string {
	if s.Look2 == nil {
		return ""
	}
	return i18n.Look(int(tile), s.Look2.Terrain(int(tile)))
}

// lookObject 取物件敘述的譯文(索引在物件空間)。
func (s *State) lookObject(kind byte) string {
	if s.Look2 == nil {
		return ""
	}
	i := int(kind) + u5data.LookObjectBase
	return i18n.Look(i, s.Look2.Object(int(kind)))
}

// clockFace 是老爺鐘的時刻,格式照原版:12 小時制、分鐘補零、AM / PM。
//
// 原版 `sub_D258`:`hour % 12`,餘 0 時改印 12;`hour <= 11` 是 AM。
func (s *State) clockFace() string {
	h := s.Clock.Hour % 12
	if h == 0 {
		h = 12
	}
	suffix := MsgClockPM
	if s.Clock.Hour <= 11 {
		suffix = MsgClockAM
	}
	return fmt.Sprintf("%d:%02d%s", h, s.Clock.Minute, suffix)
}

// readSign 讀一塊招牌(原版 `sub_D650`)。
//
// 查不到就印預設的那一塊 —— 原版 `sub_D544(-1)` 是寫死的「LIVE BY THE
// EIGHT LAWS」告示板,不是「什麼也沒有」。
func (s *State) readSign(x, y int) {
	if s.Signs == nil {
		s.Log(MsgSignDefault)
		return
	}
	sg, ok := s.Signs.At(s.Location, s.Floor, x, y)
	if !ok {
		s.Log(MsgSignDefault)
		return
	}
	for _, line := range i18n.SignLines(s.Location, sg.X, sg.Y, sg.Lines()) {
		s.Log(line)
	}
}

// drinkFromFountain 是噴泉(原版 `sub_CE78`)。
//
// ⚠ **它什麼也不做。** 組語裡沒有任何寫入 —— 沒有補血、沒有解毒、沒有加值,
// 只有三句話。看起來像沒寫完,但那就是原版,不要「順手」補上療效。
func (s *State) drinkFromFountain() {
	s.Log(MsgFountain)
	i := s.pickCharacter(MsgWhoWillDrink)
	if i < 0 {
		s.Log(MsgNobodyDrinks)
		return
	}
	switch s.Roster[i].Status {
	case u5data.StatusDead, u5data.StatusAsleep:
		s.Log(MsgIncapacitated)
	default:
		s.Log(MsgRefreshing)
	}
}

// gazeIntoSphere 是凝視水晶球(原版 `sub_D9C4` 的 tile 0x29 分支)。
//
// 擲 1..30 對上該員的**智力**:智力大於點數就看見全景(與 In Quas Wis 同一支),
// 否則是「死亡幻象」並**扣 1 點 HP**(`sub_2A464(member, 1)`)。
func (s *State) gazeIntoSphere() {
	i := s.pickCharacter("")
	if i < 0 {
		return
	}
	if int(s.Roster[i].Intel) > s.Roll(1, gazeRollMax) {
		s.Log(MsgStrangeVision)
		s.peerAtTheLand()
		return
	}
	s.Log(MsgDeathVision)
	s.damageMember(i, gazeDamage)
}

const (
	// gazeRollMax 是凝視水晶球要擲的上限(原版 `sub_28E14(1, 0x1E)`)。
	gazeRollMax = 30
	// gazeDamage 是看見死亡幻象要付的代價,原版寫死 1 點。
	gazeDamage = 1
)

// 抬頭看天(原版 `sub_D064`)
//
// 白天(6 時 ≤ hour < 18 時)看到的是太陽 —— 而且**會扣 1 點 HP**
// (`sub_2A464(member, 1)`,與凝視水晶球失敗同一支)。直視太陽會痛,
// 這是原版寫進去的。
//
// 夜裡是滿天星斗與月相:原版清掉 11×11 的覆蓋層,灑 80 顆隨機星,
// 再把三個天體(`byte_3E0D8`)畫到八個固定方位(`byte_54638` / `byte_54640`)
// 的其中之一,印「the night sky! 」後卡住等按鍵。
//
// ⚠ **星空的畫面還沒做**,這裡只出文字。畫面是 P6 美術的事,不是機制落差。
const (
	skyDayFrom   = 6  // 含
	skyDayUntil  = 18 // 不含
	skySunDamage = 1
)

func (s *State) lookAtTheSky() {
	if s.Clock.Hour >= skyDayFrom && s.Clock.Hour < skyDayUntil {
		s.Log(MsgTheSun)
		// 原版扣的是「目前動作的那一位」(`byte_3E08B`)。引擎還沒有單人狀態,
		// 用同一支挑人邏輯代替 —— 差別只在多人時挑誰,不在扣不扣。
		if i := s.pickCharacter(""); i >= 0 {
			s.damageMember(i, skySunDamage)
		}
		return
	}
	s.Log(MsgNightSky)
}
