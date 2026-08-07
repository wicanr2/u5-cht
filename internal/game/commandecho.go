package game

// 每個指令一開始會先印出自己的名字(原版 `sub_2ACF4` 各 case 的第一個 push)
//
// 這不是裝飾。原版的訊息欄是一條**逐字累加**的紙帶:按 G 先印 `Get-`,
// 接著方向鍵把 `North` 接在同一行後面,變成 `Get-North`。所以指令名末尾那個
// **`-` 本身就是「等方向」的提示** —— 少了它,玩家按下 G 之後畫面毫無反應,
// 而症狀看起來像「按鍵沒吃到」。
//
// 幾個指令的回顯**依所在位置而不同**,不是一張死表:
//
//	L  一律先印 `Look`,地牢裡再接 `...`,其餘接 `-`(問方向)
//	S  場景 / 地表印 `Search-`(問方向),地牢印 `Search...`(直接搜腳下)
//	E  地表交給處理函式,其餘印 `Enter what?`
//	A  分派器**什麼都不印** —— 由處理函式自己印 `Attack-` 或 `Attack`
//	空白 揚著帆印 `Sheets in irons!`(收帆),否則印 `Pass`
//
// ⚠ 非指令鍵(0x21..0x40 那一段)原版印 **`What?`**,不是靜靜吃掉。
//
// ⚠ **版面差異(不是機制差異)**:原版把指令名與方向接在同一行,本引擎的訊息欄
// 是一行一則的 CJK 版面。折衷是:名字以「——」結尾時,方向會被**接到同一則**
// 後面(`Append`),與原版讀起來一致;其餘情況各佔一行。

// 指令名末尾的兩種提示。`dirSuffix` 同時是「這個指令要問方向」的標記。
const (
	dirSuffix = " ——"
	ellipsis  = "……"
)

// commandEcho 是位置無關的那些指令名。
//
// 每一條後面附原版字串,方便對照;標點的形狀照原版的語意翻:
// 結尾 `-` → `——`(等方向)、`...` → `……`(接著還有東西)、`!` 保留。
var commandEcho = map[rune]string{
	'B': "登乘",             // "Board "
	'C': "施法" + ellipsis,  // "Cast..."
	'D': MsgDWhat,          // "D-What?"    —— 空鍵,沒有處理函式
	'F': "開砲" + dirSuffix, // "Fire-"
	'G': "拿取" + dirSuffix, // "Get-"
	'I': "點燃火把!",          // "Ignite torch!"
	'J': "撬鎖" + dirSuffix, // "Jimmy-"
	'K': "攀爬" + dirSuffix, // "Klimb-"
	'M': "調配藥草",           // "Mix Reagents"
	'N': "重排隊伍",           // "New Order"
	'O': "開啟" + dirSuffix, // "Open-"
	'P': "推動",             // "Push"
	'Q': "存檔:",            // "Quit:"
	'R': "裝備" + ellipsis,  // "Ready..."
	'T': "交談" + dirSuffix, // "Talk-"
	'U': "使用道具",           // "Use item"
	'V': "觀看寶石!",          // "View a gem!"
	'W': MsgWWhat,          // "W-What?"    —— 空鍵,沒有處理函式
	'X': "下載具",            // "X-it "
	'Y': "呼喊",             // "Yell "
	'Z': "角色數值" + ellipsis, // "Z-stats..."
	'H': "歇息" + dirSuffix, // "Hole up- "
}

// CommandEcho 回傳按下 key 時原版先印的那一句;空字串代表分派器不印。
//
// key 一律用大寫字母(或空白)。
func (s *State) CommandEcho(key rune) string {
	switch key {
	case ' ':
		// 空白鍵的兩句由 `Pass` 自己印(收帆 / 按兵不動),分派器這裡不重複。
		return ""
	case 'A':
		// 原版分派器對 A 沒有 push —— 三個位置各自的處理函式才印。
		return ""
	case 'E':
		if s.Location == 0 && !s.InCombat() {
			return "" // 地表:交給處理函式
		}
		return "進入什麼?" // "Enter what?"
	case 'L':
		if s.InDungeon() {
			return "觀察" + ellipsis // "Look" + "...\n"
		}
		return "觀察" + dirSuffix // "Look" + '-'
	case 'S':
		if s.InDungeon() {
			return "搜尋" + ellipsis // "Search...\n"
		}
		return "搜尋" + dirSuffix // "Search-"
	}
	return commandEcho[key]
}

// EchoCommand 印出指令名。回報有沒有印。
//
// 不認得的鍵印 `What?`(原版 `sub_2ACF4` 的 default);
// key 為 0 代表呼叫端自己判斷過了,什麼都不做。
func (s *State) EchoCommand(key rune) bool {
	if key == 0 {
		return false
	}
	if key >= 'a' && key <= 'z' {
		key = key - 'a' + 'A'
	}
	echo := s.CommandEcho(key)
	if echo == "" {
		return false
	}
	s.Log(echo)
	return true
}

// UnknownCommand 是按到非指令鍵時原版的回應(`What?`)。
func (s *State) UnknownCommand() { s.Log(MsgWhat) }

// Append 把文字接在最後一則訊息後面,而不是另起一則。
//
// 給「指令名 + 方向」這種原版在同一行的組合用。沒有訊息可接就當成 Log。
func (s *State) Append(text string) {
	if len(s.Messages) == 0 {
		s.Log(text)
		return
	}
	s.Messages[len(s.Messages)-1] += text
}

// awaitingAfterDash 回報最後一則訊息是不是「等方向」的指令名。
func (s *State) awaitingAfterDash() bool {
	if len(s.Messages) == 0 {
		return false
	}
	last := s.Messages[len(s.Messages)-1]
	return len(last) >= len(dirSuffix) && last[len(last)-len(dirSuffix):] == dirSuffix
}
