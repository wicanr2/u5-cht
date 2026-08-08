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

// Player 把曲號對到檔案並交給 Backend。
type Player struct {
	// Dir 是渲染好的音訊所在目錄。
	Dir string
	// Backend 為 nil 時整個 Player 是安靜的(headless / 沒有音效裝置)。
	Backend Backend

	// files[曲號] = 檔名(不含目錄),空字串 = 那一首沒有可播的檔案。
	files []string
	// song 是上一次交給 Backend 的曲號;`u5data.NoStartupSong` 代表還沒播過。
	song int
	// missing 收集「表裡有、目錄裡沒有」的曲子 —— 給啟動時誠實回報用。
	missing []string
}

// New 依原版的 `U5_BGM.TBL` 建一個 Player。
//
//	tblDir 是 FM Towns 的 `U5_E` 目錄(要讀 `U5_BGM.TBL`)
//	audioDir 是渲染好的 ogg 所在目錄
//
// ⚠ **讀不到 `U5_BGM.TBL` 是錯誤,不是「沒有音樂」** —— 那表示玩家給的資料
// 不完整,要說出來。而「表讀到了但 ogg 還沒渲染」是可以接受的降級
// (`CLAUDE.md §3.0`:缺素材要優雅降級並明說)。
func New(tblDir, audioDir string, b Backend) (*Player, error) {
	tracks, err := u5data.LoadBGMTable(tblDir)
	if err != nil {
		return nil, fmt.Errorf("讀 U5_BGM.TBL:%w", err)
	}
	p := &Player{
		Dir:     audioDir,
		Backend: b,
		files:   make([]string, len(tracks)),
		song:    u5data.NoStartupSong,
	}
	for i, tr := range tracks {
		name := strings.TrimSuffix(tr.File, filepath.Ext(tr.File)) + Ext
		if _, err := os.Stat(filepath.Join(audioDir, name)); err != nil {
			p.missing = append(p.missing, name)
			continue
		}
		p.files[i] = name
	}
	return p, nil
}

// Missing 回報「表裡有、目錄裡沒有」的檔案。
//
// ★ 這個數字是**完整度指標**,不是雜訊:15 個全缺表示還沒跑過轉檔,
// 缺幾個表示轉檔漏了。啟動時該把它印出來,不要靜默(同 `cjk` 的 fallback 計數)。
func (p *Player) Missing() []string { return p.missing }

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
	_ = p.Backend.Play(filepath.Join(p.Dir, name))
}

// FileFor 回報某個曲號對應的檔名(不含目錄),沒有就回空字串。
func (p *Player) FileFor(song int) string {
	if song < 0 || song >= len(p.files) {
		return ""
	}
	return p.files[song]
}

// Song 回報上一次同步過去的曲號。
func (p *Player) Song() int { return p.song }
