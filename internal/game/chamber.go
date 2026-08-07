package game

import (
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 四間 11×11 的石室(`MISCMAPS.DAT`)
//
// 聖壇、寶典、黑棘的審問、王座廳這四幕都不在正常的場景地圖裡:原版把
// 地點編號設成 0xFF,再載一張 11×11 蓋掉畫面(`sub_1DA10` / `sub_C414` /
// `sub_135FC` 都做同一件事,只差載入位移)。
//
// 引擎原本直接在原地跑對話流程 —— 規則對了,但玩家看到的還是外面的世界。
// 這一層補上那張圖:進去時記下原本的位置、把畫面換成石室,離開時還原。
//
// ⚠ **不畫走進走出的動畫。** 原版是一格一格走過去的(`sub_134CC`);
// 這裡直接放在走完之後的那一格。動畫是呈現,不是機制 —— 但站定的位置
// 照原版算(`u5data.MiscMapStandY`),因為那決定了畫面構圖。

// Chamber 是進行中的石室。
type Chamber struct {
	// Which 是四張裡的哪一張(u5data.MiscMapIndex*)。
	Which int
	// Map 是那張 11×11。
	Map *u5data.MiscMap
	// 進去之前的位置,離開時要還原。
	backLocation, backFloor, backX, backY int
}

// enterChamber 把畫面換成第 which 張石室。回傳有沒有成功。
//
// 沒有 `MISCMAPS.DAT` 時回 false —— 呼叫端照樣把流程跑完(規則不依賴畫面),
// 只是玩家看到的還是外面。誠實降級,不要假裝進去了。
func (s *State) enterChamber(which int) bool {
	if s.MiscMaps == nil || s.Chamber != nil {
		return false
	}
	c := &Chamber{
		Which:        which,
		Map:          &s.MiscMaps.Maps[which],
		backLocation: s.Location,
		backFloor:    s.Floor,
		backX:        s.X,
		backY:        s.Y,
	}
	s.Chamber = c
	s.Location = u5data.MiscMapLocation
	s.X, s.Y = u5data.MiscMapEnterX, u5data.MiscMapStandY(which)
	return true
}

// leaveChamber 還原進石室之前的位置。
func (s *State) leaveChamber() {
	c := s.Chamber
	if c == nil {
		return
	}
	s.Location, s.Floor = c.backLocation, c.backFloor
	s.X, s.Y = c.backX, c.backY
	s.Chamber = nil
}

// InChamber 回報現在是不是在石室裡。
func (s *State) InChamber() bool { return s.Chamber != nil }

// chamberTileAt 取石室裡的一格。
//
// ⚠ 11×11 之外**不是黑的,是牆** —— 原版的視窗緩衝在石室期間只被寫了
// 中間 11×11,外圈留著上一次的內容。這裡回 `TileBlank`(全黑),
// 因為讓外面的殘影透出來更糟。這是呈現上的取捨,不影響規則。
func (s *State) chamberTileAt(x, y int) byte {
	return s.Chamber.Map.At(x, y)
}

// chamberName 是狀態列上顯示的石室名字。
func chamberName(which int) string {
	switch which {
	case u5data.MiscMapIndexCell:
		return "黑棘宮殿的地牢"
	case u5data.MiscMapIndexShrine:
		return "聖壇"
	case u5data.MiscMapIndexCodex:
		return "終極智慧之寶典"
	case u5data.MiscMapIndexThrone:
		return "王座廳"
	}
	return ""
}
