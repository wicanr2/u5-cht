package game

import "github.com/wicanr2/u5-cht/internal/u5data"

// 地牢裡的 Look(原版 `sub_EEEC`,推導見 `docs/re/95`)
//
// ★ 與地面那支(`sub_D258`)是**兩支不同的程式**,差在四處:
//
//  1. **方向是相對的**(前 / 左 / 右 / 腳下),同地牢的 Search ——
//     第一人稱下「北」沒有意義。共用 `SearchRelative` 與 `dungeonRelativeSquare`。
//  2. **沒光源就是一片漆黑**,連地形都不報。
//  3. 報的是**地牢地形的十六種描述**,不是地面那套家具語。
//  4. ★★ **看到噴泉會問「要喝嗎」** —— 而喝的效果由**完整的 tile 值**決定
//     (0x50 解毒 / 0x51 補滿 / 0x52 中毒 / 其餘難喝且扣血),
//     不是高四位元。這是全支函式唯一看低位元的地方。
//
// ⚠ 原版先選隊員(`sub_E19C`)**再**問方向 —— 順序有意義:
// 取消方向時人已經選好了,而那個選擇會留在 `byte_3E08B` 裡影響下一個指令。

// dungeonLookText 是十六種地牢地形的描述,索引 = 高四位元 >> 4。
//
// ⚠⚠ **索引 8 與 12 永遠取不到** —— 原版在這張跳表**之前**就先特判了
// 0x80(力場四種)與 0xC0(崩落通道三種),所以跳表裡那兩格是死碼:
//
//	0x8 → "an energy field."     ← 走不到(力場已由前面的四路 switch 處理)
//	0xC → "SPEC WALL ERR."       ← 走不到,而且明顯是作者留著的除錯字串
//
// 照原版保留(`CLAUDE.md §3.0`),但那兩格填**哨兵值**而不是原文 ——
// 理由是可測性:原版索引 8 的字串(`aAnEnergyField`)與力場 default 的
// (`aAnEnergyField_0`)**文字完全相同**,拿字串比對根本分不出走了哪一條路。
// 原文記在 `dungeonLookDeadArms` 裡,不會遺失。
var dungeonLookText = [16]string{
	0x0: "一條通道。",
	0x1: "一道向上的梯子。",
	0x2: "一道向下的梯子。",
	0x3: "一道梯子。",
	0x4: "一只木箱。",
	0x5: "一座噴泉。",
	0x6: "一個坑洞。",
	0x7: "一只開著的箱子。",
	0x8: dungeonLookUnreachable, // ⚠ 取不到:0x80 已被力場的四路 switch 攔掉
	0x9: "沒什麼特別的。",
	0xA: "一道厚重的門。",
	0xB: "一面牆。",
	0xC: dungeonLookUnreachable, // ⚠ 取不到:0xC0 已被崩落通道那一支攔掉
	0xD: "一面牆。",
	0xE: "一道厚重的門。",
	0xF: "一道厚重的門。",
}

// dungeonLookUnreachable 是跳表裡取不到那兩格的哨兵。
//
// 任何測試看到它,就表示分派順序被改壞了(0x80 / 0xC0 漏了前面的特判)。
const dungeonLookUnreachable = "(取不到:分派順序被改壞了)"

// dungeonLookDeadArms 保留那兩格在原版裡的字串,供對碼時查。
//
//	索引 8  → "an energy field."   與力場 default 同字,所以看不出差別
//	索引 12 → "SPEC WALL ERR."     作者留著的除錯字串
var dungeonLookDeadArms = map[int]string{
	0x8: "an energy field.",
	0xC: "SPEC WALL ERR.",
}

// dungeonFieldText 是力場的四種(原版 `tile − 0x80` 的四路 switch)。
//
// ⚠ `default` 是「能量場」,而 0x80..0x83 之外的值走得到它
//(低位元還有 0x84..0x8F)。
var dungeonFieldText = [4]string{
	"一片睡眠場。",
	"一片毒氣場。",
	"一道火牆。",
	"一片電場。",
}

