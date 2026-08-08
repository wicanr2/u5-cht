# 89 — 配樂的解讀在哪裡(一個否定結論,以及它自己的更正)

| | |
|---|---|
| 輸入檔 | `re_work/fmtowns/WORRIORS.EXP`(+`.asm`)、`WORRIORJ.EXP`、`iso/U5OPEN/U5OPEN.EXP`、`iso/TBIOS.BIN`、`iso/CONFIG.SYS`、`iso/AUTOEXEC.BAT` |
| 主要函式 | `sub_313B8`(載入 `U5_BGM.TBL` + `FM_BANK.FMB`)、`sub_3453B`(★ 交出**檔名**)、`sub_34B1B`(★ 遠端呼叫驅動)、`sub_34D3D`(`AH=6`)、`sub_36D44`(`getenv("FMINST")`) |
| 起因 | `docs/formats/12` 留的兩個 ⬜:FMB 那 40 byte 的語意、tick → 秒 |
| 結論 | ✅ **不在遊戲執行檔裡**;❌ 但「不在我們手上」是**錯的** —— 它在 `TBIOS.BIN` |

---

## 0. ⚠⚠ 本篇的第一版有一條錯結論(留著當紀錄)

第一版寫:「那份驅動不在我們手上 …… **IDA 對我們手上的檔案無法回答這兩個問題**」。

**錯的。** 推翻它的證據是一條字串:

```
EUPHONY DRIVER FOR TOWNES by Joe Mizuno 1989 Copyright (C) FUJITSU LIMITED
```

它**同時出現在 `WORRIORS.EXP`、`WORRIORJ.EXP`、`U5OPEN.EXP` 三個檔案裡**
(`TMENU.EXP` 只有 `Joe Mizuno`)。⇒ 至少有一份與 EUPHONY 相關的東西在庫裡,
而第一版根本沒去找它。

⚠ **錯因**:第一版只做了「遊戲怎麼呼叫」這一半的追查(`sub_3453B` 交出檔名
⇒ 遊戲不解析),然後**直接跳到**「所以解析器不在我們手上」。
那是兩個不同的斷言 —— 前者有證據,後者只是沒找到。
(`diagnosis-notes/02-query-returned-empty`:「沒找到」與「不存在」是兩件事;
`rulebook/62`:下結論前先問「誰持有這個東西」。)

## 1. 站得住的那一半:遊戲只交出檔名

`sub_313B8` 把 `"FM_BANK.FMB"` 這 12 個位元組複製到堆疊,然後:

```asm
sub_3453B:
    mov  dword ptr byte_726DC, eax      ; ← 檔名的 offset
    mov  word  ptr byte_726DC+4, ds     ; ← 檔名的 segment(組成 far pointer)
    mov  bl, 0FFh ; mov ecx, 1
    mov  ah, 15h  ; call sub_34B1B
    mov  bl, 0    ; mov ah, 15h  ; call sub_34B1B
    mov  bl, 0FFh ; mov ecx, 0DDh
    mov  ah, 16h  ; call sub_34B1B

sub_34B1B:
    lgs  edi, fword ptr byte_726DC       ; 把檔名指標載進 gs:edi
    mov  ax, 110h ; mov fs, eax          ; ★ 固定的 selector 0x110
    call large fword ptr fs:loc_1180     ; ★ 遠端呼叫 linear 0x1180 的入口表
```

- **遞出去的是一個檔名指標,不是解析好的資料。**
- 那一族有 **28 個同形狀的 stub**(`AH` = 0..0x1B),全部 `call large fword ptr
  fs:loc_1120` / `fs:loc_1180` —— 也就是一張放在**低位線性位址的入口表**。
