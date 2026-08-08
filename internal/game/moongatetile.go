package game

import "github.com/wicanr2/u5-cht/internal/u5data"

// 月門是「寫進地圖緩衝的一格」,不是狀態(原版 `sub_DEE4` / `sub_DE74`)
//
// 推導見 `docs/re/86`。原版每次重畫都跑一遍 `sub_DEE4`:夜裡把 tile 0xDC
// 寫進八顆月石埋藏的座標,白天把計數器降到 0 之後寫回草地。
//
//	esi = 0DCh
//	if (小時 >= 14h || 小時 < 5) sub_2BBB8(&byte_3E097, 1, 10h)   ; 夜裡:累加到 16
//	else { sub_2BBFC(&byte_3E097, 1); if (byte_3E097 == 0) esi = 5 }
//	for (i = 0; i < 8; i++)
//	    if (sub_DE74(i)) { changed = (*格 != esi); *格 = esi; if (changed) 重算光照 }
//
// ⚠⚠ **節奏差異(唯一沒辦法照抄的地方)**:原版的計數器是**每次重畫**升降一格,
// 而 `sub_29D64`(重畫)一個回合裡可能被叫好幾次。引擎沒有同節奏的重畫鉤子,
// 所以改成**每回合一次**(由 `tick()` 呼叫)⇒ 淡出從「不到一秒」變成
// 「約 16 回合」。**這是近似,不是原版**,已列進 A 階段對 DOSBox 的核對清單。

// MoongateFrameMax 是月門動畫格的上限(原版 `sub_2BBB8(…, 1, 10h)`)。
//
// ⚠ 那個計數器 `byte_3E097` **會進存檔**(`sub_27D24` / `sub_284CC`),
// 而且繪圖的 `sub_29BEC` 會讀它 —— 所以它不只是開關,是**動畫格編號**:
// 月門在黃昏長出來、天亮縮回去。⬜ 引擎的繪圖還沒用它(門只有一張圖),
// 存檔位移也還沒定位。
const MoongateFrameMax = 0x10

// RefreshMoongateTiles 把月門寫進地圖緩衝,或把它擦掉(原版 `sub_DEE4`)。
//
// ★★ 這一支是「月門存在」的**唯一來源** —— 踏上去會不會被傳送,由
// `EnterMoongateHere` 讀那一格的 tile 決定(原版 `sub_E084` 就是這樣)。
// 座標 + 時段的雙重判定已經拆掉:兩個真相來源會漂。
func (s *State) RefreshMoongateTiles() {
	if s.BaseSave == nil || s.InScene() || s.InDungeon() || s.InCombat() {
		return
	}
	tile := byte(u5data.MoongateOpenTile)
	if u5data.MoongateOpenAtHour(s.Clock.Hour) {
		// 夜裡:累加到上限,而且**無論計數器多少都寫月門**。
		if s.MoongateFrame < MoongateFrameMax {
			s.MoongateFrame++
		}
	} else {
		// 白天:遞減;★ **歸零才**寫回草地 —— 這就是日出後的殘留。
		if s.MoongateFrame > 0 {
			s.MoongateFrame--
		}
		if s.MoongateFrame == 0 {
			tile = u5data.MoongateClosedTile
		}
	}
	for i := range s.Moongates {
		if !s.moongateWritesHere(i) {
			continue
		}
		m := s.Moongates[i]
		if s.TileAt(m.X, m.Y) == tile {
			continue
		}
		s.SetTileAt(m.X, m.Y, tile)
		// 原版在這裡重算光照(`sub_2E21C`)—— 月門是發光地形(`docs/re/31` §6)。
		s.relightScene()
	}
}

// moongateWritesHere 是「第 i 顆月石此刻要不要寫」(原版 `sub_DE74`)。
//
//	埋在**當前的地點**、埋在**當前的樓層**,而且(在大地圖時)落在載入視窗內。
//
// ★ 它**完全不看月相** —— 月相只決定踏進去會去哪裡(`TravelByMoongate`)。
//
// ⚠ 視窗那一條用「隊伍為中心的 32×32」近似(引擎沒有 `byte_3E0AB`/`byte_3E0AC`,
// 同 `spawnWindowOrigin`)。這一條有可觀察的後果:**離開視窗時原版不會把
// 那一格寫回草地**,所以遠處的月門 tile 會留在地圖上直到玩家再走近。
// 照原樣做 —— 那不是 bug,而且順序保護了它(回到視窗內時先重寫才可能踏上)。
func (s *State) moongateWritesHere(i int) bool {
	if i < 0 || i >= len(s.Moongates) {
		return false
	}
	m := s.Moongates[i]
	if !m.Known() {
		return false
	}
	if m.Location != s.locationCode() {
		return false
	}
	if m.Floor != s.Floor {
		return false
	}
	if s.InScene() {
		return true // ★ 場景裡不查範圍(原版 `cmp dl, al; jnz → return 1`)
	}
	ox, oy := s.spawnWindowOrigin()
	dx := (m.X - ox) & 0xFF
	dy := (m.Y - oy) & 0xFF
	return dx < u5data.SpawnWindowSpan+1 && dy < u5data.SpawnWindowSpan+1
}

// relightScene 在月門開關時重算光照(原版 `sub_2E21C`)。
//
// 引擎的照明是每次算視線時重建的(`docs/re/31`),所以這裡只要讓下一次算
// 重新來過即可 —— 沒有需要主動失效的快取。留這一支是為了標出原版在
// 這個位置做了什麼,免得下一輪以為漏了。
func (s *State) relightScene() {}
