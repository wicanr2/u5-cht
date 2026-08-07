package i18n

// `CASTLE.TLK` #8–#21 對白譯文(第 07 批)
//
// ⚠ `TOWNE.TLK#23#e14` 的來源是殘片 `caug`(對白解析的邊界情況),
// 這裡**刻意不放進表裡** —— 查不到會退回原文,比印「文字損壞」這種
// 內部狀態給玩家好。同樣的錯在第 03 批也犯過一次。

func init() {
	addTalk(map[string]string{
		// CASTLE.TLK#8: 衛兵 (Guard)
		"CASTLE.TLK#8#desc": "體格魁梧、性情凶惡的衛兵。",
		"CASTLE.TLK#8#bye":  "滾開！",
		"CASTLE.TLK#8#q0":   "喂，你！汝不知此乃禁區乎？",
		"CASTLE.TLK#8#q1":   "給吾住嘴！汝無權上樓！",
		"CASTLE.TLK#8#q2":   "汝當自行離去，還是吾得把汝從城牆扔下去？",
		"CASTLE.TLK#8#q2n":  "吾聽不懂汝的意思。汝聽明白吾的話沒？吾說的是……",

		// CASTLE.TLK#9: 弄臣 (Jester)
		"CASTLE.TLK#9#desc": "活潑跳躍的弄臣！",
		"CASTLE.TLK#9#job":  "吾在此迎賓並取悅客人！",
		"CASTLE.TLK#9#bye":  "若汝還需吾之娛樂，儘管說一聲！",
		"CASTLE.TLK#9#e0": "歡迎，歡迎，歡迎。" +
			"歡迎來到吾等聖賢且至尊的國王陛下、其威勢遍及天下、不朽永生、目前行方不明、恐已遇害、坎塔布里基亞·不列顛王陛下的城堡！" +
			"以及其信賴、效忠、幫助、友善、有禮、親切、恭順、快樂、儉樸、勇敢、潔淨、敬虔的……" +
			"……宮廷弄臣，笑笑·邦波……" +
			"……就是吾！",
		"CASTLE.TLK#9#e9":   "正是。",
		"CASTLE.TLK#9#e10":  "就在這裡。",
		"CASTLE.TLK#9#e11":  "誰？",
		"CASTLE.TLK#9#q0":   "呵唷呵唷呵唷。呵唷呵唷呵唷。呵唷呵唷呵唷。跳跳、跳跳、跳跳、跳跳！汝喜歡嗎？",
		"CASTLE.TLK#9#q0n":  "哎呀，這次吾跳得更好！",
		"CASTLE.TLK#9#q0y":  "吾也是這麼想！",

		// CASTLE.TLK#10: 黑棘 (Blackthorn, Dark Lord)
		"CASTLE.TLK#10#desc": "暗黑帝王親臨！",
		"CASTLE.TLK#10#q0":   "歡迎，%A，汝之蒞臨令吾喜出望外。汝將在此逗留甚久乎？",
		"CASTLE.TLK#10#q0n":  "吾持相反之見！汝如此好心自投羅網，吾感謝不已。準備迎接汝之末日吧！",
		"CASTLE.TLK#10#q1":   "汝既已加入吾等之壓制陣營？",
		"CASTLE.TLK#10#q1n":  "那麼汝之自首最為體貼……衛兵們！逮捕此異教徒！",
		"CASTLE.TLK#10#q2":   "既然如此，汝必知吾等之暗語……汝說出來。",
		"CASTLE.TLK#10#q2n":  "不，恐怕那並非正確答案……衛兵們！逮捕此異教徒！",
		"CASTLE.TLK#10#q2y":  "很好！有汝相助，吾等將勢不可擋！請隨意在吾之城堡與領地巡遊！",
		"CASTLE.TLK#10#q3":   "誰敢逼近強大的黑棘？我看不然，再試一次！",

		// CASTLE.TLK#11: 蠻族 (Barbarian)
		"CASTLE.TLK#11#desc": "筋骨粗壯的野蠻人。",
		"CASTLE.TLK#11#bye":  "保重。",
		"CASTLE.TLK#11#q0":   "啊，願布羅姆神賜福！有靈魂願與吾同囚。汝願隨吾逃獄乎？",
		"CASTLE.TLK#11#q0n":  "傻瓜！",
		"CASTLE.TLK#11#q1":   "若汝助吾，汝可隨吾同行否？",
		"CASTLE.TLK#11#q1n":  "嗯……",
		"CASTLE.TLK#11#q2":   "吾逃獄多次，亦被捕多次。吾如掌中物一般熟悉此城堡！",
		"CASTLE.TLK#11#q3":   "深夜時分，吾等當從身後一扇秘門潛出。登上城頂，偷偷沿著北邊梯下。入黑棘之寢房，穿過秘門，往梯下行。向北穿過秘門，沿另一梯下行，從後門逃出！待天明時分城門放下，吾等逃離此地！聽明白了？",
		"CASTLE.TLK#11#q3y":  "取吾藏在火盆內的鑰匙，吾等起身逃吧！",

		// CASTLE.TLK#12: 囚犯 (Prisoner)
		"CASTLE.TLK#12#desc": "衣衫襤褸、瘦骨嶙峋、飽經摧殘的靈魂。",
		"CASTLE.TLK#12#job":  "吾在此已逗留數月，皆因吾破犯了德性之法！",
		"CASTLE.TLK#12#bye":  "莫離吾而去！請解吾之鎖鏈！",
		"CASTLE.TLK#12#e1":   "衛兵、惡魔與影主遍處其間！汝當速離，莫貽誤時機！",
		"CASTLE.TLK#12#e4":   "是呀，他們常來此地拳打腳踢，嘲弄於吾！",
		"CASTLE.TLK#12#e7":   "此等制度皆為荒謬，彼皆詆毀德性之基礎！",
		"CASTLE.TLK#12#e8":   "他們玷污了德性之精髓！",

		// CASTLE.TLK#13: 馬夫 (Stable Hand)
		"CASTLE.TLK#13#desc": "背略彎曲、和藹可親的男子。",
		"CASTLE.TLK#13#job":  "吾在此照看馬匹。",
		"CASTLE.TLK#13#bye":  "再見。",
		"CASTLE.TLK#13#e1":   "是呀，吾等有上等的純種馬匹。",
		"CASTLE.TLK#13#e3": "是呀，吾等有最上乘的。吾等不但有犁地馬與山地馬，還有高原純血種。" +
			"唯獨缺少瓦洛里亞戰馬，吾等仍在尋覓！",
		"CASTLE.TLK#13#e4":   "就只差那一種。",
		"CASTLE.TLK#13#e7":   "是呀，吾等有。",

		// CASTLE.TLK#14: 老魔法師 (Old Mage)
		"CASTLE.TLK#14#desc": "骨瘦如柴、面容詭異扭曲的老法師。",
		"CASTLE.TLK#14#job":  "吾在此乃欲助汝。",
		"CASTLE.TLK#14#bye":  "留心腳下。",
		"CASTLE.TLK#14#e1":   "試往北邊樓梯上行，可達寶座之間，但黑棘陛下通常不見訪客！",
		"CASTLE.TLK#14#e2":   "那是囚犯。地下監獄便是。",
		"CASTLE.TLK#14#e3":   "彼大抵在馬廄中。",
		"CASTLE.TLK#14#e4":   "正是吾。",
		"CASTLE.TLK#14#e5":   "廚師多在廚房。",
		"CASTLE.TLK#14#e6":   "弄臣之蹤跡甚難尋！",
		"CASTLE.TLK#14#e7":   "北邊樓梯便是。",
		"CASTLE.TLK#14#e8":   "登上大廳中的梯子。",
		"CASTLE.TLK#14#e9":   "沿樓梯回下。",
		"CASTLE.TLK#14#e10":  "往入口近處衛兵番所後的樓梯。",
		"CASTLE.TLK#14#e11":  "攀上城堡一隅之塔。",
		"CASTLE.TLK#14#e12":  "沿梯下行。",
		"CASTLE.TLK#14#q0":   "請簽名。",
		"CASTLE.TLK#14#q1":   "汝尋何物或何人？",

		// CASTLE.TLK#15: 廚子 (Fat Cook)
		"CASTLE.TLK#15#desc": "肥胖油膩的男子。",
		"CASTLE.TLK#15#job":  "吾在此製作堆給這群飢荒的蠢貨吃的食物。",
		"CASTLE.TLK#15#bye":  "啊，去他的吧。",
		"CASTLE.TLK#15#e1":   "一點老馬肉、幾個人類幼子，那種東西。",
		"CASTLE.TLK#15#e3":   "便宜又多呀！",
		"CASTLE.TLK#15#e5":   "難得捉到的時候啦！",
		"CASTLE.TLK#15#e7":   "吾在此做食物。",
		"CASTLE.TLK#15#e9":   "嗯，他們或許不真是蠢貨，但他們肚子確實餓得要死！",
		"CASTLE.TLK#15#q0":   "要試試？",
		"CASTLE.TLK#15#q0n":  "吾自己也不吃這東西。",
		"CASTLE.TLK#15#q0y":  "嚥……哽……作嘔……吐……是呀，吞嚥時頗為粗糙。",

		// CASTLE.TLK#16: 故事人 (Pompous Storyteller)
		"CASTLE.TLK#16#desc": "自大輕浮的男子。",
		"CASTLE.TLK#16#job":  "吾以有趣的異端審判故事取悅眾人。",
		"CASTLE.TLK#16#bye":  "叮鈴鈴鈴鈴！",
		"CASTLE.TLK#16#e1":   "何，前些日子，吾等有幸目睹一位年輕淑女被車裂極刑，真是精采之至！",
		"CASTLE.TLK#16#e4":   "她拒絕向飢渴衛兵獻殷勤！",
		"CASTLE.TLK#16#q0":   "這難道不爆笑？",
		"CASTLE.TLK#16#q0n":  "或許汝會更開心於此：有位訪客因為未能對王家弄臣的故事發笑，吾等便把廚師的小侄兒扔進醬缸裡頭！",
		"CASTLE.TLK#16#q1":   "汝願加入吾等下次之娛樂乎？",
		"CASTLE.TLK#16#q1n":  "小心自己，朋友！",
		"CASTLE.TLK#16#q1y":  "吾會去尋汝。",
		"CASTLE.TLK#16#q2":   "呵呵呵呵，哈哈哈哈，嘻嘻……汝覺得這個笑話如何？",
		"CASTLE.TLK#16#q2n":  "衛兵們！",
		"CASTLE.TLK#16#q2y":  "吾感謝汝。",

		// CASTLE.TLK#17: 盲眼法師 (Blind Wizard)
		"CASTLE.TLK#17#desc": "雙眼失明的老法師。",
		"CASTLE.TLK#17#job":  "吾與汝同為囚犯。",
		"CASTLE.TLK#17#bye":  "珍重。",
		"CASTLE.TLK#17#e0":   "吾自新馬精西亞之家被擄來此。",
		"CASTLE.TLK#17#e1":   "吾在彼過著平靜的日子。",
		"CASTLE.TLK#17#e3":   "直至影主尋著吾。",
		"CASTLE.TLK#17#e5":   "吾掌有黑棘所求之秘識。",
		"CASTLE.TLK#17#e8":   "那乃吾獨知之事。",
		"CASTLE.TLK#17#e11":  "吾不知汝言何意。",
		"CASTLE.TLK#17#q0":   "汝就是%A乎？",
		"CASTLE.TLK#17#q0n":  "那汝有何所求？",
		"CASTLE.TLK#17#q0y":  "哦，好。",
		"CASTLE.TLK#17#q1":   "何人在此？汝有何所求？",
		"CASTLE.TLK#17#q2":   "吾聞過大評議會，但汝何故問吾此事？",
		"CASTLE.TLK#17#q2n":  "吾毫不知情！",
		"CASTLE.TLK#17#q2y":  "她乃貴婦。若她曾向汝言及吾，則她必有大信心於汝。",
		"CASTLE.TLK#17#q3":   "吾推斷汝乃尋海斯洛斯地牢之力量之言。我等在此談話可安全乎？",
		"CASTLE.TLK#17#q3n":  "則吾且候。",
		"CASTLE.TLK#17#q3y":  "汝所求之言乃 IGNAVUS！",

		// CASTLE.TLK#18: 女農民 (Young Woman - Field Keeper)
		"CASTLE.TLK#18#desc": "年輕的女性。",
		"CASTLE.TLK#18#job":  "吾管理田地。",
		"CASTLE.TLK#18#bye":  "對了，有件事……",
		"CASTLE.TLK#18#e0": "異於他處之不幸，吾等之莊稼長勢茁壯真摯！",
		"CASTLE.TLK#18#e3":   "正是如此！",
		"CASTLE.TLK#18#e6":   "汝必知暗黑帝王之暴掠？",
		"CASTLE.TLK#18#e9": "黑棘遣影主往許多城鎮，惟吾等靠近不列顛王陛下之城堡，故得安全。",
		"CASTLE.TLK#18#e10":  "暗黑帝王啦，傻瓜！",
		"CASTLE.TLK#18#e13":  "不知何故，他們似不願太近王陛下城堡。",
		"CASTLE.TLK#18#q0":   "%A，汝於此妙日安好否？",
		"CASTLE.TLK#18#q1":   "嗯……汝之旅程如何？",
		"CASTLE.TLK#18#q2":   "珍重，%A，願汝前程似錦。",
		"CASTLE.TLK#18#q3":   "想必汝同意不列顛王乃當然之統治者？",
		"CASTLE.TLK#18#q3n":  "滾開，蠢賤！",
		"CASTLE.TLK#18#q3y":  "很好。",

		// CASTLE.TLK#19: 農夫 (Sweaty Farmer)
		"CASTLE.TLK#19#desc": "汗泥滿身的農夫。",
		"CASTLE.TLK#19#greet": "哎呀，小兄弟，怎麼樣？",
		"CASTLE.TLK#19#job":   "吾與吾的夥伴迪布斯一起種地。",
		"CASTLE.TLK#19#e1":    "那就是吾給克里斯多福的綽號。",
		"CASTLE.TLK#19#e2":    "他與吾一起種地。",
		"CASTLE.TLK#19#e4":    "是個粗活，但吾不會幹太久。",
		"CASTLE.TLK#19#e7":    "吾打算成為藝術家！",
		"CASTLE.TLK#19#e8":    "在那之前，吾就在此幹活！",
		"CASTLE.TLK#19#q0":    "再見！",

		// CASTLE.TLK#20: 年輕農夫 (Dashing Young Farmer)
		"CASTLE.TLK#20#desc": "朝氣蓬勃的年輕農夫。",
		"CASTLE.TLK#20#greet": "嘿，%A！",
		"CASTLE.TLK#20#job":   "吾在地裡耕作以維持生計。",
		"CASTLE.TLK#20#bye":   "多謝了，夥計。",
		"CASTLE.TLK#20#e0":    "啊，吾與菲利普談過。",
		"CASTLE.TLK#20#e1":    "他是吾的同事！",
		"CASTLE.TLK#20#e2":    "朋友。",
		"CASTLE.TLK#20#e4":    "吾以認真的態度對待工作，即便並不享受其中。",
		"CASTLE.TLK#20#e7":    "吾真正的樂趣在於寫幻想故事。",
		"CASTLE.TLK#20#e9":    "吾曾寫過一部叫《時間之歌》的史詩。",
		"CASTLE.TLK#20#e11":   "它已經出版了。",
		"CASTLE.TLK#20#e13":   "當然是經由始源工坊。",
		"CASTLE.TLK#20#q0":    "汝願購吾之書乎？",
		"CASTLE.TLK#20#q0n":   "則吾今後之助力將限矣！",
		"CASTLE.TLK#20#q0y":   "好！吾甚榮幸與此開明之士相識。",
		"CASTLE.TLK#20#q1":    "汝真是良善之人。",

		// CASTLE.TLK#21: 年輕農夫 (Young Farmer - Resistance Member)
		"CASTLE.TLK#21#desc": "年輕的農夫。",
		"CASTLE.TLK#21#job":   "吾乃農夫。",
		"CASTLE.TLK#21#bye":   "旅途平安。",
		"CASTLE.TLK#21#e0":    "未曾聽聞。",
		"CASTLE.TLK#21#e1":    "一位賢能之人……咳咳，咳咳！",
		"CASTLE.TLK#21#e2":    "予眾人！",
		"CASTLE.TLK#21#q0":    "汝之職業為何？",
		"CASTLE.TLK#21#q1":    "汝喜之乎？",
		"CASTLE.TLK#21#q2":    "汝為新律所助，或所害？",
		"CASTLE.TLK#21#q2n":   "吾曾言……",
		"CASTLE.TLK#21#q2y":   "哦。",
		"CASTLE.TLK#21#q3":    "則吾可冒言汝乃反對黑棘新律乎？",
		"CASTLE.TLK#21#q3n":   "原來如此。",
		"CASTLE.TLK#21#q4":    "汝知此乃異端邪說乎？",
		"CASTLE.TLK#21#q4n":   "正是！",
		"CASTLE.TLK#21#q5":    "汝知抵抗軍之事乎？",
		"CASTLE.TLK#21#q5n":   "嗯……",
		"CASTLE.TLK#21#q6":    "汝支持之乎？",
		"CASTLE.TLK#21#q6n":   "哦。",
		"CASTLE.TLK#21#q7":    "汝知其暗語乎？",
		"CASTLE.TLK#21#q7n":   "原來如此。",
		"CASTLE.TLK#21#q8":    "其為何？",
		"CASTLE.TLK#21#q8n":   "嗯……",
		"CASTLE.TLK#21#q9": "很好……對不起，問題太多，但吾等必須謹慎。" +
			"若汝與吾等同心，於午夜至井邊會合！向在場者言及暗語，但莫告他人，間諜眾多！",
	})
}
