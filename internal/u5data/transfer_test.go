package u5data

import (
	"encoding/binary"
	"testing"
)

// makeU4Save 造一份最小的 U4 PARTY.SAV。
func makeU4Save(t *testing.T, mutate func(rec []byte, virtue []byte)) []byte {
	t.Helper()
	raw := make([]byte, u4VirtueOffset+u4VirtueSize)
	rec := raw[u4CharOffset:]
	put := func(off, v int) { binary.LittleEndian.PutUint16(rec[off:], uint16(v)) }
	put(u4HP, 180)
	put(u4MaxHP, 350) // → 等級 3
	put(u4Exp, 1234)
	put(u4Str, 30)
	put(u4Dex, 24)
	put(u4Intel, 18)
	put(u4MP, 9)
	copy(rec[u4Name:], "Avatar")
	rec[u4Gender] = GenderMale
	rec[u4Class] = 5 // Paladin
	if mutate != nil {
		mutate(rec, raw[u4VirtueOffset:])
	}
	return raw
}

// 一份正常的 U4 存檔要換算出對得上的角色。
func TestU4TransferMapsTheFields(t *testing.T) {
	got, err := ParseU4Transfer(makeU4Save(t, nil))
	if err != nil {
		t.Fatalf("轉不進來:%v", err)
	}
	c := &got.Char
	if c.Name != "Avatar" {
		t.Errorf("名字是 %q", c.Name)
	}
	if c.Gender != GenderMale {
		t.Errorf("性別是 0x%02X", c.Gender)
	}
	// ★ U4 的職業 5 是聖騎士 → U5 記成 'P'。
	if c.Class != 'P' {
		t.Errorf("職業是 %q,預期 'P'(Paladin)", c.Class)
	}
	if c.Status != StatusGood {
		t.Errorf("狀態是 %q", c.Status)
	}
	if c.Strength != 30 || c.Dex != 24 || c.Intel != 18 || c.MP != 9 {
		t.Errorf("三圍 / 法力是 %d/%d/%d/%d", c.Strength, c.Dex, c.Intel, c.MP)
	}
	if c.HP != 180 || c.MaxHP != 350 || c.Exp != 1234 {
		t.Errorf("生命 / 上限 / 經驗是 %d/%d/%d", c.HP, c.MaxHP, c.Exp)
	}
	// ★ 等級 = **最大 HP / 100**,不是照經驗值算 —— 350/100 = 3。
	// 經驗 1234 照 U5 平時的門檻(100/200/400/800/1600)會算成 4,
	// 所以這一條分辨得出有沒有照抄原版。
	if c.Level != 3 {
		t.Errorf("等級是 %d,原版是最大 HP / 100 = 3", c.Level)
	}
	// 32 B 的原始記錄要與欄位一致(存檔會直接寫它)。
	if c.Raw[CharClass] != 'P' || c.Raw[CharLevel] != 3 {
		t.Errorf("Raw 沒有跟著填:職業 %q 等級 %d", c.Raw[CharClass], c.Raw[CharLevel])
	}
	if binary.LittleEndian.Uint16(c.Raw[CharMaxHP:]) != 350 {
		t.Error("Raw 的最大 HP 沒填")
	}
}

// 八個職業碼要各自對到一個字母,而且不重複。
func TestU4ClassesAreTheEightU4Professions(t *testing.T) {
	seen := map[byte]bool{}
	for i, ch := range U4Classes {
		if seen[ch] {
			t.Errorf("職業字母 %q 重複(碼 %d)", ch, i)
		}
		seen[ch] = true
		if ch < 'A' || ch > 'Z' {
			t.Errorf("碼 %d 對到的不是大寫字母:0x%02X", i, ch)
		}
	}
	// U4 的八個職業首字母。
	for _, want := range []byte{'M', 'B', 'F', 'D', 'T', 'P', 'R', 'S'} {
		if !seen[want] {
			t.Errorf("少了職業 %q", want)
		}
	}
}

