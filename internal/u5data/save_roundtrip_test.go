package u5data

import (
	"bytes"
	"os"
	"testing"
)

// TestSaveRoundTripIsByteIdentical:讀進來再寫出去,必須與原檔**逐位元組相同**。
//
// 這是存檔寫回最強的驗收:只要有任何一個欄位寫錯位置、或是把還沒解出來的
// 區段清成 0,這裡就會炸。反過來說,通過了就代表寫出去的檔拿回原版也讀得起來。
func TestSaveRoundTripIsByteIdentical(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	for _, name := range []string{"INIT.GAM", "SAVED.GAM"} {
		raw, err := os.ReadFile(dir + "/" + name)
		if err != nil {
			t.Fatal(err)
		}
		sv, err := ParseSave(raw)
		if err != nil {
			t.Fatalf("%s:%v", name, err)
		}
		out, err := sv.Encode()
		if err != nil {
			t.Fatalf("%s:%v", name, err)
		}
		if !bytes.Equal(raw, out) {
			for i := range raw {
				if raw[i] != out[i] {
					t.Fatalf("%s:第一個不同的位元組在 0x%04X(原 0x%02X → 寫出 0x%02X)",
						name, i, raw[i], out[i])
				}
			}
		}
	}
}

// TestSaveEncodeReflectsEdits:改過的欄位要真的寫進去,而且讀得回來。
func TestSaveEncodeReflectsEdits(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	sv, err := LoadSave(dir + "/INIT.GAM")
	if err != nil {
		t.Fatal(err)
	}
	sv.Inventory.Gold = 1234
	sv.Inventory.Food = 7
	sv.Inventory.Keys = 9
	sv.Inventory.Items[ItemArrows] = 42
	sv.Inventory.Reagents[3] = 11
	sv.Karma = 33
	sv.Hour, sv.Minute = 13, 45
	sv.X, sv.Y = 12, 34
	sv.Roster[0].HP = 7
	sv.Roster[0].Status = StatusPoisoned
	sv.Roster[0].Name = "Bob"

	blob, err := sv.Encode()
	if err != nil {
		t.Fatal(err)
	}
	back, err := ParseSave(blob)
	if err != nil {
		t.Fatalf("寫出來的檔讀不回來:%v", err)
	}
	switch {
	case back.Inventory.Gold != 1234:
		t.Errorf("金幣 %d", back.Inventory.Gold)
	case back.Inventory.Food != 7:
		t.Errorf("存糧 %d", back.Inventory.Food)
	case back.Inventory.Keys != 9:
		t.Errorf("鑰匙 %d", back.Inventory.Keys)
	case back.Inventory.Items[ItemArrows] != 42:
		t.Errorf("箭矢 %d", back.Inventory.Items[ItemArrows])
	case back.Inventory.Reagents[3] != 11:
		t.Errorf("蜘蛛絲 %d", back.Inventory.Reagents[3])
	case back.Karma != 33:
		t.Errorf("業報 %d", back.Karma)
	case back.Hour != 13 || back.Minute != 45:
		t.Errorf("時間 %d:%d", back.Hour, back.Minute)
	case back.X != 12 || back.Y != 34:
		t.Errorf("座標 (%d,%d)", back.X, back.Y)
	case back.Roster[0].HP != 7:
		t.Errorf("HP %d", back.Roster[0].HP)
	case back.Roster[0].Status != StatusPoisoned:
		t.Errorf("狀態 %c", back.Roster[0].Status)
	case back.Roster[0].Name != "Bob":
		t.Errorf("名字 %q", back.Roster[0].Name)
	}
}

// TestObjectRoundTrip:物件表寫出去再讀回來要一致,而且船的耐久與小艇數
// (藏在 Raw 的 +5 / +7)不能掉。
func TestObjectRoundTrip(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	sur, und, err := LoadWorldObjects(dir)
	if err != nil {
		t.Fatal(err)
	}
	slot, ok := sur.Spawn(VehicleShip, 40, 50, 0)
	if !ok {
		t.Fatal("放不下船")
	}
	sur.Objects[slot].SetHull(6)
	sur.Objects[slot].SetSkiffs(3)

	blob := EncodeSaveObjects(sur, und)
	if len(blob) != ObjectFileSize*2 {
		t.Fatalf("寫出 %d B,預期 %d", len(blob), ObjectFileSize*2)
	}
	backSur, err := ParseObjects(blob[:ObjectFileSize])
	if err != nil {
		t.Fatal(err)
	}
	backUnd, err := ParseObjects(blob[ObjectFileSize:])
	if err != nil {
		t.Fatal(err)
	}
	o, found := backSur.At(40, 50, 0)
	if !found {
		t.Fatal("讀回來船不見了")
	}
	if o.Hull() != 6 || o.Skiffs() != 3 {
		t.Errorf("耐久 %d 小艇 %d,預期 6 / 3", o.Hull(), o.Skiffs())
	}
	if *backUnd != *und {
		t.Error("地下世界那半段對不上")
	}
}

