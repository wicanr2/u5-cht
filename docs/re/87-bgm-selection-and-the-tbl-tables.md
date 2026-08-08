# 87 — 配樂是怎麼選的(而兩張 `.TBL` 不是我以為的那種表)

| | |
|---|---|
| 輸入檔 | `re_work/fmtowns/WORRIORS.EXP.asm`、`re_work/fmtowns/iso/U5_E/U5_BGM.TBL`、`U5_SE.TBL` |
| SHA-256 | 見 `re_work/fmtowns/SHA256SUMS.txt` |
| 主要函式 | `sub_3181C`(★ 播第 n 首)、`sub_313B8`(載入 `U5_BGM.TBL`)、`sub_31DC0`(★★ 開局挑曲)、`sub_31CB8`(當前曲號)、`sub_34685`(設聲道音量) |
| 起因 | C 階段(音樂)開工;`CLAUDE.md §4.3` 把 `U5_BGM.TBL` 列為「免逆向直接讀」 |
| 狀態 | 曲號 → 檔名 ✅ 定案;開局選曲 ✅ 定案並落地;15 首的**語意**部分待 DOSBox 實測 |

---

## 0. ⚠⚠ 先更正一條寫在 `CLAUDE.md` 裡的錯

`CLAUDE.md §2.4`(FM Towns 素材盤點)寫:

> **`U5_BGM.TBL` / `U5_SE.TBL` 是純文字表**(`M1.EUP 102 87 87 87 87 87`…)
> → 場景配樂對應**免逆向,直接讀表**

**「純文字表」對,「場景配樂對應」錯。** 兩張表的真正內容:

```
U5_BGM.TBL:  M1.EUP 102 87 87 87 87 87      ← 檔名 + **六個 FM 聲道的起始音量**
U5_SE.TBL:   WALK.SND 100 3032              ← 檔名 + 音量 + **檔案位元組數**
```

裡面**沒有任何一欄是場景**。「什麼時候播第幾首」寫在程式碼裡:32 個
`sub_3181C(n)` 呼叫點,加上 `sub_31DC0` 的一張地點跳表。

⚠ **錯因**:只看了表的**形狀**(檔名 + 幾個數字)就推語意,沒有追
「誰讀這張表、讀去做什麼」。這正是 `rulebook/62` 的反面 ——
**表的欄位語意不能從表本身讀出來,要從讀它的程式碼讀出來。**
`U5_SE.TBL` 那一半其實早就被追對了(`docs/re/` 的 `.SND` 筆記與
`internal/u5data/snd.go`:第三欄逐一對上 `stat`),但 `CLAUDE.md` 的
盤點沒跟著更新 ⇒ **錯的斷言活了下來,還被當成「C 階段很便宜」的依據。**

## 1. 六個音量是怎麼確定的

`sub_3181C` 換曲前先淡出當前那一首:

```asm
eax = dword_65334                       ; 當前曲號
eax = eax*8; eax = eax + eax*2          ; ★ 一列 24 byte
eax += offset dword_651B4               ; 表的基底
esi = eax; edi = &var_12C; ecx = 6; rep movsd   ; ★ 複製 6 個 dword
...
esi = 7Fh
loc_318BF:                              ; 淡出迴圈
    for (聲道 = 0..5)
        if (var_12C[聲道] > 0) { var_12C[聲道]--; sub_34685(聲道, 值) }
    忙等 0BB8h 次
    esi--
```

- **一列 24 byte = 6 個 dword**,而 `U5_BGM.TBL` 每行剛好 **6 個數字** ⇒ 一一對應。
- 六個值各自餵給 `sub_34685(聲道, 音量)`,聲道編號 **0..5** ⇒ 它們是**每聲道音量**。
- 淡出就是把六個一起遞減。⇒ 表裡的值是**起始音量**,不是曲長、不是場景。

## 2. 曲號就是表的列號(兩個獨立來源)

```asm
sub_3181C:
    call sub_31CB8                      ; 當前曲號
    mov  dword_65338, eax               ; 前一首 = 它
    cmp  [ebp+arg_0], 0FFFFFFFFh ; jz  → 停止
    and  eax, eax                ; jl  → …
    cmp  eax, 0Eh                ; jle → 繼續            ; ★ 上限 0x0E
```

**上限 0x0E ⇒ 曲號 0..14,共 15 首**;而 `U5_BGM.TBL` 剛好 **15 行**。
兩個獨立來源一致 ⇒ **曲號 n = 表的第 n 列**:

