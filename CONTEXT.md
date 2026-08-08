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
| DOS + Exodus MIDI Upgrade(2001) | 同上 `upgrade/` | 15 首 `.XMI` 音樂;8 個被 patch 的檔可當 overlay 結構 oracle。**沒有 VGA 美術** |
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
| Grapple | 抓鉤 | ⚠ **不是繩索**;大地圖攀爬與「從洞爬回上一層」的前提 |
| Ztats | 角色數值畫面 | 原版共 **17 頁**(六名 × 2 + 裝備 + 藥草 / 咒語 / 道具 / 軍械)|
| Equipment(Ztats 第 12 頁) | 裝備頁 | 全隊共用的糧食 / 金幣 / 鑰匙 / 寶石 / 火把 / 抓鉤 |
| Armaments(Ztats 第 16 頁) | 軍械 | 48 件裝備的持有量,與「裝備頁」是不同的兩頁 |

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
- **明文文字檔**:記錄用 **NUL(0x00)分隔**;`{` 是段落起始、
  **`_` 是英文斷字提示**(譯成中文要移除)。
  `STORY`/`QUESTION`/`KARMA`/`MISCMSG`/`ENDMSG` 共 114 筆、**0 個 token,可直接翻**;
  `SHOPPE.DAT` 194 筆但有 **862 個詞典 token**(字典在 `DATA.OVL` 0x104C,118 個常用詞),
  **映射未定前不得翻譯**;`LOOK2`/`SIGNS` 格式各自不同,另案。
- **tileset 已破解**:FM Towns `EGA0–3.TIL` = 128 tile/檔 × 512 B,是原版 16×16 機械放大 2× 的
  32×32 4bpp packed;還原後 512 tile × 128 B = 65,536 B,等於 DOS `TILES.16` 宣稱的解壓後長度。
  ⚠ **FM Towns 那份的色號用 IGRB 不是 EGA 的 IRGB**(bit1↔bit2 對調);換色號的地方在
  **解碼器**(`u5data.tileColorRemap`),所以算繪端直接用 `u5data.EGAPalette` 就對。
- **DOS `TILES.16` 的壓縮已破**:是 LZW(`internal/u5data/lzw.go`)。驗收是決定性的:
  解出來的 65,536 B 與「FM Towns 四檔降回 16×16、換色號、壓回 4bpp」**逐位元組相同**
  (`TestDOSTilesMatchFMTowns`)。
- **世界地圖已組出**:`BRIT.DAT` = 205 個 16×16 chunk;**chunk 索引表在 `DATA.OVL` 0x3886**
  (256 B,`0xFF` = 該位置全水,51 + 205 = 256)。
- **音效觸發是兩層的**:`sub_2C46C(索引, 音高)` 直接放 PCM;而**多數呼叫點走
  `sub_2C598(Rate, Dura, Limit)`** —— 那是「DOS 版白噪參數 → FM Towns PCM」的**轉接層**,
  八組固定三元組各換成一個 `.SND`,**查不到的組合在 FM Towns 上完全無聲**
  (26 個呼叫點裡 7 個是這種,那就是原版行為,不要補)。索引 = `U5_SE.TBL` 的行號。
- **環境音是每次重繪掃出來的**:`sub_2BDE0` 掃隊伍周圍 11×11 取**距離平方最小**的發聲物 ——
  落地鐘 0xFA / 瀑布 0xD4 / 噴泉 0xD8 / 疊圖層的樂器 0x5C。落地鐘**十二小時制報時**,
  樂器會把配樂**壓下去**且要走到完全安靜才接回來。
- **逆向入口**:FM Towns `WORRIORS.EXP` → `tools/ida.sh analyze` → Hex-Rays 出 61K 行 C;
  檔案讀取常式錨點 `sub_2C740(?, 檔名, 緩衝區, 長度, ?)`、對話緩衝 `byte_54700`
  (見 `docs/re/00-hexrays-p3-verified.md`)。

