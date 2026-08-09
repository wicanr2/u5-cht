# 99 — 維生開銷的尾巴:再生戒指、力場消散、回合計數器,以及會自己走的馬

| | |
|---|---|
| 輸入檔 | `re_work/fmtowns/WORRIORS.EXP.asm` / `.i64` |
| SHA-256 | 見 `re_work/fmtowns/SHA256SUMS.txt` |
| 主要函式 | ★ `sub_2BCC8`(再生戒指)、`sub_2ECE8`(戒指的每回合效果)、`sub_1F4E4`(力場消散)、`sub_2A50C` 尾段、`sub_16370`、★ `sub_FEC` + `sub_FB4`(馬)、`sub_1B854` 尾段(施捨業報) |
| 主要資料 | `byte_3DDD1` = `CharRing`(記錄 0x1D)、`byte_3E09B`(回合計數器)、`byte_3E09E`(模式倒數)、`byte_3E0AE`(當前行動單位) |
| 工具 | ★ `tools/refunc.py`(新)—— 一支函式的組語(`proc`→`endp`,**不截斷**)/ 反編譯的 C / 呼叫者 |
| 起因 | 「完成遊戲邏輯」—— 從 `WORKLIST §5.2c` 的 25 支未讀函式往下清 |
| 狀態 | ✅ 八個機制落地;六條舊斷言更正 |

---

## 0. ★★ 這一輪最貴的教訓:`call` 在 `retn` 前面一行

`sub_2A50C`(每回合的維生開銷)在 `docs/re/70` 已經逐行讀過、落地過、
還寫了「三個容易寫錯的地方」。而它的**最後一行**是:

```asm
loc_2A606:
        call    sub_2BCC8       ; ★ 再生戒指 —— 整支機制掛在這裡
        pop     edi
        pop     esi
        pop     ebx
        leave
        retn
```

`docs/re/85` 已經記過同一個形狀(怪物移動在 `sub_2D38` 的最後一行)。
**同一個坑踩第二次。** ⇒ 規則寫進 `CLAUDE.md §4.5`:讀函式一律讀到 `endp`,
「看起來收尾了」不是停下來的理由 —— 原版把新東西掛在 `retn` 前面一行。

## 1. `sub_2BCC8` —— Hex-Rays 說 `return 0;`,組語有 55 行

```c
int sub_2BCC8() { return 0; }        // ← 反編譯輸出
```

```asm
for (i = 0; i < byte_3E06B; i++) {                 ; 隊伍人數
    ebx = i*32
    if (byte_3DDBF[ebx] == 'D')      continue      ; ★ 只跳過「死」
    if (byte_3DDD1[ebx] != 2Ch)      continue      ; 戒指 != 再生戒指
    if (random(0,7) != 7)            continue      ; 1/8
    var_4 = word_3DDC4[ebx]                        ; 目前 HP
    sub_2BBDC(&var_4, 1, word_3DDC6[ebx])          ; +1,上限 MaxHP
    word_3DDC4[ebx] = var_4
    byte_41989 = 1                                 ; 重畫狀態列
}
```

`byte_3DDD1 − 名冊基底 byte_3DDB4 = 0x1D` = `CharRing`,而 `0x2C = 44` 正是
`ItemRingRegeneration`(名字早就從 `DATA.OVL` 0x1806 解出來了)。

★ 三個容易寫錯的地方:

1. **只跳過 'D'。** 睡著、被惑、中毒的人**照樣回血** —— 中毒的人同一回合
   還會被維生開銷扣 1 點,兩件事各自發生。
2. **它忽略自己的參數。** `sub_2ECE8` 推了一個名冊索引進去,而 IDA 給
   `sub_2BCC8` 的堆疊框裡**連 `arg_0` 都沒有** —— 它一律掃整隊。
   ⇒ 兩個人都戴再生戒指時,每人每回合被擲**兩次**骰。
3. 上限是**那個人的** MaxHP。

⇒ 引擎此前只做了戒指的**消失**(`docs/re/70`),沒有做它的**效果**。

## 2. `sub_2ECE8` —— 差點把參數寫成 0

