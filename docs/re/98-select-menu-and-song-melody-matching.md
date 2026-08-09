# 98 — 選人選單 `sub_2A7F4`,以及用旋律把兩套音樂對上曲號

| | |
|---|---|
| 輸入檔 | `re_work/fmtowns/WORRIORS.EXP.asm` / `.i64`;`re_work/fmtowns/iso/U5_E/M*.EUP`;`gamedata/upgrade/*.xmi` |
| SHA-256 | 見 `re_work/fmtowns/SHA256SUMS.txt` |
| 主要函式 | ★ `sub_2A7F4`(「Select:」選單)、`sub_2A974`(包一層)、`sub_E19C`(選隊員)、`sub_2BC18` / `sub_2BBDC`(游標上下限) |
| 工具 | ★ `tools/match_songs.py`(新)—— 用音高差分的最長共同子串配對兩套音樂 |
| 起因 | `docs/re/97` §3b 把選單的規格寫完卻沒接(11 個呼叫端要改成 callback);同一輪要把 MT-32 那套音樂接進引擎,而它的檔名是曲名不是曲號 |
| 狀態 | ✅ 兩件都落地 |

---

## 1. `sub_2A7F4` 全文與三個關鍵行為

```
印 "Select:\n"
sel = 0
迴圈 {
    畫清單(每一列印隊員名字,游標那一列反白)
    鍵 = 等按鍵
    if ('1' <= 鍵 <= '6')  { i = 鍵 − '1'; if (i < 隊伍人數) sel = i; continue }
    if (鍵 == 1 || 鍵 == 3) { sel = sub_2BC18(sel);  continue }   ; ↑ ← 游標 −1(下限 0)
    if (鍵 == 2 || 鍵 == 4) { sel = sub_2BBDC(sel);  continue }   ; ↓ → 游標 +1(上限 人數−1)
    if (鍵 == '0')          { sel = −1; break }
    if (鍵 == 0x20 || 鍵 == 0x1B) break                          ; 空白 / ESC
    if (鍵 == 13 && arg_0)  { sel = −2; break }                  ; Enter(只有 arg_0 非 0)
}
return sel
```

`sub_E19C` 用它的那一段:

```
印 "Player: "
sel = sub_2A7F4(允許Enter)
if (sel < 0)                    → 印 "None!\n";  return −1
if (狀態 != 'G' && 狀態 != 'P')  → ★ 印 "Disabled!\n\n";  重問(jmp 回選單)
印 隊員[sel].名字
return sel
```

★★ 三件少一個手感就不對的事:

1. **游標可以停在死人身上。** 原版不把不能行動的人從清單裡藏起來 ——
   按下去才印「Disabled!」,然後**回到選單**。藏起來的話玩家不會知道
   那是狀態問題,只會覺得少了一個人。
2. **`'1'..'6'` 與方向鍵並存。** 方向鍵是鍵碼 1..4,在 `'0'`(0x30)之下,
   兩組不衝突,所以原版兩種操作都收。
   ⚠ 上限是 **`'6'`**(隊伍最多六人),而 `sub_2BD40`(數字鍵指定行動者)
   收的是 **`'0'..'9'`** —— 兩支的鍵範圍在原版就不一樣,不要「統一」。
3. **`'0'` 與空白鍵都走「不選任何人」那個出口**,而那個出口印的是
   `"None!"`,不是通用的取消訊息。

## 2. 為什麼「同步版」必須刪掉

引擎原本有一支同步的 `pickCharacter(prompt) int`,多人時直接取**最後一位**
能動的隊員。這一輪把 11 個呼叫端全部改成 `pickMember(prompt, then)` 之後,
把那一支**刪掉**而不是留著:

> 只要同步版還在,新的呼叫端就會挑它寫(同步的比較好寫),
> 而**每一個這樣的呼叫端都是一個靜靜跳過選單的地方**。

⇒ 不留同步版本,讓「要問」成為唯一的路。

### `then` 一定會被呼叫一次

