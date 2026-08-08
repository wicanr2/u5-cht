package u5data

// 船的損傷與沉船(原版 `sub_22F0`)+ 轉向要花一回合(`sub_2CCFC`)
//
// 這兩支都是**因為 Hex-Rays 安靜截斷而整批漏掉的**(`docs/re/66`)。
// `sub_22F0` 的反編譯版是空的參數列 + 幾行,而組語有完整的三層棄船階梯。
//
// # 一、`sub_22F0`:一次船身損傷判定
//
//	if ((載具 & 0xF8) != 0x20) return          ; ★ 只有大船會受損
//	                                            (0x20..0x27 = 揚帆中 + 收帆)
//	傷害 = rand(1, 30)
//	if (傷害 < 耐久) { 耐久 -= 傷害; return }   ; 撐住了
//	印 "Ship sunk!"
//	if (船上小艇 > 0) {
//	    印 "Abandon ship!"
//	    載具 = 0x28 | (載具 & 3)                ; ★ 換小艇,朝向保留
//	} else if (魔毯 > 0) {
//	    印 "Abandon ship!"
//	    魔毯--
//	    載具 = 0x14 + rand(0, 1)                ; ★ 換魔毯,朝向隨機
//	} else {
//	    載具 = 0                                ; 畫面上隊伍的載具圖消失
//	    印 "Drowning!"
//	    if (還有人能行動 或 有人睡著) {          ; sub_2B67C() != -1
//	        每個沒死的隊員受 rand(1, 8) 傷        ; sub_2A4D0
//	    }
//	}
//
// ⚠ 換成小艇那條**不扣小艇數**(組語裡沒有 `dec`)—— 照原樣。
//
// # 二、觸發點(六處,全部呼叫 `sub_22F0`)
//
//	sub_2D9D0   ★ "Rough seas!" —— 小艇(0x28..0x2B)或魔毯(0x14/0x15)
//	            站在水 tile 1 上;之後那次 sub_22F0 對這兩種載具**是空轉**
//	            (它只動大船),所以對小艇 / 魔毯來說只有訊息與動畫
//	sub_2CE70   撞擊 / 觸礁那條路(音效 sub_2C598(64h, 7D0h, 12Ch))
//	sub_23FC    ×1、sub_24DC ×2、sub_25F0 ×1(戰鬥與地形效果)
//
// # 三、`sub_2A4D0`:對全隊灑傷害
//
//	for (i = 0; i < 6 && i < 隊伍人數; i++)
//	    if (狀態[i] != 'D') sub_2A464(i, rand(1, 8))
//
// # 四、`sub_2B67C`:找第一個能行動的人(順帶把它的語意釘清楚)
//
//	掃隊員:狀態是 'G'(良好)或 'P'(中毒)→ 記下編號、**回 0**(能行動)
//	        'S'(睡著)→ 計數,繼續
//	        其他('D' 死 / 'C' 魅惑…)→ 跳出
//	掃完沒人能行動:有人睡著 → 回 1(全隊睡著);否則 → 回 −1
//
// 這條讓 `sub_1A54` 的 `if (v2 == 1) 印 "Zzzzzz..."` 讀得通,
// 也讓溺水的閘門讀得通:**回 −1 就跳過傷害**,因為那時隊伍已經全滅了。
//
// # 五、`sub_2CCFC`:船轉向要花掉這一步
//
// 這支與移動時印動詞的 `sub_7C0` 幾乎一樣,但**多了大船那一段**:
//
//	新朝向 = (載具 & 0xFC) | 方向
//	if (新朝向 != 載具) {                       ; ★ 轉向
//	    載具 = 新朝向
//	    印 "Head " + 方向名
//	    if (耐久 < 0x32) 印 "Hull weak!"        ; ★ 門檻 50,與上船時的 10 不同
//	    return 1                                ; ★ 這一步被轉向吃掉,不前進
//	}
//	; 已經朝著那個方向
//	if (載具 >= 0x24) return 0                  ; 收帆的船 → 照走
//	if (風 != 0)      return 0                  ; 揚帆 + 有風 → 照走
//	return 1                                    ; ★ 揚帆 + 無風 → 動不了
//
// 回傳 1 = 「這一步用掉了」,呼叫端 `sub_2D174` 收到非 0 就不移動。
//
// ⚠ 最後那一條修掉一個引擎的錯:`CanSail` 原本寫「無風 → 照走」,
// 依據是 `sub_2D38` 在查延遲表**之前**就把無風 `jz` 掉。那句話只對一半 ——
// 無風時**不查表**是對的,但揚著帆就是動不了,判斷在 `sub_2CCFC` 這裡。
// 收帆的船(0x24..0x27)才不受風影響。

// 船身耐久的三個常數,三個不同來源。
const (
	// ShipHullNew 是新買的船的耐久(原版 `sub_2218`:放下 tile 0x2C 時
	// `mov byte ptr dword_3E470+1[ebx*8], 64h`)。
	ShipHullNew = 100
	// ShipHullWeak 是航行中警告「Hull weak!」的門檻(`sub_2CCFC`:`cmp …, 32h`)。
	//
	// ⚠ 與 ShipHullWarning(上船時的 10)是**兩個不同的警告**,別合併。
	ShipHullWeak = 0x32
	// ShipDamageMax 是一次損傷判定的骰上限(`sub_28E14(1, 1Eh)`)。
	ShipDamageMax = 30
)

// DrownDamageMax 是溺水時每個隊員受的傷上限(`sub_2A4D0`:`sub_28E14(1, 8)`)。
const DrownDamageMax = 8

// RoughSeasTile 是會掀起「Rough seas!」的水 tile(`sub_2D9D0`:`cmp esi, 1`)。
//
// ✅ **是深水**:`internal/i18n/look_text.go` 的 `look#1` 就是「深水」,
// 而 `look#<N>` 的鍵直接是 tile 編號(`i18n.LookKey`)。
// 同一批對照還把 `sub_2CE70` 的另外三個地形值認出來(淺灘 / 碼頭 / 仙人掌),
// 四個全部與訊息語意相符 —— 那才是這條的第二份證據。
const RoughSeasTile = 1

// `sub_2CE70`(通行判定 + 撞擊)比對的三個地形。
const (
	// TileShallowWater 淺灘:揚著帆衝進去會「BREAKING UP!」。
	TileShallowWater = 3
	// TileCactus 仙人掌:撞上去「OUCH!」並讓全隊受傷。
	TileCactus = 0x2F
	// TileDock 碼頭:撞上去是「Docked!」—— 靠岸,不是撞擊。
	TileDock = 0x47
)

// ShipTakesDamage 回報這個載具碼會不會吃船身損傷。
//
// 原版是 `(載具 & 0xF8) == 0x20` —— 揚帆(0x20..0x23)與收帆(0x24..0x27)都算,
// 小艇(0x28)不算。
func ShipTakesDamage(transport byte) bool { return transport&0xF8 == VehicleSailing }

// RoughSeasAffects 回報這個載具會不會遇到「Rough seas!」
// (原版 `sub_2D9D0`:小艇 `& 0xFC == 0x28`,或魔毯 `& 0xFE == 0x14`)。
func RoughSeasAffects(transport byte) bool {
	return transport&0xFC == VehicleSkiff || transport&0xFE == VehicleCarpet
}