| 曲號 | 檔案 | 曲號 | 檔案 | 曲號 | 檔案 |
|---|---|---|---|---|---|
| 0 | `M1.EUP` | 5 | `M6.EUP` | 10 | `M11.EUP` |
| 1 | `M2.EUP` | 6 | `M7.EUP` | 11 | `M12.EUP` |
| 2 | `M3.EUP` | 7 | `M8.EUP` | 12 | `M13.EUP` |
| 3 | `M4.EUP` | 8 | `M92.EUP` | 13 | `M14.EUP` |
| 4 | `M5.EUP` | 9 | `M10.EUP` | 14 | `M152.EUP` |

⚠ 檔名的編號**不連續**(`M92` / `M152` 夾在中間),所以「曲號 = 檔名數字 − 1」
是錯的。必須讀表。

## 3. ★★ 開局挑哪一首:`sub_31DC0`

**唯一的呼叫者是 `sub_6730`(選單 / 啟動迴圈)**,位置就在載入 `UNDER.DAT`
與 `sub_2C2AC`(`initSoundEffect`)之間 ⇒ **它只跑一次,在讀檔完成之後。**

```
dword_65334 = 0 ; dword_65338 = 0
dl = byte_3DDB0
if (dl <= 0Fh) { sub_3181C(dl); dword_65334 = byte_3DDB0;
                 dword_65338 = byte_3EE18; return }   ; ★ 有有效曲號就直接播
ebx = byte_3E08C & 0FCh                                ; 載具碼
if (ebx == 20h || ebx == 24h || ebx == 28h) → 曲 2      ; ★ 揚帆 / 大船 / 小艇
if (byte_3E0A3 >= 21h) → 曲 9                          ; ★ 地牢(含戰鬥)
esi = 0; while (esi < 20h && !(byte_410F4[esi] == 隊伍X && byte_4111C[esi] == 隊伍Y)) esi++
edi = (esi < 20h) ? esi : 0                            ; 地點索引,找不到就 0
switch (edi − 10h) { … }                               ; ★★ 跳表
default → 曲 1
```

### 跳表 `jpt_31E71`

⚠ 定義域是 `地點索引 − 0x10`,而**地點表只有 32 筆** ⇒ **只有下標 16..27
走得到**。跳表裡 case 57(曲 0x0B)與 case 62(曲 7)是編譯器依
`cmp edx, 2Eh` 補出來的 —— **走不到**,所以 `M12.EUP` 在這條路上是死的。

| 地點索引 | 地點 | 曲號 | 檔案 |
|---|---|---|---|
| 16 | `CASTLE.DAT` 1(樓層 −1..3)= 不列顛王的城堡 | 8 | `M92.EUP` |
| 17 | `CASTLE.DAT` 6(樓層 −1..3)= 黑刺的城堡 | 4 | `M5.EUP` |
| 18 | WEST BRITANNY | 8 | `M92.EUP` |
| 19 | NORTH BRITANNY | 8 | `M92.EUP` |
| 20 | EAST BRITANNY | 5 | `M6.EUP` |
| 21 | PAWS | 6 | `M7.EUP` |
| 22 | COVE | 9 | `M10.EUP` |
| 23 | BUCCANEER'S DEN | 9 | `M10.EUP` |
| 24 | ARARAT | 9 | `M10.EUP` |
| 25 | BORDERMARCH | 4 | `M5.EUP` |
| 26 | FARTHING | 5 | `M6.EUP` |
| 27 | WINDEMERE | 12 | `M13.EUP` |
| 其餘 | 八座大城 / 民居 / 燈塔 / 石門 / 學院 / 修院 / 蛇之堡 | 1 | `M2.EUP` |

★ 涵蓋範圍**剛好**是 `CASTLE.DAT` 的八個地點(16..23)加 `KEEP.DAT` 的前四個
(24..27)。這不是巧合 —— `TestJumpTableCoversCastleAndKeepOnly` 把它釘住:
地點表的順序哪天被動到,會先紅在這裡,而不是等到「音樂配錯地方」。

### ⚠ 一個看起來像 bug 的結果,以及為什麼先不修

