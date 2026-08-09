#!/usr/bin/env python3
"""用**旋律**把 DOS upgrade 的 XMI 對上 FM Towns 的 `.EUP`(⇒ 曲號)。

為什麼需要這支:MT-32 那批是**曲名**檔名(`U5THEME.XMI`),而遊戲裡的配樂是
**曲號**(`U5_BGM.TBL` 的 15 列 → `M1.EUP` …)。要讓玩家能切換音源就得知道
「曲號 N 對應哪一首 XMI」,而那個對應**不在任何資料檔裡** ——
它在 upgrade 改寫進 `ULTIMA.EXE` 的程式碼裡。

⇒ 改用**一手資料**:兩邊是同一批曲子,所以拿旋律比。判準是
**音高差分序列的最長共同子串**(差分 = 不受移調影響)。

    ./tools/dev.sh python3 tools/match_songs.py

## ⚠ 兩個一開始沒想到、少了就會判錯的地方

1. **分數要對照曲子長度讀,不能看絕對值。** `M14.EUP` 每聲道只有 ~22 個音,
   所以它拿到 21 已經是**整條旋律都一樣**;而用「≥12 且是次佳兩倍」這種
   絕對門檻去看,它會被判成「沒對上」。
2. **同一句樂句會在兩首曲子裡出現。** `REUNION`(短的重奏)與 `RULEBRIT`
   共用開頭那句,所以 `REUNION×M14`、`REUNION×M152`、`RULEBRIT×M14`
   **三個組合都拿 21** —— 旋律這一個訊號**分不開它們**。
   分開它們的是**音數與聲道輪廓**(REUNION 128 音/6 聲道 ⇄ M14 126 音/6 聲道;
   RULEBRIT 460 音/8 聲道 ⇄ M152 397 音/6 聲道)。
   ⇒ 平手時用音數差當第二判準,而不是硬選一個。

⚠ 這支只**產生候選與分數**,不自動下定論。配不上的一律留白 ——
猜錯的配樂比沒有配樂糟(玩家會以為引擎壞了)。
"""
import glob
import os
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
# ⚠ **不要**在這裡自己再寫一份 `.EUP` / XMI 解碼器 —— 兩份會漂移,
# 而漂移的症狀是「對不上」,看起來像結論而其實是 bug。
from eup2ogg import parse_eup  # noqa: E402
from xmi2mid import find_evnt, parse  # noqa: E402

U5E = "re_work/fmtowns/iso/U5_E"
UPG = "gamedata/upgrade"

# 比對的最短長度:太短的共同片段是巧合(音階、琶音到處都有)。
MIN_RUN = 12


def eup_tracks(path: str) -> dict[int, list[int]]:
    """把一個 `.EUP` 解成 {聲道: 音高序列}(共用 `eup2ogg.parse_eup`)。"""
    notes, _programs, _tempo, _bars = parse_eup(path)
    per: dict[int, list[int]] = {}
    for ch, _tick, note, _vel in notes:
        per.setdefault(ch, []).append(note)
    return per


def xmi_track(path: str) -> dict[int, list[int]]:
    raw = open(path, "rb").read()
    _, evnt = find_evnt(raw)
    per: dict[int, list[int]] = {}
    for _t, _o, d in parse(evnt):
        if d[0] & 0xF0 == 0x90 and len(d) > 2 and d[2] != 0:
            per.setdefault(d[0] & 0x0F, []).append(d[1])
    return per


def diffs(seq: list[int]) -> list[int]:
    """音高差分 —— 移調之後仍然相同。"""
    return [b - a for a, b in zip(seq, seq[1:])]


def longest_common_run(a: list[int], b: list[int]) -> int:
    """最長共同**連續**子串的長度(動態規劃,只留一列)。"""
    if not a or not b:
        return 0
    prev = [0] * (len(b) + 1)
    best = 0
    for x in a:
        cur = [0] * (len(b) + 1)
        for j, y in enumerate(b, 1):
            if x == y:
                cur[j] = prev[j - 1] + 1
                if cur[j] > best:
                    best = cur[j]
        prev = cur
    return best


def best_match(xmi: dict[int, list[int]], eup: dict[int, list[int]]) -> int:
    """兩首曲子之間最長的共同差分串(取所有聲道兩兩比對的最大值)。"""
    best = 0
    for xs in xmi.values():
        dx = diffs(xs)
        if len(dx) < MIN_RUN:
            continue
        for es in eup.values():
            de = diffs(es)
            if len(de) < MIN_RUN:
                continue
            best = max(best, longest_common_run(dx, de))
    return best


