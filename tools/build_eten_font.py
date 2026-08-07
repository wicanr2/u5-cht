#!/usr/bin/env python3
"""把倚天中文系統(ETEN)的原生點陣字烘成引擎用的 atlas。

為什麼用倚天而不是 TTF rasterize:1990s DOS 中文遊戲的中文就長這樣;倚天的
16×15 / 24×24 是為該尺寸手工調的點陣,TTF 縮到這個大小會糊、筆劃比例也不對。
(方法論見 ~/.claude/knowledge-base/retro-cht/eten-bitmap-font/SKILL.md)

用法:
    python3 tools/build_eten_font.py --eten-dir /home/anr2/cht/etan_font \\
        --iso /home/anr2/cht/etan_font/ET353S.iso \\
        --size 15 --out assets/fonts/eten-16x15

    # 只烘用到的字(P5 翻譯完成後,縮小 atlas)
    python3 tools/build_eten_font.py ... --charset-file dumps/zh_chars.txt

    # 只跑驗證 oracle,不輸出
    python3 tools/build_eten_font.py --eten-dir … --size 15 --verify-only

產出:
    <out>.png   單色 atlas(每列 cols 個字)
    <out>.json  索引:codepoints[i] 就是 atlas 第 i 格的 Unicode 碼位

[HARD] 一定要一起帶 SPCFONT:STDFONT.15 從 A440「一」開始,**不含 A140–A3BF 的全形
標點**。只帶 STDFONT 的話,,。!?「」『』()《》~ 全部會掉進 fallback,畫面上
會變成「字是倚天、標點是另一套字型」。本腳本缺 SPCFONT 時會明確報錯而不是安靜跳過。
"""
import argparse
import json
import os
import shutil
import subprocess
import sys
import tempfile

# --- 倚天字庫格式(ETEN 3.53,已實測)-------------------------------------
# 檔案      內容          尺寸    stride  字數
# STDFONT.15 漢字         16×15   30 B    13094
# SPCFONT.15 全形符號     16×15   30 B    408
# STD.24M/K/L/R/B/S 漢字  24×24   72 B    13094(ETUNPACK 壓縮,六種字體)
# SPCFONT.24 全形符號     24×24   72 B    408
# 點陣佈局:每列 (W+7)//8 bytes,MSB 在左,由上而下。
SPECS = {
    15: dict(w=16, h=15, stride=30, std="STDFONT.15", spc="SPCFONT.15"),
    24: dict(w=24, h=24, stride=72, std="STD.24M", spc="SPCFONT.24"),
}

N_HANZI = 13094
N_SPC = 408
N_COMMON = 5401          # 常用字(A440–C67E)的字數


def big5_raw(hi: int, lo: int) -> int:
    """Big5 雙位元組 → 線性序號(倚天字庫的分區索引基礎)。"""
    return (hi - 0xA1) * 157 + ((lo - 0x40) if lo < 0x7F else (lo - 0x62))


LAST_SPC = big5_raw(0xA3, 0xBF)      # 407:符號區尾
BASE_A440 = big5_raw(0xA4, 0x40)     # 漢字區起點
LAST_COMMON = big5_raw(0xC6, 0x7E)   # 常用字尾
BASE_C940 = big5_raw(0xC9, 0x40)     # 次常用字起點


def glyph_slot(hi: int, lo: int):
    """回傳 (哪個字庫, 該字庫內的索引)。不在收錄範圍回 None。"""
    r = big5_raw(hi, lo)
    if r < 0:
        return None
    if r <= LAST_SPC:
        return ("spc", r)
    if r < BASE_A440:
        return None                                    # A3C0–A43F 的空隙
    if r <= LAST_COMMON:
        return ("std", r - BASE_A440)
    if r < BASE_C940:
        return None                                    # C680–C93F 的空隙
    return ("std", N_COMMON + (r - BASE_C940))


