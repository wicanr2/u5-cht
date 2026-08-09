# 100 — 最後三支、崩塌入口的那道門,以及三個存檔位移

| | |
|---|---|
| 輸入檔 | `re_work/fmtowns/WORRIORS.EXP.asm` / `.i64`;`gamedata/INIT.GAM` / `SAVED.GAM`(位移驗證)|
| SHA-256 | 見 `re_work/fmtowns/SHA256SUMS.txt` |
| 主要函式 | ★ `sub_2D72C`(Enter 的 tile 分派)、`sub_2D564`、`sub_27D24`(讀檔序列)、★ `sub_92C0` + `sub_89EC` + `sub_8BA0`(NPC 走樓梯)、★ `sub_2D944`(黑暗格)、★ `sub_BE5C` + `sub_BD5C` + `sub_BD80`(過場動畫解譯器)|
| 工具 | `tools/refunc.py` |
| 起因 | 使用者指示:完成剩下三支、探查崩塌入口的機制、把兩個存檔位移做出來 |
| 狀態 | ✅ 三支全解(兩支落地、一支歸 P6)、入口的門找到並實作、三個位移釘死並驗證 |

---

## 1. ★★★ 崩塌的入口進不去 —— 那道門在**分派表的定義域**上

`docs/re/99` §5e 留了一條 ⬜:「崩塌的入口進不進得去還沒查」,因為
`sub_2D564`(真正的進入)只查座標、載具與末日的三個位元組,**沒有封印判斷**。

答案在它的**呼叫端**。`sub_2D72C`(Enter 指令)是一個 tile 分派器:

```asm
call    sub_DB10                ; 腳下的地形
movzx   esi, byte ptr [eax]
lea     edx, [esi-10h]          ; switch 47 cases
cmp     edx, 2Eh
ja      def_2D765               ; ★ 定義域外 → default
jmp     cs:jpt_2D765[edx*4]
```

| case | tile | 動作 |
|---|---|---|
| 22 | 0x16 洞穴 | 印 "Cave" → `sub_2D564` |
| 23 | 0x17 礦坑 | 印 "Mine" → `sub_2D564` |
| 24 | 0x18 地牢 | 印 "Dungeon" → `sub_2D564` |
| 25 | 0x19 聖壇 | → `sub_1DA10` |
| default | 其餘 | 印 **"What?"** 並回 0 |

⇒ 崩塌的入口是 **0xDF**,而 `0xDF − 0x10 = 0xCF > 0x2E` ⇒ **落到 default**。
**按 E 只會印「What?」。** 門就是 tile 本身,不需要任何額外的旗標檢查。

### 引擎落地

引擎此前先查 `u5data.DungeonAt(x, y)`(**座標**),所以地形雖然顯示崩塌、
人還是走得進去。改成先看 tile:

```
腳下 == TileDungeonSealed              → 印「What?」
座標對上地牢 但 tile 不是三種入口之一  → 印「What?」
```

⚠ 第二條看起來多餘,但它擋的是「地形被改成別的東西之後」(月門、地震)
座標還對得上的情況 —— 原版是 tile 分派,座標只是第二道。

### ★ 防 soft-lock 的整條玩家路徑已驗

`TestWordOfPowerOpensTheWayThrough` 走完:
**崩塌 → 按 E 進不去 → 站在入口旁喊那個字 → 地形變回原樣 → 站上去按 E 進得去**,
最後再喊一次確認 XOR 是對稱的(封回去、又進不去)。

⚠ **喊力量之言要站在入口「旁邊」,進入要站在「上面」。** 原版掃的是視窗
緩衝裡玩家那一格的 −1 / +32 / +1 / −32(`docs/re/26` §3.1),四個鄰格。
兩件事的站位不同 —— 第一版測試站在入口上喊,得到「毫無效果」。

## 2. 三個存檔位移(`sub_27D24` 的讀取序列)

`sub_27D24` 前段是**一格一格 `fgetc`**、後段是 `fread(ptr, n, 1, f)`,
而它全程用 `ebx` 累加已讀位元組數 —— 所以位移可以從已知的錨點推出來。

| 全域 | 位移 | 長度 | 推導 | 驗證 |
|---|---|---|---|---|
| `byte_3E08B` 指定行動者 | **0x02D5** | 1 | `byte_3E08C` = 0x02D6 減 1 | `INIT.GAM` 是 **0xFF**,正是它的哨兵值 |
| `byte_3E09B` 回合計數 | **0x02E5** | 1 | `byte_3E098`(業報)= 0x02E2 加 3 | `INIT.GAM` 是 **0** |
| `byte_3E0F0` 房間清除 | **0x033A** | **14** | `byte_3E0E8` = 0x0332 加 8 | 全 0(開局沒清過房間)|

