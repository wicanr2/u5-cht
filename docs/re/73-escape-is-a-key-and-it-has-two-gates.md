# 73 — 撤離是一個鍵,而且它有兩道閘門

| | |
|---|---|
| 輸入檔 | `re_work/fmtowns/WORRIORS.EXP.asm`(FM Towns 英文版) |
| SHA-256 | 見 `re_work/fmtowns/SHA256SUMS.txt` |
| 主要函式 | `sub_18380`(ESC = 撤離)、`sub_A360`(戰鬥分派器)、`sub_42CC`(進地牢房間)、`sub_A9EC`(`byte_3E0B3` 的設定處)、`sub_2E364`(戰鬥入口,設 `byte_3E0B1`) |
| 落地 | `internal/game/combat.go` · `combatcmd.go` · `dungeon.go` · `dungeonmonster.go` · `holeup.go` |
| 從哪裡來 | `docs/re/66` 的截斷清單 |

---

## 1. 引擎裡有一句寫錯的註解

`CombatFlee()` 原本是:

```go
s.Log("汝撤離了戰場。(原版沒有撤離鍵 —— 要一步步走出戰場邊緣)")
s.EndCombat(false)
```

**原版有撤離鍵。** 而且引擎自己在 `docs/re/51` §4 就記過戰鬥可用鍵
「…加上方向鍵 / **ESC** / 空白 / '0' / '1'–'6'」—— 只是沒去追 ESC 做什麼。

## 2. 把 case 27 對回 ESC

戰鬥分派器 `sub_A360` 是 `switch (鍵 − 1)`、90 個 case。要確定 case 27 是哪個鍵,
拿三個「一眼認得出」的 case 當錨:

```
case 89 → "Yell"                'Y' = 0x59 = 89   ✓
case 90 → "Z-Stats"             'Z' = 0x5A = 90   ✓
case 32 → "Pass"                空白 = 0x20 = 32  ✓
case 48 → "Set active plr:None" '0' = 0x30 = 48   ✓
cases 49..54 → sub eax, 31h     '1'..'6'          ✓
cases 1..4  → 方向
```

⇒ **case = ASCII 碼**,四個錨一次全對。所以 **case 27 = 0x1B = ESC**,
而它 `call sub_18380`。

## 3. `sub_18380` 的兩道閘門

```asm
印 "Escape"
ebx = 0
for (edi = 0; edi < 32; edi++)                       ; 掃全部單位
    if ((unit[edi].flags & 0A0h) == 80h) { ebx = 1; break }
if (ebx) {                                            ; ★ 只有「還有活著的隊員」才檢查
    if (byte_3E0B1 & 80h)  { 印 "-Not here!"; esi = 1 }
    else if (byte_3E0B3 == 0) { 印 "-Not yet!"; esi = 1 }
}
if (!esi) {                                           ; 放行
    sub_27230(21h)
    for (edi = 0..31) if (unit[edi].flags != 0) sub_B210(−edi−1)   ; 移除每個單位
    for (edi = 0..31) if (物件[edi].tile != 0) sub_B210(edi+1)     ; 移除每個物件
    音效
}
byte_41989 = 1                                        ; 不論成不成功都設
return esi                                            ; 1 = 被拒絕
```

### ★ 遮罩是 0xA0,不是 0x80

`(flags & 0A0h) == 80h` = **隊員(0x80)而且沒死(0x20)**。
而兩道閘門**只在找到這樣的人時才檢查** ⇒ **全隊倒下時 ESC 一定放行**。

少了 `UnitDead` 那一位會怎樣:全隊倒下、`byte_3E0B3` 還是 0(勝負未定)、
ESC 被「時候未到」擋住 → **玩家被鎖在戰場上,沒有任何鍵能離開**。
那不是難度,是死鎖。`TestEscapeAlwaysWorksWhenTheWholePartyIsDown` 專門守這條。

### 第二道閘門 `byte_3E0B3` = 勝負已定

`docs/re/72` §5 已經定出它:`sub_A9EC` 開戰時設 0/1,而 `VICTORY!` 與
`BATTLE IS LOST!` **都只在它為 0 時印**,印完 victory 還把它設成 1。
⇒ **U5 的戰鬥沒有免費的退出鍵**:要跑就得一步步走出戰場邊緣
(那是另一支 `sub_2F294`,印 `escapes!`)。

