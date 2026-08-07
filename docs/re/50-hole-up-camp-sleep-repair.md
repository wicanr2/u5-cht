# 一個鍵三件事:紮營、睡床、修船

輸入檔:`WORRIORS.EXP`(FM Towns 英文版)
IDA 位址:`sub_2ACF4` case 72 → `sub_2B8CC`(紮營 / 修船)、`sub_16BA0`(睡床)、
`sub_2E8B0`(地表紮營的戰場)、`sub_2E364`(進戰場)
日期:2026-08-08

---

## 0. 為什麼這條漏了這麼久

H 是 U5 **唯一的恢復管道**:回滿 HP、補回法力、而且升級只在休息後發生
(`levelup.go` 的老人)。少了它,玩家一出城就沒得治 —— 而這種缺口
在單元測試裡完全看不出來,因為每一支被呼叫到的規則都是對的。

漏掉的原因是輸入層原本沒有一份完整的指令表(`docs/re/49`)。
指令表全解之後,H 是第一個發現「規則沒寫、鍵也沒接」的。

## 1. 分派:先看載具,再看地點

```
sub_2ACF4 case 72 ('H'):
  if (地點 == 0 || 地點 > 0x20) → sub_2B8CC()      ; 地表 或 地牢
  else                                              ; 城鎮 / 城堡(1..0x20)
      tile = 腳下
      print "Hole up- "
      if (tile != 0xAB) → "Only in bed!"           ; 0xAB = 床
      else → sub_16BA0()
```

而 `sub_2B8CC` 開頭又先看載具:

```
if ((載具 & 0xF8) == 0x20) → 修船                  ; 0x20..0x27 = 揚帆中或大船
else                       → 紮營
```

⇒ 順序是 **載具 > 地點**。在城裡的碼頭上船照樣是修船,不是「城裡不能紮營」。

## 2. 修船

```
print "Hole up & \nrepair...\n\n"
if (載具 < 0x24) { print "Sails must be\nlowered!\n\n"; return }   ; 揚帆中
for (i = 0; i < 5; i++) {
    sub_2CC8C(); sub_2E24();                       ; 動畫 + 一個世界回合
    if ((byte_3E08C & 0xFC) != 0x24) return;       ; 中途不再是大船 → 中止
    sub_29304(5);                                  ; 時鐘 +5 分
}
do {                                               ; ★ do-while
    hull += random(1,3);
    if (hull > 99) hull = 99;
} while (hull < 10);
print "Hull now <n>!\n\n"
```

★ 那是 **do-while**,不是 while:所以

1. **一定至少加一次**;
2. 破到 1 的船修一輪就會回到 10 以上,直接能航行。

寫成 `while (hull < 10) hull += …` 的話,耐久 50 的船按 H 什麼都不會發生 ——
而原版會加 1..3。差別很小,但那是「修船有沒有用」的差別。

⚠ 每一輪都重新確認還在大船上:修船途中被打下船,剩下的回合不跑。

## 3. 紮營

```
print "camp!\n\n"
if (地點 < 0x21) {                                  ; 只有地表檢查這兩條
    tile = 腳下
    if (tile != 0 && tile < 4) { print "On land or ship!\n\n"; return }
    if (載具 != 0x1C)          { print "On foot!\n";           return }
}
print "For how many hours? (1-9) "                  ; 收 '0'..'9' 與空白
if (空白 || '0') return
hours = 鍵 − '0'
able = 狀態為 'G' 或 'P' 的人數
if (able > 1) {
    print "\nWilt thou set a watch? "               ; Y / N
    if (Y) {
        print "Who will stand guard? "
        watch = 選一個人
        if (watch < 0 || 狀態[watch] != 'G') { watch = −1; print "None posted!\n\n" }
    } else watch = −1
} else watch = −1
```

三個容易寫錯的地方:

- **水的判定寫的是 `tile != 0 && tile < 4`** —— 深水(0)竟然過得了這一關,
  只有淺水 1..3 被擋。看起來是 off-by-one,但在船上這一段根本到不了
  (前面就分去修船),所以踩不到。**照抄,不「修好」它。**
