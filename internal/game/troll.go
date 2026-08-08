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
