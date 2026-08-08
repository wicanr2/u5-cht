# 86 — tile 0xDC 是開著的月門(而 0xDC 也是龍)

| | |
|---|---|
| 輸入檔 | `re_work/fmtowns/WORRIORS.EXP.asm` |
| 主要函式 | `sub_DEE4`(開關月門)、`sub_DE74`(這一顆算不算在場)、`sub_E084`(玩家踏上)、`sub_2870`(怪物踏上) |
| 起因 | `docs/re/85` 留的 ⬜:「走到 tile 0xDC 就消失,那是什麼地形?」 |
| 狀態 | ✅ 解完,順帶結案 WORKLIST 那條「埋下去的月石怎麼變成月門」 |

---

## 1. 答案:tile 0xDC 是**開著的月門**

三個獨立佐證:

1. **`look#220` 的譯文就是「月亮門!」** —— 220 = 0xDC,而地圖每格 1 byte
   ⇒ 地圖 tile 的範圍是 0..255,所以 `look#220` 就是它。
2. **`sub_E084`(玩家踏上月門)比的是同一個值**:先 `sub_DB10(隊伍X, 隊伍Y)`
   再 `cmp byte ptr [eax], 0DCh`。
3. **`sub_DEE4` 每次重畫都把這個值寫進去**(見 §2)。

⇒ `sub_2870` 尾段「新格 tile == 0xDC 就把種類碼與 tile 歸零」的意思是:
**怪物走進開著的月門,跟玩家一樣被捲走了。**

## 2. ★★★ 月門長在「月石被埋的地方」

`sub_DEE4`(由 `sub_29D64` 重畫時呼叫):

```
esi = 0DCh                                        ; 預設寫「開著的月門」
if (小時 >= 14h || 小時 < 5) {                     ; ★ 夜裡 20:00–04:59
    sub_2BBB8(&byte_3E097, 1, 10h)                ; 計數器累加到 0x10
} else {                                           ; 白天
    sub_2BBFC(&byte_3E097, 1)                     ; 計數器遞減
    if (byte_3E097 == 0) esi = 5                  ; ★ 歸零才寫回草地
}
for (i = 0; i < 8; i++) {
    if (!sub_DE74(i)) continue
    ptr = tile_at(byte_3E040[i], byte_3E048[i])   ; ★★ 月石的 X / Y
    if (*ptr != esi) 重算光照(sub_2E21C)
    *ptr = esi
    byte_4198A |= 2
}
```

`byte_3E040` / `byte_3E048` 就是**埋下去的月石的 X / Y**
(存檔位移 `SaveMoonstoneXOffset` = 0x028A、`YOffset` = 0x0292,`docs/re/71`)。

⇒ **月門就長在月石被埋的座標上。** 這是「埋月石有什麼用」的完整答案,
也結案了 WORKLIST 那條「埋下去的月石怎麼變成月門沒追(`sub_E084` 讀的是
另一組表)」—— 它讀的不是另一組表,是**同一組**;差別在 `sub_DEE4` 先把
tile 寫進地圖,`sub_E084` 才去讀那一格。

### `sub_DE74(i)` 完全不看月相

```
if (byte_3E0A3 != byte_3E050[i]) return 0     ; 地點不符(月石埋在哪個地點)
if (byte_3E058[i] != byte_3E0A5) return 0     ; 樓層不符
if (地點 != 0) return 1                        ; ★ 場景裡不查範圍
dx = (X − byte_3E0AB) & 0FFh ; dy = (Y − byte_3E0AC) & 0FFh
return (dx < 20h && dy < 20h)                 ; 在 32×32 載入視窗內
```

⇒ 三件事各由不同的東西決定,**不要混**:

| 問題 | 由什麼決定 |
|---|---|
| **開不開** | 時間(夜裡開) |
| **在哪裡** | 月石埋在哪裡 |
| **去哪裡** | 月相(`docs/re/22` 的 `TravelByMoongate`) |

