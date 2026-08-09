// Package audio 把「該播第幾首」變成實際的聲音。
//
// 分工(與 `render` 不綁 GPU 同一個理由,`docs/engineering-notes.md`):
//
//	internal/game   決定曲號(`State.CurrentSong()`,原版 `dword_65334`)
//	internal/audio  拿曲號去找檔案並播放
//
// ⇒ **headless 不需要音效裝置**:`Player` 在沒有後端時是一個安靜的空實作,
// 而「該換哪一首」的邏輯照樣可以在單元測試裡驗。
//
// ⚠ 曲號 → 檔名來自原版的 `U5_BGM.TBL`(`u5data.LoadBGMTable`),
// 而**檔名編號不連續**(`M92` / `M152` 夾在中間)⇒ 不能用「曲號 = 數字 − 1」
// (`docs/re/87` §2)。
package audio

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// Ext 是轉檔之後的音訊格式。
//
// ⚠ `.EUP` 是 FM 音源的序列,不是取樣音訊 —— 要**離線**渲染成 ogg
// (`CLAUDE.md §3`:不在遊戲內做 FM 合成)。轉出來的 ogg 衍生自原版資料,
// **不入 git**。
const Ext = ".ogg"

// Source 是音樂的來源版本。同一個曲號在不同來源是**同一首曲子的不同編曲**。
//
// 兩套都是原版素材(`CLAUDE.md §3.0`:不自製、不重繪),差在音源硬體:
//
//	SourceFMTowns  FM Towns 的 `M*.EUP`     → YM2612 六聲道 FM(`tools/eup2ogg.py`)
//	SourceMT32     DOS upgrade 的 `*.XMI`   → Roland MT-32(`tools/mt32_render.sh`)
type Source int

// 音樂來源。⚠ 順序就是 `NextSource()` 的循環順序。
const (
	// SourceFMTowns 是**遊戲內實際會響的那一套**(但我們的 ogg 是自己合成的)
	// (FM Towns 版的遊玩音樂是 `.EUP`,不是 CDDA)。
	SourceFMTowns Source = iota
	// SourceMT32 是 DOS 玩家 1988 年裝了 MT-32 會聽到的版本。
	SourceMT32
	// SourceCount 是來源數量(迴圈用)。
	SourceCount
)

// SourceNames 是給玩家看的名稱。下標是 Source。
var SourceNames = [SourceCount]string{"FM Towns", "MT-32"}

// String 讓 Source 可以直接印。
func (s Source) String() string {
	if s < 0 || s >= SourceCount {
		return "?"
	}
	return SourceNames[s]
}

// SourceFlags 是 `-music` 旗標收的值。
var SourceFlags = map[string]Source{
	"fmtowns": SourceFMTowns,
	"fm":      SourceFMTowns,
	"eup":     SourceFMTowns,
	"mt32":    SourceMT32,
	"midi":    SourceMT32,
	"xmi":     SourceMT32,
}

// ParseSource 解 `-music` 的值。空字串代表「不指定」(回 −1)。
func ParseSource(v string) (Source, error) {
	if v == "" {
		return -1, nil
	}
	if s, ok := SourceFlags[strings.ToLower(v)]; ok {
		return s, nil
	}
	return -1, fmt.Errorf("不認得的音樂來源 %q(可用:fmtowns / mt32)", v)
}

// MT32Subdir 是 MT-32 那套 ogg 相對於音訊目錄的子目錄(`tools/mt32_render.sh` 的輸出)。
const MT32Subdir = "mt32"

// Backend 是實際發聲的東西。抽出來是為了讓 headless 用空實作。
//
// ⚠ 介面刻意只有三個動作:換曲是「停掉舊的、播新的」,原版的六聲道淡出
// (`sub_3181C`)在這一層無從對應 —— 我們播的是渲染好的整首,不是 FM 聲道。
type Backend interface {
	// Play 開始循環播放 path 指的檔案。已經在播同一個檔案時要當成 no-op。
	Play(path string) error
	// Stop 停掉現在在播的。
	Stop()
	// SetVolume 設 0.0..1.0。
	SetVolume(v float64)
}

// sourceSet 是一套音樂(一個來源)。
type sourceSet struct {
	// dir 是這一套的 ogg 所在目錄。
	dir string
	// files[曲號] = 檔名(不含目錄),空字串 = 那一首沒有可播的檔案。
	files []string
	// missing 收集「該有、目錄裡沒有」的曲子 —— 給啟動時誠實回報用。
	missing []string
	// err 是這一套為什麼整套不能用(nil = 可用)。
	err error
}

// have 回報這一套至少有一個檔案。
func (s *sourceSet) have() bool {
	for _, f := range s.files {
		if f != "" {
			return true
		}
	}
	return false
}