- `sub_36D44` 用 `getenv("FMINST")` 找音色目錄、補 `\`、以 `"rb"` 開檔 ——
  那是**路徑解析**不是格式解析。

⇒ **`WORRIORS.EXP` 裡沒有任何一行在解釋那 40 byte,也沒有一行把 tick 換算成時間。**
這一半的結論不變。

## 2. 那麼是誰解析的?`TBIOS.BIN`

光碟根目錄的兩個啟動檔說得很清楚:

```
CONFIG.SYS:    DEVICE=TBIOS.SYS /TBIOS.BIN
AUTOEXEC.BAT:  CONTROL   (只是選單,errorlevel 走日文字典)
```

**沒有載入任何獨立的音源驅動。** ⇒ selector 0x110 那張入口表是
**Towns BIOS 裝起來的**,而遊戲裡那條 EUPHONY 版權字串是
**介面函式庫的 banner**(Fujitsu 隨開發套件出的那層包裝),不是驅動本體。

`TBIOS.BIN`(81,920 B,`V31L31`、`90/11/21towns`、
`Copyright (c) Fujitsu Personal computer Systems Limited 1989`)裡:

| 特徵 | 次數 | 意義 |
|---|---|---|
| `mov dx, 4D8h` | 18 | FM 音源晶片的 I/O 埠 |
| `mov ah, 3Dh` | 2 | DOS 開檔 —— **它自己會讀檔** |
| 0x4D8 / 0x4DA / 0x4E0 的 LE 位元組序列 | 23 / 8 / 11 | FM 與 PCM 埠 |

⇒ **它同時碰晶片又開檔**,正好是「收一個檔名、把 FMB 讀進來寫進 YM2612」該有的
形狀。**推論**(還不是證明):`FM_BANK.FMB` 的 40 byte 由 `TBIOS.BIN` 解析。

要證實得做 §4 那件事。

## 3. 順帶兩個發現

### (a) ★ 執行檔裡**有**函式庫的符號名(雖然沒有 CodeView)

`WORRIORS.EXP` 裡有 `_mwfopen` `fgets` `fwrite` `fread` `_mwfwrite`
`get_file_handle_and_info` `get_file_pointer` `set_flags_etc` 這類名字
—— **MetaWare High C** 的執行期符號。另外還有整組 CD-ROM API 名:

```
cdr_sdrvmd cdr_rdrvmd cdr_status cdr_restore cdr_seek cdr_tseek
cdr_mstop cdr_pause cdr_continue cdr_read cdr_tread
cdr_mtplay cdr_rmtplay cdr_mphase cdr_cdinfo cdr_read2 cdr_tread2
```

⚠ `CLAUDE.md §4.2` 寫「三版全數沒有 CodeView 符號」—— 那句話**仍然對**
(NB05/08/09/10/11 都掃過),但它容易被讀成「沒有任何符號名」。
**有函式庫層級的符號名**,而且 `cdr_mtplay`(播放音軌)證實
**CD 音軌的播放介面是連進去的** ⇒ 兩條 CDDA 不是裝飾。
⬜ 誰在什麼時候呼叫它還沒追。

### (b) ⚠ 一個差點寫進文件的誤判:位移 0x359B2 的「MZ」

在 EUPHONY 字串後面 ~2.8 KB 處有 `MZ` 兩個位元組,一度被當成**嵌在 `.EXP` 裡的
16-bit 驅動執行檔**。解檔頭就知道是垃圾:

```
pages=22506 lastpage=3 → 映像 11,522,563 byte   ← 檔案本身只有 475,719 byte
檔頭 22506 paragraph = 360,096 byte              ← 比檔案還大
```

而抽出來的內容是 High C 的執行期符號名(`_mwfopen`/`fgets`/`fwrite`)。
⇒ **`MZ` 只是兩個湊巧相鄰的位元組。** 本專案第五次踩「模式比對命中不等於資料」
(前四次是 IDA 的 `aXxx` / `sub_Xxxx` 自動命名)。
**檔頭欄位一致性是幾乎零成本的正對照,先做再說。**

## 4. 下一步(可行,但是新的子專案)

目標:在 `TBIOS.BIN` 裡找到 `AH=15h/16h` 的分派,追到 FMB 的解析。

- 16-bit binary,**沒有 MZ 檔頭**(是 BIOS 映像)⇒ 照 `CLAUDE.md §4.2/§4.3`:
  手動當 16-bit 載入 IDA、手動建 segment、手動定基址。
- **沒有 Hex-Rays**(16-bit real mode)⇒ 全程讀組語。
- 錨點:`mov dx, 4D8h` 那 18 處是 FM 暫存器寫入的所在;往上追誰餵它資料。

⚠ 這條路的成本不小,而且它只影響**音色的正確性**,不影響音序
(音高、時長、聲道、選了第幾號音色都已經解出來了)。
⇒ 值不值得做是使用者的決定,不是我的。

## 5. 已經站得住的部分(不受本篇影響)

- `.EUP` 的**音序** ✅(相位有統計佐證,`docs/formats/12` §2)。
- `FM_BANK.FMB` 的**佈局** ✅(8 + 128 × 48、名字在前 8 byte,兩份獨立資料互證)。
- 曲號 → 檔名 ✅、場景 → 曲號 ✅(`docs/re/87`)。
- 兩條 CDDA ✅ **已轉成 ogg**(`tools/cdda2ogg.sh`,180 s / 371 s 與位元組數算出來的
  秒數逐一相符)—— 這條路不需要任何 FM 知識。
- 25 個 `.SND` 音效 ✅。⚠ **音效與配樂的追查方式不一樣**:音效是遊戲**自己**
  malloc 整個檔案、自己算通道(`索引 & 7`)、自己傳音高與音量(`docs/re/63`),
  所以格式追得出來;配樂整條交給 BIOS,所以追不到。**同一個專案裡兩種形狀。**

⇒ 缺的只有「那 40 byte 怎麼變成聲音」與「一個 tick 多長」。
**音樂本身的資料一個位元組都不缺。**
