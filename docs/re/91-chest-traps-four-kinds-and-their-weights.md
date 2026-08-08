# 91 — 寶箱陷阱:四種、權重、以及「戰鬥中只有兩種」

| | |
|---|---|
| 輸入檔 | `re_work/fmtowns/WORRIORS.EXP.asm` / `.i64` |
| SHA-256 | 見 `re_work/fmtowns/SHA256SUMS.txt` |
| 主要函式 | `sub_2AB38`(★ 陷阱分派)、`sub_2A464`(單人扣血)、`sub_2AB08`(★ 中毒)、`sub_2A4D0`(★ 全隊扣血)、`sub_2B724`(1..30 骰,已見 `docs/re/15`) |
| 主要資料 | **`byte_5FFEC` = 八筆權重表**、`byte_3E0A3`(地點碼)、`byte_3E06B`(隊伍人數) |
| 起因 | 引擎的寶箱陷阱傷害是**發明的** `random(1,20)`,違反 `CLAUDE.md §3.0`(沒有證據不要實作) |
| 狀態 | ✅ 四種陷阱、權重、戰鬥例外全部定案並落地 |

---

## 0. ★★★ `sub_2AB38` 全文

```
sub_2C598(0x28, 0xBB8, 0x1F4)                  ; 音效(陷阱觸發聲)
if (byte_3E0A3 > 0x7F) ebx = random(0, 1)      ; ★ 地點碼 > 0x7F(戰鬥)只擲兩種
else                   ebx = byte_5FFEC[random(0, 7)]
switch (ebx) {
  0 → 印 "Acid"   ; sub_2A464(random(0, byte_3E06B−1), sub_2B724())
  1 → 印 "Poison" ; sub_2AB08(random(0, byte_3E06B−1))
  2 → 印 "Bomb"   ; sub_2A4D0()
  3 → 印 "Gas"    ; for (i = 0; i < 6; i++) sub_2AB08(i)
  default → 什麼都不做                          ; 走不到,原版留著
}
```

`default` 分支不可能被走到(兩條路徑都只產生 0..3),但它在原版裡確實存在 ——
照 `CLAUDE.md §3.0`「連原版的死碼也還原」,引擎的 `switch` 也不補 `default` 行為。

## 1. ★★★ 權重表 `byte_5FFEC`

dump 出來是八個位元組:

```
byte_5FFEC:  00 00 00 01 01 02 02 03
```

| 值 | 陷阱 | 出現次數 | 機率 |
|---|---|---|---|
| 0 | 酸(Acid) | 3 | 37.5% |
| 1 | 毒(Poison) | 2 | 25% |
| 2 | 炸彈(Bomb) | 2 | 25% |
| 3 | 毒氣(Gas) | 1 | 12.5% |

⇒ **一半的機率(毒 + 毒氣 = 37.5%)根本不扣血**,只改狀態。這一點直接推翻了引擎
原本的模型,也讓 `TestTrappedChestHurtsSomeone`「一定有人扣血」的舊斷言有 3/8 的機率誤判
(它靠 `SeedRandom(4)` 剛好躲過去 —— 固定種子把不確定性藏起來了)。

## 2. ★ 戰鬥中只有酸與毒

`cmp byte_3E0A3, 7Fh; ja` —— 地點碼超過 0x7F 就跳過權重表,改擲 `random(0, 1)`。
`byte_3E0A3` 是地點碼(`docs/re/xx` 已記:大地圖 0、場景 1..0x20、地牢 0x21..、戰鬥 0xFF),
0x7F 這條線把戰鬥切出去。

**為什麼要切**:炸彈與毒氣都是「掃六個隊伍槽」——
而戰鬥中的 `byte_3E06B` 語意不同(戰場上隊員在 `Units[0..5]`),
原作者顯然是為了避開這個歧義。引擎照抄這條線(`locationCode() > 0x7F`)。

## 3. 三支傷害函式

### `sub_2A464(槽, 傷害)` —— 單人扣血

引擎既有的 `damageMember`。傷害值來自 `sub_2B724`。

### `sub_2B724` —— 1..30 骰,與命中骰同一顆

`max(1, random(0, 60) / 2)`,`docs/re/15` 已記,引擎是 `(*State).AttackRoll()`。
⇒ **酸的傷害與武器命中骰、Mani 治療量共用同一顆骰子**。

