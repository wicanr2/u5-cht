package game

import "github.com/wicanr2/u5-cht/internal/u5data"

// 音效:引擎只把「該放第幾號」排進佇列,發聲在 `internal/audio`
//
// 分工同配樂(`music.go`):`internal/game` 不碰音訊 ⇒ headless 不需要音效裝置。
// 索引語意與證據等級見 `u5data/sfxindex.go` 與 `docs/re/90`。
//
// ⚠ **佇列而不是「當前音效」**:一個回合可以同時發生好幾件事(踏進沼澤中毒 +
// 受傷),原版有 8 個 PCM 通道可以疊。用單一欄位會吃掉其中幾個。

// SFXQueueMax 是一個回合最多排幾個音效。
//
// ⚠ 上限存在的理由是**防呆而不是原版限制**:世界回合會對 31 個槽跑攻擊判定,
// 若哪天有人在迴圈裡放音,沒有上限就會爆。原版是 8 個 PCM 通道(`docs/re/63`),
// 所以聽得到的本來就有限。
const SFXQueueMax = 16

// PlaySFX 把一個音效排進佇列。
func (s *State) PlaySFX(idx int) {
	if idx < 0 || idx >= u5data.SFXCount || len(s.sfx) >= SFXQueueMax {
		return
	}
	s.sfx = append(s.sfx, idx)
}

// TakeSFX 取出並清空這一輪要放的音效(由 `cmd/u5cht` 每帧呼叫)。
func (s *State) TakeSFX() []int {
	if len(s.sfx) == 0 {
		return nil
	}
	out := s.sfx
	s.sfx = nil
	return out
}

// walkSFX 是「這一步該放哪個腳步聲」。
//
// ⚠ **B 級證據**:三個檔名(`WALK` / `WALKSLOW` / `HORSE`)一望即知,但**沒有
// 追到呼叫點** —— 原版在哪裡放腳步聲、粗糙地形是否真的換 `WALKSLOW`,都還沒查。
// 引擎照名字接,理由與代價寫在 `docs/re/90` §4。
//
//	騎馬(載具碼 0x12/0x13)→ HORSE
//	要多付回合的粗糙地形    → WALKSLOW
//	其餘                    → WALK
func (s *State) walkSFX(tile int) int {
	if m := s.Transport & 0xFE; m == mountedHorse {
		return u5data.SFXHorse
	}
	if TerrainCost(tile) > 0 {
		return u5data.SFXWalkSlow
	}
	return u5data.SFXWalk
}
