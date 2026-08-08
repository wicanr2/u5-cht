# 77 — 一個假缺口(木盒),以及 U 的收尾那句 `Failed!`

| | |
|---|---|
| 輸入檔 | `re_work/fmtowns/WORRIORS.EXP.asm`(FM Towns 英文版) |
| SHA-256 | 見 `re_work/fmtowns/SHA256SUMS.txt` |
| 主要函式 | `sub_1A5E8`(Use 指令的收尾與 case 37) |
| 落地 | `internal/game/use.go` · `potion.go` · `strings.go` |

---

## 1. ★ 木盒不是缺口 —— 是我把 IDA 的自動命名當成了原版的字串

`use.go` 原本寫著:

```go
case item == UseWoodenBox:
	// 原版印「Box- How?」再依答案分支,而那條路還沒逆完。
	s.Log(MsgBoxHow)          // "(木盒的用法尚未實作)"
	return false
```

實際的 case 37(`loc_1AB32`)只有兩行:

```asm
loc_1AB32:
        push    offset aBoxHow  ; jumptable 0001A6DD case 37
loc_1AB37:
        call    sub_23C18
```

而那個字串是:

```asm
aBoxHow         db 'Box',0Ah
```

⇒ **就只有 `"Box"`。** 沒有「How?」、沒有問句、沒有分支。

`aBoxHow` 是 **IDA 依鄰近內容自動取的名字**,把後面別的字串的字尾黏進來了。
我讀的是那個**工具產生的標籤**,不是它指向的位元組。

★ 教訓很具體:**`aXxx` 這種自動名字不是資料。** 要引用一個字串的內容,
得去看它的 `db` 那一行。這與 `docs/re/67` 的
「`off_48A88 dd offset loc_202C` 其實是內嵌字面值 `", "`」是同一類 ——
**IDA 的命名與型別判斷都只是猜測**,而它猜錯時看起來與猜對時一模一樣。

檀香木盒真正的用途在別處(不列顛王城堡那條線,`docs/re/36`);
U 對它就是印一個名字。⇒ 引擎少的不是功能,是一句誠實的話。

## 2. `Failed!` 是卷軸與藥水專屬的收尾

`sub_1A5E8` 的骨架:

```asm
mov [ebp+var_10], 1                  ; ★ 預設 1
…
if (n < 8)   { eax = sub_19ED8(n);     jmp loc_1A6AB }   ; 卷軸
if (n < 16)  { eax = sub_1A0B0(n − 8); jmp loc_1A6AB }   ; 藥水
loc_1A6AB: mov [ebp+var_10], eax     ; ★ 只有這兩段會覆寫
…
if (n > 20 && n < 29) { sub_1A2F8(n − 21); jmp loc_1AB3C }   ; 月石 —— 跳過賦值
switch (n − 16) { … }                                        ; 其餘 —— 也跳過
def_1A6DD:
    if (var_10 == 0) { 印 "Failed!"; 音效 }
```

⇒ **月石與其餘 22 種道具永遠不會印 `Failed!`**,只有卷軸與藥水會。
引擎原本完全沒有這句收尾。

### ⚠ 一個容易多印一句的地方

`sub_1A0B0`(藥水)在「沒選到人」時走的是 `jmp loc_1A2F1` ——
那條路**跳過 `mov eax, esi`**,直接把 `eax`(= 負的目標索引)當回傳值。
負數不是 0 ⇒ **不會**印 `Failed!`。

所以 `DrinkPotion` 在那條路要回 `true`,不是 `false`。
回 `false` 會讓「取消選人」也吐一句「失敗!」——
而那是原版沒有的、玩家看得到的差異。

## 3. 順帶更正 `use.go` 的道具編號表

我先前在 `docs/re/71` 為 `use.go` 補的那張表,後六筆猜錯了兩筆
(把圖紙 / 徽章 / 木盒排在 32/33/34)。用**跳表自己的 case 標註**核對:

```
aSpyglass        → case 32     byte_3DFC8
aPlans           → case 33     byte_3DFC9      ↔ 存檔 0x0215(已驗)
aSextant         → case 34     byte_3DFCA
aWatchThePocket  → case 35     byte_3DFCB
aBadge           → case 36     byte_3DFCC
aBoxHow(= "Box") → case 37     byte_3DFCD      ↔ 存檔 0x0219(已驗的檀香木盒)
```

兩端都有已驗過的存檔位移夾住,中間四筆沒有滑動空間。
`use.go` 的常數本來就是對的 —— **錯的是我後來補的那張說明表**,
而那正是「文件與程式碼不一致時,先信程式碼」的一次實例(`rulebook/63`)。

## 4. 測試

- `TestTheWoodenBoxOnlyPrintsItsName` —— 要成功、要印名字、
  **而且不准再出現「尚未實作」**。
- `TestFailedIsOnlyPrintedForScrollsAndPotions` —— 卷軸失敗要印、藥水失敗要印、
  **月石失敗不准印**。
  ⚠ 藥水那一段第一版沒重試就紅了,訊息是「眼前一片通透……」= **白**藥水的句子
  (1/8 走偏,`docs/re/71`)。**同一個坑第二次**,已在測試裡註明。
