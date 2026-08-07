# 61 — 酒館的線索系統,以及轉入 U4 少掉的那一半

| | |
|---|---|
| 輸入檔 | `U5_E/WORRIORS.EXP`(FM Towns,SHA-256 見 `docs/re/00-hexrays-p3-verified.md`) |
| | `DATA.OVL`(DOS,48,464 B)、`SHOPPE.DAT`(10,135 B) |
| 位址 | `sub_21500`(酒館打聽)、`sub_27C98`(關鍵字比對)、`sub_11168`(印 `SHOPPE.DAT`) |
| | `sub_7594`(轉入的主流程)、`sub_7564`(三圍換算曲線)、`sub_71D0`(讀 U4 存檔) |
| 落地 | `internal/u5data/tavernlore.go`、`internal/game/tavern.go`、`internal/u5data/transfer.go` |

兩條 ⬜,一條是缺功能,一條是**只做了一半而不知道**。

---

## 一、酒館的打聽消息 = U5 的線索系統

`WORKLIST` 上這條掛著「要八德與地牢的知識表」。表就在 `DATA.OVL` 裡,五張連著放:

```
0x4C84  26 × u16  關鍵字指標   hone comp valo just sacr hono spir humi
                               dece desp dest wron cove sham hyth
                               crow scep amul  fals hatr cowa
                               astr oppr brit resi unde
0x4CB8  26 × u16  人名指標     Malik Greyson Trian Jeremy Rew Gruman Saul
                               Shirita Malifora Annon Trian Felespar
                               the mother of Rew Sindar Kaiko Terrance
                               Greymarch "Simon and Tessa" Shalineth
                               "a daemon" "Lord Malone" Zachariah Tactus
                               "a daemon" Terrance Jotham
0x4CEC  26 × u8   地名索引     0 1 2 3 4 5 6 7 0 1 2 3 4 5 7 1 3 9 B A C 0 4 A 1 8
0x4D06  13 × u16  地名指標     Moonglow Britain Jhelom Yew Minoc Trinsic
                               "Skara Brae" "New Magincia"
                               "a lighthouse south of Britain"
                               "a hidden mountain keep" "the desert"
                               "the Lycaeum" "Serpent's Hold"
0x4D20  26 × u16  價格         50 75 50 50 75 75 25 50 100 150 75 150 75 100
                               100 200 200 200 250 250 250 100 50 50 200 100
0x4D52  結束
```

**五段首尾相接、一個位元組的空隙都沒有**,而 26 / 26 / 26 / 13 / 26 的長度全部
剛好對上各自的筆數 —— 位移偏一格,後面四張表會一起爛掉。這個連續性就是錨。

26 個主題把整個遊戲的尋寶清單蓋完:**八德 + 七座地牢 + 三信物 + 三碎片 + 四項雜項**。
`crow` `scep` `amul` 三題各 200 金 —— 酒保會告訴你王冠與權杖在哪
(那三題與 `docs/re/57` 是同一批東西的兩個出口)。三塊碎片各 250 金,最貴;
最便宜的是靈性,25 金。

### 比對規則:純子字串,由前往後第一個中的算

`sub_27C98(關鍵字, 答案)` 把關鍵字大寫化(`al & 0x7F` 再 `sub al, 0x20`)後在答案裡
找子字串。所以打 `honesty` 或 `HONE` 都問得到誠實。

⚠ 原版在命中之後還有一段:

```asm
                and     ebx, ebx
                jz      short loc_21598      ; 0 = 沒找到 → 下一個主題
                and     ebx, ebx
                jle     short loc_2159C      ; ↓
                cmp     byte_55F37[ebx], 20h ; 命中位置前一個字元是空白?
                jnz     short loc_2159C      ; ↓
loc_2159C:      mov     edi, esi             ; ← 兩個分支與 fall-through 都到這裡
                jmp     short loc_2159F
```

**兩個條件跳與 fall-through 全部落到同一個 `mov edi, esi`** —— 那個「前一個字元是不是
空白」的檢查對控制流沒有任何影響,是編譯器留下的殘骸。所以「純子字串比對」就是原版
的行為,不是我做的簡化。

### 流程(`sub_21500`)

```
印 "Of what wouldst thou hear my lore, <隊長>?" → 讀最多 15 個字
空的 → 直接結束
比對 26 個關鍵字;都不中 → "That, I cannot help thee with." → **再問一次**(可以一直猜)
報價(SHOPPE.DAT 0x134E,`%` = 價格)+ "Fair 'nuff?" → **只收 Y / N**,其他鍵繼續等
N → "No" 結束
Y → 金幣 <= 價格?付得出來(原版是 jle,剛好等於也付得出來)
     不夠 → "Sorry, <隊長>" + 0x146A(", I must attend my PAYING customers!" says $.)
     夠   → 扣錢;`&` = 人名、`*` = 地名,四句模板**隨機挑一句**,後面接 "says <酒保>."
```

