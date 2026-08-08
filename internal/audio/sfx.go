package audio

import (
	"os"
	"path/filepath"
	"strings"

	eaudio "github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/wicanr2/u5-cht/internal/u5data"
)

// 音效(FM Towns 的 25 個 `.SND`,表裡列 23 筆)
//
// 格式見 `docs/re/63`、索引語意見 `docs/re/90` 與 `u5data/sfxindex.go`。
//
// ⚠⚠ **取樣率是推測的**(`u5data.SndAssumedRate` = 7,926,來自檔頭 +0x08 的常數,
// 25 個檔案全部相同)。原版**不傳取樣率**給驅動 —— 它傳的是**音高**
// (`sub_34834(通道, 音高, 音量, 緩衝區)`,`docs/re/63`),而音高 → 實際播放率
// 的換算在音源驅動裡(`docs/re/89`:那在 `TBIOS.BIN`)。
// ⇒ 率不對的話音效會整體偏高或偏低,但**不會錯得無法辨認**。名字裡的
// Assumed 是刻意的,不要把它當定案。

// SFXPlayer 播放音效。可為 nil(headless / 沒有音訊裝置)。
type SFXPlayer struct {
	ctx *eaudio.Context
	// sounds[索引] = 已轉成 16-bit LE 立體聲的位元組(ebiten 要這個格式)。
	sounds [u5data.SFXCount][]byte
	// missing 是表裡有、目錄裡讀不到的檔案。
	missing []string
	// live 是還在播的 player。⚠ 要留著參照,否則 GC 掉就沒聲音了。
	live []*eaudio.Player
	vol  float64
}

// NewSFXPlayer 從 FM Towns 的 U5_E 目錄讀 `U5_SE.TBL` 與 23 個 `.SND`。
//
// ctx 傳 nil 表示不出聲(仍然會回報缺件,所以 headless 也能檢查資料完整)。
func NewSFXPlayer(dir string, ctx *eaudio.Context) (*SFXPlayer, error) {
	table, err := u5data.LoadSoundTable(dir)
	if err != nil {
		return nil, err
	}
	p := &SFXPlayer{ctx: ctx, vol: 1}
	for i, e := range table {
		if i >= u5data.SFXCount {
			break
		}
		s, err := loadSoundCaseInsensitive(dir, e.File)
		if err != nil {
			p.missing = append(p.missing, e.File)
			continue
		}
		s.Volume = e.Volume
		p.sounds[i] = resampleToEbiten(s)
	}
	return p, nil
}

// Missing 回報讀不到的音效檔。
func (p *SFXPlayer) Missing() []string { return p.missing }

// Play 播第 idx 號音效。索引越界或沒有資料就什麼都不做。
//
// ⚠ **不做通道管理。** 原版有 8 個 PCM 通道、`通道 = 索引 & 7`(`docs/re/63`)
// —— 也就是同索引的音效會互相打斷、不同索引可以疊。這裡讓 ebiten 自己混音,
// ⬜ 差異:原版連放同一個音效會截斷前一個,這裡會疊起來。
func (p *SFXPlayer) Play(idx int) {
	if p == nil || p.ctx == nil || idx < 0 || idx >= u5data.SFXCount || p.sounds[idx] == nil {
		return
	}
	// 收掉播完的,避免 live 無限長大。
	kept := p.live[:0]
	for _, pl := range p.live {
		if pl.IsPlaying() {
			kept = append(kept, pl)
		} else {
			_ = pl.Close()
		}
	}
	p.live = kept

	pl := p.ctx.NewPlayerFromBytes(p.sounds[idx])
	pl.SetVolume(p.vol)
	pl.Play()
	p.live = append(p.live, pl)
}

// SetVolume 設 0.0..1.0。
func (p *SFXPlayer) SetVolume(v float64) {
	if p == nil {
		return
	}
	p.vol = v
}

// loadSoundCaseInsensitive 找檔案時不分大小寫。
//
// ⚠ 表裡寫 `Fire.SND` 而 ISO 上是 `FIRE.SND`(`docs/re/63` 記過)。
func loadSoundCaseInsensitive(dir, name string) (*u5data.Sound, error) {
	if s, err := u5data.LoadSound(filepath.Join(dir, name)); err == nil {
		return s, nil
	}
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, e := range ents {
		if strings.EqualFold(e.Name(), name) {
			return u5data.LoadSound(filepath.Join(dir, e.Name()))
		}
	}
	return nil, os.ErrNotExist
}

// resampleToEbiten 把 8-bit 帶號單聲道轉成 ebiten 要的 16-bit LE 立體聲 44.1 kHz。
//
// ⚠ 用**最近鄰**重取樣(整數步進的線性插值)。音效很短(0.07..8 秒)而且是
// 1992 年的 8-bit 素材,更講究的濾波聽不出差別;而**率本身是推測的**
// (見檔頭那一段),所以先把不確定性留在率上,不要在插值上加工。
func resampleToEbiten(s *u5data.Sound) []byte {
	if len(s.PCM) == 0 {
		return nil
	}
	// 表裡的音量 0..127 → 線性增益。
	g := float64(s.Volume) / 127.0
	if s.Volume <= 0 {
		g = 1
	}
	n := len(s.PCM) * SampleRate / u5data.SndAssumedRate
	out := make([]byte, n*4) // 每個 frame = 2 聲道 × 2 byte
	for i := 0; i < n; i++ {
		src := i * u5data.SndAssumedRate / SampleRate
		if src >= len(s.PCM) {
			src = len(s.PCM) - 1
		}
		v := int16(float64(s.PCM[src]) * 256 * g)
		lo, hi := byte(uint16(v)&0xFF), byte(uint16(v)>>8)
		out[i*4+0], out[i*4+1] = lo, hi
		out[i*4+2], out[i*4+3] = lo, hi
	}
	return out
}
