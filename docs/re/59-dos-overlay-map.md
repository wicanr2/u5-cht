# 59 — DOS 版的 overlay 機制、24 個編號,以及 `SJOG.OVL` 是什麼

| | |
|---|---|
| 輸入檔 | `ULTIMA.EXE`(DOS,36,592 B)+ 24 個 `.OVL` |
| 工具 | `tools/ovl_thunks.py`(可重跑:`tools/dev.sh python3 tools/ovl_thunks.py gamedata`) |
| 相關 | `CLAUDE.md` §4.2 把「overlay 切分與載入基址」列為 DOS 版的專屬題目 |

`WORKLIST` 上 24 個 `.OVL` 各掛一條 ⬜,而第一條是 `SJOG.OVL`:
「用途未確認(檔名待解;是第二大的 OVL,值得優先逆)」。

先解的其實不是 `SJOG`,是**整個 overlay 機制** —— 而它一解開,`SJOG` 的定位就跟著出來了。

## 名表就在主程式裡

`ULTIMA.EXE` 偏移 0x817B 起是一串 NUL 結尾字串:

```
ULTIMA.EXE\0 TOWN.OVL\0 MAINOUT.OVL\0 DUNGEON.OVL\0 INTRO.OVL\0 FLAMES.OVL\0
NPC.OVL\0 COMBAT.OVL\0 BLCKTHRN.OVL\0 LOOKOBJ.OVL\0 DNGLOOK.OVL\0 OUTSUBS.OVL\0
SHOPPES.OVL\0 ENDGAME.OVL\0 SJOG.OVL\0 CMDS.OVL\0 CAST.OVL\0 TALK.OVL\0
CAST2.OVL\0 ZSTATS.OVL\0 COMSUBS.OVL\0 SHOPPES2.OVL\0 SHOPPES3.OVL\0
FONT.OVL\0 DATA.OVL\0
```

**陣列索引就是 overlay 編號**,0 是主程式本身。這是 Microsoft C 5.x/6.0 的
overlay linker 產生的模組表 —— 與 `DATA.OVL` 裡那句
`MS Run-Time Library … 1988, Microsoft Corp`(`CLAUDE.md` §2.1)同源。

## 名表後面緊接著 164 個進入點樁

從 0x8216 起,**164 個各 12 B、完全連續**的樁:

```
9A EC 02 2E 07     call far 072E:02EC     ; 換入 overlay 的 loader(164 個樁全指這一支)
03 00              overlay 編號 = 3       ; DUNGEON.OVL
EA FE 8F 00 00     jmp  far 0000:8FFE     ; 段值由載入時的重定位填上
```

一個樁 = **一個跨 overlay 的函式進入點**。主程式(或別的 overlay)要呼叫住在
overlay 裡的函式就 call 這個樁;樁先請 loader 把 overlay 換進交換區,再遠跳過去。

### 為什麼確定格式沒認錯

三條同時成立,而且都不靠「看起來像」:

1. **164 個樁的間距全部正好 12 B**,0x8216 一路連續到 0x89C6,沒有一處例外。
2. **overlay 編號全部落在 1..23。** 沒有 0(主程式不必換入),
   也**沒有 24** —— 24 是 `DATA.OVL`。它是純資料、沒有函式,所以不該有進入點。
   這條是獨立佐證:與「`DATA.OVL` 只放表」這個既有結論互相印證,
   而那個結論當初是從別的方向(明文字串與指標表)得到的。
3. 把每個 overlay 的進入點位址取最小與最大,**跨距一律小於該 `.OVL` 的檔案大小**,
   23 個全部成立。樁格式若認錯,這 23 條不可能同時通過。

## 完整的 overlay 地圖

| # | overlay | 大小 | 進入點 | 位址範圍 | 跨距 |
|---|---|---:|---:|---|---:|
| 0 | `ULTIMA.EXE` | 36,592 | 0 | (無進入點) | |
| 1 | `TOWN.OVL` | 6,256 | 13 | 0x81D0..0x98F6 | 5,926 |
| 2 | `MAINOUT.OVL` | 7,344 | 8 | 0x81D0..0x9C30 | 6,752 |
| 3 | `DUNGEON.OVL` | 8,016 | 8 | 0x8304..0x9FE0 | 7,388 |
| 4 | `INTRO.OVL` | 8,400 | 3 | 0x85FE..0xA250 | 7,250 |
| 5 | `FLAMES.OVL` | 32 | 1 | 0xA290 | 0 |
| 6 | `NPC.OVL` | 4,912 | 4 | 0xA290..0xB570 | 4,832 |
| 7 | `COMBAT.OVL` | 7,408 | 11 | 0xA290..0xBCEC | 6,748 |
| 8 | `BLCKTHRN.OVL` | 3,184 | 2 | 0xA89E..0xABA0 | 770 |
| 9 | `LOOKOBJ.OVL` | 4,560 | 3 | 0xA5F6..0xB38C | 3,478 |
| 10 | `DNGLOOK.OVL` | 5,040 | 10 | 0xA290..0xB40E | 4,478 |
| 11 | `OUTSUBS.OVL` | 2,464 | 9 | 0xA444..0xA8E8 | 1,188 |
| 12 | `SHOPPES.OVL` | 5,936 | 11 | 0xA2B6..0xB788 | 5,330 |
| 13 | `ENDGAME.OVL` | 2,800 | 1 | 0xA8D8 | 0 |
| **14** | **`SJOG.OVL`** | **8,800** | **17** | 0xBFEC..0xE14E | 8,546 |
| 15 | `CMDS.OVL` | 7,440 | 12 | 0xBF80..0xDBA0 | 7,200 |
| 16 | `CAST.OVL` | 8,560 | 2 | 0xCD3A..0xD712 | 2,520 |
| 17 | `TALK.OVL` | 4,880 | 2 | 0xC29E..0xC39C | 254 |
| 18 | `CAST2.OVL` | 4,544 | 16 | 0xE1E0..0xF2DE | 4,350 |
| 19 | `ZSTATS.OVL` | 4,880 | 8 | 0xE63E..0xF476 | 3,640 |
| 20 | `COMSUBS.OVL` | 5,216 | 16 | 0xE1E0..0xF4BE | 4,830 |
| 21 | `SHOPPES2.OVL` | 2,848 | 2 | 0xE84C..0xEC9C | 1,104 |
| 22 | `SHOPPES3.OVL` | 2,528 | 1 | 0xEA94 | 0 |
| 23 | `FONT.OVL` | 3,744 | 4 | 0xE1E0..0xECEA | 2,826 |
| 24 | `DATA.OVL` | 48,464 | **0** | (無進入點) | |

