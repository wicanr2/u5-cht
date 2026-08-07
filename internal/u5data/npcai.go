package u5data

// NPC 的行為型別(原版 `byte_3E570[npc*16 + slot]`,跳表在 `sub_95BC`)
//
// 排程表每個 slot 有四個欄位:**行為型別**、X、Y、樓層。引擎原本只用了後三個,
// 所以 NPC 走到定點之後就完全靜止 —— 原版不是這樣:走到定點之後**才**輪到
// 行為型別接手,城裡才會有人晃來晃去、有人逃、有人追上來搭話。
//
// 八種型別由 `sub_95BC` 的八格跳表分派:
//
//	0  固定    什麼都不做
//	1  遊走    在自己的崗位 3 格內隨機走(sub_94E0(…, 3))
//	2  原地晃  同上但範圍 0 —— 只在原格轉向(sub_94E0(…, 0))
//	3  怕生    玩家走到 4 格內就往外躲(sub_8F60,離愈遠愈好)
//	4  搭話    玩家走到 4 格內就靠過來,不然在崗位 3 格內遊走
//	5  跟隨    永遠朝玩家走(sub_8F60)
//	6  警戒    同 3 的觸發距離,但 `sub_8F60` 依型別改成**靠過來**
//	7  醉步    同 5,但每步有機率亂走
//
// ⚠ 型別 3 與 6 走同一個分支、5 與 7 也是 —— 差別在 `sub_8F60` 內部再讀一次
// 自己的型別:**3 是遠離、其餘是靠近**。抄成「3/6 一樣、5/7 一樣」就會讓
// 怕生的村民追著玩家跑。
const (
	NPCAIFixed    = 0
	NPCAIWander   = 1
	NPCAIStay     = 2
	NPCAIShy      = 3
	NPCAIGreet    = 4
	NPCAIFollow   = 5
	NPCAIGuardish = 6
	NPCAIDrunk    = 7
)

// NPCAIWanderRange 是型別 1 / 4 能離開崗位幾格(`sub_95BC` 推 3)。
const NPCAIWanderRange = 3

// NPCAINoticeRange 是「玩家進到幾格內就有反應」(`cmp eax, 4`)。
//
// 用的是**曼哈頓距離**(`sub_8F3C` = |dx| + |dy|),不是直線也不是棋盤距離。
const NPCAINoticeRange = 4

// NPCAIBlocked 是 `sub_8F60` 給走不過去的方向的距離值(`0x63`)。
//
// 挑成 99 是因為它比任何真實距離都大,所以「找最近的一格」自然會跳過它。
const NPCAIBlocked = 0x63

// NPCAIHostile / NPCAIHostileBig 是「叫衛兵」之後 NPC 被改成的型別
//(`sub_B44`:生物編號 < 0x2F 的設 6,其餘設 7)。
const (
	NPCAIHostile    = 6
	NPCAIHostileBig = 7
	// NPCAIHostileSplit 是那個分界(`cmp byte_3EDB0[edi], 2Fh`)。
	NPCAIHostileSplit = 0x2F
)

// NPCAIFleeing 是被嚇跑之後的型別(`sub_B98` 設 3)。
const NPCAIFleeing = 3

// 「叫衛兵」會變敵對的三種生物(`sub_C10` 的 `cmp esi, 0FCh / 0D8h / 70h`)。
//
// ⚠ 其餘的 NPC **有一半機率直接逃跑**(`random(0,255) < 0x80` → `sub_B98`)——
// 少了這一段,叫完衛兵整條街的人還若無其事地站著。
const (
	CreatureGuard = 0x70
	// CreatureGuardCaptain 是另一種會應召的守衛(0xD8 → 生物編號 38)。
	// 名字取自它與衛兵、暗影君主並列在同一個條件裡,不是從名表猜的。
	CreatureGuardCaptain = 0xD8
)

// CallGuardsFleeChance 是非戰鬥人員在「叫衛兵」時逃跑的門檻
//(`random(0, 0xFF) < 0x80`,也就是一半)。
const CallGuardsFleeChance = 0x80

// 逮捕(`sub_1884`)

// ArrestJailLocation / ArrestJailX / ArrestJailY 是乖乖就範之後醒來的地方。
//
// 原版 `byte_3E0A3 = 4`(紫杉城 YEW)、`byte_3E0A6 = 0x19`、`byte_3E0A7 = 4`。
const (
	ArrestJailLocation = 4
	ArrestJailX        = 0x19
	ArrestJailY        = 4
)

// ArrestWakeHour 是被關進去之後醒來的時刻(`while (byte_3E08F != 8) sub_29304(20)`)。
//
// ⚠ 原版是**一次跳 20 分鐘**直到小時剛好等於 8 —— 所以醒來的分鐘數不一定是 0,
// 取決於被抓的時間。照抄比直接設成 08:00 更接近原版。
const (
	ArrestWakeHour = 8
	ArrestWakeStep = 20
)
