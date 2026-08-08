package u5data

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FM Towns 的 EUPHONY 樂曲 `.EUP` 與 FM 音色庫 `FM_BANK.FMB`
//
// 推導與佐證見 `docs/formats/12-eup-and-fmb.md`;整條音訊鏈路見 `docs/audio-pipeline.md`。
//
// ⚠⚠ **這一份的第一版把事件模型讀錯了。** 當時以為「每 6 byte 一筆事件、
// `0x9n` 是 note-on、`0x8n` 是 note-off」,而真相是**一個音符佔 12 byte**:
// `0x9n` 是前半、`0x8n` 是後半。一手資料的證據是**配對率 100%**
// (882/882 的 `0x9n` 後面緊接 `0x8n`,882/882 的 `0x8n` 前面緊接 `0x9n`)。
// 詳見 `docs/formats/12` §2.1。

const (
	// EUPSignature 是事件區前面的簽章。
	EUPSignature = "EUPHONY "
	// EUPRecordSize 是一筆記錄的位元組數。音符事件佔**兩筆**。
	EUPRecordSize = 6
	// eupHeaderAfterSig 是簽章之後、事件開始之前的位元組數(8 byte 小檔頭)。
	eupHeaderAfterSig = 16

	// EUPTempoOffset 是速度欄相對於簽章的位移。
	//
	// ★ 十五首的值是 88/90/59/67 —— 都是合理的 BPM,而且**有意義地變化**:
	// 曲 4(`M5.EUP`)是 59(最慢、也最長,73 小節 297 秒)、曲 8 是 67。
	// 若它不是速度,不會剛好落在音樂 BPM 的範圍又與曲子性格相符。
	EUPTempoOffset = 15
	// eupUnknown14 是速度欄前面那一格:十四首是 51,只有 `M92.EUP` 是 112。
	// ⬜ 語意未定(疑為拍號 / 每小節拍數)。留著給下一輪。
	eupUnknown14 = 14
)

// EUP 記錄的 status 高 nibble。
const (
	// EUPNote 是音符的**前半**(後半是 `EUPNoteTail`)。
	EUPNote = 0x9
	// EUPNoteTail 是音符的後半。⚠ **不是 note-off** —— 見本檔開頭。
	EUPNoteTail = 0x8
	// EUPProgram 是選音色(MIDI 的 program change)。`Data1` = `FM_BANK.FMB` 的索引。
	EUPProgram = 0xC
	// EUPBarEnd 是小節結束(`0xF2`)。
	//
	// ★ 它的時間欄固定是 **384** = 96 ppqn × 4 拍 ⇒ 4/4 一小節。
	// 而所有音符的小節內 tick 都 < 384(實測最大 381)—— 兩者互相佐證。
	EUPBarEnd = 0xF2
	// EUPTicksPerBar 是一小節的 tick 數。
	EUPTicksPerBar = 384
	// EUPTicksPerQuarter 是四分音符的 tick 數(`eupplayer` 的 `96 * tempo`)。
	EUPTicksPerQuarter = 96
)

// EUPNoteEvent 是一個音符(原始的 12 byte 已經拆好)。
type EUPNoteEvent struct {
	// Channel 是 FM 聲道 0..5。
	//
	// ★ 前半記錄的 `[1]` **100% 等於 status 的低 nibble**(882/882)⇒ 兩者都是聲道。
	// 後半記錄的 `[1]` 只有 6/882 相等 ⇒ 那一格是別的東西(疑為音長)。
	Channel int
	// Bar 是第幾小節(從 0 起),Tick 是小節內的位置 0..383。
	Bar  int
	Tick int
	// Note 是音高、Velocity 是力度(絕大多數 0x40)。
	Note     byte
	Velocity byte
	// Tail 是後半那 6 個位元組,原樣保留。
	//
	// ⬜ 欄位語意未定。已知:`[0]` 高 nibble 固定 8、低 nibble 是同一個聲道;
	// `[1]` 值域廣(12/8/5/13/4/0/10/11…),疑為**音長**;`[5]` 幾乎都是 0x40。
	// 要定案得追 `TBIOS.BIN`(`docs/re/89`)或比對 `eupplayer` 的音長處理。
	Tail [EUPRecordSize]byte
}

// AbsTick 是從曲子開頭算起的 tick。
func (e EUPNoteEvent) AbsTick() int { return e.Bar*EUPTicksPerBar + e.Tick }

// EUPProgramEvent 是一次選音色。
type EUPProgramEvent struct {
	Channel int
	Voice   int // `FM_BANK.FMB` 的索引 0..127
}

// EUPSong 是一首解出來的曲子。
type EUPSong struct {
	// Title 是檔頭 0x20 起的曲名("M1"、"M8"…)。
	Title string
	// Tempo 是速度欄的值(BPM)。
	Tempo int
	// Bars 是小節數(`0xF2` 的個數)。
	Bars int
	// Notes 照時間順序。
	Notes []EUPNoteEvent
	// Programs 照出現順序 —— 通常在最前面,六個聲道各一次。
	Programs []EUPProgramEvent
	// Unknown14 是速度欄前面那一格(十四首 51、`M92` 112)。⬜ 語意未定。
	Unknown14 byte
}

