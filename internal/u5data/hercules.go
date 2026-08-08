package u5data

// Hercules 的畫面幾何:640×300 置中在 720×348 —— 全部從 `HER.DRV` 讀出來
//
// `docs/re/64` 原本留了一句「blit 迴圈還沒追完(縱向 1.5 倍、加倍表怎麼與
// 圖樣表配合)」。追完了,而答案比預期乾淨:**兩張表根本不在同一個軸上**。
//
// # 一、驅動程式有一張 38 筆的進入點表
//
// `HER.DRV` 開頭是 38 個 `jmp near`(每筆 3 B,共 0x72 B):
//
//	#15 @0x09E4  設定畫筆顏色(色號 0..3 → 圖樣位元組,存在 0x0334)
//	#16 @0x09F4  畫點(內層在 0x0A3C)
//	#26 @0x118C  畫 **1bpp** 圖形(shape blit)
//	#31 @0x1282  畫字(`.HCS` 的 16×12 字模)
//	#34 @0x1C83  畫面溶解特效(LFSR 逐點消去 + 關 PC 喇叭)
//	           0x06F6 是「沒實作」的 `retf` 樁,有 8 筆指到它
//
// 有進入點表這件事本身就是禮物:不必猜哪一段是什麼,呼叫編號就是語意。
//
// # 二、掃描線位址表在 0x0072,300 筆
//
// 表**不在檔案裡**(0x0072 起 600 B 在磁碟上全是 0),由 0x0770 的初始化迴圈建:
//
//	mov si, 0072h        ; ★ 剛好接在 38 筆進入點表(0x72 B)之後
//	mov cx, 012Ch        ; ★ 300 條掃描線
//	mov bx, 005Ah        ; 每列 90 位元組(= 720 像素)
//	mov ax, 0221h        ; 第一條的位移
//	mov dx, 2000h        ; Hercules 的 bank 大小
//	@@: mov [si], ax
//	    add si, 2
//	    add ax, dx       ; 下一個 bank
//	    jns @F           ; 還沒溢出 bit15 就直接寫
//	    and ax, 7FFFh    ; 溢出 → 回到 bank 0
//	    add ax, bx       ;       並前進一列
//	@F: loop @B
//
// 這就是 Hercules 經典的**四路交錯**:位址 = bank×0x2000 + 列×90 + 欄。
// 三個獨立數字互相釘死表的長度:
//
//  1. 迴圈計數 `cx = 300`。
//  2. `0x0072 + 300×2 = 0x02CA`,而 0x02CA 正是**下一個變數**(離屏緩衝的段值,
//     `mov es, cs:[2CA]`)。多一筆或少一筆都會撞到它。
//  3. 起始位移 0x0221 = 545 = 6×90 + 5 → bank 內第 6 列、第 5 個位元組。
//     照交錯公式回推,第 i 筆對應的**實體掃描線正好是 24 + i**。
//
// # 三、於是幾何全部收斂
//
// 畫點那支(0x09FD)一開頭就把座標夾在 **0..319 / 0..199**:
//
//	cmp ax, 0        / jl  reject
//	cmp ax, 013Fh    / jg  reject      ; ★ 319
//	cmp bx, 0        / jl  reject
//	cmp bx, 00C7h    / jg  reject      ; ★ 199
//
// **呼叫端用的是原版的 320×200 座標**,放大是驅動程式自己的事:
//
//	X:  位元組 = x / 4      相位 = x & 3      ← 一個來源像素佔 2 bit = 橫向 2 倍
//	Y:  表項  = (y × 3) & ^1,而且**同一列畫兩條掃描線**(表項 n 與 n+1)
//
// 縱向那條算式值得攤開看,它就是「1.5 倍」的全部:
//
//	y=0 → 表項 0,1     y=1 → 表項 1,2      ← 表項 1 被兩列共用
//	y=2 → 表項 3,4     y=3 → 表項 4,5
//	y=4 → 表項 6,7     y=5 → 表項 7,8
//
// **每兩列來源 = 三條掃描線**,交界那條由兩列以 `or` 疊上去。沒有查表、沒有餘數
// 判斷,只有 `bx = y*3 & ~1` 加一次 `+2` —— 四支不同的常式(畫點 0x0A3C、
// 圖形 0x11DD、溶解 0x1D0D、以及清單列)算的都是這一條式子。
//
//	橫向 80 位元組 = 640 像素:清一列文字用 `mov cx,28h; rep stosw`(40 字 = 80 B)
//	                           捲動 / 淡化也是 `cx = 28h`
//
// ⇒ **實際畫面 640×300,置中在 Hercules 的 720×348**
// (上下各留 24 條、左右各留 40 像素;`(348−300)/2 = 24`、`(720−640)/2 = 40`
// 與表的起始值逐一對上)。
//
// 而 640×300 對 320×200 正好是**橫 2 倍、縱 1.5 倍** —— 與 `.HCS` 字型
// 8×8 → 16×12 是同一個比例。字格數也因此完全不變:
//
//	DOS      320×200,8×8   → 40 欄 × 25 列
//	Hercules 640×300,16×12 → 40 欄 × 25 列
//
// 驅動程式對遊戲呈現的是一個**與其他三種模式一模一樣的 320×200 / 40×25 世界**,
// 放大全部藏在驅動裡。這解釋了為什麼遊戲主體完全不必知道 Hercules 存在。
//
// # 四、加倍表與圖樣表:不在同一個軸上
//
// 這是原本那句「還沒追」真正的答案 —— 兩張表**從來不會一起用**:
//
//	0x031C(16 筆)  **幾何**:1bpp 的一個 nibble(4 個來源像素)→ 一個位元組
//	                (8 個螢幕像素)。只有畫 1bpp 圖形時用。
//	0x0318(4 筆)   **顏色**:色號 0..3 → 抖動位元組(00/55/AA/FF)。
//	                畫筆顏色與字的前景 / 背景用。
//
// 兩支 blit 各只用一張:
//
//	#26 圖形 blit:來源是 **1bpp**。`lodsb` 讀一個位元組 → 拆兩個 nibble →
//	              `mov bl, cs:[31C+bx]` 加倍 → `or es:[di], bl` 寫兩條掃描線。
//	              沒對齊時用 `ror bl, 相位×2` 再以遮罩切成 `[di]` 與 `[di+1]`。
//	              圖形目錄:`ds:[0]` = 筆數、`ds:[2+idx×2]` → `{u16 寬(像素),
//	              u16 列數, 位元…}`,寬度計數每個 nibble 減 4。
//	#31 字 blit:`.HCS` 的字模**已經是 16×12**,不需要放大 —— 所以它一列只寫
//	            一條掃描線(`cx = 0Ch`、`lodsw` 一列 2 B),而 12 = 8 的 1.5 倍
//	            已經烘在檔案裡。它用的是 0x0318:`dl`/`dh` 各取低 2 bit 當前景 /
//	            背景色 → 兩個抖動位元組 → 先用背景填滿 16 像素,再 `not`/`and`/`or`
//	            把前景合成上去。
//
// ⇒ **字型是「預先放大」,圖形是「blit 時放大」。** 同一個 1.5 倍,兩條路。
// 這也是 `.HCS` 為什麼要單獨存一份 16×12 的字型檔:8×8 拉到 16×12 沒法只靠加倍表。
//
// # 五、還沒解決的一段(誠實記帳)
//
// 驅動程式只收 **1bpp** 圖形,所以「16×16 的 2bpp tile 怎麼變成單色」這件事
// **不在驅動裡**,在遊戲本體(某個 `.OVL`)。`HerculesTile` 用的規則
// ——「取圖樣在該 x 位置的那個 bit」—— 是驅動自己的圖樣表所指的規則,
// 但**轉換發生在哪、是不是逐列換相位**尚未證實。
//
// 另外本檔只重現幾何,`internal/render` 目前仍在 320×200 的邏輯座標上畫圖,
// 沒有真的產生 640×300 的 Hercules 位元圖。要逐像素重現得再走一步。

