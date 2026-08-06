# RE-03:進入場景的分派器 + tile 語意表(部分)

> 日期:2026-08-07 ・ 來源:FM Towns `WORRIORS.EXP` 反編譯輸出 + `.asm`

## 1. ⚠ 先講一個會污染整份反編譯輸出的坑

追地點表時,`sub_31CB8()` 反編譯出來是:

```c
int sub_31CB8() { return 0; }
```

看起來是空函式。但回去讀組語:

```asm
sub_31CB8   proc near
            cmp     dword_65334, 0FFFFFFFFh
            jnz     short loc_31CC5
            sub     eax, eax              ; return 0
            jmp     short locret_31CCA
loc_31CC5:  mov     eax, dword_65334      ; return dword_65334
locret_31CCA: retn
```

實際語意是 **`return dword_65334 == -1 ? 0 : dword_65334;`** —— 它回傳當前場景索引,
不是常數 0。

### 根因與影響範圍

反編譯輸出裡到處都有這行警告:

```
// write access to const memory has been detected, the output may be wrong!
```

Hex-Rays 把某些**可寫的全域**當成唯讀常數,於是常數傳播把整段邏輯摺掉。
`WORRIORS_hexrays.c` 裡帶這個警告的函式**不只一個**。

> **紀律**:反編譯出來是「常數回傳」「條件恆真/恆假」「整段被摺掉」時,
> **一律回去讀那個函式的 `.asm` 再下結論**。Hex-Rays 是加速器,不是真值來源;
> 真值是組語與 xref 圖。(同 `rulebook/62`:一手資料贏二手推論。)

## 2. `sub_2D72C`:進入場景的分派器

依**玩家腳下的 tile** 分派 —— 所以它同時是一份 tile 語意表:

| tile | 地點 | 後續 |
|---|---|---|
| `0x10`(16) | hut(小屋) | `sub_3181C(8)` |
| `0x11`(17) | **the Shrine of the Codex**(法典聖壇) | `sub_1DA10()` |
| `0x12`(18) | keep(要塞) | |
| `0x13`(19) | village(村莊) | `sub_3181C(8)` |
| `0x14`(20) | towne(城鎮) | |
| `0x15`(21) | castle(城堡) | `sub_3181C(6)` |
| `0x16`(22) | cave(洞穴) | `sub_2D564` |
| `0x17`(23) | mine(礦坑) | `sub_2D564` |
| `0x18`(24) | dungeon(地牢) | `sub_2D564` |
| `0x19`(25) | **the shrine of …**(八德聖壇;名字取自 `off_411BC[9]`,狀態看 `byte_411FC[8]`/`byte_41204[8]`) | `sub_1DA10()` |
| `0x3D`(61) | **the palace of Blackthorn!** | `sub_3181C(11)` |
| `0x3E`(62) | **the Castle of Lord British!** | `sub_3181C(7)` |
| 其他 | `"What?\n"` | 不能進 |

這批 tile 語意與 `docs/re/02` 的通行判定互相印證:`sub_2A610` 裡的 `(mover & 0xFE) != 0x1C`、
`sub_86C` 裡的 `(v2 & 0xFE) == 0x10`(tile 16–17)都落在這個範圍。

### ⚠ 更正:`sub_3181C` 是播背景音樂,不是載入場景

一開始把 `sub_3181C(6/7/8/11)` 讀成「載入場景(組別)」。**錯的。** 它的組語開頭是:

```asm
cmp     dword_5FFF4, 1              ; debug 旗標
jnz     short loc_3185B
push    [ebp+arg_0]
push    offset aBgmSongD            ; "BGM SONG %d\n"      ← 決定性線索
...
loc_3185B:
call    sub_31CB8
mov     dword_65338, eax            ; 舊曲目
cmp     [ebp+arg_0], 0FFFFFFFFh
...
mov     eax, [ebp+arg_0]
mov     dword_65334, eax            ; 新曲目
```

⇒ **`sub_3181C(n)` = 播第 n 首 BGM**;`dword_65334` = 當前曲目、`dword_65338` = 前一首。
函式裡那個 `<= 0xE` 是**曲目數上限(0–14,共 15 首)** ——
而 FM Towns 版剛好有 **15 首 `.EUP`**(`M1`–`M152.EUP`),數字對得上。

**為什麼會讀錯**:`"BGM SONG %d"` 那段被 `dword_5FFF4 == 1` 的 debug 分支包著,
**在 Hex-Rays 的輸出裡看不到**;而反編譯版本的開頭又被 §1 那個 const 誤判污染,
於是整段看起來就像「取場景索引 → 載入」。**又是讀組語才看見真相。**

### 意外收穫:地點類型 → BGM 編號對應表

這個誤讀反而挖出 P6 音樂要用的東西 —— 進入不同地點時該播哪首曲子:

| 地點 | BGM 編號 |
|---|---|
| castle | 6 |
| **Lord British 城堡** | 7 |
| village / hut | 8 |
| **Blackthorn 宮殿** | 11 |

(其餘地點的編號要把 `sub_2D72C` 的每個 case 讀完才齊。)

## 3. 地點表:仍未找到(前一條線索是誤判)

原本以為 `sub_3181C` → `sub_31CB8` → `dword_65334` 這條鏈通往地點表。
§2 的更正推翻了它:那整條鏈是**背景音樂**的曲目狀態,與場景載入無關。

(以下的 xref 追蹤仍然有效,只是結論改成「這是 BGM 曲目狀態」)

查 `dword_65334` 的寫入 xref(`tools/ida_xref.idc`)得到 10 筆:

```
寫  0xC790   sub_C778    mov dword_65334, 1
寫  0xCC27   sub_C778    mov dword_65334, 7
寫  0x3197F  sub_3181C   mov dword_65334, eax     ← 主要路徑
讀  0x31CB8  sub_31CB8   cmp dword_65334, 0FFFFFFFFh
…
```

`0x3197F` 的組語:

```asm
loc_3197C:
        mov     eax, [ebp+arg_0]
        mov     dword_65334, eax
```

⇒ 參數直接寫進 `dword_65334`。結合 §2 的更正:**那是曲目編號**,
所以 hut 與 village 都傳 8 只是因為它們共用同一首背景音樂,不代表共用場景。

### 真正該找的地方(下一步)

`sub_2D72C` 的每個 case 除了播音樂,還會呼叫別的東西 ——
`sub_1DA10`(聖壇)、`sub_2D564`(cave/mine/dungeon)、`sub_10928`(印地名並確認)。
**載入場景的動作在那些函式裡,不在 `sub_3181C`。**

下一步:讀 `sub_10928`(它拿地名字串當參數,很可能就是「確認進入 → 載入」的主流程)。

⚠ 在找到它之前,**不要用「城鎮順序大概是這樣」去對應場景** ——
那會讓玩家走進 Britain 卻進到 Yew。

## 4. 順帶發現:文字模板的佔位符

```
"Good @, and welcome to #!"
```

`@` 與 `#` 是**執行期替換的佔位符**(推測 `@` = 時段 morning/afternoon/evening、
`#` = 地名)。中文化時這些佔位符必須保留,而且**中文語序可能要調換位置** ——
「早安,歡迎來到 #!」這種。列入 P5 翻譯流程的注意事項。

其他確認到的訊息:`"Entering room...\n"`、`"Thou art under arrest!"`(Blackthorn 暴政下
守衛會逮捕玩家)、`"Thou art subdued and blindfolded!"`。

## 5. 地點表 ✅ 已找到並實作

`sub_10928(地名字串)` 裡有一段線性搜尋:

```c
for (i = 0; i < 32 && (byte_410F4[i] != 0 || byte_4111C[i] != 0); ++i);
```

那兩個 `!= 0` **又是 §1 的 const 誤判**(玩家 X/Y 座標被當成常數 0)——
實際語意是「掃 32 筆,找座標與玩家相符的那一筆」。於是三張平行表現形:

| 符號 | 內容 |
|---|---|
| `off_41054[32]` | 地點名稱指標 |
| `byte_410F4[32]` | 地點 X 座標 |
| `byte_4111C[32]` | 地點 Y 座標 |

dump 出來就是完整的 32 個地點(實作在 `internal/u5data/locations.go`):

```
 0 MOONGLOW(232,135)   1 BRITAIN(81,106)      2 JHELOM(36,222)     3 YEW(58,43)
 4 MINOC(159,20)       5 TRINSIC(106,184)     6 SKARA BRAE(22,128) 7 NEW MAGINCIA(187,169)
 8 FOGSBANE(88,120)    9 STORMCROW(152,24)   10 GREYHAVEN(104,216) 11 WAVEGUIDE(216,120)
12 IOLO'S HUT(45,62)  13–17 (無名)           18 WEST BRITANNY      19 NORTH BRITANNY
20 EAST BRITANNY      21 PAWS(98,145)        22 COVE(136,90)       23 BUCCANEER'S DEN
24 ARARAT(49,58)      25 BORDERMARCH(15,160) 26 FARTHING(64,240)   27 WINDEMERE(248,8)
28 STONEGATE(148,74)  29 THE LYCAEUM(218,107) 30 EMPATH ABBEY(28,50) 31 SERPENT'S HOLD(146,241)
```

### 交叉驗證:26/32 的座標落在已知入口 tile 上

把每個座標拿去查已組好的世界地圖:

- 七大城市 → **tile 20 (towne)** ✓
- New Magincia / 三個 Britanny / Paws / Cove → **19 (village)** ✓
- Iolo's Hut 與三個無名點 → **16 (hut)** ✓
- Bordermarch / Farthing / Windemere / Stonegate → **18 (keep)** ✓
- Lycaeum / Empath Abbey / Serpent's Hold → **21 (castle)** ✓
- Buccaneer's Den → **20 (towne)** ✓

**不命中的 6 筆反而更有價值**:

1. **四座燈塔**(Fogsbane / Stormcrow / Greyhaven / Waveguide)全落在 **tile 27**
   ⇒ **tile 27 = lighthouse** —— 正好補上 `DATA.OVL` 0x2AB3 地點類型表裡一直對不上的 `lighthouse`。
2. **#16 (86,107) = tile 62 = Lord British 城堡**。它無名,是因為落在
   `if (i < 13 || i > 17)` 那個「不印名字」的範圍 —— 程式另外處理它。**與程式碼完全吻合。**
3. `ARARAT (49,58)` 落在 tile 8(一般地形)、`#17 (196,245)` 落在 tile 57 —— 待查。

這種「大部分命中、不命中的能解釋成新事實」的形狀,比 100% 命中更可信 ——
若三張表有錯位,不會出現這麼整齊的類型對應。

### 中文地名的處理

`Location` 同時存 `Name`(英文 canonical)與 `NameZH`。
**只填已定案的八德城市**(對齊 `CONTEXT.md` glossary 與聖者之書體系),
其餘留空、顯示時退回英文 —— 憑印象硬翻會變成二手轉譯,等《軟體世界》手冊 OCR 定案。

⚠ 玩家輸入比對一律用 `Name`(英文):玩家在遊戲中打不出中文(u4-cht 踩過的坑)。
