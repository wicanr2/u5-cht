# 80 — 做完了卻接不到:一次全庫掃描找出四個機制

| | |
|---|---|
| 輸入檔 | `re_work/fmtowns/WORRIORS.EXP.asm`(FM Towns 英文版) |
| SHA-256 | 見 `re_work/fmtowns/SHA256SUMS.txt` |
| 主要函式 | `sub_1A38C`(用碎片)、`sub_2D9D0`(風浪)、`sub_2A1E8` / `sub_2A0C4` / `sub_2A984`(狀態列) |
| 落地 | `internal/game/use.go` · `shipdamage.go` · `wind.go` · `internal/render/scene.go` · `tools/find_unwired.py` |

---

## 1. 一個反覆出現的形狀,值得一次抓乾淨

`docs/re/71` 找到「卷軸 / 藥水 / 月石撿得到卻用不了」時,病因是**入口沒接**,
不是規則沒寫。那不是偶發 —— 它是這個專案的結構性風險:

> 規則寫在 `internal/game`,入口在 `use.go` / `commands.go` / `cmd/u5cht/main.go`。
> **兩邊各自看起來完整**,而中間斷掉不會有任何測試紅。

所以這一輪先寫一支掃描器 `tools/find_unwired.py`:列出
**「有定義、但整個非測試程式碼裡沒有任何 `.Name(` 呼叫」** 的方法。

第一次跑:**27 個**。多數是純查詢或只給測試用的存取器(合理),
其中**四個是真的機制**。

## 2. ★ 三塊寶石碎片:`UseGemShard` 一個呼叫者都沒有

`use.go` 原本:

```go
case item >= UseShardFirst && item <= UseShardLast:
	// 碎片的用法是「丟進聖火」,而那條路走的是聖火那一支(`docs/re/26`),不是這裡。
	s.Log(MsgShardOnlyAtFlame)
	return false
```

**這裡就是那條路。** 跳表 case 29..31 全部指向:

```asm
loc_1A9A3:
    mov eax, [ebp+var_C] ; sub eax, 1Dh ; push eax ; call sub_1A38C
```

而 `sub_1A38C` 是「舉起碎片 → 對照聖火座標 → 投進去 → 暗影君主的末日」整條
(226 行,`Gem Shard` / `Falsehood...` / `...and cast it into the Flame of ` /
`The doom of the Shadowlord ` / `Faulinei` / `Astaroth` / `Nosfentor`)。

引擎的 `UseGemShard` **早就把它逐條實作完了**,含「北邊那一格必須是**這一位**
暗影君主」那個容易漏的第二條件。它只是**沒有任何呼叫者**,而 U 這一格反而
主動拒絕。⇒ 一個註解裡的「那不是這裡」把已完成的功能鎖在外面。

### ★ 順帶記一個 IDA 陷阱

`sub_1A38C` 比對聖火座標時是:

```asm
mov eax, offset aNoNoticeableEf   ; "\nNo noticeable effect now!\n"
cmp dl, [eax+edi+1Ch]             ; ★ 拿字串的位址當座標表的基底
```

聖火座標表**緊接在那個字串後面**,而 IDA 把整段標成字串的一部分。
與 `docs/re/67` 的 `off_48A88`(其實是內嵌 `", "`)、`docs/re/77` 的
`aBoxHow`(其實只有 `"Box"`)同一類:**IDA 的標籤與型別都只是猜測**。

## 3. ★★ 風浪:`RoughSeas` 也沒有呼叫者

`docs/re/66` 上一輪把 `RoughSeas` 寫好了、也有測試 —— 但沒有人叫它,
所以**遊戲裡的風浪永遠不會發生**。

觸發點在 `sub_2D9D0`(移動之後的分派),與沼澤中毒同一支:

```asm
loc_2DCC5:
    sub_2D998() ; sub_2A50C()
    if (esi == 1) {                       ; ★ esi = 腳下的 tile,1 = 深水
        if ((byte_3E08C & 0FCh) == 28h)   ; 小艇
            → rough
        if ((byte_3E08C & 0FEh) == 14h)   ; 魔毯
            → rough
    }
rough:
    印 "Rough seas!" ; sub_2B1C8(X, Y) ; sub_22F0()
```

新增 `RoughSeasHere()` 掛在 `moveInWorld`,就在 `swampPoisonOnArrival()` 旁邊
—— 兩者都是「移動之後、由腳下 tile 決定」的同一族。

⚠ 這裡也看到 `sub_2A50C`(維生開銷)**在這一支也被呼叫一次**,
而 `sub_1318` 也叫它。兩處是否在同一回合都跑(⇒ 中毒每回合掉 2 血)未驗,
列成 ⬜ —— **不憑推測改掉現有的一處**。

## 4. ★ 狀態列少了四樣東西

`sub_2A1E8` 畫的:

```
F:<糧食>        永遠畫
Ship:<耐久>     ★ 只在船上而且不在戰鬥中才畫;否則那一格畫 G:<金幣>(sub_2A0C4)
<月>-<日>-<年>
<模式字母>      byte_3E08A 非 0 時畫('P' 防護 / 'N' 抗魔 / 'T' 停時 / 'Q' / 'C')
```

`sub_2A984` 另外畫**風向**(`<方位> Winds`),而它有三道閘門:

```asm
byte_3E0A3 >= 0x21  → 不畫(地牢與戰鬥)
byte_3E0A3 == 0x19  → 不畫(★ 亞拉臘號殘骸)
byte_3E0A5 >= 0x80  → 走另一個分支(地下世界)
```

引擎的右欄**一樣都沒畫**,而 `WindName()` 早就寫好了、同樣沒有呼叫者。
風向不是裝飾 —— **頂風走不動**,看不到風向等於在海上盲航。

⚠ **`Ship:` 與 `G:` 共用同一格**(原版 `jnz → sub_2A0C4` 是 if/else)。
所以在船上看不到金幣 —— 那不是遺漏,是版面只有一格。照做。

## 5. 順手刪掉真的死掉的兩支

- `equip(member, slot, item)` —— `Ready` 改寫成照原版的規則之後就沒人用了
  (連測試都沒有)。留著會讓下一輪以為「換裝有兩條路」。
- `DungeonTurnAround()` —— UI 走的是 `DungeonForward` / `DungeonTurn`
  (第一人稱那一套),這支是早期的便利包裝,0 呼叫者 0 測試。

★ **刪掉死程式碼不是整理,是修正一個假訊號**:一支「看起來實作了某機制」
的函式如果沒人叫,它在下一輪盤點時會被算成「已完成」。

## 6. 標準檢查

`tools/find_unwired.py` 留在庫裡。**每次宣稱某階段完成之前跑一次** ——
它抓的是「規則寫完了但接不到」,而那是本專案已經踩過五次的形狀
(卷軸、藥水、月石、碎片、風浪)。

第一次跑 27 個,處理後 22 個,其餘都是純查詢 / 測試專用存取器。

## 7. 還沒讀的

- ⬜ `sub_2A50C`(維生開銷)被 `sub_1318` 與 `sub_2D9D0` **兩處呼叫**。
  若同一回合都跑,中毒會掉 2 血而不是 1。**未驗,先不動。**
- ⬜ `sub_2A984` 的地下世界分支(`loc_2AA44`)畫什麼沒讀。
- ⬜ `sub_2A0C4` 的金幣是**右對齊到固定寬度**(五種格式字串),
  引擎目前只印數字,沒對齊。
- ⬜ `sub_2B1C8`(風浪時呼叫,參數是座標)在截斷清單上,未讀。
- ⬜ 剩下 22 個「沒有非測試呼叫者」的方法裡,`CreationPrompt` / `Format` /
  `Field` / `IsRanged` / `Stepping` / `HasLineOfFire` 六個連測試都沒有引用 ——
  逐一確認「是該接上的入口」還是「該刪的死碼」還沒做完。
