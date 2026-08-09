# 在 Linux 上為 macOS 交叉編譯的工具鏈(Go + CGO,ebiten 要 Cocoa / Metal)。
#
#   docker build -t u5cht/osxcross -f docker/osxcross.Dockerfile docker/
#
# ⚠ **SDK 的授權只允許在 Apple 硬體上使用。** 自用交叉編是一回事,
# 散布 SDK 或含 SDK 的 image 是另一回事 —— 這個 image **不推、不散布**。
# 要對外散布 macOS 版並過 Gatekeeper(notarization)一定要真的 Mac,
# 那條走 GitHub Actions 的 macos runner(`.github/workflows/`)。
#
# ⚠ Linux 上**執行不了** macOS binary ⇒ 這條路只能做**靜態驗收**
#(架構、最低系統版本、相依、ad-hoc 簽章)。靜態全過只代表不會因結構問題
# 開不起來,不代表功能正常 —— 交付時要講明。
FROM crazymax/osxcross:15.5-debian AS osxcross

# ⚠ base **必須是 ubuntu:24.04**(或同世代)。用 `golang:1.24-bookworm`
# 會撞到 `GLIBC_2.38 not found` / `GLIBCXX_3.4.32 not found` ——
# `crazymax/osxcross` 的工具是在新世代 glibc 上編的,bookworm 只有 2.36。
# 症狀出現在 `osxcross-conf` 這種最基本的指令上,看起來像 image 壞了。
# Go 從官方 image 直接搬過來(Go 的工具鏈不吃這條相依)。
FROM ubuntu:24.04
COPY --from=golang:1.24-bookworm /usr/local/go /usr/local/go
ENV PATH="/usr/local/go/bin:${PATH}" GOTOOLCHAIN=local
RUN apt-get update && apt-get install -y --no-install-recommends \
      clang lld llvm libxml2 libssl3 liblzma5 zlib1g file xz-utils python3 \
      ca-certificates git \
    && rm -rf /var/lib/apt/lists/*
COPY --from=osxcross /osxcross /osxcross
# ⚠ 少了這一行 `ld64` 起不來(找不到 libxar.so.1),而 clang 只會轉述成
# 「unable to execute command: No such file or directory」—— 看起來像前綴打錯。
RUN echo /osxcross/lib > /etc/ld.so.conf.d/osxcross.conf && ldconfig
ENV PATH="/osxcross/bin:${PATH}"
