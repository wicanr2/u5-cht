package i18n

// `SIGNS.DAT` 的招牌與墓碑譯文。key 尾巴的數字是第幾列(從 0 起)。
// 框線(8 9 : ; a b c d e f h i j k l m n g)照原樣保留。
// 文字列: g + 內容(14欄,中文字=2欄,空格=1欄) + g

func init() {
	addSign(map[string]string{
		// sign#0#95#148#0 - 方位標示(北 東 南)
		"sign#0#95#148#0": "8lllllmllmlllll9",
		"sign#0#95#148#1": "g   北不列顛城  g",
		"sign#0#95#148#2": "g    東波斯村   g",
		"sign#0#95#148#3": "g   南特林希克  g",
		"sign#0#95#148#4": ":lllllnllnlllll;",

		// sign#0#76#113#0 - 方位標示(北 南)
		"sign#0#76#113#0": "8lllllmllmlllll9",
		"sign#0#76#113#1": "g   北不列顛城  g",
		"sign#0#76#113#2": "g    南波斯村   g",
		"sign#0#76#113#3": ":lllllnllnlllll;",

		// sign#0#85#78#0 - 地形標示
		"sign#0#85#78#0": "8llllmlllll9  ",
		"sign#0#85#78#1": "g 巨蛇堡 :l9",
		"sign#0#85#78#2": "g 山道 8lllll;",
		"sign#0#85#78#3": ":llllnl;      ",

		// sign#0#195#41#0 - 警告標示
		"sign#0#195#41#0": " 8lllllmlllll9 ",
		"sign#0#195#41#1": "8;* 謹慎! *:9",
		"sign#0#195#41#2": "g   荒漠之地   g",
		"sign#0#195#41#3": ":llllllnllllll;",

		// sign#0#245#6#0 - 私人島嶼
		"sign#0#245#6#0": "8llllllmlllllll9",
		"sign#0#245#6#1": "g 私人島嶼    g",
		"sign#0#245#6#2": "g              g",
		"sign#0#245#6#3": "g 遊客謝絕    g",
		"sign#0#245#6#4": "g   入內      g",
		"sign#0#245#6#5": "g              g",
		"sign#0#245#6#6": "g   滾開來     g",
		"sign#0#245#6#7": ":llllllnlllllll;",

		// sign#0#233#242#0 - 聖跡尋求條件
		"sign#0#233#242#0": "8lllllllllllll9",
		"sign#0#233#242#1": "g             g",
		"sign#0#233#242#2": "g 汝知悉 可方g",
		"sign#0#233#242#3": "g  得入若係  g",
		"sign#0#233#242#4": "g 神聖之任務  g",
		"sign#0#233#242#5": "g     上      g",
		"sign#0#233#242#6": "g             g",
		"sign#0#233#242#7": ":lllllllllllll;",

		// sign#0#55#66#0 - 森林警告
		"sign#0#55#66#0": "8lllllmlllll9",
		"sign#0#55#66#1": "g           g",
		"sign#0#55#66#2": "g 謹慎深林  g",
		"sign#0#55#66#3": "g           g",
		"sign#0#55#66#4": ":lllllnlllll;",

		// sign#0#45#172#0 - 舍洛克島
		"sign#0#45#172#0": "8lllllmlllll9",
		"sign#0#45#172#1": "g           g",
		"sign#0#45#172#2": "g 此乃如下島g",
		"sign#0#45#172#3": "g  名舍洛克  g",
		"sign#0#45#172#4": "g   其人     g",
		"sign#0#45#172#5": "g  怪誕奇異  g",
		"sign#0#45#172#6": "g           g",
		"sign#0#45#172#7": ":lllllnlllll;",

		// sign#0#54#143#0 - 不列顛王探險紀
		"sign#0#54#143#0": "abbbbbmbbmbbbbbc",
		"sign#0#54#143#1": "g 茲始開啟  g",
		"sign#0#54#143#2": "g 陛下之聖 g",
		"sign#0#54#143#3": "g  跡以探索  g",
		"sign#0#54#143#4": "g  勘定於新  g",
		"sign#0#54#143#5": "g  幽冥界起  g",
		"sign#0#54#143#6": "g  月日   g",
		"sign#0#54#143#7": "deeeeeneeneeeeef",

		// sign#0#103#223#0 - 墓碑: 梅里丁
		"sign#0#103#223#0": "     hlllli     ",
		"sign#0#103#223#1": "hllllk    jlllli",
		"sign#0#103#223#2": "g 於此眠臥 g",
		"sign#0#103#223#3": "g 勇敢騎士 g",
		"sign#0#103#223#4": "g          g",
		"sign#0#103#223#5": "g   梅里丁   g",
		"sign#0#103#223#6": "g          g",
		"sign#0#103#223#7": "g 安眠  137年g",
		"sign#0#103#223#8": "jlllli    hllllk",
		"sign#0#103#223#9": "     g    g    ",
		"sign#0#103#223#10": "     g    g    ",

		// sign#0#104#223#0 - 墓碑: 亞里歐尼斯
		"sign#0#104#223#0": "     hlllli     ",
		"sign#0#104#223#1": "hllllk    jlllli",
		"sign#0#104#223#2": "g 於此眠臥 g",
		"sign#0#104#223#3": "g 勇敢騎士 g",
		"sign#0#104#223#4": "g          g",
		"sign#0#104#223#5": "g 亞里歐尼斯g",
		"sign#0#104#223#6": "g          g",
		"sign#0#104#223#7": "g 安眠  137年g",
		"sign#0#104#223#8": "jlllli    hllllk",
		"sign#0#104#223#9": "     g    g    ",
		"sign#0#104#223#10": "     g    g    ",

		// sign#0#105#223#0 - 墓碑: 洛因
		"sign#0#105#223#0": "     hlllli     ",
		"sign#0#105#223#1": "hllllk    jlllli",
		"sign#0#105#223#2": "g 於此眠臥 g",
		"sign#0#105#223#3": "g 勇敢騎士 g",
		"sign#0#105#223#4": "g          g",
		"sign#0#105#223#5": "g    洛因    g",
		"sign#0#105#223#6": "g          g",
		"sign#0#105#223#7": "g 安眠 135年 g",
		"sign#0#105#223#8": "jlllli    hllllk",
		"sign#0#105#223#9": "     g    g    ",
		"sign#0#105#223#10": "     g    g    ",

		// sign#0#106#223#0 - 墓碑: 諾因
		"sign#0#106#223#0": "     hlllli     ",
		"sign#0#106#223#1": "hllllk    jlllli",
		"sign#0#106#223#2": "g 於此眠臥 g",
		"sign#0#106#223#3": "g 勇敢騎士 g",
		"sign#0#106#223#4": "g          g",
		"sign#0#106#223#5": "g    諾因    g",
		"sign#0#106#223#6": "g          g",
		"sign#0#106#223#7": "g 安眠  137年g",
		"sign#0#106#223#8": "jlllli    hllllk",
		"sign#0#106#223#9": "     g    g    ",
		"sign#0#106#223#10": "     g    g    ",

		// sign#0#107#223#0 - 墓碑: 傑拉西
		"sign#0#107#223#0": "     hlllli     ",
		"sign#0#107#223#1": "hllllk    jlllli",
		"sign#0#107#223#2": "g 於此眠臥 g",
		"sign#0#107#223#3": "g 勇敢騎士 g",
		"sign#0#107#223#4": "g          g",
		"sign#0#107#223#5": "g   傑拉西   g",
		"sign#0#107#223#6": "g          g",
		"sign#0#107#223#7": "g 安眠  137年g",
		"sign#0#107#223#8": "jlllli    hllllk",
		"sign#0#107#223#9": "     g    g    ",
		"sign#0#107#223#10": "     g    g    ",

		// sign#1#14#20#0 - 黑棘法律：誠實
		"sign#1#14#20#0": "abbbbbbbbbbbbbbc",
		"sign#1#14#20#1": "g  黑棘法律   g",
		"sign#1#14#20#2": "g    誠 實    g",
		"sign#1#14#20#3": "gllllllllllllllg",
		"sign#1#14#20#4": "g 汝不得妄言 g",
		"sign#1#14#20#5": "g 否則汝將失 g",
		"sign#1#14#20#6": "g  其舌頭矣  g",
		"sign#1#14#20#7": "deeeeeeeeeeeeeef",

		// sign#1#16#20#0 - 黑棘法律：誠實(重複)
		"sign#1#16#20#0": "abbbbbbbbbbbbbbc",
		"sign#1#16#20#1": "g  黑棘法律   g",
		"sign#1#16#20#2": "g    誠 實    g",
		"sign#1#16#20#3": "gllllllllllllllg",
		"sign#1#16#20#4": "g 汝不得妄言 g",
		"sign#1#16#20#5": "g 否則汝將失 g",
		"sign#1#16#20#6": "g  其舌頭矣  g",
		"sign#1#16#20#7": "deeeeeeeeeeeeeef",

		// sign#2#11#0#0 - 私人房間
		"sign#2#11#0#0": "abbbbbbbbbbbbbc",
		"sign#2#11#0#1": "g   私人居室   g",
		"sign#2#11#0#2": "g  禁止進入   g",
		"sign#2#11#0#3": "deeeeeeeeeeeeef",

		// sign#2#16#27#0 - 黑棘法律：慈悲
		"sign#2#16#27#0": "abbbbbbbbbbbbbbc",
		"sign#2#16#27#1": "g  黑棘法律   g",
		"sign#2#16#27#2": "g    慈 悲    g",
		"sign#2#16#27#3": "gllllllllllllllg",
		"sign#2#16#27#4": "g  汝應救濟   g",
		"sign#2#16#27#5": "g  陷入困厄者  g",
		"sign#2#16#27#6": "g 或汝將自陷g",
		"sign#2#16#27#7": "g  相同困厄內  g",
		"sign#2#16#27#8": "deeeeeeeeeeeeeef",

		// sign#2#14#27#0 - 黑棘法律：慈悲(重複)
		"sign#2#14#27#0": "abbbbbbbbbbbbbbc",
		"sign#2#14#27#1": "g  黑棘法律   g",
		"sign#2#14#27#2": "g    慈 悲    g",
		"sign#2#14#27#3": "gllllllllllllllg",
		"sign#2#14#27#4": "g  汝應救濟   g",
		"sign#2#14#27#5": "g  陷入困厄者  g",
		"sign#2#14#27#6": "g 或汝將自陷g",
		"sign#2#14#27#7": "g  相同困厄內  g",
		"sign#2#14#27#8": "deeeeeeeeeeeeeef",

		// sign#3#15#28#0 - 黑棘法律：勇氣
		"sign#3#15#28#0": "abbbbbbbbbbbbbbc",
		"sign#3#15#28#1": "g  黑棘法律   g",
		"sign#3#15#28#2": "g   勇 氣    g",
		"sign#3#15#28#3": "gllllllllllllllg",
		"sign#3#15#28#4": "g  汝應勇敢   g",
		"sign#3#15#28#5": "g  奮戰至死   g",
		"sign#3#15#28#6": "g 或汝將被逐g",
		"sign#3#15#28#7": "g   被 貶 為   g",
		"sign#3#15#28#8": "g    懦 夫    g",
		"sign#3#15#28#9": "deeeeeeeeeeeeeef",

		// sign#4#14#5#0 - 寬恕
		"sign#4#14#5#0": "8llllmlllmllll9",
		"sign#4#14#5#1": "g 願入者知悉 g",
		"sign#4#14#5#2": "g   得蒙寬恕   g",
		"sign#4#14#5#3": ":llllnlllnllllk",

		// sign#4#17#17#0 - 政治囚禁
		"sign#4#17#17#0": "8llllllmllllll9",
		"sign#4#17#17#1": "g  政治囚犯   g",
		"sign#4#17#17#2": "jllllllnllllllk",

		// sign#4#12#4#0 - 可憐的弗雷德
		"sign#4#12#4#0": "8lllllllllllll9",
		"sign#4#12#4#1": "g     哀哉    g",
		"sign#4#12#4#2": "g  可憐的弗雷德 g",
		"sign#4#12#4#3": "g   他失掉頭顱 g",
		"sign#4#12#4#4": "jlllllllllllllk",

		// sign#4#11#2#0 - 派特死於鼠咬
		"sign#4#11#2#0": "8lllllllllllll9",
		"sign#4#11#2#1": "g   於此眠臥   g",
		"sign#4#11#2#2": "g     派 特    g",
		"sign#4#11#2#3": "g   被鼠吞噬   g",
		"sign#4#11#2#4": "jlllllllllllllk",

		// sign#4#13#1#0 - 黑棘弄臣上半身
		"sign#4#13#1#0": "8llllllllllllll9",
		"sign#4#13#1#1": "g   於此眠臥    g",
		"sign#4#13#1#2": "g   黑 棘 老    g",
		"sign#4#13#1#3": "g    弄 臣 之    g",
		"sign#4#13#1#4": "g     上 半 身    g",
		"sign#4#13#1#5": "jllllllllllllllk",

		// sign#4#14#1#0 - 黑棘弄臣下半身
		"sign#4#14#1#0": "8llllllllllllll9",
		"sign#4#14#1#1": "g   於此眠臥    g",
		"sign#4#14#1#2": "g   黑 棘 老    g",
		"sign#4#14#1#3": "g    弄 臣 之    g",
		"sign#4#14#1#4": "g     下 半 身    g",
		"sign#4#14#1#5": "jllllllllllllllk",

		// sign#4#17#3#0 - 海伍德的墓碑
		"sign#4#17#3#0": "8lllllllllllll9",
		"sign#4#17#3#1": "g   於此眠臥   g",
		"sign#4#17#3#2": "g    海伍德    g",
		"sign#4#17#3#3": "g  尚非死木   g",
		"sign#4#17#3#4": "g     而 已    g",
		"sign#4#17#3#5": "g   死木爾矣   g",
		"sign#4#17#3#6": "jlllllllllllllk",

		// sign#4#18#4#0 - 萊夫的墓碑
		"sign#4#18#4#0": "     hllli",
		"sign#4#18#4#1": "hllllk   jlllli",
		"sign#4#18#4#2": "g   下埋著萊夫  g",
		"sign#4#18#4#3": "g  誰實為小賊也 g",
		"sign#4#18#4#4": "jlllli   hllllk",
		"sign#4#18#4#5": "     g   g",

		// sign#4#18#1#0 - 黃堡的墓碑
		"sign#4#18#1#0": "     hllli",
		"sign#4#18#1#1": "hllllk   jlllli",
		"sign#4#18#1#2": "g   於此眠臥   g",
		"sign#4#18#1#3": "g   黃堡冰冷   g",
		"sign#4#18#1#4": "g   如冰山矣   g",
		"sign#4#18#1#5": "jlllli   hllllk",
		"sign#4#18#1#6": "     g   g",

		// sign#4#16#2#0 - 約翰死於刑具
		"sign#4#16#2#0": "    hllli    ",
		"sign#4#16#2#1": "hlllk   jllli",
		"sign#4#16#2#2": "g  約翰多邪惡  g",
		"sign#4#16#2#3": "g 死於刑具上   g",
		"sign#4#16#2#4": "jllll   hlllk",
		"sign#4#16#2#5": "    g   g    ",

		// sign#4#14#3#0 - 傑克死於刑架
		"sign#4#14#3#0": "     hllli    ",
		"sign#4#14#3#1": "hllllk   jlllli",
		"sign#4#14#3#2": "g   於此眠臥   g",
		"sign#4#14#3#3": "g    傑 克    g",
		"sign#4#14#3#4": "g   被拉扯於   g",
		"sign#4#14#3#5": "g    刑 架 上   g",
		"sign#4#14#3#6": "jlllli   hllllk",
		"sign#4#14#3#7": "     g   g    ",

		// sign#4#17#21#0 - 通緝
		"sign#4#17#21#0": "abbbbbbbbbbbbbc",
		"sign#4#17#21#1": "g   通緝:    g",
		"sign#4#17#21#2": "g             g",
		"sign#4#17#21#3": "g             g",
		"sign#4#17#21#4": "g             g",
		"sign#4#17#21#5": "g             g",
		"sign#4#17#21#6": "g   生死不論   g",
		"sign#4#17#21#7": "deeeeeeeeeeeeef",

		// sign#4#16#28#0 - 黑棘法律：正義
		"sign#4#16#28#0": "abbbbbbbbbbbbbbc",
		"sign#4#16#28#1": "g  黑棘法律   g",
		"sign#4#16#28#2": "g    正 義    g",
		"sign#4#16#28#3": "gllllllllllllllg",
		"sign#4#16#28#4": "g  汝應認罪   g",
		"sign#4#16#28#5": "g  並受正當審g",
		"sign#4#16#28#6": "g  判 或 汝 將 g",
		"sign#4#16#28#7": "g   被 判 死 刑  g",
		"sign#4#16#28#8": "deeeeeeeeeeeeeef",

		// sign#4#14#28#0 - 黑棘法律：正義(重複)
		"sign#4#14#28#0": "abbbbbbbbbbbbbbc",
		"sign#4#14#28#1": "g  黑棘法律   g",
		"sign#4#14#28#2": "g    正 義    g",
		"sign#4#14#28#3": "gllllllllllllllg",
		"sign#4#14#28#4": "g  汝應認罪   g",
		"sign#4#14#28#5": "g  並受正當審g",
		"sign#4#14#28#6": "g  判 或 汝 將 g",
		"sign#4#14#28#7": "g   被 判 死 刑  g",
		"sign#4#14#28#8": "deeeeeeeeeeeeeef",

		// sign#5#20#14#0 - 無助者的使命
		"sign#5#20#14#0": "abbbbbbbbbbbbbbc",
		"sign#5#20#14#1": "g  無助者的   g",
		"sign#5#20#14#2": "g    使 命    g",
		"sign#5#20#14#3": "deeeeeeeeeeeeeef",

		// sign#5#14#28#0 - 黑棘法律：犧牲
		"sign#5#14#28#0": "abbbbbbbbbbbbbbc",
		"sign#5#14#28#1": "g  黑棘法律   g",
		"sign#5#14#28#2": "g    犧 牲    g",
		"sign#5#14#28#3": "gllllllllllllllg",
		"sign#5#14#28#4": "g  汝應捐獻   g",
		"sign#5#14#28#5": "g  收 入 之 半   g",
		"sign#5#14#28#6": "g 予慈善 或汝g",
		"sign#5#14#28#7": "g  將無收入矣  g",
		"sign#5#14#28#8": "deeeeeeeeeeeeeef",

		// sign#5#16#28#0 - 黑棘法律：犧牲(重複)
		"sign#5#16#28#0": "abbbbbbbbbbbbbbc",
		"sign#5#16#28#1": "g  黑棘法律   g",
		"sign#5#16#28#2": "g    犧 牲    g",
		"sign#5#16#28#3": "gllllllllllllllg",
		"sign#5#16#28#4": "g  汝應捐獻   g",
		"sign#5#16#28#5": "g  收 入 之 半   g",
		"sign#5#16#28#6": "g 予慈善 或汝g",
		"sign#5#16#28#7": "g  將無收入矣  g",
		"sign#5#16#28#8": "deeeeeeeeeeeeeef",

		// sign#6#15#28#0 - 黑棘法律：榮譽
		"sign#6#15#28#0": "abbbbbbbbbbbbbbc",
		"sign#6#15#28#1": "g  黑棘法律   g",
		"sign#6#15#28#2": "g    榮 譽    g",
		"sign#6#15#28#3": "gllllllllllllllg",
		"sign#6#15#28#4": "g 若汝失榮譽 g",
		"sign#6#15#28#5": "g 則汝應取汝 g",
		"sign#6#15#28#6": "g  自 己 的 生   g",
		"sign#6#15#28#7": "g    命 矣    g",
		"sign#6#15#28#8": "deeeeeeeeeeeeeef",

		// sign#7#5#1#0 - 吟遊詩人之末
		"sign#7#5#1#0": "     hllli    ",
		"sign#7#5#1#1": "hllllk   jlllli",
		"sign#7#5#1#2": "g   於此眠臥   g",
		"sign#7#5#1#3": "g  詩 人 之 話   g",
		"sign#7#5#1#4": "g   末 路 已 至   g",
		"sign#7#5#1#5": "jlllli   hllllk",
		"sign#7#5#1#6": "     g   g    ",

		// sign#7#15#25#0 - 黑棘法律：靈性
		"sign#7#15#25#0": "abbbbbbbbbbbbbbc",
		"sign#7#15#25#1": "g  黑棘法律   g",
		"sign#7#15#25#2": "g    靈 性    g",
		"sign#7#15#25#3": "gllllllllllllllg",
		"sign#7#15#25#4": "g  汝應施行   g",
		"sign#7#15#25#5": "g  美德之法律 g",
		"sign#7#15#25#6": "g 或汝將死作g",
		"sign#7#15#25#7": "g   異 教 徒 矣   g",
		"sign#7#15#25#8": "deeeeeeeeeeeeeef",

		// sign#8#15#27#0 - 新馬精西亞重建
		"sign#8#15#27#0": "abbbbbbbbbbbbbbc",
		"sign#8#15#27#1": "g  新馬精西亞  g",
		"sign#8#15#27#2": "g              g",
		"sign#8#15#27#3": "g    重 建     g",
		"sign#8#15#27#4": "g  136年 3月  g",
		"sign#8#15#27#5": "deeeeeeeeeeeeeef",

		// sign#8#19#7#0 - 馬精西亞市民墓地
		"sign#8#19#7#0": "abbbbbbbbbbbbbc",
		"sign#8#19#7#1": "g   於此眠臥   g",
		"sign#8#19#7#2": "g 昔日馬精西亞g",
		"sign#8#19#7#3": "g    的市民    g",
		"sign#8#19#7#4": "g 願彼等安寧   g",
		"sign#8#19#7#5": "g   安眠矣    g",
		"sign#8#19#7#6": "deeeeeeeeeeeeef",

		// sign#8#18#1#0 - 海伍德
		"sign#8#18#1#0": "8lllllllllllll9",
		"sign#8#18#1#1": "g   海伍德    g",
		"sign#8#18#1#2": "g   136年    g",
		"sign#8#18#1#3": "jlllllllllllllk",

		// sign#8#20#1#0 - 斯利姆
		"sign#8#20#1#0": "     hllli",
		"sign#8#20#1#1": "hllllk   jlllli",
		"sign#8#20#1#2": "g    斯利姆    g",
		"sign#8#20#1#3": "g   136年    g",
		"sign#8#20#1#4": "jlllli   hllllk",
		"sign#8#20#1#5": "     g   g",

		// sign#8#22#1#0 - 班特
		"sign#8#22#1#0": "     hllli",
		"sign#8#22#1#1": "hllllk   jlllli",
		"sign#8#22#1#2": "g    班 特    g",
		"sign#8#22#1#3": "g   136年    g",
		"sign#8#22#1#4": "jlllli   hllllk",
		"sign#8#22#1#5": "     g   g",

		// sign#8#24#1#0 - 德米特里
		"sign#8#24#1#0": "     hllli",
		"sign#8#24#1#1": "hllllk   jlllli",
		"sign#8#24#1#2": "g   德米特里   g",
		"sign#8#24#1#3": "g   136年    g",
		"sign#8#24#1#4": "jlllli   hllllk",
		"sign#8#24#1#5": "     g   g",

		// sign#8#26#1#0 - 布洛斯
		"sign#8#26#1#0": "8lllllllllllll9",
		"sign#8#26#1#1": "g   布洛斯    g",
		"sign#8#26#1#2": "g   136年    g",
		"sign#8#26#1#3": "jlllllllllllllk",

		// sign#8#28#1#0 - 魯欣
		"sign#8#28#1#0": "8lllllllllllll9",
		"sign#8#28#1#1": "g    魯欣     g",
		"sign#8#28#1#2": "g   136年    g",
		"sign#8#28#1#3": "jlllllllllllllk",

		// sign#8#20#20#0 - 黑棘法律：謙遜
		"sign#8#20#20#0": "abbbbbbbbbbbbbbc",
		"sign#8#20#20#1": "g  黑棘法律   g",
		"sign#8#20#20#2": "g    謙 遜    g",
		"sign#8#20#20#3": "gllllllllllllllg",
		"sign#8#20#20#4": "g  汝應謙卑   g",
		"sign#8#20#20#5": "g  於  汝  之 g",
		"sign#8#20#20#6": "g  上  位  者  g",
		"sign#8#20#20#7": "g 或汝將遭其g",
		"sign#8#20#20#8": "g   怒 火 焚 燒  g",
		"sign#8#20#20#9": "deeeeeeeeeeeeeef",

		// sign#8#20#17#0 - 黑棘法律：謙遜(重複)
		"sign#8#20#17#0": "abbbbbbbbbbbbbbc",
		"sign#8#20#17#1": "g  黑棘法律   g",
		"sign#8#20#17#2": "g    謙 遜    g",
		"sign#8#20#17#3": "gllllllllllllllg",
		"sign#8#20#17#4": "g  汝應謙卑   g",
		"sign#8#20#17#5": "g  於  汝  之 g",
		"sign#8#20#17#6": "g  上  位  者  g",
		"sign#8#20#17#7": "g 或汝將遭其g",
		"sign#8#20#17#8": "g   怒 火 焚 燒  g",
		"sign#8#20#17#9": "deeeeeeeeeeeeeef",

		// sign#9#12#28#0 - 破霧塔歡迎
		"sign#9#12#28#0": "8lllllllmllllll9",
		"sign#9#12#28#1": "g              g",
		"sign#9#12#28#2": "g    破霧塔    g",
		"sign#9#12#28#3": "g              g",
		"sign#9#12#28#4": "g    訪客歡迎   g",
		"sign#9#12#28#5": "g              g",
		"sign#9#12#28#6": "jlllllllnllllllk",

		// sign#11#18#25#0 - 灰港塔介紹
		"sign#11#18#25#0": "8lllllllmllllll9",
		"sign#11#18#25#1": "g              g",
		"sign#11#18#25#2": "g    灰港塔    g",
		"sign#11#18#25#3": "g              g",
		"sign#11#18#25#4": "g  美德之光    g",
		"sign#11#18#25#5": "g              g",
		"sign#11#18#25#6": "jlllllllnllllllk",

		// sign#14#1#1#0 - 可憐理查德
		"sign#14#1#1#0": "8llllllllllllll9",
		"sign#14#1#1#1": "g              g",
		"sign#14#1#1#2": "g   於此眠臥   g",
		"sign#14#1#1#3": "g  可憐的理查德 g",
		"sign#14#1#1#4": "g  活埋而死欲速 g",
		"sign#14#1#1#5": "g   成 終 極 五  g",
		"sign#14#1#1#6": "g              g",
		"sign#14#1#1#7": "jllllllllllllllk",

		// sign#14#3#1#0 - 列布林的棺木
		"sign#14#3#1#0": "     hlllli     ",
		"sign#14#3#1#1": "hllllk    jlllli",
		"sign#14#3#1#2": "g   於此眠臥   g",
		"sign#14#3#1#3": "g  列布林在棺內  g",
		"sign#14#3#1#4": "g              g",
		"sign#14#3#1#5": "g 龍毆過頻繁   g",
		"sign#14#3#1#6": "g              g",
		"sign#14#3#1#7": "jlllli    hllllk",
		"sign#14#3#1#8": "     g    g     ",

		// sign#14#5#1#0 - 飲酒駕駛警示墓碑
		"sign#14#5#1#0": "8llllllllllllll9",
		"sign#14#5#1#1": "g 彼已不活     g",
		"sign#14#5#1#2": "g 他欲飲酒與   g",
		"sign#14#5#1#3": "g  駕駛(馬車)  g",
		"sign#14#5#1#4": "g              g",
		"sign#14#5#1#5": "g 此為公共   g",
		"sign#14#5#1#6": "g   服務墓碑    g",
		"sign#14#5#1#7": "jllllllllllllllk",

		// sign#17#6#13#0 - 皇家監獄
		"sign#17#6#13#0": "abbbbbbbbbbbbbbc",
		"sign#17#6#13#1": "g              g",
		"sign#17#6#13#2": "g   皇家獄所   g",
		"sign#17#6#13#3": "g              g",
		"sign#17#6#13#4": "deeeeeeeeeeeeeef",

		// sign#17#24#13#0 - 北星兵工廠
		"sign#17#24#13#0": "abbbbbbbbbbbbbbc",
		"sign#17#24#13#1": "g              g",
		"sign#17#24#13#2": "g   北極星軍械 g",
		"sign#17#24#13#3": "g    庫房      g",
		"sign#17#24#13#4": "g              g",
		"sign#17#24#13#5": "g   麥克斯·恩格g",
		"sign#17#24#13#6": "g    舍主     g",
		"sign#17#24#13#7": "deeeeeeeeeeeeeef",

		// sign#17#20#24#0 - 皇家馬廄
		"sign#17#20#24#0": "abbbbbbbbbbbbbbc",
		"sign#17#20#24#1": "g              g",
		"sign#17#20#24#2": "g   皇家馬廄   g",
		"sign#17#20#24#3": "g              g",
		"sign#17#20#24#4": "g   小心 汝之   g",
		"sign#17#20#24#5": "g     步 伐     g",
		"sign#17#20#24#6": "deeeeeeeeeeeeeef",

		// sign#17#16#19#0 - 不列顛王私人居室
		"sign#17#16#19#0": "abbbbbbbbbbbbbbc",
		"sign#17#16#19#1": "g              g",
		"sign#17#16#19#2": "g   不列顛王   g",
		"sign#17#16#19#3": "g   私人房間   g",
		"sign#17#16#19#4": "g              g",
		"sign#17#16#19#5": "deeeeeeeeeeeeeef",

		// sign#18#15#26#0 - 黑棘宮殿歡迎
		"sign#18#15#26#0": "abbbbbbbbbbbbbbc",
		"sign#18#15#26#1": "g              g",
		"sign#18#15#26#2": "g   歡迎蒞臨   g",
		"sign#18#15#26#3": "g  至尊主宰之   g",
		"sign#18#15#26#4": "g    宮 殿     g",
		"sign#18#15#26#5": "g   黑 棘     g",
		"sign#18#15#26#6": "deeeeeeeeeeeeeef",

		// sign#19#13#6#0 - 盜墓警告
		"sign#19#13#6#0": "8llllllmllllll9",
		"sign#19#13#6#1": "g             g",
		"sign#19#13#6#2": "g  盜墓者    g",
		"sign#19#13#6#3": "g  必受追訴  g",
		"sign#19#13#6#4": "g             g",
		"sign#19#13#6#5": ":llllllnllllll;",

		// sign#19#1#1#0 - 可憐的柯林
		"sign#19#1#1#0": "8llllllllllllll9",
		"sign#19#1#1#1": "g              g",
		"sign#19#1#1#2": "g  於此眠臥可憐  g",
		"sign#19#1#1#3": "g    柯 林     g",
		"sign#19#1#1#4": "g  跌下懸崖身亡  g",
		"sign#19#1#1#5": "g              g",
		"sign#19#1#1#6": "jllllllllllllllk",

		// sign#19#3#1#0 - 伊恩的故事
		"sign#19#3#1#0": "8llllllllllllll9",
		"sign#19#3#1#1": "g              g",
		"sign#19#3#1#2": "g  於此埋葬伊恩  g",
		"sign#19#3#1#3": "g    何等壞孩子    g",
		"sign#19#3#1#4": "g  他曾是怎樣壞  g",
		"sign#19#3#1#5": "g  孩子遺憾他不  g",
		"sign#19#3#1#6": "g              g",
		"sign#19#3#1#7": "jllllllllllllllk",

		// sign#19#5#1#0 - 柯克之死
		"sign#19#5#1#0": "     hlllli     ",
		"sign#19#5#1#1": "hllllk    jlllli",
		"sign#19#5#1#2": "g  柯克失利於   g",
		"sign#19#5#1#3": "g  一場戰鬥矣   g",
		"sign#19#5#1#4": "g 遭刀劍錯誤之側  g",
		"sign#19#5#1#5": "g    傷 害 矣    g",
		"sign#19#5#1#6": "jlllli    hllllk",
		"sign#19#5#1#7": "     g    g     ",

		// sign#19#7#1#0 - 提姆試圖飛行
		"sign#19#7#1#0": "8llllllllllllll9",
		"sign#19#7#1#1": "g              g",
		"sign#19#7#1#2": "g  欲試飛行者   g",
		"sign#19#7#1#3": "g   可憐小提姆   g",
		"sign#19#7#1#4": "g  致命之舉矣   g",
		"sign#19#7#1#5": "g   為己念想    g",
		"sign#19#7#1#6": "g              g",
		"sign#19#7#1#7": "jllllllllllllllk",

		// sign#19#9#1#0 - 勞瑞爾的故事
		"sign#19#9#1#0": "     hlllli     ",
		"sign#19#9#1#1": "hllllk    jlllli",
		"sign#19#9#1#2": "g   強 者 且 勇  g",
		"sign#19#9#1#3": "g   年輕勞瑞爾   g",
		"sign#19#9#1#4": "g  試木桶下瀑布  g",
		"sign#19#9#1#5": "g              g",
		"sign#19#9#1#6": "jlllli    hllllk",
		"sign#19#9#1#7": "     g    g     ",

		// sign#19#1#6#0 - 安德烈之死
		"sign#19#1#6#0": "     hlllli     ",
		"sign#19#1#6#1": "hllllk    jlllli",
		"sign#19#1#6#2": "g  安德烈眠於    g",
		"sign#19#1#6#3": "g  地球傘下     g",
		"sign#19#1#6#4": "g  哀哉彼死於   g",
		"sign#19#1#6#5": "g  沙門氏菌病矣  g",
		"sign#19#1#6#6": "jlllli    hllllk",
		"sign#19#1#6#7": "     g    g     ",

		// sign#19#3#6#0 - 戴爾在酒吧
		"sign#19#3#6#0": "8llllllllllllll9",
		"sign#19#3#6#1": "g              g",
		"sign#19#3#6#2": "g 末見於酒館內  g",
		"sign#19#3#6#3": "g   人名戴爾    g",
		"sign#19#3#6#4": "g  所食之物甚   g",
		"sign#19#3#6#5": "g   於腐敗矣    g",
		"sign#19#3#6#6": "g              g",
		"sign#19#3#6#7": "jllllllllllllllk",

		// sign#19#5#6#0 - 科特調情
		"sign#19#5#6#0": "8llllllllllllll9",
		"sign#19#5#6#1": "g              g",
		"sign#19#5#6#2": "g  愛慕女性者   g",
		"sign#19#5#6#3": "g    英俊科特    g",
		"sign#19#5#6#4": "g  與不當人調情  g",
		"sign#19#5#6#5": "g              g",
		"sign#19#5#6#6": "jllllllllllllllk",

		// sign#19#7#6#0 - 珍與工作
		"sign#19#7#6#0": "     hlllli     ",
		"sign#19#7#6#1": "hllllk   jlllli",
		"sign#19#7#6#2": "g  更加努力工作  g",
		"sign#19#7#6#3": "g   曰珍女士    g",
		"sign#19#7#6#4": "g  從未再見矣   g",
		"sign#19#7#6#5": "jlllli   hllllk",
		"sign#19#7#6#6": "     g   g     ",

		// sign#19#9#6#0 - 羅伯特小氣
		"sign#19#9#6#0": "     hlllli     ",
		"sign#19#9#6#1": "hllllk   jlllli",
		"sign#19#9#6#2": "g  過度節儉者   g",
		"sign#19#9#6#3": "g   羅伯特爵士  g",
		"sign#19#9#6#4": "g  僱員私刑彼以 g",
		"sign#19#9#6#5": "g   紓解己貧苦   g",
		"sign#19#9#6#6": "jlllli   hllllk",
		"sign#19#9#6#7": "     g   g    ",

		// sign#20#17#28#0 - 北不列塔尼歡迎
		"sign#20#17#28#0": "8lllllllmllllll9",
		"sign#20#17#28#1": "g              g",
		"sign#20#17#28#2": "g   歡迎來臨   g",
		"sign#20#17#28#3": "g  北不列塔尼   g",
		"sign#20#17#28#4": "g              g",
		"sign#20#17#28#5": "jlllllllnllllllk",

		// sign#22#19#17#0 - 波斯村歡迎
		"sign#22#19#17#0": "8lllllllmllllll9",
		"sign#22#19#17#1": "g              g",
		"sign#22#19#17#2": "g    歡迎至    g",
		"sign#22#19#17#3": "g    波斯村    g",
		"sign#22#19#17#4": "g              g",
		"sign#22#19#17#5": "jlllllllnllllllk",

		// sign#28#16#23#0 - 不尋常的嚙齒動物
		"sign#28#16#23#0": "abbbbbbbbbbbbbbc",
		"sign#28#16#23#1": "g              g",
		"sign#28#16#23#2": "g    謹慎!    g",
		"sign#28#16#23#3": "g              g",
		"sign#28#16#23#4": "g  不尋常之     g",
		"sign#28#16#23#5": "g    嚙齒獸    g",
		"sign#28#16#23#6": "deeeeeeeeeeeeeef",

		// sign#28#14#23#0 - 不尋常的嚙齒動物(重複)
		"sign#28#14#23#0": "abbbbbbbbbbbbbbc",
		"sign#28#14#23#1": "g              g",
		"sign#28#14#23#2": "g    謹慎!    g",
		"sign#28#14#23#3": "g              g",
		"sign#28#14#23#4": "g  不尋常之     g",
		"sign#28#14#23#5": "g    嚙齒獸    g",
		"sign#28#14#23#6": "deeeeeeeeeeeeeef",

		// sign#30#15#11#0 - 酒館與治療
		"sign#30#15#11#0": "abbbbbbbbbbbbbbc",
		"sign#30#15#11#1": "g              g",
		"sign#30#15#11#2": "g  酒館與治癒  g",
		"sign#30#15#11#3": "g  位於二樓    g",
		"sign#30#15#11#4": "deeeeeeeeeeeeeef",

		// sign#30#10#12#0 - 圖書館
		"sign#30#10#12#0": "abbbbbbbbbbbbbbc",
		"sign#30#10#12#1": "g              g",
		"sign#30#10#12#2": "g    圖書館    g",
		"sign#30#10#12#3": "g              g",
		"sign#30#10#12#4": "g    請安靜    g",
		"sign#30#10#12#5": "deeeeeeeeeeeeeef",

		// sign#32#25#24#0 - 訓練室
		"sign#32#25#24#0": "abbbbbbbbbbbbbbc",
		"sign#32#25#24#1": "g              g",
		"sign#32#25#24#2": "g    訓練室    g",
		"sign#32#25#24#3": "g              g",
		"sign#32#25#24#4": "g    房間     g",
		"sign#32#25#24#5": "deeeeeeeeeeeeeef",
	})
}
