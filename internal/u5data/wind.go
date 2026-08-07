package u5data

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

// 風向與航行
//
// 風有五種狀態:**0 無風**,1..4 各推一個方向。原版把它放在 `byte_3E0A2`,
// 由 `sub_2A984(風)` 設定(順便把變化計時器 `byte_3E093` 歸零)。
//
// 方向是從船的漂流那一段(`sub_2D2D0`)讀出來的 —— 那裡依風值算 (dx, dy):
//
//	風 1 → dy = −1  北
//	風 2 → dy = +1  南
//	風 3 → dx = −1  西
//	風 4 → dx = +1  東
//
// 也就是**風值代表它把你推向哪邊**,不是它從哪邊來。

// 風的五種狀態。
const (
	WindCalm  = 0
	WindNorth = 1
	WindSouth = 2
	WindWest  = 3
	WindEast  = 4
	WindCount = 5
)

// WindNameZH 是風向的中文名。
var WindNameZH = [WindCount]string{"無風", "北風", "南風", "西風", "東風"}

// WindDelta 回傳這個風把東西往哪推。
func WindDelta(wind int) (dx, dy int) {
	switch wind {
	case WindNorth:
		return 0, -1
	case WindSouth:
		return 0, 1
	case WindWest:
		return -1, 0
	case WindEast:
		return 1, 0
	}
	return 0, 0
}

// 船的四個朝向 tile(`sub_2D38` 的 `(kind & 0xFC) == 0x2C`,朝向 = kind − 0x2C)。
//
// 朝向與哪一種風最搭,是從延遲表反推的 —— 每個朝向都正好有**一種**風給它 2:
//
//	朝向 0(0x2C)最搭 風 1 北  ⇒ 這是朝北的船
//	朝向 1(0x2D)最搭 風 3 西
//	朝向 2(0x2E)最搭 風 2 南
//	朝向 3(0x2F)最搭 風 4 東
//
// 所以 0x2C..0x2F 的順序是 **北 西 南 東**,不是常見的北東南西。
const (
	ShipTileBase   = 0x2C
	ShipFacings    = 4
	ShipFacingN    = 0
	ShipFacingW    = 1
	ShipFacingS    = 2
	ShipFacingE    = 3
	ShipNeverMoves = 4
)

// windDelay 是 4 朝向 × 5 風的延遲表(DOS `0x2C08`,u16;FM Towns `0x4FD50`,u32)。
//
// 值的意思是「隔幾拍才動一格」:**2 最快、3 慢、4 = 完全不動**。
// 風 0(無風)那一欄不會被讀到 —— 原版在查表前就先 `jz` 掉了。
//
//	          風1北 風2南 風3西 風4東
//	朝向 0 北   2     3     4     4
//	朝向 1 西   4     4     2     3
//	朝向 2 南   3     2     4     4
//	朝向 3 東   4     4     3     2
//
// ⚠ 「動不了」的是**橫向**的風,不是反向的。朝北的船遇南風是 3
//(同一軸線,還能搶風調向,只是慢);遇東西風才是 4。
// 直覺會以為「反向風把船擋死」—— 表說的是「同軸線的風都能用,一快一慢;
// 垂直軸線的風完全沒用」。這條是測試打臉之後才改對的。
const windDelayTable = 0x2C08

// WindDelay 是 4×5 的延遲表。
type WindDelay [ShipFacings][WindCount]int

// LoadWindDelay 從 DATA.OVL 讀延遲表。
func LoadWindDelay(dir string) (*WindDelay, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "DATA.OVL"))
	if err != nil {
		return nil, err
	}
	return ParseWindDelay(raw)
}

// ParseWindDelay 解析延遲表。
//
// ⚠ 檔案裡只有 **4 朝向 × 4 風**(風 1..4);風 0 那一欄是我們補的
// `ShipNeverMoves` —— 原版根本不查它。
func ParseWindDelay(ovl []byte) (*WindDelay, error) {
	need := windDelayTable + ShipFacings*4*2
	if len(ovl) < need {
		return nil, fmt.Errorf("DATA.OVL 只有 %d B,放不下風向延遲表", len(ovl))
	}
	w := &WindDelay{}
	for f := 0; f < ShipFacings; f++ {
		w[f][WindCalm] = ShipNeverMoves
		for k := 1; k < WindCount; k++ {
			off := windDelayTable + (f*4+(k-1))*2
			v := int(binary.LittleEndian.Uint16(ovl[off:]))
			if v < 2 || v > ShipNeverMoves {
				return nil, fmt.Errorf("朝向 %d 風 %d 的延遲是 %d,預期 2..4", f, k, v)
			}
			w[f][k] = v
		}
	}
	return w, w.validate()
}

// validate 用延遲表的對稱性擋住位移偏掉。
func (w *WindDelay) validate() error {
	// 每個朝向都要正好有一種風給它 2(順風)、一種給 3(側風)、兩種 4。
	for f := 0; f < ShipFacings; f++ {
		var n [5]int
		for k := 1; k < WindCount; k++ {
			n[w[f][k]]++
		}
		if n[2] != 1 || n[3] != 1 || n[4] != 2 {
			return fmt.Errorf("朝向 %d 的延遲分布是 2×%d 3×%d 4×%d,預期 1/1/2",
				f, n[2], n[3], n[4])
		}
	}
	// 每個朝向的順風就是它自己的方向;而垂直軸線的兩種風一定是 4。
	for _, c := range []struct{ facing, wind int }{
		{ShipFacingN, WindNorth}, {ShipFacingS, WindSouth},
		{ShipFacingW, WindWest}, {ShipFacingE, WindEast},
	} {
		if w[c.facing][c.wind] != 2 {
			return fmt.Errorf("朝向 %d 遇上它的順風 %d 卻是延遲 %d,預期 2",
				c.facing, c.wind, w[c.facing][c.wind])
		}
	}
	// 南北向的船遇東西風、東西向的遇南北風,一律動不了。
	cross := map[int][2]int{
		ShipFacingN: {WindWest, WindEast}, ShipFacingS: {WindWest, WindEast},
		ShipFacingW: {WindNorth, WindSouth}, ShipFacingE: {WindNorth, WindSouth},
	}
	for f, ws := range cross {
		for _, k := range ws {
			if w[f][k] != ShipNeverMoves {
				return fmt.Errorf("朝向 %d 遇橫風 %d 延遲 %d,預期動不了", f, k, w[f][k])
			}
		}
	}
	return nil
}

// Delay 取某個朝向在某種風下的延遲;ShipNeverMoves 代表動不了。
func (w *WindDelay) Delay(facing, wind int) int {
	if w == nil || facing < 0 || facing >= ShipFacings || wind < 0 || wind >= WindCount {
		return ShipNeverMoves
	}
	return w[facing][wind]
}

// ShipFacingForDirection 把一個方向換成船的朝向碼(0x2C..0x2F 的低兩位)。
func ShipFacingForDirection(dx, dy int) int {
	switch {
	case dy < 0:
		return ShipFacingN
	case dy > 0:
		return ShipFacingS
	case dx < 0:
		return ShipFacingW
	default:
		return ShipFacingE
	}
}
