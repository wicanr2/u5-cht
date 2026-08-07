package u5data

import (
	"encoding/binary"
	"fmt"
	"image"
	"os"
	"path/filepath"
)

// `.16` / `.4` 圖檔的內部結構(2026-08-07 破解)
//
// LZW 解壓之後(見 `lzw.go`)是一張形狀表:
//
//	u16  N                             形狀數
//	N ×  { u16 圖位移, u16 遮罩位移 }    位移 0 = 這一格沒有東西
//	每個 blob: u16 寬, u16 高, 之後是逐列的像素
//
// **列的位元組數三種深度各不相同**,這是最容易錯的一點:
//
//	EGA `.16`  4bpp packed,列寬補到 **8 像素的倍數** → roundUp(w,8)/2 B
//	CGA `.4`   2bpp packed,**不補** → ceil(w/4) B
//	遮罩       1bpp,列寬補到 8 像素 → roundUp(w,8)/8 B
//
// EGA 補到 8 像素(= 4 B)這條不是猜的:`CREATE.16` 的 51×67 那一張,
// 若照 ceil(51/2)=26 B 算,下一個形狀的位移會差 134 B;照 56/2=28 B 算
// 就正好接上。全 25 個 `.16` 檔、268 個 blob **一個不差**。
// 而 CGA 那邊反過來 —— 同一張 51×67 用 ceil(51/4)=13 B 才對得上。
// 兩種深度用不同的補齊規則,套錯就是整批位移全垮。
//
// 像素是 **packed 不是 planar**。兩種解讀在 24 px 寬時列長剛好都是 12 B,
// 分不出來 —— 是畫出來才確定的:packed 是一片有木紋的洞穴壁,
// planar 是彩色雜訊(`rulebook/64` 的「已知輸出當 oracle」)。

// 地牢透視圖組
const (
	// DungeonViewShapes 是 `DNG*.16` 的形狀格數(其中兩格是空的)。
	DungeonViewShapes = 28
	// DungeonViewHeight 是每一張透視切片的高度 —— 全部一樣高。
	DungeonViewHeight = 164
	// DungeonThemes 是三種外觀:DNG1 洞穴、DNG2 熔岩、DNG3 磚牆。
	DungeonThemes = 3
)

// dungeonPictureName 回傳第 n 個地牢圖組的檔名(n 從 1 起)。
func dungeonPictureName(n int) string { return fmt.Sprintf("DNG%d.16", n) }

// Picture 是一個形狀:寬高 + 每像素一個色號(標準 EGA),可選的透明遮罩。
type Picture struct {
	Width, Height int
	// Pix[y*Width+x] 是色號。
	Pix []byte
	// Mask 為 nil 表示整張不透明;否則 **Mask[y*Width+x] != 0 代表這一點是透明的**。
	//
	// ⚠ 極性容易搞反。原版 EGA 的疊圖是 `畫面 = (畫面 AND 遮罩) OR 圖`,
	// 所以遮罩位元 **1 = 保留背景 = 透明**。
	// 證據:`MON0.16` 第 0 個形狀 24×66,遮罩有 749 個 1,而圖裡色號 0
	// 有 1,083 個 —— 兩者同一個數量級但不相等,正是「大部分黑色是背景、
	// 少部分黑色是刻意畫的輪廓」該有的樣子。反過來解讀會把整隻怪物挖空。
	Mask []byte
}

// PictureSet 是一個 `.16` / `.4` 檔裡的全部形狀。
//
// 索引與檔案裡的位移表一一對應;**沒有內容的那幾格是 nil,不會被壓掉** ——
// 索引就是原版程式用的編號,壓掉會讓後面全部偏一格。
type PictureSet []*Picture

// pictureDepth 是像素深度。
type pictureDepth int

const (
	depthEGA pictureDepth = 4 // `.16`
	depthCGA pictureDepth = 2 // `.4`
)

// rowBytes 算一列佔幾個位元組。
func rowBytes(w int, depth pictureDepth) int {
	if depth == depthCGA {
		// CGA 不補齊。
		return (w*int(depth) + 7) / 8
	}
	// EGA 與遮罩補到 8 像素的倍數。
	return (w + 7) / 8 * 8 * int(depth) / 8
}