// Hercules 的邏輯座標空間 —— 驅動程式夾住的範圍(`HER.DRV` 0x09FD)。
//
// 與 EGA / CGA / Tandy 完全相同,這正是重點:放大藏在驅動裡。
const (
	HerculesLogicalWidth  = 320
	HerculesLogicalHeight = 200
)

// Hercules 實際畫出來的範圍,以及它所在的實體畫面。
const (
	// HerculesScreenWidth 是實際用到的寬度(80 位元組 × 8 位元)。
	HerculesScreenWidth = 640
	// HerculesScreenHeight 是掃描線位址表的筆數。
	HerculesScreenHeight = HerculesScanlineCount

	// HerculesPhysWidth / HerculesPhysHeight 是 Hercules 圖形模式的全解析度。
	HerculesPhysWidth  = 720
	HerculesPhysHeight = 348

	// HerculesLeftMargin / HerculesTopMargin 是置中留下的邊。
	HerculesLeftMargin = (HerculesPhysWidth - HerculesScreenWidth) / 2   // 40
	HerculesTopMargin  = (HerculesPhysHeight - HerculesScreenHeight) / 2 // 24
)

// 放大倍率:橫 2 倍(一個來源像素 = 2 bit)、縱 3/2 倍(兩列 = 三條掃描線)。
const (
	HerculesXScaleNum = 2
	HerculesYScaleNum = 3
	HerculesYScaleDen = 2
)

// Hercules 顯示記憶體的佈局(`HER.DRV` 0x0770 的初始化迴圈)。
const (
	// HerculesSegment 是顯示記憶體的段值(`mov ax, 0B000h`)。
	HerculesSegment = 0xB000
	// HerculesBankSize 是一個 bank 的大小(`mov dx, 2000h`)。
	HerculesBankSize = 0x2000
	// HerculesBankCount 是 bank 數 —— 由 `and ax,7FFFh` 的溢出點決定
	// (0x2000 加四次才碰到 bit15)。
	HerculesBankCount = 4
	// HerculesRowStride 是 bank 內一列的位元組數(`mov bx, 005Ah`)= 720 像素。
	HerculesRowStride = 90
	// HerculesFirstOffset 是第一條掃描線的位移(`mov ax, 0221h`)
	// = 6 × 90 + 5,也就是 bank 內第 6 列、第 5 個位元組。
	HerculesFirstOffset = 0x0221
	// HerculesBytesPerLine 是遊戲實際寫的位元組數(清畫面 / 捲動的 `cx = 28h` × 2)。
	HerculesBytesPerLine = HerculesScreenWidth / 8
)

