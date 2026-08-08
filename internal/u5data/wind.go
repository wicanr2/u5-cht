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
// ⚠⚠ **這一段整個更正過**(`docs/re/84` §3)。原本寫著
// 「風 1 → dy = −1 北 / 風 2 → dy = +1 南 / 風 3 → dx = −1 西 / 風 4 → dx = +1 東,
// 也就是風值代表它把你推向哪邊」—— **X 與 Y 認錯了,而且 3 / 4 的名字對調**。
//
// 定案靠**三個獨立來源逐一相符**(一張「值 → 方向」的表至少要兩個來源才敢定):
//
//  1. 原版自己印的名字(`sub_2A984` 的 `jpt_2A9E4`):
//     0 Calm / 1 North / 2 South / **3 East** / **4 West**
//  2. 向量表 `byte_601C7`(X)與 `byte_601CB`(Y),逐位元組展開
//     `0x601C7..0x601CF` = `00 00 00 FF 01 | 01 FF 00 00`
//     ⇒ 1→(0,+1) 2→(0,−1) 3→(−1,0) 4→(+1,0)
//     (哪張是 X 由 `sub_2D174` 的方向編碼定案:case 3 做 `ebx--` 且呼叫
//     `sub_2CCFC(0)` = North ⇒ `ebx` 是 Y,而 `sub_2D2D0` 用 `var_4` 接
//     case 1/2 並對 `byte_601C7` ⇒ `byte_601C7` 是 X)
//  3. Rel Hur 的方向 → 風值(`sub_1CDA4` 的 `jpt_1CDD0`):
//     西→4、東→3、北→1、南→2
//
// ⇒ **風的名字是它「從哪裡來」,向量是它「把你推向哪裡」。**
// 「北風把你往南推」—— 氣象學上正確,而三個來源全部相符。

// 風的五種狀態。⚠ **3 是東風、4 是西風**(原版印的名字就是這樣)。
const (
	WindCalm  = 0
	WindNorth = 1
	WindSouth = 2
	WindEast  = 3
	WindWest  = 4
	WindCount = 5
)

// WindNameZH 是風向的中文名(= 風從哪裡來)。
var WindNameZH = [WindCount]string{"無風", "北風", "南風", "東風", "西風"}

// WindDelta 回傳這個風把東西往哪推。
//
// ⚠ 與名字**相反**:北風往南推。這不是筆誤,是氣象學的慣例,
// 而原版的向量表就是這樣寫的。
func WindDelta(wind int) (dx, dy int) {
	switch wind {
	case WindNorth:
		return 0, 1
	case WindSouth:
		return 0, -1
	case WindEast:
		return -1, 0
	case WindWest:
		return 1, 0
	}
	return 0, 0
}

// 船的四個朝向 tile(`sub_2D38` 的 `(kind & 0xFC) == 0x2C`,朝向 = kind − 0x2C)。
//
// 朝向與哪一種風最搭,是從延遲表反推的 —— 每個朝向都正好有**一種**風給它 2:
//
//	朝向 0(0x2C)最搭 風 1 北  ⇒ 朝北的船
//	朝向 1(0x2D)最搭 風 3 東  ⇒ 朝東的船
//	朝向 2(0x2E)最搭 風 2 南  ⇒ 朝南的船
//	朝向 3(0x2F)最搭 風 4 西  ⇒ 朝西的船
//
// ⇒ 0x2C..0x2F 的順序就是 **北 東 南 西**。
//
// ⚠⚠ 原本這裡寫「北 西 南 東,不是常見的北東南西」——**那是連帶錯的**。
// 錯因:風值 3 / 4 的名字當初抄反了(見上面的更正),於是「朝向 1 最搭風 3」
// 被讀成「最搭西風」。兩個錯互相掩護,延遲表的對角線照樣對得起來,
// 所以自我一致性檢查抓不到。
//
// ★ 定案的第五個佐證:`sub_2CCFC` 的 `byte_3E08C = (載具 & 0FCh) + 方向`,
// 而那裡的方向是 0-based 跳表 **0=North 1=East 2=South 3=West**
// (`jpt_2CDF2` 的四個 case 直接印方向名)。低兩位就是朝向 ⇒ 北東南西。
const (
	ShipTileBase   = 0x2C
	ShipFacings    = 4
	ShipFacingN    = 0
	ShipFacingE    = 1
	ShipFacingS    = 2
	ShipFacingW    = 3
	ShipNeverMoves = 4
)

