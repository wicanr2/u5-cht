package u5data

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// FM Towns 的音效檔 `.SND`(25 個)
//
// # 檔頭 32 B
//
//	+0x00  8 B   名字(ASCII,NUL 填滿)  "Take1" "Clock2" "BariBari" "funsui2"…
//	+0x08  u32   0x1EF6 = 7926          ← 24 / 25 個檔案都是這個值
//	+0x0C  u32   **資料長度** = 檔案大小 − 32
//	+0x10  u32   迴圈起點
//	+0x14  u32   迴圈長度
//	+0x18  u32   1960 / 1568 / 1586  與取樣率有關(單位未定,見下)
//	+0x1C  u32   60 / 61              **基準音高**(播放時傳給驅動程式的那個數)
//
// `+0x0C` 對 **25 / 25** 個檔案都正好等於「檔案大小 − 32」——
// 這一條就把檔頭長度釘死了,不必猜。
//
// ★ 而且還有第二份獨立佐證:`U5_SE.TBL` 是**純文字表**,每行
// `檔名 音量 檔案大小`,23 個檔案的第三欄與實際檔案大小逐一相符。
//
// # 檔頭其實是 FM Towns 音源驅動的標準 PCM 檔頭
//
// 把載入與播放兩支追出來之後,這幾個欄位的歸屬清楚了(`WORRIORS.EXP`):
//
//	sub_2C2AC  載入:fscanf(f, "%s %d %d", 檔名, &音量, &大小)   ← ★ 三欄的讀法直接證實
//	           迴圈 `cmp ebx, 17h` → **正好 23 筆**(所以 BEEP 與 BARIBARI 真的不在表裡)
//	           malloc 整個檔案(含 32 B 檔頭),失敗印 "pcm malloc err "
//	sub_2C46C  播放:通道 = 索引 & 7,再加 0x40 → **8 個 PCM 通道**
//	           音量取 `dword_60100[索引]`(= 表的第二欄)
//	           呼叫 sub_34834(通道, 音高, 音量, 緩衝區)
//	sub_34834  `mov ah, 25h` → FM Towns 音源驅動的**播放 PCM**
//	sub_34861  `mov ah, 27h` → 停止通道
//	sub_34876  `mov ah, 28h` → 查通道狀態
//
// ★ **驅動程式收的是「音高」,不是取樣率** —— `sub_2C4F4` 有一處
// `push 3Ch; push 3` 也就是「用音高 **60** 播第 3 號音效」,而 60 / 61
// 正是檔頭 `+0x1C` 的兩個值。所以:
//
//	+0x1C  = 這個取樣的**基準音高**
//	+0x18  = 與取樣率有關的參數(1960 / 1568 / 1586,單位未定)
//
// ⇒ **取樣率不在遊戲裡**,在檔頭裡由音源驅動解讀 —— 這正好解釋了為什麼
// 7926 這個數字在整個 `WORRIORS.EXP` 裡一次都找不到。
//
// # +0x10 / +0x14 是迴圈起點與長度,而且 25 個檔案全部自洽
//
// 三個檔案的兩個欄位相加正好等於全長,而它們剛好就是該迴圈的那三個環境音:
//
//	FUNSUI2(噴泉)  0      + 60000 = 60000 = 全長 ✓
//	TAKI2  (瀑布)  38839  + 21161 = 60000 = 全長 ✓
//	BARIBARI(電光)  51367  + 14991 = 66358 = 全長 ✓
//
// 其餘檔案的兩欄長這樣:
//
//	多數一次性音效   +0x10 = 全長 − 1、+0x14 = 0    → 長度 0 的迴圈
//	BEEP / SUITEKI3  +0x10 = 全長 − 1、+0x14 = 1    → 迴圈最後**一個**取樣
//	DOKU / MAHOU1    兩個都是 0                      → 長度 0 的迴圈
//
// ★ **我一度以為後兩種形態推翻了這個讀法 —— 那是想錯了。** 在「起點 + 長度」
// 的語意下三種形態全部自洽:長度 0 是不迴圈,而「迴圈末端一個取樣」是同一件事的
// 另一種寫法(驅動程式繞著一個取樣轉,聽起來就是停住)。
// 25 個檔案沒有一個違反。
//
// ⇒ 所以欄位語意是**迴圈起點 / 迴圈長度**。`Loops()` 判的不是「欄位是什麼」
// 而是「這個迴圈聽得出來嗎」——長度 > 1 且起點 + 長度 == 全長。
// 解析器**不對欄位做任何「修正」**,日後若要改判準只動一個函式。
//
// # PCM 是 sign-magnitude,不是二補數
//
// 這是 `knowledge-base` 早就記著的陷阱,現在有量化證據:拿「相鄰取樣的平均
// 絕對差」當平滑度指標,對十個檔案分別用三種解讀跑一遍 ——
//
//	sign-magnitude  合計 43.0
//	二補數          合計 162.2   ← 約 3.8 倍
//	無號 −128       合計 162.0
//
// **十個檔案全部是 sign-magnitude 勝**,而且差距不是邊緣。原因很直觀:
// 二補數解讀會把 0x80..0xFF 這一段(在 sign-magnitude 裡是**小的負值**)
// 變成極大的負值,於是每次波形過零就跳一次。
//
// # 取樣率:確定「不在遊戲裡」,但單位還沒定案
//
// `+0x08` 的 7926 是唯一像取樣率的數字(24/25 個檔案相同 —— 第 25 個是
// `SUITEKI3` 的名字溢出把它蓋掉了),而照它算出來的長度都合理
//(`CLOCK2` 67 ms 的滴答、`BEEP` 0.5 s、`FUNSUI2` 7.6 s 的噴泉)。
//
// 追完播放路徑之後可以確定:**遊戲本身不管取樣率**,它只給驅動程式
//「通道、音高、音量、緩衝區」四樣東西,率由檔頭與驅動決定。
// 這也解釋了為什麼 7926 在整個 `WORRIORS.EXP` 裡一次都找不到。
//
// ⚠ 仍未定案的是**單位**:`+0x18` 的 1960 / 1568 / 1586 若乘 4 得
// 7840 / 6272 / 6344 Hz,算出來的長度同樣合理。要在 7926 與 `+0x18 × k`
// 之間定案,得拿 FM Towns 音源驅動的規格或實機音訊 A/B 比對 ——
// 所以常數仍叫 `SndAssumedRate`,名字裡的 Assumed 是刻意的。

