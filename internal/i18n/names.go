// Package i18n 是譯名覆蓋層:原版資料裡的英文名 → 繁體中文。
//
// # 為什麼是覆蓋層而不是改資料
//
// `[HARD]` 原版檔案唯讀。`.TLK` 與 `DATA.OVL` 的字串都靠位移表定位,
// 中文寫回去會把整張表推歪。所以譯名一律走這裡查表,`internal/u5data`
// 只負責把英文原文讀出來。
//
// # 譯名的權威順序
//
//  1. 《軟體世界》U5 中文說明書(org_game/manuel/)
//  2. 姊妹專案 u6-cht / u4-cht / u1-cht 的既有譯名 —— **系列共通名一律對齊,不另立新譯**
//  3. 現代直觀
//
// u6-cht 的對照表在 `~/u3-cht/u6-cht/dumps/glossary.json` 與
// `dumps/distilled_manual.md`(後者收了 U6 手冊 p42-57 的武器與怪物表)。
//
// # 一條硬規則
//
// **[HARD] 會拿去跟玩家輸入比對的字串,canonical 值維持英文。**
// 咒語名(玩家要打「An Nox」)、Yes/No、對話關鍵字(name/job/join)都屬於這一類。
// 這裡提供的是**顯示名**,比對照樣用 `u5data` 讀出來的英文原文。
package i18n

import "strings"

// Name 查一個英文名的中文譯名。查不到就原樣回傳 ——
// **不要回空字串**:那會讓畫面上出現空白而不是「還沒翻」,反而更難發現。
func Name(en string) string {
	if zh, ok := names[strings.TrimSpace(en)]; ok {
		return zh
	}
	return en
}

// Has 回報這個名字翻過了沒。用來統計覆蓋率。
func Has(en string) bool {
	_, ok := names[strings.TrimSpace(en)]
	return ok
}

// Bilingual 回傳「中文(English)」。
//
// 用在**玩家要打得出來**的東西上:咒語、符文。中文讓人看得懂,
// 英文讓人知道要按什麼鍵。譯名與原名相同時只回一份。
func Bilingual(en string) string {
	zh := Name(en)
	if zh == en {
		return en
	}
	return zh + "(" + en + ")"
}

// names 是全部的譯名。分區只是為了讀,查詢時是同一張表。
//
// ⚠ 新增條目前先查 u6-cht 的 glossary —— 系列共通名不另立新譯。
var names = map[string]string{}

func add(m map[string]string) {
	for k, v := range m {
		names[k] = v
	}
}

func init() {
	add(creatures)
	add(equipment)
	add(reagents)
	add(misc)
	add(people)
	add(places)
	add(npcs)
}

