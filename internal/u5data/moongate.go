package u5data

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

// 月門
//
// 八座月門散在不列顛尼亞各地,每一座就在對應城市的旁邊。踏進去會被送到
// **哪一座**,由當下的月相決定;而月相由**日期**查表得來。
//
//	sub_E2A4:  byte_3E095 = byte_41142[日 * 2]      ← 特拉梅爾(Trammel)
//	           byte_3E096 = byte_41143[日 * 2]      ← 費盧卡(Felucca)
//	sub_E084:  相位 = (小時 < 12) ? byte_3E095 : byte_3E096
//	           sub_DF84(相位 − '0')
//
// ⇒ **上午看一顆月亮、下午看另一顆。** 同一座月門在中午前後會通到不同地方
// —— 這是 U5 月門系統的核心,不是 bug。
//
// ⚠ 表裡存的是 **ASCII 數字 '0'..'7'**,不是 0..7。少減那個 0x30 會直接
// 索引越界。

// moonPhaseTable 是月相表的位址(DOS `DATA.OVL`;FM Towns `0x41142`)。
//
// 29 對 × 2 B,索引是**日**(1..28);第 0 對是填充。
const moonPhaseTable = 0x1EE8

// MoonPhaseCount 是月相的數目。
const MoonPhaseCount = 8

// DaysPerMonth 與時鐘那邊一致(每月 28 天)。
const DaysPerMonth = 28

// MoonPhases 是一張 (Trammel, Felucca) 的表,索引是日(1..28)。
type MoonPhases [DaysPerMonth + 1][2]byte

// LoadMoonPhases 從 DATA.OVL 讀月相表。
func LoadMoonPhases(dir string) (*MoonPhases, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "DATA.OVL"))
	if err != nil {
		return nil, err
	}
	return ParseMoonPhases(raw)
}

// ParseMoonPhases 解析月相表。
func ParseMoonPhases(ovl []byte) (*MoonPhases, error) {
	need := moonPhaseTable + (DaysPerMonth+1)*2
	if len(ovl) < need {
		return nil, fmt.Errorf("DATA.OVL 只有 %d B,放不下月相表", len(ovl))
	}
	m := &MoonPhases{}
	for d := 1; d <= DaysPerMonth; d++ {
		t := ovl[moonPhaseTable+d*2]
		f := ovl[moonPhaseTable+d*2+1]
		// ⚠ 存的是 ASCII '0'..'7'。
		if t < '0' || t > '7' || f < '0' || f > '7' {
			return nil, fmt.Errorf("第 %d 日的月相是 %d/%d,不在 '0'..'7'", d, t, f)
		}
		m[d] = [2]byte{t - '0', f - '0'}
	}
	return m, nil
}

// PhaseAt 回傳某一天某個時刻該用哪一個相位。
//
// 中午之前看 Trammel、之後看 Felucca(`sub_E084` 的 `cmp byte_3E08F, 0Ch`)。
func (m *MoonPhases) PhaseAt(day, hour int) int {
	if m == nil || day < 1 || day > DaysPerMonth {
		return 0
	}
	if hour < 12 {
		return int(m[day][0])
	}
	return int(m[day][1])
}

// 月門的目的地(存檔 `0x028A` 起,四張 8 B 的平行表)
//
// `sub_DF84(相位)` 一次讀四個:
//
//	byte_3E040[相位] → X       byte_3E050[相位] → 地點
//	byte_3E048[相位] → Y       byte_3E058[相位] → 樓層
//
// 位移是從已經驗過的兩個錨點推的:`byte_3DFD0`(裝備)對應 `0x021A`、
// `byte_3E060`(藥草)對應 `0x02AA`,差都是 `0x3DDB6` —— 所以
// `byte_3E040` 對應 `0x3E040 − 0x3DDB6 = 0x028A`,四張表剛好接在藥草前面。
//
// 抽出來的八組座標**每一座都貼著對應的城市**:
//
//	(224,133) 月光城   (96,102) 不列顛城  (38,224) 哲倫    (50,37)  紫衫城
//	(166,19)  米諾克   (104,194) 特林希克 (23,126) 史卡拉布雷 (187,167) 新馬精西亞
//
// 八座全中 —— 這比任何單點比對都硬。
const (
	SaveMoongateX   = 0x028A
	SaveMoongateY   = 0x0292
	SaveMoongateLoc = 0x029A
	SaveMoongateFlr = 0x02A2
)

