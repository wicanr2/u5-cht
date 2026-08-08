#!/usr/bin/env bash
# 把 FM Towns 光碟的兩條 CDDA 音軌轉成 ogg。
#
# CDDA 是 raw PCM:44,100 Hz、16-bit、立體聲、little-endian、每秒 176,400 byte。
# ⇒ 音軌長度可以直接算出來當驗收:
#     Track 2  31,752,000 / 176,400 = 180 秒
#     Track 3  65,444,400 / 176,400 = 371 秒
#
# ★ 這條路**不需要任何 FM 知識** —— 與 `.EUP` 卡住的那兩個問題(音色參數語意、
#   tick → 秒)完全無關(`docs/re/89`)。所以先做。
#
# 用法(全程 docker,`CLAUDE.md §3`):
#   tools/cdda2ogg.sh [輸出目錄]        # 預設 assets/audio
#
# ⚠ 轉出來的 ogg 衍生自原版光碟,**不入 git**(`.gitignore` 已含 assets/audio)。
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="${1:-$ROOT/assets/audio}"
ARCHIVE="$ROOT/org_game/fmtown/Ultima V - Warriors of Destiny (Japan).7z"
IMAGE=u5cht/dev

# CDDA 的取樣參數 —— 紅皮書規格,不是猜的。
RATE=44100
BYTES_PER_SEC=176400

if [[ ! -f "$ARCHIVE" ]]; then
  echo "找不到光碟壓縮檔:$ARCHIVE" >&2
  echo "(原版媒體不入庫,請自備一份合法副本放到 org_game/fmtown/)" >&2
  exit 1
fi

mkdir -p "$OUT"

docker run --rm --log-opt max-size=10m --log-opt max-file=3 \
  -v "$ROOT":/work -v "$ROOT/org_game/fmtown":/g:ro -w /tmp "$IMAGE" \
  bash -euo pipefail -c '
    OUT="$1"; RATE="$2"; BPS="$3"
    7z x -y "/g/Ultima V - Warriors of Destiny (Japan).7z" \
        "*(Track 2).bin" "*(Track 3).bin" >/dev/null
    for t in 2 3; do
      src=$(ls *"(Track $t).bin")
      size=$(stat -c%s "$src")
      secs=$(( size / BPS ))
      rem=$(( size % BPS ))
      printf "Track %s: %s byte = %s 秒" "$t" "$size" "$secs"
      if [ "$rem" -ne 0 ]; then
        # ⚠ 除不盡就是「這不是純 CDDA」或抽檔抽錯了 —— 不要默默繼續。
        printf "  ⚠ 餘 %s byte(不是 2352 的整數倍?)" "$rem"
      fi
      echo
      ffmpeg -hide_banner -loglevel error -y \
        -f s16le -ar "$RATE" -ac 2 -i "$src" \
        -c:a libvorbis -q:a 5 "$OUT/CDDA$t.ogg"
      # 驗收:轉出來的長度要對得上算出來的秒數(容許 1 秒誤差)。
      got=$(ffprobe -v error -show_entries format=duration -of csv=p=0 "$OUT/CDDA$t.ogg")
      got=${got%.*}
      if [ "$((got > secs ? got - secs : secs - got))" -gt 1 ]; then
        echo "  ✗ ogg 長度 ${got}s 與原始 ${secs}s 差太多" >&2
        exit 1
      fi
      echo "  ✓ $OUT/CDDA$t.ogg  ${got}s"
      rm -f "$src"
    done
  ' _ "/work/${OUT#$ROOT/}" "$RATE" "$BYTES_PER_SEC"
