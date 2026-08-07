#!/usr/bin/env python3
"""驗一個 Windows 交付包的編碼。

這三件事只有在**繁體中文 Windows** 上才看得出問題,開發機上永遠是對的 ——
所以要在 CI 驗:

  1. 檔名全為 ASCII(zip 沒有檔名編碼欄位;非法的 CP950 序列會讓解壓工具
     直接跳過該檔,玩家看到的是「檔案不見了」)
  2. `PLAY.bat` 是 CP950 + CRLF(cmd.exe 用當前字碼頁解讀批次檔)
  3. `README-CHT.txt` 有 UTF-8 BOM + CRLF(沒 BOM 的話舊記事本當 ANSI 讀)

用法:checkpkg.py <package.zip>
"""
import sys
import zipfile


def main() -> None:
    if len(sys.argv) != 2:
        sys.exit(__doc__)
    z = zipfile.ZipFile(sys.argv[1])

    for i in z.infolist():
        if any(ord(c) > 127 for c in i.filename):
            if not i.flag_bits & 0x800:
                sys.exit("✗ %r 檔名非 ASCII 又沒有 UTF-8 旗標" % i.filename)
            sys.exit("✗ %r 檔名非 ASCII —— 繁中 Windows 可能整個跳過這個檔" % i.filename)

    names = z.namelist()
    for want in ("PLAY.bat", "README-CHT.txt"):
        if want not in names:
            sys.exit("✗ 包裡少了 %s(有:%s)" % (want, names))

    bat = z.read("PLAY.bat")
    try:
        bat.decode("cp950")
    except UnicodeDecodeError as e:
        sys.exit("✗ PLAY.bat 不是 CP950:%s" % e)
    if b"\r\n" not in bat:
        sys.exit("✗ PLAY.bat 不是 CRLF")

    rd = z.read("README-CHT.txt")
    if rd[:3] != b"\xef\xbb\xbf":
        sys.exit("✗ README-CHT.txt 少了 UTF-8 BOM")
    if b"\r\n" not in rd:
        sys.exit("✗ README-CHT.txt 不是 CRLF")

    print("✓ %s 的編碼正確(%d 個檔)" % (sys.argv[1], len(names)))


main()