def find_font_files(eten_dir, iso, spec, cache_dir):
    """湊齊 std / spc 兩份字庫;目錄裡缺的就從倚天光碟 ISO 抽。"""
    found = {}
    want = {"std": spec["std"], "spc": spec["spc"]}

    # 1. 先在目錄裡找(大小寫不敏感)
    if eten_dir and os.path.isdir(eten_dir):
        lower = {n.lower(): os.path.join(eten_dir, n) for n in os.listdir(eten_dir)}
        for key, name in want.items():
            if name.lower() in lower:
                found[key] = lower[name.lower()]

    # 2. 缺的從 ISO 抽(7z 認 ISO9660;倚天光碟把字庫放在 DISKS/DISK*/)
    missing = {k: v for k, v in want.items() if k not in found}
    if missing and iso:
        if not shutil.which("7z"):
            sys.exit("需要 7z 才能從 ISO 抽字庫(容器內已裝 p7zip-full)")
        os.makedirs(cache_dir, exist_ok=True)
        for key, name in missing.items():
            dest = os.path.join(cache_dir, name)
            if not os.path.exists(dest):
                # -r 是必要的:倚天光碟把字庫放在 DISKS/DISK*/,不加就匹配不到
                subprocess.run(
                    ["7z", "e", "-y", "-r", f"-o{cache_dir}", iso, name],
                    check=False, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL,
                )
            if os.path.exists(dest):
                found[key] = dest

    if "std" not in found:
        sys.exit(f"找不到漢字字庫 {want['std']}(--eten-dir 或 --iso 至少一個要提供它)")
    if "spc" not in found:
        sys.exit(
            f"找不到符號字庫 {want['spc']} —— 這不是可以跳過的:\n"
            f"  STDFONT 不含全形標點,少了它,,。!?「」()《》 全會掉 fallback。\n"
            f"  倚天光碟裡的位置是 DISKS/DISK1/{want['spc']};用 --iso 指向 ET353S.iso 即可自動抽。"
        )
    return found


def load_bank(path, stride, expect_count, spec_size):
    raw = open(path, "rb").read()

    # 24 點字型是 ETUNPACK 壓縮的,需要先解開(解壓器見 kb;本專案暫不內建)
    if raw[:14] == b"ETUNPACK V1.00":
        sys.exit(
            f"{os.path.basename(path)} 是 ETUNPACK 壓縮檔,要先解開:\n"
            f"  用 ~/scummvm/kq5/workplace/tools/etunpack.py 解出裸字庫後,"
            f"把產物放進 --eten-dir 再跑一次。\n"
            f"  (提醒:STD.24L 隸書那份在 ISO 上資料本身是壞的,實務上用 STD.24M 明體。)"
        )

    n = len(raw) // stride
    if n < expect_count:
        sys.exit(f"{os.path.basename(path)} 只有 {n} 字(預期 {expect_count}),檔案可能不完整")
    return raw


