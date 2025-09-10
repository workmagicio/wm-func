package main

import (
	"time"
)

var Day = time.Hour * 24

// StreamSlice 定义了需要进行时间切分的原始时间段信息
type StreamSlice struct {
	StartTime  time.Time // 起始时间
	EndTime    time.Time // 结束时间
	DateFormat string    // 输出时间字符串的格式
	Duration   int
}

// NewStreamSlice 是 StreamSlice 的构造函数
func NewStreamSlice(start time.Time, end time.Time, dateFormat string, subtype string) *StreamSlice {
	duration := timeStepDailyMap[subtype]

	return &StreamSlice{
		StartTime:  start,
		EndTime:    end,
		DateFormat: dateFormat,
		Duration:   duration,
	}
}

// Slice 代表一个分割后的时间片
type Slice struct {
	Start string // 格式化后的起始时间字符串
	End   string // 格式化后的结束时间字符串
}

// GetSlice 按照指定的 Duration 将 StreamSlice 分割成多个 Slice
func (s *StreamSlice) GetSlice() []Slice {

	if s.Duration <= 0 {
		return []Slice{
			{
				Start: s.StartTime.Format(s.DateFormat),
				End:   s.EndTime.Format(s.DateFormat),
			},
		}
	}

	duration := time.Duration(s.Duration) * Day

	var slices []Slice
	currentStart := s.StartTime
	for currentStart.Before(s.EndTime) {
		nextStart := currentStart.Add(duration)
		if nextStart.After(s.EndTime) {
			nextStart = s.EndTime
		}
		displayEnd := nextStart.Add(-1 * time.Nanosecond)
		if displayEnd.Before(currentStart) {
			displayEnd = currentStart
		}
		slice := Slice{
			Start: currentStart.Format(s.DateFormat),
			End:   displayEnd.Format(s.DateFormat),
		}
		slices = append(slices, slice)
		currentStart = nextStart.Add(time.Hour * 24)
	}
	return slices
}

func GetStreamSlice(start, end time.Time, dateFormat string, subtype string) []Slice {
	return NewStreamSlice(start, end, dateFormat, subtype).GetSlice()
}
