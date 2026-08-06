#!/usr/bin/env python3
"""把 MODE1/2352 的 CD raw 映像(.bin)轉成 2048 B/sector 的 ISO9660。

用法:
    python3 tools/cdimg_to_iso.py <Track1.bin> <輸出.iso>

轉完用 7z 直接抽檔(7z 認 ISO9660):
    7z l <輸出.iso>
    7z x -o<目錄> <輸出.iso> "U5_E/WORRIORS.EXP"

背景:MODE1/2352 的每個 sector 是
    12 B sync(00 FF×10 00)+ 4 B header + 2048 B 使用者資料 + 288 B EDC/ECC
所以只要每 2352 B 取 offset 16..2064。音軌(TRACK … AUDIO)不要餵進來——
那是 CDDA 原始取樣,轉 ogg 用 ffmpeg 直接吃 raw s16le 44100 Hz 立體聲。

已驗證(2026-08-07):Ultima V FM Towns 日文版
    Track 1 = 31,399,200 B = 13,350 sectors × 2352 → 27,340,800 B ISO
    ISO9660 內 239 檔 / 4 目錄(含 WORRIORJ.EXP / WORRIORS.EXP)
"""
import os
import sys

RAW_SECTOR = 2352
USER_DATA = 2048
DATA_OFFSET = 16                      # 12 B sync + 4 B header
SYNC = b"\x00" + b"\xff" * 10 + b"\x00"


def convert(src, dst):
    size = os.path.getsize(src)
    if size % RAW_SECTOR:
        raise SystemExit(
            f"{src} 不是 {RAW_SECTOR} B/sector 的整數倍({size} B)——確認 .cue 裡是 MODE1/2352"
        )
    with open(src, "rb") as f:
        head = f.read(12)
        if head != SYNC:
            raise SystemExit(f"開頭沒有 MODE1 sync pattern(讀到 {head.hex(' ')})——這可能是音軌")
        f.seek(0)
        with open(dst, "wb") as o:
            n = 0
            while True:
                sector = f.read(RAW_SECTOR)
                if len(sector) < RAW_SECTOR:
                    break
                o.write(sector[DATA_OFFSET:DATA_OFFSET + USER_DATA])
                n += 1
    print(f"{src}\n  → {dst}  ({n:,} sectors, {n * USER_DATA:,} B)")


def main():
    if len(sys.argv) != 3:
        raise SystemExit(__doc__)
    convert(sys.argv[1], sys.argv[2])


if __name__ == "__main__":
    main()
