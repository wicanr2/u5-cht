# 跑 DOS 原版當 oracle 的容器(**只有原版**,重製版不在這裡跑)。
#
#   docker build -t u5cht/dosbox -f docker/dosbox.Dockerfile docker/
#
# 為什麼是 trixie 而不是 bookworm:`dosbox-x` 只在 trixie 之後才進 Debian
#(bookworm 只有 `dosbox` 0.74)。DOSBox-X 值得換一個 base ——
# 它的 `machine=ega` 與固定 `cycles` 行為在 msdostest 那 567 款上實測過
#(`~/.claude/knowledge-base/retro/dosbox-game-configs.md`)。
#
# ⚠ 這與 `tools/ida.sh` 的容器**互不相干**。`CLAUDE.md §4.5` 的
# 「不在 container 內跑遊戲」講的是 **IDA 容器**(license 唯讀掛載、
# 不出現在 log/截圖);那條規則管的是 IDA,不是禁止把原版當 oracle 跑。
FROM debian:trixie-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
      dosbox-x \
      xvfb \
      xdotool \
      x11-utils \
      imagemagick \
      python3 \
      python3-pil \
      procps \
    && rm -rf /var/lib/apt/lists/*

# ImageMagick 的預設 policy 會擋掉一部分格式;截圖只用 PNG,不需要放寬。
ENV DISPLAY=:99
