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
# ⚠ 路徑要用 **repo 相對**的:指令跑在容器裡,`/home/anr2/...` 在裡面不存在。
OUT="docs/screenshots"
mkdir -p "$ROOT/$OUT"
run() { "$ROOT/tools/dev.sh" go run ./cmd/u5dump scene "$GD" "$FM" "$@"; }

run "$OUT/00-intro.png"   --script "I.."
run "$OUT/01-world.png"   --at 81 108
run "$OUT/02-town.png"    --scene BRITAIN 0 15 15
run "$OUT/03-talk.png"    --scene BRITAIN 0 30 6 --script 'T"name"'
run "$OUT/04-shop.png"    --scene BRITAIN 0 5 19 --script 'T[b]'
run "$OUT/05-combat.png"  --at 81 108 --script '!'
run "$OUT/06-dungeon.png" --at 240 73 --script "EL"
run "$OUT/07-peer.png"    --at 81 108 --script 'P'
run "$OUT/08-word.png"    --at 239 73 --script 'Y"FALLAX"'
run "$OUT/09-shadowlord.png" --scene "EMPATH ABBEY" 0 15 15 --script 'Y"ASTAROTH"'
run "$OUT/10-blackthorn.png" --at 100 100 --script 'A"no""no""no""no"'
run "$OUT/11-codex.png"   --at 100 100 --script 'R.'
run "$OUT/12-throne.png"  --at 100 100 --script 'Zy'
run "$OUT/13-cell.png"    --at 100 100 --script 'A'
run "$OUT/14-shrine-room.png" --at 128 92 --script 'E'
# 夜晚:視野縮成身邊九格,火把再撐開一圈(docs/re/31)
run "$OUT/15-night.png"       --at 81 108 --hour 1
run "$OUT/16-night-torch.png" --at 81 108 --hour 1 --script "L"
echo "→ $OUT"
