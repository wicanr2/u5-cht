package u5data

import "testing"

// ─── 四個吃參數的對話 opcode(`docs/re/79`)────────────────────────────

// render 需要一個 Conversation;字典可以是 nil(這幾條測試不用詞典)。
func opsConv() *Conversation { return &Conversation{} }

// TestDemandGoldParsesThreeAsciiDigits —— ★ 金額寫在對話文字裡。
//
// `sub_1B854`:`(b1&7Fh −'0')×100 + (b2&7Fh −'0')×10 + (b3&7Fh −'0')`。
// `.TLK` 的文字最高位被設起來,所以每個位元組都要先 `& 0x7F`。
func TestDemandGoldParsesThreeAsciiDigits(t *testing.T) {
	cases := []struct {
		digits string
		want   int
	}{
		{"100", 100},
		{"050", 50},
		{"999", 999},
		{"000", 0},
	}
	for _, c := range cases {
		raw := []byte{OpDemandGold}
		for i := 0; i < 3; i++ {
			raw = append(raw, c.digits[i]|0x80) // 照 .TLK 的樣子把最高位設起來
		}
		_, fx := opsConv().render(raw)
		if !fx.Demands {
			t.Errorf("%q 沒被認出是索取金幣", c.digits)
			continue
		}
		if fx.DemandGold != c.want {
			t.Errorf("%q 解成 %d,預期 %d", c.digits, fx.DemandGold, c.want)
		}
	}
}

// TestArgumentBytesDoNotLeakIntoTheText —— ★★ 這是原本真正的症狀。
//
// 原本的展開器用 `for _, ch := range raw`,沒辦法跳過參數位元組 ⇒
// 0x85 本身落進 `ch >= DictLiteralMin` 被印成控制字元 0x05,
// 而三個 ASCII 數字(0x30..0x39)都小於 0x81 → 被當成**詞典索引**查表,
// 於是「索取 100 金幣」會展開成三個不相干的常用詞。
func TestArgumentBytesDoNotLeakIntoTheText(t *testing.T) {
	raw := []byte{OpDemandGold, '1' | 0x80, '0' | 0x80, '0' | 0x80}
	text, fx := opsConv().render(raw)
	if text != "" {
		t.Errorf("參數位元組漏進文字了:%q", text)
	}
	if fx.DemandGold != 100 {
		t.Errorf("金額是 %d", fx.DemandGold)
	}

	// 四個 opcode 都不准把參數印出來。
	for _, c := range []struct {
		name string
		raw  []byte
	}{
		{"0x86 給東西", []byte{OpGiveThing, GiveGold | 0x80}},
		{"0x8C 認得跳轉", []byte{OpJumpIfKnown, 0x83, 0x05}},
		{"0xFE 業報跳轉", []byte{OpJumpIfKarma, 0x1E, 0x07}},
	} {
		if text, _ := opsConv().render(c.raw); text != "" {
			t.Errorf("%s 的參數漏進文字了:%q", c.name, text)
		}
	}
}

// TestGiveThingSplitsAtSixtyFour —— 參數 < 0x40 是裝備索引,'A'..'K' 是資源。
func TestGiveThingSplitsAtSixtyFour(t *testing.T) {
	// 裝備:參數就是背包索引。
	_, fx := opsConv().render([]byte{OpGiveThing, 0x17 | 0x80})
	if !fx.Gives || fx.GiveThing != 0x17 {
		t.Errorf("裝備 0x17 解成 %#x(gives=%v)", fx.GiveThing, fx.Gives)
	}
	if fx.GiveThing >= GiveEquipmentMax {
		t.Error("0x17 該落在裝備那一側")
	}
	// 十一種資源的代碼要連續、而且全部 >= 0x40。
	codes := []byte{GiveFood, GiveGold, GiveKeys, GiveGems, GiveTorches,
		GiveGrapple, GiveCarpet, GiveSextant, GiveSpyglass, GiveBadge, GiveSkullKey}
	for i, c := range codes {
		if c < GiveEquipmentMax {
			t.Errorf("資源代碼 %#x 落在裝備那一側", c)
		}
		if want := byte('A' + i); c != want {
			t.Errorf("第 %d 個資源代碼是 %q,預期 %q(原版 switch 是連續的)",
				i, string(c), string(want))
		}
	}
	if len(codes) != 11 {
		t.Errorf("有 %d 種資源,原版跳表是 11 case", len(codes))
	}
}

// TestBothConditionalJumpsCarryTwoArguments —— 0x8C 與 0xFE 各吃兩個位元組。
func TestBothConditionalJumpsCarryTwoArguments(t *testing.T) {
	_, fx := opsConv().render([]byte{OpJumpIfKnown, 0x03 | 0x80, 0x09})
	if !fx.HasKnownJump || fx.JumpIfKnownBit != 3 || fx.JumpIfKnownTo != 9 {
		t.Errorf("0x8C 解成 bit=%d to=%d(has=%v)",
			fx.JumpIfKnownBit, fx.JumpIfKnownTo, fx.HasKnownJump)
	}
	// ★ 0xFF 是「條件成立就直接繼續」,不是跳到第 255 則。
	_, fx = opsConv().render([]byte{OpJumpIfKnown, 0x01 | 0x80, 0xFF})
	if fx.JumpIfKnownTo != 0xFF {
		t.Errorf("0xFF 沒保留下來:%d", fx.JumpIfKnownTo)
	}

	_, fx = opsConv().render([]byte{OpJumpIfKarma, 0x32 | 0x80, 0x04})
	if !fx.HasKarmaJump || fx.JumpIfKarmaAt != 0x32 || fx.JumpIfKarmaTo != 4 {
		t.Errorf("0xFE 解成 at=%d to=%d(has=%v)",
			fx.JumpIfKarmaAt, fx.JumpIfKarmaTo, fx.HasKarmaJump)
	}
}

// TestTruncatedOpcodeArgumentsDoNotPanic —— 位元組被截斷時不要當掉。
//
// `.TLK` 是原版資料,理論上不會截斷 —— 但解析器對壞資料要收得住,
// 不然一份損壞的存檔或翻譯覆蓋層就會讓遊戲崩潰。
func TestTruncatedOpcodeArgumentsDoNotPanic(t *testing.T) {
	for _, raw := range [][]byte{
		{OpDemandGold},
		{OpDemandGold, '1' | 0x80},
		{OpGiveThing},
		{OpJumpIfKnown, 0x01},
		{OpJumpIfKarma},
	} {
		if text, _ := opsConv().render(raw); text != "" {
			t.Errorf("%v 的殘餘位元組漏進文字:%q", raw, text)
		}
	}
}
