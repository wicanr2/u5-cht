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
// 索引 13–17 沒有名字(指標為 0),這與 `sub_10928` 裡
// `if (i < 13 || i > 17)` 才印名字的判斷完全吻合 —— 其中 #16 (86,107) 是
// Lord British 城堡(tile 62),程式另外處理。
//
// 中文譯名只填**已定案**的八德城市(對齊 CONTEXT.md 的 glossary 與 u4-cht/聖者之書體系);
// 其餘留空等《軟體世界》手冊 OCR 定案 —— 憑印象硬翻會變成二手轉譯。
type Location struct {
	X, Y   int
	Name   string // 原版英文名(canonical,玩家輸入比對用這個)
	NameZH string // 中文譯名;空字串 = 尚未定案
}

// Locations 是原版的 32 個地點。
var Locations = [32]Location{
	{X: 232, Y: 135, Name: "MOONGLOW", NameZH: "月光城"},       // 0
	{X: 81, Y: 106, Name: "BRITAIN", NameZH: "不列顛城"},        // 1
	{X: 36, Y: 222, Name: "JHELOM", NameZH: "哲倫"},           // 2
	{X: 58, Y: 43, Name: "YEW", NameZH: "紫衫城"},              // 3
	{X: 159, Y: 20, Name: "MINOC", NameZH: "米諾克"},           // 4
	{X: 106, Y: 184, Name: "TRINSIC", NameZH: "特林希克"},       // 5
	{X: 22, Y: 128, Name: "SKARA BRAE", NameZH: "史卡拉布雷"},    // 6
	{X: 187, Y: 169, Name: "NEW MAGINCIA", NameZH: "新馬精西亞"}, // 7
	{X: 88, Y: 120, Name: "FOGSBANE", NameZH: ""},           // 8
	{X: 152, Y: 24, Name: "STORMCROW", NameZH: ""},          // 9
	{X: 104, Y: 216, Name: "GREYHAVEN", NameZH: ""},         // 10
	{X: 216, Y: 120, Name: "WAVEGUIDE", NameZH: ""},         // 11
	{X: 45, Y: 62, Name: "IOLO'S HUT", NameZH: ""},          // 12
	{X: 176, Y: 208, Name: "", NameZH: ""},                  // 13
	{X: 201, Y: 59, Name: "", NameZH: ""},                   // 14
	{X: 153, Y: 91, Name: "", NameZH: ""},                   // 15
	{X: 86, Y: 107, Name: "", NameZH: ""},                   // 16
	{X: 196, Y: 245, Name: "", NameZH: ""},                  // 17
	{X: 84, Y: 106, Name: "WEST BRITANNY", NameZH: ""},      // 18
	{X: 86, Y: 105, Name: "NORTH BRITANNY", NameZH: ""},     // 19
	{X: 88, Y: 106, Name: "EAST BRITANNY", NameZH: ""},      // 20
	{X: 98, Y: 145, Name: "PAWS", NameZH: ""},               // 21
	{X: 136, Y: 90, Name: "COVE", NameZH: ""},               // 22
	{X: 136, Y: 158, Name: "BUCCANEER'S DEN", NameZH: ""},   // 23
	{X: 49, Y: 58, Name: "ARARAT", NameZH: ""},              // 24
	{X: 15, Y: 160, Name: "BORDERMARCH", NameZH: ""},        // 25
	{X: 64, Y: 240, Name: "FARTHING", NameZH: ""},           // 26
	{X: 248, Y: 8, Name: "WINDEMERE", NameZH: ""},           // 27
	{X: 148, Y: 74, Name: "STONEGATE", NameZH: ""},          // 28
	{X: 218, Y: 107, Name: "THE LYCAEUM", NameZH: ""},       // 29
	{X: 28, Y: 50, Name: "EMPATH ABBEY", NameZH: ""},        // 30
	{X: 146, Y: 241, Name: "SERPENT'S HOLD", NameZH: ""},    // 31
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
