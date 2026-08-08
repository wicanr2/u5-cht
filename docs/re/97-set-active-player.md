# 97 — 數字鍵是一個指令:「指定行動者」

| | |
|---|---|
| 輸入檔 | `re_work/fmtowns/WORRIORS.EXP.asm` / `.i64` |
| SHA-256 | 見 `re_work/fmtowns/SHA256SUMS.txt` |
| 主要函式 | ★ `sub_2BD40`(Set Active Plr)、`sub_E19C`(選隊員,見 `docs/re/95`)、`sub_4B14`(按鍵分派的 default) |
| 主要資料 | ★ `byte_3E08B`(當前指定的隊員,0xFF = 沒指定;**進存檔**) |
| 工具 | `tools/triage_unread.py`(新)—— 用「碰不碰遊戲狀態」排未讀函式的優先序 |
| 起因 | 三選之後 `sub_2BD40` 出現在「動到角色紀錄」那一群 |
| 狀態 | ✅ 全解並落地;收掉 `pickchar.go` 自己標的「引擎目前沒有單人狀態」 |

---

## 0. ★★ 先講三選的結果:702 支未讀裡只有 27 支動到遊戲狀態

`tools/triage_unread.py` 用「碰不碰**角色紀錄 / 物件表 / 地圖**」分群:

```
反組譯 1232 支、未讀 702 支;其中動到遊戲狀態的 27 支
剩下 675 支不碰遊戲狀態(文字排版 / 繪圖 / 檔案 / 純計算)
```

⇒ 那 675 支是 FM Towns 的執行期函式庫與繪圖層,**Go + Ebiten 已經取代掉**。
按「行數」或「引用字串數」排會把它們排到前面 —— 換掉判準之後,
剩下的遊戲邏輯是一份**有界的 27 支清單**。

⚠ 訊號是啟發式不是證明:`sub_3007C`(521 行,開新遊戲的腳本解譯器)
一個訊號都沒中。⇒ **分數低不等於可以跳過**;整批跳過某一群時
至少要抽樣看過離群的那幾筆(`WORKLIST §5.2b` 的教訓)。

## 1. `sub_2BD40(鍵碼)` 全文

```
印 "Set Active Plr:\n"
if (鍵 == '0') { 印 "None!\n"; byte_3E08B = 0FFh; 重畫狀態列; return 0 }
i = 鍵 − 0x31                                       ; '1'..'9' → 0..8
if (i >= byte_3E06B)          → 印 "Invalid!\n"; return 1
if (隊員[i].狀態 == 'D')       → 印 "Invalid!\n"; return 1
if (隊員[i].狀態 == 'S')       → 印 "Invalid!\n"; return 1
byte_3E08B = i
印 隊員[i].名字  '\n';  重畫狀態列;  return 0
```

★ 四件事:

1. **收 '0'..'9' 十個鍵**(呼叫端 `cmp edi, 30h` / `cmp edi, 39h`),不是 '0'..'6'。
   '7'..'9' 會算出超出人數的索引 → 「Invalid!」,但**仍然算它管的鍵**。
   少收三個鍵的話玩家按 8 會落到別的指令去。
2. **死了或睡著了不能被指定**,而**中毒可以** ——
   與 `sub_E19C` 自動掃描的「'G' 或 'P' 算能動」一致。
3. `'0'` 是**取消**,不是「指定第 0 位」。
4. 它**不在字母指令表裡**(`sub_2ACF4` 的 A..Z),是按鍵分派 default 的一支。

## 2. ★ 這是 `sub_E19C` 第一條分支要的狀態

`docs/re/95` §3 已解出 `sub_E19C`(選隊員)的三條路:

```
if (地點碼 > 0x80)            → 戰鬥中的當前單位
else if (byte_3E08B != 0xFF)  → ★ 這一條:玩家指定過的人,不問
else                          → 掃「'G' 或 'P'」,只有一個就不問
```

而引擎的 `pickchar.go` 自己在註解裡寫著「引擎目前沒有單人狀態,所以第 1 條先不做」——
`sub_2BD40` 就是那個狀態的來源。已接上。

### ⚠ 原版**不重驗狀態**

指定之後那個人被打死、被催眠,`sub_E19C` 照樣把他回傳(它只檢查 `!= 0xFF`)。
⇒ 喝噴泉、看水晶球那些指令真的會落在死人頭上。照抄,不「順手」加判斷。

## 3. ⬜ 它進存檔,而位移還沒定位

`sub_27D24`(讀檔)與 `sub_284CC`(寫檔)都碰 `byte_3E08B`(`docs/re/95` 的全檔掃描:
25 支函式讀寫)。⇒ 它是一個**存檔欄位**,而引擎目前只放在記憶體裡。
⬜ 位移待定 —— 存檔時會遺失「指定了誰」,重開之後回到「沒指定」。

## 4. 引擎落地

| 原版 | 引擎 |
|---|---|
| `sub_2BD40` | `(*State).SetActivePlayer`(回 false = 這個鍵不歸它管)|
| `byte_3E08B` | `State.activeMember` + `activeSet` 旗標 |
| `0xFF` | `ActiveNone` |
| `'0'..'9'` | `ActiveKeyNone` / `ActiveKeyFirst` / `ActiveKeyLast` |
| `sub_E19C` 第一條 | `pickCharacter` 的 `activeIfUsable()` 分支 |
| 按鍵分派 | `cmd/u5cht` 的 `commandKeys` 開頭(**擋在字母表之前**)|

### ★ `activeSet` 旗標為什麼必要

`State` 到處被寫成結構常值(沒有建構子),而 `activeMember` 的零值 0
會被讀成「指定了第一位」—— 那會讓喝噴泉、看水晶球全部落在同一個人頭上。
同 `song`/`songSet` 的處理。`TestFreshStateHasNoActiveMember` 釘住這一點。

### 測試

| 測試 | 驗什麼 |
|---|---|
| `TestFreshStateHasNoActiveMember` | 零值不等於「指定了第一位」|
| `TestSetActivePlayerAcceptsTenDigits` | 收十個數字鍵、非數字不歸它管 |
| `TestSetActivePlayerPicksAndClears` | 指定會印人名、`'0'` 取消 |
| `TestDeadOrAsleepCannotBeActive` | 四種狀態逐一(中毒**可以**)|
| `TestOutOfPartyIsInvalid` | 超出人數印「無效!」|
| `TestPickCharacterHonoursTheActiveMember` | 指定之後不再自動掃(**刻意挑與自動結果不同的人**才驗得出來)|
| `TestActiveMemberIsNotRevalidated` | 被指定的人死了仍回他 —— 原版行為 |
