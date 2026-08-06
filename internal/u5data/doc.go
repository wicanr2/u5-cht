// Package u5data 只做一件事:把原版 Ultima V 的資料檔解碼成 Go 型別。
//
// 邊界(重要):
//   - 唯讀。永遠不寫回原版檔案 —— 中文譯文走 i18n 覆蓋層,不改 .TLK/.DAT
//     (寫回會破壞檔頭的 offset 表)。
//   - 不含遊戲邏輯。規則與狀態機在 internal/engine。
//   - 不含繪圖。這裡回傳的是點陣與索引,不是 ebiten.Image。
//
// 支援的版本(格式不同,別混用):
//   - DOS 1988(主線):TILES.16/.4 壓縮、IBM.CH 8x8 字型、.TLK 每 byte bit7 被設為 1
//   - FM Towns 1992:EGA0-3.TIL 未壓縮、U5_J/*.JPN 是 Shift-JIS
//   - PC-98:TILES.CH 未壓縮、FONT98.CH
//
// 每個解碼器都要能指出「哪一條事實是實測驗證過的、哪一條還是推導」。
// 已驗證事實見 CLAUDE.md §2 與 docs/formats/;推導未證的地方一律留 TODO 並寫明缺什麼證據。
package u5data
