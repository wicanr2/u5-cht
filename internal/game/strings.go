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
	MsgUnderworld = "地下世界!"
	// "No\n"
	MsgNo = "否。"
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
	// "Don't hurt me!\nPlease go away!"
	MsgFrightened = "「別傷害我!請走開!」"
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
	// 尚未實作的副作用:誠實說明,不要假裝發生了(CLAUDE.md §3.0)。
	MsgGuardsNotImplemented = "(對方喚來衛兵 —— 衛兵反應尚未實作)"
	// "\nNot here!\n" —— 在不能做這件事的地方按了 B / X。
	MsgNotHere = "此處不可。"
	// "What?\n"
	MsgWhat = "此為何意?"
	// "\nOn foot\n" —— 要先下來走路才能上另一個載具。
	MsgOnFoot = "汝須先下來步行。"
	// "\nDANGER: SHIP BADLY DAMAGED!\n"
	MsgShipDamaged = "警告:此船受損嚴重!"
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
)
