# CONTEXT — Ultima V 重製 + 中文化:語彙與背景

> 進本 repo 工作前先讀。術語表隨新名詞維護(`rulebook/50-ubiquitous-language`)。
> 規則與硬約束在 `CLAUDE.md`;素材事實在 `CLAUDE.md §2`(已一手驗證,勿憑記憶推翻)。

## 專案一句話

把 1988 年 DOS 版《Ultima V: Warriors of Destiny》**用 Go + Ebitengine 從零重寫**成跨平台可玩引擎,
並**完整繁體中文化**;行為真值來自 IDA Pro 反組譯(FM Towns 32-bit 版可反編譯成 C),素材一律用原版。

## 與範本專案的關係

| 範本 | 對本專案的價值 |
|---|---|
| `../u4-cht` | 譯名體系(聖者之書 + 精訊手冊 OCR)、`dumps/*_bilingual.json`、load-time 查表架構、多平台素材與 FM 音樂逆向經驗、macOS/Windows 打包 CI |
| `../u1-cht` | 多平台 tileset 熱切換、EUPHONY/MCP 音樂格式逆向、打包路徑陷阱(cwd/存檔位置)、README 三層 voice 範例 |
| `../u3-cht` | Ultima 系列資料格式家族知識、mcmagi(Exodus Project)反組譯文件的用法 |
| `../u6-cht` | 系列共通譯名(八德/Avatar/Lord British)的既定譯法 |

**本專案與三者最大的不同**:沒有可用的開源上游引擎,**遊戲邏輯全部要自己寫**,逆向是主線而非支線。

## 素材現況(摘要;完整已驗證事實見 `CLAUDE.md §2`)

| 版本 | 位置 | 定位 |
|---|---|---|
| DOS 1988 | `org_game/Ultima_V_-_...1988.zip` | **資料格式主線**(玩家最普遍)、overlay 機制來源 |
| DOS + Exodus MIDI Upgrade(2001) | 同上 `upgrade/` | 19 首 `.XMI` 音樂;8 個被 patch 的檔可當 overlay 結構 oracle。**沒有 VGA 美術** |
| **FM Towns 日文版(1992)** | `org_game/fmtown/*.7z` | ★ **逆向主目標**(P3 32-bit,Hex-Rays 可用);英日雙執行檔;未壓縮 tileset;英日對照文字;CDDA + EUP 音樂 |
| PC-98 日文版 | `org_game/【PC98】*.rar` | 輔助:第二個未壓縮 tileset oracle、YM2203 音樂 |
| 《軟體世界》U5 手冊 | `org_game/manuel/*.rar`(61 張 JPG) | 中文譯名與背景的**權威來源**;README 必須引用(`CLAUDE.md §7.2`) |

## 術語表(glossary)

### 已定案(沿用 u4-cht / u6-cht / 聖者之書體系,不另立新譯)

| 英文 | 中文 | 備註 |
|---|---|---|
| Britannia | 不列顛尼亞 | 世界名 |
| Lord British | 不列顛王 | 系列共通;U5 開場即失蹤 |
| Avatar | 聖者 | 對齊聖者之書;玩家角色 |
| Virtue / the Eight Virtues | 美德 / 八德 | U4 建立的體系 |
| Shrine | 聖壇 | |
| Mantra | 真言 | |
| Moongate | 月門 | |
| Rune | 符文 | U5 有專用符文字型(`RUNES.CH`) |
| Codex of Ultimate Wisdom | 終極智慧法典 | |
| Karma | 業報 | 原版有 `KARMA.DAT` |
| Moonglow / Britain / Jhelom / Yew / Minoc / Trinsic / Skara Brae / New Magincia | 月光城 / 不列顛城 / 哲倫 / 紫衫城 / 米諾克 / 特林希克 / 史卡拉布雷 / 新馬精西亞 | 八德城市,對齊 u4-cht |
| Dagger / Mace / Plate / Chain | 短劍 / 釘頭鎚 / 鎧甲 / 鎖子甲 | 對齊精訊手冊譯法 |

### 待查證(手冊 OCR 完成後定案,**在那之前不要寫進程式或文件當定譯**)

| 英文 | 暫譯 | 為什麼待確認 |
|---|---|---|
| Lord Blackthorn | — | U5 的暴君,系列前作未出現;台灣譯名需查《軟體世界》手冊 |
| Shadowlords(Astaroth / Faulinei / Nosfentor) | — | U5 專有三反派;音譯需對齊當年譯法 |
| Underworld | 地底世界 | 原版 `UNDER.DAT`;暫譯待手冊確認 |
| Words of Power | — | 地牢咒語 |
| Oppression / the Great Council | — | U5 的政治主題用詞,語感要對 |

> 譯名決定與理由一律寫進 `docs/manual/術語對照.md`(格式對齊 u4-cht),**不在 commit message 裡定案**。

## 技術關鍵事實(供快速定位)

- **邏輯畫布 640×400**(原版 320×200 的乾淨 2×);底圖 nearest 整數放大;中文預設倚天 24×24
  (U5 文字欄窄,實測破版則退 16×15,理由寫 `docs/localization-notes.md`)。
- **原版字型 `IBM.CH` = 128 glyph × 8 B、8×8、索引即 ASCII 碼**(已 dump 驗證)→ 沒有 CJK 空間,
  中文必須另開點陣路徑。
- **對話 `.TLK`**:DOS/英文版每 byte **bit7 被設為 1**(清掉即明文);FM Towns 日文版 `.JPN` 是
  Shift-JIS 原樣;兩者檔頭都是 `(u16 offset, u16 index)` 交錯索引表,**可靠 index 逐筆英日對齊**。
- **明文文字檔**(`STORY` `QUESTION` `KARMA` `MISCMSG` `ENDMSG` `SHOPPE` `LOOK2`.DAT):
  `|` 分隔記錄、`{` 起始、**`_` 是斷字提示**(譯成中文要移除)。
- **`TILES.16` 檔頭前 4 byte 是解壓後長度**(65,536),壓縮演算法待破;
  **FM Towns `EGA0–3.TIL` 各 65,536 B 未壓縮**,是破解時的對答案 oracle。
- **逆向入口**:FM Towns `WORRIORS.EXP` → `tools/ida.sh analyze` → Hex-Rays 出 61K 行 C;
  檔案讀取常式錨點 `sub_2C740(?, 檔名, 緩衝區, 長度, ?)`、對話緩衝 `byte_54700`
  (見 `docs/re/00-hexrays-p3-verified.md`)。

## 硬規則(完整版見 `CLAUDE.md`)

- 編譯一律 docker;Python 走 docker uv/venv,不污染系統。
- **素材一律用原版**(`CLAUDE.md §3.0`),唯一例外是中文字型。
- 引擎與資料分離:原版資料、IDA 產物、抽出的素材**全部不入庫**;手冊 OCR 例外(使用者明示)。
- 存檔寫 `os.UserConfigDir()`,不寫 cwd。
- 驗收對 reference(DOSBox 原版)實測,不靠自家測試綠(`rulebook/65`);另驗**無 debug 的正常玩家路徑**。
- 每完成一段落 commit + push `github.com/wicanr2/u5-cht`;回應與文件用繁體中文。
