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
引擎的場景內戰鬥(`sub_C74`)還沒逆 —— 那條路目前誠實印一句說明。

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

- **場景裡與 NPC 的戰鬥**(`sub_C74`)—— 敵對的非衛兵貼上來時目前只印說明。
- `sub_154C`(對話號碼 0xFE 的特殊接觸)。
- `sub_1B52C` 回傳非零的那個條件(跟衛兵講話也會被抓的那條路)。
- 觸發逮捕的**行為來源**:偷竊、攻擊平民之類會讓衛兵敵對的事件,
  還沒對上哪一支。目前只有對話 opcode 0x8B 與 `u5dump` 的 `G` 進得去。
