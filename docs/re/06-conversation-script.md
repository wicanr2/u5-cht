# 對話腳本引擎

> 來源:FM Towns `WORRIORS.EXP`。核心是 `sub_1C3F8` —— 一個**逐位元組的解譯器**,
> 帶一張 31 路跳表;周邊有 `sub_1C840`(載入記錄)、`sub_1B52C`(分派)、
> `sub_1BB3C`(插名字)、`sub_1BB5C`(加入隊伍)、`sub_C10`(叫衛兵)、
> `sub_1C0AC`(反問玩家)。

`.TLK` 的記錄不只是文字,是一段**腳本**。把 0x81–0x9F 當文字印出來會滿畫面控制字元;
把它們丟掉則會失去加入隊伍、業報增減、呼叫衛兵這些**真的會改變遊戲狀態**的效果。

## 記錄的排法

```
段 0..4     名字 / 外貌 / 招呼 / 職業 / 道別
段 2i+5     第 i 個關鍵字        段 2i+6  它的回應
0x90 之後   提問區塊(見下)
```

算式不是猜的:`sub_1BD50(i)` 把指標重設到開頭後跳 `2i+5` 段,命中之後
`sub_1BF08` 用 `sub_1BAFC(i*2 + 6)` 印回應。

**關鍵字表在遇到位元組 0x90 時結束。** 跳段用的 `sub_1BA80(0, 0x90)` 是
「前進到 NUL(成功)或 0x90(失敗)」。少了這個終止條件,0x90 之後的提問區塊
會被當成關鍵字對 —— 而且錯得很難察覺,因為每個字都「有」答案,只是答錯。
實測影響很大:四個 `.TLK` 的關鍵字從 1,767 掉到 **1,307**,26% 是假的。

## 關鍵字比對不是「比前 4 個字母」

原版把記錄裡的關鍵字當**子字串**去玩家輸入裡找,而且命中位置必須落在詞首
(位置 0,或前一個字元是空白)——`sub_1BD8C` → `sub_27C98`,再檢查 `byte_55F37[i]`。

```
關鍵字 "bow"  ← 輸入 "bow" ✓   "bows" ✓   "my bow" ✓   "elbow" ✗
```

這與截斷比對不同:`bows` 截成 4 字仍是 `bows`,配不上 `bow`。記錄裡短關鍵字很多,
搞錯會讓一堆問話變成「聽不懂」。

## 內建關鍵字表(34 個)

`off_55E88` 是一張引擎自己的關鍵字表,**掃描順序在記錄的關鍵字之前**:

| 索引 | 字 | 行為(`sub_1BE28`) |
|---|---|---|
| 0 | NAME | 印 `"My name is "` + 段 0 |
| 1, 2 | JOB, WORK | 印段 3 |
| 3, 4 | BYE, THANK | 印段 4 並結束對話 |
| 5–33 | 29 個髒話 | 一律回 `"With language like that, how did you become an Avatar?"` |

髒話那一段是原版就有的內容,照實收錄 —— 少了它,對 NPC 罵髒話會變成沒反應,
與原版行為不同。

## 提問區塊

關鍵字表之後是一連串提問區塊:

```
0x90 <碼> <問題文字> ¶ <「否」的回答> ¶ <「是」的觸發字> ¶ <「是」的回答> ¶
```

碼是 0x91–0x9F。某個關鍵字的回應裡出現同一個碼時,引擎跳到對應區塊發問
(`sub_1C0AC` → `sub_1BCB8` 從記錄開頭找碼相同的區塊)。玩家回答後:
輸入含「是」的觸發字 → 印第 4 段(`sub_1BCF4`),否則印第 2 段(`sub_1BD0C`)。

**終端區塊**只有問題文字、沒有分支:印的過程中碰到會讓 `sub_1C3F8` 回傳 1 的指令
(0x82 結束回應、0x84 入隊)就停,不向玩家要輸入。

Gwenno 的資料是完整的兩段式範例:

```
"…Deep Forest."[8D][8D][91]          ← yew 的回應拋出 0x91
join ¶ [94] ¶                         ← join 的回應拋出 0x94
[90][91] Say, art thou from around here? ¶ I thought not. ¶ y ¶ Then perhaps just not too smart! ¶
[90][93] Aren't ye going to ask me to join with thee? ¶ I am Iolo's better half!"[82] ¶ y ¶ [94] ¶
[90][94] Iolo and I both thank thee!"[8D][8D][84]      ← 終端:道謝並入隊
[90][9F] @                            ← 結束標記
```

## 加入隊伍(`sub_1BB5C`)