// 新解出來的四個欄位:怪鑰匙、卷軸、藥水、月石。
//
// ⚠ 這一條**同時**在驗位移。存檔已有「讀進來寫出去逐位元組相同」的測試,
// 所以只要位移落在別人的欄位上,改動就會破壞那一條 —— 兩條合起來才有意義:
// 一條說「寫得回去」,一條說「寫的是這一格」。
func TestNewInventoryFieldsRoundTrip(t *testing.T) {
	raw := make([]byte, SaveFileSize)
	// 讓 validate() 過得去的最小合理值(月 / 日;其餘 0 就合法)。
	raw[SaveMonthOffset], raw[SaveDayOffset] = 4, 5
	raw[SaveOddKeysOffset] = 7
	for i := 0; i < ScrollCount; i++ {
		raw[SaveScrollsOffset+i] = byte(i + 1)
	}
	for i := 0; i < PotionCount; i++ {
		raw[SavePotionsOffset+i] = byte(10 + i)
	}
	// 月石:第 3 顆還在手上(地點 0xFF),第 5 顆埋在地點 7 的 (11, 22) 第 −1 層。
	raw[SaveMoonstoneLocOffset+3] = MoonstoneInHand
	raw[SaveMoonstoneXOffset+5] = 11
	raw[SaveMoonstoneYOffset+5] = 22
	raw[SaveMoonstoneLocOffset+5] = 7
	raw[SaveMoonstoneFloorOffset+5] = 0xFF

	s, err := ParseSave(raw)
	if err != nil {
		t.Fatal(err)
	}
	if s.Inventory.OddKeys != 7 {
		t.Errorf("怪鑰匙讀成 %d", s.Inventory.OddKeys)
	}
	for i := 0; i < ScrollCount; i++ {
		if s.Inventory.Scrolls[i] != i+1 {
			t.Errorf("卷軸 %d 讀成 %d", i, s.Inventory.Scrolls[i])
		}
	}
	for i := 0; i < PotionCount; i++ {
		if s.Inventory.Potions[i] != 10+i {
			t.Errorf("藥水 %d 讀成 %d", i, s.Inventory.Potions[i])
		}
	}
	if !s.Inventory.Moonstones[3].InHand() {
		t.Errorf("第 3 顆月石該還在手上:%+v", s.Inventory.Moonstones[3])
	}
	// ★ 地點 0 不是「在手上」——「在手上」是 0xFF。第 2 顆沒被寫過,
	// 所以它的地點是 0(大地圖),而**不是**在手上。這一條就是原本那份
	// 「16 顆布林旗標」讀法會讀反的地方。
	if s.Inventory.Moonstones[2].InHand() {
		t.Errorf("第 2 顆月石不該在手上:%+v", s.Inventory.Moonstones[2])
	}
	if got := s.Inventory.Moonstones[5]; got.X != 11 || got.Y != 22 ||
		got.Location != 7 || got.Floor != -1 {
		t.Errorf("第 5 顆月石讀成 %+v", got)
	}

	out, err := s.Encode()
	if err != nil {
		t.Fatal(err)
	}
	for i := range raw {
		if out[i] != raw[i] {
			t.Fatalf("位移 0x%04X 寫回去變了:%02X → %02X", i, raw[i], out[i])
		}
	}
}

// 卷軸 / 藥水 / 月石不能重疊到已驗過的咒語與藥草那兩段。
func TestNewInventoryOffsetsDoNotOverlap(t *testing.T) {
	type span struct {
		name     string
		off, len int
	}
	spans := []span{
		{"咒語", SaveSpellsOffset, SpellCount},
		{"卷軸", SaveScrollsOffset, ScrollCount},
		{"藥水", SavePotionsOffset, PotionCount},
		{"月石 X", SaveMoonstoneXOffset, MoonstoneCount},
		{"月石 Y", SaveMoonstoneYOffset, MoonstoneCount},
		{"月石地點", SaveMoonstoneLocOffset, MoonstoneCount},
		{"月石樓層", SaveMoonstoneFloorOffset, MoonstoneCount},
		{"藥草", SaveReagentsOffset, ReagentCount},
	}
	for i := 1; i < len(spans); i++ {
		prev, cur := spans[i-1], spans[i]
		if prev.off+prev.len > cur.off {
			t.Errorf("%s(0x%04X+%d)壓到%s(0x%04X)",
				prev.name, prev.off, prev.len, cur.name, cur.off)
		}
	}
	// 咒語到藥草之間正好 0x60 —— 與記憶體上 byte_3E000 到 byte_3E060 的距離相同。
	if SaveReagentsOffset-SaveSpellsOffset != 0x60 {
		t.Errorf("咒語到藥草距離 0x%X,應該是 0x60(記憶體佈局的錨點)",
			SaveReagentsOffset-SaveSpellsOffset)
	}
}