// creatures 是 `DATA.OVL` 0x1866 那張 48 筆的生物名表。
//
// 對齊來源:u6-cht 的 `distilled_manual.md`(U6 手冊 p46-57 的怪物表)。
// 兩代重疊的 15 種一律照抄,不重譯。
var creatures = map[string]string{
	// 人類(索引 0..15)—— 前 13 種同時是城鎮 NPC 的類別。
	"Mage":         "法師",
	"Bard":         "吟遊詩人",
	"Fighter":      "戰士",
	"Avatar":       "聖者",
	"Villager":     "村民",
	"Merchant":     "商人",
	"Jester":       "弄臣",
	"Child":        "孩童",
	"Beggar":       "乞丐",
	"Guard":        "衛兵",
	"Wanderer":     "遊民",
	"Blackthorn":   "黑棘",   // u6-cht:黑棘(不是「黑刺」)
	"Lord British": "不列顛王", // 系列共通
	"Shadow Lord":  "影主",   // u6-cht:shadowlords → 影主
	// 海上
	"Sea Horse":   "海馬",
	"Squid":       "大烏賊", // u6 手冊:Squid (giant) → 大烏賊
	"Sea Serpent": "海龍",  // u6 手冊
	"Shark":       "鯊魚",
	// 陸上
	"Giant Rat":    "巨鼠",  // u6 手冊:Rat → 巨鼠
	"Bat":          "大蝙蝠", // u6 手冊
	"Giant Spider": "巨蛛",  // u6 手冊:Spider → 巨蛛
	"Ghost":        "鬼",   // u6 手冊
	"Slime":        "黏怪",  // u6 手冊
	"Gremlin":      "小妖怪", // u6 手冊
	"Mimic":        "化形怪", // u6 手冊
	"Reaper":       "樹妖",  // u6 手冊
	"Gazer":        "多眼妖", // u6 手冊
	"Crawler":      "爬蟲",
	"Gargoyle":     "魔像族", // u6-cht 的選擇(手冊作「翼魔」,專案已定案用魔像族)
	"Insect Swarm": "蟲群",
	"Orc":          "半獸人",
	"Skeleton":     "骷髏怪", // u6 手冊
	"Python":       "巨蟒",
	"Ettin":        "雙頭巨人",
	"Headless":     "無頭怪", // u6 手冊
	"Wisp":         "幽影",  // u6-cht 的選擇(手冊作「幻靈」)
	"Daemon":       "惡魔",  // u6 手冊
	"Dragon":       "龍",   // u6 手冊
	"Sand Trap":    "流沙怪",
	"Troll":        "山怪",  // u6-cht 的選擇(手冊作「巨人」)
	"Mongbat":      "蝙猴",  // u6 手冊
	"Corpser":      "拖屍怪", // u6 手冊
	"Rot Worm":     "腐蟲",  // u6 手冊:Rotworm
}

// equipment 是 `DATA.OVL` 0x17E4 那張 48 筆的裝備名表。
//
// U5 的裝備分四段:頭盔 0-3、盾 4-8、護甲 9-15、武器 16-41、飾品 42-47。
// u6 手冊有的照抄,沒有的照系列習慣譯。
var equipment = map[string]string{
	// 頭盔
	"Leather Helm": "皮盔",
	"Chain Coif":   "鎖鏈頭巾",
	"Iron Helm":    "鐵盔",
	"Spiked Helm":  "尖釘盔",
	// 盾
	"Small Shield":  "小盾",
	"Large Shield":  "大盾",
	"Spiked Shield": "尖釘盾",
	"Magic Shield":  "魔法盾",
	"Jewel Shield":  "寶石盾",
	// 護甲
	"Cloth Armour":   "布衣",
	"Leather Armour": "皮甲",
	"Ring Mail":      "環甲",
	"Scale Mail":     "鱗甲",
	"Chain Mail":     "鎖子甲",
	"Plate Mail":     "板甲",
	"Mystic Armour":  "秘法護甲",
	// 武器
	"Dagger":         "匕首", // u6 手冊
	"Sling":          "投石索",
	"Club":           "棍棒",
	"Flaming Oil":    "燃燒油", // u6 手冊:Flask of Flaming Oil
	"Main Gauche":    "左手短劍",
	"Spear":          "長矛", // u6 手冊
	"Throwing Axe":   "投擲斧",
	"Short Sword":    "短劍",
	"Mace":           "釘頭錘", // u6 手冊:Mace → 鐵鎚 / 釘頭錘
	"Morning Star":   "流星錘", // u6 手冊:Morningstar
	"Bow":            "弓",   // u6 手冊
	"Arrows":         "箭",
	"Crossbow":       "十字弓", // u6 手冊
	"Quarrels":       "弩矢",
	"Long Sword":     "長劍",
	"2H Hammer":      "雙手大鐵鎚", // u6 手冊:Two-handed Hammer
	"2H Axe":         "雙手斧",
	"2H Sword":       "雙手劍",
	"Halberd":        "戟",
	"Sword of Chaos": "混沌之劍",
	"Magic Bow":      "魔法弓", // u6 手冊
	"Silver Sword":   "銀劍",
	"Magic Axe":      "魔斧", // u6 手冊
	"Glass Sword":    "玻璃劍",
	"Jeweled Sword":  "寶石劍",
	"Mystic Sword":   "秘法之劍",
	// 飾品
	"Ring of Invisibility": "隱形戒指",
	"Ring of Protection":   "防護戒指",
	"Ring of Regeneration": "再生戒指",
	"Amulet/Turning":       "轉化護符",
	"Spiked Collar":        "尖釘項圈",
	"Ankh":                 "安卡", // u6 手冊
}

