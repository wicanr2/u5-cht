# 文字壓縮:一份詞典,兩種極性

> 來源:FM Towns `WORRIORS.EXP` 的 `sub_1C3F8`(展開)與 `dword_41990`(槽表);
> 詞典本文在 DOS 版 `DATA.OVL` 0x104C。

遊戲裡幾乎所有英文文字都用同一份 **118 字常用詞詞典**壓縮。這件事很容易錯判,
因為錯的解法「看起來只差一點」:把 `.TLK` 整份清掉 bit7,得到

```
I study·stars.          ← 讀起來像混進一個雜訊字元
```

實際上那個 0x01 是 `the`,正確答案是 `I study the stars.`。

## 兩種極性

| | 字面文字 | token | 槽位 |
|---|---|---|---|
| `.TLK` | bit7 **有設**(取 `b & 0x7F`) | bit7 沒設 | `slot = b` |
| `.DAT`(SHOPPE 等) | bit7 **沒設**(純 ASCII) | bit7 有設 | `slot = b - 0x7F` |

兩邊對到同一張 1..128 的槽表。

`.DAT` 的偏移最初是實測推出來的,後來在商店文字的輸出函式 `sub_10FEC` 裡找到程式面佐證:
它對 `b >= 0x80` 查 `dword_41794[b*4]`,而 `0x41794 + 0x80*4 = 0x41994`,
正好是 `.TLK` 槽表 `dword_41990` 的第 1 格("the")—— 也就是 `slot = b - 0x7F`。

## 展開規則(`sub_1C3F8`)

```asm
cmp  bl, 81h
jnb  → 字面分支
push 0A0h                    ; 0xA0 = ' ' | 0x80 → token 前先輸出一個空格
call sub_1B800
edi = dword_41990[token]
while (*edi) 輸出 (*edi++ | 0x80)
if (*dword_41990[token] == 0) 輸出原位元組     ; 空槽 → 原樣輸出
else byte_55F4A = 1                            ; 設下 pendingSpace

字面分支:
or   bl, 80h
cmp  bl, 8Dh ; jnz → skip ; mov bl, 8Ah        ; CR 換成 LF
if (byte_55F4A) 輸出 ' '                       ; 補上 pendingSpace
輸出 bl; byte_55F4A = 0
```

也就是:**token 自己前置一個空格,後置空格由下一個字面字元補**。
連續的 token 之間因此不會出現雙空格 —— 這個細節錯了,句子會黏在一起或多出空白。

驗算:

```
"I study" + tok(the) + "stars."
  → "I study" + " the" + [pending] + " " + "stars."     = "I study the stars."

"a stately, white-haired" + tok(man) + tok(of) + tok(many) + "years."
  → "…white-haired" + " man" + " of" + " many" + " " + "years."
                                              = "a stately, white-haired man of many years."
```

## 槽表有 10 個洞

`dword_41990` 是**用 token 值直接索引**的指標表,中間有 10 個 NULL:
**8、28、50、65、67、70、72–75**。

DOS 版的 `DATA.OVL` 則存成緊密的 118 字清單,所以要把清單塞回槽位時得跳過這些洞。
128 − 10 = 118 ✓

洞的位置可以獨立驗證:統計四個 `.TLK` 的 token 值,這 10 個值**一次都沒出現過**。

> 這也解釋了先前分析 `SHOPPE.DAT` 時觀察到卻說不出所以然的「token 與清單索引固定差 10」
> —— 差的就是這 10 個洞;最後一個洞在 75,之後差值才固定。當時把它當成無法解釋的常數,
> 現在它是結構的直接推論。

## `.DAT` 極性的證據

同一批句子只有在 `slot = b - 0x7F` 之下才讀得通:

| 壓縮 | 展開 |
|---|---|
| `Thanks {86} nothing!"` | Thanks **for** nothing!" |
| `Come back {94} you're ready {83} buy something!"` | Come back **when** you're ready **to** buy |
| `Be off {8B} ye, then..."` | Be off **with** ye, then..." |
| `Our Iron Helms {9C} padded` | Our Iron Helms **are** padded |

最後一例特別有力:0x9C 若照 `.TLK` 的算法會落在**空槽**(28),照 `.DAT` 的算法(29)才有字。

驗收:`SHOPPE.DAT` 的 194 筆記錄展開後**沒有任何未解的 token**。

## `.TLK` 記錄結構

`sub_1C840(id)`:

```
file = TLKFiles[(地點編號-1)/8]
read(file, 0, 0x200)                 ; 檔頭:int16 筆數 + 筆數 × {int16 id, int16 offset}
線性搜尋 id 相符的那一筆                ; ← id 不是陣列下標
read(file, offset, 0x400)            ; 讀出記錄本體
```

記錄本體是 NUL 分隔的段落,**多數**記錄的前五段是:

| # | 內容 |
|---|---|
| 0 | 名字 |
| 1 | 「看」到的樣子 |
| 2 | 打招呼 |
| 3 | 問 job / work 的回答 |
| 4 | 道別 |

之後是「關鍵字 → 回應」的成對資料。

⚠ 不是每一筆都乾淨:有些記錄的名字後面還跟著控制位元組,有些第 2 段是空的。
記錄裡混著對話引擎的指令(原始 0x81–0x9F),`sub_1C3F8` 有一張跳表在處理它們
(cases 133/134/140/141/143/144/145–159)。**整套腳本語意還沒解**;
把控制碼濾掉之後文字本身是完好的,足以拿來翻譯與顯示。
