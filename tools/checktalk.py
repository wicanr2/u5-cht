#!/usr/bin/env python3
"""機械檢查一批對話譯文 —— 派給便宜模型之後的把關。

`rulebook/45`:把關不可省。這支查的都是**可機械判定**的錯,不判文筆:

  1. key 必須存在於工作單裡,而且不重複、不遺漏。
  2. 譯文不得為空。
  3. 半形標點(`,` `!` `?` `;` `:`)—— 中文句子裡混半形是看得出來的醜,
     而且倚天字庫走 Big5,全形才有對應字模。
  4. `%A`(玩家名字)—— 原文有 opcode 展開出來的名字時,譯文也要有;
     原文沒有卻多打 `%A` 會印出玩家名字在不該出現的地方。
  5. 殘留英文句子:譯文若整段沒有中日韓字元,幾乎一定是漏翻。
  6. **日文假名** —— 工作單第三欄是日文,只當語意佐證。假名跑進譯文
     代表 agent 拿日文當來源抄了(實際發生過:第 02 批 13 段)。
  7. **Big5 編不出來的字** —— 畫面用倚天點陣字,字庫是 Big5。
     簡體字、日文漢字異體、罕用字都編不出來,到玩家眼前是空框或 fallback。
     ★ 這一條同時擋掉簡繁混用(`给` / `见` / `问`),而那是肉眼最容易漏看的。
  8. 「」不成對。

用法:checktalk.py <batch.tsv> <譯好的 .go 檔>
"""
import re
import sys

KEY = re.compile(r'^\s*"([^"]+#\d+#[a-z0-9]+)":\s*(.+?),?\s*$')
HALF = ",!?;:"


def parse_go(path: str) -> dict:
    """從譯文 Go 檔抓出 key → 譯文。只認 `"key": "值",` 這種單行寫法,
    以及用 `+` 續行的多行字串。"""
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


def has_cjk(s: str) -> bool:
    return any("㐀" <= c <= "鿿" or "＀" <= c <= "￯" for c in s)


def has_kana(s: str) -> bool:
    return any("぀" <= c <= "ヿ" for c in s)


def not_big5(s: str) -> list:
    """回傳 Big5 編不出來的字。倚天字庫就是 Big5,編不出來就是畫不出來。"""
    bad = []
    for c in s:
        if ord(c) < 128:
            continue
        try:
            c.encode("big5")
        except UnicodeEncodeError:
            bad.append(c)
    return bad


def main() -> None:
    if len(sys.argv) != 3:
        sys.exit(__doc__)
    want = {}
    for line in open(sys.argv[1], encoding="utf-8"):
        parts = line.rstrip("\n").split("\t")
        if len(parts) >= 2:
            want[parts[0]] = parts[1]
    got = parse_go(sys.argv[2])

    bad = []
    for k in sorted(set(got) - set(want)):
        bad.append("多了不在工作單裡的 key:%s" % k)
    missing = sorted(set(want) - set(got))
    for k, zh in sorted(got.items()):
        if k not in want:
            continue
        en = want[k]
        if not zh.strip():
            bad.append("%s:譯文是空的" % k)
            continue
        for c in HALF:
            if c in zh:
                bad.append("%s:用了半形標點 %r → %s" % (k, c, zh))
                break
        if ("%A" in zh) != ("%A" in en or "<名字>" in en):
            # 英文那邊玩家名字已經被展開成佔位字串,所以只在譯文多打時抓。
            if "%A" in zh:
                bad.append("%s:譯文多了 %%A(原文沒有玩家名字)→ %s" % (k, zh))
        if not has_cjk(zh) and any(ch.isalpha() for ch in en):
            bad.append("%s:整段沒有中文,像是漏翻 → %s" % (k, zh))
        if has_kana(zh):
            bad.append("%s:譯文裡有日文假名(日文欄只是語意佐證)→ %s" % (k, zh))
        nb = not_big5(zh)
        if nb:
            bad.append("%s:這些字 Big5 編不出來 %s(簡體字?異體字?)→ %s"
                       % (k, "".join(sorted(set(nb))), zh))
        if zh.count("「") != zh.count("」"):
            bad.append("%s:「」不成對 → %s" % (k, zh))

    for b in bad[:40]:
        print("✗", b)
    if len(bad) > 40:
        print("… 另有 %d 條" % (len(bad) - 40))
    print("工作單 %d 段,譯出 %d 段,缺 %d 段,問題 %d 條"
          % (len(want), len(got), len(missing), len(bad)))
    if missing[:10]:
        print("  缺的前幾個:%s" % ", ".join(missing[:10]))
    sys.exit(1 if bad or missing else 0)


main()
