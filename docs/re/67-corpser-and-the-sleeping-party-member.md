# 67 — `sub_A360`:拖屍怪把人拖下水,以及一個誤名了很久的旗標

| | |
|---|---|
| 輸入檔 | `re_work/fmtowns/WORRIORS.EXP.asm`(FM Towns 英文版) |
| SHA-256 | 見 `re_work/fmtowns/SHA256SUMS.txt` |
| 主要函式 | `sub_A360`(隊員的戰鬥回合,**558 行**)、`sub_BCC4`(掙脫)、`sub_1F840`(命中結果)、`sub_2B724`/`sub_2B710`(門檻骰)、`sub_2ED50`(叫醒) |
| 落地 | `internal/u5data/combatstats.go`、`internal/game/combatai.go`、`internal/game/combat.go` |
| 從哪裡來 | `docs/re/66` 的截斷清單第一名:`sub_A360` **558 行組語 → 5 行 C、34 個字串全掉** |

---

## 1. 清單第一名是什麼

`docs/re/66` 的稽核把 `sub_A360` 排在最前面,而它反編譯出來只有五行。
組語 558 行,是**隊員在戰場上的一個回合**:

```
byte_3E0AE                → 現在輪到第幾號單位
dword_3EF50[8×n] + 2      → 旗標(bit7 隊員 / bit3 睡著 / bit2 被拖下水 …)
dword_3EF50[8×n] + 3      → 隊員:名冊索引;怪物:生物編號
dword_3EF54[8×n] + 2 / +3 → X / Y

印 <名字> ", armed with " <三件武器名> 或 "bare hands"、":"
if (旗標 & 4) { 印 "ARGH!"; sub_BCC4(n); 回合結束 }
if (旗標 & 8) { 擲醒來; 印 "Zzzzz..."; 回合結束 }
讀一個鍵 → 0x5A 路跳表:
    A 攻擊 / C "Cast..." / G "Get-" / J "Jimmy-" / O "Open-" / P "Push-"
    R "Ready..." / S "Search-" / U "Use item" / Y "Yell " / Z "Z-stats..."
    不能用的鍵 → "Cant!"
```

指令回顯那一段與 `docs/re/60` 對得上(戰鬥與大地圖共用同一批字串)。
真正沒人知道的是**上面那兩個旗標分支**。

---

## 2. ★ 旗標 0x04 不是「凍結」,是被拖屍怪拖下水

引擎的 `combat.go` 原本這樣寫:

```go
// UnitFrozen 這一回合不能行動,也不會被選為目標。
UnitFrozen = 0x04
```

那個名字是從「它會讓單位跳過回合」**反推**出來的 —— 觀察到的效果對,語意錯。
設它的地方在 `sub_1F840`(命中結果的報告):

```asm
        test    si, 80h                          ; 目標是隊員?
        jz      short 普通命中
        mov     eax, [ebp+arg_4]                 ; 攻擊者
        cmp     eax, 0FFh
        jz      short 普通命中
        cmp     byte ptr dword_3EF50+3[eax*8], 2Dh   ; ★ 攻擊者的生物編號 = 0x2D
        jnz     short 普通命中
        push    offset aDraggedUnder             ; " dragged under!"
        call    sub_23C18
        …音效…
        or      byte ptr dword_3EF50+2[edi*8], 4 ; ★ 設旗標
        movzx   eax, byte ptr dword_3EF54[edi*8]
        mov     byte ptr dword_3E46C+1[eax*8], 0 ; ★ 顯示 tile = 0(人消失了)
```

**0x2D 是誰?** 怪物名表 `off_3F564` 的第 45 筆(位址 `0x3F564 + 45×4 = 0x3F618`):

```
aCorpser        db 'Corpser',0    ; DATA XREF: .text:0003F618↑o
```

**Corpser —— 拖屍怪。** 而 `internal/i18n/names.go` 早就有這個譯名
(取自 u6 手冊)。**名字本身就在說這件事**,只是沒人把兩邊接起來。

