# Ultima V: Warriors of Destiny — Golang + Ebiten 重製 + 繁體中文化

> 專案憲章。**每個 session 動手前先讀本檔**，再依 §0 的觸發表載入對應規則。
> 立案:2026-08-06 ・ 維護:L.CY (anr2) + Claude
> 方法論母本:`~/.claude/knowledge-base/retro-cht/retro-game-remake/SKILL.md`(反編當 oracle,乾淨重寫)

---

## 0. 動手前的載入清單(不要憑記憶跳過)

| 要做什麼 | 先載入 |
|---|---|
| **任何** 逆向 / 資料格式 / 中文化工作 | `Skill: re-retro-cht-rulebook`(路由表) |
| **開 IDA 之前** | `~/.claude/knowledge-base/retro/ida-pro-9.4.md` + 本檔 §4 |
| 斷言「某欄位是 X / 這值從哪來」之前 | `rulebook/62-static-provenance-trace.md` |
| 逆向撞牆(看不出來 / 想動態 / 想 DOSBox) | `rulebook/64-re-screenshot-oracle.md` |
| 宣稱某階段「完成」之前 | `rulebook/65-verify-against-reference-not-internal-signals.md` |
| 寫 / 大改 README | `rulebook/80-retro-cht-readme-polish.md` + 本檔 §7 |
| 決定中文字尺寸 / 畫布 | `rulebook/81-retro-cjk-hires-canvas.md` + `retro-cht/eten-bitmap-font/SKILL.md` |
| 做打包後的可玩性驗收 | `retro-cht/retro-game-playtest/SKILL.md` |
| 除錯 / 找 bug | `rulebook/60-feedback-loop-priority.md` |
| 特例越補越多、「再差一點」卡很久 | `rulebook/41-whack-a-mole-stop-rethink.md` |
| 規劃模組介面 | `rulebook/70-deep-modules.md` |

---

## 1. 目標與路線

把 1988 年 DOS 版《Ultima V: Warriors of Destiny》**用 Go + Ebitengine 從零重寫成跨平台可玩的引擎,並完整繁體中文化**。

```
IDA Pro 反組譯 ──┬─ ★ FM Towns WORRIORS.EXP(32-bit P3,可能可反編譯)  ← 遊戲邏輯主目標
                 ├─ ★ WORRIORJ.EXP ⊖ WORRIORS.EXP(日/英 diff)      ← 中文化 hook 點
                 └─ DOS ULTIMA.EXE + *.OVL                          ← overlay 機制、玩家最普遍
                       (只當行為真值 oracle,不照抄)  ─────────────┐
破解原版資料格式(地圖/對話/tileset/存檔),主線讀 DOS 格式          ┼─▶ 手寫乾淨 Go + Ebiten 引擎
  └ FM Towns EGA*.TIL / PC-98 TILES.CH 未壓縮素材當 oracle          │   (deep modules + 倚天點陣 CJK)
  └ U5_J/*.JPN(日)⇄ *.TLK(英)靠 index 逐筆對齊當翻譯 oracle       │
原版資料檔(玩家自備,不散布)──────────────────────────────────────┘
```

**與 u1-cht / u4-cht 的路線差異(重要範圍認知)**

| | u1-cht | u4-cht | **u5-cht(本專案)** |
|---|---|---|---|
| 基礎 | fork `open_ultima`(C++/SDL2) | patch `xu4`(C++/Allegro) | **無上游可用,Go 從零寫** |
| 中文化手法 | 改上游文字出口 | load-time 查表替換 | **自己設計文字管線,一開始就 UTF-8** |
| 逆向角色 | 補完上游缺的行為 | 對齊 `AVATAR.EXE` 數值 | **唯一的行為真值來源** |
| 主要風險 | 上游極早期 | patch 維護 | **引擎規模(全遊戲邏輯都要自己寫)** |

⇒ 本專案的工作量在三者中最大,且**逆向是主線不是支線**。不要用「先做中文化、引擎慢慢補」的順序思考;
順序是「解資料 → 逆行為 → 寫引擎(一開始就吃 UTF-8) → 補譯文」。

---

## 2. 素材盤點(`org_game/`,已一手驗證:DOS 與 PC-98 於 2026-08-06、FM Towns 於 2026-08-07)

**⚠ 下表事實由實際開檔驗證得出,不是推論。要推翻任一條,先拿出一手證據(rulebook/62)。**

### 2.1 DOS 版本體 `Ultima_V_-_Warriors_of_Destiny_1988.zip`(322 檔)

