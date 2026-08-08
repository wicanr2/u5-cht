#!/usr/bin/env python3
"""盤點 docs/ 裡的斷言還成不成立。

三種機械可查的過期方式,對應三個檢查:

1. **不存在的函式**:筆記寫 `sub_XXXX`,但反組譯檔裡沒有這支。打錯字或憑記憶寫的
   位址 —— 而位址錯了,整條溯源就斷了。
2. **不存在的引擎符號**:「引擎對應」表格寫 `game.(*State).foo`,但那個方法已經
   改名或刪掉。`rulebook/63` 的原話:code 是唯一真相,dated 文件會過期。
3. **孤立的筆記**:`docs/re/00-function-index.md` 沒收錄的筆記檔 —— 索引本身過期。

⚠⚠ **只查函式,不查資料位址。** 這是 2026-08-08 首跑之後的決定,理由是命中率:
`byte_`/`word_`/`dword_`/`off_` 那類報了 **33 筆,其中 32 筆是誤報** ——
因為程式碼常寫成 `mov eax, offset dword_4FFB8` / `[eax+edi+42B0h]`,被讀的
`0x54268` 從頭到尾沒有符號,筆記用位址稱呼它是**合理且必要**的。
函式沒有這個問題:能成為函式就有 xref,IDA 必然建 `sub_`。

第一版的做法是「開一份白名單登記推導位址,再放寬幾個中文關鍵詞」——
那讓報告變綠,但**寬鬆的關鍵詞會整行跳過檢查**(任何含「應為」的行都不查),
等於用假訊號換綠燈。改成按符號種類切,不需要白名單也不需要關鍵詞豁免。

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
IDA_FUNC = re.compile(r"\b(sub_[0-9A-F]{1,8})\b")

# 索引收錄判斷仍要看「這份筆記有沒有提到任何 IDA 符號」,所以保留寬版樣式,
# 但**不拿它做存在性檢查**(見上面模組說明的理由)。
IDA_ANY = re.compile(r"\b((?:sub|byte|word|dword|qword|off|unk|nullsub)_[0-9A-F]{1,8})\b")

# 「引擎對應」表格裡的 Go 符號:`game.(*State).foo`、`u5data.Foo`、`game.Foo`。
# 結尾的 `*`(如 `u5data.NPCAI*`)是刻意的「這一族常數」寫法,不是單一符號。
GO_SYMBOL = re.compile(r"\b(game|u5data|i18n|render|cjk|audio)\.(?:\(\*(\w+)\)\.)?(\w+)\b(?!\*)")

# 產生的檔案不查:`00-function-index.md` 的欄位會**截斷在符號中間**
# (`sub_1DA10` → `sub_1DA1…`),那是產生器的排版,不是筆記寫錯位址。
GENERATED_DOCS = {"00-function-index.md"}

# 筆記裡刻意保留的「已被推翻」段落 —— 這些會提到已經作廢的函式名。
# ⚠ 只留「明確在說這條已作廢」的四個詞。不要為了讓報告變綠而加寬 ——
# 每加一個詞就多一批**整行不檢查**的筆記。
OVERTURNED_MARKERS = ("已被推翻", "此前這裡寫", "原本寫", "~~")


def load_asm_symbols(asm: pathlib.Path) -> set[str]:
    """反組譯檔裡出現過的所有 IDA 符號。"""
    if not asm.exists():
        return set()
    text = asm.read_text(errors="replace")
    return set(IDA_FUNC.findall(text))


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

    if not asm_symbols:
        print(f"⚠ 讀不到 {args.asm} —— 跳過 IDA 符號檢查(反組譯產物不入庫)")
    print(f"反組譯函式 {len(asm_symbols)} 支、引擎符號 {len(go_symbols)} 個、筆記 {len(docs)} 份\n")

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
                for sym in IDA_FUNC.findall(line):
                    if sym not in asm_symbols:
                        bad_ida.append((str(rel), n, sym))
            for _pkg, _recv, name in GO_SYMBOL.findall(line):
                if name and name not in go_symbols:
                    bad_go.append((str(rel), n, name))

    if bad_ida:
        print(f"## 反組譯檔裡查不到的函式({len(bad_ida)} 筆)")
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
            if not IDA_ANY.search(d.read_text(errors="replace"))
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
