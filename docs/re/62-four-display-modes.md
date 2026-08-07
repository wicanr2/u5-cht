# 62 — 四種顯示模式:驅動程式一開頭就把答案講完了

| | |
|---|---|
| 輸入檔 | `EGA.DRV`(11,654 B)、`CGA.DRV`(8,135 B)、`HER.DRV`(9,087 B)、`T1K.DRV`(7,733 B) |
| | `TILES.16`(30,414 B → 65,536)、`TILES.4`(19,347 B → 32,768) |
| 落地 | `internal/u5data/tiles.go`(`ParseCGATiles` / `CGAPalette` / `DisplayMode`)、`cmd/u5cht -display` |

`WORKLIST` 上這條寫「⬜ 素材已有(`.16`/`.4`/`.HCS`)」—— 素材在,但**沒人知道
CGA 是哪四個顏色**。猜「青 / 洋紅 / 白」是 CGA 的常識,但常識不是證據:
CGA 模式 4 有兩套調色盤各兩種亮度,共四種可能。

答案不在遊戲本體裡,在驅動程式的開頭。

## 四個 `.DRV` 的模式設定序列

四個檔案都以一串 `jmp near`(`E9 xx xx`)開頭 —— 那是驅動程式的進入點表。
往後找 `int 10h`(`CD 10`)就看到模式設定:

### `CGA.DRV` @0x058F

```asm
        mov     ah, 0
        mov     al, 4          ; 模式 4:320×200 四色
        int     10h
        mov     bl, 1
        mov     bh, 1          ; BH=1 → 設調色盤
        mov     ah, 0Bh
        int     10h            ; ★ 調色盤 = 1
        xor     bl, bl
        xor     bh, bh         ; BH=0 → 設背景 / 邊框 = BL = 0(黑)
        mov     ah, 0Bh
        int     10h
        mov     ah, 5
        xor     al, al         ; 顯示頁 0
        int     10h
```

`int 10h` AH=0Bh 的兩個子功能:BH=1 選調色盤(BL = 0 或 1)、BH=0 設背景色(BL = 0..15)。
這裡 **BL=1 → 調色盤 1**,而背景那次 **BL=0**,bit 3(亮度位元)是 0 → **低亮度**。

⇒ CGA 的四個顏色就是十六色盤的 **0(黑)、3(青)、5(洋紅)、7(淺灰)**。
不是高亮度的 11 / 13 / 15 —— 那要 BL 帶 bit 3,而原版沒有。

### `EGA.DRV` @0x0876

```asm
        mov     ah, 0
        mov     al, 0Dh        ; 模式 0Dh:320×200 十六色
        int     10h
        pop     dx
        add     dx, 24h        ; DX → 17 B 的調色盤表
        push    ds
        pop     es
        mov     al, 2
        mov     ah, 10h        ; AH=10h/AL=02h:一次載入全部調色盤暫存器
        int     10h
        mov     ah, 5
        xor     al, al
        int     10h
```

### `T1K.DRV` @0x0496

```asm
        mov     ah, 0
        mov     al, 9          ; 模式 9:Tandy 320×200 十六色
        int     10h
        pop     dx
        add     dx, 24h
        …                      ; 之後與 EGA **逐位元組相同**的 AH=10h/AL=02h
```

⇒ **Tandy 與 EGA 讀同一批 `.16`**:都是十六色、都用同一支 BIOS 呼叫載調色盤。
差別只在硬體調色盤的內容(兩份 17 B 的表還沒抽出來比對)。

### `HER.DRV`

**沒有任何 `int 10h`。** 它直接寫 `0x3B4`(CRTC 索引)、`0x3B8`(模式暫存器)、
`0x3BF`(設定暫存器)—— Hercules 不在 BIOS 的支援範圍內。單色 720×348。

## `TILES.4` 就是 CGA 的 2bpp tileset

| 檔案 | 壓縮後 | 檔頭宣稱解壓後 | = |
|---|---:|---:|---|
| `TILES.16` | 30,414 | 65,536 | 512 tile × 128 B(16×16 **4bpp**) |
| `TILES.4` | 19,347 | 32,768 | 512 tile × 64 B(16×16 **2bpp**) |

**剛好一半** —— 兩個獨立的檔頭數字互相印證,而且壓縮格式與 `.16` 相同
(同一支 LZW 解得開)。26 個 `.16` 全部有同名的 `.4` 配對。

一個位元組四個像素,**高位在左**(與 4bpp 的 hi/lo 同方向)。

### 驗收:與 EGA 逐格比形狀

`u5dump tiles-cga <gamedata> out.png` 產出 32 格一列的圖,與 EGA 那張並排看:
**圖案完全一樣,只是色數從十六色掉到四色**。深水的橫紋、草地的點、山脈的稜線、
城堡、桌椅、船、人物全部認得出來,而且四個色號整份 tileset 都用到了
(只用到兩色幾乎一定是位元取錯)。

三條自動檢查釘在 `cgatiles_test.go`:
- CGA 的每個像素色號都在 0..3(切法錯了會超出範圍)
- 整份 tileset 用滿四種色號
- **EGA 與 Tandy 的 tileset 逐 tile 相同,而 CGA 與 EGA 不同**
  (後半條擋的是「`.4` 根本沒被讀進來」——那種錯不會讓任何測試變紅)

## 落地

`u5data.DisplayMode`(EGA / CGA / Tandy / Hercules)+ `LoadTileSetFor(dir, mode)`。
CGA 的色號在載入時就換成十六色盤的 0/3/5/7(`CGAToEGA`),
所以**算繪層只有一條路徑** —— 不必為四色模式另寫一份。

命令列:`u5cht -display CGA`。

## 還沒做的:Hercules

它沒有自己的 tileset(`IBM.HCS` / `RUNES.HCS` 各 3,072 B 是**字型**,不是圖磚),
所以 `HER.DRV` 一定是把 `.16` 或 `.4` 的素材**當場轉成單色**(抖動或閾值)。
那個轉換規則要逆 `HER.DRV` 才知道。

⚠ 目前 `-display Hercules` 會**印一句「尚未實作」然後退回 EGA**,
不拿 EGA 冒充 Hercules 的畫面(`CLAUDE.md` §3.0:缺素材要優雅降級並明說)。

另外沒做的:EGA 與 Tandy 那兩份 17 B 調色盤表還沒抽出來比對。
兩者若不同,Tandy 的畫面顏色會與 EGA 有差 —— 現在兩者共用同一份 `EGAPalette`。