## 已被推翻的斷言(單一登記處)

> **這裡是本專案唯一的推翻紀錄。** 正文(本檔、`CLAUDE.md`、`WORKLIST.md`、`docs/re/*`)
> 一律只寫**現況**,不敘述「當初怎麼錯的」—— 單獨讀到某一節的人只會看到那一節,
> 而現在式的教訓(「X 還是不存在」)會在修好的那一刻變成假斷言(`rulebook/63`)。
>
> **登記規則**:推翻一條斷言時 ① 把正文改寫成正確答案 ② 在此加一列
> ③ 若錯因是可複用的判斷失誤,把它寫成 `CLAUDE.md §4` 的**規則**(不是事件)。

| 被推翻的斷言 | 現況 | 錯因(形狀) | 出處 |
|---|---|---|---|
| 明文文字檔用 `\|` 分隔記錄 | NUL(0x00)分隔 | 拿 `strings … \| tr '\n' '\|'` 的**自己加工過的輸出**當檔案內容 | `docs/re/06` |
| 日文版有 CodeView symbol table 比較好逆 | 三版**全都沒有**;日文版的價值是反編譯器 + 雙語執行檔 + 未壓縮素材 | 拿「日文版通常有」的通例當本例的事實 | `CLAUDE.md §4.2` |
| `upgrade/` 有 VGA 美術 | 它是 MIDI 音樂升級包,`TILES.16` 與原版 md5 相同 | 憑目錄名推內容 | `CLAUDE.md §2.2` |
| PC-98 單檔所以 IDA 一次吃完 | MZ 檔頭只涵蓋 40,896 / 216,880 B,**176 KB 在自動分析之外** | 憑檔案數推分析範圍 | `CLAUDE.md §4.2` |
| `BRIT.OOL` 是世界地圖的 chunk 索引表 | 索引表在 `DATA.OVL` 0x3886 | 憑檔名推用途 | `docs/re/05` |
| `TILES.16` 的壓縮不是標準 LZW | 就是 LZW | **驗收條件本身有洞** —— 拿沒換色號的 `EGA*.TIL` 當 oracle,對的解壓器在第 5 個位元組就被判失敗 | `docs/formats/02` |
| 算繪一律用 `u5data.TilePalette` | 用 `u5data.EGAPalette`;換色號在解碼器 | 引用一個**不存在的符號** | `docs/re/13` |
| 月石是 16 顆、`0xFF` = 在手上 | **8 顆 × 四欄**(X / Y / 地點 / 樓層);`0xFF` 是「地點欄沒被寫過」 | 長度夾對了就以為語意也對;而「埋在地點 0」與「沒撿到」的症狀互相掩護 | `docs/re/71` |
| 埋下去的月石怎麼變成月門「未追」 | 月門長在月石埋藏的座標上,`sub_DEE4` 寫 / `sub_E084` 讀**同一組表** | 以為讀寫兩步用的是不同的表 | `docs/re/86` |
| `byte_3DFBB` 是繩索、位移沒對出來 | 它是**抓鉤**,位移 0x0209 早就釘死了 | `hasRope()` 的 `return false` 是**陳舊標記**,而它讓兩個機制永遠不可能發生 | `docs/re/68` |
| `.TBL` 直接給場景 → 配樂的對應,免逆向 | 表只給**曲號 → 檔名**與六聲道音量;場景對應寫在程式碼裡 | 只看表的**形狀**就推語意,沒追「誰讀它、讀去做什麼」 | `docs/re/87` |
| EUPHONY 驅動不在我們手上 | `EUPHONY DRIVER FOR TOWNES by Joe Mizuno` 在三個 `.EXP` 裡,解析器在 `TBIOS.BIN` | 沒有先做「這字串在不在」的正對照 | `docs/re/89` |
| 許願井:原版不管答什麼都沒有後續 | 有一整套彩蛋(答對車名 → 生一匹馬) | 讀的是 **Hex-Rays 安靜截斷後**的函式 —— 看起來完整、會 return、毫無警告 | `docs/re/65` |
| 戰場上的隊員圖與職業無關,只有兩個值 | 第三處(`sub_C414` 開戰佈陣)寫的是 `byte_40C34[職業]`,四個值 | 用「我手上這兩支函式」回答「還有誰寫這個欄位」 | `docs/re/72` |
| 酒是全遊戲**唯一**不議價的交易 | **餐點也不議價**;議價公式全檔只在九支函式裡 | 「唯一」沒有全檔掃描佐證 | `docs/re/93` |
| 酒館的一餐是 `Haggle(單價 × 活人數)` | 就是 `單價 × 活人數` | 與上一列**互相掩護**:「酒是唯一」讓人不查餐點,「餐點會議價」又讓「唯一」看起來成立 | `docs/re/93` |
| `sub_2C598` 的第一參數是音效索引 | 是白噪的 **Rate**;它是「DOS 白噪 → FM Towns PCM」的轉接層 | 把「別的函式第一參數是索引」搬過來,沒讀被呼叫的那支 | `docs/re/92` |
| Ztats 翻頁在隊伍範圍內繞回 | **17 頁**,繞回的接縫在 Armaments 與第一名之間 | 引擎當時只有隊員頁,拿自己的實作當原版的模型 | `docs/re/94` |
| 法術代碼表在 `DATA.OVL` 0x09B3、98 項 | 0x09A5、**48** 筆 | 0x09B3 落在第 6 筆上;98 是把後面的地名一起數進去了 | `docs/re/58` |
| 創角的初值分配未做 | 早就做了(勝方三圍累加、魔力 = 智力、力量下限 20) | 同一件事在清單上出現兩次,一列打 ✅ 一列打 ⬜ | `docs/re/39` |
| 航行節奏(頂風延遲)還沒實作 | `sailRhythm` 早就有了 | **陳舊標記**:實作完成時沒回來改那一格 | `docs/re/23` |
| 0x020D 是寶珠 | 是**不列顛王的護符**;U5 沒有寶珠(那是 U6) | 拿後續作品的道具清單套到本作 | `docs/re/44` |
| SFX 只有 A 級/B 級兩批而 WALK 是 B 級 | WALK / DOKU / TAKI2 / DAME2 都是 A 級 | 只數了 `sub_2C46C` 的 17 個直接呼叫點,沒發現還有一層轉接 | `docs/re/92` |
| upgrade 的 `.XMI` 有 **19 首** | **15 首**(15 個檔各一個 `EVNT` chunk)| 數了檔名清單裡的詞而不是數檔案 | `docs/formats/13` |
| 怪物移動本身全檔找不到 | 在 `sub_2D38` 的**最後一行** | 讀了那支函式但沒讀到尾巴 —— 「找不到」的成因是自己沒讀完 | `docs/re/85` |
| `sub_27F0` 為 0 → 移動,否則不動 | 兩條分支行為相同(`sub_2B24` 的第二參數 8 寫 0 讀) | 假設參數會被用到,沒查它有沒有讀取點 | `docs/re/88` |

## 硬規則(完整版見 `CLAUDE.md`)

- 編譯一律 docker;Python 走 docker uv/venv,不污染系統。
- **素材一律用原版**(`CLAUDE.md §3.0`),唯一例外是中文字型。
- 引擎與資料分離:原版資料、IDA 產物、抽出的素材**全部不入庫**;手冊 OCR 例外(使用者明示)。
- 存檔寫 `os.UserConfigDir()`,不寫 cwd。
- 驗收對 reference(DOSBox 原版)實測,不靠自家測試綠(`rulebook/65`);另驗**無 debug 的正常玩家路徑**。
- 每完成一段落 commit + push `github.com/wicanr2/u5-cht`;回應與文件用繁體中文。
