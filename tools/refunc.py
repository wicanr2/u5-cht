#!/usr/bin/env python3
"""讀一支反組譯函式:組語本體、反編譯的 C、以及誰呼叫它。

為什麼要有這支:`WORRIORS.EXP.asm` 是 4.6 MB、反編譯的 C 是 1.3 MB,
每次用 `sed -n` 找 `proc`/`endp` 的行號既慢又容易切錯邊界 ——
而**切錯邊界正是漏掉尾巴的原因**(`docs/re/99` §0:`call` 在 `retn`
前面一行的機制被漏掉兩次)。這支一律從 `proc` 印到 `endp`。

    ./tools/dev.sh python3 tools/refunc.py sub_2A50C            # 組語(預設)
    ./tools/dev.sh python3 tools/refunc.py --c sub_2A50C        # 反編譯的 C
    ./tools/dev.sh python3 tools/refunc.py --callers sub_2BCC8  # 誰呼叫它
    ./tools/dev.sh python3 tools/refunc.py --all sub_FEC        # 三者都印

⚠⚠ **反編譯的 C 不是真值**(`CLAUDE.md §4.4`)。它有三種失效形態,
而三種都不會報錯:

    空參數列  → 函式被**截斷**(`sub_CD28` 的許願井彩蛋整段不見)
    常數回傳  → 邏輯被**摺疊**(`sub_2BCC8` 印成 `return 0;`,實際 55 行)
    常數參數  → 全域被**傳播**(`sub_2ECE8(0)` 的真身是 `sub_2ECE8(byte_3E0AE)`)

⇒ `--c` 只當導航;要下結論一律讀 `--asm`。
"""
import argparse
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
ASM = os.path.join(ROOT, "re_work/fmtowns/WORRIORS.EXP.asm")
CSRC = os.path.join(ROOT, "re_work/fmtowns/WORRIORS_hexrays.c")


def load(path: str) -> str:
    if not os.path.exists(path):
        sys.exit(f"讀不到 {path} —— IDA 產物不入庫,先跑 tools/ida.sh analyze")
    return open(path, errors="replace").read()


def asm_body(name: str) -> list[str]:
    """從 `proc` 印到 `endp` —— **不要**自己算行數截斷。"""
    lines = load(ASM).splitlines()
    start = None
    for i, ln in enumerate(lines):
        if re.match(r"^" + re.escape(name) + r"\s+proc\b", ln):
            start = i
        elif start is not None and re.match(r"^" + re.escape(name) + r"\s+endp\b", ln):
            return lines[start : i + 1]
    return []


def c_body(name: str) -> str:
    """取反編譯的**定義**(宣告以 `;` 結尾,要排除)。"""
    src = load(CSRC)
    for m in re.finditer(r"\b" + re.escape(name) + r"\(", src):
        k = src.find("\n{\n", m.end())
        if k < 0 or ";" in src[m.end() : k]:
            continue
        return src[src.rindex("\n", 0, m.start()) : src.index("\n}\n", k) + 3]
    return ""


def callers(name: str) -> dict[str, int]:
    """誰 `call` 它(從 .asm 數,反編譯的 C 會漏掉 `__usercall` 那一族)。"""
    lines = load(ASM).splitlines()
    own, cur, hits = {}, "?", {}
    for i, ln in enumerate(lines):
        if m := re.match(r"^(\w+)\s+proc\b", ln):
            cur = m.group(1)
        own[i] = cur
    for i, ln in enumerate(lines):
        if re.search(r"\bcall\s+" + re.escape(name) + r"\b", ln):
            hits[own[i]] = hits.get(own[i], 0) + 1
    return hits


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    ap.add_argument("names", nargs="+", help="函式名(sub_XXXX)")
    ap.add_argument("--c", action="store_true", help="印反編譯的 C(⚠ 不是真值)")
    ap.add_argument("--callers", action="store_true", help="印呼叫者")
    ap.add_argument("--all", action="store_true", help="組語 + C + 呼叫者")
    a = ap.parse_args()
    want_asm = a.all or not (a.c or a.callers)

    for name in a.names:
        print(f"\n{'=' * 72}\n{name}\n{'=' * 72}")
        if a.callers or a.all:
            h = callers(name)
            print("← " + (", ".join(f"{k}×{v}" for k, v in sorted(h.items())) or "(沒有 call 它)"))
        if a.c or a.all:
            body = c_body(name)
            print("\n--- 反編譯的 C(⚠ 導航用,不是真值)")
            print(body.strip() if body else "(反編譯輸出裡找不到定義)")
        if want_asm:
            body = asm_body(name)
            print(f"\n--- 組語({len(body)} 行,proc→endp)")
            if not body:
                print("(找不到 proc —— 可能不是獨立函式)")
            for ln in body:
                if ln.strip():
                    print(re.sub(r"^\s{16}", "        ", ln.rstrip()))
    return 0


if __name__ == "__main__":
    sys.exit(main())
