# 暗影君主的一生,與黑棘的審問

> 輸入檔:`org_game/fmtown/…/U5_E/WORRIORS.EXP`
> 位址:`sub_29304`(午夜遊走)、`sub_C414` → `sub_C318` → `sub_BFFC` / `sub_C098` / `sub_C13C`
> (審問)、`sub_1A38C`(寶石碎片)、`sub_1884`(逮捕)
> 日期:2026-08-07

`docs/re/26` 解完召喚之後留了三塊:遊走、玷污聖壇、消滅。這一份把三塊補齊 ——
中間那塊追下去才發現它根本不是獨立的一支,而是**黑棘審問**的結局之一。

整條線串起來是這樣:

```
午夜   sub_29304  三位活著的暗影君主各自換一座八德城市
被捕   sub_1884   在黑棘宮殿(地點 18)→ sub_C414
審問   sub_C318   逼問第一座還沒被玷污的聖壇的真言,問四次
         招了 → 那座聖壇被玷污(0x1A)、業報 −5、同伴被「仁慈地」處決
         不招 → 第四次鍘刀落下,同伴被斬;聖壇保住
復原   sub_17C2C  力量之言 + 完整美德名 + 三次真言(docs/re/26 §4)
召喚   sub_17A14  在三團聖火所在的城喊名字(docs/re/26 §5)
消滅   sub_1A38C  在對應的聖火那一格,把碎片舉起來
```

## 1. 午夜遊走(`sub_29304`)

小時從 23 進位成 0 的那一支裡:

```c
for (i = 0; i < 3; i++) {
    if (byte_3E0D8[i] >= 0x80) continue;      // 0xFF = 已消滅,不參與
    do {
        b = sub_28E14(1, 8);                  // 隨機一座八德城市
        if (byte_3E0A3 == b) b = 0;           // ★ 玩家就在那座 → 重抽
        for (j = 0; j < 3; j++)
            if (byte_3E0D8[j] == b) b = 0;    // ★ 已經有別位在那 → 重抽
    } while (b == 0);
    byte_3E0D8[i] = b;
}
```

⚠ 兩個排除條件都不能省。少了「玩家所在」那條,暗影君主會憑空出現在玩家腳下的城裡;
少了「不重複」那條,三位會疊在同一座城,剩下的城永遠遇不到。

⚠ 這個 `do…while` **理論上會空轉**(八座全被佔滿又剛好都排除時抽不到),
原版沒有防護。實務上三位加玩家最多佔四座,永遠有得選。引擎加了 1000 次上限,
只是不讓被改壞的存檔卡死主迴圈 —— 不是行為差異。

⚠ 引擎判「跨過午夜」用的是**日期變了沒**,不是「小時剛好等於 0」。
休息或進出聖壇石室一次推進超過一小時,看小時會整段跳過去。

盤據中的暗影君主怎麼影響那座城,`sub_48C` 只把索引記進 `byte_3E16A`,
用它的地方還沒追 —— 列在 §5。

## 2. 黑棘的審問(`sub_C414`)

進場條件是 `byte_3E0A3 == 0x12`(地點 18 = 黑棘宮殿,座標 (196,245),
與 `docs/re/03` 從進場地形表得到的「61 = 黑棘宮殿」對得上)。

```c
alive = 隊伍裡狀態 != 'D' 的人數;
byte_3E08C = 0x1C;                        // 強制下馬
puts("\nThou art subdued and blindfolded!");
for (v = 0; v < 8 && byte_3E0E8[v] != 0; v++) ;   // ★ 第一座還沒被玷污的聖壇
if (v == 8) goto 丟進地牢;                          // 八座全髒了 → 整幕不發生
…拖走、關進 11×11 石室、載入 MISCMAPS.DAT 與 MISCMSG.DAT…
puts("Blackthorn says:\n\n\"Ah, "); puts(聖者名); puts("!'Tis indeed an honour…");
sub_C318(v, alive);
丟進地牢: byte_3E0A5 = -1; byte_3E0A6 = 10; byte_3E0A7 = 7;
          byte_3DFB8 = 0;                 // ★ 鑰匙歸零
          byte_3E0A3 = 0x12;
```

⚠ **判的是 `byte_3E0E8[v] != 0`,整個位元組**,不只是 bit 7。
復原之後那一格留下 0x7F(仍然非 0),所以**復原過的聖壇不會再被黑棘挑中**。
寫成 `& 0x80` 的話,玩家辛苦復原一座就會被同一座反覆逼問。

⚠ **鑰匙歸零**。不然玩家一被關進去就能開門走人。

### 2.1 四輪問答(`sub_C318`)

