package u5data

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// FM Towns 版的配樂表 `U5_BGM.TBL` 與「開局播哪一首」
//
// 推導見 `docs/re/87`。音效那一半(`U5_SE.TBL` + 25 個 `.SND`)早就做完了,
// 在 `snd.go`(`LoadSoundTable` / `LoadSoundSet`)—— 這一輪差點又寫一份,
// 是先 grep 自己的程式碼才發現的。
//
// ⚠⚠ **這張表不是「場景 → 曲目」的對應表**。`CLAUDE.md §2.4` 此前寫
// 「場景配樂對應**免逆向,直接讀表**」是錯的:
//
//	U5_BGM.TBL: M1.EUP 102 87 87 87 87 87    ← 檔名 + **六個 FM 聲道的起始音量**
//
// 「什麼時候播第幾首」寫在**程式碼裡**(32 個 `sub_3181C` 呼叫點 +
// `sub_31DC0` 的地點跳表),要逆向才拿得到。錯因是只看了表的形狀就下結論,
// 沒有追「誰讀這張表、讀去做什麼」(`rulebook/62`)。

// BGMSongCount 是原版的曲目數(`sub_3181C` 的 `cmp eax, 0Eh; jle`,即 0..14)。
//
// ★ 這個上限與 `U5_BGM.TBL` 的 15 行**互相佐證**:曲號就是表的列號。
const BGMSongCount = 15

// BGMChannels 是每首曲子在表裡帶的音量欄位數。
//
// 六個 —— 對得上 FM Towns 的六個 FM 聲道,也對得上 `sub_3181C` 淡出時
// `rep movsd ecx=6` 複製六個 dword、再逐一 `sub_34685(聲道, 音量)` 遞減。
const BGMChannels = 6

// BGMTrack 是 `U5_BGM.TBL` 的一列。
type BGMTrack struct {
	File string
	// Volume 是六個 FM 聲道的**起始音量**。原版換曲時把它們一起遞減到 0
	// 當作淡出(`sub_3181C` 的 `loc_318BF` 迴圈,`esi` 從 0x7F 倒數)。
	Volume [BGMChannels]int
}

// ParseBGMTable 解 `U5_BGM.TBL`(純文字,每行「檔名 + 六個音量」)。
func ParseBGMTable(raw []byte) ([]BGMTrack, error) {
	// ⚠ 檔尾有一個 0x1A(DOS 的 Ctrl-Z EOF),同 `ParseSoundTable`。
	// 不切掉的話最後會多出一列一欄的垃圾 —— 這一條是抄過來的,不是猜的。
	text := string(raw)
	if i := strings.IndexByte(text, 0x1A); i >= 0 {
		text = text[:i]
	}
	var out []BGMTrack
	for i, line := range strings.Split(text, "\n") {
		f := strings.Fields(strings.TrimRight(line, "\r"))
		if len(f) == 0 {
			continue
		}
		if len(f) != 1+BGMChannels {
			return nil, fmt.Errorf("第 %d 行有 %d 欄,預期 %d 欄:%q", i+1, len(f), 1+BGMChannels, line)
		}
		var tr BGMTrack
		tr.File = f[0]
		for c := 0; c < BGMChannels; c++ {
			v, err := strconv.Atoi(f[1+c])
			if err != nil {
				return nil, fmt.Errorf("第 %d 行第 %d 個音量不是數字:%q", i+1, c+1, f[1+c])
			}
			tr.Volume[c] = v
		}
		out = append(out, tr)
	}
	if len(out) != BGMSongCount {
		return nil, fmt.Errorf("有 %d 首,預期 %d 首(`sub_3181C` 擋 > 0x0E)", len(out), BGMSongCount)
	}
	return out, nil
}

// LoadBGMTable 從 FM Towns 的 U5_E 目錄讀 `U5_BGM.TBL`。
func LoadBGMTable(dir string) ([]BGMTrack, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "U5_BGM.TBL"))
	if err != nil {
		return nil, err
	}
	return ParseBGMTable(raw)
}

