# RE-00:FM Towns `.EXP`(Phar Lap P3)可反編譯 —— 已驗證

> 日期:2026-08-07 ・ 這是本專案第一份逆向筆記,結論決定了 §4.2 的逆向主目標。

## 輸入檔

| 檔案 | 大小 | SHA-256 |
|---|---|---|
| `WORRIORS.EXP`(英文主程式) | 475,719 B | `a1599baa6479f7a6d488138819c48ac9ddda35e1a90d20a6f00cb3e12c3dbcc6` |
| `WORRIORJ.EXP`(日文主程式) | 500,375 B | `f732fd6d62fa09b6b1d2881964bd5a8d9d083906a5d6fc6d65ca27b49246ba8e` |

來源:`org_game/fmtown/Ultima V - Warriors of Destiny (Japan).7z`
→ Track 1(MODE1/2352)經 `tools/cdimg_to_iso.py` 轉 ISO9660 → `7z x` 抽出 `U5_E/`。

## 結論(三條,全部實測)

### 1. IDA 9.4 原生認得 P3 格式,不必手動剝 header

`tools/ida.sh analyze WORRIORS.EXP` 直接成功(exit 0)。IDA 自報:

```
Format      : Phar Lap run386-extender flat model file
.EXP file header :
   Signature word: 3350h          ← 檔案開頭的 "P3"
   Level         : LVL_FLAT
   Header size   : 0180h
   File size     : 74247h
   Image offset  : 00000200h
   Image size    : 00074047h
Registers:
   Initial ESP   : 00089FD8h
   Initial EIP   : 00039700h      ← 進入點
   Initial SS/CS : 00000000h      ← flat model
```

產出 `WORRIORS.EXP.asm`(4,644,376 B)+ `WORRIORS.EXP.i64`(5,726,521 B)。

**位元寬度佐證**(避免「以為是 32-bit 其實不是」):`.asm` 內 32-bit 暫存器參照
(`eax`/`ebx`/`ecx`/`edx`/`esi`/`edi`/`esp`/`ebp`)**52,676 處**,16-bit `ax` 僅 **1,153 處**。
函式數 **1,233**(`proc near` 計數)。

### 2. Hex-Rays 批次反編譯成功 —— 61,364 行 C

```bash
tools/ida.sh raw idat -Ohexrays:/work/WORRIORS_hexrays.c:ALL -A WORRIORS.EXP.i64
```

exit 0,產出 **61,364 行**、**1,225 個函式**的 C。Hex-Rays 版本 9.4.0.260610。

⚠ 它自報 `Detected compiler: GNU C++` —— **這是誤判**,別當事實。FM Towns 的 386 開發
慣用 MetaWare High C / Watcom;而 DOS 與 PC-98 兩版的執行檔內含 `MS Run-Time Library ... 1988,
Microsoft Corp` 字串(見 CLAUDE.md §2.1 / §2.4)。編譯器身分待另案確認,**在確認前不要依賴
它推導呼叫慣例**(反編譯輸出已標 `__cdecl`/`__fastcall`/`__usercall`,以那個為準)。

### 3. 字串錨定可直接命中邏輯

按 `retro-game-remake` 母方法論的「字串錨定找函式」,在反編譯輸出裡 grep 遊戲檔名:

```c
sub_2C740(0,  (__int16 *)"TOWNE.TLK", byte_54700, 512,  0);
sub_2C740(v0, (__int16 *)"TOWNE.TLK", byte_54700, 1024, 0);
```

一行就給出三件事:
- **`sub_2C740` 是檔案讀取常式**,參數形狀 `(?, 檔名, 緩衝區, 長度, ?)`;
- `.TLK` 是**分段讀取**(512 / 1024 B),不是整檔載入 → 對應 §2.1 觀察到的
  「檔頭 `(u16 offset, u16 index)` 索引表」用途;
- `byte_54700` 是對話緩衝區,可從它的 xref 反追整條對話處理鏈。

其餘已出現的檔名字串:`LOOK2.DAT`(6)、`MISCMAPS.DAT`(5)、`BRIT.DAT`、`SHOPPE.DAT`、
`ULTIMA.16`、`STORY1–6.16`、`STARTSC.16`、`MON0–7.16`、`END1/2.16`、`ENDSC.16`、
`TEXT.16`、`CREATE.16` —— 每一個都是一條可用的錨。

## 對計畫的影響

- §4.2 逆向主目標 = **FM Towns `WORRIORS.EXP`**(讀 C,不讀組語)。
- 中文化 hook 點的下一步 = **`WORRIORJ.EXP` 反編譯後與 `WORRIORS.EXP` 對照**(差 24,656 B,
  差異處即日文 DBCS 字型與排版邏輯)。
- DOS 版仍是資料格式主線與 overlay 機制的來源;PC-98 降為輔助。

## 待辦(從本筆記直接展開)

1. 反編譯 `WORRIORJ.EXP`,與本檔 diff → 定位 DBCS 繪字/換行/字型載入。
2. 從 `sub_2C740` 與 `byte_54700` 的 xref 反追 `.TLK` 索引表語意與控制碼(`\x01` 疑為玩家名代入)。
3. 建函式索引表(`tools/gen_func_index.py` → `docs/re/00-function-index.md`),
   之後每命名一個 `sub_XXXX` 就登記,避免重讀(kb 紀律)。
4. 確認編譯器身分(影響結構體佈局與浮點慣例的判讀)。

> 紀律備忘:`.i64`/`.asm`/反編譯 `.c` **全部 gitignore**(見 `.gitignore`);
> 位址一律寫 IDA 線性位址;「唯一/只有一處」沒有全檔掃描佐證不要寫。