不必問時同步呼叫;要問時在玩家選完**或取消**之後呼叫。取消 → `then(-1)`。
這一條是 `Picker.onCancel` 存在的理由 —— 少了它,玩家按 ESC 之後呼叫端的
流程會**靜靜停住**,而畫面上什麼異常都沒有,比報錯難查。

### 兩個呼叫端**刻意**不檢查 `-1`

`searchDungeonChest` 與炸彈坑那一段,原版拿 `sub_E19C` 的回傳值直接去查敏捷 ——
沒人可選時算出來的門檻就是「敏捷 0」。照抄。

## 3. ★ 用旋律把 MT-32 那套對上曲號

### 問題

`U5_BGM.TBL` 給的是**曲號 → `.EUP` 檔名**(`M1`、`M92`、`M152`…),
而 DOS upgrade 那套的檔名是**曲名**(`U5THEME.XMI`、`HALLS.XMI`…)。
兩邊沒有共同的鍵,而「哪一首 XMI 配哪個場景」**不在任何資料檔裡** ——
upgrade 是把它改寫進 `ULTIMA.EXE` 的程式碼。

### 做法:比旋律(`tools/match_songs.py`)

兩套是同一批曲子的不同編曲 ⇒ 拿音高**差分序列**(不受移調影響)的
最長共同子串配對,逐聲道兩兩比、取最大值。15 × 16 = 240 對,
**正反兩個方向**都算。

| 結果 | |
|---|---|
| 雙向一致 | **14 組**,分數 47..423(次佳一律 < 1/2)|
| 旋律平手 | `REUNION` ⇄ `M14`,由音數 128/126、聲道 6/6 收尾 |
| 配不上 | `AMIGA` —— 對任何 `.EUP` 與任何其他 XMI 都 ≤ 10 ⇒ **它沒有曲號** |

最終表在 `internal/u5data/mt32.go`(`MT32Tracks`),15 首全滿。

### ⚠⚠ 兩個差點判錯的地方

1. **分數要對照曲子長度讀,不能看絕對值。**
   `M14.EUP` 每聲道只有 ~22 個音,所以它拿到 21 已經是**整條旋律都一樣**;
   而「≥12 且是次佳的兩倍」這種絕對門檻會把它判成「沒對上」。
2. **同一句樂句會出現在兩首曲子裡。**
   `REUNION`(短的重奏,14 秒)與 `RULEBRIT` 共用開頭那句,所以
   `REUNION×M14`、`REUNION×M152`、`RULEBRIT×M14` **三個組合都拿 21** ——
   旋律這**一個**訊號分不開它們。分開它們的是規模:

   | | 音數 | 長度 | 聲道 |
   |---|---|---|---|
   | `REUNION.XMI` | 128 | 14.3s | 6 |
   | `M14.EUP` | 126 | 10.7s | 6 |
   | `RULEBRIT.XMI` | 460 | 44.7s | 8 |
   | `M152.EUP` | 397 | 37.3s | 6 |

   ⇒ 平手時用**第二個訊號**(音數 + 聲道數),不是硬選一個。
   這與 `CONTEXT.md` 裡「同一個觀察可以支持兩個不同的模型 ⇒ 要用能區分
   它們的檢查」是同一條規則的另一個實例。

### ★★ 三方交叉驗證

配對表可以拿 `docs/re/87` 從 32 個 `sub_3181C` 呼叫點逆出來的**曲號用途**
去對 —— 而那是完全獨立的證據來源(程式碼 vs 旋律):

| 曲號 | upgrade 作者標的曲名 | 逆向得到的用途 |
|---|---|---|
| 1 | Britannic Lands | `sub_86C` 回地表 |
| 2 | Cap'n Johne's Hornpipe | `sub_16F08` 上船 |
| 3 | Engagement and Melee | `sub_A9EC` 進戰鬥 |
| 9 | Halls of Doom | `sub_2D564` 進地牢 |
| 0Ah | Worlds Below | `sub_86C` 進幽冥界 |
| 7 | The Missing Monarch | `sub_2D72C` 進場景(城鎮)|

五條獨立吻合(第六條合理)⇒ 旋律配對、「曲號 = `U5_BGM.TBL` 的列號」、
逆出來的曲號語意**三者互為佐證**。比其中任何一條單獨的證據都強。

