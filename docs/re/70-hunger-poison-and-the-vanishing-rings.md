# 70 — 隊伍原本永遠不會餓,中毒也不會扣血

| | |
|---|---|
| 輸入檔 | `re_work/fmtowns/WORRIORS.EXP.asm`(FM Towns 英文版) |
| SHA-256 | 見 `re_work/fmtowns/SHA256SUMS.txt` |
| 主要函式 | `sub_2A50C`(每回合維生開銷)、`sub_2EE84`(開戰佈陣)、`sub_2A464`(扣一人的血)、`sub_2A4D0`(全隊傷害) |
| 落地 | `internal/game/upkeep.go`、`internal/u5data/items.go` |
| 從哪裡來 | `docs/re/66` 的截斷清單:`sub_2A50C` 113 行 → 4 行 C(`'Starving!'` 掉了)、`sub_2EE84` 214 行 → 25 行(`'A ring has vanished!'` 掉了) |

---

## 1. 症狀:買了糧食用不掉

引擎裡 `Inventory.Food` 只有**一處**會減 —— 「小妖偷糧」那個怪物特殊攻擊。
酒館買的乾糧、田裡採的作物、桌上拿的食物全部只進不出,而**中毒的狀態掛著也不痛**。

`sub_2A50C` 就是那個開銷,而它反編譯只剩四行。

---

## 2. `sub_2A50C`

```
for (每個隊員 i) {
    狀態 = byte_3DDBF[i*32]
    if (狀態 == 'D' && byte_3E08B == i) byte_3E08B = 0FFh   ; 死了就取消「被選中」
    if (狀態 == 'D' || 狀態 == 'S') continue
    if (狀態 == 'P') sub_2A464(i, 1)                        ; ★ 中毒每回合 1 血
    活人數++                                                ; ★ 睡著的不算
}
if (byte_3E08F == byte_3E090) return                        ; 這個小時已經結算過
if (word_3DFB4 == 0) {                                      ; ★ 存糧 0
    印 "Starving!"
    sub_2A4D0()                                             ; 全隊 rand(1, 8) 傷
} else if (byte_3E08F ∈ {6, 12, 18}) {                      ; ★ 6 / 12 / 18 點
    存糧 -= 活人數
    byte_3E090 = byte_3E08F                                 ; 記下「這個小時結算過了」
    byte_3E09B += 1(上限 255)
}
```

`byte_3E08F` 是**小時**(`docs/re/22` 的月門判定 `cmp byte_3E08F, 0Ch` 已經定過)。

### ★ 三個容易寫錯的地方

1. **一天扣三次,不是每回合扣。** `byte_3E090` 記著上一次結算的小時,
   同一個小時內走幾百步也只扣一次糧。
2. **睡著的人不吃** —— `'S'` 那條 `continue` 在 `活人數++` **之前**。死人也不吃。
3. ★ **餓的判定每回合都跑。** 關鍵是**餓的那一條不更新 `byte_3E090`**,
   所以只要存糧是 0,下一回合的「這個小時結算過了嗎」還是不相等,又餓一次。
   **斷糧的隊伍走一步掉一次血** —— 不是一天掉三次。所以餓死是真的會發生。

⚠ `byte_3E09B`(每次進餐 +1,上限 255)用途還沒追,先不接。

---

## 3. ★ 開戰時戒指會消失

`sub_2EE84` 是**開戰佈陣**(清空 32 個單位槽、把隊員排進去)。
迴圈裡夾了一段**與佈陣無關**的判定:

```asm
        cmp     byte_3DDD1[eax], 2Ah
        jnz     short …
        mov     [ebp+var_1], 2Ah          ; 隱形戒指
        …
        cmp     byte_3DDD1[eax], 2Ch
        jnz     short …
        mov     [ebp+var_1], 2Ch          ; 再生戒指
        cmp     [ebp+var_1], 0
        jz      short …
        push    0Fh / push 0
        call    sub_28E14                 ; rand(0, 15)
        cmp     esi, 0Bh
        jnz     short …
        push    offset aARingHasVanish    ; 'A ring has vanished!'
        …
        call    sub_2F35C                 ; 把它從身上拿掉
```

`byte_3DDD1` − 名冊基底 `byte_3DDB4` = **0x1D** = `CharRing` ✓

### 只有兩種戒指會消失

從 `DATA.OVL` 0x1806 的指標表(+0x10 偏移)逐筆解出來:

| 編號 | 名字 | 開戰會消失? |
|---|---|---|
| 0x2A = 42 | Ring of Invisibility 隱形戒指 | **會** |
| 0x2B = 43 | Ring of Protection 防護戒指 | **不會** |
| 0x2C = 44 | Ring of Regeneration 再生戒指 | **會** |

⇒ 兩個會消失的正好是**效果最強的那兩個**(隱形、再生),防護戒指沒事。
這不是隨便挑的 —— 原版只比 0x2A 與 0x2C 兩個值,中間那個跳過去。

⚠ 判定是 **`rand(0, 15) == 11`**,不是 `< 1` 也不是 `== 0`。
機率一樣是 1/16,但**換個等號就不是同一組隨機序列** —— 照原樣。

---

## 4. 落地與驗收

| | |
|---|---|
| `upkeep()` | 掛在 `State.tick()` 裡(`AdvanceTime` 之後、NPC 之前) |
| `State.mealHour` | 原版 `byte_3E090`;初值 −1,零值修正見函式內註記 |
| `MealHours` | `{6, 12, 18}` |
| `vanishRings()` | 由 `placeParty` 呼叫 —— 位置照原版(佈陣的同一趟) |
| `u5data.ItemRing*` | 三只戒指的編號 + 「防護不會消失」的說明 |

七條測試:

- `TestPoisonCostsOneHitPointPerTurn` —— 中毒的掉血、沒中毒的不掉
- `TestFoodIsEatenAtSixTwelveAndEighteen` —— **24 個小時逐一驗**
- `TestSameHourOnlyEatsOnce` —— 同一小時走 200 步只扣一次
- `TestSleepingAndDeadMembersDoNotEat` —— 睡著的與死人都不算人頭
- `TestStarvingHurtsEveryTurn` —— **第二回合還要再餓一次**(擋「順手改成一天三次」)
- `TestOnlyTwoRingsVanish` —— **防護戒指不會消失**
- `TestRingVanishRollIsOneInSixteen` —— 判準是 `== 11`,並用 4000 次抽樣擋機率差一個數量級

### ⚠ 寫測試時踩到的一個坑

一開始那兩條戒指測試**假紅**:每次迴圈都 `upkeepScene(t)` 重建 State,
而新 State 的骰子種子是**固定的 1**(`SeedRandom` 的說明就寫著
「測試不呼叫就是固定種子」)—— 所以每次都拿到同一個骰值,永遠不等於 11。

改成**重用同一個 State** 就對了。這個形狀值得記:
**「每次重建物件再擲一次」在固定種子下等於只擲了一次。**

## 5. 還沒讀的

- `byte_3E09B`(進餐計數器)的用途。
- `byte_3E08B`(「被選中的隊員」)在死亡時被清成 0xFF 的完整語意。
- `sub_2F35C`(把戒指從身上拿掉)只當黑盒用 —— 引擎直接寫 `CharRing = ItemNone`,
  原版是否還動了別的欄位未驗。
- `sub_2A50C` 尾段 `byte_3E09E` 那個倒數(歸零時 `byte_3E08A = 0`)沒追。

---

## 追記:`sub_1318` —— 腳下那一格每回合的作用,而它就是維生開銷的呼叫端

`sub_2A50C` 是誰呼叫的?`sub_1318`(197 行 → 29 行 C,三個字串全掉),
而它的最後一行是:

```asm
        and     ebx, ebx
        jnz     loc_135D          ; ★ ebx != 0 → 跳過維生開銷
        call    sub_2A50C
```

`ebx` 只在**活門**那一條被設成 1。⇒ **掉下活門的那一回合不算吃飯。**

### 四個地形效果

```
for (每個隊員)
    if (狀態 == 'S' && rand(0, 15) == 15) 狀態 = 'G'    ; ★ 睡著的 1/16 自己醒
tile = 腳下那一格
if (tile == 8Ch && (載具 & 0FEh) != 14h) {              ; ★ 鬆動的磚,不在魔毯上
    印 "A TRAPDOOR!";  sub_2A4D0()
    if (地點 == 1Dh) { …填滿熔岩 tile…;  全隊 血=0 狀態='D' }   ; ★★ STONEGATE
    樓層--;  ebx = 1
}
else if (tile == 4 && 載具 == 1Ch) {                    ; ★ 沼澤,而且**步行**
    for (每個隊員) if (狀態 不是 'D'/'P' && rand(0,29) > 敏捷) { 印 "Poisoned!"; 狀態='P' }
}
else if (tile == 0BCh || tile == 8Fh) { 印 "Burning!";  sub_2A4D0() }
```

