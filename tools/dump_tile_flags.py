#!/usr/bin/env python3
"""從原版執行檔 dump tile 通行 bitmap,產生 internal/u5data/tileflags.go 的表。

    python3 tools/dump_tile_flags.py re_work/fmtowns/WORRIORS.EXP

來源與推導見 docs/re/02-movement-and-tile-flags.md。
表在 FM Towns .EXP 的記憶體位址 0x5FF6C(檔案位移 +0x200),64 byte = 512 tile × 1 bit,
bit = 1 代表阻擋一般行走。

自我檢查:tile 1/2/3 是水,必須被阻擋(byte[0] 的 bit6/5/4 = 1)。
"""
import sys

ADDR = 0x5FF6C
IMAGE_OFF = 0x200   # Phar Lap P3 的 image offset(IDA 自報)
SIZE = 64           # 512 tile / 8


def main():
    if len(sys.argv) != 2:
        raise SystemExit(__doc__)
    d = open(sys.argv[1], "rb").read()
    off = ADDR + IMAGE_OFF
    if len(d) < off + SIZE:
        raise SystemExit(f"檔案只有 {len(d)} B,取不到 0x{off:X} 的表")
    tbl = d[off:off + SIZE]

    # oracle:水必須阻擋。不過這關就別拿去用。
    for t in (1, 2, 3):
        if not (tbl[t >> 3] & (128 >> (t & 7))):
            raise SystemExit(f"tile {t}(水)沒有被標成阻擋 —— 位址或 image offset 不對")

    blocked = sum(1 for t in range(512) if tbl[t >> 3] & (128 >> (t & 7)))
    print(f"# 阻擋 {blocked}/512,可走 {512 - blocked}", file=sys.stderr)
    for i in range(0, SIZE, 8):
        print("\t" + ", ".join(f"0x{b:02X}" for b in tbl[i:i + 8]) + ",")


if __name__ == "__main__":
    main()
