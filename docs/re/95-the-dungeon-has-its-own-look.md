# 95 — 地牢有自己的 Look,而噴泉真的能喝

| | |
|---|---|
| 輸入檔 | `re_work/fmtowns/WORRIORS.EXP.asm` / `.i64` |
| SHA-256 | 見 `re_work/fmtowns/SHA256SUMS.txt` |
| 主要函式 | ★★ `sub_EEEC`(地牢 Look)、★ `sub_13BA8`(第一人稱方向提問)、`sub_13B4C`(相對方向 → 座標)、★ `sub_E19C`(選隊員)、★ `sub_2B67C`(找第一個醒著的) |
| 主要資料 | `dword_3E16C`(地牢層緩衝)、`byte_3EE15`(朝向)、`byte_3EE16`、★ `byte_3E08B`(當前選定的隊員)、`word_3E086`/`word_3E088`(**通用暫存**) |
| 起因 | 「把剩下的遊戲邏輯補齊」—— 未讀函式按「遊戲邏輯訊號」排序後,`sub_EEEC` 是隊首(372 行、34 個字串引用) |
| 狀態 | ✅ 全解並落地;順手收掉 `docs/re/37` 的一條懸案 |

---

## 0. ★★ 地牢的 Look 是另一支程式

原版 `sub_2ACF4` 的 `L` 依地點碼分派:地牢走 `sub_EEEC`,其餘走 `sub_D258`。
兩者差在四處:

| | 地面 `sub_D258` | 地牢 `sub_EEEC` |
|---|---|---|
| 方向 | 絕對(東南西北)| ★ **相對**(前 / 左 / 右 / **腳下**)|
| 沒光源 | 照樣看得到 | ★ 直接印「一片漆黑」,連地形都不報 |
| 描述來源 | `LOOK2.DAT` 的家具語 | ★ 程式裡寫死的**十六種地牢地形** |
| 互動 | 招牌、水晶球、許願井 | ★★ **噴泉會問「要喝嗎」** |

## 1. ★ `sub_13BA8` —— 第一人稱的方向提問

```
sub_13BA8(朝向)  →  1 = 選了方向、0 = 取消
  印 "Dir-"; 等鍵(只收 空白 與鍵碼 1..4)
  空白  → 印 "Pass\n";  return 0
  3(↑) → 印 "Ahead\n"; sub_13B4C(朝向,          隊伍X, 隊伍Y)
  4(↓) → 印 "Here\n";  word_3E086 = 隊伍X; word_3E088 = 隊伍Y
  2(→) → 印 "Right\n"; sub_13B4C((朝向 + 1) & 3, 隊伍X, 隊伍Y)
  1(←) → 印 "Left\n";  sub_13B4C((朝向 + 3) & 3, 隊伍X, 隊伍Y)
  return 1
```

★ **「腳下」那一支不呼叫 `sub_13B4C`**,直接把隊伍座標寫進暫存 ——
所以它不經過地牢的環繞處理(行為等價,但結構不同)。

★ 左轉那段是 `lea esi,[edi+3]; and esi,3` 加一個負值修正
(`if (eax < 0 && esi != 0) esi -= 4`)。朝向是 0..3 時修正走不到,是死碼。

⚠ 鍵碼 1..4 就是原版方向鍵的碼(1 西 / 2 東 / 3 北 / 4 南,`docs/re/49`),
但在第一人稱下**語意被重新定義**:↑ 前、↓ 腳下、← 左、→ 右。
引擎的地牢 Search 早就這樣做了,Look 共用同一條。

## 2. `sub_EEEC` 全文

