package u5data

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
)

// 從 Ultima IV 轉入角色(原版 `sub_71D0`)
//
// 主選單的「Transfer」讀的是 **U4 磁碟上的 `PARTY.SAV`**(原版寫死 `a:party.sav`),
// 而且讀**兩塊**:
//
//	offset 0x008,39 B  → 主角的角色記錄
//	offset 0x140,182 B → 其中八個 u16 決定「是不是聖者」
//
// ★ 兩塊是**兩次獨立的 read**,而且第二塊那八個 word 與第一塊的三圍**位移相同**
//(都是 +6/+8/+0x0A/…)—— 我一開始把兩者混在一起,結論就變成
//「三圍全為 0 才是聖者」,而那對任何合法角色都不成立。
// 分辨的依據是 `mov eax, [ebp+var_4]`,而 `var_4 = offset dword_54328`(第二塊)。
//
// # U4 的角色記錄(offset 0x008 起,39 B)
//
//	+0x00  u16  HP
//	+0x02  u16  最大 HP
//	+0x04  u16  經驗值
//	+0x06  u16  力量
//	+0x08  u16  敏捷
//	+0x0A  u16  智力
//	+0x0C  u16  法力
//	+0x14  8 B  名字(NUL 結尾)
//	+0x24  1 B  性別(0x0B 男,其餘女)
//	+0x25  1 B  職業(0..7)
//
// ⚠ 三圍與 HP 在 U4 存檔裡是 **u16**,搬進 U5 只取**低位元組**
//(原版 `mov al, [edi+6]; mov [esi+0Ch], al`)。所以驗證那一關擋在 70 / 9999
// 不只是「合理範圍」——它同時保證截成一個位元組不會失真。

// U4TransferFile 是原版寫死的檔名(`a:party.sav`)。
//
// 現代環境不會有 A 磁碟,所以引擎讓呼叫端給路徑;檔名保留原樣供辨識。
const U4TransferFile = "PARTY.SAV"

// U4 存檔裡兩塊資料的位移與長度(原版兩次 `sub_2C740`)。
const (
	u4CharOffset   = 0x008
	u4CharSize     = 0x27
	u4VirtueOffset = 0x140
	u4VirtueSize   = 0xB6
)

// U4 角色記錄裡的欄位位移。
const (
	u4HP      = 0x00
	u4MaxHP   = 0x02
	u4Exp     = 0x04
	u4Str     = 0x06
	u4Dex     = 0x08
	u4Intel   = 0x0A
	u4MP      = 0x0C
	u4Name    = 0x14
	u4NameLen = 8
	u4Gender  = 0x24
	u4Class   = 0x25
)

// 驗證的上限(原版 `cmp …, 46h` / `cmp …, 270Fh` / `cmp …, 7`)。
const (
	// U4TransferStatMax 是三圍的上限 70。
	U4TransferStatMax = 0x46
	// U4TransferBigMax 是 HP / 最大 HP / 經驗值的上限 9999。
	U4TransferBigMax = 0x270F
	// U4TransferClassMax 是職業碼的上限 7。
	U4TransferClassMax = 7
)

// U4Classes 是 U4 八個職業對應的 U5 職業字母(原版 `jpt_7349` 八路)。
//
// ★ M(age) B(ard) F(ighter) D(ruid) T(inker) P(aladin) R(anger) S(hepherd)
// —— 正是 U4 的八個職業,順序也一樣。這是這張表對得上的最強證據:
// 八個字母全部命中,而且與八德的職業順序一致。
//
// ⚠ U5 自己的職業只有 A / F / B / M 四種(六名初始角色驗過),
// 但轉入會把 D / T / P / R / S 也**原樣寫進去** —— 原版沒有把它們摺到四種。
var U4Classes = [8]byte{'M', 'B', 'F', 'D', 'T', 'P', 'R', 'S'}

// U4VirtueCount 是「是不是聖者」要檢查的 u16 個數(原版八個 `cmp word ptr`)。
const U4VirtueCount = 8

// u4VirtueFirst 是第一個要檢查的 u16 在第二塊裡的位移。
const u4VirtueFirst = 6

// U4Transfer 是一次轉入的結果。
type U4Transfer struct {
	// Char 是換算好的 U5 角色。
	Char Character
	// Avatar 為真代表 Ztats 要印「乃聖者」(原版 `dword_54498`)。
	Avatar bool
}

// ErrU4Transfer 是「這份 U4 存檔轉不進來」。
//
// 原版對所有失敗印同一句 `Error: Your Ultima IV game contains …` +
// `Unable to continue transfer.` —— 不分原因。這裡把原因帶在錯誤訊息裡
// 給開發用,但**玩家看到的那一句照原版只有一種**。
var ErrU4Transfer = errors.New("這份 Ultima IV 存檔轉不進來")

// LoadU4Transfer 讀一份 U4 `PARTY.SAV` 並換算成 U5 角色。
func LoadU4Transfer(path string) (*U4Transfer, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseU4Transfer(raw)
}

