package u5data

import "strings"

// 力量之言(`sub_17CFC`,2026-08-07)
//
// 八個字,一個字對一座地牢、也對一座聖壇 —— **索引是共用的**:
//
//	0 FALLAX    欺瞞 / 誠實      4 AVIDUS     貪婪 / 犧牲
//	1 VILIS     輕蔑 / 慈悲      5 INFAMA     羞恥 / 榮譽
//	2 INOPIA    毀滅 / 勇氣      6 IGNAVUS    海斯洛斯 / 靈性
//	3 MALUM     謬誤 / 正義      7 VERAMOCOR  末日 / 謙遜
//
// ★ 「索引共用」不是推論。`sub_17CFC` 拿同一個 `edi` 去查
// `off_55DF8`(字)、`byte_41114`/`byte_4113C`(地牢入口座標)、
// `byte_55E18`(地牢入口的地形),而它呼叫 `sub_17C2C(edi, …)` 時,
// 那一支又用同一個 `edi` 去查 `off_411BC`(美德名)、`off_411DC`(真言)、
// `byte_411FC`(聖壇座標)。一路到底都是同一個索引。
//
// ★ 而且 `byte_41114` / `byte_4113C` 與 `docs/re/18` 從 `sub_2D564` 讀出來的
// **八座地牢入口座標一個不差** —— 兩份獨立來源對上。

// WordsOfPower 是八個力量之言。
//
// **[HARD] 這是玩家要打出來的字,永遠維持英文。**
var WordsOfPower = [VirtueCount]string{
	"FALLAX", "VILIS", "INOPIA", "MALUM",
	"AVIDUS", "INFAMA", "IGNAVUS", "VERAMOCOR",
}

// VirtueNames 是完整的英文美德名(`off_411BC`)。
//
// ⚠ 第 2 個是 **`Valour`**(英式拼法),不是 `Valor` ——
// 而 `u5data.Shrines` 那邊的 `Name` 是 `Valor`。原版兩處拼法就不同,
// 比對只看前四個字母所以不影響,但抄的時候別「順手統一」。
var VirtueNames = [VirtueCount]string{
	"Honesty", "Compassion", "Valour", "Justice",
	"Sacrifice", "Honor", "Spirituality", "Humility",
}

// 世界地圖上與聖壇、地牢封印有關的三個地形
const (
	// TileShrine 是聖壇(完好的)。
	//
	// 依據:`sub_17C2C` 復原成功時 `mov byte ptr [eax], 19h`。
	// 佐證:走到慈悲聖壇 (128,92) 時狀態列顯示的地形正是 25 = 0x19。
	TileShrine = 0x19
	// TileShrineDesecrated 是被玷污的聖壇。
	//
	// `sub_17CFC` 在鄰格看到 0x1A 時走「復原聖壇」那條路。
	TileShrineDesecrated = 0x1A
	// TileCodex 是終極智慧之寶典所在的那一格(`sub_2D72C` 的 17 = 法典聖壇)。
	//
	// 與八德聖壇共用同一支進場函式 `sub_1DA10`,靠這個地形分辨進去要做什麼。
	TileCodex = 0x11
	// TileDungeonSealed 是被封印的地牢入口。
	//
	// 切換的算式是 `格子 ^= (DungeonEntranceTile[i] ^ 0xDF)` ——
	// 一個 XOR 同時做開與關,所以 0xDF 與原本的地形互為彼此。
	TileDungeonSealed = 0xDF
)

// DungeonEntranceTile 是八座地牢入口在世界地圖上原本的地形(`byte_55E18`)。
//
// 三種值:0x16 / 0x17 / 0x18 —— 洞口開在不同的地形上。
var DungeonEntranceTile = [VirtueCount]byte{0x18, 0x16, 0x16, 0x18, 0x18, 0x17, 0x17, 0x16}

// WordOfPowerIndex 查一個玩家打的字是第幾個力量之言;不是就回 −1。
func WordOfPowerIndex(spoken string) int {
	for i, w := range WordsOfPower {
		if MatchPrefix(w, spoken) {
			return i
		}
	}
	return -1
}

// MatchNeedleMax 是比對時參考字被截掉的長度(`sub_27C98` 的 `cmp ebx, 9`)。
//
// ⚠ 這個 9 不是湊的:`Compassion`(10)與 `Spirituality`(12)都比它長,
// 所以原版實際上只比到 `COMPASSIO` / `SPIRITUAL`。照抄。
const MatchNeedleMax = 9

