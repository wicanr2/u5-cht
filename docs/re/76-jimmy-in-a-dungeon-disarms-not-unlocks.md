# 76 — 地牢裡的 J 撬的是陷阱,不是鎖

| | |
|---|---|
| 輸入檔 | `re_work/fmtowns/WORRIORS.EXP.asm`(FM Towns 英文版) |
| SHA-256 | 見 `re_work/fmtowns/SHA256SUMS.txt` |
| 主要函式 | `sub_14CAC`(Jimmy 指令本體)、`sub_14B2C`(地牢那一支) |
| 落地 | `internal/game/commands.go` · `strings.go` |
| 從哪裡來 | `docs/re/66` 的截斷清單(`sub_14B2C`,字串 `No keys!` / `Key broke!` / `Chest unlocked`) |

---

## 1. 症狀:地牢裡按 J 撬不了任何東西

`sub_14CAC` 開頭就分岔:

```asm
mov al, byte_3E0A3
cmp al, 20h ; jbe → 走門那條
cmp al, 29h ; jnb → 走門那條
call sub_14B2C                 ; ★ 0x21..0x28 = 地牢
```

引擎只有門那條,而 `jimmyAt` 走的是 `s.TileAt(x, y)` —— 在地牢裡那個函式
讀不到地牢格(地牢地形在 `s.Dungeons` 裡)。⇒ 按 J 只會得到「此處沒有鎖」。

## 2. ★★ 它解的是陷阱,不是鎖

成功之後那一格是:

```asm
mov eax, edi
and al, 8                      ; 保留「頭上有洞」
add al, 40h                    ; 還是 0x4x = 箱子
```

⇒ **`(tile & 8) | 0x40`** —— 箱子還在、還是關著的,**低三位元的陷阱被清掉了**。
所以 J 在地牢裡的用途是「**先解陷阱,再用 O 開**」,而不是「打開箱子」。

（`docs/re/75` 已經定出低三位元是 Open 的陷阱判準:`test di, 7`。
兩支獨立指向同一組位元。）

## 3. ★ 沒有陷阱的箱子撬不開,而且鑰匙照斷

第一個分支:

```asm
mov eax, edi ; and eax, 0F7h   ; ★ 清掉 bit 3
cmp eax, 40h
jnz short loc_14BF5
    鑰匙 == 0 → "No keys!"
    否則      → "Key broke!" 並扣鑰匙        ; ★★ 沒有擲骰,一定失敗
```

`tile & 0xF7 == 0x40` 涵蓋 **{0x40, 0x48}** —— 低三位元為 0 的箱子。
那種箱子**沒有陷阱可解**,所以原版直接跳到「鑰匙斷了」。

⇒ 把它寫成「沒陷阱就成功」會讓鑰匙變成萬能鑰匙,而原版是**在沒陷阱的箱子上
硬撬會白白折斷鑰匙**。這條沒有任何測試抓得到(兩種寫法都「有反應」),
只有讀那個 `and 0F7h` 才看得到。

## 4. 有陷阱的那條:與地牢搜尋同一條式子

```asm
dex   = byte_3DDC1[who*32]
edx   = 樓層 * 2 + 30 − dex
門檻  = signed(edx) / 2
random(1, 30) > 門檻  → "Chest unlocked"、清陷阱
否則                  → "Key broke!"、扣鑰匙
```

`(樓層×2 + 30 − 敏捷) / 2` 與 `dungeonSearchThreshold`(`docs/re/xx` 的
地牢搜尋)**是同一條式子**,引擎直接複用。

⚠ 比較方向是 **`>` 才成功**;地牢搜尋那邊是 `<=` 才成功(`sub_142EC`)。
同一條門檻、相反的方向 —— 抄的時候很容易順手抄成同一個比較。

## 5. ★ 問人與查鑰匙的順序,兩條路相反

| | 順序 |
|---|---|
| 門(`sub_14CAC`) | **先查鑰匙**(`cmp byte_3DFB8, 0` 在問方向之前)→ 沒鑰匙連方向都不問 |
| 地牢(`sub_14B2C`) | **先問人**(`sub_E19C`)→ 讀格子 → 才查鑰匙 |

所以身上沒鑰匙時,在地牢裡原版**照樣先問「Player:」**。
引擎原本 `Jimmy()` 一律先查鑰匙 —— 那對門是對的,對地牢是錯的。
現在地牢那條走自己的順序。

## 6. 落地與驗收

四條測試:

- `TestJimmyInADungeonDisarmsTheTrapNotALock` —— 成功後**仍是 0x4x**、
  陷阱位元清掉、**鑰匙不扣**
- `TestJimmyWastesAKeyOnAnUntrappedChest` —— 0x40 與 0x48 兩種都驗,
  敏捷拉滿也一樣斷鑰匙、而且**不准印「解開了」**
- `TestJimmyInADungeonAsksWhoBeforeCheckingKeys` —— 沒鑰匙的順序
- `TestJimmyOnAnOpenedChestSaysAlreadyOpen`

## 7. 還沒讀的

- ⬜ `sub_14CAC` 門那條的完整跳表(`sub_DB10` 之後那一段)引擎已有,
  但沒有逐 case 對照過 —— 目前只確認了「普通鎖擲骰、魔法鎖必斷」兩條。
- ⬜ `sub_14B2C` 的 `dword_55A48`(進來設 1、失敗設 0)語意未追。
  它在 `sub_1994C`、`sub_15374` 等處也出現,疑為「這一步要不要重畫地牢視圖」。
- ⬜ 「頭上有洞」位元(0x08)在箱子上的意義沒查 —— 它在 Open 與 Jimmy
  兩支裡都只是被**保留**,從沒被讀過。
