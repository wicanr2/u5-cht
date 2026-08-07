# 地圖物件層(`.OOL`)

> 輸入檔:`WORRIORS.EXP`(FM Towns)、`BRIT.OOL` / `UNDER.OOL` / `INIT.OOL` / `SAVED.OOL`
> 日期:2026-08-07

`.OOL` 在 `CLAUDE.md §2.1` 原本只標「疑為物件表」。啟動函式 `sub_0` 給了直接答案:

```asm
push    offset aBritDat         ; "BRIT.DAT"
call    sub_27A58               ; 等檔案就緒
push    0                       ; offset
push    100h                    ; 256 B
push    offset dword_3E46C      ; 目的地
call    sub_10910               ; → byte_3E0A5 == 0 ? "A:BRIT.OOL" : "A:UNDER.OOL"
push    eax
call    sub_2C740               ; (檔名, 緩衝, 大小, 位移)
```

256 B = **32 槽 × 8 B**。

---

## 1. 槽的欄位

由買馬那段(`sub_118CC` 尾段)逐行讀出來:

```asm
lea  esi, ds:3E46Ch[eax*8]
mov  [esi+5], al        ; al = 0
mov  [esi+7], al
mov  [esi+6], al
mov  al, 10h            ; 馬的 tile
mov  [esi+1], al
mov  [esi], al
mov  al, byte ptr [ebp+var_14]   ; X
mov  [esi+2], al
mov  al, byte ptr [ebp+var_18]   ; Y
mov  [esi+3], al
mov  al, byte_3E0A5              ; 樓層
mov  [esi+4], al
```

| 位移 | 內容 |
|---|---|
| +0 | 種類(基礎 tile);0 = 空槽 |
| +1 | 目前要畫的 tile |
| +2 | X |
| +3 | Y |
| +4 | 樓層 —— **有號數**,0xFF = −1(地下世界) |
| +5..+7 | 買馬時清零,全檔沒有其他讀寫處 → 用途不明,原樣保留 |

+0 與 +1 的差別在船的轉向函式 `sub_23FC` 看得出來:轉向只改 +1
(tile 0x2C/0x2E ⇄ 0x2D/0x2F),+0 不動。所以 **+0 是「這是什麼」、+1 是「此刻畫哪一格」**。

樓層是有號數這件事不是小節:當成無號讀會變 255,而 255 不等於任何真實樓層,
結果是地下世界的物件永遠畫不出來 —— 症狀看起來像「`UNDER.OOL` 是空的」。

---

## 2. 四個檔案的關係(逐位元組驗過)

| 檔案 | 大小 | 內容 |
|---|---|---|
| `BRIT.OOL` | 256 | 地表的預設物件。只有槽 0:tile 28(步行的隊伍) |
| `UNDER.OOL` | 256 | 地下世界的預設物件。五個,樓層欄都是 0xFF |
| `INIT.OOL` | 256 | **與 `UNDER.OOL` 逐位元組相同** |
| `SAVED.OOL` | 512 | **前 256 B 地表 + 後 256 B 地下**;前半全零,後半 == `UNDER.OOL` |

⇒ 存檔的物件表是「地表一份 + 地下一份」接在一起。512 = 2 × 256 這個算術
本身說明不了什麼,但「後半與 `UNDER.OOL` 完全相同」就是硬證據。

⚠ `BRIT.OOL` 槽 0 的座標是 (86,107) —— 那是地點表第 17 筆(一個無名地點)的位置,
**不是** `INIT.GAM` 的開場位置(IOLO'S HUT,(45,62))。兩份檔案不同步:
`BRIT.OOL` 是世界的預設,開新遊戲該讀的是 `INIT.OOL`。

一開始我把「槽 0 = 開場位置」寫進測試,結果測試炸了 —— 炸得好。

---

## 3. 場景裡的物件表

進場景時 `sub_1678` 把槽 1..31 的種類碼**全部歸零**,再載入場景:

```asm
loc_17A5:
    cmp  esi, 20h
    jge  short loc_17B5
    mov  byte ptr dword_3E46C[esi*8], 0
    inc  esi
    jmp  short loc_17A5
loc_17B5:
    ...
    call sub_5C8        ; 載入場景地圖
    call sub_48C
```

⇒ **在城裡買的馬,離開城鎮就不見了。** 那是原版行為,不是漏做。

---

## 4. 買馬:找位置的規則

`sub_118CC` 一開始就先找位置,找不到連買賣都不會開始(直接「馬廄關門了」):

```
依序看四個鄰格:南 北 東 西
  (dword_555E8 = {0, 0, 1, -1}、dword_555F8 = {1, -1, 0, 0})
條件:sub_2B360(x, y, floor) 回 0(沒東西擋著)
      而且地形 tile ∈ {5, 68, 69}
```

68 / 69 是城鎮的地面磚,5 是草地。

---

## 5. 買船**不用物件槽**

這一條是本輪最值得記的更正。

我原本照「買馬會生成物件」的模式,假設買船也生成一個船的物件,還順手寫了
`TileFrigate = 0x18` / `TileSkiff = 0x16` 兩個常數。**兩個數字都是我編的。**

實際讀 `sub_218DC`(造船廠成交):

```asm
movsx eax, word_3EF34
mov   dl, byte_57080[eax]     ; 停泊 X
mov   byte_3E165, dl
mov   al, byte_57084[eax]     ; 停泊 Y
mov   byte_3E166, al
...
mov   ax, word_3EF38
sub   word_3DFB6, ax          ; 扣錢
```

整段**沒有出現任何 tile 值,也沒有碰 `dword_3E46C`**。船的存在是靠
`byte_3EE17` 的旗標(帆船 0x82、小艇 0x40)加上兩個座標 —— 船停在碼頭等你,
不是憑空出現在腳邊。

停泊座標在 DOS `DATA.OVL`:X 表 `0x4D86`、Y 表 `0x4D8A`,各 4 byte。

船的四個朝向 tile 0x2C..0x2F 確實存在(`sub_23FC` 的轉向),但那是
「船在海上時的顯示」,與「買到的是哪一種船」還沒對上 —— 沒有證據就不填。

---

## 6. 種類碼裡認得出來的幾個

| 值 | 意義 | 依據 |
|---|---|---|
| 0 | 空槽 | `sub_1678` 清表寫 0;`sub_118CC` 用 `!= 0` 找空槽 |
| 0x10 | 馬 | `sub_118CC` 的 `mov al, 10h` |
| 0x1C | 步行的隊伍 | `BRIT.OOL` 槽 0 的實際值,與存檔的載具欄位同值 |
| ≥ 0x40 | 怪物 | `cmp byte ptr dword_3E46C[eax*8], 40h`,與生物名表的 `CreatureBase` 同源 |
| 0xFC | 某種特殊物件 | `sub_48C` 開場掃全表看在不在場;是什麼還沒追到 |

---

## 7. 還沒做的

- **Board / Exit**(上下坐騎與船):`sub_2ACF4` 的 case 66 是 `B`,但實際處理
  還沒追。買了馬能看到馬,但騎不上去。
- **怪物生成與移動**:物件表放得下怪物(種類碼 ≥ 0x40),但生成規則與
  每回合的移動(`sub_9690`)是戰鬥系統的一部分,另案。
- **0xFC 是什麼**。
- **+5..+7 三個位元組的用途**。
