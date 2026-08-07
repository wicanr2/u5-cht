# WORKLIST — 遊戲機制還原盤點

> **目標(使用者指示 2026-08-07):遊戲機制要還原成原版。** 本檔逐子系統盤點「原版有什麼、
> 我們做到哪」,並標明每一項的**一手證據**。階段順序見 `PLAN.md`,規則見 `CLAUDE.md`。
> 最後更新:2026-08-07

## 0. 盤點方法(為什麼可以相信這份清單)

不是憑對 U5 的印象列的,而是從原版檔案本身推出來的三類證據:

1. **24 個 `.OVL` 的檔名與大小** —— 原版把每個子系統切成一個 overlay,檔名直接說明它負責什麼,
   大小則是該子系統的程式碼量(16-bit 機器碼),可當工作量指標。
2. **`DATA.OVL` 的 43 個明文表** —— 裝備、怪物、店名、法術、狀態、選單、劇情關鍵字全在裡面。
3. **資料檔本身** —— `KARMA.DAT`(業報)、`QUESTION.DAT`(創角)、`*.PTH`(NPC 路徑)、
   `*.CBT`(戰鬥地圖)…每個檔案的存在就是一項機制的存在證明。

⚠ 標「待確認」的項目,是**只知道它存在、還不知道規則細節**的 —— 那些要等 P3 讀反編譯碼
或手冊 OCR,**在有證據前不要照自己對 U5 的印象實作**。

## 1. 原版子系統總覽(依 OVL)

| OVL | 大小 | 子系統 | 狀態 |
|---|---|---|---|
| `DATA.OVL` | 48,464 | 資料表(非程式碼) | 🔶 已解出世界地圖索引、詞典、43 個明文表 |
| `SJOG.OVL` | 8,800 | ⬜ 用途未確認(檔名待解;是第二大的 OVL,值得優先逆) | ⬜ |
| `CAST.OVL` | 8,560 | 施法(主) | ⬜ |
| `INTRO.OVL` | 8,400 | 開場動畫與主選單 | ⬜ |
| `DUNGEON.OVL` | 8,016 | 地牢(第一人稱) | ⬜ |
| `CMDS.OVL` | 7,440 | 指令處理(玩家輸入的動詞) | ⬜ |
| `COMBAT.OVL` | 7,408 | 戰鬥 | ⬜ |
| `MAINOUT.OVL` | 7,344 | 地表主迴圈 | 🔶 地圖顯示與移動已有,規則未做 |
| `TOWN.OVL` | 6,256 | 城鎮場景 | ⬜ |
| `SHOPPES.OVL` | 5,936 | 商店(主) | ⬜ |
| `COMSUBS.OVL` | 5,216 | 戰鬥副程式 | ⬜ |
| `DNGLOOK.OVL` | 5,040 | 地牢視角繪製 | ⬜ |
| `NPC.OVL` | 4,912 | NPC 行為與排程 | ⬜ |
| `TALK.OVL` | 4,880 | 對話系統 | 🔶 `.TLK` 已可解碼,對話流程未做 |
| `ZSTATS.OVL` | 4,880 | 角色數值畫面(Ztats) | ⬜ |
| `LOOKOBJ.OVL` | 4,560 | 觀察物件 | ⬜ |
| `CAST2.OVL` | 4,544 | 施法(續) | ⬜ |
| `FONT.OVL` | 3,744 | 字型繪製 | ✅ 自己的實作(倚天 CJK + 原版 8×8) |
| `BLCKTHRN.OVL` | 3,184 | Blackthorn(暴君)相關事件 | ⬜ |
| `SHOPPES2/3.OVL` | 2,848 / 2,528 | 商店(續) | ⬜ |
| `ENDGAME.OVL` | 2,800 | 結局 | ⬜ |
| `OUTSUBS.OVL` | 2,464 | 地表副程式 | ⬜ |
| `FLAMES.OVL` | 32 | 火焰動畫(極小,可能只是資料) | ⬜ |

**程式碼量級**:`ULTIMA.EXE` 36 KB + 24 個 OVL 約 130 KB 的 16-bit 機器碼。
FM Towns 版反編譯出 **61,364 行 C / 1,225 函式** —— 那就是要還原的行為總量。

## 2. 機制逐項盤點

### 2.1 世界與移動

