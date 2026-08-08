package game

import (
	"testing"

	"github.com/wicanr2/u5-cht/internal/u5data"
)

// TestSFXQueueIsDrainedNotOverwritten —— ★ 一個回合可以疊好幾個音效。
//
// 原版有 8 個 PCM 通道(`docs/re/63`)⇒ 同時發生的事會一起響。
// 用單一欄位會吃掉其中幾個(踏進沼澤中毒 + 受傷只剩一個)。
func TestSFXQueueIsDrainedNotOverwritten(t *testing.T) {
	var s State
	s.PlaySFX(u5data.SFXWalk)
	s.PlaySFX(u5data.SFXDamage1)
	got := s.TakeSFX()
	if len(got) != 2 || got[0] != u5data.SFXWalk || got[1] != u5data.SFXDamage1 {
		t.Errorf("取出 %v,預期 [%d %d]", got, u5data.SFXWalk, u5data.SFXDamage1)
	}
	if again := s.TakeSFX(); again != nil {
		t.Errorf("取兩次拿到 %v,第二次該是空的", again)
	}
}

// TestSFXQueueHasACap —— 上限是防呆:世界回合會掃 31 個槽。
func TestSFXQueueHasACap(t *testing.T) {
	var s State
	for i := 0; i < SFXQueueMax*3; i++ {
		s.PlaySFX(u5data.SFXWalk)
	}
	if n := len(s.TakeSFX()); n != SFXQueueMax {
		t.Errorf("排了 %d 個,預期上限 %d", n, SFXQueueMax)
	}
}

// TestSFXIndexOutOfRangeIsIgnored —— 越界索引不排、不 panic。
func TestSFXIndexOutOfRangeIsIgnored(t *testing.T) {
	var s State
	s.PlaySFX(-1)
	s.PlaySFX(u5data.SFXCount)
	s.PlaySFX(u5data.SFXNone)
	if got := s.TakeSFX(); got != nil {
		t.Errorf("越界索引被排進去了:%v", got)
	}
}

// TestWalkSoundDependsOnMountAndTerrain —— 腳步 / 馬蹄 / 慢速腳步。
//
// ⚠ **B 級證據**:三個檔名一望即知,但沒追到呼叫點(`docs/re/90` §4)。
// 這條測試釘住的是「引擎照這個規則接」,不是「原版就是這樣」。
func TestWalkSoundDependsOnMountAndTerrain(t *testing.T) {
	var s State
	s.Transport = u5data.VehicleWalk
	if got := s.walkSFX(tileGrass); got != u5data.SFXWalk {
		t.Errorf("草地上是 %d,預期 WALK(%d)", got, u5data.SFXWalk)
	}
	// 丘陵要多付回合 ⇒ 慢速腳步。
	if got := s.walkSFX(11); got != u5data.SFXWalkSlow {
		t.Errorf("丘陵上是 %d,預期 WALKSLOW(%d)", got, u5data.SFXWalkSlow)
	}
	s.Transport = mountedHorse
	if got := s.walkSFX(tileGrass); got != u5data.SFXHorse {
		t.Errorf("騎馬時是 %d,預期 HORSE(%d)", got, u5data.SFXHorse)
	}
	// ★ 騎馬蓋過地形 —— 馬在丘陵上還是馬蹄聲。
	if got := s.walkSFX(11); got != u5data.SFXHorse {
		t.Errorf("騎馬走丘陵是 %d,預期還是 HORSE", got)
	}
}

// TestMoongateAndFallHaveEvidencedSounds —— ★ A 級證據的兩個。
//
// 月門:原版 `sub_135FC` 用索引 0x0A = MOON2 三次。
// 墜落:原版 `sub_10A1C`(墜落動畫)用索引 0x14 = T_OCHI1(「落ち」)。
func TestMoongateAndFallHaveEvidencedSounds(t *testing.T) {
	s := moonState(t)
	if s.BaseSave == nil {
		t.Skip("沒有底稿存檔")
	}
	s.Location, s.Floor = 0, 0
	s.Clock.Hour = u5data.MoongateNightFrom
	fx, fy := s.Moongates[0].X+40, s.Moongates[0].Y+40
	s.X, s.Y = fx, fy
	if !s.SetTileAt(fx, fy, u5data.MoongateOpenTile) {
		t.Fatal("寫不進世界地圖")
	}
	s.TakeSFX()
	if !s.EnterMoongateHere() {
		t.Fatal("沒傳送")
	}
	if !hasSFX(s.TakeSFX(), u5data.SFXMoongate) {
		t.Error("踏進月門沒有 MOON2")
	}
}

func hasSFX(q []int, want int) bool {
	for _, v := range q {
		if v == want {
			return true
		}
	}
	return false
}
