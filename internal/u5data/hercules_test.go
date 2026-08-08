package u5data

import (
	"os"
	"path/filepath"
	"testing"
)

// herDRV 讀 `HER.DRV`;沒有 gamedata 就跳過(其餘測試的慣例)。
func herDRV(t *testing.T) []byte {
	t.Helper()
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("沒有 U5_GAMEDATA")
	}
	raw, err := os.ReadFile(filepath.Join(dir, "HER.DRV"))
	if err != nil {
		t.Skip("讀不到 HER.DRV:" + err.Error())
	}
	return raw
}

// TestHerculesScanlineTableIsBuiltAtInitNotStored 釘住「表在檔案裡是空的」。
//
// 這條擋的是「以為表是靜態資料、直接從檔案讀」——讀出來會是 600 個 0,
// 而那種錯不會讓任何算繪測試變紅。
func TestHerculesScanlineTableIsBuiltAtInitNotStored(t *testing.T) {
	raw := herDRV(t)
	tbl := raw[HerculesScanlineTableOffset:HerculesNextVarOffset]
	for i, b := range tbl {
		if b != 0 {
			t.Fatalf("0x%04X 的表在檔案裡不是全 0(第 %d 個位元組 = 0x%02X)",
				HerculesScanlineTableOffset, i, b)
		}
	}
}

// TestHerculesInitLoopConstantsMatchTheDriver 逐個常數對回 0x0770 的指令位元組。
//
// 不是「大概是這些值」——是把那五道 `mov imm16` 的立即值挖出來比。
func TestHerculesInitLoopConstantsMatchTheDriver(t *testing.T) {
	raw := herDRV(t)
	imm := func(off int, opcode byte, what string) int {
		if raw[off] != opcode {
			t.Fatalf("0x%04X 不是預期的 opcode 0x%02X(%s),實際 0x%02X",
				off, opcode, what, raw[off])
		}
		return int(raw[off+1]) | int(raw[off+2])<<8
	}
	// 0x0770: mov si,0072h / mov cx,012Ch / mov bx,005Ah / mov ax,0221h / mov dx,2000h
	cases := []struct {
		off    int
		opcode byte
		want   int
		what   string
	}{
		{0x0770, 0xBE, HerculesScanlineTableOffset, "mov si,表位址"},
		{0x0773, 0xB9, HerculesScanlineCount, "mov cx,筆數"},
		{0x0776, 0xBB, HerculesRowStride, "mov bx,列跨距"},
		{0x0779, 0xB8, HerculesFirstOffset, "mov ax,起始位移"},
		{0x077C, 0xBA, HerculesBankSize, "mov dx,bank 大小"},
	}
	for _, c := range cases {
		if got := imm(c.off, c.opcode, c.what); got != c.want {
			t.Errorf("%s:驅動裡是 %d(0x%X),常數是 %d", c.what, got, got, c.want)
		}
	}
}

// TestHerculesTableExactlyFillsUpToTheNextVariable 是筆數的第二份獨立佐證。
//
// 0x0072 + 300×2 = 0x02CA,而 0x02CA 是離屏緩衝的段值(`mov es, cs:[2CA]`)。
// 多一筆就會蓋掉它。
func TestHerculesTableExactlyFillsUpToTheNextVariable(t *testing.T) {
	if HerculesNextVarOffset != 0x02CA {
		t.Fatalf("表尾應該正好接到 0x02CA,實際 0x%04X", HerculesNextVarOffset)
	}
	raw := herDRV(t)
	// 驗 0x02CA 真的被當成段值讀:`2E 8E 06 CA 02` = mov es, cs:[02CA]
	want := []byte{0x2E, 0x8E, 0x06, 0xCA, 0x02}
	found := false
	for i := 0; i+len(want) <= len(raw); i++ {
		match := true
		for j, b := range want {
			if raw[i+j] != b {
				match = false
				break
			}
		}
		if match {
			found = true
			break
		}
	}
	if !found {
		t.Error("找不到 `mov es, cs:[02CA]` —— 0x02CA 是下一個變數這條沒佐證")
	}
}