四句模板都在 `SHOPPE.DAT`,而且早就翻好了:

| 位移 | 原文 |
|---|---|
| 0x13A2 | `"Seek ye & in *!"` |
| 0x13AE | `"Rumour has it that &, who lives in *, doth possess such knowledge."` |
| 0x13D9 | `"It may be that &, of *, may be able to help thee!"` |
| 0x13F3 | `"Mayhap & in * wilt see fit to aid thee!"` |

⇒ `&`(物品名)與 `*`(地名)這兩個佔位符的用途在此結案 —— `docs/re/08` 當時
把 `*` 記成「酒館八卦用」是對的,現在知道它填的是這 13 個地名。

---

## 二、轉入 U4:`docs/re/55` 只做了一半

`WORKLIST` 上這條寫「只剩『Exp:』『Level:』那兩個字面」。去找那兩個字面時發現:
**它們不在 Ztats 畫面上,而在轉入 U4 的報告畫面裡**(`sub_7594`)。

而 `sub_7594` 才是主選單真正呼叫的那一支(`sub_6730+95D`);
`docs/re/55` 逆的 `sub_71D0` 只是它的**第一階段**(`sub_7594+11B` 呼叫)——
把 U4 存檔讀進來、驗界線。讀完之後還有第二階段:**逐項換算**,並把每一項的變化
印在畫面上。

`docs/re/55` 記的東西沒有錯,錯在以為那樣就結束了。
「等級 = 最大 HP / 100」確實是第一階段算的,**而它接著就被第二階段蓋掉**。

### 第二階段的算式

```
經驗值 = U4 經驗值 / 10                              (word_3DDC8 /= 10)
等級   = 1; n = 經驗值 / 100; while (n > 0) { 等級++; n >>= 1 }
HP = 最大 HP = 等級 × 30                             (eax*5 → *3 → *2)
力量 = curve(力量);  if (力量 < 20) 力量 = 20        ★ 只有力量有下限
敏捷 = curve(敏捷)
智力 = curve(智力);  法力 = 智力
```

`curve` 就是 `sub_7564`,三段式:

| 輸入 | 輸出 |
|---|---|
| v < 10 | v(原樣) |
| 10 ≤ v < 30 | (v − 9) / 2 + 10 |
| v ≥ 30 | (v − 30) / 4 + 20 |

除法是朝零截斷(`cdq; sub eax,edx; sar eax,1` 這個慣用法,連做兩次 = /4)。
U4 的三圍上限 50 → 25,而畫面上印的 `"was 45(50), now 27(30)"` 裡那兩個括號
就是**兩邊的量表上限**:U4 是 50 分制,U5 這裡是 30 分制。

### ★ 只有力量有下限,敏捷與智力沒有

`cmp al, 14h; jnb; mov …, 14h` **只出現在力量那一段**。這種不對稱憑印象一定寫不出來,
而且它與建角時的 `CreateMinStrength`(同樣是 20)是同一條規則的兩個出口。

順帶一個推論:把曲線與下限合起來看,中間那一段對力量**幾乎沒有作用**
(任何 v < 30 算出來都 ≤ 20,一律被夾成 20)。但敏捷與智力沒有下限,
所以那一段對它們是實際生效的 —— 不能因為「力量看不出差別」就把它簡化掉。

### 轉入的角色最高 5 級

U4 經驗上限 9999 → /10 = 999 → /100 = 9 → 折半到 0 需要 4 次 → 等級 5,HP 150。
這是**算出來的**上限,不是寫死的 —— 所以不要在程式裡加 `if level > 5`。

---

## 還沒做的:`sub_7594` 的互動畫面

換算的算式已經落地,但原版那個**逐頁確認**的畫面沒有:

```
"Transfer Character from Ultima IV"
"Please insert the Ultima IV Player Disk" / "and press any key"
"or press <Esc> to abort transfer"
"Found:" a level N Male/Female <class> STR: DEX: INT: <name> is an Avatar./not an Avatar
Name:  → "Keep this name?"  → "Enter new name: "
Sex:   → "Keep same sex?"
Class: → "Thou art now an Avatar:" / "Class remains intact"
Exp:   → "Experience has been converted"
Level: → "Level has been converted"
STR:   → "Strength: was 45(50), now 27(30)"
DEX:   → "Dexterity: was …"
INT:   → "Intellect: was …"
"Conversion complete" → 寫出 A:SAVED.GAM / A:SAVED.OOL
```

引擎目前把換算結果一次印成幾行訊息(改名與改性別那兩問沒有)。
畫面是介面工作,算式才是機制 —— 但**改名與改性別是玩家真的做得到的選擇**,
所以那兩問算功能缺口,記在這裡。
