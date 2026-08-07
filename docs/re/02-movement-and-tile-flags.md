# RE-02:移動與 tile 通行規則(已破解並接進引擎)

> 日期:2026-08-07 ・ 來源:FM Towns `WORRIORS.EXP`(SHA-256 見 `docs/re/00`)
> 反編譯輸出:61,364 行 C / 1,225 函式(`tools/ida.sh` + Hex-Rays)

## 怎麼找到的:用「移動被擋」的訊息當錨

WORKLIST 的第一優先是通行規則 —— 沒有它,玩家可以穿牆過海。
先在 `DATA.OVL` 找 tile 屬性表,掃過整個 48 KB **沒有命中**(掃到的都是文字區)。

改用字串錨定,在反編譯輸出裡找移動相關訊息:

```
"Blocked!\n"        "Slow progress!\n"     "Very slow!\n"      "Can't!\n"
```

⇒ 原版的地形不只「能不能走」,還有**移動成本分級**(正常 / Slow progress / Very slow)。

`"Blocked!"` 在 `sub_86C` —— 那就是移動處理函式。

## `sub_86C`:移動主流程

```c
switch (dir) {
  case 1: dx = -1; ... print("West\n");   break;
  case 2: dx = +1; ... print("East\n");   break;
  case 3: dy = -1; ... print("North\n");  break;
  case 4: dy = +1; ... print("South\n");  break;
}
v8 = byte_3F789[32 * dy + dx];          // ← 場景地圖緩衝,**列寬 32**
v2 = sub_2B360(...);
if (v7 && sub_2A694(byte_3E08C, v8)) { ... }   // ← 通行判定(第一參數是載具,不是 0)
```

順帶確認一件事實:**場景地圖的列寬是 32** ——
這印證了 `TOWNE/CASTLE/KEEP/DWELLING.DAT` 各 16,384 B = **16 個 32×32 地圖**的推測。

另外這個函式裡有 `"Dost thou wish to leave? "` → `"Exit to Britannia!"`,
就是離開場景回世界地圖的流程。

## `sub_2A694`:通行判定的分派器

```c
// ⚠ Hex-Rays 把第一參數印成常數 0。組語是 `movzx eax, byte_3E08C; push eax`——
// 傳的是**隊伍當前的載具 tile**(步行 / 馬 / 飛毯 / 小船 / 船各有不同的通行規則)。
// 又一次 CLAUDE.md §4.4:反編譯的常數不可信,尤其在 const memory 警告滿天飛的時候。
int sub_2A694(int mover, int tile) {
  switch (byte_5FF8C[mover >> 2]) {          // 移動者 → 移動模式(0–10)
    case 0: return sub_2A610(mover, tile);   // 一般行走
    case 1: return sub_2A674(tile);          // 只能在水上
    case 2: return (tile & 0xF0) == 0x60 || sub_2A674(tile) || sub_2A610(mover, tile);  // 水陸兩棲
    case 3: return sub_2A610(mover, tile) && tile != 143 && tile != 4;
    case 4: return sub_2A674(tile) == 0;     // 只能在陸上
    case 5: /* 方向性通行:用 byte_5FFA8[tile] / byte_5FF6C[tile] 的方向 bit */
    case 6: return tile <= 2;
    case 7..10: return tile == 4 / 5 / 1 / 7;
  }
}
```

`8 >> (mover & 3)` 這種寫法表示 **case 5 的表是「四個方向各一個 bit」** ——
那是給船之類「只能從某些方向進出」的地形用的(待進一步確認)。

## 兩個底層判定(已完全還原)

```c
// 是不是水
int sub_2A674(int tile) { return tile < 4 || (tile & 0xF0) == 0x60; }

// 一般行走能不能過
BOOL sub_2A610(char mover, int tile) {
  BOOL ok = ((128 >> (tile & 7)) & byte_5FF6C[tile >> 3]) == 0;   // bit = 1 → 阻擋
  if ((mover & 0xFE) != 0x1C && (mover & 0xF0) != 0x40 && (tile & 0xFC) == 0x90)
    return false;                                                  // tile 0x90–0x93 限特定移動者
  return ok;
}
```

**`byte_5FF6C` 是每 tile 一個 bit 的阻擋 bitmap**,512 tile ÷ 8 = **64 byte**。
記憶體位址 `0x5FF6C` → 檔案位移 `0x5FF6C + 0x200`(P3 的 image offset)。

### 可驗證的預測 → 命中

抽出來之前先預測:tile 1/2/3 是水,應該被標成阻擋 ⇒ `byte[0]` 的 bit6/5/4 = 1。

實測 `byte[0] = 0x70 = 01110000` —— **tile 1/2/3 阻擋 ✓,而 tile 0 不阻擋**。
比預測更精確:tile 0 在 tileset 上是綠色星芒(特效,不是地形),本來就不該擋。

全表統計:**阻擋 195 / 512,可走 317**。

## 接進引擎的方式

表寫進 `internal/u5data/tileflags.go`,判定式照抄反編譯結果。

**為什麼把表寫進程式碼**:這是**遊戲規則**(等同「這把劍傷害多少」),不是美術或音樂素材。
重寫引擎本來就要還原規則;規則若不入庫,引擎就得依賴玩家手上剛好有 FM Towns 版。
原版資料檔本身仍然一律不散布(`CLAUDE.md §3.0`)。
`tools/dump_tile_flags.py` 可從自備執行檔重新產生本表核對,測試 `TestTileFlagsMatchOriginal`
會逐 byte 對回原版。

### 測試抓到的真 bug

第一版把函式簽名寫成 `TileBlocksWalking(tile byte)` —— **byte 放不下 512 個 tile**,
測試迴圈的 `byte(i)` 在 i ≥ 256 時 wrap,於是前 256 個被算了兩次(阻擋數變成 364)。
改成 `int` 並加邊界檢查(超出範圍當阻擋,比讓玩家走進未定義區域安全)。

順帶確認:`BRIT.DAT` 的值域是 **1..212** ⇒ 世界地圖每格 1 byte 只用得到 tile 0–255,
**tile 256–511 推測是 sprite(NPC/怪物/物件)**,待 P3 確認。

## 還沒做的

| 項目 | 線索 |
|---|---|
| 移動成本分級 | `"Slow progress!"` / `"Very slow!"` 在 `sub_2D0BC`,尚未讀 |
| 移動者 → 模式表 | `byte_5FF8C[mover >> 2]`,共 11 種模式;各模式對應什麼載具待確認 |
| 方向性通行(case 5) | `byte_5FFA8` / `byte_5FF6C` 的方向 bit |
| tile 0 為何被歸進「水」 | `sub_2A674` 的 `tile < 4` 把 0 併進來。**視覺上 tile 0 根本不是水**(算繪出來是一團紅黃爆裂圖案,tile 1–3 才是藍色水面),所以這不是「水的第 4 張動畫」而是把 0 當哨兵值處理 —— 多處程式在越界時回傳 0。**照抄行為,不自行「修正」** |
| tile 語意表(森林/沼澤/山…) | 目前只知道通行與否,不知道各 tile 是什麼地形 |
