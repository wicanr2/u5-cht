package u5data

// 音效的索引 —— 就是 `U5_SE.TBL` 的行號(原版 `sub_2C46C(索引, 音高)`)
//
// 推導見 `docs/re/90`。★ 索引 = 表的行號,這一點有**日文檔名交叉確認**:
//
//	sub_10A1C(墜落動畫)      → 0x14 = T_OCHI1   ★「落ち」= 掉落
//	sub_135FC(月門過場)      → 0x0A = MOON2     ★ 用了三次
//	sub_10C34                → 0x09 = MAHOU1    ★「魔法」
//	sub_2AC08                → 0x07 = DAME1     ★「ダメージ」= 傷害
//	sub_2C598                → 0x11 = FUNSUI2   ★「噴水」= 噴泉
//	sub_2C4F4                → 0x0F = CLOCK2
//
// 六個呼叫點的索引與日文檔名逐一相符 ⇒ 索引就是行號,不是別的編號。
//
// ⚠⚠ **觸發點的證據強度分兩級**,不要混:
//
//	A 級(有呼叫點):MOON2・T_OCHI1・MAHOU1・DAME1・DAME2・FUNSUI2・CLOCK2・
//	               Fire・MIRROR2・SUITEKI3・ALARM3・**WALK**・**DOKU**・
//	               **TAKI2** —— 見上表、`docs/re/90` 與 `docs/re/92`
//	B 級(只有檔名):WALKSLOW・HORSE・BLOCK・ATTACK1/2・NIGE・YOGAN・WHAT・DEATH1
//	               ⇒ 檔名一望即知(HORSE = 馬蹄),但**沒有追到呼叫點**。
//	               引擎照名字接,而這件事在 `docs/re/90` §4 誠實列出。
//
// ★★ **升級的四筆(2026-08-08,`docs/re/92`)**:`sub_2C598` 是「DOS 版白噪
// 參數 → FM Towns PCM」的**轉接層**,26 個呼叫點裡 19 個會落到 `sub_2C46C`。
// 它讓 WALK(`sub_2C118` 的兩次白噪呼叫)、DOKU(`sub_2A464` 扣血)、
// TAKI2(瀑布環境音)、DAME2 從「只有檔名」變成「有呼叫點」。
const (
	SFXWalk      = 0  // WALK.SND     ★ A 級(sub_2C118 的兩次白噪)
	SFXWalkSlow  = 1  // WALKSLOW.SND 慢速腳步(粗糙地形)
	SFXHorse     = 2  // HORSE.SND    馬蹄
	SFXBlock     = 3  // BLOCK.SND    去路受阻
	SFXAttack1   = 4  // ATTACK1.SND
	SFXAttack2   = 5  // ATTACK2.SND
	SFXFlee      = 6  // NIGE.SND     「逃げ」= 逃走
	SFXDamage1   = 7  // DAME1.SND    ★ A 級(sub_2AC08、陷阱、地形效果…共 5 處)
	SFXDamage2   = 8  // DAME2.SND    ★ A 級(sub_2B1C8 / sub_2B21C / 許願井…)
	SFXMagic     = 9  // MAHOU1.SND   ★ A 級(sub_10C34;另有三處**音高隨參數變**)
	SFXMoongate  = 10 // MOON2.SND    ★ A 級(sub_135FC ×3)
	SFXLava      = 11 // YOGAN.SND    「溶岩」= 熔岩
	SFXWhat      = 12 // WHAT.SND     無效指令
	SFXPoison    = 13 // DOKU.SND     ★ A 級(sub_2A464 扣一人的血)「毒」
	SFXFire      = 14 // Fire.SND     ★ A 級(sub_2C188)
	SFXClock     = 15 // CLOCK2.SND   ★ A 級(sub_2C4F4,落地鐘的滴答)
	SFXMirror    = 16 // MIRROR2.SND  ★ A 級(sub_CAC)
	SFXFountain  = 17 // FUNSUI2.SND  ★ A 級(sub_2BDE0 環境音)「噴水」
	SFXDrip      = 18 // SUITEKI3.SND ★ A 級(sub_35EC)「水滴」
	SFXWaterfall = 19 // TAKI2.SND    ★ A 級(sub_2BDE0 環境音)「滝」= 瀑布
	SFXFall      = 20 // T_OCHI1.SND  ★ A 級(sub_10A1C)「落ち」
	SFXDeath     = 21 // DEATH1.SND
	SFXAlarm     = 22 // ALARM3.SND   ★ A 級(sub_2BDE0,落地鐘第 4 相)

	// SFXCount 是 `U5_SE.TBL` 的列數。
	//
	// ⚠ **23,不是 25** —— 目錄裡有 25 個 `.SND`,但表只列 23 筆
	// (`BEEP.SND` 與 `BARIBARI.SND` 不在表裡,`docs/re/63` 已用
	// `sub_2C2AC` 的 `cmp ebx, 17h` 佐證)。
	SFXCount = 23
)

// SFXNone 代表「這一次不出聲」。
const SFXNone = -1

// 寶箱 / 陷阱的種類與權重(原版 `sub_2AB38` 的 `byte_5FFEC`)
//
// 推導見 `docs/re/91`。
const (
	TrapAcid   = 0 // "Acid"   單人 1..30 傷害
	TrapPoison = 1 // "Poison" 單人中毒
	TrapBomb   = 2 // "Bomb"   全隊各 random(1,8)
	TrapGas    = 3 // "Gas"    全隊中毒
)

// TrapKindRoll 是 `byte_5FFEC[8]` —— 用 `random(0,7)` 索引。
//
//	{0, 0, 0, 1, 1, 2, 2, 3}
//
// ⇒ 酸 3/8(37.5%)、毒 2/8(25%)、炸彈 2/8(25%)、毒氣 1/8(12.5%)。
//
// ★ 用「重複的項目」表示權重,與生怪的四張表同一個手法(`docs/re/82`)。
var TrapKindRoll = [8]byte{0, 0, 0, 1, 1, 2, 2, 3}

// TrapKindRollMax 是查表用的擲骰上限(原版 `sub_28E14(0, 7)`)。
const TrapKindRollMax = 7

// TrapBombDamageMax 是炸彈對每個隊員的傷害上限(原版 `sub_28E14(1, 8)`)。
const TrapBombDamageMax = 8

// TrapCombatKindMax 是**戰鬥中**的陷阱種類上限(原版 `random(0, 1)`)。
//
// ★ 地點碼 > 0x7F(戰鬥)時不查那張八筆表,只擲 `random(0,1)`
// ⇒ **戰鬥中只會是酸或毒**,不會有炸彈與毒氣。
const TrapCombatKindMax = 1
