# 建立新角色 —— 吉普賽的八德淘汰賽

輸入檔:`WORRIORS.EXP`(FM Towns 英文版)、`gamedata/QUESTION.DAT`
IDA 位址:`.text:000235B6`(流程)、`sub_23274`(一題)、`sub_23248`(抽美德)、
`sub_60BC`(主選單)
日期:2026-08-08

---

## 0. 為什麼這條要先做

在這之前引擎只能從 `SAVED.GAM` / `INIT.GAM` 開局 —— 玩家**只能扮演別人做好的聖者**。
CLAUDE.md §6.1 的驗收鐵則寫著「不要打包測試角色存檔,玩家該走建角流程」,
而那條規則在流程不存在的時候是空的。

## 1. 主選單(`sub_60BC`)

```
Journey Onward
Create New Character
Transfer from Ultima IV
Ultima V Introduction
Acknowledgements
Return to the View
```

引擎目前沒有主選單,建角先用 `-create` 旗標接上。`Transfer from Ultima IV` 未做。

## 2. 賽制:8 → 4 → 2 → 1

```
loc_235F3:  4 次 sub_23274      八德剩四
            清 byte_5717C
loc_2361B:  2 次 sub_23274      剩二
            清 byte_5717C
loc_23641:  1 次 sub_23274      決出唯一存活者
```

**一共七題。** 兩個旗標分工不同,不能併:

| 旗標 | 意思 | 何時清 |
|---|---|---|
| `byte_5717C[v]` | 本輪抽過了 | **每輪之間清空** |
| `byte_57184[v]` | 已被淘汰 | 一路留到最後 |

併成一個的話,第二輪會重新抽到已經被淘汰的美德 —— 症狀只是「題目怪怪的」,
不會壞掉,所以特別容易寫錯。

`sub_23248` 是**拒絕取樣**:隨機 0..7,抽到已抽過或已淘汰的就重抽。

## 3. 一題的效果(`sub_23274`)

```asm
loc_23398:
    mov     eax, offset dword_57164
    mov     dl, [eax+edi]           ; edi = 勝方
    add     [eax+2Ah], dl           ; ← 加到 智力
    mov     dl, [eax+edi+8]
    add     [eax+2Bh], dl           ; ← 加到 敏捷
    mov     dl, [eax+edi+10h]
    add     [eax+2Ch], dl           ; ← 加到 力量
    mov     byte ptr [eax+esi+20h], 1   ; esi = 敗方 → 淘汰
```

⚠⚠ **是 `add`,不是覆蓋。** Hex-Rays 把這三行反編譯成

```c
byte_5718E = *((_BYTE *)dword_57164 + v4);
```

照著寫的話七題只有最後一題算數,屬性少掉六題份 —— 而角色仍然「看起來正常」,
只是弱得莫名其妙。這是 CLAUDE.md §4.4 那條規則的又一個實例:
**看到賦值先回去讀組語。**

### 美德 → 屬性表(`dword_57164` 的前 24 B)

```
1000002h, 10100h, 1000200h, 10001h, 20000h, 10101h
→ 02 00 00 01 | 00 01 01 00 | 00 02 00 01 | 01 00 01 00 | 00 00 02 00 | 01 01 01 00
```

拆成三列八欄(列 0 智力 / 列 1 敏捷 / 列 2 力量):

| 美德 | 智 | 敏 | 力 | 系列對應職業 |
|---|---|---|---|---|
| 誠實 | 2 | 0 | 0 | 法師 |
| 慈悲 | 0 | 2 | 0 | 吟遊詩人 |
| 勇氣 | 0 | 0 | 2 | 戰士 |
| 正義 | 1 | 1 | 0 | 德魯伊 |
| 犧牲 | 0 | 1 | 1 | 工匠 |
| 榮譽 | 1 | 0 | 1 | 聖騎士 |
| 靈性 | 1 | 1 | 1 | 遊俠 |
| 謙遜 | 0 | 0 | 0 | 牧羊人 |

★ 與 Ultima IV 的職業對應**完全吻合** —— 這是第二重佐證,
不是拿一個數字硬套(`rulebook/62`:兩個獨立來源一致才算)。

## 4. 題目對照表

`dword_57194[8*a + b]` 是「a 對上 b」那一題在 `QUESTION.DAT` 的**檔案位移**,
8×8 對稱、對角線為 0。從執行檔抽出來(FM Towns 的檔案位移 = 線性位址 + 0x200):

```
      0     6C2   7A7   88C   960   A47   B2D   C11
    6C2     0     CDA   D8F   E77   F56  1039  111B
    7A7   CDA     0    11FF  12DB  1374  1456  14ED
    …(完全對稱)
```

28 個不重複的位移,正好是 8 取 2。換算成記錄索引就是 **2..29**;
記錄 0 是吉普賽的開場白、記錄 1 是問完之後的結語 —— 30 筆全部用上。

引擎存的是記錄索引而不是位移:`TextFile` 本來就按記錄取字,
而位移在不同語言版的 `QUESTION.DAT` 會不一樣。

## 5. 收尾(`.text:00023717`)

```asm
mov     al, byte_5718E
mov     byte_3DDC2, al      ; CharIntel  (+14)
mov     byte_3DDC3, al      ; CharMP     (+15)  ← 同一個值!
mov     al, byte_5718F
mov     byte_3DDC1, al      ; CharDex    (+13)
cmp     byte_57190, 14h
ja      short loc_23740
mov     eax, 14h            ; ← 力量下限 20
loc_23740:
movzx   eax, byte_57190
loc_23747:
mov     byte_3DDC0, al      ; CharStrength (+12)
```

兩條容易漏的:

- **初始魔力等於智力** —— 兩個位元組寫同一個值。
- **力量有下限 20**。謙遜(全 0)一路贏到底的話原始值可能低於 20。

三個累加器的**起點不是 0**:問答之前先把角色現有的三圍抄進來
(`loc_235AD` 把 `byte_3DDC0..C2` 寫進 `+0x2A..0x2C`)。所以引擎從
`INIT.GAM` 的聖者讀起,而不是憑空造值。

之後原版讀 `INIT.OOL`、寫 `A:SAVED.OOL` 與存檔 —— 也就是
**把新角色寫進那份初始存檔**,不是從零造一個。引擎照做:只改
名字 / 性別 / 三圍 / 魔力,其餘(裝備、HP、經驗、位置)沿用 `INIT.GAM`。
自己補值等於自創數值(CLAUDE.md §3.0)。

## 6. 引擎對應

| 原版 | 引擎 |
|---|---|
| `dword_57164` 前 24 B | `u5data.VirtueBonus` |
| `dword_57194` 8×8 | `u5data.VirtueQuestion`(換算成記錄索引) |
| 4 / 2 / 1 | `u5data.CreateRounds` |
| `cmp byte_57190, 14h` | `u5data.CreateMinStrength` |
| `sub_23248` | `game.State.drawVirtue` |
| `sub_23274` | `game.State.AnswerCreation` |
| 整段流程 | `game.Creation` + `PromptCreate` |

## 7. 未做

- **主選單**(`sub_60BC` 的六個項目)。目前用 `-create` 旗標代替。
- **`Transfer from Ultima IV`** —— 系列特色,`DATA.OVL` 0x3115 / 0x3529 有字串,
  轉換規則未逆。
- 建角畫面的插圖(`CREATE.16`)。
