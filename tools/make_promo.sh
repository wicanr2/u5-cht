#!/usr/bin/env bash
# 合成推廣短片(靜態投影片 + 淡入淡出 + 原版配樂),全程在容器裡。
#
# 這支跑在 `u5cht/video` 容器內,由 `tools/promo.sh` 帶進去。
#
# ── 兩條規則的位置 ────────────────────────────────────────────────────
#
# ★ **素材真實性**(`rulebook/93` 鐵則 1):畫面全部是引擎跑出來的實際截圖
#(`docs/screenshots/`,由 `tools/shots.sh` 產,與實機共用同一條 CPU 繪製路徑),
# 沒有 mockup、沒有重畫。配樂用**原版光碟的 CD-DA 音軌**(`CDDA2.ogg`,
# 44.1 kHz 立體聲 180 秒 3.3 MB —— 真實錄音)。
#
# ⚠ **不用 `M*.ogg`**:那些是我們自己的 FM 合成器算出來的(單聲道、
# 0.86 MB / 169 s),音色不對。鐵則 1 的來源教訓正是這個 ——
# u1-cht 用自寫 2-op FM 當配樂,被評為「不適合人類聽」。
#
# ⚠ **zoompan 一律不用**(`game-promo-video-ffmpeg` 雷 #1):
# `-loop 1 -t S` 配上前置 `fps` 會讓 zoompan 算出 (FPS×S)² 幀 ——
# 6 秒 25fps ≈ 22500 幀,燒滿 CPU 好幾分鐘而 output 一直空白。
# 靜態圖 + fade 對投影片式的 promo 完全夠看。
set -eu

# ===== 設計 token:**從遊戲本身萃取**,不是憑喜好挑 =====
# 全部取自引擎實際在用的顏色,所以片子的配色與遊戲畫面自然一致。
BG='#101028'        # render.ColorBackground
BGD='#08081a'       # 同色調再暗一階
GOLD='#ffd040'      # 離開確認框的邊框色(colorQuitFrame)
GOLDSH='#aa5500'    # ★ EGA 色號 6 —— 並排比對量出來的那個棕(docs/re/62)
CREAM='#e8e8d8'     # render.ColorText
BLOOD='#aa0000'     # EGA 色號 4
FB=/usr/share/fonts/opentype/noto/NotoSerifCJK-Bold.ttc
FR=/usr/share/fonts/opentype/noto/NotoSerifCJK-Regular.ttc
FM=/usr/share/fonts/opentype/noto/NotoSerifCJK-Medium.ttc
W=1280; H=720; FPS=25
SHOT=/shots; OUT=/out; MUSIC=/music/CDDA2.ogg; TMP=/tmp/promo
mkdir -p "$TMP" "$OUT"

for f in "$FB" "$FR" "$FM"; do
  [ -f "$f" ] || { echo "✗ 字型不存在:$f" >&2; exit 1; }
done
[ -f "$MUSIC" ] || { echo "✗ 配樂不存在:$MUSIC" >&2; exit 1; }

# ── 版面函式:五種,輪流用 ────────────────────────────────────────────
# 單一版面重複十二段會很單調(kb 的「版面變化」那節)。

bg() { # $1 out —— 徑向漸層底
  convert -size ${W}x${H} "radial-gradient:#1c1c44-${BGD}" "$1"
}

card() { # $1 out $2 中標 $3 英標 $4 副標 —— 鎏金浮雕標題卡
  bg "$TMP/_bg.png"
  # ⚠ 三行都掛在 center 上,而位移全是**往下** ⇒ 整塊的視覺重心會落在畫面下半。
  # 抽幀讀圖才看得出來(第一版就是這樣)。整塊往上提 70px 才是真的置中。
  convert "$TMP/_bg.png" -gravity center \
    -font "$FB" -fill "$GOLDSH" -pointsize 88 -annotate +4-66 "$3" \
    -fill "$GOLD" -pointsize 88 -annotate +0-70 "$3" \
    -font "$FB" -fill "$CREAM" -pointsize 60 -annotate +0+26 "$2" \
    -font "$FR" -fill "$GOLD" -pointsize 28 -annotate +0+110 "$4" "$1"
}