// DungeonFieldEnergy 是力場 switch 的 default(原版 `aAnEnergyField`)。
const DungeonFieldEnergy = "一片能量場。"

// 崩落通道那一族(高四位 0xC0)的低四位元語意(原版 `byte_3EE16 & 0x0F`)。
const (
	CavedStalactite = 1 // "a dripping stalactite."
	CavedPassage    = 2 // "a caved in passage."
)

// PirateOdds 是那句反盜版彩蛋的機率分母(原版 `random(1, 0xFF) == 0xFF`)。
//
// ★ **1/255** —— 其餘 254 次印的是「一位運氣沒那麼差的冒險者」。
// 這是 U5 有名的玩笑:骸骨旁邊寫著「一位不幸的軟體盜版者」。
const PirateOdds = 0xFF

// LookDungeon 是地牢裡的 L 指令(原版 `sub_EEEC`)。
func (s *State) LookDungeon() {
	d := s.Dungeon
	if d == nil {
		return
	}
	// ★ 先選人再問方向(原版 `sub_E19C` 在 `sub_13BA8` 之前)。
	// 引擎的隊員選擇仍是近似 —— 只有噴泉會用到它,所以在喝的那一步才取。
	s.AskDirection(func(dir Direction) {
		switch dir {
		case North:
			s.lookDungeonRelative(SearchAhead)
		case West:
			s.lookDungeonRelative(SearchLeft)
		case East:
			s.lookDungeonRelative(SearchRight)
		default:
			s.lookDungeonRelative(SearchHere)
		}
	})
}

// lookDungeonRelative 看相對方向的那一格。
func (s *State) lookDungeonRelative(rel SearchRelative) {
	d := s.Dungeon
	if d == nil {
		return
	}
	// ⚠ 與 Search 同一條:比的是**兩個光源計時器**,不是視野半徑
	// (地牢的基礎半徑是 2,永遠大於 0,拿半徑判會永遠有光)。
	if s.TorchTurns <= 0 && s.LightTurns <= 0 {
		s.Log(MsgThouDostSee + "一片漆黑。")
		return
	}
	x, y := s.dungeonRelativeSquare(rel)
	tile := s.DungeonTileAt(x, y)
	s.Log(MsgThouDostSee + dungeonLookDesc(tile, s))
	// ★★ 噴泉會問要不要喝 —— 判的是**高四位元**。
	if u5data.DungeonKind(tile) == u5data.DungeonFountain {
		s.beginDungeonDrink(tile)
	}
}

// dungeonRelativeSquare 把相對方向換成座標(原版 `sub_13BA8` → `sub_13B4C`)。
//
// ★ 「腳下」那一支**不呼叫** `sub_13B4C`,直接把隊伍座標寫進去 ——
// 所以它不會被地牢的環繞處理動到。行為與「先加 0 再環繞」相同,照抄結構。
func (s *State) dungeonRelativeSquare(rel SearchRelative) (int, int) {
	d := s.Dungeon
	x, y := d.X, d.Y
	switch rel {
	case SearchAhead:
		dx, dy := d.Facing.Delta()
		x, y = x+dx, y+dy
	case SearchLeft:
		dx, dy := d.Facing.TurnLeft().Delta()
		x, y = x+dx, y+dy
	case SearchRight:
		dx, dy := d.Facing.TurnRight().Delta()
		x, y = x+dx, y+dy
	case SearchHere:
		return x, y
	}
	return u5data.DungeonWrap(x), u5data.DungeonWrap(y)
}