`SJOG.OVL` 的 **17 個進入點是全部 overlay 裡最多的**。

### 共用同一個起始位址的 overlay 不能同時在記憶體裡

進入點位址分成三段,而**幾組 overlay 的最小位址完全相同**:

| 起始位址 | 共用它的 overlay |
|---|---|
| 0x81D0 | `TOWN` `MAINOUT` |
| 0xA290 | `FLAMES` `NPC` `COMBAT` `DNGLOOK` |
| 0xE1E0 | `CAST2` `COMSUBS` `FONT` |

同一個位址代表同一個交換槽 —— 它們**輪流**佔用,不能同時常駐。
這解釋了一件本來看起來多餘的事:`COMBAT.OVL` 與 `COMSUBS.OVL` 為什麼要拆兩個檔
(戰鬥主邏輯與戰鬥副程式落在**不同的槽**,才能同時在記憶體裡);
也解釋了 `SHOPPES` / `SHOPPES2` / `SHOPPES3` 的三段切法。

⚠ 這裡說的「起始位址」是**最小的進入點位址**,不是 overlay 的載入基址 ——
第一個進入點不一定在 overlay 的偏移 0。所以上表可以證明「這幾個共用一個槽」,
但**不能**用來算出每個 overlay 的精確載入基址。要精確基址得逆 loader `072E:02EC`。

## `SJOG.OVL` 是哪四個指令

名字讀起來像四個**指令字母**的縮寫,而 U5 的指令表(`docs/re/49`)裡剛好有
`S)earch` `J)immy` `O)pen` `G)et` —— 四個都是「對相鄰那一格動手」的指令。

拿各指令的招牌常數去掃 24 個 overlay(比對 16-bit 的 `cmp` 立即值編碼),
`SJOG.OVL` 是**唯一**同時命中這幾組的:

| 常數 | 屬於 | SJOG | 其他 overlay |
|---|---|---:|---|
| 0xB4 0xB5 0xB6 0xB7 | Get:碎片 / 王冠 / 權杖 / 護符(`docs/re/57`) | 各 1 | 沒有任何一個命中全部四個 |
| 0xB9 0xBB | Jimmy:上鎖門 / 魔法上鎖門 | 各 2 | `CAST.OVL` 各 1(那是 An Sanct / In Ex Por) |
| 0x97 0x98 | Open:兩種力場 | 各 2 | `CMDS` 1 / `CAST2` 各 1 |
| 0x1F 0x13 0x1E 0x0A | Search:瘟疫 `random(0,31)==19`、陷阱難度 30、簡單陷阱 < 10 | 各 1..3 | 分散 |

**G / J / O 三支是確定的**:四個信物編號在整個 overlay 集合裡只有 `SJOG` 同時比對,
而那正是 `sub_154BC`(Get 分派)的簽名 —— 它對 0xB4/0xB5/0xB6/0xB7 各有一個 `cmp`。
門與力場那兩組也只有它同時齊備。

**S 只有間接證據。** Search 最專一的常數是「有缺口的牆」0x4E,而它在 `SJOG` 裡
**一次都沒出現**(不論位元組或 16-bit 形式)。命中的只有陷阱與瘟疫的門檻,
而那幾個數字太小、太常見,單獨看不足以定案。所以:

> `SJOG.OVL` = **J)immy + O)pen + G)et**(已證),**S 很可能是 S)earch**(未證)。
> 若日後發現 Search 住在別的 overlay,那 `S` 就得另找解釋 —— 不要把這一條當成已知。

## 這解開了什麼、還剩什麼

**解開的**:overlay 編號、進入點總表、哪些 overlay 共槽、`DATA.OVL` 沒有程式碼的
獨立佐證、`SJOG` 的用途。往後要逆任一個 `.OVL`,可以先從 `tools/ovl_thunks.py`
取出它的進入點位址清單當 IDA 的函式起點 —— 不必再從「裸機器碼、無檔頭」硬猜。

**還剩的**:精確的載入基址(要逆 loader `072E:02EC`)、以及每個進入點對應哪個函式。
後者需要把主程式裡「call 哪個樁」與指令分派表對起來,那是下一步。
