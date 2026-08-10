#!/usr/bin/env python3
"""建立只含程式修正檔的繁中 patch 目錄。

用法:mkpatch.py <staging-dir> <版本> <linux|windows|macos>

patch 不含原版遊戲資料、中文字庫或音樂；玩家要把它解壓到相同平台的
完整版安裝目錄，覆蓋清單中的檔案。檔名刻意維持 ASCII，避免繁中 Windows
解壓工具誤判檔名編碼。
"""

from __future__ import annotations

import hashlib
import sys
from pathlib import Path


PAYLOAD = {
    "linux": ("u5cht",),
    "windows": ("u5cht.exe",),
    "macos": (
        "Ultima V CHT.app/Contents/Info.plist",
        "Ultima V CHT.app/Contents/MacOS/u5cht.bin",
    ),
}


def write_text(path: Path, text: str) -> None:
    with path.open("w", encoding="utf-8-sig", newline="\r\n") as f:
        f.write(text)


def main() -> None:
    if len(sys.argv) != 4 or sys.argv[3] not in PAYLOAD:
        raise SystemExit(__doc__)

    root = Path(sys.argv[1])
    version = sys.argv[2]
    target = sys.argv[3]
    payload = []
    for rel in PAYLOAD[target]:
        path = root / rel
        if not path.is_file():
            raise SystemExit(f"✗ patch 缺少必要檔案: {rel}")
        payload.append((rel, path))

    readme = (
        f"創世紀 V:命運勇士 —— {version} {target} 程式修正包\n"
        "\n"
        "這不是完整遊戲包，不能單獨啟動。它只替換程式與版本資訊，"
        "不含原版遊戲資料、中文字庫或音樂。\n"
        "\n"
        "使用方式\n"
        "---------------------------------------------------------------\n"
        "1. 先關閉遊戲並備份目前的安裝目錄。\n"
        "2. 將本包解壓到同一平台的完整版安裝目錄，允許覆蓋同名檔案。\n"
        "3. Linux／Windows 直接覆蓋根目錄的執行檔；macOS 會覆蓋\n"
        "   `Ultima V CHT.app/Contents/MacOS/u5cht.bin` 與版本資訊。\n"
        "4. 保留原本的 gamedata/ 與 assets/；patch 不會替換它們。\n"
        "\n"
        "本版包含戰鬥圖片索引修正、現代／原版 UI 切換、F2、以及日／月指示。\n"
        "完整檔案與 SHA-256 請見 PATCH-MANIFEST.txt。\n"
        "\n"
        "程式碼採 MIT 授權；原版資料、美術、音樂與字型不在本包內，"
        "仍屬原權利人。\n"
    )
    write_text(root / "PATCH-README-CHT.txt", readme)

    lines = [
        f"版本: {version}",
        f"平台: {target}",
        "說明: 以下為本 patch 會覆蓋的檔案；原版資料與資產不在包內。",
        "",
    ]
    for rel, path in payload:
        digest = hashlib.sha256(path.read_bytes()).hexdigest()
        lines.append(f"{digest}  {path.stat().st_size:>12}  {rel}")
    write_text(root / "PATCH-MANIFEST.txt", "\n".join(lines) + "\n")

    print(f"✓ {target} patch 已建立({len(payload)} 個程式檔)")


if __name__ == "__main__":
    main()
