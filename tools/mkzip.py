#!/usr/bin/env python3
"""把一個目錄打成給 Windows 玩家的 zip。

三條規則(出自 `knowledge-base/retro-cht/cjk-package-encoding`):

1. **檔名一律 ASCII。** zip 格式沒有檔名編碼欄位;繁中 Windows 用 CP950 去解讀
   UTF-8 的位元組,運氣好是亂碼,運氣差是**那串位元組在 CP950 裡非法 → 解壓工具
   直接跳過該檔**,玩家看到的是「檔案不見了」。中文留在檔案**內容**裡。
2. **仍然設 UTF-8 旗標並驗一次。** 日後有人加了中文檔名,才不會靜默回到老問題。
3. `.bat` 存 CP950、`.txt` 存 UTF-8 with BOM,換行都用 CRLF —— 那兩件事在
   `package.sh` 裡做,這支只負責打包與檢查。

用法:mkzip.py <out.zip> <來源目錄>
"""
import os
import sys
import zipfile


def main() -> None:
    if len(sys.argv) != 3:
        sys.exit(__doc__)
    out, src = sys.argv[1], sys.argv[2]

    bad = [
        os.path.relpath(os.path.join(d, n), src)
        for d, dirs, files in os.walk(src)
        for n in dirs + files
        if any(ord(c) > 127 for c in n)
    ]
    if bad:
        sys.stderr.write("✗ 這些檔名含非 ASCII,繁中 Windows 解壓會亂碼或整個跳過:\n")
        for b in bad:
            sys.stderr.write("    %s\n" % b)
        sys.exit(1)

    if os.path.exists(out):
        os.remove(out)
    n = 0
    with zipfile.ZipFile(out, "w", zipfile.ZIP_DEFLATED) as z:
        for d, _dirs, files in sorted(os.walk(src)):
            for fn in sorted(files):
                full = os.path.join(d, fn)
                z.write(full, os.path.relpath(full, src))
                n += 1

    # 保險:非 ASCII 檔名必須帶 UTF-8 旗標(bit 11)。
    with zipfile.ZipFile(out) as z:
        for i in z.infolist():
            if any(ord(c) > 127 for c in i.filename) and not (i.flag_bits & 0x800):
                sys.exit("✗ %r 沒有 UTF-8 旗標" % i.filename)

    print("✓ %s(%d 個檔,檔名全為 ASCII)" % (out, n))


main()
