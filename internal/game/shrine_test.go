package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

func shrineState(t *testing.T) *State {
	t.Helper()
	s := introState(t)
	if s == nil {
		return nil
	}
	tf, err := u5data.LoadText(gameDataDir(t) + "/MISCMSG.DAT")
	if err != nil {
		t.Fatal(err)
	}
	s.Misc = tf
	s.Location = 0
	return s
}

// atShrine 把隊伍放到第 v 座聖壇上並開始冥想。
func atShrine(t *testing.T, v int) *State {
	t.Helper()
	s := shrineState(t)
	if s == nil {
		return nil
	}
	s.X, s.Y = u5data.Shrines[v].X, u5data.Shrines[v].Y
	if !s.Meditate() {
		t.Fatal("拜不了")
	}
	if s.Shrine.Virtue != v {
		t.Fatalf("站在第 %d 座卻認成第 %d 座", v, s.Shrine.Virtue)
	}
	return s
}

// answer 一路把美德名與三次真言打對。
func answer(s *State, v int) {
	s.ShrineAnswer(u5data.Shrines[v].Name)
	for i := 0; i < ShrineMantraTries; i++ {
		s.ShrineAnswer(u5data.Shrines[v].Mantra)
	}
}

// TestShrineCoordinatesAreDistinct:八座聖壇不能撞在同一格。
//
// ⚠ **靈性那座的座標是 (0,0)**,那不是資料缺漏 —— 原版就沒把它放進表裡,
// 靠「掃不到就當靈性」的 fallback 生效(U5 的靈性聖壇在幽冥界)。
// 所以這條只檢查其餘七座互不相同。
func TestShrineCoordinatesAreDistinct(t *testing.T) {
	seen := map[[2]int]string{}
	for i, sh := range u5data.Shrines {
		if i == u5data.VirtueSpirituality {
			if sh.X != 0 || sh.Y != 0 {
				t.Errorf("靈性聖壇的座標是 (%d,%d),原版是 (0,0)", sh.X, sh.Y)
			}
			continue
		}
		k := [2]int{sh.X, sh.Y}
		if prev, dup := seen[k]; dup {
			t.Errorf("%s 與 %s 都在 (%d,%d)", prev, sh.Name, sh.X, sh.Y)
		}
		seen[k] = sh.Name
	}
	// 掃不到就是靈性。
	if got := u5data.ShrineAt(1, 1); got != u5data.VirtueSpirituality {
		t.Errorf("不在表上的座標認成第 %d 座,應該是靈性", got)
	}
}

// TestVirtuePrefixesAreDistinguishable:八個四字母前綴不能互相包含。
//
// ⚠ `hone`(誠實)與 `hono`(榮譽)只差第四個字母 —— 這正是原版比四個字母
// 而不是三個的原因。若哪天有人「順手」縮成三個,這條會紅。
func TestVirtuePrefixesAreDistinguishable(t *testing.T) {
	for i, a := range u5data.Shrines {
		if len(a.Prefix) != 4 {
			t.Errorf("%s 的前綴 %q 不是四個字母", a.Name, a.Prefix)
		}
		if !strings.HasPrefix(strings.ToLower(a.Name), a.Prefix) {
			t.Errorf("%s 的前綴 %q 對不上名字", a.Name, a.Prefix)
		}
		for j, b := range u5data.Shrines {
			if i != j && a.Prefix == b.Prefix {
				t.Errorf("%s 與 %s 的前綴都是 %q", a.Name, b.Name, a.Prefix)
			}
		}
	}
	// 打「honor」不該通過誠實聖壇,反之亦然。
	if strings.Contains("honor", u5data.Shrines[u5data.VirtueHonesty].Prefix) {
		t.Error("在誠實聖壇打 honor 會過")
	}
	if strings.Contains("honesty", u5data.Shrines[u5data.VirtueHonor].Prefix) {
		t.Error("在榮譽聖壇打 honesty 會過")
	}
}

// TestMantrasAreDistinct:八個真言互不相同。
func TestMantrasAreDistinct(t *testing.T) {
	seen := map[string]string{}
	for _, sh := range u5data.Shrines {
		if prev, dup := seen[strings.ToLower(sh.Mantra)]; dup {
			t.Errorf("%s 與 %s 的真言都是 %q", prev, sh.Name, sh.Mantra)
		}
		seen[strings.ToLower(sh.Mantra)] = sh.Name
	}
}

// TestShrineFirstVisitGivesQuest:第一次拜會領到試煉,而且不會直接給獎。
func TestShrineFirstVisitGivesQuest(t *testing.T) {
	s := atShrine(t, u5data.VirtueValor)
	if s == nil {
		return
	}
	before := s.Roster[0].Strength
	answer(s, u5data.VirtueValor)
	bit := byte(1) << u5data.VirtueValor
	if s.ShrineQuestGiven&bit == 0 || s.ShrineQuestActive&bit == 0 {
		t.Errorf("領完試煉之後旗標是 given=%02X active=%02X",
			s.ShrineQuestGiven, s.ShrineQuestActive)
	}
	if s.Roster[0].Strength != before {
		t.Error("第一次拜就加了力量 —— 獎賞應該要完成試煉才給")
	}
	if !strings.Contains(s.log(), "勇氣") && !strings.Contains(s.log(), "試煉") {
		t.Errorf("沒有印出試煉內容:\n%s", s.log())
	}
}