### ★ 白天不是立刻關

計數器從 0x10 遞減,**歸零才**把那一格寫回草地(tile 5)。所以日出之後
月門還會殘留一陣子 —— 那是原版的淡出。

## 3. ⚠⚠ `0xDC` 有兩個意思,而 `sub_2870` 同時用了兩個

```asm
cmp     eax, 0DCh              ; eax = 物件種類碼 → **龍**(免疫地形延遲)
...
cmp     byte ptr [eax], 0DCh   ; eax = 地圖 tile 指標 → **月門**
```

**同一支函式、同一個位元組值、兩個命名空間。** 把其中一個當成另一個會做出
「龍走到哪都會消失」或者「所有怪物都免疫地形」—— 而**兩種錯都不會崩,只會怪**。

★ 一般規則:物件種類碼要 **+256** 才是 tileset 索引(`docs/re/09`);
地圖 tile 直接就是 0..255。所以 `0xDC` 當種類碼指的是 tileset 的 0x1DC,
當地圖 tile 指的是 tileset 的 0xDC —— **不同的兩張圖**。

## 4. 引擎對應

| 原版 | 引擎 | 狀態 |
|---|---|---|
| tile 0xDC | `u5data.MoongateOpenTile` | ✅(原名 `CreatureVanishTile`,語意當時未定) |
| `sub_DEE4` 的時段 | `u5data.MoongateOpenAtHour` | ✅ 接進 `MoongateAt` ⇒ **白天沒有月門** |
| `sub_2870` 的月門吞噬 | `game.(*State).stepObject` | ✅ |
| `sub_DE74` | `game.(*State).moongateWritesHere` | ✅ 含 32×32 視窗檢查(用「隊伍為中心」近似) |
| 把 tile 寫進地圖緩衝 | `game.(*State).RefreshMoongateTiles` | ✅ **月門現在畫得出來** |
| 淡出計數器 `byte_3E097` | `game.State.MoongateFrame` | ✅ 開關生效;⬜ 繪圖還沒用它當動畫格 |

## 5. 落地時的三個決定

### (a) 只留一個真相來源

`EnterMoongateHere` 改成**只讀腳下那一格的 tile**(與 `sub_E084` 一致),
而「月門存在」的唯一來源是 `RefreshMoongateTiles` 寫進地圖的那一格。

⚠ 此前它查的是「座標 + 時段」,而 tile 是另一條路 —— **兩個真相來源遲早會漂**。
順帶刪掉 `MoongateAt`(拆開之後沒有非測試呼叫者,而資料本來就在 `s.Moongates`)。
`TestEnteringReadsTheTileNotTheCoordinates` 用兩個反例釘住它:
埋藏點上是草地 → 不傳送;隨便一格寫上月門 tile → 照樣傳送。

### (b) 節奏是近似,不是原版

原版的計數器**每次重畫**升降一格(`sub_29D64` 一個回合可能被叫好幾次),
引擎改成**每回合一次**(`tick()`)⇒ 淡出從「不到一秒」變成「約 16 回合」。
**這是唯一沒辦法照抄的地方**,已列進 A 階段對 DOSBox 的核對清單。

### (c) 關門寫死草地,會蓋掉原本的地形

原版 `mov esi, 5` 是寫死的:月門關上時那一格一律寫回 **tile 5(草地)**,
不管原本是什麼。⇒ 把月石埋在沙漠或雪地,門關上之後那一格會變成草地。
照原樣做(`CLAUDE.md §3.0` 不自行修正),而且它可觀察 —— 值得列進 A 階段。

⚠ 另一個可觀察的後果來自 `sub_DE74` 的視窗檢查:**離開視窗時原版不會把那一格
寫回草地**,所以遠處的月門 tile 會留在地圖上直到玩家再走近。
順序保護了它(回到視窗內時先重寫,才可能踏上去),所以不是 bug。
