package u5data

import (
	"strings"
	"testing"
)

// TestBGMTableHasFifteenTracks —— ★ 表的列數與 `sub_3181C` 的上限互相佐證。
//
// `sub_3181C` 擋 `曲號 > 0x0E`,而表剛好 15 列 ⇒ **曲號就是列號**。
// 這兩個獨立來源一致,是「索引怎麼對」唯一的證據(`rulebook/62`)。
func TestBGMTableHasFifteenTracks(t *testing.T) {
	tracks, err := LoadBGMTable(fmTownsDir(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(tracks) != BGMSongCount {
		t.Fatalf("%d 首,預期 %d 首", len(tracks), BGMSongCount)
	}
	for i, tr := range tracks {
		if !strings.HasSuffix(tr.File, ".EUP") {
			t.Errorf("第 %d 首的檔名是 %q,預期 .EUP", i, tr.File)
		}
		for c, v := range tr.Volume {
			if v <= 0 || v > 127 {
				t.Errorf("第 %d 首聲道 %d 音量 %d 超出 0..127", i, c, v)
			}
		}
	}
}

// TestStartupSongFollowsTheJumpTable —— `sub_31DC0` 的四條路。
func TestStartupSongFollowsTheJumpTable(t *testing.T) {
	// (a) 存檔記著有效曲號 → 直接用,其餘判定全部跳過。
	//     連「在地牢裡」都蓋不掉它(原版 `cmp dl, 0Fh; ja` 在最前面)。
	if got := StartupSong(7, VehicleWalk, DungeonLocationBase, 0, 0); got != 7 {
		t.Errorf("存檔曲號 7 被蓋成 %d", got)
	}
	if got := StartupSong(0x0F, VehicleWalk, 0, 0, 0); got != 0x0F {
		t.Errorf("邊界 0x0F 該直接用,得到 %d", got)
	}

	const none = 0xFF // 沒有有效曲號
	// (b) 船 —— 三種載具碼都算。
	for _, v := range []int{VehicleSailing, VehicleSailing + 3, VehicleShip, VehicleSkiff + 2} {
		if got := StartupSong(none, v, 0, 0, 0); got != SongShip {
			t.Errorf("載具 0x%02X 該是船的曲子 %d,得到 %d", v, SongShip, got)
		}
	}
	// 馬與魔毯不是船。
	for _, v := range []int{VehicleWalk, VehicleCarpet, VehicleWalk + HorseToVehicle} {
		if got := StartupSong(none, v, 0, 0, 0); got == SongShip {
			t.Errorf("載具 0x%02X 被當成船了", v)
		}
	}

	// (c) 地牢。
	if got := StartupSong(none, VehicleWalk, DungeonLocationBase, 0, 0); got != SongDungeon {
		t.Errorf("地牢該是 %d,得到 %d", SongDungeon, got)
	}

	// (d) 地點跳表:下標 16..27 有專屬曲子,其餘落 default。
	for i, want := range songByLocation {
		l := Locations[i]
		if got := StartupSong(none, VehicleWalk, 0, l.X, l.Y); got != want {
			t.Errorf("地點 %d(%s)該是曲 %d,得到 %d", i, l.Name, want, got)
		}
	}
	// 八座大城與民居(下標 0..15)不在跳表裡 ⇒ 大地圖的曲子。
	for i := 0; i < 16; i++ {
		l := Locations[i]
		if got := StartupSong(none, VehicleWalk, 0, l.X, l.Y); got != SongOverworld {
			t.Errorf("地點 %d(%s)不在跳表裡,該落 default %d,得到 %d",
				i, l.Name, SongOverworld, got)
		}
	}
	// 荒郊野外。
	if got := StartupSong(none, VehicleWalk, 0, 1, 1); got != SongOverworld {
		t.Errorf("野外該是 %d,得到 %d", SongOverworld, got)
	}
}

// TestJumpTableCoversCastleAndKeepOnly —— ★ 跳表的涵蓋範圍不是隨便一段。
//
// 下標 16..23 是 `CASTLE.DAT` 的八個地點、24..27 是 `KEEP.DAT` 的前四個。
// 若哪天地點表的順序被改動,這條會先紅 —— 而不是等到「音樂配錯地方」。
func TestJumpTableCoversCastleAndKeepOnly(t *testing.T) {
	for i := range songByLocation {
		if i < 16 || i > 27 {
			t.Fatalf("跳表出現下標 %d —— 定義域只有 16..27", i)
		}
		want := 2 // CASTLE.DAT
		if i >= 24 {
			want = 3 // KEEP.DAT
		}
		if got := Locations[i].SceneFile; got != want {
			t.Errorf("地點 %d(%s)的場景檔是 %d,預期 %d", i, Locations[i].Name, got, want)
		}
	}
}
