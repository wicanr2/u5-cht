package game

import "github.com/wicanr2/u5-cht/internal/u5data"

// 每走一步都印方向名(原版 `sub_86C` / `sub_2D174` / `sub_15F18`)
//
// 這是 2026-08-09 對 DOSBox 原版並排比對時抓到的 —— 原版的訊息欄在走路時
// 是這樣的:
//
//	>South
//	>South
//	>South
//	>Open-South
//	Opened!
//
// 而引擎走三步**一句話都不印**。差別不是裝飾:原版的訊息欄是一條逐字累加的
// 紙帶(見 `commandecho.go`),方向名是「這一步生效了」的唯一回饋。
// 少了它,撞牆時只會看到孤零零一句「去路受阻!」,而走得動的時候畫面毫無反應。
//
// ★ 三支各自印一次,而條件不同 —— **全檔掃描**過方向字串的引用點,
// 共九支函式引用 `North/South/East/West`,其中三支是「移動時印」:
//
//	sub_86C    場景移動      **無條件**印
//	sub_2D174  大地圖移動    `byte_3E167 == 0` 才印 —— 那是**帆船的航向**
//	sub_15F18  戰鬥移動      無條件印(由 sub_A360 呼叫)
//
// 其餘六支是別的東西(`sub_2B2AC` 問方向已實作、`sub_2CCFC` 帆與風、
// `sub_2A984` 大船轉向的 "Head East "、`sub_4504` 六分儀、`sub_5008` 小寫的
// 描述文字、`sub_1CC50` 施法問方向)。
//
// ⚠ **印在移動判定之前**:原版是「先印方向,再看走不走得動」——
// 所以撞牆時讀起來是「西」→「去路受阻!」兩句。倒過來寫會變成
// 「去路受阻!」→「西」,而那是另一個意思。

// echoMoveDirection 在移動前印出方向名。
//
// ⚠ 大地圖上**揚著帆的船不印** —— 原版 `sub_2D174` 的 `byte_3E167` 是
// 帆船的當前航向,揚帆時一定非 0(同一支函式在開頭就設好),
// 而靠岸 / 碰撞 / 重進大地圖三處會清 0(`sub_2CE70` / `sub_2CBEC` / `sub_2D9D0`)。
// 船改印的是轉向那一句(`MsgHead`),不是方向名 —— 兩句都印會變成重複回報。
func (s *State) echoMoveDirection(d Direction) {
	if !s.InScene() && !s.InCombat() && s.Transport&0xFC == u5data.VehicleSailing {
		return
	}
	s.Log(d.Name())
}
