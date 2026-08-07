#!/usr/bin/env python3
"""把 `talkwork.md` 切成一批批派給翻譯 agent 的工作單。

每一批是一個 TSV(`key<TAB>英文<TAB>日文`),因為:

  - agent 讀 markdown 表格容易把分隔線與標題也讀進去;TSV 沒有這個問題。
  - 一批一個檔,agent 之間**檔案邊界互不重疊** —— 那是併發的前提。

**已經翻過的段落不會進工作單**(標了 ✅ 的那些),所以重跑一次就是「剩下的」。

用法:splitwork.py <talkwork.md> <輸出目錄> <每批幾段>
"""
import os
import re
import sys

ROW = re.compile(r"^\| `([^`]+)`( ✅)? \| (.*?) \| (.*?) \|$")


def main() -> None:
    if len(sys.argv) != 4:
        sys.exit(__doc__)
    src, out, per = sys.argv[1], sys.argv[2], int(sys.argv[3])
    os.makedirs(out, exist_ok=True)
    for old in os.listdir(out):
        os.remove(os.path.join(out, old))

    rows = []
    for line in open(src, encoding="utf-8"):
        m = ROW.match(line.rstrip("\n"))
        if not m:
            continue
        key, done, en, ja = m.groups()
        if done:                       # 已翻的不再派
            continue
        if "#" not in key:             # 關鍵字那種提示列
            continue
        rows.append((key, en.strip(), ja.strip()))

    # 同一筆記錄(同一個 `檔名#id`)不要被切開 —— 上下文斷掉會譯得很怪。
    batches, cur = [], []
    last_rec = None
    for key, en, ja in rows:
        rec = key.rsplit("#", 1)[0]
        if cur and rec != last_rec and len(cur) >= per:
            batches.append(cur)
            cur = []
        cur.append((key, en, ja))
        last_rec = rec
    if cur:
        batches.append(cur)

    for i, b in enumerate(batches, 1):
        path = os.path.join(out, "batch-%02d.tsv" % i)
        with open(path, "w", encoding="utf-8") as f:
            for key, en, ja in b:
                f.write("%s\t%s\t%s\n" % (key, en, ja))
        print("%s\t%d 段" % (path, len(b)))
    print("共 %d 批 / %d 段" % (len(batches), len(rows)))


main()
