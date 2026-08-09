package audio

import (
	"os"
	"path/filepath"
	"strings"
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

// mt32Dir 在 audioDir 底下建好 MT-32 那一套的假檔案。
func mt32Dir(t *testing.T, audioDir string, tracks ...int) {
	t.Helper()
	dir := filepath.Join(audioDir, MT32Subdir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, i := range tracks {
		name := u5data.MT32Tracks[i].Base + Ext
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// TestSongNumbersMapToBothSets —— ★ 同一個曲號在兩套裡是同一首曲子。
//
// 對應是靠**旋律比對**得來的(`tools/match_songs.py`,推導寫在
// `internal/u5data/mt32.go`)。這條測試釘住幾個交叉驗證過的點:
// 曲 2 是水手舞曲(上船)、曲 3 是戰鬥、曲 9 是地牢、曲 0x0A 是幽冥界 ——
// 那五條與 `docs/re/87` 從呼叫點逆出來的用途獨立吻合。
func TestSongNumbersMapToBothSets(t *testing.T) {
	for _, tc := range []struct {
		song  int
		fm    string
		mt32  string
		title string
	}{
		{1, "M2", "BRITLAND", "Britannic Lands"},
		{2, "M3", "HORNPIPE", "Cap'n Johne's Hornpipe"},
		{3, "M4", "ENGGMNT", "Engagement and Melee"},
		{9, "M10", "HALLS", "Halls of Doom"},
		{10, "M11", "WRLDBLW", "Worlds Below"},
		// ★ 檔名編號不連續的兩首,兩套都要對得上。
		{8, "M92", "trntlla", "Villager Tarantella"},
		{14, "M152", "RULEBRIT", "Rule Britannia"},
	} {
		if got := u5data.MT32Tracks[tc.song].Base; got != tc.mt32 {
			t.Errorf("曲號 %d 的 MT-32 檔名是 %q,預期 %q", tc.song, got, tc.mt32)
		}
		if got := u5data.MT32Tracks[tc.song].Title; got != tc.title {
			t.Errorf("曲號 %d 的曲名是 %q,預期 %q", tc.song, got, tc.title)
		}
	}
	// 15 首全部有名字(空的表示配對漏了一首)。
	for i, tr := range u5data.MT32Tracks {
		if tr.Base == "" || tr.Title == "" {
			t.Errorf("曲號 %d 沒有對應的 MT-32 曲子:%+v", i, tr)
		}
	}
	// AMIGA 沒有曲號 —— 它不該出現在表裡。
	for i, tr := range u5data.MT32Tracks {
		if tr.Base == "AMIGA" {
			t.Errorf("AMIGA 被排進曲號 %d,但它對任何 .EUP 都不像(沒有曲號)", i)
		}
	}
}

// TestMT32LookupIsCaseInsensitive —— ★★ 檔名大小寫不可靠。
//
// 光碟上 15 個 `.XMI` 是大寫、`trntlla.xmi` 是小寫。這個坑咬過一次:
// `*.XMI` 的 glob 漏掉那一首,害「15 首」被寫成事實(實際 16 首)。
func TestMT32LookupIsCaseInsensitive(t *testing.T) {
	dir := tblDir(t)
	audioDir := t.TempDir()
	mdir := filepath.Join(audioDir, MT32Subdir)
	if err := os.MkdirAll(mdir, 0o755); err != nil {
		t.Fatal(err)
	}
	// 故意全部反過來寫:大寫的寫成小寫、小寫的寫成大寫。
	for _, tr := range u5data.MT32Tracks {
		flipped := strings.ToUpper(tr.Base)
		if tr.Base == flipped {
			flipped = strings.ToLower(tr.Base)
		}
		if err := os.WriteFile(filepath.Join(mdir, flipped+Ext), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	p, err := New(dir, audioDir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !p.SetSource(SourceMT32) {
		t.Fatal("大小寫翻過來就找不到 MT-32 那一套")
	}
	if n := len(p.Missing()); n != 0 {
		t.Errorf("大小寫翻過來之後回報缺 %d 首:%v", n, p.Missing())
	}
}

// TestSourceDefaultsToWhicheverIsRendered —— ★ 只渲染了一套時,預設就用那一套。
//
// 硬用 FM Towns 的話,只有 MT-32 的玩家開起來是靜音的**而且看不出原因**。
func TestSourceDefaultsToWhicheverIsRendered(t *testing.T) {
	dir := tblDir(t)
	audioDir := t.TempDir()
	mt32Dir(t, audioDir, 0, 1, 7)
	p, err := New(dir, audioDir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if p.Source() != SourceMT32 {
		t.Errorf("只有 MT-32 渲染好,預設卻是 %s", p.Source())
	}
	if got := p.FileFor(7); got != "MONARCH"+Ext {
		t.Errorf("曲號 7 的檔案是 %q,預期 MONARCH%s", got, Ext)
	}
}

// TestSwitchingSourceRestartsTheCurrentSong —— ★★ 切換要**重播當前這一首**。
//
// 不重播的話玩家按了 F5 什麼事都不會發生(曲號沒變 → `Update` 當成沒換過),
// 看起來像切換壞了。
func TestSwitchingSourceRestartsTheCurrentSong(t *testing.T) {
	dir := tblDir(t)
	audioDir := t.TempDir()
	// FM Towns 的曲號 7 = M8;MT-32 的曲號 7 = MONARCH。
	if err := os.WriteFile(filepath.Join(audioDir, "M8"+Ext), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	mt32Dir(t, audioDir, 7)
	b := &fakeBackend{}
	p, err := New(dir, audioDir, b)
	if err != nil {
		t.Fatal(err)
	}
	p.Update(7)
	if len(b.played) != 1 || filepath.Base(b.played[0]) != "M8"+Ext {
		t.Fatalf("先播的是 %v,預期 M8%s", b.played, Ext)
	}
	if !p.SetSource(SourceMT32) {
		t.Fatal("換不到 MT-32")
	}
	if len(b.played) != 2 {
		t.Fatalf("換來源之後播了 %d 次,預期 2 次(同一首要用新音源重播)", len(b.played))
	}
	if got := filepath.Base(b.played[1]); got != "MONARCH"+Ext {
		t.Errorf("換來源之後播 %q,預期 MONARCH%s", got, Ext)
	}
	// 而且曲號沒變 —— 換的是音源,不是曲子。
	if p.Song() != 7 {
		t.Errorf("換來源之後曲號變成 %d", p.Song())
	}
}

// TestSwitchingToAnUnrenderedSourceIsRefused —— ★ 沒渲染的那一套不能換過去。
//
// 靜靜換過去會變成無聲,而玩家會以為是遊戲壞了。回 false 讓呼叫端說明。
func TestSwitchingToAnUnrenderedSourceIsRefused(t *testing.T) {
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
	p.Update(7)
	if p.SetSource(SourceMT32) {
		t.Error("MT-32 一首都沒渲染卻換過去了")
	}
	if p.Source() != SourceFMTowns {
		t.Errorf("被拒之後來源變成 %s", p.Source())
	}
	// NextSource 在只有一套可用時要回原本那一套(等於什麼都沒做)。
	if got := p.NextSource(); got != SourceFMTowns {
		t.Errorf("NextSource 回 %s,預期原地不動", got)
	}
	if len(b.played) != 1 {
		t.Errorf("被拒的切換卻動了播放:%v", b.played)
	}
}

// TestParseSourceAcceptsTheFlagValues —— `-music` 收的值。
func TestParseSourceAcceptsTheFlagValues(t *testing.T) {
	for v, want := range map[string]Source{
		"": -1, "fmtowns": SourceFMTowns, "FM": SourceFMTowns, "eup": SourceFMTowns,
		"mt32": SourceMT32, "MIDI": SourceMT32, "xmi": SourceMT32,
	} {
		got, err := ParseSource(v)
		if err != nil {
			t.Errorf("%q:%v", v, err)
		}
		if got != want {
			t.Errorf("%q → %v,預期 %v", v, got, want)
		}
	}
	if _, err := ParseSource("sblaster"); err == nil {
		t.Error("不認得的值卻沒回錯")
	}
}
