#!/usr/bin/env bash
# IDA Pro 9.4 headless 包裝(全程 docker,不裝本機)。
#
#   tools/ida.sh analyze <binary> [額外參數…]   產 .i64 + .asm
#   tools/ida.sh idc <binary.i64> <script.idc>  跑 IDC 腳本
#   tools/ida.sh raw <idat 參數…>               直接下 idat 參數
#
# 工作目錄固定 re_work/(不入 git);tools/ 以唯讀掛進容器供 IDC 取用。
#
# 鐵則(見 ~/.claude/knowledge-base/retro/ida-pro-9.4.md):
#   - 寫 IDC 不要寫 IDAPython(本機實測 IDAPython 無輸出)
#   - IDC 一定要 #include <idc.idc>,少了會安靜 exit 1
#   - headless 的 print/Message() 看不到 → 結果一律寫檔到 /work
#   - 不要 grep .asm 找位址,要查 xref 圖(XrefType())
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORK="${U5_IDA_WORK:-$ROOT/re_work}"
IMAGE="${U5_IDA_IMAGE:-ida-pro-9.4-ver2}"

mkdir -p "$WORK"

run() {
  docker run --rm --log-opt max-size=10m --log-opt max-file=3 \
    -v "$WORK:/work" -v "$ROOT/tools:/work/tools:ro" -w /work \
    "$IMAGE" "$@"
}

cmd="${1:-}"
shift || true

case "$cmd" in
  analyze)
    bin="${1:?用法: tools/ida.sh analyze <binary>}"; shift
    echo "[ida] analyze $bin (工作目錄 $WORK)"
    run idat -A -B "$bin" "$@"
    ;;
  idc)
    db="${1:?用法: tools/ida.sh idc <binary.i64> <script.idc> [腳本參數]}"
    script="${2:?缺 IDC 腳本}"; shift 2
    # 腳本以 tools/ 唯讀掛載,故容器內路徑固定 /work/tools/
    run idat -A "-S/work/tools/$(basename "$script") $*" "$db"
    ;;
  raw)
    run "$@"
    ;;
  *)
    sed -n '2,20p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//'
    exit 1
    ;;
esac
