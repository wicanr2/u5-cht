# Use 指令與一個錯了很久的名字

輸入檔:`WORRIORS.EXP`(FM Towns 英文版)、`gamedata/DATA.OVL`
IDA 位址:`sub_1A5E8`(Use)、`sub_1E8D4`(可用道具清單)、`sub_1A5B0`(信物切換)、
`sub_154BC`(撿取)
日期:2026-08-08

---

## 0. 先講那個名字:0x020D 是**護符**,不是寶珠

`sub_154BC` 的撿取分支寫得毫無歧義:

```asm
loc_158DE:
        mov     byte_3DFBF, 0FFh
        push    offset aTheAmuletOfLor ; "The Amulet of Lord British!\n"
```

而 `sub_1E8D4` 建可用道具清單時,`byte_3DFBF` 落在特殊道具表**第 2 筆
(`Amulet`)** 的位置 —— 兩處獨立一致(`rulebook/62`)。

⇒ **U5 的三件信物是王冠 / 權杖 / 護符,沒有寶珠。** 「Orb of the Moons」
是 U6 的東西。這個錯名從 `SaveOrbOffset` 一路傳到 `Regalia.Orb`、
`MsgGotOrb`、乃至**結局判定的條件**,已全部更正為 Amulet。

機制本身一直是對的(結局確實要三件信物),只有名字錯 —— 而正是那個名字
讓我在做 Use 的時候差點以為存檔位移解錯了,回頭重驗了一輪。
**錯名的代價不是當下,是下一個讀它的人。**

## 1. 道具編號 = 特殊道具表索引 + 16

表在 `DATA.OVL 0x04C3`。`sub_1E8D4` 依序讀這些旗標,順序與表完全對應:

| 表索引 | 名字 | case | 旗標 |
|---|---|---|---|
| — | (月石 ×8) | — | `byte_3E050[0..7]` |
| 0 | Magic Crpt | 16 | `byte_3DFBC` |
| 1 | Skull Keys | 17 | `byte_3DFBD` |
| 2 | **Amulet** | 18 | `byte_3DFBF` |
| 3 | Crown | 19 | `byte_3DFC0` |
| 4 | Sceptre | 20 | `byte_3DFC1` |
| 5..12 | `(0`..`(7` | 21..28 | — **不可用**(跳表走 default) |
| 13..15 | Shard/… | 29..31 | `byte_3DFC4[0..2]` |
| 16 | Spyglass | 32 | `byte_3DFC8` |
| 17 | HMS Cape Plan | 33 | `byte_3DFC9` |
| 18 | Sextant | 34 | `byte_3DFCA` |
| 19 | Pocket Watch | 35 | `byte_3DFCB` |
| 20 | Black Badge | 36 | `byte_3DFCC` |
| 21 | Wooden Box | 37 | `byte_3DFCD` |

⚠ 5..12 那八筆的名字是 `(0`..`(7`,看起來像資料損毀,其實是原版留的佔位名 ——
它們在跳表裡走 default,根本用不到。**不要試著「修好」它們。**

## 2. 信物共用一個模式位元組

`sub_1A5B0(mode)`:目前的 `byte_3E08A` 已經等於 mode → 印「Removed!」並清零;
否則回 1(繼續穿上)。穿上時 `sub_1D31C(mode, 0xFF, 9)`。

| 道具 | mode |
|---|---|
| Amulet | 0x0E |
| Crown | 0x1C |
| Black Badge | 0x1D |

★ **`byte_3E08A` 就是四個咒語共用的那個位元組**(In Sanct 'P'、Rel Tym 'Q'、
Quas An Wis 'C'、In An 'N'、An Tym 'T')。所以戴上王冠會蓋掉 In Sanct、
而 An Tym 也會蓋掉王冠。寫成兩個獨立的布林值,行為就與原版不同 ——
而那種差異在遊玩時只覺得「怪」,追不到源頭。

0x1D 這一格同時是 `guard.go` 早就定出的 `BadgeMode` —— 又一次獨立對上。

⚠ **權杖不設模式。** 它只放三個音效然後化力場。寫成「跟王冠一樣」
會多出一個原版沒有的持續效果。

## 3. 各道具的效果與前置

| 道具 | 效果 | 前置 |
|---|---|---|
| 魔毯 | 上魔毯 | 不在船上、必須步行、不在場景/地牢 |
| 骷髏鑰匙 | 化掉眼前的力場,**用掉一把** | 眼前要有力場 |
| 護符 / 王冠 | 切換模式 | 身上要有 |
| 權杖 | 化力場 | 身上要有;沒力場印「No effect!」 |
| 望遠鏡 | 看星象 | 戶外、夜裡 |
| 圖紙 | 船速加倍,**不消耗** | 在船上 |
| 六分儀 | 報 chunk 座標 | 戶外、夜裡 |
| 懷錶 | 報時(與老爺鐘同格式) | — |
| 黑徽章 | 戴上(mode 0x1D) | 身上要有 |
| 碎片 | 只能投入聖火 | 走聖火那一支(`docs/re/26`) |
| 木盒 | 「Box- How?」 | ⬜ 分支未逆 |

六分儀 / 望遠鏡的兩道前置**順序有意義**:原版先擋室內、再擋白天。
反過來的話在城裡的白天會說「只有夜裡才行」,而正確的回答是「只有戶外才行」。

## 4. 未做

- **木盒**(case 37)的分支:原版印「Box- How?」再依答案走,還沒逆。
- **黑徽章的存檔位移**:`sub_1E8D4` 讀 `byte_3DFCC`,但存檔 0x0216..0x0218
  三格對三個變數還沒逐一分派完。所以 `State.HasBadge` 目前只在記憶體裡,
  存讀檔不保留 —— **留白比對錯位移好**。
- 道具選單的畫面(原版 `sub_1E418` 印「Items:」讓玩家挑字母)。
