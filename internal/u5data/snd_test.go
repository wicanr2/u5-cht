package u5data

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fmtownsDir(t *testing.T) string {
	t.Helper()
	dir := os.Getenv("U5_FMTOWNS")
	if dir == "" {
		dir = "../../re_work/fmtowns/iso/U5_E"
	}
	if _, err := os.Stat(filepath.Join(dir, "U5_SE.TBL")); err != nil {
		t.Skip("找不到 FM Towns 的 U5_E 目錄,跳過需要它的測試")
	}
	return dir
}

// `U5_SE.TBL` 的第三欄就是檔案大小 —— 逐一相符才算表與檔案是同一批。
//
// 這條同時把「檔頭 32 B」釘住:`.SND` 的 +0x0C 宣稱資料長度,
// 而 表裡的大小 − 資料長度 必須恆等於 32。
func TestSoundTableSizesMatchTheFilesAndTheHeader(t *testing.T) {
	dir := fmtownsDir(t)
	table, err := LoadSoundTable(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(table) < 20 {
		t.Fatalf("`U5_SE.TBL` 只有 %d 行,預期二十幾行", len(table))
	}
	set, err := LoadSoundSet(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range table {
		s := set[strings.ToUpper(e.File)]
		if s == nil {
			t.Errorf("%s 沒載進來", e.File)
			continue
		}
		if got := len(s.PCM) + SndHeaderSize; got != e.Size {
			t.Errorf("%s:資料 %d + 檔頭 %d = %d,表裡寫 %d",
				e.File, len(s.PCM), SndHeaderSize, got, e.Size)
		}
		if e.Volume < 1 || e.Volume > 127 {
			t.Errorf("%s 的音量是 %d,預期 1..127", e.File, e.Volume)
		}
	}
}

// PCM 是 sign-magnitude,不是二補數 —— 用「相鄰取樣的平均絕對差」量出來。
//
// ⚠ 這是 `knowledge-base` 記著的陷阱。二補數解讀會把小的負值變成極大的負值,
// 於是波形每次過零就跳一次;平滑度指標會差好幾倍。
// 這條測試把那個差距釘住:**每一個檔案**都必須是 sign-magnitude 比較平滑。
func TestSoundPCMIsSignMagnitudeNotTwosComplement(t *testing.T) {
	dir := fmtownsDir(t)
	set, err := LoadSoundSet(dir)
	if err != nil {
		t.Fatal(err)
	}
	smoothness := func(v []int) float64 {
		if len(v) < 2 {
			return 0
		}
		sum := 0
		for i := 1; i < len(v); i++ {
			d := v[i] - v[i-1]
			if d < 0 {
				d = -d
			}
			sum += d
		}
		return float64(sum) / float64(len(v)-1)
	}
	for name, s := range set {
		if len(s.PCM) < 64 {
			continue
		}
		sm := make([]int, len(s.PCM))
		tc := make([]int, len(s.PCM))
		for i, v := range s.PCM {
			sm[i] = int(v)
			// 把已解好的取樣還原成原始位元組,再用二補數解讀一次。
			b := byte(v)
			if v < 0 {
				b = byte(-v) | 0x80
			}
			if b >= 128 {
				tc[i] = int(b) - 256
			} else {
				tc[i] = int(b)
			}
		}
		a, b := smoothness(sm), smoothness(tc)
		if a >= b {
			t.Errorf("%s:sign-magnitude 的平滑度 %.2f 沒有比二補數 %.2f 好", name, a, b)
		}
	}
}

// 欄位語意是「迴圈起點 / 長度」,而 25 個檔案在這個語意下**全部自洽**。
//
// 三種形態都是合法的「不迴圈」寫法:長度 0、長度 1(繞著一個取樣轉)、起點與長度都 0。
// `Loops()` 判的是「這個迴圈聽得出來嗎」——那一層才是判準。
func TestOnlyAmbientSoundsLoop(t *testing.T) {
	dir := fmtownsDir(t)
	set, err := LoadSoundSet(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"FUNSUI2.SND": true, "TAKI2.SND": true}
	looping := []string{}
	for name, s := range set {
		if s.Loops() {
			looping = append(looping, name)
			// 迴圈範圍一定要落在資料內。
			if s.LoopStart+s.LoopLen > len(s.PCM) {
				t.Errorf("%s 的迴圈 %d+%d 超出資料 %d", name, s.LoopStart, s.LoopLen, len(s.PCM))
			}
		}
	}
	for name := range want {
		if s := set[name]; s == nil || !s.Loops() {
			t.Errorf("%s 應該要迴圈(它是環境音)", name)
		}
	}
	// 三種「不迴圈」的寫法都要被判成不迴圈。
	for _, name := range []string{"BEEP.SND", "SUITEKI3.SND", "DOKU.SND", "MAHOU1.SND"} {
		if s := set[name]; s != nil && s.Loops() {
			t.Errorf("%s 被判成迴圈(起點 %d 長度 %d),那是判準沒守住",
				name, s.LoopStart, s.LoopLen)
		}
	}
	// 迴圈的不能太多 —— 判準放寬了整個遊戲會滿是卡住的音效。
	if len(looping) > 3 {
		t.Errorf("有 %d 個音效判成迴圈(%v),預期只有環境音", len(looping), looping)
	}
}

// 檔頭宣稱的長度與實際不符要回錯,不能默默吃掉。
func TestParseSoundChecksTheDeclaredLength(t *testing.T) {
	raw := make([]byte, SndHeaderSize+10)
	copy(raw, "Test")
	raw[sndLength] = 99 // 宣稱 99 B,實際 10 B
	if _, err := ParseSound(raw); err == nil {
		t.Error("長度不符竟然解得過")
	}
	if _, err := ParseSound(make([]byte, 8)); err == nil {
		t.Error("比檔頭還短竟然解得過")
	}
}

// sign-magnitude 的兩個零:0x00 與 0x80 都是靜音。
func TestSignMagnitudeHasTwoZeros(t *testing.T) {
	if sndSample(0x00) != 0 || sndSample(0x80) != 0 {
		t.Errorf("0x00 → %d、0x80 → %d,兩者都該是 0", sndSample(0x00), sndSample(0x80))
	}
	if sndSample(0x7F) != 127 {
		t.Errorf("0x7F → %d,預期 127", sndSample(0x7F))
	}
	if sndSample(0xFF) != -127 {
		t.Errorf("0xFF → %d,預期 −127", sndSample(0xFF))
	}
}

// 迴圈欄位在「起點 + 長度」的語意下必須對**每一個**檔案都自洽。
//
// 合法的形態只有四種:真迴圈(起點 + 長度 == 全長且長度 > 1)、長度 0、
// 長度 1、起點與長度都 0。出現第五種就代表這個語意讀錯了。
func TestLoopFieldsAreConsistentForEveryFile(t *testing.T) {
	dir := fmtownsDir(t)
	set, err := LoadSoundSet(dir)
	if err != nil {
		t.Fatal(err)
	}
	for name, s := range set {
		n := len(s.PCM)
		switch {
		case s.LoopLen == 0:
		case s.LoopLen == 1 && s.LoopStart+1 == n:
		case s.LoopStart+s.LoopLen == n:
		default:
			t.Errorf("%s:起點 %d 長度 %d 全長 %d —— 不符任何一種形態",
				name, s.LoopStart, s.LoopLen, n)
		}
	}
}

// 基準音高(+0x1C)只有 60 與 61 兩種,而那正是播放時傳給驅動程式的數字。
//
// `sub_2C4F4` 裡有一處 `push 3Ch; push 3` = 「用音高 **60** 播第 3 號音效」。
func TestBaseNoteIsSixtyOrSixtyOne(t *testing.T) {
	dir := fmtownsDir(t)
	table, err := LoadSoundTable(dir)
	if err != nil {
		t.Fatal(err)
	}
	names, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	actual := map[string]string{}
	for _, e := range names {
		actual[strings.ToUpper(e.Name())] = e.Name()
	}
	seen := map[uint32]int{}
	for _, e := range table {
		// ⚠ `SUITEKI3.SND` 的名字欄位溢出成 12 B(`suiteki2suit`),
		// 後面每個欄位都往後挪了 4 B —— 它的 +0x1C 讀出來是 1,不是音高。
		// 這是那一個檔案的資料瑕疵,不是格式問題,所以照實跳過而不是放寬判準。
		if strings.EqualFold(e.File, "SUITEKI3.SND") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, actual[strings.ToUpper(e.File)]))
		if err != nil {
			t.Fatal(err)
		}
		note := binary.LittleEndian.Uint32(raw[0x1C:])
		seen[note]++
	}
	for note, n := range seen {
		if note != 60 && note != 61 {
			t.Errorf("有 %d 個檔案的基準音高是 %d,預期只有 60 / 61", n, note)
		}
	}
	if len(seen) == 0 {
		t.Error("一個檔案都沒讀到")
	}
}
