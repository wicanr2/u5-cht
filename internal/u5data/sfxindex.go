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
//	A 級(有呼叫點):MOON2・T_OCHI1・MAHOU1・DAME1・FUNSUI2・CLOCK2・Fire・
//	               MIRROR2・SUITEKI3・ALARM3 —— 見上表與 `docs/re/90`
//	B 級(只有檔名):WALK・WALKSLOW・HORSE・BLOCK・ATTACK1/2・NIGE・DAME2・
//	               YOGAN・WHAT・DOKU・TAKI2・DEATH1
//	               ⇒ 檔名一望即知(WALK = 腳步),但**沒有追到呼叫點**。
//	               引擎照名字接,而這件事在 `docs/re/90` §4 誠實列出。
const (
	SFXWalk      = 0  // WALK.SND     腳步
	SFXWalkSlow  = 1  // WALKSLOW.SND 慢速腳步(粗糙地形)
	SFXHorse     = 2  // HORSE.SND    馬蹄
	SFXBlock     = 3  // BLOCK.SND    去路受阻
	SFXAttack1   = 4  // ATTACK1.SND
	SFXAttack2   = 5  // ATTACK2.SND
	SFXFlee      = 6  // NIGE.SND     「逃げ」= 逃走
	SFXDamage1   = 7  // DAME1.SND    ★ A 級(sub_2AC08)
	SFXDamage2   = 8  // DAME2.SND
	SFXMagic     = 9  // MAHOU1.SND   ★ A 級(sub_10C34)
	SFXMoongate  = 10 // MOON2.SND    ★ A 級(sub_135FC ×3)
	SFXLava      = 11 // YOGAN.SND    「溶岩」= 熔岩
	SFXWhat      = 12 // WHAT.SND     無效指令
	SFXPoison    = 13 // DOKU.SND     「毒」
	SFXFire      = 14 // Fire.SND     ★ A 級(sub_2C188)
	SFXClock     = 15 // CLOCK2.SND   ★ A 級(sub_2C4F4)
	SFXMirror    = 16 // MIRROR2.SND  ★ A 級(sub_CAC)
	SFXFountain  = 17 // FUNSUI2.SND  ★ A 級(sub_2C598)「噴水」
	SFXDrip      = 18 // SUITEKI3.SND ★ A 級(sub_35EC)「水滴」
	SFXWaterfall = 19 // TAKI2.SND    「滝」= 瀑布
	SFXFall      = 20 // T_OCHI1.SND  ★ A 級(sub_10A1C)「落ち」
	SFXDeath     = 21 // DEATH1.SND
	SFXAlarm     = 22 // ALARM3.SND   ★ A 級(sub_2BDE0)

	// SFXCount 是 `U5_SE.TBL` 的列數。
	//
	// ⚠ **23,不是 25** —— 目錄裡有 25 個 `.SND`,但表只列 23 筆
	// (`BEEP.SND` 與 `BARIBARI.SND` 不在表裡,`docs/re/63` 已用
	// `sub_2C2AC` 的 `cmp ebx, 17h` 佐證)。
	SFXCount = 23
)

// SFXNone 代表「這一次不出聲」。
const SFXNone = -1