| 檔案 | 已驗證事實 | 中文化 / 逆向定位 |
|---|---|---|
| `ULTIMA.EXE`(36K) | MZ 16-bit real mode,409 個 relocation | IDA 主目標,遊戲主迴圈 |
| `*.OVL`(24 個) | **裸機器碼,無 MZ 檔頭**(開頭即 `55 8B EC` = `push bp; mov bp,sp`) | IDA 需**手動當 16-bit binary 載入並指定 segment**,見 §4.2 |
| `DATA.OVL`(48K) | 資料 overlay。含 `MS Run-Time Library ... 1988, Microsoft Corp` → **編譯器是 Microsoft C 5.x/6.0(cdecl)**;明文物品名表(`Leather Helm`…`Mystic Sword`、`Ring of Invisibility`) | 裝備/法術/物品名譯源;⚠ 呼叫慣例是 cdecl,**不是** Turbo Pascal(kb 裡的 Pascal 慣例別套過來) |
| `*.TLK` ×4(`TOWNE` 26K / `CASTLE` 21K / `KEEP` 17K / `DWELLING` 8K) | NPC 對話。檔頭是 `(u16 offset, u16 index)` 索引表;**文字每 byte 最高位元被設為 1**(清 bit7 後即明文:`Zachariah`、`a stately, white-haired`、`Welcome, \x01,`);內嵌控制碼(`\x01` 疑為玩家名代入、`\x03`) | **對話中文化主戰場**。索引表 base 與控制碼語意需以 IDA 確認,勿猜 |
| `STORY.DAT` `QUESTION.DAT` `KARMA.DAT` `MISCMSG.DAT` `ENDMSG.DAT` | **明文 ASCII,記錄用 NUL(0x00)分隔**;`{` 是段落起始(僅 STORY/QUESTION);**`_` 是英文斷字提示**(`be_gin`、`mys_te_ri_ous`)。實測筆數 20 / 30 / 6 / 47 / 11,**詞典 token 0 個 → 可直接翻譯** | 開場故事 / 吉普賽問答 / 業報 / 系統訊息 / 結局。⚠ 譯成中文時 `_` 要移除。<br>⚠⚠ **更正**:此前記載「`\|` 分隔」是錯的 —— `STORY.DAT` 裡 0x7C 出現 0 次。錯因是拿 `strings … \| tr '\n' '\|'` 的輸出當檔案內容(`\|` 是自己的 `tr` 加的)。**一手位元組贏二手推論** |
| `SHOPPE.DAT`(10,135 B) | 194 筆(NUL 分隔),但含 **862 個詞典 token**:位元組 ≥ 0x80 不是文字而是常用詞代碼(`Thanks\x86nothing!` → `\x86`="for" → "Thanks for nothing!")。字典在 **`DATA.OVL` 0x104C**,118 個常用詞(`the thou of to and that for…`,含 `Blackthorn`/`Shadowlords`/`Mantra`) | 商店對白。⚠ **token → index 的精確映射未定**(`for` 對得上 index 6,但另兩個 token 差 10;token 空間 128 而字典只有 118 → 有 10 個是控制碼)。**在映射定案前不得進翻譯流程**,留 P3 讀反編譯碼 |
| `LOOK2.DAT`(3,622 B) / `SIGNS.DAT`(8,364 B) | 格式與上面兩類都不同:`LOOK2` 有 218 個 NUL + 大量 0x01–0x1F 控制碼;`SIGNS` 是 u16 offset 表 + 用字元畫的框線(`8lllllmllmlllll9`)+ 地名 | 觀察敘述 / 城鎮招牌。各自另案處理 |
| `SIGNS.DAT` | 明文,但混著用字元畫的框線(`lllm`、`^_@`) | 城鎮招牌。同 u4-cht「美術內嵌字母」問題,先評估再決定譯不譯 |
| `IBM.CH` / `RUNES.CH` | **各 1024 B = 128 glyph × 8 B,8×8 點陣,索引 = ASCII 碼直接對應**(已 dump 驗證 `A`/`@`/`!`) | 原版字型。**8×8 沒有 CJK 空間 → 必須另開中文點陣路徑**(§5) |
| `IBM.HCS` | 3072 B = 128 glyph × 24 B(Hercules 高解析字型,實際尺寸待驗) | 單色顯示模式字型 |
| `TILES.16` / `TILES.4` | **前 4 byte 是解壓後長度**(`TILES.16` = 65536 = 512 tile × 16×16×4bpp);檔案 30414 B → **有壓縮(疑 LZW,待破)** | EGA / CGA tileset。P1 第一個要破的格式 |
| `BRIT.DAT`(52K) / `UNDER.DAT`(64K) | 地表 / 地底世界地圖 | 世界地圖 |
| `TOWNE.DAT` `CASTLE.DAT` `DWELLING.DAT` `KEEP.DAT`(各 16K) | 城鎮 / 城堡 / 民居 / 要塞地圖 | 場景地圖 |
| `*.NPC`(各 4608 B) | NPC 定義(對應同名 `.TLK`) | NPC 排程/座標 |
| `BRIT.CBT` / `DUNGEON.CBT` | 戰鬥地圖 | 戰鬥系統 |
| `*.PTH` / `*.OOL` | 路徑 / 疑為物件表 | 待逆 |
| `DNG1-3.16/.4`、`MON0-7.16/.4`、`STORY1-6.16/.4`、`CREATE/END/ITEMS/TEXT/ULTIMA.16/.4` | 各畫面圖組,`.16`=EGA / `.4`=CGA | 素材解碼 |
| `EGA.DRV` `CGA.DRV` `HER.DRV` `T1K.DRV` | 顯示驅動(EGA/CGA/Hercules/Tandy) | **原版就有四種顯示模式** → 顯示模式切換的素材依據 |
| `SAVED.GAM` `INIT.GAM` | 存檔 / 初始狀態 | 存檔格式;可做「匯入原版存檔」 |

### 2.2 `upgrade/` = MIDI 音樂升級包,**不是 VGA 美術**

一手證據:`Readme.txt` 自述為 **The Exodus Project — Ultima V Upgrade Release 1.0(2001-08-21)**,
目標是「add MIDI music to the game」;`upgrade/TILES.16` 與原版 **md5 完全相同**。

- **可用價值**:19 首 `.XMI`(`U5THEME` `BRITLAND` `STONES` `WRLDBLW` `RULEBRIT` `MONARCH` `GREYSON` `LADYNAN` `ENGGMNT` `REUNION` `HALLS` `FANFARE` `HORNPIPE` `AMIGA` `BLCKTHRN` `SETM` `TRNTLLA`…)+ AIL/MIDPAK 驅動。→ **離線轉 ogg**(XMI→MID→fluidsynth+SoundFont),不跑模擬器。
- **意外的 oracle**:它改了 8 個檔(`ULTIMA.EXE` `DATA.OVL` `FONT.OVL` `INTRO.OVL` `TOWN.OVL` `DUNGEON.OVL` `ENDGAME.OVL` `MAINOUT.OVL`)。**原版 vs patch 版逐 byte diff = 現成的 overlay 結構線索**(哪裡有空隙、呼叫怎麼掛)。作者 mcmagi 即 u3-cht 用的那份 U3 反組譯文件作者,其公開文件同為參考來源。
- ⚠ **不要**把 upgrade 版當基準本體。基準是原版;upgrade 只供音樂與 diff。

### 2.3 手冊 `org_game/manuel/`

《軟體世界》雜誌「說明書補完計劃」057-創世紀第V代 上/下,共 **61 張 JPG** 掃描(38 MB)。

### 2.4 日文版(PC-98 / FM Towns)——**逆向的首選入手點**(使用者指示 2026-08-06)

#### PC-98 版(`【PC98】ＵｌｔｉｍａⅤ.rar`,已抽檔驗證)

兩張 `.fdi`,各 1,265,664 B = **4096 B FDI 檔頭 + 1,232,896 B raw**;
內容是 **FAT12,磁區 1024 B、192 個根目錄項**。抽檔工具**已可用**:
`python3 tools/fdi_extract.py -l <映像>` 列目錄、`… <映像> <輸出目錄> [檔名…]` 抽檔(已實測兩張磁碟)。

