# 指令表只有一張 —— 以及我自己發明的那個 W

輸入檔:`WORRIORS.EXP`(FM Towns 英文版)
IDA 位址:`sub_2ACF4`(主指令分派器)、`sub_A360`(戰鬥指令分派器)、
`sub_4B14`(地牢迴圈的前置)、`sub_1F3A4`(Ready)、`sub_147A8` / `sub_142EC`(Search)、
`sub_13BA8`(地牢的相對方向)
日期:2026-08-08

---

## 1. 找到主分派器

原本我在找「地牢裡按 S 會怎樣」,追進 `sub_4B14`(地牢迴圈的鍵處理)只看到
方向鍵、數字鍵(彈豎琴)、'.' 與幾個控制碼。看起來像是**地牢的指令很少**。

但 `sub_4B14+13B` 把鍵交給了 `sub_2ACF4` —— 一個 **59 case 的跳表**,
`switch (鍵 − 0x20)`,涵蓋 **空白鍵到 'Z'**。

⇒ **全遊戲的字母指令只有一張表。** 地面、場景、地牢共用它;
位置差異是在**各指令內部**判的(`byte_3E0A3`:0 = 地表 / < 0x21 = 場景 / ≥ 0x21 = 地牢)。

| 鍵 | 指令 | 位置分支 |
|---|---|---|
| 空白 | Pass | 地表且揚帆 → 「Sheets in irons!」(先收帆);其餘印 `Pass` |
| A | Attack | 地表 `sub_2D478` / 場景 `sub_CAC` / **地牢 `sub_4074`** |
| B | Board | — |
| C | Cast | — |
| **D** | **`D-What?`** | ← 不是指令 |
| E | Enter | 只在地表;其餘印 `Enter what?` |
| F | Fire | — |
| G | Get | 地牢不印 `Get-` 的方向提示 |
| H | Hole up | 場景要在床上(`Only in bed!`);地表是紮營 |
| I | Ignite torch | — |
| J | Jimmy | — |
| K | Klimb | 地表 `sub_188C4` / 場景 `sub_EA0` / 地牢 `sub_417C` |
| L | Look | — |
| M | Mix reagents | — |
| N | New order | — |
| O | Open | — |
| P | Push | 不合條件 → `Push / Not here!` |
| Q | Quit(存檔並離開) | — |
| R | Ready | — |
| S | Search | 地表 / 場景 `Search-`(問絕對方向)/ **地牢 `Search...`(問相對方向)** |
| T | Talk | 地表 `sub_2B2AC`;地點 ≤ 0x20 → `Talk-Funny, no response!`;其餘 `sub_1B658` |
| U | Use item | — |
| V | View a gem | 沒寶石 → `You have none!`;地牢 `sub_F7C0` / 其他 `sub_EDD4` |
| **W** | **`W-What?`** | ← 不是指令 |
| X | Xit | — |
| Y | Yell | — |
| Z | Ztats | — |

## 2. ★ D 與 W 都是空鍵 —— 而我把 W 做成了指令

`sub_2ACF4` 對 'D' 與 'W' 只做一件事:印 `D-What?` / `W-What?`。
戰鬥的分派器 `sub_A360` 也各印一次同樣的字串(不同的字串池副本,
`aDWhat`/`aWWhat` vs `aDWhat_0`/`aWWhat_0`)—— **兩個分派器獨立佐證**。

`WORKLIST` 上原本寫「⬜ 只剩 D(`aDWhat_0`,疑為未使用的鍵)」—— 那條猜對了,
現在有分派器本身當證據。**但同一張表上的 W,我卻做成了 Wear。**

錯因很具體:U4 有 `Ready`(武器)與 `Wear`(護甲)兩個鍵,我把那個模型搬過來,
還照著它把裝備清單切成兩半。U5 沒有 —— `sub_1F3A4`(R)的清單範圍是

```asm
push edi              ; 隊員
push offset byte_3DFD0 ; 背包
push 30h              ; 48 = 全部裝備
push 0FFFFFFFFh
call sub_1E418
```

**0x30 = 48 = 頭盔 + 盾 + 護甲 + 武器 + 戒指 + 頸飾全部**。一支 R 收下六類。

> 這是 `rulebook/65` 最乾淨的一個例子:自家測試**全綠**,而且綠得很紮實 ——
> 有一支 `TestEquipListsSplitByCategory` 專門釘住「R 與 W 的清單不重疊」。
> 它每次都通過,因為它量的是**我自己發明的規則**。
> 測試會把錯的行為釘得跟對的一樣牢。

修法:`Ready` 收全部六類、`ReadyList` 列全部裝備、`Wear`/`WearList`/`BeginWear`/
`MsgCannotWear` 全部移除、W 與 D 改成印 `W—— 何事?` / `D—— 何事?`。
`equip()` 的欄位判斷與「換下來的放回背包」沒動 —— 錯的只是鍵位與清單切分。

## 3. 輸入層不該按位置分兩份清單