八座大城(`TOWNE.DAT`,下標 0..7)落到 default ⇒ **存檔在不列顛城裡,開局會先響
大地圖的曲子**。但另一條路解釋得通:存檔裡的 `byte_3DDB0` 若是有效曲號就
**直接播它**,跳表只是「沒有有效曲號時」的推導 —— 也就是**新遊戲**與
**壞掉的存檔**。哪一條實際生效,`byte_3DDB0` 到底進不進存檔,
**要對 DOSBox 實測**(從城裡存檔、重開、聽第一秒)。已列進 A 階段。

照原樣實作(`CLAUDE.md §3.0`),不自行「修正」成「城裡就放城鎮曲」。

## 4. 32 個呼叫點:一首曲子服務好幾個場合

⚠ **不要給每首曲子貼一個語意標籤。** 15 首要服務整個遊戲,原作者本來就重用;
硬貼標籤會把「Rest 用曲 4」讀成「曲 4 是黑刺城堡的曲子所以 Rest 在黑刺城堡」。

| 曲號 | 出現的場合(全部呼叫點) |
|---|---|
| 0 | `sub_A9EC` 印完 `"\nVICTORY!\n"` 之後 |
| 1 | `sub_31DC0` default、`sub_DF84`(月門傳送)、`sub_3FE4`(離開地牢)、`sub_86C`(方向樓梯)、`sub_1DA10`、`sub_177AC`、`sub_3FE4` |
| 2 | `sub_31DC0` 載具是船、`sub_16F08`(上船) |
| 3 | `sub_A9EC` 開頭(戰鬥迴圈) |
| 4 | `sub_165C8`(紮營突襲)、`sub_16BA0`、`sub_21D48`(Rest / 住宿)、`sub_1DA10`、跳表 17 / 25 |
| 5 | `sub_1884`(被逮捕,`"\nThou dost awaken to...\n"`)、跳表 20 / 26 |
| 6 | 跳表 21 |
| 7 | `sub_2D72C`(依腳下 tile 進場景)、`sub_1678`、`sub_C778`、跳表 case 62(走不到) |
| 8 | 跳表 16 / 18 / 19 |
| 9 | `sub_2D564`(洞穴 / 礦坑 / 地牢)、`sub_135FC`、地點碼 ≥ 0x21、跳表 22–24 |
| 0Ah | `sub_32244` |
| 0Bh | 只出現在跳表 case 57 ⇒ **這條路走不到**;有沒有別的路播它未知 |
| 0Ch | 跳表 27 |
| 0Dh | `sub_135FC` |
| 0Eh | `sub_12B20` |

⇒ 站得住腳的只有這幾條(有兩個以上獨立呼叫點互相佐證):
**1 = 大地圖**、**2 = 船**、**3 = 戰鬥**、**7 = 進場景(城鎮)**、**9 = 地牢**。
其餘先當「某某場合會播第 n 首」記著,不寫語意。

## 5. 引擎對應

| 原版 | 引擎 | 狀態 |
|---|---|---|
| `U5_BGM.TBL` 15 列 + 六聲道音量 | `u5data.LoadBGMTable` / `BGMTrack` | ✅ |
| 曲號 0..14 的上限 | `u5data.BGMSongCount` | ✅(與表列數互相驗證) |
| `sub_31DC0` 開局選曲 | `u5data.StartupSong` | ✅ |
| `U5_SE.TBL` + 25 個 `.SND` | `u5data.LoadSoundTable` / `LoadSoundSet` | ✅ **早就做完了**(見 §6) |
| `sub_2D72C` 進場景 → 曲 7 | `Enter()` | ✅ |
| `sub_86C` 離場 → 地表 1 / 幽冥界 0x0A | `leaveScene()` + `overworldSong()` | ✅ |
| `sub_2D564` 進地牢 → 曲 9 | `EnterDungeon()` | ✅ |
| `sub_3FE4` 離開地牢 → 字面 1 | `LeaveDungeon()` | ✅(含「出到幽冥界也放地表曲」的原版怪處) |
| `sub_DF84` 月門傳送 → 字面 1 | `TravelByMoongate()` | ✅ |
| `sub_A9EC` 進戰鬥 → 3 / 勝利 → 0 | `beginCombatFrom()` / `beginRoomCombat()` / `checkCombatOver()` | ✅ |
| `sub_16F08` 上船 / 上小艇 → 曲 2 | `Board()` | ✅(馬與魔毯不換) |
| 曲號的讀取介面 | `CurrentSong()` / `PreviousSong()` | ✅(⚠ 目前只有測試引用,見下) |
| `sub_177AC` 下載具 → 字面 1 ×2 | — | ⬜ **刻意不接**:兩個呼叫點還沒逐一讀,而 `dismount` 在**場景裡**也會被呼叫(在城裡下馬)⇒ 接錯會在城裡放大地圖的曲子。有位址,沒讀完就不做(`CLAUDE.md §3.0`) |
| `sub_165C8` 紮營 / `sub_21D48` Rest → 4,之後回 `[ebp+var_8]` | — | ⬜ 「放完回到剛才那首」的機制未讀 |
| `sub_1DA10` 聖壇 → 1 與 4 兩處 | — | ⬜ 兩處條件未分辨 |
| `sub_1678` / `sub_C778` → 曲 7 | — | ⬜ 與 `sub_2D72C` 可能重複,未確認 |
| `sub_12B20` → 0x0E、`sub_135FC` → 0x0D / 9、`sub_32244` → 0x0A | — | ⬜ 過場 / 開場,引擎沒有對應流程 |
| `sub_3181C` 的六聲道淡出 | — | ⬜ 引擎不做 FM 合成,淡出改成音量包絡 |
| `.EUP` → 可播的音訊 | — | ⬜ **C 階段主體**;`../u1-cht/tools/render_eup_music.py` 可複用,但那是**自製 2-op FM 近似**,不是原版音色(EUP 檔頭 0x254..0x6D2 疑為音色定義,未解)|
| 兩條 CDDA → ogg | — | ⬜ |
| 播放本體 | `internal/audio` | ⬜ **套件還不存在** |