| 已驗證事實 | 對本專案的意義 |
|---|---|
| **`U5.EXE` 單一檔 216,880 B**(DOS 版是 36K EXE + 24 個獨立 `.OVL`) | 程式碼集中在一個檔 |
| MZ 檔頭只宣稱 **40,896 B**,其後 **176 KB 是附加的 overlay 區**;`push bp;mov bp,sp` 在附加區出現 **467 次**、主段僅 138 次 | ⚠ **多數程式碼在 IDA 自動分析範圍外**;而且附加區是連續無邊界標記的一大塊 → **切 overlay 反而比 DOS 版難**(DOS 版每個 `.OVL` 是獨立檔案,邊界天然清楚) |
| **沒有 CodeView 符號表**(NB05/08/09/10/11 全數不存在) | ⚠ **「日文版有 symbol table」的假設不成立**。日文版的價值在下面三條,不在符號 |
| 內含 `Microsoft` / `MS Run-Time Library` 字串 | 與 DOS 版**同一編譯器族(MS C)** → 兩版程式碼結構同源,**可互相對照**;呼叫慣例同為 cdecl |
| **`TILES.CH` 98,304 B 未壓縮**(98304 = 512×192 = 768×128,實際切法待驗);DOS 版 `TILES.16` 是壓縮的 30,414 B | ★ **可當「已知輸出」反推 DOS 版壓縮格式**(`rulebook/64` 的 oracle 手法,但比截圖更硬:同源未壓縮資料)。也讓 P1 可以先用未壓縮素材把畫面跑起來,不必等壓縮破解 |
| **`STORY.DAT` 是 Shift-JIS 明文日文**,已成功解碼(`月の輝く晴れた空に,どこからともなく煙のような雲の塊が出てきた`);對應 DOS 版 `From no_where, smoky wisps of clouds be_gin to form…`。且**日文版沒有 `_` 斷字標記** | ★ **逐句雙語 oracle**(英↔日),翻譯時可交叉理解語意;並證明「2-byte 字元 + 無斷字」的排版路徑在引擎中已存在 |
| `.TLK` 索引表結構**與 DOS 版不同**:PC-98 是**純 u16 offset 陣列**(`30 00`, `64 00`, `84 03`, `71 06` …),DOS 是 `(u16 offset, u16 index)` 交錯;bit7-set 比例僅 20.7% → **日文版沒有用 bit7 混淆文字**(Shift-JIS 本身就佔高位元) | 兩版 TLK 互為結構對照;日文版索引更單純 |
| `FONT98.CH` / `RUNES98.CH` **各 4096 B**(DOS 版 `IBM.CH` 是 1024 B)→ 半角字型擴充;全角日文走 PC-98 內建字型 ROM | ★ **引擎已內建 DBCS 顯示路徑** → 中文化可沿用其機制,不必自己發明(見 §5.2) |
| `OPNDRV.COM`(7,826 B)+ `BEPDRV.COM`;`UL01–UL15.BIN`(171–1,279 B,自訂序列,開頭 `fd 04 0c 00…`) | **YM2203(OPN)FM 音源** + 15 首曲譜。後補階段可還原(參考 u4-cht 的 YM2151 / u1-cht 的 MCP/SCORE 經驗) |
| 資料檔名整組不同:`.CH` / `.DAT`(`NPC.CH` 131,072、`DNG1-3.DAT` 各 ~150 KB、`ITEMS.DAT`、`GEM.DAT`、`VIEWMAP.DAT`) | ⚠ **PC-98 是另一套資料格式**,不可假設與 DOS 版相同;繪圖是 GDC 4-plane,與 EGA 不同 |

#### FM Towns 版 ★ **逆向主目標**(`org_game/fmtown/`,已於 2026-08-07 完整驗證)

`Ultima V - Warriors of Destiny (Japan).7z`(91,454,221 B)= **CD-ROM 映像**:
`MODE1/2352` 資料軌(31,399,200 B = 13,350 sectors)+ **兩條 CDDA 音軌**(180 秒 / 371 秒)+ `.cue`。
資料軌轉 ISO:每 sector 取 offset 16..2064(`tools/cdimg_to_iso.py`),得 27,340,800 B ISO9660,
**239 檔 / 4 目錄**;之後用 `7z x` 直接抽檔(7z 認 ISO9660)。

