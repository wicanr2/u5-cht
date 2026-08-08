#!/usr/bin/env python3
"""把「還沒讀過的反組譯函式」按**是否動到遊戲狀態**分群,產出工作優先序。

為什麼需要這支:剩下的未讀函式**大多不是遊戲邏輯** —— FM Towns 的執行期函式庫、
繪圖驅動、文字排版都混在裡面,而按「行數」或「引用字串數」排序會把它們排到前面。
這支改用「碰不碰**角色紀錄 / 物件表 / 地圖**」當判準,那三個是遊戲狀態的所在。

    ./tools/dev.sh python3 tools/triage_unread.py            # 前 40 支
    ./tools/dev.sh python3 tools/triage_unread.py --all      # 全部
    ./tools/dev.sh python3 tools/triage_unread.py --tag 角色  # 只看某一群

⚠ 「未讀」= 函式名沒出現在 `docs/` 或三份根目錄文件裡。所以**寫了筆記就算讀過** ——
這與 `check_doc_claims.py` 的口徑一致,兩支可以互相對照。

⚠⚠ 訊號是**啟發式**,不是證明。實際踩過的坑:`sub_3007C`(521 行)一個訊號都沒中,
而它是開新遊戲的腳本解譯器(`WORKLIST §5.2b`)。⇒ **分數低不等於可以跳過**;
真的要整批跳過某一群時,至少抽樣看過離群的那幾筆。
"""
import argparse
import pathlib
import re
import subprocess
import sys

ASM = "re_work/fmtowns/WORRIORS.EXP.asm"
DOCS = ("docs/", "CLAUDE.md", "CONTEXT.md", "WORKLIST.md")

# 訊號 → 正規式。前三個是「動到遊戲狀態」的判準。
SIGNALS = {
    "角色": r"byte_3DDB4|byte_3DDBF|byte_3DDC2|word_3DDC4|byte_3DDBD",
    "物件": r"dword_3E46C|sub_2B6C8|sub_2B360|sub_2B57C",
    "地圖": r"sub_DB10|byte_3E0A3|byte_3E0A5|byte_3E0A6|byte_3E0A7",
    "文字": r"sub_23C18|sub_27230|sub_23A24",
    "骰": r"sub_28E14",
    "檔IO": r"dword_5AC30|sub_2C740",
    "繪圖": r"sub_29D64|sub_297F4|sub_29BEC|sub_2B740|byte_3F8F4",
}
STATEFUL = ("角色", "物件", "地圖")


def load_functions(path: pathlib.Path) -> dict[str, tuple[int, str]]:
    """回傳 {函式名: (行數, 本體)}。"""
    lines = path.read_text(errors="replace").split("\n")
    out: dict[str, tuple[int, str]] = {}
    cur, start = None, 0
    for i, line in enumerate(lines):
        m = re.match(r"^(sub_[0-9A-Fa-f]+)\s+proc", line)
        if m:
            cur, start = m.group(1), i
        elif cur and re.match(rf"^{cur}\s+endp", line):
            out[cur] = (i - start, "\n".join(lines[start:i]))
            cur = None
    return out


def mentioned() -> set[str]:
    r = subprocess.run(
        ["grep", "-rohE", "sub_[0-9A-F]+", *DOCS], capture_output=True, text=True
    )
    return set(r.stdout.split())


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--asm", default=ASM)
    ap.add_argument("--all", action="store_true")
    ap.add_argument("--tag", help="只列含這個訊號的")
    args = ap.parse_args()

    path = pathlib.Path(args.asm)
    if not path.exists():
        print(f"⚠ 讀不到 {args.asm} —— 反組譯產物不入庫,先跑 tools/ida.sh analyze")
        return 1
    funcs = load_functions(path)
    seen = mentioned()

    rows = []
    for name, (n, body) in funcs.items():
        if name in seen:
            continue
        tags = [k for k, pat in SIGNALS.items() if re.search(pat, body)]
        rows.append((n, name, tags))
    rows.sort(reverse=True)

    core = [r for r in rows if any(t in r[2] for t in STATEFUL)]
    print(f"反組譯 {len(funcs)} 支、未讀 {len(rows)} 支;"
          f"其中**動到遊戲狀態**的 {len(core)} 支")
    print(f"(判準:碰 {' / '.join(STATEFUL)} 任一)\n")

    show = rows if args.tag else core
    if args.tag:
        show = [r for r in rows if args.tag in r[2]]
    if not args.all:
        show = show[:40]

    print(f"{'行':>5} {'函式':<12} 訊號")
    for n, name, tags in show:
        print(f"{n:>5} {name:<12} {'・'.join(tags) or '(無訊號)'}")

    rest = len(rows) - len(core)
    print(f"\n剩下 {rest} 支不碰遊戲狀態(文字排版 / 繪圖 / 檔案 / 純計算)")
    print("⚠ 分數低不等於可以跳過 —— `sub_3007C` 一個訊號都沒中,卻是開新遊戲的腳本解譯器")
    return 0


if __name__ == "__main__":
    sys.exit(main())
