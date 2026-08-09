# 13 — XMI(Miles / AIL)→ 標準 MIDI

| | |
|---|---|
| 輸入檔 | `gamedata/upgrade/` 的 **17 個 XMI**(16 首音樂 + `setm` 測試檔;Exodus Project MIDI Upgrade 1.0,2001)|
| 工具 | `tools/xmi2mid.py`(純 Python,無外部依賴)、`tools/mt32_render.sh`(munt)、`tools/match_songs.py` |
| 驗收 | 16/16 解析並渲染成 ogg,長度與 tempo 積分相符;**含反對照**(餵非 XMI、多首容器都會報錯)|
| 狀態 | ✅ 完成:轉檔 → MT-32 渲染 → 對上曲號 → 接進引擎(F5 切換)|

---

## 0. 首數:16 首音樂 + 1 個測試檔

`gamedata/upgrade/` 下有 **17 個 XMI**。16 首是音樂,每個檔恰好一個 `EVNT` chunk
(XMI 是容器,一個檔可以裝多首,所以要數 chunk 不是數檔案);第 17 個 `setm.xmi`
裝了**五首**,是驅動的測試檔,不是遊戲音樂 —— 轉檔工具的「多首就報錯」那道檢查
把它擋下來,這一擋是對的。

```
AMIGA BLCKTHRN BRITLAND ENGGMNT FANFARE GREYSON HALLS HORNPIPE
LADYNAN MONARCH REUNION RULEBRIT STONES U5THEME WRLDBLW trntlla
```

⚠⚠ **`trntlla.xmi` 的檔名是小寫**,其餘 15 個是大寫。
數檔案時用 `*.XMI` 這種**大小寫敏感**的 glob 會漏掉它 —— 而漏掉之後
「15 首」看起來完全自洽(剛好等於 FM Towns 的 15 首 `.EUP`),
所以錯了也不會有任何跡象。交叉核對 `upgrade/Files.txt` 的曲名清單才發現。
⇒ **檔名清單一律用大小寫無關的方式取**(`nocaseglob` / `[xX][mM][iI]`),
引擎查檔也一樣(`internal/audio` 的 `mt32Set` 建了小寫索引)。

16 首之中 **15 首與 FM Towns 的 15 首 `.EUP` 是同一批曲子的另一個編曲**
(旋律配對的證據見 `docs/re/98` §3);多出來的 `AMIGA`(Amiga Theme,383 秒)
對任何一首都不像 ⇒ **它沒有曲號**,遊戲流程不會播它。

## 1. 容器結構

```
FORM XDIR { INFO (u16 首數) }
CAT_ XMID { FORM XMID { TIMB, [RBRN], EVNT } … }
```

IFF 慣例:chunk 長度是大端 u32,而且**補齊到偶數**(奇數長度後面多一個填充位元組)。
`TIMB` 是這首要用到的 (音色, bank) 對,6..16 B。

## 2. ★★ 與標準 MIDI 的三處差異

少處理任何一處都會得到「音全部黏在一起」或「速度快十倍」。

### ① delta 時間是**累加的裸位元組**,不是 VLQ

```
連續的 0x00..0x7F 各自貢獻自己的值,加起來才是這一步的等待;
碰到 >= 0x80 就是狀態位元組,delta 結束。
```

⚠ 這**不是** VLQ。VLQ 用最高位元當「還有下一個位元組」的旗標;
XMI 用「還是不是 < 0x80」。兩者在小數值時看起來一樣,
所以拿 VLQ 讀短曲子會「大致對」而長曲子整個垮掉 ——
典型的「同一個觀察支持兩個模型」。

### ② 沒有 note-off,note-on 自帶時長

```
9n <note> <velocity> <VLQ 時長>
```

★ 時長這一個**是** SMF 風格的 VLQ(與 ① 不同種)。同一個檔案裡兩種變長編碼並存。
轉檔時要自己排一個 note-off 到 `現在 + 時長`,並在輸出前依 tick 重新排序。

### ③ ppqn 是 60,而速度由曲子自帶的 `FF 51` 決定

**ppqn = 60**(AIL 的慣例),而 15 首**每一首**在 tick 0 都有一個 `FF 51`。
`U5THEME` 那顆是 **600,000 µs/四分音符** ⇒ 一個 tick = **10 ms**。

⚠⚠ 這一節曾寫「時基固定 1/120 秒(= 8.33 ms)」—— **錯的**,而且錯得看不出來:
轉出來的 `.mid` 一直是對的(檔案自帶的 tempo 蓋掉了預設),
只有**我自己的長度估算**用了 `tick ÷ 120`。抓到它靠的是 MT-32 渲染:

| | U5THEME |
|---|---|
| 「固定 1/120」算的 | 123.5 秒 |
| 「ppqn=60 + 檔案 tempo」算的 | **148.2 秒** |
| MT-32 渲染出來的**有聲**長度 | **149.5 秒**(尾 1.3 秒是殘響)|

⇒ 差 21%。教訓:**估算式與實際渲染要對照過才算驗過**。
只看「`.mid` 能播」不會發現估算錯了 —— 那兩件事各自獨立。

## 3. 解析結果(15/15)

