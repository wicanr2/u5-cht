#!/usr/bin/env bash
# 把已經編好的 Linux 執行檔包成 AppImage。
#
#   tools/mkappimage.sh <linux 執行檔> <輸出 .AppImage> [版本] [資料目錄]
#
# 給了**資料目錄**(第四個參數)就把它整包塞進 `usr/share/u5cht/`,
# 解出來就能玩 —— 那是 `tools/package.sh --full` 的完整版走的路,**只留本機**。
# 沒給就是交付版:**不含原版資料,也不含中文字庫**,兩者玩家自備(`CLAUDE.md §3.0`)。
#
# ⚠ 完整版少了這一段的症狀很安靜:AppImage 照樣做得出來、照樣能跑,
# 只是**沒有資料** —— 而 tar.gz 與 zip 那兩份是好的,所以三個包裡
# 只有一個是空殼,很容易到玩家手上才發現。
#
# ★ AppImage 的 cwd 是**唯讀的 squashfs 掛載點** ——
# 存檔一定要寫 `os.UserConfigDir()`(引擎早就這樣做,`internal/game/savegame.go`)。
# 這是 `retro-game-playtest` kb 記的第 2 類雷:相對 cwd 存檔在 AppImage 下必失敗。
#
# ★ 圖示用 `tools/mkicon.py` 程式畫的 ankh,**不是原版美術** ——
# 圖示一定會進交付包,所以它必須是我們自己畫的。
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BIN="${1:?用法:tools/mkappimage.sh <執行檔> <輸出.AppImage> [版本]}"
OUT="${2:?缺輸出路徑}"
VERSION="${3:-dev}"
DATADIR="${4:-}"
IMAGE="${U5_APPIMAGE_IMAGE:-u5cht/appimage}"

WORK="$(mktemp -d "$ROOT/dist/.appdir.XXXXXX")"
trap 'rm -rf "$WORK"' EXIT
APPDIR="$WORK/u5cht.AppDir"
mkdir -p "$APPDIR/usr/bin" "$APPDIR/usr/share/applications" \
         "$APPDIR/usr/share/icons/hicolor/256x256/apps"

install -m 0755 "$BIN" "$APPDIR/usr/bin/u5cht"
"$ROOT/tools/dev.sh" python3 tools/mkicon.py \
  "${APPDIR#"$ROOT/"}/usr/share/icons/hicolor/256x256/apps/u5cht.png" 256 >/dev/null
cp "$APPDIR/usr/share/icons/hicolor/256x256/apps/u5cht.png" "$APPDIR/u5cht.png"

# ⚠ `.desktop` 的 `Name` 用 UTF-8 中文沒問題(freedesktop 規定 UTF-8),
# 但**檔名一律 ASCII** —— 同 zip 的規矩(`docs/` 的交付包編碼那條)。
cat > "$APPDIR/u5cht.desktop" <<DESKTOP
[Desktop Entry]
Type=Application
Name=Ultima V (CHT)
Name[zh_TW]=創世紀 V:命運勇士 — 繁體中文
Comment=Ultima V: Warriors of Destiny — Traditional Chinese remake
Comment[zh_TW]=創世紀 V 繁體中文重製(需自備原版資料)
Exec=u5cht
Icon=u5cht
Categories=Game;RolePlaying;
Terminal=false
DESKTOP
cp "$APPDIR/u5cht.desktop" "$APPDIR/usr/share/applications/u5cht.desktop"

if [ -n "$DATADIR" ]; then
  echo "→ 把資料塞進 AppImage($DATADIR)"
  mkdir -p "$APPDIR/usr/share/u5cht"
  # ⚠ 只搬遊戲要讀的三樣,不要把整個 dist 目錄倒進去(說明檔會重複)。
  for d in gamedata assets; do
    [ -e "$DATADIR/$d" ] && cp -r "$DATADIR/$d" "$APPDIR/usr/share/u5cht/"
  done
fi

# AppRun。
#
# ★ 有內建資料時**先 `cd` 進資料根目錄**再 exec —— 引擎的預設路徑是相對的
# (`gamedata`、`assets/fonts/eten-15`),`cd` 過去就全對上,不必動旗標。
# 而 squashfs 是唯讀的沒關係:存檔走 `os.UserConfigDir()`(`savegame.go`),
# **這正是那條規則存在的理由**(u1-cht 踩過:相對 cwd 存檔在唯讀安裝下失敗)。
if [ -n "$DATADIR" ]; then
  cat > "$APPDIR/AppRun" <<'APPRUN'
#!/bin/sh
HERE="$(dirname "$(readlink -f "$0")")"
cd "$HERE/usr/share/u5cht" || exit 1
exec "$HERE/usr/bin/u5cht" "$@"
APPRUN
else
  cat > "$APPDIR/AppRun" <<'APPRUN'
#!/bin/sh
HERE="$(dirname "$(readlink -f "$0")")"
exec "$HERE/usr/bin/u5cht" "$@"
APPRUN
fi
chmod 0755 "$APPDIR/AppRun"

mkdir -p "$(dirname "$OUT")"
docker run --rm --log-opt max-size=10m --log-opt max-file=3 \
  --user "$(id -u):$(id -g)" \
  -v "$ROOT":/work -w /work -e ARCH=x86_64 -e VERSION="$VERSION" \
  "$IMAGE" appimagetool --no-appstream "${APPDIR#"$ROOT/"}" "${OUT#"$ROOT/"}"

echo "→ $OUT"
