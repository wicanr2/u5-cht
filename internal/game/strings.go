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
	// "He does not understand thee." —— 打了 NPC 不認得的關鍵字。
	MsgDoesNotUnderstand = "對方聽不懂汝所言。"
	// 尚未實作的副作用:誠實說明,不要假裝發生了(CLAUDE.md §3.0)。
	MsgJoinNotImplemented   = "(對方欲加入隊伍 —— 隊伍系統尚未實作)"
	MsgGuardsNotImplemented = "(對方喚來衛兵 —— 衛兵反應尚未實作)"
	MsgAskNotImplemented    = "(對方反問汝 —— 問答流程尚未實作)"
	// "\nWhat town?\n" —— 原版在按下「進入」卻不站在地點上時的回應。
	// 直譯「什麼城鎮?」在中文裡像在反問玩家,改成敘述句。
	MsgNothingToEnter = "此處無可進入之地。"
)
