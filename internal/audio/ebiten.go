package audio

import (
	"fmt"
	"os"

	eaudio "github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
)

// SampleRate 是音訊環境的取樣率。
//
// 渲染出來的 ogg 是 44,100 Hz(`tools/eup2ogg.py` 的預設,也是 CDDA 的原生率)
// ⇒ 用同一個率,ebiten 就不必重新取樣。
const SampleRate = 44100

// EbitenBackend 用 ebiten 的音訊播 ogg。
//
// ⚠ 這一支是 `internal/audio` 裡**唯一**碰 ebiten 的地方 —— `Player` 與
// 「該播第幾首」的邏輯都不知道 ebiten 存在,所以 headless 與單元測試不需要
// 音效裝置(同 `render` 不綁 GPU,`docs/engineering-notes.md`)。
type EbitenBackend struct {
	ctx *eaudio.Context
	cur string
	p   *eaudio.Player
	vol float64
}

// NewEbitenBackend 建一個 ebiten 音訊後端。
//
// ⚠ `eaudio.NewContext` 一個程式只能呼叫一次;重複呼叫 ebiten 會 panic。
// 所以這一支也只該被 `cmd/u5cht` 叫一次。
func NewEbitenBackend() *EbitenBackend {
	return &EbitenBackend{ctx: eaudio.NewContext(SampleRate), vol: 1}
}

// Context 回傳底層的 ebiten 音訊環境,給音效共用。
//
// ⚠ `eaudio.NewContext` 一個程式只能叫一次 ⇒ 配樂與音效必須共用同一個。
func (b *EbitenBackend) Context() *eaudio.Context { return b.ctx }

// Play 循環播放 path。已經在播同一個檔案就什麼都不做。
//
// ★ **循環**是刻意的:原版的配樂會一直放到換場景為止(`sub_3181C` 只在換曲時
// 被呼叫,`docs/re/87`)。而十五首的長度是 11..297 秒 ⇒ 不循環的話短曲會沉默。
func (b *EbitenBackend) Play(path string) error {
	if path == b.cur && b.p != nil && b.p.IsPlaying() {
		return nil
	}
	b.Stop()
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	// ⚠ 不能 `defer f.Close()` —— 串流是**惰性**解碼的,關掉檔案之後就讀不到了。
	// 播放器被 Stop / 換曲時連同串流一起丟掉,檔案由 GC 收。
	s, err := vorbis.DecodeF32(f)
	if err != nil {
		f.Close()
		return fmt.Errorf("%s:%w", path, err)
	}
	p, err := b.ctx.NewPlayerF32(eaudio.NewInfiniteLoopF32(s, s.Length()))
	if err != nil {
		f.Close()
		return fmt.Errorf("%s:%w", path, err)
	}
	p.SetVolume(b.vol)
	p.Play()
	b.cur, b.p = path, p
	return nil
}

// Stop 停掉現在在播的。
func (b *EbitenBackend) Stop() {
	if b.p != nil {
		_ = b.p.Close()
		b.p = nil
	}
	b.cur = ""
}

// SetVolume 設 0.0..1.0。
func (b *EbitenBackend) SetVolume(v float64) {
	switch {
	case v < 0:
		v = 0
	case v > 1:
		v = 1
	}
	b.vol = v
	if b.p != nil {
		b.p.SetVolume(v)
	}
}
