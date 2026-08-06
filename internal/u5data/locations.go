package u5data

// 地點表(從原版執行檔取得,2026-08-07)
//
// 來源:FM Towns `WORRIORS.EXP` 的三張平行表 ——
// `off_41054[32]`(名稱指標)、`byte_410F4[32]`(X)、`byte_4111C[32]`(Y),
// 由 `sub_10928` 以玩家世界座標線性搜尋比對(見 docs/re/03)。
//
// ⚠ 這張表**不是**從對 U5 的印象打出來的,是 dump 出來再用世界地圖交叉驗證的:
// 32 筆裡有 **26 筆的座標正好落在已知的入口 tile 上**(towne/village/hut/keep/castle),
// 而不命中的 6 筆反而補出新事實 —— 四座燈塔全落在 **tile 27**,
// 那正是 `DATA.OVL` 地點類型表裡一直對不上的 `lighthouse`。
//
// 場景對應(原版 `sub_5C8`,見 docs/re/03 §6):
//
//	檔案 = SceneFiles[(地點編號-1)/8]
//	地圖索引 = byte_41033[地點編號] + 樓層
//	檔案位移 = 索引 × 1024   (每張 32×32 = 1024 B)
//
// 索引 13–17 沒有名字(指標為 0),這與 `sub_10928` 裡
// `if (i < 13 || i > 17)` 才印名字的判斷完全吻合 —— 其中 #16 (86,107) 是
// Lord British 城堡(tile 62),程式另外處理。
//
// 中文譯名只填**已定案**的八德城市(對齊 CONTEXT.md 的 glossary 與 u4-cht/聖者之書體系);
// 其餘留空等《軟體世界》手冊 OCR 定案 —— 憑印象硬翻會變成二手轉譯。
type Location struct {
	X, Y int
	// SceneFile 是 SceneFiles 的索引 = (地點編號-1)/8。
	SceneFile int
	// SceneIndex 是該地點在場景檔內的**起始**地圖索引(地面層)。
	SceneIndex int
	// Floors 是該地點有幾層(= 同檔下一個地點的起始索引 - 自己;最後一個算到 16)。
	Floors int
	Name   string // 原版英文名(canonical,玩家輸入比對用這個)
	NameZH string // 中文譯名;空字串 = 尚未定案
}

// SceneFiles 是場景檔名,**順序就是原版的 off_4FC44**,不可調換 ——
// 檔案由 `(地點編號-1)/8` 選出。
var SceneFiles = [4]string{"TOWNE.DAT", "DWELLING.DAT", "CASTLE.DAT", "KEEP.DAT"}

// SceneOffset 回傳這個地點某一層在場景檔內的位元組位移。
//
// 原版 sub_5C8 的算法:`(SceneIndex + floor) << 10`(每張地圖 32×32 = 1024 B),
// 且 floor > 0x7F 時減 0x100 —— 那是**地下層用負數表示**。
func (l *Location) SceneOffset(floor int) int {
	if floor > 0x7F {
		floor -= 0x100
	}
	return (l.SceneIndex + floor) * SceneTiles
}

