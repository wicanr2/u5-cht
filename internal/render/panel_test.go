package render

import (
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

func panelScene(mode UIMode, members int, debug bool) *Scene {
	s := testScene()
	s.UI, s.ShowDebug = mode, debug
	for i := 0; i < members; i++ {
		s.State.Roster = append(s.State.Roster, u5data.Character{
			Name: "Elwood", HP: 60, MaxHP: 60, Status: u5data.StatusGood,
		})
	}
	s.State.PartySize = members
	return s
}

// 兩種版面要真的不一樣 —— 否則 F2 按下去玩家看不出差別。
func TestUIModesDiffer(t *testing.T) {
	a := panelScene(UIOriginal, 3, false).Render()
	b := panelScene(UIModern, 3, false).Render()
	if sameImage(a, b) {
		t.Fatal("原版版面與現代版面畫出來一模一樣")
	}
}

// ★★ 訊息欄**不能疊在右欄的資料上**。
//
// 第一版把現代版面的訊息起點寫成常數,而組間留白讓右欄長了三行 ⇒
// 訊息直接蓋在隊伍最後一行上。症狀看起來像「字疊在一起」,
// 根因是「版面高度是算出來的,不是常數」。
//
// ⚠ 這條要掃**最壞情況**:六人隊伍 + 除錯欄位 + 兩種版面。
// 只測三人隊伍會綠 —— 而滿隊伍才是玩家的常態。
func TestMessagesNeverOverlapThePanel(t *testing.T) {
	for _, mode := range []UIMode{UIOriginal, UIModern} {
		for _, debug := range []bool{false, true} {
			s := panelScene(mode, 6, debug)
			s.Render()
			if s.panelBottom >= s.messageTop() {
				t.Errorf("%s(除錯=%v):右欄畫到 y=%d,而訊息從 y=%d 開始 —— 會疊字",
					UIModeNames[mode], debug, s.panelBottom, s.messageTop())
			}
		}
	}
}

// 現代版面不畫標題(視窗標題已經有了),原版版面要畫 —— 這是兩者的定義差異之一。
func TestOnlyOriginalModeDrawsTheTitle(t *testing.T) {
	countRows := func(mode UIMode) int {
		n := 0
		for _, g := range panelScene(mode, 3, false).panelData(false) {
			n += len(g.rows)
		}
		return n
	}
	orig, modern := countRows(UIOriginal), countRows(UIModern)
	if orig != modern+2 {
		t.Errorf("原版版面 %d 列、現代版面 %d 列,預期差 2(標題兩行)", orig, modern)
	}
}

// 除錯欄位預設關著 —— 座標與地形碼不是原版的欄位。
func TestDebugFieldsAreOffByDefault(t *testing.T) {
	s := panelScene(UIModern, 3, false)
	with := len(s.panelData(true)[0].rows)
	without := len(s.panelData(false)[0].rows)
	if with != without+1 {
		t.Errorf("除錯欄位開關沒有差一列(%d vs %d)", with, without)
	}
	if s.ShowDebug {
		t.Error("ShowDebug 預設應該是關的")
	}
}
