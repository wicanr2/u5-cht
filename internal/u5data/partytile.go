package u5data

// 隊員在戰場上畫什麼 —— ★ 它**依職業**,而 `docs/re/53` §2 說反了
//
// # 上一輪的結論與它的根據
//
// `docs/re/53` §2 斷言「與職業無關,原版在戰場上只有兩個值:站著 0x1D、躺著 0x1E」,
// 根據是 `sub_2ED50`(起身)與 `sub_2EDF8`(躺下)這兩支成對的函式。
// 那兩支確實寫 0x1D / 0x1E,**而它們寫的是物件記錄的 `+1`**。
//
// # 為什麼它錯了
//
// 讀 `sub_1EC34`(Ready)時撞到第三處寫同一個欄位的地方:
//
//	idx = sub_28F5C("AMBFDTPRS", 角色.職業字母)      ; 職業字母 → 0..8
//	al  = byte_40C34[idx]
//	物件[+1] = al ;  物件[+0] = al                   ; ★ 兩個位元組都寫
//
// 順著 `byte_40C34` 的三個讀取點回查,其中 `sub_C414+1D0` 就是**開戰佈陣的迴圈**
// (同一圈裡還寫了 `+2` 旗標與 `+3` 名冊索引)。⇒ **開戰時的圖是查職業表得來的**,
// 0x1D / 0x1E 只出現在「睡著 / 醒來」與「隱形戒指」那幾條**恢復路徑**上。
//
// 而表本身在 `.asm` 裡被 IDA 拆成一個 `db` 加一個字串(因為 0x40..0x4C 都是可見字元):
//
//	byte_40C34  db 4Ch
//	aDhlllll    db '@DHLLLLL',0
//
// 展開 = `4C 40 44 48 4C 4C 4C 4C 4C`,九筆,對上九個職業字母。
//
// # ★★ 而 0x4C 正是上一輪「修掉」的那個值
//
// `docs/re/52` §5 當初記下一個矛盾:
//
//	`sub_16058`(戰鬥中的 Klimb)判「爬得過去」用的是 tile 0x4C,
//	而本專案 `partyTileFor()` 回的也是 0x4C。兩者衝突,而 `partyTileFor`
//	當初是猜的。⬜ 待查。**先不動它**。
//
// `docs/re/53` 回頭追那個矛盾,結論是「0x4C 那邊沒有依據,所以動那一邊」,
// 於是把 `partyTileFor` 改成 0x1D / 0x1E。**方向錯了** —— 0x4C 是聖者(以及
// 另外五個職業)在戰場上的圖,`sub_16058` 用它判「爬得過去」正是因為
// **那一格站著隊員**。原本那個「猜」是對的。
//
// 教訓不是「不該追矛盾」,是**追矛盾時「哪一邊有證據」要用全檔掃描回答,
// 不能用「我手上這兩支函式」回答**。當時只讀了寫 `+1` 的兩支,
// 沒有去查「還有誰寫這個欄位」。`CLAUDE.md` §4.5 已有一條
// 「『唯一』『只有一處』沒有全檔掃描佐證就不要寫」——`docs/re/53`
// 寫的是「原版只有兩個值」,是同一類斷言,而它沒有做那個掃描。

// 職業字母的順序(原版 `sub_28F5C("AMBFDTPRS", 職業字母)` 的那個字串)。
//
// A 聖者 / M 法師 / B 吟遊詩人 / F 戰士 / D 德魯伊 / T 工匠 / P 聖騎士 /
// R 遊俠 / S 牧羊人。
const ClassLetters = "AMBFDTPRS"

// PartyCombatTiles 是九個職業在戰場上的圖(原版 `byte_40C34`)。
//
// ★ 只有四個不同的值:法師 0x40、吟遊詩人 0x44、戰士 0x48,
// **其餘五個(聖者 / 德魯伊 / 工匠 / 聖騎士 / 牧羊人)共用 0x4C**。
// 這與 U4 的八職業美術規模一致 —— 只有前四種原型有專屬圖。
var PartyCombatTiles = [len(ClassLetters)]byte{
	0x4C, // A 聖者
	0x40, // M 法師
	0x44, // B 吟遊詩人
	0x48, // F 戰士
	0x4C, // D 德魯伊
	0x4C, // T 工匠
	0x4C, // P 聖騎士
	0x4C, // R 遊俠
	0x4C, // S 牧羊人
}

// PartyTileDefault 是職業字母查不到時的值。
//
// 原版 `sub_28F5C` 找不到會回 −1,而 `byte_40C34[-1]` 讀到的是表**前面**
// 一個位元組 —— 那是別的資料。引擎不照抄這個越界讀(`CLAUDE.md`:
// 原版的越界讀不是規則),改用聖者那一格。
const PartyTileDefault = 0x4C

// PartyCombatTile 回傳這名角色在戰場上該畫的圖。
//
// ⚠ 只管「正常站著」那一種。睡著 / 倒下走 `PartyTileLying`、
// 隱形戒指走 `PartyTileStanding` —— 那兩條是原版的恢復路徑,不查這張表。
func PartyCombatTile(c *Character) byte {
	if c == nil {
		return PartyTileDefault
	}
	for i := 0; i < len(ClassLetters); i++ {
		if ClassLetters[i] == c.Class {
			return PartyCombatTiles[i]
		}
	}
	return PartyTileDefault
}