```c
char sub_16370() { sub_2ECE8(0); sub_1F4E4(); return 0; }   // ← 反編譯輸出
```

```asm
sub_16370:
        movzx   eax, byte_3E0AE     ; ★ 不是 0,是「當前行動的單位」
        push    eax
        call    sub_2ECE8
        call    sub_1F4E4
        mov     al, byte_3E09E      ; ★ 模式倒數也在這裡(反編譯整段不見)
        ...
```

**兩處都被摺掉了**:參數被常數傳播成 `0`,而後半段的模式倒數整段消失。
照反編譯寫的話會得到「只有第一槽的戒指有用」—— 一條完全捏造出來的規則。

`sub_2ECE8(單位)` 本體:

```
edi = &dword_3EF50[單位*8]
if (!([edi+2] & 0x80))          return    ; 不是隊員
if ([edi+2] & 0x28)             return    ; ★ 睡著(0x08)或死了(0x20)
名冊 = [edi+3]
if (戒指 == 2Ah) { dword_3E46C[[edi+4]*8+1] = 1Dh; [edi+2] |= 10h; return }
if (戒指 == 2Ch) sub_2BCC8(名冊)
```

`0x80` / `0x28` / `0x10` 與引擎的 `UnitParty` / `UnitAsleep|UnitDead` / `UnitHidden`
逐位元對得上。隱形那條做的事與 `hideCaster`(Sanct Lor)相同 ⇒
**戴著隱形戒指的人每回合會重新隱形**,被清掉也會回來。

## 3. `sub_1F4E4` —— 力場每回合 1/16 消散

```c
for (i = 0; i < 32; ++i)
    if ((dword_3E46C[2*i] & 0xFC) == 0xE8 && sub_28E14(0, 255) < 16)
        sub_B210(i + 1);
```

★ 種類碼 0xE8..0xEB 是四種力場 —— 認出來的方法是 **look 表的物件段**
(索引 `256 + 種類碼`):`look#488` 毒力場、`489` 睡眠力場、
`490`(原版寫成 "a field of field")、`491` 力場。

⚠ **沒有訊息。** 力場就這樣不見了 —— 「力場消散了!」那句是舉權杖
(`sub_1A5E8` case 20)才印的,不要搬過來。

## 4. ★★ `sub_FEC` + `sub_FB4` —— 沒被繫住的馬會自己走

掛在**場景主迴圈** `sub_1A54` 上,與「打開的門四回合後自己關上」是鄰居 ——
而引擎只做了門。

```
for (槽 = 0; 槽 < 32; 槽++) {
    if ((kind & 0FEh) != 10h)        continue   ; 不是馬
    if (槽.樓層 != byte_3E0A5)        continue   ; 不在當前樓層
    if (random(0,1) != 0)            continue   ; ★ 1/2 才考慮動
    if (sub_FB4(x, y+1) || sub_FB4(x+1, y) ||
        sub_FB4(x, y−1) || sub_FB4(x−1, y))  continue   ; ★ 被繫住
    if (random(0,1) != 0) {
        d = 2*random(0,1) − 1
        x += d;  kind = (d > 0) ? 10h : 11h     ; ★ 朝向跟著改
    } else {
        y += 2*random(0,1) − 1                  ; ★ 垂直**不改朝向**
    }
    if (x > 1Fh || y > 1Fh || x < 0 || y < 0)  continue   ; ★ 不環繞
    if (!sub_2A694(10h, tile(x,y)))   continue   ; 馬走不上那個地形
    if (sub_2B3DC(x, y, 樓層))        continue   ; 那格有東西
    寫回 kind / tile / x / y
}

sub_FB4(x, y) = (tile(x,y) == 0A2h || tile(x,y) == 43h)
```

★ 三個名字都是從 look 表查出來的,不是猜的:

| | look 索引 | 內容 |
|---|---|---|
| 種類碼 0x10 / 0x11 | `look#272` / `look#273`(物件段 = 256 + 種類碼)| **a horse**(兩個朝向)|
| tile 0xA2 | `look#162`(地形段)| **a hitching post** 繫馬柱 |
| tile 0x43 | `look#67` | **a rail** 欄杆 |

