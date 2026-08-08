package u5data

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FM Towns 的 EUPHONY 樂曲 `.EUP` 與 FM 音色庫 `FM_BANK.FMB`
//
// 推導與佐證見 `docs/formats/12-eup-and-fmb.md`。
//
// ★ 這一份**只解碼,不合成**。FM 參數的語意還沒定案(見 `FMVoice.Params`),
// 所以不猜 —— 解出來的事件與音色先當資料放著,渲染是另一件事。

const (
	// EUPSignature 是事件區前面的簽章。
	EUPSignature = "EUPHONY "
	// EUPEventSize 是一筆事件的位元組數。
	EUPEventSize = 6
	// eupHeaderAfterSig 是簽章之後、事件開始之前的位元組數。
	//
	// ★ 這個 16 是**推出來的**:簽章 8 byte + 8 byte 小檔頭。驗證方式是
	// 「從 sig+16 起以 6 為步長讀,note-on 與 note-off 的筆數會相等」——
	// 相位錯 2 的話 status 欄會落在別的位置,兩者立刻不相等
	// (實測 sig+0x40 那個錯相位:status 值域 0..120、高 nibble 亂七八糟)。
	eupHeaderAfterSig = 16
)

// EUP 事件的 status 高 nibble。
const (
	EUPNoteOff = 0x8
	EUPNoteOn  = 0x9
	// EUPProgram 是選音色(MIDI 的 program change)。byte4 = `FM_BANK.FMB` 的索引。
	EUPProgram = 0xC
)

// EUPEvent 是一筆 6 byte 事件。
//
//	+0  status  高 nibble = 種類、低 nibble = 聲道
//	+1  step    ★ 距離**上一筆**事件的 tick 數(delta,不是絕對時間)
//	+2  gate    u16 LE(+2 低位、+3 高位)—— 疑為音長,但 note-on 常常是 0
//	+4  data1   note-on/off 是音高;note-off 一律 **0**;program change 是音色編號
//	+5  data2   note-on/off 是力度(絕大多數 0x40);program change 是 0xFF
type EUPEvent struct {
	Status byte
	Step   byte
	Gate   uint16
	Data1  byte
	Data2  byte
}

// Kind 是 status 的高 nibble。
func (e EUPEvent) Kind() byte { return e.Status >> 4 }

// Channel 是 status 的低 nibble。
func (e EUPEvent) Channel() int { return int(e.Status & 0x0F) }

// EUPSong 是一首解出來的曲子。
type EUPSong struct {
	// Title 是檔頭 0x20 起的曲名("M1"、"M8"…)。
	Title string
	// Events 是照檔案順序的事件,含 program change。
	Events []EUPEvent
	// TotalTicks 是所有 step 的和 —— 曲長的 tick 數。
	//
	// ⬜ **tick → 秒的換算還沒定案。** 檔頭 sig+8 那 8 byte 裡有兩個會隨曲子變的
	// 位元組(疑為速度),語意未追;`../u1-cht` 的 EUP 渲染器用 0.0108 s/tick
	// (≈ 48 ppqn @ 115 bpm)。要下結論得追驅動程式怎麼讀,不從長度反推。
	TotalTicks int
}

