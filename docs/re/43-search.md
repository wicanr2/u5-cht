# Search 指令

輸入檔:`WORRIORS.EXP`(FM Towns 英文版)、`gamedata/LOOK2.DAT`
IDA 位址:`sub_147A8`(主體)、`sub_13F04`(陷阱偵測)、`sub_13DD8`(翻到什麼)
日期:2026-08-08

---

## 1. 三條出路

```
問方向 → 挑一個人 → 看目標格

那一格有物件(寶箱之類)  → sub_13F04   查陷阱
那一格是 0x4E(有缺口的牆)→ 開出密門
其餘家具                  → sub_13DD8   擲「翻到了什麼」
```

## 2. 敘述句是拼出來的

原版的字串長這樣:

```
"\nIn the stump\nt"  +  "hou dost find\n"
```

地點語**以 `t` 結尾**,接上 `hou dost find` 才成句。所以地點語不是額外的一行,
它是同一句話的前半 —— 拆成兩行印的話畫面會多一行空隙,對照原版截圖會對不上。

### 15 種家具的地點語

每個 tile 都拿 `LOOK2.DAT` 的敘述交叉對過(`docs/re/37`),不是憑順序猜的:

| tile | `LOOK2` 說它是 | 地點語 |
|---|---|---|
| 0x2B | a hollow stump | In the stump |
| 0x4F | a wall | In the wall |
| 0x5A | a window shelf | On the shelf |
| 0x5C / 0x5D | a crowded bookshelf | In the bookshelf |
| 0xA1 | a deep well | Near the well |
| 0xA5 | a desk | In the desk |
| 0xA6 | an oaken barrel | In the barrel |
| 0xA8 | a vanity | In the vanity |
| 0xAB / 0xAC | a bed | Under the bed |
| 0xAD | a chest of drawers | In the dresser |
| 0xAF | a heavy footlocker | In the trunk |
| 0xB2 | a hot brazier | In the brazier |
| 0xBC | a fireplace | In the fireplace |

⇒ 「chest of drawers → dresser」「heavy footlocker → trunk」這兩組同義詞
就是交叉驗證有效的證據:光看號碼不可能對上,是敘述表補完了語意。

## 3. 密門藏在「有缺口的牆」

```asm
cmp     [ebp+var_8], 4Eh ; 'N'
jnz     short loc_14A45
push    offset aAHiddenDoor_0 ; "a hidden door!\n"
```

`LOOK2[0x4E]` = **「a wall with a nick」** —— 那個缺口就是給玩家的視覺提示;
`LOOK2[0x4F]` 是普通的「a wall」,搜了只會翻到垃圾。

★ **兩格差一號,行為完全不同。** 搞混的話密門變成隨處都有、或永遠找不到。

## 4. 陷阱偵測會出兩種錯(`sub_13F04`)

```
q = 物件的品質位元組
有陷阱 = q & 0x80
等級   = q & 0x7F

難度 = 30 − 智力        (沒有陷阱旗標時)
難度 = 等級 + 30 − 智力  (有陷阱時)
看清 = (難度 / 2) ≤ random(1, 30)

看清 ≠ 有陷阱  → 「no trap!」
否則:
    看清 且 等級 < 10 → 「a simple trap!」
    看清 且 等級 > 20 → 「a complex trap!」
    其餘              → 「a trap!」
```

真值表:

| 看清 | 有陷阱 | 印出 | |
|---|---|---|---|
| ✔ | ✔ | 依等級描述 | 正確 |
| ✔ | ✘ | no trap! | 正確 |
| ✘ | ✔ | **no trap!** | ← 漏看了 |
| ✘ | ✘ | **a trap!** | ← 幻覺 |

★ **低智力的角色既會漏看真陷阱、也會看到不存在的陷阱。** 兩種錯都要照做 ——
只做「漏看」的話,玩家會學到「說沒陷阱就一定安全」,而原版不是。

## 5. 翻到什麼(`sub_13DD8`)

```
if random(0, 7) != 0:                    ← 八分之七
    清掉那個物件槽
    if random(0, 31) == 19:              ← 三十二分之一
        "Plague!"  該員狀態 → 'P'(中毒)
    else:
        switch random(0, random(0, 3)):   ← **嵌套的亂數**
            0 "nothing"  1 "worms"  2 "guts"  3 "a bloody pulp"
else:                                    ← 八分之一,真的翻到東西
    if random(0, 3) == 0 → "food!"  物件碼 0x0F
    否則                 → "gold!"  物件碼 0x02
    數量 = random(1, 3)
```

兩個容易寫錯的:

- **八分之七是垃圾。** 把機率調高會讓搜家具變成刷錢手段;原版刻意讓它不划算。
- **`random(0, random(0,3))` 是嵌套的**,不是 `random(0,3)`。攤平的話
  「什麼也沒有」與「血肉模糊的一團」機率相同,而原版是前者遠多於後者。

## 6. 引擎對應

| 原版 | 引擎 |
|---|---|
| `sub_147A8` | `game.State.Search` / `searchAt` |
| `sub_13F04` | `game.State.reportTrap` |
| `sub_13DD8` | `game.State.rollSearchFind` / `rollSearchJunk` |
| 15 種地點語 | `game.searchPhrase` |
| 0x4E | `game.SearchSecretDoor` |

## 7. 一條測試自己踩的坑

`TestTrapDetectionMakesBothKindsOfMistake` 第一版寫成「智力 0 就一定漏看」。
不對:難度只有**一半**要壓過 `random(1,30)`,等級 25 + 智力 0 的難度是 55、
一半 27,擲到 27..30 還是看清了(4/30)。單次斷言於是隨種子飄 ——
而它「通過」的那幾次完全看不出前提是錯的。

改成統計性:跑四百次,只斷言**兩種錯都出現過**。要驗的本來就是
「兩種錯都存在」,不是「某一次一定錯」。
