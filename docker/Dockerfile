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
      python3-numpy \
      git \
      unzip \
      p7zip-full \
      ffmpeg \
      cmake \
      libsamplerate0-dev \
      libglib2.0-dev \
    && rm -rf /var/lib/apt/lists/*

# ── munt(Roland MT-32 / CM-32L 模擬器)
#
# ★ 為什麼是 munt 而不是 SoundFont:DOS 版 upgrade 的 15 首 `.XMI` 附了
# `MT32MPU.ADD` 驅動,也就是那批音樂**原本就是寫給 MT-32 的**。
# MT-32 的音色住在硬體的 CONTROL + PCM ROM 裡,而 munt 直接吃那兩顆 ROM
# ⇒ 音色來源是**真實硬體的固件**,不是第三方替代品(`CLAUDE.md §3.0`)。
#
# 只建 `mt32emu_smf2wav`(SMF → WAV 的離線轉檔器)—— 不要 GUI/ALSA 那些目標。
# ⚠ ROM **不入 image 也不入 repo**:與遊戲資料同樣是玩家自備,執行時掛載。
ARG MUNT_TAG=munt_2_8_2
RUN set -eux; \
    git clone --depth 1 --branch "$MUNT_TAG" https://github.com/munt/munt.git /tmp/munt; \
    cmake -S /tmp/munt -B /tmp/munt/build \
      -DCMAKE_BUILD_TYPE=Release \
      -Dmunt_WITH_MT32EMU_QT=FALSE \
      -Dmunt_WITH_MT32EMU_SMF2WAV=TRUE; \
    cmake --build /tmp/munt/build -j"$(nproc)"; \
    cmake --install /tmp/munt/build; \
    ldconfig; \
    rm -rf /tmp/munt

# 容器內的 Go 快取放掛載點外,避免污染 repo
ENV GOCACHE=/tmp/gocache \
    GOMODCACHE=/tmp/gomodcache \
    GOFLAGS=-buildvcs=false \
    CGO_ENABLED=1

# 軟體 GL:沒有實體 GPU 時強制走 llvmpipe
ENV LIBGL_ALWAYS_SOFTWARE=1 \
    GALLIUM_DRIVER=llvmpipe

WORKDIR /work
