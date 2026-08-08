package game

import "github.com/wicanr2/u5-cht/internal/u5data"

// 環境音:靠得夠近的東西會自己出聲(原版 `sub_2BDE0`,由 `sub_29D64` 每次重繪呼叫)
//
// 推導見 `docs/re/92`。四類發聲物、只取最近的一個、走遠就靜下來。
//
// ⚠⚠ **這不是配樂**。第四類(樂器)會把配樂**停掉**換成蜂鳴器旋律,
// 走開之後再把原本那首接回去 —— 所以它同時是音效與配樂機制。

// AmbientSound 是這一幀該發的環境音。
//
// 由 `TickAmbient` 產生;`Kind == AmbientNone` 表示附近沒有發聲物。
type AmbientSound struct {
	Kind u5data.AmbientKind

	// SFX 是這一幀要放的音效索引(`u5data.SFXNone` = 不放)。
	//
	// 落地鐘只在第 0 / 4 相出聲,所以同一個 `AmbientClock` 大多數幀是 `SFXNone`。
	SFX int

	// ClockPitch 是滴答的音高參數(原版 `sub_2C4F4` 的第一參數)。
	//
	// ★ 滴(相位 0)是 0xBB8、答(相位 4)是 0x7D0 —— **兩者音高不同**。
	// 0 = 這一幀不是滴答。
	ClockPitch int

	// BeeperNote 是樂器旋律這一步的 MIDI 音符號(0 = 休止或不適用)。
	BeeperNote byte

	// MusicStopped 為真表示配樂正被樂器壓著。
	MusicStopped bool
}

// ambientCenter 回傳掃描中心(原版 `cmp byte_3E0A3, 80h`)。
//
// ★ 地點碼 ≥ 0x80(戰鬥與石室)時中心**固定在 (5,5)** ——
// 也就是視窗中央,而不是隊伍座標。戰場的座標系本來就只有 11×11。
func (s *State) ambientCenter() (int, int) {
	if s.locationCode() >= 0x80 {
		return u5data.AmbientScanRadius, u5data.AmbientScanRadius
	}
	return s.X, s.Y
}

// TickAmbient 掃一次周圍並推進計時器(原版 `sub_2BDE0` 整支)。
//
// 每次重繪地圖呼叫一次。它有副作用:推進落地鐘的八相計時器與旋律游標,
// 並在樂器進入 / 離開範圍時停掉 / 接回配樂。
func (s *State) TickAmbient() AmbientSound {
	kind := s.scanAmbient()
	out := AmbientSound{Kind: kind, SFX: u5data.SFXNone}

	// ★ 樂器進入範圍 → 記住現在那首、停掉配樂。
	if kind == u5data.AmbientMusic && !s.musicSuppressed {
		s.suppressedSong = s.CurrentSong()
		s.musicSuppressed = true
		s.playSong(SongNone)
	}
	// ★ 附近**完全沒有**發聲物 → 把配樂接回去。
	//
	// ⚠ 條件是 `kind == 0`,不是 `kind != 4` —— 從樂器走到瀑布旁邊時
	// 配樂**還是停著**的。原版兩處判斷都寫 `and ebx, ebx; jnz`。
	if kind == u5data.AmbientNone && s.musicSuppressed {
		s.musicSuppressed = false
		s.playSong(s.suppressedSong)
	}
	out.MusicStopped = s.musicSuppressed

	switch kind {
	case u5data.AmbientClock:
		out.SFX, out.ClockPitch = s.clockVoice()
	case u5data.AmbientWaterfall:
		// ⚠ 只在**第一次**進入範圍時放(原版 `byte_600A1`)——
		// 音檔本身 60,032 B 很長,靠它自己播完而不是每幀重觸發。
		if !s.waterfallPlaying {
			s.waterfallPlaying = true
			out.SFX = u5data.SFXWaterfall
		}
	case u5data.AmbientFountain:
		if !s.fountainPlaying {
			s.fountainPlaying = true
			out.SFX = u5data.SFXFountain
		}
	case u5data.AmbientMusic:
		out.BeeperNote = u5data.BeeperNote(u5data.BeeperMelody[s.beeperStep])
		s.beeperStep = (s.beeperStep + 1) % len(u5data.BeeperMelody)
	}

	// ★ 走遠了才把「已經在放」的旗標清掉(原版 `and ebx, ebx; jnz` 那一段:
	// 兩個旗標**同時**清,而且要其中至少一個是 1 才做)。
	if kind == u5data.AmbientNone && (s.waterfallPlaying || s.fountainPlaying) {
		s.waterfallPlaying, s.fountainPlaying = false, false
	}

	// 排進音效佇列 —— 播放由 `cmd/u5cht` 的 `TakeSFX` 那條路走,
	// 這裡不碰 ebiten(同 `song`:`internal/game` 只決定要放什麼)。
	if out.SFX != u5data.SFXNone {
		s.PlaySFX(out.SFX)
	}
	s.advanceClockPhase()
	return out
}

