#!/usr/bin/env bash
# 用 Roland MT-32 的**真實 ROM** 把 DOS upgrade 的 15 首渲染成 ogg。
#
#   MT32_ROM=/path/to/roms ./tools/mt32_render.sh [輸出目錄]
#
# ★ 為什麼是 MT-32 而不是 SoundFont:那 15 首 `.XMI` 的音色編號是 General MIDI,
#   而 upgrade 附的驅動清單裡就有 `MT32MPU.ADD`(`docs/formats/13` §4)——
#   **那批音樂原本就是寫給 MT-32 的**。MT-32 的音色住在硬體的 CONTROL + PCM ROM 裡,
#   munt 直接吃那兩顆 ROM ⇒ 音色來源是真實硬體的固件,不是第三方替代品。
#
# ⚠ ROM **玩家自備**,與遊戲資料同樣不入庫也不入 image;預設找 ~/cht/mt32。
# ⚠ 機型固定 `mt32_1_07`(1987 年的 v1.07 控制 ROM)—— U5 是 1988 年的遊戲,
#   而 CM-32L 是後來的超集。要改機型就改 MT32_MACHINE。
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="${1:-$ROOT/assets/audio/mt32}"
ROM="${MT32_ROM:-$HOME/cht/mt32}"
MACHINE="${MT32_MACHINE:-mt32_1_07}"
UPG="${UPG:-$ROOT/gamedata/upgrade}"
IMAGE=u5cht/dev

[[ -d "$ROM" ]] || { echo "找不到 MT-32 ROM 目錄:$ROM(用 MT32_ROM= 指定)" >&2; exit 1; }
shopt -s nullglob nocaseglob
XMIS=("$UPG"/*.xmi)
shopt -u nocaseglob
[[ ${#XMIS[@]} -gt 0 ]] || { echo "找不到 $UPG/*.XMI" >&2; exit 1; }
mkdir -p "$OUT"

docker run --rm --log-opt max-size=10m --log-opt max-file=3 \
  -v "$ROOT":/work -v "$ROM":/rom:ro -w /work "$IMAGE" bash -c '
set -euo pipefail
OUT="$1"; MACHINE="$2"
mkdir -p build/mid build/mt32 "$OUT"
n=0
shopt -s nullglob nocaseglob
for f in gamedata/upgrade/*.xmi; do
  b=$(basename "$f"); b="${b%.*}"
  # ⚠ `setm.xmi` 是 SETM.EXE 的測試檔,裡面有 **5 段序列** —— 不是音樂,跳過。
  # 大小寫不敏感地比:那個檔在光碟上是小寫的(而 16 首音樂是大寫)。
  case "${b,,}" in setm) continue;; esac
  python3 tools/xmi2mid.py "$f" "build/mid/$b.mid" > /dev/null
  # MT-32 的原生取樣率是 32 kHz —— smf2wav 就輸出那個率,不要在這裡重新取樣。
  mt32emu-smf2wav --quiet -m /rom -i "$MACHINE" -f -o "build/mt32/$b.wav" "build/mid/$b.mid"
  # 轉 ogg 時才一併升到 44.1 kHz(與 .EUP 那批一致,ebiten 才不必重新取樣)。
  ffmpeg -v error -y -i "build/mt32/$b.wav" -ar 44100 -c:a libvorbis -q:a 5 "$OUT/$b.ogg"
  printf "  %-12s %s\n" "$b" "$(ffprobe -v error -show_entries format=duration -of csv=p=0 "$OUT/$b.ogg") 秒"
  n=$((n+1))
done
echo "✓ $n 首 → $OUT"
' _ "${OUT#$ROOT/}" "$MACHINE"
