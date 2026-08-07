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
  9. **譯名漂移** —— 原文出現 `names.go` 裡有定譯的專有名詞時,譯文必須用那個譯法。
     系列共通名要與 u4-cht / u6-cht 對齊,自己另譯就漂了
     (實際發生過:`trolls` 被譯成「巨魔」,定譯是「山怪」)。

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


# 譯名漂移檢查:只在原文**以大寫開頭、且是完整的字**出現時才要求。
#
# ⚠ 兩道限制都是必要的,少一道就全是誤報:
#
#   - **大寫**:`names.go` 裡有 `Child → 孩童`、`Smith → 史密斯` 這種
#     生物類別與人名。小寫的 `child` / `blacksmith` 是普通名詞,
#     句子裡本來就該隨文調整(「她是個很棒的孩子」沒有錯)。
#   - **完整的字**:不設邊界的話 `smith` 會命中 `blacksmith`、
#     `Bat` 會命中 `Battle`。
GLOSSARY_MIN = 4


def load_glossary(path: str = "internal/i18n/names.go") -> dict:
    """從 names.go 抓出「英文 → 中文」。單一真相來源,不另外維護一份。"""
    out = {}
    try:
        src = open(path, encoding="utf-8").read()
    except OSError:
        return out
    pat = r'"([A-Z][A-Za-z\' .-]{%d,})":\s*"([^"]+)"' % (GLOSSARY_MIN - 1)
    for en, zh in re.findall(pat, src):
        out[en] = zh
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
    want, skip = {}, set()
    for line in open(sys.argv[1], encoding="utf-8"):
        parts = line.rstrip("\n").split("\t")
        if len(parts) < 2:
            continue
        # 來源沒有半個字母的段落(`@` 這種渲染殘留)整段不列入 ——
        # 既不要求翻,也不准出現在譯文表裡(填空字串會讓那句話消失)。
        if not any(c.isalpha() for c in parts[1]):
            skip.add(parts[0])
            continue
        want[parts[0]] = parts[1]
    got = parse_go(sys.argv[2])
    glossary = load_glossary()

    bad = []
    for k in sorted(set(got) - set(want)):
        if k in skip:
            bad.append("%s:來源不是文字(`@` 之類的渲染殘留),不要放進譯文表" % k)
        else:
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
        # `%A` 要**兩邊一致**。原本只查「多打」,漏了「少打」——
        # 而少打才是會被看見的那一種:招呼語裡的玩家名字整個不見了。
        if "%A" in en and "%A" not in zh:
            bad.append("%s:原文有玩家名字 %%A,譯文漏了 → %s" % (k, zh))
        if "%A" not in en and "%A" in zh:
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
        for term, want_zh in glossary.items():
            if want_zh in zh:
                continue
            # ⚠ **句首的大寫不算專有名詞。** 英文每一句的第一個字都大寫,
            # 於是 `Guard it well!` 的 `Guard` 會被當成生物名「衛兵」
            # (實際誤報過)。要求它**不在句首**:前面得是字母、逗號或空格,
            # 而不是字串開頭、引號或 `.!?` 之後。
            if re.search(r"(?<![.!?\"'\n]\s)(?<![.!?\"'])(?<!^)\b%ss?\b"
                         % re.escape(term), en):
                bad.append("%s:原文有 %r,定譯是 %r,譯文沒用 → %s"
                           % (k, term, want_zh, zh))
                break

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
