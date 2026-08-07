# Yell 指令:收放帆、力量之言、召喚暗影君主

> 輸入檔:`org_game/fmtown/…/U5_E/WORRIORS.EXP`(FM Towns 英文版,Phar Lap `P3`)
> SHA-256:見 `docs/re/00-hexrays-p3-verified.md`
> 位址:`sub_17E74`(分派)、`sub_17CFC`(力量之言)、`sub_17C2C`(復原聖壇)、
> `sub_17A14`(召喚暗影君主)、`sub_27C98` → `sub_39554` → `sub_39C50`(字串比對)
> 日期:2026-08-07

一個鍵三種用途。原本只是想找「力量之言從哪裡輸入」,
結果 `sub_17E74` 一路把收放帆與暗影君主也帶了出來 —— 它們共用同一個指令。

## 1. 分派(`sub_17E74`)

```c
if ((byte_3E08C & 0xF8) == 0x20 && byte_3E0A3 < 0x80) {   // 在船上,而且不在戰鬥中
    if ((byte_3E08C & 0xFC) == 0x20) { puts("FURL!");  byte_3E08C += 4; }
    else                             { puts("HOIST!"); byte_3E08C -= 4; }
    return 1;
}
puts("what?\n:");
sub_239B4(buf, 12);                    // ★ 只讀 12 個字元
if (!buf[0]) { puts("Nothing\n"); return 1; }
if (byte_3E0A3 >= 1 && byte_3E0A3 <= 0x20) return sub_17A14(buf);   // 城裡:暗影君主
if (byte_3E0A3 == 0)                       return sub_17CFC(buf);   // 大地圖:力量之言
puts("\nNo effect!\n");
```

三件事值得先說清楚,因為都容易寫反:

**收放帆不問任何字。** 在船上按下去就直接切,不會跳出輸入框。
寫成「先問再判斷是不是在船上」會多出一個原版沒有的按鍵。

**`0x20..0x23` 是揚著帆的狀態。** 按下去印的是 `FURL!`(收帆)、載具碼 **+4**
變成 `0x24..0x27`。方向搞反的話,停在港口的船一按就開始跑。
`u5data.VehicleSailing = 0x20` / `VehicleShip = 0x24` 這兩個名字本來就是這樣定的。

**地牢不算「城裡」。** 判斷是 `1 ≤ 地點 ≤ 0x20`,而地牢是 `0x21..0x28`(見 `docs/re/18`)。
在地牢裡喊 → 落到最後那句 `No effect!`。

## 2. 字串比對是**前綴**,不是相等也不是子字串

這一段影響的不只 Yell —— 聖壇的真言、美德名、暗影君主的名字、力量之言
全都走同一支 `sub_27C98`。

```c
int sub_27C98(const char *needle, const char *haystack) {
    char up[10] = {0};
    int n = 0;
    for (int i = 0; i < 9; i++) {          // ★ 參考字截到 9 個字元
        if (!needle[i]) { n = i; break; }
        up[i] = islower(needle[i]) ? needle[i] - 0x20 : needle[i];
    }
    return sub_39554(up, haystack, n) == 0 ? 0 : -1;   // strncmp(up, haystack, n)
}
```

`sub_39554` 取 `min(strlen(a)+1, strlen(b)+1, n)` 個位元組交給 `sub_39C50`
(`repe cmpsb`)。所以語意是 **`haystack` 以 `needle` 開頭**:

| 參考字 | 玩家打 | 結果 |
|---|---|---|
| `Ahm` | `ahmxyz` | ✅ 過(**相等**語意會擋掉) |
| `Ahm` | `the ahm` | ❌ 不過(**子字串**語意會放過) |
| `hone` | `honesty` | ✅ |
| `hono` | `honesty` | ❌(誠實與榮譽只差第四個字母) |

⚠ 兩個截斷要記住:`Compassion`(10)與 `Spirituality`(12)都比 9 長,
原版實際上只比到 `COMPASSIO` / `SPIRITUAL`。

⚠ **大小寫還沒追完。** `sub_27C98` 只把**參考字**轉大寫,玩家打的那一行原樣交給
`repe cmpsb`。也就是說,如果輸入層(`sub_2B770` / `sub_28F40`)沒有幫忙轉大寫,
原版就得全大寫打才會過。輸入層存的是 `sub_29EEC` 的原始回傳值,沒有看到轉大寫的指令,
但 `sub_28F40` 這一層還沒讀。引擎取**不分大小寫**——這是超集,原版能過的一定過,
不會擋掉打對的玩家。追出來之後改 `u5data.MatchPrefix` 一支就好。

## 3. 力量之言(`sub_17CFC`)

八個字在 `off_55DF8`:

| # | 字 | 對應地牢 | 對應美德 |
|---|---|---|---|
| 0 | `FALLAX` | 欺瞞 Deceit | 誠實 |
| 1 | `VILIS` | 輕蔑 Despise | 慈悲 |
| 2 | `INOPIA` | 毀滅 Destard | 勇氣 |
| 3 | `MALUM` | 謬誤 Wrong | 正義 |
| 4 | `AVIDUS` | 貪婪 Covetous | 犧牲 |
| 5 | `INFAMA` | 羞恥 Shame | 榮譽 |
| 6 | `IGNAVUS` | 海斯洛斯 Hythloth | 靈性 |
| 7 | `VERAMOCOR` | 末日 Doom | 謙遜 |

★ 「地牢與美德共用同一個索引」不是推論。`sub_17CFC` 拿同一個 `edi` 去查
`off_55DF8`(字)、`byte_41114`/`byte_4113C`(地牢入口座標)、`byte_55E18`(入口地形);
它呼叫 `sub_17C2C(edi, …)` 時,那一支又用同一個 `edi` 去查 `off_411BC`(美德名)、
`off_411DC`(真言)、`byte_411FC`/`byte_41204`(聖壇座標)。一路到底都是同一個索引。

★ 而且 `byte_41114` / `byte_4113C` 與 `docs/re/18` 從 `sub_2D564` 讀出來的
**八座地牢入口座標一個不差** —— 兩份獨立來源對上。

### 3.1 掃鄰格的順序是固定的:西、南、東、北

```
byte_3F788   西   (var_4=-1, esi= 0)
byte_3F7A9   南   (var_4= 0, esi= 1)
byte_3F78A   東   (var_4= 1, esi= 0)
byte_3F769   北   (var_4= 0, esi=-1)
```

四個位址都是視窗緩衝裡玩家那一格的 −1 / +32 / +1 / −32。**取第一個命中的就停**。
兩個目標同時在旁邊時,這個順序決定動到哪一個。

命中條件是三選一:`原本的入口地形`、`0xDF`(已封印)、`0x1A`(被玷污的聖壇)。

### 3.2 封印是一個 XOR

```asm
xor  byte_3E0E0[edi], 80h        ; 旗標
call sub_DB10                     ; → 指向那一格地形的指標
mov  dl, byte_55E18[edi]
xor  dl, 0DFh
xor  [eax], dl                    ; 地形 ^= (原地形 ^ 0xDF)
```

`byte_55E18 = 18 16 16 18 18 17 17 16` —— 八個洞口開在三種不同的地形上。
一個 XOR 同時做開與關,所以 `0xDF` 與原本的地形互為彼此。
寫成兩個分支很容易某一邊漏掉,而症狀是「封起來就打不開了」。

### 3.3 ⚠ 座標不對時原版**靜靜結束**

```asm
cmp  edx, eax          ; X 對不上?
jnz  short loc_17E4B
...
cmp  eax, [ebp+var_C]  ; Y 對不上?
jnz  short loc_17E4B
...
loc_17E4B:
mov  [ebp+var_8], 1    ; ★ 仍然算「有反應」
```

`var_8 = 1` 表示不印 `No effect!`。所以:鄰格看起來像地牢入口、但不是**這個字**
對應的那一座時,原版**不繼續看別的方向,也不給任何回饋**。
這個沉默是原版的行為,不是漏寫 —— 引擎照抄(`SpeakWord` 在那裡 `return false`)。

沒有這個座標檢查的話,`FALLAX` 可以拿去開任何一座山洞。

## 4. 復原被玷污的聖壇(`sub_17C2C`)

流程與冥想(`sub_1D394`,見 `docs/re/25`)很像 —— 問美德名、問三次真言 ——
但有三處不同,都是實際會影響玩法的:

| | 冥想 `sub_1D394` | 復原 `sub_17C2C` |
|---|---|---|
| 美德名比對 | `off_55FEC` 的**四字母前綴**(`hone`) | `off_411BC` 的**完整名**(`Honesty`) |
| 輸入長度 | 12 / 8 | 15 |
| 失敗時 | 印「汝之思緒散亂」 | **什麼都不印**,只 `sub_27230(10)` 停一下 |
| 成功條件 | 真言全對 | 真言全對 **且座標是那一座聖壇** |

⚠ `off_411BC` 的第 2 個是 **`Valour`**(英式拼法),而聖壇座標表旁邊的
`off_55FEC` 前綴是 `valo`。原版兩處拼法就不同 —— 復原勇氣聖壇要打 `Valour`,
少一個 `u` 不算。**不要「順手統一」**。

成功時:`byte_3E0E8[i] &= 0x7F`(清掉玷污位元),地形寫回 `0x19`。

## 5. 召喚暗影君主(`sub_17A14`)

三個名字在 `off_55DEC`:`FAULINEI` / `ASTAROTH` / `NOSFENTOR`。

