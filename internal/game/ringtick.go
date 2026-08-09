package game

import "github.com/wicanr2/u5-cht/internal/u5data"

// 戒指的**持續效果**、力場消散、回合計數器
//
// 推導見 `docs/re/99`。這三件事此前一件都沒做,而漏掉的原因分成兩種:
//
//	sub_2BCC8(再生戒指)  Hex-Rays 反編譯成 `return 0;` —— 組語是 55 行真邏輯
//	                      (`CLAUDE.md §4.4` 的第一種失效形態)
//	sub_1F4E4(力場消散)  `sub_16370` 的第二行,而那支只被讀了第一行
//	byte_3E09B(回合數)   語意未定 ⇒ 連帶擋住乞丐的業報獎勵
//
// ★★ 而 `call sub_2BCC8` 是 **`sub_2A50C`(維生開銷)的最後一行** ——
// 同 `docs/re/85` 的「怪物移動在 `sub_2D38` 的最後一行」。**同一個形狀踩第二次**:
// 讀一支函式時讀到「看起來收尾了」就停,而原版把新東西掛在 `retn` 前面一行。
// ⇒ 讀函式一律讀到 `endp`。

// 戒指持續效果的常數。
const (
	// RegenRingRollMax 是再生戒指的擲骰上限(原版 `random(0, 7) == 7` ⇒ 1/8)。
	RegenRingRollMax = 7
	// RegenRingHeal 是中了就回幾點(原版 `sub_2BBDC(&hp, 1, maxhp)`)。
	RegenRingHeal = 1
)

// regenerateParty 是再生戒指的回血(原版 `sub_2BCC8`)。
//
//	for (i = 0; i < 隊伍人數; i++) {
//	    if (狀態 == 'D')            continue    ; ★ 只跳過「死」
//	    if (戒指 != 0x2C)           continue
//	    if (random(0,7) != 7)       continue    ; 1/8
//	    HP = min(HP + 1, MaxHP)
//	}
//
// ⚠ 三個容易寫錯的地方:
//
//  1. **只跳過 'D'。** 睡著('S')、被惑('C')、中毒('P')的人**照樣回血** ——
//     而中毒的人同一回合還會被維生開銷扣 1 點,兩件事各自獨立發生。
//  2. **它忽略自己的參數。** `sub_2ECE8` 呼叫時推了一個名冊索引進去,
//     而 `sub_2BCC8` 的堆疊框裡根本沒有 `arg_0` —— 它一律掃整隊。
//     ⇒ 兩個人都戴再生戒指時,每個人每回合會被擲**兩次**骰。照抄。
//  3. **上限是那個人的 MaxHP**,不是隊伍共用的值。
func (s *State) regenerateParty() {
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		ch := &s.Roster[i]
		if ch.Status == u5data.StatusDead {
			continue
		}
		if ch.Raw[u5data.CharRing] != u5data.ItemRingRegeneration {
			continue
		}
		if s.Roll(0, RegenRingRollMax) != RegenRingRollMax {
			continue
		}
		ch.HP = uint16(addCap(int(ch.HP), RegenRingHeal, int(ch.MaxHP)))
	}
}

