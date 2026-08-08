#!/usr/bin/env python3
"""找出 Hex-Rays 安靜截斷的函式。

CLAUDE.md §4.4 記著 Hex-Rays 的第一種失效形態:常數傳播把邏輯摺掉,
症狀是「常數回傳 / 條件恆真 / 函式被摺成一行」—— 看得出來,所以會回去讀組語。

**第二種形態沒有那個信號。** `sub_CD28`(許願井)被反編譯成一段看起來完整、
邏輯自洽、還會 return、連警告註解都沒有的函式,而組語版有三個參數、四倍長 ——
整個彩蛋不見了(`docs/re/65`)。第一次踩到時,代價是在 `strings.go` 留下一句
「原版不管答什麼都沒有後續」的錯結論,而那條錯了一整個機制。

本工具把那次的判斷法變成可重跑的清單,兩個獨立信號:

  A. 參數列對不上 —— 組語的 proc 宣告了 `arg_N`(IDA 自己的堆疊框分析),
     而反編譯的定義是 `()` 或 `(void)`。這是 `sub_CD28` 當時最明顯的跡象。
  B. 字串常數掉了 —— 組語本體 `push offset aXxx` 引用的字串個數,
     多於反編譯本體裡的字串字面值個數。內容被丟掉時這一項最靈敏。

兩個信號都不是證明,是**該回去讀組語的清單**。B 的排序等於「漏掉多少內容」,
所以從上面開始讀最划算。

用法(不需要 docker,純文字處理):

    tools/dev.sh python3 tools/hexrays_truncation.py \\
        re_work/fmtowns/WORRIORS.EXP.asm re_work/fmtowns/WORRIORS_hexrays.c

    # 只看遊戲邏輯(0x32000 以上是 Phar Lap extender 與執行期程式庫)
    … --max-addr 0x32000

    # 只看某一種信號
    … --signal args      # 只列參數列對不上的
    … --signal strings   # 只列字串掉了的(預設兩者都列)
"""

import argparse
import re
import sys

# 0x32000 以上是執行期程式庫 / Phar Lap extender —— printf 一族本來就是 varargs,
# 參數列對不上是正常的,不是截斷。
DEFAULT_MAX_ADDR = 0x32000

RE_PROC = re.compile(r"^(sub_[0-9A-F]+)\s+proc near")
RE_ENDP = re.compile(r"^(sub_[0-9A-F]+)\s+endp")
RE_ARG = re.compile(r"^(arg_[0-9A-F]+)\s*=")
RE_STR_REF = re.compile(r"offset (a[A-Za-z0-9_]+)")
RE_STR_DEF = re.compile(r"^(a[A-Za-z0-9_]+)\s+db\s+(.*)$", re.M)
RE_C_DEF = re.compile(r"\n(?:[A-Za-z_][\w \*]*?)\b(sub_[0-9A-F]+)\(([^)]*)\)\s*\n\{")
RE_C_STR = re.compile(r'"(?:[^"\\]|\\.)*"')


def parse_asm(path):
    """回傳 {函式: {args, strings, lines}} 與 {字串 label: 內容}。"""
    text = open(path, encoding="utf-8", errors="replace").read()
    literals = {}
    for m in RE_STR_DEF.finditer(text):
        literals[m.group(1)] = "".join(re.findall(r"'([^']*)'", m.group(2)))

    procs = {}
    cur = None
    for line in text.split("\n"):
        m = RE_PROC.match(line)
        if m:
            cur = m.group(1)
            procs[cur] = {"args": set(), "strings": [], "lines": 0}
            continue
        if cur is None:
            continue
        if RE_ENDP.match(line):
            cur = None
            continue
        p = procs[cur]
        p["lines"] += 1
        m = RE_ARG.match(line)
        if m:
            p["args"].add(m.group(1))
        # 註解裡的 XREF 也會出現 `offset aXxx`,先把註解切掉。
        for m in RE_STR_REF.finditer(line.split(";")[0]):
            if m.group(1) not in p["strings"]:
                p["strings"].append(m.group(1))
    return procs, literals


def parse_c(path):
    """回傳 {函式: (參數個數, 字串個數, 行數)}。只看**定義**,不看宣告。"""
    text = open(path, encoding="utf-8", errors="replace").read()
    out = {}
    for m in RE_C_DEF.finditer(text):
        fn, params = m.group(1), m.group(2).strip()
        end = text.find("\n}\n", m.start())
        body = text[m.start():end if end > 0 else len(text)]
        n_params = 0 if params in ("", "void") else params.count(",") + 1
        out[fn] = (n_params, len(RE_C_STR.findall(body)), body.count("\n"))
    return out


def main():
    ap = argparse.ArgumentParser(description=__doc__,
                                 formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("asm", help="IDA 產出的 .asm")
    ap.add_argument("c", help="Hex-Rays 批次反編譯產出的 .c")
    ap.add_argument("--max-addr", default=hex(DEFAULT_MAX_ADDR),
                    help=f"只看這個位址以下的函式(預設 {hex(DEFAULT_MAX_ADDR)})")
    ap.add_argument("--signal", choices=("args", "strings", "both"), default="both")
    ap.add_argument("--top", type=int, default=0, help="只列前 N 筆(0 = 全部)")
    args = ap.parse_args()
    max_addr = int(args.max_addr, 0)

    procs, literals = parse_asm(args.asm)
    cdefs = parse_c(args.c)
    print(f"組語 {len(procs)} 個 proc、反編譯 {len(cdefs)} 個定義;"
          f"只看 < {hex(max_addr)}", file=sys.stderr)

    rows = []
    for fn, p in procs.items():
        if fn not in cdefs:
            continue
        if int(fn[4:], 16) >= max_addr:
            continue
        n_params, n_str, n_lines = cdefs[fn]
        lost_args = len(p["args"]) > 0 and n_params == 0
        lost_str = max(0, len(p["strings"]) - n_str)
        if args.signal == "args" and not lost_args:
            continue
        if args.signal == "strings" and lost_str == 0:
            continue
        if args.signal == "both" and not lost_args and lost_str == 0:
            continue
        rows.append((lost_str, len(p["args"]) if lost_args else 0, fn,
                     n_params, len(p["args"]), n_str, len(p["strings"]),
                     n_lines, p["lines"], p["strings"][:6]))

    rows.sort(reverse=True)
    if args.top:
        rows = rows[:args.top]
    print(f"{'函式':<12} {'參數 C/asm':>10} {'字串 C/asm':>10} {'行數 C/asm':>12}  前幾個掉的字串")
    for (_, _, fn, cp, ap_, cs, as_, cl, al, sample) in rows:
        flag = "★" if cp == 0 and ap_ > 0 else " "
        peek = " | ".join(repr(literals.get(s, s))[:22] for s in sample)
        print(f"{flag}{fn:<11} {cp:>4}/{ap_:<5} {cs:>4}/{as_:<5} {cl:>5}/{al:<6}  {peek}")
    print(f"\n共 {len(rows)} 個函式值得回去讀組語。★ = 參數列也對不上。", file=sys.stderr)


if __name__ == "__main__":
    main()