```c
ebx = 0;
for (r = 0; r < 4; r++) {
    sub_BFFC(r, v);                       // 問句
    if (sub_C098(v)) {                    // 招了
        byte_3E0E8[v] = 0xFF;
        業報 -= 5;
        if (alive > 1) sub_C13C(0); else puts(#9);
        return;
    }
    if (alive < 2) { puts(#10); return; } // 只剩一人 → 下地牢
    if (!ebx) { ebx = 1; sub_C2D0(); }    // ★ 第一次拒絕只嗆一句
    else switch (r) {
        case 1: byte_3FA19 = 0xEB; break; // 刑具往前一格
        case 2: byte_3FA19 = 0xE8; break;
        case 3: sub_C13C(1); break;       // 鍘刀落下
    }
}
```

⚠ **招與不招都會死人**,只是訊息不同 —— 招了是「賞汝的同伴一個痛快」(#5),
不招到底是「汝的朋友替汝的背叛付了帳」(#6)。寫成「招了就沒事」會把這一幕
最重要的張力抹掉,而且玩家會發現招供沒有代價。

⚠ `ebx` 那個旗標:**第一次拒絕只被嗆,不動刑**。少了它,第一次拒絕就進入
處決倒數,四輪變三輪。

⚠ `switch` 的 `case 0` 是死碼(`ebx` 起始是 0,第 0 輪一定走「只嗆」那條)。

### 2.2 ★ 招供的判定是**子字串**,不是前綴

```c
int sub_C098(v) {
    sub_2B770(buf, 14);                        // 讀 14 個字元
    n = strlen(off_411DC[v]);                  // 真言長度
    for (i = 0; strlen(buf) - i >= n; i++) {   // ★ 一格一格往後滑
        ok = 1;
        for (j = 0; j < n; j++)
            if (upper(buf[i+j]) != upper(真言[j])) { ok = 0; break; }
        if (ok) return 1;
    }
    return 0;
}
```

所以「the mantra is Ahm」也算招供 —— 想含糊帶過是躲不掉的。

⚠ 這與聖壇用的 `sub_27C98`(**前綴**,見 `docs/re/26` §2)是**兩支不同的比對**。
抄成同一支的話,玩家只要在真言前面加一個字就能白嫖過關,而那正是這一幕唯一的抉擇。
引擎因此分成 `u5data.MatchPrefix` 與 `u5data.MantraSpoken` 兩支。

### 2.3 ★ 處決:被殺的人變成寶典前的骨灰罈(`sub_C13C`)

```c
找出**第二個**活著的人 esi;                  // 所以聖者本人永遠活著
把 roster[esi] 複製到暫存;
roster[esi..14] 整批往前遞補一格;
roster[15] = 暫存;
byte_3DFB3 = 0x7F;                            // = 0x3DDB4 + 15×32 + 31
byte_3E06B--;                                 // 隊伍人數 −1
if (blade) { puts(名字); puts(" is sliced in half! "); puts(#6); }
```

★ **`docs/re/27` §5 那個「還沒追」的 urn 解開了。** 0x3DFB3 就是名冊第 15 格的
位移 31,而 `sub_1DA10` 進寶典石室時掃的正是 `byte_3DDD3[i*32] == 0x7F`,
掃到就印「Thou dost see an urn marked: <名字>」並用 `sub_1D340` 擺出罈子。
也就是說:**位移 31 的 0x7F = 這個人被黑棘處決了,骨灰罈擺在寶典之前。**

⚠ 死法不是「狀態標成 'D'」。只標記的話,之後去寶典看不到罈子,而且名冊會留一個空洞。

### 2.4 ★ 十二筆問答就是 `MISCMSG.DAT` 的前十二筆

`sub_C414` 載入 `MISCMSG.DAT` 用的是 `sub_2C740(…, byte_54700, 0x3E8, 0)` ——
**位移 0**,而寶典那邊是 0x3AB(`docs/re/27` §3)。同一個緩衝區、兩個不同的起點,
只記一個會把另一邊的記錄全部算錯。

| 記錄 | 用在哪 |
|---|---|
| 0 / 1 / 2 | 前三輪的問句,**都是半句**,句尾接美德名再補 `?"` |
| 3 | 第四輪的問句,自成一句(不接美德名) |
| 4 | 「黑棘揮手一示,鍘刀落下!」 |
| 5 | 「謝了,吾友!……賜汝的同伴一個痛快。」(招供) |
| 6 | 「誰說吾不公道!……汝的朋友替汝的背叛付了帳!」(拒絕到底) |
| 7 | 「別犯下嘲笑吾的錯,蠢貨!」(第一次拒絕) |
| 8 | 「吾會一直問到沙漏見底。到那時,」**半句**,接名冊第 1 人的名字 + " die!\"" |
| 9 | 「吾在汝身上察覺到真實。這將以汝之性命為賞!」(招了而且孤身一人) |
| 10 | 「小孩都拆穿得了汝的謊……下地牢去!」(拒絕而且孤身一人) |
| 11 | 「且慢!」…(黑棘開場) |

⚠ 三筆半句(0/1/2 與 8)的譯文尾巴留了空白給接續用,而引擎的 `miscText`
會 `TrimSpace`。接的時候要自己補一個空格,否則會變成「什麼 ——Honesty?」。

## 3. 寶石碎片(`sub_1A38C`)

三個條件缺一不可,而且**兩種失敗的反應不同**:

```c
puts("Gem Shard\n\nThou dost hold above thee "); puts(["Falsehood…","Hatred…","Cowardice…"][i]);
…閃光…
if (byte_3E0A6 != X[i] || byte_3E0A7 != Y[i] ||
    byte_3E0A3 != 地點[i] || byte_3E0A5 != 樓層[i]) {
    puts("\n\nNo effect!\n"); return;           // ← 位置不對:有話說
}
puts("\n\n...and cast it into the Flame of "); puts(["Truth!","Love!","Courage!"][i]);
if (sub_2B360(x, y - 1, floor) != 0xFC) return; // ← 正北一格不是暗影君主:**沉默**
if (byte_3E0DB != i) return;                    // ← 現身的不是這一位:**沉默**
byte_3E0D8[i] = 0xFF;  byte_3DFC4[i] = 0;
dword_3E3DC |= byte_55E50[i];                   // 2 / 4 / 8
puts("\nThe doom of the Shadowlord "); puts(名字); puts(" is wrought!\n");
```

★ **四張座標表藏在字串後面。** `mov eax, offset aNoNoticeableEf` 之後用
`[eax+i+0x1C]` 定址 —— IDA 沒有替那塊資料命名,所以看起來像在讀字串。
`aNoNoticeableEf` 在 0x55E28,字串長 0x1C,之後 12 B 就是四段各 3 B:

```
+0x1C  X       0F 0F 0F
+0x1F  Y       09 03 10
+0x22  地點    1E 1F 20   = 30 / 31 / 32
+0x25  樓層    02 01 FF   = 2 / 1 / −1
```

30 / 31 / 32 正是學術之城 / 共感修道院 / 巨蛇要塞 —— 與 `sub_17A14` 的
召喚地點(`cmp al, 1Eh/1Fh/20h`)是**兩份獨立來源**,對上了。
而三個樓層都落在 `u5data.Locations` 記錄的樓層範圍內,是第三個佐證。

⚠ **`byte_3E0DB == i` 那一條不能漏。** 少了它,任何一塊碎片都能打掉任何一位,
而「哪塊碎片配哪團火」正是這條支線的全部內容。

碎片的持有旗標是 `byte_3DFC4[3]`,存檔位移 **0x0210**(4 B,第 4 B 未用)——
跟著讀取序列從 0x0208 累加,而後面 0x021A 是既有且已驗過的 `SaveItemsOffset`,
所以這一段沒有算錯的空間。

## 4. 引擎的實作

- `u5data.Flames` / `ShadowlordDoomBit` / `SaveShardsOffset` / `CharUrn`
- `u5data.MantraSpoken` —— 子字串比對(與 `MatchPrefix` 分開)
- `u5data.BlackthornQuestion` / `MsgBlackthorn*` / `BlackthornLocation`
- `game.roamShadowlords`,由 `AdvanceTime` 在**跨日**時呼叫
- `game.UseGemShard` —— 三個條件與兩種沉默
- `game.BeginInterrogation` / `AnswerBlackthorn`(`PromptBlackthorn`,**沒有 ESC**)
- `game.executeCompanion` —— 遞補名冊 + 打上 0x7F
- `u5dump` 腳本動作 `A` 開始審問,接著用 `"…"` 回答

## 5. 還沒做的

- ✅ **逮捕已經接上了** —— 見 `docs/re/29-npc-behaviour-and-arrest.md`。
  那一份順帶解出:排程表的第四個欄位(行為型別)引擎一直沒讀,
  所以 NPC 走到崗位就變蠟像;接上之後「叫衛兵 → 衛兵走過來 → 逮捕」
  整條鏈才通。**還缺的是「什麼事會讓衛兵敵對」**(偷竊、攻擊平民)。
- **盤據中的暗影君主對那座城做什麼**(`byte_3E16A` 的用處)。
- `dword_3E3DC` 的其餘位元(消滅三位之後湊齊要幹嘛)。
- 11×11 審問石室的畫面(`MISCMAPS.DAT` 位移 0)—— 與聖壇石室同一類簡化。
- 碎片本身怎麼取得(在三座地牢深處,`sub_1A5E8` 的道具使用流程)。