slide_frame() { # $1 out $2 截圖 $3 字幕 —— 金框置中
  bg "$TMP/_bg.png"
  convert "$SHOT/$2" -resize x556 -bordercolor "$GOLD" -border 3 "$TMP/_sc.png"
  convert "$TMP/_bg.png" "$TMP/_sc.png" -gravity north -geometry +0+28 -composite \
    -fill "#000000bb" -draw "rectangle 0,632 ${W},720" \
    -font "$FM" -fill "$CREAM" -gravity south -pointsize 34 -annotate +0+28 "$3" "$1"
}

slide_big() { # $1 out $2 截圖 $3 字幕 —— 放到最大但**不裁切**,字幕在下方
  # ⚠ 原本這裡用 `-resize ${W}x${H}^ -extent`(填滿再裁)—— 而遊戲畫面是
  # 640×400(1.6:1)、影片是 16:9,裁下去**右邊的狀態欄與底部的提示列會被切掉**。
  # 抽幀讀圖才發現:右欄的標題只剩半行、提示列被字幕條蓋住一半。
  # ⇒ 改成「等比放到最大再置中」,一格都不裁 —— 遊戲畫面本身就是內容,不能裁。
  convert "$SHOT/$2" -resize x608 "$TMP/_sc.png"
  bg "$TMP/_bg.png"
  convert "$TMP/_bg.png" "$TMP/_sc.png" -gravity north -geometry +0+4 -composite \
    -fill "#000000cc" -draw "rectangle 0,616 ${W},720" \
    -font "$FM" -fill "$CREAM" -gravity south -pointsize 36 -annotate +0+32 "$3" "$1"
}

split_ba() { # $1 out $2 左圖 $3 右圖 $4 左標 $5 右標 $6 字幕 —— 左右對比
  bg "$TMP/_bg.png"
  convert "$SHOT/$2" -resize 612x -bordercolor "$GOLD" -border 2 "$TMP/_l.png"
  convert "$SHOT/$3" -resize 612x -bordercolor "$GOLDSH" -border 2 "$TMP/_r.png"
  convert "$TMP/_bg.png" \
    "$TMP/_l.png" -gravity northwest -geometry +12+96 -composite \
    "$TMP/_r.png" -gravity northeast -geometry +12+96 -composite \
    -font "$FM" -fill "$CREAM" -gravity northwest -pointsize 30 -annotate +20+52 "$4" \
    -gravity northeast -pointsize 30 -annotate +20+52 "$5" \
    -fill "#000000bb" -draw "rectangle 0,632 ${W},720" \
    -font "$FM" -fill "$GOLD" -gravity south -pointsize 34 -annotate +0+28 "$6" "$1"
}

dcard() { # $1 out $2 引文 $3 出處 —— 左對齊 + 巨型引號
  bg "$TMP/_bg.png"
  convert "$TMP/_bg.png" \
    -font "$FB" -fill "#ffd0405c" -gravity northwest -pointsize 300 -annotate +48-56 '"' \
    -font "$FM" -fill "$CREAM" -gravity west -pointsize 44 -annotate +140-30 "$2" \
    -font "$FR" -fill "$GOLD" -gravity southeast -pointsize 26 -annotate +80+90 "$3" "$1"
}

seg() { # $1 png $2 out.mp4 $3 秒 —— 靜態 + 淡入淡出(**不用 zoompan**)
  FO=$(awk "BEGIN{print $3-0.5}")
  # ⚠ `ultrafast` 而不是 `veryfast`:十四段 1280×720 用 veryfast 要跑三、四分鐘,
  # 而這些段落之後**不會再重編**(concat 與配樂都走 `-c copy`)⇒
  # 畫質由這一步定案,靜態投影片在 ultrafast 下看不出差別。
  ffmpeg -y -loglevel error -loop 1 -i "$1" -t "$3" -r $FPS \
    -vf "fade=t=in:st=0:d=0.5,fade=t=out:st=$FO:d=0.5,format=yuv420p" \
    -threads 2 -c:v libx264 -preset ultrafast -pix_fmt yuv420p "$2"
}

