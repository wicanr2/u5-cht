package game

import (
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 視野半徑與視線遮蔽罩(原版 `sub_29304` 的日照段 → `sub_2E0E8` → `sub_2DDB0`)
//
// 兩件事:能看多遠(光),以及牆擋不擋得住(視線)。
// 前者是時鐘算出來的一個平方半徑,後者是那個半徑餵給 flood fill 的結果。
//
// ★ 這兩件事在原版是**分開**的。半徑不是「看得到多遠」的上限 ——
// 它是「這個距離內不必檢查有沒有被擋」的捷徑。白天在空曠處,
// 半徑 50 剛好蓋滿整個 11×11(角落的平方距離正是 50),於是視線判定
// 完全不啟用;天一黑半徑掉到 2,牆與山才開始真的擋住東西。

// 日照的平方半徑(全部來自 `sub_29304` 的 `loc_29463`)。
const (
	// SightRadiusDay 是白天(6:00–18:59)。
	//
	// = 0x32 = 50,正好等於 11×11 角落的平方距離 → 整個視窗都在「捷徑」內。
	SightRadiusDay = 0x32
	// SightRadiusNight 是夜間、地下、以及亞拉臘號殘骸。
	SightRadiusNight = 2
	// SightRadiusInLor 是 In Lor(光明咒語)保證的下限。
	SightRadiusInLor = 0x12
	// SightRadiusTorch 是火把保證的下限。
	//
	// ⚠ 比咒語**小**。看起來反直覺,但原版就是這樣:
	// `cmp byte_3E0B5, 12h`(咒語)排在 `cmp byte_3E0B5, 0Ah`(火把)前面,
	// 兩個都是「不足就補到這個值」。咒語亮、火把久。
	SightRadiusTorch = 0x0A
)

// 日出與日落的鐘點(`cmp al, 5` / `cmp al, 13h`)。
const (
	SightDawnHour = 5
	SightDuskHour = 19
)

// sightDawnRamp 是日出/日落六段漸變(原版 `byte_5FF5C`,索引 = 分鐘 / 10)。
//
// ★ IDA 把後兩個位元組誤判成 `dd offset unk_3122` —— 照 `.asm` 抄只會拿到
// 前四個(2 5 10 20),而索引跑到 0..5。直接讀執行檔位元組才看得到全部六個:
//
//	02 05 0A 14 22 31 00 00
//	         ↑ 表在這裡結束(後面兩個 0 就是證據)
//
// 從 2(夜)一路開到 49(≈ 白天的 50),十分鐘一階。
var sightDawnRamp = [6]int{2, 5, 10, 20, 34, 49}

// SightAlwaysDarkLocation 是不論日夜都暗的地點(`cmp byte_3E0A3, 19h`)。
//
// = 25 = 亞拉臘號(ARARAT)的殘骸。整個地點寫死走夜間半徑。
const SightAlwaysDarkLocation = 25

// LightRadius2 是現在的視野平方半徑(原版 `byte_3E0B5`)。
//
// 判斷順序照 `sub_29304`,不可重排 —— 尤其兩個下限是**先咒語後火把**,
// 而且是「不足才補」:同時有咒語與火把時火把那一關會被跳過。
func (s *State) LightRadius2() int {
	r := SightRadiusNight
	switch {
	case s.Location == SightAlwaysDarkLocation:
		// 亞拉臘號殘骸。
	case s.Floor < 0:
		// 地下(原版判的是 `byte_3E0A5 > 0x7F`,也就是樓層的補數表示為負)。
	case s.Clock.Hour < SightDawnHour || s.Clock.Hour > SightDuskHour:
		// 夜。
	case s.Clock.Hour == SightDawnHour:
		r = sightDawnRamp[s.Clock.Minute/10]
	case s.Clock.Hour == SightDuskHour:
		// 日落是反著走:59 分時才與夜間齊平。
		r = sightDawnRamp[(59-s.Clock.Minute)/10]
	default:
		r = SightRadiusDay
	}
	if s.LightTurns > 0 && r < SightRadiusInLor {
		r = SightRadiusInLor
	}
	if s.TorchTurns > 0 && r < SightRadiusTorch {
		r = SightRadiusTorch
	}
	return r
}

// SightMask 是這一幀的 11×11「哪一格看得到」。
//
// 回傳 nil 代表**不做視線遮蔽**,呼叫端照常全畫。原版在
// `cmp byte_3E0A3, 80h; jnb loc_29EB4` 那裡就分了岔:地點編號 ≥ 0x80
// 直接把場景視窗整份 `rep movsd` 過去,不跑 flood fill ——
// 也就是**戰鬥與四間石室沒有視線遮蔽**。地牢是另一套(第一人稱透視)。
//
// 兩輪:先算場景照明(`sub_2E21C`),再以玩家為中心算罩子(`sub_2E0E8`)。
// 順序不能顛倒 —— 第二輪要查第一輪的結果。
func (s *State) SightMask() []byte {
	if s.InCombat() || s.InDungeon() || s.InChamber() {
		return nil
	}
	r := s.LightRadius2()
	if s.SeeThroughWalls {
		// Wis An Ylem:`sub_2E0E8(-1, …)`,整張攤開。
		r = -1
	}
	half := u5data.SightSide / 2
	// 視窗左上角 —— 亮度圖與罩子共用這個原點。
	ox, oy := s.X-half, s.Y-half
	lit := u5data.ComputeLit(func(x, y int) byte {
		return s.TileAt(ox+x, oy+y)
	})
	w := u5data.SightWindow{
		Tile: func(x, y int) byte {
			return s.TileAt(ox+x, oy+y)
		},
		Lit: func(x, y int) bool {
			i := y*u5data.SightStride + x
			return i >= 0 && i < len(lit) && lit[i] != 0
		},
	}
	mask := u5data.ComputeSight(w, r)
	return mask[:]
}

// SightVisible 回報 11×11 視窗裡的某一格看不看得見。
//
// dx / dy 是相對玩家的位移(−5..+5)。mask 為 nil(戰鬥、石室、地牢)一律回 true。
func SightVisible(mask []byte, dx, dy int) bool {
	if mask == nil {
		return true
	}
	half := u5data.SightSide / 2
	x, y := dx+half, dy+half
	if x < 0 || x >= u5data.SightSide || y < 0 || y >= u5data.SightSide {
		return false
	}
	return mask[y*u5data.SightSide+x] != u5data.SightHidden
}