def glyph_bits(bank, idx, spec):
    """取出一個字的點陣,回傳 [(x, y), …] 亮點座標。"""
    stride, w, h = spec["stride"], spec["w"], spec["h"]
    row_bytes = (w + 7) // 8
    g = bank[idx * stride: idx * stride + stride]
    pts = []
    for y in range(h):
        row = g[y * row_bytes:(y + 1) * row_bytes]
        for x in range(w):
            if row[x // 8] & (1 << (7 - (x % 8))):
                pts.append((x, y))
    return pts


def ascii_art(pts, spec):
    grid = [["." for _ in range(spec["w"])] for _ in range(spec["h"])]
    for x, y in pts:
        grid[y][x] = "#"
    return "\n".join("".join(r) for r in grid)


def verify_oracle(banks, spec):
    """先過這關再往下做,否則整批字會整體偏移(症狀是「有字但都不對」)。"""
    print("== 驗證 oracle(kb 要求;不過就不要繼續)==")
    ok = True

    # 1. 漢字庫的第 0 個字必須是「一」——一條橫線
    pts = glyph_bits(banks["std"], 0, spec)
    ys = {y for _, y in pts}
    art = ascii_art(pts, spec)
    is_one = len(ys) <= 3 and len(pts) >= spec["w"] // 2
    print(f"\n[std idx 0] 應為「一」:{'✓' if is_one else '✗ 不像橫線,索引基準錯了'}\n{art}")
    ok &= is_one

    # 2.「中」A4A4、「猴」B555 應可辨識(人工目視)
    for ch, hi, lo in (("中", 0xA4, 0xA4), ("猴", 0xB5, 0x55)):
        slot = glyph_slot(hi, lo)
        if not slot:
            print(f"[{ch}] 索引落在空隙,公式有問題")
            ok = False
            continue
        bank, idx = slot
        print(f"\n[{ch} {hi:02X}{lo:02X} → {bank} idx {idx}]\n{ascii_art(glyph_bits(banks[bank], idx, spec), spec)}")

    # 3. 標點必須來自 spc 庫(這條擋住「漏帶 SPCFONT」)
    slot = glyph_slot(0xA1, 0x43)   # 「。」
    if not slot or slot[0] != "spc":
        print("\n[。A143] 沒有落在符號區 —— 分區公式錯了")
        ok = False
    else:
        print(f"\n[。A143 → spc idx {slot[1]}]\n{ascii_art(glyph_bits(banks['spc'], slot[1], spec), spec)}")
    return ok


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--eten-dir", default="/home/anr2/cht/etan_font", help="倚天字庫目錄")
    ap.add_argument("--iso", default="", help="倚天光碟映像(目錄裡缺的字庫從這裡抽)")
    ap.add_argument("--size", type=int, choices=sorted(SPECS), default=15, help="15 = 16×15,24 = 24×24")
    ap.add_argument("--out", default="assets/fonts/eten", help="輸出檔名前綴(產 .png 與 .json)")
    ap.add_argument("--cols", type=int, default=128, help="atlas 每列幾個字")
    ap.add_argument("--charset-file", default="", help="只烘這個檔案裡出現過的字(UTF-8)")
    ap.add_argument("--verify-only", action="store_true", help="只跑 oracle 驗證,不輸出檔案")
    args = ap.parse_args()

    spec = SPECS[args.size]
    cache = os.path.join(tempfile.gettempdir(), "u5cht-eten")
    paths = find_font_files(args.eten_dir, args.iso or None, spec, cache)
    banks = {
        "std": load_bank(paths["std"], spec["stride"], N_HANZI, args.size),
        "spc": load_bank(paths["spc"], spec["stride"], N_SPC, args.size),
    }
    for k, p in paths.items():
        print(f"字庫 {k}: {p}")

    if not verify_oracle(banks, spec):
        sys.exit("\noracle 未通過 —— 先修索引公式或確認字庫檔,不要拿偏移的字往下烘")
    if args.verify_only:
        print("\n✓ oracle 通過(--verify-only,未輸出檔案)")
        return

    # 決定要烘哪些字
    wanted = None
    if args.charset_file:
        with open(args.charset_file, encoding="utf-8") as f:
            wanted = {c for c in f.read() if not c.isspace()}

    entries = []          # (codepoint, bank, idx)
    for hi in range(0xA1, 0xFA):
        for lo in list(range(0x40, 0x7F)) + list(range(0xA1, 0xFF)):
            try:
                ch = bytes([hi, lo]).decode("big5")
            except UnicodeDecodeError:
                continue
            if wanted is not None and ch not in wanted:
                continue
            slot = glyph_slot(hi, lo)
            if not slot:
                continue
            bank, idx = slot
            if idx * spec["stride"] + spec["stride"] > len(banks[bank]):
                continue
            entries.append((ord(ch), bank, idx))

    if wanted is not None:
        missing = wanted - {chr(cp) for cp, _, _ in entries}
        if missing:
            # fallback 數量是品質指標:一大批字掉進來 → 先懷疑索引或字庫,不要無腦補字型
            print(f"⚠ Big5 沒有收錄 {len(missing)} 字,需另以 TTF 烘同尺寸補:{''.join(sorted(missing))[:60]}")

    try:
        from PIL import Image
    except ImportError:
        sys.exit("需要 Pillow(容器內已裝 python3-pil);或加 --verify-only 只做驗證")

    cols = args.cols
    rows = (len(entries) + cols - 1) // cols
    img = Image.new("1", (cols * spec["w"], rows * spec["h"]), 0)
    px = img.load()
    for i, (_, bank, idx) in enumerate(entries):
        ox, oy = (i % cols) * spec["w"], (i // cols) * spec["h"]
        for x, y in glyph_bits(banks[bank], idx, spec):
            px[ox + x, oy + y] = 1

    os.makedirs(os.path.dirname(args.out) or ".", exist_ok=True)
    img.save(args.out + ".png")
    meta = {
        "source": "ETEN 3.53 原生點陣字",
        "glyph_width": spec["w"],
        "glyph_height": spec["h"],
        "cols": cols,
        "count": len(entries),
        "codepoints": [cp for cp, _, _ in entries],
    }
    with open(args.out + ".json", "w", encoding="utf-8") as f:
        json.dump(meta, f, ensure_ascii=False, separators=(",", ":"))
    print(f"\n✓ {args.out}.png  {img.width}×{img.height}({len(entries)} 字,{cols} 字/列)")
    print(f"✓ {args.out}.json 索引")
    print("\n提醒:烘出的 atlas 衍生自倚天字庫(1993 商業產品),**不入 git**;"
          "開發者各自從自備的字庫重跑本腳本。")


if __name__ == "__main__":
    main()