// 開局播哪一首 —— `sub_31DC0` 在讀檔完成後跑一次
//
// 完整推導見 `docs/re/87` §3。⚠ 它**只在遊戲啟動時跑一次**(唯一的呼叫者是
// `sub_6730` 的選單迴圈,位置就在載入 `UNDER.DAT` 與 `initSoundEffect` 之間),
// 不是每次換場景。換場景的音樂由別的呼叫點負責(`sub_2D72C` 進場景、
// `sub_2D564` 洞穴 / 礦坑 / 地牢…)。

// SongOverworld 是 `sub_31DC0` 的 default(跳表的 default case 與定義域之外)。
const SongOverworld = 1

// SongShip 是「開局就在船上」的曲號(`byte_3E08C & 0xFC` 落在 0x20 / 0x24 / 0x28)。
const SongShip = 2

// SongDungeon 是「開局就在地牢」的曲號(`byte_3E0A3 >= 0x21`)。
const SongDungeon = 9

// songByLocation 是 `sub_31DC0` 的 `jpt_31E71`,key 是**地點索引**(`Locations` 的下標)。
//
// ⚠ 跳表的定義域是 `地點索引 − 0x10`,而地點表只有 32 筆 ⇒ **只有下標 16..27
// 用得到**;跳表裡 case 57(曲 0x0B)與 case 62(曲 7)是編譯器補的,
// **走不到**。所以 `M12.EUP`(曲 0x0B)在這條路上是死的。
//
// ★ 涵蓋範圍剛好是 `CASTLE.DAT` 的八個地點(16..23)+ `KEEP.DAT` 的前四個
// (24..27)。八座大城(`TOWNE.DAT`,下標 0..7)與民居(`DWELLING.DAT`,8..15)
// 落到 default ⇒ **開局在城裡會先響大地圖的曲子**。
//
// ⚠ 這聽起來像 bug,但另一條路解釋得通:存檔裡的 `byte_3DDB0` 若是有效曲號
// (≤ 0x0F)就**直接播它**,跳表只是「沒有有效曲號時」的推導。
// 兩者哪一個實際生效,要對 DOSBox 實測 —— 已列進 A 階段的核對清單。
var songByLocation = map[int]int{
	16: 8,    // CASTLE.DAT 1 —— 不列顛王的城堡(樓層 −1..3)
	17: 4,    // CASTLE.DAT 6 —— 黑刺的城堡(樓層 −1..3)
	18: 8,    // WEST BRITANNY
	19: 8,    // NORTH BRITANNY
	20: 5,    // EAST BRITANNY
	21: 6,    // PAWS
	22: 9,    // COVE
	23: 9,    // BUCCANEER'S DEN
	24: 9,    // ARARAT
	25: 4,    // BORDERMARCH
	26: 5,    // FARTHING
	27: 0x0C, // WINDEMERE
}

// NoStartupSong 代表存檔沒有記著有效曲號(原版拿 `byte_3DDB0` 與 0x0F 比)。
const NoStartupSong = 0xFF

// StartupSong 回報開局該播第幾首(原版 `sub_31DC0`)。
//
//	saved 是存檔記的曲號:**0..0x0F 就直接用**(原版 `cmp dl, 0Fh; ja`)
//	否則:在船上 → 2;地牢(地點碼 ≥ 0x21)→ 9
//	      站在地點表第 16..27 筆的座標上 → 查表
//	      其餘 → 1
//
// ⚠ 判定順序照原版:載具**先於**地點碼。在地牢裡不可能坐船所以看不出差別,
// 但順序照抄比較安全 —— 哪天發現船能進地牢,行為才會跟原版一致。
func StartupSong(saved, transport, locationCode, x, y int) int {
	if saved >= 0 && saved <= 0x0F {
		return saved
	}
	switch VehicleKind(byte(transport)) {
	case VehicleSailing, VehicleShip, VehicleSkiff:
		return SongShip
	}
	if locationCode >= DungeonLocationBase {
		return SongDungeon
	}
	for i := range Locations {
		if Locations[i].X == x && Locations[i].Y == y {
			if song, ok := songByLocation[i]; ok {
				return song
			}
			break // ★ 原版找到第一筆就停,沒對到跳表就落 default
		}
	}
	return SongOverworld
}
