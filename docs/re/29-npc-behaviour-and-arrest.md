# NPC 走到崗位之後在做什麼,以及被逮捕

> 輸入檔:`org_game/fmtown/…/U5_E/WORRIORS.EXP`
> 位址:`sub_95BC`(行為型別跳表)、`sub_94E0`(遊走)、`sub_8F60`(追擊 / 逃)、
> `sub_8F3C`(距離)、`sub_195C`(接觸)、`sub_C10`(叫衛兵)、`sub_B44` / `sub_B98`
> (敵對 / 逃跑)、`sub_1884`(逮捕)
> 日期:2026-08-07

`docs/re/28` §5 留了一句「逮捕還沒有觸發點」。追下去發現觸發點在一整層
**引擎從來沒讀過的資料**裡:排程表每個 slot 的第四個欄位。

## 1. ★ 排程表的「行為型別」欄一直是死的

`.NPC` 的排程記錄每個 slot 有四個欄位:**行為型別**、X、Y、樓層。
引擎只用了後三個 —— NPC 走到崗位、模式變成閒置,然後就完全靜止。

原版不是這樣:`sub_9690` 在 NPC **與玩家同層**時呼叫 `sub_95BC(npc, slot)`,
由行為型別接手。八格跳表:

| 型別 | 行為 | 分支 |
|---|---|---|
| 0 | 固定不動 | default |
| 1 | 在崗位 3 格內隨機走 | `sub_94E0(…, 3)` |
| 2 | 同上但範圍 0(原地) | `sub_94E0(…, 0)` |
| 3 | 玩家進 4 格內就**往外躲** | `sub_8F60` |
| 4 | 玩家進 4 格內就**靠過來**,否則遊走 3 | `sub_8F60` / `sub_94E0` |
| 5 | 永遠靠過來 | `sub_8F60` |
| 6 | 同 3 的觸發距離 | `sub_8F60` |
| 7 | 同 5 | `sub_8F60` |

⚠ 3 與 6 走同一個分支、5 與 7 也是 —— **差別在 `sub_8F60` 內部再讀一次自己的型別**:
只有 3 是遠離,其餘都是靠近。抄成「3/6 一樣」會讓怕生的村民追著玩家跑。

⚠ 型別 4 判距離用的是**崗位座標**(`[eax+esi+3]`/`[eax+esi+6]`),
不是 NPC 現在站的位置。

距離一律是 `sub_8F3C` = **曼哈頓距離** `|dx| + |dy|`。

### 1.1 遊走(`sub_94E0`)

```c
if ((random(0, 0xFF) & 8) == 0) return;        // ★ 測 bit 3,不是「< 128」
dir = (random(0, 0x3F) & 3) + 1;               // ★ 取低兩位元,不是 random(1,4)
nx, ny = step(rt.X, rt.Y, dir);
if (maxDist && manhattan(homeX, homeY, nx, ny) > maxDist) return;
if (!sub_9428(nx, ny, npc, slot)) return;
rt.X, rt.Y = nx, ny;
```

⚠ 型別 2(範圍 0)**不是「直接 return」** —— 它照樣擲兩次骰,只是任何一步都
超出範圍。省掉那兩次擲骰,之後的亂數序列就跟原版岔開了。

### 1.2 追擊與逃(`sub_8F60`)

```c
ai = byte_3E570[npc*16 + slot];
if (manhattan(playerX, playerY, rt.X, rt.Y) == 1 && ai > 3) {
    byte_3EDD0 = ((ai == 4 || ai == 5) && 對話號碼 != 0) ? 't' : 'a';
    byte_3EDD1 = npc;
    return;                                     // ★ 不移動,交給 sub_195C
}
for (dir = 1; dir <= 4; dir++)
    dist[dir] = 走得過去 ? manhattan(玩家, 那一格) : 0x63;
best = -1;
if (ai == 3) { 找 dist > 現在的;有多個時**擲硬幣**決定要不要換 }
else         { 找 dist < 現在的,取第一個;都沒有就退而求其次接受 == 現在的 }
if (ai == 5 || ai == 7)
    if (random(0, 0x3F) < 0x10) 從其餘可走的方向裡重挑一個;   // ★ 醉步
if (best > 0) 移動;
```

⚠ **貼到玩家旁邊(距離 1)時不再移動,而是「接觸」。** 這是整條逮捕鏈的關節。

⚠ 型別 5 / 7 有 25% 機率不照最佳解走。少了它,跟隨型 NPC 走得像導彈。

### 1.3 ⚠ 「可站」有兩支,不能混用

- `sub_9358`(引擎的 `npcCanStand`):**目標格永遠回可站** ——
  NPC 要走得到自己的崗位,哪怕崗位畫在櫃檯後面。