// MatchPrefix 是原版全遊戲共用的字串比對(`sub_27C98` → `sub_39554` → `sub_39C50`)。
//
// 語意是**前綴**,不是相等、也不是子字串:參考字先截到 9 個字元,
// 再拿去比玩家打的那一行的開頭。所以:
//
//	真言 Ahm    → 打 `ahmxyz` 也算(原版就是這樣)
//	美德 hone   → 打 `honesty` 算,打 `the honesty` **不算**(不是開頭)
//
// ⚠ **大小寫:原版只把參考字轉大寫,玩家打的那一行原樣拿去 `repe cmpsb`。**
// 也就是說,如果輸入層沒有幫忙轉大寫,原版就得全大寫打才會過。輸入層
// (`sub_2B770` / `sub_28F40`)有沒有轉大寫還沒追到,所以這裡取**不分大小寫** ——
// 這是一個超集:原版能過的這裡一定過,不會擋掉打對的玩家。
// 追出來之後若確定要區分大小寫,改這一支就好(全遊戲共用)。
func MatchPrefix(needle, typed string) bool {
	if len(needle) > MatchNeedleMax {
		needle = needle[:MatchNeedleMax]
	}
	if len(typed) < len(needle) {
		return false
	}
	return lowerASCII(typed[:len(needle)]) == lowerASCII(needle)
}

// lowerASCII 是只處理 ASCII 的小寫化(這些字全是 ASCII,不需要 Unicode 規則)。
func lowerASCII(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] += 'a' - 'A'
		}
	}
	return string(b)
}

// 暗影君主(`sub_17A14`)
//
// 三個名字放在 `off_55DEC`,與八個力量之言的 `off_55DF8` 相鄰 ——
// 兩組都由**同一個指令**(Yell,`sub_17E74`)分派出去。

// ShadowlordCount 是暗影君主的數量。
const ShadowlordCount = 3

// Shadowlords 是三個暗影君主的名字。
//
// **[HARD] 這是玩家要喊出來的字,永遠維持英文。**
//
// 對應的三惡與三火(語意來自遊戲背景,不是執行檔):
//
//	0 FAULINEI  虛偽 ⇄ 真理之火    1 ASTAROTH  憎恨 ⇄ 愛之火
//	2 NOSFENTOR 怯懦 ⇄ 勇氣之火
var Shadowlords = [ShadowlordCount]string{"FAULINEI", "ASTAROTH", "NOSFENTOR"}

// ShadowlordKeeps 是喊名字會有反應的三個地點編號(`sub_17A14` 的 `cmp al, 1Eh/1Fh/20h`)。
//
// 30 / 31 / 32 = 學術之城 / 慈悲修道院 / 巨蛇要塞 —— 正是三團聖火所在。
//
// ⚠ **原版不檢查「名字要配這座城」。** 在任何一座喊任何一個名字都會召來
// 那一個暗影君主。別自作主張加上配對檢查。
var ShadowlordKeeps = [ShadowlordCount]int{30, 31, 32}

// TileShadowlord 是暗影君主的 tile(`mov byte_3EDB0[ebx], 0FCh`)。
//
// 0xFC → 生物編號 (0xFC − 0x40) / 4 = 47 = `CreatureShadowLord`。
const TileShadowlord = 0xFC

// ShadowlordGone 是「這個暗影君主已經被消滅」(`sub_1A38C` 寫 0xFF)。
const ShadowlordGone = 0xFF

// ShadowlordNone 是「現在沒有召喚中的暗影君主」(`byte_3E0DB`,存檔初值就是 0xFF)。
const ShadowlordNone = 0xFF

// ShadowlordIndex 查喊出來的名字是第幾個暗影君主;不是就回 −1。
func ShadowlordIndex(spoken string) int {
	for i, n := range Shadowlords {
		if MatchPrefix(n, spoken) {
			return i
		}
	}
	return -1
}

// ShadowlordSpawnAhead 是召喚出來的暗影君主出現在玩家上方幾格
//(`sub_17A14` 的 `sub eax, 2`;同一支也因此要求 Y ≥ 2)。
const ShadowlordSpawnAhead = 2

// ShrineDesecrated 是 `byte_3E0E8[i]` 裡「已被玷污」的位元。
const ShrineDesecratedBit = 0x80

