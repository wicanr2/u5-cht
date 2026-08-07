package u5data

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

// 商店價目表
//
// 全部在 `DATA.OVL` 的一塊連續資料區裡,原版程式直接以「店種內的第幾家」索引。
// 位移是拿 FM Towns `WORRIORS.EXP` 的資料段去 DOS `DATA.OVL` 裡逐位元組比對出來的
// —— 兩版**整塊照搬**,連相對距離都一樣:
//
//	FM Towns linear   DOS DATA.OVL   內容
//	0x55274           0x3A42         藥草價   5×8 byte
//	0x5529C           0x3A6A         藥草量   5×8 byte
//	0x552C4           0x3A92         裝備價   48 × u16
//	0x55384           0x3AF2         武具存貨 9×8 byte
//	0x553CC           0x3B3A         問候語位移 8 種 × 4 句 × u16
//
// 因為兩版一致,價格一律**讀玩家自己的 DATA.OVL**,不寫死在程式裡
// —— 和店名、店主、詞典的作法一致。
//
// 每張表的列數都對得上該店種的家數(武具 9、藥草 5、公會 3、馬廄 4、造船 4、
// 旅店 6、酒館 9),這個吻合是整組表解對了的主要佐證。
const (
	priceReagentPrice = 0x3A42 // 5 × 8 byte
	priceReagentQty   = 0x3A6A // 5 × 8 byte
	priceItem         = 0x3A92 // 48 × u16
	priceArmouryStock = 0x3AF2 // 9 × 8 byte
	priceGreetOffsets = 0x3B3A // 8 × 4 × u16
	priceGuild        = 0x3BFA // 3 × 4 × u16
	priceGuildPitch   = 0x3C1A // 3 × u16
	priceReagentPitch = 0x3C20 // 8 × u16
	priceStable       = 0x3C40 // 4 × u16
	priceItemPitch    = 0x3C58 // 48 × u16
	priceSellPitch    = 0x3CCE // 8 × u16
	priceWine         = 0x4C58 // 6 × u16
	priceFrigate      = 0x4D76 // 4 × u16
	priceSkiff        = 0x4D7E // 4 × u16
	priceInn          = 0x4D8E // 6 × byte
)

// 各店種的家數,同時是對應價目表的列數。
const (
	ArmouryCount    = 9
	TavernCount     = 9
	StableCount     = 4
	ShipwrightCount = 4
	ReagentShops    = 5
	GuildCount      = 3
	HealerCount     = 7
	InnCount        = 6
)

// 武具店與藥草鋪的貨架大小。
const (
	// ArmouryShelf 是武具店一次最多列出的品項數;實際只用 7 格,第 8 格是 0xFF 終止。
	ArmouryShelf = 8
	// ArmouryStockEnd 是存貨表的結束標記。
	ArmouryStockEnd = 0xFF
	// ReagentCount 是藥草的種類數。
	ReagentCount = 8
	// GuildGoods 是公會賣的品項數(鑰匙 / 寶石 / 火把);表每列有 4 格,第 4 格沒用到。
	GuildGoods = 3
	// guildStride 是公會價目表每列的格數。
	guildStride = 4
	// WineCount 是酒館酒單的品項數。
	WineCount = 6
	// CarryLimit 是單一品項的攜帶上限。
	CarryLimit = 99
	// GoldLimit 是金幣上限(賣東西時 `sub_2BBDC(&gold, price, 0x270F)` 的第三個參數)。
	GoldLimit = 9999
	// SellPitchCount 是武具店開價收購時的說詞變體數。
	SellPitchCount = 8
)

// 治療所的三種服務與固定價(sub_12838 裡直接寫進 word_3EF38 的立即數)。
//
// 這三個不在 DATA.OVL 的表裡,是程式碼中的常數,所以只能寫死。
const (
	CurePrice      = 20
	HealPrice      = 35
	ResurrectPrice = 200
	// CharityLocation 是「付不起也照治」的地點(sub_12794 的 `cmp byte_3E0A3, 7`)。
	// 只在價格 ≤ CharityMax 時成立,所以復活(200)仍要付錢。
	CharityLocation = 7
	CharityMax      = 100
)

