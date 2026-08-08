# 93 — 酒館:一餐從來沒有議價過(以及桌上真的會出現菜)

| | |
|---|---|
| 輸入檔 | `re_work/fmtowns/WORRIORS.EXP.asm` / `.i64` |
| SHA-256 | 見 `re_work/fmtowns/SHA256SUMS.txt` |
| 主要函式 | `sub_216C8`(酒館主流程)、`sub_210D8`(點餐)、★ `sub_20E6C`(算總價)、★★ `sub_20F60`(報價 + 收錢 + **上菜**)、★ `sub_20ED0`(英文數字)、★★ `sub_20F24`(sir / milady,**有死碼**)、`sub_21108`(酒單)、`sub_1B294`(進店) |
| 主要資料 | `word_3EF38`(總價)、`dword_56E14`/`dword_56E10`(人數)、`dword_56E1C`(★ 喝了幾杯)、`dword_56E18`(上了菜的 tile)、`word_3EF34`(★ 這一類的第幾家,**不是價格**) |
| 起因 | 清 §5.2 截斷清單裡的 `sub_20F24` / `sub_20F60` |
| 狀態 | ✅ 五支全解;**推翻 `docs/re/10` 的兩條斷言**;三個新機制落地 |

---

## 0. ⚠⚠ 推翻:餐點不議價,而且酒不是「唯一」不議價的

`docs/re/10` 有兩條互相佐證的錯誤斷言:

```
| 餐點 | sub_210D8 → sub_20E6C | Haggle(單價 × 活著的人數) | …
⚠ 酒是全遊戲**唯一不議價**的交易
```

兩條都錯。做法是**全檔掃描議價公式的指紋** —— `docs/re/10` §3 自己記了那段組語,
其中 `mov ecx, 64h` 緊接 `sub ecx, edx`(= `100 − 3×INT`)是它獨有的形狀:

| 含議價公式的函式 | 是什麼 |
|---|---|
| `sub_112F8` `sub_11588` `sub_118CC` `sub_11AF0` | 四種商店的買 |
| `sub_21310` | ★ **酒館的乾糧** |
| `sub_219B0` | 造船廠 |
| `sub_21D48` `sub_22018` `sub_22280` | 旅店的住 / 寄放 / 領回 |

**九支,一支不多。** 餐點路徑的三支 —— `sub_210D8`、`sub_20E6C`、`sub_20F60` ——
**一支都不在**。而同一家酒館的乾糧在。

⇒ 餐點的價就是 `單價 × 活著的人數`,沒有智力折扣;不議價的交易有**兩項**(酒 + 餐點)。

### 錯因

`CLAUDE.md §4.5` 那條規則正中:**「唯一」「只有一處」沒有全檔掃描佐證就不要寫。**

而這次特別值得記的是**兩條錯誤斷言互相掩護**:

- 「酒是唯一不議價的」讓人不會回頭查餐點 —— 既然唯一,餐點當然議價。
- 「餐點會議價」又讓那個「唯一」看起來成立 —— 檢查過一項了。

任一條單獨存在都會被下一次閱讀抓到;兩條放在同一張表裡就形成了閉環。
⇒ **推翻一條斷言時,要順手檢查「當初支持它的那條」是不是也錯了。**

引擎因此把玩家的一餐算貴了(智力 16 的店多收 50%)。已改成原價。

## 1. `sub_20E6C(單價)` —— 算總價,死人不算

```
word_3EF38 = 0                              ; 總價
dword_56E14 = dword_56E10 = 0               ; 人數 / 份數(兩個同步遞增)
for (i = 0; i < byte_3E06B; i++) {          ; 隊伍人數
   if (隊員[i].狀態 == 'D') continue        ; ★ 死人不吃也不收錢
   dword_56E14++;  dword_56E10++
   word_3EF38 += 單價
}
印 '\n' '\n'
```

單價來自 `sub_210D8`:`sub_20E6C(dword_56E20[word_3EF34])`。

### ★ `word_3EF34` 是「這一類的第幾家」,不是價格

它有 **25 個讀取點、只有兩個寫入點**(`mov …, 0` 與 `inc …`)——
正是 `CLAUDE.md §4.5` 記的「讀很多寫很少 → 去看取址」的形狀。
但這次不是間接寫入,是我一開始**把讀取點誤認成價格來源**。真相在 `sub_1B294`(進店):

```
word_3EF36 = 店種
word_3EF34 = 0
while (byte_4185C[店種*16 + word_3EF34] != byte_3E0A3 && word_3EF34 < 0x10)
   word_3EF34++                            ; ★ 找「開在這個地點」的那一家
dword_3EF24 = off_4145C[…]                 ; 店名 "Iolo's Bows"
dword_3EF28 = off_4165C[…]                 ; 店主 "Gwenneth"
```

⇒ 它是引擎的 `Shop.TypeIndex`。

## 2. `sub_20F60` —— 報價、收錢,然後**把菜端上桌**

