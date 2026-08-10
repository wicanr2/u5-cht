#!/usr/bin/env bash
# 打交付包(Linux tar.gz + Windows zip)。全程 docker,不污染主機。
#
#   tools/package.sh [版本]            交付版(**不含**原版資料與中文字庫)
#   tools/package.sh [版本] --full     完整版(**含**資料,只留本機)
#
# 產出在 dist/。⚠ **不含原版資料,也不含烘好的中文字庫** ——
# 前者是玩家自備的合法副本,後者衍生自 1993 年的商業字型。兩者都不散布。
#
# 產出三種:
#
#	Linux    `.tar.gz`(解壓即用)+ `.AppImage`(單檔,不必解壓)
#	Windows  `.zip`(檔名全 ASCII + UTF-8 旗標;`.bat` CP950、`.txt` UTF-8 BOM)
#	macOS    另一條路 —— 見 `tools/build-mac-osxcross.sh` 與 CI 的原生 runner。
#	         ⚠ 交叉編出來的**只能靜態驗收**(Linux 跑不了 macOS binary),
#	         而要過 Gatekeeper(notarization)一定要真的 Mac。
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
VERSION="${1:-$(git describe --tags --always --dirty 2>/dev/null || echo dev)}"
# ── 完整版 ────────────────────────────────────────────────────────────
# `--full` 把**原版資料 + 烘好的中文字庫 + 轉出的配樂**一起打進去,解壓即玩。
#
# ⚠⚠ **完整版只留本機,不上傳、不散布、不進 release。**
# 裡面有:原版遊戲資料(EA / Origin 的著作)、由 1993 年商業字庫烘出的 atlas、
# 從原版媒體轉出的 ogg —— 三者都不是我們能散布的東西(`CLAUDE.md §3.0` / §7.3)。
# 所以它寫到 `dist-local/`(已 gitignore),而且每個包裡放一張「不要散布」的字條。
FULL=0
for a in "$@"; do [ "$a" = "--full" ] && FULL=1; done
DIST="dist"
[ "$FULL" = 1 ] && DIST="dist-local"
LDFLAGS="-s -w -X main.version=${VERSION}"

# ⚠ **只清自己的產物,不要 `rm -rf "$DIST"`。**
# `--full` 的輸出目錄是 `dist-local/`,而那裡還放著別的東西
#(`promo/` 的推廣片、之前的成品)—— 整個砍掉會連它們一起吃掉。
# 這條是實際踩到才加的:一次 `--full` 就把 19 MB 的推廣片刪了,
# 而腳本一聲不響、退出碼 0。
for sub in linux windows macos; do rm -rf "$DIST/$sub"; done
rm -f "$DIST"/u5cht-*.tar.gz "$DIST"/u5cht-*.zip "$DIST"/u5cht-*.AppImage
mkdir -p "$DIST"

# 1) Linux(需要 CGO:ebiten 要 X11 / GL)
echo "→ 編譯 linux/amd64"
tools/dev.sh env CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
  go build -trimpath -ldflags "$LDFLAGS" -o "$DIST/linux/u5cht" ./cmd/u5cht

# 2) Windows(純 Go 交叉編譯 —— ebiten 在 Windows 走 DirectX / OpenGL,不需要 cgo)
echo "→ 編譯 windows/amd64"
tools/dev.sh env CGO_ENABLED=0 GOOS=windows GOARCH=amd64 \
  go build -trimpath -ldflags "$LDFLAGS" -o "$DIST/windows/u5cht.exe" ./cmd/u5cht

# 3) 說明檔與啟動器(編碼規矩見 mkreadme.py)
tools/dev.sh python3 tools/mkreadme.py "$DIST/linux" "$VERSION" linux
tools/dev.sh python3 tools/mkreadme.py "$DIST/windows" "$VERSION" windows
cp LICENSE "$DIST/linux/LICENSE.txt"
cp LICENSE "$DIST/windows/LICENSE.txt"