1. 隊伍滿 6 人 → 「Thou hast no room for me in thy party」,不入隊。
2. 把腳本指標**重設到記錄開頭**(`sub_1BAA4(0)`),讀 3 個位元組 —— 那是 NPC 名字的前三個字母。
3. 從名冊**尾端往回**掃(15 → 1),找名字前三個字母相符的那一筆(遮 bit7、大小寫無關)。
4. 把它與名冊第 `PartySize` 格**對調**,人數 +1。所以「隊伍」不是另一個清單,而是名冊的前綴。
5. 把該 NPC 從場景移除。

**回應只有一個 `0x87` 代表「同下一則」**,而且可以連續好幾層。占星師 Zachariah 的
`tele`(telescope)與 `star` 就共用同一段回答:

```
tele ¶ [87] ¶ star ¶ I watch the signs amongst the planets! ¶
sign ¶ [87] ¶ plan ¶ Comets have come! A sign of evil! ¶
evil ¶ [87] ¶ come ¶ There are three comets in the firmament… ¶
```

原版 0x87 的作法是把文字指標存起來、往下讀一則再還原(`dword_55F14` 的存取還原)。

## 指令集

`sub_1C3F8` 先判 0xFF、0xFE,再用 `edx = b - 0x81` 查 31 路跳表;其餘走詞典/字面路徑。

| 碼 | 用途 | 依據 |
|---|---|---|
| 0x81 | 插入聖者的名字 | `sub_1BB3C` 逐字印 `byte_3DDB4` |
| 0x82 | 結束這一則回應 | 直接 `return 1` |
| 0x83 | 停頓(可按鍵略過) | 0x1C 次的等待迴圈 |
| 0x84 | 邀請加入隊伍 | `sub_1BB5C`,滿員時「Thou hast no room for me…」 |
| 0x85 / 0x86 / 0x8C / 0xFE | 切到子模式 | `byte_55F18 = 碼`;函式開頭會把後續位元組改送 `sub_1C1E8` |
| 0x87 | 同下一則 | 存還原 `dword_55F14` |
| 0x88 | `sub_1C2FC` | 未定 |
| 0x89 / 0x8A | **業報** +1 / −1(上限 99) | `sub_2BBB8(&byte_3E098, 1, 0x63)` / `sub_2BBFC` |
| 0x8B | 叫衛兵 | `sub_C10` 掃 32 個 NPC 槽找 tile 0x70 那批 |
| 0x8D | 換行 | 字面路徑把 0x8D 轉成 0x8A |
| 0x8E | 切換強調 | `byte_55F1A ^= 0x80`,影響後續字元的輸出屬性 |
| 0x8F | 等待按鍵 | `sub_29EEC` |
| 0x91–0x9F | 反問玩家並讀取回答(15 種) | `byte_55F32 = 碼; sub_1C0AC` → 印 "You respond-\n:" 後收輸入 |
| 0xFF | 結束整段對話 | `sub_1B760` + `sub_1BF08` |

`byte_3E098` 就是業報:上限 0x63 = 99,而遊戲裡有 `KARMA.DAT`。

## 關鍵字比對的一個坑

記錄裡有多字關鍵字(`art thou`、`who art`、`how many`)。截到 4 個字元剛好切在空白上,
變成 `"art "`。**正規化必須「先截斷再去尾空白」** —— 順序顛倒的話,存進去的是 `"art "`、
查詢時算出來的是 `"art"`,同一個字自己對不上自己,症狀是 NPC 明明列了這個關鍵字
卻回「聽不懂」。

截斷後去尾空白也順帶讓玩家打 `art` 就命中 `art thou`。原版是硬比 4 個位元組,
這樣比它寬鬆;寬鬆的方向對玩家有利,而且不會誤命中(前 4 字元相同本來就是同一個)。

## 驗收

- 四個 `.TLK` 共 **135 段對話、1,307 個關鍵字**全數解析,0 段無名,
  每個 NPC 自己列出的關鍵字都答得出來(別名鏈若成環會在這裡吊死,不會靜默通過)。
- 畫面驗收:對廚師 Justin 問 `job` → `reci` → `secr`,依序得到
  "Why, I am the cook, of course!" / "That, my friend, is a family secret!" / "I won't tell!"。
- **入隊驗收**:走到 Gwenno 旁邊問 `join` → 「Iolo and I both thank thee!」→
  她進入隊伍、從場上消失(居民 11 → 10)。
- **是非題驗收**:問 `yew` → 她反問「Say, art thou from around here?」→
  答 `y` 得「Then perhaps just not too smart!」、答 `n` 得「I thought not.」。
