package game

import (
	"fmt"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 橋下的食人妖(原版 `sub_3010` + `sub_2F48`)
//
// 從 `docs/re/66` 的截斷清單挖出來的:`sub_3010` 反編譯只剩三行,
// 而組語裡是一整個**過橋遭遇**。引擎原本走過橋什麼都不會發生。
//
// # 觸發(`sub_2D9D0` 移動之後)
//
//	腳下的 tile & 0xFE == 0x6A(橋)   →  sub_3010()
//
// # `sub_3010`:偷偷溜過去
//
//	if (rand(0, 7) != 0) return          ; ★ 1/8 才遇到
//	if (載具 != 0x1C)    return          ; ★ **只有步行**會遇到 —— 騎馬、坐船都不會
//	印 "\nThou spieth trolls under the bridge!\n\n"
//	for (每個隊員) {
//	    if (狀態 == 'D' || 狀態 == 'S') continue      ; 死的與睡著的不參加
//	    印 <名字> " sneaks across" 然後三個「.」(中間有延遲)
//	    if (敏捷 >= rand(1, 30)) continue             ; ★ 過關
//	    sub_2F48()                                    ; 被抓到 → 收過路費
//	    return
//	}
//	印 "Trolls evaded!"
//
// ⇒ **一個人被抓到就整隊停下**(`return`),後面的人不用擲。
//
// # `sub_2F48`:收過路費
//
//	印 "Caught!"
//	sub_2B67C()                                  ; 找第一個能行動的人 → word_3E086
//	通行費 = 0x63 − 那個人的**力量** × 3          ; ★ 99 − 力量×3
//	印 <通行費> "gp toll!" "Dost thou pay?"
//	等 'Y' / 'N'
//	'Y' → 扣錢;**扣完變負數就退回並開打**
//	'N' → 開打
//	開打:生成種類 **0xE4**(生物索引 41 = Troll)在腳下那一格,然後 sub_2E58C
//
// ⚠ 通行費用的是**力量**(`byte_3DDC0`,偏移 0x0C),而偷渡擲的是**敏捷**
// (`byte_3DDC1`,偏移 0x0D)—— 兩個相鄰欄位,很容易看錯成同一個。
// 兩者都能用名冊基底獨立驗:`byte_3DDBF` − `byte_3DDB4` = 0x0B = `CharStatus`。
//
// ⚠ **原版的算式沒有下限**:力量 34 以上通行費就是負數,而 `word_3DFB6 -= 費用`
// 會變成**加錢**。照原樣保留 —— U5 的力量上限雖然是 99,但真的練到 34
// 就能靠橋下的食人妖賺錢。那是原版的算式,不是這裡寫錯。

// 過橋遭遇的常數,全部來自 `sub_3010` / `sub_2F48`。
const (
	// TrollBridgeTile 是橋(`sub_2D9D0`:`and eax, 0FEh; cmp eax, 6Ah`)。
	TrollBridgeTile = 0x6A
	// TrollChanceDenominator 是遭遇機率的分母(`rand(0, 7) != 0` → 1/8)。
	TrollChanceDenominator = 8
	// TrollSneakRollMax 是偷渡擲骰的上限(`rand(1, 30)`)。
	TrollSneakRollMax = 30
	// TrollTollBase / TrollTollPerStrength 是通行費:`99 − 力量 × 3`。
	TrollTollBase        = 0x63
	TrollTollPerStrength = 3
	// TrollCreatureKind 是生成的怪物種類碼(生物索引 41 = Troll)。
	TrollCreatureKind = 0xE4
)

// crossBridge 是走上橋那一步的遭遇判定。
func (s *State) crossBridge() {
	if s.Roll(0, TrollChanceDenominator-1) != 0 {
		return
	}
	// ★ 只有步行會遇到 —— 原版 `cmp byte_3E08C, 1Ch; jnz`(**單一值**,
	// 不是 `IsOnFoot` 那個「0x1C 或 0x1D」的兩值判斷)。
	if s.Transport != u5data.VehicleWalk {
		return
	}
	s.Log(MsgTrollsUnderBridge)
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		ch := &s.Roster[i]
		if ch.Status == u5data.StatusDead || ch.Status == u5data.StatusAsleep {
			continue
		}
		s.Log(ch.Name + MsgSneaksAcross)
		if int(ch.Dex) >= s.Roll(1, TrollSneakRollMax) {
			continue
		}
		s.trollToll()
		return
	}
	s.Log(MsgTrollsEvaded)
}

