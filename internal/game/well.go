package game

import "github.com/wicanr2/u5-cht/internal/u5data"

// 許願井(原版 `sub_CD28`)
//
// 規則、字面值、兩個地點與比對方式全部寫在 `u5data/well.go` 的說明裡,
// 包含「當初為什麼會把這一格結論寫錯」——Hex-Rays 把 `sub_CD28` 截斷了
// (`CLAUDE.md` §4.4)。這裡只負責把那個流程接上輸入。

// lookAtWell 是 Look 撞到井那一格(tile 0xA1)。
func (s *State) lookAtWell() {
	s.Log(MsgLookWell)
	// Ask 自己會把 "Yes\n" / "No\n" 印出來(原版也是),所以這裡不重複印。
	s.Ask(MsgDropACoin, func(yes bool) {
		if !yes {
			return
		}
		// ★ 沒錢就**什麼都不印**直接結束 —— 原版 `cmp word_3DFB6,0; jz` 那條路
		// 連一句「你沒錢」都沒有。看起來像卡住,但那就是原版。
		if s.Inventory.Gold < u5data.WellCoin {
			return
		}
		// ★ 先扣錢,再問願望 —— 順序照原版(`dec word_3DFB6` 在讀輸入之前)。
		s.Inventory.Gold -= u5data.WellCoin
		s.AskText(MsgThyWish, u5data.WellWishMax, func(wish string) {
			s.grantWish(wish)
		})
	})
}

// grantWish 判定願望(原版 `sub_CD28` 的後半)。
func (s *State) grantWish(wish string) {
	if wish == "" {
		s.Log(MsgNothing)
		return
	}
	// ⚠ 比對用大寫字面值的**前綴**,而且原版不對輸入做大小寫轉換 ——
	// 所以小寫的願望在原版不生效。見 `u5data/well.go`。
	if u5data.WellWishMatches(wish) < 0 || !u5data.WellGrantsHere(s.Location) {
		s.Log(MsgNoEffect)
		return
	}
	s.Log(MsgPoof)
	// 原版寫死放在 (玩家X + 1, 玩家Y, 樓層),**不檢查那一格能不能站**
	// (`sub_2B6C8(10h, 10h, arg_0+1, arg_4, arg_8, 0, 空槽)`)。
	// 這裡照做 —— 買馬那條路才會挑方向,許願井不挑。
	objs := s.currentObjects()
	if objs == nil {
		return
	}
	objs.Spawn(u5data.TileHorse, s.X+1, s.Y, s.Floor)
}
