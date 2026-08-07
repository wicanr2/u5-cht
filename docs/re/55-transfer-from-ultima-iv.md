# 從 Ultima IV 轉入角色 —— 兩塊資料,位移相同

輸入檔:`WORRIORS.EXP`(FM Towns 英文版)
IDA 位址:`sub_71D0`(轉入)、`jpt_7349`(八個職業)、`sub_7594`(Ztats 那一行)
日期:2026-08-08

---

## 0. 這是系列特色,不是可選的裝飾

主選單第三項「Transfer from Ultima IV」。U4 打完的那個角色可以帶進 U5 ——
而**帶進來的東西比想像中少**:三圍、生命、經驗、名字、性別、職業,
加上一個「是不是聖者」的旗標。裝備、金錢、道具**都不帶**
(原版那兩塊 read 裡根本沒有它們的欄位)。

## 1. ★ 讀的是**兩塊**,而且欄位位移相同

```asm
; 第一塊:角色記錄
push [ebp+var_C]                     ; = 8
push 27h                             ; 39 B
push edi                             ; → dword_54300
push offset aAPartySav ; "a:party.sav"
call sub_2C740

...

; 第二塊:決定「是不是聖者」
mov  eax, 140h
mov  [ebp+var_C], eax
mov  [ebp+var_4], offset dword_54328
push eax                             ; = 0x140
push 0B6h                            ; 182 B
push [ebp+var_4]
push offset aAPartySav_0
call sub_2C740
```

然後聖者的判定讀的是**第二塊**:

```asm
mov  eax, [ebp+var_4]        ; ← dword_54328,第二塊
cmp  word ptr [eax+6],  0 ; jnz 不是
cmp  word ptr [eax+8],  0 ; jnz 不是
…                            ; 八個 u16:+6 +8 +0A +0C +0E +10 +12 +14
mov  dword_54498, 1
```

⚠⚠ **兩塊的欄位位移是一樣的**(都是 +6 起),而第一塊的 +6/+8/+0A 正是
力量 / 敏捷 / 智力。我第一次讀的時候把兩者混在一起,結論就變成
「三圍全為 0 才是聖者」—— 而那對**任何合法角色都不成立**,
所以那個結論本身就在說「你讀錯了」。

分辨的依據只有一行:`mov eax, [ebp+var_4]`,而 `var_4 = offset dword_54328`。
**追指標,不要追位移。**

## 2. 第一塊:U4 的角色記錄(offset 0x008,39 B)

| 位移 | 型別 | 內容 | → U5 欄位 |
|---|---|---|---|
| +0x00 | u16 | 生命 | `CharHP`(u16) |
| +0x02 | u16 | 最大生命 | `CharMaxHP`(u16) |
| +0x04 | u16 | 經驗值 | `CharExp`(u16) |
| +0x06 | u16 | 力量 | `CharStrength`(**只取低位元組**) |
| +0x08 | u16 | 敏捷 | `CharDex`(同上) |
| +0x0A | u16 | 智力 | `CharIntel`(同上) |
| +0x0C | u16 | 法力 | `CharMP`(同上) |
| +0x14 | 8 B | 名字 | `CharName` |
| +0x24 | 1 B | 性別 | `CharGender`(0x0B 男,其餘女) |
| +0x25 | 1 B | 職業 0..7 | `CharClass`(查表) |

⚠ 三圍在 U4 存檔裡是 **u16**,搬進 U5 只取低位元組(`mov al, [edi+6]`)。
所以下面那道驗證不只是「合理範圍」——它同時保證截成一個位元組不會失真。

### 驗證(照原版順序與界線)

```
力量 / 敏捷 / 智力  > 0x46(70)   → 拒絕
經驗 / 生命 / 上限  > 0x270F(9999)→ 拒絕
職業碼             > 7           → 拒絕
名字八個位元組:0 是結尾;0 < c < 0x20 → 拒絕
```

⚠ 比的是 `ja`(嚴格大於),所以**剛好等於界線要放行**。
而名字遇到 0 是 `continue` 不是 error —— 短名字合法。

### 八個職業

```asm
jpt_7349:  case 0 → 'M'   1 → 'B'   2 → 'F'   3 → 'D'
           case 4 → 'T'   5 → 'P'   6 → 'R'   7 → 'S'
```

★ M(age) B(ard) F(ighter) D(ruid) T(inker) P(aladin) R(anger) S(hepherd)
—— 正是 **U4 的八個職業**,順序也一樣。八個字母全部命中,這張表沒有滑動的餘地。

⚠ U5 自己的職業只有 A / F / B / M 四種(六名初始角色驗過),但轉入會把
**D / T / P / R / S 也原樣寫進去** —— 原版沒有把它們摺回四種。
「U5 只有四種職業所以要對映」是個很自然的想法,而原版沒有這麼做。

### ★ 等級 = 最大生命 / 100

```asm
movzx eax, word ptr [edi+2]   ; 最大 HP
cdq
mov   ecx, 64h                ; 100
idiv  ecx
mov   [esi+16h], al           ; 等級
```

**不是照經驗值算。** U5 平時用的門檻表(100/200/400/800/1600,見
`game/levelup.go`)在這條路上完全沒被用到 —— 最大 HP 350 的角色轉進來是
等級 3,而它的經驗值 1234 照門檻表會算成 4。測試就用這個差值把它釘住。

## 3. 失敗只有一句話

原版對**所有**失敗印同一組:

```
Error: Your Ultima IV game contains …
Unable to continue transfer.
Press any key to return to the menu.
```

不分原因、不說是哪一欄不對。引擎把原因帶在 Go 的 error 裡給開發用,
但**玩家看到的照原版只有一種**。

## 4. 路徑

原版寫死 `a:party.sav`(U4 的存檔磁碟)。現代環境沒有 A 磁碟,所以引擎用
`-u4save` 指定;**沒指定就在選單裡照實說**,不做「什麼都不做卻關掉選單」的分支
—— 那會讓玩家以為角色轉進來了。

## 5. 引擎對應

| 原版 | 引擎 |
|---|---|
| `sub_71D0` 的第一塊 + 驗證 + 換算 | `u5data.ParseU4Transfer` |
| `jpt_7349` | `u5data.U4Classes` |
| `idiv 100` | `Char.Level = 最大 HP / 100` |
| 第二塊那八個 u16 | `u5data.u4IsAvatar` → `U4Transfer.Avatar` |
| `dword_54498` | `game.State.TransferredAvatar`(Ztats 末行,`docs/re/54` §3) |
| 主選單第三項 | `game.MenuTransferU4` → `(*State).TransferFromUltimaIV` |
