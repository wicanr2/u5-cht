package i18n

// `CASTLE.TLK` 的譯文
//
// 語氣照《軟體世界》手冊的中世紀羊皮紙腔(汝 / 卿 / 吾),對齊 u4-cht / u6-cht。
// 專有名詞走 `Name()`,這裡只寫對白本身。
//
// key 的形狀與規則見 `talk.go`。工作清單用
// `u5dump talkwork gamedata out.md re_work/fmtowns/iso/U5_J` 產生
// (含原版全文,不入庫)。
//
// # 這一批的取捨
//
//   - `'Tis` / `Thou` / `Dost` 這類古語不逐字譯成「這是」「你」,
//     而是換成中文對應的文言口吻 —— 直譯會失去它與現代 NPC 的語氣差。
//   - 「the master」指的是失蹤的不列顛王,譯成「主上」;
//     `Lord British` 走既有譯名(不列顛王),不在這裡另翻。
//   - **感嘆號照留。** 原版的驚嘆密度是角色語氣的一部分,
//     中文標點統一用全形(倚天字庫的 SPCFONT 有,不會掉 fallback)。

func init() {
	addTalk(map[string]string{
		// id 1 ─ Alistair the Bard(吟遊詩人,城堡大廳)
		"CASTLE.TLK#1#desc":  "一名神情鬱鬱的樂師。",
		"CASTLE.TLK#1#greet": "願卿今日安好！",
		"CASTLE.TLK#1#job":   "吾以樂音提振眾人的心緒！",
		"CASTLE.TLK#1#bye":   "願卿日安，朋友！",
		"CASTLE.TLK#1#e1": "從前這裡是個歡快的地方，人人都能來此，把塵世的憂煩暫且擱下。" +
			"如今時勢已變，留下的只剩回憶！",
		"CASTLE.TLK#1#e2":  "那些好時光的回憶。",
		"CASTLE.TLK#1#e4":  "那時不列顛王治理這片土地，手腕堅定而公允。",
		"CASTLE.TLK#1#e6":  "君王要在權力與公義之間走得穩，難吶。",
		"CASTLE.TLK#1#e10": "吾等思念真正的王，也仍盼著他歸來！",

		// id 2 ─ Stephen(下層廚房的廚子)
		"CASTLE.TLK#2#desc": "一個高大開朗的男子，圍裙髒得厲害。",
		"CASTLE.TLK#2#job":  "吾是下層廚房的主廚。",
		"CASTLE.TLK#2#bye":  "願卿也能像吾一樣吃得飽足！",
		"CASTLE.TLK#2#e0":   "唉呀，卿還指望什麼！吾整天待在廚房裡呢！",
		"CASTLE.TLK#2#e2":   "吾最愛下廚，尤其是為宴席掌杓！",
		"CASTLE.TLK#2#e3":   "地方是小了些，備料倒是齊全。",
		"CASTLE.TLK#2#e4":   "唉，自從主上離去，就再沒辦過一場了。",
		"CASTLE.TLK#2#e6": "是啊，教人難過。不過吾仍盼他歸來 —— " +
			"好讓吾再一次把烤雉雞端到他面前！",
		"CASTLE.TLK#2#e8":  "那是他最愛的一道。",
		"CASTLE.TLK#2#e9":  "王國各地的珍饈，吾們這兒應有盡有！",
		"CASTLE.TLK#2#e10": "連雉雞都有呢！",
		"CASTLE.TLK#2#e12": "到處都有！",
	})
}
