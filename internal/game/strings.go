package game

// 引擎內建訊息的中譯
//
// 這些字串**不在任何資料檔裡**,是硬寫在原版執行檔的 .text 段中(所以翻譯不能靠改
// .DAT,得在這裡定案)。每一條都附原文,方便日後對照與校訂 —— 譯名與語氣的依據
// 見 docs/localization-notes.md。
//
// 語氣取向:原版用的是帶古風的英文(thou / dost),中譯跟著走文言一點的路子,
// 但不到艱澀 —— 這與姊妹專案 u4-cht / u6-cht 的處理一致。
const (
	// "Blocked!\n"
	MsgBlocked = "去路受阻!"
	// "\nDost thou wish to leave? "
	MsgLeaveQuestion = "汝欲離去否?(Y/N)"
	// "Yes\n\nExit to\n"
	MsgExitTo = "是 —— 前往"
	// "Britannia!\n"
	MsgBritannia = "不列顛尼亞!"
	// "Underworld!\n"
	MsgUnderworld = "幽冥界!"
	// "No\n" / "Yes\n"
	MsgNo  = "否。"
	MsgYes = "是。"
	// "Up!\n"
	MsgUp = "拾級而上!"
	// "Down!\n"
	MsgDown = "拾級而下!"
	// "What?\n" —— 原版在按 K 卻不站在梯子上時的回應。
	MsgNothingToClimb = "此處無梯可攀。"
	// "\nNobody's here!\n"
	MsgNobodyHere = "此處無人。"
	// "No response!\n" / "The guard offers\nno response!\n"
	MsgNoResponse = "無人回應。"
	// MsgNoEffect 是原版的「No effect!」—— 力量之言說錯地方、或說了不是那八個字。
	MsgNoEffect = "毫無效果!"
	// Yell 指令的四句(原版 `sub_17E74`)。
	//
	// ⚠ FURL / HOIST 的方向別搞反:載具碼 0x20..0x23 是**揚著帆**,
	// 按下去是「收帆」(+4 變 0x24..0x27);反過來才是「揚帆」。
	MsgFurl        = "收帆!"
	MsgHoist       = "揚帆!"
	MsgYellWhat    = "喊什麼?"
	MsgYellNothing = "什麼也沒喊。"
	// "\nA word of power is uttered\n"
	MsgWordUttered = "一個力量之言被說了出來。"
	// "\nA shadowlord appears\n"
	MsgShadowlordAppears = "一位暗影君主現身了!"
	// "\n\nThe Shrine is\nrestored!\n"
	MsgShrineRestored = "聖壇復原了!"
	// 黑棘的審問(原版 `sub_C414` 裡寫死在執行檔的三句)。
	//
	// "\nThou art subdued and blindfolded!"
	MsgSubdued = "汝被制伏,雙眼遭蒙!"
	// "\n\nStrong guards drag thee away!"
	MsgDraggedAway = "壯碩的衛兵把汝拖走了!"
	// "\n\nYour response?\n:"
	MsgYourResponse = "汝如何回答?"
	// 逮捕(原版 `sub_1884`)。
	//
	// "\n\"Thou art under arrest!\"\n\n" + "\"Wilt thou come quietly?\"\n\n:"
	MsgUnderArrest = "「汝被逮捕了!」「汝束手就擒否?」(Y/N)"
	// "No\n\n\"Then defend thyself, rogue!\"\n"
	MsgDefendThyself = "否 ——「那就自衛吧,惡棍!」"
	// "Yes\n\nThe guard strikes thee unconscious!"
	MsgStruckUnconscious = "是 —— 衛兵一擊把汝打昏了!"
	// "\nThou dost awaken to...\n"
	MsgAwakenTo = "汝悠悠轉醒……"
	// "\nAttacked!\n"
	MsgAttacked = "遭到攻擊!"
	// "Don't hurt me!\nPlease go away!"
	MsgFrightened = "「別傷害我!請走開!」"

	// 衛兵的盤查(原版 `sub_1B3D0`)。**密語本身維持英文**,提示可以譯。
	//
	//	aGiveNowThePass  "Give now the password, bearer of the Badge!"
	//	aThouWiltGiveHa  "Thou wilt give half thy gold to charity!"
	//	aAGuardDemandsA  "A guard demands a %d gp tribute to Blackthorn!"
	//	aDostThouPay     "Dost thou pay?"
	//	aPassFriend      "Pass, friend!"
	//	aBegoneVermin    "Begone, vermin!"
	MsgGuardPassword = "「拿出通行密語來,佩徽章者!」\n汝之回答?"
	MsgGuardHalfGold = "「汝須捐出半數家財行善!」汝可願付?(Y/N)"
	MsgGuardTribute  = "衛兵索取 %d gp,\n說是給黑棘的貢金!汝可願付?(Y/N)"
	MsgGuardYes      = "願"
	MsgGuardNo       = "不願"
	MsgGuardPass     = "「過去吧,朋友!」"
	MsgBegoneVermin  = "「滾開,害蟲!」"

	// 那個人在床上睡著(原版地形 0xAB → 「Zzzzzz」)。
	MsgAsleep = "呼……呼……"

	// 豎琴(原版 `sub_11E0`)。原版只發聲不印字。
	MsgSecretDoor = "牆上有東西動了……"

	// NPC 反問名字(原版 opcode 0x88 / `sub_1C2FC`)。
	MsgWhatIsThyName = "「汝名為何?」"
	MsgAPleasure     = "幸會!"
	MsgIfYouSaySo    = "汝說是就是吧。"

	// 施法的輸入與地點壓制(原版 `sub_1994C` / `sub_1CA0C`)。
	//
	// ⚠ 三句都不能合併:`MsgSpellNone` 是「一個符文都沒打」、
	// `MsgSpellNoEffect` 是「打了但湊不出咒語」,原版分得很清楚。
	MsgSpellName     = "咒語:"
	MsgForWhatSpell  = "為哪個咒語?"
	MsgReagents      = "藥草:"
	MsgNothingToMix  = "沒東西可調!"
	MsgHowMuch       = "要幾份?"
	MsgSpellNone     = "無!"
	MsgSpellNoEffect = "毫無效果!"
	MsgMagicAbsorbed = "被吸收了!"

	// 轉入 Ultima IV(原版 `sub_7594` 的畫面)。
	//
	// ⚠ 「Keep same sex?」的 N 是**翻轉**,不是「設成女」—— 見 `game/transfer.go`。
	MsgTransferError  = "錯誤:汝之《創世紀 IV》存檔內容有誤。"
	MsgTransferUnable = "無法完成轉入。"
	MsgTransferFound  = "尋得"
	MsgIsAnAvatar     = "乃聖者。"
	MsgIsNotAnAvatar  = "並非聖者。"
	MsgKeepThisName   = "保留此名否?(Y/N)"
	MsgEnterNewName   = "請輸入新名:"
	MsgKeepSameSex    = "保留原性別否?(Y/N)"
	MsgMale           = "男"
	MsgFemale         = "女"

	// Get 指令(原版 `sub_15A94` / `sub_154BC`)。
	MsgNothingToGet   = "這裡沒有東西可拿。"
	MsgOpenItFirst    = "得先打開它!"
	MsgCannotCarry    = "汝拿不動那個。"
	MsgBorrowed       = "借用!"
	MsgCropsPicked    = "作物採收了!"
	MsgMmmmm          = "嗯 —— 好吃!"
	MsgCantReachPlate = "碰不到那個盤子!"
	MsgGotGold        = "得到 %d 枚金幣。"
	MsgGotGems        = "得到 %d 顆寶石。"
	MsgGotKeys        = "得到 %d 把鑰匙。"
	MsgGotOddKeys     = "得到 %d 把怪鑰匙。"
	MsgGotPotion      = "一瓶%s色藥水!"
	MsgGotScroll      = "一捲卷軸 —— %s!"
	MsgGotItem        = "得到%s。"
	// "… is absorbed!"(原版 sub_161E4)—— 城堡把站在門前的人吸了進去。
	MsgIsAbsorbed = "被吸了進去!"
	MsgGotTorches     = "得到 %d 支火把。"
	MsgGotFood        = "得到 %d 份糧食。"
	MsgGotCarpet      = "一張魔毯!"
	MsgGotSandalwood  = "一只檀香木盒!"
	MsgGotPlans       = "那份圖紙!"
	MsgGotCrown       = "不列顛王的王冠!"
	MsgGotSceptre     = "不列顛王的權杖!"
	MsgGotAmulet         = "不列顛王的護符!"
	MsgGotShard       = "%s之碎片!"
	MsgGotMoonstone   = "一顆月石!"
	MsgChestEmpty     = "箱子是空的!"
	MsgChestContents  = "箱中之物:"
	// "A merchant says:\n\"Come see me at\nmy … when it's open!\""
	MsgMerchantClosed = "「營業時再來吧。」"
	// "A merchant says:\n\"GET THAT HORSE OUT …" —— 騎馬進店(馬廄除外)。
	MsgGetThatHorseOut = "把那匹馬牽出去!"
	// 交易流程尚未實作:誠實說明(CLAUDE.md §3.0)。
	MsgShopNotImplemented = "(買賣流程尚未實作 —— 缺物品表與價格表)"
	// "\"I cannot help thee with that." —— 打了 NPC 不認得的關鍵字。
	MsgDoesNotUnderstand = "「此事我幫不上汝。」"
	// "\"My name is " —— 問 NAME 時的固定開頭。
	MsgMyNameIs = "吾名"
	// "\"With language like that, how did you become an Avatar?"
	// 對 NPC 罵髒話的固定回應(原版有 29 個字都導到這一句)。
	MsgFoulLanguage = "出言如此,汝是如何成為聖者的?"
	// "\"Thou hast no room for me in thy party…\"" / "Seek me again if one of thy members doth…"
	MsgPartyFull = "「汝之隊伍已滿,容不下我。待有人離去,再來尋我。」"
	// "\nSystem Error -\nNo Match!" —— 對話記錄與名冊對不上。
	// 原版會印出來,這裡照樣讓它看得見:那是資料或解析的問題,不該靜默吞掉。
	MsgNoMatch = "(名冊中查無此人 —— 對話記錄與名冊對不上)"
	// 入隊成功的提示。原版沒有專屬訊息(直接把人放進隊伍),這是引擎自己加的回饋。
	MsgJoined = "加入了汝的隊伍。"
	// 對話 opcode 0x8B:對方喚來衛兵(原版 sub_C10)。
	MsgGuardsCalled = "「衛兵!」"
	// "\nNot here!\n" —— 在不能做這件事的地方按了 B / X。
	MsgNotHere = "此處不可。"
	// "What?\n"
	MsgWhat = "此為何意?"

	// MsgDWhat / MsgWWhat 是 D 與 W 兩個鍵的回應。
	//
	// ★ 這兩個鍵在原版**不是指令** —— 主分派器 `sub_2ACF4` 與戰鬥分派器
	// `sub_A360` 都只印一句 `D-What?` / `W-What?`(兩處獨立佐證,`docs/re/49`)。
	// 把它們做成指令(我第一版把 W 當成 U4 的 Wear)就是自創遊戲。
	MsgDWhat = "D—— 何事?"
	MsgWWhat = "W—— 何事?"
	// "\nOn foot\n" —— 要先下來走路才能上另一個載具。
	MsgOnFoot = "汝須先下來步行。"
	// "\nDANGER: SHIP BADLY DAMAGED!\n"
	MsgShipDamaged = "警告:此船受損嚴重!"
	// 船身損傷與沉船(原版 `sub_22F0` / `sub_2CCFC` / `sub_2D9D0`,見 `docs/re/66`)
	//
	// ⚠ "Hull weak!" 的門檻是 **50**,與上面 MsgShipDamaged 的 10 是
	// **兩個不同的警告**:前者每次轉向都會唸,後者只在上船時唸一次。
	MsgHullWeak    = "船身脆弱!"
	MsgShipSunk    = "船沉了!"
	MsgAbandonShip = "棄船!"
	MsgDrowning    = "溺水!"
	MsgRoughSeas   = "風浪險惡!"
	// "Head " —— 大船轉向時印的,後面接方向名。
	MsgHeading = "轉向"

	// 戰鬥中的三句(原版 `sub_A360` / `sub_BCC4` / `sub_1F840`,見 `docs/re/67`)
	//
	// ⚠ `Zzzzz...` 是**隊員**睡著時每回合都會印的 —— 少了它,玩家只會看到
	// 自己的角色莫名其妙不動。怪物睡著走另一條路(`sub_A108`),不印這句。
	MsgZzzzz = "呼……呼……"
	// "ARGH!" —— 被拖屍怪拖到水下的人每回合的哀號。
	MsgArgh = "啊啊啊!"
	// " dragged under!" / " regurgitated!"
	MsgDraggedUnder = "被拖入水中!"
	MsgRegurgitated = "被吐了出來!"
	// " hit!" —— 隊員被打中只有這一句,**不報數字**(原版 `sub_1F840`)。
	MsgWasHit = "被擊中了!"
	// 怪物的四個傷勢等級(`sub_BAFC` 回 1..4,1 最重)。
	//
	// ⚠ 原版**從不告訴玩家怪物掉了幾點血** —— 只有這四句形容詞。
	// 改成報數字等於洩漏原版刻意藏起來的資訊。
	MsgWoundCritical = "已然垂危!"
	MsgWoundHeavily  = "傷勢沉重!"
	MsgWoundLightly  = "受了輕傷!"
	MsgWoundBarely   = "只受了皮肉傷!"
	// "\nWARNING: NO SKIFFS ON BOARD!\n" / "\nNo skiffs on board!\n"
	MsgNoSkiffs = "船上沒有小艇。"
	// "\nNo land nearby!\n"
	MsgNoLandNearby = "附近無陸地可登。"
	// "Under sail" —— 揚帆中不能下船。
	MsgUnderSail = "船正揚帆而行。"
	// 腳下放不下東西 —— 原版靠「取空槽」失敗,沒有專屬訊息;這是引擎的回饋。
	MsgNoRoom = "(此處放不下)"
	// "You respond-\n:" —— 回答 NPC 的提問時的提示。
	MsgYouRespond = "汝答:"
	// 記錄裡找不到對應的提問區塊 —— 資料或解析的問題,讓它看得見。
	MsgNoQuestionBlock = "(對方欲反問,但記錄裡找不到對應的提問 —— 解析可能有誤)"
	// "\nWhat town?\n" —— 原版在按下「進入」卻不站在地點上時的回應。
	// 直譯「什麼城鎮?」在中文裡像在反問玩家,改成敘述句。
	MsgNothingToEnter = "此處無可進入之地。"

	// Look 指令(原版 `sub_D9C4` / `sub_D258` / `sub_CE78`)
	//
	// "\nThou dost see\n"
	MsgThouDostSee = "汝所見為"
	// "a deep well"
	//
	// ⚠ **更正**:此前這裡寫「原版接著問『Drop a coin?』但不管答什麼都沒有後續,
	// 組語裡沒有任何寫入」—— 那是錯的。錯因是讀了 Hex-Rays 的 `sub_CD28`,
	// 而它**把函式截斷在 Y/N 迴圈**就 return 了(`CLAUDE.md` §4.4 那條 `[HARD]`
	// 講的正是這件事)。組語裡它有三個參數、四倍長,而且是**整個許願井彩蛋**:
	// 扣一枚錢、問願望、六個字串、兩個地點、生一匹馬。見 `u5data/well.go`。
	MsgLookWell = "一口深井。"
	// "Drop a coin?" / "\nThy wish?\n" / "Nothing\n" / "\nPoof!\n"
	MsgDropACoin = "要投一枚錢幣嗎?"
	MsgThyWish   = "汝所願為何?"
	MsgNothing   = "無。"
	MsgPoof      = "砰!"
	// "a gurgling fountain!" / "Who will drink?" / "None!" /
	// "Incapacitated!" / "Refreshing..."
	//
	// ⚠ 噴泉在原版**沒有療效**,只有這幾句話。不要補。
	MsgFountain      = "一座汩汩作響的噴泉!"
	MsgWhoWillDrink  = "何人飲之?"
	MsgNobodyDrinks  = "無人飲用。"
	MsgIncapacitated = "此人動彈不得。"
	MsgRefreshing    = "沁涼舒暢……"
	// "Strange vision!" / "Death vision!" —— 凝視水晶球的兩種結果。
	MsgStrangeVision = "奇異的幻象!"
	MsgDeathVision   = "死亡的幻象!"
	// " AM.\n" / " PM.\n" —— 老爺鐘的時刻後綴。
	MsgClockAM = " 上午。"
	MsgClockPM = " 下午。"
	// 招牌查不到時原版印的那一塊(`sub_D544(-1)`):寫死的「LIVE BY THE
	// EIGHT LAWS」告示板。它是**預設值不是錯誤**,所以照樣印出來。
	MsgSignDefault = "「謹遵八律而行」"
	// "the sun!" / "the night sky! " —— 抬頭看天的兩種結果。
	MsgTheSun   = "烈日當空!"
	MsgNightSky = "滿天星斗!"

	// "Slow progress!" / "Very slow!" —— 粗糙地形(原版 `sub_2D0BC`)。
	MsgSlowProgress = "步履維艱!"
	MsgVerySlow     = "寸步難行!"

	// Push 指令(原版 `sub_18154`)。
	// "Pushed!" / "Pulled!" / "Won't budge!"
	MsgPushed    = "推開了!"
	MsgPulled    = "拉過來了!"
	MsgWontBudge = "紋風不動!"

	// Jimmy(J)、New order(N)、View a gem(V)。
	// "No Keys!" / "Unlocked!" / "Key broke!" / "What?"
	MsgNoKeys      = "沒有鑰匙!"
	MsgUnlocked    = "開了!"
	MsgKeyBroke    = "鑰匙斷了!"
	MsgNoLockHere  = "此處無鎖可撬。"
	// "\n\nSwap " / " must lead!" / "nobody!"
	MsgSwapNobody = "沒有那個人。"
	MsgMustLead   = "必須走在最前面!"
	MsgSwapped    = "%s 與 %s 換了位置。"
	// "You have none!" —— 沒有寶石可看。
	MsgYouHaveNone = "汝一顆也沒有!"

	// 選單(R / W / N / U / M 共用)。
	MsgPickWho   = "由誰?"
	MsgPickItem  = "哪一件?"
	MsgPickSwap  = "換誰?"
	MsgPickWith  = "與誰交換?"
	MsgNevermind = "作罷。"
	// 特殊道具的中文名(選單上顯示;英文原文見 `docs/re/44` 的表)。
	MsgItemCarpet    = "魔毯"
	MsgItemSkullKey  = "骷髏鑰匙"
	MsgItemAmulet    = "不列顛王的護符"
	MsgItemCrown     = "不列顛王的王冠"
	MsgItemSceptre   = "不列顛王的權杖"
	MsgItemShard     = "寶石碎片"
	MsgItemSpyglass  = "望遠鏡"
	MsgItemPlans     = "海角號圖紙"
	MsgItemSextant   = "六分儀"
	MsgItemWatch     = "懷錶"
	MsgItemBadge     = "黑徽章"
	MsgItemWoodenBox = "檀香木盒"

	// Fire(F)與 Mix(M)。
	MsgBroadsidesOnly       = "只能打舷側!"
	MsgBooom                = "轟隆!"
	MsgNoReagents           = "沒有藥草!"
	MsgNoSuchSpell          = "沒有那個咒語。"
	MsgInsufficientReagents = "藥草不足!"
	MsgMixing               = "調配中……"
	MsgMixDone              = "完成!"
	MsgMixFailed            = "調配失敗 —— 藥草浪費了。"

	// Use(U)—— 原版 `sub_1A5E8`。
	MsgNoUsableItems    = "沒有可用的道具。"
	MsgUseAmulet        = "戴上不列顛王的護符……"
	MsgUseCrown         = "汝戴上了不列顛王的王冠……"
	MsgUseSceptre       = "舉起不列顛王的權杖……"
	MsgRemoved          = "取下了。"
	MsgBoarded          = "上去了!"
	MsgXitShipFirst     = "得先下船!"
	MsgOnlyOnFoot       = "只有步行時才行。"
	MsgSkullKey         = "骷髏鑰匙 ——"
	MsgFieldDissolved   = "力場消散了!"
	MsgLooking          = "凝視著……"
	MsgNoStars          = "看不見星星。"
	MsgShipRigged       = "船已改裝,航速加倍!"
	MsgOnlyOnShipboard  = "只有在船上才用得著。"
	MsgOnlyOutdoors     = "只有在戶外才行。"
	MsgOnlyAtNight      = "只有夜裡才行。"
	MsgPosition         = "位置:"
	MsgPocketWatch      = "懷錶顯示:"
	MsgBadgeWorn        = "徽章戴上了!"
	MsgShardOnlyAtFlame = "碎片只能投入聖火。"
	MsgBoxHow           = "(木盒的用法尚未實作)"

	// Search(S)—— 原版的地點語以 "t" 結尾,接上 "hou dost find" 成句。
	MsgThouDostFind = "汝翻到了 ——"
	MsgHiddenDoor   = "一道密門!"
	MsgNoTrap       = "沒有陷阱!"
	MsgATrap        = "有陷阱!"
	MsgSimpleTrap   = "一個簡單的陷阱!"
	MsgComplexTrap  = "一個複雜的陷阱!"
	MsgPlague       = "瘟疫!"
	MsgFoundGold    = "金幣!"
	MsgFoundFood    = "糧食!"
	MsgFoundNothing = "什麼也沒有。"
	MsgFoundWorms   = "一堆蛆。"
	MsgFoundGuts    = "一團內臟。"
	MsgFoundPulp    = "血肉模糊的一團。"

	// Ready(R)與 Wear(W)。
	MsgCannotReady  = "那個拿不上手。"
	MsgDontHaveThat = "汝沒有那件東西。"
	MsgEquipped     = "%s 裝上了%s。"
	MsgUnequipped   = "%s 卸下了%s。"

	// 主選單裡還沒實作的兩項 —— 照實說,不要做一個假裝有用的分支。
	MsgTransferNeedsPath = "請以 -u4save 指定《創世紀 IV》的 PARTY.SAV。"
	MsgAcknowledgementsNotImplemented = "(製作群畫面尚未實作 —— 原版是一張圖,素材未對到)"

	// 建立新角色(原版主選單的 `Create New Character`)。
	MsgCreateName   = "汝之名為何?"
	MsgCreateGender = "汝為男(M)或女(F)?"
	MsgCreateDone   = "汝之命途已定。踏上旅途吧!"
)
