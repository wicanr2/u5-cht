# RE-00b:函式與全域索引(自動產生)

> `python3 tools/gen_func_index.py > docs/re/00-function-index.md` 重新產生。
> **讀任何 `sub_XXXX` 之前先查這裡** —— 筆記超過二三十份後,憑記憶一定會重讀已解過的東西。
>
> 目前收錄 **1105** 個符號,來源是 `docs/` 下的逆向筆記。

| 符號 | 已知語意(取自筆記) | 出處 |
|---|---|---|
| `sub_1DC` | 主要函式  /  `sub_1E74`(配物件槽)、`sub_268`(放掉)、`sub_218`(記永久移除)、`sub_2E0`(反查)、`sub_1DC`(在不在場)  / | `36-sandalwood-box-npc-objects.md` |
| `sub_218` | 主要函式  /  `sub_172C4`(F 開砲,**244 行**)、`sub_2B3DC`(查某一格上的物件)、`sub_2A4D0`(全隊傷害)、`sub_2E0`/`sub_218`/`sub_268`(NPC… | `29-npc-behaviour-and-arrest.md`, `33-get-command.md`, `36-sandalwood-box-npc-objects.md`, `57-crown-and-sceptre-placement.md`, `69-the-overworld-cannon-did-nothing.md` |
| `sub_268` | 主要函式  /  `sub_172C4`(F 開砲,**244 行**)、`sub_2B3DC`(查某一格上的物件)、`sub_2A4D0`(全隊傷害)、`sub_2E0`/`sub_218`/`sub_268`(NPC… | `29-npc-behaviour-and-arrest.md`, `32-guard-challenge.md`, `33-get-command.md`, `36-sandalwood-box-npc-objects.md`, `57-crown-and-sceptre-placement.md`, `69-the-overworld-cannon-did-nothing.md` |
| `sub_2E0` | 主要函式  /  `sub_172C4`(F 開砲,**244 行**)、`sub_2B3DC`(查某一格上的物件)、`sub_2A4D0`(全隊傷害)、`sub_2E0`/`sub_218`/`sub_268`(NPC… | `33-get-command.md`, `36-sandalwood-box-npc-objects.md`, `57-crown-and-sceptre-placement.md`, `69-the-overworld-cannon-did-nothing.md` |
| `sub_324` | 場景載入時還會依時間切換:`if (hour < 5  /  /  hour > 19) sub_324()` —— 夜間的燈光處理。 | `04-npc-schedule-and-clock.md`, `50-hole-up-camp-sleep-repair.md` |
| `sub_48C` | sub_48C();                      // byte_3E16A = 盤據這裡的是第幾位;並讓牠現身 | `11-map-objects.md`, `28-shadowlords-and-blackthorn.md`, `29-npc-behaviour-and-arrest.md` |
| `sub_5C8` | sub_5C8(0);  sub_48C();                        // 重載地圖 | `03-scene-entry-and-tile-semantics.md`, `11-map-objects.md`, `29-npc-behaviour-and-arrest.md` |
| `sub_758` | 樓層增減(`sub_758`)  /  `byte_3E0A5 = 1` / `= -1`  /  `inc` / `dec byte_3E0A5`  /  照抄的話只能在 1F 與 B1F 之間跳  / | `03-scene-entry-and-tile-semantics.md` |
| `sub_7C0` | IDA 位址:`sub_2ACF4` case 32(空白鍵)、`sub_7C0`(載具動詞與朝向)、 | `54-pass-vehicle-verbs-and-the-avatar-line.md`, `66-hexrays-truncation-audit.md`, `68-troll-bridge-and-collision.md` |
| `sub_86C` | 1  /  `sub_31DC0` default、`sub_DF84`(月門傳送)、`sub_3FE4`(離開地牢)、`sub_86C`(方向樓梯)、`sub_1DA10`、`sub_177AC`、`sub_3FE4`… | `02-movement-and-tile-flags.md`, `03-scene-entry-and-tile-semantics.md`, `87-bgm-selection-and-the-tbl-tables.md`, `92-white-noise-shim-and-ambient-sound.md`, `98-select-menu-and-song-melody-matching.md` |
| `sub_B44` | `sub_8F3C`(距離)、`sub_195C`(接觸)、`sub_C10`(叫衛兵)、`sub_B44` / `sub_B98` | `28-shadowlords-and-blackthorn.md`, `29-npc-behaviour-and-arrest.md` |
| `sub_B98` | 主要函式  /  `sub_1B3D0`(盤查)、`sub_1B52C`(對話分派)、`sub_195C`(碰到 NPC)、`sub_1884`(逮捕)、`sub_1B140`(密語比對)、`sub_154C` → `s… | `04-npc-schedule-and-clock.md`, `09-items-and-creatures.md`, `28-shadowlords-and-blackthorn.md`, `29-npc-behaviour-and-arrest.md`, `32-guard-challenge.md` |
| `sub_C10` | `sub_8F3C`(距離)、`sub_195C`(接觸)、`sub_C10`(叫衛兵)、`sub_B44` / `sub_B98` | `06-conversation-script.md`, `14-combat-maps.md`, `29-npc-behaviour-and-arrest.md`, `70-hunger-poison-and-the-vanishing-rings.md` |
| `sub_C74` | if (生物編號 >= 0x40) { 印 "Attacked!"; sub_C74(npc); } | `29-npc-behaviour-and-arrest.md`, `32-guard-challenge.md` |
| `sub_CAC` | `sub_1EFC8`(清單瀏覽器)、`sub_CAC` / `sub_4074` / `sub_2D478`(Attack 三支)  / | `49-command-table-and-two-empty-keys.md`, `60-command-echo-and-menu-keys.md`, `90-cdda-scenes-and-sound-effect-indices.md` |
| `sub_EA0` | 梯子  /  0xC8 上 / 0xC9 下 / 0x86 活板門(下)  /  站在上面按 **K**(Klimb)  /  `sub_EA0` → `sub_758(0 或 2, 196)`  / | `03-scene-entry-and-tile-semantics.md`, `49-command-table-and-two-empty-keys.md`, `52-one-coordinate-pair-one-tile-accessor.md` |
| `sub_FB4` | 主要函式  /  ★ `sub_2BCC8`(再生戒指)、`sub_2ECE8`(戒指的每回合效果)、`sub_1F4E4`(力場消散)、`sub_2A50C` 尾段、`sub_16370`、★ `sub_FEC` + … | `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_FEC` | 主要函式  /  ★ `sub_2BCC8`(再生戒指)、`sub_2ECE8`(戒指的每回合效果)、`sub_1F4E4`(力場消散)、`sub_2A50C` 尾段、`sub_16370`、★ `sub_FEC` + … | `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_1158` | ; sub_1158 —— 讀一個鍵的包裝 | `70-hunger-poison-and-the-vanishing-rings.md` |
| `sub_11E0` | 主要函式  /  `sub_11E0`(彈奏與比對)、`sub_DB10`(定位被切換的那一格)  / | `35-harp-and-the-secret-door.md` |
| `sub_1318` | 主要函式  /  `sub_10BDC`(踏進沼澤的中毒)、`sub_1318`(每回合的地形效果)、`sub_2D9D0`(移動後的分派)、`sub_1F570`(落空的訊息)、`sub_1F528`(印單位名字)、`… | `70-hunger-poison-and-the-vanishing-rings.md`, `74-two-swamp-dice-and-an-unnamed-failure.md`, `75-open-did-nothing-outside-dungeons.md`, `80-implemented-but-unreachable.md`, `81-three-mode-loops-are-mutually-exclusive.md`, `92-white-noise-shim-and-ambient-sound.md` |
| `sub_154C` | 主要函式  /  `sub_1B3D0`(盤查)、`sub_1B52C`(對話分派)、`sub_195C`(碰到 NPC)、`sub_1884`(逮捕)、`sub_1B140`(密語比對)、`sub_154C` → `s… | `29-npc-behaviour-and-arrest.md`, `32-guard-challenge.md` |
| `sub_1568` | 篩選哪些居民受影響的 `sub_1568`: | `28-shadowlords-and-blackthorn.md` |
| `sub_15C4` | 6. 補記:被盤據的城裡會發生什麼(`sub_48C` + `sub_15C4`) | `28-shadowlords-and-blackthorn.md` |
| `sub_1638` | if (byte_3E0D8[i] < 0x80) sub_1638(i);      // 「空氣中瀰漫著…」 | `28-shadowlords-and-blackthorn.md` |
| `sub_1678` | 7  /  `sub_2D72C`(依腳下 tile 進場景)、`sub_1678`、`sub_C778`、跳表 case 62(走不到)  / | `11-map-objects.md`, `28-shadowlords-and-blackthorn.md`, `36-sandalwood-box-npc-objects.md`, `81-three-mode-loops-are-mutually-exclusive.md`, `87-bgm-selection-and-the-tbl-tables.md` |
| `sub_1884` | 主要函式  /  `sub_1B3D0`(盤查)、`sub_1B52C`(對話分派)、`sub_195C`(碰到 NPC)、`sub_1884`(逮捕)、`sub_1B140`(密語比對)、`sub_154C` → `s… | `100-the-last-three-and-the-collapsed-entrance-gate.md`, `26-yell-words-of-power-shadowlords.md`, `28-shadowlords-and-blackthorn.md`, `29-npc-behaviour-and-arrest.md`, `32-guard-challenge.md`, `87-bgm-selection-and-the-tbl-tables.md` |
| `sub_195C` | 主要函式  /  `sub_1B3D0`(盤查)、`sub_1B52C`(對話分派)、`sub_195C`(碰到 NPC)、`sub_1884`(逮捕)、`sub_1B140`(密語比對)、`sub_154C` → `s… | `29-npc-behaviour-and-arrest.md`, `32-guard-challenge.md`, `33-get-command.md`, `35-harp-and-the-secret-door.md` |
| `sub_1A54` | 主要函式  /  `sub_0`(切換器)、`sub_2DD44`→`sub_2D9D0`(大地圖)、`sub_1A54`→`sub_1318`(場景)、`sub_5378`→`sub_5150`(地牢)、`sub_2A… | `32-guard-challenge.md`, `35-harp-and-the-secret-door.md`, `66-hexrays-truncation-audit.md`, `74-two-swamp-dice-and-an-unnamed-failure.md`, `75-open-did-nothing-outside-dungeons.md`, `81-three-mode-loops-are-mutually-exclusive.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_1DC8` | 每月 28 天、每年 13 個月**。一般行動每回合 **1 分鐘**(`sub_1DC8` → `sub_29304(1)`); | `04-npc-schedule-and-clock.md` |
| `sub_1E74` | 主要函式  /  `sub_1E74`(配物件槽)、`sub_268`(放掉)、`sub_218`(記永久移除)、`sub_2E0`(反查)、`sub_1DC`(在不在場)  / | `33-get-command.md`, `34-ending-trigger.md`, `36-sandalwood-box-npc-objects.md`, `57-crown-and-sceptre-placement.md` |
| `sub_1F98` | 主要函式  /  `sub_2E24`(世界回合)、`sub_1F98`(遭遇門檻)、`sub_2218`(生怪)、`sub_203C`(依地形選怪)、`sub_22B0`(這一槽是不是生物)  / | `38-terrain-movement-cost.md`, `82-the-overworld-world-turn.md` |
| `sub_2008` | 5. 怪種表(`sub_203C` + `sub_2008`)—— ★★ 四張表的權重都恰好 256 | `82-the-overworld-world-turn.md` |
| `sub_203C` | 主要函式  /  `sub_2E24`(世界回合)、`sub_1F98`(遭遇門檻)、`sub_2218`(生怪)、`sub_203C`(依地形選怪)、`sub_22B0`(這一槽是不是生物)  / | `82-the-overworld-world-turn.md` |
| `sub_215C` | sub_215C()                            ; 挑候選座標 → word_3E086 / word_3E088 | `82-the-overworld-world-turn.md` |
| `sub_2218` | 主要函式  /  `sub_2E24`(世界回合)、`sub_1F98`(遭遇門檻)、`sub_2218`(生怪)、`sub_203C`(依地形選怪)、`sub_22B0`(這一槽是不是生物)  / | `38-terrain-movement-cost.md`, `82-the-overworld-world-turn.md`, `88-the-load-window-culling-and-ship-cannon.md` |
| `sub_22B0` | 主要函式  /  `sub_2E24`(世界回合)、`sub_1F98`(遭遇門檻)、`sub_2218`(生怪)、`sub_203C`(依地形選怪)、`sub_22B0`(這一槽是不是生物)  / | `82-the-overworld-world-turn.md`, `88-the-load-window-culling-and-ship-cannon.md` |
| `sub_22F0` | ★ **這次找回來的**  /  `sub_2CCFC` 轉向 + `Hull weak!`、`sub_2D9D0` 的 `Rough seas!`、以及它們共同呼叫的 `sub_22F0` 沉船  / | `66-hexrays-truncation-audit.md`, `68-troll-bridge-and-collision.md`, `80-implemented-but-unreachable.md`, `83-monster-attacks-and-two-different-wind-rules.md`, `88-the-load-window-culling-and-ship-cannon.md` |
| `sub_23FC` | 主要函式  /  `sub_2CBEC`(★ 視窗定位)、`sub_2D014`(★ 捲動)、`sub_2E24` 尾段(★ 清場)、`sub_2B24`(⚠ 第二參數是死的)、`sub_27F0`(影主的遮罩表)、`s… | `11-map-objects.md`, `66-hexrays-truncation-audit.md`, `83-monster-attacks-and-two-different-wind-rules.md`, `88-the-load-window-culling-and-ship-cannon.md` |
| `sub_24DC` | `sub_24DC` 的接觸三條  /  `game.(*State).contactParty`  /  ✅ 含漩渦、流沙、小艇例外  / | `66-hexrays-truncation-audit.md`, `83-monster-attacks-and-two-different-wind-rules.md`, `88-the-load-window-culling-and-ship-cannon.md` |
| `sub_25F0` | 主要函式  /  `sub_2CBEC`(★ 視窗定位)、`sub_2D014`(★ 捲動)、`sub_2E24` 尾段(★ 清場)、`sub_2B24`(⚠ 第二參數是死的)、`sub_27F0`(影主的遮罩表)、`s… | `38-terrain-movement-cost.md`, `66-hexrays-truncation-audit.md`, `82-the-overworld-world-turn.md`, `83-monster-attacks-and-two-different-wind-rules.md`, `88-the-load-window-culling-and-ship-cannon.md` |
| `sub_2778` | 主要函式  /  `sub_2B24`(追人)、`sub_2A54`(隨機遊走)、`sub_2870`(真的挪一格)、`sub_2778`(進得去嗎)、`sub_27CC`(剛空出來的格子)  / | `85-monsters-move-and-terrain-slows-them.md` |
| `sub_27CC` | 主要函式  /  `sub_2B24`(追人)、`sub_2A54`(隨機遊走)、`sub_2870`(真的挪一格)、`sub_2778`(進得去嗎)、`sub_27CC`(剛空出來的格子)  / | `85-monsters-move-and-terrain-slows-them.md` |
| `sub_27F0` | 主要函式  /  `sub_2CBEC`(★ 視窗定位)、`sub_2D014`(★ 捲動)、`sub_2E24` 尾段(★ 清場)、`sub_2B24`(⚠ 第二參數是死的)、`sub_27F0`(影主的遮罩表)、`s… | `83-monster-attacks-and-two-different-wind-rules.md`, `88-the-load-window-culling-and-ship-cannon.md` |
| `sub_2870` | 主要函式  /  `sub_DEE4`(開關月門)、`sub_DE74`(這一顆算不算在場)、`sub_E084`(玩家踏上)、`sub_2870`(怪物踏上)、`sub_29BEC`(★ 重畫時挑動畫格)、`sub_2… | `85-monsters-move-and-terrain-slows-them.md`, `86-tile-0xDC-is-an-open-moongate.md`, `88-the-load-window-culling-and-ship-cannon.md` |
| `sub_2A54` | 主要函式  /  `sub_2CCFC`(★ 真正的航行閘門)、`sub_2D174`(移動分派)、`sub_2D014`(改座標)、`sub_2D2D0`(揚帆的節奏)、`sub_2B24`(★ 怪物移動)、`sub_… | `83-monster-attacks-and-two-different-wind-rules.md`, `84-the-sailing-gate-and-the-wind-table.md`, `85-monsters-move-and-terrain-slows-them.md` |
| `sub_2B24` | 主要函式  /  `sub_2CBEC`(★ 視窗定位)、`sub_2D014`(★ 捲動)、`sub_2E24` 尾段(★ 清場)、`sub_2B24`(⚠ 第二參數是死的)、`sub_27F0`(影主的遮罩表)、`s… | `84-the-sailing-gate-and-the-wind-table.md`, `85-monsters-move-and-terrain-slows-them.md`, `88-the-load-window-culling-and-ship-cannon.md` |
| `sub_2D38` | 主要函式  /  `sub_25F0`(物件的攻擊)、`sub_2D38`(物件能不能動)、`sub_2D2D0`(**玩家**的自動航行)、`dword_4FD50`(敵船的風向延遲表)、`byte_601C7`/`b… | `23-wind-and-sailing.md`, `38-terrain-movement-cost.md`, `66-hexrays-truncation-audit.md`, `82-the-overworld-world-turn.md`, `83-monster-attacks-and-two-different-wind-rules.md`, `84-the-sailing-gate-and-the-wind-table.md`, `85-monsters-move-and-terrain-slows-them.md`, `88-the-load-window-culling-and-ship-cannon.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_2E24` | 主要函式  /  `sub_2CBEC`(★ 視窗定位)、`sub_2D014`(★ 捲動)、`sub_2E24` 尾段(★ 清場)、`sub_2B24`(⚠ 第二參數是死的)、`sub_27F0`(影主的遮罩表)、`s… | `38-terrain-movement-cost.md`, `50-hole-up-camp-sleep-repair.md`, `81-three-mode-loops-are-mutually-exclusive.md`, `82-the-overworld-world-turn.md`, `83-monster-attacks-and-two-different-wind-rules.md`, `84-the-sailing-gate-and-the-wind-table.md`, `88-the-load-window-culling-and-ship-cannon.md` |
| `sub_2F48` | 主要函式  /  `sub_3010`(過橋)、`sub_2F48`(過路費)、`sub_2CE70`(通行判定 + 撞擊)、`sub_2D9D0`(移動後的分派)  / | `68-troll-bridge-and-collision.md` |
| `sub_3010` | 主要函式  /  `sub_3010`(過橋)、`sub_2F48`(過路費)、`sub_2CE70`(通行判定 + 撞擊)、`sub_2D9D0`(移動後的分派)  / | `68-troll-bridge-and-collision.md` |
| `sub_30E8` | `sub_30E8` / `sub_4834`  /  1, ?, 0x4E20  /  無聲  / | `92-white-noise-shim-and-ambient-sound.md` |
| `sub_3124` | `sub_3124`  /  寫 `dword_4FDD8/DC` 並包一對 `sub_27754`  /  疑為游標 / 中斷遮罩,不碰遊戲狀態  / | `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_31AC` | `sub_31AC`(★ 地牢 tile 讀取,`< 0x90` 時**清掉 0x08 位元** —— 要與引擎的 | `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_31F0` | `byte_3F3D8` 的語意**未解**。取用它的 `sub_31F0` 只抽 0x90 / 0x60 / 0x0F, | `48-dungeon-wandering-monster-and-arena.md` |
| `sub_33F0` | 噴泉的動畫(`sub_39A8` + `byte_4FFA0` 三格循環)與力場動畫(`sub_33F0`)。 | `18-dungeons.md` |
| `sub_35EC` | `sub_35EC`  /  0x12 = 18  /  `SUITEKI3.SND`  /  「水滴」  / | `90-cdda-scenes-and-sound-effect-indices.md` |
| `sub_36C0` | if !sub_36C0(x, y, depth): break     ; 這一格擋不擋視線(擋住就在裡面畫正面) | `18-dungeons.md` |
| `sub_3878` | sub_3878(x + 側向, y + 側向, 0, depth)   ; 左 | `18-dungeons.md` |
| `sub_39A8` | 噴泉的動畫(`sub_39A8` + `byte_4FFA0` 三格循環)與力場動畫(`sub_33F0`)。 | `18-dungeons.md` |
| `sub_3B88` | 走廊裡的物件**(梯子、寶箱、噴泉、陷阱、頭上的洞):`sub_3B88` 依 | `18-dungeons.md` |
| `sub_3D14` | 7.1 繪圖流程(`sub_3D14`) | `18-dungeons.md` |
| `sub_3ED0` | sub_3ED0(樓層 + delta, …)               ; 才檢查那一格能不能過 | `18-dungeons.md`, `78-the-bottom-of-a-dungeon-opens-into-the-underworld.md` |
| `sub_3F34` | 主要函式  /  `sub_3FE4`(離開地牢)、`sub_3F34`(換層 + 邊界)、`sub_417C`(Klimb)、`sub_4074` / `sub_100F8` / `sub_1994C`(三個呼叫端) … | `17-magic.md`, `18-dungeons.md`, `78-the-bottom-of-a-dungeon-opens-into-the-underworld.md` |
| `sub_3FE4` | 1  /  `sub_31DC0` default、`sub_DF84`(月門傳送)、`sub_3FE4`(離開地牢)、`sub_86C`(方向樓梯)、`sub_1DA10`、`sub_177AC`、`sub_3FE4`… | `78-the-bottom-of-a-dungeon-opens-into-the-underworld.md`, `87-bgm-selection-and-the-tbl-tables.md` |
| `sub_4074` | 主要函式  /  `sub_3FE4`(離開地牢)、`sub_3F34`(換層 + 邊界)、`sub_417C`(Klimb)、`sub_4074` / `sub_100F8` / `sub_1994C`(三個呼叫端) … | `34-ending-trigger.md`, `41-jimmy-neworder-gem-ztats.md`, `49-command-table-and-two-empty-keys.md`, `60-command-echo-and-menu-keys.md`, `78-the-bottom-of-a-dungeon-opens-into-the-underworld.md` |
| `sub_417C` | 主要函式  /  `sub_3FE4`(離開地牢)、`sub_3F34`(換層 + 邊界)、`sub_417C`(Klimb)、`sub_4074` / `sub_100F8` / `sub_1994C`(三個呼叫端) … | `18-dungeons.md`, `49-command-table-and-two-empty-keys.md`, `52-one-coordinate-pair-one-tile-accessor.md`, `78-the-bottom-of-a-dungeon-opens-into-the-underworld.md` |
| `sub_42CC` | 主要函式  /  `sub_18380`(ESC = 撤離)、`sub_A360`(戰鬥分派器)、`sub_42CC`(進地牢房間)、`sub_A9EC`(`byte_3E0B3` 的設定處)、`sub_2E364`(戰… | `18-dungeons.md`, `34-ending-trigger.md`, `73-escape-is-a-key-and-it-has-two-gates.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_4460` | IDA 位址:`sub_5150`(每步的分派)、`sub_4460`(生成)、`sub_4594`(找落點)、 | `48-dungeon-wandering-monster-and-arena.md` |
| `sub_4504` | `byte_3EE15` 是朝向這件事是由 `sub_4504`(地牢 HUD 的羅盤)獨立確認的: | `48-dungeon-wandering-monster-and-arena.md` |
| `sub_4594` | IDA 位址:`sub_5150`(每步的分派)、`sub_4460`(生成)、`sub_4594`(找落點)、 | `18-dungeons.md`, `48-dungeon-wandering-monster-and-arena.md` |
| `sub_4834` | `sub_48F4` 在移動時攔下 `tile == 0x83`,呼叫 `sub_4834`: | `18-dungeons.md`, `92-white-noise-shim-and-ambient-sound.md` |
| `sub_48F4` | `sub_48F4` 在左轉與右轉前檢查 `目前這一格 & 0xF0 == 0xE0`,成立就印 | `18-dungeons.md` |
| `sub_4B14` | 主要函式  /  ★ `sub_2BD40`(Set Active Plr)、`sub_E19C`(選隊員,見 `docs/re/95`)、`sub_4B14`(按鍵分派的 default)  / | `49-command-table-and-two-empty-keys.md`, `97-set-active-player.md` |
| `sub_4C6C` | 3. `sub_4C6C` 唯一豁免移動的索引是 **0x1B = 27 = Reaper** —— U5 裡它是紮了根的 | `48-dungeon-wandering-monster-and-arena.md` |
| `sub_4DC8` | 睡眠與毒**不是全隊一律中**:`sub_4DC8` / `sub_4E58` 逐人擲 `random(1,30)`, | `17-magic.md`, `18-dungeons.md`, `92-white-noise-shim-and-ambient-sound.md` |
| `sub_4E58` | 睡眠與毒**不是全隊一律中**:`sub_4DC8` / `sub_4E58` 逐人擲 `random(1,30)`, | `17-magic.md`, `18-dungeons.md`, `92-white-noise-shim-and-ambient-sound.md` |
| `sub_4EB8` | `0x61` / `0x69`  /  **陷阱坑** —— 掉下一層(`sub_4EB8`)  / | `18-dungeons.md` |
| `sub_5008` | `sub_4C6C`(移動與撲擊)、`sub_5008`(「遭到襲擊!」)、`sub_2E364`(開打)、 | `48-dungeon-wandering-monster-and-arena.md` |
| `sub_5150` | 主要函式  /  `sub_0`(切換器)、`sub_2DD44`→`sub_2D9D0`(大地圖)、`sub_1A54`→`sub_1318`(場景)、`sub_5378`→`sub_5150`(地牢)、`sub_2A… | `18-dungeons.md`, `21-chests-fields-locks.md`, `48-dungeon-wandering-monster-and-arena.md`, `81-three-mode-loops-are-mutually-exclusive.md` |
| `sub_5378` | 主要函式  /  `sub_0`(切換器)、`sub_2DD44`→`sub_2D9D0`(大地圖)、`sub_1A54`→`sub_1318`(場景)、`sub_5378`→`sub_5150`(地牢)、`sub_2A… | `18-dungeons.md`, `48-dungeon-wandering-monster-and-arena.md`, `81-three-mode-loops-are-mutually-exclusive.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_584B` | 1. `sub_584B` 是整段開場 | `24-intro.md` |
| `sub_5F98` | `sub_5F98`  /  載 `ULTIMA.16` + `WD.BIT` 的標題畫面;順手把地點碼設成 0x40  /  美術流程(P6);地點碼 0x40 = 標題畫面這件事記在這裡  / | `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_60BC` | `sub_60BC` / `sub_6730`  /  `game.MainMenu`(`-menu` 旗標進入)  / | `39-character-creation.md`, `40-push-and-the-main-menu.md`, `90-cdda-scenes-and-sound-effect-indices.md` |
| `sub_6128` | `sub_6128`  /  8  /  **製作群字幕** —— 八段各一次(`"Produced and Designed by"`、`"FM-TOWNS version by"`、`"FM-TOWNS Art Wo… | `90-cdda-scenes-and-sound-effect-indices.md`, `audio-pipeline.md` |
| `sub_6730` | 主要函式  /  `sub_310E8`(★ 填 MSF 表)、`sub_30F8C`(★ 播一段 CD)、`sub_6730`(啟動 / 選單)、`sub_6128`(製作群字幕)、`sub_22D00`(石室)、`s… | `01-tileset-and-dot16-loader.md`, `40-push-and-the-main-menu.md`, `61-tavern-lore-and-the-transfer-second-stage.md`, `87-bgm-selection-and-the-tbl-tables.md`, `90-cdda-scenes-and-sound-effect-indices.md` |
| `sub_71D0` | `dword_54498` **只有一個寫入點** —— `sub_71D0`,也就是**從 Ultima IV 轉入角色**: | `54-pass-vehicle-verbs-and-the-avatar-line.md`, `55-transfer-from-ultima-iv.md`, `61-tavern-lore-and-the-transfer-second-stage.md` |
| `sub_74BC` | ⬜ 還沒讀  /  `sub_A360`(**558 行、34 個字串**,最大的一筆)、`sub_D650`(`Wanted:` 通緝告示?)、`sub_74BC`、`sub_23D50`(`EGA*.TIL` 載入)… | `66-hexrays-truncation-audit.md` |
| `sub_7564` | `sub_7594`(轉入的主流程)、`sub_7564`(三圍換算曲線)、`sub_71D0`(讀 U4 存檔)  / | `61-tavern-lore-and-the-transfer-second-stage.md` |
| `sub_7594` | `sub_7594`(轉入的主流程)、`sub_7564`(三圍換算曲線)、`sub_71D0`(讀 U4 存檔)  / | `54-pass-vehicle-verbs-and-the-avatar-line.md`, `55-transfer-from-ultima-iv.md`, `61-tavern-lore-and-the-transfer-second-stage.md` |
| `sub_8858` | 來源:FM Towns `WORRIORS.EXP`。相關函式 `sub_8858`(載入)、`sub_9C7C`(排程 slot)、 | `04-npc-schedule-and-clock.md` |
| `sub_8924` | 0 號槽是隊伍自己**:`sub_8924` 的更新迴圈從 `esi = 1` 起跑。檔案裡 0 號槽的內容 | `04-npc-schedule-and-clock.md`, `12-npc-movement.md` |
| `sub_89EC` | 主要函式  /  ★ `sub_2D72C`(Enter 的 tile 分派)、`sub_2D564`、`sub_27D24`(讀檔序列)、★ `sub_92C0` + `sub_89EC` + `sub_8BA0`(N… | `100-the-last-three-and-the-collapsed-entrance-gate.md`, `12-npc-movement.md` |
| `sub_8A1C` | ├ sub_8BA0    尋路 —— 內含 sub_8A1C 建格子圖 + 環狀佇列 BFS | `12-npc-movement.md` |
| `sub_8BA0` | 主要函式  /  ★ `sub_2D72C`(Enter 的 tile 分派)、`sub_2D564`、`sub_27D24`(讀檔序列)、★ `sub_92C0` + `sub_89EC` + `sub_8BA0`(N… | `100-the-last-three-and-the-collapsed-entrance-gate.md`, `12-npc-movement.md` |
| `sub_8D28` | dir  /  BFS(`sub_8BA0`)  /  走路 / 回溯(`sub_8EA4`、`sub_8D28`)  / | `12-npc-movement.md` |
| `sub_8EA4` | dir  /  BFS(`sub_8BA0`)  /  走路 / 回溯(`sub_8EA4`、`sub_8D28`)  / | `12-npc-movement.md` |
| `sub_8F3C` | `sub_8F3C`(距離)、`sub_195C`(接觸)、`sub_C10`(叫衛兵)、`sub_B44` / `sub_B98` | `29-npc-behaviour-and-arrest.md` |
| `sub_8F60` | 位址:`sub_95BC`(行為型別跳表)、`sub_94E0`(遊走)、`sub_8F60`(追擊 / 逃)、 | `29-npc-behaviour-and-arrest.md` |
| `sub_91A4` | 只有模式 ≤ 1(不存在或閒置)才重新呼叫 `sub_91A4`**。正在移動中的 NPC | `12-npc-movement.md` |
| `sub_92C0` | 主要函式  /  ★ `sub_2D72C`(Enter 的 tile 分派)、`sub_2D564`、`sub_27D24`(讀檔序列)、★ `sub_92C0` + `sub_89EC` + `sub_8BA0`(N… | `100-the-last-three-and-the-collapsed-entrance-gate.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_9358` | 欄位位置由 `sub_9358` 證實(`rec[slot+3]` / `rec[slot+6]` / `rec[slot+9]`,slot ∈ 0..2)。 | `04-npc-schedule-and-clock.md`, `12-npc-movement.md`, `29-npc-behaviour-and-arrest.md` |
| `sub_9428` | if (!sub_9428(nx, ny, npc, slot)) return; | `12-npc-movement.md`, `29-npc-behaviour-and-arrest.md` |
| `sub_94E0` | 位址:`sub_95BC`(行為型別跳表)、`sub_94E0`(遊走)、`sub_8F60`(追擊 / 逃)、 | `12-npc-movement.md`, `29-npc-behaviour-and-arrest.md` |
| `sub_95BC` | 原版不是這樣:`sub_9690` 在 NPC **與玩家同層**時呼叫 `sub_95BC(npc, slot)`, | `29-npc-behaviour-and-arrest.md` |
| `sub_9690` | 原版不是這樣:`sub_9690` 在 NPC **與玩家同層**時呼叫 `sub_95BC(npc, slot)`, | `100-the-last-three-and-the-collapsed-entrance-gate.md`, `11-map-objects.md`, `12-npc-movement.md`, `29-npc-behaviour-and-arrest.md`, `50-hole-up-camp-sleep-repair.md` |
| `sub_9C7C` | 來源:FM Towns `WORRIORS.EXP`。相關函式 `sub_8858`(載入)、`sub_9C7C`(排程 slot)、 | `04-npc-schedule-and-clock.md`, `08-shops.md` |
| `sub_9CE8` | ⚠ **同一支函式裡兩個參數用兩套編號**:`sub_9CE8` 吃 tile 碼、 | `17-magic.md` |
| `sub_9E10` | `'N'`  /  In An  /  10  /  `sub_9E10` / `sub_AE20`  /  施法者放不出遠程、不能瞬移  / | `16-combat-turns-and-ai.md`, `17-magic.md` |
| `sub_9F08` | 0x8000  /  施法  /  `sub_9E10` / `sub_9F08`  /  法師、注視者、收割者、惡魔、海馬  / | `16-combat-turns-and-ai.md` |
| `sub_A108` | `'T'`  /  An Tym  /  10  /  `sub_A108` 開頭 return  /  敵人整個不動;順帶讓火把不燒  / | `16-combat-turns-and-ai.md`, `17-magic.md`, `67-corpser-and-the-sleeping-party-member.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_A310` | `, armed with ` 後面那三件武器怎麼串(`sub_A310` 逐件回傳長度,三件相加為 0 才印 `bare hands`) | `67-corpser-and-the-sleeping-party-member.md` |
| `sub_A360` | 主要函式  /  `sub_18380`(ESC = 撤離)、`sub_A360`(戰鬥分派器)、`sub_42CC`(進地牢房間)、`sub_A9EC`(`byte_3E0B3` 的設定處)、`sub_2E364`(戰… | `16-combat-turns-and-ai.md`, `17-magic.md`, `34-ending-trigger.md`, `42-equipment-slots-ready-wear.md`, `49-command-table-and-two-empty-keys.md`, `51-closing-four-known-gaps.md`, `52-one-coordinate-pair-one-tile-accessor.md`, `66-hexrays-truncation-audit.md`, `67-corpser-and-the-sleeping-party-member.md`, `73-escape-is-a-key-and-it-has-two-gates.md`, `92-white-noise-shim-and-ambient-sound.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_A9EC` | 主要函式  /  `sub_18380`(ESC = 撤離)、`sub_A360`(戰鬥分派器)、`sub_42CC`(進地牢房間)、`sub_A9EC`(`byte_3E0B3` 的設定處)、`sub_2E364`(戰… | `16-combat-turns-and-ai.md`, `72-ready-had-seven-missing-rules.md`, `73-escape-is-a-key-and-it-has-two-gates.md`, `87-bgm-selection-and-the-tbl-tables.md`, `98-select-menu-and-song-melody-matching.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_AC40` | `cmp byte_3E08A, 'T' / 'Q'`、`sub_AC40` 裡那段莫名其妙的「擲贏就把 mySide | `16-combat-turns-and-ai.md`, `17-magic.md` |
| `sub_AE20` | `'N'`  /  In An  /  10  /  `sub_9E10` / `sub_AE20`  /  施法者放不出遠程、不能瞬移  / | `16-combat-turns-and-ai.md`, `17-magic.md`, `67-corpser-and-the-sleeping-party-member.md` |
| `sub_B1D8` | 戰鬥中 → 最多 7 次隨機落點(sub_B1D8),第一個站得住的就過去 | `17-magic.md` |
| `sub_B210` | `sub_B210(槽+1)`  /  `ObjectSet.Remove`(⚠ 原版只清 +0..+5,引擎清整筆 8 B;+6/+7 對力場沒有意義) / | `17-magic.md`, `73-escape-is-a-key-and-it-has-two-gates.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_B274` | `sub_B274` 對角色讀 `byte_3DDCC[角色*32]`,而 `0x3DDCC − 0x3DDB4 = 0x18`。 | `15-combat-formulas.md`, `16-combat-turns-and-ai.md`, `67-corpser-and-the-sleeping-party-member.md` |
| `sub_B35C` | │       ├ sub_B35C   戰場單位的 +1 欄位 | `15-combat-formulas.md` |
| `sub_B398` | `sub_B398` 印證了這件事:它取 `byte_3F050[生物*8]` 當「力量那一項」、 | `15-combat-formulas.md`, `17-magic.md`, `74-two-swamp-dice-and-an-unnamed-failure.md` |
| `sub_B484` | 主要函式  /  `sub_10BDC`(踏進沼澤的中毒)、`sub_1318`(每回合的地形效果)、`sub_2D9D0`(移動後的分派)、`sub_1F570`(落空的訊息)、`sub_1F528`(印單位名字)、`… | `15-combat-formulas.md`, `16-combat-turns-and-ai.md`, `17-magic.md`, `74-two-swamp-dice-and-an-unnamed-failure.md` |
| `sub_B51C` | 否則                                   → sub_B274 算傷害 → sub_B51C 扣血 | `16-combat-turns-and-ai.md`, `17-magic.md`, `67-corpser-and-the-sleeping-party-member.md`, `74-two-swamp-dice-and-an-unnamed-failure.md`, `88-the-load-window-culling-and-ship-cannon.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_B8DC` | 0x0004 / 0x0200  /  **下毒**  /  `sub_B9A8` → `sub_B8DC`  /  巨蟒、大烏賊、巨蜘蛛 / 巨鼠、擬態怪  / | `16-combat-turns-and-ai.md`, `17-magic.md`, `67-corpser-and-the-sleeping-party-member.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_B9A8` | 0x0004 / 0x0200  /  **下毒**  /  `sub_B9A8` → `sub_B8DC`  /  巨蟒、大烏賊、巨蜘蛛 / 巨鼠、擬態怪  / | `15-combat-formulas.md`, `16-combat-turns-and-ai.md`, `17-magic.md`, `67-corpser-and-the-sleeping-party-member.md`, `74-two-swamp-dice-and-an-unnamed-failure.md` |
| `sub_BAFC` | `woundLevel(idx)`  /  四級判定 + 掛 / 清逃跑旗標,逐行照 `sub_BAFC`  / | `67-corpser-and-the-sleeping-party-member.md` |
| `sub_BBA0` | `sub_BBA0`  /  `harmUnderUnit()` + `harmStandingUnit()`(接在 `afterPlayerAction` 與 `aiTurn` 之後) / | `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_BCC4` | 主要函式  /  `sub_A360`(隊員的戰鬥回合,**558 行**)、`sub_BCC4`(掙脫)、`sub_1F840`(命中結果)、`sub_2B724`/`sub_2B710`(門檻骰)、`sub_2ED5… | `67-corpser-and-the-sleeping-party-member.md`, `92-white-noise-shim-and-ambient-sound.md` |
| `sub_BD5C` | 主要函式  /  ★ `sub_2D72C`(Enter 的 tile 分派)、`sub_2D564`、`sub_27D24`(讀檔序列)、★ `sub_92C0` + `sub_89EC` + `sub_8BA0`(N… | `100-the-last-three-and-the-collapsed-entrance-gate.md` |
| `sub_BD80` | 主要函式  /  ★ `sub_2D72C`(Enter 的 tile 分派)、`sub_2D564`、`sub_27D24`(讀檔序列)、★ `sub_92C0` + `sub_89EC` + `sub_8BA0`(N… | `100-the-last-three-and-the-collapsed-entrance-gate.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_BE5C` | 主要函式  /  ★ `sub_2D72C`(Enter 的 tile 分派)、`sub_2D564`、`sub_27D24`(讀檔序列)、★ `sub_92C0` + `sub_89EC` + `sub_8BA0`(N… | `100-the-last-three-and-the-collapsed-entrance-gate.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_BFFC` | 位址:`sub_29304`(午夜遊走)、`sub_C414` → `sub_C318` → `sub_BFFC` / `sub_C098` / `sub_C13C` | `28-shadowlords-and-blackthorn.md` |
| `sub_C098` | 位址:`sub_29304`(午夜遊走)、`sub_C414` → `sub_C318` → `sub_BFFC` / `sub_C098` / `sub_C13C` | `28-shadowlords-and-blackthorn.md` |
| `sub_C13C` | 位址:`sub_29304`(午夜遊走)、`sub_C414` → `sub_C318` → `sub_BFFC` / `sub_C098` / `sub_C13C` | `27-codex-and-the-shrine-chamber.md`, `28-shadowlords-and-blackthorn.md` |
| `sub_C2BC` | 呼叫鏈:`sub_1884`(被逮捕)→ `sub_C414` → `sub_C318` → `sub_C2BC` / `sub_C2D0` | `100-the-last-three-and-the-collapsed-entrance-gate.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_C2D0` | 呼叫鏈:`sub_1884`(被逮捕)→ `sub_C414` → `sub_C318` → `sub_C2BC` / `sub_C2D0` | `100-the-last-three-and-the-collapsed-entrance-gate.md`, `28-shadowlords-and-blackthorn.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_C318` | 位址:`sub_29304`(午夜遊走)、`sub_C414` → `sub_C318` → `sub_BFFC` / `sub_C098` / `sub_C13C` | `100-the-last-three-and-the-collapsed-entrance-gate.md`, `25-shrines.md`, `26-yell-words-of-power-shadowlords.md`, `28-shadowlords-and-blackthorn.md` |
| `sub_C414` | 主要函式  /  `sub_1EC34`(Ready,**333 行**)、`sub_1E38C`(已經穿著嗎)、`sub_1EBE8`(空手狀況)、`sub_C414+1D0`(開戰佈陣)、`sub_A9EC`(`by… | `03-picture-files.md`, `100-the-last-three-and-the-collapsed-entrance-gate.md`, `28-shadowlords-and-blackthorn.md`, `29-npc-behaviour-and-arrest.md`, `34-ending-trigger.md`, `72-ready-had-seven-missing-rules.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_C778` | `sub_29D64` / `sub_C778`  /  `rep stosd` 整塊填 **0xFF**(0x2C dword = 176 B)  / | `03-scene-entry-and-tile-semantics.md`, `34-ending-trigger.md`, `87-bgm-selection-and-the-tbl-tables.md` |
| `sub_CC44` | `sub_CC44` / `sub_D45C`  /  `u5data.LookTable.Terrain` / `.Object`  / | `37-look-signs-and-the-sky.md` |
| `sub_CD28` | 主要函式  /  `sub_CD28`(井)、`sub_D258`(Look 分派)、`sub_27C98`(字串比對)、`sub_2B770`(文字輸入)、`sub_2B6C8`(寫物件槽)、`sub_2B57C`(找… | `37-look-signs-and-the-sky.md`, `65-the-wishing-well-easter-egg.md`, `66-hexrays-truncation-audit.md`, `92-white-noise-shim-and-ambient-sound.md` |
| `sub_CE78` | `sub_CE78` 全文:印「a gurgling fountain!」「Who will drink?」,選人, | `18-dungeons.md`, `37-look-signs-and-the-sky.md`, `95-the-dungeon-has-its-own-look.md` |
| `sub_D064` | `sub_D650` / `sub_D544`(招牌)、`sub_CE78`(噴泉)、`sub_D064`(天空) | `37-look-signs-and-the-sky.md` |
| `sub_D258` | 主要函式  /  `sub_CD28`(井)、`sub_D258`(Look 分派)、`sub_27C98`(字串比對)、`sub_2B770`(文字輸入)、`sub_2B6C8`(寫物件槽)、`sub_2B57C`(找… | `37-look-signs-and-the-sky.md`, `65-the-wishing-well-easter-egg.md`, `95-the-dungeon-has-its-own-look.md` |
| `sub_D45C` | `sub_CC44` / `sub_D45C`  /  `u5data.LookTable.Terrain` / `.Object`  / | `37-look-signs-and-the-sky.md`, `57-crown-and-sceptre-placement.md` |
| `sub_D544` | `sub_D650` / `sub_D544`  /  `u5data.SignSet.At` / `Sign.Render` / `Sign.Lines`  / | `37-look-signs-and-the-sky.md` |
| `sub_D650` | ⬜ 還沒讀  /  `sub_A360`(**558 行、34 個字串**,最大的一筆)、`sub_D650`(`Wanted:` 通緝告示?)、`sub_74BC`、`sub_23D50`(`EGA*.TIL` 載入)… | `37-look-signs-and-the-sky.md`, `66-hexrays-truncation-audit.md` |
| `sub_D9C4` | IDA 位址:`sub_D9C4`(指令)、`sub_D258`(地形)、`sub_CC44` / `sub_D45C`(敘述表)、 | `37-look-signs-and-the-sky.md`, `40-push-and-the-main-menu.md`, `41-jimmy-neworder-gem-ztats.md` |
| `sub_DB10` | tile = byte_3F8F4[y*32 + x]                       ; 這個單位腳下(sub_DB10 的戰場分支) | `100-the-last-three-and-the-collapsed-entrance-gate.md`, `26-yell-words-of-power-shadowlords.md`, `35-harp-and-the-secret-door.md`, `52-one-coordinate-pair-one-tile-accessor.md`, `68-troll-bridge-and-collision.md`, `69-the-overworld-cannon-did-nothing.md`, `71-the-use-list-had-29-empty-slots.md`, `74-two-swamp-dice-and-an-unnamed-failure.md`, `75-open-did-nothing-outside-dungeons.md`, `76-jimmy-in-a-dungeon-disarms-not-unlocks.md`, `86-tile-0xDC-is-an-open-moongate.md`, `92-white-noise-shim-and-ambient-sound.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_DBC8` | `sub_DBC8`  /  把 11×11 覆蓋層填成 0xFF  /  引擎的覆蓋層是另一套資料結構  / | `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_DE74` | 主要函式  /  `sub_DEE4`(開關月門)、`sub_DE74`(這一顆算不算在場)、`sub_E084`(玩家踏上)、`sub_2870`(怪物踏上)、`sub_29BEC`(★ 重畫時挑動畫格)、`sub_2… | `86-tile-0xDC-is-an-open-moongate.md`, `88-the-load-window-culling-and-ship-cannon.md` |
| `sub_DEE4` | 主要函式  /  `sub_DEE4`(開關月門)、`sub_DE74`(這一顆算不算在場)、`sub_E084`(玩家踏上)、`sub_2870`(怪物踏上)、`sub_29BEC`(★ 重畫時挑動畫格)、`sub_2… | `31-line-of-sight.md`, `86-tile-0xDC-is-an-open-moongate.md` |
| `sub_DF84` | 1  /  `sub_31DC0` default、`sub_DF84`(月門傳送)、`sub_3FE4`(離開地牢)、`sub_86C`(方向樓梯)、`sub_1DA10`、`sub_177AC`、`sub_3FE4`… | `22-moongates.md`, `86-tile-0xDC-is-an-open-moongate.md`, `87-bgm-selection-and-the-tbl-tables.md`, `88-the-load-window-culling-and-ship-cannon.md` |
| `sub_E084` | 主要函式  /  `sub_DEE4`(開關月門)、`sub_DE74`(這一顆算不算在場)、`sub_E084`(玩家踏上)、`sub_2870`(怪物踏上)、`sub_29BEC`(★ 重畫時挑動畫格)、`sub_2… | `22-moongates.md`, `71-the-use-list-had-29-empty-slots.md`, `86-tile-0xDC-is-an-open-moongate.md` |
| `sub_E19C` | 主要函式  /  ★★ `sub_EEEC`(地牢 Look)、★ `sub_13BA8`(第一人稱方向提問)、`sub_13B4C`(相對方向 → 座標)、★ `sub_E19C`(選隊員)、★ `sub_2B67C`… | `37-look-signs-and-the-sky.md`, `58-rune-input-cast-and-mix.md`, `75-open-did-nothing-outside-dungeons.md`, `76-jimmy-in-a-dungeon-disarms-not-unlocks.md`, `95-the-dungeon-has-its-own-look.md`, `97-set-active-player.md`, `98-select-menu-and-song-melody-matching.md` |
| `sub_E2A4` | `sub_E2A4`(整點時在大地圖 / 場景才呼叫的那一支)還沒逆。 | `22-moongates.md`, `92-white-noise-shim-and-ambient-sound.md` |
| `sub_E460` | `sub_E460`  /  在第 5 列印 12 個空白(清訊息列) /  同上  / | `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_EDD4` | 4 IQW "View!"           > 7Fh → "Not here!";  < 21h → sub_EDD4(x,y) else sub_F7C0() | `17-magic.md`, `37-look-signs-and-the-sky.md`, `41-jimmy-neworder-gem-ztats.md`, `49-command-table-and-two-empty-keys.md`, `71-the-use-list-had-29-empty-slots.md` |
| `sub_EEEC` | 主要函式  /  ★★ `sub_EEEC`(地牢 Look)、★ `sub_13BA8`(第一人稱方向提問)、`sub_13B4C`(相對方向 → 座標)、★ `sub_E19C`(選隊員)、★ `sub_2B67C`… | `95-the-dungeon-has-its-own-look.md` |
| `sub_F35C` | `0x00`  /  **通道**(可走)  /  繪圖 `sub_F35C` case 0 什麼都不畫  / | `03-picture-files.md`, `18-dungeons.md` |
| `sub_F7C0` | 4 IQW "View!"           > 7Fh → "Not here!";  < 21h → sub_EDD4(x,y) else sub_F7C0() | `17-magic.md`, `18-dungeons.md`, `49-command-table-and-two-empty-keys.md`, `71-the-use-list-had-29-empty-slots.md` |
| `sub_F9A0` | `sub_F9A0` / `sub_FA20` / `sub_FA7C`  /  `markRoomCleared()` / `roomIsCleared()` / `applyClearedRooms()`  / | `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_FA20` | `sub_F9A0` / `sub_FA20` / `sub_FA7C`  /  `markRoomCleared()` / `roomIsCleared()` / `applyClearedRooms()`  / | `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_FA7C` | `sub_F9A0` / `sub_FA20` / `sub_FA7C`  /  `markRoomCleared()` / `roomIsCleared()` / `applyClearedRooms()`  / | `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_FAC4` | 0xB0 / 0xC0 / 0xD0(牆)  /  **整條外環抹成 0xFF**(`sub_FAC4`)  / | `48-dungeon-wandering-monster-and-arena.md` |
| `sub_FB6C` | 其餘(房間 0xA0 / 0xF0、門 0xE0)  /  開 **5** 格(`sub_FB6C`)  / | `48-dungeon-wandering-monster-and-arena.md` |
| `sub_FBF0` | < 0xA0(通道、梯子、陷阱、門檻…)  /  開 **7** 格(`sub_FBF0`)  / | `48-dungeon-wandering-monster-and-arena.md` |
| `sub_FC7C` | `sub_FE48`(畫戰場)、`sub_FD54`(框)、`sub_FC7C`(四面牆)、 | `48-dungeon-wandering-monster-and-arena.md` |
| `sub_FD54` | 遊蕩怪物的戰場由 `sub_FE48` → `sub_FD54` 現畫在戰鬥地圖緩衝 `byte_3F8F4` 上。 | `48-dungeon-wandering-monster-and-arena.md`, `52-one-coordinate-pair-one-tile-accessor.md`, `53-party-tile-and-the-arena-floor.md` |
| `sub_FE48` | `sub_FE48`+`FD54`+`FC7C`+`FAC4`/`FB6C`/`FBF0`  /  `u5data.BuildDungeonArena`  / | `18-dungeons.md`, `48-dungeon-wandering-monster-and-arena.md`, `50-hole-up-camp-sleep-repair.md`, `53-party-tile-and-the-arena-floor.md` |
| `sub_100F8` | 主要函式  /  `sub_3FE4`(離開地牢)、`sub_3F34`(換層 + 邊界)、`sub_417C`(Klimb)、`sub_4074` / `sub_100F8` / `sub_1994C`(三個呼叫端) … | `48-dungeon-wandering-monster-and-arena.md`, `78-the-bottom-of-a-dungeon-opens-into-the-underworld.md` |
| `sub_10208` | `sub_10208`  /  **空函式**  /  —  / | `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_102B4` | call sub_102B4 | `48-dungeon-wandering-monster-and-arena.md` |
| `sub_10334` | ⬜ `sub_10334` 的 `arg_4 == 1`/`3` 兩條(房間戰鬥走的是 3), | `48-dungeon-wandering-monster-and-arena.md` |
| `sub_1056C` | `sub_1056C`  /  `u5data.DungeonIsSealed()`(★ 極性與名字直覺相反) / | `100-the-last-three-and-the-collapsed-entrance-gate.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_105AC` | `sub_105AC`  /  `ShrineFlag[i] & ShrineDesecratedBit`(這一邊極性正常) / | `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_105E4` | 3  /  `byte_55E18` = `18 16 16 18 18 17 17 16`(八座入口的**原始**地形),正是 `sub_105E4` 會改寫成 `0xDF` 的那三個值 ⇒ 兩支函式講同一件事  / | `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_10738` | sub_10738   memchr(byte_404F4, 0x1B, 0x400)     ← 在世界地圖的 32×32 視窗裡找 tile 0x1B | `31-line-of-sight.md`, `88-the-load-window-culling-and-ship-cannon.md` |
| `sub_10878` | sub_10878(dx, dy) | `88-the-load-window-culling-and-ship-cannon.md` |
| `sub_10910` | call    sub_10910               ; → byte_3E0A5 == 0 ? "A:BRIT.OOL" : "A:UNDER.OOL" | `11-map-objects.md` |
| `sub_10928` | `sub_1DA10`(聖壇)、`sub_2D564`(cave/mine/dungeon)、`sub_10928`(印地名並確認)。 | `03-scene-entry-and-tile-semantics.md` |
| `sub_10A1C` | `sub_10A1C`(墜落動畫) /  0x14 = 20  /  `T_OCHI1.SND`  /  ★★「落ち」= 掉落  / | `66-hexrays-truncation-audit.md`, `68-troll-bridge-and-collision.md`, `88-the-load-window-culling-and-ship-cannon.md`, `90-cdda-scenes-and-sound-effect-indices.md` |
| `sub_10B3C` | `sub_10B3C` 進地下世界時當場塞進物件槽,座標寫死在程式裡。王冠與權杖用同樣的方法卻怎麼都找不到: | `33-get-command.md`, `57-crown-and-sceptre-placement.md` |
| `sub_10BC4` | ✅ **`tile == 0x8F`(熔岩)在大地圖走 `sub_10BC4`**:印 `Burning!` + `sub_2A4D0`。 | `81-three-mode-loops-are-mutually-exclusive.md` |
| `sub_10BDC` | 主要函式  /  `sub_10BDC`(踏進沼澤的中毒)、`sub_1318`(每回合的地形效果)、`sub_2D9D0`(移動後的分派)、`sub_1F570`(落空的訊息)、`sub_1F528`(印單位名字)、`… | `74-two-swamp-dice-and-an-unnamed-failure.md`, `81-three-mode-loops-are-mutually-exclusive.md` |
| `sub_10C34` | ✅ 早就從組語逆過,截斷沒造成損失  /  `sub_21108` 酒單(`tavern.go`)、`sub_1D394` 聖壇與 `ALAKAZAM`(`shrine.go`)、`sub_14CAC`/`sub_14B… | `19-levelup.md`, `66-hexrays-truncation-audit.md`, `90-cdda-scenes-and-sound-effect-indices.md` |
| `sub_10FEC` | `sub_10FEC` 的 `@`  /  `u5data.TimeOfDay` + `i18n.TimeOfDay`(中文)  / | `05-text-compression.md`, `08-shops.md`, `10-shop-prices-and-trade.md`, `47-move-modes-and-time-of-day.md` |
| `sub_11168` | 位址  /  `sub_21500`(酒館打聽)、`sub_27C98`(關鍵字比對)、`sub_11168`(印 `SHOPPE.DAT`)  / | `08-shops.md`, `61-tavern-lore-and-the-transfer-second-stage.md` |
| `sub_11190` | 金幣 -= 總價;  sub_11190() | `93-the-tavern-meal-was-never-haggled.md` |
| `sub_111CC` | 來源:FM Towns `WORRIORS.EXP`。`sub_1B294`(進店)、`sub_111CC`(挑問候語)、 | `08-shops.md` |
| `sub_112F8` | 買(`sub_11AF0`、`sub_11588`、`sub_112F8`、`sub_118CC`、`sub_219B0` 都一樣) | `10-shop-prices-and-trade.md`, `93-the-tavern-meal-was-never-haggled.md` |
| `sub_11464` | 公會  /  0x86  /  `sub_11520` → `sub_11464` → `sub_112F8`  /  3  /  ✅  / | `10-shop-prices-and-trade.md` |
| `sub_11520` | 公會  /  0x86  /  `sub_11520` → `sub_11464` → `sub_112F8`  /  3  /  ✅  / | `10-shop-prices-and-trade.md` |
| `sub_11588` | 買(`sub_11AF0`、`sub_11588`、`sub_112F8`、`sub_118CC`、`sub_219B0` 都一樣) | `10-shop-prices-and-trade.md`, `93-the-tavern-meal-was-never-haggled.md` |
| `sub_1173C` | 藥草鋪  /  0x85  /  `sub_11864` → `sub_1173C` → `sub_11588`  /  5  /  ✅  / | `10-shop-prices-and-trade.md` |
| `sub_11864` | 藥草鋪  /  0x85  /  `sub_11864` → `sub_1173C` → `sub_11588`  /  5  /  ✅  / | `10-shop-prices-and-trade.md` |
| `sub_118CC` | 買(`sub_11AF0`、`sub_11588`、`sub_112F8`、`sub_118CC`、`sub_219B0` 都一樣) | `10-shop-prices-and-trade.md`, `11-map-objects.md`, `37-look-signs-and-the-sky.md`, `47-move-modes-and-time-of-day.md`, `65-the-wishing-well-easter-egg.md`, `93-the-tavern-meal-was-never-haggled.md` |
| `sub_11AF0` | 買(`sub_11AF0`、`sub_11588`、`sub_112F8`、`sub_118CC`、`sub_219B0` 都一樣) | `10-shop-prices-and-trade.md`, `33-get-command.md`, `93-the-tavern-meal-was-never-haggled.md` |
| `sub_12060` | 賣(`sub_12060`) | `10-shop-prices-and-trade.md` |
| `sub_1258C` | 武具店  /  0x81  /  `sub_1258C`  /  9  /  ✅ 買 + 賣  / | `10-shop-prices-and-trade.md` |
| `sub_12794` | `sub_12794` 收錢時有一條**與地點綁定的例外**: | `10-shop-prices-and-trade.md` |
| `sub_12838` | 解毒 20、療傷 35、復活 200。各有前置判斷(`sub_12838`): | `10-shop-prices-and-trade.md` |
| `sub_12B20` | `sub_12B20` → 0x0E、`sub_135FC` → 0x0D / 9、`sub_32244` → 0x0A  /  —  /  ⬜ 過場 / 開場,引擎沒有對應流程  / | `87-bgm-selection-and-the-tbl-tables.md` |
| `sub_13258` | 製作名單(`sub_13258`,含 "to Lord British at Origin Systems!")。 | `30-ending.md` |
| `sub_134CC` | 王座廳的 11×11 畫面(`MISCMAPS.DAT` 位移 0x210)與走位動畫(`sub_134CC`)—— | `30-ending.md` |
| `sub_13554` | for (;;) { sub_13554(1); sub_13554(3); sub_13554(4); sub_13554(5); }  // ★ 無窮迴圈 | `30-ending.md` |
| `sub_135FC` | 主要函式  /  `sub_DEE4`(開關月門)、`sub_DE74`(這一顆算不算在場)、`sub_E084`(玩家踏上)、`sub_2870`(怪物踏上)、`sub_29BEC`(★ 重畫時挑動畫格)、`sub_2… | `03-picture-files.md`, `30-ending.md`, `34-ending-trigger.md`, `86-tile-0xDC-is-an-open-moongate.md`, `87-bgm-selection-and-the-tbl-tables.md`, `90-cdda-scenes-and-sound-effect-indices.md` |
| `sub_13B30` | `DungeonTileAt` 對照,可能是一條落差)、`sub_13B30`(★ 找空物件槽, | `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_13B4C` | 主要函式  /  ★★ `sub_EEEC`(地牢 Look)、★ `sub_13BA8`(第一人稱方向提問)、`sub_13B4C`(相對方向 → 座標)、★ `sub_E19C`(選隊員)、★ `sub_2B67C`… | `95-the-dungeon-has-its-own-look.md` |
| `sub_13BA8` | 主要函式  /  ★★ `sub_EEEC`(地牢 Look)、★ `sub_13BA8`(第一人稱方向提問)、`sub_13B4C`(相對方向 → 座標)、★ `sub_E19C`(選隊員)、★ `sub_2B67C`… | `49-command-table-and-two-empty-keys.md`, `95-the-dungeon-has-its-own-look.md` |
| `sub_13CB0` | 主要函式  /  `sub_147A8`(地面 Search 主體)、★ `sub_13FC4`(挖月石)、★★ `sub_14090`(午夜採藥)、★★ `sub_14160`(113 筆固定物品)、★ `sub_13… | `96-search-had-three-more-layers.md` |
| `sub_13DD8` | `sub_13DD8`  /  `game.State.rollSearchFind` / `rollSearchJunk`  / | `41-jimmy-neworder-gem-ztats.md`, `43-search.md`, `92-white-noise-shim-and-ambient-sound.md`, `96-search-had-three-more-layers.md` |
| `sub_13F04` | IDA 位址:`sub_147A8`(主體)、`sub_13F04`(陷阱偵測)、`sub_13DD8`(翻到什麼) | `43-search.md` |
| `sub_13FC4` | 主要函式  /  `sub_147A8`(地面 Search 主體)、★ `sub_13FC4`(挖月石)、★★ `sub_14090`(午夜採藥)、★★ `sub_14160`(113 筆固定物品)、★ `sub_13… | `41-jimmy-neworder-gem-ztats.md`, `96-search-had-three-more-layers.md` |
| `sub_14090` | 主要函式  /  `sub_147A8`(地面 Search 主體)、★ `sub_13FC4`(挖月石)、★★ `sub_14090`(午夜採藥)、★★ `sub_14160`(113 筆固定物品)、★ `sub_13… | `41-jimmy-neworder-gem-ztats.md`, `96-search-had-three-more-layers.md` |
| `sub_14160` | 主要函式  /  `sub_147A8`(地面 Search 主體)、★ `sub_13FC4`(挖月石)、★★ `sub_14090`(午夜採藥)、★★ `sub_14160`(113 筆固定物品)、★ `sub_13… | `41-jimmy-neworder-gem-ztats.md`, `96-search-had-three-more-layers.md` |
| `sub_142EC` | ✅ 早就從組語逆過,截斷沒造成損失  /  `sub_21108` 酒單(`tavern.go`)、`sub_1D394` 聖壇與 `ALAKAZAM`(`shrine.go`)、`sub_14CAC`/`sub_14B… | `49-command-table-and-two-empty-keys.md`, `51-closing-four-known-gaps.md`, `66-hexrays-truncation-audit.md`, `76-jimmy-in-a-dungeon-disarms-not-unlocks.md` |
| `sub_147A8` | 主要函式  /  `sub_147A8`(地面 Search 主體)、★ `sub_13FC4`(挖月石)、★★ `sub_14090`(午夜採藥)、★★ `sub_14160`(113 筆固定物品)、★ `sub_13… | `41-jimmy-neworder-gem-ztats.md`, `42-equipment-slots-ready-wear.md`, `43-search.md`, `49-command-table-and-two-empty-keys.md`, `96-search-had-three-more-layers.md` |
| `sub_14B2C` | ✅ 早就從組語逆過,截斷沒造成損失  /  `sub_21108` 酒單(`tavern.go`)、`sub_1D394` 聖壇與 `ALAKAZAM`(`shrine.go`)、`sub_14CAC`/`sub_14B… | `66-hexrays-truncation-audit.md`, `76-jimmy-in-a-dungeon-disarms-not-unlocks.md` |
| `sub_14CAC` | ✅ 早就從組語逆過,截斷沒造成損失  /  `sub_21108` 酒單(`tavern.go`)、`sub_1D394` 聖壇與 `ALAKAZAM`(`shrine.go`)、`sub_14CAC`/`sub_14B… | `41-jimmy-neworder-gem-ztats.md`, `66-hexrays-truncation-audit.md`, `76-jimmy-in-a-dungeon-disarms-not-unlocks.md` |
| `sub_14F68` | 種類 2 的語意確定:`sub_14F68` 對它算 `random(1, 等級 × 3)`、上限 90 | `21-chests-fields-locks.md` |
| `sub_15020` | sub_15020(var_5, …);  sub_1509C(var_5, …)      ; 擲獎品(等級 = 品質低七位) | `21-chests-fields-locks.md`, `33-get-command.md`, `75-open-did-nothing-outside-dungeons.md` |
| `sub_1509C` | sub_15020(var_5, …);  sub_1509C(var_5, …)      ; 擲獎品(等級 = 品質低七位) | `21-chests-fields-locks.md`, `75-open-did-nothing-outside-dungeons.md` |
| `sub_15108` | 主要函式  /  `sub_15374`(Open 指令本體)、`sub_152B8`(地牢的開箱)、`sub_15108`(物件層的開箱)、`sub_2B64C`(把 tile 寫回去)、`sub_1A54`(主迴圈的… | `21-chests-fields-locks.md`, `75-open-did-nothing-outside-dungeons.md` |
| `sub_152B8` | 主要函式  /  `sub_15374`(Open 指令本體)、`sub_152B8`(地牢的開箱)、`sub_15108`(物件層的開箱)、`sub_2B64C`(把 tile 寫回去)、`sub_1A54`(主迴圈的… | `75-open-did-nothing-outside-dungeons.md` |
| `sub_15374` | 主要函式  /  `sub_15374`(Open 指令本體)、`sub_152B8`(地牢的開箱)、`sub_15108`(物件層的開箱)、`sub_2B64C`(把 tile 寫回去)、`sub_1A54`(主迴圈的… | `75-open-did-nothing-outside-dungeons.md`, `76-jimmy-in-a-dungeon-disarms-not-unlocks.md` |
| `sub_154BC` | 位址  /  `sub_154BC`(Get 分派)`loc_15863` / `loc_158A2` / `loc_158D0` / `loc_158DE`  / | `33-get-command.md`, `36-sandalwood-box-npc-objects.md`, `44-use-item.md`, `57-crown-and-sceptre-placement.md`, `59-dos-overlay-map.md` |
| `sub_15930` | 主要函式  /  `sub_15A94`(Get)、`sub_154BC`(收進背包)、`sub_15930`(地牢寶箱)  / | `33-get-command.md` |
| `sub_15A94` | 主要函式  /  `sub_15A94`(Get)、`sub_154BC`(收進背包)、`sub_15930`(地牢寶箱)  / | `33-get-command.md`, `36-sandalwood-box-npc-objects.md`, `41-jimmy-neworder-gem-ztats.md` |
| `sub_15DD4` | `sub_15DD4` 每次行動後重數兩邊。⚠ 它用的兩個全域 `word_3E086` / `word_3E088` | `16-combat-turns-and-ai.md` |
| `sub_15E20` | 主要函式  /  `sub_135FC`(結局)、`sub_161E4`(觸發)、`sub_15E20`(撤離)、`sub_297F4`(疊圖)  / | `34-ending-trigger.md`, `52-one-coordinate-pair-one-tile-accessor.md` |
| `sub_15F18` | `dword_3EF50` 的 +6 / +7 是戰場 X / Y —— 這一點由 `sub_15F18` 獨立確認: | `34-ending-trigger.md` |
| `sub_16058` | `sub_16058` 判「爬得過去」用的是 tile 0x4C,而本專案 `partyTileFor()` | `34-ending-trigger.md`, `52-one-coordinate-pair-one-tile-accessor.md`, `53-party-tile-and-the-arena-floor.md`, `72-ready-had-seven-missing-rules.md` |
| `sub_161E4` | 主要函式  /  `sub_135FC`(結局)、`sub_161E4`(觸發)、`sub_15E20`(撤離)、`sub_297F4`(疊圖)  / | `34-ending-trigger.md` |
| `sub_16270` | 不可用的每一個都有自己的回應**(`sub_16270(名字, 種類)`): | `51-closing-four-known-gaps.md` |
| `sub_16370` | 主要函式  /  ★ `sub_2BCC8`(再生戒指)、`sub_2ECE8`(戒指的每回合效果)、`sub_1F4E4`(力場消散)、`sub_2A50C` 尾段、`sub_16370`、★ `sub_FEC` + … | `17-magic.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_163B0` | sub_163B0  (戰鬥收尾) … cmp byte_3E0B0, 4Dh ; jnz …  → call sub_135FC | `34-ending-trigger.md` |
| `sub_16454` | 0x02  /  逃跑中  /  `sub_AC40` 反轉方向、`sub_16454` 放行出界  / | `16-combat-turns-and-ai.md`, `17-magic.md`, `67-corpser-and-the-sleeping-party-member.md` |
| `sub_16538` | 我方全滅時原版**不會立刻判負**:先叫 `sub_16538` 找一個被魅惑的隊員, | `16-combat-turns-and-ai.md` |
| `sub_165C8` | `sub_165C8` / `sub_21D48` 的 `sub_2BCC8`  /  `camp()` 每小時 / `SleepUntilMorning()` 每 9 分鐘  / | `19-levelup.md`, `48-dungeon-wandering-monster-and-arena.md`, `50-hole-up-camp-sleep-repair.md`, `51-closing-four-known-gaps.md`, `87-bgm-selection-and-the-tbl-tables.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_16BA0` | 4  /  `sub_165C8`(紮營突襲)、`sub_16BA0`、`sub_21D48`(Rest / 住宿)、`sub_1DA10`、跳表 17 / 25  / | `41-jimmy-neworder-gem-ztats.md`, `50-hole-up-camp-sleep-repair.md`, `51-closing-four-known-gaps.md`, `81-three-mode-loops-are-mutually-exclusive.md`, `87-bgm-selection-and-the-tbl-tables.md` |
| `sub_16DA4` | 上馬 / 上毯 / 上小艇都要求**先下來走路**(`sub_16DA4` 判 `byte_3E08C` ∈ {0x1C, 0x1D}) | `11-map-objects.md`, `53-party-tile-and-the-arena-floor.md` |
| `sub_16DC8` | 上大船的限制不同(`sub_16DC8` 的跳表):放行魔毯、步行、小艇 —— | `11-map-objects.md` |
| `sub_16E58` | `sub_16E58` 是「附近有沒有陸地」:查視窗的 (4,5)(6,5)(5,4)(5,6),也就是玩家四鄰 | `11-map-objects.md` |
| `sub_16F08` | 上船唸一次 `Danger! Ship badly damaged!`(`sub_16F08`,`cmp eax, 0Ah`), | `11-map-objects.md`, `41-jimmy-neworder-gem-ztats.md`, `47-move-modes-and-time-of-day.md`, `60-command-echo-and-menu-keys.md`, `66-hexrays-truncation-audit.md`, `87-bgm-selection-and-the-tbl-tables.md`, `88-the-load-window-culling-and-ship-cannon.md`, `98-select-menu-and-song-melody-matching.md` |
| `sub_17120` | IDA 位址:`sub_172C4`(Fire)、`sub_17120`(舷側判定)、`sub_18704` / `sub_18698`(Mix) | `45-fire-and-mix.md`, `88-the-load-window-culling-and-ship-cannon.md` |
| `sub_172C4` | 主要函式  /  `sub_172C4`(F 開砲,**244 行**)、`sub_2B3DC`(查某一格上的物件)、`sub_2A4D0`(全隊傷害)、`sub_2E0`/`sub_218`/`sub_268`(NPC… | `41-jimmy-neworder-gem-ztats.md`, `42-equipment-slots-ready-wear.md`, `45-fire-and-mix.md`, `69-the-overworld-cannon-did-nothing.md` |
| `sub_17630` | 點一把火把是 `random(0,15) + 0x70`(`sub_17630`),也就是 112..127 分鐘。 | `17-magic.md`, `41-jimmy-neworder-gem-ztats.md` |
| `sub_17688` | IDA 位址:`sub_2ACF4`(A..Z 派發)、`sub_14CAC`(J)、`sub_17688`(N)、 | `41-jimmy-neworder-gem-ztats.md` |
| `sub_177AC` | `sub_177AC` 下載具 → 字面 1 ×2  /  —  /  ⬜ **刻意不接**:兩個呼叫點還沒逐一讀,而 `dismount` 在**場景裡**也會被呼叫(在城裡下馬)⇒ 接錯會在城裡放大地圖的曲子。有位址… | `11-map-objects.md`, `41-jimmy-neworder-gem-ztats.md`, `47-move-modes-and-time-of-day.md`, `87-bgm-selection-and-the-tbl-tables.md` |
| `sub_17A14` | if (byte_3E0A3 >= 1 && byte_3E0A3 <= 0x20) return sub_17A14(buf);   // 城裡:暗影君主 | `26-yell-words-of-power-shadowlords.md`, `28-shadowlords-and-blackthorn.md` |
| `sub_17C2C` | 它呼叫 `sub_17C2C(edi, …)` 時,那一支又用同一個 `edi` 去查 `off_411BC`(美德名)、 | `25-shrines.md`, `26-yell-words-of-power-shadowlords.md`, `28-shadowlords-and-blackthorn.md` |
| `sub_17CFC` | if (byte_3E0A3 == 0)                       return sub_17CFC(buf);   // 大地圖:力量之言 | `26-yell-words-of-power-shadowlords.md` |
| `sub_17E74` | 位址:`sub_17E74`(分派)、`sub_17CFC`(力量之言)、`sub_17C2C`(復原聖壇)、 | `26-yell-words-of-power-shadowlords.md`, `41-jimmy-neworder-gem-ztats.md` |
| `sub_17F58` | `sub_17F58`(可推清單)、`sub_60BC`(主選單)、`sub_6730`(選單迴圈) | `40-push-and-the-main-menu.md` |
| `sub_18028` | IDA 位址:`sub_18154`(Push)、`sub_1806C` / `sub_180E0`(搬運)、`sub_18028`(轉向)、 | `40-push-and-the-main-menu.md` |
| `sub_1806C` | IDA 位址:`sub_18154`(Push)、`sub_1806C` / `sub_180E0`(搬運)、`sub_18028`(轉向)、 | `40-push-and-the-main-menu.md` |
| `sub_180E0` | IDA 位址:`sub_18154`(Push)、`sub_1806C` / `sub_180E0`(搬運)、`sub_18028`(轉向)、 | `40-push-and-the-main-menu.md` |
| `sub_18154` | IDA 位址:`sub_18154`(Push)、`sub_1806C` / `sub_180E0`(搬運)、`sub_18028`(轉向)、 | `40-push-and-the-main-menu.md`, `41-jimmy-neworder-gem-ztats.md` |
| `sub_18380` | 主要函式  /  `sub_18380`(ESC = 撤離)、`sub_A360`(戰鬥分派器)、`sub_42CC`(進地牢房間)、`sub_A9EC`(`byte_3E0B3` 的設定處)、`sub_2E364`(戰… | `73-escape-is-a-key-and-it-has-two-gates.md` |
| `sub_18468` | 原版讓玩家**自己挑藥草**(`sub_18468`,方向鍵 + RETURN,「Type M to mix」), | `45-fire-and-mix.md`, `46-command-menus.md`, `58-rune-input-cast-and-mix.md` |
| `sub_18698` | IDA 位址:`sub_172C4`(Fire)、`sub_17120`(舷側判定)、`sub_18704` / `sub_18698`(Mix) | `45-fire-and-mix.md`, `58-rune-input-cast-and-mix.md` |
| `sub_18704` | IDA 位址:`sub_172C4`(Fire)、`sub_17120`(舷側判定)、`sub_18704` / `sub_18698`(Mix) | `17-magic.md`, `41-jimmy-neworder-gem-ztats.md`, `42-equipment-slots-ready-wear.md`, `45-fire-and-mix.md`, `58-rune-input-cast-and-mix.md` |
| `sub_188C4` | K  /  Klimb  /  地表 `sub_188C4` / 場景 `sub_EA0` / 地牢 `sub_417C`  / | `41-jimmy-neworder-gem-ztats.md`, `49-command-table-and-two-empty-keys.md`, `52-one-coordinate-pair-one-tile-accessor.md`, `68-troll-bridge-and-collision.md` |
| `sub_189BC` | 兩位打不掉。這與 `docs/re/17` 的「`sub_189BC` 只認黑刺 / 不列顛王 / 暗影領主 | `17-magic.md`, `69-the-overworld-cannon-did-nothing.md` |
| `sub_189E4` | Grav Por / Vas Flam / Xen Corp  /  `sub_189E4(0x30/0x31/0x32)`  /  指定目標的攻擊咒語  / | `17-magic.md` |
| `sub_18A08` | 咒語放出來的力場**(`sub_18A08` 的 `byte_55E24` = `82 81 80 83`)。 | `17-magic.md`, `18-dungeons.md` |
| `sub_18AF0` | An Zu  /  `sub_18AF0`  /  狀態 `'S'` → `'G'`,戰場上清掉睡著旗標  / | `17-magic.md` |
| `sub_18B88` | An Nox  /  `sub_18B88`  /  狀態 `'P'` → `'G'`  / | `17-magic.md` |
| `sub_18C00` | 其餘的效果函式看得到入口(`sub_18C00`、`sub_18D18`、`sub_18EB0`、`sub_18F2C`、 | `17-magic.md` |
| `sub_18D18` | `0x70`  /  **開過的寶箱**  /  ⚠ 原本照「怪物走得過去」猜成門,`sub_18D18` 的 `(tile & 8) \ /  0x70` 推翻了(見 `docs/re/21`)  / | `17-magic.md`, `18-dungeons.md`, `21-chests-fields-locks.md`, `75-open-did-nothing-outside-dungeons.md` |
| `sub_18EB0` | 其餘的效果函式看得到入口(`sub_18C00`、`sub_18D18`、`sub_18EB0`、`sub_18F2C`、 | `17-magic.md` |
| `sub_18F1C` | In Wis  /  `sub_18F1C` → `sub_1D0C4`  /  報出所在座標  / | `17-magic.md` |
| `sub_18F2C` | 其餘的效果函式看得到入口(`sub_18C00`、`sub_18D18`、`sub_18EB0`、`sub_18F2C`、 | `17-magic.md` |
| `sub_1904C` | `sub_1904C`、`sub_19098`、`sub_1D1B8`、`sub_1D31C`、`sub_19264`、`sub_192BC`、 | `17-magic.md` |
| `sub_19098` | `sub_1904C`、`sub_19098`、`sub_1D1B8`、`sub_1D31C`、`sub_19264`、`sub_192BC`、 | `17-magic.md`, `88-the-load-window-culling-and-ship-cannon.md` |
| `sub_19264` | `sub_1904C`、`sub_19098`、`sub_1D1B8`、`sub_1D31C`、`sub_19264`、`sub_192BC`、 | `17-magic.md` |
| `sub_192BC` | `sub_1904C`、`sub_19098`、`sub_1D1B8`、`sub_1D31C`、`sub_19264`、`sub_192BC`、 | `17-magic.md` |
| `sub_19354` | An Ex Por  /  `sub_19354`  /  0xB8 / 0xB9 → **0x97**;0xBA / 0xBB → **0x98**  / | `17-magic.md` |
| `sub_193C8` | Vas Mani  /  `sub_193C8`  /  `HP = MaxHP`  / | `17-magic.md` |
| `sub_19440` | In Vas Por Ylem  /  `sub_19440`  /  對**每個**敵人擲 1..30 對防禦,擲贏吃 `random(1,20)`  / | `17-magic.md` |
| `sub_194CC` | An Xen Ex  /  `sub_194CC`  /  「Creature: X charmed!」  / | `17-magic.md` |
| `sub_195C0` | `sub_19440`、`sub_195C0`、`sub_19674`、`sub_196A4`、`sub_19810`、`sub_1986C`、 | `17-magic.md` |
| `sub_19674` | `sub_19440`、`sub_195C0`、`sub_19674`、`sub_196A4`、`sub_19810`、`sub_1986C`、 | `17-magic.md` |
| `sub_196A4` | `sub_19440`、`sub_195C0`、`sub_19674`、`sub_196A4`、`sub_19810`、`sub_1986C`、 | `17-magic.md` |
| `sub_19810` | `sub_19440`、`sub_195C0`、`sub_19674`、`sub_196A4`、`sub_19810`、`sub_1986C`、 | `17-magic.md` |
| `sub_1986C` | `sub_19440`、`sub_195C0`、`sub_19674`、`sub_196A4`、`sub_19810`、`sub_1986C`、 | `17-magic.md`, `22-moongates.md`, `88-the-load-window-culling-and-ship-cannon.md` |
| `sub_198E0` | An Tym  /  `sub_198E0`  /  `byte_3E08A = 'T'`、`byte_3E09E = 10`  / | `17-magic.md` |
| `sub_1994C` | 主要函式  /  `sub_3FE4`(離開地牢)、`sub_3F34`(換層 + 邊界)、`sub_417C`(Klimb)、`sub_4074` / `sub_100F8` / `sub_1994C`(三個呼叫端) … | `17-magic.md`, `58-rune-input-cast-and-mix.md`, `74-two-swamp-dice-and-an-unnamed-failure.md`, `76-jimmy-in-a-dungeon-disarms-not-unlocks.md`, `78-the-bottom-of-a-dungeon-opens-into-the-underworld.md` |
| `sub_19ED8` | 主要函式  /  `sub_1E8D4`(建清單)、`sub_1A5E8`(U 的分派)、`sub_19ED8`(卷軸,201 行)、`sub_1A0B0`(藥水,254 行)、`sub_1A2F8`(月石)、`sub_… | `71-the-use-list-had-29-empty-slots.md`, `77-a-false-gap-and-the-failed-line.md` |
| `sub_1A0B0` | 主要函式  /  `sub_1E8D4`(建清單)、`sub_1A5E8`(U 的分派)、`sub_19ED8`(卷軸,201 行)、`sub_1A0B0`(藥水,254 行)、`sub_1A2F8`(月石)、`sub_… | `71-the-use-list-had-29-empty-slots.md`, `77-a-false-gap-and-the-failed-line.md` |
| `sub_1A2F8` | 主要函式  /  `sub_1E8D4`(建清單)、`sub_1A5E8`(U 的分派)、`sub_19ED8`(卷軸,201 行)、`sub_1A0B0`(藥水,254 行)、`sub_1A2F8`(月石)、`sub_… | `71-the-use-list-had-29-empty-slots.md`, `77-a-false-gap-and-the-failed-line.md` |
| `sub_1A38C` | 主要函式  /  `sub_1A38C`(用碎片)、`sub_2D9D0`(風浪)、`sub_2A1E8` / `sub_2A0C4` / `sub_2A984`(狀態列)  / | `28-shadowlords-and-blackthorn.md`, `29-npc-behaviour-and-arrest.md`, `33-get-command.md`, `80-implemented-but-unreachable.md` |
| `sub_1A5B0` | IDA 位址:`sub_1A5E8`(Use)、`sub_1E8D4`(可用道具清單)、`sub_1A5B0`(信物切換)、 | `44-use-item.md` |
| `sub_1A5E8` | ✅ 早就從組語逆過,截斷沒造成損失  /  `sub_21108` 酒單(`tavern.go`)、`sub_1D394` 聖壇與 `ALAKAZAM`(`shrine.go`)、`sub_14CAC`/`sub_14B… | `28-shadowlords-and-blackthorn.md`, `41-jimmy-neworder-gem-ztats.md`, `42-equipment-slots-ready-wear.md`, `44-use-item.md`, `56-two-sets-of-item-names.md`, `66-hexrays-truncation-audit.md`, `71-the-use-list-had-29-empty-slots.md`, `77-a-false-gap-and-the-failed-line.md`, `79-four-dialogue-opcodes-were-corrupting-the-text.md`, `92-white-noise-shim-and-ambient-sound.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_1AC20` | ⚠ **範圍的形狀還沒逆完**。`sub_1AC20` 吃一個每個咒語各自不同的參數 | `20-projectiles.md` |
| `sub_1AEB4` | `sub_1AEB4` / `sub_1C8E8` / `sub_1FA6C`  /  0x320, 變數, 0x2BC  /  **MAHOU1(音高隨參數)**  / | `17-magic.md`, `20-projectiles.md`, `92-white-noise-shim-and-ambient-sound.md` |
| `sub_1B140` | 主要函式  /  `sub_1B3D0`(盤查)、`sub_1B52C`(對話分派)、`sub_195C`(碰到 NPC)、`sub_1884`(逮捕)、`sub_1B140`(密語比對)、`sub_154C` → `s… | `32-guard-challenge.md` |
| `sub_1B18C` | (2) 隔著櫃檯談得到。** 對面那一格沒人時,如果地形在 `sub_1B18C` 的清單裡 | `35-harp-and-the-secret-door.md` |
| `sub_1B294` | 主要函式  /  `sub_216C8`(酒館主流程)、`sub_210D8`(點餐)、★ `sub_20E6C`(算總價)、★★ `sub_20F60`(報價 + 收錢 + **上菜**)、★ `sub_20ED0`(… | `08-shops.md`, `10-shop-prices-and-trade.md`, `93-the-tavern-meal-was-never-haggled.md` |
| `sub_1B3D0` | ✅ 早就從組語逆過,截斷沒造成損失  /  `sub_21108` 酒單(`tavern.go`)、`sub_1D394` 聖壇與 `ALAKAZAM`(`shrine.go`)、`sub_14CAC`/`sub_14B… | `32-guard-challenge.md`, `35-harp-and-the-secret-door.md`, `66-hexrays-truncation-audit.md` |
| `sub_1B52C` | 主要函式  /  `sub_1B3D0`(盤查)、`sub_1B52C`(對話分派)、`sub_195C`(碰到 NPC)、`sub_1884`(逮捕)、`sub_1B140`(密語比對)、`sub_154C` → `s… | `04-npc-schedule-and-clock.md`, `06-conversation-script.md`, `08-shops.md`, `29-npc-behaviour-and-arrest.md`, `32-guard-challenge.md`, `35-harp-and-the-secret-door.md` |
| `sub_1B658` | T  /  Talk  /  地表 `sub_2B2AC`;地點 ≤ 0x20 → `Talk-Funny, no response!`;其餘 `sub_1B658`  / | `35-harp-and-the-secret-door.md`, `41-jimmy-neworder-gem-ztats.md`, `49-command-table-and-two-empty-keys.md` |
| `sub_1B760` | 0xFF  /  結束整段對話  /  `sub_1B760` + `sub_1BF08`  / | `06-conversation-script.md` |
| `sub_1B800` | ⚠ `offset sub_1B800` 是 **IDA 的自動命名,不是函式** —— 它被當純數值用。 | `05-text-compression.md`, `86-tile-0xDC-is-an-open-moongate.md` |
| `sub_1B854` | 主要函式  /  ★ `sub_2BCC8`(再生戒指)、`sub_2ECE8`(戒指的每回合效果)、`sub_1F4E4`(力場消散)、`sub_2A50C` 尾段、`sub_16370`、★ `sub_FEC` + … | `79-four-dialogue-opcodes-were-corrupting-the-text.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_1B964` | 主要函式  /  `sub_1C1E8`(四個 opcode 的分派)、`sub_1B854`(索取金幣)、`sub_1B964`(給東西)、`sub_1C1C8`(認得旗標)、`sub_2BBB8`/`sub_2BBD… | `79-four-dialogue-opcodes-were-corrupting-the-text.md` |
| `sub_1BA80` | sub_1BA80(0x90, 0x9F)                ; 前進到下一個 0x90..0x9F 標記 | `06-conversation-script.md`, `79-four-dialogue-opcodes-were-corrupting-the-text.md` |
| `sub_1BAA4` | 2. 把腳本指標**重設到記錄開頭**(`sub_1BAA4(0)`),讀 3 個位元組 —— 那是 NPC 名字的前三個字母。 | `06-conversation-script.md` |
| `sub_1BAFC` | `sub_1BF08` 用 `sub_1BAFC(i*2 + 6)` 印回應。 | `06-conversation-script.md` |
| `sub_1BB3C` | `sub_1BB3C`(插名字)、`sub_1BB5C`(加入隊伍)、`sub_C10`(叫衛兵)、 | `06-conversation-script.md` |
| `sub_1BB5C` | 0x1F  /  留在旅店的標記  /  隊員 0x00、其餘 0xFF  /  入隊時 `sub_1BB5C` 寫 `byte_3DDD3[i*32]` = 0  / | `06-conversation-script.md`, `07-save-format.md`, `42-equipment-slots-ready-wear.md` |
| `sub_1BCB8` | (`sub_1C0AC` → `sub_1BCB8` 從記錄開頭找碼相同的區塊)。玩家回答後: | `06-conversation-script.md`, `79-four-dialogue-opcodes-were-corrupting-the-text.md` |
| `sub_1BCF4` | 輸入含「是」的觸發字 → 印第 4 段(`sub_1BCF4`),否則印第 2 段(`sub_1BD0C`)。 | `06-conversation-script.md` |
| `sub_1BD0C` | 輸入含「是」的觸發字 → 印第 4 段(`sub_1BCF4`),否則印第 2 段(`sub_1BD0C`)。 | `06-conversation-script.md`, `79-four-dialogue-opcodes-were-corrupting-the-text.md` |
| `sub_1BD50` | 算式不是猜的:`sub_1BD50(i)` 把指標重設到開頭後跳 `2i+5` 段,命中之後 | `06-conversation-script.md` |
| `sub_1BD8C` | (位置 0,或前一個字元是空白)——`sub_1BD8C` → `sub_27C98`,再檢查 `byte_55F37[i]`。 | `06-conversation-script.md` |
| `sub_1BE28` | 索引  /  字  /  行為(`sub_1BE28`)  / | `06-conversation-script.md` |
| `sub_1BF08` | jz  → sub_1BF08()            ; ★ 0xFF = 不跳,回到「Your interest?」 | `06-conversation-script.md`, `79-four-dialogue-opcodes-were-corrupting-the-text.md` |
| `sub_1C00C` | sub_1BCB8 / sub_1BD0C / sub_1C00C: | `79-four-dialogue-opcodes-were-corrupting-the-text.md` |
| `sub_1C0AC` | 0x91–0x9F  /  反問玩家並讀取回答(15 種)  /  `byte_55F32 = 碼; sub_1C0AC` → 印 "You respond-\n:" 後收輸入  / | `06-conversation-script.md`, `79-four-dialogue-opcodes-were-corrupting-the-text.md` |
| `sub_1C1AC` | dword_3E3E8[地點]  / = 1 << i;                // sub_1C1AC:這座城的這個人認得汝了 | `06-conversation-script.md` |
| `sub_1C1C8` | 主要函式  /  `sub_1C1E8`(四個 opcode 的分派)、`sub_1B854`(索取金幣)、`sub_1B964`(給東西)、`sub_1C1C8`(認得旗標)、`sub_2BBB8`/`sub_2BBD… | `79-four-dialogue-opcodes-were-corrupting-the-text.md` |
| `sub_1C1E8` | 主要函式  /  `sub_1C1E8`(四個 opcode 的分派)、`sub_1B854`(索取金幣)、`sub_1B964`(給東西)、`sub_1C1C8`(認得旗標)、`sub_2BBB8`/`sub_2BBD… | `06-conversation-script.md`, `79-four-dialogue-opcodes-were-corrupting-the-text.md` |
| `sub_1C2FC` | 0x88  /  **反問「汝名為何?」**  /  `sub_1C2FC`,見下  / | `06-conversation-script.md` |
| `sub_1C3F8` | 來源:FM Towns `WORRIORS.EXP` 的 `sub_1C3F8`(展開)與 `dword_41990`(槽表); | `05-text-compression.md`, `06-conversation-script.md` |
| `sub_1C840` | 帶一張 31 路跳表;周邊有 `sub_1C840`(載入記錄)、`sub_1B52C`(分派)、 | `05-text-compression.md`, `06-conversation-script.md`, `32-guard-challenge.md` |
| `sub_1C8E8` | `sub_1AEB4` / `sub_1C8E8` / `sub_1FA6C`  /  0x320, 變數, 0x2BC  /  **MAHOU1(音高隨參數)**  / | `71-the-use-list-had-29-empty-slots.md`, `92-white-noise-shim-and-ambient-sound.md` |
| `sub_1C9C0` | 主要函式  /  `sub_1E8D4`(建清單)、`sub_1A5E8`(U 的分派)、`sub_19ED8`(卷軸,201 行)、`sub_1A0B0`(藥水,254 行)、`sub_1A2F8`(月石)、`sub_… | `71-the-use-list-had-29-empty-slots.md` |
| `sub_1CA0C` | `sub_1CA0C` 是施法與調藥**共用**的輸入常式(`sub_1994C+5C` 與 `sub_18704+3E` 兩處呼叫): | `17-magic.md`, `58-rune-input-cast-and-mix.md` |
| `sub_1CC50` | 1 RH  "Wind change!"    sub_1CC50() 問方向;byte_3E0A3 >= 21h → ebx = 0 | `17-magic.md`, `20-projectiles.md`, `21-chests-fields-locks.md`, `71-the-use-list-had-29-empty-slots.md` |
| `sub_1CD3C` | Mani  /  `sub_1CD3C`  /  回 **1..30**(與命中骰同一顆 `sub_2B724`),上限 MaxHP,死人無效  / | `17-magic.md`, `71-the-use-list-had-29-empty-slots.md` |
| `sub_1CDA4` | 主要函式  /  `sub_2CCFC`(★ 真正的航行閘門)、`sub_2D174`(移動分派)、`sub_2D014`(改座標)、`sub_2D2D0`(揚帆的節奏)、`sub_2B24`(★ 怪物移動)、`sub_… | `23-wind-and-sailing.md`, `84-the-sailing-gate-and-the-wind-table.md` |
| `sub_1CE0C` | `sub_1CE0C` 做的是 `sub_2E0E8(-1, 0, 0)` —— **把視線遮蔽罩整個填成 0xFF**, | `17-magic.md`, `71-the-use-list-had-29-empty-slots.md` |
| `sub_1CE70` | 5 KXC "Summon Daemon!"  <= 7Fh → "Not here!";  else sub_1CE70(1) | `17-magic.md`, `71-the-use-list-had-29-empty-slots.md` |
| `sub_1CFC8` | 6 IMC "Resurrection!"   >= 80h → "Not here!";  else sub_1C9C0() + sub_1CFC8() | `17-magic.md`, `71-the-use-list-had-29-empty-slots.md` |
| `sub_1D0C4` | In Wis  /  `sub_18F1C` → `sub_1D0C4`  /  報出所在座標  / | `17-magic.md` |
| `sub_1D15C` | In Ex Por  /  `sub_1D15C`  /  0x97 → 0xB8;0x98 → 0xBA  / | `17-magic.md`, `21-chests-fields-locks.md` |
| `sub_1D1B8` | `sub_1904C`、`sub_19098`、`sub_1D1B8`、`sub_1D31C`、`sub_19264`、`sub_192BC`、 | `17-magic.md`, `21-chests-fields-locks.md` |
| `sub_1D310` | 主要函式  /  `sub_1E8D4`(建清單)、`sub_1A5E8`(U 的分派)、`sub_19ED8`(卷軸,201 行)、`sub_1A0B0`(藥水,254 行)、`sub_1A2F8`(月石)、`sub_… | `17-magic.md`, `71-the-use-list-had-29-empty-slots.md` |
| `sub_1D31C` | 主要函式  /  `sub_1E8D4`(建清單)、`sub_1A5E8`(U 的分派)、`sub_19ED8`(卷軸,201 行)、`sub_1A0B0`(藥水,254 行)、`sub_1A2F8`(月石)、`sub_… | `17-magic.md`, `44-use-item.md`, `71-the-use-list-had-29-empty-slots.md` |
| `sub_1D340` | 掃到就印「Thou dost see an urn marked: <名字>」並用 `sub_1D340` 擺出罈子。 | `27-codex-and-the-shrine-chamber.md`, `28-shadowlords-and-blackthorn.md` |
| `sub_1D394` | ✅ 早就從組語逆過,截斷沒造成損失  /  `sub_21108` 酒單(`tavern.go`)、`sub_1D394` 聖壇與 `ALAKAZAM`(`shrine.go`)、`sub_14CAC`/`sub_14B… | `25-shrines.md`, `26-yell-words-of-power-shadowlords.md`, `27-codex-and-the-shrine-chamber.md`, `66-hexrays-truncation-audit.md` |
| `sub_1D850` | 0x0328**  /  2  /  `byte_3E0DE`  /  已在寶典上讀到的美德(**寶典 `sub_1D850` 設的,不是聖壇**)  / | `25-shrines.md`, `26-yell-words-of-power-shadowlords.md`, `27-codex-and-the-shrine-chamber.md` |
| `sub_1DA10` | `0x19`(25)  /  **the shrine of …**(八德聖壇;名字取自 `off_411BC[9]`,狀態看 `byte_411FC[8]`/`byte_41204[8]`)  /  `sub_1DA1… | `03-picture-files.md`, `03-scene-entry-and-tile-semantics.md`, `100-the-last-three-and-the-collapsed-entrance-gate.md`, `25-shrines.md`, `27-codex-and-the-shrine-chamber.md`, `28-shadowlords-and-blackthorn.md`, `34-ending-trigger.md`, `87-bgm-selection-and-the-tbl-tables.md` |
| `sub_1DD00` | Q  /  Quit & save  /  `sub_1DD00`  /  ✅  / | `41-jimmy-neworder-gem-ztats.md` |
| `sub_1DE10` | sub_1F3A4:  sub_1DE10(0)                                → 選人(< 0 取消) | `60-command-echo-and-menu-keys.md` |
| `sub_1DECC` | 主要函式  /  ★★ `sub_1E9A0`(頁面游標與按鍵)、★ `sub_1DECC`(隊員數值頁)、`sub_1E128`(隊員裝備頁)、★ `sub_1E210`(Equipment)、`sub_1E588`(… | `94-ztats-has-seventeen-pages.md` |
| `sub_1E0A4` | `sub_1E0A4(1)` / `sub_28F80` / `sub_290B4` 是畫框與標題列,引擎用自己的文字面板代替。 | `94-ztats-has-seventeen-pages.md` |
| `sub_1E0F0` | 每格丟給 `sub_1E0F0` 印名字並累加回傳值;**六格加起來是 0** 才印 `"(None ready)"`。 | `94-ztats-has-seventeen-pages.md` |
| `sub_1E128` | 主要函式  /  ★★ `sub_1E9A0`(頁面游標與按鍵)、★ `sub_1DECC`(隊員數值頁)、`sub_1E128`(隊員裝備頁)、★ `sub_1E210`(Equipment)、`sub_1E588`(… | `94-ztats-has-seventeen-pages.md` |
| `sub_1E210` | 主要函式  /  ★★ `sub_1E9A0`(頁面游標與按鍵)、★ `sub_1DECC`(隊員數值頁)、`sub_1E128`(隊員裝備頁)、★ `sub_1E210`(Equipment)、`sub_1E588`(… | `94-ztats-has-seventeen-pages.md` |
| `sub_1E38C` | 主要函式  /  `sub_1EC34`(Ready,**333 行**)、`sub_1E38C`(已經穿著嗎)、`sub_1EBE8`(空手狀況)、`sub_C414+1D0`(開戰佈陣)、`sub_A9EC`(`by… | `60-command-echo-and-menu-keys.md`, `72-ready-had-seven-missing-rules.md` |
| `sub_1E3D8` | if ( sub_1E3D8(a1, …) != -1 ) v13 = 2;                    // 上面還有 | `60-command-echo-and-menu-keys.md` |
| `sub_1E418` | 位址  /  `sub_2ACF4`(59 case 主分派器)、`sub_1F3A4`(Ready)、`sub_1E418`(找下一個非空欄位)  / | `42-equipment-slots-ready-wear.md`, `44-use-item.md`, `46-command-menus.md`, `49-command-table-and-two-empty-keys.md`, `60-command-echo-and-menu-keys.md` |
| `sub_1E588` | 主要函式  /  ★★ `sub_1E9A0`(頁面游標與按鍵)、★ `sub_1DECC`(隊員數值頁)、`sub_1E128`(隊員裝備頁)、★ `sub_1E210`(Equipment)、`sub_1E588`(… | `94-ztats-has-seventeen-pages.md` |
| `sub_1E8D4` | 主要函式  /  `sub_1E8D4`(建清單)、`sub_1A5E8`(U 的分派)、`sub_19ED8`(卷軸,201 行)、`sub_1A0B0`(藥水,254 行)、`sub_1A2F8`(月石)、`sub_… | `44-use-item.md`, `71-the-use-list-had-29-empty-slots.md`, `79-four-dialogue-opcodes-were-corrupting-the-text.md`, `94-ztats-has-seventeen-pages.md` |
| `sub_1E9A0` | 主要函式  /  ★★ `sub_1E9A0`(頁面游標與按鍵)、★ `sub_1DECC`(隊員數值頁)、`sub_1E128`(隊員裝備頁)、★ `sub_1E210`(Equipment)、`sub_1E588`(… | `17-magic.md`, `41-jimmy-neworder-gem-ztats.md`, `94-ztats-has-seventeen-pages.md` |
| `sub_1EBE8` | 主要函式  /  `sub_1EC34`(Ready,**333 行**)、`sub_1E38C`(已經穿著嗎)、`sub_1EBE8`(空手狀況)、`sub_C414+1D0`(開戰佈陣)、`sub_A9EC`(`by… | `72-ready-had-seven-missing-rules.md` |
| `sub_1EC34` | 主要函式  /  `sub_1EC34`(Ready,**333 行**)、`sub_1E38C`(已經穿著嗎)、`sub_1EBE8`(空手狀況)、`sub_C414+1D0`(開戰佈陣)、`sub_A9EC`(`by… | `72-ready-had-seven-missing-rules.md`, `74-two-swamp-dice-and-an-unnamed-failure.md` |
| `sub_1EFC8` | `sub_1EFC8`(清單瀏覽器)、`sub_CAC` / `sub_4074` / `sub_2D478`(Attack 三支)  / | `60-command-echo-and-menu-keys.md`, `70-hunger-poison-and-the-vanishing-rings.md`, `71-the-use-list-had-29-empty-slots.md` |
| `sub_1F3A4` | 位址  /  `sub_2ACF4`(59 case 主分派器)、`sub_1F3A4`(Ready)、`sub_1E418`(找下一個非空欄位)  / | `41-jimmy-neworder-gem-ztats.md`, `42-equipment-slots-ready-wear.md`, `46-command-menus.md`, `49-command-table-and-two-empty-keys.md`, `60-command-echo-and-menu-keys.md` |
| `sub_1F48C` | 六個操控類咒語每一個都是 `!sub_1F48C(…) && !sub_189BC(…)` 這一對。 | `17-magic.md` |
| `sub_1F4E4` | 主要函式  /  ★ `sub_2BCC8`(再生戒指)、`sub_2ECE8`(戒指的每回合效果)、`sub_1F4E4`(力場消散)、`sub_2A50C` 尾段、`sub_16370`、★ `sub_FEC` + … | `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_1F528` | 主要函式  /  `sub_10BDC`(踏進沼澤的中毒)、`sub_1318`(每回合的地形效果)、`sub_2D9D0`(移動後的分派)、`sub_1F570`(落空的訊息)、`sub_1F528`(印單位名字)、`… | `74-two-swamp-dice-and-an-unnamed-failure.md` |
| `sub_1F570` | 主要函式  /  `sub_10BDC`(踏進沼澤的中毒)、`sub_1318`(每回合的地形效果)、`sub_2D9D0`(移動後的分派)、`sub_1F570`(落空的訊息)、`sub_1F528`(印單位名字)、`… | `74-two-swamp-dice-and-an-unnamed-failure.md` |
| `sub_1F5A4` | `sub_1F5A4`:魅惑(0x0040)與另外兩個位元(0x0400 / 0x0800)的遠程特殊行為。 | `16-combat-turns-and-ai.md` |
| `sub_1F840` | 主要函式  /  `sub_A360`(隊員的戰鬥回合,**558 行**)、`sub_BCC4`(掙脫)、`sub_1F840`(命中結果)、`sub_2B724`/`sub_2B710`(門檻骰)、`sub_2ED5… | `67-corpser-and-the-sleeping-party-member.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_1F9CC` | 距離是 `sub_1F9F8`:`sub_1F9CC` 算出 dx²+dy² 之後用「連續減奇數」 | `16-combat-turns-and-ai.md` |
| `sub_1F9F8` | 距離是 `sub_1F9F8`:`sub_1F9CC` 算出 dx²+dy² 之後用「連續減奇數」 | `16-combat-turns-and-ai.md` |
| `sub_1FA6C` | `sub_1AEB4` / `sub_1C8E8` / `sub_1FA6C`  /  0x320, 變數, 0x2BC  /  **MAHOU1(音高隨參數)**  / | `15-combat-formulas.md`, `92-white-noise-shim-and-ambient-sound.md` |
| `sub_1FD80` | ├ sub_1FD80   選中目標 | `15-combat-formulas.md` |
| `sub_1FDE8` | (`sub_1FDE8`)。所以失手的箭不會憑空消失,它會偏掉繼續飛。 | `20-projectiles.md` |
| `sub_1FE54` | `sub_1FE54` / `sub_20CB4` 的投射物飛行路徑(會不會被地形擋)。 | `16-combat-turns-and-ai.md`, `17-magic.md`, `20-projectiles.md` |
| `sub_200BC` | └ sub_200BC   ★ 有沒有人在旁邊干擾 | `17-magic.md` |
| `sub_20134` | 它們走 `sub_20360` → `sub_20134`,那條路是**射程 15 的指定目標攻擊**, | `15-combat-formulas.md`, `17-magic.md` |
| `sub_20360` | 戰鬥中的力場**:`sub_20360(單位, byte_55E20[種類])`,效果碼 0x33..0x36, | `17-magic.md` |
| `sub_2055C` | 1. `sub_2055C` 在 **×16 的像素空間**拉 Bresenham(每格再 +16 對到格心)。 | `20-projectiles.md` |
| `sub_20678` | └ sub_20678   畫這一幀 | `20-projectiles.md` |
| `sub_20CB4` | 命中 = sub_20CB4(dx+5, dy+5, 5, 5, 3)                      ; 投射物飛行 | `16-combat-turns-and-ai.md`, `20-projectiles.md`, `69-the-overworld-cannon-did-nothing.md`, `83-monster-attacks-and-two-different-wind-rules.md`, `88-the-load-window-culling-and-ship-cannon.md` |
| `sub_20E6C` | 主要函式  /  `sub_216C8`(酒館主流程)、`sub_210D8`(點餐)、★ `sub_20E6C`(算總價)、★★ `sub_20F60`(報價 + 收錢 + **上菜**)、★ `sub_20ED0`(… | `10-shop-prices-and-trade.md`, `93-the-tavern-meal-was-never-haggled.md` |
| `sub_20ED0` | 主要函式  /  `sub_216C8`(酒館主流程)、`sub_210D8`(點餐)、★ `sub_20E6C`(算總價)、★★ `sub_20F60`(報價 + 收錢 + **上菜**)、★ `sub_20ED0`(… | `48-dungeon-wandering-monster-and-arena.md`, `93-the-tavern-meal-was-never-haggled.md` |
| `sub_20F24` | 主要函式  /  `sub_216C8`(酒館主流程)、`sub_210D8`(點餐)、★ `sub_20E6C`(算總價)、★★ `sub_20F60`(報價 + 收錢 + **上菜**)、★ `sub_20ED0`(… | `93-the-tavern-meal-was-never-haggled.md`, `95-the-dungeon-has-its-own-look.md` |
| `sub_20F60` | 主要函式  /  `sub_216C8`(酒館主流程)、`sub_210D8`(點餐)、★ `sub_20E6C`(算總價)、★★ `sub_20F60`(報價 + 收錢 + **上菜**)、★ `sub_20ED0`(… | `10-shop-prices-and-trade.md`, `93-the-tavern-meal-was-never-haggled.md` |
| `sub_210D8` | 主要函式  /  `sub_216C8`(酒館主流程)、`sub_210D8`(點餐)、★ `sub_20E6C`(算總價)、★★ `sub_20F60`(報價 + 收錢 + **上菜**)、★ `sub_20ED0`(… | `10-shop-prices-and-trade.md`, `93-the-tavern-meal-was-never-haggled.md` |
| `sub_21108` | ✅ 早就從組語逆過,截斷沒造成損失  /  `sub_21108` 酒單(`tavern.go`)、`sub_1D394` 聖壇與 `ALAKAZAM`(`shrine.go`)、`sub_14CAC`/`sub_14B… | `10-shop-prices-and-trade.md`, `66-hexrays-truncation-audit.md`, `70-hunger-poison-and-the-vanishing-rings.md`, `93-the-tavern-meal-was-never-haggled.md` |
| `sub_21310` | 只命中九支:`sub_112F8` `sub_11588` `sub_118CC` `sub_11AF0` `sub_21310` | `10-shop-prices-and-trade.md`, `93-the-tavern-meal-was-never-haggled.md` |
| `sub_21500` | 位址  /  `sub_21500`(酒館打聽)、`sub_27C98`(關鍵字比對)、`sub_11168`(印 `SHOPPE.DAT`)  / | `10-shop-prices-and-trade.md`, `61-tavern-lore-and-the-transfer-second-stage.md` |
| `sub_216C8` | 主要函式  /  `sub_216C8`(酒館主流程)、`sub_210D8`(點餐)、★ `sub_20E6C`(算總價)、★★ `sub_20F60`(報價 + 收錢 + **上菜**)、★ `sub_20ED0`(… | `10-shop-prices-and-trade.md`, `93-the-tavern-meal-was-never-haggled.md` |
| `sub_218DC` | 實際讀 `sub_218DC`(造船廠成交): | `11-map-objects.md` |
| `sub_219B0` | 買(`sub_11AF0`、`sub_11588`、`sub_112F8`、`sub_118CC`、`sub_219B0` 都一樣) | `10-shop-prices-and-trade.md`, `93-the-tavern-meal-was-never-haggled.md` |
| `sub_21C40` | 造船廠  /  0x84  /  `sub_21C40` → `sub_219B0`  /  4  /  ✅ 報價收錢(地圖物件待補)  / | `10-shop-prices-and-trade.md` |
| `sub_21CC0` | `sub_2BCC8` 之外的 `sub_21CC0` / `sub_22234` / `sub_22254`  /  名冊游標:數 `CharInnFlag == 0` 的人數、往前 / 往後找下一個在隊伍裡的人  /… | `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_21CE4` | 客滿判斷在 `sub_21CE4`:寄放的人數不能超過 `byte_57098[旅店]`(3/4/3/2/2/2 間房)。 | `10-shop-prices-and-trade.md` |
| `sub_21D48` | `sub_165C8` / `sub_21D48` 的 `sub_2BCC8`  /  `camp()` 每小時 / `SleepUntilMorning()` 每 9 分鐘  / | `10-shop-prices-and-trade.md`, `51-closing-four-known-gaps.md`, `87-bgm-selection-and-the-tbl-tables.md`, `93-the-tavern-meal-was-never-haggled.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_22018` | `L` Leave  /  `sub_22018`  /  `Haggle(每天價)`,退房時結算  /  同伴離隊,記下寄放地點  / | `10-shop-prices-and-trade.md`, `93-the-tavern-meal-was-never-haggled.md` |
| `sub_22234` | `sub_2BCC8` 之外的 `sub_21CC0` / `sub_22234` / `sub_22254`  /  名冊游標:數 `CharInnFlag == 0` 的人數、往前 / 往後找下一個在隊伍裡的人  /… | `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_22254` | `sub_2BCC8` 之外的 `sub_21CC0` / `sub_22234` / `sub_22254`  /  名冊游標:數 `CharInnFlag == 0` 的人數、往前 / 往後找下一個在隊伍裡的人  /… | `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_22280` | `P` Pick up  /  `sub_22280`  /  `Haggle(每天價) × 月數`  /  結清後歸隊  / | `10-shop-prices-and-trade.md`, `93-the-tavern-meal-was-never-haggled.md` |
| `sub_2274C` | 旅店  /  0x88  /  `sub_2274C`  /  6  /  ✅ 住宿 / 寄放 / 領回  / | `10-shop-prices-and-trade.md` |
| `sub_22B50` | `sub_22B50` / `sub_2BDE0` ×2  /  0x14, 0x3C, 0x2710  /  **TAKI2**  / | `92-white-noise-shim-and-ambient-sound.md` |
| `sub_22D00` | 主要函式  /  `sub_310E8`(★ 填 MSF 表)、`sub_30F8C`(★ 播一段 CD)、`sub_6730`(啟動 / 選單)、`sub_6128`(製作群字幕)、`sub_22D00`(石室)、`s… | `90-cdda-scenes-and-sound-effect-indices.md`, `92-white-noise-shim-and-ambient-sound.md`, `audio-pipeline.md` |
| `sub_23248` | IDA 位址:`.text:000235B6`(流程)、`sub_23274`(一題)、`sub_23248`(抽美德)、 | `39-character-creation.md` |
| `sub_23274` | IDA 位址:`.text:000235B6`(流程)、`sub_23274`(一題)、`sub_23248`(抽美德)、 | `39-character-creation.md` |
| `sub_239B4` | sub_239B4(buf, 12);                    // ★ 只讀 12 個字元 | `26-yell-words-of-power-shadowlords.md`, `61-tavern-lore-and-the-transfer-second-stage.md` |
| `sub_23A24` | ★★ 抓鉤那一行後面**沒有 `sub_23A24`** —— 它是有/無而不是數量。 | `94-ztats-has-seventeen-pages.md` |
| `sub_23C18` | else            return sub_23C18("Yes\n"); | `54-pass-vehicle-verbs-and-the-avatar-line.md`, `60-command-echo-and-menu-keys.md`, `61-tavern-lore-and-the-transfer-second-stage.md`, `65-the-wishing-well-easter-egg.md`, `67-corpser-and-the-sleeping-party-member.md`, `70-hunger-poison-and-the-vanishing-rings.md`, `77-a-false-gap-and-the-failed-line.md` |
| `sub_23D50` | ⬜ 還沒讀  /  `sub_A360`(**558 行、34 個字串**,最大的一筆)、`sub_D650`(`Wanted:` 通緝告示?)、`sub_74BC`、`sub_23D50`(`EGA*.TIL` 載入)… | `66-hexrays-truncation-audit.md` |
| `sub_24824` | (`sub_24824(16, 80, 367, 431)` 圈的正是地圖窗),畫完卡在 | `17-magic.md` |
| `sub_24A50` | push off_41BE0 → call sub_24A50    ← 載 .16(off_41BE0 = 表的第 16 項 "CREATE.16") | `01-tileset-and-dot16-loader.md` |
| `sub_24BC0` | push off_41BA0 → call sub_24BC0    ← 載 PROPORT.PCS | `01-tileset-and-dot16-loader.md` |
| `sub_253D4` | sub_253D4(5, 5, 5) | `86-tile-0xDC-is-an-open-moongate.md` |
| `sub_25DE4` | sub_25DE4(0DCh, 5, 5)                        ; 在視窗中心畫月門 | `86-tile-0xDC-is-an-open-moongate.md` |
| `sub_265F0` | 主要函式  /  `sub_DEE4`(開關月門)、`sub_DE74`(這一顆算不算在場)、`sub_E084`(玩家踏上)、`sub_2870`(怪物踏上)、`sub_29BEC`(★ 重畫時挑動畫格)、`sub_2… | `86-tile-0xDC-is-an-open-moongate.md` |
| `sub_27034` | `while (sub_27034() == 0xFFFF)` 等一個按鍵。是一個**阻塞的畫面**,不是持續狀態。 | `17-magic.md` |
| `sub_27230` | ⬜ `sub_27230(0x21)` 與收尾的 `sub_2C188(0x4B0, 0x7D0, 1, 0x28)` 是音效 / 延遲, | `26-yell-words-of-power-shadowlords.md`, `73-escape-is-a-key-and-it-has-two-gates.md`, `94-ztats-has-seventeen-pages.md` |
| `sub_27754` | `sub_3124`  /  寫 `dword_4FDD8/DC` 並包一對 `sub_27754`  /  疑為游標 / 中斷遮罩,不碰遊戲狀態  / | `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_277C0` | sub_277C0(1, 2);                               // 從文字列 2 開始 | `60-command-echo-and-menu-keys.md` |
| `sub_27A58` | call    sub_27A58               ; 等檔案就緒 | `11-map-objects.md` |
| `sub_27BBC` | if ( sub_27BBC() == 9 ) break;             // ★ 畫到文字列 9 就停 → 7 列 | `60-command-echo-and-menu-keys.md` |
| `sub_27C98` | 主要函式  /  `sub_CD28`(井)、`sub_D258`(Look 分派)、`sub_27C98`(字串比對)、`sub_2B770`(文字輸入)、`sub_2B6C8`(寫物件槽)、`sub_2B57C`(找… | `06-conversation-script.md`, `25-shrines.md`, `26-yell-words-of-power-shadowlords.md`, `28-shadowlords-and-blackthorn.md`, `32-guard-challenge.md`, `61-tavern-lore-and-the-transfer-second-stage.md`, `65-the-wishing-well-easter-egg.md` |
| `sub_27D24` | 主要函式  /  ★ `sub_2D72C`(Enter 的 tile 分派)、`sub_2D564`、`sub_27D24`(讀檔序列)、★ `sub_92C0` + `sub_89EC` + `sub_8BA0`(N… | `07-save-format.md`, `10-shop-prices-and-trade.md`, `100-the-last-three-and-the-collapsed-entrance-gate.md`, `13-save-writing.md`, `26-yell-words-of-power-shadowlords.md`, `29-npc-behaviour-and-arrest.md`, `34-ending-trigger.md`, `73-escape-is-a-key-and-it-has-two-gates.md`, `86-tile-0xDC-is-an-open-moongate.md`, `95-the-dungeon-has-its-own-look.md`, `97-set-active-player.md` |
| `sub_284CC` | `sub_27D24`(讀檔)與 `sub_284CC`(寫檔)都碰 `byte_3E08B`(`docs/re/95` 的全檔掃描: | `07-save-format.md`, `13-save-writing.md`, `26-yell-words-of-power-shadowlords.md`, `86-tile-0xDC-is-an-open-moongate.md`, `95-the-dungeon-has-its-own-look.md`, `97-set-active-player.md` |
| `sub_28E14` | `sub_1318` 自己那一段  /  `sub_1A54`(主迴圈,每回合)  /  `sub_28E14(0, 1Dh)` = **random(0, 29)**  / | `15-combat-formulas.md`, `16-combat-turns-and-ai.md`, `28-shadowlords-and-blackthorn.md`, `48-dungeon-wandering-monster-and-arena.md`, `66-hexrays-truncation-audit.md`, `67-corpser-and-the-sleeping-party-member.md`, `70-hunger-poison-and-the-vanishing-rings.md`, `71-the-use-list-had-29-empty-slots.md`, `74-two-swamp-dice-and-an-unnamed-failure.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_28F40` | `repe cmpsb`。也就是說,如果輸入層(`sub_2B770` / `sub_28F40`)沒有幫忙轉大寫, | `26-yell-words-of-power-shadowlords.md` |
| `sub_28F5C` | ⬜ `sub_28F5C("AMBFDTPRS", ch)` 找不到時回 −1,原版接著讀 `byte_40C34[-1]` | `72-ready-had-seven-missing-rules.md` |
| `sub_28F80` | `sub_1E0A4(1)` / `sub_28F80` / `sub_290B4` 是畫框與標題列,引擎用自己的文字面板代替。 | `60-command-echo-and-menu-keys.md`, `94-ztats-has-seventeen-pages.md` |
| `sub_29008` | case 1: sub_29008(25);   break;   // ↓ | `60-command-echo-and-menu-keys.md` |
| `sub_290B4` | `sub_1E0A4(1)` / `sub_28F80` / `sub_290B4` 是畫框與標題列,引擎用自己的文字面板代替。 | `94-ztats-has-seventeen-pages.md` |
| `sub_29304` | 主要函式  /  `sub_29D64`(重畫)、`sub_2E0E8`(罩子入口)、`sub_2DDB0`(flood fill)、`sub_2E1D0`(擋不擋)、`sub_2E8D0`(距離)、`sub_29304… | `04-npc-schedule-and-clock.md`, `07-save-format.md`, `10-shop-prices-and-trade.md`, `100-the-last-three-and-the-collapsed-entrance-gate.md`, `16-combat-turns-and-ai.md`, `17-magic.md`, `27-codex-and-the-shrine-chamber.md`, `28-shadowlords-and-blackthorn.md`, `29-npc-behaviour-and-arrest.md`, `31-line-of-sight.md`, `38-terrain-movement-cost.md`, `50-hole-up-camp-sleep-repair.md`, `51-closing-four-known-gaps.md`, `81-three-mode-loops-are-mutually-exclusive.md`, `92-white-noise-shim-and-ambient-sound.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_295AC` | sub_295AC  byte_3F844[y*16 + x] = 某個位元組                執行期寫入(§4 說明是什麼) | `34-ending-trigger.md` |
| `sub_297F4` | 主要函式  /  `sub_135FC`(結局)、`sub_161E4`(觸發)、`sub_15E20`(撤離)、`sub_297F4`(疊圖)  / | `34-ending-trigger.md` |
| `sub_29A64` | 0x01 在兩種單位上意義相反**,而 `sub_29A64` 就是靠這一點把兩邊算清楚: | `16-combat-turns-and-ai.md` |
| `sub_29BEC` | 主要函式  /  `sub_DEE4`(開關月門)、`sub_DE74`(這一顆算不算在場)、`sub_E084`(玩家踏上)、`sub_2870`(怪物踏上)、`sub_29BEC`(★ 重畫時挑動畫格)、`sub_2… | `86-tile-0xDC-is-an-open-moongate.md` |
| `sub_29D64` | 主要函式  /  `sub_29D64`(重畫)、`sub_2E0E8`(罩子入口)、`sub_2DDB0`(flood fill)、`sub_2E1D0`(擋不擋)、`sub_2E8D0`(距離)、`sub_29304… | `100-the-last-three-and-the-collapsed-entrance-gate.md`, `31-line-of-sight.md`, `34-ending-trigger.md`, `50-hole-up-camp-sleep-repair.md`, `71-the-use-list-had-29-empty-slots.md`, `86-tile-0xDC-is-an-open-moongate.md`, `92-white-noise-shim-and-ambient-sound.md`, `93-the-tavern-meal-was-never-haggled.md` |
| `sub_29EEC` | do v2 = sub_29EEC(v1, v0); while ( v2 != 89 && v2 != 78 ); | `06-conversation-script.md`, `25-shrines.md`, `26-yell-words-of-power-shadowlords.md`, `65-the-wishing-well-easter-egg.md`, `70-hunger-poison-and-the-vanishing-rings.md` |
| `sub_29F4C` | `sub_29F4C` / `sub_2A33C`  /  面板上印一行角色(名字補到 9 欄 + HP + 狀態字元) /  Ebiten 那一層已經在畫。⚠ `sub_2A33C` 對第 0 位印字元 26(CP43… | `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_2A0C4` | 主要函式  /  `sub_1A38C`(用碎片)、`sub_2D9D0`(風浪)、`sub_2A1E8` / `sub_2A0C4` / `sub_2A984`(狀態列)  / | `80-implemented-but-unreachable.md` |
| `sub_2A1E8` | 主要函式  /  `sub_1A38C`(用碎片)、`sub_2D9D0`(風浪)、`sub_2A1E8` / `sub_2A0C4` / `sub_2A984`(狀態列)  / | `71-the-use-list-had-29-empty-slots.md`, `80-implemented-but-unreachable.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_2A33C` | `sub_29F4C` / `sub_2A33C`  /  面板上印一行角色(名字補到 9 欄 + HP + 狀態字元) /  Ebiten 那一層已經在畫。⚠ `sub_2A33C` 對第 0 位印字元 26(CP43… | `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_2A464` | 主要函式  /  `sub_2AB38`(★ 陷阱分派)、`sub_2A464`(單人扣血)、`sub_2AB08`(★ 中毒)、`sub_2A4D0`(★ 全隊扣血)、`sub_2B724`(1..30 骰,已見 `d… | `37-look-signs-and-the-sky.md`, `66-hexrays-truncation-audit.md`, `70-hunger-poison-and-the-vanishing-rings.md`, `91-chest-traps-four-kinds-and-their-weights.md`, `92-white-noise-shim-and-ambient-sound.md`, `95-the-dungeon-has-its-own-look.md` |
| `sub_2A4D0` | 主要函式  /  `sub_2AB38`(★ 陷阱分派)、`sub_2A464`(單人扣血)、`sub_2AB08`(★ 中毒)、`sub_2A4D0`(★ 全隊扣血)、`sub_2B724`(1..30 骰,已見 `d… | `18-dungeons.md`, `66-hexrays-truncation-audit.md`, `68-troll-bridge-and-collision.md`, `69-the-overworld-cannon-did-nothing.md`, `70-hunger-poison-and-the-vanishing-rings.md`, `81-three-mode-loops-are-mutually-exclusive.md`, `91-chest-traps-four-kinds-and-their-weights.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_2A50C` | 主要函式  /  ★ `sub_2BCC8`(再生戒指)、`sub_2ECE8`(戒指的每回合效果)、`sub_1F4E4`(力場消散)、`sub_2A50C` 尾段、`sub_16370`、★ `sub_FEC` + … | `70-hunger-poison-and-the-vanishing-rings.md`, `80-implemented-but-unreachable.md`, `81-three-mode-loops-are-mutually-exclusive.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_2A610` | case 2: return (tile & 0xF0) == 0x60  /  /  sub_2A674(tile)  /  /  sub_2A610(mover, tile);  // 水陸兩棲 | `02-movement-and-tile-flags.md`, `03-scene-entry-and-tile-semantics.md`, `47-move-modes-and-time-of-day.md`, `51-closing-four-known-gaps.md` |
| `sub_2A674` | tile 0 為何被歸進「水」  /  `sub_2A674` 的 `tile < 4` 把 0 併進來。**視覺上 tile 0 根本不是水**(算繪出來是一團紅黃爆裂圖案,tile 1–3 才是藍色水面),所以這不是… | `02-movement-and-tile-flags.md`, `47-move-modes-and-time-of-day.md`, `51-closing-four-known-gaps.md` |
| `sub_2A694` | 通行判定第一參數  /  `sub_2A694(0, tile)`  /  `movzx eax, byte_3E08C`  /  照抄的話船、馬、飛毯全都照步行規則走  / | `02-movement-and-tile-flags.md`, `03-scene-entry-and-tile-semantics.md`, `47-move-modes-and-time-of-day.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_2A7F4` | 主要函式  /  ★ `sub_2A7F4`(「Select:」選單)、`sub_2A974`(包一層)、`sub_E19C`(選隊員)、`sub_2BC18` / `sub_2BBDC`(游標上下限)  / | `97-set-active-player.md`, `98-select-menu-and-song-melody-matching.md` |
| `sub_2A974` | 主要函式  /  ★ `sub_2A7F4`(「Select:」選單)、`sub_2A974`(包一層)、`sub_E19C`(選隊員)、`sub_2BC18` / `sub_2BBDC`(游標上下限)  / | `97-set-active-player.md`, `98-select-menu-and-song-melody-matching.md` |
| `sub_2A984` | 主要函式  /  `sub_1A38C`(用碎片)、`sub_2D9D0`(風浪)、`sub_2A1E8` / `sub_2A0C4` / `sub_2A984`(狀態列)  / | `23-wind-and-sailing.md`, `80-implemented-but-unreachable.md`, `84-the-sailing-gate-and-the-wind-table.md` |
| `sub_2AB08` | 主要函式  /  `sub_2AB38`(★ 陷阱分派)、`sub_2A464`(單人扣血)、`sub_2AB08`(★ 中毒)、`sub_2A4D0`(★ 全隊扣血)、`sub_2B724`(1..30 骰,已見 `d… | `91-chest-traps-four-kinds-and-their-weights.md` |
| `sub_2AB38` | 主要函式  /  `sub_2AB38`(★ 陷阱分派)、`sub_2A464`(單人扣血)、`sub_2AB08`(★ 中毒)、`sub_2A4D0`(★ 全隊扣血)、`sub_2B724`(1..30 骰,已見 `d… | `75-open-did-nothing-outside-dungeons.md`, `91-chest-traps-four-kinds-and-their-weights.md`, `92-white-noise-shim-and-ambient-sound.md` |
| `sub_2AC08` | → 印 `EARTHQUAKE!` + `sub_2AC08`(震動畫面)+ `sub_2A4D0`(全隊 random(1,8) 傷)。 | `81-three-mode-loops-are-mutually-exclusive.md`, `90-cdda-scenes-and-sound-effect-indices.md` |
| `sub_2ACF4` | 位址  /  `sub_2ACF4`(59 case 主分派器)、`sub_1F3A4`(Ready)、`sub_1E418`(找下一個非空欄位)  / | `35-harp-and-the-secret-door.md`, `41-jimmy-neworder-gem-ztats.md`, `42-equipment-slots-ready-wear.md`, `49-command-table-and-two-empty-keys.md`, `50-hole-up-camp-sleep-repair.md`, `54-pass-vehicle-verbs-and-the-avatar-line.md`, `60-command-echo-and-menu-keys.md`, `81-three-mode-loops-are-mutually-exclusive.md`, `95-the-dungeon-has-its-own-look.md`, `97-set-active-player.md` |
| `sub_2B1C8` | if (命中) { sub_2B1C8(隊伍X, 隊伍Y); sub_22F0() }          ; ★ 打中 → 船損 | `80-implemented-but-unreachable.md`, `83-monster-attacks-and-two-different-wind-rules.md`, `88-the-load-window-culling-and-ship-cannon.md`, `92-white-noise-shim-and-ambient-sound.md` |
| `sub_2B21C` | `sub_2B21C`  /  同上 / 0x0A, 0xBB8, 0x7D0  /  DAME1 + DAME2  / | `92-white-noise-shim-and-ambient-sound.md` |
| `sub_2B2AC` | T  /  Talk  /  地表 `sub_2B2AC`;地點 ≤ 0x20 → `Talk-Funny, no response!`;其餘 `sub_1B658`  / | `35-harp-and-the-secret-door.md`, `49-command-table-and-two-empty-keys.md`, `68-troll-bridge-and-collision.md` |
| `sub_2B360` | if (sub_2B360(x, y - 1, floor) != 0xFC) return; // ← 正北一格不是暗影君主:**沉默** | `02-movement-and-tile-flags.md`, `03-scene-entry-and-tile-semantics.md`, `11-map-objects.md`, `28-shadowlords-and-blackthorn.md`, `50-hole-up-camp-sleep-repair.md`, `68-troll-bridge-and-collision.md`, `96-search-had-three-more-layers.md` |
| `sub_2B3DC` | 主要函式  /  `sub_172C4`(F 開砲,**244 行**)、`sub_2B3DC`(查某一格上的物件)、`sub_2A4D0`(全隊傷害)、`sub_2E0`/`sub_218`/`sub_268`(NPC… | `12-npc-movement.md`, `69-the-overworld-cannon-did-nothing.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_2B518` | 6. 落點與槽(`sub_215C` / `sub_2B57C` / `sub_2B518` / `sub_2B6C8`) | `82-the-overworld-world-turn.md`, `96-search-had-three-more-layers.md` |
| `sub_2B57C` | 主要函式  /  `sub_147A8`(地面 Search 主體)、★ `sub_13FC4`(挖月石)、★★ `sub_14090`(午夜採藥)、★★ `sub_14160`(113 筆固定物品)、★ `sub_13… | `36-sandalwood-box-npc-objects.md`, `65-the-wishing-well-easter-egg.md`, `82-the-overworld-world-turn.md`, `96-search-had-three-more-layers.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_2B64C` | 主要函式  /  `sub_15374`(Open 指令本體)、`sub_152B8`(地牢的開箱)、`sub_15108`(物件層的開箱)、`sub_2B64C`(把 tile 寫回去)、`sub_1A54`(主迴圈的… | `69-the-overworld-cannon-did-nothing.md`, `75-open-did-nothing-outside-dungeons.md` |
| `sub_2B67C` | 主要函式  /  ★★ `sub_EEEC`(地牢 Look)、★ `sub_13BA8`(第一人稱方向提問)、`sub_13B4C`(相對方向 → 座標)、★ `sub_E19C`(選隊員)、★ `sub_2B67C`… | `37-look-signs-and-the-sky.md`, `66-hexrays-truncation-audit.md`, `68-troll-bridge-and-collision.md`, `95-the-dungeon-has-its-own-look.md` |
| `sub_2B6C8` | 主要函式  /  `sub_CD28`(井)、`sub_D258`(Look 分派)、`sub_27C98`(字串比對)、`sub_2B770`(文字輸入)、`sub_2B6C8`(寫物件槽)、`sub_2B57C`(找… | `36-sandalwood-box-npc-objects.md`, `38-terrain-movement-cost.md`, `57-crown-and-sceptre-placement.md`, `65-the-wishing-well-easter-egg.md`, `69-the-overworld-cannon-did-nothing.md`, `75-open-did-nothing-outside-dungeons.md`, `82-the-overworld-world-turn.md`, `88-the-load-window-culling-and-ship-cannon.md`, `96-search-had-three-more-layers.md` |
| `sub_2B710` | 主要函式  /  `sub_A360`(隊員的戰鬥回合,**558 行**)、`sub_BCC4`(掙脫)、`sub_1F840`(命中結果)、`sub_2B724`/`sub_2B710`(門檻骰)、`sub_2ED5… | `15-combat-formulas.md`, `67-corpser-and-the-sleeping-party-member.md` |
| `sub_2B724` | 主要函式  /  `sub_2AB38`(★ 陷阱分派)、`sub_2A464`(單人扣血)、`sub_2AB08`(★ 中毒)、`sub_2A4D0`(★ 全隊扣血)、`sub_2B724`(1..30 骰,已見 `d… | `15-combat-formulas.md`, `17-magic.md`, `67-corpser-and-the-sleeping-party-member.md`, `68-troll-bridge-and-collision.md`, `91-chest-traps-four-kinds-and-their-weights.md` |
| `sub_2B740` | `sub_2B740(1)` 重畫、計數器 +1。**沒有任何一行寫 `byte_3E0A6`/`byte_3E0A7`。** | `100-the-last-three-and-the-collapsed-entrance-gate.md`, `74-two-swamp-dice-and-an-unnamed-failure.md`, `81-three-mode-loops-are-mutually-exclusive.md`, `83-monster-attacks-and-two-different-wind-rules.md` |
| `sub_2B770` | 主要函式  /  `sub_CD28`(井)、`sub_D258`(Look 分派)、`sub_27C98`(字串比對)、`sub_2B770`(文字輸入)、`sub_2B6C8`(寫物件槽)、`sub_2B57C`(找… | `26-yell-words-of-power-shadowlords.md`, `28-shadowlords-and-blackthorn.md`, `65-the-wishing-well-easter-egg.md` |
| `sub_2B8CC` | ✅ 早就從組語逆過,截斷沒造成損失  /  `sub_21108` 酒單(`tavern.go`)、`sub_1D394` 聖壇與 `ALAKAZAM`(`shrine.go`)、`sub_14CAC`/`sub_14B… | `38-terrain-movement-cost.md`, `41-jimmy-neworder-gem-ztats.md`, `48-dungeon-wandering-monster-and-arena.md`, `50-hole-up-camp-sleep-repair.md`, `66-hexrays-truncation-audit.md`, `73-escape-is-a-key-and-it-has-two-gates.md` |
| `sub_2BBB8` | 主要函式  /  `sub_1C1E8`(四個 opcode 的分派)、`sub_1B854`(索取金幣)、`sub_1B964`(給東西)、`sub_1C1C8`(認得旗標)、`sub_2BBB8`/`sub_2BBD… | `06-conversation-script.md`, `19-levelup.md`, `79-four-dialogue-opcodes-were-corrupting-the-text.md`, `86-tile-0xDC-is-an-open-moongate.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_2BBDC` | 主要函式  /  `sub_1C1E8`(四個 opcode 的分派)、`sub_1B854`(索取金幣)、`sub_1B964`(給東西)、`sub_1C1C8`(認得旗標)、`sub_2BBB8`/`sub_2BBD… | `10-shop-prices-and-trade.md`, `16-combat-turns-and-ai.md`, `19-levelup.md`, `79-four-dialogue-opcodes-were-corrupting-the-text.md`, `97-set-active-player.md`, `98-select-menu-and-song-melody-matching.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_2BBFC` | 0x89 / 0x8A  /  **業報** +1 / −1(上限 99)  /  `sub_2BBB8(&byte_3E098, 1, 0x63)` / `sub_2BBFC`  / | `06-conversation-script.md`, `51-closing-four-known-gaps.md`, `86-tile-0xDC-is-an-open-moongate.md` |
| `sub_2BC18` | 主要函式  /  ★ `sub_2A7F4`(「Select:」選單)、`sub_2A974`(包一層)、`sub_E19C`(選隊員)、`sub_2BC18` / `sub_2BBDC`(游標上下限)  / | `97-set-active-player.md`, `98-select-menu-and-song-melody-matching.md` |
| `sub_2BC34` | 2. 沿線一步一步走,每一步 `sub_2BC34` 查那一格擋不擋;**擋住就停在那裡**。 | `20-projectiles.md` |
| `sub_2BC70` | ├ sub_2BC70   把像素座標換回格子,順便判出不出界 | `20-projectiles.md` |
| `sub_2BCC8` | 主要函式  /  ★ `sub_2BCC8`(再生戒指)、`sub_2ECE8`(戒指的每回合效果)、`sub_1F4E4`(力場消散)、`sub_2A50C` 尾段、`sub_16370`、★ `sub_FEC` + … | `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_2BD40` | 主要函式  /  ★ `sub_2BD40`(Set Active Plr)、`sub_E19C`(選隊員,見 `docs/re/95`)、`sub_4B14`(按鍵分派的 default)  / | `97-set-active-player.md`, `98-select-menu-and-song-melody-matching.md` |
| `sub_2BDE0` | 主要函式  /  ★★ `sub_2C598`(白噪 → PCM 轉接)、★★ `sub_2BDE0`(環境音掃描)、`sub_2C46C`(放 PCM)、`sub_2C4F4`(滴答)、`sub_33D78`(蜂鳴器單… | `90-cdda-scenes-and-sound-effect-indices.md`, `92-white-noise-shim-and-ambient-sound.md` |
| `sub_2C118` | `sub_2C118`(★ 腳步) /  1, 0x19, 0x3E8 → 1, 0x19, 0x5DC  /  **WALK**  / | `100-the-last-three-and-the-collapsed-entrance-gate.md`, `92-white-noise-shim-and-ambient-sound.md` |
| `sub_2C188` | ⬜ `sub_27230(0x21)` 與收尾的 `sub_2C188(0x4B0, 0x7D0, 1, 0x28)` 是音效 / 延遲, | `67-corpser-and-the-sleeping-party-member.md`, `73-escape-is-a-key-and-it-has-two-gates.md`, `83-monster-attacks-and-two-different-wind-rules.md`, `90-cdda-scenes-and-sound-effect-indices.md` |
| `sub_2C250` | ⬜ `sub_3FE4` 開頭的 `dword_5FF34 = 0` 與 `dword_601C0 == 1 → sub_2C250` | `78-the-bottom-of-a-dungeon-opens-into-the-underworld.md` |
| `sub_2C2AC` | 與 `sub_2C2AC`(`initSoundEffect`)之間 ⇒ **它只跑一次,在讀檔完成之後。** | `63-fm-towns-sound-effects.md`, `87-bgm-selection-and-the-tbl-tables.md` |
| `sub_2C46C` | 主要函式  /  `sub_310E8`(★ 填 MSF 表)、`sub_30F8C`(★ 播一段 CD)、`sub_6730`(啟動 / 選單)、`sub_6128`(製作群字幕)、`sub_22D00`(石室)、`s… | `63-fm-towns-sound-effects.md`, `90-cdda-scenes-and-sound-effect-indices.md`, `91-chest-traps-four-kinds-and-their-weights.md`, `92-white-noise-shim-and-ambient-sound.md` |
| `sub_2C4F4` | 主要函式  /  ★★ `sub_2C598`(白噪 → PCM 轉接)、★★ `sub_2BDE0`(環境音掃描)、`sub_2C46C`(放 PCM)、`sub_2C4F4`(滴答)、`sub_33D78`(蜂鳴器單… | `03-scene-entry-and-tile-semantics.md`, `63-fm-towns-sound-effects.md`, `90-cdda-scenes-and-sound-effect-indices.md`, `92-white-noise-shim-and-ambient-sound.md` |
| `sub_2C598` | 主要函式  /  ★★ `sub_2C598`(白噪 → PCM 轉接)、★★ `sub_2BDE0`(環境音掃描)、`sub_2C46C`(放 PCM)、`sub_2C4F4`(滴答)、`sub_33D78`(蜂鳴器單… | `65-the-wishing-well-easter-egg.md`, `66-hexrays-truncation-audit.md`, `67-corpser-and-the-sleeping-party-member.md`, `68-troll-bridge-and-collision.md`, `90-cdda-scenes-and-sound-effect-indices.md`, `91-chest-traps-four-kinds-and-their-weights.md`, `92-white-noise-shim-and-ambient-sound.md` |
| `sub_2C740` | `sub_C414` / `sub_1DA10` / `sub_135FC`  /  `sub_2C740("MISCMAPS.DAT", byte_3F844, 0xB0, 位移)` —— 石室整份載入  / | `00-hexrays-p3-verified.md`, `03-scene-entry-and-tile-semantics.md`, `04-npc-schedule-and-clock.md`, `11-map-objects.md`, `14-combat-maps.md`, `18-dungeons.md`, `27-codex-and-the-shrine-chamber.md`, `28-shadowlords-and-blackthorn.md`, `30-ending.md`, `34-ending-trigger.md`, `55-transfer-from-ultima-iv.md` |
| `sub_2CB3C` | call sub_2CB3C(2) | `81-three-mode-loops-are-mutually-exclusive.md` |
| `sub_2CBEC` | 主要函式  /  `sub_2CBEC`(★ 視窗定位)、`sub_2D014`(★ 捲動)、`sub_2E24` 尾段(★ 清場)、`sub_2B24`(⚠ 第二參數是死的)、`sub_27F0`(影主的遮罩表)、`s… | `17-magic.md`, `83-monster-attacks-and-two-different-wind-rules.md`, `88-the-load-window-culling-and-ship-cannon.md` |
| `sub_2CC8C` | sub_2CC8C(); sub_2E24();                       ; 動畫 + 一個世界回合 | `50-hole-up-camp-sleep-repair.md` |
| `sub_2CCFC` | 主要函式  /  `sub_2CCFC`(★ 真正的航行閘門)、`sub_2D174`(移動分派)、`sub_2D014`(改座標)、`sub_2D2D0`(揚帆的節奏)、`sub_2B24`(★ 怪物移動)、`sub_… | `66-hexrays-truncation-audit.md`, `84-the-sailing-gate-and-the-wind-table.md` |
| `sub_2CE70` | 主要函式  /  `sub_3010`(過橋)、`sub_2F48`(過路費)、`sub_2CE70`(通行判定 + 撞擊)、`sub_2D9D0`(移動後的分派)  / | `66-hexrays-truncation-audit.md`, `68-troll-bridge-and-collision.md`, `92-white-noise-shim-and-ambient-sound.md` |
| `sub_2D014` | 主要函式  /  `sub_2CBEC`(★ 視窗定位)、`sub_2D014`(★ 捲動)、`sub_2E24` 尾段(★ 清場)、`sub_2B24`(⚠ 第二參數是死的)、`sub_27F0`(影主的遮罩表)、`s… | `68-troll-bridge-and-collision.md`, `84-the-sailing-gate-and-the-wind-table.md`, `88-the-load-window-culling-and-ship-cannon.md` |
| `sub_2D0BC` | 移動成本分級  /  `"Slow progress!"` / `"Very slow!"` 在 `sub_2D0BC`,尚未讀  / | `02-movement-and-tile-flags.md`, `38-terrain-movement-cost.md`, `40-push-and-the-main-menu.md` |
| `sub_2D174` | 主要函式  /  `sub_2CCFC`(★ 真正的航行閘門)、`sub_2D174`(移動分派)、`sub_2D014`(改座標)、`sub_2D2D0`(揚帆的節奏)、`sub_2B24`(★ 怪物移動)、`sub_… | `38-terrain-movement-cost.md`, `66-hexrays-truncation-audit.md`, `81-three-mode-loops-are-mutually-exclusive.md`, `83-monster-attacks-and-two-different-wind-rules.md`, `84-the-sailing-gate-and-the-wind-table.md` |
| `sub_2D2D0` | 主要函式  /  `sub_2CCFC`(★ 真正的航行閘門)、`sub_2D174`(移動分派)、`sub_2D014`(改座標)、`sub_2D2D0`(揚帆的節奏)、`sub_2B24`(★ 怪物移動)、`sub_… | `23-wind-and-sailing.md`, `38-terrain-movement-cost.md`, `68-troll-bridge-and-collision.md`, `83-monster-attacks-and-two-different-wind-rules.md`, `84-the-sailing-gate-and-the-wind-table.md`, `85-monsters-move-and-terrain-slows-them.md` |
| `sub_2D478` | `sub_1EFC8`(清單瀏覽器)、`sub_CAC` / `sub_4074` / `sub_2D478`(Attack 三支)  / | `49-command-table-and-two-empty-keys.md`, `60-command-echo-and-menu-keys.md` |
| `sub_2D564` | 主要函式  /  ★ `sub_2D72C`(Enter 的 tile 分派)、`sub_2D564`、`sub_27D24`(讀檔序列)、★ `sub_92C0` + `sub_89EC` + `sub_8BA0`(N… | `03-scene-entry-and-tile-semantics.md`, `100-the-last-three-and-the-collapsed-entrance-gate.md`, `18-dungeons.md`, `26-yell-words-of-power-shadowlords.md`, `87-bgm-selection-and-the-tbl-tables.md`, `98-select-menu-and-song-melody-matching.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_2D72C` | 主要函式  /  ★ `sub_2D72C`(Enter 的 tile 分派)、`sub_2D564`、`sub_27D24`(讀檔序列)、★ `sub_92C0` + `sub_89EC` + `sub_8BA0`(N… | `03-scene-entry-and-tile-semantics.md`, `100-the-last-three-and-the-collapsed-entrance-gate.md`, `27-codex-and-the-shrine-chamber.md`, `87-bgm-selection-and-the-tbl-tables.md`, `98-select-menu-and-song-melody-matching.md` |
| `sub_2D944` | 主要函式  /  ★ `sub_2D72C`(Enter 的 tile 分派)、`sub_2D564`、`sub_27D24`(讀檔序列)、★ `sub_92C0` + `sub_89EC` + `sub_8BA0`(N… | `100-the-last-three-and-the-collapsed-entrance-gate.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_2D998` | ✅ **幽冥界地震**(`sub_2D998`):`byte_3E0A5 != 0` 且 `random(0,255) == 0x69` | `80-implemented-but-unreachable.md`, `81-three-mode-loops-are-mutually-exclusive.md` |
| `sub_2D9D0` | 主要函式  /  `sub_10BDC`(踏進沼澤的中毒)、`sub_1318`(每回合的地形效果)、`sub_2D9D0`(移動後的分派)、`sub_1F570`(落空的訊息)、`sub_1F528`(印單位名字)、`… | `100-the-last-three-and-the-collapsed-entrance-gate.md`, `38-terrain-movement-cost.md`, `66-hexrays-truncation-audit.md`, `68-troll-bridge-and-collision.md`, `74-two-swamp-dice-and-an-unnamed-failure.md`, `80-implemented-but-unreachable.md`, `81-three-mode-loops-are-mutually-exclusive.md`, `83-monster-attacks-and-two-different-wind-rules.md`, `84-the-sailing-gate-and-the-wind-table.md`, `88-the-load-window-culling-and-ship-cannon.md` |
| `sub_2DD44` | 主要函式  /  `sub_2CBEC`(★ 視窗定位)、`sub_2D014`(★ 捲動)、`sub_2E24` 尾段(★ 清場)、`sub_2B24`(⚠ 第二參數是死的)、`sub_27F0`(影主的遮罩表)、`s… | `81-three-mode-loops-are-mutually-exclusive.md`, `88-the-load-window-culling-and-ship-cannon.md` |
| `sub_2DDB0` | 主要函式  /  `sub_29D64`(重畫)、`sub_2E0E8`(罩子入口)、`sub_2DDB0`(flood fill)、`sub_2E1D0`(擋不擋)、`sub_2E8D0`(距離)、`sub_29304… | `31-line-of-sight.md` |
| `sub_2E0E8` | 主要函式  /  `sub_29D64`(重畫)、`sub_2E0E8`(罩子入口)、`sub_2DDB0`(flood fill)、`sub_2E1D0`(擋不擋)、`sub_2E8D0`(距離)、`sub_29304… | `17-magic.md`, `31-line-of-sight.md`, `71-the-use-list-had-29-empty-slots.md` |
| `sub_2E1D0` | 主要函式  /  `sub_29D64`(重畫)、`sub_2E0E8`(罩子入口)、`sub_2DDB0`(flood fill)、`sub_2E1D0`(擋不擋)、`sub_2E8D0`(距離)、`sub_29304… | `31-line-of-sight.md` |
| `sub_2E21C` | 主要函式  /  `sub_29D64`(重畫)、`sub_2E0E8`(罩子入口)、`sub_2DDB0`(flood fill)、`sub_2E1D0`(擋不擋)、`sub_2E8D0`(距離)、`sub_29304… | `31-line-of-sight.md`, `86-tile-0xDC-is-an-open-moongate.md`, `88-the-load-window-culling-and-ship-cannon.md` |
| `sub_2E364` | 主要函式  /  `sub_18380`(ESC = 撤離)、`sub_A360`(戰鬥分派器)、`sub_42CC`(進地牢房間)、`sub_A9EC`(`byte_3E0B3` 的設定處)、`sub_2E364`(戰… | `16-combat-turns-and-ai.md`, `17-magic.md`, `48-dungeon-wandering-monster-and-arena.md`, `50-hole-up-camp-sleep-repair.md`, `52-one-coordinate-pair-one-tile-accessor.md`, `73-escape-is-a-key-and-it-has-two-gates.md` |
| `sub_2E51C` | call sub_2E8B0              ; → sub_2E51C(0) + sub_2E364(4, 守夜, 時數) | `14-combat-maps.md`, `48-dungeon-wandering-monster-and-arena.md`, `50-hole-up-camp-sleep-repair.md` |
| `sub_2E58C` | sub_2E58C(word_3E77C[npc*16]);                 // 開打 —— 與撞上野外怪物同一支 | `14-combat-maps.md`, `16-combat-turns-and-ai.md`, `29-npc-behaviour-and-arrest.md`, `68-troll-bridge-and-collision.md` |
| `sub_2E8B0` | call sub_2E8B0              ; → sub_2E51C(0) + sub_2E364(4, 守夜, 時數) | `48-dungeon-wandering-monster-and-arena.md`, `50-hole-up-camp-sleep-repair.md`, `73-escape-is-a-key-and-it-has-two-gates.md` |
| `sub_2E8D0` | 主要函式  /  `sub_29D64`(重畫)、`sub_2E0E8`(罩子入口)、`sub_2DDB0`(flood fill)、`sub_2E1D0`(擋不擋)、`sub_2E8D0`(距離)、`sub_29304… | `31-line-of-sight.md` |
| `sub_2E8E8` | sub_2E8E8   把某個扇區的 16 個格子在**亮度圖**上設成 0xFF(或 0) | `31-line-of-sight.md` |
| `sub_2E944` | sub_2E944   每幀:關掉扇區 n、打開扇區 n+2,n 加一,到 16 歸零 | `31-line-of-sight.md` |
| `sub_2EAE4` | ⚠ `0x14` 那個數字有兩種讀法:當生物索引是 **20 = 巨鼠**(而 `sub_2EAE4` | `17-magic.md` |
| `sub_2ECE8` | 主要函式  /  ★ `sub_2BCC8`(再生戒指)、`sub_2ECE8`(戒指的每回合效果)、`sub_1F4E4`(力場消散)、`sub_2A50C` 尾段、`sub_16370`、★ `sub_FEC` + … | `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_2ED50` | 主要函式  /  `sub_A360`(隊員的戰鬥回合,**558 行**)、`sub_BCC4`(掙脫)、`sub_1F840`(命中結果)、`sub_2B724`/`sub_2B710`(門檻骰)、`sub_2ED5… | `53-party-tile-and-the-arena-floor.md`, `67-corpser-and-the-sleeping-party-member.md`, `72-ready-had-seven-missing-rules.md` |
| `sub_2EDC0` | `sub_2EDC0` / `sub_2EDF8`(紮營讓人睡著) /  `camp()` 的「不是 'P' 就睡」分支 + `putUnitToSleep()`  / | `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_2EDF8` | `sub_2EDC0` / `sub_2EDF8`(紮營讓人睡著) /  `camp()` 的「不是 'P' 就睡」分支 + `putUnitToSleep()`  / | `17-magic.md`, `53-party-tile-and-the-arena-floor.md`, `67-corpser-and-the-sleeping-party-member.md`, `71-the-use-list-had-29-empty-slots.md`, `72-ready-had-seven-missing-rules.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_2EE84` | 從哪裡來  /  `docs/re/66` 的截斷清單:`sub_2A50C` 113 行 → 4 行 C(`'Starving!'` 掉了)、`sub_2EE84` 214 行 → 25 行(`'A ring has … | `70-hunger-poison-and-the-vanishing-rings.md`, `72-ready-had-seven-missing-rules.md`, `74-two-swamp-dice-and-an-unnamed-failure.md` |
| `sub_2F0EC` | `sub_2F0EC`  /  `sub eax, eax` 之後寫入 → 清成 0  / | `16-combat-turns-and-ai.md`, `18-dungeons.md`, `34-ending-trigger.md` |
| `sub_2F294` | ⬜ **走出戰場邊緣**那條路(`sub_2F294`,印 `escapes!`)引擎有做, | `73-escape-is-a-key-and-it-has-two-gates.md` |
| `sub_2F2BC` | │       └ sub_2F2BC  角色的裝備防禦加總 | `15-combat-formulas.md` |
| `sub_2F35C` | `sub_2F35C`(把戒指從身上拿掉)只當黑盒用 —— 引擎直接寫 `CharRing = ItemNone`, | `70-hunger-poison-and-the-vanishing-rings.md`, `72-ready-had-seven-missing-rules.md` |
| `sub_2FD78` | `sub_2FE58`  /  1  /  ⬜ 未命名(反覆呼叫 `sub_2FD78`,無字串) / | `90-cdda-scenes-and-sound-effect-indices.md` |
| `sub_2FE58` | 主要函式  /  `sub_310E8`(★ 填 MSF 表)、`sub_30F8C`(★ 播一段 CD)、`sub_6730`(啟動 / 選單)、`sub_6128`(製作群字幕)、`sub_22D00`(石室)、`s… | `90-cdda-scenes-and-sound-effect-indices.md`, `audio-pipeline.md` |
| `sub_3007C` | 呼叫者 `sub_C2BC` / `sub_C2D0` / `sub_C414`(三處)。⚠ 與 `sub_3007C` | `100-the-last-three-and-the-collapsed-entrance-gate.md`, `97-set-active-player.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `sub_30F8C` | 主要函式  /  `sub_310E8`(★ 填 MSF 表)、`sub_30F8C`(★ 播一段 CD)、`sub_6730`(啟動 / 選單)、`sub_6128`(製作群字幕)、`sub_22D00`(石室)、`s… | `90-cdda-scenes-and-sound-effect-indices.md` |
| `sub_310E8` | 主要函式  /  `sub_310E8`(★ 填 MSF 表)、`sub_30F8C`(★ 播一段 CD)、`sub_6730`(啟動 / 選單)、`sub_6128`(製作群字幕)、`sub_22D00`(石室)、`s… | `90-cdda-scenes-and-sound-effect-indices.md` |
| `sub_313B8` | 主要函式  /  `sub_313B8`(載入 `U5_BGM.TBL` + `FM_BANK.FMB`)、`sub_3453B`(★ 交出**檔名**)、`sub_34B1B`(★ 遠端呼叫驅動)、`sub_34D3D… | `87-bgm-selection-and-the-tbl-tables.md`, `89-music-playback-is-not-in-the-game-binary.md` |
| `sub_3181C` | 主要函式  /  `sub_3181C`(★ 播第 n 首)、`sub_313B8`(載入 `U5_BGM.TBL`)、`sub_31DC0`(★★ 開局挑曲)、`sub_31CB8`(當前曲號)、`sub_34685`… | `03-scene-entry-and-tile-semantics.md`, `27-codex-and-the-shrine-chamber.md`, `50-hole-up-camp-sleep-repair.md`, `87-bgm-selection-and-the-tbl-tables.md`, `98-select-menu-and-song-melody-matching.md`, `audio-pipeline.md` |
| `sub_31CB8` | 主要函式  /  `sub_3181C`(★ 播第 n 首)、`sub_313B8`(載入 `U5_BGM.TBL`)、`sub_31DC0`(★★ 開局挑曲)、`sub_31CB8`(當前曲號)、`sub_34685`… | `03-scene-entry-and-tile-semantics.md`, `87-bgm-selection-and-the-tbl-tables.md` |
| `sub_31DC0` | 主要函式  /  `sub_3181C`(★ 播第 n 首)、`sub_313B8`(載入 `U5_BGM.TBL`)、`sub_31DC0`(★★ 開局挑曲)、`sub_31CB8`(當前曲號)、`sub_34685`… | `87-bgm-selection-and-the-tbl-tables.md`, `audio-pipeline.md` |
| `sub_32244` | `sub_12B20` → 0x0E、`sub_135FC` → 0x0D / 9、`sub_32244` → 0x0A  /  —  /  ⬜ 過場 / 開場,引擎沒有對應流程  / | `87-bgm-selection-and-the-tbl-tables.md` |
| `sub_33524` | sub_33524(0, ebx)                                     ; AH=14h | `90-cdda-scenes-and-sound-effect-indices.md` |
| `sub_33575` | sub_33575(0) | `90-cdda-scenes-and-sound-effect-indices.md` |
| `sub_3359A` | sub_3359A(0) | `90-cdda-scenes-and-sound-effect-indices.md` |
| `sub_3370A` | sub_3370A(0, ebx, &byte_6095C[n*3])                   ; ★ 從起點播到終點 | `90-cdda-scenes-and-sound-effect-indices.md` |
| `sub_337DE` | sub_337DE(0, &err, &var_4, byte_60971, word_6096E)   ; 查碟(cdr_cdinfo 那一族) | `90-cdda-scenes-and-sound-effect-indices.md` |
| `sub_33D78` | 主要函式  /  ★★ `sub_2C598`(白噪 → PCM 轉接)、★★ `sub_2BDE0`(環境音掃描)、`sub_2C46C`(放 PCM)、`sub_2C4F4`(滴答)、`sub_33D78`(蜂鳴器單… | `92-white-noise-shim-and-ambient-sound.md` |
| `sub_3453B` | 主要函式  /  `sub_313B8`(載入 `U5_BGM.TBL` + `FM_BANK.FMB`)、`sub_3453B`(★ 交出**檔名**)、`sub_34B1B`(★ 遠端呼叫驅動)、`sub_34D3D… | `89-music-playback-is-not-in-the-game-binary.md` |
| `sub_34685` | 主要函式  /  `sub_3181C`(★ 播第 n 首)、`sub_313B8`(載入 `U5_BGM.TBL`)、`sub_31DC0`(★★ 開局挑曲)、`sub_31CB8`(當前曲號)、`sub_34685`… | `87-bgm-selection-and-the-tbl-tables.md` |
| `sub_34834` | call    sub_34834                ; AH=25h **播放 PCM**(通道, 音高, 音量, 緩衝區) | `63-fm-towns-sound-effects.md` |
| `sub_34861` | call    sub_34861                ; AH=27h 停止通道 | `63-fm-towns-sound-effects.md` |
| `sub_34876` | call    sub_34876                ; AH=28h 查通道狀態 | `63-fm-towns-sound-effects.md` |
| `sub_34B1B` | 主要函式  /  `sub_313B8`(載入 `U5_BGM.TBL` + `FM_BANK.FMB`)、`sub_3453B`(★ 交出**檔名**)、`sub_34B1B`(★ 遠端呼叫驅動)、`sub_34D3D… | `89-music-playback-is-not-in-the-game-binary.md` |
| `sub_34D3D` | 主要函式  /  `sub_313B8`(載入 `U5_BGM.TBL` + `FM_BANK.FMB`)、`sub_3453B`(★ 交出**檔名**)、`sub_34B1B`(★ 遠端呼叫驅動)、`sub_34D3D… | `89-music-playback-is-not-in-the-game-binary.md` |
| `sub_36D44` | 主要函式  /  `sub_313B8`(載入 `U5_BGM.TBL` + `FM_BANK.FMB`)、`sub_3453B`(★ 交出**檔名**)、`sub_34B1B`(★ 遠端呼叫驅動)、`sub_34D3D… | `89-music-playback-is-not-in-the-game-binary.md` |
| `sub_377A4` | 44 次 `fread`(`sub_377A4`)+ 82 次 `fgetc`(`sub_3806C`),寫入端完全對稱。 | `07-save-format.md` |
| `sub_37A80` | call    sub_37A80            ; fopen | `63-fm-towns-sound-effects.md` |
| `sub_3806C` | 44 次 `fread`(`sub_377A4`)+ 82 次 `fgetc`(`sub_3806C`),寫入端完全對稱。 | `07-save-format.md` |
| `sub_39034` | call    sub_39034            ; fscanf | `63-fm-towns-sound-effects.md` |
| `sub_3944C` | call    sub_3944C                      ; strcat | `67-corpser-and-the-sleeping-party-member.md` |
| `sub_39554` | return sub_39554(up, haystack, n) == 0 ? 0 : -1;   // strncmp(up, haystack, n) | `26-yell-words-of-power-shadowlords.md`, `65-the-wishing-well-easter-egg.md` |
| `sub_39C50` | `sub_17A14`(召喚暗影君主)、`sub_27C98` → `sub_39554` → `sub_39C50`(字串比對) | `26-yell-words-of-power-shadowlords.md` |
| `off_3EF3E` | (`word_3EF44` / `word_3EF42` / `word_3EF3C` / `off_3EF3E+2`),看起來是寬度。 | `20-projectiles.md` |
| `off_3F4A4` | push    off_3F4A4[edi*4]               ; 物品名表 | `67-corpser-and-the-sleeping-party-member.md` |
| `off_3F564` | 0x2D 是誰?** 怪物名表 `off_3F564` 的第 45 筆(位址 `0x3F564 + 45×4 = 0x3F618`): | `67-corpser-and-the-sleeping-party-member.md`, `74-two-swamp-dice-and-an-unnamed-failure.md` |
| `off_408F4` | 主要資料  /  ★ `off_408F4`(38 個縮寫名)、`byte_40BA0`(執行期組出的 38 筆計數)、`off_40B50`(職業名)、`off_40B8C`(狀態名)  / | `94-ztats-has-seventeen-pages.md` |
| `off_40A6C` | 兩位數的擁有量、勾記(0x0F)、藥草名(`off_40A6C[i]`)。按鍵: | `58-rune-input-cast-and-mix.md` |
| `off_40B50` | 主要資料  /  ★ `off_408F4`(38 個縮寫名)、`byte_40BA0`(執行期組出的 38 筆計數)、`off_40B50`(職業名)、`off_40B8C`(狀態名)  / | `94-ztats-has-seventeen-pages.md` |
| `off_40B8C` | 主要資料  /  ★ `off_408F4`(38 個縮寫名)、`byte_40BA0`(執行期組出的 38 筆計數)、`off_40B50`(職業名)、`off_40B8C`(狀態名)  / | `94-ztats-has-seventeen-pages.md` |
| `off_41054` | (城鎮與城堡),後 8 筆就是地牢入口,三張平行表(`off_41054` 名稱 / | `03-scene-entry-and-tile-semantics.md`, `18-dungeons.md` |
| `off_411BC` | `0x19`(25)  /  **the shrine of …**(八德聖壇;名字取自 `off_411BC[9]`,狀態看 `byte_411FC[8]`/`byte_41204[8]`)  /  `sub_1DA1… | `03-scene-entry-and-tile-semantics.md`, `26-yell-words-of-power-shadowlords.md` |
| `off_411DC` | `off_411DC`(真言)、`byte_411FC`/`byte_41204`(聖壇座標)。一路到底都是同一個索引。 | `25-shrines.md`, `26-yell-words-of-power-shadowlords.md`, `28-shadowlords-and-blackthorn.md` |
| `off_4145C` | dword_3EF24 = off_4145C[…]                 ; 店名 "Iolo's Bows" | `08-shops.md`, `93-the-tavern-meal-was-never-haggled.md` |
| `off_4165C` | dword_3EF28 = off_4165C[…]                 ; 店主 "Gwenneth" | `08-shops.md`, `93-the-tavern-meal-was-never-haggled.md` |
| `off_41BA0` | ⚠ 這一步 **grep 反編譯輸出會回零命中** —— 存取是 `off_41BA0[edi*4]` 這種間接形式, | `01-tileset-and-dot16-loader.md` |
| `off_41BB4` | `off_41BB4[0..2]`(3 檔)  /  `0xD6D8` = 55,000 B  /  55,000  / | `01-tileset-and-dot16-loader.md` |
| `off_41BC0` | `off_41BC0[0..7]`(疑為 `MON0–7.16`)  /  `0x1068` = 4,200 B  /  4,200  / | `01-tileset-and-dot16-loader.md` |
| `off_41BE0` | push off_41BE0 → call sub_24A50    ← 載 .16(off_41BE0 = 表的第 16 項 "CREATE.16") | `01-tileset-and-dot16-loader.md` |
| `off_41BE8` | (a) 插圖檔名是 DOS 的。** `off_41BE8` 存的是 `STORY1.16`…`STORY6.16` —— | `24-intro.md` |
| `off_48A88` | off_48A88       dd offset loc_202C      ; DATA XREF: sub_A310+1E↑o | `67-corpser-and-the-sleeping-party-member.md`, `77-a-false-gap-and-the-failed-line.md`, `80-implemented-but-unreachable.md`, `82-the-overworld-world-turn.md` |
| `off_4D5B8` | push    offset off_4D5B8 | `63-fm-towns-sound-effects.md` |
| `off_4FC44` | mov     eax, off_4FC44[eax*4]    ; 檔案 = {TOWNE,DWELLING,CASTLE,KEEP}[(編號-1)/8] | `03-scene-entry-and-tile-semantics.md` |
| `off_4FC90` | `off_4FC98` 與碎片名指標表 `off_4FC90` 重疊(`0x4FC98 = 0x4FC90 + 2×4`), | `36-sandalwood-box-npc-objects.md` |
| `off_4FC98` | `off_4FC98` 與碎片名指標表 `off_4FC90` 重疊(`0x4FC98 = 0x4FC90 + 2×4`), | `36-sandalwood-box-npc-objects.md` |
| `off_55DEC` | 三個名字在 `off_55DEC`:`FAULINEI` / `ASTAROTH` / `NOSFENTOR`。 | `26-yell-words-of-power-shadowlords.md` |
| `off_55DF8` | `off_55DF8`(字)、`byte_41114`/`byte_4113C`(地牢入口座標)、`byte_55E18`(入口地形); | `26-yell-words-of-power-shadowlords.md` |
| `off_55E88` | `off_55E88` 是一張引擎自己的關鍵字表,**掃描順序在記錄的關鍵字之前**: | `06-conversation-script.md` |
| `off_55FEC` | 美德名比對  /  `off_55FEC` 的**四字母前綴**(`hone`)  /  `off_411BC` 的**完整名**(`Honesty`)  / | `25-shrines.md`, `26-yell-words-of-power-shadowlords.md` |
| `off_56E9C` | 酒館的打聽消息**(`sub_21500`):`off_56E9C` 是 16 個四字母關鍵字 | `10-shop-prices-and-trade.md` |
| `off_56F04` | 搭配 `off_56F04`(人名)與 `off_56F88`(地名)回答「某某美德在哪座城」。 | `10-shop-prices-and-trade.md` |
| `off_56F88` | 搭配 `off_56F04`(人名)與 `off_56F88`(地名)回答「某某美德在哪座城」。 | `10-shop-prices-and-trade.md` |
| `byte_3DDB0` | byte_3DDB0(2)  byte_3DDB4(512 名冊)  word_3DFB4  word_3DFB6 | `13-save-writing.md`, `87-bgm-selection-and-the-tbl-tables.md` |
| `byte_3DDB4` | `byte_3DDD1 − 名冊基底 byte_3DDB4 = 0x1D` = `CharRing`,而 `0x2C = 44` 正是 | `06-conversation-script.md`, `13-save-writing.md`, `37-look-signs-and-the-sky.md`, `54-pass-vehicle-verbs-and-the-avatar-line.md`, `61-tavern-lore-and-the-transfer-second-stage.md`, `67-corpser-and-the-sleeping-party-member.md`, `68-troll-bridge-and-collision.md`, `70-hunger-poison-and-the-vanishing-rings.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `byte_3DDB8` | cmp  byte_3DDB8[eax], 6Ah ; 'j' | `16-combat-turns-and-ai.md` |
| `byte_3DDBD` | if (byte_3DDBD[edi*32] == 0x0B) 印 "sir" else 印 "milady" | `61-tavern-lore-and-the-transfer-second-stage.md`, `93-the-tavern-meal-was-never-haggled.md` |
| `byte_3DDBF` | 基底可以獨立驗:`byte_3DDBF` − `byte_3DDB4` = 0x0B = `CharStatus` ✓ | `10-shop-prices-and-trade.md`, `32-guard-challenge.md`, `53-party-tile-and-the-arena-floor.md`, `67-corpser-and-the-sleeping-party-member.md`, `68-troll-bridge-and-collision.md`, `70-hunger-poison-and-the-vanishing-rings.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `byte_3DDC0` | al  = (byte_3DDC0[who*32] >= ecx)          ; ★ 力量,setnl → 是 >= 不是 > | `39-character-creation.md`, `68-troll-bridge-and-collision.md`, `72-ready-had-seven-missing-rules.md` |
| `byte_3DDC1` | mov     byte_3DDC1, al      ; CharDex    (+13) | `39-character-creation.md`, `68-troll-bridge-and-collision.md`, `76-jimmy-in-a-dungeon-disarms-not-unlocks.md` |
| `byte_3DDC2` | movzx edx, byte_3DDC2[買家*32]         ; INT(角色紀錄 offset 0x0E) | `10-shop-prices-and-trade.md`, `37-look-signs-and-the-sky.md`, `39-character-creation.md` |
| `byte_3DDC3` | mov     byte_3DDC3, al      ; CharMP     (+15)  ← 同一個值! | `17-magic.md`, `39-character-creation.md` |
| `byte_3DDCA` | 最低等級**(`byte_3DDCA[角色*32] < 圈數` → 直接失敗) | `17-magic.md`, `19-levelup.md` |
| `byte_3DDCB` | 0x17  /  留在旅店的天數  /  0xFF 或 0..25  /  日期進位那段對 16 名角色跑 `byte_3DDCB[i*32]`  / | `09-items-and-creatures.md`, `42-equipment-slots-ready-wear.md` |
| `byte_3DDCC` | `sub_B274` 對角色讀 `byte_3DDCC[角色*32]`,而 `0x3DDCC − 0x3DDB4 = 0x18`。 | `15-combat-formulas.md`, `42-equipment-slots-ready-wear.md` |
| `byte_3DDCD` | `sub_A360` 只問 `byte_3DDCD` / `byte_3DDCF` / `byte_3DDD0`, | `67-corpser-and-the-sleeping-party-member.md` |
| `byte_3DDCF` | `sub_A360` 只問 `byte_3DDCD` / `byte_3DDCF` / `byte_3DDD0`, | `67-corpser-and-the-sleeping-party-member.md`, `72-ready-had-seven-missing-rules.md` |
| `byte_3DDD0` | `sub_A360` 只問 `byte_3DDCD` / `byte_3DDCF` / `byte_3DDD0`, | `67-corpser-and-the-sleeping-party-member.md` |
| `byte_3DDD1` | 主要資料  /  `byte_3DDD1` = `CharRing`(記錄 0x1D)、`byte_3E09B`(回合計數器)、`byte_3E09E`(模式倒數)、`byte_3E0AE`(當前行動單位)  / | `70-hunger-poison-and-the-vanishing-rings.md`, `72-ready-had-seven-missing-rules.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `byte_3DDD3` | 0x1F  /  留在旅店的標記  /  隊員 0x00、其餘 0xFF  /  入隊時 `sub_1BB5C` 寫 `byte_3DDD3[i*32]` = 0  / | `09-items-and-creatures.md`, `27-codex-and-the-shrine-chamber.md`, `28-shadowlords-and-blackthorn.md`, `42-equipment-slots-ready-wear.md` |
| `byte_3DFB3` | byte_3DFB3 = 0x7F;                            // = 0x3DDB4 + 15×32 + 31 | `27-codex-and-the-shrine-chamber.md`, `28-shadowlords-and-blackthorn.md` |
| `byte_3DFB8` | 0x07 鑰匙  /  `byte_3DFB8` 或 `byte_3DFBD`  /  99  /  ⚠ **品質 ≥ 0x80 走「怪鑰匙」那一欄**,數量取清掉最高位之後的值  / | `10-shop-prices-and-trade.md`, `28-shadowlords-and-blackthorn.md`, `29-npc-behaviour-and-arrest.md`, `33-get-command.md`, `41-jimmy-neworder-gem-ztats.md`, `76-jimmy-in-a-dungeon-disarms-not-unlocks.md`, `94-ztats-has-seventeen-pages.md` |
| `byte_3DFB9` | " Gems......."      byte_3DFB9  寬 2 | `10-shop-prices-and-trade.md`, `33-get-command.md`, `94-ztats-has-seventeen-pages.md` |
| `byte_3DFBA` | " Torches...."      byte_3DFBA  寬 2 | `10-shop-prices-and-trade.md`, `33-get-command.md`, `94-ztats-has-seventeen-pages.md` |
| `byte_3DFBB` | if (byte_3DFBB == 0) { 印 "With what?";    return }   ; ★ 沒抓鉤 | `18-dungeons.md`, `33-get-command.md`, `68-troll-bridge-and-collision.md`, `94-ztats-has-seventeen-pages.md` |
| `byte_3DFBC` | 0x1B 魔毯  /  `byte_3DFBC`  /  99  /  在地點 17 時 `sub_268(0x16)` —— **沒有** `sub_218`  / | `11-map-objects.md`, `33-get-command.md`, `44-use-item.md`, `94-ztats-has-seventeen-pages.md` |
| `byte_3DFBD` | 0x07 鑰匙  /  `byte_3DFB8` 或 `byte_3DFBD`  /  99  /  ⚠ **品質 ≥ 0x80 走「怪鑰匙」那一欄**,數量取清掉最高位之後的值  / | `33-get-command.md`, `44-use-item.md`, `94-ztats-has-seventeen-pages.md` |
| `byte_3DFBE` | if (i == 0x0E && byte_3E08E != byte_3DFBE)         continue | `33-get-command.md`, `96-search-had-three-more-layers.md` |
| `byte_3DFBF` | 而 `sub_1E8D4` 建可用道具清單時,`byte_3DFBF` 落在特殊道具表**第 2 筆 | `33-get-command.md`, `44-use-item.md`, `94-ztats-has-seventeen-pages.md` |
| `byte_3DFC0` | `Absorbed!` 的條件:`byte_3E08A != 'N'` 且 `byte_3DFC0 == 0` 且 `byte_3E0A4 == 0x12` | `17-magic.md`, `33-get-command.md`, `44-use-item.md`, `57-crown-and-sceptre-placement.md`, `67-corpser-and-the-sleeping-party-member.md`, `94-ztats-has-seventeen-pages.md` |
| `byte_3DFC1` | `byte_3DFBB`..`byte_3DFC1` 七個單位元組,**兩端都已經釘死**」—— | `33-get-command.md`, `44-use-item.md`, `57-crown-and-sceptre-placement.md`, `68-troll-bridge-and-collision.md`, `94-ztats-has-seventeen-pages.md` |
| `byte_3DFC4` | 29..31  /  `Shard/Falsehd` `Shard/Hatred` `Shard/Cowrdce`  /  `byte_3DFC4[i]`  /   / | `13-save-writing.md`, `28-shadowlords-and-blackthorn.md`, `33-get-command.md`, `44-use-item.md`, `94-ztats-has-seventeen-pages.md` |
| `byte_3DFC8` | `"Box"` = 37,而 `sub_1E8D4` 把 `byte_3DFC8`..`byte_3DFCD` 抄成 +32..+37。 | `30-ending.md`, `44-use-item.md`, `77-a-false-gap-and-the-failed-line.md`, `79-four-dialogue-opcodes-were-corrupting-the-text.md`, `94-ztats-has-seventeen-pages.md` |
| `byte_3DFC9` | 33  /  `HMS Cape Plan`  /  `byte_3DFC9 != 0 ? 0xFF : 0`  /  ★ 唯一做布林轉換的一筆  / | `33-get-command.md`, `44-use-item.md`, `77-a-false-gap-and-the-failed-line.md`, `83-monster-attacks-and-two-different-wind-rules.md`, `94-ztats-has-seventeen-pages.md` |
| `byte_3DFCA` | 34..37  /  `Sextant` `Pocket Watch` `Black Badge` `Wooden Box`  /  `byte_3DFCA`..`byte_3DFCD`  /   / | `44-use-item.md`, `77-a-false-gap-and-the-failed-line.md`, `79-four-dialogue-opcodes-were-corrupting-the-text.md`, `94-ztats-has-seventeen-pages.md` |
| `byte_3DFCB` | 19  /  Pocket Watch  /  35  /  `byte_3DFCB`  / | `44-use-item.md`, `77-a-false-gap-and-the-failed-line.md` |
| `byte_3DFCC` | 黑徽章的存檔位移**:`sub_1E8D4` 讀 `byte_3DFCC`,但存檔 0x0216..0x0218 | `44-use-item.md`, `77-a-false-gap-and-the-failed-line.md`, `79-four-dialogue-opcodes-were-corrupting-the-text.md`, `94-ztats-has-seventeen-pages.md` |
| `byte_3DFCD` | 34..37  /  `Sextant` `Pocket Watch` `Black Badge` `Wooden Box`  /  `byte_3DFCA`..`byte_3DFCD`  /   / | `30-ending.md`, `33-get-command.md`, `44-use-item.md`, `77-a-false-gap-and-the-failed-line.md`, `79-four-dialogue-opcodes-were-corrupting-the-text.md`, `94-ztats-has-seventeen-pages.md` |
| `byte_3DFD0` | 其餘 ≤ 0x0C  /  `byte_3DFD0[品質]`  /  99  /  **品質就是裝備編號**;箭矢(27)與弩矢(29)一次 **+5**  / | `10-shop-prices-and-trade.md`, `13-save-writing.md`, `17-magic.md`, `22-moongates.md`, `33-get-command.md`, `49-command-table-and-two-empty-keys.md`, `60-command-echo-and-menu-keys.md`, `72-ready-had-seven-missing-rules.md`, `79-four-dialogue-opcodes-were-corrupting-the-text.md`, `94-ztats-has-seventeen-pages.md` |
| `byte_3DFEB` | 3. 弓(0x1A)與魔法弓(0x24)裝備前查 `byte_3DFEB`、十字弓(0x1C)查 `byte_3DFED`, | `72-ready-had-seven-missing-rules.md` |
| `byte_3DFED` | 3. 弓(0x1A)與魔法弓(0x24)裝備前查 `byte_3DFEB`、十字弓(0x1C)查 `byte_3DFED`, | `72-ready-had-seven-missing-rules.md` |
| `byte_3DFF7` | if (i == 0x0F && byte_3DFF7 == 0 && sub_2B360(…) == 0) continue | `96-search-had-three-more-layers.md` |
| `byte_3E000` | (`byte_3E000 ↔ 0x024A`、`byte_3E060 ↔ 0x02AA`,兩端夾住中間 0x60 B): | `13-save-writing.md`, `17-magic.md`, `33-get-command.md`, `45-fire-and-mix.md`, `71-the-use-list-had-29-empty-slots.md`, `94-ztats-has-seventeen-pages.md` |
| `byte_3E030` | 0..7  /  `*VL` `*RH` `*IS` `*IA` `*IQW` `*KXC` `*IMC` `*AT`  /  `byte_3E030[i]`  /  卷軸,名字是咒語縮寫  / | `13-save-writing.md`, `33-get-command.md`, `71-the-use-list-had-29-empty-slots.md`, `94-ztats-has-seventeen-pages.md` |
| `byte_3E038` | 0x03 藥水  /  `byte_3E038[品質]`  /  99  /  **品質選顏色**(藍/黃/紅/綠/橙/紫/黑/白)  / | `13-save-writing.md`, `33-get-command.md`, `94-ztats-has-seventeen-pages.md` |
| `byte_3E040` | ptr = tile_at(byte_3E040[i], byte_3E048[i])   ; ★★ 月石的 X / Y | `13-save-writing.md`, `22-moongates.md`, `33-get-command.md`, `71-the-use-list-had-29-empty-slots.md`, `86-tile-0xDC-is-an-open-moongate.md`, `96-search-had-three-more-layers.md` |
| `byte_3E048` | ptr = tile_at(byte_3E040[i], byte_3E048[i])   ; ★★ 月石的 X / Y | `13-save-writing.md`, `22-moongates.md`, `71-the-use-list-had-29-empty-slots.md`, `86-tile-0xDC-is-an-open-moongate.md`, `96-search-had-three-more-layers.md` |
| `byte_3E050` | 21..28  /  `(0`..`(7`  /  `byte_3E050[i] == 0xFF ? 0xFF : 0`  /  ⚠⚠ **沒寫完的佔位名**  / | `13-save-writing.md`, `22-moongates.md`, `33-get-command.md`, `44-use-item.md`, `71-the-use-list-had-29-empty-slots.md`, `86-tile-0xDC-is-an-open-moongate.md`, `94-ztats-has-seventeen-pages.md`, `96-search-had-three-more-layers.md` |
| `byte_3E058` | if (byte_3E058[i] != 樓層)    continue       ;         樓層 | `13-save-writing.md`, `22-moongates.md`, `71-the-use-list-had-29-empty-slots.md`, `86-tile-0xDC-is-an-open-moongate.md`, `96-search-had-three-more-layers.md` |
| `byte_3E060` | (`byte_3E000 ↔ 0x024A`、`byte_3E060 ↔ 0x02AA`,兩端夾住中間 0x60 B): | `10-shop-prices-and-trade.md`, `13-save-writing.md`, `17-magic.md`, `22-moongates.md`, `33-get-command.md`, `58-rune-input-cast-and-mix.md`, `71-the-use-list-had-29-empty-slots.md`, `94-ztats-has-seventeen-pages.md` |
| `byte_3E068` | 主要資料  /  ★★ 113 筆 × 六個平行陣列(0x55A64 / 0x55AD8 / 0x55B4C / 0x55BC0 / 0x55C34 / 0x55CA8)、三個採藥點(0x55A4C 起)、`byte_3… | `96-search-had-three-more-layers.md` |
| `byte_3E06B` | 主要資料  /  **`byte_5FFEC` = 八筆權重表**、`byte_3E0A3`(地點碼)、`byte_3E06B`(隊伍人數)  / | `07-save-format.md`, `28-shadowlords-and-blackthorn.md`, `74-two-swamp-dice-and-an-unnamed-failure.md`, `86-tile-0xDC-is-an-open-moongate.md`, `91-chest-traps-four-kinds-and-their-weights.md`, `93-the-tavern-meal-was-never-haggled.md`, `97-set-active-player.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `byte_3E06C` | 主要資料  /  ★★ 113 筆 × 六個平行陣列(0x55A64 / 0x55AD8 / 0x55B4C / 0x55BC0 / 0x55C34 / 0x55CA8)、三個採藥點(0x55A4C 起)、`byte_3… | `96-search-had-three-more-layers.md` |
| `byte_3E08A` | `Absorbed!` 的條件:`byte_3E08A != 'N'` 且 `byte_3DFC0 == 0` 且 `byte_3E0A4 == 0x12` | `04-npc-schedule-and-clock.md`, `100-the-last-three-and-the-collapsed-entrance-gate.md`, `16-combat-turns-and-ai.md`, `17-magic.md`, `26-yell-words-of-power-shadowlords.md`, `32-guard-challenge.md`, `38-terrain-movement-cost.md`, `44-use-item.md`, `67-corpser-and-the-sleeping-party-member.md`, `70-hunger-poison-and-the-vanishing-rings.md`, `71-the-use-list-had-29-empty-slots.md`, `80-implemented-but-unreachable.md`, `82-the-overworld-world-turn.md`, `92-white-noise-shim-and-ambient-sound.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `byte_3E08B` | `byte_3E08B` / `byte_3E09B` / `byte_3E0F0`  /  `SaveActiveMemberOffset` / `SaveTurnCounterOffset` / `SaveRooms… | `100-the-last-three-and-the-collapsed-entrance-gate.md`, `37-look-signs-and-the-sky.md`, `52-one-coordinate-pair-one-tile-accessor.md`, `70-hunger-poison-and-the-vanishing-rings.md`, `75-open-did-nothing-outside-dungeons.md`, `93-the-tavern-meal-was-never-haggled.md`, `95-the-dungeon-has-its-own-look.md`, `97-set-active-player.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `byte_3E08C` | `byte_3E08B` 指定行動者  /  **0x02D5**  /  1  /  `byte_3E08C` = 0x02D6 減 1  /  `INIT.GAM` 是 **0xFF**,正是它的哨兵值  / | `02-movement-and-tile-flags.md`, `03-scene-entry-and-tile-semantics.md`, `08-shops.md`, `10-shop-prices-and-trade.md`, `100-the-last-three-and-the-collapsed-entrance-gate.md`, `11-map-objects.md`, `18-dungeons.md`, `22-moongates.md`, `26-yell-words-of-power-shadowlords.md`, `28-shadowlords-and-blackthorn.md`, `34-ending-trigger.md`, `38-terrain-movement-cost.md`, `50-hole-up-camp-sleep-repair.md`, `52-one-coordinate-pair-one-tile-accessor.md`, `53-party-tile-and-the-arena-floor.md`, `66-hexrays-truncation-audit.md`, `68-troll-bridge-and-collision.md`, `74-two-swamp-dice-and-an-unnamed-failure.md`, `80-implemented-but-unreachable.md`, `82-the-overworld-world-turn.md`, `83-monster-attacks-and-two-different-wind-rules.md`, `84-the-sailing-gate-and-the-wind-table.md`, `86-tile-0xDC-is-an-open-moongate.md`, `87-bgm-selection-and-the-tbl-tables.md` |
| `byte_3E08D` | byte_3E08D 月   > 13  → 設回 1 並進位 | `04-npc-schedule-and-clock.md` |
| `byte_3E08E` | if (byte_3E068[i] == byte_3E08E)  continue    ; ★ 今天已經採過這一點 | `04-npc-schedule-and-clock.md`, `96-search-had-three-more-layers.md` |
| `byte_3E08F` | } else if (byte_3E08F ∈ {6, 12, 18}) {                      ; ★ 6 / 12 / 18 點 | `04-npc-schedule-and-clock.md`, `10-shop-prices-and-trade.md`, `100-the-last-three-and-the-collapsed-entrance-gate.md`, `22-moongates.md`, `29-npc-behaviour-and-arrest.md`, `47-move-modes-and-time-of-day.md`, `50-hole-up-camp-sleep-repair.md`, `70-hunger-poison-and-the-vanishing-rings.md`, `82-the-overworld-world-turn.md`, `86-tile-0xDC-is-an-open-moongate.md`, `92-white-noise-shim-and-ambient-sound.md`, `96-search-had-three-more-layers.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `byte_3E090` | if (byte_3E08F == byte_3E090) return                        ; 這個小時已經結算過 | `70-hunger-poison-and-the-vanishing-rings.md`, `92-white-noise-shim-and-ambient-sound.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `byte_3E091` | if (byte_3E08F == 0 && byte_3E091 < 0Ah) esi = 1   ; ★★ 午夜前十分鐘不傳送 | `04-npc-schedule-and-clock.md`, `86-tile-0xDC-is-an-open-moongate.md` |
| `byte_3E092` | 每 10 個單位行動 = 遊戲內 1 分鐘**(`byte_3E092` 數到 10 → `sub_29304(1)`)。 | `16-combat-turns-and-ai.md` |
| `byte_3E093` | if ((n % 3) > byte_3E093) 走一步 else { byte_3E093 = 0; 收尾 } | `23-wind-and-sailing.md`, `83-monster-attacks-and-two-different-wind-rules.md` |
| `byte_3E094` | 主要資料  /  `byte_60038`(八相)、`byte_3E094`(剩餘鐘響)、`byte_5FFFC`(旋律游標)、**0x60060 音高表**、**0x6006C 53 步旋律**  / | `92-white-noise-shim-and-ambient-sound.md` |
| `byte_3E095` | 中午前讀 `byte_3E095`、中午後讀 `byte_3E096`(`cmp byte_3E08F, 0Ch`), | `22-moongates.md`, `86-tile-0xDC-is-an-open-moongate.md` |
| `byte_3E096` | 中午前讀 `byte_3E095`、中午後讀 `byte_3E096`(`cmp byte_3E08F, 0Ch`), | `22-moongates.md`, `86-tile-0xDC-is-an-open-moongate.md` |
| `byte_3E097` | 淡出計數器 `byte_3E097`  /  `game.State.MoongateFrame`  /  ✅ 開關生效 + ✅ 繪圖用它當動畫格(§7);⬜ 存檔位移未定位  / | `86-tile-0xDC-is-an-open-moongate.md` |
| `byte_3E098` | `byte_3E09B` 回合計數  /  **0x02E5**  /  1  /  `byte_3E098`(業報)= 0x02E2 加 3  /  `INIT.GAM` 是 **0**  / | `06-conversation-script.md`, `100-the-last-three-and-the-collapsed-entrance-gate.md`, `19-levelup.md`, `69-the-overworld-cannon-did-nothing.md`, `75-open-did-nothing-outside-dungeons.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `byte_3E09B` | `byte_3E08B` / `byte_3E09B` / `byte_3E0F0`  /  `SaveActiveMemberOffset` / `SaveTurnCounterOffset` / `SaveRooms… | `100-the-last-three-and-the-collapsed-entrance-gate.md`, `70-hunger-poison-and-the-vanishing-rings.md`, `79-four-dialogue-opcodes-were-corrupting-the-text.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `byte_3E09C` | `byte_3E09C`  /  `game.State.RestCooldown` + `(*State).tickHourly`  / | `51-closing-four-known-gaps.md` |
| `byte_3E09E` | 主要資料  /  `byte_3DDD1` = `CharRing`(記錄 0x1D)、`byte_3E09B`(回合計數器)、`byte_3E09E`(模式倒數)、`byte_3E0AE`(當前行動單位)  / | `17-magic.md`, `32-guard-challenge.md`, `70-hunger-poison-and-the-vanishing-rings.md`, `71-the-use-list-had-29-empty-slots.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `byte_3E09F` | byte_3E09F == 0(武器):攻擊碼 0x23 / 0x27 / 0x28 必中   ← 引擎的 AlwaysHitWeapons | `15-combat-formulas.md`, `17-magic.md`, `74-two-swamp-dice-and-an-unnamed-failure.md` |
| `byte_3E0A0` | 傷害減半(0x0020)      → 除以 2(除非 byte_3E0A0 成立,那個旗標還沒解) | `16-combat-turns-and-ai.md` |
| `byte_3E0A2` | if (byte_3E0A2 == 0) 走一步                      ; ★★ 無風 → 照走,不查表 | `23-wind-and-sailing.md`, `66-hexrays-truncation-audit.md`, `83-monster-attacks-and-two-different-wind-rules.md`, `84-the-sailing-gate-and-the-wind-table.md` |
| `byte_3E0A3` | if (byte_3E0A3 != 0x1E && byte_3E0A3 != 0x1F && byte_3E0A3 != 0x20) { puts("No effect!"); return 1; } | `03-scene-entry-and-tile-semantics.md`, `04-npc-schedule-and-clock.md`, `10-shop-prices-and-trade.md`, `17-magic.md`, `18-dungeons.md`, `26-yell-words-of-power-shadowlords.md`, `27-codex-and-the-shrine-chamber.md`, `28-shadowlords-and-blackthorn.md`, `29-npc-behaviour-and-arrest.md`, `31-line-of-sight.md`, `45-fire-and-mix.md`, `48-dungeon-wandering-monster-and-arena.md`, `49-command-table-and-two-empty-keys.md`, `50-hole-up-camp-sleep-repair.md`, `52-one-coordinate-pair-one-tile-accessor.md`, `54-pass-vehicle-verbs-and-the-avatar-line.md`, `65-the-wishing-well-easter-egg.md`, `71-the-use-list-had-29-empty-slots.md`, `72-ready-had-seven-missing-rules.md`, `74-two-swamp-dice-and-an-unnamed-failure.md`, `75-open-did-nothing-outside-dungeons.md`, `76-jimmy-in-a-dungeon-disarms-not-unlocks.md`, `78-the-bottom-of-a-dungeon-opens-into-the-underworld.md`, `80-implemented-but-unreachable.md`, `81-three-mode-loops-are-mutually-exclusive.md`, `86-tile-0xDC-is-an-open-moongate.md`, `87-bgm-selection-and-the-tbl-tables.md`, `88-the-load-window-culling-and-ship-cannon.md`, `91-chest-traps-four-kinds-and-their-weights.md`, `93-the-tavern-meal-was-never-haggled.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `byte_3E0A4` | `Absorbed!` 的條件:`byte_3E08A != 'N'` 且 `byte_3DFC0 == 0` 且 `byte_3E0A4 == 0x12` | `67-corpser-and-the-sleeping-party-member.md` |
| `byte_3E0A5` | 樓層增減(`sub_758`)  /  `byte_3E0A5 = 1` / `= -1`  /  `inc` / `dec byte_3E0A5`  /  照抄的話只能在 1F 與 B1F 之間跳  / | `03-scene-entry-and-tile-semantics.md`, `100-the-last-three-and-the-collapsed-entrance-gate.md`, `11-map-objects.md`, `18-dungeons.md`, `26-yell-words-of-power-shadowlords.md`, `28-shadowlords-and-blackthorn.md`, `29-npc-behaviour-and-arrest.md`, `34-ending-trigger.md`, `65-the-wishing-well-easter-egg.md`, `71-the-use-list-had-29-empty-slots.md`, `75-open-did-nothing-outside-dungeons.md`, `78-the-bottom-of-a-dungeon-opens-into-the-underworld.md`, `80-implemented-but-unreachable.md`, `81-three-mode-loops-are-mutually-exclusive.md`, `82-the-overworld-world-turn.md`, `83-monster-attacks-and-two-different-wind-rules.md`, `86-tile-0xDC-is-an-open-moongate.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `byte_3E0A6` | 邊界旗標  /  每個方向一個常數(西/北 = 1,東/南 = 0)  /  `cmp byte_3E0A6, 1 / jnb` 等四組比較  /  照抄的話往東往南永遠出不了城  / | `03-scene-entry-and-tile-semantics.md`, `10-shop-prices-and-trade.md`, `18-dungeons.md`, `26-yell-words-of-power-shadowlords.md`, `28-shadowlords-and-blackthorn.md`, `29-npc-behaviour-and-arrest.md`, `34-ending-trigger.md`, `40-push-and-the-main-menu.md`, `48-dungeon-wandering-monster-and-arena.md`, `52-one-coordinate-pair-one-tile-accessor.md`, `65-the-wishing-well-easter-egg.md`, `71-the-use-list-had-29-empty-slots.md`, `75-open-did-nothing-outside-dungeons.md`, `78-the-bottom-of-a-dungeon-opens-into-the-underworld.md`, `82-the-overworld-world-turn.md`, `83-monster-attacks-and-two-different-wind-rules.md`, `84-the-sailing-gate-and-the-wind-table.md` |
| `byte_3E0A7` | if (byte_3E0A7 != 4)                                   // ★ 玩家 Y 剛好是 4 就整段跳過 | `03-scene-entry-and-tile-semantics.md`, `26-yell-words-of-power-shadowlords.md`, `28-shadowlords-and-blackthorn.md`, `29-npc-behaviour-and-arrest.md`, `34-ending-trigger.md`, `48-dungeon-wandering-monster-and-arena.md`, `52-one-coordinate-pair-one-tile-accessor.md`, `65-the-wishing-well-easter-egg.md`, `71-the-use-list-had-29-empty-slots.md`, `75-open-did-nothing-outside-dungeons.md`, `78-the-bottom-of-a-dungeon-opens-into-the-underworld.md`, `81-three-mode-loops-are-mutually-exclusive.md`, `82-the-overworld-world-turn.md`, `83-monster-attacks-and-two-different-wind-rules.md`, `84-the-sailing-gate-and-the-wind-table.md` |
| `byte_3E0A8` | 0x02F2  /  16  /  `byte_3E0A8..B7`  /  逐一單位元組  / | `26-yell-words-of-power-shadowlords.md` |
| `byte_3E0AB` | `byte_3E0AB` / `byte_3E0AC`  /  載入視窗的左上角**什麼時候更新**(每步一格還是整塊跳)。引擎目前用「隊伍為中心的 32×32」近似,差異會影響怪出現在哪一側  / | `17-magic.md`, `38-terrain-movement-cost.md`, `82-the-overworld-world-turn.md`, `86-tile-0xDC-is-an-open-moongate.md`, `88-the-load-window-culling-and-ship-cannon.md` |
| `byte_3E0AC` | `byte_3E0AB` / `byte_3E0AC`  /  載入視窗的左上角**什麼時候更新**(每步一格還是整塊跳)。引擎目前用「隊伍為中心的 32×32」近似,差異會影響怪出現在哪一側  / | `38-terrain-movement-cost.md`, `82-the-overworld-world-turn.md`, `86-tile-0xDC-is-an-open-moongate.md`, `88-the-load-window-culling-and-ship-cannon.md` |
| `byte_3E0AD` | 三個攻擊咒語的傷害也還沒逆到底(`sub_189E4` 只把攻擊碼寫進 `byte_3E0AD`, | `16-combat-turns-and-ai.md`, `17-magic.md` |
| `byte_3E0AE` | 主要資料  /  `byte_3DDD1` = `CharRing`(記錄 0x1D)、`byte_3E09B`(回合計數器)、`byte_3E09E`(模式倒數)、`byte_3E0AE`(當前行動單位)  / | `34-ending-trigger.md`, `52-one-coordinate-pair-one-tile-accessor.md`, `67-corpser-and-the-sleeping-party-member.md`, `71-the-use-list-had-29-empty-slots.md`, `95-the-dungeon-has-its-own-look.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `byte_3E0B0` | sub_42CC   (戰鬥)     … cmp byte_3E0B0, 4Dh ; jnz …  → call sub_135FC | `34-ending-trigger.md`, `52-one-coordinate-pair-one-tile-accessor.md`, `78-the-bottom-of-a-dungeon-opens-into-the-underworld.md` |
| `byte_3E0B1` | 主要函式  /  `sub_18380`(ESC = 撤離)、`sub_A360`(戰鬥分派器)、`sub_42CC`(進地牢房間)、`sub_A9EC`(`byte_3E0B3` 的設定處)、`sub_2E364`(戰… | `34-ending-trigger.md`, `48-dungeon-wandering-monster-and-arena.md`, `50-hole-up-camp-sleep-repair.md`, `52-one-coordinate-pair-one-tile-accessor.md`, `73-escape-is-a-key-and-it-has-two-gates.md` |
| `byte_3E0B2` | `byte_3E0B2` 剩下 **0x10** 沒解(`sub_AE20` 設的,那是怪物移動那一支)。 | `16-combat-turns-and-ai.md`, `67-corpser-and-the-sleeping-party-member.md`, `72-ready-had-seven-missing-rules.md` |
| `byte_3E0B3` | 主要函式  /  `sub_18380`(ESC = 撤離)、`sub_A360`(戰鬥分派器)、`sub_42CC`(進地牢房間)、`sub_A9EC`(`byte_3E0B3` 的設定處)、`sub_2E364`(戰… | `72-ready-had-seven-missing-rules.md`, `73-escape-is-a-key-and-it-has-two-gates.md` |
| `byte_3E0B4` | ⬜ `sub_2B740(n)`(沼澤中毒後呼叫)= `if (byte_3E0B4 != 0) 重畫 n 次`。 | `81-three-mode-loops-are-mutually-exclusive.md` |
| `byte_3E0B5` | `byte_3E0B5` 是視野平方半徑**(`docs/re/31` 已解:`sub_29304` 每分鐘算它)。 | `100-the-last-three-and-the-collapsed-entrance-gate.md`, `31-line-of-sight.md` |
| `byte_3E0B6` | `sub_1D310(al)` 就一行 `byte_3E0B6 = al`;`sub_1D31C(模式, 回合, 音效)` 也就三行: | `17-magic.md`, `49-command-table-and-two-empty-keys.md`, `71-the-use-list-had-29-empty-slots.md` |
| `byte_3E0B7` | 2. **看不見就搜不到。** `byte_3E0B6`(火把)與 `byte_3E0B7`(In Lor)兩個 | `17-magic.md`, `49-command-table-and-two-empty-keys.md` |
| `byte_3E0B8` | 0x0302  /  32  /  `byte_3E0B8`  /  一整塊  / | `17-magic.md`, `26-yell-words-of-power-shadowlords.md` |
| `byte_3E0D8` | 末日地牢入口的三個閘門(`sub_2D564` 要 `byte_3E0D8 & 3E0D9 & 3E0DA >= 0x80`, | `18-dungeons.md`, `26-yell-words-of-power-shadowlords.md`, `28-shadowlords-and-blackthorn.md`, `33-get-command.md` |
| `byte_3E0DB` | if (byte_3E0DB != i) return;                    // ← 現身的不是這一位:**沉默** | `26-yell-words-of-power-shadowlords.md`, `28-shadowlords-and-blackthorn.md` |
| `byte_3E0DC` | for (i = 0; i < 8 && !(byte_3E0DC & (1 << i)); i++) ;   // ★ 由小到大,停在第一個 | `25-shrines.md`, `26-yell-words-of-power-shadowlords.md`, `27-codex-and-the-shrine-chamber.md`, `81-three-mode-loops-are-mutually-exclusive.md` |
| `byte_3E0DE` | 0x0328**  /  2  /  `byte_3E0DE`  /  已在寶典上讀到的美德(**寶典 `sub_1D850` 設的,不是聖壇**)  / | `25-shrines.md`, `26-yell-words-of-power-shadowlords.md`, `27-codex-and-the-shrine-chamber.md` |
| `byte_3E0E0` | ⇒ 喊力量之言(`xor byte_3E0E0[i], 80h`,`docs/re/26` §3.2)是把 0 變成 0x80 = | `26-yell-words-of-power-shadowlords.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `byte_3E0E8` | `byte_3E0F0` 房間清除  /  **0x033A**  /  **14**  /  `byte_3E0E8` = 0x0332 加 8  /  全 0(開局沒清過房間) / | `100-the-last-three-and-the-collapsed-entrance-gate.md`, `25-shrines.md`, `26-yell-words-of-power-shadowlords.md`, `28-shadowlords-and-blackthorn.md`, `29-npc-behaviour-and-arrest.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `byte_3E0F0` | `byte_3E08B` / `byte_3E09B` / `byte_3E0F0`  /  `SaveActiveMemberOffset` / `SaveTurnCounterOffset` / `SaveRooms… | `100-the-last-three-and-the-collapsed-entrance-gate.md`, `29-npc-behaviour-and-arrest.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `byte_3E100` | `byte_3E0E8`(8)→`byte_3E0F0`(14)→`byte_3E100`/`3E120`/`3E140`(各 32) | `29-npc-behaviour-and-arrest.md` |
| `byte_3E160` | →`byte_3E160..3E16B`(12 個單位元組)→`dword_3E16C`(512)→ **這一段**。 | `29-npc-behaviour-and-arrest.md` |
| `byte_3E161` | sub_2B64C(byte_3E161, byte_3E162, byte_3E163)    ; 把 tile 寫回去 | `69-the-overworld-cannon-did-nothing.md`, `75-open-did-nothing-outside-dungeons.md` |
| `byte_3E162` | sub_2B64C(byte_3E161, byte_3E162, byte_3E163)    ; 把 tile 寫回去 | `75-open-did-nothing-outside-dungeons.md` |
| `byte_3E163` | sub_2B64C(byte_3E161, byte_3E162, byte_3E163)    ; 把 tile 寫回去 | `75-open-did-nothing-outside-dungeons.md` |
| `byte_3E164` | byte_3E164 = 4             ; ★ 倒數 | `75-open-did-nothing-outside-dungeons.md` |
| `byte_3E165` | mov   byte_3E165, dl | `11-map-objects.md` |
| `byte_3E166` | mov   byte_3E166, al | `11-map-objects.md` |
| `byte_3E167` | if (沒揚帆) {                                   ; byte_3E167 == 0 | `54-pass-vehicle-verbs-and-the-avatar-line.md`, `68-troll-bridge-and-collision.md`, `83-monster-attacks-and-two-different-wind-rules.md` |
| `byte_3E168` | `byte_3E168` 設 1 之後誰讀它。 | `68-troll-bridge-and-collision.md` |
| `byte_3E169` | dec     byte_3E169             ; ★ 踉蹺了才扣一次 | `70-hunger-poison-and-the-vanishing-rings.md` |
| `byte_3E16A` | sub_48C();                      // byte_3E16A = 盤據這裡的是第幾位;並讓牠現身 | `28-shadowlords-and-blackthorn.md` |
| `byte_3E3AF` | 0x0E 檀香木盒  /  `byte_3DFCD`  /  —  /  另外寫死 `byte_3E3AF  / = 0x80`(見 `docs/re/36`)  / | `33-get-command.md` |
| `byte_3E570` | sub_2C740(file, edi,       0x200, byte_3E570)   ; 512 B  32 × 16 B 排程 | `04-npc-schedule-and-clock.md`, `29-npc-behaviour-and-arrest.md` |
| `byte_3E579` | `sub_2D944`、`sub_92C0`(NPC 腳下的 tile 與 `byte_3E579` 排程欄,疑為門 / 床)、 | `100-the-last-three-and-the-collapsed-entrance-gate.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `byte_3E57C` | if (byte_3E57C[npc*16 + esi]) ebx = 1;      // 有作息的居民 | `28-shadowlords-and-blackthorn.md` |
| `byte_3E970` | `byte_3E970[npc*32]` —— 路徑本身,**(步數, 方向) 成對**,共 16 段 | `12-npc-movement.md` |
| `byte_3EDB0` | sub_2C740(file, edi+0x200, 0x20,  byte_3EDB0)   ;  32 B  每個 NPC 的生物編號 | `04-npc-schedule-and-clock.md`, `07-save-format.md`, `12-npc-movement.md`, `26-yell-words-of-power-shadowlords.md`, `28-shadowlords-and-blackthorn.md`, `29-npc-behaviour-and-arrest.md`, `36-sandalwood-box-npc-objects.md` |
| `byte_3EDD0` | byte_3EDD0 = ((ai == 4  /  /  ai == 5) && 對話號碼 != 0) ? 't' : 'a'; | `29-npc-behaviour-and-arrest.md`, `50-hole-up-camp-sleep-repair.md` |
| `byte_3EDD1` | byte_3EDD1 = npc; | `29-npc-behaviour-and-arrest.md` |
| `byte_3EE14` | 只把 4 / 5(上下樓)寫進 `byte_3EE14`,所以 **`byte_3EE15` 永遠是 0..3**。 | `48-dungeon-wandering-monster-and-arena.md` |
| `byte_3EE15` | 主要資料  /  `dword_3E16C`(地牢層緩衝)、`byte_3EE15`(朝向)、`byte_3EE16`、★ `byte_3E08B`(當前選定的隊員)、`word_3E086`/`word_3E088`(… | `21-chests-fields-locks.md`, `48-dungeon-wandering-monster-and-arena.md`, `95-the-dungeon-has-its-own-look.md` |
| `byte_3EE16` | 主要資料  /  `dword_3E16C`(地牢層緩衝)、`byte_3EE15`(朝向)、`byte_3EE16`、★ `byte_3E08B`(當前選定的隊員)、`word_3E086`/`word_3E088`(… | `48-dungeon-wandering-monster-and-arena.md`, `95-the-dungeon-has-its-own-look.md` |
| `byte_3EE17` | 種類碼取 `byte_3EE17 <= 0x7F ? 0x25 : 0x29`、耐久寫死 **0x63 = 99**。 | `11-map-objects.md`, `88-the-load-window-culling-and-ship-cannon.md` |
| `byte_3EE18` | dword_65338 = byte_3EE18; return }   ; ★ 有有效曲號就直接播 | `07-save-format.md`, `87-bgm-selection-and-the-tbl-tables.md` |
| `byte_3EF1B` | ★ `byte_3EF1B` = `[0, 4, 9, 15, 8, 17, 10, 11, 10, 0, 0, …]` —— | `28-shadowlords-and-blackthorn.md` |
| `byte_3F050` | `sub_B398` 印證了這件事:它取 `byte_3F050[生物*8]` 當「力量那一項」、 | `15-combat-formulas.md` |
| `byte_3F052` | `byte_3F052[生物*8]` 當「智力那一項」,而角色走的是紀錄的 0x0C 與 0x0E | `15-combat-formulas.md`, `17-magic.md` |
| `byte_3F053` | +3  /  護甲(減傷)  /  `sub_B274` 讀 `byte_3F053[生物*8]`  / | `15-combat-formulas.md` |
| `byte_3F054` | +4  /  攻擊力  /  `sub_B274` 讀 `byte_3F054[生物*8]`  / | `15-combat-formulas.md` |
| `byte_3F055` | movzx   esi, byte_3F055[eax*8]   ; 生命上限 | `16-combat-turns-and-ai.md`, `67-corpser-and-the-sleeping-party-member.md` |
| `byte_3F056` | 同一段程式碼也證實了 `byte_3F056`(怪物屬性 +6)就是**隻數上限**: | `18-dungeons.md` |
| `byte_3F290` | `byte_3F290` 是**武器傷害表**(`docs/re/15`:`DATA.OVL` 0x160C,48 B)—— | `67-corpser-and-the-sleeping-party-member.md` |
| `byte_3F2F8` | 後面接的是一整套目標選取 + 效果分派(`byte_3F2F8` / `byte_3F330` 兩張 256 格表)。 | `15-combat-formulas.md`, `17-magic.md` |
| `byte_3F330` | 後面接的是一整套目標選取 + 效果分派(`byte_3F2F8` / `byte_3F330` 兩張 256 格表)。 | `17-magic.md` |
| `byte_3F398` | 203  /  6×32 + 11  /  敵人入場 X ×16  /  `byte_3F398`  / | `14-combat-maps.md` |
| `byte_3F3A8` | 235  /  7×32 + 11  /  敵人入場 Y ×16  /  `byte_3F3A8`  / | `14-combat-maps.md` |
| `byte_3F3B8` | 107  /  3×32 + 11  /  隊員入場 X ×6  /  `byte_3F3B8`  / | `14-combat-maps.md` |
| `byte_3F3C8` | `byte_3F3C8`  /  `game.campAmbushCreature`  / | `51-closing-four-known-gaps.md` |
| `byte_3F3D0` | `byte_3F3D0` / `byte_3F3D8`  /  `u5data.dungeonMonsterIndex` / `dungeonMonsterView`  / | `48-dungeon-wandering-monster-and-arena.md`, `51-closing-four-known-gaps.md` |
| `byte_3F3D8` | `byte_3F3D0` / `byte_3F3D8`  /  `u5data.dungeonMonsterIndex` / `dungeonMonsterView`  / | `48-dungeon-wandering-monster-and-arena.md` |
| `byte_3F6E4` | `memset(byte_3F6E4, 0xFF, 0x160)` —— 0x160 = 352 = **11 列 × 32 stride** | `03-scene-entry-and-tile-semantics.md`, `20-projectiles.md`, `31-line-of-sight.md`, `86-tile-0xDC-is-an-open-moongate.md` |
| `byte_3F769` | byte_3F769 北   byte_3F78A 東   byte_3F7A9 南   byte_3F788 西 | `26-yell-words-of-power-shadowlords.md`, `45-fire-and-mix.md` |
| `byte_3F788` | byte_3F769 北   byte_3F78A 東   byte_3F7A9 南   byte_3F788 西 | `26-yell-words-of-power-shadowlords.md`, `45-fire-and-mix.md` |
| `byte_3F789` | `byte_3F789` 看起來像一個單獨的 byte,但它被 `[32*dy + dx]` 這樣索引(dy,dx ∈ −1..1), | `02-movement-and-tile-flags.md`, `03-scene-entry-and-tile-semantics.md`, `35-harp-and-the-secret-door.md`, `45-fire-and-mix.md`, `68-troll-bridge-and-collision.md` |
| `byte_3F78A` | byte_3F769 北   byte_3F78A 東   byte_3F7A9 南   byte_3F788 西 | `26-yell-words-of-power-shadowlords.md`, `45-fire-and-mix.md` |
| `byte_3F7A9` | `docs/re/03` §7 已經定過:視窗緩衝裡**玩家那一格**;`byte_3F7A9` 比它大 0x20, | `26-yell-words-of-power-shadowlords.md`, `35-harp-and-the-secret-door.md`, `45-fire-and-mix.md`, `68-troll-bridge-and-collision.md` |
| `byte_3F844` | `sub_C414` / `sub_1DA10` / `sub_135FC`  /  `sub_2C740("MISCMAPS.DAT", byte_3F844, 0xB0, 位移)` —— 石室整份載入  / | `03-picture-files.md`, `34-ending-trigger.md` |
| `byte_3F854` | 資料  /  `byte_3E0B0`、`byte_3F844`/`byte_3F854`、地形 0x3C..0x3F  / | `34-ending-trigger.md` |
| `byte_3F8F4` | tile = byte_3F8F4[y*32 + x]                       ; 這個單位腳下(sub_DB10 的戰場分支) | `03-picture-files.md`, `100-the-last-three-and-the-collapsed-entrance-gate.md`, `14-combat-maps.md`, `31-line-of-sight.md`, `34-ending-trigger.md`, `35-harp-and-the-secret-door.md`, `48-dungeon-wandering-monster-and-arena.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `byte_3F979` | ⬜ `sub_135FC` 沒讀:它 `byte_3F979 = 0DCh`、長出月門、把 `byte_3E06B` 個物件槽 | `86-tile-0xDC-is-an-open-moongate.md` |
| `byte_3F99F` | `sub_FE48` 在**隨機遭遇**那條路徑上寫 `byte_3F99F[槽] = 生物*4 + 0x40`, | `18-dungeons.md` |
| `byte_3FA19` | case 1: byte_3FA19 = 0xEB; break; // 刑具往前一格 | `28-shadowlords-and-blackthorn.md` |
| `byte_400F4` | 其餘       (場景)          →  byte_400F4[y*32 + x]      ← 就是這一條 | `03-scene-entry-and-tile-semantics.md`, `12-npc-movement.md`, `35-harp-and-the-secret-door.md` |
| `byte_402A5` | byte_402A5 ^= 0x0B | `35-harp-and-the-secret-door.md` |
| `byte_404F3` | 座標越界           → 用 byte_404F3 當 tile(那是地圖最後一格,原版的 quirk) | `12-npc-movement.md`, `35-harp-and-the-secret-door.md` |
| `byte_404F4` | sub_10738   memchr(byte_404F4, 0x1B, 0x400)     ← 在世界地圖的 32×32 視窗裡找 tile 0x1B | `31-line-of-sight.md`, `35-harp-and-the-secret-door.md` |
| `byte_40BA0` | 主要資料  /  ★ `off_408F4`(38 個縮寫名)、`byte_40BA0`(執行期組出的 38 筆計數)、`off_40B50`(職業名)、`off_40B8C`(狀態名)  / | `60-command-echo-and-menu-keys.md`, `94-ztats-has-seventeen-pages.md` |
| `byte_40BD4` | 資料  /  `byte_40BD4` 部位碼 ×48、`byte_40C04` 重量 ×48、`byte_40C34` 職業圖 ×9  / | `72-ready-had-seven-missing-rules.md` |
| `byte_40C04` | 資料  /  `byte_40BD4` 部位碼 ×48、`byte_40C04` 重量 ×48、`byte_40C34` 職業圖 ×9  / | `72-ready-had-seven-missing-rules.md` |
| `byte_40C34` | 資料  /  `byte_40BD4` 部位碼 ×48、`byte_40C04` 重量 ×48、`byte_40C34` 職業圖 ×9  / | `72-ready-had-seven-missing-rules.md` |
| `byte_40EAC` | 若 咒語索引 >= 0 且 byte_40EAC[索引] == 遮罩 → "\nDone!\n",份數加上去(上限 99) | `58-rune-input-cast-and-mix.md` |
| `byte_41033` | dump `byte_41033[1..32]` 得到起始索引,再用「同檔下一個地點的起始索引 − 自己」算出層數: | `03-scene-entry-and-tile-semantics.md` |
| `byte_410F3` | mov     cl, byte_410F3[edx]        ; 世界座標**從地點表讀回來**(1-based 索引的同一張表) | `03-scene-entry-and-tile-semantics.md`, `78-the-bottom-of-a-dungeon-opens-into-the-underworld.md` |
| `byte_410F4` | esi = 0; while (esi < 20h && !(byte_410F4[esi] == 隊伍X && byte_4111C[esi] == 隊伍Y)) esi++ | `03-scene-entry-and-tile-semantics.md`, `18-dungeons.md`, `87-bgm-selection-and-the-tbl-tables.md` |
| `byte_41114` | `off_55DF8`(字)、`byte_41114`/`byte_4113C`(地牢入口座標)、`byte_55E18`(入口地形); | `26-yell-words-of-power-shadowlords.md` |
| `byte_4111B` | `byte_410F3[地點]` / `byte_4111B[地點]` 是**以地點編號索引**的入口座標表。 | `03-scene-entry-and-tile-semantics.md`, `78-the-bottom-of-a-dungeon-opens-into-the-underworld.md` |
| `byte_4111C` | esi = 0; while (esi < 20h && !(byte_410F4[esi] == 隊伍X && byte_4111C[esi] == 隊伍Y)) esi++ | `03-scene-entry-and-tile-semantics.md`, `18-dungeons.md`, `87-bgm-selection-and-the-tbl-tables.md` |
| `byte_4113C` | `off_55DF8`(字)、`byte_41114`/`byte_4113C`(地牢入口座標)、`byte_55E18`(入口地形); | `26-yell-words-of-power-shadowlords.md` |
| `byte_41142` | byte_3E095 = byte_41142[日 * 2]      ← 特拉梅爾 Trammel | `22-moongates.md` |
| `byte_41143` | byte_3E096 = byte_41143[日 * 2]      ← 費盧卡 Felucca | `22-moongates.md` |
| `byte_411FC` | `0x19`(25)  /  **the shrine of …**(八德聖壇;名字取自 `off_411BC[9]`,狀態看 `byte_411FC[8]`/`byte_41204[8]`)  /  `sub_1DA1… | `03-scene-entry-and-tile-semantics.md`, `25-shrines.md`, `26-yell-words-of-power-shadowlords.md` |
| `byte_41204` | `0x19`(25)  /  **the shrine of …**(八德聖壇;名字取自 `off_411BC[9]`,狀態看 `byte_411FC[8]`/`byte_41204[8]`)  /  `sub_1DA1… | `03-scene-entry-and-tile-semantics.md`, `25-shrines.md`, `26-yell-words-of-power-shadowlords.md` |
| `byte_4185C` | while (byte_4185C[店種*16 + word_3EF34] != byte_3E0A3 && word_3EF34 < 0x10) | `08-shops.md`, `10-shop-prices-and-trade.md`, `93-the-tavern-meal-was-never-haggled.md` |
| `byte_418DD` | movzx eax, byte_418DD              ; ★ 地板 | `53-party-tile-and-the-arena-floor.md` |
| `byte_418DE` | ⚠ `byte_418DE`(這是兩向梯)是**畫戰場時**記下來的 —— `sub_FD54` 對腳下 | `52-one-coordinate-pair-one-tile-accessor.md` |
| `byte_418E0` | 中央的擺設查 `byte_418E0[高四位元 >> 4]`: | `48-dungeon-wandering-monster-and-arena.md` |
| `byte_418E8` | `byte_418E8`/`byte_418F0`(隊伍散開的位置)與 `byte_41950`/`byte_41960` | `48-dungeon-wandering-monster-and-arena.md`, `50-hole-up-camp-sleep-repair.md` |
| `byte_418F0` | `byte_418E8`/`byte_418F0`(隊伍散開的位置)與 `byte_41950`/`byte_41960` | `48-dungeon-wandering-monster-and-arena.md`, `50-hole-up-camp-sleep-repair.md` |
| `byte_418F8` | `byte_418F8` 與 `byte_4190A` 內容相同({5,4,6,3,5,7});另兩張是 | `48-dungeon-wandering-monster-and-arena.md` |
| `byte_418FE` | 側 3(朝北)X = byte_4190A  Y = byte_418FE | `48-dungeon-wandering-monster-and-arena.md` |
| `byte_41904` | 側 2(朝東)X = byte_41904  Y = byte_418F8 | `48-dungeon-wandering-monster-and-arena.md` |
| `byte_4190A` | `byte_418F8` 與 `byte_4190A` 內容相同({5,4,6,3,5,7});另兩張是 | `48-dungeon-wandering-monster-and-arena.md` |
| `byte_41920` | `byte_41920 = 10 − byte_41930` 逐項成立(測試用這條對稱性把表釘住)。 | `48-dungeon-wandering-monster-and-arena.md` |
| `byte_41930` | `byte_41920 = 10 − byte_41930` 逐項成立(測試用這條對稱性把表釘住)。 | `48-dungeon-wandering-monster-and-arena.md` |
| `byte_41950` | `byte_418E8`/`byte_418F0`(隊伍散開的位置)與 `byte_41950`/`byte_41960` | `48-dungeon-wandering-monster-and-arena.md`, `50-hole-up-camp-sleep-repair.md` |
| `byte_41960` | `byte_418E8`/`byte_418F0`(隊伍散開的位置)與 `byte_41950`/`byte_41960` | `48-dungeon-wandering-monster-and-arena.md`, `50-hole-up-camp-sleep-repair.md` |
| `byte_41988` | if (byte_41988) 發聲(音階[edi] + 0x3C, 0x78, 0xC350) | `35-harp-and-the-secret-door.md` |
| `byte_41989` | byte_41989 = 1                                        ; 不論成不成功都設 | `73-escape-is-a-key-and-it-has-two-gates.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `byte_4198A` | if (var_24) { 印 "Door destroyed!"; 那一格寫成 0x44; byte_4198A = 1 } | `21-chests-fields-locks.md`, `31-line-of-sight.md`, `69-the-overworld-cannon-did-nothing.md`, `73-escape-is-a-key-and-it-has-two-gates.md`, `75-open-did-nothing-outside-dungeons.md`, `85-monsters-move-and-terrain-slows-them.md`, `86-tile-0xDC-is-an-open-moongate.md` |
| `byte_41C18` | 2. 或查 `byte_41C18` 的 xref(`sub_6730` 開頭有 `for(i=0;i<256;i++) byte_41C18[i]=i`, | `01-tileset-and-dot16-loader.md` |
| `byte_4FC54` | `byte_4FC54` 的前四個位元組是 **`3, 4, 2, 1`** —— 四個方向鍵的鍵碼 | `70-hunger-poison-and-the-vanishing-rings.md` |
| `byte_4FC80` | 資料  /  `dword_478C8`(音階)、`byte_4FC80`(曲子 13 音)、`byte_4FC8D`(進度)  / | `35-harp-and-the-secret-door.md` |
| `byte_4FC8D` | 資料  /  `dword_478C8`(音階)、`byte_4FC80`(曲子 13 音)、`byte_4FC8D`(進度)  / | `35-harp-and-the-secret-door.md` |
| `byte_4FD1C` | 陸地・地表**(`byte_4FD1C` + `dword_4FD38`)—— 順序本身就是難度曲線: | `82-the-overworld-world-turn.md` |
| `byte_4FD28` | 陸地・幽冥界**(`byte_4FD28` + `a88`):Bat 25% / Giant Rat 21.9% / Giant Spider 21.9% / | `82-the-overworld-world-turn.md` |
| `byte_4FD30` | 水域・地表**(`byte_4FD30` + `dword_4FD4C`):Shark 28.1% / Squid 28.1% / | `82-the-overworld-world-turn.md` |
| `byte_4FD35` | 水域・幽冥界**(`byte_4FD35` + `dword_4FD50+1` = `00 80 80`): | `82-the-overworld-world-turn.md` |
| `byte_4FD94` | return !(x == byte_4FD94 && y == byte_4FD95) | `85-monsters-move-and-terrain-slows-them.md` |
| `byte_4FD95` | return !(x == byte_4FD94 && y == byte_4FD95) | `85-monsters-move-and-terrain-slows-them.md` |
| `byte_4FD98` | `byte_4FD98`,索引 `11* / dy /  +  / dx / `,` / dx / ` / ` / dy / ` 各需 < 6。 | `88-the-load-window-culling-and-ship-cannon.md` |
| `byte_4FDBF` | 就不猜。表實體跨過 IDA 另外命名的 `byte_4FDBF` 與 `byte_4FDC7` | `88-the-load-window-culling-and-ship-cannon.md` |
| `byte_4FDC7` | 就不猜。表實體跨過 IDA 另外命名的 `byte_4FDBF` 與 `byte_4FDC7` | `88-the-load-window-culling-and-ship-cannon.md` |
| `byte_4FDD4` | (索引 0..60 恰好用到 `byte_4FDD4`)—— 又一次「相鄰表物理重疊」。 | `88-the-load-window-culling-and-ship-cannon.md` |
| `byte_4FDD5` | if (byte_3E08A == 'Q') { byte_4FDD5 ^= 1; if (set) return 0 }   ; Rel Tym 隔次 | `38-terrain-movement-cost.md`, `82-the-overworld-world-turn.md` |
| `byte_4FDD7` | if ((byte_3E08C & 0FEh) == 12h  /  /  == 14h) { byte_4FDD7 ^= 1; if (set) return 0 }  ; 馬/魔毯隔次 | `38-terrain-movement-cost.md`, `82-the-overworld-world-turn.md` |
| `byte_4FEEE` | 正面(`sub_36C0`,查 `byte_4FEEE[種類]`): | `18-dungeons.md` |
| `byte_4FF90` | `byte_4FF90` / `byte_4FF98` 兩張表選圖,畫的是 **`ITEMS.16`** 而不是 `DNG*.16`, | `18-dungeons.md` |
| `byte_4FF98` | `byte_4FF90` / `byte_4FF98` 兩張表選圖,畫的是 **`ITEMS.16`** 而不是 `DNG*.16`, | `18-dungeons.md` |
| `byte_4FFA0` | 噴泉的動畫(`sub_39A8` + `byte_4FFA0` 三格循環)與力場動畫(`sub_33F0`)。 | `18-dungeons.md` |
| `byte_4FFA4` | 側向來自 `byte_4FFA4` / `byte_4FFA8`(= 朝向轉 90 度); | `18-dungeons.md` |
| `byte_4FFA8` | 側向來自 `byte_4FFA4` / `byte_4FFA8`(= 朝向轉 90 度); | `18-dungeons.md` |
| `byte_54208` | (`byte_54208` / `byte_54220` / `byte_54238` / `byte_54250`)都是 FM Towns | `24-intro.md` |
| `byte_54220` | (`byte_54208` / `byte_54220` / `byte_54238` / `byte_54250`)都是 FM Towns | `24-intro.md` |
| `byte_54238` | (`byte_54208` / `byte_54220` / `byte_54238` / `byte_54250`)都是 FM Towns | `24-intro.md` |
| `byte_54250` | (`byte_54208` / `byte_54220` / `byte_54238` / `byte_54250`)都是 FM Towns | `24-intro.md` |
| `byte_54268` | 所以被讀的是 `0x4FFB8 + 0x42B0 = 0x54268` —— 反組譯檔裡 grep `byte_54268` | `24-intro.md` |
| `byte_54280` | `byte_54280[頁]`  /  用哪一個 `STORY*.16`(0..5)  / | `24-intro.md` |
| `byte_54298` | `byte_54298` / `byte_542B0`  /  擺放座標(FM Towns 640×480)  / | `24-intro.md` |
| `byte_542B0` | `byte_54298` / `byte_542B0`  /  擺放座標(FM Towns 640×480)  / | `24-intro.md` |
| `byte_542C8` | `byte_542C8[頁]`  /  頁的種類  / | `24-intro.md` |
| `byte_54524` | ⚠⚠ **`byte_54524` 不是玩家那張通行表。** 玩家走 `byte_5FF6C`,NPC 走這張, | `12-npc-movement.md`, `20-projectiles.md` |
| `byte_54700` | 2. 從 `sub_2C740` 與 `byte_54700` 的 xref 反追 `.TLK` 索引表語意與控制碼(`\x01` 疑為玩家名代入)。 | `00-hexrays-p3-verified.md`, `27-codex-and-the-shrine-chamber.md`, `28-shadowlords-and-blackthorn.md`, `30-ending.md`, `67-corpser-and-the-sleeping-party-member.md`, `79-four-dialogue-opcodes-were-corrupting-the-text.md` |
| `byte_54721` | `byte_54721`  /  0x021  /  1  /  「汝可帶來了吾的盒子?」  / | `30-ending.md` |
| `byte_54749` | `byte_54749`  /  0x049  /  2  /  「那只檀香木盒……汝可帶來了?」  / | `30-ending.md` |
| `byte_547AB` | `byte_547AB`  /  0x0AB  /  3  /  「不列顛王小心翼翼地打開盒子……」  / | `30-ending.md` |
| `byte_547D3` | `byte_547D3`  /  0x0D3  /  4  /  星界器物  / | `30-ending.md` |
| `byte_54867` | `byte_54867`  /  0x167  /  6  /  「它將帶我們脫出此獄!」  / | `30-ending.md` |
| `byte_548C9` | `byte_548C9`  /  0x1C9  /  7  /  「比蒙丹之惡更為久遠」  / | `30-ending.md` |
| `byte_54911` | `byte_54911`  /  0x211  /  8  /  「但月之球的力量,比那更為久遠!」  / | `30-ending.md` |
| `byte_5494B` | `byte_5494B`  /  0x24B  /  9  /  「隨吾來!」  / | `30-ending.md` |
| `byte_549D5` | `byte_549D5`  /  0x2D5  /  10  /  「那麼,搬張椅子坐下吧。」  / | `30-ending.md` |
| `byte_54A6D` | `byte_54A6D`  /  28  /  「……汝跪倒在聖壇之前。」  / | `27-codex-and-the-shrine-chamber.md` |
| `byte_54AE1` | `byte_54AE1`  /  31  /  「聖壇開口,一項試煉就此降下!」  / | `27-codex-and-the-shrine-chamber.md` |
| `byte_54BB9` | `byte_54BB9`  /  36  /  「做得好!」  / | `27-codex-and-the-shrine-chamber.md` |
| `byte_54BE5` | `byte_54BE5`  /  37  /  「書已翻至汝所尋的那一頁!」  / | `27-codex-and-the-shrine-chamber.md` |
| `byte_54C15` | `byte_54C15`  /  38  /  「汝在那神聖的一頁上讀到:」  / | `27-codex-and-the-shrine-chamber.md` |
| `byte_54C55` | `byte_54C55`  /  40  /  「一陣異風翻動了書頁!」  / | `27-codex-and-the-shrine-chamber.md` |
| `byte_54C7F` | `byte_54C7F`…`dword_54D80`  /  41–44  /  四段符文  / | `27-codex-and-the-shrine-chamber.md` |
| `byte_54DDB` | `byte_54DDB`  /  46  /  「終極智慧之寶典就在汝眼前……」  / | `27-codex-and-the-shrine-chamber.md` |
| `byte_55110` | `byte_55110` / `byte_55116`  /  `u5data.DungeonRoomAlwaysArmed()`  / | `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `byte_55116` | `byte_55110` / `byte_55116`  /  `u5data.DungeonRoomAlwaysArmed()`  / | `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `byte_55140` | 一律崩塌。引擎只改八個入口本身 —— 要做整區塊得先 dump `byte_55140`。 | `33-get-command.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `byte_55384` | ⚠ Hex-Rays 把 `word_3EF34` 整個折掉了:`byte_55384[edi + eax*8]` 被印成 | `10-shop-prices-and-trade.md` |
| `byte_55DD4` | 資料  /  `byte_55DD4` / `byte_55DDC` / `byte_55DE4`(地牢獎品三張並列表)  / | `33-get-command.md` |
| `byte_55DDC` | 資料  /  `byte_55DD4` / `byte_55DDC` / `byte_55DE4`(地牢獎品三張並列表)  / | `33-get-command.md` |
| `byte_55DE4` | 資料  /  `byte_55DD4` / `byte_55DDC` / `byte_55DE4`(地牢獎品三張並列表)  / | `33-get-command.md` |
| `byte_55E18` | 3  /  `byte_55E18` = `18 16 16 18 18 17 17 16`(八座入口的**原始**地形),正是 `sub_105E4` 會改寫成 `0xDF` 的那三個值 ⇒ 兩支函式講同一件事  / | `26-yell-words-of-power-shadowlords.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `byte_55E20` | 戰鬥中的力場**:`sub_20360(單位, byte_55E20[種類])`,效果碼 0x33..0x36, | `17-magic.md` |
| `byte_55E24` | 格子編號取自 `byte_55E24`(FM Towns 線性位址 0x55E24,檔案位移 +0x200): | `17-magic.md`, `18-dungeons.md` |
| `byte_55E50` | dword_3E3DC  / = byte_55E50[i];                   // 2 / 4 / 8 | `28-shadowlords-and-blackthorn.md` |
| `byte_55F18` | 0x85 / 0x86 / 0x8C / 0xFE  /  切到子模式  /  `byte_55F18 = 碼`;函式開頭會把後續位元組改送 `sub_1C1E8`  / | `06-conversation-script.md`, `79-four-dialogue-opcodes-were-corrupting-the-text.md` |
| `byte_55F19` | if (byte_55F19 == 3) call sub_1B854 | `79-four-dialogue-opcodes-were-corrupting-the-text.md` |
| `byte_55F1A` | 0x8E  /  切換強調  /  `byte_55F1A ^= 0x80`,影響後續字元的輸出屬性  / | `06-conversation-script.md`, `79-four-dialogue-opcodes-were-corrupting-the-text.md` |
| `byte_55F1B` | (byte_55F1B & 7Fh) × 100 − 12C0h        ; 12C0h = '0' × 100 | `79-four-dialogue-opcodes-were-corrupting-the-text.md` |
| `byte_55F1C` | (byte_55F1C & 7Fh) × 10  − 1E0h         ; 1E0h  = '0' × 10 | `79-four-dialogue-opcodes-were-corrupting-the-text.md` |
| `byte_55F1D` | (byte_55F1D & 7Fh)       − 30h | `79-four-dialogue-opcodes-were-corrupting-the-text.md` |
| `byte_55F32` | 0x91–0x9F  /  反問玩家並讀取回答(15 種)  /  `byte_55F32 = 碼; sub_1C0AC` → 印 "You respond-\n:" 後收輸入  / | `06-conversation-script.md`, `79-four-dialogue-opcodes-were-corrupting-the-text.md` |
| `byte_55F37` | (位置 0,或前一個字元是空白)——`sub_1BD8C` → `sub_27C98`,再檢查 `byte_55F37[i]`。 | `06-conversation-script.md`, `61-tavern-lore-and-the-transfer-second-stage.md` |
| `byte_55F38` | if (輸入是空的) → 「汝說是就是吧。」            // cmp al, byte_55F38,al 此刻是 0 | `06-conversation-script.md` |
| `byte_55F4A` | else byte_55F4A = 1                            ; 設下 pendingSpace | `05-text-compression.md` |
| `byte_55FE9` | byte_55FE9 = byte_3E0A3;               // 記住原本的地點 | `27-codex-and-the-shrine-chamber.md` |
| `byte_5606C` | `byte_5606C` / `byte_56074` / `byte_5607C`  /  完成試煉加力量 / 敏捷 / 智力  / | `25-shrines.md` |
| `byte_56074` | `byte_5606C` / `byte_56074` / `byte_5607C`  /  完成試煉加力量 / 敏捷 / 智力  / | `25-shrines.md` |
| `byte_5607C` | `byte_5606C` / `byte_56074` / `byte_5607C`  /  完成試煉加力量 / 敏捷 / 智力  / | `25-shrines.md` |
| `byte_56DE8` | `mov eax, offset byte_56DE8` + `mov dl, [eax+edx+24Ch]` ⇒ `0x56DE8 + 0x24C`。 | `10-shop-prices-and-trade.md` |
| `byte_57034` | `byte_57034[酒館]` 選出一套「菜單樣式」(0..3),樣式再決定四個字母。 | `10-shop-prices-and-trade.md` |
| `byte_57080` | mov   dl, byte_57080[eax]     ; 停泊 X | `11-map-objects.md` |
| `byte_57084` | mov   al, byte_57084[eax]     ; 停泊 Y | `11-map-objects.md` |
| `byte_57090` | `byte_57090[旅店]`(每人每天 2 或 3 金)上: | `10-shop-prices-and-trade.md` |
| `byte_57098` | 客滿判斷在 `sub_21CE4`:寄放的人數不能超過 `byte_57098[旅店]`(3/4/3/2/2/2 間房)。 | `10-shop-prices-and-trade.md` |
| `byte_5717C` | `byte_5717C[v]`  /  本輪抽過了  /  **每輪之間清空**  / | `39-character-creation.md` |
| `byte_57184` | `byte_57184[v]`  /  已被淘汰  /  一路留到最後  / | `39-character-creation.md` |
| `byte_5718E` | byte_5718E = *((_BYTE *)dword_57164 + v4); | `39-character-creation.md` |
| `byte_5718F` | mov     al, byte_5718F | `39-character-creation.md` |
| `byte_57190` | `cmp byte_57190, 14h`  /  `u5data.CreateMinStrength`  / | `39-character-creation.md` |
| `byte_5FF5C` | 資料  /  `byte_601F0`(距離表)、`dword_601D0`(擋視線地形)、`dword_601E4`(發光地形)、`byte_5FF5C`(日出漸變)  / | `31-line-of-sight.md` |
| `byte_5FF6C` | BOOL ok = ((128 >> (tile & 7)) & byte_5FF6C[tile >> 3]) == 0;   // bit = 1 → 阻擋 | `02-movement-and-tile-flags.md`, `12-npc-movement.md`, `20-projectiles.md`, `47-move-modes-and-time-of-day.md` |
| `byte_5FF8C` | switch (byte_5FF8C[mover >> 2]) {          // 移動者 → 移動模式(0–10) | `02-movement-and-tile-flags.md`, `47-move-modes-and-time-of-day.md` |
| `byte_5FFA8` | case 5: /* 方向性通行:用 byte_5FFA8[tile] / byte_5FF6C[tile] 的方向 bit */ | `02-movement-and-tile-flags.md` |
| `byte_5FFEC` | 主要資料  /  **`byte_5FFEC` = 八筆權重表**、`byte_3E0A3`(地點碼)、`byte_3E06B`(隊伍人數)  / | `91-chest-traps-four-kinds-and-their-weights.md` |
| `byte_5FFFC` | 主要資料  /  `byte_60038`(八相)、`byte_3E094`(剩餘鐘響)、`byte_5FFFC`(旋律游標)、**0x60060 音高表**、**0x6006C 53 步旋律**  / | `92-white-noise-shim-and-ambient-sound.md` |
| `byte_5FFFD` | `byte_5FFFD` 壓制  /  `State.musicSuppressed` / `suppressedSong`  / | `92-white-noise-shim-and-ambient-sound.md` |
| `byte_60018` | 1. 擋箭的表:`byte_60018`(DOS `0x6A24`) | `20-projectiles.md` |
| `byte_60038` | 主要資料  /  `byte_60038`(八相)、`byte_3E094`(剩餘鐘響)、`byte_5FFFC`(旋律游標)、**0x60060 音高表**、**0x6006C 53 步旋律**  / | `92-white-noise-shim-and-ambient-sound.md` |
| `byte_600A1` | `byte_600A1/2`  /  `State.waterfallPlaying` / `fountainPlaying`  / | `92-white-noise-shim-and-ambient-sound.md` |
| `byte_600A2` | `byte_600A1` / `byte_600A2` 兩個旗標,而清除的條件是 | `92-white-noise-shim-and-ambient-sound.md` |
| `byte_601C7` | 主要函式  /  `sub_25F0`(物件的攻擊)、`sub_2D38`(物件能不能動)、`sub_2D2D0`(**玩家**的自動航行)、`dword_4FD50`(敵船的風向延遲表)、`byte_601C7`/`b… | `83-monster-attacks-and-two-different-wind-rules.md`, `84-the-sailing-gate-and-the-wind-table.md` |
| `byte_601CB` | 主要函式  /  `sub_25F0`(物件的攻擊)、`sub_2D38`(物件能不能動)、`sub_2D2D0`(**玩家**的自動航行)、`dword_4FD50`(敵船的風向延遲表)、`byte_601C7`/`b… | `83-monster-attacks-and-two-different-wind-rules.md`, `84-the-sailing-gate-and-the-wind-table.md` |
| `byte_601F0` | 資料  /  `byte_601F0`(距離表)、`dword_601D0`(擋視線地形)、`dword_601E4`(發光地形)、`byte_5FF5C`(日出漸變)  / | `31-line-of-sight.md` |
| `byte_60948` | ebx = &byte_60948[n*3]                                ; ★ 起點 MSF(分/秒/格) | `90-cdda-scenes-and-sound-effect-indices.md` |
| `byte_6095C` | sub_3370A(0, ebx, &byte_6095C[n*3])                   ; ★ 從起點播到終點 | `90-cdda-scenes-and-sound-effect-indices.md` |
| `byte_60971` | sub_337DE(0, &err, &var_4, byte_60971, word_6096E)   ; 查碟(cdr_cdinfo 那一族) | `90-cdda-scenes-and-sound-effect-indices.md` |
| `byte_726DC` | mov  word  ptr byte_726DC+4, ds     ; ← 檔名的 segment(組成 far pointer) | `89-music-playback-is-not-in-the-game-binary.md` |
| `byte_738D8` | if (byte_738D8[al] & 2)          ; ctype 表:是小寫嗎 | `65-the-wishing-well-easter-egg.md` |
| `word_3DDC4` | 而且做兩件事:狀態 `'D'` → `'G'`,**以及** `word_3DDC4 = word_3DDC6`(目前 HP = 最大 HP)。 | `19-levelup.md`, `30-ending.md`, `37-look-signs-and-the-sky.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `word_3DDC6` | 而且做兩件事:狀態 `'D'` → `'G'`,**以及** `word_3DDC4 = word_3DDC6`(目前 HP = 最大 HP)。 | `19-levelup.md`, `30-ending.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `word_3DDC8` | 經驗值 = U4 經驗值 / 10                              (word_3DDC8 /= 10) | `19-levelup.md`, `61-tavern-lore-and-the-transfer-second-stage.md` |
| `word_3DFB4` | if (word_3DFB4 == 0) {                                      ; ★ 存糧 0 | `10-shop-prices-and-trade.md`, `13-save-writing.md`, `33-get-command.md`, `70-hunger-poison-and-the-vanishing-rings.md`, `94-ztats-has-seventeen-pages.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `word_3DFB6` | `word_3DFB6` 是**金錢**(`docs/re/11` 的買馬那段:`sub word_3DFB6, ax ; 扣錢`)。 | `10-shop-prices-and-trade.md`, `11-map-objects.md`, `13-save-writing.md`, `33-get-command.md`, `65-the-wishing-well-easter-egg.md`, `68-troll-bridge-and-collision.md`, `94-ztats-has-seventeen-pages.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `word_3E084` | word_3E084 年 | `04-npc-schedule-and-clock.md` |
| `word_3E086` | 主要資料  /  `dword_3E16C`(地牢層緩衝)、`byte_3EE15`(朝向)、`byte_3EE16`、★ `byte_3E08B`(當前選定的隊員)、`word_3E086`/`word_3E088`(… | `100-the-last-three-and-the-collapsed-entrance-gate.md`, `16-combat-turns-and-ai.md`, `37-look-signs-and-the-sky.md`, `40-push-and-the-main-menu.md`, `49-command-table-and-two-empty-keys.md`, `68-troll-bridge-and-collision.md`, `69-the-overworld-cannon-did-nothing.md`, `72-ready-had-seven-missing-rules.md`, `82-the-overworld-world-turn.md`, `95-the-dungeon-has-its-own-look.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `word_3E088` | 主要資料  /  `dword_3E16C`(地牢層緩衝)、`byte_3EE15`(朝向)、`byte_3EE16`、★ `byte_3E08B`(當前選定的隊員)、`word_3E086`/`word_3E088`(… | `100-the-last-three-and-the-collapsed-entrance-gate.md`, `16-combat-turns-and-ai.md`, `37-look-signs-and-the-sky.md`, `40-push-and-the-main-menu.md`, `49-command-table-and-two-empty-keys.md`, `82-the-overworld-world-turn.md`, `95-the-dungeon-has-its-own-look.md` |
| `word_3E770` | rec = &word_3E770[npc*16]                      // 執行期 NPC 記錄 | `12-npc-movement.md`, `36-sandalwood-box-npc-objects.md` |
| `word_3E77A` | for i in 0..31: word_3E77A[i*16] = 區域緩衝[i]   ; 對話號碼搬進執行期記錄 | `04-npc-schedule-and-clock.md` |
| `word_3E77C` | sub_2E58C(word_3E77C[npc*16]);                 // 開打 —— 與撞上野外怪物同一支 | `29-npc-behaviour-and-arrest.md`, `36-sandalwood-box-npc-objects.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `word_3ED70` | `word_3ED70[npc]` —— 路徑索引;`0xFFFF` = 目前沒有路徑 | `12-npc-movement.md` |
| `word_3EDD4` | if word_3EDD4[npc] != 0 且 random(0,2) != 1 → 這回合不重算 | `12-npc-movement.md` |
| `word_3EF34` | 主要資料  /  `word_3EF38`(總價)、`dword_56E14`/`dword_56E10`(人數)、`dword_56E1C`(★ 喝了幾杯)、`dword_56E18`(上了菜的 tile)、`word… | `10-shop-prices-and-trade.md`, `11-map-objects.md`, `93-the-tavern-meal-was-never-haggled.md` |
| `word_3EF36` | mov  word_3EF36, ax          ; 店種 = 對話碼 − 0x81 | `10-shop-prices-and-trade.md`, `93-the-tavern-meal-was-never-haggled.md` |
| `word_3EF38` | 主要資料  /  `word_3EF38`(總價)、`dword_56E14`/`dword_56E10`(人數)、`dword_56E1C`(★ 喝了幾杯)、`dword_56E18`(上了菜的 tile)、`word… | `10-shop-prices-and-trade.md`, `11-map-objects.md`, `93-the-tavern-meal-was-never-haggled.md` |
| `word_3EF3A` | `^` 0x5E  /  `word_3EF3A`  /  數量(藥草一次幾份)  / | `10-shop-prices-and-trade.md` |
| `word_3EF3C` | (`word_3EF44` / `word_3EF42` / `word_3EF3C` / `off_3EF3E+2`),看起來是寬度。 | `20-projectiles.md` |
| `word_3EF42` | (`word_3EF44` / `word_3EF42` / `word_3EF3C` / `off_3EF3E+2`),看起來是寬度。 | `20-projectiles.md` |
| `word_3EF44` | (`word_3EF44` / `word_3EF42` / `word_3EF3C` / `off_3EF3E+2`),看起來是寬度。 | `20-projectiles.md` |
| `word_3F1D0` | sub_18EB0 (An Xen Corp)  額外要求 word_3F1D0[生物] & 0x20 | `17-magic.md`, `48-dungeon-wandering-monster-and-arena.md` |
| `word_4140C` | 找到 → word_4140C = x, word_4140E = y | `31-line-of-sight.md` |
| `word_4140E` | 找到 → word_4140C = x, word_4140E = y | `31-line-of-sight.md` |
| `word_41970` | 前向來自 `word_41970` / `word_41978`(北 / 東 / 南 / 西 = `(0,−1) (1,0) (0,1) (−1,0)`)。 | `18-dungeons.md`, `21-chests-fields-locks.md` |
| `word_41978` | 前向來自 `word_41970` / `word_41978`(北 / 東 / 南 / 西 = `(0,−1) (1,0) (0,1) (−1,0)`)。 | `18-dungeons.md`, `21-chests-fields-locks.md` |
| `word_54C3E` | `word_54C3E`  /  39  /  「汝是怎麼來到這裡的?」  / | `27-codex-and-the-shrine-chamber.md` |
| `word_54DAE` | `word_54DAE`  /  45  /  「汝走近那座靜謐的聖壇……」  / | `27-codex-and-the-shrine-chamber.md` |
| `word_6096E` | sub_337DE(0, &err, &var_4, byte_60971, word_6096E)   ; 查碟(cdr_cdinfo 那一族) | `90-cdda-scenes-and-sound-effect-indices.md` |
| `dword_3E16C` | 主要資料  /  `dword_3E16C`(地牢層緩衝)、`byte_3EE15`(朝向)、`byte_3EE16`、★ `byte_3E08B`(當前選定的隊員)、`word_3E086`/`word_3E088`(… | `18-dungeons.md`, `29-npc-behaviour-and-arrest.md`, `75-open-did-nothing-outside-dungeons.md`, `95-the-dungeon-has-its-own-look.md` |
| `dword_3E368` | `dword_3E368[地點]` 是 32 個 u32,位元 i = 那個地點的第 i 個 NPC 已被永久清掉。 | `29-npc-behaviour-and-arrest.md`, `36-sandalwood-box-npc-objects.md` |
| `dword_3E36C` | 它**在存檔裡**(`sub_27D24` 讀 `dword_3E36C`,0x80 = 128 B),而引擎原本只有一個 | `29-npc-behaviour-and-arrest.md` |
| `dword_3E3DC` | dword_3E3DC  / = byte_55E50[i];                   // 2 / 4 / 8 | `28-shadowlords-and-blackthorn.md`, `29-npc-behaviour-and-arrest.md` |
| `dword_3E3E8` | dword_3E3E8[地點]  / = 1 << i;                // sub_1C1AC:這座城的這個人認得汝了 | `06-conversation-script.md`, `79-four-dialogue-opcodes-were-corrupting-the-text.md` |
| `dword_3E46C` | ≥ 0x40  /  怪物  /  `cmp byte ptr dword_3E46C[eax*8], 40h`,與生物名表的 `CreatureBase` 同源  / | `10-shop-prices-and-trade.md`, `11-map-objects.md`, `33-get-command.md`, `36-sandalwood-box-npc-objects.md`, `48-dungeon-wandering-monster-and-arena.md`, `53-party-tile-and-the-arena-floor.md`, `65-the-wishing-well-easter-egg.md`, `67-corpser-and-the-sleeping-party-member.md`, `84-the-sailing-gate-and-the-wind-table.md`, `88-the-load-window-culling-and-ship-cannon.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `dword_3E470` | dword_3E470+2[slot*8] = 0 | `36-sandalwood-box-npc-objects.md` |
| `dword_3E474` | `dword_3E474` = `dword_3E46C + 8` = 物件陣列的第 1 槽;`dword_3E47C` 是第 2 槽。 | `48-dungeon-wandering-monster-and-arena.md` |
| `dword_3E47C` | `dword_3E474` = `dword_3E46C + 8` = 物件陣列的第 1 槽;`dword_3E47C` 是第 2 槽。 | `48-dungeon-wandering-monster-and-arena.md` |
| `dword_3E54C` | 槽號也不是猜的:`dword_3E54C` 與 `3E554h[i*8]` 相對於物件表起點 `dword_3E46C` | `33-get-command.md` |
| `dword_3EF24` | dword_3EF24 = off_4145C[…]                 ; 店名 "Iolo's Bows" | `10-shop-prices-and-trade.md`, `93-the-tavern-meal-was-never-haggled.md` |
| `dword_3EF28` | dword_3EF28 = off_4165C[…]                 ; 店主 "Gwenneth" | `10-shop-prices-and-trade.md`, `93-the-tavern-meal-was-never-haggled.md` |
| `dword_3EF2C` | `&` 0x26  /  `dword_3EF2C`  /  物品名(收購對白;有別稱就用別稱)  / | `10-shop-prices-and-trade.md` |
| `dword_3EF30` | `*` 0x2A  /  `dword_3EF30`  /  地名(酒館八卦)  / | `10-shop-prices-and-trade.md` |
| `dword_3EF4C` | 在一塊 32×32 的 buffer(`dword_3EF4C` 指向)裡標出每一格的狀態: | `12-npc-movement.md` |
| `dword_3EF50` | if (地點碼 > 0x80)            esi = dword_3EF50[byte_3E0AE*8 + 3]  ; 戰鬥:當前單位 | `16-combat-turns-and-ai.md`, `17-magic.md`, `34-ending-trigger.md`, `52-one-coordinate-pair-one-tile-accessor.md`, `53-party-tile-and-the-arena-floor.md`, `67-corpser-and-the-sleeping-party-member.md`, `71-the-use-list-had-29-empty-slots.md`, `95-the-dungeon-has-its-own-look.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `dword_3EF54` | `dword_3EF50[空槽] = dword_3EF50[目標]`,連 `dword_3EF54` 一起搬 —— | `17-magic.md`, `52-one-coordinate-pair-one-tile-accessor.md`, `67-corpser-and-the-sleeping-party-member.md` |
| `dword_40C50` | `dword_40C50[字母*4]` 是以 ASCII 碼直接索引的指標表,J / O 那兩格沒有字串 —— | `58-rune-input-cast-and-mix.md` |
| `dword_41794` | 它對 `b >= 0x80` 查 `dword_41794[b*4]`,而 `0x41794 + 0x80*4 = 0x41994`, | `05-text-compression.md`, `08-shops.md` |
| `dword_41990` | 來源:FM Towns `WORRIORS.EXP` 的 `sub_1C3F8`(展開)與 `dword_41990`(槽表); | `05-text-compression.md`, `08-shops.md` |
| `dword_41D28` | │    └─ 啟動初始化:載 towns.fnt(17,280 B → dword_41D28)與 u5.fnt(0x4000 B → dword_4FFB8); | `01-tileset-and-dot16-loader.md` |
| `dword_478C8` | 資料  /  `dword_478C8`(音階)、`byte_4FC80`(曲子 13 音)、`byte_4FC8D`(進度)  / | `35-harp-and-the-secret-door.md` |
| `dword_4FD38` | 陸地・地表**(`byte_4FD1C` + `dword_4FD38`)—— 順序本身就是難度曲線: | `82-the-overworld-world-turn.md` |
| `dword_4FD4C` | 水域・地表**(`byte_4FD30` + `dword_4FD4C`):Shark 28.1% / Squid 28.1% / | `82-the-overworld-world-turn.md` |
| `dword_4FD50` | 主要函式  /  `sub_25F0`(物件的攻擊)、`sub_2D38`(物件能不能動)、`sub_2D2D0`(**玩家**的自動航行)、`dword_4FD50`(敵船的風向延遲表)、`byte_601C7`/`b… | `23-wind-and-sailing.md`, `82-the-overworld-world-turn.md`, `83-monster-attacks-and-two-different-wind-rules.md`, `84-the-sailing-gate-and-the-wind-table.md`, `85-monsters-move-and-terrain-slows-them.md` |
| `dword_4FDD8` | `sub_3124`  /  寫 `dword_4FDD8/DC` 並包一對 `sub_27754`  /  疑為游標 / 中斷遮罩,不碰遊戲狀態  / | `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `dword_4FFB8` | `u5.fnt`  /  **0x4000 = 16,384 B**  /  `dword_4FFB8`  /  `ULTIMA FONT DATA READ FAIL !!`  / | `01-tileset-and-dot16-loader.md`, `24-intro.md` |
| `dword_541B4` | `dword_541B4[頁]`  /  這一頁的文字在 `STORY.DAT` 裡的**檔案位移**  / | `24-intro.md` |
| `dword_54300` | push edi                             ; → dword_54300 | `55-transfer-from-ultima-iv.md` |
| `dword_54328` | 分辨的依據只有一行:`mov eax, [ebp+var_4]`,而 `var_4 = offset dword_54328`。 | `55-transfer-from-ultima-iv.md` |
| `dword_54498` | `dword_54498`  /  `game.State.TransferredAvatar`(Ztats 末行,`docs/re/54` §3)  / | `54-pass-vehicle-verbs-and-the-avatar-line.md`, `55-transfer-from-ultima-iv.md` |
| `dword_54828` | `dword_54828`  /  0x128  /  5  /  「來自汝與吾同稱為故鄉的那個世界」  / | `30-ending.md` |
| `dword_54A98` | `dword_54A98`  /  29  /  「汝欲冥想何種美德?」  / | `27-codex-and-the-shrine-chamber.md` |
| `dword_54D80` | `byte_54C7F`…`dword_54D80`  /  41–44  /  四段符文  / | `27-codex-and-the-shrine-chamber.md` |
| `dword_54E2C` | 0x29–0x31  /  巨集:一個位元組展開成 16 個字的整列(`dword_54E2C[c*4]`)  / | `37-look-signs-and-the-sky.md` |
| `dword_552C4` | mov  ax, word ptr dword_552C4[edi*4]  ; base | `10-shop-prices-and-trade.md` |
| `dword_553CC` | `dword_553CC[店種][4]` 是四句問候語在 `SHOPPE.DAT` 裡的**位元組位移**, | `08-shops.md` |
| `dword_555E8` | (dword_555E8 = {0, 0, 1, -1}、dword_555F8 = {1, -1, 0, 0}) | `11-map-objects.md` |
| `dword_555F8` | (dword_555E8 = {0, 0, 1, -1}、dword_555F8 = {1, -1, 0, 0}) | `11-map-objects.md` |
| `dword_55714` | 唯一的例外是 `dword_55714`(裝備的「另一種說法」,`Cloth suit`、`Two-Handed Axe` | `10-shop-prices-and-trade.md` |
| `dword_55A48` | ⬜ `sub_14B2C` 的 `dword_55A48`(進來設 1、失敗設 0)語意未追。 | `76-jimmy-in-a-dungeon-disarms-not-unlocks.md`, `96-search-had-three-more-layers.md` |
| `dword_55E6C` | push dword_55E6C             ; ★ 參數是「當前 NPC 的槽號」,不是腳本位元組 | `79-four-dialogue-opcodes-were-corrupting-the-text.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `dword_55F14` | 原版 0x87 的作法是把文字指標存起來、往下讀一則再還原(`dword_55F14` 的存取還原)。 | `06-conversation-script.md`, `79-four-dialogue-opcodes-were-corrupting-the-text.md` |
| `dword_5602C` | `dword_5602C`  /  0, 27, 57, 95, 123, 159, 185, 214  /  第 **12..19** 筆(「去寶典學 X」)  / | `25-shrines.md`, `27-codex-and-the-shrine-chamber.md` |
| `dword_5604C` | `dword_5604C`  /  256, 334, 393, 468, 555, 639, 704, 784  /  第 **20..27** 筆(八段箴言)  / | `27-codex-and-the-shrine-chamber.md` |
| `dword_56E10` | 主要資料  /  `word_3EF38`(總價)、`dword_56E14`/`dword_56E10`(人數)、`dword_56E1C`(★ 喝了幾杯)、`dword_56E18`(上了菜的 tile)、`word… | `93-the-tavern-meal-was-never-haggled.md` |
| `dword_56E14` | 主要資料  /  `word_3EF38`(總價)、`dword_56E14`/`dword_56E10`(人數)、`dword_56E1C`(★ 喝了幾杯)、`dword_56E18`(上了菜的 tile)、`word… | `93-the-tavern-meal-was-never-haggled.md` |
| `dword_56E18` | 主要資料  /  `word_3EF38`(總價)、`dword_56E14`/`dword_56E10`(人數)、`dword_56E1C`(★ 喝了幾杯)、`dword_56E18`(上了菜的 tile)、`word… | `93-the-tavern-meal-was-never-haggled.md` |
| `dword_56E1C` | 主要資料  /  `word_3EF38`(總價)、`dword_56E14`/`dword_56E10`(人數)、`dword_56E1C`(★ 喝了幾杯)、`dword_56E18`(上了菜的 tile)、`word… | `70-hunger-poison-and-the-vanishing-rings.md`, `93-the-tavern-meal-was-never-haggled.md` |
| `dword_56E20` | 單價來自 `sub_210D8`:`sub_20E6C(dword_56E20[word_3EF34])`。 | `93-the-tavern-meal-was-never-haggled.md` |
| `dword_57164` | `dword_57164` 前 24 B  /  `u5data.VirtueBonus`  / | `39-character-creation.md` |
| `dword_57194` | `dword_57194[8*a + b]` 是「a 對上 b」那一題在 `QUESTION.DAT` 的**檔案位移**, | `39-character-creation.md` |
| `dword_5AC30` | 1. 查 `dword_5AC30`(handle 表)的 xref → 找取用 handle 的函式 → 那裡才有解壓。 | `01-tileset-and-dot16-loader.md` |
| `dword_5FF34` | ⬜ `sub_3FE4` 開頭的 `dword_5FF34 = 0` 與 `dword_601C0 == 1 → sub_2C250` | `78-the-bottom-of-a-dungeon-opens-into-the-underworld.md` |
| `dword_5FFF4` | eax = offset dword_5FFF4          ; IDA 拿 debug 旗標當基底,實際上是四個獨立全域 | `03-scene-entry-and-tile-semantics.md`, `92-white-noise-shim-and-ambient-sound.md` |
| `dword_600A4` | cmp     dword_600A4[edi*4], 0    ; 沒載進來就不放 | `63-fm-towns-sound-effects.md` |
| `dword_60100` | mov     eax, dword_60100[edi*4]  ; ← 音量(表的第二欄) | `63-fm-towns-sound-effects.md` |
| `dword_6015C` | 1  /  0x19  /  0x3E8  /  `dword_6015C = 1`  /  (不出聲,只武裝旗標) / | `92-white-noise-shim-and-ambient-sound.md` |
| `dword_601C0` | ⬜ `sub_3FE4` 開頭的 `dword_5FF34 = 0` 與 `dword_601C0 == 1 → sub_2C250` | `78-the-bottom-of-a-dungeon-opens-into-the-underworld.md` |
| `dword_601D0` | 資料  /  `byte_601F0`(距離表)、`dword_601D0`(擋視線地形)、`dword_601E4`(發光地形)、`byte_5FF5C`(日出漸變)  / | `31-line-of-sight.md` |
| `dword_601E4` | 資料  /  `byte_601F0`(距離表)、`dword_601D0`(擋視線地形)、`dword_601E4`(發光地形)、`byte_5FF5C`(日出漸變)  / | `31-line-of-sight.md` |
| `dword_60974` | ⇒ `dword_60974` 是「能不能播 CD」的閘門;不能播時傳索引 **3**,而索引 3 的 | `90-cdda-scenes-and-sound-effect-indices.md` |
| `dword_651B4` | eax += offset dword_651B4               ; 表的基底 | `87-bgm-selection-and-the-tbl-tables.md` |
| `dword_65334` | ⇒ **`sub_3181C(n)` = 播第 n 首 BGM**;`dword_65334` = 當前曲目、`dword_65338` = 前一首。 | `03-scene-entry-and-tile-semantics.md`, `87-bgm-selection-and-the-tbl-tables.md` |
| `dword_65338` | ⇒ **`sub_3181C(n)` = 播第 n 首 BGM**;`dword_65334` = 當前曲目、`dword_65338` = 前一首。 | `03-scene-entry-and-tile-semantics.md`, `87-bgm-selection-and-the-tbl-tables.md` |
| `loc_12D` | if (al >= 0x21)  goto loc_12D        ; 地牢 | `81-three-mode-loops-are-mutually-exclusive.md` |
| `loc_142` | goto loc_142 | `81-three-mode-loops-are-mutually-exclusive.md` |
| `loc_1BE` | if (al == 0)     goto loc_1BE        ; 還在大地圖 ⇒ 只可能是要結束 | `81-three-mode-loops-are-mutually-exclusive.md` |
| `loc_243` | cmp   esi, 70h  ; jz  loc_243      ← 衛兵 → 跳去檢查 0xB4 | `29-npc-behaviour-and-arrest.md` |
| `loc_24B` | cmp   esi, 80h  ; jl  loc_24B      ← < 0x80 → 記 | `29-npc-behaviour-and-arrest.md` |
| `loc_630` | jbe     short loc_630 | `03-scene-entry-and-tile-semantics.md` |
| `loc_88B` | sub_2C118()                     ; ← sub_86C:loc_88B(移動)呼叫 | `92-white-noise-shim-and-ambient-sound.md` |
| `loc_1120` | fs:loc_1120` / `fs:loc_1180` —— 也就是一張放在**低位線性位址的入口表**。 | `89-music-playback-is-not-in-the-game-binary.md` |
| `loc_1180` | call large fword ptr fs:loc_1180     ; ★ 遠端呼叫 linear 0x1180 的入口表 | `89-music-playback-is-not-in-the-game-binary.md` |
| `loc_135D` | jnz     loc_135D          ; ★ ebx != 0 → 跳過維生開銷 | `70-hunger-poison-and-the-vanishing-rings.md` |
| `loc_1594` | loc_1594: | `28-shadowlords-and-blackthorn.md` |
| `loc_17A5` | jmp  short loc_17A5 | `11-map-objects.md` |
| `loc_17B5` | jge  short loc_17B5 | `11-map-objects.md` |
| `loc_202C` | off_48A88       dd offset loc_202C      ; DATA XREF: sub_A310+1E↑o | `67-corpser-and-the-sleeping-party-member.md`, `77-a-false-gap-and-the-failed-line.md` |
| `loc_22E6` | 指令長度累加得到四個標籤的位址,兩個尾巴分別是 `loc_22E6`(**回 0**)與 | `88-the-load-window-culling-and-ship-cannon.md` |
| `loc_22EA` | `loc_22EA`(**回 1**): | `88-the-load-window-culling-and-ship-cannon.md` |
| `loc_274B` | loc_274B: and ebx, ebx ; jnz 出去    ;  / dy /  != 0 → 不開 | `88-the-load-window-culling-and-ship-cannon.md` |
| `loc_2754` | loc_2754: 印 "* BOOOM! *\n\n" ; sub_23FC(槽,  / dx / ,  / dy / ) ; return 1 | `88-the-load-window-culling-and-ship-cannon.md` |
| `loc_2E11` | cmp al, 14h ; jnb loc_2E11                ; 計數器 >= 0x14 → loc_2E11(push 1) | `84-the-sailing-gate-and-the-wind-table.md`, `88-the-load-window-culling-and-ship-cannon.md` |
| `loc_2E13` | push 0 ; jmp loc_2E13                     ; 否則 push 0 | `84-the-sailing-gate-and-the-wind-table.md`, `88-the-load-window-culling-and-ship-cannon.md` |
| `loc_2E1C` | `loc_2E11` 是**移動**、`loc_2E1C` 是**跳過**。重新讀敵船那一段: | `84-the-sailing-gate-and-the-wind-table.md` |
| `loc_2E65` | jz      short loc_2E65 | `38-terrain-movement-cost.md` |
| `loc_2E7C` | jnz     short loc_2E7C | `38-terrain-movement-cost.md` |
| `loc_2ED1` | loc_2ED1:  mov   edi, 1Fh | `88-the-load-window-culling-and-ship-cannon.md` |
| `loc_2ED6` | loc_2ED6:  and   edi, edi ; jle 出去                    ; ⚠ 槽 0 掃不到 | `88-the-load-window-culling-and-ship-cannon.md` |
| `loc_3F29` | jnz short loc_3F29      ; ← tile != 0 就失敗 | `18-dungeons.md` |
| `loc_3F2B` | jnz short loc_3F2B | `18-dungeons.md` |
| `loc_50E6` | jz   short loc_50E6     ; 一樣 → 跳過 " from the X" | `48-dungeon-wandering-monster-and-arena.md` |
| `loc_61B2` | jnz  short loc_61B2 | `90-cdda-scenes-and-sound-effect-indices.md` |
| `loc_61B4` | jmp  short loc_61B4 | `90-cdda-scenes-and-sound-effect-indices.md` |
| `loc_7804` | jz   short loc_7804 | `54-pass-vehicle-verbs-and-the-avatar-line.md` |
| `loc_7809` | jmp  short loc_7809 | `54-pass-vehicle-verbs-and-the-avatar-line.md` |
| `loc_7A03` | jnz     short loc_7A03      ; ← 其他鍵回去重讀 | `61-tavern-lore-and-the-transfer-second-stage.md` |
| `loc_7A1E` | jz      short loc_7A1E | `61-tavern-lore-and-the-transfer-second-stage.md` |
| `loc_7A30` | jz      short loc_7A30       ; ★ 什麼都不打 → 保留原名 | `61-tavern-lore-and-the-transfer-second-stage.md` |
| `loc_7B92` | jnz     short loc_7B92 | `61-tavern-lore-and-the-transfer-second-stage.md` |
| `loc_7BA0` | jz      short loc_7BA0 | `61-tavern-lore-and-the-transfer-second-stage.md` |
| `loc_7BB3` | jnz     short loc_7BB3 | `61-tavern-lore-and-the-transfer-second-stage.md` |
| `loc_7BC4` | jmp     short loc_7BC4 | `61-tavern-lore-and-the-transfer-second-stage.md` |
| `loc_9E4D` | jz   short loc_9E4D | `16-combat-turns-and-ai.md` |
| `loc_9F00` | jge  loc_9F00              ; ← random(0,255) >= 128 就整個不射 | `16-combat-turns-and-ai.md` |
| `loc_A340` | jz      short loc_A340 | `67-corpser-and-the-sleeping-party-member.md` |
| `loc_A595` | ★ 關鍵在 `loc_A595` 的位置:它在 `sub_2ED50` **之後**、`push aZzzzz` 那一行。 | `67-corpser-and-the-sleeping-party-member.md` |
| `loc_AC0C` | jnz  loc_AC0C            ; 還沒輪到 | `16-combat-turns-and-ai.md` |
| `loc_B4AB` | cmp edi, 2Fh ; jle loc_B4AB | `17-magic.md` |
| `loc_B4CA` | cmp edi, 32h ; jl  loc_B4CA   ← 0x30 / 0x31 自動命中 | `17-magic.md` |
| `loc_BD57` | jle     loc_BD57                              ; 掙不脫 → 直接回 | `67-corpser-and-the-sleeping-party-member.md` |
| `loc_D0A3` | jnz     short loc_D0A3            ; 沒找到醒著的 → 跳過 | `37-look-signs-and-the-sky.md` |
| `loc_D2C1` | loc_D2C1: | `65-the-wishing-well-easter-egg.md` |
| `loc_D2EE` | jnz     short loc_D2EE | `65-the-wishing-well-easter-egg.md` |
| `loc_D3BB` | `loc_D3BB` 那一段是對 `esi`(= x)做 switch,不是查地點表: | `37-look-signs-and-the-sky.md` |
| `loc_FE53` | loc_FE53: | `53-party-tile-and-the-arena-floor.md` |
| `loc_FE83` | jge   short loc_FE83 | `53-party-tile-and-the-arena-floor.md` |
| `loc_10ECE` | ⚠ **等級沒變就整段跳過**(`cmp edx, ebx; jz loc_10ECE`)—— 連魔力都不重算。 | `19-levelup.md` |
| `loc_14A45` | jnz     short loc_14A45 | `43-search.md` |
| `loc_14BF5` | jnz short loc_14BF5 | `76-jimmy-in-a-dungeon-disarms-not-unlocks.md` |
| `loc_14DC0` | ⚠ 魔法鎖那條是「必定失敗**而且照樣扣鑰匙**」——`loc_14DC0` 直接跳到扣鑰匙那段。 | `41-jimmy-neworder-gem-ztats.md` |
| `loc_1546A` | loc_1546A: | `75-open-did-nothing-outside-dungeons.md` |
| `loc_15863` | 位址  /  `sub_154BC`(Get 分派)`loc_15863` / `loc_158A2` / `loc_158D0` / `loc_158DE`  / | `57-crown-and-sceptre-placement.md` |
| `loc_158A2` | 位址  /  `sub_154BC`(Get 分派)`loc_15863` / `loc_158A2` / `loc_158D0` / `loc_158DE`  / | `57-crown-and-sceptre-placement.md` |
| `loc_158D0` | loc_158D0:  … jmp short loc_158EA → loc_15903                            ; 權杖:只有共同尾段 | `57-crown-and-sceptre-placement.md` |
| `loc_158DE` | 位址  /  `sub_154BC`(Get 分派)`loc_15863` / `loc_158A2` / `loc_158D0` / `loc_158DE`  / | `44-use-item.md`, `57-crown-and-sceptre-placement.md` |
| `loc_158EA` | loc_158D0:  … jmp short loc_158EA → loc_15903                            ; 權杖:只有共同尾段 | `57-crown-and-sceptre-placement.md` |
| `loc_15903` | loc_158D0:  … jmp short loc_158EA → loc_15903                            ; 權杖:只有共同尾段 | `57-crown-and-sceptre-placement.md` |
| `loc_167E9` | `sub_165C8` 紮營  /  `loc_167E9`,突襲骰前面一行  /  **每睡一小時**  / | `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `loc_16C0C` | jle   short loc_16C0C | `50-hole-up-camp-sleep-repair.md` |
| `loc_1747D` | jnz     loc_1747D                ; 都不是 → 繼續飛 | `69-the-overworld-cannon-did-nothing.md` |
| `loc_1751C` | loc_1751C: | `69-the-overworld-cannon-did-nothing.md` |
| `loc_17539` | jz      short loc_17539          ; kind & F8 == 78 → 去查特例 | `69-the-overworld-cannon-did-nothing.md` |
| `loc_17547` | jnz     short loc_17547          ; ★ 這道永遠成立 —— 見下 | `69-the-overworld-cannon-did-nothing.md` |
| `loc_17E4B` | jnz  short loc_17E4B | `26-yell-words-of-power-shadowlords.md` |
| `loc_1869F` | 3. **藥草不足是重問份數,不是取消。** `sub_18698` 裡 `var_4 = 0` 之後跳回 `loc_1869F` | `45-fire-and-mix.md`, `58-rune-input-cast-and-mix.md` |
| `loc_186F1` | jle     short loc_186F1     ; 夠 → 檢查下一種 | `45-fire-and-mix.md` |
| `loc_186F6` | loc_186F6: | `45-fire-and-mix.md` |
| `loc_1881F` | 藥草**先扣再判成敗**(`loc_1881F` 扣完才到 `loc_18836` 比對配方)—— | `45-fire-and-mix.md` |
| `loc_18836` | 藥草**先扣再判成敗**(`loc_1881F` 扣完才到 `loc_18836` 比對配方)—— | `45-fire-and-mix.md` |
| `loc_19CBC` | 跳表 case 14 / 15 / 16 / 20 全部 `jmp loc_19CBC`,只差事先 push 的種類碼: | `17-magic.md` |
| `loc_1A2F1` | `sub_1A0B0`(藥水)在「沒選到人」時走的是 `jmp loc_1A2F1` —— | `77-a-false-gap-and-the-failed-line.md` |
| `loc_1A6AB` | if (n < 8)   { eax = sub_19ED8(n);     jmp loc_1A6AB }   ; 卷軸 | `77-a-false-gap-and-the-failed-line.md` |
| `loc_1A6B3` | 21..28 在 `loc_1A6B3` 那六行被送去 `sub_1A2F8`(埋月石)。 | `71-the-use-list-had-29-empty-slots.md` |
| `loc_1A9A3` | loc_1A9A3: | `80-implemented-but-unreachable.md` |
| `loc_1AB32` | 實際的 case 37(`loc_1AB32`)只有兩行: | `77-a-false-gap-and-the-failed-line.md` |
| `loc_1AB37` | loc_1AB37: | `77-a-false-gap-and-the-failed-line.md` |
| `loc_1AB3C` | if (n > 20 && n < 29) { sub_1A2F8(n − 21); jmp loc_1AB3C }   ; 月石 —— 跳過賦值 | `77-a-false-gap-and-the-failed-line.md` |
| `loc_1B14C` | jnz short loc_1B14C       ; needle 還沒完 → 繼續 | `32-guard-challenge.md` |
| `loc_1B176` | loc_1B176:  cmp byte ptr [edi], 0 | `32-guard-challenge.md` |
| `loc_1B522` | (`cmp eax, esi; jb loc_1B522`),不夠就抓。原版在這裡**不多說一句話**, | `32-guard-challenge.md` |
| `loc_1C272` | loc_1C272:                       ; case 0x8C | `79-four-dialogue-opcodes-were-corrupting-the-text.md` |
| `loc_1D3D6` | jnz short loc_1D3D6 | `25-shrines.md` |
| `loc_1D5B5` | jl  short loc_1D5B5   ; 不是數字就繼續等 | `25-shrines.md` |
| `loc_1E900` | jz      short loc_1E900      ; ← 等於 0xFF 才跳 | `94-ztats-has-seventeen-pages.md` |
| `loc_1E905` | jmp     short loc_1E905 | `94-ztats-has-seventeen-pages.md` |
| `loc_1F2E5` | `sub_1EFC8` 裡 0xD3 / 0xD4 兩個鍵碼的分支(`loc_1F2E5` / `loc_1F303`)。 | `60-command-echo-and-menu-keys.md` |
| `loc_1F303` | `sub_1EFC8` 裡 0xD3 / 0xD4 兩個鍵碼的分支(`loc_1F2E5` / `loc_1F303`)。 | `60-command-echo-and-menu-keys.md` |
| `loc_20F34` | jz      short loc_20F34         ; ← al 是 0xFF,這個 jz 永遠不跳 | `93-the-tavern-meal-was-never-haggled.md` |
| `loc_20F3B` | jmp     short loc_20F3B | `93-the-tavern-meal-was-never-haggled.md` |
| `loc_211A9` | jnz     loc_211A9                ; ★ 不等於 3 就直接進酒單 | `93-the-tavern-meal-was-never-haggled.md` |
| `loc_21598` | jz      short loc_21598      ; 0 = 沒找到 → 下一個主題 | `61-tavern-lore-and-the-transfer-second-stage.md` |
| `loc_2159C` | loc_2159C:      mov     edi, esi             ; ← 兩個分支與 fall-through 都到這裡 | `61-tavern-lore-and-the-transfer-second-stage.md` |
| `loc_2159F` | jmp     short loc_2159F | `61-tavern-lore-and-the-transfer-second-stage.md` |
| `loc_21F0C` | `sub_21D48` 旅店  /  `loc_21F0C`,`sub_29304(9)` 前面  /  **每推 9 分鐘**  / | `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `loc_23398` | loc_23398: | `39-character-creation.md` |
| `loc_235AD` | (`loc_235AD` 把 `byte_3DDC0..C2` 寫進 `+0x2A..0x2C`)。所以引擎從 | `39-character-creation.md` |
| `loc_235F3` | loc_235F3:  4 次 sub_23274      八德剩四 | `39-character-creation.md` |
| `loc_2361B` | loc_2361B:  2 次 sub_23274      剩二 | `39-character-creation.md` |
| `loc_23641` | loc_23641:  1 次 sub_23274      決出唯一存活者 | `39-character-creation.md` |
| `loc_23740` | ja      short loc_23740 | `39-character-creation.md` |
| `loc_23747` | loc_23747: | `39-character-creation.md` |
| `loc_29463` | 5. 日照半徑(`sub_29304` 的 `loc_29463`) | `31-line-of-sight.md` |
| `loc_29AA6` | jnz  short loc_29AA6 | `16-combat-turns-and-ai.md` |
| `loc_29C60` | and  al, al      ; jz  loc_29C60       ; ★ 計數器 0   → 一般畫法 | `86-tile-0xDC-is-an-open-moongate.md` |
| `loc_29EB4` | `sub_29D64` 在 `cmp byte_3E0A3, 80h; jnb loc_29EB4` 分岔:地點編號 ≥ 0x80 | `31-line-of-sight.md` |
| `loc_2A563` | loc_2A563: | `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `loc_2A5C0` | if (word_3DFB4 == 0) { 印 "Starving!"; sub_2A4D0(); jmp loc_2A5C0 } | `70-hunger-poison-and-the-vanishing-rings.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `loc_2A5CA` | if (byte_3E08F == byte_3E090)  goto loc_2A5CA         ; 這小時已結算 | `70-hunger-poison-and-the-vanishing-rings.md`, `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `loc_2A606` | loc_2A606: | `99-the-tail-of-upkeep-and-the-wandering-horses.md` |
| `loc_2AA44` | ⬜ `sub_2A984` 的地下世界分支(`loc_2AA44`)畫什麼沒讀。 | `80-implemented-but-unreachable.md` |
| `loc_2AE01` | loc_2AE01:                       ; case 32(空白) | `54-pass-vehicle-verbs-and-the-avatar-line.md` |
| `loc_2AE26` | jnz   short loc_2AE26 | `54-pass-vehicle-verbs-and-the-avatar-line.md` |
| `loc_2AE30` | jmp   short loc_2AE30 | `54-pass-vehicle-verbs-and-the-avatar-line.md` |
| `loc_2B09D` | loc_2B09D:   (按鍵 'T') | `35-harp-and-the-secret-door.md` |
| `loc_2B115` | ⚠ **不扣寶石。** `loc_2B115` 只檢查數量,看完寶石還在。這很反直覺, | `41-jimmy-neworder-gem-ztats.md` |
| `loc_2BB9E` | loc_2BB9E:                  ; 地表 | `48-dungeon-wandering-monster-and-arena.md`, `50-hole-up-camp-sleep-repair.md` |
| `loc_2C5DA` | jnz     short loc_2C5DA | `92-white-noise-shim-and-ambient-sound.md` |
| `loc_2CE50` | loc_2CE50:                          ; 想去的方向就是現在的船頭(不需轉向) | `84-the-sailing-gate-and-the-wind-table.md` |
| `loc_2CE67` | jnb  short loc_2CE67            ; ★ >= 0x24(收帆的大船)→ 回 0 = 照走 | `84-the-sailing-gate-and-the-wind-table.md` |
| `loc_2D40A` | `loc_2D40A`(上面的「走一步」)做了四件事:推時鐘、可能跑一個世界回合、 | `83-monster-attacks-and-two-different-wind-rules.md` |
| `loc_2DBAF` | 指令分派(`sub_2ACF4`)在中間的 `loc_2DBAF`。 | `81-three-mode-loops-are-mutually-exclusive.md` |
| `loc_2DBF4` | 而每回合的收尾(`sub_29304(2)` 起)在 **line 80013** 的 `loc_2DBF4`。 | `81-three-mode-loops-are-mutually-exclusive.md` |
| `loc_2DC34` | jnz     short loc_2DC34 | `68-troll-bridge-and-collision.md` |
| `loc_2DC5F` | ✅ **`(0xE9, 0xEB)` 的「Pass, Seeker!」關卡**(`sub_2D9D0` `loc_2DC5F`): | `81-three-mode-loops-are-mutually-exclusive.md` |
| `loc_2DCC5` | 大地圖  /  `sub_2DD44` → `sub_2D9D0`  /  **2**(`sub_29304(2)`)  /  `sub_2D9D0` 自己那一串  /  `loc_2DCC5`  / | `80-implemented-but-unreachable.md`, `81-three-mode-loops-are-mutually-exclusive.md` |
| `loc_2DD2F` | ✅ **每個大地圖回合都跑一個世界回合。** `loc_2DD2F: call sub_2E24` 是**無條件**的。 | `81-three-mode-loops-are-mutually-exclusive.md` |
| `loc_2E074` | loc_2E074:  var_41D = 0xFF        ; 超出半徑 → 連寫都不寫 | `31-line-of-sight.md` |
| `loc_3185B` | jnz     short loc_3185B | `03-scene-entry-and-tile-semantics.md` |
| `loc_318BF` | loc_318BF:                              ; 淡出迴圈 | `87-bgm-selection-and-the-tbl-tables.md` |
| `loc_3197C` | loc_3197C: | `03-scene-entry-and-tile-semantics.md` |
| `loc_31CC5` | loc_31CC5:  mov     eax, dword_65334      ; return dword_65334 | `03-scene-entry-and-tile-semantics.md` |