// ParseEUP 解一首 `.EUP`。
func ParseEUP(raw []byte) (*EUPSong, error) {
	sig := strings.Index(string(raw), EUPSignature)
	if sig < 0 {
		return nil, fmt.Errorf("找不到 %q 簽章", EUPSignature)
	}
	s := &EUPSong{}
	if len(raw) > 0x28 {
		s.Title = strings.TrimRight(string(raw[0x20:0x28]), "\x00 ")
	}
	for i := sig + eupHeaderAfterSig; i+EUPEventSize <= len(raw); i += EUPEventSize {
		e := EUPEvent{
			Status: raw[i],
			Step:   raw[i+1],
			Gate:   binary.LittleEndian.Uint16(raw[i+2 : i+4]),
			Data1:  raw[i+4],
			Data2:  raw[i+5],
		}
		// 0xFF 是結束(實測四首都落在檔尾前 0x2FC..0x300,後面是填充)。
		if e.Status == 0xFF {
			break
		}
		s.TotalTicks += int(e.Step)
		s.Events = append(s.Events, e)
	}
	if len(s.Events) == 0 {
		return nil, fmt.Errorf("解不出任何事件")
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

// NoteCounts 數 note-on 與 note-off 的筆數。
//
// ⚠ `find_unwired.py` 會把這一支與 `Programs` 列成「只有測試引用」——
// **那是對的**:`NoteCounts` 存在的目的就是當相位的證據(見下),
// `Programs` 是給還沒寫的離線渲染器用的。已在原地寫明,免得下一輪當死碼刪掉。
//
// ★ 這是「欄位配置對不對」的**決定性檢查**:一首正常的曲子裡兩者必須幾乎相等
// (差 0 或 1 —— 最後一個音可能沒收尾)。相位讀錯的話 status 欄根本不是 status,
// 兩個數字就不會有這個關係。`TestEUPFieldLayoutIsProvenByNoteBalance` 用它把
// 相位釘死,不必相信註解裡的 16。
func (s *EUPSong) NoteCounts() (on, off int) {
	for _, e := range s.Events {
		switch e.Kind() {
		case EUPNoteOn:
			on++
		case EUPNoteOff:
			off++
		}
	}
	return
}

// Programs 回報每個聲道最後選到的音色編號(`FM_BANK.FMB` 的索引),−1 = 沒選過。
func (s *EUPSong) Programs() [FMVoiceChannels]int {
	var out [FMVoiceChannels]int
	for i := range out {
		out[i] = -1
	}
	for _, e := range s.Events {
		if e.Kind() != EUPProgram {
			continue
		}
		if ch := e.Channel(); ch < len(out) {
			out[ch] = int(e.Data1)
		}
	}
	return out
}

// FM 音色庫 `FM_BANK.FMB`

const (
	// FMBankHeader 是檔案開頭那 8 byte(M1 那份是八個空白)。
	FMBankHeader = 8
	// FMVoiceSize 是一筆音色的位元組數。
	FMVoiceSize = 48
	// FMVoiceNameSize 是音色名的位元組數。
	FMVoiceNameSize = 8
	// FMVoiceCount 是音色數。
	//
	// ★ 8 + 128 × 48 = 6152 = 檔案大小,**整除且無餘數** ⇒ 佈局定案。
	// 而 EUP 的 program change 值域落在 0..127 ⇒ 兩個獨立來源一致。
	FMVoiceCount = 128
	// FMVoiceChannels 是 EUP 用得到的聲道數(對上 `U5_BGM.TBL` 的六欄音量)。
	FMVoiceChannels = 6
)

// FMVoice 是一筆 FM 音色。
type FMVoice struct {
	Name string
	// Params 是名字後面那 40 byte,**原樣保留**。
	//
	// ⬜ **語意未定案。** 兩個候選佈局都說得通:
	//
	//	(a) 10 個參數 × 4 個運算子(參數優先):TRUMP1 讀成
	//	    MUL=01 01 02 02、TL=18 11 00 00、AR=10 10 15 15、DR=0d 10 14 00、
	//	    SR=00 00 00 00、SL/RR=5f 40 0f 1f…(TL 前兩個有衰減、後兩個 0
	//	    ⇒ 後兩個是載波,合理)
	//	(b) 4 個運算子 × 8 個參數 + 8 個共用(運算子優先):讀出來第三、四個
	//	    運算子幾乎全 0,不太合理
	//
	// ⇒ (a) 看起來對,但「看起來對」不是證據。要定案得追驅動程式怎麼把它寫進
	// YM2612 的暫存器(`sub_34xxx` 那一族)。在那之前**不解釋這 40 byte**,
	// 免得做出一個「聽起來像但不是原版音色」的合成器(`CLAUDE.md §3.0`)。
	Params [FMVoiceSize - FMVoiceNameSize]byte
}

// ParseFMBank 解 `FM_BANK.FMB`。
func ParseFMBank(raw []byte) ([]FMVoice, error) {
	want := FMBankHeader + FMVoiceCount*FMVoiceSize
	if len(raw) != want {
		return nil, fmt.Errorf("大小 %d,預期 %d(8 + %d × %d)",
			len(raw), want, FMVoiceCount, FMVoiceSize)
	}
	out := make([]FMVoice, FMVoiceCount)
	for i := range out {
		r := raw[FMBankHeader+i*FMVoiceSize:][:FMVoiceSize]
		out[i].Name = strings.TrimRight(string(r[:FMVoiceNameSize]), "\x00 ")
		copy(out[i].Params[:], r[FMVoiceNameSize:])
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