- `sub_9428`(這一節用的):沒有那個豁免。

實作時把候選格當成自己的目標傳進去,檢查就恆真了 ——
NPC 會走進牆裡、走出地圖外、疊在別人身上。
這個錯誤當場被 `TestNPCsStayOnWalkableTiles` 抓出來,
`TestNPCsDoNotOverlap` 則直接 panic 在負座標上。引擎因此另開一支 `canStandPlain`。

## 2. 接觸之後(`sub_195C`)

```c
if (byte_3EDD0 == 'a') {
    if (對話號碼 == 0xFE)      sub_154C(npc);      // 特殊
    else if (tile == 0x70)     sub_1884();         // ★ 衛兵 → 逮捕
    else                       開打;
} else if (arg != 0 || (對話號碼 != 0 && sub_1B52C(npc))) {
    sub_1884();
}
```

所以「衛兵貼到你身上」就是逮捕,其他敵對生物就是開打。
場景內戰鬥(`sub_C74`)見 §7。

## 3. 叫衛兵(`sub_C10`)

對話 opcode 0x8B 呼叫的就是它:

```c
for (i = 0; i < 32; i++) {
    if (槽是空的) continue;
    c = byte_3EDB0[i];
    if (c == 0xFC || c == 0xD8 || c == 0x70) sub_B44(i);          // 變敵對
    else if (random(0, 0xFF) < 0x80)          sub_B98(i);          // ★ 一半逃跑
}
```

⚠ **後半段不能省。** 少了它,叫完衛兵整條街的人還若無其事地站在原地。

`sub_B44`(變敵對):型別設 **6**(生物編號 < 0x2F)或 **7**(其餘),
而且**四個排程時刻一起清成 0** —— 不清的話,下一個整點換班會把牠打回崗位,
敵意就這樣消失了。

`sub_B98`(逃跑):型別設 **3**。

三種會應召的生物:`0x70` 衛兵、`0xD8` 另一種守衛、`0xFC` 暗影君主。

## 4. 逮捕(`sub_1884`)

```c
if (byte_3E0A3 == 0x12) {                     // 黑棘宮殿
    if (沒有還醒著的隊員) return 0;
    sub_C414();                                // → 審問(docs/re/28 §2)
    goto after;
}
puts("\n\"Thou art under arrest!\"\n\n\"Wilt thou come quietly?\"\n\n:");
等一個 Y 或 N;                                 // ★ 沒有 ESC
if (Y) {
    puts("Yes\n\nThe guard strikes thee unconscious!");
    puts("\nThou dost awaken to...\n");
    byte_3E0A3 = 4;                            // ★ 紫杉城 YEW
    byte_3E0A6 = 0x19;  byte_3E0A7 = 4;        // (25, 4)
    while (byte_3E08F != 8) sub_29304(20);     // ★ 一次跳 20 分鐘直到早上八點
    byte_3DFB8 = 0;                            // ★ 鑰匙歸零
    byte_3E0A5 = 0;
} else {
    puts("No\n\n\"Then defend thyself, rogue!\"\n");
    sub_C10();                                 // 全城衛兵撲上來
    return 1;
}
```

⚠ **一次跳 20 分鐘直到小時剛好是 8** —— 所以醒來的分鐘數取決於被抓的時間,
不一定是整點。直接把時鐘設成 08:00 看起來一樣,但月相與 NPC 排程會差。

⚠ **鑰匙歸零**。不然玩家一被關進去就能開門走人。

## 5. 引擎的實作

- `u5data.NPCAI*` —— 八種型別、觸發距離 4、遊走範圍 3、0x63 這個「走不過去」值
- `u5data.CreatureGuard` / `CreatureGuardCaptain` / `CallGuardsFleeChance`
- `u5data.ArrestJail*` / `ArrestWakeHour` / `ArrestWakeStep`
- `game.npcAIStep` 掛在 `stepNPC` 的 `ModeIdle` 分支上
- `game.npcWander` / `npcApproach` / `npcContact` / `canStandPlain`
- `game.CallGuards` / `makeHostile` / `makeFlee`,對話 opcode 0x8B 已接上
- `game.Arrest` / `AnswerArrest`(`PromptArrest`,只收 Y / N)
- `game.talkToNPC` —— 從 `Talk()` 拆出來,因為**對話不只玩家能發起**
- `u5dump` 腳本動作 `G` 叫衛兵

## 6. 還沒做的

