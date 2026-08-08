# 68 — 橋下的食人妖與撞船:截斷清單挖出的兩個遭遇

| | |
|---|---|
| 輸入檔 | `re_work/fmtowns/WORRIORS.EXP.asm`(FM Towns 英文版) |
| SHA-256 | 見 `re_work/fmtowns/SHA256SUMS.txt` |
| 主要函式 | `sub_3010`(過橋)、`sub_2F48`(過路費)、`sub_2CE70`(通行判定 + 撞擊)、`sub_2D9D0`(移動後的分派) |
| 落地 | `internal/game/troll.go`、`internal/game/shipdamage.go`、`internal/u5data/shipdamage.go` |
| 從哪裡來 | `docs/re/66` 的截斷清單:`sub_3010` 78 行 → 3 行 C、`sub_2CE70` 174 行 → 35 行 |

---

## 1. 橋下的食人妖

`sub_3010` 反編譯只剩三行,而組語裡是一整個過橋遭遇。**引擎原本走過橋什麼都不會發生。**

### 觸發

`sub_2D9D0`(移動之後)取腳下的 tile:

```asm
        call    sub_DB10              ; 取 (X, Y) 的 tile
        movzx   esi, byte ptr [eax]
        mov     eax, esi
        and     eax, 0FEh
        cmp     eax, 6Ah              ; ★ 橋(0x6A / 0x6B)
        jnz     short loc_2DC34
        call    sub_3010
```

`look#106` / `look#107` 都是「橋」,對得上。

### `sub_3010`:偷偷溜過去

```
if (rand(0, 7) != 0) return                   ; ★ 1/8 才遇到
if (載具 != 0x1C)    return                   ; ★ **只有步行**
印 "\nThou spieth trolls under the bridge!\n\n"
for (每個隊員) {
    if (狀態 == 'D' || 狀態 == 'S') continue   ; 死的與睡著的不參加
    印 <名字> " sneaks across" + 三個「.」(中間有延遲)
    if (敏捷 >= rand(1, 30)) continue          ; 過關
    sub_2F48()                                 ; 被抓 → 收過路費
    return                                     ; ★ 一人被抓,整隊停下
}
印 "Trolls evaded!"
```

⚠ 載具判的是 **`cmp byte_3E08C, 1Ch`——單一值**,不是 `IsOnFoot` 那個
「0x1C 或 0x1D」的兩值判斷。用 `IsOnFoot` 代替會多放行 0x1D。

### `sub_2F48`:過路費

```
印 "Caught!"
sub_2B67C()                                  ; 找第一個能行動的人 → word_3E086
通行費 = 0x63 − 那個人的**力量** × 3
印 <通行費> "gp toll!"  "Dost thou pay?"
'Y' → 金錢 -= 通行費;**扣完變負數就退回並開打**
'N' → 開打
開打:生成種類 0xE4 在腳下那一格,然後 sub_2E58C
```

★ 種類 **0xE4**:`(0xE4 − 0x40) / 4 = 41`,而怪物名表第 41 筆(`0x3F608`)是
`aTroll db 'Troll',0` ✓。

★ **偷渡擲敏捷、過路費看力量** —— 兩個**相鄰**欄位:

| 全域 | 減名冊基底 `byte_3DDB4` | 是什麼 |
|---|---|---|
| `byte_3DDC0` | 0x0C | **力量**(通行費) |
| `byte_3DDC1` | 0x0D | **敏捷**(偷渡) |

基底可以獨立驗:`byte_3DDBF` − `byte_3DDB4` = 0x0B = `CharStatus`,
而那一格是全篇判死活用的。**兩個欄位差一,很容易看成同一個** ——
所以測試裡有一條專門把它們分開驗(敏捷 99 一定溜過、力量 10 的通行費一定是 69)。

★ **原版的算式沒有下限**:力量 34 以上通行費是負數,而 `word_3DFB6 -= 費用`
會變成**加錢**。照原樣保留 —— 那是原版的算式,不是這裡寫錯。
`TestTollFormulaHasNoFloor` 把「力量 ≥ 34 就該是負數」寫死,免得日後有人「順手修」。

---

## 2. 撞船:`sub_2CE70` 是通行判定與撞擊的同一支

`docs/re/66` 列出 `sub_22F0`(船身損傷)的六個觸發點時,只知道
「`sub_2CE70` 是撞擊 / 觸礁那條路,**觸發條件沒讀**」。讀完之後條件很具體,
而且四個地形值全部對得上語意:

```
if ((載具 & 0xFC) == 0x24) 印 "Rowing!"        ; ★ 收帆的船每走一步都印
esi = sub_2B360(X+dx, Y+dy, 樓層)              ; 目標格上有什麼
…一串「載具 × 目標」的判斷 → edi = 過不過得去…
ebx = byte_3F789[dy*32 + dx]                   ; 視窗緩衝裡那一格的地形
if (過得去) return 1

if (沒揚帆) {                                   ; byte_3E167 == 0
    if (載具 >= 0x20 && (esi & 0xFC) == 0xEC) return
    印 "Blocked!"
    if (ebx == 0x2F) { 印 "OUCH!"; sub_2A4D0() }   ; ★ 仙人掌
    音效; return
}
; 揚著帆撞上東西
if (ebx == 3)        印 "BREAKING UP!"          ; ★ 淺灘
else if (ebx != 47h) 印 "COLLISION!"
if (ebx == 47h) { 印 "Docked!"; 載具 += 4; 收帆 }  ; ★ 碼頭
else            { 音效; sub_22F0(); 收帆 }         ; ★ 船身受損
byte_3E168 = 1
```

### 四個地形值,四個語意都對上

| 值 | `look#<tile>` | 訊息 | 說得通嗎 |
|---|---|---|---|
| 1 | 深水 | `Rough seas!`(小艇 / 魔毯) | ✓ 小船在深水遇風浪 |
| 3 | 淺灘 | `BREAKING UP!` | ✓ 揚著帆衝進淺灘,船身裂開 |
| 0x2F | 仙人掌 | `OUCH!` + 全隊受傷 | ✓ |
| 0x47 | 碼頭 | `Docked!` | ✓ 靠岸,不是撞擊 |

四個一次全對,而且 `look#<N>` 的鍵**直接是 tile 編號**(`i18n.LookKey`)——
這比逐個位址對照更硬。順帶把 `docs/re/66` 留的
「`RoughSeasTile = 1` 沒解釋成水深」那一條也結掉了:**是深水**。

### 順序有一個容易寫錯的地方

- **淺灘會印兩件事**:先 `BREAKING UP!`,接著仍走「受損」那條(音效 + `sub_22F0` + 收帆)。
- **碼頭一句撞擊都不印**:第一段的 `jz` 直接跳過,只印 `Docked!`。
- 靠岸是 `add byte_3E08C, 4` —— 揚帆組(0x20..0x23)加 4 就是收帆組(0x24..0x27),
  **朝向自然保留**。寫成 `Transport = VehicleShip` 會把朝向清掉。

### `Rowing!`

`sub_7C0`(移動時印動詞)對船**不印任何動詞**,而 `sub_2CE70` 開頭對
**收帆的船**(`& 0xFC == 0x24`)印 `Rowing!`。兩支不同的函式各印一半 ——
所以揚著帆時安靜,收了帆划行時每步都唸。

---

## 3. 落地與驗收

| | |
|---|---|
| `game/troll.go` | `crossBridge` / `trollToll` / `firstAbleMember` / `trollFight` |
| `game/shipdamage.go` | `blockedMove`(四種撞擊結果) |
| `u5data/shipdamage.go` | `TileShallowWater` / `TileCactus` / `TileDock`,`RoughSeasTile` 補上證據 |
| `game/state.go` | 移動後判橋;`TileBlocksWalking` 改走 `blockedMove`;收帆的船印 `Rowing!` |

九條測試,其中幾條是專門擋「寫得太寬鬆」的:

- `TestOnlyFootTravellersMeetTheTrolls` —— 擋「用 `IsOnFoot` 代替單一值 0x1C」
  (連 `VehicleWalk + 1` 都要擋掉)
- `TestSneakUsesDexterityAndTollUsesStrength` —— 擋兩個相鄰欄位看錯
- `TestTollFormulaHasNoFloor` —— 擋「順手給通行費加下限」
- `TestCollisionOutcomesByTerrain` —— 逐地形驗,含「淺灘不印 COLLISION」
  與「碼頭不印撞擊」
- `TestDockingLowersTheSailsAndKeepsFacing` —— 四個朝向都驗
- `TestSleepingAndDeadMembersDoNotSneak` —— 睡著的人不該被點名

## 4. 還沒讀的

- `sub_2CE70` 中段那一串「載具 × 目標種類」的通行判斷(`esi` 與 0x24 / 0x2C /
  0x1B / 0x10 的比較)沒逐條追 —— 引擎目前仍用自己的 `TileBlocksWalking`
  加 `ModeOf`,兩者是否等價未驗。
