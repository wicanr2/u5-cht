# RE-00b:函式與全域索引(自動產生)

> `python3 tools/gen_func_index.py > docs/re/00-function-index.md` 重新產生。
> **讀任何 `sub_XXXX` 之前先查這裡** —— 筆記超過二三十份後,憑記憶一定會重讀已解過的東西。
>
> 目前收錄 **51** 個符號,來源是 `docs/` 下的逆向筆記。

| 符號 | 已知語意(取自筆記) | 出處 |
|---|---|---|
| `sub_5C8` | `sub_5C8` 就是載入場景地圖的函式。組語三行講完: | `03-scene-entry-and-tile-semantics.md` |
| `sub_86C` | `sub_86C` 裡的 `(v2 & 0xFE) == 0x10`(tile 16–17)都落在這個範圍。 | `02-movement-and-tile-flags.md`, `03-scene-entry-and-tile-semantics.md` |
| `sub_6730` | 2. 或查 `byte_41C18` 的 xref(`sub_6730` 開頭有 `for(i=0;i<256;i++) byte_41C18[i]=i`, | `01-tileset-and-dot16-loader.md` |
| `sub_C778` | 寫  0xC790   sub_C778    mov dword_65334, 1 | `03-scene-entry-and-tile-semantics.md` |
| `sub_10928` | `sub_1DA10`(聖壇)、`sub_2D564`(cave/mine/dungeon)、`sub_10928`(印地名並確認)。 | `03-scene-entry-and-tile-semantics.md` |
| `sub_1DA10` | `0x19`(25)  /  **the shrine of …**(八德聖壇;名字取自 `off_411BC[9]`,狀態看 `byte_411FC[8]`/`byte_41204[8]`)  /  `sub_1DA1… | `03-scene-entry-and-tile-semantics.md` |
| `sub_24A50` | push off_41BE0 → call sub_24A50    ← 載 .16(off_41BE0 = 表的第 16 項 "CREATE.16") | `01-tileset-and-dot16-loader.md` |
| `sub_24BC0` | push off_41BA0 → call sub_24BC0    ← 載 PROPORT.PCS | `01-tileset-and-dot16-loader.md` |
| `sub_2A610` | case 2: return (tile & 0xF0) == 0x60  /  /  sub_2A674(tile)  /  /  sub_2A610(mover, tile);  // 水陸兩棲 | `02-movement-and-tile-flags.md`, `03-scene-entry-and-tile-semantics.md` |
| `sub_2A674` | case 2: return (tile & 0xF0) == 0x60  /  /  sub_2A674(tile)  /  /  sub_2A610(mover, tile);  // 水陸兩棲 | `02-movement-and-tile-flags.md` |
| `sub_2A694` | if (v7 && sub_2A694(0, v8)) { ... }     // ← 通行判定 | `02-movement-and-tile-flags.md` |
| `sub_2B360` | v2 = sub_2B360(...); | `02-movement-and-tile-flags.md` |
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
| `byte_3E0A3` | movzx   eax, byte_3E0A3          ; 當前地點編號(1-based) | `03-scene-entry-and-tile-semantics.md` |
| `byte_3E0A5` | movzx   ebx, byte_3E0A5          ; 樓層 | `03-scene-entry-and-tile-semantics.md` |
| `byte_3E0A6` | byte_3E0A6 = 15;      // 場景內 X(32 格的中央) | `03-scene-entry-and-tile-semantics.md` |
| `byte_3E0A7` | byte_3E0A7 = 30;      // 場景內 Y(靠底部 → 城鎮南方入口) | `03-scene-entry-and-tile-semantics.md` |
| `byte_3F789` | v8 = byte_3F789[32 * dy + dx];          // ← 場景地圖緩衝,**列寬 32** | `02-movement-and-tile-flags.md` |
| `byte_400F4` | push    offset byte_400F4 | `03-scene-entry-and-tile-semantics.md` |
| `byte_41033` | dump `byte_41033[1..32]` 得到起始索引,再用「同檔下一個地點的起始索引 − 自己」算出層數: | `03-scene-entry-and-tile-semantics.md` |
| `byte_410F4` | for (i = 0; i < 32 && (byte_410F4[i] != 0  /  /  byte_4111C[i] != 0); ++i); | `03-scene-entry-and-tile-semantics.md` |
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