// trollToll 是被抓到之後的過路費(原版 `sub_2F48`)。
func (s *State) trollToll() {
	s.Log(MsgCaught)
	// ⚠ 收費看的是**第一個能行動的人**(原版 `sub_2B67C` 把編號留在 `word_3E086`),
	// 不是被抓到的那一個。
	payer := s.firstAbleMember()
	strength := 0
	if payer >= 0 && payer < len(s.Roster) {
		strength = int(s.Roster[payer].Strength)
	}
	// ★ 沒有下限 —— 見檔頭的說明。
	toll := TrollTollBase - strength*TrollTollPerStrength
	s.Log(fmt.Sprintf(MsgTrollToll, toll))
	s.Ask(MsgDostThouPay, func(pay bool) {
		if !pay {
			s.trollFight()
			return
		}
		// 扣完變負數就退回並開打(原版 `add word_3DFB6, ax` 之後生怪)。
		s.Inventory.Gold -= toll
		if s.Inventory.Gold < 0 {
			s.Inventory.Gold += toll
			s.trollFight()
		}
	})
}

// firstAbleMember 是 `sub_2B67C` 留在 `word_3E086` 的那個編號。
//
// 與 `anyoneCanAct` 是同一支函式的兩個出口:這裡要的是**編號**。
func (s *State) firstAbleMember() int {
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		switch s.Roster[i].Status {
		case u5data.StatusGood, u5data.StatusPoisoned:
			return i
		case u5data.StatusAsleep:
			continue
		default:
			return -1
		}
	}
	return -1
}

// trollFight 生一隻食人妖在腳下那一格然後開打。
func (s *State) trollFight() {
	o := &u5data.MapObject{X: s.X, Y: s.Y, Kind: TrollCreatureKind}
	o.Raw[u5data.ObjKind] = TrollCreatureKind
	o.Raw[u5data.ObjX], o.Raw[u5data.ObjY] = byte(s.X), byte(s.Y)
	s.beginCombatWith(o)
}

// 大地圖上的攀爬(原版 `sub_188C4`)
//
// 引擎原本在這裡寫「大地圖上的攀爬(上山、進地牢)是另一條路徑,**還沒做**」。
// `sub_188C4` 就是那條路徑,而它被 Hex-Rays 截斷成三行(`docs/re/66`)。
//
//	if (byte_3DFBB == 0) { 印 "With what?"; return }      ; ★ 沒抓鉤
//	if (載具 != 0x1C)    { 印 "On foot!";  return }       ; ★ 只有步行
//	方向 = sub_2B2AC()   ; 取消就結束
//	tile = 目標格
//	if (tile == 0x0D) { 印 "Impassable!";   return }      ; ★ 峭壁
//	if (tile != 0x0C) { 印 "Not climbable!"; return }     ; ★ 只有群山能爬
//	for (每個沒死的隊員)
//	    if (rand(1, 30) > 敏捷) { 印 "Fell!"; 那個人受 rand(1, 5) 傷 }
//	sub_2D014(dx, dy)                                     ; ★ 不論摔幾個人,整隊都過去
//
// 兩個地形值都對得上 `look#<tile>`:0x0C 是「群山」、0x0D 是「峭壁」。
// 而**峭壁與「不能爬的東西」是兩句不同的話** —— 峭壁是「過不去」,
// 其餘地形是「這不能爬」。合成一句會少掉那個區分。
//
// ⚠ 摔倒**不會擋住移動** —— 原版的 `sub_2D014` 在迴圈之後、無條件執行。
// 寫成「有人摔倒就不過去」是很自然的想像,但那不是原版。

// 攀爬用的常數。
const (
	// ClimbMountain 是可以攀的地形(`look#12` = 群山)。
	ClimbMountain = 0x0C
	// ClimbCliff 是峭壁(`look#13`)—— 過不去,而且訊息與「不能爬」不同。
	ClimbCliff = 0x0D
	// ClimbFallRollMax 是摔倒判定的骰上限(`rand(1, 30)`)。
	ClimbFallRollMax = 30
	// ClimbFallDamageMax 是摔倒的傷害上限(`rand(1, 5)`)。
	ClimbFallDamageMax = 5
)