⚠ **別拿地形段去查物件種類碼**:tile 0x10 在地形段是「a small hut」——
用錯段會得到「小屋會自己走」。索引 < 256 才是地形,≥ 256 是物件與生物。

★ 四個容易寫錯的地方:

1. **繫住的判準是「四個鄰格」不是「腳下」。** 馬站在柱子**旁邊**就算被繫住,
   所以柱子那一格自己是空的。寫成看腳下 → 城裡的馬會全部跑掉。
2. **垂直移動不改朝向**(`var_C` 還是迴圈開頭讀進來的原值)。
3. **座標不環繞**(`cmp edi, 1Fh; jg`)。用 `WrapWorld` 的話邊上的馬會瞬移到對面。
4. **1/2 的閘門在最前面**,在讀座標之前 —— 行為上看不出差別,但隨機序列不同。

## 5. `sub_2A50C` 的尾段 —— 三件事,而其中一件推翻了舊斷言

```
loc_2A563:
    if (byte_3E08F == byte_3E090)  goto loc_2A5CA         ; 這小時已結算
    if (word_3DFB4 == 0) { 印 "Starving!"; sub_2A4D0(); jmp loc_2A5C0 }
    if (byte_3E08F ∈ {6, 0Ch, 12h}) 存糧 −= 活人數
loc_2A5C0:
    byte_3E090 = byte_3E08F                               ; ★★ 三條路都設
loc_2A5CA:
    sub_2BBB8(&byte_3E09B, 1, 0FFh)                       ; 回合計數器 +1
    if (byte_3E09E != 0 && != 0FFh) {                     ; 模式倒數
        if (--byte_3E09E == 0) { byte_3E08A = 0; sub_2A1E8(); }
    }
    sub_2BCC8()                                           ; 再生戒指
```

### ★★ (a) 挨餓是**每小時**一次,不是每回合

挨餓那條路的結尾是 `jmp short loc_2A5C0`,而 `loc_2A5C0` 就是
「記下這個小時處理過了」。⇒ 斷糧的隊伍**每小時**掉一次血。

引擎此前每走一步掉一次(而且 `docs/re/70` 與 `upkeep.go` 都明文斷言
「每回合都判…斷糧是會死人的」)—— 那讓斷糧在幾十步內滅團,不是原版的難度。
錯因:只看了「用餐那條路才設 `byte_3E090`」,沒注意另外兩條路也 `jmp` 到同一點。

### ★★ (b) `byte_3E09B` 是**回合計數器**,不是進餐計數器

它的 `+1` 在 `loc_2A5CA`,而「這小時已結算」那條路**直接跳到它** ⇒
與吃不吃飯無關,每回合都 +1(上限 255)。

⇒ 這解掉了一條懸案:`sub_1B854`(對話裡的「給我錢」)扣完錢之後那一段

```
物件槽 = word_3E77C[dword_55E6C*16]
if ((dword_3E46C[槽*8] & 0FCh) != 6Ch)  return       ; ★ 只有乞丐
if (byte_3E09B < 64h)                   return       ; ★ 距上次不到 100 回合
byte_3E09B = 0
sub_2BBB8(&byte_3E098, 1, 63h)                       ; 業報 +1 上限 99
if (word_3DFB6 == 0) sub_2BBB8(&byte_3E098, 2, 63h)  ; ★ 給光錢再 +2
```

`docs/re/79` 當時寫「閘門語意未定,業報那一段先不做 —— 缺一段比多一段錯的好」。
語意定了,所以接上了。種類碼 0x6C 是乞丐(`look#364`,四個朝向共用一句)。

⚠ **兩次各自夾在 99**,不是「+3 一次算完」。

### (c) 模式倒數在**兩個**迴圈裡各一份

`sub_16370`(戰鬥中玩家回合結束)與 `sub_2A50C`(戰鬥外每回合)都遞減
`byte_3E09E`,歸零時把 `byte_3E08A` 清成 0。而兩個迴圈**互斥**
(`docs/re/81`),所以實際上每回合只減一次。⚠ `0xFF` 是「不倒數」的哨兵。

