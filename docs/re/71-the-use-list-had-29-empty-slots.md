# 71 — U 的清單有 38 格,引擎只接了 14 格

| | |
|---|---|
| 輸入檔 | `re_work/fmtowns/WORRIORS.EXP.asm`(FM Towns 英文版) |
| SHA-256 | 見 `re_work/fmtowns/SHA256SUMS.txt` |
| 主要函式 | `sub_1E8D4`(建清單)、`sub_1A5E8`(U 的分派)、`sub_19ED8`(卷軸,201 行)、`sub_1A0B0`(藥水,254 行)、`sub_1A2F8`(月石)、`sub_1D310`/`sub_1D31C`(兩個效果 setter)、`sub_1C9C0`(「On who:」) |
| 落地 | `internal/game/scroll.go` · `potion.go` · `moonstone.go` · `use.go` · `picker.go` · `internal/u5data/save.go` |
| 從哪裡來 | `docs/re/66` 的截斷清單:`sub_19ED8` **201 行 → 52 行 C、10 個字串掉了** |

---

## 1. 一個截斷函式,拉出三整族道具

`docs/re/66` 那份清單上 `sub_19ED8` 只是其中一行。追它的時候先問「誰呼叫它」,
答案是 `sub_1A5E8` —— U 指令。而讀 `sub_1A5E8` 的前 20 行就發現:

```
var_C = sub_1EFC8(...)              ; 玩家在清單上選了第幾格
if (var_C < 0) 離開
if (var_C <  8)              sub_19ED8(var_C)          ; ★ 卷軸
if (var_C < 10h)             sub_1A0B0(var_C − 8)      ; ★ 藥水
if (var_C > 14h && < 1Dh)    sub_1A2F8(var_C − 15h)    ; ★ 月石
switch (var_C − 10h)  …22 格跳表…                       ; 特殊道具
```

引擎只做了最後那張跳表。**前面三段一格都沒接**,而三族道具的 `Get` 都做了:

| | 撿得到 | 存得進檔 | 用得到 |
|---|---|---|---|
| 卷軸 ×8 | ✅ `get.go:127` | ✅ `SaveScrollsOffset` | ❌ |
| 藥水 ×8 | ✅ `get.go:113` | ✅ `SavePotionsOffset` | ❌ |
| 月石 ×8 | ✅ `get.go:140` | ✅(格式還錯,見 §5) | ❌ |

`Inventory.Scrolls` / `Potions` 這兩個欄位在整個 `internal/` 裡**只有寫、沒有讀**。
那是這一類洞的通用症狀,而且 `grep` 一次就看得到 —— 前提是想到要 grep。

## 2. ★ 一句錯註解把八格月石封了起來

`use.go` 原本寫著:

```
 5..12 "(0".."(7" → case 21..28  **不可用**(原版跳到 default)
⚠ … 它們在跳表裡走 default,根本不會被用,**不要試著「修好」它們**
```

跳表裡那八格**確實**指向 `def_1A6DD`,所以這句話看起來有一手證據。
但它們指向 default 的原因是**在進 switch 之前就已經被接走了** ——
21..28 在 `loc_1A6B3` 那六行被送去 `sub_1A2F8`(埋月石)。

⇒ 只讀跳表得到「這八格沒用」;真相在跳表**前面**六行。
`rulebook/63` 說程式碼是唯一真相 —— 這一條補充:**「讀了程式碼」不等於「讀對了範圍」**。
而那句「不要試著修好它們」還額外多了一層傷害:它主動勸退了下一輪的複查。

名字 `(0`..`(7` 也不是資料損毀,是八顆月石的短名(`(` 在原版字型裡是月相符號)。

## 3. 卷軸 `sub_19ED8` —— 三個「照抄咒語就會錯」的地方

```
byte_3E030[idx]--                  ; ★ 先扣一捲,在所有判斷之前
印 "Scroll"
0 VL  "Light!"          sub_1D310(0F0h)
1 RH  "Wind change!"    sub_1CC50() 問方向;byte_3E0A3 >= 21h → ebx = 0
2 IS  "Protection!"     sub_1D31C('P', 64h, 2)
3 IA  "Negate magic!"   sub_1D31C('N', 14h, 3)
4 IQW "View!"           > 7Fh → "Not here!";  < 21h → sub_EDD4(x,y) else sub_F7C0()
5 KXC "Summon Daemon!"  <= 7Fh → "Not here!";  else sub_1CE70(1)
6 IMC "Resurrection!"   >= 80h → "Not here!";  else sub_1C9C0() + sub_1CFC8()
7 AT  地點 1Dh / 28h → "No effect!";  else "Negate time!" + sub_1D31C('T', 14h, 7)
```

