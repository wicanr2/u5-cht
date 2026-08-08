# Ultima V 重製 + 繁體中文化 — 執行計畫

> 由 `CLAUDE.md`(專案憲章)展開為可執行工程計畫。語彙與背景見 `CONTEXT.md`。
> 立案 2026-08-06 ・ 最後更新 2026-08-08 ・ 維護:L.CY (anr2) + Claude

---

## 0. TL;DR

用 **Go + Ebitengine 從零重寫** 1988 DOS 版《Ultima V: Warriors of Destiny》,**完整繁體中文化**。
與 u1-cht / u4-cht 最大的不同:**沒有可用的開源上游,遊戲邏輯全部自己寫,逆向是主線**。

三件事撐起整個計畫:

1. **行為真值** —— IDA Pro 反組譯。FM Towns 版是 32-bit Phar Lap `P3`,**Hex-Rays 反編譯成功
   (61,364 行 C / 1,225 函式)**,所以讀 C 不讀組語。
2. **素材** —— 一律用原版(`CLAUDE.md §3.0`)。三個版本互為 oracle:DOS(主線)、
   FM Towns(未壓縮 tileset + 英日雙執行檔 + 英日對照文字)、PC-98(第二 oracle)。
3. **中文** —— 640×400 邏輯畫布 + **倚天點陣字**(`/home/anr2/cht/etan_font` + `ET353S.iso`),
   譯文走 i18n 覆蓋層,不寫回原版檔。

---

## 1. 成功標準(可驗證,不是「跑得動」)

1. 全新角色能從開場走完主要流程:創角 → 地表 → 進城鎮 → 與 NPC 對話 → 進地牢 → 戰鬥 → 存讀檔。
2. 玩家可見文字全繁中(美術 tile 內嵌字母另計並誠實揭露);字不溢框、不破版。
3. 三平台可執行檔(Linux / Windows / macOS),解壓即玩,存檔寫使用者設定目錄。
4. **對照原版實測**:同一情境下的行為與數值與 DOSBox 跑的原版一致(`rulebook/65`)。
5. **無 debug 的正常玩家路徑**可達性通過(flood-fill 連通分量檢查)。
6. README 依 `rulebook/80` 三層 voice 完稿,並依 `CLAUDE.md §7.2` 引用《軟體世界》手冊。

## 2. 範圍外(初期)

- 重繪美術、加新內容、調整平衡(素材與數值一律忠於原版)。
- Android / iOS / WASM(Ebiten 支援,但排在三平台之後)。
- 原版資料散布(引擎與資料分離,玩家自備)。

---

## 3. Repo 佈局(★ = 已存在)

```
u5-cht/
├── CLAUDE.md ★  PLAN.md ★  CONTEXT.md ★  README.md         ← 憲章 / 計畫 / 語彙 / 玩家向
├── cmd/u5cht/ ★                                            ← 執行檔(P0:640×400 + 原版字型)
├── internal/
│   ├── u5data/ ★        ← 原版資料唯讀解碼(font ★ / tlk ★ / tiles / map / npc / save)
│   ├── engine/          ← 遊戲邏輯(world · party · time&moon · npc · talk · combat · dungeon · magic)
│   ├── render/          ← Ebiten 繪圖(tile 層 · HUD · 文字層 · 顯示模式)
│   ├── cjk/             ← 倚天 atlas 載入 · Big5↔UTF-8 · CJK 斷行與寬度
│   ├── i18n/            ← UTF-8 覆蓋層:key =(來源檔, 記錄索引)→ 譯文
│   ├── audio/           ← ogg 播放與場景配樂對應
│   └── save/            ← 存檔(os.UserConfigDir,不寫 cwd)
├── tools/ ★             ← ida.sh ★ · fdi_extract.py ★ · cdimg_to_iso.py ★ · build_eten_font.py ★
├── docker/Dockerfile ★  ← Go 1.24 + CGO/GL/X11 + xvfb 軟體 GL + PIL + 7z
├── docs/
│   ├── re/ ★            ← 逆向筆記(00-hexrays-p3-verified.md ★)
│   ├── formats/         ← 資料格式規格(每份附可重跑腳本 + 驗證 oracle)
│   └── manual/          ← 《軟體世界》手冊 OCR 與術語對照(§7.2/§7.3)
├── assets/fonts/        ← 烘出的倚天 atlas(gitignore,各自重跑)
├── gamedata/            ← 玩家自備的原版資料(gitignore)
├── org_game/ ★          ← 原始素材(gitignore)
└── re_work/ ★           ← IDA 工作區與產物(gitignore)
```

