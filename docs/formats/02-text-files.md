# 格式規格 02:文字資料檔

> 狀態:**八類文字檔全部已解**。`SHOPPE.DAT` 的詞典 token 映射已定(空槽十個);
> `LOOK2` / `SIGNS` 的格式見 `docs/re/37`,兩份都只是「位移表 + 字串」。
> 實作 `internal/u5data/text.go`,驗收工具 `u5dump text`。

## 0. 一則更正(方法論教訓)

此前本專案記載「明文檔用 `|` 分隔記錄」—— **這是錯的**。`STORY.DAT` 裡 `0x7C`(`|`)
出現 **0 次**。錯誤來源是當初檢視時用了:

```bash
strings -n 5 STORY.DAT | head -2 | tr '\n' '|'     # ← 這個 `|` 是 tr 自己加的
```

輸出裡的 `|` 是**自己的管線造成的**,卻被當成檔案內容寫進了 CLAUDE.md、CONTEXT.md 與 commit。
**一手資料(原始位元組)贏二手推論(工具輸出)** —— 下任何格式斷言前,回去數那個位元組出現幾次。

## 1. 明文訊息檔(NUL 分隔)

記錄分隔是 **NUL (0x00)**。實測(2026-08-07):

| 檔案 | 大小 | 記錄數 | `_` 斷字提示 | 詞典 token | 內容 |
|---|---|---|---|---|---|
| `STORY.DAT` | 11,679 B | 20 | 654 | 0 | 開場故事(月門、Shadowlords) |
| `QUESTION.DAT` | 7,746 B | 30 | 116 | 0 | 吉普賽問答(創角道德兩難) |
| `KARMA.DAT` | 761 B | 6 | 0 | 0 | 業報訊息 |
| `MISCMSG.DAT` | 2,745 B | 47 | 4 | 0 | 系統訊息(聖壇真言提問等) |
| `ENDMSG.DAT` | 786 B | 11 | 0 | 0 | 結局訊息 |
| **`SHOPPE.DAT`** | 10,135 B | 194 | 0 | **862** | 商店對白 |

兩個標記:

- `{`:段落 / 換頁起始(只出現在 `STORY.DAT` 36 次、`QUESTION.DAT` 5 次)。
- `_`:**英文斷字提示**(`be_gin`、`mys_te_ri_ous`、`no_where`),原版用它決定一行放不下時在哪斷。
  **中文化一律移除** —— 中文不做音節斷字,留著會在畫面上變底線。
  譯文側的 `_` 數量應為 0,可當回歸檢查(`TextFile.HyphenHintCount()`)。

⇒ **前五個檔共 114 筆、0 個 token,可以直接進翻譯流程。**

## 2. 詞典壓縮(部分破解)

`SHOPPE.DAT` 裡位元組 **≥ 0x80 不是文字,而是常用詞代碼**:

```
"Thanks\x86nothing!\""            \x86 → "for"   ⇒ Thanks for nothing!
"…ready\x83buy something!\""      \x83 → "to"    ⇒ ready to buy something!
```

字典在 **DOS 版 `DATA.OVL` offset `0x104C`**,連續的 NUL 結尾字串,**118 個詞**:

```
the thou of to and that for in is have with thee this not my it me but dost know
be was Blackthorn from thy one are here many Lord am we they he would art on young
what see like only by there Blackthorn's good been must his British fine an great
thee, our who name heard as at has through once can him ye Shadowlords tell some
believe all their upon even 'tis find if about don't before these just make will
when three Great might those old hast ask unto wish man so knows still Mantra out
help well shall think where named talking more such very may lives canst which
since need I've work
```

字典結束後緊接**檔名表**(`PROPORT.PCS`、`BRITISH.BIT`、`TITLE.BIT`、`TILES.16`…),
所以邊界在第 118 個詞。字典內含 `Blackthorn`、`Shadowlords`、`Mantra`、`British` ——
這些 U5 專有詞出現在常用詞表裡,本身就佐證位置找對了。

### 2.1 token → index 的映射:未定 ⚠

| token | 需要的 index | 字典實際位置 | 差 |
|---|---|---|---|
| `\x86` = "for" | 6 | index 6 ✓ | 0 |
| `\xD7` | 87 | "about" 在 index 77 | 10 |
| `\xDE` | 94 | "when" 在 index 84 | 10 |

線性 `token - 0x80` 對第一個成立、對後兩個差 10。而 token 空間 `0x80–0xFF` 有 128 個、
字典只有 118 個 ⇒ **有 10 個 token 不是詞**(推測是控制碼:換行、玩家名代入之類),
這很可能就是那個 10 的來源,但**確切是哪 10 個、插在哪裡,還沒證據**。

> ⚠ 那兩個「需要的 index」是從英文語感推來的(`"Come back … you're ready to buy"` 猜成 "when"),
> 本身就是二手推論。硬套下去只會得到似通順但錯的譯文。
>
> **移交 P3**:FM Towns `WORRIORS.EXP` 可反編譯,裡面必然有展開 token 的字串輸出函式,
> 讀它就有確定答案。在那之前 `TextRecord.Text()` 把 token 保留成 `<XX>`,
> `HasTokens()` 為 true 的記錄**不得進翻譯流程**。

## 3. 另案處理

| 檔案 | 觀察 | 狀態 |
|---|---|---|
| `LOOK2.DAT` 3,622 B | **u16 × 512 位移表(1024 B)+ NUL 字串**;512 格對上 tileset 的 512 格 | ✅ 已解(`docs/re/37`)。⚠ 此前記載的「218 個 NUL + 大量控制碼」是誤讀:那些「控制碼」是位移的高位位元組(資料落在 0x0400–0x0E17)。錯在先數位元組分布再猜格式,而沒先試「檔頭是不是一張表」 |
| `SIGNS.DAT` 8,364 B | **u16 × 33 位移表 + 記錄** `[地點][樓層][x][y]` + 內容 + NUL,表尾 0xFF;78 塊 | ✅ 已解(`docs/re/37`)。**框是美術、字是字**,而且框的美術在字型裡(`RUNES.CH`:`l` 是橫線、`g` 是直線、`&` 字模全空 —— 所以程式才要把 `&` 特判成印 `l`)。文字整段可譯 |

## 4. DATA.OVL 裡順手發現的其他文字資源

掃字典時在 `DATA.OVL` 找到的其他明文區(都是中文化目標,待各自建表):

| offset | 內容 |
|---|---|
| `0x06EE` | 法術與材料:`Pearl` `Nightshade` `Mandrake` `In Lor` `Grav Por` `An Zu` `An Nox` `Mani` … |
| `0x0919` | 狀態 + 符文詞:`Good Health` `Poisoned` `Dead` `Asleep` `Charmed` / `AN BET CORP DES EX FLAM GRAV HUR IN KAL LOR MANI NOX POR QUAS REL SANCT TYM UUS` |
| `0x09B3` | 符文組合縮寫:`AY` `AS` `ACX` `HR` `IW` `KX` … |
| `0x0FBF` | NPC 人名:`Toama` `Enlor` `Virden` `Regina` `Leila` `Jessica` `Faye` `Donya` … |
| `0x104C` | **常用詞字典**(見 §2) |
| `0x11xx` 之後 | 檔名表 |
