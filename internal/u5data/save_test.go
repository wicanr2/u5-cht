package u5data

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSaveRejectsWrongSize(t *testing.T) {
	if _, err := ParseSave(make([]byte, 100)); err == nil {
		t.Error("大小不對卻沒報錯")
	}
}

// TestRealSaves 用原版的 SAVED.GAM 與 INIT.GAM 驗證欄位位移。
//
// 位移是把讀取序列的欄位大小累加算出來的,漏算一個欄位後面全會偏 ——
// 而偏掉之後讀出來仍然是「某個數字」,不會自己報錯。所以這裡靠**內部一致性**
// 來抓:HP 恰好等於等級 × 30、職業是可讀字母、日期在曆法範圍內。
func TestRealSaves(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	for _, name := range []string{"SAVED.GAM", "INIT.GAM"} {
		sv, err := LoadSave(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("%s:%v", name, err)
		}
		chars := 0
		for i := range sv.Roster {
			c := &sv.Roster[i]
			if !c.Present() {
				continue
			}
			chars++
			switch c.Class {
			case 'A', 'F', 'B', 'M', 'T':
			default:
				t.Errorf("%s #%d(%s)的職業碼是 0x%02X,不是可讀字母 —— 位移錯了",
					name, i, c.Name, c.Class)
			}
			if c.Gender != 0x0B && c.Gender != 0x0C {
				t.Errorf("%s #%d(%s)的性別碼是 0x%02X,預期 0x0B/0x0C",
					name, i, c.Name, c.Gender)
			}
			// 最大 HP = 等級 × 30。六名初始角色全部符合,是位移正確的內部佐證。
			if want := uint16(c.Level) * 30; c.MaxHP != want {
				t.Errorf("%s #%d(%s)Lv%d 的最大 HP 是 %d,預期 %d",
					name, i, c.Name, c.Level, c.MaxHP, want)
			}
			if c.HP > c.MaxHP {
				t.Errorf("%s #%d(%s)目前 HP %d 超過上限 %d", name, i, c.Name, c.HP, c.MaxHP)
			}
		}
		if chars < 6 {
			t.Errorf("%s 只解出 %d 名角色", name, chars)
		}
		if sv.Year != 139 {
			t.Errorf("%s 的年份是 %d,原版是不列顛尼亞 139 年", name, sv.Year)
		}
		t.Logf("%s:%d 名角色,%d 年 %d 月 %d 日 %02d:%02d,隊伍 %d 人,業報 %d",
			name, chars, sv.Year, sv.Month, sv.Day, sv.Hour, sv.Minute, sv.PartySize, sv.Karma)
	}
}

// TestKnownRoster 固定 U5 的正典夥伴名單 —— 名字讀錯就是名冊位移錯了。
func TestKnownRoster(t *testing.T) {
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	sv, err := LoadSave(filepath.Join(dir, "INIT.GAM"))
	if err != nil {
		t.Fatal(err)
	}
	// INIT.GAM 的 0 號是空名(等玩家命名),1 號之後是固定夥伴。
	want := []string{"Shamino", "Iolo", "Mariah", "Geoffrey", "Jaana"}
	for i, w := range want {
		if got := sv.Roster[i+1].Name; got != w {
			t.Errorf("名冊 #%d 是 %q,預期 %q", i+1, got, w)
		}
	}
	if sv.Roster[0].Name != "" {
		t.Errorf("INIT.GAM 的 0 號應該是空名(新遊戲讓玩家命名),實得 %q", sv.Roster[0].Name)
	}
	if sv.Roster[0].Class != 'A' {
		t.Errorf("0 號的職業是 0x%02X,預期 'A'(聖者)", sv.Roster[0].Class)
	}
}
