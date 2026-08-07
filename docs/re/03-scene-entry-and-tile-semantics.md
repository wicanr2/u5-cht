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

## 6. 地點 → 場景地圖:完整規則 ✅

`sub_5C8` 就是載入場景地圖的函式。組語三行講完:

```asm
movzx   eax, byte_3E0A3          ; 當前地點編號(1-based)
dec     eax
sar     eax, 3                   ; ÷ 8
mov     eax, off_4FC44[eax*4]    ; 檔案 = {TOWNE,DWELLING,CASTLE,KEEP}[(編號-1)/8]

movzx   eax, byte_41033[編號]    ; 該地點的起始地圖索引
movzx   ebx, byte_3E0A5          ; 樓層
add     ebx, eax
cmp     byte_3E0A5, 7Fh
jbe     short loc_630
sub     ebx, 100h                ; 樓層 > 0x7F → 減 0x100(**地下層用負數**)

loc_630:
shl     eax, 0Ah                 ; × 1024
push    400h                     ; 讀 1024 B(32×32,印證解碼)
push    offset byte_400F4
call    sub_2C740                ; 先前已知的讀檔常式
```

dump `byte_41033[1..32]` 得到起始索引,再用「同檔下一個地點的起始索引 − 自己」算出層數:

| 檔案 | 地點與樓層 |
|---|---|
| `TOWNE.DAT` | MOONGLOW 0-1、BRITAIN 2-3、JHELOM 4-6、YEW 7、MINOC 8-9、TRINSIC 10-11、SKARA BRAE 12-13、NEW MAGINCIA 14-15 |
| `DWELLING.DAT` | **四座燈塔各 3 層**(0-2 / 3-5 / 6-8 / 9-11)、IOLO'S HUT 12、三間無名小屋 13/14/15 |
| `CASTLE.DAT` | (無名 #17,**Lord British 城堡**)**1-5 共 5 層**、#18 6-9、WEST/NORTH/EAST BRITANNY 10/11/12、PAWS 13、COVE 14、BUCCANEER'S DEN 15 |
| `KEEP.DAT` | ARARAT 0-1、BORDERMARCH 2-3、FARTHING 4、WINDEMERE 5、STONEGATE 6、THE LYCAEUM 7-9、EMPATH ABBEY 10-13、SERPENT'S HOLD 14-15 |

### 畫面驗收:燈塔三層

`u5dump town gamedata <U5_E> FOGSBANE out.png` 畫出 `DWELLING.DAT` 索引 0–2:

1. **底層** —— 完整建築,有房間、桌椅、床、書櫃、樓梯
2. **二層** —— 只剩塔身與樓梯
3. **頂層** —— 圓形塔室,中央是**發光的燈**

逐層收窄、頂層是燈室 —— 這正是燈塔該有的樣子。三張圖若對應錯了,不會這麼連貫。

### 進入場景時的初始狀態(`sub_10928` 結尾)

```c
byte_3E0A3 = i + 1;   // 當前地點編號(0 = 在世界地圖)
byte_3E0A5 = 0;       // 樓層(0 = 地面層)
byte_3E0A6 = 15;      // 場景內 X(32 格的中央)
byte_3E0A7 = 30;      // 場景內 Y(靠底部 → 城鎮南方入口)
```

## 7. 場景內移動:`sub_86C` 全解 ✅

這是進場景之後每按一次方向鍵會走的路。Hex-Rays 在這個函式裡錯了**四處**,全部要看組語才對得起來。

```c
switch (按鍵) {          // 1=西 2=東 3=北 4=南
  case 1: dx=-1; facing=3; atEdge = (x < 1);    break;   // ← Hex-Rays 印成 v1=1 常數
  case 2: dx=+1; facing=1; atEdge = (x > 30);   break;
  case 3: dy=-1; facing=0; atEdge = (y < 1);    break;
  case 4: dy=+1; facing=2; atEdge = (y > 30);   break;
}
tile = 視窗[32*dy + dx];                 // 玩家鄰格
obj  = sub_2B360(x+dx, y+dy, 樓層);      // 這一格有沒有 NPC / 物件
...
if (可通行 && sub_2A694(byte_3E08C, tile)) {
    if (atEdge) { 問 "Dost thou wish to leave?"; ... }
    else { x += dx; y += dy; sub_758(facing, tile); }   // ← Hex-Rays 印成 `=` 而非 `+=`
} else { "Blocked!"; 嗶一聲 sub_2C4F4(165, 200); }
```

### Hex-Rays 在這裡的四個錯

| 位置 | 反編譯印的 | 組語實際 | 影響 |
|---|---|---|---|
| 邊界旗標 | 每個方向一個常數(西/北 = 1,東/南 = 0) | `cmp byte_3E0A6, 1 / jnb` 等四組比較 | 照抄的話往東往南永遠出不了城 |
| 座標更新 | `byte_3E0A6 = dx` | `add byte_3E0A6, al` | 照抄的話走一步就跳到 (±1, ±1) |
| 樓層增減(`sub_758`) | `byte_3E0A5 = 1` / `= -1` | `inc` / `dec byte_3E0A5` | 照抄的話只能在 1F 與 B1F 之間跳 |
| 通行判定第一參數 | `sub_2A694(0, tile)` | `movzx eax, byte_3E08C` | 照抄的話船、馬、飛毯全都照步行規則走 |

### 邊界的 off-by-one

比的是**移動前**的座標:`x < 1`、`x > 30`、`y < 1`、`y > 30`。也就是站在最外圈(0 或 31)再往外走才算離開;
站在 x=1 往西只是走到 x=0。搞反的話,城鎮最外圈那一圈永遠踩不到。

### 離開場景

```asm
cmp     byte_3E0A3, 19h            ; 地點 25(ARARAT)特判
; → "Underworld!" + 樓層 = -1 + BGM 10;其餘 → "Britannia!" + 樓層 0 + BGM 1
movzx   edx, byte_3E0A3
mov     cl, byte_410F3[edx]        ; 世界座標**從地點表讀回來**(1-based 索引的同一張表)
mov     byte_3E0A6, cl
mov     dl, byte_4111B[edx]
mov     byte_3E0A7, dl
mov     byte_3E0A3, al             ; = 0 → 回到大地圖
```

所以 `byte_3E0A6/A7` 是**共用**的一對座標:在大地圖是世界座標,在場景是場景座標。
而且在城裡走到哪一格出去都會回到城門那一格 —— 遊戲不記得你從哪裡進來的。

### 視窗緩衝

`byte_3F789` 看起來像一個單獨的 byte,但它被 `[32*dy + dx]` 這樣索引(dy,dx ∈ −1..1),
會讀到自己前後的位址。真相是它是**視窗緩衝裡玩家那一格的位址**:

- `memset(byte_3F6E4, 0xFF, 0x160)` —— 0x160 = 352 = **11 列 × 32 stride**
- `0x3F789 − 0x3F6E4 = 0xA5 = 32×5 + 5` —— 正好是 11×11 的正中央

所以視窗永遠以玩家為中心,四鄰用 ±1 / ±32 直接定址。空白填 **tile 0xFF**(算繪出來是純黑),
不是 tile 0 —— tile 0 是一團紅黃爆裂圖案,填錯會在城鎮外緣鋪出一片火花。

### 上下層是兩套機制

| 機制 | tile | 觸發 | 出處 |
|---|---|---|---|
| 方向樓梯 | 0xC4–0xC7(低 2 bit = 朝向) | **走進去**:同向上樓、反向(`facing ^ 2`)下樓 | `sub_86C` → `sub_758` |
| 梯子 | 0xC8 上 / 0xC9 下 / 0x86 活板門(下) | 站在上面按 **K**(Klimb) | `sub_EA0` → `sub_758(0 或 2, 196)` |

`sub_EA0` 的作法很巧:它不重寫升降邏輯,而是拿合成的 tile 196 去呼叫 `sub_758`,
靠 facing 0(同向)或 2(反向)決定上下。

**梯子才是主力**:四座燈塔、兩座城堡、修道院、巨蛇要塞全靠梯子,只實作樓梯的話那些地方爬不上去。

## 8. 樓層範圍:用梯子拓撲反推 ✅

§6 用「同檔下一個地點的起始索引 − 自己」算層數 —— **那個推法會錯**,因為它假設地圖索引
從地面層往上長。實際上有些地點的地下層排在 `SceneIndex` **之前**。

梯子提供了一個獨立於執行檔的驗證方式:**上下層的梯口一定落在同一格**
(某層 (x,y) 是 0xC8,上一層同一格就是 0xC9)。拿它去掃四個檔:

```
連通處:7/7、3/3、6/6、2/2… 位置全對齊
交界處:0/0                    一個都不對齊
```

再配合地面層索引表 `byte_41033`(從各地點的地面層往下追連通性),得到的分割
**四個檔各 16 張、恰好被 8 棟建築蓋滿、不重疊也不留空**:

| 檔案 | 建築(地圖範圍 → 樓層範圍) |
|---|---|
| `TOWNE` | 七大城市各 2 層;**紫衫城 6–7 → −1..0**(地下是監獄) |
| `DWELLING` | 四座燈塔各 3 層(0–2 / 3–5 / 6–8 / 9–11);四間小屋各 1 層 |
| `CASTLE` | **獅心城堡 0–4 → −1..+3**、**黑刺宮殿 5–9 → −1..+3**(都有地下);六個村莊各 1 層 |
| `KEEP` | ARARAT 0–1、BORDERMARCH 2–3、三個 1 層、書院 7–9、修道院 10–12、**巨蛇要塞 13–15 → −1..+1** |

修正了 §6 那份表的六筆:哲倫 3→2 層、紫衫城 1→2(有地下)、獅心/黑刺的範圍下移一層、
修道院 4→3、巨蛇要塞 2→3(有地下)。

驗收:`TestSceneMapsFullyPartitioned` 檢查分割不重疊不留空;`TestLadderChain` 用原版資料
實際爬過**全遊戲 101 段梯子**,每一段都檢查落點可以再下來。唯一的例外是 ARARAT ——
它的落點是 tile 134 而不是 0xC9,而 `sub_EA0` 的第一個判斷 `if (v2 != 134)` 正好就在處理它。
