# 69 — 大地圖上的砲彈原本什麼都不會發生

| | |
|---|---|
| 輸入檔 | `re_work/fmtowns/WORRIORS.EXP.asm`(FM Towns 英文版) |
| SHA-256 | 見 `re_work/fmtowns/SHA256SUMS.txt` |
| 主要函式 | `sub_172C4`(F 開砲,**244 行**)、`sub_2B3DC`(查某一格上的物件)、`sub_2A4D0`(全隊傷害)、`sub_2E0`/`sub_218`/`sub_268`(NPC → 逮捕) |
| 落地 | `internal/game/fire.go` |
| 從哪裡來 | `docs/re/66` 的截斷清單:`sub_172C4` **244 行 → 3 行 C、四個字串全掉** |

---

## 1. 症狀:BOOOM 之後什麼都沒有

`fire.go` 的 `fireToward` 原本是這樣:

```go
dx, dy := d.Delta()
s.Log(MsgBooom)
// 砲彈沿著那個方向飛,打到第一個擋住的東西為止 —— 與投射物同一條路。
s.FlyProjectile(0, s.X+dx*cannonRange, s.Y+dy*cannonRange)
```

而 `FlyProjectile` 的第一行是:

```go
c := s.Combat
if c == nil {
    return -1, tx, ty
}
```

⇒ **大地圖上開砲只會印一句 BOOOM。** 砲彈不飛、什麼都打不到。
註解寫著「與投射物同一條路」—— 那句話在戰場上成立,在大地圖上完全不成立,
而**沒有任何測試會發現**:訊息印出來了、函式回傳了、沒有錯誤。

原版的 `sub_172C4` 有 244 行,前半是「四鄰哪一格有大砲」(這部分引擎做對了),
**後半是砲彈的飛行與後果**,而那 200 行被 Hex-Rays 截斷成三行。

---

## 2. 砲彈的飛行

```
迴圈最多 5 格(var_2C = 5),每一格:
    kind = sub_2B3DC(x, y, 樓層)          ; 從槽 31 往下掃,回傳物件的種類
    if (kind != 0) → 判是不是有效目標
    else {
        tile = sub_DB10(x, y)
        if (tile ∈ 七種門) { var_24 = 1 }  ; 打到門 → 跳出
    }
迴圈結束後:
    if (var_24) { 印 "Door destroyed!"; 那一格寫成 0x44; byte_4198A = 1 }
    if (var_28 && var_20 != 0) {
        sub_2B6C8(0,0,0,0,0,0, var_20)     ; ★ 整個槽清掉
        byte_4198A |= 2
        if (byte_3E098 > 5) byte_3E098 -= 5 else byte_3E098 = 0   ; ★ 業報 −5
        n = sub_2E0(var_20)                ; 那個槽對應的 NPC
        if (n != -1) { sub_218(n); sub_268(n) }                   ; ★ 逮捕
    }
    if (var_28 && var_20 == 0) sub_2A4D0() ; ★ 打到槽 0 = 打自己 → 全隊受傷
```

`sub_2B3DC(x, y, floor)` 讀完了:從槽 31 往下掃,比對 `+2`/`+3`(X/Y)與 `+4`(樓層),
回傳該物件的**種類**並把槽號留在 `word_3E086`。找不到回 0。

### 四個後果

| 打到 | 結果 |
|---|---|
| 七種門 | `Door destroyed!`,那一格變成 **0x44 磚地**(與 An Ylem 寫的是同一個值) |
| 有效目標 | **整個物件槽直接清掉** —— 沒有血量、不進戰鬥。一砲一個 |
| 有效目標而且是個 NPC | 加上**被逮捕** |
| 槽 0(隊伍自己) | **全隊受傷**,而且**不扣業報** |

★ **業報 −5**(`byte_3E098`,`docs/re/06` 已定名),下限 0
(`cmp al,5; jbe → mov byte_3E098, 0`)。所以拿砲轟東西是有德行代價的。

### 七種門

