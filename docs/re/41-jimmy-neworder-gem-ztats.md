# 四個指令:Jimmy / New order / View a gem / Ztats

輸入檔:`WORRIORS.EXP`(FM Towns 英文版)
IDA 位址:`sub_2ACF4`(A..Z 派發)、`sub_14CAC`(J)、`sub_17688`(N)、
`sub_EDD4`(V)、`sub_1E9A0`(Z)
日期:2026-08-08

---

## 0. 指令派發表(`sub_2ACF4` 的 `jpt_2AD0D`)

一次把 A..Z 全列出來,免得下次又逐支找:

| 鍵 | 指令 | 處理函式 | 引擎 |
|---|---|---|---|
| B | Board | `sub_16F08` | ✅ |
| C | Cast | `sub_4074` | ✅ |
| D | (D-What?) | — | ⬜ |
| E | Enter | — | ✅ |
| F | Fire | `sub_172C4` | ⬜ 船上開砲 |
| G | Get | `sub_15A94` | ✅ |
| H | Hole up | `sub_2B8CC` | ✅ |
| I | Ignite torch | `sub_17630` | ✅ |
| J | Jimmy | `sub_14CAC` | ✅ 本篇 |
| K | Klimb | `sub_188C4` | ✅ |
| L | Look | `sub_D9C4` | ✅(`docs/re/37`) |
| M | Mix reagents | `sub_18704` | ⬜ |
| N | New order | `sub_17688` | ✅ 本篇 |
| O | Open | `sub_16BA0` | ✅ |
| P | Push | `sub_18154` | ✅(`docs/re/40`) |
| Q | Quit & save | `sub_1DD00` | ✅ |
| R | Ready | `sub_1F3A4` | ⬜ 換武器 |
| S | Search | `sub_147A8` | ⬜ |
| T | Talk | `sub_1B658` | ✅ |
| U | Use item | `sub_1A5E8` | ⬜ |
| V | View a gem | `sub_EDD4` | ✅ 本篇 |
| W | Wear armour | — | ⬜ |
| X | X-it | `sub_177AC` | ✅ |
| Y | Yell | `sub_17E74` | ✅ |
| Z | Ztats | `sub_1E9A0` | ✅ 本篇 |

## 1. Jimmy(`sub_14CAC`)

```
byte_3DFB8(鑰匙數)== 0 → "No Keys!"    ← 不問方向
問方向,看目標 tile:
    0xB9 / 0xBB  普通鎖門  → random(0, 29) 對上該員的**敏捷**(CharDex)
                             擲值 < 敏捷 → "Unlocked!";否則 "Key broke!"
    0x97 / 0x98  魔法鎖    → **一律 "Key broke!"**
    其他                   → 走 NPC 那條分支("No one is there!")
鑰匙數 −1
```

⚠ 魔法鎖那條是「必定失敗**而且照樣扣鑰匙**」——`loc_14DC0` 直接跳到扣鑰匙那段。
寫成「魔法鎖不能撬,什麼都不發生」會讓玩家可以無限試,而原版是會把鑰匙耗光的。
那才是真正的代價。

比較方向也要注意:`cmp eax, edx; jl` → **擲值小於敏捷才成功**。
寫反的話「手笨的人反而撬得開」,而在隨機的表象下沒有人看得出來。

## 2. New order(`sub_17688`)

```
印 "Swap ",選一個人 → -1 就 "nobody!"
若 index == 0 → "<名字> must lead!"    ← 聖者不能離開第一位
印 "with ",選第二個人 → 同樣的兩道檢查
交換兩筆**完整的 32 B 記錄**(rep movsd × 8)
```

⚠ 「聖者必須帶隊」不是介面上的講究:隊伍第 0 格是聖者,存檔格式、
對話系統(`AvatarName`)、結局判定全部假設它在那裡。少了這道檢查,
症狀會出現在很遠的地方(例如結局判定找錯人),幾乎追不回來。

## 3. View a gem(`sub_EDD4`)

★ **與 In Quas Wis 是同一支函式。** 咒語版與寶石版在原版共用同一個
32×32 全景畫面,差別只在寶石版先檢查「有沒有寶石」(`You have none!`)。

⚠ **不扣寶石。** `loc_2B115` 只檢查數量,看完寶石還在。這很反直覺,
所以引擎那邊特別註明並加測試 —— 不要「順手」補一行扣除。

## 4. Ztats(`sub_1E9A0` / `ZSTATS.OVL`)

資料全部來自存檔的角色記錄(`docs/re/07`),沒有新的解碼工作 ——
缺的一直是畫面。引擎先做文字版:一名隊員一頁,左右翻頁。

顯示用中文、比對用英文:狀態碼是 `'G'`/`'P'`/`'D'`/`'S'`/`'C'`,
治療所與復活判定比的是那個位元組,譯名不影響它們。

## 5. 未做

`F`(開砲)、`M`(調藥)、`R`(裝備武器)、`S`(搜尋)、`U`(用道具)、
`W`(穿盔甲)、`D`。處理函式都已在上表定位,下一輪從那裡開始。

`S`(`sub_147A8`)的形狀已經看過:依家具種類印不同的開場白
(「In the stump」「On the shelf」「In the bookshelf」「Near the well」「In the desk」),
還會找到隱藏門;獎品來自 `sub_13F04` / `sub_13DD8` 的表 —— 那兩張表還沒定位。
