package i18n

import "fmt"

// `SHOPPE.DAT` 商店對白的譯文覆蓋層
//
// # key 是**檔案位移**
//
// 商店對白沒有記錄編號 —— 原版 `sub_11168` 拿到的就是一個位元組位移,
// 從那裡讀到 NUL 為止(價目表 `DATA.OVL` 裡存的也是位移)。所以位移就是
// 天然的 key,而且穩定:資料檔是唯讀的。
//
// # 佔位符要原樣留著
//
// 譯文在**代換佔位符之前**接進來,所以下面這七個字元必須照抄進中文句子:
//
//	#  店名      $  店主      %  價格      &  物品名
//	*  地名      @  時段      ^  數量
//
// 少一個就少一個資訊;多打一個(例如中文裡本來就有的 `%`)會被當成價格。
//
// # 語氣
//
// 商店對白比 NPC 對話白話 —— 店家在做生意,不是在唸史詩。
// 用親切的市井口吻,不要「汝」「卿」;那組留給城堡與神職。
// 專有名詞(裝備、藥草、地名)由佔位符帶進來,已經是譯過的,這裡不要再翻一次。

// ShopKey 組出商店對白的 key。
func ShopKey(off int) string { return fmt.Sprintf("SHOPPE.DAT@%d", off) }

// Shop 查一段商店對白的譯文。查不到就回原文。
func Shop(off int, en string) string {
	if zh, ok := shops[off]; ok {
		return zh
	}
	return en
}

// ShopTranslated 回報這一段翻過了沒。
func ShopTranslated(off int) bool { _, ok := shops[off]; return ok }

// ShopCount 是已翻的段數。
func ShopCount() int { return len(shops) }

// shops 依 SHOPPE.DAT 的位元組位移索引。
var shops = map[int]string{}

// addShop 併入一批譯文。
func addShop(m map[int]string) {
	for k, v := range m {
		shops[k] = v
	}
}

// TimeOfDay 是時段字的中文(原版 `sub_10FEC` 的 `@` token)。
//
// 商店對白裡長這樣:「@好!我是 $,# 的老闆。」—— 譯文正確保留了 `@`,
// 但 `u5data.TimeOfDay` 回的是英文,不覆寫的話玩家看到「morning好!」。
//
// ⚠ 邊界照原版(`byte_3E08F < 0x0C` / `< 0x12`),不要照中文的作息直覺重畫 ——
// 「下午」在原版是 12 時到 18 時,不是 13 時起。
func TimeOfDay(hour int) string {
	switch {
	case hour < 12:
		return "早"
	case hour < 18:
		return "午"
	}
	return "晚"
}