const (
	// SndHeaderSize 是 `.SND` 的檔頭長度。
	SndHeaderSize = 32
	// SndNameLen 是名字欄位的長度。
	SndNameLen = 8
	// SndAssumedRate 是**推測**的取樣率(檔頭 +0x08 的常數)。
	//
	// ⚠ 名字裡的 Assumed 是刻意的 —— 見上面「還沒解的」那一段。
	SndAssumedRate = 7926
)

// 檔頭欄位位移。
const (
	sndLength    = 0x0C
	sndLoopStart = 0x10
	sndLoopLen   = 0x14
)

// Sound 是一個解好的音效。
type Sound struct {
	// Name 是檔頭裡的名字(不一定等於檔名:`ALARM3.SND` 的名字是 `Take1`)。
	Name string
	// PCM 是**帶正負號的 8-bit** 取樣,已從 sign-magnitude 轉好。
	PCM []int8
	// LoopStart / LoopLen 是迴圈範圍(LoopLen 為 0 代表不迴圈)。
	LoopStart, LoopLen int
	// Volume 是 `U5_SE.TBL` 給的音量 0..127;表裡沒有這個檔案時為 0。
	Volume int
}

// SndMinLoopLen 是「聽得出來的迴圈」的最小長度。
//
// ⚠ 這個 1 不是欄位語意,是**判準**:`BEEP` 與 `SUITEKI3` 的迴圈長度就是 1,
// 而繞著一個取樣轉聽起來就是停住 —— 那是原版寫「不迴圈」的另一種方式。
const SndMinLoopLen = 1

// Loops 回報這個音效**聽得出迴圈**沒有。
//
// 判準是「長度 > SndMinLoopLen 且起點 + 長度 == 全長」。欄位語意(迴圈起點 /
// 長度)是確定的;這裡判的是「這個迴圈有沒有意義」,那一層才是判準。
func (s *Sound) Loops() bool {
	return s.LoopLen > SndMinLoopLen && s.LoopStart+s.LoopLen == len(s.PCM)
}

