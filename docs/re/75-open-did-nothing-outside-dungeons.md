# 75 — Open 在地牢外什麼都不做,而門會自己關上

| | |
|---|---|
| 輸入檔 | `re_work/fmtowns/WORRIORS.EXP.asm`(FM Towns 英文版) |
| SHA-256 | 見 `re_work/fmtowns/SHA256SUMS.txt` |
| 主要函式 | `sub_15374`(Open 指令本體)、`sub_152B8`(地牢的開箱)、`sub_15108`(物件層的開箱)、`sub_2B64C`(把 tile 寫回去)、`sub_1A54`(主迴圈的倒數) |
| 落地 | `internal/u5data/open.go` · `internal/game/chest.go` · `state.go` · `strings.go` |
| 從哪裡來 | `docs/re/66` 的截斷清單(`sub_152B8`) |

---

## 1. 症狀:O 在地牢外只會說「此處沒有寶箱。」

```go
func (s *State) OpenChest() {
	if s.InDungeon() { … }
	s.Log("此處沒有寶箱。")     // ← 地表與城鎮的 Open 整條都沒有
}
```

⇒ **門打不開、地上的箱子開不了。** 而原版的 Open 是一個鍵三條路:

```
sub_15374:
if (0x21 <= 地點 <= 0x28)          → sub_152B8()        ; 地牢:開腳下的箱子
else {
    sub_2B64C(byte_3E161, …)        ; ★ 先把上一扇門關掉
    問方向 → (x, y)
    switch (tile):
      0x97 / 0x98                   → "Locked!"          魔法鎖
      0x99                          → "Too heavy!"       柵門
      0xAF                          → "It's open!"
      0xB8 / 0xBA                   → 打開(見 §3)
      0xB9 / 0xBB                   → "Locked!"
      其餘(含 0xB0..0xB7)          → sub_15108(x, y, 樓層)   物件層
}
```

四個門的 tile 與 `docs/re/69`(砲彈打得掉的七種門)完全同一族,
`look#` 名字逐一對上:0x97/0x98 奇怪的門、**0x99 柵門**(「太重了」——
柵門本來就不是用手開的)、0xB8 木門、0xB9 上鎖的門、0xBA 有窗戶的木門、
0xBB 有窗戶的上鎖的門。

## 2. ★ 地牢那條讀的是**腳下**,不是面前

`sub_152B8` 開頭:

```asm
movzx eax, byte_3E0A5 ; shl eax, 6        ; 樓層 × 64
mov dl, byte_3E0A7 ; and dl, 7            ; ★ 玩家自己的 Y
mov cl, byte_3E0A6 ; and cl, 7            ; ★ 玩家自己的 X
lea eax, dword_3E16C[eax+edx*8]
movzx edi, byte ptr [eax+ecx]
```

引擎原本走 `dungeonFacingTile`(腳下**或面前**)。那條規則是真的,
但它的出處是 **`sub_18D18` —— An Sanct(施法)選目標**用的,不是 Open。
把施法的規則套到 Open 上,玩家隔一格就開得到箱子。

而且兩支的**陷阱判準也不同**:

| | 判準 |
|---|---|
| `sub_152B8`(Open) | `test di, 7` —— **低三位元全部** |
| `sub_18D18`(An Sanct) | `test byte [eax], 1` —— 單一位元 |

⇒ 又是 `docs/re/74` §1 那個形狀:**同一件事兩支各做一半,而且細節不同**。
這是第四例(沼澤中毒、戒指消失、擦傷、地牢寶箱陷阱)。

## 3. ★★ 門會自己關上 —— 而且同時只能開一扇

開門那一段:

```asm
loc_1546A:
    byte_3E161 = tile          ; 原本是哪一種門
    byte_3E164 = 4             ; ★ 倒數
    byte_3E162 = x
    byte_3E163 = y
    sub_DB10(x, y) → *ptr = 0x44   ; 那一格變成磚地
    byte_4198A = 1
    印 "Opened!"
```

主迴圈 `sub_1A54` 每回合:

```asm
call sub_1318                   ; 地形效果
mov  al, byte_3E161
and  al, al ; jz → 跳過
dec  byte_3E164
cmp  byte_3E164, 0 ; jnz → 跳過
sub_2B64C(byte_3E161, byte_3E162, byte_3E163)    ; 把 tile 寫回去
```

而 `sub_2B64C(tile, x, y)` 只有五行:

```asm
if (byte_3E0A3 != 0 && byte_3E0A3 < 0x21 && tile != 0)
    sub_DB10(x, y) → *ptr = tile
```

⇒ **門開 4 回合就自己關上**,而且**只在場景裡**寫回
(大地圖與地牢沒有門要關)。

### 那四個變數只有一組

`sub_15374` 在**問方向之前**就先呼叫一次 `sub_2B64C(byte_3E161, …)` ——
所以開新門會立刻把上一扇關掉。**同時只能有一扇門是開的。**

