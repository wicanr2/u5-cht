# 92 — 白噪轉接層,與「靠近才聽得到」的環境音

| | |
|---|---|
| 輸入檔 | `re_work/fmtowns/WORRIORS.EXP`(一手位元組)、`WORRIORS.EXP.asm` / `.i64` |
| SHA-256 | 見 `re_work/fmtowns/SHA256SUMS.txt` |
| 主要函式 | ★★ `sub_2C598`(白噪 → PCM 轉接)、★★ `sub_2BDE0`(環境音掃描)、`sub_2C46C`(放 PCM)、`sub_2C4F4`(滴答)、`sub_33D78`(蜂鳴器單音)、`sub_29304`(報時排定) |
| 主要資料 | `byte_60038`(八相)、`byte_3E094`(剩餘鐘響)、`byte_5FFFC`(旋律游標)、**0x60060 音高表**、**0x6006C 53 步旋律** |
| 起因 | 逆 `sub_2AB38`(寶箱陷阱)時要查那個「超出範圍的音效索引」 |
| 狀態 | ✅ 兩支全解;四筆 B 級證據升成 A 級;環境音與報時已落地 |

---

## 0. ★★★ `sub_2C598` 是轉接層,不是發聲器

它的除錯分支自己說了參數是什麼:

```asm
cmp     dword_5FFF4, 1                 ; debug 旗標
jnz     short loc_2C5DA
push    ebx ; push esi ; push edi
push    offset aWhiteNoiseRate         ; "white_noise\nRate %d\nDura %d\nLimit %d"
```

⇒ `sub_2C598(Rate, Dura, Limit)` —— **DOS 版產生白噪用的三個參數**。
函式本體除了這個 debug 印字之外**沒有任何發聲呼叫**,只有八組
「三個參數都相等就換成某個 PCM 音效」的比較:

| Rate | Dura | Limit | → | 音效 |
|---|---|---|---|---|
| 1 | 0x19 | 0x3E8 | `dword_6015C = 1` | (不出聲,只武裝旗標)|
| 1 | 0x19 | 0x5DC | `sub_2C46C(0, 0x3C)` + 清旗標 | **`WALK.SND`** 腳步 |
| 0x0A | 0x640 | 0x7D0 | `sub_2C46C(0x0D, 0x3C)` | `DOKU.SND`(毒)|
| 0x14 | 0x3C | 0x2710 | `sub_2C46C(0x13, 0x3C)` | `TAKI2.SND`(瀑布)|
| 0x28 | 0xBB8 | 0x1F4 | `sub_2C46C(7, 0x3C)` | `DAME1.SND`(傷害)|
| 0x0A | 0xBB8 | 0x7D0 | `sub_2C46C(8, 0x3C)` | `DAME2.SND` |
| 0x320 | 0x1F40..0x5140 | 0x2BC | `sub_2C46C(9, 0x3C − (Dura−0x1F40)/0x640)` | `MAHOU1.SND`,**音高隨 Dura 變** |
| 0x0A | 0x1E | 0x61A8 | `sub_2C46C(0x11, 0x3C)` | `FUNSUI2.SND`(噴泉)|

★★ 這是 FM Towns 移植的**改裝痕跡**:原作的 DOS 版用 PC 喇叭產生白噪(掃頻),
移植時作者沒有改動所有呼叫點,而是在最底層攔下特定的參數組合換成取樣音效。
所以每一個「白噪音效」在 FM Towns 上都有一個對應的 `.SND`,
**而查不到的組合在 FM Towns 上是完全無聲的**。

### 26 個呼叫點:19 個有替代,7 個無聲

