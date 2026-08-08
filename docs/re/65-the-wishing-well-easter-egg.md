# 65 — 許願井:Hex-Rays 把一整個彩蛋藏起來了

| | |
|---|---|
| 輸入檔 | `re_work/fmtowns/WORRIORS.EXP`(FM Towns 英文版,Phar Lap `P3`) |
| SHA-256 | 見 `re_work/fmtowns/SHA256SUMS.txt` |
| 主要函式 | `sub_CD28`(井)、`sub_D258`(Look 分派)、`sub_27C98`(字串比對)、`sub_2B770`(文字輸入)、`sub_2B6C8`(寫物件槽)、`sub_2B57C`(找空槽) |
| 落地 | `internal/u5data/well.go`、`internal/game/well.go` |
| 這條在 WORKLIST 上原本是 | 「⬜ 0x725C 那組彩蛋名字(`Ferrari Lamborghini …`)還沒找到觸發條件」 |

---

## 1. 為什麼會找了這麼久

`internal/game/strings.go` 原本掛著這句結論:

> "a deep well"。原版接著問「Drop a coin?」但**不管答什麼都沒有後續** ——
> 組語裡沒有任何寫入,所以這裡也不做成互動,一句話帶過。

**每一個字都錯了**,而錯因不是粗心 —— 是**讀了反編譯的輸出**:

```c
char sub_CD28()                    // ← 連參數都沒了
{
  sub_23C18("a well.\n\nDrop a coin?");
  do v2 = sub_29EEC(v1, v0); while ( v2 != 89 && v2 != 78 );
  if ( v2 == 78 ) return sub_23C18("No\n");
  else            return sub_23C18("Yes\n");
}
```

看起來確實「答什麼都沒有後續」。但組語裡這支函式有 **三個參數**、
長度是反編譯版的四倍,`"Yes\n"` 那一行之後才是真正的內容。

這正是 `CLAUDE.md` §4.4 那條 `[HARD]` 講的事:
**Hex-Rays 是加速器,不是真值來源。** 而這次的症狀比「函式被摺成一行」更陰險ˇ——
它產出的是一段**看起來完整、邏輯自洽、還會 return** 的函式。沒有警告、沒有
`// write access to const memory`,只有一個沉默的截斷。

⇒ **判斷法要補一條**:反編譯出來的函式若「參數列是空的、而呼叫端明顯有推參數」,
那就是被截斷了,回去讀組語。

---

## 2. 完整流程

`sub_D258`(Look 的地形分派)在 tile **0xA1** 時呼叫:

```asm
loc_D2C1:
        cmp     edi, 0A1h
        jnz     short loc_D2EE
        movzx   eax, byte_3E0A5     ; 樓層
        push    eax
        movzx   eax, byte_3E0A7     ; 隊伍 Y
        push    eax
        movzx   eax, byte_3E0A6     ; 隊伍 X
        push    eax
        call    sub_CD28            ; ★ 三個參數
```

三個全域都是既有筆記裡定過的(`byte_3E0A5` 見 `docs/re/11`、
`byte_3E0A6` / `byte_3E0A7` 見 `docs/re/48`)—— 兩份獨立來源,不必猜。

```
印 "a well.\n\nDrop a coin?"
等 'Y' 或 'N'(其他鍵繼續等)
'N' → 印 "No\n"、結束
'Y' → 印 "Yes\n"
      if (word_3DFB6 == 0) → 結束          ★ 一句話都不印
      印 "\nThy wish?\n"
      word_3DFB6--                         ★ 先扣一枚,不論許什麼願
      sub_2B770(緩衝, 0Ch)                 讀最多 12 個字元
      緩衝[0] == 0 → 印 "Nothing\n"、結束
      六次 sub_27C98 比對:
          "Corvette" "Ferrari" "Lamborghini" "Lotus" "Porsche" "Horse"
      都不符 → 印 "\nNo effect...\n"
      符合 →
          al = byte_3E0A3                  當前地點(1-based)
          if (al != 16h && al != 1Fh) → 印 "\nNo effect...\n"
          else →
              印 "\nPoof!\n"
              sub_2C598(0Ah, 0BB8h, 7D0h)  放音效
              槽 = sub_2B57C()             找一個空的物件槽
              sub_2B6C8(10h, 10h, X+1, Y, 樓層, 0, 槽)
```

`word_3DFB6` 是**金錢**(`docs/re/11` 的買馬那段:`sub word_3DFB6, ax ; 扣錢`)。

---

## 3. 那匹馬

最後一行看起來只是一堆數字,但 `docs/re/11` 已經把物件槽的欄位定過了,
而且巧的是**定它的那段程式碼就是買馬**:

```asm
; sub_118CC 尾段(買馬)      vs      ; sub_CD28 尾段(許願井)
mov  al, 10h  ; 馬的 tile           sub_2B6C8(10h, 10h, X+1, Y, 樓層, 0, 槽)
mov  [esi+1], al                             ↑    ↑    ↑    ↑    ↑
mov  [esi], al                              +0   +1   +2   +3   +4
```

`sub_2B6C8(a1..a6, 槽)` 就是把六個位元組寫進 `dword_3E46C[2*槽]`,
順序 = 種類 / 顯示 tile / X / Y / 樓層 / +5。所以那兩個 `10h` 是
**種類與顯示 tile 都設成馬** —— 與買馬**逐位元組相同**。

⇒ **許願井會變出一匹免費的馬**,就在玩家的**正東**一格(`X + 1`),
而且**不檢查那一格能不能站**。買馬那條路會挑四個方向找空位,許願井不挑。

---

## 4. 兩個地點:PAWS 與 EMPATH ABBEY

`byte_3E0A3` 是 1-based 的地點編號(`docs/re/03`),而 `locations.go` 裡:

| 值 | 地點 |
|---|---|
| 0x16 = 22 | **PAWS** |
| 0x1F = 31 | **EMPATH ABBEY** |

兩座都有井。這不是「井的清單」而是**彩蛋的白名單** ——
別處的井照樣可以投錢、照樣扣一枚、照樣問願望,但只會回你 `No effect...`。

---

## 5. 比對規則:大寫字面值的**前綴**

`sub_27C98(字面值, 玩家輸入)` 這支值得看,因為它決定了「要打什麼才算對」:

```asm
edi = 字面值
var_14 = -1                          ; 預設不符
清空 10 B 的區域 var_10
loop ebx = 0..8:
    if (字面值[ebx] == 0) { esi = ebx; break }     ; esi = 長度
    al = 字面值[ebx] & 7Fh
    if (byte_738D8[al] & 2)          ; ctype 表:是小寫嗎
        al = (字面值[ebx] & 7Fh) - 20h              ; → 大寫
    var_10[ebx] = al
sub_39554(var_10, 玩家輸入, esi)     ; strncmp,長度 = **字面值**的長度
if (回 0) var_14 = 0
return var_14                        ; 0 = 相符、−1 = 不符
```

三個後果:

| 輸入 | 結果 | 為什麼 |
|---|---|---|
| `HORSE` | ✅ | 剛好相等 |
| `HORSEY` | ✅ | 只比字面值的 5 個位元組 |
| `HORS` | ❌ | 第 5 個位元組是 NUL vs `E` |
| `horse` | ❌ | ★ 比的是**大寫**字面值,而輸入不轉換 |

★ 最後一條是刻意保留的:`sub_2B770` 收 32..122(含小寫),而比對只認大寫。
所以小寫的願望在原版**不會**生效。這種「看起來像 bug」的地方**照原樣實作**
—— `CLAUDE.md` §3.0:忠於原版,連 bug 一起還原。

⚠ 引擎的 `SubmitText` 會去頭尾空白,原版不會。那是通用輸入層的既有行為,
不在本次範圍內;`WellWishMatches` 這一層**不做 trim**,測試也照這樣釘。

---

## 6. 五輛跑車

```
Corvette   Ferrari   Lamborghini   Lotus   Porsche   Horse
```

前五個是 1988 年 Origin 團隊的夢想車單,第六個是給老實人的。
順序照原版的比對順序保留在 `WellWishes`。

---

## 7. 落地與驗收

| | |
|---|---|
| `u5data/well.go` | 六個字面值、兩個地點、12 字上限、一枚錢、前綴比對 |
| `game/well.go` | `lookAtWell()` 接 Y/N → 扣錢 → 收願望 → 判定 → 生馬 |
| `game/look.go` | tile 0xA1 從「印一句話」改成走 `lookAtWell()` |
| `game/strings.go` | 四句新訊息,並把那段錯誤結論改成更正記錄 |

八條測試,其中三條是專門擋回歸的:

- `TestWellIsInteractiveNotJustOneLine` —— 擋回到「只印一句話」
- `TestWellOnlyWorksAtPawsAndEmpathAbbey` —— 用 21/23 夾 22、30/32 夾 31,
  擋「把地點條件寫成範圍」
- `TestNoCoinNoWishAndNoMessage` —— 沒錢時**最後一句必須停在「是。」**,
  擋「順手補一句『你沒錢』」

⚠ 沒錢時的靜默看起來像卡住,但那就是原版。要改成友善提示得先問使用者,
不能自己決定(§3.0)。