```
方向 = sub_E19C()                      ; ★ 先選隊員(見 §3)
if (方向 == -1) return
if (火把 == 0 && 光明 == 0) → "You see:\ndarkness."; return
if (sub_13BA8(byte_3EE15) == 0) return  ; 問方向,0 = 取消
tile = dword_3E16C[樓層*64 + (Y & 7)*8 + (X & 7)]
印 "You see:\n"
if (tile == 0x61) tile = 0              ; ★ 特例:'a' → 當成通道

高四位 = tile & 0xF0
if (高四位 == 0x80) {                   ; ★ 力場先攔
   switch (tile − 0x80) { 0 睡眠場 / 1 毒氣場 / 2 火牆 / 3 電場 / default 能量場 }
} else if (高四位 == 0xC0) {            ; ★ 崩落通道次攔
   n = byte_3EE16 & 0x0F
   if      (n == 1) "a dripping stalactite."
   else if (n == 2) "a caved in passage."
   else if (random(1, 0xFF) == 0xFF) "an unfortunate software pirate."   ; ★ 1/255
   else "a less fortunate adventurer."
} else 查十六路跳表
if (高四位 == 0x50) 問「Will you drink?」   ; ★★ 噴泉
```

### ★ 十六路跳表有兩格取不到

| 索引 | 原版字串 | 為什麼取不到 |
|---|---|---|
| 0x8 | `"an energy field."` | 0x80 已被力場那四路 switch 攔掉 |
| 0xC | `"SPEC WALL ERR."` | 0xC0 已被崩落通道那一支攔掉 |

`SPEC WALL ERR.` 明顯是作者留著的除錯字串。

⚠⚠ **索引 8 的字串與力場 default 的字完全相同**(`aAnEnergyField` vs
`aAnEnergyField_0`)—— 所以「有沒有走到死格」**不能用字串比對驗**。
引擎在那兩格填哨兵值(`dungeonLookUnreachable`),原文另存在
`dungeonLookDeadArms` 供對碼。這是「同一個觀察可以支持兩個模型 ⇒
要用能區分它們的檢查」的又一次。

### ★ 1/255 的反盜版彩蛋

`random(1, 0xFF) == 0xFF` —— 254/255 的機率印「a less fortunate adventurer.」,
1/255 印「**an unfortunate software pirate.**」。U5 有名的玩笑。

## 3. ★ `sub_E19C` 是「選隊員」,不是「問方向」

```
if (地點碼 > 0x80)            esi = dword_3EF50[byte_3E0AE*8 + 3]  ; 戰鬥:當前單位
else if (byte_3E08B != 0xFF)  esi = byte_3E08B                     ; ★ 已經選好的人
else {
   count = 0
   for (i = 0; i < 隊伍人數; i++)
      if (狀態 == 'G' || 狀態 == 'P') { esi = i; count++ }          ; ★ 一路覆寫
   if (count <= 1) return esi                                       ; 只有一個 → 不問
   ... 問玩家 ...
}
```

★ 迴圈把 `esi` 一路覆寫,所以自動選中的是**最後一個**符合的,不是第一個。
引擎的 `pickCharacter` 早就這樣寫了(我第一版的 `pickDrinker` 取第一個,錯了,已刪)。

### ✅ `byte_3E08B` = 當前選定的隊員(0xFF = 未選)

全檔掃描:**25 支函式讀寫**,而且 `sub_27D24`(讀檔)與 `sub_284CC`(寫檔)
都碰它 ⇒ **它進存檔**。許多指令用完就寫回 0xFF(清掉選擇)。

⇒ 這一條同時解釋了 `docs/re/93` 標 ⬜ 的那個「副作用」:
`sub_20F24`(酒館的 sir/milady)寫 `byte_3E08B = 0xFF` **不是副作用**,
是**刻意清掉選擇**。

⬜ 引擎沒有這個跨指令狀態(每次重掃)。差別只在「多人可選時原版會問、
引擎取最後一個」。要補的話它同時是一個**存檔欄位**,位移待定。

### ✅ 順手收掉 `docs/re/37` 的懸案

`docs/re/37` 記著白天看太陽那條路有一行「把隊伍的 x 座標寫進成員索引,
看起來像原版的 bug」,並說「要收掉這條得先逆 `sub_2B67C` 與
`byte_3E08B` 的全部寫入點」—— 兩件都做了:

```
sub_2B67C():                                     ; 找第一個醒著的
   for (i = 0; i < 隊伍人數; i++) {
      if (狀態 == 'G' || 'P') { word_3E086 = i; return 0 }   ; ★ 用暫存當回傳通道
      if (狀態 == 'S') 睡著數++
   }
   return 睡著數 ? 1 : -1
```

⇒ **不是 bug。** 錯因是把 `word_3E086` 當成「只可能是 x 座標」——
它是**通用暫存全域**(36 處寫、164 處讀,語意由呼叫端決定),
而 `docs/re/16` 早就記過同一個坑。

⇒ 規則:**看到一個全域被大量讀寫且語意不一致,先假設它是暫存,不是具名欄位。**

## 4. ★★ 噴泉能喝,而效果看**完整 tile 值**

```
印 "Will you drink?\n";  等鍵
N     → 印 "No.\n"
非 Y  → 重問                          ; ★ 不是當成 N
Y     → 印 "Yes.\tGulp!\n"
switch (tile) {                        ; ★★ 完整值,不是高四位
  0x50 → "Cured!\n"    隊員[選中].狀態 = 'G'
  0x51 → "Healed!\n"   隊員[選中].HP   = MaxHP
  0x52 → "Poisoned!\n" 隊員[選中].狀態 = 'P'
  default → "Bad taste.\n"  sub_2A464(選中, random(0, 7))
}
```

三件容易漏的:

1. **問要不要喝看高四位**(任何 0x5x),**喝的效果看完整值** ——
   0x53..0x5F 全部落到「難喝」。這是全支函式唯一看低位元的地方。
2. **「難喝」扣 0..7 血** —— 下限是 0,所以有 1/8 的機率一點血都不掉。
3. **補血那兩行沒有任何前置判斷** ⇒ 死人的 HP 也會被填滿(狀態仍是 'D')。
   照抄,不「順手」加判斷。

⚠ 與地面的噴泉(`sub_CE78`,`docs/re/37`)**完全不同** ——
那一支只印三句話、**什麼也不做**。兩支不能合併。

## 5. 引擎落地

| 原版 | 引擎 |
|---|---|
| `sub_2ACF4` 的 'L' 分派 | `(*State).Look` 開頭 `if s.InDungeon()` |
| `sub_EEEC` | `LookDungeon` / `lookDungeonRelative` / `dungeonLookDesc` |
| `sub_13BA8` / `sub_13B4C` | `dungeonRelativeSquare` + 既有的 `SearchRelative` |
| 十六路跳表 | `dungeonLookText`(死格填哨兵)+ `dungeonLookDeadArms` |
| 力場四種 + default | `dungeonFieldText` + `DungeonFieldEnergy` |
| `random(1,0xFF) == 0xFF` | `PirateOdds` |
| 噴泉喝水 | `beginDungeonDrink` / `drinkFromDungeonFountain` |
| `sub_E19C` | 既有的 `pickCharacter`(回**最後一個**符合的)|

### 測試

| 測試 | 驗什麼 |
|---|---|
| `TestDungeonLookIsADifferentProgram` | 地牢的 L 走另一支,↓ = 腳下 |
| `TestDungeonLookNeedsLight` | 沒光源一片漆黑;**含正對照**(點火把就看得到)|
| `TestDungeonLookFieldsComeBeforeTheJumpTable` | 分派順序;**含反對照**(哨兵真的在表裡)|
| `TestCavedPassageHasThreeOutcomes` | 石筍 / 崩落 / 1-255 彩蛋(掃 400 顆種子)|
| `TestFountainDrinkEffectsUseTheFullTile` | 四種效果由完整 tile 決定 |
| `TestFountainHealFillsToMaxRegardlessOfStatus` | 補血不看死活 |
| `TestFountainRefusalDoesNothing` | 答 N 什麼都不發生 |
| `TestDungeonLookOffersTheDrinkOnlyAtFountains` | 只有 0x5x 會問 |