| 檔 | 音數 | 聲道 | GM 音色 | 長度 | ogg |
|---|---|---|---|---|---|
| `U5THEME` | 1,559 | 11 | 35 47 48 56 57 | 148.2 s | 2789 KB |
| `WRLDBLW` | 967 | 7 | 35 49 57 68 71 | 100.3 s | 1907 KB |
| `BLCKTHRN` | 879 | 7 | 14 19 32 35 48 57 | 108.6 s | 2149 KB |
| `ENGGMNT` | 791 | 7 | 32 35 57 62 68 71 72 | 71.9 s | 1459 KB |
| `STONES` | 759 | 3 | 24 88 | 149.6 s | 2808 KB |
| `BRITLAND` | 692 | 6 | 24 73 | 81.5 s | 1761 KB |
| `AMIGA` | 687 | 6 | 0 24 | 383.4 s | 5777 KB |
| `FANFARE` | 641 | 4 | 35 57 | 71.0 s | 1346 KB |
| `HORNPIPE` | 550 | 3 | 22 60 69 | 48.4 s | 959 KB |
| `LADYNAN` | 518 | 3 | 24 46 | 89.2 s | 1889 KB |
| `RULEBRIT` | 460 | 8 | 6 35 56 | 44.7 s | 826 KB |
| `GREYSON` | 428 | 7 | 24 32 73 | 60.8 s | 1223 KB |
| `MONARCH` | 423 | 7 | 6 34 68 | 53.6 s | 977 KB |
| `trntlla` | 418 | 5 | 24 35 56 57 | 53.2 s | 1257 KB |
| `HALLS` | 267 | 6 | 35 58 72 73 119 | 98.6 s | 1688 KB |
| `REUNION` | 128 | 6 | 35 56 57 | 14.3 s | 308 KB |

長度是**依 tempo 變化積分**算的(`duration_seconds`),不是 tick 數 ÷ 120 ——
見 §2③。ogg 那一欄是 MT-32 渲染出來的大小,合計 28 MB。

## 4. 渲染:MT-32 ROM(使用者提供,已放行)

音色編號是 **General MIDI 排列**(24 吉他 / 32 貝斯 / 48 弦樂 / 56 小號 / 73 長笛),
而 upgrade 附的驅動清單說明了目標硬體:

```
GENMID.ADD:  "MPU401-General MIDI synth"
MT32MPU.ADD / SC32MPU.ADD / SBAWE32.ADD / SBFM.ADV(AdLib 後備)
```

⇒ **音色住在玩家的音源硬體裡,不在光碟上。** 1992 年每個人聽到的音色取決於
自己買的卡(Roland MT-32 / SC-55 / AWE32 / AdLib),**沒有唯一正解**。
這是 `CLAUDE.md §3.0`(素材一律取自原版)碰到的一個真實邊界:
原版資料裡確實沒有這批曲子的音色。

**使用者 2026-08-08 提供 `~/cht/mt32` 的 Roland MT-32 ROM 並指定用它渲染。**
MT-32 是 upgrade 明確支援的四種硬體之一(`MT32MPU.ADD`),
所以這不是「隨便找一份 SoundFont」——是原作者當年支援的音源之一。

| | |
|---|---|
| 合成器 | **munt**(`mt32emu-smf2wav`,由 `munt_2_8_2` 在 `docker/Dockerfile` 內建)|
| ROM | `~/cht/mt32`(唯讀掛載;**不入 git、不進 image**)|
| 機型 | `--machine-id=mt32_1_07` |
| 取樣率 | MT-32 原生 **32 kHz** → 轉檔時升到 44.1 kHz |
| 輸出 | `assets/audio/mt32/*.ogg`(16 首,28 MB)|

⚠ 三個踩過的坑:

1. 安裝出來的執行檔叫 **`mt32emu-smf2wav`**(**連字號**),不是 `mt32emu_smf2wav`。
2. `-q` 在這個版本是**取樣率**不是 quiet ⇒ 用 `--quiet`。
3. cmake 要 `libglib2.0-dev` + `libsamplerate0-dev`,少了會停在 `Could NOT find GLIB2`。

驗收:每首的 ogg 長度 = MIDI 長度 + 3.5..5.0 秒殘響尾,而且**峰值非零**
(7,445..30,521)—— 兩個都要驗,只驗長度的話「整首靜音但長度對」會過關。

## 4b. 曲號對應與引擎切換

MT-32 這套的檔名是**曲名**,而遊戲配樂用的是**曲號**。
對應是靠**旋律比對**得來的(`tools/match_songs.py`),推導與交叉驗證見
`docs/re/98` §3,表在 `internal/u5data/mt32.go`。

引擎兩套並存,**遊戲中按 F5 切換**(`-music fmtowns|mt32` 可指定開場用哪套):

```
assets/audio/*.ogg        FM Towns:15 首 .EUP + 2 條 CDDA(YM2612 FM)
assets/audio/mt32/*.ogg   DOS upgrade:16 首 .XMI(Roland MT-32)
```

切換時**同一首會用新音源重播** —— 不重播的話按了 F5 什麼都不會發生
(曲號沒變),看起來像壞掉。只渲染了一套時預設就用那一套。

## 5. 重跑

```bash
# XMI → MID(單首;--check 只解析並印統計)
./tools/dev.sh python3 tools/xmi2mid.py gamedata/upgrade/HALLS.XMI build/mid/HALLS.mid

# 16 首一次渲染成 ogg(需要 MT-32 ROM)
MT32_ROM=~/cht/mt32 ./tools/mt32_render.sh

# 重新核對曲號對應(240 對旋律比對,正反兩向)
./tools/dev.sh python3 tools/match_songs.py
```

⚠ glob 一律**大小寫無關**(`nocaseglob` / `[xX][mM][iI]`)—— 見 §0。
⚠ `build/`、`assets/audio/`、MT-32 ROM 都 gitignore:轉出來的檔案衍生自
2001 年的 upgrade 包與商業 ROM,不入庫。
