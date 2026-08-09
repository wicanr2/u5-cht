# 合成推廣片用的容器(ffmpeg + ImageMagick + 中文襯線字型)。
#
#   docker build -t u5cht/video -f docker/video.Dockerfile docker/
#
# ⚠ **預先建 image,不要每次跑都 apt** —— 這三包約 200 MB,
# 每次重裝會讓「調一句字幕再看一次」變成分鐘級的等待。
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
      ffmpeg imagemagick fonts-noto-cjk fonts-noto-cjk-extra \
    && rm -rf /var/lib/apt/lists/*
# ImageMagick 預設 policy 會擋 `@` 讀檔;只放行讀本地檔。
RUN sed -i 's/rights="none" pattern="@\*"/rights="read" pattern="@*"/' \
      /etc/ImageMagick-6/policy.xml || true