// Player 把曲號對到檔案並交給 Backend。
type Player struct {
	// Dir 是渲染好的音訊所在目錄(FM Towns 那套;MT-32 在它的 `mt32/` 子目錄)。
	Dir string
	// Backend 為 nil 時整個 Player 是安靜的(headless / 沒有音效裝置)。
	Backend Backend

	// sets 依 Source 索引。
	sets [SourceCount]sourceSet
	// source 是現在用哪一套。
	source Source
	// song 是上一次交給 Backend 的曲號;`u5data.NoStartupSong` 代表還沒播過。
	song int
}

// New 建一個 Player,兩套音樂都掛上。
//
//	tblDir 是 FM Towns 的 `U5_E` 目錄(要讀 `U5_BGM.TBL` 拿 `.EUP` 的檔名)
//	audioDir 是渲染好的 ogg 所在目錄;MT-32 那套在它的 `mt32/` 子目錄
//
// ⚠ **兩套都沒有可播的檔案才算錯誤。** 只缺一套是降級不是失敗:
// 沒有 FM Towns 光碟的 DOS 玩家(這是最普遍的情況)只會有 MT-32 那套,
// 而他照樣該有音樂。「表讀到了但 ogg 還沒渲染」同樣是降級
// (`CLAUDE.md §3.0`:缺素材要優雅降級並明說)。
func New(tblDir, audioDir string, b Backend) (*Player, error) {
	p := &Player{Dir: audioDir, Backend: b, song: u5data.NoStartupSong}
	p.sets[SourceFMTowns] = fmTownsSet(tblDir, audioDir)
	p.sets[SourceMT32] = mt32Set(filepath.Join(audioDir, MT32Subdir))

	// 預設順序:**MT-32 優先,FM Towns 次之**,挑第一個有檔案的。
	//
	// ⚠ 這與「哪一套最原生」是**兩個不同的問題**。FM Towns 那套才是
	// 遊戲內實際會響的,但我們手上的 `M*.ogg` 是**自己合成**的 ——
	// `.EUP` 序列 + `.FMB` 4-op 音色餵進 `tools/eup2ogg.py` 的 FM 合成器。
	// MT-32 那套是把 upgrade 的 `.XMI` 餵給 **munt + 真的 MT-32 ROM** 渲染,
	// 也就是 1988 年裝了 MT-32 的玩家實際聽到的聲音。
	//
	// ⇒ 使用者 2026-08-09 實機回報「背景音樂怪怪的」,而當時播的是 FM Towns 那套。
	// 在自製合成器驗過之前,預設用**已知渲染正確**的那一套;
	// 想聽 FM Towns 仍可按 F9 或 `-music fmtowns`。
	//
	// ⬜ 要把 FM Towns 那套修對,做法與調色盤那條一樣:**拿 reference 對答案** ——
	// DOSBox-X 支援 `machine=fmtowns`,用玩家自備的那張光碟開起來錄音,
	// 才知道我們的合成器差在哪(音色、包絡、速度)。不從網路抓別人的 rip:
	// 那既是散布他人重製的音檔,也繞過了「素材一律來自原版媒體」這條
	// (`CLAUDE.md §3.0`),而且抓回來也**不會告訴我們合成器哪裡錯**。
	for _, s := range []Source{SourceMT32, SourceFMTowns} {
		if p.sets[s].have() {
			p.source = s
			return p, nil
		}
	}
	// 兩套都沒有可播的檔案。這時只有「連 `U5_BGM.TBL` 都讀不到」才算錯誤 ——
	// 那表示玩家給的**原版資料**不完整,要說出來。
	//
	// ⚠ MT-32 目錄不存在**不是**錯誤:那一套要玩家自己跑 `tools/mt32_render.sh`
	// (需要 MT-32 ROM),沒跑過是常態。拿它當錯誤會讓「還沒轉檔的人」開不了音樂。
	if err := p.sets[SourceFMTowns].err; err != nil {
		return nil, err
	}
	return p, nil
}

// fmTownsSet 依 `U5_BGM.TBL` 建 FM Towns 那一套。
func fmTownsSet(tblDir, dir string) sourceSet {
	set := sourceSet{dir: dir}
	tracks, err := u5data.LoadBGMTable(tblDir)
	if err != nil {
		set.err = fmt.Errorf("讀 U5_BGM.TBL:%w", err)
		return set
	}
	set.files = make([]string, len(tracks))
	for i, tr := range tracks {
		name := strings.TrimSuffix(tr.File, filepath.Ext(tr.File)) + Ext
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			set.missing = append(set.missing, name)
			continue
		}
		set.files[i] = name
	}
	return set
}