四個 tile 都用 `look#<tile>` 查得出名字,四個都與效果相符:

| tile | 名字 | 效果 |
|---|---|---|
| 4 | 沼澤 | 中毒 |
| 0x8C | 鬆動的磚 | 活門,掉下一層 |
| 0x8F | 熔岩 | 燃燒 |
| 0xBC | 壁爐 | 燃燒 |

### ★★ STONEGATE 的活門是死路

地點 **29 = STONEGATE**(`locations.go` 第 29 筆)。踩到那裡的鬆動磚,
原版把整張地圖緩衝填滿**熔岩 tile 0x8F**、清空物件表,然後
**逐一把隊員的血設 0、狀態設 'D'** —— 那不是普通陷阱,是死路。

這一條**不能「順手加個存活判定」**。`TestStonegateTrapdoorKillsEveryone`
把它釘住,而別的地點的活門只是掉一層。

### 三個容易寫錯的地方

1. **沼澤是 `rand(0, 29) > 敏捷`(大於才中毒)** —— 敏捷 30 以上完全免疫。
2. **只有步行會中毒**,而載具判的又是**單一值 0x1C**
   (與食人妖、大地圖攀爬同一個寫法 —— 這是第三次)。
3. **睡著的人在戰鬥外每回合 1/16 自己醒**(`rand(0,15) == 15`)。
   這是第三條睡眠醒來規則:戰鬥中怪物 1/17、戰鬥中隊員 1/16 且照樣跳過回合、
   **戰鬥外 1/16 直接醒**(`docs/re/67`)。三條都不一樣。

### 寫測試時踩到的第二個坑

`TestStonegateTrapdoorKillsEveryone` 一開始也假紅:只把 `Location` 設成 29,
`TileAt` 就會去讀**場景圖**,而測試寫的是世界圖 —— 地形根本沒被讀到,
而失敗訊息長得像「全隊沒死」。改成真的 `SetScene(29, …)` 再寫 tile 才對。

**形狀與上面那個固定種子的坑同類:失敗訊息指向結果,原因在前置條件。**

---

## 追記二:酒館的酒**會讓你走不直**

`tavern.go` 原本掛著這句:

```go
case GoodsDrink:
    // 原版只把「這趟喝了幾杯」加一(dword_56E1C),沒有其他效果。
    s.Shop.drinks++
```

**錯的。** 同一支 `sub_21108`(酒單,也在截斷清單上:153 行 → 6 行 C、
17 個字串全掉)還有一行:

```asm
        mov     byte_3E169, 19h        ; ★ 25
```

而 `byte_3E169` 是**讀鍵那一層**在看的:

```asm
; sub_1158 —— 讀一個鍵的包裝
        call    sub_29EEC              ; 真正讀鍵
        xchg    eax, edi
        cmp     byte_3E169, 0
        jz      short 照原樣回傳
        push    1 / push 0
        call    sub_28E14              ; rand(0, 1)
        and     eax, eax
        jz      short 照原樣回傳       ; ★ 一半機率什麼都不做
        call    sub_C10
        dec     byte_3E169             ; ★ 踉蹺了才扣一次
        push    offset aHic            ; 'Hic!'
        call    sub_23C18
        push    3 / push 0
        call    sub_28E14              ; rand(0, 3)
        movzx   edi, byte_4FC54[eax]   ; ★★ 回傳一個隨機**方向鍵**
```

`byte_4FC54` 的前四個位元組是 **`3, 4, 2, 1`** —— 四個方向鍵的鍵碼
(與 `sub_1EFC8` 清單瀏覽器用的 1..4 同一組)。

⇒ **在酒館點一杯酒,接下來有 25 次「按鍵變成隨機走一步」**,
而且是「一半機率發生、發生了才扣一次」—— 平均要按五十次才醒。

★ 這是**輸入層**的效果,不是狀態欄上的一個圖示:玩家按 Z 想看數值,
結果往東走了一格。少了它,酒館的酒就只是「花錢買一句話」。

### 落地

| | |
|---|---|
| `State.Drunk` | 原版 `byte_3E169` |
| `GetDrunk()` | 酒館點酒時呼叫,設 25 |
| `DrunkStagger()` | 1/2 擲、扣一次、印 `Hic!`、回一個隨機方向 |
| `cmd/u5cht/main.go` | 在**字母指令與方向鍵之前**攔一手 —— 攔在這裡才會像原版一樣**所有鍵都會踉蹺** |

三條測試:喝一杯是 25 次、**踉蹺才扣**(按鍵數約兩倍)、四個方向都要出得來。