# ── 分鏡 ──────────────────────────────────────────────────────────────
echo "→ 產生分鏡圖"
card        "$TMP/00.png" '創世紀 V:命運勇士' 'Ultima V' '1988 年的不列顛尼亞 · 全程繁體中文'
dcard       "$TMP/01.png" '當美德變成統治的工具,\n汝要如何自處?' '不列顛王失蹤,黑棘登位'
slide_big   "$TMP/02.png" 01-world.png       '256×256 格的不列顛尼亞,由原版地圖分塊拼成'
slide_frame "$TMP/03.png" 02-town.png        '三十二個地點,NPC 依原版的時刻表走動'
slide_frame "$TMP/04.png" 03-talk.png        '1,712 段 NPC 對白全數中譯 —— 關鍵字對話照原版'
slide_frame "$TMP/05.png" 04-shop.png        '八種商店、194 段店家對白,價格照原版的表'
slide_big   "$TMP/06.png" 05-combat.png      '11×11 戰場,命中與傷害公式來自反組譯'
slide_frame "$TMP/07.png" 06-dungeon.png     '地牢第一人稱透視 —— 沒點火把是全黑的,原版就是這樣'
split_ba    "$TMP/08.png" 15-night.png 16-night-torch.png '入夜' '點起火把' '視線遮蔽與夜間照明照原版的光照半徑'
split_ba    "$TMP/09.png" 25-ui-modern.png 26-ui-original.png '現代版面' '原版版面' 'F2 隨時切換兩種版面'
slide_frame "$TMP/10.png" 23-help.png        'F1 指令說明 —— 原版 A–Z 全是指令,1988 年靠紙本手冊'
slide_frame "$TMP/11.png" 24-menu.png        '從建立新角色開始:吉普賽人的七題八德淘汰賽'
slide_frame "$TMP/12.png" 11-codex.png       '八座聖壇 → 終極智慧之寶典 → 力量之言 → 暗影君主'
card        "$TMP/99.png" '創世紀 V:命運勇士' 'Ultima V' 'Go + Ebitengine 從零重寫 · 倚天點陣中文\ngithub.com/wicanr2/u5-cht'

echo "→ 每段轉成影片"
LIST="$TMP/list.txt"; : > "$LIST"
# 前段慢、亮點 6s、結尾留長音(kb 的節奏那節)。
set -- "00:6" "01:6" "02:6" "03:5" "04:6" "05:5" "06:5" "07:5" "08:6" "09:6" "10:5" "11:5" "12:5" "99:7"
for pair in "$@"; do
  id="${pair%%:*}"; sec="${pair##*:}"
  seg "$TMP/$id.png" "$TMP/s_$id.mp4" "$sec"
  echo "file '$TMP/s_$id.mp4'" >> "$LIST"
done

echo "→ 接起來"
# ⚠ `-c copy`:每一段都是同一組編碼參數 ⇒ 直接串流複製,不必再編一次。
# 原本這裡重編一遍,加上最後鋪配樂又編一遍 —— **同一批畫面編了三次**。
ffmpeg -y -loglevel error -f concat -safe 0 -i "$LIST" -c copy "$TMP/silent.mp4"

DUR=$(ffprobe -v error -show_entries format=duration -of csv=p=0 "$TMP/silent.mp4")
FO=$(awk "BEGIN{print $DUR-3}")
echo "→ 鋪配樂(原版 CD-DA;總長 ${DUR}s)"
# ⚠ **不要 `-shortest`**(kb 的慘雷):配樂比影片短時它會以音軌為準,
# 把結尾卡整張砍掉,而症狀是 `ffprobe` 的**視訊**長度也變短 ——
# 看起來像分鏡少了一段。
#
# ⚠ 迴圈用 **`-stream_loop -1`(輸入層)**,不用 `aloop=size=2000000000`:
# `aloop` 的 `size` 是「載進記憶體的取樣數上限」,20 億筆 × 立體聲 32-bit
# ≈ 16 GB —— 那不是保險,是記憶體炸彈。輸入層迴圈零成本。
#
# ⚠ `-c:v copy`:這一步只是把音軌接上去,視訊不必再編一次。
ffmpeg -y -loglevel error -i "$TMP/silent.mp4" -stream_loop -1 -i "$MUSIC" \
  -filter_complex "[1:a]atrim=0:$DUR,asetpts=N/SR/TB,afade=t=in:st=0:d=2,afade=t=out:st=$FO:d=3[a]" \
  -map 0:v -map "[a]" \
  -c:v copy -c:a aac -b:a 192k -movflags +faststart "$OUT/u5cht-promo.mp4"

echo "→ 驗長度(視訊與音訊必須相等)"
ffprobe -v error -select_streams v:0 -show_entries stream=duration -of csv=p=0 "$OUT/u5cht-promo.mp4"
ffprobe -v error -select_streams a:0 -show_entries stream=duration -of csv=p=0 "$OUT/u5cht-promo.mp4"
ls -lh "$OUT/u5cht-promo.mp4"