// TestHerculesInterleaveMapsEntryToScanline24PlusI 是幾何的核心。
//
// 表的第 i 筆解回 (bank, 列, 欄),再照 Hercules 的四路交錯還原實體掃描線,
// 應該正好是 24 + i;而欄一律是 5(= 左邊留 40 像素)。
func TestHerculesInterleaveMapsEntryToScanline24PlusI(t *testing.T) {
	tbl := BuildHerculesScanlineTable()
	for i, off := range tbl {
		bank := int(off) / HerculesBankSize
		within := int(off) % HerculesBankSize
		row := within / HerculesRowStride
		col := within % HerculesRowStride
		if bank >= HerculesBankCount {
			t.Fatalf("第 %d 筆的 bank = %d,超出 %d", i, bank, HerculesBankCount)
		}
		if col != HerculesLeftMargin/8 {
			t.Fatalf("第 %d 筆的位元組欄 = %d,預期 %d(左邊留 %d 像素)",
				i, col, HerculesLeftMargin/8, HerculesLeftMargin)
		}
		if got := row*HerculesBankCount + bank; got != HerculesPhysicalScanline(i) {
			t.Fatalf("第 %d 筆解回掃描線 %d,預期 %d", i, got, HerculesPhysicalScanline(i))
		}
		// 整條線(80 位元組)都必須落在 32 KB 的顯示記憶體裡。
		if int(off)+HerculesBytesPerLine > HerculesBankCount*HerculesBankSize {
			t.Fatalf("第 %d 筆 0x%04X + %d 超出顯示記憶體", i, off, HerculesBytesPerLine)
		}
	}
	// 最後一條掃描線必須還在 348 之內,而且上下邊等寬。
	last := HerculesPhysicalScanline(HerculesScanlineCount - 1)
	if last != HerculesPhysHeight-HerculesTopMargin-1 {
		t.Errorf("最後一條掃描線 %d,上下邊不等寬(上 %d、實體 %d)",
			last, HerculesTopMargin, HerculesPhysHeight)
	}
}

// TestHerculesVerticalScaleIsThreeScanlinesPerTwoRows 把 1.5 倍寫成可驗的形狀。
//
// 每列畫兩條相鄰掃描線,相鄰兩列共用一條 —— 所以 N 列覆蓋的**不同**掃描線
// 應該是 ceil(3N/2)。
func TestHerculesVerticalScaleIsThreeScanlinesPerTwoRows(t *testing.T) {
	for _, rows := range []int{1, 2, 3, 4, 8, 12, 16, HerculesLogicalHeight} {
		seen := map[int]bool{}
		for y := 0; y < rows; y++ {
			a, b := HerculesScanlineEntries(y)
			if b != a+1 {
				t.Fatalf("第 %d 列的兩個表項不相鄰:%d, %d", y, a, b)
			}
			seen[a], seen[b] = true, true
		}
		want := (rows*HerculesYScaleNum + 1) / HerculesYScaleDen
		if len(seen) != want {
			t.Errorf("%d 列覆蓋 %d 條掃描線,預期 %d", rows, len(seen), want)
		}
	}
	// 200 列剛好用滿 300 筆,一筆不多一筆不少。
	a, b := HerculesScanlineEntries(HerculesLogicalHeight - 1)
	if b != HerculesScanlineCount-1 {
		t.Errorf("最後一列畫到表項 %d..%d,而表只有 %d 筆", a, b, HerculesScanlineCount)
	}
}