---

## 4. 階段 backlog

### P0 — 框架 ✅ 已完成(2026-08-07,commit 51b754e)

- [x] `CLAUDE.md` 專案憲章(含 IDA 優先、素材鐵則、README 手冊引用要求)
- [x] 素材三版一手驗證,結論寫進 `CLAUDE.md §2`
- [x] `tools/fdi_extract.py`(PC-98 FAT12/1024 B sector)、`tools/cdimg_to_iso.py`(MODE1/2352 → ISO)
- [x] `tools/ida.sh` + **Hex-Rays 對 P3 實測通過**(`docs/re/00-hexrays-p3-verified.md`)
- [x] `tools/build_eten_font.py` + **倚天 Big5 分區索引 oracle 通過**
- [x] `docker/Dockerfile`、`go.mod`、`.gitignore`、`CONTEXT.md`、`PLAN.md`
- [x] `internal/u5data`:`font.go`(IBM.CH 8×8)、`tlk.go`(對話檔)+ 單元/整合測試
- [x] `cmd/u5cht`:640×400 邏輯畫布 + nearest 整數放大 + 原版字型字元表 + F10/Ctrl+Q 離開
- [x] Docker 內 `go build` / `go vet` / `go test` **全綠**;整合測試對真素材通過
      (`'A'` 逐位元組相符、`TOWNE.TLK` 48 筆首筆 Zachariah、`TOWNE.JPN` 同 48 筆且 NPCIndex 逐筆一致)
- [x] `tools/dev.sh` 開發包裝(module cache 持久化)、`LICENSE`(MIT)、GitHub repo 建立並推送

### P1 — 資料解碼(順序刻意這樣排)

1. ✅ **`EGA0–3.TIL` → 512 tile**(格式完全破解 + 色號 IGRB 校正)
2. 🔶 **破 `TILES.16` 壓縮** —— 確認不是標準 LZW;已定位載入鏈但解壓函式未找到,**移交 P3**
   (引擎先用 FM Towns 未壓縮版,不擋進度)
3. 字型:`IBM.CH` ✅;`U5.FNT`(FM Towns「ULTIMA FONT」16,384 B)、`TOWNS.FNT`(「IBM FONT」
   17,280 B)、`FONT98.CH`(PC-98)佈局待驗
4. ✅ **`BRIT.DAT` → 完整 256×256 世界地圖**(205 chunk + `DATA.OVL` 0x3886 索引表);
   `UNDER.DAT` 與場景地圖(`TOWNE`/`CASTLE`/`KEEP`/`DWELLING`)待做
5. `.TLK` 三版 → JSON,**用 NPCIndex 產英日對照表**(`tlk.go` 已可解,欄位切分待 P3)
6. ✅ 明文 `.DAT` → 記錄(**NUL 分隔**,不是 `|`;`_` 斷字提示已處理)。
   `STORY`/`QUESTION`/`KARMA`/`MISCMSG`/`ENDMSG` 共 114 筆可直接翻;
   `SHOPPE.DAT` 194 筆含 862 個詞典 token,**待 token 映射定案**(字典已定位在 `DATA.OVL` 0x104C);
   `LOOK2`/`SIGNS` 格式不同,另案