⚠ 引擎的 `tickCombatMode` 會印「咒語的效力消退了。」—— **那是引擎加的**,
原版只設 `byte_41989`(重畫狀態列)。保留,但別當成原版文字。

## 5b. `sub_2BCC8` 的另外兩個呼叫端,與「被惑的人睡一覺就好了」

`sub_2BCC8` 有四個呼叫端,而尾段那一個只是其中之一:

| 呼叫端 | 位置 | 節奏 |
|---|---|---|
| `sub_2A50C` | 最後一行 | 每回合(戰鬥外)|
| `sub_16370` | 經 `sub_2ECE8` | 每個玩家單位回合結束(戰鬥中)|
| `sub_165C8` 紮營 | `loc_167E9`,突襲骰前面一行 | **每睡一小時** |
| `sub_21D48` 旅店 | `loc_21F0C`,`sub_29304(9)` 前面 | **每推 9 分鐘** |

⇒ 戴著再生戒指睡一夜會慢慢回血。旅店本來就補滿 HP 所以看不出差別,
但中毒的人在旅店**不會被補滿**(會死),那條路上戒指仍然有作用。

### ★★ `sub_2EDC0` 的 `!= 'P'` 守衛解釋了一條玩家用得到的規則

紮營時原版對「守夜的人以外」每一位呼叫 `sub_2EDC0(單位)`:

```
if (!(flags & 80h) || 狀態 != 'P')  sub_2EDF8(單位)     ; 讓他睡著
sub_2EDF8: if (狀態 == 'D') return;  狀態 = 'S';  tile = 1Eh;  flags |= 8
```

判準是「狀態**不是** 'P'」而不是「狀態是 'G'」。原因是機制性的:
**狀態是單一位元組**,設成 `'S'` 會把原本的值擦掉 ——

1. **中毒的人不睡。** 否則睡一覺就解毒了,而原版繞過這件事。
2. **被惑('C')的人會睡,而那把魅惑解掉。** 同一個副作用,這次原版**沒有**繞過。
   ⇒ **紮營是解魅惑的方法之一。**

引擎此前寫成「只有 'G' 會睡」,第 2 條整個消失。

## 5c. ★★ `sub_BBA0` —— 戰場上站在有害的格子上

由 `sub_A9EC` 在**每個單位動完之後**呼叫(`sub_A108` 怪物 AI 與 `sub_A360`
玩家指令**兩條路都會**)。這是**戰鬥地圖版**的 `terrainEffects` ——
引擎此前只有場景 / 大地圖那一份,戰場上熔岩、壁爐、沼澤與力場一件都不作用,
**把敵人推到熔岩上完全沒事**。

```
esi = 0
tile = byte_3F8F4[y*32 + x]                       ; 這個單位腳下(sub_DB10 的戰場分支)
if (tile == 8Fh || tile == 0BCh) esi = 100        ; 熔岩 / 壁爐
if (tile == 4)                   esi = 50         ; 沼澤
if (esi == 0)                                     ; ★ 地形沒中才掃物件
    for (i = 0; i < 32; i++)
        if (i != 自己的物件槽 && 物件[i].x == 我的 x && 物件[i].y == 我的 y) {
            if (kind == 0EAh) esi = 100            ; 火力場
            if (kind == 0E8h) esi = 50             ; 毒力場
            if (kind == 0E9h) esi = 150            ; 睡眠力場
            if (esi) break
        }
switch (esi) {
  case  50: if (物件[自己].kind < 80h) sub_B8DC(自己, −1)     ; ★ 高階怪物免疫
  case 100: sub_B51C(自己, random(0,10));  sub_1F840(自己, 255)
  case 150: sub_2EDF8(自己)                                   ; 睡著
}
```

三個 tile 都是從 look 表查的:`look#143` **molten lava**、
`look#188` **a fireplace**、`look#4` **swamp**。

⚠ **第四種力場(0xEB)沒有 case** —— 站上去什麼都不會發生。別「補齊」。

