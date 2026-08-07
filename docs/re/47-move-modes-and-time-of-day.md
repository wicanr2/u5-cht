# 移動者→模式表,與一個藏在中文化裡的缺陷

輸入檔:`WORRIORS.EXP`(FM Towns 英文版)、`gamedata/DATA.OVL`
IDA 位址:`sub_2A694`(通行判定分派)、`sub_2A610`(阻擋位圖)、
`sub_2A674`(是不是水)、`sub_10FEC`(商店對白的 token 代換)
日期:2026-08-08

---

## 1. `byte_5FF8C` 的 64 項

`docs/re/02` 早就記下了分派器的形狀,但**表的內容當時標「待確認」**。
從 `WORRIORS.EXP` 線性 0x5FF8C(檔案位移 = 線性 + 0x200)抽出來:

```
mover>>2    模式
0x00..0x0F   0    一般陸行
0x10..0x13   3    馬
0x14..0x17   2    魔毯(兩棲)
0x18..0x1F   0    步行
0x20..0x23   6    揚帆中
0x24..0x27   6    大船
0x28..0x2B   5    小艇
0x2C..0x2F   6    船的另一組朝向
0x80..0x8F   1    只走水(水生怪)
0x9C..0x9F   4    飛行
0xE0..0xE3  10    ⬜
0xE8..0xEB  0xFF  → 落在 switch 外,走 default
0xEC..0xEF   9    ⬜
0xF4..0xF7   8    ⬜
0xF8..0xFB   7    ⬜
0xFC..0xFF   4    飛行
```

### 為什麼是 `mover >> 2`

載具碼是「四個朝向一組」(大船 0x24..0x27、小艇 0x28..0x2B、馬 0x10..0x13),
除以 4 就把一組收成一格。

### 六個錨點同時對上

這些常數**全部是別處獨立推出來的**(買馬 `sub_118CC`、上下載具
`sub_16F08`/`sub_177AC`、風向那幾支),與 `byte_5FF8C` 毫無關係:

| 索引 | mover | 表上的模式 | 獨立來源 |
|---|---|---|---|
| 4 | 0x10..0x13 | 3 | `TileHorse` 0x10、馬的載具碼 0x12/0x13 |
| 5 | 0x14..0x17 | 2 | `VehicleCarpet` 0x14 —— 兩棲,對得上「飛過水」 |
| 7 | 0x1C..0x1F | 0 | `VehicleWalk` 0x1C |
| 8 | 0x20..0x23 | 6 | `VehicleSailing` 0x20 |
| 9 | 0x24..0x27 | 6 | `VehicleShip` 0x24 |
| 10 | 0x28..0x2B | 5 | `VehicleSkiff` 0x28 |

六個一起中,表的對齊沒有滑動的餘地(`rulebook/62`)。

## 2. 七種模式的判定

兩個基礎判定:

```
sub_2A674(tile)   是不是水: tile < 4 || (tile & 0xF0) == 0x60
sub_2A610(m,tile) 阻擋位圖: byte_5FF6C[tile>>3] 的 bit (0x80 >> (tile&7))
```

| 模式 | 判定 |
|---|---|
| 0 一般陸行 | 只看阻擋位圖 |
| 1 只走水 | 必須是水 |
| 2 兩棲 | 水可以,陸上照位圖 |
| 3 坐騎 | 不能下水,陸上照位圖 |
| 4 飛行 | 位圖與水都不管 |
| 5 小艇 | 只走水(**淺灘可以**) |
| 6 大船 | **只走深水** —— 淺灘(tile 3)會擱淺 |

★ 5 與 6 的差別是 U5 航海的核心限制:**大船靠不了岸,得放小艇。**
兩種都寫成「只要是水就行」的話,玩家可以開大船直接撞上沙灘。

⬜ 模式 7–10 的判定式還沒逐條核完。它們對應的都是 0xE0 之後的特殊移動者
(暗影君主、旋風那類),引擎先當兩棲並標 TODO —— **標 TODO 而不是假裝確定**。

## 3. 時段:早就做了,但中文化漏了一半

`sub_10FEC` 的 `@` token:

```asm
cmp     byte_3E08F, 0Ch     ; hour < 12
push    offset aMorning     ; "morning"
cmp     byte_3E08F, 12h     ; hour < 18
push    offset aAfternoon   ; "afternoon"
push    offset aEvening     ; "evening"
```

`u5data.TimeOfDay` 早就實作了,邊界也對,而且接在商店對白的 `@` 上。
WORKLIST 標 ⬜ 是過期記載。

### 但它回的是英文

譯好的店家對白長這樣:

```
"@好!我是 $,# 的老闆。有什麼能為你效勞的?」"
```

`@` 代入 `TimeOfDay(hour)` = `"morning"` → 玩家看到「**morning好!**」。

★ **譯文完全正確**(18 行原文有 `@`、18 行譯文都保留了),
壞的是代入的字。這種缺陷在譯文檢查裡**完全看不出來** ——
`checktalk` / `audittalk` 檢查的是譯文本身,而問題在代換的那一端。

修法:`Placeholders.TimeWord` 覆寫,由 `i18n.TimeOfDay` 提供「早 / 午 / 晚」。
英文原文那條路不受影響(留空就用原版的英文)。

⚠ 中文邊界要**照原版**:「下午」直覺上從 13 時起、「晚上」直覺上更晚,
但原版是 12 時與 18 時。照直覺重畫的話店家在 12 時會說「早好!」,
而那種錯只有拿原版並排才看得出來。

## 4. 引擎對應

| 原版 | 引擎 |
|---|---|
| `byte_5FF8C` | `u5data.moverMode` / `ModeOf` |
| `jpt_2A6B4` 的 11 case | `u5data.MoveModeAllows` |
| `sub_2A694` | `u5data.MoverCanEnter` |
| `sub_10FEC` 的 `@` | `u5data.TimeOfDay` + `i18n.TimeOfDay`(中文) |

## 5. 又一次「先 grep 再寫」

我在 `movemode.go` 裡重寫了一份 `TileIsWater`,`vet` 才擋下來 ——
`tileflags.go` 早就有了。這是第三次(前兩次是 `State.Mix`、`hasAnyReagent`)。
