# 94 — Ztats 有 17 頁,而繞回的接縫不在隊伍裡

| | |
|---|---|
| 輸入檔 | `re_work/fmtowns/WORRIORS.EXP`(一手位元組)、`WORRIORS.EXP.asm` / `.i64` |
| SHA-256 | 見 `re_work/fmtowns/SHA256SUMS.txt` |
| 主要函式 | ★★ `sub_1E9A0`(頁面游標與按鍵)、★ `sub_1DECC`(隊員數值頁)、`sub_1E128`(隊員裝備頁)、★ `sub_1E210`(Equipment)、`sub_1E588`(清單頁通用)、★★ `sub_1E8D4`(Items 頁的搬運) |
| 主要資料 | ★ `off_408F4`(38 個縮寫名)、`byte_40BA0`(執行期組出的 38 筆計數)、`off_40B50`(職業名)、`off_40B8C`(狀態名) |
| 起因 | 清 §5.2 截斷清單的 `sub_1E210` |
| 狀態 | ✅ 全部 17 頁定案並落地;順手釘死黑徽章的存檔位移 |

---

## 0. ★★ 一個游標,17 頁

原版**不是**「哪個人 + 看哪一頁」兩個變數,而是 `sub_1E9A0` 裡**一個** `esi`(0..16):

| 游標 | 頁 | 函式 |
|---|---|---|
| 0..11 | 六名隊員 × 兩頁(偶數 = 數值、奇數 = 裝備)| `sub_1DECC` / `sub_1E128` |
| 12 | **Equipment**(全隊共用的糧食 / 金幣 / 鑰匙 / 寶石 / 火把 / 抓鉤)| `sub_1E210` |
| 13 | Reagents(8 筆)| `sub_1E588(…, 8, byte_3E060)` |
| 14 | Spells(48 筆)| `sub_1E588(…, 0x30, byte_3E000)` |
| 15 | Items(38 筆)| `sub_1E588(…, 0x26, byte_40BA0)` |
| 16 | Armaments(48 筆)| `sub_1E588(…, 0x30, byte_3DFD0)` |

「一個游標」不是實作細節 —— **兩個接縫用兩個變數寫不出來**:

```
下一頁(鍵 2 / 4):
   if (esi == 人數×2 − 1)  esi = 0Ch        ; ★ 最後一名之後跳 Equipment
   else if (esi >= 10h)     esi = 0          ; ★ Armaments 之後繞回第一名
   else                     esi++

上一頁(鍵 1 / 3):
   if (esi == 0Ch)  esi = 人數×2 − 1         ; ★ Equipment 往前回到最後一名
   else if (esi > 0) esi−−
   else              esi = 10h               ; ★ 第 0 頁往前繞到 Armaments
```

⇒ 四人隊伍在第 7 頁就跳走,**沒有人的第 8..11 頁翻不到**。

### 按鍵

| 鍵 | 作用 |
|---|---|
| `2` / `4` | 下一頁 |
| `1` / `3` | 上一頁 |
| `0` | 跳 Equipment |
| `1`..`6` | 跳第 N 名的數值頁(`esi = 鍵碼×2 − 0x62`)|
| 空白 / ESC | 收起 |

⚠ `1` 與 `3` 同時是「上一頁」與「跳第 1 名」——原版的跳表把 `1` 導到上一頁那一支,
所以**按 `1` 是上一頁不是跳第一名**;真正能跳的是 `2`..`6`(而 `2`/`4` 又同時是下一頁)。
這團按鍵重疊是原版的樣子,引擎照抄(`ZtatsKey` 只處理 `0` 與 `1`..`6`,
方向鍵另走 `ZtatsPage`)。

★ 超出隊伍人數的數字鍵**什麼都不做** —— 原版先擋 `鍵碼 − 0x31 >= 人數`。

## 1. ★ 隊員數值頁(`sub_1DECC`)