`cmd/u5cht/main.go` 原本有兩份字母鍵清單:地面那份很完整,地牢那份只有
K / C / I / O / G / Q。結果 **S 搜尋、Z 數值、U 用道具、R 換裝、A 攻擊、
L 觀察、M 調藥、N 換位、J 撬鎖、V 看寶石在地牢裡全部按不到** ——
規則都實作了,只是按不到。

> 「做完了卻用不到,等於沒做」是我在 `picker.go` 開頭自己寫下的話,
> 然後在隔壁檔案犯了同一件事。分兩份清單就一定會漏,而漏掉的東西
> **在單元測試裡完全看不出來**:每一支規則的測試都通過。

改成一個 `commandKeys()`,兩邊共用 —— 與原版只有一個 `sub_2ACF4` 同構。
留在各自分支的只有**互動流程真的不同**的兩項:方向鍵(地牢是前進 / 轉向)
與 K(地牢要再問上 / 下)。

## 4. 地牢的 Search 是另一支程式

```
sub_147A8:  if (0x20 < 地點 < 0x29) → sub_142EC()   ← 地牢
                               else → 問絕對方向,搜鄰格
```

三處不同:

1. **方向是相對的。** `sub_13BA8` 印 `Dir-`,收到的鍵碼對應
   `1 = Left`、`2 = Right`、`3 = Ahead`、`4 = Here`(腳下!)、空白 = `Pass` 放棄。
   第一人稱下「北」沒有意義。
   ⚠ 而且在這條路徑上 `word_3E086`/`word_3E088` 存的是**絕對座標**,
   不是地面那條路的方向增量 —— 同一對全域變數在兩條路上意義不同。
2. **看不見就搜不到。** `byte_3E0B6`(火把)與 `byte_3E0B7`(In Lor)兩個
   計時器都為 0 → `You find: darkness.`
   ⚠ 判的是**計時器**不是視野半徑。地牢的基礎半徑是夜間的 2,永遠 > 0,
   拿半徑判會變成「永遠有光」。
3. **報的是地形本身**,一種地形一句話,而 **0xD0 是暗門**:

   | 高四位元 | 訊息 |
   |---|---|
   | 0x00 | `Nothing of note.` |
   | 0x10/0x20/0x30 | `Nothing hidden on the ladder.` |
   | 0x40 | `Treasure!`(另有陷阱判定) |
   | 0x50 | `Nothing hidden on the fountain.` |
   | 0x60 | `Nothing hidden in the pit.` / `A pit!` |
   | 0x70 | `Nothing hidden on the door.` |
   | 0x80 | 力場:`A sleep field` / `A poison gas field` / `A wall of fire` |
   | 0x90 | `This tile is impossible.` ← 原版自己這麼寫 |
   | 0xA0 | 房間 |
   | 0xB0 | `Nothing hidden on the wall.` |
   | 0xC0 | `Nothing in the caved in passage.` |
   | **0xD0** | **`A hidden door!`** |
   | 0xE0/0xF0 | `Nothing hidden on the door.` |

### 開暗門的那一行不能寫成 `0xE0`

```asm
mov eax, edi
and al, 8          ; 保留「頭上有洞」那一位元
add al, 0E0h       ; 變成門
```

⇒ 是 `0xE0 | (原值 & 8)`。直接寫 `0xE0` 會把洞抹掉,而那個洞是**用繩索爬回
上一層的唯一線索** —— 症狀要等到玩家想回頭時才出現,那時已經無法回推原因。
測試釘住了這一位元。

## 5. 引擎對應

| 原版 | 引擎 |
|---|---|
| `sub_2ACF4` | `cmd/u5cht.(*game).commandKeys`(地面與地牢共用) |
| `aDWhat` / `aWWhat` | `game.MsgDWhat` / `MsgWWhat` |
| `sub_1F3A4` 的 48 件清單 | `game.(*State).ReadyList` |
| `sub_142EC` | `game.(*State).searchDungeon` |
| `sub_13BA8` 的相對方向 | `game.SearchAhead/Left/Right/Here` |
| `and al,8; add al,0E0h` | `game.dungeonSecretDoorOpens` |

## 6. 還沒做的

- ⬜ **搜尋的擲骰那一層**:門檻是 `(樓層 × 2 + 30 − 敏捷) / 2`,擲過了才會
  多找到東西(寶箱的陷阱、坑裡的物品、骸骨崩解…)。骰子的上下界還沒讀出來,
  所以先只做「每種地形的那一句話」。**少一個獎勵好過憑感覺補一個機率。**
- ⬜ H(Hole up)紮營與 `Only in bed!`;A 在地表 / 場景的兩支攻擊入口。
- ⬜ `sub_A360`(戰鬥的指令表)還沒逐 case 對過 —— 戰鬥目前只綁了 A / C / 空白 / ESC。
- ⬜ 原版**每個指令都會先印指令名**(`Ready-`、`Search-`、`Yell-`…),
  引擎多數指令沒印。是體感差異,不是規則差異,但巡查文字時要一起補。