## 6. ⚠⚠ 兩個「差點重做一遍」與一個假綠燈

### (a) `.SND` 與 `U5_SE.TBL` 早就做完了

這一輪先寫了 `LoadSETable` / `SoundEffect`,**然後才 grep 到 `snd.go` 裡已經有
`LoadSoundTable` / `SoundTableEntry`,而且解得更好** —— 它處理了檔尾的
`0x1A`(DOS 的 Ctrl-Z EOF),我那份沒有,**跑起來會直接失敗**。已刪掉重複的那份。

⇒ `CLAUDE.md` 的規則早就寫了:**「要說某件事沒做之前,先 grep 自己的
`docs/` 與程式碼」**。這次是在寫完之後才 grep,只賠了十分鐘;
但它同時解釋了 §0 那條錯斷言為什麼活著 —— **盤點與程式碼各自更新,就會漂。**

### (b) ★★ 五條 `.SND` 測試從來沒跑過

`snd_test.go` 有一份自己的 `fmtownsDir`,把 `U5_FMTOWNS` 當成 **`U5_E` 目錄本身**;
而 `tools/dev.sh` 與其他測試(`tiles_test` / `tlk_test`)都把它當成 **ISO 根目錄**。
⇒ 在 gate 裡它永遠 `Skip`,而 **skip 不會讓 gate 變紅**:

```
--- SKIP: TestSoundTableSizesMatchTheFilesAndTheHeader
--- SKIP: TestSoundPCMIsSignMagnitudeNotTwosComplement
--- SKIP: TestOnlyAmbientSoundsLoop
--- SKIP: TestLoopFieldsAreConsistentForEveryFile
--- SKIP: TestBaseNoteIsSixtyOrSixtyOne
```

也就是說「`.SND` 格式已驗證」這句話,**在自動化裡一次都沒被驗證過**。
(`diagnosis-notes/03-silence-is-not-success`:沉默相容於五個世界。)
已統一成 `tiles_test.go` 的 `fmTownsDir`,五條測試現在真的跑,並且全綠。

⇒ 順手加的一條紀律:**看到 `--- SKIP` 不要當成「沒資料所以跳過」就算了** ——
先問「這台機器上明明有資料,為什麼跳過」。

## 7. 引擎的分工:`internal/game` 只決定曲號

`State.song` / `prevSong` 對應原版的 `dword_65334` / `dword_65338`,
讀取走 `CurrentSong()` / `PreviousSong()`;音訊由 `internal/audio` 拿這個值去播。
⇒ **headless 完全不需要音效裝置**,曲號切換可以在單元測試裡驗
(同 `render` 不綁 GPU 的理由,`docs/engineering-notes.md`)。

⚠ **`song` 刻意不匯出。** 曲號 **0 是勝利那一首**,而 `State` 到處被寫成
結構常值(這個專案沒有 `State` 的建構子)⇒ 匯出的話**零值會被當成
「正在播勝利曲」**。`TestFreshStateHasNoSong` 把這個陷阱釘住。