| 已驗證事實 | 對本專案的意義 |
|---|---|
| **`RUN386.EXE` + `.EXP`,magic = `P3`**(不是 `MZ`)→ **Phar Lap 386\|DOS-Extender,32-bit 保護模式** | ★★ **三版中唯一可能吃到 Hex-Rays 反編譯器的路**(16-bit real mode 永遠不支援)。載入 IDA 時選 32-bit / Phar Lap loader,**第一步先實測 Hex-Rays 是否真的能反編譯這個 P3** |
| **`U5_E/WORRIORJ.EXP`(500,375 B,日文)與 `U5_E/WORRIORS.EXP`(475,719 B,英文)在同一張光碟** | ★★ **同版本、同編譯器、同資料的雙語執行檔**。兩者 diff ≈ 24.7 KB,差的正是**日文 DBCS 字型與排版邏輯** → 這是「中文化該 hook 哪裡」的直接答案,不用猜 |
| 仍**沒有 CodeView 符號**(兩個 `.EXP` 都掃過 NB05/08/09/10/11) | 「日文版有 symbol table」在三版都不成立。優勢來自反編譯器與雙語對照,不是符號 |
| **資料檔沿用 DOS 版**:`BRIT.DAT` 52,480、`CASTLE.TLK` 21,868、`BRIT.CBT` 5,632、`CASTLE.NPC` 4,608、`BRITISH.PTH` 2,783 —— **與 DOS 版逐一同大小** | ★★ 逆 FM Towns 得到的邏輯**直接適用於 DOS 版資料格式**(不像 PC-98 是另一套)。主線引擎讀 DOS 格式的決定不變 |
| **`U5_J/*.JPN` 四份日文對話**:`TOWNE`(40,323)/`CASTLE`(31,178)/`KEEP`(26,122)/`DWELLING`(11,427),對應英文 `.TLK` 四份;**檔頭結構與 DOS 版相同**(`30 00 01 00`,`c2 00 02 00`,… = `(u16 offset, u16 index)` 交錯,前兩筆位元組完全一致);內容是 Shift-JIS 明文,已解出「よくぞ来られた。星の研究をしておる。よい旅をな!」 | ★★ **靠 index 欄逐筆英↔日對齊的翻譯 oracle**;且證明 TLK 格式可容納 2-byte 文字(英文版全 bit7-set,日文版直接放 Shift-JIS) |
| 另有 `U5_E/*.JPN` 六份:`END` `ENDMSG` `KARMA` `LOOK2J` `MISCMSG` `SHOPPE` | 連系統訊息與商店都有英日對照 → **§2.1 的八類文字全部有第二語言可校** |
| **`EGA0–EGA3.TIL` 各 65,536 B 未壓縮**;DOS `TILES.16` 的檔頭宣稱解壓後**正是 65,536** | ★ **極可能就是 DOS 壓縮 tileset 的解壓結果** → 拿它當 oracle,破 `TILES.16` 壓縮時有逐位元組對答案 |
| `U5.FNT` 16,384 B、`TOWNS.FNT` 17,160 B | 字型檔。⚠ `U5.FNT` 以 8×16 ASCII 直索引 dump 出來**不是字形**(idx 65 是橫條紋)→ 佈局待 P1 驗,**不要假設同 DOS 版的 `IBM.CH`** |
| **`U5_BGM.TBL` / `U5_SE.TBL` 是純文字表**(`M1.EUP 102 87 87 87 87 87`、`M2.EUP 102 97 …`) | ★ 場景配樂對應**免逆向,直接讀表** |
| **`M1–M152.EUP` 15 首**(EUPHONY 序列)+ 兩條 CDDA | 遊玩音樂是 EUP、非 CDDA(kb 已知陷阱);EUPHONY 格式 u1-cht 逆過,**可複用** |
| `.TIF` 49 個,固定 154,112 B = 512 B 檔頭 + 320×240×2 B | FM Towns 16-bit 直色美術。⚠ kb 陷阱:**FillOrder=2(LSB-first)** |
| `.SND` 25 個 | 音效。⚠ kb 陷阱:**sign-magnitude PCM**,不是 two's complement |
| `.16` 27 個、`DNG1–3.PNL` 各 54,560 B、`.TIL`/`.OOL`/`.GRA`/`.BIT` | 沿用 DOS 的 EGA 圖組 + FM Towns 專屬面板/圖 |

---

## 3. 技術堆疊(硬決策)

| 項目 | 決定 |
|---|---|
| 語言 / 引擎 | **Go 1.2x + Ebitengine v2**(2D、跨平台、內建 audio) |
| 邏輯畫布 | **640×400**(原版 320×200 的乾淨 2×);底圖 **nearest 整數放大**,不用線性 |
| 中文字形 | **倚天點陣字(ETEN)為預設**,見 §5 |
| 文字編碼 | 全程 **UTF-8**;原版 8×8 ASCII 路徑與 CJK 路徑分流 |
| 音訊 | Ebiten `audio/vorbis` 吃 ogg;XMI **離線**轉檔,不在遊戲內做 MIDI 合成 |
| 建置 | **[HARD] 一律 docker**(`docker/Dockerfile`:golang + CGO 需要的 X11/GL dev 套件);Python 工具走 docker uv/venv |
| 存檔位置 | `os.UserConfigDir()`,**絕不寫 cwd**(u1-cht 踩過:唯讀目錄下存不進去) |
| 授權 | 程式碼 **MIT**;原版資料 / 美術 / 音樂 / 轉出的 ogg **一律不入庫** |
| Repo | `github.com/wicanr2/u5-cht`,**每完成一個段落 commit + push**(`gh` 已裝) |

### 3.0 [HARD] 素材一律用原版(使用者 Goal,2026-08-07)

**遊戲素材全部從原版媒體抽取解碼,不自製、不重繪、不用 AI 生成、不用他人重製版。**

| 類別 | 唯一來源 |
|---|---|
| tileset / sprite / 場景圖 | 原版 `TILES.16`/`.4`、FM Towns `EGA*.TIL` 與 `.TIF`、PC-98 `TILES.CH`(§2) |
| 音樂 | FM Towns **CDDA 兩軌**、15 首 `.EUP`、upgrade 的 19 首 `.XMI`、PC-98 `UL*.BIN` |
| 音效 | 原版 `.SND`(FM Towns) |
| 文字 | 原版 `.TLK`/`.DAT`/`DATA.OVL` 抽出後翻譯,譯文走 i18n 覆蓋層 |
| 地圖 / NPC / 數值 | 原版 `.DAT`/`.NPC`/`.CBT`,不手調平衡 |

- **唯一的例外是中文字型** —— 原版沒有 CJK,必須外加;選倚天點陣字(§5.1)正是為了貼合同年代 DOS 中文觀感。
- 開發過程可以用佔位素材(placeholder)驗管線,但**不得進入交付**;缺素材時**優雅降級並明說**,
  不要拿自製圖濫竽充數(u1-cht 的做法:偵測不到該平台資產就跳過該主題)。
- 這條與 §3 的「引擎與資料分離」不衝突:**素材必須是原版,但不隨 repo 散布**,玩家自備合法副本。

### 3.1 Ebiten 的已知注意事項

- **CGO**:Linux/macOS build 需 CGO(Linux 要 `libgl1-mesa-dev`、`libxrandr/xinerama/xcursor/xi/xxf86vm-dev`);**Windows 目標可純 Go 交叉編譯**。macOS 需原生 runner(比照 u4-cht 走 GitHub Actions)。
- **[HARD] 畫面繪製不綁 GPU(2026-08-07 實測結論)**:`internal/render` 全部畫在 `image.NRGBA` 上,
  ebiten 只負責「把成品上傳成紋理 + 整數倍放大 + 收鍵盤」。
  ⚠ 這條是用五小時換來的:一開始把繪製綁在 ebiten,headless 截圖就得起 xvfb + 軟體 GL,
  結果**死鎖五小時、零輸出、最後只能砍容器**(u4-cht 也踩過「軟體 GL 死鎖」)。
  加逾時或換 GL 後端都只是治標 —— 根因是「驗證畫面」不該需要 GPU。
  改成單一 CPU 繪製路徑後:headless 是秒級純函式呼叫、CI 不需顯示環境、
  **截圖與實機畫面共用同一份 `render.Scene` 所以保證一致**。
  headless 驗收指令:`u5dump scene <gamedata> <U5_E> out.png`。