// dungeonLookDesc 把地牢地形換成一句話(原版 `sub_EEEC` 的三層分派)。
//
// 分派順序照原版:力場 → 崩落通道 → 十六路跳表。
// **順序不能換** —— 換了 0x80 與 0xC0 就會落到跳表裡那兩個死格。
func dungeonLookDesc(tile byte, s *State) string {
	switch u5data.DungeonKind(tile) {
	case u5data.DungeonMagic: // 0x80 力場
		if i := int(tile - u5data.DungeonMagic); i >= 0 && i < len(dungeonFieldText) {
			return dungeonFieldText[i]
		}
		return DungeonFieldEnergy
	case u5data.DungeonUnknownC: // 0xC0 崩落通道
		switch int(tile & 0x0F) {
		case CavedStalactite:
			return "一根滴水的石筍。"
		case CavedPassage:
			return "一條崩落的通道。"
		}
		// ★ 1/255 的反盜版彩蛋。
		if s.Roll(1, PirateOdds) == PirateOdds {
			return "一位不幸的軟體盜版者。"
		}
		return "一位運氣沒那麼差的冒險者。"
	}
	return dungeonLookText[u5data.DungeonKind(tile)>>4]
}

// 地牢噴泉的三種有效泉水(原版比的是**完整 tile 值**)。
const (
	FountainCure   = 0x50 // "Cured!"    狀態設 'G'
	FountainHeal   = 0x51 // "Healed!"   HP 補到 MaxHP
	FountainPoison = 0x52 // "Poisoned!" 狀態設 'P'
)

// FountainBadTasteMax 是難喝的泉水扣血的上限(原版 `random(0, 7)`)。
//
// ⚠ 下限是 **0** ——「難喝」有 1/8 的機率一點血都不扣。
const FountainBadTasteMax = 7

// beginDungeonDrink 問「要喝嗎」(原版 `"Will you drink?\n"`)。
//
// ⚠ 原版只收 Y / N:**其他鍵一律重問**(`jnz` 繞回去),不是當成 N ——
// 引擎的 `Ask` 把 ESC 當 N,那是既有的近似(`AnswerYesNo` 的註解已記)。
func (s *State) beginDungeonDrink(tile byte) {
	s.Ask("要喝嗎?", func(yes bool) { s.drinkFromDungeonFountain(tile, yes) })
}

// drinkFromFountain 喝下去(原版 `loc_F1E2` 之後那一段)。
func (s *State) drinkFromDungeonFountain(tile byte, yes bool) {
	if !yes {
		return
	}
	s.Log("咕嚕!")
	// ★ `pickMember` 就是 `sub_E19C`:必要時開「擇一:」選單問玩家,
	// 所以效果要寫在回呼裡 —— 寫在後面會在「要問」那條路上提前發生。
	s.pickMember("", func(who int) {
		if who < 0 {
			return
		}
		switch tile {
		case FountainCure:
			s.Log("解毒了!")
			s.setMemberStatus(who, u5data.StatusGood)
		case FountainHeal:
			s.Log("痊癒了!")
			s.healMemberFull(who)
		case FountainPoison:
			s.Log("中毒了!")
			s.setMemberStatus(who, u5data.StatusPoisoned)
		default:
			// ★ 0x53..0x5F:難喝,而且扣 0..7 血。
			s.Log("味道很糟。")
			s.damageMember(who, s.Roll(0, FountainBadTasteMax))
		}
	})
}

// setMemberStatus 寫一個隊員的狀態(界內才寫)。
//
// ⚠ 與 `poisonMember`(寶箱的毒陷阱)**不同**:那一支會擋掉死人,
// 而噴泉這裡原版是 `mov byte_3DDBF[eax], 50h` —— **沒有任何前置判斷**。
// 兩支不能合併。
func (s *State) setMemberStatus(i int, st byte) {
	if i < 0 || i >= s.PartySize || i >= len(s.Roster) {
		return
	}
	s.Roster[i].Status = st
}

// healMemberFull 把血補到上限(原版 `word_3DDC4 = word_3DDC6`)。
//
// ⚠ **不看死活** —— 原版那兩行沒有任何前置判斷,死人的 HP 也會被填滿
//(狀態仍是 'D',所以還是死的)。照抄。
func (s *State) healMemberFull(i int) {
	if i < 0 || i >= s.PartySize || i >= len(s.Roster) {
		return
	}
	s.Roster[i].HP = s.Roster[i].MaxHP
}