### (1) 卷軸不用魔力、不用藥草、不看等級

這是卷軸存在的意義 —— 一級的角色也能放出復活。所以它不能走 `Cast()`,
只能借 `spellEffect` 那一層。

### (2) ★ 持續時間與同名咒語**四個全不一樣**

`sub_1D310(al)` 就一行 `byte_3E0B6 = al`;`sub_1D31C(模式, 回合, 音效)` 也就三行:

```asm
sub_1D31C:  byte_3E08A = arg_0     ; 模式字母 'P'/'N'/'T'
            byte_3E09E = arg_4     ; ★ 回合數
            sub_1C8E8(arg_8)       ; 音效
            sub_2A1E8()            ; 重畫狀態列
```

拿卷軸押的常數對上咒語押的:

| | 卷軸 | 咒語 |
|---|---|---|
| 光明 | **240** | In Lor 100 / Vas Lor 255 |
| 防護 | **100** | In Sanct 20 |
| 抗魔 | **20** | In An 10 |
| 停時 | **20** | An Tym 10 |

⇒ **卷軸一律比咒語久。** 照 `spellEffect` 轉發會把四個時間全弄錯,
而且**不會有任何症狀**:效果照樣發生,只是長度不對。
`TestScrollDurationsDifferFromTheSpells` 把八個數字都釘住,
並且加了一條「兩邊被寫成一樣就紅」的自檢。

另外 `sub_1D310` 是**指派**不是取大值 —— 一捲光明卷軸會把 Vas Lor 的 255
**蓋成 240**。`TestLightScrollOverwritesVasLor`。

### (3) ★ 召喚惡魔與復活的場合條件是**相反**的

兩條都比 `byte_3E0A3` 與 0x7F/0x80,但一個 `jbe` 一個 `jnb`:

- **召喚惡魔只能在戰場上**(戰場外印 `Not here!`)
- **復活只能在戰場外**(戰場上印 `Not here!`)

憑印象寫一定會寫成同一個方向。`TestSummonDaemonAndResurrectionHaveOppositeGates`。

### 回傳值不是「成功了嗎」

`ebx` 一開始 1,只有「換風向而地點 ≥ 0x21」那條歸零;**三條 `Not here!` 都留著 1**。
它是 `sub_1A5E8` 的 `var_10` —— 這一回合算不算用掉了。

## 4. 藥水 `sub_1A0B0` —— ★ 每一瓶都有 1/8 機率不照顏色走

跳表**之前**兩行:

```asm
var_4 = sub_28E14(0, 0Fh)          ; random(0, 15)
if (var_4 == 0)      arg_0 = 4              ; ★ 1/16 → 不論哪一色都變成「睡著」
else if (var_4 == 1) arg_0 = sub_28E14(0,7) ; ★ 1/16 → 顏色整個重骰
```

合起來 **1/8 的瓶子會出錯**,其中一半直接把喝的人打昏。
這是原版「不明藥水」風險感的來源,而它**不在任何攻略的顏色對照表裡** ——
只有讀碼才看得到。`TestOneInEightPotionsMisfires` 用 400 瓶樣本釘住「它真的會發生」。

### 八色

| | 顏色 | 效果 | 場合 |
|---|---|---|---|
| 0 | 藍 | `Awaken` 狀態 `'S'→'G'`;戰場上若正是行動中的那個單位還要站起來 | 都可 |
| 1 | 黃 | `Healed!` 走 `sub_1CD3C` = Mani 同一支(回 1..30) | 都可 |
| 2 | 紅 | `Poison cured!` `'P'→'G'` | 都可 |
| 3 | 綠 | `POISONED!` `'G'→'P'` —— ★ **綠色是毒藥** | 都可 |
| 4 | 橙 | `Slept!` 要求狀態 `'G'`;戰場上走 `sub_2EDF8`(躺下) | 都可 |
| 5 | 紫 | `Poof!` 把 tile 換成 **0x90** | ★ 只有戰鬥中 |
| 6 | 黑 | `Invisible!` 旗標 `\|= 0x10`、tile 換成 0x1D | ★ 只有戰鬥中 |
| 7 | 白 | `sub_1CE0C` 掀開整張視線罩 | ★ 只有 `< 0x21` |