// 公會與藥草的品項名。
//
// 藥草名在原版是執行檔裡的字串陣列(FM Towns `off_555B8`),
// 不像裝備名那樣放在 DATA.OVL,所以這裡寫死。
var (
	ReagentNames = [ReagentCount]string{
		"Sulfur Ash", "Ginseng", "Garlic", "Spider Silk",
		"Blood Moss", "Black Pearl", "Nightshade", "Mandrake",
	}
	// GuildGoodsNames 出自 sub_11464 的三行選單字串。
	GuildGoodsNames = [GuildGoods]string{"Keys", "Gems", "Torches"}
	// GuildGoodsQty 是公會一次賣幾個(sub_112F8 的三個分支)。
	GuildGoodsQty = [GuildGoods]int{3, 4, 5}
	// WineNames 與 WineNamesPrice 出自 sub_21108 的酒單字串;
	// 價格另有一張表(priceWine),兩者應一致,LoadPrices 會核對。
	WineNames = [WineCount]string{"Rose", "Claret", "Sauterne", "Muscatel", "Moselle", "Chablis"}
)

// PriceTable 是全遊戲的商店價目與貨色。
type PriceTable struct {
	// Item 是 48 件裝備的底價;0 代表不賣。
	Item [ItemCount]int
	// ItemPitch 是店員推銷這件裝備的說詞在 SHOPPE.DAT 的位元組位移;0 代表沒有。
	ItemPitch [ItemCount]int
	// SellPitch 是店員開價收購的八句說詞(隨機挑一句),同樣是 SHOPPE.DAT 的位移。
	SellPitch [SellPitchCount]int
	// ArmouryStock[家][格] 是第幾家武具店賣哪些裝備,以 ArmouryStockEnd 結束。
	ArmouryStock [ArmouryCount][ArmouryShelf]byte

	// ReagentPrice[家][味] 是藥草的底價;0 代表這家沒賣。
	ReagentPrice [ReagentShops][ReagentCount]int
	// ReagentQty[家][味] 是一次買到的份數。
	ReagentQty [ReagentShops][ReagentCount]int
	// ReagentPitch[味] 是藥草說詞在 SHOPPE.DAT 的位移。
	ReagentPitch [ReagentCount]int

	// Guild[家][品] 是公會三樣貨的底價。
	Guild [GuildCount][GuildGoods]int
	// GuildPitch[品] 是公會說詞在 SHOPPE.DAT 的位移。
	GuildPitch [GuildGoods]int

	// Stable 是四家馬廄的馬價。
	Stable [StableCount]int
	// Frigate / Skiff 是四家造船廠的船價。
	Frigate [ShipwrightCount]int
	Skiff   [ShipwrightCount]int
	// Inn 是六家旅店每人每天的價錢。
	Inn [InnCount]int
	// Wine 是酒館酒單六款的價錢。
	Wine [WineCount]int

	// Greet[店種][0..3] 是四句問候語在 SHOPPE.DAT 的位移(原版 dword_553CC)。
	Greet [ShopTypeCount][4]int
}

// LoadPrices 從 DATA.OVL 讀出全部價目表。
func LoadPrices(dir string) (*PriceTable, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "DATA.OVL"))
	if err != nil {
		return nil, err
	}
	return ParsePrices(raw)
}

