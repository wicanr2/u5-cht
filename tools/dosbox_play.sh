#!/usr/bin/env bash
# 用 DOSBox-X 跑 **DOS 原版** Ultima V,照按鍵腳本逐步截圖 —— 重製版的比對基準。
#
#   tools/dosbox_play.sh <輸出目錄> <步驟檔>
#
# 步驟檔一行一步,`#` 開頭是註解:
#
#   <標籤> <TAB 或空白> <xdotool 鍵序,逗號分隔> <TAB 或空白> <這步之後等幾秒>
#
# 例:
#   01-title    -            3
#   02-journey  j            2
#   03-north    Up,Up,Up     1
#
# 鍵名用 xdotool 的寫法(`Return` `Escape` `Up` `a` `shift+n` …);`-` = 不按鍵,
# 只等待與截圖。每一步跑完存一張 `<輸出目錄>/<標籤>.png`。
#
# ── 三個設計上的決定 ────────────────────────────────────────────────
#
# ① **原版資料一律唯讀掛載,再複製到容器內的可寫目錄。** 原版會寫 `SAVED.GAM`,
#    而那份檔案同時是重製版整合測試的 fixture —— 讓 DOSBox 寫它等於在測試中途
#    改掉基準。(`CLAUDE.md`:`internal/u5data` 只讀、不寫回原版檔。)
#
# ② **畫布 640×400,與重製版相同。** Xvfb 開 640×400、DOSBox-X 全螢幕 +
#    `scaler=normal2x`,於是 EGA 的 320×200 剛好整數 2× 填滿 ——
#    兩邊截圖的幾何一致,可以並排逐區塊對。
#
# ③ **固定 `cycles=3000`。** 這不是猜的:原版附的 `run.bat` 自己寫著
#    `config -set "cpu cycles=3000"`。`cycles=auto` 是可重現性的敵人
#    (`~/.claude/knowledge-base/retro/dosbox-game-configs.md`)。
#
# ⚠ 這支腳本**有界**:每一步的等待有上限、整個容器由呼叫端的 `timeout` 收尾,
# 沒有任何無界的等檔輪詢(`rulebook/35` 禁止模式 ①)。
#
# ── 兩個實測出來的按鍵事實(省下重試,別再猜)────────────────────────
#
# ⓐ **開場動畫大約 9 秒**跑完才會進主選單。
# ⓑ **主選單選「Journey Onward」的鍵是 `KP_Enter`(數字鍵盤的 Enter)。**
#    `Return`、`space`、`j` 三個都沒有反應 —— 而且**沒有任何回饋**:
#    畫面完全不動,看起來就像「按鍵沒送進去」。
#    ⚠ 這兩種失敗長得一模一樣,所以要先用「按 `Down` 看游標會不會動」
#    把「鍵有沒有進去」與「這個鍵不對」分開,再去試別的鍵。
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT="${1:?用法:tools/dosbox_play.sh <輸出目錄> <步驟檔>}"
STEPS="${2:?缺步驟檔}"
GAME="${U5_GAMEDATA:-$ROOT/gamedata}"
IMAGE="${U5_DOSBOX_IMAGE:-u5cht/dosbox}"
# 整個容器的硬上限(秒)。步驟多就調大,但一定要有。
BUDGET="${U5_DOSBOX_BUDGET:-600}"

mkdir -p "$OUT"
[ -f "$GAME/ULTIMA.EXE" ] || { echo "✗ $GAME 裡沒有 ULTIMA.EXE" >&2; exit 1; }

# 容器內的驅動腳本。寫成檔案而不是 `-c` 的一長串,是為了讓引號與跳脫可讀。
cat > "$OUT/.drive.sh" <<'INNER'
#!/bin/bash
set -uo pipefail
OUT=/out
STEPS=/out/.steps.txt

# ① 原版資料複製到可寫目錄 —— 掛進來的 /game 是唯讀的。
mkdir -p /play
cp -r /game/. /play/
chmod -R u+w /play

