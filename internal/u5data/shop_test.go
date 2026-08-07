package u5data

import (
	"os"
	"strings"
	"testing"
)

func loadShops(t *testing.T) *ShopSet {
	t.Helper()
	dir := os.Getenv("U5_GAMEDATA")
	if dir == "" {
		t.Skip("未設 U5_GAMEDATA,跳過需要原版資料的測試")
	}
	d, err := LoadDictionary(dir)
	if err != nil {
		t.Fatal(err)
	}
	s, err := LoadShops(dir, d)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// TestShopDirectory:47 家店,店名與店主都從 DATA.OVL 讀出來,
// 而且只有一家沒有名字(類型 2 開在地點 30 的馬廄,原版指標就是 NULL)。
func TestShopDirectory(t *testing.T) {
	s := loadShops(t)
	if len(s.Shops) != 47 {
		t.Fatalf("解出 %d 家店,預期 47", len(s.Shops))
	}
	noName := 0
	for i := range s.Shops {
		sh := &s.Shops[i]
		if sh.Owner == "" {
			t.Errorf("第 %d 家(類型 %d 地點 %d)沒有店主", i, sh.Type, sh.Location)
		}
		if sh.Name == "" {
			noName++
			if sh.Type != ShopStable || sh.Location != 30 {
				t.Errorf("第 %d 家沒有店名,但它不是已知的那一家(類型 %d 地點 %d)",
					i, sh.Type, sh.Location)
			}
		}
		if sh.Location < 1 || sh.Location > len(Locations) {
			t.Errorf("第 %d 家的地點編號 %d 不合理", i, sh.Location)
		}
	}
	if noName != 1 {
		t.Errorf("有 %d 家沒有店名,預期正好 1 家", noName)
	}
}

// TestKnownShops 固定幾家對得出來的店 —— 名字錯了就是 DATA.OVL 的位移錯了。
func TestKnownShops(t *testing.T) {
	s := loadShops(t)
	cases := []struct {
		loc          int
		typ          ShopType
		name, owner  string
	}{
		{2, ShopArmoury, "Iolo's Bows", "Gwenneth"},
		{2, ShopTavern, "The Wayfarer Tavern", "Tika"},
		{2, ShopInn, "The Wayfarer Inn", "Donya"},
		{6, ShopStable, "Horse & Rider", "Hettar"},
		{24, ShopInn, "The King's Ransom Inn", "Ransack"},
	}
	for _, c := range cases {
		sh, ok := s.At(c.loc, c.typ)
		if !ok {
			t.Errorf("地點 %d 沒有 %s", c.loc, c.typ.TypeName())
			continue
		}
		if sh.Name != c.name || sh.Owner != c.owner {
			t.Errorf("地點 %d 的 %s 是 %q/%q,預期 %q/%q",
				c.loc, c.typ.TypeName(), sh.Name, sh.Owner, c.name, c.owner)
		}
	}
}

// TestShopGreetingSubstitution:問候語裡的 # $ @ 要被換掉,不能留在畫面上。
func TestShopGreetingSubstitution(t *testing.T) {
	s := loadShops(t)
	sh, ok := s.At(2, ShopTavern)
	if !ok {
		t.Fatal("找不到不列顛城的酒館")
	}
	for v := 0; v < 4; v++ {
		g := s.Greeting(sh, v, 12)
		if g == "" {
			t.Errorf("第 %d 句問候語是空的", v)
			continue
		}
		for _, ph := range []string{"#", "$", "@"} {
			if strings.Contains(g, ph) {
				t.Errorf("第 %d 句還留著佔位符 %s:%q", v, ph, g)
			}
		}
		if !strings.Contains(g, sh.Name) && !strings.Contains(g, sh.Owner) {
			t.Errorf("第 %d 句既沒有店名也沒有店主:%q", v, g)
		}
	}
	// @ 依時間變化
	for _, c := range []struct {
		hour int
		want string
	}{{8, "morning"}, {14, "afternoon"}, {20, "evening"}} {
		if got := TimeOfDay(c.hour); got != c.want {
			t.Errorf("%d 點是 %q,預期 %q", c.hour, got, c.want)
		}
	}
}

// TestEveryShopTypeHasGreetings:除了武具店之外,每種店都要有四句問候語。
//
// 武具店(類型 0)的四個位移都是 0 —— 原版它走另一條流程(還沒解),
// 這裡把它當已知例外記下來,而不是讓測試無聲通過。
func TestEveryShopTypeHasGreetings(t *testing.T) {
	s := loadShops(t)
	for typ := ShopType(0); typ < ShopTypeCount; typ++ {
		var sh *Shop
		for i := range s.Shops {
			if s.Shops[i].Type == typ {
				sh = &s.Shops[i]
				break
			}
		}
		if sh == nil {
			t.Errorf("沒有任何 %s", typ.TypeName())
			continue
		}
		g := s.Greeting(sh, 0, 12)
		if typ == ShopArmoury {
			if g != "" {
				t.Errorf("武具店竟然有問候語 %q —— 原版那四個位移是 0", g)
			}
			continue
		}
		if g == "" {
			t.Errorf("%s 沒有問候語", typ.TypeName())
		}
	}
}