// ParseSound 解一份 `.SND` 的內容。
func ParseSound(raw []byte) (*Sound, error) {
	if len(raw) < SndHeaderSize {
		return nil, fmt.Errorf("只有 %d B,放不下 %d B 的檔頭", len(raw), SndHeaderSize)
	}
	n := int(binary.LittleEndian.Uint32(raw[sndLength:]))
	if want := len(raw) - SndHeaderSize; n != want {
		return nil, fmt.Errorf("檔頭宣稱資料 %d B,實際 %d B", n, want)
	}
	s := &Sound{
		Name:      strings.TrimRight(string(raw[:SndNameLen]), "\x00"),
		LoopStart: int(binary.LittleEndian.Uint32(raw[sndLoopStart:])),
		LoopLen:   int(binary.LittleEndian.Uint32(raw[sndLoopLen:])),
	}
	// ⚠ 兩個欄位**原樣保留**,不在這裡「修正」——判斷交給 Loops()。
	// 這樣日後若讀法被推翻,只要改一個函式,不必回頭懷疑解析器動過手腳。
	s.PCM = make([]int8, n)
	for i, b := range raw[SndHeaderSize:] {
		s.PCM[i] = sndSample(b)
	}
	return s, nil
}

// sndSample 把一個 sign-magnitude 位元組轉成帶正負號的取樣。
//
// bit 7 是正負號,bit 0..6 是大小。所以 0x00 與 0x80 **都是靜音**。
func sndSample(b byte) int8 {
	v := int8(b & 0x7F)
	if b&0x80 != 0 {
		return -v
	}
	return v
}

// LoadSound 讀一份 `.SND`。
func LoadSound(path string) (*Sound, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	s, err := ParseSound(raw)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filepath.Base(path), err)
	}
	return s, nil
}

// SoundTableEntry 是 `U5_SE.TBL` 的一行。
type SoundTableEntry struct {
	File   string // 表裡寫的檔名(大小寫可能與實際檔案不同)
	Volume int    // 0..127
	Size   int    // 檔案大小(含 32 B 檔頭)
}

// ParseSoundTable 解 `U5_SE.TBL`(純文字,每行「檔名 音量 大小」)。
func ParseSoundTable(raw []byte) ([]SoundTableEntry, error) {
	var out []SoundTableEntry
	// ⚠ 檔尾有一個 0x1A —— DOS 的 EOF 標記(Ctrl-Z)。當年的文字檔慣例,
	// 不是壞資料;不切掉的話最後一行會變成一欄而解析失敗。
	text := string(raw)
	if i := strings.IndexByte(text, 0x1A); i >= 0 {
		text = text[:i]
	}
	for i, line := range strings.Split(text, "\n") {
		f := strings.Fields(strings.TrimRight(line, "\r"))
		if len(f) == 0 {
			continue
		}
		if len(f) != 3 {
			return nil, fmt.Errorf("第 %d 行有 %d 欄,預期 3 欄:%q", i+1, len(f), line)
		}
		vol, err := strconv.Atoi(f[1])
		if err != nil {
			return nil, fmt.Errorf("第 %d 行的音量不是數字:%q", i+1, f[1])
		}
		size, err := strconv.Atoi(f[2])
		if err != nil {
			return nil, fmt.Errorf("第 %d 行的大小不是數字:%q", i+1, f[2])
		}
		out = append(out, SoundTableEntry{File: f[0], Volume: vol, Size: size})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("U5_SE.TBL 是空的")
	}
	return out, nil
}

// LoadSoundTable 從 FM Towns 的 U5_E 目錄讀 `U5_SE.TBL`。
func LoadSoundTable(dir string) ([]SoundTableEntry, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "U5_SE.TBL"))
	if err != nil {
		return nil, err
	}
	return ParseSoundTable(raw)
}

// LoadSoundSet 讀 `U5_SE.TBL` 列出的每一個音效,音量一併填好。
//
// ⚠ 表裡的檔名大小寫**不一定**與實際檔案相同(表寫 `Fire.SND`,檔案是 `FIRE.SND`)
// —— ISO9660 的大小寫慣例。所以要不分大小寫找檔。
func LoadSoundSet(dir string) (map[string]*Sound, error) {
	table, err := LoadSoundTable(dir)
	if err != nil {
		return nil, err
	}
	names, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	actual := map[string]string{}
	for _, e := range names {
		actual[strings.ToUpper(e.Name())] = e.Name()
	}
	out := make(map[string]*Sound, len(table))
	for _, e := range table {
		real, ok := actual[strings.ToUpper(e.File)]
		if !ok {
			return nil, fmt.Errorf("`U5_SE.TBL` 列了 %s,但目錄裡沒有這個檔案", e.File)
		}
		s, err := LoadSound(filepath.Join(dir, real))
		if err != nil {
			return nil, err
		}
		s.Volume = e.Volume
		out[strings.ToUpper(e.File)] = s
	}
	return out, nil
}
