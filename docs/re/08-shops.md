# 商店

> 來源:FM Towns `WORRIORS.EXP`。`sub_1B294`(進店)、`sub_111CC`(挑問候語)、
> `sub_11168`(讀 SHOPPE.DAT)、`sub_10FEC`(代換佔位符)、`sub_9C7C`(營業時間)。

NPC 的對話號碼落在 **0x81–0xFC** 就是商人(`sub_1B52C` 的分派)。但實際只有
**0x81–0x88 這 8 種**有對應的店:

| 對話號碼 | 店種 | 例 |
|---|---|---|
| 0x81 | 武具店 | Iolo's Bows、Arms of Justice |
| 0x82 | 酒館 | The Honest Meal、The Sword and Keg |
| 0x83 | 馬廄 | Horse & Rider、The Stablehouse |
| 0x84 | 造船廠 | Island Shipwrights、The Crow's Nest |
| 0x85 | 藥草鋪 | The Herbalist、The Alchemist |
| 0x86 | 公會 | The Den、The Guild、The Nemesis |
| 0x87 | 治療所 | The Healers Mission、Sanctuary |
| 0x88 | 旅店 | The Wayfarer Inn、Hotel Brittany |

## 找店

```
店種 = 對話號碼 - 0x81
在 byte_4185C[店種][16] 裡找值等於**當前地點編號**的那一格 i
店名 = off_4145C[店種][i]      店主 = off_4165C[店種][i]
```

三張 16 欄的平行表,稀疏(多數欄位是 NULL)。全遊戲共 **47 家店**。

**店名與店主的字串在 DOS 的 `DATA.OVL` 裡就有**:0x0BFE 起 46 個店名、
0x0EFA 起 47 個店主(店主表結束處正好接上 0x104C 的詞典)。
兩份清單與執行檔的稀疏表**逐筆比對相符**,所以引擎只把結構寫死,字串讀玩家自己的檔。

⚠ 47 家店卻只有 46 個店名:類型 2(馬廄)開在地點 30 的那家,執行檔的名字指標是 NULL。
這不是解析漏掉,是原版資料就這樣。

## 問候語

`dword_553CC[店種][4]` 是四句問候語在 `SHOPPE.DAT` 裡的**位元組位移**,
`sub_111CC` 用 `random(0, 3)` 挑一句。

⚠ **武具店(類型 0)那四個位移都是 0** —— 它走另一條流程,還沒解。
不要拿位移 0 去讀(那會讀到 "Thanks for nothing!")。

## 佔位符(`sub_10FEC`)

| 符號 | 代換成 |
|---|---|
| `#` | 店名 |
| `$` | 店主 |
| `@` | 時段:`hour < 12` morning、`< 18` afternoon、其餘 evening |
| `%` | 數字(價格 / 數量) |
| `^` `*` | 尚未解出(要等物品表) |

同一個函式也負責展開 `.DAT` 的詞典 token:`dword_41794[b*4]`,而
`0x41794 + 0x80*4 = 0x41994 = dword_41990 + 4` —— **這正式證實了
`.DAT` 的 `slot = b - 0x7F`**(docs/re/05 當初是實測推出來的,現在有程式面佐證)。

## 騎馬

`sub_1B294` 開頭:`(byte_3E08C & 0xFE) == 0x12` 而且店種不是 0x83(馬廄)時,
回「A merchant says: "GET THAT HORSE OUT..."」。只有馬廄肯招呼騎馬的客人。

## 尚未實作

買賣、療傷、住宿、買馬買船的流程都要物品表與價格表 —— 還沒解。
引擎目前到「認出是哪家店 + 說問候語」為止,之後明說未實作而不是假裝成交。

## 驗收

- 47 家店全數解出,店名/店主與 `DATA.OVL` 逐筆相符,沒名字的正好只有已知那一家
- 站在不列顛城的三名商人旁邊按 T,分別得到
  酒館「The Wayfarer Tavern」+「…My name is Tika, thy host this afternoon.」、
  武具店「Iolo's Bows」(無問候語,符合類型 0 的已知情況)、
  旅店「The Wayfarer Inn」+「I am Donya, the innkeeper.」
