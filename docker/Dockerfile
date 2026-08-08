# Ultima V 重製專案的建置/測試容器(Go + Ebitengine)。
#
#   docker build -t u5cht/dev docker/
#   docker run --rm --log-opt max-size=10m --log-opt max-file=3 \
#     -v "$PWD":/work -w /work u5cht/dev go build ./...
#
# 這是「工具容器」:程式碼用掛載進來,不 COPY —— 改一行不必重 build image。
FROM golang:1.24-bookworm

# Ebitengine 在 Linux 走 CGO,需要 GL 與 X11 開發標頭。
# xvfb + 軟體 GL(llvmpipe)供 headless 截圖測試用(P1.5)。
RUN apt-get update && apt-get install -y --no-install-recommends \
      pkg-config \
      libgl1-mesa-dev \
      libxrandr-dev \
      libxcursor-dev \
      libxinerama-dev \
      libxi-dev \
      libxxf86vm-dev \
      libasound2-dev \
      xvfb \
      libgl1-mesa-dri \
      mesa-utils \
      x11-utils \
      xauth \
      imagemagick \
      python3 \
      python3-pil \
      git \
      unzip \
      p7zip-full \
      ffmpeg \
    && rm -rf /var/lib/apt/lists/*

# 容器內的 Go 快取放掛載點外,避免污染 repo
ENV GOCACHE=/tmp/gocache \
    GOMODCACHE=/tmp/gomodcache \
    GOFLAGS=-buildvcs=false \
    CGO_ENABLED=1

# 軟體 GL:沒有實體 GPU 時強制走 llvmpipe
ENV LIBGL_ALWAYS_SOFTWARE=1 \
    GALLIUM_DRIVER=llvmpipe

WORKDIR /work
