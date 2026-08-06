# RE-01:tileset 格式破解 + `.16` 載入鏈(進行中)

> 日期:2026-08-07 ・ 輸入檔:見 `docs/re/00-hexrays-p3-verified.md` 的 SHA-256 表。

## 一、FM Towns `EGA*.TIL` 完全破解 ✅

`U5_E/EGA0–EGA3.TIL`,各 65,536 B。**格式:128 個 tile × 512 B,每個 tile 是原版
16×16 機械放大 2 倍後的 32×32 4bpp packed**(每列 16 B × 32 列)。

推導路徑(先看資料,不猜格式):

1. 檔內位元組**只有 15–16 種唯一值,且全部是「重複 nibble」**(`00 22 AA EE FF 66 77 44 CC`)。
   4bpp packed 若沒有水平 2× 放大,圖形邊緣必然出現混合 nibble(如 `2A`)—— 實測一個都沒有。
   ⇒ 水平方向每個原始像素被複製成兩個,pack 進同一 byte。
2. 相鄰列成對相同 ⇒ 垂直也是 2×。
3. 4 檔 × 128 tile = **512 tile**;降回 16×16 後 512 × 128 B = **65,536 B**,
   **正好等於 DOS 版 `TILES.16` 檔頭宣稱的解壓後長度** —— 兩個獨立來源互相印證。

實作 `internal/u5data/tiles.go`,並把假設做成**執行期驗證條件**:降採樣時逐一檢查
「被丟掉的那一列是否真的相同」與「每個 byte 的高低 nibble 是否相同」。
512 tile × 256 px **全數通過** —— 若 FM Towns 其實是重繪過的高解圖,這裡會報錯而不是安靜吃下去。

驗收:`u5dump tiles-fmtowns` 產出的 tile sheet 可辨識出水、草、樹、山、城牆、家具、盾牌、
村民、騎士、骷髏、海蛇、龍、月相符號 —— 形狀正確。

⚠ **顏色尚未校正**:目前套標準 EGA 16 色 palette,觀感偏豔。FM Towns 的實際色號→RGB
映射需以原版截圖當 oracle 校正(u4-cht 踩過「形狀對、顏色全錯」)。**形狀正確 ≠ 顏色正確**,
兩者要分開驗收。

## 二、`TILES.16` 壓縮:未破,但已定位載入鏈

### 已確認

- `TILES.16` 前 4 byte 是 **u32LE 解壓後長度 = 65,536**;其後 30,410 B 是壓縮體(256 種位元組值,高熵)。
- **不是標準 LZW**。以 FM Towns tileset 還原出的 65,536 B 當 oracle,掃過
  {MSB, LSB} × {有/無 clear code 256} × {firstCode 256/257/258} × {earlyChange 0/1} × {skip 0/4}
  共 48 種組合,**最長相符前綴只有 4 B**。⇒ 不要再盲試參數(`rulebook/41`),改讀反編譯碼。
- **`ULTIMA.16` / `STORY1.16` / `CREATE.16` / `TEXT.16` 在 DOS 與 FM Towns 兩版 md5 完全相同**
  (`MON0.16`、`ITEMS.16` 不同)。⇒ 同格式同壓縮,而 FM Towns 執行檔可反編譯 ⇒
  **解壓演算法讀得到,只是還沒找到那個函式**。

### 載入鏈(已追出的部分)

用 IDA 的 xref 圖查 `off_41BA0`(29 個檔名的表,索引 3 = `TILES.16`)。
⚠ 這一步 **grep 反編譯輸出會回零命中** —— 存取是 `off_41BA0[edi*4]` 這種間接形式,
Hex-Rays 沒把符號名寫出來。零命中與「真的沒人用」長得一樣,所以必須查資料庫
(`tools/ida_xref.idc`,kb 紀律)。

