package u5data

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestEUPNotesAreTwelveBytesNotOnOffPairs —— ★★★ 更正過的事件模型。
//
// ⚠ 第一版把 `0x8n` 當成 note-off,而它是**音符的後半**。一手資料的證據:
// `0x9n` 的下一筆 100% 是 `0x8n`(882/882),`0x8n` 的上一筆 100% 是 `0x9n`。
// 完全配對、零例外 ⇒ 一個音符 12 byte。
//
// ⇒ `ParseEUP` 在後半不是 `0x8n` 時直接報錯,這條測試靠「十五首都解得過」
// 來釘住這件事 —— 只要模型錯了就會有檔案解不出來。
func TestEUPNotesAreTwelveBytesNotOnOffPairs(t *testing.T) {
	dir := fmTownsDir(t)
	tracks, err := LoadBGMTable(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, tr := range tracks {
		song, err := LoadEUP(dir, tr.File)
		if err != nil {
			t.Errorf("%s:%v", tr.File, err)
			continue
		}
		if len(song.Notes) == 0 {
			t.Errorf("%s 一個音符都沒有", tr.File)
		}
		for i, n := range song.Notes {
			if n.Channel < 0 || n.Channel >= FMVoiceChannels {
				t.Fatalf("%s 第 %d 個音符在聲道 %d(只有 %d 個)",
					tr.File, i, n.Channel, FMVoiceChannels)
			}
			// 後半的低 nibble 也是同一個聲道。
			if int(n.Tail[0]&0x0F) != n.Channel {
				t.Fatalf("%s 第 %d 個音符:前半聲道 %d、後半 0x%02X",
					tr.File, i, n.Channel, n.Tail[0])
			}
		}
	}
}

// TestNoteTicksStayInsideTheBar —— ★★ 時間是**小節內**的 tick,不是絕對時間。
//
// 兩個互相佐證的事實:
//   - `0xF2`(小節結束)的時間欄固定是 384;
//   - 所有音符的小節內 tick 都 **< 384**(實測最大 381)。
//
// 384 = 96 ppqn × 4 拍 ⇒ 4/4 一小節,而 96 正是 `eupplayer` 的 `60e6/(96*t)`。
//
// ⚠ 第一版把每筆的 `[1]` 當成 delta step 相加 —— 而 `[1]` 是**聲道**
// (前半 882/882 等於 status 低 nibble)。整個曲長算法因此是錯的。
func TestNoteTicksStayInsideTheBar(t *testing.T) {
	dir := fmTownsDir(t)
	tracks, err := LoadBGMTable(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, tr := range tracks {
		song, err := LoadEUP(dir, tr.File)
		if err != nil {
			t.Fatal(err)
		}
		prev := -1
		prevBar := -1
		for i, n := range song.Notes {
			if n.Tick >= EUPTicksPerBar {
				t.Errorf("%s 第 %d 個音符的小節內 tick 是 %d(上限 %d)",
					tr.File, i, n.Tick, EUPTicksPerBar)
			}
			// 同一小節內必須單調不遞減(實測 15 首、每一小節都成立)。
			if n.Bar == prevBar && n.Tick < prev {
				t.Errorf("%s 第 %d 個音符在小節 %d 內時間倒退:%d → %d",
					tr.File, i, n.Bar, prev, n.Tick)
			}
			prev, prevBar = n.Tick, n.Bar
		}
		if song.Bars == 0 {
			t.Errorf("%s 一個小節分隔都沒有", tr.File)
		}
	}
}

// TestTempoGivesSaneDurations —— ★ 速度欄的驗收:算出來的曲長要合理。
//
// `sig+15` 十五首的值是 88/90/59/67 —— 都落在音樂 BPM 的範圍,而且**有意義地
// 變化**:曲 4(`M5.EUP`)是 59(最慢,73 小節 → 297 秒)。
//
// ⚠ 這是**間接**驗收(不是從 BIOS 讀出來的),所以判準放寬成「數量級合理」:
// 每首 5..400 秒。若哪天 `sig+15` 被證實不是速度,這條會先變得可疑而不是靜默錯。
func TestTempoGivesSaneDurations(t *testing.T) {
	dir := fmTownsDir(t)
	tracks, err := LoadBGMTable(dir)
	if err != nil {
		t.Fatal(err)
	}
	total := 0.0
	for i, tr := range tracks {
		song, err := LoadEUP(dir, tr.File)
		if err != nil {
			t.Fatal(err)
		}
		if song.Tempo < 30 || song.Tempo > 240 {
			t.Errorf("%s 速度 %d 不像 BPM", tr.File, song.Tempo)
		}
		d := song.Duration()
		if d < 5 || d > 400 {
			t.Errorf("%s 算出來 %.0f 秒(%d 小節 @ %d BPM)—— 數量級不對",
				tr.File, d, song.Bars, song.Tempo)
		}
		total += d
		t.Logf("曲 %2d %-10s %3d BPM %3d 小節 %4d 音符 %6.0f 秒",
			i, tr.File, song.Tempo, song.Bars, len(song.Notes), d)
	}
	if total < 600 || total > 3600 {
		t.Errorf("十五首合計 %.0f 秒 —— 遊戲配樂該在 10..60 分鐘之間", total)
	}
}

// TestEveryChannelUsedHasAVoice —— 有音符的聲道一定選過音色。
func TestEveryChannelUsedHasAVoice(t *testing.T) {
	dir := fmTownsDir(t)
	bank, err := LoadFMBank(dir)
	if err != nil {
		t.Fatal(err)
	}
	tracks, err := LoadBGMTable(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, tr := range tracks {
		song, err := LoadEUP(dir, tr.File)
		if err != nil {
			t.Fatal(err)
		}
		var used [FMVoiceChannels]bool
		for _, n := range song.Notes {
			used[n.Channel] = true
		}
		for ch := range used {
			if !used[ch] {
				continue
			}
			v := song.VoiceOf(ch)
			if v < 0 {
				// ⚠ **這是真的資料事實,不是錯誤**:`M4.EUP` 的聲道 4 有音符
				// 卻沒有 program change。⇒ 渲染器必須對「沒選音色」有明確行為,
				// 而那個行為要有依據(原版是 BIOS 決定的,`docs/re/89`)。
				// ⬜ 已記進 `docs/audio-pipeline.md`。這裡只記錄,不判失敗。
				t.Logf("(資料事實)%s 聲道 %d 有音符但沒選音色 —— 渲染器要有預設行為",
					tr.File, ch)
				continue
			}
			if v >= len(bank) {
				t.Errorf("%s 聲道 %d 選了音色 %d,音色庫只有 %d", tr.File, ch, v, len(bank))
			}
		}
	}
}

// TestFMVoiceFieldsFitTheirBitWidths —— ★★★ 佈局的一手否證。
//
// YM2612 每個欄位的位元寬度是已知的。把佈局套到全部 126 個有名字的音色上,
// **任何一格超出該欄位上限就表示佈局錯了**。這條測試是「為什麼敢用這個佈局」
// 的證據本身(第二個來源是 `gzaffin/eupmini` 的解碼碼,見 `ParseFMBank` 註解)。
//
// 最決定性的兩個:`+32` 實測 0..61 而 FB/ALG 的寬度剛好 0..63;
// `+33` 全部 ≥ 192 ⇒ L/R 輸出位元永遠都開。
func TestFMVoiceFieldsFitTheirBitWidths(t *testing.T) {
	bank, err := LoadFMBank(fmTownsDir(t))
	if err != nil {
		t.Fatal(err)
	}
	named := 0
	var maxFBALG byte
	minPan := byte(0xFF)
	for i, v := range bank {
		if v.Name == "" {
			continue
		}
		named++
		for n, op := range v.Op {
			// 拆出來的每一格都必須落在自己的寬度內 —— 這是拆法的自我檢查。
			for _, c := range []struct {
				got, max byte
				name     string
			}{
				{op.Detune, 7, "DT"}, {op.Multiple, 15, "MUL"},
				{op.TotalLvl, 127, "TL"}, {op.KeyScale, 3, "KS"},
				{op.Attack, 31, "AR"}, {op.Decay, 31, "DR"},
				{op.Sustain, 31, "SR"}, {op.SusLevel, 15, "SL"},
				{op.Release, 15, "RR"},
			} {
				if c.got > c.max {
					t.Errorf("音色 %d(%s)運算子 %d 的 %s = %d,超出 %d",
						i, v.Name, n, c.name, c.got, c.max)
				}
			}
		}
		if v.Algorithm > 7 || v.Feedback > 7 {
			t.Errorf("音色 %d ALG=%d FB=%d 超出 0..7", i, v.Algorithm, v.Feedback)
		}
		if fb := v.Raw[32-FMVoiceNameSize]; fb > maxFBALG {
			maxFBALG = fb
		}
		if v.Pan < minPan {
			minPan = v.Pan
		}
		// ★★ **跨欄位一致性**:用 ALG(+32)挑出載波,載波的 TL(+12..15)必須夠小。
		//
		// 這比逐欄位範圍檢查強得多 —— 它同時用到兩個不同位移,
		// **任一個位移錯了這個相關性就會破**。實測 126/126 個音色的載波
		// 最小 TL ≤ 15;門檻放到 40 留餘裕。
		//
		// ⚠ 第一版寫的是「四個運算子至少一個 TL ≤ 8」—— 那是**我自己發明的**
		// 門檻,而且有九個音色不滿足(最小 TL 到 15)。發明的判準紅燈時
		// 要改判準不是改實作。
		if c := carrierMinTL(v); c > 40 {
			t.Errorf("音色 %d(%s)ALG=%d 的載波最小 TL = %d(> 40)—— 佈局可能錯了",
				i, v.Name, v.Algorithm, c)
		}
	}
	if named < FMVoiceCount/2 {
		t.Fatalf("只有 %d/%d 筆有名字", named, FMVoiceCount)
	}
	// ★ 決定性的兩個。
	if maxFBALG > 63 {
		t.Errorf("+32 實測上限 %d > 63 —— 那一格不是 FB/ALG", maxFBALG)
	}
	if minPan < 192 {
		t.Errorf("+33 最小值 %d < 192 —— L/R 位元不是永遠都開?", minPan)
	}
	t.Logf("有名字 %d/%d;+32 上限 %d(FB/ALG 寬度 63);+33 最小 %d(≥192 = L/R 全開)",
		named, FMVoiceCount, maxFBALG, minPan)
}

// algCarriers 是 YM2612 八種演算法各自的載波運算子索引。
//
// 只有載波的輸出會被聽到,調變器只影響載波的相位 ⇒ 載波必須不能被 TL 衰減光。
var algCarriers = [8][]int{
	0: {3}, 1: {3}, 2: {3}, 3: {3},
	4: {1, 3}, 5: {1, 2, 3}, 6: {1, 2, 3}, 7: {0, 1, 2, 3},
}

func carrierMinTL(v FMVoice) byte {
	min := byte(127)
	for _, n := range algCarriers[v.Algorithm&7] {
		if v.Op[n].TotalLvl < min {
			min = v.Op[n].TotalLvl
		}
	}
	return min
}

// TestFMBankVoicesHaveReadableNames —— 名字在記錄開頭 8 byte。
func TestFMBankVoicesHaveReadableNames(t *testing.T) {
	bank, err := LoadFMBank(fmTownsDir(t))
	if err != nil {
		t.Fatal(err)
	}
	for i, v := range bank {
		for _, r := range v.Name {
			if r < 0x20 || r > 0x7E {
				t.Errorf("音色 %d 的名字 %q 含不可讀字元 0x%02X", i, v.Name, r)
				break
			}
		}
	}
	joined := strings.Join([]string{bank[0].Name, bank[7].Name, bank[9].Name}, "|")
	for _, want := range []string{"TRUMP1", "FLUTE", "OBOE"} {
		if !strings.Contains(joined, want) {
			t.Errorf("前十筆裡找不到 %q(讀到的是 %q)", want, joined)
		}
	}
}

// TestFMBankRejectsWrongSize —— 大小不對就報錯,不要「盡量解」。
func TestFMBankRejectsWrongSize(t *testing.T) {
	if _, err := ParseFMBank(make([]byte, 6151)); err == nil {
		t.Error("少一個位元組也該報錯")
	}
	if _, err := ParseFMBank(nil); err == nil {
		t.Error("空檔案該報錯")
	}
}

// TestEveryBGMFileExists —— `U5_BGM.TBL` 列的十五個檔案都在。
func TestEveryBGMFileExists(t *testing.T) {
	dir := fmTownsDir(t)
	tracks, err := LoadBGMTable(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, tr := range tracks {
		if _, err := os.Stat(filepath.Join(dir, tr.File)); err != nil {
			t.Errorf("%s:%v", tr.File, err)
		}
	}
}