// klimbOverworld 是大地圖上的 K(原版 `sub_188C4`)。
func (s *State) klimbOverworld() {
	if !s.hasRope() {
		s.Log(MsgWithWhat)
		return
	}
	// ★ 單一值 0x1C —— 與食人妖那條同一個寫法。
	if s.Transport != u5data.VehicleWalk {
		s.Log(MsgOnFootOnly)
		return
	}
	s.AskDirection(func(d Direction) {
		dx, dy := d.Delta()
		nx, ny := WrapWorld(s.X+dx), WrapWorld(s.Y+dy)
		switch s.TileAt(nx, ny) {
		case ClimbCliff:
			s.Log(MsgImpassable)
			return
		case ClimbMountain:
		default:
			s.Log(MsgNotClimbable)
			return
		}
		for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
			if s.Roster[i].Status == u5data.StatusDead {
				continue
			}
			if s.Roll(1, ClimbFallRollMax) > int(s.Roster[i].Dex) {
				s.Log(s.Roster[i].Name + MsgFell)
				s.damageMember(i, s.Roll(1, ClimbFallDamageMax))
			}
		}
		// ★ 摔倒不擋移動 —— 原版無條件走這一步。
		s.X, s.Y = nx, ny
		s.tick()
	})
}

// 瀑布:掉下去,而世界上只有一個洞通往幽冥界(原版 `sub_10A1C`)
//
// 觸發有兩處,判的都是 `tile & 0xFC == 0xD4`(`look#212..215` 四格都是「瀑布」):
//
//	sub_2D9D0  腳下那一格是瀑布(移動之後)
//	sub_2D2D0  漂流之後,`byte_3F7A9`(視窗緩衝往南一列)是瀑布
//
// # `sub_10A1C`
//
//	印 "F-A-L-L-S!!!"
//	…畫面效果:載具碼暫設 0(不畫載具圖)、重畫、**再還原**…
//	for (每個沒死的隊員)
//	    門檻 = max(1, rand(0, 60) / 2)          ; sub_2B724,與拖屍怪掙脫同一支
//	    if (敏捷 <= 門檻) 那個人受 **1** 點傷
//	if (X == 0x36 && Y == 0x8A) {               ; ★ 寫死的座標 (54, 138)
//	    印 "Falling into underworld!!"
//	    樓層 = −1
//	    存 A:BRIT.OOL、載 A:UNDER.OOL、換地圖 UNDER.DAT
//	}
//
// ★ 兩件事值得記:
//
//  1. **傷害是固定 1 點**,不是骰的 —— 擲的只是「有沒有受傷」。
//     瀑布不是陷阱,是交通工具。
//  2. **通往幽冥界的座標寫死在程式裡**,世界上只有那一個瀑布下得去。
//     其餘瀑布只會讓你掉一次、扣一點血,原地不動。
//
// ⚠ 那個畫面效果**有還原**載具碼(`mov eax, esi; mov byte_3E08C, al`),
// 與 `sub_22F0` 溺水那一段的「設 0 之後不還原」正好成對 ——
// 所以那裡的 0 真的是「不畫載具圖」而不是某種載具,兩處互相印證。

// 瀑布的常數。
const (
	// FallTileGroup 是瀑布(`tile & 0xFC == 0xD4`,四格)。
	FallTileGroup = 0xD4
	// FallDamage 是掉下去受的傷 —— **固定 1 點**,不是骰的。
	FallDamage = 1
	// UnderworldHoleX / UnderworldHoleY 是唯一通往幽冥界的那個瀑布
	// (原版寫死 `cmp byte_3E0A6, 36h` / `cmp byte_3E0A7, 8Ah`)。
	UnderworldHoleX = 0x36
	UnderworldHoleY = 0x8A
)

// fallDownTheWaterfall 是踩到瀑布那一格(原版 `sub_10A1C`)。
func (s *State) fallDownTheWaterfall() {
	// ★ A 級證據:原版的墜落動畫 `sub_10A1C` 用索引 0x14 = T_OCHI1(「落ち」)。
	s.PlaySFX(u5data.SFXFall)
	s.Log(MsgFalls)
	for i := 0; i < s.PartySize && i < len(s.Roster); i++ {
		if s.Roster[i].Status == u5data.StatusDead {
			continue
		}
		// 門檻與拖屍怪掙脫同一支 `sub_2B724`。
		threshold := u5data.CorpserEscapeThreshold(s.Roll(0, u5data.CorpserEscapeRollMax))
		if int(s.Roster[i].Dex) <= threshold {
			s.damageMember(i, FallDamage)
		}
	}
	// ★ 只有那一個座標下得去。
	if s.X != UnderworldHoleX || s.Y != UnderworldHoleY {
		return
	}
	if s.Under == nil {
		// 沒載幽冥界地圖就誠實停在這裡,不假裝掉下去(`CLAUDE.md` §3.0)。
		return
	}
	s.Log(MsgFallingIntoUnderworld)
	s.Floor = -1
	// 原版的墜落動畫 `sub_10A1C` 呼叫 `sub_2CBEC` 重新定位視窗(`docs/re/88`)。
	s.resetLoadWindow()
	s.placeUnderworldItems()
}