// mt32Set 依 `u5data.MT32Tracks` 建 MT-32 那一套。
//
// ⚠⚠ **查檔要大小寫無關。** 光碟上 15 個 `.XMI` 是大寫、`trntlla.xmi` 是小寫,
// 而渲染出來的 ogg 沿用原檔名 ⇒ 寫死大小寫會漏掉那一首。
// 這個坑咬過一次(`*.XMI` 的 glob 漏了它,害「15 首」寫成事實,實際 16 首)。
func mt32Set(dir string) sourceSet {
	set := sourceSet{dir: dir, files: make([]string, u5data.BGMSongCount)}
	// 先把目錄裡的檔名建成小寫 → 實際名稱的索引,之後逐首查。
	//
	// ⚠ 讀不到目錄**刻意不當錯誤** —— 那只是「還沒渲染 MT-32 那一套」,
	// 而 `Missing()` 已經會把 15 首全部列出來。
	actual := map[string]string{}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if !e.IsDir() {
			actual[strings.ToLower(e.Name())] = e.Name()
		}
	}
	for i, tr := range u5data.MT32Tracks {
		want := tr.Base + Ext
		if name, ok := actual[strings.ToLower(want)]; ok {
			set.files[i] = name
			continue
		}
		set.missing = append(set.missing, want)
	}
	return set
}

// Missing 回報現在這一套「該有、目錄裡沒有」的檔案。
//
// ★ 這個數字是**完整度指標**,不是雜訊:15 個全缺表示還沒跑過轉檔,
// 缺幾個表示轉檔漏了。啟動時該把它印出來,不要靜默(同 `cjk` 的 fallback 計數)。
func (p *Player) Missing() []string { return p.sets[p.source].missing }

// Source 回報現在用哪一套音樂。
func (p *Player) Source() Source { return p.source }

// Available 回報某一套有沒有可播的檔案。
func (p *Player) Available(s Source) bool {
	if s < 0 || s >= SourceCount {
		return false
	}
	return p.sets[s].have()
}

// SetSource 換一套音樂,並**把當前這一首用新的音源重播**。
//
// ⚠ 重播是必要的:不重播的話玩家按了切換鍵沒有任何反應
// (曲號沒變 ⇒ `Update` 會當成沒換過),看起來像切換壞了。
//
// 回傳 false = 那一套沒有可播的檔案,**沒有換**。呼叫端該把這件事說出來,
// 而不是靜靜地換過去變成靜音。
func (p *Player) SetSource(s Source) bool {
	if s == p.source {
		return true
	}
	if !p.Available(s) {
		return false
	}
	p.source = s
	song := p.song
	p.song = u5data.NoStartupSong // 讓 Update 認為「換了一首」
	p.Update(song)
	return true
}

// NextSource 循環到下一套有檔案的音樂,回傳換到哪一套。
//
// 只有一套可用時回傳原本那一套(等於什麼都沒做)。
func (p *Player) NextSource() Source {
	for i := 1; i < int(SourceCount); i++ {
		s := (p.source + Source(i)) % SourceCount
		if p.SetSource(s) {
			return s
		}
	}
	return p.source
}

// TrackTitle 回報某個曲號的曲名,沒有就回空字串。
//
// ⚠ 只有 MT-32 那套有曲名 —— upgrade 的作者在 `Files.txt` 裡標了
// (`Ultima V Theme`、`Halls of Doom`…),而 FM Towns 那套的檔名只有 `M1`、`M92`。
// ⇒ 曲名**不分來源**都可以用(同一首曲子),但它的出處是 upgrade。
func (p *Player) TrackTitle(song int) string {
	if song < 0 || song >= len(u5data.MT32Tracks) {
		return ""
	}
	return u5data.MT32Tracks[song].Title
}

// Update 把「現在該播第幾首」同步到 Backend。每一帧呼叫都便宜。
//
//	song < 0(`game.SongNone`)         → 什麼都不做(還沒選曲)
//	song 與上次相同                     → 什麼都不做
//	那一首沒有檔案                      → 記住曲號但不出聲(下次換回來才會播)
//
// ⚠ **記住曲號**這件事很重要:少了它,缺檔的那一首會讓每一帧都重試 `os.Stat`
// 與 `Play`,而且從「缺檔的曲子」換回「有檔的曲子」時會被當成沒換過。
func (p *Player) Update(song int) {
	if song < 0 || song == p.song {
		return
	}
	p.song = song
	if p.Backend == nil {
		return
	}
	name := p.FileFor(song)
	if name == "" {
		p.Backend.Stop()
		return
	}
	// 播不出來就當成缺檔:安靜比崩掉好,而 Missing() 已經回報過缺什麼。
	_ = p.Backend.Play(filepath.Join(p.sets[p.source].dir, name))
}

// FileFor 回報某個曲號在**現在這一套**裡的檔名(不含目錄),沒有就回空字串。
func (p *Player) FileFor(song int) string {
	set := &p.sets[p.source]
	if song < 0 || song >= len(set.files) {
		return ""
	}
	return set.files[song]
}

// Song 回報上一次同步過去的曲號。
func (p *Player) Song() int { return p.song }