```c
if (byte_3E0A3 != 0x1E && byte_3E0A3 != 0x1F && byte_3E0A3 != 0x20) { puts("No effect!"); return 1; }
i = 查名字(0..2);
if (i == 3)                     goto no_effect;   // 不是那三個名字
if (byte_3E0A7 < 2)             goto no_effect;   // 玩家 Y < 2,上方擠不出位置
if (byte_3E0D8[i] == 0xFF)      goto no_effect;   // 這一位已經被消滅
for (esi = 0; esi < 32; esi++)
    if (物件[esi].tile == 0xFC) goto no_effect;   // 場上已經有一個了
…生在 (byte_3E0A6, byte_3E0A7 − 2, byte_3E0A5)…
byte_3EDB0[slot] = 0xFC;  byte_3E0DB = i;
puts("\nA shadowlord appears\n");
```

地點 **30 / 31 / 32 = 學術之城 / 共感修道院 / 巨蛇要塞** —— 正是三團聖火所在
(用 `u5data.Locations` 對過)。

⚠ **原版不檢查「名字要配這座城」。** 在三座裡的任何一座喊任何一個名字,
召來的都是那個名字的主人。別自作主張加上配對檢查。

⚠ 找 NPC 槽的方式:`ebx = 31`,`edi` 從 31 往下掃,**停在第一個空的**
(不是最小的空槽);全滿時直接蓋掉 31 號。照抄。

⚠ 尾端那段 `byte_3E08A` 存 `'T'` → 清 0 → 重畫 → 存回 `'T'` **不是**解除停時,
是「重畫時不要走停時那條路徑」的臨時開關。照抄成取消停時會多送玩家一個懲罰。

`0xFC` → 生物編號 `(0xFC − 0x40) / 4 = 47` = `CreatureShadowLord`,
與 `docs/re/16` 的戰鬥表對得上。

## 6. 存檔位移(`sub_27D24` 讀 / `sub_284CC` 寫)

跟著讀取序列從 `byte_3E0A7`(0x02F1)往後累加:

| 位移 | 大小 | 全域 | 內容 |
|---|---|---|---|
| 0x02F2 | 16 | `byte_3E0A8..B7` | 逐一單位元組 |
| 0x0302 | 32 | `byte_3E0B8` | 一整塊 |
| **0x0322** | 3 | `byte_3E0D8` | 三個暗影君主各自盤據的地點(0xFF = 已消滅) |
| **0x0325** | 1 | `byte_3E0DB` | 現在被召喚出來的是哪一個(0xFF = 沒有) |
| **0x0326** | 2 | `byte_3E0DC` | 進行中的聖壇試煉,一德一位元 |
| **0x0328** | 2 | `byte_3E0DE` | 已在寶典上讀到的美德(**寶典 `sub_1D850` 設的,不是聖壇**) |
| **0x032A** | 8 | `byte_3E0E0` | 八座地牢入口,bit 0x80 = 已封印 |
| **0x0332** | 8 | `byte_3E0E8` | 八座聖壇,bit 0x80 = 已玷污 |

★ 位移對不對有一個現成的證據:`INIT.GAM` 與 `SAVED.GAM` 在 0x0322..0x0339 這 24 B 裡
**只有 0x0325 是 0xFF,其餘全是 0** —— 而 0x0325 正是「沒有召喚中的暗影君主」。
位移若偏一格,那個 0xFF 就會落在別的欄位上。

⚠ `byte_3E0E8` **不是純旗標**:`sub_C318` 寫 0xFF,復原只清 bit 7 留下 0x7F。
引擎把這兩組存成八個位元組而不是一個位元遮罩,就是為了與存檔 1:1 ——
壓成遮罩會在存回去時把低位元洗掉。

## 7. 引擎的實作

- `u5data.WordsOfPower` / `Shadowlords` / `VirtueNames` —— 三張名字表,**維持英文**
- `u5data.MatchPrefix` —— 全遊戲共用的前綴比對(§2)
- `game.Yell` / `SubmitYell` —— 分派(`PromptYell`)
- `game.SpeakWord` —— 力量之言;`beginShrineRestore` / `shrineRestoreResolve` —— 復原聖壇
- `game.yellInTown` / `spawnShadowlord` —— 召喚
- 按鍵 **Y**;`u5dump` 腳本動作 `Y` 之後接 `"名字"`

## 8. 後續

✅ **遊走、玷污、消滅三塊已經補上**,連同它們背後的黑棘審問 ——
見 **`docs/re/28-shadowlords-and-blackthorn.md`**。那一份順帶解出:
玷污不是獨立的一支而是審問的結局之一、聖火的座標表藏在字串
`aNoNoticeableEf` 後面、招供的判定是**子字串**而不是本檔 §2 的前綴。

✅ **寶典**(`sub_1D850`)已補上 —— 見 `docs/re/27`。

還沒做的:

- 末日地牢入口的三個閘門(`sub_2D564` 要 `byte_3E0D8 & 3E0D9 & 3E0DA >= 0x80`,
  也就是三位全都不在城裡,否則 "Attacked at entrance!")。
- 逮捕的觸發(`sub_1884`)—— 見 `docs/re/28` §5。