// 三個界線各自要擋得住(70 / 9999 / 7),而剛好等於界線的要放行。
func TestU4TransferValidatesTheOriginalLimits(t *testing.T) {
	cases := []struct {
		name string
		bad  func(rec []byte, virtue []byte)
	}{
		{"力量 71", func(r, _ []byte) { binary.LittleEndian.PutUint16(r[u4Str:], U4TransferStatMax+1) }},
		{"敏捷 71", func(r, _ []byte) { binary.LittleEndian.PutUint16(r[u4Dex:], U4TransferStatMax+1) }},
		{"智力 71", func(r, _ []byte) { binary.LittleEndian.PutUint16(r[u4Intel:], U4TransferStatMax+1) }},
		{"經驗 10000", func(r, _ []byte) { binary.LittleEndian.PutUint16(r[u4Exp:], U4TransferBigMax+1) }},
		{"生命 10000", func(r, _ []byte) { binary.LittleEndian.PutUint16(r[u4HP:], U4TransferBigMax+1) }},
		{"職業 8", func(r, _ []byte) { r[u4Class] = U4TransferClassMax + 1 }},
		{"名字含控制字元", func(r, _ []byte) { r[u4Name+2] = 0x07 }},
	}
	for _, c := range cases {
		if _, err := ParseU4Transfer(makeU4Save(t, c.bad)); err == nil {
			t.Errorf("%s 竟然通過了驗證", c.name)
		}
	}
	// 剛好等於界線要放行(原版用的是 `ja`,不是 `jae`)。
	ok := makeU4Save(t, func(r, _ []byte) {
		binary.LittleEndian.PutUint16(r[u4Str:], U4TransferStatMax)
		binary.LittleEndian.PutUint16(r[u4Exp:], U4TransferBigMax)
		r[u4Class] = U4TransferClassMax
	})
	if _, err := ParseU4Transfer(ok); err != nil {
		t.Errorf("剛好等於界線卻被擋下:%v", err)
	}
	// 名字提早結束(NUL)不是錯誤 —— 原版對 0 是 continue。
	short := makeU4Save(t, func(r, _ []byte) {
		copy(r[u4Name:], []byte{'B', 'o', 'b', 0, 0, 0, 0, 0})
	})
	got, err := ParseU4Transfer(short)
	if err != nil {
		t.Fatalf("短名字被擋下:%v", err)
	}
	if got.Char.Name != "Bob" {
		t.Errorf("短名字讀成 %q", got.Char.Name)
	}
}

// 「是不是聖者」看的是**第二塊**那八個 u16,不是角色的三圍。
//
// ⚠ 兩塊的欄位位移相同(都是 +6 起),混在一起就會得到
// 「三圍全為 0 才是聖者」—— 那對任何合法角色都不成立。
func TestU4AvatarFlagReadsTheSecondBlock(t *testing.T) {
	// 第二塊全 0 → 是聖者,而角色的三圍照樣是 30/24/18。
	got, err := ParseU4Transfer(makeU4Save(t, nil))
	if err != nil {
		t.Fatal(err)
	}
	if !got.Avatar {
		t.Error("第二塊全為 0 卻不算聖者")
	}
	if got.Char.Strength == 0 {
		t.Error("三圍被第二塊蓋掉了 —— 兩塊混在一起了")
	}

	// 第二塊裡任何一個非 0 → 不是聖者。
	for i := 0; i < U4VirtueCount; i++ {
		raw := makeU4Save(t, func(_, v []byte) {})
		binary.LittleEndian.PutUint16(raw[u4VirtueOffset+u4VirtueFirst+i*2:], 1)
		got, err := ParseU4Transfer(raw)
		if err != nil {
			t.Fatal(err)
		}
		if got.Avatar {
			t.Errorf("第 %d 個 word 是 1 卻還算聖者", i)
		}
	}
}

// 檔案太短要回錯,不能 panic。
func TestU4TransferRejectsShortFiles(t *testing.T) {
	for _, n := range []int{0, 1, u4CharOffset, u4CharOffset + u4CharSize - 1} {
		if _, err := ParseU4Transfer(make([]byte, n)); err == nil {
			t.Errorf("%d B 的檔案竟然轉得進來", n)
		}
	}
	// 只有第一塊、沒有第二塊 → 角色轉得進來,但不是聖者。
	raw := makeU4Save(t, nil)[:u4CharOffset+u4CharSize]
	got, err := ParseU4Transfer(raw)
	if err != nil {
		t.Fatalf("只有第一塊卻轉不進來:%v", err)
	}
	if got.Avatar {
		t.Error("沒有第二塊卻算成聖者")
	}
}