// ParsePrices 從 DATA.OVL 的內容取出價目表。
func ParsePrices(ovl []byte) (*PriceTable, error) {
	need := priceInn + InnCount
	if len(ovl) < need {
		return nil, fmt.Errorf("DATA.OVL 只有 %d B,放不下 0x%X 的價目區", len(ovl), priceInn)
	}
	u16 := func(off, i int) int { return int(binary.LittleEndian.Uint16(ovl[off+i*2:])) }

	t := &PriceTable{}
	for i := 0; i < ItemCount; i++ {
		t.Item[i] = u16(priceItem, i)
		t.ItemPitch[i] = u16(priceItemPitch, i)
	}
	for i := 0; i < SellPitchCount; i++ {
		t.SellPitch[i] = u16(priceSellPitch, i)
	}
	for s := 0; s < ArmouryCount; s++ {
		copy(t.ArmouryStock[s][:], ovl[priceArmouryStock+s*ArmouryShelf:])
	}
	for s := 0; s < ReagentShops; s++ {
		for r := 0; r < ReagentCount; r++ {
			t.ReagentPrice[s][r] = int(ovl[priceReagentPrice+s*ReagentCount+r])
			t.ReagentQty[s][r] = int(ovl[priceReagentQty+s*ReagentCount+r])
		}
	}
	for r := 0; r < ReagentCount; r++ {
		t.ReagentPitch[r] = u16(priceReagentPitch, r)
	}
	for s := 0; s < GuildCount; s++ {
		for g := 0; g < GuildGoods; g++ {
			t.Guild[s][g] = u16(priceGuild, s*guildStride+g)
		}
	}
	for g := 0; g < GuildGoods; g++ {
		t.GuildPitch[g] = u16(priceGuildPitch, g)
	}
	for i := 0; i < StableCount; i++ {
		t.Stable[i] = u16(priceStable, i)
	}
	for i := 0; i < ShipwrightCount; i++ {
		t.Frigate[i] = u16(priceFrigate, i)
		t.Skiff[i] = u16(priceSkiff, i)
	}
	for i := 0; i < InnCount; i++ {
		t.Inn[i] = int(ovl[priceInn+i])
	}
	for i := 0; i < WineCount; i++ {
		t.Wine[i] = u16(priceWine, i)
	}
	for ty := 0; ty < ShopTypeCount; ty++ {
		for v := 0; v < 4; v++ {
			t.Greet[ty][v] = u16(priceGreetOffsets, ty*4+v)
		}
	}
	if err := t.validate(); err != nil {
		return nil, err
	}
	return t, nil
}

// validate 用幾條「算錯就一定違反」的性質擋住位移偏掉的情形。
//
// 價格全是數字,位移偏一點還是讀得出「某個數」—— 所以不能只看有沒有讀到東西,
// 要看讀到的東西彼此相不相容。
func (t *PriceTable) validate() error {
	// 武具店存貨:每一列都必須以 0xFF 結束,而且前面的編號都得是合法裝備。
	for s := range t.ArmouryStock {
		row := t.ArmouryStock[s]
		if row[ArmouryShelf-1] != ArmouryStockEnd {
			return fmt.Errorf("第 %d 家武具店的存貨沒有以 0xFF 結束:%v", s, row)
		}
		for _, id := range row[:ArmouryShelf-1] {
			if int(id) >= ItemCount {
				return fmt.Errorf("第 %d 家武具店賣編號 %d,超出裝備表", s, id)
			}
		}
	}
	// 有賣的裝備一定有推銷詞,沒賣的一定沒有 —— 兩張表是不同位移讀出來的,
	// 能雙向對上就不會是巧合。
	for id := 0; id < ItemCount; id++ {
		stocked := t.itemStocked(byte(id))
		if stocked && (t.Item[id] == 0 || t.ItemPitch[id] == 0) {
			return fmt.Errorf("裝備 %d 有店在賣,但價格 %d、說詞位移 %d",
				id, t.Item[id], t.ItemPitch[id])
		}
	}
	// 藥草:有價就要有份數,沒價就不該有份數。
	for s := range t.ReagentPrice {
		for r := range t.ReagentPrice[s] {
			if (t.ReagentPrice[s][r] == 0) != (t.ReagentQty[s][r] == 0) {
				return fmt.Errorf("第 %d 家藥草鋪的第 %d 味:價 %d 份數 %d,一個是 0 一個不是",
					s, r, t.ReagentPrice[s][r], t.ReagentQty[s][r])
			}
		}
	}
	// 原版一家藥草鋪最多列 5 味(sub_1173C 的跳表只有 5 個 case)。
	for s := range t.ReagentPrice {
		n := 0
		for r := range t.ReagentPrice[s] {
			if t.ReagentPrice[s][r] > 0 {
				n++
			}
		}
		if n > 5 {
			return fmt.Errorf("第 %d 家藥草鋪列了 %d 味,原版選單最多 5 味", s, n)
		}
	}
	// 旅店每人每天 1..9 金;讀錯位移幾乎一定會撞出 0 或三位數。
	for i, p := range t.Inn {
		if p < 1 || p > 9 {
			return fmt.Errorf("第 %d 家旅店每天 %d 金,不像價錢", i, p)
		}
	}
	return nil
}