| 項目 | 原版證據 | 狀態 |
|---|---|---|
| 地表世界地圖 256×256 | `BRIT.DAT` 205 chunk + `DATA.OVL` 0x3886 索引表 | ✅ 完整組出 |
| 地底世界 | `UNDER.DAT` 65,536 B(256 chunk,不省略) | ⬜ 解碼未做 |
| NPC 與排程 | 四檔各 4,608 B = 8 地點 × 576 B(32 × 16 B 排程 + 32 生物編號 + 32 對話號碼) | ✅ **全解**(docs/re/04):排程 slot 選擇、生物編號→tile(+256)、對話號碼分派。全遊戲 325 個 NPC、46 個商人 |
| 存檔 | `SAVED.GAM`/`INIT.GAM` 各 4,192 B:16 筆 32 B 角色紀錄 + 時間 + 隊伍 + 位置 + 業報 | ✅ **已解**(docs/re/07):逐欄位循序讀寫,位移用累加算出並以「最大HP=等級×30」交叉驗證。遊戲已能由原版存檔開局 |
| 對話腳本 | `.TLK` 記錄是腳本:0x81–0x9F 是指令,由 `sub_1C3F8` 的 31 路跳表解譯 | ✅ **全解**(docs/re/06):指令集、`2i+5`/`2i+6` 關鍵字對與 **0x90 終止**、子字串詞首比對、34 個內建關鍵字(含髒話)、提問區塊與是非分支、入隊。135 段對話 1,307 個關鍵字全數可答 |
| 文字壓縮 | 118 字詞典(DATA.OVL 0x104C)+ 128 槽表(10 個空槽) | ✅ **全解**(docs/re/05):`.TLK` 與 `.DAT` 極性相反。SHOPPE.DAT 194 筆零殘留 token |
| 場景地圖 | 四檔各 16,384 B = 16 張 32×32(列寬由 `sub_86C` 證實) | ✅ **64 張全解出;每張屬於哪個地點/樓層已用梯子拓撲獨立驗證**(不重疊不留空)。畫面驗收:進不列顛城南門、爬上燈塔頂層燈室 |
| 地點類型 | `DATA.OVL` 0x2AB3 列 9 種;`sub_2D72C` 分派 12 種;**lighthouse 由四座燈塔的座標反推出是 tile 27** | ✅ tile 對應已知 |
| 進出場景(portal) | ✅ **已實作**(`internal/game`):E 進入、走到邊界問是否離開、離開回城門格、ARARAT 通地下世界。原版分派 `sub_2D72C` 依腳下 tile:16=hut、17=法典聖壇、18=keep、19=village、20=towne、21=castle、22=cave、23=mine、24=dungeon、25=八德聖壇、**27=lighthouse**、61=Blackthorn 宮殿、62=LB 城堡 | ✅ 規則全解 + 已接進引擎 |
| 地形通行規則 | **已從原版執行檔取得**:`byte_5FF6C` 64 B bitmap(阻擋 195/512)+ `sub_2A610`/`sub_2A674` 判定式,見 `docs/re/02` | ✅ 已接進引擎(走進海裡會顯示「受阻!」) |
| 移動成本分級 | `"Slow progress!"` / `"Very slow!"` @ `sub_2D0BC` | ⬜ 訊息已定位,規則未讀 |
| 移動者→模式表 | `byte_5FF8C[mover>>2]`,11 種模式(一般/水上/兩棲/陸上/方向性…) | ⬜ |
| 世界環繞(wrap) | 已實作並測試 | ✅ |
| 方向 | `DATA.OVL` 0x2C8C:`North East South West` + `Dir:` | ✅ 中文化(北/東/南/西) |

### 2.2 時間、天候、天象

| 項目 | 原版證據 | 狀態 |
|---|---|---|
| 時段 | `DATA.OVL` 0x7836:`morning / afternoon / evening` | ⬜ |
| 日夜與時鐘 | NPC 排程依賴它(`*.PTH`) | ⬜ |
| 月相與月門 | U5 沿用系列月門機制;`MISCMAPS.DAT` 疑為相關 | ⬜ 待確認 |
| **風向**(航海) | `DATA.OVL` 0x5564:`Calm / North / South / East / West Winds` | ⬜ 影響船速與方向,規則待逆 |

### 2.3 角色與隊伍

