# 96 — 搜尋還有三層,而其中一層只在午夜有東西

| | |
|---|---|
| 輸入檔 | `re_work/fmtowns/WORRIORS.EXP`(一手位元組)、`WORRIORS.EXP.asm` / `.i64` |
| SHA-256 | 見 `re_work/fmtowns/SHA256SUMS.txt` |
| 主要函式 | `sub_147A8`(地面 Search 主體)、★ `sub_13FC4`(挖月石)、★★ `sub_14090`(午夜採藥)、★★ `sub_14160`(113 筆固定物品)、★ `sub_13CB0`(物件描述表)、`sub_2B57C`(找物件槽) |
| 主要資料 | ★★ 113 筆 × 六個平行陣列(0x55A64 / 0x55AD8 / 0x55B4C / 0x55BC0 / 0x55C34 / 0x55CA8)、三個採藥點(0x55A4C 起)、`byte_3E06C`(已撿過位元圖)、`byte_3E068`(採藥日期) |
| 工具 | `tools/dump_hidden_items.py` → `internal/u5data/hiddenitems.go` |
| 起因 | 「把剩下的遊戲邏輯補齊」——`sub_13CB0` 在未讀佇列第四位,順著它的呼叫者挖到整條鏈 |
| 狀態 | ✅ 三層全解並落地;收掉 `docs/re/41` §5 的「那兩張表還沒定位」 |

---

## 0. ★★ `sub_147A8` 的尾巴有三層,引擎一層都沒有

引擎此前的 Search 只做到「隨機翻到什麼」(`sub_13DD8` 的金幣 / 食物 / 垃圾 / 瘟疫)。
原版在那之後還有**三層固定內容**,而且是**短路**的:

```
if (tile == 0x4E)  { 印 "a hidden door!"; 地圖寫 0xB8/0xB9; return }   ; 暗門
if (tile == 0xDC)  return                                            ; ★ 月門:連一句都不印
if (sub_13FC4(X, Y, 樓層) != 0) return      ; ① 挖出自己埋的月石
if (sub_14090(X, Y)       != 0) return      ; ② 午夜的秘密採藥點
sub_14160(X, Y)                             ; ③ 113 筆固定物品 / "nothing of note."
```

⚠ 「什麼也沒有」那一句**只在第三層沒中時才印** —— 前兩層落空是靜默的
(它們會落到下一層)。少了這個順序,玩家會在採藥點看到「什麼也沒有」。

## 1. ★ ① 挖出自己埋的月石(`sub_13FC4`)

```
for (i = 7; i >= 0; i--) {                     ; ★ 倒著掃
   if (byte_3E040[i] != X)      continue       ; 月石的 X
   if (byte_3E048[i] != Y)      continue       ;         Y
   if (byte_3E058[i] != 樓層)    continue       ;         樓層
   if (byte_3E050[i] != 地點碼)  continue       ; ★★      埋在哪個地點
   if (那一格已經躺著第 i 顆) continue          ; 否則每搜一次多生一顆
   槽 = sub_2B57C();  sub_2B6C8(0x19, 0x19, X, Y, 樓層, i, 槽)
   印 "a strange rock!\n";  return 1           ; ★ 品質欄 = 月石編號
}
return 0
```

★ 三件事:

1. **倒著掃** ⇒ 同一格埋了兩顆時先挖出**編號大**的。
2. **四個欄位全比**,含**地點碼** —— 少比它的話「埋在城裡 (15,15)」與
   「埋在大地圖 (15,15)」會混在一起。這四個欄位正是 `docs/re/71` 解出的
   月石那四個平行陣列。
3. 印的是「**a strange rock!**」—— 原版**不說那是月石**,撿起來才知道。

## 2. ★★ ② 午夜的秘密採藥點(`sub_14090`)

```
for (i = 0; i < 3; i++) {
   if (表[i].X != X || 表[i].Y != Y) continue
   if (byte_3E08F != 0)              continue    ; ★★ 小時必須是 0(午夜)
   if (byte_3E068[i] == byte_3E08E)  continue    ; ★ 今天已經採過這一點
   byte_3E068[i] = byte_3E08E                    ; 記下今天
   n = random(2, 15);  藥草[表[i].藥草] += n;  夾到 0x63
   印 n " sprigs of\n" <藥草名>;  return 1
}
return 0
```

