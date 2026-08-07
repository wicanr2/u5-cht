# 60 — 每個指令先印自己的名字;而選單**沒有**字母捷徑

| | |
|---|---|
| 輸入檔 | `U5_E/WORRIORS.EXP`(FM Towns,SHA-256 見 `docs/re/00-hexrays-p3-verified.md`) |
| 位址 | `sub_2ACF4`(59 case 主分派器)、`sub_1F3A4`(Ready)、`sub_1E418`(找下一個非空欄位) |
| | `sub_1EFC8`(清單瀏覽器)、`sub_CAC` / `sub_4074` / `sub_2D478`(Attack 三支) |
| 落地 | `internal/game/commandecho.go`、`internal/game/picker.go`、`cmd/u5cht/main.go` |

`WORKLIST` 上兩條相關的 ⬜:

```
指令系統  ⬜ 剩原版每個指令會先印指令名
指令選單  ⬜ 原版的字母捷徑未核
```

第一條是真的漏了。第二條**是我自己記錯的前提** —— 原版沒有字母捷徑。

## 一、每個指令先印自己的名字

`sub_2ACF4` 的每個 case 幾乎都以一個 `push offset a…` 開頭,印完才呼叫處理函式:

```asm
                push    offset aBoard_0      ; "Board "     jumptable case 66 ('B')
                call    sub_23C18
                call    sub_16F08
```

逐 case 抽出來(索引 = 鍵碼 − 0x20):

| 鍵 | 原版印的 | 中譯 |
|---|---|---|
| 空白 | `Pass\n` / 揚帆時 `Sheets in irons!\n` | 由 `Pass` 自己印 |
| `A` | **(不印)** | 三個位置的處理函式各自印 `Attack-` / `Attack` |
| `B` | `Board ` | 登乘 |
| `C` | `Cast...` | 施法…… |
| `D` | `D-What?` | D—— 何事? |
| `E` | 地表 **(不印)** / 其餘 `Enter what?\n` | 進入什麼? |
| `F` | `Fire-` | 開砲 —— |
| `G` | `Get-` | 拿取 —— |
| `H` | `Hole up- ` | 歇息 —— |
| `I` | `Ignite torch!` | 點燃火把! |
| `J` | `Jimmy-` | 撬鎖 —— |
| `K` | `Klimb-` | 攀爬 —— |
| `L` | `Look` + 地牢 `...\n` / 其餘 `-` | 觀察…… / 觀察 —— |
| `M` | `Mix Reagents` | 調配藥草 |
| `N` | `New Order` | 重排隊伍 |
| `O` | `Open-` | 開啟 —— |
| `P` | `Push` | 推動 |
| `Q` | `Quit:` | 存檔: |
| `R` | `Ready...` | 裝備…… |
| `S` | 場景 / 地表 `Search-` / 地牢 `Search...\n` | 搜尋 —— / 搜尋…… |
| `T` | `Talk-` | 交談 —— |
| `U` | `Use item` | 使用道具 |
| `V` | `View a gem!` | 觀看寶石! |
| `W` | `W-What?` | W—— 何事? |
| `X` | `X-it ` | 下載具 |
| `Y` | `Yell ` | 呼喊 |
| `Z` | `Z-stats...\n` | 角色數值…… |
| 其他鍵 | `What?\n` | 此為何意? |

### 那個結尾的 `-` 不是標點,是「等方向」的提示

原版的訊息欄是一條**逐字累加**的紙帶:按 `G` 印 `Get-`,方向鍵再把 `North`
接在**同一行**後面 → `Get-North`。所以:

- `-` 結尾的指令(`Fire` `Get` `Jimmy` `Klimb` `Open` `Search` `Talk` `Hole up`,
  以及**非地牢的** `Look`)就是要問方向的那幾支。
- **少了它,玩家按下 G 之後畫面毫無反應**,而症狀看起來像「按鍵沒吃到」。