### ★ 紫色不是變形,只是換圖

`0x90 = CreatureBase + 0x14*4`,而 `0x14` 正是 Kal Xen 召喚的那一種
(`summonRat`)、也是 Rel Xen Bet 變形寫進去的同一個值 ——
**三處獨立命中同一個編號** ⇒ 紫藥水把人畫成老鼠。

而原版**只寫物件記錄的兩個 tile 位元組**(`[ebx]`、`[ebx+1]`),屬性一格都沒動。
所以它是純外觀,不是變形。`TestPurplePotionOnlyChangesTheSprite` 連
HP 與敏捷都一起驗,擋的就是「看起來很像變形所以順手做成變形」。

### 黑色的 0x1D 不是不明編號

`docs/re/53` 已經釘死站著 0x1D、躺著 0x1E,`combat.go` 的 `PartyTileStanding`
就是它。所以黑藥水順手把人「扶起來」,真正的效果是那個 `UnitHidden` 位元。

### 白色沒有回合數

`sub_1CE0C` = `sub_2E0E8(-1, …)`,半徑 −1 把整張視線罩填成可見
(`docs/re/17` §496、`docs/re/31`),接著**阻塞重畫 20 幀**,最後 `sub_29D64` 收尾。
⇒ 是**一瞬間看穿**,不是持續效果。

⚠ 引擎沒有阻塞動畫層,近似成「到下一個動作為止」(`State.RevealFlash`,
`tick()` 遞減)。**這是差異,不是原版有的計時器。**

### 「On who:」還是近似

`sub_1C9C0` 印 `On who: ` 讓玩家選一個隊員。引擎的隊員選單還沒做,
戰鬥外沿用 `spellTarget` 的同一套近似(傷得最重的那個);戰鬥中則是照原版取
`dword_3EF50[byte_3E0AE]` = 此刻行動的單位,**這一半是準的**。

## 5. 月石 `sub_1A2F8` —— 順手更正一條存檔格式

```
tile = sub_DB10(玩家x, 玩家y)
印 "Moonstone "
byte_3E0A3 >= 21h → "cannot be buried here!"
可埋:tile == 2Ch 或 2Dh,或 4 <= tile <= 10   (原版是 jle 3 / jge 0Bh 兩個開區間)
印 "buried!"
byte_3E040[i] = byte_3E0A6   ; X
byte_3E048[i] = byte_3E0A7   ; Y
byte_3E050[i] = byte_3E0A3   ; ★ 地點
byte_3E058[i] = byte_3E0A5   ; ★ 樓層
```

### 判準讀對了 —— 九個 `look#` 名字一次全對

```
  3 淺灘        ← 下界外,水裡埋不了
  4 沼澤   5 草地   6 灌木   7 焦灼荒漠
  8 灌木   9 樹林  10 熱帶森林                ← 界內,全是挖得動的地面
 11 山麓        ← 上界外,石頭地
0x2C 犁過的地  0x2D 豐收莊稼                  ← 兩個單獨列的例外恰好是農田
```

上下界恰好卡在「水」與「岩」上,兩個例外恰好是農地。
這不是巧合對得上,是同一個判準的九個獨立錨點(`rulebook/62`)。

### ⚠⚠ 存檔:是八顆 × 四欄,不是十六顆旗標

`save.go` 原本寫「`0x029A` 起 **16 B**,十六顆月石,0xFF = 在手上」,
並註明長度是被下一個已知欄位(`0x02AA` 藥草)夾出來的。**夾得沒錯,顆數與語意錯了。**

上面那四行連寫的四個陣列各間隔 8 B ⇒ 每個都是 8 格。位移換算的錨是同一組
(`byte_3E000 ↔ 0x024A`、`byte_3E060 ↔ 0x02AA`,兩端夾住中間 0x60 B):

