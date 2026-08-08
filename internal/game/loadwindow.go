package game

import "github.com/wicanr2/u5-cht/internal/u5data"

// 大地圖的載入視窗(原版 `byte_3E0AB` / `byte_3E0AC`)
//
// 推導見 `docs/re/88`。原版在大地圖上**不是**每格都重新載入 32×32 的地圖窗,
// 而是維護一個**對齊到 16 的倍數**的左上角,只有隊伍走進邊緣 5 格以內才
// **整塊捲動 16 格**。生怪落點、月門要不要寫、清場的距離判定全部以它為基準。
//
// ⚠ 此前引擎用「隊伍為中心的 32×32」近似。差別是可觀察的:真視窗的原點
// 一步一步**不動**,所以「離隊伍 7 格以上」的生怪落點在同一個視窗裡會落在
// 固定的一圈;近似版則跟著隊伍走,那一圈永遠貼著隊伍。

// WindowAlign 是視窗原點的對齊粒度(原版 `and al, 0F0h`)。
const WindowAlign = 0x10

// WindowScrollMargin 是「隊伍離視窗邊緣幾格以內就捲動」(原版 `cmp edx, 5; jl`)。
//
// 原版的條件是**隊伍在視窗內的座標落在 5..0x1A 之間就不捲**;兩邊各留 5 格,
// 而視窗是 32 格寬(0..0x1F)⇒ 邊緣 5 格是緩衝區。
const WindowScrollMargin = 5

// resetLoadWindow 重算視窗原點(原版 `sub_2CBEC`,進大地圖 / 讀檔時跑一次)。
//
//	原點 = 隊伍座標 & 0xF0
//	if (隊伍座標 & 0x0F) < 8 → 原點 = (原點 − 0x10) & 0xF0
//
// ★ 第二行是「挑離隊伍最近的那個對齊點」:隊伍落在 16 格區塊的前半就再退一塊,
// 這樣隊伍在視窗裡的位置永遠落在 8..23(視窗 32 格寬)⇒ 大致居中。
func (s *State) resetLoadWindow() {
	s.WindowX = alignLoadWindow(s.X)
	s.WindowY = alignLoadWindow(s.Y)
}

func alignLoadWindow(v int) int {
	o := v & 0xF0
	if v&0x0F < WindowAlign/2 {
		o = (o - WindowAlign) & 0xF0
	}
	return o
}

// scrollLoadWindow 在隊伍走到視窗邊緣時整塊捲動(原版 `sub_2D014` 的後半)。
//
//	在視窗內的座標 = (隊伍座標 − 原點) & 0x1F
//	兩軸都落在 5..0x1A → 什麼都不做
//	否則 → 原點 = ((步向 << 4) + 原點) & 0xF0        ; ★ 一次捲 16 格
//
// ⚠ **一次捲半個視窗**,不是跟著走一格。這就是原版大地圖走到邊緣時
// 「畫面整塊跳一下」的來源。
//
// ⚠ 捲動量用的是**這一步的步向**(dx / dy 各為 −1 / 0 / +1),所以斜走會兩軸一起捲。
func (s *State) scrollLoadWindow(dx, dy int) {
	inX := (s.X - s.WindowX) & u5data.SpawnWindowSpan
	inY := (s.Y - s.WindowY) & u5data.SpawnWindowSpan
	if inWindowMargin(inX) && inWindowMargin(inY) {
		return
	}
	s.WindowX = ((dx * WindowAlign) + s.WindowX) & 0xF0
	s.WindowY = ((dy * WindowAlign) + s.WindowY) & 0xF0
}

func inWindowMargin(v int) bool {
	return v >= WindowScrollMargin && v <= u5data.SpawnWindowSpan-WindowScrollMargin
}

// CullFromSlot 是清場掃描的起點(原版 `mov edi, 1Fh`)。
//
// ⚠ 迴圈是 `while (edi > 0)` ⇒ **槽 0 不清**。倒著掃到 1 就停,
// 與世界回合的移動迴圈同一個形狀。照原樣做。
const CullFromSlot = 0x1F

// cullDistantCreatures 把離視窗太遠的生物整槽清掉(原版 `sub_2E24` 的尾段)。
//
//	for (槽 = 0x1F; 槽 > 0; 槽--) {
//	    if (!sub_22B0(物件[槽].Kind)) continue        ; ★ 位移 0,不是位移 1
//	    if (((物件X − 原點X) & 0xFF) > 0x1F) → 清
//	    if (((物件Y − 原點Y) & 0xFF) > 0x1F) → 清
//	}
//
// ★ 三件事要注意:
//   - 判準是**視窗原點**不是隊伍座標 ⇒ 視窗不捲的那些回合,怪走遠了也不會被清。
//   - 只清**生物**(`IsCreatureTile`)⇒ 船、馬、小艇、寶箱、屍體不會被清掉。
//     少了這個條件,玩家停在岸邊的船會憑空消失。
//     ★ 靠的是 `sub_22B0` 對 `< 0x2C` 回 0,而載具碼是 0x10/0x14/0x24/0x28 ——
//     原版 `sub_2DD44` 把載具放回地圖時寫的種類碼是 **0x25(船)或 0x29(小艇)**,
//     兩個都 < 0x2C ⇒ 引擎的 `dismountShip` 用 `Transport`(0x24..0x2B)也落在同一區。
//   - ⚠⚠ **判準吃 `ObjKind`(位移 0),不是 `ObjTile`(位移 1)。**
//     原版是 `movzx eax, byte ptr dword_3E46C[edi*8]` —— 沒有 `+1`。
//     第一版寫成 `ObjTile`,而 `Spawn` 與測試輔助都把兩個欄位設成同一個值,
//     所以測試抓不到。兩者一旦分歧(存檔載入的物件、`turnBroadside` 改過的敵船)
//     就會清錯槽。`TestCullingJudgesByKindNotTile` 用「兩欄故意不同」釘住它。
func (s *State) cullDistantCreatures() {
	set := s.currentObjects()
	if set == nil {
		return
	}
	for slot := CullFromSlot; slot > 0; slot-- {
		o := &set.Objects[slot]
		if !u5data.IsCreatureTile(o.Raw[u5data.ObjKind]) {
			continue
		}
		dx := (int(o.Raw[u5data.ObjX]) - s.WindowX) & 0xFF
		dy := (int(o.Raw[u5data.ObjY]) - s.WindowY) & 0xFF
		if dx > u5data.SpawnWindowSpan || dy > u5data.SpawnWindowSpan {
			*o = u5data.MapObject{}
		}
	}
}