最乾淨的佐證是 `L`:它先印 `Look`,然後**依位置**選擇印 `...`(地牢,直接看腳下)
或印一個裸的 `'-'`(其餘,問方向)。同一支函式裡兩條路各接一種結尾 ——
破折號的用途在這裡寫得最清楚。

落地:`commandEcho` 表 + `CommandEcho(key)` 處理依位置變化的那五個;
名字以「——」結尾時 `AskDirection` **不另起一行**,`AnswerDirection` 把方向
`Append` 到同一則(`TestDirectionIsAppendedToTheCommandName`)。

⚠ **版面差異**:原版整條訊息欄是單一紙帶,本引擎是一行一則的 CJK 版面。
只有「指令名 + 方向」這一組合併,其餘各佔一行。這是版面,不是機制。

## 二、選單**沒有**字母捷徑 —— 一條長期誤記

`internal/game/picker.go` 的註解原本寫:

> ⚠ 原版用**字母**選(「Item: 」後面按一個字母)。引擎用方向鍵 + Enter……
> 字母對應哪一項在跳表裡看得出來有,但沒逐一核過,所以先不做。

把原版的清單瀏覽器 `sub_1EFC8` 整支掃過,**它比對過的鍵碼只有**:

| 鍵碼 | 作用 |
|---|---|
| 1, 3 | 游標上移一項 |
| 2, 4 | 游標下移一項 |
| **0xD5 / 0xD6** | 上 / 下移動 **7 項**(`var_8 = 7` 之後跑同一個迴圈) |
| 0xD3 / 0xD4 | 另兩條分支(`loc_1F2E5` / `loc_1F303`,未追) |
| 13(Enter)、0x20(空白) | 選定 |
| 0x1B(ESC) | 放棄 |

**0x41..0x5A 一次都沒出現。** 所以 `Item: ` 是「目前停在哪一項」的標籤,
不是「請按字母」的提示 —— 引擎的方向鍵 + Enter **本來就是照原版**,
那條 ⬜ 的前提從一開始就不成立。

真正漏掉的只有一件:**翻頁一次移 7 項**(不是「一整頁」——清單再長也是 7)。
已補 `PickPage` 與 PgUp / PgDn 鍵。

### 順帶釐清 Ready 的三層結構

```
sub_1F3A4:  sub_1DE10(0)                                → 選人(< 0 取消)
            sub_1E418(-1, 0x30, byte_3DFD0, 人)          → 第一個「這個人用得到」的欄位
              -1 → "Thou art empty-handed!"
            印 "Item: " + 那個人的名字
            sub_1EFC8(起始欄位, 人, 'R')                  → 清單瀏覽器(上面那套鍵)
```

`sub_1E418` 只是「往後找下一個非空且該員用得到的欄位」的掃描器,不是清單本身;
瀏覽器 `sub_1EFC8` 靠它前後移動游標。`'R'` 那個參數決定用哪張名字表
(`R` → `byte_3DFD0` 裝備、否則 `byte_40BA0`),所以 **R 與 U 共用同一支瀏覽器** ——
這與 `docs/re/46` 把五支收成一個選單的做法方向一致。

## 這次改了什麼

- 新增 `internal/game/commandecho.go`:指令名回顯表 + 依位置變化的五個 + `What?`。
- `cmd/u5cht/main.go` 的 `commandKeys` 由一個大 `switch` 改成 **`commandTable` 資料表**
  (按鍵、回顯字母、處理函式),回顯因此只有一個地方會漏 ——
  而原本那個 switch 每加一個指令都得記得補一次回顯。
- `AskDirection` / `AnswerDirection` 支援「接在指令名後面」。
- `PickPage`(7 項)+ PgUp / PgDn。
- 更正 `picker.go` 的字母捷徑誤記。

## 還沒追的

`sub_1EFC8` 裡 0xD3 / 0xD4 兩個鍵碼的分支(`loc_1F2E5` / `loc_1F303`)。
它們與 0xD5 / 0xD6 相鄰,很可能是 Home / End 或「換人」,但**沒讀就不寫**。
