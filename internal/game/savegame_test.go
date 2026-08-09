package game

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// TestSaveLoadRoundTrip:玩一段 → 存檔 → 重讀,進度要留得住。
//
// 這是「存檔真的能用」的驗收 —— 不是「Encode 沒報錯」,而是把改過的東西
// (金幣、背包、時間、位置、隊伍狀態、地圖上的馬)寫出去再讀回來還在。
func TestSaveLoadRoundTrip(t *testing.T) {
	s := shopState(t, 22)
	if err := s.SetScene(22, 0, 8, 19); err != nil {
		t.Fatal(err)
	}
	// 買一匹馬、花掉一些錢、讓時間走一段。
	shop, ok := s.Shops.At(22, u5data.ShopStable)
	if !ok {
		t.Fatal("PAWS 沒有馬廄")
	}
	s.Inventory.Gold = 9999
	s.openShop(shop)
	s.ShopChoose('a')
	s.ShopChoose('y')
	s.LeaveShop()
	s.Roster[1].Status = u5data.StatusPoisoned
	s.Clock.Hour, s.Clock.Minute = 15, 30

	// 存檔到暫存目錄(絕不碰使用者的設定目錄,也不碰 gamedata)。
	dir := t.TempDir()
	sv, err := s.ExportSave(s.BaseSave)
	if err != nil {
		t.Fatal(err)
	}
	blob, err := sv.Encode()
	if err != nil {
		t.Fatal(err)
	}
	gamPath := filepath.Join(dir, SaveGameFile)
	if err := os.WriteFile(gamPath, blob, 0o644); err != nil {
		t.Fatal(err)
	}
	oolPath := filepath.Join(dir, SaveObjectsFile)
	if err := os.WriteFile(oolPath,
		u5data.EncodeSaveObjects(s.CurrentObjects(), s.UnderObjects), 0o644); err != nil {
		t.Fatal(err)
	}

	// 重讀。
	back, err := u5data.LoadSave(gamPath)
	if err != nil {
		t.Fatalf("寫出來的存檔讀不回來:%v", err)
	}
	s2 := shopState(t, 0)
	s2.LoadFrom(back)
	if s2.Inventory.Gold != s.Inventory.Gold {
		t.Errorf("金幣 %d,存檔前是 %d", s2.Inventory.Gold, s.Inventory.Gold)
	}
	if s2.Clock.Hour != 15 || s2.Clock.Minute != 30 {
		t.Errorf("時間 %02d:%02d,存檔前是 15:30", s2.Clock.Hour, s2.Clock.Minute)
	}
	if s2.Location != 22 {
		t.Errorf("地點 %d,存檔前是 22", s2.Location)
	}
	if s2.Roster[1].Status != u5data.StatusPoisoned {
		t.Errorf("第二位隊員的狀態是 %c,存檔前中毒", s2.Roster[1].Status)
	}
	// 馬要跟著存檔留下來。
	sur, _, err := u5data.LoadSaveObjects(oolPath)
	if err != nil {
		t.Fatal(err)
	}
	horse := false
	for i := range sur.Objects {
		if sur.Objects[i].Kind == u5data.TileHorse {
			horse = true
		}
	}
	if !horse {
		t.Error("買的馬沒有跟著存檔留下來")
	}
}

// TestExportSaveNeedsBase:沒有底稿不准存 —— 引擎只解出部分欄位,
// 硬存會把還沒解的區段(魔法、任務旗標、地牢狀態)清成 0,存檔在原版裡就壞了。
func TestExportSaveNeedsBase(t *testing.T) {
	s := &State{}
	if _, err := s.ExportSave(nil); err == nil {
		t.Error("沒有底稿也讓存了")
	}
	if _, err := s.ExportSave(&u5data.Save{}); err == nil {
		t.Error("底稿長度不對也讓存了")
	}
}

// TestSaveDirIsNotCwd:[HARD] 存檔絕不寫進工作目錄。
//
// u1-cht 踩過:遊戲裝在唯讀目錄時存不進去,而且玩家的原版資料目錄
// 不該被引擎污染。
func TestSaveDirIsNotCwd(t *testing.T) {
	dir, err := SaveDir()
	if err != nil {
		t.Skipf("這個環境沒有使用者設定目錄:%v", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if dir == cwd {
		t.Fatalf("存檔目錄就是工作目錄:%s", dir)
	}
	cfg, _ := os.UserConfigDir()
	if filepath.Dir(dir) != cfg {
		t.Errorf("存檔目錄 %s 不在使用者設定目錄 %s 底下", dir, cfg)
	}
}

// ★★ 即時存檔 / 讀回(F5 / F6)走的那一條路 —— `WriteSave` → `FindSave`。
//
// 與上面 `TestSaveLoadRoundTrip` 的差別是**這一條真的走設定目錄**:
// F5 存到 `os.UserConfigDir()`,F6 再用 `FindSave` 找回來。中間任何一段
// 對不上(目錄名、檔名、物件表沒一起存),症狀都是「按了 F5 沒事,
// 按 F6 回到很久以前」——而那比完全不能存更糟。
//
// ⚠ `t.Setenv` 把 `XDG_CONFIG_HOME` 指到暫存目錄 —— **絕不能碰使用者真正的
// 設定目錄**,否則跑一次測試就把玩家的進度蓋掉了。
func TestQuickSaveThenQuickLoad(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	s := shopState(t, 22)
	if err := s.SetScene(22, 0, 8, 19); err != nil {
		t.Fatal(err)
	}
	s.Inventory.Gold = 4321
	s.Clock.Hour, s.Clock.Minute = 21, 40
	s.Roster[2].Status = u5data.StatusPoisoned

	dir, err := s.WriteSave(s.BaseSave)
	if err != nil {
		t.Fatalf("即時存檔失敗:%v", err)
	}
	for _, name := range []string{SaveGameFile, SaveObjectsFile} {
		if st, err := os.Stat(filepath.Join(dir, name)); err != nil || st.Size() == 0 {
			t.Fatalf("%s 沒有落地(err=%v)", name, err)
		}
	}

	// 存完之後繼續玩 —— 讀回時這些改動應該消失。
	s.Inventory.Gold = 1
	s.Clock.Hour = 3
	s.Roster[2].Status = u5data.StatusGood

	sv, from, err := FindSave("", "")
	if err != nil {
		t.Fatalf("讀回失敗:%v", err)
	}
	if filepath.Dir(from) != dir {
		t.Fatalf("讀回的不是剛存的那一份:%s(預期在 %s)", from, dir)
	}
	if so, uo := FindSaveObjects(); so != nil {
		s.Objects, s.UnderObjects = so, uo
	} else {
		t.Fatal("物件表沒有跟著存檔一起讀回來")
	}
	s.LoadFrom(sv)

	if s.Inventory.Gold != 4321 {
		t.Errorf("金幣 %d,預期 4321", s.Inventory.Gold)
	}
	if s.Clock.Hour != 21 || s.Clock.Minute != 40 {
		t.Errorf("時間 %d:%02d,預期 21:40", s.Clock.Hour, s.Clock.Minute)
	}
	if s.Roster[2].Status != u5data.StatusPoisoned {
		t.Errorf("狀態 %q,預期中毒", string(s.Roster[2].Status))
	}
}