- `sub_2B360`(目標格上有什麼)只當黑盒用。
- `byte_3E168` 設 1 之後誰讀它。
- 撞擊音效 `sub_2C598(64h, 7D0h, 12Ch)` 的參數含意。

---

## 追記:大地圖上的攀爬,以及一個被寫死成 `false` 的旗標

`state.go` 的 `Klimb()` 原本寫著:

```go
if !s.InScene() {
    // 大地圖上的攀爬(上山、進地牢)是另一條路徑,還沒做。
    s.Log(MsgNothingToClimb)
    return
}
```

那條路徑就是 `sub_188C4`,而它也在截斷清單上(110 行 → 3 行 C)。

```
if (byte_3DFBB == 0) { 印 "With what?";    return }   ; ★ 沒抓鉤
if (載具 != 0x1C)    { 印 "On foot!";      return }   ; ★ 只有步行(單一值)
方向 = sub_2B2AC()  ; 取消就結束
tile = 目標格
if (tile == 0x0D)    { 印 "Impassable!";   return }   ; ★ 峭壁
if (tile != 0x0C)    { 印 "Not climbable!"; return }  ; ★ 只有群山能爬
for (每個沒死的隊員)
    if (rand(1, 30) > 敏捷) { 印 "Fell!"; 受 rand(1, 5) 傷 }
sub_2D014(dx, dy)                                     ; ★ 無條件過去
```

兩個地形值都對得上 `look#<tile>`:**0x0C 群山、0x0D 峭壁**。

★ 三個容易寫錯的地方:

1. **峭壁與「不能爬」是兩句不同的話** —— 前者是「過不去」,後者是「這不能爬」。
   合成一句會少掉那個區分。
2. **摔倒不擋移動** —— `sub_2D014` 在迴圈**之後、無條件**執行。
   「有人摔倒就不過去」是很自然的想像,但那不是原版。
3. 載具判的又是 `cmp byte_3E08C, 1Ch`(**單一值**),與食人妖那條同一個寫法。

### ★★ `byte_3DFBB` 是**抓鉤**,而它的位移早就對出來了

`dungeon.go` 原本寫:

```go
// hasRope 回報身上有沒有繩索(原版 `byte_3DFBB`)。
//
// ⚠ 那個位元組在存檔裡的位移還沒對出來,所以先一律當成**沒有**。
func (s *State) hasRope() bool { return false }
```

**兩件事都不對**:

- **位移早就對出來了。** `save.go` 那一段寫著「0x0209..0x020F 是
  `byte_3DFBB`..`byte_3DFC1` 七個單位元組,**兩端都已經釘死**」——
  抓鉤就是第一格 0x0209。這個 `return false` 是典型的**陳舊標記**
  (`rulebook/63`):註解說的前提在 save.go 補完那一段之後就不成立了,
  只是沒回來改。**而寫死的 `false` 讓「用抓鉤從洞爬回上一層」與整個
  大地圖攀爬都永遠不可能發生。**
- **它不是繩索,是抓鉤(Grapple)。** `sub_188C4` 的第一道閘門就是它:
  沒有它,大地圖上按 K 只會得到 `With what?`。U5 的道具清單裡那件東西叫
  Grapple。

已補 `Inventory.Grapple` + `SaveGrappleOffset = 0x0209`,存讀檔兩邊都接上,
`hasRope()` 改成真的看那個欄位。`TestGrappleOffsetIsKnown` 同時驗
「七個位元組排滿 0x0209..0x020F」與「`hasRope` 不是寫死 false」。

### 還沒做的:`sub_10A1C` 的墜落

同一批清單裡的 `sub_10A1C` 已經讀完但**還沒實作**,先記在這裡:

```
印 "F-A-L-L-S!!!"
…畫面效果(載具碼暫設 0 再還原 —— 就是 sub_22F0 溺水那一招的「有還原」版)…
for (每個沒死的隊員)
    if (敏捷 <= max(1, rand(0,60)/2)) 受 **1** 點傷      ; 門檻與拖屍怪掙脫同一支
if (X == 0x36 && Y == 0x8A) {                            ; ★ 寫死的座標 (54, 138)
    印 "Falling into underworld!!"
    樓層 = −1;存 BRIT.OOL、載 UNDER.OOL、換地圖 UNDER.DAT
}
```

★ 值得記的兩點:傷害只有 **1 點**(不是骰的),而掉進幽冥界的**座標是寫死的**
—— 世界上只有那一個洞。要接的話得先確認引擎的換層流程能不能從大地圖直接切
`UNDER.DAT`。
