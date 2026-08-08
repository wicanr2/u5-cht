# 13 — XMI(Miles / AIL)→ 標準 MIDI

| | |
|---|---|
| 輸入檔 | `gamedata/upgrade/*.XMI`(15 個,Exodus Project MIDI Upgrade 1.0,2001)|
| 工具 | `tools/xmi2mid.py`(純 Python,無外部依賴)|
| 驗收 | 15/15 解析成功、檔頭與軌長自洽;**含反對照**(餵非 XMI 會報錯)|
| 狀態 | ✅ 格式轉換完成;⚠ **渲染成 ogg 受阻於素材規則**,見 §4 |

---

## 0. ⚠ 更正:是 15 首不是 19 首

此前五處文件寫「19 首 `.XMI`」。一手驗證:`gamedata/upgrade/` 下有 **15 個 `.XMI`**,
每個檔恰好一個 `EVNT` chunk(XMI 是容器,一個檔可以裝多首,所以要數 chunk 不是數檔案):

```
AMIGA BLCKTHRN BRITLAND ENGGMNT FANFARE GREYSON HALLS HORNPIPE
LADYNAN MONARCH REUNION RULEBRIT STONES U5THEME WRLDBLW
```

⇒ **15 首**,而且與 FM Towns 的 15 首 `.EUP` **是同一批曲子的另一個版本**。

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

### ③ 時基固定 1/120 秒

XMI 的一個 tick 是 AIL 的硬體節拍 = 1/120 秒。輸出用
**ppqn = 60 + tempo 500,000 µs/四分音符** ⇒ 60 ticks / 0.5 s = 120 ticks/s,實際速度相同。
曲子裡自帶的 `FF 51`(set tempo)照抄 —— 原版驅動也是照著跑的。

## 3. 解析結果(15/15)

| 檔 | 音數 | 聲道 | GM 音色 | 長度 |
|---|---|---|---|---|
| `U5THEME` | 1,559 | 11 | 35 47 48 56 57 | 123.5 s |
| `STONES` | 759 | 3 | 24 88 | 144.8 s |
| `WRLDBLW` | 967 | 7 | 35 49 57 68 71 | 83.5 s |
| `BLCKTHRN` | 879 | 7 | 14 19 32 35 48 57 | 95.1 s |
| `ENGGMNT` | 791 | 7 | 32 35 57 62 68 71 72 | 59.9 s |
| `BRITLAND` | 692 | 6 | 24 73 | 67.9 s |
| `AMIGA` | 687 | 6 | 0 24 | 99.8 s |
| `FANFARE` | 641 | 4 | 35 57 | 59.2 s |
| `HORNPIPE` | 550 | 3 | 22 60 69 | 40.3 s |
| `LADYNAN` | 518 | 3 | 24 46 | 78.1 s |
| `RULEBRIT` | 460 | 8 | 6 35 56 | 37.3 s |
| `GREYSON` | 428 | 7 | 24 32 73 | 50.6 s |
| `MONARCH` | 423 | 7 | 6 34 68 | 44.7 s |
| `HALLS` | 267 | 6 | 35 58 72 73 119 | 86.3 s |
| `REUNION` | 128 | 6 | 35 56 57 | 11.9 s |

## 4. ⚠ 渲染成 ogg 受阻:GM 音色**不在原版資料裡**

音色編號是 **General MIDI**(24 吉他 / 32 貝斯 / 48 弦樂 / 56 小號 / 73 長笛 / 119 合成鼓),
而 upgrade 附的驅動清單自己說明了目標硬體:

```
GENMID.ADD:  "MPU401-General MIDI synth"
MT32MPU.ADD / SC32MPU.ADD / SBAWE32.ADD / SBFM.ADV(AdLib 後備)
```

⇒ **GM 音色住在玩家的音源硬體裡,不在光碟上。** 1992 年每個人聽到的音色
取決於自己買的卡(Roland SC-55 / MT-32 / AWE32 / AdLib),**沒有唯一正解**。

要渲染成 ogg 只有兩條路,兩條都需要**原版資料以外的東西**:

| 路線 | 需要 | 與 `CLAUDE.md §3.0` 的關係 |
|---|---|---|
| A. GM SoundFont | 一份第三方音色庫(數十 MB)| 音色不是原版資料 ⇒ **要使用者放行** |
| B. OPL2 + `SBFM.ADV` 裡的音色庫 | 逆向那支 14 KB 的 Miles 驅動 + OPL2 合成 | 音色來自原版,但那是**驅動**(使用者已說不用管驅動)|

⇒ 目前**停在 `.mid`**:格式工作完成且可重跑,渲染那一步等決定。

★ 重要的是**遊玩音樂已經完整** —— FM Towns 的 15 首 `.EUP` + 兩條 CDDA 共 17 個 ogg
早就在跑(`docs/formats/12`)。XMI 是**同一批曲子在別的音源上的第二個版本**,
性質等同四種顯示模式:是主題選項,不是缺的內容。

## 5. 重跑

```bash
for f in gamedata/upgrade/*.XMI; do
  ./tools/dev.sh python3 tools/xmi2mid.py "$f" "build/mid/$(basename "$f" .XMI).mid"
done
```

`--check` 只解析並印統計。⚠ `build/` 與 `assets/audio/` 都 gitignore ——
轉出來的檔案衍生自 2001 年的 upgrade 包,不入庫。