# 3.5) 完整版:把玩家該自備的三樣東西補進去。
#
# ⚠ 這一段**只在 `--full`** 跑。少了這道判斷,交付版就會夾帶版權素材 ——
# 而那種錯誤在本機測起來完全正常(檔案在就是能玩),只有上傳之後才看得出來。
if [ "$FULL" = 1 ]; then
  for plat in linux windows; do
    echo "→ 完整版:把資料放進 $plat"
    mkdir -p "$DIST/$plat/gamedata" "$DIST/$plat/assets/fonts" "$DIST/$plat/assets/audio"
    cp -r gamedata/. "$DIST/$plat/gamedata/"
    # 字庫 atlas 與譯文用的 JSON(由 tools/build_eten_font.py 從自備字庫烘出)。
    cp assets/fonts/*.png assets/fonts/*.json "$DIST/$plat/assets/fonts/" 2>/dev/null || true
    cp -r assets/audio/. "$DIST/$plat/assets/audio/" 2>/dev/null || true
    cat > "$DIST/$plat/DO-NOT-REDISTRIBUTE.txt" <<'WARN'
This package contains ORIGINAL GAME DATA, a bitmap font baked from a
commercial 1993 font library, and music rendered from the original media.

*** LOCAL USE ONLY. DO NOT UPLOAD, SHARE OR PUBLISH THIS PACKAGE. ***

The redistributable build is the one WITHOUT gamedata/ and assets/ —
built by tools/package.sh with no --full flag.
WARN
    tools/dev.sh python3 - "$DIST/$plat" <<'PYW'
import sys, pathlib
# 繁中版的字條:給 Windows 的存 CP950、其餘 UTF-8 with BOM(同交付包的編碼規矩)。
d = pathlib.Path(sys.argv[1])
text = (
    "這一份是**完整版**:裡面含原版遊戲資料、由商業字庫烘出的中文點陣字、"
    "以及從原版媒體轉出的配樂。\r\n\r\n"
    "*** 只供本機自用。請勿上傳、分享或公開散布。 ***\r\n\r\n"
    "可以散布的版本是**不含** gamedata/ 與 assets/ 的那一份"
    "(tools/package.sh 不加 --full)。\r\n"
)
enc = "cp950" if d.name == "windows" else "utf-8-sig"
(d / "請勿散布.txt").write_text(text, encoding=enc, errors="replace")
PYW
  done
fi

# 4) 壓縮
#    tar.gz 沒有檔名編碼問題(tar 的檔名就是一串 bytes,原樣還原);
#    zip 有,所以走 mkzip.py 檢查檔名全為 ASCII 並確認 UTF-8 旗標。
NAME="u5cht-${VERSION}"
tar -C "$DIST/linux" -czf "$DIST/${NAME}-linux-amd64.tar.gz" .
# AppImage:單檔、不必解壓。⚠ 它的 cwd 是**唯讀的 squashfs** ——
# 存檔寫 `os.UserConfigDir()` 這條在這裡才真的被考到(`retro-game-playtest` 第 2 類雷)。
# ⚠ 完整版要把資料一起塞進 AppImage —— 否則三個包裡只有它是空殼。
if [ "$FULL" = 1 ]; then
  tools/mkappimage.sh "$DIST/linux/u5cht" "$DIST/${NAME}-x86_64.AppImage" "$VERSION" "$DIST/linux"
else
  tools/mkappimage.sh "$DIST/linux/u5cht" "$DIST/${NAME}-x86_64.AppImage" "$VERSION"
fi
# ⚠ 走 dev.sh 是為了帶 --user —— 直接 docker run 產出的檔會是 root 所有,
# 而 CI 後續要上傳它。
tools/dev.sh python3 tools/mkzip.py "$DIST/${NAME}-windows-amd64.zip" "$DIST/windows"
tools/dev.sh python3 tools/checkpkg.py "$DIST/${NAME}-windows-amd64.zip"

echo
ls -lh "$DIST"/*.tar.gz "$DIST"/*.zip "$DIST"/*.AppImage
echo
if [ "$FULL" = 1 ]; then
  echo "⚠⚠ 這是**完整版**(含原版資料 / 字庫 / 配樂)—— 放在 $DIST/,**不要上傳**。"
else
  echo "⚠ 包裡沒有原版資料,也沒有中文字庫 —— 兩者都要玩家自備(README-CHT.txt 有說明)。"
fi