⚠ 順帶一提:`0x2D` 在**物品**命名空間裡是 `ItemAmuletOfTurning`,毫無關係。
本專案已經在這個坑裡跌過好幾次(地形 0xB4–0xB7 是四門砲、物品 0xB4–0xB7 是信物),
所以測試裡有一條專門把兩者釘開。

### 掙脫:`sub_BCC4`

```asm
        call    sub_2B724                             ; 門檻
        movzx   edx, byte ptr dword_3EF50+1[edi*8]    ; ★ 敏捷
        cmp     edx, eax
        jle     loc_BD57                              ; 掙不脫 → 直接回
        印 <名字> " regurgitated!"
        …音效…
        and     byte ptr dword_3EF50+2[edi*8], 0FBh   ; 清掉旗標
        顯示 tile 還原成基礎 tile
```

`unit[+1]` 是**敏捷** —— 與行動倒數 `36 − [ebx+1]` 讀的是同一格,兩處互相印證。

門檻:

```asm
sub_2B724:  sub_2B710(3Ch) → rand(0, 60)
            cdq / sub eax, edx / sar eax, 1     ; 除以 2(有號)
            if (== 0) → 1                        ; ★ 硬補一個下限
sub_2B710(n): return sub_28E14(0, n)
```

⇒ **門檻 = max(1, rand(0, 60) / 2)**,範圍 1..30。
所以**敏捷 31 以上的人一定第一回合就脫身**,而敏捷 1 的人幾乎永遠出不來。
那個 `if (== 0) → 1` 不是多餘的:少了它,敏捷 1 的人有 1/61 機率靠「門檻 0」逃出。

不論掙不掙脫,**這一回合都用掉了**(`sub_A360` 的 `esi = 1`)。

---

## 3. ★ 隊員睡著與怪物睡著是**兩支不同的函式**

| | 函式 | 醒來機率 | 訊息 |
|---|---|---|---|
| 怪物 / AI | `sub_A108` | `rand(0, 16) == 16` → **1/17** | 無 |
| 隊員 | `sub_A360` | `rand(0, 255) < 16` → **1/16** | **每回合都印 `Zzzzz...`** |

```asm
; sub_A360
        test    byte ptr dword_3EF50+2[eax*8], 8
        jz      short 讀鍵
        push    0FFh / push 0
        call    sub_28E14            ; rand(0, 255)
        cmp     eax, 10h
        jge     short loc_A595       ; ★ >= 16 → 跳過叫醒,但**還是印 Zzzzz**
        call    sub_2ED50            ; 叫醒(狀態改回 'G'、顯示 tile 還原)
loc_A595:
        push    offset aZzzzz        ; 'Zzzzz...'
```

★ 關鍵在 `loc_A595` 的位置:它在 `sub_2ED50` **之後**、`push aZzzzz` 那一行。
所以醒來與沒醒來**都會印 `Zzzzz...`,而且都用掉這一回合** ——
醒來只影響**下一回合**能不能動。

引擎原本對所有單位套同一條 1/17,而且**不印那句話** ——
玩家只會看到自己的角色莫名其妙不動,連原因都不知道。

`sub_2ED50` 順帶確認是「叫醒」:它把隊員的狀態位元組寫回 `0x47`('G' 良好),
並把顯示 tile 從躺著換回站著。

---

## 4. 落地

| | |
|---|---|
| `u5data.CreatureCorpserIdx = 0x2D` | 附名表位址的推導 |
| `u5data.CorpserEscapeThreshold(roll)` | `max(1, roll/2)`,常數 `CorpserEscapeRollMax = 60` |
| `UnitFrozen` → **`UnitGrabbed`** | 連同註解改寫成真正的語意 |
| `strugglingUnderwater(u)` | `ARGH!` + 掙脫擲骰 + `被吐了出來!` |
| `corpserGrab(attacker, target)` | 與注視者的凝視同一層:命中後改成施加狀態,不算傷害 |
| `advanceCombat` 的睡眠分支 | 依「是不是玩家操控」分成 1/16 + `Zzzzz...` 與 1/17 兩條 |

五條測試,兩條是專門擋以前踩過的坑:

- `TestCorpserIndexIsFortyFive` —— 擋「同值不同命名空間」(0x2D 的物品義)
- `TestCorpserEscapeThresholdIsAtLeastOne` —— 逐個骰值驗那個下限 1
- `TestOnlyPartyMembersGetDraggedUnder` —— 擋「拖屍怪打怪物也拖下水」
- `TestStrugglingUnderwaterAlwaysCostsTheTurnAndArghs` —— 擋「掙脫了就不印 ARGH!」
- `TestSleepingPartyMemberAlwaysSaysZzzzz` —— 擋回到「兩條路共用一個機率」

## 5. `sub_A360` 還沒讀完的部分

558 行只讀了旗標分支與跳表的字串。還沒追的:

- `, armed with ` 後面那三件武器怎麼串(`sub_A310` 逐件回傳長度,三件相加為 0 才印 `bare hands`)
- `Absorbed!` 的條件:`byte_3E08A != 'N'` 且 `byte_3DFC0 == 0` 且 `byte_3E0A4 == 0x12`
  —— 三個全域都還沒定名,與 `docs/re/57` 的「兩個地點吸魔法」是**不同的閘門**
- 跳表裡沒對到字串的那些 case(A / B 等)
- `sub_1F840` 的其餘命中結果:`' killed!'`、`' slept!'`、`' hit!'`

---

## 追記:`sub_1F840` 整支讀完 —— 原版從不告訴你怪物掉了幾點血

上面的「還沒追」列了 `sub_1F840` 的其餘命中結果。讀完之後那不是三句話,
是一整套**刻意隱藏數字**的回報設計,而引擎原本把它換成了數字。

### 一、命中結果的分派

```
esi = 目標旗標;  byte_3E0B2 = 這一擊的結果旗標
byte_3E0B2 &= 0FEh
if (byte_3E0B2 & 20h) 印 <名字> " grazed!" + 音效
if (byte_3E0B2 & 20h) return            ; 擦傷就到此為止
if (byte_3E0B2 & 02h) return
if (目標旗標 & 20h) { 印 <名字> " killed!"; byte_3E0B2 |= 1 }
if (byte_3E0B2 & 04h) { 印 <名字> " slept!"; return }
if (byte_3E0B2 & 08h) return
印 <名字>
if (目標是隊員) {
    if (攻擊者是拖屍怪) → " dragged under!"(見上文)
    else                → " hit!"           ★ 就這樣,沒有數字
} else {
    等級 = sub_BAFC(目標)                    ★ 傷勢等級 1..4
    1 → " critical!"   2 → " heavily wounded!"
    3 → " lightly wounded!"   4 → " barely wounded!"
}
```

★ **隊員被打只說 `hit!`,怪物被打只給一句形容詞。** 原版一次都沒有印過
「掉了幾點血」—— 玩家判斷敵人狀況的唯一依據就是那四句話。
引擎原本兩邊都印「受了 N 點傷」,等於把原版刻意藏起來的資訊送出去。

四個 case 的對應是 IDA 自己標的(`jumptable 0001F96C case 1..4`),
不必靠猜:`case 1 → critical`、`case 4 → barely`。

### 二、★ `sub_BAFC` 同時是「逃跑判定」

`docs/re/16` 寫著:

> `sub_16454`(能不能走)對**出界**的格子只在逃跑旗標成立時放行:
> 逃跑就是靠走出邊緣完成的,**沒有另外的「逃跑判定」**。

前半對,**後半錯了**。判定就在 `sub_BAFC` 裡,與傷勢等級同一支:

```asm
        movzx   esi, byte_3F055[eax*8]   ; 生命上限
        shr     esi, 2                   ; 上限 / 4
        cmp     esi, [edi]               ; 與目前血量比
        jle     短
        mov     ebx, 1                   ; ★ 血 < 1/4 → critical
        mov     [ebp+var_4], ebx         ;   並標記要逃
        …
        add     esi, esi                 ; 上限 / 2
        cmp     esi, [edi]
        jle     短
        mov     ebx, 2                   ; 血 < 1/2 → heavily
        push    100h / call sub_2B710    ; rand(0, 256)
        cmp     esi, 0FBh
        jle     短
        mov     [ebp+var_4], 1           ; ★ 也有一小段機率要逃
        …
        sar     eax, 1 / lea esi,[esi+esi*2]   ; (上限/4/2)×3
        cmp     esi, [edi]
        jle     短
        mov     ebx, 3                   ; 血 < 3/4 → lightly
        …
        mov     ebx, 4                   ; 否則 barely
        cmp     [ebp+var_4], 0
        jz      短
        or      byte ptr [edi+2], 2      ; ★ 掛逃跑旗標
        jmp     短
        and     byte ptr [edi+2], 0FDh   ; ★ 否則**清掉**它
```

⇒ **打到剩 1/4 血以下的怪物一定會跑**;半血以下每被打一次還有
`rand(0,256) > 251`(約 1/51)的機率跑。而血回到 3/4 以上時旗標會被**清掉** ——
所以「治好自己的怪物會回頭再打」也是原版行為。

⚠ 第三個門檻原版算的是 `(上限/4/2)×3`,**不是** `上限×3/4`。
整數除法先做,兩者在上限不是 8 的倍數時會差一點。照原版的順序算。

### 三、落地

| | |
|---|---|
| `woundLevel(idx)` | 四級判定 + 掛 / 清逃跑旗標,逐行照 `sub_BAFC` |
| `woundReport(idx)` | 四句形容詞 |
| `applyDamage` | 隊員 → `MsgWasHit`;怪物 → `woundReport`。**兩邊都不再報數字** |

- `TestWoundLevelsAreMonotonic` —— 擋「四個 case 接反」(接反的話滿血的怪物
  會被報成「已然垂危」,而沒有測試會自己發現)
- `TestBadlyHurtMonstersFlee` —— 含「滿血時旗標要被清掉」那一半

### 四、還沒追的

- ✅ **0x20(擦傷)找到了**,見下面第五節。
- `byte_3E0B2` 剩下 **0x10** 沒解(`sub_AE20` 設的,那是怪物移動那一支)。
- `sub_2C188` / `sub_2C598` 兩支音效的參數含意。

---

## 五、`byte_3E0B2` 的位元語意:「這一擊的結果,已經報過了」

把所有寫入處掃一遍(14 處),語意就清楚了 —— 它不是狀態,是**回報用的一次性旗標**:

| 位元 | 誰設的 | 意思 |
|---|---|---|
| 0x01 | `sub_1F840` | killed(自己印完之後補設) |
| 0x02 | `sub_B51C` | 已經印過 `" vanishes!"`,別再印 |
| 0x04 | `sub_2EDF8` | slept |
| 0x08 | `sub_B8DC` | 已經印過中毒,別再印 |
| 0x10 | `sub_AE20` | **未解**(怪物移動那一支) |
| **0x20** | `sub_B9A8` 與 `sub_B51C` | **grazed** |

★ **擦傷有兩個來源,而且判準不同**:

```asm
; sub_B9A8(攻擊):減傷之後
        call    sub_B274                 ; 減傷後的傷害
        and     ebx, ebx
        jge     short 照打
        test    al, 80h                  ; ★ 只有目標是隊員時
        jz      short 照打
        mov     byte_3E0B2, 20h          ; 擦傷,而且**跳過扣血**

; sub_B51C(扣血)開頭:對**任何**目標
        cmp     [ebp+arg_4], 1
        jge     short 照扣
        mov     byte_3E0B2, 20h          ; 擦傷
        mov     [ebp+arg_4], eax         ; 傷害歸零
```

引擎原本只有第一條(`dmg < 0 && t.IsParty()` → 「擋下了這一擊」),
**第二條漏了** —— 怪物被減傷吃到 0 傷害時會照樣報一句傷勢等級,
玩家會以為打中了。`TestZeroDamageIsAGrazeForMonstersToo` 釘住這一條。

順帶:那句「擋下了這一擊」就是原版的 `" grazed!"`,不是引擎自己加的仁慈規則 ——
訊息已改成「只被擦過!」以對上原意。