```
標題 = 角色名(紀錄起點 3DDB4 + 索引×32)
職業 = off_40B50[strchr("AMBFDTPRS", rec[0x0A])]    ; "Avatar"…
狀態 = off_40B8C[strchr("GPDSC",     rec[0x0B])]    ; "Good Health"…

列1  印字元 rec[9]  " Lv-"  rec[0x16](寬1,空白)  ' '  職業名
列2  狀態名**置中於 15 欄**:if (0 < len < 15) 先印 (15−len)/2 個空白
列3  0xFB  "Str="  rec[0x0C](寬2,★ 填 '0')  "  HP:"  rec16[0x10](寬4,空白)
列4        "Int="  rec[0x0E](寬2,填 '0')      "  HM:"  rec16[0x12](寬4,空白)
列5        "Dex="  rec[0x0D](寬2,填 '0')      "  Ex:"  rec16[0x14](寬4,空白)
     "\n\n    Magic:"  rec[0x0F](寬2,空白)
```

★★ 三個排版細節,少一個畫面就對不上:

1. **三圍用 `'0'` 補位**(`push 30h`),HP / HM / Ex / Lv / Magic 用空白(`push 20h`)。
2. **狀態名在 15 欄裡置中**,而長度**等於** 15 時不置中也不截斷(`jge` 排除)。
3. 第一行開頭把**性別位元組當字元印**(`sub_27230(rec[9])`,0x0B / 0x0C)——
   靠 FM Towns 字型的那兩格是♂/♀。引擎沒有那套字型,改印同義的 Unicode 符號。

⚠ **顯示順序是 Str / Int / Dex,而紀錄位移是 0x0C / 0x0E / 0x0D** ——
Dex 與 Int 在紀錄裡是反的。照著顯示順序抄位移會抄錯。

## 2. 隊員裝備頁(`sub_1E128`)

掃紀錄 **0x19..0x1E** 六個部位(頭 / 甲 / 右手 / 左手 / 戒 / 護符),
每格丟給 `sub_1E0F0` 印名字並累加回傳值;**六格加起來是 0** 才印 `"(None ready)"`。

⚠ 判準是總和,不是逐格 —— 只要有一件就不印那句。

## 3. Equipment(`sub_1E210`)

```
" Food: "   word_3DFB4   寬 4,空白
" Gold: "   word_3DFB6   寬 4,空白
"\n\n Keys......."  byte_3DFB8  寬 2
" Gems......."      byte_3DFB9  寬 2
" Torches...."      byte_3DFBA  寬 2
if (byte_3DFBB != 0) " Grapple"        ; ★★ 有才顯示,而且**不印數量**
```

★★ 抓鉤那一行後面**沒有 `sub_23A24`** —— 它是有/無而不是數量。

## 4. ★★ Items 頁:38 筆散落的道具被搬進一個陣列

`sub_1E8D4` 在進 Ztats 之前把散在存檔各處的持有量搬成連續的 `byte_40BA0[38]`。
名字表 `off_408F4`(38 個指標)是從 **`WORRIORS.EXP` 的一手位元組**跟指標讀出來的 ——
不是從 `.asm` 的註解抄:**IDA 對重複字串不加註解**,靠註解一定漏(`!` 出現八次)。

| idx | 縮寫名 | 來源 | 備註 |
|---|---|---|---|
| 0..7 | `*VL` `*RH` `*IS` `*IA` `*IQW` `*KXC` `*IMC` `*AT` | `byte_3E030[i]` | 卷軸,名字是咒語縮寫 |
| 8..15 | `!` ×8 | `byte_3E038[i]` | ⚠ **八瓶藥水全叫 `!`**,顏色看不出來 |
| 16 | `Magic Crpt` | `byte_3DFBC` | 魔毯,**真的數量** |
| 17 | `Skull Keys` | `byte_3DFBD` | 骷髏鑰匙,真的數量 |
| 18 | `Amulet` | `byte_3DFBF` | 旗標 |
| 19 | `Crown` | `byte_3DFC0` | 旗標 |
| 20 | `Sceptre` | `byte_3DFC1` | 旗標 |
| 21..28 | `(0`..`(7` | `byte_3E050[i] == 0xFF ? 0xFF : 0` | ⚠⚠ **沒寫完的佔位名** |
| 29..31 | `Shard/Falsehd` `Shard/Hatred` `Shard/Cowrdce` | `byte_3DFC4[i]` | |
| 32 | `Spyglass` | `byte_3DFC8` | |
| 33 | `HMS Cape Plan` | `byte_3DFC9 != 0 ? 0xFF : 0` | ★ 唯一做布林轉換的一筆 |
| 34..37 | `Sextant` `Pocket Watch` `Black Badge` `Wooden Box` | `byte_3DFCA`..`byte_3DFCD` | |