- 中文不用 `text/v2` 的 TTF 渲染:倚天 atlas 烘成一張 `*ebiten.Image`,取字用 `SubImage`。
- 邏輯畫布用固定 offscreen image → 最後一次 `DrawImage` 放大到視窗,縮放濾鏡設 nearest。

### 3.2 模組佈局(deep modules,介面窄、內部厚)

```
cmd/u5cht/            main + flag(含 -headless)
internal/u5data/      原版資料唯讀解碼:lzw / tiles / map / tlk / npc / ovl / font / cbt / save
internal/engine/      遊戲邏輯:world · party · time&moon · npc排程 · talk · combat · dungeon · magic · vehicle
internal/render/      Ebiten 繪圖:tile 層 · HUD · 文字層 · 顯示模式(EGA/CGA/Hercules)
internal/cjk/         倚天字庫→atlas · Big5↔UTF-8 · CJK 斷行與寬度
internal/i18n/        UTF-8 覆蓋層:key =(來源檔, 記錄索引)→ 譯文
internal/audio/       ogg 播放與場景配樂對應
tools/                ida.sh · dump_*.go · xmi2ogg.sh · build_eten_font.py · gen_func_index.py
docs/re/              逆向筆記(每份標:輸入檔 + SHA-256 + IDA 線性位址)
docs/formats/         資料格式規格(每份附可重跑的解碼腳本 + 驗證 oracle)
docs/manual/          手冊 OCR 與譯名對照(見 §7.2)
org_game/             原始素材(gitignore)
```

**[HARD] 引擎與資料分離**:`internal/u5data` 只讀、不寫回原版檔;中文譯文一律走 `internal/i18n` 覆蓋層,
**絕不把中文寫回 `.TLK`/`.DAT`**(會破壞 offset 表)。

---

## 4. IDA Pro 是第一工具(使用者明示要求)

> **老遊戲重製的行為問題,優先用 IDA Pro 靜態回答,而不是猜、也不是先跑模擬器。**
> 環境:image 來源 `/home/anr2/ida_94_official/dist`,image 名稱 **`ida-pro-9.4-ver2`**。

### 4.1 呼叫方式(包成 `tools/ida.sh`)

```bash
docker run --rm -v "$WORK:/work" -v "$ROOT/tools:/work/tools:ro" -w /work \
  ida-pro-9.4-ver2 idat -A -B ULTIMA.EXE      # 產 .i64 + .asm
```

`idat` 是 headless 那支(`ida` 是 GUI)。`-A` 自動、`-B` 批次。

### 4.2 逆向目標優先序:三版並用(使用者指示:從日文版著手)

**沒有單一最佳目標,三版各有所長,依「要回答什麼問題」選檔。**

| 要回答的問題 | 用哪一版 | 理由(全部已驗證,見 §2) |
|---|---|---|
| **遊戲邏輯**(戰鬥/魔法/NPC 排程/時間月相/業報) | ★ **FM Towns `WORRIORS.EXP`** | **32-bit 保護模式(P3),唯一可能反編譯成 C 的**;且資料檔與 DOS 版同大小 → 結論直接適用主線 |
| **中文化該 hook 哪裡**(DBCS 字型與排版) | ★ **`WORRIORJ.EXP` vs `WORRIORS.EXP` diff** | 同版同資料的日/英雙執行檔,差 24.7 KB —— 差的就是雙位元組文字路徑 |
| **overlay 切分與載入基址** | **DOS** | DOS 的 24 個 `.OVL` 是獨立檔案,邊界天然清楚;PC-98 是 176 KB 連續附加區,反而更難切。**FM Towns 是 flat 32-bit,沒有這個問題** |
| **資料格式**(tileset/地圖/TLK) | **DOS 當主線;FM Towns `EGA*.TIL` + PC-98 `TILES.CH` 當 oracle** | 玩家手上最普遍的是 DOS 版 → 引擎主線讀 DOS 格式;兩個日文版都提供未壓縮素材可對答案 |
| **文字語意**(翻譯時吃不準原意) | **DOS/`WORRIORS`(英)+ `U5_J/*.JPN`(日)並排** | 靠 TLK 的 index 欄逐筆對齊 |
| PC-98 的定位 | **降為輔助** | 資料是另一套格式(`.CH`/`.DAT`、GDC 4-plane),邏輯逆向已被 FM Towns 取代;仍保留 `TILES.CH` 當第二 oracle 與 YM2203 音樂來源 |

⚠ **已被一手證據推翻的假設,不要再重複**:
- ~~「日文版有 symbol table 比較好逆」~~ → **三版全數沒有 CodeView 符號**(PC-98 `U5.EXE`、FM Towns 兩個 `.EXP` 都掃過 NB05/08/09/10/11)。日文版的優勢是**反編譯器 + 雙語執行檔 + 未壓縮素材 + 英日對照文字**,不是符號。
- ~~「PC-98 單檔所以 IDA 一次吃完」~~ → MZ 檔頭只涵蓋 40,896 B / 216,880 B,**176 KB 在自動分析之外**。
- ~~「`upgrade/` 有 VGA 美術」~~ → 它是 MIDI 音樂升級包,`TILES.16` 與原版 md5 相同(§2.2)。

### 4.3 各版的共同關卡

1. **overlay 沒有 MZ 檔頭**:DOS 版 `.OVL` 開頭即 `55 8B EC`;PC-98 版是 EXE 尾巴的 176 KB 附加區。
   兩者都要以 **16-bit binary** 載入 IDA、手動建 segment、手動定基址。
   基址線索:先逆主 EXE 的 overlay 載入常式;DOS 版另有 upgrade 版 diff 可用(§2.2)。
2. **編譯器是 Microsoft C 5.x/6.0 → cdecl**(DOS 與 PC-98 兩版皆已驗證含 MS Run-Time 字串):
   參數右至左壓棧、**呼叫者**清棧。
   ⚠ IDA kb 裡的「Turbo Pascal 參數由左至右、`retn N`」是**別的專案的**,別套到 U5。