// DungeonSealedBit 是 `byte_3E0E0[i]` 的 bit 7。
//
// ⚠⚠ **極性與名字直覺相反:這個位元「設著」代表入口是通的,「清掉」才是崩塌。**
// 四條獨立證據(`docs/re/99` §5e):
//
//  1. `sub_1056C` 是 `cmp byte_3E0E0[edx], 0; setz al` —— **等於 0 才回真**,
//     而回真那一條就是把入口地形改寫成 `0xDF`(崩塌)。
//  2. `INIT.GAM` 與 `SAVED.GAM` 的這八個位元組**全是 0** ⇒ 開局全部崩塌。
//  3. `byte_55E18`(八座入口的原始地形)= `18 16 16 18 18 17 17 16`,
//     正是 `sub_105E4` 會改寫成 `0xDF` 的那三個值(0x16 洞穴 / 0x17 礦坑 /
//     0x18 地牢)。⇒ 兩支函式講的是同一件事。
//  4. tile `0xDF` 的 look 文字是「the collapsed entrance to the dungeon」——
//     那是**預設會看到**的敘述,不是特例。
//
// ⇒ 喊力量之言(`xor byte_3E0E0[i], 80h`)是把 0 變成 0x80 = **開封**,
// 不是封印。這正是 U5 的主線:八座地牢一開始進不去,要各自喊對那個字。
const DungeonSealedBit = 0x80

// DungeonIsSealed 回報第 i 座地牢的入口是不是崩塌的。
//
// ⚠ 一定要走這一支,不要自己寫 `flag & DungeonSealedBit != 0` ——
// 那個直覺寫法**方向剛好相反**(見上面)。
func DungeonIsSealed(flag byte) bool { return flag&DungeonSealedBit == 0 }

// YellInputMax 是 Yell 指令讀幾個字元(`sub_17E74` 的 `push 0Ch`)。
const YellInputMax = 12

// MantraSpoken 回報玩家打的這一行裡**有沒有出現**這句真言。
//
// ⚠ 黑棘的審問(`sub_C098`)用的**不是** `MatchPrefix`,而是一個
// 不分大小寫的**子字串**搜尋:它拿 `edi` 從 0 開始一格一格滑過玩家打的字,
// 每一格再逐字比對真言。所以「the mantra is Ahm」也算招供 ——
// 想含糊帶過是躲不掉的。
//
// 寫成前綴比對的話,玩家只要在真言前面加一個字就能白嫖過關,
// 而那正是這一幕唯一的抉擇。
func MantraSpoken(mantra, typed string) bool {
	return strings.Contains(lowerASCII(typed), lowerASCII(mantra))
}

// ToggleDungeonSeal 回傳這一格封印切換之後變成什麼。
//
// `格子 ^= (原地形 ^ 0xDF)` —— 一個 XOR 兩個方向都對。
func ToggleDungeonSeal(tile byte, dungeon int) byte {
	if dungeon < 0 || dungeon >= VirtueCount {
		return tile
	}
	return tile ^ (DungeonEntranceTile[dungeon] ^ TileDungeonSealed)
}

// WordTargetTile 回報這個地形是不是第 i 個力量之言認得的目標。
//
// 三種:地牢入口原本的地形、被封印的 0xDF、被玷污的聖壇 0x1A。
func WordTargetTile(tile byte, i int) bool {
	if i < 0 || i >= VirtueCount {
		return false
	}
	return tile == DungeonEntranceTile[i] ||
		tile == TileDungeonSealed || tile == TileShrineDesecrated
}

// NameMatchLen 是報名字時比對的長度(原版 `sub_1C2FC` 只複製 4 個位元組)。
//
// ⚠ 與對話關鍵字的 9(`MatchNeedleMax`)不同 —— 同一支比對函式,
// 不同的截斷長度。用 9 的話「Elwo」報不進去(名字是 Elwood 時);
// 用 4 才對得上原版。
const NameMatchLen = 4

// NameSpoken 回報玩家報的名字對不對得上某個隊員(原版 `sub_1C2FC`)。
//
// 規則:把成員的名字截到 4 個字元當 needle,玩家的輸入要以它開頭
// (不分大小寫)。所以隊裡有 Elwood 時,打「Elwo」「Elwood」「Elwoodx」都算,
// 打「Elw」不算。
func NameSpoken(member, typed string) bool {
	if member == "" {
		return false
	}
	if len(member) > NameMatchLen {
		member = member[:NameMatchLen]
	}
	return MatchPrefix(member, typed)
}
