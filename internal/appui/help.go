package appui

// F1 指令說明(原版沒有這個畫面,是本重製版加的)
//
// U5 把 **A–Z 二十六個字母全部**當成指令,而畫面上只印一行提示。
// 1988 年那是靠紙本手冊解決的 —— 玩家手邊有《軟體世界》的說明書。
// 現在玩家手上沒有那本書,所以「開門是哪個鍵」這種問題必須在遊戲裡答得出來
//(使用者指示 2026-08-09)。
//
// ⚠ **這是說明文案,不是遊戲規則。** 放在 `appui` 而不是 `internal/game`:
// 那一包是原版行為的還原,混進「按鍵怎麼跟玩家解釋」會讓
// 「這條是不是原版的」變難查(`CLAUDE.md §3.0`)。
//
// ⚠ 括號裡的英文是**原版的指令名**,不要拿掉 —— 玩家看攻略、看手冊時
// 對得上的是英文;而且那也是這一列在原版真的存在的證據。

// HelpEntry 是說明畫面上的一列。
type HelpEntry struct {
	// Key 是要按的鍵(照原版的字母)。
	Key string
	// Title 是這個指令做什麼。
	Title string
	// Note 是補充:什麼時候用、要不要接方向。空字串就不印。
	Note string
}

// 原版的按鍵表(A–Z + 空白)。**兩個字母原版是空的**:`D` 與 `W`
// 按下去只會印 `D-What?` / `W-What?` —— 那是原版的樣子,照實列出來,
// 不要偷偷拿去掛新功能(借一個字母就蓋掉一個原版指令)。
var gameCommands = []HelpEntry{
	{"方向鍵", "移動", "場景裡走到地圖邊緣會問要不要離開"},
	{"空白", "原地等一回合(Pass)", "揚著帆時是收帆"},
	{"A", "攻擊(Attack)", "接一個方向"},
	{"B", "登乘(Board)", "船、小艇、馬、魔毯"},
	{"C", "施法(Cast)", "打符文首字母,例如 An Nox 是解毒"},
	{"E", "進入(Enter)", "站在城鎮 / 城堡 / 地牢入口上"},
	{"F", "開砲(Fire)", "只有在船上才有用,接一個方向"},
	{"G", "拿取(Get)", "接一個方向。拿別人桌上的東西會扣業報"},
	{"H", "歇息紮營(Hole up & camp)", "野外恢復,會消耗時間"},
	{"I", "點燃火把(Ignite torch)", "地牢裡沒火把是全黑的"},
	{"J", "撬鎖(Jimmy)", "接一個方向,要有鑰匙"},
	{"K", "攀爬(Klimb)", "梯子、繩索;地牢裡問上或下"},
	{"L", "觀察(Look)", "接一個方向;地牢裡直接看腳下"},
	{"M", "調配藥草(Mix Reagents)", "先選咒語,再勾要用的藥草"},
	{"N", "重排隊伍(New Order)", "換隊員的前後順序"},
	{"O", "★ 開門 / 開箱(Open)", "接一個方向。門要先開才走得過去"},
	{"P", "推動(Push)", "把桌椅之類推開"},
	{"Q", "存檔並離開(Quit & Save)", "原版的離開方式;本版另有 F5 / F10"},
	{"R", "裝備(Ready)", "武器、護甲、飾品"},
	{"S", "搜尋(Search)", "接一個方向;地牢裡搜腳下"},
	{"T", "交談(Talk)", "接一個方向。關鍵字打 name / job / join"},
	{"U", "使用道具(Use item)", "藥水、卷軸、月石、魔毯…"},
	{"V", "觀看寶石(View a gem)", "把整張地圖攤開一下"},
	{"X", "下載具(X-it)", "從船 / 馬 / 魔毯上下來"},
	{"Y", "呼喊(Yell)", "喊力量之言、暗影君主的名字;船上是收放帆"},
	{"Z", "角色數值(Z-stats)", "翻頁看隊員的能力、裝備、道具"},
	{"D / W", "(原版是空的)", "按下去只會印 D-What? / W-What?"},
}

// 本重製版加的鍵。**全部掛功能鍵**,一個字母都不借。
var appCommands = []HelpEntry{
	{"F1", "這張說明", "再按一次或按 ESC 收起"},
	{"F2", "切換版面", "現代版面(分組留白)與原版版面(照原版右欄)互換"},
	{"F3", "除錯欄位", "右欄多顯示座標與地形碼(原版沒有這兩欄)"},
	{"F5", "即時存檔", "原地存,不離開遊戲(原版沒有這個)"},
	{"F6", "讀回存檔", ""},
	{"F9", "切換音樂來源", "FM Towns FM 音源 與 Roland MT-32 互換"},
	{"F10", "離開遊戲", "會先問一次,離開前自動存檔"},
	{"ESC", "取消", "永遠只有取消的意思,不會結束遊戲"},
}

// HelpTitle / HelpSections 是說明畫面的內容。
const HelpTitle = "指令說明"

// HelpSections 回傳兩段(原版指令、本版加的鍵),各附標題。
func HelpSections() []struct {
	Heading string
	Entries []HelpEntry
} {
	return []struct {
		Heading string
		Entries []HelpEntry
	}{
		{"原版指令", gameCommands},
		{"本版新增", appCommands},
	}
}

// HelpPanel 是說明畫面的開關。零值 = 沒開。
type HelpPanel struct {
	open bool
	page int
}

// IsOpen 回報說明畫面開著沒有。
func (h *HelpPanel) IsOpen() bool { return h != nil && h.open }

// Page 回報現在在第幾頁(從 0 起)。
func (h *HelpPanel) Page() int {
	if h == nil {
		return 0
	}
	return h.page
}

// Close 收起說明。
func (h *HelpPanel) Close() {
	if h != nil {
		h.open, h.page = false, 0
	}
}

// Step 吃這一帧的按鍵,回報畫面有沒有變。
//
// 三條規則:
//
//  1. **F1 是開關** —— 開著再按就收起(玩家按 F1 進來,第一個反射就是再按 F1 出去)。
//  2. **ESC 收起**,與 `QuitDialog` 同一條鐵則:ESC 只取消。
//  3. 開著時 **PgDn / PgUp 或方向鍵上下翻頁**;其餘按鍵一律吃掉 ——
//     它是覆蓋層,不該讓遊戲在背後動起來。
func (h *HelpPanel) Step(k Keys, pages int) (changed bool) {
	if h == nil {
		return false
	}
	if !h.open {
		if k.Help {
			h.open, h.page = true, 0
			return true
		}
		return false
	}
	switch {
	case k.Help, k.Escape:
		h.Close()
		return true
	case k.PageDown && h.page+1 < pages:
		h.page++
		return true
	case k.PageUp && h.page > 0:
		h.page--
		return true
	}
	return false
}
