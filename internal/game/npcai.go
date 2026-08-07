package game

import (
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// NPC 走到定點之後的行為(原版 `sub_95BC` 的八格跳表)
//
// 排程表每個 slot 有四個欄位:行為型別、X、Y、樓層。引擎原本只用後三個,
// 所以 NPC 一走到崗位就完全靜止 —— 原版是**走到定點之後才輪到行為型別接手**
//(`sub_9690` 在 NPC 與玩家同層時呼叫 `sub_95BC`)。
// 少了這一段,城裡每個人都像蠟像;而且「叫衛兵」之後沒有任何人會過來。
//
// 型別的語意見 `u5data` 的 NPCAI* 常數。

// npcAIStep 是一個閒置中的 NPC 這一回合的行為。
func (s *State) npcAIStep(i int) {
	rt := &s.rtNPCs[i]
	if rt.Floor != s.Floor {
		// 原版只在同層時才跑這一段(`cmp [ebx+6], byte_3E0A5`)。
		return
	}
	sched := &s.npcs[i].Schedule
	ai := int(sched.AI[rt.Slot])
	switch ai {
	case u5data.NPCAIWander:
		s.npcWander(i, u5data.NPCAIWanderRange)
	case u5data.NPCAIStay:
		// ⚠ 範圍 0 不是「不動」—— 它照樣擲骰、照樣算目標格,
		// 只是「離崗位超過 0 格」永遠成立,所以實際上留在原地。
		// 寫成 `return` 會少掉那一次擲骰,之後的亂數序列就跟原版岔開了。
		s.npcWander(i, 0)
	case u5data.NPCAIShy, u5data.NPCAIGuardish:
		if manhattan(s.X, s.Y, rt.X, rt.Y) < u5data.NPCAINoticeRange {
			s.npcApproach(i, ai)
		}
	case u5data.NPCAIGreet:
		// 先看排程座標離玩家多遠 —— ⚠ 用的是**崗位**而不是現在站的位置
		// (原版 `sub_95BC` case 4 取的是 `[eax+esi+3]` / `[eax+esi+6]`)。
		home := manhattan(s.X, s.Y,
			int(sched.X[rt.Slot]), int(sched.Y[rt.Slot]))
		if home < u5data.NPCAINoticeRange {
			s.npcApproach(i, ai)
		} else {
			s.npcWander(i, u5data.NPCAIWanderRange)
		}
	case u5data.NPCAIFollow, u5data.NPCAIDrunk:
		s.npcApproach(i, ai)
	}
	// NPCAIFixed(0)與跳表的 default 一樣:什麼都不做。
}

// npcWander 是原版 `sub_94E0` 的隨機遊走。
//
// ⚠ **只有一半的回合會動**:原版擲一個 0..255 的亂數然後測 **bit 3**
//(`test al, 8`)—— 不是「小於 128」。兩者機率一樣但取的位元不同,
// 而同一個亂數序列下走法會完全不同。
//
// maxDist 是能離開崗位多遠;0 代表哪裡都不能去(型別 2)。
func (s *State) npcWander(i, maxDist int) {
	rt := &s.rtNPCs[i]
	if s.Roll(0, 0xFF)&8 == 0 {
		return
	}
	// ⚠ 方向是 `(random(0,0x3F) & 3) + 1` —— 取低兩位元,不是 `random(1,4)`。
	dir := (s.Roll(0, 0x3F) & 3) + 1
	nx, ny := walkStep(rt.X, rt.Y, dir)
	sched := &s.npcs[i].Schedule
	if maxDist != 0 {
		if manhattan(int(sched.X[rt.Slot]), int(sched.Y[rt.Slot]), nx, ny) > maxDist {
			return
		}
	} else if nx != rt.X || ny != rt.Y {
		// 範圍 0:任何一步都超出崗位。
		return
	}
	if !s.canStandPlain(i, nx, ny) {
		return
	}
	rt.X, rt.Y = nx, ny
}

// canStandPlain 是「這個 NPC 站得住這一格嗎」,**沒有目標格豁免**。
//
// ⚠ `npcCanStand` 走的是 `sub_9358`,那一支對**目標格**永遠回可站
//(NPC 要走得到自己的崗位,哪怕崗位畫在櫃檯後面)。行為型別這條路走的是
// 另一支 `sub_9428`,沒有那個豁免 —— 把候選格當成自己的目標傳進去的話,
// 檢查會恆真,NPC 就會走進牆裡、走出地圖外、疊在別人身上。
// (實際踩到了:`TestNPCsStayOnWalkableTiles` 當場抓出來,
// 而 `TestNPCsDoNotOverlap` 直接 panic 在負座標上。)
func (s *State) canStandPlain(self, x, y int) bool {
	return s.npcCanStand(self, x, y, -1, -1)
}

// npcApproach 是原版 `sub_8F60`:朝玩家走一步,或者(型別 3)往外躲一步。
//
// ⚠ **貼到玩家旁邊時不再移動,而是「接觸」** —— 有對話號碼的搭話、
// 沒有的動手。這就是「叫衛兵」之後衛兵走過來說「汝被逮捕了」的那條路。
func (s *State) npcApproach(i, ai int) {
	rt := &s.rtNPCs[i]
	if manhattan(s.X, s.Y, rt.X, rt.Y) == 1 && ai > u5data.NPCAIShy {
		s.npcContact(i, ai)
		return
	}
	flee := ai == u5data.NPCAIShy
	d0 := manhattan(s.X, s.Y, rt.X, rt.Y)

	// 四個方向各算一次「走過去之後離玩家多遠」;走不過去的記 0x63。
	var dist [5]int
	for dir := 1; dir <= 4; dir++ {
		nx, ny := walkStep(rt.X, rt.Y, dir)
		if !s.canStandPlain(i, nx, ny) {
			dist[dir] = u5data.NPCAIBlocked
			continue
		}
		dist[dir] = manhattan(s.X, s.Y, nx, ny)
	}

	best := -1
	for dir := 1; dir <= 4; dir++ {
		if dist[dir] == u5data.NPCAIBlocked {
			continue
		}
		if flee {
			// 逃:找更遠的。⚠ 有多個更遠的方向時原版**擲硬幣決定要不要換**
			// (`sub_28E14(0,1)`),不是取第一個也不是取最遠的。
			if dist[dir] > d0 && (best < 0 || s.Roll(0, 1) == 0) {
				best = dir
			}
		} else if dist[dir] < d0 {
			best = dir
			break
		}
	}
	if best < 0 && !flee {
		// 沒有更近的一格 → 退而求其次,接受**距離不變**的一格
		//(原版第二輪 edi 5..7 重掃 dist[1..3] 找 `== d0`)。
		for dir := 1; dir <= 3; dir++ {
			if dist[dir] != u5data.NPCAIBlocked && dist[dir] == d0 {
				best = dir
				break
			}
		}
	}

	// ⚠ 型別 5 / 7 有 25% 機率不照最佳解走(`random(0,0x3F) < 0x10`)——
	// 那是「醉步」的來源。少了它,跟隨型 NPC 會走得像導彈。
	if (ai == u5data.NPCAIFollow || ai == u5data.NPCAIDrunk) && s.Roll(0, 0x3F) < 0x10 {
		pick := best
		for dir := 1; dir <= 4; dir++ {
			if dir == best || dist[dir] == u5data.NPCAIBlocked {
				continue
			}
			if pick == best || s.Roll(0, 0x3F) < 0x10 {
				pick = dir
			}
		}
		best = pick
	}
	if best <= 0 {
		return
	}
	rt.X, rt.Y = walkStep(rt.X, rt.Y, best)
}

// npcContact 是 NPC 貼到玩家身上的那一刻(原版 `byte_3EDD0` 的 't' / 'a',
// 由 `sub_195C` 分派)。
//
// ⚠ 判斷「搭話還是動手」看的是**兩件事**:型別是不是 4 / 5,以及
// **這個人有沒有對話號碼**(`word_3E77A != 0`)。有話可說的走搭話,
// 其餘一律動手 —— 所以敵對化的衛兵(型別 6/7)永遠走動手那條。
func (s *State) npcContact(i, ai int) {
	n := &s.npcs[i]
	if (ai == u5data.NPCAIGreet || ai == u5data.NPCAIFollow) && n.Dialogue != 0 {
		s.talkToNPC(i)
		return
	}
	// 動手:衛兵是逮捕,其餘是開打。
	if n.Creature == u5data.CreatureGuard || n.Creature == u5data.CreatureGuardCaptain {
		s.Arrest()
		return
	}
	// 場景裡的 NPC 打起來要走另一條戰鬥路徑(原版 `sub_C74`),還沒逆 ——
	// 誠實說明,不要假裝什麼都沒發生(CLAUDE.md §3.0)。
	s.Log(MsgAttacked)
	s.Log(MsgSceneCombatNotImplemented)
}

// CallGuards 是對話裡的「叫衛兵」(原版 opcode 0x8B → `sub_C10`)。
//
// 兩件事同時發生,少一件都不對:
//
//	衛兵、衛兵長、暗影君主 → 變敵對(型別 6 或 7),朝玩家走過來
//	其餘每個人            → **一半機率當場逃跑**(型別 3)
//
// ⚠ 少了後者,叫完衛兵整條街的人還若無其事地站在原地。
func (s *State) CallGuards() {
	if s.npcs == nil {
		return
	}
	if len(s.rtNPCs) != u5data.NPCsPerLocation {
		s.initRuntimeNPCs()
	}
	for i := range s.npcs {
		if i == u5data.PartySlot {
			continue
		}
		n := &s.npcs[i]
		if !n.Present() || s.rtNPCs[i].Mode == ModeAbsent {
			continue
		}
		switch n.Creature {
		case u5data.CreatureGuard, u5data.CreatureGuardCaptain, u5data.TileShadowlord:
			s.makeHostile(i)
		default:
			if s.Roll(0, 0xFF) < u5data.CallGuardsFleeChance {
				s.makeFlee(i)
			}
		}
	}
}

// makeHostile 把一個 NPC 變成敵對(原版 `sub_B44`)。
//
// ⚠ 型別依生物編號分兩種:< 0x2F 設 6、其餘設 7。兩者在 `sub_8F60` 裡
// 差在「7 會醉步」—— 大型生物走得沒那麼直。
//
// ⚠ 而且**四個排程時刻要一起清成 0** —— 不然下一個整點換班時,
// 排程會把牠打回原本的崗位,敵意就這樣消失了。
func (s *State) makeHostile(i int) {
	ai := byte(u5data.NPCAIHostile)
	if s.npcs[i].Creature >= u5data.NPCAIHostileSplit {
		ai = u5data.NPCAIHostileBig
	}
	s.setNPCAI(i, ai)
}

// makeFlee 把一個 NPC 變成逃跑(原版 `sub_B98` 設型別 3)。
func (s *State) makeFlee(i int) { s.setNPCAI(i, u5data.NPCAIFleeing) }

// setNPCAI 改一個 NPC 的行為型別,並清掉排程時刻。
func (s *State) setNPCAI(i int, ai byte) {
	sched := &s.npcs[i].Schedule
	for k := range sched.AI {
		sched.AI[k] = ai
	}
	for k := range sched.Times {
		sched.Times[k] = 0
	}
	if i < len(s.rtNPCs) {
		s.rtNPCs[i].Mode = ModeIdle
	}
}

// 逮捕(原版 `sub_1884`)
//
//	在黑棘宮殿(地點 18)  → 直接進審問,不問任何話
//	其餘地方              → 「汝束手就擒否?」
//	                          是 → 被打昏,在**紫杉城的牢房**醒來(隔天早上八點,
//	                               鑰匙歸零)
//	                          否 → 「那就自衛吧,惡棍!」→ 全城衛兵撲上來

// Arrest 是衛兵抓到玩家的那一刻。回傳有沒有進到打鬥。
func (s *State) Arrest() bool {
	if s.Location == u5data.BlackthornLocation {
		s.BeginInterrogation()
		return false
	}
	s.Log(MsgUnderArrest)
	s.Prompt = PromptArrest
	return false
}

// AnswerArrest 回答「汝束手就擒否?」。
func (s *State) AnswerArrest(quietly bool) {
	if s.Prompt == PromptArrest {
		s.Prompt = PromptNone
	}
	if !quietly {
		s.Log(MsgDefendThyself)
		s.CallGuards()
		return
	}
	s.Log(MsgStruckUnconscious)
	// ⚠ 一次跳 20 分鐘直到**小時剛好是 8** —— 所以醒來的分鐘數取決於
	// 被抓的時間,不一定是整點。照抄比直接設成 08:00 接近原版。
	for n := 0; s.Clock.Hour != u5data.ArrestWakeHour && n < 24*60/u5data.ArrestWakeStep+2; n++ {
		s.AdvanceTime(u5data.ArrestWakeStep)
	}
	s.Inventory.Keys = 0
	s.Transport = u5data.VehicleWalk
	if s.Scenes == nil {
		return
	}
	if err := s.SetScene(u5data.ArrestJailLocation, 0,
		u5data.ArrestJailX, u5data.ArrestJailY); err != nil {
		s.Log("讀不到紫杉城的牢房(" + err.Error() + ")。")
		return
	}
	s.Log(MsgAwakenTo)
}