7. `DATA.OVL` 物品/武器/防具/法術名表

**驗收**:tileset PNG 與原版截圖逐格對得上;`TILES.16` 解壓結果與 `.TIL` 逐位元組相同;
`.TLK` 記錄數對得上 `.NPC`;英日逐筆對齊。

### P1.5 — 中文顯示管線

1. ✅ 倚天 atlas 烘製(16×15,13,461 字)→ `internal/cjk` 載入(png/json 同步檢查)
2. ✅ 文字層 `internal/render`:ASCII 走原版 8×8、CJK 走倚天 16×15,缺字畫紅框(不靜默跳過)
3. ✅ **量測與繪製同一個 gate**:度量與斷行抽成 `internal/textlayout`(純邏輯、不依賴 ebiten),
   `render` 只 re-export 不自己另算。property test 固定住「Wrap 出來的每一行寬度都 ≤ 上限」
4. ✅ **headless 截圖**:`u5dump scene`(純 CPU,秒級,不需 GPU)。
   ⚠ 原本走 ebiten + xvfb + 軟體 GL 的版本**死鎖五小時**,已改成 CPU 繪製路徑(見 CLAUDE.md §3.1)
5. ⬜ 24×24 版本與尺寸決策實測 → `docs/localization-notes.md`

**驗收**:容器內截出中英混排畫面,字清晰、不溢框(已達成);24×24 選項待做。

### P2 — 垂直切片(可走的 Britannia)

1. ✅ 11×11 tile 地圖視窗(tile 2× nearest 放大成 32 px,佔畫面比例與原版一致)
2. ✅ 方向鍵移動 + 世界環繞(wrap-around)+ 落點自動找陸地(避免開場泡在海裡)
3. ✅ 中文狀態欄與訊息欄(全繁中,走倚天點陣字)
4. ⬜ 碰撞與地形效果(屬遊戲邏輯,P4)
5. ⬜ Avatar tile(索引待 P3 從反編譯碼確認,目前用白框標記中心)
6. ⬜ 時間推進與日夜

**驗收**:headless 截圖為基準;鍵盤移動正確。

### P3 — 逆向 oracle 主攻

1. **`WORRIORJ.EXP` 反編譯 → 與 `WORRIORS.EXP` diff**,定位 DBCS 繪字/換行/字型載入
2. `sub_2C740` / `byte_54700` xref → `.TLK` 索引語意與控制碼(`\x01` 疑為玩家名代入)
3. 時間與月相、NPC 排程(`.NPC` + `.PTH`)、戰鬥與魔法公式、業報規則
4. DOS overlay 載入機制與基址(另案;`upgrade/` 的 8 個 patch 檔可當 diff oracle)
5. 建 `docs/re/00-function-index.md`,每命名一個 `sub_XXXX` 就登記(避免重讀)
6. 確認編譯器身分(Hex-Rays 自報 GNU C++ 是誤判)

### P4 — 引擎補完

城鎮/城堡/地底世界/地牢(第一人稱)/戰鬥/NPC 對話/魔法/船與氣球/存檔。
**驗收**:正常玩家路徑可從開場走到主要城鎮;flood-fill 可達性通過。

### P5 — 全文中文化

八類文字全譯(`.TLK` ×4、七個明文 `.DAT`、`DATA.OVL` 名表、`.OVL`/`.EXP` 硬編字串);
譯名依 `CLAUDE.md §5.3` 優先序,決定與理由寫 `docs/manual/術語對照.md`。
**驗收**:逐畫面巡查 0 殘留英文(美術內嵌字母另計)。

### P6 — 音樂與美術主題

音樂三來源:FM Towns **CDDA 兩軌**(最省)、15 首 `.EUP`、upgrade 19 首 `.XMI`;
場景對應**要逆向**(`docs/re/87`;`.TBL` 只給曲號→檔名與六聲道音量)。美術主題:EGA / CGA / Hercules / FM Towns `.TIF` / PC-98。