### ★★ `sub_B8DC` —— 毒上不了身就改成扣血

```
if (是隊員 && 狀態 == 'G') { 狀態 = 'P'; 印 "<名字> is poisoned!" }
else                        sub_B51C(自己, random(0, 20))
```

⇒ **已經中毒的人**再踩毒力場會**受傷**(狀態不是 'G'),而**怪物**踩毒力場
也是受傷(怪物沒有狀態欄)。只寫第一條的話毒力場對敵人完全無效 ——
而那正是玩家最常用它的方式。

⚠ 這一支**沒有敏捷擲骰**。地牢那條(`fieldAffectsParty`)有,戰場這條沒有 ——
兩支不同的函式,不要對齊。

## 5d. ★★ 清過的地牢房間不會再有怪(`sub_F9A0` / `sub_FA20` / `sub_FA7C`)

三支函式組成一個機制,而引擎此前**每次踏上房間格都會再打一場**:

```
sub_F9A0(房號)   打完一間房 → 在位元陣列 byte_3E0F0 上記一筆
sub_FA20(房號)   查那一筆
sub_FA7C()       ★ 進地牢時掃整座地牢的 512 格(8 層 × 64):
                 if ((tile & 0F0h) == 0F0h && sub_FA20(tile & 0Fh))
                     tile &= 0AFh            ; 0xFn → 0xAn
```

`& 0xAF` 清掉 **0x50** 兩個位元 ⇒ 房間格變成 `DungeonRoomA`(可走的空房間),
**不是**通道也不是牆。用「設成 0」之類的簡化寫法會在地圖上多出一片假通道。

呼叫點:`sub_42CC`(進房間)在**恢復地點碼之後**記(順序要緊 ——
戰鬥中 `byte_3E0A3` 是 0xFF,提前呼叫會記到錯的地牢);
`sub_5378`(地牢主迴圈)開頭套用。

### ★ 六間房永遠有怪

原版 `byte_55110`(筆數在 `byte_55116` = 6):`50h 5Bh 41h 46h 4Bh 4Ch`,
鍵 = `房號 | ((地點碼 & 0x0F) << 4)`。低四位元 4 / 5 = 地點碼 0x24 / 0x25 =
**謬誤(WRONG)** 房 1、6、11、12 與 **貪婪(COVETOUS)** 房 0、11。

### ⚠⚠ 兩件事用**兩套不同的索引**

| | 索引 |
|---|---|
| 例外清單的鍵 | `房號 \| ((地點碼 & 0x0F) << 4)` —— **原始的低四位元** |
| 位元陣列 | `DungeonRoomBlock(地點碼)*16 + 房號` —— 有「≥1 就 −1」的修正 |

⇒ **地點碼 0x21 與 0x22 在位元陣列裡共用同一批位元**,但在例外清單裡是
不同的鍵。看起來像 bug,但 `DungeonRoomBlock` 那個修正**早就存在** ——
`DUNGEON.CBT` 的房間查表(`DungeonRoomIndex`)用的是同一個函式。
⇒ 不是算錯,是原版的索引方式。`TestRoomMemoryIsPerDungeon` 把它釘住,
免得日後被「修正」掉。

⬜ **存檔位移未定位** ⇒ 重開遊戲之後房間會全部恢復(原版會記著)。
同 `byte_3E08B`(`docs/re/97` §3)的處境。

## 5e. ★★★ 開局八座地牢入口都是崩塌的(`sub_105E4` + 兩個判準)

`sub_105E4` 是**世界地圖區塊的載入**,而它在載完之後改寫兩種地形:

```
區塊裡的 tile 22 / 23 / 24(洞穴 / 礦坑 / 地牢)
    → 若 sub_1056C(區塊) 為真,改寫成 0xDF(崩塌的入口)
區塊裡的 tile 25(神秘聖壇)
    → 若 sub_105AC(區塊) 為真,改寫成 0x1A(毀壞的聖壇)
```

兩個判準的**極性相反**,而那不是我看錯:

```asm
sub_1056C:  cmp byte_3E0E0[edx], 0    ; 地牢
            setz al                    ; ★ 等於 0 才回真
            ; 區塊不在八座地牢的表裡 → 預設回 **1**

sub_105AC:  cmp byte_3E0E8[edx], 7Fh  ; 聖壇
            setnbe al                  ; ★ > 0x7F(bit 7 設著)才回真
            ; 區塊不在表裡 → 預設回 **0**
```

### 四條獨立證據 ⇒ `byte_3E0E0[i] == 0` 就是「崩塌」

| | 證據 |
|---|---|
| 1 | `sub_1056C` 是 `setz` —— 等於 0 才把入口改寫成 `0xDF` |
| 2 | **`INIT.GAM` 與 `SAVED.GAM` 的那八個位元組全是 0**(實測 offset 0x032A)⇒ 開局全部崩塌 |
| 3 | `byte_55E18` = `18 16 16 18 18 17 17 16`(八座入口的**原始**地形),正是 `sub_105E4` 會改寫成 `0xDF` 的那三個值 ⇒ 兩支函式講同一件事 |
| 4 | tile `0xDF` 的 look 文字是「the collapsed entrance to the dungeon」—— 那是**預設會看到**的敘述 |

⇒ 喊力量之言(`xor byte_3E0E0[i], 80h`,`docs/re/26` §3.2)是把 0 變成 0x80 =
**開封**。這正是 U5 的主線:八座地牢一開始進不去,要各自喊對那個字。

### 引擎此前的兩個錯

1. **地形沒有依旗標重建** ⇒ 直接用 `BRIT.DAT` 的原始地形,一開局所有地牢
   入口都是通的(包括末日)。
2. **`SpeakWord` 的兩句話是反的** ⇒ 玩家第一次喊對,畫面說「入口被封住了」。

而 `ToggleDungeonSeal`(那個 XOR)**本來就是對的** —— 它對稱地在原始地形與
`0xDF` 之間切換,所以只有初始狀態與訊息要修。

### ⚠ 三個已知差異(寫出來,不假裝一樣)

1. 原版是**整個區塊**掃 22/23/24 ⇒ 同一區塊裡**其他**的洞穴與礦坑也會一起
   崩塌;而「區塊不在表裡 → 預設回真」表示**不屬於任何地牢的**洞穴/礦坑
   一律崩塌。引擎只改八個入口本身 —— 要做整區塊得先 dump `byte_55140`。
2. **末日(DOOM)的 (128,128) 在地表是深水** —— 它從幽冥界進去,地表沒有入口。
   `applyWorldFlags` 的「目前是原始入口地形才改」那道守衛擋住了這一格,
   所以不會把海面改成崩塌的洞口。
3. ⬜ **「崩塌的入口進不進得去」還沒查。** `u5data.DungeonAt` 與原版
   `sub_2D564` **都是看座標**,不看地形 ⇒ 目前地形雖然顯示崩塌,人還是走得進去。
   原版是不是在別處擋了一道,要再讀 `sub_2D564` 的呼叫端。
   ★ 這個未知**不會造成 soft-lock**(寧可放人進去也不要把路封死),
   所以先照現在這樣,不要憑猜加一道門檻。

## 6. 引擎落地