| 項目 | 原版證據 | 狀態 |
|---|---|---|
| 創角(吉普賽問答) | `QUESTION.DAT` 30 筆 + `CREATE.16` | ⬜ 文字已解出 |
| 屬性 STR/DEX/INT | `DATA.OVL` 0x33F1 | ⬜ |
| 等級與經驗 | `DATA.OVL` 0x3529:`Exp:` `Level:` | ⬜ |
| **從 Ultima IV 轉入角色** | `DATA.OVL` 0x3115 主選單有 `Transfer from Ultima IV`;0x3529 有轉換訊息 | ⬜ 系列特色,別漏 |
| 性別 | `DATA.OVL` 0x33F1:` Male / Female ` | ⬜ |
| Avatar 身分判定 | `DATA.OVL` 0x33F1:`is an Avatar.` / `not an Avatar` | ⬜ |
| 角色數值畫面(Ztats) | `ZSTATS.OVL` 4,880 B | ⬜ |
| 狀態 | `DATA.OVL` 0x0919:`Good Health / Poisoned / Dead / Asleep / Charmed` | ⬜ |
| 隊伍成員招募 | NPC `join` 關鍵字;`DATA.OVL` 0x9D80 有 `Malik Greyson Trian Jeremy` | ⬜ |

### 2.4 戰鬥

| 項目 | 原版證據 | 狀態 |
|---|---|---|
| 戰鬥地圖 | `BRIT.CBT` 5,632 B / `DUNGEON.CBT` 39,424 B | ⬜ 未解碼 |
| 戰鬥流程 | `COMBAT.OVL` 7,408 + `COMSUBS.OVL` 5,216 | ⬜ |
| 攻擊方位 | `DATA.OVL` 0x2DA2:`Attacked from the north/east/south/west` | ⬜ |
| 連擊次數 | `DATA.OVL` 0x9AA8:`Attack- two three four` | ⬜ |
| **怪物 22 種以上** | `DATA.OVL` 0x0380/0x0404/0x0469:`SEA HORSES SQUIDS SEA SERPENTS SHARKS GIANT RATS BATS SPIDERS GARGOYLE INSECTS ORCS SKELETONS SNAKES ETTINS HEADLESSES WISPS DAEMONS MONGBATS CORPSERS ROTWORMS SHADOW LORD` + `BLACKTHORN` `LORD BRITISH` | ⬜ 名單已有,數值待逆 |
| 怪物圖 | `MON0–7.16`(8 組) | ⬜ 壓縮未破 |

### 2.5 魔法

| 項目 | 原版證據 | 狀態 |
|---|---|---|
| 咒語(符文組合) | `DATA.OVL` 0x0919 符文詞 `AN BET CORP DES EX FLAM GRAV HUR IN KAL LOR MANI NOX POR QUAS REL SANCT TYM UUS`;0x06EE 法術名 `In Lor / Grav Por / An Zu / An Nox / Mani / An Ylem / An Sanct / An Xen Cor / Rel Hur / In Wis / Kal Xen / In Xen M…` | ⬜ 名單已有 |
| 法術代碼表 | `DATA.OVL` 0x09B3:98 項縮寫(`AY AS ACX HR IW KX IMX LV FV…`) | ⬜ 對應關係待解 |
| 材料(reagents) | `DATA.OVL` 0x06EE:`Pearl / Nightshade / Mandrake` … | ⬜ |
| 施法流程 | `CAST.OVL` 8,560 + `CAST2.OVL` 4,544 | ⬜ |
| **Words of Power**(地牢咒語) | `DATA.OVL` 0x44AD:`FALLAX VILIS INOPIA MALUM AVIDUS INFAMA IGNAVUS **VERAMOCOR**` | ⬜ U5 關鍵劇情機制 |

### 2.6 物品、裝備、經濟

| 項目 | 原版證據 | 狀態 |
|---|---|---|
| 裝備 75 項 | `DATA.OVL` 0x0052:`Leather Helm … Mystic Sword`、`Ring of Invisibility` | ⬜ 名單已有,數值待逆 |
| 特殊物品 158 項 | `DATA.OVL` 0x04C3:`Magic Crpt(魔毯) / Skull Keys / Amulet / Crown / Sceptre` … | ⬜ |
| 魔法飾品 | `DATA.OVL` 0x7CFE:`Invisibility / Protection / Regeneration Ring`、`Turning Amulet` | ⬜ |
| **商店 236 家(店名)** | `DATA.OVL` 0x0C40:`North Star Armoury / Buccaneers Booty / The Shattered Shield / Siege Crafters / The Honest Meal / The Wayfarer Tavern / The Sword and Keg / The Slaughtered Lamb` … | ⬜ |
| 商店對白 | `SHOPPE.DAT` 194 筆 | 🔶 已解碼,但 **862 個詞典 token 未展開** |
| 商店流程 | `SHOPPES.OVL` + `SHOPPES2/3.OVL` 共 11,312 B | ⬜ |