| 呼叫者 | (Rate, Dura, Limit) | 結果 |
|---|---|---|
| `sub_1318`(每回合地形效果)| 0x28, 0xBB8, 0x1F4 | DAME1 |
| `sub_A360`("ARGH!")| 同上 | DAME1 |
| `sub_13DD8`("Plague!")| 同上 | DAME1 |
| `sub_2AB38`(★ 寶箱陷阱)| 同上 | DAME1 |
| `sub_2B21C` | 同上 / 0x0A, 0xBB8, 0x7D0 | DAME1 + DAME2 |
| `sub_CD28`(許願井 "Poof!")| 0x0A, 0xBB8, 0x7D0 | DAME2 |
| `sub_1A5E8` | 同上 | DAME2 |
| `sub_2B1C8` | 同上 | DAME2 |
| `sub_2A464`(★ 扣一人的血)| 0x0A, 0x640, 0x7D0 | **DOKU** |
| `sub_1AEB4` / `sub_1C8E8` / `sub_1FA6C` | 0x320, 變數, 0x2BC | **MAHOU1(音高隨參數)** |
| `sub_22B50` / `sub_2BDE0` ×2 | 0x14, 0x3C, 0x2710 | **TAKI2** |
| `sub_2BDE0` | 0x0A, 0x1E, 0x61A8 | FUNSUI2 |
| `sub_2C118`(★ 腳步)| 1, 0x19, 0x3E8 → 1, 0x19, 0x5DC | **WALK** |
| `sub_30E8` / `sub_4834` | 1, ?, 0x4E20 | 無聲 |
| `sub_4DC8` / `sub_4E58` | 1, 0x32, 0xDAC | 無聲 |
| `sub_BCC4`("regurgitated!")| 1, 0x1B58, 0x258 | 無聲 |
| `sub_22D00` | 1, 0x4B0, 0xFA0 | 無聲 |
| `sub_2CE70`(撞擊 / "Docked!")| 0x64, 0x7D0, 0x12C | 無聲 |

**7 個無聲的呼叫點就是原版的行為** —— 引擎照樣不出聲(`CLAUDE.md §3.0`)。
⚠ 如果哪天覺得「撞船應該有聲音」而自己補一個,那是自創。

### ★ 腳步聲是**兩次**呼叫夾出來的

```
sub_2C118()                     ; ← sub_86C:loc_88B(移動)呼叫
  sub_2C598(1, 0x19, 0x3E8)     ; 武裝:dword_6015C = 1
  nullsub_19(1, 0x14)           ; 空的(原本是延遲)
  sub_2C598(1, 0x19, 0x5DC)     ; 旗標還在 → 放 WALK.SND,清旗標
```

DOS 版的一個腳步是**兩段不同 Limit 的白噪**;FM Towns 用一個旗標把這一對
收斂成一次取樣播放。⇒ 這就是把 `WALK` 從 B 級(只有檔名)升成 A 級的證據。

## 1. ★★★ `sub_2BDE0` —— 環境音掃描

由 `sub_29D64`(重繪地圖)每次呼叫。掃隊伍周圍 **11×11**,找**距離平方最小**
的那一個發聲物,只讓它出聲:

```
最近距離² = 0x33                                  ; 51 = 5²+5²+1,也是上限
if (地點碼 < 0x80) { 中心 = 隊伍座標 }
else               { 中心 = (5, 5) }               ; ★ 戰鬥與石室固定中心
for (x = 中心X−5 .. 中心X+5)
  for (y = 中心Y−5 .. 中心Y+5)
    d² = (x−中心X)² + (y−中心Y)²
    if (d² >= 最近距離²) continue                  ; ★ 同距離時先掃到的贏
    tile = *sub_DB10(y, x)
    if      ((tile & 0xFE) == 0xFA) 種類 = 1       ; 落地鐘
    else if ((tile & 0xFC) == 0xD4) 種類 = 2       ; 瀑布
    else if ((tile & 0xFC) == 0xD8) 種類 = 3       ; 噴泉
    else if (那格可見 && (疊圖 & 0xFC) == 0x5C) 種類 = 4   ; 樂器
    else continue
    最近距離² = d²
```

四組 tile 都用 sprite 交叉確認過(`u5dump tiles-fmtowns` 切出來看):

