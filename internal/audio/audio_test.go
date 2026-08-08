package audio

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// fakeBackend 記下被要求做了什麼。
type fakeBackend struct {
	played []string
	stops  int
}

func (f *fakeBackend) Play(path string) error { f.played = append(f.played, path); return nil }
func (f *fakeBackend) Stop()                  { f.stops++ }
func (f *fakeBackend) SetVolume(float64)      {}

// tblDir 是 FM Towns 的 U5_E;沒有就跳過(原版資料不入庫)。
func tblDir(t *testing.T) string {
	t.Helper()
	dir := os.Getenv("U5_FMTOWNS")
	if dir == "" {
		t.Skip("未設 U5_FMTOWNS")
	}
	p := filepath.Join(dir, "U5_E")
	if _, err := os.Stat(filepath.Join(p, "U5_BGM.TBL")); err != nil {
		t.Skipf("%s 讀不到 U5_BGM.TBL", p)
	}
	return p
}

// TestFileNamesComeFromTheTableNotFromArithmetic —— ★ 檔名編號不連續。
//
// 曲號 8 對 `M92`、曲號 14 對 `M152` ⇒ 「曲號 = 檔名數字 − 1」是錯的
// (`docs/re/87` §2)。這條測試釘住「一定要讀表」。
func TestFileNamesComeFromTheTableNotFromArithmetic(t *testing.T) {
	dir := tblDir(t)
	// 假裝每一首都渲染好了。
	audioDir := t.TempDir()
	tracks, err := u5data.LoadBGMTable(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, tr := range tracks {
		name := tr.File[:len(tr.File)-len(filepath.Ext(tr.File))] + Ext
		if err := os.WriteFile(filepath.Join(audioDir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	p, err := New(dir, audioDir, &fakeBackend{})
	if err != nil {
		t.Fatal(err)
	}
	if n := len(p.Missing()); n != 0 {
		t.Fatalf("全部渲染好了卻回報缺 %d 個:%v", n, p.Missing())
	}
	for _, tc := range []struct{ song int; want string }{
		{0, "M1" + Ext},
		{1, "M2" + Ext},
		{8, "M92" + Ext},   // ★ 不是 M9
		{14, "M152" + Ext}, // ★ 不是 M15
	} {
		if got := p.FileFor(tc.song); got != tc.want {
			t.Errorf("曲號 %d → %q,預期 %q", tc.song, got, tc.want)
		}
	}
	if got := p.FileFor(u5data.BGMSongCount); got != "" {
		t.Errorf("超出範圍的曲號回了 %q", got)
	}
	if got := p.FileFor(-1); got != "" {
		t.Errorf("負曲號回了 %q", got)
	}
}

// TestMissingRendersDegradeQuietly —— 缺檔要安靜降級並且**說得出缺什麼**。
func TestMissingRendersDegradeQuietly(t *testing.T) {
	dir := tblDir(t)
	audioDir := t.TempDir() // 完全空的 —— 還沒跑過轉檔
	b := &fakeBackend{}
	p, err := New(dir, audioDir, b)
	if err != nil {
		t.Fatal(err)
	}
	if n := len(p.Missing()); n != u5data.BGMSongCount {
		t.Errorf("回報缺 %d 個,預期全部 %d 個", n, u5data.BGMSongCount)
	}
	p.Update(7)
	if len(b.played) != 0 {
		t.Errorf("沒有檔案卻播了 %v", b.played)
	}
	if b.stops != 1 {
		t.Errorf("缺檔時該把舊的停掉(Stop 呼叫 %d 次)", b.stops)
	}
	// ★ 曲號要記住 —— 否則每一帧都會重試,而且換回有檔的曲子時會被當成沒換過。
	if p.Song() != 7 {
		t.Errorf("缺檔之後曲號是 %d,預期記住 7", p.Song())
	}
}

// TestSameSongIsNotRestarted —— 重複同一首不重播。
//
// 原版 `sub_3181C` 每次呼叫都會重播,但它的呼叫點本來就不在每回合的路徑上;
// 引擎是每一帧同步一次 ⇒ 不擋掉的話音樂會一直從頭開始。
func TestSameSongIsNotRestarted(t *testing.T) {
	dir := tblDir(t)
	audioDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(audioDir, "M8"+Ext), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	b := &fakeBackend{}
	p, err := New(dir, audioDir, b)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		p.Update(7) // 曲號 7 = M8.EUP
	}
	if len(b.played) != 1 {
		t.Errorf("同一首播了 %d 次:%v", len(b.played), b.played)
	}
	if len(b.played) == 1 && filepath.Base(b.played[0]) != "M8"+Ext {
		t.Errorf("播的是 %q,預期 M8%s", b.played[0], Ext)
	}
	// 換一首再換回來 → 該重播。
	p.Update(1)
	p.Update(7)
	if len(b.played) != 2 {
		t.Errorf("換走再換回來之後播放次數是 %d,預期 2", len(b.played))
	}
}

// TestSongNoneDoesNothing —— 還沒選曲(`game.SongNone` = −1)時不出聲。
func TestSongNoneDoesNothing(t *testing.T) {
	dir := tblDir(t)
	b := &fakeBackend{}
	p, err := New(dir, t.TempDir(), b)
	if err != nil {
		t.Fatal(err)
	}
	p.Update(-1)
	if len(b.played) != 0 || b.stops != 0 {
		t.Errorf("未選曲卻動了:played=%v stops=%d", b.played, b.stops)
	}
}

// TestNilBackendIsSilentButTracksTheSong —— headless:沒有後端也要記曲號。
func TestNilBackendIsSilentButTracksTheSong(t *testing.T) {
	dir := tblDir(t)
	p, err := New(dir, t.TempDir(), nil)
	if err != nil {
		t.Fatal(err)
	}
	p.Update(9)
	if p.Song() != 9 {
		t.Errorf("沒有後端時曲號是 %d,預期 9", p.Song())
	}
}

// TestMissingTableIsAnError —— ★ 讀不到 `U5_BGM.TBL` 是錯誤,不是「沒有音樂」。
//
// 那表示玩家給的原版資料不完整,要說出來;而「表在但 ogg 還沒渲染」才是降級。
func TestMissingTableIsAnError(t *testing.T) {
	if _, err := New(t.TempDir(), t.TempDir(), nil); err == nil {
		t.Error("讀不到 U5_BGM.TBL 卻沒回錯")
	}
}
