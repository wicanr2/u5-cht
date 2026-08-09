package appui

import "testing"

// ★★★ 這一條是本檔存在的理由:**ESC 在任何狀態下都不會結束遊戲。**
//
// 來源是 wizardry-1-cht 的真實事件 —— QA 兩次誤觸 ESC,各噴掉一小時進度。
// 所以不是「試幾種情況」,而是把 ESC 與其他鍵的**所有組合**掃過一遍,
// 只要有一組回 `ActSaveAndQuit` 就紅。
func TestEscapeNeverQuits(t *testing.T) {
	for _, open := range []bool{false, true} {
		for _, other := range []Keys{
			{Escape: true},
			{Escape: true, Quit: true},
			{Escape: true, Save: true},
			{Escape: true, Load: true},
			{Escape: true, No: true},
			// ⚠ 連「ESC 與 Yes 同一帧」都要驗 —— 兩顆鍵同時按進來時
			// 順序判斷寫錯的話,ESC 會變成確認。
			{Escape: true, Yes: true},
		} {
			d := &QuitDialog{open: open}
			got := d.Step(other)
			if got == ActSaveAndQuit && !other.Yes {
				t.Fatalf("open=%v keys=%+v:ESC 讓遊戲結束了", open, other)
			}
			if !open && got == ActSaveAndQuit {
				t.Fatalf("keys=%+v:確認框沒開就結束了", other)
			}
		}
	}
}

// F10 只把框打開 —— 不存檔、不離開。原版的 `Q)uit & Save` 直接結束是原版的事,
// 而這一層是本重製版加的,手滑不該有代價。
func TestQuitKeyOnlyOpensTheDialog(t *testing.T) {
	d := &QuitDialog{}
	if got := d.Step(Keys{Quit: true}); got != ActOpenedQuit {
		t.Fatalf("F10 應該只開框,得到 %v", got)
	}
	if !d.IsOpen() {
		t.Fatal("F10 之後框應該是開著的")
	}
	// 再按一次 F10 不該有第二個效果(框已經開著,F10 不是確認鍵)。
	if got := d.Step(Keys{Quit: true}); got != ActNone {
		t.Fatalf("框開著時 F10 應該沒有作用,得到 %v", got)
	}
}

func TestYesSavesAndQuitsOnlyWhenAsked(t *testing.T) {
	d := &QuitDialog{}
	// 框沒開時按 Y 是遊戲裡的「是」,不該被這一層攔下來當離開。
	if got := d.Step(Keys{Yes: true}); got != ActNone {
		t.Fatalf("框沒開時 Y 不該離開,得到 %v", got)
	}
	d.Step(Keys{Quit: true})
	if got := d.Step(Keys{Yes: true}); got != ActSaveAndQuit {
		t.Fatalf("框開著時 Y 應該存檔並離開,得到 %v", got)
	}
	// ⚠ 回報之後框**還是開著** —— 存檔失敗時呼叫端要能留在遊戲裡,
	// 而不是框已經自己關了、玩家還以為存好了。
	if !d.IsOpen() {
		t.Fatal("ActSaveAndQuit 之後框不該自己關掉(存檔可能失敗)")
	}
}

func TestNoAndEscapeBothCancel(t *testing.T) {
	for name, k := range map[string]Keys{"N": {No: true}, "ESC": {Escape: true}} {
		d := &QuitDialog{}
		d.Step(Keys{Quit: true})
		if got := d.Step(k); got != ActCancelled {
			t.Fatalf("%s 應該取消,得到 %v", name, got)
		}
		if d.IsOpen() {
			t.Fatalf("%s 取消之後框應該關掉", name)
		}
	}
}

// ★ 確認框是 modal:開著的時候存檔 / 讀檔鍵不作用。
// 少了這一條,玩家在框上按 F5 會得到「存了、沒離開、框還開著」的怪狀態。
func TestDialogIsModalOverSaveKeys(t *testing.T) {
	d := &QuitDialog{}
	d.Step(Keys{Quit: true})
	for _, k := range []Keys{{Save: true}, {Load: true}} {
		if got := d.Step(k); got != ActNone {
			t.Fatalf("框開著時 %+v 不該有作用,得到 %v", k, got)
		}
	}
	// 反對照:框關著時同樣的鍵**要**有作用 —— 否則上面那條可能只是
	// 因為存檔鍵整個沒接而通過。
	d.Close()
	if got := d.Step(Keys{Save: true}); got != ActQuickSave {
		t.Fatalf("框關著時 F5 應該即時存檔,得到 %v", got)
	}
	if got := d.Step(Keys{Load: true}); got != ActQuickLoad {
		t.Fatalf("框關著時 F6 應該讀回存檔,得到 %v", got)
	}
}
