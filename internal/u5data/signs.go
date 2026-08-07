package u5data

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// `SIGNS.DAT` —— 城鎮招牌與墓碑
//
// # 格式(2026-08-08 一手驗證,對照 FM Towns `sub_D650` / `sub_D544`)
//
//	0x0000  u16 × 33   每個地點一筆位移(0 = 這個地點沒有招牌)
//	0x0042  …          招牌記錄,一筆接一筆
//
// 一筆記錄:
//
//	[0] 地點編號     ← 只是資料上的註記,**原版比對時不看它**
//	[1] 樓層
//	[2] x
//	[3] y
//	[4..] 招牌內容,NUL 結尾
//
// 33 = 32 個地點 + 地表(索引 0)。原版 `sub_D650` 讀 0x42 B 的表頭,
// 用 `byte_3E0A3`(目前地點)取位移,再從那裡往下**逐筆比對樓層 / x / y**,
// 碰到 0xFF 停。⚠ 比對**不含地點編號**,所以掃描會越過本地點的記錄繼續往下 ——
// 這是原版行為,照做(不同地點的招牌若座標相同會互相竄位,那是原版的 bug)。
//
// # 兩格共用一塊招牌
//
// 內容以 `0x0A` 開頭的記錄是**別名**:渲染器看到 0x0A 就往後跳 6 個位元組,
// 而那 6 個位元組是 `0x0A` + NUL + 下一筆的 4 B 表頭 —— 跳完正好落在共用的內容上。
// 掃描迴圈不認 0x0A,它只是「跳到 NUL 之後」,於是自然停在第二個表頭。
// 兩條路都通到同一段文字,招牌就同時掛在相鄰兩格上。
//
//	1916:  01 00 0e 14 | 0a 00 | 01 00 10 14 | "abbb…"
//	       └ x=0x0E ┘           └ x=0x10 ┘     └ 共用 ┘
//
// # 內容的編碼
//
//   - **bit7 = 一般字**;bit7 清掉 = 反白。原版 `sub_D544` 用它切 `sub_27754(0/1)`。
//     所以 `0xC2` 是普通的 `B`,而 `0x42`('B')是反白的 `B`。
//   - `0x29..0x31` 是**巨集**:一個位元組展開成 16 個字的整列框線
//     (見 `signMacros`)。招牌寬 16 欄,所以每一列剛好一個巨集或 16 個字。
//   - `&`(0x26)與 `'`(0x27)都印成 `l`(橫線)。
//   - `0x0D` 是**分頁**:停下來等玩家按鍵。
//   - 其餘位元組印 `c & 0x7F`。
//
// # 招牌字型就是 `RUNES.CH`(2026-08-08 逐字模驗證)
//
// 招牌不是用一般字型畫的。把 `RUNES.CH` 當 8×8 直索引 dump 出來就看得很清楚:
//
//	'l' 0x6C  ── 一條橫線          'g' 0x67  │ 一條直線
//	'[' 0x5B  符文 TH              'A' 0x41  符文 A
//	0x0E      ✳ 星芒裝飾            '&' 0x26  **空的**
//
// 最後那一條是決定性的:`&` 在 `RUNES.CH` 裡字模全空,所以原版 `sub_D544`
// 才要把 `&` 與 `'` 特判成印 `l`。**先有空字模,才有那個特判** ——
// 這解釋了一條看起來莫名其妙的程式碼分支。
//
// 相對地 `IBM.CH` 的 0x6C 是普通的小寫 `l`、0x61 是 `a`,拿它畫招牌只會得到
// 一堆字母。所以「招牌是美術內嵌字母」這個印象是對的一半:**框是美術,
// 但它是字型裡的美術,不是點陣圖**;框裡的文字則是真的文字。
//
// 框線用的是字母當美術:`a b c d e f g h i j k l m n` 與 `8 9 : ;` 在
// `RUNES.CH` 裡畫的是角、邊、直線。內容文字則是大寫 ASCII,
// 其中五個符號是**符文合字**(同樣由 `RUNES.CH` 畫成單一符文):
//
//	[ = TH    \ = EE    ] = NG    ^ = EA    _ = ST    @ = 空白
//
// 例:`NOR[@BRITAIN` = NORTH BRITAIN、`^_@PAWS` = EAST PAWS、
// `D\P` = DEEP、`TRYI]` = TRYING。
//
// ⚠ 這與 u4-cht 的「美術內嵌字母」**不是同一個問題**:這裡的文字是真的文字,
// 譯得動;只有框線是美術。中文化的取捨寫在 `docs/localization-notes.md`。