### 2.7 NPC 與對話

| 項目 | 原版證據 | 狀態 |
|---|---|---|
| NPC 對話文字 | `*.TLK` ×4 | ✅ 解碼(48 筆/檔,bit7 編碼);🔶 欄位切分未定 |
| NPC 定義 | `*.NPC` 各 4,608 B | ⬜ 未解碼 |
| NPC 排程與路徑 | `BRITISH.PTH` 2,783 B;`NPC.OVL` 4,912 B | ⬜ U5 招牌機制(NPC 有作息) |
| NPC 職業 | `DATA.OVL` 0x0342:`VILLAGER MERCHANT JESTER BARD PIRATES` | ⬜ |
| 對話關鍵字系統 | `TALK.OVL` 4,880 B;`.TLK` 內含控制碼(`\x01` 疑為玩家名代入) | ⬜ |
| 招牌 | `SIGNS.DAT` 8,364 B | ⬜ 格式為 offset 表 + 字元畫框 |
| 觀察(look) | `LOOK2.DAT` 3,622 B + `LOOKOBJ.OVL` 4,560 B | ⬜ 格式含 0x01–0x1F 控制碼 |

### 2.8 業報與美德

| 項目 | 原版證據 | 狀態 |
|---|---|---|
| 業報(Karma) | `KARMA.DAT` 761 B / 6 筆 | ⬜ 文字已解,規則待逆 |
| 聖壇與真言 | `MISCMSG.DAT`:`"What is the Mantra of the Mystic Shrine of` | ⬜ |
| 八德 | 系列共通;U5 的政治主題圍繞它 | ⬜ |

### 2.9 地牢

| 項目 | 原版證據 | 狀態 |
|---|---|---|
| 地牢地圖 | `DUNGEON.DAT` 4,096 B | ⬜ |
| 第一人稱視角 | `DUNGEON.OVL` 8,016 + `DNGLOOK.OVL` 5,040 | ⬜ |
| 地牢圖素 | `DNG1–3.16`(DOS,壓縮)/ `DNG1–3.PNL`(FM Towns,各 54,560 B) | ⬜ |
| 地牢戰鬥 | `DUNGEON.CBT` 39,424 B | ⬜ |

### 2.10 載具

| 項目 | 原版證據 | 狀態 |
|---|---|---|
| 載具動詞 | `DATA.OVL` 0x2956:`Ride / Fly / Row / Head` | ⬜ |
| 馬 | 同上 `Ride`;0x725C 有一組彩蛋名字(`Ferrari Lamborghini Lotus Porsche Horse`) | ⬜ |
| 船 + 風向 | `Row` + 0x5564 風向表;`Ship:` 出現在 0x54AD | ⬜ |
| 飛毯 | `Fly` + `Magic Crpt`(0x04C3) | ⬜ |
| 跳越(skiff?) | 待確認 | ⬜ |

### 2.11 劇情與結局

| 項目 | 原版證據 | 狀態 |
|---|---|---|
| 開場四章 | `DATA.OVL` 0xA020:`The Summoning / The Journey / The Arrival / The Welcoming`;`STORY.DAT` 20 筆 + `STORY1–6.16` | 🔶 文字已解 |
| **三位暗影君主** | `DATA.OVL` 0x484F:`doom of the Shadowlord **Faulinei / Astaroth / Nosfentor**` | ⬜ |
| **三個邪惡碎片** | `DATA.OVL` 0x47C3:`evil Shard of **Falsehood… / Hatred… / Cowardice…**` | ⬜ U5 核心:摧毀碎片才能滅暗影君主 |
| Blackthorn | `BLCKTHRN.OVL` 3,184 B;怪物表裡也有 `BLACKTHORN` | ⬜ |
| Lord British 歸來 | `ENDMSG.DAT` 11 筆(`Lord British says:`);`ENDGAME.OVL` | 🔶 文字已解 |
| 結局 | `END.DAT` / `END1–2.16` / `ENDSC.16` | ⬜ |

### 2.12 系統