```
off_41BA0 (0x41BA0)  29 個檔名
├─ 讀 0x68B6  sub_6730   mov eax, off_41BA0[edi*4]
│    └─ 啟動初始化:載 towns.fnt(17,280 B → dword_41D28)與 u5.fnt(0x4000 B → dword_4FFB8);
│       失敗訊息 "IBM FONT DATA READ FAIL" / "ULTIMA FONT DATA READ FAIL";
│       另外把表內所有 ".16" 就地改成 ".4"('1'→'4' 後截斷)—— EGA/CGA 檔名切換,
│       且它改寫的是 const 記憶體(Hex-Rays 有警告)
└─ 讀 0x233D8  (IDA 未識別為函式的碼)
     push off_41BA0 → call sub_24BC0    ← 載 PROPORT.PCS
     push off_41BE0 → call sub_24A50    ← 載 .16(off_41BE0 = 表的第 16 項 "CREATE.16")
```

`sub_24A50` 讀完了:它**只是 fread**,沒有解壓。依檔名分配緩衝區後整檔讀入,回傳 handle(`k + 256`):

| 檔名群 | 緩衝區 | fread size |
|---|---|---|
| `off_41BB4[0..2]`(3 檔) | `0xD6D8` = 55,000 B | 55,000 |
| `ITEMS.16` | `0x4074` = 16,500 B | 16,500 |
| `off_41BC0[0..7]`(疑為 `MON0–7.16`) | `0x1068` = 4,200 B | 4,200 |

⇒ **解壓是延後的**,發生在使用這些 handle 的地方。

### 下一步(P3 接手)

1. 查 `dword_5AC30`(handle 表)的 xref → 找取用 handle 的函式 → 那裡才有解壓。
2. 或查 `byte_41C18` 的 xref(`sub_6730` 開頭有 `for(i=0;i<256;i++) byte_41C18[i]=i`,
   這種 256 項恆等初始化很像字典或色號映射的初始值)。
3. 建 `docs/re/00-function-index.md`,把本筆記已命名的函式登記進去。

### 為什麼現在停

引擎可以先用 FM Towns 的**未壓縮** tileset(素材仍是原版,符合 `CLAUDE.md §3.0`),
所以壓縮不擋 P1/P2 進度(這正是 `PLAN.md` 風險表寫的對策)。而破壓縮的正確路徑是
系統性讀反編譯碼 —— 那是 P3 的工作,不是在 P1 這裡靠參數掃描硬撞。

## 三、順手得到的字型線索(餵回 P1.5)

`sub_6730` 揭露 FM Towns 版有兩套字型,且各自的載入大小是寫死的:

| 檔案 | 載入大小 | 緩衝區 | 錯誤訊息 |
|---|---|---|---|
| `towns.fnt` | **17,280 B** | `dword_41D28` | `IBM FONT DATA READ FAIL !!` |
| `u5.fnt` | **0x4000 = 16,384 B** | `dword_4FFB8` | `ULTIMA FONT DATA READ FAIL !!` |

- `U5.FNT` 實際檔案 16,384 B ✓ 與載入大小一致。
- `TOWNS.FNT` 實際檔案 17,160 B,但載入 17,280 B(差 120 B)—— 載入大小是上限而非實際,
  或格式帶尾端結構,待驗。
- 錯誤訊息把 `towns.fnt` 叫做 **IBM FONT**、`u5.fnt` 叫做 **ULTIMA FONT** ⇒
  前者是半角 ASCII 字型、後者是遊戲專用字型。這解釋了為什麼 `U5.FNT` 以
  「8×16 ASCII 直索引」去 dump 得不到字形(它不是 ASCII 表)。
- 對中文化的意義:**FM Towns 版的全角日文走哪一條路徑**還沒確認 —— 這兩個檔都不夠大
  到裝一套 JIS 字型(日文全角至少數千字),所以日文字型可能在 `WORRIORJ.EXP` 內、
  或走 FM Towns 的系統字型 ROM。**這正是 P3 第 1 項(日/英執行檔 diff)要回答的問題。**