| 原版 | 引擎 |
|---|---|
| `sub_2BCC8` | `(*State).regenerateParty` |
| `sub_2ECE8(byte_3E0AE)` | `(*State).ringUpkeep`(用 `Combat.Turn`)|
| `sub_1F4E4` | `(*State).expireFields` |
| `sub_B210(槽+1)` | `ObjectSet.Remove`(⚠ 原版只清 +0..+5,引擎清整筆 8 B;+6/+7 對力場沒有意義)|
| `sub_FEC` | `(*State).wanderHorses` |
| `sub_FB4` | `(*State).horseTied` |
| `sub_2A50C` 尾段 | `upkeep()` 的最後三行 + `settleHour()` + `tickModeOutOfCombat()` |
| `byte_3E09B` | `State.turnsSinceAlms` / `TurnsSinceAlms()` |
| `sub_1B854` 尾段 | `(*State).almsKarma` |
| `sub_2EDC0` / `sub_2EDF8`(紮營讓人睡著)| `camp()` 的「不是 'P' 就睡」分支 + `putUnitToSleep()` |
| `sub_BBA0` | `harmUnderUnit()` + `harmStandingUnit()`(接在 `afterPlayerAction` 與 `aiTurn` 之後)|
| `sub_B8DC` | `poisonOrHurt()` |
| `sub_F9A0` / `sub_FA20` / `sub_FA7C` | `markRoomCleared()` / `roomIsCleared()` / `applyClearedRooms()` |
| `byte_3E0F0` | `State.roomsCleared`(⬜ 存檔位移未定位)|
| `byte_55110` / `byte_55116` | `u5data.DungeonRoomAlwaysArmed()` |
| `sub_105E4` 的兩種改寫 | `applyWorldFlags()`(在 `LoadFrom` 尾端)|
| `sub_1056C` | `u5data.DungeonIsSealed()`(★ 極性與名字直覺相反)|
| `sub_105AC` | `ShrineFlag[i] & ShrineDesecratedBit`(這一邊極性正常)|
| `sub_165C8` / `sub_21D48` 的 `sub_2BCC8` | `camp()` 每小時 / `SleepUntilMorning()` 每 9 分鐘 |
| `sub_16370` | `afterPlayerAction()`(`ringUpkeep` → `expireFields` → `tickCombatMode`)|

### 測試

| 測試 | 驗什麼 |
|---|---|
| `TestRegenerationRingHealsOneInEight` | 1/8 的機率(800 次落在 55..160)|
| `TestOnlyTheRegenerationRingHeals` | 反對照:另外兩只戒指不回血 |
| `TestRegenerationStopsAtMaxHP` | 上限是那個人的 MaxHP |
| `TestRegenerationSkipsOnlyTheDead` | ★ 五種狀態逐一 —— 只有 'D' 不回 |
| `TestFieldsExpireOneInSixteen` | 1/16,而且**不印訊息** |
| `TestOnlyFieldsExpire` | 反對照:別的物件不會自己消失 |
| `TestTurnCounterCountsEveryTurnNotEveryMeal` | ★ 語意更正的驗收 + 上限 255 |
| `TestStarvingHurtsOncePerHour` | ★★ 同小時不重複,**而且換小時會再餓**(反對照)|
| `TestAlmsKarmaNeedsTheThrottle` | 節流 + 計數器歸零 |
| `TestAlmsGivingEverythingAddsTwoMore` | +3 與 +1 兩種情況 |
| `TestAlmsKarmaIsCappedAt99` | 兩次各自夾 99 |
| `TestAlmsOnlyForBeggars` | 四個朝向都算 / 上下緣外都不算 |
| `TestLooseHorseWanders` | 沒繫住會走 |
| `TestTiedHorseStaysPut` | ★★ 兩種地形 × 四個鄰格 = 八組 |
| `TestHorseFacingFollowsHorizontalMovesOnly` | ★ 垂直移動不換朝向(而且驗到「真的發生過垂直移動」)|
| `TestHorseStaysInsideTheScene` | ★ 不環繞 |
| `TestOnlyHorsesWander` | 反對照:別的物件不會走 |
| `TestHorseDoesNotWalkOntoAnotherObject` | 目標格有東西就不走 |
| `TestCampSleepUsesTheNotPoisonedRule` | ★★ 四種狀態 —— **被惑那一格**是唯一能分辨新舊行為的(已跑反對照確認會紅)|
| `TestRegenerationRingWorksWhileCamping` | 紮營時戒指真的有作用(拿中毒的人測,排除紮營自己的恢復)|
| `TestLavaAndFireplaceBurnInCombat` | ★ 戰場上的熔岩與壁爐會燒人 |
| `TestSwampPoisonsInCombat` | 戰場上的沼澤會上毒 |
| `TestPoisonHurtsWhenItCannotStick` | ★★ 毒上不了身改扣血(已中毒的人再踩會受傷)|
| `TestTerrainWinsOverObjects` | ★ 地形優先,中了就不掃物件 |
| `TestThreeFieldObjectsAndTheFourthDoesNothing` | ★ 0xEB 沒有 case |
| `TestSleepFieldClearsPoison` | ★★ 睡眠力場把中毒擦掉(單一狀態位元組)|
| `TestSleepingDeadStaysDead` | 死人不會被叫去睡 |
| `TestHarmlessFloorDoesNothing` | 反對照:普通地板一百回合什麼都沒發生 |
| `TestClearedRoomIsWipedOffTheMap` | ★★ `0xFn → 0xAn`,而且房號保留 |
| `TestSixRoomsAreNeverMarkedCleared` | ★ 六筆逐一 + 四條反對照(同地牢別的房、別的地牢同房號)|
| `TestAlwaysArmedRoomIsNotRemembered` | 例外生效,而且**不是整支關掉**(反對照)|
| `TestRoomMemoryIsPerDungeon` | ★ 0x21 與 0x22 共用位元 —— 釘住原版的索引怪處 |

