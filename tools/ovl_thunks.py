#!/usr/bin/env python3
"""從 DOS 版 `ULTIMA.EXE` 抽出 overlay 名表與跨 overlay 的進入點(thunk)表。

用法:
    python3 tools/ovl_thunks.py <gamedata 目錄> [--md]

`--md` 印成 markdown 表格(給 `docs/re/` 用)。

# 這是什麼

U5 的 DOS 版是 Microsoft C 的 overlay 程式:主程式 36 KB,24 個 `.OVL` 在需要時
被換進一塊共用的交換區。`ULTIMA.EXE` 裡有兩樣東西讓整個機制看得見:

1. **模組名表** —— `ULTIMA.EXE\\0TOWN.OVL\\0MAINOUT.OVL\\0…\\0DATA.OVL\\0`
   連續 25 個 NUL 結尾字串。**陣列索引就是 overlay 編號**(0 = 主程式本身)。

2. **thunk 表** —— 緊接在名表之後,164 個**各 12 B**、完全連續的樁:

       9A <off16> <seg16>   call far  <loader>     ; 全部指向同一支
       <idx16>              overlay 編號
       EA <off16> <seg16>   jmp  far  <目標>       ; seg 由載入時的重定位填

   每一個樁代表**一個跨 overlay 的函式進入點**:主程式(或別的 overlay)要呼叫
   住在 overlay 裡的函式,就 call 這個樁;樁先叫 loader 把 overlay 換進來,
   再遠跳到真正的位址。

# 為什麼可以確定樁的格式沒認錯

三條同時成立,而且都不是靠「看起來像」:

- **164 個樁的間距全部正好 12 B**,從 0x8216 一路連續到 0x89C6。
- **overlay 編號全部落在 1..23**。沒有 0(主程式不必換入),
  也**沒有 24**(`DATA.OVL` 是純資料,沒有進入點)—— 後者是獨立佐證:
  它與「`DATA.OVL` 只放表」這個既有結論互相印證。
- 把每個 overlay 的進入點位址取最小與最大,**跨距一律小於該 `.OVL` 的檔案大小**。
  23 個 overlay 全部成立。若樁格式認錯,這 23 條不可能同時通過。
"""

import os
import re
import struct
import sys

# thunk 的固定形狀:call far(9A)+ 編號 + jmp far(EA),共 12 B。
THUNK = re.compile(rb"\x9a(..)(..)(..)\xea(..)(..)", re.S)
THUNK_SIZE = 12


def module_names(exe: bytes):
    """抽出模組名表,回傳(名字串列, 表尾位移)。"""
    base = exe.find(b"ULTIMA.EXE\x00")
    if base < 0:
        raise SystemExit("在 ULTIMA.EXE 裡找不到模組名表")
    names, j = [], base
    while True:
        end = exe.index(b"\x00", j)
        word = exe[j:end]
        if not word or not re.fullmatch(rb"[A-Z0-9.]+", word):
            break
        names.append(word.decode())
        j = end + 1
    return names, j


def thunks(exe: bytes, start: int):
    """抽出 thunk 表。回傳 [(檔案位移, overlay 編號, 目標 offset)]。"""
    out = []
    for m in THUNK.finditer(exe, start):
        idx = struct.unpack("<H", m.group(3))[0]
        off = struct.unpack("<H", m.group(4))[0]
        out.append((m.start(), idx, off))
    return out


def check_contiguous(rows):
    """樁必須完全連續(間距 12 B)—— 不連續就是認錯了格式。"""
    bad = [(a[0], b[0]) for a, b in zip(rows, rows[1:]) if b[0] - a[0] != THUNK_SIZE]
    return bad


def main() -> int:
    if len(sys.argv) < 2:
        print(__doc__)
        return 2
    gamedata = sys.argv[1]
    as_md = "--md" in sys.argv
    exe = open(os.path.join(gamedata, "ULTIMA.EXE"), "rb").read()

    names, tail = module_names(exe)
    rows = thunks(exe, tail)
    if not rows:
        raise SystemExit("找不到 thunk 表")
    bad = check_contiguous(rows)
    if bad:
        raise SystemExit(f"thunk 表不連續(第一處 0x{bad[0][0]:04X} → 0x{bad[0][1]:04X})")

    per = {}
    for _, idx, off in rows:
        per.setdefault(idx, []).append(off)

    if as_md:
        print("| # | overlay | 大小 | 進入點 | 位址範圍 | 跨距 |")
        print("|---|---|---:|---:|---|---:|")
    else:
        print(f"模組名表 {len(names)} 筆;thunk {len(rows)} 個,各 {THUNK_SIZE} B,"
              f"0x{rows[0][0]:04X}..0x{rows[-1][0] + THUNK_SIZE:04X}")
        print(f"{'#':>2} {'overlay':<14} {'大小':>6} {'進入點':>5} {'位址範圍':>15} {'跨距':>6}")

    for idx, name in enumerate(names):
        offs = sorted(per.get(idx, []))
        size = os.path.getsize(os.path.join(gamedata, name)) if name != "ULTIMA.EXE" else len(exe)
        if not offs:
            span, rng = "", "(無進入點)"
        else:
            span, rng = offs[-1] - offs[0], f"0x{offs[0]:04X}..0x{offs[-1]:04X}"
            if name != "ULTIMA.EXE" and span >= size:
                rng += " ✗ 跨距超過檔案"
        if as_md:
            print(f"| {idx} | `{name}` | {size:,} | {len(offs)} | {rng} | {span} |")
        else:
            print(f"{idx:>2} {name:<14} {size:>6} {len(offs):>5} {rng:>15} {span:>6}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