// 三圍換算曲線的三段各驗兩點,加上兩個轉折點(原版 `sub_7564`)。
//
// ⚠ 轉折點才是關鍵:9→9 與 10→10 看起來一樣,但走的是不同的分支;
// 29 與 30 也是。只驗中間值的話,分段條件寫成 `<=` 或 `>` 都會過。
func TestU4StatCurveHasThreeSegments(t *testing.T) {
	cases := map[int]int{
		0: 0, 5: 5, 9: 9, // 第一段:原樣
		10: 10, 11: 11, 20: 15, 29: 20, // 第二段:(v−9)/2 + 10
		30: 20, 34: 21, 50: 25, 70: 30, // 第三段:(v−30)/4 + 20
	}
	for in, want := range cases {
		if got := U4TransferStat(in); got != want {
			t.Errorf("U4TransferStat(%d) = %d,預期 %d", in, got, want)
		}
	}
	// 曲線必須單調不減 —— 分段接不上就會有一段往回掉。
	prev := -1
	for v := 0; v <= U4TransferStatMax; v++ {
		got := U4TransferStat(v)
		if got < prev {
			t.Fatalf("v = %d 時曲線往回掉(%d → %d)", v, prev, got)
		}
		prev = got
	}
}

// 等級是「經驗/100 反覆折半」,而且最高只到 5(U4 經驗上限算出來的)。
func TestU4LevelIsHalvedFromExperience(t *testing.T) {
	cases := map[int]int{
		0: 1, 99: 1, // 經驗/100 == 0 → 1 級
		100: 2, 199: 2,
		200: 3, 399: 3,
		400: 4, 799: 4,
		800: 5,
		999: 5, // U4 經驗上限 9999 / 10 = 999
	}
	for exp, want := range cases {
		if got := U4TransferLevel(exp); got != want {
			t.Errorf("U4TransferLevel(%d) = %d,預期 %d", exp, got, want)
		}
	}
	// U4 的經驗上限換算完最高就是 5 級 —— **算出來的**上限,不是寫死的。
	if got := U4TransferLevel(U4TransferBigMax / U4TransferExpDivisor); got != 5 {
		t.Errorf("U4 經驗上限換算出 %d 級,預期 5", got)
	}
}

// 整個第二階段:經驗 /10、等級由它算、HP = 等級 × 30、三圍走曲線、法力 = 智力。
//
// ⚠ **只有力量有下限 20** —— 敏捷與智力沒有。這條不對稱是原版的,
// 兩者寫成一樣就錯了(組語裡 `cmp al, 14h` 只出現在力量那一段)。
func TestU4ConvertRunsTheSecondStage(t *testing.T) {
	raw := makeU4Save(t, func(r, _ []byte) {
		binary.LittleEndian.PutUint16(r[u4Exp:], 4000)
		binary.LittleEndian.PutUint16(r[u4Str:], 12) // 曲線 → 11,低於 20 → 夾成 20
		binary.LittleEndian.PutUint16(r[u4Dex:], 12) // 曲線 → 11,**不夾**
		binary.LittleEndian.PutUint16(r[u4Intel:], 50)
	})
	got, err := ParseU4Transfer(raw)
	if err != nil {
		t.Fatal(err)
	}
	// 第一階段:原樣搬進來。
	if got.Char.Exp != 4000 || got.Char.Strength != 12 {
		t.Fatalf("第一階段就不對:經驗 %d 力量 %d", got.Char.Exp, got.Char.Strength)
	}
	got.Convert()
	c := &got.Char
	if c.Exp != 400 {
		t.Errorf("經驗換算成 %d,預期 400(/10)", c.Exp)
	}
	if c.Level != 4 {
		t.Errorf("等級是 %d,預期 4(400/100 = 4 → 折半三次)", c.Level)
	}
	if c.HP != 120 || c.MaxHP != 120 {
		t.Errorf("生命是 %d/%d,預期 120(4 級 × 30)", c.HP, c.MaxHP)
	}
	if c.Strength != U4TransferStrengthMin {
		t.Errorf("力量是 %d,預期夾到下限 %d", c.Strength, U4TransferStrengthMin)
	}
	if c.Dex != 11 {
		t.Errorf("敏捷是 %d,預期 11(曲線值,**不夾**)", c.Dex)
	}
	if c.Intel != 25 || c.MP != 25 {
		t.Errorf("智力 / 法力是 %d / %d,預期都是 25", c.Intel, c.MP)
	}
	// Raw 要跟著換,否則存檔寫回換算前的值。
	if c.Raw[CharLevel] != c.Level || c.Raw[CharStrength] != c.Strength {
		t.Error("Raw 沒有跟著換算")
	}
	if binary.LittleEndian.Uint16(c.Raw[CharMaxHP:]) != c.MaxHP {
		t.Error("Raw 的最大 HP 沒跟著換")
	}
}