三筆(一手位元組,線性 0x55A4C 起):

| # | 座標 | 藥草編號 | 原版印的字串 |
|---|---|---|---|
| 0 | (182, 54) | 7 | `mandrake root!` |
| 1 | (97, 165) | 7 | `mandrake root!` |
| 2 | (44, 137) | 6 | `nightshade!` |

★★ **只在午夜採得到**,而 NPC 對話裡就講了「曼德拉草可於彼或亡者沼地採得」——
少了時間這一條,玩家白天去會什麼都沒有**而不知道為什麼**。這是「遊戲有講、
引擎沒做」的那種缺口,比純數值差異嚴重。

★ 記帳是**每點各一份**(`byte_3E068[i]`),不是全域一份 ⇒
同一個午夜可以跑三個點各採一次。

## 3. ★★ ③ 113 筆固定物品(`sub_14160`)

六個平行陣列,基底 `dword_55A48`(IDA 拿旗標當基底的老問題),各間隔 0x74:

| 位移 | 線性 | 內容 |
|---|---|---|
| +0x1C | 0x55A64 | 物件種類(**同時是** `sub_13CB0` 的描述索引)|
| +0x90 | 0x55AD8 | 品質 |
| +0x104 | 0x55B4C | 地點碼 |
| +0x178 | 0x55BC0 | 樓層(0xFF = −1 = 地下世界)|
| +0x1EC | 0x55C34 | X |
| +0x260 | 0x55CA8 | Y |

```
for (i = 0; i < 0x71; i++) {
   if (地點 / 樓層 / X / Y 有一個不合) continue
   ; ★ 三個特例
   if (i == 0x0D && 鑰匙數 == 0 && sub_2B360(…) == 0) continue
   if (i == 0x0E && byte_3E08E != byte_3DFBE)         continue
   if (i == 0x0F && byte_3DFF7 == 0 && sub_2B360(…) == 0) continue
   if (byte_3E06C[i>>3] & (1 << (i&7))) continue      ; 撿過了
   if (i < 0x0D || i > 0x0F) byte_3E06C[i>>3] |= bit  ; ★★ 13..15 不記帳
   找空槽; 生出物件; sub_13CB0(種類)                   ; 印描述
   return
}
印 "nothing of note.\n"
```

### ★★ 索引 13..15 可以重複拿,而第 13 筆是防卡死

`if (i < 0x0D || i > 0x0F)` 才寫位元圖 ⇒ 那三筆**永遠拿得到**。
而第 13 筆的額外條件是 `鑰匙數 == 0`:

| # | 種類 | 品質 | 地點 | 樓層 | 座標 |
|---|---|---|---|---|---|
| 13 | 7(一串鑰匙)| 9 | 18 | −1 | (8, 6) |
| 14 | 7(一串鑰匙)| 133 | 5 | 0 | (2, 2) |
| 15 | 5(武器)| 39 | 0 | 0 | (64, 80) |

⇒ **鑰匙用完了才會再長出來**。上鎖的門要鑰匙,而鑰匙會被用掉 ——
沒有這一條,玩家可能永遠打不開某扇門。

⬜ 第 14 / 15 筆的兩個條件位元組(`byte_3DFBE` / `byte_3DFF7`)語意未解 ⇒
引擎先當成永遠可拿。**留白的方向刻意選「玩家拿得到」** —— 反過來會卡死玩家。

### 表的分佈(113 筆全滿)

- **地點**:28 個地點都有;大地圖 13 筆、地點 1 有 13 筆、地點 17/18 各 9 筆
- **種類**:13 種;最多的是 4(卷軸,21 筆)、3(藥水,19 筆)、5(武器,16 筆)
- 前 12 筆全在**地下世界 (233,233)**、樓層 −1,交替是護甲(品質 15)與武器(品質 41)
  —— 同一格 12 件的藏寶處

## 4. ★ `sub_13CB0` 與 `LOOK2.DAT` 是**兩張不同的表**

`sub_13CB0(種類)` 的 17 個字串寫死在執行檔裡,而 `LOOK2.DAT` 另有一份物件名。
逐筆比對後有 **20 處不同**,其中六處是刻意的:

| 種類 | `LOOK2.DAT`(地面 Look)| `sub_13CB0`(這一張)|
|---|---|---|
| 2 | gold | a sack of gold |
| 7 | a key | a ring of keys |
| 13 | a torch | some torches |
| **25** | a moon stone | ★★ **a strange rock** |
| 30 | a corpse | a rotting body |
| 31 | a corpse | a moldy corpse |
| **14** | a sandalwood box | ★ 落到 default「nothing of note.」|

★★ 第 25 筆:**搜出來的月石認不出來**,與 `sub_13FC4` 印的同一句 ——
兩處一致,所以那是刻意的設計不是疏漏。

⚠ 第 14 筆(檀香木盒,真結局的關鍵道具)在這張表裡**沒有條目**,落到 default。
照原版保留。

## 5. ⬜ 順手發現:物件表滿了會**擠掉**現有物件

`sub_2B57C`(找物件槽)不是「找空槽,沒有就放棄」,而是**九層遞降**:

```
sub_2B518(0,    0,    0)   ; ① 真的空槽
sub_2B518(1,    0x0F, 1)   ; ② 種類 1..0x0F 且旗標 1
sub_2B518(0x80, 0xFF, 1)   ; ③
sub_2B518(0x10, 0x11, 1)   ; ④
sub_2B518(0x30, 0x7F, 1)   ; ⑤
sub_2B518(1,    0x0F, 0)   ; ⑥ 同 ②..⑤ 但旗標 0
sub_2B518(0x80, 0xFF, 0)   ; ⑦
sub_2B518(0x10, 0x11, 0)   ; ⑧
sub_2B518(0x30, 0x7F, 0)   ; ⑨
sub_2B518(0,    0xFF, 0)   ; ⑩ 最後手段
```

⇒ 地上東西太多時原版會**依優先序犧牲舊物件**,而引擎目前是「滿了就放不下」。
⬜ 要補得先讀 `sub_2B518` 的兩個範圍參數與那個旗標的語意。

## 6. 引擎落地

| 原版 | 引擎 |
|---|---|
| 三層的順序與短路 | `searchFixedContents`(接在 `searchAt` 的隨機那一層**之前**)|
| `tile == 0xDC` 沉默 | `searchAt` 的 `u5data.MoongateOpenTile` 分支 |
| `sub_13FC4` | `digUpMoonstone` + `moonstoneObjectHere` |
| `sub_14090` | `gatherHerbs` + `HerbGatherHour` / `HerbCarryLimit` |
| `sub_14160` | `findHiddenItem` + `hiddenItemAvailable` / `markHiddenItemTaken` |
| 六個平行陣列 | `u5data.HiddenItems`(產生器 `tools/dump_hidden_items.py`)|
| 三個採藥點 | `u5data.HerbSpots` |
| `byte_3E06C` / `byte_3E068` | `State.hiddenTaken` / `herbPickedDay`(⬜ 存檔位移未定位)|
| `sub_13CB0` | `u5data.DungeonRoomItemName` |
| 物件放在被搜的那一格 | 新增 `placeFoundObjectAt`(舊的 `placeFoundObject` 放腳下,保留)|

### 測試

| 測試 | 驗什麼 |
|---|---|
| `TestHiddenItemTableShape` | 113 筆、無空列(抽表位址錯會全 0)、可重複的只有 13..15 |
| `TestDungeonRoomNamesDifferFromTheLookTable` | 兩張表的六處差異;**若變成相同就是有人合併了** |
| `TestHerbSpotsNeedMidnight` | 白天採不到、午夜採到、同日同點只一次、隔天可再採 |
| `TestHerbSpotsAreTrackedPerSpot` | 記帳是每點各一份(全域一份會紅)|
| `TestHerbCapAtNinetyNine` | 夾在 0x63 |
| `TestBuriedMoonstoneCanBeDugUp` | 挖得回來、不重複生、品質欄帶編號 |
| `TestMoonstoneDigChecksTheLocationCode` | 地點碼不對挖不到 |
| `TestMoonstoneDigScansBackwards` | 倒著掃,先挖編號大的 |
| `TestSpareKeysOnlyAppearWhenYouHaveNone` | 防卡死那一條 |
| `TestHiddenItemTakenOnce` | 撿過不再出現 |
| `TestFixedContentsRunBeforeTheRandomRoll` | 物件放在**被搜的那一格**、只拿一次 |