| 遮罩 | tile | sprite | LOOK 表 |
|---|---|---|---|
| `& 0xFE == 0xFA` | 0xFA, 0xFB | 高身黃色鐘櫃 + 紅色鐘擺 | "a grandfather clock, showing: " ✓ |
| `& 0xFC == 0xD4` | 0xD4..0xD7 | 藍色瀑布,四格動畫 | "a waterfall" ✓ |
| `& 0xFC == 0xD8` | 0xD8..0xDB | 白色台座 + 藍色水花,四格動畫 | `*`(被程式特判 → 噴泉)|
| 疊圖 `& 0xFC == 0x5C` | 0x5C/0x5D 書架、0x5E/0x5F 兩格寬的樂器形物件 | | "a crowded bookshelf" / "a Guardian!" |

⚠ 第四組的遮罩**同時涵蓋書架**。原版就是這樣遮的,引擎照抄 —— 縮小範圍就是自創。

### 落地鐘:滴答 + 報時

```
byte_60038 = 八相計時器,每次掃描 +1,> 7 歸零
if (剩餘鐘響 > 0 && (相位 == 0 || 相位 == 4))  → sub_2C46C(0x16, 0x3C)   ; ALARM3
else if (相位 == 0)                            → sub_2C4F4(0xBB8, 3)     ; 滴
else if (相位 == 4)                            → sub_2C4F4(0x7D0, 3)     ; 答
匯流處:if (剩餘鐘響 > 0 && (相位 == 0 || 4)) 剩餘鐘響−−
```

★ 三件容易漏的:

1. **滴與答音高不同**(0xBB8 vs 0x7D0),不是同一個聲音重複。
2. **鐘響取代滴答**,不是疊上去。
3. 遞減在**匯流處**(switch 的 default),所以走離鐘之後鐘照樣「敲完」,只是聽不到。

剩餘鐘響 `byte_3E094` 在 `sub_29304`(時間推進)排定:

```
if (byte_3E08F != byte_3E090) {          ; ★ 小時**欄位**變了(不是「跨過幾小時」)
   if (byte_3E08F == 0)  byte_3E094 = 0x0C          ; 0 點 → 12 下
   else                  byte_3E094 = 小時 > 12 ? 小時 − 12 : 小時
}
```

⇒ **十二小時制報時**。而 `byte_3E090` 的快照在函式**最上方**、`inc byte_3E08F`
在 `byte_3E08A == 'T'`(An Tym)的跳躍**之後**
⇒ **時間停止期間時鐘不報時**。

### 瀑布與噴泉:只在進入範圍時觸發一次

`byte_600A1` / `byte_600A2` 兩個旗標,而清除的條件是
「這一幀**完全沒有**發聲物」——**不是**「換成別的發聲物」。
所以從瀑布走到噴泉旁邊時噴泉不會響,得先走到兩者都聽不到的地方。

### 種類 4:把配樂停掉,改放蜂鳴器旋律

```
if (種類 == 4 && !壓制中) { 記住當前曲號; 停配樂; 壓制中 = 1 }
if (種類 == 0 && 壓制中)  { 壓制中 = 0; 把記住的那首接回去 }
```

⚠ 接回去的條件是 `種類 == 0`,**不是** `種類 != 4` ——
從樂器走到瀑布旁邊,配樂**還是停著**。

## 2. ★★★ 那 53 步旋律

```asm
eax = offset dword_5FFF4          ; IDA 拿 debug 旗標當基底,實際上是四個獨立全域
edx = [eax+8]                     ; 0x5FFFC 游標
dl  = [eax+edx+0x78]              ; 0x6006C 序列[游標]
if (dl != 0) sub_33D78([eax+dl+0x6B], 0x78, 0x9C40)   ; 0x6005F 音高表[序列值]
[eax+8]++;  if ([eax+8] >= 0x35) [eax+8] = 0          ; ★ 週期 53
```

一手位元組(`WORRIORS.EXP`,線性位址 = 檔案位移 − 0x200):

```
0x60060  3E 40 42 43 45 47 48 49 4A      ; 音高表(九筆)
0x6006C  01 04 04 00 00 01 05 05 00 00 04 09 06 09 07 04 06 05 01 04
         00 00 00 01 05 00 00 00 04 09 06 05 06 04 00 00 00 05 08 08
         09 05 06 08 09 05 04 06 05 04 03 02 01                       ; 53 筆
```

