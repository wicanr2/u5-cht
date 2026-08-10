#!/usr/bin/env bash
# macOS 交叉編產物的靜態驗收(`osxcross-macos-cross-build` skill §5)。
#
#   tools/verify-mac-binary.sh <universal 執行檔> <osxcross target>
#
# 四道:雙弧齊 / arm64 有 ad-hoc 簽章 / 最低系統版本 / 相依只在系統路徑。
#
# ⚠ **這不是功能驗收。** Linux 跑不了 macOS binary ——
# 全過只代表結構對,不代表遊戲跑得起來。
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BIN="${1:?用法:tools/verify-mac-binary.sh <執行檔> <target>}"
TARGET="${2:?缺 osxcross target(例:darwin24.5)}"
IMAGE="${U5_OSXCROSS_IMAGE:-u5cht/osxcross}"
T="x86_64-apple-${TARGET}"

dr() {
  docker run --rm --log-opt max-size=10m --log-opt max-file=3 \
    --user "$(id -u):$(id -g)" -v "$ROOT":/work -w /work "$IMAGE" "$@"
}
REL="${BIN#"$ROOT/"}"
fail=0
say() { printf '  %-28s %s\n' "$1" "$2"; }

echo "→ 靜態驗收 $REL"
info="$(dr "$T-lipo" -info "$REL")"
say "架構" "$info"
for a in arm64 x86_64; do
  echo "$info" | grep -q "$a" || { say "✗ 缺架構" "$a"; fail=1; }
done

for a in arm64 x86_64; do
  # ⚠ **fat binary 要先 thin 再問**:`otool` 對 fat 會為每個架構印一行檔名標頭,
  # 相依檢查會永遠指著執行檔自己。
  dr "$T-lipo" -thin "$a" "$REL" -output "$REL.$a" >/dev/null
  # ★ arm64 **必須有 `LC_CODE_SIGNATURE`** —— 沒有的話 Apple Silicon 上是
  # `Killed: 9`(x86_64 沒這個限制)。ld64 會自己加,但要驗。
  load_sig="$(dr "$T-otool" -l "$REL.$a")"
  if [ "$a" = arm64 ]; then
    if printf %s "$load_sig" | grep -q LC_CODE_SIGNATURE; then
      say "arm64 ad-hoc 簽章" "有"
    else
      say "✗ arm64 沒有簽章" "Apple Silicon 會 Killed: 9"; fail=1
    fi
  fi
  # ⚠ 不要 `| grep -m1`:grep 提早關掉管線,otool 會噴 `broken pipe` ——
  # 看起來像 otool 失敗。先整段收下來再挑。
  load="$(dr "$T-otool" -l "$REL.$a")"
  say "$a 最低系統" "$(printf %s "$load" | grep -E 'minos|version' | head -1 | tr -s ' ')"
  outside="$(dr "$T-otool" -L "$REL.$a" | tail -n +2 | awk '{print $1}' \
    | grep -vE '^(/usr/lib/|/System/Library/)' || true)"
  if [ -n "$outside" ]; then
    say "✗ $a 相依在系統外" "$outside"; fail=1
  else
    say "$a 相依" "只在 /usr/lib 與 /System/Library"
  fi
  rm -f "$BIN.$a"
done

# ★ 這一行常被忽略但很有用:在 binary 裡找**這一版必然出現的字串**,
# 證明「這份 binary 真的含有那段程式碼」—— 不需要 Mac。
# ⚠ **不要用 `strings`**:GNU `strings` 預設只印可列印 ASCII 的序列,
# UTF-8 中文的高位元組被當成不可列印 ⇒ 中文字串一個都找不到,
# 而症狀看起來像「這一版沒編進去」。fat Mach-O 它也解析不了(只好 `-a`)。
# 直接對檔案做二進位子字串搜尋最實在。
for s in "指令說明" "確定離開遊戲?" "健康"; do
  if grep -aqF -- "$s" "$BIN"; then say "含字串" "$s"; else say "✗ 找不到字串" "$s"; fail=1; fi
done

[ "$fail" = 0 ] && echo "  ✓ 四道全過(⚠ 僅結構,非功能)" || { echo "  ✗ 有項目未過"; exit 1; }