### `sub_2AB08(槽)` —— 中毒

```
if (槽 >= byte_3E06B)                     return   ; 不在隊伍裡
if (隊員[槽].狀態 == 'D')                  return   ; ★ 死人不中毒
隊員[槽].狀態 = 'P'
```

⚠ 它**只寫入** `'P'` —— 不會把已中毒的人治好,也不會復活死人。
毒氣掃六槽時,死人與空槽都被這兩個 return 濾掉。

### `sub_2A4D0()` —— 全隊扣血(炸彈)

```
for (i = 0; i < 6; i++) {                  ; ★ 上限寫死 6,不是隊伍人數
  if (i >= byte_3E06B)        continue
  if (隊員[i].狀態 == 'D')     continue
  sub_2A464(i, random(1, 8))               ; ★ 每人各擲一次
}
```

兩點值得記:

1. **迴圈上限是寫死的 6**,隊伍不滿時多出來的圈是空轉。行為與 `for i < PartySize` 等價,
   但引擎照抄結構 —— 這種「固定上限 + 內層檢查」的形狀在 U5 裡到處都是,
   統一寫法之後未來對碼比較不會錯開。
2. **每人各擲一次** `random(1,8)`,不是全隊吃同一個值。

## 4. 引擎落地

| 原版 | 引擎 |
|---|---|
| `byte_5FFEC` | `u5data.TrapKindRoll` + `TrapKindRollMax` |
| `random(1,8)` | `u5data.TrapBombDamageMax` |
| `random(0,1)` 的上限 | `u5data.TrapCombatKindMax` |
| `sub_2AB38` 分派 | `(*State).chestTrapVictim` + `trapKind` |
| `sub_2AB08` | `(*State).poisonMember` |
| `sub_2A4D0` | `(*State).bombEveryone` |
| `sub_2C598(0x28, …)` | `s.PlaySFX(u5data.SFXDamage1)` —— ★ 見下 |

### ★ 音效:`0x28` 不是索引,是白噪的 Rate

第一版寫這份筆記時我把 `sub_2C598(0x28, 0xBB8, 0x1F4)` 的 `0x28` 當成
音效索引,然後發現 40 > 23 超出 `U5_SE.TBL` 的行數,於是標了一條 ⬜「未接」。

**那是位置猜錯。** `sub_2C598` 的除錯字串直接寫著參數是什麼:

```
"white_noise\nRate %d\nDura %d\nLimit %d"
```

它是「DOS 版白噪參數 → FM Towns PCM」的**轉接層**,而 `(0x28, 0xBB8, 0x1F4)`
這一組落到 `sub_2C46C(7, 0x3C)` = **`DAME1.SND`(傷害)**。完整的分派表與
其他 25 個呼叫點見 **`docs/re/92`**。

⇒ 教訓與 `CLAUDE.md §4.5` 的「IDA 自動命名不是資料」同一類:
**同一個位置在不同函式裡不是同一個意思**,別把「別處第一參數是索引」搬過來。
真值在函式本體(這裡甚至有現成的格式字串直接說)。

## 5. 測試

| 測試 | 驗什麼 |
|---|---|
| `TestChestTrapKindsFollowTheWeightTable` | 八筆表的組成(3/2/2/1)與長度 |
| `TestCombatTrapsAreOnlyAcidOrPoison` | 戰鬥中只擲 0..1;**含正對照**(非戰鬥時四種都出得來) |
| `TestGasPoisonsTheLivingOnly` | 毒氣不動死人、越界槽不 panic |
| `TestBombHurtsEveryPartySlot` | 每人各擲、上限 8、死人不扣 |
| `TestTrappedChestHurtsSomeone`(改寫) | 陷阱發動必有效果(扣血**或**中毒),且兩類都出現過 |

## 6. 這次踩到的形狀

**固定種子會把「斷言太窄」藏起來。** `TestTrappedChestHurtsSomeone` 原本
`SeedRandom(4)` + 斷言「一定有人扣血」,在舊的發明模型下永遠綠;換成真的四種陷阱之後,
種子 4 剛好擲出不扣血的那種 → 紅燈。**紅燈的成因在斷言而不是實作**(這是第十次)。

修法是把種子掃過 1..40 並放寬到「扣血或中毒」,同時要求兩類都出現過 ——
放寬斷言的同時補上鑑別力,否則「什麼都沒發生」也會綠。
