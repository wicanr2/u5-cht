# Look 指令、`LOOK2.DAT`、`SIGNS.DAT`

輸入檔:`WORRIORS.EXP`(FM Towns 英文版)、`gamedata/LOOK2.DAT`、`gamedata/SIGNS.DAT`、`gamedata/RUNES.CH`
IDA 位址:`sub_D9C4`(指令)、`sub_D258`(地形)、`sub_CC44` / `sub_D45C`(敘述表)、
`sub_D650` / `sub_D544`(招牌)、`sub_CE78`(噴泉)、`sub_D064`(天空)
日期:2026-08-08

---

## 0. 先講被推翻的兩條

`docs/formats/02-text-files.md` 原本記載:

| 原記載 | 實際 |
|---|---|
| `LOOK2.DAT` = 「218 個 NUL + 大量 0x01–0x1F 控制碼,結構化格式,未解」 | **u16 × 512 的位移表 + NUL 字串**。那些「控制碼」是位移的高位位元組(資料落在 0x0400–0x0E17,高位元組自然是 0x04–0x0E) |
| `SIGNS.DAT` = 「同 u4-cht 的美術內嵌字母問題,先評估再決定譯不譯」 | **框是美術、字是字**。而且框的美術**在字型裡**(`RUNES.CH`),不是點陣圖 → 文字整段可譯 |

兩條的錯法一樣:**先數位元組分布、再從分布猜格式**,而沒有先試最單純的假設 ——
「檔頭是不是一張表」。位元組分布對「是不是表」這個問題幾乎沒有鑑別力,
因為位移的高位元組看起來就跟控制碼一模一樣。

---

## 1. `LOOK2.DAT` —— Look 的敘述表

```
0x0000  u16 × 512   絕對檔案位移
0x0400  …           NUL 結尾的 ASCII,重複的敘述共用一段(512 格只有 216 段不重複)
```

表長 1024 B,第 0 筆位移正好是 1024 —— 表尾接資料頭。

### 兩個索引空間

```
sub_CC44(t)  →  *(u16*)(base + 2*t)          地形
sub_D45C(t)  →  *(u16*)(base + 2*t + 512)    物件(= 索引 + 256)
```

這解釋了一個乍看矛盾的現象:`TileHorse = 0x10`,但 `表[0x10]` 是 `a small hut`。
馬是**物件**,要查 `表[0x110]`,那裡才是 `a horse`。

### 交叉驗證(十一點,不是單點)

拿**與 `LOOK2.DAT` 完全無關**的來源推出來的 tile 常數對照 ——
那些常數來自移動判定、買馬、上下載具:

| 常數(出處) | 值 | `LOOK2` |
|---|---|---|
| `mountableTiles`(`sub_118CC` 的 `cmp 5`) | 5 | grass |
| `mountableTiles`(`cmp 'D'` / `'E'`) | 68 / 69 | cobble / cobble |
| `TileCodex` | 0x11 | the Shrine of the Codex |
| `TileShrine` / `TileShrineDesecrated` | 0x19 / 0x1A | a mystic shrine / a ruined shrine |
| `TileDoorA` / `TileLockedDoor` | 0xB8 / 0xB9 | a wooden door / a locked door |
| `TileStairsDown` / `TileStairsUp` | 0xC8 / 0xC9 | a ladder / a ladder |
| `TileHorse`(物件空間) | 0x10 | 0x110 = a horse |
| `TileCarpetObj`(物件空間) | 0x1B | 0x11B = an odd rug |
| `VehicleSkiff`(物件空間) | 0x28 | 0x128 = a skiff |

### 結尾的空白是接縫

```
表[0xDE] = "the Flame of "                            + 火名
表[0xDF] = "the collapsed entrance to the dungeon "   + 地牢名
表[0xFA] = "a grandfather clock, showing: "           + 時刻
```

三筆都以空格結尾,原版直接接後綴不另外補空白。trim 掉就會變成
`the Flame ofTruth`。

### 佔位符標出「被程式特判的格子」

被特判的 tile 在表上一律是 `*`:水晶球 0x29 是例外(它有字但走另一條路)、
天空 0x59、噴泉 0xD8–0xDB、招牌 0x89 / 0x8A / 0xA0 / 0xA4 / 0xF8。
這給了一個現成的**反向檢查**:哪天有一格從佔位符變成真的敘述,
就代表特判清單抄漏了一種。

