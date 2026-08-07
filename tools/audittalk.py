#!/usr/bin/env python3
"""對**目前的**工作單稽核**所有**已翻的對白。

與 `checktalk.py` 的分工:

  checktalk   一批 vs 一個檔 —— 派工當下的驗收
  audittalk   全部譯文 vs 目前的 `talkwork.md` —— 收工前的總驗

為什麼需要兩支:批次工作單是**當時**產的快照。工具改過之後
(例如讓 `%A` 在工作單上顯示出來),舊快照就過期了,拿它驗會滿江紅
而且全是假的。總驗一律對最新的 `talkwork.md`。

用法:audittalk.py <talkwork.md>
"""
import glob
import re
import sys

ROW = re.compile(r"^\| `([^`]+)`( ✅)? \| (.*?) \| (.*?) \|$")
KEY = re.compile(r'^\s*"([^"]+#\d+#[a-z0-9]+)":\s*(.+?),?\s*$')
HALF = "，!?;:".replace("，", ",")


def parse_go(path):
    out, pending_key, pending = {}, None, []
    for raw in open(path, encoding="utf-8"):
        line = raw.rstrip("\n")
        if pending_key:
            pending.append(line)
            if not line.rstrip().endswith("+"):
                out[pending_key] = "".join(
                    re.findall(r'"((?:[^"\\]|\\.)*)"', "".join(pending)))
                pending_key, pending = None, []
            continue
        m = KEY.match(line)
        if not m:
            continue
        key, val = m.groups()
        if val.rstrip().endswith("+"):
            pending_key, pending = key, [val]
            continue
        parts = re.findall(r'"((?:[^"\\]|\\.)*)"', val)
        if parts:
            out[key] = "".join(parts)
    return out


def not_big5(s):
    bad = []
    for c in s:
        if ord(c) < 128:
            continue
        try:
            c.encode("big5")
        except UnicodeEncodeError:
            bad.append(c)
    return bad


def main():
    if len(sys.argv) != 2:
        sys.exit(__doc__)
    src = {}
    for line in open(sys.argv[1], encoding="utf-8"):
        m = ROW.match(line.rstrip("\n"))
        if m and "#" in m.group(1):
            src[m.group(1)] = m.group(3)

    got, where = {}, {}
    for f in sorted(glob.glob("internal/i18n/talk_*.go")):
        if f.endswith("_test.go"):
            continue
        for k, v in parse_go(f).items():
            if k in got:
                print("✗ %s 重複定義:%s 與 %s" % (k, where[k], f))
            got[k], where[k] = v, f

    bad = []
    for k, zh in sorted(got.items()):
        en = src.get(k)
        if en is None:
            bad.append("%s(%s):工作單裡沒有這個 key" % (k, where[k]))
            continue
        if not any(c.isalpha() for c in en):
            bad.append("%s(%s):來源不是文字(%r),不該有譯文" % (k, where[k], en))
        if not zh.strip():
            bad.append("%s(%s):譯文是空的" % (k, where[k]))
            continue
        for c in HALF:
            if c in zh:
                bad.append("%s(%s):半形標點 %r → %s" % (k, where[k], c, zh))
                break
        if any("぀" <= c <= "ヿ" for c in zh):
            bad.append("%s(%s):有日文假名 → %s" % (k, where[k], zh))
        nb = not_big5(zh)
        if nb:
            bad.append("%s(%s):Big5 編不出 %s → %s"
                       % (k, where[k], "".join(sorted(set(nb))), zh))
        if ("%A" in en) != ("%A" in zh):
            bad.append("%s(%s):%%A 兩邊不一致\n    EN %s\n    ZH %s"
                       % (k, where[k], en, zh))
        if zh.count("「") != zh.count("」"):
            bad.append("%s(%s):「」不成對 → %s" % (k, where[k], zh))

    for b in bad[:30]:
        print("✗", b)
    if len(bad) > 30:
        print("… 另有 %d 條" % (len(bad) - 30))
    total = len([k for k, v in src.items() if any(c.isalpha() for c in v)])
    print("工作單 %d 段,已翻 %d 段(%.1f%%),問題 %d 條"
          % (total, len(got), 100 * len(got) / max(total, 1), len(bad)))
    sys.exit(1 if bad else 0)


main()
