#!/usr/bin/env python3
"""產生交付包用的圖示(AppImage / .desktop / macOS 用得到)。

⚠ **不用原版美術。** 原版的 tile 與插圖是版權素材,不隨本專案散布
(`CLAUDE.md §3.0`)—— 而圖示是**一定會進交付包**的東西,所以它必須是
我們自己畫的。這裡用程式畫一個生命之符(ankh):U5 的護符,
形狀簡單、與原版點陣圖無關。

用法:tools/mkicon.py <out.png> [邊長,預設 256]
"""
import sys
from PIL import Image, ImageDraw

BG = (16, 16, 40, 255)      # 與遊戲的 ColorBackground 同一個色
GOLD = (255, 208, 64, 255)  # 與離開確認框的邊框同色


def draw_ankh(size: int) -> Image.Image:
    im = Image.new("RGBA", (size, size), BG)
    d = ImageDraw.Draw(im)
    u = size / 16                      # 以十六分之一邊長當單位,縮放才穩
    cx = size / 2
    bar = max(2, int(u * 1.1))         # 筆畫寬度
    # 上面的環
    r = u * 3
    d.ellipse([cx - r, u * 1.6, cx + r, u * 1.6 + 2 * r], outline=GOLD, width=bar)
    # 直幹
    d.line([(cx, u * 1.6 + 2 * r - bar / 2), (cx, size - u * 1.6)], fill=GOLD, width=bar)
    # 橫臂
    d.line([(cx - u * 3.6, u * 9.2), (cx + u * 3.6, u * 9.2)], fill=GOLD, width=bar)
    return im


def main() -> int:
    if len(sys.argv) < 2:
        print(__doc__)
        return 2
    size = int(sys.argv[2]) if len(sys.argv) > 2 else 256
    draw_ankh(size).save(sys.argv[1])
    print(f"✓ {sys.argv[1]}({size}×{size},程式畫的 ankh,非原版美術)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
