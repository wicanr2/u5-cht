# 90 — CDDA 在哪些場景播,以及音效索引就是 `U5_SE.TBL` 的行號

| | |
|---|---|
| 輸入檔 | `re_work/fmtowns/WORRIORS.EXP`(+`.asm` / `.i64`)、`iso/U5_E/U5_SE.TBL` |
| 主要函式 | `sub_310E8`(★ 填 MSF 表)、`sub_30F8C`(★ 播一段 CD)、`sub_6730`(啟動 / 選單)、`sub_6128`(製作群字幕)、`sub_22D00`(石室)、`sub_2FE58`、`sub_2C46C`(★ 播音效) |
| 工具 | `tools/ida_cdda_hunt.idc` → `re_work/cdda_hunt.txt` |
| 起因 | 使用者指示:音效接進遊戲、用 IDA 查 CDDA 在哪個場景播 |

---

## 1. CD 音軌是用 **MSF 位址對**播的,不是「第幾軌」

`sub_30F8C(n)`:

```asm
sub_337DE(0, &err, &var_4, byte_60971, word_6096E)   ; 查碟(cdr_cdinfo 那一族)
if (err != 0) return                                  ; 讀不到就不播
sub_33575(0)
ebx = &byte_60948[n*3]                                ; ★ 起點 MSF(分/秒/格)
sub_33524(0, ebx)                                     ; AH=14h
sub_3370A(0, ebx, &byte_6095C[n*3])                   ; ★ 從起點播到終點
sub_3359A(0)
```

⇒ **`byte_60948` 是起點表、`byte_6095C` 是終點表,各 3 byte 一筆(MSF)。**

`sub_310E8` 在執行期填(靜態全 0,所以只讀檔案看不出來):

| 索引 | 起 | 迄 | 對應 |
|---|---|---|---|
| **0** | 6:00:00 | 12:08:74 | ★ **Track 3**(371 秒 = 6:11)|
| **1** | 3:00:00 | 6:00:00 | ★ **Track 2**(180 秒 = 3:00)|

對得上光碟結構:資料軌 13,350 sector ÷ 75 = 178 秒 = 2:58,加 2 秒間隙
⇒ Track 2 從 **3:00** 開始、長 3:00 ⇒ 迄 6:00 ✓;Track 3 從 6:00 開始 ✓。

## 2. ★★ 哪些場景播

`sub_6730`(啟動 / 選單迴圈)的順序,逐行讀出來:

```
載入 towns.fnt / u5.fnt / PROPORT.PCS / ibm.hcs / runes.ch
"Journey Onward"
BRITISH.PTH → a.gra → producti.gra          ; 製作商標誌
sub_310E8                                    ; ★ 讀碟、填 MSF 表
"Copyright 1988 Lord British"
sub_30F8C                                    ; ★ 播 —— 標題 / 版權畫面
sub_60BC                                     ; 主選單
"Select: "
sub_2FE58 → sub_30F8C                        ; ★ 播
"A:SAVED.GAM" / "No active game. Please create a character…"
UNDER.DAT / "initSoundEffect Error"
sub_6128                                     ; ★★ 製作群字幕(八段,每段播一次)
sub_22D00                                    ; ★★ MISCMAPS.DAT + TILES.16 → 石室
```

| 呼叫者 | 次數 | 場景 |
|---|---|---|
| `sub_6730` | 2 | **標題 / 版權畫面**(緊接 `"Copyright 1988 Lord British"`)與**選單之後** |
| `sub_6128` | 8 | **製作群字幕** —— 八段各一次(`"Produced and Designed by"`、`"FM-TOWNS version by"`、`"FM-TOWNS Art Work by"`、`"FM-TOWNS Sound by"`、`"Text transration by"`…)|
| `sub_22D00` | 1 | **石室畫面**(載 `MISCMAPS.DAT` + `TILES.16`;那份檔案的 11×11 石室有三種位移,見 `docs/re/27`/`28`)⬜ 哪一個未定 |
| `sub_2FE58` | 1 | ⬜ 未命名(反覆呼叫 `sub_2FD78`,無字串)|

### ⚠ 每個呼叫點傳的都是 0 或 3,不是 0 或 1

```asm
cmp  dword_60974, 0
jnz  short loc_61B2
push 0                  ; == 0 → 索引 0 = Track 3
jmp  short loc_61B4
loc_61B2:
push 3                  ; != 0 → 索引 3 = **全 0 的 MSF** = 空範圍
loc_61B4:
call sub_30F8C
```

⇒ `dword_60974` 是「能不能播 CD」的閘門;不能播時傳索引 **3**,而索引 3 的
MSF 從來沒被寫過(全 0)⇒ **等於什麼都不播**。

⇒ ★ **`WORRIORS.EXP` 只播 Track 3。索引 1(Track 2)被 `sub_310E8` 填好卻沒人用。**
⬜ 誰播 Track 2 未定 —— 最可能是 `U5OPEN/U5OPEN.EXP`(獨立的開場程式,
同樣連著 `cdr_*` 那組 API)。

