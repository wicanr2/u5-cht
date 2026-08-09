#!/usr/bin/env bash
# 在 Linux 上交叉編 macOS 版(arm64 + x86_64 → universal),組成 `.app` 再壓成 zip。
#
#   tools/build-mac-osxcross.sh [版本]
#
# ⚠⚠ **這條路只能做靜態驗收。** Linux 執行不了 macOS binary ——
# `tools/verify-mac-binary.sh` 全過只代表**不會因結構問題開不起來**
#(架構齊、最低系統版本對、沒有連到編譯機才有的路徑、arm64 有 ad-hoc 簽章),
# **不代表功能正常**。要真的驗功能、要過 Gatekeeper(notarization),都得有 Mac。
#
# ⚠ **bundle 不簽章。** `_CodeSignature` 在 Linux 上做不出來 ——
# 「未簽」勝過「壞簽」:壞簽直接被拒絕,未簽只是要玩家右鍵 → 打開一次。
# 說明寫在交付包的 README 裡。
#
# 對照組:CI 的 macos runner 原生編(`.github/workflows/`)。
# ★ 兩支腳本的旗標要**逐項對齊,改一邊就改另一邊** —— 否則兩個平台的產物
# 會悄悄長得不一樣,而那種差異只有玩家會撞到。
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
VERSION="${1:-$(git describe --tags --always --dirty 2>/dev/null || echo dev)}"
IMAGE="${U5_OSXCROSS_IMAGE:-u5cht/osxcross}"
OUT="dist/macos"
APP="$OUT/Ultima V CHT.app"

dr() {
  docker run --rm --log-opt max-size=10m --log-opt max-file=3 \
    --user "$(id -u):$(id -g)" \
    -v "$ROOT":/work -w /work \
    -v "$ROOT/.gocache":/gocache -e GOCACHE=/gocache/build -e GOMODCACHE=/gocache/mod \
    -e HOME=/tmp "$IMAGE" "$@"
}

rm -rf "$OUT"
mkdir -p "$APP/Contents/MacOS" "$APP/Contents/Resources" .gocache

# ★ 前綴帶 **SDK 次版號**(15.5 → darwin24.5)。不要寫死也不要憑印象 ——
# 問工具鏈自己。寫死成 `darwin24` 的症狀是
# 「unable to execute command: No such file or directory」,看起來像 clang 壞了。
TARGET="$(dr sh -c 'osxcross-conf >/dev/null 2>&1; . <(osxcross-conf); echo $OSXCROSS_TARGET' 2>/dev/null || true)"
if [ -z "$TARGET" ]; then
  TARGET="$(dr sh -c 'osxcross-conf | sed -n "s/^export OSXCROSS_TARGET=//p"')"
fi
TARGET="$(printf %s "$TARGET" | tr -d '\r\n')"
echo "→ osxcross target: $TARGET"

LDFLAGS="-s -w -X main.version=${VERSION}"
# ⚠ **每弧各編一次**,不要靠單次雙 `-arch`。Go 也不支援 fat 輸出。
for pair in "arm64:arm64" "amd64:x86_64"; do
  goarch="${pair%%:*}"; darch="${pair##*:}"
  echo "→ 編譯 darwin/$goarch(CC=${darch}-apple-${TARGET}-clang)"
  dr env GOOS=darwin GOARCH="$goarch" CGO_ENABLED=1 \
      CC="${darch}-apple-${TARGET}-clang" CXX="${darch}-apple-${TARGET}-clang++" \
      MACOSX_DEPLOYMENT_TARGET=10.13 \
      go build -trimpath -ldflags "$LDFLAGS" -o "$OUT/u5cht-$darch" ./cmd/u5cht
done

echo "→ lipo 併成 universal"
dr "x86_64-apple-${TARGET}-lipo" -create \
  "$OUT/u5cht-arm64" "$OUT/u5cht-x86_64" -output "$APP/Contents/MacOS/u5cht"
rm -f "$OUT/u5cht-arm64" "$OUT/u5cht-x86_64"

# 圖示:程式畫的 ankh,**不是原版美術**(圖示一定會進交付包)。
tools/dev.sh python3 tools/mkicon.py "${APP#"$ROOT/"}/Contents/Resources/u5cht.png" 512 >/dev/null

cat > "$APP/Contents/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleName</key>              <string>Ultima V CHT</string>
  <key>CFBundleDisplayName</key>       <string>創世紀 V:命運勇士</string>
  <key>CFBundleIdentifier</key>        <string>tw.wicanr2.u5cht</string>
  <key>CFBundleVersion</key>           <string>${VERSION}</string>
  <key>CFBundleShortVersionString</key><string>${VERSION}</string>
  <key>CFBundleExecutable</key>        <string>u5cht</string>
  <key>CFBundlePackageType</key>       <string>APPL</string>
  <key>LSMinimumSystemVersion</key>    <string>10.13</string>
  <key>NSHighResolutionCapable</key>   <true/>
</dict>
</plist>
PLIST

tools/verify-mac-binary.sh "$APP/Contents/MacOS/u5cht" "$TARGET"

tools/dev.sh python3 tools/mkreadme.py "${OUT#"$ROOT/"}" "$VERSION" macos
cp LICENSE "$OUT/LICENSE.txt"
( cd "$OUT" && zip -qr "../u5cht-${VERSION}-macos-universal.zip" . )
echo "→ dist/u5cht-${VERSION}-macos-universal.zip"
