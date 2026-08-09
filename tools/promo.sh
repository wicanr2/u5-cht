#!/usr/bin/env bash
# 產生推廣短片。
#
#   tools/promo.sh
#
# 輸出 `dist-local/promo/u5cht-promo.mp4`。
#
# ⚠⚠ **成品只留本機。** 片子裡的配樂是**原版遊戲的 CD-DA 音軌**,
# 而那是他人的著作權(作曲家的旋律)—— 用原版素材是為了「音色要真」
#(`rulebook/93` 鐵則 1),但**放到會公開的地方是另一回事**(同鐵則的但書)。
# 所以輸出目錄是 `dist-local/`(已 gitignore)。
# 要對外公開之前先換成有授權的曲子,或請權利人同意。
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
OUT="dist-local/promo"
mkdir -p "$OUT"

[ -d docs/screenshots ] || { echo "✗ 沒有截圖 —— 先跑 tools/shots.sh" >&2; exit 1; }
[ -f assets/audio/CDDA2.ogg ] || {
  echo "✗ 沒有 assets/audio/CDDA2.ogg(原版光碟音軌)—— 先跑 tools/cdda2ogg.sh" >&2; exit 1; }

# ⚠ `--cpus=2`:ffmpeg 預設會吃滿所有核心,而這台機器同時有別的專案在跑。
docker run --rm --cpus=2 --log-opt max-size=10m --log-opt max-file=3 \
  --user "$(id -u):$(id -g)" \
  -v "$ROOT/docs/screenshots":/shots:ro \
  -v "$ROOT/assets/audio":/music:ro \
  -v "$ROOT/$OUT":/out \
  -v "$ROOT/tools/make_promo.sh":/make.sh:ro \
  u5cht/video bash /make.sh

echo
echo "⚠⚠ 配樂是原版光碟音軌(他人著作權)⇒ **只留本機,不要上傳**。"
echo "   要公開發布請先換成有授權的配樂。"
