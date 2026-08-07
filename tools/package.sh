#!/usr/bin/env bash
# 打交付包(Linux tar.gz + Windows zip)。全程 docker,不污染主機。
#
#   tools/package.sh [版本]
#
# 產出在 dist/。⚠ **不含原版資料,也不含烘好的中文字庫** ——
# 前者是玩家自備的合法副本,後者衍生自 1993 年的商業字型。兩者都不散布。
#
# macOS 的包不在這裡做:那需要原生 runner(見 .github/workflows/release.yml)。
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
VERSION="${1:-$(git describe --tags --always --dirty 2>/dev/null || echo dev)}"
DIST="dist"
LDFLAGS="-s -w -X main.version=${VERSION}"

rm -rf "$DIST"
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

# 4) 壓縮
#    tar.gz 沒有檔名編碼問題(tar 的檔名就是一串 bytes,原樣還原);
#    zip 有,所以走 mkzip.py 檢查檔名全為 ASCII 並確認 UTF-8 旗標。
NAME="u5cht-${VERSION}"
tar -C "$DIST/linux" -czf "$DIST/${NAME}-linux-amd64.tar.gz" .
# ⚠ 走 dev.sh 是為了帶 --user —— 直接 docker run 產出的檔會是 root 所有,
# 而 CI 後續要上傳它。
tools/dev.sh python3 tools/mkzip.py "$DIST/${NAME}-windows-amd64.zip" "$DIST/windows"
tools/dev.sh python3 tools/checkpkg.py "$DIST/${NAME}-windows-amd64.zip"

echo
ls -lh "$DIST"/*.tar.gz "$DIST"/*.zip
echo
echo "⚠ 包裡沒有原版資料,也沒有中文字庫 —— 兩者都要玩家自備(README-CHT.txt 有說明)。"
