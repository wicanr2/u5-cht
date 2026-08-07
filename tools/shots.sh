#!/usr/bin/env bash
# 重新產生 README 用的遊戲畫面(全部 headless,不需要顯示環境)。
#
#   tools/shots.sh [gamedata 目錄] [FM Towns U5_E 目錄]
#
# 畫面來自 `internal/render` 的同一條 CPU 繪製路徑,所以與實機一模一樣
#(CLAUDE.md §3.1 的硬決策:繪製不綁 GPU)。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GD="${1:-gamedata}"
FM="${2:-re_work/fmtowns/iso/U5_E}"
OUT="$ROOT/docs/screenshots"
mkdir -p "$OUT"
run() { "$ROOT/tools/dev.sh" go run ./cmd/u5dump scene "$GD" "$FM" "$@"; }

run "$OUT/01-world.png"   --at 81 108
run "$OUT/02-town.png"    --scene BRITAIN 0 15 15
run "$OUT/03-talk.png"    --scene BRITAIN 0 30 6 --script 'T"name"'
run "$OUT/04-shop.png"    --scene BRITAIN 0 5 19 --script 'T[b]'
run "$OUT/05-combat.png"  --at 81 108 --script '!'
run "$OUT/06-dungeon.png" --at 240 73 --script 'ELff'
run "$OUT/07-peer.png"    --at 81 108 --script 'P'
echo "→ $OUT"
