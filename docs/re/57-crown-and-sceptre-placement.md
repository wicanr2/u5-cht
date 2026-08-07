# 57 — 王冠與權杖擺在哪裡:一個號碼、兩套索引

| | |
|---|---|
| 輸入檔 | `U5_E/WORRIORS.EXP`(FM Towns,SHA-256 見 `docs/re/00-hexrays-p3-verified.md`) |
| | `CASTLE.NPC` / `KEEP.NPC`(DOS 版,各 4,608 B) |
| | `LOOK2.DAT`(DOS 版,3,622 B) |
| 位址 | `sub_154BC`(Get 分派)`loc_15863` / `loc_158A2` / `loc_158D0` / `loc_158DE` |
| | `sub_10B3C`(地下世界放置)、`sub_1E74`(NPC 鏡射進物件表) |
| 落地 | `internal/u5data/get.go`、`internal/u5data/regalia_test.go`、`internal/game/regalia_test.go` |

`WORKLIST` 上「王冠與權杖擺在哪裡未找」這一條掛了很久。護符與三塊碎片早就結案了 ——
`sub_10B3C` 進地下世界時當場塞進物件槽,座標寫死在程式裡。王冠與權杖用同樣的方法卻怎麼都找不到:

- 執行檔裡沒有 `mov al, 0B5h` 這類寫入,只有 Get 分派裡的 `cmp`。
- `INIT.GAM` / `SAVED.GAM`(各 4,192 B)整份沒有 0xB5 / 0xB6 這兩個位元組。
- 四份 `.OOL` 物件表(各 256 B)也沒有。

## 卡住的原因不是資料藏得深,是我在錯的命名空間裡找

去掃地圖檔倒是撈到一堆:`CASTLE.DAT` 有 4 個 0xB5、6 個 0xB6,`KEEP.DAT` 有 2 個與 8 個。
但位置一看就不對 —— **清一色左右對稱**:

```
不列顛王城堡 +2 層(CASTLE.DAT 場景 3)
 y= 1   50 45 45 B4 88 45 50 …            … 50 45 88 B4 45 45 50
 y= 3   56 B7 45 45 45 …                  … 45 45 45 45 B5 56
 y=29   50 45 45 B6 88 45 50 …            … 50 45 88 B6 45 45 50
```

全世界只有一個的信物不會成對出現,更不會在兩個樓層各擺一份。**對稱本身就是反證**,
而我當時看到了卻沒把它當線索,還繼續往「這是不是某種展示櫃」的方向想。

答案在 `LOOK2.DAT`。它有兩套索引空間(`sub_D45C` 的 `+512` 位元組,
落地在 `u5data.LookObjectBase = 256`):

| 號碼 | 當地形看 | 當物件種類看 |
|---|---|---|
| 0xB4 | look#180 = `a cannon` | look#436 = `a Shard of the Gem of Mondain!` |
| 0xB5 | look#181 = `a cannon` | look#437 = `the Crown!` |
| 0xB6 | look#182 = `a cannon` | look#438 = `the Sceptre!` |
| 0xB7 | look#183 = `a cannon` | look#439 = `the Amulet!` |

地圖裡那些是**四個朝向的加農砲**。對稱、成對、出現在城堡上層與亞拉拉特 / 邊境哨 /
巨蛇要塞這幾座堡壘 —— 全部說得通了。

Get 分派拿到的是**物件種類**,不是地形 tile:

```asm
loc_158A2:  mov  byte_3DFC0, 0FFh
            push offset aTheCrownOfLord   ; "The Crown of Lord British!\n"
loc_158D0:  mov  byte_3DFC1, 0FFh
            push offset aTheSceptreOfLo   ; "The Sceptre of Lord British!\n"
```

## 找對命名空間之後,答案是唯一的

物件種類的來源除了程式硬塞,還有一條:`sub_1E74` 每回合把「此刻在本層」的 NPC
鏡射進物件表(`docs/re/36`),而 `.NPC` 的生物編號欄就是物件種類。掃四份 `.NPC`
共 32 個地點 × 32 槽:

| 信物 | 檔案 | 地點 | 槽 | 場景座標 | 樓層 | 行為型別 | 四個時刻 | 對話 |
|---|---|---|---|---|---|---|---|---|
| 王冠 0xB5 | `CASTLE.NPC` | 18(第二座城堡,大地圖 (196,245)) | 1 | (15,13) | +3 | 0 | 全 0 | 0 |
| 權杖 0xB6 | `KEEP.NPC` | 29 `STONEGATE` | 9 | (15,15) | 0 | 0 | 全 0 | 0 |

**全遊戲就這兩槽**(`TestRegaliaSitWhereTheNPCFilesSayTheyDo` 連「只有兩槽」一起釘住)。
排程三個 slot 指同一格、行為型別 0(`NPCAIFixed`,跳表 default,什麼都不做)、
四個時刻全 0 → `Scheduled()` 為假,所以暗影君主的影響與「叫人滾開」都不會挑到它。

兩處都是密室,周圍的地形用 `LOOK2.DAT` 讀出來是:

```
王冠(城堡 +3 層 (15,13))          權杖(STONEGATE 地面層 (15,15))
  28 28 28 28 28 28 28              4F 44 44 44 44 44 44 44 4F
  28 4F 4F 4F 4F 4F 28              44 44 44 46 44 46 44 44 44
  28 4F 45 45 45 4F 28              44 44 46 8C 8C 8C 46 44 44
  28 4F 45[45]45 4F 28              44 44 46 8C[44]8C 46 44 44
  28 4F B2 45 45 B2 4F              44 44 46 8C 8C 8C 46 44 44
  28 4F 4F 45 45 4F 28              44 44 44 46 46 46 44 44 44
  27 27 4F 97 4F 27 27              4F 44 44 44 44 44 44 44 4F

0x28 屋頂 / 0x4F 雉堞 / 0x45 石板     0x44 石板 / 0x46 石柱
0xB2 火盆 / 0x97「怪門」              0x8C「鬆動的磚」×8 團團圍住
```

## 引擎這邊一行都不用加

因為它們是 NPC,而 NPC 鏡射 + Get 掃物件表這條路早就在跑。
`internal/game/regalia_test.go` 用原版資料進場景、站到北邊一格、按 G ——
兩件都撿得到,而且**引擎沒有為信物加過任何特例**。

## ⚠ 兩者的善後不一樣(照抄,不要「修好」)

```asm
loc_158A2:  … call sub_2E0 / call sub_218 / call sub_2E0 / call sub_268   ; 王冠:全套
loc_158D0:  … jmp short loc_158EA → loc_15903                            ; 權杖:只有共同尾段
```

共同尾段 `loc_15903` 呼叫 `sub_2B6C8` 清掉物件槽,但**只有王冠配了 `sub_218`**
(寫永久移除位元)。所以:

- 王冠拿走就不見了。
- **權杖離開 `STONEGATE` 再回來會躺在原地第二次** —— 可以刷。
  與不列顛王城堡二樓那張魔毯(`CarpetNPCSlot`)同一種原版行為。

`TestOnlyTheCrownIsRemovedForGood` 把這個不對稱釘住:哪天有人「順手修好」權杖,
測試會紅,而紅的原因是它不再照抄原版。

## 順帶更正一條斷言

`internal/u5data/get.go` 原本寫「全遊戲撿得起來的物品型 NPC 只有這兩個加王冠」——
數字對(檀香木盒、魔毯、王冠),但漏了權杖,應該是**四個**。已改。