// windDelay 是 4 朝向 × 5 風的延遲表(DOS `0x2C08`,u16;FM Towns `0x4FD50`,u32)。
//
// 值的意思是「隔幾拍才動一格」:**2 最快、3 慢、4 = 完全不動**。
// 風 0(無風)那一欄不會被讀到 —— 原版在查表前就先 `jz` 掉了。
//
//	          風1北 風2南 風3東 風4西
//	朝向 0 北   2     3     4     4
//	朝向 1 東   4     4     2     3
//	朝向 2 南   3     2     4     4
//	朝向 3 西   4     4     3     2
//
// ⚠ 「動不了」的是**橫向**的風,不是反向的。朝北的船遇南風是 3
//(同一軸線,還能搶風調向,只是慢);遇東西風才是 4。
// 直覺會以為「反向風把船擋死」—— 表說的是「同軸線的風都能用,一快一慢;
// 垂直軸線的風完全沒用」。這條是測試打臉之後才改對的。
//
// ⚠⚠ **這張表是敵船的,不是玩家的**(`docs/re/84` §1–2)。它只被
// `sub_2D38` 讀,而那支只處理物件槽。而且值的極性也反了:
// **`4` 是每回合都動(最快),不是動不了** —— 原版是
// `if (延遲 == 4) → 移動`,以及「計數器 ≤ 值就移動,超過才跳一回合」。
// ⇒ 上面「2 最快 / 4 完全不動」的說明**是錯的**,留著是為了對照;
// 正確語意見 `docs/re/84` §1。引擎目前還沒有敵船,所以這張表暫時只當
// 「哪個朝向配哪個風」的對稱性佐證用。
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

// Delay 取某個朝向在某種風下的表值。
//
// ⚠⚠ **目前沒有任何非測試程式碼呼叫它,而那是有理由的**:這張表是
// **敵船**的速度表(只被原版 `sub_2D38` 讀),而引擎還沒有敵船
// (`WORKLIST §5.1b`)。`CanSail` 此前誤用了它,已改掉(`docs/re/84` §2)。
//
// 保留的兩個理由:(1) 實作敵船時就是它;(2) `LoadWindDelay` 的自我一致性
// 檢查是**風向命名的 oracle** —— 表的對角線只有在朝向與風向都用原版順序
// (北東南西 / North South East West)時才對得齊,那是第五個佐證。
//
// ⚠ 名字裡的「Delay」是抄錯極性時留下的:值越大越快(`4` = 每回合都動)。
// 等敵船實作時一起改名,現在改會動到那個 oracle 檢查。
//
// ShipNeverMoves(4)這個名字同樣是錯的極性,見上面表的說明。
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

// ShipFacingForDelta 把一步的位移換成船的朝向碼(原版 `sub_2870` 的四個 cmp)。
//
//	(0, −1) → 北 0    (+1, 0) → 東 1    (0, +1) → 南 2    (−1, 0) → 西 3
//
// ⚠ 與 `ShipFacingForDirection` 的差別:那一支對「非四方向」的輸入會落到
// 東,這一支照原版的順序逐一比對,對不上就回北(原版的 `ebx` 初值 0)。
func ShipFacingForDelta(dx, dy int) int {
	switch {
	case dx == 0 && dy == -1:
		return ShipFacingN
	case dx == 1 && dy == 0:
		return ShipFacingE
	case dx == 0 && dy == 1:
		return ShipFacingS
	case dx == -1 && dy == 0:
		return ShipFacingW
	}
	return ShipFacingN
}