// Duration 是曲長(秒)。
//
//	每 tick 的秒數 = 60 / (96 × 速度)          ← `eupplayer` 的 `60e6/(96*t)` µs
//	曲長 = 小節數 × 384 × 每 tick 的秒數
//
// ⇒ 實測十五首是 11..297 秒,合計 23.3 分鐘 —— 對遊戲配樂是合理的量。
func (s *EUPSong) Duration() float64 {
	if s.Tempo <= 0 {
		return 0
	}
	return float64(s.Bars*EUPTicksPerBar) * 60 / float64(EUPTicksPerQuarter*s.Tempo)
}

// TickSeconds 是一個 tick 幾秒。
func (s *EUPSong) TickSeconds() float64 {
	if s.Tempo <= 0 {
		return 0
	}
	return 60 / float64(EUPTicksPerQuarter*s.Tempo)
}

// ParseEUP 解一首 `.EUP`。
func ParseEUP(raw []byte) (*EUPSong, error) {
	sig := strings.Index(string(raw), EUPSignature)
	if sig < 0 {
		return nil, fmt.Errorf("找不到 %q 簽章", EUPSignature)
	}
	if sig+eupHeaderAfterSig > len(raw) {
		return nil, fmt.Errorf("簽章之後不足 %d byte", eupHeaderAfterSig)
	}
	s := &EUPSong{
		Tempo:     int(raw[sig+EUPTempoOffset]),
		Unknown14: raw[sig+eupUnknown14],
	}
	if len(raw) > 0x28 {
		s.Title = strings.TrimRight(string(raw[0x20:0x28]), "\x00 ")
	}

	at := func(i int) []byte { return raw[i : i+EUPRecordSize] }
	i := sig + eupHeaderAfterSig
	for i+EUPRecordSize <= len(raw) {
		r := at(i)
		if r[0] == 0xFF {
			break
		}
		switch r[0] >> 4 {
		case EUPNote:
			// ★ 音符佔兩筆 —— 後半一定緊跟在後面。
			if i+2*EUPRecordSize > len(raw) {
				return nil, fmt.Errorf("位移 0x%X 的音符缺後半", i)
			}
			tail := at(i + EUPRecordSize)
			if tail[0]>>4 != EUPNoteTail {
				return nil, fmt.Errorf("位移 0x%X 的音符後半是 0x%02X,預期 0x8n", i, tail[0])
			}
			n := EUPNoteEvent{
				Channel:  int(r[0] & 0x0F),
				Bar:      s.Bars,
				Tick:     int(r[2]) + 0x80*int(r[3]),
				Note:     r[4],
				Velocity: r[5],
			}
			copy(n.Tail[:], tail)
			s.Notes = append(s.Notes, n)
			i += 2 * EUPRecordSize

		case EUPProgram:
			s.Programs = append(s.Programs, EUPProgramEvent{
				Channel: int(r[0] & 0x0F), Voice: int(r[4]),
			})
			i += EUPRecordSize

		default:
			if r[0] == EUPBarEnd {
				s.Bars++
			}
			i += EUPRecordSize
		}
	}
	if len(s.Notes) == 0 {
		return nil, fmt.Errorf("解不出任何音符")
	}
	return s, nil
}

// LoadEUP 從目錄讀一首(檔名取自 `U5_BGM.TBL`)。
func LoadEUP(dir, name string) (*EUPSong, error) {
	raw, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		return nil, err
	}
	song, err := ParseEUP(raw)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", name, err)
	}
	return song, nil
}

// VoiceOf 回報某個聲道用的音色索引,沒選過就回 −1。
func (s *EUPSong) VoiceOf(ch int) int {
	v := -1
	for _, p := range s.Programs {
		if p.Channel == ch {
			v = p.Voice
		}
	}
	return v
}

// FM 音色庫 `FM_BANK.FMB`

const (
	// FMBankHeader 是檔案開頭那 8 byte。
	FMBankHeader = 8
	// FMVoiceSize 是一筆音色的位元組數。
	FMVoiceSize = 48
	// FMVoiceNameSize 是音色名的位元組數。
	FMVoiceNameSize = 8
	// FMVoiceCount 是音色數。
	//
	// ★ 8 + 128 × 48 = 6152 = 檔案大小,**整除且無餘數**;外部資料也記載
	// 「FMB 剛好 6,152 byte」;而 EUP 的 program change 值域落在 0..127
	// ⇒ **三個獨立來源一致**。
	FMVoiceCount = 128
	// FMVoiceChannels 是 EUP 用得到的 FM 聲道數(對上 `U5_BGM.TBL` 的六欄音量)。
	FMVoiceChannels = 6
	// FMOperators 是每個音色的運算子數(YM2612 是 4-op)。
	FMOperators = 4
)

