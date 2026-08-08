package u5data

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestEUPFieldLayoutIsProvenByNoteBalance —— ★★★ 欄位配置的決定性檢查。
//
// 一首正常的曲子裡 note-on 與 note-off 必須幾乎相等(差 0 或 1)。
// 相位讀錯的話 status 欄根本不是 status,這個關係就不成立 ——
// 實測錯相位(sig+0x40)時 status 的值域是 0..120、高 nibble 完全散開。
//
// ⇒ 這條測試取代了「相信註解裡那個 16」。
func TestEUPFieldLayoutIsProvenByNoteBalance(t *testing.T) {
	dir := fmTownsDir(t)
	tracks, err := LoadBGMTable(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, tr := range tracks {
		song, err := LoadEUP(dir, tr.File)
		if err != nil {
			t.Errorf("%s:%v", tr.File, err)
			continue
		}
		on, off := song.NoteCounts()
		if on == 0 {
			t.Errorf("%s 一個 note-on 都沒有", tr.File)
			continue
		}
		if d := on - off; d < -1 || d > 1 {
			t.Errorf("%s note-on %d / note-off %d 差 %d —— 相位可能讀錯了",
				tr.File, on, off, d)
		}
		if song.TotalTicks <= 0 {
			t.Errorf("%s 累積 tick 是 %d", tr.File, song.TotalTicks)
		}
	}
}

// TestNoteOffCarriesNoteZero —— note-off 的音高欄一律 0。
//
// ⇒ 收哪一個音是「該聲道當前的那一個」,不是靠音高比對。
// 渲染時的收尾邏輯要照這個來,否則同聲道連續兩個音會收錯。
func TestNoteOffCarriesNoteZero(t *testing.T) {
	dir := fmTownsDir(t)
	song, err := LoadEUP(dir, "M1.EUP")
	if err != nil {
		t.Fatal(err)
	}
	for i, e := range song.Events {
		if e.Kind() == EUPNoteOff && e.Data1 != 0 {
			t.Fatalf("第 %d 筆 note-off 的音高是 %d,預期 0", i, e.Data1)
		}
		if e.Kind() == EUPNoteOn && e.Data1 == 0 {
			t.Fatalf("第 %d 筆 note-on 的音高是 0", i)
		}
	}
}

// TestProgramChangesStayInsideTheBank —— ★ 兩個獨立來源互相佐證。
//
// EUP 的 program change 值域必須落在 `FM_BANK.FMB` 的 128 筆內。
// 音色庫的筆數是從檔案大小整除算出來的(8 + 128 × 48 = 6152),
// 而曲子裡的編號是另一份資料 —— 兩者一致才敢說佈局對。
func TestProgramChangesStayInsideTheBank(t *testing.T) {
	dir := fmTownsDir(t)
	bank, err := LoadFMBank(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(bank) != FMVoiceCount {
		t.Fatalf("音色庫 %d 筆,預期 %d", len(bank), FMVoiceCount)
	}
	tracks, err := LoadBGMTable(dir)
	if err != nil {
		t.Fatal(err)
	}
	used := 0
	for _, tr := range tracks {
		song, err := LoadEUP(dir, tr.File)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range song.Events {
			if e.Kind() != EUPProgram {
				continue
			}
			if int(e.Data1) >= len(bank) {
				t.Errorf("%s 選了音色 %d,音色庫只有 %d 筆", tr.File, e.Data1, len(bank))
			}
			used++
		}
		// 每首至少替某些聲道選過音色。
		if progs := song.Programs(); progs[0] < 0 && progs[1] < 0 {
			t.Errorf("%s 前兩個聲道都沒選音色", tr.File)
		}
	}
	if used == 0 {
		t.Error("十五首裡一個 program change 都沒有")
	}
	t.Logf("十五首共 %d 筆選音色", used)
}

// TestFMBankVoicesHaveReadableNames —— 音色名是可讀 ASCII。
//
// 這條把「名字在記錄開頭 8 byte」釘住:位移偏一格就會讀到參數位元組,
// 而參數是二進位的 ⇒ 名字立刻變亂碼。
func TestFMBankVoicesHaveReadableNames(t *testing.T) {
	bank, err := LoadFMBank(fmTownsDir(t))
	if err != nil {
		t.Fatal(err)
	}
	named := 0
	for i, v := range bank {
		if v.Name == "" {
			continue
		}
		named++
		for _, r := range v.Name {
			if r < 0x20 || r > 0x7E {
				t.Errorf("音色 %d 的名字 %q 含不可讀字元 0x%02X", i, v.Name, r)
				break
			}
		}
	}
	if named < FMVoiceCount/2 {
		t.Errorf("只有 %d/%d 筆有名字 —— 位移可能偏了", named, FMVoiceCount)
	}
	// 幾個一眼認得出來的樂器名(位移對了才讀得到)。
	joined := strings.Join([]string{bank[0].Name, bank[7].Name, bank[9].Name}, "|")
	for _, want := range []string{"TRUMP1", "FLUTE", "OBOE"} {
		if !strings.Contains(joined, want) {
			t.Errorf("前十筆裡找不到 %q(讀到的是 %q)", want, joined)
		}
	}
}

// TestFMBankRejectsWrongSize —— 大小不對就報錯,不要「盡量解」。
func TestFMBankRejectsWrongSize(t *testing.T) {
	if _, err := ParseFMBank(make([]byte, 6151)); err == nil {
		t.Error("少一個位元組也該報錯")
	}
	if _, err := ParseFMBank(nil); err == nil {
		t.Error("空檔案該報錯")
	}
}

// TestEveryBGMFileExists —— `U5_BGM.TBL` 列的十五個檔案都在。
func TestEveryBGMFileExists(t *testing.T) {
	dir := fmTownsDir(t)
	tracks, err := LoadBGMTable(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, tr := range tracks {
		if _, err := os.Stat(filepath.Join(dir, tr.File)); err != nil {
			t.Errorf("%s:%v", tr.File, err)
		}
	}
}