| 記憶體 | 存檔 | 內容 |
|---|---|---|
| `byte_3E040[8]` | **0x028A** | 月石埋的 X ← **原本整段沒解碼** |
| `byte_3E048[8]` | **0x0292** | 月石埋的 Y ← **原本整段沒解碼** |
| `byte_3E050[8]` | 0x029A | 埋在哪個**地點**(0xFF = 還在手上) |
| `byte_3E058[8]` | 0x02A2 | 埋在哪一**層** |

第二個獨立佐證:`sub_1E8D4` 建清單時只掃 `ecx < 8`,而且拿
`byte_3E050[i] == 0FFh` 當「這顆拿得出來」—— 兩處咬合。

⇒ `0xFF` **不是**「在手上」的特殊旗標,是「地點欄還沒被寫過」。
原本那份「16 顆布林旗標」的讀法會把**埋在地點 0(大地圖)的月石讀成不存在**,
而把沒撿到的讀成……也不存在,所以症狀被掩蓋了。
`save_roundtrip_test.go` 現在有一條專門驗「地點 0 ≠ 在手上」。

## 6. 落地

| 檔案 | 內容 |
|---|---|
| `internal/game/scroll.go` | `ReadScroll` + 卷軸自己的四個時間常數 + 三個場合閘門 |
| `internal/game/potion.go` | `DrinkPotion` + 1/8 走偏 + 八色 + 三個場合閘門 |
| `internal/game/moonstone.go` | `BuryMoonstone` + `MoonstoneBuryable` |
| `internal/game/use.go` | 三段在跳表之前的分派;**改掉那句錯註解** |
| `internal/game/picker.go` | 清單順序照 `sub_1E8D4`:卷軸 → 藥水 → 特殊 → 月石 → 其餘 |
| `internal/game/state.go` | `locationCode()`(把 `byte_3E0A3` 的四種比較集中翻譯一次)+ `RevealFlash` |
| `internal/u5data/save.go` | 月石四欄位 + `Moonstone` 型別 + `MoonstoneInHand` |

`locationCode()` 值得單獨說:原版有一整族「這裡能不能做」的判斷,**全部**寫成
對同一個位元組的大小比較,而引擎把三種狀態分放在 `Location` / `Dungeon` / `Combat`
三個欄位。每次遇到就重新翻譯一次 → 翻錯一次就是一條行為差異,
而且**沒有測試看得出來**(「毫無效果」與「效果照樣發生」都不會報錯)。

## 7. 還沒讀的

- ⬜ **埋下去的月石怎麼變成月門**。`sub_E084` 讀的是另一組 `Moongates` 表,
  與 `byte_3E040/48/50/58` 的關係沒追。這是月石唯一的用途,不能算完成。
- ⬜ `sub_EDD4(x, y)` 與 `sub_F7C0()` —— View 卷軸在大地圖 / 場景的兩種俯瞰,
  引擎統一走既有的 `Peer()`,兩者差異未驗。
- ⬜ 停時卷軸在地點 **0x28** 為何無效(0x1D 是 STONEGATE)。0x28 同時是地牢
  Doom 的編號,但那條是在非戰鬥下比的 —— 目前最合理的讀法是「Doom 裡不能停時」,
  **沒有第二個證據,不寫成定論**。
- ⬜ `byte_3E09E`(`sub_1D31C` 寫的回合數)在引擎裡被拆成 `CombatModeTurns` 與
  `TimeStop` 兩個欄位。原版是**同一個位元組**,兩者不可能分歧;
  引擎的 `tickCombatMode()` 只遞減前者 → 有分歧的空間。**待併。**
- ⬜ 紫 / 黑藥水寫的是物件記錄的 `+0` 與 `+1` 兩個位元組(同一個值)。
  引擎的 `Combatant` 只有一個 `Tile`;`+0`(物件種類,≥ 0x40 = 生物)沒有對應欄位。
  「把隊員的物件種類寫成 0x90」會不會讓別的程式碼把他當怪物,未驗。
- ⬜ `sub_1A0B0` 的藍藥水在戰場上那條有個 `and edx, 0E8h; cmp edx, 88h` ——
  已照原話實作,但 `0xE8` 為何要一起排除 `UnitMonster`(隊員不可能是怪物)沒追。