// SignLocations 是位移表的項數:地表 + 32 個地點。
const SignLocations = 33

// SignWidth 是招牌的欄寬。巨集展開後每一列剛好這麼寬。
const SignWidth = 16

const (
	signAliasMark = 0x0A // 內容開頭是它 → 這是別名,往後跳 signAliasSkip
	signAliasSkip = 6
	signTableEnd  = 0xFF // 掃描到它就停
	signPageBreak = 0x0D // 分頁:等玩家按鍵
	signMacroLo   = 0x29
	signMacroHi   = 0x31
	signRuleA     = 0x26 // '&'
	signRuleB     = 0x27 // '\''
	signRuleGlyph = 'l'  // 上面兩個都印成橫線
	// signOrnament 是 RUNES.CH 0x0E 的星芒裝飾(「BEWARE ✳ THE DESERT」用它)。
	// 它是**字模**不是控制碼 —— 原版照樣送進輸出層,字型自然畫得出來。
	signOrnament = 0x0E
)

// signMacros 是 0x29..0x31 這九個巨集(原版 `dword_54E2C[c*4]`)。
//
// 每一條都是 16 個字的整列,拼起來就是招牌的外框與墓碑的圓頂。
var signMacros = [signMacroHi - signMacroLo + 1]string{
	"g              g", // 0x29 空列(兩側直線)
	"jlllllllnllllllk", // 0x2A 下框,中間有接點
	"8lllllllmllllll9", // 0x2B 上框,中間有接點
	"jllllllllllllllk", // 0x2C 下框
	"8llllllllllllll9", // 0x2D 上框
	"hllllk    jlllli", // 0x2E 墓碑肩
	"jlllli    hllllk", // 0x2F 墓碑腰
	"     g    g     ", // 0x30 墓碑柱
	"     hlllli     ", // 0x31 墓碑頂
}

// signRunes 是符文合字。key 是招牌裡的 ASCII,值是它代表的字母。
var signRunes = map[byte]string{
	signOrnament: "*",
	'[':  "TH",
	'\\': "EE",
	']':  "NG",
	'^':  "EA",
	'_':  "ST",
	'@':  " ",
}

// Sign 是一塊招牌。
type Sign struct {
	Location int // 資料上註記的地點(原版比對時不看)
	Floor    int
	X, Y     int
	// Raw 是渲染前的內容位元組(巨集未展開,bit7 未清)。
	Raw []byte
}

// SignSet 是 `SIGNS.DAT` 解出來的全部招牌。
type SignSet struct {
	// offsets[loc] 是該地點在檔案裡的起掃位置;0 代表沒有招牌。
	offsets [SignLocations]int
	raw     []byte
}

// LoadSigns 讀 `SIGNS.DAT`。
func LoadSigns(dir string) (*SignSet, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "SIGNS.DAT"))
	if err != nil {
		return nil, err
	}
	return ParseSigns(raw)
}

// ParseSigns 解一份 `SIGNS.DAT` 的內容。
func ParseSigns(raw []byte) (*SignSet, error) {
	const header = SignLocations * 2
	if len(raw) < header {
		return nil, fmt.Errorf("SIGNS.DAT 只有 %d B,連 %d B 的位移表都放不下", len(raw), header)
	}
	s := &SignSet{raw: raw}
	for i := range s.offsets {
		off := int(binary.LittleEndian.Uint16(raw[i*2:]))
		if off != 0 && (off < header || off >= len(raw)) {
			return nil, fmt.Errorf("地點 %d 的位移 %d 落在資料區外", i, off)
		}
		s.offsets[i] = off
	}
	return s, nil
}

// At 找某個地點、某個座標上的招牌。
//
// ⚠ 照原版走:從該地點的位移起往下掃,**只比樓層與座標**,碰到 0xFF 停。
// 掃描會越過本地點的記錄 —— 不要「順手」加上地點比對,那會改掉原版行為。
func (s *SignSet) At(location, floor, x, y int) (*Sign, bool) {
	if location < 0 || location >= SignLocations {
		return nil, false
	}
	i := s.offsets[location]
	if i == 0 {
		return nil, false
	}
	for i+4 <= len(s.raw) && s.raw[i] != signTableEnd {
		rec := s.raw[i : i+4]
		end := i + 4
		for end < len(s.raw) && s.raw[end] != 0 {
			end++
		}
		if int(rec[1]) == floor && int(rec[2]) == x && int(rec[3]) == y {
			return &Sign{
				Location: int(rec[0]),
				Floor:    int(rec[1]),
				X:        int(rec[2]),
				Y:        int(rec[3]),
				Raw:      s.follow(i + 4),
			}, true
		}
		i = end + 1
	}
	return nil, false
}

