package u5data

// 衛兵的盤查(原版 `sub_1B3D0`,由對話號碼 0xFF 觸發)
//
// 「什麼事會惹毛衛兵」的答案不是某個行為判定,而是**對話號碼**:
// 排程表裡對話號碼是 `0xFF` 的那些 NPC 就是會攔人的衛兵。走過去跟他說話
// (或被他攔下)就進盤查,**答不出來或不肯給就當場逮捕**。
//
// 三種盤查依地點分派,順序照原版的三段 `cmp byte_3E0A3`:
//
//	18  黑棘的宮殿  要戴著徽章,而且答得出密語
//	 5  米諾克      交出一半的錢
//	其餘            每個活著的隊員 10 gp,說是「給黑棘的貢金」
//
// ⚠ 三段是 if / else if / else,不是可疊加的條件 —— 黑棘宮殿裡不會再要貢金。

// DialogueGuardChallenge 是「這個 NPC 會攔人盤查」的對話號碼。
//
// 與另外兩個特別號碼同一組:
//
//	0xFD  「別傷害我!」—— 怕你的人
//	0xFE  「滾開,害蟲!」—— 討厭你的人(說完會逃走)
//	0xFF  盤查                ← 這個
const DialogueGuardChallenge = DialogueSpecialFF

// GuardTributePerMember 是一般城鎮的貢金,**每人** 10 gp(原版 `add esi, 0Ah`)。
const GuardTributePerMember = 10

// GuardHalfGoldLocation 是「交出一半的錢」的地點編號(`cmp al, 5`)。
//
// = 5 = 米諾克(MINOC)。原版的措辭是 `Thou wilt give half thy gold to charity!`
// —— 名義上是「捐給慈善」,實際上跟別處的貢金一樣是勒索。
const GuardHalfGoldLocation = 5

// GuardTribute 是這一隊要繳多少貢金。
//
// ⚠ 算的是**活著的**成員(原版 `cmp byte_3DDBF[eax], 44h` —— 'D' 是死亡)。
// 隊裡有屍體不必為他繳錢。
func GuardTribute(livingMembers int) int {
	if livingMembers < 0 {
		return 0
	}
	return livingMembers * GuardTributePerMember
}

// BlackthornPassword 是黑棘宮殿衛兵要的密語。
//
// **[HARD] 維持英文** —— 這是玩家要打進去的字,譯成中文就打不出來了。
// 提示可以譯,答案不行(同 `wordofpower.go` 的力量之言)。
//
// ★ 值是實際比對用的那個字串,不是從遊戲印象打的:
// FM Towns `WORRIORS.EXP` 的 `aImpe` 是 `'IMPE',0`,而 DOS 版
// `DATA.OVL` 位移 0x4AAA 也是獨立一個 NUL 結尾的 `IMPE` —— 兩版一致,
// 沒有被截短。
const BlackthornPassword = "IMPE"

// BlackthornPasswordMax 是密語輸入的長度上限(`sub_2B770(buf, 0x0E)`)。
const BlackthornPasswordMax = 14

// PasswordMatches 比對密語(原版 `sub_1B140`)。
//
// ⚠ 是**完整相等**,不是詞首比對 —— 迴圈要兩邊同時碰到結尾才回 1。
// 這與對話關鍵字的 `MatchPrefix`(截到 9 字的詞首比對)是**兩套不同的規則**,
// 混用的話「IMPERA」會被誤判成通過。
//
// 比對前兩邊都先清掉最高位元(`and al, 7Fh`)再轉大寫。
func PasswordMatches(password, typed string) bool {
	norm := func(s string) string {
		out := make([]byte, len(s))
		for i := 0; i < len(s); i++ {
			c := s[i] & 0x7F
			if c >= 'a' && c <= 'z' {
				c -= 32
			}
			out[i] = c
		}
		return string(out)
	}
	return norm(password) == norm(typed)
}

// BadgeMode 是「配戴著黑棘徽章」的狀態值(原版 `byte_3E08A = 0x1D`)。
//
// ⚠⚠ **它與四個戰鬥模式咒語共用同一個位元組。** `byte_3E08A` 平常放的是
// In Sanct 'P'、Rel Tym 'Q'、Quas An Wis 'C'、In An 'N'、An Tym 'T';
// 戴上徽章時放 0x1D。也就是說 **戴著徽章時施那些咒語會把徽章「脫掉」**,
// 反過來也一樣 —— 而遊戲不會提示。
//
// 這是原版的共用儲存,照抄(同 `shadowlord.go` 的末日位元與地點 29 共用)。
const BadgeMode = 0x1D