- **兩條限制只在地表檢查。** 地牢裡騎著馬也紮得起來。
- **「能動」與「守得了夜」用不同的集合**:算人數時 'G' 與 'P'(中毒)都算,
  但真的派得上的只有 'G' —— 選了中毒的人會被回「None posted!」。
  不是筆誤,是兩個判定。

## 4. 睡覺的時間算法有個 off-by-one

`sub_16BA0`:

```asm
movzx eax, byte_3E08F      ; 現在的小時
lea   edi, [ebx-30h]       ; 要睡的小時
add   edi, eax
cmp   edi, 17h
jle   short loc_16C0C
sub   edi, 17h             ; ★ 減 23,不是 24
```

⇒ 22 時睡 4 小時 → 26 → 26 − 23 = **3 時**,而正確的環繞是 2 時。
**跨過午夜會多醒一小時。**

這是原版的 bug。CLAUDE.md §3.0 要求機制與原版一模一樣,所以照抄並用測試釘住 ——
「順手修好」會讓時間對不上原版,而那種差異只有並排跑才看得出來。

之後的流程:

```
先跑 16 個 NPC 回合(sub_9690 + sub_29D64),若 byte_3EDD0 == 'a' 中止
全隊 'G' → 'S'(睡著)
print "Zzzzzz...\n"
音樂切成 4 號(sub_3181C(4))
while (小時 != 目標) {
    sub_29304(10)                          ; 每次推 10 分鐘,不是一次加完
    if (小時變了 && (小時 == 20 || 小時 == 5)) sub_324()   ; 兩個時刻的事件
    if (sub_2B360(x, y, level)) { 被打斷; break }
}
if (被打斷) print "Thrown out of bed!\n"
全隊 'S' → 'G',套用恢復
```

時間是**一次 10 分鐘慢慢推**的,因為途中要讓 NPC 走、讓事件觸發。
一次 `AdvanceTime(hours*60)` 會跳過整個晚上的世界變化。

## 5. ★ 紮營就是進一個專用戰場

`sub_2B8CC` 的尾巴:

```asm
cmp  byte_3E0A3, 20h
jbe  short loc_2BB9E
mov  byte_3E0B1, 6          ; 地牢:模式 6
call sub_FE48
call sub_2E364              ; (6, 守夜的人, 時數)
...
loc_2BB9E:                  ; 地表
call sub_2E8B0              ; → sub_2E51C(0) + sub_2E364(4, 守夜, 時數)
```

⇒ 這解掉了 `docs/re/48` §8 列的未完項:那個 `byte_3E0B1 & 4` 的
「另一種戰場」**就是營地**。所以它那組入場位置(`byte_418E8`/`byte_418F0`
散開的隊形、`byte_41950`/`byte_41960` 的敵人位置)畫的是**營火四周**,
而「怪物種類重抽而不是沿用遭遇到的那隻」也就說得通 ——
紮營時本來就沒有「遭遇到的那隻」,是野獸自己找上來。

⬜ `sub_165C8`(模式 4 的解算:要不要被襲擊、守夜的人有沒有發現)還沒讀。
`internal/game/holeup.go` 目前紮營一定睡得安穩,並在該處標明是**已知落差**,
不是「紮營做完了」。**沒有證據就不補一個機率。**

## 6. 引擎對應

| 原版 | 引擎 |
|---|---|
| `sub_2ACF4` case 72 | `game.(*State).HoleUp` |
| `sub_2B8CC` 前半 | `game.(*State).repairShip` |
| `sub_2B8CC` 後半 | `game.(*State).canCampHere` / `askWatch` / `camp` |
| `sub_16BA0` | `game.(*State).sleepInBed` / `restHours` |
| `sub edi, 17h` | `game.holeUpWrapAt`(照抄的 off-by-one) |
| `"For how many hours? (1-9)"` | `game.(*State).AskNumber`(新增的通用數字提問) |
| `"Wilt thou set a watch?"` | `game.(*State).Ask`(新增的通用 Y/N 提問) |
