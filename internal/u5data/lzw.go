package u5data

import (
	"encoding/binary"
	"fmt"
)

// U5 的圖檔壓縮 —— 就是標準 LZW
//
// 25 個 `.16`(EGA)與 `.4`(CGA)圖檔都是同一個格式:
//
//	u32  解壓後長度(little-endian)
//	之後 LZW 位元流
//
// LZW 的參數:**LSB-first 打包、碼寬 9 → 12 bit、256 = 清表、257 = 結束、
// 258 起是字典、不做 early change**。這是 TIFF `LZWDecode` 不含 early-change
// 的那一版,也就是 GIF 的變體。
//
// ⚠ `docs/formats/01` 曾記載「不是標準 LZW(48 種參數組合最長只符 4 B)」——
// **那條是錯的**。錯在只比對了「解出來的位元組要對得上 FM Towns 的 `EGA*.TIL`」,
// 而兩者其實還差一層色號順序(見 `tileColorRemap`),所以正確的解壓器
// 在第 5 個位元組就被判定失敗。**驗收條件本身有洞,會讓對的答案看起來是錯的。**
//
// 現在的驗收是決定性的:`TILES.16` 解出來的 65,536 B 與「FM Towns 四個
// `EGA*.TIL` 降回 16×16、換過色號、壓回 4bpp」**逐位元組完全相同**
//(`TestDOSTilesMatchFMTowns`)。兩個獨立來源對上,不是「看起來對」。

const (
	lzwClear     = 256
	lzwEnd       = 257
	lzwFirstCode = 258
	lzwMaxWidth  = 12
)

// Decompress 解開一個 `.16` / `.4` 圖檔(含 4 B 長度檔頭)。
//
// 解出來的長度與檔頭宣稱的不符就報錯 —— 那是「解壓器參數不對」最直接的訊號,
// 不要讓它安靜地回一段半截資料。
func Decompress(raw []byte) ([]byte, error) {
	if len(raw) < 4 {
		return nil, fmt.Errorf("檔案只有 %d B,連 4 B 的長度檔頭都放不下", len(raw))
	}
	want := int(binary.LittleEndian.Uint32(raw))
	out := lzwDecode(raw[4:], want)
	if len(out) != want {
		return nil, fmt.Errorf("解出 %d B,檔頭宣稱 %d B —— 解壓參數不對", len(out), want)
	}
	return out, nil
}

// lzwDecode 是解壓核心。hint 只用來預先配置空間,不影響結果。
func lzwDecode(src []byte, hint int) []byte {
	out := make([]byte, 0, hint)

	// dict[c] 是碼 c 展開後的位元組。前 256 個是字面值,256/257 是控制碼。
	var dict [1 << lzwMaxWidth][]byte
	var reset func()
	width := 9
	next := lzwFirstCode
	reset = func() {
		for i := 0; i < 256; i++ {
			dict[i] = []byte{byte(i)}
		}
		width, next = 9, lzwFirstCode
	}
	reset()

	var acc uint32 // 位元暫存(LSB-first:先進來的位元在低位)
	var nbits uint
	pos := 0
	var prev []byte

	for {
		for nbits < uint(width) {
			if pos >= len(src) {
				return out
			}
			acc |= uint32(src[pos]) << nbits
			nbits += 8
			pos++
		}
		code := int(acc & (1<<uint(width) - 1))
		acc >>= uint(width)
		nbits -= uint(width)

		switch {
		case code == lzwClear:
			reset()
			prev = nil
			continue
		case code == lzwEnd:
			return out
		}

		var entry []byte
		switch {
		case code < next && dict[code] != nil:
			entry = dict[code]
		case prev != nil:
			// KwKwK:字典裡還沒有這個碼,它一定是「前一段 + 前一段的首位元組」。
			entry = append(append([]byte{}, prev...), prev[0])
		default:
			return out
		}
		out = append(out, entry...)

		if prev != nil && next < len(dict) {
			dict[next] = append(append([]byte{}, prev...), entry[0])
			next++
			// ⚠ 加寬的時機:字典**滿了才加**(不是 TIFF 的 early change 提前一格)。
			// 差這一格會讓整條位元流從第二次加寬起全錯,而症狀是
			// 「前面幾百個位元組是對的,後面變雜訊」—— 很容易誤判成別的問題。
			if next >= 1<<uint(width) && width < lzwMaxWidth {
				width++
			}
		}
		prev = entry
	}
}
