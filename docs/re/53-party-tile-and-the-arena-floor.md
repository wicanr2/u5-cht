# 隊伍在戰場上畫什麼 —— 以及戰場的底色不是空白

輸入檔:`WORRIORS.EXP`(FM Towns 英文版)
IDA 位址:`sub_2EDF8`(躺下)、`sub_2ED50`(起身)、`sub_16DA4`(在步行嗎)、
`sub_FE48` 的第一個迴圈(底色)、`sub_16058`(戰鬥中的 Klimb)
日期:2026-08-08

---

## 0. 上一輪記下的那個矛盾

`docs/re/52` §5:

> `sub_16058` 判「爬得過去」用的是 tile 0x4C,而本專案 `partyTileFor()`
> 回的也是 0x4C。兩者衝突,而 `partyTileFor` 當初是猜的。⬜ 待查。
> **先不動它** —— 沒有證據就不要為了消除矛盾而改另一邊。

「先不動」是對的做法,但那是**暫時**的做法。矛盾本身就是一條線索:
一個 tile 不會同時是「隊伍自己」與「戰場上爬得過去的東西」,所以其中一邊錯了,
而我知道哪一邊沒有依據。

## 1. 證據:兩支成對的函式直接寫那個 tile

```asm
sub_2EDF8  (躺下)
    mov  byte_3DDBF[eax], 53h              ; 狀態 → 'S'(睡著)
    or   byte ptr [esi+2], 8               ; 單位旗標:躺著
    movzx eax, byte ptr [esi+4]            ; 該單位對應的物件槽
    mov  byte ptr dword_3E46C+1[eax*8], 1Eh   ; ★ 圖 → 0x1E

sub_2ED50  (起身)
    mov  byte_3DDBF[eax], 47h              ; 狀態 → 'G'
    test byte ptr [edi+2], 10h
    movzx eax, byte ptr [edi+4]
    mov  byte ptr dword_3E46C+1[eax*8], 1Dh   ; ★ 圖 → 0x1D
```

⇒ **站著 0x1D、躺著 0x1E。** 兩支成對、同一個欄位、同一種寫法。

### 為什麼是「物件記錄的 byte +1」

戰場單位記錄在 `dword_3EF50 + i*8`:`+2` 旗標(bit 0x80 = 隊員)、`+3` 名冊索引、
`+4` **對應的物件槽**、`+6/+7` 戰場座標。畫圖用的是物件記錄
(`dword_3E46C + 槽*8`)的 `+0/+1` —— 所以「換圖」是去改那個物件。

### 第二條獨立佐證

`sub_16DA4`(「在步行嗎」)收的是 **0x1C 或 0x1D 兩個值**:

```asm
mov al, byte_3E08C
cmp al, 1Ch ; jz ok
cmp al, 1Dh ; jz ok
push offset aOnFoot_0   ; "\nOn foot\n"
```

0x1C 是世界地圖上步行的隊伍(`BRIT.OOL` 槽 0 的實際內容,早就驗過)。
0x1D 緊接在後,而且同樣算「在步行」⇒ 兩者同一族。0x4C 與這一族毫無關係。

## 2. 而且它與職業無關

我原本的註解寫「原版用角色的職業決定 sprite」——那是把「城鎮 NPC 依生物編號
選圖」的規則套過來的。原版在戰場上只有**兩個值**:站著與躺著。

另一個連帶的錯:引擎原本寫 `NPCTileBase + partyTileFor(ch)`(= 256 + 0x4C)。
0x1D / 0x1E 是**前 256 格**(地形與物件那一頁)的直接 tile 號,不是生物編號 ——
加了 256 會畫成某隻怪物。

## 3. 順手抓到的第二個錯:戰場的底色

追 `dword_3EF50` 的時候讀到 `sub_FE48` 的開頭:

```asm
loc_FE53:
    cmp   edi, 0Bh
    jge   short loc_FE83
    lea   esi, ds:3F8F4h[edi*32]
    movzx eax, byte_418DD              ; ★ 地板
    mov   ah, al ; push ax ; shl eax, 10h ; pop ax   ; 複製成四份
    mov   edi, esi
    stosd ; stosd ; stosw ; stosb      ; 11 B
```

⇒ **戰場先被填滿地板**(11 列 × 11 格),`sub_FD54` 才在上面畫牆框。

我第一版把整塊填成 0xFF,理由是別處有個 `rep stosd` 填 0xFFFFFFFF ——
但那是**畫面緩衝**的初始化,不在這條路徑上。**跑錯了來源。**

症狀會是「戰場除了牆之外全是空白」—— 而那在沒有畫面的測試裡看不出來,
因為我當時的測試只檢查牆與開口的位置。現在加了一條:框內 7×7 不得為空白。

## 4. 這一輪的方法論結論

> **矛盾是線索,不是雜訊。** 上一輪把它記下來而不硬修是對的;
> 但記下來之後要回頭追,而追的方向由「哪一邊有證據」決定 ——
> 0x4C 那邊沒有,所以動的是那一邊。

而追的過程順帶抓到第二個錯(底色),因為讀的是**同一支函式**。
這是「回頭追矛盾」比「各自猜」划算的地方。

## 5. 引擎對應

| 原版 | 引擎 |
|---|---|
| `sub_2ED50` 的 `1Dh` | `game.PartyTileStanding` |
| `sub_2EDF8` 的 `1Eh` | `game.PartyTileLying` |
| 兩者的選擇 | `game.partyTileFor`(依狀態,不依職業) |
| `sub_FE48` 的底色迴圈 | `u5data.BuildDungeonArena` 的地板填滿 |
