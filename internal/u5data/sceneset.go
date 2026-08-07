package u5data

import (
	"fmt"
	"path/filepath"
)

// SceneSet 是四個場景檔載好之後的集合,負責把「地點編號 + 樓層」解析成一張 32×32 地圖。
//
// 解析規則出自原版的場景載入函式 sub_5C8(docs/re/03 §6):
//
//	檔案 = SceneFiles[(地點編號-1)/8]
//	索引 = Locations[地點編號-1].SceneIndex + 樓層
//	位移 = 索引 × 1024
//
// 樓層在原版是 signed char:0 是地面層,往上遞增,往下遞減(0xFF = 地下一層)。
type SceneSet struct {
	Files [len(SceneFiles)][]SceneMap
}

// LoadSceneSet 從資料目錄讀入四個場景檔。任一檔缺失即失敗 —— 場景檔是遊戲本體,
// 不像字型那樣可以優雅降級。
func LoadSceneSet(dir string) (*SceneSet, error) {
	s := &SceneSet{}
	for i, name := range SceneFiles {
		maps, err := LoadSceneMaps(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		s.Files[i] = maps
	}
	return s, nil
}

// Map 回傳地點 num(1-based)在 floor 層的場景地圖。
//
// floor 用 Go 的正負整數表示(地下層是負數),不是原版的 0xFF 編碼 ——
// 那個編碼只是 signed char 的位元表示,不是語意。
func (s *SceneSet) Map(num, floor int) (*SceneMap, error) {
	loc, err := LocationByNumber(num)
	if err != nil {
		return nil, err
	}
	if floor < loc.FloorMin || floor > loc.FloorMax {
		return nil, fmt.Errorf("%s 的樓層範圍是 %+d..%+d,沒有第 %+d 層",
			loc.DisplayName(), loc.FloorMin, loc.FloorMax, floor)
	}
	idx := loc.SceneIndex + floor
	maps := s.Files[loc.SceneFile]
	if idx < 0 || idx >= len(maps) {
		return nil, fmt.Errorf("%s 第 %d 層算出索引 %d,超出 %s 的 %d 張",
			loc.Name, floor, idx, SceneFiles[loc.SceneFile], len(maps))
	}
	return &maps[idx], nil
}

// LocationByNumber 依 1-based 地點編號取地點(原版 byte_3E0A3 用的就是這個編號,0 代表在大地圖)。
func LocationByNumber(num int) (*Location, error) {
	if num < 1 || num > len(Locations) {
		return nil, fmt.Errorf("地點編號 %d 超出 1..%d", num, len(Locations))
	}
	return &Locations[num-1], nil
}

// TileBlank 是「此處無物」的 tile —— 原版把 11×11 視窗緩衝、32×32 暫存與
// 星象緩衝一律 `memset(…, 0xFF, …)`(sub_D064 等三處),而 tile 0xFF 在 tileset 裡
// 就是一格純黑。所以場景邊界外要填 0xFF,不是 0:tile 0 是一團紅黃爆裂圖案,
// 填 0 會在城鎮南緣鋪出一片火花。
const TileBlank = 0xFF

// 樓梯 tile
//
// 出自原版的 sub_758:`(tile & 0xFC) == 0xC4` 才是樓梯,低 2 bit 是樓梯的朝向
// (0=北 1=東 2=南 3=西)。朝同向走進去往上,反向走進去往下。
const (
	// StairsBase 是四個樓梯 tile 的起點(0xC4..0xC7)。
	StairsBase = 0xC4
	// StairsMask 用來判斷是不是樓梯。
	StairsMask = 0xFC
)

// 梯子 tile
//
// 樓梯(0xC4-0xC7)是**走進去**觸發的,梯子則要站在上面按 K(Klimb)——
// 原版 sub_EA0 讀腳下那一格:0xC8 呼叫 sub_758(0, 196) 等同上樓,
// 0xC9 與 0x86 呼叫 sub_758(2, 196) 等同下樓。
//
// 這兩套機制並存,而且**梯子才是主力**:四座燈塔、兩座城堡、修道院、
// 巨蛇要塞全靠梯子上下,只認樓梯的話那些地方會爬不上去。
const (
	// LadderUp 是往上的梯子。
	LadderUp = 0xC8
	// LadderDown 是往下的梯子。
	LadderDown = 0xC9
	// TrapDoor 是另一個往下的出口(原版 sub_EA0 與 0xC9 同樣處理)。
	TrapDoor = 0x86
)

// ClimbDelta 回報站在 tile 上按「攀爬」會換到哪一層:+1 上、-1 下、0 不能爬。
func ClimbDelta(tile byte) int {
	switch tile {
	case LadderUp:
		return +1
	case LadderDown, TrapDoor:
		return -1
	}
	return 0
}

// StairsFacing 回報 tile 是否為樓梯,以及它的朝向(0=北 1=東 2=南 3=西)。
func StairsFacing(tile byte) (facing int, ok bool) {
	if tile&StairsMask != StairsBase {
		return 0, false
	}
	return int(tile - StairsBase), true
}