## 7. ⬜ 這一輪讀完但**判定為顯示層**的(不需要引擎邏輯)

| 函式 | 是什麼 | 為什麼不需要 |
|---|---|---|
| `sub_29F4C` / `sub_2A33C` | 面板上印一行角色(名字補到 9 欄 + HP + 狀態字元)| Ebiten 那一層已經在畫。⚠ `sub_2A33C` 對第 0 位印字元 26(CP437 `→`)而 `sub_29F4C` 印空白 —— 兩支差這一格,語意未定 |
| `sub_E460` | 在第 5 列印 12 個空白(清訊息列)| 同上 |
| `sub_5F98` | 載 `ULTIMA.16` + `WD.BIT` 的標題畫面;順手把地點碼設成 0x40 | 美術流程(P6);地點碼 0x40 = 標題畫面這件事記在這裡 |
| `sub_3124` | 寫 `dword_4FDD8/DC` 並包一對 `sub_27754` | 疑為游標 / 中斷遮罩,不碰遊戲狀態 |
| `sub_DBC8` | 把 11×11 覆蓋層填成 0xFF | 引擎的覆蓋層是另一套資料結構 |
| `sub_10208` | **空函式** | — |
| `sub_2BCC8` 之外的 `sub_21CC0` / `sub_22234` / `sub_22254` | 名冊游標:數 `CharInnFlag == 0` 的人數、往前 / 往後找下一個在隊伍裡的人 | 旅店 / 隊伍管理畫面的清單瀏覽器 |

## 8. ⬜ 還沒讀完的(下一輪)

`sub_BE5C`(188 行)是一支**過場動畫腳本解譯器**,九個 opcode:

```
0 結束 / 1 雙物件模式 / 2 讀重複次數 / 3+4 開關逐步重畫 / 5 延遲
6 寫 byte_3F8F4[y*32+x] / 7 重畫 / 8 等 n / 9 清掉某個物件槽
≥0x10 → sub_BD80(opcode) 取物件槽,讀 word_3E086/88 當位移,跑 n 次
```

呼叫者 `sub_C2BC` / `sub_C2D0` / `sub_C414`(三處)。⚠ 與 `sub_3007C`
(開新遊戲的腳本解譯器,`WORKLIST §5.2b`)是**兩支不同的解譯器**。

其餘未讀:`sub_105E4`、`sub_BBA0`、`sub_92C0`、`sub_F9A0` / `sub_FA20`
(一組位元旗標,索引式子 `(a1−528)/8` 位元 `(a1−16)&7` 待化簡)、
`sub_2D944`、`sub_92C0`(NPC 腳下的 tile 與 `byte_3E579` 排程欄,疑為門 / 床)、
`sub_31AC`(★ 地牢 tile 讀取,`< 0x90` 時**清掉 0x08 位元** —— 要與引擎的
`DungeonTileAt` 對照,可能是一條落差)、`sub_13B30`(★ 找空物件槽,
**從 31 往下**,與引擎 `freeObjectSlot` 由 1 往上相反;⚠ 那是 `sub_2B57C`,
是**另一支**函式,不要合併)。
