#!/usr/bin/env python3
"""盤點 docs/ 裡的斷言還成不成立。

三種機械可查的過期方式,對應三個檢查:

1. **不存在的 IDA 符號**:筆記寫 `sub_XXXX` / `byte_XXXXX`,但反組譯檔裡沒有這個
   符號。多半是打錯字或憑記憶寫的位址 —— 而位址錯了,整條溯源就斷了。
2. **不存在的引擎符號**:「引擎對應」表格寫 `game.(*State).foo`,但那個方法已經
   改名或刪掉。`rulebook/63` 的原話:code 是唯一真相,dated 文件會過期。
3. **孤立的筆記**:`docs/re/00-function-index.md` 沒收錄的筆記檔 —— 索引本身過期。

⚠ 這支工具**只查「符號還在不在」,不查「敘述對不對」**。敘述錯誤(例如把兩支
互斥的函式寫成同一回合都會跑)只能靠讀碼發現,沒有機械捷徑。所以它是**下限**
不是保證:報告全綠不代表筆記正確。

用法(一律走 docker,見 CLAUDE.md §3):

    ./tools/dev.sh python3 tools/check_doc_claims.py
    ./tools/dev.sh python3 tools/check_doc_claims.py --asm re_work/fmtowns/WORRIORS.EXP.asm
"""

from __future__ import annotations

import argparse
import pathlib
import re
import sys

ROOT = pathlib.Path(__file__).resolve().parent.parent

# IDA 自動命名的符號。`loc_` 不查 —— 筆記引用區塊標籤很常見,而它們會隨重新
# 分析而變動,不是斷言的基礎。
IDA_SYMBOL = re.compile(r"\b((?:sub|byte|word|dword|qword|off|unk|nullsub)_[0-9A-F]{1,8})\b")

# 「引擎對應」表格裡的 Go 符號:`game.(*State).foo`、`u5data.Foo`、`game.Foo`。
# 結尾的 `*`(如 `u5data.NPCAI*`)是刻意的「這一族常數」寫法,不是單一符號。
GO_SYMBOL = re.compile(r"\b(game|u5data|i18n|render|cjk|audio)\.(?:\(\*(\w+)\)\.)?(\w+)\b(?!\*)")

# 產生的檔案不查:`00-function-index.md` 的欄位會**截斷在符號中間**
# (`sub_1DA10` → `sub_1DA1…`),那是產生器的排版,不是筆記寫錯位址。
GENERATED_DOCS = {"00-function-index.md"}

# 我自己從 `base + offset` 算出來的位址。IDA 不會給這種位址取名字
# (存取寫成 `[eax+edi+42B0h]`),所以「反組譯檔裡查不到」是正常的 ——
# 但**推導必須寫下來**,否則下一個人(包括我)無從複核。
COMPUTED_ADDR_DOC = "re/00-computed-addresses.md"

# 筆記裡刻意保留的「已被推翻」段落 —— 這些本來就會提到不存在的符號。
OVERTURNED_MARKERS = (
    "此前", "已被推翻", "推翻", "原本寫", "錯因", "~~",
    # 「X 不存在」「X 應為 Y」本身就是在說那個符號**不成立** —— 不是在斷言它存在。
    "不存在", "應為", "已更正", "更正:",
)


def load_asm_symbols(asm: pathlib.Path) -> set[str]:
    """反組譯檔裡出現過的所有 IDA 符號。"""
    if not asm.exists():
        return set()
    text = asm.read_text(errors="replace")
    return set(IDA_SYMBOL.findall(text))


def load_go_symbols() -> set[str]:
    """引擎裡所有 exported / unexported 的頂層識別字與方法名。

    只收名字不收 receiver 型別 —— 筆記寫 `game.(*State).foo` 時我們要問的是
    「還有沒有一個叫 foo 的方法」,而不是它掛在哪個型別上。
    """
    names: set[str] = set()
    decl = re.compile(
        r"^(?:func\s+(?:\([^)]*\)\s*)?(\w+)"  # func Foo / func (r T) Foo
        r"|type\s+(\w+)"
        r"|\t(\w+)\s*(?:[=,A-Za-z\[*]|$)"  # const / var 區塊(含 iota 的裸名一行)
        r"|(?:const|var)\s+(\w+))"
    )
    for path in (ROOT / "internal").rglob("*.go"):
        for line in path.read_text(errors="replace").splitlines():
            m = decl.match(line)
            if m:
                names.update(g for g in m.groups() if g)
    return names