// reagents 是八種施法藥草。
var reagents = map[string]string{
	"Sulfur Ash":  "硫磺灰",
	"Ginseng":     "人參",
	"Garlic":      "大蒜",
	"Spider Silk": "蛛絲",
	"Blood Moss":  "血苔",
	"Black Pearl": "黑珍珠",
	"Nightshade":  "顛茄",
	"Mandrake":    "曼德拉草",
}

// misc 是公會品項、酒、以及其他零散名詞。
var misc = map[string]string{
	"Keys":     "鑰匙", // u6-cht:key → 鑰匙
	"Gems":     "寶石", // u6 手冊
	"Torches":  "火把", // u6 手冊
	"Horse":    "馬",  // u6 手冊
	"Pirates":  "海盜",
	"Rose":     "玫瑰紅",
	"Claret":   "波爾多紅",
	"Sauterne": "索甸甜白",
	"Muscatel": "麝香甜酒",
	"Moselle":  "莫塞爾白",
	"Chablis":  "夏布利白",
}

// people 是原版寫死在資料裡的人名(不是 `.TLK` 裡的 NPC,那些另案)。
//
// 全部照 u6-cht 的 glossary。
var people = map[string]string{
	"Iolo":     "尤洛",
	"Shamino":  "夏米諾",
	"Dupre":    "杜普雷",
	"Geoffrey": "傑佛瑞",
	"Julia":    "茱莉亞",
	"Katrina":  "卡崔娜",
	"Mariah":   "瑪萊雅",
	"Jaana":    "亞娜",
	"Sentri":   "山特利",
	"Gwenno":   "葛雯",
	"Nystul":   "尼斯托",
	"Chuckles": "笑笑",
	"Sherry":   "雪莉",
	"Mondain":  "蒙丹",
	"Minax":    "米納克斯",
	"Exodus":   "出走者",
}

// places 是三十二個地點與八座地牢。
//
// ⚠ **與 u6-cht 對齊,而不是與本專案第一版對齊。** 五個地名因此換掉了:
//
//	Yew        紫杉城   → 尤伊
//	Moonglow   月光城   → 月華城
//	Jhelom     哲倫     → 傑隆
//	Skara Brae 史卡拉布雷 → 斯卡拉布雷
//	Cove       (未定)   → 海灣鎮
//
// 系列共通名不另立新譯,即使本專案先寫了別的。
var places = map[string]string{
	// 八大城
	"MOONGLOW":     "月華城",
	"BRITAIN":      "不列顛城",
	"JHELOM":       "傑隆",
	"YEW":          "尤伊",
	"MINOC":        "米諾克",
	"TRINSIC":      "特林希克",
	"SKARA BRAE":   "斯卡拉布雷",
	"NEW MAGINCIA": "新馬精西亞",
	// 燈塔
	"FOGSBANE":  "破霧塔",
	"STORMCROW": "暴風鴉塔",
	"GREYHAVEN": "灰港塔",
	"WAVEGUIDE": "導浪塔",
	// 村落與據點
	"IOLO'S HUT":      "尤洛小屋",
	"WEST BRITANNY":   "西不列塔尼",
	"NORTH BRITANNY":  "北不列塔尼",
	"EAST BRITANNY":   "東不列塔尼",
	"PAWS":            "波斯村",
	"COVE":            "海灣鎮",
	"BUCCANEER'S DEN": "海盜窩",
	"ARARAT":          "亞拉臘",
	"BORDERMARCH":     "邊境哨",
	"FARTHING":        "法辛",
	"WINDEMERE":       "溫德米爾",
	"STONEGATE":       "石門",
	// 城堡與大屋
	"THE LYCAEUM":    "學苑",
	"EMPATH ABBEY":   "共感修道院",
	"SERPENT'S HOLD": "巨蛇堡",
	// 八座地牢
	"DECEIT":   "欺瞞",
	"DESPISE":  "輕蔑",
	"DESTARD":  "毀滅",
	"WRONG":    "謬誤",
	"COVETOUS": "貪婪",
	"SHAME":    "羞恥",
	"HYTHLOTH": "海斯洛斯",
	"DOOM":     "末日",
	// 世界
	"Britannia":  "不列顛尼亞",
	"Underworld": "幽冥界",
}

