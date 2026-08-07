package i18n

import (
	"fmt"
	"strings"
)

// `.TLK` 對話本文的譯文覆蓋層
//
// 四個對話檔(`TOWNE` / `DWELLING` / `CASTLE` / `KEEP`)共數百筆記錄,
// 每筆又分名字、外觀、招呼、職業、道別、若干關鍵字回應與提問區塊。
//
// # key 的形狀
//
//	檔名#對話id#欄位        例:CASTLE.TLK#8#greet
//
// **id 是 `.TLK` 索引表裡的號碼,不是第幾筆記錄。** 原版 `sub_1C840` 也是
// 線性搜尋 id 而不是拿它當下標(見 `docs/re/06`),照著它走才不會在
// 記錄順序變動時整批錯位。
//
// # 為什麼按位置而不是按原文查
//
// 對話文字在顯示前會先跑 `Conversation.Render` 展開 opcode,其中 0x81 會**插入
// 玩家角色的名字** —— 同一筆記錄對不同存檔會渲染出不同的英文。拿渲染後的
// 英文當 key,換個角色名就查不到了。所以 key 只用位置,與內容無關。
//
// # 譯文裡的 %A
//
// 譯文要放玩家名字的地方寫 `%A`(對應原版 opcode 0x81)。查表時代入,
// 與英文那條路的行為一致。
//
// # 譯法
//
// 對白用中世紀羊皮紙腔(汝 / 卿 / 吾),對齊《軟體世界》手冊的語氣;
// 專有名詞一律走 `Name()`,不要在這裡另外翻一次(會與 u6-cht 漂移)。
//
// ⚠ **[HARD] 玩家要打得出來的字串維持英文**:關鍵字本身(`name` / `job` /
// `join` / `bye`…)是比對用的 canonical 值,**不在這一層**——
// 這一層只翻「NPC 說的話」。

// 對話欄位的名字。組 key 用,別直接寫字串字面量。
const (
	TalkFieldName  = "name"
	TalkFieldDesc  = "desc"
	TalkFieldGreet = "greet"
	TalkFieldJob   = "job"
	TalkFieldBye   = "bye"
)

// TalkEntryField 是第 i 個關鍵字回應的欄位名。
func TalkEntryField(i int) string { return fmt.Sprintf("e%d", i) }

// TalkQuestionField 是第 i 個提問區塊的欄位名(問句 / 否 / 是)。
func TalkQuestionField(i int) string  { return fmt.Sprintf("q%d", i) }
func TalkQuestionNoField(i int) string  { return fmt.Sprintf("q%dn", i) }
func TalkQuestionYesField(i int) string { return fmt.Sprintf("q%dy", i) }

// AvatarToken 是譯文裡代表玩家名字的記號(原版 opcode 0x81)。
const AvatarToken = "%A"

// TalkKey 組出對話譯文的 key。
func TalkKey(file string, id int, field string) string {
	return fmt.Sprintf("%s#%d#%s", file, id, field)
}

// Talk 查一段對話的譯文。查不到就回原文 —— 半套中文比整段亂碼好懂。
//
// avatar 用來代入譯文裡的 `%A`;傳空字串就原樣留著記號(dump 工具會這樣用)。
func Talk(file string, id int, field, en, avatar string) string {
	zh, ok := talks[TalkKey(file, id, field)]
	if !ok {
		return en
	}
	if avatar != "" {
		zh = strings.ReplaceAll(zh, AvatarToken, avatar)
	}
	return zh
}

// TalkTranslated 回報這一段翻過了沒。統計覆蓋率用。
func TalkTranslated(file string, id int, field string) bool {
	_, ok := talks[TalkKey(file, id, field)]
	return ok
}

// TalkCount 是已翻的段數。
func TalkCount() int { return len(talks) }

// talks 是譯文表。空的時候整套走英文原文,不會壞 ——
// 這一層的存在本身就是 P5 的骨架,譯文逐批補上。
var talks = map[string]string{}

// addTalk 併入一批譯文。重複的 key 直接覆蓋前一批,方便分檔維護。
func addTalk(m map[string]string) {
	for k, v := range m {
		talks[k] = v
	}
}
