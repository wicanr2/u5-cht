# 85 — 怪物怎麼移動,以及地形怎麼拖慢牠們

| | |
|---|---|
| 輸入檔 | `re_work/fmtowns/WORRIORS.EXP.asm` / `.i64` |
| 主要函式 | `sub_2B24`(追人)、`sub_2A54`(隨機遊走)、`sub_2870`(真的挪一格)、`sub_2778`(進得去嗎)、`sub_27CC`(剛空出來的格子) |
| 起因 | `docs/re/84` §1 找到 `sub_2D38` 的尾巴是 `call sub_2B24` |
| 狀態 | ✅ 全部實作(`internal/game/wandermove.go`) |

---

## 1. `sub_2B24` —— 往隊伍走一格,走不通就改遊走

```
dx = 帶號環繞(物件X − 隊伍X);dy 同理
stepX = −sign(dx);stepY = −sign(dy)          ; 往隊伍靠
if (random(0,1) == 1) 先試 X 再試 Y  else  先試 Y 再試 X
每一試:if (sub_2778(槽, 目標) && sub_27CC(目標)) → sub_2870(槽, 步); 結束
兩軸都不行 → ★ sub_2A54(槽)                  ; 退回隨機遊走
```

★ **先試哪一軸是擲出來的。** 少了這一擲,怪物會沿固定的階梯路線接近,
看起來像在走格線而不是在追人。

★★ **最後那一行是原版不會讓怪物卡在牆邊發呆的原因。**

## 2. `sub_2A54` —— 隨機遊走

最多試 **3 次**,每次 `random(0,3)` 選四方向之一,進得去就走。
⚠ 它**只查 `sub_2778`**(地形 + 佔位),**不查 `sub_27CC`** —— 與追人不同。

## 3. `sub_2870` —— 真的挪一格,而地形會拖慢

```
if ((kind & 0xFC) == 0x2C) 敵船 → 依方向換朝向圖(北東南西),不吃地形延遲
else if (kind ∈ {0xDC 龍, 0x94 蝙蝠, 0xD8 惡魔, 0xF0 Mongbat}) → 不吃地形延遲
else switch (目標格 tile) {
    4, 6, 7, 8, 30, 31   → random(0,1) == 0 才走      ; 沼澤 / 灌木 / 焦灼 / 荒漠
    9..15                → random(0,2) == 2 才走      ; 林 / 熱帶林 / 丘 / 山 / 高峰
    其餘                 → 直接走
}
走:byte_4FD94/95 = 舊座標;寫新座標;byte_4198A |= 2
   ★ 若新格 tile == 0xDC → 種類碼與 tile 都歸零(整個消失)
```

### ★★ 同一組地形,玩家付回合、怪物付機率

拖慢怪物的地形(4/6/7/8 與 9–15)**正好是玩家地形代價表的同兩級**
(`docs/re/38`:4/6/7/8 是 1 級、9–15 是 2 級)。但機制不同:
**玩家多跑 1–2 個世界回合,怪物是擲 1/2 或 1/3 決定這一步走不走得成。**
兩邊不要合併。

### ★ 免疫的四種都會飛

龍、蝙蝠、惡魔、Mongbat —— 沼澤與山地拖不慢飛行生物。這不是巧合。

⬜ `tile == 0xDC` 走上去就消失,那是什麼地形還沒查(要抽 FM Towns
`EGA*.TIL` 才看得到圖)。照編號實作、語意留白。

## 4. `sub_27CC` —— 別補進剛空出來的格子

```
return !(x == byte_4FD94 && y == byte_4FD95)
```

那對座標是**全域一份**(不是每槽一份),只由 `sub_2870` 在移動前寫、
只由這裡讀。世界回合倒著掃槽,所以它擋的是「這一輪剛移動過的那一個
留下的空格」。⇒ 兩隻怪不會在同一輪裡接力補位。

## 5. 引擎對應

| 原版 | 引擎 |
|---|---|
| `sub_2B24` | `game.(*State).chaseParty` |
| `sub_2A54` | `game.(*State).wanderRandomly` |
| `sub_2870` | `game.(*State).stepObject` + `terrainLetsCreatureThrough` |
| `sub_2778` | `game.(*State).objectCanEnter` |
| `sub_27CC` | `game.(*State).notJustVacated` |
| `sub_2D38` | `game.(*State).objectMoveGate` |
| `dword_4FD50` | `game.(*State).enemyShipSpeed`(★ 值越大越快) |
| `sub_2D2D0` 的節奏 | `game.(*State).sailRhythm` |
| 物件槽 +7 | `u5data.ObjShipTick` |

⚠ **`sub_2D2D0` 的節奏有一條反直覺而我照原樣做了**:`n % 3` ⇒
同向 1、反向 2、**垂直 0**。側風反而不多花時間。組語就是 `idiv 3` 取餘數,
`CLAUDE.md §3.0` 不自行平衡 —— 已列進 A 階段對 DOSBox 的核對清單。
