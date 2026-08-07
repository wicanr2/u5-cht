package game

import "fmt"

// 不列顛尼亞的曆法
//
// 出自原版的時間推進函式 sub_29304 與它下游的日期進位:
//
//	byte_3E091 分   +minutes;> 59 → 減 60 並進位到小時
//	byte_3E08F 時   >= 24 → 減 24 並進位到日
//	byte_3E08E 日   > 28  → 設回 1 並進位到月
//	byte_3E08D 月   > 13  → 設回 1 並進位到年
//	word_3E084 年
//
// 也就是**每月 28 天、每年 13 個月**。一般行動每回合推進 **1 分鐘**
// (sub_1DC8 呼叫 sub_29304(1));紮營之類的動作一次推 20 分鐘。
const (
	MinutesPerHour = 60
	HoursPerDay    = 24
	DaysPerMonth   = 28
	MonthsPerYear  = 13

	// MinutesPerTurn 是一般行動推進的時間。
	MinutesPerTurn = 1
)

// Clock 是遊戲內的時間。日與月是 1-based(原版進位後設回 1,不是 0)。
type Clock struct {
	Minute int
	Hour   int
	Day    int
	Month  int
	Year   int
}

// NewClock 回傳一個合法的起始時間。
//
// ⚠ 這**不是**原版新遊戲的開局時間 —— 那個值存在存檔範本裡,存檔格式還沒解。
// 在解出來之前給一個中午,免得測試與截圖都在半夜、NPC 全在睡覺看不出對錯。
func NewClock() Clock {
	return Clock{Hour: 12, Day: 1, Month: 1, Year: 139}
}

// Advance 推進 n 分鐘,依原版的進位規則。
func (c *Clock) Advance(n int) {
	c.Minute += n
	for c.Minute >= MinutesPerHour {
		c.Minute -= MinutesPerHour
		c.Hour++
	}
	for c.Hour >= HoursPerDay {
		c.Hour -= HoursPerDay
		c.Day++
	}
	for c.Day > DaysPerMonth {
		c.Day -= DaysPerMonth
		c.Month++
	}
	for c.Month > MonthsPerYear {
		c.Month -= MonthsPerYear
		c.Year++
	}
}

// HoursSince 回傳從 old 到現在跨過幾個小時邊界。
//
// ⚠ 不是「小時欄位差多少」—— 那在跨日、跨月時會變成負數。
// 一次推進整晚(紮營)時這個差別是決定性的。
func (c Clock) HoursSince(old Clock) int {
	return c.totalHours() - old.totalHours()
}

// totalHours 把時間攤平成小時數,供比較用。年的長度是 13 × 28 天。
func (c Clock) totalHours() int {
	days := ((c.Year*MonthsPerYear+(c.Month-1))*DaysPerMonth) + (c.Day - 1)
	return days*HoursPerDay + c.Hour
}

// String 是給 HUD 用的短格式。
func (c Clock) String() string {
	return fmt.Sprintf("%02d:%02d", c.Hour, c.Minute)
}

// DateString 是完整日期。
func (c Clock) DateString() string {
	return fmt.Sprintf("%d 年 %d 月 %d 日", c.Year, c.Month, c.Day)
}
