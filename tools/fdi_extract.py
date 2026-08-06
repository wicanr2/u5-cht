#!/usr/bin/env python3
"""從 PC-98 .fdi 磁碟映像抽出檔案(FAT12,磁區 1024 B)。

用法:
    python3 tools/fdi_extract.py <映像.fdi> <輸出目錄> [檔名 ...]

不指定檔名就全抽。列目錄用 `-l`:
    python3 tools/fdi_extract.py -l <映像.fdi>

已驗證(2026-08-06):Ultima V PC-98 版兩張磁碟
  ウルティマⅤ_Program.fdi  / ウルティマⅤ_Britania.fdi
  各 1,265,664 B = 4096 B FDI 檔頭 + 1,232,896 B raw
  boot sector: bytes/sector=1024, sec/clus=1, nFAT=2, rootEnt=192, secPerFAT=2
"""
import os
import struct
import sys

FDI_HEADER = 4096


def _load(path):
    """讀映像並跳過 FDI 檔頭,回傳 raw disk 內容與 BPB 參數。"""
    raw = open(path, "rb").read()[FDI_HEADER:]
    bps = struct.unpack("<H", raw[11:13])[0]
    if bps not in (512, 1024, 2048):
        raise SystemExit(f"看不出 FAT boot sector(bytes/sector={bps});這可能不是 FAT12 映像")
    spc = raw[13]
    reserved = struct.unpack("<H", raw[14:16])[0]
    nfat = raw[16]
    nroot = struct.unpack("<H", raw[17:19])[0]
    spf = struct.unpack("<H", raw[22:24])[0]
    fat = raw[reserved * bps: (reserved + spf) * bps]
    root = (reserved + nfat * spf) * bps
    data = root + nroot * 32
    return raw, fat, root, data, nroot, bps, spc


def _next_cluster(fat, c):
    """FAT12 的 12-bit 項:偶數取低 12 位,奇數取高 12 位。"""
    off = c * 3 // 2
    v = struct.unpack("<H", fat[off:off + 2])[0]
    return (v & 0xFFF) if c % 2 == 0 else (v >> 4)


def entries(path):
    """列舉根目錄項,回傳 (檔名, 起始 cluster, 大小)。跳過已刪除與 volume label。"""
    raw, fat, root, data, nroot, bps, spc = _load(path)
    out = []
    for i in range(nroot):
        e = raw[root + i * 32: root + i * 32 + 32]
        if not e or e[0] == 0:
            break
        if e[0] == 0xE5 or e[11] & 0x08:      # 已刪除 / volume label
            continue
        name = e[:8].decode("latin1").rstrip() + "." + e[8:11].decode("latin1").rstrip()
        clus = struct.unpack("<H", e[26:28])[0]
        size = struct.unpack("<I", e[28:32])[0]
        out.append((name.rstrip("."), clus, size))
    return out


def extract(path, outdir, want=None):
    raw, fat, root, data, nroot, bps, spc = _load(path)
    os.makedirs(outdir, exist_ok=True)
    got = []
    for name, clus, size in entries(path):
        if want and name not in want:
            continue
        buf = b""
        c = clus
        while 2 <= c < 0xFF0 and len(buf) < size + bps:
            off = data + (c - 2) * spc * bps
            buf += raw[off: off + spc * bps]
            c = _next_cluster(fat, c)
        open(os.path.join(outdir, name), "wb").write(buf[:size])
        got.append((name, size))
    return got


def main():
    args = sys.argv[1:]
    if not args:
        raise SystemExit(__doc__)
    if args[0] == "-l":
        for name, clus, size in entries(args[1]):
            print(f"{name:14} clus={clus:5} size={size}")
        return
    img, outdir, want = args[0], args[1], set(args[2:]) or None
    for name, size in extract(img, outdir, want):
        print(f"抽出 {name:14} {size} B")


if __name__ == "__main__":
    main()