// 掃描線位址表在驅動程式裡的位置。
const (
	// HerculesScanlineTableOffset 是表的位移 —— 剛好接在 38 筆進入點表之後。
	HerculesScanlineTableOffset = 0x0072
	// HerculesScanlineCount 是表的筆數(`mov cx, 012Ch`)。
	HerculesScanlineCount = 300
	// HerculesNextVarOffset 是表後面那個變數(離屏緩衝的段值)。
	// 它 = 0x0072 + 300×2,所以筆數不可能是別的數 —— 這是第二份獨立佐證。
	HerculesNextVarOffset = HerculesScanlineTableOffset + HerculesScanlineCount*2
)

// HerculesEntryPointCount 是 `HER.DRV` 開頭的 `jmp near` 筆數。
//
// 38 × 3 = 0x72,而掃描線表就從 0x0072 開始 —— 兩個數字互相確認。
const HerculesEntryPointCount = HerculesScanlineTableOffset / 3

// HerculesDoubleNibble 是 `HER.DRV` 0x031C 的加倍表:
// 一個 nibble(4 個 1bpp 像素)→ 一個位元組(8 個螢幕像素),每個 bit 複製成兩個。
var HerculesDoubleNibble = [16]byte{
	0x00, 0x03, 0x0C, 0x0F, 0x30, 0x33, 0x3C, 0x3F,
	0xC0, 0xC3, 0xCC, 0xCF, 0xF0, 0xF3, 0xFC, 0xFF,
}

// HerculesPhaseMask 是 `HER.DRV` 0x030C:相位 0..3 各佔哪兩個 bit(高位在左)。
var HerculesPhaseMask = [4]byte{0xC0, 0x30, 0x0C, 0x03}

// HerculesKeepMask 是 `HER.DRV` 0x0308:上面那張的補數,畫點前先把該像素挖掉。
var HerculesKeepMask = [4]byte{0x3F, 0xCF, 0xF3, 0xFC}

// Hercules 的字格 —— `.HCS` 是 16×12,而 640/16 = 40、300/12 = 25。
const (
	HerculesCellWidth  = HCSGlyphWidth
	HerculesCellHeight = HCSGlyphHeight
	HerculesCharCols   = HerculesScreenWidth / HerculesCellWidth
	HerculesCharRows   = HerculesScreenHeight / HerculesCellHeight
)

// BuildHerculesScanlineTable 重跑 `HER.DRV` 0x0770 的初始化迴圈。
//
// 逐指令對應,包含 `jns` 那個「靠 bit15 溢出偵測換列」的小把戲 ——
// 原版沒有除法也沒有取餘數,只有加法與一次遮罩。
func BuildHerculesScanlineTable() [HerculesScanlineCount]uint16 {
	var tbl [HerculesScanlineCount]uint16
	ax := uint16(HerculesFirstOffset)
	for i := 0; i < HerculesScanlineCount; i++ {
		tbl[i] = ax
		ax += HerculesBankSize
		if ax&0x8000 != 0 { // jns 沒跳 → 溢出了
			ax &= 0x7FFF
			ax += HerculesRowStride
		}
	}
	return tbl
}

// HerculesPhysicalScanline 回報表的第 i 筆對應實體畫面的哪一條掃描線。
//
// 由交錯公式反推:bank = 掃描線 & 3、bank 內列 = 掃描線 >> 2。
func HerculesPhysicalScanline(i int) int {
	return HerculesTopMargin + i
}

// HerculesScanlineEntries 回報邏輯第 y 列要畫到表的哪兩筆
// (`bx = y*3 & ^1`,然後 `+2` 再畫一次)。
//
// 回傳的兩個表項相鄰;相鄰兩列會共用一筆 —— 那就是縱向 1.5 倍。
func HerculesScanlineEntries(y int) (int, int) {
	// 原版是 `bx = y*3`(shl+add)然後 `and bx,0FFFEh`,而 bx 是**位元組**位移,
	// 所以除以 2 才是表項編號。取偶再除 2 = 向下取整,不能先除。
	byteOff := (y * HerculesYScaleNum) & 0xFFFE
	n := byteOff / 2
	return n, n + 1
}

// HerculesByteColumn / HerculesPhase 把邏輯 x 拆成位元組欄與相位
// (`shr ax,1; shr ax,1` 與 `and si,3`)。
func HerculesByteColumn(x int) int { return x / 4 }

// HerculesPhase 回報邏輯 x 落在該位元組的第幾組 2 bit。
func HerculesPhase(x int) int { return x & 3 }