// FMOperator 是一個運算子的參數(已從打包位元組拆開)。
type FMOperator struct {
	Detune   byte // DT 0..7
	Multiple byte // MUL 0..15
	TotalLvl byte // TL 0..127(越大越輕)
	KeyScale byte // KS 0..3
	Attack   byte // AR 0..31
	Decay    byte // DR 0..31
	Sustain  byte // SR 0..31
	SusLevel byte // SL 0..15
	Release  byte // RR 0..15
}

// FMVoice 是一筆 FM 音色。
type FMVoice struct {
	Name string
	Op   [FMOperators]FMOperator
	// Algorithm 是 0..7、Feedback 是 0..7(同一個位元組 +32)。
	Algorithm byte
	Feedback  byte
	// Pan 是 +33:高兩位是 L/R 輸出、低六位是 AMS/PMS。
	//
	// ★ 實測 126 個有名字的音色**全部 ≥ 192** ⇒ L 與 R 兩個輸出位元永遠都開。
	Pan byte
	// Raw 是名字後面那 40 byte,原樣保留(給比對與除錯用)。
	Raw [FMVoiceSize - FMVoiceNameSize]byte
}

// ParseFMBank 解 `FM_BANK.FMB`。
//
// # 佈局(相對於一筆記錄的開頭,名字佔 0..7)
//
//	+8 +9 +10 +11   DT/MUL    運算子 0..3   (DT = 高 4 bit & 7、MUL = 低 4 bit)
//	+12..+15        TL        0..127
//	+16..+19        KS/AR     (KS = 位元 6-7、AR = 低 5 bit)
//	+20..+23        DR        低 5 bit
//	+24..+27        SR        低 5 bit
//	+28..+31        SL/RR     (SL = 高 4 bit、RR = 低 4 bit)
//	+32             FB/ALG    (FB = 位元 3-5、ALG = 低 3 bit)
//	+33             L/R + AMS/PMS
//	+34..+47        全 0(14 byte)
//
// ★★ **這個佈局有兩個獨立來源**:
//
//  1. **一手否證**:拿 YM2612 各欄位的位元寬度去掃全部 126 個有名字的音色,
//     每一組的實測值域都剛好塞得進對應欄位,而且**互換就爆掉** ——
//     `KS/AR` 上限 223、`SL/RR` 上限 255 都放不進 `TL`(≤127)或 `SR`(≤31)。
//     最決定性的兩個:`+32` 實測 0..**61** 而 FB/ALG 的寬度剛好是 0..63;
//     `+33` 全部 ≥ 192 ⇒ L/R 位元永遠都開,正是音樂音色庫該有的樣子。
//  2. **外部來源**:`gzaffin/eupmini`(EUPHONY 播放器)的
//     `TownsFmEmulator_Operator::setInstrumentParameter` 逐行讀 `instrument[8]`
//     `[12]` `[16]` `[20]` `[24]` `[28]`,運算子以 `instrument + n`(n = 0..3)
//     取偏移 ⇒ **與 (1) 完全吻合**。
//
// ⚠ 這仍然不是「原版 BIOS 的一手證據」(`docs/re/89`:解析在 `TBIOS.BIN` 裡),
// 但兩個獨立來源互相確認,比單靠社群文件強得多。有疑慮時回頭讀 `Raw`。
func ParseFMBank(raw []byte) ([]FMVoice, error) {
	want := FMBankHeader + FMVoiceCount*FMVoiceSize
	if len(raw) != want {
		return nil, fmt.Errorf("大小 %d,預期 %d(8 + %d × %d)",
			len(raw), want, FMVoiceCount, FMVoiceSize)
	}
	out := make([]FMVoice, FMVoiceCount)
	for i := range out {
		r := raw[FMBankHeader+i*FMVoiceSize:][:FMVoiceSize]
		v := &out[i]
		v.Name = strings.TrimRight(string(r[:FMVoiceNameSize]), "\x00 ")
		copy(v.Raw[:], r[FMVoiceNameSize:])
		for n := 0; n < FMOperators; n++ {
			v.Op[n] = FMOperator{
				Detune:   (r[8+n] >> 4) & 7,
				Multiple: r[8+n] & 0x0F,
				TotalLvl: r[12+n] & 0x7F,
				KeyScale: (r[16+n] >> 6) & 3,
				Attack:   r[16+n] & 0x1F,
				Decay:    r[20+n] & 0x1F,
				Sustain:  r[24+n] & 0x1F,
				SusLevel: (r[28+n] >> 4) & 0x0F,
				Release:  r[28+n] & 0x0F,
			}
		}
		v.Algorithm = r[32] & 7
		v.Feedback = (r[32] >> 3) & 7
		v.Pan = r[33]
	}
	return out, nil
}

// LoadFMBank 從 FM Towns 的 U5_E 目錄讀 `FM_BANK.FMB`。
func LoadFMBank(dir string) ([]FMVoice, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "FM_BANK.FMB"))
	if err != nil {
		return nil, err
	}
	return ParseFMBank(raw)
}
