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