## 4. ★ 第一道閘門 `byte_3E0B1 & 0x80`:只有地牢房間

`byte_3E0B1` 是「怎麼進到這場戰鬥的」,由 `sub_2E364` 的第一個參數設定。
全檔 `mov byte_3E0B1, …` 只有六處,其中兩處是存檔序列化(`sub_27D24`)與
`sub_2E364` 自己:

| 值 | 出處 | 場合 |
|---|---|---|
| 2 | `sub_2E364(2, …)` ×2 | 地牢遊蕩怪物 |
| 4 | `sub_2E8B0` → `sub_2E364(4, …)` | 地表紮營 |
| 6 | `sub_2B8CC`(Hole up)的地牢分支 | 地牢紮營(2\|4) |
| **0x82** | **`sub_42CC`(「Entering room...」)** | ★ **地牢房間** |

⇒ **0x80 全檔只出現一次。** 地牢房間是唯一「ESC 離不開」的戰場 ——
房間要靠自己的出口走出去,不能一鍵脫離。而 `sub_42CC` 也是唯一
**不經過 `sub_2E364`** 就開戰的路(它自己 `mov byte_3E0B1, 82h` 再
`call sub_A9EC`),這也是它能設一個別人設不到的位元的原因。

## 5. 落地時的一個陷阱

引擎的 `beginRoomCombat` 被**三種場合共用**:地牢房間(`dungeon.go`)、
地牢遊蕩怪物(`dungeonmonster.go`)、地牢紮營(`holeup.go`)。

第一版我把 `Mode: CombatModeRoom` 寫死在 `beginRoomCombat` 裡面 ——
那會讓**遊蕩怪物與紮營也變成離不開的死戰**。改成由呼叫端傳:
房間 0x82、遊蕩怪物 2、紮營 6;地表紮營在 `beginCombatWith` 之後補 4。

★ 這個形狀值得記:**「共用同一支建構函式」不等於「共用同一組參數」**。
原版把模式當參數傳(`sub_2E364(mode, …)`)正是因為它會變 ——
把它寫死在被共用的地方,就是把一個參數偷偷改成常數。

## 6. 三條測試,其中一條反轉

| 測試 | 內容 |
|---|---|
| `TestEscapeIsRefusedUntilTheBattleIsDecided` | ⚠ **反轉**。前身 `TestFleeLeavesCombat` 釘住「按 ESC 就離開」,而那是引擎自己發明的 |
| `TestDungeonRoomsCanNeverBeEscaped` | 連「勝負已定」都救不了 —— 0x80 那條排在前面 |
| `TestEscapeAlwaysWorksWhenTheWholePartyIsDown` | 0xA0 遮罩;守的是死鎖 |

`TestFleeLeavesCombat` 是本專案**第五次**「測試在量自己的發明」。
與前四次(R/W 分家、picker 繞回、傷害數字、自動替換裝備)同一個形狀:
**註解裡就寫著「原版沒有 X」,而那句話沒有出處。**

⇒ 可操作的教訓:**註解裡的「原版沒有 …」是一句斷言,要有位址**。
沒有位址的否定句,下一輪應該當成 ⬜ 而不是結論。

## 7. 還沒讀的

- ⬜ `sub_B210(n)`(n 為負 = 單位槽、正 = 物件槽)當黑盒:只知道它把東西
  從戰場移除。引擎的 `EndCombat` 直接丟掉整個 `Combat`,等價 —— 但
  「移除順序會不會影響掉落物」未驗。
- ⬜ `byte_41989 = 1`(不論成不成功都設)的語意未追。與 `docs/re/69` 的
  `byte_4198A` 相鄰,疑為同一組畫面重繪旗標。
- ⬜ `sub_27230(0x21)` 與收尾的 `sub_2C188(0x4B0, 0x7D0, 1, 0x28)` 是音效 / 延遲,
  參數含意未追。
- ⬜ **走出戰場邊緣**那條路(`sub_2F294`,印 `escapes!`)引擎有做,
  但沒有對照原版驗過「哪些格算出界」。