### ⬜ 還沒解的

- `AMIGA`(Amiga Theme,383 秒)原版 upgrade 有沒有在某處播它 ——
  要逆 patch 過的 `ULTIMA.EXE`。引擎目前不播(沒證據就不接)。
- 曲號 0 的曲名是 `Ultima V Theme`,而 `docs/re/87` 從 `sub_A9EC` 逆到的是
  「打贏之後」。兩者不衝突(主題曲當凱旋樂用),但也**不算互相佐證** ——
  這一條只有一個證據來源。

## 4. 引擎落地

| 原版 | 引擎 |
|---|---|
| `sub_2A7F4` 選單 | `(*State).beginPickMember` + 共用的 `Picker` |
| `sub_E19C` 三條路 | `(*State).pickMember(prompt, then)` |
| `'1'..'6'` / `'0'` / 空白 | `PickMemberDigit`(回 false = 這個鍵不歸它管)|
| 方向鍵四顆都能動游標 | `cmd/u5cht` 的 `PickIsMember()` 分支 |
| `aDisabled` + 重問 | `MsgDisabled` + 回呼裡再開一次選單 |
| `aNone` | `Picker.cancelMsg = MsgActiveNone`(不印通用的「作罷。」)|
| 「能動」= 'G' 或 'P' | `ableMembers()` |
| ~~同步的 `pickCharacter`~~ | **已刪**(見 §2)|
| 曲號 → MT-32 曲名 | `u5data.MT32Tracks` / `MT32ExtraTracks` |
| 兩套音樂切換 | `audio.Source` / `Player.SetSource` / `NextSource`;F5 熱鍵 + `-music` 旗標 |

### 測試

| 測試 | 驗什麼 |
|---|---|
| `TestOneAbleMemberIsNotAsked` | 一個人能動就不問(`opened` 這個回傳值是判別力所在)|
| `TestNobodyAbleAnswersMinusOne` | 沒人能動:同步回 −1,**不開選單** |
| `TestTwoAbleMembersOpenTheMenu` | 2 人以上要問 —— 舊版最大的落差 |
| `TestTheMenuListsEveryoneIncludingTheDead` | ★ 清單含死人 |
| `TestDisabledMemberReAsks` | ★★ 印「無法行事!」之後選單**還開著** |
| `TestCancelStillCallsThen` | 取消也回呼一次(值 −1)|
| `TestZeroAndSpaceLeaveTheMenu` | `'0'` / 空白 → −1,而且印的是「無!」不是「作罷。」|
| `TestDigitOutsideThePartyIsSwallowed` | 超出人數不動游標但算吃掉;`'7'` 不歸它管 |
| `TestDigitOnlyWorksInTheMemberMenu` | 別的選單不吃數字鍵 |
| `TestPoisonedCanStillAct` | 中毒算能動 |
| `TestSongNumbersMapToBothSets` | 兩套的曲號對應(含 `M92` / `M152` 那兩個不連續的)|
| `TestMT32LookupIsCaseInsensitive` | ★★ 檔名大小寫翻過來也要找得到 |
| `TestSourceDefaultsToWhicheverIsRendered` | 只渲染一套時預設用那一套 |
| `TestSwitchingSourceRestartsTheCurrentSong` | ★★ 切換要重播,否則按了沒反應 |
| `TestSwitchingToAnUnrenderedSourceIsRefused` | 沒渲染的不能換過去(靜音沒有解釋比較糟)|

### ★ `actAs` 這個測試輔助為什麼不算後門

15 個既有測試在改成 callback 之後紅了 —— 它們驗的是「效果對不對」,
而選單一開,效果就發生在回呼裡。用的解法是 `actAs(t, s, 0)`:
**原版自己的機制**(數字鍵「指定行動者」`sub_2BD40`),不是測試專用的旁路。

差別很重要:後門會遮住選單那條路的 bug(`CLAUDE.md §6.1`:debug hook
會讓回歸測試全綠而玩家一開就壞),而 `actAs` 走的是玩家真的按得到的鍵。
