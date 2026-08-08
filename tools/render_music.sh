#!/usr/bin/env bash
# 把原版配樂全部渲染 / 轉檔成 ogg 放進 assets/audio/。
#
#   15 首 .EUP  → 4-op FM 合成(原版音色)→ WAV → ogg     tools/eup2ogg.py
#   2  條 CDDA  → 直接轉檔                                tools/cdda2ogg.sh
#
# 驗收兩層(見 docs/audio-pipeline.md §4):
#   A4  曲長要與 60/(96×BPM) 算出來的秒數相符   ← eup2ogg.py 自己檢查
#   A5b 單音的基頻要精確(8 演算法 × 3 載波MUL × 5 音高)← tools/verify_pitch.py
#       ⚠ 連**反對照**一起跑 —— 驗證器沒有鑑別力就是假綠燈(踩過三次)
#
# ⚠ 轉出來的 ogg 衍生自原版光碟,**不入 git**。
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
U5E="${U5E:-$ROOT/re_work/fmtowns/iso/U5_E}"
OUT="${1:-$ROOT/assets/audio}"
IMAGE=u5cht/dev

[[ -f "$U5E/U5_BGM.TBL" ]] || { echo "找不到 $U5E/U5_BGM.TBL" >&2; exit 1; }
mkdir -p "$OUT"

echo "== 渲染 15 首 .EUP(原版音色)"
docker run --rm --log-opt max-size=10m --log-opt max-file=3 \
  -v "$ROOT":/work -w /work "$IMAGE" \
  python3 tools/eup2ogg.py "${U5E#$ROOT/}" build/mus

echo "== 驗收音高(單音隔離:8 演算法 × 3 載波MUL × 5 音高)"
docker run --rm --log-opt max-size=10m --log-opt max-file=3 \
  -v "$ROOT":/work -w /work "$IMAGE" python3 tools/verify_pitch.py
echo "== 反對照(驗證器本身要有鑑別力)"
docker run --rm --log-opt max-size=10m --log-opt max-file=3 \
  -v "$ROOT":/work -w /work "$IMAGE" python3 tools/verify_pitch.py --control

echo "== 編成 ogg"
docker run --rm --log-opt max-size=10m --log-opt max-file=3 \
  -v "$ROOT":/work -w /work "$IMAGE" bash -euo pipefail -c '
    mkdir -p "$1"
    for w in build/mus/*.wav; do
      o="$1/$(basename "${w%.wav}").ogg"
      ffmpeg -hide_banner -loglevel error -y -i "$w" -c:a libvorbis -q:a 4 "$o"
      printf "  %s  %s\n" "$(basename "$o")" "$(du -h "$o" | cut -f1)"
    done
  ' _ "${OUT#$ROOT/}"

echo "== 兩條 CDDA"
"$ROOT/tools/cdda2ogg.sh" "$OUT" || echo "  (跳過:找不到光碟壓縮檔)"
echo "完成 → $OUT"
