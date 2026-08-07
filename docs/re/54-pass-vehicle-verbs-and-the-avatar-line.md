# 空白鍵、載具動詞,以及「他並不是聖者」

輸入檔:`WORRIORS.EXP`(FM Towns 英文版)
IDA 位址:`sub_2ACF4` case 32(空白鍵)、`sub_7C0`(載具動詞與朝向)、
`sub_7594`(Ztats 的聖者那一行)、`sub_71D0`(從 Ultima IV 轉入)
日期:2026-08-08

---

## 1. 空白鍵**不只是**跳過一回合

```asm
loc_2AE01:                       ; case 32(空白)
    cmp   byte_3E0A3, 0          ; 在地表?
    jnz   short loc_2AE26
    cmp   byte_3E167, 0          ; 揚著帆?
    jz    short loc_2AE26
    push  offset aSheetsInIrons  ; "Sheets in irons!\n"
    mov   byte_3E167, 0          ; ★ 收帆
    jmp   short loc_2AE30
loc_2AE26:
    push  offset aPass_3         ; "Pass\n"
```

⇒ **在海上揚著帆時,空白鍵是收帆。** 只有不在那個狀態下才是 Pass。

寫成「空白鍵就是跳過」的話,玩家在海上**找不到收帆的辦法** ——
Y(Yell)那條路是**放**帆,不是收。這種缺口不會讓程式壞掉,
會讓玩家卡在海上。

引擎的「揚帆」狀態就是載具碼落在 `VehicleSailing`(0x20..0x23);收帆是回到
`VehicleShip`(0x24..0x27)並**保留朝向那兩位元**。

## 2. 載具動詞與朝向(`sub_7C0`)

```
載具 & 0xFC == 0x10  馬     → print "Ride ";  dir 1 → 0x12、dir 3 → 0x13
              0x14  魔毯   → print "Fly ";   dir 1 → 0x14、dir 3 → 0x15
              0x28  小艇   → print "Row "  ┐
              0x20/0x24 船 → 不印動詞      ┴→ 載具 = (載具 & 0xFC) + dir
其餘(步行)                → 不印
```

★ 兩種朝向規則**不一樣**:

- **馬與魔毯只在東西向換圖**(原版只比 `dir == 1` 與 `dir == 3`),往南北走保持原樣。
- **船與小艇是四向全換**,低兩位元直接放方向碼。

寫成一條(例如一律 `(載具 & 0xFC) | dir`)會讓馬在往北走時變成另一格圖。

### 這不只是「印一個字」

朝向那兩位元同時是:

- `isBroadside`(開砲判舷側,`docs/re/45`)讀的船首方向
- `ModeOf(mover)`(通行判定,`docs/re/47`)的索引來源
- `skiffAllows` 的河道流向位元(`docs/re/51` §3.3)

**不更新朝向,船的舷側就永遠算錯。** 一個看起來只影響訊息欄的函式,
實際上餵了三個規則。

## 3. 「並非聖者」是原版行為,不是缺預設值

`sub_7594`(Ztats 畫面)最後一段:

```asm
push offset byte_3DDB4       ; 名字
call sub_23C18
push offset aIs              ; " is "
call sub_23C18
cmp  dword_54498, 0
jz   short loc_7804
push offset aAnAvatar        ; "an Avatar."
jmp  short loc_7809
loc_7804:
push offset aNotAnAvatar     ; "not an Avatar"
```

`dword_54498` **只有一個寫入點** —— `sub_71D0`,也就是**從 Ultima IV 轉入角色**:

```asm
mov  eax, [ebp+var_4]        ; U4 存檔
cmp  word ptr [eax+6],  0 ; jnz 不是
cmp  word ptr [eax+8],  0 ; jnz 不是
…                            ; 共八個 word:+6 +8 +0A +0C +0E +10 +12 +14
cmp  word ptr [eax+14h], 0 ; jnz 不是
mov  dword_54498, 1          ; ★ 八個全為 0 才算聖者
```

⇒ 沒有轉入過就一直是 0,**新建的角色一律「並非聖者」**。

⚠ 這條很容易寫錯成「主角當然是聖者」。而 U5 的開場正是
「你回來了,但這一次不是以聖者的身分」—— 那句話有它的意義,
而 Ztats 上那一行就是它的機制化身。

⬜ 轉入本身(`sub_71D0`)還沒實作,所以旗標目前恆為 false ——
但那**不是預設值錯了**,是原版在沒有轉入時的同一個結果。

## 4. 順手清掉的六條過期記載

用程式當真值(`rulebook/63`)逐條核對之後,以下六條 `WORKLIST` 的 ⬜ 是過期的:

| 項目 | 實際狀況 |
|---|---|
| 等級與經驗 | `game/levelup.go` 早就有門檻表與「升級是事件」 |
| 性別 | `u5data.GenderMale/Female`,建角寫入、Ztats 讀取 |
| Ztats 畫面 | `docs/re/41` 就做完了(Z / 翻頁 / ESC) |
| 材料 reagents | `u5data.ReagentNames` 八種 + 價目 + 配方遮罩 |
| 商店店名 | `docs/re/08`/`10`;47 家店只有 46 個店名是原版資料如此 |
| NPC 職業 | 與生物名表是**同一批名字**(索引 4..13),`i18n/names.go` 有中譯 |

> 這是本輪第 N 次同一件事:`WORKLIST` 的 ⬜ 有相當比例不是「沒做」,
> 而是「做了但沒回頭改記載」。所以每一輪都該先核對,不要照著 ⬜ 動手。
