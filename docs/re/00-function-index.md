# RE-00b:函式與全域索引(自動產生)

> `python3 tools/gen_func_index.py > docs/re/00-function-index.md` 重新產生。
> **讀任何 `sub_XXXX` 之前先查這裡** —— 筆記超過二三十份後,憑記憶一定會重讀已解過的東西。
>
> 目前收錄 **329** 個符號,來源是 `docs/` 下的逆向筆記。

| 符號 | 已知語意(取自筆記) | 出處 |
|---|---|---|
| `sub_324` | 場景載入時還會依時間切換:`if (hour < 5  /  /  hour > 19) sub_324()` —— 夜間的燈光處理。 | `04-npc-schedule-and-clock.md` |
| `sub_48C` | 0xFC  /  某種特殊物件  /  `sub_48C` 開場掃全表看在不在場;是什麼還沒追到  / | `11-map-objects.md` |
| `sub_5C8` | `sub_5C8` 就是載入場景地圖的函式。組語三行講完: | `03-scene-entry-and-tile-semantics.md`, `11-map-objects.md` |
| `sub_758` | 樓層增減(`sub_758`)  /  `byte_3E0A5 = 1` / `= -1`  /  `inc` / `dec byte_3E0A5`  /  照抄的話只能在 1F 與 B1F 之間跳  / | `03-scene-entry-and-tile-semantics.md` |
| `sub_86C` | 方向樓梯  /  0xC4–0xC7(低 2 bit = 朝向)  /  **走進去**:同向上樓、反向(`facing ^ 2`)下樓  /  `sub_86C` → `sub_758`  / | `02-movement-and-tile-flags.md`, `03-scene-entry-and-tile-semantics.md` |
| `sub_B98` | 0x40–0x73  /  **人**(`sub_B98` 判「可以被嚇跑的平民」用的就是這個範圍),4 的倍數  / | `04-npc-schedule-and-clock.md`, `09-items-and-creatures.md` |
| `sub_C10` | 0x8B  /  叫衛兵  /  `sub_C10` 掃 32 個 NPC 槽找 tile 0x70 那批  / | `06-conversation-script.md`, `14-combat-maps.md` |
| `sub_EA0` | 梯子  /  0xC8 上 / 0xC9 下 / 0x86 活板門(下)  /  站在上面按 **K**(Klimb)  /  `sub_EA0` → `sub_758(0 或 2, 196)`  / | `03-scene-entry-and-tile-semantics.md` |
| `sub_1678` | 0  /  空槽  /  `sub_1678` 清表寫 0;`sub_118CC` 用 `!= 0` 找空槽  / | `11-map-objects.md` |
| `sub_1DC8` | 每月 28 天、每年 13 個月**。一般行動每回合 **1 分鐘**(`sub_1DC8` → `sub_29304(1)`); | `04-npc-schedule-and-clock.md` |
| `sub_23FC` | 船的四個朝向 tile 0x2C..0x2F 確實存在(`sub_23FC` 的轉向),但那是 | `11-map-objects.md` |
| `sub_3F34` | Uus Por / Des Por  /  `sub_3F34(∓1, 1)`  /  地牢上 / 下一層  / | `17-magic.md` |
| `sub_6730` | 2. 或查 `byte_41C18` 的 xref(`sub_6730` 開頭有 `for(i=0;i<256;i++) byte_41C18[i]=i`, | `01-tileset-and-dot16-loader.md` |
| `sub_8858` | 來源:FM Towns `WORRIORS.EXP`。相關函式 `sub_8858`(載入)、`sub_9C7C`(排程 slot)、 | `04-npc-schedule-and-clock.md` |
| `sub_8924` | 0 號槽是隊伍自己**:`sub_8924` 的更新迴圈從 `esi = 1` 起跑。檔案裡 0 號槽的內容 | `04-npc-schedule-and-clock.md`, `12-npc-movement.md` |
| `sub_89EC` | 跨樓層的樓梯口選擇**(模式 4/5/6/7):原版先用 `sub_89EC` 找樓梯、走過去、 | `12-npc-movement.md` |
| `sub_8A1C` | ├ sub_8BA0    尋路 —— 內含 sub_8A1C 建格子圖 + 環狀佇列 BFS | `12-npc-movement.md` |
| `sub_8BA0` | dir  /  BFS(`sub_8BA0`)  /  走路 / 回溯(`sub_8EA4`、`sub_8D28`)  / | `12-npc-movement.md` |
| `sub_8D28` | dir  /  BFS(`sub_8BA0`)  /  走路 / 回溯(`sub_8EA4`、`sub_8D28`)  / | `12-npc-movement.md` |
| `sub_8EA4` | dir  /  BFS(`sub_8BA0`)  /  走路 / 回溯(`sub_8EA4`、`sub_8D28`)  / | `12-npc-movement.md` |
| `sub_91A4` | 只有模式 ≤ 1(不存在或閒置)才重新呼叫 `sub_91A4`**。正在移動中的 NPC | `12-npc-movement.md` |
| `sub_9358` | 欄位位置由 `sub_9358` 證實(`rec[slot+3]` / `rec[slot+6]` / `rec[slot+9]`,slot ∈ 0..2)。 | `04-npc-schedule-and-clock.md`, `12-npc-movement.md` |
| `sub_9428` | 建圖(`sub_8A1C`)  /  走一步(`sub_9428`)  / | `12-npc-movement.md` |
| `sub_94E0` | `sub_94E0`** —— 尋路失敗時的 fallback(疑似隨機遊走),還沒讀。 | `12-npc-movement.md` |
| `sub_9690` | ⚠ 還有一條容易漏的:`sub_9690` 用 `cmp word ptr [ebx], 1 / jg` 判斷 —— | `11-map-objects.md`, `12-npc-movement.md` |
| `sub_9C7C` | 來源:FM Towns `WORRIORS.EXP`。相關函式 `sub_8858`(載入)、`sub_9C7C`(排程 slot)、 | `04-npc-schedule-and-clock.md`, `08-shops.md` |
| `sub_9E10` | 0x8000  /  施法  /  `sub_9E10` / `sub_9F08`  /  法師、注視者、收割者、惡魔、海馬  / | `16-combat-turns-and-ai.md` |
| `sub_9F08` | 0x8000  /  施法  /  `sub_9E10` / `sub_9F08`  /  法師、注視者、收割者、惡魔、海馬  / | `16-combat-turns-and-ai.md` |
| `sub_A108` | 1. **敵人整個不動**:`sub_A108` 一開頭 `cmp byte_3E08A, 'T'` 就 return。 | `16-combat-turns-and-ai.md`, `17-magic.md` |
| `sub_A360` | 同一類的還有 `sub_A360` 開頭:**握著混沌之劍(裝備 0x23)的隊員** | `16-combat-turns-and-ai.md`, `17-magic.md` |
| `sub_A9EC` | `sub_A9EC` 是一個 0..31 的無窮掃描: | `16-combat-turns-and-ai.md` |
| `sub_AC40` | 在 `sub_AC40` 裡被當成 (dx, dy) 暫存重用** —— 同名不同義,追值的時候 | `16-combat-turns-and-ai.md` |
| `sub_AE20` | 0x2000  /  **瞬間移動**  /  `sub_AE20`  /  鬼火  / | `16-combat-turns-and-ai.md` |
| `sub_B274` | `sub_B274` 對角色讀 `byte_3DDCC[角色*32]`,而 `0x3DDCC − 0x3DDB4 = 0x18`。 | `15-combat-formulas.md`, `16-combat-turns-and-ai.md` |
| `sub_B35C` | │       ├ sub_B35C   戰場單位的 +1 欄位 | `15-combat-formulas.md` |
| `sub_B398` | `sub_B398` 印證了這件事:它取 `byte_3F050[生物*8]` 當「力量那一項」、 | `15-combat-formulas.md` |
| `sub_B484` | push edi(攻擊者); push ebx(目標); call sub_B484` → arg_0 = 目標。 | `15-combat-formulas.md`, `16-combat-turns-and-ai.md` |
| `sub_B51C` | 否則                                   → sub_B274 算傷害 → sub_B51C 扣血 | `16-combat-turns-and-ai.md` |
| `sub_B8DC` | 0x0004 / 0x0200  /  **下毒**  /  `sub_B9A8` → `sub_B8DC`  /  巨蟒、大烏賊、巨蜘蛛 / 巨鼠、擬態怪  / | `16-combat-turns-and-ai.md` |
| `sub_B9A8` | 0x0004 / 0x0200  /  **下毒**  /  `sub_B9A8` → `sub_B8DC`  /  巨蟒、大烏賊、巨蜘蛛 / 巨鼠、擬態怪  / | `15-combat-formulas.md`, `16-combat-turns-and-ai.md` |
| `sub_C778` | 寫  0xC790   sub_C778    mov dword_65334, 1 | `03-scene-entry-and-tile-semantics.md` |
| `sub_10910` | call    sub_10910               ; → byte_3E0A5 == 0 ? "A:BRIT.OOL" : "A:UNDER.OOL" | `11-map-objects.md` |
| `sub_10928` | `sub_1DA10`(聖壇)、`sub_2D564`(cave/mine/dungeon)、`sub_10928`(印地名並確認)。 | `03-scene-entry-and-tile-semantics.md` |
| `sub_10FEC` | `sub_11168`(讀 SHOPPE.DAT)、`sub_10FEC`(代換佔位符)、`sub_9C7C`(營業時間)。 | `05-text-compression.md`, `08-shops.md`, `10-shop-prices-and-trade.md` |
| `sub_11168` | `sub_11168`(讀 SHOPPE.DAT)、`sub_10FEC`(代換佔位符)、`sub_9C7C`(營業時間)。 | `08-shops.md` |
| `sub_111CC` | 來源:FM Towns `WORRIORS.EXP`。`sub_1B294`(進店)、`sub_111CC`(挑問候語)、 | `08-shops.md` |
| `sub_112F8` | 買(`sub_11AF0`、`sub_11588`、`sub_112F8`、`sub_118CC`、`sub_219B0` 都一樣) | `10-shop-prices-and-trade.md` |
| `sub_11464` | 公會  /  0x86  /  `sub_11520` → `sub_11464` → `sub_112F8`  /  3  /  ✅  / | `10-shop-prices-and-trade.md` |
| `sub_11520` | 公會  /  0x86  /  `sub_11520` → `sub_11464` → `sub_112F8`  /  3  /  ✅  / | `10-shop-prices-and-trade.md` |
| `sub_11588` | 買(`sub_11AF0`、`sub_11588`、`sub_112F8`、`sub_118CC`、`sub_219B0` 都一樣) | `10-shop-prices-and-trade.md` |
| `sub_1173C` | 藥草鋪  /  0x85  /  `sub_11864` → `sub_1173C` → `sub_11588`  /  5  /  ✅  / | `10-shop-prices-and-trade.md` |
| `sub_11864` | 藥草鋪  /  0x85  /  `sub_11864` → `sub_1173C` → `sub_11588`  /  5  /  ✅  / | `10-shop-prices-and-trade.md` |
| `sub_118CC` | 買(`sub_11AF0`、`sub_11588`、`sub_112F8`、`sub_118CC`、`sub_219B0` 都一樣) | `10-shop-prices-and-trade.md`, `11-map-objects.md` |
| `sub_11AF0` | 買(`sub_11AF0`、`sub_11588`、`sub_112F8`、`sub_118CC`、`sub_219B0` 都一樣) | `10-shop-prices-and-trade.md` |
| `sub_12060` | 賣(`sub_12060`) | `10-shop-prices-and-trade.md` |
| `sub_1258C` | 武具店  /  0x81  /  `sub_1258C`  /  9  /  ✅ 買 + 賣  / | `10-shop-prices-and-trade.md` |
| `sub_12794` | `sub_12794` 收錢時有一條**與地點綁定的例外**: | `10-shop-prices-and-trade.md` |
| `sub_12838` | 解毒 20、療傷 35、復活 200。各有前置判斷(`sub_12838`): | `10-shop-prices-and-trade.md` |
| `sub_15DD4` | `sub_15DD4` 每次行動後重數兩邊。⚠ 它用的兩個全域 `word_3E086` / `word_3E088` | `16-combat-turns-and-ai.md` |
| `sub_16370` | 倒數走在 `sub_16370`,而那是**玩家單位回合結束時**呼叫的 —— | `17-magic.md` |
| `sub_16454` | 0x02  /  逃跑中  /  `sub_AC40` 反轉方向、`sub_16454` 放行出界  / | `16-combat-turns-and-ai.md` |
| `sub_16538` | 我方全滅時原版**不會立刻判負**:先叫 `sub_16538` 找一個被魅惑的隊員, | `16-combat-turns-and-ai.md` |
| `sub_16DA4` | 上馬 / 上毯 / 上小艇都要求**先下來走路**(`sub_16DA4` 判 `byte_3E08C` ∈ {0x1C, 0x1D}) | `11-map-objects.md` |
| `sub_16DC8` | 上大船的限制不同(`sub_16DC8` 的跳表):放行魔毯、步行、小艇 —— | `11-map-objects.md` |
| `sub_16E58` | `sub_16E58` 是「附近有沒有陸地」:查視窗的 (4,5)(6,5)(5,4)(5,6),也就是玩家四鄰 | `11-map-objects.md` |
| `sub_16F08` | 8. Board / X-it(`sub_16F08` / `sub_177AC`) | `11-map-objects.md` |
| `sub_17630` | 點一把火把是 `random(0,15) + 0x70`(`sub_17630`),也就是 112..127 分鐘。 | `17-magic.md` |
| `sub_177AC` | 8. Board / X-it(`sub_16F08` / `sub_177AC`) | `11-map-objects.md` |
| `sub_18704` | ⚠ **位元順序是反的**。原版 `sub_18704` 用 `mov ebx, 80h` 起頭、每輪 `sar ebx, 1`, | `17-magic.md` |
| `sub_189E4` | Grav Por / Vas Flam / Xen Corp  /  `sub_189E4(0x30/0x31/0x32)`  /  指定目標的攻擊咒語  / | `17-magic.md` |
| `sub_18AF0` | An Zu  /  `sub_18AF0`  /  狀態 `'S'` → `'G'`,戰場上清掉睡著旗標  / | `17-magic.md` |
| `sub_18B88` | An Nox  /  `sub_18B88`  /  狀態 `'P'` → `'G'`  / | `17-magic.md` |
| `sub_18C00` | 其餘的效果函式看得到入口(`sub_18C00`、`sub_18D18`、`sub_18EB0`、`sub_18F2C`、 | `17-magic.md` |
| `sub_18D18` | 其餘的效果函式看得到入口(`sub_18C00`、`sub_18D18`、`sub_18EB0`、`sub_18F2C`、 | `17-magic.md` |
| `sub_18EB0` | 其餘的效果函式看得到入口(`sub_18C00`、`sub_18D18`、`sub_18EB0`、`sub_18F2C`、 | `17-magic.md` |
| `sub_18F2C` | 其餘的效果函式看得到入口(`sub_18C00`、`sub_18D18`、`sub_18EB0`、`sub_18F2C`、 | `17-magic.md` |
| `sub_1904C` | `sub_1904C`、`sub_19098`、`sub_1D1B8`、`sub_1D31C`、`sub_19264`、`sub_192BC`、 | `17-magic.md` |
| `sub_19098` | `sub_1904C`、`sub_19098`、`sub_1D1B8`、`sub_1D31C`、`sub_19264`、`sub_192BC`、 | `17-magic.md` |
| `sub_19264` | `sub_1904C`、`sub_19098`、`sub_1D1B8`、`sub_1D31C`、`sub_19264`、`sub_192BC`、 | `17-magic.md` |
| `sub_192BC` | `sub_1904C`、`sub_19098`、`sub_1D1B8`、`sub_1D31C`、`sub_19264`、`sub_192BC`、 | `17-magic.md` |
| `sub_193C8` | Vas Mani  /  `sub_193C8`  /  `HP = MaxHP`  / | `17-magic.md` |
| `sub_19440` | `sub_19440`、`sub_195C0`、`sub_19674`、`sub_196A4`、`sub_19810`、`sub_1986C`、 | `17-magic.md` |
| `sub_194CC` | An Xen Ex  /  `sub_194CC`  /  「Creature: X charmed!」  / | `17-magic.md` |
| `sub_195C0` | `sub_19440`、`sub_195C0`、`sub_19674`、`sub_196A4`、`sub_19810`、`sub_1986C`、 | `17-magic.md` |
| `sub_19674` | `sub_19440`、`sub_195C0`、`sub_19674`、`sub_196A4`、`sub_19810`、`sub_1986C`、 | `17-magic.md` |
| `sub_196A4` | `sub_19440`、`sub_195C0`、`sub_19674`、`sub_196A4`、`sub_19810`、`sub_1986C`、 | `17-magic.md` |
| `sub_19810` | `sub_19440`、`sub_195C0`、`sub_19674`、`sub_196A4`、`sub_19810`、`sub_1986C`、 | `17-magic.md` |
| `sub_1986C` | `sub_19440`、`sub_195C0`、`sub_19674`、`sub_196A4`、`sub_19810`、`sub_1986C`、 | `17-magic.md` |
| `sub_198E0` | An Tym  /  `sub_198E0`  /  `byte_3E08A = 'T'`、`byte_3E09E = 10`  / | `17-magic.md` |
| `sub_1994C` | └ sub_1994C   ★ 施法主流程 | `17-magic.md` |
| `sub_1AEB4` | `sub_1CE70`、`sub_1CE0C`、`sub_1AEB4`…),字串也認得出幾個 | `17-magic.md` |
| `sub_1B294` | 來源:FM Towns `WORRIORS.EXP`。`sub_1B294`(進店)、`sub_111CC`(挑問候語)、 | `08-shops.md`, `10-shop-prices-and-trade.md` |
| `sub_1B52C` | 驗收:編號 0x70 → tile 368,算繪出來是持戟的鎧甲士兵,正對應 `sub_1B52C` 那句 | `04-npc-schedule-and-clock.md`, `06-conversation-script.md`, `08-shops.md` |
| `sub_1B760` | 0xFF  /  結束整段對話  /  `sub_1B760` + `sub_1BF08`  / | `06-conversation-script.md` |
| `sub_1B800` | call sub_1B800 | `05-text-compression.md` |
| `sub_1BA80` | 關鍵字表在遇到位元組 0x90 時結束。** 跳段用的 `sub_1BA80(0, 0x90)` 是 | `06-conversation-script.md` |
| `sub_1BAA4` | 2. 把腳本指標**重設到記錄開頭**(`sub_1BAA4(0)`),讀 3 個位元組 —— 那是 NPC 名字的前三個字母。 | `06-conversation-script.md` |
| `sub_1BAFC` | `sub_1BF08` 用 `sub_1BAFC(i*2 + 6)` 印回應。 | `06-conversation-script.md` |
| `sub_1BB3C` | `sub_1BB3C`(插名字)、`sub_1BB5C`(加入隊伍)、`sub_C10`(叫衛兵)、 | `06-conversation-script.md` |
| `sub_1BB5C` | 0x84  /  邀請加入隊伍  /  `sub_1BB5C`,滿員時「Thou hast no room for me…」  / | `06-conversation-script.md`, `07-save-format.md` |
| `sub_1BCB8` | (`sub_1C0AC` → `sub_1BCB8` 從記錄開頭找碼相同的區塊)。玩家回答後: | `06-conversation-script.md` |
| `sub_1BCF4` | 輸入含「是」的觸發字 → 印第 4 段(`sub_1BCF4`),否則印第 2 段(`sub_1BD0C`)。 | `06-conversation-script.md` |
| `sub_1BD0C` | 輸入含「是」的觸發字 → 印第 4 段(`sub_1BCF4`),否則印第 2 段(`sub_1BD0C`)。 | `06-conversation-script.md` |
| `sub_1BD50` | 算式不是猜的:`sub_1BD50(i)` 把指標重設到開頭後跳 `2i+5` 段,命中之後 | `06-conversation-script.md` |
| `sub_1BD8C` | (位置 0,或前一個字元是空白)——`sub_1BD8C` → `sub_27C98`,再檢查 `byte_55F37[i]`。 | `06-conversation-script.md` |
| `sub_1BE28` | 索引  /  字  /  行為(`sub_1BE28`)  / | `06-conversation-script.md` |
| `sub_1BF08` | 0xFF  /  結束整段對話  /  `sub_1B760` + `sub_1BF08`  / | `06-conversation-script.md` |
| `sub_1C0AC` | 0x91–0x9F  /  反問玩家並讀取回答(15 種)  /  `byte_55F32 = 碼; sub_1C0AC` → 印 "You respond-\n:" 後收輸入  / | `06-conversation-script.md` |
| `sub_1C1E8` | 0x85 / 0x86 / 0x8C / 0xFE  /  切到子模式  /  `byte_55F18 = 碼`;函式開頭會把後續位元組改送 `sub_1C1E8`  / | `06-conversation-script.md` |
| `sub_1C2FC` | 0x88  /  `sub_1C2FC`  /  未定  / | `06-conversation-script.md` |
| `sub_1C3F8` | 來源:FM Towns `WORRIORS.EXP` 的 `sub_1C3F8`(展開)與 `dword_41990`(槽表); | `05-text-compression.md`, `06-conversation-script.md` |
| `sub_1C840` | 帶一張 31 路跳表;周邊有 `sub_1C840`(載入記錄)、`sub_1B52C`(分派)、 | `05-text-compression.md`, `06-conversation-script.md` |
| `sub_1CA0C` | 原版問「Spell name:」,玩家把**上古語**打進去(`sub_1CA0C`)。 | `17-magic.md` |
| `sub_1CD3C` | Mani  /  `sub_1CD3C`  /  回 **1..30**(與命中骰同一顆 `sub_2B724`),上限 MaxHP,死人無效  / | `17-magic.md` |
| `sub_1CE0C` | `sub_1CE70`、`sub_1CE0C`、`sub_1AEB4`…),字串也認得出幾個 | `17-magic.md` |
| `sub_1CE70` | `sub_1CE70`、`sub_1CE0C`、`sub_1AEB4`…),字串也認得出幾個 | `17-magic.md` |
| `sub_1CFC8` | In Mani Corp  /  `sub_1CFC8`  /  復活(對活人印「Not dead!」)  / | `17-magic.md` |
| `sub_1D1B8` | `sub_1904C`、`sub_19098`、`sub_1D1B8`、`sub_1D31C`、`sub_19264`、`sub_192BC`、 | `17-magic.md` |
| `sub_1D310` | In Lor / Vas Lor  /  `sub_1D310(100)` / `(255)`  /  光明計時器設成 100 / 255 **分鐘**  / | `17-magic.md` |
| `sub_1D31C` | `sub_1904C`、`sub_19098`、`sub_1D1B8`、`sub_1D31C`、`sub_19264`、`sub_192BC`、 | `17-magic.md` |
| `sub_1DA10` | `0x19`(25)  /  **the shrine of …**(八德聖壇;名字取自 `off_411BC[9]`,狀態看 `byte_411FC[8]`/`byte_41204[8]`)  /  `sub_1DA1… | `03-scene-entry-and-tile-semantics.md` |
| `sub_1E9A0` | sub_1E9A0       Z-stats 畫面(咒語 / 藥草 / 裝備 / 物品四張清單) | `17-magic.md` |
| `sub_1F5A4` | `sub_1F5A4`:魅惑(0x0040)與另外兩個位元(0x0400 / 0x0800)的遠程特殊行為。 | `16-combat-turns-and-ai.md` |
| `sub_1F9CC` | 距離是 `sub_1F9F8`:`sub_1F9CC` 算出 dx²+dy² 之後用「連續減奇數」 | `16-combat-turns-and-ai.md` |
| `sub_1F9F8` | 距離是 `sub_1F9F8`:`sub_1F9CC` 算出 dx²+dy² 之後用「連續減奇數」 | `16-combat-turns-and-ai.md` |
| `sub_1FA6C` | ├ sub_1FA6C   瞄準(射程來自 byte_3F2F8[武器]) | `15-combat-formulas.md` |
| `sub_1FD80` | ├ sub_1FD80   選中目標 | `15-combat-formulas.md` |
| `sub_1FE54` | `sub_1FE54` / `sub_20CB4` 的投射物飛行路徑(會不會被地形擋)。 | `16-combat-turns-and-ai.md` |
| `sub_200BC` | └ sub_200BC   ★ 有沒有人在旁邊干擾 | `17-magic.md` |
| `sub_20134` | sub_20134   一次攻擊 | `15-combat-formulas.md` |
| `sub_20CB4` | `sub_1FE54` / `sub_20CB4` 的投射物飛行路徑(會不會被地形擋)。 | `16-combat-turns-and-ai.md` |
| `sub_20E6C` | 餐點  /  `sub_210D8` → `sub_20E6C`  /  `Haggle(單價 × 活著的人數)`  /  存糧 += 活人數  / | `10-shop-prices-and-trade.md` |
| `sub_210D8` | 餐點  /  `sub_210D8` → `sub_20E6C`  /  `Haggle(單價 × 活著的人數)`  /  存糧 += 活人數  / | `10-shop-prices-and-trade.md` |
| `sub_21108` | ⚠ 酒是全遊戲**唯一不議價**的交易:`sub_21108` 直接拿 `dword_56E44[i]` 跟金幣比, | `10-shop-prices-and-trade.md` |
| `sub_21310` | 乾糧  /  `sub_21310`  /  `Haggle(單價 × 數量)`  /  存糧 += 數量  / | `10-shop-prices-and-trade.md` |
| `sub_21500` | 酒館的打聽消息**(`sub_21500`):`off_56E9C` 是 16 個四字母關鍵字 | `10-shop-prices-and-trade.md` |
| `sub_216C8` | 酒館  /  0x82  /  `sub_216C8`  /  9  /  ✅ 餐點 / 酒 / 乾糧(打聽消息待補)  / | `10-shop-prices-and-trade.md` |
| `sub_218DC` | 實際讀 `sub_218DC`(造船廠成交): | `11-map-objects.md` |
| `sub_219B0` | 買(`sub_11AF0`、`sub_11588`、`sub_112F8`、`sub_118CC`、`sub_219B0` 都一樣) | `10-shop-prices-and-trade.md` |
| `sub_21C40` | 造船廠  /  0x84  /  `sub_21C40` → `sub_219B0`  /  4  /  ✅ 報價收錢(地圖物件待補)  / | `10-shop-prices-and-trade.md` |
| `sub_21CE4` | 客滿判斷在 `sub_21CE4`:寄放的人數不能超過 `byte_57098[旅店]`(3/4/3/2/2/2 間房)。 | `10-shop-prices-and-trade.md` |
| `sub_21D48` | `R` Rest  /  `sub_21D48`  /  `Haggle(每天價 × 隊伍人數)`  /  移到床鋪,睡到早上六點  / | `10-shop-prices-and-trade.md` |
| `sub_22018` | `L` Leave  /  `sub_22018`  /  `Haggle(每天價)`,退房時結算  /  同伴離隊,記下寄放地點  / | `10-shop-prices-and-trade.md` |
| `sub_22280` | `P` Pick up  /  `sub_22280`  /  `Haggle(每天價) × 月數`  /  結清後歸隊  / | `10-shop-prices-and-trade.md` |
| `sub_2274C` | 旅店  /  0x88  /  `sub_2274C`  /  6  /  ✅ 住宿 / 寄放 / 領回  / | `10-shop-prices-and-trade.md` |
| `sub_24A50` | push off_41BE0 → call sub_24A50    ← 載 .16(off_41BE0 = 表的第 16 項 "CREATE.16") | `01-tileset-and-dot16-loader.md` |
| `sub_24BC0` | push off_41BA0 → call sub_24BC0    ← 載 PROPORT.PCS | `01-tileset-and-dot16-loader.md` |
| `sub_27A58` | call    sub_27A58               ; 等檔案就緒 | `11-map-objects.md` |
| `sub_27C98` | (位置 0,或前一個字元是空白)——`sub_1BD8C` → `sub_27C98`,再檢查 `byte_55F37[i]`。 | `06-conversation-script.md` |
| `sub_27D24` | 來源:FM Towns `WORRIORS.EXP` 的 `sub_27D24`(讀)與 `sub_284CC`(寫)。 | `07-save-format.md`, `10-shop-prices-and-trade.md`, `13-save-writing.md` |
| `sub_284CC` | 來源:FM Towns `WORRIORS.EXP` 的 `sub_27D24`(讀)與 `sub_284CC`(寫)。 | `07-save-format.md`, `13-save-writing.md` |
| `sub_28E14` | `sub_2B710(60)` → `sub_28E14(0, 60)`,而 `sub_28E14` 算範圍時是 | `15-combat-formulas.md`, `16-combat-turns-and-ai.md` |
| `sub_29304` | 每月 28 天、每年 13 個月**。一般行動每回合 **1 分鐘**(`sub_1DC8` → `sub_29304(1)`); | `04-npc-schedule-and-clock.md`, `07-save-format.md`, `10-shop-prices-and-trade.md`, `16-combat-turns-and-ai.md`, `17-magic.md` |
| `sub_29A64` | 0x01 在兩種單位上意義相反**,而 `sub_29A64` 就是靠這一點把兩邊算清楚: | `16-combat-turns-and-ai.md` |
| `sub_29EEC` | 0x8F  /  等待按鍵  /  `sub_29EEC`  / | `06-conversation-script.md` |
| `sub_2A610` | case 2: return (tile & 0xF0) == 0x60  /  /  sub_2A674(tile)  /  /  sub_2A610(mover, tile);  // 水陸兩棲 | `02-movement-and-tile-flags.md`, `03-scene-entry-and-tile-semantics.md` |
| `sub_2A674` | tile 0 為何被歸進「水」  /  `sub_2A674` 的 `tile < 4` 把 0 併進來。**視覺上 tile 0 根本不是水**(算繪出來是一團紅黃爆裂圖案,tile 1–3 才是藍色水面),所以這不是… | `02-movement-and-tile-flags.md` |
| `sub_2A694` | 通行判定第一參數  /  `sub_2A694(0, tile)`  /  `movzx eax, byte_3E08C`  /  照抄的話船、馬、飛毯全都照步行規則走  / | `02-movement-and-tile-flags.md`, `03-scene-entry-and-tile-semantics.md` |
| `sub_2B360` | obj  = sub_2B360(x+dx, y+dy, 樓層);      // 這一格有沒有 NPC / 物件 | `02-movement-and-tile-flags.md`, `03-scene-entry-and-tile-semantics.md`, `11-map-objects.md` |
| `sub_2B3DC` | 別的 NPC  /  離目標 **≥ 4 格**才算障礙  /  `sub_2B3DC`:有人就是不能走  / | `12-npc-movement.md` |
| `sub_2B710` | `sub_2B710(60)` → `sub_28E14(0, 60)`,而 `sub_28E14` 算範圍時是 | `15-combat-formulas.md` |
| `sub_2B724` | Mani  /  `sub_1CD3C`  /  回 **1..30**(與命中骰同一顆 `sub_2B724`),上限 MaxHP,死人無效  / | `15-combat-formulas.md`, `17-magic.md` |
| `sub_2BBB8` | 0x89 / 0x8A  /  **業報** +1 / −1(上限 99)  /  `sub_2BBB8(&byte_3E098, 1, 0x63)` / `sub_2BBFC`  / | `06-conversation-script.md` |
| `sub_2BBDC` | 金幣上限 **9999**(`sub_2BBDC(&gold, price, 0x270F)`)。 | `10-shop-prices-and-trade.md`, `16-combat-turns-and-ai.md` |
| `sub_2BBFC` | 0x89 / 0x8A  /  **業報** +1 / −1(上限 99)  /  `sub_2BBB8(&byte_3E098, 1, 0x63)` / `sub_2BBFC`  / | `06-conversation-script.md` |
| `sub_2C4F4` | } else { "Blocked!"; 嗶一聲 sub_2C4F4(165, 200); } | `03-scene-entry-and-tile-semantics.md` |
| `sub_2C740` | 2. 從 `sub_2C740` 與 `byte_54700` 的 xref 反追 `.TLK` 索引表語意與控制碼(`\x01` 疑為玩家名代入)。 | `00-hexrays-p3-verified.md`, `03-scene-entry-and-tile-semantics.md`, `04-npc-schedule-and-clock.md`, `11-map-objects.md`, `14-combat-maps.md` |
| `sub_2D0BC` | 移動成本分級  /  `"Slow progress!"` / `"Very slow!"` 在 `sub_2D0BC`,尚未讀  / | `02-movement-and-tile-flags.md` |
| `sub_2D564` | `sub_1DA10`(聖壇)、`sub_2D564`(cave/mine/dungeon)、`sub_10928`(印地名並確認)。 | `03-scene-entry-and-tile-semantics.md` |
| `sub_2D72C` | (其餘地點的編號要把 `sub_2D72C` 的每個 case 讀完才齊。) | `03-scene-entry-and-tile-semantics.md` |
| `sub_2E364` | 地點 > 0x7F      → 測 0x01   戰鬥中(sub_2E364 把 byte_3E0A3 設成 −1) | `16-combat-turns-and-ai.md`, `17-magic.md` |
| `sub_2E51C` | 格式看起來像「另一種入場方向」的位置,但 `sub_2E51C` 沒有搬它們 —— | `14-combat-maps.md` |
| `sub_2E58C` | 3. 進入戰鬥(`sub_2E58C`) | `14-combat-maps.md`, `16-combat-turns-and-ai.md` |
| `sub_2F0EC` | ├ sub_2F0EC  ★ 決定出現幾隻、哪幾種、站哪裡 | `16-combat-turns-and-ai.md` |
| `sub_2F2BC` | │       └ sub_2F2BC  角色的裝備防禦加總 | `15-combat-formulas.md` |
| `sub_3181C` | ⇒ **`sub_3181C(n)` = 播第 n 首 BGM**;`dword_65334` = 當前曲目、`dword_65338` = 前一首。 | `03-scene-entry-and-tile-semantics.md` |
| `sub_31CB8` | 原本以為 `sub_3181C` → `sub_31CB8` → `dword_65334` 這條鏈通往地點表。 | `03-scene-entry-and-tile-semantics.md` |
| `sub_377A4` | 44 次 `fread`(`sub_377A4`)+ 82 次 `fgetc`(`sub_3806C`),寫入端完全對稱。 | `07-save-format.md` |
| `sub_3806C` | 44 次 `fread`(`sub_377A4`)+ 82 次 `fgetc`(`sub_3806C`),寫入端完全對稱。 | `07-save-format.md` |
| `off_41054` | `off_41054[32]`  /  地點名稱指標  / | `03-scene-entry-and-tile-semantics.md` |
| `off_411BC` | `0x19`(25)  /  **the shrine of …**(八德聖壇;名字取自 `off_411BC[9]`,狀態看 `byte_411FC[8]`/`byte_41204[8]`)  /  `sub_1DA1… | `03-scene-entry-and-tile-semantics.md` |
| `off_4145C` | 店名 = off_4145C[店種][i]      店主 = off_4165C[店種][i] | `08-shops.md` |
| `off_4165C` | 店名 = off_4145C[店種][i]      店主 = off_4165C[店種][i] | `08-shops.md` |
| `off_41BA0` | ⚠ 這一步 **grep 反編譯輸出會回零命中** —— 存取是 `off_41BA0[edi*4]` 這種間接形式, | `01-tileset-and-dot16-loader.md` |
| `off_41BB4` | `off_41BB4[0..2]`(3 檔)  /  `0xD6D8` = 55,000 B  /  55,000  / | `01-tileset-and-dot16-loader.md` |
| `off_41BC0` | `off_41BC0[0..7]`(疑為 `MON0–7.16`)  /  `0x1068` = 4,200 B  /  4,200  / | `01-tileset-and-dot16-loader.md` |
| `off_41BE0` | push off_41BE0 → call sub_24A50    ← 載 .16(off_41BE0 = 表的第 16 項 "CREATE.16") | `01-tileset-and-dot16-loader.md` |
| `off_4FC44` | mov     eax, off_4FC44[eax*4]    ; 檔案 = {TOWNE,DWELLING,CASTLE,KEEP}[(編號-1)/8] | `03-scene-entry-and-tile-semantics.md` |
| `off_55E88` | `off_55E88` 是一張引擎自己的關鍵字表,**掃描順序在記錄的關鍵字之前**: | `06-conversation-script.md` |
| `off_56E9C` | 酒館的打聽消息**(`sub_21500`):`off_56E9C` 是 16 個四字母關鍵字 | `10-shop-prices-and-trade.md` |
| `off_56F04` | 搭配 `off_56F04`(人名)與 `off_56F88`(地名)回答「某某美德在哪座城」。 | `10-shop-prices-and-trade.md` |
| `off_56F88` | 搭配 `off_56F04`(人名)與 `off_56F88`(地名)回答「某某美德在哪座城」。 | `10-shop-prices-and-trade.md` |
| `byte_3DDB0` | byte_3DDB0(2)  byte_3DDB4(512 名冊)  word_3DFB4  word_3DFB6 | `13-save-writing.md` |
| `byte_3DDB4` | byte_3DDB0(2)  byte_3DDB4(512 名冊)  word_3DFB4  word_3DFB6 | `06-conversation-script.md`, `13-save-writing.md` |
| `byte_3DDB8` | cmp  byte_3DDB8[eax], 6Ah ; 'j' | `16-combat-turns-and-ai.md` |
| `byte_3DDBF` | byte_3DDBF[32*i] != 'P'  → 「汝無需此術」   ; 解毒要中毒 | `10-shop-prices-and-trade.md` |
| `byte_3DDC2` | movzx edx, byte_3DDC2[買家*32]         ; INT(角色紀錄 offset 0x0E) | `10-shop-prices-and-trade.md` |
| `byte_3DDC3` | 魔力消耗**(`sub byte_3DDC3[角色*32], 圈數`) | `17-magic.md` |
| `byte_3DDCA` | 最低等級**(`byte_3DDCA[角色*32] < 圈數` → 直接失敗) | `17-magic.md` |
| `byte_3DDCB` | 0x17 與 0x1F 的依據是位址:日期進位那段對 16 名角色跑 `byte_3DDCB[i*32]` | `09-items-and-creatures.md` |
| `byte_3DDCC` | `sub_B274` 對角色讀 `byte_3DDCC[角色*32]`,而 `0x3DDCC − 0x3DDB4 = 0x18`。 | `15-combat-formulas.md` |
| `byte_3DDD3` | 而 `0x3DDCB − 0x3DDB4 = 0x17`;入隊時寫 `byte_3DDD3[i*32]` 而差是 `0x1F`。 | `09-items-and-creatures.md` |
| `byte_3DFB8` | 而全域變數之間有對齊留下的空隙:`byte_3DFB8` 讀完 10 B 到 `0x3DFC2`, | `10-shop-prices-and-trade.md` |
| `byte_3DFB9` | `0x0207`  /  `byte_3DFB9`  /  寶石  / | `10-shop-prices-and-trade.md` |
| `byte_3DFBA` | `0x0208`  /  `byte_3DFBA`  /  火把  / | `10-shop-prices-and-trade.md` |
| `byte_3DFBC` | 上船時原本騎的東西會一起帶上:魔毯記進 `byte_3DFBC`、小艇讓船上的小艇數 +1 | `11-map-objects.md` |
| `byte_3DFC0` | 地點 == 0x12 且 byte_3DFC0 == 0 → 「Absorbed!」 | `17-magic.md` |
| `byte_3DFC4` | byte_3DFC4(4)  byte_3DFD0(48 裝備)   byte_3E000(48) | `13-save-writing.md` |
| `byte_3DFD0` | `0x021A`  /  `byte_3DFD0`  /  裝備持有數 48 B,索引 = 裝備編號  / | `10-shop-prices-and-trade.md`, `13-save-writing.md`, `17-magic.md` |
| `byte_3E000` | 3. ★ byte_3E000[咒語] −−                          ← 從這裡開始都不退 | `13-save-writing.md`, `17-magic.md` |
| `byte_3E030` | byte_3E030(8)  byte_3E038(8)  byte_3E040(8)  byte_3E048(8) | `13-save-writing.md` |
| `byte_3E038` | byte_3E030(8)  byte_3E038(8)  byte_3E040(8)  byte_3E048(8) | `13-save-writing.md` |
| `byte_3E040` | byte_3E030(8)  byte_3E038(8)  byte_3E040(8)  byte_3E048(8) | `13-save-writing.md` |
| `byte_3E048` | byte_3E030(8)  byte_3E038(8)  byte_3E040(8)  byte_3E048(8) | `13-save-writing.md` |
| `byte_3E050` | byte_3E050(8)  byte_3E058(8)  byte_3E060(8 藥草)  … | `13-save-writing.md` |
| `byte_3E058` | byte_3E050(8)  byte_3E058(8)  byte_3E060(8 藥草)  … | `13-save-writing.md` |
| `byte_3E060` | `byte_3E060`(藥草)對應 `0x02AA`,兩者的差都是 `0x3DDB6` —— 一致, | `10-shop-prices-and-trade.md`, `13-save-writing.md`, `17-magic.md` |
| `byte_3E06B` | 0x02B5  /  隊伍人數  /  3(`sub_1BB5C` 用 `cmp byte_3E06B, 6` 判滿員)  / | `07-save-format.md` |
| `byte_3E08A` | 全域 byte_3E08A == 'T'                  → 整場不動(時間停止 An Tym,10 回合) | `04-npc-schedule-and-clock.md`, `16-combat-turns-and-ai.md`, `17-magic.md` |
| `byte_3E08C` | 通行判定第一參數  /  `sub_2A694(0, tile)`  /  `movzx eax, byte_3E08C`  /  照抄的話船、馬、飛毯全都照步行規則走  / | `02-movement-and-tile-flags.md`, `03-scene-entry-and-tile-semantics.md`, `08-shops.md`, `10-shop-prices-and-trade.md`, `11-map-objects.md` |
| `byte_3E08D` | byte_3E08D 月   > 13  → 設回 1 並進位 | `04-npc-schedule-and-clock.md` |
| `byte_3E08E` | byte_3E08E 日   > 28  → 設回 1 並進位 | `04-npc-schedule-and-clock.md` |
| `byte_3E08F` | `@` 0x40  /  `byte_3E08F`  /  時段:< 12 morning、< 18 afternoon、其餘 evening  / | `04-npc-schedule-and-clock.md`, `10-shop-prices-and-trade.md` |
| `byte_3E091` | byte_3E091 分   += minutes;  > 59 → 減 60 並進位 | `04-npc-schedule-and-clock.md` |
| `byte_3E092` | 每 10 個單位行動 = 遊戲內 1 分鐘**(`byte_3E092` 數到 10 → `sub_29304(1)`)。 | `16-combat-turns-and-ai.md` |
| `byte_3E098` | 0x89 / 0x8A  /  **業報** +1 / −1(上限 99)  /  `sub_2BBB8(&byte_3E098, 1, 0x63)` / `sub_2BBFC`  / | `06-conversation-script.md` |
| `byte_3E09E` | An Tym  /  `sub_198E0`  /  `byte_3E08A = 'T'`、`byte_3E09E = 10`  / | `17-magic.md` |
| `byte_3E09F` | 判斷依據是 cdecl 的壓棧順序:`push esi(武器); push -byte_3E09F; | `15-combat-formulas.md` |
| `byte_3E0A0` | 傷害減半(0x0020)      → 除以 2(除非 byte_3E0A0 成立,那個旗標還沒解) | `16-combat-turns-and-ai.md` |
| `byte_3E0A3` | esi = byte_3E0A3 >> 3            ; 檔案 = {TOWNE,DWELLING,CASTLE,KEEP}.NPC[(編號-1)/8] | `03-scene-entry-and-tile-semantics.md`, `04-npc-schedule-and-clock.md`, `10-shop-prices-and-trade.md`, `17-magic.md` |
| `byte_3E0A5` | 樓層增減(`sub_758`)  /  `byte_3E0A5 = 1` / `= -1`  /  `inc` / `dec byte_3E0A5`  /  照抄的話只能在 1F 與 B1F 之間跳  / | `03-scene-entry-and-tile-semantics.md`, `11-map-objects.md` |
| `byte_3E0A6` | 邊界旗標  /  每個方向一個常數(西/北 = 1,東/南 = 0)  /  `cmp byte_3E0A6, 1 / jnb` 等四組比較  /  照抄的話往東往南永遠出不了城  / | `03-scene-entry-and-tile-semantics.md`, `10-shop-prices-and-trade.md` |
| `byte_3E0A7` | byte_3E0A7 = 30;      // 場景內 Y(靠底部 → 城鎮南方入口) | `03-scene-entry-and-tile-semantics.md` |
| `byte_3E0AD` | 三個攻擊咒語的傷害也還沒逆到底(`sub_189E4` 只把攻擊碼寫進 `byte_3E0AD`, | `16-combat-turns-and-ai.md`, `17-magic.md` |
| `byte_3E0B2` | 且目標是隊員時直接 `byte_3E0B2 = 0x20` 走人,完全擋下。 | `16-combat-turns-and-ai.md` |
| `byte_3E0B6` | `byte_3E0B6`(咒語)與 `byte_3E0B7`(火把)各自是一個分鐘倒數, | `17-magic.md` |
| `byte_3E0B7` | `byte_3E0B6`(咒語)與 `byte_3E0B7`(火把)各自是一個分鐘倒數, | `17-magic.md` |
| `byte_3E0B8` | byte_3E0B8[施法者] 記著「上一個攻擊我的是誰」 | `17-magic.md` |
| `byte_3E165` | mov   byte_3E165, dl | `11-map-objects.md` |
| `byte_3E166` | mov   byte_3E166, al | `11-map-objects.md` |
| `byte_3E570` | sub_2C740(file, edi,       0x200, byte_3E570)   ; 512 B  32 × 16 B 排程 | `04-npc-schedule-and-clock.md` |
| `byte_3E970` | `byte_3E970[npc*32]` —— 路徑本身,**(步數, 方向) 成對**,共 16 段 | `12-npc-movement.md` |
| `byte_3EDB0` | sub_2C740(file, edi+0x200, 0x20,  byte_3EDB0)   ;  32 B  每個 NPC 的生物編號 | `04-npc-schedule-and-clock.md`, `07-save-format.md`, `12-npc-movement.md` |
| `byte_3EE17` | `byte_3EE17` 的旗標(帆船 0x82、小艇 0x40)加上兩個座標 —— 船停在碼頭等你, | `11-map-objects.md` |
| `byte_3EE18` | 結尾多一個 2 B 欄位 `byte_3EE18` | `07-save-format.md` |
| `byte_3F050` | `sub_B398` 印證了這件事:它取 `byte_3F050[生物*8]` 當「力量那一項」、 | `15-combat-formulas.md` |
| `byte_3F052` | `byte_3F052[生物*8]` 當「智力那一項」,而角色走的是紀錄的 0x0C 與 0x0E | `15-combat-formulas.md` |
| `byte_3F053` | +3  /  護甲(減傷)  /  `sub_B274` 讀 `byte_3F053[生物*8]`  / | `15-combat-formulas.md` |
| `byte_3F054` | +4  /  攻擊力  /  `sub_B274` 讀 `byte_3F054[生物*8]`  / | `15-combat-formulas.md` |
| `byte_3F055` | movzx edx, byte_3F055[eax*8]   ; 生命上限 | `16-combat-turns-and-ai.md` |
| `byte_3F2F8` | ├ sub_1FA6C   瞄準(射程來自 byte_3F2F8[武器]) | `15-combat-formulas.md` |
| `byte_3F398` | 203  /  6×32 + 11  /  敵人入場 X ×16  /  `byte_3F398`  / | `14-combat-maps.md` |
| `byte_3F3A8` | 235  /  7×32 + 11  /  敵人入場 Y ×16  /  `byte_3F3A8`  / | `14-combat-maps.md` |
| `byte_3F3B8` | 107  /  3×32 + 11  /  隊員入場 X ×6  /  `byte_3F3B8`  / | `14-combat-maps.md` |
| `byte_3F6E4` | `memset(byte_3F6E4, 0xFF, 0x160)` —— 0x160 = 352 = **11 列 × 32 stride** | `03-scene-entry-and-tile-semantics.md` |
| `byte_3F789` | `byte_3F789` 看起來像一個單獨的 byte,但它被 `[32*dy + dx]` 這樣索引(dy,dx ∈ −1..1), | `02-movement-and-tile-flags.md`, `03-scene-entry-and-tile-semantics.md` |
| `byte_3F8F4` | push offset byte_3F8F4 | `14-combat-maps.md` |
| `byte_400F4` | if 模式 < 0 且 byte_400F4[y*32+x] == 0xC8/0xC9(樓梯) → grid = 5 | `03-scene-entry-and-tile-semantics.md`, `12-npc-movement.md` |
| `byte_404F3` | 座標越界           → 用 byte_404F3 當 tile(那是地圖最後一格,原版的 quirk) | `12-npc-movement.md` |
| `byte_41033` | dump `byte_41033[1..32]` 得到起始索引,再用「同檔下一個地點的起始索引 − 自己」算出層數: | `03-scene-entry-and-tile-semantics.md` |
| `byte_410F3` | mov     cl, byte_410F3[edx]        ; 世界座標**從地點表讀回來**(1-based 索引的同一張表) | `03-scene-entry-and-tile-semantics.md` |
| `byte_410F4` | for (i = 0; i < 32 && (byte_410F4[i] != 0  /  /  byte_4111C[i] != 0); ++i); | `03-scene-entry-and-tile-semantics.md` |
| `byte_4111B` | mov     dl, byte_4111B[edx] | `03-scene-entry-and-tile-semantics.md` |
| `byte_4111C` | for (i = 0; i < 32 && (byte_410F4[i] != 0  /  /  byte_4111C[i] != 0); ++i); | `03-scene-entry-and-tile-semantics.md` |
| `byte_411FC` | `0x19`(25)  /  **the shrine of …**(八德聖壇;名字取自 `off_411BC[9]`,狀態看 `byte_411FC[8]`/`byte_41204[8]`)  /  `sub_1DA1… | `03-scene-entry-and-tile-semantics.md` |
| `byte_41204` | `0x19`(25)  /  **the shrine of …**(八德聖壇;名字取自 `off_411BC[9]`,狀態看 `byte_411FC[8]`/`byte_41204[8]`)  /  `sub_1DA1… | `03-scene-entry-and-tile-semantics.md` |
| `byte_4185C` | loop:                        ; 在 byte_4185C[店種][0..15] 裡找當前地點 | `08-shops.md`, `10-shop-prices-and-trade.md` |
| `byte_41C18` | 2. 或查 `byte_41C18` 的 xref(`sub_6730` 開頭有 `for(i=0;i<256;i++) byte_41C18[i]=i`, | `01-tileset-and-dot16-loader.md` |
| `byte_54524` | ⚠⚠ **`byte_54524` 不是玩家那張通行表。** 玩家走 `byte_5FF6C`,NPC 走這張, | `12-npc-movement.md` |
| `byte_54700` | 2. 從 `sub_2C740` 與 `byte_54700` 的 xref 反追 `.TLK` 索引表語意與控制碼(`\x01` 疑為玩家名代入)。 | `00-hexrays-p3-verified.md` |
| `byte_55384` | ⚠ Hex-Rays 把 `word_3EF34` 整個折掉了:`byte_55384[edi + eax*8]` 被印成 | `10-shop-prices-and-trade.md` |
| `byte_55F18` | 0x85 / 0x86 / 0x8C / 0xFE  /  切到子模式  /  `byte_55F18 = 碼`;函式開頭會把後續位元組改送 `sub_1C1E8`  / | `06-conversation-script.md` |
| `byte_55F1A` | 0x8E  /  切換強調  /  `byte_55F1A ^= 0x80`,影響後續字元的輸出屬性  / | `06-conversation-script.md` |
| `byte_55F32` | 0x91–0x9F  /  反問玩家並讀取回答(15 種)  /  `byte_55F32 = 碼; sub_1C0AC` → 印 "You respond-\n:" 後收輸入  / | `06-conversation-script.md` |
| `byte_55F37` | (位置 0,或前一個字元是空白)——`sub_1BD8C` → `sub_27C98`,再檢查 `byte_55F37[i]`。 | `06-conversation-script.md` |
| `byte_55F4A` | else byte_55F4A = 1                            ; 設下 pendingSpace | `05-text-compression.md` |
| `byte_57034` | `byte_57034[酒館]` 選出一套「菜單樣式」(0..3),樣式再決定四個字母。 | `10-shop-prices-and-trade.md` |
| `byte_57080` | mov   dl, byte_57080[eax]     ; 停泊 X | `11-map-objects.md` |
| `byte_57084` | mov   al, byte_57084[eax]     ; 停泊 Y | `11-map-objects.md` |
| `byte_57090` | `byte_57090[旅店]`(每人每天 2 或 3 金)上: | `10-shop-prices-and-trade.md` |
| `byte_57098` | 客滿判斷在 `sub_21CE4`:寄放的人數不能超過 `byte_57098[旅店]`(3/4/3/2/2/2 間房)。 | `10-shop-prices-and-trade.md` |
| `byte_5FF6C` | BOOL ok = ((128 >> (tile & 7)) & byte_5FF6C[tile >> 3]) == 0;   // bit = 1 → 阻擋 | `02-movement-and-tile-flags.md`, `12-npc-movement.md` |
| `byte_5FF8C` | switch (byte_5FF8C[mover >> 2]) {          // 移動者 → 移動模式(0–10) | `02-movement-and-tile-flags.md` |
| `byte_5FFA8` | case 5: /* 方向性通行:用 byte_5FFA8[tile] / byte_5FF6C[tile] 的方向 bit */ | `02-movement-and-tile-flags.md` |
| `word_3DFB4` | byte_3DDB0(2)  byte_3DDB4(512 名冊)  word_3DFB4  word_3DFB6 | `10-shop-prices-and-trade.md`, `13-save-writing.md` |
| `word_3DFB6` | byte_3DDB0(2)  byte_3DDB4(512 名冊)  word_3DFB4  word_3DFB6 | `10-shop-prices-and-trade.md`, `11-map-objects.md`, `13-save-writing.md` |
| `word_3E084` | word_3E084 年 | `04-npc-schedule-and-clock.md` |
| `word_3E086` | `sub_15DD4` 每次行動後重數兩邊。⚠ 它用的兩個全域 `word_3E086` / `word_3E088` | `16-combat-turns-and-ai.md` |
| `word_3E088` | `sub_15DD4` 每次行動後重數兩邊。⚠ 它用的兩個全域 `word_3E086` / `word_3E088` | `16-combat-turns-and-ai.md` |
| `word_3E770` | 1. 執行期狀態(`word_3E770`,32 × 16 B) | `12-npc-movement.md` |
| `word_3E77A` | for i in 0..31: word_3E77A[i*16] = 區域緩衝[i]   ; 對話號碼搬進執行期記錄 | `04-npc-schedule-and-clock.md` |
| `word_3ED70` | `word_3ED70[npc]` —— 路徑索引;`0xFFFF` = 目前沒有路徑 | `12-npc-movement.md` |
| `word_3EDD4` | if word_3EDD4[npc] != 0 且 random(0,2) != 1 → 這回合不重算 | `12-npc-movement.md` |
| `word_3EF34` | ⚠ Hex-Rays 把 `word_3EF34` 整個折掉了:`byte_55384[edi + eax*8]` 被印成 | `10-shop-prices-and-trade.md`, `11-map-objects.md` |
| `word_3EF36` | mov  word_3EF36, ax          ; 店種 = 對話碼 − 0x81 | `10-shop-prices-and-trade.md` |
| `word_3EF38` | `%` 0x25  /  `word_3EF38`  /  價格  / | `10-shop-prices-and-trade.md`, `11-map-objects.md` |
| `word_3EF3A` | `^` 0x5E  /  `word_3EF3A`  /  數量(藥草一次幾份)  / | `10-shop-prices-and-trade.md` |
| `dword_3E46C` | ≥ 0x40  /  怪物  /  `cmp byte ptr dword_3E46C[eax*8], 40h`,與生物名表的 `CreatureBase` 同源  / | `10-shop-prices-and-trade.md`, `11-map-objects.md` |
| `dword_3EF24` | `#` 0x23  /  `dword_3EF24`  /  店名  / | `10-shop-prices-and-trade.md` |
| `dword_3EF28` | `$` 0x24  /  `dword_3EF28`  /  店主  / | `10-shop-prices-and-trade.md` |
| `dword_3EF2C` | `&` 0x26  /  `dword_3EF2C`  /  物品名(收購對白;有別稱就用別稱)  / | `10-shop-prices-and-trade.md` |
| `dword_3EF30` | `*` 0x2A  /  `dword_3EF30`  /  地名(酒館八卦)  / | `10-shop-prices-and-trade.md` |
| `dword_3EF4C` | 在一塊 32×32 的 buffer(`dword_3EF4C` 指向)裡標出每一格的狀態: | `12-npc-movement.md` |
| `dword_3EF50` | `dword_3EF50`,**32 槽 × 8 B**。0..5 是隊員、6..31 是敵人。 | `16-combat-turns-and-ai.md` |
| `dword_41794` | 它對 `b >= 0x80` 查 `dword_41794[b*4]`,而 `0x41794 + 0x80*4 = 0x41994`, | `05-text-compression.md`, `08-shops.md` |
| `dword_41990` | 來源:FM Towns `WORRIORS.EXP` 的 `sub_1C3F8`(展開)與 `dword_41990`(槽表); | `05-text-compression.md`, `08-shops.md` |
| `dword_41D28` | │    └─ 啟動初始化:載 towns.fnt(17,280 B → dword_41D28)與 u5.fnt(0x4000 B → dword_4FFB8); | `01-tileset-and-dot16-loader.md` |
| `dword_4FFB8` | `u5.fnt`  /  **0x4000 = 16,384 B**  /  `dword_4FFB8`  /  `ULTIMA FONT DATA READ FAIL !!`  / | `01-tileset-and-dot16-loader.md` |
| `dword_552C4` | mov  ax, word ptr dword_552C4[edi*4]  ; base | `10-shop-prices-and-trade.md` |
| `dword_553CC` | `dword_553CC[店種][4]` 是四句問候語在 `SHOPPE.DAT` 裡的**位元組位移**, | `08-shops.md` |
| `dword_555E8` | (dword_555E8 = {0, 0, 1, -1}、dword_555F8 = {1, -1, 0, 0}) | `11-map-objects.md` |
| `dword_555F8` | (dword_555E8 = {0, 0, 1, -1}、dword_555F8 = {1, -1, 0, 0}) | `11-map-objects.md` |
| `dword_55714` | 唯一的例外是 `dword_55714`(裝備的「另一種說法」,`Cloth suit`、`Two-Handed Axe` | `10-shop-prices-and-trade.md` |
| `dword_55F14` | 原版 0x87 的作法是把文字指標存起來、往下讀一則再還原(`dword_55F14` 的存取還原)。 | `06-conversation-script.md` |
| `dword_56E44` | ⚠ 酒是全遊戲**唯一不議價**的交易:`sub_21108` 直接拿 `dword_56E44[i]` 跟金幣比, | `10-shop-prices-and-trade.md` |
| `dword_5AC30` | 1. 查 `dword_5AC30`(handle 表)的 xref → 找取用 handle 的函式 → 那裡才有解壓。 | `01-tileset-and-dot16-loader.md` |
| `dword_5FFF4` | 為什麼會讀錯**:`"BGM SONG %d"` 那段被 `dword_5FFF4 == 1` 的 debug 分支包著, | `03-scene-entry-and-tile-semantics.md` |
| `dword_65334` | ⇒ **`sub_3181C(n)` = 播第 n 首 BGM**;`dword_65334` = 當前曲目、`dword_65338` = 前一首。 | `03-scene-entry-and-tile-semantics.md` |
| `dword_65338` | ⇒ **`sub_3181C(n)` = 播第 n 首 BGM**;`dword_65334` = 當前曲目、`dword_65338` = 前一首。 | `03-scene-entry-and-tile-semantics.md` |
| `loc_630` | jbe     short loc_630 | `03-scene-entry-and-tile-semantics.md` |
| `loc_17A5` | jmp  short loc_17A5 | `11-map-objects.md` |
| `loc_17B5` | jge  short loc_17B5 | `11-map-objects.md` |
| `loc_9E4D` | jz   short loc_9E4D | `16-combat-turns-and-ai.md` |
| `loc_9F00` | jge  loc_9F00              ; ← random(0,255) >= 128 就整個不射 | `16-combat-turns-and-ai.md` |
| `loc_AC0C` | jnz  loc_AC0C            ; 還沒輪到 | `16-combat-turns-and-ai.md` |
| `loc_29AA6` | jnz  short loc_29AA6 | `16-combat-turns-and-ai.md` |
| `loc_3185B` | jnz     short loc_3185B | `03-scene-entry-and-tile-semantics.md` |
| `loc_3197C` | loc_3197C: | `03-scene-entry-and-tile-semantics.md` |
| `loc_31CC5` | loc_31CC5:  mov     eax, dword_65334      ; return dword_65334 | `03-scene-entry-and-tile-semantics.md` |
