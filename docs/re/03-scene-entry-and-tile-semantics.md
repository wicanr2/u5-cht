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

`sub_3181C(組別)` 的參數是場景組別:**6 = castle、7 = Lord British 城堡、
8 = village/hut、11 = Blackthorn 宮殿**。組別應該對應到四個場景檔
(`TOWNE/CASTLE/KEEP/DWELLING.DAT`)的某種分派,但**對應關係還沒確認**。

## 3. 地點表:還沒找到,但知道去哪找

`sub_3181C` 內用 `sub_31CB8()` 取場景索引(且檢查 `<= 0xE`,即 0–14)。
而 §1 已證明 `sub_31CB8` 其實是回傳全域 **`dword_65334`**。

⇒ **下一步:查 `dword_65334` 的寫入 xref**(`tools/ida_xref.idc`),
找到「從玩家世界座標算出場景索引」的地方 —— 那裡就是地點表。

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
