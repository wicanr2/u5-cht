package u5data

import (
	"fmt"
	"os"
)

// `MISCMAPS.DAT` —— 四間 11×11 的石室
//
// 聖壇、寶典、黑棘的審問、王座廳這四幕都不在正常的場景地圖裡:
// 原版把地點編號設成 0xFF,再從這個檔載入一張 11×11 蓋掉畫面。
//
// 格式極簡:**每張 11 列 × 每列 16 B,只用前 11 B**。載入端一律傳 0xB0(176)
// 當長度,所以一張就是 176 B,四張連續排在檔案前 704 B:
//
//	0x000  審問的牢房   sub_C414
//	0x0B0  八德聖壇     sub_1DA10(地形 25)
//	0x160  終極寶典     sub_1DA10(地形 17)
//	0x210  王座廳       sub_135FC
//
// ★ 四個位移不是猜的 —— 它們就寫在那三支的 `sub_2C740(…, 0xB0, 位移)` 呼叫裡。
// 而且解出來的內容自己會說話:牢房裡有 0xB9 上鎖的門與 0xBB 上鎖的魔法門、
// 寶典那張正中央是 0x41、王座廳有一整排家具與 0xAB/0xAC/0xAF 的窗。
//
// ⚠ **列距是 16 不是 11。** 原版的搬運迴圈是
// `byte_3F8F4[i*32 + j] = byte_3F844[i*16 + j]`(來源列距 16、目的列距 32)。
// 照 11 讀的話第二列開始整張歪掉,而畫面看起來只是「有點亂」——
// 不會有任何東西報錯。
const (
	// MiscMapSide 是石室的邊長。
	MiscMapSide = 11
	// MiscMapStride 是檔案裡每一列佔幾個位元組。
	MiscMapStride = 16
	// MiscMapSize 是一張的大小。
	MiscMapSize = MiscMapSide * MiscMapStride
)

// 四張石室在 `MISCMAPS.DAT` 裡的位移。
const (
	MiscMapCell    = 0x000 // 黑棘的牢房
	MiscMapShrine  = 0x0B0 // 八德聖壇
	MiscMapCodex   = 0x160 // 終極智慧之寶典
	MiscMapThrone  = 0x210 // 王座廳
	MiscMapCount   = 4
)

// MiscMapOffsets 依序是上面四張。
var MiscMapOffsets = [MiscMapCount]int{
	MiscMapCell, MiscMapShrine, MiscMapCodex, MiscMapThrone,
}

// MiscMap 是一張 11×11 的石室。
type MiscMap struct {
	Tiles [MiscMapSide * MiscMapSide]byte
}

// At 取一格;超出範圍回 TileBlank。
func (m *MiscMap) At(x, y int) byte {
	if x < 0 || x >= MiscMapSide || y < 0 || y >= MiscMapSide {
		return TileBlank
	}
	return m.Tiles[y*MiscMapSide+x]
}

// MiscMapSet 是四張石室。
type MiscMapSet struct {
	Maps [MiscMapCount]MiscMap
}

// ParseMiscMaps 解析 `MISCMAPS.DAT`。
func ParseMiscMaps(raw []byte) (*MiscMapSet, error) {
	need := MiscMapThrone + MiscMapSize
	if len(raw) < need {
		return nil, fmt.Errorf("MISCMAPS.DAT 只有 %d B,四張石室要 %d B", len(raw), need)
	}
	set := &MiscMapSet{}
	for i, off := range MiscMapOffsets {
		for y := 0; y < MiscMapSide; y++ {
			// ⚠ 來源列距 16,目的列距 11。
			copy(set.Maps[i].Tiles[y*MiscMapSide:(y+1)*MiscMapSide],
				raw[off+y*MiscMapStride:])
		}
	}
	return set, nil
}

// LoadMiscMaps 讀入 `MISCMAPS.DAT`。
func LoadMiscMaps(path string) (*MiscMapSet, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("讀取 %s: %w", path, err)
	}
	set, err := ParseMiscMaps(raw)
	if err != nil {
		return nil, fmt.Errorf("解析 %s: %w", path, err)
	}
	return set, nil
}

// 石室裡的站位(全部來自原版)
//
//	sub_1DA10  進來時放在 (5, 10),然後往上走 —— 寶典 7 步、聖壇 4 步
//	sub_135FC  王座廳的隊長放在 (5, 8),走到 (5, 3)
const (
	// MiscMapEnterX / MiscMapEnterY 是走進石室時站的那一格。
	MiscMapEnterX = 5
	MiscMapEnterY = 10
	// MiscMapWalkShrine / MiscMapWalkCodex 是走幾步到定位。
	MiscMapWalkShrine = 4
	MiscMapWalkCodex  = 7
	// MiscMapThroneY 是王座廳裡站定的位置。
	MiscMapThroneEnterY = 8
	MiscMapThroneY      = 3
)

// MiscMapStandY 回傳在第 which 張石室裡站定的 Y。
func MiscMapStandY(which int) int {
	switch which {
	case MiscMapIndexShrine:
		return MiscMapEnterY - MiscMapWalkShrine
	case MiscMapIndexCodex:
		return MiscMapEnterY - MiscMapWalkCodex
	case MiscMapIndexThrone:
		return MiscMapThroneY
	}
	return MiscMapEnterY
}

// MiscMapSet 的索引(與 MiscMapOffsets 同順序)。
const (
	MiscMapIndexCell = iota
	MiscMapIndexShrine
	MiscMapIndexCodex
	MiscMapIndexThrone
)

// MiscMapLocation 是石室期間的地點編號(原版 `byte_3E0A3 = 0xFF`)。
//
// ⚠ 用 0xFF 而不是 0 —— 有很多判斷是 `地點 > 0x7F` 才成立的
//(例如魔法的「戰鬥中才能施」),進石室時那些判斷要跟原版一樣。
const MiscMapLocation = 0xFF

// WalkStep 把一個物件朝目標走一格(原版 `sub_134CC`)。
//
// 回傳走完之後的座標,以及「還在走嗎」。到了目標(或起點就是目標)回 false。
//
// 原版每呼叫一次走一格、重畫一幀,呼叫端用 `while (WalkStep(...))` 把它走完 ——
// 石室的進場與退場、結局那一幕的走位都是這一支。
//
// ⚠ **平手時走橫向。** 原版是 `cmp dy, dx; jle 走橫向` —— 距離相等
// (正斜角)時往橫的走,不是縱的。抄反了路徑會不一樣:
// 石室進場是從 (5,10) 往上走,一路 dx = 0,不受影響;
// 但結局那一幕是斜著走的,差別看得出來。
func WalkStep(x, y, targetX, targetY int) (int, int, bool) {
	if x == targetX && y == targetY {
		return x, y, false
	}
	dx, dy := abs(x-targetX), abs(y-targetY)
	if dy > dx {
		if targetY < y {
			y--
		} else {
			y++
		}
		return x, y, true
	}
	if x > targetX {
		x--
	} else {
		x++
	}
	return x, y, true
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// WalkPath 是把 `WalkStep` 走到底得到的整條路徑(不含起點)。
//
// 上限只是防呆:11×11 的石室走不了那麼多步,真的爆掉代表 WalkStep 有問題。
func WalkPath(x, y, targetX, targetY int) [][2]int {
	var out [][2]int
	for i := 0; i < MiscMapSide*MiscMapSide; i++ {
		nx, ny, moving := WalkStep(x, y, targetX, targetY)
		if !moving {
			break
		}
		x, y = nx, ny
		out = append(out, [2]int{x, y})
	}
	return out
}