音高表 = **62 64 66 67 69 71 72 73 74** = MIDI 音符號
= `D4 E4 F#4 G4 A4 B4 C5 C#5 D5`(D 大調音階,外加 C 本位)。序列值是 1-based,0 = 休止。

旋律:

```
D4 G4 G4 _  _  D4 A4 A4 _  _  G4 D5 B4 D5 C5 G4 B4 A4 D4 G4
_  _  _  D4 A4 _  _  _  G4 D5 B4 A4 B4 G4 _  _  _  A4 C#5 C#5
D5 A4 B4 C#5 D5 A4 G4 B4 A4 G4 F#4 E4 D4
```

`sub_33D78(音高, 0x78, 時長)` 是單音蜂鳴器:時長為 0 就只做「停」
(所以離開時的 `sub_33D78(0x3D, 0x78, 0)` 是靜音而不是放音)。

### ✅ 這段旋律**不是** 15 首 `.EUP` 任何一首(含正對照)

拿前 40 音的音高序列去比對全部 15 首 `M*.EUP`:

```
M1.EUP   882 音 6 聲道 → 最長相符 2 音
M12.EUP  717 音 6 聲道 → 最長相符 3 音
…
★ 正對照:共解析出 7,633 個音符,最長相符 3 音
```

**正對照是必要的** —— 如果 `ParseEUP` 壞掉、或路徑打錯,「找不到」看起來一模一樣
(`~/diagnosis-notes/docs/02-query-returned-empty/`)。7,633 這個數字才是
「真的比對過」的證據。

## 3. 引擎落地

| 原版 | 引擎 |
|---|---|
| `sub_2C598` 分派表 | 直接接目標音效(不重造白噪層)|
| `sub_2BDE0` 掃描 | `internal/game/ambient.go` 的 `TickAmbient` / `scanAmbient` |
| 四組 tile 遮罩 | `u5data.AmbientTileKind` / `AmbientOverlayIsMusic` |
| `byte_60038` 八相 | `State.clockPhase` + `advanceClockPhase` |
| `byte_3E094` 鐘響 | `State.clockStrikes` + `StartClockChime`(掛在 `AdvanceTime`)|
| `byte_5FFFC` 游標 | `State.beeperStep` |
| `byte_5FFFD` 壓制 | `State.musicSuppressed` / `suppressedSong` |
| `byte_600A1/2` | `State.waterfallPlaying` / `fountainPlaying` |
| 0x60060 / 0x6006C | `u5data.BeeperScale` / `BeeperMelody` |

### ⬜ 還沒做的

- **蜂鳴器單音沒有實際發聲**。`internal/audio` 目前只播 ogg 與 `.SND` 取樣,
  要放這段旋律得加一個方波產生器。`TickAmbient` 已經把 MIDI 音符號算出來了,
  接上去只差一個振盪器 —— 但**配樂停下來這件事已經做了**(玩家聽得到差異)。
- `sub_2C598` 的 `MAHOU1` 那一條**音高隨 Dura 變**;引擎的 `PlaySFX` 只吃索引,
  音高還沒接(同 `docs/re/90` 記的「原版傳音高不傳取樣率」)。
- `sub_E2A4`(整點時在大地圖 / 場景才呼叫的那一支)還沒逆。

## 4. 這次踩到的形狀

**「同一個位置在不同函式裡不是同一個意思」。** `docs/re/91` 第一版把
`sub_2C598(0x28, …)` 的 `0x28` 當音效索引 —— 理由是「別的函式第一參數是索引」——
然後因為 40 > 23 而標了一條「未接」的 ⬜。真值就寫在函式裡的格式字串上。

這與 `CLAUDE.md §4.5` 的「IDA 自動命名不是資料」是同一族:
**用位置、名字、形狀推語意,而沒有回去讀那支函式本體。**
代價是一條假的 ⬜ 卡在筆記裡,以及差點漏掉整個環境音系統。

⇒ 判斷法:要對某個參數下語意結論之前,**先看被呼叫的那支函式怎麼用它**。
成本是讀四十行組語,而它這次直接換來 19 個音效觸發點 + 一整套環境音機制。