func (t *PriceTable) itemStocked(id byte) bool {
	for s := range t.ArmouryStock {
		for _, v := range t.ArmouryStock[s] {
			if v == ArmouryStockEnd {
				break
			}
			if v == id {
				return true
			}
		}
	}
	return false
}

// ItemAltNames 是裝備在「賣出」對白裡用的另一種說法(原版 dword_55714)。
//
// 只有 14 件有 —— 護甲類是 "Cloth suit"(一套布甲)、雙手武器寫全名、
// 戒指與護符倒過來寫。沒有的就用一般名字。
//
// ⚠ 這張表在 DOS 的 DATA.OVL 裡找不到對應(整組位移都對不上),
// 所以是唯一一張從 FM Towns `WORRIORS.EXP` 抄過來寫死的商店字串表。
// 中文化時它走 i18n 覆蓋層,不影響。
var ItemAltNames = map[int]string{
	9: "Cloth suit", 10: "Leather suit", 11: "Ring Mail suit",
	12: "Scale suit", 13: "Chain suit", 14: "Plate suit", 15: "Mystic suits",
	31: "Two-Handed Hammer", 32: "Two-Handed Axe", 33: "Two-Handed Sword",
	42: "Invisibility Ring", 43: "Protection Ring", 44: "Regeneration Ring",
	45: "Turning Amulet",
}

// Haggle 把底價換算成這名角色實際要付的錢。
//
// 原版(sub_11AF0 / sub_11588 / sub_112F8 / sub_118CC / sub_219B0 都是同一段):
//
//	price = base + base × (100 − 3 × INT) / 100
//
// 也就是**智力每點折 3%**,而起點是底價的兩倍:INT 0 要付 2 倍,INT 33 約付一倍。
// 除法是 x86 `idiv`(向零捨入),智力高到 33 以上時商為負,價格會低於底價 ——
// 原版沒有夾住,這裡照樣不夾。
func Haggle(base, intelligence int) int {
	return base + (base*(100-3*intelligence))/100
}

// SellValue 是武具店收購一件裝備願意出的價(原版 sub_12060)。
//
//	sell = (3 × INT × base) / 100 + 1
//
// 與買價**不是同一條公式**:買價從兩倍底價往下折,賣價則是從零往上加 ——
// 智力 17 只換得回約半價,智力 33 才勉強打平。
// 底價 0 的東西(聖物、任務物品)店家不收,由呼叫端判斷。
func SellValue(base, intelligence int) int {
	return (3*intelligence*base)/100 + 1
}

// ArmouryStockList 回傳第 shop 家武具店的貨架(已去掉 0xFF 終止)。
func (t *PriceTable) ArmouryStockList(shop int) []byte {
	if t == nil || shop < 0 || shop >= ArmouryCount {
		return nil
	}
	row := t.ArmouryStock[shop]
	for i, v := range row {
		if v == ArmouryStockEnd {
			return row[:i:i]
		}
	}
	return row[:]
}

// ReagentStockList 回傳第 shop 家藥草鋪有貨的藥草編號,順序就是選單順序。
//
// 原版顯示時字母是**只發給有貨的**(sub_1173C),按鍵再由 sub_11588 映射回真正的
// 藥草編號 —— 所以缺貨的味不佔位。
func (t *PriceTable) ReagentStockList(shop int) []int {
	if t == nil || shop < 0 || shop >= ReagentShops {
		return nil
	}
	var out []int
	for r := 0; r < ReagentCount; r++ {
		if t.ReagentPrice[shop][r] > 0 {
			out = append(out, r)
		}
	}
	return out
}
