# 工程筆記

## 1. 畫面繪製不要綁 GPU(2026-08-07,五小時的教訓)

### 症狀

headless 截圖指令跑了**五小時沒有任何輸出**,`slice.png` 沒產出,輸出檔停在 `go test` 的結果。
容器 `Up 5 hours`,4 個 `go run` process 還在。最後只能 `docker stop`。

### 根因

當時的架構是「繪製邏輯綁在 ebiten 上」:

```
render.TextRenderer / render.TileSet  →  *ebiten.Image
cmd/u5cht -headless                   →  RunGame → Draw → ReadPixels → PNG
```

於是「產生一張畫面」這件事必須:起 xvfb → 建 GL context(llvmpipe 軟體渲染)→
進 ebiten 的事件迴圈 → 等它呼叫 Draw。**這條鏈上任何一環卡住,截圖就永遠不會出現。**
而軟體 GL 正是已知會卡的那一環(u4-cht 也踩過:「F2 圖形熱切換在軟體 GL 死鎖」)。

雪上加霜的是,`-headless` 的結束條件寫成「`Draw` 裡設 `done`、`Update` 裡 return Termination」——
若事件迴圈卡在 GL 初始化,`Update` 根本不會被呼叫,連逾時自殺的機會都沒有。

### 為什麼「加逾時」不是解法

加 `timeout`、換 GL 後端、調 mesa 環境變數,都只是讓失敗快一點,
**沒有改變「驗證畫面必須有 GPU」這個結構性問題**。CI 上會再遇到一次。

### 解法:單一 CPU 繪製路徑

`internal/render` 改成完全不依賴 ebiten,全部畫在 `image.NRGBA`:

```
render.Scene.Render() → *image.NRGBA (640×400)
   ├─ u5dump scene  → 直接 png.Encode          ← headless 驗收(秒級,不需 GPU)
   └─ cmd/u5cht     → ebiten.NewImageFromImage ← 實機顯示(只做上傳 + 放大 + 收鍵盤)
```

得到三件事:

1. **headless 驗證是秒級的純函式呼叫**,CI 不需要顯示環境。實測:5 小時 → 數秒。
2. **只有一份繪製實作**,截圖就是實機畫面,不會漂移。
3. `internal/render` 可以寫單元測試了 —— 綁 ebiten 時連 `go test` 都會因為
   `ui.init` 找不到 X11 而 panic。

效能不是問題:每幀重畫 640×400 約 256K 像素,而且回合制遊戲的畫面只在狀態改變時才需要重畫
(`dirty` 旗標)。

### 一般化

> **能不能在沒有硬體的地方產生輸出**,是架構問題,不是測試技巧問題。
> 當「驗證某件事」需要拉起一整條重量級相依鏈時,先問「這條鏈是本質必要的嗎」。

同源規則:`rulebook/60`(先建可重跑的 pass/fail loop)、
`rulebook/41`(特例越補越多就停下重想架構)。

## 2. 背景長跑要有活死判準

這次是使用者提醒「shell 跑了五個小時」才發現卡住。判斷依據應該內建:

- **輸出檔有沒有增長**(這次:停在 test 結果之後就沒動過)
- **預期耗時的量級**(CGO 編 ebiten 約 1–3 分鐘,五小時差了兩個數量級)
- **有沒有產出預期的檔案**

砍掉時的紀律:先 `docker ps --filter ancestor=<自己的 image>` 確認**只停自己建的容器**。
這台機器同時放著多個客戶專案的容器(civ1、pq1、moo2、portainer…),
誤停別人的就是事故(`CLAUDE.md §8`)。