// ParsePictures 解析解壓後的 `.16` / `.4` 內容。
func ParsePictures(raw []byte, ega bool) (PictureSet, error) {
	depth := depthCGA
	if ega {
		depth = depthEGA
	}
	if len(raw) < 2 {
		return nil, fmt.Errorf("只有 %d B,連形狀數都讀不到", len(raw))
	}
	n := int(binary.LittleEndian.Uint16(raw))
	if n == 0 || 2+4*n > len(raw) {
		return nil, fmt.Errorf("形狀數 %d 放不進 %d B 的檔案", n, len(raw))
	}
	set := make(PictureSet, n)
	for i := 0; i < n; i++ {
		imgOff := int(binary.LittleEndian.Uint16(raw[2+4*i:]))
		maskOff := int(binary.LittleEndian.Uint16(raw[4+4*i:]))
		if imgOff == 0 {
			continue
		}
		p, err := readBlob(raw, imgOff, depth)
		if err != nil {
			return nil, fmt.Errorf("形狀 %d: %w", i, err)
		}
		if maskOff != 0 {
			m, err := readBlob(raw, maskOff, 1)
			if err != nil {
				return nil, fmt.Errorf("形狀 %d 的遮罩: %w", i, err)
			}
			if m.Width != p.Width || m.Height != p.Height {
				return nil, fmt.Errorf("形狀 %d 的遮罩是 %d×%d,圖是 %d×%d",
					i, m.Width, m.Height, p.Width, p.Height)
			}
			p.Mask = m.Pix
		}
		set[i] = p
	}
	return set, nil
}

// readBlob 讀一個 blob(u16 寬、u16 高、逐列像素)。
func readBlob(raw []byte, off int, depth pictureDepth) (*Picture, error) {
	if off+4 > len(raw) {
		return nil, fmt.Errorf("位移 %d 超出檔案(%d B)", off, len(raw))
	}
	w := int(binary.LittleEndian.Uint16(raw[off:]))
	h := int(binary.LittleEndian.Uint16(raw[off+2:]))
	if w <= 0 || h <= 0 || w > 4096 || h > 4096 {
		return nil, fmt.Errorf("尺寸 %d×%d 不合理", w, h)
	}
	rb := rowBytes(w, depth)
	if off+4+h*rb > len(raw) {
		return nil, fmt.Errorf("%d×%d(每列 %d B)需要 %d B,但位移 %d 之後只剩 %d B",
			w, h, rb, 4+h*rb, off, len(raw)-off)
	}
	p := &Picture{Width: w, Height: h, Pix: make([]byte, w*h)}
	perByte := 8 / int(depth)
	mask := byte(1<<uint(depth)) - 1
	for y := 0; y < h; y++ {
		row := raw[off+4+y*rb:]
		for x := 0; x < w; x++ {
			b := row[x/perByte]
			// 高位在左。
			shift := uint(8 - int(depth)*(x%perByte+1))
			p.Pix[y*w+x] = (b >> shift) & mask
		}
	}
	return p, nil
}

// LoadPictures 讀一個 `.16` / `.4` 檔並解壓解析。
func LoadPictures(path string) (PictureSet, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	out, err := Decompress(raw)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filepath.Base(path), err)
	}
	set, err := ParsePictures(out, filepath.Ext(path) == ".16")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filepath.Base(path), err)
	}
	return set, nil
}

// Sheet 把一組形狀橫排成一張圖,供目視驗收。
func (s PictureSet) Sheet(gap int) *image.NRGBA {
	w, h := 0, 0
	for _, p := range s {
		if p == nil {
			continue
		}
		w += p.Width + gap
		if p.Height > h {
			h = p.Height
		}
	}
	if w == 0 || h == 0 {
		return image.NewNRGBA(image.Rect(0, 0, 1, 1))
	}
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	x := 0
	for _, p := range s {
		if p == nil {
			continue
		}
		for py := 0; py < p.Height; py++ {
			for px := 0; px < p.Width; px++ {
				if p.Mask != nil && p.Mask[py*p.Width+px] != 0 {
					continue
				}
				img.SetNRGBA(x+px, py, EGAPalette[p.Pix[py*p.Width+px]&0x0F])
			}
		}
		x += p.Width + gap
	}
	return img
}

// errf 是 test helper 也用得到的小包裝。
func errf(format string, a ...any) error { return fmt.Errorf(format, a...) }