| 項目 | 原版證據 | 狀態 |
|---|---|---|
| 主選單 | `DATA.OVL` 0x3115:`Journey Onward / Create New Character / Transfer from Ultima IV / Ultima V Introduction / Acknowledgements / Return to the View` | ⬜ |
| 存讀檔 | `SAVED.GAM` 4,192 B / `INIT.GAM`;`SAVED.OOL` | ⬜ 格式未解;可做「匯入原版存檔」 |
| 指令系統 | `CMDS.OVL` 7,440 B | ⬜ U5 是動詞驅動(Talk/Look/Get/Open/Board…) |
| 四種顯示模式 | `EGA.DRV / CGA.DRV / HER.DRV / T1K.DRV` | ⬜ 素材已有(`.16`/`.4`/`.HCS`) |
| 音樂 | upgrade 19 首 `.XMI`;FM Towns 15 首 `.EUP` + 2 條 CDDA;PC-98 `UL01–15.BIN`;**播放函式 `sub_3181C(曲目)`,曲目 0–14**;已知 castle=6、LB 城堡=7、village/hut=8、Blackthorn 宮殿=11(`docs/re/03`) | 🔶 場景→BGM 對應部分已知 |
| 音效 | FM Towns 25 個 `.SND` | ⬜ |

## 3. 目前完成度(誠實版)

**已完成的是「資料層 + 顯示層的一部分」,遊戲機制本身幾乎都還沒做。**

| 層 | 狀態 |
|---|---|
| 素材解碼 | 🔶 約 35%:tileset ✅、世界地圖 ✅、字型 ✅、對話 ✅(欄位未定)、明文文字 ✅;地底/場景/戰鬥圖/NPC/存檔/壓縮圖組 ⬜ |
| 顯示層 | 🔶 約 30%:地圖視窗 ✅、CJK 文字 ✅、HUD 骨架 ✅;場景/地牢/戰鬥/選單畫面 ⬜ |
| **遊戲機制** | ⬜ **約 45%**:地表走動、進出 32 個地點的場景、場景內移動與邊界離開、梯子與樓梯上下層(全遊戲 101 段梯子測過)。時間、NPC、戰鬥、載具仍未做 |
| 中文化 | 🔶 約 10%:管線 ✅、UI 字串 ✅;114 筆明文可翻但未翻、`.TLK` 未翻、`SHOPPE` 卡在 token |

## 4. 下一步優先序

依「先讓遊戲像遊戲」排,而不是依檔案大小:

1. ~~tile 通行規則~~ ✅ **已完成**(`docs/re/02`):不在 `DATA.OVL`(掃過 48 KB 沒命中),
   而在執行檔內的 64 B bitmap。用 `"Blocked!"` 字串當錨找到 `sub_86C` → `sub_2A694` → `sub_2A610`。
2. ~~場景地圖解碼~~ ✅(64 張 32×32);~~地點表~~ ✅(32 筆,`docs/re/03 §5`)。
   ~~地點 → 場景地圖~~ ✅  ~~進場景的狀態切換~~ ✅  ~~`.NPC` 與排程~~ ✅
   ~~文字壓縮詞典~~ ✅(全遊戲英文文字現在都讀得出來了)。
   ~~對話腳本引擎~~ ✅(docs/re/06:指令集、關鍵字比對、別名鏈;玩家可打字問話)。
   **剩下(依序)**:
   ~~存檔格式與名冊~~ ✅(docs/re/07:4,192 B 版面、16 筆角色紀錄;遊戲直接由原版存檔開局)。
   ~~加入隊伍~~ ✅  ~~反問玩家的問答~~ ✅(docs/re/06:提問區塊、是非分支、終端區塊)。
   1. **商店交易** —— `SHOPPE.DAT` 內容已可讀,缺營業時間表(`sub_9C7C` 已解)與交易流程。
   2. **NPC 每回合移動** —— 現在是直接放到排程位置;原版 `sub_9690` 會逐步尋路。
   3. **叫衛兵**(opcode 0x8B → `sub_C10`)—— 需要戰鬥系統當落點。
   4. 戰鬥、魔法、地牢、載具、存檔寫回。
3. **`.NPC` 格式 + `TALK.OVL` 對話流程** → 能跟人說話(對話文字已解好)。
4. **`WORRIORJ ⊖ WORRIORS` diff** → DBCS 排版先例(中文化)+ 順帶熟悉反編譯碼結構。
5. **存檔格式** → 能存讀進度。
6. 之後才是戰鬥、魔法、地牢、載具。

> 紀律提醒:每個機制**先找原版證據再實作**(`CLAUDE.md §3.0` 素材鐵則同樣適用於規則 ——
> 數值與規則忠於原版,不自行平衡)。宣稱任一項「完成」前,對 DOSBox 原版實測比對
> (`rulebook/65`:自家測試綠不算)。