3. **16-bit real mode 沒有 Hex-Rays**(DOS 與 PC-98 兩版全程讀組語)。
   **FM Towns 的 `.EXP` 是例外,且已實測通過**(2026-08-07,見 `docs/re/00-hexrays-p3-verified.md`):
   - IDA 9.4 **原生認得**格式:`Phar Lap run386-extender flat model file`、`LVL_FLAT`、header 0x180、
     image offset 0x200、初始 `EIP=0x39700`。**不需手動剝 header**,直接 `tools/ida.sh analyze` 即可。
   - `WORRIORS.EXP` → **1,233 函式**、4.6 MB `.asm`(32-bit 暫存器參照 52,676 處 vs 16-bit 1,153 處)。
   - **Hex-Rays 批次反編譯成功**:`idat -Ohexrays:/work/out.c:ALL -A <db>.i64` → **61,364 行 C、1,225 函式**。
   ⇒ **FM Towns 這條路讀 C,不讀組語。** 這是 §4.2 把它列為主目標的最主要理由。

### 4.4 鐵則(全部來自 kb 的踩坑,違反就是重蹈)

- **寫 IDC,不要寫 IDAPython**(本機實測 IDAPython 無輸出)。
- IDC **必須** `#include <idc.idc>`,少了會**安靜 exit 1**、無任何訊息。
- headless 的 `print`/`Message()` **看不到** → 結果一律 `fopen("/work/out.txt","w")` 寫檔。
- **不要 grep `.asm` 找位址或用途**:16-bit 的 `.asm` 只有 `segment:offset`,線性位址一個都不會出現(零命中 ≡ 真的沒人碰,分不出來)。要問「誰讀寫這塊記憶體」→ 查 xref 圖(`get_first_dref_to`/`get_next_dref_to` + **`XrefType()`**,不要自己解析指令文字)。
- 間接寫入(`ptr = &x` → `es:[di]=v`)不在 xref 裡。看到「讀很多、寫只有 1 處」→ 去看「取址」那幾筆。
- IDC 崩掉會把 `.i64` 留在壞狀態,症狀是 `Failed to initialize IDA as library (error code 4)`,**看起來像 image 壞了**。判斷:拿另一個 `.i64` 跑已知可用腳本(正對照)。壞的那個刪掉重跑 analyze。
- **讀任何 `sub_XXXX` 前先查函式索引**(`tools/gen_func_index.py` → `docs/re/00-function-index.md`)。逆向筆記過三十份之後,憑記憶一定重讀。
- **「唯一」「只有一處」沒有全檔掃描佐證就不要寫。**
- `.i64` / `.asm` / 解包後 binary **全部 gitignore**。
- 只做靜態分析與互通性研究;license 唯讀掛載、不出現在 log/截圖;**不在 container 內跑遊戲**。

### 4.5 什麼時候才離開 IDA

靜態追到「該值來自某個間接跳表 / 執行期計算」而三次嘗試都無法收斂時,才依 `rulebook/64` 換路:
用**已破解的解碼器 + 已知輸出(DOSBox 實機截圖)反推**位置。**不要**一撞牆就跳去動態 dump。

---

## 5. 中文化設計

### 5.1 字形:倚天點陣字(預設)

- **字庫來源(使用者指定 2026-08-07):`/home/anr2/cht/etan_font/`**
  —— 目錄內有 `stdfont.15`(392,820 B = 13,094 字 ✓)與 `ascfont.15`,
  **但沒有 `spcfont.15`**;缺的字庫在同目錄的倚天 3.53 光碟映像 **`ET353S.iso`** 裡
  (`DISKS/DISK1/SPCFONT.15`、`SPCFSUPP.15`;24 點六種字體在 `DISKS/DISK4–6/STD.24M/K/L/R/B/S`)。
- **烘字工具:`tools/build_eten_font.py`(已可用)**,缺的字庫會自動用 `7z -r` 從 ISO 抽:
  ```bash
  docker run --rm --log-opt max-size=10m --log-opt max-file=3 \
    -v "$PWD":/work -v /home/anr2/cht/etan_font:/eten:ro -w /work u5cht/dev \
    python3 tools/build_eten_font.py --eten-dir /eten --iso /eten/ET353S.iso \
      --size 15 --out assets/fonts/eten-16x15      # 加 --verify-only 只跑 oracle
  ```
  產出 `.png` atlas + `.json`(`codepoints[i]` = atlas 第 i 格的 Unicode 碼位)。
  **烘出的 atlas 衍生自 1993 商業字庫 → 不入 git**,各自從自備字庫重跑。
- **[HARD] 一定要一起帶 `SPCFONT`**:`STDFONT.15` 從 A440「一」起,**不含 A140–A3BF 全形標點**;
  漏帶會出現「字是倚天、標點是別的字型」。工具在缺 `SPCFONT` 時**直接報錯不繼續**(不是警告)。
- Big5 分區索引**不是線性**(公式見 `eten-bitmap-font/SKILL.md` 與工具內的 `glyph_slot`)。
  **oracle 已於 2026-08-07 通過**,任何改動後要重跑:
  `std idx 0` = 「一」(單橫線)✓、「中」A4A4 → std idx 66 ✓、「猴」B555 → std idx 2690 ✓、
  **「。」A143 → spc idx 3**(這條同時驗證了分區公式與 SPCFONT 真的有載入)✓。
- 24 點路線的前置:`STD.24*` 是 **ETUNPACK 壓縮**,要先用現成解壓器
  (`~/scummvm/kq5/workplace/tools/etunpack.py`)解成裸字庫再餵給烘字工具;
  ⚠ `STD.24L`(隸書)那份在 ISO 上資料本身是壞的,實務用 `STD.24M`(明體)。
- 尺寸:640×400 畫布下 **預設 24×24**(與原版視覺同大但細節 2×)。
  ⚠ U5 的文字欄很窄(原版每行約 16 個 8px 字元 → 640 下該區約 256px),24×24 一行只放得下 ~10 個中文字。
  **若實測破版,退 16×15**(原味 DOS 點陣,一行約 16 字),並把量測結果與決定寫進 `docs/localization-notes.md`。
  **絕不把 16×15 非整數倍放大成 24**(點陣字必醜)。
- fallback 數量是品質指標:Big5 缺字才用 TTF 補同尺寸。一大批字掉 fallback → 先懷疑索引公式或漏帶 SPCFONT。

### 5.2 文字管線