這不是最佳化,是可觀察的行為:走過一長串門會看到後面的自己關上。
`TestOnlyOneDoorCanBeOpenAtATime` 把它釘住。

⚠ 開門寫進去的 **0x44** 與 An Ylem(消除)寫的是同一個值 ——
0x44 是「什麼都沒有的地板」,不是「開著的門」。所以原版**沒有**開門的美術;
門開著的時候那一格就是地板。

## 4. `sub_15108` —— 物件層的開箱,四條容易漏的規則

```asm
for (esi = 1; esi < 32; esi++)                 ; ★ 從槽 1 開始(槽 0 是隊伍)
    座標對不上 → continue
    不在戰鬥時樓層也要對 → continue
    kind == 1    → 找到了(上鎖的箱)
    kind == 0x0E → 印 "Can't!" 並 return       ; ★ 檀香木盒打不開
找不到 → 印 "Nothing to open!"
who = sub_E19C();  who == −1 → return
var_5 = 物件[esi].品質
sub_2B6C8(…, esi)                              ; 清掉那一槽
if (1 <= 地點 <= 0x20) {                       ; ★★ 在場景裡
    if (byte_3E098 > 2) byte_3E098 -= 2 else byte_3E098 = 0    ; 業報 −2
}
if (var_5 > 0x7F) {                            ; ★ 品質最高位 = 有陷阱
    var_5 &= 0x7F
    印 "Trapped!";  sub_2AB38(who)
    if (地點 > 0x7F && 狀態[who] == 'D') { …戰鬥中被陷阱炸死的收尾… }
}
sub_15020(var_5, …);  sub_1509C(var_5, …)      ; 擲獎品(等級 = 品質低七位)
獎品一件都沒掉 → 印 "Chest empty!"
```

| ★ | 規則 |
|---|---|
| 1 | **從槽 1 開始掃** —— 槽 0 是隊伍自己 |
| 2 | **檀香木盒印 "Can't!"**,與 "Nothing to open!" 是兩句不同的話 |
| 3 | **品質一個位元組裝兩件事**:最高位 = 有陷阱、低七位 = 獎品等級。原版 `and var_5, 7Fh` 清掉陷阱位才擲獎品 —— 少了那一步,有陷阱的箱子會被當成 133 級寶箱 |
| 4 | **在場景裡開箱子扣 2 點業報**(下限 0)。大地圖上的箱子無主,不扣;地牢也不扣 |

第 4 條是「翻別人家的箱子」的代價 —— 與 `docs/re/69` 的砲擊 −5、
`get.go` 的 Get 罰則同一族。引擎完全沒有。

## 5. 落地與驗收

| 檔案 | 內容 |
|---|---|
| `u5data/open.go` | `OpenActionFor(tile)` 五種處置 + 四個常數 + 兩個物件種類 |
| `game/chest.go` | `OpenChest` 三條路、`openDungeonChestUnderfoot`、`openToward`、`openObjectChest`、`pendingDoor` / `tickDoor` / `closePendingDoor` |
| `game/state.go` | `door pendingDoor` 欄位;`tick()` 尾端 `tickDoor()` |

七條測試。兩條值得記:

- `TestOpenedDoorBecomesFloorAndClosesAfterFourTurns` —— 撐滿 3 回合都還開著、
  第 4 回合關上。**兩邊都驗**,不然「一開就關」也會過。
- `TestOpeningAChestInTownCostsKarma` 第一版紅了,訊息是「箱子還在物件層上」。
  原因不在移除,在**我用 `ObjectAt(x, y)` 找那個箱子** —— 城裡那一格上還站著
  NPC 的鏡射物件,`ObjectAt` 先撞到它。改成記住 `Spawn` 回傳的槽號。
  ⇒ **失敗訊息指向結果,原因在測試自己的取值方式**(第五次)。

## 6. 還沒讀的

- ⬜ **tile 0xAF 為什麼是 "It's open!"**:`look#175` 給它的名字是
  「沉重的行李箱」(來自 `LOOK2.DAT`),而「它是開著的」對行李箱說不太通。
  `look#<tile>` 這個鍵已用 `look#71 = 碼頭`(`TileDock = 0x47`)再驗過一次是對的,
  所以不是索引錯。**照值實作,不編故事。**
- ⬜ `sub_1509C`(第二支擲獎品的函式,與 `sub_15020` 成對)沒讀。
  引擎目前只有 `rollChestContents`(= `sub_15020`)。
- ⬜ `sub_15108` 裡「戰鬥中被陷阱炸死」那一段(改戰場單位旗標與 tile、
  清 `byte_3E08B`)未實作 —— 引擎的 `chestTrapVictim` 只扣血判死。
- ⬜ `chestTrapVictim` 的傷害仍是估計值(`random(1, 20)`);`sub_2AB38` 沒讀。
- ⬜ `byte_4198A` 的位元語意(`docs/re/69` 也列著這一條)—— 開門設 1、
  開箱 `or 2`,疑為畫面重繪旗標。