cat > /play/dosbox.conf <<'CONF'
[sdl]
fullscreen=true
fullresolution=640x400
output=surface
autolock=false
[dosbox]
machine=ega
memsize=16
[render]
scaler=normal2x
aspect=false
[cpu]
core=normal
cputype=386
cycles=3000
[speaker]
pcspeaker=false
[sblaster]
sbtype=none
[gus]
gus=false
[autoexec]
mount c /play
c:
ultima.exe
CONF

Xvfb :99 -screen 0 640x400x24 -nolisten tcp >/tmp/xvfb.log 2>&1 &
XVFB=$!
for _ in $(seq 1 40); do xdpyinfo -display :99 >/dev/null 2>&1 && break; sleep 0.25; done
export DISPLAY=:99

dosbox-x -conf /play/dosbox.conf -nomenu >/tmp/dosbox.log 2>&1 &
DB=$!

# 等視窗出現(有界:最多 20 秒)。
WIN=""
for _ in $(seq 1 80); do
  WIN="$(xdotool search --onlyvisible --class dosbox 2>/dev/null | head -1 || true)"
  [ -n "$WIN" ] && break
  sleep 0.25
done
if [ -z "$WIN" ]; then
  echo "✗ DOSBox 視窗沒出現;log:" >&2; tail -30 /tmp/dosbox.log >&2; kill $DB $XVFB 2>/dev/null; exit 1
fi
xdotool windowactivate --sync "$WIN" 2>/dev/null || true
xdotool windowfocus --sync "$WIN" 2>/dev/null || true
xdotool windowraise "$WIN" 2>/dev/null || true
echo "→ DOSBox 視窗 $WIN(焦點:$(xdotool getwindowfocus 2>/dev/null || echo 不明))"

shot() { import -display :99 -window root "$OUT/$1.png" 2>/dev/null || \
         xwd -display :99 -root -silent | convert xwd:- "$OUT/$1.png"; }

# ② 逐步執行。`read` 吃三欄;鍵序用逗號分隔。
while IFS= read -r line; do
  case "$line" in ''|'#'*) continue ;; esac
  label="$(echo "$line" | awk '{print $1}')"
  keys="$(echo  "$line" | awk '{print $2}')"
  wait_s="$(echo "$line" | awk '{print ($3=="" ? "1" : $3)}')"
  if ! kill -0 $DB 2>/dev/null; then
    echo "✗ DOSBox 在 $label 之前就結束了" >&2; break
  fi
  if [ "$keys" != "-" ]; then
    IFS=',' read -ra ks <<< "$keys"
    for k in "${ks[@]}"; do
      case "$k" in
        # `__2` = 純等待 2 秒(某些畫面要等動畫跑完才吃鍵)。
        __*) sleep "${k#__}" ;;
        # ★★ **不要加 `--window`。** 帶 `--window` 的 xdotool 走 `XSendEvent`,
        # 而 SDL 會檢查 `send_event` 旗標**直接丟掉**合成事件 ——
        # 症狀是「腳本跑完、每張截圖都一樣」,看起來像遊戲卡住,
        # 其實一個鍵都沒進去(`retro-game-playtest` kb 記過同一個坑)。
        # 不帶 `--window` 走的是 XTEST,由 X server 當成真實輸入送出。
        *)   xdotool key --clearmodifiers "$k" ;;
      esac
      sleep 0.25
    done
  fi
  sleep "$wait_s"
  shot "$label"
  echo "  ✓ $label(鍵 $keys,等 ${wait_s}s)"
done < "$STEPS"

# ③ 收尾:自己起的就自己收。**不動任何 image / volume。**
kill $DB 2>/dev/null; sleep 1; kill -9 $DB 2>/dev/null || true
kill $XVFB 2>/dev/null || true
tail -5 /tmp/dosbox.log
INNER
chmod +x "$OUT/.drive.sh"
cp "$STEPS" "$OUT/.steps.txt"

echo "→ DOS 原版:$GAME(唯讀)  輸出:$OUT  上限 ${BUDGET}s"
timeout -s KILL "$BUDGET" docker run --rm \
  --log-opt max-size=10m --log-opt max-file=3 \
  --name "u5cht-dosbox-$$" \
  -v "$GAME":/game:ro \
  -v "$(cd "$OUT" && pwd)":/out \
  "$IMAGE" bash /out/.drive.sh
echo "→ 截圖在 $OUT"