- ★★ **不要自己發明 DBCS 排版,先把日文版的答案挖出來**。兩個現成先例,優先序如下:
  1. **`WORRIORJ.EXP` ⊖ `WORRIORS.EXP` diff(首選)**:同版本、同資料、同編譯器的日/英雙執行檔,
     差 24.7 KB。差異處就是「原作者為了塞進雙位元組文字,改了哪些繪字與換行程式碼」——
     這是中文化 hook 點的直接答案。而且是 32-bit,**可能可以反編譯成 C 讀**。
  2. **PC-98 版**:`FONT98.CH`(4096 B 半角)+ 全角走字型 ROM、`STORY.DAT` 是 Shift-JIS 明文、
     **無 `_` 斷字標記**。當第二對照。
  Big5 與 Shift-JIS 同為 2-byte 結構,結論可直接類推;**日文版沒有斷字標記**這點特別重要 ——
  中文同樣不斷字,所以日文版的換行邏輯比英文版更接近我們要的。
- **量測與繪製同一個 gate**:`CharWidth()` / `Draw()` / `LineHeight()` 由**同一個條件式**決定。
  三者只要有一處不一致,斷行算出來的就跟畫出來的不同 → 文字溢框,而症狀看起來像「譯文太長」。
- 排版格寬固定為邏輯字寬,字模在格內置中 → 換字模不破版。
- **[HARD] 會拿去跟玩家輸入比對的字串,canonical 值維持英文**(u4-cht 踩過的坑):
  U5 的 `Yes/No`、法術符文(`An Nox`…)、對話關鍵字(`name`/`job`/`join`…)玩家得打得出來。
  作法:`getXxx()` 回英文 canonical 供比對,`getDisplayXxx()` 回「中文(英文)」。
  U5 特別注意:**符文(rune)輸入與 `RUNES.CH` 顯示**是另一條路徑,要單獨盤。

### 5.3 譯名政策

- 權威來源優先序:**《軟體世界》U5 手冊(§2.3)> 台灣《創世紀聖者之書》體系(`../u4-cht/manuel/`、`../u4-cht/dumps/*_bilingual.json`)> 現代直觀**。
- **日文版當語意佐證,不當譯名來源**:`U5_J/*.JPN` 與 `U5_E/*.JPN` 解決的是「這句英文到底什麼意思」
  (尤其古英文腔的對白與雙關),**不是**「這個名詞中文怎麼翻」—— 日譯多用片假名音譯,直接搬過來會
  變成二手轉譯。名詞一律回到上面的優先序。
- 系列共通名(Lord British 不列顛王、Britannia 不列顛尼亞、八德、Avatar 聖者/化身、怪物、裝備)**一律與 u1-cht / u4-cht / u6-cht 對齊**,不另立新譯。
- 原則同 u4-cht:**正確性 / 直觀 > 純懷舊**;採用或不採用官方舊譯都把理由寫進 `docs/manual/術語對照.md`。
- 對白語氣:古英文倒譯(「汝」「卿」),對齊手冊的中世紀羊皮紙腔。
- 新名詞進 `CONTEXT.md` 的 glossary(`rulebook/50`)。

---

## 6. 階段規劃(每階段一段落 → commit + push)

> 詳細 backlog 展開到 `PLAN.md` / `WORKLIST.md`;本節只定順序與驗收門檻。

| 階段 | 內容 | 驗收(決定性,不是「跑得動」) |
|---|---|---|
| **P0** | 本檔潤飾成 `PLAN.md` + `CONTEXT.md` + 鳥巢框架(目錄/docker/CI 骨架) | 框架就位,`docker build` 通過 |
| **P1** | 資料解碼 CLI。**順序刻意這樣排**:① FM Towns **未壓縮** `EGA0–3.TIL` → PNG(最快看到正確 tile);② 以它當 oracle 破 DOS `TILES.16` 壓縮(檔頭宣稱解壓後 65,536,恰等於單個 `.TIL`);③ 字型 `IBM.CH`(已知 8×8 ASCII 直索引)/ `U5.FNT`(佈局未知,要驗)→ PNG;④ `BRIT.DAT`/`UNDER.DAT` → 地圖圖;⑤ `.TLK` 三版解碼 → JSON(DOS 清 bit7 / `U5_J/*.JPN` 走 Shift-JIS / PC-98 offset 陣列),**用 index 欄產出英日對照表** | **tileset PNG 與原版截圖逐格對得上**;`TILES.16` 解壓結果與 `.TIL` **逐位元組相同**;`.TLK` 全 4 檔解出可讀對話、記錄數對得上 `.NPC`;英↔日逐筆對齊 |
| **P1.5** | headless 截圖 loop(§3.1)+ 倚天字庫烘製 + Big5 索引 oracle | 容器內截出「一句中文」的 PNG;oracle 三字驗證通過 |
| **P2** | Ebiten 垂直切片:世界地圖可走 + 狀態列 + 中文訊息欄 | 截圖比對基準;鍵盤移動、tile 正確 |
| **P3** | IDA 逆向 oracle 主攻(依 §4.2 分派)。**第一件事:實測 Hex-Rays 能否反編譯 `WORRIORS.EXP`(P3 格式)** —— 結果決定後續是讀 C 還是讀組語。接著:`WORRIORJ ⊖ WORRIORS` diff 定位 DBCS 文字路徑、`.TLK` 索引與控制碼語意、時間與月相、NPC 排程、戰鬥與魔法公式;DOS 的 overlay 載入機制另案 | 每條結論在 `docs/re/` 有筆記(輸入檔 + SHA-256 + IDA 位址),且**與 DOSBox 實測一致** |
| **P4** | 引擎補完:城鎮/城堡/地底世界/地牢/戰鬥/NPC 對話/魔法/船與氣球/存檔 | **正常玩家路徑**可從開場走到主要城鎮(見下方鐵則) |
| **P5** | 全文中文化:`.TLK` ×4 + 7 個明文 `.DAT` + `DATA.OVL` 名表 + `.OVL` 硬編字串 | 逐畫面巡查,玩家可見文字 0 殘留英文(美術內嵌字母另計並誠實揭露) |
| **P6** | 音樂與美術主題。音樂三來源:**FM Towns 兩條 CDDA**(直接轉 ogg,最省)、**15 首 `.EUP`**(EUPHONY,u1-cht 逆過可複用)、**upgrade 的 19 首 `.XMI`**(→MID→fluidsynth);場景對應直接讀明文 `U5_BGM.TBL`/`U5_SE.TBL`。美術主題:EGA / CGA / Hercules / **FM Towns 直色 `.TIF`** / PC-98 | 熱鍵切換不掉 NPC、不崩(u1-cht 踩過);顏色對得上各版實機截圖(⚠ TIF FillOrder=2、`.SND` 是 sign-magnitude PCM);音樂隨場景切換 |
| **P7** | 打包 Linux / Windows / macOS + CI + game tester 回歸 | 解壓即玩;`retro-game-playtest` 驗收通過 |
| **P8** | README(§7)+ 攻略 / 文件收尾 | §7 的引用要求全部滿足 |
| **後補** | PC-98 的 YM2203 曲譜(`UL01–15.BIN` + `OPNDRV.COM`)還原與 PC-98 美術主題;其他平台(Apple II / C64 / Amiga / Atari ST,需自備媒體) | 顏色與聲音對得上實機;每種格式的破解過程寫成 `docs/re/` 可重跑筆記 |