三個錨點交叉一致:`byte_3E08C` = 0x02D6、`byte_3E08F` = 0x02D9、
`byte_3E098` = 0x02E2 —— 差值與位址差完全相同。

### ★★ 14 這個長度本身是一條證據

索引是 `DungeonRoomBlock(地點碼)*16 + 房號`,而 `DungeonRoomBlock` 有一個
「≥1 就 −1」的修正 ⇒ 八座地牢只佔 **7** 個區塊 ⇒ 7 × 16 = 112 位元
= **剛好 14 位元組**。若索引是 0..7 就會需要 16。

⇒ 原版 `push 0Eh` 獨立佐證了 `docs/re/99` §5d 那個「0x21 與 0x22 共用區塊」
的怪處。這是它的**第三個**來源(前兩個是 `sub_1056C` 的形狀與 `DUNGEON.CBT`
的房間查表)。

### 順手修掉一個擋路的驗證

`Save.validate` 用 `s.Location > len(Locations)`(= 32)當上限,
而**地牢的地點碼是 0x21..0x28** ⇒ **在地牢裡存不了檔**。
改成 `LocationCodeMax = DungeonLocationBase + DungeonCount − 1`。
地點表其實有 40 筆,後 8 筆就是地牢(`dungeon.go` 早就記著)。

## 3. `sub_92C0` + `sub_89EC` + `sub_8BA0` —— NPC 換樓層要先走到樓梯

`sub_9690`(NPC 排程推進)模式 6 / 7:

```
if (sub_92C0(npc, 排程槽))     → 已經站在對的樓梯上 → 換層
if (這回合已經找過)            → return
目標 = sub_89EC(x, y, 要往上, npc)
if (目標) sub_8BA0(…)                   ; 往那裡走一步
if (走不動) return                      ; ★ 這一回合不動,下回合再試
```

而 `sub_89EC(x, y, up, npc)` = `sub_8BA0(x, y, 0, 0, up ? −1 : −2, npc)` ——
同一支 BFS,用 **−1 / −2 當「目標是最近的上 / 下樓梯」的特殊碼**。

`sub_92C0(npc, 槽)`:

```
tile = tileAt(該 NPC 的物件座標)
if (byte_3E0A5 /* 當前樓層 */ >  byte_3E579[npc*16 + 槽] /* 排程樓層 */)
     result = (tile == 0C9h)          ; 要往下
else result = (tile == 0C8h)          ; 要往上(★ 相等也走這條,`jle`)
return result || (tile & 0F4h) == 0C4h
```

### ★★ 順手抓到一個名字反了的常數

| 檔案 | 名字 | 值 |
|---|---|---|
| `sceneset.go` | `LadderUp` | 0xC8 |
| `sceneset.go` | `LadderDown` | 0xC9 |
| `tileflags.go`(舊)| `TileStairsDown` | **0xC8** |
| `tileflags.go`(舊)| `TileStairsUp` | **0xC9** |

兩個檔案給同一組值取了**相反的名字**。判準有兩個獨立來源:
`ClimbDelta(0xC8)` 回 **+1**(往上),而 `sub_92C0` 在「要往上」時找 **0xC8**。
⇒ 以 `sceneset.go` 為準,`tileflags.go` 的兩個改成別名。

### ⚠ 遮罩是 `0xF4` 不是 `0xFC`

`sub_92C0` 的 `(tile & 0F4h) == 0C4h` 收 **0xC4..0xC7 與 0xCC..0xCF**
(bit 3 沒被遮住);而 `u5data.StairsFacing` 用的 `StairsMask` = 0xFC
只收 0xC4..0xC7。**兩支的定義域不同,不要互相套用。**

### 已知差異

原版的「最近」是 BFS 的**步數**;引擎的 `findPath` 只吃座標,所以先用
**曼哈頓距離**挑一格再尋路。差別會在「隔著牆更近的那座樓梯」上顯出來。
要完全一致得把 `findPath` 改成支援「目標是一個謂詞」。

## 4. `sub_2D944` —— 站在「黑暗」格上視野歸零

大地圖主迴圈 `sub_2D9D0` 的第一件事:

```
tile = tileAt(隊伍X, 隊伍Y)                  ; ⚠ 反編譯寫成 sub_DB10(0, 0)
if (tile == 0FFh && byte_3E08A != 0Eh) {
    byte_3E0B5 = 0                           ; ★ 視野平方半徑歸零
    if (arg_0 == 0) { 重畫; return 1 }
    return arg_0
} else {
    sub_29304(0)                             ; 推 0 分鐘 → 重算光照
    return 0
}
```

兩件事讓它可以定案:

- **tile 0xFF 是 look 表第 255 筆「darkness!」**,而它在 `UNDER.DAT` 出現
  **106 次**、在 `BRIT.DAT` **一次都沒有** ⇒ 幽冥界專屬的格子。
- **`byte_3E0B5` 是視野平方半徑**(`docs/re/31` 已解:`sub_29304` 每分鐘算它)。

⇒ 站上去伸手不見五指,而且**連火把與 In Lor 都不補** —— 原版是直接寫 0,
不走那條「不足才補」的鏈。

### ⚠⚠ 只在大地圖上判 —— 因為 0xFF 有兩個意思

引擎的 `TileAt` 在**場景資料缺失或座標出界**時回 `u5data.TileBlank`,
而那**也是 0xFF**。少了「只在大地圖」這道閘門,任何沒載入場景的狀態
(單元測試、剛建好的 `State`)都會變成全黑,而症狀看起來像「光照公式壞了」。
第一版就是這樣紅的(`TestUndergroundAndAraratAreAlwaysDark`)。

⇒ 閘門條件用 `onOverworld()`(不在場景 / 戰鬥 / 地牢裡)—— 那正是
`sub_2D9D0` 成立的條件,三個主迴圈互斥(`docs/re/81`)。

⬜ `byte_3E08A != 0x0E` 那個豁免:`byte_3E08A` 平常放模式字母('T' 停止時間、
'Q' 速度加倍),0x0E 是誰寫進去的還沒掃。**照字面比對,不猜語意。**

## 5. `sub_BE5C` —— 過場動畫的腳本解譯器(全解,歸 P6)

九個 opcode:

| opcode | 動作 |
|---|---|
| 0 | 結束 |
| 1 | 進入「同時移動兩個物件」模式 |
| 2 | 讀一個位元組當**重複次數** |
| 3 / 4 | 開 / 關「每一步都重畫」|
| 5 | 讀一個位元組 → `sub_2B740(n)` 延遲 |
| 6 | 讀三個位元組(值, x, y)→ `byte_3F8F4[y*32 + x] = 值` |
| 7 | `sub_29D64()` 重畫 |
| 8 | `sub_BD5C(n)` 等 n 幀,然後把次數重設成 1 |
| 9 | 讀一個位元組(槽號)→ 清掉那個物件的種類與 tile |
| **≥ 0x10** | `sub_BD80(opcode)` → 取得物件槽與位移,跑 n 次把位移加上去 |

兩支輔助:

- `sub_BD5C(n)` = 等 n 幀(`sub_2C118` 垂直同步 + `sub_2B740` 延遲)
- `sub_BD80(opcode)` = 把 opcode 拆成 **(物件槽, dx, dy)**:
  `opcode & 0xFC` → 槽 6 / 7 / 8 / 0 / 1(**五個演員**),
  `opcode & 3` → 四個方向,結果寫進 `word_3E086` / `word_3E088`

呼叫鏈:`sub_1884`(被逮捕)→ `sub_C414` → `sub_C318` → `sub_C2BC` / `sub_C2D0`
→ `sub_BE5C`。⇒ 它演的是**逮捕 / 押送**那一段(衛兵把你拖走)。

⇒ **不是遊戲邏輯,是動畫。** 逮捕流程的三句文字引擎早就有
(`MsgSubdued` / `MsgDraggedAway`,`blackthorn.go`),機制上沒有缺口。
要演出來需要一個引擎目前沒有的**過場播放器**(逐幀移動物件 + 重畫),
那是 P6 的決定而不是 P4 的補完。規格記在這裡,接的時候不必再逆向。

⚠ 與 `sub_3007C`(開新遊戲的腳本解譯器,`WORKLIST §5.2b`)是**兩支不同的
解譯器**,opcode 表不通用。

## 6. 順手驗掉一件事:NPC 對白是完成的

使用者問「對白翻譯似乎還沒完成?」。查證結果:**完成**,而且**遊戲路徑
真的輸出中文**。

