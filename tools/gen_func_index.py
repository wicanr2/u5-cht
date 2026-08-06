#!/usr/bin/env python3
"""掃 docs/ 下的逆向筆記,把已命名的函式與全域彙整成一份索引。

    python3 tools/gen_func_index.py > docs/re/00-function-index.md

為什麼需要這個(IDA kb 的紀律):逆向筆記超過二三十份之後,光靠記憶一定會重讀
已經解過的函式。每次追蹤前先查索引,比重新 grep 一次反編譯輸出便宜得多。

作法:從筆記裡抓 `sub_XXXX` / `byte_XXXX` / `dword_XXXX` / `off_XXXX`,
連同**出現它的那一行**當描述(所以筆記寫得好,索引就有用)。
"""
import glob
import os
import re
from collections import defaultdict

SYMBOL = re.compile(r'\b((?:sub|byte|word|dword|off|loc)_[0-9A-F]{3,8})\b')
DOCS = "docs/**/*.md"


def main():
    hits = defaultdict(list)          # 符號 → [(檔案, 行號, 該行內容)]
    for path in sorted(glob.glob(DOCS, recursive=True)):
        if path.endswith("00-function-index.md"):
            continue
        with open(path, encoding="utf-8") as f:
            for n, line in enumerate(f, 1):
                for sym in set(SYMBOL.findall(line)):
                    hits[sym].append((path, n, line.strip()))

    print("# RE-00b:函式與全域索引(自動產生)")
    print()
    print("> `python3 tools/gen_func_index.py > docs/re/00-function-index.md` 重新產生。")
    print("> **讀任何 `sub_XXXX` 之前先查這裡** —— 筆記超過二三十份後,憑記憶一定會重讀已解過的東西。")
    print(f">")
    print(f"> 目前收錄 **{len(hits)}** 個符號,來源是 `docs/` 下的逆向筆記。")
    print()
    print("| 符號 | 已知語意(取自筆記) | 出處 |")
    print("|---|---|---|")

    def sortkey(s):
        kind, addr = s.split("_", 1)
        order = {"sub": 0, "off": 1, "byte": 2, "word": 3, "dword": 4, "loc": 5}
        return (order.get(kind, 9), int(addr, 16))

    for sym in sorted(hits, key=sortkey):
        # 取最有資訊量的一行(最長的,通常是表格列或說明句)
        path, n, line = max(hits[sym], key=lambda x: len(x[2]))
        desc = line
        # 砍掉 markdown 表格與程式碼雜訊,留可讀的部分
        desc = re.sub(r'^[|\s#>*-]+', '', desc)
        desc = desc.replace("|", " / ").strip()
        if len(desc) > 110:
            desc = desc[:110] + "…"
        where = ", ".join(sorted({f"`{os.path.basename(p)}`" for p, _, _ in hits[sym]}))
        print(f"| `{sym}` | {desc} | {where} |")


if __name__ == "__main__":
    main()
