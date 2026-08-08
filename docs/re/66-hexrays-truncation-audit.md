# 66 — 把「Hex-Rays 安靜截斷」變成一份清單(順帶找回整個沉船系統)

| | |
|---|---|
| 輸入檔 | `re_work/fmtowns/WORRIORS.EXP.asm`、`re_work/fmtowns/WORRIORS_hexrays.c` |
| 工具 | `tools/hexrays_truncation.py`(可重跑) |
| 主要函式 | `sub_22F0`(船身損傷)、`sub_2CCFC`(轉向)、`sub_2D9D0`(海上遭遇)、`sub_2A4D0`(全隊傷害)、`sub_2B67C`(誰能行動) |
| 落地 | `internal/u5data/shipdamage.go`、`internal/game/shipdamage.go`;修正 `wind.go`、`dungeon.go`、`state.go` |
| 前一篇 | `docs/re/65`(許願井)—— 第一次踩到這個形態 |

---

## 1. 一次踩坑,兩個收穫

`docs/re/65` 記了一次代價不小的錯:整個許願井機制被判定為「原版沒有」,
因為讀的是 Hex-Rays 截斷後的 `sub_CD28`。當時補上的判斷法是:

> 反編譯出來的函式參數列是空的(而呼叫端明顯在推參數),就是被截斷了。

一個判斷法只用一次太浪費。把它寫成工具跑一遍全檔 —— **90 個函式命中**,
而其中一個把整個沉船系統藏了三十多年。

## 2. 兩個獨立信號

`tools/hexrays_truncation.py` 用兩個都不靠猜的信號:

| 信號 | 怎麼判 | 靈敏度 |
|---|---|---|
| **A 參數列對不上** | 組語的 `proc` 宣告了 `arg_N`(IDA 自己的堆疊框分析),而反編譯的定義是 `()` | 精確但漏得多 —— 無參數的函式也會被截斷 |
| **B 字串常數掉了** | 組語本體 `push offset aXxx` 的字串個數 > 反編譯本體的字串字面值個數 | 靈敏,而且**差幾個就等於漏了多少內容** |

信號 B 的差值直接當排序:從上面開始讀最划算。兩個都只是「該回去讀組語的清單」,
不是證明。

⚠ **0x32000 以上要濾掉**:那是 Phar Lap extender 與執行期程式庫,
`printf` 一族本來就是 varargs,參數列對不上是正常的。工具預設 `--max-addr 0x32000`。

```bash
tools/dev.sh python3 tools/hexrays_truncation.py \
    re_work/fmtowns/WORRIORS.EXP.asm re_work/fmtowns/WORRIORS_hexrays.c --top 20
```

## 3. 前二十名與逐筆分類

```
函式             參數 C/asm   字串 C/asm     行數 C/asm  前幾個掉的字串
 sub_A360       0/0        0/34        5/558     ', armed with ' | 'bare hands' | 'ARGH!' | 'Cast...'
 sub_142EC      0/0        1/28       10/433     'You find:' | 'No trap' | 'A simple trap' | 'A pit!'
 sub_21108      0/0        0/17        6/153     '"Our wine list,' | 'a) Rose.......18' | 'Thy choice?"'
★sub_CD28       0/3        3/14       14/139     'a well.' | 'Nothing' | 'Corvette' | 'Poof!'
 sub_10C34      0/0        4/14       30/328     'An apparition!' | 'Thou art now level ' | 'stronger!'
★sub_2CCFC      0/1        0/9         3/150     'Ride ' | 'Fly ' | 'Row ' | 'Head ' | 'Hull weak!'
 sub_D650       3/3        3/12       48/246     'Wanted:   ' | 'abbbbbbbbbbbbbc'
 sub_1D394      0/0        0/9        16/400     'Mantra:' | 'ALAKAZAM' | 'Strength +1'
 sub_14CAC      0/0        1/10        3/264     'No Keys!' | 'Key broke!' | 'Unlocked!'
 sub_74BC       0/0        0/8         3/81      'Mage' | 'Bard' | 'Fighter' | 'Druid'
 sub_23D50      0/0        0/8        35/216     'EGA0.TIL' | 'EGA1.TIL' | 'EGA2.TIL'
 sub_2B8CC      0/0        8/15       52/287     'Sails must be' | 'lowered!' | 'Hull now '
 sub_1A5E8      2/0       30/37      169/511     'Item: ' | 'Boarded!' | 'X-it ship first!'
 sub_2D9D0      0/0        5/11      130/338     'Rough seas!' | 'Exit to MENU? '
 sub_1B3D0      0/0        2/7        10/127     'A guard demands' | 'gp tribute' | 'IMPE'
```
(★ = 參數列也對不上;共 90 個。)

逐筆對過既有筆記與程式碼之後:

| 狀態 | 函式 |
|---|---|
| ✅ 早就從組語逆過,截斷沒造成損失 | `sub_21108` 酒單(`tavern.go`)、`sub_1D394` 聖壇與 `ALAKAZAM`(`shrine.go`)、`sub_14CAC`/`sub_14B2C` 撬鎖(`docs/re/41`)、`sub_10C34` 升級(`docs/re/19`)、`sub_2B8CC` 紮營修船(`docs/re/50`)、`sub_1B3D0` 衛兵盤查(`docs/re/32`)、`sub_142EC` 搜索與陷阱(`docs/re/43`/`18`)、`sub_1A5E8` 指令迴圈 |
| ★ **這次找回來的** | `sub_2CCFC` 轉向 + `Hull weak!`、`sub_2D9D0` 的 `Rough seas!`、以及它們共同呼叫的 `sub_22F0` 沉船 |
| ⬜ 還沒讀 | `sub_A360`(**558 行、34 個字串**,最大的一筆)、`sub_D650`(`Wanted:` 通緝告示?)、`sub_74BC`、`sub_23D50`(`EGA*.TIL` 載入)、其餘 75 個 |

⇒ 好消息:**大多數重要子系統當初就是讀組語逆的**,所以截斷沒有造成損失。
壞消息:船是例外,而且漏得徹底。

---

## 4. 找回來的東西:船會沉

### 4.1 `sub_22F0` — 一次船身損傷判定

```
if ((載具 & 0xF8) != 0x20) return          ; ★ 只有大船會受損(0x20..0x27)
傷害 = rand(1, 30)
if (傷害 < 耐久) { 耐久 -= 傷害; return }
印 "Ship sunk!"
if (船上小艇 > 0)      { 印 "Abandon ship!"; 載具 = 0x28 | (載具 & 3) }
else if (魔毯 > 0)     { 印 "Abandon ship!"; 魔毯--; 載具 = 0x14 + rand(0,1) }
else {
    載具 = 0
    印 "Drowning!"
    if (sub_2B67C() != -1) 每個沒死的隊員受 rand(1, 8) 傷
}
```

三層階梯:**小艇 → 魔毯 → 溺水**。兩個細節照原樣保留:

- 換成小艇那條**不扣小艇數**(組語裡沒有 `dec`)。
- 溺水把載具碼設成 **0** —— 那不是「步行」(0x1C),是「不畫載具圖」。
  同樣手法在 `sub_10A1C` 的墜落動畫裡出現過(存舊值 → 設 0 → 重畫 → 還原),
  而這裡**沒有還原**,所以隊伍的圖示就這樣消失在水裡。

### 4.2 六個觸發點

| 呼叫者 | 場合 |
|---|---|
| `sub_2D9D0` | ★ **`Rough seas!`** —— 小艇或魔毯站在水 tile 1 上 |
| `sub_2CE70` | 撞擊 / 觸礁(音效 `sub_2C598(64h, 7D0h, 12Ch)`) |
| `sub_23FC` ×1、`sub_24DC` ×2、`sub_25F0` ×1 | 戰鬥與地形效果 |

⚠ `Rough seas!` 之後那次 `sub_22F0` 對小艇與魔毯**是空轉**(它只動大船)——
所以對小船來說只有訊息與動畫。這個「無效呼叫」看起來像可以省掉的程式碼,
但省掉就等於自己決定小船該不該受損。照原樣保留,並用測試釘住。

### 4.3 `sub_2B67C` 的語意順帶釘清楚

```
掃隊員:'G'(良好)或 'P'(中毒) → 記下編號,**回 0**(能行動)
        'S'(睡著)               → 計數,繼續掃
        其他('D' 死 / 'C' 魅惑) → 跳出迴圈
掃完沒人能行動:有人睡著 → 回 1;否則 → 回 −1
```

這讓兩個地方同時讀得通:`sub_1A54` 的 `if (v2 == 1) 印 "Zzzzzz..."`(全隊睡著,
跳過回合),以及溺水的閘門 —— **回 −1 就跳過傷害**,因為那時隊伍已經全滅。

★ 值得記一筆:**中毒('P')與良好('G')走同一條分支**。寫成「只有 G 能行動」
的話,中毒的隊伍會被當成全滅。

---

## 5. 順帶修掉的三個錯

### 5.1 `damageWholeParty` 的傷害值是猜的

`dungeon.go` 原本寫:

> ⚠ 傷害值還沒逆出來 —— `sub_2A4D0` 內部另有一套。這裡用陷阱最常見的
> `random(1, 20)`…並在文件裡標明是**估計值**。

`sub_2A4D0` 整支只有九行:

```
for (i = 0; i < 6 && i < 隊伍人數; i++)
    if (狀態[i] != 'D') sub_2A464(i, sub_28E14(1, 8))
```

**rand(1, 8)**,不是 1..20 —— 差了兩倍半,而地牢陷阱有四處用它。
沒讀到的原因與許願井同一個。`sub_2A464` 就是既有的 `damageMember`,
所以那段自己重寫的扣血與判死也一併刪掉了。

### 5.2 「無風照走」只對一半

`wind.go` 的 `CanSail` 原本寫「無風:原版在查表前就 `jz` 掉,船一律不受風影響 —— 照走」。
依據(`sub_2D38` 無風時不查延遲表)是對的,**結論不對**:

```asm
; sub_2CCFC,已經朝著要去的方向
        cmp     byte_3E08C, 24h
        jnb     short 照走            ; 收帆的船(0x24..0x27)不受風影響
        cmp     byte_3E0A2, 0
        jnz     short 照走            ; 揚帆 + 有風 → 走
        mov     ebx, 1               ; ★ 揚帆 + 無風 → 這一步用掉了,動不了
```

⇒ **揚著帆而無風就是動不了**;不查延遲表與「照走」是兩件事。
收帆的船(0x24..0x27)才不受風影響 —— 這也解釋了為什麼原版要分成兩組載具碼。

### 5.3 船轉向要花掉一回合

`sub_2CCFC` 與移動時印動詞的 `sub_7C0` 幾乎一樣,**差別是大船那一段**:

```
新朝向 = (載具 & 0xFC) | 方向
if (新朝向 != 載具) {
    載具 = 新朝向
    印 "Head " + 方向名
    if (耐久 < 0x32) 印 "Hull weak!"     ; ★ 門檻 50
    return 1                             ; ★ 這一步被轉向吃掉
}
```

回傳 1 = 「用掉了」,呼叫端 `sub_2D174` 收到非 0 就不移動。
引擎原本在同一步裡轉向 + 前進,所以船會像馬一樣靈活 —— 那不是 U5 的航海手感。

★ **`Hull weak!` 的門檻是 50,與上船時的 10 是兩個不同的警告**:
上船唸一次 `Danger! Ship badly damaged!`(`sub_16F08`,`cmp eax, 0Ah`),
航行中每次轉向唸 `Hull weak!`。合併成一個就會少掉「船況正在惡化」的提示。

---

## 6. 落地與驗收

| | |
|---|---|
| `u5data/shipdamage.go` | 三個耐久常數(新船 100 / 警告 50 / 骰上限 30)、溺水傷害 8、`ShipTakesDamage` / `RoughSeasAffects` |
| `game/shipdamage.go` | `DamageShip` / `abandonShip` / `drown` / `anyoneCanAct` / `RoughSeas` / `turnShipInstead` |
| `game/state.go` | `moveInWorld` 在風向判定**之前**先問 `turnShipInstead` |
| `game/wind.go` | 把「無風照走」的註解改成正確的說明 |
| `game/dungeon.go` | `damageWholeParty` 改成 `rand(1, 8)` + 六人上限 + 用 `damageMember` |
| `tools/hexrays_truncation.py` | 本篇的清單,可重跑 |

十條測試,其中幾條是專門擋「順手改」的:

- `TestOnlyBigShipsTakeHullDamage` —— 擋「把 Rough seas 的傷害套到小艇上」
- `TestRoughSeasHitsSmallCraftOnlyAndDealsNoDamage` —— 擋「把那個空轉呼叫優化掉」
- `TestHullWeakThresholdIsFiftyNotTen` —— 擋「兩個警告合併成一個」
- `TestAnyoneCanActMatchesTheThreeWayReturn` —— 逐狀態釘住,含「'P' 也能行動」
- `TestAbandonShipLadderIsSkiffThenCarpetThenDrown` —— 含「有小艇時不扣小艇數」
- `TestSailsUpWithNoWindCannotMove` —— 三種組合都驗(揚帆無風 / 揚帆有風 / 收帆無風)

## 7. 還沒做的

- **`sub_A360`(558 行、34 個字串)一個字都還沒讀。** 從掉的字串看,
  它與指令回顯、`ARGH!`、`, armed with `(裝備敘述)有關,可能是主指令分派的一大塊。
  這是清單上最大的一筆,也是下一個該讀的。
- `sub_2D9D0` 只讀了 `Rough seas!` 那一段;同一支裡還有 `Exit to MENU?` 與
  一堆座標判斷沒追。
- `sub_2CE70`(觸礁)只確認它會呼叫 `sub_22F0`,觸發條件沒讀。
- `RoughSeasTile = 1` 只記原版比的那個值,**沒有解釋成「深水 / 中水」**——
  tile 0..3 都是水,哪一個是哪一種深度缺第二份證據。