### ★ 三處看起來像 bug 而其實是原版

1. **八瓶藥水全叫 `!`** —— 玩家在這一頁分不出顏色。
2. **月石叫 `(0`..`(7`** —— 明顯是作者留著沒改的佔位字串。
3. **旗標型的東西顯示 255**:原版把「有」搬成 `0xFF`,而 `sub_1E588` 照數字印
   ⇒ 王冠那一行寫著 `255`。

⚠ 月石那一條的**極性容易看反**:

```asm
cmp     byte_3E050[ecx], 0FFh
jz      short loc_1E900      ; ← 等於 0xFF 才跳
sub     eax, eax             ;   不等於 → 0
jmp     short loc_1E905
loc_1E900:
mov     eax, 0FFh            ;   等於   → 0xFF
```

`byte_3E050` 是月石的**地點**欄,`0xFF` = 還沒埋(還在身上)
⇒ **列出來的是身上的月石,埋下去的就從清單消失**。

### ✅ 順手釘死黑徽章的存檔位移(WORKLIST §5.3 −1)

`sub_1E8D4` 把 `byte_3DFC8`..`byte_3DFCD` **六格連續**搬進清單第 32..37 筆,
而那六個名字(`Spyglass` `HMS Cape Plan` `Sextant` `Pocket Watch` `Black Badge`
`Wooden Box`)與 U 指令的 case 標註逐一相符
⇒ 中間沒有空格,**黑徽章就是 0x0218**。已接進 `u5data.Save`。

## 5. 清單頁(`sub_1E588`)

`sub_1E588(標題, 筆數, 計數陣列, 名稱表)` —— 四頁共用。
★ **數量 0 的不列**,而全部是 0 時就是**一頁空白**(沒有「(無)」那種提示)。照抄。

## 6. 引擎落地

| 原版 | 引擎 |
|---|---|
| `sub_1E9A0` 的 `esi` | `Ztats.Page`(**一個**欄位,不拆成人+頁)|
| 下一頁 / 上一頁 | `ZtatsNext` / `ZtatsPrev` |
| 鍵 `0` / `1`..`6` | `(*State).ZtatsKey` |
| `sub_1DECC` | `ztatsStatsPage`(含 `'0'` 補位與 `ztatsCenter`)|
| `sub_1E128` | `ztatsArmsPage` + `ztatsEquipSlots` |
| `sub_1E210` | `ztatsEquipmentPage` |
| `sub_1E588` | `ztatsListLines` |
| `sub_1E8D4` | `(*State).ztatsItemCounts` |
| `off_408F4` | `u5data.ZtatsItemNames` + 十四個位置常數 |
| `0FFh` 旗標 | `u5data.ZtatsFlagShown` |
| `byte_3DFCC` | `u5data.Save.Badge`(**新接**)|

### 一條被改掉的舊斷言

`TestZtatsPagesThroughTheParty` 斷言「翻頁在隊伍範圍內繞回」——
那是引擎自己發明的模型(當時只有隊員頁)。真的接縫在 Armaments 與第一名之間。
改名為 `TestZtatsWrapsThroughArmamentsNotThroughTheParty`。

### ⬜ 還沒做

- `sub_1E0A4(1)` / `sub_28F80` / `sub_290B4` 是畫框與標題列,引擎用自己的文字面板代替。
- `0xFB` / `0xFC` / `0xFE` 三個圖形字元(框線)沒有重現 —— 需要原版字型的那幾格。
- `sub_1E0F0` 印裝備名時的欄位對齊還沒逐字比對。