def rank(row: dict[str, int]) -> tuple[tuple[int, str], tuple[int, str], bool]:
    """把一列分數排成 (最佳, 次佳, 算不算對上)。"""
    scores = sorted(((v, k) for k, v in row.items()), reverse=True)
    top = scores[0] if scores else (0, "-")
    second = scores[1] if len(scores) > 1 else (0, "-")
    # 兩個條件都要:絕對長度夠(不是巧合)+ 明顯甩開次佳(不是兩首都像)。
    return top, second, top[0] >= MIN_RUN and top[0] >= second[0] * 2


def show(title: str, rows: dict[str, dict[str, int]]) -> dict[str, str]:
    print(f"\n── {title}")
    print(f"{'':<14}{'最佳':<12}{'分數':>6}   次佳(分數)")
    out: dict[str, str] = {}
    for name, row in rows.items():
        top, second, ok = rank(row)
        if ok:
            out[name] = top[1]
        print(
            f"{name:<14}{top[1]:<12}{top[0]:>6}   "
            f"{second[1]}({second[0]}){'  ✓' if ok else ''}"
        )
    return out


def main() -> int:
    eups = sorted(glob.glob(os.path.join(U5E, "M*.EUP")))
    xmis = [
        f
        for f in sorted(glob.glob(os.path.join(UPG, "*.[xX][mM][iI]")))
        if os.path.basename(f).lower() != "setm.xmi"
    ]
    if not eups or not xmis:
        print("⚠ 缺 .EUP 或 .XMI —— 原版資料玩家自備")
        return 1

    ev = {os.path.basename(f): eup_tracks(f) for f in eups}
    xv = {os.path.basename(f): xmi_track(f) for f in xmis}
    print(f".EUP {len(ev)} 首、XMI {len(xv)} 首 → 比 {len(ev) * len(xv)} 對")

    # 整張分數矩陣算一次,正反兩個方向共用。
    m = {x: {e: best_match(xt, et) for e, et in ev.items()} for x, xt in xv.items()}

    fwd = show("XMI → 最像的 .EUP", m)
    rev = show(
        ".EUP → 最像的 XMI(反對照)",
        {e: {x: m[x][e] for x in xv} for e in ev},
    )

    # ★ 只採信**雙向都指向對方**的配對 —— 單向最佳可能是「兩首 XMI 搶同一首 EUP」。
    pairs = {x: e for x, e in fwd.items() if rev.get(e) == x}
    print(f"\n① 雙向一致的配對:{len(pairs)} 組")
    for x, e in sorted(pairs.items()):
        print(f"  {x:<14}⇄ {e:<10}{m[x][e]:>5}")

    # ② 剩下的用**音數**收尾。旋律分不開的(共用樂句)由曲子規模分開。
    xn = {x: sum(len(s) for s in t.values()) for x, t in xv.items()}
    en = {e: sum(len(s) for s in t.values()) for e, t in ev.items()}
    left_x = [x for x in xv if x not in pairs]
    left_e = [e for e in ev if e not in pairs.values()]
    if left_x and left_e:
        print("\n② 旋律平手的殘局 —— 用音數與聲道數收尾")
        for x in left_x:
            for e in left_e:
                dn = abs(xn[x] - en[e])
                # 音數差在 5% 之內、聲道數相同 ⇒ 同一首的不同編曲。
                same = dn <= max(4, en[e] // 20) and len(xv[x]) == len(ev[e])
                print(
                    f"  {x:<14}× {e:<10}旋律{m[x][e]:>4}  "
                    f"音數 {xn[x]}/{en[e]}(差{dn})  "
                    f"聲道 {len(xv[x])}/{len(ev[e])}{'  ✓ 同一首' if same else ''}"
                )
                if same:
                    pairs[x] = e

    print(f"\n③ 最終配對 {len(pairs)}/{len(ev)} 首 `.EUP`")
    lost_x = [x for x in xv if x not in pairs]
    print(f"沒配到的 XMI:{lost_x or '無'}")
    print(f"沒配到的 .EUP:{[e for e in ev if e not in pairs.values()] or '無'}")

    # 落單的 XMI 可能是另一首 XMI 的改編版 —— 那表示它不是漏配,
    # 而是 FM Towns 那邊根本沒收這首(⇒ 它沒有曲號)。
    for x in lost_x:
        row = {o: best_match(xv[x], xv[o]) for o in xv if o != x}
        top, second, _ok = rank(row)
        print(
            f"  {x} 對其他 XMI 最像:{top[1]}({top[0]})、{second[1]}({second[0]})"
            f" ⇒ {'也不是改編版,是獨立曲目' if top[0] < MIN_RUN else '疑為改編版'}"
        )

    print("\n⚠ 配不上的一律留白 —— 猜錯的配樂比沒有配樂糟")
    return 0


if __name__ == "__main__":
    sys.exit(main())