```
0x97 0x98 奇怪的門      0x99 柵門
0xB8 木門   0xB9 上鎖的門   0xBA 有窗戶的木門   0xBB 有窗戶的上鎖的門
```

七個都用 `look#<tile>` 查得出名字,而且七個都是門 —— 一次全對。

---

## 3. ★ 有效目標的判準,以及一道死碼

```asm
loc_1751C:
        cmp     edi, 1Ch
        jl      short loc_17539          ; kind < 0x1C   → 去查特例
        mov     eax, edi / and eax, 0F8h
        cmp     eax, 78h
        jz      short loc_17539          ; kind & F8 == 78 → 去查特例
        mov     eax, edi / and eax, 0FCh
        cmp     eax, 2Fh
        jnz     short loc_17547          ; ★ 這道永遠成立 —— 見下
loc_17539:
        cmp     edi, 10h
        jz      short loc_17547          ; ★ 馬
        cmp     edi, 11h
        jnz     loc_1747D                ; 都不是 → 繼續飛
loc_17547:
        var_28 = 1;  var_20 = word_3E086 ; 記下槽號
```

⇒ 判準是:

```
kind == 0x10 或 0x11  → 是目標(★ 馬,編號比 0x1C 小卻明文放行)
kind <  0x1C          → 不是
kind & 0xF8 == 0x78   → 不是
其餘                  → 是目標
```

★ `kind & 0xF8 == 0x78` 涵蓋 **0x78..0x7F**,而
`0x78 = 14×4 + 0x40`(**黑刺**)、`0x7C = 15×4 + 0x40`(**不列顛王**)——
兩位打不掉。這與 `docs/re/17` 的「`sub_189BC` 只認黑刺 / 不列顛王 / 暗影領主
三個編號免疫法術抗性」是同一個設計思路的另一個出口。

⚠ **`and eax, 0FCh; cmp eax, 2Fh` 是死碼**:`& 0xFC` 會把低兩位清成 0,
結果永遠不可能等於 0x2F(低兩位是 `11`)。所以那道 `jnz` 永遠成立。
照**實際行為**實作(船 0x2C..0x2F 照樣打得掉),不照抄意圖 ——
同 `docs/re/61` 酒館關鍵字那個「兩個分支落到同一個 `mov`」的死碼。
`TestCannonCannotShootBlackthornOrLordBritish` 有一段專門把這件事釘住:
若哪天有人「照著註解」把船排除掉,那條會紅。

---

## 4. 落地與驗收

| | |
|---|---|
| `fireToward` | 戰場走 `FlyProjectile`;大地圖走新的 `fireCannonball` |
| `cannonDoors` | 七種門 |
| `cannonTargets(kind)` | 判準三條 + 馬的特例 + 死碼的說明 |
| `fireCannonball(dx, dy)` | 最多 5 格,每格先問物件再問門 |
| `cannonHit(slot, x, y)` | 清槽 / 業報 −5(下限 0)/ 打到人就逮捕 / 打到槽 0 就全隊受傷 |

四條測試:

- `TestCannonballDestroysDoors` —— 七種門逐一驗,而且草地不會被當成門
- `TestCannonCannotShootBlackthornOrLordBritish` —— 0x78..0x7F 全擋、馬要放行、
  **船要打得掉**(那道死碼)
- `TestCannonHitCostsKarma` —— 五組業報值,含下限 0 的三種寫法
- `TestShootingYourOwnSlotHurtsTheParty` —— 打自己**不扣業報**

## 5. 還沒讀的

- `byte_4198A` 的位元語意(門那條設 1、打中目標那條 `or 2`)。
- `sub_20CB4`(畫砲彈軌跡)只知道是演出。
- `sub_2E0`(槽號 → NPC 索引)當黑盒用;引擎改用既有的 `NPCAt(x, y)`,
  兩者在「同一格上有物件也有 NPC」時是否等價未驗。
- 迴圈裡 `sub_2B64C(byte_3E161..163)` 那一段(疑為砲彈的動畫起點)沒追。