---

## 2. Look 指令的分派

`sub_D9C4`:

```
問方向 → (x,y) = 隊伍位置 + 方向
tile = 地圖那一層;obj = 物件那一層

tile == 0x29(水晶球) → 凝視,不印「Thou dost see」
印 "\nThou dost see\n"
有物件            → 表[obj + 256],**到此為止**（物件會蓋掉招牌）
tile 是招牌類     → SIGNS.DAT
其餘              → sub_D258
```

`sub_D258(tile, x, y)`:

```
while tile ∈ {0xE0, 0xE1, 0xE2}:        ← 轉向格,會一路跟下去
    0xE0 → y--   0xE1 → x++   0xE2 → x--
    tile = 地圖[x][y]

tile == 0x59            → sub_D064   抬頭看天
tile == 0xA1            → sub_CD28   井
tile & 0xFC == 0xD8     → sub_CE78   噴泉
其餘 → 印 表[tile],然後:
    tile & 0xFE == 0xFA → 接時刻     hour%12(0→12):分鐘補零 + " AM."/" PM."(hour ≤ 11 是 AM)
    tile == 0xDE        → 接火名     依**地點**:0x1E Truth / 0x1F Love / 0x20 Courage
    tile == 0xDF        → 接地牢名   依**x 座標**(見下)
```

### 地牢名靠 x 座標分辨

`loc_D3BB` 那一段是對 `esi`(= x)做 switch,不是查地點表:

| x | 地牢 | x | 地牢 |
|---|---|---|---|
| 0x3A | Shame | 0x80 | Doom |
| 0x48 | Destard | 0x9C | Covetous |
| 0x5B | Despise | 0xEF | Hythloth |
| 0x7E | Wrong | 0xF0 | Deceit |

八座地牢的 x 剛好兩兩不同,一個 switch 就夠了。現代看像 hack,但那就是判斷依據。

---

## 3. `SIGNS.DAT` —— 招牌與墓碑

```
0x0000  u16 × 33   每個地點一筆位移(0 = 沒有招牌);33 = 地表 + 32 個地點
0x0042  …          記錄:[地點][樓層][x][y] + 內容 + NUL,表尾 0xFF
```

`sub_D650` 從該地點的位移起往下逐筆比對 —— **只比樓層與 x、y,不比地點編號**。
掃描會越過本地點的記錄繼續往下。座標相同的話會竄到別的地點的招牌;
那是原版行為,引擎照做(`SignSet.At`)。

### 兩格共用一塊招牌

內容只有 `0x0A` 的記錄是別名。渲染器 `sub_D544` 看到 0x0A 就 `edi += 6`,
而那 6 個位元組是 `0x0A` + NUL + 下一筆的 4 B 表頭 —— 跳完落在共用的內容上:

```
1916:  01 00 0e 14 | 0a 00 | 01 00 10 14 | "abbb…"
       └ x=0x0E ┘           └ x=0x10 ┘     └ 共用 ┘
```

掃描迴圈不認 0x0A,它只是「跳到 NUL 之後」,於是自然停在第二個表頭。兩條路同源。

### 內容的編碼

| 位元組 | 意思 |
|---|---|
| bit7 **set** | 一般字 |
| bit7 clear | **反白** |
| 0x29–0x31 | 巨集:一個位元組展開成 16 個字的整列(`dword_54E2C[c*4]`) |
| 0x26 `&` / 0x27 `'` | 都印成 `l` |
| 0x0D | 分頁,等按鍵 |
| 0x8A | 換行(印出 `c & 0x7F` = 0x0A,輸出層換行) |
| 其餘 | 印 `c & 0x7F` |

九個巨集:

```
0x29 "g              g"   0x2A "jlllllllnllllllk"   0x2B "8lllllllmllllll9"
0x2C "jllllllllllllllk"   0x2D "8llllllllllllll9"   0x2E "hllllk    jlllli"
0x2F "jlllli    hllllk"   0x30 "     g    g     "   0x31 "     hlllli     "
```

每條都是 16 個字 —— 招牌就是 16 欄寬。

### 招牌字型是 `RUNES.CH`(逐字模驗證)

把 `RUNES.CH` 當 8×8 直索引 dump:

```
'l' 0x6C  ── 一條橫線        'g' 0x67  │ 一條直線
'[' 0x5B  符文 TH            'A' 0x41  符文 A
0x0E      ✳ 星芒裝飾          '&' 0x26  **字模全空**
```

最後一條是決定性的:**`&` 的字模是空的,所以程式才要把它特判成印 `l`**。
先有空字模,才有那個分支 —— 這解釋了一條孤立看毫無道理的程式碼。

相對地 `IBM.CH` 的 0x6C 是普通的小寫 `l`、0x61 是 `a`。拿它畫招牌只會得到一堆字母。

⇒ **「美術內嵌字母」的印象對了一半**:框確實是字母當美術,但它在字型裡,
框中的文字是真的文字,譯得動。

### 符文合字

`[` = TH、`\` = EE、`]` = NG、`^` = EA、`_` = ST、`@` = 空白。

`NOR[@BRITAIN` = NORTH BRITAIN、`^_@PAWS` = EAST PAWS、`D\P` = DEEP、`TRYI]` = TRYING。

### 查不到招牌時的預設

`sub_D544(-1)` 印的是寫死在執行檔裡的告示板:

```
abbbbbbbbbbbbbbc
g              g
g  LIVE@BY@[E  g
g  EIGHT@LAWS  g
g              g
deeeeeeeeeeeeeef
```

它是**預設值不是錯誤**,所以引擎也照樣印。

---

## 4. 噴泉 —— 什麼也不做

`sub_CE78` 全文:印「a gurgling fountain!」「Who will drink?」,選人,
狀態是 `'D'` 或 `'S'` 印「Incapacitated!」,否則印「Refreshing...」。

**組語裡沒有任何寫入。** 沒有補血、沒有解毒、沒有加值。看起來像沒寫完,
但那就是原版 —— 不要「順手」補上療效(CLAUDE.md §3.0)。

## 5. 水晶球 —— 智力擲骰

`sub_D9C4` 的 `tile == 0x29` 分支:

```
member = sub_E19C()            ← 誰來看
roll   = random(1, 30)
member 的智力 > roll  →  "Strange vision!" + sub_EDD4  (= In Quas Wis 的全景)
否則                  →  "Death vision!"   + sub_2A464(member, 1)  扣 1 點 HP
```

`byte_3DDC2` − `byte_3DDB4`(名字)= 14 = 角色記錄的 `CharIntel`。
`word_3DDC4` − `byte_3DDB4` = 16 = 目前 HP。

## 6. 抬頭看天

`sub_D064`:

```
6 ≤ hour < 18  →  "the sun!"      而且 **扣 1 點 HP**(直視太陽)
否則           →  清 11×11 覆蓋層、灑 80 顆隨機星、把三個天體畫到八個方位之一、
                  印 "the night sky! " 後等按鍵
```

### 未解:白天那條路的一行

```asm
cmp     byte_3E08B, 0FFh
jnz     short loc_D0A3
call    sub_2B67C
and     eax, eax
jnz     short loc_D0A3
mov     al, byte ptr word_3E086     ← 把**隊伍的 x 座標低位元組**寫進 byte_3E08B
mov     byte_3E08B, al
```

`byte_3E08B` 是「單人狀態下是哪一位」,而 `word_3E086` 是 x 座標。
把座標寫進成員索引看起來像原版的 bug,但 `sub_2B67C` 的語意還沒追,
**不排除是我讀錯**。引擎目前沒有單人狀態,所以這一段先不實作 ——
留白比猜著寫好。要收掉這條得先逆 `sub_2B67C` 與 `byte_3E08B` 的全部寫入點。

## 7. 引擎對應

| 原版 | 引擎 |
|---|---|
| `sub_CC44` / `sub_D45C` | `u5data.LookTable.Terrain` / `.Object` |
| `sub_D650` / `sub_D544` | `u5data.SignSet.At` / `Sign.Render` / `Sign.Lines` |
| `sub_D9C4` | `game.State.Look` / `lookAt` |
| `sub_D258` | `game.State.lookTerrain` |
| `sub_CE78` | `game.State.drinkFromFountain` |
| `sub_D064` | `game.State.lookAtTheSky` |
| `sub_E19C` | `game.State.pickCharacter`(多人時的選單還沒接,見該處註解) |

譯文走 `internal/i18n`(`look#<索引>` / `looksuf#<英文>` / `sign#<地點>#<x>#<y>#<列>`)。