⚠ 我第一次抓參數時只取「最近的一個 `push`」,結果十二個呼叫點全報「3」——
**漏了分支**。`push 0` 與 `push 3` 是二選一,取最近的那個必然只看到後者。
⇒ 抓參數要看**分支結構**,不是往前找第一個 push。

## 3. `cdr_*` 的名字怎麼對上函式

`cdr_status` `cdr_mstop` `cdr_mtplay` `cdr_rmtplay` `cdr_mphase` `cdr_cdinfo`
這些字串**在程式碼區裡**(檔案 0x33xxx),是 Fujitsu 函式庫把常式名內嵌在
程式旁邊的慣例 —— 不是 CodeView 符號。**線性位址 = 檔案位移 − 0x200**(已驗證)。

`tools/ida_cdda_hunt.idc` 拿每個 banner 位址往前找函式,再列它的呼叫者。
⚠ 這一版 IDA 的 IDC **沒有 `next_func`**,用了就整個腳本靜默中止
(只寫出第一筆就停)—— 同 `CLAUDE.md §4.5` 記的「IDC 崩掉不會說話」。

## 4. ★★ 音效索引 = `U5_SE.TBL` 的行號(日文檔名交叉確認)

`sub_2C46C(索引, 音高)` 有 **17 個呼叫點**。把索引對到表的行號之後:

| 呼叫者 | 索引 | 表上的檔名 | 交叉確認 |
|---|---|---|---|
| `sub_10A1C`(墜落動畫)| 0x14 = 20 | `T_OCHI1.SND` | ★★「落ち」= 掉落 |
| `sub_135FC`(月門過場)| 0x0A = 10 | `MOON2.SND` | ★★ 用了**三次** |
| `sub_10C34` | 9 | `MAHOU1.SND` | ★「魔法」 |
| `sub_2AC08` ×2 | 7 | `DAME1.SND` | ★「ダメージ」= 傷害 |
| `sub_2C598` | 0x11 = 17 | `FUNSUI2.SND` | ★「噴水」= 噴泉 |
| `sub_2C4F4` | 0x0F = 15 | `CLOCK2.SND` | 時鐘 |
| `sub_2C188` | 0x0E = 14 | `Fire.SND` | 火 |
| `sub_CAC` | 0x10 = 16 | `MIRROR2.SND` | 鏡 |
| `sub_35EC` | 0x12 = 18 | `SUITEKI3.SND` | 「水滴」 |
| `sub_2BDE0` | 0x16 = 22 | `ALARM3.SND` | 警報 |

**六個以上的呼叫點,索引與日文檔名逐一相符** ⇒ 索引就是行號,不是別的編號。
這比「表的順序看起來像」強得多 —— 它是**語意上的**交叉確認。

### ⚠ 觸發點的證據強度分兩級,不要混

- **A 級(有呼叫點)**:上表那十個。
- **B 級(只有檔名)**:`WALK` `WALKSLOW` `HORSE` `BLOCK` `ATTACK1/2` `NIGE`
  `DAME2` `YOGAN` `WHAT` `DOKU` `TAKI2` `DEATH1`。
  檔名一望即知(`WALK` = 腳步),但**沒有追到呼叫點** —— 原版在哪裡放腳步聲、
  粗糙地形是否真的換 `WALKSLOW`,都還沒查。

引擎目前接了:走路 / 馬蹄 / 慢速腳步(B)、去路受阻(B)、月門(A)、墜落(A)、
受傷(A)、施法(A)。**每一處都在程式碼裡標了級別。**
⬜ 剩下的觸發點要逐一追 `sub_2C46C` 那 17 個呼叫點的上游。

### ⚠ 引擎與原版的兩個已知差異

1. **通道管理沒做。** 原版有 8 個 PCM 通道、`通道 = 索引 & 7`(`docs/re/63`)
   ⇒ **同索引的音效會互相打斷**、不同索引可以疊。引擎讓 ebiten 自己混音
   ⇒ 連放同一個音效會疊起來而不是截斷。
2. **取樣率是推測的**(`SndAssumedRate` = 7,926,檔頭 +0x08 的常數)。
   原版**不傳取樣率**給驅動 —— 傳的是**音高**,而音高 → 播放率的換算在
   `TBIOS.BIN` 裡(`docs/re/89`)。⇒ 音效可能整體偏高或偏低。

## 5. 引擎對應

| 原版 | 引擎 | 狀態 |
|---|---|---|
| MSF 表 + `sub_30F8C` | — | ⬜ 引擎沒有標題 / 製作群 / 石室動畫這些場景 |
| Track 3 用在標題 / 字幕 / 石室 | `assets/audio/CDDA3.ogg` 已轉出 | ⬜ 未接播放點 |
| Track 2 | `assets/audio/CDDA2.ogg` 已轉出 | ⬜ 連原版都沒播(§2)|
| 音效索引 = 表行號 | `u5data.SFX*` 常數 | ✅ |
| `sub_2C46C` 播放 | `audio.SFXPlayer.Play` | ✅(⚠ 無通道管理)|
| 事件 → 索引 | `game.PlaySFX` + `walkSFX` | 🔶 六個事件已接,其餘 ⬜ |
