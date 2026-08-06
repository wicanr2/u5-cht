#!/usr/bin/env bash
# 開發指令包裝:全程 docker,不裝本機工具鏈。
#
#   tools/dev.sh build      編譯
#   tools/dev.sh vet        靜態檢查
#   tools/dev.sh test       單元測試(無原版素材也應全綠)
#   tools/dev.sh itest      整合測試(對 gamedata/ 與 FM Towns 素材驗證,-v)
#   tools/dev.sh font [15|24]   烘倚天中文點陣字
#   tools/dev.sh sh         進容器 shell
#   tools/dev.sh <其他…>    直接在容器裡執行
#
# Go 的 module/build cache 落在 .gocache/(gitignore),不然每次 run 都重新下載依賴。
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${U5_DEV_IMAGE:-u5cht/dev}"
CACHE="$ROOT/.gocache"
ETEN_DIR="${U5_ETEN_DIR:-/home/anr2/cht/etan_font}"

mkdir -p "$CACHE/mod" "$CACHE/build"

dr() {
  local extra=()
  # 倚天字庫存在就唯讀掛進去(烘字型用)
  if [[ -d "$ETEN_DIR" ]]; then
    extra+=(-v "$ETEN_DIR:/eten:ro")
  fi
  docker run --rm --log-opt max-size=10m --log-opt max-file=3 \
    --user "$(id -u):$(id -g)" -e HOME=/tmp \
    -e GOMODCACHE=/gocache/mod -e GOCACHE=/gocache/build \
    -e U5_GAMEDATA="${U5_GAMEDATA:-/work/gamedata}" \
    -e U5_FMTOWNS="${U5_FMTOWNS:-/work/re_work/fmtowns/iso}" \
    -v "$CACHE:/gocache" -v "$ROOT:/work" "${extra[@]}" -w /work \
    "$IMAGE" "$@"
}

cmd="${1:-test}"
shift || true

case "$cmd" in
  build) dr go build -o /work/build/u5cht ./cmd/u5cht && echo "→ build/u5cht" ;;
  vet)   dr go vet ./... ;;
  # 單元測試刻意清掉素材路徑:確認「沒有原版資料時也全綠」這個性質沒有壞掉
  test)  dr env U5_GAMEDATA= U5_FMTOWNS= go test ./... "$@" ;;
  itest) dr go test -v ./... "$@" ;;
  font)
    size="${1:-15}"
    dr python3 tools/build_eten_font.py \
      --eten-dir /eten --iso /eten/ET353S.iso \
      --size "$size" --out "assets/fonts/eten-$size"
    ;;
  sh)    dr bash ;;
  *)     dr "$cmd" "$@" ;;
esac