// scanAmbient 找 11×11 裡**最近**的發聲物(原版的雙層迴圈)。
//
// 距離用**平方**比,初值 0x33 同時是「還算在附近」的上限。
// ⚠ 條件是 `d² >= 目前最小 → 跳過`,所以**同距離時先掃到的贏**
// (x 由小到大、y 由小到大)。
func (s *State) scanAmbient() u5data.AmbientKind {
	cx, cy := s.ambientCenter()
	mask := s.SightMask()
	best := u5data.AmbientMaxDistSq
	found := u5data.AmbientNone
	r := u5data.AmbientScanRadius
	for dx := -r; dx <= r; dx++ {
		for dy := -r; dy <= r; dy++ {
			d2 := dx*dx + dy*dy
			if d2 >= best {
				continue
			}
			k := u5data.AmbientTileKind(int(s.TileAt(cx+dx, cy+dy)))
			if k == u5data.AmbientNone {
				// 第四類看疊圖層,而且要**那一格看得見**。
				if !SightVisible(mask, dx, dy) {
					continue
				}
				if !u5data.AmbientOverlayIsMusic(int(s.ambientOverlayAt(cx+dx, cy+dy))) {
					continue
				}
				k = u5data.AmbientMusic
			}
			best, found = d2, k
		}
	}
	return found
}

// ambientOverlayAt 回傳疊圖層那一格的 tile(原版 `byte_3F844[y*16 + x]`)。
//
// 大地圖與場景是由物件表寫進去的(原版 `sub_295AC`),所以這裡查物件表;
// 石室與戰場沿用 `absorbOverlayAt` 那一條路。
func (s *State) ambientOverlayAt(x, y int) byte {
	if s.Chamber != nil || s.Combat != nil {
		return s.absorbOverlayAt(x, y)
	}
	objs := s.CurrentObjects()
	if objs == nil {
		return u5data.OverlayEmpty
	}
	if o, ok := objs.At(x, y, s.Floor); ok {
		return o.Raw[u5data.ObjTile]
	}
	return u5data.OverlayEmpty
}

// clockVoice 決定落地鐘這一幀的聲音(原版 case 1)。
//
//	報時中(strikes > 0)且相位是 0 或 4 → 鐘響 ALARM3
//	否則 相位 0 → 滴(0xBB8)、相位 4 → 答(0x7D0)、其餘不出聲
//
// ★ **鐘響取代滴答**,不是疊在上面。
func (s *State) clockVoice() (int, int) {
	onBeat := s.clockPhase == u5data.ClockTickPhase || s.clockPhase == u5data.ClockTockPhase
	if s.clockStrikes > 0 && onBeat {
		return u5data.SFXAlarm, 0
	}
	switch s.clockPhase {
	case u5data.ClockTickPhase:
		return u5data.SFXClock, ClockTickPitch
	case u5data.ClockTockPhase:
		return u5data.SFXClock, ClockTockPitch
	}
	return u5data.SFXNone, 0
}

// 滴與答的音高參數(原版 `sub_2C4F4(0BB8h, 3)` / `sub_2C4F4(7D0h, 3)`)。
const (
	ClockTickPitch = 0xBB8
	ClockTockPitch = 0x7D0
)

// advanceClockPhase 推進八相計時器,並在報時的節拍上遞減剩餘敲擊數。
//
// 原版在 `def_2BFB1`(switch 的 default,也就是**所有分支的匯流處**):
//
//	if (剩餘敲擊 > 0 && (相位 == 0 || 相位 == 4)) 剩餘敲擊−−
//	相位++;  if (相位 > 7) 相位 = 0
//
// ⚠ 遞減在**遞增之前**,而且不管這一幀有沒有偵測到鐘 ——
// 走離落地鐘之後鐘聲照樣「敲完」(只是聽不到)。
func (s *State) advanceClockPhase() {
	if s.clockStrikes > 0 {
		if s.clockPhase == u5data.ClockTickPhase || s.clockPhase == u5data.ClockTockPhase {
			s.clockStrikes--
		}
	}
	s.clockPhase++
	if s.clockPhase >= u5data.ClockPhases {
		s.clockPhase = 0
	}
}

// StartClockChime 在整點時排定敲鐘次數(原版 `sub_29304` 的 `loc_2954D`)。
//
//	小時 == 0  → 敲 12 下
//	小時 > 12  → 敲(小時 − 12)下
//	其餘       → 敲 小時 下
//
// ★ 也就是**十二小時制報時**。由時鐘推進處呼叫(小時真的變了才呼叫)。
func (s *State) StartClockChime(hour int) {
	switch {
	case hour == 0:
		s.clockStrikes = ClockChimeMidnight
	case hour > ClockChimeMidnight:
		s.clockStrikes = hour - ClockChimeMidnight
	default:
		s.clockStrikes = hour
	}
}

// ClockChimeMidnight 是 0 點敲的下數,也是 12 小時制的除數(原版 `0Ch`)。
const ClockChimeMidnight = 12

// ClockStrikesLeft 回報還剩幾下沒敲(給測試與存檔用)。
func (s *State) ClockStrikesLeft() int { return s.clockStrikes }
