package u5data

// 環境音:靠得夠近的東西會自己出聲(原版 `sub_2BDE0`)
//
// 推導見 `docs/re/92`。原版每次重繪地圖就掃隊伍周圍 **11×11**,
// 找**距離平方最小**的那一個發聲物,只讓它出聲 —— 所以同時只會有一種環境音。
//
// ⚠ 這不是「進場景時決定一次」:走近走遠都會即時換,離開範圍就靜下來。

// AmbientKind 是四類發聲物。順序照原版 `ebx` 的值(1..4),0 = 附近沒有。
type AmbientKind int

const (
	AmbientNone      AmbientKind = 0
	AmbientClock     AmbientKind = 1 // 落地鐘:滴答 + 第 4 相的鐘響
	AmbientWaterfall AmbientKind = 2 // 瀑布
	AmbientFountain  AmbientKind = 3 // 噴泉
	AmbientMusic     AmbientKind = 4 // ★ 會**把配樂停掉**改放蜂鳴器旋律的那個物件
)

// AmbientScanRadius 是掃描半徑。
//
// 原版 `for (x = 中心X − 5; x < 中心X + 6; x++)` ⇒ 11×11。
// ★ 這剛好是 U5 的視野大小(`byte_3F844` 是 11×11、列距 16)——
// **看得到的東西才會出聲**。
const AmbientScanRadius = 5

// AmbientMaxDistSq 是「還算在附近」的距離平方上限(原版 `var_1C` 的初值 0x33)。
//
// 51 = 5² + 5² + 1 ⇒ 剛好把 11×11 的四個角(50)包進來,而角落外的都排除。
// 迴圈用 `if (d² >= 目前最小) continue` 逐步收緊,所以初值同時是上限。
const AmbientMaxDistSq = 0x33

// AmbientTileKind 回傳這一格地形是不是發聲物(原版三個遮罩比較)。
//
//	(tile & 0xFE) == 0xFA → 落地鐘   (0xFA, 0xFB)
//	(tile & 0xFC) == 0xD4 → 瀑布     (0xD4..0xD7,四格動畫)
//	(tile & 0xFC) == 0xD8 → 噴泉     (0xD8..0xDB,四格動畫)
//
// ⚠ 第四類(`AmbientMusic`)不看地形層,看**疊圖層**且要求那一格可見
// ⇒ 由 `AmbientOverlayIsMusic` 另外判。
func AmbientTileKind(tile int) AmbientKind {
	switch {
	case tile&0xFE == 0xFA:
		return AmbientClock
	case tile&0xFC == 0xD4:
		return AmbientWaterfall
	case tile&0xFC == 0xD8:
		return AmbientFountain
	}
	return AmbientNone
}

// AmbientOverlayIsMusic 判疊圖層那一格是不是「會放旋律」的物件。
//
//	(疊圖 tile & 0xFC) == 0x5C
//
// ⚠ 這個遮罩涵蓋 **0x5C..0x5F 四格**:0x5C/0x5D 是書架的左右半,
// 0x5E/0x5F 是一個兩格寬的樂器形物件(LOOK 表寫 "a Guardian!")。
// 原版就是這樣遮的,**不縮小範圍** —— 要縮就是自創(`CLAUDE.md §3.0`)。
func AmbientOverlayIsMusic(tile int) bool { return tile&0xFC == 0x5C }

// ClockPhases 是落地鐘滴答的相位數(原版 `byte_60038`,`> 7` 就歸零)。
//
// 每次掃描 +1,所以它是「重繪次數」而不是遊戲時間:
//
//	相位 0 → 滴(sub_2C4F4(0xBB8, 3))
//	相位 4 → 答(sub_2C4F4(0x7D0, 3)),而且**這一相才可能響鐘**
//
// ★ 兩個相位的第一參數不同(0xBB8 vs 0x7D0)⇒ 滴與答**音高不一樣**。
const ClockPhases = 8

const (
	ClockTickPhase = 0 // 滴
	ClockTockPhase = 4 // 答(也是鐘響的相位)
)

// BeeperScale 是蜂鳴器旋律用的九個音(原版 0x60060 起的九個位元組)。
//
//	62 64 66 67 69 71 72 73 74  = D4 E4 F#4 G4 A4 B4 C5 C#5 D5
//
// 是 MIDI 音符號:D 大調音階(升四級)再多一個 C 本位。
// 序列裡的值是 **1-based 索引**,0 表示休止。
var BeeperScale = [9]byte{62, 64, 66, 67, 69, 71, 72, 73, 74}

// BeeperMelody 是那 53 步旋律(原版 0x6006C 起,游標 `>= 0x35` 歸零)。
//
// 每掃描一次走一步,0 是休止符。長度 53 = 0x35,與原版
// `cmp byte ptr [eax+8], 35h; jb` 的上限逐一相符。
//
// ⚠ 已確認**不是** 15 首 `.EUP` 任何一首:把 40 音的前綴拿去比對
// 7,633 個音符,最長只相符 3 音(正對照見 `docs/re/92` §4)。
var BeeperMelody = [53]byte{
	1, 4, 4, 0, 0, 1, 5, 5, 0, 0, 4, 9, 6, 9, 7, 4, 6, 5, 1, 4,
	0, 0, 0, 1, 5, 0, 0, 0, 4, 9, 6, 5, 6, 4, 0, 0, 0, 5, 8, 8,
	9, 5, 6, 8, 9, 5, 4, 6, 5, 4, 3, 2, 1,
}

// BeeperNote 把旋律的一步換成 MIDI 音符號;休止回 0。
func BeeperNote(step byte) byte {
	if step == 0 || int(step) > len(BeeperScale) {
		return 0
	}
	return BeeperScale[step-1]
}