// TestHerculesHorizontalScaleIsTwo 釘住 x 的拆法與 640 像素寬。
func TestHerculesHorizontalScaleIsTwo(t *testing.T) {
	// 最右邊的邏輯像素必須落在最後一個位元組裡。
	if got := HerculesByteColumn(HerculesLogicalWidth - 1); got != HerculesBytesPerLine-1 {
		t.Errorf("x=%d 落在位元組 %d,預期 %d",
			HerculesLogicalWidth-1, got, HerculesBytesPerLine-1)
	}
	// 四個相位剛好把一個位元組分完,而且遮罩互斥、聯集是整個位元組。
	var union byte
	for p := 0; p < 4; p++ {
		m := HerculesPhaseMask[p]
		if union&m != 0 {
			t.Errorf("相位 %d 的遮罩 0x%02X 與前面重疊", p, m)
		}
		union |= m
		if HerculesKeepMask[p] != ^m {
			t.Errorf("相位 %d:0x0308 的 0x%02X 不是 0x030C 的 0x%02X 的補數",
				p, HerculesKeepMask[p], m)
		}
	}
	if union != 0xFF {
		t.Errorf("四個相位的遮罩聯集是 0x%02X,預期 0xFF", union)
	}
	// 加倍表:每個輸入 bit 必須變成輸出的兩個相鄰 bit。
	for n := 0; n < 16; n++ {
		var want byte
		for b := 0; b < 4; b++ {
			if n&(1<<b) != 0 {
				want |= 3 << (b * 2)
			}
		}
		if HerculesDoubleNibble[n] != want {
			t.Errorf("加倍表[%d] = 0x%02X,預期 0x%02X", n, HerculesDoubleNibble[n], want)
		}
	}
}

// TestHerculesSmallTablesMatchTheDriver 把四張小表對回檔案位元組。
func TestHerculesSmallTablesMatchTheDriver(t *testing.T) {
	raw := herDRV(t)
	cases := []struct {
		off  int
		got  []byte
		what string
	}{
		{0x0308, HerculesKeepMask[:], "保留遮罩"},
		{0x030C, HerculesPhaseMask[:], "相位遮罩"},
		{0x0318, HerculesPattern[:], "顏色圖樣"},
		{0x031C, HerculesDoubleNibble[:], "加倍表"},
	}
	for _, c := range cases {
		want := raw[c.off : c.off+len(c.got)]
		for i := range c.got {
			if c.got[i] != want[i] {
				t.Errorf("%s(0x%04X)第 %d 筆:常數 0x%02X,驅動 0x%02X",
					c.what, c.off, i, c.got[i], want[i])
			}
		}
	}
}

// TestHerculesCharGridMatchesDOS 是「放大藏在驅動裡」的最後一塊。
//
// 640×300 配 16×12 的字格 = 40×25,與 DOS 的 320×200 配 8×8 一模一樣。
// 所以遊戲主體完全不必知道 Hercules 存在。
func TestHerculesCharGridMatchesDOS(t *testing.T) {
	if HerculesCharCols != HerculesLogicalWidth/GlyphWidth {
		t.Errorf("欄數 %d,DOS 是 %d", HerculesCharCols, HerculesLogicalWidth/GlyphWidth)
	}
	if HerculesCharRows != HerculesLogicalHeight/GlyphHeight {
		t.Errorf("列數 %d,DOS 是 %d", HerculesCharRows, HerculesLogicalHeight/GlyphHeight)
	}
	// 字模尺寸本身就是那個比例:8×8 → 16×12。
	if HerculesCellWidth != GlyphWidth*HerculesXScaleNum {
		t.Errorf("字寬 %d,預期 %d", HerculesCellWidth, GlyphWidth*HerculesXScaleNum)
	}
	if HerculesCellHeight != GlyphHeight*HerculesYScaleNum/HerculesYScaleDen {
		t.Errorf("字高 %d,預期 %d", HerculesCellHeight,
			GlyphHeight*HerculesYScaleNum/HerculesYScaleDen)
	}
	// 38 筆進入點表恰好停在掃描線表前面。
	if HerculesEntryPointCount*3 != HerculesScanlineTableOffset {
		t.Errorf("%d 筆進入點 × 3 = %d,而表在 0x%04X",
			HerculesEntryPointCount, HerculesEntryPointCount*3, HerculesScanlineTableOffset)
	}
}