```
印 '"'  "That will be "  <總價>  " gold for the "
sub_20ED0()                                ; ★ 人數的英文數字
印 " of ye,\n"
sub_20F24()                                ; ★ "sir" / "milady"
印 '.'
if (總價 > 金幣) {
   印 "\"\n\n\"CAN'T PAY?\nBeat it!\"\nyells " <店主> "."
   return 1
}
金幣 -= 總價;  sub_11190()
if (份數 != 0) {
   存糧 += 份數                             ; 上限 0x270F = 9999
   if (北邊那一格 == 0x95) { 改成 0x9B; dword_56E18 = 0x9B }
   else if (南邊那一格 == 0x95) { 改成 0x9A; dword_56E18 = 0x9A; 重畫 }
   else dword_56E1C++                       ; ⚠ 見下
} else dword_56E1C++
印 "\nEnjoy!\"\n\n"
return 0
```

### ★★ 桌上真的會出現菜

`0x95` 是空桌,`0x9B` / `0x9A` 是上了菜的桌(兩個朝向)。這是**直接寫進地圖緩衝**
的一格,與月門同一個手法(`docs/re/86`)—— 不是動畫、不是物件。

⬜ **原版沒有把它收回去的程式碼**:`dword_56E18` 只被寫、沒有讀取點,
所以菜就一直留在桌上到重新載入場景為止。引擎照抄(不自己補清除)。

⚠ 兩邊都有桌子時**只放北邊**(`if/else if`),而且只有南邊那一支多叫一次 `sub_29D64`
重畫 —— 北邊那一支不叫。看起來像疏漏,照抄。

### `sub_20ED0` 只認 2..6

```
switch (人數) { 2:"two" 3:"three" 4:"four" 5:"five" 6:"six" default: 什麼都不印 }
```

⇒ **一人隊伍在原版會印成 "gold for the  of ye"**(中間是空的)。這是原版行為,
引擎的 `partyCountWord` 同樣對 1 回空字串,不「順手」補上「一」。

### ★★ `sub_20F24` 的死碼:永遠看名冊第 0 筆

```asm
mov     al, 0FFh
mov     byte_3E08B, al          ; ★ 副作用:把 byte_3E08B 寫成 0xFF
and     al, al
jz      short loc_20F34         ; ← al 是 0xFF,這個 jz 永遠不跳
sub     edi, edi                ; ★ 永遠走這條:索引 = 0
jmp     short loc_20F3B
loc_20F34:
movzx   edi, byte_3E08B         ; 走不到
loc_20F3B:
if (byte_3DDBD[edi*32] == 0x0B) 印 "sir" else 印 "milady"
```

⇒ 不管跟誰說話,稱呼都由**名冊第 0 筆的性別**決定。原版的 bug,照抄。

⬜ 那個副作用(`byte_3E08B = 0xFF`)引擎**沒有**重現 —— 該位元組的語意還沒查。
要是有人讀它,這就是一條 finding。

## 3. ★ `dword_56E1C`:剛好第三杯才會被勸

`sub_21108`(酒單)開頭:

```asm
cmp     dword_56E1C, 3
jnz     loc_211A9                ; ★ 不等於 3 就直接進酒單
印 "\n\n\"I beg thy\npardon, "  <sir/milady>  ",\"\nsays "  <店主>
印 ".\n\"But haven't\nye had enough\nto drink?\" "
等 Y / N
```

**判準是 `jnz`,也就是「等於 3」** —— 不是 `>= 3`。⇒ 喝到第四杯之後那一問
就再也不出現。這是那種「看起來像 bug 所以很容易被『修好』」的原版行為。

計數器 `dword_56E1C` 在兩處遞增:

- `sub_21108`:每買一杯酒 +1(合理)
- `sub_20F60`:點餐而**份數為 0**(全隊都死)時 +1 —— 荒謬但無害,
  而且實務上走不到(能操作的隊伍必有活人)。原版留著,引擎不重現這一支
  (⬜ 它需要一個走不到的狀態,補上去只會變成永遠不執行的碼)。

## 4. 引擎落地

| 原版 | 引擎 |
|---|---|
| `sub_20E6C` 不議價 | `quoteMeal` 移除 `u5data.Haggle`(**bug 修掉**)|
| `sub_20ED0` | `partyCountWord`(只認 2..6)|
| `sub_20F24` 死碼 | `addressWord`(永遠看 `Roster[0].Gender`)|
| `sub_20F60` 尾段 | `serveOnTheTable` + `ShopItem.ServeOnTable` |
| 0x95 / 0x9B / 0x9A | `TableEmpty` / `TableServedNorth` / `TableServedSouth` |
| `cmp dword_56E1C, 3` | `TavernDrinkNagCount` + `ShopModeTavernNag` |
| `word_3EF34` | 既有的 `Shop.TypeIndex`(**不是**價格)|

### 測試

| 測試 | 驗什麼 |
|---|---|
| `TestTavernMealIsNotHaggled` | 價 = 單價 × 活人數;**含反對照**(議價值必須不等於原價,否則沒有鑑別力)|
| `TestTavernMealCountsLivingOnly`(改) | 死人不算;期望值從 `Haggle(...)` 改成原價 |
| `TestMealIsServedOnTheTable` | 四種桌子配置(只北 / 只南 / 都有 → 只放北 / 都沒有)|
| `TestBartenderNagsAtExactlyThreeDrinks` | 0..5 杯逐一驗,只有 3 會勸 |
| `TestNagLetsYouThroughOnYes` | 答 Y 進酒單、答 N 回選單 |
| `TestAddressWordAlwaysLooksAtRosterZero` | 改第 1 筆不變、改第 0 筆才變 |
| `TestPartyCountWordOnlyCoversTwoToSix` | 0..8 的邊界,1 回空字串 |
