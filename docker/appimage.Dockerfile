# 打 AppImage 用的容器。
#
#   docker build -t u5cht/appimage -f docker/appimage.Dockerfile docker/
#
# ⚠ `appimagetool` 本身是 AppImage,而 AppImage 預設要 FUSE ——
# 容器裡通常沒有(需要 `--privileged` 或 `--device /dev/fuse`,兩者都不該要)。
# 所以**先 `--appimage-extract` 拆開**,之後直接跑裡面的 `AppRun`。
#
# ⚠ 拆開之後 `appimagetool` 就是一支普通的動態連結執行檔 ⇒ **它自己的相依要補**
#(`libgpgme.so.11` / glib)。原本那份 AppImage 把它們打在裡面,拆開就散了 ——
# 症狀是 `error while loading shared libraries`,看起來像 image 沒裝好。
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates curl file zsync desktop-file-utils \
      libgpgme11 libglib2.0-0 fuse3 \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /opt
RUN curl -fsSL -o appimagetool \
      https://github.com/AppImage/AppImageKit/releases/download/continuous/appimagetool-x86_64.AppImage \
    && chmod +x appimagetool && ./appimagetool --appimage-extract >/dev/null \
    && mv squashfs-root appimagetool.d && rm appimagetool
ENV PATH="/opt/appimagetool.d/usr/bin:${PATH}" APPIMAGE_EXTRACT_AND_RUN=1
