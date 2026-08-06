# 創世紀 V:命運勇士 — 繁體中文重製

> *Ultima V: Warriors of Destiny*(1988)—— Go + Ebitengine 重寫的跨平台引擎,加上完整繁體中文化。

**現況:早期開發(P0 框架完成)。還不能玩。** 進度見 [`PLAN.md`](PLAN.md);
玩家向的完整 README(遊戲介紹、操作、畫廊)會在 P8 完稿。

---

## 這是什麼

1988 年,Richard Garriott 讓玩家在《Ultima IV》成為聖者(Avatar)之後,把不列顛尼亞交到一個
暴君手上 —— Lord British 失蹤,取而代之的是以「美德法典」之名行壓迫的統治,以及三名暗影君主。
於是《Ultima V》問的不再是「你願不願意成為有德之人」,而是**當美德變成統治工具時,你怎麼辦**。

本專案不是修補原版執行檔,也不是接手某個開源引擎(U5 沒有可用的)——是**用 Go 從零重寫遊戲引擎**,
行為以 IDA Pro 反組譯原版當真值來源,**所有素材用原版**(美術、音樂、音效、地圖、數值一律不自製),
中文採 1990 年代 DOS 中文系統的**倚天原生點陣字**,而不是把現代 TTF 縮小。

## 技術路線

| 項目 | 選擇 |
|---|---|
| 語言 / 引擎 | Go 1.24 + Ebitengine v2 |
| 邏輯畫布 | 640×400(原版 320×200 的乾淨 2×),底圖 nearest 整數放大 |
| 中文字形 | **倚天中文系統(ETEN 3.53)原生點陣字**,16×15 / 24×24 |
| 行為真值 | IDA Pro 9.4;FM Towns 版是 32-bit Phar Lap `P3`,**Hex-Rays 反編譯可用** |
| 建置 | 全程 Docker |

三個版本的原版素材互為對照組:**DOS 1988**(資料格式主線)、**FM Towns 1992**(未壓縮 tileset、
英日雙執行檔、英日對照對話、CD 音軌)、**PC-98**(第二組未壓縮素材、YM2203 音樂)。
細節見 [`CLAUDE.md §2`](CLAUDE.md)(每一條都經一手開檔驗證)與 [`docs/re/`](docs/re/)。

## 譯名與手冊

繁體譯名以 **1990 年代《軟體世界》雜誌「說明書補完計劃」的《創世紀第 V 代》中文說明書**為權威來源,
並與姊妹專案 [u4-cht](https://github.com/wicanr2/u4-cht) / [u6-cht](https://github.com/wicanr2/u6-cht)
的《創世紀聖者之書》體系對齊;系列共通名(不列顛王、八德、聖者)不另立新譯。
逐項決定與理由會整理在 `docs/manual/術語對照.md`。

> 手冊掃描與轉錄僅供研究與譯名對照之用,著作權屬《軟體世界》雜誌與原譯者;
> 非商業使用,權利人要求即撤除。

## 開發

```bash
# 建置容器(Go + CGO/GL/X11 + xvfb 軟體 GL + Pillow + 7z)
docker build -t u5cht/dev docker/

# 編譯與測試
docker run --rm --log-opt max-size=10m --log-opt max-file=3 \
  --user "$(id -u):$(id -g)" -e HOME=/tmp -v "$PWD":/work -w /work u5cht/dev \
  bash -c "go build ./... && go vet ./... && go test ./..."

# 烘倚天中文點陣字(自備倚天字庫;產物不入庫)
docker run --rm --log-opt max-size=10m --log-opt max-file=3 \
  -v "$PWD":/work -v /path/to/etan_font:/eten:ro -w /work u5cht/dev \
  python3 tools/build_eten_font.py --eten-dir /eten --iso /eten/ET353S.iso \
    --size 15 --out assets/fonts/eten-16x15
```

**原版資料由玩家自備**,放進 `gamedata/`(不入庫)。整合測試設 `U5_GAMEDATA` 指向該目錄後才會執行,
沒設就跳過。

其他工具:`tools/ida.sh`(headless 反組譯)、`tools/fdi_extract.py`(PC-98 磁碟映像抽檔)、
`tools/cdimg_to_iso.py`(CD raw 映像 → ISO)。

## 文件索引

| 文件 | 內容 |
|---|---|
| [`CLAUDE.md`](CLAUDE.md) | 專案憲章:硬規則、素材盤點(已驗證事實)、IDA 使用紀律、中文化設計、階段與驗收 |
| [`PLAN.md`](PLAN.md) | 執行計畫與 backlog |
| [`CONTEXT.md`](CONTEXT.md) | 語彙(glossary)、與姊妹專案的關係、技術關鍵事實 |
| [`docs/re/00-hexrays-p3-verified.md`](docs/re/00-hexrays-p3-verified.md) | FM Towns `.EXP`(Phar Lap P3)可反編譯的驗證紀錄 |

## 授權與邊界

- **程式碼**:MIT。
- **原版遊戲資料、美術、音樂、字型**:各原權利人所有,**不隨本專案散布**;
  repo 只提供解碼工具與引擎,玩家自備合法副本。
- *Ultima V: Warriors of Destiny* © 1988 Origin Systems / Richard Garriott。
  本專案與 Origin / EA 無關聯。