def is_overturned_context(line: str) -> bool:
    return any(mark in line for mark in OVERTURNED_MARKERS)


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--asm", default="re_work/fmtowns/WORRIORS.EXP.asm")
    ap.add_argument("--docs", default="docs")
    args = ap.parse_args()

    asm_path = ROOT / args.asm
    asm_symbols = load_asm_symbols(asm_path)
    go_symbols = load_go_symbols()
    # 根目錄那幾份也要查 —— `CONTEXT.md` 與 `PLAN.md` 正是最容易留下陳舊斷言的
    # 地方(它們寫得早、之後很少回頭改)。2026-08-08 首跑就在這兩份裡各抓到一條
    #「`TILES.16` 壓縮未破」,而它早就破了。
    docs = sorted((ROOT / args.docs).rglob("*.md"))
    docs += [ROOT / name for name in ("CONTEXT.md", "PLAN.md", "WORKLIST.md", "README.md") if (ROOT / name).exists()]

    # 推導過的位址當成已知符號 —— 但**只認登記在案的**。
    computed = ROOT / args.docs / COMPUTED_ADDR_DOC
    if computed.exists():
        registered = set(IDA_SYMBOL.findall(computed.read_text(errors="replace")))
        asm_symbols |= registered
        print(f"推導位址登記表:{len(registered)} 筆({COMPUTED_ADDR_DOC})")
    else:
        print(f"⚠ 沒有 {COMPUTED_ADDR_DOC} —— 推導出來的位址會被誤報成錯誤")

    if not asm_symbols:
        print(f"⚠ 讀不到 {args.asm} —— 跳過 IDA 符號檢查(反組譯產物不入庫)")
    print(f"引擎符號 {len(go_symbols)} 個、筆記 {len(docs)} 份\n")

    bad_ida: list[tuple[str, int, str]] = []
    bad_go: list[tuple[str, int, str]] = []

    for doc in docs:
        rel = doc.relative_to(ROOT)
        if doc.name in GENERATED_DOCS:
            continue
        for n, line in enumerate(doc.read_text(errors="replace").splitlines(), 1):
            if is_overturned_context(line):
                continue
            if asm_symbols:
                for sym in IDA_SYMBOL.findall(line):
                    if sym not in asm_symbols:
                        bad_ida.append((str(rel), n, sym))
            for _pkg, _recv, name in GO_SYMBOL.findall(line):
                if name and name not in go_symbols:
                    bad_go.append((str(rel), n, name))

    if bad_ida:
        print(f"## 反組譯檔裡查不到的 IDA 符號({len(bad_ida)} 筆)")
        print("   位址錯了整條溯源就斷了 —— 逐筆回去對 .asm。\n")
        for path, n, sym in bad_ida:
            print(f"  {path}:{n}  {sym}")
        print()

    if bad_go:
        print(f"## 引擎裡查不到的 Go 符號({len(bad_go)} 筆)")
        print("   多半是重構後改名 —— 更新「引擎對應」表格,不要留舊名字。\n")
        for path, n, sym in bad_go:
            print(f"  {path}:{n}  {sym}")
        print()

    index = ROOT / args.docs / "re" / "00-function-index.md"
    if index.exists():
        listed = set(re.findall(r"\b(\d\d-[a-z0-9-]+\.md)\b", index.read_text(errors="replace")))
        orphans = sorted(
            d.name
            for d in (ROOT / args.docs / "re").glob("*.md")
            if d.name not in listed and not d.name.startswith("00-")
        )
        # 一份筆記完全沒提到 IDA 符號(純資料格式 / 顯示模式那類)本來就不會進索引。
        symbol_free = {
            d.name
            for d in (ROOT / args.docs / "re").glob("*.md")
            if not IDA_SYMBOL.search(d.read_text(errors="replace"))
        }
        orphans = [name for name in orphans if name not in symbol_free]
        if orphans:
            print(f"## 函式索引沒收錄的筆記({len(orphans)} 份)")
            print("   `./tools/dev.sh python3 tools/gen_func_index.py > docs/re/00-function-index.md`")
            print("   ⚠ 別把輸出丟掉 —— 它印到 stdout,不自己寫檔。\n")
            for name in orphans:
                print(f"  docs/re/{name}")
            print()

    total = len(bad_ida) + len(bad_go)
    print(f"合計 {total} 筆機械可查的過期斷言。")
    print("⚠ 全綠只代表符號還在,不代表敘述正確 —— 敘述錯誤只能靠讀碼發現。")
    return 1 if total else 0


if __name__ == "__main__":
    sys.exit(main())