// follow 處理別名:內容以 0x0A 開頭就往後跳 6 個位元組,再取到 NUL 為止。
func (s *SignSet) follow(i int) []byte {
	for i < len(s.raw) && s.raw[i] == signAliasMark {
		i += signAliasSkip
	}
	end := i
	for end < len(s.raw) && s.raw[end] != 0 {
		end++
	}
	if i > end {
		return nil
	}
	return s.raw[i:end]
}

// SignCell 是招牌上的一個字。
type SignCell struct {
	Ch        byte // 已經清掉 bit7 的字元
	Highlight bool // 反白(原版 bit7 = 0 的那些)
	PageBreak bool // 這裡要停下來等按鍵;此時 Ch 無意義
	NewLine   bool // 換行;此時 Ch 無意義
}

// Render 把招牌內容展開成一格一格的字。
//
// 每一列的長度有兩種來源:用巨集畫框的招牌一律 SignWidth 欄,靠自動折行;
// **窄招牌則自己帶明確換行** —— `0x8A`(= `0x0A | bit7`)。原版 `sub_D544`
// 把它當一般字印出 `c & 0x7F` = 0x0A,而輸出層看到 0x0A 就換行。
//
// ⚠ 別把它跟開頭那個裸的 `0x0A`(別名記號)搞混:別名在 `follow` 就處理掉了,
// 走到這裡的 0x0A 一律是換行。
func (sg *Sign) Render() []SignCell {
	var out []SignCell
	for _, c := range sg.Raw {
		switch {
		case c&0x7F == signAliasMark:
			out = append(out, SignCell{NewLine: true})
		case c == signPageBreak:
			out = append(out, SignCell{PageBreak: true})
		case c == signRuleA || c == signRuleB:
			out = append(out, SignCell{Ch: signRuleGlyph, Highlight: true})
		case c >= signMacroLo && c <= signMacroHi:
			// 巨集一律是一般字(原版展開前先 sub_27754(1),
			// 而那一支的參數與 bit7 那條路的 0/1 相反 —— 框線不反白)。
			for i := 0; i < len(signMacros[c-signMacroLo]); i++ {
				out = append(out, SignCell{Ch: signMacros[c-signMacroLo][i]})
			}
		default:
			out = append(out, SignCell{Ch: c & 0x7F, Highlight: c&0x80 == 0})
		}
	}
	return out
}

// Lines 把招牌切成一列 SignWidth 個字的字串,符文合字展開成字母。
//
// 這是給工作單與沒有符文字型的路徑用的;真正上畫面時用 Render 才留得住反白。
func (sg *Sign) Lines() []string {
	var lines []string
	var row strings.Builder
	n := 0
	flush := func() {
		if n > 0 {
			lines = append(lines, row.String())
			row.Reset()
			n = 0
		}
	}
	for _, c := range sg.Render() {
		if c.PageBreak || c.NewLine {
			flush()
			continue
		}
		if r, ok := signRunes[c.Ch]; ok {
			row.WriteString(r)
		} else {
			row.WriteByte(c.Ch)
		}
		n++
		if n == SignWidth {
			flush()
		}
	}
	flush()
	return lines
}

// SignExpandRunes 把符文合字展開成字母。翻譯工作單用。
func SignExpandRunes(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if r, ok := signRunes[s[i]]; ok {
			b.WriteString(r)
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// All 走過全部地點,回傳每一塊招牌。工作單與測試用。
//
// 別名記錄(內容只有 0x0A)會解到它指向的共用內容,所以同一段文字會出現兩次 ——
// 那正是原版的樣子:兩格各有一塊招牌,只是內容相同。
func (s *SignSet) All() []Sign {
	var out []Sign
	for loc := 0; loc < SignLocations; loc++ {
		i := s.offsets[loc]
		if i == 0 {
			continue
		}
		for i+4 <= len(s.raw) && s.raw[i] != signTableEnd {
			rec := s.raw[i : i+4]
			end := i + 4
			for end < len(s.raw) && s.raw[end] != 0 {
				end++
			}
			if int(rec[0]) != loc {
				break // 掃到別的地點的記錄就停 —— 列清單不需要複製原版的越界行為
			}
			out = append(out, Sign{
				Location: int(rec[0]), Floor: int(rec[1]),
				X: int(rec[2]), Y: int(rec[3]),
				Raw: s.follow(i + 4),
			})
			i = end + 1
		}
	}
	return out
}