// npcs 是四個 `.TLK` 檔裡 135 個 NPC 的名字。
//
// 對齊規則:
//
//   - **兩代共通的 14 個一律照 u6-cht**(Charlotte 夏綠蒂、Jaana 雅娜、
//     Katrina 卡翠娜、Sutek 蘇特克、Sin'Vraal 辛弗拉、Smith 史密斯…)。
//   - 其餘照音譯,盡量用系列既有的用字習慣(-son 森、-ton 頓、th 索/瑟)。
//   - **頭銜照譯不音譯**:Lord 領主、Lady 夫人、Sir 爵士、Judge 法官、
//     Captain 船長、Squire 侍從、Monsieur 先生。
//   - 有綽號的保留綽號(`Alistair the Bard` → 吟遊詩人阿利斯泰)。
//
// ⚠ 名字**同時是對話關鍵字**(玩家可以打 NPC 的名字問別人)。
// 比對照樣用英文原文 —— 這裡只換顯示。
var npcs = map[string]string{
	"Aleyn":                  "阿蘭",
	"Alistair the Bard":      "吟遊詩人阿利斯泰",
	"Ambrose":                "安布羅斯",
	"Annon":                  "安農",
	"Anthony":                "安東尼",
	"Ava":                    "艾娃",
	"Balinor":                "巴利諾",
	"Bandaii":                "班岱",
	"Barbra":                 "芭芭拉",
	"Bidney":                 "畢德尼",
	"Bullwier":               "布爾維爾",
	"Camile":                 "卡蜜兒",
	"Chamfort":               "尚福",
	"Charlotte":              "夏綠蒂", // u6-cht
	"Christopher":            "克里斯多福",
	"Cory":                   "柯瑞",
	"David":                  "大衛",
	"Delwyn":                 "戴爾溫",
	"Desiree":                "黛西蕾",
	"Donn Piatt":             "唐恩·皮亞特",
	"Drudgeworth":            "德魯吉沃斯",
	"Dufus":                  "杜佛斯",
	"Eb":                     "艾伯",
	"Elistaria":              "艾莉絲塔莉亞",
	"Emilly":                 "艾蜜莉",
	"Felespar":               "費勒斯帕",
	"Fenelon":                "費奈隆",
	"Fiona":                  "費歐娜",
	"Flain":                  "弗萊恩",
	"Flint":                  "弗林特",
	"Foulwell":               "佛威爾",
	"Froed":                  "弗洛德",
	"Fumiko":                 "文子", // 日式名,用漢字不音譯
	"Gallrot":                "加爾洛",
	"Gardner":                "加德納",
	"Glinkie":                "葛琳琪",
	"Goeth":                  "戈斯",
	"Gorn":                   "高恩",
	"Gregory":                "葛雷格里",
	"Grendel":                "格倫德爾",
	"Greymarch":              "灰境",
	"Greyson":                "葛雷森",
	"Gruman":                 "古魯曼",
	"Hardluck":               "苦命漢", // 綽號型的名字,照意思譯
	"Hassad":                 "哈薩德",
	"Jaana":                  "雅娜", // u6-cht
	"Jacqueline":             "賈桂琳",
	"Jennifer":               "珍妮佛",
	"Jeremy":                 "傑瑞米",
	"Jerone":                 "傑羅恩",
	"Jimmy":                  "吉米",
	"Johne, Captain Johne.":  "約恩船長",
	"Joshua":                 "約書亞",
	"Jotham":                 "喬瑟姆",
	"Judge Dryden":           "德萊登法官",
	"Julia":                  "茱莉亞",
	"Justin":                 "賈斯汀",
	"Kaiko":                  "佳子",  // 日式名
	"Katrina":                "卡翠娜", // u6-cht
	"Kindor":                 "金多",
	"Kraw":                   "克勞",
	"Kristi":                 "克莉絲蒂",
	"Kurt":                   "科特",
	"Lady Hayden.":           "海登夫人",
	"Lady Janell":            "珍奈爾夫人",
	"Lady Sahra":             "莎拉夫人",
	"Lady Tessa":             "泰莎夫人",
	"Landon":                 "藍登",
	"Leof":                   "里歐夫",
	"Leona":                  "莉歐娜",
	"Lord Dalgrin":           "達葛林領主",
	"Lord Kenneth":           "肯尼斯領主", // u6-cht:Kenneth 肯尼斯
	"Lord Malone":            "馬龍領主",
	"Lord Michael":           "麥可領主",
	"Lord R'hien":            "瑞恩領主",
	"Lord Seggallion":        "塞加里昂領主", // u6-cht
	"Lord Shalineth":         "夏利尼斯領主",
	"Lord Stuart the Hungry": "飢餓的史都華領主",
	"Malifora":               "瑪莉芙拉",
	"Malik":                  "馬利克",
	"Margaret":               "瑪格麗特",
	"Mario":                  "馬里歐",
	"Maxwell":                "麥斯威爾",
	"Monsieur Loubet":        "盧貝先生",
	"Phillip":                "菲利普",
	"Quintin":                "昆汀",
	"Rew":                    "魯伊",
	"Rollo":                  "羅洛",
	"Saduj":                  "薩杜傑", // 把 Judas 倒過來拼 —— 原文的伏筆,譯名不點破
	"Saul":                   "掃羅",
	"Scally":                 "史卡利",
	"Shirita":                "希莉塔",
	"Sin'Vraal":              "辛弗拉", // u6-cht
	"Sindar...":              "辛達",
	"Sir Adam the Torch":     "火炬亞當爵士",
	"Sir Arbuthnot":          "阿布斯諾爵士",
	"Sir Sean":               "肖恩爵士",
	"Sir Simon":              "賽門爵士",
	"Smith":                  "史密斯", // u6-cht(那匹會說話的馬)
	"Squire Jimmy":           "侍從吉米",
	"Stephen":                "史蒂芬",
	"Stillwelt":              "史提威特",
	"Sutek":                  "蘇特克", // u6-cht
	"Sven":                   "斯文",
	"Tactus":                 "塔克圖斯",
	"Telila":                 "泰莉拉",
	"Temme":                  "泰姆",
	"Terrance":               "泰倫斯",
	"Tetsuo":                 "哲雄", // 日式名
	"Thentis":                "森提斯",
	"Thorkin":                "索金",
	"Thorne":                 "索恩",
	"Thrud":                  "斯魯德",
	"Tierra":                 "蒂耶拉",
	"Tim":                    "提姆",
	"Toede":                  "托德",
	"Tomoka":                 "智香", // 日式名
	"Toshi":                  "俊",  // 日式名
	"Treanna":                "崔安娜",
	"Trian":                  "崔恩",
	"Vigil":                  "維吉爾",
	"Wartow":                 "沃托",
	"Weblock":                "韋布拉克",
	"Windmire":               "溫米爾",
	"Woolfe":                 "伍爾夫",
	"Yasuda":                 "安田", // 日式名
	"Zachariah":              "撒迦利亞",
}
