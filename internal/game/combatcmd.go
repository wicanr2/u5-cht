package game

// 戰鬥中的指令表(原版 `sub_A360` 的 `jpt_A5C8`)
//
// 戰鬥有**自己的**分派器,與地面 / 場景 / 地牢那張(`sub_2ACF4`,見 `docs/re/49`)
// 是兩份程式。可用的少很多,而**不可用的每一個都有自己的回應**——
// 不是統一一句「不行」:

// CombatRefusal 是被拒絕時要接在指令名後面的話(原版 `sub_16270(名字, 種類)`)。
type CombatRefusal int

// 三種尾綴 + 「什麼都不接」。
const (
	// RefuseNone 只印指令名。
	RefuseNone CombatRefusal = 0
	// RefuseWhat 接 " what?" —— 拿來問「要對什麼做」的指令(Board / X-it)。
	RefuseWhat CombatRefusal = 1
	// RefuseNotHere 接 "-Not here" —— 這裡做不了(Enter / Fire / Hole up …)。
	RefuseNotHere CombatRefusal = 2
	// RefuseNoResponse 接 "-Funny, no response!" —— 戰鬥中沒人跟你講話(Talk)。
	RefuseNoResponse CombatRefusal = 3
)

// combatRefusals 是戰鬥中**不能用**的字母鍵,以及各自的回應。
//
// 取自 `jpt_A5C8` 逐 case:這些 case 都只走 `sub_16270(名字, 種類)`。
//
// ★ D 與 W 在這裡也是 `D-What?` / `W-What?` —— 與主分派器一致,
// 兩處獨立佐證它們不是指令(`docs/re/49` §2)。
var combatRefusals = map[rune]struct {
	Name   string
	Reason CombatRefusal
}{
	'B': {"上船", RefuseWhat},
	'D': {"D", RefuseNone},
	'E': {"進入", RefuseNotHere},
	'F': {"開砲", RefuseNotHere},
	'H': {"紮營", RefuseNotHere},
	'I': {"點火把", RefuseNotHere},
	'L': {"觀察", RefuseNotHere},
	'M': {"調藥", RefuseNotHere},
	'N': {"換位", RefuseNotHere},
	'Q': {"存檔", RefuseNotHere},
	'T': {"交談", RefuseNoResponse},
	'V': {"看寶石", RefuseNotHere},
	'W': {"W", RefuseNone},
	'X': {"下載具", RefuseWhat},
}

// CombatAllowedKeys 是戰鬥中**能用**的字母鍵(`jpt_A5C8` 有實作的那些)。
//
//	A 攻擊  C 施法  G 撿  J 撬鎖  K 攀爬  O 開  P 推  R 換裝  S 搜尋  U 用道具
//	Y 喊    Z 數值
//
// ★ 這幾支**與地圖上是同一份程式**,不需要「戰場版」——
// 原版只有一對座標(`byte_3E0A6/A7`),`sub_A360` 進來就把它設成行動中那個
// 單位在戰場上的位置;`sub_DB10(x, y)` 再依地點編號決定要讀世界、場景
// 還是戰場緩衝(`cmp byte_3E0A3, 7Fh; ja → byte_3F8F4`)。
//
// 引擎照這個形狀補了三處(`docs/re/52`):`TileAt`/`SetTileAt` 加戰鬥分支、
// 回合開始時 `focusCombatUnit` 把 `State.X/Y` 借給行動者、離場時還原。
// 補完之後這九個指令**一行都沒改**就能在戰場上運作 ——
// 先前以為要另寫「戰場版」,那是因為我把座標與地圖各拆成了兩套。
//
// K(攀爬)在戰鬥中走的是**另一支** `sub_16058`(與地圖上的 `sub_188C4`/
// `sub_EA0`/`sub_417C` 都不同),已另外實作在 `combatklimb.go`。
var CombatAllowedKeys = []rune{'A', 'C', 'G', 'J', 'K', 'O', 'P', 'R', 'S', 'U', 'Y', 'Z'}

// CombatRefuse 回報這個鍵在戰鬥中被拒絕時該印什麼;第二個回傳值為 false
// 代表這個鍵不在拒絕清單裡(可能可用,也可能是原版的 default「此為何意?」)。
func CombatRefuse(key rune) (string, bool) {
	e, ok := combatRefusals[key]
	if !ok {
		return "", false
	}
	switch e.Reason {
	case RefuseWhat:
		return e.Name + " —— 對什麼?", true
	case RefuseNotHere:
		return e.Name + " —— 此處不可。", true
	case RefuseNoResponse:
		return e.Name + " —— 無人回應。", true
	}
	return e.Name + " —— 何事?", true
}
