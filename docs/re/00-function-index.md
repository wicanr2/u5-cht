# RE-00b:函式與全域索引(自動產生)

> `python3 tools/gen_func_index.py > docs/re/00-function-index.md` 重新產生。
> **讀任何 `sub_XXXX` 之前先查這裡** —— 筆記超過二三十份後,憑記憶一定會重讀已解過的東西。
>
> 目前收錄 **58** 個符號,來源是 `docs/` 下的逆向筆記。

| 符號 | 已知語意(取自筆記) | 出處 |
|---|---|---|
| `sub_5C8` | `sub_5C8` 就是載入場景地圖的函式。組語三行講完: | `03-scene-entry-and-tile-semantics.md` |
| `sub_758` | 樓層增減(`sub_758`)  /  `byte_3E0A5 = 1` / `= -1`  /  `inc` / `dec byte_3E0A5`  /  照抄的話只能在 1F 與 B1F 之間跳  / | `03-scene-entry-and-tile-semantics.md` |
| `sub_86C` | 方向樓梯  /  0xC4–0xC7(低 2 bit = 朝向)  /  **走進去**:同向上樓、反向(`facing ^ 2`)下樓  /  `sub_86C` → `sub_758`  / | `02-movement-and-tile-flags.md`, `03-scene-entry-and-tile-semantics.md` |
| `sub_EA0` | 梯子  /  0xC8 上 / 0xC9 下 / 0x86 活板門(下)  /  站在上面按 **K**(Klimb)  /  `sub_EA0` → `sub_758(0 或 2, 196)`  / | `03-scene-entry-and-tile-semantics.md` |
| `sub_6730` | 2. 或查 `byte_41C18` 的 xref(`sub_6730` 開頭有 `for(i=0;i<256;i++) byte_41C18[i]=i`, | `01-tileset-and-dot16-loader.md` |
| `sub_C778` | 寫  0xC790   sub_C778    mov dword_65334, 1 | `03-scene-entry-and-tile-semantics.md` |
| `sub_10928` | `sub_1DA10`(聖壇)、`sub_2D564`(cave/mine/dungeon)、`sub_10928`(印地名並確認)。 | `03-scene-entry-and-tile-semantics.md` |
| `sub_1DA10` | `0x19`(25)  /  **the shrine of …**(八德聖壇;名字取自 `off_411BC[9]`,狀態看 `byte_411FC[8]`/`byte_41204[8]`)  /  `sub_1DA1… | `03-scene-entry-and-tile-semantics.md` |
| `sub_24A50` | push off_41BE0 → call sub_24A50    ← 載 .16(off_41BE0 = 表的第 16 項 "CREATE.16") | `01-tileset-and-dot16-loader.md` |
| `sub_24BC0` | push off_41BA0 → call sub_24BC0    ← 載 PROPORT.PCS | `01-tileset-and-dot16-loader.md` |
| `sub_2A610` | case 2: return (tile & 0xF0) == 0x60  /  /  sub_2A674(tile)  /  /  sub_2A610(mover, tile);  // 水陸兩棲 | `02-movement-and-tile-flags.md`, `03-scene-entry-and-tile-semantics.md` |
| `sub_2A674` | tile 0 為何被歸進「水」  /  `sub_2A674` 的 `tile < 4` 把 0 併進來。**視覺上 tile 0 根本不是水**(算繪出來是一團紅黃爆裂圖案,tile 1–3 才是藍色水面),所以這不是… | `02-movement-and-tile-flags.md` |
| `sub_2A694` | 通行判定第一參數  /  `sub_2A694(0, tile)`  /  `movzx eax, byte_3E08C`  /  照抄的話船、馬、飛毯全都照步行規則走  / | `02-movement-and-tile-flags.md`, `03-scene-entry-and-tile-semantics.md` |
| `sub_2B360` | obj  = sub_2B360(x+dx, y+dy, 樓層);      // 這一格有沒有 NPC / 物件 | `02-movement-and-tile-flags.md`, `03-scene-entry-and-tile-semantics.md` |
| `sub_2C4F4` | } else { "Blocked!"; 嗶一聲 sub_2C4F4(165, 200); } | `03-scene-entry-and-tile-semantics.md` |
| `sub_2C740` | 2. 從 `sub_2C740` 與 `byte_54700` 的 xref 反追 `.TLK` 索引表語意與控制碼(`\x01` 疑為玩家名代入)。 | `00-hexrays-p3-verified.md`, `03-scene-entry-and-tile-semantics.md` |
| `sub_2D0BC` | 移動成本分級  /  `"Slow progress!"` / `"Very slow!"` 在 `sub_2D0BC`,尚未讀  / | `02-movement-and-tile-flags.md` |
| `sub_2D564` | `sub_1DA10`(聖壇)、`sub_2D564`(cave/mine/dungeon)、`sub_10928`(印地名並確認)。 | `03-scene-entry-and-tile-semantics.md` |
| `sub_2D72C` | (其餘地點的編號要把 `sub_2D72C` 的每個 case 讀完才齊。) | `03-scene-entry-and-tile-semantics.md` |
| `sub_3181C` | ⇒ **`sub_3181C(n)` = 播第 n 首 BGM**;`dword_65334` = 當前曲目、`dword_65338` = 前一首。 | `03-scene-entry-and-tile-semantics.md` |
| `sub_31CB8` | 原本以為 `sub_3181C` → `sub_31CB8` → `dword_65334` 這條鏈通往地點表。 | `03-scene-entry-and-tile-semantics.md` |
| `off_41054` | `off_41054[32]`  /  地點名稱指標  / | `03-scene-entry-and-tile-semantics.md` |
| `off_411BC` | `0x19`(25)  /  **the shrine of …**(八德聖壇;名字取自 `off_411BC[9]`,狀態看 `byte_411FC[8]`/`byte_41204[8]`)  /  `sub_1DA1… | `03-scene-entry-and-tile-semantics.md` |
| `off_41BA0` | ⚠ 這一步 **grep 反編譯輸出會回零命中** —— 存取是 `off_41BA0[edi*4]` 這種間接形式, | `01-tileset-and-dot16-loader.md` |
| `off_41BB4` | `off_41BB4[0..2]`(3 檔)  /  `0xD6D8` = 55,000 B  /  55,000  / | `01-tileset-and-dot16-loader.md` |
| `off_41BC0` | `off_41BC0[0..7]`(疑為 `MON0–7.16`)  /  `0x1068` = 4,200 B  /  4,200  / | `01-tileset-and-dot16-loader.md` |
| `off_41BE0` | push off_41BE0 → call sub_24A50    ← 載 .16(off_41BE0 = 表的第 16 項 "CREATE.16") | `01-tileset-and-dot16-loader.md` |
| `off_4FC44` | mov     eax, off_4FC44[eax*4]    ; 檔案 = {TOWNE,DWELLING,CASTLE,KEEP}[(編號-1)/8] | `03-scene-entry-and-tile-semantics.md` |
| `byte_3E08C` | 通行判定第一參數  /  `sub_2A694(0, tile)`  /  `movzx eax, byte_3E08C`  /  照抄的話船、馬、飛毯全都照步行規則走  / | `02-movement-and-tile-flags.md`, `03-scene-entry-and-tile-semantics.md` |
| `byte_3E0A3` | cmp     byte_3E0A3, 19h            ; 地點 25(ARARAT)特判 | `03-scene-entry-and-tile-semantics.md` |
| `byte_3E0A5` | 樓層增減(`sub_758`)  /  `byte_3E0A5 = 1` / `= -1`  /  `inc` / `dec byte_3E0A5`  /  照抄的話只能在 1F 與 B1F 之間跳  / | `03-scene-entry-and-tile-semantics.md` |
| `byte_3E0A6` | 邊界旗標  /  每個方向一個常數(西/北 = 1,東/南 = 0)  /  `cmp byte_3E0A6, 1 / jnb` 等四組比較  /  照抄的話往東往南永遠出不了城  / | `03-scene-entry-and-tile-semantics.md` |
| `byte_3E0A7` | byte_3E0A7 = 30;      // 場景內 Y(靠底部 → 城鎮南方入口) | `03-scene-entry-and-tile-semantics.md` |
| `byte_3F6E4` | `memset(byte_3F6E4, 0xFF, 0x160)` —— 0x160 = 352 = **11 列 × 32 stride** | `03-scene-entry-and-tile-semantics.md` |
| `byte_3F789` | `byte_3F789` 看起來像一個單獨的 byte,但它被 `[32*dy + dx]` 這樣索引(dy,dx ∈ −1..1), | `02-movement-and-tile-flags.md`, `03-scene-entry-and-tile-semantics.md` |
| `byte_400F4` | push    offset byte_400F4 | `03-scene-entry-and-tile-semantics.md` |
| `byte_41033` | dump `byte_41033[1..32]` 得到起始索引,再用「同檔下一個地點的起始索引 − 自己」算出層數: | `03-scene-entry-and-tile-semantics.md` |
| `byte_410F3` | mov     cl, byte_410F3[edx]        ; 世界座標**從地點表讀回來**(1-based 索引的同一張表) | `03-scene-entry-and-tile-semantics.md` |
| `byte_410F4` | for (i = 0; i < 32 && (byte_410F4[i] != 0  /  /  byte_4111C[i] != 0); ++i); | `03-scene-entry-and-tile-semantics.md` |
| `byte_4111B` | mov     dl, byte_4111B[edx] | `03-scene-entry-and-tile-semantics.md` |
| `byte_4111C` | for (i = 0; i < 32 && (byte_410F4[i] != 0  /  /  byte_4111C[i] != 0); ++i); | `03-scene-entry-and-tile-semantics.md` |
| `byte_411FC` | `0x19`(25)  /  **the shrine of …**(八德聖壇;名字取自 `off_411BC[9]`,狀態看 `byte_411FC[8]`/`byte_41204[8]`)  /  `sub_1DA1… | `03-scene-entry-and-tile-semantics.md` |
| `byte_41204` | `0x19`(25)  /  **the shrine of …**(八德聖壇;名字取自 `off_411BC[9]`,狀態看 `byte_411FC[8]`/`byte_41204[8]`)  /  `sub_1DA1… | `03-scene-entry-and-tile-semantics.md` |
| `byte_41C18` | 2. 或查 `byte_41C18` 的 xref(`sub_6730` 開頭有 `for(i=0;i<256;i++) byte_41C18[i]=i`, | `01-tileset-and-dot16-loader.md` |
| `byte_54700` | 2. 從 `sub_2C740` 與 `byte_54700` 的 xref 反追 `.TLK` 索引表語意與控制碼(`\x01` 疑為玩家名代入)。 | `00-hexrays-p3-verified.md` |
| `byte_5FF6C` | BOOL ok = ((128 >> (tile & 7)) & byte_5FF6C[tile >> 3]) == 0;   // bit = 1 → 阻擋 | `02-movement-and-tile-flags.md` |
| `byte_5FF8C` | switch (byte_5FF8C[mover >> 2]) {          // 移動者 → 移動模式(0–10) | `02-movement-and-tile-flags.md` |
| `byte_5FFA8` | case 5: /* 方向性通行:用 byte_5FFA8[tile] / byte_5FF6C[tile] 的方向 bit */ | `02-movement-and-tile-flags.md` |
| `dword_41D28` | │    └─ 啟動初始化:載 towns.fnt(17,280 B → dword_41D28)與 u5.fnt(0x4000 B → dword_4FFB8); | `01-tileset-and-dot16-loader.md` |
| `dword_4FFB8` | `u5.fnt`  /  **0x4000 = 16,384 B**  /  `dword_4FFB8`  /  `ULTIMA FONT DATA READ FAIL !!`  / | `01-tileset-and-dot16-loader.md` |
| `dword_5AC30` | 1. 查 `dword_5AC30`(handle 表)的 xref → 找取用 handle 的函式 → 那裡才有解壓。 | `01-tileset-and-dot16-loader.md` |
| `dword_5FFF4` | 為什麼會讀錯**:`"BGM SONG %d"` 那段被 `dword_5FFF4 == 1` 的 debug 分支包著, | `03-scene-entry-and-tile-semantics.md` |
| `dword_65334` | ⇒ **`sub_3181C(n)` = 播第 n 首 BGM**;`dword_65334` = 當前曲目、`dword_65338` = 前一首。 | `03-scene-entry-and-tile-semantics.md` |
| `dword_65338` | ⇒ **`sub_3181C(n)` = 播第 n 首 BGM**;`dword_65334` = 當前曲目、`dword_65338` = 前一首。 | `03-scene-entry-and-tile-semantics.md` |
| `loc_630` | jbe     short loc_630 | `03-scene-entry-and-tile-semantics.md` |
| `loc_3185B` | jnz     short loc_3185B | `03-scene-entry-and-tile-semantics.md` |
| `loc_3197C` | loc_3197C: | `03-scene-entry-and-tile-semantics.md` |
| `loc_31CC5` | loc_31CC5:  mov     eax, dword_65334      ; return dword_65334 | `03-scene-entry-and-tile-semantics.md` |