// ParseU4Transfer 解析 U4 `PARTY.SAV` 的內容。
func ParseU4Transfer(raw []byte) (*U4Transfer, error) {
	if len(raw) < u4CharOffset+u4CharSize {
		return nil, fmt.Errorf("%w:只有 %d B,放不下 0x%X 起的角色記錄",
			ErrU4Transfer, len(raw), u4CharOffset)
	}
	rec := raw[u4CharOffset : u4CharOffset+u4CharSize]
	u16 := func(off int) int { return int(binary.LittleEndian.Uint16(rec[off:])) }

	// 驗證照原版的順序與界線。
	for _, f := range []struct {
		off  int
		max  int
		name string
	}{
		{u4Str, U4TransferStatMax, "力量"},
		{u4Dex, U4TransferStatMax, "敏捷"},
		{u4Intel, U4TransferStatMax, "智力"},
		{u4Exp, U4TransferBigMax, "經驗值"},
		{u4HP, U4TransferBigMax, "生命"},
		{u4MaxHP, U4TransferBigMax, "最大生命"},
	} {
		if v := u16(f.off); v > f.max {
			return nil, fmt.Errorf("%w:%s是 %d,超過 %d", ErrU4Transfer, f.name, v, f.max)
		}
	}
	if cls := int(rec[u4Class]); cls > U4TransferClassMax {
		return nil, fmt.Errorf("%w:職業碼 %d 超過 %d", ErrU4Transfer, cls, U4TransferClassMax)
	}
	// 名字裡不能有控制字元(原版 `cmp al, 20h; jnb ok`)。
	// ⚠ NUL 是結尾不是錯誤 —— 原版對 0 是 continue,不是 error。
	name := make([]byte, 0, u4NameLen)
	for i := 0; i < u4NameLen; i++ {
		c := rec[u4Name+i]
		if c == 0 {
			break
		}
		if c < 0x20 {
			return nil, fmt.Errorf("%w:名字第 %d 個位元組是 0x%02X", ErrU4Transfer, i, c)
		}
		name = append(name, c)
	}

	t := &U4Transfer{}
	c := &t.Char
	c.Name = string(name)
	c.Gender = GenderFemale
	if rec[u4Gender] == GenderMale {
		c.Gender = GenderMale
	}
	c.Class = U4Classes[rec[u4Class]]
	c.Status = StatusGood
	// ⚠ 三圍與法力只取**低位元組** —— 原版是 `mov al, [edi+6]`。
	c.Strength = byte(u16(u4Str))
	c.Dex = byte(u16(u4Dex))
	c.Intel = byte(u16(u4Intel))
	c.MP = byte(u16(u4MP))
	c.HP = uint16(u16(u4HP))
	c.MaxHP = uint16(u16(u4MaxHP))
	c.Exp = uint16(u16(u4Exp))
	// ★ **等級 = 最大 HP / 100**(原版 `movzx eax, word ptr [edi+2]; idiv 100`)。
	// 不是照經驗值算 —— U5 平時用的 `LevelForExp` 在這裡**沒有**被用到。
	c.Level = byte(u16(u4MaxHP) / 100)

	// 寫進 32 B 的原始記錄,讓存檔與其他讀 Raw 的程式一致。
	c.Raw = [CharRecordSize]byte{}
	copy(c.Raw[CharName:], name)
	c.Raw[CharGender] = c.Gender
	c.Raw[CharClass] = c.Class
	c.Raw[CharStatus] = c.Status
	c.Raw[CharStrength] = c.Strength
	c.Raw[CharDex] = c.Dex
	c.Raw[CharIntel] = c.Intel
	c.Raw[CharMP] = c.MP
	binary.LittleEndian.PutUint16(c.Raw[CharHP:], c.HP)
	binary.LittleEndian.PutUint16(c.Raw[CharMaxHP:], c.MaxHP)
	binary.LittleEndian.PutUint16(c.Raw[CharExp:], c.Exp)
	c.Raw[CharLevel] = c.Level

	t.Avatar = u4IsAvatar(raw)
	return t, nil
}

// u4IsAvatar 讀第二塊,判斷要不要設「乃聖者」旗標。
//
// 原版:`sub_2C740("a:party.sav", dword_54328, 0xB6, 0x140)`,然後檢查
// `word[+6] .. word[+0x14]` 共八個是否**全為 0**。
//
// ⚠ 讀不到第二塊時原版走的是「轉入失敗」那條路;這裡分開處理 ——
// 角色本身已經換算好了,只是旗標當 false。這與原版的差別只在
// 「破損的存檔會不會整份拒絕」,而正常存檔兩者一致。
func u4IsAvatar(raw []byte) bool {
	if len(raw) < u4VirtueOffset+u4VirtueSize {
		return false
	}
	blk := raw[u4VirtueOffset : u4VirtueOffset+u4VirtueSize]
	for i := 0; i < U4VirtueCount; i++ {
		if binary.LittleEndian.Uint16(blk[u4VirtueFirst+i*2:]) != 0 {
			return false
		}
	}
	return true
}