// TestShrineRewardNeedsTheQuest:完成試煉才給三圍與業報,而且只給一次。
//
// ⚠ 兩組旗標缺一不可。只留一個的話玩家可以站在原地無限重領。
func TestShrineRewardNeedsTheQuest(t *testing.T) {
	s := atShrine(t, u5data.VirtueValor)
	if s == nil {
		return
	}
	answer(s, u5data.VirtueValor) // 第一次:領試煉
	s.Roster[0].Strength = 10
	s.Karma = 50

	s.Meditate()
	answer(s, u5data.VirtueValor) // 第二次:領獎
	if s.Roster[0].Strength != 11 {
		t.Errorf("力量 %d,勇氣聖壇該加到 11", s.Roster[0].Strength)
	}
	if s.Karma != 53 {
		t.Errorf("業報 %d,完成試煉該 +3 變成 53", s.Karma)
	}

	// 第三次:試煉旗標已清掉,只能捐錢,不能再拿三圍。
	s.Inventory.Gold = 0
	s.Meditate()
	answer(s, u5data.VirtueValor)
	if s.Roster[0].Strength != 11 {
		t.Errorf("第三次又加了力量(變成 %d)—— 獎賞被重領了", s.Roster[0].Strength)
	}
	if s.Shrine == nil || s.Shrine.Stage != ShrineAskOffer {
		t.Error("第三次應該進到獻金,而不是結束")
	}
}

// TestWrongMantraFails:三次真言只要錯一次就失敗。
//
// ⚠ **不是「三次機會」。** 原版把三次都問完才判(`var_4`),
// 任何一次錯就整場失敗。寫成三次機會會讓聖壇好過太多。
func TestWrongMantraFails(t *testing.T) {
	for _, wrongAt := range []int{0, 1, 2} {
		s := atShrine(t, u5data.VirtueHonesty)
		if s == nil {
			return
		}
		s.ShrineAnswer("honesty")
		for i := 0; i < ShrineMantraTries; i++ {
			if i == wrongAt {
				s.ShrineAnswer("Nope")
			} else {
				s.ShrineAnswer("Ahm")
			}
		}
		if s.ShrineQuestGiven != 0 {
			t.Errorf("第 %d 次打錯卻還是領到試煉", wrongAt)
		}
		if !strings.Contains(s.log(), "散亂") {
			t.Errorf("第 %d 次打錯沒有印出失敗訊息:\n%s", wrongAt, s.log())
		}
	}
}

// TestWrongVirtueFails:美德名打錯也一樣不算。
func TestWrongVirtueFails(t *testing.T) {
	s := atShrine(t, u5data.VirtueHonesty)
	if s == nil {
		return
	}
	s.ShrineAnswer("honor") // 誠實聖壇打榮譽
	for i := 0; i < ShrineMantraTries; i++ {
		s.ShrineAnswer("Ahm")
	}
	if s.ShrineQuestGiven != 0 {
		t.Error("美德名打錯卻還是領到試煉")
	}
}

// TestOfferConvertsGoldToKarma:每重 100 金換 1 點業報,而且一次最多 9 重。
//
// ⚠ 原版只讀**一個**按鍵,所以「一次最多 900 金」不是我加的限制。
// 讀成多位數會讓玩家一口氣把業報衝滿。
func TestOfferConvertsGoldToKarma(t *testing.T) {
	s := atShrine(t, u5data.VirtueValor)
	if s == nil {
		return
	}
	answer(s, u5data.VirtueValor)
	s.Meditate()
	answer(s, u5data.VirtueValor) // 領獎
	s.Meditate()
	answer(s, u5data.VirtueValor) // 進到獻金
	if s.Shrine == nil || s.Shrine.Stage != ShrineAskOffer {
		t.Fatalf("沒有進到獻金:\n%s", s.log())
	}
	s.Inventory.Gold = 1000
	s.Karma = 10
	s.ShrineAnswer("3")
	if s.Inventory.Gold != 700 {
		t.Errorf("金子剩 %d,3 重應該扣 300", s.Inventory.Gold)
	}
	if s.Karma != 13 {
		t.Errorf("業報 %d,3 重應該 +3", s.Karma)
	}

	// 金子不夠時不扣、不加,而且留在獻金畫面。
	s.Meditate()
	answer(s, u5data.VirtueValor)
	s.Inventory.Gold = 50
	s.Karma = 10
	s.ShrineAnswer("9")
	if s.Inventory.Gold != 50 || s.Karma != 10 {
		t.Errorf("金子不夠卻扣了:金 %d、業報 %d", s.Inventory.Gold, s.Karma)
	}
	if s.Shrine == nil {
		t.Error("金子不夠應該再問一次,不是結束")
	}
}

// TestHumilityGivesDoubleKarma:謙遜不加三圍,改成雙倍業報。
//
// 八德裡只有這一座三圍全不加(`byte_5606C/56074/5607C` 第 7 筆都是 0),
// 原版用 `cmp edi, 7` 額外再 +3 補回來。
func TestHumilityGivesDoubleKarma(t *testing.T) {
	sh := u5data.Shrines[u5data.VirtueHumility]
	if sh.Str || sh.Dex || sh.Int {
		t.Fatal("謙遜聖壇不該加三圍")
	}
	s := atShrine(t, u5data.VirtueHumility)
	if s == nil {
		return
	}
	answer(s, u5data.VirtueHumility)
	s.Meditate()
	s.Karma = 10
	answer(s, u5data.VirtueHumility)
	if s.Karma != 16 {
		t.Errorf("業報 %d,謙遜應該 +3+3 變成 16", s.Karma)
	}
}

// TestKarmaCapsAt99:業報上限是 99 不是 100。
func TestKarmaCapsAt99(t *testing.T) {
	s := shrineState(t)
	if s == nil {
		return
	}
	s.Karma = 95
	s.AddKarma(20)
	if s.Karma != u5data.KarmaMax {
		t.Errorf("業報 %d,上限應該是 %d", s.Karma, u5data.KarmaMax)
	}
}