- `sub_154C`(對話號碼 0xFE 的特殊接觸)。
- `sub_1B52C` 回傳非零的那個條件(跟衛兵講話也會被抓的那條路)。
- 觸發逮捕的**行為來源**:偷竊、攻擊平民之類會讓衛兵敵對的事件,
  還沒對上哪一支。目前只有對話 opcode 0x8B 與 `u5dump` 的 `G` 進得去。

---

## 7. 補記:在城裡跟 NPC 打起來(`sub_C74`),以及「誰再也不會出現」

§6 把場景內戰鬥列為未做。追下去發現它很短,而且牽出一個**存檔欄位的漏洞**。

```c
void sub_C74(int npc) {
    sub_218(npc);                                  // ★ 記帳(只對某些生物)
    sub_2E58C(word_3E77C[npc*16]);                 // 開打 —— 與撞上野外怪物同一支
    sub_268(npc);                                  // 從場上抹掉
    sub_5C8(0);  sub_48C();                        // 重載地圖
}
```

⚠ **順序不能換。** 先開打再記帳的話,`sub_218` 要用的生物編號已經被
`sub_268` 清成 0 了。

### 7.1 ★ `sub_218`:哪些人死了會被**永久**記下

```asm
movzx esi, byte_3EDB0[edi]
and   esi, 0FCh
cmp   esi, 70h  ; jz  loc_243      ← 衛兵 → 跳去檢查 0xB4
cmp   esi, 80h  ; jl  loc_24B      ← < 0x80 → 記
loc_243: cmp esi, 0B4h ; jnz 返回
loc_24B: or dword_3E368[地點*4], 1 << npc
```

分支很容易讀反。攤開來是:

| 生物 & 0xFC | 記不記 |
|---|---|
| `0x70` 衛兵 | **不記**(跳去比 0xB4,不相等就返回) |
| `< 0x80` 其餘 | 記 |
| `0xB4` | 記 |
| 其餘 `>= 0x80` | 不記 |

讀成「衛兵要記」的話,玩家把守衛殺光之後那座城就永遠沒有衛兵了。

### 7.2 ★ `dword_3E368` 一直沒被存檔

`dword_3E368[地點]` 是 32 個 u32,位元 i = 那個地點的第 i 個 NPC 已被永久清掉。
它**在存檔裡**(`sub_27D24` 讀 `dword_3E36C`,0x80 = 128 B),而引擎原本只有一個
記憶體裡的 `removed` map —— 打死居民、或把人招進隊伍,存檔再讀回來人就復活了
(招募的那個還會變成分身:名冊裡一個、城裡一個)。

位移 **0x05B4**,跟著讀取序列從 0x0332 累加:
`byte_3E0E8`(8)→`byte_3E0F0`(14)→`byte_3E100`/`3E120`/`3E140`(各 32)
→`byte_3E160..3E16B`(12 個單位元組)→`dword_3E16C`(512)→ **這一段**。

⚠ 陣列基底是 `dword_3E368` 而索引是**地點編號 1..32**,所以存檔裡的第 0 個 u32
對應**地點 1**。大地圖(地點 0)那一格不在存檔範圍內 —— 它沒有場景 NPC。

⚠ **兩種「不在場」要分開**:存檔裡的那一份是永久的;這一次進場景中途被抹掉的
(打死的衛兵、用碎片消滅的暗影君主)在原版是放在**每次進場景重新從 `.NPC` 載入**
的暫存表裡,離場就沒了。引擎因此在 `loadNPCs` 先清掉這個地點的 session 記錄,
再套存檔的位元 —— 不清的話衛兵死一次就再也不會補上
(`TestKillingAGuardIsNotPermanent` 抓到的就是這個)。

### 7.3 ⚠ 一個疑似原版 bug:末日位元與地點 29 共用儲存

`sub_1A38C` 消滅暗影君主時 `or dword_3E3DC, 位元`,而
`0x3E3DC = 0x3E368 + 29×4` —— 那正是**地點 29(石門 Stonegate)的
永久移除遮罩**。也就是說,消滅三位暗影君主會把石門的第 1 / 2 / 3 號 NPC
標成「已清掉」。

沒有其他證據顯示這是刻意的。引擎照抄同一塊儲存(`DoomFlags` 就是
`RemovedNPC[28]` 的低位元)—— 「機制與原版一模一樣」包含把原版的 bug 也抄進來,
除非確認它會讓遊戲玩不下去。

## 8. 引擎的實作(補)

- `u5data.NPCKillIsPermanent` / `SaveRemovedNPCOffset` / `RemovedNPCLocations`
- `game.beginNPCCombat` —— 合成一個等價物件餵給既有的戰鬥流程
- `game.markNPCRemoved` / `applyRemovedNPCs`
- `BeginCombat` 拆成 `beginCombatFrom`,物件與 NPC 兩條路共用