// Locations 是原版的 32 個地點。
var Locations = [32]Location{
	{X: 232, Y: 135, SceneFile: 0, SceneIndex: 0, Floors: 2, Name: "MOONGLOW", NameZH: "月光城"},        // 1
	{X: 81, Y: 106, SceneFile: 0, SceneIndex: 2, Floors: 2, Name: "BRITAIN", NameZH: "不列顛城"},         // 2
	{X: 36, Y: 222, SceneFile: 0, SceneIndex: 4, Floors: 3, Name: "JHELOM", NameZH: "哲倫"},            // 3
	{X: 58, Y: 43, SceneFile: 0, SceneIndex: 7, Floors: 1, Name: "YEW", NameZH: "紫衫城"},               // 4
	{X: 159, Y: 20, SceneFile: 0, SceneIndex: 8, Floors: 2, Name: "MINOC", NameZH: "米諾克"},            // 5
	{X: 106, Y: 184, SceneFile: 0, SceneIndex: 10, Floors: 2, Name: "TRINSIC", NameZH: "特林希克"},       // 6
	{X: 22, Y: 128, SceneFile: 0, SceneIndex: 12, Floors: 2, Name: "SKARA BRAE", NameZH: "史卡拉布雷"},    // 7
	{X: 187, Y: 169, SceneFile: 0, SceneIndex: 14, Floors: 2, Name: "NEW MAGINCIA", NameZH: "新馬精西亞"}, // 8
	{X: 88, Y: 120, SceneFile: 1, SceneIndex: 0, Floors: 3, Name: "FOGSBANE", NameZH: ""},            // 9
	{X: 152, Y: 24, SceneFile: 1, SceneIndex: 3, Floors: 3, Name: "STORMCROW", NameZH: ""},           // 10
	{X: 104, Y: 216, SceneFile: 1, SceneIndex: 6, Floors: 3, Name: "GREYHAVEN", NameZH: ""},          // 11
	{X: 216, Y: 120, SceneFile: 1, SceneIndex: 9, Floors: 3, Name: "WAVEGUIDE", NameZH: ""},          // 12
	{X: 45, Y: 62, SceneFile: 1, SceneIndex: 12, Floors: 1, Name: "IOLO'S HUT", NameZH: ""},          // 13
	{X: 176, Y: 208, SceneFile: 1, SceneIndex: 13, Floors: 1, Name: "", NameZH: ""},                  // 14
	{X: 201, Y: 59, SceneFile: 1, SceneIndex: 14, Floors: 1, Name: "", NameZH: ""},                   // 15
	{X: 153, Y: 91, SceneFile: 1, SceneIndex: 15, Floors: 1, Name: "", NameZH: ""},                   // 16
	{X: 86, Y: 107, SceneFile: 2, SceneIndex: 1, Floors: 5, Name: "", NameZH: ""},                    // 17
	{X: 196, Y: 245, SceneFile: 2, SceneIndex: 6, Floors: 4, Name: "", NameZH: ""},                   // 18
	{X: 84, Y: 106, SceneFile: 2, SceneIndex: 10, Floors: 1, Name: "WEST BRITANNY", NameZH: ""},      // 19
	{X: 86, Y: 105, SceneFile: 2, SceneIndex: 11, Floors: 1, Name: "NORTH BRITANNY", NameZH: ""},     // 20
	{X: 88, Y: 106, SceneFile: 2, SceneIndex: 12, Floors: 1, Name: "EAST BRITANNY", NameZH: ""},      // 21
	{X: 98, Y: 145, SceneFile: 2, SceneIndex: 13, Floors: 1, Name: "PAWS", NameZH: ""},               // 22
	{X: 136, Y: 90, SceneFile: 2, SceneIndex: 14, Floors: 1, Name: "COVE", NameZH: ""},               // 23
	{X: 136, Y: 158, SceneFile: 2, SceneIndex: 15, Floors: 1, Name: "BUCCANEER'S DEN", NameZH: ""},   // 24
	{X: 49, Y: 58, SceneFile: 3, SceneIndex: 0, Floors: 2, Name: "ARARAT", NameZH: ""},               // 25
	{X: 15, Y: 160, SceneFile: 3, SceneIndex: 2, Floors: 2, Name: "BORDERMARCH", NameZH: ""},         // 26
	{X: 64, Y: 240, SceneFile: 3, SceneIndex: 4, Floors: 1, Name: "FARTHING", NameZH: ""},            // 27
	{X: 248, Y: 8, SceneFile: 3, SceneIndex: 5, Floors: 1, Name: "WINDEMERE", NameZH: ""},            // 28
	{X: 148, Y: 74, SceneFile: 3, SceneIndex: 6, Floors: 1, Name: "STONEGATE", NameZH: ""},           // 29
	{X: 218, Y: 107, SceneFile: 3, SceneIndex: 7, Floors: 3, Name: "THE LYCAEUM", NameZH: ""},        // 30
	{X: 28, Y: 50, SceneFile: 3, SceneIndex: 10, Floors: 4, Name: "EMPATH ABBEY", NameZH: ""},        // 31
	{X: 146, Y: 241, SceneFile: 3, SceneIndex: 14, Floors: 2, Name: "SERPENT'S HOLD", NameZH: ""},    // 32
}

// LocationAt 回報世界座標上有沒有地點。
func LocationAt(x, y int) (*Location, bool) {
	for i := range Locations {
		if Locations[i].X == x && Locations[i].Y == y {
			return &Locations[i], true
		}
	}
	return nil, false
}

// DisplayName 回傳顯示用名稱:有中文就用中文,否則退回英文。
//
// ⚠ 玩家輸入比對一律用 Name(英文)—— 玩家在遊戲中打不出中文(u4-cht 踩過的坑)。
func (l *Location) DisplayName() string {
	if l.NameZH != "" {
		return l.NameZH
	}
	if l.Name != "" {
		return l.Name
	}
	return "?"
}