// ringUpkeep 是戰鬥中每個單位回合結束時的戒指效果(原版 `sub_2ECE8(byte_3E0AE)`)。
//
//	單位 = 此刻行動的那一個(`byte_3E0AE`)
//	if (單位不是隊員)                    return
//	if (單位睡著或死了)                  return  ; 原版 `test al, 28h`
//	if (戒指 == 0x2A) { 掛上隱形位元; return }   ; 隱形戒指
//	if (戒指 == 0x2C) 全隊擲再生骰
//
// ⚠⚠ **參數差點寫錯成 0。** Hex-Rays 把它反編譯成 `sub_2ECE8(0)` ——
// 常數傳播把一個全域摺成了字面 0(`CLAUDE.md §4.4` 第一種失效形態)。
// 組語是 `movzx eax, byte_3E0AE; push eax`,而那個全域是**當前行動的單位**
// (`docs/re/67`)。寫成 0 的話會變成「只有第一槽的戒指有用」,
// 而那是一條完全捏造出來的規則。
//
// ★ 閘門是**行動者自己的**戒指,而 `regenerateParty` 掃的是**全隊** ——
// 兩者範圍不同是原版的樣子(被呼叫端忽略參數),不要對齊。
//
// ⚠ 隱形那條**先 return**:同時戴不到兩只戒指,所以實務上互斥,
// 但順序照抄比較安全。
func (s *State) ringUpkeep() {
	c := s.Combat
	if c == nil || c.Turn < 0 || c.Turn >= len(c.Units) {
		return
	}
	u := &c.Units[c.Turn]
	if u.Flags&UnitParty == 0 {
		return
	}
	// 原版 `test al, 28h` = 睡著(0x08)或死了(0x20)就跳過。
	if u.Flags&(UnitAsleep|UnitDead) != 0 {
		return
	}
	if u.Roster < 0 || u.Roster >= len(s.Roster) {
		return
	}
	switch s.Roster[u.Roster].Raw[u5data.CharRing] {
	case u5data.ItemRingInvisibility:
		// 每回合**重新**掛上隱形 —— 所以戴著它的人被 Sanct Lor 之類的
		// 東西清掉隱形之後,下一回合又會隱形。
		u.Flags |= UnitHidden
	case u5data.ItemRingRegeneration:
		s.regenerateParty()
	}
}

// 力場消散的常數(原版 `sub_1F4E4`)。
const (
	// FieldObjectKind 是四種力場的種類碼基底(0xE8..0xEB,原版 `kind & 0xFC == 0xE8`)。
	//
	// ★ 四種的名字來自 look 表的物件段(索引 256 + 種類碼):
	// 0xE8 毒力場 / 0xE9 睡眠力場 / 0xEA(原版寫成 "a field of field")/ 0xEB 力場。
	FieldObjectKind = 0xE8
	// FieldObjectKindMask 是原版比對用的遮罩。
	FieldObjectKindMask = 0xFC
	// FieldExpiryRollMax 是消散擲骰的上限(原版 `random(0, 255)`)。
	FieldExpiryRollMax = 255
	// FieldExpiryBelow 是「擲出小於它就消散」(原版 `< 16` ⇒ 16/256 = 1/16)。
	FieldExpiryBelow = 16
)

// expireFields 讓場上的力場自然消散(原版 `sub_1F4E4`)。
//
// 每個玩家單位回合結束時,**每一個**力場各擲一次 `random(0,255)`,
// 小於 16 就從物件表上清掉。⚠ **沒有訊息** —— 力場就這樣不見了。
// (「力場消散了!」那句是舉權杖時才印的,別搬過來。)
//
// ⚠ 掃的是**整張物件表 32 槽**,不是戰鬥單位表 —— 力場是物件不是單位。
func (s *State) expireFields() {
	objs := s.currentObjects()
	if objs == nil {
		return
	}
	for i := range objs.Objects {
		o := &objs.Objects[i]
		if o.Kind&FieldObjectKindMask != FieldObjectKind {
			continue
		}
		if s.Roll(0, FieldExpiryRollMax) < FieldExpiryBelow {
			objs.Remove(i)
		}
	}
}

// TurnCounterMax 是回合計數器的上限(原版 `sub_2BBB8(&byte_3E09B, 1, 0FFh)`)。
const TurnCounterMax = 255

// TurnsSinceAlms 是「距離上次施捨過了幾回合」(原版 `byte_3E09B`)。
//
// ★★ 這個變數的語意**此前被記成「進餐計數器」而那是錯的**:它的 `+1` 在
// `sub_2A50C` 的尾段,而那一段連「這個小時已經結算過」的路徑也會走到 ——
// 所以它**每回合 +1**,與吃不吃飯無關。
//
// 定出語意之後,`sub_1B964` 那個「`byte_3E09B >= 100` 才給業報」的閘門
// 就講得通了:**同一個乞丐(或一連串乞丐)短時間內討不到第二次業報**。
func (s *State) TurnsSinceAlms() int { return s.turnsSinceAlms }

// countTurn 是維生開銷尾段的回合計數(原版 `sub_2BBB8(&byte_3E09B, 1, 0FFh)`)。
func (s *State) countTurn() {
	s.turnsSinceAlms = addCap(s.turnsSinceAlms, 1, TurnCounterMax)
}
