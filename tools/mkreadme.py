#!/usr/bin/env python3
"""產生交付包裡的說明檔與啟動器,編碼照繁中 Windows 的規矩。

  README-CHT.txt   UTF-8 **with BOM** + CRLF —— 沒有 BOM 的話舊版記事本
                   會當成 ANSI 解讀,中文全亂。
  PLAY.bat         **CP950** + CRLF —— cmd.exe 用當前字碼頁解讀批次檔,
                   開頭 `chcp 950` 就代表檔案本身也得是 CP950。
                   ⚠ CP950 打不出的字會直接寫檔失敗 —— 那是好事,
                   逼我們在打包階段就發現用字問題,而不是讓玩家看到亂碼。

用法:mkreadme.py <輸出目錄> <版本> <linux|windows>
"""
import os
import sys

README = """創世紀 V:命運勇士 —— 繁體中文重製版  {version}

這是用 Go + Ebitengine 從零重寫的《Ultima V: Warriors of Destiny》(1988)
遊戲引擎,加上繁體中文化。行為以 IDA Pro 反組譯原版取得,不是修改原版執行檔。


一、你需要自備原版遊戲資料
---------------------------------------------------------------
本包**不含**原版的地圖、美術、音樂與文字 —— 那些的著作權屬於原權利人。
請自備一份合法的 DOS 版《Ultima V》,把資料檔複製到本目錄下的 gamedata\\ 裡。

需要的檔案至少有:
  BRIT.DAT  UNDER.DAT  TOWNE.DAT  CASTLE.DAT  DWELLING.DAT  KEEP.DAT
  TOWNE.TLK CASTLE.TLK DWELLING.TLK KEEP.TLK
  TOWNE.NPC CASTLE.NPC DWELLING.NPC KEEP.NPC
  DATA.OVL  TILES.16  SHOPPE.DAT  STORY.DAT  MISCMSG.DAT  ENDMSG.DAT
  KARMA.DAT QUESTION.DAT MISCMAPS.DAT DUNGEON.DAT DUNGEON.CBT BRIT.CBT
  SAVED.GAM INIT.GAM  (以及同一目錄的其餘檔案,整包複製最省事)


二、中文字型要自己烘
---------------------------------------------------------------
畫面上的中文用的是 1990 年代倚天中文系統的**原生點陣字**,不是把現代 TTF
縮小 —— 那份字庫是商業字型,同樣不隨本包散布。

自備倚天字庫(STDFONT.15 與 SPCFONT.15)後,用原始碼裡的
tools/build_eten_font.py 烘成 assets/fonts/ 底下的字圖。
沒有字型時遊戲仍然啟動,但中文會顯示不出來。


三、開始玩
---------------------------------------------------------------
{howto}

存檔放在系統的設定目錄(Windows 是 %APPDATA%,Linux 是 ~/.config),
不會寫進遊戲目錄,所以裝在唯讀路徑也沒問題。


四、目前的完成度
---------------------------------------------------------------
遊戲機制大致完整:世界地圖、三十二個地點、時鐘與 NPC 排程、對話、八種商店、
戰鬥、魔法、地牢第一人稱透視、聖壇、寶典、力量之言、暗影君主、黑棘審問、結局。

中文化尚未完成:介面、譯名、系統訊息與**商店對白**都是中文,
但 NPC 的對白本文絕大多數還是英文。這一點在畫面上會很明顯,先說在前頭。

音樂與音效還沒接上。


五、授權
---------------------------------------------------------------
程式碼 MIT。原版遊戲資料、美術、音樂、字型各屬原權利人,不隨本包散布。
Ultima V: Warriors of Destiny (c) 1988 Origin Systems / Richard Garriott。
本專案與 Origin / EA 無關聯。

原始碼與問題回報:https://github.com/wicanr2/u5-cht
"""

HOWTO_WIN = """把原版資料放進 gamedata\\ 之後,雙擊 PLAY.bat。
也可以直接執行 u5cht.exe;常用參數:
  u5cht.exe -gamedata D:\\Ultima5 -scale 3
"""

HOWTO_NIX = """把原版資料放進 gamedata/ 之後執行:
  ./u5cht
常用參數:
  ./u5cht -gamedata /path/to/ultima5 -scale 3
"""

BAT = """@echo off
chcp 950 >nul
cd /d "%~dp0"
if not exist "gamedata\\DATA.OVL" (
  echo.
  echo   找不到原版遊戲資料。
  echo.
  echo   請把 Ultima V 的資料檔複製到這個目錄下的 gamedata\\ 裡,
  echo   詳見 README-CHT.txt。
  echo.
  pause
  exit /b 1
)
u5cht.exe %*
if errorlevel 1 pause
"""


def main() -> None:
    if len(sys.argv) != 4:
        sys.exit(__doc__)
    out, version, target = sys.argv[1], sys.argv[2], sys.argv[3]
    os.makedirs(out, exist_ok=True)

    howto = HOWTO_WIN if target == "windows" else HOWTO_NIX
    text = README.format(version=version, howto=howto)

    # UTF-8 with BOM + CRLF:沒有 BOM 的話舊記事本會當 ANSI,中文全亂。
    with open(os.path.join(out, "README-CHT.txt"), "w",
              encoding="utf-8-sig", newline="\r\n") as f:
        f.write(text)

    if target == "windows":
        # CP950 + CRLF。打不出的字會在這裡就爆掉,不會流到玩家端。
        with open(os.path.join(out, "PLAY.bat"), "w",
                  encoding="cp950", newline="\r\n") as f:
            f.write(BAT)

    os.makedirs(os.path.join(out, "gamedata"), exist_ok=True)
    with open(os.path.join(out, "gamedata", "PUT-GAME-DATA-HERE.txt"), "w",
              encoding="utf-8-sig", newline="\r\n") as f:
        f.write("把原版 Ultima V 的資料檔複製到這個目錄裡。詳見上一層的 README-CHT.txt。\r\n")

    print("✓ %s 的說明檔已產生(%s)" % (target, out))


main()