### P7 — 打包與回歸

Linux / Windows / macOS(Go 交叉編譯 + GitHub Actions;Windows 目標可純 Go)、
game tester 正常玩家路徑驗收、存檔路徑實測。

### P8 — README 與文件

README 三層 voice + 手冊引用章節 + 譯名政策 + 逆向手記;`docs/manual/` OCR 與術語對照。

---

## 5. 驗證策略

⚠⚠ **實際跑過的只有上面三層**(2026-08-08 盤點)。下面兩層 —— 也就是唯一能
證明「機制跟原版一模一樣」的兩層 —— **一次都沒跑過**。這是目前最大的缺口,
`WORKLIST.md §4.1`(A 階段)就是為了補它。

| 層次 | 手段 | 跑過了嗎 |
|---|---|---|
| 單元 | `go test ./...`(無素材也全綠:合成資料 + 格式假設的自我一致性檢查) | ✅ 每次 commit |
| 整合 | 設 `U5_GAMEDATA` / `U5_FMTOWNS` 後跑同一批測試,對原版素材驗證 | ✅ 每次 commit |
| 格式 | 每個解碼器都要有 oracle(如倚天的「一/中/猴/。」、`TILES.16` 對 `.TIL`) | ✅ |
| 視覺 | headless 截圖存 `tests/snapshots/`,逐畫面比對基準 | 🔶 截得出來,但**沒有原版基準可比** |
| **行為** | 對 **DOSBox 原版**並排實測(`rulebook/65`:測試綠不算完成) | ⬜ **從未執行** |
| **可玩性** | game tester 跑正常玩家路徑(`retro-game-playtest`),無 debug hook | ⬜ **從未執行** |

---

## 6. 風險與對策

| 風險 | 對策 |
|---|---|
| 🟠 **引擎規模**:全遊戲邏輯自己寫,是三個姊妹專案中最大的 | 垂直切片優先(P2 先能走能看),逐系統補;deep modules 讓各系統可獨立驗 |
| ~~🟠 **`TILES.16` 壓縮未破**~~ | ✅ **已解**:是 LZW(`internal/u5data/lzw.go`)。驗收是決定性的 —— 解出來的 65,536 B 與 FM Towns 四檔還原後**逐位元組相同**(`TestDOSTilesMatchFMTowns`)。⚠ 當初判「不是標準 LZW」是因為 oracle 少了一層色號轉換,**驗收條件有洞會讓對的答案看起來是錯的** |
| ~~🟡 headless 截圖(Ebiten 要 GL context)~~ | ✅ **已解**:繪製改成純 CPU(`internal/render` 不依賴 ebiten),headless 秒級且與實機畫面共用同一份 Scene。踩坑紀錄見 `docs/engineering-notes.md` |
| 🟡 **24×24 在窄文字欄破版** | 先實測再決定,退路是 16×15;決定與量測寫進文件 |
| 🟡 **玩家輸入比對**(Yes/No、符文、對話關鍵字)中文化後打不出來 | canonical 值維持英文,顯示「中文(英文)」(u4-cht 踩過的坑) |
| 🟡 **手冊版權**(掃描 + OCR 入庫是使用者決定) | README 與 `docs/manual/README.md` 標明來源、僅供研究對照、權利人要求即撤 |
| 🟡 **IDA 產物體積**(`.i64` 5.7 MB、反編譯 C 61K 行) | 全部 gitignore;結論寫成 `docs/re/` 筆記,附輸入檔 SHA-256 與位址 |

---

## 7. 下一步

1. 進 P1 第 1–2 項:`EGA0–3.TIL` → PNG,再以它當 oracle 破 `TILES.16` 壓縮。
3. 並行:反編譯 `WORRIORJ.EXP` 並與 `WORRIORS.EXP` diff(P3 第 1 項,結果會餵回 P1.5 的字型管線設計)。