### 6.1 驗收鐵則(這些是用時間換來的)

- **[HARD] debug hook 會遮住真 bug**:發道具 / 瞬移 / 強制進城的回歸測試會**全綠但玩家一開就壞**
  (u2-cht 真實案例:新角色被放在只連城堡的 12 格小島 soft-lock)。**一定要另驗「無 debug 的正常玩家路徑」**;
  世界可達性用 **flood-fill 連通分量**檢查(落點在最大陸地分量、城鎮與落點同分量、船在鄰接玩家分量的水格)。
- **[HARD] 測試綠 ≠ 完成**(`rulebook/65`):驗收標準是**對 reference(DOSBox 跑的原版)實測比對**,
  不是自家測試通過。宣稱任何階段完成前,先拿原版並排。
- 打包要帶**全部**資料,不要只帶子集;**不要**打包測試角色存檔(玩家該走建角流程);
  headless 預設不寫回存檔(會覆蓋餵進去的 fixture)。
- 連續兩輪同類失敗 → 停,載 `rulebook/40`,說明換方法的理由。固定參數重試 N 次不是實驗。

---

## 7. README.md 的要求(使用者明確指定)

### 7.1 結構與風格

依 `rulebook/80` 的**三層 voice**(不可混用):Hero 開場信(第一人稱、溫情)→ Magazine 主體
(1990s 三大誌編輯人聲,Ultima 系用奇幻史詩體 + 「汝/卿」,**不要**用 X-COM 的「指揮官」)→
Technical Deep Dive(冷靜、可重現)。每張表格後有 70–150 字 prose 收束;章節間有橋接句;
TOC anchor 與章節 1:1 對齊。寫完過一遍 `rulebook/91`(去 AI 味)。

### 7.2 **[HARD] 必須引用遊戲手冊**

手冊(`org_game/manuel/`,《軟體世界》說明書補完計劃 057)是本專案的中文權威來源,README 要:

1. **獨立章節**介紹手冊本身:是什麼、哪一期、當年台灣玩家怎麼靠三大誌手冊翻譯玩英文遊戲。
2. **譯名政策章節明確標示**哪些詞採手冊譯法、哪些不採、理由(考古感而非批判感,見 `rulebook/80` 準則 5),
   並連到 `docs/manual/術語對照.md`。
3. **出處與致謝**:《軟體世界》雜誌與說明書補完計劃的譯者/整理者掛名致敬。
4. 附 1–2 張手冊頁面圖當視覺錨點(排版 / 世界觀圖),不是整本貼上。
5. 背景 / 操作 / 數值敘述以手冊為準時**標明出自手冊**,與逆向得到的數值分開陳述(兩者不一致就是一條 finding)。

### 7.3 手冊的版權處理(使用者已決定,附風險註記)

- **使用者 2026-08-06 明示決定:掃描 JPG + 全文 OCR 都入庫**,置於 `docs/manual/`。
- ⚠ **殘留風險(已向使用者說明,由使用者承擔決策)**:1990 年代雜誌譯稿著作權未過期,
  u4-cht 對《聖者之書》採的是「不入庫、不逐字謄錄」。本專案採較寬做法,故**必須**在 README 與
  `docs/manual/README.md` 標明:來源、僅供研究與譯名對照、非商業、**權利人要求即撤除**。
- 原版遊戲資料、PC-98 `.fdi`、XMI 轉出的 ogg、IDA 產物**一律不入庫**(與手冊處理不同,不要混淆)。

---

## 8. 執行邊界(agent 一律遵守)

**可以自動做**:`~/anr2` 內讀寫、docker build/test、IDA 靜態分析、寫文件與工具、commit + push 本 repo。

**先回報再動**:改對外 API/交付定案、大型跨模組重構、動別的 repo。

**[HARD] 一律不做**

- 碰共用 docker 資源:**禁止** `docker image prune`(含 `-a`)/`system prune`/`volume prune`/
  `builder prune`/`rmi`/`container prune`/`rm` 別人的 container。這台機器同時放多個客戶專案的 image,
  誤刪過一次事故。要空間 → 只列候選清單給使用者決定。
- 手打 `docker run` 少了 `--rm --log-opt max-size=10m --log-opt max-file=3`(370 GB log 事故)。
- 系統 Python `pip install`;非 docker 的編譯。
- force push、跳過驗證、把未驗證產出當定稿。
- **派 subagent 時**:邊界要寫進 prompt(只清自己建的 `u5cht-*` container、禁止任何 prune/rmi、
  不准改哪些目錄、不准自行 commit/push/清理)。**沒寫的等於允許**;agent 回報「我順便做了 X」→ X 當事故處理。

---

## 9. 語言與回應

- 繁體中文(README、PLAN、CONTEXT、commit message、註解、docstring);程式碼識別字保留英文。
- 中性客觀,避免浮誇詞;結論、風險、TODO 明確。
- commit message 結尾掛 `Co-Authored-By: Claude Opus 5 (1M context) <noreply@anthropic.com>`。
- 遊戲測試在背景跑,用 game tester 驗證,不阻塞對話。

---

## 10. 第一步

**先把本檔潤飾展開成 `PLAN.md`(可執行工程計畫)+ `CONTEXT.md`(語彙/背景/硬規則),
建好 §3.2 的鳥巢框架與 `docker/`,再回來與使用者對話確認,然後才進 P1。**