// MoongateDest 是一個月門相位的目的地。
type MoongateDest struct {
	X, Y     int
	Location int
	Floor    int
}

// Known 回報這個相位通不通。
//
// `sub_DF84` 第一件事就是 `cmp byte_3E050[相位], 0FFh` —— 地點 0xFF
// 代表這個相位還沒開通,直接失敗。
func (d MoongateDest) Known() bool { return d.Location != 0xFF }

// parseMoongates 從存檔取出八個目的地。
func parseMoongates(raw []byte) [MoonPhaseCount]MoongateDest {
	var out [MoonPhaseCount]MoongateDest
	for i := 0; i < MoonPhaseCount; i++ {
		out[i] = MoongateDest{
			X:        int(raw[SaveMoongateX+i]),
			Y:        int(raw[SaveMoongateY+i]),
			Location: int(raw[SaveMoongateLoc+i]),
			Floor:    int(int8(raw[SaveMoongateFlr+i])),
		}
	}
	return out
}

// encodeMoongates 把八個目的地寫回存檔。
func encodeMoongates(out []byte, g [MoonPhaseCount]MoongateDest) {
	for i := 0; i < MoonPhaseCount; i++ {
		out[SaveMoongateX+i] = byte(g[i].X)
		out[SaveMoongateY+i] = byte(g[i].Y)
		out[SaveMoongateLoc+i] = byte(g[i].Location)
		out[SaveMoongateFlr+i] = byte(g[i].Floor)
	}
}

var _ = binary.LittleEndian

// 月門開關的時段(原版 `sub_DEE4`)
//
//	esi = 0xDC                                    ; 預設要寫的 tile = 開著的月門
//	if (小時 >= 0x14 || 小時 < 5) 計數器累加到 0x10  ; ★ 夜裡(20:00–04:59)
//	else { 計數器遞減; if (計數器 == 0) esi = 5 }    ; 白天,歸零才寫回草地
//	for (i = 0; i < 8; i++)
//	    if (sub_DE74(i)) 把 esi 寫進 (byte_3E040[i], byte_3E048[i])
//
// ★★★ 那對座標就是**埋下去的月石的 X / Y**(`byte_3E040` / `byte_3E048`,
// 見 `Save` 的 `SaveMoonstoneXOffset` / `YOffset`)。
// ⇒ **月門長在月石被埋的地方** —— 這是「埋月石有什麼用」的完整答案。
//
// ★ `sub_DE74(i)` **完全不看月相**,只查三件事:那顆月石埋在**當前的地點**、
// 埋在**當前的樓層**、而且(在大地圖時)落在 32×32 的載入視窗內。
//
// ⇒ 三件事各由不同的東西決定,不要混:
//
//	開不開  → 時間(夜裡開)
//	在哪裡  → 月石埋在哪裡
//	去哪裡  → 月相(`docs/re/22`)
const (
	// MoongateNightFrom 是月門開始出現的小時(含)。
	MoongateNightFrom = 0x14
	// MoongateNightUntil 是月門消失的小時(不含)——「小時 < 5」。
	MoongateNightUntil = 5
	// MoongateLingerTicks 是天亮之後月門還留著的重畫次數(`sub_2BBB8` 的上限)。
	//
	// ★ 白天不是立刻關:計數器從 0x10 遞減,歸零才把那一格寫回草地(tile 5)。
	// 所以日出後月門還會殘留一陣子 —— 那是原版的淡出。
	MoongateLingerTicks = 0x10
	// MoongateClosedTile 是月門關上之後寫回去的地形(原版 `mov esi, 5`)。
	MoongateClosedTile = 5
)

// MoongateOpenAtHour 回報這個小時月門開不開(原版 `sub_DEE4` 的第一個判斷)。
func MoongateOpenAtHour(hour int) bool {
	return hour >= MoongateNightFrom || hour < MoongateNightUntil
}