| 量測 | 結果 |
|---|---|
| `.TLK` 工作單 | 1713 段,已翻 **1712(99.9%)**,問題 0 條 |
| 唯一沒翻的一段 | `TOWNE.TLK#23#e14`,原文是 4 字殘片 `caug` —— **刻意不翻**(`talk_b03.go` 檔頭寫了理由:查不到就退回原文,顯示 `caug` 比顯示「[資料損毀]」對玩家好)|
| 端到端(新測試) | **135 段對話開場全部掃過,0 行英文** |
| 端到端(新測試) | **1307 個關鍵字回應,0 行英文** |

★ **135 就是全部有對白的 NPC 數**(`.TLK` 四檔的相異 NPC:CASTLE 40 +
KEEP 32 + TOWNE 48 + DWELLING 15)⇒ 覆蓋率 100%,而那個數字寫成斷言了。

### ★ 為什麼要補這兩條測試

i18n 那一層早就有測試(1712 個 key、查得到、無重複),但那驗的是**表**。
玩家看到的是**遊戲路徑**:`beginConversation` → `trTalk` → `Log`。
中間任何一環(檔名對不上、`Conv.ID` 取錯、忘了呼叫 `trTalk`)都會讓
表全綠而畫面全英文。⇒ 這是 `rulebook/65`「測試綠 ≠ 完成」的具體一例。

### ⚠ 偵測器的豁免是實測逼出來的

第一版偵測器在 1307 個回應裡命中 3 條,而**三條全是誤判**:譯文正確地把
咒語符文留成英文(「吾將其名為『REL XEN BET』」)。
⇒ 判英文之前先把**全大寫的串**挖掉,而不是放寬「連續三個詞」那個門檻 ——
放寬門檻會讓真的英文句子也漏掉。反對照(`TestTheDetectorActuallyDetects`)
兩邊都驗:真英文要抓到、刻意保留的英文不能誤判。

## 7. 引擎落地

| 原版 | 引擎 |
|---|---|
| `sub_2D72C` 的 tile 分派(崩塌 → default)| `Enter()` 先查 tile 再查座標 + `u5data.IsDungeonEntranceTile` |
| `byte_3E08B` / `byte_3E09B` / `byte_3E0F0` | `SaveActiveMemberOffset` / `SaveTurnCounterOffset` / `SaveRoomsClearedOffset` |
| `Save.validate` 的地點上限 | `u5data.LocationCodeMax`(含地牢)|
| `sub_92C0` | `npcOnUsableStair()` + `npcNeedsStairUp()` |
| `sub_89EC` | `nearestNPCStair()`(⚠ 曼哈頓距離,非 BFS 步數)|
| `sub_9690` 模式 6/7 | `stepNPCToStairs()`,接在 `stepNPC` 裡 |
| `sub_2D944` | `LightRadius2()` 開頭的黑暗格判斷 + `u5data.TileDarkness` |
| `sub_BE5C` / `sub_BD5C` / `sub_BD80` | ⬜ P6 過場播放器 —— 規格在 §5 |

### 測試

| 測試 | 驗什麼 |
|---|---|
| `TestCollapsedEntranceCannotBeEntered` | ★★★ 崩塌的入口按 E 印「What?」|
| `TestWordOfPowerOpensTheWayThrough` | ★★★ **無 debug 的完整玩家路徑**(含封回去的反對照)|
| `TestThreeFieldsSurviveASaveRoundTrip` | 三個位移的**位元組**都對(不只 round-trip)|
| `TestRoomsClearedSpanIsFourteenBytes` | ★★ 14 剛好塞滿 —— 差一個位元組就表示索引方式不對 |
| `TestStairConstantsAgreeAcrossFiles` | ★★ 兩個檔案的名字不能再分岔,而且 `ClimbDelta` 要同意 |
| `TestNPCStairMaskIsWiderThanStairsFacing` | ★ 0xF4 與 0xFC 兩個定義域(含 0xCC..0xCF)|
| `TestNPCWalksToStairsInsteadOfTeleporting` | ★★ 不在樓梯上不換層 |
| `TestNPCOnStairChangesFloor` | 反對照:站對了會換(不是整支關掉)|
| `TestDarknessTileBlindsOnTheOverworld` | ★★ 視野歸零 + 豁免模式 |
| `TestDarknessOnlyAppliesOnTheOverworld` | ★★ `TileBlank` 也是 0xFF —— 不能在場景裡誤觸 |
| `TestConversationsRenderInChinese` | ★★★ 135/135 段對話開場,0 行英文 |
| `TestKeywordAnswersRenderInChinese` | ★★★ 1307 個關鍵字回應,0 行英文 |
| `TestTheDetectorActuallyDetects` | ★ 偵測器兩邊都驗 |
